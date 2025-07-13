// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/HPInc/krypton-es/es-worker/service/config"
	"github.com/HPInc/krypton-es/es-worker/service/notification"
	"github.com/HPInc/krypton-es/es-worker/service/rest"
)

func main() {
	// Initialize structured logging.
	config.Init()
	defer config.Shutdown()

	// Read and parse the configuration file.
	status := config.Load(false)
	if !status {
		panic("config load failed. cannot continue.")
	}

	eswLogger := config.GetLogger()

	// init notification client
	if notification.Init(eswLogger, &config.Settings.Notification) != nil {
		panic("notifcation client failed.")
	}
	defer notification.Shutdown()

	// init rest server for health and metrics
	rest.Init(eswLogger, config.GetServer())

	eswLogger.Info("HP CEM Enroll Worker: Goodbye!")
}
