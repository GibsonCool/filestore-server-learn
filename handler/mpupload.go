package handler

import (
	rPool "filestore-server/cache/redis"
	dblayer "filestore-server/db"
	"filestore-server/util"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
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
func InitMultipartUploadHandler(c *gin.Context) {
	//1.解析用户请求参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filesize, err := strconv.Atoi(c.Request.FormValue("filesize"))
	if err != nil {
		c.JSON(http.StatusOK, util.RespMsg{Code: -1, Msg: "params invalid"})
		return
	}

	//2.获得 redis 的一个链接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	resultCount := int(math.Ceil(float64(filesize) / float64(chunkSize)))
	log.Printf("向上取整 filesize:%d  chunkSize:%d  result%d \n", filesize, chunkSize, resultCount)
	//3.生成分块上传的初始化信息
	upinfo := MultiparUploadInfo{
		FileHash:  filehash,
		FileSize:  filesize,
		UploadId:  username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize: chunkSize,
		// 文件总大小/分块大小  然后向上取整
		ChunkCount: resultCount,
	}

	//4.将初始化信息写入到 redis 缓存
	rConn.Do("HSET", hSetKeyPrefix+upinfo.UploadId, "chunkcount", upinfo.ChunkCount)
	rConn.Do("HSET", hSetKeyPrefix+upinfo.UploadId, "filehash", upinfo.FileHash)
	rConn.Do("HSET", hSetKeyPrefix+upinfo.UploadId, "filsize", upinfo.FileSize)

	//5.将响应初始化数据返回到客户端
	c.JSON(http.StatusOK, util.RespMsg{Code: 0, Msg: "OK", Data: upinfo})
}

// UploadPartHandler：上传文件分块
func UploadPartHandler(c *gin.Context) {
	//1.解析用户请求参数
	//username := r.Form.Get("username")
	uploadId := c.Request.FormValue("uploadid")
	chunkIndex := c.Request.FormValue("index")

	log.Printf("uploadid:%s  chunkIndex:%s\n", uploadId, chunkIndex)
	//2.获得 redis 的一个链接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	//3.获得文件句柄，用于存储分块内容
	fpath := util.GetCurrentFielParentPath() + "/tmp/" + uploadId + "/" + chunkIndex
	// 数字设定法：：0表示没有权限，1表示可执行权限，2表示可写权限，4表示可读权限，然后将其相加。设置当前用户可读可写可执行权限
	os.Mkdir(path.Dir(fpath), 0744)
	file, e := os.Create(fpath)
	if e != nil {
		c.JSON(http.StatusOK, util.RespMsg{Code: -1, Msg: "Upload part failed:" + e.Error()})
		return
	}
	defer file.Close()

	//读取内存中的分块内容写入到文件中
	buf := make([]byte, 1023*1024)
	for {
		n, e := c.Request.Body.Read(buf)
		file.Write(buf[:n])
		if e != nil {
			break
		}
	}

	//4.更新 redis 缓存状态
	rConn.Do("HSET", hSetKeyPrefix+uploadId, "chkidx_"+chunkIndex, 1)

	//5.返回结果给到客户端
	c.JSON(http.StatusOK, util.RespMsg{Code: 0, Msg: "OK", Data: nil})
}

// CompleteUploadHander:通知上传合并
func CompleteUploadHander(c *gin.Context) {
	//1.解析参数
	upid := c.Request.FormValue("uploadid")
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filesize := c.Request.FormValue("filesize")
	filename := c.Request.FormValue("filename")
	//chunkCount, _ := strconv.Atoi(r.Form.Get("chunkCount"))

	log.Printf("upid:%s  username:%s  fliehash:%s  filesize:%s   filename:%s\n", upid, username, filehash, filesize, filename)

	//2.获得 redis 连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	//3.通过  uploadid 查询 redis 并判断是否所有分块上传完成

	data, err := redis.Values(rConn.Do("HGETALL", hSetKeyPrefix+upid))
	if err != nil {
		log.Println(err.Error())
		respMsg := util.RespMsg{Code: -1, Msg: "complete upload failed", Data: err.Error()}
		c.JSON(http.StatusOK, respMsg)
		return
	}

	totalCount := 0
	chunkCount := 0

	// 这通过 HGETALL 得到的结果 key 和 value 都是一个 []interface 中。所以每次取出后要  +2
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		log.Printf("k:%s   v:%s\n", k, v)
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}

	if totalCount != chunkCount {
		log.Printf("totalCount:%d  chunkCount:%d\n", totalCount, chunkCount)
		respMsg := util.NewRespMsg(-2, "invalid request", nil)
		c.JSON(http.StatusOK, respMsg)
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
			log.Printf("打开文件「%s」失败:%s", fname, err.Error())
		}

		bytes, err := ioutil.ReadAll(f)
		if err != nil {
			log.Printf("读取数据失败：%s", err.Error())
		}
		fil.Write(bytes)
		f.Close()
	}
	//写入完成，删除分块文件
	for i := 1; i <= chunkCount; i++ {
		fname := fpath + strconv.Itoa(i)
		err := os.Remove(fname)
		if err != nil {
			log.Printf("分块文件「%s」删除失败：%s", fname, err.Error())
		}
	}
	defer fil.Close()

	//5.更新唯一文件表及用户文件表
	fsize, _ := strconv.Atoi(filesize)

	/*
		这里存在的问题，并没有使用事物来执行两张表的操作，就有可能一张表操作成功，宁一张表失败的问题。可以在定义一个方法将两个 sql 操作放到一个事物中

		// db, _ := sql.Open(...)
		tx, err := db.Begin() // 创建tx对象，事务开始
		tx.Exec("写文件表sql")
		tx.Exec("写用户文件表sql")
		err := tx.Commit()    // 事务提交，要么都成功，要么都失败
		if err != nil {
		    // commit失败, 事务回滚
		    tx.Rollback()
		}

	*/
	fileUploadFinished := dblayer.OnFileUploadFinished(filehash, filename, int64(fsize), "")
	userFiledUploadFinished := dblayer.OnUserFiledUploadFinished(username, filehash, filename, int64(fsize))

	log.Printf("fileUploadFinished:%t   userFiledUploadFinished:%t", fileUploadFinished, userFiledUploadFinished)
	//6.响应处理结果
	respMsg := util.NewRespMsg(0, "OK", nil)
	c.JSON(http.StatusOK, respMsg)
}
