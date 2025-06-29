// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package jobs

import (
	"context"
	"time"

	"github.com/HPInc/krypton-es/es/service/config"
	"github.com/HPInc/krypton-es/es/service/db"
	"github.com/HPInc/krypton-es/es/service/metrics"
	"go.uber.org/zap"
)

type jobFunc func() error

var (
	esLogger *zap.Logger

	jobs *config.ScheduledJobs

	gCtx        context.Context
	gCancelFunc context.CancelFunc

	// connect job names to their runner functions
	jobsMap = map[string]jobFunc{
		"delete_expired_enrolls": db.TriggerDeleteExpiredEnrolls,
	}
)

const (
	timeLayout = "15:04:05" //24 hour time only format to read in start time
)

func Init(logger *zap.Logger, jobsConfig *config.ScheduledJobs) error {
	jobs = jobsConfig
	esLogger = logger

	gCtx, gCancelFunc = context.WithCancel(context.Background())

	for k, v := range *jobs {
		// check if this job is enabled, if not, skip
		if !v.Enabled {
			esLogger.Info("Skipping disabled job",
				zap.String("name", k))
			continue
		}
		// look up the runner func
		if _, ok := jobsMap[k]; !ok {
			esLogger.Info("Skipping job without runner config",
				zap.String("name", k))
			continue
		}
		val := v
		if err := scheduleJob(k, &val); err != nil {
			return err
		}
	}
	return nil
}

func scheduleJob(key string, job *config.ScheduledJob) error {
	esLogger.Info("Scheduling job",
		zap.String("name", key),
		zap.Bool("enabled", job.Enabled),
		zap.String("start", job.Start),
		zap.String("every", job.Every))

	t, err := time.Parse(timeLayout, job.Start)
	if err != nil {
		esLogger.Error("Error parsing start time",
			zap.String("start", job.Start),
			zap.Error(err))
		return err
	}
	every, err := time.ParseDuration(job.Every)
	if err != nil {
		esLogger.Error("Error parsing repeat string",
			zap.String("every", job.Every),
			zap.Error(err))
		return err
	}
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(),
		t.Hour(), t.Minute(), t.Second(), 0, time.UTC)

	// advance by a day if time is in the past so we get scheduled for
	// the same time next day
	if time.Until(startTime) < 0 {
		esLogger.Info("Job start time is in the past. Advancing by a day",
			zap.Duration("time_diff", time.Until(startTime)))
		startTime = startTime.AddDate(0, 0, 1)
	}
	go runner(gCtx, key, startTime, every)

	return nil
}

func runner(ctx context.Context, key string, start time.Time, repeat time.Duration) {
	esLogger.Info("Job scheduled",
		zap.String("name", key),
		zap.Duration("starts_in", time.Until(start)),
		zap.Duration("repeat", repeat))

	timer := time.NewTimer(time.Until(start))

	for {
		select {
		case <-timer.C:
			esLogger.Info("Starting job", zap.String("name", key))
			metrics.ReportJobRun(key)
			if err := jobsMap[key](); err != nil {
				esLogger.Info("There was an error running scheduled job",
					zap.String("name", key))
			}
			_ = timer.Reset(repeat) // schedule to repeat
			esLogger.Info("Re-scheduling job",
				zap.String("name", key),
				zap.Duration("next_run_in", repeat))

		case <-gCtx.Done():
			esLogger.Info("Stopping job", zap.String("name", key))
			timer.Stop()
			return
		}
	}
}

func Shutdown() {
	esLogger.Info("Device Enrollment Service: Shutting down scheduled jobs")
	if gCancelFunc != nil {
		esLogger.Info("Cancelling scheduled jobs")
		gCancelFunc()
	}
}
