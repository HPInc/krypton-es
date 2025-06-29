// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package main

import (
	"os"

	dstsclient "github.com/HPInc/krypton-es/es/service/client/dsts"
	"github.com/HPInc/krypton-es/es/service/config"
	"github.com/HPInc/krypton-es/es/service/db"
	"github.com/HPInc/krypton-es/es/service/jobs"
	"github.com/HPInc/krypton-es/es/service/metrics"
	"github.com/HPInc/krypton-es/es/service/notification"
	"github.com/HPInc/krypton-es/es/service/policy"
	"github.com/HPInc/krypton-es/es/service/rest"
	"github.com/HPInc/krypton-es/es/service/tokenmgr"
)

func main() {
	// Initialize structured logging.
	config.Init()
	defer config.Shutdown()

	// Read and parse the configuration file.
	if !config.Load(false) {
		panic("config load failed.")
	}
	metrics.RegisterPrometheusMetrics()

	// init database
	if db.Init(config.GetLogger(), &config.Settings.Database) != nil {
		panic("database init failed.")
	}
	defer db.Shutdown()

	checkMigrationMode()

	// Initialize dsts client connection
	if dstsclient.Init(config.GetLogger()) != nil {
		panic("dsts client failed to connect")
	}
	defer dstsclient.Close()

	// init notification client
	if notification.Init(
		config.GetLogger(), &config.Settings.Notification) != nil {
		panic("notifcation client failed.")
	}
	defer notification.Shutdown()

	// init token manager
	if tokenmgr.Init(config.GetLogger(),
		config.GetTokenConfigFile()) != nil {
		panic("Failed to initialize token manager.")
	}
	defer tokenmgr.Shutdown()

	// init policy
	if policy.Init(config.GetLogger(),
		config.GetDefaultPolicyFile()) != nil {
		panic("Failed to initialize policy.")
	}

	// init scheduled jobs
	if jobs.Init(config.GetLogger(), config.GetJobsConfig()) != nil {
		panic("Failed to initialize jobs.")
	}
	defer jobs.Shutdown()

	// Initialize the REST server and start listening for requests at the
	// enroll service endpoint.
	if rest.Init(config.GetLogger(), &config.Settings.Server) != nil {
		panic("server start failed.")
	}
	defer rest.Shutdown()

	rest.WaitForEvents()

	config.GetLogger().Info("HP Enroll Service: Goodbye!")
}

func checkMigrationMode() {
	if config.IsSchemaMigrationMode() {
		config.GetLogger().Info(
			"HP Enroll Service: Exit after migration")
		os.Exit(0)
	}
}
