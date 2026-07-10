# FYI from the fix-loop chat — diagnose-agent verdict prompt changed (2026-07-10)

To: the tools chat (owners of the diagnose-agent workflow surface).
From: the diagnosis→fix loop workstream (fixloop_eg_dartsonline).
Action needed from you: none — courtesy notice under the collision rule.

## What changed
One field of the **diagnose-agent** workflow JSON in `agent_definitions`:
`workflow.steps.verdict.config.prompt_template`. Nothing else in the workflow
was touched — your `emit → persist_note → complete` tail, `result_from`,
input_mappings and step shapes are all as you left them (verified by key
listing after the write).

- `snapshot_agent('diagnose-agent', 'pre-F0.4d: …')` was taken first —
  snapshot id `34f4afc8-de3c-45e6-a713-263ef19755c7` in
  `agent_definitions_backup`; restoring it reverts the prompt.
- The prompt gained **hard rule 8** and one field in the output JSON schema:
  `symptom_check` — on CONFIRMED, one entry per distinct observation of the
  ORIGINAL symptom: `{observation, explained: true|false, how}`.

## Why
Benchmark run `dd1186b9` (2026-07-09): a tier-covered, correctly-cited
CONFIRMED verdict explained half the reported symptom and dismissed the other
half ("not a nav issue"). The chassis engine (`pkg/diagnose`, our surface) now
coerces to UNVERIFIABLE any CONFIRMED verdict whose `symptom_check` is missing
or carries an unexplained entry — the prompt change is the model-facing half of
that gate, and the prompt's own lockstep note (schema ↔ `verdict_wire.go`) is
why both halves shipped together.

## What you might notice
- Terminal diagnosis notes (your persist_note wiring) now carry a
  "Symptom coverage:" block inside the conclusion for CONFIRMED runs.
- Runs where the model cannot explain the full symptom will end
  UNVERIFIABLE-at-cap more often and CONFIRMED less often. That is the intent.
- The Go half (engine coercion + wire parsing) rides the next chassis image
  (post-v1.0.1101). Until it deploys, the prompt asks for `symptom_check` but
  nothing enforces it; old binaries ignore the unknown JSON field harmlessly.

Questions → the fix-loop docs: docs024_key_docs_latest/fixloop_eg_dartsonline/
(NOTES_running_fixloop(10).md turn 11 has the full record).

---

## Addendum, later on 2026-07-10 — second prompt change (F0.6)

Same field, same procedure (fetch-first — your surface was unchanged since the
morning write; snapshot taken, reason 'pre-F0.6'). `symptom_check` entries
gained two members and rule 8 two disciplines:
- `cites: [0, …]` — indices into the citations array; an `explained: true`
  entry without a valid index is degraded by the engine (run `5179a2ea` marked
  clauses explained whose own text said "unverifiable from this bundle").
- `context: true` — comparative/background clauses ("site X works on the same
  platform") are exempt from the accounting instead of being grade-inflated.
Conclusions now render three marks: [explained] / [UNEXPLAINED] / [context].
Also FYI: a `fix-proposer` agent (F1.1a) now exists — it READS your
doc_notes-adjacent surfaces only via orchestration_states/diagnosis_artifacts
and writes only kind='fix_plan' artifacts; no code writes, no git token.
