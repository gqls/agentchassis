# HANDOFF — 2026-08-19b — `bugs_open/029`, continue here

**This supersedes `HANDOFF_2026-08-19_continue_here.md`**, which supersedes `..._2026-08-18b_...`
and `..._2026-08-18_...`. The 08-19 one is still worth reading for the *corrections* it records in
place (it is where the "evidence has expired" error was caught), but **this file is the current
state**. Then read `NOTES_retry_kills_live_child.md` (newest at the bottom; §§10–15 are today).
`README_where_we_are.md` is the owner's plain-prose log — append, never rewrite.

---

## State in one line

**029 is OPEN and its cause is still unexplained. What changed today is that the investigation is
no longer blocked: the evidence was never lost, one candidate is refuted, the freeze is narrowed to
a single transition, and the tooling that could not read the evidence is fixed, approved and
proven. Four diagnosis runs have been spent; the best of them stopped one query short.**

## ⏱ THE ONE DEADLINE THAT MATTERS

**The 08-17 evidence — all 20 instances — is deleted at `processed_at + 7 days`, i.e. about
2026-08-24.** `[VERIFIED at source]`: the rule is DB-resident, in `cleanup_expired_awaited_requests()`,
called every minute, `DELETE FROM awaited_requests WHERE status IN
('processed','expired','cancelled','error') AND processed_at < NOW() - INTERVAL '7 days'`. Terminal
rows only; keyed on `processed_at`, **not** `sent_at`; enforced continuously, so there is no "the
nightly job may not have run yet" grace. After that date the lane is back to waiting for a burst.

---

## ➡️ DO THIS FIRST: the one-variable re-file

```bash
./docs/agent_docs/docs024_key_docs_latest/bugfix_029_retry_kills_live_child/NEXT_090_single_variable.sh
```

It files a `090` that is **`d02a6958`'s symptom and seed scope byte-identical, plus the
reconstruction query and nothing else.** Everything it needs is in the lane dir:
`SYMPTOM_d02a6958_baseline.txt` (pulled from the DB, not retyped), `RECONSTRUCTION_QUERY.sql`
(executes; returns 20 rows, all `next_call_registered = false`).

**Why one variable.** Run `5d1d8f1c` changed THREE things at once and regressed from 3 iterations to
1. Nothing could be attributed. So this run tests exactly one hypothesis:

| outcome | what it establishes |
|---|---|
| reaches iteration 2 and reads the query's results | the **seed widening** caused the regression → rule: **never seed with the previous run's `NextScope`; that is what the loop is for** |
| stops again at `scope-not-narrowing` | the **SQL or the longer symptom** is implicated → deliver the reconstruction some other way |

**DO NOT add the previous run's `NextScope` symbols.** That is the variable under test.

⚠ **Trap already paid for, inside that script:** the `.sql` file has leading `--` comments, and
flattening it to one line without stripping them comments out the whole query — which then returns
**zero rows with no error**, indistinguishable from "found nothing". The script strips them
(`grep -v '^[[:space:]]*--'`). Verified both ways 2026-08-19.

---

## What is DONE — do not redo any of it

| | evidence |
|---|---|
| **Part A (RSH-010) fixed, approved, live, PROVEN** | `call_dispatch` rv1 granted **00:15:00** (its declared 900s) 2026-08-18 18:28:21Z, `status=processed` — answered INSIDE the window, against **219 pre-roll retries, none above 05:00** |
| **Live on `v1.0.1316`** | build point **`07eeba4a1`** present on both replicas; previous build point `590ca3a20` **absent** on both. `bf7646a29`, `2a3d30ec3`, `0132a3683`, `3ba384c63` all ancestors |
| **The evidence was NEVER lost** | `orchestration_states` retains ~26h and did lose it; **`awaited_requests` retains 7 days and holds all of it** — **20 instances on 08-17, two MORE than the 18 `orchestration_states` ever showed**, 20/20 with no next `call_handler` |
| **Candidate 1 (ticker's shared 60s ctx) REFUTED, three ways** | (1) batches are never shared — **31,548 of 31,548 claims are batches of exactly ONE** (98.7% row coverage, 7-day window, burst included); (2) freeze lands **~12–35 s into 60**; (3) the path it dies on has **no deadline at all** (`c.ctx`, agent-lifetime) |
| **Freeze narrowed to ONE transition** | 37/37 spawn rows `processed` at the freeze offset. The child answered, the parent handled it and wrote state, then died **inside `continueExecution` for `iter_{N+1}_call_handler`, on the response-consumer goroutine** (strictly serial, `client.go:77`, commits only on success) |
| **Bundle fix `0132a3683` — approved AND proven** | Council **APPROVED round 1, all reviewers, zero objections** (corr `e03f7122`, verdict read). Behavioural: post-fix bundle renders `awaited_requests(request_id varchar, …)`; **all four pre-fix bundles render nothing**; control `orchestration_states(` present in **all five** |
| **Evidence-loss hole closed (RSH-011)** | `wedge-evidence-capture`, hourly CronJob at `:17`, live; captures live wedges at freeze+30min and reaped ones into `doc_notes`. Verified running 2026-08-19, "newly captured: 0" |
| **Initial-wait lead REFUTED** | `call_diagnoser` rv0 is 30:00 in 29/29 rows; 18 agent/step pairs, 18 honoured, 0 mismatches |

## THE OPEN QUESTION

**What kills the parent inside `continueExecution` for `iter_{N+1}_call_handler`, on the
response-consumer goroutine, immediately after `handleCompleteResponse` stamped the spawn
processed — on the path entered when the previous iteration's `call_handler` ended in `error`?**

The same orchestration registers `call_handler` fine on healthy iterations (`23eb0107` did it three
times before wedging), so **something differs about a continuation reached via
`skipToNextLoopIteration`.** Live candidates, both `[UNVERIFIED]`:
- whether `persistAwaitingStateWithRetry`'s reload-and-copy drops what the next step resolves against
  (note `carryCollectedDataOntoFreshState`, `3ba384c63`, already addresses the *discard* half);
- **pod death** (OOM/crash) at that instant would leave an identical trace — untestable for the
  08-17 rows, cheap to settle on the next capture.

## The four diagnosis runs, and what each is worth

| run | result | what it means |
|---|---|---|
| `b346d0d4` (09:47) | NOT CONFIRMED | refuted on absence — **it was told the evidence had expired.** The premise was false |
| `d8af5f78` (11:05) | UNVERIFIABLE, 1 iter | **blocked by our harness** — `awaited_requests` had no schema in the bundle |
| **`d02a6958` (20:38)** | **UNVERIFIABLE, 3 iters** | **THE HIGH-WATER MARK.** Cited a real 08-17 row (Tier 1, `Fresh: 2026-08-17`, orch `838f8c14`) and the right code path; self-corrected a 200-row truncation; **stopped one query short of the outcome test** |
| `5d1d8f1c` (20:58) | UNVERIFIABLE, 1 iter, `scope-not-narrowing` | **a regression, caused by this session.** Three simultaneous changes ⇒ no attribution |

**Read `Citations`, `NeededEvidence`, `NextScope` and `stopped_by` — never `status` alone.** All four
runs share a status and mean four different things.

## Also open, with the work already done

1. **`workflow%` include widening — the bundle class is NOT closed.** `flow%` is a **prefix** pattern
   and never matched `workflow%`. Another lane's run `dd61df1b` stalled the same morning needing
   `workflow_templates`, `workflow_contract_chain`, `v_active_workflows`, `v_all_workflows`; my
   one-table fix does nothing for it. **Blast radius already measured:** cap **120**, ~94 in use,
   `workflow%` adds **2** (all `v_` views would add 11). ⚠ read the LIVE agent config's
   `schema_include_patterns` first — the running bundle reports *"33 of 479 tables shown"*, far
   narrower than the Go default. Per the 2026-07-29 ruling, **tell that lane** — it is their symptom.
2. **The arrival check in `persistAwaitingStateWithRetry` is keyed by STEP NAME, not request id.**
   So a re-registered step (what the takeover does) trips on the *previous* reply's marker, returns
   `nil` **without persisting**, and the caller treats it as success. `[VERIFIED at source]`; real on
   its own terms; **does not fit as the FIRST failure**. Council gate applies. NOTES §6.
3. **A source comment that does not reconcile:** it records "26 of 433 public tables" (2026-08-10);
   the same patterns select **86 of 457** today. Growth or method difference — unresolved, and the
   cap-headroom argument in (1) depends on it. NOTES §10.

## ⚠ Traps this lane has paid for — read before measuring anything

- **Post-roll quiet is NOT the fix working.** Entry condition by day: `0/448, 0/232, 0/1241, 1/1302,
  0/481, **30/1436 (08-17)**, 0/1603 (08-18), 0/385 (08-19)`. **Thirty of thirty-one are one day**,
  and 08-18 was already zero on the *unfixed* binary. Six of eight days are zero. A quiet period
  means nothing when the baseline is also quiet.
- **Retention is PER STATUS.** `SELECT min(created_at) FROM orchestration_states` reads **2026-07-13**
  — five weeks — because `CANCELLED` rows are never pruned. Grouped by status it holds two days. It
  errs *reassuringly*. Also in `LANDMINES.md`.
- **Any check against a diagnosis bundle for a string YOU authored is blind** — the symptom is quoted
  into the bundle verbatim. `LIKE '%awaited_requests%'` and `LIKE '%next_call_registered%'` both
  return true on material that is just your own text. Discriminate on the renderer's syntax
  (`awaited_requests(` **with the parenthesis**) or on output values. Caught twice in one evening.
- **Resolve a council verdict BY CORRELATION, never by recency.** `ORDER BY created_at DESC LIMIT 1`
  on `doc_notes` returned a different lane's APPROVED note. Use `AND body LIKE '%<your corr>%'`.
- **`row_cap` is 200** on data requests. An unfiltered dump `ORDER BY orchestration_id` returns the
  lexicographically-first rows only — a deterministic, non-representative slice. Filter and order by
  `sent_at`.
- **The freeze time is `last_activity`, never `updated_at`** (on a reaped row `updated_at` is when the
  *reaper* wrote, making every wedge read a uniform ~4h26m).
- **A `build provenance` log line that is absent means "scrolled", not "unstamped"** (~4 min retention).
  Use the binary probe, and **always run a control that must be absent**.

## Can 029 be closed? NO

The bar is **fixed AND live**. Part A is both, but Part A was never the wedge. The wedge mechanism is
unexplained; its absence since 08-17 is indistinguishable from six other quiet days; its one observed
burst overlaps a GitHub API incident, which nobody has pulled on.
