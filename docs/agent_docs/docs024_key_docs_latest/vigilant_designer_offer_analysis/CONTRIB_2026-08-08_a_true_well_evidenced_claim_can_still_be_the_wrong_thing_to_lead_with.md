# CONTRIB — a claim can be true, evidenced and still the wrong thing to lead with

**From the `loancalculator_couk` lane, 2026-08-08.** Sent at the owner's direction:
he raised this reviewing live homepage copy and said it belongs with the offer and
benefit analysis rather than with the copy thread, because the copy was only the
symptom. **Not started yet as far as this lane knows** — filed as an input for
whenever it opens.

## The worked example

`loancalculator.co.uk` homepage, opening block, live today:

> **"The UK's independent resource for clear, mathematically rigorous borrowing tools."**
>
> "Our calculators and guides are built to reveal the **true cost of credit**. You can
> model **exact** personal loan repayments, test debt consolidation scenarios, and
> analyze the hidden interest charges within car finance agreements."

**The owner's verdict, in his words:** the points are good, but *"just too strong… it
is positioning us as the authority in accurate tools but that's not usually a top
reader request or need — they already trust our tools, and everyone else's."*

Worth being precise about what he did **not** say. He did not say the claims are
false. On this site they are unusually well-founded — the arithmetic is golden-tested
against the original hand-built calculators, 11 of 11 reproduce exactly, and the tool
components are locked against edits. If any site on the estate has earned the right to
say "mathematically rigorous", it is this one.

**That is the point.** The claim is true, evidenced, defensible — and still the wrong
thing to lead with.

## The generalisable bit, which is why it is your problem and not the copy thread's

The failure is not tone. It is **choosing what to say by what we can defend rather
than by what the reader came for.**

A borrower arriving at a loan calculator is not asking "is this one accurate?" They
assume it is. They assume the next one is too. Accuracy is **table stakes** — a
property whose absence would be disqualifying and whose presence earns nothing. Lead
with it and you spend the most valuable line on the page answering a question nobody
asked, in the register of someone who suspects they might not be believed.

So the axis this suggests for the analysis is not "what can we claim" but something
like:

- **Is this a differentiator or table stakes?** Both may be true; only one is worth
  saying. Table stakes belong deeper down, stated flatly, or not at all.
- **Does the reader already assume it?** If yes, asserting it can actively cost
  you — insisting on something taken for granted reads as either defensive or as a
  hint that it might not be true elsewhere.
- **How strongly is it demanded?** The owner's framing on the same day: *"I can be
  helpful and offer what I can do strongly when needed or softly when a hint is more
  in order."* That says **claim strength should track reader demand, not our
  confidence.** We currently have no representation of demand anywhere — the
  `evidence_base` machinery scores whether a claim is *permitted*, and nothing scores
  whether it is *wanted*.

That last line is the one I would most want the analysis to take seriously. The
platform already has a well-developed apparatus for the permission question: banned
claims, the claims floor at the persistence seam, `evidence_base` gating, the
register. Every one of those is a filter on what we are *allowed* to say. **None of
them can tell us that a permitted claim is the wrong one to make.** A page can pass
every gate we own and still open by answering the wrong question — which is exactly
what this homepage does.

## One caveat on the example, so it is not over-read

This particular paragraph is **original hand-built copy that was never rewritten** —
it is one of three pages blocked by `bugs_open/219` during the voice rollout, so it
is not an output of our own pipeline and should not be read as evidence about the
writer.

And a finding that cuts against the obvious remedy: I re-ran this block through the
live writer twice, once with only the default house voice and once with the site's
current voice spec. **Both dropped "mathematically rigorous" and "exact" unprompted.**
The default was the softer of the two — it produced "these tools calculate the cost of
credit", while the site's own spec kept "the true cost of borrowing". So the
over-claiming here is not a voice defect and a voice fix will not reliably prevent it:
the default already forbids reaching for a word heavier than the fact, and it removed
these without being asked. What no voice can decide is **whether the sentence is worth
having at all**, which is a positioning judgement and belongs to you.

Full trace, the three generated variants and the owner's exchange:
`loancalculator_couk/NOTES_loancalculator_couk.md`, 2026-08-08 section.
