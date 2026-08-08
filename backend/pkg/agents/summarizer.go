package agents

import (
	"context"
	"fmt"
	"strings"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/proto"
)

type SummarizerAgent struct{}

func NewSummarizerAgent() *SummarizerAgent {
	return &SummarizerAgent{}
}

// SummarizeChunks generates a cohesive structural snapshot of the retrieved chunks.
// It outlines the sources, active lines, and pulls important structural headers.
func (sa *SummarizerAgent) SummarizeChunks(ctx context.Context, chunks []proto.TextChunk) (string, error) {
	if len(chunks) == 0 {
		return "No active knowledge context retrieved.", nil
	}

	var sb strings.Builder
	sb.WriteString("This context includes knowledge synthesized from the following source documents:\n")

	// Group and compile citations
	sourceMap := make(map[string][]string)
	for _, c := range chunks {
		citationLabel := fmt.Sprintf("CIT-%s (Lines %d-%d)", c.ID[:6], c.StartLine, c.EndLine)
		sourceMap[c.Filepath] = append(sourceMap[c.Filepath], citationLabel)

		// Check if we hit timeout
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
	}

	for path, citations := range sourceMap {
		sb.WriteString(fmt.Sprintf("- File: `/%s` | Citations: %s\n", path, strings.Join(citations, ", ")))
	}

	sb.WriteString("\nKey structural headings resolved across documentation:\n")
	headingCount := 0
	for _, c := range chunks {
		lines := strings.Split(c.Text, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") {
				sb.WriteString(fmt.Sprintf("  * %s (in `/%s`)\n", line, c.Filepath))
				headingCount++
			}
		}
	}

	if headingCount == 0 {
		sb.WriteString("  * [General text content - no explicit Markdown headers resolved]\n")
	}

	return sb.String(), nil
}
