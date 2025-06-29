// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cache

import (
	"os"
	"testing"

	"github.com/HPInc/krypton-es/es/service/config"
	"github.com/HPInc/krypton-es/es/service/structs"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	log *zap.Logger
)

const (
	defaultCacheUser     = "krypton"
	defaultCachePassword = "krypton"
)

func getServer() string {
	e := os.Getenv("ES_CACHE_SERVER")
	if e != "" {
		return e
	} else {
		return "localhost"
	}
}

func TestMain(m *testing.M) {
	log, _ = zap.NewProduction(zap.AddCaller())
	cacheConfig := config.Cache{
		Server:   getServer(),
		Port:     6379,
		User:     defaultCacheUser,
		Password: defaultCachePassword,
		Enabled:  true,
	}
	Init(log, &cacheConfig)
	defer Shutdown()
	os.Exit(m.Run())
}

// set enroll status and retrieve it
// expect success
func TestSuccessfulStatusCacheGet(t *testing.T) {
	id := uuid.New()
	deviceId := uuid.Nil
	status := 1
	CreateEnrollStatus(id, "", "", deviceId, status)
	cached, err := GetEnrollStatus(id)
	if err != nil {
		t.Errorf("Cache get enroll status failed. Expected no error. Got %v", err)
	}
	if status != cached.Status {
		t.Errorf("Cache get enroll status. Expected %d, got %d",
			status, cached.Status)
	}
	if deviceId != cached.DeviceId {
		t.Errorf("Cache get enroll deviceId. Expected %v, got %v",
			deviceId, cached.DeviceId)
	}
}

// set enroll status and retrieve it to check for status, tenant and user
// expect success
func TestSuccessfulStatusCacheGetEntry(t *testing.T) {
	id := uuid.New()
	deviceId := uuid.New()
	status := 1
	tenantId := uuid.New().String()
	userId := uuid.New().String()
	CreateEnrollStatus(id, tenantId, userId, deviceId, status)
	cached, err := GetEnrollStatus(id)
	if err != nil {
		t.Errorf("Cache get enroll status failed. Expected no error. Got %v", err)
	}
	if status != cached.Status {
		t.Errorf("Cache get enroll status. Expected %d, got %d",
			status, cached.Status)
	}
	if tenantId != cached.TenantId {
		t.Errorf("Cache get enroll tenantId. Expected %s, got %s",
			tenantId, cached.TenantId)
	}
	if userId != cached.UserId {
		t.Errorf("Cache get enroll userId. Expected %s, got %s",
			userId, cached.TenantId)
	}
	if deviceId != cached.DeviceId {
		t.Errorf("Cache get enroll deviceId. Expected %v, got %v",
			deviceId, cached.DeviceId)
	}
}

// try to retrieve a status that is not set
// expect invalid status
func TestInvalidEnrollIdFails(t *testing.T) {
	id := uuid.New()
	cached, err := GetEnrollStatus(id)
	if err == nil {
		t.Errorf("Cache get enroll status when not set. Expected error. Got none")
	}
	if cached != nil {
		t.Errorf("Cache get enroll status when not set. Expected nil. Got %v",
			cached)
	}
}

// try to create, then set status value
// fetch and verify. expect success
func TestSetStatusAndVerify(t *testing.T) {
	id := uuid.New()
	deviceId := uuid.New()
	var err error
	var cached *structs.EnrollStatus
	if cached, err = createEnrollStatus(id); err != nil {
		t.Errorf("Cache create. Expected no error, got %v", err)
	}
	if cached.Status != 0 {
		t.Errorf("Cache status. Expected status: %d, got %d",
			0, cached.Status)
	}
	// set status to 1
	status := 1
	SetEnrollStatus(id, deviceId, status)
	if cached, err = GetEnrollStatus(id); err != nil {
		t.Errorf("Cache set status. Expected no error, got %v", err)
	}
	if cached.Status != status {
		t.Errorf("Cache set status. Expected status: %d, got %d",
			status, cached.Status)
	}
	if cached.DeviceId != deviceId {
		t.Errorf("Cache set status. Expected deviceId: %v, got %v",
			deviceId, cached.DeviceId)
	}
}

// try to create and delete an enroll
// expect invalid status and not found error
func TestDeleteStatus(t *testing.T) {
	id := uuid.New()
	var err error
	if _, err = createEnrollStatus(id); err != nil {
		t.Errorf("Cache create. Expected no error, got %v", err)
	}
	DeleteEnrollStatusById(id)
	if _, err = GetEnrollStatus(id); err == nil {
		t.Errorf("Cache get enroll status after delete. Expected none. Got %v",
			err)
	}
}

// get status entry by id and compare status, tenant and user
// expect success
func TestEnrollStatusEntry(t *testing.T) {
	id := uuid.New()
	created, err := createEnrollStatus(id)
	if err != nil {
		t.Errorf("Cache create. Expected no error, got %v", err)
	}
	cached, err := getEnrollStatus(id)
	if err != nil {
		t.Errorf("Cache get status. Expected no error, got %v", err)
	}
	if cached.Status != created.Status {
		t.Errorf("Cache get status. Expected status: %d, got %d",
			created.Status, cached.Status)
	}
	if cached.TenantId != created.TenantId {
		t.Errorf("Cache get tenantId. Expected tenantId: %s, got %s",
			created.TenantId, cached.TenantId)
	}
	if cached.UserId != created.UserId {
		t.Errorf("Cache get userId. Expected userId: %s, got %s",
			created.UserId, cached.UserId)
	}
}

func createEnrollStatus(id uuid.UUID) (*structs.EnrollStatus, error) {
	tenantId := uuid.New().String()
	userId := uuid.New().String()
	deviceId := uuid.Nil
	CreateEnrollStatus(id, tenantId, userId, deviceId, 0)
	return GetEnrollStatus(id)
}
