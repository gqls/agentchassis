# HANDOFF 2026-08-06 — bug 122, contrast / ink slots. START HERE.

Lane opened today. **No code written yet** — a plan is at the council and a diagnosis
is at the loop. Read `PLAN_2026-08-06_contrast_ink_slots.md` for the full working;
this file is the "what to do next" list.

## The one-paragraph state

`bugs_open/122` is nine days old and **two of its three findings are fixed, with its
first fix candidate already shipped** — I re-measured 15 homepages in a browser rather
than trusting the file, and corrected it in place. What survives is a different class:
109 firm contrast failures across 12 sites, in three sub-shapes. A plan covering two of
them is with the council; the third is with the diagnosis loop because I would have had
to guess at its cause. The cheapest useful thing in the whole file is now **one
`scheduled_tasks` row** — the render audit that files contrast defects as work items is
fully built and live in v1.0.1257, and nothing ever dispatches it.

## Two things are in flight. Check these FIRST.

| what | correlation | check with |
|---|---|---|
| **council gate** — the fix plan (sub-shapes A + C) | `c4d9c841-3658-4742-85b5-961e062ecad2` | see below |
| **090 diagnosis** — sub-shape B, the six invisible headings | run `5853ee07-a49c-4571-8ea0-3eb660e43dfd` · intake `2f3d2cc0-197c-46ff-aac7-bd5e77ea782e` | see below |

```sql
-- council verdict
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='c4d9c841-3658-4742-85b5-961e062ecad2' AND kind='council_report'
ORDER BY created_at;
-- still running?
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = 'c4d9c841-3658-4742-85b5-961e062ecad2';
-- the readable note
SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;

-- diagnosis verdict
SELECT created_at, kind, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='5853ee07-a49c-4571-8ea0-3eb660e43dfd' ORDER BY created_at;
SELECT status FROM site_work_items WHERE item_type='needs_diagnosis'
  AND spec->>'dispatch_correlation_id'='5853ee07-a49c-4571-8ea0-3eb660e43dfd';
```

Last seen (2026-08-06, ~11:15 BST): council at `review_editquality / EXECUTING_STEP`;
diagnosis `diagnosing`, two bundles written. **Those DB timestamps are UTC and this
machine is BST** — an artifact stamped an hour "before" you fired the trigger is
normal, not a failure.

## Next actions, in order

1. **Read the council verdict.** APPROVED → write the code as sketched in the
   submission (the sketches are real, not gestures) and commit with
   `Council-Reviewed: c4d9c841-3658-4742-85b5-961e062ecad2`. REVISE → the objections
   come back with the reviewers' own read-only checks already answered; resubmit with
   `RESUBMIT_CORR=c4d9c841-…` so the trail accumulates. **The risks block names five
   things I expect a seat to push on** — the scheme widening is the big one, so read
   that first and do not be surprised by it.
   *If you commit code before the verdict lands, use `Council-Submitted:` with the
   same correlation — never `Council-Reviewed:` on an unread verdict.*
2. **Read the diagnosis verdict for sub-shape B.** If CONFIRMED, it names the cause of
   ai-agent-orchestration.com's six invisible headings and that becomes its own fix
   (probably its own bug file — it is currently recorded in no bug file but this
   lane's docs and 122's appended section). If REFUTED, that is a success and the
   cheapest possible place to have been wrong: record it as a visible correction in
   `NOTES` and in 122, naming what caught it.
3. **The `scheduled_tasks` row for `render-audit-agent`** — live config, no build
   needed, and the highest value-per-effort item here. **But bank a baseline first**:
   findings dedup on `contrast_failure:<page-path>#<selector>` and the next audit is
   the de-facto verifier, so a pre-fix run is what makes the fix measurable.
   `python3 scripts/render_audit.py <the 12 failing urls> > baseline_$(date +%F).txt`
   Today's run is at
   `/tmp/claude-1000/-home-ant-projects-agentchassis/b505341a-…/scratchpad/audit_fleet_2026-08-06.txt`
   — **copy it into this directory before relying on it**, the scratchpad is
   session-scoped.
4. **When you ship the seam, register it in the SAME commit** —
   `docs026_concept_register/register/visualisation-and-charts.md` (VIZ-010/012/013 are
   the neighbours) plus the index count. This is condition (2) of the ordering
   exemption and it still stands; it is not optional and it is not "later". The
   pre-commit pattern check already flagged this directory as unknown to the register.
5. **Add a LANDMINE entry** for the two-grounds constraint (see below) and run
   `./scripts/landmines-sync.py --apply`.

## What NOT to do, each for a measured reason

- **Do not add the slot to `darkSchemeDerivations`.** It compiles, logs success, and
  changes nothing: a palette slot reaches the stylesheet only through
  `{{palette "X" "literal"}}` in a layout template. `accent_text` has been derived
  since 2026-07-27 and is declared by **0 of 18** layouts, so it has never reached one
  stylesheet. That is the recorded LANDMINE, and it is why the fix emits from the
  renderer instead.
- **Do not grade any of this on a stylesheet or a palette row.** A stylesheet cannot
  resolve the cascade; a palette cannot see a literal that is in no palette. Both
  produced wrong answers in this bug's own history — its 07-28 correction records the
  regex audit "comparing the background against itself".
- **Do not repoint a palette to fix sub-shape A.** The dartsonline round proved there
  is no value satisfying both the fill and the ink role, so it trades one failure for
  another. The whole point of the plan is a *second* variable.
- **Do not touch vonc.com's Gauntlet buttons** (23 failures, same shape). The
  `gauntlet_dead_cta` lane owns that surface and 122 says to coordinate.
- **Do not read a falling failure count as repair.** It is content-dependent — 122's
  dartsonline round found the same defect reporting 1 or 2 depending on which cards a
  page happened to render.
- **Do not conclude "never ran" from an empty `orchestration_states`.** Terminal rows
  are reaped at ~24h. Ask `scheduled_tasks`, which has no reaper.

## The design constraint most likely to be lost

`--color-primary-ink` must clear AA against **background AND surface**, not one ground.
dartsonline places the same ink on both — the eyebrow on `background` (1.04) and the
card link on the derived `card_bg` (1.11). My first sketch took a single ground and
would have fixed the eyebrow while leaving the card link failing, **which would have
looked like a working fix on whichever page I happened to test**. `legibleInkFor` takes
a slice of grounds for this reason. If a future edit simplifies it to one ground, that
is a silent half-regression.

## Facts worth not re-deriving (all measured 2026-08-06)

- `header-theme-chrome` is var-driven; **0 of 19** stored header rows hold a hardcoded
  white CTA ink. 122's finding 1 is done.
- **17 of 18** layouts use `color: var(--color-primary)` as an ink (`social-lobby` 11×,
  `affiliate-hub` 9×, `magazine-grid` 8×). Census SQL is in the RUNBOOK; the `[^-]`
  prefix is load-bearing or `background-color:` inflates it.
- `primary_text`/`cta_text`/`header_text`/`footer_text`: declared by **18 of 18**
  layouts. `accent_text`: **0**. `card_bg`: 18. `surface_alt`: 3. `icon_chip_bg`: 0.
- `--color-primary-ink` is unused on all five surfaces (components, layouts, snippets,
  page_components, site_components) — checked before choosing the name.
- `write_render_audit_findings` is live in v1.0.1257 (11 hits, invented control 0);
  `render-audit-agent` steps are `site → audit → write_findings → complete`; **no**
  scheduled task dispatches it; **4** `contrast_failure` items ever, all relojistas,
  all 2026-08-04.
- Components to repoint: `image-hover-card-grid` (1 page, 1 site) and `tool-list`
  (6 pages, 4 sites). Both active and unforked.
- Sites now CLEAN that 122 lists as failing: relojistas.com, vetcomparison.uk. Also
  clean: fundamentallyai.com, leopardessconsulting.co.uk.

## Where things are

- `PLAN_2026-08-06_contrast_ink_slots.md` — the mechanism, three sub-shapes, four fix
  candidates with the two rejected ones and why.
- `RUNBOOK_contrast_ink_slots.md` — every command with its gotcha. **Read the
  browser-UA and schema-first ones before touching the DB or curl.**
- `NOTES_contrast_ink_slots.md` — the four missteps, the blind census, and the
  correction to my own first framing.
- `README_where_we_are.md` — the owner's plain-prose account.
- `bugs_open/122_…` — corrected in place, dated, append-only (146 lines added, 0
  removed).
- Council submission JSON: scratchpad,
  `submission_122_ink_slots.json`. **Copy it into this directory if you need to
  resubmit** — the scratchpad does not survive the session.
- `WRONG_CALLS.md` — today's four entries.

## No SUMMARY yet, deliberately

Nothing has shipped and both verdicts are pending, so the five headings would produce
"we measured and planned" — which is what NOTES and this file are for. Write one when
the first fix is live and proven, not on a clock.
