package fleetlock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetZincateAppID(t *testing.T) {
	assert := assert.New(t)

	id, err := GetZincateAppID()
	assert.NoError(err, "Should succeed")
	assert.NotEmpty(id, "Should return id")
}
