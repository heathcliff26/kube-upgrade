package daemon

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/kubeadm"
	rpmostree "github.com/heathcliff26/kube-upgrade/pkg/upgraded/rpm-ostree"
	"github.com/heathcliff26/kube-upgrade/pkg/version"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDoNodeUpgradeWithRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	d := &daemon{
		client:        fake.NewSimpleClientset(),
		node:          "not-a-node",
		ctx:           ctx,
		retryInterval: time.Second,
	}

	done := make(chan struct{}, 1)
	go func() {
		// Should not panic with nil
		d.doNodeUpgradeWithRetry(nil)
		done <- struct{}{}
	}()
	t.Cleanup(cancel)

	select {
	case <-done:
		t.Fatal("The upgrade should not succeed")
	case <-time.After(time.Second * 3):
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
			ObjectMeta: metav1.ObjectMeta{
				Name: "testnode",
				Annotations: map[string]string{
					constants.NodeKubernetesVersion: "v1.31.0",
				},
			},
		}

		err := d.doNodeUpgrade(node)

		assert.ErrorContains(err, "failed to acquire lock:")
	})
	for name, succeed := range map[string]bool{
		"FailedOstreeRebase":    false,
		"SucceededOstreeRebase": true,
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client, srv := NewFakeFleetlockServer(t, http.StatusOK)
			t.Cleanup(func() {
				srv.Close()
			})

			path := "testdata/exit-1.sh"
			if succeed {
				path = "testdata/exit-0.sh"
			}
			rpmOstreeCMD, err := rpmostree.New(path)
			if !assert.NoError(err, "Failed to create rpm-ostree command") {
				t.FailNow()
			}

			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			t.Cleanup(cancel)

			d := &daemon{
				ctx:       ctx,
				fleetlock: client,
				rpmostree: rpmOstreeCMD,
				client:    fake.NewSimpleClientset(),
				node:      "testnode",
			}

			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: d.node,
					Annotations: map[string]string{
						constants.NodeKubernetesVersion: "v1.31.0",
						constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusRebasing,
					},
				},
			}
			node, _ = d.client.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{})

			err = d.doNodeUpgrade(node)

			node, _ = d.client.CoreV1().Nodes().Get(ctx, node.GetName(), metav1.GetOptions{})

			if succeed {
				assert.NoError(err, "Should exit without error")
				assert.Equal(constants.NodeUpgradeStatusRebasing, node.Annotations[constants.NodeUpgradeStatus], "Should have set correct node status")
			} else {
				assert.Error(err, "Should exit with error")
				assert.Equal(constants.NodeUpgradeStatusError, node.Annotations[constants.NodeUpgradeStatus], "Should have set correct node status")
			}
		})
	}
	t.Run("FailedKubeadmUpgrade", func(t *testing.T) {
		assert := assert.New(t)

		oldHostPrefix := hostPrefix
		hostPrefix = ""
		oldCosignBinary := kubeadm.CosignBinary
		kubeadm.CosignBinary = "../../../bin/cosign"
		client, srv := NewFakeFleetlockServer(t, http.StatusOK)
		t.Cleanup(func() {
			hostPrefix = oldHostPrefix
			kubeadm.CosignBinary = oldCosignBinary
			srv.Close()
		})

		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		t.Cleanup(cancel)

		d := &daemon{
			ctx:       ctx,
			fleetlock: client,
			client:    fake.NewSimpleClientset(),
			node:      "testnode",
		}

		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: d.node,
				Annotations: map[string]string{
					constants.NodeKubernetesVersion: "v1.35.0",
					constants.NodeUpgradeStatus:     constants.NodeUpgradeStatusPending,
				},
			},
		}
		node, _ = d.client.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{})

		err := d.doNodeUpgrade(node)

		node, _ = d.client.CoreV1().Nodes().Get(ctx, node.GetName(), metav1.GetOptions{})

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
				ObjectMeta: metav1.ObjectMeta{
					Name: "testnode",
					Annotations: map[string]string{
						constants.NodeUpgradeStatus: "unset",
					},
				},
			},
		},
		{
			Name: "NoAnnotations",
			Node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testnode",
				},
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

			c := fake.NewSimpleClientset()
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
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	d := &daemon{
		ctx:    ctx,
		client: fake.NewSimpleClientset(),
		node:   "testnode",
	}

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: d.node,
		},
	}
	node, _ = d.client.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{})

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
				ObjectMeta: metav1.ObjectMeta{
					Name: "testnode",
					Annotations: map[string]string{
						constants.NodeKubernetesVersion: tCase.Version,
					},
				},
			}

			assert.Equal(tCase.Result, d.nodeHasCorrectStream(node), "Should return correct result")
		})
	}
}
