# Spark Strategic Planning Architecture

## Approach

No new tables. No chassis code changes. No RAG for v1. The mission, roadmap, and per-page content context travel as `input_data` fields in the initial domain request — the same mechanism already used for `domain` and `objective`. All agents read them from collected_data. A `write_site_spec` step persists them during the build for future maintenance cycles.

The full concept document (15,000+ words) stays as an offline reference. Content writers get what they need through structured `content_context` fields on each page spec in the roadmap. RAG becomes relevant in v2+ when the content surface area outgrows what fits in structured input_data.

## What goes where

| Data | Where | Who reads it |
|---|---|---|
| Mission (positioning, differentiators, tone, objectives) | `input_data.mission` → persisted to `site_specs` aspect `'mission'` | Classifier, planner, content writers |
| Roadmap (current phase, pages with purposes and content context) | `input_data.roadmap` → persisted to `site_specs` aspect `'roadmap'` | Planner, content writers |
| Per-page detail (archetype descriptions, mechanic explanations) | `input_data.roadmap.phases.v1_content_first.pages[].content_context` | Content writers (for that specific page) |
| Full concept document | Offline reference (not ingested) | Humans; agents in v2+ via RAG |
| Objective (short human instruction) | `input_data.objective` | Intake orchestrator, classifier |

## Why not RAG for v1?

V1 has 5 pages. The content writers already get from collected_data:
- The positioning, tagline, differentiators, target users, content tone (from mission)
- Each page's specific purpose — detailed enough to write from (from roadmap)
- Page-specific content context where needed — e.g. archetype descriptions for the archetypes page (from roadmap)
- Classifier output: identity, design_intent, content_direction

That's sufficient context for 5 pages without chunking, embedding, or knowledge_base infrastructure. The `rag_lookup` action already exists — adding it to the content writer workflow later is straightforward when the content surface area grows.

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
        "the_game_metaphor": "Everything is framed as a game. Challenges have timers, rules, scores, winners. 'What game are you playing today?' not 'what are you posting today?'",
        "engage_first": "No signup walls, no explanations. A provocation card and a text input. The product sells itself in 60 seconds of use. Engage first, understand second, commit third.",
        "ai_slop_awareness": "AI is restricted to what it's genuinely good at: framing, pattern recognition, synthesis, scoring. AI is deliberately kept away from humour, hot takes, anything meant to feel spontaneous. Deliberately bad AI takes can be used as challenge seeds — AI as straight man, humans as comedians.",
        "cold_start": "Solo mode is fully functional for 1 person. AI sparring partner fills the room transparently. Every exchange stored for next arrivals. The site is never empty — scraping runs 24/7, AI generates baseline responses clearly labelled.",
        "archetype_system": "Your profile emerges from behaviour — response speed, topic preferences, analytical vs creative lean, reaction patterns, prediction accuracy. Produces personality-type labels (Surgeon, Wildcard, Oracle, etc.) that shift over time. The Daily Gauntlet is the entry point — 5 provocations, 5 minutes, get your type. Viral mechanic: 'I'm a Surgeon-Wildcard, what are you?'"
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
          "design_direction": "Game energy, not corporate. Dark theme with vibrant accents. Cards, timers, energy indicators. Should feel like entering an arena lobby, not reading a blog. Mobile-first. The landing page should feel like something is happening right now.",
          "pages": [
            {
              "slug": "index",
              "purpose": "The product IS the landing page. Not a marketing page about Spark — the actual experience. A single provocation card filling the screen, sourced from today's real trends, framed sharply. Text input beneath. Timer ticking. No signup wall. No mission statement. Below the main card: 'The AI's take' — a response that's decent but beatable. Below that: today's other active provocations. Scroll further: brief explanation of what this is and invitation to try the Daily Gauntlet.",
              "priority": 1
            },
            {
              "slug": "provocations",
              "purpose": "Today's active provocations with AI takes. Each provocation has its own shareable URL. Functions as SEO content and social sharing source. Regenerated daily from scraping pipeline. Card-based layout. Each card shows: the provocation text, the AI's take, a timer (or 'closed' badge), reaction counts, and a 'what's your take?' call to action.",
              "priority": 2
            },
            {
              "slug": "gauntlet",
              "purpose": "The Daily Gauntlet — 5 provocations in 5 minutes. Discover your Spark archetype. Client-side JavaScript calling AI APIs. The viral Trojan horse — personality quiz mechanic. At the end, produces a shareable archetype card showing your type and key stats. 'I'm a Surgeon-Wildcard. What are you?' format.",
              "priority": 3,
              "content_context": {
                "archetypes": [
                  {
                    "name": "The Surgeon",
                    "tagline": "Precise, analytical, cuts to the core",
                    "description": "Strips arguments to their essential structure. Fast on topics with clear logical frameworks. Finds the structural flaw others miss. Can occasionally overthink simple questions.",
                    "strengths": "Analytical speed, structural clarity, finding hidden flaws",
                    "arena_style": "Devastating counterarguments, precise rebuttals"
                  },
                  {
                    "name": "The Wildcard",
                    "tagline": "Unpredictable, jumps between topics, highest remix rate",
                    "description": "Takes other people's ideas in unexpected directions. Crosses domains freely. Their responses generate surprise more than agreement. The most likely to turn a serious provocation into something nobody expected.",
                    "strengths": "Creative leaps, remix ability, cross-domain thinking",
                    "arena_style": "Lateral thinking, unexpected angles, genre-bending responses"
                  },
                  {
                    "name": "The Oracle",
                    "tagline": "Quiet but accurate, rare but impactful",
                    "description": "Doesn't respond often. But when they do, it tends to land. Best prediction accuracy. Picks battles carefully. Their silence makes their contributions feel weighty.",
                    "strengths": "Prediction accuracy, timing, selective engagement",
                    "arena_style": "Few words, high impact. Often the final word in a thread."
                  },
                  {
                    "name": "The Catalyst",
                    "tagline": "Responses generate the longest chains",
                    "description": "Their individual responses might not be the best, but they spark others. They ask the question that opens a new direction. They make the observation that 5 people want to build on. The chain-starter.",
                    "strengths": "Provocation, opening new angles, generating momentum",
                    "arena_style": "Seeds ideas that others develop. The assist, not the goal."
                  },
                  {
                    "name": "The Judge",
                    "tagline": "Mostly spectates, but their reactions predict outcomes",
                    "description": "Lives in the Gallery. Watches more than participates. But their reaction pattern is the most predictive of final challenge outcomes. When The Judge reacts 'Genius,' the crowd tends to follow. Valued for taste.",
                    "strengths": "Pattern recognition, taste, predictive judgement",
                    "arena_style": "Reacts rather than responds. Influence through curation."
                  },
                  {
                    "name": "The Maker",
                    "tagline": "Stage-dominant, consistent showcaser",
                    "description": "Shows up with creations, not opinions. Stage mode is their home. They demonstrate skill rather than argue positions. Growing niche following from consistent showcase quality. Their glow-up arc is visible.",
                    "strengths": "Creative output, consistency, skill demonstration",
                    "stage_style": "Regular showcaser. Quality improves visibly over time."
                  },
                  {
                    "name": "The Scout",
                    "tagline": "Identifies breakout creators before anyone else",
                    "description": "The curator. Finds things others miss. Their 'Fire' reactions on Stage showcases have the highest correlation with what later becomes a Moment. Valued for discovery, not creation.",
                    "strengths": "Early identification, taste, discovery",
                    "stage_style": "Reactions that signal quality. The A&R of the platform."
                  },
                  {
                    "name": "The Mentor",
                    "tagline": "Breakdowns among the most saved content",
                    "description": "When they explain how they did something, people save it. Their behind-the-scenes breakdowns are teaching material. They turn admiration into learning. Naturally builds a knowledge layer beneath the showcase layer.",
                    "strengths": "Teaching, explanation, knowledge sharing",
                    "stage_style": "Teach Me reactions trigger detailed breakdowns. Turns watchers into practitioners."
                  }
                ],
                "gauntlet_mechanics": "5 provocations mixing Arena-style opinion prompts and Stage-style creation prompts. Scored on: response speed, originality of angle (compared to predicted common response), consistency across the 5, topic preference patterns. Produces a primary archetype with secondary tendencies. Results are shareable as a visual card."
              }
            },
            {
              "slug": "about",
              "purpose": "What Spark is. Not a manifesto — a quick, energetic explanation that makes someone want to try it. Cover: the game metaphor (you play, not post), Arena vs Stage, the AI game master concept (producer not performer), how archetypes work, what makes this different from Reddit/TikTok/Twitter. End with a strong call to action back to the Gauntlet or today's provocations.",
              "priority": 4,
              "content_context": {
                "key_comparisons": {
                  "vs_reddit": "Reddit asks 'what do you think?' but responses are permanent, hierarchical (karma), and dominated by early commenters. Spark's challenges are ephemeral, anonymised during play, and scored on quality not timing.",
                  "vs_tiktok": "TikTok is passive consumption with a massive creator/consumer divide. Spark's entry point is opinion (everyone has one) not creation (intimidating). The Gauntlet proves anyone can participate in 5 minutes.",
                  "vs_twitter": "Twitter is shouting into a void hoping for engagement. Spark puts you in a room where engagement is guaranteed — the AI responds immediately, other participants are visible, the challenge structure means every response gets seen."
                },
                "the_game_master_explanation": "The AI doesn't try to be funny, creative, or clever — that's where AI fails and users smell the slop. Instead, the AI runs the game. It finds what's interesting in the world right now (scraping real trends), frames it as a sharp provocation (the question, not the answer), keeps score (reputation, streaks, archetype tracking), and synthesises what happened (recaps, crowd synthesis). Humans bring the wit, the creativity, the hot takes. The AI builds the arena around them."
              }
            },
            {
              "slug": "archetypes",
              "purpose": "The archetype system explained. Each archetype gets a section with name, tagline, full description, strengths, and play style. Shareable individual archetype cards. This page serves as both reference and share destination — when someone shares their Gauntlet result, the link goes here. Explain that archetypes are earned from real behaviour, shift over time, and combine (Surgeon-Wildcard, Oracle-Judge, etc.).",
              "priority": 5,
              "content_context": {
                "note": "Full archetype details are in the gauntlet page content_context. Reference the same data. This page expands on each archetype with more detail, example scenarios, and how combinations work.",
                "combination_examples": [
                  "Surgeon-Oracle: Rarely speaks, but when they do, it's a precisely targeted insight that predicts the outcome",
                  "Wildcard-Catalyst: Throws unexpected ideas into the room that everyone wants to riff on",
                  "Maker-Mentor: Creates consistently and teaches others how to improve",
                  "Scout-Judge: Doesn't create or opine, but has the most trusted taste on the platform"
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
          },
          "notes": "Everything is static HTML regenerated by the agent pipeline. Interactive elements (Gauntlet, future AI sparring) use client-side JS calling backend APIs. No user accounts yet."
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
          "notes": "Shape depends entirely on what we learn from v1-v3. Don't plan in detail."
        }
      }
    }
  }
}
```

## What changes in the pipeline

### Classifier prompt adjustment

For sites arriving with a `mission` in input_data, the classifier uses the mission instead of guessing from the domain name.

Prompt addition (conditional — only when input_data.mission exists):

```
This site has a pre-defined mission and strategic direction. Do not attempt to discover the business type from the domain name. Instead, derive the identity, design intent, and content direction from the provided mission.

Mission: {{.input_data.mission}}

Produce:
- identity: positioning, tone, audience — derived from mission.target_users and mission.content_tone
- design_intent: visual direction — derived from mission.positioning and roadmap design_direction
- content_direction: voice, emphasis, things to avoid — derived from mission.content_tone and mission.ai_role
- classification: site_type = "interactive_platform", archetype = "content_first_launch"
```

No code change. Prompt update to the classifier agent definition's `default_config`.

### Planner prompt adjustment

For sites with a `roadmap` in input_data, the planner uses the roadmap's current phase instead of inventing pages.

Prompt addition (conditional):

```
This site has a phased roadmap. Build only what is in the current phase.

Current phase: {{.input_data.roadmap.current_phase}}
Phase definition: {{.input_data.roadmap.phases.v1_content_first}}

Use the phase's pages array as the page list. Each page has a purpose describing what it should achieve and optional content_context with specific details the content writer needs.
```

No code change. Prompt update to the planner agent definition.

### Content writer context

The content writer already receives `input_data` through collected_data. For vonc.com, each page's `content_context` provides specific details:

- **Landing page**: no content_context needed — the purpose is detailed enough
- **Provocations**: no content_context needed — mostly structural
- **Gauntlet**: content_context includes full archetype definitions and gauntlet mechanics
- **About**: content_context includes competitor comparisons and game master explanation
- **Archetypes**: content_context includes combination examples; references gauntlet archetype data

The content writer's LLM prompt template already has access to the current page from collected_data. Adding `{{.current_page.content_context}}` to the prompt gives it everything page-specific. The mission in `{{.input_data.mission}}` gives it the global positioning, tone, and differentiators.

No workflow change needed. Just the prompt template for the content writer needs to reference these fields.

### Persist to site_specs (new workflow steps — uses existing action)

After the classifier produces specs, add steps that write mission and roadmap to site_specs:

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

Future maintenance cycles and re-planning read from site_specs without needing the original input_data.

## What needs doing (ordered)

1. **Update classifier agent definition** — add conditional prompt for sites with mission in input_data
2. **Update planner agent definition** — add conditional prompt for sites with roadmap in input_data
3. **Update content writer prompt** — reference `input_data.mission` and `current_page.content_context`
4. **Add persist_mission and persist_roadmap steps** to the intake/builder workflow
5. **Send the initial request** — the JSON above, via existing Kafka CLI mechanism
6. **Run the pipeline** — classifier → planner → content writers → deploy to S3

Steps 1-4 are prompt and workflow config updates. Step 5 is one kubectl command. Step 6 is the existing pipeline.

## Phase advancement (later)

When v1 is live and you want to move to v2:

1. Flesh out v2's page/feature detail (currently directional)
2. Update the roadmap in site_specs: v1 status → `complete`, v2 → `active`, update `current_phase`
3. Trigger re-planning — planner reads new current phase, generates new work items

Manual for now. Measurable objectives in mission tell you when to consider advancing.

## RAG integration (v2+)

When the content surface area grows beyond what fits in page-level `content_context` fields:

1. Ingest the concept doc into the knowledge_base via existing `rag_index` action
2. Add a `rag_lookup` step to the content writer workflow before content generation
3. Query constructs from page purpose: `"Spark {{current_page.slug}} {{current_section.function}}"`
4. Results flow into the LLM prompt via `{{.concept_context.rag_context}}`

The infrastructure exists. The action exists. It's just a workflow step addition when needed.

## What this does NOT require

- No new tables
- No chassis code changes
- No new actions
- No RAG infrastructure for v1
- No knowledge_base ingestion for v1
- No changes to the deploy pipeline
- No changes to Kafka topics
- No changes to agent spawning
