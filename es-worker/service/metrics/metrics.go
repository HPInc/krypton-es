// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

// Purpose:
// The metrics package is used to initialize and track Prometheus metrics for
// various components in the ESW

package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	// REST request processing latency is partitioned by the REST method. It uses
	// custom buckets based on the expected request duration.
	MetricRestLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "esw_rest_latency_milliseconds",
			Help:       "A latency histogram for REST requests served by ESW",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"method"},
	)
)

// RegisterPrometheusMetrics - register prometheus metrics.
func RegisterPrometheusMetrics() {
	prometheus.MustRegister(MetricRestLatency)
}

// ReportLatencyMetric reports the latency of the specified operation to the
// specified summary vector metric. The label is used to partition the resulting
// histogram.
func ReportLatencyMetric(metric *prometheus.SummaryVec,
	startTime time.Time, label string) {
	duration := time.Since(startTime)
	metric.WithLabelValues(label).Observe(float64(duration.Milliseconds()))
}

// Chronograph is used to measure the time taken by the specified function to
// execute
func Chronograph(logger *zap.Logger, startTime time.Time, functionName string) {
	logger.Info("Execution completed in: ",
		zap.String("Function: ", functionName),
		zap.Duration("Duration (msec): ", time.Since(startTime)),
	)
}
