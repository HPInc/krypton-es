// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package tokenmgr

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const (
	// timeout for jwks http calls
	timeoutJwksGet = time.Second * time.Duration(5)
)

// jsonWebKey represents a JSON Web Key inside a JWKS.
type jsonWebKey struct {
	Curve    string `json:"crv"`
	Exponent string `json:"e"`
	K        string `json:"k"`
	ID       string `json:"kid"`
	Modulus  string `json:"n"`
	Type     string `json:"kty"`
	Use      string `json:"use"`
	X        string `json:"x"`
	Y        string `json:"y"`
}

// rawJWKS represents a JWKS in JSON format.
type rawJWKS struct {
	Keys []*jsonWebKey `json:"keys"`
}

func startJwksRefresher() error {
	esLogger.Info("Starting JWKs refresher worker")
	refreshed := 0
	for k, v := range tokenConfig.TokenTypes {
		if v.KeysURL == "" {
			esLogger.Info("Skipping refresh",
				zap.String("name", string(k)),
				zap.String("reason", "empty keys url"))
			continue
		}
		esLogger.Info("Starting worker", zap.String("name:", string(k)))
		doRefresh(v)
		refreshed++
	}
	if refreshed == 0 {
		return ErrMissingKeySources
	}
	return nil
}

func doRefresh(keySource TokenIssuerSettings) {
	esLogger.Info("Keys source",
		zap.String("name:", keySource.KeysURL),
		zap.Int("refresh_interval:", keySource.RefreshInterval))
	ticker := time.NewTicker(time.Second * time.Duration(keySource.RefreshInterval))
	go func() {
		// do an immediate refresh
		processJWKS(keySource.KeysURL)
		for {
			select {
			case <-refreshKeysStopChannel:
				break
			case <-ticker.C:
				processJWKS(keySource.KeysURL)
			}
		}
	}()
}

func processJWKS(url string) {
	bytes, err := GetKeysFromServer(url)
	if err != nil {
		esLogger.Error("Error fetching keys.",
			zap.String("url:", url),
			zap.Error(err))
		return
	}
	if err = parseJWKS(bytes); err != nil {
		esLogger.Error("Error parsing keys.",
			zap.String("url:", url),
			zap.Error(err))
	}
}

func parseJWKS(jwksBytes json.RawMessage) (err error) {
	var rawKS rawJWKS
	err = json.Unmarshal(jwksBytes, &rawKS)
	if err != nil {
		esLogger.Error("Error unmarshalling jwks",
			zap.Error(err))
		return err
	}
	for _, key := range rawKS.Keys {
		switch keyType := key.Type; keyType {
		case ktyRSA:
			str, err := key.Pem()
			if err != nil {
				esLogger.Error("Error parsing rsa key",
					zap.String("type:", key.Type),
					zap.String("kid:", key.ID),
					zap.Error(err))
				continue
			}
			if err = setPublicKey(key.ID, key.Type, str); err != nil {
				return err
			}
		default:
			continue
		}
	}

	return nil
}

// jwks requests
// adds a default timeout for http calls
func GetKeysFromServer(url string) (keys []byte, err error) {
	ctx, cancel := context.WithTimeout(gCtx, timeoutJwksGet)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		esLogger.Error("Error creating request for keys",
			zap.String("url", url),
			zap.Error(err))
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		esLogger.Error("Error fetching keys",
			zap.String("url", url),
			zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		esLogger.Error("Get public keys failed",
			zap.String("url", url),
			zap.Int("status", resp.StatusCode),
			zap.Error(err))
		return nil, err
	}
	return io.ReadAll(resp.Body)
}
