// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/HPInc/krypton-es/es/service/db"
	"github.com/HPInc/krypton-es/es/service/structs"
	"github.com/google/uuid"
)

// device tokens
// no enroll record, no match
func TestDeviceTokenGetEnrollStatusNoEnrollRecord(t *testing.T) {
	_, bearerToken := getDeviceToken()
	enrollId := uuid.New().String()
	queryUrl := fmt.Sprintf("/api/v1/enroll/%s", enrollId)
	req, _ := http.NewRequest(http.MethodGet, queryUrl, nil)
	req.Header.Set(headerTokenType, "device")
	req.Header.Set("Authorization", bearerToken)
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusNotFound, response.Code)
}

// make an enroll record in db with a known tenant id
// make a device token with the same tenant id
// we should get not found as device id in enroll record will not match
func TestDeviceTokenGetEnrollStatusNoEnrollRecordDeviceIdMatch(t *testing.T) {
	info, bearerToken := getDeviceToken()
	enrollEntry, err := newEnroll(info)
	handleError(t, err)

	queryUrl := fmt.Sprintf("/api/v1/enroll/%s", enrollEntry.Id)
	req, _ := http.NewRequest(http.MethodGet, queryUrl, nil)
	req.Header.Set(headerTokenType, "device")
	req.Header.Set("Authorization", bearerToken)
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusNotFound, response.Code)
}

// make an enroll record in db with a known tenant id
// update enroll record to status completed (this will end up having a device id)
// make a device token with the same tenant id and device id
// we should get ok as everything matches
func TestDeviceTokenGetEnrollStatusOk(t *testing.T) {
	info, bearerToken := getDeviceToken()
	entry, err := updateEnroll(info)
	handleError(t, err)

	queryUrl := fmt.Sprintf("/api/v1/enroll/%s", entry.Id)
	req, _ := http.NewRequest(http.MethodGet, queryUrl, nil)
	req.Header.Set(headerTokenType, "device")
	req.Header.Set("Authorization", bearerToken)
	response := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusOK, response.Code)
}

// helper functions
func updateEnroll(info *testTokenInfo) (*structs.DeviceEntry, error) {
	entry, err := newEnroll(info)
	if err != nil {
		return nil, err
	}
	dc := structs.EnrollResult{
		EnrollId:    entry.Id,
		DeviceId:    uuid.MustParse(info.deviceId),
		Certificate: "cert bytes",
	}
	return entry, db.UpdateEnrollRecord(&dc)
}

func newEnroll(info *testTokenInfo) (*structs.DeviceEntry, error) {
	csrHash := uuid.New().String()
	return db.CreateEnrollRecord(info.tenantId, info.userId, csrHash)
}
