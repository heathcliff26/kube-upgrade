package constants

const (
	KubernetesVersionAnnotation = "kubernetes-version.kube-upgrade.heathcliff.eu"
	KubernetesUpgradeStatus     = "upgrade-status.kube-upgrade.heathcliff.eu"
)

const (
	NodeUpgradeStatusPending   = "pending"
	NodeUpgradeStatusRebasing  = "rebasing"
	NodeUpgradeStatusUpgrading = "upgrading"
	NodeUpgradeStatusCompleted = "completed"
)
