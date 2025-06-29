// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"context"
	"time"

	"github.com/HPInc/krypton-es/es/service/metrics"
	"go.uber.org/zap"
)

// check if key exists
func HasKey(kid string) (bool, error) {
	count := 0
	ctx, cancelFunc := context.WithTimeout(context.Background(), dbTimeout)
	defer cancelFunc()
	err := gDbPool.QueryRow(ctx, "SELECT count(*) FROM public_key WHERE kid=$1", kid).Scan(&count)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
	}
	return count == 1, nil
}

// If we have a public key saved, look it up and return
func GetPublicKey(kid string) (string, error) {
	var key string
	var err error
	start := time.Now()
	ctx, cancelFunc := context.WithTimeout(context.Background(), dbTimeout)
	defer cancelFunc()
	err = gDbPool.QueryRow(ctx, "SELECT public_key FROM public_key WHERE kid=$1", kid).Scan(&key)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return "", err
	}
	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency,
		start, operationDbGetPublicKey)
	return key, nil
}

func addPublicKey(kid string, alg string, key string) error {
	start := time.Now()
	ctx, cancelFunc := context.WithTimeout(context.Background(), dbTimeout)
	defer cancelFunc()
	_, err := gDbPool.Exec(ctx, "INSERT INTO public_key (kid, alg, public_key) VALUES($1, $2, $3)",
		kid, alg, key)
	if err != nil {
		esLogger.Error("DB: SQL Error", zap.Error(err))
		return err
	}
	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency,
		start, operationDbSetPublicKey)
	return nil
}

// strip begin and end markers and add to db
func AddKey(kid string, alg string, key string) error {
	var err error
	var ok bool
	if ok, err = HasKey(kid); err != nil {
		return err
	}
	if ok {
		return nil
	}
	return addPublicKey(kid, alg, key)
}
