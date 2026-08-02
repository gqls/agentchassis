# HANDOFF — bugs_open/174, 2026-08-02 evening — continue here

**Read this first, then `NOTES_seed_scope_relay.md` (bottom up) and `PLAN_2026-08-02_seed_scope_relay.md`.**

## State in one paragraph

`bugs_open/174` (the diagnosis dispatch loop silently dropping `seed_scope`) is
**fixed, council-approved, and LIVE on chassis v1.0.1229** — both the config half
(migration 289) and the Go half (`f51acb2bb`), pod-verified on both replicas with
a positive and a negative control. A real seeded diagnosis was fired through the
**default** path (no `DISPATCH=1`) and **the seed arrived**. The only thing
outstanding when this handoff was written was the last assertion of that run:
`scope_source` must read `seed`, and the bundle's `## In-scope code` must contain
the two symbols named. **Once that is confirmed, move the ticket to
`bugs_closed/` and nothing else is owed.**

## The one command that finishes this

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
SELECT collected_data->'assembled'->>'scope_source', current_step, status
  FROM orchestration_states
 WHERE correlation_id='12fdf121-04e8-431d-9245-38767971e9ea'
   AND owner_agent_type='diagnose-agent';"
```

- **`seed`** → the fix is proven end to end. Close the ticket (below).
- **`code_results`** → the seed was still not used. That would mean gate 3 is not
  actually closed; re-read `ExtractStringListHelper` and the pod-grep.
- The bundle itself: `SELECT body FROM diagnosis_artifacts WHERE
  correlation_id='12fdf121-04e8-431d-9245-38767971e9ea' AND kind='bundle' ORDER BY
  created_at DESC LIMIT 1;` — its `## In-scope code` must name
  `ExtractStringListHelper` and `DiagnoseAssembleBundleAction`. **Assert this, not
  just the field**: field-present and scope-used come apart, which is the entire bug.

Run correlations for this verification: intake `1a35f000-a95d-46cd-b4ea-8e61bff7bcea`,
run `12fdf121-04e8-431d-9245-38767971e9ea` (the run correlation is the one artifacts
are written under — they differ).

## To close

```bash
git mv bugs_open/174_HANDOFF_2026-08-01_dispatch_loop_drops_seed_scope_so_a_targeted_diagnosis_silently_becomes_an_untargeted_one.md \
       bugs_closed/174_HANDOFF_2026-08-01_dispatch_loop_drops_seed_scope_so_a_targeted_diagnosis_silently_becomes_an_untargeted_one.md
# LANDMINE: name BOTH paths on the commit, or a pathspec commit ships a COPY:
git commit bugs_open/174_*.md bugs_closed/174_*.md -m "close(174): ..."
# verify at HEAD, not at the tree (`ls` cannot tell you — the file is gone from disk either way):
git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 174   # exactly ONE line
```
Also flip the `016b` §10 row (line ~4927) from "FIXED, awaiting the roll" to CLOSED,
and update `MEMORY.md`'s bug line if one was added.

## What was actually wrong (the part worth carrying forward)

The ticket named **one** gate. There were **three**, and its own fix candidate 1
would have failed silently:

1. `claim_item`'s SQL `RETURNING` is **also an allow-list** and never projected the
   key — so `claimed.seed_scope`, the source path the ticket proposed, did not exist.
   An optional mapping key that does not resolve is skipped at Info level.
2. `call_handler`'s `input_mapping` — the gate the ticket named.
3. **Type.** `QueryDatabaseAction` stringifies every `[]byte` a column scan returns,
   so a jsonb value arrives as the *string* `["a","b"]`; `ExtractStringListHelper`
   returned nil for a string, indistinguishable from "nothing supplied".

**Gate 3 was confirmed in production**, not merely argued: the live orchestration
row shows `jsonb_typeof(input_data->'seed_scope') = string`. The config half alone
would not have fixed the bug.

## Deliverables, all committed

| commit | what |
|---|---|
| `f51acb2bb` | Go: `ExtractStringListHelper` JSON-text arms + `scope_source` provenance + tests; migration (then numbered 285) |
| `10789dfe6` | `config-key-audit --relay-gaps` + `scripts/audit-relay-gaps.sh` + tests; migration renumbered 285→**289** |
| `b34717c18` | standing five, 2 LANDMINES (synced), 016b §9, WFA-007, 2 WRONG_CALLS; `SafeUnmarshal` reuse |
| `e500f00b2` | disclosure of a LANDMINES same-file passenger |
| `63f4424d1` | wired `--relay-gaps` into the CLI; corrected the 016b §10 row in place |

Council: **APPROVED round 1**, corr `081d98b3-75e1-4926-a17a-b0c72e5ccece`, 6
advisory objections, 4 answered with work (see NOTES § "what the council caught").

## Open items — none blocking, all deliberate

1. **Two dispatcher relays are reported UNCOVERED** by `--relay-gaps`:
   `report-dispatch-loop.call_handler` and `build-pipeline-trigger.call_dispatch`.
   **Do not register them to make the list empty.** Their callees resolve at
   runtime and their handlers' contracts have not been read; registering them
   unread asserts something nobody has checked — the exact state 174 was in.
   Registering one properly means reading its handler's `input_contract` first.
2. **`QueryDatabaseAction`'s jsonb stringification is NOT fixed**, deliberately.
   Measured: **exactly one** live `query_database` step projects a JSON-typed
   value (the one this fix added), so there are zero currently-affected consumers
   — a prospective trap, not a live defect. Recorded in `LANDMINES.md`. If someone
   later wants to fix it properly, the blast radius must be re-measured by
   **projection position** (`-> '...' AS alias`), never by "the query mentions json".
3. `MEMORY.md` was at 22,925 bytes of a hard 25,000 cap when this lane finished;
   no line was added for 174. If you add one, merge it into an existing family
   line rather than appending.

## Traps this lane hit, so you don't

- **`min(jsonb)` does not exist.** The migration failed its first run on exactly
  that. It rolled back cleanly; verify that before retrying rather than assuming.
- **`replace()` on a query string is a silent no-op on a miss.** Assert the
  projection is present afterwards; the UPDATE's row count counts rows *touched*.
- **`validation.WalkSteps` qualifies step paths** (`steps.call_handler`) while a
  registry names them as authored. Compare through `stepName`, not equality — this
  made the detector's first live run match nothing.
- **`config-key-audit` falls through to its DEFAULT report for an unknown
  argument** (valid JSON, exit 0). `audit-relay-gaps.sh` asserts the report's own
  keys and exits 2 if absent. Do not remove that guard.
- **The 016b §10 table already had a 174 row.** I nearly added a second, and my
  first insert corrupted an unrelated line because I regex-matched over a
  truncated slice of the file. Edit that table against the whole file.
- **Two same-file collisions with live sessions in one hour** (`main.go`,
  `LANDMINES.md`) and a **migration-number collision** (another session took 285
  eight minutes after me; I renumbered mine to 289). Re-check `git status` and the
  live config immediately before acting — this tree moves in minutes.
