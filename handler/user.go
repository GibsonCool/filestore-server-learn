package handler

import (
	dblayer "filestore-server/db"
	"filestore-server/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	// 密码加盐值
	pwdSalt   = "*#890"
	tokenSalt = "_tokensalt"
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
	isSuccess := dblayer.UserSignUp(username, encPasswd)

	if isSuccess {
		w.Write([]byte("Success 注册成功"))
	} else {
		w.Write([]byte("Failed 注册失败"))
	}
}

// SignInHandler: 登录接口
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	username := r.Form.Get("username")
	passwd := r.Form.Get("password")
	encPasswd := util.Sha1([]byte(pwdSalt + passwd))

	//1.校验用户名及密码
	pwdChecked := dblayer.UserSignIn(username, encPasswd)
	if !pwdChecked {
		w.Write([]byte("Failed 用户名或密码错误"))
		return
	}

	//2.生成访问凭证（一般两种方式：① token   ② cookies/session浏览器端比较常见）这里选择第一种
	token := GenToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		w.Write([]byte("update token failed"))
		return
	}

	//3.登录成功后重定向到首页
	//w.Write([]byte("http://" + r.Host + "/static/view/home.html"))
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			"http://" + r.Host + "/static/view/home.html",
			username,
			token,
		},
	}
	w.Write(resp.JsonToBytes())
}

// GenToken: 生成用户 token
func GenToken(username string) string {
	//token(40位字符 mde5 后得到的32位字符再加上截图时间戳前8位）生成规则：md5(username+timestamp+tokenSalt)+timestamp[:8]

	ts := fmt.Sprint("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + tokenSalt))
	return tokenPrefix + ts[:8]
}
