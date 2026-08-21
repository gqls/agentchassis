# SUMMARY — 2026-08-21. Phase A complete: every small self-contained tool is rebuilt.

*Written to be read aloud. Current state only; chronology in `README_where_we_are.md`, evidence in
`NOTES_native_rebuild_of_ported_tools.md`.*

## What we are trying to do

Replace all sixty-three hand-carried tools on webdesign.co.uk with tools the framework itself
generates at the same addresses, so every future platform improvement reaches them automatically.
The owner has ruled all sixty-three go through the framework; there is no leave-alone option.

## Where we have come from

A cautious pilot became a six-step routine. Along the way we hit three platform walls (all since
fixed, two by other lanes), changed the generator's own quality contract so six recurring fault
classes cannot be generated any more, and learned the project's defining fact: most of the old tools
were quietly broken, in ways a visitor cannot report. We also lost the old-for-new swap race twice
to a notification channel that can deliver hours late, and twice a page briefly showed both tools;
both were repaired within minutes of being seen, and the working method now holds the session's
attention through the whole swap so the race cannot recur.

## What we have done

Twenty-eight tools are rebuilt — every self-contained tool under eight kilobytes, which is all of
Phase A. Twenty-four are confirmed live on the public site by fetching the real pages with caches
defeated; the last four await only a routine re-publish that is already queued. The count of old
tools that turned out to be genuinely defective keeps rising — the register of "a tool that asserts
something untrue about itself" alone stands at more than ten — and each rebuild removes the defect
class, not just the instance. Our own contribution back to the platform this week: tool build briefs
can never again be published as a page's search-engine description (council-approved, live), and we
found, reported and saw fixed within the hour a config defect that had silently broken every new
tool build fleet-wide.

## Where we are now

Thirty-five tools remain. Nothing is blocked by the platform. The in-place re-fix path another lane
built at our request is now live and proven — it corrected a small CSS-ordering defect in one of our
own rebuilds without any manual steps, and its safety gate scored its first real catch doing so.

## Where we are going

Phase C next: thirteen tools whose logic lives in external files, where writing the specification
from the tool's real behaviour in a browser is the work and the external file must be retired with
the old page. Then Phase B, the large tools, ending with the five rich applications the owner
reviews personally, one at a time. Done still means sixty-three of sixty-three, each proven at the
served bytes.
