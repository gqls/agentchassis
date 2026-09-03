# PLAN — fleet dispatch throughput & per-site turn latency

**Started 2026-08-18** (seeded from the idea.uk lane's question "how do we speed up the
drain rate?"). ~~**Owner: unclaimed — whoever picks this up owns it.**~~
**CLAIMED 2026-08-19 by the "throughput" session, at the owner's direct request — and the
scope is WIDENED**: the owner asked for the whole-architecture scale review (several
thousand domains, promotion-driven onboarding bursts of many domains/day), which brings
forward the review the site_delivery lane had seeded for "after the working site" — the
owner's direct ask supersedes that sequencing. The standing docs now exist in this
directory; baselines were re-run 2026-08-18 evening (NOTES). The wider deliverable is
`RESEARCH_2026-08-18_throughput_to_thousands_of_domains.md`; this PLAN's Phases 1–4 stand
unchanged as the dispatch-specific slice (Phase 1 in flight: 090 run
`a16b82cd-b89a-45d5-b5df-4370c754e2fd`). Owner scope ruling 2026-08-18: research +
diagnosis only — Phase 2+ config changes wait on decisions D0/D1 (see the RESEARCH doc's
decision table, which extends this file's §9).

**Companion evidence file (read first):**
`STARTER_2026-08-18_fleet_dispatch_drain_rate.md` — the measurements, the mechanism,
and the traps. This file is the *what to do*; the starter is the *why*.

---

## 1. The problem, in two numbers that must not be conflated

| target | today | mechanism |
|---|---|---|
| **T1 — fleet throughput** | ceiling ≈ **83 items/hour** for the whole fleet | one dispatch turn at a time: ~218 s per productive trigger run × ≤5 items per turn |
| **T2 — per-site turn latency** | hours (idea.uk waited ~3.5 h on 08-17) | strict fleet-wide FIFO by item age (migration 284): a new item queues behind every older item on every other site |

They have different levers. Raising T1 (Phases 2–3) shortens T2 as a side effect;
changing service *order* (Phase 4) moves T2 only, and is an owner decision.

**Root mechanism (from the starter, §2):** `build-pipeline-trigger`'s
`max_concurrent: 8` is dead config — `countInFlight` (`cmd/scheduler/main.go:414`)
counts `scheduled_tasks` **rows** per group (the `dispatch` group has one enabled row),
and `loadDueTasks` (`main.go:368`) separately enforces one execution per task row.
Median handler runtime is 36 s; the items are fast, the turns are scarce.

## 2. Scope

**IN:** the scheduler's concurrency semantics, the trigger/loop agent configs
(`build-pipeline-trigger`, `build-dispatch-loop`), batch size, service order.

**OUT — explicitly:**
- **`bugs_open/083` (unroutable `detected` rows, 698 fleet-wide).** Owned by the
  `bugfix_277_required_fields_repair` lane, active, council-approved rounds in flight.
  Routability is "can this row ever be dispatched"; this workstream is "how fast do
  dispatchable rows drain". Contribute observations into 083's file; do not fix it here.
- **Handler-side speed.** p50 is 36 s; not the constraint.
- **`report-dispatch` / `diagnose-dispatch` groups.** Same machinery, different queues;
  measure them only if someone claims they are slow.

## 3. Phase 1 — file the diagnosis, reconcile the record (no changes yet)

1. **Run the `090` on the structural claim** before it goes into any `bugs_open/` file
   (owner ruling 2026-07-31 — this is exactly the cross-cutting class). Symptom to file:
   mechanism = "`max_concurrent` on `scheduled_tasks` is compared against a count of
   task *rows*, not executions, so a group with one enabled row can never run
   concurrently regardless of the configured cap; a separate per-task guard in
   `loadDueTasks` enforces single-flight per row" — point at `cmd/scheduler/main.go`
   (`countInFlight`, `loadDueTasks`, the cap test at `:181`), `scheduled_tasks`,
   `orchestration_states`. No counts in the symptom; the loop fetches its own.
2. ~~**Reconcile `bugs_open/029`.**~~ > **CORRECTED 2026-08-19 (pickup session): this
   step's premise was stale when written — 029's file has carried the exact
   reconciliation since 2026-07-21** ("CORRECTED DIAGNOSIS 2026-07-21 … the title
   mechanism is wrong; this is bug 003's blast radius, not a concurrency-group bug",
   including the countInFlight row-count analysis). Caught by reading the file before
   editing it, after `who-owns` showed six 029 commits on 2026-08-19 alone. Nothing to
   correct; do not edit 029 — it is ACTIVELY OWNED (part A fixed + behaviourally proven
   2026-08-18, live v1.0.1309/1314).
3. ~~**Check the wedged-loop diagnosis**~~ > **CORRECTED 2026-08-19: owned — do not
   re-file.** The 029 lane ran a 090 on the wedge the morning of 08-19, found it refuted
   on a false premise (`awaited_requests` retains 7 days and held 20 instances), and has
   narrowed the death to `continueExecution` for `iter_N+1_call_handler` on the
   response-consumer goroutine (`cfdb247c4`). This workstream's stake stands: one wedged
   loop stalls the fleet for up to `timeout_seconds` per tick under single-flight —
   track their fix, contribute measurements into their file if dispatch metering
   surfaces a wedge.

**Exit criteria:** 090 verdict recorded; 029 corrected or confirmed; NOTES started.

## 4. Phase 2 — concurrency via sibling task rows (the ~N× lever, config-only)

**Recommendation: N task rows, not scheduler code.** The scheduler's only execution
state is one `(last_triggered_at, last_completed_at)` pair per row — it structurally
cannot track N in-flight executions of one row. Sibling rows make each row its own
single-flight slot, which is the semantics the code already enforces correctly. The
"proper" per-task `max_concurrent` fix needs an executions table — file it as the
long-term shape in the 090/029 record, don't build it first.

> **CORRECTED 2026-08-25 (session 2; evidence NOTES 2026-08-25 §5): the premise of this section is
> FALSE on the running scheduler and has been since 2026-03-17 (`892a289e9`).** `cmd/scheduler/main.go`
> `runTick` calls `stampCompleted` — `SET last_triggered_at = NOW(), last_completed_at = NOW()` —
> immediately after `fireTrigger`, so a `fire_message` row is never in flight: the per-row guard,
> `countInFlight`, `max_concurrent` and `timeout_seconds` are all inert, and the three
> `notify_scheduler*` stamps the "Precondition" below defuses are inert too (583's stated failure —
> "fires every 300 s, releases the original early" — could not occur). Measured: a row overlaps ITSELF
> (361 / 322 pairs in 24.5 h, min gap 0.25 s); fire cadence p50 90 s vs run p50 97 s. What 584 actually
> did: doubled the FIRE rate, ~1 s apart — the second fire co-picks the same site **94%** of the time
> (the selector cannot see a claim that lands p50 17.7 s later), so lost claims run at 39%, the gain is
> ≈ **+10–15%** claims/h (not "~N×"), and the real effect is TWO handlers on the deep site. Of the
> safety argument below: the ATOMIC-CLAIM half holds (0 double-handles in 2,579 handlers; 775 same-site
> concurrent pairs; fail rate LOWER with a partner, 1.55% vs 3.85%); the "sites interleave by
> construction" half is false at 1 s spacing and true at ≥30 s spacing (p90 time-to-first-claim
> 24.2 s). The native rate knob is `interval_seconds` (30 s tick: 60 → 90 s cadence, 30 → 60 s,
> ≤25 → every tick). Owner decision pending on which lever stays — README_where_we_are 2026-08-25.

**Precondition — the stamp hardcode (the trap that blocks the naive version):** both
agents' `notify_scheduler` steps run
`UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = 'build-pipeline-trigger'`.
A sibling row would never be stamped → its single-flight guard falls through to
`timeout_seconds` → it fires every 300 s instead of 60, and every sibling's completion
stamps the *original* row, releasing it early. Fix order matters:

1. **Migration A:** add `"task_name": "build-pipeline-trigger"` to the existing row's
   `input_data` (inert — nothing reads it yet).
2. **Migration B:** change the notify queries in **both** agents to
   `… WHERE name = $1` with `"params": ["input_data.task_name"]` —
   `QueryDatabaseAction` (`database_actions.go:31-73`) resolves `params` paths from
   collected_data with an `input_data.` fallback, so **no Go change is needed**.
   ⚠ A nil param is a step **error** (`"query param path resolved to nil"`), which is
   why A must land before B. ⚠ `build-dispatch-loop` gets `task_name` via the trigger's
   `call_dispatch.input_mapping` — add `"task_name": "input_data.task_name"` there in
   the same migration, or the loop's notify errors.
3. **Migration C:** insert ONE sibling (`build-pipeline-trigger-2`), same group/agent/
   topic/pre_query, its own name in its `input_data.task_name`. **N=2 first.** Measure
   (§7) for 24 h before adding more.

**Safety argument to verify, not assume (put the check in NOTES):**
- *Per-site serialisation survives:* `find_dispatchable_site`'s
  `NOT EXISTS (… status='claimed' … same site)` excludes a site the moment one loop
  claims there; two triggers firing in the same instant can pick the same site, but
  `claim_work_item` is atomic and the loser's loop exits via `check_claim → done`. Cost
  = a wasted spawn. **Induce it once** (two manual dispatches at one site) and confirm
  no double-handled item — `count(*) GROUP BY id HAVING count(*)>1` on claims is not
  possible from the schema, so verify at `claimed_by`/`attempt_count` and the handler's
  artefact.
- *Sites interleave by construction:* once site A holds a claimed row, the next trigger
  run's selector skips it and takes the next-oldest site. Concurrency across sites
  emerges without any new ordering logic.

**Council/authorisation:** config-only, so the gate refuses it client-side — this is
the 284 precedent: **owner authorises directly**, register the change (WDS-002's file
is the model), consumers named. The consumers are every site's queue (service becomes
N-at-a-time; FIFO *between* turns unchanged) and anything metering LLM spend — N× turns
≈ N× handler spawns ≈ N× spend rate while backlog exists. **⚠ OWNER DECISION D1: target
N and acceptable spend rate. Do not pick N>2 without it.**

**Rollback:** disable sibling rows (`enabled=false`). Migrations A/B are inert-forward
and need no rollback to restore old behaviour.

## 5. Phase 3 — batch size (small, honest gain; do second, not first)

Throughput per turn ≈ `batch / (batch × 36 s + ~40 s overhead)`. Batch 5 → ~83/hr;
batch 8 → ~88/hr; asymptote 100/hr. **Alone it buys ~10–20%** — do it after Phase 2,
where it multiplies by N instead.

- Move `load_items.max_items` **and** `process_item.max_iterations` together (they are
  both 5; one without the other is a silent no-op or a truncated loop).
- A longer turn crosses `scheduled_tasks.timeout_seconds=300` (runs already average
  218 s): raise it to 600 in the same migration or the scheduler re-fires into a live
  run — which after Phase 2 is indistinguishable from a sibling firing, but *before*
  Phase 2 it is the untested overlap. Ceilings above that: `call_dispatch.timeout_seconds`
  (900) and the chassis-side 1200 s handler timeout.
- Batch 8, timeout 600 is the recommended pairing. Config-only, live immediately,
  same authorisation route as Phase 2.

## 6. Phase 4 — service order (OWNER DECISION D2 — prepare, do not implement)

Only reach here if p90 turn wait (recipe 4 in the starter) is still unacceptable after
Phase 2 — with N concurrent turns, expected wait ≈ (sites-with-work / N) × turn length,
which at N=4 is minutes, not hours. If it is still needed: the two candidate shapes are
round-robin over sites and an aging function (`created_at + priority × k`); migration
284 already states that `k` is an owner-agreed constant and that any priority-major
order recreates starvation. Bring the owner a measured p90 table and both shapes
costed; **do not pick `k` yourself.**

## 7. Verification — the meters, and what disproves each phase

All in the starter's §7; the two that are load-bearing:

- **The concurrency meter is `count(DISTINCT site_id)` of claims per MINUTE.** Today it
  reads 1. After Phase 2 with N=2 it must read 2 in busy minutes. ⚠ A 5-minute bucket
  reads 2–6 *today* and proves nothing — the starter's §2 trap. If the per-minute meter
  still reads 1 after Migration C, the sibling never fired: check its
  `last_triggered_at` advances and that its stamps land on ITS name (`SELECT name,
  last_triggered_at, last_completed_at FROM scheduled_tasks WHERE name LIKE
  'build-pipeline-trigger%'`).
- **Throughput is completions/hour over a 24 h window** (recipe 3), compared to a
  pre-change 24 h window at similar backlog. A quiet queue reads as a failed fix —
  check `triaged+approved` depth in both windows before concluding anything
  (`a post-fix zero needs a demand control`).

Predicted, falsifiable: Phase 2 at N=2 ≈ 1.8–2× completions/hour under sustained
backlog; Phase 3 adds ~1.5× on top (batch 8 with the longer turn). If measured gain is
under half of predicted, something else binds (chassis capacity, Kafka, LLM rate
limits) — measure at the chassis before adding more siblings.

## 8. Standing traps for this workstream

- The scheduler is **its own binary** (`cmd/scheduler` → image `aqls/kafka-scheduler`);
  if Phase 1's findings ever force a code change: commit first (`build-*` takes HEAD),
  bump `IMAGE_TAG`, and verify with the binary probe + controls, not `strings`, not the
  scrolled-away provenance log line.
- `scheduled_tasks` config is **live immediately**; agent_definitions changes are live
  on the next orchestration. Snapshot before UPDATE (`snapshot_agent(...)`), md5-guard
  the UPDATE (284 is the worked example).
- `apply -k` resets replicas and a roll kills in-flight council runs — do not roll the
  fleet for a config-only phase.
- Queue depth is not a prediction, and `ORDER BY created_at` head-age is not proof of
  head-of-line blocking (both burned the idea.uk lane; its 08-16 handoff §7).
- The dispatch pipeline has a **known freeze mode** (the 08-18 wedged-loop diagnosis).
  Any "throughput fell to zero" reading during this work should check for one wedged
  `build-dispatch-loop` in `AWAITING_RESPONSES`/`EXECUTING_STEP` before blaming the
  change (`bugs_open/029`'s diagnostic queries still work for *finding* them).

## 9. Decisions needed from the owner

| id | decision | default if unanswered |
|---|---|---|
| D1 | target concurrency N (spend rate multiplies with it) | stop at N=2 |
| D2 | service-order policy + aging constant, if Phase 4 is ever reached | stay FIFO |
| D3 | is overlap past `timeout_seconds` acceptable, or must batch+timeout always move in lockstep? | lockstep |

---

# D4 — LLM SPEND GOVERNOR: design (added 2026-08-31, on the owner's shedding ruling)

**Ruling (2026-08-31, verbatim in NOTES):** shed **routine maintenance FIRST, new site
builds SECOND, research THIRD** (most protected). Supersedes the 08-21 maintenance/builds
order (corrected visibly in RESEARCH §10). Standing from 08-21: the governor must act
BEFORE the hard cap; client work stays protected via D2. Four measured cap outages are the
case: 08-17 (self-set limit), 08-25/26 (credit, ~9h), 08-27 (2h), 08-28→31 (~2.5 days).

## Shape (five parts, smallest that closes the door)

1. **METER** — spend estimate from `llm_call_log`: tokens × a `model_prices` data table
   (per-model in/out/cache-write/cache-read rates), month-to-date + trailing-24h burn rate.
   `[MEASURED 08-31, 7d io-tokens]`: page-content-writer 52.5M dominates; council-gate
   12.4M io + 393M cache-read; tool-improver 7.8M. A VIEW, so every reader agrees.
2. **CLASS MAP** — a data table `work_class_map(item_type, class, llm_bearing)` with
   class ∈ {maintenance, build, research}. Populated from the measured item→handler map
   (NOTES 08-31), owner-adjustable. **LLM-free types are NEVER shed** (page_rerender,
   undeployed_asset, the rerender-pages family — shedding them saves nothing and stops
   serving); an unmapped item_type defaults to **maintenance** (sheds earliest = the safe
   default for an unknown spender). ⚠ the map is multi-valued today (content_rewrite has
   two handlers) — class keys on item_type alone, deliberately.
3. **STATE** — a `scheduled_tasks` row (~120s, the claimed-item-timeout pattern) computes
   month-to-date spend vs the configured budget and writes `governor_state(shed_level)`:
   L0 none · L1 shed maintenance · L2 + builds · L3 + research (= everything LLM-bearing;
   one step short of the hard cap the account would impose anyway). Thresholds
   default **70% → L1, 85% → L2, 95% → L3** of monthly budget (config, not code). Every
   level CHANGE writes ONE doc_notes row (the RFC_022 cron lesson: silence must be
   distinguishable from not-running — the task also stamps a heartbeat on no-change runs).
4. **ENFORCEMENT** — at the CLAIM step (`claim_work_item_action.go`, beside the existing
   `ai_endpoint_unavailable` refusal — the proven seam): if the item's class is shed at the
   current level AND the item is llm_bearing, refuse with reason `spend_governor_shed`
   (a new claim_result reason, visible in every existing meter). Items stay triaged,
   nothing burns attempts — exactly the deferral shape the estate already handles.
   **Ships opt-in, default OFF** (`governor_config.enabled=false`; the 08-02 ruling: new
   authority on a shared seam is a field with the unsafe default off). Registered in the
   concept register in the same commit (ordering-exemption condition 2).
5. **OPTION C GATE** — once the governor is live AND exercised once (a real or induced
   shed observed), interval ≤25s unlocks: a separate migration editing VERIFY 2/7's lever
   in lockstep (637's instruction), its own council round.

## What it deliberately does NOT govern (named, not hidden)

- **council-gate spend** (393M cache-read/7d): platform self-governance, scales with change
  volume not domains; shedding it silences review — an owner lever, not a governor class.
- Direct agent runs not driven through `site_work_items` (090 loops, councils, chat).
- The hard cap itself — the governor exists to make L1–L3 happen first.

## Owner input still wanted (one number)

The governor's **monthly budget figure** (its thresholds key on it; it should sit under the
console cap so L3 fires before the account refuses). Until given: config ships with budget
NULL = governor inert even when enabled — refusing to guess a spend ceiling.

## Build order (each stage independently shippable, council per coherent commit)

A. `model_prices` + `work_class_map` + `governor_config` + `governor_state` + the meter
   view + the state task (migration, DB-only, inert without the Go half).
B. Claim-step refusal (Go; opt-in read of governor_state; unit + mutation tests; register
   entry same commit). Inert until `enabled=true` AND budget set.
C. Flip-on with the owner's budget; observe one shed cycle; then the Option-C round.

## D4 stage B — DESIGN CORRECTION (2026-09-02, found at the code before building)

The §above's "enforcement at the CLAIM step" is WRONG as the primary seam, by this lane's own
413 lesson (ranker/drainer key disagreement). Under shed level ≥1, a claim-only check makes a
site whose oldest-loadable work is shed into a selection HOG: 657's selector ranks it by that
loadable work → picked → loop loads → every claim refuses `spend_governor_shed` → no claimed
row → busy-skip never engages → picked again next fire. Starvation of NON-shed work, governor
edition.

**Corrected stage-B shape — the shed filter must live where eligibility is DEFINED, with the
claim as backstop only:**
1. **Selector** (config query, 657-style migration): eligibility excludes rows whose
   (class, llm_bearing) is shed at the CURRENT `governor_state.shed_level` — one join to
   `governor_work_class_map` + one scalar subquery on state, guarded exactly like 657 (md5
   preflight; the 657 VERIFY's md5 arm must move in LOCKSTEP — coordinate before applying).
2. **Loader** (`load_work_item_actions.go`, Go): same predicate, same lockstep — the AST
   ordering-contract test gains the shed clause; unmapped types default maintenance+bearing
   HERE (the Go default the PLAN already states).
3. **Claim** (`claim_work_item_action.go`): the backstop only — a distinct
   `spend_governor_shed` claim_result reason for races where state moved between load and
   claim. Never the primary gate.
All three read `governor_config.enabled` — false short-circuits to today's behaviour
byte-for-byte (the opt-in). The fire-gate (bugs_open/415) needs NO shed arm: it may fire on
shed-only backlogs harmlessly (selector then picks nothing — one empty turn, not a hog).
Build order: Go halves first (inert: enabled=false), image + roll, THEN the selector
migration (image-before-config), each with the governor-disabled path proven byte-identical.

## D4b — GOVERNOR AGENT SCOPE (added 2026-09-03, owner ruling the same day)

**Why it exists.** D4 went live 10:14:32Z on 2026-09-03 and was proven the same morning
(induced shed, 11:17–11:29Z). Measured that afternoon by orchestration lineage: the governor
reaches **~28% of fleet LLM spend**, because it sheds *work items* and 69.4% of spend has no
dispatch-loop ancestor — `council-gate` alone is **62%**. CLAUDE.md §council had recorded
council as ~85% of fleet LLM spend before migration 377; the two facts had never met. So D4 as
built is a **dispatch** governor and cannot defend a budget. **Owner, verbatim: "extend it,
reducing council spend is a fairly easy save if it comes to the crunch."**

**Design — two stages, mirroring D4, architecture round per AGOV-013's standing gate.**

- **Stage A (migration 751, APPLIED 2026-09-03 17:12:21Z, inert):**
  - `governor_admits_class(class, llm_bearing)` — THE level comparison, factored out so the
    r1 architecture ruling ("never re-spell the predicate") survives a second namespace.
  - `governor_admits(item_type)` rewritten as a one-line caller. Same signature; selector md5
    untouched. **Equivalence PROVEN in the verify** (legacy copy kept, 4 levels × 3 classes × 2
    bearings + unmapped compared, discriminating control, copy dropped) and mutation-proven
    (two induced breaks, each caught at its own arm).
  - `governor_agent_class_map` — a SEPARATE table (different namespace; a shared PK would
    collide silently). **UNMAPPED = ADMITTED**, the opposite default to work items, on purpose:
    an unmapped agent type is every agent in the estate.
  - `governor_admits_agent(agent_type)`.
  - `governor_withheld_runs` + `_recent` view — a refused RUN creates no row anywhere, so the
    observable must be WRITTEN, or a shed council submission reads as the documented 29-minute
    dispatch latency.
  - Seed: `council-gate → research` (L3). **THE LEVEL IS AN OPEN OWNER QUESTION** — "crunch"
    reads as late, so L3 is the conservative default; L1 is the biggest earliest saving and
    would remove advisory review for roughly the second half of each month at current burn.
    One UPDATE moves it.
- **Stage B (Go, NOT YET WRITTEN — sketched in the council submission):** one opt-in gate in
  `platform/messaging/processor.go executeWorkflow`, after `resolvedAgentType`, before
  `RecordAgentRun`. Flag `honour_spend_governor_run` on the agent's config (its own name, so a
  grep for either flag finds exactly its consumers), unsafe default OFF. On refusal: write the
  withheld-runs row, log, `return nil` (consumed, not retried). Fail-open on read error. Tests:
  flag absent/false/string → no query; admitted → runs, no row; not admitted → no run, one row
  with the correlation id; DB error → runs. Config half: a `_HOLD` migration setting the flag on
  the `council-gate` row only, with a fleet negative control, applied after the binary rolls.

**Considered and rejected for stage A:** seat-capping (run only the two always-on seats at a
shed) — roster surgery across 17 seats while the 099 mirror is suspended over 377's cache
breakpoint. Refusing the run is one seam; degradation can be a later option.

**Council:** corr `dc6d2a54-bd73-4827-8267-49c5500467ac` (architecture round, submitted
2026-09-03 ~17:15Z; six edits incl. the stage-B sketch; verify text pasted verbatim).

**Open owner questions carried:** (1) the council-gate shed LEVEL; (2) whether the other
ungoverned agent types in the 69% (landmine-verifier $7.94/24h, the auditors) should be
mapped — each is its own opt-in row and, per the standing gate, its own review.
