# Register — social-media

13 concepts, consolidated from 26 raw extractions (13 unique blocks, each
appearing twice due to exact whole-block duplication in the cluster input file)
across unit U25. No conceptual merges were needed within this category — each
block covers a genuinely distinct design facet of the Spark/Forge product
lineage, all drawn from the same `docs/social001_vonc_tiktok_social/` design
corpus.

### SOC-001 — The Forge — AI-seeded community knowledge platform
- **status:** abandoned
- **status-evidence:** 001 doc "Status: Concept stage. Parked for future development." (file dated Mar 2026; no later reference builds on it).
- **what:** Product concept predating Spark: AI answers are published as explicit first drafts into a categorised community feed; humans validate/challenge/add experience/fork; the AI synthesises human input into improved versions, visibly showing evolution. Key dynamics: correcting the AI is rewarding not adversarial; "Beat the AI" competing answers earn domain reputation; AI as fair debate synthesiser; vertical Forge embeds (e.g. vets refining pet-owner answers). Open questions recorded: moderation at scale, cold start, AI identity, expert monetisation.
- **sources:** docs/social001_vonc_tiktok_social/001_concept_the_forge_humans_edit_ai_responses.md; docs/social001_vonc_tiktok_social/002pre_whole_chat (origin transcript)
- **relations:** Spark (SOC-002, successor concept, keeps AI-as-participant-not-oracle DNA); hitl
- **verify-later:** none (never built)

### SOC-002 — Spark — AI game-master social platform (core concept)
- **status:** partial
- **status-evidence:** 002e "Status: Active exploration. All mechanics candidates for live testing"; the v1 content-first site is live on vonc.com (minilobby docs, 2026-07).
- **what:** "AI-driven provocation engine where the world is the content and your take is the game." AI occupies the game-master/producer role — framing, structure, scoring, recaps, synthesis, hosting, matchmaking — and is deliberately absent from responses, humour, personality, takes ("producer, not performer"; when something is funny, a human did it). Differentiators: opinion-first entry ("what's your take?" beats create-from-scratch), ephemeral game-structured challenges, rooms not feeds, showcase+competitive dual mode. Includes the AI-slop strategy (restrict AI to strengths; deliberately-bad AI takes as straight-man seeds) and moderation-by-design (AI referee/yellow cards, positive selection, provocation design as tone control, ephemerality limits damage, "Cursed" as community signal). Second-screen positioning: "TikTok: watch. Twitter: shout. Spark: play."
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Positioning,#AI-Role,#The-AI-Slop-Problem,#Moderation; docs/social001_vonc_tiktok_social/trigger_script/004_submit_vonc_trigger.sh (mission brief)
- **relations:** The Forge (SOC-001, predecessor); vonc.com v1 site (vonc.md VONC-009); provocation engine (SOC-007); archetype system (SOC-005)
- **verify-later:** site_specs aspects mission/roadmap for vonc site 9ec3b9ee

### SOC-003 — Arena + Stage dual modes and their mechanic families
- **status:** aspirational
- **status-evidence:** 003d roadmap: v3 "Live challenge rooms. Arena mode. Chains. Duels." status "speculative"; v4 "Stage mode … speculative".
- **what:** Two complementary energies that feed each other: Arena (competitive — provocations, reaction vocabulary Genius/Delusional/Suspicious/Based/Cursed, remix chains with shared credit, 60-second duels with AI referee, Misfit Mashup) and Stage (showcase — springboard challenges, Fire/Inspired/Stealing-This/Teach-Me/Vibe reactions, quality-curated Tune In follow, niche discovery rooms, Glow-Up progress reels, Teach-Me-triggered mentorship, AI-matchmade collabs, Rising newcomer spotlight, Taste Graph curator prestige). Flywheel: Stage showcases become Arena provocations; Arena winners get Stage suggestions; over time the platform generates its own culture.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Two-Modes,#Arena-Mechanics,#Stage-Mechanics; 002b (first appearance of the Arena/Stage split)
- **relations:** Spark core (SOC-002); rooms-not-feeds (SOC-004); emergent profiles (SOC-005)
- **verify-later:** none built (v3/v4)

### SOC-004 — Rooms-not-feeds architecture and the engagement-depth spectrum
- **status:** aspirational
- **status-evidence:** 002e: room mechanics and The Drift are design prose; v1 ships only static pages (003d "v1_content_first").
- **what:** Structural anti-feed design: a Lobby of 3–5 live rooms with energy indicators; room zones (Floor active / Gallery spectating with zero barrier); crowd-energy feedback (cards heat, reactions ripple, tug-of-war splits); ephemeral challenges with lasting recaps; Director's Cut AI replays; serialised prediction challenges; multi-format rotation; "The Crowd Speaks" synthesis; Moments (remarkable outcomes elevated to permanent shareable record). The Drift is the passive snackable stream; each depth level (Drift → Lobby → Gallery → Floor → solo modes) is complete in itself. Sound/haptics design included.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Shared-Mechanics,#The-Drift,#Sound-and-Haptics
- **relations:** Arena/Stage (SOC-003); lobby-grid component (vonc.md VONC-005, v1's static echo of the lobby)
- **verify-later:** none built

### SOC-005 — Behavioural archetype system + Daily Gauntlet
- **status:** partial
- **status-evidence:** 003d gauntlet page spec with 8 archetypes; RUNBOOK_minilobby §0: gauntlet tool + archetype-taster-quiz deployed, archetype hub (8 entity pages) live 2026-07-12.
- **what:** Identity engine: emergent profiles built from behaviour (can't be faked), producing archetypes (Surgeon, Wildcard, Oracle, Catalyst, Judge, Maker, Scout, Mentor) with secondary tendencies, earned via the Daily Gauntlet (5 provocations, 5 minutes, scored on speed/originality/consistency/topic preference) and shareable as visual cards — the "viral Trojan horse" (BuzzFeed/MBTI/Hogwarts dynamics on demonstrated behaviour, works with zero community). Radical anonymity in Arena (no usernames during play, reveal after, reputation never boosts visibility) and rotating achievable status games instead of permanent karma. On vonc v1 this exists as client-side tools + a canon archetype content set (088/089 note live archetype-combinations copy had drifted off-canon).
- **sources:** docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#gauntlet content_context; docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#The-archetype-as-viral-Trojan-horse,#Identity-and-Profiles,#Radical-Anonymity; docs/social001_vonc_tiktok_social/minilobby_task/088_archetype_entity_pages.sql (header)
- **relations:** archetype hub build (vonc.md VONC-010); content-first launch (SOC-009); Daily Gauntlet page
- **verify-later:** /tools/gauntlet/index.html and /tools/archetype-taster-quiz/ on vonc.com

### SOC-006 — Cold-start design: AI sparring partner and solo-first completeness
- **status:** aspirational
- **status-evidence:** 003d: "v2_sparring_and_interaction … status: directional, depends_on backend API infrastructure".
- **what:** The empty-room solution: first 10 seconds are a provocation card + text input + timer with no signup ("Engage first. Understand second. Commit third."); the AI is a transparent sparring partner (counter-arguments, not chatbot small-talk) so the experience is complete for one person; the scraping+AI flywheel self-fills content; invites are challenges not invitations; explicit scale thresholds (1 / 5–20 / 50–200 / 500+) each designed to feel complete. Sparring is also the top per-user cost risk with named mitigations (rate limits, smallest viable model, cached counters).
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Cold-Start-Design,#The-First-Visit; 002 (original fuller treatment: Spark Solo, landing page day 1)
- **relations:** AI cost architecture (SOC-008); content-first launch (SOC-009)
- **verify-later:** none built (v2)

### SOC-007 — Provocation engine — layered content production architecture
- **status:** aspirational
- **status-evidence:** 002e technical-layers section is design; NOTES_provocations-archive-list (2026-07-09): "the archive feed is hand-committed until the Phase-3 pipeline emits provocations.json".
- **what:** Six-layer production line: (1) Raw Feed — scrapers pull social-interest trends (not newspaper news); (2) Framing Engine — cheap local models (Mistral/Llama 8B CPU) generate 5–10 provocation candidates per item, ~2,000/day; (3) Curation Gate — stronger model or human picks the best 15–20, learns from engagement; (4) Mashup Engine — foundation models find non-obvious connections, 2–5 calls/day; (5) Serialisation Tracker — narrative threads, mostly database; (6) Niche Detector — embeddings/ML clustering for Stage rooms. Content tone contract: interesting not dark, slightly adult never gruesome, competitive without money betting.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Provocation-Engine,#Content-Tone; docs/social001_vonc_tiktok_social/002_concept_spark.md#The-Provocation-Engine (original layer detail)
- **relations:** Phase-3 provocation pipeline (vonc.md VONC-002, its v1 delivery vehicle); AI cost architecture (SOC-008); news-feed-pipeline (sibling scraping infrastructure)
- **verify-later:** any provocation-orchestrator agent definition; scheduled_tasks for provocation refresh

### SOC-008 — AI cost architecture: fixed background vs per-user scaling
- **status:** aspirational
- **status-evidence:** 002d/002e cost tables are design projections ("Scaling Scenarios … 100,000 DAU"); nothing metered in production.
- **what:** Cost-shaping principle: most AI cost is fixed (content production scales with ~15–20 challenges/day, not users) and runs on cheap local models (~£5/day background at scale); the only linearly-scaling cost is per-user interactive AI (sparring, Gauntlet scoring) and is throttled by design. Projected £0.003–0.008/user/day, so a £3–5/month subscription covers compute several times over; break-even on subscriptions before ads/brand revenue.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#AI-Cost-Architecture; 002d (first appearance of the full cost tables)
- **relations:** provocation engine (SOC-007); cold-start sparring (SOC-006); revenue model (SOC-012)
- **verify-later:** none (projection)

### SOC-009 — Content-first launch strategy for Spark (vonc.com as destination)
- **status:** partial
- **status-evidence:** 003d "current_phase: v1_content_first … Static S3 site. Provocations as SEO content. Daily Gauntlet as viral archetype quiz"; the site is live per minilobby docs.
- **what:** Don't launch a social platform; launch a content destination that happens to have interactive features. Daily provocations with the AI's take are SEO pages with shareable URLs; every provocation/response generates a self-contained share card ("47 people responded. Think you'd do better?" — the TikTok growth pattern for text); the archetype quiz works with zero community; vertical clustering happens organically by arrival path (the Reddit model — nobody joins "Reddit"). First-week content calendar defined (daily provocations, Gauntlet, weird-stat micro-content; weekly mashup + prediction post). "The first visit experience IS the pitch."
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Launch-Strategy; docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#The-initial-request
- **relations:** Spark core (SOC-002); vonc.com v1 site (vonc.md VONC-009); provocation engine (SOC-007)
- **verify-later:** vonc.com live pages; provocations archive content

### SOC-010 — Motivation hierarchy and designed user journey
- **status:** aspirational
- **status-evidence:** 002c introduced the motivation tiers; 002d the journey; both remain design prose in 002e.
- **what:** Retention design: four motivation tiers (identity/status; belonging/connection; growth/learning; financial reward — "financial reward follows demonstrated value, nobody buys prominence") mapped to user types (casual → professional creator; acquire at Tier 2). A staged journey — first 5 seconds to month 6+ — engineered as intrigue → creation → validation → habit → identity → community → mastery → purpose, with engineered social moments (first reaction, "12 people have a similar profile", first remix) and the principle that the platform reveals itself through use, never onboarding.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#What's-In-It-For-Users,#The-User-Journey
- **relations:** archetype system (SOC-005); games/puzzle retention (SOC-011); revenue model (SOC-012)
- **verify-later:** none built

### SOC-011 — Games and daily-puzzle retention ecosystem
- **status:** aspirational
- **status-evidence:** 002e: "Explore this in detail — could be the primary retention driver" (marked to-explore, not planned into any roadmap phase).
- **what:** A flagged expansion space beyond the Gauntlet: Wordle-style daily challenges with shareable result grids; daily games generated from scraping output (higher/lower with real stats, trend trivia); competitive/timed/bracket formats; streak mechanics (participation, prediction accuracy, creativity); seasonal tournaments; micro-games as retention bridges — a whole daily-play ecosystem tied to real-world content, potentially rivalling NYT Games.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Games-Daily-Puzzles-and-Gamification
- **relations:** Daily Gauntlet (SOC-005); predictions/time capsules
- **verify-later:** none built

### SOC-012 — Spark revenue model
- **status:** aspirational
- **status-evidence:** 002e "Revenue Model (future)" — all items future-tense.
- **what:** Low subscription (£3–5/month) covering AI costs; brand-sponsored challenges (creators selected on niche reputation, not follower count — meritocratic sponsorship); revenue share on high-engagement showcases with challenge-driven supply control against content mills; creator subscription channels; collab marketplace; vertical expert consultations. Prediction staking uses reputation tokens, never money.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Revenue-Model,#Tier-4
- **relations:** AI cost architecture (SOC-008); motivation hierarchy (SOC-010)
- **verify-later:** none built

### SOC-013 — Vertical integration of Spark mechanics into domain sites
- **status:** aspirational
- **status-evidence:** 003d: "v4_and_beyond … Vertical integration … speculative"; 002e "Verticals after mechanics proven".
- **what:** The same mechanics re-flavoured per vertical: vet/pet (wholesome), finance (prediction-heavy), fashion (image-dominant), food (constraint challenges), with vonc.com as the unconstrained proving ground. Echoes The Forge's vertical embed idea; also the component-library payoff — a second site with Spark features reuses the generated interactive-platform components.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Vertical-Integration; docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#Second-build-reuses-everything
- **relations:** component selector/creator; The Forge (SOC-001)
- **verify-later:** none built
