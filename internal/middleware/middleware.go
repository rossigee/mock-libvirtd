package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func getEnvInt(key string, def int) int {
	if s := os.Getenv(key); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			return v
		}
	}
	return def
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func StructuredLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID, _ := c.Get("request_id")

		c.Next()

		duration := time.Since(start)
		slog.InfoContext(c.Request.Context(), "request completed",
			slog.String("request_id", requestID.(string)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.String("client_ip", c.ClientIP()),
			slog.Int64("duration_ms", duration.Milliseconds()),
			slog.Int("response_bytes", c.Writer.Size()),
		)
	}
}

func Tracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		tracer := otel.Tracer("mock-libvirtd")
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		spanName := c.Request.Method + " " + c.FullPath()
		if spanName == " " {
			spanName = c.Request.Method + " " + c.Request.URL.Path
		}

		ctx, span := tracer.Start(ctx, spanName, trace.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.route", c.FullPath()),
			attribute.String("net.host.name", c.Request.Host),
			attribute.String("http.scheme", "http"),
		))
		defer span.End()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int64("http.response.size", int64(c.Writer.Size())),
		)

		if len(c.Errors) > 0 {
			span.RecordError(c.Errors[0])
		}
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func InputValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "application/json" && contentType != "" {
				requestID, _ := c.Get("request_id")
				c.JSON(http.StatusBadRequest, gin.H{
					"error":      "invalid content type",
					"request_id": requestID,
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func RateLimit() gin.HandlerFunc {
	requestsPerSecond := getEnvInt("RATE_LIMIT_RPS", 100)

	requests := make(chan struct{}, requestsPerSecond)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			select {
			case requests <- struct{}{}:
			default:
			}
		}
	}()

	return func(c *gin.Context) {
		select {
		case requests <- struct{}{}:
			c.Next()
		default:
			requestID, _ := c.Get("request_id")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":      "rate limit exceeded",
				"request_id": requestID,
			})
			c.Abort()
		}
	}
}
