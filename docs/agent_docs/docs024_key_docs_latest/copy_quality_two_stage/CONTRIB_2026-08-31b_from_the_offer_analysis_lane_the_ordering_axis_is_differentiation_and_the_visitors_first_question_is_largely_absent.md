# CONTRIB — from the `vigilant_designer_offer_analysis` lane, 2026-08-31 (evening)

**Reply to the follow-on owner instruction: a hypothesised QUESTION HIERARCHY as ordering input.**
What my analysis can produce, what it would cost, and — the part that matters — **what I checked
that came out against my own hypothesis.**

All figures `[MEASURED 2026-08-31]`, n = **186** ranked `lead_with` points across 32 sites.

---

## 1. The finding: the existing ordering is real, provenance-stamped, and ranked on a SELLER'S axis

`offer_ordering.lead_with[]` already is a ranked ordering with per-point provenance. **But its
ranking principle is DIFFERENTIATION**, and it is unambiguous — share of points marked
`differentiated`, by rank:

| rank | 1 | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|---|
| **% differentiated** | **100** | **100** | 97 | 61 | 31 | 30 |

Monotonic, across 186 points. **The artefact answers "what can we say that competitors cannot?"** —
which is a seller's question. The owner's instruction is about a **buyer's** question: *"what this
tool as a whole is going to get them and how much work it's going to be to get it done."* **Those are
different axes, and mine is the wrong one for hero ordering.** That is the structural answer to his
critique, and it is stronger evidence than the two Fable specimens alone: *"No vendor pays us"* ranked
where it did **because it is highly differentiated**, exactly as the artefact was built to do.

## 2. ⚠ WHAT I EXPECTED TO FIND AND DID NOT — recorded because it changes the recommendation

I expected to show that **independence claims outrank effort claims**, making this a re-ranking job.
**The data does not support that**, and I am not going to force it:

| category (regex proxy) | points | mean rank | in top 2 |
|---|---|---|---|
| answers EFFORT / *"how much work is it"* | **19** | **2.84** | 9 of 19 |
| INDEPENDENCE / *"who pays us"* | **10** | 3.00 | 3 of 10 |
| all points | 186 | 3.51 | 61 |

Effort points rank *slightly better* than independence points, and both beat the average. **On n = 19
and n = 10 with crude regex proxies, that difference is noise. There is no measured inversion.**

**So the real gap is ABSENCE, not mis-ordering: only 19 of 186 points (10%) address effort,
practicality or what it costs the visitor to get started.** The owner's *first* hypothesised doubt is
barely in the corpus at all. Re-ranking cannot surface material that was never derived.

> **This matters for scoping, and it is the opposite of what I told the owner about today's other
> three asks.** Imagery, carousels and the benefit set were all *"the machinery exists and is
> undriven"*. **This one is not.** The question hierarchy is **genuinely new derivation**, and I would
> rather say so now than deliver a re-rank that quietly answers the wrong question.

## 3. What I can produce, and what it costs

**A per-site `question_hierarchy` aspect**, same shape and discipline as `offer_ordering`: an ordered
list of the visitor's likely doubts, each with a `why` citing the strategy field it was derived from,
a `rank`, and — the new part — an `answered_by` pointer to whichever `lead_with` point addresses it,
**or an explicit `unanswered: true`.**

**That last field is the deliverable, not the list.** It turns "we think visitors ask X first" into a
checkable claim: *this site's hero answers doubts 3 and 5 and does not answer doubt 1.* A hierarchy
with no join to the copy is another artefact nobody reads — which is precisely the failure mode this
lane just measured on its own `offer_ordering` (32 sites, provenance-stamped, **zero** writer
consumers).

**Cost, stated honestly:**
- **Derivation is one analyser pass per site**, comparable to `offer_ordering` — same inputs
  (`satisfaction_condition`, `trust_threshold`, `recurring_value`, `money_flow`), so no new upstream.
- ⚠ **But `satisfaction_condition` and friends were written to answer the seller's question**, and
  §2 shows they are thin on effort/practicality. **Expect the first pass to produce hierarchies whose
  top entries are mostly `unanswered`** — that is the correct and useful result, not a failure, and
  it should be the acceptance criterion rather than a surprise.
- **The honest risk: a hypothesised hierarchy is a hypothesis.** It has no reader research behind it.
  Its `why` can cite a strategy field; it cannot cite a visitor. **Ordering input, never evidence** —
  and it must never acquire the authority of a measurement just because it has provenance and a rank.

## 4. On your boundary — agreed, and it is the same boundary I already run

Modelling the hierarchy to ORDER content is licensed; ASSERTING it in served copy stays banned as
presumption (ruling 12). **That is exactly the served/unserved split my `why` fields already sit
on**, and it is why v1 binding the served `point` only was the right call. The hierarchy is an
unserved rationale artefact. ⚠ **It must never be rendered into a prompt as prose** — *"most visitors
first ask X"* in a writer's context window is a demonstration of the presumption shape, and it will
be copied. Structured input only, same rule as `BANNED_REGISTER_v1`.

## 5. On density (ruling 13) — one caution from my side

*"Models compress; we must expand, every time."* Noted, and it interacts with my §1 finding in a way
worth flagging: **the `differentiated` axis rewards compression.** A maximally distinctive claim is
short, absolute and unqualified — *"No vendor pays us"* is the shape differentiation optimises
toward. **If density is now a scored fault, expect it to pull against the ranking principle that
produced this corpus**, and expect that tension to surface as points that score well on one axis and
badly on the other. Worth deciding which wins before the loop runs, rather than discovering it as
disagreement between two seats.

## 6. Not started

No aspect built, no analyser step added, nothing wired. A new `site_specs` aspect plus an analyser
pass is council scope and the owner's call — and on §2's evidence it is **new derivation rather than
a re-rank**, which is a larger ask than the framing implied. This CONTRIB is the view and the costing
he asked for, and it stops there.
