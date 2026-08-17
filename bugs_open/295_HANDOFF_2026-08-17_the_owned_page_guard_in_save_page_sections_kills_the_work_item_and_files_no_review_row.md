# 295 — the owned-page guard in `save_page_sections` kills the work item and files NO review row, while the estate's two other guards on the same predicate both file one

**Filed 2026-08-17** by the `vigilant_designer_offer_analysis` lane, found by reading where the
offer analyser's first drained population actually ended up (NOTES 2026-08-16/17). **OPEN, unowned.**

**One line:** three code paths guard an `owned` page with the same predicate; two of them record
the refusal as an `owned_page_review` item that a human can act on, and the third — the only one
on `page-build-handler`'s route — returns a hard error, so the work item dies `failed` and the
refusal leaves no trace once the orchestration ages out at ~24h.

---

## The evidence

**Two live work items died this way** (offer-analysis, webdesign.co.uk, 2026-08-15). Quoted from
`orchestration_states.collected_data->>'__step_error'` on the orchestrations the items name in
`result.completed_by_orchestration_id`:

- `bbee148d-03e5-46a1-8358-e395e7af39cd` (item `1279d84e`, `content_rewrite`, page `about`):
  `step save_sections failed: failed to execute action save_page_sections: page about is
  rebuild_policy=owned (tool/widget-owned): a generic section save would clobber it. Use
  apply_section_edit for targeted edits or the tool pipeline for rebuilds. Refusing to overwrite.`
- `1bdbbedc-bf29-490f-8b81-b41d21c61aec` (item `9a29ec67`, `tone_shift`, page `learn-index`):
  same message, page `learn-index`.

**The guard is right to refuse.** This is the TL-001 clobber guard (`bugs_open/208`), unified
through `pageIsOwnedForGuard`. Nothing here argues it should save the page. The defect is that it
refuses *without leaving a record*, and that its two siblings do not.

### Three call sites, one predicate, two behaviours

| call site | on `owned` | files `owned_page_review`? |
|---|---|---|
| `owned_page_guard.go:294` (selection census, `get_pages_to_build`) | excludes the page | **YES** |
| `multipage_actions.go:62-84` (`assemble_page`) | returns `skipped:true` + `skip_reason`, **no error** | **YES** (`:75`) |
| `save_page_sections_action.go:186-196` | `return nil, fmt.Errorf(...)` — **hard error** | **NO** |

`emitOwnedPageReviewItem` (`owned_page_guard.go:167-200`) is reviewed, deduped machinery: one
deterministic `item_key` per page, `ON CONFLICT DO NOTHING`, errors swallowed so reporting can
never break the guard. **Its `spec.fix` already names the correct route in prose** — *"edit via
`apply_section_edit`/section-editor, or change `pages.rebuild_policy`…"*. The save path is simply
not wired to it.

### Measured in the live DB, 2026-08-17

- **`owned_page_review` rows by `refused_by`:** 34 `(none)` at `needs_human_review` (reconcile's
  own path, 07-17→08-15) · 1 `get_pages_to_build` (cancelled, the 208 canary) · **0 from
  `assemble_page`, 0 from `save_page_sections`.** The 34 are the **positive control**: the item
  type reaches its terminal human-review state and stays there, so a zero from this path is a
  real absence, not a dead mechanism.
- **172 of 704 pages (24%) are `rebuild_policy='owned'`** — the exposed surface.
- Failed items sitting on owned pages come from several filers, not just this lane:
  `design-audit` (`needs_content_page` ×2, `content_rewrite` ×1), the checker layer with no
  `audit_source` (`literal_markdown` ×8 on 08-14, `placeholder_contact` ×2, `content_rewrite` ×3),
  offer-analysis (×2, quoted above).
  ⚠ **`[UNMEASURED]` for all but the two quoted:** their orchestrations are past the ~24h
  retention, so "failed item on an owned page" is not proof it was this guard. Do not repeat the
  aggregate as if it were.

### Why the normal flow does not catch it

`page-build-handler`'s live workflow has **no `assemble_page` step** (read from
`agent_definitions`, 2026-08-17):
`… plan_sections → check_has_ready_sections → spawn_content_writer → call_content_writer →
check_content_produced → validate_content → save_sections → update_status → …`

So the sibling that *does* file the review item is never on this route, and
`SavePageSectionsAction`'s upstream-skip honour path (keyed to `ownedPageSkipReasonPrefix` by
commit `6a9d85777`) can never fire either — there is no upstream skip to honour. The direct
ownership check was written as a backstop; on this route it is the **only** guard, and it is the
one that reports nothing.

**The correct destination already exists in the same workflow.** `validate_content`'s
`error_step` is `mark_needs_review`; `save_sections`'s `error_step` is `mark_item_failed`. A
content-validation failure gets a human; an owned-page refusal gets a silent `failed`.

## Fix candidates, ordered by what closes the door

1. **(Preferred, code, needs a roll) Call `emitOwnedPageReviewItem` in the save-path guard**,
   matching its two siblings, then keep returning the error. The item still fails — honest, the
   rewrite genuinely did not happen — but the refusal becomes a deduped, durable,
   human-routable row naming `apply_section_edit`. Closes it for **every** filer (design-audit,
   the checker layer, any future auditor), not just for offer-analysis, and needs no per-agent
   config. ~6 lines at the existing call-site pattern; `siteID` is already in scope for the
   error path via the same `CollectedData` lookup the sibling uses.
2. **(Config-only, live immediately, WRONG SCOPE) Point `save_sections.error_step` at
   `mark_needs_review`.** No roll, but it sends *every* save failure to human review, including
   transient ones — it cannot tell an ownership refusal from a database blip. Records the symptom
   at the cost of a mislabelled queue. Use only as a stopgap, and say so.
3. **(Upstream, larger) Route content findings on owned pages to `section_edit` at triage.**
   `section_edit` **completes** on owned pages — 18 of them — so the working route exists; nothing
   puts a content finding onto it. This is the real repair rather than the honest refusal, but it
   is a routing-policy change with its own design, and 1 should land first regardless.

## How to verify a fix

Positive and negative control in one run, because a zero here is otherwise ambiguous:
- Dispatch a `content_rewrite` at a **known owned** page → expect a NEW `owned_page_review` row
  with `refused_by='save_page_sections'` and the item terminal.
- Dispatch one at a **known generic** page → expect NO review row and a real save.
Running only the first cannot distinguish "the fix works" from "the emit fires unconditionally".

## Verification basis (owner ruling 2026-07-31)

**Not** put through the `090` diagnosis loop; the substitute, stated plainly: all three call
sites and both governing commits (`cb7b4d759`, `6a9d85777`) read first-hand at file:line; the
absent `assemble_page` step read from the live `agent_definitions` row, not from a repo seed; and
the data claim carries its own positive control (34 rows prove the item type reaches
`needs_human_review`, so 0 from this path is an absence rather than a dead mechanism). The two
failures are quoted verbatim from the orchestrations the items themselves name.

## Relates to

`bugs_open/208` (the guard this extends — its lane last touched it 08-08; the selection and
assemble paths it fixed are the two that DO file the review row) · `bugs_open/210` (a page-build
failure stamped deployed — same family: the build path's failure reporting) ·
`features_open/030` + BIZ-032 (the offer analyser, whose findings were the two casualties) ·
`bugs_open/115` (findings that terminate nowhere — this is the same harm reached by a different
route).
