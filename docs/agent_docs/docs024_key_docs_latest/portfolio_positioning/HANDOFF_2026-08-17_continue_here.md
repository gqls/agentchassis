# HANDOFF — Phase B CLOSED (all four verdicts approved); Phase C pilot DISPATCHED and running — 2026-08-17, continue here

Supersedes `HANDOFF_2026-08-16_continue_here.md`. Owner rulings unchanged: P9 six decisions,
pilot = remortgagecalculator.uk (M4), build order M→B→I, B8/B9/I10 HOLD, bug 270 hands-off,
copy-voice work lives in session "copy quality two stage".

## 1. THE ONE THING TO DO FIRST

**The pilot is building. Watch it, and check the three proof points in order.** Dispatched
2026-08-17 ~11:4xZ, correlation `fb048d5f-b4b3-49c8-bc02-2810bbe209aa`.
`domain-submitter` COMPLETED; `needs_domain_research` is triaged to
`domain-research-classifier`. Everything below is a query, not an inference.

```sql
-- where the build has got to
SELECT aspect, source_agent, created_at FROM site_specs ss JOIN sites s ON s.id=ss.site_id
WHERE s.domain='remortgagecalculator.uk' AND ss.is_current ORDER BY ss.created_at;

SELECT wi.item_type, wi.status, wi.handler_agent, LEFT(wi.summary,60)
FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
WHERE s.domain='remortgagecalculator.uk' ORDER BY wi.created_at;
```

**Proof point 1 — 432 (the recommender).** Once `classification` exists, it must carry the
directory flag, written by `enrich_directory_features` immediately after
`write_classification_spec`:
```sql
SELECT data->'content_features'->'mortgage_lender_directory' AS flag
FROM site_specs ss JOIN sites s ON s.id=ss.site_id
WHERE s.domain='remortgagecalculator.uk' AND ss.aspect='classification' AND ss.is_current;
```
Expect `{"recommended": true, "kind": "mortgage-lender", "separate_page": true, "reason": …}`.
**Pre-flight says this WILL fire, and says exactly why**: `industry` comes out NULL and
`site_type` `"interactive-platform"` on all four comparable finance sites, neither of which
is in `verticalDirectoryMap` — so the **domain-derived** signal is the only one that can
fire, and `remortgagecalculator.uk` contains `mortgage`. If the flag is ABSENT, that is the
finding, not a nuisance: read `bugs_open/292` first, then check what `industry` actually
came out as.

**Proof point 2 — 433/441 (the planner).** Once `site_plan` exists, the plan must contain a
`mortgage-lenders` page (name AND page_type exactly that), composed
`hero → mortgage-lender-directory-listing → call-to-action`, plus a
`mortgage-lender-directory` section on the homepage:
```sql
SELECT p.name, p.page_type, p.sections, p.in_header, p.in_footer
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE s.domain='remortgagecalculator.uk' ORDER BY p.nav_order;
```

**Proof point 3 — the silent-drop check, and DO run it even if the page looks right.**
`validate_site_plan` drops an unresolvable section name; the drop IS recorded (bugs_open/282's
fix, another lane), so ask rather than infer:
```sql
SELECT occurred_at, agent_type, action, error_message, context
FROM agent_error_log
WHERE error_code = 'PLAN_SECTION_NAME_DROPPED'
  AND site_id = (SELECT id FROM sites WHERE domain='remortgagecalculator.uk')
ORDER BY occurred_at DESC;
```
Table is `agent_error_log` (**not** `agent_errors`); payload column is `context`. Estate-wide
this returns 0 all-history, and that zero is demand-controlled (door live: 12,158 rows/7d
across 63 agent types; recorder present in v1.0.1305, probed with controls) — but it has
**never been positively exercised for `validate_plan`**, so the pilot is its first real test.

**Proof point 4 — B3f (the safety net).** If the planner missed the directory, the checks
should raise it on the first completeness sweep:
`missing_mortgage_lender_directory_page` / `_section`. A clean pilot must be verified at
proof point 2 **directly** — do not infer success from these staying quiet, because they are
the net, not the measurement.

## 2. What is DONE (this session)

- **Phase B's council trail is COMPLETE: 429 · 432 (r2) · 433 (r2) · 434 — all APPROVED.**
  433 round 2 approved 2026-08-16 16:38Z. All advisories dispositioned in NOTES.
- **One real advisory fixed** (debug_historian): the verify blocks did arithmetic on a
  possibly-NULL `p`, so `NULL <> 1` is NULL, no `IF` fires, and **the verify passed having
  inspected nothing**. Fixed in both un-run ROLLBACKs (433, 441); applied forward files left
  as the record of what ran.
- **`bugs_open/292` FILED + FIXED IN TREE (commit `e0d662243`), INERT UNTIL THE NEXT ROLL.**
  `matchVerticalDirectory` collected its domain signal by ranging a map (randomised) and
  appending EVERY match into a first-match-wins dispatch, while the map deliberately mixes
  recommending and not-recommending entries. `mortgage-refinance.co.uk` — M4, the pilot's own
  family, `"refinance"` contains `"finance"` — flipped per run. Reproduced on iteration 1;
  fix + regression test pass 600×3. **The pilot domain is unaffected** (one keyword only,
  deterministic on old and new code). Council corr `d9ca49ae-1c5d-476c-9059-361ed95531bb`,
  **verdict outstanding — read it.**
- **Pilot seeded** (`SEED_2026-08-17_…sql` + `SEED_2026-08-17b_…sql`, both applied): site row
  with email, `evidence_base`, `imagery_style_guide`. Both survived the submitter.
- **Two LANDMINES filed** (both at HEAD, both taken as passengers by other lanes' commits):
  the checks-array silent-enable, and the banned-claims double-escaping below.
- **Pattern filed in 016b §9**: *a guard written for the INNER case leaves the OUTER one, and
  the file's own comment explaining the hazard reads as proof it was handled.*

## 3. The pilot seed's own bug — read this before seeding the NEXT site

**All six `banned_claims` patterns were INERT on first apply, and the seed's verify passed.**
Dollar-quoted SQL passes bytes literally; JSON *then* unescapes — so `\\\\b` stored `\\b`, a
literal backslash. `claims.go` compiles with `regexp.Compile` and falls back to `QuoteMeta`
**only on a compile ERROR**, and a double-escaped pattern is valid regex that matches no
English. It compiled, the fallback never fired, the guard was loaded, listed, counted, dead.

The verify asserted `jsonb_array_length(banned_claims) = 6` and passed — **six inert patterns
count exactly like six working ones.** What caught it was probing: four strings that must
match, one that must not.

Correcting it took three attempts, all recorded in `SEED_2026-08-17b`'s header:
- the `£` check searched for the character it wanted to confirm, because **this authoring
  channel rewrites `\uXXXX` into the character it denotes** — so use `chr(92)`, never a typed
  escape (MEMORY: `escape-sequence-emission-trap`);
- the `LIKE`-based guards were unreadable, because **in `LIKE` the backslash IS the escape
  character** — `'%\b%'` means "the letter b" and would pass on any pattern containing `b`.
  Both now use `position()` + `chr(92)`, which has no escape semantics.
- The guard was then **run against the still-broken rows and required to flag them** (6 of 6)
  before being trusted on the fixed ones.

**Semantics are pinned in Go, not SQL** — `datahelpers/claims_banned_pattern_escaping_test.go`,
compiled exactly as production does. A `psql … ~ pattern` probe is a check in the **wrong
engine**: Postgres spells word boundary `\y` and reads `\b` as backspace.

**The facts roster is deliberately EMPTY.** No rate or threshold could be verified against a
live source, and a plausible figure with an invented source URL is the exact fabrication this
layer exists to stop. Cited lender material arrives through the directory instead. To add a
fact later: source URL + capture date, and **do not** copy loanandmortgagecalculator's SDLT
facts — those are PURCHASE facts and remortgaging is not a purchase.

## 4. Still owed

1. **Watch the pilot** (§1) — then the cost baseline from `llm_call_log` / `assets` (no
   aggregate cost figure exists anywhere in the platform; it must be computed by hand), then
   **owner sign-off before Phase E**.
2. **Read the 292 verdict** (`d9ca49ae-…`). REVISE → fix and resubmit with `RESUBMIT_CORR`.
3. **292's fix is inert until a fleet roll.** Releases are whole-fleet and owner-run; this
   lane does not roll. After a roll, verify at the artefact, never at git.
4. **A Phase-C summary** when the pilot lands — the series' next inflection.
5. Then Phase D decisions / Phase E waves per `PLAN_2026-08-12_fleet_buildout.md`.

## 5. Standing cautions

- **Council latency is HOURS, not the documented ~30 minutes** — measured 17.4h queue-to-start
  on 08-15/16. A `fix_plan` artifact with no `council_report` row means **RUNNING**, not
  dropped. Diagnose by artifact KIND, never by elapsed time. Never retry on that evidence.
- **A council sketch must be CODE**: comment-only sketches are refused client-side, and
  **never elide inside a sketch's string literals** — five rounds in this lane have been lost
  to showing reviewers less than the file contains, twice in a way that looked exactly like a
  bug they should catch.
- `improvement-sweep` is still disabled fleet-wide (since 2026-08-14) — not ours to re-enable.
- Migration numbers moved fast: **435–441 taken during the last two sessions**; next free was
  **442**. RE-CHECK.
- **Rolling back the planner pair has an ORDER: 441 before 433** (441 edits text inside 433's
  inserted block, so 433's inverse refuses until then — by design).
- `git stash` forbidden; pathspec commits; forward-only. **The index lock is contended** —
  wait for it, never remove it. **Same-file passengers are routine**: two of this lane's
  LANDMINES edits were committed by other lanes before this session's own commit ran. Compare
  the commit's printed scope block against what your message claims, before moving on.

## 6. Files of record

This dir: `PLAN_2026-08-12_fleet_buildout.md` · `SUMMARY_2026-08-16_phase_b_complete.md`
(latest milestone) · `NOTES_portfolio_positioning.md` (newest at bottom) ·
`README_where_we_are.md` · `MISSION_2026-08-17_remortgagecalculator_uk.md` ·
`SEED_2026-08-17{,b}_remortgagecalculator_uk_*.sql` · `COUNCIL_SUBMISSION_{292,433,434}_*.json`.
Migrations: `sql_for_agents/{432,433,434,441}_*` (+ROLLBACKs), all applied.
Register: `docs026_concept_register/register/directory-pipeline.md` (DIR-001).
Architecture: `architecture_review/RFC_031_hand_spliced_enrichment_steps_want_an_ordered_list.md`.
Bugs: `bugs_open/292_HANDOFF_2026-08-17_…`.
Commits this session: `07b2ea6d5` (433 approval + rollback hardening), `e0d662243` (292),
`1268ae2ef` (pilot seed + escaping fix + Go test).
