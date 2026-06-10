package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

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

func TestNetworkCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewNetworkHandler()

	if h.Count() != 0 {
		t.Fatalf("expected 0 networks, got %d", h.Count())
	}

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

	if h.Count() != 1 {
		t.Fatalf("expected 1 network, got %d", h.Count())
	}
}