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

// move enroll record to error
// find existing entry and move to enroll_error
// in addition, add error code and error message
func FailEnrollRecord(ee *structs.EnrollError) error {
	defer metrics.ReportLatencyMetric(
		metrics.MetricDatabaseLatency, time.Now(),
		operationDbFailEnroll)

	esLogger.Info("Failing enroll entry",
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
	if _, err = tx.Exec(ctx, `INSERT INTO enroll_error (
		id, request_id, tenant_id, user_id, csr_hash, status, device_id,
		certificate, error_code, error_text)
		(SELECT id, request_id, tenant_id, user_id, csr_hash, status, device_id,
		certificate, $1, $2 FROM enroll WHERE id=$3)`,
		ee.ErrorCode, ee.ErrorMessage, ee.EnrollId); err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return err
	}

	// delete enroll record
	if _, err = tx.Exec(ctx, "DELETE FROM enroll where id=$1", ee.EnrollId); err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return err
	}

	commit(tx, ctx)
	return nil
}
