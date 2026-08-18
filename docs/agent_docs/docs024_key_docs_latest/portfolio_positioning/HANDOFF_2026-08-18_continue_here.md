# HANDOFF — Phase C PROVEN AT THE ARTEFACT: the directory page is deployed and rendering cited lenders — 2026-08-18, continue here

Supersedes `HANDOFF_2026-08-17_continue_here.md` (accurate on its own history; this file
carries everything a fresh chat needs). Owner rulings unchanged: P9 six decisions, pilot =
remortgagecalculator.uk (M4), build order M→B→I, B8/B9/I10 HOLD, bug 270 hands-off,
copy-voice work lives in session "copy quality two stage".

## 1. WHERE THIS GOT TO — the whole objective is met

**The chain is closed, verified at the RENDERED HTML rather than at a status.** The
`mortgage-lenders` page is `build_status='deployed'` with `deployed_at` set, and its
`mortgage-lender-directory-listing` component (4,882 chars) renders:
- heading *"UK mortgage lenders, listed"*;
- the owner\'s NON-PRICE ruling surviving into the copy: *"It does not list rates, fees or
  APRs, because those change daily and depend on your circumstances"*;
- real cited entries — **Mansfield Building Society**, **Family Building Society** — each
  claim carrying `lender_type` / `product_types` and a **`source`** link.

researcher → quote-verified claim → register → kind-aware publish → `evaluate_directory_features`
flag → planner rule → page → deploy → **a live page naming regulated firms with citations and
no prices.** That was the Phase B/C question and it is answered.

**All proof points passed** (432 flag ✅ · 433/441 page name+type+composition ✅ · zero
`PLAN_SECTION_NAME_DROPPED` ✅ · directory checks silent for the right reason ✅).
**Cost baseline: 43 LLM calls · 389,406 in · 120,822 out · 11 assets** — a FLOOR (joined to the
pilot\'s own orchestrations, not a time window). page-content-writer ≈71% of input tokens.

## 2. Pilot state — built and PARTLY live

| | |
|---|---|
| deployed | `mortgage-lenders`, `next-steps`, `about` |
| `needs_rebuild` | `index`, `what-your-number-means`, `six-month-checklist` |
| `sites.build_status` | still `pending` |

**Remaining work (none of it directory-related):**
- `needs_new_component` ×3 FAILED at `store_generated_component` (3/3 attempts gone)
- `needs_rerender` ×1 FAILED (timeout, 3/3), `needs_imagery` ×2 FAILED
- 2 pages blocked by **20 × `unrendered_template` `{{end}}`** — a component leaking raw Go
  template syntax. **Checked: NOT the seeded `banned_claims`**; the claims guard blocked nothing.
- `component_validation_rejected` ×6 for `mortgages-repayment`
- **HITL queue: 10 × `unresolved_cta`, 4 × `needs_page`, 1 × `needs_section_data`** — work it,
  do not debug it.

## 3. Two corrections of mine that a fresh reader must not inherit

1. **"Does the `assets` count rise above 11?" was the WRONG decisive measurement** — I wrote it
   into the last handoff. It stayed at 11 and that is CORRECT: the 8 successful retries
   reference 8 asset ids that PREDATE the outage and created zero new rows, because the step
   that failed was the **deployer**, not the generator — a good retry ships an existing asset
   and must not duplicate it. **Before calling a measurement decisive, name the step it
   observes and check that it is the step that failed.**
2. **"A retry failed again → the outage is back" was too crude** and misfired within minutes:
   the failure was an asset-deployer *timeout* while base-tree 404s stayed at 0. **Only a
   renewed `%base tree%` error is evidence about that outage.**

Earlier in the same session I also read a pre-existing `complete:1` as a retry success. A count
already non-zero at t=0 evidences nothing at t+n — take the baseline, or watch a delta.

## 4. Infrastructure facts worth carrying

- **v1.0.1308 is live** (1305 → 1306 → 1307 → 1308). The same-tag reuse that made three days of
  commits inert is fixed; another lane measured it at *"24 code commits across ~10 lanes"*.
- **`bugs_closed/292`** — the random directory recommendation — is **FIXED AND LIVE**.
- **⚠ Never verify a fix by grepping the binary for its commit sha.** The binary carries ONE
  stamp (the commit it was built FROM), so ABSENT is the normal reading on a healthy build; a
  discovery grep for `[0-9a-f]{40}` is worse (20 hits, none a real commit). Use the
  `build provenance` log line, or `git merge-base --is-ancestor <fix> <tag-bump commit>`.
  All three of my failed probes: `WRONG_CALLS.md`, 2026-08-17.
- **The deploy outage is CLOSED** (~832 `base tree` 404s on 08-17, 13:31→16:14Z; zero since).
  **The 090 `75220928…` was overtaken and never returned a verdict — what it would have
  answered is still unknown: which component routed a no-repo site to git, and what changed at
  13:31.** If it recurs, that is the question; do not retrofit the roll as the cause.
- **B3f structural checks, volume MEASURED**: `head_essentials_missing` **247 / 8 sites**,
  `dead_internal_link_live` 6, `canonical_mismatch` 4; `structured_data_invalid` and
  `sitemap_entry_dead_live` at **0 — deliberately NOT interpreted** (clean vs never-exercised
  look identical; separating them is open work). All flag-only: a backlog to triage.

## 5. Next

1. **Owner sign-off on Phase C** — the cost baseline above is the number to pace Phase E waves
   against. This is the gate the plan names before any fleet dispatch.
2. Finish the pilot: the HITL queue, then the 3 component/rerender failures, then the
   `{{end}}` template leak (which is a platform bug, not a site one — worth its own file).
3. **A Phase-C-complete SUMMARY** — the series\' next inflection; the 2026-08-17 one predates
   the artefact proof.
4. Then Phase D decisions / Phase E waves per `PLAN_2026-08-12_fleet_buildout.md`.

## 6. Files of record

`PLAN_2026-08-12_fleet_buildout.md` · `SUMMARY_2026-08-17_pilot_built_and_the_machinery_proved_itself.md`
· `NOTES_portfolio_positioning.md` (newest at bottom — the 2026-08-18 entry has the artefact
proof and both corrections) · `README_where_we_are.md` · `MISSION_2026-08-17_…md` ·
`SEED_2026-08-17{,b}_…sql`. Migrations `432/433/434/441` (+ROLLBACKs), all applied and
council-approved. Register: `docs026_concept_register/register/directory-pipeline.md` (DIR-001).
RFC: `architecture_review/RFC_031_…md`. Closed: `bugs_closed/292_…md`.
