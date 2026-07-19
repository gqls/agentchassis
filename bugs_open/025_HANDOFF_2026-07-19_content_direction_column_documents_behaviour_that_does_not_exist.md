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
