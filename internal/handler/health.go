package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	ready bool
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{ready: true}
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	if !h.ready {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
