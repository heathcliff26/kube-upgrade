package constants

const (
	BaseDomain       = "kube-upgrade.heathcliff.eu/"
	NodePrefix       = "node." + BaseDomain
	ControllerPrefix = "controller." + BaseDomain
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

const (
	LabelPlanName  = BaseDomain + "plan"
	LabelNodeGroup = BaseDomain + "group"
)

const (
	ControllerResourceHash = ControllerPrefix + "checksum"
)

// TODO: Remove in future release when migration code is removed
const Finalizer = BaseDomain + "finalizer"
