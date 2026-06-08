//nolint:errcheck,dupl // Mock service ignores UUID and JSON unmarshal errors; similar CRUD handlers across storage types
package handler

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StoragePool struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Path     string `json:"path"`
	Active   bool   `json:"active"`
	Capacity int64  `json:"capacity"`
}

type StorageHandler struct {
	pools map[string]*StoragePool
	mu    sync.RWMutex
}

func NewStorageHandler() *StorageHandler {
	return &StorageHandler{
		pools: make(map[string]*StoragePool),
	}
}

func (h *StorageHandler) List(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	h.mu.RLock()
	pools := make([]*StoragePool, 0, len(h.pools))
	for _, p := range h.pools {
		pools = append(pools, p)
	}
	h.mu.RUnlock()

	slog.InfoContext(c.Request.Context(), "listed storage pools",
		slog.String("request_id", requestID.(string)),
		slog.Int("count", len(pools)),
	)

	c.JSON(http.StatusOK, gin.H{
		"data": pools,
	})
}

func (h *StorageHandler) Create(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	var req struct {
		Name     string `json:"name" binding:"required"`
		Type     string `json:"type"`
		Path     string `json:"path"`
		Capacity int64  `json:"capacity"`
	}

	if err := c.BindJSON(&req); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid storage creation request",
			slog.String("request_id", requestID.(string)),
			slog.Any("error", err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "invalid request",
			"request_id": requestID,
		})
		return
	}

	if req.Type == "" {
		req.Type = "dir"
	}
	if req.Path == "" {
		req.Path = "/var/lib/libvirt/images"
	}
	if req.Capacity == 0 {
		req.Capacity = 107374182400
	}

	pool := &StoragePool{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Type:     req.Type,
		Path:     req.Path,
		Active:   true,
		Capacity: req.Capacity,
	}

	h.mu.Lock()
	h.pools[pool.ID] = pool
	h.mu.Unlock()

	slog.InfoContext(c.Request.Context(), "created storage pool",
		slog.String("request_id", requestID.(string)),
		slog.String("pool_id", pool.ID),
		slog.String("name", pool.Name),
	)

	c.JSON(http.StatusCreated, pool)
}

func (h *StorageHandler) Get(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	id := c.Param("id")

	h.mu.RLock()
	pool, exists := h.pools[id]
	h.mu.RUnlock()

	if !exists {
		slog.WarnContext(c.Request.Context(), "storage pool not found",
			slog.String("request_id", requestID.(string)),
			slog.String("pool_id", id),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "storage pool not found",
			"request_id": requestID,
		})
		return
	}

	c.JSON(http.StatusOK, pool)
}

func (h *StorageHandler) Delete(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	id := c.Param("id")

	h.mu.Lock()
	_, exists := h.pools[id]
	if !exists {
		h.mu.Unlock()
		slog.WarnContext(c.Request.Context(), "storage pool not found",
			slog.String("request_id", requestID.(string)),
			slog.String("pool_id", id),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "storage pool not found",
			"request_id": requestID,
		})
		return
	}
	delete(h.pools, id)
	h.mu.Unlock()

	slog.InfoContext(c.Request.Context(), "deleted storage pool",
		slog.String("request_id", requestID.(string)),
		slog.String("pool_id", id),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "storage pool deleted",
	})
}
