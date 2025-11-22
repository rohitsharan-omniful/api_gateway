package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/yourorg/api-gateway/cmd/api-gateway/handlers"
	"github.com/yourorg/go-service-kit/pkg/httpservice"
)

func init() {
	RegisterProtected(func(r gin.IRouter) {
		// User management endpoints using zero-boilerplate wrapper pattern (JWT protected)
		r.POST("/users", httpservice.Wrap("CreateUser", handlers.CreateUserHandler))
		r.GET("/users/:id", httpservice.Wrap("GetUser", handlers.GetUserHandler))
		r.PUT("/users/:id", httpservice.Wrap("UpdateUser", handlers.UpdateUserHandler))
		r.DELETE("/users/:id", httpservice.Wrap("DeleteUser", handlers.DeleteUserHandler))
	})
}
