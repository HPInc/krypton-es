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

func watchEnrollErrorQueue() {
	for {
		if gCtx.Err() != nil {
			esLogger.Info(
				"Shutting down enroll error queue watch.")
			break
		}
		ee, err := getEnrollErrorMessage()
		if err != nil {
			esLogger.Error(
				"Error fetching enroll error messages: ",
				zap.Error(err))
			metrics.MetricNotificationEnrollErrorParseErrors.Inc()
		}
		// there are no messages if dc is nil
		if ee != nil {
			processEnrollError(ee)
		}
	}
}

// enroll error message represents an issue in enroll processing
// this will end up marking the enroll request as failed.
func getEnrollErrorMessage() (*structs.EnrollError, error) {
	var ee structs.EnrollError

	msgs, err := receiveMessage(
		enrollErrorQueueUrl,
		notificationSettings.EnrollErrorWatchDelay)
	if err != nil {
		esLogger.Error("Error receiving enroll error message",
			zap.Error(err))
		return nil, err
	}
	if len(msgs) == 0 {
		return nil, nil
	}
	if err = json.Unmarshal([]byte(*msgs[0].Body), &ee); err != nil {
		return nil, err
	}
	ee.ReceiptHandle = *msgs[0].ReceiptHandle
	return &ee, nil
}

func processEnrollError(ee *structs.EnrollError) {
	if ee.Type == "enroll" || ee.Type == "renew_enroll" {
		failEnrollRecord(ee)
	} else if ee.Type == "unenroll" {
		failUnenrollRecord(ee)
	} else {
		esLogger.Error("Unknown type in enroll error",
			zap.String("type", ee.Type))
	}
	deleteErrorQueueEntry(ee)
}

// fail enroll record by recording error in db
func failEnrollRecord(ee *structs.EnrollError) {
	err := db.FailEnrollRecord(ee)
	if err != nil {
		esLogger.Error("Failed updating enroll failure",
			zap.Error(err))
		metrics.MetricNotificationEnrollErrorsFailed.Inc()
	} else {
		metrics.MetricNotificationEnrollErrors.Inc()
	}
}

// fail enroll record by recording error in db
func failUnenrollRecord(ee *structs.EnrollError) {
	err := db.FailUnenrollRecord(ee)
	if err != nil {
		esLogger.Error("Failed updating unenroll failure",
			zap.Error(err))
		metrics.MetricNotificationUnenrollErrorsFailed.Inc()
	} else {
		metrics.MetricNotificationUnenrollErrors.Inc()
	}
}

// remove error queue entry. this should happen regardless of process status
func deleteErrorQueueEntry(ee *structs.EnrollError) {
	err := deleteMessage(enrollErrorQueueUrl, ee.ReceiptHandle)
	if err != nil {
		esLogger.Error("could not delete error queue entry after process",
			zap.Error(err))
		metrics.MetricNotificationEnrollErrorDeleteErrors.Inc()
	} else {
		esLogger.Info("enroll error:",
			zap.String("enrollId:", ee.EnrollId),
			zap.Int("code:", ee.ErrorCode),
			zap.String("error:", ee.ErrorMessage))
	}
}
