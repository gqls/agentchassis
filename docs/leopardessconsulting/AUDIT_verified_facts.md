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
| U9 | `/how-we-work.html` header: *"Seventy-plus agents organised across eight functional departments"* | Departments do not exist (U1). **Section DELETED 2026-07-10.** *Note on the NULL-`content_data` observation:* it was true when recorded, and an agent backfilled it at 10:29 on 2026-07-10 — `rerender_page_sections` escalates a section with missing `content_data` to a heavy path that backfills it. So the "baked into `rendered_html`" concern was real but self-healing. |
| U10 | `/who-we-help.html` FAQ asserted per-agent **token budgets**, **circuit breakers**, routing to a cheaper model tier, **Helm charts**, deployment into **AWS/GCP/Azure**, and per-agent **least-privilege IAM**. | None appear in the codebase. A circuit breaker is in fact an explicit *unwired* `TODO(provider-circuit-breaker)` in `dynamic_adapter.go`. **FAQ replaced 2026-07-10.** |
| U11 | Homepage `system-stats` published **"70%" deployed agents, "3ms" orchestration model, "Minutes+" deploy time, "99.9x" uptime**. | The `stat{N}_suffix` fields were misaligned against `stat{N}_value`, so the section rendered nonsense — *and* the uptime target was an unsupportable claim the plan bans. **Rewritten 2026-07-10** with four real, dated figures and empty suffixes. |

**Fabrication sweep, 2026-07-10:** every `page_components.content_data` on the site was
checked against 16 patterns (Peter Grenfell, eight departments, 70+, Seventy-plus,
Playwright, proxy pool, anti-bot, Orchestration Agent, Veterinary Data Aggregator,
Multi-Site Content Platform, digital transformation, Grace and Precision, 99.9, circuit
breaker, Helm chart). Result: **CLEAN.** `content_data` is the render source, so this is
the surface that matters; `rendered_html` regenerates from it at L9.

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
| **D8** | **CORRECTED 2026-07-10 — I overstated this. The image pipeline is NOT broken; images serve durably. My "83 assets on a 7-day timer / testbed logo 404s" claim was wrong.** | See the correction below. What is genuinely true is much narrower. |

### D8 — the accurate version (after re-investigation)

I made two mistakes in the first version of D8 and want them on the record, because
the whole project runs on "verify by artifact, never by report" and I violated it.

**Mistake 1 — a meaningless test.** I tested `/assets/images/logo.png` on six sites and
saw 404s. But **those sites have no `logo` asset at all** (robot-hands' assets are
`hero`, `hero_home`, `icon_*`, …; no `logo`). I was curling a path nothing was ever
deployed to. When I test the paths that *do* exist — `hero.jpg`, `hero-home.jpg`,
`icon-cycle-time.jpg` — **they all return HTTP 200.** Robot-hands' images work.

**Mistake 2 — I misread the presigned URL as a rendering bug.** It isn't. The
render-time resolver `plan_sections` (`plan_sections_action.go:193, 260, 290`) emits
`storage.DeployedWebPath(asset_key, purpose)` — the durable git path — and **never reads
`assets.url`**. Verified against the live HTML: **zero `X-Amz` presigned URLs appear in
robot-hands' or leopardess' rendered pages.** The presigned URL sitting in the
`assets.url` column is a stale *source handle*; it is bypassed at render. The 83
presigned rows make the table look alarming but do not break a single page.

**What is actually true (the narrow, real findings):**
1. **`deploy_image_asset` fails when run inline on the base `agent-chassis` pod**, which
   has no `IMAGE_BUCKET` (`agentbase/agent.go:294`, "Storage client not configured").
   I reproduced this by hand-triggering `asset-deployer` standalone. **But that is not
   how the pipeline runs it.** `spawn_actions.go` injects `IMAGE_BUCKET` + storage creds
   into spawned storage-enabled agents (`isStorageEnabledAgent`, incl. `asset-deployer`),
   and `107_image_build_handler.sql:725` documents exactly this: the base chassis "by
   design does NOT carry storage env vars," so deploys run in a **spawned** asset-deployer
   that gets them injected. Proof it works: robot-hands (34 assets, 2026-07-09) and
   idea.uk (18, 2026-06-21) were generated and committed this way and **serve 200 today.**
   My standalone trigger failed only because it skipped the spawn-time injection.
2. **`assets.url` is cosmetically stale for 83 rows** (presigned instead of git path).
   Fixable by generalising the `w9_04` backfill flip (which idea.uk already got) to all
   sites — `url = '/assets/images/'||replace(asset_key,'_','-')||'.<ext>'`. **Cosmetic,
   not urgent**, since render already ignores `url`.
3. **Leopardess DOES have five genuinely-broken images**, but for a *third* reason,
   unrelated to both of the above: its homepage `case-studies-grid` references
   `/assets/images/case-study-01.jpg … 05.jpg`, and those files were **never generated or
   committed** (they 404). They are part of the fabricated case-studies content that L5
   removes anyway. The three original brand assets (D1) were separately dead. So
   leopardess's image problem is "assets never properly produced," a content problem, not
   a platform pipeline bug.

**Net:** the imagery pipeline works. There is no platform-wide emergency. The imagery
programme's "verified end-to-end" claim stands. I withdraw the alarm. The one true
platform nit is the cosmetic `assets.url` staleness (finding 2).

**Note on the logo I deployed today:** using `commit_brand_assets.sh` to send the commit
straight to the git-adapter was still the *right* call — not because the pipeline is
broken, but because I was injecting a **specific owner-approved image** rather than
generating one, so there was no generation step to spawn. The result serves correctly.

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
can't be proven.

**Source named 2026-07-10:** the owner identifies the publication as **New Media Age**
(the UK digital-media trade magazine). The copy now names it, which turns the
12-million-uniques figure from an assertion into something a reader could in principle
check against a back issue. I have not seen the article myself — the claim is
*attributed*, not independently verified, and that is the honest standing of it. The
"third busiest" ordinal remains explicitly labelled as recollection.

---

## Re-measurement 2026-07-31 — every figure on `/services.html`, checked before rewriting it

Occasion: the owner asked for the two `/services.html` blocks to become carousels.
Rewriting a sentence means owning the numbers inside it, so every figure on that page
was re-derived from the live system first (CLAUDE.md: *ground every figure against the
live system before repeating it from another doc*). All queries below were run against
`postgres-clients-0 / clients_db` on **2026-07-31**.

| Figure as published | Live value 2026-07-31 | Query / evidence | Standing |
|---|---|---|---|
| "more than 90,790 orchestration state records" | **row count 2,364**; cumulative sequence **6,996,354** | `SELECT count(*) FROM orchestration_states` → 2,364, window 07-13→07-31. `SELECT last_value FROM orchestration_state_audit_id_seq` → 6,996,354 | **MISLABELLED — see the correction below** |
| "more than 2,000 business records" verified | **3,419** | `SELECT count(*) FROM business_intel.businesses` | TRUE and understated; refreshed |
| "enriched 937 with filed accounts" | **937** | `SELECT count(*) FROM business_intel.companies_house_data` | TRUE, unchanged |
| "6,262+ feed items collected" | **7,990** (since 2026-03-27) | `SELECT count(*) FROM content_feed_items` | TRUE and understated; refreshed |
| "5,228+ credibility-scored" | **6,794** | `… WHERE credibility IS NOT NULL` | TRUE and understated; refreshed |
| "six-hour refresh cycle" | **21600s, enabled, last completed 2026-07-31** | `SELECT interval_seconds, enabled, last_completed_at FROM scheduled_tasks WHERE name='content-feed-refresh'` | TRUE, re-verified live |
| "Nine sites built and operated" | **15** (14 `deployed` + 1 `active`), all with a domain | `SELECT count(*) FROM sites WHERE status IN ('deployed','active')` | TRUE and understated; refreshed |
| "157 definitions as of 2026-07-16" | **190 live rows**, 184 distinct types | `agent_definitions WHERE COALESCE(is_snapshot,false)=false AND deleted_at IS NULL` | grew; refreshed |
| "60 definitions are currently active" | **185** | same filter `AND is_active` | **[UNRECONCILED]** — see below |
| "40 spawn sub-agents" | **42** | `steps` is an OBJECT keyed by step name, not an array; `jsonb_each(...)` where `step->>'action' IN ('spawn_agent','spawn_group','spawn_agent_k8s')` | TRUE; refreshed |
| "Model and provider are selectable per workflow step" | **TRUE, richly** | `ai_service` on 100+ steps carries its own `{model, provider, max_tokens}` — `claude-haiku-4-5`, `claude-sonnet-4-6`, `claude-sonnet-5`, `claude-opus-4-6`, per-step `max_tokens` from 150 to 32000 | TRUE, re-verified |
| "Anthropic Claude, OpenAI, Gemini, **Mistral**, xAI, and Perplexity are all available as selectable providers" | **THREE work: anthropic, ollama, gemini** | `platform/aiservice/factory.go:24-35` — the single switch every LLM step goes through. `openai` returns `"OpenAI provider not yet implemented"`; `xai`/`perplexity`/`mistral` are not in the switch and fall to `default` → error | **FALSE as published — see CORRECTION 2** |
| Every configured step's provider | **`anthropic`, 100% of them** | no step or definition config carries any other LLM provider value | context for the row above |

### CORRECTION 1 — the 90,790 figure is a pruned table's row count, published as a cumulative total

`orchestration_states` is **pruned hourly at 24h**
(`platform/orchestration/actions/diagnose_dormant_agents_action.go:15,349`). The
published sentence — "has produced more than 90,790 state records ... every one of them
readable after the fact" — reads as a cumulative to-date total and as a durability
promise. Neither survives the pruning: the table held **2,364** rows when re-measured.

What *is* cumulative and evidenced: `orchestration_state_audit` is written by a database
trigger (`orchestration_state_audit_trigger AFTER UPDATE ON orchestration_states`), and
its identity sequence has never been reset. `last_value = 6,996,354` against a live row
count of ~100,400 spanning only 07-29→07-31 (`min(id)=6,895,954`) proves both that
~7.0M transitions have been recorded and that the table itself keeps a rolling window
of roughly three days. **So the durable figure is a rate plus a window, never a total** —
which is the fleet lesson about retention-clocked history tables, arrived at here
independently.

This drift was **already detected by the platform's own claims layer** on 2026-07-26
and routed to `needs_human_review`: `bugs_open/091` records
`C4-orchestration-state-records (live 1,900 vs published 90,790)`. It was never
resolved, so the stale figure stayed live for five more days. The claims layer worked;
the draining of what it found did not.

### CORRECTION 2 — the provider list is wrong, and my first version of THIS ROW was wrong too

> **CORRECTED 2026-07-31, before anything shipped.** My first draft of this section said
> *"five hosted providers plus a self-hosted path"*, derived from
> `grep -ioE 'case "(anthropic|openai|gemini|mistral|xai|perplexity|ollama)"'` returning
> six names. **That was a wrong call, and it would have replaced one overclaim with
> another.** What caught it: this workstream's own `RUNBOOK.md` **landmine 12**, which
> already said in plain words that only two text providers worked end to end. Reading the
> code the landmine pointed at settled it.
>
> **A `case` arm is not a working provider.** `platform/aiservice/factory.go:24-35` is the
> single switch every LLM step passes through, and it is decisive:
>
> | provider | factory behaviour |
> |---|---|
> | `anthropic` | `NewAnthropicClient` — works |
> | `ollama` | `NewOllamaClient` — works (self-hosted, in-cluster) |
> | `gemini` | `NewGeminiClient` — works (added by `bugs_closed/107`+`110`; this is what has changed since landmine 12 was written, and the landmine needs that one amendment) |
> | `openai` | `return nil, fmt.Errorf("OpenAI provider not yet implemented")` — **a stub that errors** |
> | `xai`, `perplexity`, `mistral` | not in the switch at all → `default` → `unsupported AI provider` |
>
> The `case "xai"` / `case "perplexity"` / `case "openai"` arms my grep found are all in
> **`platform/orchestration/actions/feed_actions.go`** (lines 416-457, 737-752) — a
> separate, purpose-built news-and-signal search path that calls those vendors' *search*
> APIs directly. They are real, and they are not general-purpose LLM providers.

So the published sentence is **false on two counts**: OpenAI is named as available when it
is a stub that raises an error, and Mistral/xAI/Perplexity are named as selectable
providers when the LLM factory rejects all three.

Mistral is nonetheless real here — as `mistral-small3.1` pulled by the cluster's own
`ollama-adapter` and `ollama-eval` deployments (24B, Q4, CPU-only, no GPU on any node;
`deployments/kustomize/services/ollama-adapter/base/deployment.yaml:34`,
`platform/orchestration/actions/refresh_product_specs_action.go:64`). It is a **model on
the self-hosted path**, not a provider beside Anthropic.

**The honest form, now published:** three providers work for generating text — Anthropic,
Google Gemini, and a self-hosted Ollama path that keeps inference inside the cluster — and
the claim is about *those three*. Nothing on the page names OpenAI, xAI or Perplexity as an
LLM provider any more. `RUNBOOK.md` landmine 12 stands, with Gemini added.

### [UNRECONCILED] — "60 active" against a live 185

Not asserting a cause. Candidates not investigated: `is_active` semantics changed; the
2026-07-16 count applied a filter not recorded alongside it; or ~125 definitions were
genuinely activated in the fortnight since. The 2026-07-09 audit for this site recorded
"143 agent defs, 56 active", which is consistent with the 60 rather than the 185, so the
step change is real and recent. **Flagged, not explained** — the published copy now
quotes the figure I measured myself with the query written beside it, and does not
inherit the 60.

### Also removed while here: a phantom tool URL, live on this page since 2026-07-18

`/tools/tool-monitoring-coverage-gap-finder.html` was linked from the `services-grid`
block (inside `features[2].description`, as real markup) and named again in prose in the
`info-card-grid` block (`cards[4].body`). **No such page has ever existed** — it is
absent from `pages` and returns 404. It is the same invented URL punch-list item 3
recorded the owner clicking on 2026-07-18; two further instances survived that clean-up
because they were inside body prose rather than a `link_url` field.
