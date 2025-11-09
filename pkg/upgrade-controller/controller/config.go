package controller

import (
	"strings"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
)

// Combine 2 configs, where group overrides the values used by global.
// Returns the combined configuration.
func combineConfig(global api.UpgradedConfig, group *api.UpgradedConfig) *api.UpgradedConfig {
	cfg := global
	if group == nil {
		return &cfg
	}

	if group.Stream != "" {
		cfg.Stream = group.Stream
	}
	if group.FleetlockURL != "" {
		cfg.FleetlockURL = group.FleetlockURL
	}
	if group.FleetlockGroup != "" {
		cfg.FleetlockGroup = group.FleetlockGroup
	}
	if group.CheckInterval != "" {
		cfg.CheckInterval = group.CheckInterval
	}
	if group.RetryInterval != "" {
		cfg.RetryInterval = group.RetryInterval
	}
	if group.LogLevel != "" {
		cfg.LogLevel = group.LogLevel
	}
	if group.KubeletConfig != "" {
		cfg.KubeletConfig = group.KubeletConfig
	}
	if group.KubeadmPath != "" {
		cfg.KubeadmPath = group.KubeadmPath
	}

	return &cfg
}

// Delete all config annotations from the node.
// Returns if the config changed.
func deleteConfigAnnotations(annotations map[string]string) bool {
	changed := false

	for k := range annotations {
		if strings.HasPrefix(k, constants.ConfigPrefix) {
			delete(annotations, k)
			changed = true
		}
	}
	return changed
}
