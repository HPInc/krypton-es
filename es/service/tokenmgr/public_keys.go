// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

// get and set public keys through an in memory cache
package tokenmgr

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/HPInc/krypton-es/es/service/db"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

var (
	publicKeys = make(map[string]interface{})
)

// get saved public key for kid
// first check in cache, return if found
// if not found in cache, check in db
// - if found in db, add to cache, return
// - if not found in db, return error
func getPublicKey(kid string) (interface{}, error) {
	var pubkey interface{}
	var ok bool
	var err error
	if pubkey, ok = publicKeys[kid]; !ok {
		if pubkey, err = makePublicKeyWithDbData(kid); err != nil {
			esLogger.Error("Could not find public key",
				zap.String("kid", kid),
				zap.Error(err))
			return nil, fmt.Errorf(
				"No public key to validate kid: %s", kid)
		}
		publicKeys[kid] = pubkey
	}
	return pubkey, nil
}

// set public key in db
// do not add to cache at this time as there is no
// guarantee for use. better to add to cache on first use
func setPublicKey(kid, keyType, keyString string) error {
	return db.AddKey(kid, keyType, keyString)
}

// helper method
// look up in db by kid, return parsed public key
func makePublicKeyWithDbData(kid string) (*rsa.PublicKey, error) {
	keystring, err := db.GetPublicKey(kid)
	if err != nil {
		return nil, err
	}
	return makePublicKey(keystring)
}

// helper method
// make public key from string data
func makePublicKey(keystring string) (*rsa.PublicKey, error) {
	// parse pem bytes to make public key
	block, _ := pem.Decode([]byte(keystring))
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		esLogger.Error("Failed to decode PEM block")
		if block != nil {
			esLogger.Error("Block type not supported.",
				zap.String("block_type", block.Type))
		}
		return nil, errors.New("Invalid pem block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		esLogger.Error("Error parsing public key", zap.Error(err))
		return nil, err
	}
	return pub.(*rsa.PublicKey), nil
}

// get the kid from jwt and use it to fetch public key
// from in memory cache
func getPublicKeyForJwt(token *jwt.Token) (interface{}, error) {
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, ErrInvalidTokenHeaderKid
	}
	return getPublicKey(kid)
}
