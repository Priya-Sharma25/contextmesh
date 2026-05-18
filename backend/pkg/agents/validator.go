package agents

import (
	"context"
	"regexp"
	"strings"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/proto"
	"gopkg.in/yaml.v3"
)

type ValidationAgent struct {
	linkRegex *regexp.Regexp
}

func NewValidationAgent() *ValidationAgent {
	// Regular expression to match markdown links: [label](url)
	return &ValidationAgent{
		linkRegex: regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`),
	}
}

// ValidateChunk analyzes an individual chunk for broken links, malformed YAMLs, and API drifts.
func (va *ValidationAgent) ValidateChunk(ctx context.Context, chunk proto.TextChunk) []proto.K8sViolation {
	var violations []proto.K8sViolation

	// 1. Broken Links Scanner
	links := va.linkRegex.FindAllStringSubmatch(chunk.Text, -1)
	for _, match := range links {
		if len(match) == 3 {
			url := match[2]
			if va.isBrokenLink(url) {
				violations = append(violations, proto.K8sViolation{
					FilePath:     chunk.Filepath,
					APIVersion:   "N/A",
					Kind:         "MarkdownLink",
					Severity:     "WARNING",
					Message:      "Broken or deprecated hyperlink target: " + url,
					SuggestedFix: "Verify path mapping or update relative link destination.",
				})
			}
		}
	}

	// 2. Embedded YAML Parser
	yamlBlocks := va.extractYAMLBlocks(chunk.Text)
	for _, block := range yamlBlocks {
		var mapped interface{}
		err := yaml.Unmarshal([]byte(block), &mapped)
		if err != nil {
			violations = append(violations, proto.K8sViolation{
				FilePath:     chunk.Filepath,
				APIVersion:   "Unknown",
				Kind:         "YAMLConfig",
				Severity:     "CRITICAL",
				Message:      "Embedded YAML block syntax error: " + err.Error(),
				SuggestedFix: "Fix YAML indentation or balance string quotes.",
			})
			continue
		}

		// 3. Kubernetes API Deprecation Scanner
		if strings.Contains(block, "apiVersion:") && strings.Contains(block, "kind:") {
			violations = append(violations, va.checkK8sDeprecations(block, chunk.Filepath)...)
		}
	}

	return violations
}

func (va *ValidationAgent) isBrokenLink(url string) bool {
	// Relative links checking (pointing to local files)
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "#") {
		// If link has trailing anchors, strip them
		cleanPath := strings.Split(url, "#")[0]
		if cleanPath == "" {
			return false
		}
		// In a real environment, we would verify local file existence.
		// For our mock/simulated environment: we simulate broken links for paths containing "deprecated" or "broken".
		if strings.Contains(cleanPath, "deprecated/") || strings.Contains(cleanPath, "broken-link") || strings.HasSuffix(cleanPath, ".tmp") {
			return true
		}
	}
	// Absolute links checking
	if strings.Contains(url, "localhost:") || strings.Contains(url, "example.com/broken") {
		return true
	}
	return false
}

func (va *ValidationAgent) extractYAMLBlocks(text string) []string {
	var blocks []string
	lines := strings.Split(text, "\n")
	inBlock := false
	var currentBlock []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```yaml") || (strings.HasPrefix(trimmed, "```") && inBlock) {
			if inBlock {
				blocks = append(blocks, strings.Join(currentBlock, "\n"))
				currentBlock = nil
				inBlock = false
			} else {
				inBlock = true
			}
			continue
		}
		if inBlock {
			currentBlock = append(currentBlock, line)
		}
	}
	return blocks
}

func (va *ValidationAgent) checkK8sDeprecations(yamlBlock string, filepath string) []proto.K8sViolation {
	var violations []proto.K8sViolation

	lines := strings.Split(yamlBlock, "\n")
	var apiVersion, kind string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "apiVersion:") {
			apiVersion = strings.TrimSpace(strings.TrimPrefix(l, "apiVersion:"))
			apiVersion = strings.Trim(apiVersion, "\"' ")
		}
		if strings.HasPrefix(l, "kind:") {
			kind = strings.TrimSpace(strings.TrimPrefix(l, "kind:"))
			kind = strings.Trim(kind, "\"' ")
		}
	}

	// Flag Deprecations (e.g. apps/v1beta1, extensions/v1beta1)
	if apiVersion != "" && kind != "" {
		if apiVersion == "extensions/v1beta1" && (kind == "Deployment" || kind == "ReplicaSet") {
			violations = append(violations, proto.K8sViolation{
				FilePath:     filepath,
				APIVersion:   apiVersion,
				Kind:         kind,
				Severity:     "CRITICAL",
				Message:      "Deprecated apiVersion 'extensions/v1beta1' is not supported in modern Kubernetes (v1.16+).",
				SuggestedFix: "Change apiVersion to 'apps/v1'.",
			})
		}
		if apiVersion == "networking.k8s.io/v1beta1" && kind == "Ingress" {
			violations = append(violations, proto.K8sViolation{
				FilePath:     filepath,
				APIVersion:   apiVersion,
				Kind:         kind,
				Severity:     "WARNING",
				Message:      "Deprecated apiVersion 'networking.k8s.io/v1beta1' is deprecated in v1.19+ and removed in v1.22+.",
				SuggestedFix: "Change apiVersion to 'networking.k8s.io/v1'.",
			})
		}
	}

	return violations
}
