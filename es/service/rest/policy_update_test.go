// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestPolicyUpdateDoesNotAllowPost(t *testing.T) {
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

func TestPolicyUpdateFailsForAnotherTenant(t *testing.T) {
	p1, _, err := createNewPolicy()
	if err != nil {
		t.Errorf("Error creating policy, %v", err)
	}
	data := `{"version":1,"attributes":{"BulkEnrollTokenLifetime":"value"}}`
	if err = updatePolicy(p1.Id, "", data); err == nil {
		t.Errorf("Expected not found error updating policy. Got %v", err)
	}
}

// update policy succeeds when params are correct
// note we use the same token. getBearerToken gives
// back a random tenant id for each call.
func TestPolicyUpdateSucceeds(t *testing.T) {
	bearerToken := getBearerToken()
	p1, _, err := createNewPolicyWithBearer(bearerToken)
	if err != nil {
		t.Errorf("Error creating policy, %v", err)
	}
	data := `{"version":1,"attributes":{"BulkEnrollTokenLifetime":"value"}}`
	if err = updatePolicy(p1.Id, bearerToken, data); err != nil {
		t.Errorf("Error getting policy, %v", err)
	}
}

// quick update
func updatePolicy(id uuid.UUID, bearerToken, data string) error {
	if bearerToken == "" {
		bearerToken = getBearerToken()
	}
	path := fmt.Sprintf("/api/v1/policy/%v", id)
	req, _ := http.NewRequest(http.MethodPatch, path, bytes.NewBuffer([]byte(data)))
	req.Header.Set(headerTokenType, "test")
	req.Header.Set(headerAuthorization, bearerToken)
	resp := executeTestRequest(req)
	if resp.Code != http.StatusOK {
		return errors.New(
			fmt.Sprintf("Error: %d, Failed to update policy", resp.Code))
	}
	return nil
}
