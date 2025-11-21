package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/yourorg/go-service-kit/pkg/errors"
	"github.com/yourorg/go-service-kit/pkg/httpservice"
)

func init() {
	Register(func(r *gin.Engine) {
		// Test route that returns 500 error to test Slack alerting
		r.GET("/test/error", httpservice.Wrap("TestError", func(c *gin.Context) error {
			// Return a 500 Internal Server Error
			// This will trigger the ErrorHandlerMiddleware which should send a Slack alert
			return errors.NewInternalError("Test error for Slack alert - this is intentional")
		}))

		// Test route that returns 500 with custom message
		r.GET("/test/error/:message", httpservice.Wrap("TestErrorWithMessage", func(c *gin.Context) error {
			message := c.Param("message")
			return errors.NewInternalError("Test error: " + message)
		}))
	})
}
