//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by defaulter-gen. DO NOT EDIT.

package v1alpha2

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// RegisterDefaults adds defaulters functions to the given scheme.
// Public to allow building arbitrary schemes.
// All generated defaulters are covering - they call all nested defaulters.
func RegisterDefaults(scheme *runtime.Scheme) error {
	scheme.AddTypeDefaultingFunc(&KubeUpgradePlan{}, func(obj interface{}) { SetObjectDefaults_KubeUpgradePlan(obj.(*KubeUpgradePlan)) })
	scheme.AddTypeDefaultingFunc(&KubeUpgradePlanList{}, func(obj interface{}) { SetObjectDefaults_KubeUpgradePlanList(obj.(*KubeUpgradePlanList)) })
	return nil
}

func SetObjectDefaults_KubeUpgradePlan(in *KubeUpgradePlan) {
	SetObjectDefaults_KubeUpgradeSpec(&in.Spec)
}

func SetObjectDefaults_KubeUpgradePlanList(in *KubeUpgradePlanList) {
	for i := range in.Items {
		a := &in.Items[i]
		SetObjectDefaults_KubeUpgradePlan(a)
	}
}
