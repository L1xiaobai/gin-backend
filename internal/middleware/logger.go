package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger(logger *zap.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {

		start := time.Now()

		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		cost := time.Since(start)

		statusCode := c.Writer.Status()

		logger.Info(
			"http request",
			zap.String("path", path),
			zap.String("method", method),
			zap.Int("status", statusCode),
			zap.Duration("latency", cost),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}