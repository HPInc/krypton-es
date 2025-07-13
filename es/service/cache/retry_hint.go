// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cache

const (
	retryAfterHintStrategyAverageEnrollTime = "average_enroll_time"
	retryAfterHintStrategyQueueBacklog      = "queue_backlog"
	retryAfterHintStrategySlidingWindow     = "sliding_window"
)

// Compute the average using the current
// cache strategy
func GetAverageEnrollTime() (int, error) {
	if !isEnabled {
		return 0, ErrCacheNotFound
	}
	switch gCacheConfig.RetryAfterHintStrategy {
	case retryAfterHintStrategyAverageEnrollTime:
		return getAverageEnrollElapsed()
	default:
		return getAverageEnrollElapsed()
	}
}

func AddEnrollElapsed(elapsed float64) {
	if !isEnabled {
		return
	}
	switch gCacheConfig.RetryAfterHintStrategy {
	case retryAfterHintStrategyAverageEnrollTime:
		addEnrollElapsed(elapsed)
	default:
		addEnrollElapsed(elapsed)
	}
}
