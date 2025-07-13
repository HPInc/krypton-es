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
	defaultConfigFile      = "config.yaml"
	defaultTokenConfigFile = "token_config.yaml" //#nosec G101
	defaultPolicyFile      = "default_policy.json"
	defaultServerName      = "127.0.0.1"
	defaultPort            = "7979"
	defaultLogLevel        = "info"
	envConfigFile          = "ES_CONFIG_FILE"
	envTokenConfigFile     = "ES_TOKEN_CONFIG_FILE" //#nosec G101
	envDefaultPolicyFile   = "ES_DEFAULT_POLICY_FILE"
)

var Settings Config

func Init() {
	initFlags()
	initLogger(defaultLogLevel)
}

func initFlags() {
	Settings.Flags.LogLevel = flag.String("log_level",
		defaultLogLevel,
		"Specify the logging level.")
	Settings.Flags.Version = flag.Bool("version", false,
		"Print the version of the service and exit!")

	// Parse the command line flags.
	flag.Parse()
	if *Settings.Flags.Version {
		printVersionInformation()
	}
}

func printVersionInformation() {
	fmt.Println("HP Device Enrollment Service: version information")
	fmt.Printf("- Git commit hash: %s\n - Built at: %s\n - Built by: %s\n - Built on: %s\n",
		Settings.Flags.gitCommitHash,
		Settings.Flags.builtAt,
		Settings.Flags.builtBy,
		Settings.Flags.builtOn)
}

func Load(testModeEnabled bool) bool {
	configFile := getConfigFile()

	// Open the configuration file for parsing.
	bytes, err := os.ReadFile(filepath.Clean(configFile))
	if err != nil {
		esLogger.Error("Failed to load configuration file!",
			zap.String("Configuration file:", configFile),
			zap.Error(err),
		)
		return false
	}

	// Read the configuration file and unmarshal the YAML.
	err = yaml.Unmarshal(bytes, &Settings)
	if err != nil {
		esLogger.Error("Failed to parse configuration file!",
			zap.String("Configuration file:", configFile),
			zap.Error(err),
		)
		return false
	}

	esLogger.Info("Parsed configuration from the configuration file!",
		zap.String("Configuration file:", configFile),
	)

	// override config from environment variables
	// note this only happens if environment variables are specified
	Settings.OverrideFromEnvironment()

	testModeEnvVar := os.Getenv("TEST_MODE")
	if (testModeEnvVar == "enabled") || (testModeEnabled) {
		Settings.TestMode = true
		fmt.Println("ES is running in test mode with test hooks enabled.")
		InitTestLogger()
	}

	return true
}

func Shutdown() {
	shutdownLogger()
}

func GetLogger() *zap.Logger {
	return esLogger
}

func IsSchemaMigrationMode() bool {
	return Settings.SchemaMigrationMode
}

func GetManagementServices() []string {
	return Settings.ManagementServices
}

// if ES_CONFIG_FILE env var is specified, return value
// if env value is empty, use default
func getConfigFile() string {
	configFile := os.Getenv(envConfigFile)
	if configFile != "" {
		esLogger.Info("Using config file override!",
			zap.String("Configuration file:", configFile),
		)
	} else {
		configFile = defaultConfigFile
	}
	return configFile
}

func GetTokenConfigFile() string {
	configFile := os.Getenv(envTokenConfigFile)
	if configFile != "" {
		esLogger.Info("Using token config file override!",
			zap.String("Token configuration file:", configFile),
		)
	} else {
		configFile = defaultTokenConfigFile
	}
	return configFile
}

func GetDefaultPolicyFile() string {
	policyFile := os.Getenv(envDefaultPolicyFile)
	if policyFile != "" {
		esLogger.Info("Using policy config file override!",
			zap.String("Policy file:", policyFile),
		)
	} else {
		policyFile = defaultPolicyFile
	}
	return policyFile
}

func GetJobsConfig() *ScheduledJobs {
	return &Settings.ScheduledJobs
}
