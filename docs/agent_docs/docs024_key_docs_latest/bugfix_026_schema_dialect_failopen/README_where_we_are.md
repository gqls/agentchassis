# Where we are — bug 026, the news headline that came out blank

## 2026-07-21 — picked up "bugfix 026", found most of it already fixed, built the structural half

Bug 026 was two problems in one shared component (the "news listing" block every news page
uses): on a Spanish site it showed the English words "Loading latest news..." before the page
finished loading, and its big heading came out completely blank.

When I actually looked at the live sites, most of it was already fixed — by another thread
that had been rebuilding how news pages work. On the three real news sites (relojistas,
gaswholesalers, robot-hands) the news now renders properly on the server, the loading text is
gone, and the heading is filled in (in Spanish, on the Spanish site). So the *visible*
symptoms are already sorted.

That left the one part the other thread had **deliberately left for bug 026** — and they even
wrote a note in their own change saying so. It's the interesting one, because it's the
"why did nobody notice the heading was blank?" question.

Here's the plain version. Every component has a little spec that says what content it needs
and which bits are required. There are **two** formats for that spec — an old one and a new
one — and the platform only ever learned to read the new one. The news-listing component was
still written in the old format. So two different parts of the system that were *supposed* to
catch a missing required heading both just shrugged and said "this component needs nothing"
— because they couldn't read its spec at all. One part therefore never asked the writer to
produce a heading, and the other never complained that it was empty. The heading shipped
blank and nothing flagged it.

The trap here is a general one worth remembering: a checker that only understands one format
doesn't fail *safely* on a different format — it fails *silently*. "I can't read this" comes
out looking exactly like "there's nothing to check."

**What I did.** I taught the platform to read the old format too — projecting it onto the new
one — so both the "ask the writer for it" step and the "refuse to ship it empty" step now
understand an old-format component instead of ignoring it. I added a test built around the
exact spec the news component had when the bug was filed, so this can't quietly come back.

**An honest caveat I want you to have.** Right now, **no** component is still in the old
format — the other thread's rebuild migrated the last one. So this fix doesn't repair anything
that's broken today; it closes a trap door that would re-open if an old-format spec ever comes
back (a config reload, restoring an old site snapshot, or the component-builder emitting the
old format). You chose to build it rather than just close the bug on the diagnosis, which I
think is the right call for something this cheap that sits under the whole content pipeline —
but I didn't want the "it's already extinct" part buried.

**Two loose ends that are NOT this bug.** Two news pages still have a blank/stale heading in
the database — idea.uk and ai-agent-orchestration. Both turned out to have the *wrong page
type*, so the news machinery skips them entirely. idea.uk isn't even live yet (it 404s). That's
a different known bug (015, the "mistyped page type orphans a page" one), and I've pointed it
there rather than fixing it here.

**Where it stands.** The code is written, tested, and committed. It's a change to core
plumbing, so I've put it through the reviewer council for a second opinion before it ships
(that takes about half an hour). It won't actually take effect until the next image is built
and rolled out. Once the council's happy and it's live, I'll verify it by hand — deliberately
feeding it an old-format component with an empty required field and checking it now gets
refused — and then close the bug.
