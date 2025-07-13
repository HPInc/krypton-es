// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/HPInc/krypton-es/es/service/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	// Structured logging using Uber Zap.
	esLogger *zap.Logger

	cacheClient  *redis.Client
	isEnabled    bool
	gCtx         context.Context
	gCacheConfig *config.Cache

	ErrCacheNotFound = errors.New("item not found in cache")
)

const (
	// Cache connection string.
	cacheConnStr = "%s:%d"

	// Timeout for requests to the Redis cache.
	cacheTimeout = (time.Second * 1)
	dialTimeout  = (time.Second * 5)
	readTimeout  = (time.Second * 3)
	writeTimeout = (time.Second * 3)
	poolSize     = 10
	poolTimeout  = (time.Second * 4)

	prefixEnrollStatus   = "status:%s"
	prefixUnenrollStatus = "unenroll_status:%s"
	prefixCsrHash        = "csrhash:%s"
	prefixPolicy         = "policy:%s"

	// ttl
	ttlStatus  = (time.Minute * 5)
	ttlCsrHash = (time.Minute * 10)

	// Caching operation names.
	operationCacheGet    = "get"
	operationCacheSet    = "set"
	operationCacheDelete = "delete"
)

// Init - initialize a connection to the Redis based enroll cache.
func Init(logger *zap.Logger, config *config.Cache) error {
	esLogger = logger
	gCacheConfig = config
	isEnabled = gCacheConfig.Enabled

	if !isEnabled {
		esLogger.Info("Caching is disabled - nothing to initialize!")
		return nil
	}

	// Initialize the cache client with appropriate connection options.
	cacheClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf(cacheConnStr, gCacheConfig.Server,
			gCacheConfig.Port),
		Password:     gCacheConfig.Password,
		DB:           gCacheConfig.CacheDatabase,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		PoolSize:     poolSize,
		PoolTimeout:  poolTimeout,
	})

	// Attempt to connect to the enroll cache.
	gCtx = context.Background()
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	_, err := cacheClient.Ping(ctx).Result()
	if err != nil {
		esLogger.Error("Failed to connect to the enroll cache!",
			zap.String("Cache address: ", cacheClient.Options().Addr),
			zap.Error(err),
		)
		return err
	}

	esLogger.Info("Successfully initialized the enroll cache!",
		zap.String("Cache address: ", cacheClient.Options().Addr),
	)
	return nil
}

// Shutdown the enroll cache and cleanup Redis connections.
func Shutdown() {
	if !isEnabled {
		esLogger.Info("Enroll cache was not initialized - skipping shutdown!")
		return
	}

	gCtx.Done()
	isEnabled = false

	// Close the client connection to the cache.
	err := cacheClient.Close()
	if err != nil {
		esLogger.Error("Failed to shutdown connection to the enroll cache!",
			zap.Error(err),
		)
		return
	}

	esLogger.Info("Successfully shutdown connection to the enroll cache!")
}
