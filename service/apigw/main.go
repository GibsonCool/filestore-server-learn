package main

import (
	"filestore-server/config"
	"filestore-server/service/apigw/route"
)

func main() {
	router := route.Router()
	router.Run(config.UploadServiceHost)
}
