# SUMMARY 2026-07-28 — the config that lied, and the records that could not say where they came from

## What we're trying to do

Fix two open bugs that nobody owned, `100` and `101`, and fix them at the level of the
framework rather than the individual case. They sit on the same step of the same agent,
and each bug file independently asks to be fixed alongside the other.

Underneath both is one shape: **something declares a behaviour, nothing performs it, and
nothing says so.** In `101` it is scrape settings that no code reads. In `100` it is a
provenance column populated from a field the AI model was never asked to produce. In
both cases the system reports success, and the artefact — the config, the row — reads as
evidence of something that never happened.

## Where we've come from

The estate has met this class repeatedly. A numeric config value that never reached its
action, and went unnoticed for months because the configured number happened to equal
the code's default (`042`). A news search that was a plain web search, because the
search type was dropped at a provider boundary (`127`, closed this morning). Each was
found by someone tripping over it, and each was fixed in place.

`101` was filed with unusual care by the vetcomparison thread: it named the four dead
keys, measured the cost, recorded that reading the config had already misled a session
into a false claim within an hour — and, importantly, marked one thing it had **not**
settled, with an instruction not to implement until someone did. `100` was filed the
same day, and between them they blocked that workstream's restart of vet data
collection: every row a restarted crawl produced would be born unpublishable under our
own sourcing rule.

Neither bug had an owner. The thread named against `101` was the blocked party, not a
fixer, and had written no code in two days.

## What we've done

Three code fixes and one new mechanism, all committed, with the review council's
verdict pending.

**We settled the unsettled question, and it was a third bug.** `101` asked: does
production's scraper discard page footers? If so, fetching more pages achieves nothing,
because company registration numbers live in footers. The answer is yes, and the cause
is four lines in a shared provider. There is a setting for "main content only". Our code
could send it when true. When false, it sent *nothing at all* — and Firecrawl's
documented default for "nothing at all" is **true**. So every caller explicitly asking
for the whole page has received the opposite, silently, for as long as the code has
existed. Three live steps ask for it, including the one that fetches a site's
stylesheet. The same file gets it right on its other code path, twenty lines away.

**We made provenance come from the fetcher.** The writer had been asking the LLM where
its data came from. The tempting repair — ask the model harder — is the one thing the
bug file forbids, because provenance asserted by the call that produced the facts is not
provenance. It turned out no new mechanism was needed: the component doing the fetching
already records the URL and the time, next to the HTTP call. Nobody read it. So the
writer now reads the fetch record, and the three lines that read the model are deleted
rather than kept as a fallback — a fallback would have come back to life the first time
a model volunteered a plausible URL.

**We made the four dead settings honest** — two implemented outright, two mapped onto
the multi-page dialect the adapter already supports, and a loud warning when a
single-page fetch carries settings it cannot honour.

**And we made the whole class detectable.** An action can now declare which config keys
it actually reads; anything else gets reported instead of silently ignored, with a
strict mode that refuses it outright once a contract is known complete. This deliberately
reuses a registry that already existed, that 134 files write to, and that — we
discovered — nothing had ever read.

We kept it opt-in, against the instinct to enforce everywhere. The survey is why: 228
distinct actions, 811 distinct settings between them. A strict rule applied on a guess
at that scale would start rejecting configurations that work, which is a worse bug than
the one being fixed. So there is also a report that shows how many actions haven't opted
in, to keep the gap a visible number rather than an invisible absence.

**It paid for itself immediately.** Its first run against the live system found a
*fifth* dead setting the bug file never knew about — a typo, `add_protocol` where the
code reads `add_protocol_if_missing`, in a different action entirely. That one is not
cosmetic: without it a bare domain goes to the scraper and the fetch fails. No amount of
re-reading the bug file would have surfaced it.

## Where we are now

All four pieces are committed with tests, and the tests are the kind that can fail: the
central one was run against the original code first and watched to fail before it was
trusted, and the provenance suite contains the discriminating case — a model-supplied
source URL must **not** be accepted as a fetch record, which is the only test that
separates the right fix from the rejected one.

**Nothing is live.** Two images have to be rebuilt and rolled: the web-scrape adapter
for the footer fix, the chassis for everything else. One database constraint is written
and deliberately *not* applied, because it must follow the code — applied first, it
would refuse writes the running binary cannot yet satisfy.

One thing is honestly unverified and is marked as such in the code, the bug file and the
notes: the exact shape of the data the provenance reader consumes was traced through the
source rather than observed, because no run carrying it survives — collection has been
off since March and the history table is on a retention clock. The reader accepts six
shapes and complains loudly if none matches, and the first real verification run settles
it.

## Where we're going

The next steps are mechanical and in order: council verdict, two image builds, then the
constraint, then a single vet verification run checked against two columns — the source
URL must be populated **and** the model must still not be claiming it. Both bugs stay
open until that run is green, because the bar here is fixed *and* live, not fixed.

Beyond that there is one judgement call left deliberately undone, and it belongs to
someone else. Making the two affected agents actually crawl multiple pages is a real
behaviour change to agents this thread does not own — one of which has no owner at all —
under a data lane that has been switched off since March. They now say clearly that the
setting cannot take effect and why. Turning it on is a decision, and it should be made
as one rather than arriving as a side effect of a bug fix.

The longer arc is the coverage number: 208 actions have not declared what their config
means. That figure exists now, which is the point of it.
