// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/HPInc/krypton-es/es/service/tokenmgr"
	"github.com/google/uuid"
)

const (
	getEnrollTokenUrl = "/api/v1/enroll_token"
)

// get_enroll_token should be a GET method
func TestGetEnrollTokenWithPostMethodFails(t *testing.T) {
	url := fmt.Sprintf("%s/%s", getEnrollTokenUrl, uuid.New().String())
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusMethodNotAllowed, response.Code)
	allowHeader := response.Header().Get("Allow")
	if allowHeader != "GET" {
		t.Errorf("Expected Allow: GET, Got %s\n", allowHeader)
	}
}

// get_enroll_token requests require a tenant_id path param which is a uuid
func TestGetEnrollTokenWithInvalidTenantIdFails(t *testing.T) {
	url := fmt.Sprintf("%s/%s", getEnrollTokenUrl, "123")
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusNotFound, response.Code)
}

// get enroll_token must fail if no token type header
func TestGetEnrollTokenWithMissingTypeHeaderFails(t *testing.T) {
	url := fmt.Sprintf("%s/%s", getEnrollTokenUrl, uuid.New().String())
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusBadRequest, response.Code)
}

// get_enroll_token requests need an app token
func TestGetEnrollTokenWithoutBearerFails(t *testing.T) {
	url := fmt.Sprintf("%s/%s", getEnrollTokenUrl, uuid.New().String())
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set(headerTokenType, string(tokenmgr.TokenTypeApp))
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusUnauthorized, response.Code)
}
