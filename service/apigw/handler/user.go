package handler

import (
	"context"
	"filestore-server/config"
	"filestore-server/service/account/proto"
	"filestore-server/util"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro"
	"github.com/micro/go-plugins/registry/consul"
	"log"
	"net/http"
)

var (
	userCli proto.UserService
)

func init() {

	service := micro.NewService(micro.Registry(consul.NewRegistry()))

	service.Init()

	userCli = proto.NewUserService(config.MicroServiceUserName, service.Client())
}

// SignupHandler: 处理用户注册请求
func SignupHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "http://"+c.Request.Host+"/static/view/signup.html")
}

func DoSignupHandler(c *gin.Context) {
	//解析校验参数的有效性
	username := c.Request.FormValue("username")
	passwd := c.Request.FormValue("password")

	resp, e := userCli.Signup(context.TODO(), &proto.ReqSignup{
		Username:  username,
		Passsword: passwd,
	})

	if e != nil {
		log.Println("rpc 远程调用返回错误：" + e.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, util.RespMsg{
		Code: int(resp.Code),
		Msg:  resp.Message,
	})
}
