# PLAN — bugfix 280: decision guards' "stored assembly" silently omits chrome

**Started:** 2026-08-16. Picked up after confirming bug 180 (what the owner
originally named) is closed-and-live and unrelated; the owner then confirmed
280 — the still-open sibling flagged in `bugfix_270_missing_structure`'s
handoff — was the intended target.

## The defect, in one line

`check_decision_guards.go`'s `storedPageAssemblySQL` reads
`pages.rendered_header` / `pages.rendered_footer`, the same vestigial columns
`bugs_open/270` documents as empty on all 694 pages fleet-wide. Chrome lives
in `site_components`, not those columns, so every decision guard is silently
evaluated against a document with no chrome in it — full detail, evidence and
the "why no wrong verdict has been observed yet" census: `bugs_open/280`
(read that first; not re-derived here).

## Decision: fix candidate 1 (retype to `site_components`), not candidate 2

The bug file already weighed both options (§ "Fix candidates"). Candidate 2
("declare stored assembly body-only by design") was rejected here for the
same reason 270 rejected the equivalent "delete the check" option: it would
require also correcting the file's own header comment, which currently
*already* describes the post-candidate-1 behaviour correctly ("chrome +
page_components.rendered_html") — so candidate 1 makes the code match a
comment that already exists, rather than requiring a second edit to make the
comment match the code. No new decision to record beyond that; nothing here
contradicts the bug file's own reasoning.

## What changed

`storedPageAssemblySQL` (used by both the check and its verifier — they
share the one constant deliberately, "so the two cannot drift") now
concatenates, in order:
1. the site's `header` slot `rendered_html` from `site_components` (scoped
   by `site_id`, since chrome is site-level, not page-level),
2. the site's `footer` slot, same table,
3. the existing `page_components` aggregation for the named page, unchanged.

The `FROM pages pg WHERE pg.site_id = $1 AND pg.name = $2` gate is
**unchanged** — load-bearing: it is what makes the query return zero rows
when the named page does not exist, which both the check (skip, "not this
check's finding") and the verifier (fail-closed error, RFC_017) depend on to
distinguish "page missing" from "page exists with empty chrome." A rewrite
that dropped this FROM clause in favour of a bare `SELECT <subqueries>`
would have silently broken that distinction — checked and avoided (see NOTES
for the reasoning trail).

`site_components` has a UNIQUE (site_id, slot_name) constraint, confirmed via
`\d site_components` before writing the SQL — so each header/footer subquery
returns at most one row and a scalar `SELECT` (not `string_agg`) is correct
and sufficient.

## Tests

`check_decision_guards_test.go` (new file), following
`check_missing_structure_test.go`'s established pattern for this exact bug
family:
- `TestDecisionGuardsAssemblyReadsSiteComponents` — asserts on the SQL text
  itself (the anti-regression test; a happy-path sqlmock test cannot catch a
  reversion to `pg.rendered_header`, since sqlmock returns whatever row the
  test hands it regardless of the WHERE clause). **Mutation-tested**: reverted
  the constant to the old SQL, confirmed this test — and only this test —
  fails; restored the fix, confirmed all tests pass again. (Practice: "a
  mock's own bookkeeping cannot assert a negative — mutate to prove the
  guard.")
- `TestDecisionGuardsContainsGuardSeesChromeContent` — demand control for the
  false-positive shape: a `contains` guard on header content, fed an
  assembled string that actually carries chrome, must not fire.
- `TestDecisionGuardsNotContainsGuardCatchesChromeRegression` — demand
  control for the false-negative shape: a `not_contains` guard on chrome
  content, fed an assembled string carrying that content, must fire.

The verifier (`VerifyDecisionRegressionResolved`) needed no separate edit or
test — it references `storedPageAssemblySQL` directly (confirmed by reading
the file), not a copy, so the one SQL fix covers both call sites exactly as
the file's own comment intends.

## Sequencing after the code lands

Same shape as 270's precedent (see that workstream's HANDOFF): commit
narrowly, submit to the council gate (advisory), and leave image build +
fleet roll to the owner unless explicitly asked — the owner said exactly that
for 270 one day ago and there is no reason to assume it differs here. Record
whatever the owner actually says this time in NOTES/README rather than
assuming the prior answer carries over unstated.

## Relations

- `bugs_open/280` — the bug file itself; do not re-derive its evidence here.
- `bugfix_270_missing_structure` — sibling fix, same root cause class
  (a reader of the vestigial `pages.rendered_header/footer` columns), same
  remedy shape (retype to `site_components`), different failure mode (270
  fired unconditionally; 280 silently mis-evaluates a predicate). Its
  workstream directory is the precedent for process here — read its HANDOFF
  before assuming a step (build, roll, close) is this session's to take.
- LANDMINES.md, "`pages.rendered_header` / `rendered_footer` / `rendered_head`
  are VESTIGIAL" — once 270 and 280 are both live, its "read by exactly one
  caller left in the tree" line becomes stale in the other direction (zero
  callers); update the pointer once this ships (280's own Relations section
  already flags this).
