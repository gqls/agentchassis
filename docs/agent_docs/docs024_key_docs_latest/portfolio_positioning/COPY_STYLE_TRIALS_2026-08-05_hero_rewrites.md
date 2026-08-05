# Copy style trials — loanandmortgagecalculator.co.uk hero, 2026-08-05

**Why this doc exists.** The owner flagged the live hero as LLM-styled (negative
framing etc.) and asked for read-aloud alternatives before changing the content
writer prompt. Provenance was checked FIRST (owner's request): the copy did NOT go
through the framework flow — the site was hand-built by a CLI session on 07-31 and
ADOPTED (41 pages, each one whole-document `ported-page` component; the manual
`content_direction` has never been consumed because no writer ever ran). So the
copy quality reflects an unframed LLM writing freehand, and a writer-prompt fix
only reaches these pages after decomposition (task: decompose guides → flip to
generic — the 08-02 handoff's #15).

**The instrument.** `travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md`
— built from the owner's own critique rounds; rule 3 (don't open a fact with a
negative frame, in ANY grammar) is precisely the flagged tic. All trials below
follow v3.

## The live copy (for comparison)

> **Loans and mortgages, in one place**
> Your borrowing does not come in separate boxes. Neither do these calculators.
> 23 free UK calculators for loans and mortgages — because a car loan changes what
> a mortgage lender will offer you, and a remortgage changes what your other debt
> really costs. No sign-up, no credit check, nothing sent anywhere.

Tics, named by rule: opens with a negative-frame twist and mirrors it (r3, r13);
em-dash chain carrying three ideas (r1, r2); closes on a triple negation ("No…, no…,
nothing…") — the privacy fact stated as what it isn't (r3).

## Trial A — spoken plain ("say it to a friend")

> **Loan and mortgage calculators, together**
> A car loan changes what a mortgage lender will offer you. A remortgage changes
> what your other debt costs. The sums belong together, so the calculators live
> together. There are 23 of them here, all free, all UK. Pick one and start
> typing. Everything runs in your browser and stays there.

What it does: leads with the two facts that justify the site; contractions-ready
rhythm; the privacy claim becomes a positive fact ("runs in your browser and stays
there").

## Trial B — broadsheet compact

> **Twenty-three calculators. One picture of your borrowing.**
> Loans and mortgages pull on each other. A new car loan shrinks what a mortgage
> lender will offer. A remortgage re-prices the rest of your debt. The calculators
> here work the numbers together. All 23 are free and run in your browser.

What it does: shortest; the heading carries the count; every sentence one idea.

## Trial C — warm second person (service register)

> **Work out the whole picture**
> Your car loan and your mortgage are one story to a lender, so treat them as one
> story here. Use any of the 23 calculators to see what you could borrow, what it
> really costs, and how one debt changes another. They're free, they're instant,
> and your numbers stay on your own screen.

What it does: closest to a human adviser's voice; strongest read-aloud flow; risks
being slightly softer than a tools site wants.

## Trial D — staccato open, one landing beat

> **Every borrowing sum, one place**
> Car loans. Mortgages. Remortgages. Twenty-three free UK calculators, built
> around one fact: each debt changes the others. See what a lender will offer,
> what a loan really costs, and what a remortgage does to both. It all runs in
> your browser.

What it does: rule-10 enumeration opener; lands one beat ("each debt changes the
others"); most distinctive, most likely to age into a tic if used everywhere.

## What happens after the owner picks

1. **Fleet lever:** fold the v3 rules (or the winning register) into the content
   writer's base prompt — FIRST read the live prompt (`agent_definitions`,
   page-content-writer / content_writer) to see what style guidance it carries
   today [UNVERIFIED — nobody has read it in this thread]. Platform-config change:
   measure, council gate, one trial page, owner review, then rollout.
2. **Per-site lever:** the winning register also becomes a voice block in each
   site's `content_direction` (the divergence seam already carries it to prompts —
   proven end-to-end on lendzy, #16).
3. **This site:** ported pages don't pass through the writer. Either hand-rewrite
   the hero now (fast, stays outside the framework) or decompose + regenerate
   through the improved writer (the doctrine path; task #15 decompose-first).
