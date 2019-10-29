package util

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

/*
	工具类：
		根据文件流或者文件计算 hash 值
		获取文件大小
		判断文件是否存在
*/
var CstZone = time.FixedZone("CST", 8*3600) // 东八

type Sha1Stream struct {
	_sha1 hash.Hash
}

func (obj *Sha1Stream) Update(data []byte) {
	if obj._sha1 == nil {
		obj._sha1 = sha1.New()
	}
	obj._sha1.Write(data)
}

func (obj *Sha1Stream) Sum() string {
	return hex.EncodeToString(obj._sha1.Sum([]byte("")))
}

func Sha1(data []byte) string {
	_sha1 := sha1.New()
	_sha1.Write(data)
	return hex.EncodeToString(_sha1.Sum([]byte("")))
}

func FileSha1(file *os.File) string {
	_sha1 := sha1.New()
	io.Copy(_sha1, file)
	return hex.EncodeToString(_sha1.Sum(nil))
}

func MD5(data []byte) string {
	_md5 := md5.New()
	_md5.Write(data)
	return hex.EncodeToString(_md5.Sum([]byte("")))
}

func FileMD5(file *os.File) string {
	_md5 := md5.New()
	io.Copy(_md5, file)
	return hex.EncodeToString(_md5.Sum(nil))
}

func PathExists(path string) (bool, error) {
	_, e := os.Stat(path)
	if e == nil {
		return true, nil
	}

	if os.IsNotExist(e) {
		return false, nil
	}
	return false, e
}

func GetFileSize(filename string) int64 {
	var result int64

	filepath.Walk(filename, func(path string, info os.FileInfo, err error) error {

		result = info.Size()
		return nil
	})
	return result
}

// getCurrentFilePath:获取当前执行文件绝对路径
func GetCurrentFilePath() string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		panic(" Can not get current file info")
	}
	lastIndex := strings.LastIndex(file, "/") + 1
	file = file[:lastIndex]
	return file
}

func GetCurrentFielParentPath() string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		panic(" Can not get current file info")
	}
	lastIndex := strings.LastIndex(file, "/")
	file = file[:lastIndex]
	lastIndex = strings.LastIndex(file, "/") + 1
	parentPath := file[:lastIndex]
	log.Println(parentPath)
	return parentPath
}
