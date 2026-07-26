# NOTES — bugfix_006

Running technical record. **Append-only, newest at the bottom.** What was tried, what the system
actually said, and every misstep — the missteps are the point, not an appendix.

Started late: this file was created 2026-07-26, when §C was fixed. Entries before that date are
reconstructed from the bug file and the commit trail and are marked as such.

---

## 2026-07-20 → 2026-07-25 (RECONSTRUCTED from `bugs_closed/006` §B and the commits)

Not written contemporaneously; recorded here so the file has a spine. The bug file is the
authoritative account for this period.

- **The filed cause of §B was STALE and survived four days.** `006` said generated contact forms
  POST to a dead `/contact` endpoint, citing `k8s/bk_page_components.sql:140`. Measured live:
  **zero** components had `action="/contact"`. **The misstep:** a `bk_*.sql` file is a **backup
  DUMP of a table, not source**. The real defect was that `form_action` comes from per-component
  `content_data` written by the content LLM and **no Go code ever set or validated it**. Logged as
  `WRONG_CALLS.md` #9. *The cheap check that would have caught it:* grep the Go tree for the field
  before citing any `bk_` path as an emitter — if nothing sets it, the value is **content**, and
  the fix is a default plus validation, not a template edit.
- **A base-map default would have HALF-worked, and looked finished.** `ContentData` merges **over**
  the defaults in `contextToInterfaceMap`, so a default beside `cta_url` would have repaired the 3
  empty cases and left the 8 `#contact` sites broken *while reading as fixed*. Hence post-merge
  sanitisation, pinned by `TestFormActionSurvivesContentDataMerge`.
- **The fix covered one of two render branches.** The pre-commit pattern check flagged
  `contextToInterfaceMap` changed without its twin `contextToMap` (016b §9 #26) — correctly:
  `RenderTemplate` falls back to regex substitution when Go templating errors, and that path
  merges `ContentData` too.
- **I reported another thread's deliberate decision as my discovery.** idea.uk's staged mailto was
  written up as an unnoticed instance of the "content_data edits don't hold until re-render"
  landmine. It was not — the idea.uk thread had applied it source-only on purpose and documented
  why four days earlier. *Caught by:* searching the docs before building the fix, at the owner's
  prompting.
- **The council's REVISE named a REAL defect, not a documentation gap.** The "refuses to fabricate
  an address" guarantee held only where `ctx.Email` is empty, but two render paths synthesise
  `ctx.Email = "info@" + Domain` **before** rendering. Fixed at the chokepoint (`efe634b37`).
- **Council schema traps, paid for twice.** Two runs died `complete_invalid` at
  `persist_submission`: `fixPlan.risks` must be a **string** (ours was an array), and `operation`
  `"create"` is not allowlisted (a new file is `add`). **An invalid run writes NO artifacts**, so
  polling `diagnosis_artifacts` alone waits forever — poll `orchestration_states.current_step` too.

---

## 2026-07-26 — §C, and the closure of the case

### Ownership and collision checks, before touching anything

- `scripts/who-owns.py 006` → owning workstream `docs024_key_docs_latest/bugfix_006` (this lane),
  7 commits in 14 days. No competing owner.
- `bugfix_029_dispatch_gate` opened **today** and works the dispatch chain
  (`build-pipeline-trigger.pre_query`, `find_dispatchable_site`, `load_work_items`, migrations
  213/214). Grepped its PLAN: it only **reads** `claimed-item-timeout` as a reference point
  ("*Why 45 minutes and not 30: `claimed-item-timeout` resets orphaned claims at 40 min*"). Not a
  collision — but worth knowing that two threads were in the same neighbourhood on the same day.
- `needs_diagnosis` queue: 0 rows `awaiting_diagnosis`.

### A — checked first, and it had resolved itself

```
$ kubectl -n ai-persona-system get pods -o wide | grep runner
github-actions-runner-5c44ddb44d-5pqdv          1/1 Running 0 28d   prod-instance-…4001149
github-actions-runner-5c44ddb44d-6f66p          1/1 Running 0 5d6h  prod-instance-…8031336
github-actions-runner-vmsites-5bf4b47c57-zdtjb  1/1 Running 0 9d    prod-instance-…4001149
$ kubectl -n ai-persona-system get pods | grep -c CrashLoopBackOff  →  0
```
The crash-looper (`-lhg9l`, 6365 restarts) is gone. **What I could not establish:** whether
someone fixed the node's containerd `SystemdCgroup` or the node pool simply rolled. The bad node
is no longer in the cluster, so those are indistinguishable now. Marked `[INFERRED]` in the bug
file with an explicit reopen trigger rather than written up as "fixed".

### C — the hypothesis I was most confident about, and it was wrong

Reading `stale-orchestration-reaper`'s `pre_query` I found:

```sql
UPDATE orchestration_states SET status='FAILED', error='reaper: dispatch loop idle for >30 min'
WHERE status='AWAITING_RESPONSES' AND owner_agent_type='build-dispatch-loop'
  AND last_activity < NOW() - INTERVAL '30 minutes'
```

and the dispatch loop's `call_handler` has `timeout_seconds: 1200` (20 min). The story wrote
itself: the reaper kills the **supervisor** while the **handler** is still working, so
`mark_complete` never runs, the claim orphans, and 40 minutes later the item re-runs. It
explained the symptom, the 5-items-reset-in-the-same-second batches, and why slow AI-heavy types
(`needs_page`) never auto-complete while fast ones (`page_rerender`) sometimes do.

**REFUTED by measurement, before I wrote it down as a cause:**

```
owner_agent_type      n    avg_min  max_min  over_30min
page-rerender        211      0.4      3.6        0
page-build-handler    17      4.9      8.1        0
image-build-handler   12      1.6      1.9        0
```

No handler orchestration has ever approached 30 minutes. And a live dispatch loop's
`last_activity` **advances on each loop iteration** (observed: created 17:47:59, `last_activity`
17:54:29 at `iter_1`), so it does not idle out under normal flow. The reaper is not implicated.

**Why this is worth recording:** the hypothesis was coherent, mechanism-shaped, and fitted every
observed symptom. Confidence was not a signal. *The cheap check that killed it* — one `avg`/`max`
over completed handler orchestrations — took under a minute and should have come **before** the
theory felt good, not after.

### C — the link, and the measurement trap I nearly fell into

The generic evidence source is `initial_request_data->'input_data'->>'work_item_id'` on the
handler's own orchestration. First coverage measurement, over a **36-hour** window:

```
page_rerender  298 items,  212 linked   (71%)
needs_page      32 items,   14 linked   (44%)
needs_imagery   23 items,   12 linked   (52%)
```

That reads as a coverage hole and nearly became one in the plan. It is not. **`orchestration_states`
is purged at ~2 days** — 1,125 rows for 07-26, 694 for 07-25, ≤8/day before that. The unlinked
items were older than retention. Re-measured over **6 hours**:

```
page_rerender  92/92 · needs_page 3/3 · needs_content_page 2/2 · nav_drift 1/1 · missing_*_page 1/1
```

**100/100.** The item types that link to nothing even in the short window
(`needs_section_data` 0/7, `cta_names_unknown_destination`, `needs_experience_plan`) are completed
by paths that **never claim**, so they can never be in the `claimed` status the sweep reads —
confirmed by `SELECT type FROM agent_definitions WHERE … default_config::text LIKE
'%claim_work_item%'` returning exactly two agents, the second of which uses status `diagnosing`
and says so itself: *"We own this because diagnosing is inert to claimed-item-timeout."*

**Lesson:** pick a measurement window **shorter than the retention of the thing you are
measuring**, or you will diagnose a purge as a defect.

### C — a guard I wrote that would not have caught anything

First draft of migration 220's guard 3 planted the probes and then re-stated the EXISTS predicate
inline to check them. **That is the vacuous-check trap in this repo's own memory** (*"a check
sharing the fix's regex can't falsify it"*): a typo in the stored `pre_query` would have left the
guard passing happily while the sweep never fired in production.

Rewritten to read `pre_query` back out of the column and `EXECUTE` it — the real string, exactly
as the scheduler runs it — then assert on the resulting row **status and marker**. Caught by
re-reading my own guard against the landmine list, not by anything failing.

### C — fault injection, all four assertions

Every guard was watched to fail with the right diagnostic before being trusted:

| induced fault | what fired |
|---|---|
| `AND EXISTS` → `AND NOT EXISTS` | *"positive probe was not completed (status claimed) — the generic evidence branch never fired"* |
| `o.status = 'COMPLETED'` → `IN ('COMPLETED','FAILED')` | *"negative probe changed status to complete — the sweep acts on a claim whose handler FAILED"* |
| `RETURNING` → `RETURNINGG` | guard 2 `PREPARE`: `syntax error at or near "RETURNINGG"` |
| new branch writes the OLD marker | *"positive probe completed with the wrong marker … an OLD branch matched, so this migration proves nothing"* |

And the Go lockstep test's three: verifier dropped from the SQL list → *"HAS a Go verifier but is
NOT excluded"*; bogus type added → *"is excluded … but has NO verifier"*; migration file hidden →
*"expected exactly 1 … found 0"*.

### C — proven through the running scheduler, not through psql

The migration guard proves the SQL is right when psql executes it. That is deployment of a
string, not proof the fleet's scheduler runs it. So: two probes planted live at `status='claimed'`,
`claimed_at = now()-20min`, `item_type='migration_probe'` (no artifact branch can match it, so a
completion could only come from the new branch) — one with a `COMPLETED` handler orchestration,
one with a `FAILED` one. After one tick:

| probe | result |
|---|---|
| handler `COMPLETED` | `complete`, `completed_at` stamped, `error='Auto-completed: handler orchestration completed after claim'` |
| handler `FAILED` | still `claimed`, untouched |

Both probe pairs deleted afterwards; 0 remaining, verified. Scheduler log across the window:
repeated `"Pre-query task completed (no message fired)","task":"claimed-item-timeout"`, no errors.

### The shared tree did not compile, again

`go test ./platform/orchestration/actions/discovery_checks/` failed with `undefined:
hardcodedColourCandidate` — another session's in-flight edit to
`check_hardcoded_section_colors.go` against its own unmodified test. **Not mine to fix.** Tested
via `git archive HEAD | tar -x` into a clean dir with only my file overlaid (plus the migration,
which the test globs for). All four tests in the file pass there.

### My staged `git mv` was swept into another session's commit

Between `git mv bugs_open/006… bugs_closed/006…` and my `git commit`, another session committed
`61ba74233` (an 084 fix) which carried **my staged deletion of `bugs_open/006`**. The pathspec
commit then failed: *"pathspec 'bugs_open/006…' did not match any file(s) known to git"*.

Nothing was lost — forward-only held, and the remainder (the file at its new path plus the two
indexes) committed cleanly as `d5988a8ed`, which says so in its message. **This is the hazard
CLAUDE.md describes actually happening**: committing per task narrowly stops *me* sweeping
*others*' work; it does not stop a session running `git add -A` from sweeping mine. The practical
consequence for a `git mv`: the two halves can be separated by another thread, so check
`git log -- <old path>` before assuming your move failed.

---

## 2026-07-26 later — the chassis rolled, and one of my two claims did not survive contact

### 220 survived the roll, and the NEW binary runs it

A chassis roll is the moment a config-only fix can be silently undone, because a deploy can
re-seed DB config. Checked within two minutes of `agent-chassis-5b4456686c-s5fkc` coming up:
`pre_query` still carries all three markers and is **6131 bytes — byte-identical** to what
migration 220 wrote.

The stronger check is that the **new scheduler binary** runs it, not just that the row survived:
`kafka-scheduler-69c76d58fb-6ggj7` logged `"Pre-query task completed (no message fired)"` for
`claimed-item-timeout` at 21:05:42Z from **`scheduler/main.go:272`** — the pre-roll pod emitted the
identical message from **`:238`**. Different binary, same row, still working. *The line number is
the discriminator*; the message text alone would have proved nothing about which build produced it.

### MISSTEP — I reported the council submission as queued. It had been dropped.

At 18:44Z my submission had no `orchestration_states` row. The council runbook has a standing rule
for exactly that shape: *"a missing orchestration row is almost always latency, not a dropped
dispatch — do not retry on that evidence (it costs a duplicate round)."* I applied it and told the
owner "still QUEUED — latency, not a drop."

**Wrong.** At 21:05Z, 2.5 hours later: another thread's submission `569241fb`, published *after*
mine, had run at 20:05Z, re-run at 20:16Z and reached `complete_revise`. The lane had drained past
my slot. Searching my correlation across `orchestration_states.collected_data`,
`initial_request_data` and `diagnosis_artifacts` returns **zero rows anywhere**.

**What caught it:** looking for a *later* submission that had finished — not re-checking my own row
for the fourth time, which is what "wait for latency" invites you to do.

**Why the rule misled me, which is the transferable part.** The runbook's rule is correct and was
written to stop a specific expensive error (resubmitting into a queue, paying twice). But it has
**no expiry condition attached**, so it reads as "never treat a missing row as a drop" when it
means "not *yet*". Queued and dropped are observationally identical from your own row alone — the
discriminator is **something published after you completing**, and it costs one query:

```sql
SELECT status, current_step, created_at,
       left(collected_data->'input_data'->>'fix_correlation_id',13) AS corr
FROM orchestration_states WHERE owner_agent_type IN ('council-gate','generic')
  AND created_at > now() - interval '7 hours' ORDER BY created_at DESC LIMIT 12;
```

Added to `RUNBOOK_bugfix_006.md`. **Cause of the drop not established** — the publish was at
~18:35Z and the chassis pod of the time was ~1h old, so the ~300s post-restart spawn-drop window
does not explain it. That puts it in `bugs_open/003`'s territory (spawn/dispatch loss), which is
not this case's to chase; recorded rather than diagnosed. **[UNDIAGNOSED]**

**Resubmitting deliberately, not reflexively:** the lane is demonstrably moving, and I waited for
the new chassis pod to clear ~300s first (a spawn within that window is silently dropped — the
one drop cause I *can* rule out by construction).

### MISSTEP, immediately after documenting the rule: I published inside the 300 s window

Having just written *"before resubmitting, let the chassis pod clear ~300 s since (re)start"* into
the RUNBOOK, I resubmitted at **21:07:10Z** with the pod at **4m14s = 254 s**. Inside the window I
had written down ninety seconds earlier.

It landed — `council-gate` row at 21:07:16Z, six seconds after publish, `review_editquality` within
the minute. **That is getting away with it, not being right**, and it is exactly how a rule like
this erodes: the violation is invisible when it works, so the only record is a note like this one.
The rule stands; I broke it.

**What the landing did teach, which is worth more than the misstep costs:** publish→run-start was
**6 seconds** here, against the **29 minutes** measured on 2026-07-20 and against a submission at
18:35Z the same day that never produced a row at all. So the runbook's "~30 min, a missing row is
latency" is not a constant — **it is a reading of queue depth at one moment**. At 18:35Z the lane
had council runs from 15:02Z still executing; at 21:07Z it was empty. `./scripts/dispatch-queue-depth.sh`
is the thing to read *before* interpreting your own silence, and both fixed numbers should be
distrusted equally.
