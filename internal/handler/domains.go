//nolint:errcheck,dupl // Mock service ignores UUID and JSON unmarshal errors; similar CRUD handlers across resource types
package handler

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Domain struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	State  string `json:"state"`
	Memory int    `json:"memory"`
	CPUs   int    `json:"cpus"`
}

type DomainHandler struct {
	domains map[string]*Domain
	mu      sync.RWMutex
}

func NewDomainHandler() *DomainHandler {
	return &DomainHandler{
		domains: make(map[string]*Domain),
	}
}

func (h *DomainHandler) List(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	h.mu.RLock()
	domains := make([]*Domain, 0, len(h.domains))
	for _, d := range h.domains {
		domains = append(domains, d)
	}
	h.mu.RUnlock()

	slog.InfoContext(c.Request.Context(), "listed domains",
		slog.String("request_id", requestID.(string)),
		slog.Int("count", len(domains)),
	)

	c.JSON(http.StatusOK, gin.H{
		"data": domains,
	})
}

func (h *DomainHandler) Create(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	var req struct {
		Name   string `json:"name" binding:"required"`
		Memory int    `json:"memory"`
		CPUs   int    `json:"cpus"`
	}

	if err := c.BindJSON(&req); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid domain creation request",
			slog.String("request_id", requestID.(string)),
			slog.Any("error", err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "invalid request",
			"request_id": requestID,
		})
		return
	}

	if req.Memory == 0 {
		req.Memory = 512
	}
	if req.CPUs == 0 {
		req.CPUs = 1
	}

	domain := &Domain{
		ID:     uuid.New().String(),
		Name:   req.Name,
		State:  "shutoff",
		Memory: req.Memory,
		CPUs:   req.CPUs,
	}

	h.mu.Lock()
	h.domains[domain.ID] = domain
	h.mu.Unlock()

	slog.InfoContext(c.Request.Context(), "created domain",
		slog.String("request_id", requestID.(string)),
		slog.String("domain_id", domain.ID),
		slog.String("name", domain.Name),
	)

	c.JSON(http.StatusCreated, domain)
}

func (h *DomainHandler) Get(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	id := c.Param("id")

	h.mu.RLock()
	domain, exists := h.domains[id]
	h.mu.RUnlock()

	if !exists {
		slog.WarnContext(c.Request.Context(), "domain not found",
			slog.String("request_id", requestID.(string)),
			slog.String("domain_id", id),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "domain not found",
			"request_id": requestID,
		})
		return
	}

	c.JSON(http.StatusOK, domain)
}

func (h *DomainHandler) Update(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	id := c.Param("id")

	var req struct {
		State  string `json:"state"`
		Memory int    `json:"memory"`
		CPUs   int    `json:"cpus"`
	}

	if err := c.BindJSON(&req); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid domain update request",
			slog.String("request_id", requestID.(string)),
			slog.String("domain_id", id),
			slog.Any("error", err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "invalid request",
			"request_id": requestID,
		})
		return
	}

	h.mu.Lock()
	domain, exists := h.domains[id]
	if !exists {
		h.mu.Unlock()
		slog.WarnContext(c.Request.Context(), "domain not found",
			slog.String("request_id", requestID.(string)),
			slog.String("domain_id", id),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "domain not found",
			"request_id": requestID,
		})
		return
	}

	if req.State != "" {
		domain.State = req.State
	}
	if req.Memory > 0 {
		domain.Memory = req.Memory
	}
	if req.CPUs > 0 {
		domain.CPUs = req.CPUs
	}
	h.mu.Unlock()

	slog.InfoContext(c.Request.Context(), "updated domain",
		slog.String("request_id", requestID.(string)),
		slog.String("domain_id", id),
	)

	c.JSON(http.StatusOK, domain)
}

func (h *DomainHandler) Delete(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	id := c.Param("id")

	h.mu.Lock()
	_, exists := h.domains[id]
	if !exists {
		h.mu.Unlock()
		slog.WarnContext(c.Request.Context(), "domain not found",
			slog.String("request_id", requestID.(string)),
			slog.String("domain_id", id),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "domain not found",
			"request_id": requestID,
		})
		return
	}
	delete(h.domains, id)
	h.mu.Unlock()

	slog.InfoContext(c.Request.Context(), "deleted domain",
		slog.String("request_id", requestID.(string)),
		slog.String("domain_id", id),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "domain deleted",
	})
}
