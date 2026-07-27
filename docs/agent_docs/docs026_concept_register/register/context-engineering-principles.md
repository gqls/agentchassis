# Register — context-engineering-principles

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

6 concepts, consolidated from 8 raw extractions across units U14, U15, U16, U24f
(the 2 raw blocks natively tagged NEW:context-engineering-principles, plus doctrine
-level entries reassigned here from the diagnosis-loop bucket as the closer fit —
per the consolidation instructions' license to place material in the best-fit
category among those assigned to this cluster).

### CTXE-001 — Context substrate principles (authored vs derived; salience over presence)
- **status:** aspirational
- **status-evidence:** "Documentation is code... Authored vs derived... References, not copies... Salience over presence" (NOTES_running_synthesis_principles(59) §Context substrate).
- **what:** A cluster of framing principles for how an autonomous system should hold and pass context: distinguish AUTHORED sources (owner + lifecycle, can be wrong) from DERIVED sources (auto-generated, true-by-being-actual, can't be wrong, only stale); authored layers should point at derived artifacts rather than paraphrase them, so they don't drift; models lose the bigger picture from LOCAL-DETAIL SALIENCE during reasoning, not from context-window overflow, so the lever is salience management, not window size; a bigger context window is explicitly named as "not the fix for too much context" (context rot). These principles underpin nearly every other concrete mechanism in this cluster — the bundle shape contract's provenance/pointer design, the diagnosis loop's small-bundle-plus-evidence-trail shape, and the bundle-size doctrine (CTXE-005) are all applications of it.
- **sources:** docs019/NOTES_running_synthesis_principles(59).md §Context substrate; §Building discipline "A bigger context window is not the fix."
- **relations:** diagnosis loop (embodies several of these — small bundles, references not pasted copies, register/diagnosis-loop.md); B4a retrieval ceiling (CTXE-002); bundle size doctrine (CTXE-005)
- **verify-later:** n/a — doctrine, no code artifact

### CTXE-002 — B4a: the symptom-vs-mechanism retrieval ceiling
- **status:** deployed
- **status-evidence:** thin_slice(27) "OUTCOME (2026-06-17, 2 ground-truth tasks): skinner-box lexical 0.50, semantic 0.00; resultspec lexical 0.00, semantic 0.00, fused 0.00... DECISION: embeddings do NOT earn a place in the code path on this evidence"; NOTES v4(39) "The code-retrieval channel contributes nothing (measured: flat similarity band 0.547–0.574 across all 12 seed hits; zero code citations in four full runs)."
- **what:** A rigorous, repeatedly-corrected evaluation of lexical vs. semantic (Ollama/nomic) vs. fused (RRF) code-symbol retrieval against real bugs. Finding: when a bug's cause lives in shared infrastructure named for its FUNCTION rather than its FAILURE MODE, symptom-based retrieval — lexical AND semantic alike — has a category-level ceiling: symptom words and mechanism words simply don't intersect, and no embedding closes a zero-overlap vocabulary gap; it is not a ranking problem an embedding can fix. Secondary finding: naive RRF fusion can be WORSE than lexical alone, by demoting a lone correct hit. Getting to this result took five corrected measurement setups (wrong task string, contaminated index, duplicate-symbol pollution, stale shell variable, task-wording vocabulary leakage) — "the instrument, not the system, was the fault" every time (see instrument-skepticism doctrine, CTXE-004). This finding is the empirical justification for the diagnosis loop's whole re-scope-by-following-evidence design, and a live-run follow-up later confirmed the code-retrieval channel contributed essentially nothing across real diagnoses while runtime/DB evidence carried every successful one.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#B4a; docs019/RUNBOOK_design_diagnosis_loop(7).md#the-empirical-finding,#1a; docs019/DESIGN_diagnosis_loop(3).md#1a; docs019/NOTES_running_synthesis_v2(36).md 2026-06-14/06-17; docs019/NOTES_running_synthesis_v4(39).md STATE OF THE WORLD
- **relations:** call-graph re-scope mechanism (register/diagnosis-loop.md DIAG-007); evidence-follows re-scoping (register/diagnosis-loop.md); text-vs-code embeddings split (register/context-assembly.md CTXA-016); ground-truth eval harness (register/contextkit-toolchain.md)
- **verify-later:** `groundtruth_targets.json`, `eval_targets` results in the repo; whether ground truth was ever widened beyond ~2 tasks

### CTXE-003 — Corpus enrichment policy: measure first, mechanical before authored
- **status:** aspirational
- **status-evidence:** code_retrieval_route(21) "Should every function carry a human description for embedding-match? NO... Order of investment, gated on the §7E measurement" (question raised 2026-07-02).
- **what:** A position on how to improve a retrieval corpus once a ceiling is measured (CTXE-002): (1) mechanical, rot-free enrichment FIRST — extend `composeSymbolContent` with a function's string literals, since diagnosis queries quote log lines and the literals ARE the log lines; (2) Go-convention one-sentence doc comments only on the exported surface + action entrypoints; (3) explicitly NO separate tag system — the doc's first line is the tag surface. Rationale: stale docs make retrieval confidently wrong, which is the worst failure mode for a cite-or-abstain loop — an enrichment that can rot is worse than no enrichment.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#corpus-enrichment
- **relations:** static-tier corpus gaps (register/diagnosis-loop.md DIAG-033); B4a retrieval ceiling (CTXE-002); F3 learning layer (doc enrichment feedback)
- **verify-later:** composeSymbolContent; exported_no_doc census query

### CTXE-004 — Instrument-skepticism doctrine
- **status:** convention
- **status-evidence:** design_diagnosis_loop(7) "almost every wrong measurement came from the test instrument, not the system under test — a wrong task string, a contaminated index, a stale shell variable, a task description that leaked the answer's vocabulary."
- **stage2-verified (2026-07-14):** deployed → convention — Explicitly a standing doctrine ('n/a — doctrine' in verify-later). No named file/service/table to check; 'deployed' status label is a misnomer for a methodology principle.
- **what:** A standing caution carried into every measurement and every diagnosis: apply cite-or-abstain-style suspicion to one's OWN inputs (the bundle, the query, the ground truth) before suspecting the target system. Surfaced repeatedly during the B4a evaluation (CTXE-002) and encoded directly into the eval harness's guards (task-string binding, single matched index, vocabulary-leak checks); named as the thing to watch when evaluating any model verdict, not just a retrieval score.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#a-standing-caution; docs019/RUNBOOK_thin_slice(27).md#B4a-task-1
- **relations:** ground-truth eval harness (register/contextkit-toolchain.md); B4a ceiling (CTXE-002); standing evidence rules (0-rows not decisive)
- **verify-later:** n/a — doctrine

### CTXE-005 — Bundle size doctrine: "a large bundle is a smell, not a goal"
- **status:** convention
- **status-evidence:** thin_slice(27) "Context-window facts (verified against the Claude docs, June 2026)"; "aim to keep a working bundle under ~200K tokens (~800 KB)."
- **stage2-verified (2026-07-14):** deployed → convention — Bundle-size working rule/doctrine, verify-later says 'n/a (doctrine)'; no concrete artifact named (bundle size limit is a heuristic, not code). Not a false-positive present-tense-plan case, just miscategorized as 'deployed' when it's a convention.
- **what:** A working rule for feeding assembled bundles to models: keep a working bundle under roughly 200K tokens; a full 1M-token window is not used evenly (context rot), so the fix for an oversized bundle is narrower selection, not a bigger window — a direct application of the salience-over-presence principle (CTXE-001). Includes the three practical feeding routes considered (chat paste, claude.ai Project, API with prompt caching of the stable prefix).
- **sources:** docs019/RUNBOOK_thin_slice(27).md#large-bundles
- **relations:** context substrate principles (CTXE-001); call-graph neighbourhood as the narrowing instrument (register/context-assembly.md CTXA-010); responses-are-summaries doctrine (Kafka side)
- **verify-later:** n/a (doctrine); bundle sizes in diagnosis_artifacts once built

### CTXE-006 — Reuse-check retrieval pipeline design (catalog → lexical/structural → embeddings → rerank)
- **status:** aspirational
- **status-evidence:** Directly implemented by the contextkit toolchain's `resolve_targets`/`embed`/`fuse` commands (register/contextkit-toolchain.md); the design principles themselves are framed as reusable lenses, not fully built as a standing reuse-check feature.
- **stage2-verified (2026-07-14):** partial → aspirational — resolve_targets/embed/fuse exist only as a standalone Go module (module contextkit, go.mod) under docs/agent_docs/docs019.../go_files/contextkit/cmd/{embed,fuse,resolve_targets,dedup}/main.go — a prototype never merged into platform/ build. grep -rln contextkit platform/ hits only draft actions (code_symbols_actions...
- **what:** A layered design for "has this already been solved" reuse-checking that treats it as a RETRIEVAL problem with a judgement tail, not a generation problem: a maintained capability catalog is the cheapest check (lookup, not search); "identical" (token/AST fingerprinting, algorithmic, high-precision) and "similar" (semantic, embeddings) are split into different mechanisms because lexical/structural matching misses genuine near-duplicates with different names; every narrowing layer is tuned for RECALL OVER PRECISION, since a false-negative reuse check manufactures confident duplication that is worse than no check at all; a cheap model narrows the candidate set, a strong model decides on the shortlist — never the reverse; near-duplicate detection runs POST-generation against a concrete draft (a real artifact to fingerprint), while fuzzy "what's there to build on" retrieval runs PRE-generation. Signature+docstring embeddings are framed as a general retrieval substrate (also serving target resolution and capability-catalog curation), not a narrow dedup optimisation — and at the scale of a few thousand symbols, need no vector database, just in-memory cosine.
- **sources:** docs019/NOTES_running_synthesis_principles(20).md §Reuse-checking; docs archive/_archive/.../GUIDE and go files (embed.go, fuse.go headers)
- **relations:** contextkit toolchain (register/contextkit-toolchain.md); reuse search before generation (register/context-assembly.md CTXA-020); change-layer integration contract (reuse_index_refresh trigger)
- **verify-later:** whether any capability catalog or reuse index has been built beyond the contextkit prototype
