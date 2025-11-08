package controller

import (
	"maps"
	"reflect"
	"strings"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
)

// Combine 2 configs, where group overrides the values used by global.
// Return the result as ready to use annotations.
func combineConfig(global api.UpgradedConfig, group *api.UpgradedConfig) map[string]string {
	if group == nil {
		return createConfigAnnotations(global)
	}
	cfg := global

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

	return createConfigAnnotations(cfg)
}

// Convert the provided config to node annotations
func createConfigAnnotations(cfg api.UpgradedConfig) map[string]string {
	res := make(map[string]string, 5)

	if cfg.Stream != "" {
		res[constants.ConfigStream] = cfg.Stream
	}
	if cfg.FleetlockURL != "" {
		res[constants.ConfigFleetlockURL] = cfg.FleetlockURL
	}
	if cfg.FleetlockGroup != "" {
		res[constants.ConfigFleetlockGroup] = cfg.FleetlockGroup
	}
	if cfg.CheckInterval != "" {
		res[constants.ConfigCheckInterval] = cfg.CheckInterval
	}
	if cfg.RetryInterval != "" {
		res[constants.ConfigRetryInterval] = cfg.RetryInterval
	}

	return res
}

// Apply the provided configuration annotations to the node.
// Will delete unspecified config options from node Annotations.
// Returns if the config changed.
func applyConfigAnnotations(annotations map[string]string, cfg map[string]string) bool {
	original := make(map[string]string, len(annotations))
	maps.Copy(original, annotations)

	for k := range annotations {
		if strings.HasPrefix(k, constants.ConfigPrefix) {
			delete(annotations, k)
		}
	}

	maps.Copy(annotations, cfg)
	return !reflect.DeepEqual(original, annotations)
}
