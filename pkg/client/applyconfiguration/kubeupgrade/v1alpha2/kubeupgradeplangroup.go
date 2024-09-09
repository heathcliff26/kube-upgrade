// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha2

import (
	v1 "k8s.io/client-go/applyconfigurations/meta/v1"
)

// KubeUpgradePlanGroupApplyConfiguration represents a declarative configuration of the KubeUpgradePlanGroup type for use
// with apply.
type KubeUpgradePlanGroupApplyConfiguration struct {
	DependsOn []string                            `json:"dependsOn,omitempty"`
	Labels    *v1.LabelSelectorApplyConfiguration `json:"labels,omitempty"`
	Upgraded  *UpgradedConfigApplyConfiguration   `json:"upgraded,omitempty"`
}

// KubeUpgradePlanGroupApplyConfiguration constructs a declarative configuration of the KubeUpgradePlanGroup type for use with
// apply.
func KubeUpgradePlanGroup() *KubeUpgradePlanGroupApplyConfiguration {
	return &KubeUpgradePlanGroupApplyConfiguration{}
}

// WithDependsOn adds the given value to the DependsOn field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the DependsOn field.
func (b *KubeUpgradePlanGroupApplyConfiguration) WithDependsOn(values ...string) *KubeUpgradePlanGroupApplyConfiguration {
	for i := range values {
		b.DependsOn = append(b.DependsOn, values[i])
	}
	return b
}

// WithLabels sets the Labels field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Labels field is set to the value of the last call.
func (b *KubeUpgradePlanGroupApplyConfiguration) WithLabels(value *v1.LabelSelectorApplyConfiguration) *KubeUpgradePlanGroupApplyConfiguration {
	b.Labels = value
	return b
}

// WithUpgraded sets the Upgraded field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Upgraded field is set to the value of the last call.
func (b *KubeUpgradePlanGroupApplyConfiguration) WithUpgraded(value *UpgradedConfigApplyConfiguration) *KubeUpgradePlanGroupApplyConfiguration {
	b.Upgraded = value
	return b
}
