package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/yourorg/go-service-kit/pkg/httpservice"
)

func init() {
	Register(func(r *gin.Engine) {
		r.GET("/ping", func(c *gin.Context) {
			httpservice.SuccessResponse(c, gin.H{
				"message": "pong",
			})
		})
	})
}
