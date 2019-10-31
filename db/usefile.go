package db

import (
	mydb "filestore-server/db/mysql"
	"filestore-server/util"
	"log"
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
		log.Println(e.Error())
		return false
	}
	defer stmt.Close()
	_, e = stmt.Exec(username, filehash, filename, filesize, time.Now().In(util.CstZone))

	if e != nil {
		return false
	}
	return true
}

// QueryUserFileMetas: 批量获取用户文件信息
func QueryUserFileMetas(username string, limit int) ([]UserFile, error) {
	stmt, e := mydb.DBConn().Prepare(
		"select file_sha1,file_name,file_size,upload_at,last_update from tbl_user_file where user_name=? limit ?")
	if e != nil {
		log.Println(e.Error())
		return nil, e
	}

	defer stmt.Close()

	rows, e := stmt.Query(username, limit)
	if e != nil {
		log.Println(e.Error())
	}

	var userFiles []UserFile
	for rows.Next() {
		ufile := UserFile{}

		e := rows.Scan(&ufile.FileHash, &ufile.FileName, &ufile.FileSize, &ufile.UploadAt, &ufile.LastUpdated)

		if e != nil {
			log.Println(e.Error())
			break
		}

		userFiles = append(userFiles, ufile)

	}

	return userFiles, nil
}

// UpdateUserFileName：（重命名）修改用户文件表文件名称
func UpdateUserFileName(username, newFileName, fileSha1, fileName string, fileSize int64) bool {
	log.Println(username, newFileName, fileSha1, fileName, fileSize)
	stmt, e := mydb.DBConn().Prepare(
		"update tbl_user_file set file_name=? where user_name=? and file_sha1=? and file_name=? and file_size=?")

	if e != nil {
		log.Println(e.Error())
		return false
	}
	defer stmt.Close()
	_, e = stmt.Exec(newFileName, username, fileSha1, fileName, fileSize)

	if e != nil {
		return false
	}
	return true
}
