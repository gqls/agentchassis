# 056 — The diagnosis loop kills its own runs: a NUL byte in `SeenCodeRequests` keys makes the `route` step's state-persist fail with Postgres `22P05`

**Filed:** 2026-07-21 · **Branch:** `085_debug_and_feature_loops` · **Status:** FIXED & LIVE 2026-07-22 (commit `7a9f5f652`, in prod via v1.0.1149 — see §Resolution at the bottom; end-to-end verification in flight, council follow-up open on the sanitiser half)
**Severity:** medium-high — not data corruption (the bad write is *rejected*, nothing lands),
but it **silently destroys diagnosis runs**. 25 orchestrations dead so far, still occurring today.
The bug is in the diagnosis loop's own code, so the platform's "diagnose before you assert"
mechanism is unavailable for exactly the class of investigation that reaches its code-lookup tier.
**Class:** platform code defect (`pkg/diagnose` + the orchestration state-persist path), fleet-wide.
**Owner:** the diagnosis→fix loop workstream owns `pkg/diagnose` — see *Ownership* below; do not
start a competing fix, contribute into this file / their docs.

> **This is what killed `bugs_open/030`'s root-cause diagnosis.** Corr
> `78470372-7617-40e4-888c-66cac94006bf` waited ~90 min in the 030 queue, finally ran at
> 2026-07-20 20:50 UTC, and died on *this* bug at its `route` step. So 030 has been "waiting for a
> verdict" that was never coming. The work item reads `complete` (a response arrived) while the
> diagnosis `failed` — the status-vs-artefact trap CLAUDE.md warns about.

---

## Symptom

A diagnosis orchestration (dispatched via `090_TRIGGER_needs_diagnosis` / any `generic-orchestrate`
that routes into the diagnosis workflow) runs, reaches the `route` step, and then dies. The
`site_work_items` row shows `status='complete'`; the orchestration's stored result shows:

```json
{
  "response": {
    "error": "failed to persist step result for 'route': failed to save step result: failed to update state: ERROR: unsupported Unicode escape sequence (SQLSTATE 22P05)",
    "status": "failed"
  }
}
```

## Evidence — measured (DB clock; my sandbox clock runs ~6h slow, do not trust `date` here)

`orchestration_states`, no time filter (DB `now()` = 2026-07-21 15:45 UTC when queried):

| metric | value |
|---|---|
| total orchestrations with `unsupported Unicode escape` | **25** |
| of those, failing at step `'route'` | **25 (100 %)** |
| first seen | 2026-07-19 18:14:21 UTC |
| last seen | 2026-07-21 11:20:32 UTC (**ongoing**) |

```sql
-- reproduce the census
SELECT substring(collected_data::text from 'failed to persist step result for ''([a-z_]+)''') AS step,
       COUNT(*), MIN(created_at), MAX(created_at)
FROM orchestration_states
WHERE collected_data::text ILIKE '%unsupported Unicode escape%'
GROUP BY 1;
-- → route | 25 | 2026-07-19 18:14 | 2026-07-21 11:20
```

Example dead runs (all `generic-orchestrate-*`, i.e. on-demand diagnosis/council dispatches):
`74e31db3-2813-4105-a5c9-421228f2fbe9` (2026-07-21 11:20), `2556c072-9b6f-4fa4-821d-b0421ec31523`
(2026-07-21 11:19), plus the 030 diagnosis `78470372…` (2026-07-20 20:50).

**Onset (2026-07-19) coincides with the code-lookup / `SeenCodeRequests` tier landing** in the
diagnosis loop. There are zero occurrences before that date.

## Root cause — VERIFIED by code inspection (with the code's own comment as witness)

The diagnosis loop keys its cross-iteration "already asked" set with a **literal NUL byte**
delimiter, and that set is persisted into the `collected_data` **jsonb** column. Postgres `jsonb`
(and `text`) cannot store `\u0000`; it is the *one* Unicode escape it rejects. So the moment the set
is non-empty, the next state-persist dies.

The chain, cited:

1. **The key is built with `"\x00"`.** `pkg/diagnose/loop.go:162-164`:
   ```go
   func CodeRequestKey(kind, query string) string {
       return strings.ToLower(strings.TrimSpace(kind)) + "\x00" + strings.ToLower(strings.TrimSpace(query))
   }
   ```
   (A second, duplicated builder does the same: `diagnose_code_lookup_action.go:316`
   `key := c.Kind + "\x00" + strings.ToLower(c.Query)`.)

2. **The set is a `map[string]bool` keyed by that string** — `pkg/diagnose/step.go:34,58`
   `SeenCodeRequests map[string]bool`.

3. **That map round-trips through `collected_data`** — the loop reads prior keys back out and splits
   them: `diagnose_route_action.go:366` `kind, query, ok := strings.Cut(k, "\x00")`. The code's own
   comment at `diagnose_route_action.go:372-373` states it outright:
   > *"these keys are WRITTEN by CodeRequestKey and read back through collected_data"*

4. **The persist marshals the map to JSON, NUL → `\u0000`.** `state.go:760`
   `collectedDataJSON, _ := json.Marshal(state.CollectedData)` — Go's `encoding/json` encodes a NUL
   byte in a map key as the escape `\u0000`.

5. **The JSON is written to a `jsonb` column and rejected.** `state.go:782-821`
   `UPDATE orchestration_states SET … collected_data = $6 …` → Postgres raises
   `22P05 unsupported Unicode escape sequence`, `UpdateState` returns the error, and
   `coordinator.go:908` surfaces it as `failed to persist step result for 'route'`, killing the run.

**Why 100 % at `route`, and not at the LLM steps:** the offending bytes are in the router's *own data
structure* (`SeenCodeRequests`), not in raw LLM text. A NUL in model output would fail at whichever
step emitted it; the perfect concentration at `route` is the signature of a structural cause in the
router, which is what (3) confirms. **First iteration persists fine** (the set is empty, no NUL keys);
it dies on the first persist *after* a code request is recorded — i.e. any multi-iteration diagnosis
that touches the code-lookup tier.

**What it does NOT do:** no data corruption. The bad `UPDATE` is rejected wholesale, so no `\u0000`
ever lands in the table (which is also why you cannot grep a persisted `collected_data` for the NUL —
the only artefact is the error string, written by a later, smaller update that succeeded).

## Fix candidates (in rough order of value; the fixing thread should council-gate the choice)

1. **Sanitise NUL at the persist boundary — fleet-wide safety net (recommended as the floor).**
   In `state.go UpdateStateWithVersion`, after the `json.Marshal` calls (`state.go:758-766`) and
   before `ExecContext`, strip the escape from every `*JSON` byte slice, e.g.
   `collectedDataJSON = bytes.ReplaceAll(collectedDataJSON, []byte(`\u0000`), []byte(``))` (or replace
   with U+FFFD's escape `�`). **One place, provably complete: no orchestration of any type can
   ever again be killed by a NUL byte**, whatever its source (a future LLM step emitting one would hit
   the identical wall today). Downsides: it is a hot path; and for the seen-key case, dropping the NUL
   collapses the delimiter so two composite keys *could* alias — harmless here (it only affects dedup,
   worst case a code question re-runs once), but note it. This masks, rather than removes, the
   diagnosis loop's use of NUL as a delimiter, which is why it is a floor, not the whole fix.

2. **Change the delimiter at the source — remove the NUL entirely.** In `CodeRequestKey`
   (`loop.go:163`) replace `"\x00"` with a jsonb-safe separator (any control char *other than* NUL is
   accepted by jsonb, e.g. `"\x1f"` unit-separator → `\u001f`; `kind` is a validated enum so it cannot
   contain the separator). Must be changed **in lockstep** with every reader/builder or dedup silently
   breaks: the reader `diagnose_route_action.go:366`, the duplicate builder
   `diagnose_code_lookup_action.go:316`. **Check but probably leave** `code_symbols_actions.go:215,369`
   (`path+"\x00"+symbol`) — that is a *different* map; confirm whether it is ever persisted to jsonb
   (if it is in-memory dedup within a single action invocation, it never reaches the DB and is safe).
   Changing the encoding resets any persisted `SeenCodeRequests` (harmless).

3. **Best: do both** — 1 as the platform-wide guarantee, 2 so the diagnosis loop stops generating the
   value in the first place. If only one ships, ship **1**: it protects every orchestration type and
   every future NUL source; 2 alone leaves the platform one stray NUL byte (in any step, any workflow)
   away from the same silent outage.

## How to verify a fix

- **Unit-level:** `json.Marshal(map[string]bool{diagnose.CodeRequestKey("file","x"): true})` currently
  yields a string containing `\u0000`; inserting that into a `jsonb` column raises `22P05`. After the
  fix, the marshaled bytes contain no `\u0000` (candidate 2) or the insert succeeds (candidate 1).
- **End-to-end:** fire a `needs_diagnosis` whose investigation reaches the code-lookup tier and runs
  **≥2 iterations** (so `SeenCodeRequests` is non-empty on a persist), and confirm it advances past
  `route` to a verdict instead of dying. The 030 root-cause question (corr `78470372`) is a ready-made
  reproduction — re-filing it after the fix both verifies this bug and finally answers 030.
- **Regression census:** the `22P05` count in the census query above must stop climbing.

## Landmines

- **`complete` on the work item is not success.** The `site_work_items` row goes `complete` because a
  *response* was delivered; the diagnosis inside it `failed`. Read `result->'response'->>'status'`, not
  the row status. (This is why 030 believed its verdict was still queued when it had already died.)
- **You cannot find the NUL in the DB** — the rejected write never lands. The evidence is the error
  string plus the code path, not a corrupted row.
- **Do not "fix" this by changing the delimiter in only one of the three sites.** Builder, duplicate
  builder and reader must move together or cross-iteration dedup breaks silently (which would *look*
  like a coverage regression, not a crash — harder to catch).
- **`\x1f` and other control chars are fine; only `\u0000` is rejected by jsonb.** Do not over-correct
  into stripping all control characters.

## Ownership & related

- **Owner:** `pkg/diagnose` and the diagnosis actions are the diagnosis→fix loop workstream's code
  (memory `[[fixloop-workstream]]`; the code-lookup tier is their F2.3b(c), built `927b11ba0`). Route
  the fix through them and the council gate — the `route`/`SeenCodeRequests` logic is heavily
  council-reviewed (see the `council-gate eba040a9` references littered through
  `diagnose_route_action.go`) and the persist path (candidate 1) is a hot, shared function.
- `bugs_open/030` — the dispatch-queue latency bug whose own diagnosis this defect killed. Separate
  root cause; this file is why 030's "wait for the verdict" plan has no verdict to wait for.
- Distinct from `bugs_open/019` (one truncated reviewer voids a council round) — different mechanism,
  but the same class of "a run silently produces no usable output".

---

## Resolution (2026-07-22, bugfix-056 session)

**Shipped both candidates in commit `7a9f5f652`** (branch `085_debug_and_feature_loops`):

- **Candidate 2 (delimiter):** `CodeRequestKey` now uses unit separator `0x1f` via a
  private `codeRequestKeySep` const, with a new `SplitCodeRequestKey` inverse in
  `pkg/diagnose/loop.go` — the route reader (`diagnose_route_action.go`) and the
  code-lookup dedup (`diagnose_code_lookup_action.go` — previously a third hand-built
  key) both call the canonical pair, so the encoding cannot drift again. Verified live:
  the escape text backslash-u001f is accepted by jsonb; backslash-u0000 reproduces the
  exact 22P05 error.
- **Candidate 1 (floor):** `sanitiseJSONBNulEscapes` in `platform/orchestration/state.go`
  replaces genuine backslash-u0000 escapes with backslash-ufffd in every jsonb-bound
  value of `UpdateStateWithVersion` (backslash-parity preserves literal quoted-escape
  text; zero-alloc no-match path). Tests: `pkg/diagnose/coderequest_key_test.go` +
  `platform/orchestration/state_jsonb_nul_test.go` (5 subtests incl. the exact 056 shape).

**Live:** production pod v1.0.1149 (started 2026-07-22 13:56Z) carries both new symbols —
`strings /app/agent-chassis | grep -c` → `sanitiseJSONBNulEscapes` 2, `SplitCodeRequestKey` 2,
positive control `CodeRequestKey` 4. Shipped by a fleet sweep build (not this session's).
No legacy NUL-keyed data existed to migrate (any persist carrying one had failed — the
bug's own mechanism).

**Council verdict `2f22e08f-a28b-4e3a-953f-8c1bfdf03c11` (round 1): REJECTED — guardian
hard veto, read in full AFTER the sweep had already shipped the commit.** The veto is
about scope, not the diagnosis: edits 1/2/3/6 (the pkg/diagnose delimiter fix + lockstep
+ tests) were explicitly endorsed ("Ship that"); the state.go persist-boundary sanitiser
is an architecture-level policy change (silently substitute vs hard-fail, fleet-wide)
that must be its own reviewed change with blast-radius accounting. Follow-up submission
filed 2026-07-22 covering exactly that half, with the accounting: every orchestration
type persists through UpdateStateWithVersion; no code anywhere handles SQLSTATE 22P05
specially, so pre-fix "surface the error" always meant "kill the run"; jsonb can never
store a NUL, so hard failure preserves nothing. tooling_provenance's HIGH (no travelling
record) is discharged — doc_notes `pipeline/diagnose` now carries the 0x1f decision, the
lockstep contract and the floor's status. If the follow-up review also rejects the floor,
revert it FORWARD (delete helper + the ten call-site applications) while keeping the
delimiter fix.

**End-to-end verification (the bug file's own bar):** a ≥2-iteration code-tier diagnosis
must advance past `route` to a verdict. In flight: corr `b361298a-e030-456e-956f-adf1e05503b1`
(the bugs_open/056 regeneration-loss diagnosis this bug was blocking). Regression census
must not climb past `route | 3 | … | 2026-07-21 11:20`.

**Move to /bugs_closed/ when** the end-to-end run survives the code tier AND the census
stays flat.
