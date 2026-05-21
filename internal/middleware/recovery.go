package middleware
import (
	"net/http"
	"runtime/debug"

	"go-test/pkg/response"
	"go-test/pkg/xcontext"
	
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)


func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered",
					zap.String("request_id", xcontext.GetRequestID(c.Request.Context())),
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("stack", string(debug.Stack())),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, response.FailData(50000, "系统内部错误"))
			}
		}()

		c.Next()
	}
}