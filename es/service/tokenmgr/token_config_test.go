// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package tokenmgr

import (
	"testing"

	"go.uber.org/zap"
)

func TestLoadConfiguration(t *testing.T) {
	esLogger, _ = zap.NewProduction(zap.AddCaller())
	defer esLogger.Sync()
	loadTokenConfiguration("../config/token_config.yaml")
}
