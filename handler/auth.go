package handler

import (
	"filestore-server/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

func HTTPInterceptor() gin.HandlerFunc {

	return func(c *gin.Context) {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")

		if len(username) < 3 || !IsTokenValid(token) {
			c.JSON(http.StatusUnauthorized, util.RespMsg{Code: -1, Msg: "访问未授权"})
			// 验证不通过，不在调用后续处理函数。这里其实就是直接将 handler 列表的中正在执行位置 index 移动到最后一位
			c.Abort()
			// return 可省略，只要前面执行 Abort()  就可以让后面的 handler 函数不在执行
			return
		}
		c.Next()
	}

}
