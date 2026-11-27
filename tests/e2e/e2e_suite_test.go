package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

const targetK8sVersion = "v1.37.3"

func TestE2E(t *testing.T) {
	upgradeFeat := features.New("upgrade cluster").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}

			err = decoder.ApplyWithManifestDir(ctx, r, "tests/e2e/testdata", "upgrade-cr.yaml", []resources.CreateOption{})
			if err != nil {
				t.Fatalf("Failed to apply upgrade-plan manifest: %v", err)
			}

			return ctx
		}).
		Assess("status", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}

			err = api.AddToScheme(r.GetScheme())
			if err != nil {
				t.Fatalf("Failed to add CRD to scheme: %v", err)
			}

			plan := &api.KubeUpgradePlan{}

			err = r.Get(ctx, "upgrade-plan", "", plan)
			if err != nil {
				t.Fatalf("Failed to fetch upgrade-plan: %v", err)
			}

			err = wait.For(conditions.New(r).ResourceMatch(plan, func(obj k8s.Object) bool {
				plan, ok := obj.(*api.KubeUpgradePlan)
				if !ok {
					return false
				}
				return strings.HasPrefix(plan.Status.Summary, api.PlanStatusComplete)
			}), wait.WithTimeout(time.Minute), wait.WithInterval(time.Second))
			if err != nil {
				t.Fatalf("Plan status summary is expected to be %s but is %s: %v", api.PlanStatusComplete, plan.Status.Summary, err)
			}

			assert := assert.New(t)

			assert.Equal(len(plan.Spec.Groups), len(plan.Status.Groups), "Should have a status for each group")

			return ctx
		}).
		Assess("daemonsets", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			assert := assert.New(t)

			daemons := &appsv1.DaemonSetList{}
			err := c.Client().Resources().List(ctx, daemons, resources.WithLabelSelector(fmt.Sprintf("%s=%s", constants.LabelPlanName, "upgrade-plan")))
			if err != nil || daemons.Items == nil {
				t.Fatalf("Error while getting daemonsets: %v", err)
			}

			assert.Len(daemons.Items, 1, "Should have daemonset")

			for _, ds := range daemons.Items {
				assert.Contains(ds.Spec.Template.Spec.Containers[0].Image, "localhost/kube-upgraded:", "Daemonset should use upgraded image")
			}
			return ctx
		}).
		Assess("configmaps", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			assert := assert.New(t)

			cmList := &corev1.ConfigMapList{}
			err := c.Client().Resources().List(ctx, cmList, resources.WithLabelSelector(fmt.Sprintf("%s=%s", constants.LabelPlanName, "upgrade-plan")))
			if err != nil || cmList.Items == nil {
				t.Fatalf("Error while getting configmaps: %v", err)
			}

			assert.Len(cmList.Items, 1, "Should have configmap")
			return ctx
		}).
		Assess("upgrade", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}

			err = api.AddToScheme(r.GetScheme())
			if err != nil {
				t.Fatalf("Failed to add CRD to scheme: %v", err)
			}

			plan := &api.KubeUpgradePlan{}
			err = wait.For(func(ctx context.Context) (bool, error) {
				err := r.Get(ctx, "upgrade-plan", "", plan)
				if err != nil {
					return false, fmt.Errorf("failed to fetch upgrade-plan: %w", err)
				}

				plan.Spec.KubernetesVersion = targetK8sVersion
				err = r.Update(ctx, plan)
				if err == nil {
					return true, nil
				}
				if errors.IsConflict(err) || errors.IsServerTimeout(err) || errors.IsInternalError(err) {
					return false, nil
				}
				return false, err
			}, wait.WithContext(ctx), wait.WithTimeout(time.Minute), wait.WithInterval(time.Second))
			if err != nil {
				t.Fatalf("Failed to update plan: %v", err)
			}

			err = wait.For(conditions.New(r).ResourceMatch(plan, func(obj k8s.Object) bool {
				plan, ok := obj.(*api.KubeUpgradePlan)
				if !ok {
					return false
				}
				return strings.HasPrefix(plan.Status.Summary, api.PlanStatusProgressing)
			}), wait.WithTimeout(time.Minute), wait.WithInterval(time.Second))
			if err != nil {
				t.Fatalf("Plan is expected to be progressing, but is %s: %v", plan.Status.Summary, err)
			}

			err = wait.For(conditions.New(r).ResourceMatch(plan, func(obj k8s.Object) bool {
				plan, ok := obj.(*api.KubeUpgradePlan)
				if !ok {
					return false
				}
				return strings.HasPrefix(plan.Status.Summary, api.PlanStatusComplete)
			}), wait.WithTimeout(10*time.Minute), wait.WithInterval(10*time.Second))
			if err != nil {
				t.Fatalf("Upgrade is expected to complete, but is %s: %v", plan.Status.Summary, err)
			}

			assert := assert.New(t)

			nodes := &corev1.NodeList{}
			assert.NoError(c.Client().Resources().List(ctx, nodes), "Should list nodes without error")
			for _, node := range nodes.Items {
				assert.Equal(targetK8sVersion, node.Status.NodeInfo.KubeletVersion, "Should have new kubelet version")
			}

			return ctx
		}).
		Assess("delete", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			assert := assert.New(t)

			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}

			err = decoder.DeleteWithManifestDir(ctx, r, "tests/e2e/testdata", "upgrade-cr.yaml", []resources.DeleteOption{})
			if err != nil {
				t.Fatalf("Failed to delete upgrade-plan: %v", err)
			}

			assert.Eventually(func() bool {
				plan := &api.KubeUpgradePlan{}
				err := r.Get(ctx, "upgrade-plan", "", plan)
				return err != nil && client.IgnoreNotFound(err) == nil
			}, 1*time.Minute, 5*time.Second, "Plan should be deleted")

			cmList := &corev1.ConfigMapList{}
			assert.Eventually(func() bool {
				err = c.Client().Resources().List(ctx, cmList, resources.WithLabelSelector(fmt.Sprintf("%s=%s", constants.LabelPlanName, "upgrade-plan")))
				if err != nil {
					t.Fatalf("Error while getting configmaps: %v", err)
				}
				return len(cmList.Items) == 0
			}, 1*time.Minute, 5*time.Second, "Should have deleted ConfigMaps")

			dsList := &appsv1.DaemonSetList{}
			assert.Eventually(func() bool {
				err = c.Client().Resources().List(ctx, dsList, resources.WithLabelSelector(fmt.Sprintf("%s=%s", constants.LabelPlanName, "upgrade-plan")))
				if err != nil {
					t.Fatalf("Error while getting daemonsets: %v", err)
				}
				return len(dsList.Items) == 0
			}, 1*time.Minute, 5*time.Second, "Should have deleted DaemonSets")

			return ctx
		}).Feature()

	testenv.Test(t, upgradeFeat)
}
