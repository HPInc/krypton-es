// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package tokenmgr

import (
	"fmt"
	"strings"
)

const (
	TokenTypeAzureAD    TokenType = "azuread"
	TokenTypeDevice     TokenType = "device"
	TokenTypeEnrollment TokenType = "enrollment"
	TokenTypeTest       TokenType = "test"
	TokenTypeApp        TokenType = "app"
)

type EnrollClaims struct {
	TenantId string `json:"tid"`
	// valid for unenroll and renew_enroll
	DeviceId string `json:"deviceid"`
	// user id valid for user tokens used in enroll
	UserId string `json:"userid"`
}

type TokenValidator interface {
	ValidateToken(token string) (*EnrollClaims, error)
}

func ValidateToken(tokenType string, tokenString string) (*EnrollClaims, error) {
	var validator TokenValidator

	tokenType = strings.ToLower(tokenType)
	tokenSettings, ok := tokenConfig.TokenTypes[TokenType(tokenType)]
	if !ok {
		return nil, ErrUnsupportedTokenType
	}

	switch TokenType(tokenSettings.Type) {
	case TokenTypeAzureAD:
		validator = newAzureADTokenValidator(&tokenSettings)

	case TokenTypeTest:
		validator = newTestTokenValidator(&tokenSettings)

	// no separate device validation yet. check with dsts and update
	case TokenTypeEnrollment:
		validator = newDstsEnrollmentTokenValidator(&tokenSettings)

	case TokenTypeDevice:
		validator = newDstsDeviceTokenValidator(&tokenSettings)

	case TokenTypeApp:
		validator = newAppTokenValidator(&tokenSettings)

	default:
		return nil, fmt.Errorf("invalid token type: %s",
			tokenSettings.Type)
	}

	return validator.ValidateToken(tokenString)
}

func IsAppToken(tokenType string) bool {
	return tokenType == string(TokenTypeApp)
}
