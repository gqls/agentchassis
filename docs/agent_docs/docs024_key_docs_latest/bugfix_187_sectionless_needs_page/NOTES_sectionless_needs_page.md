# NOTES — bugfix 187: needs_page items minted for section-less pages park permanently

Append-only, newest at the bottom. Evidence, commands, what the system said,
and every misstep.

## 2026-08-03 ~22:00 BST — lane opened, bug claimed, validity re-verified

- Picked 187 after an ownership sweep of bugs_open: 188 and 181 are held by
  live sessions (transcript grep: 8 and 14 focused hits respectively); 184/186/189
  were filed by still-active lanes (mortgagecalculator, thunder-reaper,
  loancalculator). 187 was filed by the bugfix_177 lane *at the council's
  direction* as a deliberate spin-out, and that lane closed same day —
  cleanest unowned ticket. Claim committed `411c6ab6f`.
- **Validity: confirmed and GROWING.** Census (live, 2026-08-03 22:00):
  image-build-handler 11→14, page-rerender 2→4 since filing; three rows minted
  TODAY (tool-ai-data-risk-checker, tool-ai-readiness-quiz, password-entropy).
  28 parked `needs_page` rows total in `needs_human_review`.
- No competing open work item in `site_work_items` (one stale `unresolved`
  needs_content_page from 07-15, unrelated tool-arena).

## 2026-08-03 ~22:20 — data-side triage of all 28 parked rows

Query joined each item to its target page BY NAME (27/28 have `page_id` NULL —
first lesson: the items don't link the page row even when it exists).

Three classes, matching the bug file's predicted split:

1. **177's shape — unsatisfiable at birth (17 rows).** All of
   image-build-handler's tool-page rows + page-rerender's rows: target page
   EXISTS (created long before the item — `page_newer_than_item = f` on every
   row), `pages.sections = []`, 0 current-plan rows, 1 slot (the finished
   tool-page shape per 177's fleet census). The emitters ask for a section
   rebuild of a page that declares no sections anywhere in the handler's
   resolution chain.
2. **Satisfiable-now or already-built, parked for ever (7 rows).**
   - grip-styles (gemini-p7), brands-index, shop-index: sections + current-plan
     rows present NOW, 0 slots — buildable today, item parked for a week+.
   - tungsten-guide, board-setup (3 slots = 3 declared), cases-index,
     thames-water (5/5): the page was BUILT since, by another route; the item
     is stale and nothing closes it.
3. **Real gaps + one moot (6 rows).** reconcile_site_plan rows for pages with
   0 sections AND 0 plan rows (directory-index, practice, guides-index,
   brand-detail, platform-log-index) — the bugs_closed/015 shape: a page that
   should have sections and doesn't; the item is SURFACING a real defect, do
   not suppress. learning-center-post: page `archived` — item moot.

## 2026-08-03 ~22:30 — the bug file's revalidator claim is FALSE

187's file says "needs_page IS drainable by the revalidator but these rows
predate/evade it — check why before hand-sweeping". Checked
(`revalidate_review_queue_action.go:149`): the `reviewRevalidators` map holds
exactly `unresolved_cta`, `required_fields_missing`, `needs_section_data`.
**`needs_page` is an UNCOVERED type — nothing drains these rows, ever.** The
check was one grep. Correction recorded in the bug file; WRONG_CALLS entry
belongs to the filing claim, recorded there with the cheap check.

Design consequence: the park-for-ever half of this bug is not "rows evading a
drain" but "no drain exists" — and the drain's question ("is this item's ask
now satisfied?") is the SAME satisfiability question the emit guard asks.
One shared resolver can serve both ends.

## 2026-08-03 ~23:00 — sweep drafted; near-miss caught by encoding the check

- 090 filed (run corr `b3dcb102-d4bf-44c1-b2a2-3068ce95acc6`); implementation
  dispatched to an opus agent per the PLAN.
- **MISSTEP (caught before apply): the sweep's first verify block asserted 3
  leftover rows — the real number is 7.** I had measured plan SECTION rows per
  page but not plan MEMBERSHIP (`site_plan_pages`), and 4 sectionless tool
  pages (grip-force-friction, gripper-cycle-time, gripper-payload, matchmatrix)
  ARE current-plan members — synthesis-eligible, so "unsatisfiable" is not
  provable and they must not be swept. What caught it: writing the DO/RAISE
  verify forced the exact leftover list, and re-running the measurement with
  the membership column contradicted the draft. The check IS the catch —
  same lesson as the marker rule: encoding the expectation is what surfaces
  the wrong one. Sweep = `sql_for_agents/300`, now asserting 7.
- Corollary worth keeping: those 4 rows mean the emit guard alone does NOT
  zero the leak for plan-member sectionless pages — the guard (correctly,
  conservatively) lets them emit, synthesis then finds no same-role donor, and
  they park. That residue is a PLAN-shape question (TL-009 territory), named
  in the bug close-out, not silently absorbed.

## 2026-08-03 ~23:25 — 090 verdict: UNVERIFIABLE (tooling), no refutation; one useful datum

- Run `b3dcb102` COMPLETED after 3 iterations: **UNVERIFIABLE**. Its own
  stated reason: both targeted code searches hit the stale code index
  (0-row = UNKNOWN per the index's own staleness notice — the known landmine,
  frozen at `d98010e8b`), and its reads of `reviewRevalidators` /
  `check_has_ready_sections` bodies failed ("tooling failure"). It neither
  confirmed nor refuted the mechanism; nothing in the citations contradicts
  the first-hand verification recorded in the bug file (per the 2026-07-31
  owner ruling, that verification is the stated substitute: emitters read at
  HEAD and quoted, handler chain read, all 28 rows measured, natural
  experiment inherited from 177).
- The run's one real find: `needs_page` rows CAN be hand-unparked — spec
  fields `promoted_by`/`promoted_reason` on rows `7e515460`/`b17664fe`
  (dartsonline `beginners` + sibling, backfilled 2026-07-29 by the
  dartsonline-traffic lane, both now `complete`). Its data_request answered
  live: exactly 2 rows fleet-wide, one actor, one day, and `promoted_by` has
  0 Go hits at HEAD (grepped in the tree, NOT the index) — a one-off manual
  intervention, not a mechanism. This is the strongest possible confirmation
  that no systematic unpark path exists — the revalidator gap is real, and
  the two promoted rows are the manual precedent for what `revalidateNeedsPage`
  mechanises with evidence.
