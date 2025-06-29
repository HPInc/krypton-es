// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"testing"

	"github.com/google/uuid"
)

const (
	ktyRSA = "RSA"
)

func TestHasKeyReturnsTrueOnMatch(t *testing.T) {
	kid, _ := createAndAddPublicKey()
	ok, err := HasKey(kid)
	if err != nil {
		handleError(t, err)
	}
	if !ok {
		t.Errorf("Expected match for HasKey")
	}
}

func TestHasKeyReturnsFalseOnNoMatch(t *testing.T) {
	ok, err := HasKey(uuid.New().String())
	if err != nil {
		handleError(t, err)
	}
	if ok {
		t.Errorf("Expected no match for HasKey")
	}
}

func TestGetPublicKeySucceedsOnMatch(t *testing.T) {
	kid, key := createAndAddPublicKey()
	keyFetched, err := GetPublicKey(kid)
	if err != nil {
		handleError(t, err)
	}
	if keyFetched == "" {
		t.Errorf("Expected key fetched: %s, Found empty string", key)
	} else if key != keyFetched {
		t.Errorf("Expected key fetched: %s, Found: %s", key, keyFetched)
	}
}

func TestGetPublicKeyFailsOnNoMatch(t *testing.T) {
	keyFetched, err := GetPublicKey(uuid.New().String())
	if err == nil {
		t.Errorf("Expected no rows error, Found nil")
	}
	if keyFetched != "" {
		t.Errorf("Expected key fetched to be empty string, Found: %s", keyFetched)
	}
}

func TestAddPublicKeySucceeds(t *testing.T) {
	kid := uuid.New().String()
	if err := addPublicKey(kid, ktyRSA, uuid.New().String()); err != nil {
		handleError(t, err)
	}
}

func TestAddPublicKeyFailsOnDuplicateKey(t *testing.T) {
	kid, _ := createAndAddPublicKey()
	if err := addPublicKey(kid, ktyRSA, uuid.New().String()); err == nil {
		t.Errorf("Expected duplicate key error. Found no error")
	}
}

// util function addPublicKey
func createAndAddPublicKey() (string, string) {
	kid := uuid.New().String()
	key := uuid.New().String()
	addPublicKey(kid, ktyRSA, key)
	return kid, key
}
