# EXTRACTION U15 — docs019 running synthesis notes (NOTES_running_* family)
Extracted 2026-07-13. Files in scope: 128. Concepts found: 51.

Five version families of running chat-synthesis notes, all append-only (each
numbered version is a strict superset of the previous one in its family —
confirmed by monotonic line-count growth and by diffing the earliest vs.
highest-numbered header set in each family, which showed zero dropped
headers). Two families (`v3`, itself archived into `v4`) are explicit
continuation threads of the same underlying project (the contextkit
diagnosis-loop build), so `principles`/`v2` (parallel forks of one ancestor)
and `v3`→`v4` (sequential continuation) together tell one long story about
building a code-diagnosis tool and pivoting it into a diagnosis→fix loop.
`fixloop` is the founding thread of that pivot's own notes file (its content
overlaps with, and is more detailed than, v4's final entries).

## Coverage
| file | treatment |
|---|---|
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_fixloop.md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_fixloop(1).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_fixloop(2).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_fixloop(3).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_fixloop(4).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_fixloop(5).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_fixloop(6).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_fixloop(7).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_fixloop(8).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_fixloop(9).md | full |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_principles(48).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_principles(49).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_principles(50).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_principles(54).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_principles(55).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_principles(56).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_principles(57).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_principles(58).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_principles(59)_last_one_archived_v1.md | full |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(0).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(1).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(2).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(3).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(4).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(5).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(6).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(7).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(8).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(9).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(10).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(11).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(12).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(13).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(14).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(15).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(16).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(17).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(18).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(19).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(20).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(21).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(22).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(23).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(24).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(25).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(27).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(28).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(29).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(30).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(31).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(32).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(33).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(34).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(35).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v2(36).md | full |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3.md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(1).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(2).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(3).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(4).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(5).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(6).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(7).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(8).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(9).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(10).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(11).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(12).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(13).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(14).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(15).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(16).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(17).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(18).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(19).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(20).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(21).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(22).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(23).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(24).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(25).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(26).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(27).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(28).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(29).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(30)-fable.md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(31).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v3(32).md | full |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4.md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(1).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(2).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(3).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(4).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(5).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(6).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(7).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(8).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(9).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(10).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(11).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(12).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(13).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(14).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(15).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(16).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(17).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(18).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(19).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(20).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(21).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(22).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(23).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(24).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(25).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(26).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(27).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(28).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(29).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(30).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(31).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(32).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(33).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(34).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(35).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(36).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(37).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(38).md | family-delta |
| docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/NOTES_running_synthesis_v4(39).md | full |

Family-delta note (applies to every row above so tagged): earliest-vs-latest
header diff was empty for every family (no section headers present in an
early version are absent from the latest), and per-family line counts are
monotonically non-decreasing version-over-version — both confirm these are
pure append-only logs with no observed drops, so the "dropped/superseded
concept" signal this treatment normally hunts for did not fire. The one
partial exception is `NOTES_running_synthesis_v2(0).md`, whose "STATE DIGEST
as of 2026-06-14" header is superseded (not deleted) by a later restated
digest in the same family — captured under status **superseded** below where
relevant.

## Concepts

### Diagnosis loop (contextkit)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** v4 STATE OF THE WORLD (2026-07-02): "Diagnosis loop (§6): DONE. §6A–§6G all passed; §6G accepted on run `51f95cda`... Engine (pkg/diagnose) + diagnose-agent workflow live."
- **what:** A read-only, human-gated agent that diagnoses chassis bugs by iterating hypothesise → gather scoped evidence → cite-or-abstain verdict → re-scope by following the evidence (call graph for code, vetted/model-written queries for data) — never re-searching the symptom, never fixing, never triggering a run. Built first as a standalone Go engine (`contextkit/internal/diagnose/`) with a tested scaffold (guards, trail, verdict parsing), then ported into the chassis as a workflow-driven agent (`diagnose-agent`/`diagnose-orchestrator`) where each iteration is an observable sequence of steps (gather → verdict via `execute_llm_prompt` → `diagnose_route` → loop-back or `diagnose_emit`).
- **sources:** NOTES_running_synthesis_v2(36).md §STATE DIGEST 2026-06-17; NOTES_running_synthesis_v3(32).md §STATE DIGEST; NOTES_running_synthesis_v4(39).md §STATE OF THE WORLD; NOTES_running_synthesis_principles(59) "diagnosis-loop design updated" entries.
- **relations:** contextkit CLI toolchain; convergence guards; verdict cite-or-abstain contract; call-graph re-scope mechanism; B4a embedding-quality finding; diagnosis→fix loop workstream (successor pivot).
- **verify-later:** `pkg/diagnose/` in chassis repo; `platform/orchestration/actions/diagnose_*.go`; `agent_definitions` rows for `diagnose-agent`/`diagnose-orchestrator`; whether the "eval gate" (reproduce the gamesdesign reversals on a live model) was ever actually run.

### Diagnosis-loop chassis integration architecture
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** v3 STATE DIGEST: "the agent rewrite landed: `diagnose-agent` now has the `diagnose_route` workflow in `default_config`"; v4: "§7 ROUTE CLOSED: run 73ed55c6 full trail read; §7E green".
- **what:** The loop is realised as a chassis AGENT (workflow of steps), not a new CLI or long-running service, following "every agent is an orchestrator": a thin `diagnose-orchestrator` spawns a `diagnose-agent` worker whose workflow (in `default_config`, not the three NULL `*_workflow` columns) is `analyse_repo → lookup_symbols → load_runtime → assemble_bundle → verdict(execute_llm_prompt) → route(diagnose_route) → [loop to assemble_bundle | emit] → complete`. The verdict step reuses the existing `execute_llm_prompt` action rather than a new action; `diagnose_route` is a router action that sets no `output_field` (its result lands under step-name `route`) and returns `next_step` per the coordinator's `getNextStepFromResult` mechanism.
- **sources:** NOTES_running_synthesis_v2(36).md (chassis integration entries, 2026-06-17); NOTES_running_synthesis_v3(32).md DECISIONS (diagnose_route seeding/state-threading fixes).
- **relations:** Diagnosis loop; Workflow default_config location convention; SagaCoordinator output_field contract.
- **verify-later:** `agent_definitions` rows for diagnose-agent/orchestrator; `coordinator.go` `getNextStepFromResult`/`ProcessResponse`.

### Verdict cite-or-abstain contract
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "the diagnosis loop is now COMPLETE as far as can be built off-chassis — design + tested scaffold + tested adapters + runnable entrypoint + the model prompt + the tested prompt↔scaffold seam" (principles(59), 2026-06-17 entry, mirrored in v2(36)).
- **what:** The model-facing prompt contract (`PROMPT_diagnosis_verdict.md`) requires every verdict to CONFIRM or REFUTE only with a verbatim-quoted citation from the bundle, else the outcome is coerced to UNVERIFIABLE; abstention is asymmetric (runtime evidence readily refutes, but confirms only on direct mechanism, never "consistent with"); the re-scope must follow what the evidence names, not re-search the symptom; and the model is told to apply the same suspicion to its own reading of the bundle that the loop applies to hypotheses. A parallel wire format (`verdict_wire.go`) parses model output as human-legible strings (CONFIRMED/REFUTED/UNVERIFIABLE) with fail-safe unknown→UNVERIFIABLE coercion.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 "VERDICT PROMPT drafted"; NOTES_running_synthesis_principles(59) DB discipline / diagnosis-loop design entries.
- **relations:** Diagnosis loop; convergence guards; data-request channel (model-written SQL).

### Convergence guards for the diagnosis loop
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "The four convergence guards (iteration-cap, scope-not-narrowing, evidence-not-growing, hypothesis-thrash) + the no-citation→UNVERIFIABLE coercion are all behaviour-tested" (v2(36) STATE DIGEST); v3(32) DECISIONS: "A new data_request counts as forward progress in the spin guards (turn 31)".
- **what:** A set of anti-spin safety mechanisms bounding the loop: an iteration cap, a rule that re-scope can't balloon past prior scope + 2, a rule that a verdict adding no new citation halts the loop, and thrash detection for hypothesis oscillation without new discriminating evidence. Later hardened so an issued (not yet cited) read-only data request also counts as forward progress, preventing the loop from stopping one iteration before a fixed query's result would have arrived.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 scaffold entries; NOTES_running_synthesis_v3(32).md DECISIONS (turns 30-31).
- **relations:** Diagnosis loop; verdict cite-or-abstain contract.

### Call-graph re-scope mechanism
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "CallGraph adapter... reads analyser `calls`, resolves callee NAMES back to defining symbols for re-scope" (v2(36) 2026-06-17).
- **what:** Re-scoping in the diagnosis loop follows the analyser's recorded (name-based, not type-resolved) call graph outward from an evidence-named site, deliberately dropping ubiquitous names (Run/String/Error/New/... plus any name resolving to more than 8 definitions) so following doesn't explode into noise — described as "the symptom-vocabulary trap in call-graph form."
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 "diagnosis-loop adapters... BUILT & tested".
- **relations:** Diagnosis loop; B4a embedding-quality finding.

### B4a embedding-quality evaluation & symptom-vs-mechanism retrieval ceiling
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** v2(36) STATE DIGEST: "embeddings do NOT earn a code-path place; the lever is the diagnosis loop... Retrieval is necessary-not-sufficient"; v4(39) STATE OF THE WORLD: "The code-retrieval channel contributes nothing (measured: flat similarity band 0.547–0.574 across all 12 seed hits; zero code citations in four full runs)."
- **what:** A rigorous two-task (later extended) evaluation of lexical vs. semantic (nomic/Ollama) vs. fused (RRF) code-symbol retrieval against real bugs, run through five corrected measurement setups (wrong task string, contaminated index, duplicate-symbol pollution, stale shell var, task-string vocabulary leakage — "the instrument, not the system, was the fault" every time). Conclusion: when a bug's cause lives in shared infrastructure named for its function, not its failure mode, symptom-based retrieval — lexical AND semantic alike — has a category-level ceiling (zero vocabulary overlap, not a ranking problem); naive RRF fusion can make results WORSE than lexical alone by demoting a lone correct hit. Later (v4) measured that the code-retrieval channel contributed essentially nothing across real runs, while runtime/DB evidence carried every successful diagnosis.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-14/06-17 B4a entries; NOTES_running_synthesis_v4(39).md STATE OF THE WORLD.
- **relations:** Diagnosis loop; call-graph re-scope mechanism; code-context retrieval infrastructure; reuse-checking retrieval architecture.
- **verify-later:** `groundtruth_targets.json`, `eval_targets` results in the repo, whether ground truth was ever widened beyond ~2 tasks.

### contextkit CLI toolchain
- **category:** NEW:contextkit-toolchain
- **status-signal:** deployed
- **status-evidence:** v2(36) STATE DIGEST: "analyser (-exclude + *(N).go skip...), resolve_targets (lexical), embed (semantic, Ollama nomic), fuse (RRF -json k=60), eval_targets..., assembler, dbcontext, dedup, thin_versions, cmd/bundle..., cmd/diagnose."
- **what:** A family of small, report-first, behaviour-tested Go CLIs built to prototype and measure context-assembly/diagnosis before chassis porting: `analyser` (Go-AST symbol index with exclude/dedup-skip), `resolve_targets` (lexical scoring), `embed`/`fuse` (semantic + RRF), `eval_targets` (recall@N/MRR against ground truth), `assembler` (composes a bundle: constitution + docs + schema + symbol bodies + runtime), `dbcontext` (read-only DB gather: `\d`, `-rows`, `-capabilities`, `-runtime-site`), `dedup`/`thin_versions` (docs-archiving tools), `cmd/bundle` (orchestration wrapper: runs dbcontext then assembler), `cmd/diagnose` (dev/test harness for the loop).
- **sources:** NOTES_running_synthesis_principles(59) multiple 2026-06-13 entries (tool-building); NOTES_running_synthesis_v2(36).md STATE DIGEST.
- **relations:** Diagnosis loop; docs archiving toolchain; code-context retrieval infrastructure.

### Code-context retrieval infrastructure (analyser adapter + code_symbols)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** "MILESTONE 2026-06-12: analyser-adapter DEPLOYED TO PRODUCTION" (principles(59)); v4(39) DECISIONS: "Fix direction: migrate code-indexer's analysis step to analyse_repo_local".
- **what:** The chassis's in-cluster code-indexing pipeline: an `analyser-adapter` (Kafka worker, tarball-fetches a repo read-only, runs the shared `internal/analysis` Go-AST walker) feeds `index_code_symbols`, which embeds symbols (nomic-embed-text via the existing `AIService`/`ollama-adapter` seam, reusing the same `rag_index`/`rag_lookup` hybrid pattern as `knowledge_base`) into a sibling `code_symbols` pgvector table (HNSW index, identity-keyed on repo/path/symbol, commit-versioned, hard-deleted not soft-deleted since it's a rebuildable cache). Later found to be indexing a year-old stale tree (fix direction: swap to `analyse_repo_local`, the in-process fetch-and-analyse path already proven in the diagnose workflow).
- **sources:** NOTES_running_synthesis_principles(59) DB discipline section (2026-06-11/12); NOTES_running_synthesis_v4(39).md 2026-07-02 "corpus check result: the index is the blocker" and DECISIONS.
- **relations:** Adapter response envelope contract; B4a embedding-quality finding; diagnosis loop.
- **verify-later:** `code_symbols` table population/freshness, `index_code_symbols` action's current data source.

### Adapter response envelope contract
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "003-vs-FOCUS contradiction RESOLVED empirically (2026-06-11) from coordinator.go + git/thunder/dynamic/websearch adapters" (principles(59)).
- **what:** The chassis's normative contract for any Kafka message replier (adapter or agent): body must be a typed struct with real bools (never `map[string]string` string-bools — this exact bug caused a real multi-day thunder production fault), sent via `ProduceWithValidation` (never bare `Produce`), with `in_response_to_request_id` as the primary matcher the coordinator claims on (`request_id` is a fallback — "reuse both" is the safest pattern), and `action`/payload read from the message BODY (not headers). Originally duplicated and drifted between doc 003 and `FOCUS_adapter_design`; single-sourced into `035_adapter_guide.md` after empirical verification against `coordinator.go` and four live adapters (websearch was the deprecated outlier).
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 envelope-contract entries; NOTES_running_synthesis_v2(36).md/v3(32).md (analyser adapter build referencing the same contract).
- **relations:** Analyser adapter build; canonical-doc-home discipline; code-context retrieval infrastructure.
- **verify-later:** `platform/orchestration/types` `ResponseHeaders`/`ResponseMessage`; `platform/validation` `ValidateOutgoingMessage` (the still-open "promote from prose to validator enforcement" TODO).

### Analyser adapter build (polyglot code-parsing service)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "MILESTONE 2026-06-12: analyser-adapter DEPLOYED TO PRODUCTION (uk_001)" (principles(59)).
- **what:** A from-scratch Kafka-worker adapter modelled structurally on thunder/git (own image, dockerfile, kustomize base+overlay, config loader, graceful `Shutdown()` with `sync.Once`, health probes) whose one genuine difference is importing the shared chassis-root `internal/analysis` package and holding a dedicated least-privilege, read-only, repo-scoped GitHub token via `secretKeyRef` (never passed through the spawning pod). Fetches via a tarball GET (no git binary, no go-git), not a clone.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11/12 adapter-build entries; NOTES_running_synthesis_v3(32).md (turns 25-27, tarball-fetcher reuse into `internal/reposource`).
- **relations:** Adapter response envelope contract; code-context retrieval infrastructure; GitHub read-token scoping pattern.

### Diagnose-agent self-contained repo fetch
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** v3(32) DECISIONS: "Symbol bodies come from a git checkout the diagnose-agent makes itself (Option 3, turns 25-26)."
- **what:** Rather than adding a body column to `code_symbols` or coupling every diagnosis iteration to a live analyser holding a checkout, the diagnose-agent fetches its own tarball (reusing the analyser's `FetchToDir`, lifted into a neutral `internal/reposource` package so both the analyser adapter and the diagnose action share one fetcher) and runs `internal/analysis` in-process for both the call graph and symbol-body slicing — one fetch, no cross-pod coupling, git stays the only source of truth for code. Fetches are pinned to the same commit the `code_symbols` index was built on (best-effort, falls back to `ref`/HEAD) so lookup-seeded symbols resolve in the fetched tree.
- **sources:** NOTES_running_synthesis_v3(32).md turns 25-27 (DECISIONS).
- **relations:** Analyser adapter build; code-context retrieval infrastructure; symbol-body slicer (ReadSymbolBody).

### Symbol-body slicer (analysis.ReadSymbolBody)
- **category:** NEW:contextkit-toolchain
- **status-signal:** deployed
- **status-evidence:** v3(32) STATE DIGEST: "`analysis.ReadSymbolBody`... TESTED... and proven BYTE-IDENTICAL to `cmd/assembler`'s real bundle output."
- **what:** The single shared implementation (in `internal/analysis`, not duplicated per-consumer) that turns a `path:Symbol` scope entry into source text by slicing the analyser's already-recorded `start_line..end_line` spans — never re-parsing — used identically by `cmd/assembler` and the chassis `diagnose_assemble_bundle` action, closing a prior stub.
- **sources:** NOTES_running_synthesis_v3(32).md STATE DIGEST + DECISIONS.
- **relations:** Diagnose-agent self-contained repo fetch; contextkit CLI toolchain.

### Model-written SQL under a three-guard read-only substrate
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** v3(32) DECISIONS: "EXPLAIN-estimate pre-flight guard on data_requests"; principles(59)/v2(36) "three-guard model-SQL... Guard 3 CONFIRMED — pgbouncer pool_mode = transaction."
- **what:** Rather than a vetted-query-catalogue-only approach (rejected as too limited for open-ended diagnosis) or an unsafe SQL-string-filter (rejected: cannot be made safe against statement stacking, data-modifying CTEs, `COPY ... TO PROGRAM`), the design lets the verdict model emit arbitrary SQL under three layered guards: (1) the prompt instructs SELECT-only; (2) a parse-lint (`sqlguard.go`, word-boundary token checks) drops anything unsafe before it can be issued, recording the drop in the evidence trail; (3) the real safety boundary — execution under a read-only DB transaction (`BeginTx(ctx,&sql.TxOptions{ReadOnly:true})`), confirmed transaction-scoped-safe under pgbouncer's `pool_mode=transaction` (session-level `SET`/`ALTER ROLE` was found NOT reliable under pgbouncer pooling). A schema digest (denylist-based `## Schema` bundle section + `EXPLAIN`-estimate size guard) is fed to the model so it writes SQL against confirmed columns.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 DB-evidence entries; NOTES_running_synthesis_principles(59) DB discipline; NOTES_running_synthesis_v3(32).md DECISIONS (turn #2 EXPLAIN guard).
- **relations:** Diagnosis loop; DB discipline / snapshot_agent convention; pgvector/rag hybrid retrieval reuse.
- **verify-later:** `diagnose_load_runtime` execution wiring (flagged as the one remaining piece not testable outside the chassis).

### DB discipline / snapshot_agent convention
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** "Snapshot before any `agent_definitions` change; check `\df` for a helper before hand-writing SQL" (principles(59)); "Backup before any agent_definitions change (standing rule, turn 11)" (v3(32) DECISIONS).
- **what:** A standing rule, reinforced repeatedly across the notes, to call `SELECT snapshot_agent('<type>','<reason>')` (with a paired `revert_agent`) before any change to `agent_definitions`, and to check `\df`/`\dx` for an existing DB helper or extension (pgvector, pgcrypto, pg_trgm all confirmed present) before hand-writing SQL — reuse-before-recreate applied at the database layer.
- **sources:** NOTES_running_synthesis_principles(59) DB discipline section; NOTES_running_synthesis_v3(32).md DECISIONS.
- **relations:** Model-written SQL guard model; schema-before-SQL discipline.

### Doc claim-verification / dated-claim convention
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** "DONE 2026-06-11: claim-verification discipline CODIFIED (016 v2_34 item 24 + 001 pointer + dated-claim convention)" (principles(59)).
- **what:** Falsifiable, load-bearing doc claims carry a `[checked YYYY-MM-DD: <evidence>]` tag (one date = last checked, updated in place); negative claims ("X isn't built") carry their falsifying command; whole-document "verified" stamps are explicitly banned (verification attaches to claims, never documents); docs update in the same change as the decision that makes them true. Motivated by a real incident where a stale negative claim in doc 019 nearly caused a reuse-before-recreate violation, and the team's own freshly-written docs went stale within hours of being written — "staleness is a coupling property, not an age property."
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 "DISCUSSED... doc up-to-dateness" and "CODIFIED" entries.
- **relations:** Doc-drift claim classifier design; canonical-doc-home discipline.

### Doc-drift claim classifier design
- **category:** documentation-system
- **status-signal:** aspirational
- **status-evidence:** "DESIGN 2026-06-13: doc-drift claim classifier... doc-drift classifier (DESIGN_doc_drift_classifier.md) is design-only, gated on testing vs known bugs" (v2(36) small-pending list).
- **what:** A designed-not-built tool that classifies individual doc claims as current/stale using tiered evidence (T1 static code/schema, T2 DB row state, T3 behavioural — reading EXISTING logs only, never triggering a run) under two hard rules: read-only (never mutate to test a sentence) and abstention asymmetry (a verdict without a citation is UNVERIFIABLE; behavioural evidence supports "stale" only on direct contradiction, since misattributing an unrelated bug/flaky run to a correct doc is worse than staying silent). Explicitly classify-don't-merge: an LLM can check a claim but must never generatively merge/rewrite docs, since a rewrite can silently drop caveats no code-check would catch.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-13 "DESIGN... doc-drift claim classifier" entry.
- **relations:** Doc claim-verification convention; docs archiving toolchain; diagnosis loop (shares its cite-or-abstain design DNA).

### Docs archiving toolchain (dedup, thin_versions, staged migration)
- **category:** documentation-system
- **status-signal:** partial
- **status-evidence:** v2(36) small-pending: "Doc archiving — tools built + tested, RUN not yet done."
- **what:** A set of report-first, behaviour-tested tools built to de-noise the docs019 directory (2,729 files → 1,917 (dir,stem) subjects; 1,734 untouchable singletons; noise concentrated in 18 "fat clusters" of 10+ versions, mostly under docs024): `dedup` (exact SHA + optional near-duplicate copies → `_archive/`), `thin_versions` (keeps newest N per subject by version>bracket>mtime rank, targets only fat clusters), and `stage_docs019_migration.sh` (deterministic archive-dir moves + dedup delegation + a human-edited `PROPOSED_MOVES.tsv` for genuinely editorial moves — canonicality of 200+ working docs cannot be inferred from filenames alone). `dedup` shipped with a real silent destructive-flag bug (Go's `flag.Parse()` stops at the first positional, so `dedup <root> -move` printed "REPORT ONLY" and moved nothing) caught only by a behaviour test asserting on post-move tree state, not by compile/vet.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-13 dedup/thin_versions/archiving entries.
- **relations:** Untested-code / behaviour-testing discipline; doc claim-verification convention; contextkit CLI toolchain.
- **verify-later:** Whether the actual cleanup run (RUNBOOK_doc_archiving.md) was ever executed against the live docs019 tree — this very U15 file enumeration still shows many `(N)`-suffixed files present.

### Canonical-doc-home / single-sourcing discipline
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** "DONE 2026-06-11: FOCUS_adapter_design merged into 035 — FOCUS retired" (principles(59)).
- **what:** A recurring lesson that contract duplication across docs (003 vs `FOCUS_adapter_design`; the "003 vs FOCUS contradiction" root-caused as duplication-then-drift, not a genuine disagreement) is fixed by promoting the contract to ONE numbered canonical doc that others link to rather than restate, plus (proposed, not built) tightening the actual validator so the contract can't silently rot behind prose. Numbered docs (001/002/003/019/020...) are canonical/permanent; `FOCUS_*` docs are transient design notes meant to be retired once their content graduates.
- **sources:** NOTES_running_synthesis_principles(59) "Doc restructure: adapter docs + 003" section and the following DONE entries.
- **relations:** Adapter response envelope contract; doc claim-verification convention.

### Tool-doc header system
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "DONE 2026-06-11: 019/020 CONSOLIDATED — splice files retired" (principles(59)); status marked apply-ready and untouched in v2(36) small-pending list.
- **what:** A standardised 6-12 line sentinel-delimited header block written into every generated tool's script (purpose, behavioural invariants, no-external-calls, version marker) at creation time, stripped at deploy-assembly (three call sites: single-page rerender, `collectJSAssets`, bulk rerender) so it never ships to visitors but is retained in the DB `html_template` for audit/parse parity. Enforced via a hard `HasToolDocHeader` gate in `create_tool_component`, tool-generator/tool-improver prompt edits, and two new `tool_health` tier-1 checks (`no_doc_header` warning, `malformed_doc_header` error). Paired with new `source_agent_type`/`source_orchestration_id` provenance columns on `content_components`, mirroring `knowledge_base`'s existing provenance pair.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 tool-doc entries (multiple DONE items).
- **relations:** JS content separation contract; doc claim-verification convention; canonical-doc-home discipline.
- **verify-later:** Whether the rollout (provenance migration → prompts SQL → binary release) was ever applied — repeatedly flagged as "apply-ready, not yet applied" across all later notes files through 2026-07-06.

### JS content separation contract
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** "js_content RESOLVED — the assets-split EXISTS (003 'JS Content Separation Contract'...)" (principles(59), 2026-06-11).
- **what:** For interactive components, `store_generated_component_action.go`'s `separateInlineJS()` extracts inline `<script>` bodies into a separate `content_components.js_content` column, replacing them in `html_template` with a `<script src="/tools/assets/{function}.js">` reference; `RerenderSinglePageAction.collectJSAssets()` assembles the resulting multi-file git commit. Verified NOT used by the library-tool pipeline (tool-generator/improver mandate one inline script; the fork INSERT omits `js_content`), creating a landmine where a library tool adopting the split would fork with a dangling script reference and a 404'ing asset.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 "js_content RESOLVED" and "VERIFIED... 019's 'isn't built yet'" entries.
- **relations:** Tool-doc header system; fork-divergence detection.

### Fork-divergence detection for library tools
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "IMMEDIATE WIN INSTEAD: FORK-DIVERGENCE detection — pure SQL discovery check (tier-1, zero cost)" (principles(59)).
- **what:** A proposed zero-cost SQL discovery check comparing a deployed fork's `html_template` hash against its `forked_from` library original to answer "which forks are unmodified / safe to bulk-push a library change" — deliberately deferred building full code-symbol indexing of tools (each tool is one IIFE, thin symbol pickings; tool discovery already solved via `semantic_tags`) until a concrete consumer needs it.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 tools/provenance/docs design entry.
- **relations:** Tool-doc header system; JS content separation contract.

### Gamesdesign silent-no-op-rebuild bug (content-regression + status-rollup)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** v2(36) STATE DIGEST: "gamesdesign silent-no-op bug — RESOLVED (the real fix is now a fixture)."
- **what:** A real production bug used repeatedly as the diagnosis loop's eval fixture, diagnosed across several sessions with TWO wrong hypotheses along the way (per-section `max_tokens:2000` cap; then recreate-mode discriminator) before the real cause was found: a January chassis regression made `SagaCoordinator.extractWorkflowResult` honour only the PLURAL `output_fields` key, while `page-content-writer` declares the SINGULAR `output_field`, so the compiled page collapsed into an oversized state-dump skip path that reported "completed" while the live page never updated. Fix: `resolveResultSpec` (new, `result_spec.go`) treats singular as FLATTEN, honouring the long-ignored mapping key. The reversals in this diagnosis are the canonical worked example baked into the diagnosis loop's verdict prompt.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-14/17 gamesdesign entries; NOTES_running_synthesis_principles(59) 2026-06-13/14 diagnosis narrative.
- **relations:** SagaCoordinator output_field contract; diagnosis loop; B4a embedding-quality finding (this bug's real fix is the "ceiling" ground-truth task).

### SagaCoordinator output_field singular/plural extraction contract
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "Fix = `resolveResultSpec` (result_spec.go) treats singular as FLATTEN, keeps plural unchanged. Pure chassis change" (v2(36)).
- **what:** The workflow-completion contract by which a step's declared `output_field` (singular) or `output_fields` (plural) determines how `extractWorkflowResult` shapes the final result; the singular case was silently mishandled before the fix, and `complete_workflow`'s `result_from` is the fixed, now-correct field the diagnose-agent's own complete step uses.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 gamesdesign entry.
- **relations:** Gamesdesign silent-no-op bug; diagnosis-loop chassis integration architecture.

### Diagnosis→fix loop workstream (founding)
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** NOTES_running_fixloop(9).md THREAD STATE AT CUTOVER: "DECIDED AND RECORDED... STILL OPEN (F2-phase)... FIRST ACTION: slice F0.1 with pre-registered criteria."
- **what:** The founding thread (2026-07-06/07) that pivots the read-only diagnosis loop into a diagnosis→fix system: documented intake (`needs_diagnosis` work item, `pipeline='diagnose'` namespace), live per-iteration reasoning persisted to a new `diagnosis_artifacts` table (kind ∈ bundle|iteration_note, written through inside the assemble action — off the parallel tools-chat's `doc_notes` surface, only the terminal note relayed there), fixes produced by a SEPARATE fixer agent with an isolated git write token (spawn-gate pattern) producing a constrained edit plan validated by gofmt+build before any PR, and a council of parallel specialist reviewers feeding a decision-maker (see hard_veto flag concept below). This is the same workstream documented in far greater operational detail in `docs024_key_docs_latest/fixloop_eg_dartsonline/` — this file is its origin notes.
- **sources:** NOTES_running_fixloop(9).md (full); NOTES_running_synthesis_v4(39).md 2026-07-06/07 entries (same founding, condensed).
- **relations:** Loop-worthiness test doctrine; hard_veto council flag; diagnosis loop; roadmap-phase enforcement gap.
- **verify-later:** `diagnosis_artifacts` table; the fixer agent's isolated write-token/spawn-gate; cross-reference against `docs024_key_docs_latest/fixloop_eg_dartsonline/` for the fuller, later-stage version of this same concept.

### Council hard_veto flag / decision-maker model
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** "Q-D veto semantics decided (owner)... Flag-based: DEFAULT = decision-maker weighs all opinions; a hard_veto flag at reviewer/pipeline/tool/component scope makes that reviewer's negative verdict a BLOCK" (NOTES_running_fixloop(9).md).
- **what:** The fix-loop's review-arbitration model: a parallel council of specialist reviewers (guidelines/reuse/bug-historian/compliance/per-pipeline guardians) feeds a decision-maker by default (advisory), except where a `hard_veto` flag is set at reviewer/pipeline/tool/component scope (accessibility and legal are the motivating cases), which makes that reviewer's negative verdict an unconditional block. A guideline-gap found during review is a SIDE-TASK (a work item that drafts an amendment PR against the guideline docs, human-terminal) rather than something that blocks the fix.
- **sources:** NOTES_running_fixloop(9).md "Q-D veto semantics decided" and "F0/F1 design settled" (DECISIONS).
- **relations:** Diagnosis→fix loop workstream founding.

### Loop-worthiness test doctrine
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Owner asked whether the loop fits the dartsonline quality problem. Answer: decomposed via the new LOOP-WORTHINESS TEST (symptom-not-feature; mechanism-plausible; not one-query-answerable; single-symptom)" (NOTES_running_fixloop(9).md); a fifth criterion (verify symptom currency at intake) added after a pilot candidate "evaporated" (was fixed live before the loop ran).
- **what:** A pre-registered five-criterion test for whether a candidate bug is worth running the diagnosis/fix loop on: it must be a genuine symptom (not a disguised feature request), the mechanism must be plausible from code, it must not be answerable by one query, it must be a single coherent symptom, and its currency must be reverified at intake (since bugs can be fixed out from under a pilot mid-triage — this happened twice in this thread alone).
- **sources:** NOTES_running_fixloop(9).md multiple 2026-07-07 pilot-selection entries.
- **relations:** Diagnosis→fix loop workstream founding.

### Roadmap-phase enforcement gap (builder item 6)
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "VERIFIED IN CODE: 082_submit_domain_unified.sh — grep confirms ONLY --mission/--mission-file exist, no --roadmap anywhere. build-site-planner prompt — the {{if .roadmap_brief}}...{{end}} block has NO else" (NOTES_running_fixloop(9).md).
- **what:** A platform-wide defect (reclassified from a single-site fix into the builder thread's main queue item) where no domain-submission path ever produces a phased roadmap, so a site's Tier-3 roadmap phase rules simply vanish rather than degrade — an absent decision point, not a hidden mechanism. Fix shape: a new post-classification hop writing a phased roadmap for commerce-shaped domains, enforced at three existing relay-wide points (strategist prompt, planner deliverability validation, built-grounded nav) rather than per-site.
- **sources:** NOTES_running_fixloop(9).md "TWO CORRECTIONS: amendment path under-specified; bug is platform-wide"; NOTES_running_synthesis_v4(39).md 2026-07-07 mirror entry.
- **relations:** Diagnosis→fix loop workstream founding; work-item relay / builder-generations architecture; curated best-in-class standing expectation.

### Curated best-in-class standing expectation
- **category:** development-guide
- **status-signal:** aspirational
- **status-evidence:** "Best-in-class/curated-list idea homed: standing expectation (guides+tools+news+non-affiliate curated top-N) + 'not-original-can-still-be-best' clause → 001_development_guide" (NOTES_running_synthesis_v4(39).md, 2026-07-07).
- **what:** A proposed platform-wide addition to the development guide requiring every commerce-shaped domain to carry a baseline of guides, tools, a news feed, and a curated non-affiliate top-N list — enforced the same way as the roadmap gap (relay-wide strategist/planner prompts, not per-message or the constitution) — with the explicit doctrine that "useful-but-unoriginal still counts as best-in-class."
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-07 "pilot candidate 2" entry; NOTES_running_fixloop(9).md "Builder queue item 7" references.
- **relations:** Roadmap-phase enforcement gap; diagnosis→fix loop workstream founding.

### Work-item relay / three-generation builder architecture
- **category:** NEW:site-build-orchestration-generations
- **status-signal:** partial
- **status-evidence:** "THREE generations coexist... GEN-3 component/spec/DB era = pageflow-builder v20 ACTIVE + site-work-orchestrator (queue-native sibling)... §B3 CLOSED: spine = the work-item relay" (NOTES_running_synthesis_v4(39).md, 2026-07-04).
- **what:** A builder-thread inventory found three coexisting generations of "build a site" orchestration on the platform (GEN-1 template era; GEN-2 in-memory multipage v1≈v2; GEN-3 component/spec/DB era) with ~8 overlapping top-level "build the site" orchestrators, only one of which (`pageflow-builder`) is the active monolith. Separately, a queue-native work-item relay (`domain-submitter → needs_domain_research → build-dispatch-loop → domain-research-classifier → needs_strategy → domain-strategist → needs_briefing → build-briefing-agent → needs_site_plan → build-site-planner → needs_page/needs_content_page → page-build-handler`) was traced end-to-end via `reconcile_site_plan`'s routing table and confirmed to reach the builder NATIVELY — established as the real spine, with `pageflow-builder` demoted to "intake convenience." A commented-out `"tool"` route in the same routing table is the mechanism gap blocking tool/infographics pages from the relay.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-04 "§B0" through "§B3 CLOSED" entries.
- **relations:** Roadmap-phase enforcement gap; adoption pipeline; vertical-exemplar-researcher hop; site-quality programme handoff.
- **verify-later:** `RUNBOOK_builder_route.md`, `load_work_item_actions.go` routing table, the un-consolidated GEN-1/2 legacy orchestrators (Q1/Q5 consolidation candidates, left open).

### Vertical-exemplar-researcher / competitor synthesis hop
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** "§B4 CLOSED on quality... Landscape verified: three real vertical leaders; causal synthesis (reasons not copies); confidence 0.82. Strategy QUOTES the hop and builds the moat on it" (NOTES_running_synthesis_v4(39).md, 2026-07-06).
- **what:** A new relay hop (`needs_vertical_research` → `vertical-exemplar-researcher`) inserted between the domain classifier and the strategist to close a gap where the classifier captured `competitors_found` names but nothing ever researched them: it runs shallow crawls of 3 vertical exemplars (vs. adoption's one deep crawl of the site itself), synthesises causal reasons (not copied content) into a `site_specs` row (`aspect=vertical_landscape`), which the strategist prompt reads wholesale and demonstrably used to shape a real site's differentiator. Its first live deployment stalled because the seed migration copied `agent_definitions` columns from a donor missing the spawn-consumed `command`/`image_tag` columns (defaulted to the stale `latest` image tag) — fixed by copying from a fresher donor and flagged as a recurring `image_tag` default-value trap.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-04 through 07-06 (§B4 sequence, full).
- **relations:** Work-item relay / builder-generations architecture; roadmap-phase enforcement gap.

### Site-quality programme handoff
- **category:** content-quality
- **status-signal:** partial
- **status-evidence:** "site-quality programme HANDED OFF to its own runbook... 0 nav / 0 img / 0 svg / 0 script on ALL pages" (NOTES_running_synthesis_v4(39).md, 2026-07-06).
- **what:** Following the platform's first recorded domain→deployed-site milestone (dartsonline.com), a measured baseline (four rendered pages, all missing nav/images/svg/script, thin CSS variable usage, near-zero internal links) triggered a dedicated handoff (`RUNBOOK_site_quality.md`) splitting remaining work into stuck-dispatch (chrome/design/imagery), delivered-but-poor (content depth, links), and never-in-scope (feeds/RSS/graphics/games, disabled improvement-sweep) — with a live hypothesis that the relay path lacks the monolith's `render_site_components` chrome step, explaining nav-zero across every page.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-06 "MILESTONE recorded" and "site-quality programme HANDED OFF" entries.
- **relations:** Work-item relay / builder-generations architecture; diagnosis→fix loop workstream founding (the same "unresolved_cta" defect class recurs across both threads).

### Kafka completion-payload size guard (message-too-large bug + shared result contract)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "§7E attempt 1 (17933a83) FAILED: 'message too large'... the number: diagnosis 6KB, cd 1.27MB ⇒ completion ships FULL collected_data... both §7E blockers fixed in one change-set" (NOTES_running_synthesis_v4(39).md headers, 2026-07-03).
- **what:** A production-triggered bug where an agent's Kafka completion message failed with "message too large" because the completion producer ships the FULL accumulated `collected_data` (1.27MB), not just the declared result (6KB) — triaged to the child-completion producer, fixed alongside a second guard-vs-expansion blocker, and generalised into "Option A": a shared result contract plus a response-size guard applied platform-wide (not just to the diagnose agent).
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-03 §7E entries (headers + surrounding turn log).
- **relations:** SagaCoordinator output_field contract; diagnosis-loop chassis integration architecture.
- **verify-later:** The "Option A" shared result contract's actual deployed shape; whether the size guard applies to all agents or only the diagnose path.

### Child-completion result key convention
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "follow-up #1 resolved: child result key = STEP NAME `call_diagnoser`" (NOTES_running_synthesis_v4(39).md header, 2026-07-03).
- **what:** Confirms (against real orchestrator rows, not guessed) that a spawned child agent's result is read by the parent under the CALLING STEP's own name/output_field (e.g. `call_diagnoser`), not a role-based or agent-type-based key — resolving several rounds of guessed migration SQL in the diagnosis-loop chassis port.
- **sources:** NOTES_running_synthesis_v4(39).md header 2026-07-03; NOTES_running_synthesis_v2(36).md "diagnose migration CORRECTED against real rows."
- **relations:** Diagnosis-loop chassis integration architecture; Workflow default_config location convention.

### Workflow default_config location convention
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "the query result overturned my main assumption... task_workflow / orchestrator_workflow / orchestration_workflow are ALL NULL on every working orchestrator... The workflow lives in default_config" (v2(36), 2026-06-17).
- **what:** A load-bearing, empirically-corrected fact about the chassis schema: an agent's actual workflow (start_step/steps graph) lives in `agent_definitions.default_config`, never in the three separately-named `*_workflow` columns that exist on the table — discovered only by querying real working rows after an entire migration draft was written on the wrong assumption (inferred from the dev-guide's prose example, which showed a workflow object but never named its column).
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 "diagnose migration CORRECTED against real rows."
- **relations:** Real-rows-beat-prose discipline; diagnosis-loop chassis integration architecture.

### GitHub read-token scoping / least-privilege adapter secrets pattern
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** v3(32) DECISIONS: "GitHub read token scoped to the diagnoser via secretKeyRef, not passthrough (turn 25)."
- **what:** `spawn_actions.go` injects `GITHUB_READ_TOKEN` from a shared platform secret only for agent types flagged `isRepoCloningAgent` (currently just `diagnose-agent`), via `secretKeyRef` so the spawning pod itself never holds the token and no other agent type is granted it — the same read-only single-repo PAT the analyser adapter uses.
- **sources:** NOTES_running_synthesis_v3(32).md DECISIONS (turn 25).
- **relations:** Analyser adapter build; diagnose-agent self-contained repo fetch.

### Trust ratchet & capability ceiling model
- **category:** NEW:autonomy-trust-model
- **status-signal:** aspirational
- **status-evidence:** "Bottleneck is trust, not capability... Automation is a per-capability ratchet, not a switch. Bidirectional ratchet." (NOTES_running_synthesis_principles(59) §Trust, reliability, and the ratchet — no implementation evidence anywhere in this file set).
- **what:** A design framework (never implemented, purely a framing document across all five families' shared preamble) for autonomous build/operate systems: trust is per-(tenant, capability), starts at the most conservative level, and moves on a bidirectional ratchet (losable, not just gainable) governed by a "trust ledger." A capability's ceiling is set by verifiability (can ground truth confirm it) × containment (blast radius), independent of how mature/trusted it currently is; the reliability cascade for any task is reuse → generate+verify → compete+judge → HITL, highest-reliability tier first; de-graduation (tightening) may auto-apply on severe evidence, but graduation (loosening) is always confirm-not-initiate — the core safety asymmetry.
- **sources:** NOTES_running_synthesis_principles(59), NOTES_running_synthesis_v2(36).md, v3(32), v4(39) — shared §Trust/§Build-vs-operate preamble (identical across all four).
- **relations:** Governance/HITL confirm-not-initiate model; onboarding/config three-layer model; requirement-mediation model.
- **verify-later:** No known code implements a "trust ledger" or per-capability ceiling table — treat as pure design framing pending stage-2 verification.

### Requirement-mediation model ("right" as balance)
- **category:** NEW:autonomy-trust-model
- **status-signal:** aspirational
- **status-evidence:** "'Right' is a requirement-relative balance among conflicting dimensions (fast/secure/generic/simple/functional). Not pick, not merge." (principles(59) §"Right" as balance, not a single answer).
- **what:** A design framing for resolving competing quality dimensions in generated artifacts: authored solutions are treated as extremes that bound a solution space, a mediator finds the requirement-relative point inside it, priority is ordered (not numerically weighted) and modulated by direction-of-travel, and a satisfied concern demotes from "author" to passive "checker" (re-promoting if a later change breaks it) — unifying single-author and multi-author review as two modes of one process. Multi-author deliberation surfaces tradeoffs but cannot itself resolve value-laden conflicts; those still land with a human/authority model.
- **sources:** NOTES_running_synthesis_principles(59) §"Right" as balance (shared preamble across all four non-fixloop families).
- **relations:** Trust ratchet & capability ceiling model; governance/HITL confirm-not-initiate model.

### Governance/HITL confirm-not-initiate model
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** "Confirm-not-initiate. Decision reasoning is agent-led; the human confirms via a decision package" (principles(59) §Governance and HITL).
- **what:** A framing where every decision publishes its reasoning (not just its outcome) so drift detection is possible; a decision requiring gating routes through exactly ONE confirming component (never reimplemented per producer); a newer proposal for an already-pending target supersedes and expires the older one (freshness over queue order); and inheritance has two precedence directions — normal entries are child-wins, sealed constraints (legal floors, mission non-negotiables) are ancestor-wins.
- **sources:** NOTES_running_synthesis_principles(59) §Governance and HITL (shared preamble).
- **relations:** Trust ratchet & capability ceiling model; diagnosis loop (embodies read-only + human-gated in practice); council hard_veto flag model.

### Onboarding/config three-layer model
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** "The config has three layers with different derivability: mechanical (discovered + probed), conventions (inferred or doc-sourced — confirmed), intent (elicited)." (principles(59) §Onboarding and config).
- **what:** A framing that treats tenant/codebase onboarding as three separate problems with different confirmation mechanisms and different climb rates on the trust ratchet: mechanical facts (probed, confirmable by reality, climb fastest), conventions (inferred-then-confirmed even in docs-authoritative mode, since hallucinated conventions would manufacture drift), and intent (elicited progressively, never "done," captured just-in-time as work happens rather than as an upfront tax).
- **sources:** NOTES_running_synthesis_principles(59) §Onboarding and config (shared preamble).
- **relations:** Trust ratchet & capability ceiling model; doc claim-verification convention (shares "inferred-then-confirmed" DNA).

### Context substrate principles (context engineering framing)
- **category:** NEW:context-engineering-principles
- **status-signal:** aspirational
- **status-evidence:** "Documentation is code... Authored vs derived... References, not copies... Salience over presence" (principles(59) §Context substrate).
- **what:** A cluster of framing principles for how an autonomous system should hold and pass context: distinguish authored (owner + lifecycle, can be wrong) from derived (auto-generated, true-by-being-actual, can't be wrong, only stale) sources; authored layers should point at derived artifacts rather than paraphrase them, so they don't drift; models lose the bigger picture from local-detail salience during reasoning, not from context-window overflow, so the lever is salience management, not window size; a bigger context window is explicitly named as "not the fix for too much context" (context rot).
- **sources:** NOTES_running_synthesis_principles(59) §Context substrate; §Building discipline "A bigger context window is not the fix."
- **relations:** Diagnosis loop (embodies several of these — small bundles, references not pasted copies); B4a embedding-quality finding.

### Reuse-checking retrieval architecture
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** "Checking for reuse is a retrieval problem with a judgement tail, not a generation problem... A maintained capability catalog... turns the first reuse question into a lookup" (principles(59) §Reuse-checking).
- **what:** A framing (partially realised in the actual contextkit/code_symbols build) that reuse-checking should be almost entirely algorithmic: a maintained signature/type/call-graph index answers "have we solved this?" as a query rather than a whole-codebase read; exact-duplicate detection is algorithmic/high-precision (fingerprinting), "similar" detection is semantic/fuzzy (embeddings); a cheap model should narrow candidates for recall, never decide; and any reuse index rots like any derived artifact, needing incremental refresh keyed to real ground-truth cases (past duplications caught in review), since the dangerous error (a missed match) leaves no trace.
- **sources:** NOTES_running_synthesis_principles(59) §Reuse-checking (finding code that already solves the problem).
- **relations:** B4a embedding-quality evaluation finding; code-context retrieval infrastructure.

### Building-discipline edge cases (pre-registered engineering checklist)
- **category:** development-guide
- **status-signal:** aspirational
- **status-evidence:** "Self-referential structures need a cycle guard... A multi-step apply must be all-or-nothing... Bulk operations need bulk confirmation" (principles(59) §Building discipline).
- **what:** A checklist of edge cases the design docs insist be caught before building anything with self-modifying or autonomous behaviour: cycle guards on any parent-link/version-chain walk; transactional multi-write-plus-event apply (outbox pattern) so a crash can't leave a dangling row with no event; reading a consistent point-in-time snapshot when assembling from multiple tables; "one live thing per target, all the way down" (dedup at every layer, not just the queue); bulk-confirmation for large batches; filtering transient/infra failures out of any trust-affecting evidence signal; and "tell not-yet apart from broken" (missing-because-unonboarded degrades gracefully, missing-because-malformed fails loudly).
- **sources:** NOTES_running_synthesis_principles(59) §Building discipline (shared preamble).
- **relations:** Trust ratchet & capability ceiling model; untested-code / behaviour-testing discipline.

### Untested-code / behaviour-testing discipline
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "PRINCIPLE 2026-06-13: untested code is a liability, surfaced by the dedup -move bug... COMPILE/gofmt/vet prove syntax, NOT behaviour." (principles(59)).
- **what:** A hard-won, explicitly codified lesson (triggered by the dedup tool's silent destructive-flag bug) that compiling/gofmt/vet only prove syntax, never behaviour; any destructive CLI operation must be report-only by default; and Go's `flag.Parse()` stopping at the first positional argument is a specific, recurring footgun requiring manual value-flag-aware argument separation in every CLI that takes a positional followed by flags — audited across `dedup`, `thin_versions`, `resolve_targets`, `embed`, `assembler`, `fuse`, `eval_targets`, `dbcontext`.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-13 "PRINCIPLE" entry; NOTES_running_synthesis_v2(36).md 2026-06-14 "runbook code audit" (a second, independent instance of the same class of bug).
- **relations:** Docs archiving toolchain (dedup); building-discipline edge cases.

### Cross-module copy-drift lesson
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "This is the 5th instance of the same root pattern (drafted/validated artefact vs what's on disk): import path, missing file, stale sibling, assumed helper API, now stale CLI." (NOTES_running_synthesis_v2(36).md, 2026-06-17).
- **what:** A hard-won lesson from porting the diagnose engine into the chassis across a real module boundary, surfacing five DISTINCT failure classes in sequence — a wrong import path on a copied file, an entire file silently omitted, a stale (pre-refactor) copy of a sibling file, an assumed helper-package API that didn't actually exist (`datahelpers.ExtractStringSlice`), and a stale CLI binary predating a library change — all of which passed silently in the source module and surfaced only on first build/run in the target. Durable prevention recorded: copy the WHOLE package directory as one unit and diff the file list, rather than cherry-picking files across versions; grep every shared-package call against the real package before authoring, not after.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 (four consecutive build-gap entries); mirrored in `016_additions_assumed_helper_and_cross_module.md` per the notes.
- **relations:** Diagnosis-loop chassis integration architecture; untested-code / behaviour-testing discipline.

### Schema-before-SQL discipline
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Schema: bundle CODE names tables; only \d gives columns AND persistence (hit 4×: page_id, no status column, file_count, the 3-NULL workflow columns)." (v2(36) STATE DIGEST standing lessons).
- **what:** A recurring, explicitly named discipline that code reliably names which DB tables are involved, but only a live `\d` (schema dump) gives real column names and reveals whether a field is persisted at all vs. computed at runtime — hit repeatedly across this project (a wrong `page_id` column, an assumed-but-nonexistent `status` column on `site_plan_sections`, a wrong `fileCount` vs `file_count` JSON key, and the workflow-column misassumption) and eventually generalised into "real rows/examples beat prose/inference" as its own standing lesson.
- **sources:** NOTES_running_synthesis_v2(36).md STATE DIGEST "Standing lessons"; NOTES_running_synthesis_principles(59) multiple 2026-06-14 gamesdesign-diagnosis entries.
- **relations:** Workflow default_config location convention; DB discipline / snapshot_agent convention; gamesdesign silent-no-op bug.

### "Every agent is an orchestrator" spawn pattern
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "ORCHESTRATOR vs long-running service (user's Q): follow 'every agent is an orchestrator' — a thin diagnose-orchestrator spawns a diagnose-agent worker pod... exactly as site-adoption-orchestrator spawns site-adoption-agent." (v2(36), 2026-06-17).
- **what:** A standing platform convention that any substantive in-chassis work (multiple iterations, LLM calls, minutes of runtime) must be wrapped: a thin coordinator/orchestrator agent spawns a dedicated worker agent as a Job pod that runs the actual work and replies to the caller's own responses topic, rather than building a bespoke long-running service that would duplicate the chassis's spawn/await/topic machinery.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 "chassis integration... STEP ZERO search."
- **relations:** Diagnosis-loop chassis integration architecture; workflow default_config location convention.

### Code-retrieval corpus staleness (§7 route)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** v4(39) headers 2026-07-02: "corpus check result: the index is the blocker... the index is of a YEAR-OLD tree."
- **what:** After the diagnosis loop was measured to gain nothing from code retrieval (see B4a finding), a follow-up investigation (§7 route) found the underlying `code_symbols` index itself was built from a year-old stale checkout of the default branch (main stale since 2025-07-14) — a corpus problem, not a retrieval-quality problem — leading to a reindexing effort, ref-pinning strategy, and ultimately the decision to migrate the code-indexer's analysis step onto the already-proven `analyse_repo_local` path.
- **sources:** NOTES_running_synthesis_v4(39).md headers 2026-07-02/03; DECISIONS section.
- **relations:** B4a embedding-quality finding; code-context retrieval infrastructure.
- **verify-later:** Current freshness of the deployed `code_symbols` index; whether the analyse-step migration was applied.

### Real-rows-beat-prose-or-assumption discipline
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Real rows/examples beat prose/inference: the agent_definitions workflow lives in default_config — the dev guide prose didn't say which column; the row did." (v2(36) STATE DIGEST standing lessons).
- **what:** A generalised standing lesson (distilled from several specific incidents across this file set) that when a dev-guide or design doc's prose is ambiguous or silent about an implementation detail, the correct source of truth is a real, live example row/file, not inference from the prose — repeatedly the deciding move that caught a wrong migration or wrong action draft before it was applied.
- **sources:** NOTES_running_synthesis_v2(36).md STATE DIGEST; NOTES_running_synthesis_principles(59) "GROUNDED... tools/provenance/docs design corrected" entries (same pattern, different subsystem).
- **relations:** Workflow default_config location convention; schema-before-SQL discipline; child-completion result key convention.

### Adoption pipeline consumption of vertical/exemplar research
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** "Adoption (user orientation): classifier CONSUMES adoption specs is CONFIRMED from its own workflow (skip-scrape conditional on site_archetype...)" (v4(39), 2026-07-04); "Q4: adoption never calls the classifier... Lean is INVERTED: adoption writes first, classifier consumes under fidelity rules later." (v4(39)).
- **what:** Clarifies the actual (initially misunderstood) relationship between the site-adoption pipeline and the domain-research-classifier: adoption crawls and fingerprints the target site first, writes specs/pages/work items via `apply_adoption_plan`, then hands off to the relay (`needs_domain_research`) — the classifier consumes adoption's output under fidelity rules, rather than adoption calling the classifier directly as first assumed.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-04 "§B2 read" entry.
- **relations:** Work-item relay / builder-generations architecture; vertical-exemplar-researcher hop.

### Data-request channel (adaptive DB-evidence gather)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** v4(39) STATE OF THE WORLD: "Correction (the data_requests channel is real and now wired)... it was dormant from a 3-part wiring gap, now fixed."
- **what:** The mechanism by which a diagnosis-loop verdict can name its own read-only SQL query as a `data_request`, which the loop lints (Guard 2), executes read-only (Guard 3), and folds into the next iteration's bundle — replacing an earlier, more limited "vetted query catalogue only" design once the read-only transaction guard was proven sufficient as the real safety boundary. The catalogue survives as a fast-path/few-shot-examples layer, not the only path. Was found dormant (misdiagnosed twice — first as "a gap to wire", then over-corrected to "dormant by design") due to a three-part wiring gap between `diagnose_route`, `diagnose_load_runtime`, and the migration's `gather_step`.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 DB-evidence design entries; NOTES_running_synthesis_v3(32).md STATE DIGEST "Correction" paragraph.
- **relations:** Model-written SQL guard model; diagnosis loop; doc/query catalogue relevance selection.

### Doc/query catalogue relevance-keyed selection
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "NEW internal/diagnose/docselect.go: DocRule{Doc, Keywords, PathGlobs, Always} + SelectDocs(hypothesis, scope, rules)... TESTS: docselect_test.go" (v2(36), 2026-06-17).
- **what:** A pure, tested, per-iteration selector (`SelectDocs`/`SelectQueries`, sharing helpers) that pulls task-specific reference documents or SQL query templates into a diagnosis bundle only when their keywords/path-globs match the current hypothesis/scope, keeping the always-on constitution small while still surfacing the relevant 003-style contract or domain query "by relevance" rather than dumping every doc into every bundle (a deliberate anti-bloat decision citing the B4a context-rot lesson).
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 "per-hypothesis -doc selection wired into the loop" and "adaptive DB-evidence gather" entries.
- **relations:** Data-request channel; context substrate principles (context rot avoidance).
