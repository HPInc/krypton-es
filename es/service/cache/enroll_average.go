// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	keyEnrollElapsed        = "enroll_elapsed"
	keyEnrollElapsedCount   = "enroll_elapsed_count"
	keyEnrollElapsedAverage = "enroll_elapsed_average"
)

// lua script to do an atomic average operation
// of an incoming enroll elapsed time
var enrollAverage = redis.NewScript(`
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
func addEnrollElapsed(elapsed float64) {
	if !isEnabled {
		return
	}
	avg, err := enrollAverage.Run(gCtx, cacheClient,
		[]string{keyEnrollElapsed}, elapsed).Int()
	if err != nil {
		esLogger.Error("Could not update average enroll time",
			zap.Error(err))
		return
	}
	esLogger.Info("Average elapsed time for enroll",
		zap.Int("Average", avg))
}

// Compute the average by doing a sum/count
// time is in seconds.
func getAverageEnrollElapsed() (int, error) {
	if !isEnabled {
		return 0, ErrCacheNotFound
	}

	// Get the requested device object from the cache.
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	enrollAverage, err := cacheClient.Get(ctx,
		keyEnrollElapsedAverage).Int()
	if err != nil {
		esLogger.Error("Could not get average enroll time",
			zap.Error(err))
		return 0, err
	}
	return enrollAverage, nil
}
