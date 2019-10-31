package main

import (
	myRedis "filestore-server/cache/redis"
	"filestore-server/config"
	mySql "filestore-server/db/mysql"
	fileMeta "filestore-server/meta"
	"filestore-server/route"
	"filestore-server/util"
	"log"
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

	router := route.Router()

	log.Printf("上传服务启动中，开始监听【%s】...\n", config.UploadServiceHost)

	err := router.Run(config.UploadServiceHost)

	util.SimplePrint(err, util.FailedStartServer)
}
