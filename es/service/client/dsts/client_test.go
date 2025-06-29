// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package dstsclient

import (
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
)

func TestGetHttpCode(t *testing.T) {
	runGetHttpCodeTest(t, codes.AlreadyExists, http.StatusConflict)
	runGetHttpCodeTest(t, codes.InvalidArgument, http.StatusBadRequest)
	runGetHttpCodeTest(t, codes.ResourceExhausted, http.StatusTooManyRequests)
	runGetHttpCodeTest(t, codes.Unauthenticated, http.StatusUnauthorized)
	runGetHttpCodeTest(t, codes.Internal, http.StatusInternalServerError)
	runGetHttpCodeTest(t, codes.Unknown, http.StatusInternalServerError)
}

func runGetHttpCodeTest(t *testing.T, grpcCode codes.Code, httpCode int) {
	code := getHttpCode(grpcCode)
	if code != httpCode {
		t.Errorf("Expected %d. Got %d", httpCode, code)
	}
}
