# SUMMARY 2026-08-04 — the element-level check exists

Written to be read aloud. Previous in the series: `SUMMARY_2026-08-03_contact_info_fabrication_complete.md`
(which closed the bug), and `SUMMARY_2026-08-02_contact_info_fabrication_closed.md` before it.
This one marks a different kind of milestone: the bug was already closed, and what has
changed is that the **class** behind it is now measurable.

## What we're trying to do

Stop the site-building platform putting things on customers' pages that shouldn't be
there. Two flavours: a component that **invents a fact** when the real one is missing (a
phone number, a price, opening hours), and a component that **renders a hole** — an empty
heading, a link with no words in it, an image tag with no image. The first is a lie; the
second is a visibly broken page. Both come from the same root: a component template asks
for a piece of content, the content isn't there, and nothing decides what should happen
instead.

## Where we've come from

The invention half was fixed and is guarded from two directions: the platform now refuses
at write time to accept a component whose fallback text asserts a fact, and a daily check
reads the whole live component library and reports any that do. Then the schema's own
"skip this field if it's missing" instruction — which sixty-eight fields declared and no
template obeyed — was made real, with each of the twenty affected components proven to
render correctly both with and without its data.

Then production contradicted the shape of that win, within hours and by accident. Another
team created two new pages using a component we had just gated. The gates worked
perfectly: no broken image, no empty meta row. And the pages still shipped an empty
heading — through a *different field of the same component*, one whose instruction said
"skip the whole section" rather than "skip this field". Our check only understood the
second phrasing. So the sixty-eight fields we fixed were never "the broken set"; they were
"the set our check could see". Across the library there are roughly two thousand fields
whose behaviour on missing content nobody has ever measured.

## What we've done

Worked the plan that finding produced — eight items, six now done and the seventh built.
The important one is last.

The small ones first, because they remove recurring friction. The commit hook now refuses
a review label that isn't a real reference — three sessions had accidentally committed the
word "pending" where an identifier belongs, and because this project never rewrites
history, each was permanent. Committing without any label still works: we haven't made
review compulsory, we've made a false claim impossible. The daily check now retries when
its data fetch is cut off mid-stream, and while adding that I found the failure had a
second form nobody had noticed, in which a truncated download was reported with the same
code the check uses for "I found a real problem". There's a new small tool that picks
markers for proving a deployment actually shipped — it caught the trap that caused
yesterday's false alarm on another thread, and then caught a subtler one of its own, which
I'll come back to. The two pages storing a raw model reply where their content should be
are now a filed bug, with the repairability of each measured rather than guessed. And the
one failed page rebuild I'd been asked to look at turned out not to be a failure at all:
the platform refused to save because the rebuild came back with a third of the content it
was replacing. It protected the page. That, and the fact that two other stalled items are
waiting on a component that can't be generated, went to the team that owns that site.

**The seventh item is the one that matters.** There is now a check that renders every
component in the live library, twice for every piece of content it references — once with
everything present, once with that one piece removed — and reports when removing it opens
a hole in the page. It never reads the instructions at all. That's the whole design: our
previous check was blind because it filtered on how a field was *described*, and this one
can only see what the page *does*. It uses the platform's own drawing code rather than a
copy, so it cannot be right about a replica and wrong about production.

Calibrated against the live library, it finds all three of the day's real incidents,
including the one that embarrassed us — and, in the same component, correctly stays silent
about the field we had already fixed. Finding the problem and not finding the non-problem,
inside the same file, is the result worth having.

## Where we are now

The check exists, is proven to discriminate, and is **not yet running on a schedule** —
being honest about that is the point. Our existing daily check is a script the cluster can
be handed directly; this one is a compiled program, which needs a little packaging first.
And before it's ever allowed to turn anything red, it needs a baseline: it currently
reports about a thousand places where removing content would leave a hole, and roughly
forty components are *legitimately* blank when they're built because a browser fills them
in afterwards. A thousand-item red light is a red light everybody learns to ignore —
which is the exact mistake we were avoiding when we chose not to make the old check fail,
arriving now from the opposite direction. What we want from it is not the total; it's
"something new appeared today".

One more thing worth saying, because it's the kind of mistake I'd rather record than bury.
The new deployment-proving tool's very first suggested command came back empty against
every running copy of the software — for a marker that was definitely inside it. The
marker contained a typographic ellipsis, and the standard tool for reading text out of a
compiled program breaks its output at any such character, so the phrase was never on one
line to be found. I had checked that the text was in the program; the command I shipped
asked a subtly different question. The safeguard we normally rely on couldn't catch it:
it reads zero when things are working, and here everything read zero. The lesson is now
written down twice, once as a trap to read before touching this, and once as a wrong call:
**run the command you're recommending, exactly as you're recommending it.**

## Where we're going

Two things, in order. Package the new check so it runs unattended, and give it a baseline
so it reports growth rather than a census. Once that's in place, the older check can
finally be allowed to fail builds — the objection to doing that until now was that it
could be satisfied by a change that clears the warning without closing the hole, and the
new check is precisely what notices that.

Beyond those, the honest position on the wider class: we have measured that the exposure
is large, and we have a way to see it. We haven't decided how much of it is worth
fixing, and that will need judgement about which holes a visitor would actually notice —
not another sweep.
