# AI Gateway

Gateway for routing and transforming LLM provider requests, including optional response caching.

## Language

### Request limits

**Completion token limit**:
The caller's requested cap on generated output tokens. `max_tokens` is the canonical internal limit for provider dispatch; `max_completion_tokens` is an OpenAI-compatible input form that fills `max_tokens` only when `max_tokens` is absent.
_Avoid_: Output cap, token budget (unless referring to broader accounting)

### Response cache

**Cache hit**:
A before-request lookup that finds a non-expired entry for the request’s cache key and returns the stored provider response without calling the provider.
_Avoid_: Hit, match (without “cache” when the distinction matters)

**TTL (time-to-live)**:
How long a cached response remains valid after it is stored (`max_age`). The clock starts (and resets) on **store** only—not on cache hit. A hot key can stay at the front of the LRU order but still expire once `max_age` has passed since its last store.
_Avoid_: Expiry window, max lifetime, sliding expiration (unless product explicitly adopts refresh-on-hit)

**Capacity eviction**:
Which entry is removed when the cache is full and a new key is stored. Eviction is by least-recently-used order among entries still in the cache; TTL does not pick the eviction victim.
_Avoid_: Earliest-expires eviction, TTL-based eviction (for capacity pressure)

**Disabled response cache**:
Configuration with `max_entries` zero (or negative): the plugin never stores responses and every lookup is a miss.
_Avoid_: Unlimited cache, capacity zero meaning unbounded

### Circuit breaker

**Circuit breaker**:
Per-target guard that stops calling a provider while it is considered unhealthy, so failures fail fast instead of hammering a down or overloaded backend.
_Avoid_: Retry policy, fallback (those are routing strategies; the breaker is a per-provider call gate)

**Circuit state (closed / open / half-open)**:
Closed — calls pass through normally. Open — calls are rejected immediately with `circuit breaker open`. Half-open — a recovery probe window after the open timeout; only probe calls should reach the provider.
_Avoid_: Half-open as “fully open again”

**Probe (half-open)**:
A single call allowed through while the circuit is half-open to test whether the provider has recovered. Issue #152: the implementation currently allows unlimited concurrent probes.
_Avoid_: Any request during half-open (only designated probes should pass)
