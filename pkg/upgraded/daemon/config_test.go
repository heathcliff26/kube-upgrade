package daemon

import (
	"os"
	"testing"
	"time"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"github.com/heathcliff26/kube-upgrade/pkg/upgraded/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func TestUpdateFromConfigFile(t *testing.T) {
	t.Run("FailedToLoadConfig", func(t *testing.T) {
		d := &daemon{
			cfgPath: "not-a-file",
		}

		assert.Error(t, d.UpdateFromConfigFile(), "Should fail to load config from file")
	})
	t.Run("UpdateSuccess", func(t *testing.T) {
		assert := assert.New(t)

		d := &daemon{
			cfgPath: "testdata/config/valid-config-full.yaml",
		}

		assert.NoError(d.UpdateFromConfigFile(), "Should update config from file")

		assert.Equal("registry.example.com/fcos-k8s", d.Stream(), "Stream should match")
		assert.Equal("https://fleetlock.example.com", d.Fleetlock().GetURL(), "Fleetlock URL should match")
		assert.Equal("compute", d.Fleetlock().GetGroup(), "Fleetlock group should match")
		assert.Equal(10*time.Minute, d.CheckInterval(), "Check interval should match")
		assert.Equal(2*time.Minute, d.RetryInterval(), "Retry interval should match")
	})
}

func TestUpdateFromConfig(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		assert := assert.New(t)

		d := &daemon{}
		cfg := &api.UpgradedConfig{
			Stream:       "registry.example.com/fcos-k8s",
			FleetlockURL: "https://fleetlock.example.com",
		}
		api.SetObjectDefaults_UpgradedConfig(cfg)

		assert.NoError(d.updateFromConfig(cfg), "Should update from config")

		assert.Equal(cfg.Stream, d.Stream(), "Stream should match")
		assert.Equal(cfg.FleetlockURL, d.Fleetlock().GetURL(), "Fleetlock URL should match")
		assert.Equal(cfg.FleetlockGroup, d.Fleetlock().GetGroup(), "Fleetlock group should match")
		checkInterval, _ := time.ParseDuration(cfg.CheckInterval)
		retryInterval, _ := time.ParseDuration(cfg.RetryInterval)
		assert.Equal(checkInterval, d.CheckInterval(), "Check interval should match")
		assert.Equal(retryInterval, d.RetryInterval(), "Retry interval should match")
	})
	tMatrix := []struct {
		Name string
		Cfg  *api.UpgradedConfig
	}{
		{
			Name: "MissingFleetlockURL",
			Cfg:  &api.UpgradedConfig{},
		},
		{
			Name: "MisformedCheckInterval",
			Cfg: &api.UpgradedConfig{
				FleetlockURL:  "https://fleetlock.example.com",
				CheckInterval: "not-a-duration",
			},
		},
		{
			Name: "MisformedRetryInterval",
			Cfg: &api.UpgradedConfig{
				FleetlockURL:  "https://fleetlock.example.com",
				RetryInterval: "not-a-duration",
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			d := &daemon{}

			api.SetObjectDefaults_UpgradedConfig(tCase.Cfg)

			assert.Error(d.updateFromConfig(tCase.Cfg), "Should fail to update config")

			assert.Empty(d.Stream(), "Should not update stream")
			assert.Nil(d.Fleetlock(), "Should not update fleetlock client")
			assert.Zero(d.CheckInterval(), "Should not update check interval")
			assert.Zero(d.RetryInterval(), "Should not update retry interval")
		})
	}

	t.Run("ShouldLockDemonOnUpdate", func(t *testing.T) {
		assert := assert.New(t)

		d := &daemon{}
		cfg := &api.UpgradedConfig{
			Stream:       "registry.example.com/fcos-k8s",
			FleetlockURL: "https://fleetlock.example.com",
		}
		api.SetObjectDefaults_UpgradedConfig(cfg)

		done := make(chan error, 1)

		d.configLock.Lock()
		go func() {
			done <- d.updateFromConfig(cfg)
		}()

		select {
		case <-done:
			t.Fatal("UpdateFromConfig should be blocked by lock")
		case <-time.After(500 * time.Millisecond):
			// Expected case
		}

		d.configLock.Unlock()

		select {
		case err := <-done:
			assert.NoError(err, "Should update from config after unlock")
		case <-time.After(100 * time.Millisecond):
			t.Fatal("UpdateFromConfig should have finished after unlock")
		}
	})
}

func TestNewConfigFileWatcher(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		d := &daemon{
			cfgPath: "testdata/config/valid-config-full.yaml",
		}

		require.NoError(d.NewConfigFileWatcher(), "Should create new config file watcher")
		t.Cleanup(func() {
			_ = d.configWatcher.Close()
		})
		assert.NotNil(d.configWatcher, "Config watcher should not be nil")
	})
	t.Run("Failure", func(t *testing.T) {
		assert := assert.New(t)

		d := &daemon{
			cfgPath: "/foo/bar/not-a-file",
		}

		assert.Error(d.NewConfigFileWatcher(), "Should fail to create config file watcher")
		if d.configWatcher != nil {
			_ = d.configWatcher.Close()
		}
	})
}

func TestWatchConfigFile(t *testing.T) {
	require := require.New(t)

	tmpDir := t.TempDir()
	cfgPath := tmpDir + "/config.yaml"

	cfg, err := config.LoadConfig("testdata/config/valid-config-full.yaml")
	require.NoError(err, "Should load initial config")
	require.NoError(saveConfigToFile(cfg, cfgPath), "Should save initial config to temp file")

	d := &daemon{
		cfgPath: cfgPath,
		ctx:     t.Context(),
	}
	require.NoError(d.UpdateFromConfigFile(), "Should load initial config from file")
	require.Equal(cfg.Stream, d.Stream(), "Initial stream should match")

	require.NoError(d.NewConfigFileWatcher(), "Should create config file watcher")
	t.Cleanup(func() {
		_ = d.configWatcher.Close()
	})
	go d.WatchConfigFile()

	cfg.Stream = "registry.example.com/updated-stream"
	require.NoError(saveConfigToFile(cfg, cfgPath), "Should save updated config to file")

	require.Eventually(func() bool {
		return cfg.Stream == d.Stream()
	}, 5*time.Second, 100*time.Millisecond, "Daemon should update stream from config file")
}

func saveConfigToFile(cfg *api.UpgradedConfig, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
