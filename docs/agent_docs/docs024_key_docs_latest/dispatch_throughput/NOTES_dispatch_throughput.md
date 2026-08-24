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
