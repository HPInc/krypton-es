// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	keyUnenrollElapsed        = "unenroll_elapsed"
	keyUnenrollElapsedCount   = "unenroll_elapsed_count"
	keyUnenrollElapsedAverage = "unenroll_elapsed_average"
)

// lua script to do an atomic average operation
// of an incoming enroll elapsed time
var unenrollAverage = redis.NewScript(`
	local key = KEYS[1]
	local value = redis.call("GET", key)
	if not value then
		value = 0
	end
	value = value + ARGV[1]
	redis.call("SET", key, value)

	local countkey = string.format("%s_%s", key, "count")
	local count = redis.call("INCR", countkey)

	local avgkey = string.format("%s_%s", key, "average")
	local avg = math.ceil(value/count)
	redis.call("SET", avgkey, avg)
	return avg
`)

// Take the incoming elapsed time and do the following
// 1. add to a sum key
// 2. increment a count key
// 3. store sum/count to an average key
func AddUnenrollElapsed(elapsed float64) {
	if !isEnabled {
		return
	}
	avg, err := unenrollAverage.Run(gCtx, cacheClient,
		[]string{keyUnenrollElapsed}, elapsed).Int()
	if err != nil {
		esLogger.Error("Could not update average unenroll time",
			zap.Error(err))
		return
	}
	esLogger.Info("Average elapsed time for unenroll",
		zap.Int("Average", avg))
}

// Compute the average by doing a sum/count
// time is in seconds.
func GetAverageUnenrollTime() (int, error) {
	if !isEnabled {
		return 0, ErrCacheNotFound
	}

	// Get the requested device object from the cache.
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	unenrollAverage, err := cacheClient.Get(ctx,
		keyUnenrollElapsedAverage).Int()
	if err != nil {
		esLogger.Error("Could not get average unenroll time",
			zap.Error(err))
		return 0, err
	}
	return unenrollAverage, nil
}
