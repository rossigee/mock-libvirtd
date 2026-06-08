package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// ============================================================
// DOMAIN HANDLER TESTS
// ============================================================

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

// ============================================================
// HEALTH HANDLER TESTS
// ============================================================

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

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// ============================================================
// NETWORK HANDLER TESTS
// ============================================================

func TestNetworkList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/networks", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("request_id", "test-123")

	h := NewNetworkHandler()
	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp struct {
		Data []Network `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
}

func TestNetworkCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewNetworkHandler()

	req := struct {
		Name   string `json:"name"`
		Bridge string `json:"bridge"`
	}{
		Name:   "test-network",
		Bridge: "virbr1",
	}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/networks", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")

	h.Create(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var network Network
	if err := json.Unmarshal(w.Body.Bytes(), &network); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if network.Name != req.Name {
		t.Fatalf("expected name %s, got %s", req.Name, network.Name)
	}
}

func TestNetworkCreateDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewNetworkHandler()

	req := struct {
		Name string `json:"name"`
	}{Name: "test-network"}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/networks", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")

	h.Create(c)

	var network Network
	if err := json.Unmarshal(w.Body.Bytes(), &network); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if network.Bridge != "virbr0" {
		t.Fatalf("expected default bridge virbr0, got %s", network.Bridge)
	}
}

func TestNetworkGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewNetworkHandler()

	createReq := struct {
		Name string `json:"name"`
	}{Name: "test-network"}
	body, _ := json.Marshal(createReq)
	w1 := httptest.NewRecorder()
	r1, _ := http.NewRequest("POST", "/api/networks", bytes.NewBuffer(body))
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = r1
	c1.Set("request_id", "test-123")
	h.Create(c1)

	var network Network
	if err := json.Unmarshal(w1.Body.Bytes(), &network); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	w2 := httptest.NewRecorder()
	r2, _ := http.NewRequest("GET", "/api/networks/"+network.ID, nil)
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = r2
	c2.Params = []gin.Param{{Key: "id", Value: network.ID}}
	c2.Set("request_id", "test-123")

	h.Get(c2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w2.Code)
	}
}

func TestNetworkGetNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewNetworkHandler()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/api/networks/nonexistent", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: "nonexistent"}}
	c.Set("request_id", "test-123")

	h.Get(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestNetworkDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewNetworkHandler()

	createReq := struct {
		Name string `json:"name"`
	}{Name: "test-network"}
	body, _ := json.Marshal(createReq)
	w1 := httptest.NewRecorder()
	r1, _ := http.NewRequest("POST", "/api/networks", bytes.NewBuffer(body))
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = r1
	c1.Set("request_id", "test-123")
	h.Create(c1)

	var network Network
	if err := json.Unmarshal(w1.Body.Bytes(), &network); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	w2 := httptest.NewRecorder()
	r2, _ := http.NewRequest("DELETE", "/api/networks/"+network.ID, nil)
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = r2
	c2.Params = []gin.Param{{Key: "id", Value: network.ID}}
	c2.Set("request_id", "test-123")

	h.Delete(c2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w2.Code)
	}
}

func TestNetworkDeleteNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewNetworkHandler()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("DELETE", "/api/networks/nonexistent", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: "nonexistent"}}
	c.Set("request_id", "test-123")

	h.Delete(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// ============================================================
// STORAGE HANDLER TESTS
// ============================================================

func TestStorageList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/storage", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("request_id", "test-123")

	h := NewStorageHandler()
	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp struct {
		Data []StoragePool `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
}

func TestStorageCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStorageHandler()

	req := struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Path     string `json:"path"`
		Capacity int64  `json:"capacity"`
	}{
		Name:     "test-storage",
		Type:     "dir",
		Path:     "/var/lib/libvirt/images",
		Capacity: 107374182400,
	}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/storage", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")

	h.Create(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var pool StoragePool
	if err := json.Unmarshal(w.Body.Bytes(), &pool); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if pool.Name != req.Name {
		t.Fatalf("expected name %s, got %s", req.Name, pool.Name)
	}
}

func TestStorageCreateDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStorageHandler()

	req := struct {
		Name string `json:"name"`
	}{Name: "test-storage"}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/storage", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")

	h.Create(c)

	var pool StoragePool
	if err := json.Unmarshal(w.Body.Bytes(), &pool); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if pool.Type != "dir" {
		t.Fatalf("expected default type dir, got %s", pool.Type)
	}
	if pool.Capacity != 107374182400 {
		t.Fatalf("expected default capacity 107374182400, got %d", pool.Capacity)
	}
}

func TestStorageGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStorageHandler()

	createReq := struct {
		Name string `json:"name"`
	}{Name: "test-storage"}
	body, _ := json.Marshal(createReq)
	w1 := httptest.NewRecorder()
	r1, _ := http.NewRequest("POST", "/api/storage", bytes.NewBuffer(body))
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = r1
	c1.Set("request_id", "test-123")
	h.Create(c1)

	var pool StoragePool
	if err := json.Unmarshal(w1.Body.Bytes(), &pool); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	w2 := httptest.NewRecorder()
	r2, _ := http.NewRequest("GET", "/api/storage/"+pool.ID, nil)
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = r2
	c2.Params = []gin.Param{{Key: "id", Value: pool.ID}}
	c2.Set("request_id", "test-123")

	h.Get(c2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w2.Code)
	}
}

func TestStorageGetNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStorageHandler()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/api/storage/nonexistent", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: "nonexistent"}}
	c.Set("request_id", "test-123")

	h.Get(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestStorageDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStorageHandler()

	createReq := struct {
		Name string `json:"name"`
	}{Name: "test-storage"}
	body, _ := json.Marshal(createReq)
	w1 := httptest.NewRecorder()
	r1, _ := http.NewRequest("POST", "/api/storage", bytes.NewBuffer(body))
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = r1
	c1.Set("request_id", "test-123")
	h.Create(c1)

	var pool StoragePool
	if err := json.Unmarshal(w1.Body.Bytes(), &pool); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	w2 := httptest.NewRecorder()
	r2, _ := http.NewRequest("DELETE", "/api/storage/"+pool.ID, nil)
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = r2
	c2.Params = []gin.Param{{Key: "id", Value: pool.ID}}
	c2.Set("request_id", "test-123")

	h.Delete(c2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w2.Code)
	}
}

func TestStorageDeleteNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStorageHandler()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("DELETE", "/api/storage/nonexistent", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "id", Value: "nonexistent"}}
	c.Set("request_id", "test-123")

	h.Delete(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
