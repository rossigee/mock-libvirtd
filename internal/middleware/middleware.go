package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
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
