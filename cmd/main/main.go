package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	slog.Info("starting mock-libvirtd", slog.String("port", port))
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Mark service as ready (after server is running)
	health.MarkReady()
	slog.Info("service ready")

	// Wait for signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	slog.Info("shutting down")
	shutdownErr := srv.Shutdown(ctx)
	cancel()

	if shutdownErr != nil {
		slog.Error("shutdown error", slog.Any("error", shutdownErr))
		os.Exit(1)
	}

	slog.Info("shutdown complete")
}
