package parser

import (
	"testing"
)

func TestParseAgentsConfig(t *testing.T) {
	markdownContent := `
# AGENTS.md Configuration

This defines the default context rules.

` + "```yaml" + `
repo_url: "https://github.com/openkruise/kruise"
retrieval:
  prioritize:
    - docs/
    - tutorials/
  ignore:
    - deprecated/

policies:
  citation_required: false

scoped_policies:
  - branch: main
    prioritize_paths:
      - docs/core/
  - release: v1.4.0
    prioritize_paths:
      - docs/v1.4-spec/
    ignore_paths:
      - deprecated/v1.4/
    citation_required: true
` + "```" + `
`

	config, err := ParseAgentsConfig(markdownContent)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config.RepoURL != "https://github.com/openkruise/kruise" {
		t.Errorf("Expected RepoURL to be 'https://github.com/openkruise/kruise', got %s", config.RepoURL)
	}

	if len(config.Prioritize) != 2 || config.Prioritize[0] != "docs/" {
		t.Errorf("Expected 2 prioritize paths, got %v", config.Prioritize)
	}

	if val, ok := config.Policies["citation_required"]; !ok || val {
		t.Errorf("Expected citation_required to be parsed as false, got %t", val)
	}

	if len(config.ScopedPolicies) != 2 {
		t.Errorf("Expected 2 scoped policies, got %d", len(config.ScopedPolicies))
	}
}

func TestResolvePolicy(t *testing.T) {
	markdownContent := `
repo_url: "https://github.com/openkruise/kruise"
retrieval:
  prioritize:
    - docs/
  ignore:
    - deprecated/
policies:
  citation_required: true
scoped_policies:
  - branch: develop
    prioritize_paths:
      - docs/draft/
    citation_required: false
  - release: v1.0
    prioritize_paths:
      - docs/release-1.0/
    ignore_paths:
      - deprecated/old/
`
	config, err := ParseAgentsConfig(markdownContent)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// 1. Resolve Global Default (no matching branch/release)
	global := ResolvePolicy(config, "main", "")
	if !global.CitationRequired {
		t.Error("Expected citation_required to be true globally")
	}
	if len(global.PrioritizePaths) != 1 || global.PrioritizePaths[0] != "docs/" {
		t.Errorf("Expected prioritize docs/, got %v", global.PrioritizePaths)
	}

	// 2. Resolve Branch Override ("develop")
	devBranch := ResolvePolicy(config, "develop", "")
	if devBranch.CitationRequired {
		t.Error("Expected citation_required to be overridden to false on branch 'develop'")
	}
	if len(devBranch.PrioritizePaths) != 2 || devBranch.PrioritizePaths[1] != "docs/draft/" {
		t.Errorf("Expected merged prioritizes, got %v", devBranch.PrioritizePaths)
	}

	// 3. Resolve Release Override ("v1.0")
	rel1 := ResolvePolicy(config, "develop", "v1.0") // both are specified
	if !rel1.CitationRequired {
		t.Error("Expected citation_required to stay true from release config")
	}
	if len(rel1.IgnorePaths) != 2 || rel1.IgnorePaths[1] != "deprecated/old/" {
		t.Errorf("Expected merged ignore paths, got %v", rel1.IgnorePaths)
	}
}
