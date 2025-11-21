package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/api-gateway/cmd/api-gateway/routes"
	"github.com/yourorg/api-gateway/pkg/config"
	apigatewaytelemetry "github.com/yourorg/api-gateway/pkg/telemetry"
	"github.com/yourorg/go-service-kit/pkg/httpservice"
	"github.com/yourorg/go-service-kit/pkg/logging"
	"github.com/yourorg/go-service-kit/pkg/middleware"
	"github.com/yourorg/go-service-kit/pkg/telemetry"
)

const ServiceName = "api_gateway"

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create logger
	logger, err := logging.NewLogger(cfg.Gateway.LogLevel, cfg.Gateway.LogFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logging.Sync(logger)

	logger.Info("Starting API Gateway", logging.NewField("service", ServiceName))

	// Initialize telemetry clients
	newRelicClient, err := apigatewaytelemetry.NewNewRelicClient(apigatewaytelemetry.NewRelicConfig{
		LicenseKey: cfg.Telemetry.NewRelic.LicenseKey,
		AppName:    cfg.Telemetry.NewRelic.AppName,
		Enabled:    cfg.Telemetry.NewRelic.Enabled,
	}, logger)
	if err != nil {
		logger.Error("Failed to initialize New Relic client", logging.NewField("error", err))
		os.Exit(1)
	}

	// Initialize Slack client (from common-service)
	slackClient := telemetry.NewSlackClient(telemetry.SlackConfig{
		WebhookURL:  cfg.Telemetry.Slack.WebhookURL,
		ServiceName: "api_gateway", // Service name for Slack alerts
		Channel:     cfg.Telemetry.Slack.Channel,
		Enabled:     cfg.Telemetry.Slack.Enabled,
	}, logger)

	// Create HTTP server with custom middleware
	serverConfig := httpservice.ServerConfig{
		Port:           cfg.Gateway.Port,
		ReadTimeout:    cfg.GetReadTimeout(),
		WriteTimeout:   cfg.GetWriteTimeout(),
		IdleTimeout:    cfg.GetIdleTimeout(),
		Logger:         logger,
		RateLimitRPS:   cfg.Gateway.RateLimitRPS,
		RateLimitBurst: cfg.Gateway.RateLimitBurst,
		AllowedOrigins: cfg.Gateway.AllowedOrigins,
		AllowedMethods: cfg.Gateway.AllowedMethods,
		AllowedHeaders: cfg.Gateway.AllowedHeaders,
		MaxBodySize:    cfg.Gateway.MaxBodySize,
	}

	// Create a custom handler that sets up routes
	handler := &GatewayHandler{
		Logger:      logger,
		Config:      cfg,
		NrClient:    newRelicClient,
		SlackClient: slackClient,
	}

	server, err := httpservice.NewServer(serverConfig, handler)
	if err != nil {
		logger.Error("Failed to create HTTP server", logging.NewField("error", err))
		os.Exit(1)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Error("Server error", logging.NewField("error", err))
			cancel()
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		logger.Info("Shutdown signal received")
	case <-ctx.Done():
		logger.Info("Context cancelled")
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", logging.NewField("error", err))
	}

	// Shutdown telemetry clients
	if newRelicClient != nil {
		newRelicClient.Shutdown(5000)
	}

	logger.Info("API Gateway stopped")
}

// GatewayHandler implements the httpservice.Handler interface.
type GatewayHandler struct {
	Logger      logging.Logger
	Config      *config.GatewayConfig
	NrClient    middleware.TelemetryClient
	SlackClient middleware.SlackClient
}

// Register registers routes with the router.
func (h *GatewayHandler) Register(router *gin.Engine) {
	routes.Init(router, h.Logger, h.Config, h.NrClient, h.SlackClient)
}
