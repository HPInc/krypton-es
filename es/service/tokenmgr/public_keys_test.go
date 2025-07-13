// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package tokenmgr

import (
	"testing"

	"go.uber.org/zap"
)

const (
	kidInvalid = "invalid"

	keyValid = `-----BEGIN RSA PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4P/Skv0UkIY6bfuj9U48
1x12Td+arbHco85yp+7dJ2speOMfJFNKNPiD0vTDoMr5KTJ0rdGlnfKUihHoRd+s
BpvF+N62vc2vZ8EvITeKp8Z/SGSgVEnoRI1wEPnw1EWX06IgeHzYV58rRKPyiIs/
ZgAf1yy/X0sMfUqKhK3uk9j6bSF9528kCUuGvEpXIHI0gGRG26GQALOjKfSYQu2H
MVEA4uCq64TC6NpwfhGIU/FvHMwsEZp7En3llalzoDmABpGInJB3TnDCzDtaYpg/
WFFLOWDLSCAPIClLONZgtzZBgvB+YI/3QdVwgfVvnsPM5Jt2i91l4vKGRLsPfGK/
JQIDAQAB
-----END RSA PUBLIC KEY-----`
	keyInvalid = "invalid"
)

// test string to key conversion with an invalid key string
// expects failure with parse
func TestMakePublicKeyWithInvalidString(t *testing.T) {
	esLogger, _ = zap.NewProduction(zap.AddCaller())
	defer esLogger.Sync()

	pubkey, err := makePublicKey(keyInvalid)
	if err == nil {
		t.Errorf("Expected error, Got %v\n", err)
	}
	if pubkey != nil {
		t.Errorf("Expected nil public key, Got %v\n", pubkey)
	}
}

// test string to key conversion with an valid key string
// expects parse success
func TestMakePublicKeyWithValidString(t *testing.T) {
	esLogger, _ = zap.NewProduction(zap.AddCaller())
	defer esLogger.Sync()

	pubkey, err := makePublicKey(keyValid)
	if err != nil {
		t.Errorf("Expected no error, Got %v\n", err)
	}
	if pubkey == nil {
		t.Errorf("Expected valid public key, Got nil\n")
	}
}
