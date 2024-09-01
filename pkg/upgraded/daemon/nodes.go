package daemon

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/kubeadm"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

// Watch for node upgrades and preform them if necessary
func (d *daemon) watchForNodeUpgrade() {
	factory := informers.NewSharedInformerFactoryWithOptions(d.client, time.Minute, informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.FieldSelector = fields.SelectorFromSet(fields.Set{"metadata.name": d.node}).String()
	}))

	informer := factory.Core().V1().Nodes().Informer()
	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(_, newObj interface{}) {
			node := newObj.(*corev1.Node)
			d.checkNodeStatus(node)
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
func (d *daemon) checkNodeStatus(node *corev1.Node) {
	if !nodeNeedsUpgrade(node) {
		return
	}

	d.doNodeUpgradeWithRetry(nil)
}

// Update the node until it succeeds
func (d *daemon) doNodeUpgradeWithRetry(node *corev1.Node) {
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

	var err error
	if node == nil {
		// Need to fetch fresh data here, as the informer might called with a stale node version
		node, err = d.getNode()
		if err != nil {
			return fmt.Errorf("failed to get node data from server: %v", err)
		}
		if !nodeNeedsUpgrade(node) {
			return nil
		}
	}

	err = d.UpdateConfigFromAnnotations(node.GetAnnotations())
	if err != nil {
		return fmt.Errorf("failed to update daemon config from node annotations: %v", err)
	}

	version := node.Annotations[constants.NodeKubernetesVersion]
	slog.Info("Attempting node upgrade to new kubernetes version", slog.String("node", node.GetName()), slog.String("version", version))

	err = d.fleetlock.Lock()
	if err != nil {
		return fmt.Errorf("failed to aquire lock: %v", err)
	}

	if version != node.Status.NodeInfo.KubeletVersion {
		slog.Info("Rebasing os to new kubernetes version", slog.String("version", version))
		err := d.updateNodeStatus(constants.NodeUpgradeStatusRebasing)
		if err != nil {
			return fmt.Errorf("failed to update node status: %v", err)
		}
		err = d.rpmostree.Rebase(d.stream + ":" + version)
		if err != nil {
			return fmt.Errorf("failed to rebase node: %v", err)
		}
		// This return is here purely for testing, as a successfull rebase does not return, but instead reboots the system
		return nil
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
	node, err := d.getNode()
	if err != nil {
		return err
	}
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}
	node.Annotations[constants.NodeUpgradeStatus] = status

	_, err = d.client.CoreV1().Nodes().Update(d.ctx, node, metav1.UpdateOptions{})
	if err == nil {
		slog.Debug("Set node status", slog.String("status", status))
	}
	return err
}

// Retrieve the node from the API
func (d *daemon) getNode() (*corev1.Node, error) {
	return d.client.CoreV1().Nodes().Get(d.ctx, d.node, metav1.GetOptions{})
}
