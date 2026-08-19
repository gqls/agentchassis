# NOTES — bugfix 311 (append-only, newest at the bottom)

## 2026-08-19 — session start, research pass

- Bug still VALID [MEASURED 2026-08-19]: class counts re-run — 26 active base section
  rows with NULL section_type (unchanged), 89 with section_type=function, 26 differing
  (was 25), 80 tool-level base (was 79). Live failures through 2026-08-18 23:26 UTC
  (`needs_new_component` failed items for loans-settlement/overpayment/credit-health/
  standard-calc/compare-loans/interest-rate-stress-test/car-finance).
- No fix in flight elsewhere: `who-owns` → portfolio_positioning lane documented it and
  explicitly parked the fix pending "a council round and the owner's say-so"
  (HANDOFF_2026-08-18c §4); no uncommitted platform-code changes touch the three files;
  no open work item routes at the fix.
- **Mechanism refinement (the load-bearing find):** Path 1 of plan_sections already
  matches by function (`loadSectionComponents` pass 2, no level filter). The incumbents
  are dropped there by `sectionTemplateValid` (no `</section>`; they are tool-shaped,
  `created_from='manual'`, ending `</script>` — measured on all three). The drop, not
  the NULL section_type, routes the flow to the selector. Independent artefact:
  work item `3d775f99` (2026-08-15) defers with "stored component 824e3309 … failed the
  template guard" — the same rows, named by the guard.
- Corollary: bug file fix candidate 2 (backfill section_type) is REFUTED for the
  guard-dropped subclass (creates a `selector_error` → "ready as-is" degrade) and a
  NO-OP for guard-passing rows (Path 1 already resolves them by function). Details in
  PLAN §"The mechanism".
- 090 filed on the refinement (per the 2026-07-31 owner ruling — durable structural
  claim): intake `1306e72c-c725-4c3b-b0c3-8a63137f35fb`, run corr
  `f1433782-6ba7-4304-a7f9-8bd830dfb7c9`. Verdict pending at time of writing.
- Adjacent machinery read before designing: CLC-004 (pre-generation field-name
  preservation — dormant for NULL-section_type rows, which is why every generation drew
  fresh field names and the guard kept firing), CLC-006 (regen-vs-create keyed on
  LLM-chosen function — this bug is its cross-site worst case), RFC_034 (DECIDED
  2026-08-17; convert-by-id programme; adjacent, not conflicting),
  `check_unresolved_sections` (the retry engine that kept re-arming the failing pages;
  converges once a valid component exists).
- `input_data` for component-creator runs carries `site_id` AND `domain` (verified on
  run for item `7a2219bc`) — the diversion needs no new plumbing.
- Schema facts that shaped the design: `content_components.name` UNIQUE (global);
  `function` unique only for tool-level base active rows; no site ownership column.
- Live chassis: v1.0.1314, pods restarted 2026-08-19 ~08:00 UTC — last night's plan
  logs gone; provenance line scrolled (expected; not evidence of anything).

### Missteps / corrections this session

- (none yet beyond inherited ones; the candidate-2 refutation is a correction to the
  BUG FILE's account, recorded there and in WRONG_CALLS.md once the 090 verdict lands.)

## 2026-08-19 — build, tests, submission (same session, later)

- 090 refinement run FAILED on infrastructure at the `verdict` step: Anthropic API
  status 400 "You have reached your specified API usage limits" — the neighbouring
  diagnosis run (`6f900e18`, another lane) failed identically at 10:25 UTC, so it is the
  fleet cap, not this run. Intake `c1d726ad` reset to `triaged` for automatic re-claim.
  Substitute first-hand verification stated in the bug file per the 2026-07-31 ruling.
- **Cross-lane arrival mid-session:** commit `73c2505e2` added to the 311 bug file —
  RFC_036 is the same wall at TOOL level (`create_tool_component`, unique index) with a
  forked_from remedy. Read RFC_036: OPEN, no code proposed, owner holds a contained
  interim. Decision: this round fixes the SECTION writer only, stated plainly in the bug
  file with the reason the remedies differ (a section-level fork would be invisible to
  every selection path; a tool-level fork escapes the partial unique index and deploy
  links pages itself).
- Built: `component_storage_identity.go` (+ tests), wiring + section_type self-heal in
  `store_generated_component_action.go`. Full actions suite green; mutation proof run
  (diversion deleted → foreign-collision test fails on uncovered incumbent UPDATE →
  restored); `go build ./...` clean (one PRE-EXISTING vet finding in
  `load_component_library_actions.go:207`, untouched file, not ours); archive-HEAD
  overlay build + full suite green (my commit cannot break shared HEAD).
- Council submitted: corr `fc3ac5f4-ee3a-4e27-88ab-a8b2536b2c1d`. The round may fail on
  the same API cap — see RUNBOOK before re-firing anything.
- Register CLC-020 appended; LANDMINES 311 entry corrected in place (the "invisible to
  the selector" mechanism was incomplete — Path 1 + template guard is what hides the
  incumbents); WRONG_CALLS row for the refuted candidate 2.
- **Declared same-file passengers at commit time:** `component-lifecycle.md` carries the
  283 lane's uncommitted CLC-021/CLC-022 entries + a CLC-017 correction; `LANDMINES.md`
  carries their new instance-scope-bindings entry. Both complete and coherent; declared
  in the commit message rather than waiting (their entries say "committed this session",
  i.e. they are mid-cycle; a pathspec commit takes the whole file either way).
