# NOTES — experience register (append-only, newest at the bottom)

## 2026-07-24 — workstream born from discussion (session "experience register")

Owner brief: a directory of small reusable experiences — carousel behaviour is approved but
what a click on a card does is undocumented and re-invented each use; wanted: a register
recording e.g. "read more" expands in place, card click → article page → onward links to
info/tools. Several thousand eventually, used as base plans, forked per site, testable end to
end; links get "probably should go" instead of guesses.

### Round 1 — four parallel doc searches (experience loop / travelling docs+tool-improver /
### concept register+feature-builder / UX+link-integrity)

Findings that held: no experience-register-like construct exists anywhere (no table, no stub);
`content_components` has no behaviour/destination fields; bug 023 is the exact complaint
("nothing pairs label and URL"); EXPERIENCE_PLAN has the right vocabulary (journeys with
page/control/action/observable-outcome, promise ledger, criteria) but is per-site, authored
from a blank page, one exists; LNK-011 reserved an agent boundary for intent-matching, never
built; concept register = proven harvest→adversarial-verify→register pipeline (1,633 entries).

Positions I took in round 1 that were later corrected:
- Proposed a concept-register-style flat-markdown register seeded into a table.
- Leaned approval-by-use over per-entry approval.

### Owner corrections (round-1 reply)

- **Travelling docs was created for this** — documents a tool's provenance and direction;
  closer ancestor than I stated. Correct: I had under-weighted it.
- **DB-based storage preferred.**
- Taxonomy: our own, loosely based on the UX industry playbook.
- **Approval per experience**; formalise the acceptance-test side.
- vonc creates the first viable product; question: does that include T4, or does owner
  initiate?

### Round 2 — three parallel searches (site-plan machinery / travelling-docs mechanics /
### vonc pilot state)

Load-bearing findings, each verified by file:line in the reports:
- `doc_plans` has **no site_id** — already library-level. subject_type CHECK today =
  tool|pipeline|experience|action (163, 184).
- **doc_plans is exact-key only** — no metadata column, no structured search; RAG indexing of
  plans is design-intent, not a verified pipe. Travelling docs give provenance/direction, NOT
  selection → the register table supplies selection, the travelling doc travels with each
  entry (the content_components + doc_plans 'tool' precedent).
- **184 split contract found** (verified first-hand this session, not just from the agent
  report): `docResolveSubject` (write_doc_plan_action.go:136-144) rejects 'action' and is
  shared by ALL THREE doc actions (write:59, append_doc_note:59, load_doc_context:56), so the
  184-seeded action rows are unreachable through the doc actions. Separately
  `persist_diagnosis_note_action.go:78` allows only tool/pipeline — it silently skips even
  'experience' (163 missed it). Filed as **bugs_open/064**.
- Criteria fence: read-time parse only; stale criteria ×3 sightings, `-EDIT` placeholders
  silently skipped, unclosed fence → silently "". Formalisation target confirmed.
- Site-plan: page set decided in build-site-planner `plan_site`; `roadmap_brief` is the
  authority-override precedent ("the roadmap is the authority") — the experience_brief hook
  copies it. Preservation set + owned-page guard are the re-plan-safety precedents.
- **Dormant `site_flows`/`flow_pages`** discovered: narrative-funnel schema
  (awareness/consideration/conversion), ZERO platform/ references. Design supersedes the
  tables, adopts the funnel-stage vocabulary (recorded in PLAN §4).
- vonc/T4: nothing auto-fires T4 (session-driven); run-12 plan REJECTED (must not build);
  re-fire 092 only after tools-api deployed+smoke-POSTed. tools-api: design approved corr
  278a37c3 (2026-07-24 09:59), implementer refused twice (max_tokens; map-not-string), round
  3 in flight, zero code landed, no PR; branch feat/278a37c3 carries no tools-api commits.
  Owner hard gates: PR merge + 4 bastion/tunnel infra tasks. T5.1 journey runner not started.

### Owner rulings (round-2 answers, recorded in PLAN §2)

Substrate = register table + travelling doc. Site-plan hook = experience_brief aspect.
Taxonomy = layered, seeded from harvest. vonc = **wait for tools-api** (my recommendation was
the feasibility veto's static cut; owner ruled wait — the full debate path is the first
viable product; harvest waits for it).

### Missteps this session

- Round 1: stated travelling docs was per-instance-tool-only prior art and proposed flat
  markdown storage — under-read; owner corrected; round-2 research confirmed the substrate
  is library-level by construction (no site_id).
- Round 1: recommended approval-by-use; owner ruled per-experience approval. Synthesis kept:
  'proven' remains an evidence upgrade after council approval.

Artifacts written this session: PLAN, RUNBOOK, this file, README_where_we_are, design/ (4
draft artifacts, nothing applied), bugs_open/064, 016b §9 pattern entry, memory pointer.
