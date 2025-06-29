// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"net/http"
	"time"

	"github.com/HPInc/krypton-es/es/service/db"
	"github.com/HPInc/krypton-es/es/service/structs"
	"go.uber.org/zap"
)

/*
/api/v1/policy
update policy for tenant

Returns:
- 200

Errors:
- 400
  - X-HP-TokenType header must be present and set to one of the user token types

- 401

  - Could not verify token

  - Token expired or not yet valid

  - 404
    There is no such policy

- 405
  - Must be PATCH

- 500
  - should not be here. yet, here we are.
*/
func UpdatePolicy(w http.ResponseWriter, r *http.Request) *enrollError {
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

	data, err := validatePolicy(r)
	if err != nil {
		esLogger.Error("Policy payload is invalid",
			zap.String("Request ID:", requestID),
			zap.String("PolicyId", policyId.String()),
			zap.String("TenantId", ei.TenantId))
		return &enrollError{err, http.StatusBadRequest}
	}

	p := &structs.Policy{
		Id:       policyId,
		TenantId: ei.TenantId,
		Data:     string(data),
	}

	err = db.UpdatePolicy(p)
	if err != nil {
		esLogger.Error("Policy update failed",
			zap.String("Request ID:", requestID),
			zap.String("PolicyId", policyId.String()),
			zap.String("TenantId", ei.TenantId))
		return &enrollError{ErrUpdatePolicy, getHttpCodeForDbError(err)}
	}

	esLogger.Info("UpdatePolicy",
		zap.String("Request ID:", requestID),
		zap.String("PolicyId", policyId.String()),
		zap.String("TenantID", ei.TenantId),
		zap.String("Elapsed", time.Since(startTime).String()))
	return nil
}
