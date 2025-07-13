// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/HPInc/krypton-es/es/service/structs"
	"github.com/google/uuid"
)

func TestPolicyGetDoesNotAllowPost(t *testing.T) {
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

func TestPolicyGetFailsForAnotherTenant(t *testing.T) {
	p1, _, err := createNewPolicy()
	if err != nil {
		t.Errorf("Error creating policy, %v", err)
	}
	_, err = getTestPolicy(p1.Id, "")
	if err == nil {
		t.Errorf("Expected not found error getting policy. Got %v", err)
	}
}

// get policy succeeds when params are correct
// note we use the same token. getBearerToken gives
// back a random tenant id for each call.
func TestPolicyGetSucceeds(t *testing.T) {
	bearerToken := getBearerToken()
	p1, _, err := createNewPolicyWithBearer(bearerToken)
	if err != nil {
		t.Errorf("Error creating policy, %v", err)
	}
	p2, err := getTestPolicy(p1.Id, bearerToken)
	if err != nil {
		t.Errorf("Error getting policy, %v", err)
	}
	if p1.Id != p2.Id {
		t.Errorf("Expected id:%v, found:%v", p1.Id, p2.Id)
	}
}

// quick get
func getTestPolicy(id uuid.UUID, bearerToken string) (*structs.Policy, error) {
	if bearerToken == "" {
		bearerToken = getBearerToken()
	}
	path := fmt.Sprintf("/api/v1/policy/%v", id)
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	req.Header.Set(headerTokenType, "test")
	req.Header.Set(headerAuthorization, bearerToken)
	resp := executeTestRequest(req)
	if resp.Code != http.StatusOK {
		return nil, errors.New(
			fmt.Sprintf("Error: %d, Failed to get policy", resp.Code))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := &structs.Policy{}
	if err = json.Unmarshal(body, r); err != nil {
		return nil, err
	}
	return r, nil
}
