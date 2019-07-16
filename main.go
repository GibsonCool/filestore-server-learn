package main

import (
	"filestore-server/handler"
	"filestore-server/util"
	"net/http"
)

func main() {
	// 配置路由
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	http.HandleFunc("/file/meta", handler.GetFileMetaHandler)

	err := http.ListenAndServe(":8080", nil)

	util.SimplePrint(err, util.FailedStartServer)
}
