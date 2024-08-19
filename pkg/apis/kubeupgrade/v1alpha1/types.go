package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	PlanStatusUnknown     = "Unknown"
	PlanStatusProgressing = "Progressing"
	PlanStatusWaiting     = "Waiting"
	PlanStatusComplete    = "Complete"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:JSONPath=.spec.kubernetesVersion,name=Version,type=string,description="The targeted kubernetes version"
// +kubebuilder:printcolumn:JSONPath=.status.summary,name=Status,type=string,description="A summary of the overall status of the cluster"
// +kubebuilder:resource:scope=Cluster,shortName=plan
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
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
	// If the actual version differs, the cluster will be upgraded
	// +required
	// +kubebuilder:example=v1.31.0
	// +kubebuilder:validation:Pattern=v[0-9]+\.[0-9]+\.[0-9]+
	KubernetesVersion string `json:"kubernetesVersion"`

	// The different groups in which the nodes will be upgraded.
	// At minimum needs to separate control-plane from compute nodes, to ensure that control-plane nodes will be upgraded first.
	// +required
	// +kubebuilder:validation:MinProperties=1
	Groups map[string]KubeUpgradePlanGroup `json:"groups"`
}

type KubeUpgradePlanGroup struct {
	// Specify group(s) that should be upgraded first.
	// Should be used to ensure control-plane nodes are upgraded first.
	// +optional
	// +kubebuilder:example=control-plane
	DependsOn []string `json:"dependsOn,omitempty"`
	// The labels by which to filter for this group
	// +required
	// +kubebuilder:validation:MinProperties=1
	// +kubebuilder:example="node-role.kubernetes.io/control-plane;node-role.kubernetes.io/compute"
	Labels map[string]string `json:"labels"`
}

type KubeUpgradeStatus struct {
	// A summary of the overall status of the cluster
	// +kubebuilder:validation:Enum=Unknown;Waiting;Progressing;Complete
	// +default="Unknown"
	Summary string `json:"summary,omitempty"`
	// The current status of each group
	Groups map[string]string `json:"groups,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
type KubeUpgradePlanList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []KubeUpgradePlan `json:"items"`
}