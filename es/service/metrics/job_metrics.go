// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// Job runs by job
	metricJobRuns = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "es_scheduled_job_runs",
			Help: "Number of job runs, partitioned by run name.",
		},
		[]string{"function"},
	)
)

func registerJobMetrics() {
	prometheus.MustRegister(
		metricJobRuns,
	)
}

func ReportJobRun(function string) {
	metricJobRuns.WithLabelValues(function).Inc()
}
