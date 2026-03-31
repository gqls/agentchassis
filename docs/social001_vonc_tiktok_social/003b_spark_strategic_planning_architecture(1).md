# Spark Strategic Planning Architecture

## Approach

No new tables. No chassis code changes. The mission and roadmap travel as `input_data` fields in the initial domain request — the same mechanism already used for `domain` and `objective`. The classifier and planner read them from collected_data. A `write_site_spec` step persists them during the build for future maintenance cycles.

The full concept document (15,000+ words) is too large for input_data. It goes into the existing RAG infrastructure. Content writers search it for relevant sections per page.

## What goes where

| Data | Where | Size | Who reads it | When |
|---|---|---|---|---|
| Mission (positioning, differentiators, tone, objectives) | `input_data.mission` → persisted to `site_specs` aspect `'mission'` | ~1,000 tokens | Classifier, planner | At build time from input_data; future cycles from site_specs |
| Roadmap (current phase, pages, features) | `input_data.roadmap` → persisted to `site_specs` aspect `'roadmap'` | ~500 tokens (v1 only) | Planner | At build time from input_data; future cycles from site_specs |
| Full concept document | RAG (existing infrastructure) | ~15,000 words | Content writers (search per page) | During content writing |
| Objective (short human instruction) | `input_data.objective` | ~100 words | Intake orchestrator, classifier | At build time |

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
        "AI as game master and producer, never the performer",
        "Opinion-first entry point — lower barrier than creation-first",
        "Ephemeral challenges with permanent emergent reputation",
        "Rooms not feeds — enter a space, don't scroll a stream",
        "Arena (competitive) + Stage (showcase) dual modes"
      ],
      "target_users": {
        "primary": "People who have opinions on current trends but no platform that asks for them",
        "secondary": "Creators wanting skill showcase with built-in audience and niche discovery",
        "tertiary": "Curators and tastemakers — people who find things, not make things"
      },
      "content_tone": {
        "is": ["interesting", "slightly adult", "competitive", "aspirational", "playful", "game-like"],
        "is_not": ["dark", "gruesome", "violent", "xxx", "newspaper news", "corporate", "generic social media"],
        "ai_role": "Producer not performer. AI frames, scores, curates, synthesises. AI does not do humour, hot takes, sarcasm, or anything meant to feel spontaneous."
      },
      "measurable_objectives": [
        {
          "id": "obj_dau_50",
          "objective": "50 daily active users within 4 weeks of launch",
          "metric": "daily_active_users",
          "target": 50,
          "status": "not_started"
        },
        {
          "id": "obj_gauntlet_rate",
          "objective": "Daily Gauntlet completion rate above 15% of visitors",
          "metric": "gauntlet_completion_rate",
          "target": 0.15,
          "status": "not_started"
        },
        {
          "id": "obj_session_duration",
          "objective": "Average session duration above 3 minutes",
          "metric": "avg_session_duration_seconds",
          "target": 180,
          "status": "not_started"
        },
        {
          "id": "obj_share_rate",
          "objective": "At least 5% of Gauntlet completions result in a share",
          "metric": "gauntlet_share_rate",
          "target": 0.05,
          "status": "not_started"
        }
      ]
    },
    "roadmap": {
      "current_phase": "v1_content_first",
      "phases": {
        "v1_content_first": {
          "description": "Static S3 site. Provocations as SEO content. Daily Gauntlet as viral archetype quiz. Engage first, understand second, commit third.",
          "hosting": "s3_static",
          "status": "active",
          "pages": [
            {
              "slug": "index",
              "purpose": "Single provocation card filling the screen. Text input beneath. Timer ticking. No signup wall. No mission statement. The product IS the landing page. Underneath: the AI's take (decent but beatable). Below that: today's other provocations (lobby in miniature).",
              "priority": 1
            },
            {
              "slug": "provocations",
              "purpose": "Today's active provocations with AI takes. Each provocation has its own shareable URL. Functions as SEO content and social sharing source. Regenerated daily from scraping pipeline.",
              "priority": 2
            },
            {
              "slug": "gauntlet",
              "purpose": "The Daily Gauntlet — 5 provocations, 5 minutes, discover your Spark archetype. Client-side JS. Produces shareable archetype card. This is the viral Trojan horse — personality quiz mechanic.",
              "priority": 3
            },
            {
              "slug": "about",
              "purpose": "What Spark is. The game master concept. Not a manifesto — a quick, energetic explanation that makes someone want to try it. Explain Arena + Stage. Explain why this isn't another social app.",
              "priority": 4
            },
            {
              "slug": "archetypes",
              "purpose": "The archetype system explained. Surgeon, Wildcard, Oracle, Catalyst, Judge, Maker, Scout, Mentor. Each with description and shareable card format. Functions as reference and share destination.",
              "priority": 5
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
          "design_direction": "Game energy, not corporate. Dark theme with vibrant accents. Cards, timers, energy indicators. Should feel like entering an arena lobby, not reading a blog. Mobile-first."
        },
        "v2_sparring_and_interaction": {
          "description": "Add AI sparring to provocations. User accounts. Live reactions. The static site becomes interactive.",
          "status": "directional",
          "depends_on": "backend API infrastructure",
          "key_features": ["ai_sparring_partner", "reaction_buttons", "user_accounts_basic", "exchange_storage"],
          "trigger": "v1 traffic proves the content-first hypothesis"
        },
        "v3_rooms_and_community": {
          "description": "Live challenge rooms. Arena mode. Chain reactions. Duels. The full competitive social experience.",
          "status": "speculative",
          "key_features": ["live_rooms", "arena_challenges", "chains_and_remixes", "duels", "reputation_system"],
          "trigger": "sustained interactive engagement from v2"
        },
        "v4_and_beyond": {
          "description": "Stage mode. Multi-format. Niche discovery. Monetisation. Vertical integration.",
          "status": "speculative",
          "notes": "Shape of this depends entirely on what we learn from v1-v3. Don't plan in detail."
        }
      }
    }
  }
}
```

## What changes in the pipeline

### Classifier prompt adjustment

The classifier currently discovers what a site should be from domain analysis. For sites that arrive with a `mission` in input_data, it should use the mission instead of guessing.

Prompt addition (conditional — only when input_data.mission exists):

```
This site has a pre-defined mission and strategic direction. Do not attempt to discover the business type from the domain name. Instead, derive the identity, design intent, and content direction from the provided mission.

Mission: {{.input_data.mission}}

Produce:
- identity: company positioning, tone, audience — derived from mission.target_users and mission.content_tone
- design_intent: visual direction — derived from mission positioning and roadmap.design_direction
- content_direction: voice, emphasis, things to avoid — derived from mission.content_tone
- classification: site type = "interactive_platform", archetype = "content_first_launch"
```

No code change. Just a prompt update to the classifier agent definition's `default_config`. The `{{.input_data.mission}}` template already works — collected_data makes input_data available to templates.

### Planner prompt adjustment

The planner currently decides what pages to build based on the classifier output and business type. For sites with a `roadmap` in input_data, it should use the roadmap's current phase.

Prompt addition (conditional):

```
This site has a phased roadmap. Build only what is in the current phase.

Current phase: {{.input_data.roadmap.current_phase}}
Phase definition: {{.input_data.roadmap.phases.v1_content_first}}

Use the phase's pages array as the page list. Use the phase's features list to understand what interactive elements are needed. Use the phase's design_direction for layout guidance.
```

Again, no code change. Prompt update to the planner agent definition.

### Persist to site_specs (new workflow step — uses existing action)

After the classifier produces specs, add a step that writes the mission and roadmap to site_specs for future reference:

```json
"persist_mission": {
  "action": "write_site_spec",
  "config": {
    "aspect": "mission",
    "site_id_field": "site_record.site_id",
    "data_field": "input_data.mission",
    "source": "intake"
  },
  "next_step": "persist_roadmap"
},
"persist_roadmap": {
  "action": "write_site_spec",
  "config": {
    "aspect": "roadmap",
    "site_id_field": "site_record.site_id",
    "data_field": "input_data.roadmap",
    "source": "intake"
  },
  "next_step": "..."
}
```

This means future maintenance cycles, re-planning, and audit agents can read the mission and roadmap from site_specs without needing the original input_data.

### Content writer RAG query

The content writer workflow already includes a research step. For vonc.com pages, the research step searches RAG for Spark concept context relevant to the page being written.

No workflow change needed — the research step's search query just needs to be informed by the page purpose. The page purpose comes from the roadmap (now in collected_data), so the content writer can construct a query like: "Spark landing page provocation card first 5 seconds engage first understand second."

## What needs doing (ordered)

1. **Ingest concept doc into RAG** — chunk concept_spark.md and store via existing pipeline
2. **Update classifier agent definition** — add conditional prompt for sites with mission in input_data
3. **Update planner agent definition** — add conditional prompt for sites with roadmap in input_data
4. **Add persist_mission and persist_roadmap steps** to the intake/builder workflow (using existing write_site_spec action)
5. **Send the initial request** — the JSON above, via the existing Kafka CLI mechanism
6. **Run the pipeline** — classifier → planner → content writers → deploy to S3

Steps 1-4 are preparation. Step 5 is one kubectl command. Step 6 is the existing pipeline running with better inputs.

## Phase advancement (later)

When v1 is live and you want to move to v2:

1. Update the roadmap in site_specs: set v1 status to `complete`, v2 to `active`, update `current_phase`
2. Flesh out v2's page/feature detail (it's currently directional)
3. Trigger re-planning — planner reads the new current phase, generates new work items

This is a manual process for now. The measurable objectives in the mission aspect tell you when to consider advancing. A future strategic review agent could automate the comparison of actuals vs targets.

## What this does NOT require

- No new tables
- No chassis code changes
- No new actions (write_site_spec already exists)
- No changes to the deploy pipeline
- No changes to the Kafka topic structure
- No changes to how agents are spawned or communicate
