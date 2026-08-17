# 291 — the whole-page shrink floor measures CSS and SCRIPT as "text", so it cannot see prose deletion — and its one refusal in the archive was a repair

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
| tag-stripped (this guard) | 3,245 | 8,492 | **262%** | ALLOWS — a stylesheet replaced the article |
| visible text (style/script/comments excluded) | 2,754 | 68 | **2%** | refuses |

And the mirror, which is the part that makes this urgent rather than academic — across all 117
archived overwrite pairs the tag-stripped axis refuses **exactly one** write, and it is the
**REPAIR** (2026-08-15 18:18Z, seed 431, putting the article back): 38% kept on that axis, 4,050%
on the visible one. A guard whose only firing in eight days of history would have blocked a
repair is a guard that gets switched off.

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
