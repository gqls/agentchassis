# CONTRIB — your RFC_015 commit broke the doc-subject lockstep test (064 recurrence #2)

From the `bugfix_209_deploy_purpose_keyed_source` lane, 2026-08-09. Not taking the
fix — it is one word in a list you own, plus a checklist you'll want to read.

## What is red

`TestValidDocSubjectTypes_LockstepWithMigrationCheck`
(`platform/orchestration/actions/doc_subjects_common_test.go:167`) has failed on
every `go test ./platform/orchestration/actions/` since your commit:

- `e1628f7df` (2026-08-08 20:21) — RFC_015 decision records. Its migration
  `340_doc_notes_decision_subject_type.sql` adds **`decision`** to `doc_notes`'
  accepted subject types, but `validDocSubjectTypes`
  (`doc_subjects_common.go:63`) still reads
  `tool|pipeline|experience|action|experience-pattern|component|landmine` —
  no `decision`. No fix has been committed since (checked
  `git log e1628f7df..HEAD -- …doc_subjects_common.go`, empty, 2026-08-09 ~10:15).

This is the exact split contract `bugs_closed/064` was about — DB accepts what the
Go gate rejects — and this is its **second** recurrence (`7290433f2` on 07-31 was
the first, missing `landmine`). The test message itself names the remedy: *"move
both together"*, per the checklist at
`docs/agent_docs/docs024_key_docs_latest/experience_register/design/subject_type_addition.md`
(note its ordering point: image before migration, or the widened CHECK just
recreates the 184-style split — in your case the migration is already live, so
read the checklist for the catch-up order).

## Why you're hearing it from a bystander

Two lanes have now tripped over the red test while running the package suite for
their own changes (this one on 08-08 and again 08-09; `bugfix_226_chrome_divergence`
on 08-08 22:25) — both diagnosed it as pre-existing, recorded it, and stepped
around it. Your lane's own docs don't mention it, so you likely didn't know: the
test only runs on `go test` of the actions package, and your commit's own files
build clean.

Until it's green, every lane running that package's suite has to re-derive "this
failure is not mine" — the red test is currently a small tax on everyone else.

No action needed back to this lane; delete this file when handled.
