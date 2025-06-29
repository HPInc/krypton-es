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

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// Ping the database to verify DSN provided by the user is valid and the
// server accessible. If the ping fails exit the program with an error.
func Ping() error {
	ctx, cancelFunc := context.WithTimeout(gCtx, dbTimeout)
	defer cancelFunc()
	if err := gDbPool.Ping(ctx); err != nil {
		esLogger.Error("DB: Connection failed", zap.Error(err))
		return err
	}
	return nil
}

// create entry for incoming device enroll
func CreateEnrollRecord(tenantId, userId, csrHash string) (*structs.DeviceEntry, error) {
	start := time.Now()
	de := structs.DeviceEntry{TenantId: tenantId, UserId: userId}
	ctx, cancelFunc := context.WithTimeout(gCtx, dbTimeout)
	defer cancelFunc()
	err := gDbPool.QueryRow(ctx,
		`INSERT INTO enroll(tenant_id, user_id, csr_hash)
		VALUES($1,$2,$3) RETURNING id, request_id`,
		tenantId, userId, csrHash).Scan(&de.Id, &de.RequestId)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return nil, err
	}
	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbCreateEnroll)
	go cache.CreateEnrollStatus(de.Id, tenantId, userId, uuid.Nil, 0)
	go cache.SetCsrHash(csrHash)
	return &de, nil
}

// entry with certificate
func UpdateEnrollRecord(dc *structs.EnrollResult) error {
	var elapsed float64
	start := time.Now()
	ctx, cancelFunc := context.WithTimeout(gCtx, dbTimeout)
	defer cancelFunc()
	err := gDbPool.QueryRow(ctx,
		`UPDATE enroll SET device_id=$1, certificate=$2,
		parent_certificates=$3, updated_at=now(), status=1 WHERE id=$4
		RETURNING extract (epoch from (updated_at - created_at))`,
		dc.DeviceId, dc.Certificate, dc.ParentCertificates,
		dc.EnrollId).Scan(&elapsed)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return err
	}
	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbUpdateEnroll)
	// update enroll time for average
	go cache.AddEnrollElapsed(elapsed)
	// set status in cache
	cache.SetEnrollStatus(dc.EnrollId, dc.DeviceId, 1)
	return nil
}

// return status for id
func GetEnrollStatus(id uuid.UUID) (*structs.EnrollStatus, error) {
	var err error
	var entry *structs.EnrollStatus = &structs.EnrollStatus{}
	start := time.Now()
	if cached, err := cache.GetEnrollStatus(id); err == nil {
		metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency,
			start, operationDbGetStatusById)
		return cached, nil
	}
	ctx, cancelFunc := context.WithTimeout(gCtx, dbTimeout)
	defer cancelFunc()

	err = gDbPool.QueryRow(ctx,
		"SELECT status, device_id, tenant_id, user_id FROM enroll WHERE id=$1",
		id).Scan(&entry.Status, &entry.DeviceId, &entry.TenantId, &entry.UserId)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return nil, err
	}
	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbGetStatusById)
	// set status in cache
	go cache.SetEnrollStatus(id, entry.DeviceId, entry.Status)
	return entry, nil
}

// get enroll details
func GetEnrollDetailsById(id uuid.UUID) (*structs.EnrollResult, error) {
	start := time.Now()
	dc := structs.EnrollResult{EnrollId: id, Id: id}
	ctx, cancelFunc := context.WithTimeout(gCtx, dbTimeout)
	defer cancelFunc()
	err := gDbPool.QueryRow(ctx,
		`SELECT device_id, certificate, parent_certificates FROM enroll
		WHERE id=$1`,
		id).Scan(&dc.DeviceId, &dc.Certificate, &dc.ParentCertificates)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
	}
	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbGetDetails)
	return &dc, err
}

// get pending enroll count
func GetPendingEnrollCount() (int, error) {
	var count int
	start := time.Now()
	ctx, cancelFunc := context.WithTimeout(gCtx, dbTimeout)
	defer cancelFunc()
	err := gDbPool.QueryRow(ctx,
		"SELECT count(*) FROM enroll WHERE status=0").Scan(&count)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return -1, err
	}
	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbGetPendingEnrollCount)
	return count, err
}

// is csr hash already in gDbPool
func HasCSRHash(csrHash string) (bool, error) {
	var count int
	start := time.Now()
	if exists, err := cache.HasCsrHash(csrHash); err == nil {
		metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency,
			start, operationDbCheckCSRHash)
		return exists, nil
	}
	ctx, cancelFunc := context.WithTimeout(gCtx, dbTimeout)
	defer cancelFunc()
	err := gDbPool.QueryRow(ctx,
		"SELECT count(*) FROM enroll WHERE csr_hash=$1",
		csrHash).Scan(&count)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return false, err
	}
	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbCheckCSRHash)
	if count == 1 {
		go cache.SetCsrHash(csrHash)
	}
	return count == 1, err
}

// get average enroll time
func GetAverageEnrollTime() (int, error) {
	t, err := cache.GetAverageEnrollTime()
	if err != nil {
		esLogger.Error("failed cache fetch for average enroll",
			zap.Error(err))
	}
	esLogger.Info("average enroll time", zap.Int("seconds", t))
	return t, nil
}
