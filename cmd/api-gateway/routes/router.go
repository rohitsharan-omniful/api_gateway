package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/yourorg/api-gateway/pkg/config"
	"github.com/yourorg/go-service-kit/pkg/jwt"
	"github.com/yourorg/go-service-kit/pkg/logging"
	gokitmiddleware "github.com/yourorg/go-service-kit/pkg/middleware"
)

var registrars []func(*gin.Engine)
var protectedRegistrars []func(gin.IRouter)

// Register adds a new route registrar to the list (public routes).
func Register(r func(*gin.Engine)) {
	registrars = append(registrars, r)
}

// RegisterProtected adds a new route registrar for protected routes (requires JWT).
func RegisterProtected(r func(gin.IRouter)) {
	protectedRegistrars = append(protectedRegistrars, r)
}

// Init initializes the router with middleware and registers all routes.
func Init(router *gin.Engine, logger logging.Logger, cfg *config.GatewayConfig, nrClient gokitmiddleware.TelemetryClient, slackClient gokitmiddleware.SlackClient, jwtService *jwt.JWTService) {
	serviceName := "api_gateway" // This could be passed as argument or constant

	// Wire middleware chain (order matters!)
	// 1. Tracing - generates/extracts trace ID (gateway creates it)
	router.Use(gokitmiddleware.TracingMiddleware(logger, serviceName))

	// 2. Gateway Request ID - generates gateway request ID (different from service request ID)
	router.Use(gokitmiddleware.ServiceRequestIDMiddleware("X-Request-ID"))

	// 3. Context Logger - attaches contextual logger to request context
	router.Use(gokitmiddleware.ContextLoggerMiddleware(logger, serviceName))

	// 4. Error Handler - centralized error handling
	router.Use(gokitmiddleware.ErrorHandlerMiddleware(logger))

	// 5. Slow Request Detector - detects slow requests and triggers alerts (also handles 5xx errors)
	router.Use(gokitmiddleware.SlowRequestMiddleware(
		cfg.Gateway.SlowMs,
		nrClient,    // Implements TelemetryClient interface
		slackClient, // Implements SlackClient interface
		logger,
	))

	// Register public routes (no JWT required)
	for _, registrar := range registrars {
		registrar(router)
	}

	// Create protected route group with JWT middleware
	protected := router.Group("/")
	protected.Use(jwt.JWTMiddleware(jwtService, logger))
	{
		// Register protected routes
		for _, registrar := range protectedRegistrars {
			registrar(protected)
		}
	}
}
