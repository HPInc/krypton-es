// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HPInc/krypton-es/es/service/metrics"
	"github.com/HPInc/krypton-es/es/service/structs"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	cacheFunctionPolicy = "Policy"
)

// create policy
func CreatePolicy(p *structs.Policy) {
	if !isEnabled {
		return
	}
	setPolicy(p.Id, p)
}

// get policy
func GetPolicy(id uuid.UUID) (*structs.Policy, error) {
	return getPolicy(id)
}

// set policy
func UpdatePolicy(p *structs.Policy) {
	if !isEnabled {
		return
	}
	setPolicy(p.Id, p)
}

// delete policy from cache
func DeletePolicy(id uuid.UUID) {
	if !isEnabled {
		return
	}

	defer metrics.ReportLatencyMetric(metrics.MetricCacheLatency,
		time.Now(), operationCacheDelete)
	// Get status from cache by id.
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	err := cacheClient.Del(ctx, fmt.Sprintf(prefixPolicy, id)).Err()
	if err != nil {
		esLogger.Error("Could not delete policy",
			zap.String("id", id.String()),
			zap.Error(err))
		metrics.ReportCacheError(operationCacheDelete, cacheFunctionPolicy)
	}
}

// internal get policy
func getPolicy(id uuid.UUID) (*structs.Policy, error) {
	var p structs.Policy
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	cacheEntry, err := cacheClient.Get(ctx,
		fmt.Sprintf(prefixPolicy, id)).Result()
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal([]byte(cacheEntry), &p); err != nil {
		esLogger.Error("Failed to unmarshal policy from cache!",
			zap.String("id", id.String()),
			zap.Error(err))
		return nil, err
	}
	return &p, nil
}

// internal set policy
func setPolicy(id uuid.UUID, p *structs.Policy) {
	defer metrics.ReportLatencyMetric(metrics.MetricCacheLatency,
		time.Now(), operationCacheSet)
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	// Marshal the file object for caching.
	cacheEntry, err := json.Marshal(p)
	if err != nil {
		esLogger.Error("Failed to marshal status for caching!",
			zap.String("id", id.String()),
			zap.String("tenant_id", p.TenantId),
			zap.Error(err),
		)
		return
	}
	key := fmt.Sprintf(prefixPolicy, id)
	if err = cacheClient.Set(ctx, key, cacheEntry, ttlStatus).Err(); err != nil {
		esLogger.Error("Could not create policy",
			zap.String("id", id.String()),
			zap.String("tenant_id", p.TenantId),
			zap.Error(err))
		metrics.ReportCacheError(operationCacheSet, cacheFunctionPolicy)
		return
	}
}
