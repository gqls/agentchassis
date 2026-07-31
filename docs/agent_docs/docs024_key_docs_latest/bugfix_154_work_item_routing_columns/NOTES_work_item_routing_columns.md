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
