// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"context"
	"time"

	"github.com/HPInc/krypton-es/es/service/cache"
	"github.com/HPInc/krypton-es/es/service/metrics"
	"github.com/HPInc/krypton-es/es/service/structs"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// renew enroll
// 1. find existing entry and move to enroll_archive
// 2. create new enroll record
func RenewEnroll(tenantId string, deviceId uuid.UUID, userId, csrHash string) (
	*structs.DeviceEntry, error) {
	defer metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, time.Now(),
		operationDbRenewEnroll)

	de := structs.DeviceEntry{}
	ctx, cancelFunc := context.WithTimeout(gCtx, dbTimeout)
	defer cancelFunc()
	err := gDbPool.QueryRow(ctx, `INSERT INTO enroll(tenant_id, user_id, device_id, csr_hash)
		VALUES($1,$2, $3, $4) RETURNING id, request_id`,
		tenantId, userId, deviceId, csrHash).Scan(&de.Id, &de.RequestId)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return nil, err
	}
	go cache.CreateEnrollStatus(de.Id, tenantId, userId, deviceId, 0)
	go cache.SetCsrHash(csrHash)

	return &de, nil
}
