# HANDOFF — vigilant designer + offer analyser (2026-08-15)

**COLD-START = this file + `PLAN_2026-08-02` (read the last six decision-log entries, 08-11 →
08-15 — two are same-day reversals and both stand as records) + `features_open/030`. NOTES tail
(08-14 evening → 08-15) has the B4 predictions, results, both honest limits, and the review pass.
This supersedes `HANDOFF_2026-08-14b_continue_here.md`.**
Re-run every liveness claim before acting on it. This is not a formality: **two claims in the
handoff this supersedes went stale within twelve hours** (272 closed, 279 fixed) — both were
right when written.

## The one-line state

> **UPDATE 2026-08-15 (afternoon, this same day): decision 1 below is TAKEN — the owner said
> ENROL, and migration 409 is applied, recorded, audit-clean. B4 is in the automatic sweep;
> the first sweep-driven run is not yet witnessed.** The paragraphs below stand as the record
> of the state this handoff was written in; NOTES 08-15 (later) has the application evidence.

**B4 — the offer analyser — is BUILT, LIVE, PROVEN END TO END, and review-passed (migrations 408
+ 421). It is NOT enrolled in the automatic sweep: migration `409_…_HOLD` is written, trial-run
clean, and deliberately held for the owner's call. The lane's next track is the claims-audit
extension over `site_specs` prose, which the owner approved on 08-14.**

## Decisions the owner has OPEN (put these to him early)

1. **ENROLMENT — the one this lane is blocked on. [TAKEN 2026-08-15: ENROL — see update above.]** `409_improvement_loop_calls_the_offer_analyser_HOLD.sql`
   puts B4 into every improvement sweep (audit-due branch, between `call_site_review` and
   `record_audit_pass`). PLAN §B5 makes this his call. The two prices, both in 409's header:
   one more LLM call per swept site (30–47KB prompt; the fleet hit its Anthropic cap on 08-14 —
   28 failed runs 15:36–16:42), and ~5 **non-parkable** dispatchable items per swept site
   (`triage_detect_items_action.go:161-173` promotes every `detected` row, no type filter).
   Bounded by the audit-due gate (fingerprint changed OR 14-day cooldown).
   **To apply on his go:** rename off `_HOLD` (verified: the suffix matches `SIDECAR_RE`, so a
   scoped `--apply` cannot take it while held) → apply → `--record-only` → run
   `./scripts/audit-single-owner-actions.sh` (must stay clean).
2. **Two decisions ANOTHER lane recorded in this lane's README on 08-15** (the `bugs_open/279`
   session; theirs to route, listed here only for completeness): cancel-or-rerun the four dead
   08-13 brief-fidelity findings on mortgagecalculator; and whether `brief-fidelity-auditor`
   becomes a real routed, scheduled check (options in bug 279 candidate 3).

## What exists (every claim re-verified live 2026-08-15 afternoon)

- **`offer-analyser`** — `agent_definitions`, config-only, 8 steps, `experimental`/`analyst`.
  Migration **408** (applied + recorded 08-14) built it; migration **421** (applied + recorded
  08-15) removed 10 doubled apostrophes the dollar-quoted authoring had put in the prompt
  (`''` inside `$prompt$` is two literal characters, not an escape — review-pass find, NOTES
  08-15). Live prompt is now 5,484 chars, 0 doubled pairs, all three load-bearing lines intact.
  `ensure_site_record → load_premise → load_offer_surface → run_offer_analysis →
   set_audit_source → write_offer_ordering → write_offer_findings → complete`
- **`site_specs` aspect `offer_ordering`** — the estate's first ranked reader-priority artefact.
  **2 of 22 sites**: gaswholesalers.com (`degraded=false`) and leopardessconsulting.co.uk
  (`degraded=true`, `inputs_missing=["recurring_value"]` — absent by owner decision). Keys:
  `reader_goal`, `lead_with[]` (≤6 × `rank`/`point`/`why`/`from_field`/`differentiated`),
  `avoid_leading_with[]`, `inputs_missing[]`, `degraded`, `primary_model`, `spec_version`.
- **10 work items** under `audit_source='offer-analysis'` — 5 per site, **all still
  `detected`/unclaimed** (no sweep hit either site since firing). Types: `content_rewrite`,
  `needs_content_page`, `nav_restructure`, `cta_improvement`, `tone_shift` — all at live handlers.
- **Register BIZ-032** (+ index row) with the four landmines AND the 08-15 staleness correction.
  **LANDMINES.md** entry on the `write_audit_findings` trap — now carrying updates from THREE
  sessions (this lane wrote it; the 272 lane fixed half 1; the 279 lane fixed half 2). Synced.
- **CONTRIB to `copy_quality_two_stage`** (08-14): the artefact they asked for exists; read query,
  LMC oneshot envelope (theirs to fire), all four limits.
- **Proof evidence:** run 1 gaswholesalers `afe600a9-…` (58s, 5 findings → 5 items, the
  falsifiable PAIR held); run 2 leopardess (degraded arm fired; the owner-protected `hitl`
  strategy row stayed **byte-identical** to its pinned md5 `cf500fcf…0208`). ⚠
  `orchestration_states` retention is ~24h — these rows will be GONE; the evidence lives in
  NOTES 08-14 evening, do not try to re-query it.

## What changed OVERNIGHT (other lanes; affects what this lane may assume)

- **`bugs_closed/272` — fixed AND live on v1.0.1301.** `write_audit_findings` now parses the
  object shape; `site-review-agent` filed its first-ever items (run `b2c82a25`). **B4's config
  stays on the array path** — belt-and-braces, guarded by 408's verify, and config must never
  assume a binary version. Do not "simplify" it.
- **`bugs_open/279` — fixed in CODE, not yet live.** Unknown categories now file `capability_gap`
  instead of minting `audit_finding_<x>` (commit `d6d56e540`; CI test pins the closed set;
  mig 416 purged the dead `work_item_type` field from the two prompts that asked for it).
  **Until the next chassis roll the old binary still mints** — check the build-provenance stamp
  before believing an unknown category files a gap row. B4's closed category vocabulary stays
  correct practice either way.
- Both fixes grew out of the two halves of this lane's 08-14 LANDMINES entry — it did its job
  twice in under a day.

## What the next session should do

1. **Put the enrolment decision to the owner** (above). It gates everything that scales B4 past
   hand-firing.
2. **The claims-audit extension over `site_specs` prose — the owner-approved next track
   (08-14, option b).** Nothing claim-checks a premise record: `check_unverified_claims` reads
   deployed HTML + stored `content_data` (`check_unverified_claims.go:1-36`), never `site_specs`,
   and it never repairs by design (`:39-41`, `:140`). No bug/feature file exists yet — **file one
   first** (grep both bug dirs again before filing; nothing covered it as of 08-14). Design
   constraints already learned the hard way, all in NOTES 08-13/08-14: a banned-term screen is
   NOT a claims check (the leopardess fabrication carried no banned term and no numerals);
   **invented specificity is the signal** — a frequency, a count, a named topic list in a field
   nobody supplied source material for; check the most checkable sentence against `pages`.
   **First population: the 13 premises refreshed on 08-12, never claim-checked** (a 3-site
   eyeball found nothing — a sample, not a check). That target is genuinely disconfirmable.
   Route findings like everything else this lane ships: with a drain, or not at all.
3. **B4 v2, two small specified changes:** (a) `load_offer_surface` passes page METADATA only —
   add a bounded head-of-hero excerpt per page so page-level findings stop being hypotheses
   (2 of the first 5 were; the model said so itself). Bounded: the surface is already 14,887
   chars at webdesign.co.uk's 101 pages. (b) One prompt line requiring attribution in the
   ordering's `why` clauses (two inherited the premise's behavioural register unattributed —
   "captures return-visit intent"). **Read of two runs, not a measurement** — confirm on more
   runs before treating it as a pattern.
4. **Watch the 10 items travel** (`detected → triaged → claimed → complete`) — IMP-016's "one
   clean cycle" is the last unwitnessed thing. The two inferential findings are the interesting
   ones: what their handlers do with an acceptance test about content the analyser never saw.
   `SELECT item_type, status, claimed_by FROM site_work_items WHERE spec->>'audit_source'='offer-analysis';`
5. **webdesign.co.uk still owes the PLAN §B5 proof run** — skipped 08-14 because another session
   was mid-work there (a `content-gap-planner` executing, 23 unresolved `needs_page`). It is the
   worst-case surface (101 pages), i.e. the truncation test: `__truncated` was absent on both
   runs so far, but neither was the big one. **Check the queue and in-flight orchestrations on
   it first, as 08-14 did.**

## Watch-outs (trimmed to what still bites; history in NOTES)

- **⚠ `findings_field` must address the ARRAY** (`offer_analysis.result.findings`), never the
  object above it — 408's verify guard asserts the exact string. The object path works only from
  v1.0.1301, and **`items_created=0` alone proves nothing**: check the PAIR (LLM findings count
  vs `items_created`). Full entry + three-session update trail in LANDMINES.md.
- **⚠ Routing is by `category`** — B4's seven values (`gap, content, differentiation, structure,
  cta, nav_restructure, tone`) are the closed set; `page` must be an EXACT `pages.name` or the
  finding becomes *create a new page*.
- **⚠ `write_site_spec` DEEP-MERGES** — an omitted `ordering` key silently keeps the previous
  run's value; the prompt demands every key every run. No timestamps in the payload.
- **⚠ The reachability predicate in `load_offer_surface` is an INLINED COPY of
  `PageMayBeLinkedPredicateFor`** — if the Go floor changes, this string does not follow.
- **⚠ B4's inputs are unverified prose** and it grades against them fleet-wide — that is item 2's
  whole argument, and `bugs_open/161`'s shape one layer up.
- **⚠ Do NOT re-run a donor run hoping for cleaner prose** — cherry-picking launders a
  fabrication into the record.
- **⚠ Three `capability_gap:revenue_shape` rows stay OPEN by design** (affiliate deferred,
  owner 08-14) on dartsonline.com, loancalculator.co.uk, loanandmortgagecalculator.co.uk.
  They self-retract if the capability is built. **Leave them.** Related drift to watch:
  dartsonline's premise says `affiliate` while the platform cannot support it.
- **⚠ `status='complete'` cannot tell a RETRACTION from a repair** — read `result->>'resolved_by'`.
- **Oneshot vehicle** (proven ~20×): `scheduled_tasks` row, `target_topic=
  'system.agent.scheduled.requests'`, `input_data={domain,site_id}` **built from a subquery
  against `sites`, never a typed UUID**, `fire_message=true`, no pre_query, **disable the moment
  `last_triggered_at` stamps** (~20s). Never `run_improvement_sweep_once.sh` for a read.
  **Check chassis pod age first** — dispatch within ~300s of a pod (re)start is silently dropped.
- **⚠ TWO schedules drive checks**: the rotation driver stamps `site_discovery_rotation`; the
  improvement loop is hand-fired by other sessions, stamps nothing, and **triages + dispatches**
  whatever it finds. `detected` is not a resting state on a site the loop reaches.
- **Before writing "so X is safe/proven/repeatable": name the property the number is about and
  check it is the property the claim is about.** This lane's WRONG_CALLS entries are all this
  shape, and the 08-15 review added the twin: **a correct claim about a moving estate needs a
  date more than it needs a proof** — two proven-true 08-14 claims were false by morning.

## Who owns what nearby

portfolio_positioning owns premise→writer wiring; brochure_component_library owns 016's first-user
relationship; bugfix_149 owns checker-layer plumbing; bugfix_230 owns SCH-025.
**`bugs_closed/272` is closed; `bugs_open/279` belongs to its fixing session** — contribute into
the bug file, do not compete; its Go half still needs a roll, which THEY are watching.
**`copy_quality_two_stage` + the LMC lane actively work loanandmortgagecalculator.co.uk** — never
fire anything at LMC while their controlled copy pair is in flight; the 08-14 CONTRIB hands them
the oneshot envelope to run B4 themselves when ready.
**webdesign.co.uk had a session mid-work at 08-14 21:30** — re-check before firing there.
This lane owns: the drain, the critic, the recompose handler, anti-brochure compose-time work,
**the offer analyser + `offer_ordering` (BIZ-032)**, WII-014, and now the claims-audit-over-specs
track once filed.

**Also carried:** `bugs_open/198` (css-patch-agent) — both fix candidates live and pod-verified,
open only for a witnessed end-to-end run. Fleet-wide round-trip-writer inventory at
`bugfix_198_roundtrip_writers/HANDOFF_2026-08-10_continue_here.md`.
