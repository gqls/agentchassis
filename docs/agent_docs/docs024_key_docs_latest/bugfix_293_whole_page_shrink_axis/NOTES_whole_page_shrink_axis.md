# NOTES — bugs_open/293, the whole-page shrink floor's axis

Append-only, newest at the bottom. Missteps are the point, not an appendix.

## 2026-08-17 — picking the bug up

- `bugs_open/293` filed this morning by the `bugfix_285_shared_template_write` lane, which fixed the
  SIBLING call site (the section editor) and states in the file that it is deliberately not touching
  this one. Ownership checked three ways before starting: `scripts/who-owns.py 293` says "OWNED or
  recently active" but every commit it names is the FILING lane's own renumbering; that lane's live
  transcript shows it has moved on to tool rebuilds (its last four user turns are about deconstructing
  and regenerating tool pages); and a grep of all live session transcripts for
  `save_sections_shrink_guard` found one other session touching the symbol, which is the `285`
  section-list lane and only in file lists. **Unclaimed.**
- **Still valid at HEAD** `[MEASURED]`: `git diff` against HEAD is empty for
  `save_sections_shrink_guard.go`, `section_visible_text.go` and `single_slot_floors.go`, and
  `shrinkGuardTagStripper = regexp.MustCompile('<[^>]*>')` is still what both sides of the whole-page
  comparison are measured with (`:58`, `:108`, and the SQL at `:137`).

## The evidence the bug file said was missing — and a better join than the one it proposed

293's "How to close" asked for pair evidence on the DELETE+INSERT path and offered two candidates:
(a) join delete rows to "the INSERT that replaced them", called *plausible but unproven*; or
(b) shadow-mode the axis and read a week of logs. **(a) turned out to be provable, and neither guess
was quite the right join.**

The rule is uniform across both ops, because migration 357's triggers bank `OLD.rendered_html` on
UPDATE *and* on DELETE: **an archive row is the content that was live until that moment**, so the
write that followed it is `row.rendered_html → (the next archive row for the same page+slot, or the
LIVE `page_components` row if there is none)`.

For the delete path the terminal case is the strong one, because `page_components.created_at` is
independent evidence that the re-insert belongs to *that* rebuild:

- 1,254 (page, slot) groups have an archived delete; 1,145 still exist live; **1,123 were re-inserted
  within 60 s of their delete, 1,109 within 5 s.**
- **Disconfirming control that came out right:** ZERO live rows are OLDER than the last delete of
  their own (page, slot). A wrong join would not produce that.
- **Positive control against a measurement made by someone else, before this join existed:** the same
  rule applied to `op='overwrite'` reproduces the 285 lane's three known refusals with identical
  figures — `idea.uk/tool-ab-test-calculator` 684→0, `webdesign.co.uk/learn-ai-builders-…` 2,143→16,
  `webdesign.co.uk/tool-ab-test-calculator` 684→0 — and their tag-stripped refusal, the repair, at
  38.1%. It also finds a fourth hollowing they did not report (`idea.uk/index/tool-list`, 1,118→0
  visible, 5,643→0 tag-stripped). A join that reproduces a number it did not produce is worth
  trusting on the rows nobody had paired.

## MISSTEP 1 — my first join manufactured a refusal, and hand-inspecting the single hit is what caught it

The first export keyed on `(page_id, slot_name)` alone. **Slot names repeat on a page** — 14 pages,
32 rows, `max 3` instances of one name — so the join emitted a CARTESIAN PRODUCT for those pages and
the run reported one refusal: `leopardessconsulting.co.uk/technical-architecture/generic-text-block`,
2,831→15. It does not exist. That page has three `generic-text-block` rows and the "pair" was one
instance's before against a different instance's after.

The cheap check that would have caught it before the run: `count(*) FROM page_components GROUP BY
page_id, slot_name HAVING count(*)>1` — one query, and I ran it only after the anomaly. Logged in
`WRONG_CALLS.md`. `(page_id, slot_name, position)` IS unique on both sides (0 non-unique of 1,618 live
and 3,603 archived), so the fixed export restricts to page+slot groups of exactly one on both sides
and pairs 1,079.

**The same defect is in the shipped guard, which is the part that matters.** `strippedIncomingBySlot`
keys a map on `ComponentName` (last write wins) and the existing side keys a map on `slot_name` (last
row scanned wins), so on those 14 pages the guard compares an ARBITRARY instance against an
ARBITRARY instance — and which one depends on DB row order and slice order, so the comparison is not
even stable. Not the axis defect 293 is about; found by tripping over it.

## What the axis measurements say

Three populations, all measured by the REAL functions (`visibleTextLength`, `evaluateSectionShrink`)
through a committed harness — `platform/orchestration/actions/shrink_axis_calibration_test.go`, which
skips unless `SHRINK_CALIBRATION_JSONL` names an export. Commands in the RUNBOOK.

**1. Would the new axis refuse writes that really happened?** Terminal pairs, 1,079, exact join.
Every pair is a write the LIVE tag-stripped guard ALLOWED (the guard shipped 2026-08-02, the archive
starts 2026-08-09, so refusals are censored out of this population by construction — which makes it
exactly the right population for the false-refusal question).

| axis | in scope | refuses |
|---|---|---|
| tag-stripped (live) | 1,062 / 1,079 | 0 |
| visible text | 492 / 1,079 | **0** |

**2. What does each axis actually protect?** The historical answer above is "nothing happened", which
is a measured absence of false refusals and is NOT an argument for the change. So the prospective
question, constructed on the same 1,079 real sections: delete every word of prose, keep the wrapper
markup and the `<style>`/`<script>` content — bugs_closed/285's exact shape — and ask each axis.

| axis | judged | REFUSES the wipe | ALLOWS it |
|---|---|---|---|
| tag-stripped (live) | 1,060 | 336 | **724 (68%)** |
| visible text | 492 | 492 | **0** |

The simulation carries three controls: the hollower must leave 0 visible chars, sections with no
prose to start with are excluded from the denominator rather than counted as protected, and the
population must contain prose-bearing sections at all.

**3. The 500-char minimum is the other half of the fix.** `minShrinkGuardChars` was chosen against
tag-stripped lengths, where a stylesheet inflates every count; the same 500 on visible text excludes
587 of 1,079 slots. Swept over both populations:

| minimum (visible chars) | terminal: in scope | refuses | intermediate: refuses | …that the guard would have judged |
|---|---|---|---|---|
| 500 | 492 | 0 | 8 | **1** |
| 300 | 729 | 0 | 13 | **1** |
| 200 | 959 | 0 | 15 | **1** |
| 120 | 1,046 | 0 | 15 | **1** |
| 50 | 1,066 | 0 | 15 | **1** |

The intermediate population (2,454 pairs, consecutive archive events) has a WEAKER join — nothing
proves the slot was re-inserted by the same rebuild — so it is used only to hunt for refusals, never
to count them. Attributing each refusal by its gap settles it: 7 of the 8 span 1,700 s to 93 hours,
and a slot absent for that long is a DROP, which the guard declines to judge (`!present → continue`).
So the cost of lowering the minimum is **zero additional refusals at every step down to 50.**

**4. The one real false refusal, inspected by hand.** `robot-hands.com/about/differentiators`,
2026-08-11 13:12:30, two rebuilds 96 s apart, both position 3: 3,724 → 1,554 visible chars (41.7%).
Read both sides — the incoming is a tighter, better-grounded rewrite ("10 gripper models from 6
manufacturers: Schunk, OnRobot, Robotiq, Zimmer Group, Festo, Schmalz", four named scoring
parameters) replacing looser prose. **A legitimate write the visible axis would have refused.** One,
fleet-wide, in eight days, against 724 prose wipes the shipped axis would wave through — and the
config escape hatch (`section_shrink_floor`) already exists for it.

**5. The axis is in THREE places, and the oldest copy is the blindest.** The page-total content
regression guard inlined at `save_page_sections_action.go:~549` tag-strips in SQL and refuses below a
quarter of the page's text. `[APPROXIMATE — paired slots only, not every deployed row]`: of 366
pages it would allow a **whole-page** prose wipe on **337**. On visible text it refuses 363 of 363.
That is the scope argument for fixing the axis once, in one shared place, rather than at the one call
site this bug names.

## 2026-08-17, later — the fix, six mutations, and a seventh that found a live false refusal

Planned with `fable` against the evidence above (brief and its answers summarised in
`PLAN_2026-08-17_whole_page_shrink_axis.md`). Three of its findings changed the design, and one
corrected my own count.

- **The axis is in FOUR places, not three.** I had missed
  `load_current_section_content_action.go:261-262`, which applies `shrinkGuardTagStripper` and
  compares against `minShrinkGuardChars` to decide whether an unclaimed stored slot is "prose-sized"
  enough to pair with an unmatched section. It refuses nothing — a pairing heuristic, not a floor.
  **This is what settled the design of the minimum:** I had intended to change
  `minShrinkGuardChars` from 500 to 200 in place, which would silently have re-tuned that action's
  pairing behaviour with no calibration covering it. Instead `evaluateSectionShrink` takes the
  minimum as a parameter and a NEW constant `minShrinkGuardVisibleChars = 200` is passed by both
  floors; 500 and the retired stripper stay, annotated with their one remaining consumer.
- **Rejected the obvious structural fix, on fable's argument.** I had planned to make
  `evaluateSectionShrink` take HTML maps so no caller *could* choose an axis. That would destroy the
  calibration harness, whose whole value is running BOTH axes through the REAL decision. The axis
  stays the caller's to supply and `shrink_axis_coverage_test.go` enforces it instead — a test can
  check "measured the right way"; a type cannot.
- **A hole in the estate's existing coverage test**, found while writing mine:
  `page_component_writer_coverage_test.go:39` matches `UPDATE page_components … SET … rendered_html =`
  and this path is DELETE+INSERT (its only `UPDATE`s set `position`). So the fleet's
  highest-volume `rendered_html` writer — 3,603 archived rebuild writes against 281 edit writes —
  was **invisible to the fleet's writer-coverage test**, and unwiring a floor there failed nothing.
  `TestSavePageSectionsWiresBothTextFloors` closes it.

### Mutations, RUN before commit — all seven bite

A mutation you did not run is a claim. Each was applied to a `git archive HEAD` tree, tested, reverted.

| id | mutation | result |
|---|---|---|
| MUT-293-A | per-slot guard back to the tag-stripped axis, both sides | **3 tests fail, in BOTH directions** — the refusal test because the retired axis allows the poison, the repair test because it refuses the repair |
| MUT-293-B | pass `minShrinkGuardChars` (500) instead of 200 | the mid-sized-prose test fails. **Nothing else in the suite could catch it** — every other fixture is comfortably over 500 visible chars |
| MUT-293-C | `+=` → `=` on both sides of the text floor | the repeated-slot test fails **for one row ORDER and passes for the other**, which is the finding |
| MUT-293-D | page-total floor back to the tag-stripped axis | it ALLOWS the whole-page wipe — so that floor is load-bearing and **not shadowed** by its per-slot sibling |
| MUT-293-E | unwire `enforcePageTotalTextFloor` from the action | coverage test fails |
| MUT-293-F | put `enforceSingleSlotFloors` back on the retired axis | coverage test fails twice (no `visibleTextLength`; retired stripper not allow-listed) |
| MUT-293-G | `+=` → `=` in the CLASS floor | fails — **and via the ALLOW arm, which is a worse finding than expected** (below) |

### MUT-293-G found a LIVE FALSE REFUSAL, not a blind spot

I applied the summing fix to the class floor for consistency, expecting it to close a blind spot of
the same shape. The mutation says the shipped behaviour is worse than that. With last-write-wins, a
page carrying two instances of one slot name with 60 class attributes between them, rebuilt to keep
**55** of them, is **REFUSED** — because only the last instance's 25 is compared against the group's
60, giving 42% against a 50% floor. So the shipped class floor can block a whole page save on a
legitimate rebuild of any of the 14 pages with a repeated slot name, and which way it falls depends
on the order the database returns rows in. That is a live defect with damage, not a gap.

### DEPLOY STATE, measured — and it corrects a figure in another lane's notes

`[MEASURED 2026-08-17]` Running chassis: `v1.0.1305`, both pods, restarted 14:43Z. Stamp read from
the binary (`grep -oa 'buildinfo.GitCommit=[0-9a-f-]*' /proc/1/exe`): **`6a782274b`, dated 08-16**.

- My fix (`6aae23e62`) is **NOT** in it, as expected — committed after the build. `IMAGE_TAG` in the
  makefile is already `v1.0.1307` while the chassis overlay still reads `v1.0.1305`, so the next
  build picks my commit up from HEAD.
- **`4b32f174c` — the 285 lane's visible-text axis at the section editor — is NOT in it either.**
  Their fence (`d7b2d9994`) is. So **the sibling's axis is still inert**, and both halves of the
  correction will now go live on the same roll. Anyone reading `PBP-043`'s "verify-later" and waiting
  for the section editor's first visible-axis refusal is waiting for a roll that has not happened.
- ⚠ **The 285 lane's NOTES record the v1.0.1305 stamp as `5e075a6f9`.** It is not: that sha is
  **absent** from the running binary, on a grep whose sanity control (`gqls/agentchassis`, present)
  passes. Not corrected in their file — it is their lane's record — but do not carry that sha forward.

**MISSTEP 2, mine, and it wasted three probes.** I ran the binary probe for a bare 40-hex sha, and
read the first "absent" as an answer before checking the positive control — which then failed, so the
whole reading was worthless. The stamp is stored next to a marker and the reliable extraction is
`grep -oa 'buildinfo.GitCommit=[0-9a-f-]*'`, which another lane's runbook already had. **The absent
control is the one that tells you the probe works; run it FIRST, not as a footnote.**

**MISSTEP 3: `kubectl logs | grep -m1 'build provenance'` returned another lane's text.** These pods
log entire council and diagnosis payloads, and a seat's SQL check quoted the phrase, so the recipe in
two runbooks matched a false line 2 hours newer than the pod's startup. It is not a startup line
scrolling out of reach (the documented failure) — it is a **collision with agent payload content**,
which no `--tail` size fixes. Use the binary marker. Filed as a LANDMINE.
