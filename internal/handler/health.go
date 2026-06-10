package handler

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

type Stats struct {
	Domains    int `json:"domains"`
	Networks   int `json:"networks"`
	Storage    int `json:"storage_pools"`
	Volumes    int `json:"volumes"`
	UptimeSec  int `json:"uptime_sec"`
	MaxDomains int `json:"max_domains"`
}

type HealthHandler struct {
	ready     atomic.Bool
	startTime time.Time
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
	}
}

func (h *HealthHandler) MarkReady() {
	h.ready.Store(true)
}

func (h *HealthHandler) MarkNotReady() {
	h.ready.Store(false)
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().Unix(),
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	if !h.ready.Load() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "initializing",
			"timestamp": time.Now().Unix(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().Unix(),
	})
}

func getStats() Stats {
	s := Stats{MaxDomains: getMaxDomains()}
	if statsFunc != nil {
		stats := statsFunc()
		s.Domains = stats.Domains
		s.Networks = stats.Networks
		s.Storage = stats.Storage
		s.Volumes = stats.Volumes
	}
	return s
}

var statsFunc func() Stats

func SetStatsFunc(f func() Stats) {
	statsFunc = f
}

func (h *HealthHandler) Stats(c *gin.Context) {
	stats := getStats()
	stats.UptimeSec = int(time.Since(h.startTime).Seconds())
	c.JSON(http.StatusOK, stats)
}

func (h *HealthHandler) Home(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, `<!DOCTYPE html>
<html>
<head>
	<title>mock-libvirtd</title>
	<style>
		body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
		h1 { color: #333; }
		.status { display: flex; gap: 20px; margin: 20px 0; }
		.card { background: #f5f5f5; padding: 20px; border-radius: 8px; }
		.card h3 { margin: 0 0 10px 0; color: #666; }
		.card .value { font-size: 32px; font-weight: bold; color: #333; }
		pre { background: #f5f5f5; padding: 15px; border-radius: 8px; overflow-x: auto; }
		a { color: #0066cc; }
	</style>
</head>
<body>
	<h1>mock-libvirtd</h1>
	<p>A mock libvirtd service for E2E testing of libvirt-based applications.</p>
	<div class="status">
		<div class="card"><h3>Domains</h3><div class="value" id="domains">-</div></div>
		<div class="card"><h3>Networks</h3><div class="value" id="networks">-</div></div>
		<div class="card"><h3>Storage Pools</h3><div class="value" id="storage">-</div></div>
		<div class="card"><h3>Uptime</h3><div class="value" id="uptime">-</div></div>
	</div>
	<h2>API Endpoints</h2>
	<pre>
GET  /health          - Liveness probe
GET  /ready           - Readiness probe
GET  /stats           - JSON stats
GET  /metrics         - Prometheus metrics
GET  /api/domains     - List domains
POST /api/domains     - Create domain
GET  /api/networks    - List networks
POST /api/networks   - Create network
GET  /api/storage     - List storage pools
POST /api/storage    - Create storage pool
	</pre>
	<script>
		async function update() {
			const r = await fetch('/stats');
			const s = await r.json();
			document.getElementById('domains').textContent = s.domains;
			document.getElementById('networks').textContent = s.networks;
			document.getElementById('storage').textContent = s.storage_pools;
			document.getElementById('uptime').textContent = s.uptime_sec + 's';
		}
		update();
		setInterval(update, 5000);
	</script>
</body>
</html>`)
}

func (h *HealthHandler) Metrics(c *gin.Context) {
	uptime := int(time.Since(h.startTime).Seconds())

	c.Header("Content-Type", "text/plain; version=0.0.4")
	c.String(http.StatusOK, "# HELP mock_libvirtd_domains Current number of domains\n")
	c.String(http.StatusOK, "# TYPE mock_libvirtd_domains gauge\n")
	c.String(http.StatusOK, "mock_libvirtd_domains %d\n", getStats().Domains)

	c.String(http.StatusOK, "# HELP mock_libvirtd_networks Current number of networks\n")
	c.String(http.StatusOK, "# TYPE mock_libvirtd_networks gauge\n")
	c.String(http.StatusOK, "mock_libvirtd_networks %d\n", getStats().Networks)

	c.String(http.StatusOK, "# HELP mock_libvirtd_storage_pools Current number of storage pools\n")
	c.String(http.StatusOK, "# TYPE mock_libvirtd_storage_pools gauge\n")
	c.String(http.StatusOK, "mock_libvirtd_storage_pools %d\n", getStats().Storage)

	c.String(http.StatusOK, "# HELP mock_libvirtd_volumes Current number of volumes\n")
	c.String(http.StatusOK, "# TYPE mock_libvirtd_volumes gauge\n")
	c.String(http.StatusOK, "mock_libvirtd_volumes %d\n", getStats().Volumes)

	c.String(http.StatusOK, "# HELP mock_libvirtd_uptime_seconds Service uptime in seconds\n")
	c.String(http.StatusOK, "# TYPE mock_libvirtd_uptime_seconds gauge\n")
	c.String(http.StatusOK, "mock_libvirtd_uptime_seconds %d\n", uptime)
}
