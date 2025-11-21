package config

import (
	"strconv"
	"strings"
	"time"

	"github.com/yourorg/go-service-kit/pkg/config"
)

// GatewayConfig holds the API Gateway configuration.
type GatewayConfig struct {
	Gateway    GatewaySettings
	Downstream DownstreamConfig
	Telemetry  TelemetryConfig
	Retry      RetrySettings
}

// GatewaySettings holds gateway-specific settings.
type GatewaySettings struct {
	Port      int
	SlowMs    int64
	TimeoutMs int
	LogLevel  string
	LogFormat string
	// Security
	RateLimitRPS   float64
	RateLimitBurst int
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxBodySize    int64 // Maximum request body size in bytes
}

// DownstreamConfig holds downstream service configuration.
type DownstreamConfig struct {
	HRMSCoreBaseURL string
}

// TelemetryConfig holds telemetry configuration.
type TelemetryConfig struct {
	NewRelic NewRelicConfig
	Slack    SlackConfig
}

// NewRelicConfig holds New Relic configuration.
type NewRelicConfig struct {
	LicenseKey string
	AppName    string
	Enabled    bool
}

// SlackConfig holds Slack configuration.
type SlackConfig struct {
	WebhookURL string
	Channel    string
	Enabled    bool
}

// RetrySettings holds retry configuration.
type RetrySettings struct {
	MaxAttempts    int
	InitialDelayMs int
	MaxDelayMs     int
}

// LoadConfig loads configuration from environment variables and config.yaml.
func LoadConfig() (*GatewayConfig, error) {
	// Use LoadConfigFromFile which handles env override automatically
	// If file doesn't exist, fall back to env only
	var fileSource config.ConfigSource
	fs, err := config.NewFileConfigSource("config.yaml")
	if err != nil {
		// File doesn't exist, use env only
		fileSource = &config.EnvConfigSource{}
	} else {
		fileSource = fs
	}

	// Create composite source (env takes precedence)
	// We need to manually create it since sources field is not exported
	envSource := &config.EnvConfigSource{}
	source := &compositeSource{
		sources: []config.ConfigSource{envSource, fileSource},
	}

	cfg := &GatewayConfig{}

	// Gateway settings
	cfg.Gateway.Port = getInt(source, "GATEWAY_PORT", "gateway.port", 8080)
	cfg.Gateway.SlowMs = int64(getInt(source, "GATEWAY_SLOW_MS", "gateway.slow_ms", 1000))
	cfg.Gateway.TimeoutMs = getInt(source, "TIMEOUT_MS", "gateway.timeout_ms", 5000)
	cfg.Gateway.LogLevel = getString(source, "LOG_LEVEL", "gateway.log_level", "info")
	cfg.Gateway.LogFormat = getString(source, "LOG_FORMAT", "gateway.log_format", "json")

	// Security settings
	cfg.Gateway.RateLimitRPS = getFloat(source, "RATE_LIMIT_RPS", "gateway.rate_limit_rps", 10.0)
	cfg.Gateway.RateLimitBurst = getInt(source, "RATE_LIMIT_BURST", "gateway.rate_limit_burst", 20)
	cfg.Gateway.AllowedOrigins = getStringSlice(source, "CORS_ALLOWED_ORIGINS", "gateway.cors.allowed_origins", []string{"*"})
	cfg.Gateway.AllowedMethods = getStringSlice(source, "CORS_ALLOWED_METHODS", "gateway.cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	cfg.Gateway.AllowedHeaders = getStringSlice(source, "CORS_ALLOWED_HEADERS", "gateway.cors.allowed_headers", []string{"Content-Type", "Authorization"})
	cfg.Gateway.MaxBodySize = int64(getInt(source, "MAX_BODY_SIZE_MB", "gateway.max_body_size_mb", 10)) * 1024 * 1024 // Default 10MB

	// Downstream services
	cfg.Downstream.HRMSCoreBaseURL = getString(source, "HRMS_CORE_BASE_URL", "downstream.hrms_core_base_url", "http://localhost:8081")

	// Telemetry - New Relic
	cfg.Telemetry.NewRelic.LicenseKey = getString(source, "NEWRELIC_LICENSE_KEY", "telemetry.newrelic.license_key", "")
	cfg.Telemetry.NewRelic.AppName = getString(source, "NEWRELIC_APP_NAME", "telemetry.newrelic.app_name", "api-gateway")
	cfg.Telemetry.NewRelic.Enabled = getBool(source, "NEWRELIC_ENABLED", "telemetry.newrelic.enabled", true)

	// Telemetry - Slack
	cfg.Telemetry.Slack.WebhookURL = getString(source, "SLACK_WEBHOOK_URL", "telemetry.slack.webhook_url", "")
	cfg.Telemetry.Slack.Channel = getString(source, "SLACK_CHANNEL", "telemetry.slack.channel", "")
	cfg.Telemetry.Slack.Enabled = getBool(source, "SLACK_ENABLED", "telemetry.slack.enabled", true)

	// Retry settings
	cfg.Retry.MaxAttempts = getInt(source, "RETRY_MAX_ATTEMPTS", "retry.max_attempts", 3)
	cfg.Retry.InitialDelayMs = getInt(source, "RETRY_INITIAL_DELAY_MS", "retry.initial_delay_ms", 100)
	cfg.Retry.MaxDelayMs = getInt(source, "RETRY_MAX_DELAY_MS", "retry.max_delay_ms", 5000)

	return cfg, nil
}

// Helper functions
func getInt(source config.ConfigSource, envKey, fileKey string, defaultValue int) int {
	// Try env first
	if val, ok := source.Get(envKey); ok {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	// Try file
	if val, ok := source.Get(fileKey); ok {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getString(source config.ConfigSource, envKey, fileKey string, defaultValue string) string {
	// Try env first
	if val, ok := source.Get(envKey); ok {
		return val
	}
	// Try file
	if val, ok := source.Get(fileKey); ok {
		return val
	}
	return defaultValue
}

func getBool(source config.ConfigSource, envKey, fileKey string, defaultValue bool) bool {
	// Try env first
	if val, ok := source.Get(envKey); ok {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	// Try file
	if val, ok := source.Get(fileKey); ok {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getFloat(source config.ConfigSource, envKey, fileKey string, defaultValue float64) float64 {
	// Try env first
	if val, ok := source.Get(envKey); ok {
		if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			return floatVal
		}
	}
	// Try file
	if val, ok := source.Get(fileKey); ok {
		if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getStringSlice(source config.ConfigSource, envKey, fileKey string, defaultValue []string) []string {
	var val string
	var ok bool

	// Try env first
	if val, ok = source.Get(envKey); !ok {
		// Try file
		val, ok = source.Get(fileKey)
	}

	if ok {
		parts := strings.Split(val, ",")
		var result []string
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultValue
}

// GetReadTimeout returns the read timeout duration.
func (c *GatewayConfig) GetReadTimeout() time.Duration {
	return time.Duration(c.Gateway.TimeoutMs) * time.Millisecond
}

// GetWriteTimeout returns the write timeout duration.
func (c *GatewayConfig) GetWriteTimeout() time.Duration {
	return time.Duration(c.Gateway.TimeoutMs) * time.Millisecond
}

// GetIdleTimeout returns the idle timeout duration.
func (c *GatewayConfig) GetIdleTimeout() time.Duration {
	return 120 * time.Second
}

// compositeSource is a local implementation that checks env first, then file
type compositeSource struct {
	sources []config.ConfigSource
}

func (c *compositeSource) Get(key string) (string, bool) {
	for _, source := range c.sources {
		if val, ok := source.Get(key); ok {
			return val, true
		}
	}
	return "", false
}

func (c *compositeSource) GetWithDefault(key, defaultValue string) string {
	for _, source := range c.sources {
		if val, ok := source.Get(key); ok {
			return val
		}
	}
	return defaultValue
}
