
<!-- SOURCE: U03_idea_uk_section_data.md -->
### idea.uk live-VM / chassis-staging duality
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** Journal project facts: "idea.uk — LIVE Go service selling £29 reports; single binary under systemd on a Hetzner VM … DNS (Cloudflare) → the VM, so chassis B2 deploys are invisible to the live site. UNCHANGED." Every checkpoint ends "idea.uk live VM untouched."
- **what:** The genuinely site-specific arrangement underpinning the whole unit's risk model: the revenue-earning idea.uk product (£29 reports, live Stripe webhook, orders in a file, reserved paths /request /confirm /approve /decline /stripe/webhook /internal/* /order/*) is a separate Go binary on a Hetzner VM; the chassis-built idea.uk site deploys to Backblaze B2 while DNS still points at the VM — so all chassis work is invisible staging and the VM cutover is a separate future decision. Two chassis site_ids exist for idea.uk (97ed2f64-… in the June thread, 1244516d-… in the July thread) — treated as separate/earlier rows, confirm before relying on either.
- **sources:** running_notes_scheme_to_components(55).md#Project-facts; HANDOFF_scheme_to_components_for_claude_code(1).md#Environment; HANDOFF_idea_uk_differentiators_section_data.md#Key-facts
- **relations:** platform mission; chassis deploy model.
- **verify-later:** sites rows for idea.uk (both ids); DNS state if a cutover is ever planned.

<!-- SOURCE: U04_idea_uk.md -->
### idea.uk mission and identity (workshop of tools; never verdicts; warm-paper identity)
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** "idea.uk mission REFRAMED away from the £29 tool… shipped as a file" and submitted via --mission-file 2026-06-21; classifier "read the mission well" on site 1244516d.
- **what:** The site's genuinely site-specific concept: idea.uk = the place to take an idea seriously — a growing workshop of genuinely good tools (the main event; free + paid), tools labelled **private (in-browser, nothing sent)** vs **AI/hosted** with private leading; cutting-edge succinct research-grounded guides; a news section; **never verdicts** — perspective, evidence, and questions framed as opinion (the legal reframe); the £29 verified report demoted to one specialised flagship tool; later a build-and-bring-to-market service; preserve the warm-paper/ink/single-rust-accent/Fraunces+IBM-Plex editorial identity. Noted honestly: the privacy and latest-research promises are stated intent the chassis can't yet enforce.
- **sources:** idea.uk/KEY_DOC_idea_uk_mission.txt; idea.uk/HANDOFF(13).md (idea.uk current state); idea.uk/running_notes(63).md (nnn/ooo 06-21)
- **relations:** liability framework (never-verdicts); mission-file mechanism; standing-ambition.
- **verify-later:** site 1244516d mission spec content.

<!-- SOURCE: U04_idea_uk.md -->
### idea.uk chassis-site build state (two site rows; staging-only; gated go-live)
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** Current position 2026-06-26: composition + stylesheet correct and verified (tool-portal-light + parchment, commit 05ef817, no LLM drift); pages still dark; rebuild + review + cutover gated on the P0 scheme thread.
- **what:** The concrete case-study state: first chassis run under site 97ed2f64 (2026-06-14: classifier→…→empty index → coordinator fix validated on it); old chassis site torn down and resubmitted fresh 2026-06-21 as site 1244516d with the mission file; re-resolved onto tool-portal-light 2026-06-25; deployed page defects catalogued (empty differentiators, unresolved CTAs, dead contact form, missing pricing spec, thin nav/footer, empty meta description, dark chrome). The live £29 VM tool is a separate stream, untouched and earning throughout — the safety property the whole thread leans on.
- **sources:** idea.uk/TODO_chassis_and_idea_uk(1).md; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (open items + current position); idea.uk/HANDOFF(13).md
- **relations:** every design/pipeline concept above (this is their test case); VM cutover.
- **verify-later:** current state of site 1244516d; whether the P0 thread completed and cutover happened.

<!-- SOURCE: U10_imagery.md -->
### robot-hands.com rebuild (testbed case study)
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** "Phase I0 ✅ COMPLETE. 33-page rebuild w/ live news (9 sources); 16 distinct per-page git-path heroes, zero expiring URLs" (2026-07-12).
- **what:** The adopted site's content layer was broken (10 zero-component pages, NULL content, features content in a pre-drift schema) while the imagery pipeline was correct — so it was rebuilt from scratch with news scope: supersede adoption-residue aspects, news-enable classification, add a mission_brief aspect, retire stale items, insert+manually-triage a needs_site_plan trigger, then a fresh 33-page plan (29 imagery rows) built unattended through dispatch. A 2026-05-20 audit first said "PATCH, do not re-plan" (foundation sound, build broken); the 2026-07-08 decision superseded it with full re-plan. Hard requirements: tools must actually work (deployed JS, resolving links) and it is the acceptance surface for all I-phases.
- **sources:** HANDOFF_robot_hands_rebuild.md, SQL_2026-07-08_robothands_rebuild_prep.sql, SQL_2026-07-08_robothands_mission_brief.sql, RUNNING_NOTES_imagery_best_in_class.md#Turns-2–4
- **relations:** news enrichment pattern; schema-contract drift bug (features {title,description} vs {icon,name,description}); orphan pre-rebuild pages cleanup still open.
- **verify-later:** current plan 7a40a0f9; page/component counts; tool pages' build_status.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Dartsonline guides defect (benchmark bug, causes A/B/C)
- **category:** site-case-studies
- **status-signal:** deployed (diagnosis confirmed; fix deliberately not applied)
- **status-evidence:** "The benchmark bug (still live, still unfixed — deliberately)" (HANDOFF_CURRENT_fixloop.md)
- **what:** dartsonline.com published a Guides nav link to a blank page. Root mechanism, hand-diagnosed with citations: (A) `build-site-planner` populated `sections` for only 5 of 15 planned pages; (B) `page-build-handler`'s `check_has_ready_sections` routes sectionless pages to `complete_error`, which is `action: complete_workflow` — a success terminal — so the work item is marked complete though the page was never built; (C) `populate_nav_tables_action.go` filters nav candidates on `pages.status` (lifecycle, defaults 'active') instead of `build_status`, publishing links to unbuilt pages. Kept deliberately live and unfixed as a repeatable benchmark; the platform fix is known and can be applied by hand any time.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#★ F0 PILOT, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#benchmark bug
- **relations:** standing hypothesis (refuted); mark_no_sections gap; two intake paths disagreement; platform-not-site-data fix philosophy
- **verify-later:** pages/site_work_items rows for site 5fe8785b-223d-41a3-88ee-c07187622381; page-build-handler workflow JSON

<!-- SOURCE: U20_legacy_docs_a.md -->
### Robot Hands website — first agent-built multi-page site
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** Working group definitions (robot-hands-website v1 with usage rows dated 2025-10-27/30), then robot-hands-complete-website (home/about/contact) with full workflow and trigger scripts.
- **what:** The platform's first end-to-end site build: an agent group spawning hero writer, image creator, about writer and contact writer; generating content and a Stability-AI hero image; assembling three HTML pages via aggregate_webpage with embedded CSS/nav; about page explicitly explains the site was built by AI agents (and "may be for sale"). Served as the proving ground for job topics, data helpers and image handling.
- **sources:** docs002_hitl_parallel/README.0100b.updated_state_of_play_for_creating_website.md; docs002_hitl_parallel/README.0100c.workflow_diagram.md; docs002_hitl_parallel/README.0100d.robot_hands_website_readme.md; docs001_flow_general/README.096b.robothandswebsite.md
- **relations:** aggregate_webpage; content-creator-about/contact agents; platform concepts evidenced: job topics, group workflows.
- **verify-later:** does robot-hands.com exist/what pipeline now owns it.

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### relojistas.com go-live + bot verdict
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-13 "Relojistas VERDICT from the access log: 14,961 reqs, 83% 404s … Human intent ≈ 0. Clean probe result (domain not worth building), not a measurement failure".
- **what:** First live domain: a Spanish watch FORUM (grounded in the Wayback snapshot), hand-made `relojistas-site/` (index + gracias, kind=search, THANKS_PATH=/gracias.html) to unblock go-live. After going live (and later Cloudflare-proxied), the access log showed overwhelmingly bot/crawler traffic (Chrome-spoof crawler, Claude-SearchBot, Semrush, Yandex) with ~0 human intent — a clean negative probe result. Later decided to static-build it anyway (RSS + crawler presence + 404/referer signal are assets).
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-relojistas-go-live-bundle, traffic_probe_running_notes(27).md#2026-06-13-relojistas-verdict, traffic_probe_running_notes(27).md#2026-06-13-b
- **relations:** hand-instance of intent-probe component; drove the passive access-log harvest decision
- **verify-later:** deploy_setup/relojistas-site/{index,gracias}.html; relojistas_notes/relojistas_golive.md

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### wayfaringlondoner.com page + THANKS_PATH-is-engine-wide
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13 "wayfaringlondoner.com page built … a 2015–16 travel blog … BLOG framing"; "Design point — THANKS_PATH is engine-wide".
- **what:** Second hand-made page: a 2015–16 travel blog (Csilla; London + Bangkok/Transylvania/Jersey), BLOG framing asking for a destination/London spot/story, tagline gained "and under new ownership". Targets the SHARED multi-vhost box. Surfaced the constraint that THANKS_PATH is one engine-wide env var, so domains on a shared box must share a thanks filename (standard `/thanks.html`); relojistas keeps `/gracias.html` on its own box.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-relojistas-verdict, traffic_probe_running_notes(27).md#2026-06-13-b
- **relations:** shared-box multi-domain onboarding
- **verify-later:** wayfaringlondoner-site/; wayfaringlondoner_notes.md (live)

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Original first-domain set (dropped surgerylight + finance/retail)
- **category:** site-case-studies
- **status-signal:** abandoned
- **status-evidence:** runbook base §3 "Suggested first set: relojistas.com, wayfaringlondoner.com, surgerylight.com, plus one finance tool and one clear retail" — absent from runbook(12) §3 which names only relojistas.
- **what:** The earliest runbook proposed a concrete 3–5 domain starter set (relojistas.com, wayfaringlondoner.com, surgerylight.com, plus a finance tool and a clear retail), each grounded via Wayback. Later versions dropped the named list down to relojistas + wayfaringlondoner; surgerylight and the finance/retail candidates silently vanished.
- **sources:** traffic_probe_runbook.md#3, traffic_probe_runbook(2).md#3, traffic_probe_plan(11).md#risks
- **relations:** relates to Wayback grounding method
- **verify-later:** n/a

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### idea.uk — AI ideation-as-a-service product
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)": "idea.uk runs as a single Go binary under systemd on a Hetzner box... Billing: Stripe Checkout — a single £29 payment per report, live and earning (proven end-to-end with a real card on 2026-06-14)."
- **what:** A paid tool at idea.uk that runs an internal ideation method (generate → cut → web-verify → score → rank) against a business domain + audience, producing a ranked report of business-idea candidates with citations. Sold as one-off £29 reports (down from an initial £199 concept) via a request-then-confirm flow with a free "audience-check" taster as the hook. Positioned as the dogfood/first customer of the idea-generation method itself.
- **sources:** `RUNBOOK_idea_uk(10).md`, `RUNBOOK_idea_uk(1).md`, `running_notes(44).md` (checkpoint 2026-05-28 pricing section, "Pricing settled for the idea.uk product")
- **relations:** idea generation method; Go engine supersedes Python; REVIEW_BEFORE_PAY billing flow; five-layer consolidation model
- **verify-later:** live idea.uk site; `idea-go/` module if present in the working tree; Stripe dashboard for the two named accounts

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Idea generation method — versioned pipeline (v0 → v3)
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** `running_notes(44).md`: v0 → v1 (durability factor + named free substitute) → v2 ("multi-lens generation + richer capability menu... audience-fit challenge... seller-bundles-support-free check") → v3 (Risk column added as a 6th factor, see separate entry). "Method v2 changes derived from the test" and "Method v2 changes — multi-lens generation" sections.
- **what:** The core reusable pipeline: generate (multi-lens: asset×capability, demand, generalist-failure, frontier, outcome) → cut (challenge against the *specific* free substitute + audience-fit challenge + seller-bundles-support-free check) → web-verify → score (Defensibility/Willingness/Buildability/Reuse/Durability, gate Def≥3 AND Will≥3) → rank. Each version fixed a concrete failure found by running the method against real domains (agritec.uk, gaswholesalers.com, robot-hands.com, websitedesign).
- **sources:** `running_notes(44).md` lines ~143-411 (method v1/v2 evolution), `idea_uk_method_v0` family (out of this unit's scope, referenced)
- **relations:** Risk-as-hazard scoring dimension; capability + event watchlists; moat/differentiator framework; cross-vendor critique
- **verify-later:** `idea_method_prompt.md`, `idea_uk_method_v0.md` (live), `idea-go/prompts.go`

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Risk-as-hazard scoring dimension
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md` "CHECKPOINT 2026-05-28 (continued — Risk column added...)": "The rubric had no dimension for the consequence of being wrong. It was caught on operator instinct, which doesn't scale," followed by implementation in `idea-go/engine.go`/`prompts.go` and Python parity, "Built + vetted + tested clean."
- **what:** A 6th scoring factor (1-5, 5=safest) scoring *consequence of being wrong*, deliberately kept separate from the fitness sum (Def+Will+Build+Reuse+Dur) so it can't be gamed by high fitness. Risk=1 auto-drops the candidate into a separate "Dropped for operator risk" section; Risk≤2 still advances but flagged "⚠ needs liability work before building"; Risk is a rank tiebreaker at equal fitness. Triggered by a near-miss: SFI single-farm assessment scored a confident test-now recommendation that could have cost a farmer £5k-50k if wrong.
- **sources:** `running_notes(44).md` (Risk rubric table + rules), `LIABILITY_AND_TERMS.md` (referenced, live)
- **relations:** idea generation method; LIABILITY_AND_TERMS / legal pages
- **verify-later:** `idea-go/engine.go` `scored` struct, `idea_method_runner.py` parity implementation

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Go engine supersedes Python reference implementation
- **category:** site-case-studies
- **status-signal:** superseded
- **status-evidence:** `RUNBOOK_idea_uk(1).md` (archive) describes `idea-go/engine.go` + `prompts.go` + `service.go` + `store.go` + `billing.go` + `main.go` as the whole stack; the live `RUNBOOK_idea_uk.md` base file's equivalent table names only `idea_method_runner.py` / `idea_service.py` / `test_idea_flow.py`. `running_notes(44).md` confirms directly: "Ported the idea.uk tooling from Python to Go (platform is Go throughout)... The Python files remain as the reference implementation but Go is now the canonical version, consistent with the rest of the platform."
- **what:** idea.uk's engine + service were first built in Python (FastAPI, `idea_method_runner.py`, `idea_service.py`, sqlite via `test_idea_flow.py`), validated end-to-end (20/20 checks), then rewritten in idiomatic stdlib-only Go (no external deps, `go vet`/`go build`/`go test` all clean, 19/19 checks) to match "the rest of the platform," which is Go throughout. The rewrite preserved the id-based (not title-based) threading bug-fix and the cross-vendor cut design. This is a genuine, confirmed language-migration superseded/replaced-by relationship, one of very few in this corpus with byte-level before/after evidence.
- **sources:** `RUNBOOK_idea_uk(1).md` §pieces table vs live `RUNBOOK_idea_uk.md`; `running_notes(44).md` ("Ported the idea.uk tooling from Python to Go")
- **relations:** idea.uk product; idea.uk deployment topology
- **verify-later:** confirm whether `idea-go/` or the Python files are what's actually running in production today (the archive/live diff plus running_notes both say Go is canonical, but verify on the actual box)

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Cross-vendor critique (multi-model critique step)
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md`: "Cross-vendor critique implemented (was the one untested multi-model claim). The runner's cut step now routes through OpenAI if OPENAI_API_KEY is set... else falls back to a different Anthropic model." Later ported into the Go engine unchanged.
- **what:** The idea-generation method's "cut" (critique) step deliberately runs on a *different* model vendor than the "generate" step where possible (OpenAI if `OPENAI_API_KEY` is set, else a different Anthropic model as a same-vendor fallback), specifically to avoid "the same model marking its own homework." A stderr log line (`[cut] cross-vendor: OpenAI (gpt-4o)` / `[cut] same-vendor: Anthropic (...)`) was added after user confusion about which vendor actually ran, so every run states its own critique provenance.
- **sources:** `running_notes(44).md` ("Built out the gaps + ran the flow end-to-end", "Added a [cut] vendor log line to engine.go")
- **relations:** idea generation method
- **verify-later:** `idea-go/engine.go` `call_other_vendor` / cross-vendor branch

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### idea.uk request-then-confirm intake with capacity throttle
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(1).md` "Flow (request-then-confirm)"; `running_notes(44).md`: "Capacity throttle (protects the 72h promise): MAX_ACTIVE_ORDERS caps orders in flight."
- **what:** The customer-facing order flow deliberately never takes payment until an operator has screened the request: submit (free) → operator `/confirm` (creates Stripe Checkout / or, post-REVIEW_BEFORE_PAY, runs the engine first) → `/decline` available at any point with a polite no-charge email → webhook-driven fulfilment → operator-reviewed delivery (`AUTO_DELIVER` off by default). A `MAX_ACTIVE_ORDERS` throttle returns HTTP 409 "at capacity" once too many orders are in flight, protecting a stated 72-hour turnaround promise; `/capacity` is a public endpoint so the page can show "currently full."
- **sources:** `RUNBOOK_idea_uk(1).md`; `running_notes(44).md` ("Built out the gaps + ran the flow end-to-end")
- **relations:** REVIEW_BEFORE_PAY billing flow; idea.uk product
- **verify-later:** `idea-go/service.go` capacity/throttle logic

<!-- SOURCE: U25_leopardess_social.md -->
### Leopardess rebuild programme (phases L0–L9)
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** PLAN phase table (2026-07-12, turn 13): "L0–L4 done, L5 partial, L6–L8 not started, L9 ongoing"; HANDOFF §3 "main pages live and verified".
- **what:** Rebuild of leopardessconsulting.co.uk (a site the platform built for itself) to be honest, well-branded and useful: evidence audit (L0), spec truth pass (L1), brand/logo (L2), palette fork (L3), 3-per-row layout (L4), copy rewrite (L5), explanatory imagery (L6), charts (L7), tools/guides/news build-out (L8), coherent deploy (L9). Includes the audience pivot (A2): sceptical, commercially-sharp, non-specialist buyer, with technical depth one click down.
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#Phases, docs/leopardessconsulting/HANDOFF.md#0, docs/leopardessconsulting/RUNNING_NOTES.md#Decision-log
- **relations:** claim-evidence audit rule; anti-hype voice spec; per-site style fork; chart component (Go SVG + JS)
- **verify-later:** site_id 4851f6fc-71cf-4160-a270-e03d6d3e0732 pages/build_status; live leopardessconsulting.co.uk pages

<!-- SOURCE: U25_leopardess_social.md -->
### Claim-evidence audit rule ("no claim ships without an audit row")
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** AUDIT header 2026-07-09: "no claim ships unless it has a row in this table"; fabrication sweep 2026-07-10 "Result: CLEAN".
- **what:** A site-truth methodology: every marketing claim is verified against code, live Postgres or an HTTP response before it may appear on the site; unverifiable claims are removed or hedged explicitly (the one allowed unproven claim — "third busiest sports site" — is published labelled as recollection). Produced a verified capability inventory (C1–C6: Companies House cascade with 2,767 verified businesses; news pipeline 5,652 items/4,672 scored; tool-generation agent family; DB-defined hierarchical agents 143 defs/56 active/40 spawners; Banana+SDXL imagery; 8 own sites) and an UNSUPPORTED list (U1–U11).
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#1, #2; docs/leopardessconsulting/PLAN_leopardess_rebuild.md#Standing-rules
- **relations:** LLM fabrication classes; verify-by-artifact operator discipline; per-category platform pipelines (Companies House, news feed, tool pipeline — this audit is dated deployment evidence for them)
- **verify-later:** business_intel.businesses counts; feed item counts; agent_definitions counts

<!-- SOURCE: U25_leopardess_social.md -->
### Reuse-not-rebuild site build-out with honest "simulation" labelling
- **category:** site-case-studies
- **status-signal:** aspirational
- **status-evidence:** PLAN L8 "not started — reuse existing tool library; surface the live news feed"; HANDOFF §6.5.
- **what:** The leopardess L8 plan: deploy/adopt existing interactive tool components (ROI estimator, quizzes, calculators — deterministic client-side widgets that must be labelled as simulations, not live inference), surface the real news-feed pipeline, pair guides with tools (tool-deployer already creates companion guides + cross-links), and note a "game" has no formal platform existence — it is simply a component_level='tool' component.
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#L8; docs/leopardessconsulting/HANDOFF.md#6
- **relations:** tool-library; news-feed-pipeline; dynamic-applications
- **verify-later:** tool component inventory; news feed surfacing on any site

<!-- SOURCE: U03_idea_uk_section_data.md -->
### idea.uk live-VM / chassis-staging duality
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** Journal project facts: "idea.uk — LIVE Go service selling £29 reports; single binary under systemd on a Hetzner VM … DNS (Cloudflare) → the VM, so chassis B2 deploys are invisible to the live site. UNCHANGED." Every checkpoint ends "idea.uk live VM untouched."
- **what:** The genuinely site-specific arrangement underpinning the whole unit's risk model: the revenue-earning idea.uk product (£29 reports, live Stripe webhook, orders in a file, reserved paths /request /confirm /approve /decline /stripe/webhook /internal/* /order/*) is a separate Go binary on a Hetzner VM; the chassis-built idea.uk site deploys to Backblaze B2 while DNS still points at the VM — so all chassis work is invisible staging and the VM cutover is a separate future decision. Two chassis site_ids exist for idea.uk (97ed2f64-… in the June thread, 1244516d-… in the July thread) — treated as separate/earlier rows, confirm before relying on either.
- **sources:** running_notes_scheme_to_components(55).md#Project-facts; HANDOFF_scheme_to_components_for_claude_code(1).md#Environment; HANDOFF_idea_uk_differentiators_section_data.md#Key-facts
- **relations:** platform mission; chassis deploy model.
- **verify-later:** sites rows for idea.uk (both ids); DNS state if a cutover is ever planned.

<!-- SOURCE: U04_idea_uk.md -->
### idea.uk mission and identity (workshop of tools; never verdicts; warm-paper identity)
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** "idea.uk mission REFRAMED away from the £29 tool… shipped as a file" and submitted via --mission-file 2026-06-21; classifier "read the mission well" on site 1244516d.
- **what:** The site's genuinely site-specific concept: idea.uk = the place to take an idea seriously — a growing workshop of genuinely good tools (the main event; free + paid), tools labelled **private (in-browser, nothing sent)** vs **AI/hosted** with private leading; cutting-edge succinct research-grounded guides; a news section; **never verdicts** — perspective, evidence, and questions framed as opinion (the legal reframe); the £29 verified report demoted to one specialised flagship tool; later a build-and-bring-to-market service; preserve the warm-paper/ink/single-rust-accent/Fraunces+IBM-Plex editorial identity. Noted honestly: the privacy and latest-research promises are stated intent the chassis can't yet enforce.
- **sources:** idea.uk/KEY_DOC_idea_uk_mission.txt; idea.uk/HANDOFF(13).md (idea.uk current state); idea.uk/running_notes(63).md (nnn/ooo 06-21)
- **relations:** liability framework (never-verdicts); mission-file mechanism; standing-ambition.
- **verify-later:** site 1244516d mission spec content.

<!-- SOURCE: U04_idea_uk.md -->
### idea.uk chassis-site build state (two site rows; staging-only; gated go-live)
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** Current position 2026-06-26: composition + stylesheet correct and verified (tool-portal-light + parchment, commit 05ef817, no LLM drift); pages still dark; rebuild + review + cutover gated on the P0 scheme thread.
- **what:** The concrete case-study state: first chassis run under site 97ed2f64 (2026-06-14: classifier→…→empty index → coordinator fix validated on it); old chassis site torn down and resubmitted fresh 2026-06-21 as site 1244516d with the mission file; re-resolved onto tool-portal-light 2026-06-25; deployed page defects catalogued (empty differentiators, unresolved CTAs, dead contact form, missing pricing spec, thin nav/footer, empty meta description, dark chrome). The live £29 VM tool is a separate stream, untouched and earning throughout — the safety property the whole thread leans on.
- **sources:** idea.uk/TODO_chassis_and_idea_uk(1).md; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (open items + current position); idea.uk/HANDOFF(13).md
- **relations:** every design/pipeline concept above (this is their test case); VM cutover.
- **verify-later:** current state of site 1244516d; whether the P0 thread completed and cutover happened.

<!-- SOURCE: U10_imagery.md -->
### robot-hands.com rebuild (testbed case study)
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** "Phase I0 ✅ COMPLETE. 33-page rebuild w/ live news (9 sources); 16 distinct per-page git-path heroes, zero expiring URLs" (2026-07-12).
- **what:** The adopted site's content layer was broken (10 zero-component pages, NULL content, features content in a pre-drift schema) while the imagery pipeline was correct — so it was rebuilt from scratch with news scope: supersede adoption-residue aspects, news-enable classification, add a mission_brief aspect, retire stale items, insert+manually-triage a needs_site_plan trigger, then a fresh 33-page plan (29 imagery rows) built unattended through dispatch. A 2026-05-20 audit first said "PATCH, do not re-plan" (foundation sound, build broken); the 2026-07-08 decision superseded it with full re-plan. Hard requirements: tools must actually work (deployed JS, resolving links) and it is the acceptance surface for all I-phases.
- **sources:** HANDOFF_robot_hands_rebuild.md, SQL_2026-07-08_robothands_rebuild_prep.sql, SQL_2026-07-08_robothands_mission_brief.sql, RUNNING_NOTES_imagery_best_in_class.md#Turns-2–4
- **relations:** news enrichment pattern; schema-contract drift bug (features {title,description} vs {icon,name,description}); orphan pre-rebuild pages cleanup still open.
- **verify-later:** current plan 7a40a0f9; page/component counts; tool pages' build_status.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Dartsonline guides defect (benchmark bug, causes A/B/C)
- **category:** site-case-studies
- **status-signal:** deployed (diagnosis confirmed; fix deliberately not applied)
- **status-evidence:** "The benchmark bug (still live, still unfixed — deliberately)" (HANDOFF_CURRENT_fixloop.md)
- **what:** dartsonline.com published a Guides nav link to a blank page. Root mechanism, hand-diagnosed with citations: (A) `build-site-planner` populated `sections` for only 5 of 15 planned pages; (B) `page-build-handler`'s `check_has_ready_sections` routes sectionless pages to `complete_error`, which is `action: complete_workflow` — a success terminal — so the work item is marked complete though the page was never built; (C) `populate_nav_tables_action.go` filters nav candidates on `pages.status` (lifecycle, defaults 'active') instead of `build_status`, publishing links to unbuilt pages. Kept deliberately live and unfixed as a repeatable benchmark; the platform fix is known and can be applied by hand any time.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#★ F0 PILOT, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#benchmark bug
- **relations:** standing hypothesis (refuted); mark_no_sections gap; two intake paths disagreement; platform-not-site-data fix philosophy
- **verify-later:** pages/site_work_items rows for site 5fe8785b-223d-41a3-88ee-c07187622381; page-build-handler workflow JSON

<!-- SOURCE: U20_legacy_docs_a.md -->
### Robot Hands website — first agent-built multi-page site
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** Working group definitions (robot-hands-website v1 with usage rows dated 2025-10-27/30), then robot-hands-complete-website (home/about/contact) with full workflow and trigger scripts.
- **what:** The platform's first end-to-end site build: an agent group spawning hero writer, image creator, about writer and contact writer; generating content and a Stability-AI hero image; assembling three HTML pages via aggregate_webpage with embedded CSS/nav; about page explicitly explains the site was built by AI agents (and "may be for sale"). Served as the proving ground for job topics, data helpers and image handling.
- **sources:** docs002_hitl_parallel/README.0100b.updated_state_of_play_for_creating_website.md; docs002_hitl_parallel/README.0100c.workflow_diagram.md; docs002_hitl_parallel/README.0100d.robot_hands_website_readme.md; docs001_flow_general/README.096b.robothandswebsite.md
- **relations:** aggregate_webpage; content-creator-about/contact agents; platform concepts evidenced: job topics, group workflows.
- **verify-later:** does robot-hands.com exist/what pipeline now owns it.

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### relojistas.com go-live + bot verdict
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-13 "Relojistas VERDICT from the access log: 14,961 reqs, 83% 404s … Human intent ≈ 0. Clean probe result (domain not worth building), not a measurement failure".
- **what:** First live domain: a Spanish watch FORUM (grounded in the Wayback snapshot), hand-made `relojistas-site/` (index + gracias, kind=search, THANKS_PATH=/gracias.html) to unblock go-live. After going live (and later Cloudflare-proxied), the access log showed overwhelmingly bot/crawler traffic (Chrome-spoof crawler, Claude-SearchBot, Semrush, Yandex) with ~0 human intent — a clean negative probe result. Later decided to static-build it anyway (RSS + crawler presence + 404/referer signal are assets).
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-relojistas-go-live-bundle, traffic_probe_running_notes(27).md#2026-06-13-relojistas-verdict, traffic_probe_running_notes(27).md#2026-06-13-b
- **relations:** hand-instance of intent-probe component; drove the passive access-log harvest decision
- **verify-later:** deploy_setup/relojistas-site/{index,gracias}.html; relojistas_notes/relojistas_golive.md

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### wayfaringlondoner.com page + THANKS_PATH-is-engine-wide
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13 "wayfaringlondoner.com page built … a 2015–16 travel blog … BLOG framing"; "Design point — THANKS_PATH is engine-wide".
- **what:** Second hand-made page: a 2015–16 travel blog (Csilla; London + Bangkok/Transylvania/Jersey), BLOG framing asking for a destination/London spot/story, tagline gained "and under new ownership". Targets the SHARED multi-vhost box. Surfaced the constraint that THANKS_PATH is one engine-wide env var, so domains on a shared box must share a thanks filename (standard `/thanks.html`); relojistas keeps `/gracias.html` on its own box.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-relojistas-verdict, traffic_probe_running_notes(27).md#2026-06-13-b
- **relations:** shared-box multi-domain onboarding
- **verify-later:** wayfaringlondoner-site/; wayfaringlondoner_notes.md (live)

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Original first-domain set (dropped surgerylight + finance/retail)
- **category:** site-case-studies
- **status-signal:** abandoned
- **status-evidence:** runbook base §3 "Suggested first set: relojistas.com, wayfaringlondoner.com, surgerylight.com, plus one finance tool and one clear retail" — absent from runbook(12) §3 which names only relojistas.
- **what:** The earliest runbook proposed a concrete 3–5 domain starter set (relojistas.com, wayfaringlondoner.com, surgerylight.com, plus a finance tool and a clear retail), each grounded via Wayback. Later versions dropped the named list down to relojistas + wayfaringlondoner; surgerylight and the finance/retail candidates silently vanished.
- **sources:** traffic_probe_runbook.md#3, traffic_probe_runbook(2).md#3, traffic_probe_plan(11).md#risks
- **relations:** relates to Wayback grounding method
- **verify-later:** n/a

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### idea.uk — AI ideation-as-a-service product
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)": "idea.uk runs as a single Go binary under systemd on a Hetzner box... Billing: Stripe Checkout — a single £29 payment per report, live and earning (proven end-to-end with a real card on 2026-06-14)."
- **what:** A paid tool at idea.uk that runs an internal ideation method (generate → cut → web-verify → score → rank) against a business domain + audience, producing a ranked report of business-idea candidates with citations. Sold as one-off £29 reports (down from an initial £199 concept) via a request-then-confirm flow with a free "audience-check" taster as the hook. Positioned as the dogfood/first customer of the idea-generation method itself.
- **sources:** `RUNBOOK_idea_uk(10).md`, `RUNBOOK_idea_uk(1).md`, `running_notes(44).md` (checkpoint 2026-05-28 pricing section, "Pricing settled for the idea.uk product")
- **relations:** idea generation method; Go engine supersedes Python; REVIEW_BEFORE_PAY billing flow; five-layer consolidation model
- **verify-later:** live idea.uk site; `idea-go/` module if present in the working tree; Stripe dashboard for the two named accounts

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Idea generation method — versioned pipeline (v0 → v3)
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** `running_notes(44).md`: v0 → v1 (durability factor + named free substitute) → v2 ("multi-lens generation + richer capability menu... audience-fit challenge... seller-bundles-support-free check") → v3 (Risk column added as a 6th factor, see separate entry). "Method v2 changes derived from the test" and "Method v2 changes — multi-lens generation" sections.
- **what:** The core reusable pipeline: generate (multi-lens: asset×capability, demand, generalist-failure, frontier, outcome) → cut (challenge against the *specific* free substitute + audience-fit challenge + seller-bundles-support-free check) → web-verify → score (Defensibility/Willingness/Buildability/Reuse/Durability, gate Def≥3 AND Will≥3) → rank. Each version fixed a concrete failure found by running the method against real domains (agritec.uk, gaswholesalers.com, robot-hands.com, websitedesign).
- **sources:** `running_notes(44).md` lines ~143-411 (method v1/v2 evolution), `idea_uk_method_v0` family (out of this unit's scope, referenced)
- **relations:** Risk-as-hazard scoring dimension; capability + event watchlists; moat/differentiator framework; cross-vendor critique
- **verify-later:** `idea_method_prompt.md`, `idea_uk_method_v0.md` (live), `idea-go/prompts.go`

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Risk-as-hazard scoring dimension
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md` "CHECKPOINT 2026-05-28 (continued — Risk column added...)": "The rubric had no dimension for the consequence of being wrong. It was caught on operator instinct, which doesn't scale," followed by implementation in `idea-go/engine.go`/`prompts.go` and Python parity, "Built + vetted + tested clean."
- **what:** A 6th scoring factor (1-5, 5=safest) scoring *consequence of being wrong*, deliberately kept separate from the fitness sum (Def+Will+Build+Reuse+Dur) so it can't be gamed by high fitness. Risk=1 auto-drops the candidate into a separate "Dropped for operator risk" section; Risk≤2 still advances but flagged "⚠ needs liability work before building"; Risk is a rank tiebreaker at equal fitness. Triggered by a near-miss: SFI single-farm assessment scored a confident test-now recommendation that could have cost a farmer £5k-50k if wrong.
- **sources:** `running_notes(44).md` (Risk rubric table + rules), `LIABILITY_AND_TERMS.md` (referenced, live)
- **relations:** idea generation method; LIABILITY_AND_TERMS / legal pages
- **verify-later:** `idea-go/engine.go` `scored` struct, `idea_method_runner.py` parity implementation

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Go engine supersedes Python reference implementation
- **category:** site-case-studies
- **status-signal:** superseded
- **status-evidence:** `RUNBOOK_idea_uk(1).md` (archive) describes `idea-go/engine.go` + `prompts.go` + `service.go` + `store.go` + `billing.go` + `main.go` as the whole stack; the live `RUNBOOK_idea_uk.md` base file's equivalent table names only `idea_method_runner.py` / `idea_service.py` / `test_idea_flow.py`. `running_notes(44).md` confirms directly: "Ported the idea.uk tooling from Python to Go (platform is Go throughout)... The Python files remain as the reference implementation but Go is now the canonical version, consistent with the rest of the platform."
- **what:** idea.uk's engine + service were first built in Python (FastAPI, `idea_method_runner.py`, `idea_service.py`, sqlite via `test_idea_flow.py`), validated end-to-end (20/20 checks), then rewritten in idiomatic stdlib-only Go (no external deps, `go vet`/`go build`/`go test` all clean, 19/19 checks) to match "the rest of the platform," which is Go throughout. The rewrite preserved the id-based (not title-based) threading bug-fix and the cross-vendor cut design. This is a genuine, confirmed language-migration superseded/replaced-by relationship, one of very few in this corpus with byte-level before/after evidence.
- **sources:** `RUNBOOK_idea_uk(1).md` §pieces table vs live `RUNBOOK_idea_uk.md`; `running_notes(44).md` ("Ported the idea.uk tooling from Python to Go")
- **relations:** idea.uk product; idea.uk deployment topology
- **verify-later:** confirm whether `idea-go/` or the Python files are what's actually running in production today (the archive/live diff plus running_notes both say Go is canonical, but verify on the actual box)

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Cross-vendor critique (multi-model critique step)
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md`: "Cross-vendor critique implemented (was the one untested multi-model claim). The runner's cut step now routes through OpenAI if OPENAI_API_KEY is set... else falls back to a different Anthropic model." Later ported into the Go engine unchanged.
- **what:** The idea-generation method's "cut" (critique) step deliberately runs on a *different* model vendor than the "generate" step where possible (OpenAI if `OPENAI_API_KEY` is set, else a different Anthropic model as a same-vendor fallback), specifically to avoid "the same model marking its own homework." A stderr log line (`[cut] cross-vendor: OpenAI (gpt-4o)` / `[cut] same-vendor: Anthropic (...)`) was added after user confusion about which vendor actually ran, so every run states its own critique provenance.
- **sources:** `running_notes(44).md` ("Built out the gaps + ran the flow end-to-end", "Added a [cut] vendor log line to engine.go")
- **relations:** idea generation method
- **verify-later:** `idea-go/engine.go` `call_other_vendor` / cross-vendor branch

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### idea.uk request-then-confirm intake with capacity throttle
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(1).md` "Flow (request-then-confirm)"; `running_notes(44).md`: "Capacity throttle (protects the 72h promise): MAX_ACTIVE_ORDERS caps orders in flight."
- **what:** The customer-facing order flow deliberately never takes payment until an operator has screened the request: submit (free) → operator `/confirm` (creates Stripe Checkout / or, post-REVIEW_BEFORE_PAY, runs the engine first) → `/decline` available at any point with a polite no-charge email → webhook-driven fulfilment → operator-reviewed delivery (`AUTO_DELIVER` off by default). A `MAX_ACTIVE_ORDERS` throttle returns HTTP 409 "at capacity" once too many orders are in flight, protecting a stated 72-hour turnaround promise; `/capacity` is a public endpoint so the page can show "currently full."
- **sources:** `RUNBOOK_idea_uk(1).md`; `running_notes(44).md` ("Built out the gaps + ran the flow end-to-end")
- **relations:** REVIEW_BEFORE_PAY billing flow; idea.uk product
- **verify-later:** `idea-go/service.go` capacity/throttle logic

<!-- SOURCE: U25_leopardess_social.md -->
### Leopardess rebuild programme (phases L0–L9)
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** PLAN phase table (2026-07-12, turn 13): "L0–L4 done, L5 partial, L6–L8 not started, L9 ongoing"; HANDOFF §3 "main pages live and verified".
- **what:** Rebuild of leopardessconsulting.co.uk (a site the platform built for itself) to be honest, well-branded and useful: evidence audit (L0), spec truth pass (L1), brand/logo (L2), palette fork (L3), 3-per-row layout (L4), copy rewrite (L5), explanatory imagery (L6), charts (L7), tools/guides/news build-out (L8), coherent deploy (L9). Includes the audience pivot (A2): sceptical, commercially-sharp, non-specialist buyer, with technical depth one click down.
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#Phases, docs/leopardessconsulting/HANDOFF.md#0, docs/leopardessconsulting/RUNNING_NOTES.md#Decision-log
- **relations:** claim-evidence audit rule; anti-hype voice spec; per-site style fork; chart component (Go SVG + JS)
- **verify-later:** site_id 4851f6fc-71cf-4160-a270-e03d6d3e0732 pages/build_status; live leopardessconsulting.co.uk pages

<!-- SOURCE: U25_leopardess_social.md -->
### Claim-evidence audit rule ("no claim ships without an audit row")
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** AUDIT header 2026-07-09: "no claim ships unless it has a row in this table"; fabrication sweep 2026-07-10 "Result: CLEAN".
- **what:** A site-truth methodology: every marketing claim is verified against code, live Postgres or an HTTP response before it may appear on the site; unverifiable claims are removed or hedged explicitly (the one allowed unproven claim — "third busiest sports site" — is published labelled as recollection). Produced a verified capability inventory (C1–C6: Companies House cascade with 2,767 verified businesses; news pipeline 5,652 items/4,672 scored; tool-generation agent family; DB-defined hierarchical agents 143 defs/56 active/40 spawners; Banana+SDXL imagery; 8 own sites) and an UNSUPPORTED list (U1–U11).
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#1, #2; docs/leopardessconsulting/PLAN_leopardess_rebuild.md#Standing-rules
- **relations:** LLM fabrication classes; verify-by-artifact operator discipline; per-category platform pipelines (Companies House, news feed, tool pipeline — this audit is dated deployment evidence for them)
- **verify-later:** business_intel.businesses counts; feed item counts; agent_definitions counts

<!-- SOURCE: U25_leopardess_social.md -->
### Reuse-not-rebuild site build-out with honest "simulation" labelling
- **category:** site-case-studies
- **status-signal:** aspirational
- **status-evidence:** PLAN L8 "not started — reuse existing tool library; surface the live news feed"; HANDOFF §6.5.
- **what:** The leopardess L8 plan: deploy/adopt existing interactive tool components (ROI estimator, quizzes, calculators — deterministic client-side widgets that must be labelled as simulations, not live inference), surface the real news-feed pipeline, pair guides with tools (tool-deployer already creates companion guides + cross-links), and note a "game" has no formal platform existence — it is simply a component_level='tool' component.
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#L8; docs/leopardessconsulting/HANDOFF.md#6
- **relations:** tool-library; news-feed-pipeline; dynamic-applications
- **verify-later:** tool component inventory; news feed surfacing on any site
