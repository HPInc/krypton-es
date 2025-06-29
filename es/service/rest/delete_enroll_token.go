// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"net/http"
	"time"

	dstsclient "github.com/HPInc/krypton-es/es/service/client/dsts"
	"go.uber.org/zap"
)

/*
/api/v1/enroll_token
Delete enroll token issued for a tenant

Returns:
- 200
  - successfully deleted

Errors:
- 400
  - X-HP-TokenType header must be present and set to one of the user token types

- 401
  - Could not verify token
  - Token expired or not yet valid

- 405
  - Must be DELETE

- 404
  - There is no enrollment token created for this tenant

- 500
  - should not be here. yet, here we are.
*/
func DeleteEnrollToken(w http.ResponseWriter, r *http.Request) *enrollError {
	startTime := time.Now()

	// parse bearer token
	ei, err := GetEnrollInfoFromToken(r)
	if err != nil {
		if IsTokenTypeHeaderError(err) {
			return &enrollError{ErrInvalidTokenType, http.StatusBadRequest}
		}
		return &enrollError{err, http.StatusUnauthorized}
	}
	// call dsts to create enrollment token
	clientErr := dstsclient.DeleteEnrollmentToken(ei.TenantId)
	if clientErr != nil {
		esLogger.Error("DeleteEnrollToken: error from dsts",
			zap.Error(clientErr.Error))
		return &enrollError{clientErr.Error, clientErr.HttpCode}
	}
	esLogger.Info(
		"DeleteEnrollToken",
		zap.String("TenantID", ei.TenantId),
		zap.String("Elapsed", time.Since(startTime).String()))
	return nil
}
