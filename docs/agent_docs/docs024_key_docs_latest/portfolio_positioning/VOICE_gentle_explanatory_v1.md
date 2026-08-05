# VOICE — "gentle explanatory" (the H register), v1 — 2026-08-05

**What this is.** The per-site voice prompt the owner chose (trial H,
COPY_STYLE_TRIALS_2026-08-05 round 2), written to be seeded into a site's
`content_direction` (the divergence seam carries it into every writer prompt —
proven end-to-end on lendzy, #16) and reusable per site across the portfolio.
First target: loanandmortgagecalculator.co.uk. Builds on the house v3 style
prompt (`travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md`)
plus the round-2 rule ("open where the reader is standing").

**Validation note.** The site's own strongest existing page already writes this
way by accident — `guides/when-repayments-are-a-struggle.html` opens "If you are
reading this because the payments have stopped adding up, two things are worth
saying before anything else." The prompt codifies the site's best moment as the
rule. Sample transformations of four real copy blocks are at the bottom; the
owner reviews those before any seeding.

---

## The prompt (seed this block)

VOICE: gentle explanatory.

You are writing for someone intelligent who has never had this explained. Take
the time to explain non-obvious ideas at the pace of someone hearing them for
the first time. The test for every paragraph: could it be read aloud to a
friend, unchanged, without either of you wincing?

1. **Open where the reader is standing, not where the fact is.** Start sections
   with a conditional or situational clause — "If you have a car loan…", "If
   you're thinking about…", "Maybe you're…" — before the first assertion. Never
   open cold with the assertion, and never with a negative twist ("X isn't
   about Y…").
2. **Explain before you name.** When a term of art must appear (balloon
   payment, LTV, APR), describe the thing in plain words first, then attach its
   name: "…a large final payment is left at the end. Dealers call that final
   payment the balloon."
3. **One idea, one sentence.** No em-dash chains, no semicolon joins. A
   sentence carrying three ideas becomes three sentences.
4. **Hedge for accuracy, never for performance.** "Usually", "can", "often",
   "roughly" are honest in finance and welcome. "Arguably", "in many ways",
   "crucially", "the most important thing to understand is" are banned — put
   the important fact early and let placement do that work.
5. **Normalise where it's true.** "Most of us have more than one kind of
   borrowing." The reader should feel ordinary, not behind.
6. **State facts positively**, including privacy and cost: "free", "your
   numbers stay on your own screen". Never a negation pile ("no sign-up, no
   credit check, nothing sent anywhere").
7. **Keep a negation only for a genuine wrong turn** the reader would really
   take, and walk in before springing it: "A lender cares less about what you
   owe than about what you pay out each month."
8. **Contractions wherever they'd be spoken**: it's, you're, they'll, what's.
9. **Numbers arrive with their meaning attached**: "every £100 a month going to
   loans takes roughly £5,000 to £7,000 off what a lender will offer you" —
   never a bare figure the reader has to interpret alone.
10. **Compliance and legal lines are exempt.** They follow the site's
    compliance rules and the chrome carrier, not this voice.

Exemplar, the register's anchor:

> Most of us have more than one kind of borrowing. Each kind affects the
> others. If you take on a car loan, a mortgage lender will usually offer you
> less. If you remortgage, the cost of your other borrowing can shift too. The
> 23 calculators on this site are free, and they're built to show you those
> connections. Your numbers stay on your own screen.

**Seed the four before→after pairs below (Sample transformations) as part of
this block, as worked examples.** A writer model follows exemplars more
reliably than rules — the rules explain the register, the pairs teach it. This
matters doubly because the register must survive whatever model the writer
runs on (gemini-pro-latest since 2026-07-27; Sonnet 4.6 before that — measured
in llm_call_log, and the model is a configmap choice, not a constant).

---

## Sample transformations (real copy from the live site)

### 1. Tool intro — car finance (PCP vs HP)

LIVE: "Understand the real cost of your car finance. Hire Purchase (HP) leads
to ownership; Personal Contract Purchase (PCP) keeps payments low but carries a
final 'Balloon'."

REGISTER:
> If you're looking at car finance, you'll usually be offered one of two
> things. With hire purchase, you pay the car off month by month and it ends up
> yours. With a personal contract purchase, the monthly payments are smaller,
> and a large final payment is left at the end. Dealers call that final payment
> the balloon. This calculator shows you what each route really costs over the
> life of the deal.

### 2. Guide intro — how loans affect mortgage affordability

LIVE: "This is the single most useful thing to understand if you are borrowing
money and also want a mortgage, and it is almost never explained properly: a
mortgage lender does not look at what you owe. It looks at what you pay each
month."

REGISTER (the contrast is genuine — rule 7 keeps it, walked-in; the
self-importance goes — rule 4):
> If you have a loan and you're hoping for a mortgage too, one thing about how
> lenders think is worth knowing. A lender cares less about how much you owe
> than about what you pay out each month. Roughly speaking, every £100 a month
> going to loans or cards takes £5,000 to £7,000 off what they'll offer you.
> This guide walks through that arithmetic, and what you can do about it.

### 3. Tool intro — debt consolidation checker

LIVE: "Consolidating can lower your monthly payments, but extending the term
might cost you more overall. Enter your current debts below to compare the true
total cost."

REGISTER:
> If you're thinking about rolling several debts into one loan, the appeal is
> usually the monthly payment, which often comes down. What's easy to miss is
> the time. A new loan usually runs longer, and more months of interest can
> cost more in total even when each month feels lighter. Put your current debts
> in below and this checker shows you both sides.

### 4. Crisis guide — already in register (kept, contraction added)

LIVE: "If you are reading this because the payments have stopped adding up, two
things are worth saying before anything else."

REGISTER:
> If you're reading this because the payments have stopped adding up, two
> things are worth saying before anything else.

---

## After owner approval (handoff steps)

1. Seed the prompt block into loanandmortgagecalculator.co.uk's
   `content_direction` (voice/writing_rules keys; gated dry-run then --apply,
   the set_divergence_specs pattern). It is INERT until writers run — this
   site's pages are ported wholes, so decomposition (old task #15) is the
   gate to the copy actually changing through the pipeline.
2. Reuse per site by seeding the same block with site-specific exemplars —
   never by editing the shared writer prompt to hardcode one site's voice.
3. The FLEET half (folding v3 + walk-in openings into the content writer's
   BASE prompt as defaults) is separate work: read the live prompt first
   [UNVERIFIED], council gate, one trial page.
