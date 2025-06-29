// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package policy

import (
	"encoding/json"

	"go.uber.org/zap"
)

func FromString(str string) (*Policy, error) {
	var p Policy
	if err := json.Unmarshal([]byte(str), &p); err != nil {
		esLogger.Error("Failed to unmarshal policy", zap.Error(err))
		return nil, err
	}
	return &p, nil
}
