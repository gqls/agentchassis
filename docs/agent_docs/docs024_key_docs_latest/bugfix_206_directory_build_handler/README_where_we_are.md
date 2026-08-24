# Where we are — directory build handler (`bugs_open/206`)

Plain-prose running log, append-only, newest at the bottom. Started late (2026-08-24) — the
lane ran from 08-06 to 08-08 without one, which was a gap in the standing five, so the first
entry back-fills what happened in plain terms before carrying on.

---

## 2026-08-24 — back-fill: what this lane was about, and what it did in August

Two pages on vetcomparison.uk had never worked. One was meant to list the site's vet practices,
the other its advice guides, and both returned "not found" — while the site's own homepage
linked to the first of them. That was the original complaint.

The cause turned out to be two gaps stacked on each other. The platform keeps a list of "which
kind of page gets built by which builder", and for these kinds of page that list said, in a
comment, that no builder existed yet. Underneath that, the component meant to render a real
business listing pointed at a data query that had never been registered anywhere. So even a page
that got as far as being planned had nothing that could fill it in.

The lane built the missing piece: a query resolver that reads a site's own directory-export
configuration and pulls from the same business data the site already exports (so the page and
the export can never disagree), plus an action that fills in a page's layout **only** when it
has none at all — written that way deliberately so it can never overwrite a layout a person or
another process had already chosen. A new handler chains those together and then hands off to
the platform's ordinary page builder. No new content-writing logic was needed.

That went through three rounds of the council review gate before approval, and two further
defects showed up only when real builds ran — the kind you cannot find by reading code, only by
watching it fail. Both were fixed by follow-up migrations the same day. Both pages then built
and deployed through the ordinary queue, with no manual dispatch, which was the whole point.
The directory page listed sixty-one real practices with real postcodes; the guides page listed
exactly the three guides that actually exist and invented nothing.

Then the lane went quiet for sixteen days.

---

## 2026-08-24 — resumed: the fix was fine, but it was never fleet-wide

Picked the lane back up today. First job was to check the August work is still true, because a
fix that was live three weeks ago is not evidence about today. It is: both pages still serve
their real content, and — the part that actually matters — a fleet-wide re-render swept them
yesterday and the real listings survived it. So the pipeline genuinely reproduces the result
rather than a good page having been frozen in place. A second team working on vetcomparison
proved a third kind of page through the same handler this morning, first attempt.

Then the surprise. I went looking for the leftovers we knew about and found instead that the
August fix, while correct, was only ever installed on **one of the doors**.

The platform decides "which builder builds this page" in more than one place. In August we
taught one of them about the new handler. A second one — the routine that reconciles a site's
plan against what has actually been built — never asks that question at all: it had the answer
"use the generic builder" typed directly into it. So pages of exactly the kind this bug is about
keep arriving at a builder that cannot build them, failing in the same way, and parking for a
human who never comes.

Five pages are sitting in that state right now across three other sites. The one that makes the
point is a directory page on garden-tools.uk: it has been waiting fifteen days for a builder that
has existed and worked the entire time. It was never told the builder's name.

The fix is to stop having two copies of that decision. There is now one function that answers
"which builder builds this kind of page", both routines ask it, and it is covered by tests — the
decision had **no** test coverage at all before today, in either copy. I also made it so that a
page whose kind genuinely has no builder yet gets filed as a visible, deferred "we need this
capability" note, rather than being sent to a builder that will fail and then parked under an
error message blaming missing data when the real cause is a missing builder.

**One thing needs a decision that is not mine.** Another team looked at this same set of stuck
pages three weeks ago and deliberately decided to leave that routine alone, on the reasoning
that these stuck items are real findings and should stay visible. I think my version keeps them
just as visible and describes them more honestly — nothing is hidden, the same record is still
written at the same moment — but they made that call on purpose and with the council's backing,
so I have put it to the council rather than quietly overruling it, told that team in their own
file, and offered to take that half straight back out if they disagree. The other half of the
change — actually routing pages to the builder that exists — is untouched by their decision and
is where the real win is.

**Three things I got wrong today**, all caught by other people's checks rather than my own, and
all written up:

- I corrected another team about something in **my own lane's history**, from memory, and was
  wrong — my own notes already held the correction. They had already written my wrong version
  into their file before I caught it.
- My first survey of how widespread this is returned **zero**, and the zero was an artefact of
  how I asked. A neighbouring session mentioned, in passing, a lesson about testing your own
  filter against a case you know is real. I tried it; the filter was blind. The corrected survey
  found the five stuck pages and **changed what the fix should be** — the problem was the
  producer, not the list I had been staring at.
- I cited a database column that does not exist in a council submission, from misreading two
  command outputs as one. The council's reviewers caught it independently and sent the
  submission back. The code was right; the description of it was not.

And the one worth remembering: I re-discovered a problem another team had already measured,
ruled on, and closed **three weeks ago**. Our standard "check nobody else is on this" tools all
search by the bug's number or name — but this problem's identity is an error message, and no
number-based search will ever find it. I only got there because the council's own check returned
a number eight times larger than mine and I went to find out why.

**Where it stands**: the fix is committed and reviewed-pending; it is Go code, so it does
nothing until the next time the fleet's images are rebuilt. The bug file stays open — not out of
caution, but because the fault is genuinely still happening on the fleet today, and it stops
happening when the fix ships, not when it is written.
