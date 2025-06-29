// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// Cache request processing latency is partitioned by cache operation. It uses
	// custom buckets based on the expected request duration.
	MetricCacheLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "es_cache_latency_milliseconds",
			Help:       "A latency histogram for cache operations issued by ES",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"method"},
	)
	// Cache errors by method and function
	MetricCacheErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "es_cache_errors",
			Help: "Number of cache errors, partitioned by method and function.",
		},
		[]string{"method", "function"},
	)
	// Cache hits by function
	MetricCacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "es_cache_hits",
			Help: "Number of cache hits, partitioned by function.",
		},
		[]string{"function"},
	)
)

func registerCacheMetrics() {
	prometheus.MustRegister(
		MetricCacheLatency,
		MetricCacheErrors,
		MetricCacheHits,
	)
}

func ReportCacheError(method, function string) {
	MetricCacheErrors.WithLabelValues(method, function).Inc()
}

func ReportCacheHit(function string) {
	MetricCacheHits.WithLabelValues(function).Inc()
}
