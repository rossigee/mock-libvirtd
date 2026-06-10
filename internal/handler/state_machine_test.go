//nolint:errcheck // Test file ignores JSON unmarshal errors for brevity
package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestStateMachineTransitions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	// Create domain
	createReq := struct {
		Name   string `json:"name"`
		Memory int    `json:"memory"`
		CPUs   int    `json:"cpus"`
	}{Name: "test-vm", Memory: 512, CPUs: 2}

	createBody, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(createBody))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")
	h.Create(c)

	var domain Domain
	json.Unmarshal(w.Body.Bytes(), &domain)
	domainID := domain.ID

	if domain.State != "shutoff" {
		t.Fatalf("expected initial state shutoff, got %s", domain.State)
	}

	// Request start (transition through starting -> running)
	updateReq := struct {
		State string `json:"state"`
	}{State: "running"}
	updateBody, _ := json.Marshal(updateReq)
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("PUT", "/api/domains/"+domainID, bytes.NewBuffer(updateBody))
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: domainID}}
	c.Set("request_id", "test-123")
	h.Update(c)

	var updated Domain
	json.Unmarshal(w.Body.Bytes(), &updated)
	if updated.State != "running" {
		t.Fatalf("expected state running (desired), got %s", updated.State)
	}

	// Poll until running (state machine transitions starting -> running)
	time.Sleep(2 * time.Second)
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/api/domains/"+domainID, nil)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: domainID}}
	c.Set("request_id", "test-123")
	h.Get(c)

	var running Domain
	json.Unmarshal(w.Body.Bytes(), &running)
	if running.State != "running" {
		t.Fatalf("expected state running after boot, got %s", running.State)
	}
	if running.StartedAt == 0 {
		t.Fatal("expected StartedAt to be set")
	}
	if running.Uptime == 0 {
		t.Fatal("expected uptime > 0 when running")
	}

	// Test pause transition
	updateReq = struct {
		State string `json:"state"`
	}{State: "paused"}
	updateBody, _ = json.Marshal(updateReq)
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("PUT", "/api/domains/"+domainID, bytes.NewBuffer(updateBody))
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: domainID}}
	c.Set("request_id", "test-123")
	h.Update(c)

	var paused Domain
	json.Unmarshal(w.Body.Bytes(), &paused)
	if paused.State != "paused" {
		t.Fatalf("expected state paused, got %s", paused.State)
	}

	// Verify uptime preserved while paused
	time.Sleep(500 * time.Millisecond)
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/api/domains/"+domainID, nil)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: domainID}}
	c.Set("request_id", "test-123")
	h.Get(c)

	var stillPaused Domain
	json.Unmarshal(w.Body.Bytes(), &stillPaused)
	if stillPaused.State != "paused" {
		t.Fatalf("expected state still paused, got %s", stillPaused.State)
	}
	if stillPaused.Uptime == 0 {
		t.Fatal("expected uptime preserved when paused")
	}

	// Test stop transition (request shutoff)
	updateReq = struct {
		State string `json:"state"`
	}{State: "shutoff"}
	updateBody, _ = json.Marshal(updateReq)
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("PUT", "/api/domains/"+domainID, bytes.NewBuffer(updateBody))
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: domainID}}
	c.Set("request_id", "test-123")
	h.Update(c)

	// Poll until shutoff
	time.Sleep(500 * time.Millisecond)
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/api/domains/"+domainID, nil)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: domainID}}
	c.Set("request_id", "test-123")
	h.Get(c)

	var shutoff Domain
	json.Unmarshal(w.Body.Bytes(), &shutoff)
	if shutoff.State != "shutoff" {
		t.Fatalf("expected state shutoff, got %s", shutoff.State)
	}
	if shutoff.Uptime != 0 {
		t.Fatal("expected uptime reset to 0 when shutoff")
	}
}

func TestInvalidStateTransition(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	// Create domain
	createReq := struct {
		Name string `json:"name"`
	}{Name: "test-vm"}
	createBody, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(createBody))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")
	h.Create(c)

	var domain Domain
	json.Unmarshal(w.Body.Bytes(), &domain)
	domainID := domain.ID

	// Try invalid transition: shutoff -> paused (should fail)
	updateReq := struct {
		State string `json:"state"`
	}{State: "paused"}
	updateBody, _ := json.Marshal(updateReq)
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("PUT", "/api/domains/"+domainID, bytes.NewBuffer(updateBody))
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: domainID}}
	c.Set("request_id", "test-123")
	h.Update(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409 for invalid transition, got %d", w.Code)
	}

	var errResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errResp)
	if errResp["error"] != "invalid state transition" {
		t.Fatalf("expected error message about invalid transition, got: %v", errResp["error"])
	}
}

func TestConcurrentDomains(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	// Create 3 domains concurrently
	domainIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		createReq := struct {
			Name string `json:"name"`
		}{Name: "vm-" + string(rune('1'+i))}
		createBody, _ := json.Marshal(createReq)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(createBody))
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		c.Set("request_id", "test-123")
		h.Create(c)

		var domain Domain
		json.Unmarshal(w.Body.Bytes(), &domain)
		domainIDs[i] = domain.ID
	}

	// Start all 3 domains
	for i, id := range domainIDs {
		updateReq := struct {
			State string `json:"state"`
		}{State: "running"}
		updateBody, _ := json.Marshal(updateReq)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("PUT", "/api/domains/"+id, bytes.NewBuffer(updateBody))
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		c.Params = []gin.Param{{Key: "id", Value: id}}
		c.Set("request_id", "test-123")
		h.Update(c)

		if w.Code != http.StatusOK {
			t.Fatalf("domain %d: expected 200, got %d", i, w.Code)
		}
	}

	// Wait for all to boot
	time.Sleep(2 * time.Second)

	// Verify all are running with different metrics
	for i, id := range domainIDs {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/api/domains/"+id, nil)
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		c.Params = []gin.Param{{Key: "id", Value: id}}
		c.Set("request_id", "test-123")
		h.Get(c)

		var domain Domain
		json.Unmarshal(w.Body.Bytes(), &domain)
		if domain.State != "running" {
			t.Fatalf("domain %d: expected running, got %s", i, domain.State)
		}
		if domain.Uptime == 0 {
			t.Fatalf("domain %d: expected uptime > 0", i)
		}
		if domain.CPUUsage == 0 {
			t.Fatalf("domain %d: expected CPU usage > 0", i)
		}
		if domain.MemUsage == 0 {
			t.Fatalf("domain %d: expected memory usage > 0", i)
		}
	}

	// List all domains
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/api/domains", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")
	h.List(c)

	var listResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	domains := listResp["data"].([]interface{})
	if len(domains) != 3 {
		t.Fatalf("expected 3 domains, got %d", len(domains))
	}
}

func TestMetricsProgression(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	// Create and start domain
	createReq := struct {
		Name string `json:"name"`
	}{Name: "test-vm"}
	createBody, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(createBody))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")
	h.Create(c)

	var domain Domain
	json.Unmarshal(w.Body.Bytes(), &domain)
	domainID := domain.ID

	// Start it
	updateReq := struct {
		State string `json:"state"`
	}{State: "running"}
	updateBody, _ := json.Marshal(updateReq)
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("PUT", "/api/domains/"+domainID, bytes.NewBuffer(updateBody))
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: domainID}}
	c.Set("request_id", "test-123")
	h.Update(c)

	// Wait and check metrics progression
	time.Sleep(2 * time.Second)

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/api/domains/"+domainID, nil)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: domainID}}
	c.Set("request_id", "test-123")
	h.Get(c)

	var running1 Domain
	json.Unmarshal(w.Body.Bytes(), &running1)
	uptime1 := running1.Uptime
	cpu1 := running1.CPUUsage
	mem1 := running1.MemUsage

	// Wait more
	time.Sleep(1 * time.Second)

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/api/domains/"+domainID, nil)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: domainID}}
	c.Set("request_id", "test-123")
	h.Get(c)

	var running2 Domain
	json.Unmarshal(w.Body.Bytes(), &running2)
	uptime2 := running2.Uptime
	cpu2 := running2.CPUUsage
	mem2 := running2.MemUsage

	if uptime2 <= uptime1 {
		t.Fatalf("uptime should increase: %d -> %d", uptime1, uptime2)
	}
	if mem2 <= mem1 {
		t.Fatalf("memory usage should increase: %.1f -> %.1f", mem1, mem2)
	}
	if cpu2 <= cpu1 {
		t.Fatalf("CPU usage should increase: %.1f -> %.1f", cpu1, cpu2)
	}
}
