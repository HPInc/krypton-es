// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package policy

import (
	"encoding/json"

	"go.uber.org/zap"
)

func ValidateBytes(bytes []byte) bool {
	var p Policy
	if err := json.Unmarshal(bytes, &p); err != nil {
		esLogger.Error("Failed to unmarshal policy", zap.Error(err))
		return false
	}
	return p.Validate() == nil
}

// validate policy data
func (p *Policy) Validate() error {
	if p.Version != defaultPolicyVersion {
		esLogger.Error("Policy version is not valid",
			zap.Int("version", p.Version))
		return ErrInvalidPolicy
	}
	return nil
}
