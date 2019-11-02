package main

import (
	"filestore-server/config"
	"filestore-server/route"
	"filestore-server/util"
	"log"
)

func main() {
	config.SetupInit()

	router := route.Router()

	log.Printf("上传服务启动中，开始监听【%s】...\n", config.UploadServiceHost)

	err := router.Run(config.UploadServiceHost)

	util.SimplePrint(err, util.FailedStartServer)
}
