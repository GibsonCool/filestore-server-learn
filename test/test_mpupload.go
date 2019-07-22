package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/gpmgo/gopm/modules/log"
	jsoniter "github.com/json-iterator/go"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// multipartUpload:模拟分块上传
func multipartUpload(filename string, targetURL string, chunkSize int) error {
	file, e := os.Open(filename)
	if e != nil {
		fmt.Println(e.Error())
		return e
	}
	defer file.Close()

	bufRedader := bufio.NewReader(file)
	index := 0

	ch := make(chan int)
	buf := make([]byte, chunkSize)

	for {
		n, e := bufRedader.Read(buf)
		if n <= 0 {
			break
		}
		index++

		bufCopied := make([]byte, 5*1024*1024)
		copy(bufCopied, buf)

		go func(b []byte, curIdx int) {
			fmt.Printf("upload_size:%d\n", len(b))

			response, err := http.Post(
				targetURL+"&index="+strconv.Itoa(curIdx),
				"multipart/form-data",
				bytes.NewBuffer(b),
			)
			if err != nil {
				fmt.Println(err.Error())
			}

			body, err := ioutil.ReadAll(response.Body)
			fmt.Printf("%+v   %+d\n", string(body), curIdx)
			response.Body.Close()

			ch <- curIdx
		}(bufCopied[:n], index)

		//遇到任何错误立即返回，并忽略 EOF 错误信息
		if e != nil {
			if e == io.EOF {
				break
			} else {
				fmt.Println(e.Error())
			}
		}
	}

	for idx := 0; idx < index; idx++ {
		select {
		case res := <-ch:
			fmt.Printf("已接受完：%d\n", res)
		}
	}

	return nil
}

/*
	测试分块上传
*/
func main() {
	username := "admin"
	token := "d2d123327dfc6e25d24a73da7bd6007b5d328b77"
	filehash := "c0b9096a93c3320ea576bab9144815b55ce000ee"

	// 1.请求初始化分块上传接口
	resp, err := http.PostForm("http://localhost:8080/file/mpupload/init",
		url.Values{
			"username": {username},
			"token":    {token},
			"filehash": {filehash},
			"filesize": {"268435456"},
		})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	//2.得到 uploadID 以及服务端指定的分块大小 chunkSize
	uploadID := jsoniter.Get(body, "data").Get("UploadId").ToString()
	chunkSize := jsoniter.Get(body, "data").Get("ChunkSize").ToInt()
	fmt.Printf("uploadid:%s   chunksize: %d\n", uploadID, chunkSize)

	//3.请求分块上传接口
	filename := "/Users/coulson/Downloads/test.zip"
	tURL := "http://localhost:8080/file/mpupload/uppart?" +
		"username=" + username + "&token=" + token + "&uploadid=" + uploadID
	err = multipartUpload(filename, tURL, chunkSize)

	if err != nil {
		log.Error(err.Error())
		//os.Exit(-1)
	}

	//4.请求分块完成接口
	response, err := http.PostForm("http://localhost:8080/file/mpupload/complete",
		url.Values{
			"username": {username},
			"token":    {token},
			"filehash": {filehash},
			"filesize": {"268435456"},
			"filename": {"test.zip"},
			"uploadid": {uploadID},
		})
	if err != nil {
		log.Error(err.Error())
		//os.Exit(-1)
	}
	defer response.Body.Close()

	all, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error(err.Error())
		//os.Exit(-1)
	}
	fmt.Sprintf("complete result:%s\n", string(all))

}
