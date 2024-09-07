package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha1"
	"github.com/heathcliff26/kube-upgrade/pkg/client/clientset/versioned/scheme"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func init() {
	ctrl.SetLogger(klog.NewKlogr())
}

type controller struct {
	client.Client
	manager manager.Manager
	nodes   clientv1.NodeInterface
}

// Run make generate when changing these comments
// +kubebuilder:rbac:groups=kubeupgrade.heathcliff.eu,resources=kubeupgradeplans,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubeupgrade.heathcliff.eu,resources=kubeupgradeplans/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=nodes,verbs=list;update

func NewController(name string) (*controller, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	ns, err := GetNamespace()
	if err != nil {
		return nil, err
	}

	mgr, err := ctrl.NewManager(config, manager.Options{
		Scheme:                        scheme.Scheme,
		LeaderElection:                true,
		LeaderElectionNamespace:       ns,
		LeaderElectionID:              name,
		LeaderElectionReleaseOnCancel: true,
		LeaseDuration:                 Pointer(time.Minute),
		RenewDeadline:                 Pointer(10 * time.Second),
		RetryPeriod:                   Pointer(5 * time.Second),
		HealthProbeBindAddress:        ":9090",
	})
	if err != nil {
		return nil, err
	}
	err = mgr.AddHealthzCheck("healthz", healthz.Ping)
	if err != nil {
		return nil, err
	}
	err = mgr.AddReadyzCheck("readyz", healthz.Ping)
	if err != nil {
		return nil, err
	}

	return &controller{
		manager: mgr,
		nodes:   client.CoreV1().Nodes(),
		Client:  mgr.GetClient(),
	}, nil
}

func (c *controller) Run() error {
	err := ctrl.NewControllerManagedBy(c.manager).For(&api.KubeUpgradePlan{}).Complete(c)
	if err != nil {
		return err
	}

	err = ctrl.NewWebhookManagedBy(c.manager).
		For(&api.KubeUpgradePlan{}).
		WithDefaulter(&planMutatingHook{}).
		Complete()
	if err != nil {
		return err
	}

	return c.manager.Start(signals.SetupSignalHandler())
}

func (c *controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := klog.LoggerWithValues(klog.NewKlogr(), "plan", req.Name)

	var plan api.KubeUpgradePlan
	err := c.Get(ctx, req.NamespacedName, &plan)
	if err != nil {
		logger.Error(err, "Failed to get Plan")
		return ctrl.Result{}, err
	}

	err = c.reconcile(ctx, &plan, logger)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = c.Status().Update(ctx, &plan)
	if err != nil {
		logger.Error(err, "Failed to update plan status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{
		Requeue:      plan.Status.Summary != api.PlanStatusComplete,
		RequeueAfter: time.Minute,
	}, nil
}

func (c *controller) reconcile(ctx context.Context, plan *api.KubeUpgradePlan, logger logr.Logger) error {
	if plan.Status.Groups == nil {
		plan.Status.Groups = make(map[string]string, len(plan.Spec.Groups))
	}

	nodesToUpdate := make(map[string][]corev1.Node, len(plan.Spec.Groups))
	newGroupStatus := make(map[string]string, len(plan.Spec.Groups))

	for name, cfg := range plan.Spec.Groups {
		upgradedCfg := combineConfig(plan.Spec.Upgraded, plan.Spec.Groups[name].Upgraded)

		nodeList, err := c.nodes.List(ctx, metav1.ListOptions{
			LabelSelector: labels.SelectorFromSet(cfg.Labels).String(),
		})
		if err != nil {
			logger.WithValues("group", name).Error(err, "Failed to get nodes for group")
			return err
		}

		status, update, nodes, err := c.reconcileNodes(plan.Spec.KubernetesVersion, nodeList.Items, upgradedCfg)
		if err != nil {
			logger.WithValues("group", name).Error(err, "Failed to reconcile nodes for group")
			return err
		}

		newGroupStatus[name] = status

		if update {
			nodesToUpdate[name] = nodes
		} else if plan.Status.Groups[name] != newGroupStatus[name] {
			logger.WithValues("group", name, "status", newGroupStatus[name]).Info("Group changed status")
		}
	}

	for name, nodes := range nodesToUpdate {
		if groupWaitForDependency(plan.Spec.Groups[name].DependsOn, newGroupStatus) {
			logger.WithValues("group", name).Info("Group is waiting on dependencies")
			newGroupStatus[name] = api.PlanStatusWaiting
			continue
		} else if plan.Status.Groups[name] != newGroupStatus[name] {
			logger.WithValues("group", name, "status", newGroupStatus[name]).Info("Group changed status")
		}

		for _, node := range nodes {
			_, err := c.nodes.Update(ctx, &node, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update node %s: %v", node.GetName(), err)
			}
		}
	}

	plan.Status.Groups = newGroupStatus
	plan.Status.Summary = createStatusSummary(plan.Status.Groups)

	return nil
}

func (c *controller) reconcileNodes(kubeVersion string, nodes []corev1.Node, cfgAnnotations map[string]string) (string, bool, []corev1.Node, error) {
	if len(nodes) == 0 {
		return api.PlanStatusUnknown, false, nil, nil
	}

	completed := true
	needUpdate := false

	for i := range nodes {
		if nodes[i].Annotations == nil {
			nodes[i].Annotations = make(map[string]string)
		}

		if applyConfigAnnotations(nodes[i].Annotations, cfgAnnotations) {
			needUpdate = true
		}

		if nodes[i].Annotations[constants.NodeKubernetesVersion] == kubeVersion {
			if nodes[i].Annotations[constants.NodeUpgradeStatus] != constants.NodeUpgradeStatusCompleted {
				completed = false
			}
			continue
		}

		completed = false
		nodes[i].Annotations[constants.NodeKubernetesVersion] = kubeVersion
		nodes[i].Annotations[constants.NodeUpgradeStatus] = constants.NodeUpgradeStatusPending

		needUpdate = true
	}

	var status string
	if completed {
		status = api.PlanStatusComplete
	} else {
		status = api.PlanStatusProgressing
	}
	return status, needUpdate, nodes, nil
}
