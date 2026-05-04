package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var dummyHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestCORS_Wildcard_WhenNoOriginsConfigured(t *testing.T) {
	mw := CORS()
	handler := mw(dummyHandler)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected wildcard origin, got %q", got)
	}
}

func TestCORS_Wildcard_WhenOnlyEmptyStrings(t *testing.T) {
	mw := CORS("", "  ")
	handler := mw(dummyHandler)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected wildcard origin, got %q", got)
	}
}

func TestCORS_AllowedOrigin_SetsHeaderAndVary(t *testing.T) {
	mw := CORS("https://example.com", "https://other.com")
	handler := mw(dummyHandler)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("expected https://example.com, got %q", got)
	}
	if got := w.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("expected Vary: Origin, got %q", got)
	}
}

func TestCORS_DisallowedOrigin_NoAllowOriginHeader(t *testing.T) {
	mw := CORS("https://example.com")
	handler := mw(dummyHandler)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no Allow-Origin header, got %q", got)
	}
}

func TestCORS_PreflightOptions_Returns204(t *testing.T) {
	mw := CORS("https://example.com")
	handler := mw(dummyHandler)

	r := httptest.NewRequest(http.MethodOptions, "/", nil)
	r.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestCORS_PreflightOptions_DoesNotCallNext(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	})
	mw := CORS("https://example.com")
	handler := mw(next)

	r := httptest.NewRequest(http.MethodOptions, "/", nil)
	r.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if called {
		t.Fatal("next handler should not be called for OPTIONS preflight")
	}
}

func TestCORS_StandardHeaders_AlwaysSet(t *testing.T) {
	mw := CORS("https://example.com")
	handler := mw(dummyHandler)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if got := w.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Fatalf("unexpected Allow-Methods: %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type, Authorization, X-Provider" {
		t.Fatalf("unexpected Allow-Headers: %q", got)
	}
	if got := w.Header().Get("Access-Control-Max-Age"); got != "86400" {
		t.Fatalf("unexpected Max-Age: %q", got)
	}
}

func TestCORS_NonOptions_CallsNextHandler(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	mw := CORS()
	handler := mw(next)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if !called {
		t.Fatal("next handler should be called for non-OPTIONS request")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCORS_TrimsWhitespaceFromOrigins(t *testing.T) {
	mw := CORS("  https://example.com  ")
	handler := mw(dummyHandler)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("expected https://example.com, got %q", got)
	}
}
