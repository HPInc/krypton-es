// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

import (
	"encoding/json"

	caclient "github.com/HPInc/krypton-es/es-worker/service/client/ca"
	"go.uber.org/zap"
)

/*
Pending registration queue is an internal de-coupling queue for es-worker.
es-worker posts to this queue once a ca certificate is obtained from ca
This function retreives the next pending registration message.
*/
func GetPendingRegistrationMessage() (*caclient.DeviceCertificate, error) {
	msgs, err := receiveMessage(pendingRegistrationQueueUrl)
	if err != nil {
		eswLogger.Error("Error recieving pending registration message",
			zap.Error(err))
		return nil, err
	}

	var payload caclient.DeviceCertificate
	if len(msgs) == 0 {
		return nil, nil
	}
	if err = json.Unmarshal([]byte(*msgs[0].Body), &payload); err != nil {
		eswLogger.Error("Error unmarshaling pending registration message",
			zap.Error(err))
		return nil, err
	}
	payload.ReceiptHandle = *msgs[0].ReceiptHandle
	return &payload, nil
}

/*
Internal message to pending registration queue. This is the second
stage of enroll worker processing pending enroll messages.
This message will serve as a de-coupling point between ca and dsts
processing.
*/
func SendPendingRegistrationMessage(dc *caclient.DeviceCertificate) error {
	jsonstring, err := json.Marshal(dc)
	if err != nil {
		eswLogger.Error("Error sending pending registration message",
			zap.Error(err))
		return err
	}
	return sendMessage(pendingRegistrationQueueUrl, string(jsonstring))
}

/*
delete a pending registration message once it is processed or
the process resulted in an error.
*/
func DeletePendingRegistrationMessage(receiptHandle string) error {
	err := deleteMessage(pendingRegistrationQueueUrl, receiptHandle)
	if err != nil {
		eswLogger.Error(
			"Error deleting pending registration message",
			zap.String("receipthandle: ", receiptHandle),
			zap.Error(err))
	}
	return err
}
