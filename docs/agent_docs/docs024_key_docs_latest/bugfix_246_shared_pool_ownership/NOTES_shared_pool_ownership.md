# NOTES — bugfix 246, the shared `*sql.DB` and who owns its sizing

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-11 — session start: picking the bug, and proving it is still real

### Ownership check (before any work)

`bugs_open/` has ~95 files. Rather than trust the numbering, I greped the **live
session transcripts** for bug numbers, because `scripts/who-owns.py` reads COMMITS
and is blind to a session mid-fix (memory: `who-owns-is-blind-to-uncommitted-sessions`):

```bash
cd ~/.claude/projects/-home-ant-projects-agentchassis/
for f in $(ls -t *.jsonl | head -32); do
  nums=$(tail -c 400000 "$f" | grep -oE 'bugs_open/[0-9]{3}' | sort -u | tr '\n' ' ')
  echo "$f :: $nums"
done
```

31 sessions were active within the last ~20 minutes. `246` appeared in **none** of
them. `who-owns.py 246` also returned "(none identified)" with no commit whose
subject is about it. Third, independent corroboration: the **filing lane's own
handoff** says so in writing —
`bugfix_239_dispatch_fail_closed/HANDOFF_2026-08-11_continue_here.md`:

> ## Related work filed by this lane, unowned, NOT started
> - `bugs_open/246` — `NewMessageProcessor` re-shrinks the SHARED `*sql.DB` …

Three independent checks agreeing is worth more than any one of them, and they are
not blind the same way (transcripts see uncommitted sessions; `who-owns` sees
commits; the handoff is the filer's own statement of intent).

### Is the bug still valid? — yes, and the bug file's own quote is wrong

The bug file quotes `agent.go` as:

```go
maxOpen := 12 // production value via CHASSIS_DB_MAX_OPEN_CONNS
```

**That is not what the code says.** `platform/agentbase/agent.go:277-289` reads:

```go
maxConns := 4
if v, err := strconv.Atoi(os.Getenv("CHASSIS_DB_MAX_OPEN_CONNS")); err == nil && v > 0 {
    maxConns = v
}
db.SetMaxOpenConns(maxConns)
db.SetMaxIdleConns(1)
db.SetConnMaxLifetime(time.Minute * 10)
```

The **default is 4**, and 12 only arrives from the environment. This matters more
than a typo: it is the whole blast-radius argument (below). The bug file's
paraphrase would have led a fixing session to believe the code hard-codes 12 and
that every agent is affected. It does not, and they are not. [VERIFIED — read the
file at HEAD 2026-08-11.]

`platform/messaging/processor.go:68-74` then does, unconditionally, on the handle
it was **passed**:

```go
db.SetMaxOpenConns(4)
db.SetMaxIdleConns(1)
db.SetConnMaxLifetime(time.Minute * 10)
```

### The mechanism, proven in isolation WITH A CONTROL

A `*sql.DB` is a pool object, not a connection, so the second call re-sizes the
first pool. I did not want to assert that from the documentation, so I proved it
(scratch program, no database needed — `sql.Open` does not connect):

```
after agentbase sizing: MaxOpenConnections = 12
after processor sizing:  MaxOpenConnections = 4
control (set to 9):      MaxOpenConnections = 9
nil *sql.DB SetMaxOpenConns panicked: runtime error: invalid memory address ...
```

The **control line is the point**: the probe reports 9 when the pool is set to 9,
so the reading of 4 could have come out otherwise. Without it, "it printed 4" is
compatible with a probe that always prints 4.

The fourth line is a **second finding the bug file does not mention**: `agentbase`
leaves `a.db` nil when `a.config.DatabaseURL == ""`, and `(*sql.DB)(nil).SetMaxOpenConns`
**panics**. So a DB-less agent panics inside this constructor today. Deleting the
three calls removes that crash as a side effect.

### The live half — the env var IS set, so the defect is not theoretical

```bash
kubectl -n ai-persona-system exec <chassis-pod> -- sh -c 'echo "$CHASSIS_DB_MAX_OPEN_CONNS $CHASSIS_INTAKE_MODE"'
# => 12 worker_pool_all      (both replicas)
```

Both live chassis replicas set `CHASSIS_DB_MAX_OPEN_CONNS=12` **and**
`CHASSIS_INTAKE_MODE=worker_pool_all`. That second one matters: the worker-pool
intake is precisely the workload the 12 was raised for, and the `agent.go` comment
predicts this exact failure —

> 4 workers + two consume loops + the response path + the retry driver against 4
> connections is a queue, and a freeze if anything holds a transaction while
> acquiring a second conn.

So the fleet is running the workload that motivated the raise, on the pool size the
raise was meant to replace.

### A finding neither the bug nor CS-2 records: the env keys are LIVE-ONLY

```bash
kubectl kustomize deployments/kustomize/services/agent-chassis/overlays/production/uk_001 | grep CHASSIS_
# => (nothing)
```

`CHASSIS_DB_MAX_OPEN_CONNS`, `CHASSIS_INTAKE_MODE` and `CHASSIS_RESPONSES_START_AT`
render **nowhere** in the production overlay. They exist only on the live
Deployment object.

> **CORRECTION, same session, before I wrote it down anywhere durable.** My first
> reaction was "then the next `apply -k` strips them" — the same shape as the
> known `kubectl scale` landmine. **That is wrong**, and I checked before asserting
> it. `kubectl apply` does a three-way merge: a field present in the live object but
> absent from BOTH `last-applied-configuration` and the incoming config is
> **preserved**, not pruned. The keys are absent from the annotation:
>
> ```bash
> kubectl -n ai-persona-system get deploy agent-chassis \
>   -o jsonpath='{.metadata.annotations.kubectl\.kubernetes\.io/last-applied-configuration}'
> # => CHASSIS keys in last-applied: NONE   (19 env keys total)
> ```
>
> So they survive an ordinary deploy. What is true is weaker and still worth
> saying: **the repo does not record how the fleet is actually configured**, so a
> reader of `deployments/` cannot learn that the chassis runs `worker_pool_all`
> with 12. That is a documentation defect, not a time bomb. Logging it as a near-miss
> rather than a finding, because the alarming version is the one I nearly wrote.

### The instrument, and why its zero proves nothing yet

The 239 lane left an instrument and told the next session to use it before touching
the pool: a transient agent-definition lookup fault now logs `DISPATCH_LOOKUP_RETRYABLE`.

```
agent-chassis-...-l2bwt: DISPATCH_LOOKUP_RETRYABLE=0  (1808 log lines available, pod up ~1h45m)
agent-chassis-...-twzdn: DISPATCH_LOOKUP_RETRYABLE=0  (129 log lines available)
```

**Zero — and I am not going to report that as "the pool is healthy".** A zero needs
a demand control (memory: `a-post-fix-zero-needs-a-demand-control`), so:

```sql
SELECT date_trunc('hour', received_at) AS hr, kind, status, count(*)
FROM chassis_intake_events
WHERE received_at > now() - interval '6 hours'
GROUP BY 1,2,3 ORDER BY 1 DESC;
```

~60–100 events/hour across both replicas, i.e. **1–2 messages per minute**. There
was real work in the window, so the instrument was not idle — but 1–2/min is
nowhere near saturating even a 4-connection pool. **The measurement cannot
discriminate** between "4 is fine at this load" and "4 would freeze under load".

That is the honest framing of the whole bug, and it agrees with the filer, who
wrote "Severity: unknown, and that is the point":

> **246 is a latent config-inertness defect, not an observed outage.** The case for
> fixing it is an ownership/correctness argument — a deliberately-set platform knob
> is inert and nothing can tell you — NOT a performance claim. Any submission that
> argues impact will be arguing something it cannot evidence.

`[UNMEASURED]` — whether a 4-connection pool actually throttles the worker-pool
intake at peak. Nothing surfaces `db.Stats().WaitCount` / `WaitDuration`, so there
is **no instrument that can see pool saturation today**. Settling it needs either
that instrumentation or a load test, and neither is required to justify the fix.

### pgbouncer — checking the bug's "does this just move the queue?" question

`deployments/kustomize/services/pgbouncer/pgbouncer-configmap.yaml`:
`pool_mode = transaction`, `max_client_conn = 200`, `default_pool_size = 15`,
`min_pool_size = 2`, `reserve_pool_size = 5`.

Transaction pooling is exactly the mechanism that absorbs more idle client
connections, and 2 replicas × 12 = 24 client connections against a 200 ceiling is
comfortable. So raising the effective client cap does **not** simply move the queue.

Two things to say out loud rather than bury:
- `default_pool_size = 15` is the **server-side** ceiling per user/database pair and
  is shared by every client of `clients_user`/`clients_db`, not just the chassis.
  That ceiling exists today and this change does not create it — but it does mean
  the chassis could now ask for a larger share of it.
- The configmap's own comment reasons from **"3 chassis replicas × 4 conns"**. That
  rationale is stale in two ways at once (it is 2 replicas, and the intent is 12).
  This is a consumer that must be *told*, per the owner ruling of 2026-07-29 §3.

### Where the pool-sizing calls live, fleet-wide

```bash
grep -rn "SetMaxOpenConns" --include=*.go . | grep -v vendor/
```
`cmd/scheduler/main.go:76` (3), `internal/adapters/thunder/adapter.go:136` (10),
`platform/agentbase/agent.go:283` (env), `platform/database/mysql.go:38` (10),
`platform/database/postgres.go:161` (10), `platform/messaging/processor.go:68` (4),
plus tests. Only **one** of these sizes a handle it did not open: `processor.go`.

`NewMessageProcessor` has exactly one non-test caller: `platform/agentbase/agent.go:341`.
</content>

---

## 2026-08-11, later — the plan, the loop that could not answer, and the fix

### The 090 run came back with NO VERDICT, and that is the interesting part

Filed as intake `99c300b9`, run correlation `105970e4-dd02-4654-9536-84a2dd6a3da2`.
It ran to `COMPLETED` through **5 iterations**, wrote **5 `bundle` artifacts, zero
`council_report`, and no `doc_notes` row.** A completed run with no verdict reads
exactly like a run that found nothing worth saying.

Reading the final bundle's `data_requests` section shows what actually happened,
and it is worth writing down:

> **The model asked exactly the right question.** Its own words: *"finding either
> key in any `agent_definitions.env_vars` row would show a case where the override
> to 4 is actually consequential rather than a no-op."*
>
> It got `(0 rows)`. It re-ran the identical query in a later iteration under the
> F0.5 persistence rule and got `(0 rows)` again — **two checks blind in the same
> way, agreeing with each other.** Then it exhausted its iterations.

The loop's evidence surface is **the repo plus `clients_db`**. `CHASSIS_DB_MAX_OPEN_CONNS`
is set on the **Kubernetes Deployment object**. The loop has no `kubectl` and
structurally cannot see it.

**Why this is a landmine rather than a shrug:** the evidence it *could* reach pointed
the wrong way. `(0 rows)` reads as "nothing sets this key, so the override is a
harmless no-op" — and the truth is that 2 pods of 95 set it, *both of them the pods
the symptom was about*. Had it emitted a verdict on what it had, the natural verdict
was a confident REFUTED, and a filing session would have closed a live bug on it.
Appended to `LANDMINES.md` and synced.

**Was the 090 wasted?** No — but not for the reason I expected. It did not confirm
the mechanism (I had already done that in isolation, with a control). What it
produced was a *fact about the diagnosis loop*, which is now recorded for everyone.

### MISSTEP: fable was unavailable and I did not have a fallback

I dispatched the implementation plan to a `fable`-model Plan agent. It died at ~6
minutes: `You've hit your session limit · resets 6:50pm`. **No plan was produced.**
I wrote the plan myself from the analysis already done — the evidence gathering was
the expensive part and it was already finished, so the loss was small. Recording it
because the failure mode is invisible in the output: a plan doc exists either way,
and nothing in it says which mind wrote it. `PLAN_2026-08-11_shared_pool_ownership.md`
is **mine**, not fable's.

### The design decision I want on record, because I rejected the "best" answer

The theoretically strongest fix is to narrow the `db` parameter to a query-only
interface, so the processor **structurally cannot** resize a pool — "make the bad
state unrepresentable", which is this estate's stated ordering rule for fix
candidates. **I rejected it, deliberately.**

The reason is a measurement, not a preference: `p.db` is passed onward to
`orchestration.NewStateRepository(db *sql.DB)` (`state.go:174`) and
`NewSagaCoordinator(db *sql.DB)` (`coordinator.go:92`) — **both public constructors
used across the platform.** Narrowing the type ripples into shared signatures far
beyond this bug, to fix a three-line defect that has **exactly one caller**. The
behavioural test gets most of the protection (a re-introduction fails loudly) at a
fraction of the risk. If a second caller ever appears, revisit.

I note this is the one place I am knowingly not taking the most robust option, so a
reviewer can disagree with it on the record rather than discover it.

### The test seam — and why the existing suite could never have caught this

Every existing test in `platform/messaging` builds `&MessageProcessor{logger: ...}`
as a **struct literal** (`error_step_plan_test.go:21`, `processor_response_status_test.go:61`,
`processor_dispatch_resolution_test.go:46`). That is convenient and it means **no
existing test calls `NewMessageProcessor` at all** — so an entire class of defect,
"the constructor does something wrong", was invisible to the suite by construction.
The new test calls the real constructor. Verified safe to pass nils: the constructor
only stores producer/orchestrator/validator/initializer, and `NewStateRepository`
stores its arguments without dereferencing (`state.go:174-176`).

### MUTATION PROOF — the test earns its place

A passing test proves nothing on its own. Re-introducing the three deleted lines:

```
--- FAIL: .../operator_configured_12,_as_the_live_chassis_does
        NewMessageProcessor re-sized the pool it was handed: got 4, want 12.
--- FAIL: .../control:_a_different_value_must_also_survive
        NewMessageProcessor re-sized the pool it was handed: got 4, want 9.
--- FAIL: TestConstructorSizesThePoolItOpensItself
        the caller's pool was re-sized while opening our own: got 4, want 12
```

The **9** control failing is the load-bearing line: it proves the assertion is not a
hardcoded comparison against 12 that would pass for the wrong reason.

`go build ./...` clean · `gofmt -l` clean · `go vet` clean · package tests pass.

### MISSTEP: I validated the council JSON against the RUNBOOK's TRAP LIST instead of the schema

The submission bounced client-side: `ERROR: .plan.summary is empty`. I had put
`grounded_in` and `risks` at the top level and omitted `plan.summary`. The real
schema is in the **097 script header, lines 24-36** — `summary`, `edits`,
`grounded_in`, `risks` all live **inside `plan`**.

What makes this worth recording is *how* I got it wrong: I read the RUNBOOK's traps
section carefully and validated against **those** (`risks` must be a string not an
array; `operation: create` is refused; no-op phrase blocklist) — and arrived
confident, having checked the hard parts, with the trivial part wrong. **A trap list
is not a schema.** Checking documented gotchas can substitute for reading the
definition and feels more rigorous than reading it. Cost: one invocation, no credits,
because 097 validates client-side. Logged to `WRONG_CALLS.md`.

### Committed

`039fcce84` — `platform/messaging/processor.go` + the new test, pathspec commit, 2
files, both mine (the commit-scope block confirmed it). Trailer
`Council-Submitted: c94d73ac-2a15-40cb-98a9-1185a2b7435a` — **not** `Council-Reviewed:`,
because no verdict had been read at commit time.
