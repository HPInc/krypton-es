// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT
// Purpose:
// HTTP handler for health

package rest

import (
	"net/http"
)

func getHealthHandler(w http.ResponseWriter, r *http.Request) {
	_ = sendJsonResponse(w, http.StatusOK, nil)
}
