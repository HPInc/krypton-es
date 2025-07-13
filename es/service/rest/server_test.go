// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/HPInc/krypton-es/es/service/config"
	"github.com/HPInc/krypton-es/es/service/db"
	"github.com/HPInc/krypton-es/es/service/policy"
	"github.com/HPInc/krypton-es/es/service/tokenmgr"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type testTokenInfo struct {
	tenantId string
	deviceId string
	userId   string
}

var (
	log          *zap.Logger
	serverConfig config.Server = config.Server{
		Host: getServer(),
		Port: 7979,
	}
	testApiPath    = "api/v1"
	tokenUrl       = fmt.Sprintf("%s/%s/token", getJwtServer(), testApiPath)
	deviceTokenUrl = fmt.Sprintf("%s/%s/device_token", getJwtServer(), testApiPath)
)

func getServer() string {
	e := os.Getenv("ES_SERVER")
	if e != "" {
		return e
	}
	return "localhost"
}

func getJwtServer() string {
	e := os.Getenv("ES_TEST_JWT_SERVER")
	if e != "" {
		return e
	}
	return "http://localhost:9090"
}

func TestMain(m *testing.M) {
	log, _ = zap.NewProduction(zap.AddCaller())
	os.Setenv("ES_DB_SCHEMA_MIGRATION_SCRIPTS", "../db/schema")
	db.InitTestDefault(log)
	tokenmgr.Init(log, "../config/token_config_test.yaml")
	policy.Init(log, "../config/default_policy.json")
	Init(log, &serverConfig)
	defer Shutdown()
	os.Exit(m.Run())
}

func executeTestRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return rr
}

func checkTestResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func getBearerToken() string {
	invalid := "Bearer invalid"
	r, err := http.Get(tokenUrl)
	if err != nil {
		log.Error("Error getting test jwt token",
			zap.Error(err))
		return invalid
	}
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("Failed to get jwt token response",
			zap.Error(err))
		return invalid
	}
	return "Bearer " + string(data)
}

func getDeviceToken() (*testTokenInfo, string) {
	tenantId := uuid.New().String()
	deviceId := uuid.New().String()
	return getDeviceTokenWithParams(tenantId, deviceId)
}

func getDeviceTokenWithParams(tenantId, deviceId string) (*testTokenInfo, string) {
	invalid := "Bearer invalid"
	info := &testTokenInfo{
		tenantId: tenantId,
		deviceId: deviceId,
	}
	req, _ := http.NewRequest(http.MethodGet, deviceTokenUrl, nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	data := url.Values{}
	data.Set("tenant_id", tenantId)
	data.Set("device_id", deviceId)
	req.URL.RawQuery = data.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Failed to execute device token req",
			zap.Error(err))
		return info, invalid
	}

	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to get test device token response",
			zap.Error(err))
		return info, invalid
	}
	return info, "Bearer " + string(bytes)
}

func handleError(t *testing.T, err error) {
	if err == nil {
		return
	}
	t.Errorf("Expected no error. Got %v", err)
}
