package daemon

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	fleetlock "github.com/heathcliff26/fleetlock/pkg/client"
	rpmostree "github.com/heathcliff26/kube-upgrade/pkg/upgraded/rpm-ostree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestWatchForUpgrade(t *testing.T) {
	tMatrix := []struct {
		Name            string
		CheckExitCode   string
		UpgradeExitCode string
		ExpectedCalls   string
	}{
		{
			Name:          "NoUpgrade",
			ExpectedCalls: "upgrade --check\n",
		},
		{
			Name:          "Upgrade",
			CheckExitCode: "0",
			ExpectedCalls: "upgrade --check\nupgrade --reboot\n",
		},
		{
			Name:            "UpgradeFail",
			CheckExitCode:   "0",
			UpgradeExitCode: "1",
			ExpectedCalls:   "upgrade --check\nupgrade --reboot\nupgrade --reboot\n",
		},
		{
			Name:          "CheckFail",
			CheckExitCode: "1",
			ExpectedCalls: "upgrade --check\nupgrade --check\n",
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			require := require.New(t)

			logfile := filepath.Join(t.TempDir(), "args.log")
			t.Setenv("RPM_OSTREE_LOG_FILE", logfile)

			if tCase.CheckExitCode != "" {
				t.Setenv("RPM_OSTREE_CHECK_EXIT_CODE", tCase.CheckExitCode)
			}
			if tCase.UpgradeExitCode != "" {
				t.Setenv("RPM_OSTREE_UPGRADE_EXIT_CODE", tCase.UpgradeExitCode)
			}

			rpmOstreeCMD, err := rpmostree.New("testdata/rpm-ostree.sh")
			require.NoError(err, "Failed to create rpm-ostree command")

			client, srv := NewFakeFleetlockServer(t, http.StatusOK)
			t.Cleanup(srv.Close)

			ctx, cancel := context.WithCancel(context.Background())
			d := &daemon{
				rpmostree:     rpmOstreeCMD,
				checkInterval: 300 * time.Millisecond,
				retryInterval: 100 * time.Millisecond,
				ctx:           ctx,
				cancel:        cancel,
				fleetlock:     client,
			}

			done := make(chan struct{})
			go func() {
				d.watchForUpgrade()
				close(done)
			}()

			// Wait for test to run
			time.Sleep(150 * time.Millisecond)

			d.cancel()
			select {
			case <-done:
			case <-time.After(2 * d.checkInterval):
				t.Fatalf("Timed out waiting for watchForUpgrade to exit")
			}

			buf, err := os.ReadFile(logfile)
			require.NoError(err, "Failed to read logfile")

			require.Equal(tCase.ExpectedCalls, string(buf), "Should have made the expected calls")
		})
	}
}

func TestDoUpgrade(t *testing.T) {
	fakeDaemon := func(fleetlock *fleetlock.FleetlockClient, rpmostree *rpmostree.RPMOStreeCMD) *daemon {
		node := &corev1.Node{
			Name: "testnode",
		}
		return &daemon{
			fleetlock: fleetlock,
			rpmostree: rpmostree,
			node:      node.GetName(),
			client:    fake.NewClientset(node),
		}
	}

	t.Run("LockAlreadyReserved", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		client, srv := NewFakeFleetlockServer(t, http.StatusLocked)
		t.Cleanup(func() {
			srv.Close()
		})
		rpmOstreeCMD, err := rpmostree.New("testdata/exit-0.sh")
		require.NoError(err, "Failed to create rpm-ostree command")

		d := fakeDaemon(client, rpmOstreeCMD)

		err = d.doUpgrade()

		assert.ErrorContains(err, "failed to acquire lock:")
	})
	t.Run("FailedOstreeUpgrade", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		client, srv := NewFakeFleetlockServer(t, http.StatusOK)
		t.Cleanup(func() {
			srv.Close()
		})
		rpmOstreeCMD, err := rpmostree.New("testdata/exit-1.sh")
		require.NoError(err, "Failed to create rpm-ostree command")

		d := fakeDaemon(client, rpmOstreeCMD)

		err = d.doUpgrade()

		assert.Error(err, "Should exit with error")
	})
	// This case is kinda sketchy, as in reality the system would reboot on success, thus the method should never return
	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		client, srv := NewFakeFleetlockServer(t, http.StatusOK)
		t.Cleanup(func() {
			srv.Close()
		})
		rpmOstreeCMD, err := rpmostree.New("testdata/exit-0.sh")
		require.NoError(err, "Failed to create rpm-ostree command")

		d := fakeDaemon(client, rpmOstreeCMD)

		err = d.doUpgrade()

		assert.NoError(err, "Should succeed")
	})
}
