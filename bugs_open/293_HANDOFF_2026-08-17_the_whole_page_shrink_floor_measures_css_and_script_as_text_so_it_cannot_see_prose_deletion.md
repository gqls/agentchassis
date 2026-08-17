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

---

## FIX COMMITTED 2026-08-17 — steps 2 and 3 done. STAYS OPEN until it rolls.

By the `bugfix_293_whole_page_shrink_axis` lane. `6aae23e62` (axis, minimum, page-total extraction,
keying, tests) · `9cd887ddf` (the class floor's keying) · `e42d57adf` (council round 1 revisions) ·
`3a69ea16c` (a pattern-check allow-list entry my own extraction un-exempted). Council
`823679dc-43d5-4f93-8b2d-746c41250290`, round 1 REVISE with all five objections acted on, round 2
resubmitted on the same correlation.

**Open, not closed, and the bar is the reason:** this is Go, so it is inert until an image is built and
rolled. `[MEASURED 2026-08-17]` the running chassis `v1.0.1305` is stamp `6a782274b` (08-16) —
**and `4b32f174c`, the sibling axis this file's own header describes as shipped, is NOT in it
either.** Both halves go live on the same roll. `IMAGE_TAG` is already at `v1.0.1307`.

### What shipped, against this file's "What the fix would be"

| this file said | what shipped, and why it differs |
|---|---|
| "One line: feed `visibleTextLength` … instead of `shrinkGuardTagStripper`" | done, on **three** floors not one — the per-slot whole-page guard, the section editor, and the page-total guard extracted out of `save_page_sections_action.go` |
| "`minShrinkGuardChars` (500) … stay as they are" | **changed.** 500 on visible text leaves 587 of 1,079 slots unjudged, so `evaluateSectionShrink` takes the minimum as a parameter and both floors pass `minShrinkGuardVisibleChars = 200`: scope 492 → 959, guard-judged refusals pinned at ONE from 500 down to 50. 200 not 120 because 200 is the deepest step the section editor's own 263 pairs also cover (scope 153 → 204, same 4 refusals). `minShrinkGuardChars` keeps its value — its remaining consumer (`load_current_section_content_action.go:262`) is a pairing heuristic that refuses nothing |
| "the config opt-outs stay as they are" | the two existing ones did; the page-total floor **gained** `page_total_text_floor` (default 0.25 = today's behaviour, 0 disables). It had no escape hatch at all — the only way past it was rolling a binary |
| — | **`shrink_axis_coverage_test.go`**: a caller of `evaluateSectionShrink` that does not measure with `visibleTextLength` fails the build; any `enforce*Floor` with no caller fails the build. This is what replaces the "unaffected by construction" claim |
| — | **keying**: both sides of both floors now SUM per slot name. On the class floor MUT-293-G shows last-wins is a live **false refusal**, not a blind spot: two instances holding 60 class attributes, rebuilt to keep 55, is refused because only the last instance's 25 is compared |

### Step 3 — the sibling's calibration, re-run

263 archived overwrite pairs through the real implementation (up from 117; the archive has filled):
the visible axis refuses **4**, all real hollowings — the three this file lists plus
`idea.uk/index/tool-list` 1,118→0 — and the count is the SAME 4 at every minimum from 500 to 50 while
scope rises 153 → 261. So lowering that path's minimum costs nothing measurable, on its own evidence.

### Residuals, honestly

1. **The page-total floor still fails OPEN** on a measurement error, as the inline rule did. Its two
   siblings fail closed. It now files a `save_guard_unmeasured` work item so a later content loss on
   that page is not mis-diagnosed as a floor that should have caught it — but reconciling the three is
   a behavioural change needing its own evidence and is **not done**.
2. **120 of 1,079 pairs (11%)** still sit below the 200-char minimum and are unjudged by the text
   floors; the class-attribute floor is their only cover. The remedy if a defect turns up there is a
   lower minimum plus a harness re-run, which is now one command.
3. **The fourth copy of the retired axis stays**, deliberately: `load_current_section_content_action.go:262`.
   Allow-listed with its reason in `shrink_axis_coverage_test.go`.
4. **`page_component_writer_coverage_test.go` is still blind to this path** — it matches
   `UPDATE … SET rendered_html`, and the rebuild is DELETE+INSERT. My coverage test pins the three
   floors this action must call and fails on any unwired enforcer, but it does not generalise that
   older regex, so a NEW writer on another DELETE+INSERT path stays invisible there.

### How to close (revised)

1. Roll a chassis image built from `HEAD` ≥ `e42d57adf`. Verify at the artefact, not at git:
   `kubectl … exec <pod> -- sh -c "grep -oa 'buildinfo.GitCommit=[0-9a-f-]*' /proc/1/exe | head -1"`
   then `git merge-base --is-ancestor e42d57adf <that sha>`. ⚠ **Do not** grep the pod LOG for
   `build provenance` — the chassis logs whole agent payloads and they quote the phrase (LANDMINE).
2. Induce a refusal and check the ARTEFACT: the page's `page_components` rows byte-identical
   afterwards, plus the allow arm (a legitimate full-content save of the same page must succeed).
3. Watch a week with the **demand control** alongside — a zero refusal count means nothing unless
   `page_component_history` shows rebuild writes in the window. Expect ~0–1 refusals; hand-read any
   hit against the robot-hands precedent before classifying it.
4. Then move to `bugs_closed/` and add the pattern to 016b §9.

Commands for all of it: `docs024_key_docs_latest/bugfix_293_whole_page_shrink_axis/RUNBOOK`.
