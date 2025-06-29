// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package tokenmgr

import (
	dstsclient "github.com/HPInc/krypton-es/es/service/client/dsts"
)

type DstsEnrollmentTokenValidator struct {
	settings *TokenIssuerSettings
}

func newDstsEnrollmentTokenValidator(tokenSettings *TokenIssuerSettings) *DstsEnrollmentTokenValidator {
	return &DstsEnrollmentTokenValidator{settings: tokenSettings}
}

func (v DstsEnrollmentTokenValidator) ValidateToken(tokenString string) (*EnrollClaims, error) {
	// enrollment token is issued by DSTS. delegate to dsts for validation
	tid, err := dstsclient.ValidateEnrollmentToken(tokenString)
	if err != nil {
		return nil, err.Error
	}

	return &EnrollClaims{TenantId: tid}, nil
}
