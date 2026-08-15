# 282 — validate's name resolver silently drops every tool section the planner places (407's missing half)

**Filed:** 2026-08-15, loancalculator rebuild lane. **Status:** OPEN.
**Diagnosis trail:** 090 run `4a02a4e1-3972-450a-8163-28d6bb0a79fd` (verdict
UNVERIFIABLE at budget, disjunctive hypothesis + named next_scope) → next_scope
walked first-hand same session; the disjunction closed with the drop site cited.
Owner-ruling 2026-07-31 note: the 090 ran BEFORE this root cause was asserted;
this file is the completion of its own named evidence requests.

## Mechanism (the whole chain, each link cited)

1. **The menu shows the tools.** Migration 407 (PLAN-049): `load_components`
   includes `component_level='tool'` rows when the site's structure spec has
   `plan_includes_tools='true'` AND the component is deployed on that site.
   Proven at run corr `2f74a975-1a87-40a8-af88-a9bd2ecc1510`:
   `available_components`=151; diagnosis data_request `menu_has_tool1|menu_has_tool2
   = true|true`.
2. **The planner PLACES the tools.** `llm_call_log` id
   `ca3c22f4-5e4c-4ccf-9d4a-02719d976c8e` (`plan_site`, 2026-08-15 14:25) —
   raw response proposes `"sections": ["hero","tool-loan-repayment",…]` for
   index and `"hero","tool-<fn>","ported-prose","faq","tool-cta"` for tool
   pages (application-tracker, overpayment, settlement read directly; the
   response text is the artefact).
3. **Validate's resolver cannot resolve them.** `loadComponentNameResolver`
   (`platform/orchestration/actions/v3_site_actions.go:3804-3809`) builds
   `validFunctions` from `WHERE component_level IN ('section','element')` —
   tool-level functions are structurally absent, so `resolve()` returns
   ("", false) for every proposed tool section and the section is dropped from
   the plan write. No error, no tell; the write succeeds.
4. **The persisted plan lacks the tools.** Plan `dcbae4df` (2026-08-15
   14:25:49): 0 of the 12 locked tool functions appear in `site_plan_sections`;
   index reads `hero,info-card-grid,tool-list,guide-list,call-to-action`
   (5 sections — the raw response proposed 6).

The measured effect that surfaced it: a `recompose_pages` release of the 12
tool-carrying pages on loancalculator.co.uk (0162cde4…) produced compositions
that would build every calculator page WITHOUT its calculator. Nothing live was
damaged — the tool-role review gate (TP-004) + the 12 permanent locks held.

## Why this is 407's missing half, not a new seam

407 widened the MENU under an opt-in flag and added the placement gate; the
resolver that validates proposed names against reality was never widened to
match. Two lists that must agree, maintained in two places — the
`idx_swi_dedup ↔ workItemTerminalStatuses` lockstep class exactly.

## Fix candidates (ranked by what closes the door)

1. **Mirror the menu's own predicate in the resolver** (preferred): include
   tool-level functions in `validFunctions` for the site being planned iff the
   site's `plan_includes_tools` flag is on AND the component is deployed on
   that site — the same subquery `load_components` uses, factored into ONE
   shared helper so the two surfaces cannot drift again (the lockstep made
   structural). Needs the site_id plumbed to `loadComponentNameResolver`
   (callers have it).
2. Widen the resolver to all active tool-level functions unconditionally —
   smaller diff, but lets a planner place a tool the menu never showed it
   (hallucination path back open); rejected by the same reasoning that made
   407's menu opt-in and site-scoped.
3. Prompt-level workaround (tell the LLM to use a different marker) — rejected:
   markers invent a second vocabulary; the estate resolves REAL function names.

Whoever fixes: council-gate the change (platform code, shared seam — but note
RFC_022's narrowed trigger: this mirrors an existing opt-in whose consumers are
enumerable by the flag query). After it ships and a replan runs, verify at
`site_plan_sections` (the 12 functions present on their pages) and only then
work the held D2 tickets (11 `owned_page_review` + `needs_page:index` on
loancalculator — see the lane's HANDOFF 2026-08-15).

## How to verify a fix

Re-fire the lane's `phase2_recompose_26.sh` (12-page scope). Expect: the 12
locked tool functions in `site_plan_sections` on their own pages; zero
RECOMPOSE_INTENT_NOT_REALISED rows; locks 12/12 untouched. The negative control
already exists: plan `dcbae4df` is the no-fix baseline.
