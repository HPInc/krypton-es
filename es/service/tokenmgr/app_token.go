// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package tokenmgr

import (
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type AppTokenValidator struct {
	settings *TokenIssuerSettings
}

const (
	// type values for "typ" claim in token
	appType = "app"
)

// holder to get extended claims that we expect
// device tokens and app tokens have differing claims
// but we take a union approach to keep it simple
type AppTokenClaims struct {
	Type string `json:"typ"`
	jwt.RegisteredClaims
}

func newAppTokenValidator(tokenSettings *TokenIssuerSettings) *AppTokenValidator {
	return &AppTokenValidator{settings: tokenSettings}
}

// get bearer token and do common validation
func (v AppTokenValidator) ValidateToken(tokenStr string) (*EnrollClaims, error) {
	var claims AppTokenClaims

	token, err := jwt.ParseWithClaims(tokenStr, &claims, getPublicKeyForJwt)
	if err != nil {
		return nil, err
	} else if !token.Valid {
		return nil, ErrInvalidToken
	}

	if token.Header["alg"] == nil {
		return nil, ErrInvalidTokenHeaderSigningAlg
	}
	if claims.Type != appType {
		return nil, ErrInvalidTypeClaim
	}
	if !strings.HasPrefix(claims.Issuer, v.settings.Issuer) {
		return nil, ErrInvalidIssuerClaim
	}
	if !v.hasAppIdSubject(claims.Subject) {
		return nil, ErrInvalidSubjectClaim
	}
	return &EnrollClaims{}, nil
}

// app token subjects should match one of the registered apps
func (v AppTokenValidator) hasAppIdSubject(subject string) bool {
	for _, id := range v.settings.AllowedAppIds {
		if id == subject {
			return true
		}
	}
	return false
}
