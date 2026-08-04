# NOTES — bugfix 107 homepage skeleton (append-only, newest at the bottom)

## 2026-08-04 — lane opened

- Picked 107 after a fleet-wide ownership sweep: who-owns + live-transcript
  greps + `site_work_items` queue. The only recent touch was a triage `head -45`
  read at 09:17 today by an idle session.
- On the way in, found `bugs_open/121` (house voice) was fixed AND live since
  2026-07-27 — the file simply never moved. Re-verified all four layers today
  (canonical row, template placeholder, pod-grep v1.0.1251, llm_call_log
  artefact check: 8/8 today's page-content-writer prompts carry the resolved
  voice block) and closed it (`4c449273e`).
- Re-validated 107 against the live fleet (RUNBOOK §1). Strongest single fact:
  **lendzy.co.uk, built 2026-08-02 — six days after filing — has the exact
  skeleton the bug describes** (`hero > brief-explanation > info-card-grid >
  mechanism-flow > call-to-action`). The bug's original table remains accurate
  for the older sites.
- Sites that differ (vonc, gamesdesign, dartsonline) are hand-directed;
  ported sites (loancalculator family) bypass the planner entirely
  (`ported-prose`/`ported-page`). Neither refutes the claim — both are paths
  around the planner, not the planner varying.
- [INFERRED] the planner default is the cause — that is the bug file's reading
  and matches the composition census, but I have not yet read
  `plan_sections_action.go` end to end. Two read-only research agents
  dispatched (code map + docs prior art). 090 to be filed once the mechanism
  is grounded in symbols, per the 2026-07-31 owner ruling.

## 2026-08-04 — the vocabulary already exists at BOTH ends, measured

The bug said "the estate already has the vocabulary"; it is stronger than the
file knew. Measured live:

- **Per-site kind exists:** `site_specs.aspect='classification'` (17 current
  rows) carries `category` — e.g. lendzy.co.uk = **"hub"**, fundamentallyai =
  "brochure", idea.uk/gamesdesign/mortgagecalculator = "interactive",
  dartsonline = "ecommerce". A separate `site_archetype` aspect (6 rows, the
  adopted/ported sites) carries a rich label + design block.
  `resolved_composition` also exists as an aspect (13 rows) — read it before
  inventing anything.
- **Per-component suitability exists:** `content_components.suitable_site_types`
  — **86 of 178 active components** carry values like `["brochure","saas",
  "landing-page"]`, `["interactive-platform"]`, `["affiliate","comparison"]`.
  `experience_patterns` has the same column.
- **And lendzy proves the two ends do not meet:** classified "hub", built
  `hero > brief-explanation > info-card-grid > mechanism-flow > call-to-action`
  — while `features`/`testimonials`/`faq` are marked brochure-ish and nothing
  marked for hubs was preferred. The wiring between classification and section
  planning is the missing piece, exactly as filed.
- NOTE the two vocabularies are NOT the same enumeration ("hub"/"interactive"
  vs "interactive-platform"/"saas"/...). Any fix must reconcile or map them,
  not assume they join.
- `sites` has NO classification column (the bug file's "`classification`
  carries `site_type`" is loosely worded — it lives in site_specs, keyed by
  aspect). Schema-first paid off here.

## 2026-08-04 — code map back (read-only agent over the repo; citations verified spot-wise)

The composition is born in **`build-site-planner`'s `plan_site` LLM step**,
prompt in `agent_definitions.default_config` (in-repo mirror
`docs/agent_docs/sql_for_agents/053_build_site_planner.sql:1978-2228`), NOT in
`plan_sections_action.go` (which only triages readiness of a list it is
handed). Twelve load-bearing facts recorded; the ones that reshape the fix:

1. **The menu is unfiltered**: `load_components` = `SELECT … FROM
   content_components WHERE component_level IN ('section','element') AND
   is_active` — no `suitable_site_types` predicate. Every site is offered the
   whole library; the prompt then says "use ONLY these".
2. **The prompt's one-shot example IS the brochure skeleton**
   (`"sections": ["hero","features","testimonials","call-to-action"]`) — the
   shape we keep getting is literally the example the model is shown.
3. **Rule 11 is the only site-data-conditional section rule** (news_feed →
   add latest-news). Nothing forbids a section, nothing constrains order,
   nothing branches on site_type.
4. **The LLM already emits `site_type` in its plan JSON and
   `WriteSitePlanAction` discards it** — `site_plans` has no column for it. A
   plan cannot be audited against the archetype it was built for.
5. **`component_selector.go:176` already scores `suitable_site_types @>
   site_type` (+0.35) but NO live step config passes `site_type`** — the
   selector context gets "" and every candidate takes the ELSE 0.05 branch.
   Wired, starved. (065/246/309 step configs all pass only
   site_id/sections/page_name.)
6. **Two fallbacks re-impose the brochure after planning**:
   `defaultSectionsForPage` (`apply_gap_plan_action.go:953-978`, hardcoded
   `hero > … > call-to-action` defaults, site_type not a parameter) and the
   modal same-role sibling synthesis
   (`load_page_sections_from_spec_action.go:281-370`) which persists the
   borrowed skeleton into `pages.sections`, indistinguishable from a plan.
7. **`validate_components: true` silently drops unresolvable section names**
   (`v3_site_actions.go:3055-3062`) — contradicting the roadmap block's
   promise that unknown section_types pass through to the selector. A
   brief-specified novel section dies here.
8. **Three incompatible site_type vocabularies** in flight (classifier 7+1,
   strategist 7, old site-classifier 4) — and `content_components.
   suitable_site_types` values ("saas", "affiliate", "comparison",
   "interactive-platform") match none of them exactly.
9. Built pages' compositions are force-preserved on re-plan
   (`reconcilePlanWithRealised`); the release valve is `recompose_pages`.
10. `site_plan_directives` already has `category='structural'` at section
    scope with lock transfer — modelled, unused for composition.
11. `site_plan_sections.ordering` is first-class + unique — an ordering rule
    has somewhere to live.
12. Correction to the bug file's cited fallback: `plan_sections_action.go:
    1210-1218` is the 024 template-truncation guard, not a composition
    fallback.

## 2026-08-04 — MISSTEP, and the lane converts: the bug is owned, and was parked

The docs-prior-art agent returned two facts my ownership sweep missed:
**the owner parked 107 on 2026-07-27** (oufe handoff :78-82), and the
owner-approved **vigilant_designer_offer_analysis** programme (active TODAY)
carries candidate 1 verbatim as Phase 4.1 and candidate 3 as Phase 4.2.

- My sweep (who-owns + transcript greps + queue) was keyed on `107_HANDOFF` —
  a spelling neither the park nor the programme ever uses. The check that
  would have caught it: read the FILING lane's latest handoff, and grep other
  lanes' PLANs for the bug's mechanism words, not its number. Logged in
  WRONG_CALLS.md this date.
- Actions taken instead of a fix: mechanism map appended to the bug file
  (dated section), `CONTRIB_2026-08-04_phase4_mechanism_map_from_107_sweep.md`
  filed into vigilant_designer_offer_analysis/, PLAN here corrected in place.
- The 090 diagnosis was NOT filed — deliberately. The root-cause assertion now
  lives in the bug file as a *contribution to an owned lane*, marked as a
  snapshot to re-verify at implementation time; firing a diagnosis run at a
  parked, owned bug would spend credits answering a question whose owner has
  not asked it yet. If the vigilant lane wants the loop's independent check
  before Phase 4, the symptom text is ready to lift from the bug file's
  mechanism section.
- This lane is now CLOSED as a fix lane. Nothing further owed here.
