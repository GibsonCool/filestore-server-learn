package route

import (
	apigw "filestore-server/service/apigw/handler"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	router := gin.Default()
	// 处理静态资源
	router.Static("/static/", "./static")

	//不需要经过验证就能访问的接口
	router.GET("/user/signup", apigw.SignupHandler)
	router.POST("/user/signup", apigw.DoSignupHandler)

	return router
}
