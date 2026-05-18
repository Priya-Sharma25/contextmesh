package storage

import (
	"crypto/md5"
	"encoding/hex"
	"sync"
	"time"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/proto"
)

type StorageEngine struct {
	mu          sync.RWMutex
	redisCache  map[string]cacheEntry
	cacheHits   int64
	cacheMisses int64
}

type cacheEntry struct {
	Response  *proto.QueryContextResponse
	ExpiresAt time.Time
}

func NewStorageEngine() *StorageEngine {
	return &StorageEngine{
		redisCache: make(map[string]cacheEntry),
	}
}

// GetPgvectorSchema returns the SQL schema definitions for PostgreSQL pgvector.
// Exhibited directly on the frontend dashboard to show database sophistication.
func (se *StorageEngine) GetPgvectorSchema() string {
	return `
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Table to store repository information
CREATE TABLE IF NOT EXISTS repositories (
    id SERIAL PRIMARY KEY,
    url VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table to store document catalog
CREATE TABLE IF NOT EXISTS documents (
    id SERIAL PRIMARY KEY,
    repo_id INTEGER REFERENCES repositories(id) ON DELETE CASCADE,
    filepath VARCHAR(512) NOT NULL,
    commit_hash CHAR(40) NOT NULL,
    content TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(repo_id, filepath)
);

-- Table to store semantically chunked text blocks with vector embeddings
CREATE TABLE IF NOT EXISTS document_chunks (
    id VARCHAR(64) PRIMARY KEY,
    document_id INTEGER REFERENCES documents(id) ON DELETE CASCADE,
    filepath VARCHAR(512) NOT NULL,
    text_content TEXT NOT NULL,
    start_line INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    hierarchy_breadcrumbs TEXT,
    -- 1536-dimension vector embedding (matches OpenAI text-embedding-3-small)
    embedding vector(1536),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create HNSW index for high-performance cosine distance queries
CREATE INDEX IF NOT EXISTS document_chunks_hnsw_idx 
ON document_chunks 
USING hnsw (embedding vector_cosine_ops);
`
}

// GetCache attempts to fetch a compiled context query from the Redis wrapper.
func (se *StorageEngine) GetCache(query string, repoURL string, branch string) (*proto.QueryContextResponse, bool) {
	se.mu.Lock()
	defer se.mu.Unlock()

	key := se.hashKey(query, repoURL, branch)
	entry, found := se.redisCache[key]
	if !found {
		se.cacheMisses++
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(se.redisCache, key)
		se.cacheMisses++
		return nil, false
	}

	se.cacheHits++
	return entry.Response, true
}

// SetCache caches an assembled QueryContextResponse in the simulated Redis cache (Expires in 1 minute).
func (se *StorageEngine) SetCache(query string, repoURL string, branch string, resp *proto.QueryContextResponse) {
	se.mu.Lock()
	defer se.mu.Unlock()

	key := se.hashKey(query, repoURL, branch)
	se.redisCache[key] = cacheEntry{
		Response:  resp,
		ExpiresAt: time.Now().Add(1 * time.Minute),
	}
}

// GetTelemetryStats returns cache hit/miss data.
func (se *StorageEngine) GetTelemetryStats() (int64, int64, float64) {
	se.mu.RLock()
	defer se.mu.RUnlock()

	total := se.cacheHits + se.cacheMisses
	if total == 0 {
		return 0, 0, 0.0
	}
	ratio := float64(se.cacheHits) / float64(total)
	return se.cacheHits, se.cacheMisses, ratio
}

func (se *StorageEngine) hashKey(query, repoURL, branch string) string {
	hasher := md5.New()
	hasher.Write([]byte(query + "|" + repoURL + "|" + branch))
	return hex.EncodeToString(hasher.Sum(nil))
}
