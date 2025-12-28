package config

import (
	"log/slog"
	"testing"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"github.com/stretchr/testify/assert"
)

func TestValidConfig(t *testing.T) {
	c := DefaultConfig()
	c.LogLevel = "debug"
	c.FleetlockURL = "https://fleetlock.example.com"

	res, err := LoadConfig("testdata/valid-config.yaml")

	assert := assert.New(t)

	if !assert.NoError(err) {
		t.Fatalf("Failed to load config: %v", err)
	}
	assert.Equal(c, res)

}

func TestSetLogLevel(t *testing.T) {
	tMatrix := []struct {
		Name  string
		Level slog.Level
		Error error
	}{
		{"debug", slog.LevelDebug, nil},
		{"info", slog.LevelInfo, nil},
		{"warn", slog.LevelWarn, nil},
		{"error", slog.LevelError, nil},
		{"DEBUG", slog.LevelDebug, nil},
		{"INFO", slog.LevelInfo, nil},
		{"WARN", slog.LevelWarn, nil},
		{"ERROR", slog.LevelError, nil},
		{"Unknown", 0, &ErrUnknownLogLevel{"Unknown"}},
	}
	t.Cleanup(func() {
		err := setLogLevel(api.DefaultUpgradedLogLevel)
		if err != nil {
			t.Fatalf("Failed to cleanup after test: %v", err)
		}
	})

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			err := setLogLevel(tCase.Name)

			assert := assert.New(t)

			if !assert.Equal(tCase.Error, err) {
				t.Fatalf("Received invalid error: %v", err)
			}
			if err == nil {
				assert.Equal(tCase.Level, logLevel.Level())
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	c := DefaultConfig()

	assert := assert.New(t)

	assert.Equal(api.DefaultUpgradedStream, c.Stream)
	assert.Equal(api.DefaultUpgradedFleetlockGroup, c.FleetlockGroup)
	assert.Equal(api.DefaultUpgradedCheckInterval, c.CheckInterval)
	assert.Equal(api.DefaultUpgradedRetryInterval, c.RetryInterval)
	assert.Equal(api.DefaultUpgradedLogLevel, c.LogLevel)
	assert.Equal(api.DefaultUpgradedKubeletConfig, c.KubeletConfig)
	assert.Equal("", c.KubeadmPath)
}

func TestLoadConfigError(t *testing.T) {
	tMatrix := map[string]string{
		"WrongFile":              "testdata/not-yaml.txt",
		"EmptyConfigFile":        "testdata/empty-file.yaml",
		"ConfigFileDoesNotExist": "not-a-file",
		"EmptyStream":            "testdata/empty-stream.yaml",
		"EmptyFleetlockURL":      "testdata/empty-fleetlockUrl.yaml",
		"EmptyFleetlockGroup":    "testdata/empty-fleetlockGroup.yaml",
		"EmptyKubeletConfig":     "testdata/empty-kubeletConfig.yaml",
	}

	for name, path := range tMatrix {
		t.Run(name, func(t *testing.T) {
			c, err := LoadConfig(path)

			assert := assert.New(t)

			assert.Error(err)
			assert.Nil(c)
		})
	}
	t.Run("InvalidLogLevel", func(t *testing.T) {
		cfg, err := LoadConfig("testdata/invalid-log-level.yaml")

		assert := assert.New(t)

		assert.Nil(cfg)
		assert.Equal(NewErrUnknownLogLevel("not-a-log-level"), err)
	})
}
