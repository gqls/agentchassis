# AUDIT — leopardessconsulting.co.uk: verified facts vs claims

**Date:** 2026-07-09
**Site ID:** `4851f6fc-71cf-4160-a270-e03d6d3e0732`
**Method:** every claim below was checked against code, live Postgres, or an HTTP
response. Nothing here is inferred from the marketing copy itself. Where a claim
could not be substantiated it is marked UNSUPPORTED, and that is a finding, not a
gap to be filled with a better sentence.

This document is the evidence base for the rebuild. The rule for the whole
project: **no claim ships unless it has a row in this table.**

---

## 1. What the platform genuinely does (verified)

| # | Claim on site today | Verdict | Evidence |
|---|---|---|---|
| C1 | Companies House verification pipeline: area discovery → entity scraping → automated verification, with financial enrichment and a name-matching cascade | **TRUE** | `platform/orchestration/actions/companies_house_actions.go` (1054 lines) hits `api.company-information.service.gov.uk`. `companies_house_local_match_action.go` (716 lines) implements the tiered cascade (≥0.90 + geo; ≥0.90 + unique; postcode prefix + ≥0.50; trigram residual → LLM review). Live: `business_intel.businesses` = **2,767 verified**; `companies_house_data` = **937 enriched**; `ch_vet_companies` = **5,798**. |
| C2 | Multi-source news pipeline with LLM credibility scoring, four source types, 6-hour refresh | **TRUE** | `feed_actions.go`: `FetchRSSAction`, `FetchLLMNewsAction` (xAI `/v1/responses` with `web_search`+`x_search` tools, plus OpenAI/Perplexity). `feed_triage_actions.go` writes `relevance_score`, `credibility`, `credibility_reason`, `source_tier`. Live: **5,652 feed items, 4,672 credibility-scored**. `scheduled_tasks` has `content-feed-refresh` at `interval_seconds = 21600`. Source types live: news_search ×13, api_news ×3, scrape ×1, rss ×1. |
| C3 | LLM-driven tool generation: evaluates what tools suit an audience, generates self-contained HTML/CSS/JS, deploys with nav + companion guide, cross-links related pages | **TRUE** | `discovery_checks/check_missing_tools.go` (delegates the choice to `tool-suggester`), `create_tool_component_action.go`, `deploy_tool_action.go` (builds tool page, nav entry, companion guide), `create_tool_cross_link_items.go`. Agent family present: `tool-suggester`, `tool-generator`, `tool-deployer`, `tool-auditor`, `tool-improver`. |
| C4 | Hierarchical agents: DB-defined workflows, sub-agent spawning, stateless horizontally-scaling pods, zero-downtime workflow updates | **TRUE** | `platform/messaging/processor.go` re-reads `agent_definitions.default_config.workflow` **per message** (hence zero-downtime edits). `spawn_actions.go` / `call_agent.go` / `spawn_group.go`. **40 of 148** agent defs call `spawn_agent`. State in `orchestration_states` = **75,061 rows**, so pods hold nothing durable. |
| C5 | Image generation | **TRUE** | `internal/adapters/imagegenerator/dynamic_adapter.go` routes `kind=="icon"` → Banana (Google Gemini `gemini-3-pro-image-preview` default), everything else → Stability **SDXL v1.0**. Banana supports reference images (brand consistency); SDXL's reference-image field is a no-op. |
| C6 | Sites built and operated by the platform | **TRUE, and understated** | `sites` table: **8 deployed** — ai-agent-orchestration.com, dartsonline.com, finetuning.uk, gamesdesign.co.uk, idea.uk, leopardessconsulting.co.uk, robot-hands.com, vonc.com. |

### Caveats that must shape the copy

- **C1 "dissolved companies":** the matcher filters `WHERE company_status='active'`. It
  *excludes* dissolved companies rather than reconciling them. Saying we "handle
  dissolved companies" overstates it.
- **C2 "Cloudflare-hosted":** wrong. GitHub Actions writes to **Backblaze B2**;
  Cloudflare is DNS plus an edge Worker (`scripts/cloudflare/worker.js`). Say what
  is true or say nothing.
- **C4 "each named agent is itself an orchestrator managing teams of specialists":**
  overstated. 40 of 148 spawn sub-agents; the rest are leaf specialists.
- **C6:** these are **our own sites**, not client engagements. They demonstrate the
  platform. They are not a client roster and must never be implied to be one.

---

## 2. Claims that are UNSUPPORTED (must be removed)

| # | Claim | Why it fails |
|---|---|---|
| U1 | "Over 70 specialised AI agents in **8 departments**" | **No department concept exists anywhere.** `information_schema.columns WHERE column_name ILIKE '%department%'` → **0 rows**. No table, enum, or Go constant. The nearest real taxonomy is `agent_definitions.category` (19 values) and `agent_category` (6 values). Neither is 8. The per-department `agent_count` figures in the identity spec (Strategy: 8, Research: 10, Content: 12…) are fabricated. |
| U2 | "Over 70 specialised AI agents" (as a running fleet) | 143 distinct types are **defined** in `agent_definitions`; 56 are `status='active'`. At any moment only ~7 agent types run as pods. True as a catalogue count; false as a fleet. |
| U3 | Case studies attributed to clients: "Veterinary Data Aggregator", "Multi-Site Content Platform", "Industry News Aggregation", "Interactive Tool Platform" | These are **the platform's own subsystems**, renamed to look like client engagements. The engineering is real (C1–C4). The clients are not. |
| U4 | "Six production sites deployed and self-maintaining" for a client | Eight sites exist, and they are ours (C6). Both the number and the framing are wrong. |
| U5 | Leadership team: "Peter Grenfell, Founder & Principal Consultant" | `identity.key_people` = `[]`. The bio is written in exactly the register the voice spec bans ("digital transformation", "agile innovation"). **Owner confirmed 2026-07-09: invented. Delete.** |
| U6 | `/how-we-work.html` (live): *"Playwright-based agents handle dynamic JavaScript rendering, session management, and anti-bot navigation. Supervisor agents detect rate-limit signals and reroute traffic across proxy pools."* | **Playwright is not used.** It appears in exactly two places: a comment in `internal/adapters/webscrape/adapter.go:130` — *"Could add other providers here (Playwright, Puppeteer, etc.)"* — and a commented-out line in a dev configmap. The real scraper is **Firecrawl**. `proxy_pool`/`proxypool`/`anti-bot`/`antibot` → **0 matches** in the entire Go codebase. Every specific in that sentence is fabricated. |
| U7 | `/about.html` (live) lists **two AI agents as team members**: "Orchestration Agent: Operations" and "Orchestration Agent: Research & Intelligence", each with a `photo` filename, under the heading "The People Behind the Platform". | They are not people. The section also asserts *"the orchestration platform powering our client deployments"* and *"our published case studies"* — both refer to things that do not exist (U3). The intro also contains a competitor swipe: *"not consultants who advise from a distance"*. |
| U8 | The three `leadership-team` `photo` files | `founder-principal-engineer.jpg`, `orchestration-agent-operations.jpg`, `orchestration-agent-research.jpg` all return **HTTP 404**. The About page has three broken portraits of people who do not exist. |
| U9 | `/how-we-work.html` header: *"Seventy-plus agents organised across eight functional departments"* | Rendered into `page_components.rendered_html` (7 KB) with **`content_data IS NULL`** — the copy is baked into the artifact, so fixing the spec alone will not change the page. Departments do not exist (U1). |

---

## 3. Live defects (verified by artifact, not by report)

| # | Defect | Evidence |
|---|---|---|
| D1 | **The logo and hero image are dead links.** | All 3 rows in `assets` for this site store a **presigned S3 URL with `X-Amz-Expires=604800`** (7 days), signed `20260128` and `20260204`. `curl` → **HTTP 401 `Request has expired`**. `storage_path` is empty on all three, so `asset-deployer` cannot resolve a git path either. These rows predate commit `84f07d38` ("resolve deployed git path, not presigned S3 url"). They need regenerating, not repairing. |
| D2 | **Nine pages have zero sections.** | `/ai-readiness-quiz.html`, five `case-study-*.html`, `/for-engineering-leaders.html`, `/guides/llm-cost-calculator-guide.html`, and others return a page shell with no content components. Six are `needs_rebuild`. |
| D3 | **Nav labels are raw `<title>` strings.** | e.g. `nav_label` = `"Multi-Stage Data Pipeline with Companies House Verification \| Leopardess Consulting"`. Should be a short label. |
| D4 | **`sites.tagline` is stale and uses retired language.** | Row reads `"Agile AI Agents: Move Swiftly into Digital Transformation with Grace and Precision"`. The `identity` spec explicitly retires *"digital transformation"* and *"grace and precision"*. The DB row was never updated when the spec was. |
| D5 | **Near-total absence of imagery.** | `assets`: leopardess = **3** (all broken). robot-hands.com = **41**. No `site_plan_imagery` rows. No favicon, no OG cards, no card images, no infographics. |
| D6 | **`design_intent` contradicts itself.** | `style_direction: "professional-dark"` and a `colour_mood` describing "deep charcoal and near-black backgrounds", but `color_scheme.background = "#ffffff"` and `text = "#333333"`. The rendered site follows the light `color_scheme`; the prose describes a dark site that does not exist. |
| D7 | **`brand_assets`, `deploy_config`, `logo_url`, `github_repo` are all empty** on the site row. Two sibling sites (finetuning.uk, gaswholesalers.com) do set `logo_url = /assets/images/logo.png` — that is the working convention to follow. **RESOLVED 2026-07-10**: `logo_url` now set. |
| **D8** | **PLATFORM-WIDE: image asset deployment is broken. `deploy_image_asset` can never succeed on `agent-chassis`.** | `platform/agentbase/agent.go:294` only builds a storage client when `IMAGE_BUCKET` is set, logging *"Storage client not configured (IMAGE_BUCKET not set)"*. The `agent-chassis` deployment has `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY` but **no `IMAGE_BUCKET`, `S3_ENDPOINT` or `S3_REGION`** (those are literal env vars on `image-generator-adapter`, and live in the `storage-config` configmap + `personae-storage-secrets`, neither of which `agent-chassis` references). So `deploy_image_asset_action.go:147` returns `storage client not available`. **Verified by running it:** triggering `asset-deployer` for this site produced `orchestration_states.status = FAILED`, `error = "step deploy_asset failed: failed to execute action deploy_image_asset: storage client not available"`. |

### D8's blast radius (measured, 2026-07-10)

| Fact | Value |
|---|---|
| Active assets platform-wide | **102** |
| …with a durable git path (`/assets/…`) | **19** (18 of them `idea.uk`, 1 the leopardess logo landed today) |
| …with an **expiring presigned URL** | **83** |
| …with a non-empty `storage_path` | 19 |

`/assets/images/logo.png` returns **404 on five of six sites checked** (robot-hands.com,
vonc.com, idea.uk, dartsonline.com, gaswholesalers.com); only finetuning.uk serves one,
from an older deploy route.

**This directly contradicts the imagery programme's premise.**
`PLAN_imagery_best_in_class.md` states the imagery pipeline is "verified end-to-end …
`asset-deployer` commits optimised images to the site's git repo." It does not, and
cannot, in this cluster. **robot-hands.com — the imagery testbed — has 34 active assets,
every one of them a presigned URL that expires 7 days after signing, and its logo 404s.**
The images on that site are on a timer. This is the same failure that killed the
leopardess logo (D1); it was never a leopardess-specific problem.

**The fix** is one of:
1. Add `IMAGE_BUCKET`, `S3_ENDPOINT`, `S3_REGION` to the `agent-chassis` deployment
   (it already has the AWS credentials) — e.g. `envFrom` the `storage-config` configmap,
   noting its keys are `image_bucket` / `S3-ENDPOINT`, which do **not** match the env
   names the Go code reads, so explicit `env:` entries are needed. Smallest correct fix.
2. Or give `asset-deployer` a k8s config that schedules it on a pod that has storage.

**Workaround used today** (`scripts/commit_brand_assets.sh`): `deploy_image_asset`'s only
job after optimising is to publish a commit message to the git-adapter
(`system.adapter.git.requests`). With correctly sized files already in hand, we send that
same message directly. Same adapter, same repo, same contract — no storage client needed.
This is a workaround for one site's brand assets, **not** a fix for the pipeline.

---

## 4. Spec-vs-brief conflicts (owner decisions required)

| # | Existing spec says | The brief says | Status |
|---|---|---|---|
| X1 | `design_intent.avoid`: *"Literal leopard or animal imagery — the brand metaphor is applied through language, not visuals."* | *"We want a great leopardess logo and branding consistent across the site."* | Brief overrides. Degree of literalness is a design choice. |
| X2 | `identity`: *"Non-technical SMB buyers are explicitly out of scope and no page should contain copy that targets or accommodates them."* Audience = CTO / VP Eng. | *"Target an intelligent future buyer of AI to help their business but they may not know much about what AI can do for them… they may have heard the hype and are not convinced."* | **Direct contradiction.** This is a repositioning, not a restyle. |
| X3 | `identity.tagline`: *"…Deployed on Kubernetes and Kafka, Not Prototyped on Wishful Thinking."* | *"no negative framing, no claims that are too bold… keep our promises small but deliverable and not join the hype fest."* | Tagline is a swipe at competitors and is negative framing. Must be rewritten. |
| X4 | `portfolio.case_studies` present platform subsystems as client work (U3). | *"no inventing projects or staff, but we can say what we might be able to do provided it is not too pie in the sky."* | Reframe from "case studies" to demonstrable capability, honestly labelled. |

---

## 4b. Positioning research — data-sovereignty / model-choice angle (2026-07-10)

Prompted by a specific problem the owner hit with a real prospective legal client:
data-confidentiality concerns ruled out sending their material to a foundation LLM.
Investigated whether the platform can honestly support "route only the steps that
need it away from third-party APIs."

| # | Claim | Verdict | Evidence |
|---|---|---|---|
| P1 | Model/provider can be chosen per workflow **step**, not just per agent, with no new code | **TRUE** | `ExecuteLLMPromptAction` (`ai_actions.go`) resolves `ai_service` in a three-tier lookup, tier 2 being `workflow.steps[<step>].config.ai_service`. Proven tooling exists to change it live: `swap_agent_model()` (migration 083/`021_model_swap_and_rollback.sql`), actively used in production code paths. |
| P2 | A self-hosted step genuinely never leaves the cluster | **TRUE** | `ollama-adapter` runs the stock `ollama/ollama` image, `ClusterIP` only, no internet exposure. `platform/aiservice/ollama.go` calls it in-cluster. |
| P3 | Only two text-generation providers work end-to-end today | **TRUE, and a real constraint** | `createAIClient` switch: `anthropic` and `ollama` only; `openai` is a stubbed error; nothing else is wired for text. "Mistral" is not a separate provider — it is the open-weight model name (`mistral-small3.1`) run *through* the same self-hosted Ollama pod. Imagery is broader: Gemini + Stability SDXL both work (routed by kind, not config). News has its own separate hand-rolled path (xAI, OpenAI, Perplexity) that bypasses the generic `AIService` entirely. Firecrawl (web-scrape/search) may have a model behind its API; out of scope — we don't control or need to describe that layer. |
| P4 | The larger self-hosted model (llama3.3:70b) is deployed and inference-ready | **NOT YET** | `model_lifecycle.training_runs`: one `complete` run (2026-06-03→04), several `failed`/stuck `pending` — real but experimental, not routine. GPU provisioning is genuinely dynamic (`thunder_instances`, ThunderCompute, decommissioned after each run). **No `agent_definitions` row points at `llama3.3:70b`** — trained and tested, never used for production inference. TODO logged: `docs/agent_docs/docs024_key_docs_latest/009_model_infrastructure.md` "Future" checklist. |
| P5 | Client data is isolated from other clients/sites on this platform | **NOT TRUE TODAY, but a real capability to offer** | Single shared Postgres (no row-level security anywhere in the schema), single shared Kafka cluster, single shared `ollama-adapter` pod — separated only by a `site_id` column in shared tables. **However**: genuine physical isolation is achievable and cheap to reason about — standing up a dedicated cluster per client (the owner's suggestion: Rackspace or similar) gives total isolation, and the scaffolding for cross-cluster dispatch already exists (`remote-job-spawner`, `DispatchAgentAction`), just not exercised in production yet. **Position this as a capability we build with a client, not a standing guarantee.** |
| P6 | The stack is UK/EU-hosted end to end | **NOT TODAY** | Compute: Rackspace, UK-hosted (owner-confirmed). Storage: Backblaze B2, `us-east-005` — US. Models: Anthropic (US) and Google Gemini (US) for the cloud paths; the self-hosted path is wherever the cluster sits (currently UK, via Rackspace). So "UK-based company, UK-hosted compute" is true today; "all data stays in the UK" is not. Full detail in memory: `uk-sovereign-stack-exploration` (deferred to a separate chat, not part of this rebuild). |

**Net honest claim for the site:** *"We can architect a workflow so the steps that
touch your data run on infrastructure you control, and only the steps that don't
need to leave that boundary call out to a foundation model."* Not a standing
guarantee of isolation from other clients (P5) or of UK-only data residency (P6) —
both are real, buildable, and worth offering as part of a scoped engagement, not
claiming as already shipped.

## 5. Open questions — all resolved 2026-07-09/10

1. ~~Is Peter Grenfell real?~~ No. Invented, deleted (A1). The `team[0]` background
   (worldsoccernews.com, ex-Bumble) is the owner's own.
2. ~~Audience pivot?~~ Confirmed (A2).
3. ~~Logo literalness?~~ Stylised head profile (A3).
4. ~~Palette?~~ Dark chrome, light reading surfaces (A4).

See `RUNNING_NOTES.md` decision log for A1–A9 and the full reasoning.

### Resolved: the worldsoccernews.com figure (2026-07-10)

Web research (Wayback Machine CDX) found no independent ranking evidence anywhere —
no Alexa/SimilarWeb history, no press, no forum mentions — for "third busiest sports
site." The site's own "About Us" page separately states "15 million page impressions"
(unchanged across 2000–2012 snapshots).

**Owner's decision, overriding my recommendation to use only the sourced figure:**
state ~12 million unique users a month at peak (a different, owner-recalled metric,
not the site's self-published impressions figure), note it was covered at the time in
a media trade magazine and referenced in Microsoft's own advertising material, and
publish the "third busiest" recollection **explicitly labelled as unproven**, with a
real boundary for credibility: bigger than the BBC's sports coverage at the time,
smaller than ESPN's Soccernet. This is a deliberate, informed choice by the primary
source about their own history — not a claim I generated or independently verified.
It is drafted into `specs/identity.json` with the hedge intact, which is the pattern
this whole rebuild is built on: say the true thing, and say plainly when a part of it
can't be proven. The 12M-uniques, magazine and Microsoft details are owner-asserted
and not independently checked by me; flagging that here for completeness, not to
relitigate a decision the owner has already made.
