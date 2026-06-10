package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHealthHandler()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/health", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r

	h.Health(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestReady(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHealthHandler()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/ready", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r

	h.Ready(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d before MarkReady, got %d", http.StatusServiceUnavailable, w.Code)
	}

	h.MarkReady()
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/ready", nil)
	c, _ = gin.CreateTestContext(w)
	c.Request = r

	h.Ready(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d after MarkReady, got %d", http.StatusOK, w.Code)
	}
}

func TestHealthStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHealthHandler()

	SetStatsFunc(func() Stats {
		return Stats{
			Domains:    5,
			Networks:   2,
			Storage:    3,
			Volumes:    10,
			UptimeSec:  100,
			MaxDomains: 1000,
		}
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/stats", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r

	h.Stats(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var stats Stats
	if err := json.Unmarshal(w.Body.Bytes(), &stats); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if stats.Domains != 5 {
		t.Fatalf("expected domains 5, got %d", stats.Domains)
	}
}

func TestHealthHome(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHealthHandler()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r

	h.Home(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "text/html" {
		t.Fatalf("expected HTML content type")
	}
}

func TestHealthMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHealthHandler()

	SetStatsFunc(func() Stats {
		return Stats{
			Domains:    5,
			Networks:   2,
			MaxDomains: 1000,
		}
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/metrics", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r

	h.Metrics(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "text/plain; version=0.0.4" {
		t.Fatalf("expected Prometheus content type")
	}
}