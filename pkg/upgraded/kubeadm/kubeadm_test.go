package kubeadm

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)

	cmd, err := New("", "not-a-file")
	assert.Error(err, "Should not succeed")
	assert.Nil(cmd, "Should not return a command")

	cmd, err = New("", "testdata/print-args.sh")
	assert.NoError(err, "Should succeed")
	assert.NotNil(cmd, "Should return a command")
	assert.Equal("version --output short", cmd.version, "Should have set the version without newline")
}

func TestApply(t *testing.T) {
	assert := assert.New(t)

	cmd, err := New("", "testdata/print-args.sh")
	if !assert.NoError(err, "Should create a command") {
		t.FailNow()
	}

	actualStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	err = cmd.Apply("test-version")

	wOut.Close()
	stdout, _ := io.ReadAll(rOut)
	os.Stdout = actualStdout

	assert.NoError(err, "Command should succeed")
	assert.Equal("upgrade apply --yes test-version\n", string(stdout), "Should have added version to command args")
}
