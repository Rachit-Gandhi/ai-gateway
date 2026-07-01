package providerstest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ferro-labs/ai-gateway/providers/core"
)

type ChatCompleter interface {
	Complete(context.Context, core.Request) (*core.Response, error)
}

type ChatCompleterFactory func(baseURL string) (ChatCompleter, error)

func AssertMaxCompletionTokensMapped(t *testing.T, factory ChatCompleterFactory, model string, limitFieldPath ...string) {
	t.Helper()

	var captured []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		captured = append(captured, body)
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte(`{"error":{"message":"captured request"}}`))
	}))
	defer server.Close()

	provider, err := factory(server.URL)
	if err != nil {
		t.Fatalf("provider factory returned error: %v", err)
	}

	maxCompletionTokens := 17
	_, _ = provider.Complete(context.Background(), core.Request{
		Model:               model,
		Messages:            []core.Message{{Role: core.RoleUser, Content: "hello"}},
		MaxCompletionTokens: &maxCompletionTokens,
	})

	maxTokens := 23
	_, _ = provider.Complete(context.Background(), core.Request{
		Model:               model,
		Messages:            []core.Message{{Role: core.RoleUser, Content: "hello"}},
		MaxTokens:           &maxTokens,
		MaxCompletionTokens: &maxCompletionTokens,
	})

	if len(captured) != 2 {
		t.Fatalf("captured %d requests, want 2", len(captured))
	}
	assertJSONNumberAtPath(t, captured[0], limitFieldPath, maxCompletionTokens)
	assertJSONNumberAtPath(t, captured[1], limitFieldPath, maxTokens)
}

func assertJSONNumberAtPath(t *testing.T, body map[string]any, path []string, want int) {
	t.Helper()

	var current any = body
	for _, key := range path {
		m, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("request body path %v reached %T before %q", path, current, key)
		}
		current, ok = m[key]
		if !ok {
			t.Fatalf("request body missing path %v; body=%v", path, body)
		}
	}

	got, ok := current.(float64)
	if !ok {
		t.Fatalf("request body path %v = %T(%v), want number %d", path, current, current, want)
	}
	if int(got) != want {
		t.Fatalf("request body path %v = %v, want %d", path, got, want)
	}
}
