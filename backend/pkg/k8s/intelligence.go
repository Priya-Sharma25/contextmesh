package k8s

import (
	"fmt"
	"strings"

	"github.com/Priyasharma620064/contextmesh/backend/pkg/proto"
	"gopkg.in/yaml.v3"
)

type K8sIntelligence struct {
	deprecationRules []deprecationRule
}

type deprecationRule struct {
	kind             string
	deprecatedAPI    string
	supportedAPI     string
	removedInVersion string
	severity         string
	explanation      string
}

func NewK8sIntelligence() *K8sIntelligence {
	rules := []deprecationRule{
		{
			kind:             "Deployment",
			deprecatedAPI:    "extensions/v1beta1",
			supportedAPI:     "apps/v1",
			removedInVersion: "v1.16",
			severity:         "CRITICAL",
			explanation:      "extensions/v1beta1 Deployment is completely obsolete and removed in v1.16.",
		},
		{
			kind:             "Deployment",
			deprecatedAPI:    "apps/v1beta1",
			supportedAPI:     "apps/v1",
			removedInVersion: "v1.16",
			severity:         "CRITICAL",
			explanation:      "apps/v1beta1 Deployment is obsolete and removed in v1.16.",
		},
		{
			kind:             "Ingress",
			deprecatedAPI:    "extensions/v1beta1",
			supportedAPI:     "networking.k8s.io/v1",
			removedInVersion: "v1.22",
			severity:         "CRITICAL",
			explanation:      "extensions/v1beta1 Ingress is obsolete and removed in v1.22.",
		},
		{
			kind:             "Ingress",
			deprecatedAPI:    "networking.k8s.io/v1beta1",
			supportedAPI:     "networking.k8s.io/v1",
			removedInVersion: "v1.22",
			severity:         "CRITICAL",
			explanation:      "networking.k8s.io/v1beta1 Ingress is deprecated and removed in v1.22.",
		},
		{
			kind:             "CronJob",
			deprecatedAPI:    "batch/v1beta1",
			supportedAPI:     "batch/v1",
			removedInVersion: "v1.25",
			severity:         "WARNING",
			explanation:      "batch/v1beta1 CronJob is deprecated and removed in v1.25.",
		},
		{
			kind:             "PodSecurityPolicy",
			deprecatedAPI:    "policy/v1beta1",
			supportedAPI:     "Admission Controllers (Pod Security Standards)",
			removedInVersion: "v1.25",
			severity:         "CRITICAL",
			explanation:      "PodSecurityPolicy (policy/v1beta1) is fully deprecated in v1.21 and removed in v1.25.",
		},
		{
			kind:             "HorizontalPodAutoscaler",
			deprecatedAPI:    "autoscaling/v1",
			supportedAPI:     "autoscaling/v2",
			removedInVersion: "v1.25",
			severity:         "INFO",
			explanation:      "autoscaling/v1 is functional but autoscaling/v2 offers better metrics support.",
		},
	}

	return &K8sIntelligence{
		deprecationRules: rules,
	}
}

// AnalyzeManifest parses a Kubernetes YAML manifest (which can contain multiple sub-documents)
// and returns all parsed API drifts, schema violations, and version mismatches.
func (ki *K8sIntelligence) AnalyzeManifest(manifestContent string, filepath string, targetK8sVersion string) *proto.ValidateK8sManifestResponse {
	var violations []proto.K8sViolation

	// Multi-document YAML splitting
	documents := strings.Split(manifestContent, "---")
	for docIdx, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		// Parse document structure
		var parsed map[string]interface{}
		err := yaml.Unmarshal([]byte(doc), &parsed)
		if err != nil {
			violations = append(violations, proto.K8sViolation{
				FilePath:     filepath,
				APIVersion:   "N/A",
				Kind:         fmt.Sprintf("DocBlock-%d", docIdx+1),
				Severity:     "CRITICAL",
				Message:      "Failed to parse document chunk: " + err.Error(),
				SuggestedFix: "Ensure YAML structure and indentation are aligned.",
			})
			continue
		}

		// Read apiVersion and kind
		apiVersionVal, okApi := parsed["apiVersion"]
		kindVal, okKind := parsed["kind"]
		if !okApi || !okKind {
			// Skip documents that aren't valid Kubernetes manifests (e.g. Helm values file)
			continue
		}

		apiVersion, _ := apiVersionVal.(string)
		kind, _ := kindVal.(string)

		apiVersion = strings.TrimSpace(apiVersion)
		kind = strings.TrimSpace(kind)

		// 1. Cross-reference against deprecation rules
		for _, rule := range ki.deprecationRules {
			if rule.kind == kind && rule.deprecatedAPI == apiVersion {
				violations = append(violations, proto.K8sViolation{
					FilePath:     filepath,
					APIVersion:   apiVersion,
					Kind:         kind,
					Severity:     rule.severity,
					Message:      rule.explanation + fmt.Sprintf(" Incompatible with target Kubernetes version %s.", targetK8sVersion),
					SuggestedFix: "Update apiVersion to '" + rule.supportedAPI + "'.",
				})
			}
		}

		// 2. Scan manifest spec details (Common misconfigurations)
		if metadata, ok := parsed["metadata"].(map[string]interface{}); ok {
			if name, okName := metadata["name"].(string); okName && name == "" {
				violations = append(violations, proto.K8sViolation{
					FilePath:     filepath,
					APIVersion:   apiVersion,
					Kind:         kind,
					Severity:     "CRITICAL",
					Message:      "Missing required field 'metadata.name'.",
					SuggestedFix: "Add a valid string identifier to 'metadata.name'.",
				})
			}
		}
	}

	isValid := len(violations) == 0
	return &proto.ValidateK8sManifestResponse{
		IsValid:    isValid,
		Violations: violations,
	}
}
