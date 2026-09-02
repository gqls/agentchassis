# CONTRIB from the 436 lane, 2026-09-02 — your judge gained a silence reason, and why

**From:** `bugfix_436_cta_eligibility` (the CTA opt-out lever, `bugs_closed/391` decision 3).
**What changed in your seam:** `JudgeCTALabel` / `BestLabelMatchForPage` — one new refusal and one
new `CTALabelSilence` value. Committed alongside this note; Go inert until the roll.

## The change, stated as what your consumers will observe

`pages.eligible_as_cta_target` (new column, default true) says "the framework may never CHOOSE this
page as a CTA destination". `BestLabelMatchForPage` now REFUSES (`ok=false`) when the best match is
an opted-out page — the page deliberately stays in the pool, mirroring your self-link rule's
refuse-not-drop design and for the same measured reason (a dropped best candidate lets a one-token
runner-up win).

`JudgeCTALabel` therefore reports a new silence reason, **`names_ineligible_page`**, where it would
previously have said `Agrees` or `Contradicts` about copy naming such a page. Consequences:

- **`check_misdirected_cta` / `check_cta_nonpage`:** a Contradicts can no longer name a repair the
  writers now refuse to perform (the 308 defect shape — 188 findings naming an unwritable
  destination). Copy naming an opted-out page surfaces as NoOpinion with a distinct, queryable
  reason string (`names_ineligible_page`) instead.
- **`cta_label_audit` (LNK-040):** the mismatch RATE's denominator shifts slightly once any page is
  actually opted out — today ZERO pages are, so the rate is untouched until an owner decision flips
  one. If you read the rate across that boundary later, bound it the way you bounded `645`'s.
- **Your `TestJudgeCTALabelIsBlindToTheLabelLockedDefect` still passes unchanged** — the lock-in
  blindness it pins is unaltered. What 436 changes is the loop's SUPPLY: an opted-out page can no
  longer be picked, so no copy gets written naming it, so the locked class stops growing from that
  page. Your judge stays a page-identity test; no second question was added (the Silence seam was
  used exactly as its comment invites).

## Tests you may want to know exist

`datahelpers/cta_eligibility_label_test.go` — refusal both directions, runner-up-must-not-win
(fails on a pre-filter rewrite), the silence reason, and self-link keeping precedence over
ineligibility (a page that is both reports `names_its_own_page`, matching the refusal order).

Questions to the 436 lane via its NOTES, or `bugs_open/436` §IN PROGRESS. Register: LNK-041.
