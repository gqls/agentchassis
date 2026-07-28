# HANDOFF — architecture seat, continue here (2026-07-28, morning)

**COLD-START ENTRY POINT. Supersedes `HANDOFF_2026-07-27c_continue_here.md`**,
which said layer 1 was built but not live. **It is live, proven, and complete.**
Go back to `c` only for its verification recipe; back to `b`/the original for
their §5 landmine lists, which are unchanged and still the most expensive part of
this directory.

Prose: `SUMMARY_2026-07-28_the_seat_can_see.md` (current state, written to be read
aloud), `README_where_we_are.md` (owner's plain-English log, append-only),
`NOTES_architecture_seat.md` (technical log + every misstep),
`DECISIONS_open_for_owner_2026-07-26_architecture_seat.md` (D1–D11),
`RUNBOOK_architecture_seat.md` (every command with its gotcha).

---

## 1. Nothing is waiting on the owner, EXCEPT one thing that is not a code question

Design settled 2026-07-27: **D7(a)** guardian keeps its veto · **D7(b)** do NOT
narrow the guardian · **D9** do NOT add a second forward seat · **D11** seats must
be able to look things up. Do not reopen D7(b) or D9 on a reversal trigger firing;
after a ruling those are the evidence needed to *ask* to revisit, a higher bar.

> ## ✅ RESOLVED 2026-07-28 12:47 — the branch was pushed, and the ref is now PINNED
>
> The staleness is gone: `ref='086_experience_loop'`, commit `d98010e8b`, **0
> commits behind the branch tip**, **4,992 rows / 4,992 bodies**. Another session
> also shipped the `ref` and `commit_time` columns (108 candidate 1), so the
> freshness verdict now keys on the indexed COMMIT's own date, not the row clock.
>
> **Migration 252 pins the ref deterministically** — `code-index-refresh`'s
> `pre_query` is now `SELECT '086_experience_loop'::text AS ref`. Proven through
> the REAL scheduled path (not a manual dispatch): scheduler fired 12:46:27,
> `index-orchestrator` → `code-indexer`, both `COMPLETED`, ref stored as pinned.
>
> **⚠ REVERSAL TRIGGER — CHANGE THE LITERAL IN 252 TO 'main' AS PART OF THE
> MERGE.** Not after. Two reasons, and the second is the quiet one: the pin will
> keep indexing a branch that has stopped moving while main becomes the live tree;
> and the pre_query it REPLACED inferred the ref from `^[0-9]{3}_`-shaped refs, so
> `main` would never have matched it — the day of the merge that old query would
> have gone dry and **the refresh would have stopped silently**, since a pre_query
> returning no rows makes the scheduler skip the task entirely rather than fall
> back (`cmd/scheduler/main.go:198-212`).
>
> **Migration 251 is INERT — do not copy its pattern.** It put a static `ref` in
> the task's `input_data`. That can never be read: the pre_query result is the
> OVERLAY in `mergeJSON` (`:216`, `:480`) so it overwrites a static key of the same
> name, and the no-rows path skips the task rather than falling back. Committed as
> applied rather than rewritten, so its checksum still matches the ledger.
>
> ### The original ask, kept because the reasoning still applies
>
> `code_symbols` mirrors the **last pushed tip**. `origin/086_experience_loop` has
> sat at `e19aa5d10` since 2026-07-24 while local HEAD moved ~955 commits. **A
> reindex cannot close that and provably does not** — one ran successfully at
> 07:04 on 07-28 and the distance did not move by one commit.
>
> This is not theoretical. At **07:07:27 on 07-28** a live diagnosis asked whether
> `RepairPageLinks` exists and was told *"the query was RUN and matched none"*. It
> exists (`platform/orchestration/datahelpers/link_repair.go:139`). It is absent
> only from the snapshot. The banner said *"refreshed 17h ago"*.
>
> Every council review and every diagnosis is currently reasoning about a
> fortnight-old tree while being told the index is fresh. **Raised with the owner
> in `README_where_we_are.md`; it is a shared-branch decision, not ours to take.**

## 2. D11 layer 1 — DONE. Live, proven, and measured.

| | |
|---|---|
| implementation | **`37f7deff9`**, trailer `Council-Reviewed: 18fe4035-4fa6-4079-ab44-8541d6e58944` |
| perf follow-up | **`a4f06f83a`** (the COALESCE fix — no trailer, correctly: it was not the approved plan) |
| migrations | **243** (body column + trigram index) and **244** (doc_notes trail) — both APPLIED |
| live on | chassis **`v1.0.1182`** (layer 1 landed on `v1.0.1180`) |
| bodies | **4,535 of 4,535**, 0 empty strings |
| `content_hash` drift | **0** — nothing was silently re-embedded |
| `stop_reason` hits | **0 → 6** (the contract's own example, working for the first time) |
| query plan | **BitmapOr across BOTH trigram indexes** (was a Seq Scan at 125.9 ms; now 5.5 ms warm) |
| indexer log | `bodies_sliced=4536 slice_errors=0 file_read_errors=0` |

**What it changed:** `code_symbols.body` holds each symbol's source text, sliced at
index time by the now-exported `internal/analysis.SliceLines` from the LOCAL
checkout. `diagnose_code_lookup` kind=`content` searches `body OR content`, marks
each hit `[body]` or `[decl]`, and excerpts **around** the match. An empty result
renders as an ANSWER — *"answered: 0 rows — searched 4,535 indexed symbols"* — in
**both** lanes that answer code checks (`diagnose_code_lookup` and
`diagnose_load_runtime`).

**Re-verify at any time** (all five must be green; check 5 must show a **BitmapOr**,
a Seq Scan is a regression):
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db \
  -f - < docs/agent_docs/sql_for_agents/243_code_symbols_body_column_VERIFY.sql
```

## 3. What is open, in order

1. **`bugs_open/108` candidate 1 — freshness by COMMIT DISTANCE, not clock.**
   **This is the next job, and its priority was INVERTED by layer 1 landing** (see
   §4). The small enabling change: the indexer already receives an explicit `ref`
   and knows what it fetched, but **does not store it**. Recording `ref` beside
   `commit_sha` makes "N commits behind `<ref>`" computable at read time without
   giving the shared chassis pod a GitHub token it deliberately does not hold.
   Note the constraint already filed in 108: the `code-index-refresh` scheduled
   task's `input_data` is `{repo, owner, language}` — **there is no ref parameter**,
   so any induced test needs that added first.
2. **The seat still has 0 reviews, and waiting will not change that.**
   `review_architecture` exists only on `feature-designer`, which refuses anything
   without an owner-approved `capability_gap` spec (`check_spec_approved` wants
   BOTH `owner_approval` and `code_pointers`). 5 such items, 2 approved, **both
   owned by other threads**. **A zero here is a rate limit, not a fault** — do NOT
   manufacture one by firing at another thread's ticket.
3. **When it speaks, read it honestly.** `./scripts/council-adoption-report.sh` §5
   exists to say the seat is producing confident noise and should be pulled. That
   is a real option, not a formality.
4. **Markdown is still unreachable.** 0 of 4,535 rows are markdown, so
   `WRONG_CALLS.md`, `/bugs_open/` and the concept register are invisible to every
   seat. Layer 1 built the mechanism; markdown additionally needs the `kind` CHECK
   constraint relaxed — its own migration. Rank by the concept register's
   rediscovery-frequency signal rather than inventing one. **Do this AFTER 108
   candidate 1**, for exactly the reason in §4.
5. **D11 layer 3 — the dynamic round-trip.** A seat emits `code_checks` and gets
   answers *next round*, so it cannot look while reasoning. `[UNSCOPED]`; deserves
   its own RFC.
6. **`council-gate` still gets no code answers at all.** Its missing `code_lookup`
   is DELIBERATE (`099_SYNC_gate_roster.py:24-29` — it serves the blind reproposer,
   which the gate has no equivalent of), so the fix is surfacing code results into
   the verdict note, not bolting on a reproposer-shaped step. In `bugs_open/108`.

## 4. THE FINDING THAT MATTERS MOST — read this before ordering any further work

**Fixing `108` defect B made defect A worse.** The empty-answer rendering shipped
with layer 1 deliberately replaced `"(no symbol matches in the index)"` with
`"answered: 0 rows — the query was RUN and matched none; this is not an unanswered
question."` That is the correct fix for the empty-vs-unanswered confusion — **and
it is a stronger denial when the corpus is stale.** The old wording at least kept
the word *index* in the sentence.

> **Two fixes that each increase honesty can combine to increase HARM, when one
> raises confidence in a signal the other has not yet made correct.** Confidence and
> correctness are separate axes; repairing confidence first is a regression wearing
> a fix's clothes, and it passes every test because the sentence it emits is exactly
> the sentence you wanted.

Consequence for this workstream: **land the fix that makes the DATA correct before
the one that makes the PRESENTATION confident.** That is why §3 item 1 outranks
item 4. Full write-up in `bugs_open/108`; transferable pattern in `016b` §9.

## 5. Landmines — new since the `b`/`c` handoffs

- **THE SEED IS NOT THE SYSTEM.** The approved plan asserted *"the indexer is
  already walking the file"*. It is not — `flattenSymbols` walks a JSON-decoded
  `analysis.Output` (line spans, no source text), and the sketch sliced bytes out
  of a `fileSrc` that exists nowhere in scope. Twelve seats approved it. It works
  ONLY because the LIVE `code-indexer` workflow uses `analyse_repo_local`, which
  leaves a real checkout in the pod; the repo's own seed
  `118_code_indexer_for_analyser.sql` **still shows `request_repo_analysis`**, under
  which every body would be NULL and the feature would deploy looking finished and
  doing nothing. **Read `agent_definitions`, not the seed.**
- **`content_hash` does NOT change when only a BODY changes.** It covers
  `composeSymbolContent` output — kind + symbol + signature + doc + path. So there
  is no cheap test for "is this stored body current", which is why the upsert
  assigns `body` plainly and must never `COALESCE` it onto the previous value: that
  would preserve text contradicting the `line_start`/`line_end` written from
  `EXCLUDED` in the same statement.
- **`COALESCE(col, …)` in a predicate disqualifies `col`'s plain-column index.**
  Cost us a Seq Scan at 125.9 ms on the only query path the new index serves.
  Removing it is semantically identical here: `WHERE` discards NULL exactly as it
  discards false.
- **THREE ways a check passed without checking, all in this one piece of work:**
  (a) the pod-grep marker the council submission proposed
  (`idx_code_symbols_body_trgm`) is **SQL-only and appears nowhere in the binary** —
  it would have read 0 from a correct deploy; (b) comparing the two predicates on
  live rows returns 6 and 6 because there are **0 NULL bodies**, so the
  distinguishing case cannot occur — test it over a `VALUES` list instead; (c) the
  VERIFY script's `EXPLAIN` stayed pinned to the **old** predicate, so after the
  roll it would have shown a Seq Scan and read as a failed fix. **When a fix changes
  the shape of the thing asserted, the assertion is part of the change.**
- **The migration runner has no single-file apply.** `--apply` takes EVERY pending
  file, including other sessions' half-finished work. Use
  `MIGRATIONS_DIR=<dir containing only your file>` — same ledger, same probe.
- **Migration numbers move under you.** 243 was approved as 241; two sessions took
  241 and 242 within four hours. Re-read from `ls` immediately before writing.
- **Best pod-grep marker is a string your change DELETED.** It flips the other way,
  so a stale image cannot pass both. Pre-validate it against the OLD image before
  the new one exists — `v1.0.1182` gave `1/0` where `v1.0.1180` gave `0/1`.

## 6. Where the measurement stands

Baseline (pre-cutover **13:44:56**, from `agent_definitions.updated_at`, never
guessed): guardian 210 reviews, 90 invoked the stability preference, **2 of those
90 cited precedent (2.2%)**. The old "6 of 90" was structurally wrong (two
independent `FILTER`s, never a subset) and is corrected everywhere.

**The metric is wrong in BOTH directions and the script says so** — `%deflect%`
matches the bare word and the new prompt itself says "deflected upward", so a seat
echoing its instructions scores a citation; and the 14:18 guardian reasoned
correctly about recurrence without quoting a past report and scored zero.
**At low n, read the review text.**
