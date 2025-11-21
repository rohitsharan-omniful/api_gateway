package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/yourorg/go-service-kit/pkg/logging"
	"github.com/yourorg/go-service-kit/pkg/middleware"
)

func TestTracingMiddleware(t *testing.T) {
	// Setup
	logger, _ := logging.NewLogger("info", "json")
	gin.SetMode(gin.TestMode)

	// Create router with tracing middleware
	router := gin.New()
	router.Use(middleware.TracingMiddleware(logger, "api_gateway"))
	router.Use(middleware.ServiceRequestIDMiddleware("X-Request-ID"))
	router.GET("/test", func(c *gin.Context) {
		traceID := middleware.GetTraceIDFromGin(c)
		requestID := middleware.GetRequestIDFromGin(c)
		c.JSON(http.StatusOK, gin.H{
			"trace_id":   traceID,
			"request_id": requestID,
		})
	})

	// Test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Trace-ID"))
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestTracingMiddlewareWithExistingTraceID(t *testing.T) {
	// Setup
	logger, _ := logging.NewLogger("info", "json")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.TracingMiddleware(logger, "api_gateway"))
	router.GET("/test", func(c *gin.Context) {
		traceID := middleware.GetTraceIDFromGin(c)
		c.JSON(http.StatusOK, gin.H{"trace_id": traceID})
	})

	// Test request with existing trace ID
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Trace-ID", "existing-trace-id")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "existing-trace-id", w.Header().Get("X-Trace-ID"))
}
