package config

import (
	"log/slog"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DEFAULT_CONFIG_PATH = "/etc/kube-upgraded/config.yaml"

	DEFAULT_LOG_LEVEL       = "info"
	DEFAULT_KUBECONFIG      = "/etc/kubernetes/kubelet.conf"
	DEFAULT_RPM_OSTREE_PATH = "/usr/bin/rpm-ostree"
	DEFAULT_KUBEADM_PATH    = "/usr/bin/kubeadm"
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
	// The path to the rpm-ostree binary, default "/usr/bin/rpm-ostree"
	RPMOStreePath string `yaml:"rpm-ostree-path,omitempty"`
	// The path to the kubeadm binary, default "/usr/bin/kubeadm"
	KubeadmPath string `yaml:"kubeadm-path,omitempty"`
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
		LogLevel:      DEFAULT_LOG_LEVEL,
		Kubeconfig:    DEFAULT_KUBECONFIG,
		RPMOStreePath: DEFAULT_RPM_OSTREE_PATH,
		KubeadmPath:   DEFAULT_KUBEADM_PATH,
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
