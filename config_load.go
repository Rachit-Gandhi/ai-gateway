package aigateway

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ferro-labs/ai-gateway/internal/tracingpolicy"
	"gopkg.in/yaml.v3"
)

// LoadConfig reads and parses a config file from the given path.
// Supported formats: JSON (.json), YAML (.yaml, .yml).
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file extension %q: use .json, .yaml, or .yml", ext)
	}

	resolveEnv(&cfg)

	return &cfg, nil
}

// resolveEnv materialises environment-variable references in config sections
// that carry user/plugin-owned secrets. It intentionally leaves
// observability.tracing.headers alone: those are still resolved lazily by
// internal/otel so programmatic configs keep the historical behaviour.
func resolveEnv(cfg *Config) {
	if cfg == nil {
		return
	}

	for i := range cfg.MCPServers {
		cfg.MCPServers[i].Headers = resolveEnvStringMap(cfg.MCPServers[i].Headers, true)
	}

	for i := range cfg.Observability.Exporters {
		cfg.Observability.Exporters[i].Config = resolveEnvAnyMap(cfg.Observability.Exporters[i].Config)
	}

	for i := range cfg.Plugins {
		cfg.Plugins[i].Config = resolveEnvAnyMap(cfg.Plugins[i].Config)
	}
}

func resolveEnvStringMap(raw map[string]string, skipEmpty bool) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	out := make(map[string]string, len(raw))
	for k, v := range raw {
		resolved := os.Expand(v, os.Getenv)
		if skipEmpty && resolved == "" {
			continue
		}
		out[k] = resolved
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func resolveEnvAnyMap(raw map[string]any) map[string]any {
	if raw == nil {
		return nil
	}
	out := make(map[string]any, len(raw))
	for k, v := range raw {
		out[k] = resolveEnvAnyValue(v)
	}
	return out
}

func resolveEnvAnyValue(v any) any {
	switch val := v.(type) {
	case string:
		return os.Expand(val, os.Getenv)
	case map[string]any:
		return resolveEnvAnyMap(val)
	case map[string]string:
		return resolveEnvStringMap(val, false)
	case []any:
		out := make([]any, len(val))
		for i, elem := range val {
			out[i] = resolveEnvAnyValue(elem)
		}
		return out
	case []string:
		out := make([]string, len(val))
		for i, elem := range val {
			out[i] = os.Expand(elem, os.Getenv)
		}
		return out
	default:
		return v
	}
}

// ValidateConfig validates a Config for correctness.
func ValidateConfig(cfg Config) error {
	// Default to single strategy when mode is omitted to match runtime behavior.
	mode := cfg.Strategy.Mode
	if mode == "" {
		mode = ModeSingle
	}

	switch mode {
	case ModeSingle, ModeFallback, ModeLoadBalance, ModeConditional, ModeLatency, ModeCostOptimized,
		ModeContentBased, ModeABTest:
	default:
		return fmt.Errorf("unknown strategy mode: %q", cfg.Strategy.Mode)
	}

	if len(cfg.Targets) == 0 {
		return fmt.Errorf("at least one target is required")
	}

	if mode == ModeConditional && len(cfg.Strategy.Conditions) == 0 {
		return fmt.Errorf("conditional strategy requires at least one condition")
	}

	if mode == ModeContentBased && len(cfg.Strategy.ContentConditions) == 0 {
		return fmt.Errorf("content-based strategy requires at least one content_condition")
	}

	if mode == ModeABTest && len(cfg.Strategy.ABVariants) == 0 {
		return fmt.Errorf("ab-test strategy requires at least one ab_variant")
	}

	if mode == ModeCostOptimized {
		switch cfg.Strategy.UnpricedStrategy {
		case "", unpricedStrategyFallback, unpricedStrategySkip, unpricedStrategyAllow:
		default:
			return fmt.Errorf("cost-optimized unpriced_strategy must be one of fallback, skip, allow")
		}
	}

	if mode == ModeLoadBalance {
		var sum float64
		for _, t := range cfg.Targets {
			if t.Weight < 0 {
				return fmt.Errorf("target %q has negative weight", t.VirtualKey)
			}
			sum += t.Weight
		}
		if sum <= 0 {
			return fmt.Errorf("loadbalance strategy requires total weight > 0")
		}
	}

	// Validate observability.tracing.privacy_level against the single source of
	// truth in the internal tracingpolicy package (shared with internal/otel).
	if err := tracingpolicy.ValidatePrivacyLevel(cfg.Observability.Tracing.PrivacyLevel); err != nil {
		return fmt.Errorf("observability.tracing: %w", err)
	}

	// Validate aliases: no alias may point to another alias (no cycles/chains).
	for name, target := range cfg.Aliases {
		if name == "" {
			return fmt.Errorf("alias name must not be empty")
		}
		if target == "" {
			return fmt.Errorf("alias %q must not map to an empty string", name)
		}
		if name == target {
			return fmt.Errorf("alias %q must not point to itself", name)
		}
		if _, chainedAlias := cfg.Aliases[target]; chainedAlias {
			return fmt.Errorf("alias %q points to another alias %q; chained aliases are not supported", name, target)
		}
	}

	return nil
}
