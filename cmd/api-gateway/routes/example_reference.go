package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/go-service-kit/pkg/httpservice"
	"github.com/yourorg/go-service-kit/pkg/logging"
)

func init() {
	Register(func(r *gin.Engine) {
		r.GET("/example/logging", ExampleReferenceHandler)
	})
}

// ExampleReferenceHandler demonstrates proper logging practices using centralized utilities.
func ExampleReferenceHandler(c *gin.Context) {
	// 1. Log at the appropriate level using one-liners
	// INFO: Significant events in the normal flow
	httpservice.LogInfo(c, "Starting example reference handler",
		logging.NewField("user_action", "view_reference"),
		logging.NewField("endpoint", "/example/logging"),
	)

	// Simulate some work
	processTime := 50 * time.Millisecond

	// DEBUG: Granular details useful for debugging
	httpservice.LogDebug(c, "Processing request details",
		logging.NewField("process_time_ms", processTime.Milliseconds()),
		logging.NewField("step", "validation"),
	)

	// Simulate a warning condition
	// WARN: Something unexpected but handled
	if processTime > 20*time.Millisecond {
		httpservice.LogWarn(c, "Processing took longer than expected",
			logging.NewField("actual_duration_ms", processTime.Milliseconds()),
			logging.NewField("threshold_ms", 20),
		)
	}

	// Simulate an error condition (optional, for demonstration)
	// ERROR: Operation failed
	// err := errors.New("simulated error")
	// httpservice.RespondErrorWithLog(c, "Failed to process item", err,
	// 	logging.NewField("item_id", "12345"),
	// )

	// 3. Return a standardized response
	httpservice.RespondSuccess(c, gin.H{
		"message": "Check the logs to see the structured output! (Now using centralized helpers)",
		"guide": []string{
			"LogInfo: Normal flow events",
			"LogWarn: Unexpected but handled issues",
			"RespondErrorWithLog: Log error and return response",
			"RespondSuccess: Standard success response",
		},
	})
}
