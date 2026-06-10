package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

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

func TestStorageCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewStorageHandler()

	if h.Count() != 0 {
		t.Fatalf("expected 0 pools, got %d", h.Count())
	}

	req := struct {
		Name string `json:"name"`
	}{Name: "test-pool"}
	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/storage", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Set("request_id", "test-123")
	h.Create(c)

	if h.Count() != 1 {
		t.Fatalf("expected 1 pool, got %d", h.Count())
	}
}