// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/HPInc/krypton-es/es/service/db"
)

var (
	ErrInvalidTokenType        = errors.New("an invalid token type was specified")
	ErrAppTokenNotProvided     = errors.New("app token expected but none was specified")
	ErrDeviceTokenNotProvided  = errors.New("device token expected but none was specified")
	ErrDeviceIdMismatch        = errors.New("device id does not match claim in bearer token")
	ErrDuplicateCsr            = errors.New("specified csr has been used previously")
	ErrNoAuthorizationHeader   = errors.New("request does not have an authorization header")
	ErrNoBearerTokenSpecified  = errors.New("authorization header does not contain a bearer token")
	ErrTokenTypeHeaderNotFound = errors.New("the X-HP-Token-Type header was not found in the request")
	ErrRequestInProgress       = errors.New("request is being processed. Please see 'Retry-After' for a wait hint")
	ErrTenantIdNotProvided     = errors.New("param tenant_id is not provided")
	ErrTenantIdMismatch        = errors.New("tenant id does not match claim in bearer token")
	ErrPayloadRead             = errors.New("payload read error")
	ErrPayloadMissing          = errors.New("payload missing")
	ErrInvalidPolicyVersion    = errors.New("invalid policy version")
	ErrLookupCsr               = errors.New("there was an error while looking up this csr")
	ErrCreateEnroll            = errors.New("there was an error creating enroll entry")
	ErrRenewEnroll             = errors.New("there was an error renewing enroll")
	ErrUnenroll                = errors.New("there was an error while unenroll")
	ErrHandoffEnroll           = errors.New("there was an error handing off enroll for processing")
	ErrInternal                = errors.New("server encountered an internal error")
	ErrCreateEnrollToken       = errors.New("could not create enroll token")
	ErrGetEnrollToken          = errors.New("could not get enroll token")
	ErrLookupEnroll            = errors.New("could not find enroll entry")
	ErrCreatePolicy            = errors.New("could not create policy")
	ErrDeletePolicy            = errors.New("could not delete policy")
	ErrGetPolicy               = errors.New("could not get policy")
	ErrUpdatePolicy            = errors.New("could not update policy")
	ErrInvalidPolicy           = errors.New("invalid policy data")
)

// translate db error to http code
func getHttpCodeForDbError(err error) int {
	// all db errors are unexpected. mapping to 500
	httpCode := http.StatusInternalServerError

	// if we have a deadline exceeded error, map to 429
	if errors.Is(err, context.DeadlineExceeded) {
		httpCode = http.StatusTooManyRequests
	} else if db.IsDbErrorNoRows(err) {
		httpCode = http.StatusNotFound
	}

	return httpCode
}
