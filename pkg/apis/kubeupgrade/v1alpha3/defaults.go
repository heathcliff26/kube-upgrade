package v1alpha3

const (
	DefaultStatus                 = "Unknown"
	DefaultUpgradedStream         = "ghcr.io/heathcliff26/fcos-k8s"
	DefaultUpgradedFleetlockGroup = "default"
	DefaultUpgradedCheckInterval  = "3h"
	DefaultUpgradedRetryInterval  = "1m"
)

func SetObjectDefaults_KubeUpgradeSpec(spec *KubeUpgradeSpec) {
	if spec.Groups == nil {
		spec.Groups = make(map[string]KubeUpgradePlanGroup)
	}
	for name, group := range spec.Groups {
		if group.Labels == nil {
			group.Labels = make(map[string]string)
		}
		spec.Groups[name] = group
	}
	SetObjectDefaults_UpgradedConfig(&spec.Upgraded)
}

func SetObjectDefaults_UpgradedConfig(cfg *UpgradedConfig) {
	if cfg.Stream == "" {
		cfg.Stream = DefaultUpgradedStream
	}
	if cfg.FleetlockGroup == "" {
		cfg.FleetlockGroup = DefaultUpgradedFleetlockGroup
	}
	if cfg.CheckInterval == "" {
		cfg.CheckInterval = DefaultUpgradedCheckInterval
	}
	if cfg.RetryInterval == "" {
		cfg.RetryInterval = DefaultUpgradedRetryInterval
	}
}
