package eval

import (
	"testing"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/proto"
)

func TestNewBenchmarkSuite(t *testing.T) {
	suite := NewBenchmarkSuite()
	if suite == nil {
		t.Fatal("Expected BenchmarkSuite to be created, got nil")
	}
	if len(suite.goldDataset) != 3 {
		t.Errorf("Expected 3 gold standard queries, got %d", len(suite.goldDataset))
	}
}

func TestRunSuiteWithEmptyChunks(t *testing.T) {
	suite := NewBenchmarkSuite()

	result := suite.RunSuite("https://example.com/repo", "main", nil)
	if result == nil {
		t.Fatal("Expected non-nil response even with empty chunks")
	}
	if result.RunID == "" {
		t.Error("Expected a generated RunID")
	}
	if result.SuiteName == "" {
		t.Error("Expected a SuiteName")
	}
	if len(result.Comparisons) != 3 {
		t.Errorf("Expected 3 comparisons (one per gold query), got %d", len(result.Comparisons))
	}
}

func TestRunSuiteWithActiveChunks(t *testing.T) {
	suite := NewBenchmarkSuite()

	chunks := []proto.TextChunk{
		{
			ID:        "chunk-arch-1",
			Text:      "The orchestrator runs parallel goroutines with a 4-second timeout boundary.",
			Filepath:  "docs/core/architecture.md",
			StartLine: 5,
			EndLine:   10,
			Metadata:  map[string]string{"hierarchy": "Platform Core > Architecture"},
		},
		{
			ID:        "chunk-deploy-1",
			Text:      "Deploy the contextmesh-engine with 3 replicas using v1.0.0 deployment image.",
			Filepath:  "docs/tutorials/getting-started.md",
			StartLine: 8,
			EndLine:   15,
			Metadata:  map[string]string{"hierarchy": "Tutorials > Getting Started"},
		},
	}

	result := suite.RunSuite("https://example.com/repo", "main", chunks)
	if result == nil {
		t.Fatal("Expected non-nil response")
	}

	// Strategy A should outperform Strategy B
	for i, comp := range result.Comparisons {
		if comp.StrategyAName == "" || comp.StrategyBName == "" {
			t.Errorf("Comparison %d: strategy names should be non-empty", i)
		}
	}

	// Average metrics should be populated
	avg := result.AverageMetrics
	if avg.LatencyMs < 0 {
		t.Error("Expected non-negative latency")
	}
	if avg.HallucinationRisk < 0.05 {
		t.Errorf("Expected hallucination risk >= 0.05 (clamped floor), got %f", avg.HallucinationRisk)
	}
}

func TestCalculatePrecisionRecallEmpty(t *testing.T) {
	suite := NewBenchmarkSuite()
	gq := GoldStandardQuery{
		Query:             "test",
		ExpectedFilepath:  "docs/test.md",
		ExpectedLineRange: [2]int32{1, 10},
	}

	prec, rec := suite.calculatePrecisionRecall(gq, nil)
	if prec != 0.0 || rec != 0.0 {
		t.Errorf("Expected (0, 0) for empty results, got (%f, %f)", prec, rec)
	}
}

func TestCalculatePrecisionRecallMatch(t *testing.T) {
	suite := NewBenchmarkSuite()
	gq := GoldStandardQuery{
		Query:             "test",
		ExpectedFilepath:  "docs/test.md",
		ExpectedLineRange: [2]int32{5, 10},
	}

	chunks := []proto.TextChunk{
		{Filepath: "docs/test.md", StartLine: 5, EndLine: 10},
		{Filepath: "docs/other.md", StartLine: 1, EndLine: 5},
	}

	prec, rec := suite.calculatePrecisionRecall(gq, chunks)
	if prec != 0.5 {
		t.Errorf("Expected precision 0.5 (1 match / 2 results), got %f", prec)
	}
	if rec != 1.0 {
		t.Errorf("Expected recall 1.0 (match found), got %f", rec)
	}
}

func TestCalculateHallucinationRiskEmpty(t *testing.T) {
	suite := NewBenchmarkSuite()
	gq := GoldStandardQuery{
		Query:               "test",
		ExpectedIdentifiers: []string{"keyword1", "keyword2"},
	}

	risk := suite.calculateHallucinationRisk(gq, nil)
	if risk != 0.95 {
		t.Errorf("Expected max risk 0.95 for empty results, got %f", risk)
	}
}

func TestCalculateHallucinationRiskPartialCoverage(t *testing.T) {
	suite := NewBenchmarkSuite()
	gq := GoldStandardQuery{
		Query:               "test",
		ExpectedIdentifiers: []string{"orchestrator", "parallel", "goroutines", "timeout"},
	}

	chunks := []proto.TextChunk{
		{Text: "The orchestrator runs parallel tasks."},
	}

	risk := suite.calculateHallucinationRisk(gq, chunks)
	// 2 out of 4 identifiers matched -> coverage 0.5, risk = 0.5 + 0.15 (no hierarchy) = 0.65
	if risk < 0.5 || risk > 0.75 {
		t.Errorf("Expected risk between 0.5-0.75 for partial coverage, got %f", risk)
	}
}

func TestCalculateCitationConfidence(t *testing.T) {
	suite := NewBenchmarkSuite()
	gq := GoldStandardQuery{
		ExpectedFilepath:  "docs/test.md",
		ExpectedLineRange: [2]int32{5, 10},
	}

	// Empty results
	conf := suite.calculateCitationConfidence(gq, nil)
	if conf != 0.0 {
		t.Errorf("Expected 0 confidence for empty, got %f", conf)
	}

	// Non-matching filepath
	conf = suite.calculateCitationConfidence(gq, []proto.TextChunk{
		{Filepath: "docs/other.md", StartLine: 5, EndLine: 10},
	})
	if conf != 0.15 {
		t.Errorf("Expected fallback 0.15 for non-matching file, got %f", conf)
	}

	// Exact match
	conf = suite.calculateCitationConfidence(gq, []proto.TextChunk{
		{Filepath: "docs/test.md", StartLine: 5, EndLine: 10},
	})
	if conf != 1.0 {
		t.Errorf("Expected 1.0 confidence for exact match, got %f", conf)
	}
}
