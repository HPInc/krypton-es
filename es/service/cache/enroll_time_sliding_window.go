// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	keyEnrollLastN        = "enroll_last_n"
	keyEnrollLastNAverage = "enroll_last_n:average"
	keyEnrollLastNSum     = "enroll_last_n:sum"
)

// lua script to do an atomic average operation
// of an incoming window of enroll elapsed time
// there is no time based purge of accumulated values
// this is just a fifo of enrolled times and an average
// of the items in the queue at a given time
var enrollLastNAverage = redis.NewScript(`
	local key = KEYS[1]
	local value = ARGV[1]
	local max_count = ARGV[2]
	local popped = 0.0

	redis.call("RPUSH", key, value)
	local count = redis.call("LLEN", key)
	if count >= tonumber(max_count) then
	  popped = redis.call("LPOP", key)
	end

	local sumkey = string.format("%s:%s", key, "sum")
	value = value - popped
	local sum = redis.call("INCRBYFLOAT", sumkey, value)

	local avg = math.ceil(sum/count)
	local avgkey = string.format("%s:%s", key, "average")
	redis.call("SET", avgkey, avg)
	return avg
`)

// Take the incoming elapsed time and do the following
// 1. add to a running sum key
// 2. increment a running count key
// 3. store sum/count to a running average key
func AddLastNEnrollElapsed(elapsed float64) {
	if !isEnabled {
		return
	}
	avg, err := enrollLastNAverage.Run(gCtx, cacheClient,
		[]string{keyEnrollLastN}, elapsed,
		gCacheConfig.EnrollTimeWindowSize).Int()
	if err != nil {
		esLogger.Error("Could not update last n average enroll time",
			zap.Error(err))
		return
	}
	esLogger.Info("Last n average elapsed time for enroll",
		zap.Int("Average", avg))
}

// Compute the average by doing a sum/count
// time is in seconds.
func GetLastNAverageEnrollTime() (int, error) {
	if !isEnabled {
		return 0, ErrCacheNotFound
	}

	// Get the requested device object from the cache.
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	enrollAverage, err := cacheClient.Get(ctx,
		keyEnrollLastNAverage).Int()
	if err != nil {
		esLogger.Error("Could not get last n average enroll time",
			zap.Error(err))
		return 0, err
	}
	return enrollAverage, nil
}
