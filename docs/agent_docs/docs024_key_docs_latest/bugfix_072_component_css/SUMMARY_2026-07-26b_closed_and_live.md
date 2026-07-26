# SUMMARY 2026-07-26b — closed and live

Supersedes nothing: `SUMMARY_2026-07-26_component_css_diagnosed_and_fixed.md` stands as
what we believed before the roll, and it said plainly that the fix was not live. It now is.
This is the inflection.

## What we're trying to do

Make sure that when a page ships markup, the styling for that markup ships with it.

## Where we've come from

A bug filed on 25 July: news cards rendering bare on two sites out of five, cause unknown
and honestly declared unknown. Diagnosed on 26 July — the site stylesheet is written once,
by the design agent, and never regenerated, so it freezes while the site keeps changing.
One of the two sites had been asking for styling that was never written for eighty days.

## What we've done

Fixed it in two places that cover different populations, and made them aware of each other
so they never both apply the same rules. The page assembler now collects its components'
styles as it builds; and the two news components carry their own styling, which is what 86
of our 94 components already did. The CSS was copied from the existing source rather than
retyped, which is what let us change two sites without touching three others.

Then the chassis was rebuilt and rolled (v1.0.1171), and we verified it properly rather
than declaring victory:

- the code is in the **running binary**, checked with a phrase that only exists because of
  this change, alongside a pre-existing phrase and a nonexistent one so a false pass would
  have shown;
- both broken homepages were re-rendered and **now render styled cards**;
- the two working sites came back **byte-identical** — the rules now inside their pages
  match what their stylesheet already served, so we added the missing styling rather than
  repainting anyone.

That control was the whole point of the test, and it held.

## Where we are now

**The bug is closed and the file has moved to `/bugs_closed/`.** The defect is no longer
reproducible on any of the five sites.

Two honest qualifications. First, the half of the fix that did the visible work was the
simpler half — the components carrying their own styles. The page-assembly collector
deliberately stood aside, because it detects that a component already carries its own CSS.
That is the designed behaviour, and it means the collector is installed but not yet
exercised in the wild, so we verified it by running its query directly, both ways: nothing
to add against today's world, and exactly the 3,355 missing characters against a simulated
pre-fix page. The safety net is real; nothing has fallen into it yet.

Second, a mistake worth recording. Meaning to fire the light page refresh, we passed an
empty argument where the script's documentation says to pass none — and in shell those are
not the same thing, so the *heavy* refresh ran instead, on two live sites. Nothing was
harmed, because that script guards against exactly the case that would have regenerated
copy and both pages were clean. But it was luck rather than care, and the heavier path is
what carried the fix onto the page, so the mistake was rewarded. It is in the standing
wrong-calls log and beside the command in the runbook.

One loose end: the change went to the reviewer council and no verdict ever came back — no
trace of the submission exists, which looks like a lost dispatch rather than a slow queue.
We have not resubmitted, because an absent row is also what a queued job looks like. The
practical consequence is that both commits are honestly marked un-reviewed rather than
carrying a stamp they did not earn.

## Where we're going

Nothing further is owed on this bug. Two things were surfaced and left deliberately
unbuilt, both named in the plan:

- **No automatic detection of this class.** A check would have to read what is actually
  served to a browser, and our detection machinery only reads the database — where the
  served stylesheet does not exist. A database-only check would have reported these two
  sites healthy throughout. Better a named gap than a check that cannot see the failure.
- **The snippet-matching granularity problem.** Roughly 17 of 21 style snippets are keyed
  on generic words (`card`, `cta`) that match no real component name, so they have never
  reached any site. That makes sites plainer than intended; it does not make markup ship
  unstyled, so it was correctly out of scope here — but it is a real, documented, untouched
  defect and someone should take it.
