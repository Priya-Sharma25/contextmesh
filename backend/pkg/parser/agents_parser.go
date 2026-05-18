package parser

import (
	"errors"
	"strings"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/proto"
	"gopkg.in/yaml.v3"
)

// ParseAgentsConfig parses an AGENTS.md content string.
// It can parse either raw YAML or YAML enclosed in a markdown code block (```yaml ... ```).
func ParseAgentsConfig(content string) (*proto.AgentsConfig, error) {
	yamlContent := extractYAML(content)
	if yamlContent == "" {
		return nil, errors.New("empty configuration or invalid structure")
	}

	// Define a shadow struct to parse the YAML structure accurately
	type ScopedPolicyYAML struct {
		Branch           string   `yaml:"branch"`
		Release          string   `yaml:"release"`
		CitationRequired *bool    `yaml:"citation_required"`
		PrioritizePaths  []string `yaml:"prioritize_paths"`
		IgnorePaths      []string `yaml:"ignore_paths"`
	}

	type AgentsConfigYAML struct {
		RepoURL        string             `yaml:"repo_url"`
		Retrieval      struct {
			Prioritize []string `yaml:"prioritize"`
			Ignore     []string `yaml:"ignore"`
		} `yaml:"retrieval"`
		Policies       map[string]bool    `yaml:"policies"`
		ScopedPolicies []ScopedPolicyYAML `yaml:"scoped_policies"`
	}

	var rawConfig AgentsConfigYAML
	err := yaml.Unmarshal([]byte(yamlContent), &rawConfig)
	if err != nil {
		return nil, err
	}

	// Map the raw parsed YAML to the proto-compatible AgentsConfig struct
	config := &proto.AgentsConfig{
		RepoURL:        rawConfig.RepoURL,
		Prioritize:     rawConfig.Retrieval.Prioritize,
		Ignore:         rawConfig.Retrieval.Ignore,
		Policies:       make(map[string]bool),
		ScopedPolicies: make([]proto.ScopedPolicy, 0),
	}

	// Copy policies and supply defaults
	for k, v := range rawConfig.Policies {
		config.Policies[k] = v
	}
	if _, ok := config.Policies["citation_required"]; !ok {
		config.Policies["citation_required"] = true // default to true
	}

	// Copy scoped policies
	for _, sp := range rawConfig.ScopedPolicies {
		citation := true
		if sp.CitationRequired != nil {
			citation = *sp.CitationRequired
		}
		config.ScopedPolicies = append(config.ScopedPolicies, proto.ScopedPolicy{
			Branch:           sp.Branch,
			Release:          sp.Release,
			CitationRequired: citation,
			PrioritizePaths:  sp.PrioritizePaths,
			IgnorePaths:      sp.IgnorePaths,
		})
	}

	return config, nil
}

// ResolvePolicy merges the global config with any matching scoped policies (by branch or release).
// A release match takes highest precedence, followed by a branch match, and finally the global policy.
func ResolvePolicy(config *proto.AgentsConfig, branch string, release string) *proto.ScopedPolicy {
	// Start with base global configuration
	resolved := &proto.ScopedPolicy{
		Branch:           branch,
		Release:          release,
		CitationRequired: true,
		PrioritizePaths:  append([]string{}, config.Prioritize...),
		IgnorePaths:      append([]string{}, config.Ignore...),
	}

	if val, ok := config.Policies["citation_required"]; ok {
		resolved.CitationRequired = val
	}

	// Look for a branch-specific or release-specific match
	var matchedBranchPolicy *proto.ScopedPolicy
	var matchedReleasePolicy *proto.ScopedPolicy

	for i := range config.ScopedPolicies {
		sp := &config.ScopedPolicies[i]
		if sp.Release != "" && sp.Release == release {
			matchedReleasePolicy = sp
			break
		}
		if sp.Branch != "" && sp.Branch == branch {
			matchedBranchPolicy = sp
		}
	}

	// Merge logic: Release policy overrides all if present, else Branch overrides global.
	var activeOverride *proto.ScopedPolicy
	if matchedReleasePolicy != nil {
		activeOverride = matchedReleasePolicy
	} else if matchedBranchPolicy != nil {
		activeOverride = matchedBranchPolicy
	}

	if activeOverride != nil {
		resolved.CitationRequired = activeOverride.CitationRequired
		
		// If prioritizes/ignores are specified, append/merge them
		if len(activeOverride.PrioritizePaths) > 0 {
			resolved.PrioritizePaths = append(resolved.PrioritizePaths, activeOverride.PrioritizePaths...)
		}
		if len(activeOverride.IgnorePaths) > 0 {
			resolved.IgnorePaths = append(resolved.IgnorePaths, activeOverride.IgnorePaths...)
		}
	}

	// Remove duplicate paths for cleaner policies
	resolved.PrioritizePaths = uniqueStrings(resolved.PrioritizePaths)
	resolved.IgnorePaths = uniqueStrings(resolved.IgnorePaths)

	return resolved
}

// Helper to extract YAML content from markdown code fences if present.
func extractYAML(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	// Check if there is a yaml block in markdown
	startTag := "```yaml"
	endTag := "```"
	if strings.Contains(content, startTag) {
		startIdx := strings.Index(content, startTag) + len(startTag)
		remaining := content[startIdx:]
		endIdx := strings.Index(remaining, endTag)
		if endIdx != -1 {
			return strings.TrimSpace(remaining[:endIdx])
		}
	}

	// Also check general code fence block
	startTagGeneric := "```"
	if strings.HasPrefix(content, startTagGeneric) {
		startIdx := len(startTagGeneric)
		remaining := content[startIdx:]
		endIdx := strings.Index(remaining, endTag)
		if endIdx != -1 {
			return strings.TrimSpace(remaining[:endIdx])
		}
	}

	// Otherwise, treat the entire text as raw YAML/Markdown mix
	// If it doesn't contain braces/colons it might be plain text, but yaml.Unmarshal will check that.
	return content
}

func uniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
