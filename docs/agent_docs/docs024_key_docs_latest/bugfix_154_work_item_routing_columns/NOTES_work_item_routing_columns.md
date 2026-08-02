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
