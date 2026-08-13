# HANDOFF — loancalculator.co.uk framework rebuild · continue here (2026-08-13)

> Supersedes `HANDOFF_2026-08-10_framework_rebuild_continue_here.md` as the
> continue-point (that file's §0-§5 context still applies; its step table is stale —
> this file is current). The copy/voice thread and `bugs_open/227` remain separate.

```
site      loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
serving   26/26 clean (verified 2026-08-12 evening: false claims 0/26, footer control
          26/26, no page under 2000 B, calculators' inline arithmetic intact)
chassis   v1.0.1295 (2026-08-13 13:53Z), built from 69612d692a4a…, artefact-verified
state     REBUILD WAITING on one code change (planner half) + lock release at fire time
```

## Owner decisions on record (all 2026-08-12, in chat)

1. **Brief approved** — `MISSION_DRAFT_2026-08-11_for_owner_review.md`, WITH the
   keep-pages-and-URLs pin.
2. **Keep the 12 calculator locks** for the first pass; release the other 8 (4 css
   carriers + homepage prose-0 + 3 chrome) at fire time.
3. **Fix the planner** so it sees the calculators and doesn't delete them.
4. Park the maintenance queue — DONE (45 items + the earlier 17; 64 deferred total).
5. **Cut the two false claims now** (owner accepted the wait-for-planner-half
   sequencing) — DONE and verified on all 26 live pages.

## What is DONE and LIVE (nothing here needs re-doing)

| thing | where | proof |
|---|---|---|
| url_shape flag + both canonicalisation surfaces | `site_url_shape.go`; live | v1.0.1288+; literal probes both replicas, controls clean |
| seed `url_shape:'flat'` on this site | `SEED_2026-08-11_url_shape_flat.sql`, applied | current spec row verified (27-entry pages list intact) |
| adoption carry-forward (re-adoption can't drop opt-in flags) | `19acfc895` | LIVE v1.0.1294+, literal + ancestry; council 70256656 APPROVED |
| matchLockedRow identity arm (half 1 of decision 3) | `f4820a877` | LIVE v1.0.1294+, ancestry + stamp; council a625c326 APPROVED (10 advisories, dispositioned in bugs_open/241) |
| false claims cut (footer + guide CTA), rendered_html AND content_data | DB + `chrome/footer.html` (`3d49ae8c2`) | 26/26 live pages, positive controls; purity baseline untouched (digest left NULL; assemble path proven zero-DB-write) |
| backups ×4 (repo tag `loancalc-pre-rebuild-20260811` + tar; bak tables 27/63/3/12; off-cluster dump; snapshot `0d1b55f0` pages=27 chrome=3 locked=17) | see NOTES 08-11 | each verified at creation |
| identity chain for the planner half | measured 08-13 | 12/12 locked tools are MASTERS, functions unique, enrichment lookup has NO level filter |

## THE ONE REMAINING CODE TASK — the planner half (owner decision 3, half 2)

Full design + traps: **`PLAN_2026-08-12_planner_sees_locked_tools.md`** (updated
08-13 — read it first, it is the working doc). Summary:

- Target: `agent_definitions` type=`build-site-planner`, step `load_components`
  (the ONLY definition of six planners that has the step — verified).
- Shape: self-gating SQL — tools appear only if the site's structure spec has
  `plan_includes_tools='true'` AND the tool is already placed on that site's own
  pages. Byte-identical result set for unflagged sites (control-run this claim
  against the live DB before applying).
- **Trap 3 is the blocker**: a `params` path that resolves nil HARD-FAILS the step
  fleet-wide. `$ctx.` has NO site_id (verified in `execution_context_params.go`).
- **The puzzle**: `orchestration_states` has ZERO build-site-planner rows ALL-TIME
  and zero retained rows mentioning load_components — yet `site_plans` keep being
  written (`noted.co.uk` 08-12, created_by=write_site_plan). Retention is suspected
  short; resolution steps are in the PLAN file (read the definition's own
  workflow_plan for the input shape; catch one live planner run and dump its
  collected_data keys). `site_record.site_id` is the LIKELIEST key (write_site_plan's
  own input_mapping example) — but per trap 3, verify on a live run, don't infer.
- Fifth-flag accumulation (RFC_022) must be named in the council submission.
- After applying: council submission, roll, pod-grep a literal + control, then a
  canary replan on ONE flagged site before this rebuild fires.

## THEN the fire sequence (unchanged)

1. Release the 8 non-calculator locks (chrome 3 + css carriers 4 + homepage prose-0)
   — the 12 calculator locks STAY (owner decision 2). Full 20-row pre-release state
   in NOTES 08-11.
2. Extract the MISSION block from the draft to a plain .txt; fire
   `082_submit_domain_unified.sh loancalculator.co.uk --email uk@websy.uk
   --mission-file <txt>`.
3. Monitor via `parent_orchestration_id` (printed id logs nothing); publish→start
   can be ~30 min; no dispatch within ~300s of a chassis restart.
4. Verify per the original handoff step 8: purity query, 26/26 serving, pre/post URL
   diff EMPTY, toolgolden 11/11, calculators IN PLACE (the identity arm's job).

## Standing cautions for this lane

- **Query by correlation id, never by `now() - interval`** after any interruption
  (WRONG_CALLS 08-12 — a successful whole-site redeploy nearly got re-fired).
- The 26 hand-raised page_rerender rows VANISHED from site_work_items unexplained
  (08-12, recorded in NOTES). If the cause surfaces, it belongs in NOTES + LANDMINES.
- The 090 on the matchLockedRow mechanism FAILED on API 529 (not a refutation).
  Re-run when healthy:
  `./…/090_TRIGGER_needs_diagnosis_v1.sh` with the symptom text from NOTES 08-12,
  or accept the stated substitute (first-hand read + mutation test + the
  loanandmortgagecalculator lane's independent test) — owner-ruling 07-31 statement
  is already in NOTES.
- Grep LANDMINES for **your own plan's symbols at submission time** (WRONG_CALLS
  08-11) — the corpus can gain an entry about your seam between coding and
  submitting.
