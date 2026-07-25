# 064 — doc-subject split contract: the DB accepts subject types the doc actions reject

Filed 2026-07-24 (session "experience register", found during register design research and
verified first-hand). Status: **CLOSED 2026-07-25 — fixed `c9cc95a5a` AND LIVE on
v1.0.1156, both branches behaviourally verified on the running pod** (see "Live verification"
below). Severity: low-medium — no data loss, but a seeded capability was silently unreachable
and one note path silently dropped valid subjects.

## Mechanism

The `doc_plans`/`doc_notes` `subject_type` contract has **four enforcement points**, and the
last two subject-type additions each moved a different subset:

| # | Point | Accepts today |
|---|---|---|
| 1 | `doc_plans_subject_type_check` (DB, last set by migration 184) | tool, pipeline, experience, **action** |
| 2 | `doc_notes_subject_type_check` (DB, last set by 184) | tool, pipeline, experience, **action** |
| 3 | `docResolveSubject` — `platform/orchestration/actions/write_doc_plan_action.go:136-144`, shared by `write_doc_plan` (:59), `append_doc_note` (`append_doc_note_action.go:59`), `load_doc_context` (`load_doc_context_action.go:56`) | tool, pipeline, experience |
| 4 | `persist_diagnosis_note` subject gate — `persist_diagnosis_note_action.go:78` | tool, pipeline |

- **Migration 163** (+`experience`) moved 1, 2, 3 but missed 4.
- **Migration 184** (+`action`) moved 1 and 2 **only** — its stated purpose was to give the
  three shared fix-loop actions travelling PLAN+NOTES, answering council-gate objection
  `5a65ec4c` (`184_travelling_action_subjects.sql:6-16`), and it seeded three `action`
  PLANs by raw INSERT.

Point 3's own comment states the invariant this violates (`write_doc_plan_action.go:138-141`):
*"Kept in lockstep with the doc_plans/doc_notes subject_type CHECK constraint… A value the
DB accepts but this gate rejects — or vice versa — is a split contract; move both together."*

## Consequences

1. **The three 184-seeded `action` docs are unreachable through every doc action.**
   `load_doc_context` errors on `subject_type='action'`, so no workflow step can consult
   `diagnose_read_repo_files` / `diagnose_prepare_fix_commit` / `diagnose_build_gate`'s
   PLANs; `append_doc_note`/`write_doc_plan` likewise refuse, so the docs can only evolve by
   raw SQL/migration. The convention the council seat cited is therefore *still* not
   operational — the rows exist for humans reading the DB, not for agents.
2. **Any future step configured `subject_type='action'` fails** at `docResolveSubject`.
3. **`persist_diagnosis_note` silently skips `experience` subjects** (163's miss): a
   diagnosis about an experience subject returns `persisted: false` — and logs a
   **misleading** reason, `"no explicit subject — skipping"`, when the subject was perfectly
   explicit and only its type fell outside the stale allowlist
   (`persist_diagnosis_note_action.go:78-82`).

## Root cause

Enum-like contract with four hard-coded enforcement points and no single-source discipline —
the same class as the dedup-index/`workItemTerminalStatuses` lockstep bug (fixed v1.0.1127).
See 016b §9: "A schema CHECK and its code gates are one contract".

## Fix candidates

1. **(Preferred) Single-source the Go side**: one shared `validDocSubjectTypes` set used by
   both `docResolveSubject` and `persist_diagnosis_note` (point count 4 → 2), widened to
   include `action`; decide point-4 policy deliberately (diagnosis notes for `action` and
   `experience` subjects are exactly what the note trail is for); fix the misleading log
   line; add a table-driven gate test + a lockstep test asserting the Go set matches the
   live CHECK values.
2. Minimal: add `'action'` at point 3 and `'experience','action'` at point 4.

## Verify

- Unit: table-driven test over all subject types × both Go gates.
- Live: a scratch `load_doc_context` step with `subject_type='action'`,
  `subject_key='diagnose_build_gate'` returns the 184-seeded plan (today: error).
- `SELECT conname, pg_get_constraintdef(oid) FROM pg_constraint WHERE conname IN
  ('doc_plans_subject_type_check','doc_notes_subject_type_check');`

## Ownership / coordination

184 was seeded by the fixloop/feature-builder thread; `travelling_docs/` owns the substrate.
The **experience_register** workstream plans to add subject_type `'experience-pattern'` in
its Phase-2 change-set and will fix this in passing — the full four-point checklist is
`docs/agent_docs/docs024_key_docs_latest/experience_register/design/subject_type_addition.md`.
If another thread touches doc subjects first: take the fix, note it here, follow that
checklist (image before migration, or the widened CHECK just recreates 184's split).

## Fix taken — 2026-07-24, session "bugfix 064 subject split contract"

Taken per the invitation above (checked first: `who-owns.py` verdict read, experience_register
P2 not started, no open `site_work_items`, no competing commits on the four files). Fix
candidate 1 (preferred) implemented, commit **`c9cc95a5a`**:

- NEW `platform/orchestration/actions/doc_subjects_common.go` — canonical
  `validDocSubjectTypes = tool|pipeline|experience|action` (the `work_items_common.go`
  v1.0.1127 lockstep idiom) + `docSubjectGateReason` with **distinct** skip reasons.
- `docResolveSubject` and the `persist_diagnosis_note` gate both consume it: point count
  4 → 2. Point-4 policy decided deliberately (per the checklist): any subject the substrate
  accepts can carry a diagnosis note; the misleading `"no explicit subject"` log for
  explicit-but-unsupported types is fixed (`"unsupported subject_type …"`).
- Tests: table-driven vocabulary × both gates; distinct-reasons regression; a
  **migration-lockstep test** that parses the newest `sql_for_agents` migration recreating
  `doc_plans_subject_type_check` and fails on drift — verified by induced fault (widening the
  Go list alone fails naming `184_travelling_action_subjects.sql`). Run in a clean
  `git archive HEAD` overlay (shared tree was broken by unrelated WIP): all pass.
- Scope: **Go-side only, no migration** — the DB was already wider than the code, so there is
  no image-before-migration ordering risk. `'experience-pattern'` deliberately NOT added
  (that stays in the experience_register P2 change-set; the lockstep test now catches a
  P2 migration that lands without the Go entry).
- Council gate: submission corr `2b03e56d-d770-4d8d-a1e2-4d3a46494927` — **APPROVED round 1**
  (1 low-severity advisory: edit 5 — the test-comment tidy — is non-substantive; accepted,
  it stays because the comment was factually stale). bug_historian's residual concern
  (a possible THIRD hard-coded allowlist beyond the two gates) answered with evidence
  post-verdict: `grep -rn '"tool"' --include='*.go' platform/ internal/ pkg/ | grep
  '"pipeline"'` → only hit is `validDocSubjectTypes` itself; the only other subject-type
  comparison grep hit is `page_role_validator.go` (page *roles*, unrelated domain).

## Live verification — 2026-07-25, image v1.0.1156 (CLOSED)

Image rolled (pod `agent-chassis-bdd85599c-l8sv2`, deployment image
`docker.io/aqls/agent-chassis:v1.0.1156`, `c9cc95a5a` confirmed ancestor of HEAD).

1. **Discriminating pod-grep PASSED** — both strings the change CREATED present in the
   running binary (`strings /app/agent-chassis`): `unsupported subject_type` ×1,
   `subject_type must be one of` ×1; positive control `persist_diagnosis_note` ×6.
2. **Positive branch PROVEN live** — scratch one-step agent_definition
   `scratch-docsubject-064` (sole step `load_doc_context`, config
   `subject_type='action'`, `subject_key='diagnose_build_gate'`), fired via the kcat
   `orchestrate` envelope (durable-write-guard 021 harness pattern). Orch `91d550a2`:
   `complete | COMPLETED`, `collected_data->doc_context_result` = `has_plan: true` with the
   full 184-seeded PLAN body + 1 note — the exact call that errored at `docResolveSubject`
   before the fix.
3. **Negative branch PROVEN live** (per [[verify-the-failing-branch]] — the gate must still
   reject, not fail open) — same definition flipped to `subject_type='site'`, orch
   `f74d2442`: `load_docs | FAILED` with the new error verbatim:
   `load_doc_context: subject_type must be one of 'tool', 'pipeline', 'experience',
   'action', got "site"`.
4. **Cleanup verified 0 remaining** — deleted the scratch agent_definition, both
   orchestration_states, 10 audit rows; `agent_error_log` had no rows for either run
   (checked by orchestration_id AND by message content), so nothing for the immune-system
   sweep to mis-triage; the three `action` doc_plans untouched (loader is read-only).

Both consequences 1 and 2 are therefore dead on the live fleet; consequence 3's gate now
accepts `experience`/`action` and logs distinct reasons (unit-tested; the shared
`docSubjectGateReason` is the same code path pod-grepped above). Fixed AND live → CLOSED,
moved to `bugs_closed/`.
