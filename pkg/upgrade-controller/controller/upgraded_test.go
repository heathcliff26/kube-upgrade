package controller

import (
	"testing"

	"github.com/heathcliff26/kube-upgrade/pkg/constants"
	"github.com/stretchr/testify/assert"
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
