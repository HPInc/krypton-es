// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	// REST request headers and expected header values.
	headerContentType        = "Content-Type"
	headerContentTypeOptions = "X-Content-Type-Options"
	headerRequestID          = "request_id"
	headerRetryAfter         = "Retry-After"
	headerAuthorization      = "Authorization"
	headerTokenType          = "X-HP-Token-Type" //#nosec G101
	bearerToken              = "Bearer "

	contentTypeFormUrlEncoded = "application/x-www-form-urlencoded"
	contentTypeJson           = "application/json"
	contentTypeJsonUtf8       = "application/json; charset=utf-8"
	contentTypeOptionNoSniff  = "nosniff"

	// Request parameters
	paramTenantID = "tenant_id"
	paramDeviceID = "device_id"
	paramEnrollID = "enroll_id"
	paramPolicyId = "policy_id"

	requestPayloadTypeEnroll   = "enroll"
	requestPayloadTypeReenroll = "renew_enroll"
	requestPayloadTypeUnenroll = "unenroll"
)

func getUUIDParam(r *http.Request, name string) (uuid.UUID, *enrollError) {
	var err error
	vars := mux.Vars(r)
	idstring, ok := vars[name]
	if !ok {
		err = fmt.Errorf("require %s in path", name)
		return uuid.Nil, &enrollError{err, http.StatusBadRequest}
	}
	id, err := uuid.Parse(idstring)
	if err != nil {
		err = fmt.Errorf("%s must be a uuid", name)
		return uuid.Nil, &enrollError{err, http.StatusBadRequest}
	}
	return id, nil
}
