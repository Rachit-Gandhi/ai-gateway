package middleware

import (
	"net/http"
	"strings"

	"github.com/ferro-labs/ai-gateway/internal/apierror"
	"github.com/ferro-labs/ai-gateway/internal/metrics"
	"github.com/ferro-labs/ai-gateway/internal/ratelimit"
)

// RateLimit returns middleware that enforces per-IP token-bucket rate limiting.
func RateLimit(store *ratelimit.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				parts := strings.SplitN(xff, ",", 2)
				ip = strings.TrimSpace(parts[0])
			}
			if !store.Allow(ip) {
				metrics.RateLimitRejections.WithLabelValues("ip").Inc()
				apierror.WriteOpenAI(w, http.StatusTooManyRequests,
					"rate limit exceeded", "rate_limit_error", "rate_limit_exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
