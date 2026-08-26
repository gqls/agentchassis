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
