# Register — conversion-playbooks

4 concepts, consolidated from 8 raw extractions across unit U20. (The cluster input file contained this category's raw blocks twice, back-to-back and byte-identical; each pair is merged into one entry below. No cross-unit duplication found — all raw blocks for this category came from U20_legacy_docs_a.md.)

### CVP-001 — Playbook > Strategic Pattern > Component hierarchy (Librarian as system brain)
- **status:** abandoned
- **status-evidence:** Extensive design (Playbooks/Strategic_Patterns/Pattern_Component_Slots/Components schema, success_score feedback) across website_analysis 001-003; no implementation era follows — the MVP path (chief-strategist + in-house components) shipped instead, and the schema never reappears.
- **what:** "Strategy-to-website engine": the library stores business solutions, not just components — Playbooks (objective+vertical strategies with success scores, e.g. affiliate product-review), containing Strategic Patterns (comparison-table, best-of listicle), containing Components. Learn loop classifies scraped winners into this hierarchy; Execute loop queries "best playbook for objective X in vertical Y" and assembles it; A/B results feed success_score back. The Librarian is the sole read/write gatekeeper; "the link is the database schema".
- **sources:** docs004_website_capture_project/website_analysis/README.001.prompt_of_objectives.md; docs004_website_capture_project/website_analysis/README.003.summary_for_development.md; docs004_website_capture_project/website_analysis/README.002.summary_of_plan_agents_groups.md
- **relations:** CVP-002 behavioural models library (the surviving cousin); site-spec-and-classifier archetype system is the spiritual live successor; affiliate content-type placement knowledge (reviews/comparisons/listicles) embedded here
- **verify-later:** confirm no Playbooks/Strategic_Patterns tables exist

### CVP-002 — Behavioural models library and functional component labelling
- **status:** partial
- **status-evidence:** "PAS" shipped as a real input (`"model": "PAS"` in mvp-site-builder trigger messages; chief-strategist prompt takes {{.model}}); the wider library (AIDA, Fogg B=MAP, Cialdini, Hook) and deep inference labelling remained design.
- **what:** Components are labelled by behavioural function, not visual pattern: not "hero" but "attention_capture"/"problem_statement"/"social_proof", drawn from marketing science (AIDA, PAS, Fogg Behaviour Model, Cialdini's persuasion principles, the Hook model). Build plans map a chosen behavioural model to a sequence of functional sections; the architect assembles "a psychological argument, not just a visual page". Self-critiques recorded: inference black-box risk (LLM can't reliably tell "agitation" from "interest"), theory-vs-reality gap, new-generic monoculture trap.
- **sources:** docs004_website_capture_project/website_analysis/README.006.visual_to_code.md; docs004_website_capture_project/website_analysis/README.007.behavioural_models.md; docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md
- **relations:** CVP-003 Minimal Viable Funnel; data-function contract; content strategy in current content pipeline (content-quality docs) is the descendant
- **verify-later:** whether current build plans still carry a behavioural model field

### CVP-003 — Minimal Viable Funnel (pragmatic-first Day-1 build)
- **status:** superseded
- **status-evidence:** Fully built as mvp-site-builder (boxing-tickets.com runs); superseded within docs004 itself by the briefing→specialist-architect pipeline and later by the current work-item site build.
- **what:** Anti-boil-the-ocean strategy: start with one behavioural model (PAS) and three generic in-house components (problem/agitate/solution blocks) so a strategically coherent landing page can be built with zero scraped data — solving the cold-start problem. Scraping demoted to an "iteration engine" suggesting upgrades.
- **sources:** docs004_website_capture_project/website_analysis/README.006.visual_to_code.md#minimal-viable-funnel; docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** MVP site builder pipeline; CVP-004 strategic fallback stubs; in-house forge
- **verify-later:** —

### CVP-004 — Strategic fallback stubs for non-replicable components
- **status:** abandoned
- **status-evidence:** Design-only ("Store 'Stubs' with 'Fallbacks'… two-pronged output") in website_analysis 001/003; no stub tables or developer-task topics appear later.
- **what:** When ingestion finds a component it can't replicate (e.g. a mortgage calculator), record a Stub with its strategic goal (lead-gen-quote) and a linked simple fallback component (CTA form). The live site ships the working fallback; simultaneously a developer task goes to a HITL queue ("developer.tasks.required") to build the real thing as v2. The site is always complete and strategically sound.
- **sources:** docs004_website_capture_project/website_analysis/README.001.prompt_of_objectives.md#strategic-fallback; docs004_website_capture_project/website_analysis/README.003.summary_for_development.md
- **relations:** dynamic-applications (the current interactive app generation finally addresses "non-replicable dynamic apps"); HITL queue
- **verify-later:** none — idea registry
