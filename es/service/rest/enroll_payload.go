// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/HPInc/krypton-es/es/service/config"
	"github.com/google/uuid"
)

type enrollPayload struct {
	ID                uuid.UUID `json:"id"`
	CSR               string    `json:"csr"`
	RequestId         string    `json:"request_id"`
	TenantId          string    `json:"tenant_id"`
	DeviceId          uuid.UUID `json:"device_id,omitempty"`
	CSRHash           string    `json:"-"`
	Type              string    `json:"type"`
	ManagementService string    `json:"mgmt_service"`
	HardwareHash      string    `json:"hardware_hash"`
}

func GetEnrollPayload(r *http.Request) (*enrollPayload, error) {
	var ep enrollPayload
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(body, &ep); err != nil {
		return nil, err
	}
	if ep.CSRHash, err = ep.getCSRHash(); err != nil {
		return nil, err
	}
	ep.Type = requestPayloadTypeEnroll
	return &ep, nil
}

func GetRenewEnrollPayload(r *http.Request) (*enrollPayload, error) {
	ep, err := GetEnrollPayload(r)
	if err != nil {
		return nil, err
	}
	ep.Type = requestPayloadTypeReenroll
	return ep, nil
}

func (ep enrollPayload) getCSRHash() (string, error) {
	_, err := base64.StdEncoding.DecodeString(ep.CSR)
	if err != nil {
		return "", err
	}
	bs := sha256.Sum256([]byte(ep.CSR))
	return fmt.Sprintf("%x\n", bs), nil
}

func (ep enrollPayload) ValidateManagementService() error {
	if ep.ManagementService == "" {
		return errors.New("please specify mgmt_service in payload")
	} else if !ep.HasManagementService() {
		return fmt.Errorf(
			"%s is not a valid mgmt_service. Valid values are %v",
			ep.ManagementService, config.GetManagementServices())
	}
	return nil
}

func (ep enrollPayload) HasManagementService() bool {
	for _, v := range config.GetManagementServices() {
		if ep.ManagementService == v {
			return true
		}
	}
	return false
}
