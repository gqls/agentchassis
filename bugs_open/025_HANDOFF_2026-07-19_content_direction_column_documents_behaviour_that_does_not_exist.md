# 025 — `pages.content_direction` documents per-page writer steering that does not exist

**Filed 2026-07-19** (relojistas thread). **Status: OPEN.** Fleet-wide. Not a runtime
failure — a *false affordance*: the column's own comment promises a capability nothing
implements, so a thread that reads the schema plans against it and loses the time.

## Symptom

`pages.content_direction` (jsonb) exists and its `COMMENT ON COLUMN` states:

> `Optional per-page content direction for rebuilds. Passed to content-writer prompt.`
> `Structure: { "instruction": "...", "format": "...", "examples": [...], "avoid": [...] }`

**No Go code reads the column.** Nothing writes it either. Setting it has no effect on
generated content whatsoever, and there is no error — the value is silently inert.

## Evidence

Added by `docs/agent_docs/sql_for_tables/005_content_components.sql:6313`; comment at
`:6315-6317`.

Neither action that loads a page row for the build path selects it:
- `platform/orchestration/actions/get_pages_to_build_actions.go:98-104` and `:120-127` —
  selects `id, site_id, name, url, title, page_type, status, build_status, sections,
  nav_label, nav_order, in_header, in_footer, version, meta_description`.
- `platform/orchestration/actions/load_page_record_action.go:167` — `SELECT id::text, name,
  title, page_type, sections::text, url, build_status, nav_label, nav_order FROM pages`.

Every `content_direction` reference in Go is the **site-level `site_specs` aspect**, a
different thing that happens to share the name — `site_spec_actions.go:210`,
`apply_adoption_plan_action.go:251,285`, `write_site_plan_action.go:126`,
`datahelpers/format_content_direction.go:22`. The writer prompt only ever dereferences
`.site_specs.specs.content_direction.formatted`, never `.current_page.content_direction`.

## Why it matters

The name collision is the trap. A thread searching for per-page steering finds a
`content_direction` column *and* finds `content_direction` all over the Go source, and
reasonably concludes the column is wired up. It is not; the Go hits are the site-wide spec.
This thread planned a per-page cite-or-omit rule against the column before checking, and
only caught it by grepping the two loaders for the column name.

## Fix candidates

1. **Delete the column and its comment.** Cheapest and most honest if per-page steering is
   not wanted. Nothing reads it, so removal is behaviour-preserving.
2. **Wire it up as documented** — add `content_direction` to the two `SELECT`s above, thread
   it into `page-content-writer`'s `input_fields` (currently `current_section`,
   `render_context`, `reviewed_brief`, `current_page`, `link_context`, `site_plan`,
   `site_specs`, `existing_content`, `build_mode`, `rewrite_guidance`), and render it in
   the prompt alongside the existing site-level block.
3. **Correct the comment to say "reserved, not implemented"** — the stopgap if neither of
   the above is scheduled. Strictly better than the current state, which actively misleads.

Recommend (1) unless someone wants per-page steering: the live per-section hook
`.current_section.component.content_brief` (`purpose` / `tone_direction` /
`section_guidance`) already provides finer-grained steering than the column would, and it
demonstrably reaches the prompt.

## How to verify a fix

- For (1): column absent from `\d pages`; `grep -rn "content_direction" platform/ internal/ pkg/`
  returns only site-spec hits.
- For (2): set `pages.content_direction` on one page, rebuild it, and confirm the
  instruction's effect in the generated section text — **verify against the rendered
  artifact or the saved section, not a `complete` status.**

## Related

Same family as the general lesson in `016b §9`: *the repo SQL under `sql_for_agents/` is a
migration history, not the live truth* — live prompts live in
`agent_definitions.default_config`. Here the inverse bites: a `sql_for_tables/` comment is
the only documentation of a capability, and it is wrong.

## Addendum 2026-07-20 — the false claim also lives in the canonical contracts doc

`003_contracts_and_standards(8).md` §"content_direction (Page-Level Edit Instructions)"
states "Flows to content-writer's prompt when present" — the same claim as the column
comment, in the doc every developer is told to check changes against. Corrected in place
(dated block pointing here) rather than deleted, since it may be the *intended* design.
This strengthens fix candidate (2) (wire it up as documented): the behaviour is documented
in two places, so either the code should match the docs or both docs should change, not one.

## Fix in progress 2026-07-21 (bugfix 025 thread) — candidate (2), owner-chosen

**Owner chose "wire it up as documented", not delete.** Status stays **OPEN** until
the Go half is fixed AND live (inert until the next chassis image roll).

**Correction to the handoff's data-flow model.** The handoff listed both loaders but
implied `current_page` for the rebuild path comes from the work-item `spec` (as seed
`055_page_build_handler.sql:85` shows: `"current_page": "input_data.spec"`). **The LIVE
`page-build-handler` does not** — it maps `"current_page": "page_record"`, the fresh
output of `load_page_record`. So the load-bearing loader for the acceptance test
("rebuild the page") is **`load_page_record_action.go`**, not the spec. Verified by
dumping the live `page-build-handler` default_config; the seed is migration history.
Full path table in `docs024_key_docs_latest/bug025_content_direction/PLAN_2026-07-21_*`.
Both loaders were wired anyway (pageflow-builder / page-rebuild loop over
`get_pages_to_build`; site-work-orchestrator reads the spec).

**Done (committed):**
- Go: `load_page_record_action.go` and `get_pages_to_build_actions.go` now SELECT
  `content_direction::text` and put the parsed value on the page map (only when
  present, so the writer guard stays false otherwise). Compiles + gofmt clean.
- Config (LIVE now): migration `sql_for_agents/187_page_content_direction_wireup.sql`
  renders `.current_page.content_direction` in the `page-content-writer` prompt
  (guarded block, `{ instruction, format, examples, avoid }`), and corrects the
  column COMMENT. Applied to live DB (UPDATE 1). Full-template parse+execute verified
  under `text/template` for absent / full / partial cases — block emits zero
  `<no value>` when the key is absent.
- Docs: `003_contracts_and_standards(8).md` §content_direction updated to IMPLEMENTED.

**Remaining to CLOSE:** chassis image build + roll (Go half), then the acceptance
test — set `pages.content_direction` on one real page, rebuild, confirm the
instruction's effect in the **saved section**, not a `complete` status.
