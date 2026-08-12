# dispatch/pool lane — HANDOFF 2026-08-12: 239 is CLOSED; the lane continues with D2 (pool instrumentation) and 259 (dead sqlDB guards)

**Cold-start for a new chat. Read this, then `bugs_open/259_HANDOFF_2026-08-12_three_processor_paths…`
and `bugfix_246_shared_pool_ownership/HANDOFF_2026-08-11_continue_here.md` §D2.**
The previous handoff in this directory (2026-08-11) is history — its task completed and
everything below supersedes it.

## What just finished (all committed, nothing pending)

- **`bugs_closed/239`** — the fail-closed dispatch fix. Fully proven live (refusal on
  v1.0.1284; the FAILED-row trace on v1.0.1286, corr `cc7bd91a`, baseline zero
  all-history). **Moved to `bugs_closed/` at owner direction 2026-08-12** (`2aa3014a3`) —
  the owner restored CLAUDE.md's fixed-AND-live bar, superseding the 08-06 stay-direction.
- **RFC_023 is RULED** (owner, 2026-08-12, recorded at the foot of the RFC): the
  architecture-scope trigger is **BEHAVIOUR** (does a consumer's success path change),
  not the diff's package count — a vertical slice is not "many packages". **Rider:** code
  bloat/streamlining is a strong standing concern; the owner is open to adding
  agents/seats for it if the existing council seats do not hold that line. Next
  silent-success fix: ordinary gate, cite the ruling, no RFC.
- **Nested-envelope branch (`DISPATCH_OWN_DEFAULT`)**: owner 2026-08-12 — leave it to
  accumulate real fleet data; revisit with the measured population (was 7 msgs/8 days).
  Recorded in SYS-090.
- **v1.0.1288 is live** (stamp `bb5348642`): carries 239's fix (unregressed), 246's fix
  (`039cfce84` + follow-ups `6ba3fca28`), and 247's deletions (`8cb8938bb`).

## TASK 1 — D2: surface `db.Stats()` so pool saturation is measurable at all

**Ownership settled 2026-08-12:** this lane builds it; the shared-pool-ownership lane
(messages from the "bugfix 238" socket) stood down on it at our confirmation and takes
D4 (overlay honesty) + the pgbouncer `SHOW POOLS` credential (terraform, 047-base-configs).
Their constraints, all verified or credible, carried here so you need not re-derive:

- **Instrument the pool `agentbase` OPENS** (`agent.go:277-289`, sized from
  `CHASSIS_DB_MAX_OPEN_CONNS`, live value 12 since v1.0.1288). `p.sqlDB` is a different,
  nil-in-production handle — instrumenting it measures nothing (that mistake is exactly
  bug 259's shape).
- **The counters are CUMULATIVE since process start.** A raw gauge misleads across a
  roll. Export as Prometheus counters via the existing `observability` package (cumulative
  is idiomatic there; `rate()` belongs to the query side) — NOT a hand-rolled snapshot log.
- 2 pods of 95 carry the env var, one binary imports `agentbase` — chassis-only in
  practice; instrument unconditionally anyway, it is cheap and universal.
- **Baselines are not like-for-like across the 08-11 roll:** pre-roll readings
  (`DISPATCH_LOOKUP_RETRYABLE`=0 at 60–100 intake events/hr) were against the OLD
  4-connection pool; post-roll load is 182–228/hr against 12.
- This does **NOT** replace pgbouncer's `cl_waiting`/`maxwait` (`SHOW POOLS`) — Go side
  vs server side; the other lane is unblocking that credential. Both instruments wanted.
- **Register entry owed in the same commit** (the 246 ratchet names D2 as REGISTRABLE,
  not ratchetable). Platform code → council gate submission, one round, cite RFC_023's
  ruling if scope is questioned.

## TASK 2 — `bugs_open/259` (slug `three_processor_paths…`): three more dead `p.sqlDB` guards

⚠ **259 is an AMBIGUOUS NUMBER** — an unrelated `259_…one_provision_request…gpus` was
filed the same day. Resolve by slug, `git log` the file path.

Filed by the shared-pool lane, offered to and accepted by this lane as 239's natural
continuation (instance #4 of the same guard was `recordDispatchFailureState`, fixed at
`209917d15`). **This session re-verified all three at source** (`processor.go`, v. HEAD):

- **A (~:351)** child-workflow-completed early-return — dead; UNASSESSED whether that
  causes duplicate parent responses or is belt-and-braces.
- **B (~:582)** final-result read — dead, so **every response from this path carries the
  literal `{"status":"completed"}` instead of `CollectedData`.** Control-flow certain;
  downstream effect UNMEASURED. The live-behaviour question of the three.
- **C (~:1486)** the two-phase dedup claim — dead AND REDUNDANT: `agentbase` runs the
  same claim on a live handle (`agent.go:1149-1173`), proven working (449 rows/82 writers
  in one hour, lifecycle visible in `processed_messages`).

**The trap, verbatim intent from the filer and endorsed here: do NOT mechanically apply
239's `db := p.db; if db == nil { db = p.sqlDB }` fallback to these.** On C it would
switch on a second dedup layer that has NEVER run fleet-wide — deleting C is strictly
safer than fixing it. On A and B the fallback CHANGES LIVE BEHAVIOUR (B stops sending the
placeholder and starts sending real `CollectedData` to whatever parses responses today).
246's "no-op everywhere" symmetry argument does NOT transfer. Measure B's consumers
before choosing between fix-and-enable and delete. A diagnosis (`090`) run is warranted
for B if its consumer set is not quickly enumerable — but note the landmine the other
lane filed: **the diagnosis loop cannot see Deployment env vars** (no kubectl), so
env-dependent questions must be answered by hand first.

## TASK 3 (small) — memory-index compaction is due

The index hook fires at 92% of the 25,000B cap. The owner's standing ruling was "revisit
when bugs actually close" — they now are (239, 247, 091/184, 108, 170…). Sanctioned exit:
closed-and-live bug entries move to `MEMORY_closed.md`; count is the binding axis.

## Standing facts and traps for this lane

- **Prove a deploy**: probe `/proc/1/exe` for a KNOWN full sha with an absent-control;
  single-pass method for finding a stamp: pipe all candidate shas into ONE
  `kubectl exec -i <pod> -- grep -aoFf - /proc/1/exe`. The chassis's `build provenance`
  log line scrolls within the hour, and a full-log grep can false-match a council-gate
  payload QUOTING the phrase. Then `git merge-base --is-ancestor <fix> <stamp>`.
- **kcat -P sends one message per LINE** — single-line `<<<` envelopes only.
- **A sha is generated output, never retyped** (`git rev-parse`, or pasted from `git log`)
  — a c/f transposition cost a round of confusion on 246.
- Peer coordination: the lane on the "bugfix 238" socket is the shared-pool-ownership
  lane (`bugfix_246_shared_pool_ownership/`), NOT the 238 lane — commit `3b225ca84`'s
  body has that wrong; `d56a85346` is the correction.
- Council: one run per coherent task; `Council-Submitted:` when committing before the
  verdict; a scope veto is answered by recording, not resubmitting — and RFC_023's ruling
  now defines the scope line.

## Where everything is written down

- `bugs_closed/239_…` — the whole 239 story, closing verification at the foot.
- `architecture_review/RFC_023_…` — the ruling (foot of file).
- Register `SYS-090` (system-architecture.md) — the seam, its consumers, the open
  own-default question with the owner's accumulate decision.
- `bugfix_246_shared_pool_ownership/` — the other lane's standing five (D1–D5, traps).
- `bugs_open/246` (still open pending their post-roll verification) · `bugs_open/259`
  (three_processor_paths slug).
