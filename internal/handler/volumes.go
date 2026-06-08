package handler

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StorageVolume struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	PoolID       string `json:"pool_id"`
	Type         string `json:"type"`
	Size         int64  `json:"size"`
	Path         string `json:"path"`
	Format       string `json:"format"`
	BackingStore string `json:"backing_store,omitempty"`
}

type VolumeHandler struct {
	volumes map[string]*StorageVolume
	mu      sync.RWMutex
}

func NewVolumeHandler() *VolumeHandler {
	return &VolumeHandler{
		volumes: make(map[string]*StorageVolume),
	}
}

func (h *VolumeHandler) List(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	poolID := c.Param("pool_id")

	h.mu.RLock()
	volumes := make([]*StorageVolume, 0)
	for _, v := range h.volumes {
		if v.PoolID == poolID {
			volumes = append(volumes, v)
		}
	}
	h.mu.RUnlock()

	slog.InfoContext(c.Request.Context(), "listed storage volumes",
		slog.String("request_id", requestID.(string)),
		slog.String("pool_id", poolID),
		slog.Int("count", len(volumes)),
	)

	c.JSON(http.StatusOK, gin.H{
		"data": volumes,
	})
}

func (h *VolumeHandler) Create(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	poolID := c.Param("pool_id")

	var req struct {
		Name         string `json:"name" binding:"required"`
		Type         string `json:"type"`
		Size         int64  `json:"size" binding:"required"`
		Format       string `json:"format"`
		BackingStore string `json:"backing_store"`
	}

	if err := c.BindJSON(&req); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid volume creation request",
			slog.String("request_id", requestID.(string)),
			slog.String("pool_id", poolID),
			slog.Any("error", err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "invalid request",
			"request_id": requestID,
		})
		return
	}

	if req.Type == "" {
		req.Type = "file"
	}
	if req.Format == "" {
		req.Format = "qcow2"
	}

	volume := &StorageVolume{
		ID:           uuid.New().String(),
		Name:         req.Name,
		PoolID:       poolID,
		Type:         req.Type,
		Size:         req.Size,
		Path:         "/var/lib/libvirt/images/" + req.Name,
		Format:       req.Format,
		BackingStore: req.BackingStore,
	}

	h.mu.Lock()
	h.volumes[volume.ID] = volume
	h.mu.Unlock()

	slog.InfoContext(c.Request.Context(), "created storage volume",
		slog.String("request_id", requestID.(string)),
		slog.String("pool_id", poolID),
		slog.String("volume_id", volume.ID),
		slog.String("name", volume.Name),
		slog.Int64("size", volume.Size),
	)

	c.JSON(http.StatusCreated, volume)
}

func (h *VolumeHandler) Get(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	poolID := c.Param("pool_id")
	volumeID := c.Param("volume_id")

	h.mu.RLock()
	volume, exists := h.volumes[volumeID]
	h.mu.RUnlock()

	if !exists || volume.PoolID != poolID {
		slog.WarnContext(c.Request.Context(), "volume not found",
			slog.String("request_id", requestID.(string)),
			slog.String("pool_id", poolID),
			slog.String("volume_id", volumeID),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "volume not found",
			"request_id": requestID,
		})
		return
	}

	c.JSON(http.StatusOK, volume)
}

func (h *VolumeHandler) Delete(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	poolID := c.Param("pool_id")
	volumeID := c.Param("volume_id")

	h.mu.Lock()
	volume, exists := h.volumes[volumeID]
	if !exists || volume.PoolID != poolID {
		h.mu.Unlock()
		slog.WarnContext(c.Request.Context(), "volume not found",
			slog.String("request_id", requestID.(string)),
			slog.String("pool_id", poolID),
			slog.String("volume_id", volumeID),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "volume not found",
			"request_id": requestID,
		})
		return
	}
	delete(h.volumes, volumeID)
	h.mu.Unlock()

	slog.InfoContext(c.Request.Context(), "deleted storage volume",
		slog.String("request_id", requestID.(string)),
		slog.String("pool_id", poolID),
		slog.String("volume_id", volumeID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "volume deleted",
	})
}
