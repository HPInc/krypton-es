// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestPolicyDeleteDoesNotAllowPost(t *testing.T) {
	path := fmt.Sprintf("/api/v1/policy/%v", uuid.New())
	req, _ := http.NewRequest(http.MethodPost, path, nil)
	executeTestRequest(req)
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusMethodNotAllowed, response.Code)
	allowHeader := response.Header().Get("Allow")
	if allowHeader != "GET,DELETE,PATCH" {
		t.Errorf("Expected Allow: GET,DELETE,PATCH, Got %s\n", allowHeader)
	}
}

func TestPolicyDeleteFailsForAnotherTenant(t *testing.T) {
	p1, _, err := createNewPolicy()
	if err != nil {
		t.Errorf("Error creating policy, %v", err)
	}
	if err = deletePolicy(p1.Id, ""); err == nil {
		t.Errorf("Expected not found error getting policy. Got %v", err)
	}
}

// delete policy succeeds when params are correct
// note we use the same token. getBearerToken gives
// back a random tenant id for each call.
func TestPolicyDeleteSucceeds(t *testing.T) {
	bearerToken := getBearerToken()
	p1, _, err := createNewPolicyWithBearer(bearerToken)
	if err != nil {
		t.Errorf("Error creating policy, %v", err)
	}
	if err = deletePolicy(p1.Id, bearerToken); err != nil {
		t.Errorf("Error deleting policy, %v", err)
	}
}

// quick delete
func deletePolicy(id uuid.UUID, bearerToken string) error {
	if bearerToken == "" {
		bearerToken = getBearerToken()
	}
	path := fmt.Sprintf("/api/v1/policy/%v", id)
	req, _ := http.NewRequest(http.MethodDelete, path, nil)
	req.Header.Set(headerTokenType, "test")
	req.Header.Set(headerAuthorization, bearerToken)
	resp := executeTestRequest(req)
	if resp.Code != http.StatusOK {
		return errors.New(
			fmt.Sprintf("Error: %d, Failed to delete policy", resp.Code))
	}
	return nil
}
