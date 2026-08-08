# SUMMARY 2026-08-08 — why this is hard, and what we have learned about rules

Second summary in this workstream. The first, on the 6th, reported that we could not
find the fault by measuring. This one reports what happened when we tried to fix it, and
it is mostly a record of things that did not work. Written to be read aloud.

---

## What we're trying to do

Have the framework write copy that reads as though an intelligent person wrote it, for a
reader with something they are trying to get done. Not salesy. Explaining what deserves
explanation and stating the quiet non-obvious things plainly. Offering what we can
genuinely do, strongly where that helps and as a hint where that is better, without
narrating our own limits.

The mechanism we have for steering this is a prompt: a house style shared across the
fleet, plus a per-site voice specification of twenty-odd written rules.

## Where we've come from

We tried the obvious thing first, twice. In July the owner's own writing style was
reverse-engineered into a prompt and refined over three rounds. In August a "gentle
explanatory" voice was developed with him and seeded onto two sites. Both are rule sets.
Between the house style and a site's own spec, a writer can be carrying thirty-odd rules
at once.

Every time the copy has been wrong, the instinct has been to tune the rules. This week
was the first time we tested that instinct instead of following it.

## What we've done

Three measurements, then three fixes, then a controlled removal. Almost all of it failed,
and the failures were informative in a way the successes would not have been.

We measured first, because if a fault is machine-visible we can detect it and gate it,
and if it is not then no rule will ever hold it. We checked whether the sentences were
monotonous, whether the copy constantly announced its limits, and whether sections
repeated each other. All three came back clean, across six sites and nearly nine hundred
sentences. Our copy is mechanically well-formed by every proxy we can count and it still
was not good.

Then the owner read a page and found the fault in about a minute. He has since found
three more, and none of them was reachable by anything we can measure or any rule we
could write.

Finally we ran the experiment properly: removed the rules that seemed to be causing the
mechanical feel, and rebuilt three pages. The tic we were trying to remove survived
untouched, and the removal introduced a new fault on a live page.

## Where we are now

We have a set of things that are now well enough evidenced to be treated as working
knowledge. They are all about the behaviour of rules, and most of them are uncomfortable.

**A writer model follows exemplars more reliably than rules.** This was written down in
our own voice document as an argument for including worked examples, and this week it bit
us from the other side: we deleted the rule that produced an unwanted opening and left the
three worked examples that demonstrate it, and the behaviour continued exactly as before.
**A rule and its example are not two statements of one instruction. The example is the
instruction; the rule is commentary.** Any change to a rule that does not also change its
examples is theatre.

**A rule can only name a form. What goes wrong is an instinct.** The owner's style prompt
banned one construction in round two, caught the same instinct wearing different grammar
in round three, and today the same instinct is live on a site in a third costume. Three
spellings, two patches, still shipping.

**Prescriptions become tics; prohibitions do not.** Tell a writer what to avoid and a
hundred sections still open a hundred ways. Tell it how to open and a hundred sections
open identically. Prescriptions about structure are templates wearing a rule's clothes,
and a template everywhere is exactly what "mechanical" means.

**Rules that are cheap to check crowd out rules that need judgement.** "No em dashes" is
unambiguous and costs nothing to satisfy. "Explain what deserves explanation" is neither.
Under pressure a model satisfies the first kind, so every round of tuning adds more of
them and buries the ones we actually care about.

**Rules do not only fail to catch faults. They generate them.** Of the faults the owner
found this week, two were produced by rules working correctly: a privacy sentence in an
opening about money, because a rule specified that phrasing; and an over-strong claim that
survived our site voice but was removed by the plainer default. Adding rules to fix copy
carries a real chance of introducing the next fault, and the fault arrives pre-authorised,
so nothing downstream questions it.

**A rule that exists may not fire.** Our spec already exempts legal pages from the site
voice. The legal page has the site voice's tic in its first sentence. So a new rule to fix
that page would be a second copy of a rule already being ignored.

**Some rules are load-bearing by accident.** A paragraph-length cap looked like pure
style. It was the only thing preventing independently-written sections from wandering into
each other's territory, because no writer on a page can see any other section. We relaxed
it and a page immediately said the same thing twice. **Before removing a constraint, ask
what it is incidentally holding up.**

And one that is not about rules at all, but is the reason none of them is sufficient. **A
claim can be true, well-evidenced, correctly phrased and still be the wrong thing to say.**
The homepage led on the accuracy of its calculators. The claim is unusually well-founded
here. It is still wrong to lead with, because a borrower already assumes the arithmetic
works and came to find out what a loan will cost them. Nothing in our platform represents
what a reader wants — every mechanism we own filters what we are permitted to say.

## Where we're going

Immediately: fix the exemplars, which is the same experiment with the variable actually
changed, and cheap. Then rebuild the same three pages and read them.

Then the two things this week has pointed at repeatedly. Give the writer the page rather
than one section at a time, because several of these faults are page-level judgements that
no section-scoped call can make. And find some way to represent what the reader came for,
because that is the axis the owner keeps judging on and the one we cannot currently see.

The measuring script stays where it is. It has found nothing four times now, which is its
job: it cannot tell us the copy is good, only that a change has made it mechanically
worse. That is worth having before each of these experiments, not after.
