package controller

import (
	"context"
	"fmt"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-kubeupgrade-heathcliff-eu-v1alpha3-kubeupgradeplan,mutating=false,failurePolicy=fail,groups=kubeupgrade.heathcliff.eu,resources=kubeupgradeplans,verbs=create;update,versions=v1alpha3,name=kubeupgrade.heathcliff.eu,admissionReviewVersions=v1,sideEffects=None

// planValidatingHook validates the plan
type planValidatingHook struct{}

// Validate all values of the plan and check if they are sensible
func (*planValidatingHook) validate(obj runtime.Object) (admission.Warnings, error) {
	plan, ok := obj.(*api.KubeUpgradePlan)
	if !ok {
		return nil, fmt.Errorf("expected a KubeUpgradePlan but got a %T", obj)
	}

	err := api.ValidateObject_KubeUpgradePlan(plan)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateCreate validates the object on creation.
// The optional warnings will be added to the response as warning messages.
// Return an error if the object is invalid.
func (p *planValidatingHook) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	return p.validate(obj)
}

// ValidateUpdate validates the object on update.
// The optional warnings will be added to the response as warning messages.
// Return an error if the object is invalid.
func (p *planValidatingHook) ValidateUpdate(_ context.Context, _ runtime.Object, newObj runtime.Object) (admission.Warnings, error) {
	return p.validate(newObj)
}

// ValidateDelete validates the object on deletion.
// The optional warnings will be added to the response as warning messages.
// Return an error if the object is invalid.
func (*planValidatingHook) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}
