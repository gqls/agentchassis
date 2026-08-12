# PLAN — fleet build-out: from 4 hand-built sites to a full-featured .uk portfolio

Continuation of this workstream, not a new one. Supersedes nothing in
`PLAN_2026-07-31_differentiation_axes.md` (that plan is positioning/twin-splitting;
this one is build execution). Full context, ground-truth findings and the plan-mode
file this was written from: see `README_where_we_are.md`'s 2026-08-12 entry and
`NOTES_portfolio_positioning.md`'s same-date entry for the measured evidence.

## Why this plan exists

Owner asked this session to move toward all ~152 portfolio domains live, 20+ pages
each, with images, copy, tools, guides, infographics/graphs where appropriate,
newsfeeds, and directory listings sourced from live web search. Ground-truthing
(three parallel research agents + one design agent, all against live DB and code)
found the situation more foundational than the 2026-08-11 handoff implied — see
NOTES for the measured figures. In summary: only 4 of 152 domains have ever been
built at all, all 4 by hand or adoption, not by the generative pipeline; the
pipeline's page planner is capped at ≤20 pages (a ceiling, not a floor); newsfeeds
work end-to-end; directory listings have a proven mechanism but zero wiring for
finance/insurance; the structural-validity gate and bug 161's fleet-wide fact-
discipline fix are both still open.

## Owner rulings this session (govern everything below)

1. **Nothing new goes through the pipeline until the structural-validity gate AND
   bug 161's fix are both live** — not even a single pilot domain.
2. Newsfeeds and directory listings: wanted from day one, not a later pass.
3. Generated tools stay on manual per-site review (no gate-loosening work this phase).
4. Graphs: `evidence-chart` bar charts only, wherever a site has verified cited
   data — no new infographic/chart-generator build this phase. (Spot-checked
   2026-08-12 against `oufe.com`'s live pages — see NOTES; claim holds.)
5. Directory listings: non-price facts only (FRN, regulator status, product types,
   underwriter, established year) — no APR/premium figures, for financial-promotion
   compliance reasons.

## Phases

**Phase 0** — housekeeping: re-check ownership/tree state before each work session
(this tree runs 3+ concurrent sessions); this plan + rulings recorded here.

**Phase A — guardrails (blocking)**:
- A1. Structural-validity gate: new
  `platform/orchestration/actions/discovery_checks/check_site_structural_validity.go`,
  four checks (dead_internal_link_live, canonical_mismatch, structured_data_invalid,
  head_essentials_missing), generalising
  `loanandmortgagecalculator_couk/verify_site.py`. Standing rotation dispatch
  (clone of migration 372's pattern), flag-only routing initially, SEO-004 register
  entry, council-submitted same commit.
- A2. Bug 161 RFC — the long pole. `source.artifact_check` mechanism, architecture
  review round-trip. Starts in parallel with A1, not after it.
- A3. Fidelity dial — no action, already ruled acceptable inert.
- A4. Raise `build-site-planner`'s `max_pages` from 20 to ~24–25.
- A5. Confirm bugs 251 (canonical)/252 (og:/lang) have shipped and rolled, before
  enabling `canonical_mismatch` auto-repair and before Phase C.

**Phase B — directory-listing producer for finance/insurance**: reuses
`directory_claims.go`'s verified-claim mechanism unchanged; new concept-register
entry (prerequisite — none exists today); one kind per provider class (~6-10 kinds,
not one filtered kind — the filtered-kind shape would trip architecture review);
new `SEED_finance_directory_researcher_agent.sql`; new deterministic
`evaluate_directory_features` action mirroring `evaluate_news_feed` (today nothing
writes the opt-in, so nothing would ever fire without this); fix the publish-
trigger's kind-blind find-sites query; sequence components → researcher → claims
populated → only then let domains through.

**Phase C — pilot**: one real, unambiguous proposition (not lendzy, not a
HOLD/undecided one) through the fully-gated pipeline end to end. Manually compute
real LLM-call and image cost from `llm_call_log`/`assets` (no aggregate cost figure
exists anywhere in the platform). Owner sign-off before Phase E.

**Phase D — remaining owner-only decisions**: `loanzy.uk` (L9) conflict with the
`webdesign` lane; B8/B9/I10 hold/undecided status; build order across the ~140
remaining domains (register itself says this "remains the owner's commercial call").

**Phase E — fleet dispatch in waves**, paced against the Phase C cost/time
baseline, manual tool-review queue worked per site, register status tags kept
current as domains land.

**Explicitly out of scope this phase**: general infographic/chart generation,
directory pricing/rates, loosening the tool review gate, backfilling the 4 existing
sites to the new bar.

## What this plan does NOT decide

Same three items `PLAN_2026-07-31_differentiation_axes.md` already named as
undecided (twin build order within Tier C pairs where still open, overall build
order, serving-infrastructure specifics) — this plan doesn't re-litigate those.
