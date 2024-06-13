package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
	"net/http"
	"time"
)

// RateLimitMiddleware 令牌桶限流
func RateLimitMiddleware(fillInterval time.Duration, cap int64) func(c *gin.Context) {
	bucket := ratelimit.NewBucket(fillInterval, cap)
	return func(c *gin.Context) {
		//	如果取不到令牌就中断本次请求并且返回 rate limit
		if bucket.TakeAvailable(1) < 1 {
			c.String(http.StatusOK, "服务器请求繁忙")
			c.Abort()
			return
		}
		c.Next()
	}
}
