// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package config

import "go.uber.org/zap"

const ServiceName = "HP Device Enrollment Service"

type TokenType string

// Database server settings
type Database struct {
	// Hostname/IP address of the enroll database.
	Server string `yaml:"server"`
	// Database port
	Port int `yaml:"port"`
	// Database user
	User string `yaml:"user"`
	// Database password
	Password string `yaml:"password"`
	// Database name
	Name string `yaml:"name"`
	// Path to schema migration scripts for enroll database.
	SchemaMigrationScripts string `yaml:"schema"`
	// Whether to perform schema migration.
	SchemaMigrationEnabled bool `yaml:"migrate"`
	// expiry time in minutes for enroll records
	EnrollExpiryMinutes int `yaml:"enroll_expiry_minutes"`
	// max number of records to delete on expiry run
	EnrollExpiryDeleteLimit int `yaml:"enroll_expiry_delete_limit"`
	// Maximum number of open SQL connections
	MaxOpenConnections int `yaml:"max_open_connections"`
	// SSL mode to use for connections to the database.
	SslMode string `yaml:"ssl_mode"`
	// SSL root certificate to use for connections.
	SslRootCertificate string `yaml:"ssl_root_cert"`
}

// Cache server settings
type Cache struct {
	// The hostname/IP address of the enroll cache.
	Server string `yaml:"server"`
	// The port at which the cache is available.
	Port int `yaml:"port"`
	// Cache user
	User string `yaml:"user"`
	// Cache password
	Password string `yaml:"password"`
	// Whether enroll caching is enabled.
	Enabled bool `yaml:"enabled"`
	// The Redis database number to be used for the enroll cache.
	CacheDatabase int `yaml:"cache_db"`
	// strategy to use for retry-after hints
	RetryAfterHintStrategy string `yaml:"retry_after_hint_strategy"`
	// enroll time sliding window size
	EnrollTimeWindowSize int `yaml:"enroll_time_window_size"`
	// enroll update sliding window size
	EnrollUpdateWindowSize int `yaml:"enroll_update_window_size"`
}

// Configuration settings for the REST server.
type Server struct {
	Host string `yaml:"host"`
	// Port on which the REST service is available.
	Port int `yaml:"port"`
	// Max Retry-After default value
	MaxRetryAfterSeconds int `yaml:"max_retry_after_seconds"`
	// Retry-After default start value
	RetryAfterSeconds int `yaml:"retry_after_seconds"`
	// Debug rest requests
	DebugRestRequests bool `yaml:"debug_rest_requests"`
}

// Notification configuration settings
type Notification struct {
	// endpoint is only loaded from env. not from config.
	// endpoint is only applicable for local tests
	// cloud runs will load from cloud env
	Endpoint              string
	PendingEnrollName     string `yaml:"pending_enroll_name"`
	EnrollName            string `yaml:"enroll_name"`
	EnrollWatchDelay      int    `yaml:"enroll_watch_delay"`
	EnrollErrorName       string `yaml:"enroll_error_name"`
	EnrollErrorWatchDelay int    `yaml:"enroll_error_watch_delay"`
}

type ScheduledJob struct {
	Enabled bool   `yaml:"enabled"`
	Start   string `yaml:"start"`
	Every   string `yaml:"every"`
}

type ScheduledJobs map[string]ScheduledJob

type Config struct {
	// Rest Server
	Server Server
	// Notification Settings
	Notification Notification
	// Cache server
	Cache Cache
	// Database settings
	Database Database
	// Configuration settings for CA server connection
	// ES service creates a grpc client connection to CA server
	CA struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}
	// Configuration settings for DSTS server connection
	// es service creates a grpc client connection to DSTS server
	// for enrollment token validation
	DSTS struct {
		Host    string `yaml:"host"`
		RpcPort int    `yaml:"rpc_port"`
	}
	ManagementServices []string      `yaml:"management_services"`
	ScheduledJobs      ScheduledJobs `yaml:"scheduled_jobs"`
	Flags              struct {
		// --config_file: specifies the path to the configuration file.
		ConfigFile *string
		// --log_level: specify the logging level to use.
		LogLevel *string
		// --token_config_file: specify the path to token config file.
		TokenConfigFile *string
		// --version: displays versioning information.
		Version *bool
		//
		gitCommitHash string
		builtAt       string
		builtBy       string
		builtOn       string
	}
	Logger              *zap.Logger
	TestMode            bool
	SchemaMigrationMode bool
}
