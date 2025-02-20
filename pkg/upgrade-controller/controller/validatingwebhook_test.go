package controller

import (
	"testing"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha2"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidate(t *testing.T) {
	minimumValidPlan := &api.KubeUpgradePlan{
		Spec: api.KubeUpgradeSpec{
			KubernetesVersion: "v1.31.0",
			Groups: map[string]api.KubeUpgradePlanGroup{
				"control-plane": {
					Labels: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "node-role.kubernetes.io/control-plane",
								Operator: metav1.LabelSelectorOpExists,
							},
						},
					},
				},
			},
			Upgraded: api.UpgradedConfig{
				FleetlockURL: "https://fleetlock.example.com",
			},
		},
	}

	validMultipleGroups := minimumValidPlan.DeepCopy()
	validMultipleGroups.Spec.Groups["compute"] = api.KubeUpgradePlanGroup{
		DependsOn: []string{"control-plane"},
		Labels: &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "node-role.kubernetes.io/compute",
					Operator: metav1.LabelSelectorOpExists,
				},
			},
		},
	}

	validGroupWithUpgradedConfig := minimumValidPlan.DeepCopy()
	validGroupWithUpgradedConfig.Spec.Groups["compute"] = api.KubeUpgradePlanGroup{
		DependsOn: []string{"control-plane"},
		Labels: &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "node-role.kubernetes.io/compute",
					Operator: metav1.LabelSelectorOpExists,
				},
			},
		},
		Upgraded: &api.UpgradedConfig{
			FleetlockGroup: "compute",
		},
	}

	validKubernetesVersionWithPreRelease := minimumValidPlan.DeepCopy()
	validKubernetesVersionWithPreRelease.Spec.KubernetesVersion = "v1.31.0-rc.0"

	invalidKubernetesVersion := minimumValidPlan.DeepCopy()
	invalidKubernetesVersion.Spec.KubernetesVersion = "testv1.0.0"

	invalidMissingKubernetesVersion := minimumValidPlan.DeepCopy()
	invalidMissingKubernetesVersion.Spec.KubernetesVersion = ""

	invalidKubernetesVersionMajorOnly := minimumValidPlan.DeepCopy()
	invalidKubernetesVersionMajorOnly.Spec.KubernetesVersion = "v1"

	invalidMissingGroups := minimumValidPlan.DeepCopy()
	invalidMissingGroups.Spec.Groups = map[string]api.KubeUpgradePlanGroup{}

	invalidGroupDependsOn := minimumValidPlan.DeepCopy()
	invalidGroupDependsOn.Spec.Groups["compute"] = api.KubeUpgradePlanGroup{
		DependsOn: []string{"control-plane", "infra"},
		Labels: &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "node-role.kubernetes.io/compute",
					Operator: metav1.LabelSelectorOpExists,
				},
			},
		},
	}

	invalidMissingGroupLabel := minimumValidPlan.DeepCopy()
	invalidMissingGroupLabel.Spec.Groups["compute"] = api.KubeUpgradePlanGroup{
		DependsOn: []string{"control-plane"},
	}

	invalidGroupLabel := minimumValidPlan.DeepCopy()
	invalidGroupLabel.Spec.Groups["compute"] = api.KubeUpgradePlanGroup{
		DependsOn: []string{"control-plane"},
		Labels: &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "node-role.kubernetes.io/compute",
					Operator: "InvalidOperator",
				},
			},
		},
	}

	invalidGroupUpgradedConfig := minimumValidPlan.DeepCopy()
	invalidGroupUpgradedConfig.Spec.Groups["compute"] = api.KubeUpgradePlanGroup{
		DependsOn: []string{"control-plane"},
		Labels: &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "node-role.kubernetes.io/compute",
					Operator: metav1.LabelSelectorOpExists,
				},
			},
		},
		Upgraded: &api.UpgradedConfig{
			FleetlockURL: "not-a-url",
		},
	}

	invalidMissingUpgradedFleetlockURL := minimumValidPlan.DeepCopy()
	invalidMissingUpgradedFleetlockURL.Spec.Upgraded.FleetlockURL = ""

	invalidStream := minimumValidPlan.DeepCopy()
	invalidStream.Spec.Upgraded.Stream = "not-a-valid-stream- -"

	invalidFleetlockURL := minimumValidPlan.DeepCopy()
	invalidFleetlockURL.Spec.Upgraded.FleetlockURL = "not-a-url"

	invalidCheckInterval := minimumValidPlan.DeepCopy()
	invalidCheckInterval.Spec.Upgraded.CheckInterval = "not-a-duration"

	invalidRetryInterval := minimumValidPlan.DeepCopy()
	invalidRetryInterval.Spec.Upgraded.RetryInterval = "not-a-duration"

	tMatrix := []struct {
		Name  string
		Plan  *api.KubeUpgradePlan
		Error bool
	}{
		{
			Name:  "EmptyPlan",
			Plan:  &api.KubeUpgradePlan{},
			Error: true,
		},
		{
			Name: "MinimumPlan",
			Plan: minimumValidPlan,
		},
		{
			Name: "ValidMultipleGroups",
			Plan: validMultipleGroups,
		},
		{
			Name: "ValidGroupWithUpgradedConfig",
			Plan: validGroupWithUpgradedConfig,
		},
		{
			Name: "ValidKubernetesVersionWithPreRelease",
			Plan: validKubernetesVersionWithPreRelease,
		},
		{
			Name:  "InvalidKubernetesVersion",
			Plan:  invalidKubernetesVersion,
			Error: true,
		},
		{
			Name:  "InvalidMissingKubernetesVersion",
			Plan:  invalidMissingKubernetesVersion,
			Error: true,
		},
		{
			Name:  "InvalidKubernetesVersionMajorOnly",
			Plan:  invalidKubernetesVersionMajorOnly,
			Error: true,
		},
		{
			Name:  "InvalidMissingGroups",
			Plan:  invalidMissingGroups,
			Error: true,
		},
		{
			Name:  "InvalidGroupDependsOn",
			Plan:  invalidGroupDependsOn,
			Error: true,
		},
		{
			Name:  "InvalidMissingGroupLabel",
			Plan:  invalidMissingGroupLabel,
			Error: true,
		},
		{
			Name:  "InvalidGroupLabel",
			Plan:  invalidGroupLabel,
			Error: true,
		},
		{
			Name:  "InvalidGroupUpgradedConfig",
			Plan:  invalidGroupUpgradedConfig,
			Error: true,
		},
		{
			Name:  "InvalidMissingUpgradedFleetlockURL",
			Plan:  invalidMissingUpgradedFleetlockURL,
			Error: true,
		},
		{
			Name:  "InvalidStream",
			Plan:  invalidStream,
			Error: true,
		},
		{
			Name:  "InvalidFleetlockURL",
			Plan:  invalidFleetlockURL,
			Error: true,
		},
		{
			Name:  "InvalidCheckInterval",
			Plan:  invalidCheckInterval,
			Error: true,
		},
		{
			Name:  "InvalidRetryInterval",
			Plan:  invalidRetryInterval,
			Error: true,
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			if !assert.NoError((&planMutatingHook{}).Default(t.Context(), tCase.Plan), "Should add defaults to plan") {
				t.FailNow()
			}

			warn, err := (&planValidatingHook{}).validate(tCase.Plan)

			assert.Nil(warn, "Should not output any warnings")

			if tCase.Error {
				assert.Error(err, "Plan should be invalid")
			} else {
				assert.NoError(err, "Plan should be valid")
			}
		})
	}
	t.Run("InvalidObject", func(t *testing.T) {
		assert := assert.New(t)

		warn, err := (&planValidatingHook{}).validate(&corev1.Pod{})

		assert.Nil(warn, "Should not return a warning")
		assert.Error(err, "Should return an error")
	})
}

func TestValidateCreate(t *testing.T) {
	assert := assert.New(t)

	ctx := t.Context()

	warn, err := (&planValidatingHook{}).ValidateCreate(ctx, &api.KubeUpgradePlan{})

	assert.Nil(warn, "Should not return a warning")
	assert.Error(err, "Should return an error")

	plan := &api.KubeUpgradePlan{
		Spec: api.KubeUpgradeSpec{
			KubernetesVersion: "v1.31.0",
			Groups: map[string]api.KubeUpgradePlanGroup{
				"control-plane": {
					Labels: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "node-role.kubernetes.io/control-plane",
								Operator: metav1.LabelSelectorOpExists,
							},
						},
					},
				},
			},
			Upgraded: api.UpgradedConfig{
				FleetlockURL: "https://fleetlock.example.com",
			},
		},
	}

	warn, err = (&planValidatingHook{}).ValidateCreate(ctx, plan)

	assert.Nil(warn, "Should not return a warning")
	assert.NoError(err, "Should not return an error")
}

func TestValidateUpdate(t *testing.T) {
	assert := assert.New(t)

	ctx := t.Context()

	warn, err := (&planValidatingHook{}).ValidateUpdate(ctx, &api.KubeUpgradePlan{}, &api.KubeUpgradePlan{})

	assert.Nil(warn, "Should not return a warning")
	assert.Error(err, "Should return an error")

	plan := &api.KubeUpgradePlan{
		Spec: api.KubeUpgradeSpec{
			KubernetesVersion: "v1.31.0",
			Groups: map[string]api.KubeUpgradePlanGroup{
				"control-plane": {
					Labels: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "node-role.kubernetes.io/control-plane",
								Operator: metav1.LabelSelectorOpExists,
							},
						},
					},
				},
			},
			Upgraded: api.UpgradedConfig{
				FleetlockURL: "https://fleetlock.example.com",
			},
		},
	}

	warn, err = (&planValidatingHook{}).ValidateUpdate(ctx, &api.KubeUpgradePlan{}, plan)

	assert.Nil(warn, "Should not return a warning")
	assert.NoError(err, "Should not return an error")
}

func TestValidateDelete(t *testing.T) {
	assert := assert.New(t)

	warn, err := (&planValidatingHook{}).ValidateDelete(t.Context(), nil)

	assert.Nil(warn, "Should not return a warning")
	assert.NoError(err, "Should not return an error")
}
