# AGENTS.md Configuration

This file configures the scoped documentation retrieval, synchronization, and indexing rules for the ContextMesh Agentic Context Engineering Platform.

```yaml
repo_url: "https://github.com/Priyasharma620064/contextmesh"

# Default global retrieval guidelines
retrieval:
  prioritize:
    - docs/
    - tutorials/
    - examples/
  ignore:
    - deprecated/
    - temp/
    - "*.tmp"

# Base platform policies
policies:
  citation_required: true
  enable_semantic_ranking: true
  enforce_strict_crd_schemas: true

# Fine-grained, branch-aware & release-aware target policies
scoped_policies:
  - branch: main
    prioritize_paths:
      - docs/core/
      - docs/production/
    citation_required: true

  - branch: develop
    prioritize_paths:
      - docs/experimental/
      - docs/draft/
    ignore_paths:
      - deprecated/alpha/
    citation_required: false

  - release: v1.0.0
    prioritize_paths:
      - docs/stable-v1.0/
    ignore_paths:
      - docs/experimental/
      - temp/
    citation_required: true
```
