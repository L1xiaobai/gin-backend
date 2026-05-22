package middleware

import (
	"strconv"
	"time"

	"go-test/pkg/metrics"

	"github.com/gin-gonic/gin"
)

func Prometheus() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		method := c.Request.Method

		metrics.HTTPRequestsInFlight.Inc()
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		metrics.HTTPRequestsInFlight.Dec()
		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path, status).Observe(duration)
	}
}