# agentchassis — Framework Capabilities Inventory (2026-07-21)

*Compiled from a hard sweep of the repo + live docs. Self-contained for pasting
into another chat. Status tags: **[LIVE]** in production; **[PARTIAL]** built but
mid-rollout / not fully wired; **[DESIGNED]** specced, not built.*

## What it is
An autonomous multi-agent platform that **plans, builds, operates, verifies and
repairs a fleet of ~1,000+ content websites**. It is a Kafka-orchestrated,
Postgres-backed "agent chassis" on Kubernetes: every pod is both worker and
orchestrator, loads its workflow from the database, and runs a crash-resumable
saga. "Sites" are the universal unit of work; even internal jobs (news pools,
system tasks) are modelled as synthetic sites so all per-site machinery is reused.
Humans make taste/truth decisions; agents do everything in between and inspect
their own output.

## How it works (architecture, newcomer view)
- Stateless saga coordinator: any pod picks up any job because all state lives in
  Postgres, not memory. Workflows survive crashes and resume.
- Workflows are **data, not code**: each agent's steps live in
  `agent_definitions.default_config.workflow.steps` — versionable, inspectable and
  changeable by SQL with no redeploy.
- ~300+ pluggable named actions dispatched from step config (site build, imagery,
  RAG, HITL checkpoints, diagnosis/council review, work-item claim/complete…).
- Remote agent calls and cross-cluster work go over Kafka topics, correlated by
  correlation/causation IDs; a remote-job-spawner creates K8s Jobs in other
  clusters (federated/multi-cluster).
- Runaway protection: circuit breaker (max step executions), cycle detection,
  optimistic-lock retries with jittered backoff, fuel/governance metering.
- Human-in-the-loop pause/resume without suspending the orchestration.

---

## 1. Website & content generation  [LIVE]
- Plan a full multi-page site from a one-line brief (pages + design direction +
  content strategy + imagery plan) and build it end to end.
- Assemble complete HTML pages from a library of stored, validated components;
  plan each page's sections from component input-schemas + resolved data sources.
- Render shared site chrome (header/footer/head) once and reuse across pages;
  build navigation from DB nav tables with a pages-table fallback.
- Incrementally re-render a single page, a page's sections, or a whole-site batch.
- Rules-driven page inclusion (e.g. auto-place a news page when the site's
  classification enables a feed).
- Brand/theme/palette: deterministically render CSS from a palette + layout +
  typography spec; **re-theme an entire site** (~1h restyle); fork a site's CSS
  into a reusable library theme (pending human review).
- Inject brand head tags (favicon, OpenGraph/Twitter cards) into every page.
- Blog/news: rebuild blog listing pages, render a live news section, emit an
  outbound RSS 2.0 feed.xml.
- Render per-site JavaScript bundles for interactivity.

## 2. Interactive tools + self-verification  [LIVE]
- Generate interactive tools (calculators, simulators, games) as LLM-authored
  HTML → a tool component + a tool page, cross-linked contextually into existing
  content pages.
- **Self-verifying tools**: each tool carries a living spec + change-history in the
  DB; the platform drives it in a *real headless browser* and confirms it works
  (`page_status_ok`, `selector_exists`, `no_console_errors`) — filing its own fix
  ticket if it breaks. Checks it can't run are reported as skipped, never faked.
- Anti-fabrication net: a data-backed tool cannot be recreated with invented data.

## 3. Imagery generation  [LIVE]
- Two lanes: plan-driven imagery (logo/hero/illustration/icon/infographic/sprite)
  and content-driven imagery attached to articles/news/products.
- **Dual image-model routing by kind**: Stability SDXL for photographic work;
  Google "Banana" (`gemini-3-pro-image-preview`) for flat icons/illustrations/
  heroes, honouring up to 20 reference images + negative prompts.
- Per-site machine-readable **imagery style guide** (palette, medium, mood, avoid
  list, reference anchors) applied to every generation, with per-kind overrides.
- Per-article "content hero" auto-generated from title+description; cards + OG
  crops derived from the one hero generation.
- Logo permanence: generate → human-approve → **lock**; regeneration refuses to
  touch locked assets; favicon/OG derived from the approved logo.
- Sprite sheets for themed list bullets; Lucide webfont icons for UI chrome.
- Deploy pipeline: optimise → commit image to git (durable path via Cloudflare
  worker, not an expiring S3 URL) → re-render pages. Per-kind byte/dimension
  budgets + image-integrity checks (placeholder-in-use, 404, undeployed, missing).
- **Hard rule: data charts are code-rendered from real series (go-echarts), never
  drawn by a diffusion model** — the model only adds an annotation layer.

## 4. Claims / fact verification — the "honesty machinery"  [LIVE, V5 DESIGNED]
Per-site fact register (`evidence_base`): facts with value/kind/source (sql |
artifact | attested | citation), tolerances, a banned-claims blacklist and an
allowed-entities list. Layered checks:
- **V1 [LIVE]** build-time gate blocks known-false claims and flags business
  numbers matching no registered fact → routes to human review; a post-deploy
  scan re-checks live pages for drift/hand-edits (caught an invented service name
  within hours).
- **V2 [LIVE]** the writer prompt is given the registered facts as *the only*
  numbers/entities it may assert — bounded "use only these", not "never invent".
- **V3 [LIVE]** an LLM claims-auditor classifies every prose assertion as
  supported / could-be-framed / unsupported against the register; findings go to
  humans, never auto-fixed.
- **V4 [LIVE]** a scheduled freshness pass re-runs SQL-sourced facts, updates
  values, raises `stale_evidence` on drift (incl. *under*-claiming) with
  compare-and-swap so a human edit is never lost (caught a site overclaiming by
  476 records).
- **V5 [DESIGNED]** an evidence-researcher: web-search/scrape → extract atomic
  claims with verbatim quotes → verify by re-fetching the URL and asserting the
  quote still appears (defeats plausible-but-fake citations).
- Origin cases it exists to prevent: invented client case studies, "2,767 Awards
  Won", fabricated vet prices with legal exposure. Governing rule: deterministic
  checks first, evidence-is-data, **truth decisions are always human**.

## 5. Link / CTA / section integrity  [PARTIAL / LIVE]
- CTA-integrity: enforces "a label implies a real destination" — buttons with no
  resolvable target simply don't render (catches frozen wrong labels, empty href,
  dead in-page anchors, AI-invented hostnames).
- Link-integrity verifies against *live* sites (fetch every page, follow every
  link) — found 312 broken links across 117 pages incl. footer 404s.
- Legal pages (privacy/terms) can be marked "owned" so auto-rebuild never
  silently rewrites reviewed legal text.
- Loop-integrity gate: a per-item-type verifier registry at work-item completion,
  so a handler no-op can no longer stamp a defect "fixed"; empty query-filled
  sections render a localised "more coming soon" instead of a blank box; an
  LLM-apology guard blocks meta-commentary from shipping as page copy.

## 6. Multi-agent loops — the core differentiator  [MIXED]
Every loop is a declarative DB workflow with adversarial verification + human
gates, not a single LLM call.
- **Diagnosis loop [LIVE]** — read-only: turns a bug symptom into a *cited* root
  cause by forming a hypothesis and gathering evidence across three tiers (static
  Go code / live DB / runtime records), re-scoping by following what the evidence
  names. Issues CONFIRMED only with citations across all tiers, else ABSTAIN — a
  REFUTED hypothesis is scored a success.
- **Fix loop [LIVE]** — turns a confirmed diagnosis into a *constrained* edit plan
  (≤8 edits, file allowlist, expected symbols), council-reviews it, then applies
  it through a **caged implementer** (read-only token; a separate git-adapter
  holds the write credential; hard file allowlist + build gate before any PR; a
  red gate yields no PR) to a branch/PR.
- **Council / council-gate [LIVE, advisory]** — multi-seat adversarial review of a
  proposed change: **16 named seats** (2 always-on: edit-quality + guardian[hard
  veto]; the rest relevance-gated so a wide panel runs near 2-seat cost),
  aggregated by a deterministic router into **APPROVED / REVISE / REJECTED**. Any
  thread can submit a plan+rationale for review before committing. A safety check
  that *abstains* still halts approval (fail-closed on omission).
- **Feature-builder loop [PARTIAL]** — a *designer* turns an approved capability
  spec into a staged plan (ordered edits, per-stage allowlist, per-stage gates),
  council-reviewed; an *implementer* then builds stage-by-stage (one branch, one
  commit + build gate per stage, a derived `go test` end gate, one PR). Designer
  half proven; implementer half built but had never fired end-to-end as of the
  docs.
- **Experience loop [LIVE]** — a planner writes what a whole visitor *experience*
  should be (journeys, promises, data needs); a council of 4 critics attacks it
  (journey-completion, feasibility, fabrication/honesty, scope-cutting) until it
  holds.
- **Concept register [PARTIAL]** — extracts ~1,633 concepts from ~4,111 docs,
  verifies each against live code/DB, and mints one expert council *seat per
  concept area*; concepts rediscovered 4–6× flag which seats to build first.
- **Immune system [LIVE]** — automated sweeps scan every recorded fleet failure,
  route genuine platform-wide bugs into the diagnosis queue automatically, and
  surface open/escalated items in a digest.
- Track record cited in docs: the experience council refused a plan 7 runs
  running and every refusal exposed a real harness defect; the diagnosis council
  escalated 8 runs running, each correct.

## 7. Data ingestion, news pooling, research, RAG  [LIVE]
- Fetch news per-site from RSS, scrape, news-API and news-search; dedupe on
  source URL; triage each item for relevance + credibility against the site's
  vertical/values/audience.
- Deterministically decide whether a site should even *have* a news feed (industry
  → recommendation + keywords + source types).
- **Pool the expensive work once across the 1,000+ portfolio** by modelling each
  news pool as a synthetic site; rank which pooled articles matter per-site *for
  free* using in-DB pgvector + an in-house Ollama embedding (cost scales with
  world-news volume, not domain count).
- **Audience profiling** = a structured, versioned `audience.v1` spec per site
  (who the reader is, position vs sibling domains, editorial directives); the
  who+position feed selection embeddings so near-duplicate domains never render
  identical feeds. (Per-site distinctiveness is treated as *the product*.)
- Shared knowledge base with RAG: chunk → embed → pgvector search (trigram-text
  fallback).
- Research workflow: web search → pick top authoritative URLs → batch-scrape →
  collect snippets as LLM context.

## 8. Web, scraping, browser, archive  [LIVE]
- Web search over an adapter (url/title/snippet/date/source).
- Web scraping single-URL and **batch** (backed by Firecrawl + a Playwright
  provider); Firecrawl-style crawl/map.
- Browser automation (headless desktop profile) for real page interaction/render
  and tool verification.
- Archive research via Internet Archive CDX/Wayback to recover a dead domain's
  old vertical, paths and traffic origins.

## 9. Analytics / traffic  [LIVE / PARTIAL]
- Intent probe: a JS-free capture page (search/category/free-text) that logs
  visitor stated-intent server-side; rank domains by demand (events, 7d/30d
  volume, distinct terms, dominant-cluster share, referer breakdown, recency).
- Per-host visit/event counters exposed at `/stats` behind an internal key;
  passive signals from nginx logs (referer, landing path, 404-intent paths,
  bot classification, status distribution).
- Centralised outbound HTTP-call logging (agent, action, method, URL, status,
  latency) and 24h LLM-call stats (calls, model, latency, in/out tokens by agent).
- Caveat: honest per-visitor measurement needs the proxy to pass real client IPs
  (a known prerequisite, not yet universal).

## 10. Human-in-the-loop (admin dashboard)  [LIVE, one gap]
- React SPA + API gateway over auth-service + core-manager.
- Any agent can pause at a review point *without* suspending the orchestration by
  creating a `needs_human_review` work item with `on_approve` instructions.
- Reviewer queue: sites overview, per-site work items with error previews, review
  flows (placeholder-content, editable checkpoint+approve, retry/resolve),
  "Retry All Failed"; confirm a finding into a follow-up item with edited text /
  severity / notes; edit content directly (HTML/fields save+deploy with
  auto-lock+rerender, or queue an LLM rewrite); spec/Direction editor with
  pin/unpin + propagate; media/asset browser.
- Documented gap: some human-resolution routes were left unwired (a large review
  queue nobody could action) — being addressed.

## 11. Scheduling, ops, deployment  [LIVE]
- DB-driven scheduler: reads `scheduled_tasks`, publishes Kafka triggers on
  `interval_seconds`; a new schedule is a plain INSERT (no redeploy). Optional
  SQL pre-query gates a task and injects the first row's columns as input
  (e.g. rotate through the least-recently-built site). Concurrency groups with
  max-concurrent + timeout; atomic double-fire prevention.
- Work-item queue: agents atomically claim items (optimistic lock) to prevent
  double-dispatch; a verification step re-checks a fix before marking complete.
- Migration runner: applies pending `NNN_name.sql` in order, records each,
  dry-run by default, refuses if DB unreachable, `--record-only` for hand-applied
  files, idempotency lint.
- Static-site deploy: a GitHub Action on push diffs changed domains, syncs each to
  Backblaze B2, purges the matching Cloudflare cache. Platform services deploy via
  per-service Kustomize bases + production overlays + Terraform.

## 12. Dataset extraction (reasoning corpus)  [LIVE]
- Turns the platform's *own* decisions into a trajectory dataset (theory,
  evidence, decision, whether it was right), mined from diagnosis artifacts, LLM
  call logs and orchestration trails; ~820 records / 112 trajectories, bad rows
  flagged not dropped, with provenance so incomparable rows aren't pooled. Ships a
  small hand-graded gold benchmark (eval-only, too small to train on).

---

## External AI models / modalities wired in
- **Text/reasoning:** Anthropic Claude (default `claude-sonnet-4-6`; alias table up
  to `claude-sonnet-5` / `claude-opus-4-8`), incl. extended-thinking/reasoning.
  Provider is swappable per agent (Anthropic ↔ Ollama).
- **Local LLMs + embeddings:** self-hosted Ollama, embeddings via
  `nomic-embed-text` (the only vector modality).
- **Image:** Stability SDXL + Google Banana (`gemini-3-pro-image-preview`, up to 20
  reference images).
- **Web/browser:** web search, web scrape (Firecrawl + Playwright), headless
  browser automation.
- **Remote GPU:** Thunder Compute offload (+ B2 storage) for heavy jobs.
- **⚠ MODALITY GAP — no speech/voice/TTS is wired in.** The "voice" surfaces in the
  code are brand *writing-voice* text checks, not audio. No video assembly, no
  animation pipeline, no code-execution/programming agent beyond the caged
  implementer's git+build flow. Any operation we design that needs narrated video,
  voiceover, or animation requires **adding** those modalities — they are not
  present today.

## Honest status caveats (read before quoting)
- Council-gate is **advisory** — it records verdicts, it cannot block a
  hand-commit. The blocking PR-mode gate is designed but owner-gated.
- Feature-builder **implementer** half built but unproven end-to-end.
- V5 citation-verification is **designed, not built**.
- The domain→live-site build path is the platform's core competency but is also
  its most bug-fought path (delivery/render gaps actively worked).
- Rosters/counts change frequently — the "16 seats", "1,633 concepts", "~820
  records" figures were true at compile time; verify against the live DB before
  repeating.
