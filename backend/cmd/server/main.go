package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/agents"
	"github.com/Priyasharma620064/contextmesh/backend/pkg/eval"
	"github.com/Priyasharma620064/contextmesh/backend/pkg/k8s"
	"github.com/Priyasharma620064/contextmesh/backend/pkg/mcp"
	"github.com/Priyasharma620064/contextmesh/backend/pkg/parser"
	"github.com/Priyasharma620064/contextmesh/backend/pkg/proto"
	"github.com/Priyasharma620064/contextmesh/backend/pkg/storage"
)

type Server struct {
	orchestrator *agents.AgentOrchestrator
	storage      *storage.StorageEngine
	mcpServer    *mcp.McpServer
	k8sIntel     *k8s.K8sIntelligence
}

// maxBodySize limits POST request bodies to 1MB to prevent OOM attacks.
const maxBodySize = 1 << 20 // 1 MB

func main() {
	orchestrator := agents.NewAgentOrchestrator()
	srv := &Server{
		orchestrator: orchestrator,
		storage:      storage.NewStorageEngine(),
		mcpServer:    mcp.NewMcpServer(orchestrator),
		k8sIntel:     k8s.NewK8sIntelligence(),
	}

	mux := http.NewServeMux()

	// REST/API endpoints
	mux.HandleFunc("/api/agents/config", srv.handleParseConfig)
	mux.HandleFunc("/api/sync/trigger", srv.handleTriggerSync)
	mux.HandleFunc("/api/sync/status", srv.handleGetSyncStatus)
	mux.HandleFunc("/api/query", srv.handleQueryContext)
	mux.HandleFunc("/api/k8s/validate", srv.handleValidateK8s)
	mux.HandleFunc("/api/eval/benchmark", srv.handleRunBenchmark)
	mux.HandleFunc("/api/storage/schema", srv.handleGetStorageSchema)
	mux.HandleFunc("/api/storage/stats", srv.handleGetStorageStats)
	
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	
	// MCP Endpoint
	mux.HandleFunc("/mcp", srv.handleMcpRequest)

	// Determine allowed CORS origins from environment (defaults to localhost:3000)
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:3000"
	}
	allowedOriginsSet := make(map[string]bool)
	for _, origin := range strings.Split(allowedOrigins, ",") {
		allowedOriginsSet[strings.TrimSpace(origin)] = true
	}

	// Wrap mux with configurable CORS middleware
	corsMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestOrigin := r.Header.Get("Origin")
		if allowedOriginsSet[requestOrigin] {
			w.Header().Set("Access-Control-Allow-Origin", requestOrigin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Vary", "Origin")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		mux.ServeHTTP(w, r)
	})

	port := ":8080"
	slog.Info("----------------------------------------------------------------------")
	slog.Info("🚀 ContextMesh Agentic Context Engineering Platform booting...")
	slog.Info("🌐 Go Microservices engine running on port " + port)
	slog.Info("🛠️  MCP Tool Server enabled at http://localhost" + port + "/mcp")
	slog.Info("🤖 In-Memory storage initialized and preloaded with openkruise/kruise mock indices")
	slog.Info("----------------------------------------------------------------------")

	if err := http.ListenAndServe(port, corsMux); err != nil {
		slog.Error("Server shutdown failed", "error", err)
		os.Exit(1)
	}
}

func (s *Server) handleParseConfig(w http.ResponseWriter, r *http.Request) {
	var req proto.ParseAgentsConfigRequest
	if !s.parseRequest(w, r, &req) {
		return
	}

	config, err := parser.ParseAgentsConfig(req.Content)
	if err != nil {
		s.jsonSuccess(w, proto.ParseAgentsConfigResponse{
			Valid:        false,
			ErrorMessage: err.Error(),
		})
		return
	}

	s.jsonSuccess(w, proto.ParseAgentsConfigResponse{
		Valid:        true,
		ParsedConfig: *config,
	})
}

func (s *Server) handleTriggerSync(w http.ResponseWriter, r *http.Request) {
	var req proto.TriggerSyncRequest
	if !s.parseRequest(w, r, &req) {
		return
	}

	state, err := s.orchestrator.SyncAgent.TriggerSync(req.RepoURL, req.Branch, req.Release)
	if err != nil {
		s.jsonError(w, "Sync failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonSuccess(w, proto.TriggerSyncResponse{
		SyncID:  state.LastCommitHash[:8],
		Message: fmt.Sprintf("Incremental sync completed. Total chunks indexed: %d", state.TotalChunksIndexed),
	})
}

func (s *Server) handleGetSyncStatus(w http.ResponseWriter, r *http.Request) {
	repoURL := r.URL.Query().Get("repo")
	if repoURL == "" {
		repoURL = getDefaultRepo()
	}

	states := s.orchestrator.SyncAgent.GetSyncStatus(repoURL)
	s.jsonSuccess(w, proto.GetSyncStatusResponse{
		SyncStates: states,
	})
}

func (s *Server) handleQueryContext(w http.ResponseWriter, r *http.Request) {
	var req proto.QueryContextRequest
	if !s.parseRequest(w, r, &req) {
		return
	}

	if req.MaxTokens <= 0 {
		req.MaxTokens = 4000
	}

	// 1. Caching Layer Check
	if !req.BypassCache {
		if cachedResp, found := s.storage.GetCache(req.Query, req.RepoURL, req.Branch); found {
			s.jsonSuccess(w, cachedResp)
			return
		}
	}

	// 2. Perform Context Orchestration
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.orchestrator.ExecuteOrchestration(ctx, &req)
	if err != nil {
		s.jsonError(w, "Orchestration query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Cache the output
	s.storage.SetCache(req.Query, req.RepoURL, req.Branch, resp)

	s.jsonSuccess(w, resp)
}

func (s *Server) handleValidateK8s(w http.ResponseWriter, r *http.Request) {
	var req proto.ValidateK8sManifestRequest
	if !s.parseRequest(w, r, &req) {
		return
	}

	if req.TargetK8sVersion == "" {
		req.TargetK8sVersion = "v1.25"
	}

	res := s.k8sIntel.AnalyzeManifest(req.ManifestContent, req.FilePath, req.TargetK8sVersion)
	s.jsonSuccess(w, res)
}

func (s *Server) handleRunBenchmark(w http.ResponseWriter, r *http.Request) {
	repoURL := r.URL.Query().Get("repo")
	branch := r.URL.Query().Get("branch")
	if repoURL == "" {
		repoURL = getDefaultRepo()
	}
	if branch == "" {
		branch = "main"
	}

	// Fetch active index chunks to run evaluations
	activeChunks := s.orchestrator.SyncAgent.GetIndexedChunks(repoURL, branch)

	benchmark := eval.NewBenchmarkSuite()
	res := benchmark.RunSuite(repoURL, branch, activeChunks)
	s.jsonSuccess(w, res)
}

func (s *Server) handleGetStorageSchema(w http.ResponseWriter, r *http.Request) {
	schema := s.storage.GetPgvectorSchema()
	s.jsonSuccess(w, map[string]string{
		"schema": schema,
	})
}

func (s *Server) handleGetStorageStats(w http.ResponseWriter, r *http.Request) {
	hits, misses, ratio := s.storage.GetTelemetryStats()
	s.jsonSuccess(w, map[string]interface{}{
		"cache_hits":   hits,
		"cache_misses": misses,
		"hit_ratio":    ratio,
	})
}

func (s *Server) handleMcpRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := s.readBody(r)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	respBytes, err := s.mcpServer.HandleRequest(ctx, body)
	if err != nil {
		s.jsonError(w, "MCP handle failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(respBytes)
}

func (s *Server) jsonSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to encode JSON response", "error", err)
	}
}

func (s *Server) jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// readBody reads the request body with a size limit to prevent OOM attacks.
func (s *Server) readBody(r *http.Request) ([]byte, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodySize)
	return io.ReadAll(r.Body)
}

func getDefaultRepo() string {
	repo := os.Getenv("DEFAULT_REPO_URL")
	if repo == "" {
		return "https://github.com/Priyasharma620064/contextmesh"
	}
	return repo
}

// parseRequest is a helper that ensures method is POST, limits body size, and decodes JSON.
func (s *Server) parseRequest(w http.ResponseWriter, r *http.Request, req interface{}) bool {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	body, err := s.readBody(r)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusBadRequest)
		return false
	}
	if err := json.Unmarshal(body, req); err != nil {
		s.jsonError(w, "Failed to parse JSON body: "+err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}
