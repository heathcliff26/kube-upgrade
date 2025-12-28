package rpmostree

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)

	cmd, err := New("not-a-file")
	assert.Error(err, "Should not succeed")
	assert.Nil(cmd, "Should not return a command")

	cmd, err = New("testdata/exit-0.sh")
	assert.NoError(err, "Should succeed")
	assert.NotNil(cmd, "Should return a command")
}

func TestCheckForUpgrade(t *testing.T) {
	tMatrix := []struct {
		Name   string
		Path   string
		Result bool
		Error  bool
	}{
		{
			Name:   "NeedUpgrade",
			Path:   "testdata/exit-0.sh",
			Result: true,
		},
		{
			Name:   "DoesNotNeedUpgrade",
			Path:   "testdata/exit-77.sh",
			Result: false,
		},
		{
			Name:  "UnknownExitCode",
			Path:  "testdata/exit-1.sh",
			Error: true,
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			cmd, err := New(tCase.Path)
			require.NoError(t, err, "Should create a command")
			res, err := cmd.CheckForUpgrade()

			if tCase.Error {
				assert.Error(err, "Should not succeed")
				assert.False(res, "Result should be default false on failure")
			} else {
				assert.NoError(err, "Should succeed")
				assert.Equal(tCase.Result, res, "Result should match")
			}
		})
	}
}

func TestRebase(t *testing.T) {
	tMatrix := map[string]bool{
		"Verified":   false,
		"Unverified": true,
	}

	for name, unverified := range tMatrix {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd, err := New("testdata/print-args.sh")
			require.NoError(t, err, "Should create a command")

			actualStdout := os.Stdout
			rOut, wOut, _ := os.Pipe()
			os.Stdout = wOut

			err = cmd.Rebase("test-image", unverified)

			wOut.Close()
			stdout, _ := io.ReadAll(rOut)
			os.Stdout = actualStdout

			assert.NoError(err, "Command should succeed")
			if unverified {
				assert.Equal("rebase --reboot ostree-unverified-registry:test-image\n", string(stdout), "Should have added image to command args")
			} else {
				assert.Equal("rebase --reboot ostree-image-signed:docker://test-image\n", string(stdout), "Should have added image to command args")
			}

		})
	}

}

func TestGetBootedImageRef(t *testing.T) {
	assert := assert.New(t)

	cmd, err := New("testdata/print-status.sh")
	require.NoError(t, err, "Should create a command")

	ref, err := cmd.GetBootedImageRef()
	assert.NoError(err, "Should succeed")
	assert.Equal("ostree-unverified-registry:ghcr.io/heathcliff26/fcos-k8s:v1.34.2", ref, "Should return correct image ref")
}
