# SUMMARY — why the copy reads wrong, and what actually fixes it

**2026-08-12.** Written for the owner, to be read aloud. Covers the copy-quality
problem as it stands after six rewrite rounds on one live page, and the fixes now
applied or proposed. Two threads are working this: this one (the
loanandmortgagecalculator lane, which found the faults on a live site) and the
`copy_quality_two_stage` lane the owner opened this afternoon (which did the prior-art
sweep and corrected two of my claims). Where they disagree, the sweep wins and it is
marked.

---

## What we are trying to do

Get machine-written copy to read like a person wrote it, on a site where a real
customer is trying to work out a real money decision. Not "better prose" as a polish
item — the owner's framing is that AI-slop writing is a product defect, damaging on
any site, and on a consumer-finance site it also decides whether anyone trusts the
numbers.

## Where we have come from

The calculator site's pages were hand-built during its port, then converted so the
framework controls them. The homepage was the first one the framework rewrote, and it
was rewritten six times in two days. Each round was rejected for a different reason,
and **the sequence is the finding** — same model and same prompt every time, only the
instructions changing:

1. **No brief at all** → a 235-word essay. The platform refused to save it (a guard
   that stops a page losing half its text), so nothing shipped.
2. **A brief describing the structure** → right structure, and **every design class
   stripped**: cards, grids and buttons gone, the page a flat run of text. Filed as a
   platform bug; my own comparison missed it because it diffed the words, not the
   markup.
3. **Design vocabulary added to the brief** → the design came back, and the owner
   rejected the copy: a negatively-framed title, consequences fired at the reader in
   clipped sentences, and a paragraph about the website's own inventory.
4. **The three objections written into the brief** → all three fixed, and a new fault:
   the point had migrated to the *end* of its sentence, because I had banned opening
   with something unwelcome and the writer complied by burying it.
5. **The site's own identity spec reframed** → the body register moved properly, the
   four banned phrases went, and the headline invented a fresh fault.
6. **The heading rule replaced** → in flight as this is written.

## What we have done

**Found where the negativity actually comes from, and it was not the model.** The
site's own identity spec said its top differentiator was *"how a car finance payment
reduces what a lender will offer"*. A brief telling the writer to lead with the most
differentiated thing was therefore telling it to lead with a loss. Two more of the
phrases the owner objected to trace to the same spec — including a reader defined *by
exclusion* ("not the single-subject researcher"), which is where "only makes sense"
came from. The strategy spec's tone was literally set to `authoritative`, which on a
subject where the reader is the weaker party pulls straight to
what-they-won't-tell-you.

**Found the same thing again in the heading rule, which is why rounds 4 and 5 fixed
the prose and not the title.** The site's `heading_style` asked in terms for H1s to be
*"a reframing of conventional wisdom"* — tell the reader what they believe is wrong —
and it held the owner-rejected title as a **model example to copy**. The prose rules
were not ignored; they were overruled by a more specific instruction demanding the
opposite.

**Measured the corpus, because one page cannot escape it.** Across all 41 pages,
titles and descriptions ran roughly 16 loss-framed to 11 gain-framed: *How Your Loans
Cut What You Can Borrow*, *The Fees Nobody Quotes You*, *Total Cost, Not Monthly
Payment*, *where an application is most likely to fail*. The writer reads the site it
is writing into, so matching that register is doing its job. **The cure was already on
five pages of the same site** — *what overpaying saves*, *which way round leaves you
better off*, *which pound works harder* — same arithmetic, opposite direction.

**Checked what the industry does, as the owner asked.** Skipton: of six body
sentences, exactly one is loss-framed and it is the mandated regulatory warning.
Nationwide's vocabulary is *borrow more*, *boost for borrowers*, *enhanced
affordability*, *Helping Hand*; its borrowing calculator states existing commitments
*"neutrally as assessment criteria rather than as reductions or losses to the
borrower."* The pattern is not relentless positivity — it is: state the constraint as
**criteria or as a lever**, and spend your one loss sentence where the regulator
requires it.

**Applied, all reversible, prior versions superseded rather than deleted:** the
differentiator reframed to the lever direction (*"the £5,000 to £7,000 of borrowing
power that comes back for every £100 a month cleared"*); tone changed on both specs to
explanatory and non-adversarial; a direction-of-travel rule that also says *put the
point at the front and do not solve negativity by burying it*; a say-it-aloud rule
banning the owner's four phrases by name; the heading rule replaced with
name-the-thing plus worked examples in both directions; and four loss-framed guide
titles reframed, because the homepage links them by title.

**Two things I got wrong, both now corrected in place.** I reported round 2 a success
while it had lost the site's entire design, because my comparison stripped the markup
before diffing it. And I justified building a new mechanism on a query that returned
"nothing consumes the copy auditor's findings" — the value I filtered on had never
existed, so I had proved the absence of a spelling. The other lane caught it: that
auditor has run 34 times, the rewriter has 83 completed jobs, and the applier exists.
The real defect there is small and strange — copy findings are filed under the design
audit's name because of a default in the code, so the work happens and is invisible.

## Where we are now

The body copy reads properly and the design is intact for the third round running. The
title is the last visible fault and the rule that caused it has been replaced. What
the two threads have between them is a diagnosis with the same shape at every level:
**every fault so far was a faithful rendering of an instruction we had written down**,
and the instruction was usually one nobody had read since it was generated at
adoption.

The other lane's sweep sharpens that, and its findings should be believed over my
earlier optimism:

- **A writer copies your examples far more reliably than it follows your rules.** They
  proved it by deleting a rule, leaving its worked examples, and watching the
  behaviour not budge. This is why the heading fix above changed the *examples*, not
  just the wording.
- **An owner decision from 9 August was never delivered.** The gentle-explanatory
  voice was chosen as the fleet default and one rule was to change form; all seven
  writing prompts still carry the old version, including the one that wrote the
  rejected page. Some of what the owner has been reading is a rule he already ordered
  removed.
- **Rule-tuning has already been tried at scale and mostly failed** — six sites, nine
  hundred sentences, three ways of measuring, nothing found, and the owner then found
  a fault in about a minute.
- **The writer only ever sees one section of a page at a time.** My proposed second
  stage read one section too, which would have rebuilt exactly that blind spot.
- **Nothing shares a copy rule between sites.** Two sites produced the same banned
  headline construction within an hour, one of them with an owner-approved fix already
  live in the same database.

## Where we are going

In the order the evidence supports, which is not the order I would have picked
yesterday:

1. **Deliver the voice decision that was already made** — into all seven prompts, and
   change the worked examples at the same time, because the examples are what gets
   obeyed. Cheapest item on the list and it is not a new design.
2. **Fix the attribution default** so copy findings are visible as copy work. Small,
   and until it is done nobody can see whether any of this is working.
3. **Turn the claims checker on for this site.** It is built, live and approved, and
   this site simply never opted in — which is why nothing but a query I wrote by hand
   checked the new £5,000–£7,000 figure that appeared in round 5. That figure turned
   out to be the site's own established fact, stated on four other pages, but we found
   that out by luck rather than by control.
4. **Share the heading examples between sites.** Both rejected lists are now in one
   place on this site; they belong to the fleet, and this is cheaper than any new
   agent.
5. **Then, and only then, the second stage.** With two constraints the sweep
   establishes: it must **read the whole page while writing one section**, and it must
   not be shown the brief, because the brief is where the framing it is looking for
   was introduced.

**Two questions only the owner can answer**, unchanged and now sharper: may an
editorial pass change live copy without a human reading it first — the machinery
currently says no, deliberately — and when the voice change ships, do we make seven
edits or build the one shared place all seven should read from. They have drifted apart
once already without anyone meaning it.

**What is honestly still unproven:** that the reframed specs hold. Round 6 was in
flight when this was written. Every claim above about *where the faults came from* is
measured and cited; the claim that the fixes stick is one round old at best, on one
page, on one site.
