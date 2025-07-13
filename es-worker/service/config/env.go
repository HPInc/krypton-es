// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package config

import (
	"os"
	"strconv"

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
		// SERVER
		"ESW_SERVER_PORT":                    {v: &c.Server.Port},
		"ESW_SERVER_DEBUG_LOG_REST_REQUESTS": {v: &c.Server.DebugLogRestRequests},

		//NOTIFICATION
		"ESW_NOTIFICATION_ENDPOINT":                         {v: &c.Notification.Endpoint},
		"ESW_NOTIFICATION_PENDING_ENROLL_NAME":              {v: &c.Notification.PendingEnrollName},
		"ESW_NOTIFICATION_PENDING_ENROLL_WATCH_DELAY":       {v: &c.Notification.PendingEnrollWatchDelay},
		"ESW_NOTIFICATION_PENDING_REGISTRATION_NAME":        {v: &c.Notification.PendingRegistrationName},
		"ESW_NOTIFICATION_PENDING_REGISTRATION_WATCH_DELAY": {v: &c.Notification.PendingRegistrationWatchDelay},
		"ESW_NOTIFICATION_ENROLL_NAME":                      {v: &c.Notification.EnrollName},
		"ESW_NOTIFICATION_ENROLL_ERROR_NAME":                {v: &c.Notification.EnrollErrorName},
		// operation mode
		"ESW_OPERATION_MODE": {v: &c.OperationMode},
		// ca
		"ESW_CA_HOST":     {v: &c.CA.Host},
		"ESW_CA_RPC_PORT": {v: &c.CA.Port},
		// dsts
		"ESW_DSTS_HOST":     {v: &c.DSTS.Host},
		"ESW_DSTS_RPC_PORT": {v: &c.DSTS.Port},
	}
	for k, v := range m {
		e := os.Getenv(k)
		if e != "" {
			eswLogger.Info("Overriding env variable",
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
	case *int:
		i, err := strconv.Atoi(envValue)
		if err != nil {
			eswLogger.Error("Bad integer value in env")
		} else {
			*t.v.(*int) = i
		}
	}
}

func getLoggableValue(secret bool, value string) string {
	if secret {
		return "***"
	}
	return value
}
