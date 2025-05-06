package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HttpRequestsTotal counts the number of HTTP requests processed
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "padel_alert_http_requests_total",
			Help: "The total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HttpRequestDuration tracks the duration of HTTP requests
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "padel_alert_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// PlaytomicApiRequests counts the number of requests to Playtomic API
	PlaytomicApiRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "padel_alert_playtomic_api_requests_total",
			Help: "The total number of requests to Playtomic API",
		},
		[]string{"endpoint", "status"},
	)

	// PlaytomicApiDuration tracks the duration of Playtomic API requests
	PlaytomicApiDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "padel_alert_playtomic_api_duration_seconds",
			Help:    "Playtomic API request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)

	// NotificationsSent counts the number of notifications sent
	NotificationsSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "padel_alert_notifications_sent_total",
			Help: "The total number of notifications sent",
		},
		[]string{"type", "status"},
	)

	// RulesCount tracks the current number of rules
	RulesCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "padel_alert_rules_count",
			Help: "The current number of rules",
		},
		[]string{"type"},
	)
)
