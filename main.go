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
	http.HandleFunc("/file/query", handler.FileQueryHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.HandleFunc("/file/update", handler.FileUpdateMetaUpdateHandler)
	http.HandleFunc("/file/delete", handler.FiledeleteHandler)

	err := http.ListenAndServe(":8080", nil)

	util.SimplePrint(err, util.FailedStartServer)
}
