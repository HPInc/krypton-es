// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"testing"

	"github.com/HPInc/krypton-es/es/service/cache"
	"github.com/HPInc/krypton-es/es/service/structs"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// create single policy
func TestCreatePolicy(t *testing.T) {
	tenantId := uuid.New().String()
	data := "{}"
	p, err := newPolicyWithParams(tenantId, data)
	handleError(t, err)

	if p.Id.String() == "" {
		t.Errorf("Expected valid uuid as id. got %v", p.Id)
	}
	if tenantId != tenantId {
		t.Errorf("Expected tenantId: %s, got %s", tenantId, p.TenantId)
	}
	if data != p.Data {
		t.Errorf("Expected data: %s, got %s", data, p.Data)
	}
}

// get a policy
func TestGetPolicy(t *testing.T) {
	// create a policy
	p1, err := newPolicy()
	handleError(t, err)

	// get the policy by id that just got created
	p2, err := GetPolicy(p1.Id, p1.TenantId)
	handleError(t, err)

	// ensure both policies are equal
	if p1.Id != p2.Id {
		t.Errorf("Expected id: %v. got %v", p1.Id, p2.Id)
	}
	if p1.TenantId != p2.TenantId {
		t.Errorf("Expected tenant id: %v. got %v", p1.TenantId, p2.TenantId)
	}
	if p1.Enabled != p2.Enabled {
		t.Errorf("Expected enabled: %v. got %v", p1.Enabled, p2.Enabled)
	}
	if !p1.CreatedAt.Equal(p2.CreatedAt) {
		t.Errorf("Expected createdAt: %v. got %v", p1.CreatedAt, p2.CreatedAt)
	}
	if !p1.UpdatedAt.IsZero() || !p2.UpdatedAt.IsZero() {
		t.Errorf("Expected empty updatedAt. Got: %v, %v", p1.UpdatedAt, p2.UpdatedAt)
	}
}

// update a policy
func TestUpdatePolicy(t *testing.T) {
	// create a policy
	p1, err := newPolicy()
	handleError(t, err)

	newData := `{"updated":true}`
	p1.Data = newData

	// update policy
	err = UpdatePolicy(p1)
	handleError(t, err)

	// get the policy
	p2, err := GetPolicy(p1.Id, p1.TenantId)
	handleError(t, err)

	// ensure both policies are equal
	if p1.Id != p2.Id {
		t.Errorf("Expected id: %v. got %v", p1.Id, p2.Id)
	}
	if p1.TenantId != p2.TenantId {
		t.Errorf("Expected tenant id: %v. got %v", p1.TenantId, p2.TenantId)
	}
	if !p1.CreatedAt.Equal(p2.CreatedAt) {
		t.Errorf("Expected createdAt: %v. got %v", p1.CreatedAt, p2.CreatedAt)
	}
	if !p2.UpdatedAt.After(p1.CreatedAt) {
		t.Errorf("Expected updatedAt after createdAt. Got: %v, %v",
			p1.CreatedAt, p2.UpdatedAt)
	}
	if p2.Data != newData {
		t.Errorf("Expected data: %v. got %v", newData, p2.Data)
	}
}

// delete policy
func TestDeletePolicy(t *testing.T) {
	// create a policy
	p, err := newPolicy()
	handleError(t, err)

	err = DeletePolicy(p.Id, p.TenantId)
	handleError(t, err)

	// check by get
	_, err = GetPolicyId(p.TenantId)
	expectSomeError(t, err)
}

// delete non existent policy
func TestDeleteNonExistentPolicy(t *testing.T) {
	err := DeletePolicy(uuid.New(), uuid.New().String())
	expectError(t, err, ErrNoRows)
}

// get policy by tenant
func TestGetPolicyId(t *testing.T) {
	// create a policy
	p, err := newPolicy()
	handleError(t, err)

	id, err := GetPolicyId(p.TenantId)
	handleError(t, err)

	if *id != p.Id {
		t.Errorf("Expected: %v, got %v", p.Id, id)
	}
}

func TestGetPolicyIdFails(t *testing.T) {
	id, err := GetPolicyId(uuid.New().String())
	if err != ErrNoRows {
		t.Errorf("Expected %v. got %v", ErrNoRows, err)
	}
	if id != nil {
		t.Errorf("Expected nil id. got %v", id)
	}
}

// tests for cache
// creating a db policy should create a cache entry
func TestCreatePolicyCreatesCache(t *testing.T) {
	// create a policy
	p, err := newPolicy()
	handleError(t, err)

	// make sure we can get the policy back.
	// we are not counting on cache here
	_, err = GetPolicyId(p.TenantId)
	handleError(t, err)

	// policy cache gets created from a goroutine
	// this might need retry
	p1, err := cache.GetPolicy(p.Id)
	handleError(t, err)

	if p.Id != p1.Id {
		t.Errorf("Expected %v. got %v", p.Id, p1.Id)
	}
	if p.Data != p1.Data {
		t.Errorf("Expected %v. got %v", p.Data, p1.Data)
	}
}

// updating a db policy should update cache entry
func TestUpdatePolicyUpdatesCache(t *testing.T) {
	// create a policy
	p1, err := newPolicy()
	handleError(t, err)

	newData := `{"updated":true}`
	p1.Data = newData

	// update policy
	err = UpdatePolicy(p1)
	handleError(t, err)

	// get the policy
	_, err = GetPolicy(p1.Id, p1.TenantId)
	handleError(t, err)

	p2, err := cache.GetPolicy(p1.Id)
	expectNoError(t, err)

	if p2.Data != newData {
		t.Errorf("Expected %v. got %v", newData, p2.Data)
	}
}

func getCachePolicy(id uuid.UUID) func() bool {
	return func() bool {
		_, err := cache.GetPolicy(id)
		return err != redis.Nil
	}
}

// removing a db policy should remove cache entry
func TestDeletePolicyDeletesCache(t *testing.T) {
	// create a policy
	p, err := newPolicy()
	handleError(t, err)

	if !retryWait(5, getCachePolicy(p.Id)) {
		t.Errorf("could not get cached policy: %v", p.Id)
	}

	err = DeletePolicy(p.Id, p.TenantId)
	handleError(t, err)

	// check by get.
	_, err = GetPolicyId(p.TenantId)
	expectSomeError(t, err)

	_, err = cache.GetPolicy(p.Id)
	expectSomeError(t, err)
}

func newPolicy() (*structs.Policy, error) {
	return newPolicyWithParams(uuid.New().String(), "{}")
}

func newPolicyWithParams(tenantId, data string) (*structs.Policy, error) {
	return CreatePolicy(tenantId, data)
}
