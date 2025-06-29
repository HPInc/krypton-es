// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// Database request processing latency is partitioned by the Postgres method. It uses
	// custom buckets based on the expected request duration.
	MetricDatabaseLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "es_db_latency_milliseconds",
			Help:       "A latency histogram for database operations issued by ES",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"method"},
	)
	// Total number of errors committing database transactions.
	MetricDatabaseCommitErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "es_db_commit_errors",
			Help: "Total number of errors committing transactions",
		})

	// Total number of errors rolling back database transactions.
	MetricDatabaseRollbackErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "es_db_rollback_errors",
			Help: "Total number of errors rolling back transactions",
		})
)

func registerDatabaseMetrics() {
	prometheus.MustRegister(
		MetricDatabaseLatency,
		MetricDatabaseCommitErrors,
		MetricDatabaseRollbackErrors,
	)
}
