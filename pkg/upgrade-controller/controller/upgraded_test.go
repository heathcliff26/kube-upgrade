package controller

import (
	"testing"

	api "github.com/heathcliff26/kube-upgrade/pkg/apis/kubeupgrade/v1alpha3"
	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	upgradedconfig "github.com/heathcliff26/kube-upgrade/pkg/upgraded/config"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestNewEmptyUpgradedDaemonSet(t *testing.T) {
	assert := assert.New(t)

	c := &controller{
		namespace: "test-namespace",
	}
	daemon := c.NewEmptyUpgradedDaemonSet("testplan", "testgroup")

	assert.Equal("upgraded-testgroup", daemon.Name, "Should have correct name")
	assert.Equal("test-namespace", daemon.Namespace, "Should have correct namespace")
	assert.Equal(upgradedLabels("testplan", "testgroup"), daemon.Labels, "Should have correct labels")
}

func TestNewUpgradedDaemonSetSpec(t *testing.T) {
	assert := assert.New(t)

	c := &controller{
		namespace:     "test-namespace",
		upgradedImage: "registry.example.com/kube-upgraded:latest",
	}
	labels := upgradedLabels("testplan", "testgroup")
	spec := c.NewUpgradedDaemonSetSpec("testplan", "testgroup")

	assert.Equal(labels, spec.Selector.MatchLabels, "Should have correct selector labels")
	assert.Equal(labels, spec.Template.Labels, "Should have correct template labels")
	assert.Equal(c.upgradedImage, spec.Template.Spec.Containers[0].Image, "Should have correct container image")

	for _, vol := range spec.Template.Spec.Volumes {
		if vol.Name == "config" {
			assert.Equal("upgraded-testgroup", vol.ConfigMap.Name, "Should have correct config map name")
			break
		}
	}
}

func TestAttachUpgradedConfigMapData(t *testing.T) {
	assert := assert.New(t)

	c := &controller{}
	cm := &corev1.ConfigMap{}
	cfg := &api.UpgradedConfig{}
	api.SetObjectDefaults_UpgradedConfig(cfg)

	assert.NoError(c.AttachUpgradedConfigMapData(cm, cfg), "Should attach config map data without error")
	data, exists := cm.Data[upgradedconfig.DefaultConfigFile]
	assert.True(exists, "Config map should have data for upgraded config file")
	assert.NotEmpty(data, "Config map data for upgraded config file should not be empty")
	assert.Contains(data, "logLevel:", "Config map data should contain logLevel field")
}

func TestNewEmptyUpgradedConfigMap(t *testing.T) {
	assert := assert.New(t)

	c := &controller{
		namespace: "test-namespace",
	}
	cm := c.NewEmptyUpgradedConfigMap("testplan", "testgroup")

	assert.Equal("upgraded-testgroup", cm.Name, "Should have correct name")
	assert.Equal("test-namespace", cm.Namespace, "Should have correct namespace")
	assert.Equal(upgradedLabels("testplan", "testgroup"), cm.Labels, "Should have correct labels")
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
