# Register — portfolio-evolution

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

3 concepts, consolidated from 6 raw extractions across units U20, U21. (The cluster input file contained this category's raw blocks twice, back-to-back and byte-identical; each pair is merged into one entry below. The three entries are closely related sub-parts of one larger 2026-era strategy — the founding 4-phase engine (PEV-003) and its two named phases/mechanisms (PEV-001, PEV-002) — kept as separate entries because each is independently named and cited, but cross-referenced.)

### PEV-001 — Pragmatic Evolution model (explore/exploit portfolio cohorts)
- **status:** abandoned
- **status-evidence:** Full strategy synthesis in 008/014 (cohorts: top-10% untouched, middle-40% careful P1-P2 A/B tests, bottom-50% high-velocity churn; site-specific optimisation "no monoculture"); no subsequent doc era operates a portfolio this way — the platform pivoted to per-site quality loops.
- **what:** An evolutionary algorithm over a large site portfolio: select worst performers, radically mutate them with mixed component "genes", evaluate fitness after 3 months. Critique recorded and resolved into an explore/exploit design: attribution black hole and SEO destabilisation confine chaos to a "loser" cohort where attribution is deliberately ignored; winners graduate. Winning changes are applied only to individual sites where they actually won, and content evolves on a separate continuous track from layout to protect SEO.
- **sources:** docs004_website_capture_project/website_analysis/README.008.evolutionary_algorithm_of_site_portfolio.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** improvement-loop (the live per-site discovery→audit→fix cycle is the surviving descendant in spirit); traffic-analytics (fitness signal dependency); PEV-003 Pragmatic Evolution Engine (the 4-phase parent plan this is Phase 3/4 of)
- **verify-later:** none — strategy registry; check if any cohort/experiment tables exist

### PEV-002 — Hypothesis priority list (learn loop as idea generator, not fact finder)
- **status:** abandoned
- **status-evidence:** 008/014: "All scraped data is treated as messy, high-correlation ideas, not truth… Librarian generates a Hypothesis Priority List (P1-P5)"; scorecard interrogation of sites against all behavioural models; no implementation follows.
- **what:** Epistemics for the scraping programme: accept that ingestion finds correlation ("cargo cults"), rank target sites by external success metrics (Ahrefs/Semrush APIs via an seo_api_adapter), interrogate each against every behavioural model to produce confidence scorecards, and emit a prioritised backlog of testable hypotheses for the Evolve loop to convert into causation.
- **sources:** docs004_website_capture_project/website_analysis/README.008.evolutionary_algorithm_of_site_portfolio.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md; docs004_website_capture_project/website_analysis/README.003.summary_for_development.md
- **relations:** Prospector/seo_api_adapter (never built); llm-quality-testing shares the evaluation mindset; PEV-003 Pragmatic Evolution Engine (this is Phase 2, "Idea Generator")
- **verify-later:** none

### PEV-003 — Pragmatic Evolution Engine (portfolio build/learn/test/optimize)
- **status:** abandoned
- **status-evidence:** docs009/001 full 4-phase plan ("Internal Library of Effectiveness", "Controlled Evolutionary Cohorts"); no subsequent doc implements cohort testing, manifests, or the Librarian.
- **what:** The founding mission statement for a large-scale website portfolio: Phase 1 pragmatic-first MVP builds from behavioural models (AIDA/PAS) with intelligent component fallback; Phase 2 "Idea Generator" evidence gathering from winner sites (Prospector via Ahrefs-type metrics, Capture Bot producing dom+screenshot+layout_map "Rosetta Stone", Pattern Deconstructor VLM scoring components against behavioural models, Librarian producing a Hypothesis Priority List — see PEV-002); Phase 3 large-scale single-variable A/B cohort tests turning correlation into causation, with content and layout evolved on separate tracks for SEO stability — see PEV-001; Phase 4 site-specific optimization (winners applied only where they won — no monoculture), manifest.json component "genes" per site, git_hook_adapter flagging human-edited repos as desynchronized for HITL review, and exporter agents (WordPress XML/SQL) for client handoff.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#Core-Mission, #Phase-2, #Phase-4
- **relations:** site interrogation/pattern library; adoption-pipeline; improvement-loop is the maintenance-shaped descendant; llm-quality-testing; PEV-001 Pragmatic Evolution model; PEV-002 Hypothesis priority list; BIZ-019 WordPress export/handoff idea (the exporter-agent detail)
- **verify-later:** any manifest.json in site repos; git_hook_adapter; cohort/experiment tables (expected absent)
