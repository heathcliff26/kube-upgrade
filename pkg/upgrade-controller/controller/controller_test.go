package controller

import (
	"testing"
	"time"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	upgradedconfig "github.com/heathcliff26/kube-upgrade/pkg/upgraded/config"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeFake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"
	controllerFake "sigs.k8s.io/controller-runtime/pkg/client/fake"
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
							Labels: map[string]string{labelControl: labelValue},
						},
						groupCompute: {
							DependsOn: []string{groupControl, groupInfra},
							Labels:    map[string]string{labelCompute: labelValue},
						},
						groupInfra: {
							DependsOn: []string{groupControl},
							Labels:    map[string]string{labelInfra: labelValue},
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
							Labels: map[string]string{labelControl: labelValue},
						},
						groupCompute: {
							DependsOn: []string{groupControl, groupInfra},
							Labels:    map[string]string{labelCompute: labelValue},
						},
						groupInfra: {
							DependsOn: []string{groupControl},
							Labels:    map[string]string{labelInfra: labelValue},
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
							Labels: map[string]string{labelControl: labelValue},
						},
						groupCompute: {
							DependsOn: []string{groupControl, groupInfra},
							Labels:    map[string]string{labelCompute: labelValue},
						},
						groupInfra: {
							DependsOn: []string{groupControl},
							Labels:    map[string]string{labelInfra: labelValue},
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
							Labels: map[string]string{labelControl: labelValue},
						},
						groupCompute: {
							DependsOn: []string{groupControl, groupInfra},
							Labels:    map[string]string{labelCompute: labelValue},
						},
						groupInfra: {
							DependsOn: []string{groupControl},
							Labels:    map[string]string{labelInfra: labelValue},
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
							Labels: map[string]string{labelControl: labelValue},
						},
						groupCompute: {
							DependsOn: []string{groupControl},
							Labels:    map[string]string{labelCompute: labelValue},
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
							Labels: map[string]string{labelControl: labelValue},
						},
						groupCompute: {
							DependsOn: []string{groupControl},
							Labels:    map[string]string{labelCompute: labelValue},
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
							Labels: map[string]string{labelControl: labelValue},
						},
						groupCompute: {
							DependsOn: []string{groupControl},
							Labels:    map[string]string{labelCompute: labelValue},
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
			Name: "DeleteConfigurationNotations",
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
							Labels: map[string]string{labelControl: labelValue},
							Upgraded: &api.UpgradedConfig{
								FleetlockGroup: "control-plane",
							},
						},
						groupCompute: {
							DependsOn: []string{groupControl},
							Labels:    map[string]string{labelCompute: labelValue},
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
				constants.ConfigStream:          "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:    "https://fleetlock.example.org",
				constants.ConfigFleetlockGroup:  "control-plane",
				constants.ConfigCheckInterval:   "2m",
				constants.ConfigRetryInterval:   "3m",
			},
			AnnotationsCompute: map[string]string{
				constants.NodeKubernetesVersion: "v1.30.4",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
				constants.ConfigStream:          "registry.example.com/test-stream",
				constants.ConfigFleetlockURL:    "https://fleetlock.example.org",
				constants.ConfigFleetlockGroup:  groupCompute,
				constants.ConfigCheckInterval:   "2m",
				constants.ConfigRetryInterval:   "3m",
			},
			ExpectedSummary: api.PlanStatusComplete,
			ExpectedGroupStatus: map[string]string{
				groupControl: api.PlanStatusComplete,
				groupCompute: api.PlanStatusComplete,
			},
			ExpectedAnnotationsControl: map[string]string{
				constants.NodeKubernetesVersion: "v1.30.4",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
			},
			ExpectedAnnotationsCompute: map[string]string{
				constants.NodeKubernetesVersion: "v1.30.4",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusCompleted,
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
							Labels: map[string]string{labelControl: labelValue},
						},
						groupCompute: {
							DependsOn: []string{groupControl, groupInfra},
							Labels:    map[string]string{labelCompute: labelValue},
						},
						groupInfra: {
							DependsOn: []string{groupControl},
							Labels:    map[string]string{labelInfra: labelValue},
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
							Labels: map[string]string{labelControl: labelValue},
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
			c := createFakeController(t, tCase.AnnotationsControl, tCase.AnnotationsCompute, tCase.AnnotationsInfra, &tCase.Plan)

			assert := assert.New(t)

			ctx := t.Context()

			err := c.reconcile(ctx, &tCase.Plan, klog.NewKlogr())

			if !assert.NoError(err, "Reconcile should succeed") {
				t.FailNow()
			}

			assert.Equal(tCase.ExpectedSummary, tCase.Plan.Status.Summary, "Summary should be correct")

			if !assert.Equal(len(tCase.Plan.Spec.Groups), len(tCase.Plan.Status.Groups), "Group lengths should match") {
				t.FailNow()
			}

			assert.Equal(tCase.ExpectedGroupStatus, tCase.Plan.Status.Groups, "Group status should match")

			nodeControl, _ := c.client.CoreV1().Nodes().Get(ctx, nodeControl, metav1.GetOptions{})
			nodeCompute, _ := c.client.CoreV1().Nodes().Get(ctx, nodeCompute, metav1.GetOptions{})
			nodeInfra, _ := c.client.CoreV1().Nodes().Get(ctx, nodeInfra, metav1.GetOptions{})

			assert.Equal(tCase.ExpectedAnnotationsControl, nodeControl.GetAnnotations(), "Control group should have expected annotations")
			assert.Equal(tCase.ExpectedAnnotationsCompute, nodeCompute.GetAnnotations(), "Compute group should have expected annotations")
			assert.Equal(tCase.ExpectedAnnotationsInfra, nodeInfra.GetAnnotations(), "Infra group should have expected annotations")
		})
	}
}

func TestReconcileNodes(t *testing.T) {
	c := &controller{
		client: kubeFake.NewSimpleClientset(),
	}

	nodeControl := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeControl,
			Labels: map[string]string{
				labelControl: labelValue,
			},
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				KubeletVersion: "v1.31.1",
			},
		},
	}
	nodeControl, _ = c.client.CoreV1().Nodes().Create(t.Context(), nodeControl, metav1.CreateOptions{})

	assert := assert.New(t)

	status, needUpdate, nodes, err := c.reconcileNodes("v1.31.0", false, []corev1.Node{*nodeControl})

	assert.Equal(api.PlanStatusError, status, "Should return error status")
	assert.False(needUpdate, "Should not request update")
	assert.Nil(nodes, "Should not return nodes")
	assert.Error(err, "Should return an error")

	status, needUpdate, nodes, err = c.reconcileNodes("v1.31.0", true, []corev1.Node{*nodeControl})

	assert.NotEqual(api.PlanStatusError, status, "Should not return error status")
	assert.True(needUpdate, "Should request update")
	assert.Equal("v1.31.0", nodes[0].GetAnnotations()[constants.NodeKubernetesVersion], "Should set kubernetes wanted version on node")
	assert.NoError(err, "Should not return an error")
}

func TestReconcileUpgradedDaemons(t *testing.T) {
	tMatrix := []struct {
		Name                   string
		InitialDaemons, Groups []string
	}{
		{
			Name:   "CreateAllDaemons",
			Groups: []string{groupControl, groupCompute, groupInfra},
		},
		{
			Name:   "DeleteExtraDaemons",
			Groups: []string{groupControl, groupCompute},
			InitialDaemons: []string{
				groupControl,
				groupCompute,
				groupInfra,
			},
		},
		{
			Name:   "CreateMissingDaemons",
			Groups: []string{groupControl, groupCompute, groupInfra},
			InitialDaemons: []string{
				groupControl,
				groupCompute,
			},
		},
		{
			Name:   "CreateAndDeleteDaemons",
			Groups: []string{groupControl, groupCompute},
			InitialDaemons: []string{
				groupControl,
				groupInfra,
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			plan := &api.KubeUpgradePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "upgrade-plan",
				},
				Spec: api.KubeUpgradeSpec{
					KubernetesVersion: "v1.31.0",
					Groups:            make(map[string]api.KubeUpgradePlanGroup, 3),
				},
			}
			for _, group := range tCase.Groups {
				plan.Spec.Groups[group] = api.KubeUpgradePlanGroup{
					Labels: map[string]string{"node-role.kubernetes.io/" + group: labelValue},
				}
			}
			c := createFakeController(t, nil, nil, nil, plan)
			c.upgradedImage = "registry.example.com/kube-upgrade:latest"
			for _, group := range tCase.InitialDaemons {
				addFakeUpgradedConfigMap(t, c, plan.Name, group)
				addFakeUpgradedDaemonset(t, c, plan.Name, group)
			}

			assert.NoError(c.reconcile(t.Context(), plan, klog.NewKlogr()), "Reconcile should succeed")

			daemonsList, err := c.client.AppsV1().DaemonSets(c.namespace).List(t.Context(), metav1.ListOptions{})
			assert.NoError(err, "Should list daemonsets without error")

			for _, daemon := range daemonsList.Items {
				assert.Equalf(plan.Name, daemon.Labels[constants.LabelPlanName], "Daemonset %s should have plan name as label", daemon.Name)
				assert.Containsf(tCase.Groups, daemon.Labels[constants.LabelNodeGroup], "Daemonset %s should belong to a valid group", daemon.Name)
				assert.Len(daemon.Spec.Template.Spec.NodeSelector, 1, "Should have exactly 1 label")
				assert.Equal("registry.example.com/kube-upgrade:latest", daemon.Spec.Template.Spec.Containers[0].Image, "Daemonset should have correct upgraded image")
			}

			cmList, err := c.client.CoreV1().ConfigMaps(c.namespace).List(t.Context(), metav1.ListOptions{})
			assert.NoError(err, "Should list configmaps without error")

			for _, cm := range cmList.Items {
				assert.Equalf(plan.Name, cm.Labels[constants.LabelPlanName], "ConfigMap %s should have plan name as label", cm.Name)
				assert.Containsf(tCase.Groups, cm.Labels[constants.LabelNodeGroup], "ConfigMap %s should belong to a valid group", cm.Name)
				assert.Len(cm.Data, 1, "ConfigMap should have exactly 1 data entry")
			}
		})
	}
	t.Run("UpdateDaemonSet", func(t *testing.T) {
		assert := assert.New(t)

		plan := &api.KubeUpgradePlan{
			ObjectMeta: metav1.ObjectMeta{
				Name: "upgrade-plan",
			},
			Spec: api.KubeUpgradeSpec{
				KubernetesVersion: "v1.31.0",
				Groups: map[string]api.KubeUpgradePlanGroup{
					groupControl: {
						Labels: map[string]string{labelControl: labelValue},
					},
				},
			},
		}
		c := createFakeController(t, nil, nil, nil, plan)
		daemon := c.NewEmptyUpgradedDaemonSet(plan.Name, groupControl)
		daemon.Spec = c.NewUpgradedDaemonSetSpec(plan.Name, groupControl)
		daemon.Spec.Template.Spec.HostNetwork = true
		daemon.Spec.Template.Spec.HostPID = false
		_, _ = c.client.AppsV1().DaemonSets(c.namespace).Create(t.Context(), &daemon, metav1.CreateOptions{})

		assert.NoError(c.reconcile(t.Context(), plan, klog.NewKlogr()), "Reconcile should succeed")

		result, err := c.client.AppsV1().DaemonSets(c.namespace).Get(t.Context(), daemon.Name, metav1.GetOptions{})
		assert.NoError(err, "Should get daemonset without error")
		assert.False(result.Spec.Template.Spec.HostNetwork, "Daemonset HostNetwork should be updated to false")
		assert.True(result.Spec.Template.Spec.HostPID, "Daemonset HostPID should be updated to true")
	})
	t.Run("UpdateConfigMap", func(t *testing.T) {
		assert := assert.New(t)

		cfg := &api.UpgradedConfig{}
		api.SetObjectDefaults_UpgradedConfig(cfg)

		plan := &api.KubeUpgradePlan{
			ObjectMeta: metav1.ObjectMeta{
				Name: "upgrade-plan",
			},
			Spec: api.KubeUpgradeSpec{
				KubernetesVersion: "v1.31.0",
				Groups: map[string]api.KubeUpgradePlanGroup{
					groupControl: {
						Labels:   map[string]string{labelControl: labelValue},
						Upgraded: cfg.DeepCopy(),
					},
				},
			},
		}
		c := createFakeController(t, nil, nil, nil, plan)
		cm := c.NewEmptyUpgradedConfigMap(plan.Name, groupControl)
		cfg.Stream = "registry.example.com/updated-stream"
		_ = c.AttachUpgradedConfigMapData(&cm, cfg)
		_, _ = c.client.CoreV1().ConfigMaps(c.namespace).Create(t.Context(), &cm, metav1.CreateOptions{})

		assert.NoError(c.reconcile(t.Context(), plan, klog.NewKlogr()), "Reconcile should succeed")

		result, err := c.client.CoreV1().ConfigMaps(c.namespace).Get(t.Context(), cm.Name, metav1.GetOptions{})
		assert.NoError(err, "Should get daemonset without error")
		assert.Contains(result.Data[upgradedconfig.DefaultConfigFile], api.DefaultUpgradedStream, "ConfigMap data should be updated")
	})
}

func TestPlanFinalizer(t *testing.T) {
	assert := assert.New(t)

	plan := &api.KubeUpgradePlan{
		ObjectMeta: metav1.ObjectMeta{
			Name: "upgrade-plan",
		},
		Spec: api.KubeUpgradeSpec{
			KubernetesVersion: "v1.31.0",
			Groups: map[string]api.KubeUpgradePlanGroup{
				groupControl: {
					Labels: map[string]string{labelControl: labelValue},
				},
			},
		},
	}
	c := createFakeController(t, nil, nil, nil, plan)

	assert.NoError(c.reconcile(t.Context(), plan, klog.NewKlogr()), "First reconcile should succeed")
	assert.Contains(plan.GetFinalizers(), constants.Finalizer, "Plan should have finalizer after first reconcile")

	daemonsList, err := c.client.AppsV1().DaemonSets(c.namespace).List(t.Context(), metav1.ListOptions{})
	assert.NoError(err, "Should list daemonsets without error")
	assert.NotEmpty(daemonsList.Items, "There should be daemonsets present")

	plan.SetDeletionTimestamp(&metav1.Time{Time: time.Now()})
	tmpCtrl := createFakeController(t, nil, nil, nil, plan)
	c.Client = tmpCtrl.Client

	assert.NoError(c.reconcile(t.Context(), plan, klog.NewKlogr()), "Second reconcile should succeed")
	assert.NotContains(plan.GetFinalizers(), constants.Finalizer, "Finalizer should be removed for deletion")

	daemonsList, err = c.client.AppsV1().DaemonSets(c.namespace).List(t.Context(), metav1.ListOptions{})
	assert.NoError(err, "Should list daemonsets without error")
	assert.Empty(daemonsList.Items, "DaemonSet should be deleted after plan deletion")
}

func createFakeController(t *testing.T, annotationsControl, annotationsCompute, annotationsInfra map[string]string, plan *api.KubeUpgradePlan) *controller {
	scheme := runtime.NewScheme()
	_ = api.AddToScheme(scheme)
	fakeCtrlClient := controllerFake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(plan).Build()
	c := &controller{
		client:    kubeFake.NewSimpleClientset(),
		Client:    fakeCtrlClient,
		namespace: "kube-upgrade",
	}

	nodeControl := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeControl,
			Labels: map[string]string{
				labelControl: labelValue,
			},
			Annotations: annotationsControl,
		},
	}
	nodeCompute := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeCompute,
			Labels: map[string]string{
				labelCompute: labelValue,
			},
			Annotations: annotationsCompute,
		},
	}
	nodeInfra := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeInfra,
			Labels: map[string]string{
				labelInfra: labelValue,
			},
			Annotations: annotationsInfra,
		},
	}

	ctx := t.Context()
	_, _ = c.client.CoreV1().Nodes().Create(ctx, nodeControl, metav1.CreateOptions{})
	_, _ = c.client.CoreV1().Nodes().Create(ctx, nodeCompute, metav1.CreateOptions{})
	_, _ = c.client.CoreV1().Nodes().Create(ctx, nodeInfra, metav1.CreateOptions{})

	return c
}

func addFakeUpgradedConfigMap(t *testing.T, c *controller, plan, group string) {
	cfg := &api.UpgradedConfig{}
	api.SetObjectDefaults_UpgradedConfig(cfg)
	cm := c.NewEmptyUpgradedConfigMap(plan, group)
	_ = c.AttachUpgradedConfigMapData(&cm, cfg)

	_, _ = c.client.CoreV1().ConfigMaps(c.namespace).Create(t.Context(), &cm, metav1.CreateOptions{})
}

func addFakeUpgradedDaemonset(t *testing.T, c *controller, plan, group string) {
	daemon := c.NewEmptyUpgradedDaemonSet(plan, group)
	daemon.Spec = c.NewUpgradedDaemonSetSpec(plan, group)

	_, _ = c.client.AppsV1().DaemonSets(c.namespace).Create(t.Context(), &daemon, metav1.CreateOptions{})
}
