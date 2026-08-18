# 083 — discovery findings written as `status='detected'` never reach a handler: the promoter runs only inside a task disabled since May

> **CONTRIB 2026-08-17 (`bugs_open/296` lane) — a SECOND stranding cause, measured: a type that
> HAS a live, correct machine closer, whose 225 rows will still never move.**
> Relevant to the "named owner per type" half of today's owner decision, because this type would
> pass an is-there-an-owner test and strand anyway.
> **The population:** `contrast_failure`, **225 rows, all `deferred`** — inside the *"467 rows
> across 6 statuses"* ceiling already noted in the 08-09 notice below, and the largest single
> block of it.
> **What is different from this file's thesis.** 083's model of stranding is *absence*: no handler
> owns the type, so nothing ever closes it. `contrast_failure` is not that shape. It has an
> audit-path retraction (`work_item_retraction.go` + `retractResolvedContrastFindings`) that is
> **built, shared, wired to this exact type, deliberately reaches `deferred` rows** (WII-016 —
> `deferred` is not in `workItemClosedStatuses`), and **confirmed live in the running chassis**
> (binary probe, v1.0.1305, with controls). It has closed **zero** — and that zero is honest twice
> over:
> 1. It has never had one *opportunity*. Committed 2026-08-12 20:03; the last render audit of any
>    site holding these rows was 2026-08-11 12:04. The re-audit rotation is hourly, **`LIMIT 1`
>    site**, 7-day eligibility, and all 23 eligible sites were swept 08-10/11 — so **no site was
>    due between 08-11 12:04 and 08-17 14:54**.
> 2. When it does arrive it will correctly close **almost none of them, for ever**, because the
>    findings are **TRUE**. Four parked pairings on four sites, re-measured 2026-08-17 in a live
>    headless browser, **all reproduce at exactly their parked ratio** (1.06:1, 1.00:1, 3.35:1,
>    1.76:1), and live failure counts now *exceed* parked counts on each of those pages.
>
> **So the check "does this type have an owner/closer?" returns YES here and the rows strand
> regardless.** The stranding cause is not a missing closer but a **correct decline** — and a
> correct decline and a silent no-op are indistinguishable from outside, which is this file's own
> shape one level up. A time limit keyed on "held too long" would fire on all 225 of these and be
> *right to*, but what they need is a **fixer or an explicit accept**, not a promotion:
> `css-patch-agent` has still never processed a single work item, and migration `389` parked them
> precisely to stop 225 ungraded completions being minted.
> **Full working, every query re-runnable:** `bugs_open/296` §8. Defect population itself:
> `features_open/026` (primary-as-ink), not `bugs_closed/122`'s article-body ink.

> **CONSUMER NOTICE 2026-08-09 — `claims_unverified` is no longer parked-for-ever either, and this
> is a FACTUAL-claims surface, so read this one even if you skimmed the last.**
> The review-queue sweep gained a `claims_unverified` revalidator (`4030cadb9`, CQ-021, council
> `b67eb26a-14ef-45d7-b755-3e489fd57ef0`, inert until the next chassis roll).
> **What changed about your guarantee:** an item of this type can now be CLOSED without a human,
> when a re-scan of the page against the site's *current* `evidence_base` register finds no
> unsupported claim. Retraction never edits copy and never dispatches a rewrite, but it IS a
> machine closing a human-review row about a **truth** claim. The CLOSER census run first returned
> ZERO rows, and no row carries a `deploy_result` block — so neither a handler nor a fix pipeline
> owned this type, which is the condition this file is about.
> **Live population when this was written: 23 items across 7 sites**, all `needs_human_review`.
> ⚠ **The register is a MOVING STANDARD and it is DATA** — adding a fact row retracts an item with
> the copy untouched. A `resolved` stamp is not proof the page was corrected.
> ⚠ **AND A CEILING THIS FILE SHOULD KNOW ABOUT:** the sweep only ever selects
> `needs_human_review` and `unresolved`. Findings parked in `blocked`, `detected`, `triaged` or
> `deferred` — **467 rows across 6 statuses, measured 2026-08-08** — are reached by neither the
> sweep nor its coverage report. `image_url_404` (26 rows, 0 ever closed, no handler) is the worked
> example. Under diagnosis as `a174b184-dac2-47a1-95ca-df2d192e183a`; do not treat it as settled.
>
> **CONSUMER NOTICE 2026-08-08 — `voice_tells` is no longer parked-for-ever, and this file's tally of it will move.**
> **UPDATE 2026-08-09: this is now LIVE and has closed its first item** — `voice:ecfd0bfd-…`, page
> `ai-readiness-quiz`, retracted unattended by the 08:38:53Z scheduled sweep.
> The review-queue sweep (`revalidate_review_queue`) gained a `voice_tells` revalidator
> (`ef80216be`, council `4d430ca8-7e34-479a-95f3-71fdc12fdef6`, inert until the next chassis roll).
> **What changed about your guarantee:** an item of this type can now be CLOSED without a human,
> when a re-scan of the page against the site's own voice gate finds no tells. It is not the
> auto-rewrite `check_voice_tells.go`'s `fix` text forbids — retraction never edits copy and never
> dispatches a rewrite; it withdraws a finding the current page no longer supports. The CLOSER
> census run first returned ZERO rows (nothing had ever closed one), and this file's own
> classification of the type as *advisory / machine-fixable in principle* rather than *needs a human
> ANSWER* is part of why it was judged safe to drain. **Live population when this was written: 25
> items, all `needs_human_review`, all filed 2026-07-17, all on leopardessconsulting.co.uk.**
> ⚠ The voice gate is a MOVING STANDARD: a site that loosens its thresholds retracts items whose
> copy never changed. A `resolved` stamp is not by itself proof the prose was rewritten.
> Registered as CQ-020. Raised by the `bugfix_168_deployed_asset_path` lane.



**Filed:** 2026-07-26 · **Branch:** `086_experience_loop` · **Status:** OPEN, diagnosed with evidence, not fixed
**Severity:** high, structural and silent. Nothing errors. Detectors report success, rows are
written, and the work is never done. **98 items are parked fleet-wide.**
**Class:** dormant-machinery / fail-silent delivery gap — the same family as `bugs_open/071`
(a gate that detects every broken link then discards the finding) and `063`.
**Found by:** the bugs_open/049 session, while establishing why three months of correct
phantom-link detection had produced no change on any live site.

---

## Symptom

`phantom_internal_link` has been detected 22 times. It has been **fixed zero times, ever.**

```sql
SELECT item_type,
       count(*) FILTER (WHERE status='detected') AS stuck_detected,
       count(*) FILTER (WHERE status='complete') AS ever_complete,
       count(*) AS total
FROM site_work_items
WHERE item_type IN ('phantom_internal_link','unbuilt_internal_link','empty_internal_href')
GROUP BY 1;
```
```
 phantom_internal_link  | 18 |  0 | 22
 empty_internal_href    |  7 |  2 | 14
 unbuilt_internal_link  |  — |  — |  0     <- has never fired at all; see "second-order" below
```

Fleet-wide, **98 rows sit in `status='detected'`**, the oldest from 2026-07-14, spanning 21
item types — `page_rerender` (28), `undeployed_asset` (19), `phantom_internal_link` (18),
`empty_internal_href` (7), `empty_section` (4), and 16 others.

## Root cause — one promoter, and it lives inside a disabled scheduled task

Discovery checks write findings with `Status: "detected"` (the convention across
`platform/orchestration/actions/discovery_checks/*.go`). The dispatch loop cannot see that
status: both `claim_work_item_action.go:102` and `load_work_item_actions.go:559` filter
`status IN ('triaged','approved')`.

Exactly one thing bridges the gap — `TriageDetectedItemsAction`
(`platform/orchestration/actions/triage_detect_items_action.go:91-103`), whose own header says so:

> `TriageDetectedItemsAction promotes discovery findings from status='detected' to
> status='triaged' with pipeline='build' so the dispatch loop picks them up.`
> `Used by: improvement-loop agent (after all discovery agents complete)`

It is registered (`registry.go:819`) and it is correct — no type filter, it promotes everything
for the site. **It is simply never called**, because the only workflow carrying that step is the
`improvement-loop` agent, and the only thing that fires that agent is the `improvement-sweep`
scheduled task, which is **disabled**:

```sql
SELECT name, enabled, last_triggered_at FROM scheduled_tasks WHERE name='improvement-sweep';
--  improvement-sweep | f | 2026-05-02 10:11
```

The agent definition itself is alive and correctly configured — this is not a broken workflow,
it is an unscheduled one:

```sql
SELECT type, is_active FROM agent_definitions WHERE type='improvement-loop'
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;             --  improvement-loop | t
```

**Two other code paths write `status='triaged'` and neither helps**: `claim_work_item_action.go:186`
and `load_work_item_actions.go:917` both *release an already-claimed item* back to the queue when
an AI endpoint is unhealthy. Nothing else promotes `detected`.

So: **a detector that files a finding as `detected` is writing to a queue with no consumer.**

## Why this is not `bugs_open/033`

033 is the `needs_human_review` pile (376 rows) — items routed to a human surface that nobody
works. This is a different status with a different intended consumer: `detected` is meant to be
picked up by *machinery*, automatically, and the machinery is switched off.

The distinction matters because **`detected` was deliberately chosen over `needs_human_review`
precisely to avoid 033.** `bugs_closed/054`'s record states it plainly:

> On that basis `bugs_open/054` wired the new **chrome dead-control** finding into a
> *draining* pathway (`status='detected'` + `handler_agent='nav-link-fixer'`, the
> phantom-links convention) rather than adding a 125th row to the unread
> `needs_human_review` pile.

That reasoning was sound and the premise was false: the "draining pathway" does not drain, and
the "phantom-links convention" it copied has a lifetime completion count of zero. **Anyone
choosing `detected` today to avoid 033 is choosing a quieter version of the same problem** —
which is the real cost of this bug, because it keeps being chosen.

## Second-order effect: detector work ships and changes nothing

This is why it is worth fixing rather than noting.

`unbuilt_internal_link` (bugs_open/049 candidate 4) was built, tested, committed and is **live in
the running pod** — `docker.io/aqls/agent-chassis:v1.0.1165`, pod-grep count 2. It has **never
produced a single row**. Not because it is broken: because discovery only runs when a human fires
it, and the sites carrying its target class had not been swept. Last discovery item written:

```
finetuning.uk      2026-05-01     gaswholesalers.com  2026-04-25     vetcomparison.uk  never
```

Those were the three worst sites in 049's audit. Coverage and delivery fail together and for the
same reason: `improvement-sweep` was the periodic driver of *both* the discovery run and the
triage that follows it.

## Fix candidates

1. **Re-enable `improvement-sweep`** (`UPDATE scheduled_tasks SET enabled=true`). Live
   immediately, no image roll. ⚠️ **Owner decision, not a platform one** — it fires fix agents
   fleet-wide on a 180s interval and spends credits, and something disabled it deliberately on
   2026-05-02. `[UNVERIFIED]` why it was disabled; I did not find a record and did not assume one.
   Read that reason before flipping it.
2. **Give triage its own scheduled task**, decoupled from the fix loop: promote `detected` →
   `triaged` on a slow cadence for item types whose handlers are known-good, leaving discovery
   itself manual. Smaller blast radius than (1); it drains the existing pile without also
   re-arming fleet-wide auto-fixing.
3. **Refuse the write instead of parking it.** Make `insertWorkItem` reject (or loudly warn on)
   `status='detected'` when nothing is scheduled to promote it — the platform's own
   silent-empty-render class, applied to queues. Turns a silent park into a visible error at the
   point the finding is filed.
4. **Check the handler is real before celebrating the detector.** Related but distinct:
   `bugs_open/077` (detector predicate wider than its handler) and `078` (NULL handler_agent
   livelocks the dispatcher). A promoted `phantom_internal_link` routes to `page-build-handler`,
   and `[UNVERIFIED]` whether a whole-page rebuild actually repairs an LLM-authored bad href.
   **Do not enable (1) or (2) without checking that**, or the pile moves from `detected` to
   `failed` and nothing improves.

## How to verify a fix

1. `SELECT count(*) FROM site_work_items WHERE status='detected'` falls and **stays** fallen —
   a one-off drain that refills is not a fix.
2. `phantom_internal_link` reaches a non-zero `complete` count for the first time.
3. The items that complete are verified against the **live page**, not the row status —
   `complete` is not proof the work happened (016b, and `bugs_open/017`'s whole subject).

## Landmines

- **A zero completion count is not evidence the handler is broken.** Here it means the items
  were never dispatched at all. Distinguish "never claimed" from "claimed and failed" before
  blaming a handler: `claimed_at IS NULL` on all 18.
- **`approved` is a status no Go code ever writes** (recorded in `bugs_open/033`), so the
  dispatch filter `status IN ('triaged','approved')` is effectively `= 'triaged'`.
- Do not "fix" this by having detectors write `triaged` directly. That bypasses the triage
  stage's judgement and would auto-dispatch every finding the moment it is detected — including
  the ones a human should see first. The gap is the missing promoter, not the status.

## Related

- `bugs_open/049` — where this was found; its mechanism 2 detector is the worked example of a
  live, correct, never-fired check.
- `bugs_open/033` — the `needs_human_review` pile. Sibling, different status, different consumer.
- `bugs_closed/054` — chose `detected` believing it drained. The quote above is from its record.
- `bugs_open/071` — the build-time gate that detects broken links then discards them. Same
  fail-silent shape one stage earlier.
- `bugs_open/077` / `078` — handler-side hazards that fix candidates 1 and 2 must clear first.

---

## Corroborating instance, 2026-07-27 — the claims sweep is downstream of this too

Contributed from the oufe.com workstream; not a competing diagnosis, and no fix
attempted here.

`check_unverified_claims` — the post-deploy half of the claims layer, which scans
**stored `rendered_html` of live pages** and raises `claims_unverified` at
severity high — is reachable only via `quality-discovery-agent`, which is
dispatched only by `improvement-loop`, which is driven only by the
`improvement-sweep` scheduled task recorded here as **disabled since
2026-05-02**.

Consequence worth adding to this file's impact section: **the estate's only
automatic detector of published-content drift has effectively never run.** It is
not broken and it is not unwired — it has no cadence.

Observed on oufe.com 2026-07-26: a page shipped four promises of the site's own
infallibility and stayed live until a human read it. Had the site been armed with
patterns (`sql_for_agents/226`, and see `bugs_open/104` for why only 5 of 15 sites
are), the build gate would have blocked it — but for anything that reaches
production by another route, or drifts afterwards, this sweep is the only net, and
it is not in the water.

Two things that make the silence hard to notice, both relevant to this bug's
"findings never reach a handler" framing:

- every finding in this family terminates at `needs_human_review` with an **empty
  handler by design** (`check_unverified_claims.go:135-150`) — so "no handler" is
  correct behaviour here, and the queue surface (`bugs_open/033`) is what makes it
  actionable or not;
- the sibling that *does* have a fleet cadence, `evidence-freshness` / V4
  (`refresh_evidence_base_action.go:172-199`), sweeps every site with a register
  daily — which makes the claims layer *look* actively swept when only its
  fact-refresh half is.

---

## Re-measured 2026-07-27, post-roll (v1.0.1174) — the pile grew 61% in one day, and it is now stranding shipped code

Verification sweep, not a fix. Nothing here contradicts the diagnosis above; it sharpens
the cost and adds one consequence the original filing could not have seen.

**The sweep is still off and the pile is still filling:**

```sql
SELECT name, enabled, last_triggered_at FROM scheduled_tasks WHERE name='improvement-sweep';
-- improvement-sweep | f | 2026-05-02 10:11:07+00      (unchanged)

SELECT min(created_at) AS oldest, count(*) FROM site_work_items WHERE status='detected';
-- 2026-07-14 15:08:03+00 | 158
```

**98 → 158 in one day** (this file's own figure was measured 2026-07-26). The oldest row is
unchanged at 2026-07-14, so nothing drained — 60 arrived on top. Top types now:
`undeployed_asset` 35, `page_rerender` 28, `phantom_internal_link` 18, `needs_rerender` 13,
`needs_imagery` 10, `empty_internal_href` 7. `phantom_internal_link` still stands at **18
detected, 0 complete, lifetime** — unchanged, so the count is not the growth driver; the
growth is elsewhere and broad.

Note `75df951c9` (migration 233) already inserts *re-render* items as `triaged` rather than
`detected`. That is a correct local mitigation and it does not touch this bug: it changes
what one producer writes, while the missing consumer stays missing for the other twenty item
types.

**The new consequence: this bug now strands code that has already shipped.**
`bugs_open/093`'s fix (a stored-`content_data` stat audit, built into
`check_unverified_claims`) is **live in the running binary and has never executed**, because
that check is reachable only via `quality-discovery-agent` ← `improvement-loop` ←
`improvement-sweep`. Measured: `claims_unverified` has **0** rows live and **1** in
`site_work_items_archive`, dated 2026-07-17 — nine days before the fix shipped.

This is the second-order effect in § "Detector work ships and changes nothing", except the
first example (`unbuilt_internal_link`) was a detector nobody had swept a site for. This one
is a fix built *deliberately* to close a council-escalated gap, reviewed over six rounds, and
it cannot fire. **When estimating the value of fix candidate 1 or 2, count that too**: the
sweep being off is no longer only a backlog, it is now a silent tax on shipped work.

**Fix candidate 4's caution stands and applies here as well** — do not flip the sweep on
without first checking that the handlers for the top item types have a real remit
(`bugs_open/077`), or 158 rows move from `detected` to `failed` and the tax is unchanged.

---

## 2026-07-27 — fleet-wide measurement, contributed (brochure_component_library thread)

**Contributing measurements, not forking a diagnosis** — `who-owns.py 083` says OWNED,
and `bugs_open/115` (mine, same day) is a *symptom* of this file's mechanism, not a
rival account. What follows is the scale, which I could not find quantified here.

I arrived from the other end: the owner reported a live site looked wrong, and it turned
out our own brief-fidelity audit had filed three correct findings about it **three days
earlier** that nobody ever read, because nothing consumes that item type. That is one
item type. The fleet number is much larger.

### Every non-terminal work item, split by whether it can EVER be actioned

```sql
SELECT CASE WHEN handler_agent IS NULL OR handler_agent = '' THEN 'no handler named'
            WHEN NOT EXISTS (SELECT 1 FROM agent_definitions ad
                              WHERE ad.type = wi.handler_agent AND ad.is_active
                                AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL)
                 THEN 'handler named but NOT REGISTERED'
            ELSE 'has a live handler' END AS routing,
       count(*), count(DISTINCT item_type)
  FROM site_work_items wi
 WHERE status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled')
 GROUP BY 1;
```

| routing | items | item types |
|---|---|---|
| **no handler named** | **298** | 13 |
| has a live handler | 238 | 25 |
| **handler named but NOT REGISTERED** (`bugs_closed/077`'s exact shape) | **9** | 5 |

The 9 all name **`human-review`**, which is not a registered agent. Same mechanism as
077's `forced-text-color-fixer`: filed, unclaimable, invisible.

### The 298, by type — and NOT all of them are defects

| item_type | items | oldest (days) | sites |
|---|---|---|---|
| `cta_names_unknown_destination` | 70 | 13 | 6 |
| `unresolved_cta` | 70 | 35 | 7 |
| `required_fields_missing` | 45 | 13 | 5 |
| **`needs_section_data`** | **44** | **135** | **10** |
| `voice_tells` | 25 | 10 | 1 |
| `image_source_unsatisfiable` | 17 | 10 | 1 |
| `owned_page_review` | 6 | 10 | 4 |
| `dead_control` | 6 | 10 | 2 |
| `capability_gap` | 5 | 4 | 4 |
| `image_url_404` | 5 | 1 | 1 |
| others (3 types) | 5 | ≤12 | — |

> **The honest caveat, and it matters for how this file is read: `capability_gap` is
> SUPPOSED to have no handler.** It is the platform's deliberate "found work nobody can
> action", read as a roadmap. So "no handler" is not automatically a defect and the
> headline number must not be quoted as 298 broken things. Subtracting it gives **293**
> — but the right subtraction is per type, and deciding which of the other twelve are
> deliberate is exactly the work this bug needs and I have not done it.

`needs_section_data` is the one I would look at first regardless: **44 items, 10 sites,
oldest 135 days.** Whatever it detects, it has been detecting it since March.

### Why this compounds rather than merely accumulating

A detector whose output nobody drains is not neutral — it is **actively misleading**,
because `detected` reads like a live state and behaves like a grave. Three of these rows
described a live commercial site accurately and the site was described as sound in its
own handoff for three days, by me, while they sat there.

### One candidate this file may not have

**A routing audit, run on a cadence, not a new handler for each type.** The query above
is the whole detector: any item type whose rows sit non-terminal with no live handler is
a *routing* defect by construction, independent of what the type means. It generalises
past every instance — 077, 115 and this file are three sightings of one thing, and the
fix that makes the next one visible is cheaper than the fix that drains this batch.

Contributed rather than acted on: the drain itself is this file's owner's call.

### 2026-07-28 — correcting my own contribution above: routing is NOT the bottleneck

> **The measurement I added yesterday framed this as unroutable items needing a route.
> That framing is wrong, and one query kills it.** I was about to recommend surfacing
> the 298 on the admin dashboard, having found that
> `internal/core-manager/admin/confirm_work_item_handler.go:79` only accepts items with
> `status='needs_human_review'` — "one status value away from a human's screen".

There are already **325 items in `needs_human_review`**, the status that screen shows,
and the oldest is from **2026-03-15**. The queue humans can see is fuller and older than
the one they cannot. Routing more items into it changes nothing.

```sql
SELECT status, count(*), count(DISTINCT item_type), min(created_at)::date
  FROM site_work_items WHERE status IN ('needs_human_review','detected') GROUP BY 1;
--  needs_human_review | 325 | 23 | 2026-03-15
--  detected           | 157 | 23 | 2026-07-14
```

**So the defect this file names is real but its cause is one level up: there is no
reader for ANY of the queues.** A router would move rows between three unread lists.

### The unroutable 298, classified by what would actually resolve them

| kind | items | types |
|---|---|---|
| **needs a human ANSWER** | **50** | `needs_section_data`, `owned_page_review`, `incomplete_page_group` |
| advisory / machine-fixable in principle | 186 | `unresolved_cta`, `cta_names_unknown_destination`, `required_fields_missing`, `voice_tells`, `dead_control`, `image_source_unsatisfiable`, `image_url_404`, `contact_form_undeliverable`, `truncated_component` |
| deliberate roadmap row | 5 | `capability_gap` |

**The 50 are not backlog, they are blocked work.** Sample:
*"Section 'pricing' on engagement-model needs: Pricing tier names…"* — 17 of them on
leopardessconsulting.co.uk, waiting since **2026-03-15** for information only the owner
can supply. The platform has been asking a question for four and a half months and the
question was never put to anyone.

### What this changes about the fix

Not a router. **An item type must declare who reads it, and "nobody" must be an
answerable and refusable answer.** A queue with no drain rate is a claim about intent
that the data contradicts — and the standing owner ruling (2026-07-25, *"the queue
should not FILL"*) already says as much for the human one.

Put to the owner 2026-07-28 as an explicit call, since it is about what he wants to
exist rather than about mechanism. Recommendation given: promote the 50, turn the 186
into a periodic report rather than a queue, and stop writing any type whose reader is
nobody. **Decision pending — do not act on this section until it is recorded here.**

## 2026-07-29 — contributed (gauntlet_dead_cta / vonc 6): the acceptance ladder is a consumer nobody has listed, and the pile is now 250

**Arrived here independently**, which is worth recording as corroboration: I was
witnessing `bugs_open/131` B, watched a genuine Tier-4 catch produce an
`improve_tool` item, and followed it to the same disabled promoter this file
already names. Same root cause, reached from a different end — **no new
diagnosis, one new consumer and a fresher number.**

**The pile: 157 (07-27) → 250 (07-29 14:2xZ)**, still 23-ish types. Same query as
the section above. That is roughly **+45/day** over two days, so the growth this
file measured at 61%/day has not slowed just because the branch moved on.

**The new consumer: Tier-2 and Tier-4 tool acceptance — the whole ladder is
write-only.**

```sql
SELECT status, count(*), min(created_at)::date AS oldest
FROM site_work_items WHERE item_type='improve_tool' GROUP BY 1;
--  detected    | 7 | 2026-07-17     <- every one filed since 07-17
--  cancelled   | 5 | 2026-07-10
--  complete    | 2 | 2026-07-25
--  wont_fix    | 2 | 2026-07-25
--  unresolved  | 1 | 2026-04-23
```

`judge_acceptance_results` inserts the fixer's work item with the status
**hardcoded** at `'detected'` (`platform/orchestration/actions/tool_acceptance_actions.go:698-702`,
the literal is in the VALUES list), so it lands in exactly the queue this file is
about. **Every `improve_tool` item created since 2026-07-17 is parked**; the two
`complete` rows both predate that and neither came from the current path.

**Why this consumer matters more than its seven rows.** The acceptance ladder is
the platform's own quality immune system, and a large amount of shipped, verified
machinery terminates in these parked rows:

- `bugs_closed/010`'s convergence guard (escalate after 2 failed fix cycles) can
  never reach cycle 1 — the cycle starts with a dispatched `improve_tool`.
- `bugs_open/126`'s concern — that a repair loop inherits the authority of
  whatever test it is given — is currently theoretical for the same reason.
- A **live instance from today**: `e7ea0125-2a58-4e34-97c3-027f664588e6`, filed
  12:30Z by a true-positive Tier-4 failure on `tool-loot-table-balancer`
  (a `<select>` laid out entirely off-screen on mobile, confirmed by geometry and
  by eye — `bugs_open/131` § "B check-side"). Detected, attributed, evidenced with
  screenshots, and parked within the hour. **The page is still broken.**

**What it adds to the fix question, not a new answer.** This file's standing
recommendation — *an item type must declare who reads it, and "nobody" must be an
answerable and refusable answer* — holds, and I am not acting on it (the owner
call is still pending, per the section above). But `improve_tool` is a clean case
FOR that framing rather than against it: its reader is real, named, built and
tested (`tool-improver`, with a convergence guard and an escalation path). It is
not an orphan type awaiting a decision about whether anyone wants it. **So if the
50/186 split is what gets triaged, `improve_tool` belongs in neither bucket** — it
is a type whose reader exists and whose only defect is that nothing promotes the
row to where the reader looks.

`[UNVERIFIED]` whether promoting these seven by hand is safe today. It would
dispatch `tool-improver` at four live tool pages, and `bugs_open/126` says a fixer
can be aimed at the wrong thing by a badly-specified criterion. I have not done it,
and I would not without the owner call this file is already waiting on.

---

## Contribution, 2026-07-29 ~17:10Z — the promoter was RUN, and the premise needs narrowing

**The owner instructed one firing of the disabled `improvement-sweep`** (see the
section above; the `[UNVERIFIED]` question about hand-promotion is now moot for
one of the seven — it went through the machinery, not by hand). Fired at
gamesdesign.co.uk via
`docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/scripts/run_improvement_sweep_once.sh`,
orchestration `30692439-43d2-4406-9fe8-9734c3f5689a`, 17:05:43→17:10:01Z.

**The promoter works.** 67 rows moved `detected`→`triaged` in one statement
(identical `triaged_at 17:08:32.778827`), including `e7ea0125` — the parked
Tier-4 catch — which is now `triaged`, `pipeline=build`, `handler_agent=tool-improver`.
Dispatch followed on its own: first claim 17:09:48, `rerender-pages` working the
site by 17:10:08. **So the parking is not a broken mechanism. It is an unrun one**
— which is what this file said, now demonstrated rather than argued.

**But this file's title over-reaches, and the run shows where.** Not every
finding needs the promoter, because **not every agent inserts at `detected`**.
Grouped by author, for the 97 triaged rows on that site afterwards:

| created_by | inserted DIRECTLY as `triaged` | inserted as `detected` (needed the promoter) |
|---|---|---|
| `rerender-pages` | **32** | 0 |
| `design-discovery-agent` | 3 | 37 |
| `completeness-discovery-agent` | 0 | 19 |
| `design-audit` | 0 | 5 |
| `tool-acceptance-agent` | 0 | 1 |

`rerender-pages` created 32 already-triaged follow-up items within a minute of
starting work, and `design-discovery-agent` does BOTH depending on the path.
**That is why nobody noticed.** The build pipeline never looked starved, because
a large share of the fleet's work items skip the promoter entirely — so the
symptom is not "the queue is empty", it is "one class of finding is missing from
a queue that is visibly busy". Anyone measuring pipeline health to decide whether
this bug is real will measure it healthy.

**Consequence for the fix question.** The standing recommendation (an item type
must declare its reader) is unchanged, but there is a second, cheaper question
underneath it: **why does an item's initial status depend on which agent created
it?** Two writers of the same table disagree about whether `detected` or
`triaged` is the correct initial state, and nothing reconciles them. Fixing that
would retire this bug for the affected types without needing the promoter to run
at all.

**Corrected in the same session, before it reached a conclusion here:** I first
wrote that the downstream build pipeline was "idle for want of promotion" and had
"woken up" on this run. False. Grouping `build-pipeline-trigger`'s own
orchestrations by hour shows 6–17 dispatches per hour all day for other sites;
the two `complete_idle` rows I saw were a lull. Parked items do not block the
pipeline — they are invisible to it. Logged in `WRONG_CALLS.md`.

**A separate defect found by the same run is filed as `bugs_open/150`**: the loop
terminated at `complete_clean` ("No issues found — site is clean") immediately
after promoting those 67 findings, and skipped its own dispatch branch. Relevant
here only as a warning — anyone who runs the sweep to clear this backlog will get
a success message asserting the opposite of what happened.

---

## Contribution, 2026-07-30 — the first complete cycle in this bug's history, and what it exposed

The 07-29 contribution above showed the promoter works. **The cycle has now
finished, and it worked**: `e7ea0125` → `tool-improver` (46s) → component
rewritten → page redeployed → **re-verified clean on the served page with a
positive control in the same batch** (`bugs_open/131` § "B check-side … FIXED AND
VERIFIED"). Queue fully drained overnight: 139 complete, 0 triaged.

**This is the first `improve_tool` item ever to complete.** The 07-29 measurement
in this file — 7 of 7 parked since 07-17, none dispatched — means every downstream
mechanism built for this item type had never once run. So the value of unparking
was not the one page; it was that **four separate mechanisms got their first
execution**, and half of them failed:

| what ran for the first time | outcome |
|---|---|
| `tool-improver` on an acceptance-raised item | **worked** (46s, verified fix on the live page) |
| `tool-improver` on a `tool-auditor`-raised item | **failed at step 1**, twice ⇒ `bugs_open/154` |
| `bugs_closed/010`'s convergence guard | **still unexercised** — nothing reached cycle 1 |
| `bugs_open/126`'s fixer-authority concern | **still theoretical** for the same reason |

**The failure is worth this file's attention because it is the same class of
defect.** 083 says a finding never reaches its handler; 154 says that for one of
the two creators, reaching the handler is not enough — `input_data.component_id`
resolves to nil at `load_tool`, and (the counterintuitive part) **it is the items
whose `component_id` column is SET that fail.** A row can satisfy every visible
precondition and still not carry what the workflow reads.

**What that means for this file's standing recommendation.** "An item type must
declare who reads it, and 'nobody' must be an answerable and refusable answer"
survives, but it is now demonstrably not sufficient: `improve_tool` HAD a
declared, built, tested reader, and 50% of the population still could not be
processed by it. **A declared reader is not a working path.** The check that
would have caught this at any point in the last two weeks is one dispatch — which
is exactly what nothing was doing.

Also recorded, so nobody re-derives it: the other 9 failures on that site were
8 × `needs_content_image` (all one S3 download error) and 1 × `audit_tool`
("Claim timed out (attempts exhausted)"). Neither is about promotion.

---

## Consumer notice, 2026-08-09 (bugfix_230_discovery_driver) — the UPSTREAM half is getting a clock; this file's drain question is untouched and will get busier

`bugs_open/230` (no recurring driver for the three discovery agents) is being fixed:
migration `346` seeds a fair-rotation driver (`site_discovery_rotation` stamp table +
three `site-discovery-rotation-*` scheduled tasks, one site per agent per hourly tick,
7-day per-site period, observe-only), council corr `2281fc48`. **What changes about this
file's world:** detection stops being attention-driven, so the `detected` pile this file
is about will start accumulating *fresh, true* findings from unattended sites on a steady
~9 runs/day — the honest version of the state this file documents, rather than silence.
Nothing here promotes, drains, or adds a `triage_detected_items` carrier: the standing
owner decision recorded above ("Decision pending — do not act…") remains the blocker it
was, now with better data accruing for it.

Two answers this file recorded as open, found in the concept register during that work:
- The `[UNVERIFIED]` why of the 2026-05-02 disable: **on record as deliberate** —
  IMP-016, *"intentionally paused during core build"*, with a gated re-enable sequencing
  (check handlers exist, observe-only first, watch one clean cycle).
- Fix candidate 1's warning ("read that reason before flipping it") gains a second,
  measured reason: the sweep's live pre_query caps at <50 open build items and its
  selection starves (IMP-010) — today that cap excludes webdesign.co.uk (85) and
  dartsonline.com (79). **Re-enabling improvement-sweep as-is would examine everything
  except the sites most worked on.** Landmine filed.

---

# OWNER RULING 2026-08-15 — candidate 2 chosen and BUILT: the promoter is its own scheduled task (seed 430, SCH-026)

Put to the owner by the `bugfix_277_required_fields_repair` lane, whose council trail
(corr `7b0e2833`) drew a high-severity objection to a producer born `triaged` — the
workaround this file predicted every producer would reach for. The owner ruled: **do 083,
as candidate 2** — *"give triage its own scheduled task, decoupled from the fix loop:
promote `detected → triaged` on a slow cadence for item types whose handlers are
known-good, leaving discovery itself manual."* Not candidate 1: re-enabling
`improvement-sweep` wholesale re-arms the whole audit+fix loop and its own pre_query
excludes the busiest sites (the last section above).

**Built the same day, live within the hour.** `detected-item-promoter`
(`sql_for_agents/430_detected_item_promoter_task.sql`, ledger-recorded, register
**SCH-026**): a pure-SQL scheduled task on `feasibility-recheck`'s exact pattern
(`fire_message=false`, pre_query = worker, verified in `cmd/scheduler/main.go:269-277`),
900s cadence, **≤20 promotions per tick oldest-first**, and a **known-good rule that is
data, not opinion**: the handler is a live active agent definition AND the exact
`(item_type, handler_agent)` pair has ≥1 lifetime `complete` (this file's own candidate 4:
"check the handler is real before celebrating the detector"). A pair with zero completes is
HELD until a human promotes one row by hand — every new type's first dispatch is a
deliberate canary. It mirrors `TriageDetectedItemsAction`'s UPDATE in effect incl.
`pipeline='build'` — which turned out to be load-bearing for a reason this file did not
record: the **scheduler gate** (`build-pipeline-trigger` pre_query) only wakes the loop for
a site holding a `pipeline='build'` triaged item; the loop's own loader does not filter by
pipeline.

Pile at apply (verify block's partition assertion, disconfirmable): **70 detected = 66
promotable + 4 held (`page_component_status_drift → component-template-fixer`, 0 lifetime
dispatches) + 0 unroutable.** Re-measured before building: every one of the 18 pairs in the
pile has an active handler; 17 have completes (`content_rewrite` 107, `needs_rerender` 111,
`undeployed_asset` 260 …).

**Consequence:** the born-`triaged` workaround is no longer needed. `bugs_open/277`'s
producer went back to `Status: "detected"` in the same hour (inert until the next chassis
roll). Other born-triaged producers (migration 233's rerender items, `check_integrity`,
`check_tool_acceptance_due`) are their lanes' call.

**Status: FIX LIVE for the triage half; OPEN until this file's own verify criteria are met** —
(1) the pile drains and *stays* drained; (2) `phantom_internal_link` reaches a non-zero
`complete` count for the first time (none in the pile today — waits for the next re-raise);
(3) completions verified at the live page. `improvement-sweep` itself stays paused (IMP-016)
— discovery has SCH-025, triage has SCH-026; the auto-fix half remains a separate decision.

## VERIFY CRITERION 1 MET — 2026-08-16 morning (bugfix_277 session, post-roll)

`detected-item-promoter` ran overnight on its 15-minute cadence. **Pile: 70 → 4, and the 4
are exactly the held pair** (`page_component_status_drift → component-template-fixer`, zero
lifetime completes — awaiting a hand canary). 100 rows promoted since apply (the pile refilled
from discovery rotation and drained again): **93 complete, 4 failed (all downstream
page-build-handler failures — save_sections/verification, not promoter defects), 3 parked**.
Zero rows born after apply remain at `detected`. Criterion 1 ("drains and STAYS drained") holds.
Criterion 2 (`phantom_internal_link` first-ever complete) still waits for a re-raise; criterion 3
(verify completions at the live page) is not yet done — a sample is the next session's job.

**The producer revert is LIVE**: chassis `v1.0.1303` (uniform, 9 pods), stamp `5e075a6f…`
carries `3c6354059`; `check_required_fields_missing.go` files born-`detected` again.

**Council on the promoter (corr `05a3d1c8`): REVISE**, every objection measured the same morning:
- "pipeline='build' unconditional could misroute diagnose/report items" — promoted rows' original
  pipelines: build 97 / content 2 / design 1; **no diagnose or report item has ever sat at
  `detected`** (lifetime). Same rewrite the original promoter always did. A cheap door-closer for
  the next session: `AND wi.pipeline NOT IN ('diagnose','report')` in the candidates CTE — as a
  NEW numbered migration (430 is ledger-recorded; never edit a recorded file).
- "stale reaper keyed on created_at would reap promoted rows" — every enabled reaper pre_query keys
  on `claimed_at` (`claimed-item-timeout`, `stale-work-item-reaper`, `report-dispatch`); none on
  `created_at`. Does not apply.
- "sibling born-triaged producers left unaudited" — grep: exactly two, `check_integrity.go` and
  `check_tool_acceptance_due.go` (the precedents cited when 277 went born-triaged). Their lanes'
  call whether to return to `detected`; both pairs are known-good so the promoter would carry them.
- "sole live carrier premise" — re-verified live twice: only `improvement-loop` carries
  `triage_detected_items` (migration 286 removed the other two); `improvement-sweep` enabled=f.
- reuse_agent's "invoke the Go action instead of mirroring in SQL" — the honest answer is that
  the Go action is site-scoped and workflow-embedded (needs an orchestration per site per tick);
  the SQL mirror is the estate's own SCH-006 pattern for exactly this shape. Worth stating in
  round 2, not conceding.

## 2026-08-17 — criterion 3 MET; criteria 1 and 2 CORRECTED; and the promoter's own risk 2 fired

Measured by the `bugfix_277_required_fields_repair` lane against chassis `v1.0.1305`
(OCI `revision=6a782274b`, positive+negative binary controls both behaved). Everything
below is a fresh query, not a figure carried forward — two of the three corrections exist
precisely because the previous block carried figures forward.

### VERIFY CRITERION 3 — MET (4 of 4 verified at the served page)

Sampled four promoted-and-completed rows on `mortgagecalculator.co.uk`, all promoted by
**this** promoter (they carry `spec.original_pipeline` with `triaged_at` after apply):

| item_type | page | page `updated_at` | item `completed_at` | live |
|---|---|---|---|---|
| `tone_shift` | `/guides/missed-payments/index.html` | 10:37:33 | 10:37:38 | 200, 2,009 visible chars, h1 *"A missed payment doesn't rule you out"* |
| `content_rewrite` | `/investor/index.html` | 10:33:12 | 10:33:54 | 200, 2,641 chars, h1 *"Buy-to-Let Investor Dashboard"* |
| `needs_content_page` | `/contact/index.html` | 10:27:45 | 10:27:57 | 200, 1,426 chars, h1 *"Something not quite add up…"* |
| `needs_content_page` | `/guides/index.html` | 10:24:04 | 10:24:14 | 200, 1,981 chars, h1 *"Guides to help you decide"* |

Two independent instruments agree, and each could have come out otherwise. (a) The page was
written **5–42 seconds BEFORE** the item was closed, in all four cases — the causal order a
real handler produces; a row marked `complete` without work would leave `page.updated_at` at
some unrelated earlier time. (b) The pages serve real, on-topic prose. The two
`needs_content_page` rows are the strongest: those pages exist **at all** only because the
handler built them.

> ⚠ **Method trap, recorded because it nearly produced a false finding.** The first fetch of
> all four URLs returned **404**, which reads exactly like "the work never reached the page".
> It was my own artefact: the rows store `/guides/index.html` and this site does not serve the
> trailing-slash form. The control that caught it was fetching the site ROOT (200, 37 KB) —
> proving the instrument worked and the domain was live — after which re-fetching the **exact
> stored URL** returned 200 on all four. Fetch the URL the row actually names; and when a
> whole sample 404s, suspect the request before the subject.

### CRITERION 2 — met, but NOT by this fix, and it is not discriminating evidence

Criterion 2 asks that `phantom_internal_link` reach a non-zero `complete` count *for the first
time*. It has **7** — and every one of them completed **2026-08-09 to 2026-08-11**, before this
promoter was built on 2026-08-15. All 7 carry `original_pipeline`, so they were promoted by the
*original* `TriageDetectedItemsAction` path (the hand-run recorded in this file's 2026-07-29
section), not by candidate 2.

> **CORRECTED 2026-08-17:** the 2026-08-15 block stated *"criterion 2 … waits for the next
> re-raise"* and the 2026-08-16 block repeated it. Both were false when written — the count had
> been non-zero for six days. The criterion was authored 2026-07-26 when it was genuinely zero,
> and carried forward twice without re-measuring. **What caught it:** querying the completion
> DATES rather than the count. The cheap check that would have: `SELECT status, count(*),
> max(completed_at) FROM site_work_items WHERE item_type='phantom_internal_link' GROUP BY 1`.
> Logged in WRONG_CALLS. **Consequence for closing this bug:** criterion 2 is satisfied on its
> literal wording but proves nothing about candidate 2 — do not cite it as evidence the new
> promoter works. Criteria 1 and 3 carry that weight.

### CRITERION 1 — the meter is now wrong, and would read "failing" for ever

`SELECT count(*) FROM site_work_items WHERE status='detected'` is **82** today, up from the 4
recorded yesterday. That is not a regression. Partitioned:

```
flag-only, NO handler_agent (promoter can never touch these)   77   (image_url_404 41, head_essentials_missing 36)
has a handler                                                   5   (page_component_status_drift 4, placeholder_contact 1)
of those, pairs that pass the known-good rule                   0
```

The 77 are findings whose producers **deliberately file no handler**, and `detected` is where
they belong permanently — 40 of them were restored to that state on purpose by the concurrent
`bugs_closed/284` lane's migration `442`, and its migration `443` now enforces the invariant
with `CHECK swi_no_handlerless_promotable`. The remaining 5 are the promoter correctly holding
two never-completed pairs for a hand canary. **So the promoter has zero promotable work and is
behaving exactly as designed, while the criterion-1 meter reads 82 and climbing.**

> **CORRECTED 2026-08-17 — restate criterion 1 as:** *the count of `detected` rows that have a
> `handler_agent` whose pair passes the known-good rule falls and stays at or near zero.*
> The raw count conflates two populations with opposite meanings and will now grow with normal
> flag-only discovery for ever. [MEASURED] promotable pile = **0**.
> `443`'s CHECK constraint is also a hard backstop *behind* this promoter's handler predicate:
> a future promoter bug that promoted a handler-less row would abort the tick loudly rather
> than misroute silently.

### The promoter's OWN risk 2 fired — and is now closed by migration `444`

430's submitted risk 2 read: *"the known-good rule trusts ONE lifetime complete per pair; a pair
whose only complete was a false success qualifies."* It fired. Of **85** rows promoted since
apply, **19 ended `failed`**. The counterfactual — each pair's record **as it stood at the
instant of its promotion**:

| pair | complete / failed at promotion | promoted | ended failed |
|---|---|---|---|
| `literal_markdown → page-build-handler` | 1 / 28 = **3%** | 6 | **5** |
| `audit_tool → tool-auditor` | 18 / 21 = 46% | 12 | 8 |
| `phantom_internal_link → page-build-handler` | 7 / 8 = 47% | 2 | 2 |
| every other pair | ≥ 60% | — | 4 |

`literal_markdown` cleared the gate on **one** lifetime success against 28 failures, and the
promoter then fed it six more. **Migration `444`** (applied and ledger-recorded 2026-08-17,
with a separate `_ROLLBACK.sql`) adds two predicates to the candidates CTE — `430` is
ledger-recorded and was not edited:

1. **pipeline allow-list** `wi.pipeline IN ('build','content','design')` — the council's
   pipeline objection. An allow-list, not the deny-list the objection implied, because
   [MEASURED] `report` does not exist on this table at all (0 rows) while `experience` (7) and
   `maintenance` (1) do: a `NOT IN ('diagnose','report')` deny-list names one value that cannot
   fire and misses two that can. Lifetime promotions across both implementations (n=1562) are
   build 1055 / design 310 / content 197 — zero diagnose, zero report.
2. **success floor** — a pair with ≥5 terminal outcomes must still be succeeding at ≥25%.
   The threshold is set by the census, not chosen: the 28 pairs holding ≥1 complete run
   3%, then 41, 42, 46, 50, 67, 79, … so any floor in 10–35% isolates the single pathological
   pair and touches nothing else. The <5-terminal exemption keeps the canary path intact.

Both doors hold **0 rows today** — they are doors, not repairs. The migration's verify block
carries a **positive control** (`literal_markdown` must fail the floor, else the predicate is
inert and the "holds 0" assert is vacuous) and a **negative control** (`page_rerender`,
4017/21, must pass); both came out the required opposite ways at apply.

**The 41–47% pairs are deliberately NOT addressed here.** "Is `page-build-handler` defective for
`literal_markdown` and `phantom_internal_link`?" is a different bug from "should the promoter
keep feeding a pair that has stopped working", and bundling them is the shape the guardian seat
has objected to repeatedly. Named so the silence reads as a decision.

**Status: unchanged — FIX LIVE, bug OPEN.** Criterion 1 holds under its corrected wording,
criterion 3 is met, criterion 2 is met but non-discriminating. Remaining: let `444`'s doors sit
long enough to show they hold nothing they should not.

### COUNCIL APPROVED — round 2, corr `05a3d1c8`, 2026-08-17 11:27Z

**APPROVED**, 12 seats approving (including `architecture` and `prior_art_librarian`, the seat
whose HIGH objection gated round 1), 3 abstained, not truncated. Two advisory objections, none
high. Both were checkable, so both were checked rather than banked — and one of them changed
what I believe about the door-closer.

**1. `guardian` (LOW) — "the per-tick cost of the new correlated subquery isn't measured, only
the 0-and-0 row-count outcome."** Fair: I had measured the effect and not the cost. `EXPLAIN
(ANALYZE, BUFFERS)` on the live table, candidates query only:

| | execution time |
|---|---|
| 430's predicates (before) | **65.3 ms** |
| 444's predicates (after) | **78.1 ms** |
| delta | **+12.9 ms** |

On a 900,000 ms tick that is **+0.0014%** of the interval. No starvation risk.

**2. `guardian` (MEDIUM) — "confirm the fleet-wide pipeline distribution; I can't read
`scheduled_tasks` from here."** Confirmed, and it is `grounded_in [3]` re-stated so the seat's
check is answerable from the file: pipelines across all 8,421 rows are build 7749, content 490,
design 130, **diagnose 44, experience 7, maintenance 1**. The allow-list covers the first three;
the last three are held by design. The seat's real point is the *silent* failure mode — a new
pipeline value stops being promotable with no signal beyond a growing pile that looks like
normal backlog. **That is a live residual and it is sharper than my risk 4 admitted**, because
this file's criterion-1 meter is *already* known to conflate populations. Not closed here.
The cheap control, when someone takes it: have the pre_query return a third column counting
`detected` rows a door held, so the scheduler's own log makes it visible.

**3. `bug_historian` (MEDIUM) — "the floor treats `complete` as ground truth, and
`bugs_closed/028` is the documented case of `page-build-handler` reporting complete while
deploying hollow content. If that pair's ONE lifetime complete is hollow, the floor is computed
on a corrupted signal."** The sharpest form of this is one row, so I checked it.

The complete: `literal_markdown` on `gaswholesalers.com/how-pricing-works.html`, 2026-08-15
16:43. Fetched the live page and ran five literal-markdown patterns over its visible text
(`**bold**`, ATX headings, `[text](url)`, leading bullets, `_italic_`): **0 hits in 8,120
visible characters.**

**A zero from a detector I had just written proves nothing, so it was given a demand control** —
the same five patterns, same code path, against three pages whose `literal_markdown` items are
`failed` or `needs_human_review`:

| page | item status | markdown hits |
|---|---|---|
| `gaswholesalers.com/how-pricing-works.html` | **complete** | **0** |
| `ai-agent-orchestration.com/news.html` | needs_human_review | 9 |
| `robot-hands.com/gripper-catalog/index.html` | failed | 5 |
| `fundamentallyai.com/news/index.html` | failed | 13 |

**The objection is answered, and the answer strengthens the floor rather than weakening it.**
That one `complete` is real work, not a `028`-shaped hollow completion — and the failures are
real too, with the defect still visible on the served pages. So `1 complete / 28 failed` is an
honest measurement of a handler that mostly cannot do this job, and holding the pair is right.
It also means the floor's input signal was not corrupted **in this case**; the general concern
stands for pairs nobody has spot-checked, and is recorded as risk 5.

> The failures are already owned: `bugs_open/184` (LLM markdown reaches the page as literal
> asterisks) and `bugs_open/201` (its dispatch path). Not re-filed here — contributed there.

**Trailers:** the two commits carrying this work went out with `Council-Submitted: 05a3d1c8-…`
before the verdict landed, which `098` resolves to the approval at report time. No amend
(forward-only), and no `Council-Reviewed:` written on a verdict that had not been read.

---

# 2026-08-17 evening — the first canary RAN, the whole arc is demonstrated, and the residual's stated remedy turned out not to work

Session `bugfix-083`, picking up the close-out this file's own handoff invited. Everything below
is a fresh query or a fetch of a live page; nothing is carried forward.

## 1. The held-pair canary — run, completed, and verified at the artefact

The owner ruled (2026-08-17) that a human should canary a held pair now, and that a time limit
should exist for the future. The time-limit half was built the same afternoon by the
`bugfix_277` lane (migration `453`, `held-pair-canary-escalation`, which had already fired and
moved 4 rows out of `detected`). **This is the other half: a human actually being the human.**

`page_component_status_drift → component-template-fixer` was the oldest held pair — **20 rows,
`claimed_at IS NULL` on every one, never dispatched in its life.**

**Validity-checking the four escalated rows first is what made this worth doing carefully, and it
found a trap.** `453`'s escalation payload says *"promote ONE row of this pair"*. One of the four
was stale: `bc041cfb` named `page_components.0f02ca76…`, which **no longer exists** — the page was
re-rendered on 08-15 and the slot now holds `a9550607…` at `build_status='deployed'`. The finding
was true when filed and had since been repaired by ordinary re-render. Dispatching *that* row would
have failed inside the handler (`fixPageComponentStatus` returns a hard error on `sql.ErrNoRows`),
and — post-`444` — **that failure would have counted against the pair's success ratio.** A canary
failing for an artefact reason teaches the gate the exact opposite of the truth. Filed as
`bugs_open/300`; closed by hand as `complete` / `resolution_path='manual:revalidated'` with the
evidence in `result.revalidation`.

**Canaried `d15536c2` instead** (drift verified still true at promotion time), promoted from
`needs_human_review` — note `430`'s prescribed statement ends `AND status='detected'` and would
have matched **0 rows**, silently, since `453` had already moved it.

| | |
|---|---|
| promoted by hand | 21:47:33Z |
| claimed | 21:48:37Z (**64s** later — dispatch works) |
| component `06c80367` `approved → deployed` | 21:49:05Z |
| item `complete` | 21:49:16Z |

The artefact moved **11 seconds before** the row was closed — the causal order real work leaves.
Evidence was taken at `page_components.build_status` and the served page, **not** at the item's
`result`, because `bugs_open/287` is still open for `build-dispatch-loop`.

**Demand control, taken at that moment:** the two un-canaried siblings (`c881b7df` tool-1,
`424986fd` prose-2) were still `approved`, with their original `updated_at` of 08-06 and 08-15.
So the instrument could tell repaired from not.

## 2. Then the mechanism did the rest by itself — which is the discriminating evidence criterion 2 never was

With one lifetime completion the pair now passes the known-good rule, so the two remaining
still-true rows were returned to `detected` (they were stranded — see §3) and **the promoter took
them on its own next tick, unattended**:

| | |
|---|---|
| promoter tick | 21:52:46Z — both rows promoted at 21:52:45, stamped `spec.original_pipeline='build'` |
| `c881b7df` | claimed 21:53:03 → component `a39ffa7b` `approved→deployed` 21:53:21 → complete 21:53:28 |
| `424986fd` | claimed 21:53:28 → component `c8c34743` `approved→deployed` 21:53:43 → complete 21:53:50 |

Pair lifetime record now **4 complete, 0 failed**, from a standing start of zero dispatches ever.
(⚠ subject to the 18:30Z OPEN RISK above: "lifetime" is only "whatever survives in the table". If
those four completes were to leave it, this pair reverts to held and the canary would have to be
run again — which is an argument for the durable per-pair tally that section proposes.)

**The served page is unharmed, checked before and after:** `/loans/consolidation.html` returned
200 / 13,512 bytes / 2,190 visible characters both times, **byte-identical visible text**, h1
*"Debt Consolidation Risk Checker"*, and the canaried slot's own prose still present (probe: *"If
you're thinking about rolling several debts into one loan"* — with a nonsense negative control that
correctly returned absent). All three components' `rendered_html` lengths are unchanged (469 / 744
/ 5720): this fix repairs a status, and it did not rewrite content.

**This is the arc the bug was filed about, end to end, for the first time:** a human canaries one
row → the pair becomes known-good → the promoter dispatches the rest unattended → the handler
repairs them → verified at the served page. Unlike criterion 2, it is discriminating: none of it
could have happened before 2026-08-15.

## 3. A composition gap between `453` and SCH-026 — the escalation is a ONE-WAY door

`453` moves a held row `detected → needs_human_review`. The promoter only ever selects
`status='detected'`. **Nothing moves an escalated row back.** So when a pair is canaried and becomes
known-good, its other findings do *not* rejoin the queue — they sit in the human pile for ever,
which is `bugs_open/033`, i.e. the disease this file exists to cure, one step further along.

Handled by hand this time (the two rows above were returned to `detected` with a
`result.returned_to_detected` stamp, and the promoter then took them within 4 minutes). **Not
fixed in the mechanism** — flagged for the `453` author as the natural follow-on: a successful
canary should return that pair's escalated rows to `detected`.

## 4. The guardian's residual: its stated remedy does not work, and the real gap is one level down

This file records the residual with a suggested cheap control — *"have the pre_query return a third
column counting `detected` rows a door held, so the scheduler's own log makes it visible."*

**[MEASURED at the running service] That would have been write-only.** Every tick of this task
emits one line and it carries no numbers at all:

```
{"caller":"scheduler/main.go:274","msg":"Pre-query task completed (no message fired)",
 "task":"detected-item-promoter"}
```

Six such lines in `--tail=3000`. For a `fire_message=false` task the pre_query result is merged into
`inputData` and then **discarded unlogged** — so `promoted` and `pairs` have never been readable
either. That is SCH-006's observability gap and **eight CTE-only tasks share it**, so it was fixed
there rather than here (owner-approved this session: *fix the framework*).

Shipped as `03012d862`, council `8dc58e2a-2a0f-4b0c-aa9d-738177db079b` (`FORCE=1` — `097` scopes to
`platform/`/`internal/`/`pkg/`, and this change is `cmd/` plus config, the blind spot this lane
already noted):

- **`cmd/scheduler/main.go`** — every CTE-only scheduled task now logs its own pre_query result.
  Capped at 2 KB, and **the cap announces itself**: a silent truncation would rebuild, inside the
  fix, the defect the fix exists to close. Test pins the marker, not the logging.
  **INERT until `kafka-scheduler` is rebuilt and rolled — verify at the artefact, not at the commit.**
- **Migration `458`** (applied + ledger-recorded, `_ROLLBACK.sql` alongside) — supplies the numbers,
  with a reason per held pair. The doors move into a `scored` CTE evaluated **once** per row, of
  which `candidates` and `held` are complementary halves; deliberately **not** a second negated copy
  of the predicates, which would have drifted within four hours (`444` added them, `454` corrected
  them, same day). Guarded on a verbatim pre-image md5 so a concurrent edit by the still-open 277
  lane **aborts** it rather than silently reverting that lane's work.

**`458` also closes a second defect found while writing it.** The final SELECT's
`WHERE (SELECT COUNT(*) FROM promoted) > 0` suppresses **nothing** — aggregate-only target list, no
`GROUP BY`, so one row returns regardless (verified read-only: the `WHERE` form returns `('0',NULL)`,
the `HAVING` form returns 0 rows). Consequences, both live until today: the scheduler could not
distinguish an acting tick from an idle one, and the task claimed its `maintenance` concurrency slot
on **every** tick, defeating `bugs_open/048`'s release-on-no-op. Now suppressed on
`promoted=0 AND held=0` — *not* on `promoted` alone, because an idle tick that is holding rows is
exactly the state this residual is about.
**Honest sizing: a door, not a repair.** No `maintenance` group-mate is starved today (all five
within 1.04 intervals of due).

**What it reports on its first day is the argument for it** — and it is a positive control, not a
zero:

```
promoted=0  held=2
literal_markdown->page-build-handler (pair below the 25% success floor);
placeholder_contact->page-build-handler (pair has never completed one (awaiting a hand canary))
```

`444` recorded both doors as holding **zero** rows at apply and called them *"doors, not repairs"*.
**One of them is now load-bearing and nothing could see it.**

> ⚠ **This counter inherits the OPEN RISK recorded above at 18:30Z** (*"work-item history SHRANK
> — 14 completes left the table, cause UNDIAGNOSED"*), and I am flagging it rather than leaving the
> two accounts to drift. `held` is a live `COUNT` over `site_work_items`, exactly like the known-good
> rule and the floor it reports on. **If completed rows can leave the table, a pair can be reported
> as held for a reason that is an artefact of row disappearance**, and the held *reason* string will
> say "pair has never completed one" — which would be a confident, wrong explanation. That makes the
> visibility worth *more*, not less (the shrinkage was invisible precisely because nothing reported
> per-pair state), but it is not independent evidence: **if `090` finds the actor and the answer is a
> durable per-pair tally, this counter should read that tally, not the table.** Whoever takes that
> item owns this line too.

## 5. Corrections to this file's own record

> **CORRECTED 2026-08-17 (evening) — concept register SCH-007 is stale and this file leaned on it.**
> SCH-007 states a pre_query *"must return at least one row, or `last_triggered_at`/
> `last_completed_at` never advance"*. **False since `dc2e4b61a`** (the `bugs_open/048` fix): the
> zero-row path stamps both timestamps and logs *"Pre-query found no rows — task ran with nothing to
> do"* (`cmd/scheduler/main.go:200-216`, read at source). It is the rule that made the
> non-suppressing `WHERE` look deliberate. Register entry corrected.

> **A generic "is this finding still true?" predicate was considered and REFUSED by its own
> measurement.** Sizing the class on `spec.page_component_id` gives one item type and one stale row
> of twenty. A generic proxy — *"was the page redeployed after the finding was filed?"* — came out
> **backwards**: promoted rows whose page was redeployed after filing succeeded 530/618 (86%) against
> 24/59 (41%) for those not redeployed. That reads as staleness *helping*, which it does not:
> `pages.deployed_at` is updated **by the handler's own work**, so the column is a consequence of the
> outcome it was meant to predict. The instrument cannot separate the two. Logged in `WRONG_CALLS.md`;
> no predicate built.

> **A HIGH hazard raised against the canary was REFUTED at the live config.** The warning was that
> `component-template-fixer`'s `create_rerender` step INSERTs into a `domain` column dropped from
> `site_work_items`, which would fail *after* the repair had been applied. The **live** agent
> definition names `pipeline`, not `domain`; the repo artefact saying otherwise
> (`k8s/bk_agent_definitions_backup.sql`) is a snapshot dated 2026-03-12. What caught it: reading the
> live row instead of the seed. Recorded because the hazard was plausible, specific, and wrong.

## 6. Where the three criteria stand now

| criterion | state |
|---|---|
| 1 — the promotable pile drains and stays drained (corrected wording) | **HOLDS.** Promotable = **0**; the raw `detected` count is 152, of which 147 are flag-only rows with no handler and 5 have one. `458`'s verify block asserted the partition. |
| 2 — `phantom_internal_link` reaches a non-zero `complete` | met on its literal wording, **non-discriminating** (already true six days before the fix) — unchanged, do not bank it |
| 3 — completions verified at the live page | **MET twice over**: the 08-17 morning sample of four, and now three more this evening on `loanandmortgagecalculator.co.uk`, before-and-after with a negative control |

**Status: FIX LIVE, bug still OPEN**, for two reasons and no others:

1. `cmd/scheduler/main.go` is **committed but not rolled** — the held count is computed and stored
   in the pre_query's output today and nothing reads it until `kafka-scheduler` ships. The residual
   is not closed until a tick's held count is visible in the service's own log.
2. `444`'s doors and `458`'s counter should sit a week before anyone declares they hold nothing they
   should not.

Everything else this file asked for is done. When those two land, close it — **naming both paths on
the commit** (`git mv` landmine) and verifying at HEAD with `git ls-tree`.

## 7. COUNCIL APPROVED round 1 (corr `8dc58e2a`, 22:07:31Z) — and the two objections worth checking were checked

**APPROVED**, 4 advisory objections, none high-severity, not truncated; `architecture` returned
`point_fix` (strictly additive observability, no new contract), `constitution`, `mission`,
`guidelines`, `reuse_agent` and `prior_art_librarian` approving. Three objections were checkable, so
they were checked rather than banked — and **one of them was simply right**.

**1. `guardian` (MEDIUM) — "risk 1 asserts the other CTE-only tasks 'select counts, ids and reasons
only', but that is not verified against the actual pre_query bodies; logging every task's raw result
is a real blast-radius change (log volume, potential secret exposure)."** The right objection, and I
had asserted where I should have measured. Now measured, over all **9** enabled `fire_message=false`
tasks:

| | |
|---|---|
| mention `password`/`secret`/`credential`/`api_key`/`private_key`/`b2_`/`aws_`/`smtp`/`authorization` | **0** |
| mention `email` | **0** |
| mention `token` | **2** — `council-seat-token-pressure`, `fleet-step-token-pressure` |

Both "token" tasks are **LLM token-pressure monitors** (output vs max tokens), not credentials, and
both end `SELECT id::text AS note_id FROM ins` — a `doc_notes` UUID. So the fleet-wide logging
change exposes nothing sensitive **today**; the standing hazard (a future author selecting a secret)
remains, which is why the code comment at the log site says to select counts and ids, never secrets.

**2. `guardian` (LOW) — "does the same non-suppressing `WHERE` exist in the other CTE-only tasks? If
so, `048`'s release-on-no-op is silently defeated elsewhere too."** Checked: **exactly one other
does** — `held-pair-canary-escalation` (migration `453`), which this file already named as carrying
the identical idiom. The other seven do not. So the defect is bounded at 2 of 9, both in this lane,
and `453`'s correction is its author's call rather than something to bundle here.

**3. `editquality` (missing) + `tooling_provenance` (MEDIUM) — "the plan claims it corrects
concept-register SCH-007, but no edit delivers that; a claimed correction not written back to the
register is indistinguishable from the stale-half problem the register itself has."** **Correct, and
it is a defect in the submission rather than in the work.** SCH-007 *was* corrected (strike-through,
date, and the cost named — the stale wording is what made the broken suppression idiom look
mandatory), along with SCH-026 — but the edit list did not include the register file, so the seats
could not see it. Recorded as a lesson rather than argued away: **an edit list that omits work you
have actually done reads exactly like a claim you cannot support.**

**4. `debug_historian` (MEDIUM) — "did the 'held set measured live' evidence come from running the
LIVE pre_query by hand? It embeds `UPDATE … RETURNING`, so gathering evidence could itself have
promoted work items."** Exactly the right question to ask, and the answer is no: the held census was
run as a **read-only mirror** (the `scored` CTE alone, no `promoted` CTE), and when the full query
*was* exercised it was inside `BEGIN … ROLLBACK`. No row was promoted as a side effect of measuring.

**Two advisories accepted and NOT fixed, named so the silence is a decision:**

- `debug_historian` (LOW): the forward migration `RAISE`s on any md5 mismatch, **including the
  already-applied case**, where needle-gate discipline would prefer a 0-row no-op. The rollback file
  has that symmetric branch and the forward file does not. Real, minor, and a re-runner would get a
  loud abort rather than silent success — which is the safe direction to be wrong in.
- `reuse_agent` / `constitution` / `prior_art_librarian` (LOW, all three): check for an existing
  shared truncation helper before adding `truncateForLog`. Checked — there are **five** `truncate*`
  helpers in the tree (`truncateCell`, `truncateStr`, `truncateTail`, `truncateNewsSummary`,
  `truncatePreservingRealised`) and **every one is unexported inside `platform/orchestration/actions`**;
  the only exported one, `fetchguard.LimitedRead`, is for HTTP bodies. Nothing `cmd/scheduler` can
  import, so a package-local helper is right. The pre-existing near-duplication inside `actions` is
  real but is not this change's to fix.

The commit carrying this work went out with `Council-Submitted: 8dc58e2a-…` before the verdict
landed, which `098` resolves to the approval at report time — no amend, per forward-only.

---

# 2026-08-18 — THE RESIDUAL IS CLOSED AND VERIFIED AT THE SERVICE, and the first thing it showed was that the held pile had grown 8× overnight while nobody could see it

Session `bugfix-083`, after the fresh fleet build. This is the post-roll verification the section
above said was owed, plus what the new visibility immediately revealed.

## 1. The deploy, proven per SERVICE with a control that behaves

| | |
|---|---|
| pod | `kafka-scheduler-5d9f8c7bf7-wcsdq`, image `docker.io/aqls/kafka-scheduler:v1.0.1309` |
| started | 2026-08-18 15:45:39Z |
| the service's own statement | `main.go:63` `"build provenance"`, `git_commit=f0117fb8b93ea3e1f32298daeb9751bcff4b90c7` |
| carries the fix? | `git merge-base --is-ancestor 03012d862 f0117fb8b` → **YES** |

> ⚠ **My first control was not a control, and I am recording it because it "failed" in a way that
> looked alarming.** I checked that my *last* commit (`58f6ad360`) was NOT in the build — it read as
> an ancestor, which I briefly took as the control failing. It was not: the build post-dates all of
> yesterday's work, so of course it contains it. A commit that cannot be absent cannot be a control.
> Re-run properly against a commit made **after** the build revision (31 exist on this branch):
> `3f1426a8d` is correctly NOT an ancestor. **The check only means something when the negative case
> is reachable** — the same demand-control discipline `444`'s verify block used, applied to git.

## 2. Criterion: the residual is CLOSED — a tick now reports its own work

The line the old binary emitted was `{"caller":"scheduler/main.go:274","msg":"Pre-query task
completed (no message fired)","task":"detected-item-promoter"}` and nothing else, on every tick,
whether it promoted twenty rows or none. What it emits now (`main.go:286`, the line moved because
the code did):

```json
{"msg":"Pre-query task completed (no message fired)","task":"detected-item-promoter",
 "pre_query_result":"{\"held\":\"16\",\"held_detail\":\"dead_fragment_link->page-build-handler
 (pair has never completed one (awaiting a hand canary)); empty_internal_href->page-build-handler
 (…); literal_markdown->page-build-handler (pair below the 25% success floor);
 missing_conversion_path->content-gap-planner (…); placeholder_contact->page-build-handler (…)\",
 \"pairs\":null,\"promoted\":\"0\"}"}
```

**And the two branches now discriminate**, which was half the point: `thunder-reaper` and
`thunder-training-monitor` take the zero-row branch at `main.go:218` (*"Pre-query found no rows —
task ran with nothing to do"*), while the promoter — which has something to say — takes the
completion branch. Before this build both said the same thing.

That is criterion-complete for the residual: not "the column exists" but **a real tick's held count
is legible in the service's own log**, which is what the guardian seat actually asked for.

## 3. What it showed on day one — the held pile is 16, not 2, across 5 pairs

Yesterday's measurement, taken through a read-only mirror because nothing reported it, was **2 rows
in 2 pairs**. Today, measured off the same predicate the promoter uses:

| pair | held | sites | oldest | pair record (ok / failed) | why held |
|---|---|---|---|---|---|
| `literal_markdown → page-build-handler` | **10** | 2 | 0.9d | 3 / 24 = **11%** | below the 25% floor |
| `placeholder_contact → page-build-handler` | 3 | 2 | **1.9d** | 0 / 4 | never completed one |
| `dead_fragment_link → page-build-handler` | 1 | 1 | 0.6d | 0 / 0 | never dispatched at all |
| `empty_internal_href → page-build-handler` | 1 | 1 | 0.5d | 0 / 1 | never completed one |
| `missing_conversion_path → content-gap-planner` | 1 | 1 | 0.7d | 0 / 0 | never dispatched at all |

**An 8× growth in one day, entirely invisible until this build.** It is not a regression — the
promoter is refusing work exactly as designed, and `literal_markdown`'s 10 rows are `444`'s floor
doing precisely the job it was built for. But "the mechanism is behaving" and "sixteen real findings
about live pages are parked" are both true, and only the second is a problem anyone can now see.

## 4. ⚠ TIME-SENSITIVE: `453`'s one-way door fires on 3 of these rows TOMORROW

Measured: `result ? 'held_pair_escalation'` is **0 across all 16** — `453` has not escalated any of
them yet, because its limit is 3 days and the oldest held row (`placeholder_contact`, 2026-08-16) is
at 1.9 days. **It crosses the limit on 2026-08-19.**

At that point those 3 rows move `detected → needs_human_review`, which the promoter never selects —
and, as recorded above, **nothing moves them back.** So they leave the automated path permanently,
joining a human queue that stands at 829 rows. If the pair is later canaried successfully, the
promoter will pick up *future* `placeholder_contact` findings and these three will still be sitting
there. The composition defect I flagged yesterday at 5 rows now has 16 rows queued into it, with the
first 3 due in under a day.

This is a decision, not a defect I should quietly patch: the door was built deliberately, and
whether a successful canary should reclaim its pair's escalated rows is the `453` author's call or
the owner's. Named here with its clock so it cannot be missed.

## 5. Where 083 now stands

| criterion | state |
|---|---|
| 1 — promotable pile drains and stays drained (corrected wording) | **HOLDS.** Promotable = 0; the 16 are all correctly *held*, not stranded-by-absence-of-a-promoter |
| 2 — `phantom_internal_link` first `complete` | met on wording, **non-discriminating**, unchanged |
| 3 — completions verified at the live page | **MET** twice (08-17 morning ×4, 08-17 evening ×3 with before/after and a negative control) |
| residual — held rows visible | **CLOSED 2026-08-18**, verified at the running service with the two log branches discriminating |

**The bug's own fix is complete and proven.** What remains is not 083's mechanism but the questions
its new visibility has surfaced — the four never-canaried pairs, `literal_markdown`'s broken handler
(`bugs_open/184` / `201`), `453`'s one-way door, and `bugs_open/300`'s unstable key. Those are named
owners or owner decisions, not this file's outstanding work.

**Recommendation: 083 can close** once `444`/`458`'s doors have sat their week (they are visibly
holding the right things today, which is the first time that sentence could be checked rather than
asserted). Everything it was filed for — a queue with no consumer — is answered, mechanically and
with artefact evidence.

### `454` PROVEN on the next tick — 2026-08-17 16:43Z

The correction to `444` (count `verified` as success, not just `complete`) released exactly what it
should, on the **first promoter tick after it applied**:

```
empty_section -> page-build-handler   2 rows   detected -> triaged at 16:43:05Z
```

Before `454` that pair read 3 `complete` / 12 `failed` = **20%** and was held by the floor; counting
its 9 `verified` rows it is 12/12 = **50%** and passes. Disconfirmable: had `454` been wrong in
either direction, the rows would have stayed held (no effect) or `literal_markdown` would have been
released too (floor disabled). Neither happened — `literal_markdown → page-build-handler` remains
held at 1+0 / 28.

`placeholder_contact → page-build-handler` correctly remains at `detected`: 0 lifetime successes, so
the known-good rule holds it, and at 1 day old it is inside `453`'s 3-day limit. It should escalate
on ~**2026-08-19**, which is the second discriminating test of that clock.

### OPEN RISK, observed 2026-08-17 18:30Z — work-item history SHRANK, cause NOT diagnosed

**The observation, measured twice.** `required_fields_missing` read **complete 64 /
needs_human_review 31** at 11:00Z and **complete 50 / needs_human_review 31** at 18:30Z. The 14
are not re-statused — `verified` is 0 and the type has only two statuses — so they left the table.
Re-measured at 18:29 and stable at 50, so it was a discrete event, not measurement drift.

**What I did NOT establish, stated plainly because the temptation is to fill it in.** I did not find
the actor. There is no `DELETE FROM site_work_items` in a scheduled task (an earlier grep of mine
that appeared to find three was matching `deleted_at`, not `DELETE` — a bad query, corrected here),
no work-item retention window in `platform/`/`internal/`/`sql_for_agents/`, and no CronJob that
names the table. The `295` lane's phrase *"the rest are past retention"* is what put me on to it,
but that may refer to `orchestration_states` (which is reaped) rather than to work items. Against a
blanket retention sweep: **the oldest surviving row fleet-wide is 2026-03-15**, and 89 completed
rows created before 2026-08-01 still exist. So this looks targeted, not systemic — but *"looks"* is
the right word and this is `[UNDIAGNOSED]`, not a mechanism.

**Why it matters here regardless of cause, which is the actionable part.** Both of this promoter's
success tests read *lifetime* history:

- 430's known-good rule — *"the pair has ≥1 `complete`"*;
- 444/454's floor — *"the pair is succeeding at ≥25%"*.

If completed rows can leave the table, **"lifetime" is really "within whatever window survives"**,
and a pair that worked well but has been quiet can lose its evidence and read as *never having
worked* — held for ever. That is exactly the latent failure `454` was applied to prevent, arriving
by a second, independent route. It is also the third instance today of one class: **the population I
was measuring was not the population I assumed** (`failed` rows carry no `completed_at`; `verified`
is a second success status; and now the row set itself is not stable).

**Not fixed here, and deliberately not patched blind.** A durable claim about a cross-cutting
mechanism is what `090`'s diagnosis loop is for (CLAUDE.md), and guessing at a guard before knowing
the actor would be a fix aimed at a mechanism nobody has identified. What a next session should do,
in order: (1) run `090` on *"completed `site_work_items` rows disappear; 14 `required_fields_missing`
completes left the table between 11:00Z and 18:30Z on 2026-08-17 with no status change"*; (2) only
then decide whether the promoter's tests need a durable counter (e.g. a per-pair success tally that
survives row deletion) rather than a live COUNT over the table.

### THE OPEN RISK IS DIAGNOSED — 2026-08-18. "Lifetime" meant the last 7 DAYS, and it was already stranding a good pair. Fixed by `465`.

The section above recorded, as `[UNDIAGNOSED]`, that completed work items were disappearing. **The
actor is `work-item-archiver`** — an ENABLED daily scheduled task (`fire_message=true`, agent
`work-item-archiver`, 86400s) whose own description reads: *"Archives terminal work items older than
7 days to `site_work_items_archive`"*.

**It is not a cleanup nobody runs.** `site_work_items_archive` holds **20,184 rows against 8,702
live** — most of this platform's work-item history is not in `site_work_items` at all. My 14 missing
`required_fields_missing` completes are there (49 archived rows of the type, all `complete`), as are
12 archived `literal_markdown → page-build-handler` rows, which is why that pair's failures appeared
to *fall* from 28 to 24 overnight.

> **Why I did not find it yesterday, recorded because the search looked thorough and was not.** I
> grepped for `DELETE FROM site_work_items` and for "retention", and checked `pre_query` bodies. The
> archiver is invisible to all three: it **moves** rows rather than deleting them, its scheduled-task
> row has a **NULL `pre_query`** (it fires a message to an agent, so the SQL is not in the table I
> was reading), and its description says "Archives", not "retention". I then reasoned from *"the
> oldest surviving row is 2026-03-15"* that there was no systematic sweep — but that row is
> **non-terminal**, and the archiver only takes terminal ones. The control I chose could not have
> come out otherwise. What found it was noticing `work-item-archiver` in an unrelated list of
> `maintenance` group-mates inside migration `458`'s header.

**Both of the promoter's success tests read `site_work_items` only, so both meant "in the last 7
days".** That is `bugs_open/083`'s own disease reintroduced by the mechanism built to cure it: a pair
that works well but has been quiet for eight days reads as *never having worked* and is held for ever.

**It had already happened.** Live-table-only versus live+archive, measured 2026-08-18:

| pair | live-only | TRUE | verdict under the old scope |
|---|---|---|---|
| `empty_internal_href → page-build-handler` | 0/1 = 0% | **9/5 = 64%** | **HELD as "never completed" while holding NINE lifetime successes** |
| `empty_section → page-build-handler` | 12/16 = 43% | **316/33 = 91%** | a 316-success workhorse reading as marginal |
| `literal_markdown → page-build-handler` | 3/24 = 11% | 3/36 = 8% | correctly held either way |
| `placeholder_contact → page-build-handler` | 0/4 = 0% | 0/6 = 0% | genuinely never succeeded |

**Migration `465`** (applied + ledger-recorded, `_ROLLBACK.sql` alongside) makes both tests read
`site_work_items UNION ALL site_work_items_archive`. Three controls, and the two negative ones carry
the weight: `literal_markdown` must **stay** held (if the archive rescued the very pair `444` exists
for, the fix would be dissolving the floor rather than correcting its scope) and `placeholder_contact`
must stay unknown (no success in either table). Both hold. A read-only mirror of the live predicate
confirms the next tick takes **exactly** the one stranded row.

Cost: 78 ms → 134.6 ms per tick on a 900,000 ms interval.

**Deliberately not taken:** a shared VIEW over both tables, which is the tidier estate-wide answer
and the right RFC. A new shared object other pipelines may adopt is a shared-seam change (owner
ruling 2026-07-28); this file's job was to stop a live pair being stranded. Named so the omission
reads as a decision.

**This is the fourth member of one family in this lane** — `failed` rows have no `completed_at`;
`verified` is a second success status; the row set is not stable; and now the row set is only a
7-day *window*. Every one is **the population measured not being the population assumed**, and none
was caught by review — twelve council seats approved `444`.
