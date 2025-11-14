package controller

import (
	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
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
