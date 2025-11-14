package constants

const (
	BaseDomain   = "kube-upgrade.heathcliff.eu/"
	NodePrefix   = "node." + BaseDomain
	ConfigPrefix = "config." + BaseDomain
)

const (
	NodeKubernetesVersion = NodePrefix + "kubernetesVersion"
	NodeUpgradeStatus     = NodePrefix + "status"
	NodeUpgradedVersion   = NodePrefix + "upgradedVersion"
)

const (
	NodeUpgradeStatusPending   = "pending"
	NodeUpgradeStatusRebasing  = "rebasing"
	NodeUpgradeStatusUpgrading = "upgrading"
	NodeUpgradeStatusCompleted = "completed"
	NodeUpgradeStatusError     = "error"
)

// TODO: Remove when removing migration code in v0.7.0
const (
	ConfigStream         = ConfigPrefix + "stream"
	ConfigFleetlockURL   = ConfigPrefix + "fleetlock-url"
	ConfigFleetlockGroup = ConfigPrefix + "fleetlock-group"
	ConfigCheckInterval  = ConfigPrefix + "check-interval"
	ConfigRetryInterval  = ConfigPrefix + "retry-interval"
)

const (
	LabelPlanName  = BaseDomain + "plan"
	LabelNodeGroup = BaseDomain + "group"
)

// TODO: Remove in future release when migration code is removed
const Finalizer = BaseDomain + "finalizer"
