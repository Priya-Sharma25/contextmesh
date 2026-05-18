package agents

import (
	"context"
	"math"
	"strings"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/proto"
)

type RetrieverAgent struct {
	syncAgent *SyncAgent
}

func NewRetrieverAgent(syncAgent *SyncAgent) *RetrieverAgent {
	return &RetrieverAgent{
		syncAgent: syncAgent,
	}
}

// Retrieve searches the active SyncAgent indexing store.
// It scores chunks using TF-IDF lexical overlap combined with a simulated cosine similarity vector.
func (ra *RetrieverAgent) Retrieve(ctx context.Context, query string, repoURL string, branch string) ([]proto.TextChunk, error) {
	// Retrieve all indexed chunks for this repository branch
	allChunks := ra.syncAgent.GetIndexedChunks(repoURL, branch)
	if len(allChunks) == 0 {
		return nil, nil
	}

	// Fetch priority/ignore settings resolved from the repository's AGENTS.md
	priorityPaths := []string{}
	ignorePaths := []string{}
	config := ra.syncAgent.GetConfig(repoURL)
	if config != nil {
		priorityPaths = config.Prioritize
		ignorePaths = config.Ignore
		// Check branch overrides
		for _, sp := range config.ScopedPolicies {
			if sp.Branch == branch {
				priorityPaths = append(priorityPaths, sp.PrioritizePaths...)
				ignorePaths = append(ignorePaths, sp.IgnorePaths...)
			}
		}
	}

	var results []proto.TextChunk
	queryTerms := strings.Fields(strings.ToLower(query))

	for _, chunk := range allChunks {
		// 1. Enforce AGENTS.md ignore paths
		shouldIgnore := false
		for _, ip := range ignorePaths {
			if ip != "" && strings.Contains(chunk.Filepath, ip) {
				shouldIgnore = true
				break
			}
		}
		if shouldIgnore {
			continue
		}

		// 2. Lexical Scoring (TF-IDF approximation)
		lexicalScore := 0.0
		textLower := strings.ToLower(chunk.Text)
		for _, term := range queryTerms {
			count := strings.Count(textLower, term)
			if count > 0 {
				// Simple TF term frequency log curve
				lexicalScore += math.Log1p(float64(count))
			}
		}

		// Normalize lexical score
		if len(queryTerms) > 0 {
			lexicalScore = lexicalScore / float64(len(queryTerms))
		}

		// 3. Dense Semantic Similarity Simulation (Cosine similarity)
		// We simulate cosine similarity by calculating high-dimensional character bigram overlap
		semanticScore := ra.computeBigramOverlap(strings.ToLower(query), textLower)

		// 4. Hybrid Fusion (70% semantic + 30% lexical)
		hybridScore := (0.7 * semanticScore) + (0.3 * lexicalScore)

		// 5. Apply AGENTS.md priority pathway boosting
		isPrioritized := false
		for _, pp := range priorityPaths {
			if pp != "" && strings.Contains(chunk.Filepath, pp) {
				isPrioritized = true
				break
			}
		}
		if isPrioritized {
			hybridScore += 0.25 // 25% Boost for prioritizations
		}

		// Clamp the final relevance score [0.0, 1.0]
		if hybridScore > 1.0 {
			hybridScore = 1.0
		}

		// Filter low relevance matches
		if hybridScore > 0.15 {
			chunkCopy := chunk
			chunkCopy.Score = hybridScore
			results = append(results, chunkCopy)
		}
	}

	// Rerank and sort chunks by score descending
	ra.rerank(results)

	// Token budget capping - limit to top 5 results for query context matching
	maxResults := 5
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results, nil
}

func (ra *RetrieverAgent) computeBigramOverlap(s1, s2 string) float64 {
	b1 := ra.getBigrams(s1)
	b2 := ra.getBigrams(s2)

	if len(b1) == 0 || len(b2) == 0 {
		return 0.0
	}

	intersection := 0
	for bg := range b1 {
		if b2[bg] {
			intersection++
		}
	}

	// Cosine similarity equivalent for sets (Ochiai coefficient)
	return float64(intersection) / math.Sqrt(float64(len(b1)*len(b2)))
}

func (ra *RetrieverAgent) getBigrams(s string) map[string]bool {
	bigrams := make(map[string]bool)
	runes := []rune(s)
	for i := 0; i < len(runes)-1; i++ {
		if runes[i] != ' ' && runes[i+1] != ' ' {
			bigrams[string(runes[i:i+2])] = true
		}
	}
	return bigrams
}

func (ra *RetrieverAgent) rerank(chunks []proto.TextChunk) {
	// Simple bubble sort for descending order
	n := len(chunks)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if chunks[j].Score < chunks[j+1].Score {
				chunks[j], chunks[j+1] = chunks[j+1], chunks[j]
			}
		}
	}
}
