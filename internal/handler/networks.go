//nolint:errcheck,dupl // Mock service ignores UUID and JSON unmarshal errors; similar CRUD handlers across resource types
package handler

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxNetworks = 100
)

type Network struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Bridge string `json:"bridge"`
	Active bool   `json:"active"`
}

type NetworkHandler struct {
	networks map[string]*Network
	mu       sync.RWMutex
}

func NewNetworkHandler() *NetworkHandler {
	return &NetworkHandler{
		networks: make(map[string]*Network),
	}
}

func (h *NetworkHandler) List(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	h.mu.RLock()
	networks := make([]*Network, 0, len(h.networks))
	for _, n := range h.networks {
		networks = append(networks, n)
	}
	h.mu.RUnlock()

	slog.InfoContext(c.Request.Context(), "listed networks",
		slog.String("request_id", requestID.(string)),
		slog.Int("count", len(networks)),
	)

	c.JSON(http.StatusOK, gin.H{
		"data": networks,
	})
}

func (h *NetworkHandler) Create(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	var req struct {
		Name   string `json:"name" binding:"required"`
		Bridge string `json:"bridge"`
	}

	if err := c.BindJSON(&req); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid network creation request",
			slog.String("request_id", requestID.(string)),
			slog.Any("error", err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "invalid request",
			"request_id": requestID,
		})
		return
	}

	if len(req.Name) == 0 || len(req.Name) > maxNameLength {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "name must be 1-255 characters",
			"request_id": requestID,
		})
		return
	}

	h.mu.Lock()
	if len(h.networks) >= maxNetworks {
		h.mu.Unlock()
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":      "network limit reached",
			"request_id": requestID,
		})
		return
	}
	h.mu.Unlock()

	if req.Bridge == "" {
		req.Bridge = "virbr0"
	}

	network := &Network{
		ID:     uuid.New().String(),
		Name:   req.Name,
		Bridge: req.Bridge,
		Active: true,
	}

	h.mu.Lock()
	h.networks[network.ID] = network
	h.mu.Unlock()

	slog.InfoContext(c.Request.Context(), "created network",
		slog.String("request_id", requestID.(string)),
		slog.String("network_id", network.ID),
		slog.String("name", network.Name),
	)

	c.JSON(http.StatusCreated, network)
}

func (h *NetworkHandler) Get(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	id := c.Param("id")

	h.mu.RLock()
	network, exists := h.networks[id]
	h.mu.RUnlock()

	if !exists {
		slog.WarnContext(c.Request.Context(), "network not found",
			slog.String("request_id", requestID.(string)),
			slog.String("network_id", id),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "network not found",
			"request_id": requestID,
		})
		return
	}

	c.JSON(http.StatusOK, network)
}

func (h *NetworkHandler) Delete(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	id := c.Param("id")

	h.mu.Lock()
	_, exists := h.networks[id]
	if !exists {
		h.mu.Unlock()
		slog.WarnContext(c.Request.Context(), "network not found",
			slog.String("request_id", requestID.(string)),
			slog.String("network_id", id),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "network not found",
			"request_id": requestID,
		})
		return
	}
	delete(h.networks, id)
	h.mu.Unlock()

	slog.InfoContext(c.Request.Context(), "deleted network",
		slog.String("request_id", requestID.(string)),
		slog.String("network_id", id),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "network deleted",
	})
}

func (h *NetworkHandler) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.networks)
}
