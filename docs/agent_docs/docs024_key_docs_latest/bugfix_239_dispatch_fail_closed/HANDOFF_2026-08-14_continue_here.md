# dispatch/pool lane — HANDOFF 2026-08-14: **259 is CLOSED, LIVE and VERIFIED. The lane's bug work is DONE.**

> **UPDATED at the end of 2026-08-14 — BOTH owed items are now discharged.** The fresh chassis
> build (`v1.0.1298`, pods up 08:58Z) carries the fix; it is verified at the artefact and
> behaviourally; `bugs_open/259` has moved to **`bugs_closed/259`** (`c5072e142`). Register SYS-090
> updated. **Nothing on 259 remains** except one optional observation noted in §2(b).
> The lane's remaining work is §3 — memory-index compaction, and the 040 lane's podmonitor item.

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

## 2. ~~THE TWO THINGS OWED~~ — **BOTH DISCHARGED. Kept for the queries and the technique.**

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

### (b) ~~VERIFY AFTER THE ROLL, and only then close the bug~~ — **DONE on v1.0.1298. Bug CLOSED (`c5072e142`).**

Full evidence lives in **`bugs_closed/259`**; the technique is now a 016b §9 entry. Summary:

- **Deploy proven at the artefact by ABSENCE with a PRESENT-control**, because a deletion has no
  new literal to grep. ⚠ **Two things that do NOT work here, both learned the hard way:** the
  `build provenance` line had already **scrolled past `--tail=6000`** on both pods (empty there
  means *not in range*, never *unstamped*); and a **sha probe is the wrong tool for a deletion** —
  the release builds from HEAD, so the stamp is a **later** commit than the fix and grepping for
  `e37f79b65` fails on a *correct* deploy. What worked: two deleted literals absent on both pods
  plus one **kept** literal present, so a broken probe cannot pass as a good deploy. Establish
  attributability first — the deleted literal must be >0 at `<fix>^` and 0 at HEAD.
- **CHECK 1 `recordDispatchFailureState` — PASS, measured.** Corr `8234e56e`, 13:55:58Z:
  `no-such-agent-259-postroll | FAILED | DISPATCH_UNRESOLVABLE: DISPATCH_FAIL_CLOSED …`. The
  `owner_agent_type` is the **requested** type, not `generic`. This is the load-bearing check
  because a nil `p.db` yields **no row at all** — the defect's own signature.
- **CHECK 3 `processed_messages` — PASS, measured.** Continuous across the roll, no loss of
  writers. ⚠ **Compare ACROSS the roll, not against the filing's `449 rows/82 writers`** — that
  was a busier day; today runs 71–82/hr and the difference is fleet load, visible in the writer
  count. The `08:20 260/48` bucket is release churn, not a signal.
- **CHECK 2 `recordDroppedValidationError` — INCONCLUSIVE by DEMAND, and that is the honest
  answer, not a pass.** ⚠ **`severity='warning'` does NOT identify this writer** — four writers
  share it. The discriminator is **`context ? 'matched_needle'`**. That writer fires **~2/day**
  and its last row (`02:37:58Z`) **predates the roll by 6¼ hours**, so ~0.4 rows were expected in
  the window and zero could not have come out otherwise. Closure does not rest on it: the edit
  was `db := p.sqlDB; if db == nil { db = p.db }` → `db := p.db`, and since `p.sqlDB` was *always*
  nil in production the old code *always* fell through to the exact handle the new code reads —
  identical by construction — while its only premise (`p.db` works on the live pod) is what
  CHECK 1 measured.

**The one optional loose end**, if you want the direct observation for completeness:
```sql
SELECT occurred_at, agent_type, action, context->>'matched_needle'
FROM agent_error_log WHERE context ? 'matched_needle' AND occurred_at > '2026-08-14 08:58:00+00'
ORDER BY occurred_at DESC;
```
⚠ domain-column trap on that table: `COALESCE(domain,'') = ''`, never `domain IS NULL`.

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
