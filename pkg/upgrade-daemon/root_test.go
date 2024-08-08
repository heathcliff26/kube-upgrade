package upgradedaemon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewUpgradeDaemon()

	assert.Equal(t, Name, cmd.Use)
}
