package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/agents"
	"github.com/Priyasharma620064/contextmesh/backend/pkg/eval"
	"github.com/Priyasharma620064/contextmesh/backend/pkg/k8s"
	"github.com/Priyasharma620064/contextmesh/backend/pkg/proto"
)

type McpServer struct {
	orchestrator *agents.AgentOrchestrator
	k8sIntel     *k8s.K8sIntelligence
}

type JsonRpcRequest struct {
	JsonRpc string          `json:"jsonrpc"`
	Id      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JsonRpcResponse struct {
	JsonRpc string      `json:"jsonrpc"`
	Id      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RpcError   `json:"error,omitempty"`
}

type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Tool represents an MCP-exposed tool structure
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string                `json:"type"`
	Properties map[string]Property   `json:"properties"`
	Required   []string              `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

func NewMcpServer(orchestrator *agents.AgentOrchestrator) *McpServer {
	return &McpServer{
		orchestrator: orchestrator,
		k8sIntel:     k8s.NewK8sIntelligence(),
	}
}

// ListTools returns the registered MCP tools.
func (s *McpServer) ListTools() []Tool {
	return []Tool{
		{
			Name:        "semantic_search",
			Description: "Search repository documentation utilizing Agentic Context Engineering rules, priority boosts, and compression.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"query": {Type: "string", Description: "The semantic search query."},
					"repo":  {Type: "string", Description: "Target repository URL."},
					"branch": {Type: "string", Description: "Target branch name."},
				},
				Required: []string{"query", "repo"},
			},
		},
		{
			Name:        "validate_k8s_docs",
			Description: "Scan Kubernetes manifest documentation chunks for deprecated API drifts and linting issues.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"manifest": {Type: "string", Description: "Raw yaml content of the manifest file."},
					"filepath": {Type: "string", Description: "Path designation for the file."},
				},
				Required: []string{"manifest"},
			},
		},
		{
			Name:        "run_benchmarks",
			Description: "Trigger the Side-by-Side Retrieval Evaluation Benchmark Suite using gold-standard datasets.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"repo":   {Type: "string", Description: "Repository target URL."},
					"branch": {Type: "string", Description: "Repository branch target."},
				},
				Required: []string{"repo"},
			},
		},
	}
}

// HandleRequest routes incoming JSON-RPC 2.0 payloads for MCP.
func (s *McpServer) HandleRequest(ctx context.Context, body []byte) ([]byte, error) {
	var req JsonRpcRequest
	err := json.Unmarshal(body, &req)
	if err != nil {
		return s.errorResponse(nil, -32700, "Parse error: "+err.Error())
	}

	if req.JsonRpc != "2.0" {
		return s.errorResponse(req.Id, -32600, "Invalid Request: missing jsonrpc version")
	}

	switch req.Method {
	case "tools/list":
		return s.successResponse(req.Id, map[string]interface{}{
			"tools": s.ListTools(),
		})

	case "tools/call":
		var callParams struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		err := json.Unmarshal(req.Params, &callParams)
		if err != nil {
			return s.errorResponse(req.Id, -32602, "Invalid params: "+err.Error())
		}

		result, err := s.CallTool(ctx, callParams.Name, callParams.Arguments)
		if err != nil {
			return s.errorResponse(req.Id, -32603, "Internal tool error: "+err.Error())
		}

		return s.successResponse(req.Id, result)

	default:
		return s.errorResponse(req.Id, -32601, "Method not found: "+req.Method)
	}
}

func (s *McpServer) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	switch name {
	case "semantic_search":
		query, _ := args["query"].(string)
		repo, _ := args["repo"].(string)
		branch, ok := args["branch"].(string)
		if !ok {
			branch = "main"
		}

		if query == "" || repo == "" {
			return nil, errors.New("missing query or repo arguments")
		}

		res, err := s.orchestrator.ExecuteOrchestration(ctx, &proto.QueryContextRequest{
			Query:     query,
			RepoURL:   repo,
			Branch:    branch,
			MaxTokens: 4000,
		})
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"content": []map[string]string{
				{
					"type": "text",
					"text": res.CompiledContext,
				},
			},
			"metrics": map[string]interface{}{
				"relevance_score":   res.OverallRelevanceScore,
				"total_tokens":      res.TotalTokens,
				"compression_ratio": res.CompressionRatio,
			},
		}, nil

	case "validate_k8s_docs":
		manifest, _ := args["manifest"].(string)
		filepath, ok := args["filepath"].(string)
		if !ok {
			filepath = "manifest.yaml"
		}

		if manifest == "" {
			return nil, errors.New("missing manifest argument")
		}

		res := s.k8sIntel.AnalyzeManifest(manifest, filepath, "v1.25")
		return map[string]interface{}{
			"is_valid":   res.IsValid,
			"violations": res.Violations,
		}, nil

	case "run_benchmarks":
		repo, _ := args["repo"].(string)
		branch, ok := args["branch"].(string)
		if !ok {
			branch = "main"
		}

		if repo == "" {
			return nil, errors.New("missing repo argument")
		}

		// Pull active index chunks for context evaluation runs
		activeChunks := s.orchestrator.SyncAgent.GetIndexedChunks(repo, branch)
		
		benchmark := eval.NewBenchmarkSuite()
		res := benchmark.RunSuite(repo, branch, activeChunks)

		return map[string]interface{}{
			"run_id":          res.RunID,
			"suite":           res.SuiteName,
			"average_metrics": res.AverageMetrics,
			"comparisons":     res.Comparisons,
		}, nil

	default:
		return nil, fmt.Errorf("tool '%s' not supported", name)
	}
}

func (s *McpServer) successResponse(id interface{}, result interface{}) ([]byte, error) {
	resp := JsonRpcResponse{
		JsonRpc: "2.0",
		Id:      id,
		Result:  result,
	}
	return json.Marshal(resp)
}

func (s *McpServer) errorResponse(id interface{}, code int, message string) ([]byte, error) {
	resp := JsonRpcResponse{
		JsonRpc: "2.0",
		Id:      id,
		Error: &RpcError{
			Code:    code,
			Message: message,
		},
	}
	return json.Marshal(resp)
}
