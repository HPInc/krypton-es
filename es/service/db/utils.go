// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"context"
	"errors"

	"github.com/HPInc/krypton-es/es/service/metrics"
	pgx "github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var (
	ErrNoRows = pgx.ErrNoRows
)

func IsDbErrorNoRows(err error) bool {
	return err == ErrNoRows
}

func commit(tx pgx.Tx, ctx context.Context) {
	err := tx.Commit(ctx)
	if err != nil {
		esLogger.Error("Failed to commit transaction!",
			zap.Error(err),
		)
		metrics.MetricDatabaseCommitErrors.Inc()
	}
}

func rollback(tx pgx.Tx, ctx context.Context) {
	if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		esLogger.Error("Failed to rollback transaction!",
			zap.Error(err),
		)
		metrics.MetricDatabaseRollbackErrors.Inc()
	}
}
