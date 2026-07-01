# Issue #139: Tool/function calling is dropped on native non-OpenAI providers

- URL: https://github.com/ferro-labs/ai-gateway/issues/139
- State: OPEN
- Created: 2026-05-24T07:44:25Z
- Updated: 2026-06-15T09:08:36Z
- Labels: bug, priority: high, release-1.1.4
- Assignee: Rachit-Gandhi

## Problem
Tool / function calling is silently dropped on native non-OpenAI paths:
- Anthropic / Gemini / Bedrock / Cohere request structs don't include `tools` / `tool_choice` (the Anthropic provider has no `Tools` reference at all).
- OpenAI streaming drops `tool_calls` deltas, and `buildMessages` drops assistant `ToolCalls` on multi-turn round-trips.

Agentic / function-calling requests come back text-only with no error.

## Fix
Forward `tools` / `tool_choice` per native API; assemble assistant tool-call round-trips (and stream tool-call deltas) correctly.

## Acceptance
Tool-call round-trip integration test per native provider (call → tool result → final answer).

_Internal audit → R-H1._
