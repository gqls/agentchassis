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
