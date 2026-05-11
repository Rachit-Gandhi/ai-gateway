package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/ferro-labs/ai-gateway/internal/admin"
)

// ProxyAuth returns a middleware that requires auth on proxy routes by default.
// Set ALLOW_UNAUTHENTICATED_PROXY=true to disable (local dev only).
func ProxyAuth(store admin.Store, masterKey string) func(http.Handler) http.Handler {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("ALLOW_UNAUTHENTICATED_PROXY")), "true") {
		return func(next http.Handler) http.Handler { return next }
	}
	return admin.AuthMiddleware(store, masterKey)
}
