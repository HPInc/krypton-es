// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

const (
	headerPolicyId = "x-hp-policy-id"
)

type policyExistsResponse struct {
	Id string `json:""`
}

func sendPolicyExistsResponse(w http.ResponseWriter, policyId string) {
	p := policyExistsResponse{
		Id: policyId,
	}
	if err := sendJsonResponse(w, http.StatusConflict, &p); err != nil {
		esLogger.Error("Failed to send json response",
			zap.Error(err))
	}
}

func sendInternalServerErrorResponse(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
}

func sendUnsupportedMediaTypeResponse(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusUnsupportedMediaType),
		http.StatusUnsupportedMediaType)
}

// JSON encode and send the specified payload & the specified HTTP status code.

// JSON encode and send the specified payload & the specified HTTP status code.
func sendJsonResponse(w http.ResponseWriter, statusCode int,
	payload interface{}) error {
	w.Header().Set(headerContentType, contentTypeJson)
	w.WriteHeader(statusCode)

	if payload != nil {
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		err := encoder.Encode(payload)
		if err != nil {
			esLogger.Error("Failed to encode JSON response!",
				zap.Error(err),
			)
			sendInternalServerErrorResponse(w)
			return err
		}
	}

	return nil
}

// modeling after http.Error source
// https://go.dev/src/net/http/server.go?s=61907:61959#L2131
func sendJsonError(w http.ResponseWriter, jsonError *JsonError) {
	w.Header().Set(headerContentType, contentTypeJsonUtf8)
	w.Header().Set(headerContentTypeOptions, contentTypeOptionNoSniff)

	w.WriteHeader(jsonError.Code)
	if err := json.NewEncoder(w).Encode(jsonError); err != nil {
		esLogger.Error("Error encoding response to json", zap.Error(err))
	}
}
