# HANDOFF — architecture seat, continue here (2026-07-27, late evening)

**COLD-START ENTRY POINT. This supersedes `HANDOFF_2026-07-27b_continue_here.md`**,
whose §4 item 1 says *"LAYER 1 IS APPROVED — build it. This is the next job."*
**It is built.** Go back to `b` for its §5 landmines (still accurate, and the two
new ones below are additions, not replacements) and to `HANDOFF_2026-07-27_continue_here.md`
§5 for the Go-contract landmines, which are the most expensive ones and unchanged.

Prose, if you want it: `README_where_we_are.md` (owner's plain-English log,
append-only, newest at bottom — the last entry is this work),
`NOTES_architecture_seat.md` (technical log; last entry is this work),
`RUNBOOK_architecture_seat.md` (every command with its gotcha),
`DECISIONS_open_for_owner_2026-07-26_architecture_seat.md` (D1–D11),
`SUMMARY_2026-07-27c_rulings_and_the_seat_that_cannot_see.md` (current state as of
the evening — **still broadly right; layer 1 moving from approved to built is not
a new inflection, so no new SUMMARY was written**).

---

## 1. The design is settled. Nothing waits on the owner.

Unchanged from handoff `b` §2 and worth not re-litigating:
**D7(a)** the guardian keeps its veto · **D7(b)** do NOT narrow the guardian ·
**D9** do NOT add a second forward seat · **D11** seats must be able to look
things up (the honesty caveat is an interim, not a destination).

Do not reopen D7(b) or D9 on a reversal trigger firing — after a ruling those are
the evidence you would need to ask the owner to revisit, which is a higher bar.

## 2. THE ONE THING OWED, and it is a verification, not a build

**Layer 1 is BUILT, COMMITTED, and IN THE IMAGE. It is not yet LIVE.**

| | |
|---|---|
| implementation commit | **`37f7deff9`**, carries `Council-Reviewed: 18fe4035-4fa6-4079-ab44-8541d6e58944` |
| working docs commit | `53ff8c80a` (NOTES/RUNBOOK/README/WRONG_CALLS + migration 244) |
| migration 243 (`body` column + trigram index) | **APPLIED, live** — 0 hash drift on 4,535 rows |
| migration 244 (doc_notes trail) | **APPLIED, live** — 4 rows, defect-then-fix per action |
| chassis image `v1.0.1179` | **BUILT and contains the change** (verified by grep, below) |
| running pod | still **`v1.0.1177`** as of 21:20 BST — **the roll had not happened** |
| `code_symbols.body` | **0 of 4,535 populated**, correctly: the column is live, the Go that fills it is not |

**That the image contains it was verified DISCRIMINATINGLY, and the marker the
council submission proposed would NOT have worked.** It said to grep the pod for
`idx_code_symbols_body_trgm` — that string exists only in the migration SQL and
appears nowhere in the Go, so it would have returned 0 from a correct deploy.
Use these instead (run against the running pod after the roll):

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
kubectl -n ai-persona-system exec $POD -- sh -c '
  strings /app/agent-chassis | grep -c "sliced symbol bodies from the local checkout"  # expect 1
  strings /app/agent-chassis | grep -c "BODIES ARE NOT INDEXED"                        # expect 1
  strings /app/agent-chassis | grep -c "no content matches in the index"               # expect 0 — REMOVED by this change
  strings /app/agent-chassis | grep -c "zzz_string_that_appears_nowhere_zzz"           # expect 0 — never existed
'
```
The third line is the good one: it is a string this change **deleted**, so it
flips the other way and distinguishes the new binary from a rebuild of the old.
Measured on the local images: `1179` → 1/1/0/0, `1177` → 0/0/1/0.

### After the roll, in order

1. **Wait ~300s** after the chassis pod restarts before dispatching anything —
   a spawn inside that window is silently dropped.
2. **Pod-grep** with the block above.
3. **Trigger a reindex** (the `code-index-refresh` scheduled task runs on a 24h
   cadence; the index was last refreshed 2026-07-27 13:36 UTC, so it is fresh but
   body-less). Watch for `with_body` in the action result — it now sits beside
   `upserted` precisely so a run that indexed everything and sliced nothing is
   legible without querying the column.
4. **Re-run the VERIFY**, which was already run before:
   ```bash
   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
     psql -U clients_user -d clients_db \
     -f - < docs/agent_docs/sql_for_agents/243_code_symbols_body_column_VERIFY.sql
   ```
   Expected change: check 3 `bodies_populated` 0 → ~4,535, check 4
   `stop_reason_hits` 0 → >0. **Checks 1, 2 and the negative control must still
   be 0** — check 1 is the one that matters (content/content_hash byte-identical,
   i.e. nothing was silently re-embedded).
5. **Settle `guardian`'s open low objection with check 5's `EXPLAIN`.** With
   `body` NULL it already shows a **Seq Scan** (cost 547) — but that measurement
   **decides nothing** while there is no text to plan over. Against populated
   bodies: if it still seq-scans and the cost is material, switch the predicate to
   a single expression index on `(COALESCE(body,'') || ' ' || content)`. If it is
   fast, say so with the number and leave the OR alone.
6. **Record the real `max(length(body))`.** I have asserted nothing about body
   size and neither should you until it is measured — the excerpt cap (400 chars)
   and row cap (40) bound what reaches a prompt regardless, so this is about index
   size, not prompt safety.

## 3. What layer 1 actually changed, in one place

- `code_symbols.body` (nullable text) + `idx_code_symbols_body_trgm`.
- `index_code_symbols` slices each symbol's body from `out.Root` — the LOCAL
  checkout — one file read per file, no re-parse, no extra fetch. On any read or
  slice failure the body stays **NULL, never `""`**.
- `diagnose_code_lookup` kind=`content` searches `body OR content`, marks each hit
  `[body]` or `[decl]`, and takes the excerpt **around the match**.
- An empty result renders as an **answer** — *"answered: 0 rows — searched 4,535
  indexed symbols"* — in **both** lanes that answer code checks
  (`diagnose_code_lookup` and `diagnose_load_runtime`), plus a scope line stating
  whether bodies are indexed at all.
- `internal/analysis.sliceLines` → **exported** `SliceLines`. One function owns
  the `[start,end]` inclusive 1-indexed convention, so a stored body and a
  freshly sliced one cannot silently differ.

## 4. NEW LANDMINES — the two that cost the most today

- **THE SEED IS NOT THE SYSTEM, and this one nearly shipped an inert feature.**
  The approved plan asserted *"the indexer is already walking the file, so this
  needs no new file I/O pass"*. **It is not.** `flattenSymbols` walks a
  JSON-decoded `analysis.Output` — line spans and names, **no source text** — and
  the plan's sketch sliced bytes out of a `fileSrc` that exists nowhere in scope.
  Twelve council seats approved it. It works **only** because the LIVE
  `code-indexer` workflow uses `analyse_repo_local`, which leaves a real checkout
  in the pod; **the repo's own seed `118_code_indexer_for_analyser.sql` still
  shows `request_repo_analysis`**, under which every body would be NULL and the
  feature would deploy looking finished and doing nothing. Read
  `agent_definitions`, not the seed. Full entry in `WRONG_CALLS.md`.
- **`content_hash` does NOT change when only a function's BODY changes.** It
  covers `composeSymbolContent` output — kind + symbol + signature + doc + path.
  So there is no cheap test for "this stored body is still current", which is why
  the upsert assigns `body` plainly and never `COALESCE`s it onto the previous
  value: the "safe" form would preserve text contradicting the
  `line_start`/`line_end` written from `EXCLUDED` in the same statement.
- Two smaller ones now in the RUNBOOK: **the migration runner has no single-file
  apply** (`--apply` takes every pending file, including other sessions'
  half-finished work — use `MIGRATIONS_DIR=<dir with just your file>`), and
  **the migration number moves under you** (243 was approved as 241; two sessions
  took 241 and 242 within four hours).

## 5. Everything else that was open, unchanged

1. **The seat still has 0 reviews, and waiting will not change that.**
   `review_architecture` exists only on `feature-designer`, which refuses anything
   without an owner-approved `capability_gap` spec. 5 such items, 2 approved,
   **both owned by other threads**. Its first review arrives on the colour-fixer
   thread's round 4 or on a newly approved spec. **A zero here is a rate limit,
   not a fault** — do NOT manufacture one by firing at another thread's ticket.
2. **When it does speak, read it honestly.** `./scripts/council-adoption-report.sh`
   §5 exists to say the seat is producing confident noise and should be pulled.
   That is a real option.
3. **D11 layer 3 — the dynamic round-trip.** A seat emits `code_checks` and gets
   answers **next round**, so it cannot look while reasoning. `[UNSCOPED]`; it
   deserves its own RFC, not folding into layer 1 or 2.
4. **Markdown is still unreachable.** 0 of 4,535 index rows are markdown, so
   `WRONG_CALLS.md`, `/bugs_open/` and the concept register are invisible to every
   seat. Layer 1 built the mechanism; markdown additionally needs the `kind` CHECK
   constraint relaxed — a separate migration, **after** layer 1 is proven live.
   Rank by the concept register's own rediscovery-frequency signal rather than
   inventing one.
5. **`council-gate` still gets no code answers at all**, and its verdict note
   cannot distinguish "searched, found nothing" from "nobody ran the query". The
   missing `code_lookup` is **deliberate** (`099_SYNC_gate_roster.py:24-29` — it
   serves the blind reproposer, which the gate has no equivalent of), so the fix
   is surfacing code results into the verdict note, not bolting on a
   reproposer-shaped step. Recorded in `bugs_open/108` as owed.
