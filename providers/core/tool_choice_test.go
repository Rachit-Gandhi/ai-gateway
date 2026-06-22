package core

import "testing"

func TestNormalizeToolChoice(t *testing.T) {
	tests := []struct {
		name     string
		choice   any
		wantKind ToolChoiceKind
		wantName string
	}{
		{name: "nil", choice: nil, wantKind: ToolChoiceUnset},
		{name: "auto string", choice: "auto", wantKind: ToolChoiceAuto},
		{name: "none string", choice: "none", wantKind: ToolChoiceNone},
		{name: "required string", choice: "required", wantKind: ToolChoiceRequired},
		{
			name: "forced function map",
			choice: map[string]any{
				"type":     "function",
				"function": map[string]any{"name": "lookup_weather"},
			},
			wantKind: ToolChoiceFunction,
			wantName: "lookup_weather",
		},
		{
			name: "forced function struct",
			choice: struct {
				Type     string `json:"type"`
				Function struct {
					Name string `json:"name"`
				} `json:"function"`
			}{
				Type: "function",
				Function: struct {
					Name string `json:"name"`
				}{Name: "lookup_weather"},
			},
			wantKind: ToolChoiceFunction,
			wantName: "lookup_weather",
		},
		{name: "unknown string", choice: "bogus", wantKind: ToolChoiceUnset},
		{
			name: "function without name",
			choice: map[string]any{
				"type":     "function",
				"function": map[string]any{"name": ""},
			},
			wantKind: ToolChoiceUnset,
		},
		{
			name: "unknown object type",
			choice: map[string]any{
				"type":     "other",
				"function": map[string]any{"name": "lookup_weather"},
			},
			wantKind: ToolChoiceUnset,
		},
		{name: "unsupported scalar", choice: 123, wantKind: ToolChoiceUnset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKind, gotName := NormalizeToolChoice(tt.choice)
			if gotKind != tt.wantKind || gotName != tt.wantName {
				t.Fatalf("NormalizeToolChoice(%#v) = (%v, %q), want (%v, %q)", tt.choice, gotKind, gotName, tt.wantKind, tt.wantName)
			}
		})
	}
}
