package v1alpha1

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"time"
)

var (
	validStatusValues = []string{"Unknown", "Waiting", "Progressing", "Complete"}
)

func ValidateObject_KubeUpgradePlan(plan *KubeUpgradePlan) error {
	err := ValidateObject_KubeUpgradeSpec(plan.Spec)
	if err != nil {
		return err
	}

	return ValidateObject_KubeUpgradeStatus(plan.Status)
}

func ValidateObject_KubeUpgradeSpec(spec KubeUpgradeSpec) error {
	versionRegex := regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
	if !versionRegex.MatchString(spec.KubernetesVersion) {
		return fmt.Errorf("invalid input for spec.kubernetesVersion, \"%s\" is not a valid version", spec.KubernetesVersion)
	}

	if len(spec.Groups) < 1 {
		return fmt.Errorf("need at least one node group for upgrades")
	}
	for name, group := range spec.Groups {
		for _, dependency := range group.DependsOn {
			if _, ok := spec.Groups[dependency]; !ok {
				return fmt.Errorf("group \"%s\" depends on non-existing group \"%s\"", name, dependency)
			}
		}
		if len(group.Labels) < 1 {
			return fmt.Errorf("group \"%s\" needs at least one label", name)
		}
		err := ValidateObject_UpgradedConfig(group.Upgraded)
		if err != nil {
			return fmt.Errorf("group \"%s\" has an invalid upgraded config: %v", name, err)
		}
	}

	err := ValidateObject_UpgradedConfig(spec.Upgraded)
	if err != nil {
		return err
	}
	if spec.Upgraded != nil && spec.Upgraded.FleetlockURL == "" {
		return fmt.Errorf("missing parameter spec.upgraded.fleetlock-url")
	}

	return nil
}

func ValidateObject_UpgradedConfig(cfg *UpgradedConfig) error {
	if cfg == nil {
		return nil
	}

	if cfg.Stream != "" {
		_, err := url.ParseRequestURI("http://" + cfg.Stream)
		if err != nil {
			return fmt.Errorf("invalid input \"%s\" for stream: %v", cfg.Stream, err)
		}
	}

	if cfg.FleetlockURL != "" {
		_, err := url.ParseRequestURI(cfg.FleetlockURL)
		if err != nil {
			return fmt.Errorf("invalid input \"%s\" for fleetlock-url: %v", cfg.FleetlockURL, err)
		}
	}

	if cfg.CheckInterval != "" {
		_, err := time.ParseDuration(cfg.CheckInterval)
		if err != nil {
			return fmt.Errorf("invalid input \"%s\" for check-interval: %v", cfg.CheckInterval, err)
		}
	}

	if cfg.RetryInterval != "" {
		_, err := time.ParseDuration(cfg.RetryInterval)
		if err != nil {
			return fmt.Errorf("invalid input \"%s\" for retry-interval: %v", cfg.RetryInterval, err)
		}
	}

	return nil
}

func ValidateObject_KubeUpgradeStatus(status KubeUpgradeStatus) error {
	// Mutating/Validation webhooks for subresources are called later, so it is ok if the status does not exist
	if status.Summary == "" && len(status.Groups) == 0 {
		return nil
	}

	if !slices.Contains(validStatusValues, status.Summary) {
		return fmt.Errorf("found unknown status \"%s\" in summary, accepted values are: %v", status.Summary, validStatusValues)
	}

	for group, value := range status.Groups {
		if !slices.Contains(validStatusValues, value) {
			return fmt.Errorf("found unknown status \"%s\" in group \"%s\", accepted values are: %v", value, group, validStatusValues)
		}
	}

	return nil
}
