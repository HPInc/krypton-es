// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cache

import (
	"math"
	"testing"

	"github.com/redis/go-redis/v9"
)

// lua script to do an atomic clear of keys
// used in enroll average
var enrollAverageClear = redis.NewScript(`
	local key = KEYS[1]
	local value = ARGV[1]
	redis.call("SET", key, value)
	local countkey = string.format("%s_%s", key, "count")
	redis.call("SET", countkey, value)
	local avgkey = string.format("%s_%s", key, "average")
	redis.call("SET", avgkey, value)
	return 0
`)

// setup enroll average cache and check result
func TestCacheAverageEnrollTime(t *testing.T) {
	initCacheAverageEnrollKeys(t)

	testEnrollTimes := []float64{1.0, 1.25, 1.50, 1.75}
	expectedAverage := func() int {
		var total float64
		for _, val := range testEnrollTimes {
			addEnrollElapsed(val)
			total += val
		}
		return int(math.Ceil(total / float64(len(testEnrollTimes))))
	}()
	cacheAverage, err := getAverageEnrollElapsed()
	if err != nil {
		t.Errorf("Cache get average enroll time failed. Expected no error. Got %v", err)
	}
	if cacheAverage != expectedAverage {
		t.Errorf("Cache get average enroll time. Expected %d. Got %d", cacheAverage, expectedAverage)
	}
}

func initCacheAverageEnrollKeys(t *testing.T) {
	// clear cache keys
	_, err := enrollAverageClear.Run(gCtx, cacheClient,
		[]string{keyEnrollElapsed}, 0).Int()
	if err != nil {
		t.Errorf("Failed resetting enroll average keys. %v", err)
	}
	startAverage, err := getAverageEnrollElapsed()
	if err != nil {
		t.Errorf("Cache get average enroll time failed. Expected no error. Got %v", err)
	}
	if startAverage != 0 {
		t.Errorf("Cache get average enroll time not properly initialized. Expected 0. Got %d", startAverage)
	}
}
