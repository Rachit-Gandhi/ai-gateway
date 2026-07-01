# Issue #152: Cache & breaker correctness: cache key omits logprobs, eviction not LRU, half-open probe cap

- URL: https://github.com/ferro-labs/ai-gateway/issues/152
- State: OPEN
- Created: 2026-05-24T07:45:24Z
- Updated: 2026-06-08T07:12:05Z
- Labels: bug, priority: medium, release-1.1.6
- Assignee: Rachit-Gandhi

## Problem
Response-cache correctness issues:
- **Cache key omits `logprobs` / `top_logprobs`** — a request with `logprobs=true` can receive a cached response that lacks log-probs.
- **Eviction is by earliest `expiresAt`, not LRU** — hot entries with short TTLs are evicted before cold entries with long TTLs.
- **Circuit-breaker half-open allows unlimited concurrent probes** — thundering herd on recovery.

## Fix
Include the missing fields in the cache key; switch eviction to LRU (or reuse `internal/cache`); cap half-open probes.

_Internal audit → R-M1(cache)/cache eviction/half-open._
