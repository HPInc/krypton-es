// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"testing"

	"github.com/HPInc/krypton-es/es/service/structs"
	"github.com/google/uuid"
	pgx "github.com/jackc/pgx/v5"
)

// fail enroll test
func TestFailEnrollRecord(t *testing.T) {
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
	ee := &structs.EnrollError{
		EnrollId:     er.Id.String(),
		ErrorCode:    123,
		ErrorMessage: "failed to generate certificate",
	}
	if err = FailEnrollRecord(ee); err != nil {
		handleError(t, err)
	}
	_, err = GetEnrollDetailsById(er.Id)
	if err == nil || !IsDbErrorNoRows(err) {
		t.Errorf("Expected = %v, got %v", pgx.ErrNoRows, err)
	}
}
