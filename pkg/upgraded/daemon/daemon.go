package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/config"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/fleetlock"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/kubeadm"
	rpmostree "github.com/heathcliff26/kube-upgrade/pkg/upgraded/rpm-ostree"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type daemon struct {
	fleetlock     *fleetlock.FleetlockClient
	checkInterval time.Duration
	retryInterval time.Duration

	rpmostree *rpmostree.RPMOStreeCMD
	kubeadm   *kubeadm.KubeadmCMD

	image string
	node  string

	client kubernetes.Interface
	ctx    context.Context
	cancel context.CancelFunc

	upgrade sync.Mutex
}

// Create a new daemon
func NewDaemon(cfg *config.Config) (*daemon, error) {
	fleetlockClient, err := fleetlock.NewClient(cfg.Fleetlock.URL, cfg.Fleetlock.Group)
	if err != nil {
		return nil, fmt.Errorf("failed to create fleetlock client: %v", err)
	}

	rpmOstreeCMD, err := rpmostree.New(cfg.RPMOStreePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create rpm-ostree cmd wrapper: %v", err)
	}
	kubeadmCMD, err := kubeadm.New(cfg.KubeadmPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubeadm cmd wrapper: %v", err)
	}

	if cfg.Kubeconfig == "" {
		return nil, fmt.Errorf("no kubeconfig provided")
	}
	config, err := clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeconfig: %v", err)
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	if cfg.Image == "" {
		return nil, fmt.Errorf("no image provided for kubernetes updates")
	}

	node, err := findNodeByMachineID(kubeClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes node name for host: %v", err)
	}
	slog.Info("Found node name for this host", slog.String("node", node))

	return &daemon{
		fleetlock:     fleetlockClient,
		checkInterval: cfg.CheckInterval,
		retryInterval: cfg.RetryInterval,

		rpmostree: rpmOstreeCMD,
		kubeadm:   kubeadmCMD,

		image:  cfg.Image,
		node:   node,
		client: kubeClient,
	}, nil
}

// Retries the given function until it succeeds
func (d *daemon) retry(f func() bool) {
	for !f() {
		select {
		case <-d.ctx.Done():
			return
		case <-time.After(d.retryInterval):
		}
	}
}

// Will try to release the lock until successfull
func (d *daemon) releaseLock() {
	d.retry(func() bool {
		err := d.fleetlock.Release()
		if err == nil {
			return true
		}

		slog.Warn("Failed to release lock", "err", err)
		return false
	})
}

// Run the main daemon loop
func (d *daemon) Run() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	d.ctx = ctx
	d.cancel = cancel
	go func() {
		<-stop
		cancel()
	}()

	node, err := d.client.CoreV1().Nodes().Get(d.ctx, d.node, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node status: %v", err)
	}

	if !nodeNeedsUpgrade(node) {
		slog.Debug("Releasing any log that may be held by this machine")
		d.releaseLock()
		if d.ctx.Err() != nil {
			return nil
		}
	} else {
		slog.Info("Node is in the middle of a kubernetes upgrade, not releasing the lock")
	}

	slog.Info("Starting daemon")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		d.watchForUpgrade()
		slog.Info("Stopped watching for upgrades")
	}()
	go func() {
		defer wg.Done()
		d.watchForKubernetesUpgrade()
		slog.Info("Stopped watching for kubernetes upgrades")
	}()

	wg.Wait()
	return nil
}

// Check for os upgrades and perform them if necessary.
// Runs until context is cancelled
func (d *daemon) watchForUpgrade() {
	var needUpgrade bool
	for {
		d.retry(func() bool {
			var err error
			slog.Debug("Checking for upgrades via rpm-ostree")
			needUpgrade, err = d.rpmostree.CheckForUpgrade()
			if err == nil {
				return true
			}
			slog.Error("Failed to check if there is a new upgrade", "err", err)
			return false
		})

		if needUpgrade {
			slog.Info("New upgrade is necessary, trying to start update")
			d.retry(func() bool {
				err := d.doUpgrade()
				if err == nil {
					return true
				}
				slog.Error("Failed to perform rpm-ostree upgrade", "err", err)
				return false
			})
		} else {
			slog.Debug("No upgrades found")
		}

		select {
		case <-d.ctx.Done():
			return
		case <-time.After(d.checkInterval):
		}
	}
}

// Perform rpm-ostree upgrade
func (d *daemon) doUpgrade() error {
	d.upgrade.Lock()
	defer d.upgrade.Unlock()

	err := d.fleetlock.Lock()
	if err != nil {
		return fmt.Errorf("failed to aquire lock: %v", err)
	}

	err = d.rpmostree.Upgrade()
	if err != nil {
		return err
	}

	// This should not be reached, as rpmostree.Upgrade() reboots the node on success.
	// I included it here mainly for completness sake.

	d.releaseLock()
	return nil
}

// Watch for kubernetes upgrades and preform them if necessary
func (d *daemon) watchForKubernetesUpgrade() {
	factory := informers.NewSharedInformerFactoryWithOptions(d.client, time.Minute, informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.FieldSelector = fields.SelectorFromSet(fields.Set{"metadata.name": d.node}).String()
	}))

	informer := factory.Core().V1().Nodes().Informer()
	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: d.checkNodeStatus,
		UpdateFunc: func(_, newObj interface{}) {
			d.checkNodeStatus(newObj)
		},
		DeleteFunc: func(_ interface{}) {
			slog.Error("Node has been deleted from cluster")
			d.cancel()
		},
	})
	if err != nil {
		slog.Error("Failed to add event handlers to kubernetes informer")
		d.cancel()
		return
	}
	err = informer.SetWatchErrorHandler(cache.DefaultWatchErrorHandler)
	if err != nil {
		slog.Error("Failed to set watch error handler to kubernetes informer")
		d.cancel()
		return
	}
	slog.Info("Watching for new kubernetes upgrades")
	informer.Run(d.ctx.Done())
}

// Check if we need to upgrade the node and trigger the upgrade if needed
func (d *daemon) checkNodeStatus(obj interface{}) {
	node := obj.(*corev1.Node)

	if !nodeNeedsUpgrade(node) {
		return
	}

	d.retry(func() bool {
		err := d.doNodeUpgrade(node)
		if err == nil {
			return true
		}
		slog.Error("Failed to upgrade node", "err", err, slog.String("node", node.GetName()))
		return false
	})
}

// Update the node by first rebasing to a new version and then upgrading kubernetes
func (d *daemon) doNodeUpgrade(node *corev1.Node) error {
	d.upgrade.Lock()
	defer d.upgrade.Unlock()

	version := node.Annotations[constants.KubernetesVersionAnnotation]
	slog.Info("Attempting node upgrade to new kubernetes version", slog.String("node", node.GetName()), slog.String("version", version))

	err := d.fleetlock.Lock()
	if err != nil {
		return fmt.Errorf("failed to aquire lock: %v", err)
	}

	if version != node.Status.NodeInfo.KubeletVersion {
		slog.Info("Rebasing os to new kubernetes version", slog.String("version", version))
		err := d.updateNodeStatus(constants.NodeUpgradeStatusRebasing)
		if err != nil {
			return fmt.Errorf("failed to update node status: %v", err)
		}
		err = d.rpmostree.Rebase(d.image + ":" + version)
		if err != nil {
			return fmt.Errorf("failed to rebase node: %v", err)
		}
	}

	slog.Info("Updating node via kubeadm")

	err = d.updateNodeStatus(constants.NodeUpgradeStatusUpgrading)
	if err != nil {
		return fmt.Errorf("failed to update node status: %v", err)
	}

	kubeadmConfigMap, err := d.client.CoreV1().ConfigMaps("kube-system").Get(d.ctx, "kubeadm-config", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to fetch kubeadm-config: %v", err)
	}
	if kubeadmConfigMap.Data == nil {
		return fmt.Errorf("kubeadm configmap contains no data")
	}
	var kubeadmConfig kubeadm.ClusterConfiguration
	err = yaml.Unmarshal([]byte(kubeadmConfigMap.Data["ClusterConfiguration"]), &kubeadmConfig)
	if err != nil {
		return fmt.Errorf("failed to parse kubeadm-config: %v", err)
	}

	if version != kubeadmConfig.KubernetesVersion {
		slog.Info("kubeadm-config kubernetesVersion does not match requested version, initializing upgrade", slog.String("kubernetesVersion", kubeadmConfig.KubernetesVersion), slog.String("version", version))
		err = d.kubeadm.Apply(version)
	} else {
		slog.Debug("Cluster upgrade is already initialized, upgrading node")
		err = d.kubeadm.Node()
	}
	if err != nil {
		return fmt.Errorf("failed run kubeadm: %v", err)
	}

	err = d.updateNodeStatus(constants.NodeUpgradeStatusCompleted)
	if err != nil {
		return fmt.Errorf("failed to update node status: %v", err)
	}

	slog.Info("Finished node upgrade, releasing lock")
	d.releaseLock()
	return nil
}

// Update the kube-upgrade node status annotation with the given status
func (d *daemon) updateNodeStatus(status string) error {
	node, err := d.client.CoreV1().Nodes().Get(d.ctx, d.node, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}
	node.Annotations[constants.KubernetesUpgradeStatus] = status

	_, err = d.client.CoreV1().Nodes().Update(d.ctx, node, metav1.UpdateOptions{})
	if err == nil {
		slog.Debug("Set node status", slog.String("status", status))
	}
	return err
}
