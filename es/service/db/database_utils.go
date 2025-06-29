// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"context"
	"os"

	"github.com/HPInc/krypton-es/es/service/config"
	"go.uber.org/zap"
)

const (
	defaultPostgresUser = "postgres"
	defaultPostgresPass = "postgres"
	defaultCacheUser    = "krypton"
	defaultCachePass    = "krypton"
	invalidEnrollStatus = -1
	SslModeDisabled     = "disable"
)

var (
	testDbConfig config.Database = config.Database{
		Server:                  getServer(),
		Port:                    5432,
		User:                    defaultPostgresUser,
		Password:                defaultPostgresPass,
		Name:                    "es",
		SchemaMigrationEnabled:  true,
		SslMode:                 SslModeDisabled,
		EnrollExpiryDeleteLimit: 5,
	}
)

func getServer() string {
	e := os.Getenv("ES_DB_SERVER")
	if e != "" {
		return e
	} else {
		return "localhost"
	}
}

func getCacheServer() string {
	e := os.Getenv("ES_CACHE_SERVER")
	if e != "" {
		return e
	} else {
		return "localhost"
	}
}

func getMigrationScripts() string {
	e := os.Getenv("ES_DB_SCHEMA_MIGRATION_SCRIPTS")
	if e != "" {
		return e
	} else {
		return "schema"
	}
}

// init for test
func InitTest(logger *zap.Logger, dbConfig *config.Database) error {
	dbConfig.SchemaMigrationScripts = getMigrationScripts()
	config.Settings.Cache = config.Cache{
		Server:   getCacheServer(),
		Port:     6379,
		User:     defaultCacheUser,
		Password: defaultCachePass,
		Enabled:  true,
	}
	return Init(logger, dbConfig)
}

func InitTestDefault(logger *zap.Logger) error {
	var err error
	if err = InitTest(logger, &testDbConfig); err != nil {
		return err
	}
	return Ping()
}

func CountEnrollRecords() (int, error) {
	ctx, cancelFunc := context.WithTimeout(gCtx, dbTimeout)
	defer cancelFunc()

	i := 0
	err := gDbPool.QueryRow(ctx, "SELECT count(*) FROM enroll").Scan(&i)
	return i, err
}
