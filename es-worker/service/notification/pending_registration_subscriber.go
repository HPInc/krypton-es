// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

import (
	caclient "github.com/HPInc/krypton-es/es-worker/service/client/ca"
	dstsclient "github.com/HPInc/krypton-es/es-worker/service/client/dsts"
	"go.uber.org/zap"
)

func ProcessPendingRegistrationQueue() {
	for {
		if gCtx.Err() != nil {
			eswLogger.Info(
				"Shutting down pending registration queue watch.")
			break
		}
		dc, err := GetPendingRegistrationMessage()
		if err != nil {
			eswLogger.Error("Failed fetching registration messages.",
				zap.Error(err))
		}
		// there are no messages if dc is nil
		if dc != nil {
			processPendingRegistration(dc)
		}
	}
}

// process and remove from queue
func processPendingRegistration(dc *caclient.DeviceCertificate) {
	var err error
	if err = processPendingSingleRegistration(dc); err != nil {
		if err = SendErrorMessage(&ErrorMessage{
			EnrollId:  dc.EnrollId,
			Message:   err.Error(),
			Type:      dc.Type,
			RequestId: dc.RequestId,
		}); err != nil {
			eswLogger.Error("Failed fetching registration messages.",
				zap.Error(err))
			// no need to retry. proceed to delete
		}
	}
	if err = DeletePendingRegistrationMessage(dc.ReceiptHandle); err != nil {
		eswLogger.Error(
			"Removing pending registration message failed.",
			zap.Error(err))
	}
}

func processPendingSingleRegistration(dc *caclient.DeviceCertificate) error {
	var err error
	eswLogger.Info("Processing pending registration",
		zap.String("enroll_id", dc.EnrollId),
		zap.String("tenant_id", dc.TenantId),
		zap.String("type", dc.Type),
		zap.String("service", dc.ManagementService))
	if dc.Type == caclient.CertificateTypeEnroll {
		if err = dstsclient.CreateDevice(gDSTSClient, dc); err != nil {
			eswLogger.Error(
				"Create device failed.", zap.Error(err))
			return err
		}
	} else if dc.Type == caclient.CertificateTypeRenew {
		if err = dstsclient.UpdateDevice(gDSTSClient, dc); err != nil {
			eswLogger.Error(
				"Update device failed.", zap.Error(err))
			return err
		}
	} else {
		eswLogger.Error(
			"Invalid device certificate type",
			zap.String("type", dc.Type))
		return err
	}
	if err = SendEnrolledMessage(dc); err != nil {
		eswLogger.Error("Sending enrolled message failed.", zap.Error(err))
		return err
	}
	eswLogger.Info("Enrolled message sent.",
		zap.String("enroll_id", dc.EnrollId),
		zap.String("tenant_id", dc.TenantId),
		zap.String("device_id", dc.DeviceId),
		zap.String("type", dc.Type),
		zap.String("service", dc.ManagementService))
	return nil
}
