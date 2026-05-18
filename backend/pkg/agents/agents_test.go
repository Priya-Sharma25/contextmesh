package agents

import (
	"context"
	"strings"
	"testing"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/proto"
)

func TestRetrieverAndSyncAgents(t *testing.T) {
	syncAgent := NewSyncAgent()
	repoURL := "https://github.com/Priyasharma620064/contextmesh"

	// 1. Check indexed chunks
	chunksMain := syncAgent.GetIndexedChunks(repoURL, "main")
	if len(chunksMain) == 0 {
		t.Fatal("Expected prepopulated chunks on main branch, got none")
	}

	// 2. Query Retriever
	retriever := NewRetrieverAgent(syncAgent)
	ctx := context.Background()

	// Query that matches "architecture" content
	results, err := retriever.Retrieve(ctx, "architecture and goroutines", repoURL, "main")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected retriever to return matches for 'architecture', got 0")
	}

	// Top result should be from architecture.md since it contains "Architecture" and "goroutines"
	topResult := results[0]
	if !strings.Contains(topResult.Filepath, "architecture.md") {
		t.Errorf("Expected top match from 'architecture.md', got: %s", topResult.Filepath)
	}

	// Check if priority path boosting works (docs/core/ is prioritized on main branch in mock prepopulate)
	if topResult.Score < 0.25 {
		t.Errorf("Expected top prioritized match to get a score boost, got score: %f", topResult.Score)
	}
}

func TestValidationAgent(t *testing.T) {
	validator := NewValidationAgent()
	ctx := context.Background()

	// Test 1: Validate malformed YAML
	chunkMalformed := proto.TextChunk{
		ID:       "test-1",
		Text:     "Some tutorial text\n\n```yaml\napiVersion: apps/v1\n  kind: Deployment\n   spec: malformed indentation\n```",
		Filepath: "docs/malformed.md",
	}
	violations := validator.ValidateChunk(ctx, chunkMalformed)
	foundYAML := false
	for _, v := range violations {
		if v.Kind == "YAMLConfig" && v.Severity == "CRITICAL" {
			foundYAML = true
			if !strings.Contains(v.Message, "error") {
				t.Errorf("Expected syntax error message, got: %s", v.Message)
			}
		}
	}
	if !foundYAML {
		t.Error("Expected to find YAML indentation syntax violation")
	}

	// Test 2: Validate Deprecated API
	chunkDeprecated := proto.TextChunk{
		ID:       "test-2",
		Text:     "Installing volcano...\n\n```yaml\napiVersion: extensions/v1beta1\nkind: Deployment\nmetadata:\n  name: volcano-scheduler\n```",
		Filepath: "deploy/volcano.yaml",
	}
	violations = validator.ValidateChunk(ctx, chunkDeprecated)
	foundDep := false
	for _, v := range violations {
		if v.Kind == "Deployment" && v.APIVersion == "extensions/v1beta1" {
			foundDep = true
			if v.Severity != "CRITICAL" {
				t.Errorf("Expected CRITICAL severity, got: %s", v.Severity)
			}
			if !strings.Contains(v.SuggestedFix, "apps/v1") {
				t.Errorf("Expected suggestion to use 'apps/v1', got: %s", v.SuggestedFix)
			}
		}
	}
	if !foundDep {
		t.Error("Expected to detect deprecated extensions/v1beta1 deployment API")
	}

	// Test 3: Validate Broken Relative Link
	chunkBrokenLink := proto.TextChunk{
		ID:       "test-3",
		Text:     "Check our [Stale Manual](file:///docs/deprecated/old-manual.tmp) or [Broken URL](http://localhost:8080/broken-link) guidelines.",
		Filepath: "docs/index.md",
	}
	violations = validator.ValidateChunk(ctx, chunkBrokenLink)
	linkCount := 0
	for _, v := range violations {
		if v.Kind == "MarkdownLink" && v.Severity == "WARNING" {
			linkCount++
		}
	}
	if linkCount != 2 {
		t.Errorf("Expected 2 broken link warnings, got: %d", linkCount)
	}
}

func TestSummarizerAgent(t *testing.T) {
	summarizer := NewSummarizerAgent()
	ctx := context.Background()

	chunks := []proto.TextChunk{
		{
			ID:        "abc123456",
			Text:      "# Core Architecture\n\nThe context core orchestrates autonomous operations.",
			Filepath:  "docs/architecture.md",
			StartLine: 1,
			EndLine:   3,
		},
		{
			ID:        "xyz789012",
			Text:      "## Quick installation guide\n\nRun the helm install script.",
			Filepath:  "docs/install.md",
			StartLine: 4,
			EndLine:   6,
		},
	}

	summary, err := summarizer.SummarizeChunks(ctx, chunks)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !strings.Contains(summary, "architecture.md") || !strings.Contains(summary, "install.md") {
		t.Error("Expected summary to list active source files")
	}
	if !strings.Contains(summary, "CIT-abc123") {
		t.Error("Expected summary to list citation tags matching chunk ID hashes")
	}
	if !strings.Contains(summary, "Core Architecture") {
		t.Error("Expected summary to outline parsed markdown header structures")
	}
}

func TestOrchestratorParallelExecution(t *testing.T) {
	orchestrator := NewAgentOrchestrator()
	ctx := context.Background()

	req := &proto.QueryContextRequest{
		Query:       "architecture goroutine templates",
		RepoURL:     "https://github.com/Priyasharma620064/contextmesh",
		Branch:      "main",
		MaxTokens:   4000,
		BypassCache: false,
	}

	res, err := orchestrator.ExecuteOrchestration(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error during parallel agent execution, got %v", err)
	}

	if len(res.RetrievedChunks) == 0 {
		t.Fatal("Expected retrieved chunks in final compiled context")
	}

	if res.TotalTokens == 0 {
		t.Error("Expected token count metrics to be tracked")
	}

	if res.OverallRelevanceScore <= 0.0 {
		t.Error("Expected a positive relevance ranking average score")
	}

	// Verify that anomalies were detected in the compiled output
	// because getting-started.md has a deprecated file reference in its test content
	if !strings.Contains(res.CompiledContext, "CONTEXT SUMMARY") {
		t.Error("Expected compiled prompt context to contain synthesized summaries")
	}
	if !strings.Contains(res.CompiledContext, "SEMANTIC KNOWLEDGE SOURCE CHUNKS") {
		t.Error("Expected compiled prompt context to list citation blocks")
	}
}
