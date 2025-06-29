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
Delete enroll call.
Internal Delete is a maintenance call

Returns:
- 200
  - Returns {"count": <int>} indicating deleted record count

Errors:
- 405
  - Must be DELETE

- 500
  - internal server error
*/
func DeleteExpiredEnrolls(w http.ResponseWriter, r *http.Request) *enrollError {
	startTime := time.Now()

	// delete will use service config to determine expired records
	count, err := db.DeleteExpiredEnrolls(0)
	if err != nil {
		return &enrollError{err, http.StatusInternalServerError}
	}

	esLogger.Info(
		"Delete expired enroll records",
		zap.Int64("Count", count),
		zap.String("Elapsed", time.Since(startTime).String()))
	return nil
}
