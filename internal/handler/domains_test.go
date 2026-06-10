package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestDomainList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/domains", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("request_id", "test-123")

	h := NewDomainHandler()
	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp struct {
		Data []Domain `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
}

func TestDomainCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	req := struct {
		Name   string `json:"name"`
		Memory int    `json:"memory"`
		CPUs   int    `json:"cpus"`
	}{
		Name:   "test-vm",
		Memory: 1024,
		CPUs:   2,
	}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")

	h.Create(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var domain Domain
	if err := json.Unmarshal(w.Body.Bytes(), &domain); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if domain.Name != req.Name {
		t.Fatalf("expected name %s, got %s", req.Name, domain.Name)
	}
}

func TestDomainCreateInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	body := []byte(`{"invalid": "json"}`)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")

	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDomainCreateDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	req := struct {
		Name string `json:"name"`
	}{
		Name: "test-vm",
	}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")

	h.Create(c)

	var domain Domain
	if err := json.Unmarshal(w.Body.Bytes(), &domain); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if domain.Memory != 512 {
		t.Fatalf("expected default memory 512, got %d", domain.Memory)
	}
	if domain.CPUs != 1 {
		t.Fatalf("expected default CPUs 1, got %d", domain.CPUs)
	}
}

func TestDomainGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	createReq := struct {
		Name string `json:"name"`
	}{Name: "test-vm"}
	body, _ := json.Marshal(createReq)
	w1 := httptest.NewRecorder()
	r1, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(body))
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = r1
	c1.Set("request_id", "test-123")
	h.Create(c1)

	var domain Domain
	if err := json.Unmarshal(w1.Body.Bytes(), &domain); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	w2 := httptest.NewRecorder()
	r2, _ := http.NewRequest("GET", "/api/domains/"+domain.ID, nil)
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = r2
	c2.Params = []gin.Param{{Key: "id", Value: domain.ID}}
	c2.Set("request_id", "test-123")

	h.Get(c2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w2.Code)
	}

	var retrieved Domain
	if err := json.Unmarshal(w2.Body.Bytes(), &retrieved); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if retrieved.ID != domain.ID {
		t.Fatalf("expected ID %s, got %s", domain.ID, retrieved.ID)
	}
}

func TestDomainGetNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/api/domains/nonexistent", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: "nonexistent"}}
	c.Set("request_id", "test-123")

	h.Get(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDomainUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	createReq := struct {
		Name string `json:"name"`
	}{Name: "test-vm"}
	body, _ := json.Marshal(createReq)
	w1 := httptest.NewRecorder()
	r1, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(body))
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = r1
	c1.Set("request_id", "test-123")
	h.Create(c1)

	var domain Domain
	if err := json.Unmarshal(w1.Body.Bytes(), &domain); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	updateReq := struct {
		State string `json:"state"`
	}{State: "running"}
	updateBody, _ := json.Marshal(updateReq)
	w2 := httptest.NewRecorder()
	r2, _ := http.NewRequest("PUT", "/api/domains/"+domain.ID, bytes.NewBuffer(updateBody))
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = r2
	c2.Params = []gin.Param{{Key: "id", Value: domain.ID}}
	c2.Set("request_id", "test-123")

	h.Update(c2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w2.Code)
	}

	var updated Domain
	if err := json.Unmarshal(w2.Body.Bytes(), &updated); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if updated.State != "running" {
		t.Fatalf("expected state running, got %s", updated.State)
	}
}

func TestDomainUpdateNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	updateReq := struct {
		State string `json:"state"`
	}{State: "running"}
	updateBody, _ := json.Marshal(updateReq)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PUT", "/api/domains/nonexistent", bytes.NewBuffer(updateBody))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: "nonexistent"}}
	c.Set("request_id", "test-123")

	h.Update(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDomainDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	createReq := struct {
		Name string `json:"name"`
	}{Name: "test-vm"}
	body, _ := json.Marshal(createReq)
	w1 := httptest.NewRecorder()
	r1, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(body))
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = r1
	c1.Set("request_id", "test-123")
	h.Create(c1)

	var domain Domain
	if err := json.Unmarshal(w1.Body.Bytes(), &domain); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	w2 := httptest.NewRecorder()
	r2, _ := http.NewRequest("DELETE", "/api/domains/"+domain.ID, nil)
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = r2
	c2.Params = []gin.Param{{Key: "id", Value: domain.ID}}
	c2.Set("request_id", "test-123")

	h.Delete(c2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w2.Code)
	}
}

func TestDomainDeleteNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("DELETE", "/api/domains/nonexistent", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: "nonexistent"}}
	c.Set("request_id", "test-123")

	h.Delete(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDomainCreateTooMany(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_ = os.Setenv("MAX_DOMAINS", "2")
	defer func() { _ = os.Unsetenv("MAX_DOMAINS") }()

	h := NewDomainHandler()

	for i := 0; i < 2; i++ {
		req := struct {
			Name string `json:"name"`
		}{Name: fmt.Sprintf("test-vm-%d", i)}
		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(body))
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		c.Set("request_id", "test-123")
		h.Create(c)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
		}
	}

	req := struct {
		Name string `json:"name"`
	}{Name: "test-vm-3"}
	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")
	h.Create(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestDomainCreateInvalidName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	req := struct {
		Name string `json:"name"`
	}{Name: ""}
	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")
	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDomainCreateInvalidMemory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	req := struct {
		Name   string `json:"name"`
		Memory int    `json:"memory"`
	}{Name: "test-vm", Memory: 64}
	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")
	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDomainCreateInvalidCPUs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	req := struct {
		Name string `json:"name"`
		CPUs int    `json:"cpus"`
	}{Name: "test-vm", CPUs: 1000}
	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")
	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDomainUpdateStateConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	req := struct {
		Name string `json:"name"`
	}{Name: "test-vm"}
	body, _ := json.Marshal(req)
	w1 := httptest.NewRecorder()
	r1, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(body))
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = r1
	c1.Set("request_id", "test-123")
	h.Create(c1)

	var domain Domain
	if err := json.Unmarshal(w1.Body.Bytes(), &domain); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	updateReq := struct {
		State string `json:"state"`
	}{State: "paused"}
	updateBody, _ := json.Marshal(updateReq)
	w2 := httptest.NewRecorder()
	r2, _ := http.NewRequest("PUT", "/api/domains/"+domain.ID, bytes.NewBuffer(updateBody))
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = r2
	c2.Params = []gin.Param{{Key: "id", Value: domain.ID}}
	c2.Set("request_id", "test-123")

	h.Update(c2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, w2.Code)
	}
}

func TestDomainCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	if h.Count() != 0 {
		t.Fatalf("expected 0 domains, got %d", h.Count())
	}

	req := struct {
		Name string `json:"name"`
	}{Name: "test-vm"}
	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")
	h.Create(c)

	if h.Count() != 1 {
		t.Fatalf("expected 1 domain, got %d", h.Count())
	}
}

func TestStateMachineTransitions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

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
	_ = json.Unmarshal(w.Body.Bytes(), &domain)
	domainID := domain.ID

	if domain.State != "shutoff" {
		t.Fatalf("expected initial state shutoff, got %s", domain.State)
	}

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
	_ = json.Unmarshal(w.Body.Bytes(), &updated)
	if updated.State != "running" {
		t.Fatalf("expected state running (desired), got %s", updated.State)
	}

	time.Sleep(2 * time.Second)
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/api/domains/"+domainID, nil)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: domainID}}
	c.Set("request_id", "test-123")
	h.Get(c)

	var running Domain
	_ = json.Unmarshal(w.Body.Bytes(), &running)
	if running.State != "running" {
		t.Fatalf("expected state running after boot, got %s", running.State)
	}
	if running.StartedAt == 0 {
		t.Fatal("expected StartedAt to be set")
	}
	if running.Uptime == 0 {
		t.Fatal("expected uptime > 0 when running")
	}

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
	_ = json.Unmarshal(w.Body.Bytes(), &paused)
	if paused.State != "paused" {
		t.Fatalf("expected state paused, got %s", paused.State)
	}

	time.Sleep(500 * time.Millisecond)
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/api/domains/"+domainID, nil)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: domainID}}
	c.Set("request_id", "test-123")
	h.Get(c)

	var stillPaused Domain
	_ = json.Unmarshal(w.Body.Bytes(), &stillPaused)
	if stillPaused.State != "paused" {
		t.Fatalf("expected state still paused, got %s", stillPaused.State)
	}
	if stillPaused.Uptime == 0 {
		t.Fatal("expected uptime preserved when paused")
	}

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

	time.Sleep(500 * time.Millisecond)
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/api/domains/"+domainID, nil)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: domainID}}
	c.Set("request_id", "test-123")
	h.Get(c)

	var shutoff Domain
	_ = json.Unmarshal(w.Body.Bytes(), &shutoff)
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
	_ = json.Unmarshal(w.Body.Bytes(), &domain)
	domainID := domain.ID

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
	_ = json.Unmarshal(w.Body.Bytes(), &errResp)
	if errResp["error"] != "invalid state transition" {
		t.Fatalf("expected error message about invalid transition, got: %v", errResp["error"])
	}
}

func TestConcurrentDomains(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

	domainIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		createReq := struct {
			Name string `json:"name"`
		}{Name: fmt.Sprintf("vm-%d", i+1)}
		createBody, _ := json.Marshal(createReq)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/api/domains", bytes.NewBuffer(createBody))
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		c.Set("request_id", "test-123")
		h.Create(c)

		var domain Domain
		_ = json.Unmarshal(w.Body.Bytes(), &domain)
		domainIDs[i] = domain.ID
	}

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

	time.Sleep(2 * time.Second)

	for i, id := range domainIDs {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/api/domains/"+id, nil)
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		c.Params = []gin.Param{{Key: "id", Value: id}}
		c.Set("request_id", "test-123")
		h.Get(c)

		var domain Domain
		_ = json.Unmarshal(w.Body.Bytes(), &domain)
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

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/api/domains", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")
	h.List(c)

	var listResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &listResp)
	domains := listResp["data"].([]interface{})
	if len(domains) != 3 {
		t.Fatalf("expected 3 domains, got %d", len(domains))
	}
}

func TestMetricsProgression(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDomainHandler()

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
	_ = json.Unmarshal(w.Body.Bytes(), &domain)
	domainID := domain.ID

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

	time.Sleep(2 * time.Second)

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/api/domains/"+domainID, nil)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: domainID}}
	c.Set("request_id", "test-123")
	h.Get(c)

	var running1 Domain
	_ = json.Unmarshal(w.Body.Bytes(), &running1)
	uptime1 := running1.Uptime
	cpu1 := running1.CPUUsage
	mem1 := running1.MemUsage

	time.Sleep(1 * time.Second)

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/api/domains/"+domainID, nil)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: domainID}}
	c.Set("request_id", "test-123")
	h.Get(c)

	var running2 Domain
	_ = json.Unmarshal(w.Body.Bytes(), &running2)
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