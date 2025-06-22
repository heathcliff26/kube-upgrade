package v1alpha2

import (
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

	// The label selector by which to filter for this group
	// +required
	// +kubebuilder:example="node-role.kubernetes.io/control-plane;node-role.kubernetes.io/compute"
	Labels *metav1.LabelSelector `json:"labels"`

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
	// +default="ghcr.io/heathcliff26/fcos-k8s"
	// +kubebuilder:example="ghcr.io/heathcliff26/fcos-k8s"
	Stream string `json:"stream,omitempty"`

	// URL for the fleetlock server. Is required to be set globally.
	// +optional
	// +kubebuilder:example="https://fleetlock.example.com"
	FleetlockURL string `json:"fleetlock-url"`

	// The group to use for fleetlock
	// +default="default"
	// +kubebuilder:example="control-plane;compute"
	FleetlockGroup string `json:"fleetlock-group,omitempty"`

	// The interval between regular checks
	// +optional
	// +kubebuilder:validation:Format=go-duration
	// +default="3h"
	// +kubebuilder:example="3h;24h;30m"
	CheckInterval string `json:"check-interval,omitempty"`

	// The interval between retries when an operation fails
	// +optional
	// +kubebuilder:validation:Format=go-duration
	// +default="5m"
	// +kubebuilder:example="5m;1m;30s"
	RetryInterval string `json:"retry-interval,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
type KubeUpgradePlanList struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []KubeUpgradePlan `json:"items"`
}
