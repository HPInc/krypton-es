// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// unenroll requires delete. check if post fails with 405
func TestUnenrollFailsForPost(t *testing.T) {
	unEnrollUrl := fmt.Sprintf("/api/v1/enroll/%s", uuid.New())
	req, _ := http.NewRequest(http.MethodPost, unEnrollUrl, nil)
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusMethodNotAllowed, response.Code)
}

// Unenroll should only allow device tokens as authorization
func TestUnenrollFailsForNonDeviceTokens(t *testing.T) {
	unEnrollUrl := fmt.Sprintf("/api/v1/enroll/%s", uuid.New())
	req, _ := http.NewRequest(http.MethodDelete, unEnrollUrl, nil)
	req.Header.Set(headerTokenType, "azuread")
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusBadRequest, response.Code)
}

// Unenroll should fail with 401 when initial requirements are met
// method is DELETE
// token type expected is device
func TestUnenrollFailsAsExpectedAfterPreReqs(t *testing.T) {
	unEnrollUrl := fmt.Sprintf("/api/v1/enroll/%s", uuid.New())
	req, _ := http.NewRequest(http.MethodPatch, unEnrollUrl, nil)
	req.Header.Set(headerTokenType, "device")
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusUnauthorized, response.Code)
}
