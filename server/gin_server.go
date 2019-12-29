package server

import (
	"github.com/gin-gonic/gin"
	"histore/order/job/xxl"
	"net/http"
)

func GinSupport() *gin.Engine {
	ginServer := gin.New()
	setXxlClientRoute(ginServer)
	return ginServer
}

func setXxlClientRoute(r *gin.Engine) {
	r.POST("/", func(c *gin.Context) {
		buf := make([]byte, 1024)
		int, _ := c.Request.Body.Read(buf)
		if int == 0 {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		resBy, err := xxl.RequestHandler(buf)
		if err != nil || resBy == nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write(resBy)
	})
}
