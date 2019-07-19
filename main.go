package main

import (
	"filestore-server/handler"
	"filestore-server/util"
	"net/http"
)

func main() {
	// 配置路由
	//配置静态资源处理
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("./static"))))

	// 文件操作
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	http.HandleFunc("/file/meta", handler.GetFileMetaHandler)
	http.HandleFunc("/file/query", handler.FileQueryHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.HandleFunc("/file/update", handler.FileUpdateMetaUpdateHandler)
	http.HandleFunc("/file/delete", handler.FiledeleteHandler)

	// 用户操作
	http.HandleFunc("/user/signup", handler.SignupHandler)
	http.HandleFunc("/user/signin", handler.SignInHandler)
	http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler))

	err := http.ListenAndServe(":8080", nil)

	util.SimplePrint(err, util.FailedStartServer)
}
