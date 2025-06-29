// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"net/http"
	"strings"

	"github.com/HPInc/krypton-es/es/service/tokenmgr"
)

type EnrollInfo struct {
	UserId   string `json:"user_id"`
	TenantId string `json:"tenant_id"`
	DeviceId string `json:"device_id"`
}

func GetEnrollInfoFromToken(r *http.Request) (*EnrollInfo, error) {
	// Retrieve the token type specified in the request.
	tokenType := r.Header.Get(headerTokenType)
	if tokenType == "" {
		esLogger.Error(ErrTokenTypeHeaderNotFound.Error())
		return nil, ErrTokenTypeHeaderNotFound
	}

	token, err := extractTokenFromAuthorizationHeader(r)
	if err != nil {
		return nil, err
	}

	// Invoke the token manager to validate the access token.
	// specific claims like deviceid are validated by the
	// corresponding validators
	claims, err := tokenmgr.ValidateToken(tokenType, token)
	if err != nil {
		return nil, err
	}

	// tenantid is required for all types so its validated here
	if !tokenmgr.IsAppToken(tokenType) && claims.TenantId == "" {
		esLogger.Error("Could not get tenant id claim from token")
		return nil, tokenmgr.ErrInvalidToken
	}

	return &EnrollInfo{
		DeviceId: claims.DeviceId,
		TenantId: claims.TenantId,
		UserId:   claims.UserId,
	}, nil
}

func extractTokenFromAuthorizationHeader(r *http.Request) (string, error) {
	tokenString := r.Header.Get(headerAuthorization)
	if tokenString == "" {
		esLogger.Error(ErrNoAuthorizationHeader.Error())
		return "", ErrNoAuthorizationHeader
	}
	if !strings.HasPrefix(tokenString, bearerToken) {
		esLogger.Error(ErrNoBearerTokenSpecified.Error())
		return "", ErrNoBearerTokenSpecified
	}
	tokenString = strings.TrimPrefix(tokenString, bearerToken)
	return tokenString, nil
}
