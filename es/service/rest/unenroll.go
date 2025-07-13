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
Unenroll call. Requires the following input
1. device token obtained from dsts (from authorization header)
This function will do the following
1. Verify the access token for a dsts token
Requires:
- Custom header: X-HP-Token-Type
  - Value: "enrollment"
  - Authorization header: Bearer <Token>

Returns:
- 202 (request is accepted and will eventually be processed)
  - id of pending enroll. This id can be used to follow up.

Errors:
- 400
  - Malformed or missing Authorization header

- 401
  - Could not verify token
  - Token expired or not yet valid
  - Audience or subject has invalid values (note: these are configurable in es)

- 405
  - Must be DELETE

- 500
  - should not be here. yet, here we are.
*/
func Unenroll(w http.ResponseWriter, r *http.Request) *enrollError {
	startTime := time.Now()

	// token must be a valid device token obtained from dsts
	if err := validateDeviceToken(r); err != nil {
		return &enrollError{err, http.StatusBadRequest}
	}

	ei, err := GetEnrollInfoFromToken(r)
	if err != nil {
		if IsTokenTypeHeaderError(err) {
			return &enrollError{err, http.StatusBadRequest}
		}
		return &enrollError{err, http.StatusUnauthorized}
	}

	deviceId, enroll_err := getUUIDParam(r, "device_id")
	if enroll_err != nil {
		return enroll_err
	}

	if ei.DeviceId != deviceId.String() {
		return &enrollError{
			ErrDeviceIdMismatch,
			http.StatusBadRequest,
		}
	}

	// create an unenroll record in db
	de, err := db.Unenroll(ei.TenantId, deviceId)
	if err != nil {
		return &enrollError{ErrUnenroll, getHttpCodeForDbError(err)}
	}

	// make unenroll payload here as we have all required
	// values available.
	ep := &enrollPayload{
		ID:        de.Id,
		RequestId: de.RequestId,
		TenantId:  ei.TenantId,
		DeviceId:  deviceId,
		Type:      requestPayloadTypeUnenroll,
	}

	if err = pushToPendingEnrollQueue(ep); err != nil {
		return &enrollError{err, http.StatusInternalServerError}
	}

	sendEnrollResponse(w, de, startTime)

	esLogger.Info(
		"Unenroll queued",
		zap.String("unenroll_id", de.Id.String()),
		zap.String("request_id", de.RequestId),
		zap.String("tenant_id", ei.TenantId),
		zap.String("device_id", deviceId.String()),
		zap.String("Elapsed", time.Since(startTime).String()))
	return nil
}
