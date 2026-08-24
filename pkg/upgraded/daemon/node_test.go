package daemon

import (
	"context"
	"net/http"
	"testing"
	"time"

	fleetlock "github.com/heathcliff26/fleetlock/pkg/client"
	fleetfake "github.com/heathcliff26/fleetlock/pkg/fake"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/kubeadm"
	rpmostree "github.com/heathcliff26/kube-upgrade/pkg/upgraded/rpm-ostree"
	"github.com/heathcliff26/kube-upgrade/pkg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDoNodeUpgradeWithRetry(t *testing.T) {
	d := &daemon{
		client:        fake.NewClientset(),
		node:          "not-a-node",
		ctx:           t.Context(),
		retryInterval: time.Millisecond,
	}

	done := make(chan struct{}, 1)
	go func() {
		// Should not panic with nil
		d.doNodeUpgradeWithRetry(nil)
		done <- struct{}{}
	}()

	select {
	case <-done:
		t.Fatal("The upgrade should not succeed")
	case <-time.After(time.Millisecond * 20):
	}
}

func TestDoNodeUpgrade(t *testing.T) {
	t.Run("LockAlreadyReserved", func(t *testing.T) {
		assert := assert.New(t)

		client, srv := NewFakeFleetlockServer(t, http.StatusLocked)
		t.Cleanup(func() {
			srv.Close()
		})

		d := &daemon{
			fleetlock: client,
		}
		node := &corev1.Node{
			Name: "testnode",
			Annotations: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
			},
		}

		err := d.doNodeUpgrade(node)

		assert.ErrorContains(err, "failed to acquire lock:")
	})
	t.Run("FailedOstreeRebase", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		rpmOstreeCMD, err := rpmostree.New("testdata/exit-1.sh")
		require.NoError(err, "Failed to create rpm-ostree command")

		d, node := newTestDoNodeUpgradeSetup(t, constants.NodeUpgradeStatusRebasing)
		d.rpmostree = rpmOstreeCMD

		err = d.doNodeUpgrade(node)

		node, _ = d.client.CoreV1().Nodes().Get(t.Context(), node.GetName(), metav1.GetOptions{})

		assert.Error(err, "Should exit with error")
		assert.Equal(constants.NodeUpgradeStatusError, node.Annotations[constants.NodeUpgradeStatus], "Should have set correct node status")
	})
	t.Run("SucceededOstreeRebase", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		rpmOstreeCMD, err := rpmostree.New("testdata/exit-0.sh")
		require.NoError(err, "Failed to create rpm-ostree command")

		d, node := newTestDoNodeUpgradeSetup(t, constants.NodeUpgradeStatusRebasing)
		d.rpmostree = rpmOstreeCMD

		err = d.doNodeUpgrade(node)

		node, _ = d.client.CoreV1().Nodes().Get(t.Context(), node.GetName(), metav1.GetOptions{})

		assert.NoError(err, "Should exit without error")
		assert.Equal(constants.NodeUpgradeStatusRebasing, node.Annotations[constants.NodeUpgradeStatus], "Should have set correct node status")
	})
	t.Run("FailedKubeadmUpgrade", func(t *testing.T) {
		assert := assert.New(t)

		oldHostPrefix := hostPrefix
		hostPrefix = ""
		oldCosignBinary := kubeadm.CosignBinary
		kubeadm.CosignBinary = "../../../bin/cosign"
		t.Cleanup(func() {
			hostPrefix = oldHostPrefix
			kubeadm.CosignBinary = oldCosignBinary
		})

		d, node := newTestDoNodeUpgradeSetup(t, constants.NodeUpgradeStatusPending)

		err := d.doNodeUpgrade(node)

		node, _ = d.client.CoreV1().Nodes().Get(t.Context(), node.GetName(), metav1.GetOptions{})

		assert.ErrorContains(err, "failed to fetch kubeadm-config", "Should fail to fetch kubeadm config map")
		assert.Equal(constants.NodeUpgradeStatusError, node.Annotations[constants.NodeUpgradeStatus], "Should have set correct node status")
	})
}

func TestUpdateNodeStatus(t *testing.T) {
	tMatrix := []struct {
		Name  string
		Node  *corev1.Node
		Error bool
	}{
		{
			Name: "Success",
			Node: &corev1.Node{
				Name: "testnode",
				Annotations: map[string]string{
					constants.NodeUpgradeStatus: "unset",
				},
			},
		},
		{
			Name: "NoAnnotations",
			Node: &corev1.Node{
				Name: "testnode",
			},
		},
		{
			Name:  "NoNode",
			Error: true,
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := t.Context()

			c := fake.NewClientset()
			d := &daemon{
				client: c,
				ctx:    ctx,
			}
			if tCase.Node != nil {
				_, _ = c.CoreV1().Nodes().Create(ctx, tCase.Node, metav1.CreateOptions{})
				d.node = tCase.Node.GetName()
			} else {
				d.node = "not-a-node"
			}

			if tCase.Error {
				assert.Error(d.updateNodeStatus("new-status"), "Should fail")
			} else {
				assert.NoError(d.updateNodeStatus("new-status"), "Should succeed")
				node, _ := c.CoreV1().Nodes().Get(ctx, d.node, metav1.GetOptions{})
				assert.Equal("new-status", node.GetAnnotations()[constants.NodeUpgradeStatus], "Should have set status")
			}
		})
	}
}

func TestAnnotateNodeWithUpgradedVersion(t *testing.T) {
	ctx := t.Context()
	node := &corev1.Node{
		Name: "testnode",
	}
	d := &daemon{
		ctx:    ctx,
		client: fake.NewClientset(node),
		node:   node.GetName(),
	}

	assert := assert.New(t)

	node, err := d.annotateNodeWithUpgradedVersion(node)
	assert.NoError(err, "Should set version when no annotations are set")
	assert.Equal(version.Version(), node.Annotations[constants.NodeUpgradedVersion], "Should return updated node when no annotations are set")
	node, _ = d.client.CoreV1().Nodes().Get(ctx, node.GetName(), metav1.GetOptions{})
	assert.Equal(version.Version(), node.Annotations[constants.NodeUpgradedVersion], "Should set version when no annotations are set")

	node.Annotations[constants.NodeUpgradedVersion] = "old-version"
	node, _ = d.client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})

	node, err = d.annotateNodeWithUpgradedVersion(node)
	assert.NoError(err, "Should update version")
	assert.Equal(version.Version(), node.Annotations[constants.NodeUpgradedVersion], "Should return updated node")
	node, _ = d.client.CoreV1().Nodes().Get(ctx, node.GetName(), metav1.GetOptions{})
	assert.Equal(version.Version(), node.Annotations[constants.NodeUpgradedVersion], "Should update version")

	_ = d.client.CoreV1().Nodes().Delete(ctx, node.GetName(), metav1.DeleteOptions{})
	_, err = d.annotateNodeWithUpgradedVersion(node)
	assert.NoError(err, "Should not update node when version already matches")
}

func TestNodeHasCorrectStream(t *testing.T) {
	tMatrix := []struct {
		Name           string
		Version        string
		Stream         string
		BootedImageRef string
		Result         bool
	}{
		{
			Name:           "CorrectStream",
			Version:        "v1.34.2",
			Stream:         "registry.example.com/fcos-k8s",
			BootedImageRef: "ostree-unverified-registry:registry.example.com/fcos-k8s:v1.34.2",
			Result:         true,
		},
		{
			Name:           "WrongImage",
			Version:        "v1.34.2",
			Stream:         "registry.example.org/fcos-k8s",
			BootedImageRef: "ostree-unverified-registry:registry.example.com/fcos-k8s:v1.34.2",
		},
		{
			Name:           "WrongVersion",
			Version:        "v1.34.2",
			Stream:         "registry.example.com/fcos-k8s",
			BootedImageRef: "ostree-unverified-registry:registry.example.com/fcos-k8s:v1.33.5",
		},
		{
			Name:           "MissingVersionAnnotation",
			Version:        "v1.34.2",
			Stream:         "registry.example.com/fcos-k8s",
			BootedImageRef: "ostree-unverified-registry:registry.example.com/fcos-k8s:v1.34.2",
			Result:         true,
		},
	}
	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			d := &daemon{
				bootedImageRef: tCase.BootedImageRef,
				stream:         tCase.Stream,
			}
			node := &corev1.Node{
				Name: "testnode",
				Annotations: map[string]string{
					constants.NodeKubernetesVersion: tCase.Version,
				},
			}

			assert.Equal(tCase.Result, d.nodeHasCorrectStream(node), "Should return correct result")
		})
	}
}

func newTestDoNodeUpgradeSetup(t *testing.T, nodeStatus string) (*daemon, *corev1.Node) {
	t.Helper()

	client, srv := NewFakeFleetlockServer(t, http.StatusOK)
	t.Cleanup(func() {
		srv.Close()
	})

	node := &corev1.Node{
		Name: "testnode",
		Annotations: map[string]string{
			constants.NodeKubernetesVersion: "v1.35.0",
			constants.NodeUpgradeStatus:     nodeStatus,
		},
	}

	d := &daemon{
		ctx:       t.Context(),
		fleetlock: client,
		client:    fake.NewClientset(node),
		node:      node.GetName(),
	}

	return d, node
}

func TestNodeKubeadmUpgrade(t *testing.T) {
	oldK8sTmpDir := kubernetesTMPDir
	kubernetesTMPDir = "/not/a/valid/path"
	t.Cleanup(func() {
		kubernetesTMPDir = oldK8sTmpDir
	})

	t.Run("SuccessApply", func(t *testing.T) {
		assert := assert.New(t)

		node := &corev1.Node{
			Name: "testnode",
			Annotations: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
			},
		}

		configMap := &corev1.ConfigMap{
			Name:      "kubeadm-config",
			Namespace: "kube-system",
			Data: map[string]string{
				"ClusterConfiguration": "kubernetesVersion: v1.30.4\n",
			},
		}

		d, _ := newTestNodeKubeadmUpgradeSetup(t, "testdata/fake-kubeadm.sh", node, configMap)

		err := d.nodeKubeadmUpgrade("v1.31.0")

		assert.NoError(err)
		node, _ = d.client.CoreV1().Nodes().Get(t.Context(), "testnode", metav1.GetOptions{})
		assert.Equal(constants.NodeUpgradeStatusUpgrading, node.Annotations[constants.NodeUpgradeStatus])
	})

	t.Run("SuccessNode", func(t *testing.T) {
		assert := assert.New(t)

		node := &corev1.Node{
			Name: "testnode",
			Annotations: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
			},
		}

		configMap := &corev1.ConfigMap{
			Name:      "kubeadm-config",
			Namespace: "kube-system",
			Data: map[string]string{
				"ClusterConfiguration": "kubernetesVersion: v1.31.0\n",
			},
		}

		d, _ := newTestNodeKubeadmUpgradeSetup(t, "testdata/fake-kubeadm.sh", node, configMap)

		err := d.nodeKubeadmUpgrade("v1.31.0")

		assert.NoError(err)
		node, _ = d.client.CoreV1().Nodes().Get(t.Context(), "testnode", metav1.GetOptions{})
		assert.Equal(constants.NodeUpgradeStatusUpgrading, node.Annotations[constants.NodeUpgradeStatus])
	})

	t.Run("FailedFetchConfigMap", func(t *testing.T) {
		assert := assert.New(t)

		node := &corev1.Node{
			Name: "testnode",
			Annotations: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
			},
		}

		d, _ := newTestNodeKubeadmUpgradeSetup(t, "testdata/fake-kubeadm.sh", node, nil)

		err := d.nodeKubeadmUpgrade("v1.31.0")

		assert.Error(err)
		assert.ErrorContains(err, "failed to fetch kubeadm-config")
		node, _ = d.client.CoreV1().Nodes().Get(t.Context(), "testnode", metav1.GetOptions{})
		assert.Equal(constants.NodeUpgradeStatusError, node.Annotations[constants.NodeUpgradeStatus])
	})

	t.Run("ConfigMapNoData", func(t *testing.T) {
		assert := assert.New(t)

		node := &corev1.Node{
			Name: "testnode",
			Annotations: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
			},
		}

		configMap := &corev1.ConfigMap{
			Name:      "kubeadm-config",
			Namespace: "kube-system",
		}

		d, _ := newTestNodeKubeadmUpgradeSetup(t, "testdata/fake-kubeadm.sh", node, configMap)

		err := d.nodeKubeadmUpgrade("v1.31.0")

		assert.Error(err)
		assert.ErrorContains(err, "kubeadm configmap contains no data")
		node, _ = d.client.CoreV1().Nodes().Get(t.Context(), "testnode", metav1.GetOptions{})
		assert.Equal(constants.NodeUpgradeStatusError, node.Annotations[constants.NodeUpgradeStatus])
	})

	t.Run("ConfigMapInvalidYAML", func(t *testing.T) {
		assert := assert.New(t)

		node := &corev1.Node{
			Name: "testnode",
			Annotations: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
			},
		}

		configMap := &corev1.ConfigMap{
			Name:      "kubeadm-config",
			Namespace: "kube-system",
			Data: map[string]string{
				"ClusterConfiguration": "not: valid: yaml: [",
			},
		}

		d, _ := newTestNodeKubeadmUpgradeSetup(t, "testdata/fake-kubeadm.sh", node, configMap)

		err := d.nodeKubeadmUpgrade("v1.31.0")

		assert.Error(err)
		assert.ErrorContains(err, "failed to parse kubeadm-config")
		node, _ = d.client.CoreV1().Nodes().Get(t.Context(), "testnode", metav1.GetOptions{})
		assert.Equal(constants.NodeUpgradeStatusError, node.Annotations[constants.NodeUpgradeStatus])
	})

	t.Run("KubeadmApplyFails", func(t *testing.T) {
		assert := assert.New(t)

		node := &corev1.Node{
			Name: "testnode",
			Annotations: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
			},
		}

		configMap := &corev1.ConfigMap{
			Name:      "kubeadm-config",
			Namespace: "kube-system",
			Data: map[string]string{
				"ClusterConfiguration": "kubernetesVersion: v1.30.4\n",
			},
		}

		d, _ := newTestNodeKubeadmUpgradeSetup(t, "testdata/fake-kubeadm-fail.sh", node, configMap)

		err := d.nodeKubeadmUpgrade("v1.31.0")

		assert.Error(err)
		assert.ErrorContains(err, "failed run kubeadm")
		node, _ = d.client.CoreV1().Nodes().Get(t.Context(), "testnode", metav1.GetOptions{})
		assert.Equal(constants.NodeUpgradeStatusError, node.Annotations[constants.NodeUpgradeStatus])
	})

	t.Run("KubeadmNodeFails", func(t *testing.T) {
		assert := assert.New(t)

		node := &corev1.Node{
			Name: "testnode",
			Annotations: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
			},
		}

		configMap := &corev1.ConfigMap{
			Name:      "kubeadm-config",
			Namespace: "kube-system",
			Data: map[string]string{
				"ClusterConfiguration": "kubernetesVersion: v1.31.0\n",
			},
		}

		d, _ := newTestNodeKubeadmUpgradeSetup(t, "testdata/fake-kubeadm-fail.sh", node, configMap)

		err := d.nodeKubeadmUpgrade("v1.31.0")

		assert.Error(err)
		assert.ErrorContains(err, "failed run kubeadm")
		node, _ = d.client.CoreV1().Nodes().Get(t.Context(), "testnode", metav1.GetOptions{})
		assert.Equal(constants.NodeUpgradeStatusError, node.Annotations[constants.NodeUpgradeStatus])
	})
}

func newTestNodeKubeadmUpgradeSetup(t *testing.T, kubeadmPath string, node *corev1.Node, configMap *corev1.ConfigMap) (*daemon, *corev1.Node) {
	t.Helper()
	require := require.New(t)

	client := fake.NewClientset()
	if node != nil {
		_, err := client.CoreV1().Nodes().Create(t.Context(), node, metav1.CreateOptions{})
		require.NoError(err, "Failed to create node")
	}
	if configMap != nil {
		_, err := client.CoreV1().ConfigMaps("kube-system").Create(t.Context(), configMap, metav1.CreateOptions{})
		require.NoError(err, "Failed to create configmap")
	}

	kubeadmCMD, err := kubeadm.NewFromPath("", kubeadmPath)
	require.NoError(err, "Failed to create kubeadm command")

	d := &daemon{
		ctx:     t.Context(),
		client:  client,
		node:    "testnode",
		kubeadm: kubeadmCMD,
	}

	if node != nil {
		return d, node
	}
	return d, nil
}

func TestWatchForNodeUpgrade(t *testing.T) {
	t.Run("NodeUpdateTriggersUpgrade", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		client := fake.NewClientset()

		node := &corev1.Node{
			Name: "testnode",
			Annotations: map[string]string{
				constants.NodeKubernetesVersion: "v1.31.0",
				constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusPending,
			},
		}
		_, err := client.CoreV1().Nodes().Create(t.Context(), node, metav1.CreateOptions{})
		require.NoError(err)

		configMap := &corev1.ConfigMap{
			Name:      "kubeadm-config",
			Namespace: "kube-system",
			Data: map[string]string{
				"ClusterConfiguration": "kubernetesVersion: v1.30.4\n",
			},
		}
		_, err = client.CoreV1().ConfigMaps("kube-system").Create(t.Context(), configMap, metav1.CreateOptions{})
		require.NoError(err)

		srv := fleetfake.NewFakeServer(t, http.StatusOK, "")
		srv.Group = "default"
		t.Cleanup(func() {
			srv.Close()
		})

		fleetlockClient, err := fleetlock.NewClient(srv.URL(), "default")
		require.NoError(err)

		rpmOstreeCMD, err := rpmostree.New("testdata/exit-0.sh")
		require.NoError(err)

		kubeadmCMD, err := kubeadm.NewFromPath("", "testdata/fake-kubeadm.sh")
		require.NoError(err)

		ctx, cancel := context.WithCancel(t.Context())
		t.Cleanup(cancel)

		d := &daemon{
			ctx:            ctx,
			cancel:         cancel,
			client:         client,
			node:           "testnode",
			fleetlock:      fleetlockClient,
			rpmostree:      rpmOstreeCMD,
			kubeadm:        kubeadmCMD,
			bootedImageRef: "ostree-unverified-registry:registry.example.com/fcos-k8s:v1.31.0",
			stream:         "registry.example.com/fcos-k8s",
			retryInterval:  time.Millisecond,
		}

		go func() {
			d.watchForNodeUpgrade()
		}()

		time.Sleep(100 * time.Millisecond)

		node, _ = client.CoreV1().Nodes().Get(t.Context(), "testnode", metav1.GetOptions{})
		node.Annotations[constants.NodeUpgradeStatus] = constants.NodeUpgradeStatusPending
		_, err = client.CoreV1().Nodes().Update(t.Context(), node, metav1.UpdateOptions{})
		require.NoError(err)

		assert.Eventually(func() bool {
			n, err := client.CoreV1().Nodes().Get(t.Context(), "testnode", metav1.GetOptions{})
			if err != nil {
				return false
			}
			return n.Annotations[constants.NodeUpgradeStatus] == constants.NodeUpgradeStatusCompleted
		}, time.Second*5, time.Millisecond*10)

		cancel()
	})

	t.Run("ContextCancellationStopsInformer", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		t.Cleanup(cancel)

		d := &daemon{
			ctx:           ctx,
			cancel:        cancel,
			client:        fake.NewClientset(),
			node:          "testnode",
			retryInterval: time.Millisecond,
		}

		done := make(chan struct{})
		go func() {
			defer close(done)
			d.watchForNodeUpgrade()
		}()

		time.Sleep(100 * time.Millisecond)

		cancel()

		select {
		case <-done:
			// Success
		case <-time.After(time.Second):
			t.Fatal("watchForNodeUpgrade did not stop after context cancellation")
		}
	})
}
