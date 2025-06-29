// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package cache

import (
	"testing"

	"github.com/google/uuid"
)

// set csrhash and check if exists
// expect success
func TestSuccessfulCsrHashCheck(t *testing.T) {
	id := uuid.New().String()
	SetCsrHash(id)
	exists, err := HasCsrHash(id)
	if err != nil {
		t.Errorf("Cache get csr hash failed. Expected no error. Got %v", err)
	}
	if !exists {
		t.Errorf("Cache check csr hash. Expected true. Got false")
	}
}

// try to check a csrhash that is not set
// expect false on exists and cache not found error
func TestNonExistentCsrHashFails(t *testing.T) {
	id := uuid.New().String()
	exists, err := HasCsrHash(id)
	if err == nil {
		t.Errorf("Cache check csrhash when not set. Expected error. Got none")
	}
	if exists {
		t.Errorf("Cache check csrhash when not set. Expected false. Got true")
	}
}
