// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

import (
	"encoding/json"

	dstsclient "github.com/HPInc/krypton-es/es-worker/service/client/dsts"
	"go.uber.org/zap"
)

// send delete device completion message
// note that this is re-using enroll complete queue
func SendDeleteDeviceComplete(msg *dstsclient.DeviceDetails) error {
	jsonString, err := json.Marshal(msg)
	if err != nil {
		eswLogger.Error("Error sending delete device message",
			zap.String("error_message:", string(jsonString)),
			zap.Error(err))
		return err
	}
	eswLogger.Info("Sending delete device complete message",
		zap.String("message", string(jsonString)))
	return sendMessage(enrolledQueueUrl, string(jsonString))
}
