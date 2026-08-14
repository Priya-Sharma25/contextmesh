package storage

import (
	"testing"
	"time"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/proto"
)

func TestNewStorageEngine(t *testing.T) {
	se := NewStorageEngine()
	if se == nil {
		t.Fatal("Expected non-nil StorageEngine")
	}
	if se.redisCache == nil {
		t.Fatal("Expected initialized cache map")
	}
}

func TestGetPgvectorSchema(t *testing.T) {
	se := NewStorageEngine()
	schema := se.GetPgvectorSchema()

	if schema == "" {
		t.Fatal("Expected non-empty schema string")
	}

	requiredKeywords := []string{
		"CREATE EXTENSION",
		"vector",
		"repositories",
		"documents",
		"document_chunks",
		"embedding vector(1536)",
		"hnsw",
	}
	for _, kw := range requiredKeywords {
		found := false
		if len(schema) > 0 {
			for i := 0; i <= len(schema)-len(kw); i++ {
				if schema[i:i+len(kw)] == kw {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("Expected schema to contain '%s'", kw)
		}
	}
}

func TestCacheSetAndGet(t *testing.T) {
	se := NewStorageEngine()

	resp := &proto.QueryContextResponse{
		CompiledContext:       "test context",
		TotalTokens:          100,
		CompressionRatio:     0.85,
		OverallRelevanceScore: 0.92,
	}

	se.SetCache("test query", "https://repo.com", "main", resp)

	cached, found := se.GetCache("test query", "https://repo.com", "main")
	if !found {
		t.Fatal("Expected cache hit after SetCache")
	}
	if cached.CompiledContext != "test context" {
		t.Errorf("Expected cached context 'test context', got '%s'", cached.CompiledContext)
	}
	if cached.TotalTokens != 100 {
		t.Errorf("Expected 100 tokens, got %d", cached.TotalTokens)
	}
}

func TestCacheMiss(t *testing.T) {
	se := NewStorageEngine()

	_, found := se.GetCache("nonexistent", "https://repo.com", "main")
	if found {
		t.Error("Expected cache miss for non-existent key")
	}
}

func TestCacheDifferentKeys(t *testing.T) {
	se := NewStorageEngine()

	resp1 := &proto.QueryContextResponse{CompiledContext: "context-1"}
	resp2 := &proto.QueryContextResponse{CompiledContext: "context-2"}

	se.SetCache("query-1", "https://repo.com", "main", resp1)
	se.SetCache("query-2", "https://repo.com", "main", resp2)

	cached1, found := se.GetCache("query-1", "https://repo.com", "main")
	if !found || cached1.CompiledContext != "context-1" {
		t.Error("Expected to get context-1 for query-1")
	}

	cached2, found := se.GetCache("query-2", "https://repo.com", "main")
	if !found || cached2.CompiledContext != "context-2" {
		t.Error("Expected to get context-2 for query-2")
	}
}

func TestCacheExpiration(t *testing.T) {
	se := NewStorageEngine()

	resp := &proto.QueryContextResponse{CompiledContext: "will expire"}
	se.SetCache("expiry-test", "https://repo.com", "main", resp)

	// Manually expire the entry
	key := se.hashKey("expiry-test", "https://repo.com", "main")
	se.mu.Lock()
	se.redisCache[key] = cacheEntry{
		Response:  resp,
		ExpiresAt: time.Now().Add(-1 * time.Second), // Already expired
	}
	se.mu.Unlock()

	_, found := se.GetCache("expiry-test", "https://repo.com", "main")
	if found {
		t.Error("Expected cache miss for expired entry")
	}
}

func TestTelemetryStatsInitial(t *testing.T) {
	se := NewStorageEngine()

	hits, misses, ratio := se.GetTelemetryStats()
	if hits != 0 || misses != 0 || ratio != 0.0 {
		t.Errorf("Expected zero stats initially, got hits=%d misses=%d ratio=%f", hits, misses, ratio)
	}
}

func TestTelemetryStatsAfterOperations(t *testing.T) {
	se := NewStorageEngine()

	resp := &proto.QueryContextResponse{CompiledContext: "cached"}
	se.SetCache("q1", "repo", "main", resp)

	// 1 hit
	se.GetCache("q1", "repo", "main")
	// 2 misses
	se.GetCache("q2", "repo", "main")
	se.GetCache("q3", "repo", "main")

	hits, misses, ratio := se.GetTelemetryStats()
	if hits != 1 {
		t.Errorf("Expected 1 hit, got %d", hits)
	}
	if misses != 2 {
		t.Errorf("Expected 2 misses, got %d", misses)
	}

	expectedRatio := 1.0 / 3.0
	if ratio < expectedRatio-0.01 || ratio > expectedRatio+0.01 {
		t.Errorf("Expected ratio ~0.33, got %f", ratio)
	}
}

func TestHashKeyConsistency(t *testing.T) {
	se := NewStorageEngine()

	key1 := se.hashKey("query", "repo", "branch")
	key2 := se.hashKey("query", "repo", "branch")

	if key1 != key2 {
		t.Error("Expected identical hash keys for identical inputs")
	}

	key3 := se.hashKey("different", "repo", "branch")
	if key1 == key3 {
		t.Error("Expected different hash keys for different inputs")
	}
}
