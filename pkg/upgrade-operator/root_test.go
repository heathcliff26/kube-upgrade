package upgradeoperator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewUpgradeOperator()

	assert.Equal(t, Name, cmd.Use)
}
