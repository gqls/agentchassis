# ASD-STE100 vs. the house voice — a measured comparison

**Written 2026-08-09.** Input to §4a of
`loancalculator_couk/HANDOFF_2026-08-09b_continue_here.md` — the fleet-wide base-prompt
change, where the owner has decided voice H becomes the default "for now" with named review
triggers. This document exists so that decision has a real alternative to compare against.

**Nothing here is shipped.** No agent definition was edited, no site copy was touched, no
work item was fired. The candidate block is
[`VOICE_CANDIDATE_ste_block.md`](VOICE_CANDIDATE_ste_block.md); the measurement script is
[`ste_audit.py`](ste_audit.py).

## 1. What was compared, and against what

Three things, all read live rather than from a doc:

| | what it is | where it was read from |
|---|---|---|
| **STE** | ASD-STE100 Simplified Technical English, extracted into writer-prompt shape | local skill at `~/Downloads/ste_skill_folder/ste/` (SKILL.md + 2 reference files) |
| **House voice** | the fleet default block, live in **7** agents | `agent_definitions` where `default_config::text ILIKE '%size of the fact%'` |
| **Voice H** | loancalculator.co.uk's own spec, which overrides the house voice | `site_specs` aspect `content_direction`, `is_current`, site `0162cde4…` |

The house voice block was pulled out of `page-content-writer` — the agent that wrote every
page on this site. Reading it needs the nesting-proof path the handoff warns about;
`jsonb_path_query(default_config, 'strict $.**.prompt_template')` finds all of them and the
naive `->'workflow'->'steps'` walk does not.

## 2. The measurement

All 26 live pages, fetched through the size+DOCTYPE guard (26/26 passed, no B2 error blobs),
prose extracted from `<p>/<li>/<h2>/<h3>` with script/style/nav/footer/code removed, then
scored against STE's *mechanically checkable* rules. **542 sentences.**

| STE rule | sentences hit | % of corpus |
|---|---:|---:|
| over 25 words (descriptive cap) | 110 | 20.3% |
| *(over 20 words — procedural cap)* | *202* | *37.3%* |
| contractions | 187 | 34.5% |
| phrasal verbs | 69 | 12.7% |
| banned modals (should/would/could/may/might) | 28 | 5.2% |
| British spelling | 14 | 2.6% |
| -ing used as a verb | 12 | 2.2% |
| "check"/"via" as one-meaning violations | 12 | 2.2% |
| unapproved vocabulary | 10 | 1.8% |
| present perfect / continuous | 1 | 0.2% |
| semicolons · Latin abbreviations · "There is" openers | 0 | 0.0% |
| **at least one violation** | **302** | **55.7%** |
| clean under the checked rules | 240 | 44.3% |

Sentence length: **mean 18.0 words, median 18, p90 29, max 49.**

**[MEASURED]** — the script is committed alongside this file and re-runnable.
**This is a FLOOR, not a ceiling.** The checker cannot see synonym rotation, noun-cluster
length, article omission, or warning ordering. Those are the rules a human would catch and a
regex cannot, and every one of them would add hits.

> **CORRECTED before publication.** The first run returned **60.5%** and I nearly wrote that
> number down. Two false-positive classes were inflating it: `\w*ise[sd]?\b` was scoring
> *promise, rises, advertised, raise, exercise, otherwise* as British spellings, and
> `conduct` was matching inside **"Financial Conduct Authority"** — a proper noun, which STE
> explicitly exempts from substitution. Tightening both to explicit lists moved British
> spelling 25→14 and unapproved vocabulary 12→10, and the headline 60.5%→**55.7%**. The
> caught-it check was reading every hit rather than the count.

> **A second near-miss, worth recording.** I was about to argue that STE's modal ban was
> disqualifying for regulated finance copy, because forcing *"you should check with your
> lender"* into *"you must check"* converts a suggestion into a direction. **The site writes
> "should" exactly 0 times in 542 sentences.** The modals it actually uses are *might* (13),
> *would* (9), *could* (3), *may* (3) — and *might/could/may* all map cleanly onto STE's
> "can". The argument was sound and the premise was false. Only the counterfactual *would*
> ("the interest they **would have made**") has no clean STE form.

## 3. Where STE and what we have already agree

More than expected, and this is the useful half.

- **One idea per sentence.** House voice: *"One idea per sentence. Split anything chaining
  clauses with commas, semicolons or dashes."* Voice H: *"One idea, one sentence: no em-dash
  chains and no semicolon joins."* STE: *"One instruction per sentence."* Same rule, three
  authors.
- **No semicolons.** Both ban them. The site has **0** — already compliant.
- **Condition before command.** STE makes it a rule. Voice H's `paragraph_style` already
  reports it as the site's dominant opening pattern (*"If you stay under this limit"*), so
  the site does this by habit.
- **Start with the fact.** House voice bans the negative-first opening. STE's warning rule —
  command first, risk second — is the same instinct applied to the highest-stakes sentence.
- **Plainness over emphasis.** House voice's *"match the word to the size of the fact"* and
  its cut-list (crucially, seamless, robust, leverage, delve) is a smaller, hand-rolled
  version of STE's substitution table. Voice H's `things_to_avoid` bans superlatives outright.
- **Active voice and imperatives for instructions.** Compatible with voice H's CTA rule
  (*"run the numbers", "try", "use"* — never *"sign up", "get started"*).

## 4. Where they contradict, and which one has to lose

These are not tensions to be managed. They are direct opposites, and adopting STE means
overruling a decision this lane already made deliberately.

| | STE says | we say | live cost |
|---|---|---|---|
| **Contractions** | "Do not use contractions." | Voice H: *"Use contractions wherever they would be spoken: it's, you're, they'll, don't."* House voice says the same. | **187 sentences (34.5%)** |
| **Spelling** | "American English spelling." | British English — a platform convention *and* a UK consumer-finance brand fact. `amortisation`, `licence`, `favour`, `instalments`. | 14 sentences, but the real cost is credibility, not count |
| **Sentence length** | hard cap 20 (procedural) / 25 (descriptive) | Voice H `sentence_style` explicitly wants *"longer explanatory sentences (25-35 words)"* for rhythm | **110 sentences (20.3%)** — the top of H's own intended band is banned outright |
| **Phrasal verbs** | banned | the domain's native register: *pay off, pay down, take out a loan, hand the car back* | **69 sentences (12.7%)** |
| **Rhythm** | uniformity: one term, repeated; consistency is the goal | Voice H: *"never stays in one rhythm for more than two consecutive sentences"*. House voice: *"Vary paragraph length"*, *"Don't repeat a sentence shape"* | not countable, but it is the deeper conflict |
| **Hedging** | "A hedge becomes a fact or a 'can'." | Voice H: *"Hedge for accuracy, never for performance: 'usually', 'can', 'often', 'roughly' are honest in lending and welcome."* | 28 sentences; mostly survivable (see §2 correction) |

The rhythm row is the one that matters most and measures least. STE is built so that a
reader under time pressure, possibly in a second language, cannot misread a procedure.
Repetition is a *feature* there. Voice H is built so a nervous borrower keeps reading and
feels spoken to rather than processed. Those two goals do not reconcile by compromise — you
pick one.

## 5. Three worked rewrites, on real live sentences

### 5a. Descriptive prose — STE costs us something real

**LIVE** (`index.html`): *"If you borrow money, you'll pay it back in monthly instalments.
But that monthly payment isn't just the loan amount divided by the number of months you're
borrowing it for. Instead it's split into two parts: Principal (the actual money you
borrowed) and Interest (the fee the lender charges you for borrowing it)."*

**STE**: *"If you borrow money, you repay it in monthly installments. The monthly payment is
not the loan amount divided by the number of months. The payment has two parts. The
principal is the money you borrowed. The interest is the fee that the lender charges you to
borrow it."*

Every sentence now passes (10/14/5/7/13 words). It is also flatter. The `But… Instead…`
turn is gone, and with it the sense of someone correcting a misconception you might actually
hold. `instalments` became `installments`, which is wrong for a UK site.

### 5b. Procedural — roughly a draw

**LIVE** (`tools/overpayment-calculator.html`): *"This calculator puts the two numbers side
by side: the interest you'd save by overpaying, and any ERC your lender might charge for it.
Enter your loan details below to see whether the saving still comes out ahead once the fee is
accounted for."* (first sentence 24 words — over the 20-word procedural cap)

**STE**: *"This calculator shows two numbers together. The first number is the interest that
you save when you overpay. The second number is the Early Repayment Charge that your lender
can apply. Enter your loan details below. The calculator shows if the saving is more than
the fee."*

Longer in total, shorter per sentence, and it spells out ERC instead of leaning on the
abbreviation. Little lost, little gained.

### 5c. The warning — **STE straightforwardly wins**

**LIVE** (`tools/overpayment-calculator.html`): *"There's a catch worth checking before you
commit to it. Some lenders let you overpay up to 10% of your balance a year for free, but
charge an Early Repayment Charge (ERC) if you go over that."* (second sentence 27 words)

**STE**: *"CAUTION: Examine your loan agreement before you overpay. Many lenders allow an
overpayment of 10% of the balance each year with no fee. If you pay more than 10%, the
lender can apply an Early Repayment Charge (ERC)."*

The live version buries the action inside *"a catch worth checking"* and makes the reader
extract the instruction. The STE version leads with what to do, then gives the number, then
the consequence. **On the sentences where being misread actually costs the reader money,
STE's form is better than ours.** That is the finding I did not expect and it is the one
worth acting on.

## 6. The part of STE worth stealing is its SHAPE, not its rules

Independent of whether we like the rules, the skill is packaged in a way that speaks
directly to the open question in
`SUMMARY_2026-08-09_h_becomes_the_default.md` — *"First, decide how it ships. Seven edits
will drift again."*

STE splits into three files with different load rules:

- `SKILL.md` — the rules, always loaded
- `references/word-substitutions.md` — the dictionary, *"read it before drafting"*
- `references/examples.md` — the exemplars, *"read it when rewriting or when unsure"*

That is exactly the **one shared carrier read at assembly time** option, already worked out,
with a precedent for how the pieces get pulled in on demand rather than pasted seven times.

**But note the trap, using this lane's own evidence.** The lane's clearest finding this week
is *"the example is the instruction; the rule is commentary"* — a writer follows exemplars
more reliably than rules. STE puts its exemplars in the file loaded **last and conditionally**
(*"when unsure"*), and its rules in the file loaded **always**. By our own measurement of how
writers behave, that is backwards. If we borrow the carrier structure, the exemplars belong
in the always-loaded half.

The fourth thing STE has that our prompt entirely lacks is a **post-draft self-check pass** —
nine numbered scans the writer runs over its own output before returning it. Our block has
rules and (elsewhere) exemplars, and no verification step at all. Given the lane's finding
that rules underperform exemplars, a check is a *third* lever, and it is the untried one.

## 7. Recommendation

**Do not adopt STE as the house voice.** It would rewrite 55.7% of a finished, owner-approved
site, mandate American spelling on a UK finance brand, and overturn the contraction rule that
voice H states twice and the house voice states independently. Its uniformity is a virtue in
a maintenance manual and a liability in copy whose job is to keep a worried reader reading.

**Do take four things from it into the §4a submission:**

1. **A hard per-sentence word cap, checkable.** Our specs say *"averaging 18-25 words"*. An
   average is not a check — measured mean is 18.0, sitting at the floor of its own band while
   p90 is 29 and the longest sentence is 49. A per-sentence ceiling is enforceable; an
   average is not.
2. **Classify the section before writing it.** STE's Step 0 (procedural vs descriptive,
   decided per section) is the single cheapest idea here. This site genuinely has both, and
   §5a/§5c show the same ruleset helping one and hurting the other.
3. **The substitution list as a TABLE, not a sentence.** Our cut-list is 14 items buried in
   prose. A two-column lookup is greppable, testable, and extensible without a prompt rewrite.
4. **A post-draft self-check pass.** The untried lever, and the one that does not depend on
   the writer having internalised anything.

**And the narrow, high-value one:** adopt STE's **warning form — command first, then the
risk** — for the site's ⚠️ callouts specifically. Voice H already has a callout rule but says
nothing about ordering. §5c is the evidence, and these are the sentences where being misread
costs the reader money.

## 8. What was NOT done here

- No `agent_definitions` row was read for anything except extraction, and none was written.
- No site copy was changed. loancalculator.co.uk is finished and `index`/`prose-0` is locked.
- **No live arm was run.** The strongest version of this comparison is the one the lane
  already knows how to do: a controlled two-arm test through the framework, of the kind
  recorded in `SUMMARY_2026-08-09` (default house voice vs H on the homepage). A third arm
  with the STE block would be the real test, dispatched via `voiceh_rewrite_v3.sh` against a
  **scratch page, never a locked one**. That is an owner call, not a thing to fire at a
  finished site, so it is named here and not done.
