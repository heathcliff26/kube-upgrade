package constants

const (
	BaseDomain   = "kube-upgrade.heathcliff.eu/"
	NodePrefix   = "node." + BaseDomain
	ConfigPrefix = "config." + BaseDomain
)

const (
	NodeKubernetesVersion = NodePrefix + "kubernetesVersion"
	NodeUpgradeStatus     = NodePrefix + "status"
)

const (
	NodeUpgradeStatusPending   = "pending"
	NodeUpgradeStatusRebasing  = "rebasing"
	NodeUpgradeStatusUpgrading = "upgrading"
	NodeUpgradeStatusCompleted = "completed"
	NodeUpgradeStatusError     = "error"
)

const (
	ConfigStream         = ConfigPrefix + "stream"
	ConfigFleetlockURL   = ConfigPrefix + "fleetlock-url"
	ConfigFleetlockGroup = ConfigPrefix + "fleetlock-group"
	ConfigCheckInterval  = ConfigPrefix + "check-interval"
	ConfigRetryInterval  = ConfigPrefix + "retry-interval"
)
