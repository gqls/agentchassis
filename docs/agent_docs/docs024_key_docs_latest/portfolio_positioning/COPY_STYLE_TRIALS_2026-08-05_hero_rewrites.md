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

## Round 2 (2026-08-05, owner feedback)

Owner chose **C**, and named the residual defect in A: it opens mid-context ("A
car loan changes…") — assumptive. Wanted the reader walked in: "If you have a car
loan, then that will affect…" — explanatory, conditional, not assumptive. Four
more drafted around A/C with that gentler entry:

### Trial E — conditional open, plainest
> **Loan and mortgage calculators, together**
> If you have a car loan, it can change what a mortgage lender is willing to
> offer you. And if you remortgage, that can change what your other borrowing
> really costs. Everything here is built around those connections. There are 23
> free calculators to try, covering loans and mortgages across the UK. Pick
> whichever fits your question, and your numbers stay on your own screen.

### Trial F — question-led walk-in
> **Work out what your borrowing really costs**
> Maybe you're weighing up a car loan while saving for a mortgage. Or you're
> thinking about remortgaging and wondering what it means for the rest of your
> debt. Either way, the answer depends on how your borrowing fits together.
> These 23 free calculators are built to show you exactly that. They're instant,
> and your numbers stay on your own screen.

### Trial G — C's register with the conditional entry (recommended)
> **Work out the whole picture**
> If you're juggling a car loan, a mortgage, or both, each one changes what the
> other costs you. Lenders look at your borrowing as one story, so it helps to
> work it out that way too. Use any of the 23 free calculators here to see what
> you could borrow, what it really costs, and how one debt affects another.
> They're instant, and your numbers stay on your own screen.

### Trial H — fully explanatory, softest pace
> **See how your borrowing fits together**
> Most of us have more than one kind of borrowing. Each kind affects the others.
> If you take on a car loan, a mortgage lender will usually offer you less. If
> you remortgage, the cost of your other borrowing can shift too. The 23
> calculators on this site are free, and they're built to show you those
> connections. Your numbers stay on your own screen.

### The generalisable rule this round surfaced (for the writer-prompt change)

**"Open where the reader is standing, not where the fact is."** A conditional or
situational clause ("If you have a car loan…", "Maybe you're weighing up…")
before the first assertion. It is the positive counterpart to v3 rule 3: rule 3
bans the false-twist opening; this prescribes the walk-in opening. Candidate
rule 14 for a v4 of the style prompt — to be folded into the content-writer
prompt work the handoff carries.
