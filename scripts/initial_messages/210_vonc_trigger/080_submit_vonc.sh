#!/bin/bash
# ============================================================================
# submit-vonc.sh — Submit vonc.com (Spark) to the build pipeline
# Tier 3 submission: domain + mission + roadmap + briefs
# ============================================================================

# Build JSON via python3 (handles multiline strings in briefs)
MESSAGE_BODY=$(python3 << 'PYEOF'
import json, sys

mission_brief_text = """Spark is an AI-driven social platform where the world is the content and your take is the game.

Positioning: Not another social app. A daily game powered by what is happening in the world. You don't post. You don't scroll. You play.

Tagline: TikTok is where you watch. Twitter is where you shout. Spark is where you play.

Key differentiators:
- AI as game master and producer, never the performer — when something is funny or impressive, a human did it
- Opinion-first entry point — you need a take, not an idea
- Ephemeral challenges with permanent emergent reputation — disposable content, lasting identity
- Rooms not feeds — enter a space with energy, don't scroll a stream
- Arena (competitive takes) + Stage (creative showcase) dual modes
- Archetypes earned from behaviour, not self-reported — your profile emerges from how you engage

Target users:
- Primary: People with opinions on current trends but no platform that asks for them
- Secondary: Creators wanting skill showcase with built-in audience and niche discovery
- Tertiary: Curators and tastemakers — people who find things, not make things

Content tone: interesting, slightly adult, competitive, aspirational, playful, game-like, energetic
Not: dark, gruesome, violent, corporate, generic social media, AI slop
AI role: Producer not performer. AI frames, scores, curates, synthesises. Never does humour, hot takes, sarcasm."""

roadmap_brief_text = """Current phase: v1_content_first
Static S3 site. Provocations as SEO content. Daily Gauntlet as viral archetype quiz.
Design: Game energy, not corporate. Dark theme with vibrant accents. Cards, timers, energy indicators. Arena lobby feel. Mobile-first.

Pages for this phase:
- index: The product IS the landing page. Single provocation card filling screen. Text input beneath. Timer. No signup wall.
  Sections: provocation-card, lobby-grid, brief-explanation, gauntlet-cta
- provocations: Today's active provocations with AI takes. Each has shareable URL. SEO content. Card-based layout.
  Sections: provocation-feed
- gauntlet: Daily Gauntlet — 5 provocations, 5 minutes, discover your archetype. Client-side JS. Produces shareable archetype card.
  Sections: gauntlet-interface, archetype-result-card
- about: What Spark is. Quick, energetic explanation. Game metaphor. Arena vs Stage. AI game master concept. How archetypes work.
  Sections: hero, platform-comparison, game-master-explanation, gauntlet-cta
- archetypes: Archetype system explained. Each gets section with name, tagline, description, strengths. Shareable individual cards.
  Sections: hero, archetype-grid, archetype-combinations

Future phases (not for building now):
- v2: AI sparring on provocations, user accounts, live reactions
- v3: Live challenge rooms, arena mode, chains, duels, reputation
- v4: Stage mode, monetisation, vertical integration"""

msg = {
    "action": "orchestrate",
    "config": {"agent_type": "domain-submitter"},
    "input_data": {
        "domain": "vonc.com",
        "objective": "Build a content-first launch site for Spark, an AI-driven social platform where the world is the content and your take is the game. V1 is static S3 with daily provocation generation and a viral archetype quiz. The landing page IS the product.",

        "mission_brief": {"text": mission_brief_text},
        "roadmap_brief": {"text": roadmap_brief_text},

        "mission": {
            "mission": "AI-driven social platform where the world is the content and your take is the game",
            "positioning": "Not another social app. A daily game powered by what is happening in the world. You don't post. You don't scroll. You play.",
            "tagline": "TikTok is where you watch. Twitter is where you shout. Spark is where you play.",
            "key_differentiators": [
                "AI as game master and producer, never the performer",
                "Opinion-first entry point — you need a take, not an idea",
                "Ephemeral challenges with permanent emergent reputation",
                "Rooms not feeds — enter a space with energy, don't scroll a stream",
                "Arena (competitive takes) + Stage (creative showcase) dual modes",
                "Archetypes earned from behaviour, not self-reported"
            ],
            "target_users": {
                "primary": "People with opinions on current trends but no platform that asks for them",
                "secondary": "Creators wanting skill showcase with built-in audience and niche discovery",
                "tertiary": "Curators and tastemakers — people who find things, not make things"
            },
            "content_tone": {
                "is": ["interesting", "slightly adult", "competitive", "aspirational", "playful", "game-like", "energetic"],
                "is_not": ["dark", "gruesome", "violent", "xxx", "newspaper news", "corporate", "generic social media", "AI slop"],
                "ai_role": "Producer not performer. AI frames, scores, curates, synthesises. Never does humour, hot takes, sarcasm."
            },
            "core_concepts": {
                "the_game_metaphor": "Everything is framed as a game. Challenges have timers, rules, scores, winners.",
                "engage_first": "No signup walls, no explanations. A provocation card and a text input. Engage first, understand second, commit third.",
                "ai_slop_awareness": "AI restricted to genuine strengths. Deliberately bad AI takes as challenge seeds.",
                "cold_start": "Solo mode fully functional for 1 person. AI sparring partner fills the room transparently.",
                "archetype_system": "Profile emerges from behaviour. Daily Gauntlet produces shareable archetype cards."
            },
            "measurable_objectives": [
                {"id": "obj_dau_50", "objective": "50 DAU within 4 weeks", "metric": "daily_active_users", "target": 50},
                {"id": "obj_gauntlet_rate", "objective": "Gauntlet completion >15% of visitors", "metric": "gauntlet_completion_rate", "target": 0.15},
                {"id": "obj_session_duration", "objective": "Avg session >3 minutes", "metric": "avg_session_duration_seconds", "target": 180},
                {"id": "obj_share_rate", "objective": ">5% Gauntlet completions shared", "metric": "gauntlet_share_rate", "target": 0.05}
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
                            "purpose": "The product IS the landing page. Single provocation card filling screen. Text input beneath. Timer. No signup wall.",
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
                            "purpose": "Daily Gauntlet — 5 provocations, 5 minutes, discover your archetype. Client-side JS.",
                            "section_types": ["gauntlet-interface", "archetype-result-card"],
                            "priority": 3,
                            "content_context": {
                                "archetypes": [
                                    {"name": "The Surgeon", "tagline": "Precise, analytical, cuts to the core", "description": "Strips arguments to essential structure. Fast on logical frameworks.", "strengths": "Analytical speed, structural clarity, finding hidden flaws"},
                                    {"name": "The Wildcard", "tagline": "Unpredictable, highest remix rate", "description": "Takes ideas in unexpected directions. Crosses domains freely.", "strengths": "Creative leaps, remix ability, cross-domain thinking"},
                                    {"name": "The Oracle", "tagline": "Quiet but accurate, rare but impactful", "description": "Does not respond often but when they do it lands. Best prediction accuracy.", "strengths": "Prediction accuracy, timing, selective engagement"},
                                    {"name": "The Catalyst", "tagline": "Responses generate longest chains", "description": "Individual responses might not be best, but they spark others.", "strengths": "Provocation, opening angles, generating momentum"},
                                    {"name": "The Judge", "tagline": "Spectates, but reactions predict outcomes", "description": "Lives in the Gallery. Reaction pattern most predictive of final outcomes.", "strengths": "Pattern recognition, taste, predictive judgement"},
                                    {"name": "The Maker", "tagline": "Stage-dominant, consistent showcaser", "description": "Shows up with creations not opinions. Demonstrates skill.", "strengths": "Creative output, consistency, skill demonstration"},
                                    {"name": "The Scout", "tagline": "Identifies breakout creators early", "description": "The curator. Finds things others miss.", "strengths": "Early identification, taste, discovery"},
                                    {"name": "The Mentor", "tagline": "Breakdowns among most saved content", "description": "Explanations are teaching material. Turns admiration into learning.", "strengths": "Teaching, explanation, knowledge sharing"}
                                ],
                                "gauntlet_mechanics": "5 provocations mixing opinion and creation prompts. Scored on response speed, originality, consistency, topic preferences. Produces primary archetype with secondary tendencies. Shareable visual card."
                            }
                        },
                        {
                            "slug": "about",
                            "purpose": "What Spark is. Quick, energetic explanation. Game metaphor. Arena vs Stage. AI game master concept.",
                            "section_types": ["hero", "platform-comparison", "game-master-explanation", "gauntlet-cta"],
                            "priority": 4,
                            "content_context": {
                                "key_comparisons": {
                                    "vs_reddit": "Reddit: permanent, karma-hierarchical, dominated by early commenters. Spark: ephemeral, anonymised during play, scored on quality not timing.",
                                    "vs_tiktok": "TikTok: passive consumption, massive creator/consumer divide. Spark: opinion entry point, Gauntlet proves anyone can participate in 5 minutes.",
                                    "vs_twitter": "Twitter: shouting into a void. Spark: put in a room where engagement is guaranteed."
                                },
                                "game_master_explanation": "AI does not try to be funny or clever. AI runs the game: finds what is interesting (scraping), frames it sharply (provocation), keeps score (reputation, archetypes), synthesises results (recaps). Humans bring wit and creativity. AI builds the arena."
                            }
                        },
                        {
                            "slug": "archetypes",
                            "purpose": "Archetype system explained. Each gets section with name, tagline, description, strengths. Shareable individual cards.",
                            "section_types": ["hero", "archetype-grid", "archetype-combinations"],
                            "priority": 5,
                            "content_context": {
                                "combination_examples": [
                                    "Surgeon-Oracle: Rarely speaks, but precisely targeted insight that predicts outcomes",
                                    "Wildcard-Catalyst: Unexpected ideas everyone wants to riff on",
                                    "Maker-Mentor: Creates consistently and teaches others to improve",
                                    "Scout-Judge: Does not create or opine, but most trusted taste on the platform"
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
                    "status": "directional"
                },
                "v3_rooms_and_community": {
                    "description": "Live challenge rooms. Arena mode. Chains. Duels. Reputation.",
                    "status": "speculative"
                },
                "v4_and_beyond": {
                    "description": "Stage mode. Multi-format. Niche discovery. Monetisation.",
                    "status": "speculative"
                }
            }
        }
    }
}

sys.stdout.write(json.dumps(msg, ensure_ascii=False, separators=(",", ":")))
PYEOF
)


CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"
DOMAIN="vonc.com"

echo "========================================="
echo "Submitting vonc.com (Spark)"
echo "========================================="
echo "  Correlation: ${CORRELATION_ID}"
echo "  Time:        ${TIMESTAMP}"
echo "========================================="

# Send to domain-submitter
kubectl -n kafka run -i --rm kcat-vonc-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=submit-vonc-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=$CLIENT_ID \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.generic.responses <<JSON
${MESSAGE_BODY}
JSON

echo ""
echo "========================================="
echo "Submitted: ${DOMAIN}"
echo "========================================="
echo ""
echo "SAVE: CORRELATION_ID=${CORRELATION_ID}"
echo ""
echo "Check progress:"
echo "  SELECT id, domain, status FROM sites WHERE domain = '${DOMAIN}';"
echo ""
echo "  SELECT aspect, LEFT(data::text, 80) FROM site_specs ss"
echo "  JOIN sites s ON s.id = ss.site_id"
echo "  WHERE s.domain = '${DOMAIN}' AND ss.is_current = true ORDER BY aspect;"
echo ""
echo "  SELECT wi.item_type, wi.handler_agent, wi.status, wi.priority"
echo "  FROM site_work_items wi JOIN sites s ON s.id = wi.site_id"
echo "  WHERE s.domain = '${DOMAIN}' ORDER BY wi.priority;"
echo ""
echo "  SELECT function, section_type, created_from, LENGTH(html_template) as len"
echo "  FROM content_components WHERE created_from = 'generated' ORDER BY created_at DESC;"









----
${MESSAGE_BODY}
