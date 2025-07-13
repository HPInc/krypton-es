// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/HPInc/krypton-es/es/service/db"
	"github.com/HPInc/krypton-es/es/service/policy"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type createPolicyResult struct {
	Id        uuid.UUID `json:"id"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	Elapsed   string    `json:"elapsed"`
	RequestId string    `json:"request_id"`
}

/*
/api/v1/policy
create policy for tenant

Returns:
- 200

Errors:
- 400
  - X-HP-TokenType header must be present and set to one of the user token types

- 401
  - Could not verify token
  - Token expired or not yet valid

- 405
  - Must be POST

- 409
  - There is already a policy created for this tenant. Use update

- 500
  - should not be here. yet, here we are.
*/
func CreatePolicy(w http.ResponseWriter, r *http.Request) *enrollError {
	startTime := time.Now()

	requestID := r.Header.Get(headerRequestID)
	// Check if the contents of the POST were provided using JSON encoding.
	if r.Header.Get(headerContentType) != contentTypeJson {
		esLogger.Error("CreatePolicy request does not have JSON encoding!",
			zap.String("Request ID:", requestID),
		)
		sendUnsupportedMediaTypeResponse(w)
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
			zap.String("TenantId", ei.TenantId))
		return &enrollError{err, http.StatusBadRequest}
	}

	// check if a policy exists
	id, err := db.GetPolicyId(ei.TenantId)
	if err == nil {
		sendPolicyExistsResponse(w, id.String())
		esLogger.Info(
			"CreatePolicy",
			zap.String("Request ID:", requestID),
			zap.String("TenantID", ei.TenantId),
			zap.String("Elapsed", time.Since(startTime).String()))
		return nil
	} else {
		esLogger.Info("Existing policy not found",
			zap.String("Request ID:", requestID),
			zap.String("TenantId", ei.TenantId),
			zap.Error(err))
	}

	p, err := db.CreatePolicy(ei.TenantId, string(data))
	if err != nil {
		esLogger.Error("Policy already exists",
			zap.String("Request ID:", requestID),
			zap.String("TenantId", ei.TenantId))
		return &enrollError{ErrCreatePolicy, getHttpCodeForDbError(err)}
	}

	policyResult := createPolicyResult{
		Id:        p.Id,
		Enabled:   p.Enabled,
		CreatedAt: p.CreatedAt,
		Elapsed:   time.Since(startTime).String(),
	}

	res, err := json.Marshal(policyResult)
	if err != nil {
		return &enrollError{ErrInternal, http.StatusInternalServerError}
	}
	fmt.Fprintf(w, "%s", res)

	esLogger.Info(
		"CreatePolicy",
		zap.String("Request ID:", requestID),
		zap.String("TenantID", ei.TenantId),
		zap.String("Elapsed", time.Since(startTime).String()))
	return nil
}

// validate incoming policy
func validatePolicy(r *http.Request) ([]byte, error) {
	var err error
	if r.ContentLength == 0 {
		return nil, ErrPayloadMissing
	}
	defer r.Body.Close()
	data, err := validatePolicySchema(r.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// validate json schema
func validatePolicySchema(b io.ReadCloser) ([]byte, error) {
	body, err := io.ReadAll(b)
	if err != nil {
		esLogger.Error("Failed to read the request body!",
			zap.Error(err),
		)
		return nil, ErrPayloadRead
	}

	if !policy.ValidateBytes(body) {
		err = ErrInvalidPolicy
		esLogger.Error("Failed to validate policy", zap.Error(err))
		return nil, err
	}

	return body, nil
}
