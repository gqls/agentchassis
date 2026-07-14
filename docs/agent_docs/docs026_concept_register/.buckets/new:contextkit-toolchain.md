
<!-- SOURCE: U15_docs019_running_notes.md -->
### contextkit CLI toolchain
- **category:** NEW:contextkit-toolchain
- **status-signal:** deployed
- **status-evidence:** v2(36) STATE DIGEST: "analyser (-exclude + *(N).go skip...), resolve_targets (lexical), embed (semantic, Ollama nomic), fuse (RRF -json k=60), eval_targets..., assembler, dbcontext, dedup, thin_versions, cmd/bundle..., cmd/diagnose."
- **what:** A family of small, report-first, behaviour-tested Go CLIs built to prototype and measure context-assembly/diagnosis before chassis porting: `analyser` (Go-AST symbol index with exclude/dedup-skip), `resolve_targets` (lexical scoring), `embed`/`fuse` (semantic + RRF), `eval_targets` (recall@N/MRR against ground truth), `assembler` (composes a bundle: constitution + docs + schema + symbol bodies + runtime), `dbcontext` (read-only DB gather: `\d`, `-rows`, `-capabilities`, `-runtime-site`), `dedup`/`thin_versions` (docs-archiving tools), `cmd/bundle` (orchestration wrapper: runs dbcontext then assembler), `cmd/diagnose` (dev/test harness for the loop).
- **sources:** NOTES_running_synthesis_principles(59) multiple 2026-06-13 entries (tool-building); NOTES_running_synthesis_v2(36).md STATE DIGEST.
- **relations:** Diagnosis loop; docs archiving toolchain; code-context retrieval infrastructure.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Symbol-body slicer (analysis.ReadSymbolBody)
- **category:** NEW:contextkit-toolchain
- **status-signal:** deployed
- **status-evidence:** v3(32) STATE DIGEST: "`analysis.ReadSymbolBody`... TESTED... and proven BYTE-IDENTICAL to `cmd/assembler`'s real bundle output."
- **what:** The single shared implementation (in `internal/analysis`, not duplicated per-consumer) that turns a `path:Symbol` scope entry into source text by slicing the analyser's already-recorded `start_line..end_line` spans — never re-parsing — used identically by `cmd/assembler` and the chassis `diagnose_assemble_bundle` action, closing a prior stub.
- **sources:** NOTES_running_synthesis_v3(32).md STATE DIGEST + DECISIONS.
- **relations:** Diagnose-agent self-contained repo fetch; contextkit CLI toolchain.

<!-- SOURCE: U15_docs019_running_notes.md -->
### contextkit CLI toolchain
- **category:** NEW:contextkit-toolchain
- **status-signal:** deployed
- **status-evidence:** v2(36) STATE DIGEST: "analyser (-exclude + *(N).go skip...), resolve_targets (lexical), embed (semantic, Ollama nomic), fuse (RRF -json k=60), eval_targets..., assembler, dbcontext, dedup, thin_versions, cmd/bundle..., cmd/diagnose."
- **what:** A family of small, report-first, behaviour-tested Go CLIs built to prototype and measure context-assembly/diagnosis before chassis porting: `analyser` (Go-AST symbol index with exclude/dedup-skip), `resolve_targets` (lexical scoring), `embed`/`fuse` (semantic + RRF), `eval_targets` (recall@N/MRR against ground truth), `assembler` (composes a bundle: constitution + docs + schema + symbol bodies + runtime), `dbcontext` (read-only DB gather: `\d`, `-rows`, `-capabilities`, `-runtime-site`), `dedup`/`thin_versions` (docs-archiving tools), `cmd/bundle` (orchestration wrapper: runs dbcontext then assembler), `cmd/diagnose` (dev/test harness for the loop).
- **sources:** NOTES_running_synthesis_principles(59) multiple 2026-06-13 entries (tool-building); NOTES_running_synthesis_v2(36).md STATE DIGEST.
- **relations:** Diagnosis loop; docs archiving toolchain; code-context retrieval infrastructure.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Symbol-body slicer (analysis.ReadSymbolBody)
- **category:** NEW:contextkit-toolchain
- **status-signal:** deployed
- **status-evidence:** v3(32) STATE DIGEST: "`analysis.ReadSymbolBody`... TESTED... and proven BYTE-IDENTICAL to `cmd/assembler`'s real bundle output."
- **what:** The single shared implementation (in `internal/analysis`, not duplicated per-consumer) that turns a `path:Symbol` scope entry into source text by slicing the analyser's already-recorded `start_line..end_line` spans — never re-parsing — used identically by `cmd/assembler` and the chassis `diagnose_assemble_bundle` action, closing a prior stub.
- **sources:** NOTES_running_synthesis_v3(32).md STATE DIGEST + DECISIONS.
- **relations:** Diagnose-agent self-contained repo fetch; contextkit CLI toolchain.
