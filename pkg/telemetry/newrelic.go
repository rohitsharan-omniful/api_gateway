package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/yourorg/go-service-kit/pkg/logging"
)

const ServiceName = "api_gateway"

// NewRelicClient wraps the New Relic agent.
type NewRelicClient struct {
	app    *newrelic.Application
	logger logging.Logger
	enabled bool
}

// NewRelicConfig holds New Relic configuration.
type NewRelicConfig struct {
	LicenseKey string
	AppName    string
	Enabled    bool
}

// NewNewRelicClient creates a new New Relic client.
func NewNewRelicClient(cfg NewRelicConfig, logger logging.Logger) (*NewRelicClient, error) {
	if !cfg.Enabled || cfg.LicenseKey == "" {
		logger.Info("New Relic disabled or license key not provided")
		return &NewRelicClient{
			enabled: false,
			logger:  logger,
		}, nil
	}

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(cfg.AppName),
		newrelic.ConfigLicense(cfg.LicenseKey),
		newrelic.ConfigDistributedTracerEnabled(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create New Relic application: %w", err)
	}

	logger.Info("New Relic client initialized", logging.NewField("app_name", cfg.AppName))

	return &NewRelicClient{
		app:    app,
		logger: logger,
		enabled: true,
	}, nil
}

// RecordTransaction records a transaction in New Relic.
func (n *NewRelicClient) RecordTransaction(ctx context.Context, name string, durationMs int64, statusCode int, traceID, requestID string) {
	if !n.enabled || n.app == nil {
		return
	}

	txn := newrelic.FromContext(ctx)
	if txn == nil {
		txn = n.app.StartTransaction(name)
		defer txn.End()
	}

	txn.SetName(name)
	txn.AddAttribute("trace_id", traceID)
	txn.AddAttribute("request_id", requestID)
	txn.AddAttribute("status_code", statusCode)
	txn.AddAttribute("duration_ms", durationMs)
	txn.AddAttribute("service", ServiceName)

	if statusCode >= 500 {
		txn.NoticeError(fmt.Errorf("HTTP %d", statusCode))
	}
}

// RecordCustomEvent records a custom event in New Relic.
func (n *NewRelicClient) RecordCustomEvent(eventType string, attributes map[string]interface{}) {
	if !n.enabled || n.app == nil {
		return
	}

	n.app.RecordCustomEvent(eventType, attributes)
}

// RecordSlowRequest records a slow request event.
// Implements middleware.TelemetryClient interface.
func (n *NewRelicClient) RecordSlowRequest(ctx interface{}, path string, durationMs int64, traceID, requestID string) {
	if !n.enabled {
		return
	}

	// Extract downstream service from path (e.g., /hrms/* -> hrms-core)
	downstreamService := extractDownstreamServiceFromPath(path)

	attributes := map[string]interface{}{
		"event_type":        "SlowRequest",
		"service":           ServiceName,
		"path":              path,
		"downstream_service": downstreamService,
		"duration_ms":      durationMs,
		"trace_id":         traceID,
		"request_id":       requestID,
	}

	n.RecordCustomEvent("SlowRequest", attributes)
	
	// Convert ctx to context.Context if possible
	var ctxContext context.Context
	if c, ok := ctx.(context.Context); ok {
		ctxContext = c
		n.RecordTransaction(ctxContext, path, durationMs, 200, traceID, requestID)
	}
	
	n.logger.Warn("Slow request detected",
		logging.NewField("path", path),
		logging.NewField("downstream_service", downstreamService),
		logging.NewField("duration_ms", durationMs),
		logging.NewField("trace_id", traceID),
		logging.NewField("request_id", requestID),
	)
}

// extractDownstreamServiceFromPath extracts the downstream service name from the path.
func extractDownstreamServiceFromPath(path string) string {
	if len(path) > 1 {
		parts := path[1:] // Remove leading /
		for i, char := range parts {
			if char == '/' {
				serviceName := parts[:i]
				switch serviceName {
				case "hrms", "employees":
					return "hrms-core"
				default:
					return serviceName
				}
			}
		}
		if len(parts) > 0 {
			return parts
		}
	}
	return "unknown"
}

// RecordError records an error event.
// Implements middleware.TelemetryClient interface.
func (n *NewRelicClient) RecordError(ctx interface{}, path, errorMsg string, statusCode int, traceID, requestID string) {
	if !n.enabled {
		return
	}

	// Extract downstream service from path
	downstreamService := extractDownstreamServiceFromPath(path)

	attributes := map[string]interface{}{
		"event_type":        "GatewayError",
		"service":           ServiceName,
		"path":              path,
		"downstream_service": downstreamService,
		"error":            errorMsg,
		"status_code":      statusCode,
		"trace_id":         traceID,
		"request_id":       requestID,
	}

	n.RecordCustomEvent("GatewayError", attributes)
	
	// Convert ctx to context.Context if possible
	var ctxContext context.Context
	if c, ok := ctx.(context.Context); ok {
		ctxContext = c
		n.RecordTransaction(ctxContext, path, 0, statusCode, traceID, requestID)
	}
	
	n.logger.Error("Gateway error detected",
		logging.NewField("path", path),
		logging.NewField("downstream_service", downstreamService),
		logging.NewField("error", errorMsg),
		logging.NewField("status_code", statusCode),
		logging.NewField("trace_id", traceID),
		logging.NewField("request_id", requestID),
	)
}

// Shutdown gracefully shuts down the New Relic client.
func (n *NewRelicClient) Shutdown(timeoutMs int) {
	if n.enabled && n.app != nil {
		n.app.Shutdown(time.Duration(timeoutMs) * time.Millisecond)
	}
}

