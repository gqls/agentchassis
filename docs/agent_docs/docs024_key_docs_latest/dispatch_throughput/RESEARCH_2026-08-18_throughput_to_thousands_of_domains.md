# RESEARCH — system throughput to thousands of domains (the whole-architecture scale review)

**Written 2026-08-19** by the dispatch_throughput workstream at the owner's direct request
("research deeply the options for increasing the throughput of the system … do we need more
than one repo or deploy path … should we add another cluster … several thousands of domains
… when I promote it I will have many per day as a sort of burst traffic"). This executes —
brought forward — the "whole-architecture scale review" the site_delivery lane seeded for
after the working site. Evidence markers: `[MEASURED <date>]` with source, `[INFERRED]`,
`[UNMEASURED]`. Baseline queries: `RUNBOOK_dispatch_throughput.md`. Scope ruling (owner,
2026-08-18): research + diagnosis only; no live changes were made for this document.

---

## 0. The answer in one paragraph

The computers are not the constraint — the cluster runs at 6–20% CPU `[MEASURED 08-18]`
while the platform takes work one piece at a time in four separate places (dispatcher,
chassis worker pool, adapters, deploy runners). Nor is the repo: the pain in the platform
repo is session coordination, not size, and repo-per-site is already ruled out. Nor is a
second cluster: it multiplies capacity we are not using. What actually binds, in order of
when you will hit it: (1) the dispatch turn ceiling (~83 items/hour fleet-wide), (2) the
LLM account — spend cap hit twice in eleven days at 1/50th of target volume, with no
client-side governor anywhere, (3) the site-deploy fan-out (2 runner slots, one Actions run
per commit), and (4) at promotion time, the Cloudflare zone cap (~1k, weeks away at
50 domains/day). The burst answer is not peak provisioning; it is a priority lane for
paying builds, a spend governor, an honest ETA in the signup flow, and a drain rate the
account can afford.

## 1. The target, in three axes

**Axis 1 — the unlock sequence** (what binds *next*, mechanically):
dispatch turns → chassis workers (8 fleet-wide) → adapters (sequential by design) →
Kafka topic churn → LLM limits. Each is evidenced in §2.

**Axis 2 — destination constraints** (what binds at ~3,000 domains): LLM economics,
per-domain cadence, deploy path, CF zones, Postgres growth. These drive the Tier-2
investments regardless of the unlock order.

**Axis 3 — load shape.** Steady-state maintenance and promotion-burst onboarding are
different problems:

| scenario | rate needed | today's ceiling | gap |
|---|---|---|---|
| D0a low: 1 maintenance item/domain/day × 3k domains | ~125/hr | 83/hr | ~1.5× — config levers suffice |
| D0a high (owner 07-20: "hundreds of thousands of jobs/day") | 4,000–12,000/hr | 83/hr | ~50–150× — structural |
| D0b burst: 50 signups/day × ~213 items/build `[MEASURED 08-18, loanzy]` | ~10,650 items/day just for builds | ~2,000/day ceiling | builds alone exceed the fleet |

The 100× swing between D0a-low and D0a-high is why **D0 is the first decision**: nobody has
ever stated the intended per-domain cadence (chassis_replica_scaling PLAN §13 has awaited it
since 07-20).

## 2. The capacity chain, ranked (unlock order)

1. **Dispatch: one site at a time, ~83 items/hour fleet-wide** `[MEASURED 08-18, STARTER §2]`.
   `max_concurrent: 8` is dead config — `countInFlight` counts task *rows* per group (the
   dispatch group has one), and `loadDueTasks` separately enforces single-flight per row
   (`cmd/scheduler/main.go:414, :368, :181`). Independently verified by the 090 loop
   (run `a16b82cd`, verdict recorded in NOTES). Config-only remedy exists (PLAN Phases 2–3:
   sibling rows + batch; predicted ~1.8–3× at N=2, ceiling ~×12 at N=8×batch-8).
2. **Chassis worker pool: 8 concurrent orchestrations fleet-wide** (`CHASSIS_INTAKE_WORKERS`
   default 4 × 2 pods; the env var is set nowhere) `[CONFIG-READ 08-18]`. One env var + a
   replica bump — but gated on the never-taken `orchestration_states` write-amplification
   measurement (chassis lane's own 07-28 gate).
3. **Adapters: 6 of 8 sequential by design**, git-adapter effectively 1 consumer,
   browser-runner/analyser topics 1 partition. The "treadmill" 0/5 failure was measured at
   FIVE concurrent items (07-28): callee queue delay + response-lane delay > 3-min await,
   and each retry re-queues at the back. `[MEASURED 07-28, chassis lane]` One exception the
   other way: image-generator fans out **unbounded** (`go a.handleMessage`) — the burst-
   sensitive paid API has no limiter.
4. **Kafka per-job topics**: 2 created per spawn; 3,947 topics live `[MEASURED 08-18]`
   (3,022 `job.*`), down from 25,042 on 08-10 (bugs_open/240, OPEN — `MetadataTopics` still
   blank; which reaper is now holding the line is `[UNVERIFIED]`, and the only sweep known
   to work is a laptop crontab). Creation ∝ dispatch rate: ~1,400–2,000/hr at ×8.
5. **LLM account**: no client-side rate limiter, concurrency limiter, or budget anywhere in
   the tree; retry is a fixed ladder; `ai_endpoint_health` is the de-facto circuit breaker
   and never re-probes. Three provider outages in 11 days (07-31, 08-10 [3h20m fleet-wide],
   08-17; plus Gemini 429s, bugs_open/202). The 68% council caching saving dilutes with
   concurrency (each overlapping run pays its own cache write).
6. **Nodes/pods**: 5 spot nodes ("replaced without notice"), a K8s Job + 2 topics per spawned
   handler, no HPA/PDB/ResourceQuota. At D0a-high ≈ 1 Job/second sustained `[INFERRED]`.
7. **Postgres**: real headroom on rates (chassis PLAN §10) but one non-HA instance behind one
   non-HA pgbouncer (`default_pool_size=15` is the true ceiling), and 100Gi is the physical
   bound. Availability risk before throughput risk.
8. **Per-site Go actions** (features_open/024): 296 registry entries, 9 exist for 2 sites —
   "the only true scale blocker" by the estate's own record; couples site count to binary
   size. Remedy pattern exists (`CHVerticalProfile` config tables); owner already ruled
   "get it adopted".

## 3. The three questions, answered

**More repos?** No. There are already three (`agentchassis`, `sites`, `vm-sites`) and the
boundary is the decided one (deploy-key scoping, sink separation — traffic_probe REPORT).
Repo-per-site is owner-rejected verbatim ("several thousand domain names so individual
repos will be ungainly"). The platform repo's pain is *coordination* (390 commits/day
`[MEASURED 08-18]`, 4× the rate at which worktrees were deferred on 07-18) — that is
decision D11 (worktrees/branch model), not a repo split. Cohort-sharded site repos (10–30,
never per-site) only become worth tabling if D8 keeps the Actions deploy path.

**Another cluster?** Not for capacity — the one we have is idle; every binding constraint
in §2 is a serialisation choice or an external account, and none is relieved by a second
cluster. When a single cluster genuinely fills (D0a-high), the sanctioned shape is the
**satellite** (register BIZ-014): a self-contained cluster + Postgres + Kafka + scheduler
per shard of domains, sites partitioned by `site_id` — never a second cluster reaching into
uk_001's in-cluster singletons (`*.svc.cluster.local` DNS, exec'd Postgres, a scheduler
whose only mutual exclusion is a DB timestamp). The makefile already parameterises
ENVIRONMENT/REGION, so satellite wiring is cheap; the state layer is the design work.
Prerequisite either way: BIZ-014's five seams (owner_id, entitlement gate, billing adapter,
credential parameterisation, build-tier flag) — "cheap now, a forensic untangling later".

**A different deploy path?** Yes — this is the one that scales with domain count and is
already at its ceiling: one shared repo + branch serialising at GitHub's updateRef (4
retries), 2 in-cluster runner slots, one Actions run + `b2 sync` + full CF purge per
commit; 377 page rerenders once queued ~230 redundant whole-estate syncs with measured
out-of-order regressions. The fork is decision **D8**: platform writes B2 directly (git
stays as the audit/rebuild record — the platform already writes B2 in DGH-008/DGH-011) vs
staying with Actions and batching (one commit per site per dispatch turn) + more runners.
Interim regardless of D8: build the batching, it is a recorded-but-never-built lesson.

## 4. The onboarding burst path (new, from the owner's burst steer)

One measured case, end to end — **loanzy.uk, born 2026-08-18 12:53** `[MEASURED 08-19]`:

| dimension | value |
|---|---|
| work items filed day one | ~213 (37 needs_page, 35 needs_imagery, 32 page_rerender, …) |
| orchestrations attributable | **410**, of which **87 FAILED** (21% internal failure/retry) |
| LLM calls / tokens | ~180–200 calls; **~1.1M input + ~1.1M output** (component-creator is output-heavy) |
| LLM cost at Sonnet list | **≈ $20/build** (pre-caching, pre-Batch; imagery API `[UNMEASURED]`) |
| wall-clock submit → mostly built | **~10.5 h** (residuals next morning) — much of it queue wait, not work |

Caveats: n=1; a second case (remortgagecalculator.uk) could not be attributed — its
orchestrations carry neither domain nor site-id in payloads, so the query returns a
well-formed zero (recorded in NOTES; the attribution seam itself is a finding — nothing
today can answer "what did this site's build cost" without archaeology).

**Burst arithmetic** `[INFERRED from the one case]`: 50 signups/day ≈ 10,650 items,
~20,500 orchestrations, ~$1,000/day LLM (+imagery), ~100 Kafka-topic-creating spawns/hour,
~1,600 deploy commits — every number past today's whole-fleet daily totals. And under
today's strict fleet FIFO, signup #50's build queues behind 49 builds *and* all older
maintenance items — days, visible to a paying customer.

**The burst design is therefore** (decisions D0b/D2/D4/D15): a priority lane for paid
builds ahead of maintenance; an LLM spend governor so a burst cannot hit the monthly cap
mid-build (it would today — twice in 11 days at far lower volume); admission control
(intake cap + pause-maintenance-during-burst); and an honest ETA in the signup flow
(coordinates with site_delivery's request-phase build wait). Elasticity, not peak
provisioning.

**Zone timeline warning (D12):** at 50 domains/day the ~1k CF zone cap is ~3 weeks of
promotion away (39 zones today). The already-ruled plan B (own authoritative DNS +
CF-for-SaaS) takes time to build — its readiness may need to *precede* the first big
promotion, not follow the 500-domain milestone.

## 5. Tiered options

**Tier 0 — config only, days, reversible** (execution waits on D0/D1 per the owner's scope
ruling): sibling dispatch rows at N=2 (migrations A→B→C in the PLAN — the notify-stamp
hardcode is the trap); batch 5→8 with `timeout_seconds` 300→600 in lockstep (D3).
Predicted 1.8–3×; meters in §7. Do not exceed N=2 before the adapter decision — the
measured cliff is at 5 concurrent items.

**Tier 1 — small code, weeks**: LLM client concurrency limiter + spend governor + health
re-probe (promotion prerequisite); one commit per site per dispatch turn (deploy batching);
class-level item dedupe (606 `head_essentials_missing` rows are ONE defect class — the
cheapest "throughput" is fewer items); PDB/HPA/ResourceQuota + an on-demand node floor;
chassis workers raise (after the write-amplification measurement); `MetadataTopics` fix +
an in-cluster topic reaper that actually fires; shared response consumer group.

**Tier 2 — structural, decide direction before building**: D9 dispatch end-state
(worker-pull from `site_work_items` — the claims-CAS pattern the chassis already proved —
vs an executions-table scheduler; mutually exclusive); D8 deploy end-state; retire per-job
topics (MCL-013 shared pools); per-site actions → config (024 adoption); Anthropic Batch
API for latency-tolerant maintenance classes (~50%); Postgres growth plan (measure write
amplification first, archival ratchet, then managed/replica/partition-by-site_id);
BIZ-014's five seams.

**Tier 3 — before ~500–1,000 domains (or before the first big promotion, whichever is
sooner)**: DNS plan B; per-hostname metering (Workers Analytics Engine / TRF-006 beacon);
satellite #2 shape + trigger; CI + release delegation (today: one operator, no CI, 14
serial docker builds, fleet mixed for hours after each roll).

## 6. Per-domain economics — what the split actually is `[MEASURED 08-19, 7-day window]`

- **Council-gate dominates LLM spend**: 2,207 calls, 7.9M in + 6.4M out + **271.9M
  cache-read** tokens in 7 days ≈ $200/week at list. This is platform self-governance —
  it scales with *change volume* (commits, fixes), NOT with domain count. Don't divide it
  by domains.
- **Site builds**: ≈ $20 one-off per site (§4).
- **Steady-state maintenance is mostly LLM-free**: completions are dominated by
  `page_rerender` (4,808 of the last 14 days' items) which renders and deploys without
  model calls; the LLM-bearing maintenance classes (content_rewrite, section_edit,
  component work) are a small slice.
- So the earlier worst-case ($27–$160/domain/month) is what happens only if build-phase
  intensity is applied as steady state. The honest model prices three things separately:
  builds (one-off), LLM-bearing maintenance per class × D0a cadence, and council overhead
  (per change, amortised across the estate). **The per-class maintenance figure is the
  remaining `[UNMEASURED]`** — measure it over a clean week before pricing tiers.

## 7. Falsifiable predictions and their meters

| change | prediction | meter | disconfirming reading |
|---|---|---|---|
| sibling row N=2 | 1.8–2× completions/hr under sustained backlog | per-MINUTE `count(DISTINCT site_id)` of claims (never 5-min buckets) | still 1 site/minute ⇒ sibling never fired (check its own `last_triggered_at`/stamp) |
| batch 8 + timeout 600 | +~50% on top | completions/hr, 24h vs 24h at similar backlog depth | gain <half ⇒ something else binds (chassis workers/adapters) — stop, measure there |
| deploy batching | Actions runs ≈ sites touched, not items completed | runs/day vs commits/day | runs still ≈ commits ⇒ batching not on the path that commits |
| LLM governor | zero fleet-wide cap outages during a burst | `ai_endpoint_health` + llm failure rate | a cap outage with governor on ⇒ budget mis-set or a bypassing caller |
| any post-change zero | — | **demand control first**: `triaged+approved` depth in both windows | a quiet queue reads as a failed fix |

## 8. Decisions for the owner

| id | decision | default if unanswered |
|---|---|---|
| **D0a** | steady-state maintenance cadence per domain (items/day, which classes) — sizes everything | none — blocking |
| **D0b** | burst profile: peak signups/day to plan for + time-to-live-site promise | none — blocking for promotion |
| D1 | dispatch concurrency N + acceptable spend rate | stop at N=2 |
| D2 | service order: priority lane for paid builds over maintenance (+ aging constant) | stays FIFO — but under D0b that is customer-visible |
| D3 | batch/timeout lockstep | lockstep |
| D4 | LLM spend governor: budget + at-cap behaviour (pause / degrade model / Batch-only) | none exists — promotion prerequisite |
| D5 | second LLM provider for availability (243 c2); does stay-on-Sonnet (council-scoped) extend to maintenance? | single provider |
| D6 | Batch API adoption: which classes accept ≤24h latency for ~50% cost | not wired |
| D7 | Anthropic account tier / commitment (two exhaustions in 11 days at 1/50th of target) | current tier |
| D8 | deploy end-state: platform-writes-B2 + git-as-record vs Actions-per-commit (+ interim batching, runner count) | Actions, 2 slots |
| D9 | dispatch end-state: worker-pull/event-driven vs executions-table scheduler | neither built; sibling rows are the stopgap either way |
| D10 | CI + who may run releases | owner-only, serial, no CI |
| D11 | session worktree/branch model (deferred 07-18 at ¼ of today's commit rate) | shared tree |
| D12 | DNS plan B trigger — calendar-based if promotion is planned (zone cap ≈ 3 weeks at 50/day) | "near 500–1,000 domains" |
| D13 | satellite model: domains/satellite, second-satellite trigger; build the five seams now | seams unbuilt |
| D14 | node economics: on-demand floor, HPA | 5 spot nodes, no HPA |
| D15 | admission control: backlog SLO / detector throttling / pause-maintenance-in-burst | detectors file freely |
| D16 | retention/archival for orchestration_states + llm_call_log (100Gi bound) | unbounded growth |

## 9. Closed options — ruled out, with the ruling

- Repo-per-site (owner, traffic_probe REPORT: "ungainly").
- Class C dynamic rendering on boxes (same report: "the VM is a second sink, not a second renderer").
- Kafka-native parallelism via ~125 partitions (chassis PLAN §9: "wrong tool shape"; partition
  count is a one-way door whose key was never decided).
- Thousands of clusters / cluster-per-domain (BIZ-014).
- Mixed council model roster (owner 08-10: caches are model-scoped; mixed is *more* expensive).
- Coupled multi-cluster (shared Kafka/Postgres across clusters) as an isolation mechanism
  (concept register: "it's a coupling mechanism").
