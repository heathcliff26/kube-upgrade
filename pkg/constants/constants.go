package constants

const (
	baseDomain = "kube-upgrade.heathcliff.eu/"
	nodePrefix = "node." + baseDomain
)

const (
	NodeKubernetesVersion = nodePrefix + "kubernetesVersion"
	NodeUpgradeStatus     = nodePrefix + "status"
)

const (
	NodeUpgradeStatusPending   = "pending"
	NodeUpgradeStatusRebasing  = "rebasing"
	NodeUpgradeStatusUpgrading = "upgrading"
	NodeUpgradeStatusCompleted = "completed"
)
