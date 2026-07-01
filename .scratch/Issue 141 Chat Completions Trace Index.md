# Issue 141 Chat Completions Trace

Investigation trace map for issue #141 (`max_completion_tokens` ignored by non-OpenAI providers).

Open the canvas board: [[Issue 141 Chat Completions Trace.canvas]]

## Flow

Router ingress -> handler -> decode -> validate -> `providers.Request` -> stream fork -> `Route` / `RouteStream` -> strategy/provider resolution -> provider payload mapping.

## Artifacts

| File | Purpose |
|------|---------|
| `.scratch/issue-141-trace-graph.html` | Draggable HTML graph with live edges |
| `.scratch/Issue 141 Chat Completions Trace.canvas` | Obsidian canvas (native edges + drag) |

## Checkpoint

Both `max_tokens` and `max_completion_tokens` survive decode + validation and enter `Route` / `RouteStream` unchanged. A normalization before provider resolution would be provider-blind: it can coalesce `MaxCompletionTokens` into `MaxTokens` globally, but it cannot know "non-OpenAI only" for fallback/load-balance/cost strategies. The first non-streaming provider-aware seam is inside strategy execution after `lookup(...)`; the streaming path becomes provider-aware after `resolveStreamingProviderLocked(...)`. The current drop still happens in provider payload mapping where non-OpenAI adapters read `req.MaxTokens` only.
