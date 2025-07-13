// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	dstsclient "github.com/HPInc/krypton-es/es/service/client/dsts"
	"github.com/HPInc/krypton-es/es/service/db"
	"github.com/HPInc/krypton-es/es/service/policy"
	"go.uber.org/zap"
)

/*
/api/v1/enroll_token
Get enroll token for tenant

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
  - There is already an enrollment token created for this tenant

- 500
  - should not be here. yet, here we are.
*/
func CreateEnrollToken(w http.ResponseWriter, r *http.Request) *enrollError {
	startTime := time.Now()

	// parse bearer token
	ei, err := GetEnrollInfoFromToken(r)
	if err != nil {
		if IsTokenTypeHeaderError(err) {
			return &enrollError{ErrInvalidTokenType, http.StatusBadRequest}
		}
		return &enrollError{err, http.StatusUnauthorized}
	}
	// get tenant policy to get enroll token lifetime
	p, err := getPolicy(ei.TenantId)
	if err != nil {
		esLogger.Error("CreateEnrollToken: error looking up policy",
			zap.Error(err))
		return &enrollError{ErrCreateEnrollToken, http.StatusInternalServerError}
	}

	// get token lifetime in days specified in policy
	lifetimeDays, err := p.GetAttributeInt(policy.BulkEnrollTokenLifetimeDays)
	if err != nil {
		lifetimeDays = dstsclient.DefaultEnrollmentTokenLifetimeDays
	}

	// call dsts to create enrollment token
	token, clientErr := dstsclient.CreateEnrollmentToken(
		ei.TenantId, int32(lifetimeDays))
	if clientErr != nil {
		esLogger.Error("CreateEnrollToken: error from dsts",
			zap.Error(clientErr.Error))
		return &enrollError{ErrCreateEnrollToken, clientErr.HttpCode}
	}
	jsonString, err := json.Marshal(token)
	if err != nil {
		return &enrollError{ErrInternal, http.StatusInternalServerError}
	}
	esLogger.Info(
		"CreateEnrollToken",
		zap.String("TenantID", ei.TenantId),
		zap.String("Elapsed", time.Since(startTime).String()))
	fmt.Fprintf(w, "%s", string(jsonString))
	return nil
}

func getPolicy(tenantId string) (*policy.Policy, error) {
	id, err := db.GetPolicyId(tenantId)
	if err != nil {
		return policy.GetDefault(), nil
	}
	p, err := db.GetPolicy(*id, tenantId)
	if err != nil {
		return nil, err
	}
	return policy.FromString(p.Data)
}
