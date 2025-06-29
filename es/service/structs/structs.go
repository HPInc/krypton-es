// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package structs

import (
	"time"

	"github.com/google/uuid"
)

// payload from worker via sqs
// contains certificate info
type EnrollResult struct {
	Id                 uuid.UUID `json:"id"`
	EnrollId           uuid.UUID `json:"enroll_id"`
	DeviceId           uuid.UUID `json:"device_id"`
	Certificate        string    `json:"certificate"`
	ParentCertificates string    `json:"parent_certificates"`
}

// payload from worker via sqs
// contains unenroll or delete device entry
type UnenrollResult struct {
	// this id is allocated by es and sent over
	// as part of unenroll request
	UnenrollId uuid.UUID `json:"unenroll_id"`
	// request id from es for trace. used in logging
	RequestId string `json:"request_id"`
	TenantId  string `json:"tenant_id"`
	DeviceId  string `json:"device_id"`
}

type DeviceEntry struct {
	Id          uuid.UUID `json:"id"`
	TenantId    string    `json:"-"`
	UserId      string    `json:"-"`
	RequestId   string    `json:"request_id"`
	CreatedTime string    `json:"created_time"`
}

// enroll error message
type EnrollError struct {
	Id            string `json:"id"`
	EnrollId      string `json:"enroll_id"`
	ErrorCode     int    `json:"error_code"`
	ErrorMessage  string `json:"error_message"`
	ReceiptHandle string `json:"receipt_handle,omitempty"`
	Type          string `json:"type"`
}

// enroll status details
type EnrollStatus struct {
	Status   int       `json:"status"`
	TenantId string    `json:"tenant_id"`
	UserId   string    `json:"user_id"`
	DeviceId uuid.UUID `json:"device_id"`
}

// unenroll status details
type UnenrollStatus struct {
	Status   int    `json:"status"`
	TenantId string `json:"tenant_id"`
	DeviceId string `json:"device_id"`
}

// Policy
type Policy struct {
	Id        uuid.UUID `json:"id"`
	TenantId  string    `json:"tenant_id"`
	Data      string    `json:"data"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_time"`
	UpdatedAt time.Time `json:"updated_time,omitempty"`
}
