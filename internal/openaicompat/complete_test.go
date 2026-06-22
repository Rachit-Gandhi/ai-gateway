package openaicompat

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ferro-labs/ai-gateway/providers/core"
)

func TestAPIError(t *testing.T) {
	tests := []struct {
		name string
		body []byte
		want string
	}{
		{
			name: "structured envelope",
			body: []byte(`{"error":{"message":"rate limited"}}`),
			want: "groq API error (429): rate limited",
		},
		{
			name: "raw body fallback",
			body: []byte(`temporarily unavailable`),
			want: "groq API error (429): temporarily unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := APIError("groq", http.StatusTooManyRequests, tt.body).Error(); got != tt.want {
				t.Fatalf("APIError = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPostChat_Non200ReturnsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(w, `{"error":{"message":"slow down"}}`)
	}))
	defer srv.Close()

	_, err := PostChat(context.Background(), chatParamsForTest(srv.URL), requestForTest())
	if err == nil {
		t.Fatal("PostChat returned nil error, want API error")
	}
	if got, want := err.Error(), "groq API error (429): slow down"; got != want {
		t.Fatalf("PostChat error = %q, want %q", got, want)
	}
}

func TestPostChat_InvalidJSONBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{not json`)
	}))
	defer srv.Close()

	_, err := PostChat(context.Background(), chatParamsForTest(srv.URL), requestForTest())
	if err == nil {
		t.Fatal("PostChat returned nil error, want unmarshal error")
	}
	if !strings.Contains(err.Error(), "failed to unmarshal response") {
		t.Fatalf("PostChat error = %q, want unmarshal context", err.Error())
	}
}

func TestPostStream_Non200ReturnsAPIErrorBeforeChannel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, `maintenance`)
	}))
	defer srv.Close()

	ch, err := PostStream(context.Background(), chatParamsForTest(srv.URL), requestForTest())
	if err == nil {
		t.Fatal("PostStream returned nil error, want API error")
	}
	if ch != nil {
		t.Fatalf("PostStream channel = %#v, want nil on non-200", ch)
	}
	if got, want := err.Error(), "groq API error (503): maintenance"; got != want {
		t.Fatalf("PostStream error = %q, want %q", got, want)
	}
}

func chatParamsForTest(url string) ChatParams {
	return ChatParams{
		HTTPClient: http.DefaultClient,
		URL:        url,
		Headers:    map[string]string{"Authorization": "Bearer test-key"},
		Provider:   "groq",
		Label:      "groq",
	}
}

func requestForTest() core.Request {
	return core.Request{
		Model:    "test-model",
		Messages: []core.Message{{Role: core.RoleUser, Content: "hi"}},
	}
}

var errBoom = errors.New("boom")
