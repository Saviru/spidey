package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/saviru/spidey/pkg/core"
	"golang.org/x/time/rate"
)

type RateLimitConfig struct {
	MaxRequests int           // Maximum tokens/requests
	Window      time.Duration // Time window to refill tokens
	Store       string        // "redis" (distributed) or "memory" (single-node)

	KeyFunc    func(c *core.Context) string // Custom Key (e.g. API key)
	SkipFunc   func(c *core.Context) bool   // Bypass logic
	RejectFunc func(c *core.Context)        // Custom JSON error response
}

var memoryLimiters = sync.Map{}

// extracts the real IP address
func DefaultKeyFunc(c *core.Context) string {
	ip := c.Request.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = c.Request.RemoteAddr
	}

	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}
	return ip
}

const redisRateLimitScript = `
local key = KEYS[1]
local window = tonumber(ARGV[1])
local max = tonumber(ARGV[2])

local current = redis.call("INCR", key)
if current == 1 then
    redis.call("PEXPIRE", key, window)
end

local ttl = redis.call("PTTL", key)
return {current, ttl}
`

// RateLimit creates an advanced rate limiting middleware
func RateLimit(config RateLimitConfig) func(*core.Context, func()) {
	// Apply Defaults
	if config.MaxRequests <= 0 {
		config.MaxRequests = 100
	}
	if config.Window <= 0 {
		config.Window = time.Minute
	}
	if config.KeyFunc == nil {
		config.KeyFunc = DefaultKeyFunc
	}
	if config.Store == "" {
		config.Store = "memory"
	}
	if config.RejectFunc == nil {
		config.RejectFunc = func(c *core.Context) {
			c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "Too Many Requests",
			})
		}
	}

	// Calculate rate for memory limiter (tokens per second)
	requestsPerSecond := float64(config.MaxRequests) / config.Window.Seconds()

	return func(c *core.Context, next func()) {
		// 1. Skip logic
		if config.SkipFunc != nil && config.SkipFunc(c) {
			next()
			return
		}

		// 2. Generate Key
		key := "spidey:ratelimit:" + config.KeyFunc(c)
		var remaining int
		var resetTime int64
		var allowed bool

		// 3. Evaluate Limits
		useMemory := true

		if config.Store == "redis" && rdb != nil {
			// Try Redis first
			result, err := rdb.Eval(redisCtx, redisRateLimitScript, []string{key}, int64(config.Window/time.Millisecond), config.MaxRequests).Result()
			if err == nil {
				useMemory = false
				res := result.([]interface{})
				current := int(res[0].(int64))
				ttlMs := res[1].(int64)

				remaining = config.MaxRequests - current
				if remaining < 0 {
					remaining = 0
				}
				resetTime = time.Now().Add(time.Duration(ttlMs) * time.Millisecond).Unix()
				allowed = current <= config.MaxRequests
			}
			// If Redis fails, useMemory remains true (Graceful Fallback)
		}

		if useMemory {
			// Fallback to in-memory Token Bucket using x/time/rate
			limiterInterface, _ := memoryLimiters.LoadOrStore(key, rate.NewLimiter(rate.Limit(requestsPerSecond), config.MaxRequests))
			limiter := limiterInterface.(*rate.Limiter)

			allowed = limiter.Allow()
			remaining = int(limiter.Tokens())
			// Rough estimate for reset time in token bucket
			resetTime = time.Now().Add(config.Window).Unix()
		}

		// 4. Inject standard Rate Limit Headers
		c.Writer.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.MaxRequests))
		c.Writer.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Writer.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

		// 5. Reject or Proceed
		if !allowed {
			c.Writer.Header().Set("Retry-After", strconv.FormatInt(resetTime-time.Now().Unix(), 10))
			config.RejectFunc(c)
			return // DO NOT CALL NEXT()
		}

		next()
	}
}
