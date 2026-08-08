# Where we are — the stale site chrome case (bug 117)

Plain-prose log, append-only, newest at the bottom.

---

**2026-08-07, first entry.**

I went looking for the next open bug nobody else was working on. There are 49
files in `bugs_open/` and about twenty other sessions running against this tree,
so most of the work was proving a bug was *free* rather than picking one. Eight
had no mention in any live session. Six of those turned out to be spoken for or
already finished — two of them (126 and 181) are fixed and live and only sitting
in `bugs_open/` because you asked for finished bugs to stay there. Worth knowing:
a file being in `bugs_open/` no longer means the bug is open.

That left 117, filed on 27 July and untouched since.

**What 117 is about.** Every page on every site is built as head + header +
sections + footer. The three chrome parts are not generated when the page is
built — they are looked up as frozen strings in a table and pasted in. So if you
change the shared footer, nothing on any existing page changes. You can
re-render all eighteen pages of a site, successfully, and see no difference. The
person who filed it lost two correct changes that way before working out why.

**What I found that the bug file does not say.** The bug asks for someone to
measure how widespread this is before designing a fix, and marks that as not
done. I did it, and it turned up something more awkward than a missing check.

There is already a check for exactly this. It is live, it runs, and it files real
work. It is called "stale site components". And it compares the wrong two things:
it asks whether the stored chrome is older than the site's most recently
regenerated *page content*. Chrome is not built from page content. The two have
nothing to do with each other.

So I cross-tabbed what the check says against what is actually true, across all
57 chrome rows on the fleet. Of the 39 times it currently fires, 36 are for a
reason unrelated to chrome being stale. And there is one site — oufe.com — whose
footer genuinely is out of date, where the check stays silent. It is wrong in
both directions at once.

That matters more than a check that does nothing, because this one is believed.
It files rebuild jobs, they get picked up, and they complete. Real capacity is
being spent on a signal that is mostly noise, while the case it exists to catch
slips past.

**A wrong turn worth recording.** I wanted to prove those out-of-date footers
actually serve different HTML, not just carry an older timestamp. I found one
tag that appeared in the new renders and not the old, and nearly used it as
proof. It was a dead end: that block only appears on sites that have compliance
text, so it was telling me about the site, not about the template version. When
I then tested every candidate tag properly, none of them separated old renders
from new. So I have to be straight about this — I have shown those footers were
rendered from an older version of the template, and I have *not* shown that
re-rendering them would change anything a visitor sees.

That sounds like a weakness in the case. I think it is the opposite, and it is
the thing that should decide the fix. A timestamp cannot tell "out of date and it
matters" from "out of date and it makes no difference". That is precisely why 36
of 39 firings are wasted. The question worth asking is not "is this older than
that" but "would re-rendering actually change anything?" — and you answer that by
recording a fingerprint of what the chrome was built from, and checking whether
it still matches.

I also nearly shipped the same mistake myself. My first idea for the corrected
comparison was to widen it — check the template *and* the nav *and* the site
record. I ran it before writing it up, and it marks essentially everything stale,
because the site record gets touched constantly for unrelated reasons. A wider
timestamp is not a better one.

**One more thing that needs fixing regardless.** There is a repair job that
rewrites component templates to strip hardcoded colours, and it updates the
template without updating the "last changed" timestamp. It specifically targets
chrome components. So a change made by that job is invisible to any
timestamp-based check, including a corrected one. Every other writer of that
table sets the timestamp; this is the only one that does not, which makes it a
small fix and an easy thing to break again.

**Where it stopped.** I filed the formal diagnosis run, which completed. I could
not retrieve its verdict — the run stores the evidence it gathered but the
conclusion appears not to be saved, which is a known problem noted elsewhere in
the repo. So I am not claiming it confirmed me; the case rests on my own
measurements, and I have said so in the notes.

I then handed the design work to a planning agent and hit the session limit
partway through. **No code has been changed and no plan exists yet.** What exists
is the evidence, the measurements, the queries that produced them, and a clear
statement of which fix shape the evidence points at. The next session picks up at
designing the fingerprint. I have written a handoff with the specifics.
