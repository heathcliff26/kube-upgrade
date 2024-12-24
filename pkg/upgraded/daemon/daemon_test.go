package daemon

import (
	"context"
	"testing"
	"time"

	fleetlock "github.com/heathcliff26/fleetlock/pkg/client"
	"github.com/heathcliff26/fleetlock/pkg/fake"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/config"
	"github.com/stretchr/testify/assert"
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
				Kubeconfig:    "testdata/kubeconfig",
				RPMOStreePath: "testdata/exit-0.sh",
				KubeadmPath:   "testdata/exit-0.sh",
			},
			Error: "failed to get kubernetes node name for host",
		},
		{
			Name: "NoRPMOstree",
			CFG: config.Config{
				Kubeconfig:    "testdata/kubeconfig",
				RPMOStreePath: "",
				KubeadmPath:   "testdata/exit-0.sh",
			},
			Error: "failed to create rpm-ostree cmd wrapper:",
		},
		{
			Name: "NoKubeadm",
			CFG: config.Config{
				Kubeconfig:    "testdata/kubeconfig",
				RPMOStreePath: "testdata/exit-0.sh",
				KubeadmPath:   "",
			},
			Error: "failed to create kubeadm cmd wrapper:",
		},
		{
			Name: "EmptyKubeconfig",
			CFG: config.Config{
				Kubeconfig:    "",
				RPMOStreePath: "testdata/exit-0.sh",
				KubeadmPath:   "testdata/exit-0.sh",
			},
			Error: "no kubeconfig provided",
		},
		{
			Name: "KubeconfigFileNotFound",
			CFG: config.Config{
				Kubeconfig:    "not-a-file",
				RPMOStreePath: "testdata/exit-0.sh",
				KubeadmPath:   "testdata/exit-0.sh",
			},
			Error: "failed to read kubeconfig:",
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
