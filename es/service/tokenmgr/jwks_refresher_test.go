// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package tokenmgr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

// test if the timeout mechanism works for jwks refresh
func TestJwksRefreshTimeout(t *testing.T) {
	esLogger, _ = zap.NewProduction(zap.AddCaller())
	defer esLogger.Sync()

	gCtx = context.Background()
	defer gCtx.Done()

	// create a test server
	svr := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(timeoutJwksGet + 5)
			fmt.Fprintf(w, "{}")
		}))
	defer svr.Close()

	expected := context.DeadlineExceeded
	_, err := GetKeysFromServer(svr.URL)
	if err == nil || !errors.Is(err, expected) {
		t.Errorf("Expected error %v, Got %v\n", expected, err)
	}
}
