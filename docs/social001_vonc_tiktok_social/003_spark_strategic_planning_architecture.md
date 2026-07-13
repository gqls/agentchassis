# Spark Strategic Planning Architecture

## Problem

The Spark concept document is a 15,000+ word product strategy that operates on a completely different timescale and abstraction level from anything the current site-building pipeline handles. The pipeline builds websites from specs. The Spark document describes a product vision with phases, changing priorities, conditional strategies, and objectives that shift as we learn.

It's not site configuration. It's not research output. It's not a brief. It's a living product roadmap that happens to have a website as one of its outputs.

The current pipeline is stateless in terms of strategy — classifier runs, produces specs, planner builds work items, agents execute, done. There's no mechanism for "we said we'd measure X, three weeks passed, we didn't hit it, so now change Y."

## What agents actually need (not the same thing)

Different agents need different slices of the concept at different times:

| Agent | What it needs | How much | When |
|---|---|---|---|
| Classifier | Positioning, tone, audience, visual direction | ~500 words of stable summary | Once per classification |
| Planner | What pages/features to build in current phase | Structured phase definition | Per build cycle |
| Content writers | Enough context to write aligned copy | ~2,000 words of relevant sections | Per page |
| Strategic review (future) | Objectives, metrics, phase conditions | Structured objectives + actuals | Periodic |

No agent needs the full 15,000 words loaded as a blob. The concept doc should be decomposed into the structures the system already has.

## The Decomposition

### 1. Mission — in site_specs, aspect 'mission'

Stable, structured. The elevator pitch. Changes rarely. The classifier reads this and produces identity/design_intent/content_direction.

```json
{
  "mission": "AI-driven social platform where the world is the content and your take is the game",
  "positioning": "Not another social app. A daily game powered by what's happening in the world. You don't post. You don't scroll. You play.",
  "tagline": "TikTok is where you watch. Twitter is where you shout. Spark is where you play.",
  "key_differentiators": [
    "AI as game master and producer, never the performer",
    "Opinion-first entry point — lower barrier than creation-first",
    "Ephemeral challenges with permanent emergent reputation",
    "Rooms not feeds — enter a space, don't scroll a stream",
    "Arena (competitive) + Stage (showcase) dual modes",
    "Radical anonymity during Arena challenges, reveal after"
  ],
  "target_users": {
    "primary": "People who have opinions on current trends but no platform that asks for them",
    "secondary": "Creators wanting skill showcase with built-in audience and niche discovery",
    "tertiary": "Curators and tastemakers — people who find things, not make things"
  },
  "content_tone": {
    "is": ["interesting", "slightly adult", "competitive", "aspirational", "playful"],
    "is_not": ["dark", "gruesome", "violent", "xxx", "newspaper news"],
    "ai_role": "Producer not performer. AI frames, scores, curates, synthesises. AI does not do humour, hot takes, sarcasm, or anything meant to feel spontaneous."
  },
  "measurable_objectives": [
    {
      "id": "obj_dau_50",
      "objective": "Achieve 50 daily active users within 4 weeks of launch",
      "metric": "daily_active_users",
      "target": 50,
      "timeframe_weeks": 4,
      "status": "not_started"
    },
    {
      "id": "obj_gauntlet_rate",
      "objective": "Daily Gauntlet completion rate above 15% of visitors",
      "metric": "gauntlet_completion_rate",
      "target": 0.15,
      "timeframe": "ongoing",
      "status": "not_started"
    },
    {
      "id": "obj_session_duration",
      "objective": "Average session duration above 3 minutes",
      "metric": "avg_session_duration_seconds",
      "target": 180,
      "timeframe": "ongoing",
      "status": "not_started"
    },
    {
      "id": "obj_share_rate",
      "objective": "At least 5% of Gauntlet completions result in a share",
      "metric": "gauntlet_share_rate",
      "target": 0.05,
      "timeframe": "ongoing",
      "status": "not_started"
    }
  ]
}
```

### 2. Roadmap — in site_specs, aspect 'roadmap'

Structured, phase-gated build plan. The planner reads the current phase and plans accordingly. When a phase is complete, a human (or eventually a review agent) advances the roadmap and triggers re-planning.

```json
{
  "current_phase": "v1_content_first",
  "phases": {
    "v1_content_first": {
      "description": "Static S3 site. Provocations as content. Gauntlet as viral mechanic. Engage first, understand second, commit third.",
      "entry_condition": "launch",
      "hosting": "s3_static",
      "pages": [
        {
          "slug": "index",
          "purpose": "Single provocation card filling the screen. Text input. Timer. No signup wall. The product IS the landing page.",
          "priority": 1
        },
        {
          "slug": "provocations",
          "purpose": "Today's active provocations with AI takes. SEO content. Each provocation has its own shareable URL.",
          "priority": 2
        },
        {
          "slug": "gauntlet",
          "purpose": "Daily Gauntlet — 5 provocations, discover your Spark archetype. Viral personality quiz mechanic. Client-side JS calling AI API.",
          "priority": 3
        },
        {
          "slug": "about",
          "purpose": "What Spark is. The game master concept. Not a manifesto — a quick explanation that makes someone want to try it.",
          "priority": 4
        },
        {
          "slug": "archetypes",
          "purpose": "The archetype system explained. Surgeon, Wildcard, Oracle, Catalyst, Judge, Maker, Scout, Mentor. Shareable archetype cards.",
          "priority": 5
        }
      ],
      "features": [
        "daily_provocation_generation_from_scraping",
        "ai_take_per_provocation",
        "shareable_provocation_cards",
        "daily_gauntlet_client_side",
        "archetype_generation_and_share_cards",
        "hourly_or_daily_static_regeneration"
      ],
      "success_criteria": {
        "daily_visitors": 50,
        "gauntlet_completion_rate": 0.15,
        "avg_session_seconds": 180
      },
      "status": "active",
      "notes": "Everything is static HTML regenerated by the agent pipeline. Interactive elements (Gauntlet, AI sparring) use client-side JS calling backend APIs. No user accounts yet."
    },
    "v1b_sparring": {
      "description": "Add AI sparring partner to provocations. Visitors can argue with AI. Exchanges stored for next visitors to see.",
      "entry_condition": "v1 deployed and basic traffic flowing",
      "depends_on": "backend_api_endpoint",
      "pages_additions": [
        {
          "slug": "provocation_detail",
          "purpose": "Individual provocation page with AI sparring interface. Shows previous exchanges. React to others' responses."
        }
      ],
      "features": [
        "ai_sparring_api_endpoint",
        "exchange_storage_and_display",
        "reaction_buttons_client_side"
      ],
      "success_criteria": {
        "sparring_sessions_per_day": 20,
        "return_visitor_rate": 0.15
      },
      "status": "planned"
    },
    "v2_interactive": {
      "description": "User accounts. Live rooms. Real-time reactions. The full Arena experience for text challenges.",
      "entry_condition": "v1 success criteria met OR 8 weeks elapsed",
      "depends_on": "backend_infrastructure",
      "features": [
        "user_accounts",
        "live_challenge_rooms",
        "real_time_reactions",
        "chain_remix_mechanic",
        "reputation_system_basic",
        "arena_mode_text_only"
      ],
      "success_criteria": {
        "daily_active_users": 200,
        "challenges_with_10plus_responses": 5
      },
      "status": "planned"
    },
    "v3_stage_and_media": {
      "description": "Stage mode. Multi-format (image, audio, video). Niche discovery. Glow-up tracking.",
      "entry_condition": "v2 success criteria met",
      "features": [
        "stage_showcase_mode",
        "multi_format_challenges",
        "niche_detection_and_rooms",
        "glow_up_tracking",
        "tune_in_mechanic",
        "collab_matchmaking"
      ],
      "status": "planned"
    },
    "v4_monetisation": {
      "description": "Creator subscriptions. Brand challenges. Revenue share. Expert consultations.",
      "entry_condition": "v3 active with 500+ DAU",
      "features": [
        "creator_subscription_channels",
        "brand_sponsored_challenges",
        "revenue_share_on_showcases",
        "prediction_staking_reputation_tokens"
      ],
      "status": "future"
    }
  }
}
```

### 3. Concept document — into existing RAG/research infrastructure

The full Spark concept doc gets chunked and stored in the same RAG pipeline used for competitor research, industry analysis, and other context. No new storage silo.

Content writers search for relevant sections: "Spark positioning game master provocation" returns the relevant paragraphs. Same mechanism already used for research-backed content writing.

The concept doc is versioned by re-ingestion — when the doc is updated, re-chunk and re-store. Old versions naturally get superseded.

### 4. Phase advancement — HITL for now, automatable later

You review metrics against the measurable objectives in the mission aspect and the success criteria in each roadmap phase. When conditions are met (or you decide to advance regardless):

1. Update `roadmap` aspect — set current phase's status to `complete`, next phase's status to `active`, update `current_phase`
2. Update relevant `mission` objective statuses
3. Trigger re-planning — the planner reads the new current phase and generates work items for the new features/pages

For now this is manual (update site_specs rows, trigger planner). Later, a scheduled "strategic review" agent could read analytics, compare to success criteria, and either advance automatically or flag for human review.

## How the pipeline changes for vonc.com

### Before any agents run (human setup)

1. Create the site record for vonc.com in the `sites` table
2. Insert the `mission` aspect into `site_specs`
3. Insert the `roadmap` aspect into `site_specs`
4. Ingest the concept doc into RAG

### Classifier behaviour

The classifier reads the `mission` aspect and produces:

- `identity` — derived from mission positioning, target users, content tone
- `design_intent` — derived from "rooms not feeds" visual metaphor, game energy, the lobby concept
- `content_direction` — derived from content tone (interesting not dark, AI as producer not performer)
- `classification` — site type, archetype (this is a new kind of site, not a brochure)

The classifier does NOT need the full concept doc. The mission gives it everything it needs for spec generation. If it needs more context on a specific point (e.g. what "rooms not feeds" means visually), that's a RAG query.

### Planner behaviour

The planner reads:
- `mission` — to understand objectives and priorities
- `roadmap` — to know what phase we're in and what pages/features to plan
- Classifier output (`identity`, `design_intent`, `content_direction`) — to ensure implementation matches intent

For v1_content_first, the planner produces work items for: landing page, provocations feed, gauntlet page, about page, archetypes page. It knows from the roadmap that these are S3-hosted static pages with client-side JS for interactive elements.

### Content writer behaviour

The content writer reads:
- `identity` and `content_direction` from site_specs (tone, voice, audience)
- Relevant sections from the concept doc via RAG (for rich context on specific topics)
- The page's purpose from the roadmap (what this specific page needs to achieve)

When writing the landing page, the writer pulls RAG context about "engage first understand second commit third" and "single provocation card filling the screen" to understand what the page should feel like.

## What this does NOT require

- No new tables (mission and roadmap are site_spec aspects; concept doc goes in existing RAG)
- No changes to the agent chassis or workflow engine
- No new actions beyond what exists (write_site_spec, read_site_spec, RAG search)
- No changes to the deploy pipeline (still git → GitHub Actions → S3)

## What needs doing

1. **Define the site record** for vonc.com in the sites table
2. **Write the mission aspect** — structured JSON as above, INSERT into site_specs
3. **Write the roadmap aspect** — structured JSON as above, INSERT into site_specs
4. **Ingest concept doc into RAG** — chunk the markdown, store via existing RAG pipeline
5. **Update classifier prompt** — teach it to read the mission aspect and produce aligned specs instead of guessing from domain name. For sites with a mission aspect, skip domain-analysis discovery and use the mission directly.
6. **Update planner** — teach it to read the roadmap's current phase and use the phase's pages/features as its planning input instead of (or in addition to) the classifier's page suggestions
7. **Run the pipeline** — classifier → planner → content writers → deploy to S3

Steps 1-4 are data setup. Steps 5-6 are prompt/workflow adjustments to existing agents. Step 7 is running what already exists with better inputs.

## Future: automated strategic review

When analytics are in place, a scheduled agent could:

1. Read measurable objectives from mission aspect
2. Query analytics for actual metrics
3. Compare actuals vs targets
4. If phase success criteria met → propose phase advancement (HITL approval or auto-advance)
5. If objectives missed → flag for review, suggest strategy adjustments
6. Write updated mission/roadmap aspects with new status values

This closes the loop: strategy → build → measure → adjust → strategy. But it's not needed for v1. Manual review is fine while learning what works.
