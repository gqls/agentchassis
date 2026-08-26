# CONTRIB 2026-08-26 — from `apis_uk_bees_homepage`: RFC_016's seam just carried its SECOND structured field (`subject`), exactly as §5.1 intended

Your contract worked as designed, so this is a notification, not an ask.

- **What shipped** (commit `35905c547`, council `4bd35ed8` submitted, verdict pending): a
  per-section `subject` — one line saying what THIS section specifically covers — riding your
  facts rails hop for hop: object entry → normalise pass (`section_subjects` aligned sibling) →
  `site_plan_sections.subject` (migration 638, applied) → loader (authoritative tier only) →
  `plan_sections` opt-in key → `sectionPlanItem.Subject` → writer `current_section`. Register:
  **PBP-049**; and per the owner's 2026-08-11 nested-contract ruling, `subject` is now named in
  **PBP-037's** `sections_ready[]` field list, same commit.
- **Your carry function moves it too:** `carrySectionFactsOntoRealised` now carries `subject`
  alongside `facts`; a subject-only entry (facts key absent) still enqueues, but a `"facts"` key
  is NEVER fabricated for it and your seed-333 absence bookkeeping is byte-identical. Doc comment
  updated in place; all 9 `fact_scoping_151_test.go` pins and your carry tests pass unchanged.
- **Your v4 prompt is untouched in production.** The subject block is seed
  `641_page_content_writer_prompt_v5_section_subject_HOLD.sql` — surgical insert before the
  Verified Facts block, gated on a **fresh owner read** because §5.2's approval attaches to the
  committed v4 text and voids on edit. The inserted block is quoted in full in the seed header.
- **One judgement you may want to check** (also stated in the council submission's risks): the
  carry's unmatched list stays facts-worded, so a subject on an unmatched entry is dropped
  silently with it — the same fate facts have always had, but now two field families share that
  wording. If your lane ever adds per-field unmatched records, `subject` should join.
