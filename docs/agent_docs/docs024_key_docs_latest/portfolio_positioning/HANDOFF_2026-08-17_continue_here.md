# HANDOFF — Phase B CLOSED; Phase C pilot BUILT and all three proof points PASSED; deploy path is the open problem — 2026-08-17 evening, continue here

Supersedes `HANDOFF_2026-08-16_continue_here.md`. Owner rulings unchanged: P9 six decisions,
pilot = remortgagecalculator.uk (M4), build order M→B→I, B8/B9/I10 HOLD, bug 270 hands-off,
copy-voice work lives in session "copy quality two stage".

## 1. THE ONE THING TO DO FIRST

> ## ✅ THE PILOT BUILT, AND ALL THREE PROOF POINTS PASSED — 2026-08-17 evening
>
> The earlier blocker (a fleet-wide head-of-line stall) **cleared on its own at 12:12Z**; the
> pilot then ran the whole pipeline unattended, 12:20 → 13:13, and produced pages.
> **The 090 I filed (`5fbb7f4c…`) was overtaken by events** — read it for the structural half
> (the selector has no skip/backoff), but do NOT read its verdict as a live outage.
>
> | proof | result |
> |---|---|
> | **432** — the flag | `content_features.mortgage_lender_directory = {recommended:true, kind:"mortgage-lender", separate_page:true}` ✅ exactly as the pre-flight predicted |
> | **433/441** — the plan | page `mortgage-lenders`, page_type `mortgage-lenders`, sections `[hero, mortgage-lender-directory-listing, call-to-action]`, `in_header=true`; homepage carries `mortgage-lender-directory` ✅ |
> | **drop check** | zero `PLAN_SECTION_NAME_DROPPED` for this site ✅ — so that composition is what the planner EMITTED, not what survived a cull |
>
> **Cost baseline (Phase C deliverable): 43 LLM calls · 389,406 in · 120,822 out · 11 assets**,
> joined to the pilot's own 83 orchestrations — NOT a time window, which would have swept in
> other lanes' council/diagnosis spend. **A floor, not a total.** page-content-writer is 29 of
> the 43 calls and ~71% of input tokens.
>
> ### ⚠ THE ROLL SHIPPED NO NEW CODE — CHECKED, NOT ASSUMED
> Pods restarted 14:42Z but the tag is **unchanged at `v1.0.1305`**, and the binary carries
> **none** of this session's commits (POS control present, NEG absent). **`bugs_open/292`'s fix
> is STILL NOT LIVE.** Another lane measured the same thing today: *203 commits unshipped while
> pods looked new.* **A restart on the same tag is not a release — `IMAGE_TAG` must be bumped.**
> (Do not reuse my `grep -aq "bestDomainKey"` probe arm: that is a LOCAL VARIABLE, which Go
> strips, so it reads ABSENT even on a binary that has the fix. The commit-stamp probe is the
> one that carries the result.)
>
> ### What actually failed — the pilot's real yield, and none of it is this lane's work
> 1. **11 × `failed to get latest commit/base tree`** (10 `needs_imagery`, 1 `needs_page`
>    deploy), logged under the misleading code `LLM_API_ERROR`. **`sites.github_repo` is EMPTY
>    and that is NORMAL** — 6 of the 8 most recent sites are empty and serve by the B2 route,
>    so "the repo was never created" is the WRONG reading. The question is why the deployer
>    took a git path on a B2 site. **This is the top item: one cause, eleven failures.**
> 2. **20 × `unrendered_template` `{{end}}` blockers** on 2 pages. **Checked specifically:
>    NOT the seeded `banned_claims`** — the guard did not over-fire and blocked nothing. A
>    component is leaking raw Go template syntax into rendered output.
> 3. `component_validation_rejected` ×6 (`mortgages-repayment`), `needs_new_component` failed
>    ×2 at `store_generated_component`, `needs_rerender` failed after 3 retries.
> 4. **10 × `unresolved_cta` + `needs_section_data` + 2 × `needs_page` at
>    `needs_human_review`** — the HITL queue, expected on a new site. Work it, don't debug it.
>
> **"The machinery works" and "the site is finished" are different statements — only the first
> is true.** Nothing is deployed; the site is built but not live.

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
- **`bugs_open/292` FILED + FIXED + COUNCIL-APPROVED (commit `e0d662243`, corr
  `d9ca49ae-…`, unanimous round 1). INERT UNTIL THE NEXT ROLL.**
  `matchVerticalDirectory` collected its domain signal by ranging a map (randomised) and
  appending EVERY match into a first-match-wins dispatch, while the map deliberately mixes
  recommending and not-recommending entries. `mortgage-refinance.co.uk` — M4, the pilot's own
  family, `"refinance"` contains `"finance"` — flipped per run. Reproduced on iteration 1;
  fix + regression test pass 600×3. **The pilot domain is unaffected** (one keyword only,
  deterministic on old and new code). Council corr `d9ca49ae-1c5d-476c-9059-361ed95531bb` —
  **APPROVED, nothing outstanding but the fleet roll.**
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
2. ~~Read the 292 verdict~~ **DONE — APPROVED unanimously, round 1** (2026-08-17 11:15Z,
   *"all reviewers approve"*, 6 abstained). One LOW advisory (editquality): the test assumes
   `matchVerticalDirectory`'s signature, and a wrong guess would fail to compile — already
   answered, the test compiles and passes 600×3. Nothing outstanding on 292 but the roll.
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
