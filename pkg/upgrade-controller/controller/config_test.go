package controller

import (
	"testing"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha2"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/stretchr/testify/assert"
)

func TestCombineConfig(t *testing.T) {
	tMatrix := []struct {
		Name   string
		Global api.UpgradedConfig
		Group  *api.UpgradedConfig
		Result map[string]string
	}{
		{
			Name: "OverrideAll",
			Global: api.UpgradedConfig{
				Stream:         "registry.example.org/test-stream",
				FleetlockURL:   "https://fleetlock.example.org",
				FleetlockGroup: "not-default",
				CheckInterval:  "10m",
				RetryInterval:  "15m",
			},
			Group: &api.UpgradedConfig{
				Stream:         "registry.example.com/test-stream",
				FleetlockURL:   "https://fleetlock.example.com",
				FleetlockGroup: "default",
				CheckInterval:  "2m",
				RetryInterval:  "3m",
			},
			Result: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m",
				constants.ConfigRetryInterval:  "3m",
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
			},
			Group: &api.UpgradedConfig{
				FleetlockGroup: "default",
				CheckInterval:  "2m",
				RetryInterval:  "3m",
			},
			Result: map[string]string{
				constants.ConfigStream:         "registry.example.org/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.org",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m",
				constants.ConfigRetryInterval:  "3m",
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
			},
			Group: &api.UpgradedConfig{
				Stream:       "registry.example.com/test-stream",
				FleetlockURL: "https://fleetlock.example.com",
			},
			Result: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "not-default",
				constants.ConfigCheckInterval:  "10m",
				constants.ConfigRetryInterval:  "15m",
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
			},
			Result: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m",
				constants.ConfigRetryInterval:  "3m",
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
			},
			Result: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m",
				constants.ConfigRetryInterval:  "3m",
			},
		},
		{
			Name:   "AllNil",
			Result: map[string]string{},
		},
	}
	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert.Equal(t, tCase.Result, combineConfig(tCase.Global, tCase.Group), "Should combine the 2 configs")
		})
	}
}

func TestApplyConfigAnnotations(t *testing.T) {
	tMatrix := []struct {
		Name                     string
		Original, Config, Result map[string]string
		Changed                  bool
	}{
		{
			Name:     "ApplyNewConfig",
			Original: map[string]string{},
			Config: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m0s",
				constants.ConfigRetryInterval:  "3m0s",
			},
			Result: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m0s",
				constants.ConfigRetryInterval:  "3m0s",
			},
			Changed: true,
		},
		{
			Name: "OverrideOldConfig",
			Original: map[string]string{
				constants.ConfigStream:         "registry.example.org/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.org",
				constants.ConfigFleetlockGroup: "not-default",
				constants.ConfigCheckInterval:  "3h0m0s",
				constants.ConfigRetryInterval:  "5m0s",
			},
			Config: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m0s",
				constants.ConfigRetryInterval:  "3m0s",
			},
			Result: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m0s",
				constants.ConfigRetryInterval:  "3m0s",
			},
			Changed: true,
		},
		{
			Name: "DeleteOldAnnotations",
			Original: map[string]string{
				constants.ConfigStream:         "registry.example.org/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.org",
				constants.ConfigFleetlockGroup: "not-default",
				constants.ConfigCheckInterval:  "3h0m0s",
				constants.ConfigRetryInterval:  "5m0s",
			},
			Config: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m0s",
			},
			Result: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m0s",
			},
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
			Config: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m0s",
				constants.ConfigRetryInterval:  "3m0s",
			},
			Result: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m0s",
				constants.ConfigRetryInterval:  "3m0s",
				"example.com/test":             "true",
			},
			Changed: true,
		},
		{
			Name: "Unchanged",
			Original: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m0s",
				constants.ConfigRetryInterval:  "3m0s",
			},
			Config: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m0s",
				constants.ConfigRetryInterval:  "3m0s",
			},
			Result: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m0s",
				constants.ConfigRetryInterval:  "3m0s",
			},
			Changed: false,
		},
		{
			Name: "DeleteAll",
			Original: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m0s",
				constants.ConfigRetryInterval:  "3m0s",
			},
			Config:  map[string]string{},
			Result:  map[string]string{},
			Changed: true,
		},
		{
			Name: "ConfigNil",
			Original: map[string]string{
				constants.ConfigStream:         "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:   "https://fleetlock.example.com",
				constants.ConfigFleetlockGroup: "default",
				constants.ConfigCheckInterval:  "2m0s",
				constants.ConfigRetryInterval:  "3m0s",
			},
			Config:  nil,
			Result:  map[string]string{},
			Changed: true,
		},
	}
	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tCase.Changed, applyConfigAnnotations(tCase.Original, tCase.Config), "Should indicate if annotations changed")
			assert.Equal(tCase.Result, tCase.Original, "Should have applied the annotations")
		})
	}
}
