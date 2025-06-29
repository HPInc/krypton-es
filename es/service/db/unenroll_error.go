// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"context"
	"time"

	"github.com/HPInc/krypton-es/es/service/metrics"
	"github.com/HPInc/krypton-es/es/service/structs"
	"go.uber.org/zap"
)

// move unenroll record to error
// find existing entry and move to enroll_error
// in addition, add error code and error message
func FailUnenrollRecord(ee *structs.EnrollError) error {
	defer metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, time.Now(),
		operationDbFailUnenroll)
	esLogger.Info("Failing unenroll entry",
		zap.String("enroll_id:", ee.EnrollId),
		zap.Int("code:", ee.ErrorCode),
		zap.String("error:", ee.ErrorMessage))

	ctx, cancel := context.WithTimeout(gCtx, dbTimeout)
	defer cancel()

	tx, err := gDbPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx, ctx)

	// make error record
	if _, err = tx.Exec(ctx, `INSERT INTO unenroll_error (
		id, request_id, tenant_id, device_id, status,
		error_code, error_text)
		(SELECT id, request_id, tenant_id, device_id, status,
		$1, $2 FROM unenroll WHERE id=$3)`,
		ee.ErrorCode, ee.ErrorMessage, ee.EnrollId); err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return err
	}

	// delete unenroll record
	if _, err = tx.Exec(ctx, "DELETE FROM unenroll where id=$1", ee.EnrollId); err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return err
	}

	commit(tx, ctx)
	return nil
}
