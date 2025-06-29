// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package tokenmgr

import (
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type DstsDeviceTokenValidator struct {
	settings *TokenIssuerSettings
}

type DeviceTokenClaims struct {
	// tenant id from enroll token propagated back from dsts
	TenantId string `json:"tid"`
	// audience will work per rfc with registeredclaims
	jwt.RegisteredClaims
}

func newDstsDeviceTokenValidator(tokenSettings *TokenIssuerSettings) *DstsDeviceTokenValidator {
	return &DstsDeviceTokenValidator{settings: tokenSettings}
}

func (v DstsDeviceTokenValidator) ValidateToken(tokenString string) (*EnrollClaims, error) {
	var claims DeviceTokenClaims

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
		!token.Claims.(*DeviceTokenClaims).VerifyAudience(v.settings.Audience, true) {
		return nil, ErrInvalidAudienceClaim
	}

	if !strings.HasPrefix(claims.Issuer, v.settings.Issuer) {
		return nil, ErrInvalidIssuerClaim
	}

	if claims.Subject == "" {
		esLogger.Error("Could not get the device ID from device token")
		return nil, ErrInvalidToken
	}

	return &EnrollClaims{
		TenantId: claims.TenantId,
		DeviceId: claims.Subject,
	}, nil
}
