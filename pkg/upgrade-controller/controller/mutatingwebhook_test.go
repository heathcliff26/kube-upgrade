package controller

import (
	"testing"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
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
					Upgraded: api.UpgradedConfig{
						Stream:         api.DefaultUpgradedStream,
						FleetlockGroup: api.DefaultUpgradedFleetlockGroup,
						CheckInterval:  api.DefaultUpgradedCheckInterval,
						RetryInterval:  api.DefaultUpgradedRetryInterval,
						LogLevel:       api.DefaultUpgradedLogLevel,
						KubeletConfig:  api.DefaultUpgradedKubeletConfig,
					},
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
					Upgraded: api.UpgradedConfig{
						Stream:         api.DefaultUpgradedStream,
						FleetlockGroup: api.DefaultUpgradedFleetlockGroup,
						CheckInterval:  api.DefaultUpgradedCheckInterval,
						RetryInterval:  api.DefaultUpgradedRetryInterval,
						LogLevel:       api.DefaultUpgradedLogLevel,
						KubeletConfig:  api.DefaultUpgradedKubeletConfig,
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
						"compute": {
							DependsOn: []string{"control-plane"},
							Upgraded: &api.UpgradedConfig{
								Stream:         "docker.io/heathcliff26/fcos-k8s",
								FleetlockURL:   "https://fleetlock.example.com",
								FleetlockGroup: "test",
								CheckInterval:  "1m",
								RetryInterval:  "30s",
							},
						},
					},
				},
			},
			Result: &api.KubeUpgradePlan{
				Spec: api.KubeUpgradeSpec{
					Groups: map[string]api.KubeUpgradePlanGroup{
						"control-plane": {
							Labels:   map[string]string{},
							Upgraded: &api.UpgradedConfig{},
						},
						"compute": {
							DependsOn: []string{"control-plane"},
							Labels:    map[string]string{},
							Upgraded: &api.UpgradedConfig{
								Stream:         "docker.io/heathcliff26/fcos-k8s",
								FleetlockURL:   "https://fleetlock.example.com",
								FleetlockGroup: "test",
								CheckInterval:  "1m",
								RetryInterval:  "30s",
							},
						},
					},
					Upgraded: api.UpgradedConfig{
						Stream:         api.DefaultUpgradedStream,
						FleetlockGroup: api.DefaultUpgradedFleetlockGroup,
						CheckInterval:  api.DefaultUpgradedCheckInterval,
						RetryInterval:  api.DefaultUpgradedRetryInterval,
						LogLevel:       api.DefaultUpgradedLogLevel,
						KubeletConfig:  api.DefaultUpgradedKubeletConfig,
					},
				},
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			err := (&planMutatingHook{}).Default(t.Context(), tCase.Plan)

			assert.NoError(err, "Should succeed")
			assert.Equal(tCase.Result, tCase.Plan)
		})
	}
	t.Run("InvalidObject", func(t *testing.T) {
		assert.Error(t, (&planMutatingHook{}).Default(t.Context(), &corev1.Pod{}), "Should return an error")
	})
}
