// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"errors"
	"net/http"

	"github.com/HPInc/krypton-es/es/service/metrics"
	"go.uber.org/zap"
)

/*
consolidate error handling. See https://go.dev/blog/error-handling-and-go
*/
type enrollError struct {
	Error error
	Code  int
}

type JsonError struct {
	ErrorString string `json:"error"`
	Code        int    `json:"code"`
}

type enrollHandler func(http.ResponseWriter, *http.Request) *enrollError

func esHandlerFunc(inner enrollHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := inner(w, r)
		if err != nil {
			// Skip logging errors for status requests on pending enroll
			// requests.
			if !errors.Is(err.Error, ErrRequestInProgress) {
				esLogger.Error("Error serving http request", zap.Error(err.Error))
				metrics.ReportRestError(r.Method, err.Code)
			}
			// send errors as json
			sendJsonError(w, &JsonError{err.Error.Error(), err.Code})
		}
	})
}
