package controller

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

const logLevelEnv = "LOG_LEVEL"

const (
	logLevelDebug = 4
	logLevelInfo  = 0
	logLevelWarn  = -4
)

func init() {
	var fs flag.FlagSet
	klog.InitFlags(&fs)
	err := fs.Set("v", strconv.Itoa(getLogLevel()))
	if err != nil {
		klog.Fatalf("Failed to set klog verbosity: %v", err)
	}
	logger := klog.NewKlogr()
	ctrl.SetLogger(logger)
	logger.WithValues(logLevelEnv, os.Getenv(logLevelEnv), "level", getLogLevel()).Info("Logger initialized")
}

func getLogLevel() int {
	levelStr := os.Getenv(logLevelEnv)
	switch strings.ToLower(levelStr) {
	case "debug":
		return logLevelDebug
	case "info":
		return logLevelInfo
	case "warn":
		return logLevelWarn
	case "":
		return logLevelInfo
	default:
		klog.Warningf("Unknown log level '%s', defaulting to 'info'", levelStr)
		return logLevelInfo
	}
}
