# 026 — Infrastructure Layers, Site Adoption, and Backend Capabilities

How the platform serves sites, adopts existing sites, handles human direction at any stage, and scales to support backend functionality.

---

## Infrastructure Separation

The platform operates in three distinct layers. Layer 1 never sees external traffic.

### Layer 1 — Core Platform (existing K8s cluster)

The factory. Builds websites, runs agent orchestration, collects data, generates artifacts.

- Agent chassis, Kafka, Postgres
- Content generation, design, improvement loops
- Data collection (vet crawls, medicine prices, research)
- Produces: HTML, CSS, JS, JSON data, config files, Terraform, Docker images
- Internal traffic only — no public-facing endpoints

### Layer 2 — Client Delivery Infrastructure (separate)

The showroom. Serves websites, handles client API endpoints, runs client backends.

- Static site serving via S3/Backblaze (exists now)
- Site API router for form handling, data queries (planned)
- Client-facing Postgres for site data (planned)
- OVH VMs or dedicated servers, not the core cluster
- Provisioned and managed by Layer 1

### Layer 3 — Framework Builder (future)

Building other factories. Deploys agent frameworks and AI pipelines for clients.

- Provisions infrastructure for client-specific deployments
- Deploys known frameworks (LlamaIndex, LangGraph, custom)
- Each client gets isolated infrastructure
- Managed, monitored, and maintained by the platform

### Data Flow Between Layers

```
Core cluster (Layer 1)              Client server (Layer 2)
┌─────────────────────┐            ┌──────────────────────┐
│ Agent orchestration  │            │ nginx                │
│ Content generation   │──git──────→│ S3 (static files)    │
│ Data collection      │──db sync──→│ Postgres (site data) │
│ Terraform/Ansible    │──deploy───→│ site-api-router      │
│ Monitoring agents    │←─metrics──│ Prometheus            │
└─────────────────────┘            └──────────────────────┘
```

---

## Backend Capability Tiers

Not every site needs a server. Most don't.

### Tier 1 — Static + client-side JS

Brochure sites, portfolios, calculators, comparison tools. Everything runs in the browser. Data served as JSON from S3.

- Deployment: S3 + CDN
- Monthly cost: ~£0
- Status: fully supported now

### Tier 2 — Static + API router

Contact forms, newsletter signup, quote requests, error reports, simple data submissions.

- Deployment: S3 + site-api-router on client server
- Monthly cost: ~£5-15 (shared server allocation)
- Status: planned

### Tier 3 — Application with persistent state

User accounts, saved data, booking systems, small e-commerce.

- Deployment: S3 + site-api-router + Postgres + auth service
- Monthly cost: ~£25-50
- Status: planned

### Tier 4 — Application with real-time features

Chat, live updates, collaborative editing, notifications.

- Deployment: dedicated service with WebSocket support
- Monthly cost: ~£50-100
- Status: future

### Tier 5 — Full platform

Complex business logic, message queues, orchestration, background processing.

- Deployment: K8s namespace or dedicated cluster
- Monthly cost: ~£200+
- Status: future

### The vetcomparison.uk pattern (Tier 1 reference)

Demonstrates how much can be done without a server:

```
Data pipeline:  crawl → process → export JSON → git commit → S3
Client-side:    browser loads JSON index, does search/filter/sort/calculate
No API calls, no server queries, no backend
```

5,000 vets at ~500 bytes each ≈ 2.5MB JSON. Client-side search handles this. Medicine data lazy-loaded by letter chunk. The "backend" is a batch export that runs periodically on Layer 1 and pushes to S3.

This pattern works for up to ~50k items. Beyond that, pre-built search indexes (Pagefind) extend it to ~500k items without adding a server.

---

## Site API Router (Layer 2 Backend)

For sites needing POST endpoints or server-side queries. A shared Go service loading route configurations from the DB.

```go
// site_api_routes configuration in site_specs:
{
  "routes": [
    {
      "method": "POST",
      "path": "/api/contact",
      "actions": ["validate_input", "store_submission", "send_email"],
      "config": {"table": "form_submissions", "notify": "owner@example.com"}
    }
  ]
}
```

Reuses the existing action library. Configuration-driven — add routes via DB, no code deploy. The core cluster manages route configs as part of site specs.

### Minimum viable Layer 2

```
OVH VM (extending existing Terraform pattern)
  ├── nginx (reverse proxy, SSL via certbot, rate limiting)
  ├── site-api-router (Go binary, systemd service)
  ├── Postgres (form submissions, entity data, route configs)
  ├── Prometheus + Grafana (monitoring)
  ├── fail2ban (security)
  └── cron jobs (data exports, cert renewal)
```

---

## Human Direction

Humans influence sites at three levels. Each has different persistence and routing.

### Direction channels

| Channel | What it does | Persistence | Route |
|---------|-------------|-------------|-------|
| Work item request | "Do this specific thing" | Until completed | Dashboard → work item → dispatch loop |
| Direction update | "The site should be like this" | Permanent until changed | Dashboard → site_specs → strategist/planner/auditors |
| Reference suggestion | "Look at this site/idea" | Feeds next planning cycle | Dashboard → research context |

### Build-time direction

Already works. The trigger `input_data` carries instructions alongside domain/email/phone. The classifier sees it, incorporates it into specs. No changes needed.

```bash
# Existing pattern — freeform instructions in input_data
{"domain": "loancalculator.co.uk", "email": "...", 
 "notes": "Focus on first-time buyers, include repayment calculator"}
```

### Post-build direction (any time)

The site is live. The human wants to add features, change tone, suggest references. This needs a dashboard panel that writes to `site_specs` as a `direction` aspect.

The direction spec is pinned — agents can read it but not supersede it. Only humans change direction.

```json
site_specs aspect: "direction" (pinned)
{
  "requested_features": [
    {"type": "tool", "name": "loan repayment calculator", "priority": "must_have"},
    {"type": "content", "name": "first-time buyer guides", "priority": "should_have"}
  ],
  "reference_sites": [
    {"url": "https://moneysavingexpert.com/mortgages", "what_to_take": "content depth"}
  ],
  "notes": "Keep it simple, no jargon, focus on first-time buyers"
}
```

The strategist/planner reads direction and creates appropriate work items. The improvement loop respects it — auditors won't remove `must_have` features.

### HITL-requested content and lock lifecycle

When humans request specific content or features, the resulting components need protection from auditors. But they also shouldn't be frozen forever.

**Lock types:**

| Lock type | Behaviour | Use case |
|-----------|-----------|----------|
| `permanent` | Never expires, manual unlock only | Brand elements, legal disclaimers, human-crafted content |
| `timed` | Expires after N days (e.g. 90) | HITL-requested content that should eventually re-enter improvement cycle |
| `review` | Creates HITL review item on expiry | Content needing human approval before agents touch it again |

Implementation: `lock_type` and `lock_expires_at` columns on `page_components` and `site_components`. Discovery check queries expand from `AND locked_at IS NULL` to `AND (locked_at IS NULL OR lock_expires_at < NOW())`.

When a `review` lock expires, a discovery check creates a `needs_lock_review` work item. Human decides: re-lock, release, or update. If no human response within a configurable period, the system can auto-release (per-site config).

### Audit pass cap reset

The 3-pass audit cap prevents unbounded improvement cycles. Resets:

| Trigger | Mechanism |
|---------|-----------|
| Time-based (default) | Improvement sweep checks `last_audit_reset_at`, resets after N days (e.g. 60) |
| Direction change | Human updates direction spec → auto-reset |
| Major rebuild | N pages rebuilt in one cycle → auto-reset |
| Manual | Human clicks "re-audit" in dashboard |

Combined with timed lock expiry, this creates a natural rhythm:

```
Build → audit × 3 → cap reached → site quiet
  ... 60 days pass ...
  → pass counter resets, expired locks release
  → improvement loop runs fresh
  → finds new issues (content aged, design dated, new opportunities)
  → audit × 3 → quiet again
```

Sites breathe and improve rather than either constant churn or permanent stasis. HITL involvement is optional — the system self-maintains, humans intervene when they want to.

---

## Site Adoption

### Core principle

Adoption is a one-off data capture, not a permanent state. After adoption, the site is like any other — evolving, improving, with the strategist setting aspirational goals.

### What adoption is NOT

Adoption does not write a spec that describes the status quo as the target. If the adopted site has 5 pages, the spec shouldn't say "this site should have exactly 5 pages." The adoption captures what exists. The strategist then writes what the site SHOULD BECOME — which goes beyond what was adopted.

### Separation of concerns

```
adoption_source data    → transient, flows through collected_data during adoption
                           strategist reads it, uses it, then it's gone
                           
identity spec           → what the site is now (evolves with improvements)
content_direction spec  → where it should go (aspirational, written by strategist)
structure spec          → page layout (evolves as site grows)
```

The only trace of adoption in the permanent specs is a reference field:

```json
identity spec: {"adopted_from": "mortgagecalculator.co.uk", "adopted_at": "2026-03-30", ...}
```

No raw crawl data stored in specs. If someone needs the crawl data again, they re-crawl. The site has probably changed anyway.

### Adoption flow

Adoption goes through the existing build pipeline, not a separate agent. The classifier handles the deep crawl as part of its research when `adopt_from` is present in `input_data`.

```
Trigger: domain-submitter with adopt_from in input_data
  → classifier: deep crawl of adopt_from URL as part of research
  → classifier writes identity spec (from crawl + research)
  → strategist reads crawl data + identity + industry research
  → strategist writes aspirational content_direction (goes BEYOND adopted state)
  → planner creates pages matching adopted structure
  → content writer uses existing_content from crawl for initial content
  → design agent matches/improves the palette
  → site deploys
  → improvement loop has room to push toward content_direction
```

The content writer receiving `mode: "recreate"` with `existing_content` uses the original text as source material, adapting it to our component templates. This preserves the site's content while fitting it into our system.

### Adoption modes

| Mode | Trigger | What happens |
|------|---------|-------------|
| Take over domain | `{"domain": "mortgagecalculator.co.uk", "adopt_from": "https://mortgagecalculator.co.uk"}` | Deep crawl, recreate content faithfully, then improve |
| Build inspired by | `{"domain": "loancalculator.co.uk", "reference_sites": ["https://mortgagecalculator.co.uk"]}` | Classifier uses reference in research, strategist takes what's useful, content is fresh |
| Normal build | `{"domain": "newsite.co.uk"}` | Classifier researches the vertical, finds competitors organically |

All three go through the same pipeline. The difference is how much existing content feeds into the specs. The planner and content writer adapt based on what's in the specs — they don't need to know whether it came from adoption, reference, or pure generation.

---

## Research, Patterns, and the Component Library

### Research data retention

Currently, classifier research is transient — lives in `collected_data`, feeds into classification, then gone. This wastes valuable crawl data.

Change: the classifier persists its research findings in `site_specs` as a `research` aspect. This is a summary of what was found (industry patterns, competitor analysis, opportunities), not the raw crawl dump.

### Pattern extraction (Phase 3)

A separate `pattern-extraction-agent` analyses research data and extracts reusable patterns for the component library:

- **Tool specs:** "A mortgage calculator that accepts loan amount, rate, term. Shows monthly payment, amortisation breakdown, multiple scenarios." Detailed enough for the tool-builder to create a functionally equivalent tool.
- **Layout patterns:** "Financial calculator sites typically have: hero with key stat, calculator tool, comparison table, educational guides, FAQ, trust signals."
- **Content patterns:** "Guide structure: intro, key takeaways box, detailed sections with examples, related links, disclaimer."
- **UX observations:** Good and bad examples with descriptions. "Live-updating results as inputs change (good). Results only after clicking calculate (bad)."

Runs as a side effect after classification, or on a schedule, or manually triggered. Patterns accumulate over time and become available to future planners/builders.

### Code as reference input

For complex tools (games, interactive visualisations, advanced calculators), descriptions alone aren't enough for the LLM to understand the intended behaviour. The reference code is included in the tool-builder's prompt as context:

```
"Here is reference code for a similar tool. Study the mechanics,
interaction patterns, and scoring system. Create an ORIGINAL 
implementation achieving similar functionality with [these differences].
Do not copy the code."

Reference code:
[actual JS from research]
```

The LLM sees the code, understands how it works, produces something new. This is not copying — it's the same way a developer studies an existing implementation before building their own.

**Copyright considerations:** We never deploy extracted code directly. The research stores code snippets as reference material. The tool-builder's prompt explicitly instructs original implementation. The output goes through the normal validation pipeline (which includes cross-site contamination checks). The reference material is internal — it's never served to users.

### Integration with RAG and training pipelines

Stored LLM prompts and successful outputs (tools, components, content) feed the existing RAG pipeline. When the tool-builder generates a mortgage calculator, the prompt + output pair is stored. Future tool-builder calls retrieve relevant examples via RAG. This improves output quality over time without fine-tuning.

The pattern library (tool specs, good/bad examples) enriches RAG retrieval — when building a calculator, the system retrieves both the abstract spec ("should have amortisation breakdown") and concrete examples ("here's a successful calculator we built previously").

---

## Phase Plan

### Phase 1 — Adoption and direction

- Adoption via existing pipeline with `adopt_from` in input_data
- Deep crawl during classifier research
- Content writer `existing_content` support (prompt change)
- Dashboard direction panel for post-build HITL
- Persist classifier research in site_specs

### Phase 2 — Tool and entity adoption

- Tool extraction: detailed specs from reference code
- Entity extraction: structured data from directory/listing pages
- Data export action: DB → JSON → git → S3
- Lock types (permanent, timed, review) on components

### Phase 3 — Patterns and research

- Pattern extraction agent
- Tool spec library (reusable across sites)
- Good/bad example library for LLM prompts
- RAG integration for pattern retrieval
- Audit pass auto-reset (time-based)

### Phase 4 — Client-facing backend (Layer 2)

- Site-api-router Go service
- OVH VM provisioning with Terraform
- Form handling, data queries
- Auth and persistent state

### Phase 5 — Framework deployments (Layer 3)

- Infrastructure provisioning agents
- Framework deployment handlers (LlamaIndex, LangGraph)
- Server entity management
- Self-managing infrastructure

---

## Principles

1. **Layer 1 never serves external traffic.** It produces artifacts and pushes them.

2. **No vendor lock-in.** Use third-party services but always with the exit path designed. The site-api-router is our own code. Infrastructure provisioning supports multiple providers.

3. **Static-first.** Default to the lowest tier that serves the need. The JSON-on-S3 pattern handles more cases than expected.

4. **Same pipeline, richer specs.** Adoption, new builds, maintenance, and framework deployment all flow through the work item pipeline. The orchestration layer is the same.

5. **Adoption is a starting point, not a ceiling.** The adopted state is the baseline. The strategist writes aspirational direction. The improvement loop pushes toward it. No spec describes the status quo as the target.

6. **Research is reference, not template.** Crawled code and content inform generation but are never deployed directly. The system creates original implementations inspired by research. Descriptions and specs, not copies.

7. **Human direction persists.** HITL requests go into pinned specs. Agents respect them. Lock types control how long protection lasts. The system self-maintains but humans set the direction.

8. **Servers are entities.** Infrastructure follows the same patterns as site data — state-based lifecycle, discovery checks for health, work items for management tasks.

9. **The framework is a framework builder.** The same system that generates websites can generate infrastructure configs, framework deployments, and management automation.
