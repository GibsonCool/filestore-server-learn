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

	//1.get请求直接返回注册页面内容
	if r.Method == http.MethodGet {
		bytes, e := ioutil.ReadFile("./static/view/signup.html")
		if e != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(bytes)
	}

	//2.解析校验参数的有效性
	r.ParseForm()

	username := r.Form.Get("username")
	passwd := r.Form.Get("password")

	if len(username) < 3 || len(passwd) < 5 {
		w.Write([]byte("Invalid parameter: 用户名或密码不符合规范"))
		return
	}

	//3.用户密码加盐处理
	encPasswd := util.Sha1([]byte(pwdSalt + passwd))
	//4.存入数据库 tbl_user 表并返回结果
	isSuccess := dblayer.UserSignUp(username, encPasswd)

	if isSuccess {
		w.Write([]byte("Success 注册成功"))
	} else {
		w.Write([]byte("Failed 注册失败"))
	}
}

// SignInHandler: 登录接口
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.Redirect(w, r, "/static/view/signin.html", http.StatusFound)
	}

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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//3.登录成功后重定向到首页 并组装返回 username,token 重定向url等信息
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

	//fmt.Println(resp.Data)
	w.Write(resp.JsonToBytes())
}

// UserInfoHandler: 查询用户信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析请求参数
	r.ParseForm()

	username := r.Form.Get("username")
	//token := r.Form.Get("token")

	// 2. 验证 token 是否有效			// token 校验逻辑使用同一拦截器 HTTPInterceptor 处理
	//isTokenValid := IsTokenValid(token)
	//if !isTokenValid {
	//	w.WriteHeader(http.StatusForbidden)
	//	return
	//}

	// 3. 查询用户信息
	user, e := dblayer.GetUserInfo(username)
	if e != nil {
		fmt.Println(e.Error())
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 4. 组装并且响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JsonToBytes())
}

// GenToken: 生成用户 token
func GenToken(username string) string {
	//token(40位字符 mde5 后得到的32位字符再加上截取时间戳前8位）生成规则：md5(username+timestamp+tokenSalt)+timestamp[:8]

	ts := fmt.Sprintf("%x", time.Now().In(util.CstZone).Unix())
	tokenPrefix := util.MD5([]byte(username + ts + tokenSalt))
	return tokenPrefix + ts[:8]
}

// IsTokenValid: token 是否有效
func IsTokenValid(token string) bool {
	if len(token) != 40 {
		return false
	}
	// TODO:判断 token 的时效性，是否过期

	// TODO:从数据库表 tbl_user_token 查询 username 对应的 token 信息

	// TODO: 对比两个 token 是否一致

	return true
}
