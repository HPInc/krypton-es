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
	cacheFunctionEnrollStatus = "EnrollStatus"
)

// create enroll status
func CreateEnrollStatus(id uuid.UUID, tenantId, userId string, deviceId uuid.UUID, status int) {
	if !isEnabled {
		return
	}
	entry := structs.EnrollStatus{
		Status:   status,
		TenantId: tenantId,
		UserId:   userId,
		DeviceId: deviceId,
	}
	setEnrollStatus(id, &entry)
}

// set status for enroll id
func SetEnrollStatus(id, deviceId uuid.UUID, status int) {
	if !isEnabled {
		return
	}
	// fetch entry
	entry, err := getEnrollStatus(id)
	if err != nil {
		esLogger.Error("Could not update enroll status",
			zap.String("id", id.String()),
			zap.Error(err))
		metrics.ReportCacheError(operationCacheSet, cacheFunctionEnrollStatus)
		return
	}
	entry.DeviceId = deviceId
	entry.Status = status
	setEnrollStatus(id, entry)
}

// get status for enroll id
func GetEnrollStatus(id uuid.UUID) (*structs.EnrollStatus, error) {
	if !isEnabled {
		return nil, ErrCacheNotFound
	}

	defer metrics.ReportLatencyMetric(metrics.MetricCacheLatency,
		time.Now(), operationCacheGet)
	// Get status from cache by id.
	entry, err := getEnrollStatus(id)
	if err != nil {
		esLogger.Debug("Could not get enroll status",
			zap.String("id", id.String()),
			zap.Error(err))
		metrics.ReportCacheError(operationCacheGet, cacheFunctionEnrollStatus)
		return nil, err
	}
	metrics.ReportCacheHit(cacheFunctionEnrollStatus)
	return entry, nil
}

// delete enroll id from status cache
func DeleteEnrollStatusById(id uuid.UUID) {
	if !isEnabled {
		return
	}

	defer metrics.ReportLatencyMetric(metrics.MetricCacheLatency,
		time.Now(), operationCacheDelete)
	// Get status from cache by id.
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	err := cacheClient.Del(ctx, fmt.Sprintf(prefixEnrollStatus, id)).Err()
	if err != nil {
		esLogger.Error("Could not delete enroll status by id",
			zap.String("id", id.String()),
			zap.Error(err))
		metrics.ReportCacheError(operationCacheGet, cacheFunctionEnrollStatus)
	}
}

// internal get enroll status entry
func getEnrollStatus(id uuid.UUID) (*structs.EnrollStatus, error) {
	var entry structs.EnrollStatus
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	cacheEntry, err := cacheClient.Get(ctx,
		fmt.Sprintf(prefixEnrollStatus, id)).Result()
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal([]byte(cacheEntry), &entry); err != nil {
		esLogger.Error("Failed to unmarshal status from cache!",
			zap.String("id", id.String()),
			zap.Error(err))
		return nil, err
	}
	return &entry, nil
}

// internal set enroll status
func setEnrollStatus(id uuid.UUID, entry *structs.EnrollStatus) {
	defer metrics.ReportLatencyMetric(metrics.MetricCacheLatency,
		time.Now(), operationCacheSet)
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	// Marshal the file object for caching.
	cacheEntry, err := json.Marshal(entry)
	if err != nil {
		esLogger.Error("Failed to marshal status for caching!",
			zap.String("id", id.String()),
			zap.String("tenant_id", entry.TenantId),
			zap.String("user_id", entry.UserId),
			zap.Error(err),
		)
		return
	}
	key := fmt.Sprintf(prefixEnrollStatus, id)
	if err = cacheClient.Set(ctx, key, cacheEntry, ttlStatus).Err(); err != nil {
		esLogger.Error("Could not create enroll status",
			zap.String("id", id.String()),
			zap.String("tenant_id", entry.TenantId),
			zap.String("user_id", entry.UserId),
			zap.Error(err))
		metrics.ReportCacheError(operationCacheSet, cacheFunctionEnrollStatus)
		return
	}
}
