# PLAN — bugfix_428_planner_deferral

Started 2026-09-02, by the session working bugs_open/427 (renamed at start; kept the
name across both bugs since they share the boxingonline.com context). Picked up
`gap_planner`'s diagnosis in bugs_open/428 and carried it through to a shipped fix —
though a narrower one than the bug's own §6 first proposed.

## 0. The critical correction that reshaped this bug's fix

Bug 428 diagnosed the site-planner LLM deliberately omitting strategy-named
`entity-page`/`entity-directory` roles (~76% of the time when named), citing its own
"final say" license. Its §6 candidate 2 — "wire the existing 13 deferred verdicts to
a dispatcher, lowest-risk, no planner-prompt change" — turned out to be the opposite
of lowest-risk: every one of those 13 rows carries `spec.filing_mode='record'`,
which is `write_audit_findings_action.go`'s RFC_056 (2026-08-25) — a deliberate
owner-ruled circuit breaker that stops LLM-audit-seat findings from auto-dispatching
as page rewrites, built specifically because an earlier auto-dispatch of exactly
this finding class destroyed live content (`bugs_closed/238`: five broken images and
an unrecoverable rewrite of the case-study copy underneath them, finetuning.uk).

This was checked at the artefact (`spec->>'filing_mode'` on every one of the 13/16
candidate rows), not assumed from the shape. `gap_planner` independently confirmed
after I flagged it, and separately caught that the same rows also can't just be
dispatched wholesale for a second reason: `bugs_open/206`'s builder mechanism is
live, but `business_intel` only covers 3 verticals, none matching any of the 13
affected sites — so most of them would hit a live builder with nothing to build
from anyway (that's `bugs_open/427`'s territory, not this bug's).

**Put to the user given the stakes (this touches a mechanism the owner personally
ruled on 8 days prior): fix the two safe, independent pieces (prompt formatting +
license tightening) and build a human-reviewed release surface for the existing
verdicts — explicitly NOT an automated dispatcher.** Approved; this is what shipped.

## 1. Shipped, piece 1 — the prompt (migration 687)

Two edits to `build-site-planner`'s live `plan_site` prompt, applied same-session:

1. **Formatting** (bug 428 §6 candidate 3): `{{.site_specs.specs.strategy}}` →
   `{{toJSON .site_specs.specs.strategy}}`. `toJSON` is an existing, already-live
   template function — no Go change, no image roll needed, not a `_HOLD`. Confirmed
   this is NOT the omission's cause (the model reads the ugly dump correctly
   regardless, per bug 428 §2's own `llm_call_log` citation) — purely a legibility
   fix so the next thread auditing "did the model see X" can read the prompt text
   directly instead of pulling `llm_call_log.prompt_rendered`.
2. **License tightening, softer form** (bug 428 §6 candidate 1's lower-risk end):
   the "FINAL SAY... note why in strategy_notes" rule now requires the model to name
   every omitted `recommended_page_types` entry individually, with a real per-type
   reason — not a generic "keeping it lean" covering any number of drops (which is
   literally what boxingonline's own call did: named both roles, gave one shared
   rationale for dropping both). **Deliberately not the harder option** (a
   `validate_site_plan` hard failure) — the model keeps its licensed final say in
   full; this only makes the existing obligation concrete enough to audit.

Applied and verified live at the artefact (both new strings present in the live
`prompt_template` after running). `snapshot_agent()` taken first;
`agent_definitions_bak_687` + a paired `_ROLLBACK.sql` are the recovery path.
Submitted for council review (`3f9cdfea-7287-4ab3-afad-9c386fbb7365`).

## 2. Shipped, piece 2 — the human-reviewed release surface

Respects RFC_056 exactly as designed (`spec.release_recipe`: "a human or a later
migration releases it") rather than routing around it:

- `HandleReleaseRecordVerdict` (new admin endpoint): releases ONE row via a
  parameterised UPDATE performing the identical operation the stored
  `release_recipe` text describes — never executes that stored string as SQL.
  Gated in the WHERE clause on `status='deferred' AND filing_mode='record' AND`
  both routed fields present; structurally cannot touch anything else. Requires a
  non-empty `released_by`, checked before any query runs. Stamps
  `filing_mode='released'` + who/when/why, so a row can never be double-released.
- `HandleListWorkItems` gains an optional `filing_mode` filter, so these rows are
  findable among the ~1,284 general `deferred` rows fleet-wide (measured by
  `gap_planner` during this bug's own investigation) — previously only reachable
  by a raw SQL query.
- Admin dashboard: the `deferred` status was **entirely absent** from the filter
  dropdown before this — no route to these rows existed in the UI at all. Added
  it, added a "Record verdicts only" toggle, added a "Review & Release" button
  gated on the exact same predicate the backend enforces, prompting for a
  reviewer name and reason first (mirrors the existing Resolve button's `prompt()`
  pattern rather than a new UI convention).

Every WHERE-clause guard is pinned by its own test (mutation-tested: removing the
`filing_mode` predicate from the handler fails its corresponding test). Committed,
submitted for council review (`38be9226-d5b5-48b7-9b87-20efbaf3dec3`).

## 3. Explicitly not built

- **An automated dispatcher/promoter for `filing_mode='record'` verdicts** — bug
  428 §6 candidate 2 as originally worded. Would reverse RFC_056. Needs an owner
  decision, not a council submission; not something either this session or
  `gap_planner` builds unilaterally. Both of us flagged it to our respective users
  rather than proceeding.
- **A `validate_site_plan` hard-failure enforcement of the license** — the harder
  end of candidate 1. Left for a future round if the softer prompt change proves
  insufficient (see §4).

## 4. How to tell if this actually worked

No before/after A-B test was run — the prompt change is live for every future
`plan_site` call, and the honest read is "wait for real calls and check the
`strategy_notes` shape," not "this is proven." Follow-up worth doing in a future
session: sample the next N `plan_site` calls against a strategy naming an entity
role, and check whether `strategy_notes` now names the specific `page_type`
per-omission rather than a blanket justification — the same measurement bug 428
§7 already specifies for verifying candidate 1.

## 5. Cross-references

- `bugs_open/428` — the filed bug, diagnosis, and both corrections (this session's
  RFC_056 finding, `gap_planner`'s 206/data-coverage finding) live there.
- `bugs_closed/238` — the incident that motivated RFC_056; read before ever
  proposing to loosen `write_audit_findings_action.go`'s record-mode gating.
- `bugs_open/427` / `bugfix_427_event_render` — the sibling bug this session was
  already working; boxingonline.com is the shared motivating site, and 427's own
  candidate #3 (entity-directory render target) is downstream of whether this
  bug's planner-licensing fix actually changes future behaviour.
- `bugs_open/206` — the entity-directory builder; live but data-starved for every
  site this bug's sample affects (per `gap_planner`'s correction to 428 §5).

## 6. Status, 2026-09-02 (later same day) — see bugs_open/428 §10/§11 for the live account

Both council verdicts came back APPROVED. Fresh chassis build confirmed live for
`agent-chassis` and `core-manager` (`ebf27c60377f`) — the release-surface
BACKEND is genuinely live. Gap found: the admin-dashboard FRONTEND (the actual
button) has not been redeployed in 170 days, so the surface is API-only until
someone runs `make admin-dashboard`. Logged as a near-miss in
`docs024_key_docs_latest/WRONG_CALLS.md` (2026-09-02, "committed... stood in
for usable"). **This file is not being kept as the live status record going
forward — bugs_open/428 §10/§11 is; read there.**
