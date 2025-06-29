// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"testing"
)

func TestGetCSRHashFailsForNonBase64(t *testing.T) {
	ep := enrollPayload{CSR: "123"}
	_, err := ep.getCSRHash()
	if err == nil {
		t.Errorf("Expected error for non base64 encoded csr. Got none")
	}
}
