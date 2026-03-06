# Site Classifier — Evolving Responsibilities

## Current State

The classifier takes a domain name (and optional objectives) and produces:

```json
{
    "build_approach": "page-assembled | application | hybrid",
    "primary_archetype": "descriptive label",
    "page_types": ["content", "tool", "commerce", "entity-directory", "news"],
    "entity_types": [],
    "hosting_trajectory": "static_only | static_now_api_later | needs_server",
    "detected_industry": "industry name",
    "recommended_builder": "pageflow-builder"
}
```

This is a routing decision — it tells the system which builder to use and what
kind of site to expect. It doesn't describe what the site should look like.

## What It Needs to Become

The classifier is the first agent in the pipeline. Everything downstream
depends on its output — the planner reads it, the design agent reads it,
the content writer reads it, and now the audit/review agents read it too.

The classifier should produce a **unified site spec** — one document that
describes everything the site should be, from pages to features to design
intent. Items in the spec are tagged by readiness: deployable now, planned
for later, or blocked pending new capabilities.

### Unified Spec Structure

```json
{
    "classification": {
        "build_approach": "page-assembled",
        "primary_archetype": "brochure-with-lead-gen",
        "detected_industry": "energy / wholesale fuel",
        "hosting_trajectory": "static_only",
        "recommended_builder": "pageflow-builder"
    },

    "identity": {
        "company_name": "Gas Wholesalers",
        "tagline": "Wholesale Gas Supply Solutions",
        "tone": "professional, direct, trustworthy",
        "target_audience": "commercial fuel buyers, fleet managers, facility managers",
        "key_messages": [
            "Reliable supply network",
            "Competitive wholesale pricing",
            "24/7 availability"
        ],
        "differentiators": [
            "Multiple supply sources for redundancy",
            "No hidden fees or surprise charges",
            "Dedicated account management"
        ]
    },

    "design_intent": {
        "style_direction": "professional-dark",
        "colour_mood": "dark blue and orange — energy, petroleum, industrial strength",
        "typography_mood": "clean sans-serif — modern corporate, not stuffy",
        "imagery_direction": "industrial facilities, fuel infrastructure, fleet vehicles",
        "layout_preference": "spacious sections, clear hierarchy, prominent CTAs",
        "reference_sites": [],
        "avoid": ["clip-art icons", "stock photo people", "generic corporate blue"]
    },

    "pages": [
        {
            "name": "index",
            "title": "Gas Wholesalers | Wholesale Fuel Distribution",
            "purpose": "hero + value proposition + services overview + CTA",
            "sections": ["hero", "features", "services-grid", "differentiators", "social-proof", "call-to-action"],
            "status": "deployed",
            "priority": 1
        },
        {
            "name": "about",
            "title": "About Us",
            "purpose": "company story, team, values",
            "sections": ["hero", "about-content", "team-grid", "call-to-action"],
            "status": "deployed",
            "priority": 2
        },
        {
            "name": "services",
            "title": "Our Services",
            "purpose": "detailed service descriptions with individual CTAs",
            "sections": ["hero", "services-detail", "call-to-action"],
            "status": "deployed",
            "priority": 3
        },
        {
            "name": "pricing",
            "title": "Pricing",
            "purpose": "transparent pricing tiers or quote request",
            "sections": ["hero", "pricing-tiers", "faq", "call-to-action"],
            "status": "planned",
            "priority": 60,
            "handler": "page-build-handler"
        },
        {
            "name": "faq",
            "title": "FAQ",
            "purpose": "common questions about wholesale fuel supply",
            "sections": ["hero", "faq-accordion", "call-to-action"],
            "status": "planned",
            "priority": 70,
            "handler": "page-build-handler"
        },
        {
            "name": "client-portal",
            "title": "Client Portal",
            "purpose": "order tracking, account management",
            "status": "blocked",
            "blocked_reason": "no auth/portal agent",
            "priority": 200
        }
    ],

    "features": [
        {
            "name": "newsletter_signup",
            "description": "Email capture for market updates",
            "status": "planned",
            "handler": "tool-deployer",
            "priority": 80
        },
        {
            "name": "quote_request_form",
            "description": "Multi-field form for supply enquiries",
            "status": "planned",
            "handler": "tool-deployer",
            "priority": 40
        },
        {
            "name": "live_chat",
            "description": "Real-time customer support",
            "status": "blocked",
            "blocked_reason": "no chat integration agent",
            "priority": 150
        }
    ],

    "content_direction": {
        "voice": "We speak as experienced industry insiders, not salespeople",
        "avoid_phrases": ["synergy", "cutting-edge", "world-class", "solutions provider"],
        "emphasis": "reliability, transparency, no-nonsense service",
        "social_proof_style": "company commitments rather than fabricated testimonials",
        "blog_strategy": "industry insights, market commentary, regulatory updates"
    },

    "seo": {
        "primary_keywords": ["wholesale fuel distribution", "natural gas supply UK", "bulk fuel delivery"],
        "local_seo": true,
        "schema_types": ["LocalBusiness", "Service"],
        "meta_description_tone": "direct and benefit-focused"
    },

    "maintenance_profile": {
        "audit": {
            "visual_design": { "enabled": true, "every": "7d" },
            "content_quality": { "enabled": true, "every": "7d" },
            "strategic_review": { "enabled": true, "every": "30d" }
        },
        "content_refresh": { "enabled": false },
        "link_checking": { "enabled": true, "every": "24h" }
    }
}
```

### Key Properties of the Unified Spec

**One document, not two.** There is no separate "dream spec" and "build spec."
Every item has a `status` field: `deployed`, `planned`, `blocked`. The
"dream" is the full document. The "build" is the subset that isn't blocked.

**Status flow:** `planned → triaged → claimed → complete → deployed`
or `planned → blocked` (can't do yet). The classifier sets initial statuses.
The planner may adjust priorities. The dispatch loop processes items in
priority order, skipping blocked items.

**Handlers are declared per item.** The classifier knows which agent handles
each type of work. When a new agent is deployed, blocked items with that
handler become unblocked automatically (via the feasibility re-check task).

**Design intent is explicit.** Instead of the webdesign-agent guessing from
the industry name, the classifier states what the design should feel like.
The webdesign-agent implements this intent. The design-audit-agent checks
whether the implementation matches.

**Content direction is explicit.** Instead of the page-content-writer
guessing tone, the classifier specifies voice, emphasis, and things to avoid.
The content-quality-auditor checks whether the content matches.

---

## What the Classifier Needs to Know

### Inputs

| Source | What | Used for |
|--------|------|----------|
| Domain name | `gaswholesalers.com` | Industry detection, company name |
| Human objective | "sell leads to gas suppliers" | Purpose, audience, messaging |
| Existing site (if any) | Scraped content | Current state, what to keep |
| Industry research | Web search results | Competitor patterns, industry norms |
| Agent registry | Available handlers | Feasibility assessment |
| Component library | Available sections | Section selection |
| Style collections | Available designs | Design direction |

### How It Gets Them

The classifier's workflow could be:

```
scrape_existing_site (if exists)
  → research_industry (web search for competitors, best practices)
  → load_agent_registry (query agent_definitions for available handlers)
  → load_component_library (query content_components for available sections)
  → load_style_collections (query style_collections for design options)
  → generate_unified_spec (LLM call with all context)
  → write_spec (store in site_specs)
  → write_work_items (create site_work_items from spec pages/features)
  → complete
```

The LLM call receives all the context and produces the unified spec.
The `write_work_items` step creates `site_work_items` entries for every
`planned` page and feature. Blocked items are also created but with
`status: blocked`.

### Feasibility Assessment

The classifier checks the agent registry to determine what's achievable:

```go
// Pseudocode for feasibility check
func assessFeasibility(feature string, agentRegistry map[string]bool) string {
    handlerMapping := map[string]string{
        "content_page":     "page-build-handler",
        "newsletter":       "tool-deployer",
        "contact_form":     "tool-deployer",
        "quote_form":       "tool-deployer",
        "live_chat":        "chat-integration-agent",
        "booking_system":   "booking-agent",
        "client_portal":    "portal-agent",
        "ecommerce":        "commerce-agent",
    }

    handler := handlerMapping[feature]
    if agentRegistry[handler] {
        return "planned"  // handler exists, can build
    }
    return "blocked"      // handler doesn't exist yet
}
```

This is a Go action, not LLM-based. The LLM decides what the site *should*
have. The feasibility check determines what it *can* have right now.

---

## Downstream Impact

### Planner
Currently the planner decides pages, sections, and components. With the
unified spec, the planner's role shifts to **validation and enrichment** —
it receives the classifier's page list and fills in section-level detail
(which component templates to use, content direction per section, research
requirements). It doesn't decide what pages exist — the classifier does.

### Design Agent
Currently guesses design direction from industry name. With `design_intent`
in the spec, it implements explicit direction. The design-audit-agent can
then check whether the implementation matches.

### Content Writer
Currently generates content with minimal guidance. With `content_direction`
in the spec, it has explicit voice, emphasis, and avoid-lists. The
content-quality-auditor checks alignment.

### Audit/Review Agents
Read the unified spec as their ground truth. Deviations from the spec
become findings → work items. The spec is what the audit enforces.

### HITL
The unified spec is the primary HITL review point. A human can adjust
the spec before anything is built — change tone, add pages, remove features,
adjust priorities. Everything downstream follows the spec.

---

## Implementation Path

1. **Extend the current classifier output** — add `identity`, `design_intent`,
   `content_direction` fields to the existing classification response.
   Store as `site_specs.spec_type = 'unified_spec'`.

2. **Add feasibility checking** — new Go action `check_feasibility` that
   queries agent_definitions and component library. Runs after the LLM
   classification, before writing work items.

3. **Write work items from spec** — extend `WriteBuildItemsAction` to
   create items for all `planned` pages/features with correct handlers
   and priorities. Create `blocked` items for infeasible features.

4. **Planner reads unified spec** — instead of generating its own page list,
   the planner enriches the classifier's page list with section-level detail.

5. **Audit agents read unified spec** — `design_intent` and `content_direction`
   become the ground truth for audit checks.

Steps 1-3 can happen incrementally. The current classifier still works — the
new fields are additive. Steps 4-5 are refactors that reduce duplication
between classifier and planner.


====

the planner/discovery/manual entry puts in the intended handler name even if it doesn't exist yet. The system handles the rest:

Planner creates: handler_agent = 'tool-deployer', status = 'triaged'
→ Dispatch loop claims it
→ Claim action checks agent_definitions: 'tool-deployer' not found
→ Item marked 'blocked', error = 'Handler agent not registered: tool-deployer'
→ Item sits in 'blocked'
→ ... weeks later, tool-deployer is deployed ...
→ Feasibility-recheck task finds it: agent_definitions now has 'tool-deployer'
→ Item promoted to 'triaged'
→ Next dispatch loop picks it up and processes it

The classifier's feasibility check (from the unified spec design) serves a different purpose — it sets status: blocked on the spec entry so the planner can decide whether to create the work item at all or create it pre-blocked. But even if the planner creates it as triaged, the claim action catches it.

