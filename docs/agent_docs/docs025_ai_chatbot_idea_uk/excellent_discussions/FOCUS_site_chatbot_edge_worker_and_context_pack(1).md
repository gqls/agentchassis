# FOCUS — Site Chatbot: Edge Worker and Context Pack

Canonical design for adding a bounded, per-domain AI chatbot to sites that are
deployed as static assets on object storage (Backblaze S3). Covers where the
runtime endpoint lives, why it is **not** an orchestrated agent, how the edge
worker is kept provider-agnostic, the build-time **context pack** that bounds
the conversation, and how each prompt/answer turn is recorded.

This document is written to stand alone. The findings from the two sister
projects that motivated the security decision (`ai-persona-system` and
`terraform-nginx-reverse-proxy`) are preserved in **Appendix A**, and the
platform facts this design leans on are summarised in **Appendix B**, because
those repos and docs may not be to hand when this is read.

---

## TL;DR

- The chatbot runtime is **synchronous request/response**, so it does **not**
  run through Kafka/the agent chassis. The chassis is a durable-async fabric;
  using it for live chat costs streaming and inherits async failure modes.
- The endpoint runs on a **serverless edge worker**, not a central nginx VM. A
  shared VM reintroduces exactly the long-lived-Linux-box attack surface we have
  repeatedly seen compromised, and — critically — forces the domain's DNS off
  object storage and *behind the box*, so we would lose the hack-resistance that
  static-on-S3 gives us. The edge worker keeps static on S3 and adds only a
  narrow `/api/*` compute path.
- To avoid lock-in, the worker is written against **Web-platform APIs only**,
  with a tiny platform shim. Cloudflare is the first host; Deno Deploy, Fastly
  Compute, Vercel Edge, or a self-hosted Node/Bun process behind our own thin
  proxy are all drop-in targets.
- The conversation is bounded by a **build-time context pack**: Layer 1
  publishes a per-domain JSON pack (identity + selected grounding + limits) to
  static storage. The worker loads it, composes a bounded prompt, calls the LLM,
  streams the answer, and records the turn. Layer 1 never receives inbound
  traffic — it publishes the pack and later *pulls* the recorded turns.
- Installing chat is a **separate orchestration**, triggered via a
  `maintenance_queue` task, **not** a stage of the (already complex) build
  pipeline. The build pipeline is untouched.
- **Ollama embeddings + pgvector** do the grounding selection at **install time
  on Layer 1** (always). Runtime per-question retrieval is an optional v2 that
  ships chunk vectors in the pack and does the search in-worker, so pgvector and
  any vendor edge-vector store stay out of the edge.
- Turns are recorded in **`086_site_chat_turns.sql`** (schema-checked), kept
  separate from the build-time `llm_call_log`.

---

## 1. Constraints (the hard ones)

These are non-negotiable and drive every decision below.

1. **Sites are static on object storage.** A deployed domain resolves to S3.
   There is no server in the request path and no way to run server-side code or
   hold a secret in the page. (This is also our main hack-resistance: there is
   nothing to compromise.)
2. **Layer 1 (the core cluster) never serves external traffic.** It builds and
   improves sites and holds all credentials and data. It may *publish outward*
   (it already pushes site assets and data exports to S3) and *pull inward*, but
   it must not expose an inbound endpoint to the public internet.
3. **No API keys on the static site.** A chatbot needs an LLM call, which needs
   a key and a server. That key cannot live in the page.
4. **Each prompt and answer must be recorded**, durably, per domain, in a way
   that does not pollute the build-time training log and respects that
   user-submitted text carries PII/GDPR weight (these are UK sites).
5. **Bounded per domain.** The bot answers about *this* site's subject and
   declines off-topic questions.
6. **Provider-agnostic edge.** We use Cloudflare today but must be able to move
   the edge elsewhere, or self-host it, without rewriting the chat logic.
7. **Reuse before rebuild.** Grounding reuses the existing RAG store; the LLM
   call reuses the existing prompt-composition shape; the widget reuses the
   existing client-side tool/component pipeline. See Appendix B.

---

## 2. Why the runtime turn is *not* an orchestrated agent

The platform's governing principle is "every agent is an orchestrator," and the
default reflex is to handle work by spawning an agent that consumes from Kafka
and replies on its parent's responses topic. For the **build-time** pieces of
this feature (deciding a site gets a bot; assembling its context pack) that is
exactly right and they *are* agents (Section 4). For the **runtime chat turn**
it is the wrong tool, and this is a deliberate, documented exception.

Two runtime shapes were considered:

- **Option A — synchronous handler.** The endpoint receives the POST and runs
  the flow in-process: load context → compose prompt → call LLM → stream back →
  record turn.
- **Option B — orchestrated agent.** The endpoint drops the turn onto Kafka; a
  `site-chat-agent` consumes it, does the work, and replies on the response
  topic; the endpoint waits and relays.

### Latency of Option B

It splits by whether the agent is warm:

| Sub-case | Added cost before any token |
|---|---|
| Agent **spawned per turn** (fresh Job pod) | Pod schedule + image pull + chassis boot. The codebase's own remote-startup wait is ~12s *before* work begins. A non-starter for chat. |
| Agent **long-lived** (warm consumer) | Kafka round-trip + coordinator path (claim → build collected_data → process → await → write back → route response). Roughly a few hundred ms to ~1s on a healthy cluster, *on top of* the LLM call. |

The LLM call itself (~1–5s for a short Claude answer) is identical under both
options, so Option B's *added* cost is the orchestration overhead.

### Why latency is not even the main reason

Two structural facts make Option B wrong for request/response regardless of how
warm the consumer is:

- **No streaming.** A synchronous handler streams tokens to the browser (SSE),
  so the first words appear in ~1s. A Kafka round-trip delivers the whole answer
  at once; the user watches a spinner for the full generation. The felt-latency
  difference dwarfs the raw millisecond difference.
- **Async failure modes leak into a live UX.** Consumer-group offset replay on
  pod restart, large-`collected_data` memory pressure, phantom-complete, and
  multi-minute awaited-request retry timeouts are all acceptable for durable
  background build work. In a live chat they surface as dropped or duplicated
  answers.

**Decision:** the runtime turn is synchronous (Option A). Kafka is for durable
async build work; chat is synchronous request/response. The edge worker *is* the
Option-A handler, hosted serverless rather than on a VM.

---

## 3. Where the endpoint runs: central nginx VM vs serverless edge

The remaining question was purely *where* the synchronous handler lives. The
candidates were a shared nginx reverse-proxy VM (the pattern in the
`terraform-nginx-reverse-proxy` sister project — see Appendix A) versus a
serverless edge worker.

### Security and maintenance comparison

| Axis | Central nginx VM | Serverless edge worker |
|---|---|---|
| **Attack surface** | A whole long-lived Ubuntu host: SSH, nginx/OpenSSL, plus anything else listening. Persists indefinitely. | Function code + its bindings. No SSH, no OS, no ports we own. |
| **Effect on S3 hack-resistance** | **Lost.** DNS must point at the VM, so the VM is in the path for *static* traffic too. The reason we like S3 (nothing to hack) is gone for any proxied site. | **Preserved.** Static stays on S3/edge; the worker handles only `/api/*`. A narrow stateless compute path is added, not a box in front of everything. |
| **Blast radius** | One shared front door for all sites; one compromise affects all of them and takes their static content down with it. | Per-site function, isolated. A bad deploy or compromise is scoped to one site's API path. |
| **Secrets** | LLM key + SSH creds on disk, in Terraform state/logs (in the sister project, secrets were marked `sensitive = false` and SSH password auth was enabled — see Appendix A). | Provider secret store, never on a filesystem we manage; rotatable; can be per-deployment. |
| **Patching** | We own kernel, nginx, OpenSSL, certbot, fail2ban, and (in the sister project) Grafana/Prometheus — forever. | Provider owns the runtime and TLS. Nothing to patch. |
| **TLS / DDoS** | certbot cron we maintain; DDoS is our problem. | Managed certificates and edge DDoS absorption. |
| **Uptime** | Box down ⇒ every proxied site down. | No single box; edge-distributed. |
| **Config drift** | High. The sister project's `main.tf` shipped with unresolved git merge-conflict markers and hand-edited per-domain confs that had diverged from the template. | Deploy *is* the artifact; published from the repo. |
| **Ongoing ops** | Indefinite sysadmin work. | Near-zero. |

### The honest costs of serverless

- **Vendor lock-in.** Real, but made thin by Section 5: the worker is
  Web-standard code behind a platform shim, and the heavy domain logic stays in
  Layer 1 / the context pack.
- **No local database.** A worker can't host Postgres. Turn recording needs a
  defined sink (Section 8). This is the one genuinely new design obligation
  serverless creates.
- **Per-request CPU/time limits.** Fine for a proxied, streamed LLM call;
  long-running work would not belong here anyway.

### Recommendation

Serverless edge worker. It preserves S3's hack-resistance, removes the box we
keep getting burned by, keeps secrets off any filesystem we manage, and confines
blast radius per site. The central nginx VM reintroduces precisely the risk the
sister project demonstrates and drags static content behind it.

---

## 4. End-to-end architecture

Two paths: an **install path** that is a standalone orchestration (deliberately
**not** part of the site build pipeline), and a runtime path that is the
synchronous edge worker.

### Install path — a separate orchestration, not a build-pipeline step

The build pipeline is already complex, and chat is an opt-in add-on to an
*already-built, already-deployed* site, not a stage of building one. So chat
installation is its own orchestration with its own workflow, kept entirely out
of the build pipeline. This also matches the platform preference to spawn
sub-agents with their own workflows (clear logs, separate responsibilities)
rather than grow an existing workflow.

**Trigger (reuse, no new mechanism).** Installation is requested by enqueuing a
maintenance task — `task_type = 'install_chat'` on the existing
`maintenance_queue` (which already has `site_id`, `payload`, `status`, retries).
Nothing in the build pipeline is touched; the build pipeline does not know chat
exists.

**`site-chat-installer` (orchestrator).** Claimed from the queue. Its workflow is
thin; it checks preconditions (site built + deployed) and spawns three
sub-agents, each with its own workflow, each replying on the installer's
responses topic:

1. **`chat-context-builder`** — assembles the **context pack** (Section 7):
   reads identity from `site_specs`, embeds the site's own content with Ollama
   and selects per-domain grounding via pgvector (Section 7.1), applies a token
   budget, and publishes the pack as static JSON next to the site assets on S3
   (egress only).
2. **`chat-widget-installer`** — generates/forks the chat widget as a
   client-side component and integrates it (page placement, nav, deploy) by
   reusing the existing component/tool pipeline actions. The *only* difference
   from a normal tool is that its script `POST`s to `/api/chat` instead of
   computing in-browser.
3. **`chat-route-registrar`** — records the `/api/chat` route + bounding config
   for the domain (where the worker reads it from) and marks chat installed on
   the site (e.g. `sites.settings`), so the state is queryable and reversible.

All are ordinary orchestrators that put complexity in Go action code, per
platform convention. The whole feature can be removed for a site by an
`uninstall_chat` task without disturbing the build pipeline.

### Runtime (synchronous edge worker)

```
Browser (static page on S3/edge)
   │  POST /api/chat  { sessionId, message, history? }
   ▼
Edge worker  (provider-agnostic; Cloudflare today)
   1. resolve domain from Host
   2. ContextStore.get(domain) → context pack   (edge-cached static JSON)
   3. guard: size/turn/rate limits from pack
   4. compose bounded system prompt (identity + grounding + refusal rule)
   5. LLMClient.complete(prompt, userMessage) → token stream
   6. stream tokens back to browser (SSE)
   7. TurnSink.append(turn record)              (fire-and-forget)
   ▼
Layer 1 later PULLS recorded turns → flywheel log + per-site analytics
```

Layer 1 publishes the pack and pulls the turns. It never receives an inbound
request. The LLM key lives only in the worker's secret store. Static content
never leaves S3/edge.

---

## 5. Keeping the worker provider-agnostic

The portability strategy is a **platform-agnostic core function plus a thin
per-platform shim**. The core never imports anything vendor-specific.

### What is allowed in the core (Web-platform only)

- The `fetch`-handler shape: a pure `handleChat(request, deps) → Response`.
- `fetch()` for the outbound LLM HTTPS call.
- `ReadableStream` / `TransformStream` / `TextEncoder` for SSE streaming.
- `Request` / `Response` / `Headers` / `URL`.
- `crypto.randomUUID()`, `crypto.subtle`.

### What is forbidden in the core (vendor-specific)

- Cloudflare Durable Objects, Vectorize, `caches.default`, or any `cf`-flavoured
  binding referenced directly.
- Any direct reference to a specific KV/D1/R2 client. These are reached only
  through the `deps` adapters below.

### The dependency contract (`deps`)

The core receives three small adapters. Each has a Cloudflare implementation and
a portable HTTP implementation, wired by the shim.

```ts
interface ContextStore {
  // Returns the bounded context pack for a domain, or null if no bot.
  get(domain: string): Promise<ContextPack | null>;
}

interface LLMClient {
  // Streams completion tokens. Default impl: Anthropic Messages API over fetch.
  // Swap-able for a self-hosted model endpoint later.
  complete(input: {
    system: string;
    messages: { role: "user" | "assistant"; content: string }[];
    maxTokens: number;
  }): Promise<ReadableStream<Uint8Array>>;
}

interface TurnSink {
  // Records one prompt/answer turn. Fire-and-forget; must not block the response.
  append(turn: TurnRecord): Promise<void>;
}
```

### Reference implementations

| Adapter | Cloudflare impl | Portable / self-hosted impl |
|---|---|---|
| `ContextStore` | `fetch(packUrl)` with edge cache, **or** KV `get` | `fetch(packUrl)` against the S3/CDN URL of the pack — no binding needed, since the pack is static JSON |
| `LLMClient` | `fetch` to Anthropic Messages API; key from `env` secret | identical `fetch`; key from process env / secret manager |
| `TurnSink` | D1 insert, **or** Queue produce | HTTPS `POST` to a turn-ingest queue, or write to SQLite/Postgres if self-hosted |

The most portable `ContextStore` is simply an HTTP `GET` of the pack's static URL
— it needs no platform binding at all and works identically everywhere. KV is an
optional latency optimisation, not a requirement.

### The shim (per platform)

Each platform gets a ~20-line entrypoint that constructs the adapters from its
own `env` and calls the core:

```ts
// cloudflare entrypoint
export default {
  async fetch(request, env, ctx) {
    const deps = {
      contextStore: makeHttpContextStore(env.PACK_BASE_URL),
      llm: makeAnthropicClient(env.LLM_API_KEY),
      turnSink: makeQueueSink(env.TURN_INGEST_URL, env.INGEST_TOKEN),
    };
    return handleChat(request, deps);
  }
};
```

A Deno/Bun/Node entrypoint differs only in how it reads env and starts the
server; `handleChat` is byte-for-byte the same. **Rate limiting** is the one
concern that is least portable (it needs shared state); treat it as a fourth,
optional adapter or lean on the host platform's WAF rule, and keep a coarse
in-pack per-session cap as the portable floor.

---

## 6. Worker flow sketch

Provider-agnostic core. Pseudocode, not final code; error handling abbreviated.

```ts
async function handleChat(request: Request, deps: Deps): Promise<Response> {
  if (request.method !== "POST") return json(405, { error: "method" });

  const domain = new URL(request.url).hostname;          // bound to this site
  const pack = await deps.contextStore.get(domain);
  if (!pack) return json(404, { error: "no_chat_for_domain" });

  const body = await request.json();                     // { sessionId, message, history? }
  const message = (body.message ?? "").slice(0, pack.limits.maxInputChars);
  if (!message.trim()) return json(400, { error: "empty" });

  // Operational bounding
  const history = (body.history ?? []).slice(-pack.limits.maxHistoryTurns);
  if (history.length >= pack.limits.maxTurnsPerSession) {
    return json(200, { reply: pack.scope.turnLimitMessage, capped: true });
  }

  // Prompt bounding: identity + grounding + explicit refusal rule
  const system = composeSystemPrompt(pack);

  const started = Date.now();
  const stream = await deps.llm.complete({
    system,
    messages: [...history, { role: "user", content: message }],
    maxTokens: pack.limits.maxOutputTokens,
  });

  // Tee the stream: one copy to the browser, one accumulates for recording
  const [toClient, toRecord] = stream.tee();

  // Fire-and-forget recording; never blocks or fails the user response
  recordWhenDone(deps.turnSink, {
    turnId: crypto.randomUUID(),
    siteId: pack.siteId,
    domain,
    sessionId: body.sessionId ?? null,
    question: message,
    // answer + tokens + latency filled in by the accumulator
    accumulate: toRecord,
    model: pack.model.id,
    packVersion: pack.version,
    startedAt: started,
  });

  return sseResponse(toClient);                          // tokens stream to browser
}

function composeSystemPrompt(pack: ContextPack): string {
  const grounding = pack.grounding
    .map((g, i) => `[${i + 1}] ${g.title}\n${g.text}`)
    .join("\n\n");
  return [
    `You are the assistant for ${pack.identity.name}, a ${pack.identity.industry} site.`,
    pack.identity.description,
    pack.scope.instructions,                              // tone/voice from site_specs
    `Answer ONLY using the reference material below and general knowledge`,
    `directly relevant to ${pack.identity.name}'s subject. If a question is`,
    `outside that scope, reply exactly: "${pack.scope.refusalMessage}"`,
    ``,
    `Reference material:`,
    grounding,
  ].join("\n");
}
```

Notes:

- The refusal rule is enforced in the prompt; for a cheaper first line of
  defence an optional pre-classifier (a small local model) can reject obviously
  off-topic input before spending an LLM call, but that is a later optimisation
  and is kept out of the core to preserve portability.
- `recordWhenDone` consumes the accumulator copy of the stream, assembles the
  full answer, computes tokens/latency, and calls `TurnSink.append`. It runs
  after the response has begun streaming, so recording never adds user-visible
  latency. On Cloudflare this uses `ctx.waitUntil`; the portable shim uses an
  equivalent "don't await" pattern.

---

## 7. Context-pack shape

One JSON document per domain, produced at build time by the
`chat-context-builder` agent and published to static storage. It is the entire
bounded context the worker needs — the worker holds no per-site logic.

```jsonc
{
  "version": 3,                          // bump on rebuild; lets the worker cache-bust
  "siteId": "ste_01H...",
  "domain": "example-vets.co.uk",
  "generatedAt": "2026-05-26T10:00:00Z",
  "contentHash": "sha256:…",             // hash of grounding; for cache + change detection

  "identity": {
    "name": "Example Veterinary Clinic",
    "industry": "veterinary",
    "description": "A small-animal vet practice in Bristol offering …",
    "voice": "warm, plain-English, reassuring"
  },

  "scope": {
    "instructions": "Speak in the practice's warm, plain-English voice.",
    "refusalMessage": "I can only help with questions about Example Veterinary Clinic and pet care. Is there something about the practice I can help with?",
    "bannedTopics": ["medical diagnosis", "pricing of competitors"]
  },

  "grounding": [                          // build-time selected, bounded by token budget
    {
      "id": "kb_…",
      "title": "Opening hours and location",
      "sourceUrl": "https://example-vets.co.uk/contact",
      "text": "Open Mon–Fri 8:30–18:00 …"
    },
    {
      "id": "kb_…",
      "title": "Services offered",
      "sourceUrl": "https://example-vets.co.uk/services",
      "text": "Vaccinations, dental, microchipping …"
    }
    // … top-N chunks within the budget
  ],

  "model": {
    "id": "claude-haiku-…",              // suggested model for this site's bot
    "fallbackId": "claude-…"
  },

  "limits": {
    "maxInputChars": 1000,
    "maxOutputTokens": 600,
    "maxTurnsPerSession": 20,
    "maxHistoryTurns": 8
  }
}
```

### 7.1 Grounding strategy — Ollama embeddings + pgvector, placed correctly

We already run Ollama embeddings (nomic-embed-text, 768-dim, matching the
existing `knowledge_base vector(768)`) and pgvector. The design uses **both** —
the only real question is *where* each runs, because pgvector lives in Layer 1's
Postgres and Layer 1 must not serve external inbound traffic, while the worker is
at the edge and must stay portable. So we split by phase:

- **Install time (always): Ollama + pgvector on Layer 1.** `chat-context-builder`
  embeds the site's own content with Ollama and uses pgvector to select the
  per-domain grounding set (and, for larger sites, to cluster it into topics).
  This is where both tools do their work, fully inside Layer 1, no exposure. The
  selected chunks are frozen into the pack.

- **Runtime — phased:**
  - **v1: no runtime retrieval.** The pack carries a bounded grounding set and
    the worker passes all of it into the prompt. No edge embedding, no edge
    vector store; the worker stays fully portable and self-contained. Best when
    a site's material fits a sensible token budget.
  - **v2: per-question retrieval without exposing Layer 1.** The pack *also*
    carries the chunk vectors (produced at install time by the same Ollama
    model, so they share the pgvector index's vector space). At runtime the
    worker does the similarity search **in-process** — cosine over a few hundred
    768-dim vectors is trivial compute — selecting the top chunks for the
    question. The only thing it cannot do alone is embed the *question* in the
    same space; that needs a nomic embedding reachable from the edge.

  For the question-embedding step in v2, keep portability and keep Layer 1
  private by putting nomic behind a **narrow, stateless embedding-only
  endpoint** (it holds no data — just the model — so it carries almost none of
  the VM risk discussed in Section 3, and can itself be serverless or a tiny
  service). The worker reaches it through a fourth adapter, `Embedder.embed(text)
  → number[]`, with the same shim pattern as the others; a self-host can point it
  at its own Ollama. We deliberately do **not** put pgvector or a vendor edge
  vector store at the edge — pgvector stays the install-time engine, and the
  runtime search is in-worker over pack-shipped vectors.

This answers the earlier tension directly: pgvector is used every time (install
time), Ollama embeddings are used at install time always and at runtime only in
v2, and neither forces an inbound path to Layer 1 or a vendor lock-in at the
edge. A middle option remains available — **topic-scoped packs** picked by cheap
keyword match — when per-question semantic retrieval is overkill.

---

## 8. Recording each turn

### The turn record

```jsonc
{
  "id": "uuid",                    // edge-generated; becomes the PK (idempotent ingest)
  "siteId": "ste_…",
  "domain": "example-vets.co.uk",
  "sessionId": "client-generated, opaque",
  "question": "…",                 // user text — treat as PII
  "answer": "…",
  "model": "claude-haiku-…",
  "packVersion": 3,
  "groundingIds": ["kb_…", "kb_…"],// chunks fed into the prompt (debugging "why did it say that")
  "inputTokens": 812,              // names mirror llm_call_log
  "outputTokens": 143,
  "latencyMs": 1840,
  "refused": false,                // hit the off-topic refusal path
  "capped": false,                 // hit a turn/length limit
  "createdAt": "2026-05-26T10:01:12Z"
}
```

### Where turns land — and why a *separate* store

Turns are recorded via `TurnSink`. They must **not** go into the build-time LLM
call log: that log is the training flywheel (sliced by agent type and work item,
with its own retention), and end-user chat is a different owner, a different
privacy profile (user-submitted PII on public UK sites), and a different access
pattern (a site owner wants their own chat analytics). Mixing them entangles
retention policy and pollutes the training slices.

Two sink implementations, by host:

- **Cloudflare:** insert into D1 (SQL at the edge) or produce to a Queue that a
  Layer-1 puller drains. A queue is preferable — it keeps the edge write trivial
  and lets Layer 1 own the durable copy.
- **Self-hosted:** `POST` to a turn-ingest queue, or write to SQLite/Postgres
  co-located with the self-hosted worker.

Either way, **Layer 1 pulls** the durable copy on its own schedule (it does not
receive an inbound push), then writes it to the chat-turn table below and,
separately, can feed sanitised turns into the flywheel as a distinct slice.

### The Layer-1 table

Written as migration `086_site_chat_turns.sql` (numbering to be reconciled
against the live migrations directory before applying). It was checked against
the live schema: `sites.id` is `uuid`/`gen_random_uuid()`, so `site_id` is
`uuid REFERENCES sites(id) ON DELETE CASCADE`, and `domain` is
`varchar(255)` to match `sites.domain`. Token/latency columns
(`input_tokens`, `output_tokens`, `latency_ms`, `created_at`) are named to match
`llm_call_log` so the two LLM-logging tables read alike. The edge-supplied turn
uuid is the primary key, so the Layer-1 puller ingests idempotently with
`ON CONFLICT (id) DO NOTHING`. `created_at` is the edge time the turn happened;
`ingested_at` is when Layer 1 stored it. Raw IPs are never stored — only an
optional salted `client_ip_hash` — and a short retention/redaction policy for
`question`/`answer` should be set before go-live given the PII content.

---

## 9. Bounded context — the three layers

"Bounded" means three distinct things; conflating them is where chatbots drift.

1. **Retrieval bounding** — only this site's material is available. Achieved at
   build time: grounding is selected from the RAG store filtered to the domain
   and frozen into the pack. The worker cannot reach anything else.
2. **Prompt bounding** — the system prompt pins the bot to the site identity and
   instructs it to emit the exact refusal message for out-of-scope questions
   (see `composeSystemPrompt`).
3. **Operational bounding** — input length, output tokens, turns per session,
   history window, and rate limiting. These live in `pack.limits` and the rate
   adapter, enforced in the worker before and around the LLM call.

---

## 10. Open decisions and next steps

Open decisions:

- **Turn sink:** edge SQL (D1) vs edge Queue drained by Layer 1. Recommendation:
  Queue + Layer-1 pull, to keep the durable copy on Layer 1.
- **Rate-limiting mechanism:** host WAF rule vs a portable shared-state adapter.
  Recommendation: start with the in-pack per-session cap (portable) plus the
  host's WAF, defer a cross-request limiter.
- **Per-site vs shared LLM key:** per-deployment key narrows blast radius;
  shared key is simpler. Recommendation: per-site or per-batch key if the host's
  secret model makes it cheap.
- **Model choice per site:** set in `pack.model`; default to the cheapest model
  that holds the bounded-Q&A quality bar.
- **Runtime retrieval (v2) trigger:** when to move a site from v1 (whole-pack
  grounding) to v2 (in-worker vector search + narrow embedding endpoint), i.e.
  the pack-size/quality threshold that justifies standing up the embedder.

Resolved:

- *Install is a separate orchestration*, triggered via a `maintenance_queue`
  `install_chat` task — **not** a build-pipeline stage (Section 4).
- *Embeddings/vectors*: pgvector + Ollama at install time on Layer 1 always;
  runtime retrieval is the optional v2 with the vectors shipped in the pack
  (Section 7.1).
- *Turn table*: written as `086_site_chat_turns.sql`, schema-checked (Section 8).

Suggested build order (structural first):

1. Define the **context-pack schema** and the **`handleChat` core** with the
   adapter interfaces — this is the contract everything else depends on.
2. Implement the **portable adapters** (HTTP `ContextStore`, Anthropic
   `LLMClient`, Queue `TurnSink`) and the Cloudflare shim. (`Embedder` is v2.)
3. Apply **`086_site_chat_turns.sql`** (after reconciling the migration number)
   and build the **Layer-1 turn puller** that drains the sink into it.
4. Build the **`site-chat-installer`** orchestration and its sub-agents
   (`chat-context-builder` using Ollama+pgvector, `chat-widget-installer`
   reusing the component pipeline, `chat-route-registrar`), triggered by the
   `install_chat` maintenance task. Nothing in the build pipeline changes.
5. Wire DNS/edge for opted-in domains so `/api/chat` reaches the worker while
   static stays on S3.
6. (v2, if needed) Stand up the narrow embedding endpoint and ship chunk vectors
   in the pack for in-worker per-question retrieval.

---

## Appendix A — Sister-project findings (preserved)

These are the relevant facts from the two sister projects, recorded here because
those repos may be unavailable when this document is read.

**`ai-persona-system`** — the older multi-agent system. Its infrastructure
Terraform (`deployments/terraform/{1-cluster,2-strimzi,3-resources}`) provisions
a Kubernetes cluster with Strimzi/Kafka via the `kubernetes`/`helm` providers
against a local kubeconfig. It contains no OVH provisioning and no public proxy;
it is not relevant to the edge decision beyond confirming the agent/Kafka
lineage.

**`terraform-nginx-reverse-proxy`** — a single OVH VM acting as an internet-
facing nginx reverse proxy in front of a Kubernetes NodePort. It is the concrete
example of the "central nginx box" option and demonstrates why that option was
rejected:

- The VM is created **out-of-band** (a fixed IP, `51.89.148.216`); Terraform
  only configures it via `null_resource` + `remote-exec` over SSH. There is no
  OVH API provider — the box is hand-created.
- The SSH `connection` block uses **password authentication**
  (`password_for_ovh_ssh`) alongside a key, and both that password and the nginx
  `htpasswd_password` are declared `sensitive = false`, so they land in
  Terraform state and logs in clear.
- Hardening present: `ufw` (22/80/443/9100, plus 9090/3000 in one variant),
  `fail2ban` (`nginx-http-auth` jail, ban after 5 retries), nginx
  `limit_req_zone` rate limiting (10 r/s, burst 20), `certbot --nginx` for TLS,
  `logrotate`, gzip, and Basic-Auth on admin endpoints.
- Grafana (3000) and Prometheus (9090) were exposed on the same box in one
  variant — additional long-lived attack surface.
- IP allow-listing was done with the nginx `if ($remote_addr = …)` anti-pattern.
- `main.tf` shipped with unresolved git merge-conflict markers
  (`<<<<<<< Updated upstream` … `>>>>>>> Stashed changes`), and the
  `original_confs/*` per-domain files had diverged from the `.tpl` template —
  i.e. observable configuration drift.

The takeaway: even a reasonably hardened single VM is a long-lived, broad,
hand-maintained attack surface, and pointing site DNS at it would put that
surface in front of otherwise-static content. This is the lived "nginx box keeps
getting hacked" experience, and it is the expected outcome of the pattern.

---

## Appendix B — Platform facts this design depends on

- **Layers.** Layer 1 is the core Kubernetes cluster (agents, Kafka, Postgres,
  all credentials, all build work); it publishes outward and pulls inward but
  never serves inbound public traffic. Layer 2 is client delivery; today that is
  static assets on Backblaze S3 with nothing in the request path. This design
  adds the edge worker as the only Layer-2 compute.
- **Deploy path.** Build agents commit site assets to git; CI publishes them to
  Backblaze S3. Data exports use the same publish-to-S3 mechanism, so
  "Layer 1 publishes a JSON artifact to S3" (the context pack) is an existing,
  proven motion.
- **Orchestration model.** Work is normally done by agents that consume from
  Kafka and reply on their parent's responses topic, keeping workflow logic thin
  and complexity in Go action code. The build-time chat pieces follow this; the
  runtime turn is the documented synchronous exception (Section 2).
- **RAG store.** A pgvector-backed knowledge store keyed by collection / domain /
  industry with metadata already exists, with index and lookup actions and a
  trigram fallback. Per-domain grounding selection reuses it at build time.
- **LLM call + prompt composition.** An existing LLM-call action performs
  endpoint-health-aware routing and logs build-time calls to the training
  flywheel; an established prompt-composition shape assembles system prompts.
  The worker mirrors the prompt-composition shape but calls the provider
  directly over HTTPS (the key cannot live on Layer 1's behalf at the edge).
- **Tool/component pipeline.** Interactive client-side components are generated,
  integrated, navigated, and deployed to S3 by an existing pipeline. The chat
  widget is one more component on that path; its only difference is that its
  script calls `/api/chat`.
