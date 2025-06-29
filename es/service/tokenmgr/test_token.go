// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package tokenmgr

import (
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type TestTokenValidator struct {
	settings *TokenIssuerSettings
}

type TestTokenClaims struct {
	TenantId string `json:"tid"`

	// audience will work per rfc with registeredclaims
	jwt.RegisteredClaims
}

func newTestTokenValidator(tokenSettings *TokenIssuerSettings) *TestTokenValidator {
	return &TestTokenValidator{settings: tokenSettings}
}

func (v TestTokenValidator) ValidateToken(tokenString string) (*EnrollClaims, error) {
	var claims TestTokenClaims

	token, err := jwt.ParseWithClaims(tokenString, &claims,
		getPublicKeyForJwt)
	if err != nil {
		return nil, err
	} else if !token.Valid {
		return nil, ErrInvalidToken
	}

	if token.Header["alg"] == nil {
		return nil, ErrInvalidTokenHeaderSigningAlg
	}

	if v.settings.Audience != "" &&
		!token.Claims.(*TestTokenClaims).VerifyAudience(v.settings.Audience, true) {
		return nil, ErrInvalidAudienceClaim
	}

	if !strings.HasPrefix(claims.Issuer, v.settings.Issuer) {
		return nil, ErrInvalidIssuerClaim
	}

	return &EnrollClaims{TenantId: claims.TenantId}, nil
}
