package oss

import (
	"filestore-server/config"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var ossCli *oss.Client

// OssClient: 获取 oss client
func OssClient() *oss.Client {
	if ossCli != nil {
		return ossCli
	}

	client, e := oss.New(config.OSSEndpoint, config.OSSAccessKeyID, config.OSSAccessKeySecret)
	if e != nil {
		fmt.Println(e.Error())
		return nil
	}

	return client
}

// OssBucket: 获取 bucket 存储空间
func OssBucket() *oss.Bucket {
	client := OssClient()
	if client != nil {
		bucket, e := client.Bucket(config.OSSBucket)
		if e != nil {
			fmt.Println(e.Error())
			return nil
		}

		return bucket
	}
	return nil
}

// GetDownloadSignedUrl:临时授权下载URL
func GetDownloadSignedUrl(objName string) string {
	signedUrl, err := OssBucket().SignURL(objName, oss.HTTPGet, 3600)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return signedUrl
}
