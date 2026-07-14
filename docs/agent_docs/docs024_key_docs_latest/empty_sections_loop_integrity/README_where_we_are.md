Done — the project trio is in docs/agent_docs/docs024_key_docs_latest/empty_sections_loop_integrity/, following the imagery-workstream conventions:

- PLAN_empty_sections_loop_integrity.md — the problem in two sentences, the three-hole root cause, and five phases with live checkboxes: Phase 1 loop-integrity gate (code deployed, SQL 149 + live re-drive
  pending), Phase 2 required_fields_missing check (deployed, SQL 150 pending), Phase 3 meta-commentary guard (code done, needs the rebuild), Phase 4 the robot-hands product-page decision (options A/B/C
  table, B/C recommended, owner call), Phase 5 spin-outs (sectionHasVisibleContent measuring resolved data, more verifiers, fixloop benchmark export). Ends with five success criteria.
- RUNNING_NOTES_empty_sections_loop_integrity.md — the "what this project is about" preamble, a decision log (8 decisions from today with rationale), and the full session-1 record: evidence trail
  (identical 19,364-byte result payloads, pages.sections=[], the two-strike zombie mechanism), the table of every file changed, and the deploy-state verification including the 14:35→14:38 build-timing miss.
- RUNBOOK_empty_sections_loop_integrity.md — operator procedures: pod-binary verification greps (with the build-timing trap written down), applying + verifying SQL 149/150, the re-drive procedure with an
  expected-outcomes table (including what counts as a regression), how to read result._verification verdicts, the dartsonline zero-flag cross-check for the new discovery check, and zombie-backlog triage for
  the ~36 old items.

I also added the workstream to my persistent memory index so future chats recall it, linked to the fixloop and build/deploy-practice memories.

Still open from earlier, now tracked in the plan: rebuild for the meta-commentary guard, apply SQL 149/150, then the live re-drive.


---

