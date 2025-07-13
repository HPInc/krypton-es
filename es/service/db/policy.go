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
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

// create new policy record
// tenantId = tenant id
// data = policy data as a json string
func CreatePolicy(tenantId, data string) (
	*structs.Policy, error) {
	start := time.Now()
	var createdAt pgtype.Timestamptz

	p := structs.Policy{TenantId: tenantId, Data: data}
	ctx, cancelFunc := context.WithTimeout(context.Background(), dbTimeout)
	defer cancelFunc()
	err := gDbPool.QueryRow(ctx, `INSERT INTO policy(tenant_id, data, enabled)
		VALUES($1,$2,true) RETURNING id, enabled, created_at`,
		p.TenantId, p.Data).Scan(&p.Id, &p.Enabled, &createdAt)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return nil, err
	}
	if createdAt.Valid {
		p.CreatedAt = createdAt.Time
	}
	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbCreatePolicy)

	go cache.CreatePolicy(&p)

	return &p, nil
}

// Get policy by id
func GetPolicy(id uuid.UUID, tenantId string) (*structs.Policy, error) {
	var err error
	p := &structs.Policy{}
	start := time.Now()
	var createdAt, updatedAt pgtype.Timestamptz

	ctx, cancelFunc := context.WithTimeout(gCtx, dbTimeout)
	defer cancelFunc()
	err = gDbPool.QueryRow(ctx,
		`SELECT id, tenant_id, data, enabled, created_at, updated_at
		FROM policy WHERE id=$1 AND tenant_id=$2`,
		id, tenantId).Scan(
		&p.Id, &p.TenantId, &p.Data, &p.Enabled, &createdAt, &updatedAt)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return nil, err
	}
	if createdAt.Valid {
		p.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		p.UpdatedAt = updatedAt.Time
	}
	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbGetPolicy)
	return p, nil
}

// update policy
func UpdatePolicy(p *structs.Policy) error {
	var id uuid.UUID
	start := time.Now()
	ctx, cancelFunc := context.WithTimeout(context.Background(), dbTimeout)
	defer cancelFunc()
	err := gDbPool.QueryRow(ctx,
		`UPDATE policy SET
		updated_at=now(), data=$1 WHERE id=$2 AND tenant_id=$3
		RETURNING id`,
		p.Data, p.Id, p.TenantId).Scan(&id)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return err
	}
	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbUpdatePolicy)

	go cache.UpdatePolicy(p)

	return nil
}

// delete policy
func DeletePolicy(id uuid.UUID, tenantId string) error {
	start := time.Now()

	ctx, cancelFunc := context.WithTimeout(context.Background(), dbTimeout)
	defer cancelFunc()

	// optimistic cache purge
	cache.DeletePolicy(id)

	tx, err := gDbPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx, ctx)

	res, err := tx.Exec(ctx,
		`DELETE FROM policy WHERE id=$1 AND tenant_id=$2`, id, tenantId)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNoRows
	}

	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbDeleteEnroll)

	esLogger.Debug("Deleted policy",
		zap.String("id", id.String()),
		zap.String("tenantId", tenantId))

	commit(tx, ctx)
	return nil
}

// get policy by tenant_id
func GetPolicyId(tenantId string) (*uuid.UUID, error) {
	start := time.Now()

	ctx, cancelFunc := context.WithTimeout(context.Background(), dbTimeout)
	defer cancelFunc()

	var id uuid.UUID
	err := gDbPool.QueryRow(ctx,
		`SELECT id FROM policy WHERE tenant_id=$1`,
		tenantId).Scan(&id)
	if err != nil {
		return nil, err
	}

	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbGetPolicyByTenant)

	return &id, nil
}
