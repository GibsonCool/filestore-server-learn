package db

import (
	mydb "filestore-server/db/mysql"
	"fmt"
	"time"
)

type UserFile struct {
	UserName    string
	FileHash    string
	FileName    string
	FileSize    int64
	UploadAt    string
	LastUpdated string
}

// OnUserFiledUploadFinished: 更新用户文件表
func OnUserFiledUploadFinished(username, filehash, filename string, filesize int64) bool {
	stmt, e := mydb.DBConn().Prepare(
		"insert ignore into tbl_user_file (`user_name`,`file_sha1`,`file_name`,`file_size`,`upload_at`) values (?,?,?,?,?)")

	if e != nil {
		fmt.Println(e.Error())
		return false
	}
	defer stmt.Close()
	_, e = stmt.Exec(username, filehash, filename, filesize, time.Now())

	if e != nil {
		return false
	}
	return true
}
