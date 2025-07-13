// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// renew requires patch. check if post fails with 405
func TestRenewEnrollFailsForPost(t *testing.T) {
	renewEnrollUrl := fmt.Sprintf("/api/v1/enroll/%s", uuid.New())
	req, _ := http.NewRequest(http.MethodPost, renewEnrollUrl, nil)
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusMethodNotAllowed, response.Code)
}

// Renew enroll should only allow enrollment token types
func TestRenewEnrollFailsForNonEnrollmentTokenTypes(t *testing.T) {
	renewEnrollUrl := fmt.Sprintf("/api/v1/enroll/%s", uuid.New())
	req, _ := http.NewRequest(http.MethodPatch, renewEnrollUrl, nil)
	req.Header.Set(headerTokenType, "azuread")
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusBadRequest, response.Code)
}

// Renew enroll should fail with 401 when initial requirements are met
// method is PATCH
// token type is device
func TestRenewEnrollFailsAsExpectedAfterPreReqs(t *testing.T) {
	renewEnrollUrl := fmt.Sprintf("/api/v1/enroll/%s", uuid.New())
	req, _ := http.NewRequest(http.MethodPatch, renewEnrollUrl, nil)
	req.Header.Set(headerTokenType, "device")
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusUnauthorized, response.Code)
}
