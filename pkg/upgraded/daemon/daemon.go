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

	fleetlock "github.com/heathcliff26/fleetlock/pkg/client"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/config"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/kubeadm"
	rpmostree "github.com/heathcliff26/kube-upgrade/pkg/upgraded/rpm-ostree"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultStream         = "ghcr.io/heathcliff26/fcos-k8s"
	defaultFleetlockGroup = "default"
	defaultCheckInterval  = 3 * time.Hour
	defaultRetryInterval  = 5 * time.Minute
)

type daemon struct {
	fleetlock     *fleetlock.FleetlockClient
	checkInterval time.Duration
	retryInterval time.Duration

	rpmostree *rpmostree.RPMOStreeCMD
	kubeadm   *kubeadm.KubeadmCMD

	stream string
	node   string

	client kubernetes.Interface
	ctx    context.Context
	cancel context.CancelFunc

	upgrade sync.Mutex
}

// Create a new daemon
func NewDaemon(cfg *config.Config) (*daemon, error) {
	fleetlockClient, err := fleetlock.NewEmptyClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create fleetlock client: %v", err)

	}
	err = fleetlockClient.SetGroup(defaultFleetlockGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to set fleetlock group: %v", err)
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

	machineID, err := utils.GetMachineID()
	if err != nil {
		return nil, fmt.Errorf("failed to get machine-id: %v", err)
	}
	node, err := findNode(kubeClient, machineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes node name for host: %v", err)
	}
	slog.Info("Found node name for this host", slog.String("node", node))

	return &daemon{
		fleetlock:     fleetlockClient,
		checkInterval: defaultCheckInterval,
		retryInterval: defaultRetryInterval,

		rpmostree: rpmOstreeCMD,
		kubeadm:   kubeadmCMD,

		stream: defaultStream,
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

	node, err := d.getNode()
	if err != nil {
		return fmt.Errorf("failed to get node status: %v", err)
	}

	err = d.UpdateConfigFromAnnotations(node.GetAnnotations())
	if err != nil {
		return fmt.Errorf("failed to update daemon config from node annotations: %v", err)
	}

	node, err = d.annotateNodeWithUpgradedVersion(node)
	if err != nil {
		return fmt.Errorf("failed to annotate node with upgraded version: %v", err)
	}

	if !nodeNeedsUpgrade(node) {
		slog.Debug("Releasing any log that may be held by this machine")
		d.releaseLock()
		if d.ctx.Err() != nil {
			return nil
		}
	} else {
		slog.Info("Node needs upgrade or is in the middle of one, upgrading node before starting daemon")
		d.doNodeUpgradeWithRetry(node)
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
		d.watchForNodeUpgrade()
		slog.Info("Stopped watching for kubernetes upgrades")
	}()

	wg.Wait()
	return nil
}
