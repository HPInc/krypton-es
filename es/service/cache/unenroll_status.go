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
	cacheFunctionUnenrollStatus = "UnenrollStatus"
)

// create unenroll status
func CreateUnenrollStatus(id uuid.UUID, tenantId, deviceId string, status int) {
	if !isEnabled {
		return
	}
	entry := structs.UnenrollStatus{
		Status:   status,
		TenantId: tenantId,
		DeviceId: deviceId,
	}
	setUnenrollStatus(id, &entry)
}

// set status for unenroll id
func SetUnenrollStatus(id uuid.UUID, status int) {
	if !isEnabled {
		return
	}
	// fetch entry
	entry, err := getUnenrollStatus(id)
	if err != nil {
		esLogger.Error("Could not update unenroll status",
			zap.String("id", id.String()),
			zap.Error(err))
		metrics.ReportCacheError(operationCacheSet, cacheFunctionUnenrollStatus)
		return
	}
	entry.Status = status
	setUnenrollStatus(id, entry)
}

// get status for unenroll id
func GetUnenrollStatus(id uuid.UUID) (*structs.UnenrollStatus, error) {
	if !isEnabled {
		return nil, ErrCacheNotFound
	}

	defer metrics.ReportLatencyMetric(metrics.MetricCacheLatency,
		time.Now(), operationCacheGet)
	// Get status from cache by id.
	entry, err := getUnenrollStatus(id)
	if err != nil {
		esLogger.Debug("Could not get unenroll status",
			zap.String("id", id.String()),
			zap.Error(err))
		metrics.ReportCacheError(operationCacheGet, cacheFunctionUnenrollStatus)
		return nil, err
	}
	metrics.ReportCacheHit(cacheFunctionUnenrollStatus)
	return entry, nil
}

// delete unenroll id from status cache
func DeleteUnenrollStatusById(id uuid.UUID) {
	if !isEnabled {
		return
	}

	defer metrics.ReportLatencyMetric(metrics.MetricCacheLatency,
		time.Now(), operationCacheDelete)
	// Get status from cache by id.
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	err := cacheClient.Del(ctx, fmt.Sprintf(prefixUnenrollStatus, id)).Err()
	if err != nil {
		esLogger.Error("Could not delete unenroll status by id",
			zap.String("id", id.String()),
			zap.Error(err))
		metrics.ReportCacheError(operationCacheGet, cacheFunctionUnenrollStatus)
	}
}

// internal get unenroll status entry
func getUnenrollStatus(id uuid.UUID) (*structs.UnenrollStatus, error) {
	var entry structs.UnenrollStatus
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	cacheEntry, err := cacheClient.Get(ctx,
		fmt.Sprintf(prefixUnenrollStatus, id)).Result()
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

// internal set unenroll status
func setUnenrollStatus(id uuid.UUID, entry *structs.UnenrollStatus) {
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
			zap.String("device_id", entry.DeviceId),
			zap.Error(err),
		)
		return
	}
	key := fmt.Sprintf(prefixUnenrollStatus, id)
	if err = cacheClient.Set(ctx, key, cacheEntry, ttlStatus).Err(); err != nil {
		esLogger.Error("Could not create unenroll status",
			zap.String("id", id.String()),
			zap.String("tenant_id", entry.TenantId),
			zap.String("device_id", entry.DeviceId),
			zap.Error(err))
		metrics.ReportCacheError(operationCacheSet, cacheFunctionUnenrollStatus)
		return
	}
}
