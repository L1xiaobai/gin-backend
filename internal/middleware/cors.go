package middleware

import (
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	// 允许的来源列表，可以根据需要进行调整
	allowedOrigins := map[string]bool{
        "http://localhost:3000": true,
        "http://127.0.0.1:3000": true,
    }

    return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
        if allowedOrigins[origin] {
            c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
            c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        } 	// 允许来自XXX的跨域请求

        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true") // 允许携带凭证（如 cookies, HTTPS认证信息）
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE") // 允许的 HTTP 方法
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With") // 允许的请求头
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type") // 允许客户端访问的响应头
		c.Writer.Header().Set("Access-Control-Max-Age", "3600") // 预检请求的缓存时间（单位：秒）
		// 通常不用手动设置
		// c.Writer.Header().Set("Access-Control-Request-Headers", "Authorization, Content-Type, X-Requested-With") // 预检请求中允许的请求头
		// c.Writer.Header().Set("Access-Control-Request-Method", "POST, OPTIONS, GET, PUT, DELETE") // 预检请求中允许的 HTTP 方法

		// 处理预检请求
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}