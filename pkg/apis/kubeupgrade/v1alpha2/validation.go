package v1alpha2

import (
	"fmt"
	"net/url"
	"time"

	"golang.org/x/mod/semver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ValidateObject_KubeUpgradePlan(plan *KubeUpgradePlan) error {
	return ValidateObject_KubeUpgradeSpec(plan.Spec)
}

func ValidateObject_KubeUpgradeSpec(spec KubeUpgradeSpec) error {
	if !semver.IsValid(spec.KubernetesVersion) {
		return fmt.Errorf("invalid input for spec.kubernetesVersion, \"%s\" is not a valid semantic version", spec.KubernetesVersion)
	}
	if semver.Prerelease(spec.KubernetesVersion) == "" && semver.Canonical(spec.KubernetesVersion) != spec.KubernetesVersion {
		return fmt.Errorf("invalid input for spec.kubernetesVersion, \"%s\" needs to be a full version like vX.Y.Z", spec.KubernetesVersion)
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

		if group.Labels == nil || (len(group.Labels.MatchExpressions) < 1 && len(group.Labels.MatchLabels) < 1) {
			return fmt.Errorf("group \"%s\" needs at least one label selector", name)
		}
		_, err := metav1.LabelSelectorAsSelector(group.Labels)
		if err != nil {
			return fmt.Errorf("invalid label selector for group \"%s\": %v", name, err)
		}

		if group.Upgraded != nil {
			err := ValidateObject_UpgradedConfig(*group.Upgraded)
			if err != nil {
				return fmt.Errorf("group \"%s\" has an invalid upgraded config: %v", name, err)
			}
		}
	}

	err := ValidateObject_UpgradedConfig(spec.Upgraded)
	if err != nil {
		return err
	}
	if spec.Upgraded.FleetlockURL == "" {
		return fmt.Errorf("missing parameter spec.upgraded.fleetlock-url")
	}

	return nil
}

func ValidateObject_UpgradedConfig(cfg UpgradedConfig) error {
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
