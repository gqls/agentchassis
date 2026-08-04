# NOTES — `bugs_open/154`, work-item routing columns

Append-only, newest at the bottom. Technical log: evidence, commands, what the
system actually said, and every misstep.

---

## 2026-07-31 — session "bugfix 19", picking the bug

Asked to take the next `bugs_open/` case nobody else is on. The ownership check
is the part worth recording, because the naive version of it is wrong.

`git log` over `bugs_open/` is **lagging** — it cannot see a session mid-fix. So
I read the live `.jsonl` transcripts instead, tallying `bugs_open/NNN` mentions
in the tail of every session touched today. That produced a clear map of what 31
concurrent sessions are on (149, 157, 138, 161, 151, 156, 155, 099, 072, 139,
135, 153, 143, 125, 083, 018, 120, 017 all dominantly owned).

Then, per the memory note, I re-checked the *shortlist* by **code symbol rather
than by number** — `tool-improver|load_tool|improve_tool|tool-auditor`. That
turned up two sessions with high counts:

- `c4daed6f` (158 hits) — last written 10:05, i.e. ended. This is the `131`/`083`
  lane that *filed* 154.
- `631baa00` (15 hits) — **active**, 16:10. Worth a second look, and the second
  look mattered: it says *"I deliberately did NOT fire a cluster acceptance run …
  On a failing verdict the judge inserts an `improve_tool` work item routed to
  `handler_agent='tool-improver'` — an automated fixer, pointed at a page whose
  … [owner] … isn't mine to choose."* That is a session **declining** to touch
  tool-improver, which is the opposite of owning the fix.

**Lesson, and it is the one the memory note already makes:** a high symbol count
is not ownership. It can be a session explaining why it is staying away. Read
the surrounding text, not the tally.

## Confirming the cause rather than inheriting it

154 marked its own explanation `[INFERRED — not yet read in the code]` and named
the two things to read first. Both reads, live (`agent_definitions`, not seeds):

`build-dispatch-loop`, `process_item.sub_workflow.call_handler.input_mapping`:

```json
"component_id?": "current_item.spec.component_id"
```

`tool-improver`, `load_tool`:

```json
{"action": "query_database",
 "config": {"params": ["input_data.component_id"],
            "query": "SELECT ... FROM content_components cc JOIN page_components pc ON pc.component_id = cc.id JOIN pages p ON pc.page_id = p.id WHERE cc.id = $1::uuid AND cc.is_active = true LIMIT 1"}}
```

Then `LoadWorkItemsAction`'s SELECT — and there it is:

```sql
wi.severity, wi.summary, wi.spec, wi.page_id,
```

`page_id` and **not** `component_id`, though `\d site_work_items` shows
`component_id | uuid` as a real column alongside `entity_id` and `affected_url`.
So `current_item` never carries them, and `current_item.spec.component_id` is the
only reachable path. `ResolveInputMapping` skips an unresolved `?` path silently
(`input_mapping.go:122-129`), so the value dies with no log line and surfaces two
agents later as a nil query param.

**The framework, not either agent, is what drops the value.** The creator that
used the schema properly is the one whose items cannot be dispatched.

## Census, 2026-07-31 (all `site_work_items`)

| column | column set | column set AND no spec key | spec only |
|---|---|---|---|
| `component_id` | 21 | **16** | 235 |
| `entity_id` | 0 | 0 | 12 |
| `affected_url` | **0** | 0 | **0** |
| `page_id` | 1681 | 100 | **218** |

`improve_tool` alone: **4 of 4** `tool-auditor` rows are column-only, all
`failed`/`wont_fix`; **16 of 16** rows from `tool-acceptance-agent`,
`design-discovery-agent` and `generic` are spec-only and fine.

**Still live:** the newest row is `a5d11c86`, robot-hands.com, dated
**2026-07-31** — a fresh failure one day after filing. The bug is not stale.

## The check that could have made the whole fix inert, and nearly went unrun

`load_tool` queries `content_components`, but `page_components.id` is a
*different* id — and 016b already records "do not key re-validation on a stored
`component_id`" because `page_components.id` is not stable across re-renders. If
`tool-auditor` had been writing a `page_components.id` into the column, then
delivering it would satisfy the mapping and **still** find no row, and I would
have shipped a fix that changed the error message and nothing else.

```
item     | in_content_components | active | in_page_components | joins_to_a_page
a5d11c86 | t                     | t      | f                  | t
ee745694 | t                     | t      | f                  | t
7c2d898a | t                     | t      | f                  | t
5b4fd5cc | t                     | t      | f                  | t
```

4/4 satisfy **every clause** of `load_tool`'s query. The fix is not inert.
`[VERIFIED]` — query above, run 2026-07-31.

## The design candidate I rejected, and what killed it

The obvious fix is to **backfill the column value into the item's `spec` map**:
it needs no config change at all, because `current_item.spec.component_id` would
simply start resolving. I had written most of the argument for it before
checking who else reads `spec.component_id`.

`rerender-pages` does. And `create_rerender_items_action.go:219`:

```go
scoped := (reason == "section_data_resolved" || reason == "image_landed") && componentIDStr != ""
```

So a write into `spec` can flip a **site-wide rerender into a component-scoped
one** — in a path with nothing to do with this bug. Rejected on that.

**Misstep worth recording:** my first instinct was to reassure myself with the
data — the two live `needs_rerender` rows in that shape both carry an empty
`reason`, so `scoped` would be false today anyway. That is a *data-dependent*
safety argument about a *code* property, and it goes stale the moment a rerender
is raised with one of those two reasons. The design should not rest on it, and
the note in the code says so. I kept the structural argument and threw the
comforting measurement away.

## What I deliberately did NOT change

- **`page_id` keeps its column-only behaviour.** Extending the spec fallback to
  it "for symmetry" would newly expose `current_item.page_id` on **218** rows.
  Widening what reaches a handler changes it without editing it. There is no bug
  here to close — `page_id` is already exposed — so there is no reason to.
- **The key stays ABSENT, not `""`,** when neither source has a value. An
  optional mapping path that *resolves* is forwarded; one that is *missing* is
  skipped. Materialising `""` would turn "not supplied" into "supplied as empty"
  for handlers that gate on presence — `create_rerender_items` again.
- **`tool-auditor`'s site-scoped `item_key`** (`audit_fix_<domain>`, one key per
  site on a per-tool fix) — 154's second finding. A separate defect in item
  *creation* with fleet-wide dedup consequences, marked `[UNVERIFIED]` in the bug
  file as possibly intended. Left open and recorded, not silently dropped.

## Reuse, caught by the compiler

Wrote a `uuidPtrString` helper; `go build` refused it —
`fork_theme_from_site_action.go:450` already had a byte-identical one. Deleted
mine. Small, but it is the CLAUDE.md rule ("reuse existing machinery before
building new") being enforced by the toolchain rather than by me remembering.

## Proving the test can fail

A passing test proves nothing until you have watched it fail. Deleted the
`component_id` exposure line and re-ran:

```
--- FAIL: TestLoadWorkItems_ExposesRoutingColumns
    column-only item: component_id = <nil>, want e56f1d80-… — this is the bugs_open/154 failure
    spec-only item:   component_id = <nil>, want c5bb011e-… — the fallback regressed the 235-row majority
```

Both arms fire — the bug itself **and** a regression of the majority path.
Restored, full `./platform/orchestration/...` suite green.

## Filed and submitted

- Diagnosis loop (`090`), per CLAUDE.md's default for a durable structural
  claim: `RUN_CORRELATION_ID=21758756-d7b3-444a-844e-b37e09b5c9ce`.
  **Retention trap, exactly as 154 warned:** the `orchestration_states` rows for
  it were already gone ~25 minutes later. Track it via
  `site_work_items` (`7d5d34eb`) and `diagnosis_artifacts` instead.
- Council gate: `SUBMISSION_CORR=10be5ed9-3bd0-45ed-b6bb-4385a887967d`.
  Committing before the verdict, so the commit carries `Council-Submitted:`,
  never `Council-Reviewed:` on a verdict I have not read.

## 2026-07-31, later — the diagnosis loop CONFIRMED it, and found a better citation than mine

`21758756-…` returned **CONFIRMED** on the first iteration, re-reading
`LoadWorkItemsAction` independently and quoting the same SELECT and the same
`item := map[string]interface{}{…}` literal. Worth recording because it did not
merely agree — it produced a **contrast pair on one site** that is cleaner than
anything in my own write-up:

| item | column | spec | status |
|---|---|---|---|
| `a5d11c86` | set (`121a2f4f…`) | `{}` | **failed** 13:21:26, one second after the `load_tool` "resolved to nil" line |
| `265f0c41` | NULL | `component_id` present | **complete** |

Same mechanism, both directions, same site, an hour apart. All four symptom
points came back `[explained]`.

**Lesson:** I had the census (4/4 vs 16/16) and treated that as sufficient. A
census establishes a correlation across a population; a *matched pair* shows the
mechanism operating in both directions. Cheaper to read, harder to argue with,
and I did not think to look for one. This is the second time this lane has been
reminded that a perfect discriminator across a population is not the same as
evidence about the mechanism (see the `072` note in memory).

## The council round I wasted, and why the validator was right

Round 1 died at `persist_submission`, before any seat ran:
`edit 4: file path must be repo-relative with no traversal or whitespace`.
Edit 4 was the config half, and I had written
`"file": "agent_definitions/build-dispatch-loop (live DB row, NOT in this commit)"`
— prose, in a field that takes a path, trying to be honest about the change not
living in the repo.

The validator was right in a way I did not expect: a config change with **no
file** is not reviewable and cannot be rolled back. "It lives in the DB" was a
description of a gap, not an excuse for one. Rewriting it as
`sql_for_agents/278_dispatch_loop_reads_component_id_column.sql` made the
submission valid *and* the change better — it now carries a pre-image snapshot,
`jsonb_set(create_missing=false)` so a drifted mapping fails loud instead of
growing a second key, an idempotent `WHERE` matching the exact prior value, the
pod-grep gate with its positive control, and a rollback statement.

Cost: one dispatch, zero reviewer time. Logged in `WRONG_CALLS.md`.

## Status at hand-off

Fix committed (`4667db235`, `9bde2aa14`) and **not live**. Per CLAUDE.md's bar —
*fixed AND live* — 154 stays in `bugs_open/`: the defect is still reproducible
until the image ships, and `a5d11c86` reproduced it **today**. What remains is
the roll, the pod-grep on every replica, then `sql_for_agents/278`. In that
order, for the reason written at the top of that file.

## 2026-07-31, 19:40 — council APPROVED, and four objections that improved the migration

`10be5ed9` → **APPROVED, 6 advisory objections, none high-severity**, round 1 of
the valid submission (8 seats ran). Because the config half had **not been
applied yet**, answering the objections cost nothing — which is an argument for
submitting the DB half alongside the Go half rather than after it.

**Two objections I could only answer by running a check, and both were right to ask:**

- `editquality` (medium): *the jsonb_set path assumes a specific depth … same
  shape as a known landmine.* It was **assumed**, not proven. Ran it read-only:
  the path resolves to `current_item.spec.component_id` at exactly that depth.
  Verified — and step 1 of the migration now re-asserts it **at apply time**,
  because the row is shared and mutable and my comment is not evidence.
- `tooling_provenance` (medium): *mutates `default_config` with a raw UPDATE; the
  platform has `snapshot_agent()` for exactly this.* Checked — it exists, two
  overloads. Now used. My first draft's "snapshot" was a `SELECT` whose output
  lived in my scrollback, which is not a snapshot at all. Same reuse rule the
  compiler enforced on `uuidPtrString` earlier, caught by a reviewer this time.

**The objection that was sharper than my own answer would have been.**
`bug_historian` (medium): the Go change exposes **three** columns, the migration
rewires **one** mapping — so `entity_id`/`affected_url` sit in the "fixed the
mechanism, forgot the sibling call site" position. I would have replied "zero
rows carry them, so there is nothing to route". Measuring it properly gave a
better and less comfortable answer:

| | column set | who reads it |
|---|---|---|
| `entity_id` | 0 rows | 1 agent — `asset-deployer`, via `input_data.spec.entity_id` |
| `affected_url` | 0 rows | nothing |

`asset-deployer` reads the **spec passthrough**, not a dispatcher mapping — so
the column-first coalesce **cannot reach it at all**, and `build-dispatch-loop`
maps no `entity_id` in the first place. The first creator to write `entity_id`
on the column hits this identical bug, and fixing it then takes **two** edits,
not one. Nothing is broken today (no failing population), so pre-fixing it would
be the same speculative widening I refused for `page_id` — but **my "close the
door on the class" claim was too broad and is now narrowed in place** in both
the register (WDS-014) and `LANDMINES.md`.

**Recorded, not actioned:** `guardian` (medium) — edit 4 should have been
`operation: config_change` naming the owning pipeline, not `add`; the file now
states its surface explicitly. `architecture` (medium) — the three new top-level
keys are a shared wire-shape change to `current_item`, and **no RFC describes
that map as a contract**; the seat explicitly did not call for a block, and by
the 2026-07-29 owner ruling an addition that is additive-and-inert (measured: 0
agents referenced the new path) is not architecture-scope. Its real finding is
the missing contract doc, which is now the register's open review question.

**No resubmission**: APPROVED with nothing high-severity. The revisions above
went into the un-applied migration and the two docs, not into a new round.

## 2026-07-31, ~21:45 — the image landed; both halves are now LIVE

**Go half — pod-grepped on BOTH replicas, `v1.0.1219`:**

```
=== agent-chassis-59cb674798-t7dgn ===   NEW: 1   CTRL: 1
=== agent-chassis-59cb674798-z84n8 ===   NEW: 1   CTRL: 1
```

`NEW` = `"routing field left unset"` (a string this change added).
`CTRL` = `"LoadWorkItemsAction: Starting"` (pre-existing). The control is the
point: without it a `0` on the new symbol is ambiguous between "not shipped" and
"my grep is broken", which is the `bugs_open/153` trap.

Timing reconciles independently: commits at **18:25:37Z / 18:27:55Z**, pods
started **19:09:31Z / 19:09:52Z**. Recorded because the commit timestamps display
in `+01:00` and the pod times in `Z` — comparing them raw makes the fix look like
it *post-dates* the roll by 18 minutes. It does not.

**Config half — `sql_for_agents/278` applied:**

```
rows_to_change_expect_1 → 1
NOTICE: Snapshot captured: type=build-dispatch-loop, source_version=1,
        source_id=099b51e0-6dd0-4856-8f82-805a379e8b1d
UPDATE 1
build-dispatch-loop | current_item.component_id
COMMIT
```

The pre-flight count and the `snapshot_agent()` call are the council's
`debug_historian` and `tooling_provenance` objections, and both earned their
place: the count proved the row was still the shape the migration was written
against (it is shared and mutable, and ~4 hours had passed), and the snapshot is
a real rollback artefact rather than a `SELECT` in a terminal.

## Verifying at the artefact — and why it did not complete inside this session

Owner's call (asked, because dispatching runs an automated LLM rewrite against a
live customer tool): dispatch **one** item on `gamesdesign.co.uk` —
`ee745694`, `tool-bayesian-ranking`. Chosen over robot-hands.com, which several
active lanes are working.

Reset to `triaged` with `attempt_count=0`. **That reset is legitimate rather than
an override of a judgement:** all three of its failures were caused by *this*
bug, so the attempts were spurious. `5b4fd5cc` was deliberately NOT touched —
it is `wont_fix`, which is a human decision, not collateral damage.

Pre-state captured for the change-detection:
`tool-bayesian-ranking` `c345a76a`, `md5=5de92eba982c30315b3886096b52dd87`,
9158 bytes, `updated_at=2026-07-29 18:48:23Z`.

**It did not dispatch within ~20 minutes, and the reason is not this bug.**
Checked rather than assumed:

| check | result |
|---|---|
| dispatch lane alive? | `build-dispatch-loop` last claimed **28 min ago** — alive, slow. (Grouped by `claimed_by`; a bare `max(claimed_at)` would have shown 168 min and blamed `diagnose-dispatch-loop`.) |
| site eligible? | `gamesdesign.co.uk` not locked, `deployed`, **0** claimed items (so not the `NOT EXISTS` whole-site blocker), **36** dispatchable |
| item reachable in the 5-item window? | **rank 2 of 36** (`priority 60`, behind one `nav_drift` at 30) — comfortably inside `max_items: 5` |

So the item is correctly queued and correctly ranked; what it is waiting on is
`find_dispatchable_site` picking **this** site, which is one site per tick and
effectively arbitrary among eligible sites (WDS-002 — lowest-UUID sites can
starve others). Nothing here is evidence about the fix either way.

**[UNPROVEN AT HAND-OFF]** the live join of the two halves on a real row. What IS
proven: the coalesce is in both running binaries, the mapping reads the column
path, the induced-fault test covers the exact column-only shape, and all four
stuck items' `component_id`s satisfy every clause of `load_tool`'s query. The
single remaining question is whether `load_tool` clears on a real dispatch.

## 2026-07-31, 22:42 — the watcher expired without a dispatch; the gap is now 70 min

40 polls, 22:02 → 22:42 local, every 60s. `ee745694` stayed `triaged`,
`attempt_count=0`, unclaimed, no error, throughout. Then, DB-computed:
`mins_since_last_claim_anywhere = 70`.

**The sequence of gaps I have now measured: 28 → 39 → 40 → 70 minutes.** It is
growing, and nothing has been claimed fleet-wide across the whole watch.

**I am deliberately NOT concluding from this**, because I have already read this
same signal wrong twice today in opposite directions (see `WRONG_CALLS.md`: first
"alive, slow", then a confidently wrong "dead — 90-minute drought" built on a
local-vs-UTC arithmetic error, then "bursty" from the histogram). What I will
record is the measurement and the discriminator, and nothing beyond it:

- The longest gap already observed today, from the histogram, is the ~90-minute
  quiet spell **18:30–20:00Z**, which ended in a burst of 19 claims.
- The current gap is **70 minutes**. So it is **inside** previously-observed
  behaviour, at the top of the range.
- **The discriminator is already in the data:** if the gap passes ~90 minutes and
  keeps climbing, that is outside anything seen today and is a real dispatch
  problem. If a burst arrives first, it was another lull.

That check belongs to whoever owns dispatch, not to this lane. The one thing this
lane must not do is convert "my verification has not fired" into a diagnosis of
someone else's subsystem — which is precisely the mistake I made at 22:10.

**Standing position for `154`:** both halves live and verified (re-verified across
a second roll to `v1.0.1223`); the fix is inert-until-exercised, not unproven; the
outstanding item is one observation, blocked on dispatch scheduling. The bug file
and the handoff both now say exactly that, and neither claims the observation was
made.

## 2026-08-01 08:02Z — the discriminator fired: it WAS a real stall, 11.5 hours

Overnight the gap ran to **688 minutes** — far past the ~90-minute maximum I had
set as the discriminator. So the pre-registered test resolved cleanly, and it
resolved *against* my third reading:

| reading | verdict |
|---|---|
| 1. "alive, slow / intermittent" | wrong |
| 2. "dead — a real drought" | **directionally right, arrived at wrongly** (local-clock-vs-UTC arithmetic; the 90 min I quoted was really 31) |
| 3. "bursty; 39 min is within observed range" | wrong — correct about the histogram, wrong about what it predicted |

**The honest reckoning: reading 2 reached the right conclusion from a broken
measurement, and I retracted it for that reason. Retracting it was still
correct.** A conclusion that happens to be true, computed from arithmetic that is
false, is not knowledge — it cannot be relied on, and the next inference drawn
from it inherits the error rather than the luck. What actually settled this was
not any of the three readings but the **pre-registered discriminator**: "if the
gap passes ~90 min and keeps climbing, that is outside observed behaviour."
Writing that down *before* the data arrived is the only reason this ended in a
fact instead of a fourth opinion.

**Lesson for the lane:** when a signal has already fooled you twice, stop
producing readings and publish a falsifiable threshold instead. The threshold
costs one sentence and it is the only thing here that survived contact.

**Timeline, DB-computed throughout:**

```
2026-07-31 20:33:49Z   last claim before the stall (19 claims in that hour)
     ...  11h 29m with ZERO claims fleet-wide, 354-380 items dispatchable ...
2026-08-01 08:02:45Z   claims resume (3 in the 08:00 hour)
```

Through all of it `build-pipeline-trigger` stayed enabled, fired every 120s and
**completed**, and all five reapers stayed enabled — including
`claimed-item-timeout` at 120s, which I watched clear a stuck row between two
queries a minute apart. So the stall was **not** a dead scheduler and **not** an
unreaped claim, and I am recording that as the two hypotheses it rules out —
nothing more. **The cause is still unknown and is not this lane's to find.**

**Still not diagnosing it, and now with a stronger reason:** several other
sessions are active on dispatch symbols right now (`693556a1`, `3bec7dd7`,
`956e5263`, `078d18fd`, `8871a7d4`, `957623aa`), so an 11.5-hour fleet-wide stall
is very likely already being worked by someone who owns it. The contribution from
here is the measured window above, not a theory.

**For `154`:** dispatch is live again as of 08:02Z and the item is still `triaged`
at rank 2 on its site, so the outstanding observation is finally *possible*.
Watching now.

## 2026-08-01 08:33Z — dispatch recovered, but the target site is being STARVED

Second watch: 30 polls, 08:03 → 08:33Z. `ee745694` never left `triaged`.

Dispatch is demonstrably working — claims in the 45 minutes after recovery:

| site | claims | window |
|---|---|---|
| loancalculator.co.uk | 15 | 08:04–08:14 |
| finetuning.uk | 5 | 08:07–08:10 |
| robot-hands.com | 5 | 08:01–08:09 |
| vetcomparison.uk | 5 | 08:07–08:10 |
| system.internal | 3 | 08:02–08:19 |
| **gamesdesign.co.uk** | **0** | **never selected** |

gamesdesign was eligible throughout — re-checked at 08:30: unlocked, `deployed`,
**0** blocking claims, **36** dispatchable, and `ee745694` still at **rank 2**.
Five other sites were served repeatedly and this one was not selected once.

**This is the starvation mode the register already documents** (WDS-002:
`find_dispatchable_site` picks ONE site per tick via `DISTINCT ON` with **no outer
`ORDER BY`** — "effectively arbitrary among eligible sites, and lowest-UUID sites
can starve others"). I am naming the match, not diagnosing it: I have not read the
selection SQL this session, so **`[INFERRED]`** that this instance is that
mechanism. What is **measured** is the table above.

**Consequence for `154`, stated plainly: waiting is not a terminating strategy.**
The verification has been queued for ~12 hours across two watches and the site has
never been picked. It may be picked in the next tick or not this week; nothing in
the queue's state predicts it, because the selection does not consider queue age
or depth. So "wait for a natural dispatch" is not a plan, and continuing to sit on
it would be mistaking patience for progress.

Two honest ways forward, and the choice is the owner's because they differ in
blast radius, not in confidence:

1. **Fire `build-dispatch-loop` at gamesdesign by hand.** The platform's own
   mechanism for exactly this (the queue-stalled bypass pattern). **But it is
   BROADER than what was approved:** the loop loads `max_items: 5`, so it would
   process up to five of that site's queued items — rank 1 is a `nav_drift`, ranks
   3+ are `page_rerender`s against a live site — not only mine. Legitimate queued
   work the platform would do anyway, but not "dispatch one item".
2. **Close on the mechanism evidence, with the gap stated.** Both halves live and
   pod-verified across two rolls, mapping reads the column path, induced-fault
   test covers the exact column-only shape, and all four stuck items' ids satisfy
   every clause of `load_tool`'s query. The close would say **no live dispatch was
   observed** — which is honest, and leaves the last step for whoever next sees an
   `improve_tool` item run.

I am NOT taking option 1 unilaterally: the earlier go-ahead was for one item on
one site, and five items including live-page rerenders is a different action.

---

## 2026-08-02 — starvation MEASURED, then FIXED: `284` makes dispatch oldest-waiting-first (owner-directed)

Asked "what is starvation mode?", I read the live selector instead of answering
from the register — and the `[INFERRED]` marker above is discharged, twice over:
it was starvation, and it was **worse than WDS-002 said**. `DISTINCT ON (site_id)`
forces `ORDER BY` to lead with `site_id`, so the old order was not "effectively
arbitrary" — it was **deterministically lowest-UUID-first**. `priority` never
influenced which site won (it picked each site's representative row and was then
projected away). Measured: 17 eligible sites; gamesdesign 14th by UUID, holding
the fleet's OLDEST eligible item (3d10h, double the runner-up); robot-hands
(1st by UUID) winning every idle tick on a priority-**110** item ahead of a
priority-**5** item at mortgagecalculator.

> **CORRECTED 2026-08-02:** my 07-31 entry above says "nothing in the queue's
> state predicts" when gamesdesign would be picked. False — the state predicts it
> exactly: picked iff all 13 lower-UUID sites are simultaneously busy-or-drained.
> I wrote "unpredictable" without having read the one query that decides it.
> Caught by reading `find_dispatchable_site`'s SQL (one command). → WRONG_CALLS.

**Owner ruling (this session): fix the starvation gap, and rid the trigger of the
dead `wi.domain` column.** Fairness rule chosen by the owner from four costed
options (FIFO / aging / round-robin / priority-major): **oldest-waiting-first**.

**The fix — `sql_for_agents/284`, applied 2026-08-02 ~09:36Z, config-only (no
image dependency, live immediately):**

- New selector: `ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1`,
  `DISTINCT ON` dropped (redundant under LIMIT 1 — the oldest item's site IS the
  site whose oldest item is oldest). Output shape unchanged.
- Starvation-free by construction: a new item's `created_at` ≥ any waiter's, so
  no future arrival outranks a waiting item indefinitely — bounded wait, not
  better odds. Tiebreaks are load-bearing: batch inserts share one transaction
  timestamp, so `created_at` alone is not a total order.
- Deliberately NOT done: cross-site priority (never existed; priority-major
  recreates starvation keyed on priority, aging needs an owner-agreed scale) and
  any change to within-site claim order (still priority ASC, created_at ASC).
- Known bounded bias, stated up front: within-site claiming is priority-major, so
  a site keeps its old fairness key until its oldest item drains and can be
  re-picked whenever idle — bounded by ceil(backlog/5) batches, biased toward the
  most-starved site, which is the correct direction.
- Apply evidence: pre-flight count **1** (md5-guarded against the exact old query
  text), `snapshot_agent` → `bb291e23-…` reason "WDS-002 fairness ORDER BY (284)",
  `UPDATE 1`, verify shows the new text. Idempotent: the md5 guard makes a re-run
  (or a bulk `--apply` by another session) a 0-row no-op.
- **Council: not submitted** — the gate's scope is `platform/`,`internal/`,`pkg/`
  and refuses docs/config-only submissions client-side; this change has no Go
  half. Owner authorised directly in-session. Register updated in the same
  commit (WDS-002 entry, which also corrects its own "effectively arbitrary").

**Verified at the artefact, first tick:** 09:36:35Z, `build-dispatch-loop`
claimed `2c3a203a` (`nav_drift`) on **gamesdesign.co.uk** — the exact site the
new ORDER BY test-run predicted, after 3½ days of never being selected. The
target item `ee745694` is rank 2 in the same batch.

**Second finding, same session — the dead column in seed 052 (owner's first
instruction).** `site_work_items.domain` was renamed `pipeline` (WDS-003); live
config lost the filter via migration 067 and the live scheduler pre_query reads
`wi.pipeline` — but seed `052_build_pipeline_trigger.sql` still carried
`wi.domain = 'build'` in BOTH operative INSERT strings (a re-seed would produce a
trigger that errors at RUNTIME on its first tick, not at INSERT time) **and in an
operative `UPDATE scheduled_tasks` statement that would REINTRODUCE the dead
column into the live scheduler row if re-run**. All three corrected; the two
`-- backup` psql dumps and migration-067's own removal machinery keep their
`wi.domain` text deliberately (history, and the pattern being removed). Fleet
check: no LIVE `agent_definitions` row references `wi.domain` anywhere.

Observed while here, not acted on: the scheduler pre_query gates on
`status='triaged'` only (not `approved`) and `s.locked_at IS NULL`, while
`find_dispatchable_site` accepts both statuses and ignores locks — an asymmetry
that can skip a tick for approved-only queues. `[UNMEASURED]` whether any site is
currently affected; noting, not fixing.

---

## 2026-08-02 — WITNESSED, and CLOSED. The last step happened by itself, 4 minutes after 284 went live

Timeline, all times UTC, all captured live (orchestration retention is minutes):

- 09:36:35 — first tick under the new selector claims gamesdesign's rank-1 item
  (`2c3a203a`, nav_drift). The site nothing had picked in 3½ days, picked in
  under 2 minutes.
- 09:39:56 — `ee745694` claimed (rank 2, same batch). **This is the witness item:
  tool-auditor-raised, column-only, the exact structurally-undispatchable shape.**
- 09:40:13 — orchestration `76c1a3e1` (`tool-improver`) starts.
- 09:40:48.17 — `content_components c345a76a` (`tool-bayesian-ranking`)
  rewritten: `html_template` 10,520 chars (full shape, not a truncation stub).
- 09:40:48.41 — `section_edit` item `3971205f` created BY tool-improver.
- 09:40:53 — work item `complete`, `error` NULL. 57 seconds end-to-end, on the
  item class that used to die ~1s after claim with `resolved to nil`.
- Orchestration final state: `current_step=complete`, `COMPLETED`, `error` NULL,
  `__step_error` EMPTY (checked deliberately — bugfix-099's landmine), outputs
  present for EVERY step incl. `load_tool`, `improved_html`, `update_component`.

Three verification missteps worth their lines, none fatal:
1. My watcher's orchestration query used `id` — the column is
   `orchestration_id`. The watcher still caught the work-item transition; the
   orchestration evidence came from the manual follow-up. Schema first, even in
   throwaway watchers: the trap fired twice more in the same session
   (`scheduled_tasks.name` not `task_name`, `content_components` not
   `site_components.name`).
2. My first "batch progress" query filtered `NOT IN (terminal…)` but forgot
   `needs_human_review` is non-terminal-non-dispatchable — it showed 6 irrelevant
   rows and none of the ones I wanted. The filter defined the world again.
3. The component lookup tried `site_components` by two different columns (0 rows
   both) before asking `information_schema` where `component_id`-shaped things
   live. Tool components are `content_components` rows; `site_work_items
   .component_id` points at `content_components.id` for improve_tool items.

**CLOSED**: file moved to `bugs_closed/` in this commit (both paths named — the
git-mv landmine). The closing bar "fixed AND live" was met 07-31; today adds
"witnessed", which is what the file said it was waiting for.

Left open deliberately, recorded in WDS-014 and the bug file: the
`entity_id`/`affected_url` siblings (fix on first failing row), the `page_id`
asymmetry (218-row blast radius, owner call), tool-auditor's site-scoped
`item_key` (own measurement wanted), and the scheduler pre_query asymmetry noted
above (`triaged`-only + `locked_at`, vs the selector's `triaged|approved`, no
lock check) `[UNMEASURED]`.

---

## 2026-08-02 — verifying 284 uncovered a SECOND defect: the selector and the loader disagree about "dispatchable" (`285`)

284 worked, and watching it work is what exposed this. Post-284 the fleet served
gamesdesign (5 claims, 09:36–09:41) and then **went silent again for 68 minutes**
while `build-dispatch-loop` ran 16 times and completed cleanly every time.

**The mechanism.** Two queries decide dispatchability and they are not the same
query:

- SELECTOR — `find_dispatchable_site` — picks the SITE on three predicates.
- LOADER — `LoadWorkItemsAction` (`load_work_item_actions.go:~624`) — picks the
  ITEMS, on those three **plus two**: `(COALESCE(approval_mode,'auto')='auto' OR
  status='approved')` and the `depends_on` unresolved-dependency clause.

So the selector hands the loop a site whose only eligible item the loader
refuses. The loop loads zero, reports `has_items:false` / `rows_dropped:0`,
notifies the scheduler and completes **COMPLETED, no error**. No claim → the site
is still eligible → picked again. The queue never advances and nothing logs a
failure. `rows_dropped:0` is the tell: not "dropped after loading", but "the SQL
never matched a row another query calls eligible".

**Measured.** robot-hands.com held exactly one selector-eligible item, `93f2a3b7`
(content_rewrite, created 07-31 12:27:28 — the fleet's oldest), depending on
`0733a7a4`, a `needs_content_page` in **`needs_human_review`**: a state no
automation ever moves to complete/verified. Fleet-wide that was **1 blocked row
out of 366** across 17 sites — and because it sat at the head, it stalled
everything.

```
08:03–08:06  robot-hands drains its last 5 claimable items
08:06–09:36  ZERO claims fleet-wide (89 min) — lowest-UUID selector picks
             robot-hands every tick; every loop loads 0
09:36        284 live → gamesdesign served (starved 3d10h), 5 claims
09:41–10:14  ZERO claims again (68 min) — robot-hands back at the head, now
             by AGE, and permanently
10:14        285 live
10:16–10:22  relojistas 5, vetcomparison 2, webdesign 1 — exact FIFO order
```

**Why 284 made it worse, and why the two changes are only safe together.** Under
lowest-UUID a blocked site held the head only while it happened to sort lowest.
Under oldest-waiting-first the key is `created_at` — it never changes and only
ages — so **an unloadable item, once at the head, is at the head for ever.**
284 without 285 is a permanent fleet stall behind one row. Fairness ordering was
still the right change; it is just not sufficient on its own, and anyone
reasoning about either in isolation will get its blast radius wrong.

**The fix (`sql_for_agents/285`, applied 10:14Z).** Give the selector the
loader's two clauses verbatim. **Strictly narrowing** — every site it removes is
one where the loop would have loaded 0 items and claimed nothing — so no site
that would have been served loses service; only the wasted pick that blocks
everyone behind it goes. Pre-flight 1, snapshot, `UPDATE 1`, verified.

**Proof at the artefact**, before/after on the loop's own output rather than on
run status (every run was COMPLETED throughout, which is exactly why status is
useless here):

```
site              item_count   when
robot-hands.com   0            10:03, 10:06, 10:08, 10:11, 10:13   <- pre-285
relojistas.com    5            10:16                                <- post-285
vetcomparison.uk  5            10:19
webdesign.co.uk   1            10:21
```

Negative control, deliberately chosen so a fix that "worked" by simply dispatching
more could not pass it: `93f2a3b7` must remain `triaged`, unclaimed,
`attempt_count 0` — it is genuinely undispatchable and the point is that the queue
stops **waiting** for it, not that it gets dispatched. Confirmed.

**Deliberately NOT fixed, recorded instead:**
- The loader's dependency subquery is **site-scoped** (`WHERE site_id = $1`), so a
  cross-site `depends_on` can never resolve. 285 copies that faithfully — the
  selector's job is to AGREE with the loader, not to be independently right.
  Fixing it means changing Go and moving both queries together.
- robot-hands' actual blockage: `0733a7a4` needs a person. That is work-item
  triage, not a dispatch defect.
- The loader's optional `item_pipeline`/`handler_agent` filters are NOT mirrored —
  the live `load_items` config carries neither, so mirroring them would invent a
  filter the loader does not apply, and re-introduce the dead `domain='build'`
  clause migration 067 removed.

**Misstep, logged in WRONG_CALLS:** I had twice recorded these gaps as "comparable
to the ~90-minute quiet spell already observed, so not yet outside known
behaviour". There was no known behaviour — only a symptom I had seen twice and
never explained, which I turned into a baseline by comparing the second instance
to the first. I measured the symptom (time since last claim) three times with
increasing precision and never once read what the dispatcher had decided; one
query against `collected_data->'load_items'` would have ended it.

---

## 2026-08-02 — triage of the blocked pair (`286`), and the generator defect behind it (`bugs_open/177`)

Owner asked for the blocked item to be triaged. Findings, in the order they
changed the disposition:

1. **`0733a7a4` is SPURIOUS, not slow.** tool-generator built the tool page, its
   component and its `page_components` row, then raised "write content for this
   page" **45 ms later** (component 12:27:28.312 → page .326 → slot .342 → item
   .387). The page is `deployed` with `deployed_at` set and a 10,336-char tool
   component.
2. **One slot IS the finished shape of a tool page** — the measurement that makes
   it spurious rather than early: of 141 tool pages fleet-wide, **116 have exactly
   1 slot**; all five deployed robot-hands tool pages have 1. So page-build-handler
   found nothing to build because there is nothing to build, and its refusal to
   report success is WDS-004 working, not failing.
3. **8 of 8 `tool_content:%` items ever created are in needs_human_review**, each
   at attempt 1, byte-identical error, 07-14 → 07-31, 4 sites. 0% success over the
   class's entire history ⇒ `bugs_open/177`.
4. **`93f2a3b7` is REAL work**: verified none of the 3 `page_components` on
   `/how-to-specify-a-gripper.html` contains `gripper-safety-factor-calculator`.
   It is also the FIRST `tool_crosslink` ever to carry a `depends_on` (the other
   5 have none) — so the coupling is new and spreading.

**The disposition trap, and it is the durable lesson here: a dependency can only
be released by `complete`/`verified`.** The loader's clause is `dep_id NOT IN
(SELECT id … WHERE status IN ('complete','verified'))`, so `wont_fix`,
`rejected`, `cancelled` and `failed` ALL leave the dependent blocked for ever.
There is no "dismissed" state a blocker can reach. That makes `complete` the
*convenient* disposition for any awkward blocker — which is exactly why it must
be refused when no work was done. Marking `0733a7a4` complete would have been the
silent-completion pathology (WDS-004) committed by hand, by me, for convenience.

So: `0733a7a4` → **`wont_fix`** (true statement: should never have been raised),
original handler error preserved inside the new reason string; `93f2a3b7` →
**`depends_on` cleared**, left `triaged` to be attempted on merit. Neither
fabricates a success; neither abandons real work. `286`, 5 pre-flight assertions
all true, `UPDATE 1` + `UPDATE 1`, verified before commit.

**Prediction recorded BEFORE the outcome, so it cannot be retrofitted:** 5 of 5
previous `tool_crosslink` items failed at `validate_page_content`. If `93f2a3b7`
does the same, that is a separate defect in the crosslink path — **do not
re-diagnose it as a dependency problem**, which is the trap now that the
dependency is the thing I just touched.

**Not swept, deliberately:** the other 7 `tool_content` rows. They block nothing
(only `0733a7a4` ever had a dependent), so there is no urgency, and sweeping 7
rows across 4 sites is a different action from triaging the one that stalled the
fleet. Their disposition belongs with 177's fix.

### And a prior-art check I should have run before writing a line of 284

`bugs_open/169` part B, filed 2026-07-31 by the dartsonline lane, **had already
found the UUID starvation** — same query, same reading, same conclusion — and
closed by explicitly asking for the owner ruling ("should site selection be FIFO
by oldest-eligible-item, or priority-weighted across sites?"). It also already
flagged the `wi.domain` seed drift. I found neither, because I went from the
owner's instruction straight to the live config without grepping `bugs_open/` for
the mechanism — the check CLAUDE.md puts first and that my own memory file
records me getting wrong before.

It cost nothing this time, and only by luck: the owner's answer *was* the ruling
169 was waiting for, so the work was the completion of that thread rather than a
duplicate of it. 169 is now updated — part B fixed, part A (the spawn hang)
explicitly untouched and still open. **The lesson stands regardless of the
outcome:** an instruction from the owner is not a reason to skip prior art, and
"I was told to do this" would have been no defence if another session had been
mid-fix on the same query.

---

## 2026-08-02 — the triaged crosslink SUCCEEDED, my prediction was wrong, and the success hid a content regression (`bugs_open/178`)

Claimed 10:38:54 (33 s after `286` committed), `complete` 10:42, attempt 0, no
error. **It added the link correctly** — one contextual anchor to
`/tools/gripper-safety-factor-calculator/index.html` in each of the three slots,
pointing at the right URL, with sensible prose around it. So robot-hands is a
normal dispatch participant again and the crosslink work is genuinely done.

**My recorded prediction (validate_content failure, on the 0-of-5 record) was
WRONG.** Recording it in advance was still right — it is why the outcome got a
second look at all.

**And the second look found the real problem.** Slot lengths against the numbers
I had queried an hour earlier:

```
slot                 before   after   delta
call-to-action         2658    2717     +59   <- one anchor
hero                   3171    3198     +27   <- one anchor
generic-text-block     4617    1980   -2637   <- 57% GONE
```

`content_data` 4439 → 1806. The item's summary is *"Add … tool reference to …
page"*; it regenerated the whole section, 7 paragraphs → 4, and changed the
heading. Lost: workpiece-first methodology, kinematic requirements (grip force at
contact vs actuator force, repeatability), cycle-time vs actuation trade-offs,
ISO 9409-1 flange integration parameters, closing synthesis. On a page titled
"Engineer's Reference". Filed `bugs_open/178`; **prior content is recoverable** in
`page_component_history` (`ecb4b420`, the 4439-char snapshot the writer itself
took at 10:41:22.93696 immediately before overwriting).

**NOT restored by me.** The naive restore removes the anchor the item was
legitimately raised to add, so the merge is an editorial act on a live customer
page — an owner call, and there is no urgency because nothing is lost.

Fleet check, 7 days, `content_data` shrink >25% vs last history snapshot:
`fundamentallyai /tools/review-council-simulator` -92% (3→3 slots),
`vonc /about` -50% (12→6), `relojistas /glosario` -48% (3→2),
`vetcomparison /about` -48% (4→4), `robot-hands` -43% (3→3),
`gamesdesign /tools/bayesian-ranking` -27% (4→4). **The slot column
discriminates**: vonc 12→6 is almost certainly `bugs_closed/156`'s legitimate
duplicate removal. The four with UNCHANGED slot counts match this signature; only
robot-hands is CONFIRMED (I read both versions), the rest are `[UNVERIFIED]`.

> **Checked, and it clears 154:** the witnessed `tool-improver` run that closed
> this bug is NOT an instance — `component_versions` for `c345a76a` shows
> 7944 → 9158 → 10520, i.e. it GREW. The morning's verification stands.

**Lesson, in WRONG_CALLS:** my prediction named only a failure mode, so `complete`
had no stated bar left to clear and read as vindication. The success criterion
needed writing down with the same specificity as the failure one — "length
unchanged apart from ~90 chars of anchor" — because that sentence, written in
advance, is what turns a green status into a question. I had applied exactly this
discipline to the 154 witness an hour earlier (hence the `component_versions`
check above) and did not carry it to the item I had personally unblocked.

---

## 2026-08-02 (afternoon) — restores applied (`287`), per-slot shrink guard shipped (`2da3e08e5`, INERT until roll)

Owner directed: restore all reduced pages + stop the class. Full dispositions in
`bugs_open/178`'s update block (single source; not duplicated here). Highlights
that belong to THIS lane's record:

- **Verification before restoring changed two dispositions.** gamesdesign's −27%
  was a legitimate rewrite (old blob = context dump, new fields purposeful) — NOT
  restored; restoring on the size signal alone would have REGRESSED a good page.
  fundamentallyai's live render was intact (32,876) with content_data NULL — the
  012 class — so it got a content_data restore and deliberately NO rerender.
- **The guard gap was measured, not assumed**: the existing content regression
  guard is page-TOTAL with a 25% wipe threshold; 178's case was −57% slot /
  −24% page. Two independent blindnesses. New per-slot floor (≥500 stripped
  chars, keep ≥50%, `section_shrink_floor` config, fail closed, refusal work
  item). Council `e64f8576` pending (~30 min queue); trailer Council-Submitted.
- **Guard is INERT until an image roll.** Post-roll pod-grep: added marker
  `"SECTION SHRINK"` ≥1, positive control `"CONTENT REGRESSION BLOCKED"` ≥1,
  both replicas, same exec.
- Council schema cost two rejected client-side attempts (plan must be an OBJECT
  `{summary, edits, grounded_in}`, grounded_in INSIDE plan) — cheap (no credits)
  but the same mistake family as 154's malformed edit path. Read the trigger's
  header before writing the JSON, not after the first error.

---

## 2026-08-03 — inducing the shrink-guard refusal (prediction recorded BEFORE the outcome)

Guard verified in the running binary FIRST: a roll to **v1.0.1234** landed minutes
before this session's check (pods 40–63s old at first look); pod-grep of both new
replicas: `SECTION SHRINK` 2, `section_shrink_floor` 1, control 1. The guard
survived the roll.

**Induction lever** (inverse of 165's plan-inflation, same logic): inflate the
STORED side. dartsonline `/blog/beginners.html` (`5009f5c8`, the fix loop's own
example site; passes all four gates from the 165 runbook, zero open items, no
open `save_refused_incomplete` for the site — dedup would swallow the emit).
Backed up all 3 slots (`tmp_induce_178_backup` + scratchpad), then appended
~10.8k stripped chars of `INDUCE-178-GUARD-PROOF` marker text to `article-body`'s
`rendered_html` (md5-guarded UPDATE). Existing side now 15,443 stripped; a
regeneration from intact content_data should produce ~4,644 → ratio ~0.30 < 0.5
floor. Page total 20,155 → ~9,356 incoming ≈ 0.46, above the page-total guard's
0.25, so the run REACHES the shrink guard.

**Prediction — success criteria written in advance, ALL must hold:**
1. Save REFUSED; step error contains `SECTION SHRINK REFUSED for page "beginners"`
   with `article-body 15443→~4644` and `floor 50%`. Read `__step_error`
   (bugfix 099: a FAILED step can show COMPLETED with `error` NULL).
2. Refusal work item EMITTED: `item_type='save_refused_incomplete'`,
   `item_key='save_refused_incomplete:beginners'`, `status='needs_human_review'`,
   spec.reason carrying the SECTION SHRINK text.
3. NOTHING written: all 3 `page_components` md5s unchanged (article-body = the
   inflated value), zero new `page_component_history` rows for the page, no
   deploy commit.
4. The induction `page_rerender` item ends terminal-failed, NOT complete.

A `complete` item or a changed row = the guard did not protect; STOP and
investigate before any cleanup. Blast radius (council answer 4, measured): **6
live agent types invoke `save_page_sections` as an action** — page-build-handler,
page-rerender, tool-recreation-handler (top-level step `save_sections`) +
pageflow-builder, page-rebuild, site-work-orchestrator (nested sub-workflow);
council-gate/diagnose-agent/fix-proposer only MENTION it in footprint maps.

### Outcome — ALL FOUR criteria met; the refusal is NOT masked on the measured path

The guard fired TWICE (the dispatcher retried the failed item once):
orchestrations `33fded0c` (22:56:00Z) and `085bcb22` (22:58:34Z), both status
**FAILED** at `save_sections`, `orchestration_states.error` carrying the full
refusal verbatim: `SECTION SHRINK REFUSED for page "beginners" — article-body
15443→4644 chars (30% kept, floor 50%)`. The 4644 is exact — the regeneration
reproduced the pre-inflation stripped length byte-deterministically, which is
itself evidence the lever measured what it claimed to.

1. ✓ Refusal with the predicted numbers, in `error` — NOT masked green on this
   path (`__step_error` was `(none)`; the failure surfaced in the real column).
2. ✓ Refusal item `ebc1dda8` `save_refused_incomplete:beginners`,
   `needs_human_review`, created 17ms into the FIRST refusal; the second emit
   collapsed by `idx_swi_dedup` — one open item per site+key, as designed.
3. ✓ Nothing written: all 3 slots md5+`updated_at` identical to baseline, **0**
   new `page_component_history` rows, `deploy_page` never reached. The
   article-body row still carried the marker — the refused save did not touch
   even the slot it objected to.
4. ✓ Item never `complete` (it looped `triaged`→claimed→failed; cancelled by me
   after evidence capture rather than left burning attempts).

Cleanup verified: slot restored byte-exact (md5 `82707c59…` re-checked), marker
grep across `page_components` = **0**, refusal item + induction item both
`cancelled`, `tmp_induce_178_backup` dropped (scratchpad copy retained).
Evidence files: `scratchpad/induce178/{orch_evidence,refusal_item_evidence}.jsonl`
— captured at the moment, since `orchestration_states` retention is ~24h.

### Round-2 code answer + resubmission (same correlation)

- **Locked-slot exclusion committed** (`5f00dcba9`): prior_art_librarian's round-1
  objection was CORRECT — for a locked slot the save discards the incoming copy
  (bugs_open/058), so comparing it against the locked existing is the sibling
  floor's false-refusal trap. Existing-side query now carries
  `pageComponentAgentWritableSQL("")`. Verified against `git archive HEAD` + the
  file — the in-tree build was broken by ANOTHER session's in-flight
  `load_work_item_actions.go` edit, exactly the shared-tree case the memory
  warns about; the archive method separated their breakage from my change.
- **Resubmitted 2026-08-03 with `RESUBMIT_CORR=e64f8576-…`**:
  `SUBMISSION_2026-08-03_shrink_guard_round2.json` — all 9 round-1 objections
  answered (HIGH: the induction; blast radius: 6 invoking agent types measured,
  3 top-level + 3 nested, the footprint-map mentions separated out; pod-grep:
  run twice, 1233 and 1234). Verdict watch armed.

A roll to **v1.0.1234** landed mid-session (pods 40–63s old when first checked);
guard markers re-verified on both new replicas before inducing. The ~300s
no-dispatch window after a chassis restart was waited out before queueing.

### Found by the induction, not by any reviewer: the refusal item's SUMMARY is the wrong sentence

`ebc1dda8`'s queue-visible summary reads *"beginners" returned too few sections
to replace what is stored* — `savePageSectionsRefusal` hard-codes the
completeness floor's wording, and the shrink guard reuses the whole helper. For
a shrink refusal that sentence is FALSE (all sections were returned; one was
too small), and an operator triaging the queue is pointed at a count problem
that does not exist. `spec.reason` carries the correct text, so the detail is
one click away, but the one-liner is the thing the queue shows. Nine council
seats across two rounds did not catch this; watching the real item land did —
which is the induction argument in miniature.

**Deliberately NOT fixed mid-round** (round 2 is in council now; a fix commit
the round cannot see helps nobody). The fix is small but has one design edge:
keep `item_type`+`item_key` shared (one open refusal per page is the right
dedup) and parameterise ONLY the Summary/Fix strings. Fold into a REVISE if one
comes back; otherwise a follow-up commit after the verdict.

---

## 2026-08-03 — v1.0.1238 rolled; every commit on this thread now LIVE and pod-proven

Owner rolled a fresh build (v1.0.1238, tags 1235–1237 skipped like 1230–1232).
Pod-grep BOTH running replicas, one exec each:

```
shrank past the floor                              1   (77b58fd4d — violation wording)
could not measure the page's existing sections     1   (0913d5754 — measurement-error sentence)
SECTION SHRINK                                     2   (guard marker, unchanged)
CONTENT REGRESSION BLOCKED                         1   (positive control)
INDUCE-178-NEVER-IN-BINARY                         0   (nonsense negative control — grep sane)
```

`5f00dcba9` (locked-slot exclusion) adds no rodata; proven in by ANCESTRY — the
0913d5754 literal postdates it and builds come from committed HEAD. Counts are
substring `grep -c` used as PRESENCE checks (≥1), not occurrence counts —
debug_historian's caution about linker-packed literals applies to counting, not
to presence.

**This thread is closed end to end**: guard live + induction-proven + APPROVED
(e64f8576 r2); refusal wording true on all four paths + APPROVED (98aa9103) +
its advisory implemented; all of it in the running binary. What 178 still owes
is the handler root cause and the sibling writers — a different thread.

---

## 2026-08-03 — 178 root cause: 090 diagnosis DISPATCHED

Handoff item 1 picked up. Pre-dispatch checks, in order:

- Dedup: no open `needs_diagnosis` item covers this mechanism (the four open
  rows are 156-duplicates / anthropic max_tokens / phantom links / spawn race —
  all sitting `failed`, incidentally). No `/bugs_open/` file other than 178.
- Refusal queue: only the known dartsonline induction row (cancelled) + one
  older prune-floor item. **No new class instances since the guard went live.**
- **The handoff's suggested symptom said `tool_crosslink` — the real
  `item_type` is `content_rewrite`** (verified on row 93f2a3b7; handler
  `page-build-handler`, produced by `create_tool_cross_link_items.go`). Symptom
  authored with the real type.
- Origin was **406 commits behind** local HEAD and the diagnosis reads origin —
  `save_sections_shrink_guard.go` did not exist there at all. Fast-forward
  pushed before dispatch (3e11f2518..30dde02d1); a session committed again
  mid-push, leaving a 1-commit advisory that the seam does not depend on.
- Coverage probe hits: two `needs_links` rows on the target page, `unresolved`
  since 07-17, completeness-discovery-agent — parked backlog, not in-flight.
  Read, then FORCE=1 per the documented path. Seed-scope probe clear.

Dispatched via the loop (not a direct publish):
`item_key = needs_diagnosis:178-crosslink-regenerates-whole-section`,
intake corr `0c4b57be`, **RUN corr `aece2920-f85a-46e2-a53f-235a4b6e9ab1`**
(artifacts key). Seeds: `create_tool_cross_link_items.go` +
`save_page_sections_action.go`; subject `pipeline/page-build-handler`;
runtime site robot-hands.com. Verdict pending — capture evidence promptly,
`orchestration_states` retention is ~24h.

---

## 2026-08-03 — 177: dispatched a 090 into a lane that had JUST been taken (misstep, recorded)

While 178's run iterated, took handoff item 3 (177, listed "unstarted"). Quick
look verified the two emit sites, `sections=[]` on the traced page, and that
`content_guidance` is written four times and read by nobody on the work-item
path. Dispatched 090: intake `2e566eb2`, RUN corr
`da59941f-8d16-4c3a-9812-e9f76064de28`.

**who-owns run AFTER dispatch surfaced `bugfix_177_tool_content_items/`** —
PLAN created 11:41 BST, ~4 minutes before my dispatch, untracked, so invisible
to who-owns (reads commits) and to the trigger's probes (reads
`site_work_items`). The owning lane is further along than my quick look: the
deploy path DECLARES four sections (`deploy_tool_action.go:343-346`), every
dead item is create-path, and the plan-sections edge case is measured. **My
symptom's both-paths framing is wrong on the deploy path** — caveat recorded
in `bugs_open/177` so the verdict is read with the right lens.

Run left to complete deliberately: the 07-31 ruling wants a loop pass over a
first-hand-verified structural root cause (155 precedent), and this is that
pass for their PLAN's claim. NOT fixing 177 here — contribution routed into
the bug file per who-owns. The check the misstep earns (transcript grep BEFORE
a 090 dispatch, not after; three lagging surfaces are not one live one) is in
`WRONG_CALLS.md` 2026-08-03.

---

## 2026-08-03 — handoff item 4 CLOSED: the relojistas slot was a recorded back-out, not a mystery

Traced the writer instead of the row. `page_component_history` has no
slot_name column (pointer is `component_id`, FK `ON DELETE SET NULL` — hence
the anonymous snapshot). Zero `page_components` rows fleet-wide hold a
`jsonld` content key today; no Go code writes `DefinedTermSet`. The trail
ended in `traffic_probe/relojistas_rebuild_running_notes.md` 2026-07-28 (3):
the `structured-data-block` component (JSON-LD DefinedTermSet, 8 glossary
terms) was built AND **deliberately backed out the same day** — a JSON-LD-only
section is a `<script>` and nothing else, `sectionHasVisibleContent()` drops
it by design, so it can never reach the served page. Verified live: component
`b51dbc8f` `is_active=false`, refusal reason in its description. Snapshot
`b0e119a4` is the back-out's own pre-write copy (source
`save_page_sections_overwrite`, 07-28 16:21Z).

**Do not restore; no owner call needed.** Corrections written into
`bugs_open/178` (three visible markers + a dated update) and the handoff.
The 178 fleet-signature table loses its relojistas row: 3→2 slots was the
back-out.

---

## 2026-08-03 — 177 verdict: REFUTED in one iteration, re-deriving the owning lane's asymmetry

The run took ~3 minutes (claimed 10:43Z, complete 10:46:34Z — far under the
~30-minute budget; the loop dispatch really is faster than a direct publish).
Outcome REFUTED: the loop caught the symptom's both-paths framing on
`deploy_tool_action.go`'s explicit four-section `sectionsJSON` and noted the
create path's INSERT omits `sections` entirely — i.e. it independently
re-derived the 177 lane's central finding from seed code alone, then stopped
honestly (scope-not-narrowing, UNVERIFIABLE, no auto-conclusion). Exactly the
"REFUTED verdict is a success" case: my framing was the wrong part, and one
run caught it. Verdict JSON committed as
`EVIDENCE_2026-08-03_177_verdict_da59941f.json` (this dir); full trail written
into `bugs_open/177` for the owning lane.

---

## 2026-08-03 — watch item MEASURED: the pre_query/selector asymmetry is real in code, vacuous in data

The 08-02 `[UNMEASURED]` marker (scheduler pre_query gates `status='triaged'`
+ `locked_at IS NULL`; loader accepts `triaged|approved` at
`load_work_item_actions.go:633` with the approval_mode gate at :635, and
skips locked sites itself at :127-137). Measured live:

- Dispatchable build items right now: finetuning.uk 114 triaged /
  vetcomparison.uk 2 triaged, both unlocked; **zero `approved` items on any
  site**. No one is in the starvation gap today.
- Deeper: **no row in the retained table has EVER had `status='approved'` or
  a non-auto `approval_mode`** (0 rows, any pipeline, any status). The
  selector's approved arm has never matched anything — a designed-but-unexercised
  approval flow, not a live path. (site_work_items retains at least back to
  07-14; this is not an orchestration-style 24h window.)

So: the asymmetry bites only when the first writer sets `approved` — at which
point an approved-only queue starves invisibly whenever no site fleet-wide
holds a `triaged` item (the pre_query returns 0 sites and the loop never
ticks). Noting, still not fixing: the right moment is whenever an approval
flow is actually built, and the fix belongs with that work. Marker resolved
to `[MEASURED 2026-08-03: inert]` — not deleted, downgraded.

---

## 2026-08-03 — 178 root cause: CONFIRMED, closing the gap the 090 run named but couldn't reach

Run `aece2920` ran 5 full iterations (10:39Z→11:00Z) and stopped honestly at
UNVERIFIABLE. Its own "still needed" text named the exact gap: *"the
page-build-handler writer/content-generation step definition (absent from
this bundle)"* — a scope/tooling limit (the step config wasn't in what got
gathered), not evidence against the hypothesis. Read that config directly
from live `agent_definitions` after the run ended:

- `load_existing_content` step: `mode: input_data.spec.mode`. The action
  (`load_existing_content_action.go:64-69`) is a hard gate:
  `if mode != "recreate" { return has_existing:false }`. Doc comment: *"For
  non-adoption pages (no mode: recreate), returns empty — no-op."*
- `93f2a3b7.spec` has no `mode` key (checked the live row).
  `create_tool_cross_link_items.go` never sets one (grepped, zero hits).
- `call_content_writer`'s `input_mapping`: `existing_content?` ← that no-op;
  `current_page` ← `page_record`, whose own step description says it carries
  only "sections, title, page_type" — confirmed no prose channel exists
  there either (`page_components.content_data` is never loaded into it).
- So the writer gets `rewrite_guidance` (the item's instruction text) and
  literally nothing to edit. It fabricates a replacement section that
  satisfies the instruction's shape — exactly the observed defect: correct
  new link, shorter/restructured prose, changed heading.
- **Bonus finding, corrects fix candidate 2's premise**: `mode="recreate"`,
  even set, sources `research_results` (the ORIGINAL adoption crawl) per
  `load_existing_content_action.go`'s own doc comment — never current
  `page_components`. So it can't be the fix; there is no existing channel
  that passes a page's LIVE stored content to its writer for editing.
  Candidate 1 (edit-not-regenerate) is the only one the plumbing supports
  without adding a new channel.
- **Generalises past cross-links**: the mechanism is "emitter omits `mode`",
  not anything cross-link-specific. `apply_gap_plan_action.go`'s
  `content_rewrite` emission (:243) was not checked for a `mode` key — left
  as a flag for the fix, not verified either way.
- This is the diagnosis-loop escape hatch (07-31 ruling) used as designed:
  the automated run was executed, iterated fully, and named the precise
  missing evidence; I then supplied that evidence first-hand rather than
  re-running a 6th iteration, and say so here plainly.
- Landmine written (shared mechanism, will bite the next thread touching any
  content-writer path) + verifier dispatched (`98ca06a2`). Evidence JSON for
  both runs (`aece2920` here, `da59941f` for 177 above) committed before the
  ~24h `orchestration_states` reaper.
- **This closes ROOT CAUSE, not the bug.** Fix candidates 1 (edit channel)
  and 3 (emit the delta) are unimplemented; the sibling-writers item and the
  fourth-floor tracked deferral are unchanged. Full write-up in
  `bugs_open/178`'s final update.

---

## 2026-08-03 ~20:15 — fresh chassis build (v1.0.1243) verified; 177 confirmed live, done by its owning lane

Owner reported a fresh build. Verified at the pod, not the tag: deploy image
`v1.0.1243` (was 1238 at this lane's last check), `raiseToolContentItem`
count 7 + `recurrenceExpected` count 1 on both replicas — the 177 lane's fix
(`74655b709`, council `982507b0`), not anything from this session (this
session shipped no code, only diagnosis + docs).

`git log --since` on the four files this lane's 178 diagnosis names
(`load_existing_content_action.go`, `save_page_sections_action.go`,
`create_tool_cross_link_items.go`, `apply_gap_plan_action.go`) shows nothing
since the diagnosis completed — 178's mechanism is exactly where this
session left it. Their commit message names why they left the two blocked
`content_rewrite` dependents blocked: *"the 178 interlock — dependents stay
blocked, their dispatch is destructive today"* — a direct, correct
consumption of this session's 178 root-cause finding by an independent lane.

HANDOFF rewritten (rev 2) to make the 178 FIX the explicit next task, with a
full design direction (opt-in field per the 2026-08-02 owner ruling on
shared-seam authority, not a blanket behaviour change to
`call_content_writer`) so a fresh session can start writing code without
re-deriving the mechanism. Given this session's length (two full 090
diagnosis runs, extensive DB/code reads, multiple large captures), the fix
implementation itself is being left for a new session/context rather than
started here.

---

## 2026-08-03/04 — 178 FIX IMPLEMENTED, committed, built, LIVE, migration applied; live verification IN PROGRESS

Fresh session (post-`/clear`), picked up exactly where the rev-2 handoff left
off. Design decided by reading the LIVE `page-build-handler` and
`page-content-writer` workflows directly (not the historical
`sql_for_agents` migration files, which are historical and drift from
current state — confirmed `load_existing_content`'s config is
`"mode": "input_data.spec.mode"`, matching the diagnosis exactly) rather than
re-deriving from docs alone.

**Design chosen, and why the alternatives were rejected:**
- A third `spec.mode` value, `"edit_live"`, alongside the existing
  `"recreate"` — reuses the field the diagnosis already names rather than
  inventing a new one.
- New Go action `load_current_section_content_action.go`, a NEW workflow
  step (not a change to `load_existing_content_action.go`, which stays
  adoption-only per its own doc comment) inserted between
  `check_has_ready_sections` and `spawn_content_writer`. It reuses
  `plan_sections`' own `section_plan` OUTPUT FIELD NAME — same key,
  overwritten — specifically so `call_content_writer`'s existing
  `input_mapping` needs zero changes. Confirmed live: that mapping already
  carries `"section_plan": "section_plan"` and `"existing_content?":
  "existing_content"` before this session touched anything.
- Considered and rejected: threading a per-slot map through the template via
  Go's `index` builtin (`RenderPromptTemplate` uses plain `text/template`, so
  `index` IS available — verified in `data_helpers.go:1129-1152` — but doing
  the join ahead of time in Go and handing each section its OWN
  `existing_content_html` via the loop variable is simpler and needs no
  template-engine gymnastics).
- Content source: `page_components.rendered_html`, not `content_data`.
  Guaranteed prose-complete and needs no per-component schema knowledge;
  `content_data`'s shape varies by component and isn't guaranteed to be
  flowing text. Recorded as an OPEN REVIEW QUESTION in the register entry
  (PBP-028) rather than resolved — unverified whether markup-vs-field-shape
  costs the model any fidelity.
- Matching key: `slot_name` (page_components) == `sectionPlanItem.Name`
  (`plan_sections_action.go`) == `component_function` (what
  `save_page_sections_action.go` writes AS slot_name). Confirmed by reading
  all three, not assumed.

**Both live emitters updated** (the ONLY two `content_rewrite` producers,
confirmed by search): `create_tool_cross_link_items.go`'s
`emitToolCrossLinkItems` and `apply_gap_plan_action.go`'s `applyAddToPage`
(read in full — it looks the page up BY NAME from `pages` before building
the spec, so bugs_open/178's flagged-but-unchecked `:243` emission does
target an existing page, same as the cross-link path).

**Tests**: 4 new sqlmock tests, the load-bearing two being NEGATIVE proofs
(no `mode` set; `mode="recreate"`) — `mock.ExpectationsWereMet()` on a mock
with NO query expectations registered, so a stray query fails the test, not
just a wrong return value. `go build ./...` and the whole `actions` package
suite green (one unrelated pre-existing broken file in
`discovery_checks/` — another session's WIP, confirmed via `git log` on that
file predating this session, left untouched).

**Committed** `08d0515f3` (code+SQL+register, 8 files, scope report clean —
no passengers) then `0a2a94b89` (IMAGE_TAG bump, separate per the
one-commit-per-task rule). **Submitted to council**: `Council-Submitted:
97ebadcf-bbe6-485f-8231-ff16fc4e679f`. ⚠ **Checked 2026-08-04 09:xx: this
run STALLED** — `orchestration_state_audit` shows it reached
`review_constitution` at 20:09:59Z on 08-03 and never advanced again; no
`council_report` diagnosis_artifact was ever written, and the
`orchestration_states` row itself is gone (reaped or never-written; audit
trail is the only surviving evidence). Not investigated further — council
review is advisory only and this doesn't block anything, but flagging so
nobody waits on a verdict that isn't coming. Whoever next touches
council-gate reliability: this is one more data point for that lane, not
this one's to fix.

**Built + pushed**: `v1.0.1244`, from committed HEAD (`0a2a94b89`).
**Did NOT self-deploy** — a same-day memory note records the owner wants
whole-fleet releases run by him personally (`make release redeploy-agents`),
after a previous single-service deploy fragmented the fleet's tags. Asked
the user to run it and stopped there.

**Owner ran the release.** Live tag confirmed `v1.0.1247` (three more builds
landed between 1244 and the release — expected on a shared tree). Pod-grepped
BOTH replicas before touching anything further: `LoadCurrentSectionContent:
attached current content for edit mode` count 1, `load_current_section_content`
count 2, both pods — genuinely live, not just the tag. Re-verified the
migration's anchor text was still untouched (`check_has_ready_sections.then_step`
still `spawn_content_writer`, prompt anchor still occurs exactly once) before
applying — nobody else had touched either definition in the interim.

**Migration 299 applied** by hand (`psql -v ON_ERROR_STOP=1 < ...`, both
`DO $$ ... RAISE EXCEPTION $$` verify blocks passed, real assertions not bare
SELECTs — see the migration-runner landmine on why that distinction matters).
Recorded via `run-migrations.sh --record-only` rather than a hand-written
ledger row.

**Live verification, using REAL parked production items rather than a
synthetic test.** `bugs_open/178`'s own note names two crosslink items
(`9e9ec430`, `18bc832c`, both on `vetcomparison.uk`, targeting
`guide-cma-compliance` and `guide-independent-strategy`) that the 177 lane
deliberately parked behind a dead `wont_fix` dependency specifically because
dispatching them pre-178-fix would be destructive. **Caught before acting**:
those two items were created 2026-08-02, BEFORE this fix's emitter change, so
their `spec` carries no `mode` key — releasing them as-is would still hit
the OLD path and reproduce the exact damage they were parked to avoid. Fixed
by `UPDATE site_work_items SET spec = spec || '{"mode":"edit_live"}'::jsonb,
depends_on = NULL WHERE id IN (...) AND status='triaged'` in one transaction,
then confirmed both rows show `mode=edit_live` and `depends_on` empty.

Baseline recorded before dispatch: `d8c51ace-...` (guide-cma-compliance)
`generic-text-block` content_data length **6034** (rendered_html 6212);
`2a347990-...` (guide-independent-strategy) `generic-text-block` **3637**
(rendered_html 3815).

**Discovered no scheduler drives `build` pipeline dispatch at all** —
`scheduled_tasks` has `report-dispatch` and `diagnose-pipeline-trigger`, no
`build-dispatch-loop` entry, ever. Same family as the
`detection-works-schedule-and-dispatch-do-not` pattern already in memory,
now with a THIRD instance. Fired `build-dispatch-loop` directly via `kcat`
(`config.agent_type: "build-dispatch-loop"`, `input_data: {site_id, domain}`
— confirmed its `input_contract` first rather than guessing the shape) for
`vetcomparison.uk`'s site_id, since both target items are on that site.
**IN PROGRESS as this note is written**: orchestration claimed both items,
is calling `page-build-handler` for item 0
(`process_item_iter_0_call_handler`, `AWAITING_RESPONSES`). Result not yet
known — continues below or in the next handoff, whichever this session
reaches first.

---

## 2026-08-04 — dispatch RESULT: the fix is proven correct at the data level; end-to-end blocked by an UNRELATED bug, filed as 192

Both items claimed and dispatched. **9e9ec430** (guide-cma-compliance) ran
`page-content-writer` orchestration `0883b1aa-d5d6-45ad-a596-df0cc06744ec`.
**18bc832c** never got processed by this same run — the dispatch loop
claimed both but appears to only fully process one item per invocation
(iter_0); the second stayed `claimed` until I fired a second read below.
Not investigated further — a re-fire will pick it up, and it isn't this
fix's mechanism.

**The proof that matters**: `0883b1aa`'s `collected_data.input_data.section_plan`
(read directly from `orchestration_states`, not inferred) shows
`sections_ready[0].existing_content_html` holding the page's full, exact,
CURRENT `generic-text-block` HTML — the real CMA-compliance prose, matched
by slot name (`generic-text-block`), attached read-only, nothing
paraphrased or truncated. **This is exactly what `load_current_section_content`
was built to do, and it did it correctly on a real production page on the
first live dispatch.**

**Then it failed** — `process_sections_loop` inside `page-content-writer`,
same step, `sections_for_render.sections_ready not found`. Traced far enough
to be confident this is NOT this fix: (1) the exact same failure hit
`df69efd6-19b7-4788-8fe1-668ea769f3fc` in the same few minutes, for
`tool-gripper-payload-calculator-guide` on an entirely different site,
via `needs_content_page`/adoption — no relation to `edit_live` or a
crosslink at all; (2) `orchestration_states` history shows this exact
signature spiking at 2026-08-03 21:00-23:00 (11, 14, 12 failures those
hours) — **hours before this fix's own image (`v1.0.1244`, pushed ~20:2x)
had been deployed anywhere** (the owner's release, and this fix's actual
rollout, happened the following morning — pods were ~11 min old when
checked ~08:2x on 08-04). A regression that predates your own code cannot
be caused by your own code. **Did NOT chase the real root cause** — read
`ExtractFieldsAction` (`v3_site_actions.go:4232`) far enough to see it DOES
null-check candidates (so a naive "extract_fields doesn't skip nulls"
theory is wrong), confirmed `input_data.section_plan.sections_ready` DOES
hold the real data in the SAME collected_data row at the SAME time, and
stopped there rather than spending more of this session chasing a second
bug — filed as `bugs_open/192` with everything gathered, a `090` run
flagged as owed, not run.

Also incidentally confirmed: `bugs_open/087`'s own text ("page-build-handler
... its writer children always carry a real section_plan, 26 of 26 COMPLETED")
was true when written and is **not** the same failure as 192 despite the
identical error string — 087 is `page-rebuild`-specific (no section_plan
supplied at all); 192 hits the build-handler path 087 called the healthy
control. Two different causes, same symptom string — logged so nobody
conflates them.

**Both work items are safe**: `failed`, `attempt_count=1`, `depends_on`
cleared, no content touched (failure is upstream of any save). Re-dispatch
once 192 is fixed; do not re-try before then, it will hit the same wall.
Recovery of these two rows and the fresh `092`-adjacent one is unnecessary —
nothing was written.

**178's own status**: fix code DONE, LIVE (v1.0.1247, pod-verified both
replicas), migration applied+recorded, register entry PBP-028 written,
council submitted (that run stalled — see below — advisory only, doesn't
block). The ONE thing not yet directly observed is the writer's OWN
behaviour with `existing_content_html` in hand (does it actually edit rather
than replace) — logically it should, since the prompt block instructs
exactly that and the field is populated correctly, but the assertion in
178's "how to verify a fix" section (before/after `content_data` length) has
not been run to completion. That is the one thing 192 is blocking.

**Council run status, checked 2026-08-04**: `orchestration_state_audit` for
`6837e104-5924-4833-a526-eeb6f58a1f65` (submission `97ebadcf`) shows it
reached `review_constitution` at 20:09:59Z on 08-03 and never advanced —
no `council_report` artifact was ever written and the `orchestration_states`
row itself is gone. Stalled, not rejected. Not investigated (advisory only,
doesn't block); flagging for whoever next looks at council-gate reliability.
