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
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
func DownloadHandler(c *gin.Context) {
	fsha1 := c.Request.FormValue("filehash")

	// 根据下载参数文件的 hash 值查询出 文件元信息
	fileMeta, _ := meta.GetFileMetaDB(fsha1)

	c.FileAttachment(fileMeta.Location, fileMeta.FileName)
}

// FileNameUpdateHandler: 修改文件名称
func FileNameUpdateHandler(c *gin.Context) {
	opType := c.Request.FormValue("op")
	fileSha1 := c.Request.FormValue("filehash")
	newFileName := c.Request.FormValue("filename")
	userName := c.Request.FormValue("username")

	if opType != "0" {
		c.JSON(http.StatusOK, util.RespMsg{
			Code: -1,
			Msg:  "op 操作错误",
		})
		return
	}

	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusOK, util.RespMsg{
			Code: -2,
			Msg:  "该接口不支持这种 http method:" + c.Request.Method,
		})
		return
	}

	curFileMeta, e := meta.GetFileMetaDB(fileSha1)
	if e != nil {
		c.JSON(http.StatusOK, util.RespMsg{
			Code: -2,
			Msg:  fmt.Sprintf("获取文件信息失败,filehash:%s", fileSha1),
		})
		return
	}

	tblUserFileSuc := meta.UpdateUserFileNameDB(userName, newFileName, *curFileMeta)
	if !tblUserFileSuc {
		c.JSON(http.StatusOK, util.RespMsg{
			Code: -3,
			Msg:  "更新用户文件表失败",
		})
		return
	}

	c.JSON(http.StatusOK, util.RespMsg{
		Code: 0,
		Msg:  "文件名修改成功！~",
	})
}

// FiledDeleteHandler: 删除文件及元信息
func FiledDeleteHandler(c *gin.Context) {
	fileSha1 := c.Request.FormValue("filehash")
	getFileMeta, e := meta.GetFileMetaDB(fileSha1)
	if e != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// 删除文件
	os.Remove(getFileMeta.Location)
	// 删除文件元信息
	meta.RemoveFileMeta(fileSha1)
	// TODO：删除表文件信息

	c.Status(http.StatusOK)

}

// TryFastUploadHandler:尝试秒传接口
func TryFastUploadHandler(c *gin.Context) {
	//1.解析请求参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	filesize := c.Request.FormValue("filesize")

	//2.从文件列表查询相同hash的文件记录
	fileMeta, e := meta.GetFileMetaDB(filehash)
	if e != nil {
		log.Println(e.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	//3.查不到记录则返回秒传失败
	if fileMeta.FileSha1 == "" {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		c.Data(http.StatusOK, "application/json", resp.JsonToBytes())
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
		c.Data(http.StatusOK, "application/json", resp.JsonToBytes())
		return
	}

	resp := util.RespMsg{
		Code: -2,
		Msg:  "秒传失败，请稍后重试",
	}
	c.Data(http.StatusOK, "application/json", resp.JsonToBytes())
	return

}

// DownloadURLHandler: 生成文件的下载地址
func DownloadURLHandler(c *gin.Context) {
	filehash := c.Request.FormValue("filehash")
	log.Println(filehash)

	// 从文件表查找记录
	fileMeta, e := meta.GetFileMetaDB(filehash)
	if e != nil {
		log.Println(e.Error())
	}

	if strings.HasPrefix(fileMeta.Location, "./tmp") {
		//  获取本地下载路径
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")
		tmpUrl := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
			c.Request.Host, filehash, username, token)

		c.Data(http.StatusOK, "octet-stream", []byte(tmpUrl))
	} else if strings.HasPrefix(fileMeta.Location, "test/") {
		// 获取阿里云 oss 下载路径
		signedUrl := oss.GetDownloadSignedUrl(fileMeta.Location)

		c.Data(http.StatusOK, "octet-stream", []byte(signedUrl))
	} else {

		c.Data(http.StatusOK, "octet-stream", []byte("无法识别的下载路径："+fileMeta.Location))
	}

}
