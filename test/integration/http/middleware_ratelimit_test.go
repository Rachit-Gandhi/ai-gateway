//go:build integration
// +build integration

package http_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

func TestMiddlewareRateLimit_ExceedLimit_Returns429(t *testing.T) {
	// Configure rate limit: 1 RPS, burst 1.
	// The first request consumes the burst token; subsequent rapid requests
	// should be rejected before the bucket refills.
	env := newTestServer(t, withRateLimit(1, 1))

	// Fire 10 concurrent requests from the same "IP" — at least some must get 429.
	// We set X-Forwarded-For to force a consistent key (by default, httptest
	// connections use different ephemeral ports in RemoteAddr).
	var wg sync.WaitGroup
	var count429 atomic.Int32
	var count200 atomic.Int32

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("GET", env.Server.URL+"/health", nil)
			req.Header.Set("X-Forwarded-For", "10.0.0.1")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			resp.Body.Close()
			switch resp.StatusCode {
			case http.StatusTooManyRequests:
				count429.Add(1)
			case http.StatusOK:
				count200.Add(1)
			}
		}()
	}
	wg.Wait()

	if count429.Load() == 0 {
		t.Fatalf("expected at least one 429 response, got %d OK and %d 429", count200.Load(), count429.Load())
	}
	t.Logf("rate limit results: %d OK, %d 429", count200.Load(), count429.Load())
}

func TestMiddlewareRateLimit_UsesXRealIPFromRouterRealIP(t *testing.T) {
	// The full router must keep chi's RealIP middleware before the rate limiter:
	// X-Real-IP clients behind the same trusted proxy should get independent
	// buckets even though their socket RemoteAddr is identical.
	env := newTestServer(t, withRateLimit(1, 1))
	handler := env.Server.Config.Handler

	requestFrom := func(clientIP string) int {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.RemoteAddr = "10.0.0.10:12345"
		req.Header.Set("X-Real-IP", clientIP)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	if got := requestFrom("203.0.113.10"); got != http.StatusOK {
		t.Fatalf("first request from client A: got %d, want 200", got)
	}
	if got := requestFrom("203.0.113.10"); got != http.StatusTooManyRequests {
		t.Fatalf("second request from client A: got %d, want 429", got)
	}
	if got := requestFrom("203.0.113.11"); got != http.StatusOK {
		t.Fatalf("first request from client B behind same proxy: got %d, want 200", got)
	}
}

func TestMiddlewareRateLimit_NotEnabled_NoRejection(t *testing.T) {
	// Default server with no rate limiter — all requests should pass.
	env := newTestServer(t)

	for i := 0; i < 10; i++ {
		resp, err := http.Get(env.Server.URL + "/health")
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, resp.StatusCode)
		}
	}
}
