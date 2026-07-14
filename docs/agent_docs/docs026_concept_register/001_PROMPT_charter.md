# Charter — Concept Register extraction (rewritten prompt)

Original request: 2026-07-13 (voice), rewritten for precision. An earlier draft of
the same intent exists at
`docs019_documentation_audit_autonomous_build_and_operate/README_comprehensive_documentation_categorisation.md`
(which proposed a `docs/intentions` directory that was never created).

## The task

Build a complete **Concept Register** for the platform:

1. Sweep **every file under `docs/`** — every paragraph of every documentation
   file. Account for every file in a coverage ledger (text docs read; version
   families read latest-fully + earlier-for-deltas; code/SQL headers scanned;
   binaries listed as skipped).
2. Extract every distinct **concept** — a nameable scope, responsibility, or
   behaviour: missions, pipelines, mechanisms, conventions, contracts, agents,
   tools, and unfulfilled ideas alike.
3. **Classify** each concept. Start from the docs024 documentation-index
   categories as a reference spine, but the taxonomy is open — create new
   categories wherever the content demands. The end state should be categories
   that could each back an expert agent.
4. Tag each concept with a **status signal** from the documents alone:
   deployed / partial / aspirational / superseded / abandoned / unknown —
   with dated evidence. (True state is verified against code in stage 2.)
5. Record **provenance** (source files/sections) and **relations** so later
   stages can trace every concept back to its origins.
6. Write everything into `docs/agent_docs/docs026_concept_register/` only.
   Do not modify any existing document.

## Why (the three-stage programme)

- **Stage 1 (this):** the register — list and classify the concepts.
- **Stage 2:** analyse agent-chassis code, workflows, and DB to determine the
  real state of each concept (deployed and working? drifted? never built?).
- **Stage 3:** create an expert agent per concept area — fully versed in its
  responsibilities and provenance — to join the council of decision makers in
  the diagnosis/fix loop (currently a 2-agent skeleton; see
  `docs024_key_docs_latest/fixloop_eg_dartsonline/SUMMARY_write_step_position_2026-07-12.md`).
  The documentation categories and the council seats should roughly correlate.

## Context the task rests on

- `docs024_key_docs_latest/` is the target format for all documentation: small
  scoped documents indexed in `000_documentation_index(2).md`; cross-cutting
  FOCUS documents collect per-feature views.
- The diagnosis tool (docs019 project) digs out documentation, code, and DB
  evidence for an error, loops until certain, and issues a verdict; it has
  evolved into a fix loop that proposes fixes, with multiple agents contributing
  perspectives and holding veto power over what enters a fix.
