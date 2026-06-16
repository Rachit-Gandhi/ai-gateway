# Trace Flow Graph — Reference

## Obsidian canvas color codes

| code | color | use |
|------|-------|-----|
| `"1"` | red | errors |
| `"2"` | orange | handlers, non-stream |
| `"3"` | yellow | issue focus, notes |
| `"4"` | green | router / ingress |
| `"5"` | cyan | decode, stream, decision |
| `"6"` | purple | gateway, dependencies |

## Node HTML template

```html
<div class="draggable node" id="NODE_ID" data-x="X" data-y="Y">
  <div class="drag-handle"><span class="tag TAG">Label</span><span class="grip">⠿</span></div>
  <div class="node-body">
    <h3>TITLE</h3>
    <div class="file">path/to/file.go:LINE</div>
    <ul>
      <li>fact one</li>
    </ul>
  </div>
</div>
```

Classes: `node wide` (300px), `node xwide` (400px), `node issue-highlight` (red border).

Fork pill: `class="draggable fork-label"`.

Annotation: `class="draggable note"` with drag-handle + `note-body`.

## EDGES JavaScript

```javascript
const EDGES = [
  { from: 'node-a', to: 'node-b' },                    // flow
  { from: 'node-a', to: 'node-b', type: 'decision' },
  { from: 'node-a', to: 'node-b', type: 'error' },
  { from: 'node-a', to: 'node-b', type: 'dep' },
  { from: 'node-a', to: 'node-b', type: 'issue' },
];

function pickSides(fromRect, toRect) {
  const dx = (toRect.x + toRect.w/2) - (fromRect.x + fromRect.w/2);
  const dy = (toRect.y + toRect.h/2) - (fromRect.y + fromRect.h/2);
  if (Math.abs(dx) > Math.abs(dy))
    return dx > 0 ? { from: 'right', to: 'left' } : { from: 'left', to: 'right' };
  return dy > 0 ? { from: 'bottom', to: 'top' } : { from: 'top', to: 'bottom' };
}

function anchor(rect, side) {
  switch (side) {
    case 'top': return { x: rect.x + rect.w/2, y: rect.y };
    case 'bottom': return { x: rect.x + rect.w/2, y: rect.y + rect.h };
    case 'left': return { x: rect.x, y: rect.y + rect.h/2 };
    case 'right': return { x: rect.x + rect.w, y: rect.y + rect.h/2 };
  }
}

function bezierPath(p1, p2, fromSide, toSide) {
  const dist = Math.hypot(p2.x - p1.x, p2.y - p1.y);
  const offset = Math.min(120, Math.max(40, dist * 0.35));
  const cp1 = { ...p1 }, cp2 = { ...p2 };
  if (fromSide === 'bottom') cp1.y += offset;
  if (fromSide === 'top') cp1.y -= offset;
  if (fromSide === 'right') cp1.x += offset;
  if (fromSide === 'left') cp1.x -= offset;
  if (toSide === 'bottom') cp2.y += offset;
  if (toSide === 'top') cp2.y -= offset;
  if (toSide === 'right') cp2.x += offset;
  if (toSide === 'left') cp2.x -= offset;
  return `M ${p1.x} ${p1.y} C ${cp1.x} ${cp1.y}, ${cp2.x} ${cp2.y}, ${p2.x} ${p2.y}`;
}

function redrawEdges() {
  edgesGroup.innerHTML = '';
  EDGES.forEach(({ from, to, type }) => {
    const fromEl = document.getElementById(from);
    const toEl = document.getElementById(to);
    if (!fromEl || !toEl) return;
    const fromRect = { x: +fromEl.dataset.x, y: +fromEl.dataset.y, w: fromEl.offsetWidth, h: fromEl.offsetHeight };
    const toRect = { x: +toEl.dataset.x, y: +toEl.dataset.y, w: toEl.offsetWidth, h: toEl.offsetHeight };
    const sides = pickSides(fromRect, toRect);
    const p1 = anchor(fromRect, sides.from);
    const p2 = anchor(toRect, sides.to);
    const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
    path.setAttribute('d', bezierPath(p1, p2, sides.from, sides.to));
    path.setAttribute('class', 'edge' + (type ? ' ' + type : ''));
    edgesGroup.appendChild(path);
  });
}
```

Call `redrawEdges()` after every position change and on `loadPositions()`.

## SVG edge CSS classes

```css
.edge { stroke: #4a5d78; stroke-width: 2; fill: none; marker-end: url(#arrow); }
.edge.decision { stroke-dasharray: 7 5; stroke: #6ea8fe; marker-end: url(#arrow-blue); }
.edge.error { stroke-dasharray: 5 4; stroke: #f85149; marker-end: url(#arrow-red); }
.edge.dep { stroke-dasharray: 3 4; stroke: #a371f7; opacity: 0.85; marker-end: url(#arrow-purple); }
.edge.issue { stroke: #ff7b72; stroke-width: 2.5; marker-end: url(#arrow-issue); }
```

Define matching SVG `<marker>` ids for each arrow color in `<defs>`.

## Drag + persist essentials

```javascript
const STORAGE_KEY = '<slug>-trace-graph-positions-v1';

function applyPosition(el, x, y) {
  el.style.left = x + 'px';
  el.style.top = y + 'px';
  el.dataset.x = String(x);
  el.dataset.y = String(y);
  redrawEdges();
}

// Drag: pointerdown on .drag-handle or .fork-label → pointermove → pointerup + savePositions()
// Pan: pointerdown on .viewport (not .draggable) → scroll viewport
```

## Obsidian canvas node example

```json
{
  "id": "chat-completions",
  "type": "text",
  "text": "## ChatCompletions(gw)\n`internal/handler/chat.go:14`\n\n- Decode\n- Validate\n- Branch on Stream",
  "x": -500,
  "y": 160,
  "width": 300,
  "height": 200,
  "color": "2"
}
```

## Obsidian canvas edge example

```json
{
  "id": "e7",
  "fromNode": "chat-completions",
  "fromSide": "right",
  "toNode": "decode",
  "toSide": "left"
}
```

Add `"color": "1"` for error edges, `"6"` for dep edges.

## Column layout starting points (HTML data-x/y)

| column | x | contents |
|--------|---|----------|
| ingress | 60 | bootstrap → router → middleware → mount → route |
| handler | 480 | handler func |
| decode | 880 | decode, validate |
| meta | 1380 | error, wire DTO |
| object | 520 | domain object (lower, y≈1180) |
| fork | 120 / 980 | stream left, non-stream right (y≈1540) |

Adjust per graph; keep ≥400px horizontal between columns.

## Codebase search hints

When tracing a request path, search in order:

1. Route registration (`Post(`, `HandleFunc(`, `Mount(`)
2. Handler function body (decode → validate → branch)
3. Core orchestrator (`Route`, `Execute`, `Process`)
4. Provider/client mapping (where fields are read or dropped)

Grep for struct field names when investigating "field X ignored" bugs.
