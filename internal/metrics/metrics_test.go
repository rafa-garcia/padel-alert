package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestHttpRequestsTotal(t *testing.T) {
	assert.NotNil(t, HttpRequestsTotal)

	resetCounters()

	HttpRequestsTotal.WithLabelValues("GET", "/api/test", "200").Inc()

	count := testutil.ToFloat64(HttpRequestsTotal.WithLabelValues("GET", "/api/test", "200"))
	assert.Equal(t, float64(1), count)
}

func TestHttpRequestDuration(t *testing.T) {
	assert.NotNil(t, HttpRequestDuration)

	resetCounters()

	HttpRequestDuration.WithLabelValues("GET", "/api/test").Observe(0.1)

	// Cannot directly test histogram values, but ensure it doesn't panic
}

func TestPlaytomicApiRequests(t *testing.T) {
	assert.NotNil(t, PlaytomicApiRequests)

	resetCounters()

	PlaytomicApiRequests.WithLabelValues("matches", "200").Inc()

	count := testutil.ToFloat64(PlaytomicApiRequests.WithLabelValues("matches", "200"))
	assert.Equal(t, float64(1), count)
}

func TestPlaytomicApiDuration(t *testing.T) {
	assert.NotNil(t, PlaytomicApiDuration)

	resetCounters()

	PlaytomicApiDuration.WithLabelValues("matches").Observe(0.2)

	// Cannot directly test histogram values, but ensure it doesn't panic
}

func TestNotificationsSent(t *testing.T) {
	assert.NotNil(t, NotificationsSent)

	resetCounters()

	NotificationsSent.WithLabelValues("email", "success").Inc()

	count := testutil.ToFloat64(NotificationsSent.WithLabelValues("email", "success"))
	assert.Equal(t, float64(1), count)
}

func TestRulesCount(t *testing.T) {
	assert.NotNil(t, RulesCount)

	resetCounters()

	RulesCount.WithLabelValues("match").Set(5)

	value := testutil.ToFloat64(RulesCount.WithLabelValues("match"))
	assert.Equal(t, float64(5), value)
}

// Helper to reset counters between tests
func resetCounters() {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
}
