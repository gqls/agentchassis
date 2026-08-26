# PLAN 2026-08-26 — bugs_open/399: make the CTA label and its destination answer one question

## The defect

The framework computes each CTA's real destination, writes that destination's TITLE into the same
jsonb object as the copy, hands the title to the content writer *specifically so it can write copy
for the actual destination*, and never checks whether it did.

## What I am building

One shared predicate, asked once before the row is persisted, recorded when it disagrees.

1. `datahelpers.JudgeCTALabel` — "does this copy name the page its destination is?", built on the
   existing `BestLabelMatchForPage` + `LoadCTALabelUniverse`.
2. `check_misdirected_cta`'s `ctaClassifyAnchor` becomes a thin adaptor over it. Its existing tests
   pass unchanged; that is the proof the extraction changed nothing.
3. `actions/cta_label_audit.go` — a write-time pass in `SavePageSectionsAction`'s guard chain,
   opt-in (`audit_cta_label_agreement`, unsafe default OFF), recording `CTA_LABEL_MISMATCH` to
   `agent_error_log`.
4. Migration 643 arms it on all six live `save_page_sections` steps.

## Decisions, and the reasons

### D1 — the remedy is a RECORD, not a refusal and not a repair

Both alternatives were designed and rejected on measurement.

- **Repair is nearly unreachable.** `[MEASURED 2026-08-26]` of 186 mismatched pairs, the copy names
  exactly one other page in **13**, two or more in **78** (RFC_047: refuse, never guess), and no
  page at all in **95**. A repoint reaches 7% and inherits `bugs_open/248`'s clobber — a repair
  turned a correct `/contact.html` into a wrong link on 2026-08-24.
- **Refusal has nowhere to go.** At 14.6% it fails ~1 CTA write in 7 (~29 sections/day at the
  2026-08-24/25 rate), nothing can auto-satisfy it, and on a page that re-authors itself several
  times a day one cosmetic mismatch becomes an indefinitely withheld refresh.

### D2 — CORRECTION to the bug's own fix candidate 1

399 proposes comparing the label to `_target_title` and, on mismatch, *"regenerate the label from
the title"*. **Both halves are wrong and the second is harmful.**

- The comparison would be a THIRD definition of "misdirected" beside the detector's and the
  writers' — the re-drift RFC_047 §9 explicitly rejects — and it is the shape
  `bugfix_203/CALIBRATION_2026-08-11` measured brittle (nine correct CTAs flipped over a hyphen in
  "Break-Even").
- Regenerating the label from the title is the same operation `stampCTADestinationGuidance` already
  performs, by force. It converts a mismatch into a **lock**, moving the row out of the ~60
  label-less bucket a ranking fix reaches into the ~20 label-locked bucket only an LLM copy pass
  clears (`bugs_open/391`).

### D3 — the seam is `save_page_sections`, not the render

`RenderComponentAction` is where the fresh label first meets the resolved destination and it is the
**wrong** seam: `RerenderPageSectionsAction` bypasses it entirely
(`rerender_page_sections_action.go:662` calls `RenderTemplate` directly), so a gate there is blind
to the repair loop — the loop actually minting the churn. Both writers converge on
`save_page_sections`, verified in the live `agent_definitions` on 2026-08-26.

### D4 — arm all six save steps, not two

A recursive census found `save_page_sections` on **six** live agents, four inside loop
sub_workflows. For an instrument this matters more than for a guard: a guard armed on half its
writers is visibly partial, an instrument armed on half its writers reports a **rate** that reads
fleet-wide and is silently biased. Migration 643 asserts the census.

## Stated blind spots

- **Blind to `bugs_open/391`'s label-locked class by construction.** When the framework wrote both
  sides they agree, and the button is still wrong (16/17 of the password-entropy family). Pinned by
  a test whose comment says so.
- **Population is the six `ctaFieldNames` components only** — a `_target_title` exists only where
  `setCTAField` ran.
- **Site chrome is out of scope** — `site_components` carries no cta/target_title keys.

## Sequencing against live lanes

`bugfix_389_cta_relevance` (bug 391) is active. Its **step 4** re-resolves ~60 label-less fields,
writing new URLs under old generic copy — contradictions **by construction**. The audit must read
that window as expected, not as damage. Told them directly.
