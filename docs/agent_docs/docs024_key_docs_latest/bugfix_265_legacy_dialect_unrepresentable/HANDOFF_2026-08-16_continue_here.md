# HANDOFF 2026-08-16 — bug 265 — **SUPERSEDED: THE BUG IS CLOSED.** Kept for the trail; the closure section in `bugs_closed/265_*` is the record. Nothing below is outstanding except the two items explicitly owned by other lanes.

**State (updated 10:26Z):** council **APPROVED r1, all reviewers** (`aba82416…`, verdict 10:22:53Z). Migration 437 **APPLIED 10:24Z + recorded** (3 converted, constraint present, census 0, refusal induced and rolled back). Go half `58b0111ac` inert until the next chassis roll. Steps 1–2 below are DONE; start at step 3.

**Do, in order:**
1. Read the verdict: `SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts WHERE
   correlation_id='aba82416-de79-4452-8730-3e35ca0a15bb' AND kind='council_report';` and the
   note (`doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC`). REVISE →
   revise 437/Go and resubmit with `RESUBMIT_CORR=aba82416…`. APPROVED → continue.
2. Re-run the census FIRST (`count(*) … WHERE input_schema ? 'properties'` must be 3 with the
   3 ids in 437's guard; a 4th row → extend the conversion, don't bypass the guard). Apply
   437 by hand (RUNBOOK), then `--record-only … --note "3 rows converted, constraint present,
   refusal induced"`. Induce the refusal (RUNBOOK) — a zero census is evidence only after
   that. Do NOT run a bare `--apply` (takes other lanes' pending files); the runner dry run
   currently takes >110 s, so scope or skip it.
3. Update: CLC-015 status (register + index row) → "LIVE at the table, Go awaits roll";
   bug file STATUS block; NOTES; README. Bug stays in `bugs_open/` until the Go half is
   rolled AND the constraint has refused something real or been induced post-apply — then
   move to `bugs_closed/` naming both paths on the commit.
4. Post-roll: `kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1
   'build provenance'` → `git merge-base --is-ancestor 58b0111ac <stamp>`.

**Same-file passenger:** `store_generated_component_action.go` in `58b0111ac` carries the
185 lane's `PageWantedLivePredicateFor` hunk (named in the message). Nothing to undo.

**Surfaced for other lanes (in the bug file):** report-dossier `body` `source: llm` vs its
seed's "never LLM-authored" (gripper-dossier lane); `site-header` wrapper-less v2 shape.
