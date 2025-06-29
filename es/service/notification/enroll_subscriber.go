// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

import (
	"encoding/json"

	"github.com/HPInc/krypton-es/es/service/db"
	"github.com/HPInc/krypton-es/es/service/metrics"
	"github.com/HPInc/krypton-es/es/service/structs"
	"go.uber.org/zap"
)

// this struct will hold a type
// and the appropriate enroll/unenroll data
type EnrollEnvelope struct {
	ReceiptHandle  string                 `json:"receipt_handle,omitempty"`
	Type           string                 `json:"type"`
	EnrollResult   structs.EnrollResult   `json:"-"`
	UnenrollResult structs.UnenrollResult `json:"-"`
}

func watchEnrollQueue() {
	for {
		if gCtx.Err() != nil {
			esLogger.Info(
				"Shutting down enroll queue watch.")
			break
		}
		dc, err := getEnrolledMessage()
		if err != nil {
			esLogger.Error(
				"Error fetching enrolled messages: ",
				zap.Error(err))
			metrics.MetricNotificationEnrollParseErrors.Inc()
		}
		// there are no messages if dc is nil
		if dc != nil {
			// unenroll needs to be routed to unenroll path
			if dc.Type == payloadTypeUnenroll {
				processUnenrolled(dc)
			} else {
				processEnrolled(dc)
			}
		}
	}
}

// enrolled message represents a successful enroll
// and contains a device certificate.
func getEnrolledMessage() (*EnrollEnvelope, error) {
	var result EnrollEnvelope
	msgs, err := receiveMessage(
		enrollQueueUrl,
		notificationSettings.EnrollWatchDelay)
	if err != nil {
		esLogger.Error("Error receiving enrolled message",
			zap.Error(err))
		return nil, err
	}
	if len(msgs) == 0 {
		return nil, nil
	}
	if err = json.Unmarshal([]byte(*msgs[0].Body), &result); err != nil {
		return nil, err
	}
	// do a specialized unmarshall based on type
	if result.Type == payloadTypeUnenroll {
		if err = json.Unmarshal([]byte(*msgs[0].Body), &result.UnenrollResult); err != nil {
			return nil, err
		}
	} else {
		if err = json.Unmarshal([]byte(*msgs[0].Body), &result.EnrollResult); err != nil {
			return nil, err
		}
	}
	result.ReceiptHandle = *msgs[0].ReceiptHandle
	return &result, nil
}

// process enrolled message by updating the enroll record
// as successful.
func processEnrolled(ee *EnrollEnvelope) {
	err := db.UpdateEnrollRecord(&ee.EnrollResult)
	if err != nil {
		esLogger.Error("could not update enroll record", zap.Error(err))
		metrics.MetricNotificationEnrollsFailed.Inc()
	} else {
		metrics.MetricNotificationEnrolls.Inc()
	}
	deleteEnrolledQueueEntry(ee)
}

// once enrolled message is processed, remove the queue entry
// if not removed, notification listeners will pick up duplicate
// entries
func deleteEnrolledQueueEntry(ee *EnrollEnvelope) {
	err := deleteMessage(enrollQueueUrl, ee.ReceiptHandle)
	if err != nil {
		esLogger.Error("could not delete queue entry after process",
			zap.Error(err))
		metrics.MetricNotificationEnrollDeleteErrors.Inc()
	} else {
		esLogger.Info("enrolled:",
			zap.String("enrollId:", ee.EnrollResult.EnrollId.String()),
			zap.String("deviceId:", ee.EnrollResult.DeviceId.String()))
	}
}
