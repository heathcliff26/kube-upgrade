package daemon

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	fleetlock "github.com/heathcliff26/fleetlock/pkg/client"
	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/config"
)

// Reload the daemon configuration from it's config file
func (d *daemon) UpdateFromConfigFile() error {
	slog.Info("Attempting to update daemon configuration from config file", slog.String("path", d.cfgPath))

	cfg, err := config.LoadConfig(d.cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config from file: %v", err)
	}

	return d.updateFromConfig(cfg)
}

// Update the daemon configuration from the provided config object.
// Ensures that no changes will be made if there are errors in the config.
func (d *daemon) updateFromConfig(cfg *api.UpgradedConfig) error {
	checkInterval, err := time.ParseDuration(cfg.CheckInterval)
	if err != nil {
		return fmt.Errorf("failed to parse check interval \"%s\": %v", cfg.CheckInterval, err)
	}
	retryInterval, err := time.ParseDuration(cfg.RetryInterval)
	if err != nil {
		return fmt.Errorf("failed to parse retry interval \"%s\": %v", cfg.RetryInterval, err)
	}

	fleetlockClient, err := fleetlock.NewClient(cfg.FleetlockURL, cfg.FleetlockGroup)
	if err != nil {
		return fmt.Errorf("failed to create fleetlock client with url '%s' and group '%s': %v", cfg.FleetlockURL, cfg.FleetlockGroup, err)
	}

	d.Lock()
	defer d.Unlock()

	d.stream = cfg.Stream
	d.fleetlock = fleetlockClient
	d.checkInterval = checkInterval
	d.retryInterval = retryInterval

	slog.Info("Finished updating configuration")
	return nil
}

// Create a new config file watcher that needs to be closed when done
func (d *daemon) NewConfigFileWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create config file watcher: %v", err)
	}

	// Need to watch the directory instead of the file, as kubernetes uses symbolic links
	err = watcher.Add(filepath.Dir(d.cfgPath))
	if err != nil {
		return fmt.Errorf("failed to add config file to watcher: %v", err)
	}

	d.configWatcher = watcher
	return nil
}

func (d *daemon) WatchConfigFile() {
	slog.Info("Started watching the config file for changes", slog.String("path", d.cfgPath))

	for {
		select {
		case event, ok := <-d.configWatcher.Events:
			if !ok {
				slog.Info("Config file watcher events channel closed")
				return
			}
			// Ignore chmod, rename and remove events, they are not relevant
			if event.Has(fsnotify.Chmod | fsnotify.Rename | fsnotify.Remove) {
				continue
			}
			slog.Debug("Received event on config file directory", slog.String("Op", event.Op.String()), slog.String("name", event.Name))
			// Check if the event is for our config file or the ..data symlink
			if event.Name == d.cfgPath || event.Name == filepath.Join(filepath.Dir(d.cfgPath), "..data") {
				err := d.UpdateFromConfigFile()
				if err != nil {
					slog.Error("Failed to update configuration from config file", slog.String("path", d.cfgPath), slog.String("error", err.Error()))
				}
			}
		case err, ok := <-d.configWatcher.Errors:
			if !ok {
				slog.Info("Config file watcher errors channel closed")
				return
			}
			slog.Error("Error watching config file", slog.String("path", d.cfgPath), slog.String("error", err.Error()))
		case <-d.ctx.Done():
			// Daemon is stopping
			return
		}
	}
}
