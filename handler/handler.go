package handler

import (
	"encoding/json"
	dblayer "filestore-server/db"
	"filestore-server/meta"
	"filestore-server/store/oss"
	"filestore-server/util"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

//UploadHandler： 用于处理文件上传
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// 返回上传 html 页面
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internel server error")
			return
		}

		io.WriteString(w, string(data))
	} else if r.Method == http.MethodPost {
		/*
			接受文件流及存储到本地目录
		*/

		//读取文件内容
		file, header, e := r.FormFile("file")
		util.SimplePrint(e, util.FailedGetData)
		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName: header.Filename,
			Location: "./tmp/" + header.Filename,
			UploadAt: time.Now().In(util.CstZone).Format("2006-01-02 15:04:05"),
		}

		//创建内容接收文件
		newFile, e := os.Create(fileMeta.Location)
		if e != nil {
			fmt.Printf("Failed to create file, err:%s", e.Error())
			return
		}
		defer newFile.Close()

		//将网络文件内容从内存拷贝到创建的文件中，并复制文件大小 FileSize 字段
		fileMeta.FileSize, e = io.Copy(newFile, file)
		if e != nil {
			fmt.Printf("Failed to save data into file ,err:%s", e.Error())
			return
		}

		//将文件的句柄移到头部，计算文件的 sha1 值
		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)

		/*
			同时将文件写入 OSS 存储
		*/
		newFile.Seek(0, 0)
		ossPath := "test/" + fileMeta.FileName
		e = oss.OssBucket().PutObject(ossPath, newFile)
		if e != nil {
			fmt.Println(e.Error())
			w.Write([]byte("upload failed!"))
			return
		}
		fileMeta.Location = ossPath

		//meta.UpdateFileMeta(fileMeta)
		_ = meta.UpdateFileMetaDB(fileMeta)

		// 更新用户文件记录
		r.ParseForm()
		username := r.Form.Get("username")
		isSuc := dblayer.OnUserFiledUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
		if isSuc {
			// 上传完成，跳转到home页面
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("upload failed:  更新用户文件表记录失败"))
		}

	}
}

// UploadSucHandler:上传已完成
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

// GetFileMetaHandler:获取文件元信息
// 浏览器访问---》http://localhost:8080/file/meta?filehash=5913ebee4876c3a3265851e9855b75d1898377f3
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	//解析请求参数
	r.ParseForm()
	//获取参数第一个值
	filehash := r.Form["filehash"][0]
	//fMeta := meta.GetFileMeta(filehash)
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, e := json.Marshal(fMeta)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(data)

}

// FileQueryHandler: 查询批量的文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	//不直接查询文件表，改去查询用户文件表
	//userFile, e := dblayer.GetFileMetaList(limitCnt)

	userFile, e := dblayer.QueryUserFileMetas(username, limitCnt)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, e := json.Marshal(userFile)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
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
		fmt.Println(e.Error())
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
