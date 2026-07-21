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
