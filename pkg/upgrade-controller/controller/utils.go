package controller

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha1"
)

var serviceAccountNamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

const namespaceKubeUpgrade = "kube-upgrade"

// Read the namespace from the inserted serviceaccount file. Fallback to default if the file does not exist.
func GetNamespace() (string, error) {
	data, err := os.ReadFile(serviceAccountNamespaceFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", NewErrorGetNamespace(serviceAccountNamespaceFile, err)
		} else {
			return namespaceKubeUpgrade, nil
		}
	}

	ns := strings.TrimSpace(string(data))
	if len(ns) == 0 {
		return "", NewErrorGetNamespace(serviceAccountNamespaceFile, fmt.Errorf("file was empty"))
	}
	return ns, nil
}

// Return a pointer to the variable value
func Pointer[T any](v T) *T {
	return &v
}

// Check if the given group needs to wait on another one
func groupWaitForDependency(deps []string, status map[string]string) bool {
	for _, d := range deps {
		if status[d] != v1alpha1.PlanStatusComplete {
			return true
		}
	}
	return false
}

// Return the status summary from the given input
func createStatusSummary(status map[string]string) string {
	if len(status) == 0 {
		return v1alpha1.PlanStatusUnknown
	}
	waiting := false
	unknown := false
	progressing := false

	for _, v := range status {
		switch v {
		case v1alpha1.PlanStatusComplete:
		case v1alpha1.PlanStatusProgressing:
			progressing = true
		case v1alpha1.PlanStatusWaiting:
			waiting = true
		default:
			unknown = true
		}
	}

	if unknown {
		return v1alpha1.PlanStatusUnknown
	} else if progressing {
		return v1alpha1.PlanStatusProgressing
	} else if waiting {
		return v1alpha1.PlanStatusWaiting
	} else {
		return v1alpha1.PlanStatusComplete
	}
}
