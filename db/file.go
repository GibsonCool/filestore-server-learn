package db

import (
	"database/sql"
	myDb "filestore-server/db/mysql"
	"fmt"
)

/*
	预编译语句(PreparedStatement)提供了诸多好处, 因此我们在开发中尽量使用它. 下面列出了使用预编译语句所提供的功能:

		PreparedStatement 可以实现自定义参数的查询
		PreparedStatement 通常来说, 比手动拼接字符串 SQL 语句高效.
		PreparedStatement 可以防止SQL注入攻击

	一般用Prepared Statements和Exec()完成INSERT, UPDATE, DELETE操作。
*/

// OnFileUploadFinished: 文件上传完成，保存 FileMeta 数据
func OnFileUploadFinished(filehash string, filename string, filesize int64, fileaddr string) bool {
	stmt, err := myDb.DBConn().Prepare(
		"insert ignore into tbl_file (`file_sha1`,`file_name`,`file_size`,`file_addr`,`status`) values (?,?,?,?,1)")

	if err != nil {
		fmt.Println("Failed to prepare statement ,err :", err.Error())
	}

	defer stmt.Close()

	result, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		fmt.Println("插入数据失败,err:", err.Error())
		return false
	}

	if rowsAffected, err := result.RowsAffected(); err == nil {
		// 到这里说明 sql 执行成功了，但是还需要判断下文件是否已经存在，是否有数据插入 sql
		if rowsAffected <= 0 {
			fmt.Printf("File with hash:%s has been uploaded before", filehash)
		}
		return true
	}

	return false
}

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// GetFileMeta: 从 mysql 获取文件元信息
func GetFileMeta(filehash string) (*TableFile, error) {
	stmt, e := myDb.DBConn().Prepare("select file_sha1,file_addr,file_name,file_size from tbl_file where file_sha1=? and status =1 limit 1")
	if e != nil {
		fmt.Println(e.Error())
		return nil, e
	}
	defer stmt.Close()

	tfile := TableFile{}
	e = stmt.QueryRow(filehash).Scan(&tfile.FileHash, &tfile.FileAddr, &tfile.FileName, &tfile.FileSize)
	if e != nil {
		fmt.Println(e.Error())
		return nil, e
	}

	return &tfile, nil

}
