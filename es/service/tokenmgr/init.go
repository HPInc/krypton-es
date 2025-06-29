// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package tokenmgr

import (
	"context"

	"go.uber.org/zap"
)

var (
	// Structured logging using Uber Zap.
	esLogger *zap.Logger

	// stop channel for keys refresh
	refreshKeysStopChannel chan bool

	// base context for package
	gCtx context.Context
)

func Init(logger *zap.Logger, tokenConfigFile string) error {
	esLogger = logger
	gCtx = context.Background()

	if !loadTokenConfiguration(tokenConfigFile) {
		return ErrTokenConfigurationInitFailure
	}

	refreshKeysStopChannel = make(chan bool)
	return startJwksRefresher()
}

func Shutdown() {
	esLogger.Info("HP Enrollment service: signalling shutdown to JWKs refresher")
	refreshKeysStopChannel <- true
	gCtx.Done()
}
