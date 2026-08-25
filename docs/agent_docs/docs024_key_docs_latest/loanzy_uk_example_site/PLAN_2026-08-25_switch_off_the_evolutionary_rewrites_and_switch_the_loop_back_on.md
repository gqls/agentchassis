# PLAN — switch off the "evolutionary" rewrites, switch the improvement loop back on

**Written 2026-08-25** by the `loanzy_uk_example_site` lane at the owner's request: *"turn off the
evolutionary aspect of the improvement loop for now — the bit that keeps rewriting pages that have
been judged to be good for the sake of aspirational improvements. It should stay because it's a
great thing but it is causing too many bad/unexpected renders, so switch just that bit off and turn
the improvement loop back on."*

**Status: PLAN. Nothing live has been changed.** The one migration Phase 1 needs is written and
rehearsed as a `_HOLD` file (never taken by the runner); applying it is your word. Every figure
below carries its date; every claim not measured says so.

---

## 0. The answer in four sentences

The improvement loop has two kinds of reviewer. The mechanical ones find **defects** (a broken link,
an empty section, a 404 asset) and their fixes are wanted. The four LLM ones — the design audit, the
strategic site review, the offer analyser, the brief-fidelity auditor — file **opinions** about pages
that already work, and the router turns those opinions into page **regenerations**. The
"evolutionary bit" is exactly those four seats' findings being dispatched; the plan takes the four
seats off the loop's path (one edge, reversible), turns the sweep back on, and then — with a small
Go change that is already written and tested — brings the seats back as **verdicts that dispatch
nothing** until a person releases them.

---

## 1. What the improvement loop IS (plain terms)

It is a scheduled sweep. A task called `improvement-sweep` fires every 15 minutes, picks **one**
site (the one least recently touched, skipping any with a build in flight or more than 50 open
build items), and runs the `improvement-loop` workflow on it. That workflow, in order:

1. refreshes the news-feed and directory recommendations for the site (deterministic, no LLM);
2. fingerprints the site (a hash of every rendered component, the palette and the chrome) and
   decides whether an audit is due — due if the fingerprint changed or 14 days have passed;
3. if due, runs the **seats** — first the three mechanical ones, then the four LLM ones;
4. records the audit against the fingerprint;
5. **promotes** every `detected` finding on the site to `triaged` so the dispatcher can take it;
6. queues a re-render of the site and calls the dispatch loop, which runs the fixers.

It has been **off since 2026-08-17 12:30Z** (`scheduled_tasks.improvement-sweep.enabled = false`).
It ran 08-13 → 08-17 and was hand-fired once on 08-22 18:3xZ. `[MEASURED 2026-08-25]`

## 2. What "the evolutionary bit" IS, measured

The four LLM seats do not repair anything themselves. Each ends in the same shared action,
`write_audit_findings`, which **routes a finding by its category to a handler**. The routes that
matter here (`platform/orchestration/actions/write_audit_findings_action.go`, Rule 4):

| the seat says | becomes | handled by | what the handler does |
|---|---|---|---|
| content / structure / differentiation on an existing page | `content_rewrite` | `page-build-handler` | **regenerates the page** |
| gap on an existing page | `needs_content_page` | `page-build-handler` | **rebuilds the page** |
| gap / content on a page that does not exist | `needs_content_planning` | `content-gap-planner` | plans **new** pages |
| tone | `needs_copy_edit` (since 08-24; was `tone_shift` → regenerate) | `copy-editor` | proposes, parks for a human |
| colour / typography | `needs_design_review` | `webdesign-agent` | analysis only |
| spacing / responsive | `spacing_fix` / `responsive_fix` | `component-template-fixer` | template surgery |

**How much of this there has been** `[MEASURED 2026-08-25, site_work_items UNION archive, lifetime]`:
from the `design-audit` source alone, **976** `content_rewrite`, **399** `needs_content_page`,
**964** `needs_content_planning`, **26** `tone_shift`, **1,197** `needs_design_review`. Since the
loop's 08-13 relaunch the five newer sources (visual-design, content-quality, site-review,
offer-analysis, brief-fidelity) filed about **150** items, of which the regenerating kinds
(`content_rewrite` / `needs_content_page` / `needs_content_planning` / `tone_shift`) are the majority.

**What one of them does to a page that was fine**: `bugs_open/238` — a `tone_shift` on
finetuning.uk's homepage regenerated the case-studies grid, dropped every image URL, and served five
empty `<img src="">` on the live page. Nobody asked for a rewrite; the loop had been fired to
serve five images.

**"Pages that have been judged to be good"** — the loop has no memory of a page being good. It
audits by fingerprint and by a 14-day clock, so a site that passed last fortnight is re-audited this
fortnight and any new opinion is dispatched. There is no per-page or per-site "approved, leave it"
switch anywhere in the loop (`IMP-006` in the register proposed one four times since March; it was
never built).

## 3. The rule this plan proposes

**Defects dispatch; opinions record.** A finding from a mechanical check (a predicate that could
have come out the other way) may be dispatched to a fixer as today. A finding from an LLM seat is a
**verdict**: it is written down as a row, with the handler it *would* have gone to preserved on
the row, and it dispatches nothing until a person (or a later deliberate migration) releases it.
The seats stay — they are the site acceptance council the owner asked for this morning
(`REFERENCE_2026-08-25_site_acceptance_council.md`) — but their output changes from *work* to
*evidence*.

## 4. How the current loop measures against that rule

- **Every LLM finding is dispatched, by two doors, not one.** The loop's own `triage_findings`
  step promotes *every* `detected` row on the site (no type filter — `triage_detect_items_action.go`).
  And a second promoter, `detected-item-promoter` (live, every 15 minutes since 08-15), promotes any
  `detected` row whose (item type, handler) pair has ever completed — which every regenerating pair
  has, thousands of times. `[MEASURED 2026-08-25]` **26 LLM-audit rows were promoted between 08-20
  and 08-24 while the sweep was OFF.** So "the sweep is off" was never "the rewrites are off";
  it was "no new opinions are being filed".
- **Nothing is parked at `detected` today** (`[MEASURED 2026-08-25]` 0 LLM-audit rows at
  `detected`; 59 at `needs_human_review`, 16 complete, 14 cancelled, 6 deferred, 6 wont_fix, 2 failed).
  The promoter has drained the queue. Turning the sweep on with the seats in place would refill it.
- **Seven of the eight seat calls fail open** (the peer lane's measurement, same day): a seat that
  errors or times out routes exactly where success routes, and the eighth
  (`call_completeness_discovery`) jumps past four seats to triage. The record that would say "the
  offer analyser failed on site X" lives in `orchestration_states`, which is reaped in ~25h. So a
  clean audit can be recorded over seats that never ran. This is not the rewrite problem, but it is
  the same loop and the RFC carries it.

## 5. The plan, in phases

### Phase 1 — take the four seats off the path; sweep back on. *Config only. Today, on your word.*

**The change:** one edge. `call_completeness_discovery.next_step` moves from `spawn_design_audit`
to `record_audit_pass`. The eight LLM-seat steps stay in the workflow, inert and unreachable —
nothing is deleted, and the rollback is the same edge flipped back. Migration:
`docs/agent_docs/sql_for_agents/619_improvement_loop_bypasses_the_llm_audit_seats_HOLD.sql`
(+ `_ROLLBACK`). **Rehearsed 2026-08-25 against the live row inside a rolled-back transaction:**
drift guard passed, edit + verify passed (31 steps kept, 0 dangling edges), live row untouched.

**Then, and only then, the sweep:**
```sql
UPDATE scheduled_tasks SET enabled = true, updated_at = now() WHERE name = 'improvement-sweep';
```
(The `vigilant_designer_offer_analysis` lane is holding this switch until ordering is settled
between us; I will tell them this plan is the ordering. Migration `389` re-enabled the sweep the
same way on 08-19, quoting you: *"it will be expensive so I am wary of costs."*)

**What then runs, per site, per 15 minutes:** enrichment → fingerprint → quality, design and
completeness discovery (algorithmic, no LLM) → audit recorded → promotion → re-render → dispatch.
**What no longer runs:** the design audit, the site review, the offer analyser, the brief-fidelity
auditor, and therefore every `content_rewrite` / `needs_content_page` / `needs_content_planning`
they would have filed.

**What it costs `[INFERRED, not measured]`:** the discovery seats are free; the fixers a mechanical
finding routes to are not always — `empty_sections` files `needs_content_page` at
`page-build-handler`, which *does* regenerate a page, and that is a defect fix you want. The
re-render step is the bulk of the wall-clock. One site per firing → the 31 active sites in about
eight hours, then the 14-day cooldown means most firings find nothing due.

**How we will know Phase 1 did what it says (run after the first hour):**
```sql
-- 1. the sweep is firing and completing
SELECT enabled, last_triggered_at, last_completed_at FROM scheduled_tasks WHERE name='improvement-sweep';
-- 2. zero new LLM-audit rows since the switch-on (the seats are bypassed)
SELECT count(*) FROM site_work_items
 WHERE created_at > '<switch-on time>'
   AND spec->>'audit_source' IN ('design-audit','visual-design-audit','content-quality-audit','site-review','offer-analysis','brief-fidelity-audit');
-- 3. the four seats' LLM call counts stay flat
SELECT agent_type, count(*) FROM llm_call_log
 WHERE created_at > '<switch-on time>'
   AND agent_type IN ('visual-design-auditor','content-quality-auditor','site-review-agent','offer-analyser','brief-fidelity-auditor')
 GROUP BY 1;
-- 4. what the mechanical seats DID file
SELECT source, created_by, item_type, handler_agent, status, count(*) FROM site_work_items
 WHERE created_at > '<switch-on time>' GROUP BY 1,2,3,4,5 ORDER BY 6 DESC;
```
Query 2 must be **0** and query 3 must return **no rows**. A non-zero in either means a seat is
reachable by a route this plan did not see — stop and read `spec->>'audit_source'` on the rows.

### Phase 2 — make "record, don't dispatch" a real mode. *Go; council; needs a roll. Written today.*

`write_audit_findings` gains an opt-in setting, **`filing_mode: record`**. In that mode every
routable finding is filed as the **same row** — same type, same dedup key, same spec — but parked:
status `deferred`, handler blank, the handler it would have gone to kept in `spec.routed_handler`,
who parked it and why in `spec.deferred_by` / `spec.deferred_reason`, and a release recipe on the
row. Both promoters refuse such a row by construction (one needs a handler, the other needs
`detected`). A typo in the setting is an **error**, never a silent dispatch.

Written and tested today (`platform/orchestration/actions/write_audit_findings_action.go`,
`..._filing_mode_test.go`, four tests, package green). It goes to the council with the three new
mechanical seats below as one submission. **Inert until the next chassis roll.**

Also in Phase 2, three of the four missing acceptance-council seats, as discovery checks that are
**registered but not enabled** until a migration names them (the estate's safe-rollout pattern):
`build_prerequisites`, `heading_promise`, `structure_floor`. All flag-only: they file verdict rows
and can never dispatch. Definitions are in `RFC_056`.

### Phase 3 — the seats back on, as verdicts. *Config only; HOLD until the roll and the verdict.*

One migration: name the three checks in the completeness agent's `checks` array; create the fourth
seat (`reader-experience-auditor`, an LLM seat with `filing_mode: record` from birth); put
`filing_mode: record` on the five existing LLM seats' `write_audit_findings` steps; restore the edge
Phase 1 moved. From then on the loop runs the full acceptance council after the fact, every seat's
finding is a row, and **nothing is rewritten unless a person releases the row** (the recipe is on
the row: `UPDATE site_work_items SET status = spec->>'routed_status', handler_agent =
spec->>'routed_handler' WHERE id = … AND spec->>'filing_mode' = 'record'`).

**Also in Phase 3, the fail-open surface** (§4, third bullet): each seat call gets a tiny
"record that this seat failed" step (a `deferred` `capability_gap` row per site per seat, the
loop's existing `record_not_converging` shape) and the audit is *not* stamped as passed when a seat
failed. Config only; carried by the same migration so it lands with the seats.

## 6. What this plan does NOT cover, named so it is not mistaken for covered

- **The render-audit rotation is a second, separate source of bad renders and it is ON.**
  `site-render-audit-rotation` (hourly) files `contrast_failure` at `css-patch-agent`;
  `[MEASURED 2026-08-25, last 14 days]` **239** completed, 15 cancelled, 13 at human review.
  `bugs_open/198`, `/390` and `/396` (the CSS one) are its damage. It is not part of the improvement
  loop and switching the loop's seats off does nothing to it. Owned by the `390` lane; **a decision
  for you, not this plan** — pausing it is one row (`enabled=false` on that task).
- **The 37 disabled one-shot tasks** (`offer-analyser-oneshot-*`, `oneshot-*-discovery-*`) must
  stay off. They are site-pinned with no stopping condition; re-enabled they re-run one site every
  five minutes for ever (the peer lane's measurement, same day).
- **Copy quality** is the `copy_quality_two_stage` lane's; the reader seat judges purpose, not prose.

## 7. The choices that are yours

1. **Apply 619 and switch the sweep on** — Phase 1, today. (My recommendation: yes; the seats are
   preserved, the rollback is one edge, and the verification queries say within an hour whether a
   seat is still reachable.)
2. **The render-audit rotation** — leave on, or pause until the `390` lane's commits 2 and 3 land.
   I have no measurement that says the appended `!important` rules of `616` are worse than nothing;
   I do have `396`'s that a design run erases them all and the items stay `complete`.
3. **Phase 3 — the seats back on as record-only** once the roll ships. This is the acceptance
   council. The cost is the seats' LLM calls (the same cost as before 08-17) with the rewrites
   removed; the benefit is a verdict row per seat per site, which is the census you asked for.

## 8. Falsifiers — what would show this plan is wrong

- Query 2 in §5 non-zero after Phase 1 → an LLM seat has a route this plan did not see.
- Bad renders continue at the same rate with the seats bypassed → the cause was the mechanical
  fixers or the render-audit rotation, not the opinions; §6 first bullet becomes the plan.
- A record-mode row being dispatched → one of the two promoter doors has changed; the tests in
  `write_audit_findings_filing_mode_test.go` name both doors and one of them is no longer the door.
- The sweep enabled and the loop filing nothing routable → `REFERENCE` §11 first bullet.
