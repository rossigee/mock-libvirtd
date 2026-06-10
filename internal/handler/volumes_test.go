package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestVolumeList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVolumeHandler()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/api/storage/pool1/volumes", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "pool_id", Value: "pool1"}}
	c.Set("request_id", "test-123")

	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestVolumeCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVolumeHandler()

	req := struct {
		Name   string `json:"name"`
		Size   int64  `json:"size"`
		Type   string `json:"type"`
		Format string `json:"format"`
	}{
		Name:   "disk.qcow2",
		Size:   10737418240,
		Type:   "file",
		Format: "qcow2",
	}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/storage/pool1/volumes", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "pool_id", Value: "pool1"}}
	c.Set("request_id", "test-123")

	h.Create(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var vol StorageVolume
	if err := json.Unmarshal(w.Body.Bytes(), &vol); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if vol.Name != req.Name {
		t.Fatalf("expected name %s, got %s", req.Name, vol.Name)
	}
}

func TestVolumeCreateInvalidName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVolumeHandler()

	req := struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}{
		Name: "      ",
		Size: 10737418240,
	}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/storage/pool1/volumes", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "pool_id", Value: "pool1"}}
	c.Set("request_id", "test-123")

	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestVolumeCreateInvalidSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVolumeHandler()

	req := struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}{
		Name: "disk.qcow2",
		Size: -1,
	}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/storage/pool1/volumes", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "pool_id", Value: "pool1"}}
	c.Set("request_id", "test-123")

	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestVolumeGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVolumeHandler()

	createReq := struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}{
		Name: "disk.qcow2",
		Size: 10737418240,
	}
	body, _ := json.Marshal(createReq)
	w1 := httptest.NewRecorder()
	r1, _ := http.NewRequest("POST", "/api/storage/pool1/volumes", bytes.NewBuffer(body))
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = r1
	c1.Params = []gin.Param{{Key: "pool_id", Value: "pool1"}}
	c1.Set("request_id", "test-123")
	h.Create(c1)

	var vol StorageVolume
	if err := json.Unmarshal(w1.Body.Bytes(), &vol); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	w2 := httptest.NewRecorder()
	r2, _ := http.NewRequest("GET", "/api/storage/pool1/volumes/"+vol.ID, nil)
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = r2
	c2.Params = []gin.Param{
		{Key: "pool_id", Value: "pool1"},
		{Key: "volume_id", Value: vol.ID},
	}
	c2.Set("request_id", "test-123")

	h.Get(c2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w2.Code)
	}
}

func TestVolumeGetNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVolumeHandler()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/api/storage/pool1/volumes/nonexistent", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{
		{Key: "pool_id", Value: "pool1"},
		{Key: "volume_id", Value: "nonexistent"},
	}
	c.Set("request_id", "test-123")

	h.Get(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestVolumeDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVolumeHandler()

	createReq := struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}{
		Name: "disk.qcow2",
		Size: 10737418240,
	}
	body, _ := json.Marshal(createReq)
	w1 := httptest.NewRecorder()
	r1, _ := http.NewRequest("POST", "/api/storage/pool1/volumes", bytes.NewBuffer(body))
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = r1
	c1.Params = []gin.Param{{Key: "pool_id", Value: "pool1"}}
	c1.Set("request_id", "test-123")
	h.Create(c1)

	var vol StorageVolume
	if err := json.Unmarshal(w1.Body.Bytes(), &vol); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	w2 := httptest.NewRecorder()
	r2, _ := http.NewRequest("DELETE", "/api/storage/pool1/volumes/"+vol.ID, nil)
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = r2
	c2.Params = []gin.Param{
		{Key: "pool_id", Value: "pool1"},
		{Key: "volume_id", Value: vol.ID},
	}
	c2.Set("request_id", "test-123")

	h.Delete(c2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w2.Code)
	}
}

func TestVolumeDeleteNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVolumeHandler()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("DELETE", "/api/storage/pool1/volumes/nonexistent", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{
		{Key: "pool_id", Value: "pool1"},
		{Key: "volume_id", Value: "nonexistent"},
	}
	c.Set("request_id", "test-123")

	h.Delete(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestSanitizeVolumeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"disk.qcow2", "disk.qcow2"},
		{"../etc/passwd", "etcpasswd"},
		{"foo/bar", "foobar"},
		{"foo\\bar", "foobar"},
		{"  spaces  ", "spaces"},
		{"", ""},
	}

	for _, tt := range tests {
		result := sanitizeVolumeName(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeVolumeName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestVolumeCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewVolumeHandler()

	if h.Count() != 0 {
		t.Fatalf("expected 0 volumes, got %d", h.Count())
	}

	req := struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}{Name: "disk.qcow2", Size: 10737418240}
	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/storage/pool1/volumes", bytes.NewBuffer(body))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Params = []gin.Param{{Key: "pool_id", Value: "pool1"}}
	c.Set("request_id", "test-123")
	h.Create(c)

	if h.Count() != 1 {
		t.Fatalf("expected 1 volume, got %d", h.Count())
	}
}