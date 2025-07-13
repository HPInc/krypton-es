// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/HPInc/krypton-es/es/service/db"
	"github.com/HPInc/krypton-es/es/service/structs"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

/*
	Get /enroll/{enroll_id}
	Get enroll status by id

Returns:
- 200
  - {"device_id": <uuid>, "cert": <base64 encoded cert>}

Errors:
- 400
  - Malformed or missing Authorization header
  - Malformed or missing payload

- 401
  - Could not verify token
  - Token expired or not yet valid

- 405
  - Must be GET

- 429
  - Not ready yet / too many requests.
  - "Retry-After:<delay seconds>" header is included in response.
  - See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After

- 500
  - should not be here. yet, here we are.
*/
var EnrollmentStatusHandler = enrollHandler(EnrollStatus)

func EnrollStatus(w http.ResponseWriter, r *http.Request) *enrollError {
	ei, err := GetEnrollInfoFromToken(r)
	if err != nil {
		return &enrollError{err, http.StatusUnauthorized}
	}

	id, enroll_error := getUUIDParam(r, paramEnrollID)
	if enroll_error != nil {
		return enroll_error
	}
	return getEnrollStatus(w, id, ei)
}

func getEnrollStatus(w http.ResponseWriter, id uuid.UUID, ei *EnrollInfo) *enrollError {
	var entry *structs.EnrollStatus
	var err error
	entry, err = db.GetEnrollStatus(id)
	// if there is a lookup error, consider an unenroll status if we have a device token
	if err != nil && ei.DeviceId != "" {
		return getUnenrollStatus(w, id, ei)
	}
	// if there is an error or if entry does not match token details, dont go further.
	if err != nil || entry.TenantId != ei.TenantId || entry.UserId != ei.UserId {
		esLogger.Error("Failed to match enroll record",
			zap.String("token_tenant_id", ei.TenantId),
			zap.String("token_user_id", ei.UserId),
			zap.String("entry_tenant_id", entry.TenantId),
			zap.String("entry_user_id", entry.UserId),
			zap.Error(err))
		return &enrollError{
			fmt.Errorf("id: %s is not found", id),
			http.StatusNotFound,
		}
	}
	// collect device id from record if non empty
	deviceId := ""
	if entry.DeviceId != uuid.Nil {
		deviceId = entry.DeviceId.String()
	}
	// if token has device id (device token), it should only access
	// records matching the device id
	if ei.DeviceId != "" {
		if deviceId != ei.DeviceId {
			esLogger.Error("Failed to match enroll record",
				zap.String("token_device_id", ei.DeviceId),
				zap.String("entry_device_id", deviceId))
			return &enrollError{
				fmt.Errorf("id: %s is not found", id),
				http.StatusNotFound,
			}
		}
	}
	switch entry.Status {
	case 0:
		writeRetryAfter(w, getRetryAfterHint())
		return &enrollError{
			ErrRequestInProgress,
			http.StatusTooManyRequests,
		}
	case 1:
		return getCompletedEnroll(w, id)
	default:
		return &enrollError{
			fmt.Errorf("id: %s is not found", id),
			http.StatusNotFound,
		}
	}
}

func writeRetryAfter(w http.ResponseWriter, retryAfter int) {
	w.Header().Add(headerRetryAfter, fmt.Sprintf("%d", retryAfter))
}

// compute retry after with current strategy
func getRetryAfterHint() int {
	// db's average enroll time will engage the current cache
	// strategy for enroll times. check cache config for details
	avgEnrollSeconds, err := db.GetAverageEnrollTime()
	if err != nil {
		esLogger.Error("Failed to get average enroll time",
			zap.Int("Defaulting to max retry", gServerConfig.MaxRetryAfterSeconds),
			zap.Error(err))
		return gServerConfig.MaxRetryAfterSeconds
	}
	if avgEnrollSeconds <= 0 {
		avgEnrollSeconds = gServerConfig.RetryAfterSeconds
	}
	// just an optimistic iota add to average.
	return avgEnrollSeconds + 1
}

func getCompletedEnroll(w http.ResponseWriter, id uuid.UUID) *enrollError {
	dc, err := db.GetEnrollDetailsById(id)
	if err != nil {
		return &enrollError{ErrLookupEnroll, http.StatusNotFound}
	}
	jsonstring, err := json.Marshal(dc)
	if err != nil {
		return &enrollError{ErrInternal, http.StatusInternalServerError}
	} else {
		fmt.Fprintf(w, "%s", jsonstring)
	}
	return nil
}
