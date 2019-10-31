package oss

import (
	"filestore-server/config"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"log"
)

var ossCli *oss.Client

// OssClient: 获取 oss client
func OssClient() *oss.Client {
	if ossCli != nil {
		return ossCli
	}

	client, e := oss.New(config.OSSEndpoint, config.OSSAccessKeyID, config.OSSAccessKeySecret)
	if e != nil {
		log.Println(e.Error())
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
			log.Println(e.Error())
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
		log.Println(err.Error())
		return ""
	}
	return signedUrl
}

// BuildLifecycleRule: 针对指定的 bucket 设置生命周期规则
func BuildLifecycleRule(bucketName string) {
	// 表示前缀为 test/ 的对象文件距离最后修改时间10天后过期
	ruleTest1 := oss.BuildLifecycleRuleByDays("rule1", "test/", true, 10)
	rules := []oss.LifecycleRule{ruleTest1}

	OssClient().SetBucketLifecycle(bucketName, rules)
}
