package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/agents"
)

func newTestServer() *McpServer {
	orchestrator := agents.NewAgentOrchestrator()
	return NewMcpServer(orchestrator)
}

func TestListTools(t *testing.T) {
	srv := newTestServer()
	tools := srv.ListTools()

	if len(tools) != 3 {
		t.Fatalf("Expected 3 tools, got %d", len(tools))
	}

	expectedNames := map[string]bool{
		"semantic_search":  false,
		"validate_k8s_docs": false,
		"run_benchmarks":   false,
	}

	for _, tool := range tools {
		if _, ok := expectedNames[tool.Name]; !ok {
			t.Errorf("Unexpected tool: %s", tool.Name)
		}
		expectedNames[tool.Name] = true
		if tool.Description == "" {
			t.Errorf("Tool %s has empty description", tool.Name)
		}
		if tool.InputSchema.Type != "object" {
			t.Errorf("Tool %s has unexpected schema type: %s", tool.Name, tool.InputSchema.Type)
		}
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("Expected tool %s not found", name)
		}
	}
}

func TestHandleRequestInvalidJSON(t *testing.T) {
	srv := newTestServer()
	ctx := context.Background()

	resp, err := srv.HandleRequest(ctx, []byte("{invalid"))
	if err != nil {
		t.Fatalf("HandleRequest should not return Go error, got %v", err)
	}

	var rpcResp JsonRpcResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if rpcResp.Error == nil {
		t.Fatal("Expected error response for invalid JSON")
	}
	if rpcResp.Error.Code != -32700 {
		t.Errorf("Expected parse error code -32700, got %d", rpcResp.Error.Code)
	}
}

func TestHandleRequestInvalidVersion(t *testing.T) {
	srv := newTestServer()
	ctx := context.Background()

	body, _ := json.Marshal(JsonRpcRequest{
		JsonRpc: "1.0",
		Id:      1,
		Method:  "tools/list",
	})

	resp, err := srv.HandleRequest(ctx, body)
	if err != nil {
		t.Fatalf("HandleRequest should not return Go error, got %v", err)
	}

	var rpcResp JsonRpcResponse
	json.Unmarshal(resp, &rpcResp)

	if rpcResp.Error == nil {
		t.Fatal("Expected error for invalid jsonrpc version")
	}
	if rpcResp.Error.Code != -32600 {
		t.Errorf("Expected invalid request code -32600, got %d", rpcResp.Error.Code)
	}
}

func TestHandleRequestToolsList(t *testing.T) {
	srv := newTestServer()
	ctx := context.Background()

	body, _ := json.Marshal(JsonRpcRequest{
		JsonRpc: "2.0",
		Id:      1,
		Method:  "tools/list",
	})

	resp, err := srv.HandleRequest(ctx, body)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	var rpcResp JsonRpcResponse
	json.Unmarshal(resp, &rpcResp)

	if rpcResp.Error != nil {
		t.Fatalf("Unexpected error: %s", rpcResp.Error.Message)
	}
	if rpcResp.Result == nil {
		t.Fatal("Expected result to be non-nil")
	}
}

func TestHandleRequestMethodNotFound(t *testing.T) {
	srv := newTestServer()
	ctx := context.Background()

	body, _ := json.Marshal(JsonRpcRequest{
		JsonRpc: "2.0",
		Id:      1,
		Method:  "nonexistent/method",
	})

	resp, err := srv.HandleRequest(ctx, body)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	var rpcResp JsonRpcResponse
	json.Unmarshal(resp, &rpcResp)

	if rpcResp.Error == nil {
		t.Fatal("Expected error for unknown method")
	}
	if rpcResp.Error.Code != -32601 {
		t.Errorf("Expected method not found code -32601, got %d", rpcResp.Error.Code)
	}
}

func TestCallToolSemanticSearch(t *testing.T) {
	srv := newTestServer()
	ctx := context.Background()

	result, err := srv.CallTool(ctx, "semantic_search", map[string]interface{}{
		"query":  "architecture goroutine templates",
		"repo":   "https://github.com/Priyasharma620064/contextmesh",
		"branch": "main",
	})
	if err != nil {
		t.Fatalf("CallTool semantic_search failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map result from semantic_search")
	}

	if _, ok := resultMap["content"]; !ok {
		t.Error("Expected 'content' key in result")
	}
	if _, ok := resultMap["metrics"]; !ok {
		t.Error("Expected 'metrics' key in result")
	}
}

func TestCallToolSemanticSearchMissingArgs(t *testing.T) {
	srv := newTestServer()
	ctx := context.Background()

	_, err := srv.CallTool(ctx, "semantic_search", map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected error for missing query/repo args")
	}
}

func TestCallToolValidateK8s(t *testing.T) {
	srv := newTestServer()
	ctx := context.Background()

	manifest := `apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: test-app`

	result, err := srv.CallTool(ctx, "validate_k8s_docs", map[string]interface{}{
		"manifest": manifest,
		"filepath": "test.yaml",
	})
	if err != nil {
		t.Fatalf("CallTool validate_k8s_docs failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map result")
	}

	isValid, ok := resultMap["is_valid"].(bool)
	if !ok {
		t.Fatal("Expected 'is_valid' bool in result")
	}
	if isValid {
		t.Error("Expected is_valid=false for deprecated API")
	}
}

func TestCallToolValidateK8sMissingManifest(t *testing.T) {
	srv := newTestServer()
	ctx := context.Background()

	_, err := srv.CallTool(ctx, "validate_k8s_docs", map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected error for missing manifest argument")
	}
}

func TestCallToolRunBenchmarks(t *testing.T) {
	srv := newTestServer()
	ctx := context.Background()

	result, err := srv.CallTool(ctx, "run_benchmarks", map[string]interface{}{
		"repo":   "https://github.com/Priyasharma620064/contextmesh",
		"branch": "main",
	})
	if err != nil {
		t.Fatalf("CallTool run_benchmarks failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map result")
	}
	if _, ok := resultMap["run_id"]; !ok {
		t.Error("Expected 'run_id' in benchmark result")
	}
}

func TestCallToolRunBenchmarksMissingRepo(t *testing.T) {
	srv := newTestServer()
	ctx := context.Background()

	_, err := srv.CallTool(ctx, "run_benchmarks", map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected error for missing repo argument")
	}
}

func TestCallToolUnknownTool(t *testing.T) {
	srv := newTestServer()
	ctx := context.Background()

	_, err := srv.CallTool(ctx, "nonexistent_tool", map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected error for unknown tool name")
	}
}

func TestHandleRequestToolsCallViaJsonRpc(t *testing.T) {
	srv := newTestServer()
	ctx := context.Background()

	params, _ := json.Marshal(map[string]interface{}{
		"name": "semantic_search",
		"arguments": map[string]interface{}{
			"query":  "architecture",
			"repo":   "https://github.com/Priyasharma620064/contextmesh",
			"branch": "main",
		},
	})

	body, _ := json.Marshal(JsonRpcRequest{
		JsonRpc: "2.0",
		Id:      42,
		Method:  "tools/call",
		Params:  params,
	})

	resp, err := srv.HandleRequest(ctx, body)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	var rpcResp JsonRpcResponse
	json.Unmarshal(resp, &rpcResp)

	if rpcResp.Error != nil {
		t.Fatalf("Unexpected error: %s", rpcResp.Error.Message)
	}
	if rpcResp.Id != float64(42) {
		t.Errorf("Expected ID 42, got %v", rpcResp.Id)
	}
}
