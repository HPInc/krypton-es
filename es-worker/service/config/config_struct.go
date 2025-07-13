// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package config

import "go.uber.org/zap"

// Server represents contains configuration settings for the ESW http server.
type Server struct {
	// Hostname of the ESW service.
	Host string `yaml:"host"`

	// Port on which the REST server is available.
	Port int `yaml:"port"`

	// Control debug logging on REST requests.
	DebugLogRestRequests bool `yaml:"log_rest_requests"`
}

// Notification configuration settings
type Notification struct {
	// endpoint is set via env var as needed
	// mostly only needed for local runs
	Endpoint                      string
	PendingEnrollName             string `yaml:"pending_enroll_name"`
	PendingEnrollWatchDelay       int    `yaml:"pending_enroll_watch_delay"`
	PendingRegistrationName       string `yaml:"pending_registration_name"`
	PendingRegistrationWatchDelay int    `yaml:"pending_registration_watch_delay"`
	EnrollName                    string `yaml:"enroll_name"`
	EnrollErrorName               string `yaml:"enroll_error_name"`
}

type Config struct {
	// Server settings
	Server Server

	// Notification settings
	Notification Notification

	// Configuration settings for CA server connection
	// es-worker service creates a grpc client connection to CA server
	CA struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}
	// Configuration settings for DSTS server connection
	// es-worker service creates a grpc client connection to DSTS server
	DSTS struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}
	// Structured logging configuration settings.
	Logging struct {
		// Default logging level to use.
		LogLevel string `yaml:"log_level"`
	} `yaml:"logging"`
	Flags struct {
		// --config_file: specifies the path to the configuration file.
		ConfigFile *string
		// --log_file: specifies the path to be used to store log files.
		LogFile *string
		// --version: displays versioning information.
		Version *bool
		//
		gitCommitHash string
		builtAt       string
		builtBy       string
		builtOn       string
	}
	Logger        *zap.Logger
	TestMode      bool
	OperationMode string `yaml:"operation_mode"`
}
