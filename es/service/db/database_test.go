// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/HPInc/krypton-es/es/service/structs"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	log *zap.Logger
)

func TestMain(m *testing.M) {
	log, _ = zap.NewProduction(zap.AddCaller())
	InitTest(log, &testDbConfig)
	defer Shutdown()
	os.Exit(m.Run())
}

// ping test
func TestDatabasePing(t *testing.T) {
	if err := Ping(); err != nil {
		t.Errorf("Ping failed. Expected no error. Got %v", err)
	}
}

// shutdown and re-init test
func TestShutdownAndReInit(t *testing.T) {
	Shutdown()
	Init(log, &testDbConfig)
	TestDatabasePing(t)
}

// create single enroll
func TestCreateEnrollRecord(t *testing.T) {
	userId := uuid.New().String()
	tenantId := uuid.New().String()
	csrHash := uuid.New().String()
	_, err := CreateEnrollRecord(userId, tenantId, csrHash)
	if err != nil {
		handleError(t, err)
	}
}

// create single enroll and get back by id
func TestGetEnrollStatusById(t *testing.T) {
	er, err := newEnroll()
	if err != nil {
		handleError(t, err)
	}
	_, err = GetEnrollStatus(er.Id)
	if err != nil {
		handleError(t, err)
	}
}

// fetch single enroll by an invalid id
func TestGetEnrollStatusByNonExistentIdFails(t *testing.T) {
	entry, err := GetEnrollStatus(uuid.Nil)
	if err == nil {
		t.Errorf("Expected error. Got no error")
	} else if entry != nil {
		t.Errorf("Expected nil status, got %v", entry)
	}
}

// create and update enroll record
func TestNonExistentUpdateEnrollRecordFails(t *testing.T) {
	dc := structs.EnrollResult{
		EnrollId:    uuid.New(),
		DeviceId:    uuid.New(),
		Certificate: "cert bytes",
	}
	err := UpdateEnrollRecord(&dc)
	if err == nil {
		t.Errorf("Expected error on non existent update. Got no error")
	}
}

// create and update enroll record
func TestUpdateEnrollRecord(t *testing.T) {
	er, err := newEnroll()
	if err != nil {
		handleError(t, err)
	}
	dc := structs.EnrollResult{
		EnrollId:    er.Id,
		DeviceId:    uuid.New(),
		Certificate: "cert bytes",
	}
	if err = UpdateEnrollRecord(&dc); err != nil {
		handleError(t, err)
	}

}

func TestGetEnrollDetailsById(t *testing.T) {
	er, err := newEnroll()
	if err != nil {
		handleError(t, err)
	}
	dc := structs.EnrollResult{
		EnrollId:    er.Id,
		DeviceId:    uuid.New(),
		Certificate: "cert bytes",
	}
	err = UpdateEnrollRecord(&dc)
	if err != nil {
		handleError(t, err)
	}
	dcOut, err := GetEnrollDetailsById(er.Id)
	if err != nil {
		handleError(t, err)
	}
	if dc.EnrollId != dcOut.EnrollId {
		t.Errorf("Expected EnrollId = %v, got %v", dc.EnrollId, dcOut.EnrollId)
	} else if dc.DeviceId != dcOut.DeviceId {
		t.Errorf("Expected DeviceId = %v, got %v", dc.DeviceId, dcOut.DeviceId)
	} else if dc.Certificate != dcOut.Certificate {
		t.Errorf("Expected Certificate = %v, got %v", dc.Certificate, dcOut.Certificate)
	}
}

func TestPendingEnrollCount(t *testing.T) {
	cleanEnrollTable()
	i, err := GetPendingEnrollCount()
	if err != nil {
		handleError(t, err)
	} else if i != 0 {
		t.Errorf("Expected pending enroll: 0. Got: %d", i)
	}
	newEnroll()
	i, err = GetPendingEnrollCount()
	if err != nil {
		handleError(t, err)
	} else if i != 1 {
		t.Errorf("Expected pending enroll: 1. Got: %d", i)
	}
}

func TestHasCSRHash(t *testing.T) {
	csrHash := uuid.New().String()
	CreateEnrollRecord(uuid.New().String(), uuid.New().String(), csrHash)
	ok, err := HasCSRHash(csrHash)
	if err != nil {
		handleError(t, err)
	} else if !ok {
		t.Errorf("Expected ok. Got %v", ok)
	}
}

func TestNonExistentCSRHashFails(t *testing.T) {
	ok, err := HasCSRHash(uuid.New().String())
	if err != nil {
		handleError(t, err)
	} else if ok {
		t.Errorf("Expected not ok. Got %v", ok)
	}
}

func handleError(t *testing.T, err error) {
	if err == nil {
		return
	}
	t.Errorf("Expected no error. Got %v", err)
}

func expectNoError(t *testing.T, err error) {
	handleError(t, err)
}

func expectSomeError(t *testing.T, err error) {
	if err != nil {
		return
	}
	t.Errorf("Expected error. Got %v", err)
}

func expectError(t *testing.T, err, expected error) {
	if err != expected {
		t.Errorf("Expected %v. got %v", expected, err)
	}
}

func cleanEnrollTable() {
	ctx, cancelFunc := context.WithTimeout(gCtx, dbTimeout)
	defer cancelFunc()
	gDbPool.Exec(ctx, "DELETE FROM enroll")
}

func newEnroll() (*structs.DeviceEntry, error) {
	return newEnrollWithTenantId(uuid.New().String())
}

func newEnrollWithTenantId(tenantId string) (*structs.DeviceEntry, error) {
	userId := uuid.New().String()
	csrHash := uuid.New().String()
	return CreateEnrollRecord(userId, tenantId, csrHash)
}

func retryWait(count int, fn func() bool) bool {
	for i := 0; i < count; i++ {
		time.Sleep(1 * time.Second)
		if fn() {
			return true
		}
	}
	return false
}
