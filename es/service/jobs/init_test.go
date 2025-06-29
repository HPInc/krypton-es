// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package jobs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/HPInc/krypton-es/es/service/config"
	"github.com/HPInc/krypton-es/es/service/db"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	log, _ := zap.NewProduction(zap.AddCaller())

	if os.Getenv("ES_SKIP_JOBS_TEST") == "true" {
		log.Warn("Skipping jobs test",
			zap.String("es_skip_jobs_test", "true"))
		return
	}
	os.Setenv("ES_DB_SCHEMA_MIGRATION_SCRIPTS", "../db/schema")
	db.InitTestDefault(log)

	// set up job to run in 2 seconds
	d, _ := time.ParseDuration("2s")
	start := time.Now().Add(d)
	t := fmt.Sprintf("%02d:%02d:%02d",
		start.Hour(), start.Minute(), start.Second())
	var jobsConfig config.ScheduledJobs
	jobsConfig = map[string]config.ScheduledJob{
		"delete_expired_enrolls": {
			Enabled: true, Start: t, Every: "1s",
		},
	}
	Init(log, &jobsConfig)
	defer Shutdown()
	os.Exit(m.Run())
}

// ping database test
func TestDatabasePing(t *testing.T) {
	if err := db.Ping(); err != nil {
		t.Errorf("Database ping failed. Expected no error. Got %v", err)
	}
}

func TestScheduledDelete(t *testing.T) {
	var i, enrollCount int = 0, 10
	for ; i < enrollCount; i++ {
		newEnroll()
	}
	time.Sleep(5 * time.Second)
	i, err := db.CountEnrollRecords()
	if err != nil {
		t.Errorf("Enroll count failed. Expected no error. Got %v", err)
	}
	if i > 0 {
		t.Errorf("Expected enroll count 0. Got %d", i)
	}
}

func handleError(t *testing.T, err error) {
	if err == nil {
		return
	}
	t.Errorf("Expected no error. Got %v", err)
}

func newEnroll() {
	userId := uuid.New().String()
	tenantId := uuid.New().String()
	csrHash := uuid.New().String()
	db.CreateEnrollRecord(userId, tenantId, csrHash)
}
