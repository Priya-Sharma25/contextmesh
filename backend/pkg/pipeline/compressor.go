package pipeline

import (
	"strings"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/proto"
)

type TokenCompressor struct{}

func NewTokenCompressor() *TokenCompressor {
	return &TokenCompressor{}
}

// Compress fits chunks inside a strict token budget (maxTokens).
// It retains high-relevance chunks and prunes low-scoring ones until the budget is satisfied.
func (tc *TokenCompressor) Compress(chunks []proto.TextChunk, maxTokens int32) ([]proto.TextChunk, float64) {
	if len(chunks) == 0 {
		return nil, 0.0
	}

	var accepted []proto.TextChunk
	var currentTokens int32 = 0
	originalLength := 0
	compressedLength := 0

	for _, chunk := range chunks {
		originalLength += len(chunk.Text)
		
		// Approximate token calculation: 1 word ~ 1.3 tokens
		wordCount := int32(len(strings.Fields(chunk.Text)))
		chunkTokens := int32(float64(wordCount) * 1.3)
		if chunkTokens == 0 {
			chunkTokens = 1
		}

		// Check if we can include this chunk within the token budget
		if currentTokens+chunkTokens <= maxTokens {
			accepted = append(accepted, chunk)
			currentTokens += chunkTokens
			compressedLength += len(chunk.Text)
		}
		// If we exceed token limits, we ignore lower priority chunks
	}

	ratio := 0.0
	if originalLength > 0 {
		ratio = float64(compressedLength) / float64(originalLength)
	}

	return accepted, ratio
}
