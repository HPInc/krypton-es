// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package tokenmgr

import "errors"

var (
	ErrTokenConfigurationInitFailure = errors.New("failed to initialize token configuration")
	ErrMissingKeySources             = errors.New("no JWKs sources found in token configuration")
	ErrKIDNotFound                   = errors.New("the given key ID was not found in the JWKS")
	ErrMissingAssets                 = errors.New("required assets are missing to create a public key")
	ErrValidatorNotImplemented       = errors.New("the requested token validator is not implemented")
	ErrUnsupportedTokenType          = errors.New("X-HP-Token-Type header contains an unsupported token type")
	ErrInvalidToken                  = errors.New("invalid token provided")
	ErrInvalidTokenHeaderKid         = errors.New("invalid token signing kid specified")
	ErrInvalidTokenHeaderSigningAlg  = errors.New("invalid token signing algorithm specified")
	ErrInvalidIssuerClaim            = errors.New("specified token contains an invalid issuer claim")
	ErrInvalidAudienceClaim          = errors.New("specified token contains an invalid audience claim")
	ErrInvalidTypeClaim              = errors.New("specified token contains an invalid typ claim")
	ErrInvalidSubjectClaim           = errors.New("specified token contains an invalid sub claim")
)
