# HANDOFF — vigilant designer + offer analyser (2026-08-14b, evening)

**COLD-START = this file + `PLAN_2026-08-02` (programme + the full decision log, read the last
four entries) + `features_open/030`. NOTES tail (08-14 evening) has the predictions, the results
and both honest limits. This supersedes `HANDOFF_2026-08-14_continue_here.md`.**
Re-run every liveness claim before acting on it — this tree moves, and the file this supersedes
had two claims corrected within a day.

## The one-line state

**B4 IS BUILT, LIVE AND PROVEN END TO END — the programme's headline deliverable exists.** Two
live runs, both COMPLETED, 5 findings → 5 work items each, degraded arm exercised. **It is NOT
enrolled in the automatic sweep: that migration is written and deliberately HELD for the owner's
call.** The lane's next track is the claims-audit extension he approved the same evening.

## What the owner decided (2026-08-14 evening, in a fresh chat)

1. **B4 v1 = ONE analysis, TWO outputs** — the ranked ordering artefact AND the findings, from a
   single LLM pass. Chosen over auditor-only (PLAN §B4 as written) and over ordering-first.
2. **Extend the claims audit to cover `site_specs` prose** — option (b) of the three the previous
   session left open on leopardess's `recurring_value`. Chosen over leaving the field omitted and
   over merging it knowingly. **`recurring_value` stays absent meanwhile**, and B4's degraded arm
   now announces that in every ordering artefact for that site.
3. **Sequencing (B4 first, claims audit second) was MY call, not his** — flagged to him as such,
   because the option text he chose said plainly that the claims work delays B4, and "B4 first"
   had been his standing instruction two days running. **He may reverse it; ask before assuming.**

## What exists now (all verified live 2026-08-14 21:30–21:40 UTC)

- **`offer-analyser`** — `agent_definitions`, config-only, 8 steps, `status='experimental'`,
  `category='analyst'`. Migration `408_offer_analyser_agent.sql`, applied **and** ledger-recorded.
  No image roll was needed: all six actions were already registered.
  `ensure_site_record → load_premise → load_offer_surface → run_offer_analysis → set_audit_source
   → write_offer_ordering → write_offer_findings → complete`
- **`site_specs` aspect `offer_ordering`** — NEW, the estate's first ranked reader-priority
  artefact. **2 of 22 sites** carry one (gaswholesalers.com `degraded=false`,
  leopardessconsulting.co.uk `degraded=true`). Keys: `reader_goal`, `lead_with[]` (≤6, each with
  `rank`/`point`/`why`/`from_field`/`differentiated`), `avoid_leading_with[]`, `inputs_missing[]`,
  `degraded`, `primary_model`, `spec_version`.
- **5 work items** under `audit_source='offer-analysis'` on gaswholesalers.com, all `detected`,
  all at live handlers: `content_rewrite` ×2, `needs_content_page`, `nav_restructure`,
  `cta_improvement`. Plus 5 more on leopardess.
- **Register BIZ-032** (+ its `000_concept_index.md` row) with four landmines. **LANDMINES.md**
  entry on the `write_audit_findings` trap, synced to `doc_notes`.
- **CONTRIB to `copy_quality_two_stage`** telling them the artefact they asked for exists, with the
  read query, the LMC envelope (theirs to fire) and all four limits.
- **`409_improvement_loop_calls_the_offer_analyser_HOLD.sql`** — written, trial-run clean
  (snapshot + guards + 1 row updated + six verify checks, then rolled back), **NOT APPLIED.**

## What the next session should do

1. **PUT THE ENROLMENT DECISION TO THE OWNER — it is the one thing owed.** `PLAN §B5` says
   "enrolment order = owner calls at the time", and migration `409` IS the enrolment: it puts B4
   into every improvement sweep on all 22 sites at once. Two prices he should set rather than
   discover, both in `409`'s header:
   - **one more LLM call per swept site** (30–47KB prompt, ~3–4k output). The fleet **hit its
     Anthropic spend cap on 08-14**: 28 runs failed `usage limits` between 15:36 and 16:42, five of
     them `site-review-agent`. It IS bounded by the audit-due gate (fingerprint changed OR 14-day
     cooldown), so an unchanged site in cooldown costs nothing.
   - **~5 non-parkable dispatchable items per site.** `triage_detect_items_action.go:161-173`
     promotes every `detected` row with no type filter. The item types all have live drains
     (`content_rewrite` 122 rows in 30d, `needs_content_page` 46, `cta_improvement` 36,
     `nav_restructure` 21), so the novelty is volume, not mechanism.
   **To apply once he says go:** rename off the `_HOLD` suffix (it matches `SIDECAR_RE` in
   `run-migrations.sh`, verified — so a scoped `--apply` cannot pick it up while held), apply,
   record, then **run `./scripts/audit-single-owner-actions.sh` — it must stay clean.**
2. **The claims-audit extension to `site_specs` prose — the owner-approved next track.** No
   existing bug covers it; file one (or a `features_open/`) first. The gap, read in the code:
   `check_unverified_claims` scans deployed `page_components`/`site_components` HTML and stored
   `content_data` (`check_unverified_claims.go:1-36`) and **never `site_specs`**; and it never
   repairs, by design (*"Truth decisions are human — auditors raise work items, they never rewrite
   content"*, `:39-41` and `:140`). **The motivating case is leopardess twice over** — "eight
   departments" was audited out of that site's pages, and the same site then produced a fresh
   fabrication one layer back, in the surface the audit was never extended to cover.
   ⚠ **13 refreshed premise records have never been claim-checked** (08-12 refresh). A 3-site
   eyeball found nothing of that class, which is a sample, not a check. That is the population the
   new detector should be pointed at first, and it is a real disconfirmable test.
3. **B4 v2, one change, already specified:** `load_offer_surface` passes page METADATA only — name,
   type, nav membership, title, meta description — and **not one word of what any page says.** So
   B4's page-level findings can be hypotheses (2 of the first 5 were; the model said so itself in
   the finding text). Add a **bounded head-of-hero excerpt per page** — bounded because the surface
   is already 14,887 chars at webdesign.co.uk's 101 reachable pages. Also one prompt line requiring
   attribution in the ordering's `why` clauses (two of them inherited the premise's behavioural
   language unattributed — *"captures return-visit intent"*).
4. **Watch the 5 gaswholesalers items travel** — `detected → triaged → claimed → complete`. That is
   IMP-016's "one clean cycle", and it is the last thing unwitnessed for B4. Two of the five are the
   inferential ones, so what their handlers do with them is the interesting part.
   `SELECT item_type, status, claimed_by FROM site_work_items WHERE spec->>'audit_source'='offer-analysis';`
5. **webdesign.co.uk is still owed the PLAN §B5 proof run.** It was skipped tonight on purpose: it
   had a `content-gap-planner` **executing** and 23 unresolved `needs_page` rows, so another session
   was mid-work there. It is also the worst-case surface (101 pages, 14.9KB), which makes it the
   truncation test — `__truncated` was absent on both runs so far, but neither was the big one.

## Watch-outs

- **⚠ NEVER point a `findings_field` at the object above the array.** `write_audit_findings` has
  no `case map[string]interface{}` — findings vanish, the step SUCCEEDS, `items_created` is 0 and
  the run's `error` is NULL (`bugs_open/272`). **`items_created = 0` proves nothing on its own** —
  it is also what a clean site looks like. Check the PAIR: LLM findings count vs `items_created`.
  B4 points at `offer_analysis.result.findings` and a migration guard asserts that exact string.
  Full entry now in `LANDMINES.md`.
- **⚠ Routing is by `category`, not by `work_item_type`** — that field is read by nothing. B4's
  seven allowed categories all have live routes; an off-vocabulary one mints
  `audit_finding_<x>` at content-gap-planner (6 such rows fleet-wide, 5 still `detected`).
  And `page` must be the EXACT `pages.name` or the finding becomes *create a new page*.
- **⚠ `write_site_spec` DEEP-MERGES: an omitted key survives while looking current.** Maps recurse,
  arrays replace (`site_spec_actions.go:513`). B4's prompt demands every `ordering` key every run.
  Never put a timestamp in the payload — the ROW carries freshness, and an LLM asked for the time
  invents one.
- **⚠ The reachability predicate in `load_offer_surface` is an INLINED COPY of Go**
  (`PageMayBeLinkedPredicateFor` — deliberately not `PageHasShippedPredicateFor`, which would drop
  11 pages that serve HTTP 200). A config-only agent cannot call the helper, so if that floor
  changes in Go, this string does not follow it.
- **⚠ B4's inputs are unverified prose and it grades against them fleet-wide.** That is the whole
  argument for item 2 above, and it is `bugs_open/161`'s shape one layer up.
- **⚠ A banned-term screen is NOT a claims check** (carried forward, and it is item 2's core
  design constraint): the regex from leopardess's 2026-07-16 ruling passed prose containing a
  brand-new fabrication with no banned term and no numerals at all. **Invented specificity is the
  signal** — a frequency, a count, a named topic list, in a field nobody supplied source material
  for. Read it, then check the most checkable sentence against the database.
- **⚠ Do NOT re-run a donor run hoping for cleaner prose.** Same generator, same failure rate;
  cherry-picking until one clears the gate launders a fabrication into the record.
- **⚠ TWO schedules drive our checks** (carried forward). `site_discovery_rotation` covers only the
  rotation driver (7-day period, `LIMIT 1`, fires every 3h). The improvement loop runs the same
  checks as a CHILD orchestration, is hand-fired by other sessions, does **not** stamp that table,
  and **triages and dispatches** what our checks file.
- **⚠ `status='complete'` cannot tell a RETRACTION from a repair** — read `result->>'resolved_by'`.
- **⚠ Three `capability_gap:revenue_shape` rows stay OPEN by design** (`handler_missing`,
  `deferred`, empty handler) on dartsonline.com, loancalculator.co.uk,
  loanandmortgagecalculator.co.uk. Affiliate is DEFERRED (owner, 08-14). They are the standing
  record that three live sites are outside the offer checker's reach, and they retract themselves
  if the capability is built. **Leave them.**
- **Remediation vehicle, proven ~20 times now:** oneshot rows in `scheduled_tasks`
  (`target_topic='system.agent.scheduled.requests'`, `input_data={domain,site_id}`,
  `fire_message=true`, no pre_query), **disabled the moment `last_triggered_at` is stamped** (~20s).
  **Build `input_data` from a subquery against `sites`, never from a UUID you typed.**
  **Never `run_improvement_sweep_once.sh` for a read** — its triage promotes on every path.
- **Check the pod age before dispatching:** no orchestration dispatch within ~300s of a chassis
  pod (re)start — the spawn is silently dropped. The fleet rolled at 20:36 on 08-14.
- **Two claims from this lane were corrected within a day of being written** (carried forward,
  because it keeps being the right warning): a predicted outcome recorded in an evidence column,
  and a real measurement generalised past its axis. Both passed every marker discipline.
  **Before writing "so X is safe/proven/repeatable", name the property the number is about and
  check it is the same property as the claim.**

## Who owns what nearby

portfolio_positioning owns premise→writer wiring; brochure_component_library owns 016's
first-user relationship; bugfix_149 owns checker-layer plumbing; bugfix_230 owns SCH-025.
**`bugs_open/272` is the bug_backlog_clearing lane's** (filed 08-14) — B4 routes around it rather
than fixing it, so **do not start a competing fix**; contribute into their file if you learn more.
**`copy_quality_two_stage` + the loanandmortgagecalculator lane are actively working LMC** —
nothing was fired at LMC tonight and nothing should be while their controlled round-3/round-4 copy
pair is in flight; the CONTRIB hands them the envelope to fire it themselves.
**webdesign.co.uk had another session mid-work at 21:30** — check before firing there.
This lane owns: the drain, the critic, the recompose handler, anti-brochure compose-time work,
**the offer analyser and `offer_ordering` (BIZ-032)**, and WII-014.

**Also carried:** `bugs_open/198` (css-patch-agent) — both fix candidates live and pod-verified,
open only for a witnessed end-to-end run. And the fleet-wide round-trip-writer inventory at
`bugfix_198_roundtrip_writers/HANDOFF_2026-08-10_continue_here.md`.
