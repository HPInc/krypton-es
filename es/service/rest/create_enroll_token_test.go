// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"net/http"
	"testing"
)

const (
	createEnrollTokenUrl = "/api/v1/enroll_token"
)

// create_enroll_token should be a post method
func TestCreateEnrollTokenWithGetMethodFailsWith405(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, createEnrollTokenUrl, nil)
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusMethodNotAllowed, response.Code)
	allowHeader := response.Header().Get("Allow")
	if allowHeader != "DELETE,POST" {
		t.Errorf("Expected Allow: DELETE,POST, Got %s\n", allowHeader)
	}
}

// create enroll_token must fail if no token type header
func TestCreateEnrollTokenWithoutTokenTypeHeaderFailsWith400(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, createEnrollTokenUrl, nil)
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusBadRequest, response.Code)
}

// requests without token type header fails with 400
func TestCreateEnrollTokenWithInvalidTokenTypeHeaderFailsWith400(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, createEnrollTokenUrl, nil)
	req.Header.Set(headerTokenType, "invalid_token_type")
	req.Header.Set("Authorization", "Bearer 123")
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusBadRequest, response.Code)
}

// create_enroll_token requests needs a bearer token
func TestCreateEnrollTokenWithoutBearerFailsWith401(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, createEnrollTokenUrl, nil)
	req.Header.Set(headerTokenType, "azuread")
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusUnauthorized, response.Code)
}
