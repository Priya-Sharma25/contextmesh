package pipeline

import (
	"strings"
	"testing"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/proto"
)

func TestMarkdownChunker(t *testing.T) {
	chunker := NewMarkdownChunker()

	docContent := `
# Platform Manual

Welcome to the context mesh.

## Section 1: Orchestration
The orchestrator drives parallel runs.

### Section 1.1: goroutines
Goroutines utilize channels for synchronization.

# Quick Reference
Key concepts and layouts.
`

	chunks := chunker.ChunkDocument(docContent, "docs/manual.md")

	// There should be at least 3 chunks: Header 1 section, Subsection 1.1 section, and Section 2 section.
	if len(chunks) < 3 {
		t.Fatalf("Expected at least 3 chunks, got: %d", len(chunks))
	}

	foundSub := false
	for _, c := range chunks {
		if strings.Contains(c.Text, "Goroutines utilize channels") {
			foundSub = true
			hierarchy := c.Metadata["hierarchy"]
			expected := "Platform Manual > Section 1: Orchestration > Section 1.1: goroutines"
			if hierarchy != expected {
				t.Errorf("Expected nested hierarchy breadcrumbs to be '%s', got: '%s'", expected, hierarchy)
			}
		}
	}

	if !foundSub {
		t.Error("Expected to find the subsection chunk with its nested metadata hierarchy")
	}
}

func TestTokenCompressor(t *testing.T) {
	compressor := NewTokenCompressor()

	chunks := []proto.TextChunk{
		{
			ID:   "c1",
			Text: "This is a brief first chunk of text.", // ~10 tokens
		},
		{
			ID:   "c2",
			Text: "This is another chunk that represents middle details.", // ~12 tokens
		},
		{
			ID:   "c3",
			Text: "This is a very long third chunk of text designed to contain substantial detailed parameters and configuration data that takes up a lot of space in the context window.", // ~38 tokens
		},
	}

	// 1. Fully fits within budget
	accepted, ratio := compressor.Compress(chunks, 100)
	if len(accepted) != 3 {
		t.Errorf("Expected all 3 chunks to be accepted, got: %d", len(accepted))
	}
	if ratio != 1.0 {
		t.Errorf("Expected compression ratio of 1.0, got: %f", ratio)
	}

	// 2. Fits only top 2 chunks under strict 25 token budget
	acceptedStrict, ratioStrict := compressor.Compress(chunks, 25)
	if len(acceptedStrict) != 2 {
		t.Errorf("Expected only 2 chunks to be accepted under strict budget, got: %d", len(acceptedStrict))
	}
	if ratioStrict >= 1.0 || ratioStrict == 0.0 {
		t.Errorf("Expected compression ratio to be fractional, got: %f", ratioStrict)
	}
}
