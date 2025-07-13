// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFile = "config.yaml"
	defaultLogLevel   = "info"
	envConfigFile     = "ESW_CONFIG_FILE"
)

var Settings Config

func Init() {
	initFlags()
	initLogger(defaultLogLevel)
}

func initFlags() {
	Settings.Flags.LogFile = flag.String("log_file", "", "Specify the path for log files.")
	Settings.Flags.Version = flag.Bool("version", false,
		"Print the version of the service and exit!")

	// Parse the command line flags.
	flag.Parse()
	if *Settings.Flags.Version {
		printVersionInformation()
	}
}

func printVersionInformation() {
	fmt.Println("HP CEM Enroll Service: version information")
	fmt.Printf("- Git commit hash: %s\n - Built at: %s\n - Built by: %s\n - Built on: %s\n",
		Settings.Flags.gitCommitHash,
		Settings.Flags.builtAt,
		Settings.Flags.builtBy,
		Settings.Flags.builtOn)
}

func Load(testModeEnabled bool) bool {
	filename := getConfigFile()

	// Open the configuration file for parsing.
	bytes, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		eswLogger.Error("Failed to load configuration file!",
			zap.String("Configuration file:", filename),
			zap.Error(err),
		)
		return false
	}

	// Read the configuration file and unmarshal the YAML.
	err = yaml.Unmarshal(bytes, &Settings)
	if err != nil {
		eswLogger.Error("Failed to parse configuration file!",
			zap.String("Configuration file:", filename),
			zap.Error(err),
		)
		return false
	}

	// override config from environment variables
	// note this only happens if environment variables are specified
	Settings.OverrideFromEnvironment()

	eswLogger.Info("Parsed configuration from the configuration file!",
		zap.String("Configuration file:", filename),
	)

	testModeEnvVar := os.Getenv("TEST_MODE")
	if (testModeEnvVar == "enabled") || (testModeEnabled) {
		Settings.TestMode = true
		eswLogger.Info(
			"ES worker is running in test mode with test hooks enabled.")
	}

	displayConfiguration()
	return true
}

// if ES_WORKER_CONFIG_FILE env var is specified, return value
// if env value is empty, use default
func getConfigFile() string {
	configFile := os.Getenv(envConfigFile)
	if configFile != "" {
		eswLogger.Info("Using config file override!",
			zap.String("Configuration file:", configFile),
		)
	} else {
		configFile = defaultConfigFile
	}
	return configFile
}

func GetLogger() *zap.Logger {
	return eswLogger
}

func GetServer() *Server {
	return &Settings.Server
}

func Shutdown() {
	shutdownLogger()
}

func displayConfiguration() {
	eswLogger.Info("HP Enrollment Worker Service - current configuration")
	eswLogger.Info("Server settings",
		zap.String(" - Hostname:", Settings.Server.Host),
		zap.Int(" - Rest Port:", Settings.Server.Port),
		zap.Bool(" - Debug logging:", Settings.Server.DebugLogRestRequests),
	)
	eswLogger.Info("CA settings",
		zap.String(" - CA Hostname:", Settings.CA.Host),
		zap.Int(" - CA Port:", Settings.CA.Port),
	)
	eswLogger.Info("DSTS settings",
		zap.String(" - DSTS Hostname:", Settings.DSTS.Host),
		zap.Int(" - DSTS Port:", Settings.DSTS.Port),
	)
	eswLogger.Info("Notification settings",
		zap.String(" - Endpoint:", Settings.Notification.Endpoint),
		zap.String(" - Pending enrollment queue:", Settings.Notification.PendingEnrollName),
		zap.String(" - Enroll queue:", Settings.Notification.EnrollName),
		zap.String(" - Pending registration queue:", Settings.Notification.PendingRegistrationName),
		zap.String(" - Error queue:", Settings.Notification.EnrollErrorName),
	)

}
