# 286 — the tool-generator cannot rebuild a tool at an existing page: `create_tool_component` has no attach path, so a same-URL replacement dies on `pages_site_id_name_key` and deletes its own component

**Filed 2026-08-16 by the `webdesign_tool_rebuilds` lane. Status: FIX BUILT + council APPROVED (`27d0f428`, 2026-08-15 23:14Z), NOT LIVE — Go rides the next chassis roll; the config seed that arms it is HOLD until then.**
Diagnosis: 090 run corr `3050effc-f882-4e9b-b5fa-c4b6a247d672`, **CONFIRMED first iteration** (verdict in `orchestration_states.collected_data->'verdict'`; code-tier run, no `doc_notes` row by design).

## Symptom

`add_tool_novel_webdesign.co.uk` (pilot: rebuild the ported `tool-aspect-ratio` natively at the SAME URL, PLAN_2026-08-15) ran the tool-generator for 47 s at 2026-08-15 18:28–18:29Z and produced nothing: orchestration `5ef53886…` `COMPLETED`/`complete_error`, `final_result` empty, no `tool-aspect-ratio` in `content_components`. The work item read `complete`. Only `agent_error_log` says why:

```
2026-08-15 18:29:17Z tool-generator | save_tool | create_tool_component
step save_tool failed: failed to execute action create_tool_component: failed to create tool page:
ERROR: duplicate key value violates unique constraint "pages_site_id_name_key" (SQLSTATE 23505)
```

The two OTHER "empty `final_result`" generator runs the same evening (`tool-storage-risk-explainer` 18:36, `tool-overpayment-priority` 19:27, other sites) were NOT this: both created their components (`content_components.created_at` matches) — `final_result` empty is that agent's normal shape, not a handshake failure. **The handoff's "matches the spawn→call handshake failure" hypothesis was WRONG; recorded in the lane NOTES.**

## Mechanism (read at HEAD, cited by the 090 run)

`platform/orchestration/actions/create_tool_component_action.go`, `CreateToolComponentAction`:
- page name comes from `CanonicalisePage{Role:"tool", Slug:function}` ⇒ always `tool-<bare>` == `function`;
- the ONLY early exit is a COMPONENT probe (`content_components.function` + `component_level='tool'` + site) — a page probe does not exist;
- the page write is a bare `INSERT INTO pages (…) VALUES (…)` — no `ON CONFLICT`, no lookup, no `page_id` input anywhere in `CreateToolComponentInputSpec` (`site_id, html_content, function, display_name` + `description, category, related_pages`);
- the error branch runs `DELETE FROM content_components WHERE id = $1` on the row it just inserted.

So: existing page named `tool-<bare>` (canonical — every page this action or the deployer created since 08-04) ⇒ unique-key collision, component destroyed. Existing page under a LEGACY bare name ⇒ no collision and a SECOND page row (bugs_open/080's duplicate). Neither is an attach.

**Why the ab-test build "worked" the night before:** it did not go through this action. `tool-ab-test-calculator`'s native fork arrived via **`tool-deployer` / `deploy_tool_to_site`** (item `add_tool_webdesign.co.uk`, complete 00:08:28Z), whose page path is `resolveToolPageIdentity` (existing row keeps its identity) + `UpsertPageForRole` (`ON CONFLICT (site_id,name) DO NOTHING` → explicit role branches) + `page_components` insert at position 2. Same binary, different action. No change to `create_tool_component_action.go` shipped between 2026-08-11 and the failure (`git log` vs roll sha `5e075a6f9`).

## Fix (built 2026-08-16, `create_tool_component_action.go` + `create_tool_component_adopt_test.go`; register **TL-044**)

Opt-in step-config key **`adopt_existing_page`** (bool, default **OFF** — 2026-08-02 §2 shape). ON ⇒ identity via `resolveToolPageIdentity`, row via `UpsertPageForRole{PageType:"tool", AdoptUnshippedRows:true, Refresh:[]}` — a live same-role page is attached to AS IT STANDS (nothing about the served page changes as a side effect; the lane retires the ported slot and re-renders as its own step). `Refused()` (live page of another role) ⇒ component cleaned up, honest failure. `page_adopted` in the result map. **Cleanup asymmetry:** a failed link deletes the page only if this call created it (mutation-proven). Flag absent ⇒ byte-identical old path (pinned by test).

Consumers: exactly one live step names the action (`tool-generator.save_tool`) — measured, so RFC_022 says not architecture-scope.

## How to close (fixed AND live)

1. Roll lands: `kubectl -n ai-persona-system exec <chassis-pod> -- grep -aq "<commit sha>" /proc/1/exe` with a control.
2. Apply the HOLD seed (`sql_for_agents/435_tool_generator_adopt_existing_page_HOLD.sql` → rename/apply) — sets `adopt_existing_page: true` on `tool-generator.save_tool` ONLY.
3. Refile the aspect-ratio pilot (RUNBOOK INSERT); expect `create_result.page_adopted = true`, new `page_components` row on the EXISTING page id at position 2, NO new `pages` row.
4. Retire the ported slot → rerender → grade at the artefact (RUNBOOK). Then this file → `bugs_closed/`.

## Related finding recorded here (not a separate bug yet — needs its own 090 before it is one)

The ab-test fork (`cd60486c`) is a **hollow shell**: its LLM-generated template externalises 31 UI-copy fields (`section_heading`, `calculate_button_label`, …) that nothing fills. Served verbatim by the deployer ⇒ 47 raw `{{.` tags on the live page (measured 2026-08-16 09:55Z); after the 08-15 `section_edit` re-rendered it against `content_data = {}` ⇒ 13,284 chars of markup with **zero visible text**, which the slot floors (bugs 178/253) did NOT catch — they measure class-attribute retention, and every class survived; only the text nodes emptied. The lane REVERTED the page to the working ported slot 2026-08-16 (`page_rerender:owner-gate:tool-ab-test-calculator:revert-to-ported`); ab-test is rebuild candidate #2 via this route. **Open question for a floor owner: a text-content floor** (visible chars before/after) beside the class-attr floor.

## Relations
`bugs_open/080` (identity rule reused) · `RFC_010` (`AdoptUnshippedRows`) · `bugs_open/281`/`285` (the ported-instance arc this route replaces) · `bugs_open/204` (does NOT bite: single named slot) · lane: `docs024_key_docs_latest/webdesign_tool_rebuilds/`.
