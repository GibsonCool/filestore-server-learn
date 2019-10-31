package handler

import (
	dblayer "filestore-server/db"
	"filestore-server/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

const (
	// 密码加盐值
	pwdSalt   = "*#890"
	tokenSalt = "_tokensalt"
)

// SignupHandler: 处理用户注册请求
func SignupHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "http://"+c.Request.Host+"/static/view/signup.html")
}

func DoSignupHandler(c *gin.Context) {
	//解析校验参数的有效性
	username := c.Request.FormValue("username")
	passwd := c.Request.FormValue("password")

	if len(username) < 3 || len(passwd) < 5 {
		c.JSON(http.StatusOK, util.RespMsg{
			Code: -1,
			Msg:  "Invalid parameter: 用户名或密码不符合规范",
		})
		return
	}

	//3.用户密码加盐处理
	encPasswd := util.Sha1([]byte(pwdSalt + passwd))
	//4.存入数据库 tbl_user 表并返回结果
	isSuccess := dblayer.UserSignUp(username, encPasswd)

	if isSuccess {
		c.JSON(http.StatusOK, util.RespMsg{
			Code: 0,
			Msg:  "注册成功",
			Data: "/static/view/signin.html",
		})
	} else {
		c.JSON(http.StatusOK, util.RespMsg{
			Code: -1,
			Msg:  "注册失败",
		})
	}
}

// SignInHandler: 登录接口
func SignInHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "http://"+c.Request.Host+"/static/view/signin.html")
}

func DoSignInHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	passwd := c.Request.FormValue("password")
	encPasswd := util.Sha1([]byte(pwdSalt + passwd))

	//1.校验用户名及密码
	pwdChecked := dblayer.UserSignIn(username, encPasswd)
	if !pwdChecked {
		c.JSON(http.StatusOK, util.RespMsg{
			Code: -1,
			Msg:  "Failed 用户名或密码错误",
		})
		return
	}

	//2.生成访问凭证（一般两种方式：① token   ② cookies/session浏览器端比较常见）这里选择第一种
	token := GenToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		c.JSON(http.StatusInternalServerError, util.RespMsg{
			Code: -2,
			Msg:  "update token failed",
		})
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
			"http://" + c.Request.Host + "/static/view/home.html",
			username,
			token,
		},
	}

	//log.Println(resp.Data)
	c.JSON(http.StatusOK, resp)
}

// UserInfoHandler: 查询用户信息
func UserInfoHandler(c *gin.Context) {
	username := c.Request.FormValue("username")

	// 3. 查询用户信息
	user, e := dblayer.GetUserInfo(username)
	if e != nil {
		log.Println(e.Error())
		c.JSON(http.StatusForbidden,
			gin.H{})
		return
	}

	// 4. 组装并且响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	c.Data(http.StatusOK, "octet-stream", resp.JsonToBytes())
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
