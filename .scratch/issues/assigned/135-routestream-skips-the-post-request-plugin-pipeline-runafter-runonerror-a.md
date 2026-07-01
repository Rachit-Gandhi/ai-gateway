# Issue #135: RouteStream skips the post-request plugin pipeline (RunAfter/RunOnError) and cache hits

- URL: https://github.com/ferro-labs/ai-gateway/issues/135
- State: CLOSED
- Created: 2026-05-24T07:43:52Z
- Updated: 2026-06-15T09:00:19Z
- Labels: bug, priority: high, release-1.1.3
- Assignee: Rachit-Gandhi

## Problem
`RouteStream` runs only `RunBefore`. It never calls `RunAfter` / `RunOnError`, and it ignores `pctx.Skip`. `Route` handles all of these via `runBeforePlugins` + the after/error stages. So for **all streaming traffic**:
- response-cache never stores responses,
- request-logger never logs,
- `on_error` plugins never fire,
- a cache `Skip` hit is discarded and the provider is called anyway.

## Fix
Run the full plugin pipeline in `RouteStream`: handle `pctx.Skip` (convert a cached response into a single-chunk stream), and invoke `RunAfter` / `RunOnError` from the goroutine that drains the stream.

## Acceptance
Integration tests: streaming request triggers `RunAfter` (cache store + request log) and `RunOnError`; a cached entry short-circuits a streaming request.

_Internal audit → H4 / R-C3 (pipeline). Note: the budget-enforcement consequence is tracked separately._
