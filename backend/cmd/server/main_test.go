package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/agents"
	"github.com/Priyasharma620064/contextmesh/backend/pkg/k8s"
	"github.com/Priyasharma620064/contextmesh/backend/pkg/mcp"
	"github.com/Priyasharma620064/contextmesh/backend/pkg/storage"
)

func newTestServer() *Server {
	orchestrator := agents.NewAgentOrchestrator()
	return &Server{
		orchestrator: orchestrator,
		storage:      storage.NewStorageEngine(),
		mcpServer:    mcp.NewMcpServer(orchestrator),
		k8sIntel:     k8s.NewK8sIntelligence(),
	}
}

func TestHealthzEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	handler.ServeHTTP(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("Expected 'ok', got '%s'", string(body))
	}
}

func TestReadyzEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	handler.ServeHTTP(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}

func TestHandleParseConfigSuccess(t *testing.T) {
	srv := newTestServer()

	agentsMD := `repo_url: "https://github.com/test/repo"
retrieval:
  prioritize:
    - docs/
  ignore:
    - temp/
policies:
  citation_required: true`

	body, _ := json.Marshal(map[string]string{"content": agentsMD})
	req := httptest.NewRequest(http.MethodPost, "/api/agents/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleParseConfig(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["valid"] != true {
		t.Errorf("Expected valid=true, got %v", result["valid"])
	}
}

func TestHandleParseConfigMethodNotAllowed(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/agents/config", nil)
	w := httptest.NewRecorder()

	srv.handleParseConfig(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", resp.StatusCode)
	}
}

func TestHandleParseConfigInvalidJSON(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodPost, "/api/agents/config", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleParseConfig(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleTriggerSyncSuccess(t *testing.T) {
	srv := newTestServer()

	body, _ := json.Marshal(map[string]string{
		"repo_url": "https://github.com/Priyasharma620064/contextmesh",
		"branch":   "main",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/sync/trigger", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleTriggerSync(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["sync_id"] == nil {
		t.Error("Expected sync_id in response")
	}
}

func TestHandleGetSyncStatus(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/sync/status?repo=https://github.com/Priyasharma620064/contextmesh", nil)
	w := httptest.NewRecorder()

	srv.handleGetSyncStatus(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}

func TestHandleQueryContextSuccess(t *testing.T) {
	srv := newTestServer()

	body, _ := json.Marshal(map[string]interface{}{
		"query":        "architecture goroutine",
		"repo_url":     "https://github.com/Priyasharma620064/contextmesh",
		"branch":       "main",
		"max_tokens":   4000,
		"bypass_cache": true,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleQueryContext(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["compiled_context"] == nil {
		t.Error("Expected compiled_context in response")
	}
}

func TestHandleValidateK8s(t *testing.T) {
	srv := newTestServer()

	body, _ := json.Marshal(map[string]string{
		"manifest_content": "apiVersion: extensions/v1beta1\nkind: Deployment\nmetadata:\n  name: test",
		"file_path":        "test.yaml",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/k8s/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleValidateK8s(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["is_valid"] == true {
		t.Error("Expected is_valid=false for deprecated API")
	}
}

func TestHandleRunBenchmark(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/eval/benchmark?repo=https://github.com/Priyasharma620064/contextmesh&branch=main", nil)
	w := httptest.NewRecorder()

	srv.handleRunBenchmark(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}

func TestHandleGetStorageSchema(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/storage/schema", nil)
	w := httptest.NewRecorder()

	srv.handleGetStorageSchema(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	if result["schema"] == "" {
		t.Error("Expected non-empty schema")
	}
}

func TestHandleGetStorageStats(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/storage/stats", nil)
	w := httptest.NewRecorder()

	srv.handleGetStorageStats(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}

func TestHandleMcpRequestSuccess(t *testing.T) {
	srv := newTestServer()

	rpcBody, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	})
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(rpcBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleMcpRequest(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}

func TestHandleMcpRequestMethodNotAllowed(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	w := httptest.NewRecorder()

	srv.handleMcpRequest(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", resp.StatusCode)
	}
}

func TestParseRequestHelper(t *testing.T) {
	srv := newTestServer()

	// Test with valid JSON
	body, _ := json.Marshal(map[string]string{"content": "test"})
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	w := httptest.NewRecorder()

	var parsedReq struct {
		Content string `json:"content"`
	}

	ok := srv.parseRequest(w, req, &parsedReq)
	if !ok {
		t.Error("Expected parseRequest to succeed")
	}
	if parsedReq.Content != "test" {
		t.Errorf("Expected content 'test', got '%s'", parsedReq.Content)
	}
}

func TestParseRequestHelperGetMethod(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	var parsedReq struct{}
	ok := srv.parseRequest(w, req, &parsedReq)
	if ok {
		t.Error("Expected parseRequest to reject GET method")
	}
}

func TestJsonSuccessAndError(t *testing.T) {
	srv := newTestServer()

	// Test jsonSuccess
	w := httptest.NewRecorder()
	srv.jsonSuccess(w, map[string]string{"status": "ok"})
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected application/json content type, got '%s'", resp.Header.Get("Content-Type"))
	}

	// Test jsonError
	w = httptest.NewRecorder()
	srv.jsonError(w, "test error", http.StatusBadRequest)
	resp = w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.StatusCode)
	}

	var errResult map[string]string
	json.NewDecoder(resp.Body).Decode(&errResult)
	if errResult["error"] != "test error" {
		t.Errorf("Expected error 'test error', got '%s'", errResult["error"])
	}
}
