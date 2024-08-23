package daemon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	fleetlockclient "github.com/heathcliff26/fleetlock/pkg/server/client"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/config"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/fleetlock"
	rpmostree "github.com/heathcliff26/kube-upgrade/pkg/upgraded/rpm-ostree"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewDaemon(t *testing.T) {
	tMatrix := []struct {
		Name  string
		CFG   config.Config
		Error string
	}{
		{
			Name: "NoNodeFound",
			CFG: config.Config{
				Kubeconfig: "testdata/kubeconfig",
				Image:      config.DEFAULT_IMAGE,
				Fleetlock: config.FleetlockConfig{
					URL:   "https://fleetlock.example.com",
					Group: config.DEFAULT_FLEETLOCK_GROUP,
				},
				RPMOStreePath: "testdata/exit-0.sh",
				KubeadmPath:   "testdata/exit-0.sh",
				CheckInterval: config.DEFAULT_CHECK_INTERVAL,
				RetryInterval: config.DEFAULT_RETRY_INTERVAL,
			},
			Error: "failed to get kubernetes node name for host",
		},
		{
			Name: "NoFleetlockUrl",
			CFG: config.Config{
				Kubeconfig: "testdata/kubeconfig",
				Image:      config.DEFAULT_IMAGE,
				Fleetlock: config.FleetlockConfig{
					URL:   "",
					Group: config.DEFAULT_FLEETLOCK_GROUP,
				},
				RPMOStreePath: "testdata/exit-0.sh",
				KubeadmPath:   "testdata/exit-0.sh",
				CheckInterval: config.DEFAULT_CHECK_INTERVAL,
				RetryInterval: config.DEFAULT_RETRY_INTERVAL,
			},
			Error: "failed to create fleetlock client:",
		},
		{
			Name: "NoRPMOstree",
			CFG: config.Config{
				Kubeconfig: "testdata/kubeconfig",
				Image:      config.DEFAULT_IMAGE,
				Fleetlock: config.FleetlockConfig{
					URL:   "https://fleetlock.example.com",
					Group: config.DEFAULT_FLEETLOCK_GROUP,
				},
				RPMOStreePath: "",
				KubeadmPath:   "testdata/exit-0.sh",
				CheckInterval: config.DEFAULT_CHECK_INTERVAL,
				RetryInterval: config.DEFAULT_RETRY_INTERVAL,
			},
			Error: "failed to create rpm-ostree cmd wrapper:",
		},
		{
			Name: "NoKubeadm",
			CFG: config.Config{
				Kubeconfig: "testdata/kubeconfig",
				Image:      config.DEFAULT_IMAGE,
				Fleetlock: config.FleetlockConfig{
					URL:   "https://fleetlock.example.com",
					Group: config.DEFAULT_FLEETLOCK_GROUP,
				},
				RPMOStreePath: "testdata/exit-0.sh",
				KubeadmPath:   "",
				CheckInterval: config.DEFAULT_CHECK_INTERVAL,
				RetryInterval: config.DEFAULT_RETRY_INTERVAL,
			},
			Error: "failed to create kubeadm cmd wrapper:",
		},
		{
			Name: "EmptyKubeconfig",
			CFG: config.Config{
				Kubeconfig: "",
				Image:      config.DEFAULT_IMAGE,
				Fleetlock: config.FleetlockConfig{
					URL:   "https://fleetlock.example.com",
					Group: config.DEFAULT_FLEETLOCK_GROUP,
				},
				RPMOStreePath: "testdata/exit-0.sh",
				KubeadmPath:   "testdata/exit-0.sh",
				CheckInterval: config.DEFAULT_CHECK_INTERVAL,
				RetryInterval: config.DEFAULT_RETRY_INTERVAL,
			},
			Error: "no kubeconfig provided",
		},
		{
			Name: "KubeconfigFileNotFound",
			CFG: config.Config{
				Kubeconfig: "not-a-file",
				Image:      config.DEFAULT_IMAGE,
				Fleetlock: config.FleetlockConfig{
					URL:   "https://fleetlock.example.com",
					Group: config.DEFAULT_FLEETLOCK_GROUP,
				},
				RPMOStreePath: "testdata/exit-0.sh",
				KubeadmPath:   "testdata/exit-0.sh",
				CheckInterval: config.DEFAULT_CHECK_INTERVAL,
				RetryInterval: config.DEFAULT_RETRY_INTERVAL,
			},
			Error: "failed to read kubeconfig:",
		},
		{
			Name: "NoImage",
			CFG: config.Config{
				Kubeconfig: "testdata/kubeconfig",
				Image:      "",
				Fleetlock: config.FleetlockConfig{
					URL:   "https://fleetlock.example.com",
					Group: config.DEFAULT_FLEETLOCK_GROUP,
				},
				RPMOStreePath: "testdata/exit-0.sh",
				KubeadmPath:   "testdata/exit-0.sh",
				CheckInterval: config.DEFAULT_CHECK_INTERVAL,
				RetryInterval: config.DEFAULT_RETRY_INTERVAL,
			},
			Error: "no image provided for kubernetes updates",
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			d, err := NewDaemon(&tCase.CFG)

			assert.ErrorContains(err, tCase.Error, "Should return the given error")
			assert.Nil(d, "Should not return a daemon")
		})
	}
}

func TestRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	d := &daemon{
		retryInterval: time.Millisecond,
		ctx:           ctx,
	}

	cancelOnTimeout(t, ctx, cancel)

	count := 0
	d.retry(func() bool {
		count++
		return count > 5
	})
	t.Cleanup(cancel)

	assert.Equal(t, 6, count, "Should have run the function exactly 6 times")
}

func TestDoUpgrade(t *testing.T) {
	t.Run("LockAlreadyReserved", func(t *testing.T) {
		assert := assert.New(t)

		client, srv := NewFakeFleetlockServer(t, http.StatusLocked)
		t.Cleanup(func() {
			srv.Close()
		})
		rpmOstreeCMD, err := rpmostree.New("testdata/exit-0.sh")
		if !assert.NoError(err, "Failed to create rpm-ostree command") {
			t.FailNow()
		}

		d := &daemon{
			fleetlock: client,
			rpmostree: rpmOstreeCMD,
		}

		err = d.doUpgrade()

		assert.ErrorContains(err, "failed to aquire lock:")
	})
	t.Run("FailedOstreeUpgrade", func(t *testing.T) {
		assert := assert.New(t)

		client, srv := NewFakeFleetlockServer(t, http.StatusOK)
		t.Cleanup(func() {
			srv.Close()
		})
		rpmOstreeCMD, err := rpmostree.New("testdata/exit-1.sh")
		if !assert.NoError(err, "Failed to create rpm-ostree command") {
			t.FailNow()
		}

		d := &daemon{
			fleetlock: client,
			rpmostree: rpmOstreeCMD,
		}

		err = d.doUpgrade()

		assert.Error(err, "Should exit with error")
	})
	// This case is kinda scetchy, as in reality the system would reboot on success, thus the method should never return
	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)

		client, srv := NewFakeFleetlockServer(t, http.StatusOK)
		t.Cleanup(func() {
			srv.Close()
		})
		rpmOstreeCMD, err := rpmostree.New("testdata/exit-0.sh")
		if !assert.NoError(err, "Failed to create rpm-ostree command") {
			t.FailNow()
		}

		d := &daemon{
			fleetlock: client,
			rpmostree: rpmOstreeCMD,
		}

		err = d.doUpgrade()

		assert.NoError(err, "Should succeed")
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
						constants.KubernetesUpgradeStatus: "unset",
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

			c := fake.NewSimpleClientset()
			d := &daemon{
				client: c,
				ctx:    context.Background(),
			}
			if tCase.Node != nil {
				_, _ = c.CoreV1().Nodes().Create(context.Background(), tCase.Node, metav1.CreateOptions{})
				d.node = tCase.Node.GetName()
			} else {
				d.node = "not-a-node"
			}

			if tCase.Error {
				assert.Error(d.updateNodeStatus("new-status"), "Should fail")
			} else {
				assert.NoError(d.updateNodeStatus("new-status"), "Should succeed")
				node, _ := c.CoreV1().Nodes().Get(context.Background(), d.node, metav1.GetOptions{})
				assert.Equal("new-status", node.GetAnnotations()[constants.KubernetesUpgradeStatus], "Should have set status")
			}
		})
	}
}

func cancelOnTimeout(t *testing.T, ctx context.Context, cancel context.CancelFunc) {
	go func() {
		select {
		case <-time.After(time.Second * 5):
			t.Fail()
			t.Log("Timeout waiting for retry to succeeed")
			cancel()
		case <-ctx.Done():
		}
	}()
}

func NewFakeFleetlockServer(t *testing.T, statusCode int) (*fleetlock.FleetlockClient, *httptest.Server) {
	assert := assert.New(t)

	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(statusCode)
		b, err := json.MarshalIndent(fleetlockclient.FleetLockResponse{
			Kind:  "ok",
			Value: "Success",
		}, "", "  ")
		if !assert.NoError(err, "Error in fake server: failed to prepare response") {
			return
		}

		_, err = rw.Write(b)
		assert.NoError(err, "Error in fake server: failed to send response")
	}))
	c, err := fleetlock.NewClient(srv.URL, "default")
	assert.NoError(err, "Error in creating fake server: failed to create client")
	return c, srv
}
