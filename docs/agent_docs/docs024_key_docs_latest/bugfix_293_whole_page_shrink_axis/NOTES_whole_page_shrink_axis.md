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

## 2026-08-17, council round 1 — REVISE, and the gating objection was the right question

`823679dc-43d5-4f93-8b2d-746c41250290`, `complete_revise`, ~8 minutes. Five objections across four
seats (guidelines approved; 5 seats abstained on relevance). **Two of the five changed the code, one
was refuted by a measurement I should have run before submitting, and two were answered as asked.**

### The gating one: is `page_components.build_status` ever `'deployed'`? — REFUTED

Guardian (high), and editquality and guidelines independently flagged the same thing. The reasoning
was excellent: `LANDMINES.md` records the SIBLING table's `site_components.build_status` as
`'rendered'` and **never** `'deployed'`. If `page_components` matched, then the page-total floor I
extracted returns zero rows for every page, `existingTotal <= minPageTotalTextChars` short-circuits
every time, the floor **never engages** — and my sqlmock tests would certify it anyway, because a mock
supplies whatever rows you tell it to. Guardian's words: *"extracting the block is the moment to
verify the population predicate rather than carry a possibly-vacuous one forward with a new test suite
certifying it."*

`[MEASURED 2026-08-17]` It is the dominant value, so the analogy does not transfer:

```sql
SELECT build_status, count(*) AS rows, count(DISTINCT page_id) AS pages
FROM page_components GROUP BY 1 ORDER BY 2 DESC;
--  deployed  1575  617   |  approved  85  49  |  pending  19  19  |  removed  4  4
```
and per page, the guard's exact population: **617 of 729 pages (84.6%)** have at least one row with
`build_status='deployed' AND rendered_html IS NOT NULL`.

**This is the check I should have run before submitting**, and the reason I did not is instructive: I
carried the predicate over *verbatim* precisely so as not to change behaviour, and treated "unchanged"
as "already validated". An inherited predicate has whatever validity it always had — which may be
none — and **extraction is the moment that question becomes yours.** No code changed for this
objection; the measurement is the deliverable.

### The two that changed the code

- **bug_historian (medium): the fail-open path stood down silently.** *"A floor that fails open on
  error is indistinguishable from no floor at all on the very incident class this whole bug is
  about"* — so a later content loss gets diagnosed as "the floor should have caught this" when it
  never ran. It now files a work item under its **own** type `save_guard_unmeasured`, not the
  `save_refused_incomplete` its siblings use: nothing was refused on that path and a row saying
  otherwise is a false claim in a queue a human reads. Producer set, key shape and "no automated
  consumer, by design" stated at the constructor per the owner ruling of 2026-08-02 §1. It still
  fails OPEN — that stays a separate question.
- **reuse_agent (low): three `*IncomingBySlot` helpers, one shape.** Collapsed onto
  `incomingBySlot(sections, measure)`. Worth more than tidiness: **that duplication is how the
  last-write-wins keying defect came to exist in two of them independently.** The keying question is
  now answered once and the measure is the parameter — which is also the separation the harness needs.
- **editquality (low): "closes a hole" overstated what shipped.** It does not generalise the older
  writer-coverage test's UPDATE-only regex, so a new writer on some other DELETE+INSERT path stays
  invisible exactly as before. Corrected to "narrows" in the file header, and made *partly true* by
  adding `TestEveryFloorEnforcerHasACaller` — ANY `enforce…Floor` in the package with no caller fails
  the build, which is the generic case the claim was reaching for. **MUT-293-H** proves it bites.

### The one that turned into a new measurement

**bug_historian (medium): is "visible" defined once or three times?** `visibleTextLength` (mine, and
now all three floors) versus `sectionHasVisibleContent` in `rerender_single_page_action.go:910`, which
the page assembler uses to decide a section is empty and droppable. If they disagree, *a save can pass
this floor and the assembler can drop the same content anyway* — a second, independent silent-drop
path for exactly the class this floor protects. The seat said plainly it could not check this from SQL
or the code index and needed a human diff.

They **do** differ, by construction: parser walk (dropping script/style/noscript/template/code/pre/
svg/iframe/textarea/select/option/head with their content, entities **decoded**) versus a regex chain
(style blocks, script blocks, tags, entities, whitespace) with a **>10**-char threshold and an
explicit `reRuntimeFill` escape. Only one direction matters, though: nothing this floor judges as
prose-bearing may be droppable by the assembler.

`[MEASURED]` over all three populations — **6,585 sections in the floor's scope, ZERO droppable by the
assembler**; 944 in the harmless direction (the floor declines to judge, the assembler keeps).
`TestVisibleDefinitionsAgree` holds it, with a demand control that FAILS on an empty population. A
code reading reaches the same conclusion, but it cannot see the entity edge case — content that is
almost entirely entities counts ~0 for the regex and >0 for the parser — which is why this runs over
real rows rather than in my head.

**Round 2 resubmitted on the same correlation** (`RESUBMIT_CORR`), so the trail accumulates.
Eight mutations now run: A–H.

## 2026-08-17, council round 2 — APPROVED, and acting on the advisories changed the design

`823679dc`, run `d8974ec3`, `complete_approved` at 17:03Z (~6 minutes; round 1 was ~10). Twelve seats
reported, four abstained. **3 advisory objections, none high — and two of them were the same objection
from opposite directions, which is what made me look again.**

- **bug_historian (medium):** filing a breadcrumb for the fail-open path *"patches the SYMPTOM
  (visibility after the fact) while leaving the mechanism (fail-open on a content-guard) live and
  generic"*, and it named the estate's own rule back at me — *"a recorded decision with no enforcement
  point is decorative"*.
- **reuse_agent (medium), independently:** a work-item type whose whole point is that nothing acts on
  it **does not belong in the dispatch table**. `site_work_items` has `handler_agent`, `claimed_by`,
  `attempt_count`; `agent_error_log` is the table built for a record with no consumer.

**Both were right, and the thing I had not measured is what made the fix cheap.** I deferred
fail-closed in round 1 on the reasoning that it was a behavioural change and must not ride an axis
correction. That reasoning is fine; the estimate under it was wrong. This floor runs **FIRST** of the
three (`save_page_sections_action.go:593`, then `:603`, then `:614`) and **both of the others query the
same table for the same page moments later and REFUSE on a query error**. So the fail-open window was
never "an error means the page saves unguarded" — it was only "an error hitting this one statement and
not the next two", a blip between statements. Failing closed costs almost nothing and makes the three
consistent. `save_guard_unmeasured`, `savePageSectionsUnmeasured` and `pageTotalUnmeasuredFix` are
**deleted**: the change adds no shared vocabulary at all now, which also retires the architecture
seat's note about a new `item_type` and the constitution seat's low flag about *"the one place a
workaround-shaped deferral persists across two rounds"*. **MUT-293-I** proves the new path bites.

> **CORRECTED:** the PLAN's decision **D7** ("the page-total floor keeps failing OPEN") and the
> submission's risks paragraph are both now wrong. Left in place with this correction rather than
> edited away — the reason it changed is the useful part: *I scoped a deferral without measuring how
> big the thing I was deferring actually was.*

- **guardian (medium):** enumerate the OTHER consumers that inherit the 500 → 200 minimum, since this
  is a guard every content-mutating workflow shares. `[MEASURED]` four, and **none pins
  `section_shrink_floor`**, so all four take the new default: `page-build-handler`, `page-rerender`,
  `tool-recreation-handler` (via `save_page_sections`) and `section-editor` (via `apply_section_edit`).
  Recorded in the code, per the owner ruling of 2026-07-29 §3 — the consumers must be TOLD.
- **guardian (low):** the coverage test is a source scan and "not a structural guarantee against a
  fourth axis; fine as an interim guard, not as the permanent enforcement mechanism". Accepted as
  stated — it is the same weakness `page_component_writer_coverage_test.go` documents about itself, and
  it is disclosed in the file header rather than claimed away.
- **tooling_provenance (advisory):** no `doc_notes` row recording the round-2 dispositions for the next
  fixer. Written: `subject_key='save_page_sections-text-floors-axis-293'`, 5,124 chars, categories
  `decision`/`decision-record`/`council-gate`/`provenance` — all six decisions with their evidence,
  the enumerated consumers, the residuals, and how to re-run any of it.
- **prior_art_librarian (missing):** whether the 263-pair overwrite population was re-verified for
  this round or only cited from round 1. **Re-run this session**, not cited: the sweep over
  `pairs_overwrite.jsonl` through the real `evaluateSectionShrink` is where "the same 4 refusals at
  every minimum from 500 to 50, scope 153 → 261" comes from.
- **editquality → approve**, explicitly noting the round-1 correction was proportionate and that the
  remaining gap is "now honestly disclosed in the risks section rather than overclaimed". The
  architecture seat: `ARCHITECTURE_SIGNAL: point_fix | DEFLECTIONS: 0`.

Committed with `Council-Reviewed: 823679dc-43d5-4f93-8b2d-746c41250290` (`cd610a006`).

**Nine mutations run in total (A–I).** The estate's own line holds up: a REVISE round is cheaper than
the defect it finds. Round 1's gating objection cost ~10 minutes and asked the one question my
337-of-366 figure silently depended on; round 2's advisories cost ~6 and produced a simpler design
than the one I submitted.

## 2026-08-17, POST-ROLL — the axis is LIVE and has been exercised; two later commits are not in the image

`v1.0.1307`, both `agent-chassis` pods, started 17:05Z. Stamp read from the binary (never the log —
see the landmine): **`a6d1c53c0`**, identical on both pods, with the probe's own sanity control
(`gqls/agentchassis` present) passing.

| commit | in the image? | what it is |
|---|---|---|
| `6aae23e62` | **YES** | the axis on all three floors, minimum 200, the page-total extraction, the summing keying |
| `9cd887ddf` | **YES** | the class floor's keying |
| `4b32f174c` | **YES** | the section editor's axis — the 285 lane's half, live at last |
| `e42d57adf` | no | council round-1 revisions (the fail-open work item, `incomingBySlot`, two tests) |
| `cd610a006` | no | council round-2: the page-total floor failing **CLOSED**, and the item type's removal |
| `c04018a3e` | no | *negative control — committed after the stamp, must be absent, and is* |

**So the substance of 293 is live.** What is not: the page-total floor still fails OPEN *and silently*
in the running binary (the round-1 breadcrumb is not in it either). That is not a regression — the
inline block it replaced failed open silently too — and both sit behind the same one-statement-blip
window described above. They ride the next roll.

### Was it exercised, or merely present? — the demand control, and then the scope control behind it

`[MEASURED]` 11 whole-page rebuild writes across 3 pages since 17:05Z, and **zero** refusals naming
`VISIBLE text` or `PAGE CONTENT REGRESSION`. A zero refusal count is worth nothing on its own, so:

- **Demand control:** rebuilds are running — the 11 writes, 17:38 to 17:49Z.
- **Scope control, which is the one that mattered:** ran the committed harness over those 11 pairs.
  **10 of the 11 are IN SCOPE** on the live axis (≥200 visible chars on the existing side), so the
  guard genuinely judged ten real writes and allowed all ten — consistent with the calibration's zero
  false refusals in 1,079 pairs. Without this, "zero refusals" was equally consistent with a guard
  that judged nothing at all.
- And on those same ten sections the retired axis would have **allowed a total prose wipe on 10 of
  11**, where the live one refuses all 10. The protection is real on the pages that rebuilt this hour.

### What is NOT established, stated rather than glossed

**No refusal has fired yet**, so the refusal path — message text, work item, nothing written — is
proven only by the mutation-backed wiring tests and not at the artefact. The calibration predicts
about one a week fleet-wide, so waiting is the proportionate way to get it.

**I did not induce one, deliberately.** The recipe exists (the 285 RUNBOOK's), but on this path a save
that is *not* refused hollows a live page — the exact damage of `bugs_closed/285` — and the marginal
information over ancestry + nine mutations is small against that. The safe variant (raise
`section_shrink_floor` in a step's config so the next legitimate rebuild is refused, then revert)
writes nothing and destroys nothing, but it blocks another lane's build and edits a shared live agent
definition, which is an outward-facing change to make on request rather than unilaterally. **Flagged
for the owner as an option, not done.**

### MISSTEP 4 — I tried to prove execution from the pod logs, and the pods were already gone

The diagnostic line I added (`visible_text_total`, logged unconditionally on every save) was meant to
be the proof that the new code path ran. It returned nothing on either `agent-chassis` pod — because
`save_page_sections` does not run there. The archive's `application_name` gave the three connections
(`10.20.39.5/.10/.15`) and **none of them maps to a live pod**: page builds run in ephemeral
`agent-build-dispatch-loop-*`-style pods (22 pods share the chassis image), which are gone, and their
logs with them. **A log-based proof of execution has the lifetime of the pod that produced it, and for
an ephemeral handler that is minutes.** The durable substitutes are what I used instead: the binary
stamp for "is it in the image", and the archive plus the harness for "was it exercised, in scope".
