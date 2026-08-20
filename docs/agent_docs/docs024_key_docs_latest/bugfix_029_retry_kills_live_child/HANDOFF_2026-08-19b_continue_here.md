# HANDOFF — 2026-08-19b — `bugs_open/029`, continue here

Supersedes `HANDOFF_2026-08-19_continue_here.md` (still accurate on Part A, RSH-011 and the
baseline trap — read it for those). Then `NOTES_retry_kills_live_child.md` §9 and §10.
`README_where_we_are.md` is the owner's plain-prose log — append, never rewrite.

## State in one line

**029 stays OPEN.** Part A fixed/approved/live/proven, unchanged. The wedge now has a **named class
with three verified source sites** and a **retired wrong subsystem**; a fourth 090 is in flight; and
one real framework gap in the diagnosis harness was found and is **not yet fixed**.

> **UPDATED ~17:00Z — the 090 verdict is IN and the plan is written. Both are recorded below;
> the "IN FLIGHT" table is kept for the audit trail, not as outstanding work.**

## ✅ RESOLVED SINCE THIS FILE WAS WRITTEN

- **090 `d52c3407` verdict: `UNVERIFIABLE` — and it is a REAL abstention, the first run to READ the
  wedge.** Four citations, three `tier: state` quoting actual rows: it pulled `04518118`'s rows and
  the control and confirmed the signature. It abstained because the bundle carries no *bodies* for
  `handleCompleteResponse`/`continueExecution` and `orchestration_states` — the only place the
  map-vs-table divergence is observable — is purged for all six. Compatible-with is not
  observation-of. **NOTES §11.**
- **Its challenge to the rv0 finding is REFUTED at source. §9 stands.** It asked whether a retry path
  could reset `retry_version` to 0 on a fresh `request_id`. Only two `INSERT INTO awaited_requests`
  sites exist (`state.go:1611`, `spawn_actions.go:166`), **neither on a retry path**, and all three
  `retry_version` writers are `UPDATE … WHERE request_id`. Its own citation settles it inside one
  orchestration: `04518118`'s `call_handler` is **1 row at rv3**, its `spawn_handler` **2 rows at rv0
  with distinct ids**. Also found: `UpdateAwaitedRequestForRetry` (`state.go:1962`) is **DEAD**; the
  live writer is `UpdateAwaitedRequestRetry` (`coordinator.go:3337`). **NOTES §11a.**
- **C1 survives a test that could have refuted it. NOTES §12.** 17/17 duplicate gaps exceed the 300 s
  `StuckOrchestrationTimeout`, at 414–577 s. C1 and C2 both survive (identical table signature,
  nothing separates them); C3 refuted, C4/C5 disfavoured. Separating C1 from C2 needs logs the 08-17
  pods no longer have → next burst, RSH-011 armed.
- **The fix plan is written and GRADED:** `PLAN_2026-08-19_wedge_fix_park_advance_divergence.md`.
  Fable-drafted, checked against the live system. **Two draft claims did not survive** — a dead
  citation, and a "divergence census 0 and 0 `[MEASURED]`" that is **vacuous** (zero orchestrations
  are `AWAITING_RESPONSES` and zero `awaited_requests` rows are non-terminal fleet-wide, so 0/0 was
  forced). It was headed for a council submission; do not carry it.

## ⏳ IN FLIGHT — read these two first

| | |
|---|---|
| **090 run `d52c3407-14e7-4b9e-be46-c8ee741b2532`** | filed 16:14Z with six frozen `orchestration_id`s + a healthy control seeded into the symptom. **Verdict UNREAD.** `SELECT owner_agent_type,current_step,status FROM orchestration_states WHERE correlation_id::text='d52c3407-14e7-4b9e-be46-c8ee741b2532' ORDER BY created_at;` then `collected_data->'verdict'` on the `diagnose-agent` row |
| **Landmine verifier `a709e01b-afac-48ee-9e49-6aaf43fb8617`** | armed for the new 090-trigger entry |

## DONE 2026-08-19 (this session), all committed

- **`d609cedad`** — the rv0 measurement + NOTES §9 + the LANDMINE.
- **`805eba66b`** — `bugs_open/029` gains the 2026-08-19 section; owner log appended.
- **`7853605ce`** — NOTES §10, the framework gap.
- Council **APPROVED, all reviewers, round 1** on `e03f7122-7895-4b81-8add-5a93f69ed553`
  (the `schemaAlwaysTables` fix, commit `0132a3683`). `098` credits it automatically from the
  `Council-Submitted:` trailer — **do not amend, forward-only.**

## THE FINDING THAT RETIRES A WEEK OF SUSPECTS

`[MEASURED]` from `awaited_requests` (7-day table). 20 instances, all 08-17, none since.
**17/20 register the same `iter_{N+1}_spawn_handler` twice; `retry_version` is 0 on all 37 spawn
rows; 0/20 ever register the next `call_handler`.**

**A retry bumps `retry_version`. So the duplicate is NOT the takeover re-running the step — the step
BODY executed twice.** `handleSpawnRetry` and the whole retry-machinery candidate set are aimed at
the wrong subsystem. The "06:54–09:37 gap consistent with a >5-min takeover" agreed with the
hypothesis and could not have disagreed with it. `WRONG_CALLS.md` 2026-08-19.

## THE CLASS — three sites, all `[VERIFIED at source]`, composition `[UNVERIFIED]`

**The outstanding-awaited set has two representations and nothing reconciles them.**

1. `coordinator.go:2113` `persistAwaitingStateWithRetry` — the "did a reply beat the park?" check is
   keyed on **StepName**, not the request id. On a hit it returns `nil` **without** writing the
   awaited entry or setting `AWAITING_RESPONSES`; the caller reads `nil` as success and still calls
   `InsertAwaitedRequest` (`state.go:1609`). Row in the **table**, nothing in the **map**.
2. `coordinator.go:2671` `handleCompleteResponse` — `allDone := len(freshState.AwaitedRequests)==0`,
   from the **map alone**. That one boolean decides advance-vs-park and whether `continueExecution`
   runs. **The table is never consulted.**
3. `coordinator.go:848` `continueExecution` — silent early `return nil` when the loaded status is
   `AWAITING_RESPONSES`.

(1) creates divergence, (2) turns divergence into a wrong decision, (3) makes it silent.

## ⚠ THE FRAMEWORK GAP — found today, NOT fixed, and it is the best next platform change

`diagnose_load_runtime_action.go` renders row sections for `agent_error_log` (:274),
`site_work_items` (:309) and `orchestration_states` (:344). **There is no `### awaited_requests`
section.** `orchestration_states` retains ~26h; `awaited_requests` retains 7 days. So for any hang
older than a day the bundle renders rows from the table that is empty and describes the columns of
the table that is full. `0132a3683` fixed the schema half only.

**Three 090 stalls, three causes, all ours:** wrong table → right table/no schema → right
table/right schema/**no rows and no ids**. Rule earned: **when a run refutes on absence, establish
WHICH absence from its `needed_evidence` before re-filing.**

> **SUPERSEDED 2026-08-19 — this is now LIVE and PROVEN; the paragraph below is kept as the record
> of what was true at 17:00Z.** `v1.0.1316` rolled 17:13Z, `0132a3683` is an ancestor of build point
> `07eeba4a1`, and the bundle renders the table — proven behaviourally twice, independently
> (NOTES §11 from `diagnosis_artifacts`, §17a from `orchestration_states`, with controls both times).

⚠ **`0132a3683` is Go and INERT until the next chassis roll.** Pods are on `v1.0.1315` (12:15Z);
the commit is 16:00Z, so it is **not aboard**. Verify after the roll at the build stamp, then by
checking a fresh bundle's Schema section actually describes the table.

## What is left

1. ~~**Read `d52c3407`'s verdict.**~~ **DONE — see above.** Next diagnosis action is NOT another
   re-file: the loop's abstention turns on evidence that no longer exists for this cohort. **Wait for
   the next burst**, which RSH-011 captures automatically with the full `awaited_requests` set.
   ~~ CONFIRMED changes what to build; REFUTED is a result and must be
   recorded as a visible correction here.
2. ~~**Build the `### awaited_requests` rows section** in `diagnose_load_runtime` — council-gate it.
   Ranked above any symptom-text workaround by the close-the-door rule: the workaround must be
   remembered by every future filer, and those who forget get a bundle that looks complete.~~
   > **CORRECTED 2026-08-19 — DO NOT BUILD THIS AS SPECIFIED (NOTES §17c).** It would render
   > **empty**. Every static row section is scoped to the diagnosis TARGET
   > (`diagnose_load_runtime_action.go:279-280`, `:314`, `:349-350`), and `[MEASURED]` the target
   > correlation has **0** `awaited_requests` rows against **1,469** in the 08-17 incident window.
   > The `orchestration_states` section is already scoped that way and already renders
   > `(no orchestration rows for this correlation/site)` — that empty string is the verdict's own
   > first citation. The **schema** half (`0132a3683`, now live) was the whole blocker: the loop
   > writes its own `awaited_requests` query and `runDataRequests` returns the rows. The residual
   > defect is narrower — an unfiltered dump meets `row_cap=200` with an alphabetical `ORDER BY`.
3. **Answer "what runs the step body twice at rv0"** — loop expansion / `ErrLoopExpansionHandled`,
   the recursive `continueExecution`, or a second consumer. Do NOT re-enter the retry machinery.
4. **Do not close.** Bar is fixed AND live; nothing about the wedge is fixed. Quiet since 08-17 is a
   baseline (6 of 8 surrounding days are also zero), not evidence.

---

# APPENDED 2026-08-19 ~22:15Z by a SECOND, CONCURRENT SESSION — and an apology in the form of a warning

**I overwrote this file with `cat >` at 22:15Z and have restored it byte-identically from
`9857c0b2e`. Nothing above this line is mine; nothing above it was lost.** I did not run
`scripts/who-owns.py 029` and did not check `git log` on this directory before starting, so I worked
this lane all day without knowing you were on it. **The recovery worked only because the file was
committed** — exactly the case CLAUDE.md makes for the versioned-memory hook. `WRONG_CALLS.md`
2026-08-19.

**Read YOUR sections above as primary.** Where we disagree, you are right and I say so below.

## Where your findings CORRECT mine

- **The duplicate spawn is NOT the takeover — `retry_version` is 0 on all 37 spawn rows.** My
  NOTES §14 called the 06:54–09:37 gap "consistent with the already-established >5-min takeover".
  **That is refuted by your measurement, and worse, I had the refuting data on screen** — my own
  query output shows `retry_version 0` on both spawn rows of every pair. I read past it because the
  takeover story already fitted. Your §"THE FINDING THAT RETIRES A WEEK OF SUSPECTS" is right and my
  line is withdrawn.
- **Your three-site class subsumes my NOTES §6.** I found site (1), the step-name-keyed arrival
  check, independently and stopped there. You have the composition — (2) `allDone` computed from the
  **map alone** at `handleCompleteResponse`, (3) `continueExecution`'s silent early return — which is
  what turns divergence into a wrong decision and then hides it. Mine is a fragment of yours.
- **Your ranking on the framework gap is better than what I did.** You wrote: build the
  `### awaited_requests` **rows** section, and rank it above any symptom-text workaround, because a
  workaround must be remembered by every future filer. **I did the workaround. It failed** (below).

## What is genuinely NEW from my session

- ~~**⚠ YOUR LINE "`0132a3683` … is NOT aboard" IS NOW STALE**~~ — **CORRECTED 2026-08-20: it is
  not stale, it is deliberately SUPERSEDED IN PLACE.** Its author added a banner directly above it
  (line ~96) recording the roll, and kept the original sentence beneath as the record of what was
  true at 17:00Z. **Leave the pair intact; do not delete the original line** — without it the next
  reader loses why the file ever said that. What follows was my addition to the same fact:** Pods are on
  **`v1.0.1316`** (17:13Z), build point **`07eeba4a1`** present on both replicas with the previous
  build point absent as a control; `0132a3683` **is an ancestor**. So the schema half is live.
- **And it is behaviourally PROVEN, by the check you specified.** A fresh bundle's Schema section
  renders `awaited_requests(request_id varchar, …, retry_version integer, …,
  processing_started_at timestamp, …)`. Control: **four pre-fix bundles render nothing** for it,
  while `orchestration_states(` is present in **all five** — so the section was fine before and the
  gap was specific to this table. NOTES §11 (mine — note our NOTES numbering has collided; see below).
- **The 7-day window is now grounded at SOURCE, not observed:** `cleanup_expired_awaited_requests()`
  runs `DELETE … WHERE status IN ('processed','expired','cancelled','error') AND processed_at <
  NOW() - INTERVAL '7 days'`. Keyed on **`processed_at`**, terminal rows only, enforced per minute.
  **The 08-17 cohort dies ~2026-08-24.** A Go-only `grep "DELETE FROM awaited_requests"` returns
  **nothing**, which is why this was assumed rather than checked.
- **Two more 090s were spent, against your "do not re-file" instruction, which I had not read.**
  `d02a6958` (3 iterations) cited a real 08-17 row and the right code path but stopped one query
  short. `5d1d8f1c` **regressed to 1 iteration** (`stopped_by = scope-not-narrowing`) because I
  changed **three things at once** — so nothing can be attributed. **Your instruction stands and mine
  is withdrawn: do not re-file; wait for the burst.** The artefacts
  (`NEXT_090_single_variable.sh`, `RECONSTRUCTION_QUERY.sql`, `SYMPTOM_d02a6958_baseline.txt`) are
  left in the lane dir for whenever a re-file IS warranted; **they are not a recommendation to fire
  one now.**

## ⚠ HOUSEKEEPING THE NEXT READER MUST KNOW

- **`NOTES_retry_kills_live_child.md` now has TWO §9–§11 sequences**, yours and mine, appended by two
  sessions that could not see each other. Neither is wrong; the numbering is. **Resolve by date and
  content, not by number** — mine are timestamped 2026-08-19 from ~10:45Z onward and concern
  retention, the ticker refutation, the bundle fix and the two runs above.
- **Traps I paid for that are not in your sections:** any check against a diagnosis bundle for a
  string YOU authored is blind (the symptom is quoted into the bundle verbatim — `LIKE
  '%awaited_requests%'` returns true on a *pre-fix* bundle); resolve a council verdict **by
  correlation, never by recency** (`ORDER BY created_at DESC LIMIT 1` returned another lane's
  approval); `row_cap` is **200** and an unfiltered dump `ORDER BY orchestration_id` returns a
  lexicographic slice; and flattening a `.sql` file to one line **without stripping its `--`
  comments** comments out the whole query and returns **zero rows with no error**.

---

# 2026-08-20 09:10Z — START HERE IN A FRESH SESSION

Appended by the session that wrote the appendix above. **THREE sessions have worked 029** (the
file's original author, the §17 contributor, and me). Everything below is current as of this
timestamp and supersedes anything above it that conflicts.

## FIRST, BEFORE YOU TOUCH ANYTHING

```bash
python3 scripts/who-owns.py 029          # I skipped this and worked an owned lane all day
git log --oneline -8 -- docs/agent_docs/docs024_key_docs_latest/bugfix_029_retry_kills_live_child/
ListAgents                                # "bugfix 029" is a live peer session
```
**Use the Write tool, never `cat >`, on any file in this directory.** I overwrote this very file with
a redirect; Write would have refused. Recovered from git only because it was committed.

**SETTLED 2026-08-20 — nothing was lost, and note WHAT KIND of fact that is.** A `SUPERSEDED` banner
of this file's author first appears in a commit **four minutes after my restore**, and they recalled
writing it about an hour earlier — which would have put it inside my clobber window. **Git cannot
separate "written once, late" from "destroyed and re-typed from memory": `git log -S` dates the first
COMMIT, never the first keystroke.** It was settled only by the authoring session reading its own
transcript (one `Edit`, running continuously into that commit), so it is **attested by them and
unverifiable from git** — not measured. The transferable move, now a `LANDMINES.md` entry: after any
overwrite of a shared file, **ask the authoring session "did you type this once or twice?"**

## DO NOT FIRE `NEXT_090_single_variable.sh` YET

It is armed, correct, and **not** the right next action. The standing instruction from this file's
author holds: **wait for the next burst**, which RSH-011 captures automatically. Three 090s have been
spent; the loop's remaining abstention turns on evidence that no longer exists for the 08-17 cohort.
The script exists for whenever a re-file IS warranted.

## The scope guard is fully explained — do not re-open it

`[VERIFIED at source + MEASURED 2026-08-20]`, NOTES **§18**. Guard: `pkg/diagnose/loop.go:432`,
`next.size() > prevSize+2`; `size()` = `len(Symbols)` (:205); `namedScope` (:398) takes Symbols from
`v.NextScope` **alone**; init `PrevScopeSize = seed.size()+1` (`advance.go:68`); on a stop
`advance.go` returns at :104-111 **before** :120 overwrites it, so persisted `prev_scope_size` is
pre-guard. **Read `route.stopped_by`** — `diagnosis.stopped_reason` is a different, empty key.

| run | seed | init prev | trip | persisted |
|---|---|---|---|---|
| `d02a6958` | 5 | 6 | iter1 named 8 (8>8 **false**) → prev 8; iter2 named 5 → prev 5; **iter3 named 12: 12>7 TRIPS** | 5 ✓ |
| `5d1d8f1c` | 6 | 7 | **iter1 named 13: 13>9 TRIPS** | 7 ✓ |

**Three claims are now dead. Do not resurrect any of them:**
- ~~`d02a6958` "ran out of iterations one query short"~~ — it stopped at 3 of 5 **on the guard**.
- ~~seed widening caused `5d1d8f1c`'s halt~~ — **wrong in direction.** Threshold is `prevSize+2` and
  `prevSize` starts at `seed.size()+1`, so a **wider seed is PROTECTIVE**. `5d1d8f1c` had the wider
  seed *and* the higher threshold and tripped anyway.
- ~~`5d1d8f1c` does not reconcile~~ (peer §17) — it does; their operand was named 9, it is **13**.

What actually differed: the model named **13** symbols in iteration 1 where the baseline named **8**.
Symptom length or variance — `[UNVERIFIED]`, and one run per condition cannot separate them.
`d02a6958` survived iteration 1 **by exactly one symbol** (8 > 8 is false).

## The `### awaited_requests` rows section: DO NOT BUILD IT

Ranked as "best next platform change" earlier in this file, and I endorsed it in NOTES §16(c).
**Both withdrawn** (peer §17, verified). Every static row section is scoped to the diagnosis **target**
(`diagnose_load_runtime_action.go:279-280, :314, :349-350`); the target correlation has **0**
`awaited_requests` rows against **1,469** in the 08-17 window. The natural experiment is already in the
bundle: `orchestration_states` is scoped identically, renders *"(no orchestration rows for this
correlation/site)"*, and **that empty string is a verdict's own first citation.** A new section renders
`(no rows…)` for any historical incident described in prose.

**The schema half was the whole blocker and it is shipped, live and proven.** Residual is narrow:
unfiltered dump + `row_cap=200` + alphabetical `ORDER BY`.

## Live on `v1.0.1316` — verified at the artefact

Build point **`07eeba4a1`** present on both replicas; previous build point `590ca3a20` **absent** as a
control. `bf7646a29`, `2a3d30ec3`, `0132a3683`, `3ba384c63` are all ancestors. **This is the same fact as the SUPERSEDED banner at
line ~96 — that banner and the original "not aboard" sentence beneath it are a deliberate PAIR.
Do not delete either half.** Behavioural proof of the bundle fix: a fresh
bundle renders `awaited_requests(request_id varchar, …)`; **four pre-fix bundles render nothing**;
control `orchestration_states(` present in **all five**.

## ⏱ EVIDENCE DEADLINE: ~2026-08-24 (about four days)

`[VERIFIED at source]` `cleanup_expired_awaited_requests()`, called every minute:
`DELETE FROM awaited_requests WHERE status IN ('processed','expired','cancelled','error') AND
processed_at < NOW() - INTERVAL '7 days'`. Keyed on **`processed_at`**, **terminal rows only**,
continuous — no nightly-job grace. A Go-only grep for that DELETE returns **nothing**; the retention
is DB-resident.

## What is actually left

1. **Wait for the next burst.** RSH-011 (`wedge-evidence-capture`, hourly at `:17`) captures it with
   the full `awaited_requests` set. That is the whole remaining path on the wedge — separating the
   author's C1 from C2 needs logs the 08-17 pods no longer have.
2. **The live hypothesis is the author's three-site class** (divergence created at
   `persistAwaitingStateWithRetry`'s step-name-keyed arrival check; turned into a wrong decision by
   `allDone` computed from the **map alone** at `handleCompleteResponse`; hidden by
   `continueExecution`'s silent early return). My independently-found arrival-check defect is a
   **fragment** of this — fix it as part of the class, not solo.
3. **Answer "what runs the step body twice at rv0"** — `retry_version` is 0 on all 37 spawn rows, so
   it is **not** the takeover. Loop expansion / `ErrLoopExpansionHandled`, the recursive
   `continueExecution`, or a second consumer. **Do not re-enter the retry machinery.**
4. **`workflow%` include widening — TOLD, not taken.** `flow%` is a prefix pattern that never matched
   `workflow%`; the 301 lane's run `dd61df1b` stalled on exactly that. **They have been told**
   (`bugfix_301_owned_guard_ordering/NOTES_owned_guard_ordering.md`, CONTRIB 2026-08-20) with the
   blast radius measured: cap **120**, ~94 in use, `workflow%` adds **2**. ⚠ read the LIVE config's
   `schema_include_patterns`, not the Go default — a running bundle says *"33 of 479 shown"*.
5. **Do not close 029.** Bar is fixed AND live; nothing about the wedge is fixed. Quiet since 08-17 is
   the baseline (six of eight surrounding days are also zero), not evidence.

## ⚠ A mis-attribution in a commit message, correctable only by note

`400269574`'s message credits **me** with two `WRONG_CALLS.md` rows that are **not mine** — the
*"I called 20 rows an open queue"* row and the *"cannot be faked"* row. Both self-identify inside
the file as the **`bugfix_313_internal_linker`** lane's (`bugs_open/313`, `/298`). They were sitting
uncommitted in the shared tree and rode along in a pathspec commit, which is unavoidable and was
declared in good faith — the credit is simply wrong. **The rows themselves are correctly attributed;
only the commit message is not**, and forward-only forbids an amend. If you are counting this file's
tally by lane, read the rows, not `git log`.

## ⚠ NOTES numbering is collided across three sessions

`NOTES_retry_kills_live_child.md` has **two §9–§11 sequences** and then §§16–18 from two different
sessions. **Resolve by date and subject, never by number** — several are already cited by number from
other documents, so renumbering would break references.

## Traps paid for, that are not in the sections above

- **Any check against a bundle for a string YOU authored is blind** — the symptom is quoted in
  verbatim, *and* `orchestration_states` has a **column** named `awaited_requests jsonb` rendered in
  every bundle ever. So `LIKE '%awaited_requests%'` reads true for two independent wrong reasons.
  Match the renderer's `(`.
- **Resolve a council verdict BY CORRELATION.** `ORDER BY created_at DESC LIMIT 1` on `doc_notes`
  returned another lane's APPROVED note; I nearly recorded it as ours.
- **Retention is PER STATUS.** `min(created_at)` on `orchestration_states` reads five weeks because
  `CANCELLED` is never pruned; grouped by status it holds two days. Errs *reassuringly*.
- **Flattening a `.sql` to one line without stripping `--` comments** comments out the whole query and
  returns **zero rows with no error**.
- **`row_cap` is 200.** An unfiltered dump `ORDER BY orchestration_id` returns a lexicographic slice.


---

## 2026-08-20 14:30Z — status check on `v1.0.1319`. NOTHING HAS CHANGED ON THE GROUND

| check | result |
|---|---|
| chassis build | pods on **`v1.0.1319`**, started 10:18Z |
| build point | **NOT ESTABLISHED — stated rather than guessed.** The `build provenance` line had long scrolled (~4h uptime, ~4min retention) and the binary probe could not be completed: 160 candidate commits, one `exec` each is ~5s and a single alternation grep over the binary both exceeded the command timeout. **This does not gate anything** — no 029 code has shipped since `v1.0.1316`, where Part A and `0132a3683` were both verified present with controls, and both are ancestors of every later commit. If you need it, narrow the candidate window first |
| **wedge recurrence** | **NONE.** Still 20 instances, all 2026-08-17 |
| entry condition (terminal-`error` `call_handler`) | 08-18 **0/1595**, 08-19 **0/736**, 08-20 **0/380**. The 08-17 spike (30/1432) remains the only one |
| evidence expiry | 2,901 `awaited_requests` rows for 08-17; oldest `processed_at` 00:32:27, so deletion begins **2026-08-24 00:32** and rolls through the day. **~3.5 days left** |
| RSH-011 capture cron | **alive** — last scheduled 14:17Z, job `Complete` in 15s, as it has been hourly |

**So the lane is exactly where the 08-20 09:10Z block left it: rare, bursty, unexplained, not
reproducing, with a capture armed and an evidence clock running.**
