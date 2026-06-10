package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func getEnvDuration(key string, def time.Duration) time.Duration {
	if s := os.Getenv(key); s != "" {
		if ms, err := strconv.Atoi(s); err == nil {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return def
}

func getEnvInt(key string, def int) int {
	if s := os.Getenv(key); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			return v
		}
	}
	return def
}

type Domain struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	State     string  `json:"state"`
	Memory    int     `json:"memory"`
	CPUs      int     `json:"cpus"`
	CreatedAt int64   `json:"created_at"`
	StartedAt int64   `json:"started_at,omitempty"`
	Uptime    int     `json:"uptime,omitempty"`
	CPUUsage  float64 `json:"cpu_usage,omitempty"`
	MemUsage  float64 `json:"mem_usage,omitempty"`

	mu            sync.RWMutex
	desiredState  string
	cancelFunc    context.CancelFunc
	stateUpdateCh chan struct{}
}

type DomainHandler struct {
	domains map[string]*Domain
	mu      sync.RWMutex
}

func (h *DomainHandler) Shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, d := range h.domains {
		if d.cancelFunc != nil {
			d.cancelFunc()
		}
	}
}

func (h *DomainHandler) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.domains)
}

var (
	stateTickRate = getEnvDuration("STATE_TICK_RATE_MS", 100*time.Millisecond)
	maxMemoryMB   = getEnvInt("MAX_MEMORY_MB", 1048576)
	maxCPUs       = getEnvInt("MAX_CPUS", 256)
	minMemoryMB   = getEnvInt("MIN_MEMORY_MB", 128)
	minCPUs       = getEnvInt("MIN_CPUS", 1)
	maxNameLength = getEnvInt("MAX_NAME_LENGTH", 255)
)

func getMaxDomains() int {
	return getEnvInt("MAX_DOMAINS", 1000)
}

var validTransitions = map[string]map[string]bool{
	"shutoff":  {"running": true},
	"starting": {"running": true, "shutoff": true},
	"running":  {"shutoff": true, "paused": true},
	"stopping": {"shutoff": true},
	"paused":   {"running": true, "shutoff": true},
}

func NewDomainHandler() *DomainHandler {
	return &DomainHandler{
		domains: make(map[string]*Domain),
	}
}

func (d *Domain) trySetDesiredState(nextState string) (string, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	current := d.State
	allowed, exists := validTransitions[current]
	if !exists || !allowed[nextState] {
		return current, false
	}
	d.desiredState = nextState
	return current, true
}

func (d *Domain) updateMetrics() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.State == "shutoff" || d.StartedAt == 0 {
		d.Uptime = 0
		d.CPUUsage = 0
		d.MemUsage = 0
		return
	}

	if d.State != "running" && d.State != "paused" {
		d.CPUUsage = 0
		d.MemUsage = 0
		return
	}

	uptime := time.Since(time.UnixMilli(d.StartedAt))
	d.Uptime = int(uptime.Seconds())

	cpuRamp := float64(d.Uptime) / 5.0
	if cpuRamp > 1 {
		cpuRamp = 1
	}
	d.CPUUsage = 10.0 + (cpuRamp * 30.0)

	memRamp := float64(d.Uptime) / 3.0
	if memRamp > 1 {
		memRamp = 1
	}
	d.MemUsage = 20.0 + (memRamp * 40.0)
}

func (d *Domain) runStateMachine(ctx context.Context) {
	ticker := time.NewTicker(stateTickRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.mu.Lock()
			d.State = "shutoff"
			d.StartedAt = 0
			d.mu.Unlock()
			return
		case <-ticker.C:
			d.mu.Lock()
			current := d.State
			desired := d.desiredState
			d.mu.Unlock()

			if current == desired {
				d.updateMetrics()
				continue
			}

			switch current {
			case "starting":
				d.mu.Lock()
				d.StartedAt = time.Now().UnixMilli()
				d.State = "running"
				d.mu.Unlock()
				slog.Info("domain transitioned", slog.String("domain", d.ID), slog.String("state", "running"))

			case "stopping":
				d.mu.Lock()
				d.State = "shutoff"
				d.StartedAt = 0
				d.mu.Unlock()
				slog.Info("domain transitioned", slog.String("domain", d.ID), slog.String("state", "shutoff"))

			case "shutoff":
				if desired == "running" {
					d.mu.Lock()
					d.State = "starting"
					d.mu.Unlock()
					slog.Info("domain transitioned", slog.String("domain", d.ID), slog.String("state", "starting"))
				}

			case "running":
				switch desired {
				case "shutoff":
					d.mu.Lock()
					d.State = "stopping"
					d.mu.Unlock()
					slog.Info("domain transitioned", slog.String("domain", d.ID), slog.String("state", "stopping"))
				case "paused":
					d.mu.Lock()
					d.State = "paused"
					d.mu.Unlock()
					slog.Info("domain transitioned", slog.String("domain", d.ID), slog.String("state", "paused"))
				}

			case "paused":
				switch desired {
				case "running":
					d.mu.Lock()
					d.State = "running"
					d.mu.Unlock()
					slog.Info("domain transitioned", slog.String("domain", d.ID), slog.String("state", "running"))
				case "shutoff":
					d.mu.Lock()
					d.State = "stopping"
					d.mu.Unlock()
					slog.Info("domain transitioned", slog.String("domain", d.ID), slog.String("state", "stopping"))
				}
			}

			d.updateMetrics()
		}
	}
}

func (d *Domain) snapshot() *Domain {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return &Domain{
		ID:        d.ID,
		Name:      d.Name,
		State:     d.State,
		Memory:    d.Memory,
		CPUs:      d.CPUs,
		CreatedAt: d.CreatedAt,
		StartedAt: d.StartedAt,
		Uptime:    d.Uptime,
		CPUUsage:  d.CPUUsage,
		MemUsage:  d.MemUsage,
	}
}

func (h *DomainHandler) List(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	h.mu.RLock()
	domains := make([]*Domain, 0, len(h.domains))
	for _, d := range h.domains {
		domains = append(domains, d.snapshot())
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

	if len(req.Name) == 0 || len(req.Name) > maxNameLength {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "name must be 1-255 characters",
			"request_id": requestID,
		})
		return
	}

	h.mu.Lock()
	if len(h.domains) >= getMaxDomains() {
		h.mu.Unlock()
		slog.WarnContext(c.Request.Context(), "domain limit reached",
			slog.String("request_id", requestID.(string)),
		)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":      "domain limit reached",
			"request_id": requestID,
		})
		return
	}
	h.mu.Unlock()

	if req.Memory == 0 {
		req.Memory = 512
	}
	if req.CPUs == 0 {
		req.CPUs = 1
	}

	if req.Memory < minMemoryMB || req.Memory > maxMemoryMB {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      fmt.Sprintf("memory must be between %d and %d MB", minMemoryMB, maxMemoryMB),
			"request_id": requestID,
		})
		return
	}
	if req.CPUs < minCPUs || req.CPUs > maxCPUs {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      fmt.Sprintf("CPUs must be between %d and %d", minCPUs, maxCPUs),
			"request_id": requestID,
		})
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	domain := &Domain{
		ID:            uuid.New().String(),
		Name:          req.Name,
		State:         "shutoff",
		Memory:        req.Memory,
		CPUs:          req.CPUs,
		CreatedAt:     time.Now().UnixMilli(),
		desiredState:  "running",
		cancelFunc:    cancel,
		stateUpdateCh: make(chan struct{}, 1),
	}

	h.mu.Lock()
	h.domains[domain.ID] = domain
	h.mu.Unlock()

	go domain.runStateMachine(ctx)

	slog.InfoContext(c.Request.Context(), "created domain",
		slog.String("request_id", requestID.(string)),
		slog.String("domain_id", domain.ID),
		slog.String("name", domain.Name),
	)

	c.JSON(http.StatusCreated, domain.snapshot())
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

	domain.updateMetrics()
	c.JSON(http.StatusOK, domain.snapshot())
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

	var responseState string
	if req.State != "" {
		currentState, ok := domain.trySetDesiredState(req.State)
		if !ok {
			slog.WarnContext(c.Request.Context(), "invalid state transition",
				slog.String("request_id", requestID.(string)),
				slog.String("domain_id", id),
				slog.String("current_state", currentState),
				slog.String("requested_state", req.State),
			)
			c.JSON(http.StatusConflict, gin.H{
				"error":            "invalid state transition",
				"current_state":    currentState,
				"requested_state":  req.State,
				"request_id":       requestID,
			})
			return
		}
		responseState = req.State
	}

	if req.Memory > 0 {
		domain.mu.Lock()
		domain.Memory = req.Memory
		domain.mu.Unlock()
	}

	if req.CPUs > 0 {
		domain.mu.Lock()
		domain.CPUs = req.CPUs
		domain.mu.Unlock()
	}

	domain.updateMetrics()
	slog.InfoContext(c.Request.Context(), "updated domain",
		slog.String("request_id", requestID.(string)),
		slog.String("domain_id", id),
		slog.String("new_state", req.State),
	)

	snapshot := domain.snapshot()
	if responseState != "" {
		snapshot.State = responseState
	}
	c.JSON(http.StatusOK, snapshot)
}

func (h *DomainHandler) Delete(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	id := c.Param("id")

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
	delete(h.domains, id)
	h.mu.Unlock()

	if domain.cancelFunc != nil {
		domain.cancelFunc()
	}

	slog.InfoContext(c.Request.Context(), "deleted domain",
		slog.String("request_id", requestID.(string)),
		slog.String("domain_id", id),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "domain deleted",
	})
}
