# 293 — the whole-page shrink floor measures CSS and SCRIPT as "text", so it cannot see prose deletion — and its one refusal in the archive was a repair

**Filed 2026-08-17** by the `bugfix_285_shared_template_write` lane, which fixed the SIBLING call
site and is deliberately not touching this one. **This is the residual half of a corrected axis,
not a new discovery**: the mechanism is measured, the fix is known, and what is missing is
EVIDENCE FOR THIS PATH — see "Why this was not fixed with its sibling".

**First-hand verification substituted for a 090 run (2026-07-31 ruling):** every number below is
a query over `page_component_history` / live rows, run 2026-08-17 and quoted with its predicate;
the code claims are lines read at HEAD. No downstream-consequence claims.

## The defect

`save_sections_shrink_guard.go` refuses a whole-page save whose per-slot text falls below
`section_shrink_floor` (default 0.5). It measures with
`shrinkGuardTagStripper = regexp.MustCompile(`<[^>]*>`)` (:58) — which strips TAGS but not what is
INSIDE `<style>` and `<script>`. So CSS declarations and JavaScript source count as "text".

Consequences, both measured on the real pair that `bugs_closed/285` produced
(`page_component_history ab400131-2a41-434b-bd95-d44c9f064a32` vs the poison row):

| axis | existing | incoming | kept | verdict at floor 0.5 |
|---|---|---|---|---|
| tag-stripped (this guard) | 3,236 | 8,491 | **262%** | ALLOWS — a stylesheet replaced the article |
| visible text (`datahelpers.VisibleTextFromHTML`) | 2,143 | 16 | **0.7%** | refuses |

And the mirror, which is the part that makes this urgent rather than academic — across all 117
archived overwrite pairs, run through the real implementation, the tag-stripped axis refuses
**exactly one** write and it is the **REPAIR** (2026-08-15 18:18Z, seed 431, putting the article
back): 38% kept on that axis, 16 → 2,143 on the visible one. The visible axis refuses **three**,
all real hollowings — `idea.uk/tool-ab-test-calculator` 684 → 0 visible (2026-08-11, while
tag-stripped GREW 10,399 → 12,929), this bug's article 2,143 → 16, and
`webdesign.co.uk/tool-ab-test-calculator` 684 → 0 (2026-08-15, `bugs_open/286`'s hollow fork).
The two axes agree on **zero** pairs. A guard whose only firing in eight days of history would
have blocked a repair is a guard that gets switched off.

**PRIORITY — this is the HIGHER-VOLUME path, not a lower-priority follow-up** (council
`3279156b`, bug_historian seat, medium: this plan is an instance of 016b §9's "one call site of a
shared judgement gets the rigorous fix; the sibling stays heuristic"). The delete-arm count below
is the volume argument: 3,603 rebuild writes against 281 edit writes in the same 8 days. The
sibling is fixed because it could be CALIBRATED, not because it mattered more.

## Why this was not fixed with its sibling

The section-editor call site (`single_slot_floors.go`) was corrected to the visible-text axis and
is council-submitted (`3279156b-d2ba-41f0-9115-aa2275bfb27e`), because the archive can PAIR that
path's writes: it UPDATEs a row, and migration 357's trigger banks the prior. This path cannot be
paired the same way — its writes are DELETE+INSERT rebuilds, archived as delete rows with no
successor:

```sql
SELECT count(*) FILTER (WHERE op='overwrite') AS pairable,   -- 281
       count(*) FILTER (WHERE op='delete')    AS unpairable, -- 3,603
       min(created_at)::date AS archive_since               -- 2026-08-09
FROM page_component_history WHERE op IS NOT NULL;
```

Changing an axis fleet-wide on evidence that does not cover the path is how a guard starts
refusing good work — so this is filed rather than shipped.

## What the fix would be (and its measured cost)

One line: feed `visibleTextLength` (`section_visible_text.go`, live at the sibling call site) to
`strippedExistingBySlot` / `strippedIncomingBySlot` instead of `shrinkGuardTagStripper`. The pure
decision, `minShrinkGuardChars` (500) and the config opt-outs stay as they are.

Known cost, measured on what CAN be paired: applying the 500-char minimum to visible text takes
**31 of 117 pairs (26%)** out of the text floor's scope — slots that are mostly CSS/JS. The
tag-stripped axis refused 0 of those 31, and on a CSS-dominated slot its ratio is dominated by the
stylesheet, so its "protection" there cannot see the failure mode anyway; structural collapse on
those slots is still covered by the class-attribute floor (`bugs_open/253`, unchanged, its own
minimum of 10 class attributes).

## How to close

1. Get evidence for the DELETE+INSERT path. Two candidates, neither run:
   (a) reconstruct pairs from `page_component_history` delete rows joined to the INSERT that
   replaced them within the same rebuild (same `page_id` + `slot_name`, next `created_at`) — the
   join is plausible but unproven, and a wrong join manufactures false pairs, so prove it on a
   rebuild whose outcome you already know;
   (b) shadow-mode the corrected axis: compute both, refuse on neither, log the disagreement, and
   read a week of it. Costs a roll, cannot break anything, and produces exactly the pairs (a)
   guesses at.
2. Then change the axis, or record why not.
3. Either way, re-run the sibling's calibration (117 pairs, both directions) — it is in
   `section_visible_text.go`'s header and it is the thing that must not rot.

## Relations

- `bugs_closed/285` — the incident both halves come from; its delivery landed on an arbitrary
  ported page and the served article was empty for ~23 h.
- `bugs_open/178` (the text floor), `bugs_open/253` (the class floor, and the council round that
  wired both into the section editor), `bugs_open/286` (the rebuild lane's ask for a text-content
  floor on tool slots — the same missing axis, found from the other end).
- 016b §9: "a census scoped to ONE writer measures that writer, not the surface" — this is the
  same shape one level down: an AXIS scoped to one path measures that path.

---

## CONTRIBUTION 2026-08-17 — step 1 is DONE: the evidence exists, and it says something the plan above did not expect

By the `bugfix_293_whole_page_shrink_axis` lane. Full working, controls and commands:
`docs/agent_docs/docs024_key_docs_latest/bugfix_293_whole_page_shrink_axis/` (NOTES + RUNBOOK), and
the harness is committed at `platform/orchestration/actions/shrink_axis_calibration_test.go` —
so the calibration `section_visible_text.go` tells you to re-run is now a command.

### The join was provable, and neither of the two candidates above is quite it

Candidate (a) proposed joining a delete row to "the INSERT that replaced it within the same rebuild"
and called the join plausible but unproven. **The replacement is not an unarchived INSERT — it is the
LIVE `page_components` row, and `page_components.created_at` is independent proof it belongs to that
rebuild.** The rule is uniform for both ops (357's triggers bank `OLD.rendered_html` on UPDATE *and*
DELETE): an archive row is the content live until that moment, so the write after it is
`row.rendered_html → (next archive row for the same page+slot, else the live row)`.

- 1,254 (page, slot) groups have an archived delete; **1,123 were re-inserted within 60 s, 1,109
  within 5 s**; and — the disconfirming control — **ZERO live rows are older than their own last
  delete.** A wrong join does not produce that zero.
- **Positive control against a number this lane did not produce:** the same rule on `op='overwrite'`
  reproduces this file's three known refusals with identical figures (684→0, 2,143→16, 684→0) and the
  tag-stripped repair at 38.1%. It also finds a **fourth** hollowing not in the table above —
  `idea.uk/index/tool-list`, 5,643→0 tag-stripped, 1,118→0 visible, 2026-08-10.
- ⚠ **Do not join on `(page_id, slot_name)` alone.** Slot names repeat on 14 pages, so it is a
  cartesian product; on the first run it manufactured a refusal
  (`leopardessconsulting.co.uk/technical-architecture/generic-text-block` 2,831→15) that does not
  exist. 1,079 pairs survive the exact join.

### The axis swap alone is SAFE but nearly INERT on this path — the 500-char minimum is the other half

Every archived pair is a write the live guard ALLOWED (the guard shipped 2026-08-02, the archive
starts 2026-08-09), which makes this population right for false refusals and useless for true catches.

| | tag-stripped (shipped) | visible text |
|---|---|---|
| in scope, of 1,079 exact rebuild pairs | 1,062 | **492** |
| refuses any of them | 0 | **0** |

So "it would have caught X" is **not available** on this path, and the 26% scope loss the plan above
costed is **52% here**. The change has to be argued on mechanism, and measured prospectively:
construct this bug's failure (delete every word of prose, keep the wrapper markup and the
`<style>`/`<script>` content) on all 1,079 real sections and ask each axis —

| | judged | REFUSES the prose wipe | ALLOWS it |
|---|---|---|---|
| tag-stripped (shipped) | 1,060 | 336 | **724 (68%)** |
| visible text | 492 | 492 | **0** |

Confirmed by a second, independent instrument (crude regex, not the parser): for **728 of 1,062**
sections, over half of what the guard counts as "text" is style/script content — 85–89% on some
(`relojistas.com/hero`, `finetuning.uk/hero-tool`).

**And the minimum is why the visible column only judges 492.** `minShrinkGuardChars = 500` was chosen
against CSS-inflated lengths. Swept over both populations, the number of writes **the guard would
actually have judged** and refused stays at exactly **ONE** at every step from 500 down to 50, while
scope rises 492 → 959 (min 200) → 1,046 (min 120):

| minimum | in scope (of 1,079) | refuses real writes | …the guard would have judged |
|---|---|---|---|
| 500 | 492 | 0 | 0 |
| 200 | 959 | 0 | **1** |
| 120 | 1,046 | 0 | **1** |

That one is real and was read by hand: `robot-hands.com/about/differentiators`, 2026-08-11 13:12:30,
two rebuilds 96 s apart, 3,724 → 1,554 visible chars — a **legitimate tightening rewrite** the
visible axis would have refused. One false refusal, fleet-wide, in eight days, against 724 prose
wipes the shipped axis waves through; `section_shrink_floor` is the existing escape hatch for it.
(Every other apparent refusal in the weak-join population spans a gap of 1,700 s to 93 hours — a slot
absent that long is a DROP, which the guard declines to judge.)

### Two things found on the way that widen this bug

1. **The axis is chosen in THREE places, and the oldest is the blindest.** The inline page-total
   content-regression guard (`save_page_sections_action.go:~549`, refuses below a quarter of the
   page's text) tag-strips in SQL. `[APPROXIMATE — paired slots only]`: of 366 pages it would allow a
   **whole-page** prose wipe on **337**, and refuses **0** real rebuilds on the visible axis. There is
   a fourth copy at `:393`, a diagnostic log line that advertises itself as "the stripped-text total
   the regression guard will compute" — which becomes false the moment the guard's axis changes.
2. **Both sides of this guard's comparison are keyed on slot name alone**, and slot names legitimately
   repeat (`LANDMINES.md`: 11 of 17 duplicate groups are legitimate — `generic-text-block` used 2–3×
   with differing content). `strippedIncomingBySlot` does `m[ComponentName] = len` (last write wins);
   the existing side does `existing[slot] = len` (last row scanned wins). **On those 14 pages the
   guard compares an arbitrary instance against an arbitrary instance, and which one depends on DB row
   order and slice order.** The sibling class floor (`save_sections_component_floor.go:180`) has the
   identical keying. Fixable without any uniqueness assumption by summing per slot name.

Step 2 (change the axis, or record why not) is this lane's next move; option (b), shadow-mode, is no
longer needed — it would cost a roll and a week to produce pairs that already exist.
