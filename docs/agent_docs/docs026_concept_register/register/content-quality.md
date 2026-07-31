# Register — content-quality

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

17 concepts, consolidated from 40 raw extractions (20 unique blocks, each duplicated
once in the source cluster file) across units U02, U05, U07, U09, U13, U15, U19,
U20, U21, U24a, U24c, U24d, U25.

### CQ-001 — Content-quality defect catalogue (gamesdesign.co.uk)
- **status:** partial
- **status-evidence:** FOCUS_content_quality(2) 2026-06-10 status table: defect 1 "addressed (Step 1)", defects 2-5 "open"; HANDOFF_2026-06-15 §7 lists the parked items "next package, EXPECTED to recur on readopt" — this supersedes the earlier 2026-06-09 snapshot which had no defects yet addressed.
- **what:** A maintained catalogue of content-correctness defects on built pages (CATALOGUE_gamesdesign_post_sync_fix_defects as source of record): hero CTA text↔destination mismatch site-wide (lead item), tool-flavoured guide copy, brand-suffix in card titles, empty footer brand-tagline/contact, empty tool descriptions, empty meta descriptions. Content quality is explicitly scoped as "words and per-component data" — distinct from design fidelity and from link destinations. Work order: settle CTA field-vs-template first (reuse component-template-fixer's CTA handling, later found not to exist — see CQ-014), then footer/titles/descriptions batch, then guide re-flavouring. Defects are worked as separate threads, each read-the-code-first. The three-way finding classification and specialist agents referenced in the original work order were flagged as PROPOSED, not confirmed built.
- **sources:** FOCUS_content_quality.md (2026-06-09); content_quality_and_internal_linking/FOCUS_content_quality(2).md (2026-06-10); HANDOFF_2026-06-15_index_stale_rebuild(2).md#7; running_notes_17_internal_linking_phantom_fixes(21).md#content-quality-observations
- **relations:** recommendation specialist architecture (content-governance CGV-026); internal linking; validate_page_content gate (CQ-002); component-template-fixer CTA assumption (CQ-014)
- **verify-later:** CATALOGUE_gamesdesign_post_sync_fix_defects(9).md; live gamesdesign.co.uk pages; site_work_items for content_rewrite items

### CQ-002 — validate_page_content gate (pre-deploy content validator)
- **status:** deployed
- **status-evidence:** docs018/003 (design-era): "validate_page_content action runs before review mode determination... Validation errors force HITL review, blocking auto-approval." Confirmed live and fixed by 2026-06-09/10: "the content validator and the gate … routes validate_content error_step → mark_needs_review → needs_human_review (Mode 2 of the silent-completion work, already confirmed fixed)"; machinery "verified this round" per FOCUS_content_quality(2).
- **what:** Blocker-detecting validator (placeholder text, unrendered templates, empty required sections, cross-site contamination, broken internal links, unverified emails) that any content fix must pass; failures route consistently to human review (mark_needs_review → needs_human_review). `validateInternalLinks` (now on datahelpers) flags `phantom_link` and `empty_internal_href` as non-blocking warnings, tolerating planned-but-unbuilt pages. Companion mechanisms: prepare_link_context injects an only-link-to-these-pages allowlist into writer prompts; rerender-time contact injection replaces hallucinated phone/email with DB truth. Known gaps: adopted content referencing the source domain trips the contamination heuristic (needs an adopted-from whitelist for mode=recreate); legitimate emails (contactforsales.com) also flagged; it does not flag brand-suffix titles, empty contact, or empty descriptions (content/spec issues, not link/placeholder issues).
- **GUARANTEE CHANGED 2026-07-29 — this gate can now refuse a page on a site that never opted in** (`bugs_closed/104`, claims-verification **CLM-015**, RFC 003). Check 8's claims half was opt-in: *"sites without an evidence_base skip both silently."* Nine **fleet-wide** banned-claim patterns are now scanned on **every** site, register or not, at severity **blocker** — so `content-reviewer`'s `validate_content` step can fail a build for a claim no site may make. The unregistered-**number** half is unchanged and still strictly opt-in; the two halves of check 8 now have deliberately different opt-in rules. **Withdrawal is config, not a release:** `check_claims_fleet_wide: false` on the step, live immediately, restores the previous behaviour exactly and logs a Warn naming the site. Newly refutable population when this landed: 6 live sites with no register — dartsonline, gaswholesalers, idea.uk, system.internal, vetcomparison.uk, webdesign.co.uk. Measured 0 findings across all 908 live components before shipping.
- **sources:** FOCUS_content_quality.md#machinery; FOCUS_content_quality(2).md#the-machinery; HANDOFF_2026-04-23(1).md Bug 7; HANDOFF-pipeline-triage-april-2026.md#queue; docs018_rerendering/003_website_builder_architecture_status_report.md#3,#6; docs018_rerendering/002_summary_link_constraints.md
- **relations:** silent completion mode 2; phantom-link check hook candidate; content-quality catalogue gaps (CQ-001); content-reviewer workflow (content-governance CGV-023); link_registry
- **verify-later:** validate_page_content.go blocker classes; adopted-domain whitelist; prepare_link_context in registry

### CQ-003 — Shared-component regen clobber failure mode (silent overwrite of dependent pages)
- **status:** deployed
- **status-evidence:** "RESOLVED + RECOVERED (verified)" (HANDOFF(7), 2026-07-04); "R6b pass 2026-07-03: distinct md5s, needles true"; root cause section dated + confirmed in NOTES §4.
- **what:** Regenerating a shared component overwrites its `input_schema`/`html_template` field contract in place without migrating dependent pages' `content_data` to the new field names; rendering binds by exact field name and silently empties misses, so every dependent renders a content-free shell that the assembler silently drops — fanning out across every page/site sharing the component. Confirmed on `system-stats` (`fdd92ad4`, regen 2026-06-24 15:06): 24 old keys vs 22 new, five live pages on three sites byte-identical empty. `content_data` stayed intact and per-page, so the breakage was recoverable without an LLM.
- **sources:** NOTES_component_regen_clobber(43).md §1, §4, §8; HANDOFF_component_regen_clobber(7).md §Incident 1; RUNBOOK_pre_cleanup_backup.md §The problem
- **relations:** F1 field-contract guard (fix); F3 scoped rerender (repair path, CQ-004); F5/F8 (sibling facets, CQ-005); RenderTemplate silent-empty mechanism
- **verify-later:** platform/orchestration/actions/store_generated_component_action.go (regen branch ~L354-432); content_components/component_versions/page_components tables; component `fdd92ad4-521a-4602-89cf-7ee1a66c10f1`

### CQ-004 — Recovery playbook for stranded dependents (Route A rebuild vs Route B re-key + scoped re-render)
- **status:** deployed
- **status-evidence:** "R5b step 8 PASS — all five RECOVERED and verified" (NOTES §9ad, 2026-07-03); leopardess confirmed live by screenshot (§9ae).
- **what:** Two recovery routes for pages stranded by a shared-contract change: Route A — full `needs_page` writer rebuild (regenerates content_data under the new schema; simplest, costs LLM); Route B — re-key each page's `content_data` old→new (explicit reviewable jsonb_build_object mapping, dry-run first, CTAS backup, non-1:1 fields handled explicitly) then trigger the F3-scoped section re-render (no LLM, preserves per-page values). Route B executed for the five, doubling as F3's end-to-end proof; gated on fleet image, freshness check, and a cta-schema decision.
- **sources:** NOTES(43).md §6, §9q, §9s-§9t, §9ad; RUNBOOK(49).md Part A; PLAN(1).md Phase 4
- **relations:** shared-component regen clobber (CQ-003); section readiness model (the cta_url blocker it hit); optimistic-lock co-management (content-governance CGV-008); snapshot-before-change (CGV-009)
- **verify-later:** page_components content_data keys for the five; backup tables page_components_bak_sysstats_20260702 / _briefexp_20260703 (may be dropped)

### CQ-005 — F8 shared-component contamination: site-specific copy baked into shared machinery (three carriers)
- **status:** partial
- **status-evidence:** "STEP-12 SWEEP PASSES — contamination cleared board-wide" (NOTES §9bg, 2026-07-06) — incident remediated; "F8 mitigation… WHAT: a guard/lint so shared-component fallbacks and llm_guidance must be site-neutral" still an open Part E flag (structural mitigation unbuilt).
- **what:** A pre-guard regen (2026-07-01) baked vonc's product pitch ("Spark", the daily Gauntlet) into the shared `brief-explanation` component via three carriers invisible to the name-only F1 guard: (1) static-field fallback values; (2) those values merged into dependents' content_data by the stored⊕resolved merge; (3) per-field `llm_guidance` — the strongest, actively instructing every future writer pass on any site to write vonc's product (reproduced verbatim on robot-hands and idea.uk; contamination also migrated into generated LLM copy on pages built pre-fix). Remediation playbook executed: snapshot v2/v3 → neutralize fallbacks → strip merged keys with CTAS backup → scoped F3 re-renders → writer rebuilds under cleaned guidance → board-wide strpos sweep (clean except vonc's own legitimate copy). Proposed structural mitigation (unbuilt): store-time site-neutrality lint over fallbacks + llm_guidance.
- **sources:** NOTES(43).md §9an-§9bb, §9bg; RUNBOOK(49).md Part C + Part E F8; HANDOFF(7).md §Incident 2
- **relations:** F1 guard (its blind spot); llm_guidance surface; stored⊕resolved merge; neutralize-in-place remediation (CQ-006); shared-component regen clobber (CQ-003)
- **verify-later:** brief-explanation input_schema (neutral guidance ×11, stats source=llm no fallback); component_versions v1-v3; store-side lint absence

### CQ-006 — Neutralize-in-place remediation pattern for contaminated shared components
- **status:** deployed
- **status-evidence:** "v1 not a restore candidate… NEUTRALIZE-IN-PLACE chosen" (NOTES §9ao); Steps 1-2 landed with optimistic lock held (§9ap); Steps 6-7 landed §9bd.
- **what:** When a shared component's history offers no clean restore point (v1 predated the current field contract; restoring would regress dependents on the new architecture), the fix is surgical in-place neutralization: manual snapshot first, then targeted jsonb patches replacing only the offending attributes (fallbacks, guidance) under an optimistic lock, preserving names/types/structure — followed by per-dependent cleanup (strip merged keys, scoped re-renders, writer rebuilds) mapped per consumer (vonc's own copy untouched; robot-hands stripped; old-architecture pages escalated to rebuilds).
- **sources:** NOTES(43).md §9ao-§9aq, §9bb, §9bd; RUNBOOK(49).md Part C Steps 1-9
- **relations:** F8 (CQ-005); optimistic-lock co-management (content-governance CGV-008); component versioning; recovery playbook (CQ-004)
- **verify-later:** the CTE jsonb_object_agg patch shape in RUNBOOK(49) Step 7 as reusable SQL

### CQ-007 — Adoption content-quality defect families (polish batch)
- **status:** partial
- **status-evidence:** Open Groups 3 items as of HANDOFF_2026-06-09: "GameDesign.uk brand-suffix in card titles; footer mailto/tagline empty; one empty tool description; guide tables render poorly; guides should cross-link to tools"; hero H1 reuse and empty meta descriptions from the catalogue remain untracked-as-fixed.
- **stage2-verified (2026-07-14):** unknown → partial — RUNBOOK_linking_phantom_fixes(5).md:176 (content_quality_and_internal_linking/) explicitly lists these exact defects (brand-suffix titles, tool-flavoured guide copy, empty tool descriptions, footer metadata) as 'EXPECTED recurrences (adopt-path, untouched) ... next package's input' as of a later runbook than the sta...
- **what:** The residual content-quality class after build mechanics were fixed: source-brand `<title>` suffixes used as display names (preserving the source brand, not the destination), empty footer contact/tagline (no graceful no-data path — components render empty structure instead of hiding), hero H1 duplicated across hubs, meta_description populated in DB but emitted empty, tool-flavoured guide copy (user open to real embedded interactive demos in guides), guide→tool cross-linking as enhancement.
- **sources:** CATALOGUE(9)#family-e, HANDOFF_2026-06-09#next-task, running_notes_14(25)#part-14n
- **relations:** silent-fallback link family (content-governance CGV-010); page-content-writer prompts; internal linking
- **verify-later:** current gamesdesign deployed HTML; hero/footer component schemas' no-data paths

### CQ-008 — Post-build validation of structured components (Fix D, unimplemented)
- **status:** aspirational
- **status-evidence:** TODO_remaining_work.md "Post-build validation (Fix D): assert a component whose input_schema declares a required structured field ... actually has it populated before deploy" — listed under open/structural, not done.
- **what:** Proposed check that runs after a build, asserting that any component whose `input_schema` declares a required structured field actually has that field populated in `content_data`; if empty, flags the page instead of deploying a silently-empty component. Catches the bug class regardless of which planner or writer path produced the empty result.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md#D-Post-build-validation, js_snippets_news_gaswholesalers/TODO_remaining_work.md
- **relations:** FAQ duplicate content-surface bug; per-section briefs gap
- **verify-later:** grep/inspect `input_schema`; `content_data`

### CQ-009 — Site-quality programme handoff
- **status:** partial
- **status-evidence:** "site-quality programme HANDED OFF to its own runbook... 0 nav / 0 img / 0 svg / 0 script on ALL pages" (NOTES_running_synthesis_v4(39).md, 2026-07-06).
- **what:** Following the platform's first recorded domain→deployed-site milestone (dartsonline.com), a measured baseline (four rendered pages, all missing nav/images/svg/script, thin CSS variable usage, near-zero internal links) triggered a dedicated handoff (`RUNBOOK_site_quality.md`) splitting remaining work into stuck-dispatch (chrome/design/imagery), delivered-but-poor (content depth, links), and never-in-scope (feeds/RSS/graphics/games, disabled improvement-sweep) — with a live hypothesis that the relay path lacks the monolith's `render_site_components` chrome step, explaining nav-zero across every page.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-06 "MILESTONE recorded" and "site-quality programme HANDED OFF" entries
- **relations:** Work-item relay / builder-generations architecture; diagnosis→fix loop workstream founding (the same "unresolved_cta" defect class recurs across both threads)
- **verify-later:** n/a — dated handoff, check RUNBOOK_site_quality.md for current state

### CQ-010 — Placeholder-content suppression sweep
- **status:** deployed
- **status-evidence:** Executed SQL in 018: find deployed sections containing 'NEEDS HUMAN REVIEW'/'Lorem ipsum'/'[INSERT'/'<no value>', replace with hidden comment, create per-page placeholder_content items (handler 'human-review', status needs_human_review) plus per-site needs_rerender items.
- **what:** A validation pattern: placeholder or unreviewed text must never stay live — offending sections are hidden behind an HTML comment, a needs_human_review work item requests the real data (team names, photos...), and a rerender item republishes. Companion flows later resolve needs_section_data items as wont_fix when data arrives via site_specs (team, departments) or the section is dropped (pricing → engagement process).
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#placeholder-sweep and #075b-075e
- **relations:** work-item queue; site_specs identity enrichment; hitl approval
- **verify-later:** validation agent producing these; recurrence of placeholder text

### CQ-011 — Audited content pipeline (persona → research → draft → veracity/copyright audits)
- **status:** aspirational
- **status-evidence:** "Content Pipeline cannot be a single agent… Step 4 (Audit - Veracity)… Step 5 (Audit - Copyright)" (001); "Purifier Agent" and "Copywriters with Character" in the phase summary (014); site_persona step defined in 011.
- **what:** Content generation as an orchestrated sub-system: define a site persona/style guide, research via search/scrape adapters, persona-driven drafting, fact-check against research (separate agent, possible HITL), plagiarism/copyright audit (images only from licensed/free sources), then inject into template slots found by parsing data-function attributes. Motivated by veracity/copyright being "mission-critical legal and reputational risks".
- **sources:** docs004_website_capture_project/website_analysis/README.001.prompt_of_objectives.md#content-bottleneck; docs004_website_capture_project/website_analysis/README.011.mvp_content_generation_workflow.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** live successors: content-quality docs, research-agents; persona idea → persona architecture across the platform
- **verify-later:** whether any veracity/plagiarism audit step exists in the current content pipeline

### CQ-012 — Prompt composition asymmetry (text cascade vs image)
- **status:** aspirational
- **status-evidence:** loop_closure(9) decisions: "Step 5 image prompt cascade — Defer. Keep the single-prepend `imagery_direction` cascade for v1"; references `FOCUS_prompt_composition_pattern.md` — "the text pattern itself is fragile and shouldn't be copied wholesale … a composer step that produces a parameter envelope (for both text and images) is the strongest candidate."
- **what:** Deliberate design opinion that image prompts use only a single-prepend `imagery_direction` cascade, not the richer page-content-writer text composition — because the text cascade is considered fragile and a better target is a unified composer producing a parameter envelope for both text and images (likely landing in 2H, not a step-5 extension).
- **sources:** imagery/old/PLAN_imagery_loop_closure(9).md#decisions-taken, #image-prompt-cascade-deferred
- **relations:** live FOCUS_prompt_composition_pattern.md; image request shape; directive cascade
- **verify-later:** composeImagePromptWithDirection; getImageryDirectionForSite

### CQ-013 — Input sanitisation (sanitizeValue, Cc/Cf stripping, NFD survives)
- **status:** deployed
- **status-evidence:** running_notes 2026-06-12 "Sanitisation v2 … now strips Cc AND Cf … Real bug found by the new tests: checking IsControl before IsSpace silently JOINED words".
- **what:** The engine's `sanitizeValue()` strips control (Cc) and format (Cf: zero-widths, bidi overrides incl. U+202E, BOM, soft hyphen) chars, collapses whitespace runs (IsSpace checked FIRST to avoid joining words like `gmt\t\tmaster`→`gmtmaster`), and caps by RUNES not bytes (MaxValueLen semantic changed). NFD combining marks deliberately survive; NFC normalisation + lowercasing are deferred to the P4 collector (needs x/text; engine stays stdlib-only). (Same underlying mechanism is also described independently under traffic-analytics — see TRF-007.)
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-retention-timer, traffic_probe_running_notes(27).md#2026-06-12-debug-guide, traffic_probe_plan(11).md#p4
- **relations:** pairs with P4 ingest validation contract (NFC there); capture-side input sanitisation (traffic-analytics TRF-007)
- **verify-later:** service.go sanitizeValue; MAX_VALUE_LEN handling

### CQ-014 — `component-template-fixer` CTA-handling reuse assumption — corrected, replaced by dedicated agent
- **status:** superseded
- **status-evidence:** Archived `FOCUS_content_quality.md` (2026-06-09): "The plan notes `component-template-fixer` 'already handles CTA fixes' — verify and extend rather than build new." Live `FOCUS_content_quality(2).md` (2026-06-10): "`component-template-fixer` exists but **explicitly punts on CTAs** (`cta_improvement`/`cta` → `fixed:false, action:'needs_review'`)... So the PLAN's 'already handles CTA fixes' was wrong; there was no CTA resolver to reuse — hence the dedicated `internal-link-resolver` (Step 3)."
- **what:** `PLAN_design-note-recommendation-specialists.md` asserted `component-template-fixer` already had CTA-fix handling that could be reused/extended for the hero-CTA phantom-link defect. Verification against the live agent's actual routing table found it explicitly declines CTA improvements, routing them to `needs_review` instead of fixing them. This wrong assumption, once corrected, directly motivated building a new dedicated agent (`internal-link-resolver`) rather than extending the wrong one.
- **sources:** content_quality_and_internal_linking/FOCUS_content_quality.md (archived); live FOCUS_content_quality(2).md; running_notes_17(16)
- **relations:** internal-link-resolver agent; identity-advisor/sites.approval_mode (CQ-015); content-quality defect catalogue (CQ-001)
- **verify-later:** `component-template-fixer`'s current action set

### CQ-015 — `identity-advisor` agent and `sites.approval_mode` gate — proposed, confirmed never built
- **status:** abandoned
- **status-evidence:** Archived `FOCUS_content_quality.md` (2026-06-09) lists them as PROPOSED pieces needing verification ("Before relying on `identity-advisor` / `component-template-fixer` / `sites.approval_mode`, confirm each exists"). Live `FOCUS_content_quality(2).md`/`FOCUS_internal_linking(1).md` (2026-06-10) confirm: "`identity-advisor` does **not** exist. `sites.approval_mode` does **not** exist. The three-way `finding_type` classification and those specialists are PROPOSED, not built."
- **what:** `PLAN_design-note-recommendation-specialists.md`'s three-way finding-routing design (bug / gap / recommendation) named `identity-advisor` as the specialist for contact/email findings and `sites.approval_mode` as the gate for whether recommendation-type findings auto-apply. Neither was ever implemented — a clean case of a documented plan whose specific pieces were checked against the live schema/agent_definitions and found absent.
- **sources:** content_quality_and_internal_linking/FOCUS_content_quality.md (archived) and live FOCUS_content_quality(2).md; running_notes_17(16)
- **relations:** component-template-fixer CTA-reuse assumption (CQ-014); recommendation-specialist architecture (content-governance CGV-026); two sources of truth for contact email (CGV-006)
- **verify-later:** re-check `agent_definitions` and the `sites` table for these names in case they were built later

### CQ-016 — LLM fabrication classes in self-built site content
- **status:** deployed
- **status-evidence:** AUDIT U1-U11 with removal dates ("Section DELETED 2026-07-10", "FAQ replaced 2026-07-10"); "Fabrication sweep, 2026-07-10 … CLEAN".
- **what:** Catalogue of what unconstrained content agents invented on a live site: fictional staff ("Peter Grenfell"), a nonexistent "8 departments" taxonomy, platform subsystems dressed as client case studies, AI agents listed as human team members with 404 portraits, capabilities that don't exist (Playwright scraping, proxy pools, circuit breakers, Helm/IAM), and misaligned stat suffixes ("99.9x uptime"). Removal required both spec rewrites and component deletion because some copy was baked into rendered_html.
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#2; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-5, #Turn-11; docs/leopardessconsulting/scripts/L5_pages.sql (header)
- **relations:** claim-evidence audit rule; site_specs pinned gap (content-governance CGV-028, fabrications regenerate while specs are wrong); anti-hype voice spec (CQ-017)
- **verify-later:** page_components.content_data pattern sweep on all sites; content-gap-planner rewrite history in site_specs

### CQ-017 — Anti-hype voice and claim-discipline spec
- **status:** deployed
- **status-evidence:** HANDOFF §3 "Specs — identity/voice/design_intent/portfolio rewritten (source_agent operator-rebuild, pinned)".
- **what:** A reusable voice contract for LLM-written site copy: positive framing (no strawmen, no competitor swipes), prefer the smaller exactly-true claim, plain language over compression, a banned-language list ("digital transformation", "leverage", "seamless"…), an LLM-tells-to-avoid list (reflexive triads, em-dash rhythm, summarising flourishes), CTA governance (name the next thing that happens; vary per page — repetition "signals content shallowness"), and honest uncertainty framing ("we have not done that one yet").
- **sources:** docs/leopardessconsulting/specs/voice.json; docs/leopardessconsulting/specs/identity.json#content_posture; docs/leopardessconsulting/scripts/L5_nav_and_ctas.sql (header)
- **relations:** LLM fabrication classes (CQ-016); portfolio honest-labelling pattern ("Not yet done for a client")
- **verify-later:** site_specs aspect 'voice' for leopardess; whether content writers consume voice spec

### CQ-018 — Cross-page content-duplication checker, split by what each half may be confident about
- **status:** built, **council-APPROVED**, INERT (no discovery agent's config names the check yet — verified 2026-07-31 across all five workflow-bearing columns of `agent_definitions`: zero references)
- **status-evidence:** commits `feat(151 cand 3)` + **`43492ec94` (identity fix)** 2026-07-31; `deduplicate-sections` seeded and verified live (`269_deduplicate_sections_handler.sql`); **13** behavioural tests pass (against a clean `git archive HEAD`); council rounds 1 and 2 REVISE, **round 3 APPROVED** 2026-07-31 15:39Z on `da3f2d9b-ae6f-492d-ad3b-748323b66367` (12 approve / 2 advisory object, none high-severity; the round-2 gating seat `prior_art_librarian` approved). Round 3's first attempt was killed mid-flight by another session's chassis roll and re-fired unchanged. Shipped rule measured over all 1,023 live `page_components` rows post-fix: **0 in-remit groups, 0 deletions**.
- **what:** The permanent form of `bugs_open/151`'s hand-run census ("is this site saying the same things all over itself?"), plus its deterministic repair. `check_content_duplication` splits its population per `remit.go`: **in-remit** is same-page content-IDENTICAL sections, which get a dispatchable `content_duplication` item routed at the new `deduplicate-sections` agent → `remove_duplicate_page_sections` action (delete the later rows, keep the earliest, renumber positions, queue an assemble-only rerender as a SEPARATE work item so a failed render cannot leave the DB correct and the site doubled); **residue** is near-duplicate and cross-page copy, which gets exactly ONE `capability_gap` per site carrying `do_not_auto_rewrite: true` and naming 151 candidate 1 as the structural fix. A verifier is registered for `content_duplication` (re-runs the predicate for the page; a vanished page is an error, never a success). Shared helpers in `datahelpers`, all three of which any comparison over section content should CALL rather than reimplement: `NormaliseSectionText` (prose, for the similarity SCREEN and the residue), `SectionTokenSet` (the Jaccard token set), and **`SectionIdentityKey(slotName, canonicalBlob)` — the in-remit identity: same slot + byte-identical `content_data::text`.** The two are deliberately different rulers: prose for screening, bytes for a DELETE. Ad-hoc/by-hand version: `gauntlet_dead_cta/scripts/dedup_census.py` (fact census + difflib + exact blocks, any site).
- **sources:** platform/orchestration/actions/discovery_checks/check_content_duplication.go, platform/orchestration/actions/remove_duplicate_page_sections_action.go, platform/orchestration/datahelpers/section_text.go, docs/agent_docs/sql_for_agents/269_deduplicate_sections_handler.sql, bugs_open/151 (candidate 3), bugs_open/156, gauntlet_dead_cta/SUMMARY_2026-07-31
- **relations:** implements `bugs_open/151` candidate 3 (candidate 1, the structural fix, remains with the brochure_component_library lane — CONTRIB_2026-07-31 in their directory); `remit.go` in-remit ∪ residue doctrine; the claims gate (`bugs_open/149` C1) is a prerequisite for any future rewrite handler; `bugs_open/156` is the duplicate-ROW defect this repairs, distinct from 151's independently-generated near-duplicates
- **verify-later:** whether the fact-census floor of 6 is the right number; whether any site other than vonc has content-identical rows once the check is enabled fleet-wide; **whether a pre-delete guard should consult `site_plan_sections`** (the system of record) and refuse where the PLAN itself specifies the repetition — measured 2026-07-31: exactly one upstream duplicate exists fleet-wide (`webdesign.co.uk`/`index`/`info-card-grid`, 2 plan entries) and it does not overlap any in-remit item, so the mechanism `debug_historian` named is real but unexercised. Raised as the top residual risk in council round 2.
- **landmines:**
  - **`content_data` often holds the SITE-WIDE BOILERPLATE, not section content — so prose identity can collapse two unrelated components.** Measured 2026-07-31: `vonc.com/index.html`'s `provocation-card` and `lobby-grid` rows (different `component_id`) both carry the byte-identical site-context blob (`year`/`email`/`domain`/`nav_items`/`_built_at`) and no section content, and the original prose-identity rule would have **DELETED the lobby-grid row from a live home page** — the only in-remit group it found fleet-wide. Fixed in `43492ec94`: slot equality is now NECESSARY *and* identity is the canonical blob. The asset-key filter does NOT save this (it strips `url`/`id`/`class`, not `email`/`year`/`domain`/nav labels). See `LANDMINES.md` and `gauntlet_dead_cta/RUNBOOK` §16b — and note this coexists with the next landmine without contradicting it: slot equality as *sufficient* breaks 10 real pages, as *necessary* it strictly shrinks the set.
  - **Measure what this rule WILL DO by compiling the shipped function, never by reimplementing it in SQL.** A SQL/Python restatement is a second definition of "identical" — exactly the drift `section_text.go` exists to prevent. An `md5(content_data)` census undercounts the prose rule by construction and produced a false "verified no-op" (`WRONG_CALLS.md` 2026-07-31). Recipe: `RUNBOOK` §16b.
  - **The discriminator is CONTENT IDENTITY, never slot name.** Fleet-wide 17 duplicate `(page_id, slot_name)` groups exist and **11 are legitimate** (repeated slots, differing content, five sites). A unique index or a slot-keyed rule deletes real sections. A test asserts this false positive stays unflagged — keep it.
  - **A clean fact census is indistinguishable from a clean site.** With 4 approved facts (vonc) no section can reach a 3-fact overlap, so the fact half is blind. The check reports `fact_census_blind` below a pool of 6 rather than returning quietly clean.
  - **The similarity number never causes an edit.** It sizes the residue only. 151 measured two sections asserting identical facts at 18% similarity, so a threshold alone reports them clean.
  - **Enabling order is not optional:** chassis image carrying `remove_duplicate_page_sections` → pod-grep with a control → THEN add `content_duplication` to a discovery agent's check list. The agent was seeded BEFORE `knownHandlerAgents` gained it, so that ratchet states something true rather than aspirational.
