package handler

import (
	"bytes"
	"encoding/json"
	"filestore-server/config"
	dblayer "filestore-server/db"
	"filestore-server/meta"
	"filestore-server/mq"
	"filestore-server/store"
	"filestore-server/store/oss"
	"filestore-server/util"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

//UploadHandler： 用于处理文件上传
func UploadHandler(c *gin.Context) {
	// 返回上传 html 页面
	data, err := ioutil.ReadFile("./static/view/index.html")
	if err != nil {
		c.String(http.StatusNotFound, "网页不存在！~")
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(data))
}

func DoUploadHandler(c *gin.Context) {
	errCode := 0
	errMsg := ""
	defer func() {
		if errCode < 0 {
			c.JSON(http.StatusOK, util.RespMsg{
				Code: errCode,
				Msg:  errMsg,
			})
		}
	}()

	// 1.读取文件内容
	file, header, e := c.Request.FormFile("file")
	util.SimplePrint(e, util.FailedGetData)
	defer file.Close()

	// 2.将文件转为 []byte
	buf := bytes.NewBuffer(nil)
	if _, e := io.Copy(buf, file); e != nil {
		errCode = -1
		errMsg = "failed to get file data, err:" + e.Error()
		log.Println(errMsg)
		return
	}

	// 3.构建文件元信息
	fileMeta := meta.FileMeta{
		FileName: header.Filename,
		FileSha1: util.Sha1(buf.Bytes()),
		Location: "./tmp/" + header.Filename,
		FileSize: int64(len(buf.Bytes())),
		UploadAt: time.Now().In(util.CstZone).Format("2006-01-02 15:04:05"),
	}

	// 4.将文件写入临时的存储位置
	newFile, e := os.Create(fileMeta.Location)
	if e != nil {
		log.Printf("Failed to create file, err:%s", e.Error())
		return
	}
	defer newFile.Close()

	nByge, e := newFile.Write(buf.Bytes())
	if int64(nByge) != fileMeta.FileSize || e != nil {
		errCode = -2
		errMsg = "Failed to save data into file, writtenSize:" + string(int64(nByge)) + "  fileSize: " + string(fileMeta.FileSize)
		log.Println(errMsg)
		if e != nil {
			log.Printf("err:%s\n", e.Error())
			errMsg += e.Error()
		}
		return
	}

	// 5.同步或异步将文件转移到 oss 中
	newFile.Seek(0, 0) //将游标重新移回到文件头部
	ossPath := "test/" + fileMeta.FileName
	if !config.AsyncTransferEnable {
		// 同步任务直接写入
		e = oss.OssBucket().PutObject(ossPath, newFile)
		if e != nil {
			errCode = -3
			errMsg = "oss upload failed! err:" + e.Error()
			log.Println(errMsg)
			return
		}
		fileMeta.Location = ossPath
	} else {
		// 异步任务转移到 rabbmitMq 任务队列
		transferData := mq.TransferData{
			FileHash:      fileMeta.FileSha1,
			CurlLocation:  fileMeta.Location,
			DestLocation:  ossPath,
			DestStoreType: store.StoreOss,
		}
		pubData, _ := json.Marshal(transferData)
		log.Printf("异步任务转移信息：%s", pubData)
		pubSuc := mq.Publish(
			config.TransExchangeName,
			config.TransOSSRoutingKey,
			pubData,
		)

		if !pubSuc {
			//TODO: 当前任务发送转移消息失败，加入错误队列稍后重试。
		}

	}

	// 6.更新用户文件表记录
	//meta.UpdateFileMeta(fileMeta)
	_ = meta.UpdateFileMetaDB(fileMeta)

	username := c.Request.FormValue("username")
	isSuc := dblayer.OnUserFiledUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
	if isSuc {
		// 上传完成，跳转到home页面
		c.Redirect(http.StatusFound, "/static/view/home.html")
	} else {
		errCode = -4
		errMsg = "upload failed:  更新用户文件表记录失败"
	}
}

// UploadSucHandler:上传已完成
func UploadSucHandler(c *gin.Context) {
	c.JSON(http.StatusOK, util.RespMsg{
		Code: 0,
		Msg:  "Upload finished!",
	})
}

// GetFileMetaHandler:获取文件元信息
// 浏览器访问---》http://localhost:8080/file/meta?filehash=5913ebee4876c3a3265851e9855b75d1898377f3
func GetFileMetaHandler(c *gin.Context) {
	//获取参数第一个值
	filehash := c.Request.FormValue("filehash")
	//fMeta := meta.GetFileMeta(filehash)
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.RespMsg{
			Code: -1,
			Msg:  "Upload failed!",
		})
		return
	}

	if fMeta != nil {

		data, e := json.Marshal(fMeta)
		if e != nil {
			log.Printf("Upload failed! err:%s\n", e.Error())
			c.JSON(http.StatusInternalServerError, util.RespMsg{
				Code: -2,
				Msg:  "Upload failed!",
			})
			return
		}
		c.Data(http.StatusOK, "application/json", data)
	} else {
		c.JSON(http.StatusOK, util.RespMsg{
			Code: -3,
			Msg:  "No sub file",
		})
	}

}

// FileQueryHandler: 查询批量的文件元信息
func FileQueryHandler(c *gin.Context) {

	limitCnt, _ := strconv.Atoi(c.Request.FormValue("limit"))
	username := c.Request.FormValue("username")
	//不直接查询文件表，改去查询用户文件表
	//userFile, e := dblayer.GetFileMetaList(limitCnt)

	userFile, e := dblayer.QueryUserFileMetas(username, limitCnt)
	if e != nil {
		c.JSON(http.StatusInternalServerError, util.RespMsg{
			Code: -1,
			Msg:  "Query filed!",
		})
		return
	}

	//log.Println(userFile)
	c.JSON(http.StatusOK, util.RespMsg{
		Code: 0,
		Msg:  "获取成功",
		Data: userFile,
	})
}

// DownloadHandler: 根据参数 filehash 下载文件
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fsha1 := r.Form.Get("filehash")

	// 根据下载参数文件的 hash 值查询出 文件元信息
	fileMeta, e := meta.GetFileMetaDB(fsha1)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(e.Error()))
		return
	}

	// 根据元信息的文件路径打开文件，读取并返回给请求方
	file, e := os.Open(fileMeta.Location)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(e.Error()))
		return
	}
	defer file.Close()

	data, e := ioutil.ReadAll(file)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 设置一下 response header 让浏览器识别支持文件下载, 如果不设置是直接在浏览器展示
	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Content-Disposition", "attachment;filename=\""+fileMeta.FileName+"\"")
	w.Write(data)
}

// FileUpdateMetaUpdateHandler: 修改文件元信息
func FileUpdateMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	curFileMeta, e := meta.GetFileMetaDB(fileSha1)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(e.Error()))
		return
	}
	curFileMeta.FileName = newFileName
	meta.UpdateFileMetaDB(*curFileMeta)

	data, e := json.Marshal(curFileMeta)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)

}

// FiledeleteHandler: 删除文件及元信息
func FiledeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fileSha1 := r.Form.Get("filehash")
	getFileMeta, e := meta.GetFileMetaDB(fileSha1)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(e.Error()))
		return
	}

	os.Remove(getFileMeta.Location)

	meta.RemoveFileMeta(fileSha1)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("删除成功，文件名称：" + getFileMeta.FileName))

}

// TryFastUploadHandler:尝试秒传接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	//1.解析请求参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize := r.Form.Get("filesize")

	//2.从文件列表查询相同hash的文件记录
	fileMeta, e := meta.GetFileMetaDB(filehash)
	if e != nil {
		log.Println(e.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//3.查不到记录则返回秒传失败
	if fileMeta.FileSha1 == "" {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JsonToBytes())
		return
	}

	//4.上传过则将文件信息写入用户文件表，返回成功
	parseFileSize, _ := strconv.ParseInt(filesize, 10, 64)

	suc := dblayer.OnUserFiledUploadFinished(username, filehash, filename, parseFileSize)
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JsonToBytes())
		return
	} else {
		resp := util.RespMsg{
			Code: -2,
			Msg:  "秒传失败，请稍后重试",
		}
		w.Write(resp.JsonToBytes())
		return
	}

}

// DownloadURLHandler: 生成文件的下载地址
func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	log.Println(filehash)

	// 从文件表查找记录
	fileMeta, e := meta.GetFileMetaDB(filehash)
	if e != nil {
		log.Println(e.Error())
	}
	signedUrl := oss.GetDownloadSignedUrl(fileMeta.Location)
	w.Write([]byte(signedUrl))

}
