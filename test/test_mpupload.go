package main

import (
	"bufio"
	"bytes"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// multipartUpload:模拟分块上传
func multipartUpload(filename string, targetURL string, chunkSize int) error {
	file, e := os.Open(filename)
	if e != nil {
		log.Println(e.Error())
		return e
	}
	defer file.Close()

	// 文件内容
	bufRedader := bufio.NewReader(file)
	index := 0

	ch := make(chan int)
	// 每次读取 chunkSize 大小内容
	buf := make([]byte, chunkSize)

	for {
		// 每次从文件中读取 buf 大小的内容n,下次会在已读取的位置后面开始在读取 n 的大小内容
		n, e := bufRedader.Read(buf)
		if n <= 0 {
			break
		}
		index++

		bufCopied := make([]byte, 5*1024*1024)
		copy(bufCopied, buf)

		go func(b []byte, curIdx int) {
			log.Printf("upload_size:%d\n", len(b))

			response, err := http.Post(
				targetURL+"&index="+strconv.Itoa(curIdx),
				"multipart/form-data",
				bytes.NewBuffer(b),
			)
			if err != nil {
				log.Println(err.Error())
			}

			body, err := ioutil.ReadAll(response.Body)
			log.Printf("%+v   %+d\n", string(body), curIdx)
			response.Body.Close()

			ch <- curIdx
		}(bufCopied[:n], index)

		//遇到任何错误立即返回，并忽略 EOF 错误信息
		if e != nil {
			if e == io.EOF {
				break
			} else {
				log.Println(e.Error())
			}
		}
	}

	for idx := 0; idx < index; idx++ {
		select {
		case res := <-ch:
			log.Printf("已接受完：%d\n", res)
		}
	}
	log.Println("分块上传完成~~~~~~~~~~")
	return nil
}

/*
	测试分块上传
*/
func main() {
	username := "test1"
	token := "15e6be7c2e0ede83193960b1aebc9db65db7a560"
	filehash := "12a54dddd54e6cb083c09a692d5c579c094a2e69"
	fileSize := "64549977"
	oldFileName := "1.mp4"

	// 1.请求初始化分块上传接口
	resp, err := http.PostForm("http://localhost:8080/file/mpupload/init",
		url.Values{
			"username": {username},
			"token":    {token},
			"filehash": {filehash},
			"filesize": {fileSize},
		})

	if err != nil {
		log.Println(err.Error())
		os.Exit(-1)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		os.Exit(-1)
	}

	//2.得到 uploadID 以及服务端指定的分块大小 chunkSize
	uploadID := jsoniter.Get(body, "data").Get("UploadId").ToString()
	chunkSize := jsoniter.Get(body, "data").Get("ChunkSize").ToInt()
	log.Printf("uploadid:%s   chunksize: %d\n", uploadID, chunkSize)

	//3.请求分块上传接口
	filename := "/Users/coulson/Downloads/" + oldFileName
	tURL := "http://localhost:8080/file/mpupload/uppart?" +
		"username=" + username + "&token=" + token + "&uploadid=" + uploadID
	err = multipartUpload(filename, tURL, chunkSize)

	if err != nil {
		log.Println(err.Error())
		//os.Exit(-1)
	}

	//4.请求分块完成接口
	response, err := http.PostForm("http://localhost:8080/file/mpupload/complete",
		url.Values{
			"username": {username},
			"token":    {token},
			"filehash": {filehash},
			"filesize": {fileSize},
			"filename": {oldFileName},
			"uploadid": {uploadID},
		})
	if err != nil {
		log.Println(err.Error())
		//os.Exit(-1)
	}
	defer response.Body.Close()

	all, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err.Error())
		//os.Exit(-1)
	}
	fmt.Sprintf("complete result:%s\n", string(all))

}
