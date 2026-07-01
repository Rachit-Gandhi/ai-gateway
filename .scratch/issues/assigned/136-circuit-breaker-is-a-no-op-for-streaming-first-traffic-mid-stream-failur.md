# Issue #136: Circuit breaker is a no-op for streaming-first traffic; mid-stream failures don't count

- URL: https://github.com/ferro-labs/ai-gateway/issues/136
- State: CLOSED
- Created: 2026-05-24T07:43:53Z
- Updated: 2026-06-15T09:00:20Z
- Labels: bug, priority: high, release-1.1.3
- Assignee: Rachit-Gandhi

## Problem
Two streaming circuit-breaker gaps:
1. `circuitBreakers` is populated only inside `getStrategy`, which `RouteStream` never calls — so for streaming-first traffic the breaker map is empty and the breaker is a **no-op**.
2. `cbProvider.CompleteStream` records **success when the stream opens**, so mid-stream failures never trip the breaker.

## Fix
Initialize breakers independently of `getStrategy` (call from both routes / `New`); record success/failure from `streamwrap.Meter` at stream completion (callback in `MeterMeta`).

## Acceptance
Tests: breaker trips on a streaming-only failing provider; a provider that always errors mid-stream opens the circuit.

_Internal audit → H1/H2._
