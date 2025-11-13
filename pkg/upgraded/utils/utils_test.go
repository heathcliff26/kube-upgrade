package utils

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMachineID(t *testing.T) {
	assert := assert.New(t)

	id, err := GetMachineID()
	assert.NoError(err, "Should succeed")
	assert.NotEmpty(id, "Should return id")
}

func TestCreateCMDWithStdout(t *testing.T) {
	assert := assert.New(t)

	actualStdout := os.Stdout
	actualStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr
	err := CreateCMDWithStdout("/bin/bash", "-c", "echo stdout && echo stderr >&2").Run()

	wOut.Close()
	stdout, _ := io.ReadAll(rOut)
	os.Stdout = actualStdout
	wErr.Close()
	stderr, _ := io.ReadAll(rErr)
	os.Stderr = actualStderr

	assert.NoError(err, "Command should succeed")
	assert.Equal("stdout\n", string(stdout), "Should have written to fake stdout")
	assert.Equal("stderr\n", string(stderr), "Should have written to fake stderr")
}

func TestCreateChrootCMDWithStdout(t *testing.T) {
	assert := assert.New(t)

	cmd := CreateChrootCMDWithStdout("/chroot", "/bin/echo", "chrooted")
	require.NotNil(t, cmd.SysProcAttr)

	assert.Equal("/chroot", cmd.SysProcAttr.Chroot, "Chroot path should be set")
	assert.Equal("/bin/echo", cmd.Path, "Command path should be set")
	assert.Equal([]string{"/bin/echo", "chrooted"}, cmd.Args, "Command args should be set")
	assert.Equal(os.Stdout, cmd.Stdout, "Command should use stdout")
	assert.Equal(os.Stderr, cmd.Stderr, "Command should use stderr")
}

func TestCheckExistsAndIsExecutable(t *testing.T) {
	tMatrix := []struct {
		Name    string
		Perms   os.FileMode
		Success bool
	}{
		{
			Name:    "Executable",
			Perms:   0744,
			Success: true,
		},
		{
			Name:  "NotExecutable",
			Perms: 0644,
		},
		{
			Name: "NotFound",
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			testfilePath := "/tmp/kube-upgrade-testfile." + tCase.Name

			if tCase.Perms != 0 {
				err := os.WriteFile(testfilePath, []byte("#!/bin/bash"), tCase.Perms)
				if !assert.NoError(err, "Should create test file") {
					t.FailNow()
				}
				t.Cleanup(func() {
					err := os.Remove(testfilePath)
					if err != nil {
						t.Logf("Failed to clean up testfile %s: %v", testfilePath, err)
					}
				})
			}

			err := CheckExistsAndIsExecutable(testfilePath)

			if tCase.Success {
				assert.NoError(err, "Should succeed")
			} else {
				assert.Error(err, "Should not succeed")
			}
		})
	}
}
