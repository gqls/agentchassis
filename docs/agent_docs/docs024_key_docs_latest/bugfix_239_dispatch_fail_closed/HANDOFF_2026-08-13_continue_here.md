# dispatch/pool lane — HANDOFF 2026-08-13: D2 is CLOSED and proven; 259 is fully diagnosed and ready to fix

**Cold-start for a new chat. Read this first; it is self-contained for the next task.**
Supersedes `HANDOFF_2026-08-12b_continue_here.md` and `HANDOFF_2026-08-12_continue_here.md`,
both now history. **12a contains a claim that is false** (corrected in 12b §2 and again
below) — do not mine it for facts.

## 1. What is finished, with its evidence

**D2 / SYS-091 — pool instrumentation: DONE, LIVE, SCRAPED, PROVEN.** Nothing owed.

- Council verdict **APPROVED** (`e3aa14c5-adcd-4472-b0ee-213ae043e378`, 2026-08-12 17:44:34Z).
  `098` credits the `Council-Submitted:` trailer on `11abe7a41` automatically.
- `[MEASURED 2026-08-13 14:07Z, v1.0.1295]` both chassis pods declare `8080,9090`, both are
  active Prometheus targets (**2 of 35**, `health:"up"`, `:9090/metrics`), and
  `go_sql_max_open_connections{pod=~"agent-chassis-.*"}` returns **two series reading 12** —
  matching `CHASSIS_DB_MAX_OPEN_CONNS=12`. First chassis contention baseline:
  `wait_count_total 0`, `wait_duration_seconds_total 0`, `open 2`, `in_use 1`.
- **⚠ Those two zeroes are a BASELINE, not a result.** Cumulative since process start, pods
  were 14 minutes old. And the target denominator moves with fleet load (141 → 35 in a day
  as ephemeral agent Jobs come and go), so never state a chassis finding as "N of M targets"
  without its date.

**The defect found on the way, because it is the lane's cautionary tale.** D2 was built,
council-approved, mutation-tested and serving `12` on both pods — and Prometheus was scraping
**neither**. A PodMonitor's numeric `targetPort: 9090` compiles to
`keep __meta_kubernetes_pod_container_port_number == "9090"`, which matches the port a pod
**DECLARES**, not the port it serves; the chassis Deployment declared only 8080. The spawned
agent pods declare 9090, so the metric returned **108 healthy series with zero rows from the
two pods it existed for**, and because the `job` label is the *PodMonitor's name* they all
read as `job="ai-persona-system/agent-chassis"`. Fixed in `889a7c055`
(`base/deployment.yaml` now declares `9090 name: metrics`). Corrections landed in SYS-091,
`podmonitor.yaml`, `LANDMINES.md`, `WRONG_CALLS.md`, and a contribution into `bugs_open/040`
(that lane owns the file — **do not compete on it**).

**The lesson worth carrying:** an aggregate over a superset is not evidence about a member.
Filter by the thing you mean (`{pod=~"agent-chassis-.*"}`) and ask the **target list for the
pod name**, never the metric for a health status.

## 2. `bugs_open/259` — fully diagnosed 2026-08-13, ready to fix

⚠ **AMBIGUOUS NUMBER** — resolve by slug `three_processor_paths_guard_on_a_handle…`; an
unrelated GPU-provisioning `259` was filed the same day. `git log` the file path.

**All three sites are now assessed and all three are DELETIONS**, each on its own separate
proof — the conclusions converged, the justifications did not:

| site | where | verdict | proof |
|---|---|---|---|
| **A** | `processor.go:350-360`, child-workflow completion | **INERT** | both the guarded branch and the fall-through `return nil`; the whole skipped region is one log line; `process()` (`:141-:364`) has an **unnamed** `error` return and **no `defer`**, so nothing can distinguish the two returns |
| **B** | `sendWorkflowSuccessResponse` `:567-596` (+ inner `sendWorkflowResponse` `:789-803`) | **UNREACHABLE** | zero callers repo-wide, proven by grep with a live-sibling control; the live path reaches `sendWorkflowResponseWithStatus` (`:804`) directly |
| **C** | `processor.go:1486…`, two-phase dedup claim | **REDUNDANT** | `agentbase` runs the same claim on a live handle (`agent.go:1149-1173`), evidenced 449 rows / 82 writers in one hour |

**A was the last open question** and the filing expected it to need a live behavioural
measurement ("duplicate parent response, or belt-and-braces?"). The answer is **neither** —
the premise was false at source: that `return nil` suppresses nothing, because there is no
response-sending code for it to skip. Full quoted proof is in the bug file. `[MEASURED
2026-08-13, v1.0.1295]` `DATABASE_URL` is empty on both chassis pods while
`CLIENTS_DATABASE_URL` is set, so `p.sqlDB` is nil in production exactly as filed.

### THE NEXT TASK — candidate 1: delete `p.sqlDB` entirely

Recommended in the bug file and now **unblocked** (A's inertness dissolved the ordering
constraint that said A had to be treated first). It makes the bad state unrepresentable:
there is no second handle left to guard on. **Re-locate every line number before editing** —
this file is edited by several lanes.

**Delete outright:**
- `processor.go:34` — the `sqlDB *sql.DB` field.
- `processor.go:~85-110` — the `DATABASE_URL` open block in `NewMessageProcessor` (read its
  comment first: it documents a real second-pool sizing hazard, which disappears with it).
- **A** `:350-360` — delete the whole `if msgCtx.IsChildOrchestration()` block at `:351`.
  Provable no-op. ⚠ **Not** the one at `:294`, which is live and stores parent context.
- **B** `:567-596` and `:789-803` — both dead functions. This also removes the only
  fleet-wide `{"status":"completed"}` placeholder literal.
- **C** `:1486…` to its closing brace, plus the `zap.Bool("is p.sqlDB exists…)` field in the
  log call just above it at `:1483`.

**Simplify, do NOT delete — these are the LIVE paths.** Three fallback readers, one of them
outside `processor.go` and **not listed in the bug file**:
- `processor.go:~649-658` and `:~1192` — `db := p.db; if db == nil { db = p.sqlDB }`
  (left by `bugs_open/239`) → collapse to `p.db`.
- `platform/messaging/validation_drop.go:79-82` — `db := p.sqlDB; if db == nil { db = p.db }`
  — **note the operands are in the OPPOSITE order** to the other two, so read it before
  editing; it prefers the nil handle and falls through. → collapse to `p.db`.

**Two tests reference the field and will not compile after it goes:**
- `processor_pool_ownership_test.go:107-110` — asserts `sqlDB` **is** opened when
  `DATABASE_URL` is set, and sizes it. This test's whole subject is being deleted; remove it,
  and say so in the commit rather than letting the coverage quietly drop.
- `processor_dispatch_resolution_test.go:433-454` — asserts the fixture has `sqlDB` **nil**
  ("the production shape"). Its guard becomes vacuous once there is one handle; keep the test
  (it covers 239's fix) and drop only the `sqlDB` assertions.

**Duties owed on it:** platform code → **council gate** submission, one round, cite RFC_023's
behaviour-not-package-count ruling if scope is questioned. **Register touch in the same
commit.** The 246 lane's D5 was the original "collapse `p.sqlDB` into `p.db`" item, so check
`bugfix_246_shared_pool_ownership/` before starting in case they have moved on it —
`scripts/who-owns.py 259` reads commits only and is blind to a session mid-fix, so check the
tree too.

**Verification that could come out either way:** the deletion should be a no-op fleet-wide.
The honest check is not "does it still build" but the two live paths — after the roll, confirm
`recordDispatchFailureState` still writes its FAILED rows (239's proof, corr trace) and that
`agent_error_log` still gains rows from `recordDroppedValidationError`. Remember the domain
column trap on that table (`COALESCE(domain,'') = ''`, never `domain IS NULL`).

## 3. Also still open on the lane

- **Memory-index compaction.** The hook fires at 92% of the 25,000-byte cap. Sanctioned exit:
  move closed-and-live bug entries to `MEMORY_closed.md`. **Count is the binding axis, not
  bytes** — an arrival must displace one. 239, 246, 247, 091/184, 108, 170 are all closed now.
  I deliberately did **not** add a memory entry for the PodMonitor trap: it is in
  `LANDMINES.md` (the system of record, synced to `doc_notes`), and adding an index line while
  the index is over budget would worsen the documented problem.
- **`podmonitor.yaml` is live but not in the kustomize build** (`base/kustomization.yaml`
  lists only `deployment.yaml`) — hand-applied, reconciled by nothing, drift silent both
  ways. Recorded in the 040 contribution and left to that lane: wiring it in changes what a
  whole-fleet release applies.

## 4. Standing traps for this lane

- **Prove a deploy at the artefact.** The `build provenance` log line **scrolls within the
  hour** on the chassis, and a full-log grep for it **false-matches council-gate payloads that
  quote the phrase** — hit again on 2026-08-13, so treat any match inside a giant JSON log
  line as a false positive. Better: ask the pod what it declares/serves. A `/proc/1/exe` sha
  probe needs a **narrow** candidate list — the fleet commits fast enough that the 40 most
  recent commits spanned ~28 minutes on 2026-08-12, and a 400-pattern `grep -f` against the
  binary times out.
- **`kcat -P` sends one message per LINE** — single-line `<<<` envelopes only.
- **A sha is generated output, never retyped** (`git rev-parse`, or paste from `git log`).
- Deployment manifests and docs are **outside council scope** (it takes `platform/`,
  `internal/`, `pkg/`) — which is why `889a7c055` carries no trailer, correctly.
- Peer lane on the "bugfix 238" socket is the shared-pool-ownership lane
  (`bugfix_246_shared_pool_ownership/`), NOT the 238 lane.

## 5. Where everything is written down

- `bugs_open/259` (slug `three_processor_paths…`) — the three sites, each with its proof, and
  the fix candidates with the 2026-08-13 corrections.
- `bugs_closed/239`, `bugs_closed/246` — the closed predecessors.
- Register **SYS-090** (dispatch seam), **SYS-091** (pool instrumentation, status now
  LIVE/SCRAPED/PROVEN with the wrong reader-check struck through).
- `LANDMINES.md` — "a PodMonitor's numeric `targetPort` keys on the port a pod DECLARES".
- `WRONG_CALLS.md` — 2026-08-12, "the reader was checked first".
- `bugs_open/040` — the metrics plumbing, with this lane's contribution at the foot.
- `architecture_review/RFC_023` — the scope ruling to cite at the gate.
