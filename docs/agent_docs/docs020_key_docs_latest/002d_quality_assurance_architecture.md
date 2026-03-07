The document covers:
The three-layer QA pipeline — structural checks (algorithmic), design/content audit (LLM-assisted), strategic review (LLM-required). Each produces work items processed by the same dispatch loop.
The agent hierarchy — why group agents rather than individual check agents or a registry of mini-actions. The key argument: grouping by shared context avoids redundant LLM calls, while keeping each group as a full orchestrator preserves the "every agent is an orchestrator" principle.
The promotion pattern — how a check goes from a query_database action step to a standalone agent with its own workflow. One line changes in the calling workflow, nothing downstream breaks. This is the practical answer to "when does an action become an agent?"
Algorithmic vs LLM-assisted vs LLM-required — how group agents run algorithmic checks first, include the results in the LLM prompt, and make one holistic LLM call per group. Some checks have both components.
Site-type-specific configurations — different sites enable different audit groups via maintenance_profile.audit. Adding new groups is creating an agent definition, not changing Go code.
Responsibility boundaries — the chain from classifier → planner → design agent → audit agent. The audit agent enforces the upstream decisions, it doesn't override them. Exception: adopted sites with no classifier output get "propose" mode.
Dream spec — aspirational spec from the classifier, feasibility annotations, gap analysis by the review agent, and how the backlog refills as capabilities grow.


# 002d — Quality Assurance Agent Architecture

Extension to 002 System Architecture. Covers the audit/review agent hierarchy, the relationship between structural checks and subjective assessment, the agent promotion pattern, and the dream spec concept.

---

## The Quality Assurance Pipeline

Three layers of quality assurance, each operating at a different level of abstraction.

```
Layer 1: Structural Checks (algorithmic, no LLM)
  → "Is component_id NULL?" "Does slot_name match data-component?"
  → Discovery checks in run_discovery_checks (existing pattern)
  → Binary right/wrong, cheap to run, run every cycle

Layer 2: Design & Content Audit (LLM-assisted, contextual)
  → "Is the colour palette cohesive?" "Is the typography hierarchy clear?"
  → Group auditor agents that combine algorithmic checks with one LLM call
  → Subjective assessment, moderate cost, run periodically

Layer 3: Strategic Review (LLM-required, holistic)
  → "Is this site achieving its purpose?" "What's the biggest gap?"
  → Top-level review agents that load full site context
  → Business-level judgement, higher cost, run on demand or infrequently
```

Each layer produces `site_work_items`. The dispatch loop processes them identically regardless of which layer created them.

---

## Agent Hierarchy

```
design-audit-agent (top-level orchestrator)
  ├── visual-design-auditor (group agent)
  │     Steps: load context → algorithmic checks → LLM visual assessment → write findings
  │     Checks: colour consistency, spacing, typography, dark sections, responsive
  │
  └── content-quality-auditor (group agent, reusable)
        Steps: load brief → load page samples → detect empty pages → LLM content review → write findings
        Checks: tone alignment, content gaps, CTA effectiveness, differentiation

site-review-agent (top-level orchestrator)
  ├── content-quality-auditor (same agent, reused via spawn+call)
  └── strategic alignment review (own LLM call)
        Checks: purpose alignment, page structure, dream spec gaps, conversion path
```

### Why group agents instead of individual check agents

Individual checks that need LLM assessment would each need to load the same context (CSS theme, colour palette, rendered HTML). Sending 2000 tokens of context with every small question is wasteful. Grouping checks by shared context means one context load and one LLM call per group.

However, each group agent IS a full orchestrator. If a specific check grows complex enough to need its own workflow (e.g. calling a vision AI for screenshot analysis, or spawning a research agent for competitor comparison), it can be promoted from an action step to a spawned sub-agent. The group agent's workflow changes one line — see "The Promotion Pattern" below.

### Why not a registry of mini-actions

The existing discovery check registry (`DiscoveryCheck` interface in `discovery_checks/`) works well for structural checks — SQL queries that return binary results. But LLM-assisted checks have different characteristics: they share context, they need prompt composition, and they may grow into multi-step workflows.

Making them a second class of object (registered functions that aren't agents) breaks the principle that every agent is an orchestrator. It creates something that can't be spawned, can't have workflows, and can't be extended independently.

The group-agent approach keeps one consistent pattern. Checks start as action steps within a group agent's workflow. If they need independence, they become agents. The calling pattern is the same either way.

---

## The Promotion Pattern

A check starts as an action step in a group agent's workflow:

```json
"check_colour_consistency": {
    "action": "query_database",
    "config": {
        "query": "SELECT ... hardcoded hex values ...",
        "params": ["site_record.site_id"]
    },
    "next_step": "check_typography",
    "output_field": "colour_findings"
}
```

When it needs more capability (e.g. calling a vision AI), promote it to an agent:

```json
"spawn_colour_checker": {
    "action": "spawn_agent",
    "config": { "role": "colour_checker", "agent_type": "check-colour-consistency" },
    "next_step": "call_colour_checker"
},
"call_colour_checker": {
    "action": "call_agent",
    "config": {
        "target_role": "colour_checker",
        "input_mapping": { "site_id": "site_record.site_id", "domain": "site_record.domain" }
    },
    "next_step": "check_typography",
    "output_field": "colour_findings"
}
```

The `colour_findings` output field stays the same. Nothing downstream changes. The check now owns its own workflow, can call external services, can be versioned and tested independently.

**Rule: Start as an action. Promote to an agent when the action needs multiple steps, external calls, or independent scaling.**

---

## Algorithmic vs LLM-Assisted vs LLM-Required

Each check falls into one of three categories:

| Category | Cost | Example | Implementation |
|----------|------|---------|----------------|
| Algorithmic | Free | "Is component_id NULL?" | SQL query or Go string operation |
| LLM-assisted | Moderate | "Is the colour palette cohesive?" | Could be algorithmic (compare hex values) but benefits from LLM judgement for edge cases |
| LLM-required | Higher | "Does the tone match the brief?" | Cannot be done algorithmically — needs language understanding |

Group agents run algorithmic checks first (via `query_database` steps), then make ONE LLM call that covers all LLM-assisted and LLM-required checks in the group. The algorithmic results are included in the LLM prompt as context — so the LLM doesn't re-check things already found algorithmically, but can use them to inform its assessment.

This means some checks have both an algorithmic and LLM component. The algorithmic part catches the obvious cases (hardcoded hex `#1a1a2e` that should be `var(--color-primary)`). The LLM catches subtle cases ("this shade of blue doesn't match the overall warm tone of the palette").

---

## Site-Type-Specific Audit Configurations

Different site types need different audit groups. A brochure site doesn't need feed quality checks. A news site doesn't need pricing page audits.

The site's `maintenance_profile` (in `sites.settings`) controls which audit groups run:

```json
{
    "maintenance_profile": {
        "audit": {
            "visual_design": { "enabled": true, "every": "7d" },
            "content_quality": { "enabled": true, "every": "7d" },
            "strategic_review": { "enabled": true, "every": "30d" },
            "feed_quality": { "enabled": false },
            "entity_accuracy": { "enabled": false }
        }
    }
}
```

As new site types are built, new group auditor agents are created:

| Site Type | Audit Groups |
|-----------|-------------|
| Brochure | visual_design, content_quality, strategic_review |
| News | visual_design, content_quality, feed_quality, freshness |
| E-commerce | visual_design, content_quality, product_accuracy, pricing |
| Events | visual_design, content_quality, entity_accuracy, event_lifecycle |

Each group is an agent with its own workflow. The top-level audit orchestrator spawns only the groups enabled for that site type. Adding a new group is creating an agent definition — no Go code changes to the orchestrator.

---

## Responsibility Boundaries

### Classifier → Planner → Design Agent → Audit Agent

The pipeline has a clear chain of authority:

| Stage | Agent | Decides | Stores in |
|-------|-------|---------|-----------|
| Classification | site-classifier | Industry, site type, build approach, archetype | `site_specs` |
| Planning | site-planner | Pages, sections, components, tone, audience | `site_specs`, `content_data` |
| Design | webdesign-agent | Colour palette, typography, spacing, CSS | `style_collections`, `css_themes` |
| Build | page-content-writer et al | Actual content, rendered HTML | `page_components` |
| Audit | design-audit-agent | Whether the build matches the plan | `site_work_items` |

**The audit agent is a consultant, not a dictator.** It reads the classifier's design intent and the planner's specifications, then checks whether the current implementation matches. It doesn't override design decisions — it enforces them.

```
Classifier decides: "professional-dark, blue/orange palette, serif headings"
  → stored in site_specs

Design agent creates: CSS theme with those colours and fonts
  → stored in css_themes

Audit agent checks: "The classifier said blue/orange but I see green in 3 sections"
  → creates work item: colour_fix, change green to var(--color-accent)
```

**Exception: no design intent exists.** For adopted sites (imported without running the classifier), the audit agent can propose a design direction. This is a different mode — "propose" vs "enforce" — and should be flagged in the work item for HITL review.

### Fix Agent Independence

Fix agents (handlers) don't know they were triggered by an audit. They receive `site_id`, `domain`, and spec fields. They load their own context. They could be called from CLI with the same inputs.

This means audit findings route to the same handlers used by the build pipeline and the improvement loop. There's one webdesign-agent, not separate ones for build vs audit vs manual request.

---

### Storage

```json
// In sites.content_data.dream_spec
{
    "ideal_pages": ["index", "about", "services", "case-studies", "blog",
                    "pricing", "faq", "testimonials", "team", "resources"],
    "ideal_features": ["live chat", "booking form", "portfolio gallery",
                      "newsletter signup", "client testimonials with photos"],
    "ideal_content_depth": "6-8 detailed service pages with case studies each",
    "ideal_design": "custom photography, animated hero, micro-interactions",
    "ideal_seo": "local SEO optimized, schema markup, blog with 2x/week posts",
    "feasibility": {
        "live_chat": { "possible": false, "reason": "no chat agent yet" },
        "booking_form": { "possible": true, "agent": "tool-deployer" },
        "newsletter_signup": { "possible": true, "agent": "tool-deployer" },
        "blog_posts": { "possible": true, "agent": "page-build-handler" },
        "custom_photography": { "possible": false, "reason": "only AI-generated images" }
    }
}
```

### Gap Analysis

The `site-review-agent` compares `dream_spec` against current state:

```
dream_spec says: 10 pages
current state: 7 pages
gap: 3 pages (pricing, faq, resources)
feasibility: all possible via page-build-handler
→ work items: needs_content_page × 3
```

The `feasibility` field prevents creating work items for things we can't build yet. As new agents come online, feasibility changes and previously blocked items become actionable.

### Updating the Dream Spec

see the classifier architecture doc for this - currently: 015_consolidated_site_spec_classifier_architecture.md
---

## Audit vs Fix: Same Checks, Different Modes

A design check can operate in two modes:

| Mode | Trigger | Action | Output |
|------|---------|--------|--------|
| **Audit** | Improvement loop, periodic schedule | Scan all components, find problems | Work items (`status: detected`) |
| **Fix** | Dispatch loop, work item handler | Fix a specific problem identified by audit | Updated component, needs_rerender item |

The check logic is the same code. What differs is scope (scan everything vs fix one thing) and authority (detect vs modify).

For algorithmic checks, the audit and fix functions can share Go code. For LLM-assisted checks, the audit prompt asks "what's wrong?" and the fix prompt asks "here's what's wrong, fix it."

This means we don't need separate "audit agent" and "fix agent" for each concern. The group auditor detects problems. The existing handler agents (webdesign-agent, component-template-fixer, page-build-handler) fix them. The handler doesn't know whether it was triggered by a human, an audit, or the build pipeline.

---

## Resolved Decisions

18. **Every agent is an orchestrator.** Quality checks start as actions within group agents. They are promoted to standalone agents when they need multi-step workflows or external calls. The promotion path is clean: one line changes in the calling workflow.

19. **Group agents own shared context.** Checks that need the same data (CSS theme, colour palette, page samples) are grouped into one agent that loads context once and runs checks in sequence. One LLM call per group, not per check.

20. **Audit agents enforce, not override.** They read the classifier/planner's intent and check whether the build matches. They don't make design decisions — they flag deviations from the stated intent.

21. **Dream spec drives the improvement backlog.** The gap between the aspirational spec and current reality generates work items. Feasibility annotations prevent creating impossible items. As capabilities grow, the backlog refills automatically.

22. **Site type determines audit configuration.** Different site types enable different audit groups via `maintenance_profile.audit`. New groups are new agents, not code changes to existing ones.

23. **Handlers don't know their trigger.** The same webdesign-agent handles `needs_design` from the build pipeline, `needs_design_review` from the audit, and manual design requests. It receives `site_id` + `domain` and does its job.
