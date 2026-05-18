package k8s

import (
	"strings"
	"testing"
)

func TestK8sAPIAnalyzer(t *testing.T) {
	ki := NewK8sIntelligence()

	multiDocManifest := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: healthy-dashboard
spec:
  replicas: 2
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: outdated-scheduler
spec:
  replicas: 1
---
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: old-ingress
spec:
  rules:
  - host: mesh.io
---
# Malformed YAML block
apiVersion: batch/v1
kind: Job
metadata:
  name: malformed-job
spec:
  template:
   spec:
    containers:
    - name: runner
      image: alpine:latest
     indentation_error: true
`

	res := ki.AnalyzeManifest(multiDocManifest, "deploy/manifest.yaml", "v1.25")

	if res.IsValid {
		t.Fatal("Expected manifest check to fail due to multiple apiVersion deprecations and YAML syntax errors")
	}

	foundDeploymentDep := false
	foundIngressDep := false
	foundSyntaxError := false

	for _, v := range res.Violations {
		if v.APIVersion == "extensions/v1beta1" && v.Kind == "Deployment" {
			foundDeploymentDep = true
			if v.Severity != "CRITICAL" {
				t.Errorf("Expected Deployment extensions/v1beta1 to have CRITICAL severity, got: %s", v.Severity)
			}
			if !strings.Contains(v.SuggestedFix, "apps/v1") {
				t.Errorf("Expected suggestion to update to apps/v1, got: %s", v.SuggestedFix)
			}
		}

		if v.APIVersion == "networking.k8s.io/v1beta1" && v.Kind == "Ingress" {
			foundIngressDep = true
			if v.Severity != "CRITICAL" {
				t.Errorf("Expected Ingress networking.k8s.io/v1beta1 to have CRITICAL severity, got: %s", v.Severity)
			}
			if !strings.Contains(v.SuggestedFix, "networking.k8s.io/v1") {
				t.Errorf("Expected suggestion to update to networking.k8s.io/v1, got: %s", v.SuggestedFix)
			}
		}

		if v.Severity == "CRITICAL" && strings.Contains(v.Message, "parse") {
			foundSyntaxError = true
		}
	}

	if !foundDeploymentDep {
		t.Error("Expected to flag deprecated Deployment (extensions/v1beta1)")
	}

	if !foundIngressDep {
		t.Error("Expected to flag deprecated Ingress (networking.k8s.io/v1beta1)")
	}

	if !foundSyntaxError {
		t.Error("Expected to catch YAML indentation syntax error in the Job block")
	}
}
