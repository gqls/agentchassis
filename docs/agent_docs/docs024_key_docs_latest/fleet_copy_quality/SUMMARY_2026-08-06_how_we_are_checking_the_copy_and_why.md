# SUMMARY — how we are checking the copy, and why (2026-08-06)

First summary in this workstream. Written to be read aloud.

---

## What we're trying to do

Make the copy the framework writes sound like it was written by an intelligent human
offering a service. Not salesy, not pushy. Explaining the things that deserve
explanation, and saying the quiet non-obvious things plainly rather than dressing them
up. Written for the reader's benefit, because the reader is interested in themselves
and what they are trying to get done. Offering what we can genuinely do — strongly when
that is what's needed, as a hint when that is more in order — without constantly
narrating what we can and can't do.

This is generic. It applies to every site we build, now and later.

## Where we've come from

We already had two attempts at this, and both were rule-based.

The first is the owner's own **style prompt**, reverse-engineered last month from a
passage he hand-edited until he liked it, then refined over three rounds. It lives in
`travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md`, and a lift of
it sits at the bottom of the content writer's prompt as the "house voice".

The second is the **gentle-explanatory voice** developed with the owner yesterday for
loanandmortgagecalculator, and seeded onto loancalculator: ten rules and some worked
examples.

Between the house voice and a site's own spec, a writer can be carrying **thirty-odd
style rules** at once. The instinct all along has been that if the copy isn't right, the
rules need tuning. That instinct is what this workstream is questioning.

The history that matters is the correction trail on the owner's own prompt. Round two
banned a particular construction — "it isn't X, it's Y" — because it spells out a
contrast nobody needed spelled out. Round three found the same instinct wearing
different clothes: "Nothing here is exotic. One choice is…" is the same move, and the
round-two rule didn't catch it because it only knew the literal words.

## What we've done

We stopped guessing and measured, then read.

**The method, and why it is in this order.** A style complaint is easy to agree with and
hard to act on, so the first job is to find out whether the fault is something a machine
can see. If it is, we can detect it, gate it and prove a fix. If it isn't, then no rule
and no automated check will ever hold it, and we should stop trying to build one. So we
took the three most plausible mechanical explanations and tried to confirm each against
live pages — pulling real served copy off six sites, stripping the chrome and the
scripts, and counting.

The point of counting was **to try to prove ourselves right and fail**. A measurement
that can only come out one way is worth nothing, which is why each probe had a
comparison built into it: sentence variance against the known human range, repetition
within a page against repetition between unrelated pages, our copy against copy written
by a different model outside the framework.

All three came back clean.

**Rhythm.** The idea was that "one idea per sentence" produces sentences that are all
the same length. An early two-page sample seemed to show exactly that, and it was wrong
— widening to eight hundred and eighty sentences across six sites gave a spread that
sits comfortably in the normal human range. The monotony was an artefact of a small
sample, and it went away as soon as the sample got honest.

**Disclaiming.** The owner's specific complaint was copy that keeps announcing its own
limits. Searched for as phrases, it is essentially absent — under one occurrence per
two thousand words on every site measured. The copy is also already strongly
reader-facing: on one site "you" outnumbers "we" by eighteen to one.

**Repetition.** If each section is written by a model that cannot see the others, the
sections should end up saying the same things. Measured properly, they don't: overlap
between paragraphs on the same page is no higher than between paragraphs on unrelated
pages.

**Then we read it.** And the fault is there, plainly, on the first pages we opened.

One site says, in three consecutive paragraphs, that it won't tell you whether your idea
will succeed, that it doesn't deliver verdicts, and that it can still be wrong and the
final call is yours. Another opens with "No preferred platforms. No black boxes." and
then, a paragraph later, "We don't have a large org chart or a department for every
service. What we have is…".

That last one is the finding in miniature. **It is the very construction the owner's
prompt banned two rounds ago, live on a site today.** Every search we wrote missed all
of these, because they are a *move*, not a phrase.

## Where we are now

The copy is mechanically well-formed by every proxy we can count, and it still isn't
good. That is not a disappointing result. It is the answer: **the fault does not live
anywhere a rule can reach.**

We think there are two reasons rules keep failing here, and they compound.

The first is that a rule can only name a *form*, and the thing going wrong is an
*instinct*. Ban "isn't/it's" and it reappears as "Nothing here is". Ban that and it
reappears as "We don't have X. What we have is Y." Three spellings of one habit, two
patches, still shipping.

The second is quieter and worse. Some rules are trivially checkable — no em dashes, use
contractions — and some require judgement, like knowing which ideas deserve explaining.
A model under pressure satisfies the checkable ones, because they are unambiguous and
cheap. So every round of tuning adds more checkable rules, and the checkable ones crowd
out the ones we actually care about. **We have been making the problem worse by working
on it.**

One structural fact is worth recording even though our prediction from it was wrong. The
writer works one section at a time and cannot see any other section on the page. We
expected that to cause repetition; measurement says it doesn't. What it does prevent is
judgement across a whole page — deciding where the weight goes, what to expand and what
to state flatly, whether the offer has already been made and should now only be hinted
at. Those are page-level decisions, and nothing in the pipeline is ever holding the page.

## Where we're going

Three things to settle with the owner, in order of confidence.

**Judge instead of ruling.** The one method with a track record on prose here is the one
that produced his own style prompt: write it both ways, hide which is which, and have a
reader pick. That is also how the council gate works, and this estate already trusts a
second opinion over a checklist. The proposal is to add a critic that reads the finished
page as a reader and repairs what is weak, briefed with the goal rather than the rules.
A critic can catch a move. A rule can only catch a form.

**Name the instinct, not its shapes.** Replace the several anti-negation rules with one
that states the underlying judgement: define a thing by what it is, and use a negation
only to correct an expectation the reader actually arrived with. That permits the quiet
useful contrast and forbids the performed one, which is the distinction the current
rules keep failing to draw.

**Give the writer the page.** Not because of repetition — that was refuted — but because
modulation is impossible without it.

And three questions we cannot answer from the code. Whose voice is this: the owner's own
voice, which the style prompt was built from, or a service speaking to a stranger with a
task? What should decide whether an offer is made strongly or hinted at? And is there a
page already on the estate that is close to right — because a real target beats any
amount of specification, and it is exactly how the style prompt was built in the first
place.

The measuring script is kept in this directory. It has found nothing three times, which
is its value: it is now a **negative control**. It cannot tell us the copy is good, but
it will tell us if a future change makes it mechanically worse, and that is worth having
before we start changing prompts.
