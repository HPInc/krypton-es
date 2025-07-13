// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/HPInc/krypton-es/es/service/metrics"
	"go.uber.org/zap"
)

const (
	cacheFunctionCsrHash = "CsrHash"
)

// set csr
func SetCsrHash(csrhash string) {
	if !isEnabled {
		return
	}
	defer metrics.ReportLatencyMetric(metrics.MetricCacheLatency,
		time.Now(), operationCacheSet)
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	key := fmt.Sprintf(prefixCsrHash, csrhash)
	err := cacheClient.Set(ctx, key, true, ttlCsrHash).Err()
	if err != nil {
		esLogger.Error("Could not csrhash entry",
			zap.String("csrhash", csrhash),
			zap.Error(err))
		metrics.ReportCacheError(operationCacheSet, cacheFunctionCsrHash)
		return
	}
}

// lookup by csrhash
func HasCsrHash(csrhash string) (bool, error) {
	if !isEnabled {
		return false, ErrCacheNotFound
	}

	defer metrics.ReportLatencyMetric(metrics.MetricCacheLatency,
		time.Now(), operationCacheGet)
	// check if entry exists by key.
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	key := fmt.Sprintf(prefixCsrHash, csrhash)
	status, err := cacheClient.Get(ctx, key).Bool()
	if err != nil {
		esLogger.Debug("Could not lookup csrhash",
			zap.String("csrhash", csrhash),
			zap.Error(err))
		metrics.ReportCacheError(operationCacheGet, cacheFunctionCsrHash)
		return false, err
	}
	metrics.ReportCacheHit(cacheFunctionCsrHash)
	return status, nil
}
