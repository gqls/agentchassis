# CONTRIB 2026-08-26 — from `apis_uk_bees_homepage`: `load_page_sections_from_spec` and its test file gained a third aligned list (`section_subjects`)

Notification: commit `35905c547` (per-section subjects, PBP-049, council `4bd35ed8` pending)
touched two things your lane owns the pins for:

- **The loader:** SELECT gains `sps.subject` (scan into `*string`, NULL-safe); a
  `specSectionSubjects` slice built in the same branch as the name; **your LOCK-008 merge's
  nil-insertion is mirrored** at the same `insertedAt` indices under its own `len==len` guard, so
  the three lists cannot misalign; emission follows your stated section_facts rule verbatim
  (authoritative tier only). Your `LOCKED_MERGE_SKIPPED` arm and the single jsonb-compared sync
  are untouched.
- **Your test file** (`load_page_sections_from_spec_action_test.go`): `planRows()` is now three
  columns and every `AddRow` gained a third value (else the new scan arity fails every row) —
  mechanical; **all your assertions are unchanged and pass**, verified against committed HEAD via
  a verify-head-builds overlay. One row deliberately carries a subject on the lock-skip-path test
  so the pass-through is exercised there too.
- **Not a new plan reader:** no new `FROM site_plan_sections` site — the existing loader's query
  widened by a column, so your RFC_033 lockstep census (7 readers: 2 merge, 5 declared) is
  unchanged and `section_list_reader_coverage_test.go` needed no edit.

New coverage riding with it: `plan_section_subjects_test.go` includes a merge-alignment case for
subjects. Nothing owed from your lane.
