# 405 — `detected-item-promoter`'s known-good door tests the HANDLER's competence, never the FINDING's provenance — so LLM-audit opinions ride a door built for mechanical defects

**Filed 2026-08-25** by the `loanzy_uk_example_site` lane, at the owner's direction ("find the
thread that produced that promoter or file a new bug for it"), while executing
`PLAN_2026-08-25_switch_off_the_evolutionary_rewrites_and_switch_the_loop_back_on.md`.
**Status: OPEN. Severity: medium** (the acute inflow is stopped by migration 623 + the record-mode
seam; this files the STRUCTURAL residue so it is not folklore).
**Producing thread, named:** the promoter was built by the **`bugs_closed/083`
(detected_findings_never_reach_a_handler) lane** — commit `3c6354059` 2026-08-15, "083 candidate 2
BUILT (owner ruling 2026-08-15)", lineage migrations `430 → 444 → 453/454 → 458 → 465 → 480` —
and 083 was closed 2026-08-22 by the **`bugs_open/277` lane** (live session notified today).
`scripts/who-owns.py 083`: closed; no open owner. Hence a NEW file rather than a CONTRIB.

## 1. The mechanism (read from the live `pre_query`, 2026-08-25 — not inferred)

`detected-item-promoter` (scheduled task, 900s, live since 2026-08-15) promotes any
`site_work_items` row at `status='detected'` with a non-empty handler through four doors:
pipeline ∈ (build, content, design); handler is a live agent; **the (item_type, handler_agent)
pair has ≥1 lifetime completion** (live ∪ archive); the pair is above a 25% success floor.

Every door interrogates the HANDLER and the PAIR's history. **No door asks what produced the
row.** A `content_rewrite → page-build-handler` row passes the known-good test identically whether
it was filed by a mechanical check that measured a defect ("this required field is empty") or by an
LLM auditor's opinion ("this page could be better") — because the pair's thousands of completions
were earned by BOTH populations under one item_type.

## 2. Why that is a defect and not a design choice, with the evidence

- **The promoter's own description says** "Does NOT re-enable improvement-sweep" — its authors
  understood it as draining a triage backlog, not as a dispatch route for auditors. But
  **[MEASURED 2026-08-25, live ∪ archive]** 26 LLM-audit rows (`spec->>'audit_source'` ∈ the six
  model-seat sources) were promoted between 2026-08-20 and 2026-08-24 **while the sweep was
  disabled** (triaged_at 08-20 14:59, 08-22 11:26, 08-24 10:27, 08-24 22:21; the 08-17 12:40 and
  08-22 18:4x batches excluded as sweep/hand-run triage). Rewrites the owner believed were off
  were arriving by this door.
- **IMP-054's premise is voided** (register): "a lone discovery run files findings nothing can
  ever see" was true when written (2026-08-09) and false six days later. `detected` is a queue,
  not a shelf — LANDMINES.md "`detected` is a QUEUE, not a shelf" (2026-08-25) is the prospective
  half of this file.
- The known-good doors were REVIEWED and tightened three times (444, 454, 465, 480) and every
  tightening was about handler competence — the provenance axis was never on the table, because
  083's population WAS mechanical. The blind spot is inherited scope, not an error by its authors.

## 3. Why this is filed on first-hand verification rather than a `090` run (owner ruling 2026-07-31)

The claim asserts no hidden cause: every clause of §1 is the live `pre_query` read verbatim, and
§2's counts are dated queries over `site_work_items ∪ archive` reproduced in
`loanzy_uk_example_site/NOTES` (2026-08-25 evening). There is no competing hypothesis to refute —
the door demonstrably does not read provenance, and rows demonstrably passed it. A `090` run would
re-read the same two artefacts. (The related JUDGEMENT question — whether opinions should dispatch
at all — went through the council instead: RFC_056, trail `d1342f2a`.)

## 4. What already contains it (so the residue is stated exactly)

- **Migration 623 (APPLIED 2026-08-25)**: the four model seats are off the improvement loop's
  path, so no NEW opinion rows are being filed. **0 LLM-audit rows sit at `detected` today.**
- **The record-mode seam (`filing_mode: record`, in the running binary since v1.0.1339)**: once
  RFC_056's round-2 verdict lands and migration 624 applies, model-seat findings are born
  `deferred` + `handler ''` — never `detected`, so never this promoter's candidates.
- **Residue A:** any OTHER route that files an opinion-shaped row at `detected` with a proven
  pair — a hand-fired auditor without record mode, a future seat whose author does not know this
  file — re-opens the door. The promoter itself still cannot tell.
- **Residue B:** `tool-acceptance-tier4` and similar non-improvement-loop audit sources also file
  `detected` rows the promoter handles; they are DEFECT-shaped today, but the same one-door
  design means nobody would notice if one became opinion-shaped.

## 5. Fix candidates, ordered by what closes the door (not by effort)

1. **An origin class the promoter can read** (closes the door): `write_audit_findings` stamps
   `spec.origin = 'model_opinion'` on every finding whose seat ran `execute_llm_prompt` (or
   simply: always, with values `mechanical|model`), and the promoter's `scored` CTE gains a fifth
   door: `COALESCE(wi.spec->>'origin','') <> 'model_opinion'`, with the held reason
   "model opinion — release by hand or via record-mode". Go + pre_query in lockstep (one
   migration + one Go change; the drift pair should be pinned by a test the way
   `dedup-index-go-list-lockstep` is).
2. **A hand-kept source list in the pre_query** (cheaper, drifts): door on
   `wi.spec->>'audit_source' NOT IN (<the six>)`. Wrong by omission the day a seventh seat ships;
   acceptable only as an interim if (1) stalls.
3. **Do nothing beyond 623/624** and accept Residues A/B as documented. Honest, free, and the
   next session to hand-fire an auditor pays for it.

## 6. How to verify a fix

Induce, don't wait: file one synthetic `detected` row with a proven pair and
`spec.origin='model_opinion'` (or an audit_source from the six), wait two promoter ticks, assert
it is HELD with the named reason in the tick's `doc_notes` row — and file the mechanical control
row in the same breath, asserting it IS promoted. Both directions, same run.

## 7. Routing

- The promoter's living owners: the `083`/`277` lineage — `bugs_open/277` session notified
  2026-08-25 with this file's path. If that lane takes candidate 1, the Go half touches
  `write_audit_findings` — coordinate with RFC_056 (this lane), whose record-mode seam is
  adjacent but NOT a substitute (record mode changes what the six seats FILE; candidate 1 changes
  what the promoter ADMITS, whoever filed it).
- 016b §9 pattern (transferable): **a "known-good" test scoped to one axis certifies every other
  axis by accident** — the pair's history voted, the row's provenance never did.
