package v1alpha3

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PlanStatusUnknown     = "Unknown"
	PlanStatusProgressing = "Progressing"
	PlanStatusWaiting     = "Waiting"
	PlanStatusComplete    = "Complete"
	PlanStatusError       = "Error"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:JSONPath=.spec.kubernetesVersion,name=Version,type=string,description="The targeted kubernetes version"
// +kubebuilder:printcolumn:JSONPath=.status.summary,name=Status,type=string,description="A summary of the overall status of the cluster"
// +kubebuilder:resource:scope=Cluster,shortName=plan
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
type KubeUpgradePlan struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec KubeUpgradeSpec `json:"spec" valid:"required"`

	// +optional
	Status KubeUpgradeStatus `json:"status,omitempty"`
}

type KubeUpgradeSpec struct {
	// The kubernetes version the cluster should be at.
	// If the actual version differs, the cluster will be upgraded.
	// +required
	// +kubebuilder:example=v1.31.0
	KubernetesVersion string `json:"kubernetesVersion"`

	// Allow downgrading to older kubernetes versions.
	// Only enable if you know what you are doing.
	// +optional
	// +default=false
	AllowDowngrade bool `json:"allowDowngrade,omitempty"`

	// The different groups in which the nodes will be upgraded.
	// At minimum needs to separate control-plane from compute nodes, to ensure that control-plane nodes will be upgraded first.
	// +required
	// +kubebuilder:validation:MinProperties=1
	Groups map[string]KubeUpgradePlanGroup `json:"groups"`

	// The configuration for all upgraded daemons. Can be overwritten by group specific config.
	// +required
	Upgraded UpgradedConfig `json:"upgraded,omitempty"`
}

type KubeUpgradePlanGroup struct {
	// Specify group(s) that should be upgraded first.
	// Should be used to ensure control-plane nodes are upgraded first.
	// +optional
	// +listType=atomic
	// +kubebuilder:example=control-plane
	DependsOn []string `json:"dependsOn,omitempty"`

	// The labels by which to filter nodes for this group
	// +required
	// +kubebuilder:example="node-role.kubernetes.io/control-plane;node-role.kubernetes.io/compute"
	Labels map[string]string `json:"labels"`

	// Enable the upgraded pods to be scheduled on tainted nodes like control-planes.
	// +optional
	// +listType=atomic
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// The configuration for all upgraded daemons in the group. Overwrites global parameters.
	// +optional
	// +nullable
	Upgraded *UpgradedConfig `json:"upgraded,omitempty"`
}

type KubeUpgradeStatus struct {
	// A summary of the overall status of the cluster
	Summary string `json:"summary,omitempty"`

	// The current status of each group
	Groups map[string]string `json:"groups,omitempty"`
}

type UpgradedConfig struct {
	// The container image repository for os rebases
	// +optional
	// +kubebuilder:example="ghcr.io/heathcliff26/fcos-k8s"
	Stream string `json:"stream,omitempty"`

	// URL for the fleetlock server. Is required to be set globally.
	// +optional
	// +kubebuilder:example="https://fleetlock.example.com"
	FleetlockURL string `json:"fleetlockUrl"`

	// The group to use for fleetlock
	// +kubebuilder:example="control-plane;compute"
	FleetlockGroup string `json:"fleetlockGroup,omitempty"`

	// The interval between regular checks
	// +optional
	// +kubebuilder:validation:Format=go-duration
	// +kubebuilder:example="3h;24h;30m"
	CheckInterval string `json:"checkInterval,omitempty"`

	// The interval between retries when an operation fails
	// +optional
	// +kubebuilder:validation:Format=go-duration
	// +kubebuilder:example="5m;1m;30s"
	RetryInterval string `json:"retryInterval,omitempty"`

	// The log level used by slog, default "info"
	// +optional
	// +kubebuilder:validation:Enum=debug;info;warn;error
	// +kubebuilder:example="debug;info;warn;error"
	LogLevel string `json:"logLevel,omitempty"`

	// The path to the kubelet config file on the node
	// +optional
	// +kubebuilder:example="/etc/kubernetes/kubelet.conf"
	KubeletConfig string `json:"kubeletConfig,omitempty"`

	// The path to the kubeadm binary on the node. Upgraded will download kubeadm if no path is provided.
	// +optional
	// +kubebuilder:example="/usr/bin/kubeadm"
	KubeadmPath string `json:"kubeadmPath,omitempty"`

	// Allow unsigned ostree images for rebase. It is recommended to use signed images instead.
	// +optional
	AllowUnsignedOstreeImages bool `json:"allowUnsignedOstreeImages,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
type KubeUpgradePlanList struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []KubeUpgradePlan `json:"items"`
}
