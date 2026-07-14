| DOC-001 | Documentation consolidation system (numbered canonical docs + index) | deployed | docs024 canonical set, consolidation notes, version families closed by full diffs | documentation-system.md |
| DOC-002 | Anthropic product-knowledge skill (verify, don't recall) | deployed | Skill instructs consulting official Anthropic docs over memory for Claude facts | documentation-system.md |
| DOC-003 | Per-tool travelling documentation convention (PLAN_/NOTES_ files + taxonomy) | partial | 037 convention: PLAN_/NOTES_ files per tool, instantiated across 3 project trees | documentation-system.md |
| DOC-004 | Running-notes checkpoint journal + distilled HANDOFF discipline (idea.uk) | deployed | Memory-off journal + HANDOFF pattern, run on main + sub-thread | documentation-system.md |
| DOC-005 | Docubundle context packagers + curated attach-lists (idea.uk) | deployed | Bash packagers assembling go-live and chassis-engine context bundles | documentation-system.md |
| DOC-006 | Interactive HTML runbook checklist | deployed | Tickable HTML companion mirroring markdown runbook steps | documentation-system.md |
| DOC-007 | Packaged canonical-doc copies as debug context (003 contracts copy) | deployed | Packaging workflow drops canonical docs + code dump alongside notes | documentation-system.md |
| DOC-008 | Epistemic tagging and handoff-correction discipline | deployed | Claims tagged verified/assumed/gap; correction log vs stale handoffs | documentation-system.md |
| DOC-009 | Cold-start documentation bundle practice (BUNDLE/HANDOFF/PLAN/RUNBOOK + cmd/bundle) | deployed | Four-doc travelling set per investigation via cmd/bundle | documentation-system.md |
| DOC-010 | Travelling documentation (PLAN + NOTES) in Postgres | deployed | doc_plans/doc_notes tables; umbrella system, adopted by other workstreams | documentation-system.md |
| DOC-011 | doc_plans supersede versioning (one current row, never edit history) | deployed | Supersede tx + partial unique index enforce one current PLAN row | documentation-system.md |
| DOC-012 | doc_notes append-only log with jsonb category roll-up | deployed | One row per NOTES entry; GIN-indexed categories jsonb | documentation-system.md |
| DOC-013 | DB-as-truth storage decision (knowledge_base = derived index; git = mirror) | deployed | Postgres is truth; KB is derived RAG index; git optional mirror | documentation-system.md |
| DOC-014 | Abandoned: flat-file docs-repo as truth + docselect catalogue retrieval | superseded | Rev-1 design reversed to DB-as-truth within a day | documentation-system.md |
| DOC-015 | Doc subject convention — ('tool', function) and ('pipeline', ...) | deployed | Docs keyed by subject_type/subject_key; tool vs pipeline conventions | documentation-system.md |
| DOC-016 | The dangling-doc prevention rule | deployed | NOTES subject must reference an artifact the agent actually owns | documentation-system.md |
| DOC-017 | The four doc actions (write_doc_plan, append_doc_note, load_doc_context, persist_diagnosis_note) | deployed | Chassis write/read surface for travelling docs, in production | documentation-system.md |
| DOC-018 | PLAN-at-birth write hook in tool-generator | deployed | compose_plan → write_plan → index_plan after save_tool succeeds | documentation-system.md |
| DOC-019 | NOTES-at-every-fix hook on the three fix agents | deployed | compose_note → append_note wired on fixer/improver/recreation success paths | documentation-system.md |
| DOC-020 | "Docs never fail the work" containment principle — and its limit | deployed | error_step containment covers errors, not crashes/stalls | documentation-system.md |
| DOC-021 | Pipeline documentation model — derive the topology, author the intent | partial | Step map generated from agent_definitions; PLAN bodies still pending | documentation-system.md |
| DOC-022 | Workflow-altering migrations write pipeline NOTES | deployed | Every workflow-altering migration appends a pipeline/build doc_notes entry | documentation-system.md |
| DOC-023 | NOTES category taxonomy | deployed | GIN-queryable tag vocabulary extending 037's taxonomy | documentation-system.md |
| DOC-024 | Deliberate-decisions sections + the graduation rule (prose → structured → enforced) | deployed | PLAN carries do-not-re-fix prose; enforcement deferred until recurrence | documentation-system.md |
| DOC-025 | Framing: plan = enforced desired state; pipeline = compiled runbook; NOTES = reasoning log | deployed | Agreed 2026-07-04 framing across site_plans/pipeline/NOTES/contracts | documentation-system.md |
| DOC-026 | load_doc_context fix-time retrieval | deployed | Composes current PLAN + latest NOTES + criteria_json into one block | documentation-system.md |
| DOC-027 | tool_docs knowledge-base indexing of PLANs (rag_index derived index) | deployed | rag_index chunks/embeds PLAN bodies into tool_docs collection | documentation-system.md |
| DOC-028 | EDIT-marker / -EDIT check-id convention | deployed | Fill-later blanks in seeded docs; -EDIT checks skipped until real selectors land | documentation-system.md |
| DOC-029 | Pilot PLAN seeding by SQL (dogfooding the format) | deployed | First tool PLAN hand-seeded before workflow wiring existed | documentation-system.md |
| DOC-030 | Provenance stamps the chassis, not the logical agent | deployed | Config-declared source fields are the reliable provenance, not agent headers | documentation-system.md |
| DOC-031 | Handoff-document discipline (updated-every-turn, supersede chain, turn log) — travelling_docs thread | deployed | Three-generation HANDOFF chain with newest-first turn log | documentation-system.md |
| DOC-032 | Standing opens ledger of the travelling-docs arc | partial | Carried-forward small items repeated across every revision | documentation-system.md |
| DOC-033 | Context-bundle seeding for fresh agent threads (imagery) | deployed | Script assembles imagery workstream's cold-start context bundle | documentation-system.md |
| DOC-034 | Traffic-probe context packaging (docubundle) | deployed | Packager bundles task brief, domain list, deploy docs for cold start | documentation-system.md |
| DOC-035 | Single-source relocation with pointer / canonical-doc-home discipline | deployed | Duplicated topics consolidated into one numbered doc + pointer sentence | documentation-system.md |
| DOC-036 | Full heading+content-line diff across all forked copies before consolidating | deployed | Diff-before-promote methodology for travelling/forked docs | documentation-system.md |
| DOC-037 | verify_before_migration pre-flight convention | deployed | Companion pre-flight SQL script before any hand-applied migration | documentation-system.md |
| DOC-038 | doc_notes / travelling-docs integration boundary | deployed | Fix-loop persists terminal notes via another workstream's persist_note gate | documentation-system.md |
| DOC-039 | Guideline-compliance review methodology (001/002/003 walkthrough before shipping) | deployed | Point-by-point guideline walkthrough producing a test plan | documentation-system.md |
| DOC-040 | Doc claim-verification / dated-claim convention | deployed | [checked YYYY-MM-DD] tags on falsifiable claims; whole-doc stamps banned | documentation-system.md |
| DOC-041 | Doc-drift claim classifier (design only, named across three units, never built) | aspirational | Evidence-or-abstain claim classifier, tiered T1-T3, consistently deferred | documentation-system.md |
| DOC-042 | Claim taxonomy: code-checkable / superseded-but-not-wrong / code-invisible | aspirational | Three buckets of doc claims by checkability | documentation-system.md |
| DOC-043 | Classify, do NOT merge (the human consolidates) | deployed | Standing rule: LLM finds/cites, human decides/writes canonical docs | documentation-system.md |
| DOC-044 | Date/version as triage, not truth | deployed | Dates order the verification queue; never override a code check | documentation-system.md |
| DOC-045 | Standing conformance suite (carved out, deliberately not built) | aspirational | Continuous behave-as-documented monitor, deliberately scoped out | documentation-system.md |
| DOC-046 | Documentation archiving subproject (docs019 cleanup) | partial | dedup+thin_versions+staged migration plan; manifests prove partial execution | documentation-system.md |
| DOC-047 | dedup (cmd/dedup) | deployed | Exact/near-duplicate file finder with report-only default and undo manifest | documentation-system.md |
| DOC-048 | thin_versions (cmd/thin_versions) | deployed | Keeps newest N versions per document subject, archives the rest | documentation-system.md |
| DOC-049 | docs019 migration staging script (stage_docs019_migration.sh) | partial | Automates deterministic archive moves; editorial moves stay human-gated | documentation-system.md |
| DOC-050 | Bundle-first handoff practice (context packs; broad script vs lean assembler) | deployed | Task handoffs pair problem statement with cmd/bundle invocation | documentation-system.md |
| DOC-051 | Engines docs tree + single _archive graveyard | aspirational | Target restructure separating engine code/docs/archive | documentation-system.md |
| DOC-052 | Travelling-docs pattern (runbook = plan, notes = history, handoff = session) | deployed | General cross-project framing of the runbook/notes/handoff triad | documentation-system.md |
| DOC-053 | Three parallel threads with hard boundaries | deployed | Concurrent chat threads own non-overlapping territories | documentation-system.md |
| DOC-054 | Concept register and the council-of-concept-experts mission | aspirational | The docs026 programme itself: extract, classify, verify, seed council agents | documentation-system.md |
| DOC-055 | Four-layer documentation model for automation | aspirational | Standards + context substrate + known-good library + trust ledger | documentation-system.md |
| DOC-056 | Published reasoning as substrate + drift detection | aspirational | Decisions publish their premise, enabling drift detection | documentation-system.md |
| DOC-057 | Objective tree vs concern tree (two orthogonal axes) | aspirational | Vertical mission tree kept separate from horizontal concern tree | documentation-system.md |
| DOC-058 | Authored vs derived context (one substrate, change layer between) | aspirational | Authored (owned) vs derived (emitted) context, with a change layer between | documentation-system.md |
| DOC-059 | Debugging-guide fork-and-merge maintenance (cumulative 016b copy) | deployed | 016b guide forks across chat threads, periodically merged back | documentation-system.md |
| DOC-060 | Handoff document convention (stand-alone dated brief for a fresh chat) — vonc thread | deployed | Dated self-contained handoff with orientation, DONE state, backlog | documentation-system.md |
| DOC-061 | API documentation system (OpenAPI external + per-service internal API.md) | unknown | Two-tier API doc practice, predates vonc corpus, unverified currency | documentation-system.md |
| DOC-062 | Classic pre-docs024 documentation tree (emptied) | abandoned | Original top-level doc set now all zero-byte archived files | documentation-system.md |
| DOC-063 | 2026-05-24 launcher build handoff (superseded Option A) | superseded | First training-launcher build handoff, carries two disproven claims | documentation-system.md |
| DOC-064 | Deploy-from-context-packs guide — six deploy mechanisms (A–F) — dropped from the live tree | abandoned | Cross-cutting deploy-mechanism guide, absent from the live idea.uk tree | documentation-system.md |
| DOC-065 | HANDOFF permanent-thread scope split (Threads A–D) | partial | traffic-probe work split into four labeled threads | documentation-system.md |
| DOC-066 | docs019 working/main snapshot bundle (duplicate early-draft staging copy) | superseded | Nested archive-of-archive with zero unique content vs live docs | documentation-system.md |
| SQ-001 | Site Quality Programme — the three-way split and seven legs | partial | Programme closing deploys-vs-best-in-class gap via A/B/C split, 7 legs | site-quality.md |
| SQ-002 | Site-chrome gap hypothesis (relay path lacks chrome rendering) | unknown | Hypothesis: relay build path never renders nav/header/footer chrome | site-quality.md |
| SQAM-001 | Three-way split quality-gap diagnostic method (stuck / poor / out-of-scope) | aspirational | Triage method: dispatched-but-stuck vs delivered-but-poor vs never-in-scope | site-quality-audit-methodology.md |
| SQAM-002 | Baseline mechanical quality measurement methodology | deployed | Pre-LLM deterministic HTML metric pass to form falsifiable hypotheses | site-quality-audit-methodology.md |
