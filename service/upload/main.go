package main

import (
	myRedis "filestore-server/cache/redis"
	"filestore-server/config"
	mySql "filestore-server/db/mysql"
	"filestore-server/handler"
	fileMeta "filestore-server/meta"
	"filestore-server/util"
	"log"
	"net/http"
)

/*
	将各种初始化操作统一入口，替代 init() 方法。也便于理解和和控制初始化顺序
*/
func setupInit() {
	mySql.Setup()
	myRedis.Setup()
	fileMeta.Setup()
}

func main() {
	setupInit()

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
	// 秒传接口
	http.HandleFunc("/file/fastupload", handler.HTTPInterceptor(handler.TryFastUploadHandler))

	// 鉴权下载URL
	http.HandleFunc("/file/downloadurl", handler.DownloadURLHandler)

	//分块上传
	http.HandleFunc("/file/mpupload/init", handler.HTTPInterceptor(handler.InitMultipartUploadHandler))
	http.HandleFunc("/file/mpupload/uppart", handler.HTTPInterceptor(handler.UploadPartHandler))
	http.HandleFunc("/file/mpupload/complete", handler.HTTPInterceptor(handler.CompleteUploadHander))

	// 用户操作
	http.HandleFunc("/user/signup", handler.SignupHandler)
	http.HandleFunc("/user/signin", handler.SignInHandler)
	http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler))

	log.Printf("上传服务启动中，开始监听【%s】...\n", config.UploadServiceHost)
	err := http.ListenAndServe(config.UploadServiceHost, nil)

	util.SimplePrint(err, util.FailedStartServer)
}
