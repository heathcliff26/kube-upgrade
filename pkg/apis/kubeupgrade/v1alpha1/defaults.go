package v1alpha1

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
	if cfg.CheckInterval == "" {
		cfg.CheckInterval = "3h"
	}
	if cfg.RetryInterval == "" {
		cfg.RetryInterval = "5m"
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
