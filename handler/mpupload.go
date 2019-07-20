package handler

import (
	"filestore-server/cache/redis"
	"filestore-server/util"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"
)

// MultiparUploadInfo: 初始化信息
type MultiparUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadId   string //标记当前分块上传的唯一id 规则：username+当前时间戳
	ChunkSize  int    //当前分块的大小
	ChunkCount int    //分块的数量
}

// 每块的大小
var chunkSize = 5 * 1024 * 1024 //5MB
var hSetKeyPrefix = "MP_"

// 初始化分块上传
func InitMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	//1.解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.RespMsg{Code: -1, Msg: "params invalid"}.JsonToBytes())
		return
	}

	//2.获得 redis 的一个链接
	rConn := redis.RedisPool().Get()
	defer rConn.Close()

	//3.生成分块上传的初始化信息
	upinfo := MultiparUploadInfo{
		FileHash:  filehash,
		FileSize:  filesize,
		UploadId:  username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize: chunkSize,
		// 文件总大小/分块大小  然后向上取整
		ChunkCount: int(math.Ceil(float64(filesize / chunkSize))),
	}

	//4.将初始化信息写入到 redis 缓存
	rConn.Do("HSET", hSetKeyPrefix+upinfo.UploadId, "chunkcount", upinfo.ChunkCount)
	rConn.Do("HSET", hSetKeyPrefix+upinfo.UploadId, "filehash", upinfo.FileHash)
	rConn.Do("HSET", hSetKeyPrefix+upinfo.UploadId, "filsize", upinfo.FileSize)

	//5.将响应初始化数据返回到客户端
	w.Write(util.RespMsg{Code: 0, Msg: "OK", Data: upinfo}.JsonToBytes())
}
