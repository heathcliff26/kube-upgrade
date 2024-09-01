package v1alpha1

import "time"

func SetObjectDefaults_KubeUpgradeSpec(spec *KubeUpgradeSpec) {
	if spec.Groups == nil {
		spec.Groups = make(map[string]KubeUpgradePlanGroup)
	}
	for name, group := range spec.Groups {
		if group.Labels == nil {
			group.Labels = make(map[string]string)
		}
		if group.Upgraded != nil {
			SetObjectDefaults_UpgradedConfig(group.Upgraded)
		}
		spec.Groups[name] = group
	}
	if spec.Upgraded != nil {
		SetObjectDefaults_UpgradedConfig(spec.Upgraded)
	}
}

func SetObjectDefaults_UpgradedConfig(cfg *UpgradedConfig) {
	if cfg.Stream == "" {
		cfg.Stream = "ghcr.io/heathcliff26/fcos-k8s"
	}
	if cfg.FleetlockGroup == "" {
		cfg.FleetlockGroup = "default"
	}
	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = time.Hour * 3
	}
	if cfg.RetryInterval == 0 {
		cfg.RetryInterval = time.Minute * 5
	}
}

func SetObjectDefaults_KubeUpgradeStatus(status *KubeUpgradeStatus) {
	if status.Summary == "" {
		status.Summary = "Unknown"
	}
	if status.Groups == nil {
		status.Groups = make(map[string]string)
	}
}
