// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// REST request processing latency is partitioned by the REST method. It uses
	// custom buckets based on the expected request duration.
	MetricRestLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "es_rest_latency_milliseconds",
			Help:       "A latency histogram for REST requests served by ES",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"method"},
	)

	// Number of REST requests received by ES.
	MetricRequestCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "es_rest_requests",
		Help:        "Number of requests received by ES",
		ConstLabels: prometheus.Labels{"version": "1"},
	})

	// Enroll request errors by code
	MetricEnrollErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "es_rest_enroll_errors",
			Help: "Number of enroll errors, partitioned by api and error code.",
		},
		[]string{"method", "code"},
	)
)

func registerRestMetrics() {
	prometheus.MustRegister(
		MetricRestLatency,
		MetricRequestCount,
		MetricEnrollErrors)
}

func ReportRestError(method string, code int) {
	MetricEnrollErrors.WithLabelValues(method, strconv.Itoa(code)).Inc()
}
