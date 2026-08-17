# PLAN — bugs_open/293, the whole-page shrink floor's axis

Design, phasing, decisions **and their reasons**. Corrections to the originating brief live here,
marked as corrections, never edited away.

## The problem in one paragraph

Three guards on the page-save path exist to refuse a rebuild that has deleted most of a page's
writing. All three measured "text" as `<[^>]*>`-stripped length, which strips TAGS but not what is
inside `<style>` and `<script>` — so CSS declarations and JavaScript source counted as prose. The
failure mode they exist to stop therefore made their number go **up**: on the write that emptied a
live webdesign.co.uk article (`bugs_closed/285`) the per-slot floor read 262% retained. The section
editor's copy of this judgement was corrected on 2026-08-17; this bug is the other two, and it is the
higher-volume path — 3,603 rebuild writes against 281 edit writes in the same eight days.

## What the brief asked for, and where it was wrong

`bugs_open/293` set the gate as "get evidence for the DELETE+INSERT path", offering (a) a
reconstructed join it called *plausible but unproven* and (b) shadow-mode for a week.

> **CORRECTION 1 — the successor was never missing, so (b) is unnecessary.** The brief's premise was
> that a rebuild's "after" state is an unarchived INSERT. It is the **live `page_components` row**,
> and `page_components.created_at` is independent evidence that the re-insert belongs to that
> rebuild. That gives 1,079 exactly-paired writes with a disconfirming control that came out right
> (no live row is older than its own last delete) and a positive control that reproduces another
> lane's three known refusals to the character. Shadow-mode would have cost a roll and a week to
> produce pairs that already existed.

> **CORRECTION 2 — the brief's one-line fix is not sufficient, and the evidence says so.** It read:
> "One line: feed `visibleTextLength` … instead of `shrinkGuardTagStripper`. The pure decision,
> `minShrinkGuardChars` (500) and the config opt-outs stay as they are." Measured, that change alone
> takes the guard from judging 1,062 slots to judging **492**, because 500 was calibrated against
> CSS-inflated lengths. On this path the axis swap by itself trades "measuring the wrong thing" for
> "not measuring at all" on more than half the population. The minimum had to move with the axis.

> **CORRECTION 3 — the axis is in FOUR places, not two.** The brief scoped the bug to
> `save_sections_shrink_guard.go`. There is also the page-total content-regression guard inlined in
> the action (the blindest of them), a diagnostic log line that advertised itself as reporting the
> guard's own arithmetic, and a fourth consumer in
> `load_current_section_content_action.go` that is deliberately left alone.

## Decisions, and why

**D1 — visible text on all three floors, not one.** 016b §9's defect class is "one call site of a
shared judgement gets the rigorous fix; the sibling stays heuristic". Fixing the call site this bug
names and leaving the page-total copy would have reproduced that class *inside its own fix*. Evidence
for including it: on 366 pages the page-total guard would allow a **whole-page** prose wipe on 337,
where visible text refuses 363 of 363 and refuses **zero** real rebuild writes.

**D2 — the argument is mechanism, not history, and the history is stated as such.** Over 1,079 real
rebuild writes the new axis refuses nothing — because no rebuild in the archived window hollowed a
page. That is a measured absence of *false refusals* and it is **not** an argument for the change;
quoting it as one would be "a post-fix zero needs a demand control" in a new costume. The
justification is a constructed measurement on the real population (delete every word of prose, keep
the wrapper and its CSS/JS): the shipped axis allows that on **724 of 1,060** sections.

**D3 — the minimum becomes a parameter, and 200 not 120.** `section_visible_text.go`'s own header
predicted both the need and the shape ("the fix is a lower minimum for this axis — which needs
`evaluateSectionShrink` to take it as a parameter, not a second copy of the ratio rule here"). The
sweep shows the guard-judged refusal count pinned at **one** at every step from 500 down to 50, so
120 was free on *this* path. 200 was chosen because it is the deepest step the **section editor's**
own population also covers (263 overwrite pairs: scope 153 → 204, the same 4 refusals, all real
hollowings). Taking 120 fleet-wide would move that path onto a minimum with no evidence of its own —
the exact mistake this bug was filed to avoid, one parameter over. The 200 → 120 step buys 87 slots
and is not worth the gap; lowering later is one constant plus a re-run of the committed harness.

**D4 — `minShrinkGuardChars` (500) keeps its value.** Not decoration: its remaining consumer,
`load_current_section_content_action.go:262`, uses it with the retired stripper to judge whether an
unclaimed stored slot is "prose-sized" enough to pair with an unmatched section. It **refuses
nothing** — a pairing heuristic, not a floor — and re-tuning it would change which slots pair, with
no calibration covering that decision. This is why the minimum is a new constant rather than an
edit to the old one, and it was fable's find, not mine.

**D5 — the anti-drift mechanism is a coverage test, not a type.** The obvious structural fix is to
make `evaluateSectionShrink` take HTML and measure for itself, so no caller *can* choose an axis. It
was rejected: the calibration harness's whole value is running BOTH axes through the REAL decision on
real pairs, and a decision that measures for itself can only be trusted, never calibrated. So the
axis stays in the caller and `shrink_axis_coverage_test.go` enforces it — a caller of the decision
that does not measure with `visibleTextLength` fails the build, as does a file applying the retired
stripper without a reason in `retiredAxisConsumers`, with a vacuity control so a scan that matches
nothing cannot pass. Precedent and its stated weakness both come from
`page_component_writer_coverage_test.go`: it reads SOURCE, so it proves wiring EXISTS, not that
wiring EXECUTES. The behavioural half is the mocked wiring tests.

**D6 — summing per slot name, not keying by position.** Both sides of both floors were maps keyed on
slot name, last-write-wins, so on the 14 pages with a repeated slot name the comparison was between
an arbitrary instance and an arbitrary instance, decided by DB row order. Position was available and
unique — and rejected: `LANDMINES.md` records that a repeated slot name is **normal** (11 of 17
groups legitimate), and position survives a rebuild in only 1,051 of 1,079 pairs. Summing needs no
uniqueness and no position assumption, is bit-identical for the ~97% of groups with one instance, and
is more faithful than before because the insert loop writes *every* instance. Stated consequence:
instances individually under the minimum can aggregate over it, which brings the group into scope.

**D7 — the page-total floor keeps failing OPEN, and keeps its own population.** Its two siblings fail
closed on a measurement error; it does not, because the inline rule it replaced did not. Making it
fail closed is a behavioural change deserving its own evidence, and it must not ride an axis
correction. Likewise it keeps `build_status = 'deployed'` rather than the agent-writable predicate:
the two answer different questions ("is the page about to lose what it is currently SERVING?" versus
"may automation rewrite this row?") and folding them together would silently change one.

## Phasing, and what is inert until when

1. Evidence + the committed calibration harness — `458a85e28`.
2. The axis, the minimum, the page-total extraction, the keying on the two text floors, and the
   tests — `6aae23e62`. **Go, so inert until an image is built and rolled.**
3. The class floor's keying — `9cd887ddf`, separately because it is a different floor and a
   different defect (and MUT-293-G showed it is a live *false refusal*, not a blind spot).
4. Council submitted alongside, `823679dc-43d5-4f93-8b2d-746c41250290`, `Council-Submitted:` trailer.
5. Post-roll verification at the artefact, with a demand control — RUNBOOK.

`[MEASURED 2026-08-17]` The running chassis is `v1.0.1305`, stamp `6a782274b` (08-16). Neither this
fix nor the 285 lane's sibling axis (`4b32f174c`) is in it, so **both halves go live on the same
roll** and `bugs_open/293` stays open until they do.

## What would falsify the design

- A refusal in the first week that reads like the robot-hands case — a legitimate tightening rewrite
  blocked. One in eight days was the measured rate; several would mean the floor (0.5) is wrong for
  this path, not the axis, and the remedy is `section_shrink_floor` in config, live-immediate.
- A hollowing that ships anyway with visible text ≥ 200 on the existing side and the class floor
  silent. That would mean the 0.5 ratio is too slack, not that the axis is wrong.
- The 11% of pairs still below the 200-char minimum producing a defect the class floor misses. The
  remedy is a lower minimum plus a harness re-run, which is now one command.
