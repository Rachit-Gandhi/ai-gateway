# Issue 141 Chat Completions Trace

Investigation trace map for issue #141 (`max_completion_tokens` ignored by non-OpenAI providers).

Open the canvas board: [[Issue 141 Chat Completions Trace.canvas]]

## Flow

Router ingress → handler → decode → validate → `providers.Request` → stream fork.

## Artifacts

| File | Purpose |
|------|---------|
| `.scratch/issue-141-trace-graph.html` | Draggable HTML graph with live edges |
| `.scratch/Issue 141 Chat Completions Trace.canvas` | Obsidian canvas (native edges + drag) |

## Checkpoint

Both `max_tokens` and `max_completion_tokens` survive decode + validation. The drop happens in per-provider mapping (next layer).
