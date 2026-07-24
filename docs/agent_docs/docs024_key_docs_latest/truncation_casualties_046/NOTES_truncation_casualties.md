# NOTES — bugs_open/046 truncation casualties (append-only, newest at bottom)

## 2026-07-21 — session start, re-grounding

- Re-ran the census against live DB: **still exactly 9** (unchanged from the
  07-20 filing). Good — the bug's evidence held.
- The 5-pair imbalance predicate (`<script/<style/<section/<div/<fieldset`) catches
  **exactly the same 9** as the script-only census — measured fleet-wide,
  `n_script_only == n_any_pair == 9`, zero over-fire from div/section/etc. So the
  general predicate is safe to use and matches the census precisely.
- `ends-mid-token` (the other half of `toolTemplateValid`) would add **36** false
  positives fleet-wide. So the discovery sweep must use imbalance ONLY.
  `toolTemplateValid` can afford ends-mid-token because it is a load-time schema
  drop against the 19 healthy *tools* (all end cleanly), not a fleet queue item.
  This is the key calibration decision — recorded so nobody "improves" the sweep
  by adding the heuristic and floods the HITL queue with 36 items.

## 2026-07-21 — design constraints found while reading the machinery

- Two regeneration paths exist, NOT one: `needs_component_regeneration` →
  component-creator (sections; 52 complete / 5 failed in prod), and
  `needs_tool_recreation` → tool-recreation-handler (tools). The census is 8 tools
  + 1 section, so the "right" auto-remedy would differ by level.
- **But auto-routing either is unsafe.** `bugs_open/020`: tool recreation invented
  a 2,100-practice directory and shipped it live. So the detection check must
  **detect-and-surface**, not auto-fix — the `dead_controls` precedent exactly.
  Decision: `needs_human_review`, no handler, with `intact_version_available` in
  the spec so triage is immediate.
- **Package cycle:** `actions` imports `discovery_checks` (registry wiring), so a
  discovery check CANNOT import `actions.balancedPairs`. Mirrored the list locally
  with a drift-guard test (`TestTruncationTagPairsMirrorGuard`) that fails if the
  two diverge. Not ideal (one duplication) but guarded.
- **Delivery is owned and buggy.** `bugs_open/024` (travelling-docs, active today)
  — a template fix does not reach the live page; defect 6 open. So restoring a
  template fixes the source, not the page. Did NOT touch the re-render pipeline.

## 2026-07-21 — built + committed the check

- `check_truncated_component.go` + test + verifier, commit `1e5cb6fdc`.
- Misstep 1: imported `github.com/google/uuid` but never used it (WorkItemSpec
  already takes `dctx.SiteID` which is a uuid). Build failed; removed the import.
- Misstep 2: my first `censusCut` test fixture left `<section>`/`<div>` unclosed
  as well as `<script>`, so the predicate (correctly) returned three tags, not
  one. Fixed the fixture to close section/div so it asserts the clean
  "script-only" signature. The test caught my own sloppy fixture — working as
  intended.
- Verifier: the package enforces (via `verifier_coverage_test.go`) that every
  check-produced item_type is verifiable or an acknowledged gap. Registered
  `VerifyTruncatedComponentResolved` (re-checks the current template). NOTE:
  `contact_form_undeliverable` still fails that test — **pre-existing**, from
  another thread's commit `3913a0adf`, reproduces with my files moved aside. Left
  it; not mine, and touching their file risks a same-file passenger.

## 2026-07-21 — grip-force restored (source), live page still broken

- Restored `tool-grip-force-friction-calculator-robot-hands-com` html_template
  from intact v2 (23,526, balanced). **Census 9 → 8.** Backed up the damaged
  24,409-byte version to scratchpad first.
- Checked the delivery surface honestly: the page is `build_status=needs_rebuild`,
  its `rendered_html` is still the 23,874-char damaged v1 render (1 `<script`, 0
  `</script>`), and the **live URL still serves broken JS** (curl: `<script`×3,
  `</script>`×2 — matches the bug's original evidence). So the source is fixed;
  the page is not, and won't be until a re-render runs (024's pipeline). Recorded
  as NOT done, deliberately, rather than left to look finished.
- Why restoring is still safe & useful: while the template was damaged,
  `toolTemplateValid` REJECTED it and the re-render CARRIED the stored damaged
  HTML (no change, no harm). Now that it ACCEPTS the restored template, the next
  re-render renders good bytes. The restore cannot make the already-broken live
  page worse.

## 2026-07-22 — detection LIVE + ENABLED + PROVEN end-to-end

- Chassis **v1.0.1149** on prod carries the check. Pod-verified with the
  discriminating literals I created (`truncated_component query failed`: 1;
  `still truncated: unterminated`: 1; positive controls present; negative control
  0). Pod up 24m — past the 300s post-restart dispatch window.
- Numbering-collision scare, resolved: migration 192's ledger note claimed "186
  already applied by hand". It is NOT — my check was `already_enabled=f`, and 186
  is uniquely my file (only `186_*.sql` in the tree). 192's author saw the file
  exist and assumed it was applied; it was correctly waiting image-first. No real
  collision — applied 186 cleanly.
- Applied seed 186 (psql -f; its own guards + snapshot `b05773e0`), then
  `run-migrations.sh --record-only` with a note. Verified: checks array now has
  `truncated_component`.
- **Induced the check** (verify-the-failing-branch discipline, not a pod-grep):
  triggered `completeness-discovery-agent` on vonc.com (corr `c6721ab9`), built
  the kcat publish by hand to SKIP the trigger script's hardcoded finetuning.uk
  auto-approve tail. Orchestration COMPLETED clean; item `ae5ab628` created for
  `tool-arena-interface-vonc-com` with the exact expected spec
  (`unterminated:["<script"]`, `intact_version_available:false`, priority 35,
  needs_human_review). The unplaced archetype-clash-calculator was correctly NOT
  flagged (0 page_components). Correctness proven, not just deployment.

## 2026-07-24 — grip-force FULLY repaired live; delivery blocker gone

- Re-grounded: census still 8. Both blockers I'd deferred to are now CLOSED by
  other threads: **024 (delivery)** — sanctioned path is section-editor
  `apply_section_edit`/`content_edit` (features_open/009, migration 195 wired
  tool-improver's tail to it); **020 (fabrication)** — prompt mig 183 +
  `check_tool_fabrication` gate, live. Chassis now v1.0.1151.
- **Delivered grip-force LIVE** via the section-editor content_edit path (a pure
  re-render from the current, now-good template — no LLM). Drove it with the 086
  direct-orchestrator wrapper (spawn section-editor → call_agent), input_data
  carrying site_id/page_component_id/page_name/slot_name/edit_type=content_edit/
  field_updates={}. Wrote a reusable drive script (scripts/deliver_via_section_editor.sh).
- Result (corr 06c6c158): section-editor orchestration COMPLETED; rendered_html
  23,874(1/0) → 23,526(1/1); build_status deployed; **live page 3/3 script tags
  (was 3/2).** Full end-to-end repair proven.
- Trap banked: the `content_edit field_updates={}` empty-object DID pass cleanly
  through the wrapper's call_agent input_mapping — apply_section_edit accepted it
  (features_009 said {} satisfies the non-nil requirement; confirmed live).
- Trap banked: curl to a missing scratchpad dir returns "HTTP 200" with 0 bytes
  (the -o write fails, not the fetch). mkdir the dir; the live page was fine (200,
  38,318 bytes) once the file could be written.
- For the 8 remaining: recipe is regenerate (tool-recreation, fabrication-gated) →
  deliver (section-editor). Left as an owner decision — LLM-heavy + changes 8 live
  customer tools + each regenerated tool is a NEW design.

## 2026-07-24 (later) — regeneration proof: arena-interface FULLY repaired (census 8 → 7)

- Owner chose "one full proof first". Investigated the regeneration mechanism:
  - `needs_tool_recreation`/tool-recreation-handler is for ADOPTED sites (rebuilds
    from adoption-crawl interactive_features; its own comments defer generation to
    the tool-suggester path) — wrong shape for our platform-generated tools.
  - tool-generator's `create_tool_component` NO-OPs on an existing active tool
    (`already_exists`) and its birth path always INSERTs a NEW page — a
    regeneration would need deactivate-first + risks a duplicate page.
  - **tool-improver is the right mechanism**: rewrites the existing component's
    html_template in place from the current (truncated) template + an issue
    statement (prompt = "Issue to Fix" + current HTML, sonnet-5 @ 32k); the write
    guard's comparative checks deliberately allow a rewrite onto an already-broken
    row; mig 195's tail auto-emits the section_edit delivery item.
- MISSTEP + trap re-confirmed: first drive (corr 502ba210) published 57s after a
  chassis pod restart (another thread's roll, pod start 12:44:01Z, publish
  12:44:58) → spawn silently dropped, wrapper stuck AWAITING_RESPONSES. The
  CLAUDE.md ~300s no-dispatch window, hit live. Re-fired once the pod was ~1h old
  → ran immediately.
- Proof run (corr 32a77a00, COMPLETED ~2.5min): arena template 23,353 (1/0, cut)
  → **38,342, script 1/1, style 1/1, ends clean `})();</script>`**. No
  fabrication tells (no PRNG/seed arrays, no external fetch — self-contained).
  Census 8 → 7. mig-195 section_edit item bd51ff8a auto-emitted (triaged).
- Delivery: drove section-editor directly (corr 64a6a599, COMPLETED — the queued
  bd51ff8a rides the starved dispatch lane; a direct drive is idempotent with it).
  rendered_html 38,342 (1/1); **live page vonc.com/tools/arena/index.html: HTTP
  200, 55,870 bytes, 3/3 script tags** (bug's original evidence: 39,646 @ 3/2 —
  the swallowed page tail is restored).
- Closed tracked item ae5ab628 (status complete, evidence in result jsonb).
- Recipe script: scripts/regenerate_via_tool_improver.sh (+ the existing
  deliver_via_section_editor.sh). STOPPED here per the owner's instruction —
  6 placed casualties + 1 unplaced-section remain for the scale-out decision.
