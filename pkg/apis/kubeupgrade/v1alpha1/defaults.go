package v1alpha1

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
}

func SetObjectDefaults_KubeUpgradeStatus(status *KubeUpgradeStatus) {
	if status.Summary == "" {
		status.Summary = "Unknown"
	}
	if status.Groups == nil {
		status.Groups = make(map[string]string)
	}
}
