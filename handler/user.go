package handler

import (
	delayed "filestore-server/db"
	"filestore-server/util"
	"io/ioutil"
	"net/http"
)

const (
	// 密码加盐值
	pwdSalt = "*#890"
)

// SignupHandler: 处理用户注册请求
func SignupHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		bytes, e := ioutil.ReadFile("./static/view/signup.html")
		if e != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(bytes)
	}

	r.ParseForm()

	username := r.Form.Get("username")
	passwd := r.Form.Get("password")

	if len(username) < 3 || len(passwd) < 5 {
		w.Write([]byte("Invalid parameter: 用户名或密码不符合规范"))
		return
	}

	encPasswd := util.Sha1([]byte(pwdSalt + passwd))
	isSuccess := delayed.UserSignUp(username, encPasswd)

	if isSuccess {
		w.Write([]byte("Success 注册成功"))
	} else {
		w.Write([]byte("Failed 注册失败"))
	}
}
