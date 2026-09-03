# NOTES — dispatch throughput / whole-architecture scale review

Append-only, newest at the bottom. Technical log: what was tried, what the system said,
every misstep.

---

## 2026-08-18 evening — workstream claimed; scope widened at the owner's direct request

Claimed by the "throughput" session (uk@websy.uk). The owner asked for a deep research pass
on ALL options for increasing system throughput — explicitly including whether more than one
repo, deploy path or cluster is needed — for an estate of **several thousand domains**, and
(added on review, 08-18 late) **promotion-driven onboarding bursts of many domains per day**.
This brings forward the "whole-architecture scale review" the owner seeded in the
site_delivery lane ("after the working site" — the owner's direct ask supersedes that
sequencing; noted, not relitigated).

### Baselines re-run [MEASURED 2026-08-18 ~18:00–18:30 UTC], per the PLAN's own instruction

- sites: **43** rows (all; no status filter — vocabulary landmine)
- open work items: **3,051**; completions 7d: **6,595** (~39/hr average vs ~83/hr ceiling)
- llm_call_log 24h: **858 calls**, ~6.1M input / 2.3M output tokens (claude-sonnet-5
  dominant); 7d: 5,784 calls, 69 failures, only 5 rate-limit-shaped; latency p50 **18.6s**,
  p90 **62s** — LLM latency dominates handler runtime (item p50 36s)
- orchestration_states 24h: **5,616**
- DB: **5.9 GB** (orchestration_states 2.7 GB, llm_call_log 1.4 GB)
- cluster: 5 workers × 7.5 CPU / ~58 GiB, **6–20% CPU, 9–40% mem** — compute is NOT the
  binding constraint; the fleet serialises while the nodes idle
- Kafka: **3,947 topics** = 3,022 `job.*` per-spawn + 925 `system.*`;
  `system.agent.copywriter.tasks.high` = 3 partitions RF2; scheduler pod healthy, 0 restarts

### bugs_open/240 topic-count reconciliation

240 measured **25,042** topics on 08-10; live read 08-18 ~18:15 UTC = **3,947**, with a
`kafka-topic-cleanup` job in ns `kafka` having completed minutes before the read. 240's own
tail records the one-off sweep (→354) plus a C4 cron firing on 08-11. So the number is
reconciled as: sweep + some reaper holding it to ~4k under load. **[UNVERIFIED] which reaper
is actually doing the work now** (in-cluster idle-gated `agent-job-cleanup`, the
`kafka-topic-cleanup` job observed today, or the owner-laptop crontab) — establish before
citing topic lifecycle as "managed" at N× dispatch. 240 remains OPEN (`MetadataTopics` still
blank in `platform/kafka/dialer.go`).

### Diagnosis filed (090), per the 2026-07-31 owner ruling

Symptom = the scheduler single-flight mechanism (countInFlight counts rows not executions;
loadDueTasks per-row guard). Filed 2026-08-19 ~00:0x UTC:
- intake `CORRELATION_ID=8237de92-d873-4033-afea-93ba919d2435`
- run   `RUN_CORRELATION_ID=a16b82cd-b89a-45d5-b5df-4370c754e2fd` ← artifacts key
Loop claimed it within the 180s wait. Verdict to be recorded here when it lands.

### Wedged-loop diagnosis check (PLAN §3.3)

`needs_diagnosis` 9d2e3963 (filed 08-18 12:14, "build-dispatch-loop freezes in
EXECUTING_STEP at process_item_iter_N_spawn_handler") still **failed** at 08-18 18:44 UTC
read. To re-file via 090 after the scheduler diagnosis returns (avoid two concurrent runs
on the same subsystem).

### Exploration corrections worth keeping (full evidence in the session's three reports)

- The 18 dirty uk_001 overlays are release tag bumps (v1.0.1310), NOT multi-domain work.
- RFC_034 "per-instance scope" = component-on-page instances, not infrastructure.
- Multi-cluster: MCL-001 code built-and-unused; va001 never existed (register: aspirational).
- The `job.*`-topic and adapter-serialisation facts above are the 08-18 state; the adapter
  "treadmill" 0/5 failure and the chassis 8-concurrent ceiling are from
  chassis_replica_scaling's measured record (07-28), not re-measured today.

## 2026-08-19 — PLAN §3 items 2 and 3 dissolved on contact with the record

- §3.2 (reconcile 029's mechanism): **already done, 2026-07-21, in 029's own file** —
  "CORRECTED DIAGNOSIS 2026-07-21 … the title mechanism is wrong; this is bug 003's blast
  radius, not a concurrency-group bug", with the countInFlight row-count analysis this PLAN
  wanted written. The PLAN's author (08-18) missed it; I nearly re-derived it. The cheap
  check that caught it: grep the bug file for "CORRECTED" BEFORE planning a correction.
- §3.3 (wedged-loop diagnosis): **actively owned by the 029 lane as of this morning** —
  six 029 commits on 08-19; their `cfdb247c4` shows the morning's 090 refuted the wedge on
  a false premise (awaited_requests retains 7d, 20 instances live) and narrows the death to
  `continueExecution` for `iter_N+1_call_handler` on the response-consumer goroutine.
  NOT re-filing. Also noted: 029 "part A" (inverted retry window) FIXED, behaviourally
  proven 08-18 18:28Z, live on v1.0.1309; fleet now at **v1.0.1314** (my 08-18 baselines
  were taken on the 1310 fleet — dated accordingly).
- Both corrections written into PLAN §3 visibly (strikethrough + CORRECTED blocks).

## 2026-08-19 — build-cost measurement (the burst-sizing number), n=1 + one dead end

**loanzy.uk** (site row born 2026-08-18 12:53, `deployed`): ~213 work items filed day one;
**410 orchestrations attributable, 87 FAILED (21%)**; LLM via two attribution routes
(work_item_id join ∪ orchestration-payload ILIKE on the domain): ~180–200 calls,
**~1.1M input + ~1.1M output tokens ≈ $20 at Sonnet list**; 21 image assets (imagery API
cost [UNMEASURED]); wall-clock submit→mostly-built **~10.5h** (much of it fleet-queue wait,
not work). All `[MEASURED 2026-08-19]`, queries in RUNBOOK.

**Misstep worth keeping:** attribution by payload text is silently partial. The second case
(remortgagecalculator.uk, 10 pages 08-17/18) returned a well-formed ZERO on both the domain
string AND the site UUID — its orchestrations embed neither. So "what did this site's build
cost" is unanswerable today without archaeology for some sites; the n=1 figure is marked
accordingly, and per-site cost attribution is itself a gap the RESEARCH doc records
(§4 caveats). The trap class is measurement-discipline's "interrogating what doesn't exist
returns a well-formed answer".

**LLM spend split, 7d window `[MEASURED 08-19]`:** council-gate dominates — 2,207 calls,
7.9M in + 6.4M out + **271.9M cache-read** ≈ $200/wk; it scales with CHANGE volume, not
domains. Steady-state maintenance completions are dominated by page_rerender (4,808/14d),
which is LLM-free. Per-class LLM-bearing maintenance cost remains [UNMEASURED] — the one
figure still owed before pricing tiers (RESEARCH §6).

## 2026-08-19 — deliverable written; 090 verdict pending

`RESEARCH_2026-08-18_throughput_to_thousands_of_domains.md` written (9 sections, decision
table D0a–D16). 090 run a16b82cd at `verdict` step (EXECUTING_STEP) as of ~15:45 UTC;
verdict to be appended here when it lands.

## 2026-08-19 — 090 outcome: run FAILED at verdict (max_tokens), NO verdict exists

The a16b82cd run completed handler+diagnoser then died at `verdict`:
`stop_reason=max_tokens (output_tokens=32000 …, 0 chars recovered)` [MEASURED 15:32:06Z].
So the RESEARCH doc's "independently verified by the 090" line was FALSE as written —
corrected visibly in §2.1, substituting first-hand verification per the 2026-07-31 ruling
(measured cadence + two independent code reads + matching ceiling arithmetic). NOT
re-filing: a re-run would face the same cap, and the failure class is owned by
bugs_open/183 (step token pressure) — observation contributed into that file. The claim's
verification status is: strong first-hand, loop-unverified, stated as such wherever cited.

## 2026-08-21 — OWNER RULINGS on the decision table (chat; discussion continuing)

- **D0b**: 50 signups/day is a welcome MAXIMUM; expect a fraction. NEW REQUIREMENT: a
  human review gate — owner checks each site before release; if not OK, a CLI-assisted
  fix loop (framework+site) BEFORE the site goes out. **That mechanism is being designed
  in ANOTHER THREAD** — do not build it here; the burst path must leave the seam
  (build → owner review → fix loop → release). Adds some load (fix iterations).
- **D0a**: THREE PORTFOLIOS decided in shape (numbers still owed): (1) client/third-party
  retained sites — high attention, client pays so AI spend secondary; most third-party
  sites will be HANDED OFF, only some retained; (2) own high-attention portfolio;
  (3) own low-attention group. Own-domain maintenance spend to be kept LOW.
- **D2 RULED: clients served first.** Owner also floats a fully separate cluster for
  client domains as the clean separation (ties into D13).
- **D3 RULED: lockstep** (timeout moves with concurrency).
- **D4 direction** (reading to confirm): as spend approaches the cap, shed own-domain
  build/improvement work FIRST, keep maintenance running, protect client work longest.
  (Refinement noted: the governor must act BEFORE the hard cap, since at the cap the API
  refuses everything indiscriminately.)
- **D5 RULED: no second provider now** — maintenance too buggy to add model-choice
  complexity; optimise later.
- **D6 RULED: Batch API yes** (classes to be picked).
- **D7 CORRECTED by the docs** (platform.claude.com/docs/en/api/rate-limits, fetched
  2026-08-21): tiers DO exist — Start $500/mo cap, Build $1,000/mo, Scale $200,000/mo,
  Custom uncapped; orgs move up automatically with usage history, and there is a
  "Request rate limit increase" flow on Console → Settings → Rate limits for the monthly
  spend cap too. ⚠ The 08-17 outage error text ("You have reached your SPECIFIED API
  usage limits") is the documented signature of the owner's OWN self-set Billing-page
  limit (HTTP 400), distinct from the tier cap (HTTP 429 `enforced_spend_limit_reached`)
  — so at least one recent outage was self-inflicted config, fixable in Console today.
  Cache reads do NOT count toward ITPM — caching raises effective throughput directly.
- **D8 RULED: keep GitHub Actions for now**, scale runners; revisit later. (Interim
  batching still worth building — it cuts runs ~5× regardless.)
- **D9: discussion requested** — what breaks if the polling dispatcher is removed for
  worker-pull; answered in chat + README (summary: governor first, policy moves into the
  claim query, staged flag-gated cutover per the chassis worker-pool precedent; fits
  satellites BETTER, provided each satellite pulls from its OWN DB).
- **D10: options requested** — answered in chat + README (CI on push via self-hosted
  runner pattern + scheduled release train; the real gap is that the working branch is
  never pushed — origin is ~7.7k commits behind).
- **D11: explanation requested** — answered in chat + README (adopt for platform-code
  sessions, docs stay on shared tree, exactly one deployed branch).
- **D12 RULED: start plan B** (own authoritative DNS + CF-for-SaaS). Execution belongs
  to the domain programme lane — pointer to be left there.
- **D13 direction**: first split = a CLIENT satellite (cluster + CF account) for clean
  client/own separation; per-mega-client satellites an option, not default; owner wants
  to MANAGE client domains sooner rather than later ⇒ five seams become near-term.
- **D14 RULED: spot OK while mainly own domains** (revisit with the client satellite).
- **D15 RULED: maintenance pauses during bursts.** Burst-scaling question answered in
  chat (burst profile: pause maintenance + raise N/workers + governor budget shift;
  node autoscale is the missing piece; LLM tier is the true burst ceiling).
- **D16 RULED: review retention/archival** — "a small database is an easily managed
  database"; proposal owed.

## 2026-08-24 — PHASE 2 EXECUTED: N=2 live via migrations 582→583→584

Fleet context first: fresh chassis roll this morning — **v1.0.1332**, both pods started
09:39Z (was 1314 at my 08-19 reads). Nothing of this lane's rides that image (config-only
lane); the provenance startup line had already scrolled, tag+time recorded instead.

**Pre-change baseline [MEASURED 10:3x UTC]:** completions 629/24h; dispatchable backlog
141 across 6 sites (modest demand — the N=2 gain will show in WAIT TIMES more than raw
completions at this depth; demand control noted for the 24h comparison).

**The PLAN's migration B was incomplete and the live config said so:** THREE notify
stamps carry the hardcode (trigger `notify_scheduler` + `notify_scheduler_idle` + loop
`notify_scheduler` — occurrence count **3** as of 2026-08-24; one md5
`64db3df8551b60a2098443ce00569604`), not the PLAN's two. The idle path would have
released the original row on every idle sibling tick. Found by counting occurrences
before writing, not by reading the PLAN's list. Also verified before writing: the
scheduler DOES deliver input_data for fire_message tasks (`main.go:192`, `fireTrigger`;
the scary "never sent anywhere" comment at :195 applies to CTE-only tasks), and
`QueryDatabaseAction` params error on nil — hence strict A→B ordering.

**Applied 2026-08-24 ~10:38Z, in order, each ON_ERROR_STOP clean:**
- 582 (inert task_name seed): UPDATE 1.
- 583 (parameterise all three + call_dispatch mapping): snapshot_agent both agents,
  4× UPDATE 1, post-check DO/RAISE clean. **Ordering guard INDUCED before applying** —
  ran 583's pre-flight against the pre-582 row and it RAISEd as designed.
- 584 (sibling insert): INSERT 1, shape post-check clean.

**Artefact verification [MEASURED 10:40Z]:**
- `build-pipeline-trigger-2` last_triggered_at = last_completed_at = 10:40:08 — **its
  stamps land on ITS OWN row**; original 10:40:39 on its own. The stamp trap is dead.
- **Two `call_dispatch` orchestrations in AWAITING_RESPONSES simultaneously** — two
  concurrent dispatch turns, directly observed.
- Idle path exercised live through the parameterised query (multiple `complete_idle`
  COMPLETED, `error IS NULL` on every recent trigger/loop run) — the nil-param hazard
  did not fire.
- Per-minute distinct-sites meter still reads 1 at this backlog (141 items / 6 sites,
  many idle ticks) — expect 2 in busy minutes; 24h throughput comparison owed, with
  backlog depth held beside it.

**Council:** migrations are gate-scope since 08-19 — submitted, `Council-Submitted:
db9b7cbf-7b94-471a-a4cf-26a6679fa47f` (schema note: `plan` is an OBJECT
{summary, edits[{file,symbol,operation,rationale,sketch}], grounded_in, risks} — my
first attempt used a bare edits array and was refused client-side).

**Register:** WDS-002 gains the concurrency-lever paragraph (same commit).

**OWED / next:** (a) the same-site double-pick induction (two manual dispatches at one
site → confirm wasted-spawn-not-double-handle at claimed_by/attempt_count); (b) 24h
throughput + wait-time comparison with demand control; (c) do NOT add sibling #3 —
gated on adapter decision + D4 governor; (d) Phase 3 (batch 8 + timeout 600, D3
lockstep ruled) after the 24h read.

## 2026-08-25 — council round 1 REVISE, answered with measurements; round 2 submitted; HANDOFF cut

Round 1 (corr db9b7cbf): REVISE, gating = guardian HIGH — feared a FOURTH stamp nested in
sub_workflow/substeps, invisible to a top-level step walk. Answered by census, not
argument: fleet-wide WHOLE-TEXT scan of all active agent_definitions [MEASURED 08-25]
returns **0** hardcoded stamps (positive control: the same scan sees substep-level text —
6 configs match `spawn_handler`), **0** in scheduled_tasks.pre_query, and the `$1` form
lives in exactly the 2 edited agents. Guardian MEDIUM (apply atomicity): the 582→583→584
sequence is self-gating (DO/RAISE cross-file pre-flights, 583's induced pre-apply).
Guardian LOW (INSERT..SELECT stale identity): the INSERT's explicit 12-column list
excludes id and both stamp columns. reuse LOW (native knob): max_concurrent is dead
config (STARTER §2); executions-count fix = the D9 fork, owner-deferred. architecture
LOW: register note added to WDS-002 — occurrence-counting is not verified enumeration;
the whole-text census is the repeatable check. Round 2 submitted on the SAME correlation
(RESUBMIT_CORR). Verdict unread at handoff time — first task for the next session.
`HANDOFF_2026-08-25_continue_here.md` cut; workstreams memory repointed at it.

## 2026-08-25 (session 2, 16:10Z→) — council r2 REVISE read; the 24h N=2 read; and the lane's mechanism claim REFUTED at the artefact

### 1. Council round 2 = REVISE again (verdict 12:23:42Z, `diagnosis_artifacts` kind=`council_report`, corr db9b7cbf; 8 abstained)

Gating = guardian HIGH on edit 3: the same-site double-pick induction "still OWED" sits inside a
subsystem with a documented DISAGREE landmine; wants the induction run or a time-boxed monitoring
commitment. Others: editquality MEDIUM ("two call_dispatch turns in flight" proves the trigger
fired twice, not that two DIFFERENT sites were dispatched); debug_historian MEDIUM ×2 (no
re-runnable post-apply verify; task_name resolution seen in ONE snapshot, not across cycles);
guardian MEDIUM (task_name key + input_mapping = key-deletion fragility for later migrations);
architecture MEDIUM (WDS-002 must say N=3 needs D9, not another clone triple); reuse LOW (is there
a `-2` sibling idiom elsewhere?); guidelines (WDS-002 should NAME `input_data.task_name`, not just
cite 582-584). All answered by measurement below, and one of them turned out to be the thread that
unravelled the lane's central claim.

### 2. The 24h read, WITH the demand control `[MEASURED 2026-08-25 16:1xZ]`

| window (24h) | arrivals | completions | claims | wait p50/p90 (min) | run p50/p90 (s) |
|---|---|---|---|---|---|
| pre  08-23 10:38 → 08-24 10:38 | 843 | 627 | 632 | 31 / 148 | 41 / 311 |
| post 08-24 10:38 → 08-25 10:38 | 2,577 | 2,332 | 2,340 | 49 / 171 | 30 / 51 |

Completions ×3.7 is **demand, not the lever**: arrivals ×3.06, and the growth is `page_rerender`
393→2,010 (LLM-free; a webdesign.co.uk rerender flood). What is attributable: completions/arrivals
74%→90% (the fleet kept up with 3× the demand); handler runtime p90 311→51s (the cheap item mix).
Wait p50 rose 31→49 min — demand-driven, not a clean reading either way. Backlog at the read:
triaged **158** across **5** sites (was 141/6 at the apply). Verdict on the raw number: unusable
alone; the honest figure is in §4.

### 3. The per-minute distinct-sites meter is NOT a concurrency meter (first crack)

Control run on the PRE window (nominally N=1): **27.7%** of claim-minutes had ≥2 distinct sites,
max **6** in one minute. Post: 32.4%, max 4. A true single-flight row at 60s cannot produce 6
sites in a minute. STARTER §7's meter (and RUNBOOK's copy) was measuring something other than
what it said. The orchestration-level meters below are the real ones. ⚠ `orchestration_states`
is retained ~24–27h (`min(created_at)` = 2026-08-24 13:09 at a 16:1xZ read on 08-25; trigger/loop
rows from 15:40 on) — the pre-window loop rows are GONE, so every orchestration census below is
N=2-period only. That is the period that matters, but no orchestration-level pre/post exists.

### 4. Orchestration-level census, N=2 period (08-24 15:40 → 08-25 16:13Z) — answers to the seats

- **task_name resolves on every cycle (debug_historian):** trigger runs **777 / 777** per row,
  loops **548 / 546**, each row's loops across **28** sites; **0** `query param path resolved to
  nil` anywhere; the only 4 FAILED rows are reaper kills (`spawn_dispatch` stale >4h ×2,
  `notify_scheduler_idle` stale 1h, loop `call_handler` idle >30min) — the spawn-handshake /
  wedged-loop class, all on trigger-2 (n=4, identical config; one at 09:26 = the 09:27 chassis
  roll to v1.0.1337). `__step_error` keys: 63 loop-level + 37 `__step_errors` = handler failures
  routed through `error_step: mark_failed`, normal.
- **Two DIFFERENT sites at once (editquality):** minutes with trigger-spawned loops alive on ≥2
  distinct sites: **631 of 1,179**; alive loops per minute 1–**8** (mode 1: 482 min, 2: 378,
  3: 198, 4: 245, 5: 89, 6: 59, 7: 16, 8: 6). Little's law on the window: 1,089 loops × mean
  135s / window = **1.65** mean alive.
- **DOUBLE-HANDLE CENSUS (guardian HIGH — the induction, answered by the population):**
  handler orchestrations that are children of dispatch loops: **2,579 over 2,502 items**; items
  with >1 handler: **41**, every one `handlers = attempt_count = 3`, sequential (`overlapped=f`),
  all ended `failed` — legitimate retries; **overlapping handler pairs for the SAME item: 0**.
  The atomic `claim_work_item` (`WHERE id=$1 AND status IN ('triaged','approved') … RETURNING`)
  held on every one of the 4,265 claim attempts. Same-site concurrent handler PAIRS: **775**
  across **27** sites — so the race the PLAN said to induce once has occurred ~hundreds of times
  naturally, and the "wasted spawn" is in fact a lost CLAIM (claim→check_claim→done in ms, no
  spawn): **1,670 of 4,265 claim attempts lost (39%)**; loops that loaded items but claimed
  nothing: 6 + 5.
- **Causal test — does co-handling one site raise failures? NO.** Handler fail rate WITH a
  concurrent same-site partner **1.55%** (22/1,421) vs WITHOUT **3.85%** (45/1,169). Failure
  classes: `OWNED_PAGE_GUARD` (page-rerender, 33), `store_generated_component` (19),
  `save_page_sections` (9) — none concurrency-shaped; git-shaped errors on items 20→33 over a
  3× demand rise (2.4%→1.3% of arrivals). Retried share of completions 0.2%→0.0%.
- **Sibling idiom (reuse):** `scheduled_tasks` = 100 rows / 50 enabled; the ONLY `-N` name is
  `build-pipeline-trigger-2`; 39 date-suffixed `oneshot-*`/`offer-analyser-oneshot-*` rows are
  disabled one-offs, not multiplicity. Group `dispatch` = 4 rows (2 mine + `improvement-sweep`,
  `intent-collection`, both disabled). So this IS a novel idiom, not reuse of one.

### 5. THE REFUTATION — a `scheduled_tasks` row is NOT a single-flight slot; the scheduler is fire-and-forget

Three artefact-level facts, any one of which is decisive:

1. **A row overlaps ITSELF**: trigger runs of the SAME row with overlapping lifetimes —
   **361 pairs** (`build-pipeline-trigger`), **322** (`-2`), min gap **0.25 s**. Fire cadence per
   row p50 **90 s**, p90 91 s (interval 60 + the 30 s tick), while a run lasts p50 **97 s**, p90
   242 s. A single-flight row cannot do that.
2. **Both stamps are set at FIRE**: `scheduled_tasks` read 39 s after a fire shows
   `last_triggered_at = last_completed_at` to the microsecond on both rows
   (`in_flight_per_guard = f`), with runs of p50 97 s still executing.
3. **The code says so, in the fire path**: `cmd/scheduler/main.go` `runTick` → after
   `fireTrigger(...)` → `stampCompleted(...)` — "For fire-and-forget tasks, mark completed
   immediately so the concurrency slot opens for the next tick. The message has been published —
   we don't wait for the orchestration to finish." `stampCompleted` = `UPDATE scheduled_tasks SET
   last_triggered_at = NOW(), last_completed_at = NOW()`. Introduced **`892a289e9` 2026-03-17
   "for fire and forget tasks dont wait for a response, send complete immediately"**; refactored
   into the named function by `dc2e4b61a` 2026-07-21 (bugs_open/048). Live scheduler:
   `kafka-scheduler:v1.0.1337` (provenance line scrolled; facts 1–2 are the artefact proof).

Consequences, all `[MEASURED 2026-08-25]` unless marked:
- `loadDueTasks`' per-row guard (`last_completed_at >= last_triggered_at`), `countInFlight`,
  `max_concurrent`, and `timeout_seconds` are **all inert for every `fire_message=true` row** —
  satisfied/zero the instant a task fires. `loadDueTasks`' own doc comment ("prevents the same
  task from spawning multiple concurrent pods") is false and has been since March.
- The agents' three `notify_scheduler*` stamps (`SET last_completed_at = NOW()`) are **inert**:
  they bump a column the guard already passes. So **583's rationale was false on this binary**
  (a sibling under the hardcode would NOT have fired at 300 s — the scheduler stamps the sibling
  itself at fire — and the hardcoded stamp on the original row changed nothing). 582/583 are
  harmless hygiene (correct IF the scheduler ever reverts to awaiting completion), not a fix. The
  08-24 "its stamps land on ITS OWN row" verification read the SCHEDULER's stamp and attributed
  it to the notify step — a writer inferred from a reader.
- **What 584 actually did: doubled the FIRE rate — two fires per ~90 s, ~1 s apart** (16:22:20.96
  / 16:22:22.03). Because `find_dispatchable_site` cannot see a claim that has not happened yet
  (time from loop spawn to first claim p50 **17.7 s**, p90 **24.2 s**, p99 131 s), the second
  fire picks the SAME site **94%** of the time (475/505 and 479/508 loaded loops co-picked within
  15 s) and loads an overlapping batch (typically 4 of 5 shared — e.g. loops `769c3b46`/`c720c890`
  16:10Z: `[8548e940,a391bc2e,bca87604,2188fcf4,1be4a901]` vs `[a391bc2e,bca87604,2188fcf4,
  1be4a901,c6b310a3]`). The pair then interleaves claims: avg **2.6** items claimed per loaded
  loop (was up to 5 for a lone loop). Net: busiest hour **227** claims (118 + 109) against a
  single-row ceiling of ~40 fires/h × 5 = ~200 → **≈ +10–15% claim throughput, not 2×**; the real
  gain is **two handlers concurrently on the deep site** (per-site latency), measured safe (§4).
- **The PLAN's "per-site serialisation survives" is false** (775 same-site concurrent handler
  pairs) — and measured harmless at 2. The safety that DOES hold is the atomic claim.
- **The native rate knob exists and is live: `interval_seconds`** (the reuse seat's question).
  With a 30 s tick: interval 60 → 90 s cadence (measured); 30 → 60 s; ≤25 → every tick (30 s).
  Fires spaced ≥30 s apart see the previous claim (p90 24.2 s) and steer to DISTINCT sites — the
  behaviour N=2 was designed to give and does not.
- Phase 3's D3 "timeout moves with concurrency" — `timeout_seconds` is moot on this binary; only
  the batch half (`load_items.max_items` + `process_item.max_iterations`) is real.
- STARTER §2's OWN data contradicted its conclusion: 17 runs in 25 min at avg 218 s duration ⇒
  mean concurrency 17×218/1500 = **2.5** (Little's law). It was read as "the real tick is 218 s".
  Both code reads (08-18) opened `loadDueTasks`/`countInFlight` (the READERS of the stamps) and
  neither opened the fire path 80 lines above (the WRITER). The 090 that might have caught it
  died at `max_tokens` (§2026-08-19). Logged in `WRONG_CALLS.md` 2026-08-25.

### 6. Also found: the sibling-parity trap is already loaded

`docs/agent_docs/sql_for_agents/213_dispatch_gate_matches_dispatcher.sql` — **untracked on disk
since 2026-08-12 18:39, never committed, never applied** (live `pre_query` md5 `200246f7…` is the
pre-213 form + the 503 `retry_after` clause; CONTRIB 2026-08-03 in `bugfix_029_dispatch_gate/`
says one clause was applied by hand; 029-dispatch closed 08-20). It rewrites `pre_query`
`WHERE name = 'build-pipeline-trigger'` — by name — so applying it now would desync the sibling
(584 byte-copied `pre_query`). LANDMINE appended; the re-runnable parity check is in the new
`584_…_VERIFY.sql`.

### 7. What this session did NOT do, deliberately

Did not disable the sibling, did not touch `interval_seconds`, did not change the selector. The
owner authorised N=2 by name on a premise that is now refuted; which lever replaces it is his call
(options costed in README_where_we_are, this date). N=2 stays live: measured safe (0 double-handles
in 2,579 handlers; failure rate lower with a partner), instant rollback, ~+10–15% claims.

## 2026-08-25 (session 2, ~19:50Z) — council round 3: APPROVED; two advisories acted on immediately

Report 19:5xZ (3rd on corr db9b7cbf): **APPROVED — "approved with 4 advisory objection(s), none
high-severity"**, 7 abstained. The honesty held up: editquality's notes call the plan "honest and
appropriately scoped … targets the actual causal path established in this round (fire-and-forget
stampCompleted, not per-row single-flight)". Both commits (`dc76d1c30` change, `e80561f04` docs)
carry `Council-Submitted:` and are auto-credited by 098; this entry's commit carries
`Council-Reviewed:` (verdict read in full before writing the trailer).

The four advisories, and what was done with each:
1. **editquality MEDIUM (a real defect in my VERIFY, caught on its first review):** liveness keyed
   on `orchestration_states.owner_agent_type`, a column a LANDMINE says reads ZERO for some active
   agents — a daily check that can false-RAISE. **FIXED:** liveness now keys on
   `input_data.task_name` alone (carried only by the trigger's runs and its loops, so presence
   proves delivery); re-run clean.
2. **guardian/architecture MEDIUM+LOW (no guard against sibling #3; "provisional with no deadline
   becomes permanent by default"):** **FIXED half:** assertion 6/6 added to the VERIFY — >2 trigger
   rows RAISEs, naming interval_seconds/D9 as the sanctioned paths; mutation-proved (inverted
   predicate RAISEs at m=2, exit 3). **Deadline set:** if the owner has not ruled options A–D by
   **2026-09-01**, re-present them — recorded in HANDOFF-b OWED 1 and WDS-002.
3. **reuse/constitution MEDIUM/LOW (sibling duplicates the native interval_seconds knob):** already
   the substance of options B/C; owner decision pending, deadline above.
4. **debug_historian LOW (582 re-applies jsonb_set unguarded on rerun):** content-idempotent,
   row-touch on replay only; noted, not worth a migration — the pattern to copy next time is
   `WHERE input_data->>'task_name' IS DISTINCT FROM <value>`.

VERIFY post-hardening run: **all 6 hold** (parity ×2 rows · identity · 0 hardcoded stamps with
positive control · liveness · 0 double-handles/24h · no third sibling).

## 2026-08-26 — OWNER RULED B; migration 637 APPLIED; the sibling retired for `interval_seconds`

**Ruling verbatim (chat, 2026-08-26):** "Decision - B as you suggest and we can do C when we have
the governor." B = retire the sibling, original row `interval_seconds` 60→30; C (interval 25, ~3×)
waits for the D4 governor.

- **Council:** new coherent change → new submission, `SUBMISSION_CORR
  69a04e0a-8e45-4f5a-b0bb-285f00c544ee` (DRY_RUN admission clean first). Applied before the
  verdict per the 07-29 ruling (review is after the fact by design); commit carries
  `Council-Submitted:`.
- **Applied 08:51:0xZ, ON_ERROR_STOP clean:** pre-flight gate-parity DO passed; `UPDATE 1` ×2
  (both guarded on pre-state — rerun-safe per r3's debug_historian advisory); post-check DO
  passed. State: `build-pipeline-trigger` enabled @ **interval 30**; `build-pipeline-trigger-2`
  **DISABLED**, row kept as the rollback path (`637_..._ROLLBACK.sql`, guarded; its lockstep
  instruction: re-edit the VERIFY lever assertion in the same commit).
- **VERIFY reworked to 7 assertions in lockstep:** 1/7 gate parity narrowed to
  pre_query/agent/topic/fire_message but across ALL rows INCLUDING the disabled sibling (it is
  the rollback path — a by-name UPDATE desyncing it would make a future re-enable fire a stale
  gate); NEW 2/7 LEVER: exactly one enabled trigger row at interval ≥ 30 — **"C only after D4"
  is now mechanical, not documentation**. All 7 hold post-apply `[MEASURED 08:53Z]`; the lever
  assertion mutation-proved (inverted predicate RAISEs at 30, exit 3). ⚠ full VERIFY runtime was
  **2m49s** this morning (the informational co-pick table dominates) — do not run it under a
  2-minute timeout; the DO block alone is seconds.
- **Overnight read that strengthens B:** the phase-locked pair's lost-claim share over the
  trailing 24h had risen to **59.6% / 57.9%** (was 38.7/39.7 at yesterday's 24.5h read)
  `[MEASURED 08:53Z]` — the co-pick cost was growing with the backlog shape, not shrinking.
- Housekeeping observation: `improvement-sweep` (dispatch group, interval 900) reads `enabled=t`
  this morning — another lane flipped it since NOTES 08-25 recorded it disabled. Unrelated to
  this change (the group cap never binds — 398); noted so the next group census doesn't trip.
- **Post-apply artefact reads (~09:2xZ):** [appended below when taken]

### Post-apply artefact reads `[MEASURED 2026-08-26 09:04 + 09:20Z]`

- **Cadence:** 27 fires 08:52→09:20, gap p50 **60 s**, p90 65 s, max 180 s (one slow tick). Only
  `build-pipeline-trigger` fires; the sibling's `last_triggered_at` is frozen at 08:51:09.23.
- **Steering (the point of B):** steady-state window ≥ 08:55: **24 loops across 15 distinct
  sites** — under the phase-locked pair, 94% of loops co-picked one site.
- **Lost claims: 12.9%** (9 of 70; was 38.7/39.7% at the 08-25 read and **58–60%** in the
  trailing 24h to this morning). Anatomy of the 9: **3 = `ai_endpoint_unavailable` on the SAME
  item** (`efccd5d8`, site `8ff093d5`) retried by successive loops — an endpoint-health refusal
  at the claim step, NOT a race (observed, not diagnosed; not this lane's); **6 =
  `already_claimed_or_ineligible` mid-batch** (claim_result_2/4) across 2 same-site overlap
  pairs — the residual consecutive-fire overlap when a deep site's loop releases between items.
- **Batches:** avg loaded 3.83, avg won 2.54 per loaded loop (12-min transition read: 12 loops /
  5 sites — small-sample, superseded by the above).
- **Demand jumped overnight:** 1,398 triaged across **31** sites at 09:20Z (was 158/5 at the
  08-25 read) — a filing wave; tomorrow's 24h post-B read is therefore a REAL throughput test:
  ceiling at this cadence ≈ 60 turns/h × up to 5 = ~300 claims/h.
- 637 council verdict not landed at 09:20Z (~30 min budget from 08:50 submission) — next session
  reads it (`69a04e0a`, doc_notes / diagnosis_artifacts).

### 2026-08-26 ~10:2xZ — the 637 council round died on an ACCOUNT-CREDIT outage, not on content; resubmitted

The 08:50 round completed `complete_invalid` at `council_decide`: "no reviewer produced a readable
opinion (9 abstained, 8 unreadable)" — every seat's LLM call got HTTP 400 **"Your credit balance is
too low to access the Anthropic API"**. Measured from `llm_call_log`: credit-balance failures ran
**2026-08-25 23:47:10Z → ~08:55Z** (~9 h; 100% of calls failed 00:00–08:00, ~600 calls, every
LLM-bearing agent type), recovery ~09:00Z (09:00 hour: 296 calls, 0 credit fails; last success
10:19). The round landed at 08:50–08:52 — minutes before recovery. **This is a DIFFERENT signature
from the 08-17 outage** (that was the self-set Billing-page limit, "specified API usage limits";
this is the prepaid credit balance — both HTTP 400, distinct messages — D7 notes updated by this
entry, not re-litigated). Consequences noted:
- The morning read's `ai_endpoint_unavailable` claim refusals (item `efccd5d8`) were this outage's
  tail, not a dispatch defect.
- **The 24h post-B read carries a caveat:** LLM-bearing handlers were dead 23:47→08:55 (mostly the
  pre-B side of the window) — completions in that stretch are rerender-only; hold this beside any
  pre/post comparison.
- D4's case just wrote itself: at an empty balance the API refuses EVERYTHING indiscriminately for
  hours — the governor exists to shed deliberately before that point. (Also the only reason the
  fleet half-worked overnight: page_rerender is LLM-free.)
- Resubmitted on the SAME correlation (`RESUBMIT_CORR=69a04e0a…`) so the `Council-Submitted:`
  trailer on `a5fd1651e` resolves; an infra-killed round leaves no verdict to answer.

### 2026-08-26 ~10:4xZ — 637 council: APPROVED on resubmission

Report on corr 69a04e0a: **APPROVED — "2 advisory objection(s), none high-severity"**, 9 abstained.
Both advisories are one point: `637_..._ROLLBACK.sql` is named in the rationale as part of the
safety story but absent from the `edits` array. Disposition: **the file exists and was committed
with 637 in `a5fd1651e`** (guarded UPDATEs, its own DO/RAISE post-check, and the lockstep
instruction to re-edit VERIFY 2/7); the omission was in the submission's edit list only — I left
it out believing council-scope refuses `_ROLLBACK` files as edits, though the `_VERIFY` edit in
r3 was accepted, so listing it would likely have worked. Lesson for the next submission: list
every artefact the rationale leans on, and let the scope filter do the refusing. Nothing to
build; verdict read in full; this commit carries `Council-Reviewed: 69a04e0a…`.

### 2026-08-26 ~11:1xZ — ~2h post-B interim read; first zombie-tail pair found and the VERIFY 6/7 narrowed for it

All meters `[MEASURED 2026-08-26 10:50–10:55Z]`, windows as stated; taken now partly because
`orchestration_states` retains ~24–27 h — tomorrow's 24h read loses this morning if it runs late.

- **Cadence:** 100 fires 08:52→10:49, gap p50 **60 s**, p90 100 s, max 276 s — p90 heavier than
  the 30-min read (65 s); slow ticks under load, watch at 24 h. Sibling frozen at 08:51:09
  (disabled, interval column still 60 — rollback row untouched).
- **Steering:** 100 loops across **22 distinct sites** (≥ 08:55 steady state).
- **Lost claims: 11.4%** (49 of 431). Anatomy: 46 `already_claimed_or_ineligible` across 43
  distinct items = the residual consecutive-fire overlap, ~1 per item; 3 `ai_endpoint_unavailable`
  all on `efccd5d8` at 08:55–08:58 — entirely the credit-outage tail, none after ~09:00 recovery.
  **Post-recovery loss mode is overlap only, 46/(382+46) = 10.7%** — close to but not yet the
  single-digit target; the 24h read judges.
- **Throughput with demand beside it:** claims/h 30 → 118 → 199 (08/09/10h UTC), completions/h
  29 → 111 → 220, sites claimed/h 7 → 15 → 18. Arrivals/h 131 → 226 → 135; open depth now:
  triaged 1,268 / 30 sites (demand is not the constraint). 199 claims/h vs the ~300 ceiling
  estimate. ⚠ the pre-08:51 hours are outage-confounded (LLM handlers dead) — not a clean pre-B
  control; the trailing-24h pre-B stretch tomorrow carries the same caveat (handoff top block).
- **Batch cap binds:** avg loaded 4.66 per loaded loop against `max_items=5` (was 3.83 at 09:20)
  — Phase 3's case (5→8 both knobs) is now measured, queued behind the 24h read per HANDOFF-b.
- **Wait to claim** (items claimed ≥ 08:55, n=318): p50 **8.8 h**, p90 10.6 h — a backlog-drain
  figure (outage + filing wave), not steady-state; hold arrivals beside it tomorrow.
- **SAFETY — first non-zero in the double-handle census, investigated, NOT a double-handle:**
  post-B window census read 371 handlers / 349 items / **1 "overlapping pair"** (was 0 in 2,579
  on 08-25). The pair (`a52ac67f` / `d0f7ea9e`, item `b82f5cf5`): handler A stuck at
  `call_content_writer` ("Request … timed out after 3 retries" — the outage tail), sat zombie,
  stale-reaped 10:37:42 ("running for 1h21m20s, max age 1h0m0s"); the item's claim released
  earlier, B claimed 10:34:57 via the normal claim path and completed 10:48. Overlap = A's
  2-minute reap-lag tail vs B's live run. **Mechanism note: for a stale-reaped orchestration,
  `updated_at` is the REAP stamp, not end-of-life — raw interval overlap counts zombie tails.**
  Structural, will recur: the item-claim release and the orchestration reaper are separate
  clocks, and under a deep backlog the successor claims fast.
- **VERIFY 6/7 narrowed in consequence** (same-day, the r3-advisory precedent — a daily check
  that false-RAISEs teaches sessions to dismiss RAISEs): pairs whose FIRST-started member is
  stale-reaped (`error LIKE 'Orchestration stale%'`) AND whose second started > 10 min later are
  excluded from the violation count and reported as a **NOTICE** — never silently dropped. A
  stale-reaped member with a near-simultaneous partner still counts (a real claim race starts
  within seconds — first-claim p50 17.7 s). **Mutation-proved against LIVE data** `[MEASURED
  ~11:0x–11:2xZ]`: the scratch mutant with the exclusion disabled (`AND false`) RAISEs 6/7 on
  this very pair, exit 3; the real file then passed **all 7 + the zombie NOTICE ("1 zombie-tail
  pair(s) excluded")** — the strongest possible mutation pair: same live data, one predicate
  apart. ⚠ both runs took 10–20+ min, not the morning's 2m49s — the DB was IO-saturated
  (5+ concurrent full jsonb scans of orchestration_states from other lanes, 2 autovacuums);
  budget the daily VERIFY by DB load, not by its best time. LANDMINES entry appended for the
  general trap (updated_at-as-aliveness on stale-reaped rows) and the verifier dispatched
  (corr 808c4de2).
- Also observed (not diagnosed, not this lane's): the item was claimable ~3 min BEFORE the
  orchestration reaper stamped its stuck handler — two staleness clocks, claim-level and
  orchestration-level, by design or not.

### 2026-08-26 ~15:1xZ — afternoon artefact reads + a cross-lane inquiry answered (sibling ≠ outage pause)

- **Throughput held near ceiling all afternoon** `[MEASURED 15:0xZ]`: claims/h 206 → 134 → 265 →
  278 → 271 (10:00–14:00 UTC hours), completions/h 237 → 132 → 260 → 279 → 269 — against the
  ~300/h ceiling estimate. The 11:00 dip coincides with the DB IO saturation noted above.
- **The VERIFY's informational table grew a THIRD group with blank `row_task_name`** — resolved
  at the artefact, NOT a 637 defect: post-B loops stamp `task_name` correctly (306 of them,
  08:52→14:58). The blank group is a separate, pre-existing population of `build-dispatch-loop`
  spawns from another producer (42 pre-B / 21 post-B in the trailing 30 h, present since at
  least 08-25 21:20; 0 co-picks, ~10.7% lost). Liveness (4/7) is unaffected — the trigger's own
  loops carry the stamp. Left unattributed deliberately: which producer spawns them is a
  question for a quiet moment, not a defect hunt.
- **Cross-lane inquiry (bugs_open/391 lane, ~14:5xZ): "build-pipeline-trigger-2 is disabled —
  is that the credit-outage pause outliving its cause?"** Answered in full by SendMessage: no —
  it is owner ruling B (migration 637), applied 08:51Z, coincidentally minutes before credit
  recovery; re-enabling is blocked by VERIFY 2/7 and would REINTRODUCE the 94% co-pick waste.
  Their throughput figure ("146–154/h") disagrees with the artefact (265–278/h) — asked them to
  remeasure their meter. Their independent read of `find_dispatchable_site` (priority orders
  items WITHIN a site; `created_at` orders BETWEEN sites; skip-if-site-has-claimed-item) matches
  ours — a useful second derivation, cited here. They correctly did NOT flip the row. The
  incident validates 637's design choice of keeping the ruling mechanical (2/7) rather than
  documentary: the next session that pattern-matches "disabled = paused" hits a RAISE, not a
  silent regression.

### 2026-08-26 ~15:4xZ — the 391 lane's starvation handover DIAGNOSED: the selector and the loader disagree on ordering → bugs_open/413

The 391 lane handed over (CONTRIB addendum, explicitly undiagnosed): finetuning.uk, 73 eligible
items, zero claims in 6+ h while the fleet ran 265–278/h. Full evidence and mechanism now in
`bugs_open/413_HANDOFF_2026-08-26_selector_and_loader_disagree_on_ordering_so_a_pinned_item_starves_younger_sites.md`;
090 run `250188a7-29ae-4b3d-ace6-638694612c8b` filed BEFORE committing to the root cause
(07-31 ruling). The investigation path, for the record:

1. **The discriminator the handover couldn't see:** "never selected" vs "selected-but-loaded-zero"
   are identical in `site_work_items`; `orchestration_states` separates them. finetuning.uk had
   exactly ONE `build-dispatch-loop` in 12 h (05:08:53) → selection-side. `[MEASURED 15:1xZ]`
2. **Live selector read from the artefact** (agent_definitions, not a mirror — the 391 lane's
   mirror was wrong twice today, their own admission): full clause list incl. lock-exception,
   approval_mode, depends_on, busy-skip. All 73 finetuning items pass EVERY clause (cumulative
   breakdown 73/73/73; depends_on all NULL — the archived-dependency hypothesis from the
   [[a-closer-census-cannot-see-what-it-succeeded-at]] memory was checked and is NOT this case).
3. **The pin, measured:** 13 sites / ~570 eligible rows older than finetuning's oldest; 10 busy at
   the instant, 2 also starving (gaswholesalers 04:22, loanandmortgagecalculator 04:39). Pinned
   rows live: audit_tool @prio 140 attempt 0 on mortgagecalculator (created 23:53; the site took
   22 loops/95 claims in 3 h at prios 30–130 and never touched them), oufe, ai-agent-orchestration;
   fail-bounce flavour on loancash (@140, attempt_count 2). Sites drop out of service the moment
   their own old rows drain (fundamentallyai 35 loops → none after 14:13; idea.uk 19 → none after
   13:59) — the NOW-census can't show their history (rolling window), the loop cessation can.
4. **Ruled out en route:** loader-drop black hole (bugs_closed/078 shape — every looped site
   loads ~5/claims ~90%+), the NULL-task_name second producer (real, ~3-4 loops/h fleet-wide,
   serves some younger sites, neither causes nor masks), and everything the 391 lane had already
   measured (lock, busy-skip, retry_after/307, attempt exhaustion).

Lane consequences, all applied this entry: RUNBOOK gains §"Per-site starvation floor" (the
aggregate meters structurally cannot see an absence — quote the WORST site); the 24h post-B read
spec now includes it (HANDOFF top block); Phase 3 is flagged in 413 as cutting both ways (deeper
loads reach @140 rows sooner; longer busy windows per pick) — measure the floor across it, do not
assume. Ruling B neither caused nor cures 413 — pre-B the pair co-picked one deep site 94%, so
starvation was WORSE, just unmeasured.

### 2026-08-26 ~17:1xZ — 090 verdict on 413: UNVERIFIABLE (iteration-cap), and the round's real yield was a correction to MY sentence

Verdict: NOT confirmed, stopped at iteration cap, no fix proposed — neither CONFIRMED nor
REFUTED. Full reconciliation in 413's ~17:1x addendum: the loop's "oldest rows" sample was
`status='detected'` rows (selector-INELIGIBLE, July dates), its zero-row pin query is not
auditable (bundles record descriptions, not SQL), and its "dozen+ sites cycling" observation is
what 413 predicts, aimed at a claim the file does not make. Mechanism unamended; the file now
stands on declared first-hand verification per the 07-31 ruling, loop trail linked not hidden.
What the round DID catch: my symptom's "no trigger-driven dispatch at all" was an overclaim
(fall-through service exists — fundamentallyai's 35 loops were in my own evidence); corrected
visibly in 413, logged in WRONG_CALLS ("a graded verdict grades the SENTENCE you filed").
Deliberately NOT re-filing the 090 on the corrected sentence — the open question is now a
fix-candidate choice, and the census is repeatable from the RUNBOOK. 391 lane told (they asked
only if the mechanism was amended — it was not; their NOTES need no correction).

### 2026-08-26 ~21:4xZ — Phase 3 PREPARED (migration 658 written + council-submitted); tomorrow's protocol locked; two pieces of lane lore corrected by the exploration

Owner sanctioned Phase 3 + tomorrow's read this evening ("413 is being looked at in another
thread so you can move onto Phase 3 and tomorrow's read"), and chose "After the read" when the
sequencing was put to them directly. Coordination with the 413 fix session (their migration is
657): they apply ≥12:00Z tomorrow, after my boundaries, and their selector reads max_items LIVE
— no lockstep either way. Full protocol in HANDOFF-b top block.

**Exploration findings that changed the design (all verified at code/artefact, not lore):**
- Iterations are NOT unrolled in config — one `process_item` step with a sub_workflow; the
  `iter_N` names are runtime-generated (`loop_expansion_handler.go:101`). Batch 8 = two integer
  writes on paths `{workflow,steps,load_items,config,max_items}` and
  `{workflow,steps,process_item,config,max_iterations}`.
- "One without the other is a silent no-op" CONFIRMED (`loop_actions.go:197-203` truncation)
  with refinements: a THROUGHPUT no-op, harmless (the loader is a pure SELECT and claims
  nothing — surplus rows stay triaged); max_items-alone Warns to pod logs, max_iterations-alone
  is fully silent.
- ⚠ Both readers fall back to Go defaults on type mismatch: max_items→**50**
  (`load_work_item_actions.go:699`), max_iterations→**20** (`loop_actions.go:52-55`). So the
  values are bare jsonb numbers, the guard asserts `jsonb_typeof='number'`, and **the rollback
  sets 5 explicitly — key deletion (633's rollback shape) would run batch 50/20, ten times
  Phase 3.**
- Claims are per-iteration (only writer: `claim_work_item_action.go:96-113`; claim_result_N
  keys accrue one per iteration on live rows) → batch 8 adds NO claim-reaper exposure; what
  lengthens ~60% is the site's busy window (the selector's NOT-EXISTS clause) — the 413
  interaction, measured by the floor.
- **Honest sizing `[MEASURED 2026-08-26, 24h, n=1,588]`:** 80.3% of turns loaded exactly the
  cap; 26/29 sites ≥5 eligible, 22/29 ≥8 (full loader predicate incl. retry_after); one pass
  drains 134 vs 204 items. But turns lengthen ~60%, so expected throughput is **~+7%**
  (overhead amortisation) — NOT +52%, NOT PLAN:195's ~1.5× (that assumed the dead timeout
  half). Framed exactly this way in the council submission.
- New risk found and metered, not blocking: collected_data ×1.6 → ~1.9 MB avg / ~7.3 MB max
  vs the 8 MiB warn tripwire (`state.go:891-912`); watch post-apply (RUNBOOK query).
- D3 recorded satisfied-VACUOUSLY in WDS-002 (timeout half dead per bug 398); do not re-run
  051 (it would replace the whole loop workflow, not reset a knob).

**Files:** `658_dispatch_phase3_batch_8_HOLD.sql` + `_ROLLBACK` (refusal-BEFORE-snapshot per
the LANDMINES 2026-08-26 replay-decoy trap — 633's own order is backwards; whole-fleet md5
negative control in the guard). Council: DRY_RUN clean, submitted,
`SUBMISSION_CORR 95099f95-b32b-4656-bfe1-28f140de3717` — read the verdict before tomorrow's
apply if landed; `Council-Submitted:` trailer on the commit either way.

Note on who-owns: `who-owns.py 413` names THIS lane (we filed it) — the human owner's routing
of the FIX to the "bugs_open/413" session supersedes commit-history inference; levers + meters
stay here, coordination contract in HANDOFF-b.

## 2026-08-26 evening — 413 FIX BUILT (bugs_open/413 session, resuming the fix by owner routing; levers/meters stay with the lane)

Coordination first: throughput session confirmed the split in writing (~20:4xZ) — fix here,
levers there; **apply agreed NOT BEFORE 12:00Z 08-27** (after the 24h post-B read ~09:00Z and
658 ~09:30Z), stamp + ping on apply. Bug re-verified live before building anything
`[MEASURED 2026-08-26 ~20:1xZ]`: selector text unchanged (md5 d6f98acdb5aec385d5eb4077eac530fc,
the 633-era text); pin census **16 of 25 eligible sites pinned** — webdesign.co.uk
oldest_load_rank **135**, loanzy 62, idea.uk 53; worst wait ~5h (loanzy, last claim 15:10).
Evening inflow DEEPENS pins (15:5x census read 13/25) — dated trend, in the council submission.

**What was built (candidate 1, generalised): `657_selector_ranks_sites_by_loadable_work_HOLD.sql`
+ `_ROLLBACK` + `_VERIFY`.** Selector ranks sites by min(created_at) over each site's top-K
eligible rows under THE LOADER'S ordering (priority ASC, created_at ASC); **K read LIVE from
`load_items.max_items`** (the K agreement with 658 — their header defers the lockstep to 657,
657 reads their knob, nobody owes an edit; COALESCE→1 on a broken path degrades pin-free).
Eligibility clauses byte-identical to the 633-era text; output shape (one row site_id/domain)
unchanged; cost same class (~120ms real work vs ~115ms, both ~1.2-1.5s JIT-dominated, 1 fire/60s).

**Proofs run tonight, all at the artefact:**
- Full dry run in a rolled-back transaction (migration body → VERIFY → rollback body): preflight
  refuses drift by md5; guards pass; VERIFY passes on applied state (K=5; census 28 eligible /
  15 pinned at 20:46Z); rollback restores md5-exact; nothing persisted (exit 0).
- **Divergence, same instant:** OLD text picks webdesign.co.uk (the rank-135 pin), NEW picks
  vetcomparison.uk (oldest work a pick would actually drain). That pair of picks IS the bug and
  the fix in one measurement.
- VERIFY run standalone pre-apply FAILS on its md5 arm (exit 3) — the designed mutation proof.
- Go lockstep test `load_work_items_ordering_contract_test.go` (AST, house pattern): passes;
  mutation (ordering swapped at line 789) FAILS with the routing message; reverted, re-passes.
  ⚠ misstep-class near-miss, self-caught: I reverted the deliberate mutation with a bare
  `git checkout -- <file>` on the shared tree — safe THIS time (numstat clean, file untouched
  by others), but the safe form is a targeted inverse edit; a peer's uncommitted change in the
  same file would have been destroyed. Noted here rather than WRONG_CALLS (no false claim was
  recorded; the check that catches it is `git diff --numstat <file>` BEFORE any checkout --).

**Registered in the same commit** (ordering-exemption condition 2): WDS-002 gains the
"ordering contract CLOSED" bullet + a visible CORRECTION striking its own "bounded by
ceil(backlog/5)" claim; same correction block added to 284's header (the file that stated the
bias and called it bounded); LANDMINES entry "ONE ordering contract" appended + verifier
dispatched (corr 8aed215e). **Adjacent finding filed as `bugs_open/415`** (fire-gate pre_query
narrower than the selector: no 'approved', pipeline='build' — an approved-only backlog never
fires the trigger; theoretical at today's volume, first-hand verified, scoped OUT of 657).

**Deliberately NOT shipped: candidate 2** (age floor / positional-wait bound) — a policy trade
(per-site fairness vs item-age fairness) for the owner; presented in README_where_we_are with
the floor as its evidence base. With pins unrepresentable, every site's representative age
advances when served, so residual long waits are CAPACITY, not ordering.

### 2026-08-26 ~20:5xZ — 658 council APPROVED (round 1); advisories acted on the same hour; chassis roll noted for tomorrow's window

**Verdict:** APPROVED — "1 advisory objection(s), none high-severity", 7 abstained, corr
`95099f95`. Read in full. Three actionable points across editquality (medium) and guardian
(medium + low), each now DONE with evidence `[MEASURED 2026-08-26 ~20:5xZ]`:

1. *"jsonb paths asserted but not verified; jsonb_set create_missing would mint a wrong-path key
   and the guard reads the path it wrote"* (editquality) — covered MECHANICALLY by the file's own
   refusal-first block: it reads BOTH paths and RAISEs 'knob path missing' on NULL before any
   write, so a wrong path aborts instead of minting. Additionally the paths were read verbatim
   from the live row during design (step JSON in the exploration report). No edit needed.
2. *"four agent types carry TWO active definition rows — confirm build-dispatch-loop is not one,
   and which row the loop LOADS"* (guardian, medium) — row census: build-dispatch-loop has
   exactly ONE row in ANY state (id 099b51e0, active, non-snapshot, version 1, not deleted).
   The four duplicate-active types are content-creator, content-creator-contact,
   chief-strategist, site-component-architect — ours is not among them; "which row loads" has a
   unique answer. Check recorded here; run the same census again immediately before hand-apply.
3. *"enumerate other configs reading these knobs by convention"* (guardian, low) — census over
   all live configs: 23 agents carry `max_items` and/or `max_iterations` somewhere, but exactly
   ONE other agent uses the `load_work_items` action: **site-work-orchestrator** (the
   human-initiated build path, the known second caller from the 396 work). 658 deliberately does
   not touch it — human-initiated batch stays as-is, now stated as a census, not an assumption.
   No agent reads build-dispatch-loop's values across rows.

**Chassis roll ~20:26Z (owner: "a fresh chassis build has been deployed"):** pods
`agent-chassis-5864bf97c5-*` age 25m at 20:51Z. Dispatch continuity measured across it: trigger
fires 1/min and loops ~1/min unbroken 20:06→20:51 — **no dispatch gap** (the ~300s
no-dispatch-after-restart caution did not manifest; likely staggered pod replacement).
Provenance line not in `--tail=300` (startup line scrolled — "not in range", not "unstamped");
what shipped is not this lane's to verify tonight. Tomorrow's 24h read holds the roll timestamp
beside the window.

Commit carrying this entry carries `Council-Reviewed: 95099f95` — verdict read in full,
advisories dispositioned above.

### 2026-08-26 ~21:2xZ — 657 council r1 REVISE (a REAL catch), r2 in flight; one loader fact adopted into this lane's practice

Reported by the 413 session (their lane's work, recorded here because tomorrow's all-clear gates
on it): r1 on `ecf2e542` returned REVISE ~12 min in — debug_historian HIGH, and RIGHT: their K
subquery selected the loop row by `ORDER BY updated_at DESC`, but **the runtime loader selects
by `version DESC`** (`loadAgentDefinition`, `processor.go:371-389`); `updated_at` is degenerate
fleet-wide (199/200 live rows share one microsecond — the LANDMINES entry this lane already
carries). Moot today (one active row, their preflight asserts it) but a latent wrong-row read.
Fixed (`9ff62d8a4`: K mirrors the loader's rule verbatim; UPDATE pinned by captured row id),
proofs re-run green 21:10Z, r2 submitted ~21:13Z on the same correlation.

**Adopted here:** any query this lane writes that reads an agent's LIVE config must select the
row the way the LOADER does — `version DESC` among active rows — never by `updated_at`. 658 is
unaffected (its WHERE hits the single row and RAISEs on ROW_COUNT≠1; the one-row-in-any-state
census is re-run before apply), but the rule goes into tomorrow's practice: when in doubt about
which row the runtime reads, mirror `processor.go:371-389`, not a timestamp.

Slot logic unchanged and restated: **no acted-on verdict = no 657 apply.** If r2 lands APPROVED
it will be before the 09:00Z read; a further REVISE is theirs to act on before the ≥12:00Z slot.

## 2026-08-27 — the 24h post-B read: GATE PASSED; VERIFY false-RAISEd on a SECOND reaper spelling (fixed + mutation-proven); floor baseline handed to 657

**Window honesty first:** the read ran `[MEASURED 2026-08-27 08:37–09:00Z]`, not 09:00Z sharp —
my wake timer parsed "09:00" in LOCAL time (BST) so it fired immediately; the window is explicit
in every query (`>= 2026-08-26 08:55Z`, ~23.8h) and retention was verified INTACT back to 08-26
08:06Z before anything ran, so nothing was lost — the read is simply dated 08:37Z, not 09:00Z.
(Misstep + check: `date -u -d 'today 09:00'` parses the TIME in local TZ; give the zone in the
string. Snapshot queries — floor, pin census — executed 08:48–08:50Z.)

Caveats held beside, per handoff: credit outage 08-25 23:47→08:55Z sits on the PRE-B side (its
only in-window trace: 3 `ai_endpoint_unavailable` claim refusals 08:55–08:58 on item efccd5d8);
chassis roll ~20:26Z inside the window (no dispatch gap, measured last night).

### The meters `[MEASURED 2026-08-27 08:37–08:50Z, window ≥ 2026-08-26 08:55Z]`

- **Cadence:** 1,307 fires, gap p50 **60s** / p90 105s / max 724s. 13 gaps >300s, clustered
  02:00–04:30Z overnight (worst in-cluster 650s) plus the known 11:18Z DB-saturation window —
  slow ticks under load, same shape as yesterday's p90; not a stall (loops continued).
- **Sibling frozen** ✓: `-2` disabled, last_triggered_at 08-26 08:51:09 unchanged; parity
  one-liner clean at 08:28Z (one pre_query md5 across both rows).
- **Lost claims: 16.5%** (1,067 of 6,480 attempts; was 58–60% under the pair). Anatomy: 1,066
  `already_claimed_or_ineligible` over 755 distinct items (~1.4 per item — consecutive-fire
  overlap residue), 3 outage-tail endpoint refusals. ⚠ trend: 10.7% at yesterday's 2h read →
  16.5% over 24h — rising with depth, hold beside Phase 3 (longer turns may raise overlap).
- **Steering:** 1,376 loops / **32 distinct sites**; 10–24 sites per hour, no co-pick mode.
- **Batch cap binds hard:** avg loaded 4.91/5; **1,319 of 1,375 loaded loops (95.9%) at cap** —
  Phase 3's premise measured again, stronger than yesterday's 80.3%.
- **Throughput:** claims/h 170–309 sustained against the ~300/h ceiling (peak 309 at 19:00Z);
  completions track claims (up to 307/h). Demand beside it: arrivals 130–485/h all night;
  open depth NOW **716 triaged / 31 sites** (was 1,268/30 yesterday — the backlog is DRAINING
  while arrivals continue: throughput > arrivals over the window).
- **Wait-to-claim** (n=4,945): p50 **1.9h** / p90 13.3h (was p50 8.8h at the 2h read — the
  drain figure improving as predicted).
- **Double-handles: 0 true.** Raw census 5,333 handlers / 5,050 items / 181 items with 2+
  handlers (sequential retries) / **3 raw overlapping pairs**, every one discriminated at the
  artefact as a zombie tail (below).

### VERIFY 6/7 false-RAISE — the SECOND reaper spelling (fixed same-day, commit adebc2d11)

Daily 584 VERIFY (started 08:37Z, RAISEd ~16 min in under load): **6/7 DOUBLE-HANDLE: 1 pair**,
exit 3. Investigated before ruling the gate:

- The 3 raw pairs: two match yesterday's exclusion exactly (first member FAILED `Orchestration
  stale — running for 1h2xm`, successors 79–81 min later) and were excluded as NOTICEs. The
  third — the RAISE — is pair `0d699d65`/`fb7e9e0f` on item `61265835`: tool-generator wedged
  at `suggest_related_pages`, stamped FAILED by the **step-level reaper** with the spelling
  **`reaper: stale EXECUTING_STEP for >4h; step=…`**, which `LIKE 'Orchestration stale%'`
  cannot match.
- **The claim invariant HELD — proven serial at the item:** loop 3ae202fb claimed 09:10:45,
  spawned A 09:14:21; claim released by the claim-level staleness clock; loop e089cf20
  re-claimed 11:49 (`claimed=true`, normal atomic path), its handler B completed the item
  11:54:14. A sat zombie until the reap stamp 13:37:26. Successor gap 2.6h — nothing like a
  claim race (first-claim p50 17.7s).
- **Fix:** exclusion widened to both reaper spellings, OR properly parenthesised inside the AND
  chain (the 396 lane's CONTRIB precedence lesson, applied the same morning it arrived);
  FAILED-status and >10-min arms kept. **Mutation-proven with the window PINNED to 08-26
  08:55Z** — the pair's loop ages out of the trailing-24h window at ~09:10Z, so an unpinned
  re-run could pass VACUOUSLY ([[pin-the-clock-to-before-the-failure]]): edited exclusion
  **0 violations / 3 excluded**; new-arm-disabled mutant **1 / 2**. Same live data, one
  predicate apart. Full VERIFY re-run on the edited file: [appended below when landed]
- Residual honesty: the two stale-reap clocks (claim-level release vs orchestration-level reap)
  remain separate by design or accident — yesterday's open observation stands, not this lane's.

### Per-site floor + pin census `[MEASURED 2026-08-27 08:48–08:50Z]` — the 657 pre-fix baseline

- ⚠ **The floor query needs a LOCK CONTROL:** its worst row, adversecreditmortgage.co.uk (70
  eligible, no claim in 27h), is **locked since 08-18 with except_n=0** — parked by design (the
  396 lane's "held items"), NOT starvation; absent from the pin census because that predicate
  checks locks. Next floor read: join `sites.locked_at` before quoting the worst site.
- **Worst genuinely-starving site: lendzy.co.uk** — 55 eligible, oldest waiting **10.6h**
  (created 08-26 22:16Z), last claim 00:57Z (~7.9h before the read), **pinned at
  oldest_load_rank 44**. The 413 mechanism live and quotable.
- Census: **10 pinned of 25 sites with eligible work** (webdesign.co.uk rank 22,
  loanandmortgagecalculator 23, gaswholesalers 70, lendzy 44, robot-hands 39, lampenkap 30,
  finetuning 17, loanzy 13, oufe 10, leopardessconsulting 7). Dynamic snapshot, dated. Several
  pinned sites show recent last_claims (their YOUNG rows get served; their oldest never does) —
  pinned ≠ unserved, pinned = the oldest row never drains and age-order victims queue behind it.
- dartsonline.com: 2 eligible, loadable at rank 2, unserved 5.6h — a pure positional victim.

### GATE: **PASS** (ruled ~09:1xZ)

p50 60 ≤ ~65 ✓ · lost 16.5% ≪ 58–60% ✓ · 0 true double-handles ✓ (all 3 raw pairs
artefact-discriminated) · VERIFY: red this morning on the proven-false detector arm, green
pending the re-run of the corrected file — the mutation pair above is the evidence the gate
rests on. Proceeding to 658 per the handoff (~09:30Z): one-row census re-run first, then
hand-apply, artefact-verify both knobs.

### 2026-08-27 09:15Z — VERIFY re-run GREEN; 658 APPLIED and artefact-verified

- **Full VERIFY on the corrected file: all 7 hold**, exit 0, with the honest NOTICE "3
  zombie-tail pair(s) excluded" `[MEASURED ~09:1xZ]`. Informational table: 1,303 trigger loops,
  **0 co-picked**, lost 17.4%; the blank-task_name second producer at 80 loops / 1.9% lost
  (still unattributed, still benign). Detector correction committed `adebc2d11` (pathspec, one
  file; _VERIFY is outside council scope by design).
- **658 hand-applied 09:15:06Z**, ON_ERROR_STOP clean: refusal-first passed, snapshot captured
  (source 099b51e0 v1), guard NOTICE "8/8, both jsonb numbers, config shapes intact, no other
  row changed", COMMIT. One-row census re-run seconds before (guardian advisory): still exactly
  ONE build-dispatch-loop row in any state, knobs read 5/5 pre-apply.
- **Artefact verify `[MEASURED 09:16Z]`:** max_items **8** (jsonb number), max_iterations **8**
  (jsonb number) on the live row selected by the loader's predicate.
- Honest expectation on record: **~+7%** (overhead amortisation) — the cap binds (95.9% of
  loaded loops at cap this window) but turns lengthen ~60%. Watch: capability probe (loaded up
  to 8, claim_result keys past _4) ~09:30Z; collected_data sizes (~×1.6 → ~7.3MB max tail vs
  8MiB warn); the ~11:30Z 2h read cuts the Phase-3 window; 657 applies ≥12:00Z on our all-clear.

### 2026-08-27 ~09:5xZ — cross-session amplifier for 413 (from bugs_open/414): a dropped-spawn claim darkens its whole site, unreapably; class confirmed at oufe

The 414 session messaged (incidental find): the selector's busy-skip clause EXCLUDES a site
with any `claimed` row, and a claim whose spawn was silently dropped (zero orchestrations,
ever) is covered by NO reaper — their lendzy `content_rewrite` sat claimed 35 min this morning
(08:51:52→hand-release) with the site invisible throughout. Actioned here:

- **Verified what was verifiable:** lendzy's rows had already archived out (rolling window —
  their event stands cited, not re-verified). But the CLASS is live independently:
  **oufe.com, 2 claimed rows, zero orchestrations** (`page_rerender` 09:08:38, `audit_tool`
  09:11:03) `[MEASURED 09:48Z]` — oufe dark inside our Phase-3 window. 14 sites / 30 claimed
  rows fleet-wide at the same instant (busy-skip exposure, mostly legitimate).
- 413 gains the dated addendum (mechanism, worked case cited, discriminator query, the
  MUST-run-before-grading note for 657's acceptance reads); RUNBOOK floor gains the
  stuck-claim control beside the lock control.
- **TTL experiment on oufe, deliberate non-intervention until ~10:15Z:** the 08-26 zombie pair
  had a claim-level release ~1h18m in — but WITH an orchestration. If oufe's rows self-release
  by ~10:15 (age ~1h05), "unreapable" is overstated and a TTL exists; if not, second-site
  confirmation, and this session releases them under 414's guard (status still claimed,
  >30 min, zero orchestrations) so oufe is back in service before the ~11:30Z read either way.
- Their "~80 claims/h fleet-wide" disagrees with our measured 170–309/h — asked them to
  remeasure (site_work_items is a rolling window; a NOW-census undercounts a rate).
- **Confounder flag for our own reads:** the 2h/24h Phase-3 floors must run the discriminator —
  a dark site may be EXCLUDED (this), not OUT-ORDERED (413) or slow (capacity).

### 2026-08-27 ~10:2xZ (recorded 13:2xZ) — oufe TTL experiment: 414's "unreapable" REFUTED; the mechanism is `claimed-item-timeout`

The ~10:15Z check found oufe's two stuck rows already RELEASED at ages ~42 min (09:50:33 /
09:53:29 — both PREDATE this lane's 09:58Z message naming oufe, so mechanically, not tipped),
signature: status→triaged, claimed_at NULL, attempt_count+1, retry_after now+30min. The
mechanism, read in full: **`claimed-item-timeout`** (scheduled_tasks, interval 120s, enabled —
not named "reaper", which is how BOTH sessions' censuses missed it). Three stages:
auto-complete claims >15 min whose handler orchestration COMPLETED after claim; auto-complete
>15 min on artefact evidence; **reset >40 min** with per-type backoff (`reaper_policies`,
default 30×attempt min), error 'Claim timed out — handler pod likely died'. So a dropped-spawn
dark window is **BOUNDED ~40–42 min**; 414's lendzy hand-release at 35 min preempted the sweep
by ~5 min, which is why their observation never saw it. 413 addendum corrected visibly;
RUNBOOK control re-worded (a >45 min stuck claim now means the timeout task ITSELF is broken).
**Frequency [MEASURED 10:20Z]: ≥89 resets today by 10:20Z across 27 sites** (oufe 11, loanzy
10, idea.uk 9) — caveats in the 413 correction (decay confounds day comparison; includes
failed handlers; the 15 verifier-gated types ALWAYS land in reset even on success).

### 2026-08-27 13:1xZ — the "2h" Phase-3 read ran at ~4h (session paused ~10:40→13:11 on MY account's usage limit); read window 09:15:06→13:11Z; a FLEET LLM outage started 11:30Z and is ONGOING

Session-pause honesty: the ~11:30Z read fired at 13:11Z (my Claude session hit its own usage
cap mid-morning). The window is explicit, so the read is simply LONGER; the 657 slot slipped
past ≥12:00Z with the all-clear unsent — sent 13:2xZ (they hold until it arrives; selector md5
read d6f98acd at 13:15Z = correctly unapplied, read by version DESC).

**Phase-3 capability, both knobs PROVEN at the artefact:**
- 163 of 239 loaded loops loaded exactly **8** (distribution 1–8, cap binds at 68%).
- max claim_result suffix **_7** on completed loops (0-indexed = **8 claim attempts/loop**).
- collected_data: completed avg 343KB / max 1.4MB vs 8MiB warn — headroom holds.
- Cadence through everything: p50 59s / p90 121s / max 424s.

**The window is TRIPLY confounded — do not grade ~+7% on it:**
1. ~09:50–11:00: claim-churn stretch (the reset storm's +30min backoffs thinned
   instant-eligibility; claims 42–52/h with loops normal; ZERO endpoint refusals here).
2. **11:30Z→ongoing: fleet LLM outage, 100% call failure** `[MEASURED llm_call_log 13:15Z:
   12h 61/61 failed, 13h 17/17, last-15min 20/20]` — error "You have reached your **specified
   API usage limits**" = the SELF-SET Billing-page limit (08-17 signature, NOT the 08-25/26
   prepaid-credit one; D7 distinguishes them). Dispatch-side trace: `ai_endpoint_unavailable`
   claim refusals 0 before 11:30, then 84/106/56/48 per half-hour. LLM-free work (rerenders)
   still completes; LLM-bearing claims refuse at the claim step. **D4's case, third instance,
   live.** Owner action needed on the account; noted in README.
3. Lost-claim split over the window: 45.5% headline = 283 endpoint (outage artifact) + 293
   overlap. Overlap-only ≈ 29.8%, concentrated pre-outage (273 of 293 before 11:30) — HIGHER
   than the 24h read's 16.5%; whether that is batch-8 (longer turns → more consecutive-fire
   overlap) or reset-storm churn is NOT decidable on this window. The clean post-outage
   window judges; flagged for the 24h Phase-3 read.

**Floor/pins at 13:11Z (with lock + stuck-claim controls):** worst unlocked =
**cookly.uk, 6.2h unserved, NOT pinned (rank 2)** — a pure positional victim, 657's exact
target (its 3 rows re-eligible since ~08:00 after claim-timeout resets at ~07:0x). Batch 8
already UNPINNED some of yesterday's pins by widening the load window
(loanandmortgagecalculator rank 23→6, loanzy 13→2); still pinned at K=8: **10 of 25**
(lendzy 35 — oldest now ~15h, gaswholesalers 66, idea.uk 54, relojistas 48). These are the
657 pre-fix baselines alongside the 09:00Z read.

**657 all-clear judgement: SENT.** The outage is orthogonal to the selector swap (DB-only
apply, DB-only VERIFY, no LLM in the query); starvation damage is live and ongoing (cookly);
their acceptance reads must (a) cut windows on the outage boundary, (b) run the stuck-claim +
lock controls before grading a dark site, (c) during the outage measure SELECTION fairness
(loops per starving site), not drain.

### 2026-08-27 13:18:19Z — 657 APPLIED (their session, on our all-clear); windows cut; floors scheduled

Their stamp 13:18:19Z, VERIFY green 13:18:30Z (md5 d29807313; **K=8 read LIVE from 658's
knob — the K-agreement worked unedited across the 5→8 change**); census at verify 28 eligible /
11 pinned (pins persist as rows; they no longer freeze the age order). Confirmed from this side
at the artefact: selector md5 d2980731 by version DESC `[MEASURED 13:19Z]`. Their apply probe
picked loanandmortgagecalculator.co.uk — one of yesterday's measured starvers. Their commit
25e92db4c records all four context items verbatim.

**Division agreed:** this lane takes the **+2h (~15:20Z)** and **+6h (~19:20Z)** per-site
floors against the 09:00Z and 13:11Z baselines — acceptance bar per 413 §"How to verify"
(no site with eligible work > ~1h unserved while pins exist elsewhere), graded with the lock +
stuck-claim discriminators, and during the LLM outage (still 100% fail at 13:19Z) on
loops-per-starving-site rather than drain. They spot-check the first ~10 fires. Bug 413 stays
OPEN until measured.

## 2026-08-27 ~13:3xZ — 657 APPLIED (13:18:19Z, VERIFY green, K=8) + the 396 CONTRIB adopted (413 session)

Applied on the lane's all-clear; stamp confirmed at the artefact from both sessions
(md5 d2980731 by version DESC). Full apply record + acceptance context in 413's apply
section; lane takes the +2h/+6h floors. **The deferred_work_item_park CONTRIB (guard
presence-tests cannot see a precedence break) is ADOPTED**: 657's FOREACH literals now pin
the four OR-bearing fragments WITH their wrapping parens (their measured case: paren drop =
1,104 → 15,683 admitted rows, no clause "dropped"). Their honest caveat kept verbatim in the
guard comment — a substring cannot prove parens BALANCE; the leading '(' catches the
realistic edit, and the VERIFY's md5 arm pins the live text byte-exactly. The live row is
unaffected (already applied; replay refused by md5); this hardens the file-as-template for
the next migration author who copies the guard shape.

### 2026-08-27 ~13:3xZ — reset-census composition MEASURED (answering 414's open question): gated types do NOT dominate

414 predicted the verifier-gated fallthrough types "will dominate any count". Measured today
(127 resets by 13:3x): **gated = 5/127** (all `unbuilt_internal_link`); `page_rerender` = 66
(has an evidence-completion arm, so its resets are genuine 40-min holds), then
undeployed_asset 12, content_rewrite 11. So the incident character of the census stands; the
413 caveat updated in place. Unattributed residue: why 66 LLM-free rerenders held claims
>40 min today (deploy latency? the 09:50 reset batch?) — noted, not chased; not this lane's.

### 2026-08-27 ~13:4xZ — the "66 rerender resets" residue answered same-hour (414's deploy-step lead)

414's lead: their own repair saw `deploy_page failed: timed out after 3 retries` AFTER the save
had succeeded — so a rerender hold can be work-done-claim-aged. Measured on today's 66
`[13:4xZ]`: **54 pages deployed at/after the claim window; 56 of 66 items have since reached
complete; 2 never deployed; 9 on retry backoff; 1 failed.** Approximation stated: claim time
taken as reset-stamp minus 40 min, and "since completed" cannot split already-done-at-reset
from done-on-retry without per-item timelines. Conclusion for this lane: these resets cost
40-min darkness windows + retry churn, NOT lost work — the floors' discriminator handles the
darkness; nothing further owed here. (If anyone later chases the deploy step's own timeout:
414's about-page case is the worked example.)

### 2026-08-27 ~13:2x–13:3xZ — 657 first-fires spot check (their session): selection landing on the formerly starved

Their read, 13:18–13:26Z (8 fires / 9 loops): loops went to loanandmortgagecalculator ×4
(yesterday's 04:39 starver), loanzy ×3 (rank-62 pin), idea.uk ×1 (rank-53), robot-hands ×1 —
i.e. the new selector immediately serves old-loadable-work sites. Same-site repeats every
~2 min are the outage shape as predicted (LLM claims refuse → rows release → site stays
oldest-loadable and non-busy): selection fairness GOOD, drain blocked by the outage. Cadence
normal (~60s). No anomalies; the 15:20Z floor is this lane's next read.

### 2026-08-27 ~13:36Z — OUTAGE OVER: owner added credit; recovery at ~13:35:00Z

Owner (chat): "I have added a bit more credit". Call-by-call at the artefact: failures through
13:34:41, successes from 13:35:13, none failing after `[MEASURED 13:36Z]`. **Outage window:
11:30 → ~13:35Z (~2h05m), 100% failure 12:00→13:35.** All Phase-3/657 windows cut on both
edges. The 15:20Z floor lands ~1h45m post-recovery — drain gradable again. D4 case 3 closed
as an incident, standing as the governor's evidence.

### 2026-08-27 15:20:08Z — +2h post-657 floor: ACCEPTANCE BAR MET (wide margin)

All four sections `[MEASURED 15:20:08Z]`, windows: 657 live 13:18:19Z, drain gradable from
13:35:00Z (outage end).

- **Acceptance question** (unlocked sites, eligible >1h old, unserved >1h): TWO rows — cookly
  and loancash, each ONE ~67-min-old eligible row on a site served 66–70 min ago, oldest at
  load-rank 1 (no pin involved). Normal rotation spacing at 29 active sites, not starvation.
  Baselines: 09:00Z lendzy 10.6h pinned-44; 13:11Z cookly 6.2h. **Worst is now ~68 min.**
- **Old starvers resolved:** loanandmortgagecalculator oldest 04:32-yesterday → 14:35-TODAY
  (backlog drained, rank 1, 10 loops post-657); lendzy 46→15 eligible, oldest (22:16-y) now
  rank 5 — about to drain; finetuning oldest 01:51→06:20 draining, served 15:19; idea.uk rank
  54→15, served 15:19; cookly served 14:10.
- **Stuck-claim discriminator: 0 rows** (timeout task healthy, no dark sites at read).
- **Drain at ceiling post-recovery:** 473 claims / 480 completions in 105 min (~272/h, 29 sites).
- **Residual as designed:** pinned ROWS persist and age (farmerinsurance @23:20-y rank 77/80,
  gaswholesalers @23:41-y rank 64/68, noted.co.uk fresh burst rank 16) — 657 stops pins
  freezing the AGE ORDER; it does not drain them. That is candidate 2's policy question,
  already with the owner in README.

**Verdict: PASS at +2h.** +6h read ~19:20Z (timer armed) closes the day's acceptance.

### 2026-08-27 19:21Z — +6h read BLOCKED at the door: kubeconfig token expired 19:11:20 (the 3-day cycle)

The read fired on time (19:21:09Z) and hit `Unauthorized` on every kubectl — token expiry
confirmed by decoding the JWT (expired 19:11:20, ten minutes before the read; the
[[kubeconfig-token-expires-every-3-days]] shape exactly). Owner notified (terminal push);
refresh watch armed — the read runs the moment the token is back. The +6h therefore becomes
+Nh with the actual timestamp recorded; the +2h PASS stands as the day's primary acceptance
evidence. Nothing about the fix is in doubt from this; it is an access outage on the meter,
not the mechanism.

## 2026-08-30 ~21:1xZ — session resumed after ~3 frozen days (credit); long-run floor PASSES; VERIFY green at lost 3.9%; a NEW 2-day LLM outage live; one meter lesson

**Continuity honesty:** the session froze ~08-27 15:45Z (credit) with the +6h read already
blocked by the kubeconfig 3-day expiry (19:11:20Z); resumed 08-30 ~21:09Z with the kubeconfig
fresh. The 08-28 clean-window Phase-3 read is UNRECOVERABLE (orchestration_states retention;
say plainly: lost). No other session ran the lane's reads meanwhile (git log clean).

**Long-run (+3.3d) post-657 floor `[MEASURED 2026-08-30 21:09–21:15Z]`: PASS, trivially.**
The 08-27 ~700-row backlog is fully DRAINED: six eligible rows fleet-wide (five sites, all
`needs_page` @99 created 18:37–18:41 tonight, attempt 0, cycling on retry_after deferrals —
consistent with the live LLM outage below). Zero pins, zero stuck claims, trigger 29 fires/h.
2,273 claims / 2,312 completions still visible since 08-27 13:35 (rolling window: an
undercount of the true 3-day drain).

- **Meter lesson (RUNBOOK caveat added):** the floor's `last_claim` = max(claimed_at) is
  BLIND to claim-release cycles — a release (deferral or timeout reset) clears claimed_at.
  mortgagecalculator read "last claim 08-28, 2.5h unserved" while taking 9 loops in 2h on a
  deferring row. The service meter is LOOPS; the acceptance query's loops column is what
  exposed the contradiction.

**Daily 584 VERIFY (3 days missed — those windows lost): all 7 hold**, 5 zombie tails
NOTICEd; informational: 562 loops / 0 co-picks / **lost claims 3.9%** (19 of 482) in the
trailing 24h — the best figure the lane has recorded (58–60% pre-B → 16.5% → 3.9%, the last
step on a shallow queue with 657 live).

**⚠ NEW FLEET LLM OUTAGE, ~2 days old, LIVE:** llm_call_log fails/day — 08-28: 482/1,356
(36%); 08-29: **758/760**; 08-30: **654/659**; last 30 min **25/25**; last success 16:12Z
today; same "You have reached your specified API usage limits" 400 family. This is why the
queue drained to near-zero demand and why the six needs_page rows defer. Owner push-notified
(terminal; Remote Control inactive). D4's case count: four.

Windowing note for any future Phase-3 grading: clean post-657 LLM-healthy windows are scarce
— 08-27 13:35→~08-28 morning is the main one, and it is already out of retention. The ~+7%
question may only be answerable after the account is topped up AND demand returns; do not
force it.

## 2026-08-31 ~09:45Z — LLM outage OVER (recovery ~09:00Z today); post-recovery sanity read CLEAN; lane idle pending D4

Timeline correction against my own first read this morning (an aggregate trap, caught in
minutes): "4 failed of 49 last hour" at 09:42 was the RECOVERY EDGE, not a recovered day —
the night ran 100% failure (02:00–07:00 hours all-fail; one isolated success 22:05 08-30),
08:00 hour 32/42 failed, **09:00 hour 0/35** `[MEASURED 09:45Z]`. So the outage ran
~2026-08-28 → 2026-08-31 ~08:40–09:00Z (~2.5 days), D4 case 4's full extent.

Sanity read per the handoff's NEXT-1 `[MEASURED 09:45Z]`: trigger 53 fires/h; arrivals
ramping (16/60/29 last 3h) and claims following (4→14/h); floor CLEAN — three sites with
work, all fresh (farmerinsurance oldest 06:37 with 19 loops/3h = actively serviced on
deferral cycles; finetuning 8 rows @09:38; vetcomparison 1 @09:39); **zero stuck claims;
zero pins**. Nothing owed dispatch-side. The lane is idle on its build queue: **D4 governor
first, blocked on ONE owner decision — the at-cap shedding policy** (what gets refused first
when spend approaches the cap: maintenance vs build vs research classes; RESEARCH §6 has the
per-class cost sketch). Everything else per HANDOFF_2026-08-30.

## 2026-08-31 — OWNER RULED the D4 shedding policy; the governor build UNBLOCKS

**Ruling verbatim (chat, 2026-08-31):** "the shedding policy is routine maintenance first,
new site builds second and research third."

So the shed ORDER at-cap: (1) routine maintenance — first to be refused; (2) new site
builds; (3) research — most protected, shed last. D4 is now unblocked as the lane's first
build item. Design work starts from RESEARCH §6 (per-class cost sketch) + the 08-21 decision
table rulings; council review before/alongside the commit as platform code.

## 2026-08-31 — D4 STAGE A BUILT, COUNCIL-SUBMITTED, APPLIED; the meter's first real figure

Design in PLAN §D4 (committed b2e18b9b8); the 08-21/08-31 shed-order conflict marked visibly
in RESEARCH §10 and flagged to the owner in chat (they can object; the 08-31 direct answer
supersedes). Prices verified at platform.claude.com same day — NOT from memory (sonnet-5
$2/$10/$0.20 introductory made permanent). Migration **671** + guarded ROLLBACK:
- Dry-run in a rolled-back transaction on the live DB: exit 0, zero objects persisted.
- Council corr `80df0963-12d9-46dd-b122-30258f57a8e9`, submitted after clean DRY_RUN admission
  (first attempt used a flat plan array — the schema is plan{summary,edits,grounded_in,risks});
  commit `82683fe07` carries `Council-Submitted:`. Verdict owed a read (~30 min).
- **Applied 12:40:31Z**, artefact-verified: governor_state level 0 / **$2,113.08 August MTD** /
  3.68M unpriced io tokens (gemini — surfaced, not dropped); config enabled=f, budget NULL.
- ⚠ owed proof: the verify's own EXECUTE stamped computed_at — the SCHEDULER firing the task
  is a separate fact ([[thunder-reaper-fires-but-has-never-reaped]]); check computed_at
  advances past 12:40:32 unprompted. [result below]
- **For the owner's budget number: August ran ~$2,113 with four outages truncating it.**
  Stage B (Go claim-step refusal, opt-in) is next; C (interval ≤25s) stays gated on the
  governor being live + exercised.

### 2026-08-31 ~13:1xZ — 671 council r1 REVISE (fair catch: NO prior-art search was done); answered with evidence + migration 672; r2 in flight

**The gating objection was RIGHT procedurally**: reuse_agent (HIGH) — nothing in grounded_in
showed `platform/governance/fuel.go` had been checked before designing a new governor; and
prior_art_librarian — the load-bearing "nothing sheds deliberately" absence claim carried no
census. I had not looked. The checks, run after the round:
- **fuel.go read in full (74 lines): ORTHOGONAL** — a per-TASK fuel budget in a Kafka header
  (abstract units, stale model vocabulary, 2 call sites: coordinator.go:80/104,
  contentcreator/agent.go:241/262/845). Depth-limiting per message, not an account-month
  dollar authority. Ruled out with the quotes in r2, and recorded in the new doc_plans entry.
- **fleet-step-token-pressure + council-seat-token-pressure pre_queries read**: both audit
  max_tokens CAPS against truncation — neither prices tokens nor aggregates spend.
- **Censuses**: scheduled_tasks ~ 'pressure|budget|govern|spend|fuel' = exactly 3 rows; Go
  grep ShedLevel|SpendGovernor|spend_governor|monthly_budget = zero hits outside this work.
**The three LOW asks became migration 672** (dry-run rolled-back, applied 13:12:13Z, commit
2b522026f): advisory lock woven into the spine CTE (an unreferenced SELECT CTE may never
evaluate — it must be on the chain); verify DRIVES the level-change note both directions
(0→3→0, 2 notes, deleted by captured id — self-cleaning probe); doc_plans travelling design
for pipeline/spend-governor. Replay guard = md5 tri-arm (not-applied 1c371a33 / replay
838f8cd1 / drift = refuse). **r2 submitted on the same corr 80df0963.**
Lesson, same shape as the lane's own rule about measuring before submitting: **the absence
claim IS a census, and it goes in grounded_in BEFORE the design** — the seats exist because
sessions (this one included) design first and search never.

### 2026-08-31 ~13:3xZ — 671/672 r2 APPROVED; the advisory closed with its own sharper form (673)

**APPROVED** on corr 80df0963 ("1 advisory objection, none high-severity"; the "(round 1)"
header is the template literal — this is round 2). Verdict read in full. The advisory
(editquality): the xact-scoped advisory lock protects only the statement's span. Disposition:
the statement span IS the whole race window — but chasing the concern found the REAL residual:
under READ COMMITTED a fire that BLOCKS on the lock keeps its pre-block SNAPSHOT, so its
`old` CTE reads a stale level on unblock → duplicate/missed note. **Migration 673** (dry-run
rolled-back, applied ~13:3xZ): `FOR UPDATE` on the `old` read (EvalPlanQual re-reads the
committed row on lock release), plus debug_historian's rowcount assertion at the mutation
site (GET DIAGNOSTICS = 1). guardian's hashtext-collision note: harmless (serialisation
only); tooling_provenance's Go-side validDocSubjectTypes point: the doc_plans row was written
by SQL and the DB CHECK passed live — the Go list gates Go writers only. **D4 stage A is
DONE: live, hardened, approved, inert. Stage B (Go claim-step refusal, opt-in) is next, and
the owner's monthly budget number is still the open input (August measured ~$2,113,
outage-suppressed).**

## 2026-09-02 — the 314 lane's lint heads-up surfaced SEVEN unrecorded hand-applies; all --record-only'd

The bugs_open/314 session messaged: mid-caps migration names (our _sibling_A/B/C_, _lever_B_)
are invisible to pattern-check.py's idempotency lint (their fix, their lane; our naming stays
— it is a deliberate ordered-set convention, and now immovable because the ledger keys on
filename). Chasing their "applied out of band and left unrecorded" concern against OUR files:
**582/583/584/637/671/672/673 — all hand-applied, all runner-appliable by name, NONE in
schema_migrations** (0 rows of 481). The probe would have graced most (guards RAISE with
'already') but NOT 672 post-673: its drift arm's message lacks 'already' and would read as
'drifted, investigate' on a plain replay of the runner. All seven recorded via
`run-migrations.sh --record-only` with dated verification notes `[MEASURED 2026-09-02:
ledger now returns all 7]`. **Lane practice from here: a hand-apply of a runner-appliable
file is FINISHED only when --record-only has run** — the apply and the record are two halves
of one act (same shape as commit-then-build). _HOLD files stay outside the ledger by design
(never runner-probed; suffix never dropped). Answered 314 in full; offered them the
appliable-but-unrecorded census as a possible sibling check, their call.

### 2026-09-02 — my _HOLD answer to 314 was WRONG; corrected, and 658 completed per the documented sequence

I told 314 "_HOLD is never renamed, ever" from this lane's live behaviour — then the
migration-runner-practice memory (bugs_closed/150, 2026-08-01) showed the DOCUMENTED
sequence is apply → DROP THE SUFFIX → --record-only (the runner refuses to record a
sidecar on purpose; the rename is a required step). The lane had left 657/658 half-done at
the hold stage and I described the drift as the norm. What caught it: reading the topic
file before merging a "new" lesson into it. Corrected to 314 in writing; **658 renamed to
`658_dispatch_phase3_batch_8.sql` + recorded (commit df0d718dd, both paths on the commit,
ls-tree clean at HEAD)**; 657 flagged to its session (their file, their timing). The
rename commit carries no council trailer deliberately: content 100% identical to the
95099f95-approved file — the reviewed change is unchanged, only the name moved.

Addendum, same hour: 314's fleet census landed as my correction crossed it in flight —
**26 renamed _HOLDs fleet-wide (26/26 stuck on disk), none this lane's** `[THEIR MEASUREMENT
2026-09-02]` — so both readings were true of their samples and wrong as generalisations;
the tree carries both practices. They file the "dry run exists/free/mandated/UNDRIVEN" item
(their territory; our seven-unrecorded is its live evidence, the 672 vocabulary-degrade
detail quoted into it, the one-doc_notes-row-per-run cron shape offered as the candidate
driver). 658/657 close the orphan-rename path: 658 recorded under its new name only;
657's note to the 413 session says rename+record as one motion.

### 2026-09-02 — 413 CLOSED by its session; 426 filed by 314's; the arc's SUMMARY cut

The 413 session closed the bug (their 8a26d04e5: moved to bugs_closed with our +2h/+3.3d
PASS verdicts cited, caveats carried, residuals routed — candidate 2 stays an owner option
in README, 415 stays open) and completed 657's _HOLD lifecycle per the flag. The 314 session
filed **bugs_open/426** ("the applied-by-hand-never-recorded check exists, is free, is
mandated, and nobody runs it" — our seven-unrecorded as live evidence, the 672
vocabulary-degrade detail quoted, the cron shape as candidate 1). On their _HOLD-lint
question I conceded in writing: my "rename is the jurisdiction moment" conflated the
RUNNER's jurisdiction with the LINT's — write time is the only useful lint moment for a
file that is guaranteed to be hand-applied. **SUMMARY_2026-09-02 cut** (second in the
series): the ruling-B arc closed, D4 open. Lane state: awaiting the owner's budget number +
stage B go-ahead (both asked, 08-31/09-01 chat).

## 2026-09-02 (later) — OWNER'S PROVISIONAL RULING on job ordering; routing asserted; decisions restated to the owner

**Owner (chat, verbatim): "My decision would like to be, even if I can't make it yet, that
there is no need to reorder because everything flows through well and we can scale to meet
excess demand."** Marked PROVISIONAL at their own word ("even if I can't make it yet").
Scope as applied by this lane:
- **Candidate 2 (age floor / per-item wait ceiling): DECLINED provisionally.** Accepts
  unbounded per-item waits during sustained same-site inflow (the 08-31 ~16h pinned-row
  shape); in practice bounded by the queue now draining to empty. Mechanical revisit
  trigger: the pin census's oldest-row age tail growing across reads.
- Default for any further reordering proposal: decline unless it closes a MEASURED defect.
- **Does NOT touch D2** (clients-first in bursts — explicit 08-21 ruling, stands, unbuilt)
  and does not touch 415 (admission-correctness, not ordering).
Routing asserted to the 413 session: ordering/throughput decisions come through this lane
(the 08-26 levers/meters split), owner relieved of direct asks; their precise list requested.
Owner also asked for the open decisions restated fully — done in chat (budget number; stage-B
go-ahead; candidate 2 now provisionally answered); firm-up of the provisional ruling remains
theirs, no urgency while the queue drains clean.

### 2026-09-02 — 415's fix (migration 688) pinged at commit+apply per contract; no lockstep owed

The 415 session widens the fire gate on BOTH trigger rows (superset admission: triaged+
approved, pipeline filter dropped, lock-exception arm; new md5 2ebd918b, rollback restores
200246f7; council corr 5f0cb450 pending). Confirmed from this side: VERIFY 1/7 pins PARITY
not text and both rows move in one statement → stays green; 688 is scheduled_tasks-only so
stage B's agent_definitions/governor anchors are untouched. Watch beside the next routine
read: fire cadence p50 (~60s) and no-op turn share — a spare fire is a cheap tick, a TREND
is a report to them.

## 2026-09-02 (evening) — D4 STAGE B BUILT END-TO-END (Go committed, config held); council r1 in flight

Owner's "carry on" taken as the stage-B go-ahead (ships off — stated to them, no objection).
**The design CORRECTION found at the code before building** (PLAN §stage-B correction,
3c1b81d91): claim-only enforcement would re-create the 413 selection hog under shed —
the filter lives at SELECTOR + LOADER (where eligibility is defined) with the claim as the
race backstop. **Go half committed `dec5ad61b` (+gofmt `c0a18f37d`)**: renderer
workItemNotGovernorShedSQL with three mutation-tested posture rules (fail-open
NOT COALESCE(...,false); unmapped type = maintenance+bearing, sheds earliest; enabled=false
= identity), loader flag honour_spend_governor (honour_site_lock's exact shape, helper
extracted so byte-identical-off is a TESTED branch), claim backstop with its own
spend_governor_shed reason. Both named mutations executed and killed by the intended
assertions `[MEASURED 2026-09-02]`; one TRANSIENT peer half-save broke the package mid-proof
(all three runs "[build failed]" incl. post-revert — the shared-tree reality), detected via
numstat + rebuild, proofs re-run clean. A lucky structural fact, checked not assumed: the
predicate binds only wi.item_type (no $1), so the IDENTICAL spelling embeds in the selector —
unlike siteLockExceptionSQL's ⚠⚠ non-transplant. **Config half = 674_HOLD + ROLLBACK,
committed after a chained dry-run** (apply→rollback in one rolled-back txn: clause lands
once, selector still EXECUTEs, flags bare-true, governor stays disabled, restore md5-exact
d29807313, nothing persisted; snapshot_agent signature is (type, reason) — first dry-run
caught my (id, reason) guess). Council corr `8f4bb57d` r1 pending (verdict monitor armed).
**Sequence to live**: verdict → chassis image with dec5ad61b rides the owner's next release
(releases are WHOLE-FLEET, owner runs them) → stamp-check both replicas → hand-apply 674 →
drop suffix + record → owner sets budget → owner flips enabled (the one deliberate act).

### 2026-09-02 — baseline caveat from the 357 lane (bugs_open/408 fixed in code, rides the next roll)

Their FYI: the assemble_page stack-overflow class (skipped content writer / typo'd
content_field → pod crash + orchestration wedged EXECUTING_STEP until the 4h stale reaper;
three destroyed 08-26) becomes a clean page-skip after the next chassis roll. Two
consequences for THIS lane's measurements, held for the next reads:
- **Pre-roll turn-time baselines overstate tail latency** (they contain 4h wedges); do not
  grade post-roll p90/max against them without this beside.
- The next roll carries BOTH their fix and stage B's Go half — post-roll distribution
  changes are not attributable to either alone (stage B stays inert, but the tail change
  is theirs). Expect the daily VERIFY's zombie-tail NOTICE count to drop too (their wedge
  class feeds the 'reaper: stale EXECUTING_STEP' spelling the 08-27 widening covered).

### 2026-09-02 (later) — stage B r1 REVISE: the council made the design BETTER; r2 in flight

r1 verdict (corr 8f4bb57d): REVISE, gating = bug_historian (selector/loader shedding is
INVISIBLE — a withheld item indistinguishable from a stuck one), with the architecture seat
naming the right fix for a different objection: no canonical source for a predicate
hand-copied four times. **Both built rather than argued** `[all MEASURED 2026-09-02]`:
- **Migration 675** (applied + recorded same hour): `governor_admits(item_type)` — the ONE
  canonical predicate; Go and the selector now emit a one-line call; the posture rules moved
  from string assertions to EXECUTION probes (every shed level driven against synthetic
  states inside the verify, incl. DELETE-the-state-row fail-open). Plus
  **`governor_withheld_now`** — the shed-event view (withheld vs stuck, one query).
- Go collapse committed `6a84e3dc1`; 674 rewritten to the call + function-exists preflight;
  chained dry-run green post-rewrite; 675 pair committed `c2a95fc90`.
- Evidence answers: 671 CREATES the map table (r1's editquality couldn't see schemas);
  674's verify already EXECUTE-probes the full query; consumer censuses — load_work_items =
  {build-dispatch-loop, site-work-orchestrator}, claim_work_item = {build-dispatch-loop,
  **diagnose-dispatch-loop** (new datum)}, fleet flag census 0; the 413-hog argument
  RE-DERIVED from the live selector's quoted clauses (prior_art's fair procedural point —
  same lesson as the stage-A round: derive, don't retell).
r2 resubmitted on the same corr; verdict monitor to be re-armed. The travelling doc_notes
decision record written (tooling_provenance).

### 2026-09-02 (r2→r3) — stage B r2 REVISE dispositioned; the big objection was a SCOPE MISREADING, stated precisely rather than re-argued

r2 gating (debug_historian HIGH, replace() silent no-op): the shipped 674 ALREADY carried the
demanded checks (md5-exact preflight + post-text occurrence count + busy-skip survival before
the EXECUTE probe) — the r2 sketch under-showed them; an explicit pre-replace anchor-count arm
added anyway (aeab60fcb, commented redundant-by-construction), chained dry-run re-run green.
The three-seat "conceded monopoly, no mitigation" cluster mis-scoped the r2 derivation: it
condemned the REJECTED claim-only design; the built 674 excludes withheld rows INSIDE elig,
BEFORE ranking — a fully-shed site is invisible, not monopolist. True residual now a TRACKED
RISK: the claim-step race costs ≤1 loop turn per level flip (120s-bounded), self-healing.
New evidence attached to r3 `[all MEASURED 2026-09-02]`: EXPLAIN ANALYZE — 101 candidate rows
through governor_admits() in 55ms incl. sort (map PK confirmed); dated zero-flag census;
doc_notes decision row id 1032c8f4; ai_endpoint_unavailable at claim_work_item_action.go:290
(availability) vs spend_governor_shed (policy) — distinct reasons, same seam, compose not
compete; fuel.go's stage-A ruling re-cited (cross-corr invisibility strikes again — seats
cannot see 80df0963 from 8f4bb57d). r3 submitted; monitor re-armed.

### 2026-09-02 ~14:4xZ — STAGE B APPROVED (r3, corr 8f4bb57d); advisories acted on; D4 build COMPLETE pending roll + owner switches

**APPROVED — "3 advisory objections, none high-severity."** Verdict read in full. Dispositions:
- **guardian (REAL CATCH, acted on):** applying 674 changes the selector md5 that
  `657_..._VERIFY.sql` pins — without a lockstep edit the daily contract check would read as
  selector DRIFT. Now step 3 of 674's written apply procedure (update the VERIFY's md5 list,
  run green, commit together).
- **editquality:** the LIMIT-0 probe proves parse, not runtime behaviour — true; the selector
  text is parameter-free (verified at the artefact) and the Go halves carry their own tests;
  a first-10-fires canary is now step 4 of the apply procedure.
- **debug_historian:** the snapshot_agent calls they asked for have been in 674 since r1
  (lines ~57-58) — THIRD instance this corr of the sketch under-showing the file. Lane lesson,
  now twice-earned: **paste the verify/guard text into the sketch verbatim; summaries of
  guards read as absences.**
- **architecture (low):** second-consumer opt-in promoted from risk-note to STANDING GATE
  (674 header + AGOV-013): any further `honour_spend_governor` flag is architecture-scope.
**D4 state: stages A and B both council-APPROVED. Everything is built.** Remaining, in order:
owner's next release carries dec5ad61b → stamp-check both replicas → apply 674 (procedure in
its header, incl. the 657-VERIFY lockstep + canary) → drop suffix + record → owner sets the
budget (recommendation $4,500 given) → owner flips enabled. Three commits this arc carry
`Council-Submitted:`; this NOTES commit carries the earned `Council-Reviewed:`.

### 2026-09-02 — 384's ask answered: the floor meter has a SECOND blind spot (zero-eligible sites), and the detected→triaged promoter is firing-but-strict fleet-wide

The 384 lane asked three dispatch questions (leopardess blog listing never rebuilds — no
build-dispatch-loop in 24h). Answers, all `[MEASURED 2026-09-02]`:
1. **Promotion = `detected-item-promoter`** (scheduled_tasks, 900s, fire_message=f) —
   ENABLED and ticking (18:18Z today), INDEPENDENT of site selection: their
   selection-first-deadlock hypothesis refuted. But it promotes through deliberate DOORS
   (pipeline IN build/content/design; handler registered+active; a known-good door — the
   444/430/454 door-closers), and **30 sites hold detected>0 with ZERO triaged/approved**
   (webdesign.co.uk 158 since 08-04, finetuning 77 since 07-26, leopardess 51). Whether the
   doors are correctly parking low-value work or over-holding is OPEN — reading the full
   pre_query + the three cited bugs decides it; flagged to 384, either lane may file after
   that read. ⚠ my own 08-30 "backlog fully drained" partly re-reads under this: empty
   triaged ≠ no demand — some demand was parked at 'detected'.
2. `sites.build_status` is NOT read by the selector (the verified clause list) — inert for
   dispatch.
3. **Their instinct confirmed: the floor is BLIND to zero-eligible sites** (its WHERE
   requires eligible rows). RUNBOOK gains the zero-eligible census beside the floor, with
   the promoter-state read attached (a big census = a promotion question, not a selector one).
Their born-terminal find (insertWorkItem two-strike arm counting another producer's
successes) is 389/CONTRIB territory — noted, not this lane's.

### 2026-09-02 (evening) — the promoter question RESOLVED (no bug: the flag layer is designed); my own census meter corrected TWICE same-day; the sawtooth probe run

The 384 lane's two corrections landed within the hour and both were right:
1. **Zero-eligible-at-an-instant is steady state** (their every-site control) — my hour-old
   RUNBOOK census read ~the whole estate and discriminated nothing. The 413 file's own
   rolling-window warning, walked into by ME an hour after citing it to them.
2. Their handler-door finding (1,386 detected rows, 100% empty handler) then resolved
   AGAINST both our readings by the promoter's FULL text `[READ 2026-09-02]`: rows without
   handlers are EXCLUDED from scoring BY DESIGN — "Flag-only rows … 'detected' is where they
   belong permanently". The flag layer is records, not parked work; the doors govern only
   handler-bearing rows; **no promoter bug exists**. RUNBOOK section rewritten (visible
   double-correction) to the meaningful meter: handler-BEARING detected age, with the
   promoter's own held_detail as the why. Lesson, twice in one day: **read the mechanism's
   FULL text before metering its population — the first screen of a pre_query is not the
   design.**
3. **Sawtooth probe** `[MEASURED 2026-09-02, 21d live+archive]`: leopardess's rerender-family
   completions froze at 08-28 (179→5→4→2→1) and — the discriminating signal — UNIQUELY
   failed to recover post-outage (09-01/02: dartsonline 268/106, gaswholesalers 52/139,
   leopardess 2/1). Supports 384's two-strike freeze at the site level; the weekly
   PERIODICITY is untestable in this window (the 08-28–31 trough is the fleet LLM outage,
   confounding every site); their 09-02→04 self-resume prediction remains the clean test.
4. Their `unresolved`-is-terminal overcount trap adopted as a RUNBOOK caveat on the
   work_items_open baseline figure.

## 2026-09-02 ~18:26Z — OWNER SET THE BUDGET: $2,000/month; the meter is ARMED (enforcement still off)

**Owner (chat, verbatim): "my budget is $2000 per month."** Set at 18:26Z
(`UPDATE governor_config SET monthly_budget_usd = 2000`), read back: enabled still FALSE,
thresholds 70/85/95 → **L1 $1,400 · L2 $1,700 · L3 $1,900**. September MTD at set time:
$241.92, level 0.

**Stated to the owner plainly (this is a THROTTLE, not a guard-rail):** the measured
post-recovery burn is ~$99/day ≈ $3,000/natural month, so at $2,000 the governor — once
enabled — will actively constrain: at current burn L1 (maintenance sheds) lands ~day 14,
L2 (builds) ~day 17, L3 (research) ~day 19 of each month. That is a legitimate owner
choice (hold spend to $2k) and the system will do exactly that; flagged so it is chosen,
not discovered. Their console cap must sit ABOVE $2,000 (~$2,200–2,500+) or the account
wall still fires first. One line changes the number at any time; September's metering
(running regardless of enablement) will show the real crossing dates.
Remaining to live enforcement: the release (dec5ad61b in the chassis) → 674 apply (its
header procedure incl. the 657-VERIFY lockstep + canary) → owner's "enable".

## 2026-09-02 ~21:1xZ — owner reports a fresh chassis DEPLOYED; stamp check BLOCKED by the kubeconfig 3-day expiry (21:08Z, three minutes before the check); session handed off

Owner (chat ~21:00Z): "A fresh chassis build has been deployed." The 674 sequence's step 1
(merge-base `dec5ad61b` against BOTH replicas' provenance stamps) hit `Unauthorized` —
token decoded: **expired 21:08:03Z**, dead on the 3-day cycle, minutes after the 18:31Z
governor-tick read worked. So the roll's content is **UNVERIFIED** — recorded as such, not
assumed (a roll is not evidence the code shipped). **HANDOFF_2026-09-02_continue_here.md
cut** (supersedes 08-30): full 674 procedure incl. the 657-VERIFY lockstep + canary +
suffix-drop+record, the enable sequence with the console-cap dependency, the reference
card, and the day's traps (verbatim-sketch, uniform-split, flag-layer, unresolved-terminal,
record-on-apply). Governor state at close: budget $2,000 armed, level 0, $244 MTD,
heartbeat live, enforcement three switches from real (verified roll → 674 → enable).

## 2026-09-02 ~21:4xZ — TOKEN REFRESHED; THE WHOLE 674 SEQUENCE EXECUTED CLEAN; D4 is ONE OWNER ACT FROM LIVE

Owner refreshed the kubeconfig; the handoff's NEXT ran end-to-end `[all MEASURED 2026-09-02]`:
1. **Stamp check → capability probes** (the provenance grep matched an in-log QUOTATION of
   docs about the provenance check — prose-poisons-detector, live; fell back to the binary
   probe per the memory): both replicas (`8ddbf8958-cd2h9`/`-vppjz`, started 20:56/20:57Z)
   carry `governor_admits(`/`spend_governor_shed`/`honour_spend_governor`, absent-control 0
   on both. ⚠ the absent-control TIMED OUT the first combined run (cannot stop early — the
   396 lane's exact experience); re-run alone, clean.
2. **674 APPLIED 21:27:14Z** — refusal arms passed, NOTICE green, governor still disabled.
   New selector md5 **fcbe8821a2a56512911955735796460e**.
3. **657-VERIFY lockstep same sitting** (r3 guardian): md5 constant → fcbe8821 with a
   comment recording WHY touching it was sanctioned; run green (K=8 live, 6 eligible/0
   pinned, next pick gaswholesalers). Commit 50c47efd5.
4. **Canary clean** (21:27→21:37): 11 fires (~60s cadence), 11 loops/8 sites, 29 claims won,
   **0 spend_governor_shed refusals, 0 withheld rows** — disabled = byte-identical, proven
   on live traffic.
5. **Suffix dropped + ledger-recorded in one motion** (6385e6d00; both paths on the commit;
   ls-tree one appliable 674… which revealed a NUMBER COLLISION: another lane's
   `674_farmer_cull_spec_wash_three_prose_aspects.sql` — numbering-not-a-mutex; resolve by
   SLUG, as with bug numbers).
**Remaining for D4, both the owner's:** console cap above $2,000, then
`UPDATE governor_config SET enabled=true WHERE id=1;` — after which watch the first shed
cycle and unlock option C.

## 2026-09-03 10:14:32Z — OWNER SAID "enable" — THE GOVERNOR IS LIVE

**Owner (chat, verbatim): "enable".** Executed 10:14:32Z; read back in the same transaction:
enabled=t, budget $2,000, thresholds 70/85/95. State at enable: **level 0, $373.42 MTD**
(day 3 — running ~$124/day, hotter than the $99 estimate: at this pace L1 ~Sep 11-12,
L2 ~Sep 14, L3 ~Sep 16). Immediate verification: `governor_admits` TRUE for all four probe
classes (maintenance/build/llm-free/unmapped) at level 0; `governor_withheld_now` = 0 rows.
Post-enable live-dispatch watch appended below. **D4's build→live arc is COMPLETE**: four
measured blackouts → owner-ruled policy → three tables of council rounds → a governor that
now watches every claim with the owner's numbers in it. Console-cap-above-$2,000 reminder
issued a third time (owner's console, not verifiable from here). **Option C's gate is now
"one real or induced shed observed"** — at the current burn the first REAL shed arrives
~Sep 11; an INDUCED one (briefly lower the budget in a controlled window, watch L1 fire,
withheld view populate, level-change doc_note write, restore) is the faster path and a
next-session choice.

### 2026-09-03 ~10:40–10:50Z — the first post-enable watch: the governor is inert-as-designed at L0, and the shed staircase is PROVEN at all four levels without withholding any live work

Picked the lane up ~25 min after the enable. Everything below `[all MEASURED 2026-09-03]`.

**1. Post-enable watch — dispatch is unchanged, which is the whole claim at level 0.**
Adjacent 25-minute windows either side of the 10:14:32Z enable: build-pipeline-trigger fires
6 vs 6; dispatch loops 6 (2 sites) vs 7 (7 sites), 0 FAILED either side; **0
`spend_governor_shed` mentions in any orchestration since the enable; `governor_withheld_now`
= 0 rows;** `governor_admits` TRUE for all four class/bearing groups and for an unmapped
type. ⚠ handler counts across the two windows (48 pre vs 19 post) are NOT comparable — the
post window is right-truncated, its loops were still spawning. Recorded so nobody grades a
throughput change off it.

**2. The heartbeat scare that was not one.** First read: `computed_at` age **211 s** against
a 120 s task. Second read three minutes later: **25 s**. Interval 120 s + the scheduler's
30 s tick makes ~150 s ordinary and a single tail read meaningless; the RUNBOOK now says two
consecutive reads over ~300 s is the signal. Not "fine" by assumption — re-read, twice.

**3. MY OWN WRONG TURN, caught before it was asserted anywhere durable: I read the wrong
object and nearly filed a revert.** Checked "does the selector carry the governor clause?"
against `scheduled_tasks.pre_query` — it came back **f**, with an md5 matching neither the
pre-674 nor the post-674 value. That reads exactly like "another session reverted 674".
It is not: the selector 674 edits is
`agent_definitions.default_config#>>'{workflow,steps,find_dispatchable_site,config,query}'`,
and `scheduled_tasks.pre_query` is a different query that was never in 674's scope. **The
word "selector" names two live objects in this lane** — 657's VERIFY pins the agent_definitions
one; the trigger's `pre_query` is the wake-up gate (the 415/688 lane's). Read at the right
object: md5 `fcbe8821a2a56512911955735796460e`, carries_gov **t**, both step flags bare
jsonb `true`, fleet negative control 0 other rows. Wiring intact.

**4. What that wrong turn turned up anyway — a release rewrites EVERY live agent row ~70 s
before the pods start.** 208 rows / 203 types, all stamped `2026-09-03 08:56:53.045885+00`
(one statement, one microsecond, snapshots untouched), with **no matching
`schema_migrations` row** — so it is the release's own seeding step, not a migration. The
chassis pods started 08:57:46 / 08:58:07Z on `v1.0.1356`. **The hand-applied governor clause
SURVIVED it** (md5 read AFTER the roll). But 674 edited the LIVE row and no repo seed carries
the clause, so the window is real and a silent revert would remove the governor's primary
gate with nothing reporting it. RUNBOOK gains "re-run the wiring check after EVERY release".
657's VERIFY pins the same md5 and would also catch it — but it is a hand-run habit, not a cron.

**5. The enable landed on pods that had never been probed.** Yesterday's capability probe was
on `8ddbf8958-cd2h9`/`-vppjz`; the 08:57Z roll replaced them. Re-probed both live pods:
`governor_admits(` / `spend_governor_shed` / `honour_spend_governor` = 1 on each, absent
control `governor_forbids(` = 0 on each (run separately — the combined form times the exec
out, as on 09-02). The Go halves are live on `v1.0.1356`.

**6. THE SHED STAIRCASE, PROVEN — and the meter that could not have come out otherwise.**
Drove the LIVE selector against synthetic `shed_level` 0→3 inside one transaction, rolled
back (MVCC keeps it invisible to live dispatch; post-rollback control clean each time).
First attempt reported `selector_candidate_sites = 1` at **every** level — which reads as
"the governor changes nothing" and is in fact **the selector's trailing `LIMIT 1`**: the
meter could not have produced any other number at any level, true or false. Stripped the
trailing LIMIT (with an abort arm if the strip does not match) and re-ran:

| level | dispatchable sites | withheld items | by class |
|---|---|---|---|
| L0 | 14 | 0 | — |
| L1 | 13 | 51 | maintenance/llm 51 |
| L2 | 13 | 112 | + build/llm 61 |
| L3 | 13 | 112 | (no research-class item eligible in the window) |

Exactly the owner's ruled order, and the class ladder flips in the right sequence. **Sites
barely move while items move a lot** — correct, because shedding is per `item_type`: a mixed
site stays dispatchable on its llm-free work. That is the council r2/r3 "withheld, not
monopolist" property, now measured rather than argued. L3 showing no further withholding is
a population fact (one research type in the map, nothing of it eligible at 10:47Z), not a
defect — say so when quoting it.

**What none of this proves:** the Go loader and claim backstop reading a NON-ZERO level on
live traffic. The DB predicate, the selector clause and the view are all proven; the Go
halves are proven present but only ever exercised at L0, where they are identity. Option C's
gate ("one real or induced shed observed") therefore still stands.

**Owner, in chat ~10:41Z, verbatim: "I have increased the cap in anthropic to $3000."** That
closes the console-cap dependency that had been flagged three times: the account wall now
sits $1,000 above the budget and $1,100 above L3, so the governor's staged brake always
arrives first. Every precondition D4 was waiting on is now met.

### 2026-09-03 11:14–11:34Z — THE INDUCED SHED (owner-authorised): the whole chain fires, and the alarm that announces it does not

Owner, in chat: **"induce it today"**. Window run and restored; everything below
`[all MEASURED 2026-09-03]`, log at `scratchpad/induced_shed.log`, markers inline.

**Design, and why NOT L1.** The demand control decided it: claims in the hour before, by class
— maintenance/**llm-free** 47 (page_rerender ×44), maintenance/llm 8, build/llm 6, unmapped/llm
3. At L1 only ~11 claims/hour would be silenced (~2 in a short window), so a zero would have
been indistinguishable from ordinary quiet — an absence argument with no power. **L2 sheds L1's
class plus builds (~17/hour, 27% of claims), and the llm-free `page_rerender` stream continuing
throughout is the positive control that dispatch is alive rather than stalled.** Budget dropped
to make MTD cross 85%.

**⚠ My own sizing was stale within 20 minutes.** I computed budget 430 against MTD $388,
measured ~20 min earlier; at window start MTD was **$398.61** — 92.7% of 430, close enough to
the 95% L3 line to drift across mid-window. Nudged to 450 (88.6%) before the level moved. The
burn between those two reads was ~$35/hour, far above the $124/day (~$5/hour) figure the daily
average implies — **MTD is lumpy, so size a threshold experiment against a read taken minutes
before, not the day's average.**

**Timings — the governor is about TWICE as slow as its stated 120 s, in both directions.**
Onset: budget set 11:14:33 → level 2 at 11:17:09 (**156 s**, caught mid-cycle). Release: budget
restored 11:29:25 → level 0 at 11:33:34 (**249 s**, a full cycle). Heartbeat ages across the
poll log climb 129→151→173→196→217→239 then reset: the state task's real cadence is **~250 s**,
not 120 s — interval 120 + the scheduler's 30 s tick, running ~2× under load. So a budget
crossing takes up to ~4 minutes to bite AND up to ~4 minutes to release. **The release lag is
the one nobody would predict**: the budget was correct again at 11:29:25 while 115 items stayed
withheld for a further 4 minutes.

**What fired, all of it:**
- **Withheld view**: 114–115 rows for the whole 12m16s, classed `build/llm 60–61` +
  `maintenance/llm 54`. Correct ladder, correct classes, stable.
- **The loader half — the discriminating measurement.** Per-loop census over the shed window:
  3 dispatch loops handled **24 items, every one of them llm-FREE maintenance**, while 100+
  llm-bearing items sat eligible and withheld. Zero llm-bearing items loaded.
- **The Go claim backstop fired**: **1** `spend_governor_shed` refusal in the window — the
  load→claim race the backstop exists for, observed live rather than argued.
- **Dispatch never stalled**: 24 items handled in the shed window vs 13 in the control window
  before it. The fleet worked *harder* on llm-free work while the llm-bearing half was held.
- Restore clean: level 0, withheld 0, budget 2000, enabled true.

**⚠ The one llm-bearing claim inside the window, and what it means.** A `needs_diagnosis` item
was claimed at 11:27:09 — during L2, by `diagnose-dispatch-loop`. Not a defect: **only
`build-dispatch-loop` carries `honour_spend_governor`.** Live census: `build-dispatch-loop` t,
`diagnose-dispatch-loop` f, `report-dispatch-loop` f, `zip-deliverable-dispatch` f. So **the
governor governs one of four dispatch loops** — by design (opt-in default off, and a second
consumer is the STANDING ARCHITECTURE GATE), but it must be said plainly rather than left for
someone to discover at L3. By count it is small (3 of 268 claims over 3 hours, ~1%); **by SPEND
it is unmeasured and plausibly not small**, since `needs_diagnosis` drives whole diagnosis runs.
That is an open question for D4's next round, not a claim.

**THE FIND — the alarm is dead: `bugs_open/459`.** Two real level changes, **zero** `doc_notes`
rows. Root cause reproduced by A/B on the live stored text inside `BEGIN … ROLLBACK`: with
`FOR UPDATE` on the `old` CTE (migration **673**) the note's `INSERT … FROM old, new` selects
nothing; delete that one token and `level_changed` goes 0 → 1 and the row lands. A control arm
(the `old` CTE alone) returns `shed_level=3` against `new.lvl=0`, ruling out "the inputs did not
differ". **672 installed the alarm AND proved it** (verify drives 0→3→0, asserts exactly 2
notes); **673 hardened the statement and its verify checked only that its token was present and
that the text still ran** — blind by construction to the sole behaviour that token could break.
Pattern filed to 016b §9. Fix candidates ordered in the bug file; an appliable migration is
council scope, so the fix is NOT hand-rolled here.

**Honest limit on today's evidence.** The before/shed/after claim comparison is weak on its own:
llm-bearing claims run ~2 per 12 minutes, so the counts (2 before, 1 in-window from the
ungoverned loop, 0 after by 11:39) cannot carry the argument. **The per-loop LOAD census is the
measurement that discriminates**, because every loop in the window is a trial and the withheld
pool proves demand existed throughout. Quote that one.

**Cross-lane datum received** (the `bugs_open/329` lane, unprompted, on our own meter): a fresh
24 h double-handle census — **3,044 handlers, 2,911 distinct items, 71 with ≥2 handlers, 0
overlapping pairs**. The 71 are sequential retries; no stale-reap shapes needed the
discriminator. Independent re-run of the RUNBOOK census by a lane with no stake in its result.
