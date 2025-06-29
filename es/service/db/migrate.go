// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/HPInc/krypton-es/es/service/config"
)

var (
	connStrBase = "postgres://%s:%s@%s:%d/%s?sslmode=%s"
)

// Migrates the database schema using the migration scripts specified in the
// configuration file.
func migrateDatabaseSchema(dbConfig *config.Database) error {
	// If database schema migration is disabled, do nothing.
	if !dbConfig.SchemaMigrationEnabled {
		esLogger.Info("Database schema migration is disabled. Skipping ...")
		return nil
	}

	esLogger.Info("Starting database schema migration ...",
		zap.String("Migration script location: ", dbConfig.SchemaMigrationScripts),
	)

	connStr := fmt.Sprintf(connStrBase, dbConfig.User, dbConfig.Password,
		dbConfig.Server, dbConfig.Port, dbConfig.Name, dbConfig.SslMode)

	// Open the database for migration.
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		esLogger.Error("Failed to open database for schema migration!",
			zap.Error(err),
		)
		return err
	}
	defer db.Close()

	// Initialize the migration driver for Postgres.
	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		esLogger.Error("Failed to connect to the database instance for migration!",
			zap.Error(err),
		)
		return err
	}

	mig, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s",
		dbConfig.SchemaMigrationScripts), databaseName, driver)
	if err != nil {
		esLogger.Error("Failed to initialize a new migration instance!",
			zap.Error(err),
		)
		return err
	}

	// Attempt to migrate up the schema for the database. If we are currently
	// at the highest available schema, migration with fail with the error code
	// ErrNoChange. In this case, there is no migration to be performed and we
	// are good to proceed.
	err = mig.Up()
	if err != nil && err != migrate.ErrNoChange {
		esLogger.Error("Failed to upgrade database schema!",
			zap.Error(err),
		)
		return err
	}

	esLogger.Info("Successfully completed schema migration for the database!")
	return nil
}
