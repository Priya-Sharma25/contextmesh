package agents

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/proto"
)

type SyncAgent struct {
	mu           sync.RWMutex
	syncStates   map[string][]proto.RepositorySyncState // repoURL -> state slices
	indexStore   map[string][]proto.TextChunk          // "repoURL:branch" -> chunks
	agentsConfig map[string]*proto.AgentsConfig        // repoURL -> config
}

func NewSyncAgent() *SyncAgent {
	sa := &SyncAgent{
		syncStates:   make(map[string][]proto.RepositorySyncState),
		indexStore:   make(map[string][]proto.TextChunk),
		agentsConfig: make(map[string]*proto.AgentsConfig),
	}
	sa.prepopulateMockKnowledge()
	return sa
}

func (sa *SyncAgent) GetSyncStatus(repoURL string) []proto.RepositorySyncState {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	return sa.syncStates[repoURL]
}

func (sa *SyncAgent) GetIndexedChunks(repoURL, branch string) []proto.TextChunk {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	key := fmt.Sprintf("%s:%s", repoURL, branch)
	return sa.indexStore[key]
}

func (sa *SyncAgent) GetConfig(repoURL string) *proto.AgentsConfig {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	return sa.agentsConfig[repoURL]
}

// TriggerSync simulates high-fidelity branch-aware delta indexing with hashing.
func (sa *SyncAgent) TriggerSync(repoURL string, branch string, release string) (*proto.RepositorySyncState, error) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	// 1. Resolve active mock documents for this repo/branch
	docs := sa.getMockDocuments(repoURL, branch, release)
	
	// Create or update sync states
	states := sa.syncStates[repoURL]
	var currentState *proto.RepositorySyncState
	for i := range states {
		if states[i].Branch == branch {
			currentState = &states[i]
			break
		}
	}

	// Calculate file deltas using simple simulated content hashing (MD5)
	filesChanged := int64(0)
	filesUnchanged := int64(0)
	var newChunks []proto.TextChunk

	for _, doc := range docs {
		hasher := md5.New()
		hasher.Write([]byte(doc.content))
		contentHash := hex.EncodeToString(hasher.Sum(nil))

		// Check against pre-existing files (simulated delta logic)
		isChanged := true
		if currentState != nil && currentState.SyncStatus == "COMPLETED" {
			// In our simulation, odd-numbered ticks trigger modification
			if time.Now().UnixNano()%2 == 0 {
				isChanged = false
			}
		}

		if isChanged {
			filesChanged++
		} else {
			filesUnchanged++
		}

		// Chunk document (split by paragraph boundaries for simplified semantic segmentation)
		paragraphs := strings.Split(doc.content, "\n\n")
		lineOffset := int32(1)
		for pIdx, p := range paragraphs {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			pLines := int32(strings.Count(p, "\n") + 1)
			
			chunkID := fmt.Sprintf("%s-%s-%s-p%d", repoURL, branch, doc.path, pIdx)
			newChunks = append(newChunks, proto.TextChunk{
				ID:        sa.md5Hash(chunkID),
				Text:      p,
				Filepath:  doc.path,
				StartLine: lineOffset,
				EndLine:   lineOffset + pLines - 1,
				Score:     0.0,
				Metadata: map[string]string{
					"repo_url": repoURL,
					"branch":   branch,
					"release":  release,
					"hash":     contentHash[:8],
				},
			})
			lineOffset += pLines + 1 // Add 1 for the double newline divider
		}
	}

	// Persist chunks in our indexing store
	storeKey := fmt.Sprintf("%s:%s", repoURL, branch)
	sa.indexStore[storeKey] = newChunks

	// Update git metrics and commit log simulation
	hasher := md5.New()
	hasher.Write([]byte(fmt.Sprintf("%s-%s-%d", repoURL, branch, time.Now().UnixNano())))
	latestCommit := hex.EncodeToString(hasher.Sum(nil))

	syncState := proto.RepositorySyncState{
		RepoURL:            repoURL,
		Branch:             branch,
		LastCommitHash:     latestCommit,
		TotalChunksIndexed: int64(len(newChunks)),
		FilesChanged:       filesChanged,
		FilesUnchanged:     filesUnchanged,
		SyncStatus:         "COMPLETED",
		UpdatedAt:          time.Now(),
	}

	if currentState != nil {
		*currentState = syncState
	} else {
		sa.syncStates[repoURL] = append(sa.syncStates[repoURL], syncState)
	}

	return &syncState, nil
}

type mockDoc struct {
	path    string
	content string
}

func (sa *SyncAgent) getMockDocuments(repoURL, branch, release string) []mockDoc {
	// Returns documents adjusted dynamically by branch or release scope
	docs := []mockDoc{
		{
			path: "docs/core/architecture.md",
			content: `# Core Platform Architecture

The ContextMesh platform is designed for cloud-native documentation intelligence. It orchestrates Retriever, Validator, and Summarizer agents.

The orchestration layer is responsive, maintaining parallel goroutines with context boundary timeouts to prevent thread blocking under heavy AI loads.

Semantic caching uses a high-performance Redis cache layer combined with persistent PostgreSQL vector indices (pgvector).`,
		},
		{
			path: "docs/tutorials/getting-started.md",
			content: `# Getting Started Tutorial

Follow these instructions to set up the context engine.

First, place an AGENTS.md configuration in your repository root. Then configure priorities:
[Prioritize Documents](file:///docs/core/architecture.md)

Example manifest deploy config:
` + "```yaml" + `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: contextmesh-engine
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: backend
        image: contextmesh/backend:v1.0.0
` + "```" + ``,
		},
		{
			path: "deploy/k8s/deprecated-app.yaml",
			content: `# Deprecated Manifest Template

This contains an older deployment setup for legacy checks.

` + "```yaml" + `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: legacy-scheduler
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: volcano-compat
        image: volcano/legacy:v0.8
` + "```" + `

Also, check the broken links section:
[Refer to Stale Guidelines](file:///docs/deprecated/old-guidelines.tmp)`,
		},
	}

	// Add release/branch targeted documentation to simulate realistic sync outputs
	if branch == "develop" {
		docs = append(docs, mockDoc{
			path: "docs/experimental/alpha-feature.md",
			content: `# Experimental Features

This contains bleeding-edge drafts for context compression algorithms.
Avoid using this on production clusters.`,
		})
	}
	if release == "v1.0.0" {
		docs = append(docs, mockDoc{
			path: "docs/stable-v1.0/release-notes.md",
			content: `# Stable Release v1.0.0

Official release notes for the stable API. Fully compliant with Kubernetes v1.28 structures.`,
		})
	}

	return docs
}

func (sa *SyncAgent) prepopulateMockKnowledge() {
	defaultRepo := "https://github.com/Priyasharma620064/contextmesh"

	// Pre-parse the sample AGENTS.md we created earlier
	sa.agentsConfig[defaultRepo] = &proto.AgentsConfig{
		RepoURL:    defaultRepo,
		Prioritize: []string{"docs/"},
		Ignore:     []string{"temp/"},
		Policies:   map[string]bool{"citation_required": true},
		ScopedPolicies: []proto.ScopedPolicy{
			{
				Branch:           "main",
				PrioritizePaths:  []string{"docs/core/"},
				CitationRequired: true,
			},
			{
				Branch:           "develop",
				PrioritizePaths:  []string{"docs/experimental/"},
				IgnorePaths:      []string{"deprecated/alpha/"},
				CitationRequired: false,
			},
		},
	}

	// Trigger simulated sync for main branch of our default repo
	_, _ = sa.TriggerSync(defaultRepo, "main", "")
	_, _ = sa.TriggerSync(defaultRepo, "develop", "")
}

func (sa *SyncAgent) md5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
