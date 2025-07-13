// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPolicyCreateDoesNotAllowGet(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/policy", nil)
	executeTestRequest(req)
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusMethodNotAllowed, response.Code)
	allowHeader := response.Header().Get("Allow")
	if allowHeader != "POST" {
		t.Errorf("Expected Allow: POST, Got %s\n", allowHeader)
	}
}

// must provide content-type: application/json
func TestPolicyCreateWithoutContentType(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/policy", nil)
	req.Header.Set(headerTokenType, "azuread")
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusUnsupportedMediaType, response.Code)
}

// must provide auth
func TestPolicyCreateWithoutAuthHeader(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/policy", nil)
	req.Header.Set(headerTokenType, "azuread")
	req.Header.Set(headerContentType, contentTypeJson)
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusUnauthorized, response.Code)
}

// should have content for call
func TestPolicyCreateWithoutContent(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/policy", nil)
	req.Header.Set(headerTokenType, "test")
	req.Header.Set(headerContentType, contentTypeJson)
	req.Header.Set(headerAuthorization, getBearerToken())
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusBadRequest, response.Code)
}

// should have valid json content
func TestPolicyCreateWithoutValidJson(t *testing.T) {
	invalidJson := []byte("<>")
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/policy", bytes.NewBuffer(invalidJson))
	req.Header.Set(headerTokenType, "test")
	req.Header.Set(headerContentType, contentTypeJson)
	req.Header.Set(headerAuthorization, getBearerToken())
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusBadRequest, response.Code)
}

// should have non empty json
func TestPolicyCreateWithEmptyJson(t *testing.T) {
	invalidJson := []byte(`{}`)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/policy", bytes.NewBuffer(invalidJson))
	req.Header.Set(headerTokenType, "test")
	req.Header.Set(headerContentType, contentTypeJson)
	req.Header.Set(headerAuthorization, getBearerToken())
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusBadRequest, response.Code)
}

// should pass
func TestPolicyCreate(t *testing.T) {
	_, _, err := createNewPolicy()
	if err != nil {
		t.Errorf("Error creating policy, %v", err)
	}
}

// attempting to create policy again should fail with 409
func TestPolicyCreateAgain(t *testing.T) {
	bearerToken := getBearerToken()
	_, _, err := createNewPolicyWithBearer(bearerToken)
	if err != nil {
		t.Errorf("Error creating policy, %v", err)
	}
	_, resp, err := createNewPolicyWithBearer(bearerToken)
	if resp.Code != http.StatusConflict {
		t.Errorf("Expected %d, got %d", http.StatusConflict, resp.Code)
	}
}

// use a bearer token with a random tenant
func createNewPolicy() (*createPolicyResult, *httptest.ResponseRecorder, error) {
	return createNewPolicyWithBearer(getBearerToken())
}

// quick create
func createNewPolicyWithBearer(bearerToken string) (*createPolicyResult, *httptest.ResponseRecorder, error) {
	if bearerToken == "" {
		bearerToken = getBearerToken()
	}
	validJson := []byte(`{"version":1}`)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/policy", bytes.NewBuffer(validJson))
	req.Header.Set(headerTokenType, "test")
	req.Header.Set(headerContentType, contentTypeJson)
	req.Header.Set(headerAuthorization, bearerToken)
	resp := executeTestRequest(req)
	if resp.Code != http.StatusOK {
		return nil, resp, errors.New(
			fmt.Sprintf("Error: %d, Failed to create policy", resp.Code))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	r := &createPolicyResult{}
	if err = json.Unmarshal(body, r); err != nil {
		return nil, nil, err
	}
	return r, resp, nil
}
