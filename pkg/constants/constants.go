package constants

const (
	baseDomain   = "kube-upgrade.heathcliff.eu/"
	nodePrefix   = "node." + baseDomain
	configPrefix = "config." + baseDomain
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

const (
	ConfigStream         = configPrefix + "stream"
	ConfigFleetlockURL   = configPrefix + "fleetlock-url"
	ConfigFleetlockGroup = configPrefix + "fleetlock-group"
	ConfigCheckInterval  = configPrefix + "check-interval"
	ConfigRetryInterval  = configPrefix + "retry-interval"
)
