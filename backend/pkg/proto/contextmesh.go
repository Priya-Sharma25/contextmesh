package proto

import "time"

// ScopedPolicy represents sub-branch or release-specific retrieval rules.
type ScopedPolicy struct {
	Branch           string   `json:"branch"`
	Release          string   `json:"release"`
	CitationRequired bool     `json:"citation_required"`
	PrioritizePaths  []string `json:"prioritize_paths"`
	IgnorePaths      []string `json:"ignore_paths"`
}

// AgentsConfig matches the YAML-parsed config definitions in AGENTS.md.
type AgentsConfig struct {
	RepoURL        string          `json:"repo_url"`
	Prioritize     []string        `json:"prioritize"`
	Ignore         []string        `json:"ignore"`
	Policies       map[string]bool `json:"policies"`
	ScopedPolicies []ScopedPolicy  `json:"scoped_policies"`
}

// ParseAgentsConfigRequest for reading/validating AGENTS.md.
type ParseAgentsConfigRequest struct {
	Content string `json:"content"`
}

// ParseAgentsConfigResponse for validation outputs.
type ParseAgentsConfigResponse struct {
	Valid        bool         `json:"valid"`
	ErrorMessage string       `json:"error_message"`
	ParsedConfig AgentsConfig `json:"parsed_config"`
}

// TriggerSyncRequest initiates multi-repo delta indexing.
type TriggerSyncRequest struct {
	RepoURL string `json:"repo_url"`
	Branch  string `json:"branch"`
	Release string `json:"release"`
}

// TriggerSyncResponse contains tracking metrics.
type TriggerSyncResponse struct {
	SyncID  string `json:"sync_id"`
	Message string `json:"message"`
}

// RepositorySyncState represents status of active repos.
type RepositorySyncState struct {
	RepoURL            string    `json:"repo_url"`
	Branch             string    `json:"branch"`
	LastCommitHash     string    `json:"last_commit_hash"`
	TotalChunksIndexed int64     `json:"total_chunks_indexed"`
	FilesChanged       int64     `json:"files_changed"`
	FilesUnchanged     int64     `json:"files_unchanged"`
	SyncStatus         string    `json:"sync_status"` // "SYNCING", "COMPLETED", "FAILED"
	UpdatedAt          time.Time `json:"updated_at"`
}

// GetSyncStatusRequest queries active repo indexing states.
type GetSyncStatusRequest struct {
	RepoURL string `json:"repo_url"`
}

// GetSyncStatusResponse lists active indexing.
type GetSyncStatusResponse struct {
	SyncStates []RepositorySyncState `json:"sync_states"`
}

// QueryContextRequest searches the engine.
type QueryContextRequest struct {
	Query       string `json:"query"`
	RepoURL     string `json:"repo_url"`
	Branch      string `json:"branch"`
	MaxTokens   int32  `json:"max_tokens"`
	BypassCache bool   `json:"bypass_cache"`
}

// TextChunk holds segmented markdown information.
type TextChunk struct {
	ID        string            `json:"id"`
	Text      string            `json:"text"`
	Filepath  string            `json:"filepath"`
	StartLine int32             `json:"start_line"`
	EndLine   int32             `json:"end_line"`
	Score     float64           `json:"score"`
	Metadata  map[string]string `json:"metadata"`
}

// QueryContextResponse returns formatted results.
type QueryContextResponse struct {
	CompiledContext       string      `json:"compiled_context"`
	RetrievedChunks       []TextChunk `json:"retrieved_chunks"`
	TotalTokens           int32       `json:"total_tokens"`
	CompressionRatio      float64     `json:"compression_ratio"`
	OverallRelevanceScore float64     `json:"overall_relevance_score"`
}

// RunBenchmarkRequest triggers retrieval evaluations.
type RunBenchmarkRequest struct {
	SuiteName string `json:"suite_name"`
	RepoURL   string `json:"repo_url"`
	Branch    string `json:"branch"`
}

// BenchmarkMetrics rates retrieval precision and hallucination indicators.
type BenchmarkMetrics struct {
	PrecisionAtK       float64 `json:"precision_at_k"`
	RecallAtK          float64 `json:"recall_at_k"`
	HallucinationRisk  float64 `json:"hallucination_risk"`
	ChunkQuality       float64 `json:"chunk_quality"`
	CitationConfidence float64 `json:"citation_confidence"`
	LatencyMs          float64 `json:"latency_ms"`
}

// SideBySideComparison tests two indexing/chunking strategies.
type SideBySideComparison struct {
	StrategyAName string           `json:"strategy_a_name"`
	MetricsA      BenchmarkMetrics `json:"metrics_a"`
	ChunksA       []TextChunk      `json:"chunks_a"`

	StrategyBName string           `json:"strategy_b_name"`
	MetricsB      BenchmarkMetrics `json:"metrics_b"`
	ChunksB       []TextChunk      `json:"chunks_b"`
}

// RunBenchmarkResponse displays test summaries.
type RunBenchmarkResponse struct {
	RunID          string                 `json:"run_id"`
	SuiteName      string                 `json:"suite_name"`
	AverageMetrics BenchmarkMetrics       `json:"average_metrics"`
	Comparisons    []SideBySideComparison `json:"comparisons"`
	CompletedAt    time.Time              `json:"completed_at"`
}

// K8sViolation models problematic manifests.
type K8sViolation struct {
	FilePath     string `json:"file_path"`
	APIVersion   string `json:"api_version"`
	Kind         string `json:"kind"`
	Severity     string `json:"severity"` // "CRITICAL", "WARNING", "INFO"
	Message      string `json:"message"`
	SuggestedFix string `json:"suggested_fix"`
}

// ValidateK8sManifestRequest scans raw manifests.
type ValidateK8sManifestRequest struct {
	ManifestContent  string `json:"manifest_content"`
	FilePath         string `json:"file_path"`
	TargetK8sVersion string `json:"target_k8s_version"`
}

// ValidateK8sManifestResponse reports deprecations and linting rules.
type ValidateK8sManifestResponse struct {
	IsValid    bool           `json:"is_valid"`
	Violations []K8sViolation `json:"violations"`
}
