// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cache

import (
	"testing"

	"github.com/HPInc/krypton-es/es/service/structs"
	"github.com/google/uuid"
)

// Set a policy
func TestCreateAndGetPolicy(t *testing.T) {
	_, err := createPolicy()
	if err != nil {
		t.Errorf("Cache get policy failed. Expected no error. Got %v", err)
	}
}

// check returned policy params
func TestGetPolicy(t *testing.T) {
	p := &structs.Policy{
		Id:       uuid.New(),
		TenantId: uuid.New().String(),
		Data:     `{"key":"value"}`,
		Enabled:  true,
	}
	CreatePolicy(p)
	cached, err := GetPolicy(p.Id)
	if err != nil {
		t.Errorf("Cache get policy failed. Expected no error. Got %v", err)
	}
	if p.Id != cached.Id {
		t.Errorf("Cache get policy. Expected %v, got %v",
			p.Id, cached.Id)
	}
	if p.TenantId != cached.TenantId {
		t.Errorf("Cache get policy tenantId. Expected %s, got %s",
			p.TenantId, cached.TenantId)
	}
	if p.Data != cached.Data {
		t.Errorf("Cache get policy data. Expected %s, got %s",
			p.Data, cached.Data)
	}
}

// try to retrieve a policy that is not set
func TestGetNonExistentPolicy(t *testing.T) {
	id := uuid.New()
	cached, err := GetPolicy(id)
	if err == nil {
		t.Errorf("Cache get policy when not set. Expected error. Got none")
	}
	if cached != nil {
		t.Errorf("Cache get policy when not set. Expected nil. Got %v",
			cached)
	}
}

// try to create, then set policy data
// fetch and verify. expect success
func TestSetPolicyData(t *testing.T) {
	p, err := createPolicy()
	if err != nil {
		t.Errorf("Cache get policy failed. Expected no error. Got %v", err)
	}

	// set data
	p.Data = `{"key":"value","key1":"value1"}`
	UpdatePolicy(p)
	cached, err := GetPolicy(p.Id)
	if err != nil {
		t.Errorf("Cache set policy. Expected no error, got %v", err)
	}
	if p.Data != cached.Data {
		t.Errorf("Cache set data. Expected status: %s, got %s",
			p.Data, cached.Data)
	}
}

// try to create and delete a policy
// expect invalid status and not found error
func TestDeletePolicy(t *testing.T) {
	p, err := createPolicy()
	if err != nil {
		t.Errorf("Cache get policy failed. Expected no error. Got %v", err)
	}
	DeletePolicy(p.Id)
	if _, err = GetPolicy(p.Id); err == nil {
		t.Errorf("Cache get policy after delete. Expected none. Got %v",
			err)
	}
}

func createPolicy() (*structs.Policy, error) {
	p := &structs.Policy{
		Id:       uuid.New(),
		TenantId: uuid.New().String(),
		Data:     "{}",
		Enabled:  true,
	}
	CreatePolicy(p)
	return GetPolicy(p.Id)
}
