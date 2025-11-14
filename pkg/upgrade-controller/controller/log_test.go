package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLogLevel(t *testing.T) {
	tMatrix := map[string]int{
		"debug":   logLevelDebug,
		"info":    logLevelInfo,
		"warn":    logLevelWarn,
		"unknown": logLevelInfo,
		"Debug":   logLevelDebug,
		"WARN":    logLevelWarn,
	}
	for levelStr, logLevel := range tMatrix {
		t.Run(levelStr, func(t *testing.T) {
			t.Setenv(logLevelEnv, levelStr)
			assert.Equal(t, logLevel, getLogLevel(), "Should return correct log level")
		})
	}
	t.Run("VariableEmpty", func(t *testing.T) {
		t.Setenv(logLevelEnv, "")
		assert.Equal(t, logLevelInfo, getLogLevel(), "Should return default value")
	})
}
