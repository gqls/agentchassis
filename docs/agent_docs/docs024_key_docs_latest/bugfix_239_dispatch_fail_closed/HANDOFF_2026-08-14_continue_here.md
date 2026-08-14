# dispatch/pool lane — HANDOFF 2026-08-14: 259's fix is IN THE TREE and APPROVED; ONE thing is owed

> **UPDATED later on 2026-08-14 — item (a) below is DONE.** The council verdict landed
> **APPROVED** first round (`0ff072ef-ee02-465e-8a70-f5461c585ec9`, 07:57:26Z; 10 reviewers,
> 8 approve, 2 advisory objections, none high-severity, not truncated). Both medium objections
> were answered with measurements and the answers are in `bugs_open/259` — the short version is
> that `a.db` and `a.stateRepo` are assigned on adjacent lines inside one `if`, so agentbase's
> dedupe gate is provably **as strong as** "the processor has a live handle", and separately the
> deletion is a no-op regardless of that gate because site C never ran. **The only remaining
> owed item is (b): verify after the roll, then close the bug.**

**Cold-start for a new chat. Read this first; it is self-contained for the next task.**
Supersedes `HANDOFF_2026-08-13_continue_here.md`, and with it `-08-12b` and `-08-12`, all now
history. **`HANDOFF_2026-08-12` contains a claim that is false** (corrected in 12b §2) — do not
mine it for facts.

**The lane now has real standing docs** (created late, 2026-08-14 — the lane had run four days
on handoffs alone): `NOTES_dispatch_fail_closed.md` for the technical log and missteps,
`README_where_we_are.md` for the owner's plain-prose history. **Update those as you go**; do not
put the running record in the next handoff.

## 1. What is finished

**D2 / SYS-091 — pool instrumentation: DONE, LIVE, SCRAPED, PROVEN on v1.0.1295.** Nothing owed.
Full evidence in `HANDOFF_2026-08-13`, register SYS-091, and its cautionary tale (an instrument
built, approved and serving the right number on both pods while Prometheus scraped **neither** —
a PodMonitor's numeric `targetPort` keys on the port a pod DECLARES). The index row that still
read `built` was corrected 2026-08-14 in `82b159ee9`.

**`bugs_open/259` (slug `three_processor_paths…`) — candidate 1 APPLIED, 2026-08-14.**
`e37f79b65`, plus `f894b1a38` (gofmt). `p.sqlDB` no longer exists; `MessageProcessor` has one
handle. Six dependents went with it: the `DATABASE_URL` open block, the unread `stateRepo` field,
site A (inert), site B's two callerless functions (taking the only fleet-wide
`{"status": "completed"}` placeholder), site C (the redundant dedupe claim), and the callerless
`createSQLDB`. The three live fallback readers were **simplified to `p.db`, not deleted** —
including the one in `validation_drop.go` whose operands were the opposite way round.

⚠ **259 IS AN AMBIGUOUS NUMBER** — an unrelated GPU-provisioning `259` was filed the same day.
Resolve by slug; `git log` the file path, never the bare number.

## 2. THE TWO THINGS OWED — this is the next task

### (a) ~~READ THE COUNCIL VERDICT, and act on it~~ — **DONE, APPROVED. Kept for the queries.**

`Council-Submitted: 0ff072ef-ee02-465e-8a70-f5461c585ec9`. Submitted 2026-08-14, still running
at handoff time (last seen at `gate_tooling_provenance`). **The code is already on the shared
branch**, so a REVISE or REJECTED is work owed, not a formality.

```sql
-- the run
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '0ff072ef-ee02-465e-8a70-f5461c585ec9';
-- the verdict
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='0ff072ef-ee02-465e-8a70-f5461c585ec9' AND kind='council_report' ORDER BY created_at;
-- the human-readable note
SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
```

A missing orchestration row is almost always **latency, not a dropped dispatch** — do not
resubmit on that evidence. On APPROVED: nothing to do, `098` credits the `Council-Submitted:`
trailer automatically at report time, with no amend. On REVISE: resubmit with
`RESUBMIT_CORR=0ff072ef-…` so the trail accumulates. The submission JSON is worth reusing as a
base — its `risks` block already names the four things a reviewer should attack, in the order
they matter.

### (b) VERIFY AFTER THE ROLL, and only then close the bug

**`bugs_open/259` stays OPEN.** The bar is fixed **and live**, and this is a Go change — inert
until a fleet roll, so the defect is still reproducible on every chassis pod. **Do not move the
file to `bugs_closed/` on the strength of the commit.**

Prove the roll at the artefact, not at git:
```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor e37f79b65 <the stamp>   # "did my fix ship?" is a query
```
⚠ That line **scrolls within the hour** on the chassis, and a full-log grep for the phrase
**false-matches council-gate payloads that quote it** — treat any match inside a giant JSON log
line as a false positive. Read the stamp **per service**, not per fleet.

Then the honest check — the deletion should be a fleet-wide no-op, and these two could come out
either way:

1. **`recordDispatchFailureState` still writes its FAILED rows** (`bugs_closed/239`'s proof).
   Dispatch a **single-line** envelope naming a non-existent agent (⚠ `kcat -P` sends one
   message per LINE — a multi-line heredoc becomes N invalid fragments), then:
   `SELECT owner_agent_type, status, error FROM orchestration_states WHERE correlation_id = '<corr>';`
2. **`agent_error_log` still gains rows from `recordDroppedValidationError`** — the reader whose
   operands were reversed, so the one most worth watching. ⚠ domain column trap on that table:
   `COALESCE(domain,'') = ''`, never `domain IS NULL`.
3. **C's deletion changed nothing about dedupe** — `processed_messages` keeps its write rate:
   ```sql
   SELECT date_trunc('hour', processed_at) hr, status, count(*), count(DISTINCT processed_by)
   FROM processed_messages WHERE processed_at > now() - interval '4 hours' GROUP BY 1,2 ORDER BY 1 DESC;
   ```
   Baseline to beat: **449 rows / 82 writers** in the 13:00 hour, 2026-08-12. **Do not** use the
   `Duplicate message ignored` / `DEDUPE_CLAIM_LOST` log markers — they fire only when a
   duplicate actually occurs, so zero there means nothing.

A silent path in 1 or 2, or a collapse in 3, means a live path was misclassified as dead.

## 3. Also still open on the lane

- **Memory-index compaction.** The hook fires at 92% of the 25,000-byte cap. **Count is the
  binding axis, not bytes** — an arrival must displace one; closed-and-live bug entries →
  `MEMORY_closed.md` is the sanctioned exit. 239, 246, 247, 091/184, 108, 170 are all closed.
- **`podmonitor.yaml` is live but not in the kustomize build** (`base/kustomization.yaml` lists
  only `deployment.yaml`) — hand-applied, reconciled by nothing, drift silent both ways. Left to
  the `bugs_open/040` lane, which owns the file; wiring it in changes what a whole-fleet release
  applies. **Do not compete on 040.**
- **Not ours, but do not inherit the blame:** `internal/adapters/thunder/api/client_test.go:113`
  does not compile at HEAD (`unknown field Identifier in struct literal of type Instance`).
  Committed by another session; the package is clean in the tree and references nothing this
  lane touched. ⚠ `go build ./...` does **not** compile test files, so a build-only baseline
  cannot see it — take baselines with `go test`/`go vet` or you will be blamed for it.

## 4. Standing traps for this lane

- **`gofmt -l <files>` BEFORE you commit.** A comment inserted between struct fields splits
  gofmt's alignment group; the pre-commit check caught it only *after* `e37f79b65` had landed,
  and the build gate rejects un-gofmt'd code.
- **A diff-grep cannot see a changed markdown BULLET.** `git diff | grep '^[+-][^+-]'` returns 0
  on a `- ` bullet line, because the second character is the bullet's own dash. **Gate on
  `git diff --numstat` first** — it printed `1 1` where the grep printed `0`, on this lane, on
  2026-08-14, with the landmine already on screen from the SessionStart hook.
- **A redirected test that passes is not a test that still checks anything.** Mutate the
  production code to prove it. Copy the file to the scratchpad and restore from that copy — not
  `git stash`/`checkout`, which on this tree can take another session's work with it.
- **"No non-test callers" answers reachability, not "what will stop compiling."** Site B's
  wrapper had a test caller the fix list had not carried forward.
- **A sha is generated output, never retyped** (`git rev-parse`, or paste from `git log`).
- Deployment manifests and docs are **outside council scope** (`platform/`, `internal/`, `pkg/`).
- Peer lane on the "bugfix 238" socket is `bugfix_246_shared_pool_ownership/`, NOT the 238 lane.
  246 is CLOSED and handed its D5 to `bugs_open/259`; its newest doc is 2026-08-11.

## 5. Where everything is written down

- `bugs_open/259` (slug `three_processor_paths…`) — the three sites with their separate proofs,
  the 08-13 corrections, and the 08-14 status block recording what shipped and what stayed open.
- `NOTES_dispatch_fail_closed.md` · `README_where_we_are.md` — this lane's standing docs.
- `bugs_closed/239`, `bugs_closed/246` — the closed predecessors.
- Register **SYS-090** (dispatch seam; its open review question is now struck through and
  resolved), **SYS-091** (pool instrumentation, LIVE/SCRAPED/PROVEN).
- `LANDMINES.md` — the PodMonitor `targetPort` trap.
- `WRONG_CALLS.md` — 2026-08-12, "the reader was checked first".
- `architecture_review/RFC_023` — the scope ruling to cite at the gate.
