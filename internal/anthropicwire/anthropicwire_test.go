package anthropicwire

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ferro-labs/ai-gateway/providers/core"
)

func TestMapTools(t *testing.T) {
	t.Run("nil tools returns nil", func(t *testing.T) {
		if got := MapTools(nil); got != nil {
			t.Fatalf("MapTools(nil) = %#v, want nil", got)
		}
	})

	t.Run("defaults empty schema", func(t *testing.T) {
		got := MapTools([]core.Tool{{
			Type: "function",
			Function: core.Function{
				Name:        "lookup_weather",
				Description: "Look up weather",
			},
		}})

		if len(got) != 1 {
			t.Fatalf("MapTools len = %d, want 1", len(got))
		}
		if got[0].Name != "lookup_weather" || got[0].Description != "Look up weather" {
			t.Fatalf("tool metadata = %#v", got[0])
		}
		var schema map[string]any
		if err := json.Unmarshal(got[0].InputSchema, &schema); err != nil {
			t.Fatalf("unmarshal input_schema: %v", err)
		}
		want := map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		}
		if !reflect.DeepEqual(schema, want) {
			t.Fatalf("input_schema = %#v, want %#v", schema, want)
		}
	})

	t.Run("preserves provided schema", func(t *testing.T) {
		schema := json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}`)
		got := MapTools([]core.Tool{{
			Type: "function",
			Function: core.Function{
				Name:        "lookup_weather",
				Description: "Look up weather",
				Parameters:  schema,
			},
		}})

		if len(got) != 1 {
			t.Fatalf("MapTools len = %d, want 1", len(got))
		}
		if string(got[0].InputSchema) != string(schema) {
			t.Fatalf("input_schema = %s, want %s", got[0].InputSchema, schema)
		}
	})
}
