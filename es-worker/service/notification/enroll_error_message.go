// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

import (
	"encoding/json"

	"go.uber.org/zap"
)

// error message to be posted in error queue
type ErrorMessage struct {
	// enroll id for es to look up record
	// this is present in pending_enroll messages
	// and is propagated through all intermediate
	EnrollId string `json:"enroll_id"`
	// error code
	Code int `json:"error_code"`
	// error message
	Message string `json:"error_message"`
	// type of the event which caused error
	Type string `json:"type"`
	// Request id of the original messsage
	RequestId string `json:"request_id"`
}

// send messages to error queue
func SendErrorMessage(msg *ErrorMessage) error {
	jsonString, err := json.Marshal(msg)
	if err != nil {
		eswLogger.Error("Error sending enroll error message",
			zap.String("error_message:", string(jsonString)),
			zap.Error(err))
		return err
	}
	eswLogger.Info("Sending error message",
		zap.String("message", string(jsonString)))
	return sendMessage(enrollErrorQueueUrl, string(jsonString))
}
