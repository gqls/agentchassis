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

## 2026-08-04 ~00:15 — implementation done; opus capped at the finish line; verified at HEAD+overlay

- Implementation ran as an opus agent per the owner's instruction and died on
  the weekly usage cap having completed ALL edits (resolver file including
  `revalidateNeedsPage` with a live-measured slot-matching rationale — 95.7%
  slot_name vs 87.5% function match, thames-water's double-mount case; both
  emitter guards; map registration; 24 tests). This session (Fable) finished
  the remaining verification. **Deviation from "use opus to fix": the cap is
  hard until Aug 7; the last ~5% (test-harness fix + verification) is mine.**
- **MISSTEP (the agent's, caught by running the tests): 3 FlagPageImageRebuild
  tests fed `site_id` as literals in `StepConfig.Config`** —
  `ExtractActionInputs` treats config strings as REFERENCES into collected
  data (the `bugs_open/042` family), so extraction failed. Fixed by moving
  inputs to `CollectedData`, which is also how the live workflow delivers
  them. The production code was right; only the harness was wrong.
- Verified against `git archive HEAD` + my 8 files overlaid (the tree carries
  other sessions' WIP in the same package): `go build ./platform/...` exit 0,
  `go test ./platform/orchestration/actions/...` exit 0. First attempt filled
  /tmp (16G tmpfs at 97% — the concurrent chassis build lives there); moved
  the check to $HOME/.cache and freed the archive immediately.
- Mutation spot-checks, live, in the overlay: guard disabled via
  `satisfiable && false` → the skip test FAILS on the unexpected `Begin` in
  BOTH emitters (first attempt `if false` didn't compile — a compile error is
  a guard in series, not a behavioural kill). Old symbol
  `toolPageDeclaredSections`: 0 hits repo-wide — the pod-grep negative
  control is armed. gofmt clean; 177's tests unchanged except the two
  permitted call-site renames.
- Council submitted: SUBMISSION_CORR `e2e87b04-2847-4ca5-bd63-f6f2d4aa1990`
  (schema lessons: `plan` is an OBJECT with summary/edits/grounded_in/risks;
  `create` is spelled `add`). Committing with `Council-Submitted:` per the
  2026-07-30 rule.

## 2026-08-04 ~00:40 — council APPROVED round 1 (5 advisories, none high); the checkable ones checked

Corr `e2e87b04`, decision `approved`, 13 seats reviewed / 4 abstained.
Advisories and their answers, each checked rather than nodded at:

1. **editquality (sweep matches `pages.sections`)** — rests on a false
   premise: the objection says the build "never reads pages.sections", but
   the loader reads it as fallback 3 (`load_page_sections_from_spec_action.go`)
   and the resolver mirrors exactly that. A stale non-empty cache means the
   HANDLER would find sections, so excluding such rows from the sweep is
   correct, and exclusion is the safe direction regardless. No change.
2. **bug_historian (declaredPageSections fails open to "declares nothing")** —
   real, inherited verbatim from 177's guard (the extraction is pure by
   design). Bounded: the emit guard also requires NOT plan-member before
   skipping, and membership fails open to TRUE, so suppressing a legit emit
   needs all three section sources erroring while the membership query
   succeeds-false. On the revalidator side the same failure lands on
   `unknown` (stays queued) — safe. Recorded as an open line in WII-010.
3. **reuse_agent + guardian + prior_art_librarian (revalidator judges EVERY
   needs_page producer fleet-wide)** — measured, not argued: every parked
   needs_page row across ALL producers (7 sources once the error filter is
   dropped) carries `spec.page_name`, so none hits the unknown-spec arm
   blind; reconcile's real-gap rows resolve no sections → `unknown`, parked,
   exactly as the 015 caution demands; `rejected` manual rows are terminal
   and never loaded. The verdict asymmetry (close only on positive
   name-matched evidence) is the containment.
4. **debug_historian (how is the roll confirmed before the sweep?)** — the
   pod-grep pair in the PLAN's verify section IS the answer (positive
   `declaredPageSections`/`skipped_sectionless_page`, negative
   `toolPageDeclaredSections`, both replicas, one exec) and will be executed
   before `300` is applied.
5. **tooling_provenance (check doc_notes before trusting the file register)**
   — 74 notes mention needs_page/page-build-handler (latest today); the
   load-bearing ones (177's fix note, the b3dcb102 verdict) are already cited
   in the lane docs.

Census moved overnight: one more unsatisfiable row
(`tool-spawn-rate-balancer`, decl 0, planless, image-build-handler 22:59Z) —
the sweep predicate catches it and the DO/RAISE leftover set is unchanged.
The leak stays live until the image rolls, which is the point of rolling it.
