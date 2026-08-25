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

---

## 2026-08-24 — the fix is written, and I ended up doing the opposite of what I set out to do

The change is committed and will go live with the next chassis build. It is with the review council
now.

**What I set out to do was repair the number. What I actually shipped removes it from the decision.**
The reason is worth reading, because I nearly got it wrong and the thing that caught me was mundane.

I had measured that the broken number currently changes nothing — no selection anywhere comes out
differently because of it. I then wrote, in three separate places, that "removing **or replacing**"
it was therefore free. That is not what I measured. I measured what happens if it is taken *away*. I
never measured what happens if it is replaced with a *working* number, which is precisely what I then
spent the afternoon building.

When I finally ran my own new query against the live database, the answer was that it changes the
chosen component in **3,246** of the 4,888 situations I could test — against **0** for simply
removing it. The largest change available in this bug, and I had been calling it free. What caught it
was not a test or a review: it was running the query and looking at the numbers on the screen. The
code compiled perfectly throughout.

**And then the mistake turned out to be the useful part.** Being forced to look honestly at what a
*working* version of the number does is what killed the idea. The number rewards components that have
already been used a lot. Make it accurate and you get a loop: chosen → score rises → chosen again. We
have an open complaint — bug 107 — that every site comes out looking like the last one. **A working
"prefer what we already use" rule is an engine for exactly that.** Repairing the number would have
closed this bug while quietly making that one worse.

So: the number is no longer part of choosing a component. It is still calculated honestly and written
to the logs, so we can see it — we just do not decide on it. If we ever want a "prefer proven
components" rule back, it is one line away, and we would be adding it deliberately, knowing what it
costs.

### Something worth your attention about bug 107

While measuring this I found something that may matter more than anything in this bug.

The reason the sameness complaint exists is probably **not** that a scorer keeps picking the same
favourite from a rich field of options. It is that **for almost every slot on a page, there is only
one option to pick.** Of all the section types in the library, only **four** have more than one
candidate at all. The chooser is not choosing badly; most of the time it is not choosing.

That means no amount of adjusting the scoring will fix the sameness. Only having more things to
choose from will. I have written that into bug 107 for whoever picks it up, along with the warning
that adding variety is exactly what would have switched the old rule on, in favour of whatever was
already popular.

### Honest about what is not settled

The independent diagnosis check I commissioned this morning never came back — it is still running
after an hour with nothing in it. So this rests on my own evidence, which I have set out in full,
with the controls that could have proved me wrong. If that check lands and contradicts me, it goes on
the record as a correction.

The council has not returned a verdict yet either. The code is on the shared branch, so if the review
comes back asking for changes I act on it there.

---

## 2026-08-24, after the build — it is live, and the reviewer who worried most turned out to be wrong in a useful way

The new chassis went out at 18:55 and the change is in it. I checked that three separate ways rather
than trusting the deployment, and each check had a control designed to fail if I was fooling myself:
the running binary contains the new query, it no longer contains the old one, and the commit is an
ancestor of what was built.

**One of those controls was worthless and I nearly wrote it down.** To prove "my fix is in the build"
I compared it against another commit that was supposed to *fail* the test — except I picked one that
also predated the build, so both came back "yes" and the test proved nothing. I only noticed because
the answer was too convenient. Replaced it with a commit made after the build, which correctly
reports "no". Worth mentioning because it is the same mistake in a smaller form as the one earlier
today: a check that cannot come out the other way is not a check.

### The reviewer's objection was right in shape and backwards in direction

The most serious concern from the review council was about a second, quieter change I made: which
component counts as the official template for a given kind of section — the one that gets overwritten
and enforced when the platform regenerates it. The reviewer's point was that changing that could
silently re-shape the contract for pages already built against the old one. It is a fair worry; that
family of mistake has bitten this platform before.

I went and checked, and **the change fixes that problem rather than causing it.** The regeneration
machinery decides what to overwrite by the component's *function name*. Across all 117 kinds of
section, the old ordering's answer matched what the machinery actually enforces in 88 cases; the new
ordering matches in 90. Both of the two that changed moved from disagreeing to agreeing.

**And that check turned up something nobody was looking for: 27 of those 117 still disagree.** For a
quarter of section types, the platform tells the component writer to preserve one template's fields
while the storage layer overwrites a different one. That is not caused by this change and it is not
mine to fix, but nobody owns it and it should probably be its own bug.

### What is still missing, and it is the reason I am not closing this

The fix is live but it has not yet *done* anything, because **nothing has been built since the roll.**
The broken counter has stopped moving, which is what I want to see — but that is worthless as evidence
right now, because with no page builds the old code would not have moved it either.

So the honest position is: proven present, not yet proven working. One page build settles it, and the
recipe is written down. I would rather leave it open one more day than record a pass from a check that
could not have failed — which is exactly the trap I fell into twice today and caught twice.

### Where to pick this up

`docs/agent_docs/docs024_key_docs_latest/bugfix_378_usage_count_derived/HANDOFF_2026-08-24_continue_here.md`

---

## 2026-08-25 — it works, and we can close it

Yesterday I said the fix was live but I would not call it done, because nothing had been built since
it shipped, so the broken counter sitting still proved nothing. Overnight the platform did plenty of
work, and the answer is now unambiguous.

Since the fix went live the platform has built **403** component slots across **125** pages, run the
page builder **73** times, run **880** re-renders, and created **5** new components. Through all of
that the old counter has not moved by a single digit — every value identical to the snapshot I took
before the change. Under the old code it would have been ticking up throughout.

Nothing broke either, which is the other half of the check: requests for a brand-new component ran at
**5** overnight against **6** the previous day, and requests for missing section data went from three
to none. If my change had broken the component chooser, both would have spiked.

**So: the bug is closed.** It has moved to the closed folder with the full evidence attached.

### The thing I keep learning

Three times in this piece of work I produced a clean, reassuring result and the control showed it was
worthless. Yesterday I "proved" my fix was in the build by comparing it against a commit that turned
out not to test anything. The day before, a frozen counter meant nothing because nothing had run.
And this morning I checked the logs for errors, found none, then checked whether those logs contained
*any* relevant lines at all — and they did not. The subsystem writes to different pods entirely.

None of those were near-misses in the sense of nearly shipping a bug. They were near-misses of a
different kind: I nearly wrote down three separate "verified" claims that could not have come out any
other way. **The pattern is that the comforting answer is precisely the one that needs the control.**

### One new bug, found sideways

While answering the reviewers I turned up something unrelated and larger, now filed as **bug 388**.
When the platform asks a component to be rebuilt, it tells the writer *"preserve these field names"* —
but the part that then enforces the result picks a **different** component to compare against. Two
different pieces of code look up the same thing in two different ways, and on **27 of 117** section
types they disagree.

It is not new and it is not caused by my change — my change actually reduced the disagreement from 29
to 27. But nobody owned it, so it now has its own file with the evidence and a reproducible query.

### One small thing still owed

The old column is now dead — nothing writes it, nothing reads it. I deliberately left it in the
database so the code could be rolled back safely; now that the code has been live for a day, that
reason has expired. A one-line migration should mark it as superseded so nobody quotes it in six
months. It needs its own review round, so it is a small separate job rather than part of this one.

### Where the record is

`docs/agent_docs/docs024_key_docs_latest/bugfix_378_usage_count_derived/HANDOFF_2026-08-24_continue_here.md`

---

## 2026-08-25, later — the dead column is dealt with

You asked me to finish off the abandoned counter rather than leave it noted as a to-do. Done, and it
turned out to be more worth doing than I had made it sound.

**The column was not merely dead — its description was actively lying.** Every table column in the
database can carry a note explaining what it is, and this one said:

> *"Times this component has been assigned to a page. Incremented by selector. Higher = more
> battle-tested."*

All three of those are false, and that note is exactly what the next person sees when they inspect the
table. So the trap was never the leftover numbers; it was the sentence promising they meant something.
That sentence is now replaced with one saying the column is superseded, that nothing writes or reads
it, that its values both miss real usage and count things that never happened, and where to look
instead. **Applied and checked by reading it back out of the database, not by trusting the command.**

**I also removed the last thing that was still writing to it.** One piece of code still named the
column when creating a component — not counting anything, just setting it to zero. Harmless in itself,
but it meant the column could not be deleted without breaking every component creation.

**And I have written the deletion, tested it, and deliberately not run it.** It has to wait for the
next build, because until the code above ships, deleting the column would break things. The file is
named so the automatic runner will not pick it up, and it carries a safety check that refuses to run
if the old counter has started moving again — which would mean the old code was still live.

**I tested that safety check by trying to trip it**, rather than just watching it pass: I pretended a
component had been counted once more, and it correctly refused with the right message. Nothing was
saved. This is the same discipline as the rest of this lane — a check that has never failed is not
yet a check.

### What is left

**One scheduled action, and it is not a question:** run the deletion after the next chassis build.
The instructions and the safety check are in the file itself. Everything else on this lane is closed.

I have also put the follow-up through the review council, as the rules require for anything that
changes the database.

### One thing I want to be clear about

`agent_definitions.usage_count` is a **different** column with the same name on a different table. It
is alive, it works, and it is deliberately untouched — a previous piece of work fixed that one
properly last month. Anyone reading this should not generalise today's deletion to it.
