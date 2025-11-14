package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"golang.org/x/mod/semver"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

const (
	defaultUpgradedImage = "ghcr.io/heathcliff26/kube-upgraded"
	upgradedImageEnv     = "UPGRADED_IMAGE"
	upgradedTagEnv       = "UPGRADED_TAG"
)

func init() {
	ctrl.SetLogger(klog.NewKlogr())
}

type controller struct {
	client.Client
	manager       manager.Manager
	namespace     string
	upgradedImage string
}

// Run make generate when changing these comments
// +kubebuilder:rbac:groups=kubeupgrade.heathcliff.eu,resources=kubeupgradeplans,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubeupgrade.heathcliff.eu,resources=kubeupgradeplans/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=nodes,verbs=list;update
// +kubebuilder:rbac:groups="",namespace=kube-upgrade,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups="coordination.k8s.io",namespace=kube-upgrade,resources=leases,verbs=create;get;update
// +kubebuilder:rbac:groups="apps",namespace=kube-upgrade,resources=daemonsets,verbs=list;watch;create;update;delete
// +kubebuilder:rbac:groups="",namespace=kube-upgrade,resources=configmaps,verbs=list;watch;create;update;delete

func NewController(name string) (*controller, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	ns, err := GetNamespace()
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()
	err = api.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}
	err = clientgoscheme.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}

	mgr, err := ctrl.NewManager(config, manager.Options{
		Scheme:                        scheme,
		LeaderElection:                true,
		LeaderElectionNamespace:       ns,
		LeaderElectionID:              name,
		LeaderElectionReleaseOnCancel: true,
		LeaseDuration:                 Pointer(time.Minute),
		RenewDeadline:                 Pointer(10 * time.Second),
		RetryPeriod:                   Pointer(5 * time.Second),
		HealthProbeBindAddress:        ":9090",
		Cache: cache.Options{
			DefaultNamespaces: map[string]cache.Config{ns: {}},
		},
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
		Client:        mgr.GetClient(),
		manager:       mgr,
		namespace:     ns,
		upgradedImage: GetUpgradedImage(),
	}, nil
}

func (c *controller) Run() error {
	err := ctrl.NewControllerManagedBy(c.manager).
		For(&api.KubeUpgradePlan{}).
		Owns(&appv1.DaemonSet{}).
		Owns(&corev1.ConfigMap{}).
		Complete(c)
	if err != nil {
		return err
	}

	err = ctrl.NewWebhookManagedBy(c.manager).
		For(&api.KubeUpgradePlan{}).
		WithDefaulter(&planMutatingHook{}).
		WithValidator(&planValidatingHook{}).
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

	// Migration from v0.6.0: Remove the finalizer as it is not needed
	// TODO: Remove in future release
	if controllerutil.RemoveFinalizer(plan, constants.Finalizer) {
		err := c.Update(ctx, plan)
		if err != nil {
			return fmt.Errorf("failed to remove finalizer from plan %s: %v", plan.Name, err)
		}
	}

	cmList := &corev1.ConfigMapList{}
	err := c.List(ctx, cmList, client.InNamespace(c.namespace), client.MatchingLabels{
		constants.LabelPlanName: plan.Name,
	})
	if err != nil {
		logger.WithValues("plan", plan.Name).Error(err, "Failed to fetch upgraded ConfigMaps")
		return err
	}

	dsList := &appv1.DaemonSetList{}
	err = c.List(ctx, dsList, client.InNamespace(c.namespace), client.MatchingLabels{
		constants.LabelPlanName: plan.Name,
	})
	if err != nil {
		logger.WithValues("plan", plan.Name).Error(err, "Failed to fetch upgraded DaemonSets")
		return err
	}

	daemons := make(map[string]*appv1.DaemonSet, len(plan.Spec.Groups))
	for i := range dsList.Items {
		daemon := &dsList.Items[i]
		group := daemon.Labels[constants.LabelNodeGroup]
		if _, ok := plan.Spec.Groups[group]; ok {
			daemons[group] = daemon
		} else {
			err = c.Delete(ctx, daemon)
			if err != nil {
				return fmt.Errorf("failed to delete DaemonSet %s: %v", daemon.Name, err)
			}
			logger.WithValues("name", daemon.Name).Info("Deleted obsolete DaemonSet")
		}
	}

	cms := make(map[string]*corev1.ConfigMap, len(plan.Spec.Groups))
	for i := range cmList.Items {
		cm := &cmList.Items[i]
		group := cm.Labels[constants.LabelNodeGroup]
		if _, ok := plan.Spec.Groups[group]; ok {
			cms[group] = cm
		} else {
			err = c.Delete(ctx, cm)
			if err != nil {
				return fmt.Errorf("failed to delete ConfigMap %s: %v", cm.Name, err)
			}
			logger.WithValues("name", cm.Name).Info("Deleted obsolete ConfigMap")
		}
	}

	nodesToUpdate := make(map[string][]corev1.Node, len(plan.Spec.Groups))
	newGroupStatus := make(map[string]string, len(plan.Spec.Groups))

	for name, cfg := range plan.Spec.Groups {
		err = c.reconcileUpgradedConfigMap(ctx, plan, logger, cms[name], name)
		if err != nil {
			return fmt.Errorf("failed to reconcile ConfigMap for group %s: %v", name, err)
		}

		err = c.reconcileUpgradedDaemonSet(ctx, plan, logger, daemons[name], name, cfg)
		if err != nil {
			return fmt.Errorf("failed to reconcile DaemonSet for group %s: %v", name, err)
		}

		nodeList := &corev1.NodeList{}
		err = c.List(ctx, nodeList, client.MatchingLabels(cfg.Labels))
		if err != nil {
			logger.WithValues("group", name).Error(err, "Failed to get nodes for group")
			return err
		}

		status, update, nodes, err := c.reconcileNodes(plan.Spec.KubernetesVersion, plan.Spec.AllowDowngrade, nodeList.Items)
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
			err = c.Update(ctx, &node)
			if err != nil {
				return fmt.Errorf("failed to update node %s: %v", node.GetName(), err)
			}
		}
	}

	plan.Status.Groups = newGroupStatus
	plan.Status.Summary = createStatusSummary(plan.Status.Groups)

	return nil
}

func (c *controller) reconcileNodes(kubeVersion string, downgrade bool, nodes []corev1.Node) (string, bool, []corev1.Node, error) {
	if len(nodes) == 0 {
		return api.PlanStatusUnknown, false, nil, nil
	}

	completed := 0
	needUpdate := false
	errorNodes := make([]string, 0)

	for i := range nodes {
		if nodes[i].Annotations == nil {
			nodes[i].Annotations = make(map[string]string)
		}

		// Step to cleanup after migration to v0.6.0
		// TODO: Remove in v0.7.0
		if deleteConfigAnnotations(nodes[i].Annotations) {
			needUpdate = true
		}

		if !downgrade && semver.Compare(kubeVersion, nodes[i].Status.NodeInfo.KubeletVersion) < 0 {
			return api.PlanStatusError, false, nil, fmt.Errorf("node %s version %s is newer than %s, but downgrade is disabled", nodes[i].GetName(), nodes[i].Status.NodeInfo.KubeletVersion, kubeVersion)
		}

		if nodes[i].Annotations[constants.NodeKubernetesVersion] == kubeVersion {
			switch nodes[i].Annotations[constants.NodeUpgradeStatus] {
			case constants.NodeUpgradeStatusCompleted:
				completed++
			case constants.NodeUpgradeStatusError:
				errorNodes = append(errorNodes, nodes[i].GetName())
			}
			continue
		}

		nodes[i].Annotations[constants.NodeKubernetesVersion] = kubeVersion
		nodes[i].Annotations[constants.NodeUpgradeStatus] = constants.NodeUpgradeStatusPending

		needUpdate = true
	}

	var status string
	if len(errorNodes) > 0 {
		status = fmt.Sprintf("%s: The nodes %v are reporting errors", api.PlanStatusError, errorNodes)
	} else if len(nodes) == completed {
		status = api.PlanStatusComplete
	} else {
		status = fmt.Sprintf("%s: %d/%d nodes upgraded", api.PlanStatusProgressing, completed, len(nodes))
	}
	return status, needUpdate, nodes, nil
}
