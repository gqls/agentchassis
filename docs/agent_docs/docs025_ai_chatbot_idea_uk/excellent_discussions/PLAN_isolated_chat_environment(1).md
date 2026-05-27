# PLAN — Isolated Chat Environment

Plan for running the site chatbot's server-side pieces on infrastructure that is
**separate from the core build cluster**, so that live-traffic load, a
compromise, or a chat-code bug cannot interfere with the webdesign/build
workflows. Companion to `FOCUS_site_chatbot_edge_worker_and_context_pack.md`
(the chatbot design) and `086_site_chat_turns.sql` (the turn table). Written to
stand alone; grounding facts are in the Appendix.

---

## TL;DR

- **Live requests already never hit a cluster.** They hit the serverless edge
  worker, which talks only to object storage (read the context pack), the LLM
  provider, and a turn sink. The cluster is touched *after the fact*, draining
  the sink. So the live request path is already off-cluster.
- What still needs isolating is the **turn data**, the **drain/analytics
  processing**, and any **chat workflow code** — to sever three distinct blast
  vectors: **load**, **hack**, and **bug**.
- We do **not** reuse the existing multi-cluster dispatch (Phase 4a). That model
  shares cluster A's Kafka and Postgres on purpose; for chat that shared path is
  exactly the coupling we must remove. We reuse the chassis *binaries*, the
  `agent_definitions` schema, the action code, and the B2 storage pattern — but
  deployed against **separate** Kafka/DB/storage.
- The boundary is **one-directional and async**: the core publishes outward
  (the install trigger and the site content/identity the pack needs); the chat
  environment consumes and runs everything else on its own infra; **nothing on
  the chat side has synchronous or write access back into core.**
- The central decision is **how much to move** (Section 5). The minimal option
  severs all three live-traffic vectors and may need **no Kafka/chassis at all**
  on the chat side. The full option also isolates chat *install* code at the
  cost of duplicating the chassis + Ollama + pgvector.
- **Current lean (open):** Option **Y-copy** — run the full chassis on a separate
  cluster, curate the seed, experiment in a sandbox. The chassis image is
  monolithic, so copy is *config not code*; trimming the seed does not shrink the
  binary/attack surface (only a purpose-built slim image does, which is the real
  code job and a reversible later decision). See Section 5 "Current lean".
- **What the framework buys chat** is lane-dependent (Section 11): a fast lane
  (no framework, ships independently), a slow lane (agentic, well-warned
  latency), and a job lane (long-running submissions). The
  **building-as-a-service** example (Section 12) reframes the satellite as a
  customer-facing instance of the whole platform and largely settles X-vs-Y
  toward full chassis for that use case.

---

## 1. What we are isolating, precisely — the three vectors

"Don't let chat interfere with build" decomposes into three independent threats,
and isolation addresses each differently. Naming them keeps the plan honest
about what each piece of work actually buys.

1. **Load.** A traffic spike or a chatty bot produces a flood of turns. If the
   thing that ingests and stores turns shares a database or a chassis with the
   build workflows, that write-load competes with builds and can degrade them.
2. **Hack.** The edge worker is the only internet-facing surface. If it is
   compromised, the question is *what it can reach*. If its turn sink drains into
   the core database, a compromised worker becomes a path to poison core data or
   pivot inward. Isolation means the worker's reachable blast radius contains
   nothing belonging to core.
3. **Bug.** Chat workflow/action code (install, pack-build, drain) running on the
   *core* chassis can have a fault — a leak, a runaway spawn, a bad query — that
   degrades the shared chassis/Kafka/DB the build relies on.

The plan severs each vector explicitly; Section 5 is about how far to go.

---

## 2. The live path already bypasses the cluster

From the chatbot design, the runtime flow is:

```
browser → edge worker → { storage (read context pack),
                          LLM provider (generate),
                          turn sink (write turn) }
```

No cluster is in this path, and the **turn sink is a buffer** between live
traffic and anything we run. The chat environment drains that sink on its own
schedule:

```
chat environment: puller → reads sink → writes chat Postgres → analytics
```

Two consequences:

- A live-traffic spike fills the sink, not our database directly. The puller
  drains at its own rate; back-pressure stays at the buffer, never reaching
  build workflows. (This addresses **load** before we even add a second
  cluster.)
- Whatever the chat environment is, it is **not latency-critical to the user** —
  the user already has their answer by the time we drain. That gives us freedom
  to make the chat environment as small and as cut-down as we like.

So isolation is not about moving a hot request path off the cluster (there is
none); it is about making sure the **data, the drain, and the chat code** live
somewhere a build workflow never shares.

---

## 3. Why not reuse the existing multi-cluster dispatch

The platform already runs agents on a second cluster (Phase 4a, Rackspace Spot
`va001`). That mechanism (`remote-job-spawner`, `dispatch_agent`,
`system.dispatch.*`) is built so the remote agents **talk back over cluster A's
shared Kafka and use cluster A's Postgres** — "no federation needed" by design.
It is a way to borrow compute for *trusted internal build work* while keeping one
brain.

For chat that shared Kafka/DB path is the precise thing we are trying to remove:
it is a synchronous, write-capable channel from the chat side into core. Reusing
it would re-introduce all three vectors. So:

- **Do not reuse:** the shared-Kafka dispatch, the external Strimzi listener that
  exposes core's Kafka, the shared Postgres DSN.
- **Do reuse:** the chassis binaries, the `agent_definitions` schema and action
  registry, the action code (`execute_llm_prompt`, `rag_lookup`/`rag_index`,
  the storage actions), and the B2 storage client pattern — all deployed against
  the chat environment's *own* Kafka/DB/storage.

This is the inverse topology to Phase 4a: Phase 4a couples two clusters into one
system; the chat environment is a deliberately *decoupled* satellite.

---

## 4. The boundary contract

One rule: **the boundary is one-directional, async, and egress-from-core only.**

| Direction | Allowed | Mechanism |
|---|---|---|
| Core → chat env | Install trigger; the site identity + content the pack needs | Core publishes to shared object storage (a content export and/or the finished pack), exactly the egress-to-S3 motion the platform already uses for site assets and data exports. |
| Edge worker → chat env | Recorded turns | Worker writes to the sink (managed queue or chat-env-hosted queue); chat env drains it. |
| Chat env → core | **Nothing synchronous, nothing with write access** | If core ever needs chat analytics, the chat env *publishes* a summary outward and core pulls it — never the reverse. |

Nothing on the chat side holds a core credential, a core DB DSN, or a core Kafka
address. A separate B2 key (ideally a separate B2 *account*, not just a separate
bucket) scopes the chat env's storage so a compromise cannot touch core's
buckets.

---

## 5. The central decision — how much to move

Two coherent options, differing in **where context-pack building runs**, plus a
sub-question of **whether the chat env needs a chassis at all**.

### Option X — minimal satellite (recommended MVP)

Pack building **stays on core** (it already has the content, Ollama, and
pgvector; it is internal, controlled, not live-traffic-driven). Core publishes
finished packs to storage. The chat environment owns only the runtime-adjacent
pieces: the **turn store**, the **puller**, and **analytics**.

Because those pieces are not an orchestration — they are "drain a queue into a
table" and "run read queries" — **the chat env may need no Kafka and no chassis
at all**: a small managed Postgres (with the turn table) + one stateless
ingest/analytics service + a storage bucket. This is the most faithful reading of
"cut down to just what we need."

- Severs **load** (separate DB + the sink buffer) and **hack** (separate
  storage/DB, no inbound to core) fully.
- Severs **bug** for the *drain/analytics* code (it runs off-core).
- Leaves *pack-build* code on core — but that code runs on controlled internal
  triggers with the same blast radius as any other build agent, not on live
  traffic.

### Option Y — full chat chassis (end-state if install-code isolation is wanted)

Everything chat moves to the chat cluster, including pack building. Core
publishes raw site content/identity exports; the chat cluster runs a cut-down
chassis (its own Kafka + Postgres + pgvector + Ollama + only the chat
`agent_definitions` and actions) and builds packs itself via the
`site-chat-installer` orchestration.

- Additionally severs **bug** for the *install/pack-build* code.
- Costs: duplicate Ollama + pgvector + chassis + Kafka, plus a core→chat content
  export pipeline, plus running install-time embeddings on the satellite.

### Recommendation

Start with **Option X**. It removes every vector that is actually tied to live
traffic, with a fraction of the infrastructure — possibly no chassis at all.
Move to **Option Y** only if isolating internal install-code from core becomes a
real requirement (e.g. chat install needs to scale or fail independently). The
two are compatible: X's turn store/puller/analytics are unchanged by a later move
to Y; Y just adds the install side onto the same satellite.

### Current lean (not committed — kept open)

The working lean is toward **Option Y via copy** — stand up the **full chassis**
on a separate cluster, curate the `agent_definitions` *seed* down to what chat
needs, and **experiment** until the right shape emerges. Two flavours of Y to
keep distinct:

- **Y-copy** — deploy the existing (monolithic) chassis image against new
  Kafka/Postgres/storage; "fewer workflows" means a smaller seed, not different
  code. Experimenting = inserting/removing `agent_definition` rows and toggling
  adapters; no recompile.
- **Y-slim** — purpose-build a chat-only image with build actions compiled out.
  Smaller attack surface, more defensible for an internet-adjacent cluster, but a
  build-tooling job and a second artifact to maintain.

Key insight that makes the lean cheaper than it sounds: the chassis image is
**monolithic**, so Y-copy is a *deployment/config* exercise, not development. The
genuine code work is Y-slim. Therefore: running the full image now is the *less*
rebuild-y path, and trimming the seed does **not** shrink the binary or the
latent attack surface — only Y-slim does. The slim decision is reversible and
best made *after* we know which capabilities we are keeping.

Sandbox-first framing keeps the heavy, harder-to-reverse choice downstream:
stand the full-chassis copy up as an **experimentation sandbox** (driven by our
own candidate questions/capabilities, not live users) to learn what chat should
*do*; let that learning decide whether production chat runs the full image, a
slim image, or a warm-adapter shape. The **fast-lane bounded chatbot
(Section 11) can ship independently** of this whenever a real one is wanted in
front of users.

---

## 6. What the isolated environment contains (cut down)

### Under Option X (minimal)

| Piece | Notes |
|---|---|
| **Postgres (small, managed)** | Holds `site_chat_turns` and nothing else required. pgvector **not** needed here (no runtime/install retrieval on the satellite under X). |
| **Turn sink** | Managed queue (e.g. the edge provider's queue) or a tiny chat-env-hosted queue. Buffers live traffic. |
| **Ingest/analytics service** | One stateless binary: drains the sink → inserts turns (idempotent) → exposes/exports per-site analytics. No Kafka, no chassis. |
| **Object storage** | Separate bucket/account for the sink overflow and any chat artifacts. Packs themselves can stay in the existing site-assets storage (read by the worker) — decide per Section 10. |

### Under Option Y (full), additionally

| Piece | Notes |
|---|---|
| **Kafka (single small broker)** | Only the topics the chat workflows use. Its **own** cluster, not core's. |
| **Chassis (reused binary)** | Runs only chat `agent_definitions` + the actions they need. |
| **pgvector + Ollama** | For install-time grounding selection on the satellite. |
| **Content ingest** | Consumes core's site content/identity export to feed pack building. |

### Schema consequence — depends on X vs Y-copy

`086_site_chat_turns.sql` declares `site_id uuid REFERENCES sites(id) ON DELETE
CASCADE`, which assumes a `sites` table in the same database. The isolation route
changes whether that FK can exist:

- **Option X (no full schema):** there is no `sites` table on the satellite, so
  the FK cannot exist. Keep `site_id` as a plain `uuid` logical reference (no
  FK), or FK to a minimal local `sites_mirror`. Needs an isolated-DB variant of
  086 (drop the FK; keep the indexes).
- **Option Y-copy (full schema present):** the satellite *has* a `sites` table,
  so the FK **can** live as written — but now there are **two `sites` tables**
  with different populations (see Section 11's service example: core owns the
  portfolio; the satellite owns customer/SaaS sites). The open question shifts
  from "no table" to "how is the satellite's `sites` populated and how stale may
  it be" — answered per use case (a one-directional feed from core for rows the
  satellite needs, or the satellite originating its own rows).

Flag at apply time which database 086 targets — core vs satellite — and under Y
whether the satellite's `sites` is fed from core or locally originated.

---

## 7. Reuse map

| Reused from core (do not recreate) | New for the satellite |
|---|---|
| Chassis binaries; `agent_definitions` schema; action registry (Option Y only) | Separate Kafka cluster (Option Y); separate Postgres |
| Action code: `execute_llm_prompt`, `rag_lookup`/`rag_index`, storage actions | Ingest/analytics service (Option X) |
| B2 storage client pattern (`storage.NewS3Client` per call) | Separate B2 account/key scoped to chat buckets |
| The egress-to-S3 publish motion (site assets / data exports) | Core→chat content/pack export (publish step) |
| `086_site_chat_turns.sql` | Isolated-DB variant of 086 (FK dropped) |

Explicitly **not** reused: the shared-Kafka multi-cluster dispatch, the external
Strimzi listener, any core DSN/credential on the chat side.

---

## 8. Where to run it

- Not the existing `va001` collector cluster — that runs the coupled spawner and
  reaches back to core; repurposing it would defeat the isolation.
- Same datacentre is **not** required and arguably undesirable: a separate
  provider/account improves credential and failure isolation. Candidates: a fresh
  small managed Postgres + a container host anywhere (Option X needs very little),
  or a fresh small K8s cluster (Option Y).
- The edge worker is provider-agnostic already; it does not care where the
  satellite lives, since it only writes to the sink.

---

## 9. Phased plan (structural first)

1. **Settle Section 5** (X vs Y) and the Section 10 decisions — this sizes
   everything.
2. **Stand up the satellite store**: managed Postgres + the isolated-DB variant
   of `086`. Separate B2 account/key.
3. **Choose and stand up the sink** (managed queue preferred) and point the edge
   worker's `TurnSink` at it.
4. **Build the ingest/analytics service** (Option X) — drain → idempotent insert
   → analytics. This is the whole satellite under X.
5. **Wire the core→chat publish step**: under X, core publishes finished packs
   (it already builds them); under Y, core publishes content exports and the
   satellite builds packs.
6. **(Option Y only)** Stand up the cut-down chassis + Kafka + pgvector + Ollama
   and port the `site-chat-installer` orchestration to the satellite.
7. **Verify the boundary**: confirm nothing on the chat side can reach a core
   credential, DSN, or Kafka address; confirm the only core→chat flow is the
   publish step.

---

## 10. Risks and open decisions

Open decisions:

- **X vs Y, and Y-copy vs Y-slim** (Section 5). Current lean: Y-copy as a
  sandbox; keep open. The slim-vs-full *image* choice (latent attack surface vs
  build-tooling cost) is the surface-area decision and is best made after the
  capability list firms up.
- **Capability classes for chat** (Section 11): what we want chat to *do*
  (research / data queries / task completion / in-answer charts / build-as-a-
  service intake) — this list sizes how many adapters the slow lane needs and
  therefore how heavy the satellite must be. Still open by design; needs further
  conversation.
- **Where the context pack is stored**: in the existing site-assets storage (the
  worker already reads site assets from there) vs the satellite's own bucket.
  Keeping packs with site assets is simpler for the worker; a separate bucket is
  cleaner for isolation. Likely: packs stay with site assets (read-only,
  published by core), turns/sink overflow go to the satellite bucket.
- **Sink technology**: managed edge queue vs self-hosted. Recommendation: managed
  queue, to keep the buffer off any box we maintain.
- **Analytics access**: does the site owner / core need to see chat analytics,
  and if so via what egress (Section 4 says chat env publishes, core pulls).

Risks:

- **Two stores of site_id with no FK** (Option X/Y both). The satellite cannot
  enforce referential integrity against core's `sites`. Orphaned turns (site
  deleted on core, turns still on satellite) are possible; decide a reconciliation
  / retention policy. This is an accepted cost of isolation, not a bug.
- **Spot preemption** if the satellite runs on Rackspace Spot (nodes are
  ephemeral; IPs change). For a store-of-record (turns) this argues for managed
  Postgres over Spot-hosted Postgres.
- **Pack staleness**: under X, packs are built on core and published; a core
  outage stops pack *updates* but the worker keeps serving the last published
  pack from storage. Acceptable; note it.

---

## 11. What the agent framework buys chat (lanes and maturation)

For most site-chatbot questions — hours, services, location — the framework buys
**nothing** and costs latency: that is bounded Q&A, best served by a single
synchronous RAG call (the fast lane). The framework earns its place only when the
chat stops *answering about* the site and starts *acting with* the platform's
capabilities. This splits chat into lanes:

- **Fast lane (no framework).** Bounded Q&A, synchronous and streamed,
  sub-second-to-seconds. This is the edge worker + context pack from the
  companion FOCUS doc. **It can and should ship independently** of any satellite
  decision.
- **Slow lane (framework).** Turns that require *doing work*: live bounded
  research (web-search/scrape adapters), querying structured datasets the
  platform already builds (directories, pricing, enrichment), running one of the
  site's own tools on the user's behalf, generating an in-answer chart/infographic,
  or multi-step task completion. A cheap intent classifier routes here; the user
  is **clearly warned** the answer will take longer. Agentic task completion is
  the class a simple solution structurally **cannot** do.
- **Job lane (framework).** Submit a long-running job (e.g. *build me a site* —
  Section 12 example), acknowledge, track status, deliver on completion. This is
  job-submission + status, not conversation.

**Maturation / latency path.** Prove a slow-lane capability as a **spawned
agent** first (flexible, but pays the ~12s remote-spawn cost). Promote the
popular ones to **warm adapters** (long-running single-replica services).
End-state: a **resident chat-orchestrator adapter** that fans out to capability
adapters without spawning per turn. The slow lane gets faster
capability-by-capability; the end-state is not "agents per turn forever." This is
the concrete meaning of the earlier note: start as agents (some answers slow,
well warned), switch to adapters to improve latency.

This is why the satellite's *shape* (X vs Y, full vs slim image) cannot be
settled until the **capability classes** are chosen: three capabilities and
fifteen capabilities imply very different satellite weights.

---

## 12. Worked example — building-and-hosting as a service via chat

Recorded because it sharply reframes the satellite and largely settles the X-vs-Y
question for this use case. (Discussion artefact; revisit as it firms up.)

**Scenario.** `design.co.uk` is a site *core* built and on which the chatbot is
installed. A prospective customer uses that chat box to enter a domain + initial
spec; the agent framework runs the full build workflows and produces a new,
hosted site — itself with news, features, and a chat box.

**What this reframes.** The chat is no longer a Q&A box; it is the **intake +
orchestration front-end to the whole build platform, offered as a service.**
That has three consequences:

1. **The build must run on the satellite, not core.** Building customer sites on
   core would re-introduce the exact load/hack/bug vectors isolation exists to
   remove (anonymous, internet-triggered, cost-bearing builds competing with the
   portfolio). So the satellite becomes a **second, customer-facing instance of
   the entire platform** — which *requires* the full chassis. This is a much
   stronger justification for Y-copy than experimentation alone, and it
   effectively rules out Option X *for this use case*.
2. **The hosting model is already ours.** "Hybrid S3 + lambdas" *is* the edge
   pattern from the companion FOCUS doc: static on S3 + serverless functions for
   the dynamic bits. The chatbot was the first such function; build-intake and
   forms are more. So the chatbot architecture generalises into the SaaS hosting
   product.
3. **Two `sites` populations, not a stale copy.** Core owns the portfolio; the
   satellite owns customer SaaS sites (born and living on the satellite).
   `design.co.uk` straddles: built by core, but its chat triggers satellite
   builds. The satellite still needs the platform's **reusable building blocks**
   (component/tool library, style collections, knowledge_base patterns) — fed
   one-directionally from core — but not core's site data.

**End-to-end flow (mapped to existing workflows).**

1. **Conversational intake/briefing.** The chat detects build-intent and runs the
   briefing **as a dialogue** rather than a static form — a `briefing-agent`
   conducting a structured interview that fills the briefing questionnaire fields
   (reusing `018_briefing_questionnaire` / `002_intake_orchestrator`). *This is
   the highest-value "framework in chat" use: conversational intake beats a form,
   and it drives the core product.*
2. **Hand-off to intake orchestrator** on the satellite → new site row in the
   satellite's `sites` → build orchestration kicked off.
3. **Build workflows run on the satellite** (full chassis): classify
   (`003_site_classifier`) → plan → content (research/adoption sub-agents) →
   design/theme → components/tools → news/features → imagery. Minutes-to-hours,
   async; the customer is warned and tracks status (job lane).
4. **Hosting on hybrid S3 + lambdas.** Static published to S3; the new site's
   dynamic bits (its own chatbot, forms) are serverless functions. The new site
   gets its own chatbot via the install orchestration from the FOCUS doc — so the
   output is "a copy of itself with news, features, and a chat box." (Recursion:
   the platform builds sites that themselves have chatbots.)
5. **Delivery + recording.** Customer notified; the new site's chatbot records its
   turns to the satellite turn store via the same machinery.

**New concerns this surfaces (flag, not yet design):**

- **Cost/abuse exposure.** Anonymous "type a domain, get a site" is a spam/cost
  magnet — each build spends real tokens/compute. Must gate builds behind
  account/payment/quota and moderate the brief. The satellite isolation is *more*
  justified here, not less.
- **Accounts/billing/quotas.** Net-new SaaS concerns absent from the current
  platform.
- **Feeding the satellite's building blocks.** Decide what reusable platform
  assets the satellite needs and how they are fed one-directionally from core.
  Spot, US-East) runs `remote-job-spawner` and dispatched agents that connect
  back to the **primary cluster's** Kafka (via an added external Strimzi
  nodeport listener) and **primary Postgres**. Topic contract:
  `system.dispatch.requests` / `system.dispatch.responses`; agents talk back on
  `parent_responses_topic` on the shared Kafka. Cluster identity `CLUSTER_ID`
  (primary `uk_001`). This is a *coupled* model and is the wrong template for
  chat isolation.
- **Core Kafka.** Strimzi cluster `personae-kafka-cluster` in namespace `kafka`;
  in-cluster listeners only (plain 9092, tls 9093) by default. Core workloads run
  in namespace `ai-persona-system`.
- **Storage.** Backblaze B2 (S3-compatible). Credentials in K8s secret
  `personae-platform-secrets` (`B2_APPLICATION_KEY_ID`, `B2_APPLICATION_KEY`).
  Canonical pattern: construct an S3 client per call via `storage.NewS3Client`;
  the injected `params.StorageClient` is deprecated for new code.
- **Chatbot design (companion doc).** Live path is browser → edge worker →
  {storage, LLM, sink}; Layer 1 never serves inbound; the context pack is a
  build-time artifact published to storage; turns are recorded via a `TurnSink`
  and drained by a puller into `site_chat_turns`; embeddings/vectors (Ollama +
  pgvector) are used at install time, with optional in-worker runtime retrieval
  (v2).
