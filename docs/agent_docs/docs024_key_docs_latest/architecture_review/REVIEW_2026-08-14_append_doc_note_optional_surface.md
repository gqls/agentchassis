# REVIEW — `append_doc_note`'s accumulated optional surface (RFC_022 budget, first of the three standing findings)

**Reviewed 2026-08-14 at the owner's direction** ("run note-writer's review"), under his
same-day rulings: budget N=10 on shared actions; sharing is estate design, so this
reviews the ACCUMULATED SURFACE as a whole, never the reuse. The action is carried by
**8 live agents** (component-template-fixer, council-gate, domain-research-classifier,
experience-planner, landmine-verifier, tool-acceptance-agent, tool-improver,
tool-recreation-handler) and declares **11 optional keys** — over the budget of 10.

**VERDICT: ACKNOWLEDGE AT 11. No trims.** Baseline recorded in
`optional_key_budget_acks.json` (and the cron's `ACKED_LEVELS` mirror); the counter goes
quiet on this action unless the surface grows PAST 11.

## Method

Read the whole implementation (`append_doc_note_action.go`, 190 lines, plus
`docResolveSubject` in `write_doc_plan_action.go:136-155`); censused live usage with
`config-key-audit --live-pairs` over the live fleet export; dated every key by the
file's full history. Each claim below carries its evidence.

## The surface, decomposed — 11 keys are 6 concepts, three of them doubled by the estate's standard literal/field duality

| concept | keys | read at | live-configured? |
|---|---|---|---|
| subject (what the note is about) | `subject_type`; `subject_key` / `subject_key_field` | `docResolveSubject` (gated on `validDocSubjectTypes`, the split-contract guard from bugs_open/064) | all three |
| body | `note_body_field` | `:97` (default `doc_note_body`; empty body is a refusal, `:100`) | yes |
| categories | `note_categories` / `note_categories_field` | `docCategoriesJSON` `:161/:164` — literal wins, field is the default path | literal yes; field via its default |
| site linkage | `note_site_id_field` | `:110` (default `input_data.site_id`) | yes |
| provenance | `note_source`, `created_by`, `source_item_id_field` | `:118/:119/:112` — each defaults to the calling agent or `input_data.item_id`, and each maps 1:1 onto a `doc_notes` column | first two yes; item-id via its default |
| the one AUTHORITY-bearing key | `note_body_suffix_field` | `:104-105` | yes (seed 365, landmine-verifier) |

Two keys are configured by no live step — `note_categories_field` and
`source_item_id_field` — and **neither is dead**: both are the defaulted half of a
duality, exercised through their defaults on every run that omits them
(`GetStringField(..., default)` at `:164` and `:112`). Removing either would break the
literal/field contract every other resolver in this action honours, to save nothing.

## The growth history — one addition in the action's life, and it was the reviewed one

Ten of the eleven keys arrived **at the action's birth** (`a18e6875d`, 2026-07-04), as
the write-path mirror of the `doc_notes` schema itself. The eleventh —
`note_body_suffix_field` — landed `1058b5366` (2026-08-10, bugs_open/223 phase 1),
went through a full council round (corr `495df717`, APPROVED), ships opt-in with the
unsafe default OFF per the 2026-08-02 ruling, is applied only AFTER the empty-body
refusal (a suffix can never make an empty verdict look written), and composes its text
mechanically in Go so the model whose verdict it qualifies cannot soften it. It has a
live consumer (seed 365). This is not an action accumulating drift; it is a schema
mirror plus one reviewed capability. The count crossing the budget is an artefact of
the schema having ~10 writable dimensions, not of unreviewed growth.

## What would reopen this

- Any twelfth optional key: the counter pages again (baseline 11), and under the ruled
  trigger the GROWTH itself is architecture-scope for a shared action.
- A new authority-bearing key (one that changes behaviour for existing callers, or is
  on by default): architecture-scope regardless of count, per the RFC_022 interim's
  three conditions.
- The census learning per-step usage it cannot see today: `--live-pairs` reports
  distinct (action, key) pairs, not per-carrier counts — good enough here (the
  question was "is anything dead", answer no), noted for the next reviewer.

## Baseline mechanics applied with this review

`optional_key_budget_acks.json`: `append_doc_note` at count 11, pointing here.
`check.py`'s `ACKED_LEVELS` mirror updated + configmap re-applied; parity test
`TestBudgetCronAckedLevelsMatchTheAcksFile` pins the pair. The daily check will keep
paging on `analyse_repo_local` (12) and `diagnose_prepare_fix_commit` (11) — **that is
correct behaviour, not noise**: those two owe their own reviews (owner: they wait for
natural contact), and the red is the system telling the truth until then.
