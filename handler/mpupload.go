package handler

import (
	rPool "filestore-server/cache/redis"
	dblayer "filestore-server/db"
	"filestore-server/util"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gpmgo/gopm/modules/log"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
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

// InitMultipartUploadHandler：初始化分块上传
func InitMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	//1.解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write((&util.RespMsg{Code: -1, Msg: "params invalid"}).JsonToBytes())
		return
	}

	//2.获得 redis 的一个链接
	rConn := rPool.RedisPool().Get()
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
	w.Write((&util.RespMsg{Code: 0, Msg: "OK", Data: upinfo}).JsonToBytes())
}

// UploadPartHandler：上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	//1.解析用户请求参数
	r.ParseForm()
	//username := r.Form.Get("username")
	uploadId := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

	fmt.Printf("uploadid:%s  chunkIndex:%s\n", uploadId, chunkIndex)
	//2.获得 redis 的一个链接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	//3.获得文件句柄，用于存储分块内容
	fpath := util.GetCurrentFielParentPath() + "/tmp/" + uploadId + "/" + chunkIndex
	// 数字设定法：：0表示没有权限，1表示可执行权限，2表示可写权限，4表示可读权限，然后将其相加。设置当前用户可读可写可执行权限
	os.Mkdir(path.Dir(fpath), 0744)
	file, e := os.Create(fpath)
	if e != nil {
		w.Write((&util.RespMsg{Code: -1, Msg: "Upload part failed:" + e.Error(), Data: nil}).JsonToBytes())
		return
	}
	defer file.Close()

	//读取内存中的分块内容写入到文件中
	buf := make([]byte, 1023*1024)
	for {
		n, e := r.Body.Read(buf)
		file.Write(buf[:n])
		if e != nil {
			break
		}
	}

	//4.更新 redis 缓存状态
	rConn.Do("HSET", hSetKeyPrefix+uploadId, "chkidx_"+chunkIndex, 1)

	//5.返回结果给到客户端
	w.Write((&util.RespMsg{Code: 0, Msg: "OK", Data: nil}).JsonToBytes())
}

// CompleteUploadHander:通知上传合并
func CompleteUploadHander(w http.ResponseWriter, r *http.Request) {
	//1.解析参数
	r.ParseForm()

	upid := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")
	//chunkCount, _ := strconv.Atoi(r.Form.Get("chunkCount"))

	fmt.Printf("upid:%s  username:%s  fliehash:%s  filesize:%s   filename:%s\n", upid, username, filehash, filesize, filename)

	//2.获得 redis 连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	//3.通过  uploadid 查询 redis 并判断是否所有分块上传完成

	data, err := redis.Values(rConn.Do("HGETALL", hSetKeyPrefix+upid))
	if err != nil {
		log.Error(err.Error())
		respMsg := util.RespMsg{Code: -1, Msg: "complete upload failed", Data: err.Error()}
		w.Write(respMsg.JsonToBytes())
		return
	}

	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		fmt.Printf("k:%s   v:%s\n", k, v)
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}

	if totalCount != chunkCount {
		respMsg := util.NewRespMsg(-2, "invalid request", nil)
		w.Write(respMsg.JsonToBytes())
		return
	}

	//4.合并分块
	fpath := util.GetCurrentFielParentPath() + "/tmp/" + upid + "/"

	resultFile := fpath + filename
	fil, err := os.OpenFile(resultFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic(err)
		return
	}

	for i := 1; i <= chunkCount; i++ {
		fname := fpath + strconv.Itoa(i)
		f, err := os.OpenFile(fname, os.O_RDONLY, os.ModePerm)
		if err != nil {
			fmt.Printf("打开文件「%s」失败:%s", fname, err.Error())
		}

		bytes, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Printf("读取数据失败：%s", err.Error())
		}
		fil.Write(bytes)
		f.Close()
	}
	//写入完成，删除分块文件
	for i := 1; i <= chunkCount; i++ {
		fname := fpath + strconv.Itoa(i)
		err := os.Remove(fname)
		if err != nil {
			fmt.Printf("分块文件「%s」删除失败：%s", fname, err.Error())
		}
	}
	defer fil.Close()

	//5.更新唯一文件表及用户文件表
	fsize, _ := strconv.Atoi(filesize)
	fileUploadFinished := dblayer.OnFileUploadFinished(filehash, filename, int64(fsize), "")
	userFiledUploadFinished := dblayer.OnUserFiledUploadFinished(username, filehash, filename, int64(fsize))

	fmt.Printf("fileUploadFinished:%t   userFiledUploadFinished:%t", fileUploadFinished, userFiledUploadFinished)
	//6.响应处理结果
	respMsg := util.NewRespMsg(0, "OK", nil)
	w.Write(respMsg.JsonToBytes())
}
