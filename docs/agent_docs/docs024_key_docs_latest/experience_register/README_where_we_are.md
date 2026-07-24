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
