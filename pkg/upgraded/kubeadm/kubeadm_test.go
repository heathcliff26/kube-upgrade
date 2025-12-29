package kubeadm

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFromPath(t *testing.T) {
	assert := assert.New(t)

	cmd, err := NewFromPath("", "not-a-file")
	assert.Error(err, "Should not succeed")
	assert.Nil(cmd, "Should not return a command")

	cmd, err = NewFromPath("", "testdata/print-args.sh")
	assert.NoError(err, "Should succeed")
	assert.NotNil(cmd, "Should return a command")
	assert.Equal("version --output short", cmd.version, "Should have set the version without newline")

	cmd, err = NewFromPath("", "testdata/exit-1.sh")
	assert.Error(err, "Should not succeed")
	assert.Nil(cmd, "Should not return a command")
}

func TestNewFromVersion(t *testing.T) {
	oldCosignBinary := CosignBinary
	CosignBinary = "../../../bin/cosign"
	t.Cleanup(func() {
		CosignBinary = oldCosignBinary
	})

	t.Run("Download", func(t *testing.T) {
		t.Parallel()
		require := require.New(t)
		tmpDir := t.TempDir()

		cmd, err := NewFromVersion(tmpDir, "v1.35.0")
		require.NoError(err, "Should create a command")
		require.Equal("v1.35.0", cmd.Version(), "Downloaded version should match")
	})
	t.Run("DownloadFails", func(t *testing.T) {
		t.Parallel()
		require := require.New(t)
		tmpDir := t.TempDir()

		cmd, err := NewFromVersion(tmpDir, "invalid-version")
		require.Error(err, "Should fail to download invalid version")
		require.Nil(cmd, "Should not return a command")
	})
	t.Run("CosignError", func(t *testing.T) {
		oldCosignBinary := CosignBinary
		CosignBinary = "testdata/exit-1.sh"
		t.Cleanup(func() {
			CosignBinary = oldCosignBinary
		})
		assert := assert.New(t)
		tmpDir := t.TempDir()

		cmd, err := NewFromVersion(tmpDir, "v1.35.0")
		assert.ErrorContains(err, "invalid kubeadm binary:", "Should fail due to cosign error")
		assert.Nil(cmd, "Should not return a command")
	})
}

func TestApply(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	cmd, err := NewFromPath("", "testdata/print-args.sh")
	require.NoError(err, "Should create a command")

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
