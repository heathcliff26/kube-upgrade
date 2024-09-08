package controller

import (
	"context"
	"testing"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestDefault(t *testing.T) {
	tMatrix := []struct {
		Name         string
		Plan, Result *api.KubeUpgradePlan
	}{
		{
			Name: "EmptyPlan",
			Plan: &api.KubeUpgradePlan{},
			Result: &api.KubeUpgradePlan{
				Spec: api.KubeUpgradeSpec{
					Groups: map[string]api.KubeUpgradePlanGroup{},
				},
			},
		},
		{
			Name: "Groups",
			Plan: &api.KubeUpgradePlan{
				Spec: api.KubeUpgradeSpec{
					Groups: map[string]api.KubeUpgradePlanGroup{
						"control-plane": {},
					},
				},
			},
			Result: &api.KubeUpgradePlan{
				Spec: api.KubeUpgradeSpec{
					Groups: map[string]api.KubeUpgradePlanGroup{
						"control-plane": {
							Labels: map[string]string{},
						},
					},
				},
			},
		},
		{
			Name: "UpgradedConfig",
			Plan: &api.KubeUpgradePlan{
				Spec: api.KubeUpgradeSpec{
					Groups: map[string]api.KubeUpgradePlanGroup{
						"control-plane": {
							Upgraded: &api.UpgradedConfig{},
						},
					},
					Upgraded: &api.UpgradedConfig{},
				},
			},
			Result: &api.KubeUpgradePlan{
				Spec: api.KubeUpgradeSpec{
					Groups: map[string]api.KubeUpgradePlanGroup{
						"control-plane": {
							Labels:   map[string]string{},
							Upgraded: &api.UpgradedConfig{},
						},
					},
					Upgraded: &api.UpgradedConfig{
						Stream:         "ghcr.io/heathcliff26/fcos-k8s",
						FleetlockGroup: "default",
						CheckInterval:  "3h",
						RetryInterval:  "5m",
					},
				},
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			err := (&planMutatingHook{}).Default(context.Background(), tCase.Plan)

			assert.NoError(err, "Should succeed")
			assert.Equal(tCase.Result, tCase.Plan)
		})
	}
	t.Run("InvalidObject", func(t *testing.T) {
		assert.Error(t, (&planMutatingHook{}).Default(context.Background(), &corev1.Pod{}), "Should return an error")
	})
}
