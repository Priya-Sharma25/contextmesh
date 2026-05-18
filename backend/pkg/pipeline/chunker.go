package pipeline

import (
	"fmt"
	"strings"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/proto"
)

type MarkdownChunker struct{}

func NewMarkdownChunker() *MarkdownChunker {
	return &MarkdownChunker{}
}

// ChunkDocument splits a markdown file into hierarchical semantic chunks.
// It tracks current heading contexts (e.g. "# Overview > ## Installation") and injects
// this active lineage directly into the metadata of child content.
func (mc *MarkdownChunker) ChunkDocument(content string, filepath string) []proto.TextChunk {
	var chunks []proto.TextChunk
	lines := strings.Split(content, "\n")

	var activeHeadings [6]string // Supports h1 through h6 nesting
	var currentChunkText []string
	var chunkStartLine int32 = 1

	for idx, line := range lines {
		lineNum := int32(idx + 1)
		trimmed := strings.TrimSpace(line)

		// Check if this is a header line (e.g., "# H1" or "## H2")
		if strings.HasPrefix(trimmed, "#") {
			// Save current active block before resetting headers
			if len(currentChunkText) > 0 {
				chunks = append(chunks, mc.buildChunk(currentChunkText, filepath, chunkStartLine, lineNum-1, activeHeadings))
				currentChunkText = nil
			}

			// Parse heading level and text
			level := 0
			for level < len(trimmed) && trimmed[level] == '#' {
				level++
			}
			headerText := strings.TrimSpace(trimmed[level:])

			// Update heading hierarchy
			if level > 0 && level <= 6 {
				activeHeadings[level-1] = headerText
				// Clear lower-level children
				for i := level; i < 6; i++ {
					activeHeadings[i] = ""
				}
			}
			chunkStartLine = lineNum
		}

		currentChunkText = append(currentChunkText, line)
	}

	// Dump remaining final block
	if len(currentChunkText) > 0 {
		chunks = append(chunks, mc.buildChunk(currentChunkText, filepath, chunkStartLine, int32(len(lines)), activeHeadings))
	}

	return chunks
}

func (mc *MarkdownChunker) buildChunk(lines []string, filepath string, startLine, endLine int32, headings [6]string) proto.TextChunk {
	text := strings.Join(lines, "\n")
	
	// Create active breadcrumb hierarchy (e.g., "Core Architecture > Initialization")
	var breadcrumbs []string
	for _, h := range headings {
		if h != "" {
			breadcrumbs = append(breadcrumbs, h)
		}
	}
	hierarchy := strings.Join(breadcrumbs, " > ")
	if hierarchy == "" {
		hierarchy = "General Content"
	}

	// Generate clean ID from filepath and lines
	id := fmt.Sprintf("%s-L%d-L%d", filepath, startLine, endLine)

	return proto.TextChunk{
		ID:        id,
		Text:      text,
		Filepath:  filepath,
		StartLine: startLine,
		EndLine:   endLine,
		Score:     0.0,
		Metadata: map[string]string{
			"hierarchy": hierarchy,
			"length":    fmt.Sprintf("%d", len(text)),
		},
	}
}
