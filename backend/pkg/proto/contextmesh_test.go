package proto

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAgentsConfigSerialization(t *testing.T) {
	config := AgentsConfig{
		RepoURL:    "https://github.com/test/repo",
		Prioritize: []string{"docs/", "tutorials/"},
		Ignore:     []string{"temp/", "deprecated/"},
		Policies:   map[string]bool{"citation_required": true},
		ScopedPolicies: []ScopedPolicy{
			{
				Branch:          "main",
				PrioritizePaths: []string{"docs/core/"},
				IgnorePaths:     []string{"deprecated/alpha/"},
			},
		},
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal AgentsConfig: %v", err)
	}

	var deserialized AgentsConfig
	if err := json.Unmarshal(data, &deserialized); err != nil {
		t.Fatalf("Failed to unmarshal AgentsConfig: %v", err)
	}

	if deserialized.RepoURL != config.RepoURL {
		t.Errorf("Expected RepoURL '%s', got '%s'", config.RepoURL, deserialized.RepoURL)
	}
	if len(deserialized.Prioritize) != 2 {
		t.Errorf("Expected 2 prioritize paths, got %d", len(deserialized.Prioritize))
	}
	if len(deserialized.ScopedPolicies) != 1 {
		t.Errorf("Expected 1 scoped policy, got %d", len(deserialized.ScopedPolicies))
	}
}

func TestTextChunkSerialization(t *testing.T) {
	chunk := TextChunk{
		ID:        "chunk-abc123",
		Text:      "# Architecture\n\nThe system uses goroutines.",
		Filepath:  "docs/architecture.md",
		StartLine: 1,
		EndLine:   3,
		Score:     0.95,
		Metadata:  map[string]string{"hierarchy": "Core > Architecture"},
	}

	data, err := json.Marshal(chunk)
	if err != nil {
		t.Fatalf("Failed to marshal TextChunk: %v", err)
	}

	var deserialized TextChunk
	json.Unmarshal(data, &deserialized)

	if deserialized.ID != "chunk-abc123" {
		t.Errorf("Expected ID 'chunk-abc123', got '%s'", deserialized.ID)
	}
	if deserialized.Score != 0.95 {
		t.Errorf("Expected Score 0.95, got %f", deserialized.Score)
	}
	if deserialized.Metadata["hierarchy"] != "Core > Architecture" {
		t.Errorf("Expected hierarchy metadata, got '%s'", deserialized.Metadata["hierarchy"])
	}
}

func TestQueryContextRequestDefaults(t *testing.T) {
	reqJSON := `{"query": "test query", "repo_url": "https://repo.com", "branch": "main"}`

	var req QueryContextRequest
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if req.Query != "test query" {
		t.Errorf("Expected query 'test query', got '%s'", req.Query)
	}
	if req.MaxTokens != 0 {
		t.Errorf("Expected MaxTokens 0 (default), got %d", req.MaxTokens)
	}
	if req.BypassCache != false {
		t.Error("Expected BypassCache false by default")
	}
}

func TestK8sViolationSerialization(t *testing.T) {
	violation := K8sViolation{
		FilePath:     "deploy/app.yaml",
		APIVersion:   "extensions/v1beta1",
		Kind:         "Deployment",
		Severity:     "CRITICAL",
		Message:      "Deprecated API version detected",
		SuggestedFix: "Migrate to apps/v1",
	}

	data, err := json.Marshal(violation)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var deserialized K8sViolation
	json.Unmarshal(data, &deserialized)

	if deserialized.Severity != "CRITICAL" {
		t.Errorf("Expected CRITICAL severity, got '%s'", deserialized.Severity)
	}
	if deserialized.Kind != "Deployment" {
		t.Errorf("Expected Kind 'Deployment', got '%s'", deserialized.Kind)
	}
}

func TestBenchmarkMetricsSerialization(t *testing.T) {
	metrics := BenchmarkMetrics{
		PrecisionAtK:       0.85,
		RecallAtK:          0.92,
		HallucinationRisk:  0.10,
		ChunkQuality:       0.88,
		CitationConfidence: 0.95,
		LatencyMs:          12.5,
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var deserialized BenchmarkMetrics
	json.Unmarshal(data, &deserialized)

	if deserialized.PrecisionAtK != 0.85 {
		t.Errorf("Expected PrecisionAtK 0.85, got %f", deserialized.PrecisionAtK)
	}
	if deserialized.LatencyMs != 12.5 {
		t.Errorf("Expected LatencyMs 12.5, got %f", deserialized.LatencyMs)
	}
}

func TestRunBenchmarkResponseSerialization(t *testing.T) {
	resp := RunBenchmarkResponse{
		RunID:     "abc12345",
		SuiteName: "Test Suite",
		AverageMetrics: BenchmarkMetrics{
			PrecisionAtK: 0.90,
		},
		Comparisons: []SideBySideComparison{
			{
				StrategyAName: "Strategy A",
				StrategyBName: "Strategy B",
			},
		},
		CompletedAt: time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var deserialized RunBenchmarkResponse
	json.Unmarshal(data, &deserialized)

	if deserialized.RunID != "abc12345" {
		t.Errorf("Expected RunID 'abc12345', got '%s'", deserialized.RunID)
	}
	if len(deserialized.Comparisons) != 1 {
		t.Errorf("Expected 1 comparison, got %d", len(deserialized.Comparisons))
	}
}

func TestParseAgentsConfigRequestSerialization(t *testing.T) {
	reqJSON := `{"content": "repo_url: https://test.com"}`

	var req ParseAgentsConfigRequest
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if req.Content != "repo_url: https://test.com" {
		t.Errorf("Unexpected content: '%s'", req.Content)
	}
}

func TestValidateK8sManifestRequestSerialization(t *testing.T) {
	req := ValidateK8sManifestRequest{
		ManifestContent:  "apiVersion: apps/v1\nkind: Deployment",
		FilePath:         "deploy/app.yaml",
		TargetK8sVersion: "v1.25",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var deserialized ValidateK8sManifestRequest
	json.Unmarshal(data, &deserialized)

	if deserialized.TargetK8sVersion != "v1.25" {
		t.Errorf("Expected 'v1.25', got '%s'", deserialized.TargetK8sVersion)
	}
}
