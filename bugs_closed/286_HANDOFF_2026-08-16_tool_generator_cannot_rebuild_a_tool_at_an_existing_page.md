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

## CLOSED 2026-08-19 ~20:50Z — FIXED AND LIVE since 2026-08-16; every "How to close" step done; moved to `bugs_closed/`

**The roll.** The fix `88897190e` first shipped in `v1.0.1304` (pods 10:41Z 2026-08-16; stamp `5de6cddbe`,
ancestry TRUE, junk-hex control absent — the lane's NOTES 2026-08-16 15:10Z). Re-verified today at the
artefact on the current fleet: `v1.0.1316` (pods 17:13Z), binary stamp `07eeba4a1` PRESENT in
`agent-chassis-5ddd9744-86nqf` (`/proc/1/exe` grep; controls in the same breath: junk 40-hex ABSENT,
the fix's own literal `88897190e` ABSENT as expected — a binary carries only its own sha);
`git merge-base --is-ancestor 88897190e 07eeba4a1` → TRUE. Also an ancestor of the `v1.0.1315` stamp
`590ca3a20` probed with both controls by the 306-closing session (`5ac03f247`).

**The config half.** Seed 435 applied 2026-08-16 15:15Z (`ee0228813`, HOLD lifted after the roll); read live
today 16:03Z: `tool-generator.save_tool.config.adopt_existing_page = true`, and the consumer census is still
exactly ONE step naming `create_tool_component` (`agent_definitions`, active, non-snapshot).

**Behaviour, measured with demand, not inferred from the roll.** `orchestration_states`
(`owner_agent_type='tool-generator'`, `created_at > 2026-08-16`): **7 runs with
`collected_data->'create_result'->>'page_adopted' = 'true'`, all COMPLETED** (2026-08-18 → 08-19), 1 greenfield
(`false`), 0 with `already_exists`. The pilot itself (item `99734862`, orch `72f0737e`, 15:48Z 08-16) adopted
the EXISTING page `00979b9e`, minted no `pages` row named `tool-aspect-ratio`, linked the new component at
position 2 — then the lane retired the ported slot, re-rendered and graded at the served page; aspect-ratio is
one of the 12 tools "replaced, live, graded" (lane HANDOFF 2026-08-19). The instrument-alive control:
`agent_error_log` duplicate-key failures on this action all-time are **exactly three rows, one per
constraint** — `pages_site_id_name_key` **once, 2026-08-15 18:29Z (the filing case) and never since**;
`idx_cc_tool_function_unique` once 08-17; `content_components_name_key` once 08-18 (both below).

**Residual — tracked elsewhere by design, and it is NOT small.** This file fixed the PAGE half of "rebuild a
tool at an existing page". The COMPONENT half is three gates on the same INSERT, and the write history above
shows the lane walking into each on consecutive days:
- `idx_cc_tool_function_unique` (fleet-wide library claim) — `RFC_036 §9.3`, built by the `bugs_open/311`
  lane (`e24bc9c0f`, council APPROVED r1 `ceae30f2`) and **LIVE on `v1.0.1316`** (ancestor of `07eeba4a1`,
  checked above — the §11 addendum still says "inert until a roll"; it is not, as of 17:13Z today).
- the action's own per-site `already_exists` probe + `content_components_name_key` (`UNIQUE(name)`, name =
  `<function>-<domainSlug>`): **the generator can build a given tool for a given site exactly once, ever** —
  a re-fix of a native tool needs the old row deactivated AND renamed by hand (lane RUNBOOK / NOTES 08-18
  13:51Z), and the old slot retired by hand before the generator's own rerender claims (RUNBOOK "the retire
  race", margins 2–96 min, lost once). Filed as its own bug from this session:
  `bugs_open/331_HANDOFF_2026-08-19_create_tool_component_cannot_regenerate_its_own_tool.md (numbered 330 in commit 1f70a645b's message and in this file's first write; 330 was taken by another session minutes earlier — renumbered)` (with the
  `090` run and the RFC_036 §12 note). Not this bug's: this bug's page collision is gone.
- The "related finding" above (ab-test hollow shell, text-content floor) stays an open question for a floor
  owner; it was never this bug's mechanism.
