// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"testing"

	"github.com/HPInc/krypton-es/es/service/structs"
	"github.com/google/uuid"
)

// renew enroll record
func TestRenewEnrollRecord(t *testing.T) {
	tenantId := uuid.New().String()
	er, err := newEnrollWithTenantId(tenantId)
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
	er2, err := RenewEnroll(
		tenantId, dc.DeviceId, er.UserId, "csrhash2")
	if err != nil {
		handleError(t, err)
	}
	dc.EnrollId = er2.Id
	dc.Certificate = "cert bytes2"
	if err = UpdateEnrollRecord(&dc); err != nil {
		handleError(t, err)
	}
	dcRe, err := GetEnrollDetailsById(er2.Id)
	if err != nil {
		handleError(t, err)
	}
	if dcRe.DeviceId != dc.DeviceId {
		t.Errorf("Expected DeviceId = %v, got %v", dc.DeviceId, dcRe.DeviceId)
	}
	if dcRe.Certificate != dc.Certificate {
		t.Errorf("Expected Certificate = %v, got %v", dc.Certificate, dcRe.Certificate)
	}
}

// renew enroll is expected to be done with device tokens
// device tokens will not have user id
func TestRenewEnrollWithoutUserId(t *testing.T) {
	tenantId := uuid.New().String()
	er, err := newEnrollWithTenantId(tenantId)
	if err != nil {
		handleError(t, err)
	}
	dc := structs.EnrollResult{
		EnrollId:    er.Id,
		DeviceId:    uuid.New(),
		Certificate: "cert bytes3",
	}
	if err = UpdateEnrollRecord(&dc); err != nil {
		handleError(t, err)
	}
	er2, err := RenewEnroll(
		tenantId, dc.DeviceId, "", "csrhash3")
	if err != nil {
		handleError(t, err)
	}
	dc.EnrollId = er2.Id
	dc.Certificate = "cert bytes4"
	if err = UpdateEnrollRecord(&dc); err != nil {
		handleError(t, err)
	}
	es, err := GetEnrollStatus(er2.Id)
	if err != nil {
		handleError(t, err)
	}
	if es.TenantId != tenantId {
		t.Errorf("Expected TenantId = %v, got %v", es.TenantId, tenantId)
	}
	if es.UserId != "" {
		t.Errorf("Expected UserId = %v, got %v", "", es.UserId)
	}

	dcRe, err := GetEnrollDetailsById(er2.Id)
	if err != nil {
		handleError(t, err)
	}
	if dcRe.DeviceId != dc.DeviceId {
		t.Errorf("Expected DeviceId = %v, got %v", dc.DeviceId, dcRe.DeviceId)
	}
	if dcRe.Certificate != dc.Certificate {
		t.Errorf("Expected Certificate = %v, got %v", dc.Certificate, dcRe.Certificate)
	}
}
