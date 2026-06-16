---
name: trace-flow-graph
description: Builds draggable investigation trace graphs from a written flow or codebase search. Produces HTML (live edges + drag) and Obsidian .canvas files. Use when the user asks for a trace map, flow graph, investigation graph, architecture trace, or Obsidian canvas for a code path.
---

# Trace Flow Graph

Build **dual artifacts** for code-path / request-flow investigation:

1. **HTML graph** — draggable nodes, SVG edges that redraw on drag, pan canvas, layout saved to localStorage
2. **Obsidian canvas** — same nodes/edges, native drag + connections in Obsidian Canvas view
3. **Index note** (optional) — short markdown with wikilinks to the canvas

Default output directory: `.scratch/` in the repo (create if missing). Ask only if the user specifies another path.

## When to use

- User describes a flow in prose ("trace from router to provider")
- User asks to search the codebase and map a path
- User is investigating a bug and wants a living graph to extend layer by layer
- User mentions Obsidian canvas, trace map, or draggable flow graph

## Workflow

```
Task Progress:
- [ ] 1. Gather flow (written or search)
- [ ] 2. Draft node + edge list (internal schema)
- [ ] 3. Lay out nodes (spread columns, no overlap)
- [ ] 4. Write HTML graph
- [ ] 5. Write Obsidian .canvas
- [ ] 6. Write index .md (optional)
```

### Step 1 — Gather the flow

**Written flow:** User gives steps, function names, or layers. Confirm scope (how deep, streaming vs non-streaming, error paths).

**Codebase search:** Trace the real path. For each node record:
- Display title (function/type/route name)
- File path + line when known
- 3–6 bullet facts (calls, branches, key fields)
- Category (see legend below)

Read call sites, handlers, and type definitions — do not invent nodes.

### Step 2 — Internal schema

Build this list before generating files:

```yaml
meta:
  title: "Issue #141 — Chat Completions Trace"
  subtitle: "ingress → req fork"
  issue_ref: "#141"          # optional

nodes:
  - id: kebab-case-id          # stable, used in HTML id= and canvas id
    tag: router                # legend category
    title: "NewRouter(...)"
    file: "internal/httpserver/router.go:31"
    bullets:
      - "chi.NewRouter()"
      - "mountOpenAIRoutes(r, gw, ...)"
    highlight: false           # true for investigation focus fields
    x: 60
    y: 260
    width: 280                 # wide: 300, xwide: 400

edges:
  - from: bootstrap
    to: new-router
    type: flow                 # flow | decision | error | dep | issue
```

**Edge types**

| type | meaning | HTML class | Canvas color |
|------|---------|------------|--------------|
| `flow` | main execution path | (default) | none |
| `decision` | branch / fork | `decision` | `"5"` cyan |
| `error` | error / early return | `error` | `"1"` red |
| `dep` | dependency / side reference | `dep` | `"6"` purple |
| `issue` | bug-relevant data path | `issue` | `"3"` yellow |

**Node categories (tag → color)**

| tag | use for |
|-----|---------|
| `router` | HTTP routes, mounts, middleware |
| `handler` | Handler funcs, entrypoints |
| `gateway` | Orchestrators, facades, core structs |
| `decode` | Parsing, validation, DTO mapping |
| `request` | Domain request/object types |
| `stream` | Streaming path |
| `nonstream` | Non-streaming path |
| `error` | Error responses |
| `note` | Annotations (use `.note` class in HTML) |

Add custom tags to the legend if needed; keep colors consistent within one graph.

### Step 3 — Layout rules

- **Spread nodes** — use columns (~400px apart), ~180–220px vertical gap within a column
- Canvas size: at least `(maxX + 400) × (maxY + 300)`; default ~2200×1700 for medium graphs
- Put the **fork/branch** near the object that drives it
- Put **error nodes** above/right of the step that triggers them
- Put **dependency** nodes (shared structs) beside the route that closes over them
- Never stack all nodes in one pile — investigation graphs must be readable on first open

Obsidian canvas uses negative coordinates freely; center the graph around (-500, 0) for large layouts.

### Step 4 — HTML graph

Filename: `.scratch/<slug>-trace-graph.html` (slug from title, kebab-case).

Must include:
- Sticky header with title, legend, hint, **Reset layout** + **Fit to view**
- `#graph` container with `#edges-svg` **behind** `.draggable` nodes
- Each node: `id`, `class="draggable node"`, `data-x`, `data-y`, `.drag-handle` with grip
- `EDGES` array in script matching internal schema
- `redrawEdges()` on load, drag, and resize
- Pointer drag on handles; pan on empty viewport
- `localStorage` persist (bump key suffix when layout defaults change)

For full HTML/CSS/JS skeleton, see [reference.md](reference.md).

### Step 5 — Obsidian canvas

Filename: `.scratch/<Title Case Name>.canvas`

```json
{
  "nodes": [{ "id", "type": "text", "text", "x", "y", "width", "height", "color" }],
  "edges": [{ "id", "fromNode", "toNode", "fromSide", "toSide", "color?" }]
}
```

- `text`: markdown — `## title`, `` `file:line` ``, bullet list
- `fromSide` / `toSide`: `top` | `right` | `bottom` | `left`
- Reuse the **same node ids** as HTML
- Map edge `type` → canvas `color` per table above

### Step 6 — Index note (optional)

`.scratch/<Title Case Name> Index.md` — links to canvas, lists artifacts, one-paragraph checkpoint.

## Updating an existing graph

When the user adds the next layer ("now go into RouteStream"):

1. Read existing HTML + canvas + any index note
2. Append nodes/edges; do not rename existing ids
3. Place new nodes below/right of the frontier nodes they extend
4. Bump HTML `localStorage` key version so stale layouts reset cleanly
5. Update index checkpoint paragraph

## Investigation mode

If the user says **don't fix yet** / **root cause only**:
- Mark suspect fields with `highlight: true` and `issue` edges
- Add a `note` node for the checkpoint finding
- Do not propose code fixes in the graph

## Quality checklist

- [ ] Every edge `from`/`to` id exists as a node
- [ ] File paths verified against codebase (search mode)
- [ ] HTML edges redraw when nodes drag
- [ ] Obsidian canvas opens with connected nodes
- [ ] Initial layout has no overlapping nodes
- [ ] Legend matches tags used

## Additional resources

- HTML template, JS edge logic, Obsidian color codes: [reference.md](reference.md)
- Worked example (issue #141): [examples/](examples/) in this directory
