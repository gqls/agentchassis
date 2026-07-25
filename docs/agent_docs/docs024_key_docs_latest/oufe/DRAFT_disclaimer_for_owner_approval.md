# Draft standing wording for oufe.com — for owner approval

Status: **DRAFT, not applied anywhere.** Nothing below is live. This is not legal
advice; it is a starting text for the owner to approve, amend, or take to a
solicitor. Written 2026-07-25.

## Why this needs deciding before launch, and why a footer line is not enough

Three separate exposures, which need three different placements:

1. **The financial promotions perimeter (s21 FSMA).** Analysis and education sit
   outside it. Language that invites or induces someone to engage in investment
   activity does not. The original brief for this site used the phrase *"tools
   that can find opportunities"*, which drifts toward the wrong side of that line.
   Fixed by framing, consistently, throughout the site — not by a disclaimer.
2. **Negligent misstatement (Hedley Byrne).** This is the real exposure route for
   research someone pays for, and it is the one our own idea.uk liability work
   already identified. It is answered by a disclaimer only if that disclaimer is
   **conspicuous and proximate** — in the tool, at the top of the deliverable —
   not buried in a footer the reader never scrolled to.
3. **Defamation and malicious falsehood.** Characterising a named real company as
   distressed, vulnerable or in difficulty is a statement of fact about that
   company. Answered by sourcing, not by disclaiming: every such statement traces
   to a document we can link and date, or it does not appear.

## A. Footer line (every page)

> OUFE publishes educational analysis of financial and legal mechanism. Nothing
> here is investment advice, a recommendation, or an offer, and nothing here
> should be relied on when making a financial decision.

## B. Tool box (inside every interactive tool, above the fold)

> These figures are yours to set and the arithmetic is illustrative. The model
> shows how a priority waterfall distributes value under the assumptions you
> enter — it is not a valuation, not a forecast, and not advice. Real outcomes
> turn on security, guarantees, structural subordination, intercompany claims and
> contested valuation evidence, none of which are modelled here. Everything runs
> in your browser: nothing you type is sent anywhere or stored.

*(The current tool build already carries a version of this. It is the sentence
that matters most, because a tool that outputs a number is the thing a reader is
most likely to treat as an answer.)*

## C. Case dossier header (every page about a named real company)

> This page describes a live situation involving a real company. Every figure and
> quoted term on it links to the document it came from, with the date we read it.
> Where we have not verified something, we say so rather than estimating.
> Situations move; a page can be out of date before you read it. Nothing here is
> advice, and none of it should be treated as a view on any security.

## D. `/disclaimer` page (linked in the footer's legal group)

Longer form, covering: what OUFE is and is not; the sourcing standard; that
analysis of mechanism is not a prediction of outcome; the limits of every model on
the site; that no relationship is created by reading; correction and removal
route; and a plain statement that OUFE is an independent research publication, not
an authorised firm and not a company.

**Note the machinery:** `disclaimer` is already a recognised legal-nav page name,
so a page called that is picked up into the footer's legal group automatically.
Two things to check when building it: the nav label map has entries for privacy
and terms but **not** for `disclaimer`, so it may render with a default label; and
legal pages must be created the way migration 182 does it — content hand-written,
`rebuild_policy='owned'` plus a permanent component lock, and `rendered_html`
written at the same time — or a later rebuild will quietly rewrite the text.

## E. The correction and removal route (a commitment, not just wording)

From the vetcomparison rails, which this site inherits wholesale:

> If anything on this site is wrong, tell us and we will correct it and say that
> we have. If you are the subject of a page and believe it misstates your
> position, write to us and we will review it promptly.

This needs a real monitored address behind it before it is published. Proposed:
the site's contact address, currently `oufe@contactforsales.com`.

## What the owner needs to decide

1. **Approve, amend or replace** A–E. They are drafted to be true rather than
   defensive; a solicitor may want them more defensive.
2. **Whether to say "not a company" explicitly.** oxenunity.com was resolved by
   making no entity claims at all. oufe.com is harder, because a research
   publication naturally invites "who are you?" — and the honest answer is
   currently "an independent research project, not an incorporated firm". My
   recommendation is to say exactly that on the about page and in the disclaimer.
   It costs a little authority and buys the thing this site is actually selling.
3. **Whether to take this to a solicitor before launch or after.** The idea.uk
   precedent is a ~£200–500 fixed-fee UK review, and its own terms are still
   flagged as drafts pending one. The vetcomparison precedent is that the owner
   can decide to proceed ahead of review on defined narrowing conditions, provided
   the decision and its conditions are recorded contemporaneously. Either is
   defensible; drifting into launch without choosing is not.
4. **Limitation of liability, if and when anything is sold.** Not needed for the
   free Phase 1. The recorded position from idea.uk is information-not-advice plus
   liability capped at the fee.
