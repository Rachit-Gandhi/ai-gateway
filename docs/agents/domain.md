# Domain Docs

How the engineering skills should consume this repo's domain documentation when exploring the codebase.

## Layout

This is a single-context repository.

The expected domain-doc locations are:

- `CONTEXT.md` at the repo root for shared domain language and glossary terms.
- `docs/adr/` for architecture decision records.

## Before exploring, read these

- `CONTEXT.md` at the repo root, if it exists.
- ADRs under `docs/adr/` that touch the area being changed, if they exist.

If these files do not exist, proceed silently. Do not flag their absence or suggest creating them upfront. Producer skills create them lazily when terms or decisions actually get resolved.

## File structure

```text
/
├── CONTEXT.md
├── docs/adr/
│   ├── 0001-example-decision.md
│   └── 0002-example-decision.md
└── ...
```

## Use the glossary's vocabulary

When output names a domain concept, use the term as defined in `CONTEXT.md`. Do this in issue titles, refactor proposals, hypotheses, and test names.

If the concept is missing from the glossary, either reconsider whether the term belongs to this project or note the gap for a future documentation pass.

## Flag ADR conflicts

If output contradicts an existing ADR, surface the conflict explicitly rather than silently overriding it.
