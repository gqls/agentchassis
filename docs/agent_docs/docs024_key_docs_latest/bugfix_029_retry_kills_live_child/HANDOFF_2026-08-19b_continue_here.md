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

## 2026-08-20 ~14:5xZ — AGREED on the ground state, but the 30 was worth one more question (NOTES §19)

A second session re-ran the recurrence check independently and **agrees with every row above** —
`v1.0.1319` since 10:18Z, no recurrence, build point not established and **not gating anything**
(nothing in this lane has been pending a roll since `0132a3683` was proven on `1316`).

**What is new is what the entry condition DID NEXT.** The block above measures how OFTEN it fired
(08-17: 30). Nobody had asked how those 30 ended:

| population | n | went on to register ANOTHER `call_handler` |
|---|---|---|
| `call_handler` at `retry_version=3`, `status='error'` | **31** (30 on 08-17, 1 on 08-15) | **0** |
| **positive control** — healthy `processed` `call_handler`s, 08-17 | 1,387 | **1,054 (76%)** |

**Nothing that reaches rv3/error on a `call_handler` ever recovers.** The control is load-bearing and
came out otherwise: the identical `EXISTS` finds continuation 1,054 times, so the zero is about that
population, not a broken query.

**The 30 split into two exhaustive modes, and only one of them is this bug:**
- **20 WEDGED** — iteration N+1's `spawn_handler` registered, its `call_handler` never. This file's population.
- **10 NEVER ADVANCED** — no N+1 `spawn_handler` either. `[UNVERIFIED]` benign (error on the loop's
  last item) vs a second freeze; **not decidable from retained data** — `orchestration_states` is purged.

**Two consequences for whoever picks this up:**
1. **The suspect widens from the post-spawn continuation to the ERROR PATH itself.** "What kills the
   parent after the spawn" presumes the spawn half is where it breaks; a 0/31 recovery rate says the
   error path is fatal in both modes. This **promotes `PLAN_2026-08-19` §2's P2**
   (`skipToNextLoopIterationForAsync` losing a delete-plus-advance to a non-retrying `UpdateState`),
   filed there as a path never shown to fire — the path now carries 31 terminal outcomes and no survivors.
2. **There is a within-day control sitting in the retained rows: the 20 that advanced vs the 10 that
   did not.** Same day, same entry condition, different outcomes, and it does **not** need
   `orchestration_states`. It is the one comparison still available if no burst arrives before 08-24.

**One door CLOSED since (NOTES §20): the freeze is NOT a dead worker.** `build-dispatch-loop` runs an
ephemeral pod per orchestration (`spawn_actions.go:2388`) with `BackoffLimit: 3`, so "the pod died and
K8s replaced it" was a live infrastructure explanation that would have retired the whole
`coordinator.go` candidate set. It is **refuted**: the 17 duplicate spawn pairs are **17/17 same
`processing_pod`**, and all 98 rows across those orchestrations resolve to **exactly 17 pods** — one
each, never replaced. `ActiveDeadlineSeconds` is 86400, so a job deadline is out too. C1 survives
(a dedicated pod consumes its own message, so same-pod is what a takeover predicts); C5 stays
disfavoured on the minutes-scale gaps.

**Unchanged:** 029 OPEN, nothing fired, `NEXT_090_single_variable.sh` unrun, rows-section withdrawn,
wait-for-the-burst standing. **Also unchanged: none of this explains the first death** — it sharpens
where to look, and no edit should be described as if it had.

## 2026-08-20 15:50Z — the evidence is PRESERVED, and a measurement that may retire a standing claim

**The 08-24 clock no longer binds.** `EVIDENCE_2026-08-15_to_17_awaited_requests.tsv` holds all
**6,484** rows across 1,019 orchestrations, **round-trip proven** to rebuild 31 / 20 / 11 / **0** from
the file alone. ⚠ It took three attempts: `kubectl exec … psql > file` **truncated silently twice**
(4,243 rows, then 5,186, then 6,484 — same command, well-formed final line, no error in the file).
Accepted only on an explicit row-count assertion. Now a `LANDMINES.md` entry; **assert the count
against the DB on any export of this size.**

**Child response times, measured from that file** `[MEASURED 2026-08-20]`, n=3,150 healthy
`call_handler` rows: median **10.4s**, p90 121.4s, p95 226.1s, p99 **454.9s**, max **971.3s**.
Over 300s: **91 (2.89%)**. Over 600s: 11. **Over 1200s: ZERO.**

⚠ **The peer's 300s-exhaustion finding is confirmed** (all 31 errored `call_handler`s gave up at
299.9996–300.4814s) **but the 300s figure is DEFINITIONAL** — pre-Part-A, every rv≥1 window was
capped to 300s for a step declaring ≤30 min. Any rv3 row must read 300s. It is not evidence of cause.

**Two readings, unresolved, and the second would retire a standing claim:**
- **(a)** a retry re-dispatches to a fresh child; the old 300s cap abandoned the **2.89%** that need
  longer; Part A's 1200s covers **100%** of observed responses with 229s of headroom.
- **(b)** reaching rv3 means the request already blew a **1200s** rv0 window, and **no child exceeds
  1200s** — so it was **hung, not slow**, and no window would have saved it. Part A would then be
  **irrelevant to all 31**, and the bug file's *"plausibly makes the entry condition rarer `[INFERRED]`"*
  should be **retired rather than sharpened**.

**What decides it:** whether each retry re-produces the request to a new child or re-waits on the
original (`handleRecoverableError` → `UpdateAwaitedRequestRetry`). **Held by the peer session** — it
is their "what does a thrice-retried, abandoned await do to the loop" thread. Do not duplicate it.

**If (b) holds**, the 08-17 burst is 30 children that never answered, which points away from the
coordinator entirely and at **whatever was happening to the children that day — and the GitHub API
incident overlapping 08-17 is still an unpulled thread.**

## 2026-08-20 16:20Z — THE BURST IS EXPLAINED. The wedge is not. Read this before deciding anything

**The 08-17 entry condition was a GitHub outage, not a coordinator defect.** `agent_error_log`
carries **954** rows matching `%github%` on 2026-08-17 against **1–3 on every other retained day**
(08-10, 08-09, 08-05, 07-27, 07-24) — ~300× base rate, confined to one day. Hourly 13:00–18:00,
peak 495 at 14:00; the wedge window (14:52 → 18:52) sits **inside** it. `page-rerender` — the wedges'
child type — dominates every hour. The error is **GitHub `503 Service Unavailable`** on
`create blob` / `create commit` / `get latest commit-base tree`.

**The controlled join** — 08-17 `call_handler` correlations, split by whether the call was abandoned,
joined on `agent_error_log.context->>'correlation_id'`:

| group | correlations | with a GitHub error | rate |
|---|---|---|---|
| **abandoned (rv3/error)** | 30 | **30** | **100.0%** |
| healthy, same day | 337 | 71 | **21.1%** |

Same day, same agent, same hours; **the control could have come out otherwise.**

**Chain, now closed end to end:** GitHub 503s → `page-rerender` children fail and never answer →
the parent's **1200s** rv0 window expires (§22: no observed child ever exceeds 1200s, max 971.3s, so
this can never have been slowness) → replay to the **same** child (§23, at source) → rv3 → abandoned.

### What this changes

- **Part A converts none of the 31** and the bug file's "plausibly makes the entry condition rarer"
  is **retired** (peer §23). Part A remains correct and worth having — it fixed a real inversion for
  steps whose children are genuinely slow. It is simply not a fix for this.
- **`PLAN_2026-08-19`'s coordinator candidate set may still explain the WEDGE but cannot explain the
  BURST** — and the burst is what made 029 look active.
- **The wedge is now the ONLY open question**: what the parent does after an abandoned await, and the
  20-wedged vs 11-never-advanced split, are exactly as unexplained as before. **A 503 is not
  sufficient on its own** — 21% of healthy correlations saw one and answered anyway.

**INDEPENDENTLY REPRODUCED at a different key and unit** (peer §25), which is what makes it not a
join artefact: keyed on the **page identity** (`context->>'page_name'`) over distinct pages rather
than on correlation over orchestrations, it returns **17/17 = 100% abandoned vs 86/413 = 20.8%
healthy**. Two independent routes, same separation. Limits unchanged: ~1 in 5 healthy pages saw a 503
and answered anyway, so this is association, not mechanism.

⚠ **AND THERE IS A FOURTH BLIND JOIN, WORSE THAN MINE, NOW LANDMINED.** Joining on the **parent's
`orchestration_id`** — a real, populated, obviously-correct column — returns **abandoned 8/30 (27%)
vs healthy 66/337 (20%)**: no separation, reading as a **clean refutation** of everything above.
Structurally blind because **parent and child log under different `orchestration_id`s and
`agent_error_log` has no column spanning them** (verified: 47 `page-rerender` rows under this cohort
carry 47 distinct orchestration ids, none the parent's; all carry the correlation in `context`). A
parent-keyed join therefore sees only `Request <id> timed out after 3 retries` — true, and carrying
none of the child's cause. **The wrong key does not look wrong; it looks like evidence against the
true answer.** Full entry, with all three wrong-key shapes and the same-query non-zero control that
catches them: `LANDMINES.md`.

> ⚠ **The blind join that nearly closed this lead.** Matching the 31 `request_id`s against
> `error_message` finds all 31 and **zero** GitHub rows — the only rows carrying those ids are the
> *parent's own* "timed out after 3 retries" entries. Joining children on
> `orchestration_id LIKE '<topic prefix>%'` returns **0 of 31 with any log row**, which reads as
> "the children were silent, GitHub exonerated". **Both are blind:** the topic prefix is the
> **correlation** id, and `agent_error_log` has **no `correlation_id` column** — it keys it inside
> `context` jsonb. **A join on a column that cannot match returns a clean, confident zero.** Always
> run the same join over rows that MUST match before believing one.

## 2026-08-20 — LANE STATE CHANGED: the BURST is explained, the WEDGE is the only open question

**Confirmed independently by a second session (NOTES §25), at a different join key and unit —
100% vs 20.8% against §24's 100% vs 21.1%, plus the 954-vs-1/3/3/2/1 census reproduced exactly.**
So the association is not a join artefact.

| | status |
|---|---|
| **the 08-17 BURST** | **EXPLAINED.** A GitHub 503 outage that day (~300× base rate, 13:00–18:00, wedge window inside it) stopped `page-rerender` children answering. A child that never answers blows its **1200 s** rv0 window — 229 s beyond the slowest response ever observed — so the entry condition is an **outage artefact, not a coordinator phenomenon** |
| **the WEDGE** | **UNCHANGED and OPEN.** What the parent does with an abandoned await, and the 20-advanced-vs-10-stopped split, are exactly as unexplained as before |
| `PLAN_2026-08-19`'s candidates | **never wrong, but they only ever spoke to the WEDGE.** They cannot explain the burst — and the burst is what made 029 look active |
| a 503 as cause | **NOT sufficient.** ~1 in 5 healthy pages saw one and answered anyway. Association, not mechanism |

**⚠ Before you write ANY parent↔child join against `agent_error_log`, read NOTES §25.** Parent and
child are logged under **different `orchestration_id`s** and the table holds **no key spanning them**
(no `correlation_id` column; 0 of 367 correlation ids appear as an `orchestration_id`; neither 8-hex
topic prefix matches a child row — the second is the PARENT's id, 367/367). The obvious key, the
parent's `orchestration_id`, is **populated, non-null, and structurally blind**: it returns 27% vs
20%, which **reads as a refutation of the outage finding**. Four joins in this family have now come
out blind between two sessions. **The only sound linkage is the payload's `page_name`**
(`request_payload → message.body.input_data.spec.page_name` = the child's `context->>'page_name'`),
and the check that catches it is a **must-be-non-zero control in the SAME query as the claim**.

### 2026-08-20 17:10Z status, `v1.0.1320` (pods 16:09Z) — no change, nothing owed

Wedges **20, all 08-17** (no recurrence). Abandoned calls **0/439 today, 0/736, 0/1595** — three
clear days. GitHub-`503` rows **still 08-17 only** (954, against 1–3 base) — the outage has not
returned. Capture cron last scheduled **16:17Z**. Evidence preserved, so nothing expires.
Build point not probed: no 029 code has shipped since `v1.0.1316` and both fixes are ancestors of
everything since, so it is not load-bearing — see the 08-20 14:30Z note.

**Nothing is owed on this lane.** The only open question is the wedge, it is not reproducing, its
evidence is preserved, and the split is with the owner.

### 2026-08-20 ~17:20Z — THE SPLIT IS BEING ACTIONED BY THE OTHER SESSION. THIS SESSION HAS STOOD DOWN

**Do not start a re-file and do not edit `bugs_open/029`.** The peer session reports that the owner
ruled **YES, split it** — close 029 for the retry-window defect, re-file the wedge under its own
number — and is actioning it now. They will publish the new bug number and the commit.

> **⚠ Epistemic status, stated because it matters for who acts:** the ruling landed in **their**
> session. I have **not** heard it directly and have **not** verified it — it is **attested by the
> peer session**, exactly like the banner-authorship question earlier today. That is sufficient
> grounds for me to **stand down** (refraining needs no proof), and it is **not** sufficient grounds
> for me to act on 029 myself. If you are picking this lane up and the new bug does not exist,
> **ask the owner rather than assuming the split was approved.**

**What the new bug will carry** (their plan, recorded so it is not lost if either session ends): the
wedge as the sole open question; the 20-advanced vs 10-stopped split; 0-of-31 recovery with the
971.3s response ceiling; **the outage as explained CONTEXT, not cause**; the preserved-evidence path;
the four-blind-joins warning pointing at the `LANDMINES.md` entry; and `PLAN_2026-08-19` marked
still-valid-but-wedge-only. They will pathspec both paths on the move and verify at HEAD with
`git ls-tree` (the `git mv` + pathspec copy landmine).

**The close-out will carry this file's top banner wording verbatim** rather than a re-draft, including
*"Do not read a fixed Part A as a fixed hang."*

**Still unowned and NOT part of the split:** the `workflow%` bundle-include widening (one line, blast
radius measured, the 301 lane told, that lane closed). It needs someone to ship it or formally drop
it. It is not blocked by anything.
