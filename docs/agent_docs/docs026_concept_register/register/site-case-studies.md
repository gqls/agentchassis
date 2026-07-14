# Register — site-case-studies

18 concepts, consolidated from 36 raw extractions (the cluster input file contains
the entire site-case-studies block set duplicated exactly twice — 18 unique raw
blocks appearing twice each, with no further cross-unit duplication found) across
units U03, U04, U10, U13, U20, U24c, U24e, U25.

### CASE-001 — idea.uk live-VM / chassis-staging duality
- **status:** deployed
- **status-evidence:** Journal project facts: "idea.uk — LIVE Go service selling £29 reports; single binary under systemd on a Hetzner VM … DNS (Cloudflare) → the VM, so chassis B2 deploys are invisible to the live site. UNCHANGED." Every checkpoint ends "idea.uk live VM untouched."
- **what:** The site-specific arrangement underpinning the whole idea.uk work's risk model: the revenue-earning idea.uk product (£29 reports, live Stripe webhook, orders in a file, reserved paths /request /confirm /approve /decline /stripe/webhook /internal/* /order/*) is a separate Go binary on a Hetzner VM; the chassis-built idea.uk site deploys to Backblaze B2 while DNS still points at the VM, so all chassis work is invisible staging until a separate future cutover decision. Two chassis site_ids exist for idea.uk (97ed2f64-… from the June thread, 1244516d-… from the July thread), treated as separate/earlier rows — confirm before relying on either.
- **sources:** running_notes_scheme_to_components(55).md#Project-facts; HANDOFF_scheme_to_components_for_claude_code(1).md#Environment; HANDOFF_idea_uk_differentiators_section_data.md#Key-facts
- **relations:** idea.uk mission and identity (CASE-002); idea.uk chassis-site build state (CASE-003); chassis deploy model
- **verify-later:** sites rows for idea.uk (both ids); DNS state if a cutover is ever planned

### CASE-002 — idea.uk mission and identity (workshop of tools; never verdicts; warm-paper identity)
- **status:** deployed
- **status-evidence:** "idea.uk mission REFRAMED away from the £29 tool… shipped as a file" and submitted via --mission-file 2026-06-21; classifier "read the mission well" on site 1244516d.
- **what:** idea.uk's genuinely site-specific concept: the place to take an idea seriously — a growing workshop of genuinely good tools (the main event; free + paid), labelled private (in-browser, nothing sent) vs AI/hosted with private leading; cutting-edge succinct research-grounded guides; a news section; and never verdicts — perspective, evidence, and questions framed as opinion (the legal reframe). The £29 verified report was demoted to one specialised flagship tool, with a later build-and-bring-to-market service planned; the warm-paper/ink/single-rust-accent/Fraunces+IBM-Plex editorial identity was preserved throughout. Noted honestly: the privacy and latest-research promises are stated intent the chassis can't yet enforce.
- **sources:** idea.uk/KEY_DOC_idea_uk_mission.txt; idea.uk/HANDOFF(13).md; idea.uk/running_notes(63).md
- **relations:** liability framework (never-verdicts); mission-file mechanism; idea.uk chassis-site build state (CASE-003); idea.uk product (CASE-010)
- **verify-later:** site 1244516d mission spec content

### CASE-003 — idea.uk chassis-site build state (two site rows; staging-only; gated go-live)
- **status:** partial
- **status-evidence:** Current position 2026-06-26: composition + stylesheet correct and verified (tool-portal-light + parchment, commit 05ef817, no LLM drift); pages still dark; rebuild + review + cutover gated on the P0 scheme thread.
- **what:** The concrete build-state history: first chassis run under site 97ed2f64 (2026-06-14, classifier→…→empty index → coordinator fix validated on it); torn down and resubmitted fresh 2026-06-21 as site 1244516d with the mission file; re-resolved onto tool-portal-light 2026-06-25; deployed page defects catalogued (empty differentiators, unresolved CTAs, dead contact form, missing pricing spec, thin nav/footer, empty meta description, dark chrome). The live £29 VM tool remained a separate stream, untouched and earning throughout — the safety property the whole thread leans on.
- **sources:** idea.uk/TODO_chassis_and_idea_uk(1).md; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md; idea.uk/HANDOFF(13).md
- **relations:** every design/pipeline concept this build exercised; VM cutover; idea.uk live-VM duality (CASE-001)
- **verify-later:** current state of site 1244516d; whether the P0 thread completed and cutover happened

### CASE-004 — robot-hands.com rebuild (testbed case study, 2026-07)
- **status:** deployed
- **status-evidence:** "Phase I0 ✅ COMPLETE. 33-page rebuild w/ live news (9 sources); 16 distinct per-page git-path heroes, zero expiring URLs" (2026-07-12).
- **what:** The adopted site's content layer was broken (10 zero-component pages, NULL content, features content in a pre-drift schema) while the imagery pipeline was correct, so it was rebuilt from scratch with news scope: supersede adoption-residue aspects, news-enable classification, add a mission_brief aspect, retire stale items, insert+manually-triage a needs_site_plan trigger, then build a fresh 33-page plan (29 imagery rows) unattended through dispatch. An earlier 2026-05-20 audit said "PATCH, do not re-plan" (foundation sound, build broken); the 2026-07-08 decision superseded it with a full re-plan. Hard requirement: tools must actually work (deployed JS, resolving links) since this is the acceptance surface for all imagery-pipeline phases.
- **sources:** HANDOFF_robot_hands_rebuild.md; SQL_2026-07-08_robothands_rebuild_prep.sql; SQL_2026-07-08_robothands_mission_brief.sql; RUNNING_NOTES_imagery_best_in_class.md#Turns-2–4
- **relations:** news enrichment pattern; schema-contract drift bug (features {title,description} vs {icon,name,description}); Robot Hands website — original 2025 build (CASE-006, same site, earlier era)
- **verify-later:** current plan 7a40a0f9; page/component counts; tool pages' build_status; orphan pre-rebuild pages cleanup (still open)

### CASE-005 — Dartsonline guides defect (benchmark bug, causes A/B/C)
- **status:** deployed (diagnosis confirmed; fix deliberately not applied)
- **status-evidence:** "The benchmark bug (still live, still unfixed — deliberately)" (HANDOFF_CURRENT_fixloop.md).
- **what:** dartsonline.com published a Guides nav link to a blank page. Root mechanism, hand-diagnosed with citations: (A) build-site-planner populated `sections` for only 5 of 15 planned pages; (B) page-build-handler's `check_has_ready_sections` routes sectionless pages to `complete_error`, which is `action: complete_workflow` — a success terminal — so the work item is marked complete though the page was never built; (C) `populate_nav_tables_action.go` filters nav candidates on `pages.status` (lifecycle, defaults 'active') instead of `build_status`, publishing links to unbuilt pages. Kept deliberately live and unfixed as a repeatable benchmark bug for the diagnosis→fix loop; the platform fix is known and can be applied by hand any time.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#★ F0 PILOT; fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1; fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#benchmark bug
- **relations:** standing hypothesis (refuted); mark_no_sections gap; two intake paths disagreement; platform-not-site-data fix philosophy
- **verify-later:** pages/site_work_items rows for site 5fe8785b-223d-41a3-88ee-c07187622381; page-build-handler workflow JSON

### CASE-006 — Robot Hands website — first agent-built multi-page site (2025-10, legacy)
- **status:** deployed
- **status-evidence:** Working group definitions (robot-hands-website v1 with usage rows dated 2025-10-27/30), then robot-hands-complete-website (home/about/contact) with full workflow and trigger scripts.
- **what:** The platform's first end-to-end site build: an agent group spawning a hero writer, image creator, about writer, and contact writer; generating content and a Stability-AI hero image; assembling three HTML pages via aggregate_webpage with embedded CSS/nav. The about page explicitly explains the site was built by AI agents (and "may be for sale"). Served as the proving ground for job topics, data helpers, and image handling — a full generation before the 2026-07 imagery rebuild of the same domain.
- **sources:** docs002_hitl_parallel/README.0100b.updated_state_of_play_for_creating_website.md; docs002_hitl_parallel/README.0100c.workflow_diagram.md; docs001_flow_general/README.096b.robothandswebsite.md
- **relations:** aggregate_webpage; content-creator-about/contact agents; robot-hands.com rebuild (CASE-004, same site, later era)
- **verify-later:** does robot-hands.com exist / what pipeline now owns it

### CASE-007 — relojistas.com go-live + bot verdict
- **status:** deployed
- **status-evidence:** running_notes 2026-06-13 "Relojistas VERDICT from the access log: 14,961 reqs, 83% 404s … Human intent ≈ 0. Clean probe result (domain not worth building), not a measurement failure".
- **what:** First live domain: a Spanish watch FORUM (grounded in a Wayback snapshot), hand-made relojistas-site/ (index + gracias, kind=search, THANKS_PATH=/gracias.html) to unblock go-live. After going live (and later Cloudflare-proxied), the access log showed overwhelmingly bot/crawler traffic (Chrome-spoof crawler, Claude-SearchBot, Semrush, Yandex) with ~0 human intent — a clean negative probe result. It was later decided to static-build it anyway (RSS + crawler presence + 404/referer signal treated as assets).
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-relojistas-go-live-bundle, #2026-06-13-relojistas-verdict, #2026-06-13-b
- **relations:** hand-instance of intent-probe component; passive access-log harvest decision; wayfaringlondoner.com (CASE-008, same probe programme)
- **verify-later:** deploy_setup/relojistas-site/{index,gracias}.html; relojistas_notes/relojistas_golive.md

### CASE-008 — wayfaringlondoner.com page + THANKS_PATH-is-engine-wide
- **status:** partial
- **status-evidence:** running_notes 2026-06-13 "wayfaringlondoner.com page built … a 2015–16 travel blog … BLOG framing"; "Design point — THANKS_PATH is engine-wide".
- **what:** Second hand-made probe page: a 2015–16 travel blog (Csilla; London + Bangkok/Transylvania/Jersey), BLOG framing asking for a destination/London spot/story, tagline gained "and under new ownership". Targets the shared multi-vhost box, and surfaced the constraint that THANKS_PATH is one engine-wide env var, so domains sharing a box must share a thanks filename (standard /thanks.html) — relojistas keeps /gracias.html only because it sits on its own box.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-relojistas-verdict, #2026-06-13-b
- **relations:** shared-box multi-domain onboarding; relojistas.com go-live (CASE-007)
- **verify-later:** wayfaringlondoner-site/; wayfaringlondoner_notes.md (live)

### CASE-009 — Original first-domain set (dropped surgerylight + finance/retail)
- **status:** abandoned
- **status-evidence:** runbook base §3 "Suggested first set: relojistas.com, wayfaringlondoner.com, surgerylight.com, plus one finance tool and one clear retail" — absent from runbook(12) §3, which names only relojistas.
- **what:** The earliest traffic-probe runbook proposed a concrete 3–5 domain starter set, each grounded via Wayback. Later versions dropped the named list down to relojistas + wayfaringlondoner; surgerylight and the finance/retail candidates silently vanished without an explicit decision recorded.
- **sources:** traffic_probe_runbook.md#3; traffic_probe_runbook(2).md#3; traffic_probe_plan(11).md#risks
- **relations:** Wayback grounding method; relojistas.com (CASE-007); wayfaringlondoner.com (CASE-008)
- **verify-later:** n/a

### CASE-010 — idea.uk — AI ideation-as-a-service product
- **status:** deployed
- **status-evidence:** "idea.uk runs as a single Go binary under systemd on a Hetzner box... Billing: Stripe Checkout — a single £29 payment per report, live and earning (proven end-to-end with a real card on 2026-06-14)."
- **what:** A paid tool at idea.uk running an internal ideation method (generate → cut → web-verify → score → rank) against a business domain + audience, producing a ranked report of business-idea candidates with citations. Sold as one-off £29 reports (down from an initial £199 concept) via a request-then-confirm flow with a free "audience-check" taster as the hook, and positioned as the dogfood/first customer of the idea-generation method itself.
- **sources:** RUNBOOK_idea_uk(10).md; RUNBOOK_idea_uk(1).md; running_notes(44).md (2026-05-28 pricing checkpoint)
- **relations:** idea generation method (CASE-011); Go engine supersedes Python (CASE-013); request-then-confirm intake (CASE-015); idea.uk mission (CASE-002)
- **verify-later:** live idea.uk site; idea-go/ module if present in the working tree; Stripe dashboard for the two named accounts

### CASE-011 — Idea generation method — versioned pipeline (v0 → v3)
- **status:** partial
- **status-evidence:** running_notes(44).md traces v0 → v1 (durability factor + named free substitute) → v2 (multi-lens generation + richer capability menu + audience-fit challenge + seller-bundles-support-free check) → v3 (Risk column added as a 6th factor).
- **what:** The core reusable pipeline: generate (multi-lens: asset×capability, demand, generalist-failure, frontier, outcome) → cut (challenge against the specific free substitute + audience-fit challenge + seller-bundles-support-free check) → web-verify → score (Defensibility/Willingness/Buildability/Reuse/Durability, gated on Def≥3 AND Will≥3) → rank. Each version fixed a concrete failure found by running the method against real domains (agritec.uk, gaswholesalers.com, robot-hands.com, websitedesign).
- **sources:** running_notes(44).md lines ~143-411 (method v1/v2 evolution)
- **relations:** Risk-as-hazard scoring dimension (CASE-012); cross-vendor critique (CASE-014); idea.uk product (CASE-010)
- **verify-later:** idea_method_prompt.md; idea_uk_method_v0.md (live); idea-go/prompts.go

### CASE-012 — Risk-as-hazard scoring dimension
- **status:** deployed
- **status-evidence:** "The rubric had no dimension for the consequence of being wrong. It was caught on operator instinct, which doesn't scale," followed by implementation in idea-go/engine.go/prompts.go and Python parity, "Built + vetted + tested clean."
- **what:** A 6th scoring factor (1-5, 5=safest) scoring the consequence of being wrong, deliberately kept separate from the fitness sum (Def+Will+Build+Reuse+Dur) so it can't be gamed by high fitness. Risk=1 auto-drops a candidate into a "Dropped for operator risk" section; Risk≤2 still advances but flagged "needs liability work before building"; Risk is a rank tiebreaker at equal fitness. Triggered by a near-miss: an SFI single-farm assessment scored a confident test-now recommendation that could have cost a farmer £5k-50k if wrong.
- **sources:** running_notes(44).md (Risk rubric table + rules); LIABILITY_AND_TERMS.md (referenced, live)
- **relations:** idea generation method (CASE-011); LIABILITY_AND_TERMS / legal pages
- **verify-later:** idea-go/engine.go scored struct; idea_method_runner.py parity implementation

### CASE-013 — Go engine supersedes Python reference implementation
- **status:** superseded
- **status-evidence:** running_notes(44).md confirms directly: "Ported the idea.uk tooling from Python to Go (platform is Go throughout)... The Python files remain as the reference implementation but Go is now the canonical version, consistent with the rest of the platform."
- **what:** idea.uk's engine and service were first built in Python (FastAPI, idea_method_runner.py, idea_service.py, sqlite via test_idea_flow.py), validated end-to-end (20/20 checks), then rewritten in idiomatic stdlib-only Go (no external deps, go vet/build/test all clean, 19/19 checks) to match the rest of the Go-throughout platform. The rewrite preserved the id-based (not title-based) threading bug-fix and the cross-vendor cut design — a genuine, confirmed language-migration relationship with byte-level before/after evidence, one of very few in this corpus.
- **sources:** RUNBOOK_idea_uk(1).md §pieces table vs live RUNBOOK_idea_uk.md; running_notes(44).md
- **relations:** idea.uk product (CASE-010); idea.uk deployment topology (CASE-001)
- **verify-later:** confirm whether idea-go/ or the Python files are what's actually running in production today

### CASE-014 — Cross-vendor critique (multi-model critique step)
- **status:** deployed
- **status-evidence:** running_notes(44).md: "Cross-vendor critique implemented (was the one untested multi-model claim). The runner's cut step now routes through OpenAI if OPENAI_API_KEY is set... else falls back to a different Anthropic model." Later ported into the Go engine unchanged.
- **what:** The idea-generation method's "cut" (critique) step deliberately runs on a different model vendor than the "generate" step where possible (OpenAI if OPENAI_API_KEY is set, else a different Anthropic model as a same-vendor fallback), specifically to avoid the same model marking its own homework. A stderr log line was added after user confusion about which vendor actually ran, so every run states its own critique provenance.
- **sources:** running_notes(44).md ("Built out the gaps + ran the flow end-to-end", "Added a [cut] vendor log line to engine.go")
- **relations:** idea generation method (CASE-011)
- **verify-later:** idea-go/engine.go call_other_vendor / cross-vendor branch

### CASE-015 — idea.uk request-then-confirm intake with capacity throttle
- **status:** deployed
- **status-evidence:** RUNBOOK_idea_uk(1).md "Flow (request-then-confirm)"; running_notes(44).md: "Capacity throttle (protects the 72h promise): MAX_ACTIVE_ORDERS caps orders in flight."
- **what:** The customer-facing order flow deliberately never takes payment until an operator has screened the request: submit (free) → operator /confirm (creates Stripe Checkout, or post-REVIEW_BEFORE_PAY runs the engine first) → /decline available at any point with a polite no-charge email → webhook-driven fulfilment → operator-reviewed delivery (AUTO_DELIVER off by default). A MAX_ACTIVE_ORDERS throttle returns HTTP 409 "at capacity" once too many orders are in flight, protecting a stated 72-hour turnaround promise; /capacity is a public endpoint so the page can show "currently full."
- **sources:** RUNBOOK_idea_uk(1).md; running_notes(44).md
- **relations:** REVIEW_BEFORE_PAY billing flow; idea.uk product (CASE-010)
- **verify-later:** idea-go/service.go capacity/throttle logic

### CASE-016 — Leopardess rebuild programme (phases L0–L9)
- **status:** partial
- **status-evidence:** PLAN phase table (2026-07-12, turn 13): "L0–L4 done, L5 partial, L6–L8 not started, L9 ongoing"; HANDOFF §3 "main pages live and verified".
- **what:** Rebuild of leopardessconsulting.co.uk (a site the platform built for itself) to be honest, well-branded, and useful: evidence audit (L0), spec truth pass (L1), brand/logo (L2), palette fork (L3), 3-per-row layout (L4), copy rewrite (L5), explanatory imagery (L6), charts (L7), tools/guides/news build-out (L8), and coherent deploy (L9). Includes an audience pivot (A2): sceptical, commercially-sharp, non-specialist buyer, with technical depth one click down.
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#Phases; docs/leopardessconsulting/HANDOFF.md#0; docs/leopardessconsulting/RUNNING_NOTES.md#Decision-log
- **relations:** claim-evidence audit rule (CASE-017); reuse-not-rebuild build-out (CASE-018); anti-hype voice spec; per-site style fork
- **verify-later:** site_id 4851f6fc-71cf-4160-a270-e03d6d3e0732 pages/build_status; live leopardessconsulting.co.uk pages

### CASE-017 — Claim-evidence audit rule ("no claim ships without an audit row")
- **status:** deployed
- **status-evidence:** AUDIT header 2026-07-09: "no claim ships unless it has a row in this table"; fabrication sweep 2026-07-10 "Result: CLEAN".
- **what:** A site-truth methodology: every marketing claim is verified against code, live Postgres, or an HTTP response before it may appear on the site; unverifiable claims are removed or hedged explicitly (the one allowed unproven claim — "third busiest sports site" — is published labelled as recollection). Produced a verified capability inventory (Companies House cascade with 2,767 verified businesses; news pipeline 5,652 items/4,672 scored; tool-generation agent family; DB-defined hierarchical agents 143 defs/56 active/40 spawners; Banana+SDXL imagery; 8 own sites) and an explicit UNSUPPORTED list.
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#1, #2; docs/leopardessconsulting/PLAN_leopardess_rebuild.md#Standing-rules
- **relations:** LLM fabrication classes; verify-by-artifact operator discipline; Leopardess rebuild programme (CASE-016)
- **verify-later:** business_intel.businesses counts; feed item counts; agent_definitions counts

### CASE-018 — Reuse-not-rebuild site build-out with honest "simulation" labelling
- **status:** aspirational
- **status-evidence:** PLAN L8 "not started — reuse existing tool library; surface the live news feed"; HANDOFF §6.5.
- **what:** The leopardess L8 plan: deploy/adopt existing interactive tool components (ROI estimator, quizzes, calculators — deterministic client-side widgets that must be labelled as simulations, not live inference), surface the real news-feed pipeline, pair guides with tools (tool-deployer already creates companion guides + cross-links), and note that a "game" has no formal platform existence — it is simply a component_level='tool' component.
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#L8; docs/leopardessconsulting/HANDOFF.md#6
- **relations:** tool-library; news-feed-pipeline; dynamic-applications tool builder tiers; Leopardess rebuild programme (CASE-016)
- **verify-later:** tool component inventory; news feed surfacing on any site
