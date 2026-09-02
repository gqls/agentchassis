# Where we are — the empty listing pages bug (444)

Plain-prose log for the owner. Append-only, newest at the bottom.

## 2026-09-02 (late evening) — lane opened, cause fully understood, fix designed

Your critique of designblog ("explaining the brief and not answering it") turned out to be
a class: seven pages across the remakes — now eight, we found seotools' directory is empty
too — serve nicely-written pages that contain zero of the thing they exist to list. Every
automatic check passes because the pages are full-size, well-formed and return 200; the
only test that fails is counting the items, which nothing was doing.

Tonight we got to the bottom of WHY, and it's more interesting than a missing feature.
The platform already has everything needed to fill or refuse these pages — a news pipeline,
a directory pipeline, and even a safety rule on the directory component that says "don't
render me without at least one entry". The pages shipped empty because of three separate
disconnects:

1. The page planner plans a news page whether or not the news machinery has been switched
   on for that site — its instructions literally tell it to plan the ideal site and let
   the build system sort it out. The build system doesn't.
2. The directory component's safety rule never fired because of an unlucky combination of
   two GOOD safety decisions: the data lookup treats a missing configuration as an error
   (so it can't be mistaken for "no businesses"), and the section builder deliberately
   doesn't hide errors (so they can't be mistaken for "no data"). Result: the error was
   logged once and the section was built empty anyway. Each rule is right; together they
   produced exactly what each was preventing.
3. Glossary pages have no machinery behind them at all — they were planned as ordinary
   text pages, so the writer wrote text ABOUT a glossary.

The fix we've designed closes the door where it should have been closed: when a plan is
validated, any listing page whose data source doesn't exist for that site is held back and
recorded as a named capability gap ("this site needs a news source", "this directory needs
a kind"), instead of being built as prose. That list of gaps becomes the enablement
checklist the remake process needs before firing the next 18 briefs. A second small repair
makes an errored data lookup put the section in the human review queue instead of shipping
it hollow.

Next: the plan goes to the council for review, then we build it. The other threads
(portfolio positioning, designblog, the feed lane, the copy lane) are all corresponding —
each holds a piece of the per-site clean-up, and the WebProNews feed you flagged is the
proving case for switching a site's news on properly.

## 2026-09-02, late — built, reviewed three times, approved, half live

The fix is built and went through the council three times the same evening. The first two
rounds sent it back — and both were worth it: reviewers caught that a redesign of an
existing site could have deleted a page that was already built and live (fixed — built
pages are never removed, they just get a note filed against them), that a rare error case
would have quietly demoted a working fallback (fixed), and that I'd copied two lookup
tables by hand that should be read from their one true source (fixed — they're now derived,
so future additions can't drift). Third round: approved.

Where that leaves things tonight: the planner's INSTRUCTIONS are already updated live — it
is now told a listing page may only be planned when the thing that feeds it exists, and to
record what's missing instead of building around it. The enforcement code — the part that
actually holds an unfillable page back and writes the "this site still needs X" ticket —
is committed and rides the next platform release. Until that release, the next 18 site
briefs are protected by the updated instructions and by the release checklist the
portfolio thread adopted; after it, the protection is mechanical.

The empty pages already out there (the two sites you critiqued, plus one more we found on
seotools) still need their per-site fixes — those are with the threads that own each site,
and the "what to enable" recipe is written down in their release runbook.
