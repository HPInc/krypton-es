// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"testing"
	"time"

	"github.com/HPInc/krypton-es/es/service/cache"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func getCachedEnroll(id uuid.UUID) func() bool {
	return func() bool {
		_, err := cache.GetEnrollStatus(id)
		return err != redis.Nil
	}
}

// create single enroll, then delete it
func TestDeleteEnrollById(t *testing.T) {
	er, err := newEnroll()
	handleError(t, err)

	if !retryWait(5, getCachedEnroll(er.Id)) {
		t.Errorf("could not get cached enroll: %v", er.Id)
	}

	// remove this enroll entry
	err = DeleteEnroll(er.Id)
	handleError(t, err)

	_, err = GetEnrollStatus(er.Id)
	expectError(t, err, ErrNoRows)
}

// create multiple enrolls, then delete it with a short interval
func TestDeleteExpiredEnrolls(t *testing.T) {
	cleanEnrollTable()

	var i, enrollCount int64 = 0, 10
	for ; i < enrollCount; i++ {
		newEnroll()
	}

	// wait 2 seconds
	time.Sleep(2 * time.Second)

	// match enroll delete limit to added records so test will
	// pass in single run
	testDbConfig.EnrollExpiryDeleteLimit = 10

	// remove entries older than 1 seconds
	count, err := DeleteExpiredEnrolls(1)
	handleError(t, err)

	if count < enrollCount {
		t.Errorf("Expected records deleted >= %d, found: %d", enrollCount, count)
	}
}
