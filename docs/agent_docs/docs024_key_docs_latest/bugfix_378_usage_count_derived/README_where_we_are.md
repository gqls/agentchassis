# Where we are — the component "usage count" that doesn't count usage

Plain-prose running log for the owner. Append-only, newest at the bottom.

---

## 2026-08-24 — picking this up, and what I found when I checked it

There is a column on the component library called `usage_count`. The idea is sensible: when the
platform has several components that could fill the same slot on a page, it should lean towards the
ones that have actually been used a lot, on the grounds that they are proven. The code comment says
exactly that — "battle-tested components score higher".

It does not do that. It never has.

**The short version: the number counts how often one particular route through the code happened to
pick a component. It is not a count of how often the component was used.**

There are three different ways the builder can decide which component fills a slot. It can use the
component the page has already recorded for that slot; it can match the slot's name directly against
a component's name; or, failing both, it can ask a selector to search by section type and score the
candidates. **Only the third route increments the counter.** The other two bind a component to a
real, live page and walk past the counter without touching it. So a component that gets used
constantly by the first two routes reads as never used.

That was the bug as it was filed. When I went to check it was still true, I found it is worse in a
way that changes what we should do about it.

**The counter also counts things that never happened.** The increment fires at the moment the
selector *picks* a candidate — before the builder has decided whether that section is actually
usable, and before anything is written to the page. If the section is then deferred or skipped, the
count still went up. And if the page is planned again tomorrow, it goes up again. So it is not
"uses"; it is "times this route considered it".

The evidence is blunt. The two largest numbers in the entire column belong to components that have
never been on a page at all:

- `testimonials-modern` was created **yesterday**. It reads **12 uses**. It has never been bound to
  a single page — I checked the live table, I checked including rows marked removed, and I checked a
  snapshot taken before it existed.
- `bayesian-ranking-hero-tool_pre_037` reads **20 uses** and has **zero** pages. It is a *backup
  copy* of another component, left switched on, and it is still eligible to be chosen.

So the column is wrong in both directions at once: it misses roughly 1,900 real uses, and its top
two values are not uses at all.

**This matters more than the bug file thought, because the number is used in two places, not one.**
The known one is a small nudge in the scoring — worth a tenth of the score. The one nobody had
noticed is in a different file, where the same column is the *primary sort key* used to decide which
component is the official contract for a section type — the row that gets overwritten and enforced
when a component is regenerated. That is a much heavier decision to be making on a number that means
nothing.

### The one piece of good news, and it is time-limited

I checked whether this bias is currently changing any real choice, and it is not. I simulated the
selector across every combination of section type, site type and page type it could face — nearly
five thousand of them — and removing the bad number changes **no** decision at all.

I did not take that zero at face value, because a zero is only worth something if the test could
have come out the other way. So I re-ran it having artificially given the *worst* candidate in each
contest a full score on that term, and it changed 52 decisions. The test works. The zero is real.

The reason is almost funny: the platform hardly ever has a contest. Only **four** section types have
more than one candidate to choose between — and every candidate in all four has a count of zero. The
twelve components that carry a number are all the only option for their slot, so they were going to
win anyway.

**That gives us a clean window.** Right now we can take this number out of the decision entirely and
know, with a check that could have failed, that not one current choice changes. That window closes
the moment the library grows and contests become normal.

### What I am proposing, in principle

Don't fix the counter — **stop keeping one**. The platform already has a durable, honest record of
which components are actually on pages: the table that binds components to pages. Read that when we
need to know how used something is, and the number can never drift again, because there is nothing
to maintain.

The alternative — keep the counter but remember to increment it on the other routes too — is the
one I want to argue against, and there is a hard number behind that. There are **seven** places in
the codebase that bind a component to a page. A maintained counter is seven places to remember, and
an eighth gets written next month. We have been bitten by exactly this before: a recent count of ten
writers on a related table was correct when written and wrong within a fortnight, because an
eleventh was born.

### One connection worth flagging

You have an open complaint — bug 107 — that every site comes out looking like the last one. This
term is a small engine for exactly that: get picked, score higher, get picked again. It is doing no
harm today only because there are so few contests. But **any fix for the sameness problem works by
adding more candidates per slot — which is precisely what switches this term on**, and it switches
on in favour of whichever component is already the incumbent. So fixing the sameness without fixing
this first would partly undo itself. I have raised it with that lane.

### Where this is now

I have a diagnosis run in flight for an independent check on the mechanism, and a plan being drawn
up. Nothing has been changed yet. Next: get the plan reviewed by the council before committing
anything, because this touches a shared mechanism every site build goes through.

---

## 2026-08-24 later — another lane offered me a better number, and I turned it down

I had flagged this to the lane working bug 357, because their bug is about page rows that name the
wrong component, and my whole plan is to start trusting those rows. They came back quickly and
usefully, and corrected me on three things.

The one that mattered: the bad population is **22 rows, not 9**. The 9 was the figure in the opening
line of their bug file, which their own query had already outgrown. I re-ran their query myself
rather than take the number, and got 22 — all on a single component, the shared `hero`.

Then they offered me something better. As of about ten this morning, page rows carry a stamp proving
which component actually produced the bytes on the page. All 22 of their bad rows are unstamped and
cannot become stamped. So if I counted only stamped rows, their problem would vanish from my numbers
automatically, for free.

**I have declined it, and I think the reason is worth you knowing, because it is the whole point of
this bug.**

The stamp only exists on pages rebuilt since this morning. So counting only stamped rows would mean
counting only components that happen to have been rebuilt recently. A component that has been sitting
happily on twenty pages for two months, untouched because nothing was wrong with it, would read as
unproven — and the components that look best would be the ones that were rebuilt most recently.

That is the same mistake as the one I am here to fix. Today the number measures *which route found
the component*; the stamp version would measure *how recently it was rebuilt*. Both are facts about
the plumbing wearing the costume of a quality judgement. I would have fixed this instance and
reinstalled the identical shape one layer over, and whoever measured it in six weeks would find the
same suspicious perfect correlation I found this week.

The numbers, for the record: counting any real page binding gives me a signal on **108** of 151
components. Counting only stamped ones gives me **39**. Today's broken counter gives **12**.

It also turns out the stamp cannot do the specific job it was offered for. It excludes their 22 bad
rows, but it excludes about 1,750 perfectly good older rows too, because those are unstamped as well.
So it does not separate honest from dishonest — it separates recent from old.

So: I count real page bindings, and I declare their 22 bad rows openly as a known contamination
(about 1% of rows, all on one component) rather than quietly filtering them out. Their own pending
migration retypes those rows anyway, at which point my number corrects itself with no work from me.

The stamp *should* become the better signal once most pages carry it. So I am putting the definition
in one place, so that switch is a one-line change later instead of three.

Nothing has been changed in the platform yet. The independent diagnosis check I commissioned is still
running, and I have not submitted anything to the council.
