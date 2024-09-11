package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha2"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestE2E(t *testing.T) {
	certManagerFeat := features.New("deploy cert-manager").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			res, err := http.Get(fmt.Sprintf("https://github.com/cert-manager/cert-manager/releases/download/%s/cert-manager.yaml", certManagerVersion))
			if err != nil || res.StatusCode != http.StatusOK {
				t.Fatalf("Failed to fetch cert-manager manifests for %s: %v", certManagerVersion, err)
			}

			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}

			err = decoder.DecodeEach(ctx, res.Body, decoder.CreateHandler(r))
			if err != nil {
				t.Fatalf("Failed to deploy cert-manager: %v", err)
			}

			return ctx
		}).
		Assess("available", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			err := wait.For(conditions.New(c.Client().Resources()).DeploymentAvailable("cert-manager-webhook", "cert-manager"), wait.WithTimeout(5*time.Minute), wait.WithInterval(10*time.Second))
			if err != nil {
				t.Fatalf("Failed to wait for cert-manager-webhook deployment: %v", err)
			}

			return ctx
		}).
		Feature()

	controllerDeploymentFeat := features.New("deploy upgrade-controller").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}

			err = decoder.ApplyWithManifestDir(ctx, r, "manifests/release", "upgrade-controller.yaml", []resources.CreateOption{})
			if err != nil {
				t.Fatalf("Failed to apply upgrade-controller manifest: %v", err)
			}

			return ctx
		}).
		Assess("available", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var dep appsv1.Deployment
			if err := c.Client().Resources().Get(ctx, "upgrade-controller", namespace, &dep); err != nil {
				t.Fatalf("Failed to get upgrade-controller-deployment: %v", err)
			}

			err := wait.For(conditions.New(c.Client().Resources()).DeploymentConditionMatch(&dep, appsv1.DeploymentAvailable, corev1.ConditionTrue), wait.WithTimeout(time.Minute*1))
			if err != nil {
				t.Fatalf("Failed to wait for upgrade-controller deployment to be created: %v", err)
			}

			pods := &corev1.PodList{}
			err = c.Client().Resources(namespace).List(ctx, pods)
			if err != nil || pods.Items == nil {
				t.Fatalf("Error while getting pods: %v", err)
			}
			if len(pods.Items) != 2 {
				t.Fatalf("Not enough upgrade-controller pods, expected 2 but got %d", len(pods.Items))
			}
			for _, pod := range pods.Items {
				err = wait.For(conditions.New(c.Client().Resources()).PodConditionMatch(&pod, corev1.PodReady, corev1.ConditionTrue), wait.WithTimeout(time.Minute*1))
				if err != nil {
					t.Fatalf("Failed to wait for pod %s to be ready: %v", pod.GetName(), err)
				}
			}

			return ctx
		}).
		Feature()

	testenv.Test(t, certManagerFeat, controllerDeploymentFeat)
	if t.Failed() {
		t.Fatal("Failed to deploy controller, can't run tests")
	}

	examplePlanFeat := features.New("create example plan").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}

			err = decoder.ApplyWithManifestDir(ctx, r, "manifests/release", "upgrade-cr.yaml", []resources.CreateOption{})
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
				return plan.Status.Summary == api.PlanStatusUnknown
			}), wait.WithTimeout(1*time.Minute), wait.WithInterval(1*time.Second))
			if err != nil {
				t.Fatalf("Plan status summary is expected to be %s but is %s: %v", api.PlanStatusUnknown, plan.Status.Summary, err)
			}

			assert := assert.New(t)

			assert.Equal(len(plan.Spec.Groups), len(plan.Status.Groups), "Should have a status for each group")
			assert.Equal(api.PlanStatusProgressing, plan.Status.Groups["control-plane"], "control-plane group should be progressing")
			assert.Equal(api.PlanStatusUnknown, plan.Status.Groups["compute"], "compute group should be unknown")

			return ctx
		}).
		Assess("nodes", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			nodes := &corev1.NodeList{}
			err := c.Client().Resources().List(ctx, nodes)
			if err != nil || nodes.Items == nil {
				t.Fatalf("Error while getting nodes: %v", err)
			}

			for _, node := range nodes.Items {
				if _, ok := node.GetAnnotations()["node-role.kubernetes.io/control-plane"]; ok {
					value := node.GetAnnotations()[constants.NodeUpgradeStatus]
					assert.Equalf(t, constants.NodeUpgradeStatusPending, value, "control-plane node %s should have upgrade status annotation")
				}
			}

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}

			err = decoder.DeleteWithManifestDir(ctx, r, "manifests/release", "upgrade-cr.yaml", []resources.DeleteOption{})
			if err != nil {
				t.Fatalf("Failed to delete upgrade-plan: %v", err)
			}

			return ctx
		}).Feature()

	validationWebhookFeat := features.New("validation webhook").
		Assess("rejected", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}

			err = decoder.ApplyWithManifestDir(ctx, r, "tests/testdata", "invalid-plan.yaml", []resources.CreateOption{})
			assert.ErrorContains(t, err, "admission webhook \"kubeupgrade.heathcliff.eu\" denied the request", "Plan should be rejected by webhook")

			return ctx
		}).Feature()

	testenv.Test(t, examplePlanFeat, validationWebhookFeat)
}
