package config

import (
	myRedis "filestore-server/cache/redis"
	mySql "filestore-server/db/mysql"
	fileMeta "filestore-server/meta"
)

/*
	将各种初始化操作统一入口，替代 init() 方法。也便于理解和和控制初始化顺序
*/
func SetupInit() {
	mySql.Setup()
	myRedis.Setup()
	fileMeta.Setup()
}
