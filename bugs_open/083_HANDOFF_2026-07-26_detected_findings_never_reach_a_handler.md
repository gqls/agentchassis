# 083 — discovery findings written as `status='detected'` never reach a handler: the promoter runs only inside a task disabled since May

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
