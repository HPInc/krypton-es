// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"net/http"
	"time"

	"github.com/HPInc/krypton-es/es/service/db"
	"go.uber.org/zap"
)

/*
/api/v1/policy
policy info for tenant

Returns:
- 200

Errors:
- 400
  - X-HP-TokenType header must be present and set to one of the user token types

- 401
  - Could not verify token
  - Token expired or not yet valid

- 405
  - Must be HEAD

- 500
  - should not be here. yet, here we are.
*/
func GetPolicyInfo(w http.ResponseWriter, r *http.Request) *enrollError {
	startTime := time.Now()

	// parse bearer token
	ei, err := GetEnrollInfoFromToken(r)
	if err != nil {
		if IsTokenTypeHeaderError(err) {
			return &enrollError{err, http.StatusBadRequest}
		}
		return &enrollError{err, http.StatusUnauthorized}
	}
	id, err := db.GetPolicyId(ei.TenantId)
	if err != nil {
		esLogger.Error("No policy found for tenant",
			zap.String("TenantId", ei.TenantId))
		return &enrollError{ErrGetPolicy, getHttpCodeForDbError(err)}
	}

	// write id header for info
	w.Header().Set(headerPolicyId, id.String())

	esLogger.Info(
		"PolicyInfo",
		zap.String("TenantID", ei.TenantId),
		zap.String("Elapsed", time.Since(startTime).String()))
	return nil
}
