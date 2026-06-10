// Mock libvirtd service for E2E testing.
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
	"github.com/rossigee/mock-libvirtd/internal/logging"
	"github.com/rossigee/mock-libvirtd/internal/middleware"
	"github.com/rossigee/mock-libvirtd/internal/tracing"
)

func main() {
	// Configure logging
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logging.Init(logLevel)

	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "release"
	}
	gin.SetMode(ginMode)

	// Initialize tracing (only if OTLP_ENDPOINT is set)
	if os.Getenv("OTLP_ENDPOINT") != "" {
		tracing.Init("mock-libvirtd")
	}

	// Create router
	router := gin.New()

	var (
		domainsHandler   *handler.DomainHandler
		networksHandler  *handler.NetworkHandler
		storageHandler   *handler.StorageHandler
		volumesHandler   *handler.VolumeHandler
	)

	// Apply middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.RateLimit())
	router.Use(middleware.Tracing())
	router.Use(middleware.StructuredLogging())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.InputValidation())

	// Health check endpoints
	health := handler.NewHealthHandler()
	router.GET("/", health.Home)
	router.GET("/health", health.Health)
	router.GET("/ready", health.Ready)
	router.GET("/stats", health.Stats)
	router.GET("/metrics", health.Metrics)

	// Configuration endpoints
	router.GET("/loglevel", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"level": logging.GetLevel()})
	})
	router.POST("/loglevel", func(c *gin.Context) {
		var req struct {
			Level string `json:"level"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		logging.SetLevel(req.Level)
		c.JSON(http.StatusOK, gin.H{"level": logging.GetLevel()})
	})

	// API endpoints
	api := router.Group("/api")
	{
		domains := handler.NewDomainHandler()
		api.GET("/domains", domains.List)
		api.POST("/domains", domains.Create)
		api.GET("/domains/:id", domains.Get)
		api.PUT("/domains/:id", domains.Update)
		api.DELETE("/domains/:id", domains.Delete)
		domainsHandler = domains

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

		volumes := handler.NewVolumeHandler()
		api.GET("/storage/:pool_id/volumes", volumes.List)
		api.POST("/storage/:pool_id/volumes", volumes.Create)
		api.GET("/storage/:pool_id/volumes/:volume_id", volumes.Get)
		api.DELETE("/storage/:pool_id/volumes/:volume_id", volumes.Delete)

		networksHandler = networks
		storageHandler = storage
		volumesHandler = volumes
	}

	// Wire up stats
	handler.SetStatsFunc(func() handler.Stats {
		return handler.Stats{
			Domains:  domainsHandler.Count(),
			Networks: networksHandler.Count(),
			Storage:  storageHandler.Count(),
			Volumes:  volumesHandler.Count(),
		}
	})

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

	// Wait for server to be ready before marking service as ready
	time.Sleep(1000 * time.Millisecond)
	health.MarkReady()
	slog.Info("service ready")

	// Wait for signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	slog.Info("shutting down")
	domainsHandler.Shutdown()
	tracing.Shutdown()
	shutdownErr := srv.Shutdown(ctx)
	cancel()

	if shutdownErr != nil {
		slog.Error("shutdown error", slog.Any("error", shutdownErr))
		os.Exit(1)
	}

	slog.Info("shutdown complete")
}
