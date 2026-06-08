package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rossigee/mock-libvirtd/internal/handler"
	"github.com/rossigee/mock-libvirtd/internal/middleware"
)

func main() {
	// Configure logging
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	// Set up slog with JSON format
	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	slog.SetDefault(slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}),
	))

	// Configure Gin
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "release"
	}
	gin.SetMode(ginMode)

	// Create router
	router := gin.New()

	// Apply middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.StructuredLogging())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	// Health check endpoints
	health := handler.NewHealthHandler()
	router.GET("/health", health.Health)
	router.GET("/ready", health.Ready)

	// API endpoints
	api := router.Group("/api")
	{
		domains := handler.NewDomainHandler()
		api.GET("/domains", domains.List)
		api.POST("/domains", domains.Create)
		api.GET("/domains/:id", domains.Get)
		api.PUT("/domains/:id", domains.Update)
		api.DELETE("/domains/:id", domains.Delete)

		networks := handler.NewNetworkHandler()
		api.GET("/networks", networks.List)
		api.POST("/networks", networks.Create)
		api.GET("/networks/:id", networks.Get)
		api.DELETE("/networks/:id", networks.Delete)

		storage := handler.NewStorageHandler()
		api.GET("/storage", storage.List)
		api.POST("/storage", storage.Create)
		api.GET("/storage/:id", storage.Get)
		api.DELETE("/storage/:id", storage.Delete)
	}

	// Start server
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("starting mock-libvirtd", slog.String("port", port))
	if err := router.Run(":" + port); err != nil {
		slog.Error("failed to start server", slog.Any("error", err))
		os.Exit(1)
	}
}
