package db

import (
	myDb "filestore-server/db/mysql"
	"fmt"
)

// UserSignUp: 通过用户名以及密码完成 user 表的注册操作
func UserSignUp(username string, passwd string) bool {
	stmt, e := myDb.DBConn().Prepare(" insert ignore into tbl_user (`user_name`,`user_pwd`) values (?,?)")
	if e != nil {
		fmt.Println("failed to insert user,err:" + e.Error())
		return false
	}
	defer stmt.Close()

	result, e := stmt.Exec(username, passwd)
	if e != nil {
		fmt.Println("Failed to inser user, err:" + e.Error())
		return false
	}

	// 多一步判断是否重复注册。执行了 sql 并没有插入数据
	if rowsAffected, e := result.RowsAffected(); e == nil && rowsAffected > 0 {
		return true
	}

	return false
}
