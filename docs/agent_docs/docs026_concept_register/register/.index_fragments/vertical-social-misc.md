| VET-001 | Vet med pricing pipeline (discovery → scrape+evidence → export) | deployed | 3-stage business-intel pod pipeline: URL discovery, price collection+evidence, JSON export | vet-med-pricing.md |
| VET-002 | Unified polymorphic products/product_prices schema (kind discriminator) | aspirational | Unifies business_prices+product_prices under products via a kind column | vet-med-pricing.md |
| VET-003 | business_prices deprecation migration pattern | aspirational | Phased table retirement: COMMENT ON TABLE deprecated, drop only after Go cutover | vet-med-pricing.md |
| VET-004 | vetcomparison.uk V1 rebuild scope | aspirational | Narrow relaunch: medicine search, vet directory, news, guides; no price-panel yet | vet-med-pricing.md |
| VET-005 | LLM-driven content_features recommendation | aspirational | Moves news/tools/guides decision from hardcoded Go map into classifier LLM prompt | vet-med-pricing.md |
| VET-006 | vet-json-exporter / vet-export-orchestrator agent pair | aspirational | New wrapper pair for vet-practice service prices, modelled on med-json-exporter | vet-med-pricing.md |
| VET-007 | Vet med pricing schema (products / retailers / listings / snapshots) | deployed | 4 tables + matview: med_products, med_retailers, med_retailer_listings, snapshots | vet-med-pricing.md |
| VET-008 | Med scrape evidence store | deployed | One row per page fetch: markdown, content hash, variants accounting, 90-day retention | vet-med-pricing.md |
| VET-009 | Med URL discovery via Firecrawl /map | partial | Site-wide product-URL discovery alongside category-page crawling | vet-med-pricing.md |
| VET-010 | Configurable med price JSON export to site repos | deployed | Generic export action serves many sites via config, commits JSON into site git repos | vet-med-pricing.md |
| VET-011 | thunder-reaper + cost gate (spend backstop) | deployed | GPU instance uptime cap + spend gate; miscategorized here, belongs with Thunder infra | vet-med-pricing.md |
| CH-001 | Companies House enrichment pipeline (bulk collect → match → review → detail fetch) | deployed | 5-stage ch-* agent chain; 634/5,780 (23.2%) confirmed matches as of Mar 2026 | companies-house-enrichment.md |
| CH-002 | Vertical profile registry (generic-words/keywords/suffixes per industry) | deployed | Matching heuristics live in a Go registry keyed by vertical_slug — verticals are config | companies-house-enrichment.md |
| CH-003 | Companies House enrichment with succession-risk signals | deployed | Schema: financials, officers/PSC, owner-age/succession-risk derivation | companies-house-enrichment.md |
| CH-004 | Companies House matching cascade (revised 7-tier signal architecture) | partial | v2 plan: 7 tiers incl. website-scrape and corporate-group mapping, targets 70-90% match | companies-house-enrichment.md |
| CANB-001 | Canine biology knowledge tree (1M-agent demo) | aspirational | 7-level, ~1M-agent swarm building a citable Labrador knowledge tree; demoted to showcase | canine-biology.md |
| CANB-002 | Canine-biology per-vertical knowledge + LoRA project | partial | RAG-seeding (chunks→Ollama embed) plus text/image LoRA fine-tuning for vet vertical | canine-biology.md |
| CANB-003 | Interactive Biological Explorer + experiment engine (aspirational vision) | abandoned | Three.js/Neo4j explorer + ODE experiment engine; dropped for the practical RAG plan | canine-biology.md |
| BIC-001 | Business-intel sweep/verify collection pipeline (vet-intel) | deployed | Area-sweep → collection_tasks → batch-verify pipeline building the vet business directory | business-intel-collection.md |
| SOC-001 | The Forge — AI-seeded community knowledge platform | abandoned | Predecessor to Spark: AI drafts, humans validate/fork/improve; parked concept | social-media.md |
| SOC-002 | Spark — AI game-master social platform (core concept) | partial | AI as producer not performer; opinion-first provocation game; v1 live on vonc.com | social-media.md |
| SOC-003 | Arena + Stage dual modes and their mechanic families | aspirational | Competitive Arena vs showcase Stage, feeding each other in a content flywheel | social-media.md |
| SOC-004 | Rooms-not-feeds architecture and the engagement-depth spectrum | aspirational | Anti-feed design: Lobby/Floor/Gallery zones, ephemeral challenges, Moments | social-media.md |
| SOC-005 | Behavioural archetype system + Daily Gauntlet | partial | 8 archetypes earned via a 5-provocation daily quiz; archetype hub live 2026-07-12 | social-media.md |
| SOC-006 | Cold-start design: AI sparring partner and solo-first completeness | aspirational | No-signup provocation+AI-sparring first 10 seconds; complete for a lone user | social-media.md |
| SOC-007 | Provocation engine — layered content production architecture | aspirational | 6-layer pipeline: raw feed → framing → curation → mashup → serialisation → niche | social-media.md |
| SOC-008 | AI cost architecture: fixed background vs per-user scaling | aspirational | Content-gen cost is fixed (~£5/day); only sparring/scoring scales per-user | social-media.md |
| SOC-009 | Content-first launch strategy for Spark (vonc.com as destination) | partial | Launch a content destination, not a social platform; provocations as SEO pages | social-media.md |
| SOC-010 | Motivation hierarchy and designed user journey | aspirational | 4 motivation tiers mapped to a staged first-5-seconds-to-month-6 user journey | social-media.md |
| SOC-011 | Games and daily-puzzle retention ecosystem | aspirational | Flagged expansion: Wordle-style daily games generated from scraping output | social-media.md |
| SOC-012 | Spark revenue model | aspirational | £3-5/mo subscription + meritocratic brand sponsorship + revenue share, no pay-to-win | social-media.md |
| SOC-013 | Vertical integration of Spark mechanics into domain sites | aspirational | Same mechanics re-flavoured per vertical (vet, finance, fashion, food) | social-media.md |
| VONC-001 | Spark daily-provocation product (vonc.com) | partial | One provocation/day + Gauntlet + Archetype; landing page IS the product | vonc.md |
| VONC-002 | Phase-3 provocation pipeline (automated provocations.json emission) | aspirational | Missing daily generator for provocations.json; still hand-committed as of 2026-07-11 | vonc.md |
| VONC-003 | provocations.json data contract (today / lobby / arena / archive) | deployed | Versioned JSON feed contract for Spark's runtime-fill sections; v3 live | vonc.md |
| VONC-004 | provocation-card component (daily hero card) + mini-lobby trim | partial | Daily hero card runtime-filled from JSON; mini-lobby trim blocked on bundle verdict | vonc.md |
| VONC-005 | lobby-grid arena component (six-room grid) | deployed | 6-card Arena grid runtime-filled from `arena`; reference loader-builder implementation | vonc.md |
| VONC-006 | brief-explanation static explainer (regeneration, not a loader) | deployed | Stable "how Spark works" content fixed by build-time regeneration, not a JS loader | vonc.md |
| VONC-007 | provocations-archive-list component + provocations archive page | deployed | Runtime-fill archive page with clone-template rows and a visible empty state | vonc.md |
| VONC-008 | Option 1 — build-time static content for the daily shells (rejected alternative) | abandoned | Rejected fix that would have frozen provocations permanently at build time | vonc.md |
| VONC-009 | vonc.com Spark v1 site (the live testbed) | deployed | 8-page v1 build: index, archive, about, contact, archetype hub, tools | vonc.md |
| VONC-010 | Archetype hub built with existing machinery (entity pages + query-resolved grid) | deployed | 8 entity pages + query-resolved grid; fixed a zero-archetypes page_type bug | vonc.md |
| CHAT-001 | site_chat_turns table (per-domain chatbot turn logging) | aspirational | Per-turn PII log, separate from llm_call_log; migration number disputed (046 vs 086) | site-chatbot.md |
| CHAT-002 | Site chatbot edge worker (synchronous, not an orchestrated agent) | aspirational | Deliberate exception to "every agent is an orchestrator": sync SSE edge handler | site-chatbot.md |
| CHAT-003 | Build-time context pack (per-domain bounded context) | aspirational | Per-domain JSON pack (identity, scope, grounding chunks, limits) built at install time | site-chatbot.md |
| CHAT-004 | site-chat-installer orchestration (install_chat maintenance task) | aspirational | 3 sub-agents: chat-context-builder, chat-widget-installer, chat-route-registrar | site-chatbot.md |
| CHAT-005 | Provider-agnostic worker (deps adapters) | aspirational | handleChat core + ContextStore/LLMClient/TurnSink adapters; Cloudflare-first, portable | site-chatbot.md |
| CHAT-006 | Three-layer bounding (retrieval / prompt / operational) | aspirational | Retrieval/prompt/operational bounding decomposition to stop chatbot topic drift | site-chatbot.md |
| CHAT-007 | Isolated chat environment (satellite; load/hack/bug vectors) | aspirational | Satellite infra severing load/hack/bug vectors from core; Option X vs Y undecided | site-chatbot.md |
| CHAT-008 | Simple paid multi-domain chat (freemium + day-pass) | aspirational | Fast-lane paid chat: stateless signed entitlement token via Stripe guest-checkout | site-chatbot.md |
| CHAT-009 | Chat lanes (fast/slow/job) + warm-adapter maturation | aspirational | Fast/slow/job lane split; spawned-agent-to-warm-adapter maturation path | site-chatbot.md |
| SAAS-001 | Isolated chat/satellite architecture ("Y-copy") for SaaS build isolation | aspirational | Same satellite architecture as CHAT-007, escalated to a build-as-a-service framing | saas-isolation-architecture.md |
| SAAS-002 | Conversational build-intake via briefing-agent chat | aspirational | briefing-agent chat intake hands off to intake-orchestrator to kick a build | saas-isolation-architecture.md |
| EMAIL-001 | Operator email identity: leopardess.uk + deterministic per-site addresses | partial | One operator domain fronts all sites' mail; deterministic address encoding | email-infrastructure.md |
| EMAIL-002 | Transactional email sending realities (587-only, relay filtering, SES + DKIM) | deployed | Hard-won SMTP truths: 587-only, MailChannels blocks, dedicated SES sender adopted | email-infrastructure.md |
| ADM-001 | Admin dashboard + nginx gateway architecture | deployed | React SPA + nginx gateway to auth-service/core-manager; Sites/Work Items/Pages/Direction views | admin-dashboard-and-api.md |
| ADM-002 | Admin API current state: dual-auth gateway, inventory, and fix blocks | partial | Two-service gateway audit: concrete bugs + hardcoded values, fixes sequenced A-F | admin-dashboard-and-api.md |
| ADM-003 | Core-manager API server surface (spec pin/unpin among admin routes) | deployed | core-manager exposes spec pin/unpin, keeping Pattern B lock semantics alive | admin-dashboard-and-api.md |
| ADM-004 | Work-item HITL model: approve/reject endpoints on pending_review status | superseded | Binary approval gate replaced by needs_human_review + PATCH /specs retry flow | admin-dashboard-and-api.md |
| ADM-005 | Admin work-item reassign + force-complete override endpoints | superseded | Two narrow overrides replaced by generic PATCH + shared retry/resolve pair | admin-dashboard-and-api.md |
| ADM-006 | WireGuard VPN admin-access implementation detail | superseded | 3 access options (WireGuard-in-cluster, VM bastion, port-forward); configs dropped from live doc | admin-dashboard-and-api.md |
| ADM-007 | Public REST API for the site-building pipeline | aspirational | Plan to expose sites/pages/work-items/specs over /api/v1/sites/*; never built | admin-dashboard-and-api.md |
| ADM-008 | site_ownership table / ownership model | abandoned | Proposed junction table for per-user site scoping; never created | admin-dashboard-and-api.md |
| ADM-009 | React admin dashboard for build review | partial | site-admin-dashboard.jsx: Dashboard/Review Queue/Review Detail views on mock data | admin-dashboard-and-api.md |
| ADM-010 | AI Persona Platform public API | superseded | Legacy v1 REST surface from the "AI personas" productisation era | admin-dashboard-and-api.md |
| RES-001 | vertical-exemplar-researcher — the exemplar-research relay hop | deployed | New relay hop researching 3 vertical exemplars; verified end-to-end on dartsonline | research-agents.md |
| RES-002 | research-agent (cited web research into research_results) | deployed | Web-search specialist: search→scrape→synthesise→cite, spawned by writer/classifier | research-agents.md |
| RES-003 | research_results with source attribution | deployed | Table storing research findings + full source attribution + expiry, per site/page | research-agents.md |
| RES-004 | Chat differentiator ideation agent | aspirational | Proposed agent ranking payable differentiators from asset × AI-capability combos | research-agents.md |
| RES-005 | Wayback/archive.org grounding method + limitation | partial | Probe pages grounded via Wayback snapshots; sandbox can't reach archive.org directly | research-agents.md |
| RES-006 | Capability watchlist + real-world event watchlist (dual standing research workflows) | aspirational | Two proposed recurring workflows tracking AI capabilities and scheme/event windows | research-agents.md |
| RES-007 | Deep-research domain insight agent | abandoned | Classifier deciding when multi-platform social research pays off; domain-flipping era | research-agents.md |
| ENT-001 | Entity data agent family (structured data drives pages) | partial | Typed JSONB entities with state-based lifecycle drive template-rendered pages | entity-data.md |
| ENT-002 | Events/tickets vertical (boxing first target) | abandoned | Planned first entity-driven site type (event/performer/venue/ticket); never shipped | entity-data.md |
| DMR-001 | Chassis deploy-mechanism reference (targets A–F) | deployed | Taxonomy of 6 deploy mechanisms across chassis/sites/idea.uk targets | deploy-mechanics-reference.md |
| PUB-001 | Public API plan: site_ownership junction + user-facing build/HITL endpoints | aspirational | site_ownership junction + endpoints for sites/pages/work-items/specs/assets; unbuilt | public-api.md |
| SCR-001 | Polite-scraping throttle (REQUEST_THROTTLE_MS) | aspirational | Optional per-adapter delay env var for polite bulk scraping | adopting-and-scraping.md |
