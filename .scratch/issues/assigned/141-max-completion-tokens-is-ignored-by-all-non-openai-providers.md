# Issue #141: max_completion_tokens is ignored by all non-OpenAI providers

- URL: https://github.com/ferro-labs/ai-gateway/issues/141
- State: OPEN
- Created: 2026-05-24T07:44:28Z
- Updated: 2026-06-15T09:08:26Z
- Labels: bug, priority: high, release-1.1.4
- Assignee: Rachit-Gandhi

## Problem
Only `max_tokens` is read; `max_completion_tokens` (the o-series convention) is ignored by all non-OpenAI providers. Anthropic additionally hard-defaults 1024, so a client sending `max_completion_tokens=4096` receives ≤1024 tokens.

## Fix
Coalesce `MaxCompletionTokens` → `MaxTokens` in every provider.

## Acceptance
Test: `max_completion_tokens` is honored across providers.

_Internal audit → R-H3._
