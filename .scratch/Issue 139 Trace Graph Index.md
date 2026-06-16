# Issue #139 — Tool / Function Calling Trace

**GitHub:** [#139 — Tool/function calling is dropped on native non-OpenAI providers](https://github.com/ferro-labs/ai-gateway/issues/139)

## Artifacts

| File | Purpose |
|------|---------|
| [[Issue 139 Trace Graph.canvas]] | Obsidian canvas — drag nodes, native connections |
| [issue-139-trace-graph.html](./issue-139-trace-graph.html) | Interactive HTML graph — pan, drag, layout persisted in localStorage |

Open the HTML file in a browser for the draggable investigation graph. Use **Reset layout** if nodes look misplaced after updates.

## Checkpoint (investigation only)

Tool metadata (`tools`, `tool_choice`, `tool_calls`, `tool_call_id`) is preserved through HTTP decode (`internal/handler/chatrequest.go`) and gateway orchestration (`Gateway.Route` / `Gateway.RouteStream`). The silent failure happens at the **provider translation layer**: native providers (anthropic, gemini, bedrock, cohere) use request structs without tools; ~20 slim OpenAI-compat providers omit top-level `tools`/`tool_choice`; OpenAI streaming drops assistant `ToolCalls` in `buildMessages` and omits `tool_calls` in delta mapping. `streamwrap` and `sse.Write` are not drop points. The MCP agentic loop never runs when `resp.Choices[].Message.ToolCalls` is empty.

**Probe order:** after decode → before provider call → provider outbound HTTP body → raw API response → `StreamChunk.Delta.ToolCalls` before SSE.

## Node index

- **Ingress:** `client-post` → `chat-handler` → `decode-request` → `validate-request`
- **Fork:** `stream-fork` → `route-ns` (non-stream) | `route-stream` (stream)
- **Non-stream:** plugins → MCP inject → `strategy-execute` → `provider-complete` → MCP loop → JSON
- **Stream:** MCP redirect → `resolve-stream-provider` → `provider-complete-stream` → `streamwrap-meter` → `sse-write`
- **Providers (drop zone):** `openai-complete` ✓ · `openai-stream` ⚠ · `ollama-cloud` partial · `slim-oai-compat` ✗ · `anthropic-native` · `gemini-native` · `bedrock-native` · `cohere-native`
- **Focus:** `build-messages` · `issue-checkpoint`
