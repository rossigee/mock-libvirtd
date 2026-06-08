package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
