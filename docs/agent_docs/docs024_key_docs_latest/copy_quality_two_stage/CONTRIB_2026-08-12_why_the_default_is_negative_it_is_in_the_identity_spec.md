# CONTRIB 2026-08-12 — the negativity is not a writing habit, it is the site's IDENTITY SPEC. Found, traced, and fixed at the root on one site

**For the `copy_quality_two_stage` lane** (owner started a thread on it at ~14:45;
this is the LMC lane feeding it, not competing). **Owner's instruction that produced
this:** *"spend some time trying to pin down why everything is negatively framed as a
default and we should change that first."*

Answer, in one line: **the site's own stated proposition was a subtraction, so every
writer told "lead with what is most differentiated" was being told to lead with a
loss.** It was never the model's default. It was ours, written down.

> **⚠ CORRECTED 2026-08-12 (later the same day) — one claim I imported from the PLAN
> was FALSE, and the lane session refuted it with better evidence.** §5 of the PLAN as
> I wrote it said nothing consumes the copy auditor's findings ("0 rows, all-history").
> **My query named `design-audit-agent`, a producer string that has never existed** —
> the real value is `design-audit` — so I proved the absence of a spelling, not the
> absence of a mechanism. `content-quality-auditor` has in fact run **34 times, all
> COMPLETED**, `content_rewrite` holds **83 complete** items, and the copy applier
> exists (`page-build-handler` via `write_audit_findings_action.go`). There is also a
> whole earlier lane, `fleet_copy_quality`, plus two shipped Go scanners
> (`ScanVoiceTells`, `ScanDeployedClaims`) that I did not find. **The real defect is
> ATTRIBUTION** — the auditor stamps `audit_source: "content-quality-audit"` into a
> field that defaults to `"design-audit"`, so copy findings are real, consumed, and
> invisible to any query anyone would write.
>
> **What this does and does not change here.** It does not touch §1–§4 or §6: the trace
> from `identity.key_differentiators[0]` to the rejected sentence, the corpus
> measurement, the industry evidence and the overcorrection finding were all measured
> directly and stand. It does change the *prescription*: stage 2 is a smaller,
> narrower job than "build the missing applier", and anyone reading this CONTRIB for a
> build plan should read the PLAN's corrected §2 first. **What caught it:** the lane
> session ran a proper prior-art sweep instead of trusting my query. The cheap check I
> skipped was `SELECT DISTINCT created_by FROM site_work_items` before filtering on a
> guessed value.

---

## 1. The trace, from spec to slop

`site_specs.identity.key_differentiators[0]`, live on LMC until this afternoon,
written at adoption on 2026-07-31 and never revisited:

> *"Covers the crossing points between unsecured borrowing and a mortgage — **how a
> car finance payment reduces what a lender will offer**, whether to consolidate debt
> into a remortgage, whether the next £1,000 should go on the deposit or clear a
> loan."*

My round-4 page brief said *"lead with what is most useful to the reader, and put the
most useful thing first… the most differentiated thing"*. The writer obeyed both
documents. What came out:

> *"A lender looking at a mortgage application counts what you pay out each month
> against what it will offer you, so a loan instalment or a car finance payment
> reduces that offer before you've even applied."*

**That is differentiator #1 rendered.** The owner's objection to it —
*"slips in the point we're trying to make at the end and adds a negativity"* — is an
objection to the spec, delivered to the copy.

Two of the other four phrases he flagged trace the same way:

| flagged phrase | where it came from |
|---|---|
| *"if the next £1,000 could go either way"* | differentiator #1's *"whether the next £1,000 should go on the deposit or clear a loan"* — the writer had the choice but not the words for it, and reached for a vague idiom |
| *"the questions that only make sense once both sides are in view"* | `identity.target_audience`, which defines the reader **by exclusion**: *"Not the single-subject researcher… is served better by a single-subject site"* |

Two contributing settings, both also spec:

- `strategy.tone: "authoritative"` + `site_type: "authority-portal"` — "authoritative"
  on a subject where the reader is the weaker party pulls straight to the
  reveal-what-they-don't-tell-you register. The sibling lane found the same thing and
  wrote the antidote as prose: *"Lenders are simply the people who decide, and their
  criteria can be explained calmly."*
- `identity.services[2]` described the guides as covering *"…fixed vs variable rates,
  and **repayment struggles**"* — the problem, not the help.

## 2. The corpus then reinforces it, which is why one page could not escape

The build path runs `load_existing_content`, so a writer sees the site it is writing
into. Measured over all 41 pages' `title + meta_description`:

```
pages 41 | loss/error/concealment markers 16 | gain markers 11
```

`[MEASURED 2026-08-12; crude — a page can match both, and the marker lists are mine.
Treat the direction as the finding, not the ratio.]` The titles carry it plainly:
*How Your Loans **Cut** What You Can Borrow* · *The Fees **Nobody Quotes** You* ·
*Total Cost, **Not** Monthly Payment* ("the single most **expensive habit** in
personal finance") · *When Repayments Become a **Struggle**** · *Car Finance Return
**Damage** Checker* · fact-finder: *"where an application is **most likely to fail**"*
· fixed-vs-variable: *"that is often exactly the **wrong** way round"*.

**So asking one page to read positively while forty pages say cut / struggle / nobody
quotes / most likely to fail puts the writer in a conflict it resolves toward the
corpus** — because matching the site *is* its job.

**And the site already contains the cure, on five pages.** Same arithmetic, opposite
direction: *"What overpaying an unsecured loan **saves** in interest and time — usually
a much higher **return** than overpaying a mortgage"*; *"see which way round **leaves
you better off**"*; *"Which pound **works harder**"*; *"How Much **Could You Borrow**?"*.
Proof the register is reachable inside this subject and this site, not an import.

## 3. Why the subject invites it, stated so it can be designed against

This site's differentiated fact **is a negative quantity**: existing payments reduce
borrowing power by ~£5,000–£7,000 per £100/month. There is no way to lead with the
site's unique value and stay neutral — the number itself has a direction. Unless you
flip it, which costs nothing and changes everything:

> *your car finance cuts what a lender will offer* → **clearing £100 a month gives
> back £5,000 to £7,000 of borrowing power**

Identical arithmetic. One is a subtraction happening to the reader; the other is a
lever the reader controls. **This is the general rule for any site whose subject is a
constraint** — affordability, eligibility, tax, penalties, deadlines. The fleet has
several.

## 4. What the industry actually does (owner asked; three sources, 2026-08-12)

- **Skipton Building Society** (`/mortgages`): of six body sentences pulled, exactly
  **one** is loss-framed — *"If you don't keep up with your mortgage payments you could
  lose your home"* — and it is the mandated regulatory warning, standing apart from
  the copy. The explanatory voice is enabling throughout.
- **Nationwide** — its whole vocabulary is capability-increasing: *"borrow more"*,
  *"Boost for borrowers"*, *"enhanced affordability"*, *"increases support"*,
  *"Helping Hand"*.
- **Nationwide's borrowing calculator**, on the same mechanism our homepage was
  writing about, reads: *"Final borrowing would be subject to our lending limits,
  affordability checks, and property value"* — assessed as *"framed neutrally as
  assessment criteria rather than as reductions or losses to the borrower."*

**The pattern is consistent and it is not "be relentlessly positive":** the industry
states the constraint as *criteria* or as a *lever*, and spends its one loss sentence
where the regulator requires it. We had made loss the theme.

## 5. What was changed (LMC only, config, reversible, verified in-transaction)

Prior rows superseded not deleted (`is_current=false, superseded_at=now()`), so every
previous value is recoverable.

1. **`identity.key_differentiators[0]`** → the lever direction, same facts:
   *"…what clearing a loan, keeping it, or moving it into a remortgage each do to what
   a lender will offer — including the roughly £5,000 to £7,000 of borrowing power
   that comes back for every £100 a month of payments cleared, and which of the
   deposit or the debt the next £1,000 does more for."*
2. **`identity.tone`** → adds *"Calm and non-adversarial: a lender is simply the party
   that decides, and its criteria are explained rather than exposed. The reader is
   someone working something out, not someone being warned."*
3. **`identity.services[2]`** → *"repayment struggles"* → *"what to do if repayments
   get hard."*
4. **`strategy.tone`** → `"authoritative"` → *"Plain and practical; explanatory rather
   than adversarial or authoritative."*
5. **`content_direction.framing_direction`** (new) — the direction-of-travel rule,
   including the industry evidence and: *"PUT THE POINT AT THE FRONT of the sentence…
   Do not solve negativity by hedging, deferring or burying — that is a worse failure
   than stating the thing plainly."*
6. **`content_direction.say_it_aloud`** (new) — the read-aloud test, with the owner's
   four flagged phrases banned **by name**.
7. **Three `things_to_avoid`** — loss/error/concealment as a page's THEME; headings
   built on negation or aphorism; burying the point to avoid an unwelcome opening.

**`formatted` regenerated by `datahelpers.FormatContentDirection` itself** (a scratch
Go program with a `replace` directive, so no repo surface), never by hand — the writer
reads `formatted`, so a hand-edit would have left `data` and `formatted` disagreeing
and the new rules inert. The verify block asserts the new text is present in
`formatted`, not just in `data`, and would have aborted if the regeneration had failed.

## 6. The overcorrection the owner spotted, and its lesson for stage 2

He said *"I think some of these changes are overcorrecting"*, and he is right about the
mechanism. My round-4 brief banned *"any framing where the reader is losing… as the
FIRST thing they read"*. The writer complied by **moving the point to the end of the
sentence** — producing exactly the *"slips in the point at the end"* he then objected
to. **A prohibition displaces a problem rather than solving it**; it also keeps the
banned concept salient. The round-5 brief replaces the ban with the positive
instruction (lead with the lever, mechanism after).

**For the lane's design this is the sharpest evidence yet that stage 2 cannot be a
list of bans.** Stage 1 wrote what the specs said. Round 4 banned the symptom and got
a new symptom. What actually moved the register was **changing what the site says it
IS** — which no rewrite pass, however good, could have reached, because it would have
been rewriting a faithful rendering of a wrong premise.

**So a third question for the lane, alongside the PLAN's three:** does stage 2 get to
raise a finding *against the site spec*, or only against the prose? On this evidence
the highest-value finding it could ever produce is *"this page reads badly because the
identity spec is framed as a loss"* — and that is not a copy edit.

## 7. Verification status

Round 5 is running against the reframed specs at the time of writing; the result and
whether the four banned phrases recur will be appended to
`loanandmortgagecalculator_couk/NOTES` and this lane's NOTES. **The claim that the
spec change fixes the register is UNPROVEN until then** — what is proven is the trace
in §1 and the corpus measurement in §2.

— LMC lane, 2026-08-12. Evidence: live `site_specs` rows (superseded copies hold the
prior values), the 41-page corpus query, three fetched industry pages, and the four
rounds recorded in `NOTES_two_stage_copy.md`.
