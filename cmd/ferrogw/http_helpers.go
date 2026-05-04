package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	httppprof "net/http/pprof"
	"os"
	"strings"

	"github.com/ferro-labs/ai-gateway/internal/apierror"
	"github.com/ferro-labs/ai-gateway/internal/metrics"
	"github.com/ferro-labs/ai-gateway/internal/ratelimit"
	webassets "github.com/ferro-labs/ai-gateway/web"
	"github.com/go-chi/chi/v5"
)

var pageTemplates = make(map[string]*template.Template)

func init() {
	pages := []string{
		"getting-started", "overview", "keys", "logs",
		"providers", "config", "analytics", "playground",
	}
	for _, page := range pages {
		tmpl, err := template.ParseFS(webassets.Assets,
			"templates/layout.html",
			"templates/pages/"+page+".html",
		)
		if err != nil {
			panic("failed to parse template " + page + ": " + err.Error())
		}
		pageTemplates[page] = tmpl
	}
}

func renderWebTemplate(w http.ResponseWriter, pageName string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl, ok := pageTemplates[pageName]
	if !ok {
		return fmt.Errorf("unknown page template: %s", pageName)
	}
	return tmpl.ExecuteTemplate(w, "layout.html", data)
}

func mountPprofRoutes(r chi.Router) {
	if !pprofEnabled() {
		return
	}

	r.Route("/debug/pprof", func(r chi.Router) {
		r.Get("/", httppprof.Index)
		r.Get("/cmdline", httppprof.Cmdline)
		r.Get("/profile", httppprof.Profile)
		r.Post("/symbol", httppprof.Symbol)
		r.Get("/symbol", httppprof.Symbol)
		r.Get("/trace", httppprof.Trace)
		r.Get("/allocs", httppprof.Handler("allocs").ServeHTTP)
		r.Get("/block", httppprof.Handler("block").ServeHTTP)
		r.Get("/goroutine", httppprof.Handler("goroutine").ServeHTTP)
		r.Get("/heap", httppprof.Handler("heap").ServeHTTP)
		r.Get("/mutex", httppprof.Handler("mutex").ServeHTTP)
		r.Get("/threadcreate", httppprof.Handler("threadcreate").ServeHTTP)
	})
}

func pprofEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("ENABLE_PPROF")))
	return v == "1" || v == "true" || v == "yes"
}

func rateLimitMiddleware(store *ratelimit.Store) func(http.Handler) http.Handler {
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

func serveLogo(w http.ResponseWriter) {
	data, err := fs.ReadFile(webassets.Assets, "logo.png")
	if err != nil {
		apierror.WriteOpenAI(w, http.StatusNotFound, "logo not found", "not_found_error", "resource_not_found")
		return
	}
	w.Header().Set("Content-Type", "image/png")
	_, _ = w.Write(data)
}
