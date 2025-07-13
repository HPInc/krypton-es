// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package policy

import (
	"strconv"
)

// look up attribute value, convert to int32 and return
func (p *Policy) GetAttributeInt(a PolicyAttribute) (int, error) {
	if val, ok := p.Attributes[a]; ok {
		i, err := strconv.Atoi(val)
		if err != nil {
			return 0, err
		}
		return i, nil
	}
	return 0, ErrPolicyUnknownAttribute
}
