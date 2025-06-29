// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

import (
	"github.com/HPInc/krypton-es/es/service/db"
	"github.com/HPInc/krypton-es/es/service/metrics"
	"go.uber.org/zap"
)

const (
	payloadTypeUnenroll = "unenroll"
)

// process unenrolled message by updating the enroll record
// as successful.
func processUnenrolled(ee *EnrollEnvelope) {
	r := &ee.UnenrollResult
	err := db.UpdateUnenrollRecord(r)
	if err != nil {
		esLogger.Error("could not update unenroll record",
			zap.Error(err),
			zap.String("unenroll_id", r.UnenrollId.String()),
			zap.String("device_id", r.DeviceId),
			zap.String("request_id", r.RequestId),
		)
		metrics.MetricNotificationUnenrollsFailed.Inc()
	} else {
		metrics.MetricNotificationUnenrolls.Inc()
	}
	deleteUnenrolledQueueEntry(ee)
}

// once unenrolled message is processed, remove the queue entry
// if not removed, notification listeners will pick up duplicate
// entries
func deleteUnenrolledQueueEntry(ee *EnrollEnvelope) {
	r := ee.UnenrollResult
	err := deleteMessage(enrollQueueUrl, ee.ReceiptHandle)
	if err != nil {
		esLogger.Error("could not delete unenroll queue entry after process",
			zap.String("unenroll_id", r.UnenrollId.String()),
			zap.String("device_id", r.DeviceId),
			zap.String("request_id", r.RequestId),
			zap.Error(err))
		metrics.MetricNotificationUnenrollDeleteErrors.Inc()
	} else {
		esLogger.Info("unenrolled:",
			zap.String("unenroll_id", r.UnenrollId.String()),
			zap.String("device_id", r.DeviceId),
			zap.String("request_id", r.RequestId))
	}
}
