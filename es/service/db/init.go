// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/HPInc/krypton-es/es/service/cache"
	"github.com/HPInc/krypton-es/es/service/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	// Name of the database hosting device objects.
	databaseName = "es"

	// Maximum number of connection retries.
	maxDbConnectionRetries = 3

	// Database connection retry interval
	connectionRetryInterval                = (time.Second * 5)
	dbTimeout                              = (time.Second * 3)
	defaultIdleInTransactionSessionTimeout = (time.Second * 10)
	defaultStatementTimeout                = (time.Second * 10)

	// db metrics
	operationDbCreateEnroll               = "create_enroll"
	operationDbUpdateEnroll               = "update_enroll"
	operationDbRenewEnroll                = "renew_enroll"
	operationDbDeleteEnroll               = "delete_enroll"
	operationDbGetStatusById              = "get_status_by_id"
	operationDbGetDetails                 = "get_enroll_details"
	operationDbGetPendingEnrollCount      = "get_pending_enroll_count" //#nosec G101
	operationDbCheckCSRHash               = "check_csr_hash"
	operationDbGetStatusByTenantAndDevice = "get_status_by_tenant_and_device"
	operationDbFailEnroll                 = "failed_enroll"
	operationDbUnenroll                   = "unenroll"
	operationDbUpdateUnenroll             = "update_unenroll"
	operationDbFailUnenroll               = "failed_unenroll"
	operationDbGetPublicKey               = "get_publickey"
	operationDbSetPublicKey               = "set_publickey"
	operationDbCreatePolicy               = "create_policy"
	operationDbGetPolicy                  = "get_policy"
	operationDbUpdatePolicy               = "update_policy"
	operationDbGetPolicyByTenant          = "get_policy_by_tenant"
	// internal calls
	operationDbDeleteExpiredEnrolls = "delete_expired_enrolls"
)

var (
	// Structured logging using Uber Zap.
	esLogger *zap.Logger

	// Connection to the devices database.
	gDbPool *pgxpool.Pool

	// global context
	gCtx context.Context

	// Connection string for the Postgres enroll database.
	postgresDsn = "host=%s port=%d user=%s dbname=%s password=%s sslmode=%s"

	// db config
	gDbConfig *config.Database
)

func Init(logger *zap.Logger, dbConfig *config.Database) error {
	esLogger = logger
	gDbConfig = dbConfig

	gCtx = context.Background()

	// Configure the connection to the enroll database. We configure the SSL
	// mode as disabled.
	dsn := fmt.Sprintf(postgresDsn,
		gDbConfig.Server,
		gDbConfig.Port,
		gDbConfig.User,
		gDbConfig.Name,
		gDbConfig.Password,
		gDbConfig.SslMode)

	// Connect to the database and initialize it.
	err := loadEnrollDatabase(dsn)
	if err != nil {
		esLogger.Error("Failed to initialize enroll database!",
			zap.Error(err),
		)
		return err
	}

	// init cache
	if err = cache.Init(esLogger, &config.Settings.Cache); err != nil {
		esLogger.Error("Failed to initialize cache!",
			zap.Error(err),
		)
		return err
	}

	return nil
}

// Shutdown - close the connection to the device database. Also close the
// connection to the device cache.
func Shutdown() {
	// Shutdown the device database and close connections.
	shutdownEnrollDatabase()
	cache.Shutdown()
	gCtx.Done()
}

// Shutdown the connection to the device database.
func shutdownEnrollDatabase() {
	if gDbPool == nil {
		esLogger.Info("Not shutting down db as it was not initialized")
	} else {
		gDbPool.Close()
		esLogger.Info("Database successfully shutdown")
		gDbPool = nil
	}
}

func loadEnrollDatabase(dsn string) error {
	var (
		err           error
		dbInitialized = false
		tlsConfig     *tls.Config
	)

	// Load the root CA certificates and initialize TLS configuration.
	if gDbConfig.SslMode != "disable" {
		certs, err := loadTlsCert(gDbConfig.SslRootCertificate)
		if err != nil {
			esLogger.Error("Failed to load the root CA certificate for SSL connections to the database!",
				zap.String("SSL Root CA path", gDbConfig.SslRootCertificate),
				zap.Error(err),
			)
			return err
		}

		tlsConfig = &tls.Config{
			RootCAs:    certs,
			ServerName: gDbConfig.Server,
			MinVersion: tls.VersionTLS12,
		}
	}

	pgxConfig, err := initPgxConfig(dsn)
	if err != nil {
		esLogger.Error("Failed to initialize database connection configuration!",
			zap.String("Database server: ", gDbConfig.Server),
			zap.Error(err),
		)
		return err
	}
	pgxConfig.ConnConfig.TLSConfig = tlsConfig

	// Give ourselves 3 retry attempts to connect to the enroll database.
	for i := maxDbConnectionRetries; i > 0; i-- {
		ctx, cancelFunc := context.WithTimeout(gCtx, dbTimeout)
		defer cancelFunc()
		gDbPool, err = pgxpool.New(ctx, dsn)
		if err != nil {
			esLogger.Error("Failed to connect to enroll database!",
				zap.Error(err),
			)
			time.Sleep(connectionRetryInterval)
		} else {
			if err = Ping(); err != nil {
				esLogger.Error("Failed to ping the files database!",
					zap.String("Database host: ", gDbConfig.Server),
					zap.Error(err),
				)
				gDbPool.Close()
				time.Sleep(connectionRetryInterval)
			} else {
				dbInitialized = true
				break
			}
		}
	}

	if !dbInitialized {
		esLogger.Error("All retry attempts to load enroll database exhausted. Giving up!",
			zap.Error(err),
		)
		return err
	}

	// Perform database schema migrations.
	err = migrateDatabaseSchema(gDbConfig)
	if err != nil {
		esLogger.Error("Failed to migrate database schema for enroll database!",
			zap.String("Database server: ", gDbConfig.Server),
			zap.Error(err),
		)
		shutdownEnrollDatabase()
		return err
	}

	esLogger.Info("Connected to the enroll database!",
		zap.String("Database server: ", gDbConfig.Server),
		zap.Int("Database port: ", gDbConfig.Port),
	)
	return nil
}

// Initialize Pgx configuration settings to connect to database.
func initPgxConfig(connStr string) (*pgxpool.Config, error) {
	pgxConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	if gDbConfig.MaxOpenConnections == 0 {
		pgxConfig.MaxConns = int32(runtime.NumCPU()) * 5
	} else {
		pgxConfig.MaxConns = int32(gDbConfig.MaxOpenConnections)
	}

	runtimeParams := pgxConfig.ConnConfig.RuntimeParams
	runtimeParams["application_name"] = config.ServiceName
	runtimeParams["idle_in_transaction_session_timeout"] =
		strconv.Itoa(int(defaultIdleInTransactionSessionTimeout.Milliseconds()))
	runtimeParams["statement_timeout"] =
		strconv.Itoa(int(defaultStatementTimeout.Milliseconds()))

	return pgxConfig, nil
}

func loadTlsCert(rootCertPath string) (*x509.CertPool, error) {
	certs := x509.NewCertPool()

	pemData, err := os.ReadFile(filepath.Clean(rootCertPath))
	if err != nil {
		esLogger.Error("Failed to read the root CA certificate file!",
			zap.String("CA certificate path", rootCertPath),
			zap.Error(err),
		)
		return nil, err
	}
	if !certs.AppendCertsFromPEM(pemData) {
		esLogger.Error("Failed to read the root CA certificate file!",
			zap.String("CA certificate path", rootCertPath),
			zap.Error(err),
		)
		return nil, errors.New("failed to append root ca cert")
	}

	return certs, nil
}
