package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/saviru/spidey/pkg/core"
)

var rdb *redis.Client
var redisCtx = context.Background()

func InitRedis(addr, password string, db int) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

type responseRecorder struct {
	http.ResponseWriter
	body       []byte
	statusCode int
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

type CachedResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

func generateCacheKey(r *http.Request) string {
	raw := r.URL.Path
	if r.URL.RawQuery != "" {
		raw += "?" + r.URL.RawQuery
	}
	hash := sha256.Sum256([]byte(raw))
	return "spidey:cache:" + hex.EncodeToString(hash[:])
}

func Cache(ttl time.Duration) func(*core.Context, func()) {
	return func(c *core.Context, next func()) {
		if rdb == nil || c.Request.Method != http.MethodGet {
			// Skip caching if Redis isn't initialized or method isn't GET
			next()
			return
		}

		key := generateCacheKey(c.Request)

		// Check Redis
		val, err := rdb.Get(redisCtx, key).Result()
		if err == nil && val != "" {
			var cached CachedResponse
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				// Cache HIT
				for k, v := range cached.Headers {
					c.Writer.Header().Set(k, v)
				}
				c.Writer.WriteHeader(cached.Status)
				c.Writer.Write(cached.Body)
				return
			}
		}

		// Cache Miss
		recorder := &responseRecorder{
			ResponseWriter: c.Writer,
			statusCode:     http.StatusOK, // Default to 200
		}

		// Temporarily replace the writer in the context
		originalWriter := c.Writer
		c.Writer = recorder

		next()

		// Restore the original writer
		c.Writer = originalWriter

		if recorder.statusCode >= 200 && recorder.statusCode < 300 {
			// Save to Redis
			headers := make(map[string]string)
			for k, v := range recorder.Header() {
				if len(v) > 0 {
					headers[k] = v[0]
				}
			}

			cachedResponse := CachedResponse{
				Status:  recorder.statusCode,
				Headers: headers,
				Body:    recorder.body,
			}

			if data, err := json.Marshal(cachedResponse); err == nil {
				rdb.Set(redisCtx, key, data, ttl)
			}
		}
	}
}

// removes a specific path from the cache (e.g. "/api/users")
func InvalidateCache(path string) error {
	if rdb == nil {
		return nil
	}

	req, _ := http.NewRequest("GET", path, nil)
	key := generateCacheKey(req)

	return rdb.Del(redisCtx, key).Err()
}
