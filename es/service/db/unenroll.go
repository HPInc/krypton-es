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

// unenroll
// 1. create new unenroll record denoting an unenroll entry
// this db entry will be used to track the queue result
func Unenroll(tenantId string, deviceId uuid.UUID) (
	*structs.DeviceEntry, error) {
	start := time.Now()

	de := structs.DeviceEntry{}
	ctx, cancelFunc := context.WithTimeout(context.Background(), dbTimeout)
	defer cancelFunc()
	err := gDbPool.QueryRow(ctx, `INSERT INTO unenroll(tenant_id, device_id)
		VALUES($1,$2) RETURNING id, request_id`,
		tenantId, deviceId).Scan(&de.Id, &de.RequestId)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return nil, err
	}

	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbUnenroll)
	go cache.CreateUnenrollStatus(de.Id, tenantId, deviceId.String(), 0)
	return &de, nil
}

// update unenroll entry as complete
func UpdateUnenrollRecord(res *structs.UnenrollResult) error {
	var elapsed float64
	start := time.Now()
	ctx, cancelFunc := context.WithTimeout(context.Background(), dbTimeout)
	defer cancelFunc()
	err := gDbPool.QueryRow(ctx,
		`UPDATE unenroll SET
		updated_at=now(), status=1 WHERE id=$1
		RETURNING extract (epoch from (updated_at - created_at))`,
		res.UnenrollId).Scan(&elapsed)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return err
	}
	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbUpdateUnenroll)
	// update unenroll time for average
	cache.AddUnenrollElapsed(elapsed)
	// set status in cache
	go cache.SetUnenrollStatus(res.UnenrollId, 1)
	return nil
}

// return status for id
func GetUnenrollStatus(id uuid.UUID) (*structs.UnenrollStatus, error) {
	var err error
	entry := &structs.UnenrollStatus{}
	start := time.Now()
	if cached, err := cache.GetUnenrollStatus(id); err == nil {
		metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency,
			start, operationDbGetStatusById)
		return cached, nil
	}
	ctx, cancelFunc := context.WithTimeout(gCtx, dbTimeout)
	defer cancelFunc()
	err = gDbPool.QueryRow(ctx,
		"SELECT status, tenant_id, device_id FROM unenroll WHERE id=$1",
		id).Scan(&entry.Status, &entry.TenantId, &entry.DeviceId)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return nil, err
	}
	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbGetStatusById)
	// set status in cache
	go cache.SetUnenrollStatus(id, entry.Status)
	return entry, nil
}
