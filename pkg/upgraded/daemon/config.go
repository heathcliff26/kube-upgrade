package daemon

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/heathcliff26/kube-upgrade/pkg/constants"
)

// Update the daemon configuration based on the annotations of the node.
// Returns on the first error, but will change all configs before that.
func (d *daemon) UpdateConfigFromNode() error {
	node, err := d.getNode()
	if err != nil {
		return err
	}
	return d.UpdateConfigFromAnnotations(node.GetAnnotations())
}

// Update the daemon configuration from the given annotations.
// Returns on the first error, but will change all configs before that.
func (d *daemon) UpdateConfigFromAnnotations(annotations map[string]string) error {
	for key, value := range annotations {
		switch key {
		case constants.ConfigStream:
			if value == "" {
				return fmt.Errorf("stream annotation %s is empty", constants.ConfigStream)
			}
			if d.stream != value {
				slog.Info("Updated stream configuration from node annotation", slog.String("annotation", constants.ConfigStream), slog.String("value", value))
				d.stream = value
			}
		case constants.ConfigFleetlockURL:
			if d.fleetlock.GetURL() != value {
				slog.Info("Updating fleetlock url from node annotation", slog.String("annotation", constants.ConfigFleetlockURL), slog.String("value", value))
			} else {
				continue
			}

			err := d.fleetlock.SetURL(value)
			if err != nil {
				return fmt.Errorf("failed to update fleetlock url to \"%s\": %v", value, err)
			}
		case constants.ConfigFleetlockGroup:
			if d.fleetlock.GetGroup() != value {
				slog.Info("Updating fleetlock group from node annotation", slog.String("annotation", constants.ConfigFleetlockGroup), slog.String("value", value))
			} else {
				continue
			}

			err := d.fleetlock.SetGroup(value)
			if err != nil {
				return fmt.Errorf("failed to update fleetlock group to \"%s\": %v", value, err)
			}
		case constants.ConfigCheckInterval:
			interval, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("failed to parse \"%s\" as duration: %v", value, err)
			}
			if d.checkInterval != interval {
				slog.Info("Updated check interval from node annotation", slog.String("annotation", constants.ConfigStream), slog.String("value", value))
				d.checkInterval = interval
			}
		case constants.ConfigRetryInterval:
			interval, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("failed to parse \"%s\" as duration: %v", value, err)
			}
			if d.retryInterval != interval {
				slog.Info("Updated retry interval from node annotation", slog.String("annotation", constants.ConfigStream), slog.String("value", value))
				d.retryInterval = interval
			}
		default:
			continue
		}
	}

	if d.fleetlock.GetURL() == "" {
		return fmt.Errorf("missing fleetlock server url")
	}

	return nil
}
