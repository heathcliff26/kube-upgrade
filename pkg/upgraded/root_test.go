package upgraded

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewUpgraded()

	assert.Equal(t, Name, cmd.Use)
}
