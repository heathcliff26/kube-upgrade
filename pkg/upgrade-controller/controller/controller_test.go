package controller

import (
	"context"
	"testing"
	"time"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha1"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeFake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"
)

const (
	groupControl = "control"
	nodeControl  = "node-control"
	groupCompute = "compute"
	nodeCompute  = "node-compute"
	groupInfra   = "infra"
	nodeInfra    = "node-infra"
	groupLabel   = "group"
)

func TestNewController(t *testing.T) {
	c, err := NewController("test")

	assert := assert.New(t)

	assert.Nil(c, "Should not return a client")
	assert.Error(err, "Client creation should fail")
}

func TestReconcile(t *testing.T) {
	tMatrix := []struct {
		Name                                                                             string
		Plan                                                                             api.KubeUpgradePlan
		AnnotationsControl, AnnotationsCompute, AnnotationsInfra                         map[string]string
		ExpectedSummary                                                                  string
		ExpectedGroupStatus                                                              map[string]string
		ExpectedAnnotationsControl, ExpectedAnnotationsCompute, ExpectedAnnotationsInfra map[string]string
	}{
		{
			Name: "InitialReconcile",
			Plan: api.KubeUpgradePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "upgrade-plan",
				},
				Spec: api.KubeUpgradeSpec{
					KubernetesVersion: "v1.31.0",
					Groups: map[string]api.KubeUpgradePlanGroup{
						"control": {
							Labels: map[string]string{
								groupLabel: groupControl,
							},
						},
						"compute": {
							DependsOn: []string{"control", "infra"},
							Labels: map[string]string{
								groupLabel: groupCompute,
							},
						},
						"infra": {
							DependsOn: []string{"control"},
							Labels: map[string]string{
								groupLabel: groupInfra,
							},
						},
					},
				},
			},
			ExpectedSummary: api.PlanStatusProgressing,
			ExpectedGroupStatus: map[string]string{
				groupControl: api.PlanStatusProgressing,
				groupCompute: api.PlanStatusWaiting,
				groupInfra:   api.PlanStatusWaiting,
			},
			ExpectedAnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusPending,
			},
		},
		{
			Name: "2ndReconcile",
			Plan: api.KubeUpgradePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "upgrade-plan",
				},
				Spec: api.KubeUpgradeSpec{
					KubernetesVersion: "v1.31.0",
					Groups: map[string]api.KubeUpgradePlanGroup{
						"control": {
							Labels: map[string]string{
								groupLabel: groupControl,
							},
						},
						"compute": {
							DependsOn: []string{"control", "infra"},
							Labels: map[string]string{
								groupLabel: groupCompute,
							},
						},
						"infra": {
							DependsOn: []string{"control"},
							Labels: map[string]string{
								groupLabel: groupInfra,
							},
						},
					},
				},
				Status: api.KubeUpgradeStatus{
					Summary: api.PlanStatusProgressing,
					Groups: map[string]string{
						groupControl: api.PlanStatusProgressing,
						groupCompute: api.PlanStatusWaiting,
						groupInfra:   api.PlanStatusWaiting,
					},
				},
			},
			AnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			ExpectedSummary: api.PlanStatusProgressing,
			ExpectedGroupStatus: map[string]string{
				groupControl: api.PlanStatusComplete,
				groupCompute: api.PlanStatusWaiting,
				groupInfra:   api.PlanStatusProgressing,
			},
			ExpectedAnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			ExpectedAnnotationsInfra: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusPending,
			},
		},
		{
			Name: "3rdReconcile",
			Plan: api.KubeUpgradePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "upgrade-plan",
				},
				Spec: api.KubeUpgradeSpec{
					KubernetesVersion: "v1.31.0",
					Groups: map[string]api.KubeUpgradePlanGroup{
						"control": {
							Labels: map[string]string{
								groupLabel: groupControl,
							},
						},
						"compute": {
							DependsOn: []string{"control", "infra"},
							Labels: map[string]string{
								groupLabel: groupCompute,
							},
						},
						"infra": {
							DependsOn: []string{"control"},
							Labels: map[string]string{
								groupLabel: groupInfra,
							},
						},
					},
				},
				Status: api.KubeUpgradeStatus{
					Summary: api.PlanStatusProgressing,
					Groups: map[string]string{
						groupControl: api.PlanStatusComplete,
						groupCompute: api.PlanStatusWaiting,
						groupInfra:   api.PlanStatusProgressing,
					},
				},
			},
			AnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			AnnotationsInfra: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			ExpectedSummary: api.PlanStatusProgressing,
			ExpectedGroupStatus: map[string]string{
				groupControl: api.PlanStatusComplete,
				groupCompute: api.PlanStatusProgressing,
				groupInfra:   api.PlanStatusComplete,
			},
			ExpectedAnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			ExpectedAnnotationsInfra: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			ExpectedAnnotationsCompute: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusPending,
			},
		},
		{
			Name: "4thReconcile",
			Plan: api.KubeUpgradePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "upgrade-plan",
				},
				Spec: api.KubeUpgradeSpec{
					KubernetesVersion: "v1.31.0",
					Groups: map[string]api.KubeUpgradePlanGroup{
						"control": {
							Labels: map[string]string{
								groupLabel: groupControl,
							},
						},
						"compute": {
							DependsOn: []string{"control", "infra"},
							Labels: map[string]string{
								groupLabel: groupCompute,
							},
						},
						"infra": {
							DependsOn: []string{"control"},
							Labels: map[string]string{
								groupLabel: groupInfra,
							},
						},
					},
				},
				Status: api.KubeUpgradeStatus{
					Summary: api.PlanStatusProgressing,
					Groups: map[string]string{
						groupControl: api.PlanStatusComplete,
						groupCompute: api.PlanStatusProgressing,
						groupInfra:   api.PlanStatusComplete,
					},
				},
			},
			AnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			AnnotationsInfra: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			AnnotationsCompute: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			ExpectedSummary: api.PlanStatusComplete,
			ExpectedGroupStatus: map[string]string{
				groupControl: api.PlanStatusComplete,
				groupCompute: api.PlanStatusComplete,
				groupInfra:   api.PlanStatusComplete,
			},
			ExpectedAnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			ExpectedAnnotationsInfra: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			ExpectedAnnotationsCompute: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
		},
		{
			Name: "NewUpdate",
			Plan: api.KubeUpgradePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "upgrade-plan",
				},
				Spec: api.KubeUpgradeSpec{
					KubernetesVersion: "v1.31.0",
					Groups: map[string]api.KubeUpgradePlanGroup{
						"control": {
							Labels: map[string]string{
								groupLabel: groupControl,
							},
						},
						"compute": {
							DependsOn: []string{"control"},
							Labels: map[string]string{
								groupLabel: groupCompute,
							},
						},
					},
				},
				Status: api.KubeUpgradeStatus{
					Summary: api.PlanStatusComplete,
					Groups: map[string]string{
						groupControl: api.PlanStatusComplete,
						groupCompute: api.PlanStatusComplete,
					},
				},
			},
			AnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.30.4",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			AnnotationsCompute: map[string]string{
				constants.NodeKubernetesVersion: "v1.30.4",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			ExpectedSummary: api.PlanStatusProgressing,
			ExpectedGroupStatus: map[string]string{
				groupControl: api.PlanStatusProgressing,
				groupCompute: api.PlanStatusWaiting,
			},
			ExpectedAnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusPending,
			},
			ExpectedAnnotationsCompute: map[string]string{
				constants.NodeKubernetesVersion: "v1.30.4",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
		},
		{
			Name: "Update2ndReconcile",
			Plan: api.KubeUpgradePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "upgrade-plan",
				},
				Spec: api.KubeUpgradeSpec{
					KubernetesVersion: "v1.31.0",
					Groups: map[string]api.KubeUpgradePlanGroup{
						"control": {
							Labels: map[string]string{
								groupLabel: groupControl,
							},
						},
						"compute": {
							DependsOn: []string{"control"},
							Labels: map[string]string{
								groupLabel: groupCompute,
							},
						},
					},
				},
				Status: api.KubeUpgradeStatus{
					Summary: api.PlanStatusProgressing,
					Groups: map[string]string{
						groupControl: api.PlanStatusProgressing,
						groupCompute: api.PlanStatusWaiting,
					},
				},
			},
			AnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			AnnotationsCompute: map[string]string{
				constants.NodeKubernetesVersion: "v1.30.4",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			ExpectedSummary: api.PlanStatusProgressing,
			ExpectedGroupStatus: map[string]string{
				groupControl: api.PlanStatusComplete,
				groupCompute: api.PlanStatusProgressing,
			},
			ExpectedAnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			ExpectedAnnotationsCompute: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusPending,
			},
		},
		{
			Name: "Update3rdReconcile",
			Plan: api.KubeUpgradePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "upgrade-plan",
				},
				Spec: api.KubeUpgradeSpec{
					KubernetesVersion: "v1.31.0",
					Groups: map[string]api.KubeUpgradePlanGroup{
						"control": {
							Labels: map[string]string{
								groupLabel: groupControl,
							},
						},
						"compute": {
							DependsOn: []string{"control"},
							Labels: map[string]string{
								groupLabel: groupCompute,
							},
						},
					},
				},
				Status: api.KubeUpgradeStatus{
					Summary: api.PlanStatusProgressing,
					Groups: map[string]string{
						groupControl: api.PlanStatusComplete,
						groupCompute: api.PlanStatusProgressing,
					},
				},
			},
			AnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			AnnotationsCompute: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			ExpectedSummary: api.PlanStatusComplete,
			ExpectedGroupStatus: map[string]string{
				groupControl: api.PlanStatusComplete,
				groupCompute: api.PlanStatusComplete,
			},
			ExpectedAnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			ExpectedAnnotationsCompute: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
		},
		{
			Name: "ApplyConfiguration",
			Plan: api.KubeUpgradePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "upgrade-plan",
				},
				Spec: api.KubeUpgradeSpec{
					KubernetesVersion: "v1.30.4",
					Upgraded: &api.UpgradedConfig{
						Stream:         "registry.example.com/test-stream",
						FleetlockURL:   "https://fleetlock.example.org",
						FleetlockGroup: "default",
						CheckInterval:  time.Minute * 2,
						RetryInterval:  time.Minute * 3,
					},
					Groups: map[string]api.KubeUpgradePlanGroup{
						"control": {
							Labels: map[string]string{
								groupLabel: groupControl,
							},
							Upgraded: &api.UpgradedConfig{
								FleetlockGroup: "control-plane",
							},
						},
						"compute": {
							DependsOn: []string{"control"},
							Labels: map[string]string{
								groupLabel: groupCompute,
							},
							Upgraded: &api.UpgradedConfig{
								FleetlockGroup: "compute",
							},
						},
					},
				},
				Status: api.KubeUpgradeStatus{
					Summary: api.PlanStatusComplete,
					Groups: map[string]string{
						groupControl: api.PlanStatusComplete,
						groupCompute: api.PlanStatusComplete,
					},
				},
			},
			AnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.30.4",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
				constants.ConfigFleetlockGroup:  "not-the-group",
			},
			AnnotationsCompute: map[string]string{
				constants.NodeKubernetesVersion: "v1.30.4",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
				constants.ConfigFleetlockGroup:  "not-the-group-either",
			},
			ExpectedSummary: api.PlanStatusComplete,
			ExpectedGroupStatus: map[string]string{
				groupControl: api.PlanStatusComplete,
				groupCompute: api.PlanStatusComplete,
			},
			ExpectedAnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.30.4",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
				constants.ConfigStream:          "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:    "https://fleetlock.example.org",
				constants.ConfigFleetlockGroup:  "control-plane",
				constants.ConfigCheckInterval:   "2m0s",
				constants.ConfigRetryInterval:   "3m0s",
			},
			ExpectedAnnotationsCompute: map[string]string{
				constants.NodeKubernetesVersion: "v1.30.4",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
				constants.ConfigStream:          "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:    "https://fleetlock.example.org",
				constants.ConfigFleetlockGroup:  "compute",
				constants.ConfigCheckInterval:   "2m0s",
				constants.ConfigRetryInterval:   "3m0s",
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			c := createFakeController(tCase.AnnotationsControl, tCase.AnnotationsCompute, tCase.AnnotationsInfra)

			assert := assert.New(t)

			err := c.reconcile(context.Background(), &tCase.Plan, klog.NewKlogr())

			if !assert.NoError(err, "Reconcile should succeed") {
				t.FailNow()
			}

			assert.Equal(tCase.ExpectedSummary, tCase.Plan.Status.Summary, "Summary should be correct")

			if !assert.Equal(len(tCase.Plan.Spec.Groups), len(tCase.Plan.Status.Groups), "Group lengths should match") {
				t.FailNow()
			}

			assert.Equal(tCase.ExpectedGroupStatus, tCase.Plan.Status.Groups, "Group status should match")

			nodeControl, _ := c.nodes.Get(context.Background(), nodeControl, metav1.GetOptions{})
			nodeCompute, _ := c.nodes.Get(context.Background(), nodeCompute, metav1.GetOptions{})
			nodeInfra, _ := c.nodes.Get(context.Background(), nodeInfra, metav1.GetOptions{})

			assert.Equal(tCase.ExpectedAnnotationsControl, nodeControl.GetAnnotations(), "Control group should have expected annotations")
			assert.Equal(tCase.ExpectedAnnotationsCompute, nodeCompute.GetAnnotations(), "Compute group should have expected annotations")
			assert.Equal(tCase.ExpectedAnnotationsInfra, nodeInfra.GetAnnotations(), "Infra group should have expected annotations")
		})
	}
}

func createFakeController(annotationsControl, annotationsCompute, annotationsInfra map[string]string) *controller {
	c := &controller{
		nodes: kubeFake.NewSimpleClientset().CoreV1().Nodes(),
	}

	nodeControl := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeControl,
			Labels: map[string]string{
				groupLabel: groupControl,
			},
			Annotations: annotationsControl,
		},
	}
	nodeCompute := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeCompute,
			Labels: map[string]string{
				groupLabel: groupCompute,
			},
			Annotations: annotationsCompute,
		},
	}
	nodeInfra := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeInfra,
			Labels: map[string]string{
				groupLabel: groupInfra,
			},
			Annotations: annotationsInfra,
		},
	}
	_, _ = c.nodes.Create(context.Background(), nodeControl, metav1.CreateOptions{})
	_, _ = c.nodes.Create(context.Background(), nodeCompute, metav1.CreateOptions{})
	_, _ = c.nodes.Create(context.Background(), nodeInfra, metav1.CreateOptions{})

	return c
}
