package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gin_backend",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "gin_backend",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "gin_backend",
			Name:      "http_requests_in_flight",
			Help:      "Current number of HTTP requests being processed.",
		},
	)

	RateLimitAllowedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gin_backend",
			Name:      "rate_limit_allowed_total",
			Help:      "Total number of requests allowed by rate limiter.",
		},
		[]string{"path"},
	)

	RateLimitBlockedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gin_backend",
			Name:      "rate_limit_blocked_total",
			Help:      "Total number of requests blocked by rate limiter.",
		},
		[]string{"path"},
	)

	RateLimitRedisErrorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gin_backend",
			Name:      "rate_limit_redis_error_total",
			Help:      "Total number of Redis errors during rate limiting.",
		},
		[]string{"path"},
	)

	MySQLUp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "gin_backend",
			Name:      "mysql_up",
			Help:      "Whether MySQL is available. 1 means up, 0 means down.",
		},
	)

	RedisUp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "gin_backend",
			Name:      "redis_up",
			Help:      "Whether Redis is available. 1 means up, 0 means down.",
		},
	)

	MySQLOpenConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "gin_backend",
			Name:      "mysql_open_connections",
			Help:      "Current number of open MySQL connections.",
		},
	)

	MySQLInUseConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "gin_backend",
			Name:      "mysql_in_use_connections",
			Help:      "Current number of MySQL connections in use.",
		},
	)

	MySQLIdleConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "gin_backend",
			Name:      "mysql_idle_connections",
			Help:      "Current number of idle MySQL connections.",
		},
	)
)

func InitMetrics() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		HTTPRequestsInFlight,

		RateLimitAllowedTotal,
		RateLimitBlockedTotal,
		RateLimitRedisErrorTotal,

		MySQLUp,
		RedisUp,
		MySQLOpenConnections,
		MySQLInUseConnections,
		MySQLIdleConnections,
	)
}