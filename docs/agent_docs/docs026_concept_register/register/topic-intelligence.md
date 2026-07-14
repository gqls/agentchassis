# Register — topic-intelligence

3 concepts, consolidated from 6 raw extractions (3 unique blocks, each duplicated
once in the source cluster file) across unit U26.

### TPI-001 — Audio-monitoring topic discovery with auto-spawned topic agents
- **status:** abandoned
- **status-evidence:** 017/018 fully design the pipeline (Bloomberg/podcast transcription via Whisper → topic extraction → novelty check → spawn agent) with a phased plan starting "Week 1: financial podcasts"; nothing downstream ever references it.
- **what:** A self-expanding intelligence network: audio streams/podcasts are transcribed, novel topic clusters detected (novel-phrase and frequency-spike detection against a 30-day corpus), and a specialised monitoring agent is automatically spawned per new topic (sources, sentiment, players, trajectory, content generation, subscriber alerts) — "Bloomberg mentions topic at 9:00 AM, your system publishes analysis by 9:30". Included a Domain Intelligence Orchestrator (DIO) deciding which intelligence strategy fits each domain.
- **sources:** docs/architecture/017-audio-monitoring-discussion; docs/architecture/018-audio-monitoring-tech.md#realistic-implementation-path
- **relations:** topic amplifier engine (TPI-002); agent spawning; cross-domain intelligence network (TPI-003)
- **verify-later:** n/a

### TPI-002 — Topic amplifier / deep digger engine
- **status:** abandoned
- **status-evidence:** 019/020 catalogue the hard problems and Python component designs (MinHash LSH dedup, spaCy extraction, verification engine, source discovery, PG+Elasticsearch+Redis storage) with a 6-week plan; no implementation trace exists.
- **what:** The engineering backbone for topic intelligence: data collection (news/social/RSS/scraping), temporal tracking with velocity/anomaly detection, structured extraction (dates, money, entities), claim verification against trusted sources, source discovery (link following, social-graph expansion, citation mining), scalable near-duplicate detection, and a hybrid division of labour — LLMs for context understanding/relevance/noise-filtering (rated "very strong"), traditional code for collection, temporal/quantitative analysis, dedup and storage. Honest bootstrap/noise/evolution problem analysis included.
- **sources:** docs/architecture/020-topic-amplifier-deep-digger.md; docs/architecture/019-information-discovery-agent-spawning#the-honest-assessment; docs/architecture/019-information-discovery-agent-spawning#llms-in-the-loop
- **relations:** audio-monitoring topic discovery (TPI-001); deep-research domain insight agent
- **verify-later:** n/a

### TPI-003 — Cross-domain intelligence network and subscription tiers
- **status:** abandoned
- **status-evidence:** 016's "Hidden Superpowers" section (living knowledge graphs, insight arbitrage between domains, $10/$99/$999/$9,999 subscription tiers, "Organizational OS") is pure vision with no follow-through in later documentation.
- **what:** Developed domains share intelligence: patterns detected on one site alert sibling sites to opportunities ("vehicle-hire.com notices courier demand spike → couriervans.com gets alert"); accumulated contextual memory, relationship mapping and time-series pattern recognition become sellable subscriptions (industry intelligence, trend prediction, competitive clusters) and ultimately an org-wide agent deployment ("every employee gets a personal agent dashboard").
- **sources:** docs/architecture/016-competitive-advantge.md#the-hidden-superpowers-of-your-system; docs/architecture/016-competitive-advantge.md#the-organizational-os-concept
- **relations:** EBORG; audio-monitoring topic discovery (TPI-001); business-strategy subscription models
- **verify-later:** n/a
