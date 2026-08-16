# HANDOFF — webdesign tool rebuilds. START HERE. Written 2026-08-16 ~10:15Z, UPDATED 15:40Z. Supersedes `HANDOFF_2026-08-15_continue_here.md`.

Read `PLAN_2026-08-15_…` (design), `RUNBOOK_…` (commands, incl. the new "Is the adopt path live?" block),
`NOTES_…` (evidence + missteps, newest at the bottom), `bugs_open/286` (the pilot's real cause + fix).

## Verified state (2026-08-16 ~15:40Z — supersedes the 10:10Z table)

| thing | state |
|---|---|
| fleet binary | **`v1.0.1304`**, pods 10:41Z, stamp `5de6cddbe`; `88897190e` IS an ancestor (probed with a junk-hex control) ⇒ **286 fix LIVE**. RUNBOOK: probe the STAMP, then ancestry — never grep your own sha. |
| seed 435 | **APPLIED 15:15Z** (`ee0228813`): `tool-generator.save_tool.adopt_existing_page = true` (SELECT-verified). |
| pilot | **REFILED 15:2xZ** as item `99734862-ed24-4841-9d8f-7e6ce8c1de6b` (`add_tool_novel_webdesign.co.uk`, `triaged`), 11 items ahead in the serial dispatcher. Description written from the live tool, self-contained copy demanded. |
| seed 434 proof | 3 re-armed audit fixes **graded PASS at the served pages** (0 raw tags, real copy, inputs/scripts intact). |
| ab-test | REVERTED to the ported tool + **verified served** (0 raw tags, `class="ported-page"` present, fork gone). Rebuild candidate #2. |
| 285 | owning lane ran the induce-a-refusal 09:59Z (fence refused, wrapper intact) — theirs. |
| aspect-ratio page `00979b9e…` | ported slot `deployed`, no native slot yet. |

## Next actions, in order

1. **Grade the pilot's generator run** (item `99734862`): `SELECT status, error FROM site_work_items WHERE id='99734862-…'`; then the tool-generator orchestration for it — `create_result.page_adopted` must be `true`; `page_components` on page `00979b9e-db47-4c26-819e-add95b0f8fd6` must gain ONE row (`slot_name='tool-aspect-ratio'`, position 2) and `pages` NO new row for the site with name `tool-aspect-ratio`. If it collided again ⇒ the flag is not being read: re-check the SELECT and `agent_error_log` for `save_tool`.
2. **Grade the generated component itself BEFORE retiring anything** (the ab-test lesson): `content_components` row for `function='tool-aspect-ratio'` — `regexp_matches(html_template,'\{\{\.','g')` count must be 0 (no externalised copy fields), visible-chars > 300, has `<script>`, no `<script src=`.
3. Retire the ported slot on that page (RUNBOOK guarded UPDATE, exactly 1 row) → assemble-only `page_rerender` (spec `{domain,page_id,page_name,filename}`, no `reason`) → grade at the served URL: `{{\.` 0, `class="ported-page"` 0, the new tool's container present, inputs/buttons present.
4. Only after 1–3 succeed ONCE: batch (PLAN §2), simple tools first, serial, ab-test second — **CONFIRMED from code (`create_tool_component_action.go:197-206`): the already-exists probe joins `page_components` with NO `build_status` filter, so the hollow fork `cd60486c` (active, linked `removed`) WILL short-circuit the generator into `already_exists` and write nothing.** Before filing ab-test: `UPDATE content_components SET is_active=false, updated_at=now() WHERE id='cd60486c-f5e1-4d80-9676-0d65024f0372' AND function='tool-ab-test-calculator' AND is_active;` (1 row; the removed placement row stays as history), then file.
5. Owed to others, unchanged (122 ink lane's audit checks ~08-18; mindmap junk text = owner's localStorage).

## Traps this session paid for (all in LANDMINES / WRONG_CALLS / NOTES)

- An empty `final_result` is the workflow's `output_fields` shape, not a failure; read `agent_error_log`
  for the window and check whether the thing EXISTS before naming a mechanism.
- `spec_data` as a MAP is silently unread — census `jsonb_typeof` on any copied `create_work_item` step.
- A tool slot can be 13 KB and have zero visible text; the class-attr floors will not tell you.
- Guard "open items on this page" on the dispatchable statuses only; `unresolved`/`failed` pile up.
- The 4 audit_fix items I re-armed were the SAME class as the 233 dead ones — when a "two-item cleanup"
  turns out to be a class, say so in the seed and the bug file, and fix the producer, not the rows.

## OWNER RULING 2026-08-16 (added ~15:45Z by the bugfix_285 lane) — PLAN §3 REVERSED

**The rich hand-built apps ARE rebuild candidates after all** (Mind Map Studio Pro, Meme Studio,
Logic Architect Pro, Flat-File Micro CMS, Pasteboard Manager, and the 13 external-script tools):
the owner chose option (a) — generator rebuild anyway, **accepting that it is a reimplementation,
not a preservation**. The trade PLAN §3 warned about was put to him in those words and accepted.

Consequences for this lane's next actions:
- **No excluded class remains** — all 63 ported tools are in scope for the native route.
- **`bugs_open/204` / the byte-faithful decomposition route is no longer a prerequisite for ANY
  webdesign tool.** (204 still matters for other decomposed sites; a session picked it up 15:33Z.)
- Unchanged and now load-bearing: spec written from the LIVE tool's behaviour; grade the generated
  component BEFORE retiring the ported slot (ab-test: 13 KB, zero visible text, 47 raw tags);
  retire with `build_status='removed'`, never delete — **note the `page_component_history` archive
  row id per tool in NOTES**, that is what makes a bad reimplementation a one-statement revert.
- RECOMMENDED (yours to take or leave): rich apps LAST and one at a time, after the simple batch
  proves the recipe, each seen by the owner at the served page.

Full ruling + reasoning: PLAN §"OWNER RULING 2026-08-16", README_where_we_are 2026-08-16 afternoon.
