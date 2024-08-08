package daemon

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/heathcliff26/kube-upgrade/pkg/drain"
	"github.com/heathcliff26/kube-upgrade/pkg/lock"
	rpmostree "github.com/heathcliff26/kube-upgrade/pkg/rpm-ostree"
	"github.com/heathcliff26/kube-upgrade/pkg/utils"
)

const (
	defaultCheckIntervall = time.Hour * 3
	defaultRetryIntervall = time.Minute * 5
)

type daemon struct {
	lockUtility *lock.LockUtility
	drainer     *drain.Drainer

	node  string
	group string

	checkIntervall time.Duration
	retryIntervall time.Duration
}

// Create a new daemon
func NewDaemon(name, namespace, group, node string) (*daemon, error) {
	client, err := utils.CreateNewClientset("")
	if err != nil {
		return nil, err
	}

	l, err := lock.NewLockUtility(client, name, namespace)
	if err != nil {
		return nil, err
	}

	d, err := drain.NewDrainer(client)
	if err != nil {
		return nil, err
	}

	return &daemon{
		lockUtility:    l,
		drainer:        d,
		group:          group,
		node:           node,
		checkIntervall: defaultCheckIntervall,
		retryIntervall: defaultRetryIntervall,
	}, nil
}

// Checks if an initial cleanup from a previous instance is necessary
func (d *daemon) initialCleanup() error {
	ok, err := d.lockUtility.HasLock(d.group, d.group)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return d.cleanup()
}

// Cleans up after an upgrade by uncordoning the node and releasing the lease
func (d *daemon) cleanup() error {
	err := d.drainer.UncordonNode(d.node)
	if err != nil {
		return err
	}

	return d.lockUtility.Release(d.group)
}

// Run the main daemon loop
func (d *daemon) Run() error {
	err := d.initialCleanup()
	if err != nil {
		return err
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	intervall := d.checkIntervall
	firstTime := true

	for {
		if !firstTime {
			select {
			case <-stop:
				return nil
			case <-time.After(intervall):
			}
		} else {
			firstTime = false
		}

		needUpgrade, err := rpmostree.CheckForUpgrade()
		if err != nil {
			slog.Error("Failed to check if there is a new upgrade", "err", err)
			intervall = d.retryIntervall
			continue
		}

		if needUpgrade {
			ok, err := d.doUpgrade()
			if err != nil {
				slog.Error("Failed to perform rpm-ostree upgrade", "err", err)
				intervall = d.retryIntervall
				continue
			}
			if !ok {
				intervall = d.retryIntervall
				continue
			}
		}
		intervall = d.checkIntervall
	}
}

// Perform rpm-ostree upgrade
func (d *daemon) doUpgrade() (bool, error) {
	ok, err := d.lockUtility.Lock(d.group, d.node)
	if err != nil {
		return false, err
	} else if !ok {
		slog.Info("Could not lock the resource, another process already has it")
		return false, nil
	}

	err = d.drainer.DrainNode(d.node)
	if err != nil {
		return false, err
	}

	err = rpmostree.Upgrade()
	if err != nil {
		return false, err
	}

	// This should not be reached, as rpmostree.Upgrade() reboots the node on success.
	// I included it here mainly for completness sake.

	err = d.cleanup()
	if err != nil {
		return false, err
	}
	return true, nil
}
