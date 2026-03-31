# Spark Strategic Planning Architecture

## Approach

No new tables beyond the schema migration for component selection metadata (which is needed regardless of vonc.com — it's part of the parallel component selector work). No chassis code changes. The mission and roadmap travel as `input_data` fields. Content writers get context through structured `content_context` fields per page. No RAG for v1.

The component library grows through use: the planner requests section types, the selector finds matching components or flags gaps, the component-creator handler generates new templates following the component contracts, and the library fills up for reuse.

## What goes where

| Data | Where | Who reads it |
|---|---|---|
| Mission (positioning, differentiators, tone, objectives) | `input_data.mission` → persisted to `site_specs` aspect `'mission'` | Classifier, planner, content writers |
| Roadmap (current phase, pages with purposes and content context) | `input_data.roadmap` → persisted to `site_specs` aspect `'roadmap'` | Planner, content writers |
| Per-page detail (archetype descriptions, mechanic explanations) | `input_data.roadmap.phases.v1_content_first.pages[].content_context` | Content writers (for that specific page) |
| Full concept document | Offline reference (not ingested for v1) | Humans; agents in v2+ via RAG |

---

## How the pipeline handles a new site type

### The problem with the current pipeline

The classifier knows: brochure, landing page, portfolio, blog, e-commerce. The planner selects from existing components: hero-split, differentiators-3-column, testimonials, CTA blocks. The design agent picks from existing style collections. None of these fit an interactive social platform.

### The solution: component selector + component creator (parallel conversation)

The parallel conversation designed a component selection system that separates:

- **Planner** decides WHAT section types a page needs — structural decision based on page type, site type, roadmap
- **Component selector** decides WHICH template to use — queries DB by `section_type`, scores by site_type match, quality, usage
- **Component creator** handles "no suitable component" — generates template following contracts, stores for reuse

This means new site types work without special-casing. The system discovers it doesn't have what it needs and creates it.

### The flow for vonc.com

```
1. Classifier reads mission
   → outputs identity, design_intent, content_direction
   → site_type = "interactive-platform"

2. Planner reads roadmap current phase
   → outputs section_types per page:
     index:        [provocation-card, lobby-grid, gauntlet-cta, brief-explanation]
     provocations: [provocation-feed]
     gauntlet:     [gauntlet-interface, archetype-result-card]
     about:        [about-hero, platform-comparison, game-master-explanation, gauntlet-cta]
     archetypes:   [archetype-grid, archetype-combinations]

3. Component selector queries for each section_type:
   SELECT * FROM content_components
   WHERE section_type = 'provocation-card'
     AND component_level = 'section'
   ORDER BY
     CASE WHEN suitable_site_types @> '"interactive-platform"' THEN 0.4 ELSE 0.1 END
     + COALESCE(avg_quality_score, 0.3) * 0.3
     + LEAST(usage_count::float / 50.0, 1.0) * 0.1
   DESC LIMIT 3

   → No match for provocation-card → creates needs_new_component work item
   → No match for lobby-grid → creates needs_new_component work item
   → gauntlet-cta might match an existing CTA variant → selector scores it
   → etc.

4. Component creator processes each needs_new_component item:
   - Reads: section_type, description (from roadmap), design_direction,
     component contracts (from prompt)
   - LLM generates: html_template + input_schema following all contracts
   - Stores in content_components with full selection metadata
   - Marks work item complete

5. Content writer fills each component's template variables via LLM
   - provocation-card: {{.provocation_text}}, {{.ai_take}}, {{.time_remaining}}
   - For v1 static: filled with example/seed content
   - Template stays same when dynamic content arrives later

6. Design agent generates CSS theme from design_intent
   "Game energy, dark theme, vibrant accents, arena lobby feel"

7. Page assembly → git commit → S3 deploy

8. Quality audit → scores per component → feeds back to selection metadata
```

### Second build reuses everything

When vonc.com is rebuilt, or a vertical site gets Spark features:
- Selector finds `provocation-card` with `section_type` match and usage history
- No component creation needed — template exists, scored, proven
- Content writer fills template with new content
- Components improve over time through audit feedback and manual refinement

---

## Component creation contracts

The component-creator handler's LLM prompt must include the full component contract so every generated template follows the system's rules. Compiled from docs 003 (contracts), 018 (dynamic application guidelines), and the input schema v2 contract.

### The contract prompt (included in component-creator agent definition)

```
You are creating a reusable HTML component template for the agent system.

SECTION TYPE: {{.section_type}}
SUITABLE SITE TYPES: {{.suitable_site_types}}
DESCRIPTION: {{.description}}
DESIGN DIRECTION: {{.design_direction}}
REFERENCE CONTENT: {{.reference_content}}

== COMPONENT CONTRACT — YOU MUST FOLLOW ALL OF THESE ==

1. STRUCTURE:
   <style> scoped CSS </style>
   <section class="{function}-section" data-component="{function}">
     HTML using {{.variable}} template placeholders
   </section>
   <script> if interactive, self-contained JS in IIFE </script>

2. NAMING:
   - function value in kebab-case: lowercase, digits, hyphens only
   - Root element has data-component="{function}" matching exactly
   - Class on root element: {function}-section

3. TEMPLATE VARIABLES:
   - Use {{.field_name}} for all content that varies per instance
   - Generate an input_schema declaring each field:
     {"fields": {"field_name": {"type": "text|array|image|url|boolean",
      "source": "llm|site_specs.{path}|site_assets.{type}|renderer|static",
      "required": true|false, "llm_guidance": "hint for content writer"}}}

4. CSS RULES:
   - ALL colours via CSS variables with fallbacks
   - Light sections: color: var(--color-text); headings: var(--color-heading)
   - Dark sections: color: var(--section-text, rgba(255,255,255,0.9));
     headings: var(--section-heading, #ffffff)
   - NEVER hardcode hex colours on text elements
   - Scope ALL CSS to .{function}-section — no global element rules (h1 {}, p {})
   - Include @media (max-width: 768px) responsive rules
   - Mobile-first: touch targets >= 44px

5. DARK SECTIONS (if is_dark_section = true):
   Set these CSS custom properties on the root container:
   --section-text: rgba(255,255,255,0.9);
   --section-text-muted: rgba(255,255,255,0.7);
   --section-heading: #ffffff;
   --section-surface: rgba(255,255,255,0.05);
   --section-border: rgba(255,255,255,0.2);

6. CSS VARIABLES AVAILABLE:
   Colours: --color-primary, --color-primary-hover, --color-primary-text,
     --color-secondary, --color-accent, --color-text, --color-text-muted,
     --color-heading, --color-background, --color-surface, --color-card-bg,
     --color-border, --color-header-bg, --color-header-text,
     --color-footer-bg, --color-footer-text, --color-white
   Layout: --container-max-width (1200px), --spacing-section (5rem 2rem),
     --border-radius, --shadow

7. INTERACTIVE ELEMENTS (if section has JS):
   - Client-side only, no external API calls
   - Wrap in IIFE: (function() { ... })();
   - No global variable pollution
   - Progressive enhancement — works without JS where possible
   - No external CDN imports unless explicitly listed

8. QUALITY:
   - No placeholder text (Lorem ipsum, TODO, [INSERT], NEEDS HUMAN REVIEW)
   - No unrendered template variables in output
   - Semantic HTML (section, article, nav, header — not div soup)
   - Accessible: labels on inputs, ARIA where needed, focus states
   - No fabricated content (stats, testimonials, quotes)

== END CONTRACT ==

Generate the html_template and input_schema for this component.
```

### Where to store the contract

For now: in the component-creator agent definition's `default_config.prompt_template`. It's static enough.

Later: if the contracts evolve frequently, extract into a `component_creation_guidelines` entry in the knowledge_base that the component-creator loads via `rag_lookup` at the start of each run. This way contract updates propagate without redeploying the agent definition.

### Component metadata on creation

When the component-creator stores a new component:

```sql
INSERT INTO content_components (
  name, function, display_name, category, component_level,
  section_type, suitable_site_types, suitable_page_types,
  content_shape, visual_density,
  semantic_tags, description, html_template, input_schema,
  is_dark_section, render_mode, created_from, is_active,
  usage_count, avg_quality_score
) VALUES (
  'spark-provocation-card',
  'spark-provocation-card',
  'Provocation Card',
  'interactive-platform',
  'section',
  'provocation-card',
  '["interactive-platform", "community", "game"]'::jsonb,
  '["landing", "provocation-index", "challenge"]'::jsonb,
  'structured_card',
  'medium',
  '["provocation", "challenge", "game", "interactive", "timer", "spark"]'::jsonb,
  'Challenge card with provocation text, text input, countdown timer, AI take',
  '<style>...</style><section class="spark-provocation-card-section" data-component="spark-provocation-card">...</section>',
  '{"fields": {"provocation_text": {"type": "text", "source": "llm", "required": true}, "ai_take": {"type": "text", "source": "llm", "required": true}, "time_remaining": {"type": "text", "source": "renderer", "required": false}, "response_count": {"type": "text", "source": "renderer", "required": false}}}'::jsonb,
  true,
  'template',
  'generated',
  true,
  0,
  NULL
);
```

The `created_from: 'generated'` distinguishes LLM-created components from manually crafted or adopted ones. Quality scoring starts at NULL (unproven). Usage count starts at 0. Both accumulate through use and audit feedback.

---

## The initial request for vonc.com

```json
{
  "action": "orchestrate",
  "config": {"agent_type": "intake-orchestrator"},
  "input_data": {
    "domain": "vonc.com",
    "objective": "Build a content-first launch site for Spark — an AI-driven social platform where the world is the content and your take is the game. V1 is static S3 with daily provocation generation and a viral archetype quiz (the Daily Gauntlet). The landing page IS the product — a single provocation card with a text input, not a marketing page about the product.",
    "mission": {
      "mission": "AI-driven social platform where the world is the content and your take is the game",
      "positioning": "Not another social app. A daily game powered by what's happening in the world. You don't post. You don't scroll. You play.",
      "tagline": "TikTok is where you watch. Twitter is where you shout. Spark is where you play.",
      "key_differentiators": [
        "AI as game master and producer, never the performer — when something is funny or impressive, a human did it",
        "Opinion-first entry point — lower barrier than creation-first. You don't need an idea, you need a take",
        "Ephemeral challenges with permanent emergent reputation — disposable content, lasting identity",
        "Rooms not feeds — enter a space with energy, don't scroll a stream passively",
        "Arena (competitive takes) + Stage (creative showcase) dual modes",
        "Archetypes earned from behaviour, not self-reported — your profile emerges from how you engage"
      ],
      "target_users": {
        "primary": "People who have opinions on current trends but no platform that asks for them",
        "secondary": "Creators wanting skill showcase with built-in audience and niche discovery",
        "tertiary": "Curators and tastemakers — people who find things, not make things"
      },
      "content_tone": {
        "is": ["interesting", "slightly adult", "competitive", "aspirational", "playful", "game-like", "energetic"],
        "is_not": ["dark", "gruesome", "violent", "xxx", "newspaper news", "corporate", "generic social media", "AI slop"],
        "ai_role": "Producer not performer. AI frames, scores, curates, synthesises. AI does not do humour, hot takes, sarcasm, or anything meant to feel spontaneous. When the AI is visible, it's as game master — setting the board, not playing the game."
      },
      "core_concepts": {
        "the_game_metaphor": "Everything is framed as a game. Challenges have timers, rules, scores, winners.",
        "engage_first": "No signup walls, no explanations. A provocation card and a text input. Engage first, understand second, commit third.",
        "ai_slop_awareness": "AI restricted to genuine strengths. Deliberately bad AI takes as challenge seeds — AI as straight man.",
        "cold_start": "Solo mode fully functional for 1 person. AI sparring partner fills the room transparently.",
        "archetype_system": "Profile emerges from behaviour. Daily Gauntlet produces shareable archetype cards. Viral mechanic."
      },
      "measurable_objectives": [
        {"id": "obj_dau_50", "objective": "50 DAU within 4 weeks", "metric": "daily_active_users", "target": 50, "status": "not_started"},
        {"id": "obj_gauntlet_rate", "objective": "Gauntlet completion >15% of visitors", "metric": "gauntlet_completion_rate", "target": 0.15, "status": "not_started"},
        {"id": "obj_session_duration", "objective": "Avg session >3 minutes", "metric": "avg_session_duration_seconds", "target": 180, "status": "not_started"},
        {"id": "obj_share_rate", "objective": ">5% Gauntlet completions shared", "metric": "gauntlet_share_rate", "target": 0.05, "status": "not_started"}
      ]
    },
    "roadmap": {
      "current_phase": "v1_content_first",
      "phases": {
        "v1_content_first": {
          "description": "Static S3 site. Provocations as SEO content. Daily Gauntlet as viral archetype quiz.",
          "hosting": "s3_static",
          "status": "active",
          "design_direction": "Game energy, not corporate. Dark theme with vibrant accents. Cards, timers, energy indicators. Arena lobby feel. Mobile-first.",
          "pages": [
            {
              "slug": "index",
              "purpose": "The product IS the landing page. Single provocation card filling screen. Text input beneath. Timer. No signup wall. Below: AI's take (decent but beatable). Below that: today's other provocations. Scroll further: brief explanation + Gauntlet invitation.",
              "section_types": ["provocation-card", "lobby-grid", "brief-explanation", "gauntlet-cta"],
              "priority": 1
            },
            {
              "slug": "provocations",
              "purpose": "Today's active provocations with AI takes. Each has shareable URL. SEO content. Card-based layout.",
              "section_types": ["provocation-feed"],
              "priority": 2
            },
            {
              "slug": "gauntlet",
              "purpose": "Daily Gauntlet — 5 provocations, 5 minutes, discover your archetype. Client-side JS. Produces shareable archetype card.",
              "section_types": ["gauntlet-interface", "archetype-result-card"],
              "priority": 3,
              "content_context": {
                "archetypes": [
                  {"name": "The Surgeon", "tagline": "Precise, analytical, cuts to the core", "description": "Strips arguments to essential structure. Fast on logical frameworks. Finds structural flaws others miss.", "strengths": "Analytical speed, structural clarity, finding hidden flaws"},
                  {"name": "The Wildcard", "tagline": "Unpredictable, highest remix rate", "description": "Takes ideas in unexpected directions. Crosses domains freely. Most likely to turn serious into surprising.", "strengths": "Creative leaps, remix ability, cross-domain thinking"},
                  {"name": "The Oracle", "tagline": "Quiet but accurate, rare but impactful", "description": "Doesn't respond often but when they do it lands. Best prediction accuracy. Picks battles carefully.", "strengths": "Prediction accuracy, timing, selective engagement"},
                  {"name": "The Catalyst", "tagline": "Responses generate longest chains", "description": "Individual responses might not be best, but they spark others. Opens new directions. The chain-starter.", "strengths": "Provocation, opening angles, generating momentum"},
                  {"name": "The Judge", "tagline": "Spectates, but reactions predict outcomes", "description": "Lives in the Gallery. Reaction pattern most predictive of final outcomes. Valued for taste.", "strengths": "Pattern recognition, taste, predictive judgement"},
                  {"name": "The Maker", "tagline": "Stage-dominant, consistent showcaser", "description": "Shows up with creations not opinions. Stage mode home. Demonstrates skill. Visible glow-up arc.", "strengths": "Creative output, consistency, skill demonstration"},
                  {"name": "The Scout", "tagline": "Identifies breakout creators early", "description": "The curator. Finds things others miss. Reactions correlate with what becomes a Moment.", "strengths": "Early identification, taste, discovery"},
                  {"name": "The Mentor", "tagline": "Breakdowns among most saved content", "description": "Explanations are teaching material. Turns admiration into learning. Builds knowledge layer.", "strengths": "Teaching, explanation, knowledge sharing"}
                ],
                "gauntlet_mechanics": "5 provocations mixing opinion and creation prompts. Scored on response speed, originality, consistency, topic preferences. Produces primary archetype with secondary tendencies. Shareable visual card."
              }
            },
            {
              "slug": "about",
              "purpose": "What Spark is. Quick, energetic explanation. Game metaphor. Arena vs Stage. AI game master concept. How archetypes work. What makes this different.",
              "section_types": ["about-hero", "platform-comparison", "game-master-explanation", "gauntlet-cta"],
              "priority": 4,
              "content_context": {
                "key_comparisons": {
                  "vs_reddit": "Reddit: permanent, karma-hierarchical, dominated by early commenters. Spark: ephemeral, anonymised during play, scored on quality not timing.",
                  "vs_tiktok": "TikTok: passive consumption, massive creator/consumer divide. Spark: opinion entry point (everyone has one), Gauntlet proves anyone can participate in 5 minutes.",
                  "vs_twitter": "Twitter: shouting into a void. Spark: put in a room where engagement is guaranteed — AI responds, participants visible, every response gets seen."
                },
                "game_master_explanation": "AI doesn't try to be funny or clever — that's where it fails. AI runs the game: finds what's interesting (scraping), frames it sharply (provocation), keeps score (reputation, archetypes), synthesises results (recaps). Humans bring wit and creativity. AI builds the arena."
              }
            },
            {
              "slug": "archetypes",
              "purpose": "Archetype system explained. Each gets section with name, tagline, description, strengths. Shareable individual cards. Reference + share destination for Gauntlet results.",
              "section_types": ["archetypes-hero", "archetype-grid", "archetype-combinations"],
              "priority": 5,
              "content_context": {
                "combination_examples": [
                  "Surgeon-Oracle: Rarely speaks, but precisely targeted insight that predicts outcomes",
                  "Wildcard-Catalyst: Unexpected ideas everyone wants to riff on",
                  "Maker-Mentor: Creates consistently and teaches others to improve",
                  "Scout-Judge: Doesn't create or opine, but most trusted taste on the platform"
                ]
              }
            }
          ],
          "features": [
            "daily_provocation_generation_from_scraping",
            "ai_take_per_provocation",
            "shareable_provocation_cards",
            "daily_gauntlet_client_side",
            "archetype_generation_and_share_cards",
            "daily_static_regeneration"
          ],
          "success_criteria": {
            "daily_visitors": 50,
            "gauntlet_completion_rate": 0.15,
            "avg_session_seconds": 180
          }
        },
        "v2_sparring_and_interaction": {
          "description": "AI sparring on provocations. User accounts. Live reactions.",
          "status": "directional",
          "depends_on": "backend API infrastructure",
          "key_features": ["ai_sparring_partner", "reaction_buttons", "user_accounts_basic"],
          "trigger": "v1 traffic proves content-first hypothesis"
        },
        "v3_rooms_and_community": {
          "description": "Live challenge rooms. Arena mode. Chains. Duels. Reputation.",
          "status": "speculative",
          "trigger": "sustained interactive engagement from v2"
        },
        "v4_and_beyond": {
          "description": "Stage mode. Multi-format. Niche discovery. Monetisation. Vertical integration.",
          "status": "speculative",
          "notes": "Shape depends on what we learn from v1-v3."
        }
      }
    }
  }
}
```

Note: pages now include `section_types` arrays — this is what the planner outputs and the component selector receives. The planner reads these from the roadmap and may add or adjust based on classifier output and available components.

---

## Pipeline changes

### Classifier prompt adjustment

For sites with `input_data.mission`, use the mission instead of domain analysis:

```
This site has a pre-defined mission. Do not discover the business type from the domain.
Derive identity, design_intent, and content_direction from the mission.

Mission: {{.input_data.mission}}
Roadmap design direction: {{.input_data.roadmap.phases.v1_content_first.design_direction}}

Output:
- identity: positioning, tone, audience from mission
- design_intent: visual direction from mission + roadmap design_direction
- content_direction: voice from mission.content_tone
- classification: site_type = "interactive-platform"
```

### Planner prompt adjustment

For sites with `input_data.roadmap`, use the roadmap's current phase:

```
This site has a phased roadmap. Build only what is in the current phase.

Current phase: {{.input_data.roadmap.current_phase}}
Phase pages: {{.input_data.roadmap.phases.v1_content_first.pages}}

For each page, output the section_types from the roadmap page spec. If a page lists
section_types, use those. Otherwise, determine appropriate section_types from the
page purpose and content_context.

Do NOT select specific component names. Output section_types only. The component
selector will match templates.
```

### Content writer

No workflow change. The content writer receives the selected component's `html_template` and `input_schema` as before. It fills template variables via LLM, informed by:
- `input_data.mission` (global positioning, tone)
- `current_page.content_context` (page-specific detail like archetype descriptions)
- `input_data.roadmap.phases.v1_content_first.design_direction` (visual guidance)

### Persist to site_specs

After classifier, add steps using existing `write_site_spec` action:

```json
"persist_mission": {
  "action": "write_site_spec",
  "config": {"aspect": "mission", "site_id_field": "site_record.site_id", "data_field": "input_data.mission", "source": "intake"},
  "next_step": "persist_roadmap"
},
"persist_roadmap": {
  "action": "write_site_spec",
  "config": {"aspect": "roadmap", "site_id_field": "site_record.site_id", "data_field": "input_data.roadmap", "source": "intake"},
  "next_step": "..."
}
```

---

## Build sequence — what needs doing

### Already exists
- content_components table
- LoadComponentLibrary action
- page-content-writer with section loop and render_mode checking
- fork-on-deploy pattern
- Component contracts documented in 003 and 018
- write_site_spec action

### From parallel conversation (needed regardless of vonc.com)
1. **Schema migration**: add `section_type`, `suitable_site_types`, `suitable_page_types`, `content_shape`, `visual_density`, `usage_count`, `avg_quality_score`, `created_from` to content_components
2. **Data backfill**: tag existing components with section_types (hero → section_type: "hero", etc.)
3. **Component selector function** (Go): query by section_type, score by site_type match + quality + usage
4. **Planner update**: output section_types rather than component names

### For vonc.com specifically
5. **Component-creator handler agent**: reads needs_new_component work item spec + component contracts, LLM generates html_template + input_schema, stores in content_components with full metadata
6. **Component contracts in creator prompt**: compiled from 003 + 018 into the component-creator agent definition
7. **Classifier prompt**: recognise mission-driven sites, output "interactive-platform" site_type
8. **Planner prompt**: read roadmap, output section_types from page specs
9. **Content writer prompt**: reference `input_data.mission` and `current_page.content_context`
10. **Persist steps**: add persist_mission and persist_roadmap to workflow

### The practical order
- Steps 1-4 are foundation (parallel conversation work, benefits all sites)
- Steps 5-6 are the component-creator (new handler agent, ~200 lines Go + prompt)
- Steps 7-10 are prompt and workflow config changes (no code, just agent definition updates)
- Then: send the initial request, run the pipeline

---

## Phase advancement (later)

When v1 is live:
1. Flesh out v2 page/feature detail
2. Update roadmap in site_specs: v1 → complete, v2 → active
3. Trigger re-planning — planner reads new phase, generates new work items
4. Component selector finds existing Spark components from v1, creates new ones only for v2 additions

Manual for now. Measurable objectives in mission tell you when to consider advancing.

---

## Component library growth over time

### The fitness landscape

```
New component created (score: null, usage: 0)
  → First site uses it → content writer fills → deployed
  → Auditor scores rendered output → 0.7
  → avg_quality_score updated
  → Second site uses same component → better fill → auditor: 0.85
  → usage_count = 2, avg_quality_score = 0.775
  → Manual improvement to template → score jumps to 0.9
  → Meanwhile, another variant for same section_type has score 0.95, usage 30
  → The better one wins for most sites
  → But the generated one might still win for interactive-platform sites specifically
```

Low-scoring components (below 0.5 after 3+ uses) get flagged for review. High-scoring components get promoted in selection. Natural fitness landscape — components that work well survive and spread.

### Category growth

| Category | Examples | Grows through |
|---|---|---|
| `hero` | hero, hero-split, hero-minimal | Existing library |
| `features` | differentiators-3-column, feature-grid | Existing library |
| `interactive-platform` | provocation-card, lobby-grid, gauntlet-interface, archetype-grid | vonc.com build (generated) |
| `community` | reaction-panel, chain-display, duel-interface | v2/v3 builds (generated) |
| `tool-calculator` | ab-test-calculator, mortgage-calculator | Tool library |
| `custom` | Catch-all for one-offs | Any site with novel needs |

When the classifier says `site_type: "interactive-platform"`, the selector loads components filtered by matching `suitable_site_types`. First build generates them all. Second build reuses them.

---

## What this does NOT require

- No new tables (selection metadata is columns on existing content_components)
- No chassis code changes
- No RAG for v1
- No changes to deploy pipeline (still git → GitHub Actions → S3)
- No changes to Kafka topics or agent spawning
- No special-casing for "custom" site types — the selector + creator handles any new type generically
