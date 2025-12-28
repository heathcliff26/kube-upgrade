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

	"github.com/fsnotify/fsnotify"
	fleetlock "github.com/heathcliff26/fleetlock/pkg/client"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/config"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/kubeadm"
	rpmostree "github.com/heathcliff26/kube-upgrade/pkg/upgraded/rpm-ostree"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	hostPrefix       = "/host"
	rpmOstreeCMDPath = "/usr/bin/rpm-ostree"
)

type daemon struct {
	cfgPath string

	stream        string
	fleetlock     *fleetlock.FleetlockClient
	checkInterval time.Duration
	retryInterval time.Duration

	rpmostree *rpmostree.RPMOStreeCMD
	kubeadm   *kubeadm.KubeadmCMD

	node           string
	bootedImageRef string

	client kubernetes.Interface
	ctx    context.Context
	cancel context.CancelFunc

	configWatcher *fsnotify.Watcher

	configLock sync.RWMutex
	upgrade    sync.Mutex
}

// Create a new daemon
func NewDaemon(cfgPath string) (*daemon, error) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}
	if cfgPath == "" {
		cfgPath = config.DefaultConfigPath
	}

	// Hardcoded path, as it will be executed in a container
	rpmOstreeCMD, err := rpmostree.New(rpmOstreeCMDPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create rpm-ostree cmd wrapper: %v", err)
	}
	var kubeadmCMD *kubeadm.KubeadmCMD
	if cfg.KubeadmPath != "" {
		kubeadmCMD, err = kubeadm.NewFromPath(hostPrefix, cfg.KubeadmPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create kubeadm cmd wrapper: %v", err)
		}
	}

	config, err := clientcmd.BuildConfigFromFlags("", hostPrefix+cfg.KubeletConfig)
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
	node, err := nodeName(kubeClient, machineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes node name for host: %v", err)
	}
	slog.Info("Found node name for this host", slog.String("node", node))

	bootedImageRef, err := rpmOstreeCMD.GetBootedImageRef()
	if err != nil {
		return nil, fmt.Errorf("failed to get ostree image ref: %v", err)
	}

	d := &daemon{
		cfgPath: cfgPath,

		rpmostree: rpmOstreeCMD,
		kubeadm:   kubeadmCMD,

		node:           node,
		bootedImageRef: bootedImageRef,
		client:         kubeClient,
	}

	err = d.updateFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// Retries the given function until it succeeds
func (d *daemon) retry(f func() bool) {
	for !f() {
		select {
		case <-d.ctx.Done():
			return
		case <-time.After(d.RetryInterval()):
		}
	}
}

// Will try to release the lock until successful
func (d *daemon) releaseLock() {
	d.retry(func() bool {
		err := d.Fleetlock().Release()
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
	var wg sync.WaitGroup
	wg.Add(3)

	err := d.NewConfigFileWatcher()
	if err != nil {
		return fmt.Errorf("failed to create config file watcher: %v", err)
	}
	defer d.configWatcher.Close()
	go func() {
		defer wg.Done()
		d.WatchConfigFile()
		slog.Info("Stopped watching config file")
	}()

	err = d.rpmostree.RegisterAsDriver()
	if err != nil {
		return fmt.Errorf("failed to register upgraded as driver for rpm-ostree: %v", err)
	}

	node, err := d.getNode()
	if err != nil {
		return fmt.Errorf("failed to get node status: %v", err)
	}

	node, err = d.annotateNodeWithUpgradedVersion(node)
	if err != nil {
		return fmt.Errorf("failed to annotate node with upgraded version: %v", err)
	}

	if !nodeNeedsUpgrade(node) && d.nodeHasCorrectStream(node) {
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
