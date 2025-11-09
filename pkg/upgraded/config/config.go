package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"sigs.k8s.io/yaml"
)

const (
	DefaultConfigDir  = "/etc/kube-upgraded/"
	DefaultConfigFile = "config.yaml"
	DefaultConfigPath = DefaultConfigDir + DefaultConfigFile
)

var logLevel = &slog.LevelVar{}

// Initialize the logger
func init() {
	logLevel = &slog.LevelVar{}
	opts := slog.HandlerOptions{
		Level: logLevel,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	slog.SetDefault(logger)
}

// Parse a given string and set the resulting log level
func setLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "info":
		logLevel.Set(slog.LevelInfo)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		return NewErrUnknownLogLevel(level)
	}
	return nil
}

func DefaultConfig() *api.UpgradedConfig {
	cfg := &api.UpgradedConfig{}
	api.SetObjectDefaults_UpgradedConfig(cfg)
	return cfg
}

// Loads the config from the given path.
// When path is empty, it checks the default path "/etc/kube-upgraded/config.yaml".
// When no config is found in the default path, it returns the default config.
// Returns error when the given config is invalid.
func LoadConfig(path string) (*api.UpgradedConfig, error) {
	c := DefaultConfig()

	if path == "" {
		path = DefaultConfigPath
	}

	// #nosec G304: File will be passed by controller
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return nil, err
	}

	err = ValidateConfig(c)
	if err != nil {
		return nil, err
	}

	err = setLogLevel(c.LogLevel)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Validates the given config to ensure all required fields are set.
func ValidateConfig(cfg *api.UpgradedConfig) error {
	if cfg == nil {
		return fmt.Errorf("invalid config, config is nil")
	}
	if cfg.Stream == "" {
		return fmt.Errorf("invalid config, missing stream")
	}
	if cfg.FleetlockURL == "" {
		return fmt.Errorf("invalid config, missing fleetlock-url")
	}
	if cfg.FleetlockGroup == "" {
		return fmt.Errorf("invalid config, missing fleetlock-group")
	}
	if cfg.KubeletConfig == "" {
		return fmt.Errorf("invalid config, missing kubelet-config")
	}
	if cfg.KubeadmPath == "" {
		return fmt.Errorf("invalid config, missing kubeadm-path")
	}
	return nil
}
