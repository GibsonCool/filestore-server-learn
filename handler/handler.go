package handler

import (
	"encoding/json"
	dblayer "filestore-server/db"
	"filestore-server/meta"
	"filestore-server/util"
	"fmt"
	"io"
	"io/ioutil"
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
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
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
	fileMetas := meta.GetLastFileMetas(limitCnt)
	data, e := json.Marshal(fileMetas)
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
	fileMeta := meta.GetFileMeta(fsha1)

	// 根据元信息的文件路径打开文件，读取并返回给请求方
	file, e := os.Open(fileMeta.Location)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
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

	curFileMeta := meta.GetFileMeta(fileSha1)
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)

	data, e := json.Marshal(curFileMeta)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

// FiledeleteHandler: 删除文件及元信息
func FiledeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fileSha1 := r.Form.Get("filehash")
	getFileMeta := meta.GetFileMeta(fileSha1)
	os.Remove(getFileMeta.Location)

	meta.RemoveFileMeta(fileSha1)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("删除成功，文件名称：" + getFileMeta.FileName))

}
