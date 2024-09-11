package controller

import (
	"context"
	"testing"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha2"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeFake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"
)

const (
	groupControl = "control"
	labelControl = "node-role.kubernetes.io/control-plane"
	nodeControl  = "node-control"

	groupCompute = "compute"
	labelCompute = "node-role.kubernetes.io/compute"
	nodeCompute  = "node-compute"

	groupInfra = "infra"
	labelInfra = "node-role.kubernetes.io/infra"
	nodeInfra  = "node-infra"

	labelValue = "true"
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
						groupControl: {
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelControl: labelValue,
								},
							},
						},
						groupCompute: {
							DependsOn: []string{groupControl, groupInfra},
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelCompute: labelValue,
								},
							},
						},
						groupInfra: {
							DependsOn: []string{groupControl},
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelInfra: labelValue,
								},
							},
						},
					},
				},
			},
			ExpectedSummary: api.PlanStatusProgressing + ": Upgrading groups [control]",
			ExpectedGroupStatus: map[string]string{
				groupControl: api.PlanStatusProgressing + ": 0/1 nodes upgraded",
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
						groupControl: {
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelControl: labelValue,
								},
							},
						},
						groupCompute: {
							DependsOn: []string{groupControl, groupInfra},
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelCompute: labelValue,
								},
							},
						},
						groupInfra: {
							DependsOn: []string{groupControl},
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelInfra: labelValue,
								},
							},
						},
					},
				},
				Status: api.KubeUpgradeStatus{
					Summary: api.PlanStatusProgressing + ": Upgrading groups [control]",
					Groups: map[string]string{
						groupControl: api.PlanStatusProgressing + ": 0/1 nodes upgraded",
						groupCompute: api.PlanStatusWaiting,
						groupInfra:   api.PlanStatusWaiting,
					},
				},
			},
			AnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			ExpectedSummary: api.PlanStatusProgressing + ": Upgrading groups [infra]",
			ExpectedGroupStatus: map[string]string{
				groupControl: api.PlanStatusComplete,
				groupCompute: api.PlanStatusWaiting,
				groupInfra:   api.PlanStatusProgressing + ": 0/1 nodes upgraded",
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
						groupControl: {
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelControl: labelValue,
								},
							},
						},
						groupCompute: {
							DependsOn: []string{groupControl, groupInfra},
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelCompute: labelValue,
								},
							},
						},
						groupInfra: {
							DependsOn: []string{groupControl},
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelInfra: labelValue,
								},
							},
						},
					},
				},
				Status: api.KubeUpgradeStatus{
					Summary: api.PlanStatusProgressing + ": Upgrading groups [infra]",
					Groups: map[string]string{
						groupControl: api.PlanStatusComplete,
						groupCompute: api.PlanStatusWaiting,
						groupInfra:   api.PlanStatusProgressing + ": 0/1 nodes upgraded",
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
			ExpectedSummary: api.PlanStatusProgressing + ": Upgrading groups [compute]",
			ExpectedGroupStatus: map[string]string{
				groupControl: api.PlanStatusComplete,
				groupCompute: api.PlanStatusProgressing + ": 0/1 nodes upgraded",
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
						groupControl: {
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelControl: labelValue,
								},
							},
						},
						groupCompute: {
							DependsOn: []string{groupControl, groupInfra},
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelCompute: labelValue,
								},
							},
						},
						groupInfra: {
							DependsOn: []string{groupControl},
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelInfra: labelValue,
								},
							},
						},
					},
				},
				Status: api.KubeUpgradeStatus{
					Summary: api.PlanStatusProgressing + ": Upgrading groups [compute]",
					Groups: map[string]string{
						groupControl: api.PlanStatusComplete,
						groupCompute: api.PlanStatusProgressing + ": 0/1 nodes upgraded",
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
						groupControl: {
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelControl: labelValue,
								},
							},
						},
						groupCompute: {
							DependsOn: []string{groupControl},
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelCompute: labelValue,
								},
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
			ExpectedSummary: api.PlanStatusProgressing + ": Upgrading groups [control]",
			ExpectedGroupStatus: map[string]string{
				groupControl: api.PlanStatusProgressing + ": 0/1 nodes upgraded",
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
						groupControl: {
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelControl: labelValue,
								},
							},
						},
						groupCompute: {
							DependsOn: []string{groupControl},
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelCompute: labelValue,
								},
							},
						},
					},
				},
				Status: api.KubeUpgradeStatus{
					Summary: api.PlanStatusProgressing + ": Upgrading groups [control]",
					Groups: map[string]string{
						groupControl: api.PlanStatusProgressing + ": 0/1 nodes upgraded",
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
			ExpectedSummary: api.PlanStatusProgressing + ": Upgrading groups [compute]",
			ExpectedGroupStatus: map[string]string{
				groupControl: api.PlanStatusComplete,
				groupCompute: api.PlanStatusProgressing + ": 0/1 nodes upgraded",
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
						groupControl: {
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelControl: labelValue,
								},
							},
						},
						groupCompute: {
							DependsOn: []string{groupControl},
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelCompute: labelValue,
								},
							},
						},
					},
				},
				Status: api.KubeUpgradeStatus{
					Summary: api.PlanStatusProgressing + ": Upgrading groups [compute]",
					Groups: map[string]string{
						groupControl: api.PlanStatusComplete,
						groupCompute: api.PlanStatusProgressing + ": 0/1 nodes upgraded",
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
					Upgraded: api.UpgradedConfig{
						Stream:         "registry.example.com/test-stream",
						FleetlockURL:   "https://fleetlock.example.org",
						FleetlockGroup: "default",
						CheckInterval:  "2m",
						RetryInterval:  "3m",
					},
					Groups: map[string]api.KubeUpgradePlanGroup{
						groupControl: {
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelControl: labelValue,
								},
							},
							Upgraded: &api.UpgradedConfig{
								FleetlockGroup: "control-plane",
							},
						},
						groupCompute: {
							DependsOn: []string{groupControl},
							Labels: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									labelCompute: labelValue,
								},
							},
							Upgraded: &api.UpgradedConfig{
								FleetlockGroup: groupCompute,
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
				constants.ConfigCheckInterval:   "2m",
				constants.ConfigRetryInterval:   "3m",
			},
			ExpectedAnnotationsCompute: map[string]string{
				constants.NodeKubernetesVersion: "v1.30.4",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
				constants.ConfigStream:          "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:    "https://fleetlock.example.org",
				constants.ConfigFleetlockGroup:  groupCompute,
				constants.ConfigCheckInterval:   "2m",
				constants.ConfigRetryInterval:   "3m",
			},
		},
		{
			Name: "LabelSelectorWithExpression",
			Plan: api.KubeUpgradePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "upgrade-plan",
				},
				Spec: api.KubeUpgradeSpec{
					KubernetesVersion: "v1.31.0",
					Groups: map[string]api.KubeUpgradePlanGroup{
						groupControl: {
							Labels: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      labelControl,
										Operator: metav1.LabelSelectorOpExists,
									},
								},
							},
						},
						groupCompute: {
							DependsOn: []string{groupControl, groupInfra},
							Labels: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      labelCompute,
										Operator: metav1.LabelSelectorOpExists,
									},
								},
							},
						},
						groupInfra: {
							DependsOn: []string{groupControl},
							Labels: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      labelInfra,
										Operator: metav1.LabelSelectorOpExists,
									},
								},
							},
						},
					},
				},
			},
			ExpectedSummary: api.PlanStatusProgressing + ": Upgrading groups [control]",
			ExpectedGroupStatus: map[string]string{
				groupControl: api.PlanStatusProgressing + ": 0/1 nodes upgraded",
				groupCompute: api.PlanStatusWaiting,
				groupInfra:   api.PlanStatusWaiting,
			},
			ExpectedAnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusPending,
			},
		},
		{
			Name: "GroupHasErrorNode",
			Plan: api.KubeUpgradePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "upgrade-plan",
				},
				Spec: api.KubeUpgradeSpec{
					KubernetesVersion: "v1.31.0",
					Groups: map[string]api.KubeUpgradePlanGroup{
						groupControl: {
							Labels: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      labelControl,
										Operator: metav1.LabelSelectorOpExists,
									},
								},
							},
						},
					},
				},
			},
			AnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusError,
			},
			ExpectedSummary: api.PlanStatusError + ": Some groups encountered errors [control]",
			ExpectedGroupStatus: map[string]string{
				groupControl: api.PlanStatusError + ": The nodes [node-control] are reporting errors",
			},
			ExpectedAnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusError,
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
				labelControl: labelValue,
			},
			Annotations: annotationsControl,
		},
	}
	nodeCompute := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeCompute,
			Labels: map[string]string{
				labelCompute: labelValue,
			},
			Annotations: annotationsCompute,
		},
	}
	nodeInfra := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeInfra,
			Labels: map[string]string{
				labelInfra: labelValue,
			},
			Annotations: annotationsInfra,
		},
	}
	_, _ = c.nodes.Create(context.Background(), nodeControl, metav1.CreateOptions{})
	_, _ = c.nodes.Create(context.Background(), nodeCompute, metav1.CreateOptions{})
	_, _ = c.nodes.Create(context.Background(), nodeInfra, metav1.CreateOptions{})

	return c
}
