// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/HPInc/krypton-es/es/service/db"
	"github.com/HPInc/krypton-es/es/service/structs"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func getUnenrollStatus(w http.ResponseWriter, id uuid.UUID, ei *EnrollInfo) *enrollError {
	var entry *structs.UnenrollStatus
	var err error
	entry, err = db.GetUnenrollStatus(id)
	// if there is a lookup error or if entry does not match token details, dont go further.
	if err != nil {
		esLogger.Error("Could not find unenroll id",
			zap.Error(err))
		return &enrollError{
			fmt.Errorf("id: %s is not found", id),
			http.StatusNotFound,
		}
	}
	if entry.TenantId != ei.TenantId || entry.DeviceId != ei.DeviceId {
		esLogger.Error("Could not find unenroll id",
			zap.String("token_tenant_id", ei.TenantId),
			zap.String("token_device_id", ei.DeviceId),
			zap.String("tenant_id", entry.TenantId),
			zap.String("device_id", entry.DeviceId))
		return &enrollError{
			fmt.Errorf("id: %s is not found", id),
			http.StatusNotFound,
		}
	}
	switch entry.Status {
	case 0:
		writeRetryAfter(w, getRetryAfterHint())
		return &enrollError{
			ErrRequestInProgress,
			http.StatusTooManyRequests,
		}
	case 1:
		return getCompletedUnenroll(w, entry, id)
	default:
		return &enrollError{
			fmt.Errorf("id: %s is not found", id),
			http.StatusNotFound,
		}
	}
}

func getCompletedUnenroll(
	w http.ResponseWriter,
	entry *structs.UnenrollStatus,
	id uuid.UUID) *enrollError {
	type unenrollStatus struct {
		Id     uuid.UUID `json:"id"`
		Status string    `json:"status"`
	}
	jsonstring, err := json.Marshal(&unenrollStatus{
		Id:     id,
		Status: "success",
	})
	if err != nil {
		return &enrollError{err, http.StatusBadRequest}
	} else {
		fmt.Fprintf(w, "%s", jsonstring)
	}
	return nil
}
