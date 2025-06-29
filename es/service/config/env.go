// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package config

import (
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type value struct {
	secret bool
	v      interface{}
}

// loadEnvironmentVariableOverrides - check values specified for supported
// environment variables. These can be used to override configuration settings
// specified in the config file.
func (c *Config) OverrideFromEnvironment() {
	m := map[string]value{
		//Server
		"ES_SERVER":                  {v: &c.Server.Host},
		"ES_PORT":                    {v: &c.Server.Port},
		"ES_MAX_RETRY_AFTER_SECONDS": {v: &c.Server.MaxRetryAfterSeconds},
		"ES_RETRY_AFTER_SECONDS":     {v: &c.Server.RetryAfterSeconds},
		"ES_DEBUG_REST_REQUESTS":     {v: &c.Server.DebugRestRequests},

		//DSTS
		"ES_DSTS_HOST":     {v: &c.DSTS.Host},
		"ES_DSTS_RPC_PORT": {v: &c.DSTS.RpcPort},
		//DB
		"ES_DB_SERVER":                     {v: &c.Database.Server},
		"ES_DB_PORT":                       {v: &c.Database.Port},
		"ES_DB_USER":                       {v: &c.Database.User},
		"ES_DB_PASSWORD":                   {secret: true, v: &c.Database.Password},
		"ES_DB_NAME":                       {v: &c.Database.Name},
		"ES_DB_SCHEMA_MIGRATION_SCRIPTS":   {v: &c.Database.SchemaMigrationScripts},
		"ES_DB_SCHEMA_MIGRATION_ENABLED":   {v: &c.Database.SchemaMigrationEnabled},
		"ES_DB_ENROLL_EXPIRY_MINUTES":      {v: &c.Database.EnrollExpiryMinutes},
		"ES_DB_ENROLL_EXPIRY_DELETE_LIMIT": {v: &c.Database.EnrollExpiryDeleteLimit},
		"ES_DB_SSL_MODE":                   {v: &c.Database.SslMode},
		"ES_DB_SSL_ROOT_CERT":              {v: &c.Database.SslRootCertificate},
		// Notification settings
		"ES_NOTIFICATION_ENDPOINT":                 {v: &c.Notification.Endpoint},
		"ES_NOTIFICATION_PENDING_ENROLL_NAME":      {v: &c.Notification.PendingEnrollName},
		"ES_NOTIFICATION_ENROLL_NAME":              {v: &c.Notification.EnrollName},
		"ES_NOTIFICATION_ENROLL_WATCH_DELAY":       {v: &c.Notification.EnrollWatchDelay},
		"ES_NOTIFICATION_ENROLL_ERROR_NAME":        {v: &c.Notification.EnrollErrorName},
		"ES_NOTIFICATION_ENROLL_ERROR_WATCH_DELAY": {v: &c.Notification.EnrollErrorWatchDelay},
		//CACHE
		"ES_CACHE_SERVER":                    {v: &c.Cache.Server},
		"ES_CACHE_PORT":                      {v: &c.Cache.Port},
		"ES_CACHE_USER":                      {v: &c.Cache.User},
		"ES_CACHE_PASSWORD":                  {secret: true, v: &c.Cache.Password},
		"ES_CACHE_ENABLED":                   {v: &c.Cache.Enabled},
		"ES_CACHE_RETRY_AFTER_HINT_STRATEGY": {v: &c.Cache.RetryAfterHintStrategy},
		"ES_CACHE_ENROLL_TIME_WINDOW_SIZE":   {v: &c.Cache.EnrollTimeWindowSize},
		"ES_CACHE_ENROLL_UPDATE_WINDOW_SIZE": {v: &c.Cache.EnrollUpdateWindowSize},
		//Management services
		"ES_MANAGEMENT_SERVICES": {v: &c.ManagementServices},
		// modes of operation
		"ES_MODE_SCHEMA_MIGRATION": {v: &c.SchemaMigrationMode},
	}
	for k, v := range m {
		e := os.Getenv(k)
		if e != "" {
			esLogger.Info("Overriding env variable",
				zap.String("variable: ", k),
				zap.String("value: ", getLoggableValue(v.secret, e)))
			val := v
			replaceConfigValue(os.Getenv(k), &val)
		}
	}
}

// envValue will be non empty as this function is private to file
func replaceConfigValue(envValue string, t *value) {
	switch t.v.(type) {
	case *string:
		*t.v.(*string) = envValue
	case *[]string:
		*t.v.(*[]string) = strings.Split(envValue, ",")
	case *bool:
		b, err := strconv.ParseBool(envValue)
		if err != nil {
			esLogger.Error("Bad bool value in env")
		} else {
			*t.v.(*bool) = b
		}
	case *int:
		i, err := strconv.Atoi(envValue)
		if err != nil {
			esLogger.Error("Bad integer value in env",
				zap.Error(err))
		} else {
			*t.v.(*int) = i
		}
	default:
		esLogger.Error("There was a bad type map in env override",
			zap.String("value", envValue))
	}
}

func getLoggableValue(secret bool, value string) string {
	if secret {
		return "***"
	}
	return value
}
