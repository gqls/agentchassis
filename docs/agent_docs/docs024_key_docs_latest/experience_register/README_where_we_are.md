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
