package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
)

// +kubebuilder:webhook:path=/mutate-kubeupgrade-heathcliff-eu-v1alpha3-kubeupgradeplan,mutating=true,failurePolicy=fail,groups=kubeupgrade.heathcliff.eu,resources=kubeupgradeplans,verbs=create;update,versions=v1alpha3,name=kubeupgrade.heathcliff.eu,admissionReviewVersions=v1,sideEffects=None

// planMutatingHook sets the defaults for the plan
type planMutatingHook struct{}

func (*planMutatingHook) Default(_ context.Context, obj runtime.Object) error {
	plan, ok := obj.(*api.KubeUpgradePlan)
	if !ok {
		return fmt.Errorf("expected a KubeUpgradePlan but got a %T", obj)
	}

	api.SetObjectDefaults_KubeUpgradePlan(plan)

	return nil
}
