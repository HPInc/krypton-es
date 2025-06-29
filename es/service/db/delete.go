// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"context"
	"fmt"
	"time"

	"github.com/HPInc/krypton-es/es/service/cache"
	"github.com/HPInc/krypton-es/es/service/metrics"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// delete enroll record by id
// this is usually called by processing after it
// determines an enroll record is no longer needed.
func DeleteEnroll(id uuid.UUID) error {
	start := time.Now()

	esLogger.Debug("Deleting enroll record",
		zap.String("id", id.String()))

	// optimistic cache purge
	cache.DeleteEnrollStatusById(id)

	ctx, cancelFunc := context.WithTimeout(context.Background(), dbTimeout)
	defer cancelFunc()

	tx, err := gDbPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx, ctx)

	_, err = tx.Exec(ctx, `DELETE FROM enroll WHERE id=$1`, id)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return err
	}

	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbDeleteEnroll)

	esLogger.Debug("Deleted enroll record",
		zap.String("id", id.String()))

	commit(tx, ctx)
	return nil
}

// entrypoint for scheduled delete expired enroll calls
func TriggerDeleteExpiredEnrolls() error {
	_, err := DeleteExpiredEnrolls(0)
	return err
}

// delete expired enroll records
// expiry criteria is controlled by service config
func DeleteExpiredEnrolls(enrollExpirySeconds int) (int64, error) {
	start := time.Now()
	if enrollExpirySeconds <= 0 {
		enrollExpirySeconds = gDbConfig.EnrollExpiryMinutes * 60
	}
	esLogger.Info("Deleting expired enroll records",
		zap.Int("expired_since", enrollExpirySeconds),
		zap.Int("delete_limit", gDbConfig.EnrollExpiryDeleteLimit))
	ctx, cancelFunc := context.WithTimeout(context.Background(), dbTimeout)
	defer cancelFunc()

	tx, err := gDbPool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer rollback(tx, ctx)
	sql := fmt.Sprintf(
		`DELETE FROM enroll WHERE id IN (
		SELECT id FROM enroll WHERE 
		created_at < NOW() - INTERVAL '%d seconds' LIMIT $1)`,
		enrollExpirySeconds)
	result, err := tx.Exec(ctx, sql, gDbConfig.EnrollExpiryDeleteLimit)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return 0, err
	}

	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbDeleteExpiredEnrolls)

	esLogger.Info("Deleted expired enroll records",
		zap.Int64("count", result.RowsAffected()),
		zap.Int("expired_since", enrollExpirySeconds))

	commit(tx, ctx)
	return result.RowsAffected(), nil
}
