package controller

import (
	"testing"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/stretchr/testify/assert"
)

func TestCombineConfig(t *testing.T) {
	tMatrix := []struct {
		Name   string
		Global api.UpgradedConfig
		Group  *api.UpgradedConfig
		Result *api.UpgradedConfig
	}{
		{
			Name: "OverrideAll",
			Global: api.UpgradedConfig{
				Stream:         "registry.example.org/test-stream",
				FleetlockURL:   "https://fleetlock.example.org",
				FleetlockGroup: "not-default",
				CheckInterval:  "10m",
				RetryInterval:  "15m",
				LogLevel:       "error",
				KubeletConfig:  "/foo/kubelet.conf",
				KubeadmPath:    "/foo/kubeadm",
			},
			Group: &api.UpgradedConfig{
				Stream:         "registry.example.com/test-stream",
				FleetlockURL:   "https://fleetlock.example.com",
				FleetlockGroup: "default",
				CheckInterval:  "2m",
				RetryInterval:  "3m",
				LogLevel:       "debug",
				KubeletConfig:  "/foo/bar/kubelet.conf",
				KubeadmPath:    "/foo/bar/kubeadm",
			},
			Result: &api.UpgradedConfig{
				Stream:         "registry.example.com/test-stream",
				FleetlockURL:   "https://fleetlock.example.com",
				FleetlockGroup: "default",
				CheckInterval:  "2m",
				RetryInterval:  "3m",
				LogLevel:       "debug",
				KubeletConfig:  "/foo/bar/kubelet.conf",
				KubeadmPath:    "/foo/bar/kubeadm",
			},
		},
		{
			Name: "PartialOverride#1",
			Global: api.UpgradedConfig{
				Stream:         "registry.example.org/test-stream",
				FleetlockURL:   "https://fleetlock.example.org",
				FleetlockGroup: "not-default",
				CheckInterval:  "10m",
				RetryInterval:  "15m",
				LogLevel:       "error",
				KubeletConfig:  "/foo/kubelet.conf",
				KubeadmPath:    "/foo/kubeadm",
			},
			Group: &api.UpgradedConfig{
				FleetlockGroup: "default",
				CheckInterval:  "2m",
				RetryInterval:  "3m",
				LogLevel:       "info",
			},
			Result: &api.UpgradedConfig{
				Stream:         "registry.example.org/test-stream",
				FleetlockURL:   "https://fleetlock.example.org",
				FleetlockGroup: "default",
				CheckInterval:  "2m",
				RetryInterval:  "3m",
				LogLevel:       "info",
				KubeletConfig:  "/foo/kubelet.conf",
				KubeadmPath:    "/foo/kubeadm",
			},
		},
		{
			Name: "PartialOverride#2",
			Global: api.UpgradedConfig{
				Stream:         "registry.example.org/test-stream",
				FleetlockURL:   "https://fleetlock.example.org",
				FleetlockGroup: "not-default",
				CheckInterval:  "10m",
				RetryInterval:  "15m",
				LogLevel:       "error",
				KubeletConfig:  "/foo/kubelet.conf",
				KubeadmPath:    "/foo/kubeadm",
			},
			Group: &api.UpgradedConfig{
				Stream:        "registry.example.com/test-stream",
				FleetlockURL:  "https://fleetlock.example.com",
				LogLevel:      "warn",
				KubeletConfig: "/foo/bar/kubelet.conf",
				KubeadmPath:   "/foo/bar/kubeadm",
			},
			Result: &api.UpgradedConfig{
				Stream:         "registry.example.com/test-stream",
				FleetlockURL:   "https://fleetlock.example.com",
				FleetlockGroup: "not-default",
				CheckInterval:  "10m",
				RetryInterval:  "15m",
				LogLevel:       "warn",
				KubeletConfig:  "/foo/bar/kubelet.conf",
				KubeadmPath:    "/foo/bar/kubeadm",
			},
		},
		{
			Name: "GlobalEmpty",
			Group: &api.UpgradedConfig{
				Stream:         "registry.example.com/test-stream",
				FleetlockURL:   "https://fleetlock.example.com",
				FleetlockGroup: "default",
				CheckInterval:  "2m",
				RetryInterval:  "3m",
				LogLevel:       "error",
				KubeletConfig:  "/foo/kubelet.conf",
				KubeadmPath:    "/foo/kubeadm",
			},
			Result: &api.UpgradedConfig{
				Stream:         "registry.example.com/test-stream",
				FleetlockURL:   "https://fleetlock.example.com",
				FleetlockGroup: "default",
				CheckInterval:  "2m",
				RetryInterval:  "3m",
				LogLevel:       "error",
				KubeletConfig:  "/foo/kubelet.conf",
				KubeadmPath:    "/foo/kubeadm",
			},
		},
		{
			Name: "GroupNil",
			Global: api.UpgradedConfig{
				Stream:         "registry.example.com/test-stream",
				FleetlockURL:   "https://fleetlock.example.com",
				FleetlockGroup: "default",
				CheckInterval:  "2m",
				RetryInterval:  "3m",
				LogLevel:       "error",
				KubeletConfig:  "/foo/kubelet.conf",
				KubeadmPath:    "/foo/kubeadm",
			},
			Result: &api.UpgradedConfig{
				Stream:         "registry.example.com/test-stream",
				FleetlockURL:   "https://fleetlock.example.com",
				FleetlockGroup: "default",
				CheckInterval:  "2m",
				RetryInterval:  "3m",
				LogLevel:       "error",
				KubeletConfig:  "/foo/kubelet.conf",
				KubeadmPath:    "/foo/kubeadm",
			},
		},
		{
			Name:   "AllNil",
			Result: &api.UpgradedConfig{},
		},
	}
	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert.Equal(t, tCase.Result, combineConfig(tCase.Global, tCase.Group), "Should combine the 2 configs")
		})
	}
}

func TestDeleteConfigAnnotations(t *testing.T) {
	tMatrix := []struct {
		Name             string
		Original, Result map[string]string
		Changed          bool
	}{
		{
			Name: "DeleteConfigAnnotations",
			Original: map[string]string{
				constants.ConfigStream:         "registry.example.org/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.org",
				constants.ConfigFleetlockGroup: "not-default",
				constants.ConfigCheckInterval:  "3h0m0s",
				constants.ConfigRetryInterval:  "5m0s",
			},
			Result:  map[string]string{},
			Changed: true,
		},
		{
			Name: "DoNotTouchUnrelatedAnnotations",
			Original: map[string]string{
				constants.ConfigStream:         "registry.example.org/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.org",
				constants.ConfigFleetlockGroup: "not-default",
				constants.ConfigCheckInterval:  "3h0m0s",
				constants.ConfigRetryInterval:  "5m0s",
				"example.com/test":             "true",
			},
			Result: map[string]string{
				"example.com/test": "true",
			},
			Changed: true,
		},
		{
			Name: "Unchanged",
			Original: map[string]string{
				"example.com/test": "true",
			},
			Result: map[string]string{
				"example.com/test": "true",
			},
			Changed: false,
		},
		{
			Name:     "AnnotationsNil",
			Original: nil,
			Result:   nil,
			Changed:  false,
		},
	}
	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tCase.Changed, deleteConfigAnnotations(tCase.Original), "Should indicate if annotations changed")
			assert.Equal(tCase.Result, tCase.Original, "Should have applied the annotations")
		})
	}
}
