package upgradehandler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewUpgradeHandler()

	assert.Equal(t, Name, cmd.Use)
}
