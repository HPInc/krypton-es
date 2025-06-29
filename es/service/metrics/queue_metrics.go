// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// Total number of enroll notifications processed successfully.
	MetricNotificationEnrolls = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "es_queue_enroll_notifications",
			Help: "Total number of enroll notifications processed successfully",
		})

	// Total number of enroll notifications that failed processing.
	MetricNotificationEnrollsFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "es_queue_enroll_notifications_failed",
			Help: "Total number of enroll notifications that failed",
		})

	// Total number of errors parsing enroll notifications.
	MetricNotificationEnrollParseErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "es_queue_enroll_notification_parse_errors",
			Help: "Total number of errors parse enroll notifications",
		})

	// Total number of errors deleting enroll notifications.
	MetricNotificationEnrollDeleteErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "es_queue_enroll_notification_delete_errors",
			Help: "Total number of errors deleting enroll notifications",
		})

	// Total number of enroll error notifications processed successfully.
	MetricNotificationEnrollErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "es_queue_enroll_errors_processed",
			Help: "Total number of enroll errors processed successfully",
		})

	// Total number of enroll errorsthat failed processing.
	MetricNotificationEnrollErrorsFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "es_queue_enroll_errors_failed",
			Help: "Total number of enroll errors that failed",
		})

	// Total number of errors parsing enroll errors.
	MetricNotificationEnrollErrorParseErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "es_queue_enroll_errors_parse_errors",
			Help: "Total number of errors parse enroll errors",
		})

	// Total number of errors deleting error queue entries.
	MetricNotificationEnrollErrorDeleteErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "es_queue_enroll_errors_delete_errors",
			Help: "Total number of errors deleting enroll errors",
		})

	// Total number of unenroll notifications processed successfully.
	MetricNotificationUnenrolls = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "es_queue_unenroll_notifications",
			Help: "Total number of unenroll notifications processed successfully",
		})

	// Total number of unenroll notifications that failed processing..
	MetricNotificationUnenrollsFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "es_queue_unenroll_notifications_failed",
			Help: "Total number of unenroll notifications that failed processing",
		})

	// Total number of enroll error notifications processed successfully.
	MetricNotificationUnenrollErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "es_queue_unenroll_errors_processed",
			Help: "Total number of unenroll errors processed successfully",
		})

	// Total number of enroll errors that failed processing.
	MetricNotificationUnenrollErrorsFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "es_queue_unenroll_errors_failed",
			Help: "Total number of unenroll errors that failed",
		})

	// Total number of errors deleting unenroll queue entries.
	MetricNotificationUnenrollDeleteErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "es_queue_unenroll_delete_errors",
			Help: "Total number of errors deleting unenroll entries",
		})
)

func registerQueueMetrics() {
	prometheus.MustRegister(
		MetricNotificationEnrolls,
		MetricNotificationEnrollsFailed,
		MetricNotificationEnrollParseErrors,
		MetricNotificationEnrollDeleteErrors,
		MetricNotificationEnrollErrors,
		MetricNotificationEnrollErrorsFailed,
		MetricNotificationEnrollErrorParseErrors,
		MetricNotificationEnrollErrorDeleteErrors,
		MetricNotificationUnenrolls,
		MetricNotificationUnenrollsFailed,
		MetricNotificationUnenrollErrors,
		MetricNotificationUnenrollErrorsFailed,
		MetricNotificationUnenrollDeleteErrors,
	)
}
