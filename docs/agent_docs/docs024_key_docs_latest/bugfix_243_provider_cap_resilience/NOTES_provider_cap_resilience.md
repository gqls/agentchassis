# NOTES — 243-anthropic-cap: the platform's resilience to a provider refusal

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep.

Lane opened 2026-08-24. Subject: `bugs_open/243_HANDOFF_2026-08-10_anthropic_account_usage_limit_reached_every_llm_step_fleetwide_fails_until_september.md`
(**243-anthropic-cap** — the number is shared with an unrelated tool-acceptance bug; resolve by slug).

## 2026-08-24 — session 1: why this lane exists, and the disambiguation

`bugs_open/243` names TWO unrelated cases (CLAUDE.md's documented trap; `who-owns.py 243`
prints the ambiguity warning rather than an owner). I took the anthropic-cap one on the
owner's rule "if it already has an active thread, leave it":

- **243-tool-acceptance** (`…_tool_acceptance_look_step_has_no_storage_client_…`) is OWNED by
  `staged_component_build`, which committed to its own lane dir today (`926c8ffcd`,
  `c97d50b3c`, `53dedc4d2`, 2026-08-24). All three of its candidates are answered. **Left alone.**
- **243-anthropic-cap** states "OPEN, unowned" and has **no lane directory** — I checked
  `docs024_key_docs_latest/` for any anthropic/cap/provider/resilience dir: none. Six different
  lanes have *contributed* to the file (`bugfix_153_build_provenance` filed it;
  `bugfix_236_site_availability`, `bugfix_209/231`, `webdesign_uk_build_service`,
  `bugfix_281`/loop-engine, `bugfix_302_design_repair_verification`, `bugfix_305_negation_gate`
  each appended an occurrence) and **none owns it**. That is the gap this lane fills.

### The bug is still valid, and the file understates it `[MEASURED 2026-08-24 13:1x Z]`

Cap failures per day, `llm_call_log`, last 16 days:

```sql
SELECT date_trunc('day', created_at) AS d,
       count(*) FILTER (WHERE NOT success AND error_message ILIKE '%usage limit%') AS cap_failures,
       count(*) FILTER (WHERE success) AS ok,
       count(*) FILTER (WHERE NOT success) AS all_failed
FROM llm_call_log WHERE created_at > now() - interval '16 days' GROUP BY 1 ORDER BY 1;
```

| day | cap failures | ok | all failures |
|---|---|---|---|
| 08-10 | 7 | 457 | 7 |
| 08-14 | 28 | 751 | 38 |
| 08-17 | 4 | 867 | 6 |
| 08-19 | 5 | 914 | 13 |
| 08-21 | 3 | 1223 | 6 |
| 08-22 | **113** | 1063 | 116 |
| 08-23 | 32 | 1109 | 36 |
| 08-24 | 0 | 750 | 2 |

**Seven of the last fifteen days carried cap failures**, and on 08-22 the cap was essentially
the *only* failure mode (113 of 116). The bug file's framing — "third occurrence in 22 days" —
counts *narrated incidents*, not days with failures, and understates the rate by ~2×. The
file is not wrong, it is a chronicle of the events someone happened to be present for.

**State right now:** no cap active. `llm_call_log` shows 0 cap failures today across 750
successful calls; `ai_endpoint_health` claude row `healthy=t`, last_checked 12:42:10Z. So this
lane is working on a **quiescent** system, which is the right time to do it and also means every
fix needs a demand control — a green reading today proves nothing about the capped case.

### The monthly-limit shape, which the file does not state

The refusal always quotes a reset on the **1st** ("regain access 2026-09-01"), i.e. it is a
**monthly** spend limit. The events cluster toward month-end (08-19, 21, 22, 23) as cumulative
spend approaches the cap. `[INFERRED]` from the reset-date shape plus the clustering, not from
any billing API — but it makes a falsifiable prediction: **expect recurrence between now and
08-31**, and expect the first days of September to be quiet regardless of what we fix.

### 244's caching fix landed and did NOT prevent this — 243's standing conclusion is overtaken

The bug file says in four places that `bugs_open/244` (council prompt caching) is "the cheaper
prevention" and "fix that and this bug's trigger largely goes away". **244 is FIXED AND LIVE**
(its own banner; commits `3d6851d9b`, `071adc44c`, migration 376). It worked:

| day | full-price input tokens | cache read | cache write |
|---|---|---|---|
| 08-09 (pre-cache) | 42,716,865 | 0 | 0 |
| 08-21 | 5,708,697 | 121,127,386 | 12,338,987 |
| 08-22 | 6,664,088 | 102,528,376 | 9,972,791 |

Full-price input fell ~7×. But **total prompt volume grew several-fold**, so on the standard
weighting (reads 0.1×, writes 1.25×) effective input cost fell only ~30% (42.7M → ~29-33M
equivalent) — and the two worst cap days in the whole record are *after* the fix. **So the
prevention landed and the bug got worse.** Recorded here because the bug file still tells the
next reader to wait for 244.

### The mechanism, read first-hand — two seams, both unfixed today

**Seam A — the health gate is asymmetric: live traffic can only mark the endpoint DOWN.**

- `platform/orchestration/actions/ai_actions.go:617-645` — on any `isAIUnavailable(err)`
  (which matches `"usage limit"`, `ai_errors.go:96`), `execute_llm_prompt` fires
  `SELECT update_endpoint_health($1,false,$2)` in a goroutine and returns `*AIUnavailableError`.
- `update_endpoint_health` has **exactly one Go caller** — that one — and it **only ever passes
  `false`**. Verified: `grep -rn "update_endpoint_health" --include=*.go platform/ internal/ pkg/`
  → one hit. No trigger on the table (`pg_trigger` → 0 rows), and it is the only pg_proc
  mentioning `ai_endpoint_health`.
- The **only** writer of `healthy=true` is the probe loop,
  `check_endpoint_health_action.go:138`, which runs per `check_interval_seconds`. Live row:
  `claude / active / **3600** / claude_ping` (the two ollama rows are 30s and 60s).
- `claim_work_item_action.go:226-263` gates **every work-item claim fleet-wide** on that one
  boolean: `healthy=false` → claim released, item back to `triaged`,
  `{"claimed":false,"reason":"ai_endpoint_unavailable"}`.
- `find_dispatchable_site` is `ORDER BY created_at ASC … LIMIT 1` across all sites, so the whole
  fleet queues behind the single oldest item, which can never clear.

**So one 400 anywhere costs up to an hour of total fleet dispatch, and the recovery path is not
the one that failed.** The 08-17 addendum measured exactly this: `healthy=false` 11:09:53Z →
12:10:18Z = **60m25s** of fleet-wide dispatch stop, while 93 of 99 live Anthropic calls in the
same window SUCCEEDED. Nothing alerted; no item was marked failed; the only durable trace is
`claim_result.reason` inside orchestration `collected_data`.

The asymmetry is the part that was not previously written down. The 08-17 addendum blamed the
**probe's** lost coin-flip. That is one trigger, but the more common one is *live traffic*:
any single capped call marks the endpoint down immediately, and nothing any successful call
does can mark it up again.

**Seam B — the error TYPE is destroyed before routing, so a transient is indistinguishable
from bad output.**

- `ai_errors.go`'s own header documents the intended consumer:
  `// Usage in coordinator/dispatch: var aiErr *AIUnavailableError; if errors.As(err, &aiErr) { … }`.
- **That consumer does not exist.** `grep -rn "errors.As" --include=*.go platform/orchestration/
  platform/agentbase/ platform/messaging/` returns no `AIUnavailableError` consumer. The only
  user of the classifier is `work_item_failure_ladder.go:378`, which re-runs `isAIUnavailable`
  on the error **string** at the work-item layer.
- At the orchestration/step layer the type is flattened: sync path `coordinator.go:945` →
  `routeToErrorStepOrFail(…, fmt.Sprintf("step %s failed: %v", state.CurrentStep, err))`;
  async path `coordinator.go:3634` passes an already-flattened `errorMsg` to `routeToErrorStep`.
- Consequence, live: `agent_definitions` type='council-gate' has **17** `review_*`/`gate_*`
  steps with `config.error_step='complete_invalid'` (12 with none) — unchanged since the
  08-19 contribution measured it. One seat's transient 400 discards the **entire round**,
  including seats that already answered, while a seat returning *unparseable* output is
  tolerated as an abstention. **Garbage is survivable; a transient is fatal.**
  ⚠ nesting trap, from the 08-19 contribution and re-confirmed: `error_step` is inside
  `config`, not at the step level. Read at step level the same query says "(none) | 29".

This is a doc comment that enforces nothing — MEMORY's
`a-doc-comment-is-not-an-enforcement-mechanism` shape, in the file that *documents its own
missing consumer*.

**Seam C — nobody is told.** Every occurrence in the record was found by a lane chasing an
unrelated symptom. No alert, no work item, no named error class surfaced anywhere.

### The 08-23 event was the WRONG ACCOUNT — and that is a diagnosis trap, not a code defect

The 08-22→23 event (113 + 32 = 145 cap failures, ~16h) was resolved when the owner discovered
the fleet's key is **not on the Anthropic org that `platform.claude.com` opens by default** for
his address. Billing read `$0.00 spent / 0% used` against a $2,000 limit *while* the API was
refusing calls; the decisive column is **Organization settings → API keys → `Last used`**
("30+ days ago" on the org he was looking at, while `llm_call_log` had failures minutes old —
a failed call is still a use, so a live key can never read 30+ days). Recorded in
`~/.claude/…/memory/the-fleet-key-is-not-on-the-default-console-org.md` and the 308 lane's
NOTES (`7224021c7`). Not this lane's to fix, but it is why "the owner adds credit" can take
16 hours instead of 20 minutes, and it belongs in the bug file's §5 candidate 1.

### Constraint on the fix: a same-file passenger is already sitting in seam A's file

`git status` → `M platform/orchestration/actions/check_endpoint_health_action.go`. Another
session is mid-edit there (adding `CheckConfig: true` plus a long doc comment to
`CheckEndpointHealthInputSpec`, unrelated to this bug). This is the **same reason the 08-17
addendum declined to fix it** — and it is still true seven days later. A pathspec commit still
takes a same-file passenger, so any edit I make to that file commits theirs too, half-finished.
Any candidate touching that file must say how it handles this; `git stash` is banned and
hook-blocked, so "clean the tree first" is not available.

### Filed 090 rather than asserting the root cause unrun

Per CLAUDE.md's diagnosis norm — this is a cross-cutting structural claim about a shared seam,
which is the exact "always file" case. Diagnosis queue was **empty** before I filed
(`item_type='needs_diagnosis' AND status='awaiting_diagnosis'` → 0 rows), so no duplicate.

- intake correlation `0ec4ec81-62d4-4829-8aa7-a0ca6d8e9ad9`
- **run correlation `6c834cc7-de0d-4b1f-b283-d6b82b8dffda`** (the key the artifacts carry)

## 2026-08-24 — the probe CANNOT see the cap, so the 08-17 "falsifier resolved" conclusion does not hold

Reading `check_endpoint_health_action.go` to design the seam-A fix produced a finding that
corrects the bug file rather than adding to it.

`ai_endpoint_health` row `claude` has `ping_path='claude_ping'`, so the probe loop calls
`pingClaude` (`check_endpoint_health_action.go:125-126`). Its status switch (`:220-231`):

| status | result |
|---|---|
| 200 | `true` |
| 402 | `false` ("credits exhausted") |
| 401 | `false` ("authentication failed") |
| 529 | `true` (overloaded but reachable) |
| **default** | **`true`** — comment: *"any non-auth error means API is reachable"* |

**The usage-limit refusal is a 400** — the bug file's §1 says so explicitly and emphasises it
("Note **400**, not 429"). A 400 therefore falls to `default:` and the probe returns
**healthy, with no error message**.

Three consequences, and the third is a correction to another lane's recorded conclusion:

1. **The probe can never mark claude unhealthy for a cap.** Only live traffic can
   (`ai_actions.go:634`, the sole `update_endpoint_health` caller, which only ever passes
   `false`). So the flag is set by traffic and cleared by the probe.
2. **The probe is therefore a TIMER, not a health check** — for this condition. It clears the
   flag on its next tick regardless of whether the cap is still biting. That is why the wedge
   is bounded by `check_interval_seconds` (3600) and not by the outage: the fleet resumes
   dispatch after ≤1 hour whether or not the provider is still refusing.
3. **⚠ CORRECTION to the bug file's 08-17 addendum, "Falsifier resolved … the FIRST branch".**
   That addendum posed two branches — (a) intermittent cap, probe lost a coin-flip; (b) the cap
   bites the probe's model specifically — and resolved to (a) on the evidence that *"the
   12:09:53Z-due probe ran at 12:10:18Z and returned `healthy=true`"*. **That evidence cannot
   discriminate between the two branches**, because `pingClaude` returns `true` on a 400 by
   construction. A probe that is capped and a probe that is fine produce the identical row.
   The conclusion may still be true — the *other* evidence in the same addendum (93 of 99 live
   calls succeeding) does support an intermittent cap — but the probe result contributes
   nothing to it, and the "lost coin-flip" framing of the probe is wrong: **the probe cannot
   lose that coin-flip.** This is MEMORY's `a-pass-from-a-blind-check-outlives-the-blindness`
   shape, and the blind pass is now quoted in the bug file as a resolved falsifier.

**What this changes about the fix.** The 08-17 addendum's first suggested shape — *"require N
consecutive failures before marking unhealthy (a single sample should not gate a fleet)"* — was
aimed at the probe. The probe is not the writer of `false`; **live traffic is**, and there is
exactly one such writer. So an N-consecutive rule belongs at `ai_actions.go:626-637`, not in
the probe. Recorded because building the addendum's version as written would have hardened a
code path that never fires for this condition.

## 2026-08-24 — the planning subagent died on a usage limit, which is this bug's own shape one layer up

I delegated the fix plan to a `fable` planning subagent (the owner asked for fable). It
terminated early: *"Agent terminated early due to an API error: You've hit your session limit ·
resets 6:40pm (Europe/London)"*. Its last recorded action was that it was **about to** check
the council-decide abstention machinery and whether any success path writes health back up —
i.e. it died just before the two questions I had already answered first-hand above
(`diagnose_council_decide_action.go:310-325` for abstention,
`grep update_endpoint_health` for the single-direction writer).

**No plan content was produced, so nothing here rests on it.** I wrote the plan myself from the
evidence already in this file. Recorded because (a) an unrecorded dead subagent is exactly the
"a subagent's report is another doc" trap in reverse — the absence of a report is easy to
mistake for "it agreed with me" — and (b) it is worth noticing that this session's own tooling
hit a usage limit while diagnosing a usage limit. **They are NOT the same limit**: this is the
Claude Code *session* limit for this interactive session, not the fleet's `ANTHROPIC_API_KEY`
account cap that `bugs_open/243` is about. Conflating them would be the "wrong account" error
of `the-fleet-key-is-not-on-the-default-console-org` in a third variety. The fleet's own
liveness query at the same moment read 750 successes / 0 cap failures today.

## 2026-08-24 — the 090 run produced NO verdict, by a third mechanism, and I am taking the declared-substitute path

Run `6c834cc7-de0d-4b1f-b283-d6b82b8dffda` finished. Outcome:

```sql
SELECT (SELECT count(*) FROM diagnosis_artifacts WHERE correlation_id='<RUN>' AND kind<>'bundle') AS verdicts,
       (SELECT status FROM site_work_items WHERE spec->>'dispatch_correlation_id'='<RUN>') AS item;
--  0 | complete
```

`0 | complete` is the LANDMINES settling query's terminal answer: the run is over and wrote no
verdict artifact. **But the mechanism was not the one the landmine describes, and its
diagnostic check came back clean.**

- The landmine's tell is the body-budget omission line. I ran it on all five bundles:
  **no omission line in any of them** (lengths 48,860 / 98,827 / 111,150 / 110,674 / 118,627).
  So the "check the budget line" remedy — the one the landmine calls *"the check that actually
  discriminates"* — **returns clean on a run that still produced no verdict.**
- The actual stop was the **iteration cap**. It is recorded not in `diagnosis_artifacts` at all
  but on the work item's own `result`:
  `{"status":"UNVERIFIABLE","summary":"Diagnosis NOT confirmed (stopped: iteration-cap). Best-effort trail attached for a human; no fix proposed.","conclusion":"NOT CONFIRMED (stopped: iteration-cap). last hypothesis: <my symptom, verbatim>"}`.

So there are at least three ways a 090 ends without a verdict — body-budget truncation at
bundle 1 (2026-08-12), budget exhaustion at a later bundle on a single-symbol scope
(2026-08-24, bugfix_206), and now **iteration-cap exhaustion with no truncation at all** — and
the landmine's stated discriminator only sees the first two. The settling query sees all three
*as an absence*, which is right, but a reader who then goes looking for *why* will find the
omission line empty and have nothing. **The `UNVERIFIABLE` conclusion on `site_work_items.result`
is the missing half**; LANDMINES entry owed (below).

**What this means for the claim.** `UNVERIFIABLE` is not REFUTED — the loop neither confirmed
nor contradicted the mechanism; it ran out of iterations while working on it, and its last
hypothesis is my symptom quoted back verbatim. So the loop contributes **nothing** either way,
and I must not present it as support. Per the **owner ruling of 2026-07-31**, I therefore take
the ruling's named escape hatch explicitly rather than silently:

> **DECLARED SUBSTITUTE FOR THE 090 LOOP.** The 090 was run (intake
> `0ec4ec81-62d4-4829-8aa7-a0ca6d8e9ad9`, run `6c834cc7-de0d-4b1f-b283-d6b82b8dffda`) and
> returned `UNVERIFIABLE` on its iteration cap, contributing no verdict. In its place, every
> link in the chain was read first-hand from the live code and the live DB, with the citation
> inline at each step and each one individually falsifiable: the sole `update_endpoint_health`
> caller and its single direction (`grep`, one hit); the absence of any SQL trigger or second
> pg_proc on `ai_endpoint_health` (`pg_trigger` → 0 rows; `pg_proc` → 1 name); the sole writer
> of `healthy=true` (`check_endpoint_health_action.go:138`); the live `check_interval_seconds`
> (3600); the claim gate (`claim_work_item_action.go:226-263`); the absent `errors.As` consumer
> (`grep` over three packages, 0 hits); both flattening sites (`coordinator.go:945`, `:3634`);
> the live count of seats routing to `complete_invalid` (17, queried today); the abstention /
> unreadable split and the approval downgrade (`diagnose_council_decide_action.go:310-325`,
> `:460`); and `pingClaude`'s 400→`true` branch (`:220-231`). **No inferred link.** The two
> claims that are NOT first-hand are marked as such above: the monthly-limit shape `[INFERRED]`
> and the effective-spend arithmetic `[MEASURED input, DERIVED weighting]`.

## 2026-08-24 — council APPROVED round 1; two of three halves shipped; the third is blocked by a shared-tree hazard

**Verdict: APPROVED, round 1** — `82f07fa6-1c42-46ad-bdf6-1d58892c44a7`, *"approved with 4
advisory objection(s) — none high-severity"*, 6 abstained, **0 unreadable** (so not
truncation-gated, and no seat was lost — which is worth noting given what this fix is about).

### All four objections answered by QUERY, and two of them changed the code

| seat | sev | objection | answer |
|---|---|---|---|
| editquality | med | the migration's `review\_%` name filter could miss a `gate_*` seat | **Right in principle. Measured:** the 19 steps with `error_step='complete_invalid'` are the 17 `review_*` seats plus `council_decide` and `persist_submission` — so no `gate_*` seat is affected today. **Adopted anyway**: the migration now filters on `error_step` with those two named as exceptions, so a *future* `gate_*` seat is covered automatically. Their version closes the door; mine only happened to be standing in front of it |
| editquality | low | the field→step mapping is an unverified assumption | **Measured:** all 17 `review_fields`' first segments name a real step, 0 misses. Code splits on the first `.` (not `TrimSuffix(".result")`, which is what the sketch said) and the test pins the derivation |
| guardian | med | `routeToErrorStep` is fleet-wide, not council-scoped | **Right, and it produced a real change.** The map is now **capped at 50** with a `__truncated` marker: a loop expanding into many failing iterations makes a distinct step name per iteration, so an unbounded map would grow `collected_data` without limit on exactly the runs already going badly |
| reuse_agent | med | why a sibling key rather than making `__step_error` itself the map, "with a migration path for its few readers"? | **"Few" is the load-bearing word and it is wrong. Measured: 33 Go references outside tests + 6 live agent configs = 39 readers.** Changing its shape breaks all of them. Additive is the cheaper correct choice — and it is now a number, not a preference |

**And the one "missing" item was REFUTED by reading**, which is worth recording because it was
the most serious-sounding of the five: editquality said the `AIUnavailableError` consumer for
work-item retries is still absent, so *"work items still accrue attempt_count on a transient
400"*. They do not. `work_item_failure_ladder.go:378` already classifies this and issues a
transient release with a cooldown (`reason: ai_unavailable`). The consumer exists; it reads the
classifier by **string** rather than by **type**, which is untidy but not the defect claimed.
Scoped out with the citation rather than argued.

### The test is mutation-proven in BOTH directions

Not "it passed". `reviewStepFailed` forced to always-false → the *lost opinion* arm fails with
its own message; forced to always-true → the *skipped seat* arm fails with its own message.
Both reverted, suite green. A test that only failed in one direction would have permitted the
opposite defect, which here is "every gated-off seat blocks every approval".

### ⚠ THE THIRD HALF IS NOT COMMITTED, and the reason is a live hazard for the whole tree

`coordinator.go` in the shared working tree carries the **`bugs_open/354` lane's uncommitted**
call to `errorRouteTermination`, whose definition lives in `error_route_completion.go` —
**untracked** (`??`). Measured: HEAD's `coordinator.go` has **0** references, the working tree
has **1**.

```
$ ./scripts/verify-head-builds.sh --with platform/orchestration/coordinator.go
platform/orchestration/coordinator.go:4489:37: undefined: errorRouteTermination
verify-head-builds: FAILED
```

A pathspec commit takes the whole file, so committing my ~20-line `routeToErrorStep` hunk would
take their half-written change **without** the callee and break HEAD for the estate. **So I did
not commit it.** Consequences, all deliberate:

- the council-side reader shipped and is **inert** — `reviewStepFailed` fails closed with no
  `__step_errors`, giving exactly today's behaviour;
- migration `588` stays `_HOLD` and **must not be applied** until the writer lands;
- `bugs_open/354` now carries a note with the two ways out. **It is their change; I have
  touched none of their files**, including the untracked ones.

This is not only my problem and that is why it was worth writing down: **any** session that
commits `coordinator.go` for **any** reason breaks HEAD, and nothing at commit time would warn
them — the commit-scope report cannot see a same-file passenger, and their `git status` will
not volunteer another lane's untracked file.

### What shipped, and what it may NOT be said to have proven

`e521cde3e` — `Council-Reviewed: 82f07fa6…`. Registered as **MDL-044** (symmetric health
write) and **WFA-023** (`__step_errors`), both entered with their limits stated.

**Neither has a behavioural proof and neither may be described as working.** The estate was
quiescent all day (`llm_call_log`: 0 cap failures across 750 successful calls), so the
discriminating case has not occurred, and Go is inert until the next chassis roll anyway. The
proofs owed are in `PLAN` C1 and C4 — and both need a **demand control**: a green reading on a
day with no refusals is exactly what the pre-fix binary would also produce.

## 2026-08-24 (later) — chassis v1.0.1334 carries both shipped halves; the probe interval is APPLIED; the writer is still absent

### The roll: capability-probed, not inferred, on BOTH replicas

Chassis `v1.0.1334`, pods up 15:39Z. The `build provenance` startup line had already **scrolled**
out of `--tail=3000` — the documented case, and an empty result there means "not in range", not
"unstamped". So the check was a **capability probe** against `/proc/1/exe`, which has no shelf
life, with **both controls in the same breath** `[MEASURED 2026-08-24 ~16:0x Z]`:

| probe string | fr8dn | xl2zk | meaning |
|---|---|---|---|
| `failed to clear endpoint health after a successful call` | **1** | **1** | C1 / MDL-044 **PRESENT** |
| `recorded as unreadable, NOT abstained` | **1** | **1** | C4a **PRESENT** |
| `diagnose_council_decide` | 15 | 15 | present-control ✓ (the probe can find things) |
| `__step_errors_absent_control_xyz` | 0 | 0 | absent-control ✓ (it is not matching everything) |

Both controls behaved, so the two 1s mean something. **MDL-044 and the council-side reader are
LIVE on both replicas.**

### And the writer is confirmed ABSENT — which is the reading that gates migration 588

The reader and the writer both mention `__step_errors`, so that literal **cannot** discriminate
between them — probing it would return 1 either way and read as "the writer shipped". The
writer has its own literal (the cap marker), and that is the one to ask for:

| probe string | result | meaning |
|---|---|---|
| `step-error record capped at` | **0** | the coordinator **WRITER is absent** (still uncommitted) |
| `__step_errors` | 1 | the reader's key literal only — **not** evidence of a writer |

So the shipped state is exactly the designed one: the reader is live and **inert**, failing
closed with no `__step_errors` to read. **588's council half stays held.** Recorded because the
obvious probe here is the wrong one, and it fails in the direction that would have licensed
applying the migration.

### Migration 596 applied — the interval half, split out on the owner's instruction

`596_claude_probe_interval_60s.sql`, applied by hand and recorded `--record-only`.
**Deliberately not `--apply`**, which takes EVERY pending file fleet-wide — not just mine.

- pre-state: `claude` / `check_interval_seconds = 3600`, ledger rows for 588/596: **none**
- post-state: `claude` / **60**, matching the `cpu-ollama` row that already carried 60
- `BEGIN / UPDATE 1 / DO / COMMIT`, exit 0; the `DO` block RAISEs rather than SELECTs, so a
  bad state aborts the COMMIT

**588 was AMENDED rather than merely trimmed**: its header now records that the interval half
moved and where to. Without that a future reader could re-add it, and the ledger would carry
two files claiming one change — which is the shape the "a committed `_HOLD` migration is
indistinguishable from an applied one" landmine warns about, one step along.

### ⚠ The first "it moved" reading was NOT evidence, and I nearly took it as such

After applying, `last_checked` moved from `15:45:07` to `16:22:38`. My watcher reported
`PROBE MOVED` and exited, and the natural reading is "the new interval is working".

**That gap is 37½ minutes.** It is the OLD 3600s schedule finally coming due, and it is exactly
what the pre-change system would have produced. **One movement is consistent with BOTH
intervals**, so it discriminates nothing — the same shape as this lane's own finding about the
08-17 probe result, and the same shape as `two-clean-runs-cannot-establish-stability`.

The measurement that *can* come out otherwise is the **tick RATE**: sample `last_checked`
repeatedly over ~3 minutes and count distinct values. A 60s interval gives ~3–4; a 3600s
interval gives exactly 1. Result recorded below.

### The tick-rate measurement, and a correction it forced on my own migration `[MEASURED 2026-08-24 16:22–16:26Z]`

Sampled `last_checked` every 15s for 3 minutes and counted distinct values:

```
2026-08-24 16:22:38.297329+00
2026-08-24 16:24:12.988514+00
2026-08-24 16:25:44.134144+00
--- distinct ticks in ~3 min: 3 ---
```

**3 ticks, so the new interval IS being honoured** — a 3600s interval would have produced
exactly 1, which is what makes this measurement able to come out otherwise.

**But the gaps are 94s and 92s, not 60s** — and my own migration comment said the change
"bounds that worst case to about a minute". That was wrong, and it is now corrected in the
file (the applied SQL is untouched; only the prose claim changed).

**Why ~92s and not 60s:** the probe fires only when the scheduled task
`ai-endpoint-health-check` ticks **and** the endpoint's own `check_interval_seconds` has
elapsed since `last_checked`. That task's own `interval_seconds` is **also 60**
(`scheduled_tasks`, enabled, last_triggered 16:27:07Z), so the two compose: a tick at T only
probes if `last_checked + 60s <= T`, which in practice means roughly every other tick.

**So the honest bound is one to two minutes, phase-dependent.** Still a ~39× improvement on
3600s, which is the whole point — but "one minute" is the kind of round number that gets
quoted forward for ever, and it is not what the system does. If a tighter bound is ever
wanted, the lever is **this row at 30s** (the value `gpu-ollama` already carries) against the
task's 60s tick — not lowering the endpoint interval alone, which cannot beat the tick.

**Two things I got to check because I distrusted a green reading**, both worth keeping:
the first "PROBE MOVED" was a 37½-minute gap (the old schedule coming due) and proved nothing;
and the corrected figure came from asking *how often*, not *whether*.

## 2026-08-24 (evening) — MDL-044 is PROVEN on live traffic, after four attempts of which three proved nothing

Forced the endpoint unhealthy by hand and watched whether a successful live call would clear
it. Took five attempts. **Attempt 5 is the proof; attempts 1–4 are the lesson.**

### The proof `[MEASURED 2026-08-24 16:51:32 → 16:52:34 Z]`

```
t0 (forced healthy=false)  : 2026-08-24 16:51:31.969
probe last ran BEFORE t0   : 2026-08-24 16:51:07.419
CLEARED after ~39s         : healthy=YES
                             last_checked  = 2026-08-24 16:51:33.504   <- my forced write, UNMOVED since
                             last_healthy  = 2026-08-24 16:52:34.402   <- 61s LATER than last_checked
                             error         = NULL                       <- was my forced message
                             ok_calls_since_t0 = 1                      <- the demand control
```

**Attribution is airtight, on four independent grounds, and the disconfirming result was
available:**

1. **`last_healthy` is 61 seconds LATER than `last_checked`.** Every other writer in the
   system sets the two together — the prober's UPDATE assigns
   `last_checked = NOW(), last_healthy = CASE WHEN $2 THEN NOW() END` in one statement, and so
   does `update_endpoint_health`. **Only the new writer sets `last_healthy` without touching
   `last_checked`**, so that ordering is a signature no other code path can produce.
2. **`last_checked` never moved from my own forced write**, so **no probe ran** in the window.
   The prober cannot be the cause.
3. **`error` went to NULL** — the new writer's `error = NULL`, against the message the forced
   write had put there.
4. **The demand control fired: exactly one successful call**, and the clear coincides with it.

Had the prober been responsible, `last_checked` would have moved and `last_healthy` would
equal it. That is the result this test could have produced and did not.

### ⚠ Attempts 1–4: three of them proved nothing, in two different ways, and I nearly filed the second as a broken mechanism

**Attempts 2 and 4 were VACUOUS — zero demand.** Both reported "still false after 50s", which
reads exactly like a broken mechanism. Then I checked traffic: **0 successful LLM calls in
either window.** There was nothing for a clear-on-success writer to react to. This is
`a-post-fix-zero-needs-a-demand-control` from the memory index, walked into twice in ten
minutes — and the fix was to move the control *inside* the experiment
(`ok_calls_since_t0` in the same query as the row), not to remember to check afterwards.

**Attempt 1 and my detector: the check was STRUCTURALLY INCAPABLE of reporting success.**
Attempt 1 had real demand — 5 successful `landmine-verifier` / `council-gate` calls — and my
loop reported "DID NOT FLIP". It had not failed; **my test could not see it.**

```sql
SELECT healthy, healthy||'', (healthy)::text FROM ai_endpoint_health WHERE name='claude';
-- t | true | true
```

`psql -t -A` renders a bare boolean column as **`t`**, but the moment it is concatenated
(`healthy||'|'||…`) the cast renders **`true`**. My loop built a concatenated row and then
tested `[ "$H" = "t" ]`. **That comparison can never be true**, for any state of the system.
So attempt 1's negative was manufactured, attempt 3's positive came only from a DB read I did
*afterwards* out of suspicion, and attempt 4's `case "$R" in t\|*)` carried the same defect.

**I was one step from writing "the mechanism did not fire" into `bugs_open/243`** — a false
claim about my own shipped code, on the day it shipped, supported by a test that could only
ever say no. What caught it was not the harness: it was noticing that
`last_healthy > last_checked` in a state dump, which is impossible unless the writer *had*
fired.

**The cheap check, and it is now the first line of the experiment: print the rendering control
next to the comparison.** One `SELECT healthy, healthy||''` shows the two forms are different
strings. Better still, make the query answer in a vocabulary you chose —
`CASE WHEN healthy THEN 'YES' ELSE 'NO' END` — so the value can never depend on a driver's
boolean rendering. Logged in `WRONG_CALLS.md`.

### Status after this

**MDL-044: LIVE + PROVEN on live traffic** (chassis v1.0.1334). The register entry's
verify-later is discharged; its "may not claim the mechanism has ever fired" caveat is lifted,
and only for this half.

**WFA-023 is unchanged: reader live and inert, writer still absent, 588 still held.**

## 2026-08-24 (evening, last) — a second independent firing, and a claim I nearly made off ONE observation

**Second independent proof of MDL-044, unprompted by me** `[MEASURED 2026-08-24 16:54Z]`: a
routine state read showed `healthy=f`, and moments later `healthy=t` with
`last_checked = 16:54:00.987` but `last_healthy = 16:54:11.817` — the same
`last_healthy > last_checked` signature, matching a successful `council-gate` call at
`16:54:11.836` to **19 milliseconds**. Nobody forced anything; the probe had set the row false
and a live call healed it in ~11s. That is the mechanism doing its job in production.

**And here is where I nearly overclaimed.** From that single event I formed the reading *"the
probe is intermittently marking claude unhealthy, so the pre-fix system was losing fleet
dispatch far more often than the 7-cap-days figure suggests"* — which would have been a
striking finding, and would have gone in the bug file.

So I sampled the row every 3s for 2 minutes before writing it:

```
samples: 40      healthy=T: 40      healthy=F: 0
```

**Zero false readings.** The claim is **not supported**, and the sample also cannot refute it —
at a ~92s probe cadence, two minutes covers barely one or two probe ticks, and any false the
writer healed inside my 3s sampling gap is invisible to it. **So the honest statement is: one
probe-induced `false` was observed, and the rate is `[UNMEASURED]`** — my sample was
structurally incapable of bounding it in either direction.

What it *does* establish, and this is the useful part: the row was false and **self-healed in
seconds without anyone acting**. Before MDL-044 that same event would have held the fleet's
claim gate shut until the next probe tick — up to an hour before today's migration, and one to
two minutes after it. **Measuring the rate properly needs a sampler at ~1s over hours, or a
history table; neither exists, and inventing one is not this lane's job.** Recorded so the next
session neither repeats my inference nor mistakes my two-minute sample for a bound.

**Also for the record — a same-file passenger, in the direction that lands on me.** My edit to
`docs026_concept_register/register/000_concept_index.md` was swept into another session's commit
`51293533e` ("333 round-2 record") before I committed it. The content is at HEAD and nothing is
lost; forward-only holds. It is the exact scenario CLAUDE.md describes — committing per task
stops *me* sweeping *others'* work, and cannot stop theirs sweeping mine.

## 2026-08-24 (night) — UNBLOCKED. All three halves are committed; the writer awaits the next roll

The `bugs_open/354` blocker was cleared by an **owner-approved sweep**, after establishing it
was **parked, not in progress**: both untracked files last modified **2026-08-22 19:20–19:21**
(two days), and no commit touching that lane since — the only 354-related commit in the window
was my own note.

**What I verified before committing code I did not write** (the bar for vouching, not a
formality):

- builds — `verify-head-builds` with all three files, OK; and HEAD green after both commits;
- **their own tests pass in full** in a clean HEAD tree (declared-terminal recorded, recovered
  untouched, skip untouched, undeclared untouched, outcome read strictly, malformed marker
  inert across 5 shapes, prefix not doubled);
- **read in full**, including the body: `errorRouteTermination` returns false unless the ending
  step declares `config.outcome == "error"` **AND** a non-empty `__step_error.message` exists —
  inert by construction.

**Two commits, so no intermediate state fails to compile:**

1. `893a12d47` — `sweep:` their two callee files alone. HEAD's `coordinator.go` had no call at
   that point, so HEAD still built.
2. `dbd865ee8` — my `routeToErrorStep` writer, **declaring their `completeWorkflow` hunk as a
   same-file passenger**. A pathspec commit takes the whole file; the alternative was editing
   another lane's hunk out and back on a shared tree, which is worse.

### ⚠ A finding for THEM that I nearly accepted at face value

Their file's measurement table reads `36 dishonest (declared terminal + marker) -> recorded
100%`, which reads as measured against live config. **It is not live.** Checked at all three
placements today — `v->'config'->>'outcome'='error'` = **0**, step-level `v->>'outcome'='error'`
= **0**, and the string `"outcome"` anywhere in any live `default_config` = **0**.

So their change is the **code half of a code/config split whose config half does not exist**,
and `bugs_open/354` is **not fixed** — the discriminator is present and unreachable. Told them,
with the sequencing advice from this lane transferred. I checked all three placements
specifically because the `error_step` nesting trap bit this lane earlier today; one placement
would have returned a confident 0 that happened to be right for the wrong reason.

### Where 243 stands now

| half | state |
|---|---|
| **MDL-044** symmetric health writer | **LIVE + PROVEN TWICE** on v1.0.1334 |
| probe interval 3600→60 (mig 596) | **APPLIED**, measured cadence 92–94s |
| **WFA-023** reader | **LIVE** on v1.0.1334, inert by design |
| **WFA-023** writer | **COMMITTED** (`dbd865ee8`), **inert until the next roll** |
| mig `588_..._HOLD` (council repoint) | **STILL HELD** — gate below |

**The gate before 588 may be applied, and it is a threshold not a glance:**
`grep -ac "step-error record capped at" /proc/1/exe` must read **≥ 1 on EVERY replica**, with a
known-absent control in the same breath. **Never probe `__step_errors`** — the reader mentions
it too, so it returns 1 whether or not the writer shipped.

**The proof still owed after that:** a council round in which one seat errors must reach a
verdict, report that seat under `unreadable` (not `abstained`), and — if the rest would have
approved — return **REVISE** naming the lost seat. A round approving with zero unreadable is
the negative control.

## 2026-08-25 — the last half is LIVE: mig 588 applied on v1.0.1337, and the SQL was wrong the first time

### The gate, checked as the file demanded rather than by glancing at a tag

Chassis **v1.0.1337** (pods up 09:27Z). Two independent confirmations:

- **Ancestry against the running binary's own stamp.** The `build provenance` line was in range
  this time: `git_commit 4c996e1b5cb9b2513d88ec9fe2bae220c38fb6c2`. Then
  `git merge-base --is-ancestor <commit> 4c996e1b5` → **IN** for all three (`dbd865ee8` writer,
  `e521cde3e` reader, `893a12d47` sweep). This is the query CLAUDE.md prescribes and it beats a
  marker hunt.
- **Binary probe, both replicas, both controls:** `step-error record capped at` = **1/1**
  (the writer), `diagnose_council_decide` = 15/15 (present control), an invented string = 0/0
  (absent control). `__step_errors` also read 1 — **and that is not evidence**, which is why the
  cap marker is the discriminating probe.

### ⚠ THE FIRST APPLY FAILED AND ROLLED BACK — my SQL, not the plan

```
psql:<stdin>:72: ERROR:  invalid reference to FROM-clause entry for table "ad"
LINE 8: FROM LATERAL jsonb_each(ad.default_config #> '{workflow,step...
HINT:  There is an entry for table "ad", but it cannot be referenced from this part of the query.
```

The reviewed sketch used `UPDATE agent_definitions ad … FROM LATERAL jsonb_each(ad.default_config …)`.
**Postgres refuses that outright**: the UPDATE target cannot be referenced from a `LATERAL` in
its own `FROM`. `ON_ERROR_STOP` aborted inside the transaction; verified afterwards that
**19 seats still read `complete_invalid`**, i.e. nothing changed. The `BEGIN`/guard/`COMMIT`
discipline did exactly its job.

Rewritten to a **correlated scalar subquery** (which *may* reference `ad`), rebuilding the whole
steps object in one `jsonb_object_agg` pass. **The rule is unchanged from the reviewed version;
only the SQL expressing it.**

**And the lesson generalises: a council seat reviews a SKETCH, and a sketch is not executable.**
This class of defect cannot be caught at the gate and no seat should be blamed for missing it —
it is the SQL sibling of the existing `go-build-cannot-parse-your-sql` lesson. **Compile-check
any SQL you submit.**

**Second defect in the same file, also mine:** it omitted `snapshot_agent`, which CLAUDE.md
states opens every migration touching `agent_definitions`. Added; the successful apply captured
`source_version=2` as `be2a7614-9096-4425-adba-55b0cd730756` before changing anything.

### Post-state, measured

| | |
|---|---|
| review seats routing `error_step` → their **own** `next_step` | **17 of 17** |
| steps still carrying `complete_invalid` | **2** — `persist_submission`, `council_decide` |

Spot-checked: `review_editquality` → `review_constitution`; `review_guardian` → `council_decide`;
both terminals unchanged. That is exactly the intended shape.

### Renamed off `_HOLD` in the same commit as the apply

`git mv` to `588_council_seat_transient_costs_one_seat.sql`, **both pathspecs named in the
commit** so no stale twin survives at HEAD — verified `git ls-tree HEAD` returns the new name
only. This heeds the landmine another session filed the same week: inside the repo an applied
`_HOLD` file is indistinguishable from one still waiting, and the filename is the only tell.
The runner also refuses `--record-only` on a `_HOLD` name (SIDECAR_RE), which is what forced
the issue — a useful bit of mechanism to know.

### NEGATIVE CONTROL DISCHARGED, positive arm still owed

`[MEASURED 2026-08-25 09:49:00Z]` The first council round to run after the migration reached
**`complete_approved`** — decision `approved`, **5 abstained, 0 unreadable**. So repointing all
17 `error_step`s did **not** break the ordinary path: a round where nothing errors is
unaffected, exactly as intended.

**That is the regression check, not the proof.** The positive arm — a seat whose call *errors*
producing `unreadable > 0`, a verdict rather than `complete_invalid`, and REVISE if the rest
would have approved — needs a real provider transient and **cannot be forced**. It will arrive:
cap failures hit 7 of the 15 days to 08-24. Do not fake it; do not close on its absence.

## 2026-08-25 (evening) — post-roll re-verification on v1.0.1339, and a zero that is NOT a result

A roll is the moment to re-check, not to assume: Go can regress via a build from an older ref
(the estate has hit this), and DB config can be overwritten by a seed.

**Nothing regressed** `[MEASURED 2026-08-25 ~19:15Z]`:

- chassis **v1.0.1339** (pods 19:07Z), provenance stamp `a7459a44b68b8c67b7d7bb0ca7c064e0729d59f5`;
- `git merge-base --is-ancestor <c> a7459a44b` → **IN** for `dbd865ee8`, `e521cde3e`, `893a12d47`;
- DB config survived: probe interval **60s**, seats repointed **17/17**, `complete_invalid`
  still exactly **2** (`persist_submission`, `council_decide`).

### 47 rounds, zero deaths — and why that is NOT the proof

Since mig 588 applied this morning: **27 `complete_approved` + 20 `complete_revise` + 1 in
flight, and ZERO `complete_invalid`.** Against the 08-19 measurement of roughly a coin-flip per
round dying there, that looks like a decisive win.

**It is not, and the reason is the demand control — for the third time in this lane.**
`llm_call_log` shows **0 cap failures on 08-24 and 0 on 08-25**. No transient has occurred, so
nothing has exercised the new path; all 47 rounds report `unreadable = 0`. **The zero means
"nothing tested it", not "it works".**

Recording it in those terms deliberately: 47-of-47 is exactly the kind of figure that gets
quoted as proof, and this lane has already produced two vacuous zeros and one structurally
incapable detector. What 47-of-47 *does* establish is **no regression on the ordinary path** —
which was the real risk of repointing seventeen `error_step`s, and is worth having.

### ⚠ A prediction of mine that has NOT materialised

I wrote `[INFERRED]` on 08-24 that the cap is a **monthly** limit and that recurrence was likely
before 08-31, from the reset-date shape plus the month-end clustering.

| day | cap | ok |
|---|---|---|
| 08-21 | 3 | 1,223 |
| 08-22 | **113** | 1,063 |
| 08-23 | 32 | 1,109 |
| 08-24 | **0** | 1,850 |
| 08-25 | **0** | 1,395 |

**Two clean days at the two highest call volumes in the whole record.** Most likely the account
was properly funded once the wrong-account error was found on 08-23 — which is the same shape as
the 08-10 event, where the owner acted and the fleet came back 21 days before the stated
calendar date.

**Two days cannot settle it**, and six days of the month remain, so the prediction is
**unresolved, not refuted** — and I am not upgrading it to "fixed" on a quiet fortnight either.
The honest status is: the trigger has been absent for 48 hours, cause plausibly known, and the
figure goes stale by the day. **Re-run the histogram before repeating any of this.**

## 2026-08-25 (late) — the positive arm has NOT landed, and the four near-misses sharpen what it actually requires

`[MEASURED 2026-08-25 ~21:20Z]` Since mig 588 applied: **60 council rounds completed**
(38 `complete_approved`, 22 `complete_revise`, 4 in flight), **ZERO `complete_invalid`**, and
**zero rounds with `unreadable > 0`**.

**But there WERE four council-seat LLM failures in that window**, so the absence is not simply
"nothing went wrong":

| time | agent | step | error |
|---|---|---|---|
| 12:18:52 | council-gate | `review_debug_historian` | `TOLERATED (step continued on the partial): response truncated: stop_reason=max_tokens` |
| 12:49:49 | council-gate | `review_debug_historian` | same |
| 16:19:49 | council-gate | `review_debug_historian` | same |
| 21:14:05 | council-gate | `review_debug_historian` | same |

**None of them could have exercised this fix, and the reason is worth writing down.** They are
`success=false` rows in `llm_call_log`, which makes them *look* like the case the fix is for.
They are not: **`TOLERATED (step continued on the partial)`** means the truncation machinery
(`bugs_open/019` / `076` / `138`) salvaged the partial *upstream*, the step **completed**, and it
therefore **never routed to `error_step`** — so `routeToErrorStep` never ran, `__step_errors` was
never written, and there was nothing for `reviewStepFailed` to find. `diagnose_council_decide`
saw a normal (degraded) review, not an absent field. Hence `unreadable = 0`, correctly.

### ⚠ Refinement to the close condition — the proof needs a TERMINAL seat error, not any failure

The condition as previously written ("a seat whose call errors") is too loose and would let a
future session read these four rows as the proof. Precisely:

> The positive arm requires a seat whose call **fails terminally enough to route to its
> `error_step`**. A truncation that is tolerated, salvaged, or degraded does **not** qualify — it
> is absorbed before the error route. In practice the qualifying cases are the provider refusing
> us (the 400 this bug is about), a connection-level failure, or an unhandled 4xx/5xx.

**The `llm_call_log` shape does not discriminate this.** A tolerated truncation and a terminal
failure are both `success=false`. **Check the ORCHESTRATION, not the call log:** the round must
reach a verdict with `unreadable > 0`, or (pre-fix behaviour) `complete_invalid`.

### One observation handed on rather than chased

`review_debug_historian` truncated **four times in one day** and was the only seat to do so —
each time tolerated, so no round was lost. That is not this lane's bug and I have not
investigated it, but a single seat repeatedly hitting `max_tokens` is the shape `bugs_open/019`
and `138` care about, and its `max_tokens` may simply be undersized for what its prompt now
carries. Recorded here so it is not lost; **not** filed as a bug on four observations.
