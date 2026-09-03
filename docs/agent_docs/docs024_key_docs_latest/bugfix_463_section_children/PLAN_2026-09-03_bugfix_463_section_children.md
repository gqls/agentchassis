# PLAN — bugfix 463, a section index's new children (2026-09-03)

## The problem, and the shape of it

`bugs_open/463`, filed by the `gamedesign.uk` lane, which declined the fix.
`reconcilePlanWithRealised`'s Pass C dropped every newly planned CHILD of a section index,
because it compared the FIRST PATH SEGMENT of each side and a child and a collider give the
same string. Pass A's union restores realised pages immediately afterwards, so the damage is
invisible everywhere except on a hub that is empty today — which can then never be filled.

## Decisions, and their reasons

**D1 — fix the producer, not the symptom.** Candidate 1 in the bug file (make Pass C
distinguish a child from a collider) is the only one that makes the bad state
unrepresentable. Candidate 3 (flat article URLs outside the section prefix) was refused: it
trades a dropped page for an orphaned one and breaks the `section_children` resolver that
`bugs_open/444`'s gate depends on.

**D2 — reuse `datahelpers.PagePathKey` rather than invent a comparator.** It is already the
estate's collision key, already tested, and — the load-bearing part — it is *complementary by
construction* to `countSectionChildren`'s prefix test, which is what 444's gate uses. "Claims
the hub's path" and "lives under the hub" partition the old first-segment test, so after this
the two guards in series cannot disagree about one page. Before, each one's evidence read as a
reason for the other.

**D3 — the second defect is IN SCOPE, and this was the significant call.** Fixing Pass C
alone changes nothing at the artefact: both write surfaces discard the planner's url and
re-derive it from `CanonicalisePage`'s role default, so the surviving child lands in `/blog/`
and the hub still resolves zero children. `bugs_open/463` §5 says the bug is "NOT about
`parent_section`"; that is true of Pass C and false of the write path, and taken as written it
would have produced a fix that passed its own tests and changed nothing on the served page.
Corrected in the bug file in place.

**D4 — the derivation lives in `ValidateRoles`, gated three ways.** It has exactly two
production callers and they are the two write surfaces `bugs_open/241` requires never to
disagree, so one definition serves both. Gates: leaf roles only (exactly the roles
`CanonicalisePage` reads `parent` for); absent value only; and **not** for an entry the
reconciler paired with a realised page — deriving there would honour a realised identity with
`honour_realised_identity` OFF, which `stampSameNameRealisedIdentity` deliberately refuses
(`bugs_open/215`). Placed after the role ladder and kept out of `declaredParents`, so no role
decision changes.

**D5 — do NOT build the observability half.** 463's candidate 2 asks for a durable finding on
a drop. The `428` lane shipped `recommended_type_reconciliation.go` for that class the same
day, classifying by STAGE rather than by pass — a vocabulary that survives these passes being
renumbered. Two lanes filing overlapping findings for one event is the drift this council
reviews for. This lane's contribution is the producer-side record (`reconcileCounts.DroppedPages`
names the page, url and pass), which their detector cannot derive.

**D6 — `bugs_open/467` filed, not fixed.** `truncatePreservingRealised` drops every net-new
page once the preserved set reaches `max_pages` (20). Same silent-shrink signature, different
pass, and `[MEASURED]` it affects 26 of 42 sites. It is a distinct decision needing an owner
ruling on what a re-plan may ADD, so it is a separate bug — but it means 463 must be verified
on a site under the cap.

## Phasing

1. Coordinate first (the file was being edited by another session when I started). ✅
2. Both code halves + mutation-proven tests, one commit. ✅ `9b540c2e6`
3. Council gate, submitted alongside rather than holding the code. ✅ `9f6c6374`
4. Docs: correct 463 §5 in place, file 467, `WRONG_CALLS`, `LANDMINES`. ✅
5. Verify at the artefact after the next chassis roll. ⏳ **not done — the fix is inert until then.**
