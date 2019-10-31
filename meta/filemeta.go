package meta

import (
	myDb "filestore-server/db"
	"log"
	"sort"
)

// FileMeta: 文件元信息结构
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var fileMetas map[string]FileMeta

/*
	程序首次运行的时候执行一次
	https://zhuanlan.zhihu.com/p/34211611
	init函数的主要特点：
		init函数先于main函数自动执行，不能被其他函数调用；
		init函数没有输入参数、返回值；
		每个包可以有多个init函数；
		包的每个源文件也可以有多个init函数，这点比较特殊；
		同一个包的init执行顺序，golang没有明确定义，编程时要注意程序不要依赖这个执行顺序。
		不同包的init函数按照包导入的依赖关系决定执行顺序。
*/

//func init() {
//	fileMetas = make(map[string]FileMeta)
//}

func Setup() {
	fileMetas = make(map[string]FileMeta)
}

// UpdateFileMeta: 新增/更新 文件元信息
func UpdateFileMeta(fmeta FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}

// UpdateFileMetaDB: 新增/更新 文件元信息到 mysql 中
func UpdateFileMetaDB(fmeta FileMeta) bool {
	return myDb.OnFileUploadFinished(
		fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.Location,
	)
}

// UpdateUserFileMetaDB: 重命名用户文件名
func UpdateUserFileNameDB(userName, newFileName string, fmeta FileMeta) bool {
	return myDb.UpdateUserFileName(userName, newFileName, fmeta.FileSha1, fmeta.FileName, fmeta.FileSize)
}

// GetFileMeta: 通过 sha1 值获取文件的元信息
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

// GetFileMetaDB: 从 mysql 通过 sha1 值获取文件的元信息
func GetFileMetaDB(fileSha1 string) (*FileMeta, error) {
	tableFile, e := myDb.GetFileMeta(fileSha1)
	if e != nil {
		return nil, e
	}
	fmeta := FileMeta{
		FileSha1: tableFile.FileHash,
		FileName: tableFile.FileName.String,
		FileSize: tableFile.FileSize.Int64,
		Location: tableFile.FileAddr.String,
	}
	log.Println("fmeta 数据：")
	log.Println(fmeta)
	return &fmeta, nil
}

// GetLastFileMetas: 获取批量的文件元信息列表
func GetLastFileMetas(count int) []FileMeta {
	fMetaArray := make([]FileMeta, len(fileMetas))
	for _, V := range fileMetas {
		fMetaArray = append(fMetaArray, V)
	}
	sort.Sort(ByUploadTime(fMetaArray))
	return fMetaArray[0:count]
}

// RemoveFileMeta: 简单删除元信息
func RemoveFileMeta(fileSha1 string) {
	delete(fileMetas, fileSha1)
}
