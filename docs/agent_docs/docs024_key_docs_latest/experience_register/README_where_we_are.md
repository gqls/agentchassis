# Where we are — experience register

Append-only. Owner's plain-prose log; add below, never rewrite above.

## 2026-07-24 — the idea, two rounds of discussion, design decided

We want a library of small user experiences. Today, when a site uses our approved carousel,
nobody has written down what clicking a card should do — every build makes it up again. The
library will record things like: "read more" expands the summary right there; clicking the
card takes you to that item's own page; that page offers related reading and tools. Each
entry is a base plan. A site takes a copy and fills in the blanks — which real page the card
should lead to — and because the blanks are filled in explicitly, we can check the links
automatically instead of guessing what the builder meant.

We searched everything related first. The short version: nothing like this exists yet, and
three different efforts have each hit the missing piece from their own side — the link
checkers can tell a link goes *somewhere real* but not whether it goes *where it should*; the
tool documentation work found that design intent "lived only in a conversation that's gone";
and the experience loop writes each site's journey plan from scratch every time, which is why
only one exists.

Decisions made today, after two rounds:
- Entries live in the database: a register table for the machine-readable part (so the
  planner can search it), plus each entry gets its own travelling document for the story of
  why it is the way it is — exactly how tools already work.
- The site planner learns about experiences through the same kind of instruction the roadmap
  already uses: "these pages must exist because these experiences need them."
- The naming system is ours, loosely borrowed from the UX industry's pattern names, and we
  only add entries by harvesting things that already work — starting with tens, not
  authoring thousands.
- Every entry is approved individually by a review council, and its acceptance tests are
  formal: checked when written, checked again when a site fills in the blanks, and run
  against the live site.
- vonc.com finishes its full product first (the AI debate gauntlet, real backend included) —
  we decided not to ship a cut-down static version first. Its provocations journey (teaser →
  full article → related links) becomes the register's first harvested entry, taken from a
  working site rather than invented.

To be clear about what happens next and who moves: nothing builds itself. The vonc backend
is waiting on its code generator to produce a valid result (third attempt running), then on
the owner to merge the result and do four pieces of infrastructure (the subdomain, the
bastion machine, the network peering, the tunnel). Only after that does the experience plan
get re-run and the site rebuilt. Separately, the register's own build (the table, the
planner hook, the validation) is designed and written up but waits for the owner's go.

One genuine bug was found during the research and filed (064): a previous change let the
database accept a new kind of travelling document that the code still refuses to read or
write — so those documents exist but are unreachable. Our build will fix that in passing,
and it taught us the checklist of every place that must change together.

---

**2026-07-26 — the gate lifted, and the first four experiences are on paper.**

The thing we were waiting for happened: the vonc gauntlet went live end to end. A visitor can
read today's provocation, start a real twenty-minute round against a live AI opponent, file a
position, get a written counter-argument back, defend it, and receive a judged verdict — all
against a real backend on a real domain. That was the condition we set for starting the
harvest, so the harvest started today.

I checked the site myself rather than taking the other session's word for it: both pages
answer 200, the data feed the pages read is byte-for-byte the one in the repo, and the
JavaScript that drives the archive is genuinely in the file the live site serves. The
journeys were verified in a real browser by the session that built them — 72 checks of 73,
the one failure being the backend occasionally erroring, which is filed separately as its own
bug and does not touch anything we took.

Four experiences came out of it. Two are about single components: *a list built from a data
feed, where whether a row is clickable is decided by the data and not by the template*, and
*a call-to-action whose words and whose destination live in the same record*. Two are
journeys: *click a teaser and the full piece opens in place at a web address you can share*,
and *a timed exchange with a remote engine where nothing on screen changes unless the engine
really answered*.

The most interesting number of the day: all four of those components have been used exactly
once each, on this one site. The components don't repeat — but the rules above already do. One
of them we harvested from two different components on the same site. That is the whole
argument for the register, and it is now evidence rather than a hunch.

Harvesting also corrected our own design in ten places, which is exactly why we insisted on
harvesting from something real instead of authoring a catalogue from a textbook. Three worth
telling you about. First, our design could only describe things a visitor can click; the most
valuable rule in the live code is the opposite — *a row with nothing behind it must not be
clickable at all* — and we had nowhere to put it. Second, we had named this first pattern
"teaser → detail → related links", and the real thing has no "related links" step; we had
invented a leg that nobody built. Third, and most awkward: four of the rules that make these
experiences worth having **cannot be tested by our own testing machinery at all**. It cannot
ask whether a link is dead, it cannot follow a link to see whether the promised page exists,
it cannot check that a region is empty, and it cannot wait for anything slower than a third of
a second — while the AI in the gauntlet takes eight to twenty-three seconds to reply. That
last one means the approved plan for the gauntlet contains two tests that would fail a
perfectly good page. The fix is to the testing machinery, never to the page — making the page
paint fake text would make those tests pass with the engine switched off.

So where that leaves us. The design is now written from something that exists rather than
something imagined, and it is better for it. The build of the register itself — the table, the
planner hook, the validation — is still waiting on your go, and it is the next thing. One
piece of good news there: a bug we filed on Friday (064) has been fixed by another session and
is now live in the running system, which makes our build a little smaller than it was.

---

**2026-07-26 (later) — the brochure components, and a rule we keep re-inventing.**

You chose to harvest the brochure set before building the register itself, and it earned its
place. Five more experiences, taken from the five components on fundamentallyai.com — the card
carousel, the hover-reveal image grid, the swipeable text track, the counting statistics band,
and the illustrated statement block. Each one checked on its own live page rather than in the
repository.

The important thing that came out of it is not a component at all. It is a rule that six
different pieces of code implement independently, in five different ways: **a control that
cannot do anything must not be presented as a control.** The archive strips the link from a row
with no article behind it. The call-to-action isn't rendered when there is nowhere to go. The
carousel hides its arrows when there is only one card, and never draws a pause button when
there is nothing rotating. Two templates simply don't emit a link when no address was supplied.
Five authors, none of them talking to each other, all arriving at the same rule.

That changes the design. Rules like this should be written down once and *referred to* by each
experience, not copied into each one — copying is exactly how you end up with six slightly
different versions of the same idea, which is the problem we are trying to solve rather than a
way to solve it. So the register gains a short list of named invariants alongside the
experiences themselves.

There is also a live problem, found by taking one of these experiences seriously. The carousel's
rule says a card takes you to a real page. So I checked where the four cards on the capabilities
page actually point — and all four are dead twice over: the address they use returns "not
found", and the place on the page they are aiming at doesn't exist either. The hover-grid cards
on another page are the same. This is already a known bug belonging to another thread, so I
added the evidence there rather than starting a competing fix. But note what it shows: the site
passed its own link check on Friday with "43 targets, none broken", because that check tidies
the address before testing it and ignores the part after the "#". The experience says what the
card was *supposed* to reach; that is what turns "is this link right?" from someone's judgement
into a check that runs.

Two things I got wrong today and caught before they went anywhere, since they are the useful
part of a log: I nearly reported an accessibility fault in someone else's component from a
careless search (the component is correct — I searched for the wrong thing), and I nearly
carried forward my own note that the statistics band has no behaviour, when in fact it has the
most careful behaviour of the five.

Next is the register's own build, which is what you green-lit the sequence for.

---

**2026-07-27 — the register exists in the database now, not just on paper.**

The thing we were waiting for happened on its own. Yesterday's build of the platform went out
this morning, and it contained the small code change the register needed. That was the condition
we had set for ourselves: the code has to go out *before* the database change, because doing it
the other way round recreates a bug we spent real time closing last week — a database that
accepts something every piece of code still refuses.

So this afternoon the database change went in. The register now has its three tables: the library
of experiences itself, the per-site copies that bind an experience to real pages, and the small
set of rules that hold across many experiences at once. Two of those rules are seeded, the ones
harvest actually found evidence for. The library itself is deliberately **empty** — entries have
to be written through the validating path, so that the first things in the register are things
the register's own rules accepted. Seeding it by hand would have been quicker and would have
proved nothing.

I checked it rather than assuming it. The interesting check is the negative one: it is easy to
confirm that the new value is now allowed, and that tells you almost nothing, because a
constraint that was accidentally *deleted* would also accept it. So I also confirmed a value that
should still be refused is still refused. It is. Both checks ran inside a transaction that I
rolled back, so nothing was left behind.

Two things I got wrong and fixed, both worth saying because they are the same mistake in
different clothes. First, the database change went out yesterday without its safety block — the
bit that checks its own work and undoes everything if the work is incomplete. Our own written
convention requires one; I had skipped it. Worse, I had been careful in exactly this way about
the *code* the day before, deliberately breaking each check to watch it fail. I just hadn't held
the database change to the same standard. Added it, then broke it on purpose to watch it fail
before trusting it — which it did, naming precisely the half-finished state it exists to prevent.

Second, applying it turned out to be a small trap. The tool that applies these changes applies
*everything* that is waiting, and twenty were waiting — nineteen belonging to other people
working on this system, some of them parked deliberately. One of them says, in its own text, that
the situation it was written for no longer exists. So I applied only mine, by hand, and then
registered it properly so it doesn't get applied a second time later.

**The review we were waiting on never happened.** Not rejected, not sent back — it simply died.
It got stuck four hours into one reviewer's turn and was cleaned up by a watchdog. That is a
third possible outcome I hadn't allowed for, and it's the sort that quietly looks like patience:
you keep checking and it keeps saying "in progress". It wasn't only mine — eight runs died that
way on the same day, and none on any other day that week. I have **not** claimed to know why. The
day had several platform restarts, which is a plausible culprit, but plausible is not diagnosed,
and I'd rather write down the question than invent an answer. If it happens on a second day, that
becomes a bug worth filing and the query to prove it is written down.

I've resubmitted, and this time I've said in the submission that the change has already gone
live. That sounds like an odd thing to volunteer, but it changes what the reviewers should care
about: objections to the *shape of the tables* are now the expensive ones, because a live table
is much harder to change than code that hasn't shipped.

Next is the part that makes the register do something: the path that writes an entry into it, the
path that binds an entry to a real site's pages, and then the first check that actually runs. That
last one is what turns Sunday's manual discovery — four carousel cards pointing at a page that
doesn't exist — into something the system notices by itself.
