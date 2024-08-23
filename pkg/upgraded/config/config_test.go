package config

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidConfig(t *testing.T) {
	c := DefaultConfig()
	c.LogLevel = "debug"
	c.Fleetlock.URL = "http://fleetlock.example.com"
	c.CheckInterval = time.Hour * 1
	c.RetryInterval = time.Hour * 5

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
		err := setLogLevel(DEFAULT_LOG_LEVEL)
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

	assert.Equal(DEFAULT_LOG_LEVEL, c.LogLevel)
	assert.Equal(DEFAULT_KUBECONFIG, c.Kubeconfig)
	assert.Equal(DEFAULT_IMAGE, c.Image)
	assert.Equal("", c.Fleetlock.URL)
	assert.Equal(DEFAULT_FLEETLOCK_GROUP, c.Fleetlock.Group)
	assert.Equal(DEFAULT_RPM_OSTREE_PATH, c.RPMOStreePath)
	assert.Equal(DEFAULT_KUBEADM_PATH, c.KubeadmPath)
	assert.Equal(DEFAULT_CHECK_INTERVAL, c.CheckInterval)
	assert.Equal(DEFAULT_RETRY_INTERVAL, c.RetryInterval)
}

func TestInvalidConfigs(t *testing.T) {
	cfg, err := LoadConfig("testdata/invalid-log-level.yaml")

	assert := assert.New(t)

	assert.Nil(cfg)
	assert.Equal(NewErrUnknownLogLevel("not-a-log-level"), err)
}

func TestWrongFile(t *testing.T) {
	cfg, err := LoadConfig("testdata/not-yaml.txt")

	assert := assert.New(t)

	assert.Nil(cfg)
	assert.Error(err)
}

func TestDefaultConfigPath(t *testing.T) {
	c, err := LoadConfig("")

	assert := assert.New(t)

	assert.NoError(err)
	assert.Equal(DefaultConfig(), c)
}

func TestConfigFileDoesNotExist(t *testing.T) {
	c, err := LoadConfig("not-a-file")

	assert := assert.New(t)

	assert.Error(err)
	assert.Nil(c)
}
