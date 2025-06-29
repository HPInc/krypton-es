// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

import (
	"encoding/base64"
	"fmt"

	caclient "github.com/HPInc/krypton-es/es-worker/service/client/ca"
	dstsclient "github.com/HPInc/krypton-es/es-worker/service/client/dsts"
	"go.uber.org/zap"
)

// this function is called for each configured timer tick.
// once called, this function will continue to process all pending
// messages before returning.
// if queue is empty at the time of invocation, function will
// immediately return.
func ProcessPendingEnrollQueue() {
	for {
		if gCtx.Err() != nil {
			eswLogger.Info(
				"Shutting down pending enroll queue watch.")
			break
		}
		en, err := GetPendingEnrollMessage()
		if err != nil {
			eswLogger.Error("Error fetching enrolled messages: ",
				zap.Error(err))
		}
		// there are no messages if en is nil.
		if en != nil {
			processPendingEnroll(en)
		}
	}
}

// process pending enroll
func processPendingEnroll(en *EnrollPayload) {
	var err error
	// if there is an error in processing, the corresponding
	// message is sent to enroll_error queue. es consumes the
	// errors and marks the enroll as failed.
	if err = processSinglePendingEnroll(en); err != nil {
		if err = SendErrorMessage(&ErrorMessage{
			EnrollId:  en.ID,
			Message:   err.Error(),
			Type:      en.Type,
			RequestId: en.RequestId,
		}); err != nil {
			eswLogger.Error("Error sending message",
				zap.Error(err))
			// no need to retry. proceed to delete
		}
	}
	if err = DeletePendingEnrollMessage(en.ReceiptHandle); err != nil {
		eswLogger.Error(
			"Removing pending registration message failed:",
			zap.Error(err))
	}
}

// pending enroll processing.
// payload contains csr and tenant information among other book keeping bits
// like enroll id.
// validates csr for base64 encoding, then forwards to ca for the appropriate
// create or renew based on enroll type. if there is an error in processing at
// any stage, an error message is posted to error queue.
// es will watch for errors and will immediately get to respond to any errors.
// if there are no errors, es will not get an immediate result as the
// message will be forwarded to pending_registration queue.
func processSinglePendingEnroll(en *EnrollPayload) error {
	eswLogger.Info("Pending enroll message",
		zap.String("type", en.Type),
		zap.String("request_id", en.RequestId),
		zap.String("enroll_id", en.ID),
		zap.String("tenant_id", en.TenantId))
	// if message type is unenroll, take a separate path
	if en.Type == EnrollTypeUnenroll {
		return DeleteDevice(en)
	}

	// we expect enroll or renew_enroll after this
	// these messages have same payload so we can do common
	// processing
	csr, err := base64.StdEncoding.DecodeString(en.CSR)
	if err != nil {
		eswLogger.Error("csr is not base64 encoded", zap.Error(err))
		return err
	}
	var dc *caclient.DeviceCertificate
	if en.Type == EnrollTypeEnroll {
		dc, err = caclient.CreateDeviceCertificate(
			gCAClient, en.RequestId, en.TenantId, csr)
		if err != nil {
			eswLogger.Error("Create device certificate failed:",
				zap.Error(err))
			return err
		}
	} else if en.Type == EnrollTypeRenew {
		dc, err = caclient.RenewDeviceCertificate(
			gCAClient, en.RequestId, en.TenantId, en.DeviceId, csr)
		if err != nil {
			eswLogger.Error("Renew device certificate failed:",
				zap.Error(err))
			return err
		}
	} else {
		eswLogger.Error("Invalid enroll type",
			zap.String("type", en.Type))
		return fmt.Errorf(
			"Invalid enroll type: %s. Expected enroll / renew_enroll",
			en.Type)
	}
	eswLogger.Info("Created device certificate",
		zap.String("type", en.Type),
		zap.String("enroll_id", en.ID),
		zap.String("tenant_id", en.TenantId))
	//inject the enroll id
	dc.EnrollId = en.ID
	dc.TenantId = en.TenantId
	dc.ManagementService = en.ManagementService
	dc.HardwareHash = en.HardwareHash
	err = SendPendingRegistrationMessage(dc)
	if err != nil {
		eswLogger.Error("Send pending registration message failed:",
			zap.Error(err))
		return err
	}
	eswLogger.Info("Pending registration message sent",
		zap.String("enroll_id", en.ID),
		zap.String("tenant_id", en.TenantId),
		zap.String("service", en.ManagementService))
	return nil
}

// trigger a delete in dsts
func DeleteDevice(en *EnrollPayload) error {
	eswLogger.Info("Processing delete device",
		zap.String("device_id", en.DeviceId))

	dd := &dstsclient.DeviceDetails{
		UnenrollId: en.ID,
		RequestId:  en.RequestId,
		TenantId:   en.TenantId,
		DeviceId:   en.DeviceId,
		Type:       en.Type,
	}

	// ask dsts to delete device
	err := dstsclient.DeleteDevice(gDSTSClient, dd)
	if err != nil {
		eswLogger.Error("Delete device failed:", zap.Error(err))
		return err
	}

	// send delete completion message
	return SendDeleteDeviceComplete(dd)
}
