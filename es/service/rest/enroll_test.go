// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"net/http"
	"testing"
)

func TestEnrollWithGetMethodFailsWith405(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/enroll", nil)
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusMethodNotAllowed, response.Code)
	allowHeader := response.Header().Get("Allow")
	if allowHeader != "POST" {
		t.Errorf("Expected Allow: POST, Got %s\n", allowHeader)
	}
}

func TestEnrollWithoutTokenTypeHeaderFailsWith400(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/enroll", nil)
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestEnrollWithInvalidTokenTypeHeaderFailsWith400(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/enroll", nil)
	req.Header.Set(headerTokenType, "invalid_token_type")
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusUnauthorized, response.Code)
}

func TestEnrollWithoutBearerFailsWith401(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/enroll", nil)
	req.Header.Set(headerTokenType, "azuread")
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusUnauthorized, response.Code)
}
