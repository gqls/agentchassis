# 083 — discovery findings written as `status='detected'` never reach a handler: the promoter runs only inside a task disabled since May

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
