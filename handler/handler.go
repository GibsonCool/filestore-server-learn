package handler

import (
	"encoding/json"
	"filestore-server/meta"
	"filestore-server/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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

		//将网络文件内容从内存拷贝到创建的文件中
		fileMeta.FileSize, e = io.Copy(newFile, file)
		if e != nil {
			fmt.Printf("Failed to save data into file ,err:%s", e.Error())
			return
		}

		//将文件的句柄移到头部，计算文件的 sha1 值
		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		meta.UpdateFileMeta(fileMeta)

		// 上传完成，重定向提示用户
		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)

	}
}

// UploadSucHandler:上传已完成
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

// GetFileMetaHandler:获取文件元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form["filehash"][0]
	fMeta := meta.GetFileMeta(filehash)
	data, e := json.Marshal(fMeta)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(data)

}
