package config

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DEFAULT_CONFIG_PATH = "/etc/kube-upgraded/config.yaml"

	DEFAULT_LOG_LEVEL       = "info"
	DEFAULT_KUBECONFIG      = "/etc/kubernetes/kubelet.conf"
	DEFAULT_STREAM          = "ghcr.io/heathcliff26/fcos-k8s"
	DEFAULT_FLEETLOCK_GROUP = "default"
	DEFAULT_RPM_OSTREE_PATH = "/usr/bin/rpm-ostree"
	DEFAULT_KUBEADM_PATH    = "/usr/bin/kubeadm"
	DEFAULT_CHECK_INTERVAL  = 3 * time.Hour
	DEFAULT_RETRY_INTERVAL  = 5 * time.Minute
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

type Config struct {
	// The log level used by slog, default "info"
	LogLevel string `yaml:"logLevel,omitempty"`
	// The path to the kubeconfig file, default is the kubelet config under "/etc/kubernetes/kubelet.conf"
	Kubeconfig string `yaml:"kubeconfig,omitempty"`
	// The container image repository for os rebases
	Stream string `yaml:"stream,omitempty"`
	// Configuration for fleetlock node locking
	Fleetlock FleetlockConfig `yaml:"fleetlock"`
	// The path to the rpm-ostree binary, default "/usr/bin/rpm-ostree"
	RPMOStreePath string `yaml:"rpm-ostree-path,omitempty"`
	// The path to the kubeadm binary, default "/usr/bin/kubeadm"
	KubeadmPath string `yaml:"kubeadm-path,omitempty"`

	// The interval between regular checks, default 3h
	CheckInterval time.Duration `yaml:"check-interval,omitempty"`
	// The interval between retries when an operation fails, default 5m
	RetryInterval time.Duration `yaml:"retry-interval,omitempty"`
}

type FleetlockConfig struct {
	// URL to fleetlock server
	URL string `yaml:"url"`
	// The node group to use for fleetlock, default group is "default"
	Group string `yaml:"group,omitempty"`
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

func DefaultConfig() *Config {
	return &Config{
		LogLevel:   DEFAULT_LOG_LEVEL,
		Kubeconfig: DEFAULT_KUBECONFIG,
		Stream:     DEFAULT_STREAM,
		Fleetlock: FleetlockConfig{
			Group: DEFAULT_FLEETLOCK_GROUP,
		},
		RPMOStreePath: DEFAULT_RPM_OSTREE_PATH,
		KubeadmPath:   DEFAULT_KUBEADM_PATH,
		CheckInterval: DEFAULT_CHECK_INTERVAL,
		RetryInterval: DEFAULT_RETRY_INTERVAL,
	}
}

// Loads the config from the given path.
// When path is empty, it checks the default path "/etc/kube-upgraded/config.yaml".
// When no config is found in the default path, it returns the default config.
// Returns error when the given config is invalid.
func LoadConfig(path string) (*Config, error) {
	c := DefaultConfig()

	p := path
	if path == "" {
		p = DEFAULT_CONFIG_PATH
	}

	f, err := os.ReadFile(p)
	if os.IsNotExist(err) && path == "" {
		slog.Info("No config file specified and default file does not exist, falling back to default values.", slog.String("default-path", DEFAULT_CONFIG_PATH))
		return c, nil
	} else if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return nil, err
	}

	err = setLogLevel(c.LogLevel)
	if err != nil {
		return nil, err
	}

	return c, nil
}
