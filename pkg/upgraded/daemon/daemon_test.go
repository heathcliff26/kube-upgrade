package daemon

import (
	"context"
	"testing"
	"time"

	fleetlock "github.com/heathcliff26/fleetlock/pkg/client"
	"github.com/heathcliff26/fleetlock/pkg/fake"
	rpmostree "github.com/heathcliff26/kube-upgrade/pkg/upgraded/rpm-ostree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func TestNewDaemon(t *testing.T) {
	oldHostPrefix := hostPrefix
	hostPrefix = ""
	oldRPMOStreeCMDPath := rpmOstreeCMDPath
	rpmOstreeCMDPath = "testdata/exit-0.sh"
	t.Cleanup(func() {
		hostPrefix = oldHostPrefix
		rpmOstreeCMDPath = oldRPMOStreeCMDPath
	})

	tMatrix := []struct {
		Name, Path string
		Error      string
	}{
		{
			Name:  "NoNodeFound",
			Path:  "testdata/config/valid-config.yaml",
			Error: "failed to get kubernetes node name for host",
		},
		{
			Name:  "NoKubeadm",
			Path:  "testdata/config/invalid-kubeadm.yaml",
			Error: "failed to create kubeadm cmd wrapper:",
		},
		{
			Name:  "KubeconfigFileNotFound",
			Path:  "testdata/config/invalid-kubeletConfig.yaml",
			Error: "failed to read kubeconfig:",
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			d, err := NewDaemon(tCase.Path)

			assert.ErrorContains(err, tCase.Error, "Should return the given error")
			assert.Nil(d, "Should not return a daemon")
		})
	}
}

func TestRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
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

func TestRun(t *testing.T) {
	tMatrix := []struct {
		Name, RPMOStreePath string
		Error               string
	}{
		{
			Name:          "FailedToRegisterAsDriver",
			RPMOStreePath: "testdata/exit-1.sh",
			Error:         "failed to register upgraded as driver for rpm-ostree:",
		},
		{
			Name:          "FailedToGetNode",
			RPMOStreePath: "testdata/exit-0.sh",
			Error:         "failed to get node status:",
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			oldRPMOstreeCMDPath := rpmOstreeCMDPath
			t.Cleanup(func() {
				rpmOstreeCMDPath = oldRPMOstreeCMDPath
			})

			rpmOstreeCMD, err := rpmostree.New(tCase.RPMOStreePath)
			require.NoError(err, "Should create rpm-ostree cmd wrapper")
			config, err := clientcmd.BuildConfigFromFlags("", "testdata/kubeconfig")
			require.NoError(err, "Should read kubeconfig")
			kubeClient, err := kubernetes.NewForConfig(config)
			require.NoError(err, "Should create kubernetes client")

			d := &daemon{
				rpmostree: rpmOstreeCMD,
				client:    kubeClient,
			}

			result := make(chan error, 1)
			go func() {
				result <- d.Run()
			}()

			select {
			case err = <-result:
				assert.ErrorContains(err, tCase.Error, "Should return the given error")
			case <-time.After(time.Second * 5):
				t.Fatal("Timeout waiting for Run to return")
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

func NewFakeFleetlockServer(t *testing.T, statusCode int) (*fleetlock.FleetlockClient, *fake.FakeServer) {
	testGroup := "default"

	srv := fake.NewFakeServer(t, statusCode, "")
	srv.Group = testGroup

	c, err := fleetlock.NewClient(srv.URL(), "default")
	assert.NoError(t, err, "Error in creating fake server: failed to create client")
	return c, srv
}
