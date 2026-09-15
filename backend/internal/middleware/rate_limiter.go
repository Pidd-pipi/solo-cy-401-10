package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gigmatch/gigmatch/internal/constants"
	"github.com/gigmatch/gigmatch/internal/util"
)

type ipWindow struct {
	count   int
	resetAt time.Time
}

type ipCounter struct {
	mu       sync.Mutex
	requests map[string]*ipWindow
}

var limiter = &ipCounter{requests: map[string]*ipWindow{}}

// RateLimit caps requests per IP at maxRequests per minute.
func RateLimit(maxRequests int, logger *slog.Logger) gin.HandlerFunc {
	if maxRequests <= 0 {
		maxRequests = 120
	}
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		limiter.mu.Lock()
		w, ok := limiter.requests[ip]
		if !ok || now.After(w.resetAt) {
			w = &ipWindow{count: 0, resetAt: now.Add(time.Minute)}
			limiter.requests[ip] = w
			if len(limiter.requests) > 10_000 {
				for key, val := range limiter.requests {
					if now.After(val.resetAt) {
						delete(limiter.requests, key)
					}
				}
			}
		}
		w.count++
		exceeded := w.count > maxRequests
		limiter.mu.Unlock()

		if exceeded {
			if logger != nil {
				logger.Warn("rate limit exceeded", "client_ip", ip)
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests, util.Response{
				Code:    constants.CodeTooManyRequests,
				Message: constants.MsgTooManyRequests,
			})
			return
		}
		c.Next()
	}
}
