package db

import (
	myDb "filestore-server/db/mysql"
	"log"
)

// User : 用户表model
type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

// UserSignUp: 通过用户名以及密码完成 user 表的注册操作
func UserSignUp(username string, passwd string) bool {
	stmt, e := myDb.DBConn().Prepare(" insert ignore into tbl_user (`user_name`,`user_pwd`) values (?,?)")
	if e != nil {
		log.Println("failed to insert user,err:" + e.Error())
		return false
	}
	defer stmt.Close()

	result, e := stmt.Exec(username, passwd)
	if e != nil {
		log.Println("Failed to insert user, err:" + e.Error())
		return false
	}

	// 多一步判断是否重复注册。执行了 sql 并没有插入数据
	if rowsAffected, e := result.RowsAffected(); e == nil && rowsAffected > 0 {
		return true
	}

	return false
}

// UserSignIn: 判断密码是否一致
func UserSignIn(username string, encpwd string) bool {
	stmt, e := myDb.DBConn().Prepare("select * from tbl_user where user_name=? limit 1")
	if e != nil {
		log.Println("failed to query user,err:" + e.Error())
		return false
	}

	rows, e := stmt.Query(username)
	if e != nil {
		log.Println("Failed to query user by name:" + username + ",err:" + e.Error())
		return false
	} else if rows == nil {
		log.Println("username not found:" + username)
		return false
	}

	// 如果查询到了用户数据
	parseRows := myDb.ParseRows(rows)
	if len(parseRows) > 0 && string(parseRows[0]["user_pwd"].([]byte)) == encpwd {
		return true
	}
	return false
}

// UpdateToken: 刷新用户登录的token
func UpdateToken(username string, token string) bool {
	stmt, e := myDb.DBConn().Prepare("replace into tbl_user_token (`user_name`,`user_token`) values (?,?)")
	if e != nil {
		log.Println(e.Error())
		return false
	}
	defer stmt.Close()

	_, e = stmt.Exec(username, token)
	if e != nil {
		log.Println(e.Error())
		return false
	}

	return true
}

// GetUserInfo:用户信息查询
func GetUserInfo(username string) (User, error) {
	user := User{}

	stmt, err := myDb.DBConn().Prepare(
		"select user_name,signup_at from tbl_user WHERE user_name=? limit 1")
	if err != nil {
		log.Println(err.Error())
		return user, err
	}
	defer stmt.Close()

	// 执行查询的操作
	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		return user, err
	}
	return user, nil

}
