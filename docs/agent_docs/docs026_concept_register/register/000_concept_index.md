# Concept Index — master register

1,685 concepts across 109 category register files (**recounted 2026-07-31**; the
headline had already drifted 1 behind again by that evening, before PBP-025 was
added — and it moved again minutes later when TL-036 landed, and again when DOC-070
landed that evening: its author watched this very line go 1,683 → 1,684 under them
mid-edit, from another session's row, which is the same
lesson three times in one evening: re-take it from the grep, never increment from the
line you just read — so it is re-taken from the grep below rather than incremented, not
accumulated: the running "+N more" chain below had drifted 13 behind the actual
row count, so the headline is now taken from `grep -c "^| [A-Z]*-[0-9]" ` and the
chain is kept only as a history of *why* entries postdate the freeze).
1,627 consolidated from
2,185 raw extraction blocks (32 extraction-unit files, ~4,111 source documents
under `docs/`) as of 2026-07-13; 4 more (STY-049, FIX-051/052/053) added
2026-07-16 for a subsystem (fixloop's triage/escalation layer, and the
missingkey=zero structural defect it surfaced) that shipped after extraction
froze; 2 more (MDL-038/039) added 2026-07-17 for two live platform bugs the
fix-loop's own first real-case run found — see the addition notes further
down; and 3 more (LCO-005/006, OPP-003) added 2026-07-28; and 1 more (OPP-004) added
2026-07-28 closing `bugs_open/106` — the register's own coverage check, wired to the
commit path so a new subsystem announces its own absence; and 2 more (TL-032/033) added
2026-07-29 for the orphan element-reference gate and the acceptance-ladder eligibility
defect that let 63 broken tools ship unwatched; and 1 more (ADP-017) added 2026-07-30
for the shared reply-delivery policy `bugs_open/133` extracted from the one adapter
path that had it; and 1 more (ADO-037) added 2026-07-30 for verbatim adoption —
`fidelity=locked` plus the `deploy_mode` component key — which made doc 028's
long-inert fidelity dial (ADO-011) read by something for the first time; and 1 more (STY-050) added 2026-07-30 for the per-site
chrome-config seam that finally gave `bugs_open/018`'s schema-driven fill a real
consumer
for capability that shipped this week — a structural fingerprint for LLM
responses, the diagnosability layer on the island's tools-api, and the
pre-commit detector that stops model text reaching a log; and 3 more
(SCR-002/003/004) added 2026-07-28 from the `bugs_open/100`+`101` fix —
fetch-recorded provenance, the declared config-key contract that makes an inert
step-config key detectable, and its fleet-wide coverage report; and 1 more
(WFA-002) added 2026-07-28 from the `bugs_open/124` fix — the `$ctx.`
execution-context parameter namespace, which lets any queue-driven workflow's
SQL record which run claimed a row; and 1 more (DMR-002) added 2026-07-28 —
single-service deploy with a registry pre-flight, built while rolling the
`bugs_open/131` fix because the all-or-nothing `deploy-agents` would have
ImagePullBackOff'd twelve healthy services; and 1 more (LNK-024) added 2026-07-28
from the `bugs_open/079` REOPENING — dead-link repair moved to the persistence
point, because a transformation that lands only in `clean_html` is discarded by
the structured save path and 4 of the 6 live persistence paths had no repair at
all; and 1 more (FIX-054) added 2026-07-28 — the forward-fitness council seat
extended to the FIX lane and the gate, on the owner's reversal of decision D9,
because `bugs_closed/124` and `129` were each vetoed on scope and told to route
their seam to an architecture review reachable from neither lane; and 5 more
(TL-031, PLAN-043/044/045/046) added 2026-07-28 for the experience register — its
substrate, its validating write path, its bind/verify consumer, its five-seat
approval council, and the attribute assertion that closed its largest harness gap;
and 1 more (IMG-065) added 2026-07-29 — the operator asset-amend path
(staging BYTEA → ingest_staged_asset → S3 → in-place assets amend), built from
the `bugs_open/131` og-card finding that the platform had NO path for a human
to supply corrected image bytes; and 1 more (CLC-012) added 2026-07-29 — a second
component implementing an ALREADY-REGISTERED experience pattern
(`teaser-detail-deeplink`) rather than declaring a new shape, which is the property
that makes a shape vocabulary worth keeping: two components that look nothing alike
share one micro-journey; and 1 more (FIX-055) added 2026-07-29 — truncation-gate
attribution on the council report, from `bugs_open/138`, because a reviewer that
ran out of tokens was gating rounds under a label that named the SEAT, making an
advisory seat blocking and a working seat look noisy enough to retire; and 1 more
(WFA-003) added 2026-07-29 from the `bugs_open/144` fix — sub-workflow validation
and the exported step traversal that comes with it, because the runtime validator
and the offline config-key audit had each written their own top-level-only walk,
were therefore blind in the same direction, and agreed with each other over 85
live steps that nothing had ever checked; and 1 more (SCH-023) added 2026-07-29 —
firing a DISABLED scheduled task once at a target you choose, built on owner
instruction to run the `detected`→`triaged` promoter, whose own `pre_query`
selects `ORDER BY updated_at ASC LIMIT 1` and therefore points AWAY from a fresh
finding, because filing the work item is what bumps that column.
That last group carries a note on method: the first of the five was added alone,
and adding it was enough to make `102_CHECK_register_coverage.py` treat the whole
workstream as covered while four callable mechanisms were still absent. The
coverage check matches on workstream NAME, so ONE entry silences it for
everything — which means dropping a ratchet line obliges you to register the
mechanisms, not merely one of them. Status tags were
documentary signals from the source material unless independently verified (see
below).

**2026-07-30 addition:** 1 more (**WFA-004**) added closing `bugs_open/148` —
an offline detector, `config-key-audit --unregistered-actions`, that replays
the exact "unregistered action, no topic" rule the runtime validator enforces
against every live agent definition, reusing WFA-003's `WalkSteps` traversal
rather than writing a fourth hand-rolled walk. Measured against the live fleet
(178 agents): exactly the 3 findings the bug documents.

**2026-07-31 addition:** 1 more (**PUB-004**) — round publication on the island's
`tools-api`: two endpoints and two columns that let a completed gauntlet round
become a public, linkable record, but **only because the visitor who argued it
pressed share**. Registered in the same commit that ships the seam, per the
2026-07-28 ordering ruling. It is tools-api's first public read surface, so the
entry states the guarantee it adds rather than only the keys: one row class is
now readable by anyone holding its unguessable slug, gated on `published_at IS
NOT NULL` in the read query itself. Two things there deliberately for the next
reader: the public projection is its **own type** so a new column cannot silently
widen what is served, and the slug alphabet/validator live in **one** place after
briefly living in two — a drift that fails silently and totally.

And 2 more (**PUB-002, PUB-003**) added 2026-07-29 by the consolidation
programme, for the two shared packages it built on 07-28 — `platform/httpguard`
(one client key, one banded limiter, one intake gate) and `platform/mailer` (the
first SMTP sender inside the built code). Both were council-approved on the day
they were built and **neither was registered then**, which is exactly the gap
`bugs_open/106`'s coverage check exists to catch. Registered late rather than not
at all, and both carry the honest status: **built, approved, and called by
nothing.** PUB-002 also records the seam added on 07-29 — `ClientIP` now requires
the caller to name the proxy in front of it, because its previous hard-coded
rules were nginx's and are false on the estate's other front-end.

**Stage 2 (code/DB verification) ran 2026-07-14 and is COMPLETE** — see
`006_VERIFICATION_stage2.md` for method and full findings. Every one of the
original 1,627 concepts was checked against the live codebase at least once,
across three batches: all 314 partial/unknown, all 871 deployed (a false-positive
sweep added after batch 1 found the riskiest bucket wasn't the one originally
planned), and all 174 superseded/abandoned. 124 corrections were confirmed
total (each independently adversarially re-checked before acceptance) and
applied below — a ~7.6% error rate. A 7th status, **convention**, was added
for design doctrines/methodologies that were tagged `deployed` but are not
code artifacts at all (see status vocabulary in `README.md`). One duplicate
(PUB-001) was retired to a pointer entry rather than a status change. Status
column below now reflects verified ground truth for the 124 corrected rows;
all other rows held up under verification and keep their original signal.

**2026-07-27 addition (third instance of the same gap — see `bugs_open/106`):**
the entire **claims-verification** subsystem was absent. Its first plan is dated
2026-07-16, three days after extraction froze, so V0–V5, the `evidence_base`
register, the banned-claim scanner and the citation verifier were never
extracted; `grep -rl evidence_base register/` returned nothing until today. Added
as `claims-verification.md`, **18 concepts (CLM-001..018)**, grounded in code and
DB read first-hand on 2026-07-27 rather than carried from other documents.

Two of the twelve are recorded defects rather than descriptions — `CLM-009`
(`EvidenceFact.Kind` declared and read nowhere) and `CLM-010` (the fleet-share
deferral whose "until two sites" precondition lapsed at eight) — because both are
dormant capability, which is exactly the class this register exists to make
findable and the class code search cannot surface.

**The pattern is now three-for-three and all three were found by coincidence**
(fixloop 07-16, model-directory 07-17, claims-verification 07-27), each by a
session that happened to be working beside the hole. Measured today: **51 of 76
workstream directories postdate the freeze — 67%.** `bugs_open/106` proposes a
coverage sensor on the model of `verifier_coverage_test.go`, and a
`covers-through:` stamp per file so a reader can see where the register stopped
looking.

**2026-07-16 addition:** a coordination pass with the fixloop workstream (its
tool went complete 2026-07-16, all 4 triage/escalation phases live) found a
genuine gap — that whole subsystem shipped after extraction froze on 2026-07-13,
so none of it was in the register. Added and independently verified 4 new
concepts, cross-checked against live code (registry entries, exact file/line
citations, and — for FIX-051/052/053 — an independent research pass that also
confirmed every cited commit against `git log`): `FIX-051` (triage router,
Phase 1), `FIX-052` (silent-check verifier, Phase 2), `FIX-053` (feedback
close-out, Phase 3), and `STY-049` (the `missingkey=zero` structural defect the
fixloop's real-case queue surfaced, root-caused via the image-landing/article-body
incident). `FIX-034` (the pre-existing base digest) was updated in place to
record the Phase 4 escalation-section enhancement built on top of it. This is
targeted incremental extraction of genuinely new material, not a re-sweep of
frozen stage-1 corpus.

**2026-07-17 addition:** the fix-loop delivered its first real-case CONFIRMED
diagnosis (correlation `e505f70f`) and, in the same session, an owner-directed
model swap surfaced a second live bug — both genuinely new platform defects,
not previously in the register. Added `MDL-038` (BUG A: `GenerateText` never
decodes `stop_reason`, so max_tokens-truncated LLM responses silently look
complete — CONFIRMED by the loop itself on 3 citations including live
`llm_call_log` state evidence, independently re-confirmed here by direct code
read) and `MDL-039` (BUG B: an agent's root-level `ai_service` config silently
shadows its step-level config — proven by direct experiment on `diagnose-agent`,
17-agent fleet blast radius, terminal state PARTIAL rather than CONFIRMED
since the loop's own two-evidence-family guard correctly withheld CONFIRMED
pending a state-tier citation). Also confirmed by direct re-read: `fix-proposer`
(home of the bug-historian reviewer added 2026-07-16) has no root-level
`ai_service` key, so it is NOT among the 17 affected agents.

Sorted by register file, then by ID. Use your editor's search for a concept name,
an ID prefix, or a status word.

| ID | Concept | Status | Summary | Register file |
|---|---|---|---|---|
| DES-023 | Layout: technical-precise | deployed | "Engineered" layout — glass-effect header (backdrop-filter blur) as its signature moment, tight border-radius,... | design-composition.md |
| CTS-011 | CSS colour inheritance model (--section-*, --color-* fallback) | deployed | "Single most important rule" — element colours resolve via two-level var() fallback chain | contracts-and-standards.md |
| DBI-013 | QueryDatabaseAction parameterised queries & schema-drift discipline | deployed | $1 placeholders required; live DB, not dumps, is the source of truth for columns | database-and-infrastructure.md |
| PAY-006 | One-off credit vs recurring subscription billing model | aspirational | $5 build credit ledger vs tier subscription; billing_credits table proposed, not built | payments.md |
| DBG-017 | sites.status is informational; never scope blast-radius by status='active' | deployed | 'active' is legacy; dispatch keys on site_work_items instead | debugging.md |
| DEV-026 | F6: dedup status-list mismatch and itemsCreated overcount (open defect) | aspirational | 'unresolved' status disagreement between Go guard and idx_swi_dedup; parked, unfixed. | development-guide.md |
| DBG-073 | Workflow monitoring REST endpoints (built but apparently unused) | unknown | /monitor/* endpoints documented as built but never seen used; psql/db-inspector instead | debugging.md |
| DOC-059 | Debugging-guide fork-and-merge maintenance (cumulative 016b copy) | deployed | 016b guide forks across chat threads, periodically merged back | documentation-system.md |
| DOC-003 | Per-tool travelling documentation convention (PLAN_/NOTES_ files + taxonomy) | partial | 037 convention: PLAN_/NOTES_ files per tool, instantiated across 3 project trees | documentation-system.md |
| STY-021 | R6f — theming vocabulary drift (defined vs consumed CSS custom properties) | deployed | 11 gap names between template var() usage and generated styles.css :root | styling-render-pipeline.md |
| SPEC-015 | Intake orchestrator workflow (classify -> brief -> spawn builder), legacy | partial | 11-step HITL orchestration ancestor of the work-item relay | site-spec-and-classifier.md |
| IMP-018 | Corrupted component templates and the quality→regeneration bridge | deployed | 14 components fleet-wide had html_template saved as RENDERED OUTPUT (literal `<no value>`, zero `{{…}}` vars) —... | improvement-loop.md |
| MDL-012 | thunder-reaper scheduled task + uptime deadline | deployed | 15-min task decommissions instances past max_uptime_hours; deadline is ours | model-infrastructure.md |
| SYS-031 | collected_data growth causing OOM-kills | partial | 18MB collected_data blew the 512Mi pod limit, causing phantom-completed orchestrations | system-architecture.md |
| DEV-063 | Role-based agent pools / atomic work-claim queue (proposal, superseded) | superseded | 2025 proposal for role-queue atomic work-claiming; ancestor of today's work-item pipeline. | development-guide.md |
| DEV-055 | v1 monolithic LLM-chain site builders | superseded | 2025-era chain of one-LLM-call specialists producing free-form HTML; no component library. | development-guide.md |
| IMG-003 | Imagery best-in-class programme (I0–I8) | partial | 2026-07-08 successor programme; I0/I1 complete, I2 in progress, I3-I8 not started. | imagery.md |
| FIX-034 | fixloop-digest / awareness surface | deployed | 24h activity digest built, awaiting chassis image | fix-loop.md |
| ADM-006 | WireGuard VPN admin-access implementation detail | superseded | 3 access options (WireGuard-in-cluster, VM bastion, port-forward); configs dropped from live doc | admin-dashboard-and-api.md |
| CHAT-004 | site-chat-installer orchestration (install_chat maintenance task) | aspirational | 3 sub-agents: chat-context-builder, chat-widget-installer, chat-route-registrar | site-chatbot.md |
| BIZ-023 | Domain content strategy framework (15-question) | aspirational | 3-layer, 15-question method for deciding what content a domain needs | business-strategy.md |
| VET-001 | Vet med pricing pipeline (discovery → scrape+evidence → export) | deployed | 3-stage business-intel pod pipeline: URL discovery, price collection+evidence, JSON export | vet-med-pricing.md |
| BLD-003 | Build pump and the queue immune system | deployed | 30s scheduled trigger plus self-healing reapers/watchdogs keep the relay moving | build-pipeline.md |
| SOC-010 | Motivation hierarchy and designed user journey | aspirational | 4 motivation tiers mapped to a staged first-5-seconds-to-month-6 user journey | social-media.md |
| VET-007 | Vet med pricing schema (products / retailers / listings / snapshots) | deployed | 4 tables + matview: med_products, med_retailers, med_retailer_listings, snapshots | vet-med-pricing.md |
| BIZ-022 | Verticals designed (revenue models + knowledge clusters) | aspirational | 5 verticals with monetisation + 24-month revenue projections, no live revenue yet | business-strategy.md |
| CH-001 | Companies House enrichment pipeline (bulk collect → match → review → detail fetch) | deployed | 5-stage ch-* agent chain; 634/5,780 (23.2%) confirmed matches as of Mar 2026 | companies-house-enrichment.md |
| DBG-052 | "Renders empty" diagnostic method (data-binding, not template) | deployed | 5-step method proving an empty shell is a data-binding bug, not a template failure | debugging.md |
| VONC-005 | lobby-grid arena component (six-room grid) | deployed | 6-card Arena grid runtime-filled from `arena`; reference loader-builder implementation | vonc.md |
| SOC-007 | Provocation engine — layered content production architecture | aspirational | 6-layer pipeline: raw feed → framing → curation → mashup → serialisation → niche | social-media.md |
| NEWS-001 | News feed pipeline (sources -> async ingest -> triage -> JSON render -> commit) | deployed | 6h heartbeat -> orchestrator -> ingesters -> triage -> render_news_section -> git commit | news-feed-pipeline.md |
| IDEA-002 | Operator-risk column: hazard scored separately from fitness, with gates | deployed | 6th scored dimension for consequence-of-being-wrong; paused the SFI assessment | idea-product.md |
| CASE-012 | Risk-as-hazard scoring dimension | deployed | 6th scoring factor for consequence-of-being-wrong, kept separate from fitness | site-case-studies.md |
| CANB-001 | Canine biology knowledge tree (1M-agent demo) | aspirational | 7-level, ~1M-agent swarm building a citable Labrador knowledge tree; demoted to showcase | canine-biology.md |
| FTW-012 | Base-model decision: Llama 3.3 70B QLoRA | deployed | 70B chosen for available hardware; 8B ablation planned but never run | finetuning-flywheel.md |
| SOC-005 | Behavioural archetype system + Daily Gauntlet | partial | 8 archetypes earned via a 5-provocation daily quiz; archetype hub live 2026-07-12 | social-media.md |
| STY-006 | Three-part styles.css assembly and core/specialised palette merge rules | deployed | 8 core slots spec-wins, specialised slots theme-wins; caused mixed output on leopardess | styling-render-pipeline.md |
| VONC-010 | Archetype hub built with existing machinery (entity pages + query-resolved grid) | deployed | 8 entity pages + query-resolved grid; fixed a zero-archetypes page_type bug | vonc.md |
| VONC-009 | vonc.com Spark v1 site (the live testbed) | deployed | 8-page v1 build: index, archive, about, contact, archetype hub, tools | vonc.md |
| DBG-065 | Mode A/Mode B broken-template taxonomy + pre-extraction JS-shell class | deployed | <no value> repair vs regeneration routing; pre-separateInlineJS shells never got JS | debugging.md |
| DES-066 | Template helper system (`{{palette}}`/`{{typo}}`/`{{token}}` with mandatory fallback) | deployed | A Go-template-style substitution convention embedded directly in the `css_template` CSS text: `{{palette "key"... | design-composition.md |
| TLIB-023 | intent-probe capture component | deployed | A NEW content-library section (built after a survey found nothing reusable among 83 existing sections) rendering... | tool-library.md |
| DES-002 | Style collections (data bundle + design bridge ancestry) | deployed | A `style_collections` row bundles the components and tokens defining a site's visual identity:... | design-composition.md |
| DES-050 | Library-row cleanup pattern for failed composition cascades | deployed | A bad composition cascade leaves one set of library rows... | design-composition.md |
| TL-034 | `has_visible_area` — the Tier-4 check that separates "exists" from "usable" | built, mismeasuring until the next roll | Measures the rendered box; three tools served work areas of 1146x0 that selector_exists passed... **bugs_open/157: every WHOLE-NUMBER axis measured 0 and failed as "too small to click" in all images ≤ v1.0.1215; fixed at HEAD 71680ad513** | tool-lifecycle.md |
| TL-035 | `capture_renders` — a screenshot of a page that PASSED | built | Opt-in; the only screenshot path fired exclusively on failure, so three defects on all-passing pages were found by a human looking... | tool-lifecycle.md |
| TL-032 | Orphan element-reference detection (`OrphanElementRefs` + check `orphan_element_refs`) | built | A static detector for a page whose own JavaScript addresses element ids the page never contains or creates, so... | tool-lifecycle.md |
| TL-033 | Ported tools were invisible to the whole acceptance ladder (eligibility + subject-key widening) | deployed | Every tier opened with `cc.component_level='tool'`, so 63 ported tools in one shared 'section' row were never... | tool-lifecycle.md |
| TL-013 | Tier-2 static acceptance checker + the anchor rule (discovery check `tool_acceptance`) | deployed | A browserless discovery check (sibling of tool_health): loads the current travelling PLAN's criteria fence,... | tool-lifecycle.md |
| CTXA-023 | "Verified against ALL forks" claim did not actually reconcile | partial | A canonical debugging-guide consolidation missed a parallel fork that kept diverging independently | context-assembly.md |
| ONB-010 | Convention coverage IS capability reliability | aspirational | A capability is only as reliable as its manual convention's audit coverage | onboarding-config.md |
| DEV-044 | needs_section_data review items appearing on successful builds (open question) | unknown | A clean isolated build still spawned an unexplained needs_human_review section-data work item. | development-guide.md |
| DBG-025 | CLI/ops data-transfer pitfalls (COPY/psql jsonb, kubectl exec/cp, tnr scp) | deployed | A cluster of verified transfer traps beyond the kcat heredoc issue | debugging.md |
| TL-030 | Mode-B rendered-artifact templates (components stored as rendered output) | deployed | A component corruption class: html_template full of bare `<no value>`, zero `{{.}}` slots, empty input_schema —... | tool-lifecycle.md |
| DES-075 | Post-025 CSS content flow: empty css_content is by design, not a bug | deployed | A debugging-relevant clarification of the deployed pipeline (DES-003/DES-071): the design pipeline runs... | design-composition.md |
| TL-004 | Post-adoption tool-recreation detection check (T2) + tool-recreation-handler workflow | deployed | A discovery_checks package check finds page_type IN ('tool','game'), status='active' pages with no widget,... | tool-lifecycle.md |
| IMP-021 | backend_unreachable discovery check | partial | A discovery_checks/ check giving the improvement loop eyes on the VM-hosted class: per-site, NOOPs unless... | improvement-loop.md |
| REB-002 | Assemble-only vs section re-render distinction (the `reason` gate) | deployed | A dropped reason anywhere in the chain silently downgrades a fix to assemble-only | rebuild-cascade.md |
| TLIB-005 | system-stats component key-contract break (regen renames fields, dependents empty) | deployed | A durable cross-site platform bug found via gamesdesign's empty stats band: component-creator regenerated the... | tool-library.md |
| IMP-042 | build-dispatch-loop self-chaining removal | deployed | A fix (migration 063) removing the build-dispatch-loop's self-respawn pattern (spawn_next_dispatch →... | improvement-loop.md |
| TP-006 | Reuse-checking retrieval architecture | partial | A framing that reuse-checking (for code or components) should be almost entirely algorithmic: a maintained... | tool-pipeline.md |
| IMP-014 | Fixes-land-at-initial-render principle (loop fixers backstop only) | deployed | A governing platform principle carried into every artefact of the scheme-to-components work: correctness must... | improvement-loop.md |
| DEV-036 | psql read-only PreToolUse gate | deployed | A hook auto-approves read-only SELECT/\d psql while mutations still prompt the human. | development-guide.md |
| DEV-030 | Correct-while-touching norm (bounded repair of adjacent inert bugs) | convention | A migration touching a workflow also fixes known-inert bugs in that same workflow, declared explicitly. | development-guide.md |
| DES-078 | Spec ownership / silent-override failure-mode principle (doc 028) | deployed | A named failure-mode taxonomy governing all spec-writing agents in the design-composition pipeline: an agent... | design-composition.md |
| DIAG-005 | SeenRequests / data_request progress-counting mechanism | deployed | A new unseen read-only data_request counts as loop progress so guards don't stop one iteration too early | diagnosis-loop.md |
| TL-002 | Adoption interactivity misroute (canonical prefix desync, "M2") | deployed | A page_type='tool' page rendering description-only prose but no widget can have two distinct causes needing... | tool-lifecycle.md |
| IMP-033 | Trust ledger + gate-policy engine | aspirational | A per-capability store of automation level, gate policy, and supporting evidence, plus a small evaluator mapping... | improvement-loop.md |
| IMP-032 | Reliability cascade (reuse → generate+verify → compete+judge → HITL) | aspirational | A per-task router for producing any unit of work in descending reliability order: known-good reuse, then... | improvement-loop.md |
| IMP-020 | Dormant discovery check: checkEmptyPageSections / validate_component_standards wrapper | abandoned | A pre-existing sub-check (`checkEmptyPageSections`, inside... | improvement-loop.md |
| VMB-003 | Probe as Layer 4 build + thin Layer 5 VM deploy (D1-D4) | deployed | A probe is a normal chassis site whose only difference is the deploy target | vm-backend-sites.md |
| IMP-025 | Content-writer chrome double-injection bug and chrome-ownership rule | deployed | A production bug rule: site chrome (header/footer/head) must be injected exactly once, only at the... | improvement-loop.md |
| TL-023 | Fork-divergence detection for library tools (proposed) | aspirational | A proposed zero-cost SQL discovery check comparing a deployed fork's html_template hash against its forked_from... | tool-lifecycle.md |
| DES-068 | Seed-driver transactional load pattern | deployed | A psql driver script wrapping all 15 numbered layout `\ir` includes in a single transaction with `\set... | design-composition.md |
| DEV-085 | workflow field-path audit query (jsonb_path_query over agent_definitions) | unknown | A recursive jsonb_path_query sweep extracting every field-path value referenced across all workflows. | development-guide.md |
| DEV-039 | Development-guide gotcha: max_tokens must live inside ai_service | deployed | A root-level step-config max_tokens is silently ignored; Anthropic client defaults to 2048 tokens. | development-guide.md |
| DEV-040 | Development-guide gotcha: verify deployed contents against the pod, never tag/git | convention | A same-tag deploy can silently ship a stale binary; only the running pod binary is reliable proof. | development-guide.md |
| DES-074 | Parallel legacy HTML-assembly render path (getThemeByID / GetThemeByName) | partial | A second, older render path reads `css_themes.css_content` directly into assembled HTML, independent of the... | design-composition.md |
| CLC-001 | Shared content-component reuse model (one content_components row, N page_components instances) | deployed | A section component is a single shared content_components row (keyed by `function`, with section_type,... | component-lifecycle.md |
| DES-009 | Semantic CSS theme and snippet system (theme_tags, css_themes, css_snippets, js_snippets) — superseded | partial | A semantic tagging vocabulary (mood/style/industry/audience/functional/colour, with related_tags pairing)... | design-composition.md |
| DES-012 | Design pipeline guiding principles (mottos) | unknown | A shared decision-shorthand invoked repeatedly to settle scope questions across the design-composition work:... | design-composition.md |
| DEV-082 | item_key canonicalization (workItemKey builder) | partial | A shared workItemKey builder to fix item_key prefix drift across work-item creators; code prepared, not applied. | development-guide.md |
| DEV-032 | The seam rule — every prompt consuming a spec field must render it | deployed | A spec field surviving analysis is still ignored at generation if the generation prompt never renders it. | development-guide.md |
| TL-019 | Behavioural QA loop for tools & games (planned Tier 3+ headless-browser testing) | aspirational | A standalone, slower-cadence QA loop motivated by real defects (Jelly Invaders degrading over time, P2P host... | tool-lifecycle.md |
| TLIB-013 | Known-good solution library (proposed) | aspirational | A store of proven solutions as reusable parameterised templates, indexed by capability/domain, versioned,... | tool-library.md |
| CTXP-003 | Adoption context pack (skinner-box) as a worked fresh-thread starter | deployed | A structured resume pack demonstrating the pack contract, including its own inherited staleness | context-pack-tooling.md |
| DES-047 | Computed-styles extraction via browser JS injection | aspirational | A supplementary fingerprint step: scrape a homepage with injected JS calling `getComputedStyle()`, write the... | design-composition.md |
| TLIB-018 | system.internal canonical library site | deployed | A synthetic site record that owns library-level work (component regeneration, library maintenance) so the... | tool-library.md |
| PLAN-039 | Plan storage authority: 029 Q1 and the withdrawn table-first alteration | superseded | A table-first fix was tried and reverted same-day pending the still-open Q1 decision | site-plan-and-reconciler.md |
| TL-001 | Tool widget clobber hazard (interactive content silently destroyed by content rebuild) | deployed | A tool/game lives as a section's rendered_html, not a planned section, so any full rebuild (needs_page,... | tool-lifecycle.md |
| WFA-001 | Workflow Builder & Validator (YAML DSL) | abandoned | A validation-first YAML-to-workflow-JSON authoring system; no later doc references it as ever used. | workflow-authoring.md |
| WFA-002 | `$ctx.` execution-context parameter namespace for `query_database` | deployed | Any workflow's SQL can bind the identity of the run executing it ($ctx.correlation_id, …), so a claim step can record which run took a queued row. | workflow-authoring.md |
| WFA-003 | Sub-workflow validation + one shared step traversal | built (inert until roll) | Steps nested in a loop's sub_workflow are validated at last (85 live steps, 18 agents); WalkSteps is exported so audits stop writing their own top-level-only walk. | workflow-authoring.md |
| WFA-004 | Offline unregistered-action detector (`config-key-audit --unregistered-actions`) | deployed (standalone CLI tool) | Reports, before any message arrives, which live steps name an action that is neither locally registered nor topic-routed — closing `bugs_open/148`. Reuses WFA-003's `WalkSteps`. | workflow-authoring.md |
| IMP-044 | Defect-cataloguing discipline (enumerate-before-fixing) | deployed | A working method for a real adoption-run defect sweep: group symptoms into lettered families by shared mechanism... | improvement-loop.md |
| DBG-059 | orchestration_state_audit: temporary attachable trigger for state races | deployed | AFTER UPDATE trigger capturing every transition; explicitly removed after use | debugging.md |
| SOC-002 | Spark — AI game-master social platform (core concept) | partial | AI as producer not performer; opinion-first provocation game; v1 live on vonc.com | social-media.md |
| CTXA-010 | Go analyser + call-graph neighbourhood (and the wiring-include gap) | deployed | AST-based structural index and call-graph neighbourhood; registry.go-style init wiring needs forced -include | context-assembly.md |
| VMB-012 | idea.uk VM deployment: Path A manual setup.sh, systemd binary | deployed | Abandoned Docker/S3 plan for a single embedded-page Go binary on a Hetzner VM | vm-backend-sites.md |
| DEV-038 | Roadmap-phases enforcement gap (routed to builder thread) | deployed | Absent a roadmap, Tier-3 phase constraints vanish rather than degrade gracefully; confirmed in code. | development-guide.md |
| CTXK-015 | cmd/bundle operational usage lore across workstreams | deployed | Accumulated field lessons (registry-based action resolution, -step gotchas, path quoting) from repeated real use | contextkit-toolchain.md |
| DES-007 | Superseded design-agent family split (brand-designer / layout-architect / style-generator) — replaced by composition/execution | superseded | Across at least two archive generations, the plan was to decompose the monolithic `webdesign-agent` (brand... | design-composition.md |
| VMB-006 | setup.sh — idempotent multi-vhost box provisioning | deployed | Adapted from idea.uk's script; idempotent nginx+TLS+hardening provisioner | vm-backend-sites.md |
| STG-003 | Presigned-URL data plane / storage boundary | deployed | Adapter mints URLs; only URLs cross Kafka, bytes go direct VM↔B2 over HTTPS | storage-architecture.md |
| TL-024 | Component quality tracking (0–100 score) | deployed | Additive quality fields on content_components computed by a compute_component_quality action, with indexes for... | tool-lifecycle.md |
| ADO-016 | Interactive fingerprint extraction gap (Path C) | aspirational | Adoption misses script/canvas machinery; tools rebuild as prose | adoption-pipeline.md |
| ADO-006 | Adoption -> classifier handoff: adoption writes first, classifier consumes | deployed | Adoption never calls classifier directly; emits needs_domain_research only | adoption-pipeline.md |
| PLAN-018 | Adoption calls the canonicaliser + reconciler orphan pruning (unbuilt) | aspirational | Adoption's own URL computation diverges from planner canonicaliser; orphan pruning unbuilt | site-plan-and-reconciler.md |
| PLAN-007 | Architectural Tension #2: page identity derived in multiple places that undo each other | partial | Adoption/planner/convergence each re-derive identity with no single canonical owner | site-plan-and-reconciler.md |
| HITL-004 | Decision authority (co-equal voices, abstention, creator veto) | aspirational | Advocate/curator disagreement escalates to human; creator holds veto | hitl.md |
| SYS-002 | Agent message contract & "agent = row" orchestrator convention | deployed | Agent = DB row with default_config.workflow; spawn-before-call; reply to caller's topic; house rules | system-architecture.md |
| DEV-086 | Workflow-as-configuration (JSON workflows in agent definitions) | deployed | Agent behaviour is a JSON workflow (start_step + steps) stored in default_config, not compiled code. | development-guide.md |
| DIAG-023 | analyse_repo_local: the diagnose-agent's self-contained repo fetch | deployed | Agent fetches its own tarball and analyses in-process, pinned to the code_symbols index commit | diagnosis-loop.md |
| SCH-009 | last_completed_at ownership contract and fire_message known-gap | deployed | Agent tasks must explicitly set last_completed_at; scheduler never reads fire_message | scheduler-and-tasks.md |
| VKA-004 | Vertical research handler + knowledge accumulation loop | aspirational | Agent turning research gaps into indexed knowledge benefiting all future domains | vertical-knowledge-architecture.md |
| SYS-037 | Workflow default_config location convention | deployed | Agent workflow lives in default_config, never the separate *_workflow columns | system-architecture.md |
| HITL-002 | Confirm-not-initiate governance/HITL model (decision package) | aspirational | Agent-led reasoning + decision package, human confirms; new version deprecates old | hitl.md |
| SYS-053 | Stateless-first agent principle + DB-backed orchestration state | deployed | Agents are ephemeral executors; all workflow state persists in the database | system-architecture.md |
| IMG-063 | Human taste-gate operating model (runbook rituals) | convention | Agents author/deploy; humans hold credentials, budget sign-off, and visual approval gates. | imagery.md |
| DEV-035 | ExtractActionInputs Strategy-0 explicit dot-paths lesson | partial | Aggressive recursive field search can mis-match; explicit Strategy-0 dot-paths resolve first and win. | development-guide.md |
| DES-024 | Layout: high-energy | deployed | Aggressive, kinetic layout — uppercase headings, 80vh dark hero, diagonal clip-path section separators, zero... | design-composition.md |
| DOC-025 | Framing: plan = enforced desired state; pipeline = compiled runbook; NOTES = reasoning log | convention | Agreed 2026-07-04 framing across site_plans/pipeline/NOTES/contracts | documentation-system.md |
| MDL-014 | Orphan-sweep for stale thunder_instances rows | aspirational | Agreed but unbuilt design to reconcile DB rows against Thunder's live list | model-infrastructure.md |
| NAV-009 | Navigation maintenance: nav-updater and nav-link-fixer | deployed | Algorithmic, no-LLM agents refreshing nav tables and fixing anchor-slug templates | navigation.md |
| DEV-003 | Actions are the unit of work — no wrapper+core split | deployed | All action logic lives in the XxxAction func; no exported "core logic" duplicate API surface. | development-guide.md |
| CTXA-022 | contextkit module packaging and the graduation seam | deployed | All harness tools share two contracts in one module; graduation moves them chassis-side unchanged | context-assembly.md |
| CTS-014 | Query parameterisation contract ($1 + params) | partial | All new SQL must use $1 placeholders; tool-suggester/tool-improver still unmigrated | contracts-and-standards.md |
| IMG-037 | Assets table with full provenance (schema) | deployed | All-binary-assets table with origin_type/prompt/model/asset_key; products side stalled. | imagery.md |
| IMP-034 | Cascade router | aspirational | An action/agent that picks a cascade tier per leaf task from three inputs — the capability's... | improvement-loop.md |
| TL-014 | tool-acceptance-agent — Tier 4 self-driving orchestrator + continuous sweep + iteration loop | partial | An agent (migration 145) closing the loop with zero humans: ensure_site_record → load_docs → request_browser_run... | tool-lifecycle.md |
| DES-010 | Site-design-planner agent (structure × identity × effects) — abandoned proposal | abandoned | An earlier, more ambitious proposal (Option B) for a dedicated site-design-planner decomposing site design into... | design-composition.md |
| TLIB-019 | Semantic component library with vector embeddings (superseded vision) | superseded | An early vision of a Postgres/pgvector library of deconstructed web components: cleaned HTML/CSS with... | tool-library.md |
| IMP-041 | maintenance-triage agent | aspirational | An orchestrator that scans deployed sites for maintenance issues (stale pages, missing pages, broken links, CSS... | improvement-loop.md |
| DIAG-038 | Index hygiene: exclude archived code copies, prune by commit | deployed | Analyser skips archived docs/-copies and (N).go duplicates; prune clears rows from superseded commits | diagnosis-loop.md |
| ADO-026 | site-scraper (Firecrawl scrape -> site_context) | deployed | Ancestor design-transfer mechanism, standardized site_context schema | adoption-pipeline.md |
| DBG-021 | LLM API shape disciplines (server-tool injection, per-model shapes, timeouts) | deployed | Anthropic API wire-shape differences across model generations; long-call timeout sizing | debugging.md |
| SOC-004 | Rooms-not-feeds architecture and the engagement-depth spectrum | aspirational | Anti-feed design: Lobby/Floor/Gallery zones, ephemeral challenges, Moments | social-media.md |
| PBP-006 | needs_llm routing via detectNeedsLLMContent | deployed | Any non-empty input_schema routes a section to LLM generation regardless of render_mode | page-build-pipeline.md |
| AME-001 | entity_state_log — append-only cross-orchestration memory | abandoned | Append-only cross-orchestration memory log with accumulation patterns | agent-memory-and-evolution.md |
| AGOV-003 | Decision log (premise vs rule_trace; inputs_used) | aspirational | Append-only reasoning log for drift detection and audit | autonomy-governance.md |
| CTS-054 | Adapter Response Envelope Contract — conditional traffic-probe application | superseded | Applicability decision demoted once P4 redesigned to need no adapter at all | contracts-and-standards.md |
| CTS-015 | Schema enforcement: flexible vs strict mode | abandoned | Approval-locks-schema design; later shown the strict-mode trigger was stillborn | contracts-and-standards.md |
| CGV-017 | Schema-mode strict/flexible subsystem | abandoned | Approval-snapshot lock regime built, never wired up, dropped 2026-07-09 | content-governance.md |
| HITL-014 | Content-type-aware approval capabilities (text edit / image replace) | aspirational | Approvals adapt UI/edit affordances by content type (text vs image) | hitl.md |
| HITL-022 | Conditional approval branching | aspirational | Approved/rejected outcomes drive finalize-vs-regenerate workflow branching | hitl.md |
| CTXP-001 | Context-pack tooling: analyser/assembler vs package_*.sh trade-off | superseded | Archived guide's discussion of when to use lean call-graph bundling vs broad directory-walk scripts was dropped | context-pack-tooling.md |
| DBI-006 | agent_instances/agent_spawn_history column-shape drift & correction | superseded | Archived manual-DDL column list didn't match spawn_actions.go; corrected twice | database-and-infrastructure.md |
| BIC-001 | Business-intel sweep/verify collection pipeline (vet-intel) | deployed | Area-sweep → collection_tasks → batch-verify pipeline building the vet business directory | business-intel-collection.md |
| DEV-021 | Array-item field contract for the page-content-writer (item_fields fix) | partial | Array fields need per-element shape rendered in the prompt or the model guesses wrong item keys. | development-guide.md |
| DES-063 | Layout CTA-pair curation with WCAG contrast gates | deployed | As part of the section-contrast arc's fix work, `tool-portal-light` gained a missing CTA pair... | design-composition.md |
| IMG-002 | Imagery subsystem pre-plan assessment | superseded | As-is baseline (hardcoded Stability adapter, one-purpose assets) that motivated the loop-closure plan. | imagery.md |
| PBP-019 | sectionHasVisibleContent assembler filter | deployed | Assembler drops sections with ≤10 visible chars; a second silent-drop path for interactive content | page-build-pipeline.md |
| STY-031 | Rerender pipeline (rerender-pages/page-rerender/render-site-components) | deployed | Assembly/deployment half of the system, separated from content generation | styling-render-pipeline.md |
| BIZ-005 | Sale-readiness / separability discipline (idea.uk) | deployed | Assets passed as data, minimal action set, billing behind a provider interface | business-strategy.md |
| DES-017 | Layout: portfolio-kinetic | deployed | Asymmetric, motion-forward, display-type-led layout for creative-studio energy — animated underline text-links... | design-composition.md |
| SYS-030 | claim_work_item atomic claim + load_work_items first_item | deployed | Atomic UPDATE...RETURNING claim prevents double-processing of work items | system-architecture.md |
| RSN-002 | Salience over presence (context bundle) | aspirational | Attention follows the concrete, not mere bundle presence | reasoning.md |
| PAY-004 | Two-plane billing architecture (auth-service truth + chassis cache) | aspirational | Auth service owns billing truth; chassis reads a fed cache table | payments.md |
| PAY-002 | Chassis-wide Stripe billing integration plan (client_entitlements cache) | aspirational | Auth-DB truth + one-way-fed chassis cache; every DDL marked PROPOSED | payments.md |
| DOC-058 | Authored vs derived context (one substrate, change layer between) | aspirational | Authored (owned) vs derived (emitted) context, with a change layer between | documentation-system.md |
| CTXE-001 | Context substrate principles (authored vs derived; salience over presence) | aspirational | Authored sources point at derived artifacts, not copy them; salience management beats bigger context windows | context-engineering-principles.md |
| ABO-003 | context substrate model (authored vs derived) | aspirational | Authored=owned/fallible vs derived=no-owner readout framing | autonomous-build-operate.md |
| DOC-049 | docs019 migration staging script (stage_docs019_migration.sh) | partial | Automates deterministic archive moves; editorial moves stay human-gated | documentation-system.md |
| IMP-031 | Automation ratchet (per-capability trust levels) | aspirational | Automation is not global; each capability (create action, provision nginx, reshard DB) carries its own trust... | improvement-loop.md |
| SYS-040 | Lifecycle map by verifiability + containment (Tier A/B/C) | aspirational | Autonomy ceiling set by verifiability and containment, independent of agent capability | system-architecture.md |
| FIX-035 | Owner standing rule: awareness before autonomy | deployed | Awareness must precede any council/roster widening | fix-loop.md |
| VMB-010 | requires-backend capability gate (Decision 5) | partial | Backend-requiring components gate on deploy-target capability, not site type | vm-backend-sites.md |
| CTS-049 | Capability gate D5 — requires-backend semantic tag | partial | Backend-requiring components gated by class tag; supersedes invented intent-probe site type | contracts-and-standards.md |
| STY-029 | CSS component-list fallback bug (fake 5-item list) | deployed | Bad status filter emptied component list, triggering a hardcoded fallback; fixed | styling-render-pipeline.md |
| DYN-012 | Generation-time guards for dynamic components (archive-list reference build) | deployed | Bakes runtime-fill lessons into generation instead of post-hoc repair | dynamic-applications.md |
| RSN-009 | Mediator as multi-objective optimiser | aspirational | Balance point among authored extremes via priority profile | reasoning.md |
| IMG-049 | Reference-image style anchoring | partial | Banana-native reference-image anchoring shipped via style guide; IP-Adapter/LoRA not built. | imagery.md |
| STY-035 | Inline JS extraction contract (separateInlineJS / js_content) | deployed | Bare script blocks extracted to external per-component JS files at store time | styling-render-pipeline.md |
| DBG-071 | Marker/attribute REPLACE anchoring + hidden-vs-author-CSS landmines | deployed | Bare-string attribute replace corrupts inline querySelector; hidden loses to author CSS | debugging.md |
| AME-002 | Agent variants + snapshot versioning | partial | Base agents versioned/frozen; task variants reference a snapshot version | agent-memory-and-evolution.md |
| IMG-011 | Spawned asset-deployer / storage-env isolation (Phase 2F) | deployed | Base chassis pod carries no IMAGE_BUCKET by design; deploys spawn asset-deployer instead. | imagery.md |
| STY-019 | Visible-content filter (≤10 chars) + data-runtime-fill assembler exemption | deployed | Base filter plus marker exempting intentionally-empty interactive shells | styling-render-pipeline.md |
| CQ-009 | Site-quality programme handoff | partial | Baseline measurement (0 nav/img/svg/script) triggered a dedicated site-quality runbook | content-quality.md |
| DOC-005 | Docubundle context packagers + curated attach-lists (idea.uk) | deployed | Bash packagers assembling go-live and chassis-engine context bundles | documentation-system.md |
| STY-039 | Batched multipage generation (assemble_multipage_site) | partial | Batch-of-3-5 generation strategy, replaced by sequential loop-based generation | styling-render-pipeline.md |
| SYS-005 | Work-item relay spine / dispatch-loop pattern | deployed | Baton = site_work_items row; 30s pump seeds queue; dispatch loop claims and spawns handlers | system-architecture.md |
| DES-033 | webdesign-agent install/render ordering bug ("first render wrong layout") | partial | Before site-design-planner existed, webdesign-agent ran `generate_css → deploy_css → ... | design-composition.md |
| PBP-005 | Render-time item-key reconciler (schema-sourced, non-fatal) | deployed | Belt-and-braces remap of LLM-drifted array item keys onto schema-expected keys | page-build-pipeline.md |
| MDL-028 | Model quality assessment & per-agent model assignment | partial | Benchmark data behind the Claude/Llama70B/Mistral per-agent routing table | model-infrastructure.md |
| FTW-033 | CheckpointUploader trainer callback | deployed | Best-effort checkpoint upload, hard-gated final adapter upload | finetuning-flywheel.md |
| ADM-004 | Work-item HITL model: approve/reject endpoints on pending_review status | superseded | Binary approval gate replaced by needs_human_review + PATCH /specs retry flow | admin-dashboard-and-api.md |
| DBG-032 | Code-ahead-of-DB schema drift (SQLSTATE 42703, latent until first caller) | deployed | Binary referenced columns before migration ran; migration file was mis-parked | debugging.md |
| AGOV-005 | Capabilities catalog: ceiling on the capability | aspirational | Blast-radius-capped trust ceiling, never full autonomy for chassis edits | autonomy-governance.md |
| FIX-007 | Known-answer benchmark methodology | convention | Blind reruns scored against pre-registered rubric | fix-loop.md |
| CQ-002 | validate_page_content gate (pre-deploy content validator) | deployed | Blocker validator (placeholder/contamination/links/email) routing failures to human review | content-quality.md |
| DBI-023 | VM sizing / Hetzner box selection | deployed | Boxes sized by disk/log headroom; EU-only Hetzner, x86-only build caveat | database-and-infrastructure.md |
| CQ-018 | Cross-page content-duplication checker + deterministic handler | built, inert | Same-page identical sections repaired automatically; near-duplicate copy queued as capability_gap, never rewritten | content-quality.md |
| CQ-007 | Adoption content-quality defect families (polish batch) | partial | Brand-suffix titles, empty footer contact, duplicated H1s, empty meta descriptions | content-quality.md |
| CGV-021 | Page-content-writer + admin content brief regeneration flow | partial | Bridge doc: admin edits brief -> Regenerate -> content_rewrite item with brief in prompt | content-governance.md |
| NAV-007 | Hardcoded fallback nav/header defaults inventing structure | partial | Brochure-default fallbacks fabricate URLs; later found not the live-path cause for one incident | navigation.md |
| IMG-009 | asset_key multi-image model (Phase 2B–2D) | deployed | Broke one-asset-per-purpose constraint; enables N heroes/icons per site via asset_key. | imagery.md |
| CASE-004 | robot-hands.com rebuild (testbed case study, 2026-07) | deployed | Broken content layer rebuilt from scratch as the imagery-pipeline acceptance surface | site-case-studies.md |
| WDS-010 | Unified build & maintenance via site_work_items (single queue, same code) | deployed | Build and maintenance are the same process over one queue table | work-dispatch.md |
| SYS-039 | Build-vs-operate asymmetry | aspirational | Build work is sandboxable/competition-safe; operate work is live and risk-gated | system-architecture.md |
| LQT-003 | Verification harness (build + ops) | partial | Build-check/validator/canary/rollback; ops side is the thinnest part | llm-quality-testing.md |
| AGOV-012 | Contributors vs checkers | deployed | Build-path reviewers vs deployed-site monitors, settled distinction | autonomy-governance.md |
| LNK-011 | internal-link-resolver agent | deployed | Build-time sub-agent resolving hero/CTA destinations to real pages, not a patcher | link-management.md |
| BIZ-030 | AI-native orchestration positioning (vs Temporal/Airflow) | abandoned | Build-vs-buy rationale for a purpose-built orchestrator; adapters never built | business-strategy.md |
| CTXK-004 | assembler (cmd/assembler) | deployed | Builds one paste-ready bundle: constitution + task + full in-scope code + signature neighbourhood + schema | contextkit-toolchain.md |
| PLAN-038 | Three section sources for a page build + plan storage triple shape (029 Q1) | superseded | Builds read site_specs blob then pages.sections then sibling synthesis; plan table unread | site-plan-and-reconciler.md |
| CTXK-007 | embed (cmd/embed) | partial | Builds/queries a semantic vector index over symbols; Ollama for real recall, local stand-in for pipeline-proving | contextkit-toolchain.md |
| DIAG-024 | Symptom anchor (F0.4a) | deployed | Bundle always renders the original symptom above the current hypothesis so drift stays visible | diagnosis-loop.md |
| CTXA-009 | Codebase-conditional capabilities (degrade, don't break) | aspirational | Bundle capabilities rest on structural facts that may not hold elsewhere; degrade gracefully, don't break | context-assembly.md |
| CTXA-003 | Three kinds of database data (definition / operational / content) | aspirational | Bundle distinguishes definition rows, operational telemetry, and gated content data | context-assembly.md |
| DIAG-032 | Same-file sibling signatures + fair-share budgeting (F0.4c) | deployed | Bundle lists a scoped symbol's file-siblings too, with fair-share-per-file budgeting fixing starved small files | diagnosis-loop.md |
| STY-028 | site-asset-renderer: deterministic per-site JS snippet bundling | deployed | Bundles js_snippets into assets/js/snippets.js per site, closing the loader gap | styling-render-pipeline.md |
| DIAG-035 | Reasoning-state as a first-class handoff artefact | partial | Bundles lack persistent reasoning state across iterations; partially covered by the evidence trail today | diagnosis-loop.md |
| DIAG-028 | Diagnosis persistence: diagnosis_artifacts write-through + persist_diagnosis_note | deployed | Bundles persisted per iteration; terminal diagnosis written as a NOTES row only when a subject is given | diagnosis-loop.md |
| DEV-074 | Agent/group categorisation taxonomy (category, status, domain_tags) | unknown | CHECK-constrained category/status plus GIN-indexed domain_tags; no doc confirms it was ever applied. | development-guide.md |
| CTXK-003 | analyser (cmd/analyser) | deployed | CLI wrapper emitting a structural-summary JSON of a Go tree, skipping archived/duplicate files | contextkit-toolchain.md |
| DIAG-012 | Tier-coverage guard (F0.4e) | deployed | CONFIRMED verdict must carry both a static AND a state/runtime citation or degrades to Unverifiable | diagnosis-loop.md |
| DIAG-011 | Symptom-coverage gate family (symptom_check → context disposition) | deployed | CONFIRMED verdicts must map every symptom observation to explained/unexplained with citations | diagnosis-loop.md |
| FTW-039 | LLM fallback extraction doubling as training data (med pricing) | deployed | CPU Mistral price-extraction fallback incidentally accrues fine-tune data | finetuning-flywheel.md |
| MDL-003 | Ollama adapter (CPU embeddings + local classification) | deployed | CPU provider for embeddings + small-model classification, same AIService interface | model-infrastructure.md |
| CTS-038 | Call metadata vs response-data convention (output_field.response) | deployed | Call metadata at output_field; called agent's payload at output_field.response | contracts-and-standards.md |
| BIZ-021 | Early portfolio inventory (honest capability notes) | unknown | Candid Feb-2026 snapshot: no site had generated a real lead yet | business-strategy.md |
| FTW-020 | Held-out eval set v1 | deployed | Canonical 20-case (of 50) comparison set reused across all iterations | finetuning-flywheel.md |
| DIAG-025 | gamesdesign resolveResultSpec fixture (reference bug trajectory) | superseded | Canonical REFUTE→REFUTE→CONFIRM eval scenario; retired as yardstick once the site stopped exhibiting it | diagnosis-loop.md |
| ADP-003 | Adapter design pattern / guide (adapter vs agent vs inline) | deployed | Canonical decision rule and Kafka/HTTP microservice structure for adapters | adapters.md |
| TLIB-017 | Seeded standalone tool library | deployed | Canonical interactive tools stored whole in content_components as `<style>+<main>+<script>` with... | tool-library.md |
| CTS-007 | page_type vocabulary and "landing, not index" | deployed | Canonical kebab page_types; homepage TYPE=landing, NAME=index | contracts-and-standards.md |
| HITL-020 | HITL content-approval demo agent and group | deployed | Canonical minimal HITL example: simple-content-writer-with-approval agent/group | hitl.md |
| DBG-001 | The 016/016b debugging guide: assumption checklist + durable invariants | deployed | Canonical, versioned operational runbook + heuristics; extracted piecemeal by nearly every unit | debugging.md |
| ADP-008 | Agent-to-adapter capability maturation path (fast/slow/job lanes) | aspirational | Capabilities prove out as spawned agents, then promote to warm adapters | adapters.md |
| FIX-017 | Revise loop (F2.2) | deployed | Capped repropose cycle on revise decisions | fix-loop.md |
| ADO-031 | Website-builder orchestrator (maximal clone-and-improve vision) | abandoned | Capture->vision->code->synthesis->content->library master workflow never built | adoption-pipeline.md |
| CTXA-021 | DB capabilities capture (\dx/\df into the bundle) | partial | Captures installed extensions/helper functions so generation reuses snapshot_agent instead of hand-rolling | context-assembly.md |
| ONB-014 | Intent-elicitation agent (progressive, value-returning interview) | aspirational | Captures why-chain/priority profiles via proposal-confirmation + elicitation | onboarding-config.md |
| DOC-032 | Standing opens ledger of the travelling-docs arc | partial | Carried-forward small items repeated across every revision | documentation-system.md |
| DEV-065 | CollectedData normalisation and data_helpers safe-access layer | deployed | Central data_helpers.go layer normalises every inbound message into a canonical CollectedData shape. | development-guide.md |
| BLD-007 | pageflow-builder (component-based site build orchestration) | deployed | Central v2-era monolithic builder; GEN-3's active all-in-one orchestrator | build-pipeline.md |
| SYS-042 | Mediator routing model | aspirational | Change descriptor matched against doc-tree metadata routing table to select consultees | system-architecture.md |
| DES-048 | No runtime re-compose path — layout change via the 025 FK-swap pattern | partial | Changing an existing site's layout is deliberately unsupported at runtime: `install_site_composition` refuses... | design-composition.md |
| IMG-046 | Data-graph / chart pipeline (code-rendered, never diffusion) (I4) | aspirational | Charts must be code-rendered from real data, never diffusion-generated; not built yet. | imagery.md |
| DOC-017 | The four doc actions (write_doc_plan, append_doc_note, load_doc_context, persist_diagnosis_note) | deployed | Chassis write/read surface for travelling docs, in production | documentation-system.md |
| BIZ-003 | Five-layer platform stack (L0 chassis → L5 VM backend deploy) | partial | Chassis→idea engine→idea.uk→vertical tools→tool-rich sites→VM deploy, each a customer | business-strategy.md |
| CTS-047 | Training-data export format (ChatML + metadata sidecar) | deployed | ChatML messages + ignored metadata sidecar; prose-not-JSON rows kept as DPO rejects | contracts-and-standards.md |
| TL-008 | The tool verification ladder (Tiers 0–4) | deployed | Cheap-to-expensive tiers, each catching a different class. Tier 0: generation-time output integrity... | tool-lifecycle.md |
| LCO-003 | Possibility-A-vs-B diagnostic method for silent LLM config failures | partial | Cheapest observability fix first, then audit-query to localise the real bug | llm-call-observability.md |
| CGV-004 | Page growth budget (growth_config, three-tier weekly limits) | deployed | CheckPageGrowthBudget caps content/blog/structural page creation per site per week | content-governance.md |
| SYS-036 | Parent/child result key = step-name convention | deployed | Child response is stored under the calling step's own name, never a synthetic key | system-architecture.md |
| STY-011 | Chrome selection path and the dead header_component_id column | deployed | Chrome resolution chain always falls through a NULL-forever column to fallbacks | styling-render-pipeline.md |
| RAGK-003 | rag_index action (chunk, embed, dedup, store) | partial | Chunks + SHA256 dedup + embed + insert; stores even if embedding fails | rag-knowledge-base.md |
| BIP-006 | vet-batch-processor agent | deployed | Claims and processes a batch of verification tasks sequentially; reworked in production | business-intelligence-platform.md |
| DOC-008 | Epistemic tagging and handoff-correction discipline | convention | Claims tagged verified/assumed/gap; correction log vs stale handoffs | documentation-system.md |
| CTS-019 | {function}-section class contract + data-component naming | partial | Class convention honoured unevenly; data-component attribute is the reliable escape hatch | contracts-and-standards.md |
| SPEC-011 | Classifier as strategic brain (always runs full) | partial | Classifier decides site destiny on every site; blocked items await feasibility | site-spec-and-classifier.md |
| RES-007 | Deep-research domain insight agent | abandoned | Classifier deciding when multi-platform social research pays off; domain-flipping era | research-agents.md |
| WDS-012 | needs_section_data: resolvable-by-query vs genuinely-human classification | partial | Classifies deferred section data into mechanically-resolvable vs permanently-HITL | work-dispatch.md |
| MCL-010 | RTT-based agent-type viability classification | aspirational | Classifies which agent types tolerate cross-cluster DB latency | multicluster.md |
| ONB-021 | Intake orchestrator with two HITL gates and per-group briefing questionnaires | partial | Classify→HITL confirm→group questionnaire→HITL review brief→spawn builder | onboarding-config.md |
| CASE-007 | relojistas.com go-live + bot verdict | deployed | Clean negative probe result: access log showed ~0 human intent | site-case-studies.md |
| VMB-009 | Pull-not-push off-cluster data return | partial | Cluster pulls over key-gated HTTPS on a schedule; box never holds credentials | vm-backend-sites.md |
| CHRT-001 | Chart component: Go static-SVG emitter + JS progressive enhancement | aspirational | Code-rendered chart plan: dependency-free Go SVG emitter plus inline JS enhancement, no CDN. | data-charts.md |
| SPEC-017 | write_site_spec spec_data string coercion (bugfix) | deployed | Coercion block accepts plain-string mission/roadmap briefs as JSON objects | site-spec-and-classifier.md |
| PEV-001 | Pragmatic Evolution model (explore/exploit portfolio cohorts) | abandoned | Cohort-based site mutation/evaluation; chaos confined to a "loser" cohort | portfolio-evolution.md |
| LCO-001 | Temperature/max_tokens logging gap in llm_call_log | partial | Columns exist but the Go writer never populates them from the actual API call | llm-call-observability.md |
| TLIB-016 | Component selector metadata and scoring (schema) | deployed | Columns that let a selector score components for a slot: section_type (kebab), suitable_site_types /... | tool-library.md |
| DOC-037 | verify_before_migration pre-flight convention | convention | Companion pre-flight SQL script before any hand-applied migration | documentation-system.md |
| SOC-003 | Arena + Stage dual modes and their mechanic families | aspirational | Competitive Arena vs showcase Stage, feeding each other in a content flywheel | social-media.md |
| FIX-030 | Whole-file rewrite strategy | deployed | Complete file bodies only, no diffs; caps near ~41KB | fix-loop.md |
| MDL-037 | llama3.3:70b trained but never used for inference | partial | Completed training run never wired to any agent's production ai_service | model-infrastructure.md |
| MDL-038 | BUG A: GenerateText never decodes stop_reason | deployed | Truncated LLM responses (max_tokens hit) silently look like complete successes; 17 live instances found | model-infrastructure.md |
| MDL-039 | BUG B: root ai_service SHADOWS step-level ai_service | deployed | Runbook rule was backwards; 17 agent defs have dead step-level max_tokens config (10 content-creator-* affected) | model-infrastructure.md |
| WII-003 | complete_work_item flag-preservation guard (Fix A) | deployed | Completion no longer clobbers deliberately-set needs_human_review flags | work-item-integrity.md |
| SYS-089 | Agent teams: composite/family/service-agent patterns (abandoned) | abandoned | Complex team-composition designs abandoned in favour of simpler agent groups | system-architecture.md |
| CTS-008 | JS content separation contract (js_content → assets) | deployed | Component JS split from html_template into js_content asset file; js_snippets for shared utils | contracts-and-standards.md |
| PLAN-027 | Display-name leak into section arrays + validate_components resolver | deployed | Component display_name leaked into sections array; resolver now maps to real function name | site-plan-and-reconciler.md |
| DYN-008 | Two JS delivery paths + inline-script truncation bug class | deployed | Component js_content extraction vs js_snippets bundle, plus a truncation bug | dynamic-applications.md |
| REB-001 | F3 scoped reason-stamped rerender (dependent-page scoping + reason propagation) | deployed | Component regen creates reason-stamped rerender items scoped to dependent pages only | rebuild-cascade.md |
| CAP-001 | Component asset coupling not enforced (JS/data file existence) | aspirational | Component templates reference external JS/data files with no pipeline guarantee they exist. | component-asset-pipeline.md |
| CVP-002 | Behavioural models library and functional component labelling | partial | Components labelled by behavioural function (AIDA/PAS/Fogg/Cialdini), not visual pattern | conversion-playbooks.md |
| CGV-010 | Silent-fallback link family (phantom /contact.html, /services.html) | unknown | Components link to nonexistent pages instead of resolving or degrading gracefully | content-governance.md |
| TLIB-022 | Shared component library semantics + field-set guard + neutral-base/fork rule | deployed | Components with `forked_from IS NULL` form a cross-site SHARED library keyed by `function` (e.g.... | tool-library.md |
| DOC-026 | load_doc_context fix-time retrieval | deployed | Composes current PLAN + latest NOTES + criteria_json into one block | documentation-system.md |
| IMG-025 | Image generation as parameter shaping (deferred composer step) | aspirational | Composition-by-parameter design for images deferred; partially realised by per-kind gating. | imagery.md |
| CASE-003 | idea.uk chassis-site build state (two site rows; gated go-live) | partial | Concrete build history across two site_ids with catalogued page defects | site-case-studies.md |
| DOC-053 | Three parallel threads with hard boundaries | convention | Concurrent chat threads own non-overlapping territories | documentation-system.md |
| AGOV-007 | Two gated paths: config changes vs deliverables | aspirational | Config confirmer and ledger ratchet are two distinct gates | autonomy-governance.md |
| DOC-030 | Provenance stamps the chassis, not the logical agent | deployed | Config-declared source fields are the reliable provenance, not agent headers | documentation-system.md |
| FIX-040 | config_change edit operation type | deployed | Config-only edits labelled but left for a human to apply | fix-loop.md |
| ABO-005 | governance/HITL principles | aspirational | Confirm-not-initiate, sealed-ancestor inheritance, one privileged transition | autonomous-build-operate.md |
| CQ-015 | identity-advisor agent and sites.approval_mode gate — never built | abandoned | Confirmed absent: three-way finding_type routing and its specialist agents never built | content-quality.md |
| PLAN-031 | New-domain build pipeline stage chain (domain-submitter → page-build-handler) | partial | Confirmed happy-path chain for building a brand-new domain end to end | site-plan-and-reconciler.md |
| WII-006 | Work-item routing map (item_type → handler agent) | deployed | Confirmed live mapping from item_type to the four content/rerender handler agents | work-item-integrity.md |
| SCH-017 | Thunder unreachable-probe counter | deployed | Consecutive-unreachable-probe counter distinguishes SSH blip from truly-lost box | scheduler-and-tasks.md |
| ADO-035 | Doc-tree adoption plan (category mismatch: docs retrieval, not site adoption) | aspirational | Constitution + tag/embedding retrieval plan for the doc corpus | adoption-pipeline.md |
| FIX-025 | Build gate (diagnose_build_gate) | deployed | Containerized gofmt+build gate; red blocks the PR | fix-loop.md |
| I18N-001 | Language handling: implicit mechanism plus minimal explicit prompt support | partial | Content language rides implicitly on context; a ## Language prompt section is the only explicit signal. | language-i18n.md |
| SOC-008 | AI cost architecture: fixed background vs per-user scaling | aspirational | Content-gen cost is fixed (~£5/day); only sparring/scoring scales per-user | social-media.md |
| HITL-019 | HITL review flow (needs_human_review → retry/resolve/spec-edit) | deployed | Content-quality flags create needs_human_review items with 3 resolution paths | hitl.md |
| BIZ-025 | Content-site valuation model (24-32x) | aspirational | Content/affiliate sites value at 24-32x monthly profit; underpins portfolio strategy | business-strategy.md |
| DOC-045 | Standing conformance suite (carved out, deliberately not built) | aspirational | Continuous behave-as-documented monitor, deliberately scoped out | documentation-system.md |
| SNAP-004 | Snapshots and revert for agent definitions (snapshot_agent/revert_agent) | deployed | Convention to snapshot before patching default_config; audit trail table | site-snapshots-and-revert.md |
| LQT-004 | LLM model governance: aliases, per-step model choice, llm_call_log | deployed | Conventions tying aliases, tiering, and call logging together across ~90 agents | llm-quality-testing.md |
| PLAN-013 | Adoption-faithfulness convergence (reconcilePlanWithRealised): sequential root-cause fixes | deployed | Convergence was inert through two sequential bugs before finally working | site-plan-and-reconciler.md |
| BIZ-019 | WordPress export/handoff idea (XML export, shortcodes, subscription plugin) | abandoned | Convert generated sites to WordPress for client handoff; killed by competitive analysis | business-strategy.md |
| ONB-015 | Onboarding orchestrator (dependency-graph flow; active-with-pending) | aspirational | Coordinates three layer agents by dependency; terminal state active-with-pending | onboarding-config.md |
| CTS-048 | Local-step input resolution: input_mapping dead, key_path for loops | deployed | Coordinator doesn't resolve input_mapping for local action/loop substeps; use key_path | contracts-and-standards.md |
| DEV-045 | Cross-module port / copy-drift discipline | convention | Copy the WHOLE package as a unit and diff file lists; grep real helper APIs before authoring. | development-guide.md |
| DEV-081 | Specialist agent design doctrine (agents own their domain) | deployed | Core agent-design rulebook: self-contained, raw-identifier callers, declarative workflows, spawn-before-call. | development-guide.md |
| SYS-068 | Layer-1 / Layer-2 hack-resistance model | deployed | Core cluster never serves inbound traffic; Layer 2 is static-on-S3 with nothing to compromise | system-architecture.md |
| HITL-012 | await_approval / HITL pause-resume mechanism (AwaitedRequests, token matching) | deployed | Core pause/resume primitive: token-based approval request, matched Kafka resume | hitl.md |
| FIX-005 | diagnosis_artifacts table (unified egress store) | deployed | Correlation-keyed egress table, kind grows bundle→escalation | fix-loop.md |
| ADP-013 | Thunder consecutive-unreachable probe streak | deployed | Counter-based durability so one transient SSH blip doesn't kill a training run | adapters.md |
| ADO-002 | Adoption is a one-off capture, not a ceiling (specs separation) | deployed | Crawl data goes to research_results, never site_specs; strategist extends beyond baseline | adoption-pipeline.md |
| DOC-064 | Deploy-from-context-packs guide — six deploy mechanisms (A–F) — dropped from the live tree | partial | Cross-cutting deploy-mechanism guide, absent from the live idea.uk tree | documentation-system.md |
| DEV-019 | Standing session/working-contract rules (house rules) | convention | Cross-thread working contract: Go not Python, British English, schema-first, no summary docs. | development-guide.md |
| CTXA-008 | Diagnostic playbooks / failure fingerprints as authored knowledge | aspirational | Curated failure signature + confirm commands + fix pattern, surfaced into debug bundles like standards | context-assembly.md |
| NEWS-006 | News publishing gap (curation -> deployed posts) | aspirational | Curated news items never become deployed blog posts; Path B design to close the gap | news-feed-pipeline.md |
| ASG-006 | Controlled group evolution (observed mutation with rules) | aspirational | Curated seeds + metrics-triggered mutations; platform/evolution/ confirmed real | agent-spawning-and-groups.md |
| PLAN-035 | build-site-planner + roadmap-overrides-components rule | deployed | Current needs_site_plan handler; roadmap overrides the component list when present | site-plan-and-reconciler.md |
| CASE-014 | Cross-vendor critique (multi-model critique step) | deployed | Cut step runs on a different model vendor than generate, logged explicitly | site-case-studies.md |
| ABO-006 | autonomous-system building-block hardening checklist | aspirational | Cycle guards, outbox transactions, bulk confirmation, no self-recovering rollback | autonomous-build-operate.md |
| DEV-047 | Building-discipline edge cases (pre-registered engineering checklist) | aspirational | Cycle guards, outbox-pattern apply, consistent snapshots, bulk confirmation, not-yet vs broken. | development-guide.md |
| SYS-048 | awaited_requests global request/response registry | deployed | DB-backed registry matching Kafka responses to waiting orchestrations | system-architecture.md |
| DBG-064 | Orchestration debug log taxonomy (early ancestor of the formal guide) | superseded | DEBUGaa grep targets + pg_stat_activity lock triage; ancestral to 016/016b | debugging.md |
| VONC-004 | provocation-card component (daily hero card) + mini-lobby trim | partial | Daily hero card runtime-filled from JSON; mini-lobby trim blocked on bundle verdict | vonc.md |
| SCH-021 | Retention prune timer | deployed | Daily systemd timer prunes events-*.jsonl older than RETENTION_DAYS (default 90) | scheduler-and-tasks.md |
| DES-030 | Layout: tool-portal-dark | partial | Dark developer-utility portal layout supporting three page shapes in one template — portal/index, tool pages,... | design-composition.md |
| STY-033 | Section-contrast model (is_dark_section + --section-* variables) | deployed | Dark-background components must define the --section-* contract on container | styling-render-pipeline.md |
| DEV-072 | Dynamic agent routing (route_by_field / conditional_call_agent) | deployed | Data-driven agent selection inside workflows via a dot-path value → agent-type mapping. | development-guide.md |
| NEWS-017 | Blog-listing / orphan-page routing session handoff | deployed | Dated fix package for blog-listing rendering and three-way orphan-page reclassification | news-feed-pipeline.md |
| DOC-060 | Handoff document convention (stand-alone dated brief for a fresh chat) — vonc thread | deployed | Dated self-contained handoff with orientation, DONE state, backlog | documentation-system.md |
| DOC-044 | Date/version as triage, not truth | convention | Dates order the verification queue; never override a code check | documentation-system.md |
| DOC-056 | Published reasoning as substrate + drift detection | aspirational | Decisions publish their premise, enabling drift detection | documentation-system.md |
| DEV-052 | Confirmed chassis workflow model (group agents, promotion pattern, wrapper orchestrator) | deployed | Declarative JSON workflows, generic action library reused first, promote to spawned sub-agent later. | development-guide.md |
| OPD-002 | Parallel-thread boundary and handoff convention | convention | Declared ownership boundaries and handoff docs across concurrent threads | operating-doctrine.md |
| ADO-004 | Source vs destination separation (target_url / destination_domain) | deployed | Decouples crawled site from built site; mismatch used to silently drop content | adoption-pipeline.md |
| VMB-013 | VM launch plan (idea.uk): dedicated hardened box | deployed | Dedicated VM chosen over reusing the shared OVH multi-domain reverse proxy | vm-backend-sites.md |
| IMG-017 | Legacy image_prompts age-out check — retired (Phase 2G.6) | superseded | Dedicated migration check retired; reframed as operational deregistration decision. | imagery.md |
| SYS-049 | Message deduplication (processed_messages) | deployed | Dedup key + composite PK blocks duplicate delivery within a retry generation | system-architecture.md |
| IMP-048 | Work-item dedup and two-strike semantics (partial index behaviour) | deployed | Dedup semantics that shape operations: the partial unique index suppresses a new work item only while an open one... | improvement-loop.md |
| STY-044 | Head-inside-body bug and positional injection fixes | deployed | Dedup-by-size heuristic kept the wrong misplaced head; fixed to dedup by position | styling-render-pipeline.md |
| ADO-030 | Playwright capture adapter + website-capture agent | superseded | Deep browser capture deferred in favour of managed Firecrawl | adoption-pipeline.md |
| DBI-008 | sites.build_status vestigial column | aspirational | Defaults to pending and is never advanced by any code path | database-and-infrastructure.md |
| DIAG-009 | Three-guard read-only SQL enforcement model | deployed | Defence in depth for model SQL: prompt constraint, lint, and the real guarantee — read-only transaction | diagnosis-loop.md |
| RAGK-006 | Concept-document RAG for content writers (v2+) | aspirational | Deferred design to ingest full concept docs into knowledge_base for v2+ | rag-knowledge-base.md |
| FTW-041 | Text LoRA — veterinary knowledge extractor | aspirational | Deferred recipe to fine-tune a local 7-8B knowledge-extractor model | finetuning-flywheel.md |
| STY-037 | Content/structure separation: JSON content + html-assembler | superseded | Defined the modern content/template split; taxonomy ancestor of current pipeline | styling-render-pipeline.md |
| DBG-012 | Open problem: nav-updater never spawns | unknown | Definition/topics exist but no pod ever starts; nav_drift items always claim-timeout | debugging.md |
| CQ-012 | Prompt composition asymmetry (text cascade vs image) | aspirational | Deliberate choice to keep single-prepend image cascade separate from text composition | content-quality.md |
| CHAT-002 | Site chatbot edge worker (synchronous, not an orchestrated agent) | aspirational | Deliberate exception to "every agent is an orchestrator": sync SSE edge handler | site-chatbot.md |
| SYS-076 | idea.uk topology exception | deployed | Deliberate exception to the serverless-edge default: a small always-on backend | system-architecture.md |
| SYS-086 | HTML-first progressive enhancement delivery | deployed | Deliberate plain HTML/CSS/JS generation strategy that survived into the renderer | system-architecture.md |
| TRF-016 | Probe content restraint — no results, no imagery, no anchoring | deployed | Deliberately no results page and no photos in v1 so displayed content can't bias the signal | traffic-analytics.md |
| RAGK-005 | knowledge-indexer agent (deferred) | aspirational | Deliberately unbuilt owning-agent; actions suffice until a use case demands it | rag-knowledge-base.md |
| DIAG-015 | Live schema section in the bundle (gatherSchema) | deployed | Denylist-driven live information_schema section stops the model guessing table names | diagnosis-loop.md |
| CASE-018 | Reuse-not-rebuild site build-out with honest "simulation" labelling | aspirational | Deploy existing tool library, label deterministic widgets as simulations | site-case-studies.md |
| PBP-024 | Deploy-observability bookkeeping gap | partial | Deploy path never writes deploy_commit/last_built_at, weakening change-detection evidence | page-build-pipeline.md |
| PBP-025 | componentless_pages discovery check (repair half of PBP-023) | built, not enabled | Active+deployed page with sections but ZERO page_components serves chrome only and no check could see it | page-build-pipeline.md |
| PLAN-004 | built_from_plan_version drift stamp + removal of deployed→needs_rebuild flip (Option B) | deployed | Deploy-time plan-version stamp replaces the blunt sync-time rebuild flip | site-plan-and-reconciler.md |
| STY-046 | CSS generation bug (webdesign-agent design_spec not applied) | superseded | Deployed CSS reverts to default blue template despite a correct design_spec | styling-render-pipeline.md |
| DEV-017 | Agent re-registration vs re-seed risk (DB row authoritative) | deployed | Deploys bump updated_at but don't overwrite default_config; DB-edited prompts survive deploys. | development-guide.md |
| PLAN-006 | Architectural Tension #1: infer-and-repair vs deterministic structure derivation | partial | Derive page structure from LLM naming signal, not vertical-hardcoded repair heuristics | site-plan-and-reconciler.md |
| ONB-005 | Config as a maintained artifact (wizard is first pass; lifecycle is deliverable) | aspirational | Derived config gets periodic re-derivation, drift flagging, provenance | onboarding-config.md |
| ABP-001 | Automated Go action build pipeline (compiler pod) | aspirational | Design for an in-cluster compiler pod that builds/tests/deploys LLM-written Go actions | action-build-pipeline.md |
| DES-003 | Composition pipeline: direction → composition → execution (site-design-planner + webdesign-agent) | deployed | Design is deliberately a two/three-stage pipeline, not one agent. | design-composition.md |
| IMG-004 | Two lanes of imagery: plan-driven vs content-driven (Lane B) | aspirational | Design split between fixed plan-time imagery and continuous content-attached imagery. | imagery.md |
| STY-014 | Phase 4.5 data-section-bg surface generalisation (deferred) | aspirational | Designed attribute-based surface decouple, deferred as unneeded dark-site concern | styling-render-pipeline.md |
| DYN-011 | loader-builder agent + section descriptor + Tier E runtime-feed source | aspirational | Designed autonomy path from hand-built loaders to LLM-generated ones | dynamic-applications.md |
| CTXA-002 | Bundle shape contract (the task-scoped context package) | aspirational | Designed fixed-section bundle contract: metadata, authored layer, code context, DB data, pointers, provenance | context-assembly.md |
| WDS-013 | Side-effect rules engine (deterministic follow-on items) | partial | Deterministic Go rules meant to emit nav/sitemap/redirect/rerender follow-ons per completion | work-dispatch.md |
| PLAN-014 | First-plan branch: "no current plan + pages exist ⇒ adopted pages" | deployed | Deterministic detection of a faithful first pass via adoption_locked flag | site-plan-and-reconciler.md |
| CTXK-013 | docselect / queryselect: per-hypothesis doc and query selection | deployed | Deterministic keyword/path-glob selectors pulling only relevant docs or vetted queries into each iteration | contextkit-toolchain.md |
| CTXK-008 | resolve_targets (cmd/resolve_targets) | deployed | Deterministic lexical-overlap baseline proposing ranked candidate symbols for a task | contextkit-toolchain.md |
| PLAN-017 | Bare-sibling duplicate pages (planner re-invents adopted topics) | deployed | Deterministic stem-dedup guard drops differently-slugged planner-invented siblings | site-plan-and-reconciler.md |
| FIX-013 | Plan validation / hard allowlist for edit plans | deployed | Deterministic validator fails closed on malformed plans | fix-loop.md |
| ONB-009 | Three checking tiers + three-bucket audit output (coverage honesty) | aspirational | Deterministic/heuristic/judgement-only tiers; audit reports three numbers | onboarding-config.md |
| BIZ-028 | Domain value maximisation pipeline (domain flipping) | abandoned | Develop parked domains to multiply resale value; platform pivoted to operating own sites | business-strategy.md |
| FIX-009 | Blinding discipline for benchmark runs | convention | Diagnose-agent can't read docs; symptom string is the only leak | fix-loop.md |
| DIAG-002 | Chassis workflow architecture (diagnose_route, next_step override) | deployed | Diagnose-agent workflow: analyse→lookup→runtime→bundle→verdict→route loop-back via next_step override | diagnosis-loop.md |
| STY-010 | Hazard-class vs band-class declarer taxonomy (library blast radius) | partial | Diagnostic split sizing every scheme-fix decision; non-idea.uk tail still open | styling-render-pipeline.md |
| CTXA-007 | Run signatures: expected-vs-actual sequence diff | aspirational | Diff a live run against a captured healthy-run signature to surface the divergence point directly | context-assembly.md |
| DOC-036 | Full heading+content-line diff across all forked copies before consolidating | convention | Diff-before-promote methodology for travelling/forked docs | documentation-system.md |
| FIX-049 | Fix-loop value proposition: unattended, cited, consistent | convention | Differentiator is auditable consistency, not superhuman insight | fix-loop.md |
| SYS-050 | Orchestration ↔ site linkage (orchestration_states.site_id) | deployed | Direct nullable site_id column replaces JSONB spelunking for per-site orchestrations | system-architecture.md |
| IMG-062 | Components declare imagery contracts / many-images-per-page | aspirational | Direction for components to own typed imagery contracts instead of hero_image-only. | imagery.md |
| IMP-007 | Discovery agents on dead/stub sites (noise at scale) | unknown | Discovery agents keep generating remediation items for sites that are deleted, stubs, or mid-adoption. Proposed... | improvement-loop.md |
| LNK-009 | check_phantom_internal_links post-deploy audit + surface routing | partial | Discovery check built and gate-cleared but deliberately not yet enabled | link-management.md |
| DBG-047 | Pipeline field as a soft routing label (needs_imagery excluded by pipeline='design') | deployed | Discovery checks inherited wrong pipeline label; dispatcher filter also loosened | debugging.md |
| NAV-004 | Nav discovery checks and fix agents | deployed | Discovery/fixer pairs for anchor-slug links, stacked nav, unlinked components, orphans | navigation.md |
| DEV-061 | Loop-action dispatch (build-dispatch-loop, migration 071) | deployed | Dispatch loop claims/spawns/calls one work item per iteration via claim→spawn→call→mark_complete. | development-guide.md |
| WDS-005 | Dispatcher response-stall and missing claim/orchestration timeout cleanup | partial | Dispatch orchestrations stall despite handler response arriving; no sweeper existed | work-dispatch.md |
| SCH-011 | Pipeline-blind dispatch surfaces (discovered platform defect) | deployed | Dispatch queries lack pipeline filters; any item on a claimable site gets dispatched | scheduler-and-tasks.md |
| SCH-005 | Dispatch throughput bottleneck (Family J): one-site-per-tick, NOT-EXISTS-blocked | unknown | Dispatcher processes one site per tick and blocks entire site on any claimed item | scheduler-and-tasks.md |
| DEV-013 | Authoring rules pack (20-rule bundle) | deployed | Distilled 20-rule authoring discipline (schema-check, parameterised SQL, code-fence stripping...). | development-guide.md |
| DES-080 | site-design-planner spec aspects as originally scoped (navigation / layout / resolved_composition) | partial | Doc 103 defined three site_specs aspects site-design-planner was originally slated to write, separated by... | design-composition.md |
| CTXA-018 | Documentation indexing rides the prose rag path | aspirational | Docs are prose, so index them via existing rag_index/knowledge_base rather than a new mechanism | context-assembly.md |
| ONB-007 | Docs-authoritative conventions for our own repo (the free drift audit) | aspirational | Docs authoritative for our repo; code disagreements recorded as free drift audit | onboarding-config.md |
| DOC-015 | Doc subject convention — ('tool', function) and ('pipeline', ...) | deployed | Docs keyed by subject_type/subject_key; tool vs pipeline conventions | documentation-system.md |
| IMP-022 | Colour-fix algorithmic detail (countHardcodedColorComponents / findForcedTextColors) | superseded | Documented exact algorithmic mechanics for two design-discovery-agent colour checks: `hardcoded_section_colors`... | improvement-loop.md |
| CTXK-016 | contextkit bundle regeneration procedure | deployed | Documented, tested human-facing bundle regeneration procedure distinct from the live loop's own retrieval | contextkit-toolchain.md |
| SPEC-020 | Catch-all email forwarding, abandoned for per-site forwarders | abandoned | Domain-level catch-all repeatedly bounced; specific forwarders chosen instead | site-spec-and-classifier.md |
| DEV-060 | build_queue site seeding | deployed | Domain-level intake queue for new sites; seed_build_queue is the entry point into the work-item pipeline. | development-guide.md |
| DBG-016 | SQL/DB template-and-data surgery method (needle-gate + Postgres pitfalls) | deployed | Dominant safe method for mutating production templates/prompts/configs by SQL | debugging.md |
| INVD-002 | Verify-before-acting investigation discipline | convention | Don't act on a theory until code-search verifies it | investigation-discipline.md |
| DGH-003 | sites.github_repo deploy-target selector / resolveGitRepoName patch | aspirational | Dormant deploy-target column; 3-touch patch to actually wire it up | deployment-github.md |
| STG-005 | asset-deployer (S3 → optimize-by-purpose → git) | deployed | Downloads S3 asset, optimizes by purpose (logo/hero), commits to git | storage-architecture.md |
| BIZ-026 | Data-sovereignty / pilot-first / startup fast-start positioning | aspirational | Drafted client-engagement angles; explicitly "not yet done for a client" | business-strategy.md |
| DOC-035 | Single-source relocation with pointer / canonical-doc-home discipline | convention | Duplicated topics consolidated into one numbered doc + pointer sentence | documentation-system.md |
| STY-018 | Stored⊕resolved merge writes resolved values back into content_data | deployed | Durable recoveries, but also a contamination carrier for bad fallbacks | styling-render-pipeline.md |
| FIX-003 | needs_diagnosis intake route (F0.1c) | deployed | Durable site_work_items intake joined to loop by correlation_id | fix-loop.md |
| MDL-024 | Static vs dynamic agent deployment + GPU cost strategy | aspirational | Dynamic spawned GPU workers claimed as 95% cheaper than static GPU deployment | model-infrastructure.md |
| DES-067 | structure_tokens JSONB convention | deployed | Each `layouts` row carries a `structure_tokens` JSONB column holding non-colour design tokens — spacing, radii,... | design-composition.md |
| DBI-005 | Schema-per-client multi-tenancy (create_client_schema) | deployed | Each client gets an isolated client_<id> schema created by a SQL function | database-and-infrastructure.md |
| RSN-008 | Multi-author generation | aspirational | Each concern authors a full solution instead of guarding one | reasoning.md |
| IMP-024 | Acceptance-test cheap-LLM verification call gating lock + retry | superseded | Each finding carried an `acceptance_test` enabling a cheap follow-up LLM call after a fix: feed fixed HTML back,... | improvement-loop.md |
| CLC-010 | llm_guidance as a per-field generation-steering surface | deployed | Each input_schema.fields entry may carry llm_guidance, which page-content-writer renders into its... | component-lifecycle.md |
| ADO-021 | Section recipes for adoption | aspirational | Each section captured as purpose+structure+reference implementation recipe | adoption-pipeline.md |
| IMP-040 | Per-site maintenance profile with budgets | partial | Each site declares which maintenance domains run, at what cadence, with which sub-agents and regulatory bodies,... | improvement-loop.md |
| NAV-011 | Global context injection for navigation (superseded) | superseded | Earlier Global.Sitemap design, superseded by nav tables + GetNavItems | navigation.md |
| NEWS-019 | News & content feed pipeline (mid-era design, superseded) | superseded | Earlier article-rewriter/publisher/lifecycle-decay design, ancestor of the deployed pipeline | news-feed-pipeline.md |
| DIAG-020 | Abandoned design: diagnose_run single engine-wrapping action | abandoned | Earlier design ran the whole capped loop inside one action; dropped for the observable workflow-driven loop | diagnosis-loop.md |
| ABO-001 | trust ledger — earlier draft | aspirational | Earlier draft of the same contract now live as AGOV-004 | autonomous-build-operate.md |
| ABO-002 | change-layer integration contract — earlier draft | aspirational | Earlier draft of the same contract now live as AGOV-006 | autonomous-build-operate.md |
| DES-041 | Component-creation via HITL work-item triage — superseded | superseded | Earlier plan for seeding new library components via work items routed through HITL triage. | design-composition.md |
| BLD-012 | MVP build squad lineage (chief-strategist → architect → content-creator → deployer) | superseded | Earliest 4-agent builder lineage, ancestor of the GEN-1/2/3 generations | build-pipeline.md |
| STY-036 | aggregate_webpage HTML assembly action (first-gen renderer) | partial | Earliest page renderer, one action call per page, long since replaced | styling-render-pipeline.md |
| CASE-009 | Original first-domain set (dropped surgerylight + finance/retail) | abandoned | Early 5-domain starter set silently trimmed to two named domains | site-case-studies.md |
| DBG-060 | Message-flow logging / observability plan (never fully built) | aspirational | Early MessageFlowLogger aspiration; only zap logs + processing_history ever built | debugging.md |
| IMG-035 | Image storage and display URL strategy (S3/B2 dual URI) | deployed | Early dual s3://+https:// URI decision; public-bucket/CDN chosen over presigned URLs. | imagery.md |
| SCH-019 | Superseded checkpoint-JSON / events-per-1k ranking (early P4) | superseded | Early explicit checkpoint-file design dropped for structural idempotency | scheduler-and-tasks.md |
| DEV-077 | Deliberate discovery + human-approved agent evolution (abandoned) | abandoned | Early governance: agents only self-modify at planning/review stages, always with human approval. | development-guide.md |
| SYS-084 | Inter-agent invocation protocol v1 (superseded) | superseded | Early invoke_agent/agent_invocations design replaced by call_agent + headers | system-architecture.md |
| PLAN-037 | Multi-page site support (wrap_multipage, multipage-site-builder) | superseded | Early multi-page extension of the single-page pipeline; superseded by the plan/pages domain | site-plan-and-reconciler.md |
| SYS-029 | Self-spawning flat dispatch-loop (pre-scheduler, superseded) | superseded | Early one-item-then-respawn design replaced by scheduler tick + in-workflow loop | system-architecture.md |
| DBG-074 | kcat + db-inspector operational runbook | deployed | Early ops playbook for triggering/tracing workflows in the live cluster | debugging.md |
| DIAG-040 | Base-runbook gated-items framing (documentation-style lineage note) | superseded | Early runbook style deferred the roadmap to a separate PLAN.md; superseded by inlined self-contained runbooks | diagnosis-loop.md |
| ADO-029 | website-analyzer conditional scraping group | deployed | Early smart capture entry point routing scrape/extract/crawl | adoption-pipeline.md |
| ONB-020 | Briefing agent (early industry-brief / clarifying-question stage, pre-questionnaire) | partial | Early two-era briefing agent generating brief JSON, later superseded | onboarding-config.md |
| SYS-066 | Agent families architecture | partial | Eight specialist-agent families each owning a data domain, mixed completion | system-architecture.md |
| DES-057 | Colour Inheritance Model (two-tier `var(--section-*, var(--color-*))` fallback) | deployed | Element-level colour rules (headings, body text, links) resolve via a two-tier CSS custom-property fallback... | design-composition.md |
| CTXA-016 | Text-vs-code embeddings: share the mechanism, separate the policy (B4b) | deployed | Embedder/pgvector/hybrid pattern shared with prose RAG; model, tuning, and evaluation kept separate per domain | context-assembly.md |
| RAGK-002 | rag_lookup action (vector search + trigram fallback) | deployed | Embeds query, cosine search within collection, trigram fallback when Ollama down | rag-knowledge-base.md |
| IMG-015 | check_unfulfilled_imagery_plan discovery check (Phase 2G.4) | deployed | Emits needs_imagery per unfulfilled site_plan_imagery row, priority-banded by scope. | imagery.md |
| REB-007 | page-build-handler writes only planned sections (sections=0 → silent no-op rebuild) | deployed | Empty planned-section list makes a rebuild complete having written nothing | rebuild-cascade.md |
| DYN-007 | Runtime-fill mechanism (data-runtime-fill shells + client loaders) | deployed | Empty shells filled client-side from a JSON feed, proven three times over | dynamic-applications.md |
| IMG-055 | Section-scope imagery pipeline — idea.uk verification | deployed | End-to-end plan→emit→generate→deploy→rebuild chain exercised live on idea.uk. | imagery.md |
| FIX-001 | Diagnosis→fix loop programme / council loop (F0–F3) | deployed | End-to-end symptom-to-PR pipeline; PR #1 merged 2026-07-13 | fix-loop.md |
| TRF-009 | Access-log passive harvest and /access-digest | deployed | Engine reads its own nginx log for referer/404/UA signals the event stream can't see | traffic-analytics.md |
| PAY-008 | REVIEW_BEFORE_PAY billing flow supersedes charge-first flow (idea.uk) | deployed | Engine runs and operator reviews before any pay link is sent; charge-first kept as fallback | payments.md |
| TRF-007 | Capture-side input sanitisation with deferred normalisation | deployed | Engine strips Cc/Cf and collapses whitespace; NFC/lowercasing deferred to the collector | traffic-analytics.md |
| CQ-013 | Input sanitisation (sanitizeValue, Cc/Cf stripping, NFD survives) | deployed | Engine strips control/format chars, collapses whitespace correctly, defers NFC to collector | content-quality.md |
| DEV-073 | spawn_group action with DB group lookup and dynamic group_type | aspirational | Enhanced version resolves the whole agent group (agents+workflow+questionnaire) from the database. | development-guide.md |
| CTXE-003 | Corpus enrichment policy: measure first, mechanical before authored | aspirational | Enrich retrieval with rot-free mechanical signals (string literals) before hand-authored docs or tags | context-engineering-principles.md |
| FTW-040 | Thunder adapter (credential-boundary GPU provisioning) | deployed | Ephemeral credential-free VMs, spend/uptime/concurrency caps, 15-min reaper | finetuning-flywheel.md |
| ONB-013 | Sandboxed probing — the tenant-code security envelope | aspirational | Ephemeral read-only sandbox gates the first agent allowed to run tenant code | onboarding-config.md |
| FTW-004 | llm_call_log as ops visibility + training-data capture | deployed | Every LLM call logged fire-and-forget with flywheel columns for training exports | finetuning-flywheel.md |
| BIZ-012 | Fractal agent architecture claim (self-similar recursive orchestration) | deployed | Every agent is an orchestrator; traced 7 levels deep with identical code paths | business-strategy.md |
| DEV-068 | agent_definitions registry (DB-driven agent config and versioning) | deployed | Every agent type is a database row; creating an agent is an insert, not a code change. | development-guide.md |
| IMG-054 | DeployedWebPath committed-path convention (two-URL serving model) | deployed | Every asset has a throwaway presigned URL and a durable committed git path; render uses latter. | imagery.md |
| DIAG-008 | Three-tier citation standard (static / state / runtime) | deployed | Every citation tagged static/state/runtime with freshness; strongest verdicts ground across all three | diagnosis-loop.md |
| TL-020 | Validation observability: structured rejection logging (recordValidationRejection) | deployed | Every component pre-store validation rejection writes a structured agent_error_log row (severity warning for... | tool-lifecycle.md |
| CASE-017 | Claim-evidence audit rule ("no claim ships without an audit row") | deployed | Every marketing claim verified against code/DB/HTTP before it may ship | site-case-studies.md |
| TL-007 | Tool doc header system (sentinel header, stripped at deploy, provenance) | deployed | Every new tool's script opens with one sentinel-delimited block (function/purpose/behaviour/inputs/outputs; never... | tool-lifecycle.md |
| CTS-020 | Paired-variable ("on-colour") standard | deployed | Every paintable band colour has a matching curated text colour, overridable per site | contracts-and-standards.md |
| DEV-058 | site_work_items unified work queue and lifecycle | deployed | Every piece of platform work is a row: source, pipeline, spec JSONB, lifecycle statuses, dedup key. | development-guide.md |
| PLAN-040 | Planner role-aware ≥1-section invariant + role→pipeline mapping | aspirational | Every planned page whose role page-build-handler owns must have ≥1 section (not yet enforced) | site-plan-and-reconciler.md |
| SYS-056 | SagaCoordinator engine: embedded, distributed, no central orchestrator | deployed | Every pod embeds a full orchestrator loading JSON workflows from the DB | system-architecture.md |
| LNK-022 | CTA-graph integrity (dead-end and circular primary actions) | partial | Every primary CTA once 404'd, then became circular; retarget deliberately parked | link-management.md |
| TRF-021 | Per-domain notes and living-docs convention | deployed | Every probe domain gets a living notes file; project knowledge lives in plan/runbook/notes | traffic-analytics.md |
| BIP-003 | Data observations provenance model | deployed | Every scrape/search logged as a data_observations row; staleness columns track freshness | business-intelligence-platform.md |
| DOC-022 | Workflow-altering migrations write pipeline NOTES | deployed | Every workflow-altering migration appends a pipeline/build doc_notes entry | documentation-system.md |
| SCH-022 | claimed-item-timeout evidence-gated completion + reset (Lever A/C) | deployed | Evidence-gated auto-completion/reset already existed; avoided duplicate watchdog build | scheduler-and-tasks.md |
| SCH-023 | Firing a DISABLED scheduled task once, at a target you choose | deployed | One-off fireTrigger() envelope; target is an argument because a selecting pre_query picks the wrong row | scheduler-and-tasks.md |
| DOC-041 | Doc-drift claim classifier (design only, named across three units, never built) | aspirational | Evidence-or-abstain claim classifier, tiered T1-T3, consistently deferred | documentation-system.md |
| STY-030 | CSS applies_to granularity mismatch (known issue, unfixed) | partial | Exact-text overlap matching means only 2 of ~21 snippets ever ship | styling-render-pipeline.md |
| DOC-047 | dedup (cmd/dedup) | deployed | Exact/near-duplicate file finder with report-only default and undo manifest | documentation-system.md |
| DBI-011 | Ownership hierarchy reuse for entitlement scoping | deployed | Existing clients→networks→sites FK chain reused for billing/entitlement instead of new ownership | database-and-infrastructure.md |
| STY-015 | Explicit RenderContext.Scheme signal (Q1) — abandoned design | abandoned | Explicit scheme plumbing dropped once implicit palette signal was found sufficient | styling-render-pipeline.md |
| CTS-035 | Priority profile (order not weights; sealed constraints) | aspirational | Exploratory mediator-framework design; adjacent-project material, not core platform | contracts-and-standards.md |
| STG-008 | Persistence design: tiered one-way data flow (box → B2 → chassis) | partial | Exposed box writes B2 dead-drop; chassis pulls and ingests, never the reverse | storage-architecture.md |
| MCL-011 | Cross-cloud cluster expansion (Phase 4: AWS EKS / GCP GKE) | aspirational | Extends adjacent-cluster pattern to a genuinely remote cloud provider | multicluster.md |
| IMG-018 | Image-generator request shape + per-kind defaults (Phase 2H) | deployed | Extends request with negative_prompt/seed/reference_image_uri and Go kindDefaults map. | imagery.md |
| ONB-008 | Conventions agent (extract-cite-confirm, then audit) | aspirational | Extracts convention atoms, human-confirms, then audits code for drift | onboarding-config.md |
| LOCK-005 | Adoption faithfulness via 90-day timed locks | partial | Faithful-first-pass lock originates at first re-plan; re-plan-window enforcement undeployed | locks.md |
| LNK-013 | page-content-writer ↔ resolver wiring with regression-safe fallback chain | deployed | Fallback-to-prior-behaviour wiring later masked a bug for two weeks | link-management.md |
| IDEA-014 | SFI26 Diff Alerts (first vertical tool) — replacing single-farm assessment | aspirational | Farm-advisor digest tool; replaced a Risk-2 single-farm assessment product | idea-product.md |
| DBG-023 | Send-before-register await race (preRegisterAwaitedRequest fix) | deployed | Fast adapter reply beat awaited_requests insert; fixed by register-before-send | debugging.md |
| RSN-005 | Direction-of-travel (trajectory layer) | aspirational | Fast-churn heading layer, freshness-stamped, human-confirmed | reasoning.md |
| CHAT-008 | Simple paid multi-domain chat (freemium + day-pass) | aspirational | Fast-lane paid chat: stateless signed entitlement token via Stripe guest-checkout | site-chatbot.md |
| CHAT-009 | Chat lanes (fast/slow/job) + warm-adapter maturation | aspirational | Fast/slow/job lane split; spawned-agent-to-warm-adapter maturation path | site-chatbot.md |
| IMG-029 | Lucide icon strategy and validator wiring | partial | Features grid uses Lucide webfont icons; validator written but not yet wired in. | imagery.md |
| IMP-035 | Bidirectional ratchet (trust can be lost) | aspirational | Feedback is two-directional: success accrues evidence toward graduation; repeated/severe failure drops the trust... | improvement-loop.md |
| FIX-028 | diagnose_read_repo_files action | deployed | Fetches plan's current file bodies; modify-404 is a hard error | fix-loop.md |
| CQ-016 | LLM fabrication classes in self-built site content | deployed | Fictional staff, fake taxonomies, nonexistent capabilities invented and later removed | content-quality.md |
| STY-017 | Section readiness model (planSection source tiers, spec resolver) | deployed | Field source tiers + required/fallback semantics decide defer-vs-carry | styling-render-pipeline.md |
| DOC-028 | EDIT-marker / -EDIT check-id convention | deployed | Fill-later blanks in seeded docs; -EDIT checks skipped until real selectors land | documentation-system.md |
| RAGK-004 | RAG best practices — filter-first, quality gating, token budget | aspirational | Filter by metadata before ranking; 20-30% context budget; nomic task prefixes | rag-knowledge-base.md |
| MCL-004 | Dispatch confirmation observability gap (Gap B / agent_dispatch_log) | aspirational | Fire-and-forget dispatch has no near-real-time failure signal; table proposed, not built | multicluster.md |
| DEV-041 | Development-guide gotcha: rebalance window after chassis restart | convention | Firing an orchestration within ~300s of a chassis restart risks a Kafka rebalance-window death. | development-guide.md |
| DEV-071 | MVP site builder pipeline (strategist → architect → content-creator → deployer) | superseded | First end-to-end production pipeline (boxing-tickets.com); superseded by the work-item pipeline. | development-guide.md |
| DOC-029 | Pilot PLAN seeding by SQL (dogfooding the format) | deployed | First tool PLAN hand-seeded before workflow wiring existed | documentation-system.md |
| DOC-063 | 2026-05-24 launcher build handoff (superseded Option A) | superseded | First training-launcher build handoff, carries two disproven claims | documentation-system.md |
| BLD-009 | site-work-orchestrator (unified build/maintenance over site_work_items) | deployed | First unified build/maintenance queue orchestrator, refined into build-dispatch-loop | build-pipeline.md |
| NAV-010 | Navigation tables (site_nav_groups / site_nav_items) | deployed | First-class nav model replacing scattered pages-table queries and the old cache | navigation.md |
| IMG-034 | generate_image action + image-generator adapter pipeline (legacy) | superseded | First-generation Stability-only image workflow, superseded by site_plan_imagery pipeline. | imagery.md |
| MDL-030 | Flywheel C Phase 2 — HTTP-job-server automation | abandoned | First-generation automation design, superseded before being built | model-infrastructure.md |
| DBG-048 | Early pipeline-failure triage priorities dropped by root-cause diagnosis | abandoned | First-pass symptom triage superseded within a day by deeper root-cause fixes | debugging.md |
| DBG-028 | Kafka topic-creation race self-heal (transient "Topic not yet on broker") | deployed | First-publish race on new per-spawn topics self-heals on retry; not a real fault | debugging.md |
| RSN-001 | Chain-of-thought prompt pattern catalog | unknown | Five CoT archetypes; no confirmed wiring into any agent prompt | reasoning.md |
| DIAG-031 | Loop-worthiness test (fix-loop intake doctrine) | deployed | Five criteria gate whether a bug enters the fix loop, including a mandatory cheap-query pre-check | diagnosis-loop.md |
| DBG-049 | Probe-project debugging-guide entries #24-#28 | deployed | Five field-earned checklist entries: runtime-path config, harness input, export, interfaces, UNIQUE | debugging.md |
| TL-006 | tool-game-* duplicate pages (T5) | unknown | Five page_type=tool, build_status=planned pages surfaced that duplicate the five existing games by name... | tool-lifecycle.md |
| CTS-030 | CSS section-colour model evolution (inheritance→hardcoded→painting) | superseded | Five-source archive history of how dark-section styling was hardened over ~a year | contracts-and-standards.md |
| DBG-056 | Stage-by-stage rebuild verification and the false-complete rule | deployed | Five-stage A-E method; complete + unchanged components = the old false-complete | debugging.md |
| ADP-006 | git-adapter as sole write credential holder | deployed | Fix-implementer never holds a GitHub write token; git-adapter does all writes | adapters.md |
| DOC-038 | doc_notes / travelling-docs integration boundary | deployed | Fix-loop persists terminal notes via another workstream's persist_note gate | documentation-system.md |
| IMG-019 | parseAspectRatio SDXL v1.0 whitelist snap fix | deployed | Fixed aspect-ratio snapping to SDXL's strict dimension whitelist, unblocking hero gen. | imagery.md |
| PBP-018 | render_mode derivation + LLM routing condition (migration 002) | deployed | Fixed check_render_mode to key off llm_field_specs, not the hardcoded render_mode column | page-build-pipeline.md |
| ONB-018 | Governed vocabularies and the hand-authored first constitution (prerequisites) | aspirational | Fixed concern/priority vocabularies must exist before conventions/intent agents run | onboarding-config.md |
| FIX-033 | Round-counting scope bug (correlation vs orchestration) | superseded | Fixed in source; one-cycle deploy gap via same-tag trap | fix-loop.md |
| MCL-014 | Shared topic pools (replace per-agent topics) | aspirational | Fixed partitioned pool topics route by agent-ID header instead of per-agent topics | multicluster.md |
| SPEC-006 | Structured design_intent from the classifier (palette/typography) | deployed | Fixed prose-buried hex colours so consumers read structured reference_values | site-spec-and-classifier.md |
| IMG-051 | Per-page hero resolver + flag_page_image_rebuild trigger (June fix) | deployed | Fixed site-wide hero overwrite + baked fallback + non-reresolving rerender, page scope. | imagery.md |
| LNK-016 | nav-link-fixer agent (template-anchor scope only) | deployed | Fixes #{{.slug}} anchors in templates; cannot reach hardcoded ContentData | link-management.md |
| WII-004 | item_key canonicalization + dedup namespace decisions | partial | Fixes item_key/item_type drift that caused double builds and silent dedup collisions | work-item-integrity.md |
| ASG-007 | Dynamic prompt improvement loop (Prompt Improvement Agent) | aspirational | Flag-for-improvement dispatches failing prompt to specialist, saves new version | agent-spawning-and-groups.md |
| SOC-011 | Games and daily-puzzle retention ecosystem | aspirational | Flagged expansion: Wordle-style daily games generated from scraping output | social-media.md |
| IMG-059 | image_source_unsatisfiable discovery check | deployed | Flags image fields sourced from an asset path nothing can supply; 0 flags = healthy. | imagery.md |
| CTS-029 | Thin-slice constitution (always-on rules) | deployed | Flat-file always-on rules doc; destined to become `standards` rows scope=constitution | contracts-and-standards.md |
| DEV-054 | thin-slice constitution (always-on rules doc) | deployed | Flat-file always-on rules pasted into every assembler/bundle output; may migrate to `standards` rows. | development-guide.md |
| FIX-032 | Fork isolation / NO FORK decision | superseded | Fork-isolation proposal raised then explicitly closed | fix-loop.md |
| SYS-071 | Standalone "probe-go" service (abandoned) | abandoned | Forked multi-vhost service rejected as too far from the website-building chassis | system-architecture.md |
| IMG-023 | Vision-capable LLM path (Phase 5) | aspirational | Foundational aiservice image-input support required before any imagery auditor can see images. | imagery.md |
| PEV-003 | Pragmatic Evolution Engine (portfolio build/learn/test/optimize) | abandoned | Founding 4-phase mission for a large-scale site portfolio; no cohort testing ever implemented | portfolio-evolution.md |
| DES-004 | Component-based headers replacing LLM-generated chrome | deployed | Founding decision that page chrome (header/footer/head) is never LLM-generated per page: tested templates... | design-composition.md |
| SYS-063 | Early long-term platform ambitions | aspirational | Founding roadmap: self-organising teams, marketplace, multi-tenant, cross-cluster | system-architecture.md |
| CTXA-001 | "Documentation is code": context-assembly tool and paid-service vision | partial | Founding thesis: assemble task-scoped bundles from docs+code, verify against ground truth, dogfood then sell | context-assembly.md |
| STY-034 | JS delivery paths & the js_snippets loader gap (historical) | partial | Four coexisting JS paths catalogued; loader gap later closed by STY-028 | styling-render-pipeline.md |
| DES-062 | Chrome linkage tangle: four overlapping header/footer default stores and the hardcoded dark fallback | partial | Four coexisting default stores for site chrome: `style_collections.header/footer_component_id` (the store... | design-composition.md |
| BIZ-010 | Audience-tuned elevator pitch variants (V1-V4 method) | deployed | Four ~30s pitches tuned per audience plus 10-second openers | business-strategy.md |
| DOC-009 | Cold-start documentation bundle practice (BUNDLE/HANDOFF/PLAN/RUNBOOK + cmd/bundle) | convention | Four-doc travelling set per investigation via cmd/bundle | documentation-system.md |
| IDEA-001 | Ideation method v0→v3 (staged, multi-model, web-verified pipeline) | deployed | Frame→generate→cut→verify→score→rank method behind idea.uk's paid reports | idea-product.md |
| RSN-006 | Step-type-aware prompt composition (altitude-aware) | aspirational | Framing gets full why-chain; generation collapses to a tether | reasoning.md |
| CTXA-004 | Altitude: step type decides what the bundle emphasises | deployed | Framing/implementation/debug steps get different mixes of intent, code, and runtime evidence | context-assembly.md |
| IDEA-007 | Free audience-check taster endpoint | deployed | Free rate-limited taster exposing method step 1; replaced voluntary-pay as the hook | idea-product.md |
| PLAN-026 | FAQ empty-items bug: duplicate content-surface planning (Defect 1) | deployed | Freeform + structured components both planned for same content; writer filled only one | site-plan-and-reconciler.md |
| ADO-010 | Fresh vs adoption entry paths converge on one cascade | deployed | Fresh-build and adoption both converge on needs_domain_research | adoption-pipeline.md |
| MDL-025 | Model-tiering by task ("the 3B problem") | aspirational | Frontier models for synthesis only; tiny specialised models for extraction | model-infrastructure.md |
| CTS-057 | Component creation contract (generator's embedded rulebook) | deployed | Full LLM-facing component-generation rulebook compiled from docs 003+018+schema v2 | contracts-and-standards.md |
| FIX-018 | Decision router (F2.3) | deployed | Full approved/revise/reframe/escalate router | fix-loop.md |
| CGV-019 | page_component_history full-snapshot content history | deployed | Full content_data snapshot before every write; rollback/audit substrate for edits | content-governance.md |
| TPI-002 | Topic amplifier / deep digger engine | abandoned | Full engineering design for topic collection/verification/dedup; no implementation trace | topic-intelligence.md |
| FTW-003 | Fine-tuning path (log→export→LoRA→GGUF→Ollama→swap) | partial | Full pipeline shape; last-mile production wiring explicitly still outstanding | finetuning-flywheel.md |
| AFFC-001 | Affiliate and first-party product commerce layer (schema built, resolver gap) | partial | Full products/affiliate schema shipped; resolver populating it is "a wired socket with no plug" | affiliate-commerce.md |
| CTS-031 | Component Quality Contract (scoring formula) | abandoned | Full quality-scoring contract vanished from docs v6→v7; residual fields still in live JSON | contracts-and-standards.md |
| PBP-022 | Two re-render paths + assemble-only rerender distinction | deployed | Full rebuild vs light no-LLM re-render vs assemble-only reassembly are three distinct operations | page-build-pipeline.md |
| STY-025 | Interactive-section clobber + interactivity-aware save guard | partial | Full rebuilds silently discard interactive tools stored only as rendered_html | styling-render-pipeline.md |
| TRF-012 | Ingest validation contract | aspirational | Full spec for what the collector must enforce (shape checks, dedupe, NFC); not yet enabled | traffic-analytics.md |
| SPEC-001 | Dream spec / gap analysis / feasibility - one spec, not two | aspirational | Full spec is the dream; per-item status makes gap analysis mechanical | site-spec-and-classifier.md |
| WDS-001 | Work-item state machine (detected → triaged → claimed → complete/failed) + site-exclusion by stuck claim | deployed | Full state machine plus the common "one stuck claim blocks the whole site" failure mode | work-dispatch.md |
| BATCH-003 | Dispatch loop & detected→triaged→claimed state machine | deployed | Full work-item state chain from detection through claim to completion | batch-processing.md |
| CQ-004 | Recovery playbook for stranded dependents (Route A vs Route B) | deployed | Full writer rebuild vs re-key + scoped re-render to recover pages after a contract change | content-quality.md |
| DES-028 | Layout: tool-first-landing | deployed | Full-container (up to 1400px) tool-dominated landing page where "the tool IS the page" — defining primitive... | design-composition.md |
| BIZ-027 | UK-sovereign stack exploration (deferred) | aspirational | Future fully-UK-hosted compute+storage+model stack; explicitly deferred by owner | business-strategy.md |
| ADO-007 | Pattern extraction, code-as-reference, and RAG-fed generation | aspirational | Future pattern-extraction-agent mining research into reusable specs | adoption-pipeline.md |
| BLD-002 | Three coexisting builder generations + the work-item relay spine (baton/hop model) | deployed | GEN1/2/3 builder archaeology; settled spine is the site_work_items baton relay | build-pipeline.md |
| DOC-023 | NOTES category taxonomy | deployed | GIN-queryable tag vocabulary extending 037's taxonomy | documentation-system.md |
| STG-002 | Hostile-VM threat model for the training data plane | deployed | GPU box holds no B2 key/DB access, only time-limited presigned URLs | storage-architecture.md |
| VET-011 | thunder-reaper + cost gate (spend backstop) | deployed | GPU instance uptime cap + spend gate; miscategorized here, belongs with Thunder infra | vet-med-pricing.md |
| DIAG-014 | diagnose_ro role and pooler-aware read-only enforcement | partial | GRANT-only SELECT role for the CLI harness; pgbouncer doctrine — enforce by GRANT, never session settings | diagnosis-loop.md |
| IMP-030 | Audit gap-finding routing fix (existing-page gaps → needs_content_page) | deployed | Gap findings on existing pages were being routed to content_rewrite (edits, not rebuilds), causing... | improvement-loop.md |
| PLAN-028 | Stale site_plan: gap-planned pages never written back (Concern 2) | aspirational | Gap-added pages never appended back to site_specs.site_plan, causing drift | site-plan-and-reconciler.md |
| DOC-052 | Travelling-docs pattern (runbook = plan, notes = history, handoff = session) | deployed | General cross-project framing of the runbook/notes/handoff triad | documentation-system.md |
| IMG-061 | Orphaned generated assets (component consumes nothing) | partial | Generated imagery with no consuming component slot, or stale post-replan assets. | imagery.md |
| SYS-088 | Human-readable orchestration and correlation names | deployed | Generated readable names alongside UUIDs for narrative-style debugging | system-architecture.md |
| PBP-009 | page-content-writer (task specialist, no persistence) | deployed | Generates content only; persistence/deploy live in the page-build-handler wrapper | page-build-pipeline.md |
| TLIB-004 | Component-creator agent (observed-pattern section components) — deployment specifics | deployed | Generates new section component templates (hero, feature-grid, etc. — distinct from tool-generator) when a page... | tool-library.md |
| TLIB-011 | component-creator (LLM component template generation) + CSS variable naming contract | deployed | Generates reusable HTML component templates from section-type descriptions, storing them in content_components... | tool-library.md |
| VET-010 | Configurable med price JSON export to site repos | deployed | Generic export action serves many sites via config, commits JSON into site git repos | vet-med-pricing.md |
| DEV-007 | Field-name collision via the nested-source loop (required AND optional) | deployed | Generic field names can silently bind to the wrong nested source, including required fields. | development-guide.md |
| LNK-003 | Hero CTA brochure-default defect (text↔destination mismatch) | deployed | Generic hero schemas defaulted every CTA to /contact.html and phantom /services.html | link-management.md |
| SYS-065 | relationships table — first-class entity relationships | partial | Generic relationship entity modelled on links, later earmarked for semantic page links | system-architecture.md |
| CGV-024 | maintenance_queue + claim/complete/fail functions | partial | Generic site-maintenance work queue with SKIP LOCKED claim/complete/fail functions | content-governance.md |
| BIZ-001 | Platform mission: best possible site per domain via one unified pipeline | partial | Given any domain, produce the best site end-to-end; revenue model shapes the site | business-strategy.md |
| STY-032 | CSS responsibility barrier + colour inheritance model | deployed | Global CSS owns colour/typography; component CSS owns layout only | styling-render-pipeline.md |
| CTS-006 | String-value naming convention (snake identifiers, kebab data) | deployed | Go-identifier values snake_case, data-shaped values kebab-case; migration 051 applied | contracts-and-standards.md |
| ADO-020 | Two-stage -> three-stage adoption processing (historical evolution) | superseded | Go-only design-fingerprint stage inserted ahead of LLM classification | adoption-pipeline.md |
| FTW-030 | run.sh launch chain + RUN_SH markers | deployed | Grep-able marker protocol; RUN_SH_DONE now implies durable upload | finetuning-flywheel.md |
| DEV-076 | output_field / input_fields group-memory data mapping contract | deployed | Group-agents-era plumbing: call_agent's output_field names the group-memory key for the child's result. | development-guide.md |
| SPEC-009 | Guide as first-class page_type (classifier vocabulary + canonical URLs) | deployed | Guide added to classifier enum, retyped, URLs migrated to canonical form | site-spec-and-classifier.md |
| FIX-036 | Council roster expansion vision | aspirational | Guidelines/reuse/bug-historian/compliance bench never built | fix-loop.md |
| ADO-015 | guide as a first-class page_type (adoption classifier) | deployed | Guides folded into blog-post then given own page_type + canonical URL | adoption-pipeline.md |
| CTXA-014 | analyse_repo_local in-process analysis and the stale-index incident | deployed | HEAD-ref resolution silently indexed a year-old commit; fixed by explicit-ref in-process tarball analysis | context-assembly.md |
| DIAG-037 | Stale-corpus class: HEAD pinning, explicit refs, CI-triggered indexing doctrine | partial | HEAD/latest pins silently track ancient artefacts; doctrine adopted for explicit refs, CI-indexing still queued | diagnosis-loop.md |
| CGV-023 | Content review flow with rejection -> needs_attention | deployed | HITL/auto-eval gate; rejected pages marked needs_attention and queued for maintenance | content-governance.md |
| FIX-045 | SEED_first_writestep_diagnosis / seeded-bug strategy | deployed | Hand-authored CONFIRMED row exercised the real write chain | fix-loop.md |
| DEV-018 | Work-item manual-crafting discipline (real shapes, truthful provenance, never-guess) | convention | Hand-inserted site_work_items must mirror real rows; never guess specs, URLs, or paths. | development-guide.md |
| DBI-016 | Auth database provisioning | deployed | Hand-provisioned auth_db/auth_user; file preserves a live credential (hygiene finding) | database-and-infrastructure.md |
| ONB-019 | build-briefing-agent (spec-reading briefing) | deployed | Handler answers briefing questionnaire autonomously from site_specs, no human | onboarding-config.md |
| DEV-010 | Specialist vs handler: the persistence boundary | deployed | Handlers must persist their own outputs; specialists-as-handlers need a save/deploy wrapper. | development-guide.md |
| CTS-016 | Handler dispatch input-path contract (input_data.spec.*) | deployed | Handlers must read spec via input_data.spec, not top-level flattening; rediscovered twice | contracts-and-standards.md |
| EMAIL-002 | Transactional email sending realities (587-only, relay filtering, SES + DKIM) | deployed | Hard-won SMTP truths: 587-only, MailChannels blocks, dedicated SES sender adopted | email-infrastructure.md |
| LNK-007 | Layer 1b header/footer phantom fix (shared site components) | deployed | Hardcoded ContentData in render_site_components fixed at the Go source | link-management.md |
| IMG-042 | Header logo resolution from plan imagery | deployed | Header component fixed to resolve locked logo via site_plan_imagery, not dead sites.logo_url. | imagery.md |
| CTS-028 | Chrome templates must be variable-driven | aspirational | Header/footer LLM-hardcode links; pre-store hardcoded-link gate designed, not built | contracts-and-standards.md |
| SYS-015 | Four overlapping chrome default stores | partial | Header/footer defaults split across 4 stores; intended chain deliberately left unrepaired | system-architecture.md |
| SCH-004 | Work-item claim/retry behaviour and the claim-timeout class | deployed | Heavy builds collide with claim durations producing retried-then-complete items | scheduler-and-tasks.md |
| CTS-055 | Section resolvers override content_data on every render | deployed | Hero image/static fields re-resolve on every render, ignoring stored instance edits | contracts-and-standards.md |
| DES-016 | Layout: brochure-bold | deployed | High-energy conversion variant of brochure-formal — tall hero, gradient accents, display-bold typography,... | design-composition.md |
| FTW-023 | Fine-tuning candidate selection/prioritisation | aspirational | High-volume structured-JSON agents ranked as swap candidates | finetuning-flywheel.md |
| MDL-019 | GPU/AI-endpoint scheduling design evolution (superseded) | superseded | Historical four-option debate resolved into the single ai_endpoint_health table | model-infrastructure.md |
| PLAN-020 | site_plan page-role enum naming (underscore → hyphen; index → landing) | superseded | Historical rename of the role vocabulary to kebab-case, homepage role to landing | site-plan-and-reconciler.md |
| ADP-011 | thunder-adapter — GPU provisioning adapter | deployed | Holds Thunder/B2 creds, provisions ephemeral GPU VMs, verified end-to-end 2026-05-22 | adapters.md |
| SYS-032 | Page content-creation build pipeline trace | deployed | Hop-by-hop trace from load_page_record through SavePageSectionsAction | system-architecture.md |
| FTW-032 | Checkpoint & final-adapter durability via presigned PUT manifest | partial | Hostile-VM-safe upload via pre-minted URLs; O(K²) loop retired via batch presign | finetuning-flywheel.md |
| DEV-022 | Sub-agent modelling conventions (agent_definitions row shape) | deployed | How a called sub-agent's row/workflow/topics/seed-SQL should be modelled, per research-agent. | development-guide.md |
| DES-064 | Section→component resolution: direct-function Path 1 vs scoring selector Path 2 | deployed | How a planned section becomes an actual rendered component: Path 1 matches the section name directly against... | design-composition.md |
| DIAG-036 | code_symbols retrieval as used by the diagnosis loop | deployed | How the loop seeds iteration-1 scope from the code_symbols index via lookup_code_symbols | diagnosis-loop.md |
| FTW-025 | Eval gate before promotion | partial | Human deployment_decision required; also the integrity boundary for uploads | finetuning-flywheel.md |
| HITL-008 | Human change-request work items | deployed | Human-submitted edits enter the same priority-ordered work queue as agent items | hitl.md |
| SQ-002 | Site-chrome gap hypothesis (relay path lacks chrome rendering) | partial | Hypothesis: relay build path never renders nav/header/footer chrome | site-quality.md |
| TRF-018 | Global bot-IP blocklist (Thread D) | aspirational | Idea to block illegitimate-crawler IPs globally across all boxes from the access-digest rollup | traffic-analytics.md |
| ADP-007 | git-adapter new actions (create_branch, create_pull_request) | deployed | Idempotent branch creation and PR-as-human-review-terminal actions | adapters.md |
| SYS-051 | Sites contact-identity denormalisation | deployed | Identity/contact fields promoted from content_data JSONB to first-class columns | system-architecture.md |
| TP-005 | deploy_page files_field contract (co-located JS must ship) | deployed | If page deploys use content_field (HTML only), component JS (/tools/assets/*.js) is silently dropped — news... | tool-pipeline.md |
| DGH-005 | Chassis build/deploy practice (local Makefile builds) | deployed | Images build from local tree, decoupled from commits; verify against the running pod | deployment-github.md |
| CLC-003 | F1 field-contract guard (reject regens that rename/drop retained fields) | deployed | In StoreGeneratedComponentAction's Layer-1 validation, on isRegeneration the guard diffs old vs new... | component-lifecycle.md |
| MDL-021 | Code-context retrieval infrastructure (analyser adapter) | deployed | In-cluster code indexing into a pgvector code_symbols table; found stale | model-infrastructure.md |
| SYS-061 | Child-orchestration timeout monitor | partial | In-memory per-child timeout goroutine; pod-restart recovery was a known gap | system-architecture.md |
| DYN-006 | Tool builder tiers (static / dynamic / application) | partial | Interactive functionality classified by creation risk; matured to tool-pipeline | dynamic-applications.md |
| DEV-025 | Reply-topic derivation rules (own topic vs parent topic) | deployed | Intermediate calls await on the child's own responses topic; only the final notify uses the parent's. | development-guide.md |
| DBG-069 | Launcher reply-topic own-vs-parent derivation (Decision D4) | deployed | Intermediate replies must use own ResponsesTopic, never parent's | debugging.md |
| SPEC-005 | Superseding a spec doesn't undo installed artefacts (re-queue rule) | convention | Invalidating a spec must also queue the re-run work item | site-spec-and-classifier.md |
| BLD-001 | Builder route method: map what exists before building (§B0 census) | convention | Inventory of ~147 agent types against problem-statement capabilities before building anything | build-pipeline.md |
| WDS-007 | Two-strike rule for work items (dedup/anti-churn) | deployed | Item_key with 2+ terminal attempts in 7 days inserts as unresolved rather than re-dispatching | work-dispatch.md |
| BATCH-002 | Work item lifecycle (blocking, unblocking, unresolved) | deployed | Items blocked three ways; unresolved mechanism suppresses rapid re-emission | batch-processing.md |
| DIAG-004 | Four convergence guards plus engine-level failsafes | deployed | Iteration cap, scope-not-narrowing, evidence-not-growing, hypothesis-thrash guards plus timeout/fuel caps | diagnosis-loop.md |
| NEWS-011 | News rendering three-layer architecture (data/behaviour/structure+style) | deployed | JSON data, component JS, and template/CSS deploy independently, joining only in-browser | news-feed-pipeline.md |
| SNAP-001 | Site snapshots: point-in-time capture and revert | deployed | JSONB full-site-state snapshot/revert functions, migration 085, used in production | site-snapshots-and-revert.md |
| FTW-021 | Claude-as-judge anonymised A/B + self-recognition bias | deployed | Judge design controls for position and self-recognition bias, empirically observed | finetuning-flywheel.md |
| LNK-019 | Links agent family (algorithmic, no-LLM link health) | aspirational | Judgment-free crawler/validator/redirect-manager family, still unimplemented | link-management.md |
| IMG-058 | Image-role alias resolver + authoritative overlay (I0 rewrite) | deployed | July rewrite unifying 3 incompatible hero-resolution patterns via a shared alias table. | imagery.md |
| CGV-022 | Legal content agent + legal constraint rules | aspirational | Jurisdiction-aware legal pages and machine-readable disclaimer/forbidden-phrase rules | content-governance.md |
| ADP-015 | Firecrawl scraping adapter and actions | deployed | Kafka adapter exposing scrape/crawl/extract; v2 owns screenshot/image S3 copies | adapters.md |
| CTXA-013 | Analyser adapter: in-cluster polyglot parsing service | deployed | Kafka adapter fetching a repo tarball read-only and parsing it via a pluggable per-language Analyser seam | context-assembly.md |
| MDL-011 | Thunder Compute adapter (provision/decommission lifecycle) | deployed | Kafka adapter wrapping Thunder API; provision/decommission mechanics + API gotchas | model-infrastructure.md |
| CTXE-005 | Bundle size doctrine: "a large bundle is a smell, not a goal" | convention | Keep working bundles under ~200K tokens; fix oversized bundles with narrower selection, not a bigger window | context-engineering-principles.md |
| DOC-048 | thin_versions (cmd/thin_versions) | deployed | Keeps newest N versions per document subject, archives the rest | documentation-system.md |
| SCH-018 | P4 off-box collection (intent_events + CollectIntentEventsAction) | partial | Key-gated HTTPS intent pull with structural idempotency via engine_event_id | scheduler-and-tasks.md |
| CTS-052 | /events export endpoint (P4 collector interface) | deployed | Key-gated NDJSON event stream with since/host/limit params, lock-free by design | contracts-and-standards.md |
| TRF-008 | /events export endpoint and checkpoint contract | deployed | Key-gated NDJSON export with strictly-after since param, lock-free, duplicate-free pulls | traffic-analytics.md |
| CTS-051 | /stats endpoint + INTERNAL_API_KEY | deployed | Key-gated per-host stats summary via X-Internal-Key header | contracts-and-standards.md |
| IMP-006 | Three-way audit-finding classification (bug/recommendation/gap) — recurring proposal, still unbuilt | aspirational | LLM auditors mix factually-broken bugs with subjective opinions, but the pipeline auto-fixes both uniformly as if... | improvement-loop.md |
| CGV-015 | Blog/content planning agents (blog-content-planner, content-gap-planner, internal-linker) | deployed | LLM planners turning content gaps into pages, sections, or internal links | content-governance.md |
| NEWS-002 | Feed triage: relevance + credibility + source-attribution provenance | partial | LLM scores relevance/credibility/attribution; credibility field never actually populated | news-feed-pipeline.md |
| CTS-045 | CSS variable naming convention (--color-*) + STRICT RULE | deployed | LLM was emitting nonexistent --primary-color names; prompt now enforces real variable names | contracts-and-standards.md |
| PLAN-024 | Lazy per-page brief generation via build_page_brief step (abandoned; replaced by Go renderer) | abandoned | LLM-generated lazy briefs replaced by a deterministic Go brief renderer | site-plan-and-reconciler.md |
| IMG-027 | LLM-generated SVG icon path (sleeper option) | aspirational | LLM-written SVG icons retained as a possible future replacement for raster icon gen. | imagery.md |
| FTW-036 | knowledge_base RAG store + Flywheel B verification | deployed | Lane-B deployment/verification record; full mechanism lives in RAGK-001 | finetuning-flywheel.md |
| ASS-001 | Agent swarm simulation ideas (never built — hierarchical/fractal use-case brainstorm) | aspirational | Large brainstormed catalogue of 1M-agent hierarchical/fractal use cases | agent-swarm-simulations.md |
| SOC-009 | Content-first launch strategy for Spark (vonc.com as destination) | partial | Launch a content destination, not a social platform; provocations as SEO pages | social-media.md |
| ADO-001 | Infrastructure three layers (core/client delivery/framework builder) | partial | Layer 1 (factory) built, Layer 2 (client delivery) planned, Layer 3 future | adoption-pipeline.md |
| IMP-001 | QA three-layer architecture + concrete audit agent hierarchy | deployed | Layer 1 = structural checks (algorithmic, every cycle); Layer 2 = group LLM audits (shared context, ONE LLM call... | improvement-loop.md |
| CTXE-006 | Reuse-check retrieval pipeline design | aspirational | Layered catalog→lexical/structural→embeddings→rerank design tuned for recall over precision | context-engineering-principles.md |
| DES-058 | Dark Section Variable Contract / buildSectionDefaults renderer behaviour | partial | Layout templates must NOT declare `--section-*` defaults on section containers themselves; a Go renderer... | design-composition.md |
| DES-038 | Theme/layout library growth: fork-with-review gate + design-asset lineage columns | partial | Layouts are a curated shared grammar — no auto-generated bespoke layout per site. | design-composition.md |
| ADO-032 | Adopting existing external sites ("Adopt" workflow), legacy precursor | superseded | Learn loop against existing sites with component-match confidence scores | adoption-pipeline.md |
| DBG-045 | Kafka per-spawn response-topic partition race (adapter reply lost) | partial | LeastBytes balancer picks out-of-range partition on fresh per-spawn topics | debugging.md |
| IMG-052 | Legacy site-level hero_url shadow (content_data last-write-wins) | deployed | Legacy content_data hero_url still shadows per-page heroes with one site-wide image. | imagery.md |
| IMG-036 | pageflow-builder retirement | superseded | Legacy monolithic site builder deliberately left un-extended; architecture moved on. | imagery.md |
| ADM-010 | AI Persona Platform public API | partial | Legacy v1 REST surface from the "AI personas" productisation era | admin-dashboard-and-api.md |
| IMG-041 | Manual brand-asset commit workaround (derivation gap) | partial | Leopardess site hand-derived favicon/OG and committed via a standalone shell script. | imagery.md |
| TLIB-001 | Fork-on-deploy tool ownership model | deployed | Library tools are canonical rows (component_level='tool', forked_from IS NULL) — blueprints never referenced... | tool-library.md |
| DYN-009 | js_snippets library + render_js_snippets_for_site + site-asset-renderer | deployed | Library-wide JS behaviours bundled per site by applies_to overlap | dynamic-applications.md |
| DES-060 | Hazard-class vs band-class self-declarer split; is_dark_section demoted to metadata | convention | Library-wide diagnosis, generation-4 of the section-contrast arc: of 84 active section components, 37... | design-composition.md |
| DES-031 | Layout: social-lobby | partial | Light, colour-forward social-platform layout built around a room/lobby metaphor. | design-composition.md |
| CTS-041 | Query-resolver list components (pages_where_type) | deployed | List components resolve items dynamically by page_type; no template change on page add | contracts-and-standards.md |
| FIX-038 | Guardian veto surfacing an architecture-level fix | deployed | Live instance: guardian caught a disguised architecture change | fix-loop.md |
| DIAG-034 | Verdict-quality wrinkles + measured-dead code-retrieval channel | partial | Live runs show the code lookup channel contributes almost nothing; two distinct verdict-quality defects found | diagnosis-loop.md |
| FIX-020 | Schema hint for reviewers (F2.3b(a)) | deployed | Live schema hint fixed hallucinated-column check failures | fix-loop.md |
| IDEA-013 | Real-door streaming progress page + programmatic refund endpoint | aspirational | Live-progress UX and Stripe refund API both designed, not built; refunds are manual | idea-product.md |
| CTS-034 | Chassis conventions verified (text+CHECK, deleted_at) | deployed | Live-schema verification pass corrected contract docs to match reality | contracts-and-standards.md |
| SYS-045 | Architectural tensions catalogue | partial | Living catalogue of genre-level design tensions (infer-and-repair; page identity) | system-architecture.md |
| LQT-001 | Model quality assessment: local 70B comparable for some tasks | deployed | Llama 70B near-parity with Claude on classification/content; Mistral weak | llm-quality-testing.md |
| PLAN-029 | site_plan as authoritative build source, overwriting pages.sections | deployed | Loader syncs plan sections back into pages.sections every build; fixes must target the plan | site-plan-and-reconciler.md |
| MCL-009 | Cross-cluster Postgres reachability strategy (Option C) | aspirational | Local PgBouncer per remote cluster tunnels back to primary Postgres | multicluster.md |
| DEV-075 | Aggregation patterns (aggregate_data, aggregator agent) | partial | Local aggregate_data broke on verbose child state; redesigned as a spawned aggregator agent. | development-guide.md |
| LOCK-002 | Lock semantics: hard gate discovery, soft gate execution, read-only rerender | deployed | Lock means human-controls, not read-only; discovery skips locked, execution doesn't | locks.md |
| IMG-039 | Logo permanence: generate → human-approve → lock (D5) | deployed | Logo generated once, human-approved, then locked; favicon/OG derive from it, never regen. | imagery.md |
| DEV-016 | pgbouncer per-batch transaction discipline | deployed | Long transactions through pgbouncer are fragile; bulk work must commit per small batch. | development-guide.md |
| MCL-015 | Worker pool architecture (replace per-agent Jobs) | aspirational | Long-running pods run many agent workflows as goroutines instead of per-agent Jobs | multicluster.md |
| DES-040 | Visual identity library and effects library (composable design assets) — aspirational | aspirational | Longer-term plan for two accumulating libraries: a visual identity library of palettes/typography/effects... | design-composition.md |
| CTXA-017 | code_symbols repo-label symmetry (shared owner/repo resolver) | deployed | Lookup queried a bare repo name against composed owner/repo rows and found nothing; fixed with one shared resolver | context-assembly.md |
| DIAG-017 | Deterministic scaffold / model-only-verdict split | deployed | Loop control/guards are pure tested Go; only the verdict judgement is model-dependent, its own workflow step | diagnosis-loop.md |
| DIAG-018 | Falsification-first evaluation gate (scaffold correct ≠ reasons well) | deployed | Loop must reproduce reversals and abstain on known bugs before being trusted, not just pass scaffold tests | diagnosis-loop.md |
| FIX-010 | Standing hypothesis refuted (reconcile_site_plan routing table) | superseded | Loop refused to confirm the wrong-file hypothesis | fix-loop.md |
| SYS-044 | Loop mechanisms | deployed | Loop steps expand into N×M dynamic workflow steps, not Go for-loops | system-architecture.md |
| DEV-012 | Loop mechanisms: dynamic workflow expansion | deployed | Loops inject N×M steps at runtime via a coordinator-side expansion handler; never nest loops. | development-guide.md |
| RSH-001 | Dual-signal self-heal on missing spec dependency | deployed | Loud error log + queued recovery work item; two-strike rule caps retries | resilience-self-heal.md |
| TP-007 | Toolchain validator + repo read/search (net-new for a self-coding pipeline) | aspirational | Low-regret net-new pieces identified for a hypothetical self-coding pipeline: a toolchain validator giving... | tool-pipeline.md |
| CGV-027 | Privacy posture (no cookies/JS/IP; UK GDPR/PECR) | deployed | Low-risk privacy stance baked into the traffic-probe engine and pages | content-governance.md |
| IMG-010 | Hero-variant routing through image-build-handler (Phase 2E) | deployed | Made hero_<page> variants routable via a new classification and deploy-path branch. | imagery.md |
| CQ-001 | Content-quality defect catalogue (gamesdesign.co.uk) | partial | Maintained catalogue of hero-CTA, brand-suffix, empty-footer/description defects | content-quality.md |
| IDEA-003 | Capability watchlist + real-world event-window watchlist | aspirational | Maintained lists of AI capabilities and scheme deadlines feeding re-runs; unbuilt as workflow | idea-product.md |
| DEV-001 | STEP ZERO — reuse-before-create discipline | convention | Mandatory pre-flight search of agent_definitions/registry/code before creating anything new. | development-guide.md |
| FTW-011 | Flywheel C training pipeline (scripts 00-03) | deployed | Manual Unsloth QLoRA scripts with a smoke-gates-full discipline | finetuning-flywheel.md |
| MCL-007 | Cross-cluster KafkaUser + secret replication pattern | aspirational | Manual kubectl-copied KafkaUser secret authenticates a remote cluster's agents | multicluster.md |
| DEV-080 | Data-flow verification matrix practice | deployed | Manual per-step config-vs-implementation trace practice; ancestor of automated contract validation. | development-guide.md |
| SPEC-022 | Roadmap phase advancement and automated strategic review | aspirational | Manual phase advancement now; automated strategy-review loop deferred | site-spec-and-classifier.md |
| TL-029 | Component-creator invocation contract (dual placement + quote-free description) | partial | Manually invoking component-creator (spawn+call) must satisfy BOTH the input_contract (top-level required fields... | tool-lifecycle.md |
| SPEC-018 | Chassis-native idea engine (Phase D / Layer 4) | aspirational | Mapped-but-unbuilt plan to express idea-generation as chassis actions | site-spec-and-classifier.md |
| MCL-013 | Multi-cluster scaling tiers (10K/100K/1M agents) | aspirational | Maps each agent-count tier to its bottleneck and the one architectural fix | multicluster.md |
| MDL-033 | run.sh RUN_SH markers + set -e durability hard-gate | deployed | Marker protocol lets DONE imply "trained AND uploaded" | model-infrastructure.md |
| BIZ-016 | Portfolio/use-case spec seeds (ai-agent-orchestration.com) | deployed | Marketing case-study data doubling as a platform capability inventory | business-strategy.md |
| LOCK-003 | Site-level lock (sites.locked_at) | deployed | Master switch stopping all automated agent activity on a site | locks.md |
| FTW-022 | iter_0 verdict: shippable for low-stakes | convention | Matches Claude on JSON/schema; voice fidelity is the main iter_1 lever | finetuning-flywheel.md |
| CH-002 | Vertical profile registry (generic-words/keywords/suffixes per industry) | deployed | Matching heuristics live in a Go registry keyed by vertical_slug — verticals are config | companies-house-enrichment.md |
| DOC-004 | Running-notes checkpoint journal + distilled HANDOFF discipline (idea.uk) | convention | Memory-off journal + HANDOFF pattern, run on main + sub-thread | documentation-system.md |
| FIX-031 | PR as human terminal / nothing merges itself | convention | Merge is permanently human across the whole fix-loop | fix-loop.md |
| ONB-006 | Inference quality scales with codebase quality — surface uncertainty | aspirational | Messy repos yield confident-but-bad convention inference; surface as questions | onboarding-config.md |
| TRF-017 | Traffic-claim verification and the bot-vs-human verdict method | deployed | Method for testing marketplace visit claims against beacon and access-log ground truth | traffic-analytics.md |
| ADO-017 | Adoption resume logic (never built) | abandoned | Mid-workflow resume plumbing exists but no subscriber; re-crawl is the answer | adoption-pipeline.md |
| DES-035 | webdesign-agent post-merge loop bug and generate_css stuck mystery | unknown | Migration 010 left every non-fork path out of `deploy_css` looping back to `generate_css`... | design-composition.md |
| PLAN-036 | site_plan_pages schema repair (plan-domain drift) | deployed | Migration reconciles two schema drafts' columns, drops orphan site_plan_partials | site-plan-and-reconciler.md |
| AGOV-009 | Thin vertical slice before six-contract infrastructure | deployed | Minimal bundle harness built and used before any contract shipped | autonomy-governance.md |
| DBG-054 | Isolated build test methodology (throwaway test-page pattern) | deployed | Minimal test page through the full build path attributes a bug to one pipeline layer | debugging.md |
| DES-019 | Layout: utility-tool | deployed | Minimal-chrome layout where "the tool is the reason" — narrowest container (800px), compact header, single tool... | design-composition.md |
| PLAN-010 | page_type vocabulary gap forcing game→tool re-type (Gap B) | unknown | Missing 'game' page-type forces re-typing to tool, duplicating pages | site-plan-and-reconciler.md |
| DBG-058 | Spawn-consumed columns lesson: seeds must copy infra columns from a live donor | deployed | Missing command column boots generic entrypoint; dispatcher's call goes unheard | debugging.md |
| CTS-010 | Site component linkage contract (slot_name↔function) | deployed | Missing component_id link falls to generic lookup then hardcoded fallback header | contracts-and-standards.md |
| VONC-002 | Phase-3 provocation pipeline (automated provocations.json emission) | aspirational | Missing daily generator for provocations.json; still hand-committed as of 2026-07-11 | vonc.md |
| DBG-036 | Env-prefix trap: VAR=x on its own line never reaches the child process | deployed | Missing export/same-line prefix silently uses defaults; banner-tell mitigation | debugging.md |
| FTW-037 | Nomic task prefixes load-bearing | deployed | Missing search_query/search_document prefixes silently broke ranking 5x | finetuning-flywheel.md |
| BIZ-004 | Payable-differentiator / moat framework (asset × AI × paying audience) | deployed | Model is never the moat; a hard-to-reproduce asset + AI for a paying audience is | business-strategy.md |
| DIAG-003 | Verdict cite-or-abstain contract + wire format seam | deployed | Model must CONFIRM/REFUTE/UNVERIFIABLE with citations; verdict_wire.go parses it, fail-safe to UNVERIFIABLE | diagnosis-loop.md |
| DIAG-016 | EXPLAIN pre-flight size guard on data requests | deployed | Model queries are EXPLAIN-planned before execution and skipped with feedback if estimated rows are too big | diagnosis-loop.md |
| IMG-028 | Diffusion transparency abandoned → flat-grey chip icons | deployed | Models can't produce real alpha; icons locked to flat grey background inside a CSS chip. | imagery.md |
| FTW-035 | Monitor enablement gate: DONE must mean durable | partial | Monitor stays disabled until upload path proven, to avoid destroying adapters | finetuning-flywheel.md |
| DBG-027 | Scheduler-fired chassis-resident agents report owner_agent_type='generic' | deployed | Monitoring filters must key on collected_data config.agent_type, not owner_agent_type | debugging.md |
| VET-005 | LLM-driven content_features recommendation | aspirational | Moves news/tools/guides decision from hardcoded Go map into classifier LLM prompt | vet-med-pricing.md |
| FIX-016 | Hard-veto flag at multiple scopes — early design | superseded | Multi-scope veto design narrowed to single guardian flag | fix-loop.md |
| DBI-021 | setup.sh box provisioning script | deployed | Multi-vhost box installer: nginx, certbot, systemd, hardening, deploy hook | database-and-infrastructure.md |
| SYS-009 | business-intel shared-pod pattern | deployed | Multiple agent defs share one static pod via message routing; ai_service must live in step config | system-architecture.md |
| DBI-004 | Three-database architecture | deployed | MySQL auth + PostgreSQL clients_db + PostgreSQL templates_db via pgbouncer | database-and-infrastructure.md |
| DOC-016 | The dangling-doc prevention rule | deployed | NOTES subject must reference an artifact the agent actually owns | documentation-system.md |
| REB-004 | Auto-escalation: empty content_data → needs_page writer rebuild | deployed | NULL content_data escalates a page to a full writer rebuild instead of carrying garbage | rebuild-cascade.md |
| BIZ-017 | AI persona team and departments marketing model | deployed | Named AI personas (Archivist/Sentinel/Quartermaster) + 8-dept/70+ agent structure | business-strategy.md |
| SPEC-004 | Spec aspect ownership and read-and-extend (anti-silent-overwrite) | deployed | Named owners per aspect; classifier reads-and-extends adoption output | site-spec-and-classifier.md |
| ASG-002 | Agent groups (reusable multi-agent teams) | deployed | Named versioned team definitions; spawn_group instantiates and starts workflow | agent-spawning-and-groups.md |
| FTW-006 | training_exports Postgres schema | deployed | Named, versioned snapshot datasets in Postgres, not ephemeral JSONL files | finetuning-flywheel.md |
| IMP-036 | Fixer agents: color-variable-fixer, site-component-linker, component-template-fixer, css-patch-agent | deployed | Narrow algorithmic/LLM fixers dispatched from the queue: color-variable-fixer replaces hardcoded hex in component... | improvement-loop.md |
| RSN-007 | Checker model (single-axis parallel checkers) | aspirational | Narrow parallel checkers reconciled by singular arbitration | reasoning.md |
| VET-004 | vetcomparison.uk V1 rebuild scope | aspirational | Narrow relaunch: medicine search, vet directory, news, guides; no price-panel yet | vet-med-pricing.md |
| FTW-014 | GPU environment version pinning (cu124 stack) | deployed | Narrow torch/transformers/flash-attn pin set required for training to work | finetuning-flywheel.md |
| DIAG-006 | Named-scope guard vs capped call-graph expansion | deployed | Narrowing guard compares model-named scope only; call-graph expansion capped at 18 for the gather | diagnosis-loop.md |
| CASE-005 | Dartsonline guides defect (benchmark bug, causes A/B/C) | deployed | Nav link to a blank page kept live deliberately as a fixloop benchmark | site-case-studies.md |
| NAV-012 | Header nav from pages.in_header + nav-label hygiene | deployed | Nav membership is a data flag; label hygiene is a companion defect | navigation.md |
| DES-055 | Three-per-row no-orphan grid rule as a content fix | convention | Neither a global `repeat(3,1fr)` nor a per-component `auto-fit,minmax()` avoids orphan/stretched last cards in... | design-composition.md |
| DOC-066 | docs019 working/main snapshot bundle (duplicate early-draft staging copy) | superseded | Nested archive-of-archive with zero unique content vs live docs | documentation-system.md |
| DOC-068 | `subject_type='component'`: travelling docs for a section component (the ladder's substrate) | both halves LIVE + observed; capability proven, NOT used | A component can carry a PLAN + criteria fence and per-site NOTES; PLAN is the fleet-wide contract, NOTES the per-site verdicts. Go gate watched at runtime 07-31 (corr `8f564028`, v1.0.1215), not inferred from a build date. LANDMINE: `doc_plans` is back to 0 component rows — no real contract exists yet, and doc_notes also allows 'landmine', so re-adding its CHECK from doc_plans' array orphans the live corpus | documentation-system.md |
| DOC-070 | `PROBE_doc_subject_go_gate.sh`: make the running binary print its own doc-subject vocabulary | built, used to a PASS | The runtime half of the four-enforcement-point checklist: an invalid `subject_type` makes the pod render `validDocSubjectTypes` as compiled. Workflow travels INLINE in the message, so no agent row is written. LANDMINE: two Go gates over one list with DIFFERENT messages — `load_doc_context` cannot produce `unsupported subject_type`; and a green run with no control arm is indistinguishable from a probe that never ran (VOID) | documentation-system.md |
| TL-036 | Offline criteria-fence trial + mutation prover (try_fence / prove_fence_can_fail) | built, cluster-verified | A fence can be run and proven able to fail BEFORE it is published as a tool contract; uses the fleet's own RunChecksAction. LANDMINE: a check's ID is an assertion nothing validates — `selector_count` does not assert a count | tool-lifecycle.md |
| TL-037 | Tool numeric-equivalence gate (`toolgolden.py`) — capture what a tool COMPUTES | built, proven able to fail | Nothing in the platform verified that a tool computes the right numbers: Tier 2 validates selectors and "CONFIRMs, never refutes"; `toolaudit.py`'s RESPONDS is satisfiable by a page with one number input and NO SCRIPT AT ALL (proven by construction). Records every id-bearing element's text+display per input vector; caught a divisor 12→11 error toolaudit scores RESPONDS. LANDMINE: a mid-parse capture records £0.00 for everything and reports success | tool-lifecycle.md |
| DBG-051 | Assumed-status-values trap | deployed | Never assume a status column's vocabulary; always SELECT DISTINCT first | debugging.md |
| MIGG-001 | Proposed migration runner/ledger for hand-applied agent-def changes | aspirational | Never built; only manual "2d state check" stands in; responds to DBG-010's incident | migration-governance.md |
| OPP-002 | Operator discipline: verify-by-artifact, dated backups, kcat | deployed | Never trust a status; diff bytes, dated backups, kcat trigger convention | operator-practice.md |
| OPP-004 | `check_register_coverage`: the register coverage cadence, on the commit path | deployed | A commit creating a workstream the register has never heard of now says so, to the person creating it; closes bugs_open/106. | operator-practice.md |
| HITL-011 | HITL approval-as-specialised-agent architecture (human-reviewer plan) | aspirational | Never-built plan: dedicated human-reviewer agent, approval_tasks/versions tables | hitl.md |
| CTS-018 | system.internal site convention | deployed | Never-deployed sites row hosting maintenance/library work items | contracts-and-standards.md |
| DEV-015 | Chassis action input conventions (dual registration) | deployed | New actions need BOTH ActionInputSpec registration AND a GlobalActionRegistry entry (IsLocal:true). | development-guide.md |
| TRF-015 | intent-probe content component | deployed | New content_components row: plain HTML form + beacon, no JS, capturing anonymous intent | traffic-analytics.md |
| IMG-016 | needs_imagery branch in image-build-handler (Phase 2G.5) | deployed | New generic branch handling structured needs_imagery items alongside legacy branches. | imagery.md |
| DEV-043 | Observability signature-fields pattern | deployed | New marker fields in a result map prove which code path (old vs new) actually executed. | development-guide.md |
| SYS-074 | VM-Hosted Backend Sites class (proposed) | aspirational | New persistent internet-facing VM class with DNS/TLS/data-return lifecycle | system-architecture.md |
| SCH-010 | Private inert pipeline statuses pattern | deployed | New pipeline uses unrecognized statuses so it's inert to existing sweeps by construction | scheduler-and-tasks.md |
| LNK-010 | B4/B5 Browse-All hub links via `section_index_for` queryresolve verb | deployed | New queryresolve verb replaces empty *_index_url specs for hub buttons | link-management.md |
| RES-001 | vertical-exemplar-researcher — the exemplar-research relay hop | deployed | New relay hop researching 3 vertical exemplars; verified end-to-end on dartsonline | research-agents.md |
| VET-006 | vet-json-exporter / vet-export-orchestrator agent pair | aspirational | New wrapper pair for vet-practice service prices, modelled on med-json-exporter | vet-med-pricing.md |
| SCH-003 | content-feed-trigger workflow shape bug (array vs object count) | deployed | News trigger broken for weeks by array-vs-object count field shape mismatch | scheduler-and-tasks.md |
| NEWS-009 | Price-news TTL and news->infographic enhancements | aspirational | Nice-to-have backlog: short-expiry price news and news-driven infographic generation | news-feed-pipeline.md |
| SYS-019 | sites.status informational lifecycle label (system-architecture angle) | deployed | No build-time code filters on sites.status; build dispatch keys on site_work_items instead | system-architecture.md |
| TRF-004 | Minimal-data privacy posture (UK GDPR/PECR) | deployed | No cookies/JS/IP/names logged; declared load-invariant even under traffic pressure | traffic-analytics.md |
| FIX-037 | Architecture-change visibility (Q-E signals / detector) | partial | No formal detector; guardian's informal judgement substitutes | fix-loop.md |
| DYN-004 | Games as a content type (largest pipeline gap) | aspirational | No game generator/library/spec aspect exists; page_type missing | dynamic-applications.md |
| DBG-006 | jsonb && operator class bug | deployed | No jsonb&&jsonb operator; css_snippets errored silently for months | debugging.md |
| PBP-007 | No component-level regeneration trigger (whole-page rebuild remedy) | deployed | No mechanism to regen one component; only remedy for bad content is a full page rebuild | page-build-pipeline.md |
| SYS-052 | Universal orchestration principle & agent_group_definitions elimination | deployed | No orchestrator/worker distinction; groups became agents with spawn/call workflows | system-architecture.md |
| CASE-015 | idea.uk request-then-confirm intake with capacity throttle | deployed | No payment until operator screens the request; MAX_ACTIVE_ORDERS throttle | site-case-studies.md |
| DBG-010 | Hand-applied agent-def migrations have no ledger; re-run reverts later ones | deployed | No schema_migrations ledger; idempotent only vs own prior application | debugging.md |
| AGOV-008 | The outcome-record gap | aspirational | No signal records whether a deliverable actually succeeded | autonomy-governance.md |
| MCL-008 | Kafka cluster-wide authorization gap (ACLs decorative) | partial | No spec.kafka.authorization block means declared ACLs are unenforced | multicluster.md |
| ADO-022 | Adopt-from vs deploy-to separation (unbuilt staging) | aspirational | No staging area distinct from live deploy target for adopted rebuilds | adoption-pipeline.md |
| PLAN-033 | Roadmap-phases scope decision gap (nav grounded in built reality) | partial | No submission path produces a phased roadmap; planner has no ELSE branch without one | site-plan-and-reconciler.md |
| TRF-006 | Visit beacon and events-per-1k metric | deployed | No-JS 1x1 beacon counts human visits as denominator for the core intent-rate metric | traffic-analytics.md |
| IMG-007 | Algorithmic imagery discovery checks (Phase 1 trio) | deployed | No-LLM checks: unfulfilled_image_prompt, placeholder_image_in_use, image_url_404. | imagery.md |
| SPEC-013 | spec-updater (mechanical site_specs merge from findings) | deployed | No-LLM handler applying suggested_value patches to site_specs | site-spec-and-classifier.md |
| DIAG-030 | Anchorless (code-only) diagnosis degrade at load_runtime + error_step lesson | partial | No-anchor runs hard-errored until a config-level error_step fix degraded them to code+schema bundles | diagnosis-loop.md |
| SOC-006 | Cold-start design: AI sparring partner and solo-first completeness | aspirational | No-signup provocation+AI-sparring first 10 seconds; complete for a lone user | social-media.md |
| CVP-004 | Strategic fallback stubs for non-replicable components | abandoned | Non-replicable components get a working fallback plus a HITL developer task for v2 | conversion-playbooks.md |
| FIX-004 | Superseded: null-site-allowed intake design | superseded | Null-site design proved schema-impossible, replaced | fix-loop.md |
| FLW-003 | Voice parameters (numeric stage-tuned voice) | abandoned | Numeric voice dials per flow stage; superseded by persona selection | flows-and-narrative.md |
| LQT-002 | LLM reliability strategy for component generation | partial | Observability first, then shrink the LLM's bookkeeping contract | llm-quality-testing.md |
| DBG-020 | Deployed-binary-predates-disk failure class | deployed | Observed behaviour contradicts correct code because the deployed image predates the repo | debugging.md |
| OPP-001 | In-chassis replicability requirement for operator work | deployed | Off-platform actions must map to chassis ops or a named gap | operator-practice.md |
| SYS-023 | work-site-orchestrator (monolith) vs build-site-planner (thin planner) | deployed | Old inline monolith replaced by a thin planner that delegates via work items | system-architecture.md |
| MDL-027 | RAG best practices (superseded v1) | superseded | Older doc superseded by the v2 best-practices doctrine (see RAGK-004) | model-infrastructure.md |
| CLC-002 | StoreGeneratedComponentAction regeneration branch | deployed | On storing a generated component whose `function` matches an existing row (`WHERE function=$1 AND forked_from IS... | component-lifecycle.md |
| ONB-003 | Three-layer onboarding/config model (mechanical / conventions / intent) | aspirational | Onboarding as three problems with different derivability/confirmation needs | onboarding-config.md |
| DES-076 | Fork_theme step double-creation guard | aspirational | Once site-design-planner runs, the pre-existing `fork_theme` step still present in webdesign-agent risks... | design-composition.md |
| IMG-043 | Sprite-sheet bullets and list treatment (Phase I2) | partial | One N×M glyph grid per site sliced by CSS background-position; active build phase. | imagery.md |
| FIX-026 | git_adapter_request generic adapter caller | deployed | One allowlisted-verb caller for all git-adapter ops | fix-loop.md |
| SPEC-014 | Specialist architects per site type (legacy) | partial | One architect agent per site type with its own component filter | site-spec-and-classifier.md |
| CVP-003 | Minimal Viable Funnel (pragmatic-first Day-1 build) | superseded | One behavioural model + 3 generic components solves cold-start; shipped as mvp-site-builder | conversion-playbooks.md |
| NEWS-014 | rerender-pages refresh-flag coupling | aspirational | One boolean conflates three independent refresh operations; split proposed, not done | news-feed-pipeline.md |
| AGOV-001 | Confirm-not-initiate + single central confirmer | aspirational | One component applies all proposed→active transitions | autonomy-governance.md |
| CTXA-006 | Runtime evidence keyed by orchestration_id (the run narrative) | partial | One correlation key reconstructs a coherent single-run story from three cheap DB/log reads | context-assembly.md |
| VMB-007 | Multi-domain single-binary hosting and domain onboarding/relocation | partial | One engine binary behind many domains; THANKS_PATH is engine-wide per box | vm-backend-sites.md |
| CGV-012 | Standards curation & governance — concern curators | aspirational | One flat curator agent per top-level concern, reusing the auditor pattern | content-governance.md |
| DEV-087 | execute_llm_prompt generic action with DB prompt templates | deployed | One generic action renders the agent's DB-stored prompt_template and calls the configured LLM. | development-guide.md |
| EMAIL-001 | Operator email identity: leopardess.uk + deterministic per-site addresses | partial | One operator domain fronts all sites' mail; deterministic address encoding | email-infrastructure.md |
| VONC-001 | Spark daily-provocation product (vonc.com) | partial | One provocation/day + Gauntlet + Archetype; landing page IS the product | vonc.md |
| SYS-077 | Agent chassis — generic configurable agent executor | deployed | One reusable Go binary becomes any agent type via database configuration | system-architecture.md |
| DOC-012 | doc_notes append-only log with jsonb category roll-up | deployed | One row per NOTES entry; GIN-indexed categories jsonb | documentation-system.md |
| VET-008 | Med scrape evidence store | deployed | One row per page fetch: markdown, content hash, variants accounting, 90-day retention | vet-med-pricing.md |
| TRF-005 | Intent event record (fields, omissions, landing_query enrichment) | deployed | One row per submission with kind/value/ref_host/country; no IP/UA/cookies ever recorded | traffic-analytics.md |
| FTW-010 | Dataset profile/schema heterogeneity iter_0 | deployed | One training slice spans 3 component JSON schemas; anchors max_seq choice | finetuning-flywheel.md |
| SPEC-002 | Site spec unification: site_specs aspect-versioned store | deployed | One versioned spec per site as independent aspect rows with provenance | site-spec-and-classifier.md |
| ADO-018 | Single-agent adoption trigger, superseded by wrapper orchestrator | superseded | One-agent positional-domain trigger replaced by spawn->call wrapper | adoption-pipeline.md |
| FTW-017 | fp16 adapter save decision | aspirational | One-line fix to halve adapter size (fp32→fp16); agreed but never shipped | finetuning-flywheel.md |
| TRF-003 | Probe page pattern — one invited action, plausible framing | deployed | One-line tagline, single action, no JS/cookies, framing follows the domain's heritage | traffic-analytics.md |
| TRF-013 | intent_site_stats visit-count snapshot | partial | One-row-per-host cumulative /stats snapshot feeding the events-per-1k rate calculation | traffic-analytics.md |
| MDL-036 | Text-provider wiring reality (two providers end-to-end) | deployed | Only Anthropic and Ollama actually work end-to-end for text | model-infrastructure.md |
| ADO-011 | Adoption fidelity dial (locked/high/medium/low; phases 1-4) | partial | Only Phase 1 implicit-high exists; real per-item dial unbuilt | adoption-pipeline.md |
| ADO-037 | Verbatim adoption (fidelity=locked) + deploy_mode component key | deployed | Preserves crawled URLs and bytes; skips recreate and the restyle cascade | adoption-pipeline.md |
| SPEC-003 | Fidelity dial (locked/high/medium/low + no-adoption confidence mode) | partial | Only Phase 1 implicit-high fidelity exists at the platform level | site-spec-and-classifier.md |
| NAV-001 | Nav agent family and the three-tier authority model | partial | Only Tier 1 (strategist, new-build) is fully implemented of the three tiers | navigation.md |
| DEV-049 | Schema-before-SQL discipline | convention | Only a live \d schema dump gives real column names and persistence; code names tables, not columns. | development-guide.md |
| PAY-001 | Stripe webhook-as-truth payments pattern (idea.uk implementation) | deployed | Only a verified webhook grants entitlement; live £29 payments proven end-to-end | payments.md |
| DIAG-022 | Repo-cloning token gate (isRepoCloningAgent) | deployed | Only allowlisted repo-cloning agent types get the GitHub read token injected at spawn time | diagnosis-loop.md |
| DBG-031 | R6c artifact-forensics method: cache-busted, metric-consistent comparisons | deployed | Only compare artefacts with identical metrics; md5 before concluding stale-cache | debugging.md |
| LNK-012 | unresolved_cta build-time HITL signal | deployed | Only place a correctly-dropped CTA button's absence is detectable pre-deploy | link-management.md |
| WDS-008 | Discovery auto-triage and scheduled-audit open questions | aspirational | Open questions on whether discovery emissions should auto-triage and audits should schedule | work-dispatch.md |
| IDEA-015 | idea.uk standalone service page-serving and deploy gotchas | deployed | Operational failure catalogue for the single-binary Go service, each tied to a live incident | idea-product.md |
| HITL-010 | Manual HITL continuation runbook | convention | Operational procedure for un-sticking HITL flows via awaited_requests + kcat | hitl.md |
| PBP-010 | Re-render vs rebuild distinction (which path fixes what) | deployed | Operational rule: re-render fixes header/footer; only rebuild re-resolves schema CTAs | page-build-pipeline.md |
| MDL-034 | iter0 pre-trigger + Phase B/C/D deploy runbooks | deployed | Operational runbooks staging the launcher and checkpoint-upload rollout | model-infrastructure.md |
| DEV-083 | Basic operations reference (kcat spawn, scale, monitoring) | convention | Operator layer: scale deployments, post spawn_group/orchestrate via kcat, monitor by correlation_id. | development-guide.md |
| AGOV-011 | Morality review as configured, layered standard | aspirational | Operator-chosen base standard, not a baked-in moral view | autonomy-governance.md |
| BIZ-014 | Operator-vs-vendor business model fork / SaaS commercial model | aspirational | Operator-primary, vendor-optional; five cheap-now seams for future separability | business-strategy.md |
| SCR-001 | Polite-scraping throttle (REQUEST_THROTTLE_MS) | aspirational | Optional per-adapter delay env var for polite bulk scraping | adopting-and-scraping.md |
| SCR-002 | Fetch-recorded provenance (datahelpers.ExtractFetchProvenance) | deployed | Provenance read from the fetch record, never asked of the model; live on v1.0.1192 | adopting-and-scraping.md |
| SCR-003 | Declared config-key contract + unknown-key detection (ActionInputSpec) | deployed | Actions declare the step-config keys they read; three states, unknown/recognised/conditionally honoured. Opt in with CheckConfig: true | adopting-and-scraping.md |
| SCR-004 | Config-key COVERAGE report (scripts/audit-config-keys.sh, cmd/config-key-audit) | deployed | Joins the binary's declarations against every live agent_definitions step config; separates unknown keys from undeclared actions | adopting-and-scraping.md |
| SCR-005 | Config-key ADOPTION report (cmd/config-key-audit --specs) | deployed | Dumps every action's FULL spec, so "who is one line away from opting in?" is answerable without hand-reading 208 actions | adopting-and-scraping.md |
| DIAG-029 | Diagnosis subject threading through orchestrator input_mapping | deployed | Optional subject_type/subject_key fields required three coordinated edits across mapping and both contracts | diagnosis-loop.md |
| MDL-018 | Anthropic client temperature parameter removed unconditionally | superseded | Opus 4.7+ rejects any non-default temperature; client omits it on every call | model-infrastructure.md |
| VMB-011 | Cloudflare-proxied-in-front option | deployed | Orange-cloud CF in front of the VM origin; real-IP nginx config required | vm-backend-sites.md |
| CQ-011 | Audited content pipeline (persona -> research -> draft -> veracity/copyright audits) | aspirational | Orchestrated content generation with fact-check and plagiarism/copyright audit stages | content-quality.md |
| SCH-012 | diagnose-dispatch-loop (automatic dispatch) | partial | Orchestrator claims awaiting_diagnosis items; shipped with scheduled trigger disabled | scheduler-and-tasks.md |
| FTW-027 | model-trainer orchestration chain / Phase 5 kickoff | deployed | Orchestrator spawns data-preparer→provisioner→launcher over Kafka/saga | finetuning-flywheel.md |
| IDEA-006 | idea.uk service: request-then-confirm flow, REVIEW_BEFORE_PAY, capacity cap | deployed | Order state machine; live and earning, proven end-to-end with a real card | idea-product.md |
| BIZ-018 | EBORG — Evidence-Based Organisational Planning (venture concept) | unknown | Org maps roles to AI-agent frameworks; used as the HITL demo client | business-strategy.md |
| ADO-028 | 11-agent website analysis framework (legacy) | superseded | Original 4-group 11-agent web-capture master plan | adoption-pipeline.md |
| LNK-021 | link registry, cached navigation structures, and redirects (foundation) | deployed | Original MVP schema: link_registry + versioned nav cache + redirects table | link-management.md |
| DBI-007 | Clients → networks → sites hierarchy (early spine) | superseded | Original multi-tenancy spine; networks/clients now rarely referenced vs sites | database-and-infrastructure.md |
| PLAN-042 | Website-builder agent group (six-specialist pipeline) | superseded | Original six-agent site-creation flow, replaced by the site_plans/webdesign pipeline | site-plan-and-reconciler.md |
| CTS-042 | data-function contract + P1/P2/P3 fallback | superseded | Original structure/content decoupling; superseded by kebab function naming contract | contracts-and-standards.md |
| DOC-062 | Classic pre-docs024 documentation tree (emptied) | abandoned | Original top-level doc set now all zero-byte archived files | documentation-system.md |
| SNAP-003 | Milestone-tagged site-spec history with inline git-snapshot function | partial | Original unbounded milestone history replaced by pruned specs + snapshot-agent | site-snapshots-and-revert.md |
| BIZ-006 | idea.uk as an instance of the paid multi-domain chat plan | superseded | Originated as a chat-domain day-pass product; shipped as an always-on report service instead | business-strategy.md |
| IDEA-012 | Multi-tenant branded intake pages on one central engine (white-label) | aspirational | Other sites offer the product via branded pages POSTing to the central service | idea-product.md |
| SYS-010 | CollectedData pathologies | deployed | Overloaded single-channel data structure with documented duplication/namespace pathologies | system-architecture.md |
| SYS-018 | Oversize-result delivery: fail-loud hardening + size guards | deployed | Oversize completions now error loudly instead of shipping full collected_data as a stub | system-architecture.md |
| BIZ-029 | Underserved-niche and vertical showcase strategy | abandoned | Own narrow verticals with showcase domains funnelling to purchasable workflows | business-strategy.md |
| FIX-039 | Platform-not-site-data fix philosophy | deployed | Owner ruling: fixes must target platform, not one site's data | fix-loop.md |
| ONB-011 | Stack-discovery agent (inspect → interpret → probe → confirm) | aspirational | Owns mechanical layer: facts, interpreted proposals, declared probe plan | onboarding-config.md |
| IMP-008 | April 2026 pipeline triage fix set (P1–P5) | deployed | P1: component_id plumbed through create_work_item into site_work_items (unblocking tool-improver's load_tool).... | improvement-loop.md |
| DOC-024 | Deliberate-decisions sections + the graduation rule (prose → structured → enforced) | convention | PLAN carries do-not-re-fix prose; enforcement deferred until recurrence | documentation-system.md |
| FIX-044 | Q-H human-facing result package | deployed | PR body carries diagnosis + plan + council decision together | fix-loop.md |
| DOC-034 | Traffic-probe context packaging (docubundle) | deployed | Packager bundles task brief, domain list, deploy docs for cold start | documentation-system.md |
| DOC-007 | Packaged canonical-doc copies as debug context (003 contracts copy) | convention | Packaging workflow drops canonical docs + code dump alongside notes | documentation-system.md |
| SYS-021 | pages.sections is the build-read field | deployed | Page build resolves sections via site_specs.site_plan then falls back to pages.sections | system-architecture.md |
| PBP-021 | load_page_record lookup semantics (name-first, page_id fallback) | deployed | Page lookup resolves by name first, page_id fallback; returns sections+count that gates builds | page-build-pipeline.md |
| DBG-013 | Zero-planned-sections silent no-op success (planning gap) | partial | Page with 0 sections routes to complete_error, a success-labelled step | debugging.md |
| PLAN-012 | Planned-but-uncomposed pages gap (catalogued, never composed) | aspirational | Pages exist with nav intent but no sections/plan rows; two-phase re-plan-then-build needed | site-plan-and-reconciler.md |
| CASE-010 | idea.uk - AI ideation-as-a-service product | deployed | Paid £29-report tool running the internal ideation method, live and earning | site-case-studies.md |
| BIZ-013 | Honest-delta disclosure discipline (built vs admitted-not-built table) | convention | Pairs every ambitious claim with an honest ledger of proven vs possible vs unbuilt | business-strategy.md |
| MCL-001 | Multi-cluster dispatch contract (dispatch_agent + remote-job-spawner) | partial | Parent action publishes to Kafka; remote spawner creates the K8s Job on the target cluster | multicluster.md |
| DBG-063 | Parent-timeout vs child-HITL race | deployed | Parent call_agent timeout fires before child HITL answered, losing the pause | debugging.md |
| TRF-001 | Traffic-probe mission/program (residual-traffic intent capture) | deployed | Parked domains get a probe page inviting one action to rank real rebuild demand | traffic-analytics.md |
| WII-005 | Work-item dedup index + two-strike anti-churn rule | deployed | Partial unique index over non-terminal statuses plus insertWorkItem's two-strike suppression | work-item-integrity.md |
| DEV-027 | Work-item dedup mechanics (idx_swi_dedup, suppression window, two-strike) | deployed | Partial unique index plus 3h suppression window plus two-strike escalation to unresolved. | development-guide.md |
| BIZ-007 | Voluntary pay and "free goes" rejected → free taster + paid report | abandoned | Pay-if-satisfied and N-free-goes rejected; replaced by £0.02 taster + £29 report | business-strategy.md |
| STY-009 | Hero ink model and the structural-dark exception | deployed | Per-branch --hero-ink variable drives hero contrast; imageless heroes are the common case | styling-render-pipeline.md |
| MDL-007 | LLM tiering + cluster-then-slot-fill scaling pattern | aspirational | Per-call-site llm_tier annotation mapped to endpoint via flippable config | model-infrastructure.md |
| CHAT-003 | Build-time context pack (per-domain bounded context) | aspirational | Per-domain JSON pack (identity, scope, grounding chunks, limits) built at install time | site-chatbot.md |
| MDL-009 | thunder-training-monitor | partial | Per-instance SSH probe classifying ALIVE/DONE_OK/DONE_FAIL/GONE_UNKNOWN | model-infrastructure.md |
| DBG-009 | Presign/awaited-loop O(K²) state bloat, fixed by batch adapter calls | deployed | Per-item awaited loops re-persist full state; batch adapter call is the structural fix | debugging.md |
| DIAG-027 | diagnose_assemble_bundle action | deployed | Per-iteration action composing hypothesis + code + runtime into the bundle the verdict step reads | diagnosis-loop.md |
| IMG-048 | Image performance budgets (Phase I7 / D8) | aspirational | Per-kind byte/dimension ceilings enforced at deploy with a weight-over-budget check. | imagery.md |
| IMG-045 | News imagery (Phase I5) | aspirational | Per-news-item chart/illustration/none classification with a grace-interval fallback image. | imagery.md |
| SYS-060 | Fuel budget resource management | partial | Per-orchestration compute budget plumbed through headers, enforcement unconfirmed | system-architecture.md |
| PBP-008 | page-build-handler build path | deployed | Per-page orchestrator: load, plan_sections, write, validate, save, deploy, one commit per page | page-build-pipeline.md |
| SYS-012 | Response-topic consumer group race | partial | Per-pod consumer groups fan every response out to every pod, causing version races | system-architecture.md |
| MDL-015 | Adapter-managed SSH access (ed25519 in k8s Secrets) | deployed | Per-provision keypair, deterministic Secret naming, ubuntu not root login | model-infrastructure.md |
| CTS-017 | Legal rules schema and content_direction | aspirational | Per-site legal_rules + page-level content_direction; legal-content-agent still planned | contracts-and-standards.md |
| IMG-038 | imagery_style_guide — per-site brand guide as data (Phase I1) | deployed | Per-site palette/medium/mood/avoid/reference guide driving generation with per-kind gating. | imagery.md |
| AGOV-004 | Trust ledger + bidirectional ratchet | aspirational | Per-tenant-capability trust level with asymmetric mutation | autonomy-governance.md |
| NAV-008 | Tool nav integration | partial | Per-tool nav entries work but grouping/label-length design remains open | navigation.md |
| CHAT-001 | site_chat_turns table (per-domain chatbot turn logging) | aspirational | Per-turn PII log, separate from llm_call_log; migration number disputed (046 vs 086) | site-chatbot.md |
| SCH-014 | Stale orchestration sweeper/reaper | deployed | Periodic DB sweep classifies expired awaited requests after goroutine death on restart | scheduler-and-tasks.md |
| SYS-004 | Stale orchestration sweeper | partial | Periodic DB sweep classifies/repairs expired awaited_requests instead of in-memory timeouts | system-architecture.md |
| SCH-016 | thunder-training-monitor + worker (probe/classify/reconcile/decommission) | partial | Periodic orchestrator probing training boxes; terminal/decommission branch unverified live | scheduler-and-tasks.md |
| TLIB-014 | component-quality-auditor (library health scoring) | deployed | Periodically scores content_components via compute_component_quality and creates needs_component_regeneration... | tool-library.md |
| HITL-009 | input_requests HITL persistence | deployed | Persisted human input requests with pending view, expiry, and Kafka reply routing | hitl.md |
| HITL-013 | approval_requests table | partial | Persistence for pending approvals; write path initially stubbed | hitl.md |
| DBG-015 | agent_error_log/llm_call_log/http_request_log as primary forensic sources | deployed | Persistent DB logs outlive pod stdout; read them before kubectl logs | debugging.md |
| VMB-001 | VM-hosted backend sites — a new infrastructure class | partial | Persistent non-reaped internet-facing VM class outside k8s, generalised from idea.uk | vm-backend-sites.md |
| ADO-033 | Site interrogation & pattern library | aspirational | Persistent unfulfilled idea: learn patterns from successful sites | adoption-pipeline.md |
| PERS-002 | Persona cognitive architecture (swappable cognitive components) | abandoned | Personas as cognitive entities with swappable memory/reasoning subsystems | persona-architecture.md |
| DEV-057 | Remove-loops plan: input_mapping, contract validation, sequential_fan_out | partial | Phase 1 (input_mapping/contracts) landed; phases 2-4 superseded by the dispatch-loop architecture. | development-guide.md |
| PLAN-032 | Design/composition work-item emission gap (planner reorg unclosed seam) | deployed | Phase-1 planner emits only needs_page+needs_rerender, dropping needs_design/composition | site-plan-and-reconciler.md |
| VET-003 | business_prices deprecation migration pattern | aspirational | Phased table retirement: COMMENT ON TABLE deprecated, drop only after Go cutover | vet-med-pricing.md |
| IMG-031 | Per-kind prompt gating and five-place new-kind checklist | deployed | Photographic brand direction gated off icons/logos; standing checklist for new kinds. | imagery.md |
| SYS-067 | "Database is source of truth, Git is the deployment artifact" | deployed | Pivotal data-ownership doctrine enabling rerendering and granular editing | system-architecture.md |
| CQ-010 | Placeholder-content suppression sweep | deployed | Placeholder text hidden behind HTML comment + needs_human_review item + rerender | content-quality.md |
| DBI-009 | layouts.updated_at trigger and reuse-before-create gate | deployed | Plain CREATE FUNCTION collision gate routed the trigger onto a shared function | database-and-infrastructure.md |
| FIX-048 | Hard deterministic gates between every LLM step | convention | Plain Go gates decide what proceeds; models only propose | fix-loop.md |
| IDEA-010 | Deliverable quality standards for reports and product emails | deployed | Plain-English, professionally-designed deliverables; base64-encode all HTML emails | idea-product.md |
| PLAN-041 | Autonomous section composition: per-section descriptor {role, kind, data_feed} | aspirational | Plan should declare each section's static/dynamic kind and data feed for the pipeline | site-plan-and-reconciler.md |
| ADM-007 | Public REST API for the site-building pipeline | aspirational | Plan to expose sites/pages/work-items/specs over /api/v1/sites/*; never built | admin-dashboard-and-api.md |
| TRF-019 | Relojistas static-rebuild manifest (Thread A) | aspirational | Plan to package a domain's heritage/RSS/inbound-link signals into a static multi-page rebuild | traffic-analytics.md |
| MDL-026 | Self-hosted LLM inference (vLLM/GPU at scale) | aspirational | Plan to serve 7B models via vLLM continuous batching to escape API cost | model-infrastructure.md |
| CQ-014 | component-template-fixer CTA reuse assumption — corrected | superseded | Plan wrongly assumed CTA-fix reuse existed; agent actually punts CTAs to needs_review | content-quality.md |
| IMG-020 | Build-time imagery trigger: emit_imagery_items + imageryplan.go | deployed | Plan-time emission of needs_imagery sharing selection logic with the discovery check. | imagery.md |
| HITL-017 | HITL API for approvals (REST endpoint replacing manual Kafka messages) | aspirational | Planned /api/v1/hitl/respond endpoint; guide written, "For Future Implementation" | hitl.md |
| DYN-002 | Interactive fingerprint parse stage (C1-C6) | partial | Planned Go extractor for canvas/script/library signals feeding intent LLM | dynamic-applications.md |
| NEWS-010 | "Insights section" as the Tier-2 news-feed expansion target | superseded | Planned curated-articles tier was displaced by the archive-first news-index page | news-feed-pipeline.md |
| ENT-002 | Events/tickets vertical (boxing first target) | abandoned | Planned first entity-driven site type (event/performer/venue/ticket); never shipped | entity-data.md |
| IMG-050 | Per-vertical LoRA fine-tunes | aspirational | Planned per-vertical image LoRA training, deprioritised behind reference-image approach. | imagery.md |
| SYS-064 | Environment variable validation framework (abandoned) | abandoned | Planned pre-spawn env var validation framework, silently dropped | system-architecture.md |
| IMG-024 | imagery-quality-auditor agent (Phase 6 / I8) | aspirational | Planned vision-capable auditor sibling dedicated to imagery direction/brand/quality checks. | imagery.md |
| IMG-014 | Planner imagery block prompt extension (Phase 2G.3) | deployed | Planner JSON output gains a structured imagery key in the same LLM call as pages. | imagery.md |
| PLAN-025 | Per-section briefs gap (planner depth): bare section-name strings, no intent | aspirational | Planner emits bare section names with no per-section brief, hiding competing surfaces | site-plan-and-reconciler.md |
| IMG-033 | Planner key stability across replans | aspirational | Planner freely renames imagery keys across replans, causing spurious regenerations. | imagery.md |
| PLAN-015 | Planner ignores adopted state (generic-skeleton overlay) | superseded | Planner invented parallel pages ignoring realised state; fixed by convergence work | site-plan-and-reconciler.md |
| WDS-009 | Terminal work items: every pipeline ends with assembly + deployment | deployed | Planner must emit needs_rerender terminal items or the pipeline produces no website | work-dispatch.md |
| IMG-006 | Planner ignores site_archetype imagery constraints (Bug 4) | aspirational | Planner produced lavish hero prompts despite archetype saying minimal/no photography. | imagery.md |
| IMG-032 | One-entry-one-image decomposition rule (planner prompt patch) | deployed | Planner prompt bans multi-image prompts like "a set of six icons". | imagery.md |
| ADO-034 | Bare-guide / spurious duplicate pages from planner ignoring adopted state | deployed | Planner re-invents differently-slugged sibling pages; cleanup applied | adoption-pipeline.md |
| PLAN-001 | Plan as declarative artefact + reconciler (Kubernetes-style desired-vs-realised) | partial | Planner writes desired state; deterministic Go reconciler diffs vs realised, emits needs_page | site-plan-and-reconciler.md |
| FIX-023 | Write step / fix-implementer agent (F1.1b(c)) | deployed | Plan→branch→PR write organ, proven live 2026-07-13 | fix-loop.md |
| DBI-010 | system.internal pseudo-site anchor pattern | deployed | Platform-wide work items anchor to a fixed pseudo-site instead of a null site_id | database-and-infrastructure.md |
| ADP-005 | Tier-4 browser-runner adapter (headless Chromium over Kafka) | deployed | Playwright/Chromium adapter runs page/selector/console checks, P0 deployed and smoke-passed | adapters.md |
| TPI-001 | Audio-monitoring topic discovery with auto-spawned topic agents | abandoned | Podcast transcription -> novel-topic detection -> auto-spawned monitoring agent, unbuilt | topic-intelligence.md |
| DOC-039 | Guideline-compliance review methodology (001/002/003 walkthrough before shipping) | deployed | Point-by-point guideline walkthrough producing a test plan | documentation-system.md |
| DBI-015 | Early schema inventory and since-dropped tables | superseded | Point-in-time \dt+ snapshot preserving tables since dropped or absorbed elsewhere | database-and-infrastructure.md |
| ADP-009 | Analyser adapter (build + migration path) | deployed | Polyglot code-parsing adapter, deployed to production 2026-06-12 | adapters.md |
| ONB-016 | Config-maintenance agent (drift detection as trust ratchet's signal source) | aspirational | Post-baseline drift detection across all three layers, feeds trust ratchet | onboarding-config.md |
| SEO-001 | SEO content agent | aspirational | Post-content sweep owning meta/schema/canonical/OG across all pages; no dedicated category | seo.md |
| DBG-042 | Defect-catalogue discipline: enumerate by root cause, read-pin-confirm-fix | deployed | Post-deployment audit method grouping defects by root-cause family, not symptom | debugging.md |
| STG-009 | Result storage split (DB paper-trail + S3 artefacts) | deployed | Postgres holds the record of what happened; S3 holds the actual product | storage-architecture.md |
| DOC-013 | DB-as-truth storage decision (knowledge_base = derived index; git = mirror) | deployed | Postgres is truth; KB is derived RAG index; git optional mirror | documentation-system.md |
| SQAM-002 | Baseline mechanical quality measurement methodology | deployed | Pre-LLM deterministic HTML metric pass to form falsifiable hypotheses | site-quality-audit-methodology.md |
| CTXA-020 | Reuse search before generation (code AND definition rows) | aspirational | Pre-generation reuse check should cover jsonb workflow/agent definitions, not just Go functions | context-assembly.md |
| FIX-047 | Loop-worthiness test doctrine (five-criteria) | convention | Pre-registered intake test applied three times in founding thread | fix-loop.md |
| PLAN-034 | site-planner (v2, single-LLM-call site plan) | superseded | Predecessor single-call planner; superseded by build-site-planner for work-item builds | site-plan-and-reconciler.md |
| SOC-001 | The Forge — AI-seeded community knowledge platform | abandoned | Predecessor to Spark: AI drafts, humans validate/fork/improve; parked concept | social-media.md |
| IMG-005 | imagery_direction prepend + origin_model/origin_prompt provenance | deployed | Prepends site style direction to prompts; fixed provenance columns being silently dropped. | imagery.md |
| MDL-017 | Thunder checkpoint upload + O(K²) batch-presign retirement | partial | Presigned PUT manifest for durability; batch presign fixed an O(K²) slowdown | model-infrastructure.md |
| VMB-004 | vm-sites content repo and deploy-to-vm Action | deployed | Private repo + rsync-over-SSH Action mirrors the B2 action for VM targets | vm-backend-sites.md |
| RES-005 | Wayback/archive.org grounding method + limitation | partial | Probe pages grounded via Wayback snapshots; sandbox can't reach archive.org directly | research-agents.md |
| ONB-012 | Confirmation by reality (mechanical layer climbs the ratchet first) | aspirational | Probed commands are strongest confirmation; first capability past confirm_every | onboarding-config.md |
| DES-026 | Layout: affiliate-hub | deployed | Product-review/buyer-guide layout — persistent disclosure strip, vertical product "picks" cards, pros/cons... | design-composition.md |
| SQ-001 | Site Quality Programme — the three-way split and seven legs | partial | Programme closing deploys-vs-best-in-class gap via A/B/C split, 7 legs | site-quality.md |
| DBI-012 | Model-written SQL under a three-guard read-only substrate | deployed | Prompt guard + parse-lint + read-only transaction let a model emit arbitrary SELECT SQL safely | database-and-infrastructure.md |
| CTS-022 | component-creator prompt re-aim (painting rules, vocabulary) | deployed | Prompt rewritten from literal dark-section block to four painting models | contracts-and-standards.md |
| FTW-009 | `<no value>` contamination + iter_1 filter floor | partial | Prompt-builder bug polluted iter_0 data; fix-date filter never recorded | finetuning-flywheel.md |
| ADP-016 | vmhost/service-deployer adapter — persistent-VM provisioning | aspirational | Proposed adapter to automate what idea.uk and traffic_probe both did by hand | adapters.md |
| ADP-017 | Shared reply-delivery policy (undeliverable reply → deliverable error) | deployed | One implementation of degrade-resend-once-else-error; the rule held at 1 of 9 produce sites. Registration does NOT pre-clear adoption at the other 8 — architecture seat says that is an RFC moment | adapters.md |
| RES-004 | Chat differentiator ideation agent | aspirational | Proposed agent ranking payable differentiators from asset × AI-capability combos | research-agents.md |
| AME-003 | improvement_proposals — HITL-gated agent evolution queue | abandoned | Proposed agent/variant changes wait for human approval before applying | agent-memory-and-evolution.md |
| TLIB-008 | Component selector by functional requirement (never-built proposal) | aspirational | Proposed capability-based search over content_components — finding a component by what it does rather than by... | tool-library.md |
| SPEC-019 | Email identity in site_spec (deterministic address encoding + email aspect) | aspirational | Proposed convention for per-domain inbound/outbound email identity | site-spec-and-classifier.md |
| DEV-046 | Curated best-in-class standing expectation | aspirational | Proposed dev-guide addition: every commerce domain needs guides+tools+news+curated top-N. | development-guide.md |
| CGV-007 | Standing-ambition default in the mission aspect | aspirational | Proposed domain-submitter default mission_brief so builds lead the vertical, not mirror it | content-governance.md |
| CGV-026 | Recommendation-specialist architecture (bug vs recommendation vs gap) | abandoned | Proposed finding_type routing (bug/gap/recommendation) with approval_mode gate; never built | content-governance.md |
| ADM-008 | site_ownership table / ownership model | abandoned | Proposed junction table for per-user site scoping; never created | admin-dashboard-and-api.md |
| CQ-008 | Post-build validation of structured components (Fix D) | aspirational | Proposed post-build assertion that required structured fields are actually populated | content-quality.md |
| DES-077 | Vertical-specific planner variants | aspirational | Proposed separate agent definitions using the same planner Go code but vertical-tuned prompt templates, so a... | design-composition.md |
| LCO-004 | Default temperature hardening (chassis-level fallback ~0.4) | aspirational | Proposed ~0.4 default once the read path is proven, gated on earlier steps | llm-call-observability.md |
| LCO-002 | Per-field LLM config resolution fallback chain | aspirational | Proposes lifting temperature to the same multi-level fallback max_tokens has | llm-call-observability.md |
| GML-001 | Games quality lifecycle parity (new game_health / game-auditor / game-behavioral-tester / game-improver) | aspirational | Proposes mirroring the entire tool-lifecycle quality apparatus for games, which today have no equivalent:... | games-lifecycle.md |
| MDL-016 | Thunder Prototyping vs Production mode economics | partial | Prototyping proven fine for inference; untested for long training runs | model-infrastructure.md |
| MCL-005 | Same-cluster loopback test requirement (Gap C) | aspirational | Prove dispatch round-trip locally before adding a real second cluster | multicluster.md |
| SPEC-008 | Build-standard classifier migration (best-in-class quality/fit) | partial | Proven-correct prompt migration for quality standard, not yet applied | site-spec-and-classifier.md |
| PAY-005 | Pluggable billing provider abstraction (Stripe as implementation #1) | aspirational | Provider interface sketch normalising webhook payloads; zero retrofit cost claimed | payments.md |
| BATCH-001 | Universal LLM batch-processing architecture (queue, three-gate control, callback contract) | aspirational | Provider-agnostic batch queue, three-gate control, context-free callbacks | batch-processing.md |
| DES-018 | Layout: magazine-grid | deployed | Publication-feel layout: top-level 2/3 main + 1/3 sidebar grid, article cards, featured-article variant,... | design-composition.md |
| FIX-015 | Deterministic council decision + hard veto | deployed | Pure Go aggregation of reviewer verdicts | fix-loop.md |
| STY-045 | Slot-based modular page assembly (proposal, partially adopted) | partial | Pure-concatenation assembly proposal; site_components shipped, JSON-first landed differently | styling-render-pipeline.md |
| SYS-017 | Hosting split: static-serverless front + small always-on backend | deployed | Pure-static sites serverless on B2; multi-LLM/webhook jobs need a small always-on service | system-architecture.md |
| RSN-003 | Four axes governing a development step | aspirational | Purpose/How-well/Where-heading/What-is; trajectory was the gap | reasoning.md |
| VMB-005 | site-engine deploy Action and narrow-sudo privilege model | deployed | Push-to-deploy engine binary via a scoped sudo hook, no root key in CI | vm-backend-sites.md |
| IDEA-004 | Cross-vendor critique (the cut step on a different vendor) | deployed | Quality gate run by a different model/vendor so the method doesn't mark its own work | idea-product.md |
| LQT-005 | Flywheel D — Claude vs local-model quality eval (replay harness) | partial | Quality-testing-category anchor for the paused Flywheel-D eval lane | llm-quality-testing.md |
| CTXA-005 | Multipass fetch: probe → gate → include/reduce/point | partial | Queries probed with LIMIT N+1, then gated by size/sensitivity into full include, reduction, or pointer | context-assembly.md |
| TL-005 | Recreation-loss defect (correctly-routed recreation still produces no deployed widget) | unknown | Query evidence showed all five games on gamesdesign.co.uk had routed correctly to tool-recreation-handler and... | tool-lifecycle.md |
| PLAN-009 | Section-data deferral + reconciler (reconcile_section_data / needs_section_data) | deployed | Query-resolvable section data reconciled automatically; human-sourced fields stay HITL | site-plan-and-reconciler.md |
| ONB-002 | build_queue domain queue with direction spectrum | partial | Queue table with direction spectrum from null to fork_from; seed_build_queue | onboarding-config.md |
| BLD-004 | Two front doors and duplicate classifiers (Q5) | partial | Queue-door and HITL intake-orchestrator both classify sites; consolidation undecided | build-pipeline.md |
| BIZ-002 | finetuning.uk RAG platform product strategy & business plan | aspirational | RAG platform for SMEs, data curation as differentiator; tiers £199-1,499/mo | business-strategy.md |
| FTW-002 | Three compounding improvement channels | partial | RAG, LoRA, and prompt-variant A/B as three independent compounding quality levers | finetuning-flywheel.md |
| CANB-002 | Canine-biology per-vertical knowledge + LoRA project | aspirational | RAG-seeding (chunks→Ollama embed) plus text/image LoRA fine-tuning for vet vertical | canine-biology.md |
| NEWS-004 | Render source-diversity interleaving | aspirational | ROW_NUMBER partition by source caps any one source at ~2 of 6 display slots | news-feed-pipeline.md |
| SYS-087 | Workflow status state machine | deployed | RUNNING/AWAITING_RESPONSES/COMPLETED/FAILED vocabulary, minor drift across eras | system-architecture.md |
| TL-009 | Canonical tool-page section-shape design question and fix options | aspirational | Raises and answers (as a design decision, not yet built) whether a tool page even wants generic... | tool-lifecycle.md |
| TRF-020 | Domain shortlist and selection policy | deployed | Ranked parked-domain export and a policy for choosing which domains to probe first | traffic-analytics.md |
| STY-041 | Assembly action consolidation (3 clear actions) | deployed | Rationalising 6 overlapping assembly actions down to assemble_page and siblings | styling-render-pipeline.md |
| PLAN-011 | Planner re-plan union safety (normaliseRealisedToPlanPage) | deployed | Re-plans union in realised pages with sections so a re-plan can't clobber built pages | site-plan-and-reconciler.md |
| PBP-013 | No-LLM re-render path (rerender_page_sections, Part 2 / Option Y) | partial | Re-renders all sections from stored content_data + fresh resolved_data, no LLM spend | page-build-pipeline.md |
| DIAG-007 | Call-graph re-scope mechanism (evidence-follows re-scoping) | deployed | Re-scope follows evidence-named symbols + call graph, dropping ubiquitous names, not the symptom text | diagnosis-loop.md |
| FTW-031 | Scripts bundle in B2 as deployment unit | deployed | Re-uploading bundle.tar.gz IS the deploy; must stay flat, no DB change | finetuning-flywheel.md |
| ADM-001 | Admin dashboard + nginx gateway architecture | deployed | React SPA + nginx gateway to auth-service/core-manager; Sites/Work Items/Pages/Direction views | admin-dashboard-and-api.md |
| FIX-027 | isRepoCloningAgent spawn gate / token injection | deployed | Read-only GitHub token injected into dedicated pods | fix-loop.md |
| FIX-002 | fix-proposer agent / constrained edit plan (F1.1a) | deployed | Read-only agent drafts ≤8-edit plans from CONFIRMED diagnoses | fix-loop.md |
| DIAG-001 | Read-only, cite-or-abstain diagnosis loop (core concept) | deployed | Read-only agent: hypothesise, gather evidence, cite-or-abstain verdict, re-scope by following evidence | diagnosis-loop.md |
| DIAG-013 | sqlguard stripQuoted: lint false-positive on quoted literals | partial | Read-only lint was tripped by keywords inside string literals (a slug containing "drop"); fix blanks literals | diagnosis-loop.md |
| ADP-010 | GitHub read-token scoping / least-privilege adapter secrets | deployed | Read-only repo-scoped PAT injected only for isRepoCloningAgent types | adapters.md |
| FTW-013 | iter_0 baseline training run | deployed | Real cost/time/loss anchors: ~9h, ~$20, final_loss 0.27 | finetuning-flywheel.md |
| MDL-031 | Phase 5 training-launcher + model-trainer chain | deployed | Real launcher (migration 102) driven by the model-trainer orchestrator | model-infrastructure.md |
| MDL-010 | Monitor/reaper responsibility split | deployed | Reaper is a dependency-free cost backstop; monitor depends on adapter+SSH | model-infrastructure.md |
| WII-002 | Evidence-gated claimed-item-timeout reaper (positive-evidence completion + stale-claim reset) | deployed | Reaper only auto-completes with positive per-item-type artifact evidence, else resets the claim | work-item-integrity.md |
| SCH-013 | Reaper mechanisms, the work-item-claim reaper gap, and the reaper-location correction | superseded | Reapers are SQL pre_query entries, not Go code; no sweep for stuck claimed items | scheduler-and-tasks.md |
| IMG-057 | flag_page_image_rebuild section-scope mapping (Edit H) | partial | Rebuild flag no-op'd for section scope; prefix-split fix pending code apply. | imagery.md |
| CASE-016 | Leopardess rebuild programme (phases L0-L9) | partial | Rebuild of the platform's own consulting site to be honest and useful | site-case-studies.md |
| PBP-012 | Interactive/deferred-section clobber on plan-driven rebuild + carry-forward fix | deployed | Rebuilds plan-drive and DELETE+INSERT, dropping tools/deferred sections absent from the plan | page-build-pipeline.md |
| STY-027 | Render-off-build_status debt (planned-vs-rendered diff) | partial | Rebuilds skip planned-but-missing sections on already-deployed pages | styling-render-pipeline.md |
| BLD-015 | page-rebuild (rebuild pages without re-planning) | deployed | Rebuilds specific pages skipping planner/assets/nav, reusing standard build-loop agents | build-pipeline.md |
| SYS-058 | Perspective transformation | deployed | Receiver's own orchestration becomes primary; sender is responsible for headers | system-architecture.md |
| CTXK-009 | fuse (cmd/fuse) | deployed | Reciprocal-rank fusion (k=60) merging lexical and semantic candidate lists by rank, not raw score | contextkit-toolchain.md |
| MDL-008 | Ollama CPU adapter operational envelope | deployed | Recreate strategy, load timeouts, memory headroom rule, measured throughput | model-infrastructure.md |
| DEV-084 | Guidelines audit (001/002/003 compliance) | convention | Recurring audits confirming an engine/collector honours the dev-guide/architecture/contracts house rules. | development-guide.md |
| STY-040 | Asset bubble-up deduplication (proposal, never shipped) | abandoned | Recursive per-component asset merge proposal; barrier model shipped instead | styling-render-pipeline.md |
| DYN-003 | Four-stage interactive-content pattern (parse/assess/generate/integrate) | aspirational | Reference shape for building any interactive content type | dynamic-applications.md |
| DES-021 | Layout: docs-sidebar | deployed | Reference-grade documentation layout — 3-zone CSS grid (fixed sidebar nav, main reading column, collapsing... | design-composition.md |
| BIZ-011 | Substrate-vs-application pitch framing | aspirational | Reframe from "website builder" to "domain-agnostic orchestration substrate" | business-strategy.md |
| PBP-023 | UpdatePageStatusAction zero-component deploy guard (Option B) | deployed | Refuses to mark a page deployed with zero rendered components | page-build-pipeline.md |
| CQ-003 | Shared-component regen clobber failure mode | deployed | Regenerating a shared component silently emptied every dependent page using old field names | content-quality.md |
| ASG-003 | Agent and group discovery by capability and performance | deployed | Registry service matching capability/performance; platform/discovery/ confirmed real | agent-spawning-and-groups.md |
| LNK-001 | link_registry as first-class link index + planned links-orchestrator family | partial | Registry/extraction/validation shipped; the agent family on top remains unbuilt | link-management.md |
| PLAN-021 | site_plan_partials: single JSONB-blob partial storage (abandoned) | abandoned | Rejected JSONB blob storage in favour of row-per-thing tables for HITL locking at scale | site-plan-and-reconciler.md |
| PLAN-023 | Separate BuildPageURL path-resolver helper (abandoned) | abandoned | Rejected a new helper in favour of extending CanonicalisePage with ParentSection | site-plan-and-reconciler.md |
| STG-004 | Storage credential architecture decision (no storage-adapter) | partial | Rejected a storage-adapter service since multi-MB blobs would wreck Kafka brokers | storage-architecture.md |
| VONC-008 | Option 1 — build-time static content for the daily shells (rejected alternative) | abandoned | Rejected fix that would have frozen provocations permanently at build time | vonc.md |
| PLAN-022 | Three sequential per-partial plan-builder LLM calls (abandoned) | abandoned | Rejected splitting the planner into three LLM calls; one coherent call was sufficient | site-plan-and-reconciler.md |
| FIX-024 | Hard file allowlist (diagnose_prepare_fix_commit) | deployed | Rejects any file outside the approved plan before git | fix-loop.md |
| FTW-034 | Resume path | partial | Relaunch auto-resumes from highest B2 checkpoint; unproven in prod | finetuning-flywheel.md |
| SYS-034 | Site-chrome rendering gap | partial | Relay build path may never invoke the chrome-rendering step; zero <nav> measured live | system-architecture.md |
| FIX-012 | mark_no_sections — referenced-but-never-built step | abandoned | Remedy step named in a comment, never implemented | fix-loop.md |
| STY-026 | Theme-layer render resolution (style_collection → css_theme) | deployed | Render path resolves colour exclusively via style_collection, ignoring site_specs | styling-render-pipeline.md |
| CTS-013 | CSS theme template contract (renderer vs template ownership) | deployed | Renderer owns palette/luminance defaults; theme template owns layout/typography only | contracts-and-standards.md |
| STY-007 | buildSectionDefaults: luminance-keyed dark-only --section-* defaults | deployed | Renderer's only live per-section adaptation; emits nothing on light palettes | styling-render-pipeline.md |
| DEV-070 | evaluate_condition — template-based conditional branching | deployed | Renders a Go text/template expression against CollectedData; next_step becomes a true/false map. | development-guide.md |
| BLD-013 | Sequential page generation (Phase 0 multipage fix) | superseded | Replaced parallel batch page spawning with a sequential per-page loop | build-pipeline.md |
| DES-037 | Scheme-aware weighted layout matcher + needs_new_layout_candidate HITL signal | deployed | Replaced the tags-only, scheme-blind `resolveLayoutByTags` (exact-overlap count, alphabetical ties — the... | design-composition.md |
| FTW-018 | Flywheel D replay-eval methodology + CPU-eval-pod evolution | partial | Replay stored prompts against candidate model; CPU attempt superseded by GPU | finetuning-flywheel.md |
| STY-013 | Dual chrome render paths (build-fresh vs stale rerender-injected) | deployed | Repoint-before-force_rerender ordering prevents stale dark chrome re-render | styling-render-pipeline.md |
| PBP-003 | plan_sections field-source resolution semantics (on_missing, required, defer) + needs_section_data escalation | deployed | Required field defaulting to skip_field silently falls through to defer, dropping sections | page-build-pipeline.md |
| CTS-027 | plan_sections required-field deferral trap | deployed | Required+unresolved field hits switch default, silently drops the whole section | contracts-and-standards.md |
| ABO-004 | mediator model for competing design concerns | aspirational | Requirement-relative balance among fast/secure/simple/etc. | autonomous-build-operate.md |
| REB-005 | Rerender fossilisation (reassembly re-ships stale renders; template changes need full rebuild) | deployed | Rerender reassembles stored HTML; only a full rebuild reaches new component templates | rebuild-cascade.md |
| SYS-014 | Observability gaps: owner_agent_type "generic" | unknown | Rerouted generic workflows keep misleading owner_agent_type, breaking searches | system-architecture.md |
| DIAG-039 | Evidence-fed fuzzy-scope resolver (§7D) | deployed | Resolves English next_scope descriptions to real path:Symbol handles via embedding search before re-scoping | diagnosis-loop.md |
| CGV-016 | page_components.build_status CHECK constraint | deployed | Restricts build_status to a fixed enum after an invented 'approved' value hid a section | content-governance.md |
| DES-027 | Layout: ecommerce-storefront | deployed | Retail-clean, product-forward storefront — promo hero, image-overlay category tiles, product grid, add-to-cart... | design-composition.md |
| CHAT-006 | Three-layer bounding (retrieval / prompt / operational) | aspirational | Retrieval/prompt/operational bounding decomposition to stop chatbot topic drift | site-chatbot.md |
| PAY-007 | Existing but non-functional auth-service subscription scaffold | partial | Reusable subscription package verified as unwired: no SDK, mock usage stats, no webhook | payments.md |
| CQ-017 | Anti-hype voice and claim-discipline spec | deployed | Reusable voice contract: banned hype language, smallest-true-claim, CTA governance | content-quality.md |
| ASG-004 | Workflow template library, lineage and marketplace | abandoned | Reusable workflow templates with lineage, ratings, monetised marketplace idea | agent-spawning-and-groups.md |
| CGV-025 | maintenance_queue as generic install/uninstall trigger surface | aspirational | Reused maintenance_queue as a generic per-site add-on trigger, first for chatbot install | content-governance.md |
| DOC-014 | Abandoned: flat-file docs-repo as truth + docselect catalogue retrieval | superseded | Rev-1 design reversed to DB-as-truth within a day | documentation-system.md |
| CASE-001 | idea.uk live-VM / chassis-staging duality | deployed | Revenue VM untouched while chassis builds deploy invisibly to B2 staging | site-case-studies.md |
| FIX-019 | Verify step (diagnose_run_checks) | deployed | Reviewer SQL checks run under read-only containment | fix-loop.md |
| FIX-043 | Q-G reviewer context (answered narrowly) | partial | Reviewers share one role prompt; no per-reviewer corpora yet | fix-loop.md |
| HITL-007 | content-reviewer (HITL + auto-eval dual mode with pre-validation) | deployed | Reviews page content via human or auto-eval mode with link/email pre-validation | hitl.md |
| DBG-014 | Content↔template key-contract drift (system-stats class) | partial | Rewritten template shares zero keys with stored content_data; visible-content filter drops it | debugging.md |
| ASG-005 | Multi-dimensional template classification and semantic search | abandoned | Rich behavioral/performance/embedding classification for template discovery | agent-spawning-and-groups.md |
| HITL-018 | system.notifications.ui topic and the missing HITL UI service | partial | Rich notification topic once had no consumer; later matured into admin dashboard | hitl.md |
| SYS-079 | Message header contract | deployed | Rich sender identity, retry, and status-enum header set on every message | system-architecture.md |
| DBI-020 | clients_db vs templates_db agent_definitions source-of-truth | deployed | Rich-schema agent_definitions load from clients_db, not templates_db | database-and-infrastructure.md |
| NEWS-005 | Content diversity & original research pipeline | aspirational | Roadmap: topic splitting, readership-segment writers, timelines, scenario analysis | news-feed-pipeline.md |
| AGOV-010 | External rollback + recursive self-improvement risk | aspirational | Rollback must not depend on the agents it rolls back | autonomy-governance.md |
| BIP-009 | vet-pipeline-orchestrator (rolling pipeline) | deployed | Rolling coordinator advancing sweep→promote→verify each run; reworked multiple times | business-intelligence-platform.md |
| IMP-015 | Supervised fixer first-run protocol (disposable specimen site) | aspirational | Rollout protocol for a re-aimed automated fixer: confirm the deployed pod carries it; capture the specimen's... | improvement-loop.md |
| LNK-004 | sourceResolver `pages` fabrication bug — the phantom generator | deployed | Root cause of hero phantoms; also corrected an earlier wrong-cause diagnosis | link-management.md |
| RSN-004 | Why-chain (objective-tree traversal) | aspirational | Root-to-node purpose path used as an anti-drift question | reasoning.md |
| RSN-010 | N-round convergence (author/checker modes) | aspirational | Rounds shrink the active concern set; non-convergence escalates | reasoning.md |
| WDS-011 | Work-item routing: content rebuild vs re-render (needs_page/page_rerender/needs_rerender/link_resolution_rebuild) | deployed | Route by item_type on whether copy is regenerated; link_resolution_rebuild hazard noted | work-dispatch.md |
| VKA-001 | Vertical knowledge architecture (overview) | aspirational | Route domains to specialised verticals with own KB/research/monetisation; Phase 0 only | vertical-knowledge-architecture.md |
| ADO-024 | Tool routing fix deployment status (T1+T2 deployed, symptom unconfirmed) | partial | Routing + detection fix live; widget-deploy acceptance criteria unverified | adoption-pipeline.md |
| BATCH-004 | Work item processing_tier (standard / batch_gpu) | aspirational | Routing column for holding items until a GPU batch window opens | batch-processing.md |
| CGV-002 | Granular editing spectrum & three edit paths | deployed | Routing model: direct edit / brief regenerate / page regenerate / direction propagate | content-governance.md |
| FTW-007 | ChatML export format with metadata sidecar | deployed | Row shape: chat messages + metadata JSONB back-linking to llm_call_log | finetuning-flywheel.md |
| CTS-036 | Atomic standard (generated-views doc tree) | aspirational | Rule-atoms as smallest unit; docs are generated views. Same exploratory track as CTS-035 | contracts-and-standards.md |
| BIZ-020 | Two-tier commercialisation model (sell output → sell setup) | aspirational | Run service in a niche, then sell the whole setup as a business-in-a-box | business-strategy.md |
| SPEC-007 | Phase 0 classifier-only positioning read | deployed | Running just the classifier as a near-zero-cost positioning brief | site-spec-and-classifier.md |
| IMP-003 | improvement-loop orchestrator (discovery→triage→fix→rerender cycle, pass cap, auto-reset) | deployed | Runs after initial build or on schedule/manual trigger: pass-limit gate (≥3 → complete_clean) → spawns the three... | improvement-loop.md |
| CTXK-006 | bundle (cmd/bundle): orchestration wrapper and pure-composer boundary | deployed | Runs dbcontext then assembler; assembler itself never touches SQL or triggers anything | contextkit-toolchain.md |
| VONC-007 | provocations-archive-list component + provocations archive page | deployed | Runtime-fill archive page with clone-template rows and a visible empty state | vonc.md |
| CTS-044 | Generation-time guards for dynamic components | deployed | Runtime-fill marker + no-inline-script baked in at generation, not patched post-hoc | contracts-and-standards.md |
| IMG-026 | Icon generation lessons and image-model comparison | deployed | SDXL judged wrong for flat icons; model comparison drove switch to Banana for icons. | imagery.md |
| MKT-001 | Marketing as work items + OpenClaw adapter | aspirational | SEM/landing/email/social as work items via an adapter to Google/Meta/LinkedIn; unbuilt | marketing.md |
| FTW-038 | Thunder instance lifecycle: reaper + training monitor | deployed | SQL-migration-level summary of the reaper/monitor pair as a unit | finetuning-flywheel.md |
| BLD-010 | Work-item build pipeline: domain-submitter → dispatch loop → handler agents | deployed | SQL-migration-traced current architecture from domain submission to per-page handlers | build-pipeline.md |
| BLD-008 | page-content-writer (section-by-section content generation) | deployed | SQL/prompt history of the writer's anti-fabrication and section-generation contract | build-pipeline.md |
| DBI-003 | API keys logged in plaintext — exposure & rotation | deployed | STABILITY/BANANA keys exposed in logs for 7 weeks, then scrubbed and rotated | database-and-infrastructure.md |
| DBG-067 | Secret hygiene: image-provider API keys logged in plaintext | partial | STABILITY/BANANA keys in logs; rotation repeatedly deferred across sessions | debugging.md |
| IDEA-008 | Click-through operator approval links (HMAC per-order tokens) | deployed | Safe-GET approval pages with HMAC tokens for one-click confirm/approve/decline | idea-product.md |
| STY-012 | Scheme-aware fallback chrome (RenderFallbackHeader/Footer) | deployed | Safety-net chrome functions rewritten from hardcoded-dark to var()-driven | styling-render-pipeline.md |
| DBG-050 | gamesdesign silent-staleness: result-contract stub (output_field vs output_fields) | deployed | SagaCoordinator honoured only plural key; resolveResultSpec fix shipped 2026-06-18 | debugging.md |
| ATM-002 | Requirement-mediation model ("right" as balance) | aspirational | Same balance framing as ABO-004, from the shared preamble | autonomy-trust-model.md |
| FIX-050 | Transferable machinery: legacy-migration and feature intakes | aspirational | Same gate/council scaffolding proposed for other intake types | fix-loop.md |
| FIX-051 | Triage router (Phase 1): deterministic failure sorter | deployed | No-LLM router sorts every fleet failure into bug/blip/no-error/capability-gap | fix-loop.md |
| FIX-052 | Silent-check verifier (Phase 2): the class no work item ever records | deployed | Finds bugs nothing flags (darts nav-page signature), routes via triage router | fix-loop.md |
| FIX-053 | Feedback close-out (Phase 3): all-time resolution recheck + auto-reescalation | deployed | Closes parked escalations whose pattern genuinely resolved; re-escalates if it returns | fix-loop.md |
| FIX-055 | Truncation-gate attribution: `gated_by_truncation` on every council report | built, not yet live | A revise now says whether a seat judged it or merely ran out of tokens; gate unchanged | fix-loop.md |
| FIX-056 | `Council-Submitted:` trailer — review credit that survives committing before the verdict | deployed | Records the correlation at commit time; 098 resolves it later, so approval credits with no amend | fix-loop.md |
| FIX-057 | Recoverable structural plan refusal (`repair_step` on `diagnose_persist_fix_plan`) | built, not yet live | A validator rule no longer discards a completed design; bounded repair loop, gate unchanged when unset | fix-loop.md |
| FIX-058 | Council seat token-pressure instrument: pull report + CTE-only push alert | deployed | Measures HEADROOM, not truncations — the lagging count reads ~0 because the caps were raised, and a raise moves the cliff | fix-loop.md |
| FIX-059 | Seat length budget applier (one block, snapshot-then-write, refuses hand-authored blocks) | built, not applied | Candidate 4's reorder was refuted by measurement; the length budget is the half the evidence credits | fix-loop.md |
| SOC-013 | Vertical integration of Spark mechanics into domain sites | aspirational | Same mechanics re-flavoured per vertical (vet, finance, fashion, food) | social-media.md |
| SQLC-001 | SQL needle-gate surgery pattern (guarded, idempotent, reversible DB edits) | convention | Same method as DBG-016, extracted independently under this proposed category | sql-change-management.md |
| DBG-070 | gpu-provisioner output-shape flattening (output_fields plural vs singular) | deployed | Same output_field/output_fields bug class as DBG-050 on a different agent | debugging.md |
| SYS-082 | Retry semantics | deployed | Same request_id with incremented retry_version, unconfirmed as shipped | system-architecture.md |
| SAAS-001 | Isolated chat/satellite architecture ("Y-copy") for SaaS build isolation | aspirational | Same satellite architecture as CHAT-007, escalated to a build-as-a-service framing | saas-isolation-architecture.md |
| STY-024 | Ambient pass-through pattern for surface painters | deployed | Sanctioned --section-x: var(--color-x) pass-through for fallback-less consumers | styling-render-pipeline.md |
| CHAT-007 | Isolated chat environment (satellite; load/hack/bug vectors) | aspirational | Satellite infra severing load/hack/bug vectors from core; Option X vs Y undecided | site-chatbot.md |
| STG-006 | Checkpoint & final-adapter upload to B2 | partial | Save-index-keyed checkpoint uploads with a hard-gate final-adapter upload | storage-architecture.md |
| WDS-006 | build-pipeline-trigger site targeting via pre_query | aspirational | Scheduled dispatcher defaults to system.internal with no real site targeting | work-dispatch.md |
| CTS-053 | Wrapper-orchestrator requirement finding (001:405-462) | partial | Scheduler-reached + substantive-work agents must not run in shared chassis pod | contracts-and-standards.md |
| CTS-024 | Component schema/template/prompt three-way consistency invariant | deployed | Schema item fields, template tokens, and prompt output must agree; info-card-grid still violates | contracts-and-standards.md |
| CH-003 | Companies House enrichment with succession-risk signals | deployed | Schema: financials, officers/PSC, owner-age/succession-risk derivation | companies-house-enrichment.md |
| DES-059 | Scheme derivation cascade + the drop-at-render gap | deployed | Scheme (light/dark) is derived at composition time from `design_intent.style_direction` by substring matching,... | design-composition.md |
| DBI-022 | Deploy privilege model (site-engine-deploy sudo hook) | deployed | Scoped sudoers rule lets a deploy user swap the binary with no root CI key | database-and-infrastructure.md |
| STY-003 | Component quality tracking (quality_score et al.) | deployed | Scored fields on content_components drive planner/auditor regeneration targeting | styling-render-pipeline.md |
| CTXK-010 | eval_targets + ground-truth eval set + measurement-trap discipline | deployed | Scores resolver output against ground truth (recall@N/MRR) with hard-won task-binding/leak-proofing guards | contextkit-toolchain.md |
| PEV-002 | Hypothesis priority list (learn loop as idea generator, not fact finder) | abandoned | Scraped data treated as correlation; Librarian emits a prioritised testable-hypothesis backlog | portfolio-evolution.md |
| DOC-033 | Context-bundle seeding for fresh agent threads (imagery) | deployed | Script assembles imagery workstream's cold-start context bundle | documentation-system.md |
| DES-025 | Layout: comparison-aggregator | deployed | Search-first, data-dense, trust-oriented layout — hero IS a search input, sticky filter bar, dense horizontal... | design-composition.md |
| MCL-002 | Adjacent-cluster Phase 4a rollout: va001 second cluster | aspirational | Second Rackspace Spot cluster shares primary's Kafka/Postgres for trusted dispatch | multicluster.md |
| CASE-008 | wayfaringlondoner.com page + THANKS_PATH-is-engine-wide | partial | Second probe page surfaced a shared-box thanks-filename constraint | site-case-studies.md |
| IDEA-011 | Chassis-native idea engine (Phase D idea-orchestrator) | aspirational | Second way to run the method as a chassis agent/workflow; not started, needs schema pass | idea-product.md |
| PBP-014 | content_data ⊕ resolved_data persistence model | deployed | Section content_data merges LLM copy with resolved_data, enabling no-LLM re-render | page-build-pipeline.md |
| ADO-013 | Tool/game pages never deployed (A1): section-only parser + flip churn | deployed | Section-only HTML parser missed div-based tool output; fix verified in prod | adoption-pipeline.md |
| REB-003 | Carry-forward path and the carry fingerprint | deployed | Sections failing readiness are carried forward unchanged, re-fossilising stale content | rebuild-cascade.md |
| CTS-012 | Section painting contract (four painting models) | deployed | Sections re-export --section-* as token references; literal colours forbidden | contracts-and-standards.md |
| CTS-003 | content_data is the source of truth; HTML patching rejected | deployed | Sections store content_data (truth) + derived rendered_html; patching HTML is a bridge at best | contracts-and-standards.md |
| BLD-005 | image_tag 'latest' stale-default trap | partial | Seeded agents default to an ancient chassis image via the image_tag column default | build-pipeline.md |
| BIP-002 | Business verticals registry (business_intel) | deployed | Seeds veterinary/online-pharmacy/seaweed-farming with default_agent_type per vertical | business-intelligence-platform.md |
| CTXP-002 | Docubundle / package_*.sh context-packaging practice | deployed | Self-contained package_<subject>_debug.sh scripts + dbcontext + a deploy-mechanisms taxonomy (A-F) | context-pack-tooling.md |
| CTXK-017 | Dogfooding bundle for building the diagnosis loop itself | aspirational | Self-referential bundle recipe scoping the diagnosis loop's own code for its own continued development | contextkit-toolchain.md |
| DBG-068 | Adapter-vs-chassis deployment drift | partial | Separate K8s resources; chassis rebuild doesn't refresh the adapter binary | debugging.md |
| BIP-001 | business_intel schema (multi-vertical business intelligence platform) | deployed | Separate schema modelling businesses for data-collection verticals; ~3,500 vet practices loaded | business-intelligence-platform.md |
| IMG-001 | Imagery loop-closure master plan (Phases 0–6) | partial | Sequenced plan closing spec-to-delivery imagery gaps; Phase 2G/2H verified, 3-6 pending. | imagery.md |
| CTS-050 | Class-level rename (probe → site-engine) and env-var churn | superseded | Service/paths/env-vars renamed from probe-specific to class-generic site-engine naming | contracts-and-standards.md |
| SYS-075 | Pull architecture / no collector VM | deployed | Serving boxes buffer JSONL; the cluster pulls over key-gated HTTPS, no push | system-architecture.md |
| BIZ-015 | domain-strategist (strategy vs architecture separation) | deployed | Sets revenue model/content strategy per domain; planner has final say on architecture | business-strategy.md |
| DBG-029 | Loose dispatch item-status semantics (complete ≠ done) | aspirational | Seven dated sightings of dispatch bookkeeping bugs; fix parked as hygiene | debugging.md |
| DBG-062 | Early message-routing failure-mode catalogue | deployed | Seven traced early bugs behind every core architectural convention | debugging.md |
| LNK-020 | site_specs `cta` aspect + CTA graph audit (parked) | partial | Shared CTA URL source fixed dependants; graph found circular, retarget parked | link-management.md |
| CTXK-014 | ranked-candidate contract (internal/candidates) | deployed | Shared Candidate/File JSON contract emitted and consumed across resolve_targets/embed/fuse/eval_targets | contextkit-toolchain.md |
| STY-042 | Component library unification (component_library.go) | deployed | Shared Go module: one source of truth for component ops and chrome rendering | styling-render-pipeline.md |
| CTXK-002 | internal/analysis package + ReadSymbolBody symbol-body slicer | deployed | Shared analyser output contract and the one symbol-body slicing implementation used by all consumers | contextkit-toolchain.md |
| DYN-005 | Generator architecture convergence (shared interactive-artefact-generator) | aspirational | Shared base generator anticipated once games exist alongside tools | dynamic-applications.md |
| RAGR-001 | knowledge_base: shared pgvector RAG store | deployed | Shared embedded content store (vector(768)+trigram fallback) across industries/collections | rag-retrieval.md |
| SYS-011 | Flat-namespace collision risk and compensating-mechanism accretion | deployed | Shared flat map lets actions silently pick up wrong fields; workarounds accreting | system-architecture.md |
| DBG-072 | Problem-category taxonomy for component/tool defects | deployed | Shared greppable vocabulary tagging incidents so patterns roll up to the guide | debugging.md |
| DBG-026 | configOrInput numeric config coercion (expiry_minutes silently dropped) | deployed | Shared helper type-asserted to string; JSON numbers silently fell through to defaults | debugging.md |
| ATM-001 | Trust ratchet & capability ceiling model | aspirational | Shared preamble framing behind AGOV-004/ABO-001 | autonomy-trust-model.md |
| RSN-011 | N-round candidate ownership | aspirational | Shared seeded candidate, changed only by adjudicated proposals | reasoning.md |
| RAGK-001 | RAG knowledge_base (shared pgvector store) | deployed | Shared table, vector(768), ivfflat+trigram fallback, SHA256 dedup | rag-knowledge-base.md |
| CTXK-005 | dbcontext (cmd/dbcontext) | deployed | Shells out to psql for schema/rows/capabilities/runtime-evidence with multipass row sizing | contextkit-toolchain.md |
| MDL-001 | Model aliases and the model selection strategy | deployed | Short aliases resolved in code; sonnet/haiku/opus/ollama per-step defaults | model-infrastructure.md |
| ADO-014 | Sectionless-page durability stack | partial | Sibling fallback + discovery check + no-sibling flag for zero-section pages | adoption-pipeline.md |
| SYS-013 | Kafka empty partition assignment on simultaneous pod join | unknown | Simultaneous deploy join can leave a partition unassigned; workaround is killing a pod | system-architecture.md |
| ADO-019 | Unified design spec aspect, superseded by design_reference/design_intent split | superseded | Single blended design aspect replaced by concrete/semantic split | adoption-pipeline.md |
| PLAN-002 | CanonicalisePage + ValidateRoles: deterministic page-shape vocabulary | deployed | Single helper maps role/slug/parent to canonical name/url/page_type from adoption+planner | site-plan-and-reconciler.md |
| CTS-040 | Tier D items-array component schema shape | deployed | Single items array + sub-schema replaces numbered-flat anti-pattern; pre-store validator | contracts-and-standards.md |
| STY-043 | page_components: component instances as the page's stored form | deployed | Single most consequential schema decision enabling rerender/edit/lock | styling-render-pipeline.md |
| PBP-011 | save_page_sections: DELETE+INSERT persistence with layered guards | deployed | Single save path with content-regression and interactivity-preservation guards | page-build-pipeline.md |
| DBI-024 | No tenant isolation today; dedicated-cluster-per-client as offered capability | partial | Single shared Postgres/Kafka with no RLS; per-client cluster positioned as buildable | database-and-infrastructure.md |
| LNK-008 | datahelpers/links.go — canonical link classification library | deployed | Single shared normaliser used by both deploy gate and post-deploy audit | link-management.md |
| MDL-005 | ai_endpoint_health: multi-endpoint model routing / GPU scheduler | deployed | Single table is the GPU scheduler; healthy→flow, unhealthy→back-to-triage | model-infrastructure.md |
| TLIB-015 | Component library (content_components) — the base schema | deployed | Single table of reusable renderables: name, html_template, input_schema, `function` (identity), display_name,... | tool-library.md |
| BIP-005 | vet-practice-verifier agent | partial | Single-practice verification workflow; long trail of live production fixes | business-intelligence-platform.md |
| SCH-002 | Kafka scheduler (DB-driven heartbeat service + scheduled_tasks table) | deployed | Single-replica Go service ticks scheduled_tasks every 30s, publishes triggers | scheduler-and-tasks.md |
| CGV-001 | Section-editor architecture (content_data as source of truth) | deployed | Single-section edits via content_data update + re-render, never HTML patching | content-governance.md |
| SYS-016 | Coordinator result-extraction contract (resolveResultSpec) & silent-stub bug family | deployed | Singular output_field silently dropped results; centralised result_spec.go fix shipped | system-architecture.md |
| REB-006 | Chrome refresh gating (render_site_components, force_rerender, repoint-before-rerender) | deployed | Site chrome render skips non-empty slots unless force_rerender; repoint before forcing | rebuild-cascade.md |
| FLW-001 | Multi-track flows (journeys, narrative arcs, layered context) | abandoned | Site modelled as audience journeys with hierarchical context inheritance | flows-and-narrative.md |
| BIZ-009 | Building-and-hosting as a service via chat (recursive satellite platform) | aspirational | Site's own chatbot becomes intake for the whole build platform, recursively | business-strategy.md |
| DIAG-041 | backend_unreachable discovery check | partial | Site-health discovery check that alerts on VM backend down; filed here as a category mismatch | diagnosis-loop.md |
| NAV-005 | Duplicate header/footer pathology | partial | Site-level components leaking into pages.sections cause double-rendered chrome | navigation.md |
| FLW-002 | Brand DNA invariants with bounded variance | abandoned | Site-level immutable identity layer plus allowed voice-variance ranges | flows-and-narrative.md |
| CASE-002 | idea.uk mission and identity (workshop of tools; never verdicts) | deployed | Site-specific brand concept reframing away from the single £29 tool | site-case-studies.md |
| CQ-005 | F8 — shared-component contamination (three carriers) | partial | Site-specific product pitch baked into a shared component via fallbacks/merge/llm_guidance | content-quality.md |
| VET-009 | Med URL discovery via Firecrawl /map | partial | Site-wide product-URL discovery alongside category-page crawling | vet-med-pricing.md |
| PBP-002 | rerender-pages v6 workflow (refresh_site_components gate) | deployed | Site-wide rerender agent; chrome force-render gated on spec.refresh_site_components | page-build-pipeline.md |
| TRF-014 | Ranking queries and graduation criteria | partial | Six read-only queries answer "is there demand"; graduation threshold still a proposal | traffic-analytics.md |
| PERS-001 | Copywriter persona roster | abandoned | Six seeded copywriter personas with style agents, assigned by flow stage | persona-architecture.md |
| HITL-021 | HITL kcat test harness | deployed | Six shell scripts for manually testing the HITL approval loop end-to-end | hitl.md |
| DOC-002 | Anthropic product-knowledge skill (verify, don't recall) | deployed | Skill instructs consulting official Anthropic docs over memory for Claude facts | documentation-system.md |
| CTS-043 | Recursive component tree ("everything is a component") | abandoned | Slot-placeholder recursive RenderNode design; shipped system uses flat sections instead | contracts-and-standards.md |
| FTW-016 | GPU training performance model | convention | Smoke rate ≠ steady state; FA2 lever; O(N²) attention cost of seq length | finetuning-flywheel.md |
| DBG-022 | Operator/assistant division-of-labour + DB-change safety conventions | deployed | Snapshot-before-change, verified replace(), workflow-vs-Go deploy distinction | debugging.md |
| DYN-010 | js-bundle-stale gap (site-asset-renderer not wired into ongoing builds) | aspirational | Snippet bundle only rebuilt at initial design/full rerender, not on change | dynamic-applications.md |
| DEV-009 | Agent vs infrastructure boundary test | convention | Something is an agent only if it owns a domain and needs independent spawn/debug. | development-guide.md |
| SYS-055 | Two-phase agent lifecycle (spawn + initialize handshake) | deployed | Spawn creates the pod; a separate initialize handshake precedes real work | system-architecture.md |
| DEV-069 | Spawn/step naming conventions | deployed | Spawn steps must start spawn_<descriptor>; step names must be unique within a workflow. | development-guide.md |
| DIAG-021 | Abandoned design: diagnostician per-iteration re-invocation | abandoned | Spawn-a-fresh-child-per-iteration design considered and dropped once next_step loop-back was confirmed to work | diagnosis-loop.md |
| DEV-062 | spawn_agent — database-definition-driven Kubernetes Job spawning | deployed | SpawnAgentAction reads the child's agent_definitions row and launches a K8s Job from it. | development-guide.md |
| TLIB-003 | Component selector/creator architecture: section_type vs function split, and the self-extending library narrative | partial | Splits "what role does this section play" (section_type) from "which template" (function). Planner emits... | tool-library.md |
| DES-065 | Component selector + creator (section_type vs function split) | partial | Splits the planner's historically conflated role: the planner decides WHAT section types a page needs; a Go... | design-composition.md |
| IMG-030 | Image provider abstraction and kind→provider routing | partial | Stability/Banana kind-based routing; A6 extension committed but not yet deployed. | imagery.md |
| VONC-006 | brief-explanation static explainer (regeneration, not a loader) | deployed | Stable "how Spark works" content fixed by build-time regeneration, not a JS loader | vonc.md |
| VMB-014 | VM cutover: nginx front door with reserved tool paths | aspirational | Staging-in-place cutover plan for a chassis site sharing a domain with a live tool | vm-backend-sites.md |
| SCH-008 | Concurrency group starvation problem and prevention rules | deployed | Stalled task in a shared concurrency_group can starve the whole pipeline | scheduler-and-tasks.md |
| CTXK-001 | contextkit CLI toolchain (module overview) | deployed | Standalone Go module of context-bundle CLIs; production diagnose-agent is its deployed descendant | contextkit-toolchain.md |
| DBG-061 | Orchestration environment reset runbook (clean-slate test-cycle procedure) | deployed | Standard truncate/scale/topic-delete procedure repeated across early docs | debugging.md |
| DOC-055 | Four-layer documentation model for automation | aspirational | Standards + context substrate + known-good library + trust ledger | documentation-system.md |
| HITL-003 | User-representative advocate (intent + conflict triage) | aspirational | Standing advocate triaging claimed conflicts before they reach the user | hitl.md |
| BLD-006 | Coverage baseline: guides, tools, news, curated top-N on most sites | aspirational | Standing content-coverage policy; curated top-N list mechanism is the one genuine new build | build-pipeline.md |
| DGH-001 | Commit-is-deploy: git → Actions → Backblaze B2 (+ chassis image-tag deploys) | deployed | Standing deploy mechanism: commit triggers B2 sync; chassis code ships via image tag | deployment-github.md |
| DOC-043 | Classify, do NOT merge (the human consolidates) | convention | Standing rule: LLM finds/cites, human decides/writes canonical docs | documentation-system.md |
| SYS-006 | Entity data model | aspirational | State-based lifecycle entities driving pages, news triggers, client-side real-time data | system-architecture.md |
| DBG-039 | 0-rows rule: zero rows decisive only after query AND run completion ruled in | deployed | State-dump substitute for evidence past the idle-reaper's 3600s capture window | debugging.md |
| SYS-020 | Aspiration: agent-creation & message logging workstream | aspirational | Stated desire to log agent creation/inter-agent messages as its own workstream; never built | system-architecture.md |
| DEV-020 | Launch idioms: orchestrate vs work-item insert | deployed | Static agents launch via kcat orchestrate; dynamic handlers launch via a site_work_items insert. | development-guide.md |
| IDEA-009 | Fake-door → intent-capture-first launch discipline | superseded | Static landing page capturing intent without charging; superseded by live Stripe flow | idea-product.md |
| DYN-001 | Dynamic applications direction (three tiers; thin generated backends) | aspirational | Static/dynamic components -> agent-powered backends -> full app generation | dynamic-applications.md |
| VMB-002 | site-engine — API-only capture backend for the class | deployed | Stdlib-only Go binary forked from idea.uk, live for relojistas.com | vm-backend-sites.md |
| DOC-021 | Pipeline documentation model — derive the topology, author the intent | partial | Step map generated from agent_definitions; PLAN bodies still pending | documentation-system.md |
| STY-022 | D2a — buildTokenAliases renderer-enforced compatibility bridge | deployed | Step-11 post-pass appends missing canonical/alias :root definitions | styling-render-pipeline.md |
| DBG-055 | error_step must live inside step.Config, not at step level | deployed | Step-level error_step silently ignored; dormant instances existed across tool agents | debugging.md |
| CTS-009 | Component creation & regeneration contract | deployed | StoreGeneratedComponentAction create/regenerate branches; regen keyed by LLM-emitted function | contracts-and-standards.md |
| SYS-057 | Reply-to metadata (__work_request__) convention | deployed | Stores request_id + caller's responses topic at receipt time, used at completion | system-architecture.md |
| SPEC-021 | Mission + roadmap as site_specs aspects (strategy-driven intake) | deployed | Strategic context persisted as mission/roadmap aspects, built vonc.com | site-spec-and-classifier.md |
| PLAN-008 | site_specs vs site_plan two-layer architecture + aspect ownership contract | partial | Strategic slow-changing specs vs per-build row-shaped plan tables, with ownership rules | site-plan-and-reconciler.md |
| CVP-001 | Playbook > Strategic Pattern > Component hierarchy (Librarian as brain) | abandoned | Strategy-to-website engine storing business solutions, not just components | conversion-playbooks.md |
| LOCK-006 | auto_lock_on_deploy trigger — assumed live, later found stillborn | abandoned | Strict-mode lock trigger never functionally fired; dropped via migration 009 | locks.md |
| FTW-008 | Response cleaning + SFT negative-example exclusion | deployed | Strip fences, exclude prose edge-cases from SFT, reserve them for future DPO | finetuning-flywheel.md |
| FTW-019 | Three-level evaluation pipeline (L1/L2/L3) | deployed | Structural checks, Claude-as-judge, and spot-check folded into one report | finetuning-flywheel.md |
| SYS-073 | Phased plan P0–P5 (traffic probe) | partial | Structural decisions through off-box collection and a future registry adapter | system-architecture.md |
| LNK-005 | Correct-or-absent principle + loud-but-non-blocking phantom policy | deployed | Structural rule: never fabricate a link target; absence is a warning, not an error | link-management.md |
| SYS-047 | Pages / page_components split | deployed | Structure/workflow in pages; actual rendered content in page_components | system-architecture.md |
| IMG-012 | site_plan_imagery structured plan table (Phase 2G.1) | deployed | Structured per-image plan rows (scope, kind, prompt) succeeding legacy image_prompts dict. | imagery.md |
| DES-015 | Layout: brochure-formal | deployed | Structured, understated, CTA-driven brochure layout with corporate restraint. | design-composition.md |
| DBG-046 | Work-item re-drive and zombie-claim operational semantics | partial | Stuck claims block a whole site; re-drive requires resetting attempt_count + claim metadata | debugging.md |
| IMG-047 | Product illustration pipeline (copyright-safe sketches) (I6) | aspirational | Stylised product sketches to avoid trade-dress exposure from scraped affiliate photos. | imagery.md |
| DOC-011 | doc_plans supersede versioning (one current row, never edit history) | deployed | Supersede tx + partial unique index enforce one current PLAN row | documentation-system.md |
| BIP-010 | prepare_extraction_context / scan_discovery_candidates actions | partial | Supporting Go actions formatting LLM context and scanning for unknown practices | business-intelligence-platform.md |
| CQ-006 | Neutralize-in-place remediation pattern | deployed | Surgical jsonb patch to strip contamination when no clean component restore point exists | content-quality.md |
| CTXE-004 | Instrument-skepticism doctrine | convention | Suspect your own inputs (bundle, query, ground truth) before suspecting the system under test | context-engineering-principles.md |
| ADO-009 | Duplicate sites-row on re-adoption (open investigation) | deployed | Suspected duplicate site row on re-adopting an existing destination domain | adoption-pipeline.md |
| BIP-007 | Geographic area-sweep discovery system | partial | Sweeps every UK postcode district for new vet practices via Firecrawl search | business-intelligence-platform.md |
| CTXE-002 | B4a: the symptom-vs-mechanism retrieval ceiling | deployed | Symptom-based retrieval (lexical or semantic) has a hard ceiling when the cause lives in function-named infra | context-engineering-principles.md |
| PLAN-003 | Two canonicalisation-surfaces divergence (WriteSitePlanAction vs SyncPagesToDBAction) + fix | deployed | Sync ran CanonicalisePage without ValidateRoles, flattening hubs; fixed by unifying pipeline | site-plan-and-reconciler.md |
| BIZ-024 | Deep research domain authority strategy | aspirational | Synthesise primary/authoritative sources for E-E-A-T content moat via a 6-step pipeline | business-strategy.md |
| VKA-002 | vertical_registry table + knowledge-base provenance extensions | aspirational | Table mapping vertical_slug to orchestrators/KB/monetisation; not yet applied to DB | vertical-knowledge-architecture.md |
| RES-003 | research_results with source attribution | deployed | Table storing research findings + full source attribution + expiry, per site/page | research-agents.md |
| DOC-051 | Engines docs tree + single _archive graveyard | aspirational | Target restructure separating engine code/docs/archive | documentation-system.md |
| MDL-002 | Agent model-assignment upgrade sweeps (migration 081) | deployed | Targeted UPDATEs upgrade agent model refs; stale claude-3.x replaced globally | model-infrastructure.md |
| DOC-050 | Bundle-first handoff practice (context packs; broad script vs lean assembler) | deployed | Task handoffs pair problem statement with cmd/bundle invocation | documentation-system.md |
| BIP-004 | collection_tasks queue + batch claiming | deployed | Task queue with SKIP LOCKED atomic batch claiming and orphan-recovery | business-intelligence-platform.md |
| DMR-001 | Chassis deploy-mechanism reference (targets A–F) | deployed | Taxonomy of 6 deploy mechanisms across chassis/sites/idea.uk targets | deploy-mechanics-reference.md |
| DMR-002 | `make deploy-<service>` — single-service deploy with a registry pre-flight | deployed | Deploys ONE named service at $(IMAGE_TAG) and refuses if that tag is not in the registry; `deploy-agents` is all-or-nothing | deploy-mechanics-reference.md |
| DES-014 | Layout archetype library (15/17/18 named layouts) — overview | deployed | Taxonomy of named structural/visual archetypes (brochure-formal, portfolio-kinetic, utility-tool, media-grid,... | design-composition.md |
| LNK-018 | Semantic linking domain decomposition (5 link types) | partial | Taxonomy splitting link work by lifecycle/complexity; most agents unbuilt | link-management.md |
| FIX-042 | F3 learning layer: bug_records + guideline side-tasks | aspirational | Taxonomy/amendment mechanism designed, build status unconfirmed | fix-loop.md |
| ADO-012 | Readopt-as-acceptance-test pattern | aspirational | Tear down and re-adopt as the from-scratch acceptance test after a fix batch | adoption-pipeline.md |
| DBG-053 | rendered_html is a snapshot, not a live view | deployed | Template migrations don't retroactively affect already-built pages' frozen renders | debugging.md |
| AGOV-002 | config_work_items contract | aspirational | Tenant-scoped mirror of site_work_items for config proposals | autonomy-governance.md |
| FIX-046 | F0.3 per-iteration notes / doc_notes reuse | partial | Terminal note wired; per-iteration rows never fully landed | fix-loop.md |
| IMG-060 | Rerender reassembles, it does not re-resolve | partial | Terminal rerender patches stored HTML rather than re-running section resolution. | imagery.md |
| DEV-059 | Work item archival | deployed | Terminal work items older than a configurable age move to site_work_items_archive in batches. | development-guide.md |
| LGL-001 | Liability framework and live legal pages (risk-tiered mitigations, T&Cs) | partial | Terms/refund/privacy live on idea.uk with AI-disclosure wording; solicitor review pending | legal-and-compliance.md |
| FTW-001 | Finetuning flywheel four-lane programme (A/B/C/D) | partial | The A/B/C/D flywheel: export, RAG, LoRA training, Claude-vs-local eval | finetuning-flywheel.md |
| DES-045 | Design fingerprint extraction pipeline (rawHTML → design_reference) | deployed | The Go action (`extract_design_fingerprint`) that parses a crawled site's rawHTML `<style>` blocks, CSS custom... | design-composition.md |
| TL-011 | check_tool_health INNER JOIN blind spot | partial | The Tier-1 tool health check joins content_components to page_components with an INNER JOIN, so a... | tool-lifecycle.md |
| IMP-017 | needs_section_data semantics and the abandoned standalone handler | superseded | The abandoned idea of a dedicated `needs_section_data` handler agent that would fetch list data asynchronously.... | improvement-loop.md |
| DEV-053 | Development Guide (agent-build daily reference) [doc artifact] | superseded | The archived 001 dev-guide doc; has a live successor in docs024_key_docs_latest. | development-guide.md |
| TL-027 | component-quality-auditor auto-regeneration threshold (boundary bug) | deployed | The auditor raises regeneration work items for low-quality components — but its strict `< 50` condition meant... | tool-lifecycle.md |
| IMP-027 | Triage drain loop fix (bounded audit passes + structured findings + section locking) | deployed | The audit→fix→re-audit loop originally ran unbounded, consuming most of the token budget. Fix: auditors emit... | improvement-loop.md |
| IMP-013 | fix_forced_text_colours re-aim: painting classifier + declaration rewriter | partial | The backstop fixer rebuilt around the new paired-variable contract: a `paintClass` classifier... | improvement-loop.md |
| IMP-045 | `needs_rebuild` is inert without an explicit work item | deployed | The build-dispatch-loop reads site_work_items and never scans pages; only write_build_items converts... | improvement-loop.md |
| TP-004 | Commented-out tool route and the planned-tool-page seam | partial | The build-relay's reconcile routing table carries a commented "tool" → tool-build-handler route, so planned tool... | tool-pipeline.md |
| DES-005 | resolved_composition pointer spec + install_site_composition semantics | deployed | The composition install contract: a `css_themes` row is created with all three FKs but empty `css_content`... | design-composition.md |
| DES-079 | Composition resolution architecture: three resolvers + install action (implementation detail of DES-003) | deployed | The concrete Go action sequence behind site-design-planner's composition stage: `validate_composition_inputs` →... | design-composition.md |
| ADO-003 | Site-adoption pipeline: wrapper, Go fingerprint, LLM classify, apply_adoption_plan | deployed | The core 16-step adoption mechanism: crawl, design fingerprint, LLM analysis, write specs+pages+items | adoption-pipeline.md |
| DOC-054 | Concept register and the council-of-concept-experts mission | aspirational | The docs026 programme itself: extract, classify, verify, seed council agents | documentation-system.md |
| IMP-049 | Fleet generalisation doctrine (four rules + artifact verification) | deployed | The doctrine for turning incident fixes into fleet guarantees: (1) fix the writer, not the row — a psql repair... | improvement-loop.md |
| TL-021 | Mandatory minimum tool-suggestion count (2–5, no "suggest zero" option) — superseded | superseded | The earliest tool-suggester design forced the LLM to always propose at least two tools per site. Replaced by an... | tool-lifecycle.md |
| IMP-012 | Improvement-loop colour/nav fixer suite is scheme-blind (pre-re-aim state) | superseded | The established fixer infrastructure before the re-aim: color-variable-fixer runs `fix_hardcoded_colors` (dark... | improvement-loop.md |
| IMP-038 | Heartbeat maintenance model (findings-based, pre-work-items) | superseded | The first full maintenance architecture: K8s CronJob (8h) → agent-chassis spawns maintenance-batch-scheduler →... | improvement-loop.md |
| DES-061 | Paired-variable design direction (curated bg+text pairs, completion of the existing standard) | partial | The fix direction for the section-contrast arc's final generation: a light scheme must be able to render fully... | design-composition.md |
| DES-013 | Composable theme migration 025 (palettes / layouts / typography_sets split from css_themes) | partial | The foundational data-model migration: `css_themes.css_template` conflated palette, typography, and layout... | design-composition.md |
| IMP-037 | Improvement-sweep and build-pipeline-trigger scheduling | deployed | The improvement loop's cadence lives in scheduled_tasks: build-pipeline-trigger (2 min) finds sites with... | improvement-loop.md |
| IMP-047 | Runtime-fill guards & discovery-check wiring gaps (registered-not-enabled / enabled-not-implemented / sweep off) | deployed | The improvement loop's configuration surface has three drift modes: checks registered in Go but named in no... | improvement-loop.md |
| IMP-010 | Improvement-sweep site starvation | deployed | The improvement sweep's site selection starves some sites the same way find_dispatchable_site's arbitrary... | improvement-loop.md |
| IMP-016 | improvement-sweep pause + gated re-enable sequencing | partial | The improvement-loop's triage sweep is deliberately off during core build; the detect→fix loop depends on it, and... | improvement-loop.md |
| CLC-007 | F5 — regen-added required fallback-less fields strand renderability | aspirational | The incident's second facet: a regen also ADDED a required field (Tier-C source, no fallback) that no affected... | component-lifecycle.md |
| CLC-011 | Superseded hypothesis: update_component_html re-renders dependents inline | superseded | The initial working theory held that update_component_html performed an inline dependent re-render (inferred from... | component-lifecycle.md |
| CLC-012 | teaser-reveal-panel: a second component implementing an existing experience pattern | deployed | Teasers that open in place at a shareable URL; native <details> so it works with JS disabled, body always in the DOM so the claims gate can read it, and an item with no body renders with NO control rather than a dead one | component-lifecycle.md |
| TLIB-006 | Tier-D list components: queryresolve + items-array contract (vs numbered-flat fabrication) | deployed | The list-component contract: a Tier-D component sources an `items` array field with `source: query.<name>` (e.g.... | tool-library.md |
| DEV-051 | Workflow lives in default_config, not the workflow columns | deployed | The loader reads default_config; task_workflow/orchestrator_workflow columns are dead for working agents. | development-guide.md |
| TL-015 | Criteria contract v0 (check-type vocabulary + profiles) + browser-runner-adapter design | deployed | The machine-readable criteria schema consumed by Tier 4: `profiles: [desktop, mobile]`; check types... | tool-lifecycle.md |
| SYS-080 | Orchestration-as-identity model (AgentID = PodName) | deployed | The orchestration record, not the pod, is the persistent "agent doing a task" | system-architecture.md |
| DEV-078 | Website build overall plan v0 (first multi-agent website roadmap) | superseded | The original 6-phase/12-step "build a website with agents" plan; origin of the site-building programme. | development-guide.md |
| DES-046 | fpExtractCSSVars regex-based CSS variable extraction (superseded internal bug fix) | superseded | The original design-fingerprint CSS-variable extractor used one whole-stylesheet regex, producing false... | design-composition.md |
| IMP-029 | plan_sections pre-check → plan-then-reconcile evolution | deployed | The original fix for wasteful LLM re-sends on sections with pending needs_section_data was a pre-check that... | improvement-loop.md |
| TLIB-010 | Planned assets-table template/JS split for large tools — superseded plan | superseded | The original plan routed oversized tool templates through the assets table/S3 pipeline. What was actually built... | tool-library.md |
| DES-072 | Legacy monolithic CSS renderer internals (removed) | abandoned | The original renderer held a flat Go struct populated by `extractDesignColors`/`designColorMaps`, loading one... | design-composition.md |
| MDL-004 | RAG pipeline deployment bundle | deployed | The original rollout PR that added ollama-adapter + rag actions + migrations | model-infrastructure.md |
| TLIB-009 | Tag-based deterministic tool-to-site matching (matchToolToSite) — superseded | superseded | The original tool-suggestion mechanism was a deterministic Go function comparing a library tool's semantic_tags... | tool-library.md |
| DES-043 | Palette/typography resolution cascade + the dead-slot bug and fingerprint-fallback hardening | partial | The palette source cascade is design_reference → mission → `design_intent.palette.reference_values` → layout... | design-composition.md |
| TL-017 | Acceptance criteria live in the tool's PLAN (fenced ```criteria JSON block) | deployed | The per-tool definition of *working*. Candidates judged on key/lifecycle/owner and rejected: site_specs (right... | tool-lifecycle.md |
| IMP-039 | Unified build & maintenance work items (site_work_items) | deployed | The pivotal unification: every piece of work — building a page, fixing stale content, adding a tool, publishing... | improvement-loop.md |
| TLIB-012 | JS tools documentation and provenance gap | aspirational | The platform's JS tools have no prose docs and no code-symbol provenance; the only documentation is origin... | tool-library.md |
| DES-001 | Three-layer design system (content_components / css_themes / style_collections) | deployed | The platform's design system has always had three independently-varying layers: Layer 1 self-contained HTML... | design-composition.md |
| CASE-006 | Robot Hands website - first agent-built multi-page site (2025-10) | deployed | The platform's first end-to-end agent site build, proving ground for job topics | site-case-studies.md |
| DBG-007 | Silent-completion / "trust the artefact, not the status" family | deployed | The platform's most recurring failure shape: success status masking no-op work | debugging.md |
| TLIB-021 | In-House Forge — content_components with data-function semantic contract (historical ancestor) | deployed | The platform's own component library from its earliest era: rows with name, function (semantic purpose),... | tool-library.md |
| IMP-011 | Scheme-coherence audit guard (Q8) — abandoned proposal | abandoned | The proposed regression guard: an auditor/improvement-loop check flagging "section scheme does not match site... | improvement-loop.md |
| CLC-005 | Store-driven retry on field-drift rejection (Option B) — abandoned alternative | abandoned | The rejected alternative to Option A (CLC-004): on a field-drift rejection the guard would return the existing... | component-lifecycle.md |
| DES-049 | Composition re-resolve procedure (gated, file-based, backup-first) | deployed | The safe pattern for re-running composition on an already-built site, given that install refuses overwrites... | design-composition.md |
| DES-053 | Per-site style fork chain (palette → css_theme → style_collection) | deployed | The safe pattern for restyling one site without affecting others sharing the same seed collection: clone... | design-composition.md |
| TLIB-020 | Intelligent fallback component matching (P1/P2/P3) — historical | deployed | The site architect resolves each build-plan section against the component library in tiers: P1 perfect function... | tool-library.md |
| DES-036 | Layout-resolution-by-tags gap (classifier not emitting industry_tags) and its migration-008 fix | deployed | The site-design-planner's original layout picker (`resolveLayoutByTags`) intersected a site's classification... | design-composition.md |
| TL-012 | "Completeness + validation passed" ≠ working — twice demonstrated | convention | The standing empirical argument for the behavioural tier: structural/validation checks measure output integrity,... | tool-lifecycle.md |
| TP-003 | Inline-JS extraction ("Path 1" /tools/assets/<fn>.js) — designed, partly real, never on the live deploy path | partial | The store path's separateInlineJS extracts a bare inline `<script>` into js_content, nominally replaced by a... | tool-pipeline.md |
| IMG-064 | Imagery work-item economy end-to-end chain | deployed | The umbrella planner→build→deploy→rebuild chain every imagery phase concept composes into. | imagery.md |
| IMG-065 | Operator asset-amend path (ingest_staged_asset) | built | First human path to replace an asset's bytes: staging BYTEA → validate → S3 new key → in-place assets amend, lock-honouring. | imagery.md |
| DES-011 | chief-strategist (build-plan LLM) + component placement dedup rules — superseded | superseded | The v1/v2 planning agent that produced sections/component_details build plans before build-site-planner existed. | design-composition.md |
| DES-008 | Brand designer agent (theme selection) — earliest design decision point (superseded) | superseded | The very first brand/design decision point in the pipeline's history: an LLM agent analysing domain + objective... | design-composition.md |
| DES-039 | Early "visual identity poles" layout taxonomy (dropped) | superseded | The very first migration draft described layout diversity as nine named "poles" tied to specific reference... | design-composition.md |
| DES-071 | webdesign-agent CSS rendering pipeline (LLM spec → deterministic Go template → git commit) | deployed | The webdesign flow: `analyze_design` (LLM → design-spec JSON: color_scheme/typography/spacing) →... | design-composition.md |
| CGV-013 | Coordinator role (arbitrates and frames) | aspirational | Thin layer above curators owning taxonomy, cross-concern conflicts, human framing | content-governance.md |
| CTXA-015 | Code-indexer agent, index-orchestrator wrapper, and CI-triggered indexing | partial | Thin orchestrator indexes a repo into code_symbols; spawn-wrapped for the GitHub token; CI trigger still queued | context-assembly.md |
| TRF-011 | Intent collection topology (collector under wrapper-orchestrator) | partial | Thin orchestrator spawns a collector worker to pull events/stats from all VM-hosted sites | traffic-analytics.md |
| DIAG-019 | diagnose-orchestrator spawn-wrapper pattern + trigger envelope | deployed | Thin orchestrator spawns the diagnose-agent worker pod; triggered via the generic-request kafka envelope | diagnosis-loop.md |
| SYS-043 | Wrapper-orchestrator pattern (pod lifecycle) | deployed | Thin spawn→call→complete wrapper gives real work its own dedicated Job pod | system-architecture.md |
| FIX-029 | fix-implementer-orchestrator (dedicated-pod wrapper) | deployed | Thin wrapper fixed a shared-pod spawn-gate bypass bug | fix-loop.md |
| SCH-020 | intent-collection-orchestrator + intent-collector agents | partial | Thin wrapper-orchestrator spawning intent-collector, mirrors med-export pair | scheduler-and-tasks.md |
| MCL-006 | Cross-cluster Kafka external listener (nodeport→loadbalancer) | aspirational | Third Strimzi listener lets a second cluster reach primary Kafka, no MirrorMaker | multicluster.md |
| DES-056 | Section-contrast / dark-section model: the four-generation evolution arc | partial | This is a dedicated lineage entry, not a single mechanism — it exists because the same underlying problem (how... | design-composition.md |
| SNAP-002 | component_versions population and change_source provenance | deployed | Three best-effort writers populate version history; lesson on silent best-effort | site-snapshots-and-revert.md |
| DOC-042 | Claim taxonomy: code-checkable / superseded-but-not-wrong / code-invisible | aspirational | Three buckets of doc claims by checkability | documentation-system.md |
| FIX-008 | Dartsonline guides pilot selection history | convention | Three candidates dropped before confirmed differential pilot | fix-loop.md |
| IDEA-005 | Engine implementations: single-shot prompt → Python runner → Go engine | deployed | Three coexisting method implementations; shipped Go engine is stdlib-only, offline-buildable | idea-product.md |
| IMP-023 | Per-site, per-audit-type cadence configuration (maintenance_profile.audit.{type}) — abandoned | abandoned | Three consecutive doc versions carried a designed-but-never-built configuration surface: per-site JSON config... | improvement-loop.md |
| CLC-004 | F1-prompt generation-time field-name preservation (loader + dormant rule + function pin) | deployed | Three coupled pieces so regens preserve names instead of being rejected by CLC-003: (1) load_existing_component... | component-lifecycle.md |
| IMP-004 | Discovery agents architecture (design/quality/completeness) & check registry | deployed | Three domain-scoped analyst agents — quality-discovery-agent (build), design-discovery-agent (design),... | improvement-loop.md |
| SYS-001 | Kafka topic model evolution | deployed | Three eras of topic naming; current model has generic entry, per-spawn job topics, fixed adapter topics | system-architecture.md |
| SPEC-012 | Classifier lineage: v1 Haiku label -> v2 domain_profile -> domain-research-classifier | deployed | Three generations of site classification culminating in current agent | site-spec-and-classifier.md |
| FTW-026 | Flywheel C phase-2 automation architecture (evolution) | superseded | Three generations: HTTP job server → SSH-exec → adapter dispatch | finetuning-flywheel.md |
| HITL-005 | Human direction channels + lock lifecycle + audit-pass cap | partial | Three human-steering channels, content locks, 3-pass audit cap | hitl.md |
| STY-004 | Pre-store component validation gates + planning deferrals + empty-section filter | deployed | Three layers stop broken components/sections reaching pages | styling-render-pipeline.md |
| WDS-004 | Silent completion pathology and the positive-evidence rule | deployed | Three modes marked work complete without evidence it succeeded; fix is positive-evidence-only | work-dispatch.md |
| DBG-004 | Timeout chain ordering contract (claim > call_handler > workflow) | deployed | Three nested timeouts must strictly decrease or duplicate/orphaned handling results | debugging.md |
| NEWS-007 | Feed triage scoring repair (config reads + wrapper unwrap) | deployed | Three stacked bugs left 200+ items unscored; truncation/config/wrapper-unwrap fixes | news-feed-pipeline.md |
| DES-051 | Design/composition flow gaps A–B–C and the plan-time trigger fixes | partial | Three stacked gaps behind themeless/off-palette built sites, investigated as one thread. | design-composition.md |
| SCH-015 | claimed-item-timeout & timeout chain | partial | Three timeouts must stay ordered; two-phase evidence-based/blind reset | scheduler-and-tasks.md |
| STY-038 | HTML action architecture (generate → process → validate) | superseded | Three-action LLM page pipeline, replaced by component-template rendering | styling-render-pipeline.md |
| DOC-031 | Handoff-document discipline (updated-every-turn, supersede chain, turn log) — travelling_docs thread | convention | Three-generation HANDOFF chain with newest-first turn log | documentation-system.md |
| DBI-017 | Database password rotation runbook | deployed | Three-holder password chain rotated in a safe PG→secret→PgBouncer order | database-and-infrastructure.md |
| SYS-025 | Quality Assurance Agent Architecture | partial | Three-layer QA model folded into the main system-architecture doc, not abandoned | system-architecture.md |
| PBP-017 | Sectionless-page durability stack (2b sibling fallback + S1 check + S2 flag) | partial | Three-layer defence against a planned page reaching build with empty sections | page-build-pipeline.md |
| DEV-006 | Standardized input extraction (input_mapping/input_fields/ActionInputSpec, `?` suffix) | deployed | Three-layer input contract with optional `?`-suffixed mapping keys; real site-plan-contamination bug. | development-guide.md |
| SYS-046 | Site / area / page component hierarchy | partial | Three-level slot resolution: area_components → site_components → assembly | system-architecture.md |
| MDL-035 | Per-workflow-step model routing (data-sovereignty) | deployed | Three-tier ai_service resolution lets any step stay in-cluster | model-infrastructure.md |
| CANB-003 | Interactive Biological Explorer + experiment engine (aspirational vision) | abandoned | Three.js/Neo4j explorer + ODE experiment engine; dropped for the practical RAG plan | canine-biology.md |
| DES-020 | Layout: media-grid | deployed | Thumbnail-dominant, continuous-scroll discovery layout — auto-fill fluid grid, optional featured/pinned item,... | design-composition.md |
| FTW-015 | Snapshot economics: setup script beats VM snapshots | convention | Thunder snapshots uneconomic below ~18 runs/month; use idempotent setup script | finetuning-flywheel.md |
| DOC-006 | Interactive HTML runbook checklist | deployed | Tickable HTML companion mirroring markdown runbook steps | documentation-system.md |
| DBG-030 | F2 tiered guard-verification methodology (unit→integration→live fixtures) | deployed | Tier 1/2/3 verification with KEEP/REJECT fixtures; evidence discriminator ordering | debugging.md |
| ADP-002 | Adapter response-header tier taxonomy and validator-coverage gap | partial | Tier-2 routing fields aren't validator-enforced; unfiled tracking issue | adapters.md |
| NAV-006 | Nav quality mechanisms of 2026-04-17 | deployed | Tiered priority, child-page exclusion, label trust, footer quick links shipped together | navigation.md |
| RSH-002 | Composition resolver orphan-rows policy | aspirational | Tolerate cheap orphaned rows from failed installs; sweep via database-cleanup | resilience-self-heal.md |
| RSH-003 | Retry-as-replay: an awaited request is re-sent, never rebuilt | deployed | A retry replays the recorded original; rebuilding it sent the PARENT id, empty body, wrong action | resilience-self-heal.md |
| TLIB-002 | Never load html_template in listing queries (storage discipline) | deployed | Tool/component templates are large; listing and discovery queries must select metadata only, loading... | tool-library.md |
| SYS-085 | Project Manager / User Representative agent hierarchy (abandoned) | abandoned | Top-level PM/user-rep persona hierarchy vanished; review intent moved to HITL | system-architecture.md |
| DBG-003 | LLM step config field-path shadowing (ai_service/max_tokens/temperature) | partial | Top-level ai_service shadows step overrides; misplaced max_tokens silently defaults to ~2048 | debugging.md |
| MDL-022 | LLM step config shadowing bug | partial | Top-level ai_service silently shadows step-level model/max_tokens overrides | model-infrastructure.md |
| DBG-043 | Kafka consumer-group recovery: restart-to-rejoin, never replay-from-earliest | deployed | Topic wipe broke group membership; replay-from-earliest risked duplicate adoptions | debugging.md |
| SYS-072 | Layer-4-build + thin-Layer-5-VM-deploy framing | deployed | Traffic probe reuses the existing build+deploy pipeline, swapping only the target | system-architecture.md |
| MDL-032 | setsid detached launch command | deployed | Training launch backgrounded via setsid so ssh_exec returns immediately | model-infrastructure.md |
| MDL-029 | Flywheel C — LoRA fine-tuning path & iter0 adapter output | deployed | Training pipeline + the first closed-out adapter artefact (828MB) | model-infrastructure.md |
| MCL-012 | Multi-cluster environment re-discovery handoff practice | convention | Treat prior FOCUS docs' IPs/names as illustrative; re-derive live facts each session | multicluster.md |
| SQAM-001 | Three-way split quality-gap diagnostic method (stuck / poor / out-of-scope) | aspirational | Triage method: dispatched-but-stuck vs delivered-but-poor vs never-in-scope | site-quality-audit-methodology.md |
| SPEC-010 | site_type taxonomy drift between classifier and strategist | partial | Two canonical vocabularies for the same concept in one spec chain | site-spec-and-classifier.md |
| ADM-005 | Admin work-item reassign + force-complete override endpoints | superseded | Two narrow overrides replaced by generic PATCH + shared retry/resolve pair | admin-dashboard-and-api.md |
| BLD-014 | Selective rebuild via build_status | deployed | Two orthogonal page-state columns let rebuilds target only marked pages | build-pipeline.md |
| STY-048 | page-rerender mode contract and site-uniformity reconcile pattern | deployed | Two page-rerender modes with different skip semantics; idempotent reconcile scripts | styling-render-pipeline.md |
| STY-049 | missingkey=zero silent-empty-render root pattern + escalate-not-blank guard | partial | Root cause of the image-landing trap (now recovered separately); one call site guarded live, root template behaviour still generic/unpatched | styling-render-pipeline.md |
| STY-050 | Per-site chrome config via a gated input_schema field (config.* -> site_specs) | deployed | Puts a per-site value into SHARED chrome without forking it; live on idea.uk (GTM), 8 co-tenant sites byte-identical | styling-render-pipeline.md |
| RES-006 | Capability watchlist + real-world event watchlist (dual standing research workflows) | aspirational | Two proposed recurring workflows tracking AI capabilities and scheme/event windows | research-agents.md |
| TL-016 | Composer selector invention & the delivered-reality principle (Option B) | deployed | Two recurring failure classes in machine-written acceptance criteria, and their durable remedies. (1) The... | tool-lifecycle.md |
| DIAG-026 | Diagnose loop-back plumbing fault class (state threading + scope encoding) | deployed | Two silent producer/consumer field mismatches left guards and re-scope inert while the loop "worked" | diagnosis-loop.md |
| PLAN-019 | Deferred plumbing stubs: scheduled reconciler tick, domain-aware ensure_pages | aspirational | Two small Phase-1 deferrals; status conflict on whether the tick now exists | site-plan-and-reconciler.md |
| VKA-003 | Research/build cluster separation | aspirational | Two-cluster model separating shared research from per-site build; deferred | vertical-knowledge-architecture.md |
| ADM-002 | Admin API current state: dual-auth gateway, inventory, and fix blocks | partial | Two-service gateway audit: concrete bugs + hardcoded values, fixes sequenced A-F | admin-dashboard-and-api.md |
| ADO-027 | tool-recreation-handler | deployed | Two-stage analyze_tool/recreate_tool JS-heavy page recreation | adoption-pipeline.md |
| DOC-061 | API documentation system (OpenAPI external + per-service internal API.md) | unknown | Two-tier API doc practice, predates vonc corpus, unverified currency | documentation-system.md |
| CTS-046 | API documentation convention (OpenAPI + internal API.md) | deployed | Two-tier docs: external OpenAPI spec + internal per-service API.md with CI lint | contracts-and-standards.md |
| DBG-041 | Convergence inertness: []map[string]interface{} vs []interface{} assertion | deployed | Type assertion always failed; whole convergence feature silently dead since deploy | debugging.md |
| ENT-001 | Entity data agent family (structured data drives pages) | aspirational | Typed JSONB entities with state-based lifecycle drive template-rendered pages | entity-data.md |
| CGV-018 | content_items reusable content layer | aspirational | Typed reusable content rows (headline/tagline/faq) built but apparently never written to | content-governance.md |
| NEWS-018 | News feed pipeline: content_sources and feed-item lifecycle schema | deployed | Typed source configs and the ingested->published/expired/duplicate item lifecycle | news-feed-pipeline.md |
| AGOV-006 | Change-layer integration (change_events, in_band) | aspirational | Typed triggers fan out from a change-event table; in_band closes self-mod loop | autonomy-governance.md |
| CTS-033 | Adapter response envelope contract (single-sourced) | deployed | Typed-struct Kafka envelope resolved empirically; now single-sourced in 035_adapter_guide | contracts-and-standards.md |
| STY-001 | Styling render pipeline reference: two assembly paths and the scheme gap | deployed | Umbrella finding: CSS render and section render are separate paths meeting only in the browser | styling-render-pipeline.md |
| SYS-038 | Autonomous Build-and-Operate — trust-not-capability thesis | aspirational | Umbrella vision bounding LLM uncertainty to progressively remove the human | system-architecture.md |
| PBP-016 | Save-failure visibility fix (mark_save_failed) + engine error_step ambiguity | aspirational | Unbuilt fix to surface save failures instead of laundering into complete_error | page-build-pipeline.md |
| DBG-038 | Pod label agent-type (hyphen) vs log field agent_type (underscore) | deployed | Underscore selector silently matches zero pods; type-wide selectors mix in stale runs | debugging.md |
| SYS-024 | Snapshot-shadowing agent-definition loader defect | deployed | Unfiltered ORDER BY version reads let snapshots silently shadow live agent rows | system-architecture.md |
| VET-002 | Unified polymorphic products/product_prices schema (kind discriminator) | aspirational | Unifies business_prices+product_prices under products via a kind column | vet-med-pricing.md |
| DBI-014 | Database cleanup and log retention policy | deployed | Uniform per-table retention functions with distinct success/error windows | database-and-infrastructure.md |
| WII-001 | Silent-completion failure family ("work reports success but doesn't happen") | partial | Unifying defect class: several ways a work item reaches complete without the artifact existing | work-item-integrity.md |
| PLAN-016 | Union-clobber bug and the carry fix | deployed | Union of adopted pages wiped sections/meta/nav_order; fixed by carrying fields forward | site-plan-and-reconciler.md |
| FIX-021 | Reframe step (post-veto) | partial | Unit-tested reframe path never fired live since v4 | fix-loop.md |
| SPEC-016 | Feasibility / blocked-handler pattern | partial | Unknown handlers block work items; feasibility-recheck promotes them later | site-spec-and-classifier.md |
| VMB-008 | Dedicated vs shared box policy and VM sizing | deployed | Unknown-traffic experiments get their own box; low-traffic domains share one | vm-backend-sites.md |
| BIP-008 | Discovery candidates + promotion pipeline | partial | Unmatched search finds staged as candidates, then promoted into businesses table | business-intelligence-platform.md |
| HITL-016 | process_approval_decision and rejection routing | deployed | Unpacks approval decision into CollectedData; branches continue/stop/reject | hitl.md |
| CTS-021 | is_dark_section demoted to catalogue metadata | deployed | Unreliable self-declared flag; styling must derive from what CSS actually paints | contracts-and-standards.md |
| RSN-012 | Self-development coding pipeline — positions A/B/C | aspirational | Unresolved cross-area coordination model; lean toward spawn-fresh mediator | reasoning.md |
| CTXA-019 | cmd/bundle robustness contract (validate early, fail loud, manifest input) | partial | Validate cheap inputs before slow gathers; fail loud on a missing path rather than silently omitting context | context-assembly.md |
| ADO-025 | Adoption faithfulness - WriteSitePlanAction identity strip | partial | ValidateRoles/CanonicalisePage interaction strips identity on some page types | adoption-pipeline.md |
| DBI-018 | sites.status vocabulary and the blast-radius filter trap | deployed | Validated status vocabulary; heuristic against scoping blast-radius queries on 'active' | database-and-infrastructure.md |
| DEV-034 | call_agent contract validation vs input_data.spec.* convention (dual-placement) | deployed | Validator checks only top-level required keys while handlers read input_data.spec.*; needs both. | development-guide.md |
| ONB-004 | Progressive onboarding — a ramp, never "done" | aspirational | Value from mechanical layer first; onboarding never terminates | onboarding-config.md |
| CTXA-011 | code_symbols: the per-repo code index (pgvector sibling table) | deployed | Vector+trigram code index keyed (repo,path,symbol), SHA-versioned, HNSW, pruned hard on reindex | context-assembly.md |
| DIAG-010 | data_requests channel: model-authored read-only SQL gather | deployed | Verdict emits read-only SELECTs; diagnose_route/load_runtime execute them in a READ ONLY transaction | diagnosis-loop.md |
| SYS-081 | Optimistic locking on orchestration state | deployed | Version-column CAS design specified but unconfirmed as shipped | system-architecture.md |
| DBI-019 | training_exports Postgres schema | deployed | Versioned ChatML dataset snapshots in Postgres TOAST instead of S3 | database-and-infrastructure.md |
| VONC-003 | provocations.json data contract (today / lobby / arena / archive) | deployed | Versioned JSON feed contract for Spark's runtime-fill sections; v3 live | vonc.md |
| TL-025 | Component versioning (component_versions table) — schema-mode origin | deployed | Versioned snapshots of component templates (html_template, css_template, input_schema per version_number),... | tool-lifecycle.md |
| DES-029 | Layout: industry-hub | deployed | Vertical information-authority layout — "About this site" independence-claim banner,... | design-composition.md |
| DOC-057 | Objective tree vs concern tree (two orthogonal axes) | aspirational | Vertical mission tree kept separate from horizontal concern tree | documentation-system.md |
| TPI-003 | Cross-domain intelligence network and subscription tiers | abandoned | Vision of sibling-domain intelligence sharing and paid subscription tiers; pure vision | topic-intelligence.md |
| CGV-008 | Optimistic-lock co-management of shared rows across parallel chats | deployed | WHERE updated_at=<last-known> UPDATE; 0 rows means stop and coordinate | content-governance.md |
| ADO-036 | Vertical-slice dogfooding of the automation ratchet (category mismatch) | aspirational | Walk one capability end-to-end before generalising the ratchet | adoption-pipeline.md |
| DES-022 | Layout: soft-editorial | deployed | Warm, reading-first, organic layout — tinted background, pill-shaped buttons, barely-there card borders, serif... | design-composition.md |
| STY-023 | D2b — canonical-token prevention (contract rule 11 + lint) | partial | Warn-only lint + contract rule stop new orphan tokens at the source | styling-render-pipeline.md |
| RES-002 | research-agent (cited web research into research_results) | deployed | Web-search specialist: search→scrape→synthesise→cite, spawned by writer/classifier | research-agents.md |
| IMP-028 | wont_fix/superseded dedup and needs_section_data data-honesty pattern | deployed | When a recurring issue is detected while an older item is stuck, the loop creates a new item and marks the old... | improvement-loop.md |
| DES-042 | Palette merge rule: core slots vs specialised slots | deployed | When a site composes a theme, core palette slots let the site's own spec (or, in the later two-stage pipeline,... | design-composition.md |
| DEV-050 | Real-rows-beat-prose-or-assumption discipline | deployed | When doc prose is ambiguous, a real live row/file is the source of truth, not inference. | development-guide.md |
| CLC-006 | F4 — regen-vs-create keyed on the LLM-chosen function (silent fork) | partial | Whether a store is a regeneration or a creation depends on whether the LLM happened to choose the existing... | component-lifecycle.md |
| ORG-001 | Organizational framework (roles, listeners, policy-as-filters) | abandoned | Whole-company modelling thought experiment: roles, listeners, policy filters | org-framework.md |
| CTXK-011 | diagnose (cmd/diagnose): the CLI dev/test harness | deployed | Wires the scaffold to real gatherer/call-graph adapters with a stubbed (non-model) verdict step | contextkit-toolchain.md |
| CGV-005 | Human direction channels and the pinned direction spec | partial | Work-item / direction-update / reference-suggestion channels; direction resets audit pass | content-governance.md |
| LNK-017 | prepare_link_context available_pages gap on the work-item path | partial | Work-item rebuild path leaves the LLM's link-context constraint empty | link-management.md |
| SYS-026 | site_work_items domain → pipeline column rename | deployed | Work-routing column renamed to eliminate collision with the site's own domain name | system-architecture.md |
| DEV-079 | Data-path resolution problem (agent vs local action nesting) | superseded | Workflow config referenced CollectedData paths that didn't match where actions actually stored results. | development-guide.md |
| DEV-028 | Deploy-ordering hard gate for coupled Go action + workflow-config changes | deployed | Workflow jsonb is live instantly; wiring to a not-yet-deployed action breaks every run of the agent. | development-guide.md |
| SYS-078 | Local vs remote actions and the action registry | deployed | Workflow steps run synchronously in-process or dispatch to another agent's topic | system-architecture.md |
| IMG-044 | Content-linked card imagery (Phase I3) | aspirational | Would give every linking card an image re-cropped from its article's own generated asset. | imagery.md |
| IMG-022 | Visual auditor imagery awareness (Phase 4, text-only) | aspirational | Would give visual-design-auditor a sixth IMAGERY check category from text context only. | imagery.md |
| IMG-021 | Adoption image mirror (Phase 3) | aspirational | Would persist crawled imagery as assets instead of discarding it after adoption. | imagery.md |
| BLD-011 | page-build-handler (content-page handler with section planning and validation gates) | deployed | Wrapper adding plan_sections/validate/save/deploy around the content-writer specialist | build-pipeline.md |
| FIX-011 | Two intake paths disagreement | deployed | WriteBuildItemsAction and reconcile_site_plan skip different page types | fix-loop.md |
| CGV-028 | site_specs `pinned` flag not honoured by the write path | partial | WriteSiteSpec ignores pinned; only disabled improvement-sweep currently protects specs | content-governance.md |
| PBP-004 | Array item-fields prompt contract (019 migration + ItemFields) | deployed | Writer prompt now lists array element shape so the LLM doesn't guess mismatched item keys | page-build-pipeline.md |
| PBP-015 | Index stale-rebuild defect (writer output ≠ save input path) | deployed | Writer's compiled result was silently replaced by a size-limit stub before save | page-build-pipeline.md |
| SYS-033 | extractResponseContent flat-string hypothesis (superseded) | superseded | Writer-can't-populate-structured-fields hypothesis disproven by an isolated build test | system-architecture.md |
| LNK-014 | select_sections path-mismatch bug (phantom CTA root cause) | deployed | Wrong JSON path silently discarded resolver output; one-line jsonb_set fix | link-management.md |
| DBG-011 | CrashLoop exec "./X" — image lacks the binary | deployed | Wrong/stale Docker image content; "no guard between built and running" recurred 3x | debugging.md |
| PBP-020 | complete_error silent-success family (page build completes having built nothing) | partial | Zero-ready-sections routes to a step literally named complete_error, masking failure as success | page-build-pipeline.md |
| SCH-007 | CTE-only scheduled tasks pattern ("Always Return a Row" rule) | deployed | Zero-row pre_query silently stalls last_triggered_at/last_completed_at | scheduler-and-tasks.md |
| DOC-040 | Doc claim-verification / dated-claim convention | convention | [checked YYYY-MM-DD] tags on falsifiable claims; whole-doc stamps banned | documentation-system.md |
| MDL-020 | agent_definitions backup naming convention | superseded | _preNNN suffix ties a backup to its guarding migration; never-drop rule | model-infrastructure.md |
| DES-034 | Phased belt-and-braces removal plan for webdesign-agent install_theme (abandoned same-day) | superseded | `026_design_and_site_planner_v1.md` proposed a cautious two-phase removal of webdesign-agent's defensive... | design-composition.md |
| STY-016 | Exact-field-name template binding with silent empty on miss (RenderTemplate) | deployed | `<no value>` strip is why renamed/missing fields fail silently | styling-render-pipeline.md |
| DES-073 | css_templating.go theme-forking bridge (known-broken legacy path) | partial | `TemplateCSSFromSpec` converts a rendered CSS snapshot into old flat-field-name placeholders and writes it to... | design-composition.md |
| DES-054 | Deterministic contrast gate missing on specialised palette slots | aspirational | `color_util.go` has correct WCAG code (`relativeLuminance`, `wcagContrastRatio`, `pickReadableOnBackground`),... | design-composition.md |
| TL-003 | Two divergent tool-creation paths (novel vs fork) | deployed | `create_tool_component_action.go` (the "novel" path) never sets pages.sections, leaving it default `[]`;... | tool-lifecycle.md |
| DES-044 | design_reference vs design_intent spec-aspect model: extraction, three-way priority, palette-lock policy | deployed | `design_reference` holds concrete values (hex colours, font stacks, CSS variables, spacing, a... | design-composition.md |
| TLIB-007 | create_tool_component updates in place by function; unique index covers active library originals | deployed | `idx_cc_tool_function_unique` = UNIQUE(function) WHERE component_level='tool' AND forked_from IS NULL AND... | tool-library.md |
| IMP-043 | Flywheel B — RAG knowledge base with nomic task prefixes | deployed | `knowledge_base` pgvector(768) table read/written by rag_lookup/rag_index actions on the cpu-ollama... | improvement-loop.md |
| DES-069 | palettes table / seed (CSS-theme-extracted colour slots) | deployed | `palettes` stores one row per design palette (`name`, `display_name`, `colours` JSONB slot map, `category`,... | design-composition.md |
| IMP-005 | Blog listing rebuild and slot-detection strategy | deployed | `rebuild_blog_listing` runs in rerender-pages before get_pages: finds the actual listing slot via a priority list... | improvement-loop.md |
| DES-032 | Renderer theme-resolution cascade and the emergency fallback | deployed | `render_css_from_spec` resolves theme by `config.theme_id` → `config.theme_name` → `sites.style_collection_id`... | design-composition.md |
| DES-070 | typography_sets table / seed (6 named font/scale bundles) | deployed | `typography_sets` stores 6 named bundles — sans-modern, serif-editorial, display-bold, mono-technical,... | design-composition.md |
| IMP-026 | Audit finding dedup + blocked-item filtering algorithm (write_audit_findings) | deployed | `write_audit_findings` was documented as implementing three dedup/safety layers: bulk-preloading blocked item... | improvement-loop.md |
| DEV-033 | Manual agent trigger via kcat orchestrate envelope (never hand-roll spawn+call) | deployed | action=orchestrate to system.agent.generic.requests is the proven manual-trigger mechanism. | development-guide.md |
| CTS-037 | Input/output contracts on agent definitions | deployed | agent_definitions.input_contract/output_contract now enforced at call-site, not just docs | contracts-and-standards.md |
| ADR-001 | Agent definition snapshot/revert via backup table | deployed | agent_definitions_backup table; snapshot_agent/revert_agent eliminate wrong-row bug | agent-definition-registry.md |
| DEV-066 | Agent groups — versioned project-recipe teams (discovery, versioning, pinning) | partial | agent_group_definitions are immutable-versioned project recipes; EvolutionService appears aspirational. | development-guide.md |
| STY-002 | CSS assembly pipeline (composable theme → styles.css) | deployed | analyze_design → render_css_from_spec (deterministic) → deploy_css → CDN sync | styling-render-pipeline.md |
| NEWS-003 | Real-time-search news providers (Grok Responses API decision) | deployed | api_news routes to Grok/OpenAI/Perplexity real-time search after chat-completions hallucinated URLs | news-feed-pipeline.md |
| IMP-046 | build_status 'approved' invisibility defect and its layered fleet fix | deployed | apply_section_edit left an edited live section at build_status='approved' while every discovery check filters... | improvement-loop.md |
| TRF-002 | Wayback grounding of probe pages | deployed | archive.org snapshot fixes vertical/language/invited-action before building a probe page | traffic-analytics.md |
| SYS-028 | Asset self-resolving storage URI (dispatch loop) | superseded | asset-deployer resolves its own s3:// URI from asset_id instead of orchestrator pre-resolving | system-architecture.md |
| IMG-008 | Asset locking mirrors page_components (Phase 2A) | deployed | assets gains locked_at/lock_type so audits/discovery skip locked (e.g. approved) assets. | imagery.md |
| IMG-053 | Presigned-URL expiry and deploy-time asset localisation (Edit F) | deployed | assets.url presigned links died after 7 days; deploy now records the durable local path. | imagery.md |
| SYS-069 | Gateway proxy pattern (auth-service → core-manager) | deployed | auth-service is the only HTTP ingress; core-manager re-validates JWTs independently | system-architecture.md |
| SAAS-002 | Conversational build-intake via briefing-agent chat | aspirational | briefing-agent chat intake hands off to intake-orchestrator to kick a build | saas-isolation-architecture.md |
| ONB-022 | Per-builder briefing questionnaires | deployed | briefing_questionnaire JSONB per builder agent; fetch_agent_questionnaire action | onboarding-config.md |
| MDL-023 | Extended thinking configuration | deployed | budget_tokens enables Anthropic extended thinking; strips temperature | model-infrastructure.md |
| DEV-011 | Extended thinking config and the no-temperature-to-Anthropic rule | deployed | budget_tokens enables extended thinking; Anthropic client sends no temperature since 2026-05-27. | development-guide.md |
| SCH-001 | Build pipeline trigger: 30s heartbeat, fire-and-forget, one item per dispatch orchestration | deployed | build-pipeline-trigger seeds queue, picks one dispatchable site per tick | scheduler-and-tasks.md |
| ADO-023 | Adoption interactivity misroute - canonical-prefix key desync | deployed | buildPageFeatureMap keyed raw not canonical name, missing tool pages | adoption-pipeline.md |
| NAV-003 | Stale pages polluting nav + config-driven deactivation fix | deployed | build_status filter + deactivate_stale_pages flag close the stale-nav gap | navigation.md |
| DEV-004 | spawn→call pattern and role-based targeting | deployed | call_agent finds a spawned agent via target_role scanning CollectedData spawn results. | development-guide.md |
| SYS-083 | Agent-centric architecture: steps call agents, not topics | deployed | call_agent with agent_type is the primary abstraction over raw topic addressing | system-architecture.md |
| DEV-014 | agent_definitions three-column semantics (category/agent_category/status) | deployed | category is free-text, agent_category is CHECK-constrained (no 'orchestrator'), status is lifecycle. | development-guide.md |
| TP-001 | Tool pipeline end-to-end (suggest → route → generate/fork → cross-link → rewrite → improve → audit) | deployed | check_missing_tools / missing_tools discovery check auto-seeds add_tool items → tool-suggester (LLM judgment over... | tool-pipeline.md |
| IMP-002 | Audit enforces intent, doesn't override (chain of authority) + propose mode for spec-less sites | deployed | classifier decides intent → planner implements → composition installs → webdesign renders → audit checks build vs... | improvement-loop.md |
| DBG-057 | Code-retrieval corpus staleness masquerading as retrieval-quality problem | deployed | code_symbols index built from a year-old stale checkout, not a retrieval bug | debugging.md |
| DIAG-033 | Static-tier corpus gaps: workflow-JSON invisibility + error-log enrichment | partial | code_symbols indexes only .go files; agent_definitions workflow JSON is invisible without explicit enrichment | diagnosis-loop.md |
| ADP-004 | Adapter deployment essentials & troubleshooting checklist | deployed | command vs args, KafkaTopic CRDs, RBAC globs — real thunder-adapter deploy lessons | adapters.md |
| DEV-048 | Untested-code / behaviour-testing discipline | deployed | compile/gofmt/vet prove syntax not behaviour; destructive CLI ops must default to report-only. | development-guide.md |
| DEV-024 | Child-result shaping: output_fields (plural) contract | deployed | complete step's output_fields (plural) governs result shape; singular output_field is silently ignored. | development-guide.md |
| FIX-022 | Escalation as first-class success terminal | deployed | complete_escalated treats architecture dead-ends as success | fix-loop.md |
| CLC-009 | Component versioning via component_versions (and unversioned-write provenance) | partial | component_versions snapshots (component_id, version_number, schema, template, change_description, changed_by,... | component-lifecycle.md |
| DOC-019 | NOTES-at-every-fix hook on the three fix agents | deployed | compose_note → append_note wired on fixer/improver/recreation success paths | documentation-system.md |
| DOC-018 | PLAN-at-birth write hook in tool-generator | deployed | compose_plan → write_plan → index_plan after save_tool succeeds | documentation-system.md |
| DBG-008 | save_page_sections: sole writer, HTML-fallback bug, guard laundered to success | partial | content-regression guard's refusal routed through a SUCCESS-labelled complete_error step | debugging.md |
| CGV-020 | Section governance columns: content_brief and suppressed_sections | deployed | content_brief enables regeneration; suppressed_sections stops discovery resurrecting removals | content-governance.md |
| CTS-005 | Component naming contract (function = canonical kebab ID) | deployed | content_components.function is kebab, unique-per-active-row, matches data-component attr | contracts-and-standards.md |
| CGV-014 | vonc.com mini-lobby content-edit re-render scope-boundary question | deployed | content_data-only edit rule established; correct path for a structural trim left unclear | content-governance.md |
| NEWS-008 | News pipeline replication and the news enrichment pattern | deployed | content_sources rows as a pure-data replication template for adding news to a new site | news-feed-pipeline.md |
| ADM-003 | Core-manager API server surface (spec pin/unpin among admin routes) | deployed | core-manager exposes spec pin/unpin, keeping Pattern B lock semantics alive | admin-dashboard-and-api.md |
| SYS-054 | ExecutionContext unified message envelope and ID semantics | deployed | correlation/orchestration/request id semantics; sender constructs, receiver trusts | system-architecture.md |
| OPD-001 | Standing evidence rules (working-method contract) | deployed | correlation_id reads, snapshot-before-UPDATE, 0-rows skepticism | operating-doctrine.md |
| TL-010 | Canonicalise tool page identity across surfaces (T3) | aspirational | create_tool_component and deploy_tool build page name/url/page_type ad hoc, diverging from the canonical... | tool-lifecycle.md |
| TL-022 | forked_from NULL collision risk on novel tools | unknown | create_tool_component omits forked_from, so novel/generated tools are classified as library tools by the partial... | tool-lifecycle.md |
| CGV-003 | Spatial addressing for natural-language editing | partial | data-pc-id/slot/position shipped at section level; element-level addressing unfulfilled | content-governance.md |
| DOC-046 | Documentation archiving subproject (docs019 cleanup) | partial | dedup+thin_versions+staged migration plan; manifests prove partial execution | documentation-system.md |
| DBG-002 | Agent is a DB row; trust default_config over prose | deployed | default_config.workflow is the real behaviour; description can lie; two possible source DBs | debugging.md |
| NEWS-012 | files_field vs content_field git_commit deploy bug | deployed | deploy_page misconfigured field silently dropped all component JS from git since inception | news-feed-pipeline.md |
| IMG-040 | Brand-head derived assets (favicon + OG card) | deployed | derive_brand_head_assets deterministically derives favicon/OG card from locked logo. | imagery.md |
| DOC-010 | Travelling documentation (PLAN + NOTES) in Postgres | deployed | doc_plans/doc_notes tables; umbrella system, adopted by other workstreams | documentation-system.md |
| DOC-001 | Documentation consolidation system (numbered canonical docs + index) | deployed | docs024 canonical set, consolidation notes, version families closed by full diffs | documentation-system.md |
| ONB-001 | Domain submission tiers and mission/roadmap briefs (domain-submitter entry point) | deployed | domain-submitter entry point; three tiers up to mission/roadmap briefs | onboarding-config.md |
| FIX-014 | Two-reviewer council (F2.1) | deployed | edit-quality + guardian reviewers, guardian holds veto | fix-loop.md |
| TRF-010 | intent_events table with structural idempotency | deployed | engine_event_id UNIQUE + ON CONFLICT makes overlapping collector pulls safely idempotent | traffic-analytics.md |
| IMG-056 | ensureAssets scope gap: hero/logo-only surfacing (Edit B / kind-alias) | deployed | ensureAssets only surfaced hero/logo; extended to section/illustration scope across 3 sites. | imagery.md |
| CTS-001 | Page-build-handler pipeline (plan_sections Layer 0 + validate_content) | deployed | ensure_site_record→plan_sections triage→writer→validate_content→save/deploy pipeline | contracts-and-standards.md |
| DOC-020 | "Docs never fail the work" containment principle — and its limit | deployed | error_step containment covers errors, not crashes/stalls | documentation-system.md |
| DEV-031 | error_step mechanics (config-level placement, existing target, loop corollary) | deployed | error_step must be nested inside step.Config; a step-level sibling key is silently ignored. | development-guide.md |
| ADO-008 | Firecrawl capability escalation ladder | aspirational | executeJavascript/waitFor/structured json as upgrades to plain rawHtml parsing | adoption-pipeline.md |
| DEV-064 | Prompt resolution priority hierarchy | deployed | execute_llm_prompt resolves prompt: caller override > agent's own prompt_template > workflow fallback. | development-guide.md |
| FIX-006 | Retention/expiry knob on diagnosis_artifacts | aspirational | expires_at/pinned columns exist; no sweep job built | fix-loop.md |
| SYS-062 | Fan-out and awaited-response correlation | deployed | fan_out step dispatches parallel sub-tasks matched back via causation_id | system-architecture.md |
| NEWS-015 | rebuild_blog_listing does not handle news-index pages | partial | findBlogPage never matches news-index pages, so news-only sites silently no-op | news-feed-pipeline.md |
| WDS-002 | Dispatch chain + NOT-EXISTS whole-site claim blocker + one-site-per-tick throughput | deployed | find_dispatchable_site excludes any site with a claimed item; picks one site arbitrarily per tick | work-dispatch.md |
| DGH-002 | Git-adapter non-fast-forward commit race (shared sites repo) | aspirational | force:false + no retry on updateRef can silently lose a concurrent multi-site commit | deployment-github.md |
| CASE-011 | Idea generation method - versioned pipeline (v0 -> v3) | partial | generate->cut->verify->score->rank pipeline refined across four versions | site-case-studies.md |
| DBG-040 | Untracked-file deploy trap: verify by ancestry, not tag/commit message | deployed | git commit -a misses untracked files; verify via git merge-base --is-ancestor | debugging.md |
| CHAT-005 | Provider-agnostic worker (deps adapters) | aspirational | handleChat core + ContextStore/LLMClient/TurnSink adapters; Cloudflare-first, portable | site-chatbot.md |
| STY-005 | Scheme-to-components P0: light-resolved site renders dark | deployed | idea.uk deployed dark chrome/sections despite resolving light; fixed via paired-variable standard | styling-render-pipeline.md |
| CASE-013 | Go engine supersedes Python reference implementation | superseded | idea.uk engine ported Python->Go to match the rest of the Go-throughout platform | site-case-studies.md |
| SYS-008 | Idle timeout for spawned agents + topic cleanup strategy | deployed | idle_timeout_seconds env var exits idle pods; CronJob + Kafka retention clean up orphan topics | system-architecture.md |
| ADP-001 | Adapter/response message envelope contract (normative) | deployed | in_response_to_request_id + typed bool headers + ProduceWithValidation, or replies silently vanish | adapters.md |
| CTXA-012 | Hybrid code retrieval: index/lookup_code_symbols | deployed | index/lookup_code_symbols actions mirror rag_index/rag_lookup with RRF fusion moved into SQL | context-assembly.md |
| DEV-023 | input_mapping semantics: call_agent-only vs local-action/loop dot-paths | deployed | input_mapping is dead config on plain local action steps; only call_agent and loop fan-out honour it. | development-guide.md |
| DBG-019 | Discovery-checks list maintenance and the workflow-replace landmine | deployed | jsonb array append is safe; whole-workflow jsonb_set is a future silent-erase risk | debugging.md |
| DBG-018 | Kafka trigger payload discipline: multi-line kcat bodies mis-route | deployed | kcat -P is line-delimited; multi-line JSON silently mis-pairs headers/body | debugging.md |
| SYS-035 | Generic orchestrate envelope as universal manual trigger | deployed | kcat-produce shape for hand-running any agent via the generic entry point | system-architecture.md |
| DEV-042 | Development-guide gotcha: BST/UTC timestamp mismatch | deployed | last_activity lacks time zone while created_at has one; NOW()-last_activity math is silently wrong. | development-guide.md |
| NEWS-013 | Two distinct news components as a multi-view pattern | deployed | latest-news and news-listing are separate components, template for future filtered views | news-feed-pipeline.md |
| CTS-002 | Component input-schema source vocabulary (Tier A-D + renderer, proposed E) | partial | llm/static/site-data/query source tiers for component fields; feed.* Tier E undecided | contracts-and-standards.md |
| LOCK-004 | Timed lock-expiry project and lock-model coherence plan | partial | lock_type/lock_expires_at added via migration 115 (schema only); Go predicate sweep pending | locks.md |
| LOCK-001 | Pattern A (locked_at/locked_by) canonical; Pattern B (pinned) dead | deployed | locked_at/locked_by is the uniform lock mechanism; pinned boolean never wired | locks.md |
| CTXK-012 | internal/diagnose scaffold package (file-level map) | deployed | loop.go/step.go/advance.go/callgraph.go/verdict_wire.go/sqlguard.go implementing the tested loop scaffold | contextkit-toolchain.md |
| ATN-001 | Agent hierarchy tree navigation (ltree paths + subtree summaries + live viewer) | aspirational | ltree tree_path + subtree summaries + REST/WebSocket viewer for massive trees | agent-tree-navigation.md |
| DGH-004 | Site manifest + external-edit desynchronisation detection | aspirational | manifest.json + git webhook would flag human edits and halt agent overwrites | deployment-github.md |
| SYS-022 | Chassis config location bugs | partial | max_tokens shadowing, dead step-level temperature, dropped error_step field | system-architecture.md |
| DBG-044 | Manual work-item insertion as an operational rebuild lever | deployed | needs_page/needs_content_page hand-inserts are claimed normally by dispatch | debugging.md |
| PBP-001 | Rebuild vs rerender semantics and stale-render fossilisation | deployed | needs_rerender reassembles stored HTML without re-rendering templates; only rebuild does | page-build-pipeline.md |
| SYS-003 | Orchestration state and collected_data as the workflow data bag | deployed | orchestration_states row holds workflow_plan/collected_data/current_step/status | system-architecture.md |
| DBG-034 | EXECUTING_STEP frozen forever means the worker pod died (OOMKill triage) | deployed | orchestration_states written by the worker; a dead pod freezes the row, not a hang | debugging.md |
| IMP-019 | mark_item_failed error honesty (flag-before-complete) | deployed | page-build-handler's step-level error routing pointed at `complete_error`, a SUCCESS-labelled complete_workflow —... | improvement-loop.md |
| PRC-001 | Mega-prompt fragility and candidate replacement patterns | aspirational | page-content-writer's 6KB mega-prompt flagged as technical debt; five replacement patterns proposed. | prompt-composition.md |
| WII-007 | Positive-evidence deploy guard (0-component page never marked deployed) | deployed | pageHasComponents gates the deployed status write, keeping build_status trustworthy | work-item-integrity.md |
| IMP-009 | Component linking enrichment saga (component_id NULL on rebuilt pages) | deployed | page_components.component_id was wiped on every rebuild because sections_metadata from the content writer carries... | improvement-loop.md |
| LNK-002 | Internal linking machinery and its defects | partial | pages table as link-target authority, catalogued against known defects | link-management.md |
| STY-020 | Assembly membership and chrome model (page_components by position) | deployed | pages.sections is metadata only; three coexisting head shapes confuse forensics | styling-render-pipeline.md |
| STG-001 | Storage: per-call S3 client construction is canonical | deployed | params.StorageClient is unreliable (nil at startup); construct per action instead | storage-architecture.md |
| PAY-003 | Entitlement gate architecture (build-submission + maintenance-run gates) | aspirational | pending_entitlement hold plus a maintenance heartbeat join-filter; both unbuilt | payments.md |
| CTS-026 | Anti-fabrication content path (llm_field_specs, merge_with) | deployed | plan_sections resolves query data pre-LLM; RenderComponentAction overlays it as authoritative | contracts-and-standards.md |
| CTS-025 | Numbered-flat-fields anti-pattern (25 components) | partial | postN_title schemas force LLM fabrication; only game-list migrated to Tier D so far | contracts-and-standards.md |
| SCH-006 | pre_query SQL-worker/gate pattern (one message per tick, not fan-out; self-healing tasks) | deployed | pre_query is gate + SQL worker; scheduler never fans out per row | scheduler-and-tasks.md |
| FTW-028 | training-launcher real workflow | deployed | presign→manifest→detached SSH launch→mark_running, full path ~26s | finetuning-flywheel.md |
| DEV-067 | Workflow selection priority (inline override > group > agent default) | deployed | processor.selectWorkflow's three-tier priority: inline message config > group workflow > agent default. | development-guide.md |
| CTS-032 | query.{name} field-source resolution timing | superseded | query.* resolution moved from render-time to plan_sections-time | contracts-and-standards.md |
| PLAN-030 | queryresolve reality-vs-invention architectural promise | deployed | queryresolve enforces the boundary between LLM-authored content and DB-derived facts | site-plan-and-reconciler.md |
| DOC-027 | tool_docs knowledge-base indexing of PLANs (rag_index derived index) | deployed | rag_index chunks/embeds PLAN bodies into tool_docs collection | documentation-system.md |
| DEV-008 | RAG actions and knowledge_base shared store | partial | rag_lookup/rag_index over shared knowledge_base table; registered but not fully workflow-tested. | development-guide.md |
| FIX-041 | F1.2 deferred work items | aspirational | ref/base as input, fix_pr artifact, diff strategy all deferred | fix-loop.md |
| ADO-005 | Adoption variants A-D and the unwired selector | partial | reference/structure/clone/analysis modes defined but selector never wired | adoption-pipeline.md |
| CTS-039 | Component render modes (template/agent/composite/standalone) | partial | render_mode field designed with 4 modes; live data shows only template/standalone used | contracts-and-standards.md |
| SYS-059 | MessageType semantics | deployed | request = actively working now; response = reporting back, not a history marker | system-architecture.md |
| NEWS-016 | Two rerender trigger paths (site-wide batch vs single-page orchestration) | deployed | rerender-pages creates work items; page-rerender is a direct no-work-item orchestration | news-feed-pipeline.md |
| CTS-004 | Workflow result contract + dead-key stub bug class | deployed | result_from/output_fields contract; historical bug shipped entire collected_data, now fixed | contracts-and-standards.md |
| SYS-041 | Autonomous control loop | aspirational | route-produce-verify-gate-apply-feedback wraps the existing orchestrator unchanged | system-architecture.md |
| CGV-011 | Content-regression guard on section save | deployed | save_page_sections blocks thinner regenerations from overwriting richer live content | content-governance.md |
| DBI-002 | Migrations ledger system | deployed | schema_migrations table + guarded run-migrations.sh runner, live since 2026-07-10 | database-and-infrastructure.md |
| DBG-037 | Two failure envelopes: COMPLETED parent ≠ child succeeded | deployed | sendWorkflowResponse hides failure in body; notifyParentOfFailure is the other shape | debugging.md |
| STY-047 | http2 deprecation fix at the nginx conf generator | deployed | setup.sh now emits version-neutral listen directives | styling-render-pipeline.md |
| ADM-009 | React admin dashboard for build review | deployed | site-admin-dashboard.jsx: Dashboard/Review Queue/Review Detail views on mock data | admin-dashboard-and-api.md |
| DES-006 | site-design-planner scope: "Choice B" (composition-only) and its declared spec aspects | deployed | site-design-planner was scoped to write exactly one spec aspect, `resolved_composition`... | design-composition.md |
| STG-007 | JSON store scaling evolution (whole-file → daily JSONL) | deployed | site-engine's store evolved from write-cliff whole-file to bounded daily JSONL | storage-architecture.md |
| CTS-023 | Image fields optional-with-gate contract | deployed | site_assets.* fields must be required:false + skip_field + template-gated | contracts-and-standards.md |
| NAV-002 | Two nav systems and the GetNavItems fallback | deployed | site_nav tables vs legacy pages.in_header flags; partial population mixes both | navigation.md |
| PUB-001 | Public API plan (duplicate — see ADM-007 + ADM-008) | aspirational | site_ownership junction + endpoints for sites/pages/work-items/specs/assets; unbuilt; same plan as ADM-007+ADM-008 | public-api.md |
| PLAN-005 | Strategic vs plan-time guidance split + directive cascade + brief renderer + HITL lock transfer | partial | site_plan_directives cascade site->page->section; locks transfer across plan rebuilds | site-plan-and-reconciler.md |
| HITL-001 | Work item approval_mode (auto / hitl / eval) | partial | site_work_items.approval_mode column; auto/hitl live, eval defined but unused | hitl.md |
| WDS-003 | pipeline column as soft routing namespace/label | partial | site_work_items.pipeline is a mostly-unused routing label distinct from handler_agent | work-dispatch.md |
| CGV-006 | Two sources of truth for site contact email | partial | sites.email vs site_specs.identity.email can drift; COALESCE band-aid, no consolidation | content-governance.md |
| SYS-007 | Maintenance profile per-site configuration | deployed | sites.settings.maintenance_profile controls per-domain cadence, budgets, audit config | system-architecture.md |
| LNK-006 | Step 1 / Layer 1a hero+CTA schema/template hardening | deployed | skip_field + gated buttons so an unresolved CTA renders nothing | link-management.md |
| ADP-014 | Thunder Compute API specifics (field/casing/template traps) | deployed | snake_case create vs camelCase status, real template names, ubuntu login user | adapters.md |
| MDL-006 | Model swap/snapshot/revert control plane (migration 083) | deployed | snapshot/swap/revert functions; agent_definitions is the routing control plane | model-infrastructure.md |
| DEV-029 | Prompt/workflow-jsonb migration convention (snapshot-first, anchored, idempotent) | deployed | snapshot_agent() first, anchor-drift-checked edit, idempotency marker, live-row-only filter. | development-guide.md |
| DBG-066 | Snapshot-shadowing defect (version+1000 outranks active row) | superseded | snapshot_agent() rows sorted ahead of active in naive ORDER BY version loaders | debugging.md |
| CGV-009 | Snapshot-before-change backup conventions | deployed | snapshot_agent, manual component_versions inserts, CTAS bak tables before every mutation | content-governance.md |
| DBI-001 | Snapshot-before-mutate discipline | deployed | snapshot_agent/take_site_snapshot + backup naming convention before any mutation | database-and-infrastructure.md |
| CTS-056 | Static-source schema fields force fleet-generic labels/suffixes | partial | source:static label fields re-apply generic fallback text on every render; fix is static→llm | contracts-and-standards.md |
| ASG-001 | Agent spawning (agents as DB records claimed by generic pods) | deployed | spawn_agent creates agent_instances row; generic chassis pod loads config | agent-spawning-and-groups.md |
| DEV-005 | Wrapper-orchestrator pattern (every pod-running agent needs a parent) | deployed | spawn→call→complete thin wrapper gives substantive work its own dedicated K8s Job pod. | development-guide.md |
| SYS-027 | Dispatch-loop input_mapping path mismatch | unknown | spec JSONB mapped nested but handlers read flat, causing path-resolution errors | system-architecture.md |
| FTW-029 | setsid detached launch + false-success gap | deployed | ssh_exec returns immediately; exit_code 0 doesn't prove the training started | finetuning-flywheel.md |
| DBG-035 | chunkContent() infinite loop — the OOM root cause | deployed | start=end-overlap stepped backwards forever; fixed with forward-progress guard | debugging.md |
| SYS-070 | site-engine (API-only capture backend) | deployed | stdlib-only Go binary capturing intent events server-side for VM-hosted sites | system-architecture.md |
| TL-026 | Component regeneration in place (store_generated_component mechanics) + a naming-collision incident | deployed | store_generated_component looks up an existing component by the LLM's EMITTED function (forked_from IS NULL); if... | tool-lifecycle.md |
| TL-028 | Store-path template validation (+ pending `<script>`-balance hardening) | partial | store_generated_component's pre-store validation gate rejects Mode-A/B artifacts and unclosed `<style>` but NOT... | tool-lifecycle.md |
| MCL-003 | Cluster-filter gap in remote-job-spawner (Gap A) | partial | target_cluster filter exists but logs at Debug not Info on the skip path | multicluster.md |
| LNK-015 | link_registry — records but never validates (dormant substrate, abandoned) | abandoned | target_page_id never populated; live audit reads rendered_html directly instead | link-management.md |
| DBG-024 | agent_definitions source-of-truth is clients_db, not templates_db | deployed | templates_db has only the legacy 8-agent old-schema catalogue | debugging.md |
| ONB-017 | Active-config schema (four tables, computed-on-read effective values) | aspirational | tenant_configs/mechanical_config/standards/objectives contract specification | onboarding-config.md |
| DBG-033 | Prompt-template rendering resolvers differ by output_format | deployed | text→bare string, json→map, action-config→different resolver keeping .result | debugging.md |
| ADP-012 | Thunder adapter schema and provisioning gates | deployed | thunder_instances/thunder_config/thunder_provision_check enforce cost + concurrency caps | adapters.md |
| MDL-013 | Thunder spend gating (DB-side check) | deployed | thunder_provision_check view enforces a rolling daily spend cap before create | model-infrastructure.md |
| HITL-015 | HITL approval timeouts (config mapping, defaults, restart recovery) | partial | timeout_seconds never mapped to Step.Timeout; fix plan incl. goroutine recovery | hitl.md |
| TP-002 | Tool creation never enqueues the final page deploy (planned-pages gap) | partial | tool-generator creates component + page + nav but leaves the page build_status='planned'; nothing enqueues the... | tool-pipeline.md |
| INVD-001 | Abandoned "no owner" claim (checked and found false) | abandoned | tool-recreation-handler already owned the responsibility claimed missing | investigation-discipline.md |
| TL-018 | Recreation writes page sections — component-less tools and their visibility gap | deployed | tool-recreation-handler ends save_page_sections → update_status → deploy_page and never creates a... | tool-lifecycle.md |
| DOC-065 | HANDOFF permanent-thread scope split (Threads A–D) | partial | traffic-probe work split into four labeled threads | documentation-system.md |
| FTW-005 | Training-data export as chassis agent + action | deployed | training_data_export action/worker/orchestrator, v1→v3.2 evolution | finetuning-flywheel.md |
| FTW-024 | model_lifecycle schema | deployed | training_runs/artefacts/evaluations/deployable_adapters lifecycle namespace | finetuning-flywheel.md |
| CLC-008 | F7 — unguarded template swap in update_component_html (residual) | partial | update_component_html swaps a shared component's template (snapshotting versions — its old silent snapshot... | component-lifecycle.md |
| DEV-056 | Batched multi-page generation and chunked HTML generation | superseded | v1-era anti-token-limit strategies made unnecessary by the component architecture. | development-guide.md |
| HITL-006 | intake-orchestrator (classify → brief → HITL confirm → spawn builder → rerender) | partial | v1/v2 entry pipeline with two HITL gates, superseded by domain-submitter | hitl.md |
| DBG-005 | Claimed-item-timeout evidence-based auto-completion (false-positive family) | partial | v1→v2 fix history for false-completing stuck claims; homepage zero-component incident | debugging.md |
| CH-004 | Companies House matching cascade (revised 7-tier signal architecture) | partial | v2 plan: 7 tiers incl. website-scrape and corporate-group mapping, targets 70-90% match | companies-house-enrichment.md |
| DES-052 | `analyze_design` requires structured palette.reference_values (else the LLM invents a palette) | deployed | webdesign-agent's `analyze_design` LLM step reads colours only from `design_intent.palette.reference_values`,... | design-composition.md |
| IMG-013 | flattenImageryBlock write path + lock transfer (Phase 2G.2) | deployed | write_site_plan inserts site_plan_imagery rows and transfers HITL locks across replans. | imagery.md |
| DEV-037 | Whole-blob input_data passthrough mapping (anti-pattern) | superseded | {"input_data":"input_data"} double-nests the caller's data; replaced by explicit per-field mapping. | development-guide.md |
| DEV-002 | Canonical field-path resolution helpers (datahelpers) vs duplicated resolvers | partial | ~18 near-duplicate dot-path resolvers exist; datahelpers.ExtractNestedField etc. are canonical. | development-guide.md |
| STY-008 | SectionStyles: built-but-disconnected per-section CSS mechanism, retired | abandoned | ~80% built renderer mechanism no active layout ever consumes | styling-render-pipeline.md |
| BIZ-008 | Unit economics, pricing, and sourcing decisions (idea.uk) | deployed | £29 flat, cost-plus; self-hosted LLMs deferred to 2027 | business-strategy.md |
| SOC-012 | Spark revenue model | aspirational | £3-5/mo subscription + meritocratic brand sponsorship + revenue share, no pay-to-win | social-media.md |
| VIZ-001 | evidence-chart: magnitudes resolved through fact ids | deployed | CSS bars, no SVG; a chart point cannot carry its own number — every value resolves via fact_id and renders its verified date | visualisation-and-charts.md |
| VIZ-002 | evidence-timeseries: one measurement over time | live, exercised | Companion to evidence-chart; one column per dated observation, each point rendering its OWN citation. First live use 2026-07-29: Thames leakage 2020-25 (migs 265/266), claimscan + render audit clean | visualisation-and-charts.md |
| VIZ-003 | Series facts: the substrate a time axis needs | deployed | Observation{as_of,value,source}; as_of is the date the value APPLIES to, distinct from the three provenance dates | visualisation-and-charts.md |
| VIZ-004 | The honesty gate had to learn about series | deployed | numberSupported skips Value==nil, so without a series branch every plotted point reads as an unregistered number | visualisation-and-charts.md |
| VIZ-005 | Generated images explain, code-rendered output states | designed | Diffusion imagery is wrong for anything exact, selectable or translatable. Stated in features_open/023; nothing enforces it | visualisation-and-charts.md |
| VIZ-006 | mechanism-flow: drawing a process, with no numeric field | deployed | Numbered flow with decision branches. Has NO number slot by design — the absence is the control | visualisation-and-charts.md |
| VIZ-007 | No arithmetic in the render funcmap; a missing func is a PARSE error | deployed | No inc/add. Rules out SVG coordinate maths in templates, which is why charts pass values to CSS custom properties | visualisation-and-charts.md |
| VIZ-008 | $facts is declared by the template, not supplied by the engine | deployed | {{$facts := .facts}}. An undeclared variable is a parse error, so the component renders nothing rather than degrading | visualisation-and-charts.md |
| VIZ-009 | Text inside <svg> is invisible to the claims gate | deployed | extractAssertions never reaches SVG text, so an SVG diagram could assert anything and scan clean | visualisation-and-charts.md |
| VIZ-010 | scripts/render_audit.py: the post-deploy render witness | built, unwired; capability superseded by VIZ-012 | Renders every element in headless Chromium; the only thing that catches 026 family 3 (a component hard-coding ink over a themed fill), which check_palette_contrast states it cannot see | visualisation-and-charts.md |
| VIZ-011 | Chart furniture is a graphical object, so 3.0 applies | deployed | Axis lines and connectors need 3.0. --color-border scores 1.66 on oufe and fails; the accent scores 6.86 | visualisation-and-charts.md |
| VIZ-012 | render-audit-agent: the render audit as a dispatchable orchestration | live, exercised | Chassis action → dedicated pod → Chromium over every deployed page; firm vs over_image separated. First full-site run found a real 2.61 (fixed same morning). Audits deployed ROWS only; seed key is start_step, never initial_step | visualisation-and-charts.md |
| CLM-013 | Series facts: many dated observations, each independently sourced | deployed | Every observation carries its OWN source, never inherited; a rule enforced only in a validator is not enforced | claims-verification.md |
| LNK-023 | repairOutboundPageLinks: shared rerender-path link repair | deployed | The build gate's dead-link repair applied where rerendered HTML leaves for deploy, both paths, origin-stamped log | link-management.md |
| CTXA-024 | GitHubSource.CommitInfo: commit identity + committer date | deployed | Resolves short sha to full + committer date so index freshness keys on the commit, never the row clock | context-assembly.md |
| CLM-015 | The fleet-wide banned-claim set: ten patterns no site may assert about itself | LIVE v1.0.1196, council-APPROVED | Nil-safe, so an unarmed site is protected; NOT unioned at parse time because EvidenceBase is marshalled back to site_specs. UPDATED 07-29: the negation-prone pattern is now ARMED behind CLM-017's guard, and this entry's '0 findings' headline was an ARTEFACT of its absence — armed, the set finds 2 live overclaims (bugs_open/147) | claims-verification.md |
| CLM-014 | cmd/claimscan: run the live gate's own engine over exported page HTML, offline | deployed | The only way to test a candidate pattern set against copy other than the site it was written for; a session nearly rebuilt it. Prints BANNED/NUMBER, never the string "banned_claim" | claims-verification.md |
| CLM-016 | ClaimSurface: the page's structural type gates the prose number heuristic | committed, INERT until roll | 124 live findings -> 63, suppressing 61 measured false positives and nothing else; ONLY the heuristic is gated (banned claims and stat fields still scan every page type); zero value = UNKNOWN = scanned | claims-verification.md |
| CLM-017 | The negation guard: a banned PHRASE is not a banned CLAIM | committed, INERT until roll | Clause-local (stops at the first comma) so a negation in another clause cannot launder an overclaim; applies to per-site registers too — one matcher, not two that drift; 'without' and bare 'no' deliberately EXCLUDED as intensifiers. Two landmines: a pattern-set test passes VACUOUSLY when the pattern is absent, and narrowing a pattern by reasoning made it match nothing across 919 live components | claims-verification.md |
| CLM-018 | The CLAIMS FLOOR: claims checking at the persistence seam, not in workflow config | committed 07-30, INERT until roll | 6 agents persist page sections and only 2 gate — so the check moves to save_page_sections, where a workflow author cannot forget it. Blocker (banned claim) REFUSES the save, error (unregistered number) records and allows: severity-driven, not check-by-check. CORRECTS bugs_open/149 C1 — page-content-writer, the agent it named, PERSISTS NOTHING and is called by 4 of the 6. Blast radius measured first: 3 of 949 live components (0.32%) can no longer re-render, all 3 asserting something untrue | claims-verification.md |
| LNK-024 | repairSectionsBeforePersist: dead-link repair at the PERSISTENCE point | deployed | The gate repairs clean_html, which the structured save path never reads — so repair moves to where sections are written; 4 of 6 persistence paths had none by any route | link-management.md |
| NAV-013 | The nav-membership contract: one declaration, one writer, and a rebuild REQUEST | LIVE v1.0.1215, council-APPROVED | pages.in_header/in_footer DECLARES membership; URL shape decides WHERE, never WHETHER. Collapses two overlapping never-primary rules — the URL-keyed one ran BEFORE the flags were read, so a nav_drift item completed having placed nothing (proven twice, with a positive control where the same handler succeeded). site_nav_items now has ONE writer: addToolToNav is DELETED (NAV-008 superseded in part). RequestNavRebuild asks nav-updater for the rebuild instead — a nav row is NOT a link, and writing one would have silenced check_orphan_pages. LANDMINE: a third caller outside platform/orchestration/actions is an RFC moment; recurrenceExpected is load-bearing or the THIRD request per site is born terminal | navigation.md |
