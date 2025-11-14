package controller

import (
	"testing"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	upgradedconfig "github.com/heathcliff26/kube-upgrade/pkg/upgraded/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpgradedDaemonSet(t *testing.T) {
	assert := assert.New(t)

	c := &controller{
		namespace:     "test-namespace",
		upgradedImage: "registry.example.com/kube-upgraded:latest",
	}
	labels := upgradedLabels("testplan", "testgroup")

	daemon := c.NewUpgradedDaemonSet("testplan", "testgroup")

	assert.Equal("upgraded-testgroup", daemon.Name, "Should have correct name")
	assert.Equal("test-namespace", daemon.Namespace, "Should have correct namespace")
	assert.Equal(labels, daemon.Labels, "Should have correct labels")

	assert.Equal(labels, daemon.Spec.Selector.MatchLabels, "Should have correct selector labels")
	assert.Equal(labels, daemon.Spec.Template.Labels, "Should have correct template labels")
	assert.Equal(c.upgradedImage, daemon.Spec.Template.Spec.Containers[0].Image, "Should have correct container image")

	found := false
	for _, vol := range daemon.Spec.Template.Spec.Volumes {
		if vol.Name == "config" {
			assert.Equal("upgraded-testgroup", vol.ConfigMap.Name, "Should have correct config map name")
			found = true
			break
		}
	}
	assert.True(found, "Should have config volume defined")
}

func TestNewUpgradedConfigMap(t *testing.T) {
	assert := assert.New(t)

	c := &controller{
		namespace: "test-namespace",
	}
	cfg := &api.UpgradedConfig{}
	api.SetObjectDefaults_UpgradedConfig(cfg)

	cm, err := c.NewUpgradedConfigMap("testplan", "testgroup", cfg)
	require.NoError(t, err, "Should create upgraded config map without error")

	assert.Equal("upgraded-testgroup", cm.Name, "Should have correct name")
	assert.Equal("test-namespace", cm.Namespace, "Should have correct namespace")
	assert.Equal(upgradedLabels("testplan", "testgroup"), cm.Labels, "Should have correct labels")

	data, exists := cm.Data[upgradedconfig.DefaultConfigFile]
	assert.True(exists, "Config map should have data for upgraded config file")
	assert.NotEmpty(data, "Config map data for upgraded config file should not be empty")
	assert.Contains(data, "logLevel:", "Config map data should contain logLevel field")
}

func TestUpgradedLabels(t *testing.T) {
	assert := assert.New(t)

	labels := upgradedLabels("testplan", "testgroup")

	expected := map[string]string{
		constants.LabelPlanName:  "testplan",
		constants.LabelNodeGroup: "testgroup",
	}
	assert.Equal(expected, labels, "Should have correct upgraded labels")
}
