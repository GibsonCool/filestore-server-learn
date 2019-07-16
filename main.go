package main

import (
	"filestore-server/errorUtils"
	"filestore-server/handler"
	"net/http"
)

func main() {
	// 配置路由
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)

	err := http.ListenAndServe(":8080", nil)

	errorUtils.SimplePrint(err, errorUtils.FailedStartServer)
}
