package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/yourorg/api-gateway/cmd/api-gateway/handlers"
	"github.com/yourorg/go-service-kit/pkg/httpservice"
)

func init() {
	RegisterProtected(func(r gin.IRouter) {
		// Example wrapped handlers - zero boilerplate, automatic logging (JWT protected)
		r.POST("/orders", httpservice.Wrap("CreateOrder", handlers.CreateOrderHandler))
		r.GET("/example/error", httpservice.Wrap("ErrorExample", handlers.ErrorExampleHandler))
		r.GET("/items/:id", httpservice.Wrap("ComplexBusinessLogic", handlers.ComplexBusinessLogicHandler))
	})
}
