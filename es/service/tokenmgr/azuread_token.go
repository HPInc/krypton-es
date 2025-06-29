// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package tokenmgr

import (
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type AzureADTokenValidator struct {
	settings *TokenIssuerSettings
}

type AzureADTokenClaims struct {
	TenantId string `json:"tid"`

	UserId string `json:"user_id"`

	// audience will work per rfc with registeredclaims
	jwt.RegisteredClaims
}

func newAzureADTokenValidator(tokenSettings *TokenIssuerSettings) *AzureADTokenValidator {
	return &AzureADTokenValidator{settings: tokenSettings}
}

func (v AzureADTokenValidator) ValidateToken(tokenString string) (*EnrollClaims, error) {
	var claims AzureADTokenClaims

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
		!token.Claims.(*AzureADTokenClaims).VerifyAudience(v.settings.Audience, true) {
		return nil, ErrInvalidAudienceClaim
	}

	if !strings.HasPrefix(claims.Issuer, v.settings.Issuer) {
		return nil, ErrInvalidIssuerClaim
	}

	return &EnrollClaims{TenantId: claims.TenantId, UserId: claims.Subject}, nil
}
