package upgradecontroller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewUpgradeController()

	assert.Equal(t, Name, cmd.Use)
}
