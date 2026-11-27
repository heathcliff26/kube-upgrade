package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var minimalPlan = &api.KubeUpgradePlan{
	Name: "minimal-plan",
	Spec: api.KubeUpgradeSpec{
		KubernetesVersion: "v1.37.0",
		Upgraded: api.UpgradedConfig{
			FleetlockURL: "https://fleetlock.example.com",
		},
		Groups: map[string]api.KubeUpgradePlanGroup{
			groupControl: {
				Labels: map[string]string{
					labelControl: "true",
				},
				Upgraded: &api.UpgradedConfig{
					FleetlockGroup: groupControl,
				},
			},
		},
	},
}

const (
	namespace = "kube-upgrade"

	groupControl    = "control-plane"
	labelControl    = "node-role.kubernetes.io/control-plane"
	nodeControlName = "node-control"

	groupCompute    = "compute"
	labelCompute    = "node-role.kubernetes.io/compute"
	nodeComputeName = "node-compute"

	reconcileTimeout = time.Second * 5
	reconcileTick    = time.Millisecond * 100
)

func TestExamplePlan(t *testing.T) {
	require := require.New(t)

	f, err := os.ReadFile(filepath.Join(repoRoot, "examples", "upgrade-cr.yaml"))
	require.NoError(err, "Should read plan")

	var plan *api.KubeUpgradePlan
	err = yaml.Unmarshal(f, &plan)
	require.NoError(err, "Should parse plan")

	err = k8sClient.Create(t.Context(), plan)
	require.NoError(err, "Should create plan")
	t.Cleanup(func() {
		err := k8sClient.Delete(context.Background(), plan)
		require.NoError(err, "Should delete plan")
	})

	require.Eventually(func() bool {
		err := k8sClient.Get(t.Context(), client.ObjectKeyFromObject(plan), plan)
		require.NoError(err, "Should retrieve plan")
		return strings.HasPrefix(plan.Status.Summary, api.PlanStatusProgressing)
	}, reconcileTimeout, reconcileTick, "Should set plan status")

	require.Equal(len(plan.Spec.Groups), len(plan.Status.Groups), "Should have a status for each group")
	require.Equal(api.PlanStatusProgressing+": 0/1 nodes upgraded", plan.Status.Groups[groupControl], "control-plane group should be progressing")
	require.Equal(api.PlanStatusWaiting, plan.Status.Groups[groupCompute], "compute group should be waiting")

	node := &corev1.Node{Name: nodeControlName}
	err = k8sClient.Get(t.Context(), client.ObjectKeyFromObject(node), node)
	require.NoError(err, "Should get node")
	require.Equal(constants.NodeUpgradeStatusPending, node.GetAnnotations()[constants.NodeUpgradeStatus], "Should have annotated node")

	daemons := &appsv1.DaemonSetList{}
	err = k8sClient.List(t.Context(), daemons, client.MatchingLabels{constants.LabelPlanName: plan.GetName()})
	require.NoError(err, "Should list daemonsets")
	require.Len(daemons.Items, 2, "Should have 2 daemonsets")

	for _, ds := range daemons.Items {
		require.Equal(ds.Spec.Template.Spec.Containers[0].Image, "localhost/kube-upgraded:envtest", "Daemonset should use upgraded image")
	}

	cmList := &corev1.ConfigMapList{}
	err = k8sClient.List(t.Context(), cmList, client.MatchingLabels{constants.LabelPlanName: plan.GetName()})
	require.NoError(err, "Should list configmaps")
	require.Len(cmList.Items, 2, "Should have 2 configmaps")
}

func TestInvalidPlan(t *testing.T) {
	require := require.New(t)

	plan := &api.KubeUpgradePlan{
		Name: "invalid-plan",
		Spec: api.KubeUpgradeSpec{
			KubernetesVersion: "v1.37.0",
			Groups: map[string]api.KubeUpgradePlanGroup{
				groupControl: {
					Labels: map[string]string{
						labelControl: "true",
					},
					Upgraded: &api.UpgradedConfig{
						FleetlockGroup: groupControl,
					},
				},
			},
		},
	}
	err := k8sClient.Create(t.Context(), plan)
	require.ErrorContains(err, "admission webhook \"kubeupgrade.heathcliff.eu\" denied the request", "Plan should be rejected by webhook")
}

func TestMutatingWebhook(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	plan := minimalPlan.DeepCopy()
	err := k8sClient.Create(t.Context(), plan)
	require.NoError(err, "Should create plan")
	t.Cleanup(func() {
		err := k8sClient.Delete(context.Background(), plan)
		require.NoError(err, "Should delete plan")
	})

	assert.Equal(api.DefaultUpgradedStream, plan.Spec.Upgraded.Stream, "Should set default upgraded stream")
	assert.Equal(api.DefaultUpgradedFleetlockGroup, plan.Spec.Upgraded.FleetlockGroup, "Should set default upgraded fleetlockGroup")
	assert.Equal(api.DefaultUpgradedCheckInterval, plan.Spec.Upgraded.CheckInterval, "Should set default upgraded checkInterval")
	assert.Equal(api.DefaultUpgradedRetryInterval, plan.Spec.Upgraded.RetryInterval, "Should set default upgraded retryInterval")

	assert.Empty(plan.Spec.Groups["control-plane"].Upgraded.Stream, "Should leave group upgraded stream empty")
	assert.Equal("control-plane", plan.Spec.Groups["control-plane"].Upgraded.FleetlockGroup, "Should set fleetlock group for control-plane")
	assert.Empty(plan.Spec.Groups["control-plane"].Upgraded.CheckInterval, "Should leave group upgraded checkInterval empty")
	assert.Empty(plan.Spec.Groups["control-plane"].Upgraded.RetryInterval, "Should leave group upgraded retryInterval empty")
}

func TestMultiplePlans(t *testing.T) {
	require := require.New(t)

	plan := minimalPlan.DeepCopy()
	err := k8sClient.Create(t.Context(), plan)
	require.NoError(err, "Should create plan")
	t.Cleanup(func() {
		err := k8sClient.Delete(context.Background(), plan)
		require.NoError(err, "Should delete plan")
	})

	p2 := minimalPlan.DeepCopy()
	err = k8sClient.Create(t.Context(), p2)
	require.ErrorContains(err, "admission webhook \"kubeupgrade.heathcliff.eu\" denied the request: KubeUpgradePlan already exists", "Should reject 2nd plan")
}
