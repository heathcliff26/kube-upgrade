package v1alpha2

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	DefaultStatus                 = "Unknown"
	DefaultUpgradedStream         = "ghcr.io/heathcliff26/fcos-k8s"
	DefaultUpgradedFleetlockGroup = "default"
	DefaultUpgradedCheckInterval  = "3h"
	DefaultUpgradedRetryInterval  = "5m"
)

func SetObjectDefaults_KubeUpgradeSpec(spec *KubeUpgradeSpec) {
	if spec.Groups == nil {
		spec.Groups = make(map[string]KubeUpgradePlanGroup)
	}
	for name, group := range spec.Groups {
		if group.Labels == nil {
			group.Labels = &metav1.LabelSelector{}
		}
		spec.Groups[name] = group
	}
	SetObjectDefaults_UpgradedConfig(&spec.Upgraded)
}

func SetObjectDefaults_UpgradedConfig(cfg *UpgradedConfig) {
	if cfg.Stream == "" {
		cfg.Stream = "ghcr.io/heathcliff26/fcos-k8s"
	}
	if cfg.FleetlockGroup == "" {
		cfg.FleetlockGroup = "default"
	}
	if cfg.CheckInterval == "" {
		cfg.CheckInterval = "3h"
	}
	if cfg.RetryInterval == "" {
		cfg.RetryInterval = "5m"
	}
}
