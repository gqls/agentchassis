# Standing wording for oufe.com

> **APPROVED by the owner, 2026-07-26** — choices and wording, sections A–F.
> Section G (limitation of liability) was added on his instruction at the same
> time and is **new text he has not yet seen**, so it is the one part still
> awaiting a read.
>
> **Live now:** B and C (they shipped with the tool build).
> **Approved, not yet on the site:** A, D, E, F — they need the legal pages
> building, which is the migration-182 pattern (hand-written content,
> `rebuild_policy='owned'` + permanent component lock, `rendered_html` written in
> the same migration, or a later rebuild silently rewrites the legal text).

This is not legal advice; it is text for the owner to approve, amend, or take to a
solicitor.

Written 2026-07-25. **Rewritten 2026-07-26 on the owner's direction** — see the
correction note below, because the change is a change of posture and not a change
of wording.

> **CORRECTED 2026-07-26.** The first draft's central promise was *"every figure
> here links to the document it came from"*. The owner struck it. It is a claim
> about our reliability, and it implies that a linked figure is a correct figure —
> which does not follow. A citation proves where something came from. It does not
> prove we read the source properly, that we picked the right passage, that the
> source itself is right, or that the model did not invent the sentence around
> it. The replacement posture is the opposite of a promise: **we make mistakes,
> the tools can be wrong, the sources can be wrong, and our reading of them can
> be wrong — and we cite everything anyway so you can check us.** Citation
> becomes a tool for the reader to catch us with, rather than a warrant that we
> are right.

## Why this needs deciding before launch, and why a footer line is not enough

Four exposures, needing four different placements:

1. **Our own fallibility.** This site is assembled with substantial machine
   assistance. Models invent things fluently, and the most dangerous output is
   the plausible one: a real-looking figure in a well-formed sentence next to a
   real citation. Our own estate has done exactly this — invented prices for
   3,124 named real vet practices, and separately written invented figures back
   over correct ones during a routine rebuild *with both safety systems live*.
   This is answered by saying so, prominently, in the reader's path.
2. **The financial promotions perimeter (s21 FSMA).** Analysis and education sit
   outside it; language inviting someone to engage in investment activity does
   not. The original brief said *"tools that can find opportunities"*, which
   drifts the wrong way. Answered by consistent framing, not by a disclaimer.
3. **Negligent misstatement (Hedley Byrne).** The real exposure route for
   research someone relies on. Answered by a disclaimer only if it is
   **conspicuous and proximate** — in the tool, in the output, at the top of the
   page — never only in a footer.
4. **Defamation and malicious falsehood.** Calling a named real company
   distressed is a statement of fact about it. Answered by sourcing and by
   marking our reading as our reading, not by disclaiming.

## A2. Persistent banner, every page — DECIDED 2026-07-27, LIVE

The owner chose a persistent banner over a footer-only line and over an entry
interstitial, in a plain professional register. It sits directly under the header
on every page, needs no click, and cannot be dismissed and forgotten:

> **Analysis of mechanism, not investment advice.**
> We cite our sources so you can check them, and our reading of them can be wrong.

Two words on why it is phrased that way. It does **not** say every figure is
sourced, which is the promise struck on 07-26 and which we cannot keep. And it
does not apologise: for a professional reader an apologetic banner reads as
amateur, while a flat statement of what the site is reads as a research house.

Styling uses `--color-text-muted` on `--color-surface`, measured at **5.12**
against AA's 4.5, with the bolded lead in `--color-text` at 11.32. Checked before
writing it, not after — the site had a 1.00 contrast bug the same day.

The entry interstitial was rejected for the reason already argued against staged
warnings in the tool: repeated click-throughs go blind fast, and a warning nobody
reads is weaker in practice than one they cannot miss.

## A. Footer line (every page)

> OUFE publishes educational analysis of financial and legal mechanism. We make
> mistakes, and some of what is here is assembled with AI assistance that can
> invent convincing detail. Check anything that matters against the primary
> source. Nothing here is investment advice or a recommendation.

## B. Tool: condition of use, shown before the tool becomes usable

The owner asked whether a disclaimer could be a condition of using the tool. Yes,
and the recommended form is **one acknowledgement before first use**, not a
caveat at every step — see the note on staging below.

> **Before you use this tool.**
>
> This model can give you a wrong answer. It is a simplification: it computes one
> arithmetic rule over the numbers you type, and real outcomes turn on security,
> guarantees, structural subordination, intercompany claims, contingent
> liabilities and contested valuation evidence — none of which it models. Any
> pre-filled figures are illustrative and may be out of date or simply wrong.
>
> Treat every result as **a possibly inaccurate worked example**, useful for
> understanding how the mechanism behaves, and not as a statement about any real
> company or any real decision. Do not rely on it.
>
> [ I understand this tool can be wrong ] → reveals the tool

## C. Caveat inside the tool's own output

The most important placement, and the one that is easy to miss: **put the caveat
in the result, not just around the tool.** A result gets screenshotted, pasted
into a deck, and forwarded; the surrounding page does not travel with it. So the
output block itself carries a line such as:

> Illustrative worked example from figures you entered — simplified, may be
> wrong, not advice. oufe.com

## D. Case dossier header (every page about a named real company)

> This page describes a real situation, and it is our reading of public
> documents. **Our reading may be wrong, and the documents themselves may be
> wrong, incomplete or superseded.** Every figure is cited to what we read and
> when we read it, so you can check it — a citation shows you our source, it does
> not prove we understood it. Situations move; this page can be out of date
> before you reach it. Where we have not verified something we say so, rather
> than estimating. Nothing here is advice.

## E. `/disclaimer` page (linked in the footer legal group)

Longer form covering: what OUFE is and is not; that it is an independent research
publication, **not an incorporated company and not an authorised firm**; the sourcing standard and its limits; that AI assistance is used and what
that means for reliability; that the tools are simplifications that can be wrong;
that analysis of mechanism is not prediction of outcome; that no relationship is
created by reading; and the correction route.

**Build notes when this page is made:** `disclaimer` is already a recognised
legal-nav page name so it is picked up into the footer's legal group
automatically — but the nav label map has entries for privacy and terms and
**not** for disclaimer, so check what label it renders with. Build it the way
migration 182 does: hand-written content, `rebuild_policy='owned'` plus a
permanent component lock, `rendered_html` written in the same migration. Otherwise
a later rebuild rewrites the legal text.

## F. Correction and removal route (a commitment, not just wording)

Inherited from the vetcomparison rails:

> If anything here is wrong, tell us. We will correct it and say that we have.
> If you are the subject of a page and believe it misstates your position, write
> to us and we will review it promptly.

Needs a real monitored address before publication. Currently seeded:
`oufe@contactforsales.com`.

## G. Limitation of liability — paid products only (owner instruction, 2026-07-26)

Owner's instruction: cap our liability on anything paid for at **the refund
only**, i.e. at the fee that was paid for that item.

Proposed clause, for the terms attached to any paid deliverable:

> **Our liability to you.** If something we sell you is wrong, late, or not what
> we described, tell us and we will refund what you paid for it. **That refund is
> the limit of what we owe you.** We are not liable for any loss, cost or
> decision that follows from your use of anything we publish or sell — including
> lost profit, lost opportunity, or the outcome of any transaction — beyond
> refunding the fee for the item concerned.
>
> Nothing here limits our liability for death or personal injury caused by our
> negligence, for fraud or fraudulent misrepresentation, or for anything else
> that cannot lawfully be limited.

### Three things the owner should know about that clause

**1. The cap is only as strong as the refund is real.** A cap at the fee is
clean and easy to explain precisely because it converts every dispute into "you
get your money back". That only holds if the refund route is genuinely honoured,
promptly and without argument. If refunds are slow or contested, the cap stops
being a fair bargain and starts being the thing a court looks at hardest.
Practically: the refund promise in F is now load-bearing, not a courtesy.

**2. The cap does nothing for the free content, which is most of the site.**
Liability capped at "the fee paid" is a cap of zero where no fee was paid — which
sounds strong and is actually just silence. Protection on the free side comes
from a different place entirely: the not-advice posture, the proximate
disclaimers, and the fact that we say plainly we can be wrong. **This is worth
saying because it is easy to approve a liability cap and feel more covered than
you are.** The cap addresses only the paid tier, and the paid tier does not exist
yet.

**3. The carve-out paragraph is not optional decoration.** Certain liabilities
cannot be excluded or limited by contract, and a clause that tries to exclude
everything can be read down or struck out entirely — losing you the cap you
wanted as well as the one you could not have. Keeping the carve-out visible is
what makes the rest of the clause look reasonable.

> `[UNVERIFIED — for the solicitor]` The specific statutory footing for point 3,
> whether the buyer is a consumer or a business, which reasonableness or fairness
> test applies, and whether a fee-capped clause survives it in this context. I
> have drafted to the shape that is conventional for paid research and to our own
> recorded position from idea.uk (information-not-advice, liability capped at the
> fee). **I am not stating the law here** — that is exactly the class of claim
> this workstream refuses to make from memory, and it would be incoherent to
> apply that rule to the site's content and not to its own terms.

### Where this sits relative to the rest

G applies **only** to paid items. A, D and E already establish that nothing on
the site is advice and that we can be wrong, which is what does the work on the
free side. G is not needed until something is actually for sale — but it should
be settled **before** the first sale rather than during it, because a liability
position drafted while a customer is unhappy is drafted badly.

## On staging the disclaimer through the tool

The owner raised putting disclaimers "as stages in the tool". My recommendation is
**one gate, then persistence, then the caveat in the output** — and *not* a
caveat at each step. Repeated interstitials get click-through-blind within about
two uses, and a warning nobody reads is weaker in practice and arguably weaker in
law than a single one they had to act on. The three placements above cover the
realistic paths: you cannot reach the tool without acknowledging (B), the caveat
stays on screen while you use it (already in the build), and it travels with any
result you copy out (C).

If the owner wants a stronger gate later, the sharper version is not more
warnings but a **per-session** acknowledgement rather than a remembered one, so
each visit re-consents.

## What the owner needs to decide

1. **Approve, amend or replace A–F.** They are drafted to be true rather than
   defensive. A solicitor may want them more defensive; my view is that on this
   particular site, plain admission reads as more credible than hedging, and is
   also simply accurate.
2. **The audience question** — the owner asked whether targeting students is
   safer. It is a real strategic question and the answer is not a straight yes:
   written up separately in `PLAN §7`, because it changes more than the wording.
3. **Whether to say "not a company" explicitly.** Recommended: yes, on the about
   page and in the disclaimer. It costs a little authority and buys the thing
   this site is actually selling.
4. **Solicitor review before or after launch.** The idea.uk precedent is a
   ~£200–500 fixed-fee UK review (whose own terms are still flagged as drafts
   pending one). The vetcomparison precedent is that proceeding ahead of review
   is defensible on defined narrowing conditions, provided the decision and its
   conditions are recorded contemporaneously. Either is defensible; drifting into
   launch without choosing is not.
5. **Limitation of liability**, if and when anything is sold. Not needed for the
   free Phase 1. The recorded position from idea.uk is information-not-advice
   plus liability capped at the fee.
