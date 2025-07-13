// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/HPInc/krypton-es/es/service/db"
	"go.uber.org/zap"
)

/*
/api/v1/policy
get policy for tenant

Returns:
- 200

Errors:
- 400
  - X-HP-TokenType header must be present and set to one of the user token types

- 401
  - Could not verify token
  - Token expired or not yet valid

- 405
  - Must be GET

- 500
  - should not be here. yet, here we are.
*/
func GetPolicy(w http.ResponseWriter, r *http.Request) *enrollError {
	startTime := time.Now()
	requestID := r.Header.Get(headerRequestID)

	// Retrieve the specified policy identifier.
	policyId, eErr := getUUIDParam(r, paramPolicyId)
	if eErr != nil {
		return eErr
	}

	// parse bearer token
	ei, err := GetEnrollInfoFromToken(r)
	if err != nil {
		if IsTokenTypeHeaderError(err) {
			return &enrollError{err, http.StatusBadRequest}
		}
		return &enrollError{err, http.StatusUnauthorized}
	}

	p, err := db.GetPolicy(policyId, ei.TenantId)
	if err != nil {
		esLogger.Error("Policy get failed",
			zap.String("Request ID:", requestID),
			zap.String("TenantId", ei.TenantId))
		return &enrollError{err, getHttpCodeForDbError(err)}
	}

	res, err := json.Marshal(p)
	if err != nil {
		return &enrollError{ErrInternal, http.StatusInternalServerError}
	}
	fmt.Fprintf(w, "%s", res)

	esLogger.Info(
		"GetPolicy",
		zap.String("TenantID", ei.TenantId),
		zap.String("Elapsed", time.Since(startTime).String()))
	return nil
}
