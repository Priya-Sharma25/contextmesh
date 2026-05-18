package agents

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/proto"
)

// AgentOrchestrator coordinates Retriever, Validator, Summarizer, and Sync agents.
type AgentOrchestrator struct {
	Retriever  *RetrieverAgent
	Validator  *ValidationAgent
	Summarizer *SummarizerAgent
	SyncAgent  *SyncAgent
}

// NewAgentOrchestrator creates a new multi-agent orchestrator instance.
func NewAgentOrchestrator() *AgentOrchestrator {
	syncAgent := NewSyncAgent()
	return &AgentOrchestrator{
		Retriever:  NewRetrieverAgent(syncAgent),
		Validator:  NewValidationAgent(),
		Summarizer: NewSummarizerAgent(),
		SyncAgent:  syncAgent,
	}
}

// ExecuteOrchestration runs Retriever, Validator, and Summarizer agents in parallel
// to compile a high-fidelity context window for an AI query.
func (ao *AgentOrchestrator) ExecuteOrchestration(ctx context.Context, req *proto.QueryContextRequest) (*proto.QueryContextResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var errs []error
	var errMu sync.Mutex

	var chunks []proto.TextChunk
	var validationReports []proto.K8sViolation
	var summaryText string

	// 1. Trigger Retriever Agent (Fetch semantic chunks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		retrieved, err := ao.Retriever.Retrieve(ctx, req.Query, req.RepoURL, req.Branch)
		if err != nil {
			errMu.Lock()
			errs = append(errs, fmt.Errorf("retriever failed: %w", err))
			errMu.Unlock()
			return
		}
		chunks = retrieved
	}()

	// 2. Wait for Retriever chunks because Validator and Summarizer depend on them
	c := make(chan struct{})
	go func() {
		wg.Wait()
		close(c)
	}()

	select {
	case <-c:
		// Retriever finished
	case <-ctx.Done():
		return nil, fmt.Errorf("orchestration timeout waiting for retrieval: %w", ctx.Err())
	}

	if len(errs) > 0 {
		return nil, errs[0]
	}

	if len(chunks) == 0 {
		return &proto.QueryContextResponse{
			CompiledContext:       "No relevant documentation found in active context policies.",
			RetrievedChunks:       nil,
			TotalTokens:           0,
			CompressionRatio:      0.0,
			OverallRelevanceScore: 0.0,
		}, nil
	}

	// 3. Trigger Validation and Summarization in parallel based on chunks
	var wgSecond sync.WaitGroup

	wgSecond.Add(2)
	// A. Validate chunks (Broken links, YAML specs, outdated API warnings)
	go func() {
		defer wgSecond.Done()
		for _, chunk := range chunks {
			violations := ao.Validator.ValidateChunk(ctx, chunk)
			if len(violations) > 0 {
				errMu.Lock()
				validationReports = append(validationReports, violations...)
				errMu.Unlock()
			}
		}
	}()

	// B. Summarize active chunks (Context snapshot & changelog insights)
	go func() {
		defer wgSecond.Done()
		sum, err := ao.Summarizer.SummarizeChunks(ctx, chunks)
		if err != nil {
			errMu.Lock()
			errs = append(errs, fmt.Errorf("summarizer failed: %w", err))
			errMu.Unlock()
			return
		}
		summaryText = sum
	}()

	cSecond := make(chan struct{})
	go func() {
		wgSecond.Wait()
		close(cSecond)
	}()

	select {
	case <-cSecond:
		// Validation and Summarizer completed
	case <-ctx.Done():
		return nil, fmt.Errorf("orchestration timeout during validation/summarization: %w", ctx.Err())
	}

	if len(errs) > 0 {
		return nil, errs[0]
	}

	// 4. Compile the final context window
	// Inject validation alerts, citations, and summaries directly into the AI prompt
	compiledContext := ao.compileContext(chunks, validationReports, summaryText)

	// Calculate overall score (average of top chunk scores)
	var sumScore float64
	for _, ch := range chunks {
		sumScore += ch.Score
	}
	avgScore := 0.0
	if len(chunks) > 0 {
		avgScore = sumScore / float64(len(chunks))
	}

	// Calculate tokens (rough estimates: 1 word = 1.3 tokens)
	totalTokens := int32(len(compiledContext) / 4)
	compRatio := float64(len(compiledContext)) / float64(ao.rawChunksLength(chunks)+1)

	return &proto.QueryContextResponse{
		CompiledContext:       compiledContext,
		RetrievedChunks:       chunks,
		TotalTokens:           totalTokens,
		CompressionRatio:      compRatio,
		OverallRelevanceScore: avgScore,
	}, nil
}

func (ao *AgentOrchestrator) compileContext(chunks []proto.TextChunk, violations []proto.K8sViolation, summary string) string {
	var sb sync.WaitGroup // wait, no need for waitgroup, string builder is fine
	_ = sb

	out := "=== CONTEXT SUMMARY ===\n"
	out += summary + "\n\n"

	if len(violations) > 0 {
		out += "=== DOCUMENTATION ANOMALIES DETECTED ===\n"
		for _, v := range violations {
			out += fmt.Sprintf("[%s] File: %s | Issue: %s | Suggested Fix: %s\n", v.Severity, v.FilePath, v.Message, v.SuggestedFix)
		}
		out += "\n"
	}

	out += "=== SEMANTIC KNOWLEDGE SOURCE CHUNKS ===\n"
	for idx, chunk := range chunks {
		out += fmt.Sprintf("--- Chunk %d | Source: %s [Lines %d-%d] | Citation ID: CIT-%s ---\n",
			idx+1, chunk.Filepath, chunk.StartLine, chunk.EndLine, chunk.ID[:6])
		out += chunk.Text + "\n\n"
	}

	return out
}

func (ao *AgentOrchestrator) rawChunksLength(chunks []proto.TextChunk) int {
	var total int
	for _, c := range chunks {
		total += len(c.Text)
	}
	return total
}
