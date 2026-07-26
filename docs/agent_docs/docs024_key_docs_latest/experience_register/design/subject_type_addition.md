# Adding doc-subject type `experience-pattern` — the full change-set (P2)

> **UPDATE 2026-07-24 (bugfix-064 thread, commit `c9cc95a5a`):** 064 is fixed ahead of P2 —
> change-set items 1 and 3 below are DONE, and item 5's tests exist. The Go side is now
> single-sourced in `platform/orchestration/actions/doc_subjects_common.go`
> (`validDocSubjectTypes`, consumed by both `docResolveSubject` and the
> `persist_diagnosis_note` gate, which now also has the distinct-reason log fix). A
> migration-lockstep test (`doc_subjects_common_test.go`) parses the newest migration that
> recreates `doc_plans_subject_type_check` and fails on drift. **P2 therefore shrinks to:**
> add `"experience-pattern"` to `validDocSubjectTypes`, ship the migration (item 2), and
> keep the image-before-migration order (item 6 unchanged; item 4 unchanged). If the P2
> migration file lands without the Go entry, the lockstep test fails the build gate by
> design.
>
> **UPDATE 2026-07-26: 064 is CLOSED** — `bugs_closed/064…`, commit `eb81de7b5`. The fix went
> live on chassis v1.0.1156 (2026-07-25) and the closing session proved both branches with
> live runs: `subject_type='action'` now returns the 184-seeded PLAN, and an invalid type
> fails with the new error verbatim. Re-confirmed in today's v1.0.1167 binary. Nothing in this
> checklist is waiting on it any longer.

The subject_type contract currently has **four enforcement points**, and every addition so
far has missed at least one (bugs_open/064): migration 163 (+`experience`) moved the DB
CHECKs and `docResolveSubject` but missed `persist_diagnosis_note`; migration 184
(+`action`) moved the DB CHECKs **only**, leaving the seeded action rows unreachable through
every doc action. This file is the checklist so `experience-pattern` moves them ALL — and
fixes 064 in passing.

## The four enforcement points (verified 2026-07-24)

| # | Point | Location | Today accepts |
|---|---|---|---|
| 1 | `doc_plans_subject_type_check` | DB constraint (last set by 184) | tool, pipeline, experience, action |
| 2 | `doc_notes_subject_type_check` | DB constraint (last set by 184) | tool, pipeline, experience, action |
| 3 | `docResolveSubject` | `platform/orchestration/actions/write_doc_plan_action.go:136-144` — shared by `write_doc_plan` (:59), `append_doc_note` (append_doc_note_action.go:59), `load_doc_context` (load_doc_context_action.go:56) | tool, pipeline, experience |
| 4 | `persist_diagnosis_note` subject gate | `persist_diagnosis_note_action.go:78` | tool, pipeline |

Point 3's own comment states the rule: *"A value the DB accepts but this gate rejects — or
vice versa — is a split contract; move both together."*

## Change-set

1. **Single-source the Go side** (preferred 064 fix): one shared
   `validDocSubjectTypes` slice/set in the actions package, used by points 3 and 4, with a
   comment binding it to the DB CHECK. Collapses four points to two (DB + one Go source).
2. **Migration**: guarded DROP/ADD of both CHECKs (copy 184's idempotent shape) →
   `('tool','pipeline','experience','action','experience-pattern')`.
3. **Point-4 policy, decided deliberately** (not silently inherited): diagnosis notes
   SHOULD be persistable for `action` and `experience-pattern` subjects (a diagnosis about
   a shared action or a pattern is exactly what the note trail is for); `experience` too
   (163's miss). Also fix the misleading log line — today a valid-key/'experience'-type call
   logs "no explicit subject" when the subject was perfectly explicit and only the type was
   outside the allowlist.
4. **Rename/rekey coverage**: `RekeyTravellingDocs` is invoked only by
   `rename_tool_identity` (tool subjects). Rule for the register instead of new plumbing:
   **pattern names are immutable once `approved`** — supersede under a new name. Enforce in
   the register write path; document in the PLAN.
5. **Tests**: table-driven test over all five types × the two Go gates; a regression test
   that reads the live CHECK values and asserts the Go source matches (the
   dedup-index/Go-list lockstep pattern, v1.0.1127 lesson).
6. **Live verify**: `SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conname IN
   ('doc_plans_subject_type_check','doc_notes_subject_type_check');` + a scratch
   `load_doc_context` step with `subject_type='action'`, `subject_key='diagnose_build_gate'`
   returning the 184-seeded plan (the 064 proof), and one with `experience-pattern` after
   the first entry is written.

## Sequencing

Ships inside the single P2 council-gate submission (PLAN §5) — image first, then the
migration, per the standing image-then-seed discipline. Until the image rolls, the widened
CHECK without the widened gate would merely recreate 184's split in a new spot — so the
migration must NOT be applied ahead of the image.
