// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package policy

import "errors"

var (
	ErrLoadDefaultPolicy      = errors.New("error loading default policy")
	ErrInvalidPolicy          = errors.New("invalid policy")
	ErrPolicyUnknownAttribute = errors.New("unknown policy attribute")
)
