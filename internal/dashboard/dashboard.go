// Package dashboard provides template rendering and asset helpers for the
// gateway web dashboard.
package dashboard

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	httppprof "net/http/pprof"
	"os"
	"strings"

	"github.com/ferro-labs/ai-gateway/internal/apierror"
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

// RenderWebTemplate writes the named page template to w.
func RenderWebTemplate(w http.ResponseWriter, pageName string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl, ok := pageTemplates[pageName]
	if !ok {
		return fmt.Errorf("unknown page template: %s", pageName)
	}
	return tmpl.ExecuteTemplate(w, "layout.html", data)
}

// MountPprofRoutes registers /debug/pprof/* routes on r when ENABLE_PPROF is set.
func MountPprofRoutes(r chi.Router) {
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

// ServeLogo writes the embedded logo.png to w.
func ServeLogo(w http.ResponseWriter) {
	data, err := fs.ReadFile(webassets.Assets, "logo.png")
	if err != nil {
		apierror.WriteOpenAI(w, http.StatusNotFound, "logo not found", "not_found_error", "resource_not_found")
		return
	}
	w.Header().Set("Content-Type", "image/png")
	_, _ = w.Write(data)
}
