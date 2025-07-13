// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

import (
	"encoding/json"

	"go.uber.org/zap"
)

type EnrollPayload struct {
	// enroll id. this is the id of the enroll record in es
	ID string `json:"id"`
	// tenant id from enroll token. es adds this.
	TenantId string `json:"tenant_id"`
	// request id from es to track message
	RequestId string `json:"request_id"`
	// base64 encoded csr presented to es as part of enroll request
	// this is copied as is from es after validating for base64
	CSR string `json:"csr"`
	// es-worker will populate upon message receipt.
	// this is the queue receipt handle.
	ReceiptHandle string `json:"receipt_handle"`
	// intermediate device id placeholder. es-worker will
	// populate and use for internal purposes.
	DeviceId string `json:"device_id,omitempty"`
	// type of payload. enroll or renew_enroll or delete_enroll
	Type string `json:"type"`
	// destination management service
	// eg. HP Connect
	ManagementService string `json:"mgmt_service"`
	// hardware hash from device
	// if specified, pass to dsts device add
	HardwareHash string `json:"hardware_hash"`
}

// Return next incoming pending enroll message.
// Pending enroll messages are posted by the enroll service.
// This is a mechanism to de-couple enroll service so that
// it can scale better and return enroll pending messages to
// clients as quick as possible without getting into the rest
// of the enroll flow.
func GetPendingEnrollMessage() (*EnrollPayload, error) {
	msgs, err := receiveMessage(pendingEnrollQueueUrl)
	if err != nil {
		eswLogger.Error("Error receiving pending enroll message",
			zap.Error(err))
		return nil, err
	}

	var payload EnrollPayload
	if len(msgs) == 0 {
		return nil, nil
	}
	if err = json.Unmarshal([]byte(*msgs[0].Body), &payload); err != nil {
		eswLogger.Error("Error unmarshalling pending enroll message",
			zap.Error(err))
		return nil, err
	}
	payload.ReceiptHandle = *msgs[0].ReceiptHandle
	return &payload, nil
}

// A processed pending enroll message is removed from queue here.
func DeletePendingEnrollMessage(receiptHandle string) error {
	err := deleteMessage(pendingEnrollQueueUrl, receiptHandle)
	if err != nil {
		eswLogger.Error(
			"could not delete pending enroll message",
			zap.String("receipthandle", receiptHandle),
			zap.Error(err))
	}
	return err
}
