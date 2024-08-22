package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha1"
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
	err := ctrl.NewControllerManagedBy(c.manager).For(&v1alpha1.KubeUpgradePlan{}).Complete(c)
	if err != nil {
		return err
	}
	return c.manager.Start(signals.SetupSignalHandler())
}

func (c *controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := klog.LoggerWithValues(klog.NewKlogr(), "plan", req.Name)

	var plan v1alpha1.KubeUpgradePlan
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
		Requeue:      plan.Status.Summary != v1alpha1.PlanStatusComplete,
		RequeueAfter: time.Minute,
	}, nil
}

func (c *controller) reconcile(ctx context.Context, plan *v1alpha1.KubeUpgradePlan, logger logr.Logger) error {
	if plan.Status.Groups == nil {
		plan.Status.Groups = make(map[string]string, len(plan.Spec.Groups))
	}

	for name, cfg := range plan.Spec.Groups {
		if groupWaitForDependency(cfg.DependsOn, plan.Status.Groups) {
			logger.WithValues("group", name).Info("Group is waiting on dependencies")
			plan.Status.Groups[name] = v1alpha1.PlanStatusWaiting
			continue
		}

		nodes, err := c.nodes.List(ctx, metav1.ListOptions{
			LabelSelector: labels.SelectorFromSet(cfg.Labels).String(),
		})
		if err != nil {
			logger.WithValues("group", name).Error(err, "Failed to get nodes for group")
			return err
		}

		status, err := c.reconcileNodes(ctx, plan.Spec.KubernetesVersion, nodes.Items)
		if err != nil {
			logger.WithValues("group", name).Error(err, "Failed to reconcile nodes for group")
			return err
		}

		if plan.Status.Groups[name] != status {
			logger.WithValues("group", name, "status", status).Info("Group changed status")
			plan.Status.Groups[name] = status
		}
	}

	plan.Status.Summary = createStatusSummary(plan.Status.Groups)

	return nil
}

func (c *controller) reconcileNodes(ctx context.Context, kubeVersion string, nodes []corev1.Node) (string, error) {
	if len(nodes) == 0 {
		return v1alpha1.PlanStatusUnknown, nil
	}

	completed := true

	for _, node := range nodes {
		if node.Annotations == nil {
			node.Annotations = make(map[string]string)
		}

		if node.Annotations[constants.KubernetesVersionAnnotation] == kubeVersion {
			if node.Annotations[constants.KubernetesUpgradeStatus] != constants.NodeUpgradeStatusCompleted {
				completed = false
			}
			continue
		}

		completed = false
		node.Annotations[constants.KubernetesVersionAnnotation] = kubeVersion
		node.Annotations[constants.KubernetesUpgradeStatus] = constants.NodeUpgradeStatusPending

		_, err := c.nodes.Update(ctx, &node, metav1.UpdateOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to update node %s: %v", node.GetName(), err)
		}
	}

	var status string
	if completed {
		status = v1alpha1.PlanStatusComplete
	} else {
		status = v1alpha1.PlanStatusProgressing
	}
	return status, nil
}
