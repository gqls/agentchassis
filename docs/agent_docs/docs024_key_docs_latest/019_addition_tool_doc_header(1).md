# 019 addition — Tool Doc Header (splice as a new section after "Design conventions")

> Splice target: `019_tool_library.md`, new `## Tool Doc Header` section
> following `### Design conventions`. Also add one row to 020's Go Actions /
> Discovery Checks tables as noted at the end. This file is the staging copy;
> 019 is the canonical home once spliced.

## Tool Doc Header

Every NEW tool's `html_template` begins its `<script>` with ONE sentinel-
delimited block comment — the tool-doc header. It states what the tool is and
the invariants it must keep, so the auditor reviews against a spec, rewrites
preserve intent, and the future JS parser has a doc to index.

**It never ships.** `StripToolDocHeader` (platform/content/tool_doc_header.go)
removes the block at the OUTBOUND deploy assembly — three call sites:
`RerenderSinglePageAction`'s page HTML, `collectJSAssets`' emitted `.js`
values, and the bulk path's `finalHTML` in `rerender_pages_actions.go` —
so the public page and assets carry none of it. `rendered_html` in the DB
deliberately RETAINS the header (audit/parse parity with the template).
That's why it can be richer than a one-liner — and why general
comment-stripping is NOT used: removing arbitrary JS comments safely needs a
JS-aware minifier (a regex destroys `"https://…"` strings and regex literals);
exact-sentinel scanning cannot corrupt code and is idempotent. A malformed
block (opener without closer) is deliberately left in place — shipping an inert
comment is harmless, truncating a script is not — and `tool_health` flags it.

### Format

The sentinels are exact byte sequences (`ToolDocOpen` / `ToolDocClose` consts):

```html
<script>
/* === tool-doc ===
function: tool-vat-calculator
purpose: One sentence — what the tool computes/produces for the visitor.
behaviour:
  - the invariants an audit must hold the tool to (units, ranges, rounding)
  - no external calls; all computation client-side
inputs: what the user provides (fields, units)
outputs: what the tool renders (results, generated code, downloads)
=== /tool-doc === */
(function() {
    // tool logic
})();
</script>
```

Rules:
- ONE block, first thing inside `<script>`. 6–12 lines.
- `function:` matches the component's `function` column exactly.
- NEVER include orchestration ids, agent names, or dates — provenance lives in
  the `source_*` columns, not in code (the header would ship if stripping ever
  regressed; internals are not published).
- The block must not contain `*/` (it is a JS block comment).
- The SUBSTANTIVE documentation is not here: the same generation step writes
  the full prose to the `description` column and the `knowledge_base`
  `tool_docs` row. The header is the audit/parse anchor only.

### Lifecycle

| Stage | Responsibility |
|---|---|
| Creation (`tool-generator`) | Prompt requires the header (snippet below); `create_tool_component` validates `HasToolDocHeader` before insert and rejects with a clear error otherwise. |
| Fork (`deploy_tool_to_site`) | Forks copy the template, header included — no action needed. |
| Modification (`tool-improver`, `update_component_html`) | Prompts instruct: preserve the header; update `behaviour:` lines when behaviour changes. The header is the anti-drift anchor across LLM rewrites. |
| Audit (`tool-auditor`) | Reads `html_template` (DB), audits the code AGAINST the stated `behaviour:` invariants — "does the code match its own spec". |
| Render/deploy | `StripToolDocHeader` runs at deploy assembly — `RerenderSinglePageAction` page HTML, `collectJSAssets` `.js` values, bulk `finalHTML` (cheap no-op when absent). `rendered_html` retains the header; `tool_health`'s `no_script` check is unaffected — only the comment block is removed from what ships. |
| Sweep (`tool_health`) | `no_doc_header` (warning) on tools missing the header; `malformed_doc_header` (error) on opener-without-closer. Old tools converge on the normal sweep — no retrofit campaign. |

### tool-generator prompt snippet (splice into its generation prompt)

> Begin the `<script>` with exactly one tool-doc header block in this form:
> `/* === tool-doc ===` … `=== /tool-doc === */` containing `function:`
> (matching the component function), `purpose:` (one sentence), `behaviour:`
> (the invariants your code keeps — units, ranges, no external calls),
> `inputs:`, `outputs:`. Do not put names, ids, or dates in it. Do not use
> `*/` inside it.

Apply via the usual agent-definition discipline: pull the row, LOCATE the
nested prompt path first (the 072 bug — prompts sit inside loop
`sub_workflow`s, not top-level steps), `snapshot_agent('tool-generator')`,
then the targeted `jsonb_set`. Same for the one preserve-the-header line in
`tool-improver`.

### 020 table additions

- Go Actions: `StripToolDocHeader` / `HasToolDocHeader` — `tool_doc_header.go`
  — strip the doc header at render; validate presence at creation.
- Discovery Checks (`tool_health` tiers): `no_doc_header` (warning),
  `malformed_doc_header` (error) — header presence/shape on library tools and
  forks.
