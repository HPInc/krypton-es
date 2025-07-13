// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	dstsclient "github.com/HPInc/krypton-es/es/service/client/dsts"
	"github.com/HPInc/krypton-es/es/service/tokenmgr"
	"go.uber.org/zap"
)

/*
/enroll_token/{tenant_id}
Get enroll token for tenant

Returns:
- 200
  - {"access_token": <jwt token>}

Errors:
- 400
  - Malformed or missing Authorization header
  - Malformed or missing payload

- 401
  - Could not verify token
  - Token expired or not yet valid

- 405
  - Must be GET

- 500
  - should not be here. yet, here we are.
*/
func GetEnrollToken(w http.ResponseWriter, r *http.Request) *enrollError {
	startTime := time.Now()

	tenantId, enroll_error := getUUIDParam(r, paramTenantID)
	if enroll_error != nil {
		return enroll_error
	}
	// check if X-HP-TokenType header is provided and set to "app"
	if err := validateAppToken(r, tokenmgr.TokenTypeApp); err != nil {
		return &enrollError{err, http.StatusBadRequest}
	}
	// parse bearer token
	ei, err := GetEnrollInfoFromToken(r)
	if err != nil {
		return &enrollError{err, http.StatusUnauthorized}
	}
	token, clientErr := dstsclient.GetEnrollmentToken(tenantId.String())
	if clientErr != nil {
		esLogger.Error("GetEnrollToken: error from dsts",
			zap.Error(clientErr.Error))
		return &enrollError{ErrGetEnrollToken, clientErr.HttpCode}
	}

	jsonString, err := json.Marshal(token)
	if err != nil {
		return &enrollError{ErrInternal, http.StatusInternalServerError}
	}
	esLogger.Info(
		"GetEnrollToken",
		zap.String("TenantID", ei.TenantId),
		zap.String("Elapsed", time.Since(startTime).String()))
	fmt.Fprintf(w, "%s", string(jsonString))
	return nil
}

// bearer token should be an app token obtained from dsts
func validateAppToken(r *http.Request, tt tokenmgr.TokenType) error {
	// Retrieve the token type specified in the request.
	tokenType := r.Header.Get(headerTokenType)
	if tokenType == "" {
		esLogger.Error(ErrTokenTypeHeaderNotFound.Error())
		return ErrTokenTypeHeaderNotFound
	}
	if tokenmgr.TokenType(tokenType) != tt {
		return ErrInvalidTokenType
	}
	return nil
}
