package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
	
	// MCP Endpoint
	mux.HandleFunc("/mcp", srv.handleMcpRequest)

	// Wrap mux with standard CORS middleware
	corsMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		mux.ServeHTTP(w, r)
	})

	port := ":8080"
	log.Printf("----------------------------------------------------------------------")
	log.Printf("🚀 ContextMesh Agentic Context Engineering Platform booting...")
	log.Printf("🌐 Go Microservices engine running on port %s", port)
	log.Printf("🛠️  MCP Tool Server enabled at http://localhost%s/mcp", port)
	log.Printf("🤖 In-Memory storage initialized and preloaded with openkruise/kruise mock indices")
	log.Printf("----------------------------------------------------------------------")

	if err := http.ListenAndServe(port, corsMux); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
}

func (s *Server) handleParseConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req proto.ParseAgentsConfigRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.jsonError(w, "Failed to parse JSON body: "+err.Error(), http.StatusBadRequest)
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req proto.TriggerSyncRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.jsonError(w, "Failed to parse JSON body: "+err.Error(), http.StatusBadRequest)
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
		repoURL = "https://github.com/Priyasharma620064/contextmesh" // default
	}

	states := s.orchestrator.SyncAgent.GetSyncStatus(repoURL)
	s.jsonSuccess(w, proto.GetSyncStatusResponse{
		SyncStates: states,
	})
}

func (s *Server) handleQueryContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req proto.QueryContextRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.jsonError(w, "Failed to parse JSON body: "+err.Error(), http.StatusBadRequest)
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req proto.ValidateK8sManifestRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.jsonError(w, "Failed to parse JSON body: "+err.Error(), http.StatusBadRequest)
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
		repoURL = "https://github.com/Priyasharma620064/contextmesh"
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
	body, err := io.ReadAll(r.Body)
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
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

func (s *Server) jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
