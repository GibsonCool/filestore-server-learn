package main

import (
	"bufio"
	"encoding/json"
	"filestore-server/config"
	dblayer "filestore-server/db"
	"filestore-server/mq"
	"filestore-server/store/oss"
	"log"
	"os"
)

func ProcessTransfer(msg []byte) bool {
	// 1.解析msg
	pubData := mq.TransferData{}
	e := json.Unmarshal(msg, &pubData)
	if e != nil {
		log.Println(e.Error())
		return false
	}

	// 2.根据零食㽾文件路径，创建文件句柄
	file, e := os.Open(pubData.CurlLocation)
	if e != nil {
		log.Println(e.Error())
		return false
	}

	// 3.通过文件句柄将文件内容读出来并上传到 oss
	e = oss.OssBucket().PutObject(
		pubData.DestLocation,
		bufio.NewReader(file),
	)
	if e != nil {
		log.Println(e.Error())
		return false
	}

	// 4.更新文件的存储路径到文件表
	suc := dblayer.UploadFileLocation(pubData.FileHash, pubData.DestLocation)
	if !suc {
		return false
	}
	return true
}

func main() {
	if !config.AsyncTransferEnable {
		log.Println("异步转移上传任务功能被禁用，请查看相关配置。")
		return
	}

	log.Println("开始监听转移任务队列...")

	mq.StartConsume(
		config.TransOSSQueueName,
		"transfer_oss",
		ProcessTransfer,
	)
}
