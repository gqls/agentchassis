# Where we are — webdesign tool rebuilds

2026-08-15. The owner asked us to speed up replacing the 63 imported ("ported") tools on
webdesign.co.uk with tools the framework builds and manages itself. Today we found out how
native tools actually get built (a suggester files a work item, a generator builds it), and
discovered nothing was replacing the imported ones — the generator was only adding new tools.
We also cleaned up two hazards first: a shared template that a fixer bot had overwritten with
the wrong tool's markup (restored; nothing had broken yet), and the one page where a native
tool had been dropped next to its imported twin, showing both a leftover and a raw template
tag to visitors (repair queued, leftover retired).

The plan: prove the replacement recipe on one simple tool (the aspect-ratio calculator — its
rebuild is queued), then work through the simple tools one at a time. The rich, hand-built
apps (mind-map studio, meme studio, mini-CMS and friends) are deliberately excluded — an
AI rebuild from a one-line description would quietly lose what makes them good; they wait for
the faithful-conversion route or a per-tool decision.

## 2026-08-16 morning — what happened overnight, in plain words

The pilot did not fail because of the flaky agent handshake I guessed at in the handoff. It failed
because the tool-generator's "save the tool" step can only ever CREATE a page — it has no way to
say "attach this to the page that already exists". Since our whole plan is "rebuild the tool at the
same address", every rebuild would have hit this wall. I had the diagnosis loop confirm it (it did,
first go), built the fix as an opt-in switch (off by default, so nothing else changes), put it
through the council (approved), and committed it. It is NOT live yet: it needs the next chassis roll,
and then a small config seed that turns the switch on for the generator only. Both steps are
written down. Note: at 09:52 this morning the cluster was still running last night's build, so if
a newer one was pushed it had not landed anywhere I could see.

Two other things fell out. First, the "two odd audit items" were not two — a whole class of
audit-fix items had been filed with an EMPTY spec for weeks (233 of 233), because four agent
configs wrote the spec in a shape the platform silently ignores. Fixed at the source last night;
the three items I re-armed all ran and applied within minutes. Second, the ab-test page: the native
copy of that tool turned out to be a hollow shell (its labels and headings were never filled in),
and the page was serving 47 raw template tags. I put the old working tool back on that page and
queued a re-publish. ab-test will be rebuilt properly through the new route, second in line after
aspect-ratio.

The 285 close-out (proving the shared-template fence refuses) is being done by its own lane right
now — I saw the refusal land at 09:59 and left it to them.

## 2026-08-16 afternoon — the owner has settled the rich apps: rebuild them too

The one open question on this lane was what to do with the handful of hand-built applications —
the mind-map studio, meme studio, logic architect, the mini-CMS, the pasteboard — and the tools
whose code sits in separate script files. The plan had them set aside, because a rebuild from a
written description is a fresh implementation rather than a copy, and something is usually lost.
That trade was put to the owner and he has accepted it: **rebuild them as well.** So every one of
the 63 imported tools is now in scope for this route, and nothing here waits on the faithful-copy
conversion any more.

Three things stay exactly as they were, and matter more for these apps than for the calculators:
the description we hand the generator has to be written from watching the live tool work, not from
reading its page; each rebuild is judged before the old one is switched off, because we have
already had one come back as an empty shell; and the old version is retired rather than deleted, so
a disappointing rebuild can be put back in one step. My suggestion, which is the lane's to take or
leave: do these last and one at a time, once the simple tools have proved the process, so the owner
is only asked to look at the difficult ones after the easy ones are known to work.

## 2026-08-16 late afternoon — the first rebuilt tool is built and graded; the old one is switched off

The aspect-ratio calculator is the first tool the framework has rebuilt in place, and the part that
had never worked before now works: the generator attached its new tool to the page that already
existed, instead of trying to create a second page at the same address and dying. That was the whole
point of yesterday's fix, and it did it in 54 seconds.

Before switching anything off I graded what it built, because last time we assumed and got an empty
shell. This one is real: it reduces a width and height to a simplified ratio the proper way, it works
the other direction too (give it a target ratio and one dimension and it gives you the other), and
the 16:9 / 4:3 / 1:1 / 21:9 shortcut buttons are there. All the wording is written into the tool
itself rather than left as blanks for something else to fill in — which is exactly what went wrong
with the A/B test tool.

So the old imported version of that tool is now switched off. Switched off, not deleted: its content
is untouched, and I checked its fingerprint before and after to prove the bytes did not change, so
putting it back is a single flip if anyone dislikes the new one. The page will rebuild itself with
just the new tool on it shortly — it is sitting in a queue about twenty jobs deep. I will check the
live page once it has run; until then the rebuild is done but unproven, and I am not treating it as
finished.

Two things I got told to do that turned out to be impossible as written, both now corrected in the
plan. First, we were meant to record an "archive row" for each tool as the undo handle — there is no
such row, because the archive only records changes to a page's content, and switching a tool off
changes only its status. The undo handle is the old tool's own row, which is better anyway. Second,
the address we were checking the live page at, `/tools/aspect-ratio/`, is a 404 — the real page is
at `/tools/aspect-ratio/index.html`. That matters more than it sounds: every one of our "is it clean
now?" checks passes perfectly against a 404 page, because a page that does not exist contains none of
the things we are looking for.

While waiting I counted up what is left, and found a couple of traps worth knowing about. There are
97 imported pages on the site but only 63 are tools — the rest are learning pages, and one of the
63-looking ones is just the tools index listing. And there are two separate groups of exactly 13
tools that are easy to mistake for each other: 13 whose code lives in separate files (the awkward
ones), and 13 carrying a marker from an earlier repair effort. Only 4 tools are in both groups. If
someone used one list as a shortcut for the other they would get nine tools wrong in each direction.

I also read the generator's code and found something that would have bitten us on the second rebuild:
when it checks "does this tool already exist?", it does not care whether the existing one is switched
off. So the A/B test tool, whose failed rebuild we withdrew this morning, would block its own
replacement — and the run would finish successfully having built nothing at all. That is now written
down with the one-line fix, before it cost us a cycle rather than after.

## 2026-08-19, late evening — from the session that closed bug 286 (plain account)

Bug 286 — "the tool generator couldn't rebuild a tool at an address that already had a page" — has
been fixed and running since the 16th. Nobody had moved the file, so I checked it properly (the
running binary really carries the fix; the switch is on; seven rebuilds have gone through it; the
original error has not happened since the day it was filed) and closed it.

While doing that I found the next wall, which this lane has been climbing by hand every time it
re-fixes a tool it already rebuilt once: the generator can CREATE a tool for a site, but it cannot
REPLACE one it made. Three separate safety catches each stopped it on three consecutive days in
August, and the recipe grew three manual database edits plus a race to retire the old copy before
the page re-rendered. I filed that as bug 331, had the diagnosis loop confirm it (it did, first go),
and built the fix: an item can now say "replace the existing one", and the generator rewrites the
tool in place — same identity, old version kept automatically for a revert, page never shows two
tools. It is written, tested, and submitted to the reviewers; it goes live on the next chassis roll
plus one config line (held until then). After that, a re-fix is one filing and nothing else.

Also: the library-collision fix the 311 session built today is already live on the 17:13 roll, so
the two "blocked" tools (ab-test, meme-generator) are not blocked any more.

## 2026-08-19 late evening into 2026-08-20 morning — the two stuck tools, and a checker we mostly did not need to build

**The two tools that could not be rebuilt are no longer stuck, and one of them is done.** The other
session's fix for the library collision went live on the 17:13 roll, so I picked up the A/B test
significance calculator — the one that had defeated three attempts across three days — and it rebuilt
first time. I checked their fix the way you would want it checked: not "the version number went up",
but the running program on both machines, and then the database row afterwards to prove the shared
library copy of that tool had not been touched. It had not. I have written the result back into their
notes, because their fix was half of a pair the owner asked to be proven together.

**We are at 14 of 63, and every one of the 14 has been checked on the live page, not just in the
database.** Two of them (the A/B tool and the blob maker) were finished late last night and their pages
re-published overnight; I confirmed both this morning by fetching the real pages with the caches
defeated, and by checking that identifiers which only the OLD versions contained are now absent. That
last part is the bit that actually rules out looking at a stale copy.

**What the tools were doing wrong, again.** The A/B test calculator would tell you "not significant,
keep running the test" when you had entered no data at all — clear the visitor count and it printed a
verdict with a Z-score of "NaN". It also announced a variant that had performed significantly WORSE
using the same wording as a win. The blob maker had three controls and no wiring: dragging a slider or
picking a colour did nothing, and the only button that worked also re-rolled the random shape, so you
could never keep a blob you liked and change its colour. The shadow stacker, which is building now,
turns the whole shadow off if you clear any one of six number fields, and still offers you the broken
text as CSS to paste. That is now ten of the fourteen tools we have opened that were measurably broken,
and four of those ten lie about themselves in their own output — a claim, a comment or a control that
says work happened when it did not. Those are the ones nobody would ever catch by looking at the page.

**The checker we said we owed: most of it exists, and one part of it would have been wrong.** The plan
was to write something that scans the tools we have not rebuilt yet and finds these faults
automatically. Before writing it I went and read what the platform already has. Two things came out.
First, we already own an automatic reviewer for tools, and it had ALREADY reported the A/B test
division-by-zero — in its own words, four days before I found it by hand. So the value of a new scanner
is much smaller than we thought. Second, and more useful: one of the six rules we wanted to check
cannot be checked this way at all. The obvious test for "does this tool validate its numbers" is to
look for the standard guard against a bad number — and the A/B tool we built last night does not
contain that guard, because it uses a stricter check earlier on instead. A scanner written the obvious
way would have raised a complaint against the best-validated tool on the site. So what is worth
building shrank to two small rules that cannot be wrong (a tool must not use pop-up alert boxes, and
must not wire its buttons the old inline way), and both go inside a check that already exists and
already runs.

**One scare that turned out to be a fixed bug.** While measuring whether it was even worth filing these
findings, I found 205 tool-improvement jobs sitting in a dead state — which reads as "the fixer never
works, do not bother". Looking at the actual rows: twenty of them had been filed under one shared name
instead of one name per tool, so the system's own "this has failed twice already" rule stamped them all
as failures before anything ran. Then the timestamps: a migration fixing exactly that was applied at
17:17 that afternoon, and from 17:24 the jobs carry proper individual names and do get done. So the
graveyard is a scar from before the fix, not the current state. I nearly wrote the opposite down.

**Where we go next.** Finish the shadow stacker, then keep going down the list smallest first — the
recipe is boring now, which is what we want. There is one more piece of good news I have written into
the handoff: another session has just built a "replace the tool in place" capability, which removes the
three manual database edits and the small race we have been running on every single rebuild. It needs
the next chassis roll and one configuration line, and they have explicitly asked this lane to be the
one to test it on our next re-fix. After that, each rebuild is one filing and nothing else.

## 2026-08-20, evening — same session, short

The reviewers sent yesterday's "replace a tool in place" change back with one good objection: nothing
stopped a replacement that was an empty shell from overwriting a working tool. That guard is now
built — a replacement must actually contain visible content, and can't shrink the tool's text past
the same limit every other rewriter honours — and the change is back with the reviewers. Today's
fresh fleet build carries yesterday's code switched off, so nothing is exposed while we wait. Also
wrote a new milestone summary (SUMMARY_2026-08-20): five tools rebuilt has become 22 of 63, and the
three platform walls are down or falling.

2026-08-21 — Where we are, in plain terms. Twenty-four of the sixty-three tools are now rebuilt,
live, and checked on the real website. Yesterday's fresh platform build carries both of the fixes
this lane cared about: our own fix that stops a tool's internal build instructions ever being
published as the page's search-engine description (approved by the review council on the first
round), and another lane's mechanism that will soon let us re-fix a tool in place without the
delicate manual swap we do today. Two embarrassments to own: twice this week a page briefly showed
BOTH the old and the new tool at once, because the automatic wake-up we relied on can arrive hours
late. Both were caught and repaired within minutes of being seen, and we have replaced the wake-up
with a method that cannot be late. What is left: four small tools, then thirteen whose logic lives
in separate files, then the big ones you asked to review personally, one at a time. Four decisions
are yours and none is urgent; they are listed at the top of the new handoff file.

## 2026-08-21, afternoon — bug 331 is closed; re-fixing a tool is now one filing

The "generator can never fix a tool it built" problem is done, end to end, and proven the hard way.
The new build went live last night; the switch went on this morning; and the first real use caught
three things in one afternoon. First, the safety gate we were made to add refused my own first
attempt — rightly, because my instructions forgot to say "keep the explanatory text", and the rebuild
would have dropped it. Second, the switch itself had a wiring fault that quietly broke every ordinary
tool request across the fleet for an hour and a half — another session spotted it within the hour and
I corrected it on the spot (one request affected, already redone; both fixes went back through the
reviewers and passed). Third, with the wiring corrected and the instructions written properly, the
oklch colour picker was rebuilt in place — same identity, old version kept automatically for undo,
the CSS ordering defect fixed — and the live page now serves the corrected tool. The runbook's
three-step manual workaround and its race are retired: fixing a built tool is now a single request
with one extra line in it.

## 2026-08-22 — four finished pages briefly went backwards, and it was not our doing

Overnight discovery: four of the tools we had already rebuilt and proven (the grid generator, the
JSON cleaner, the noise generator and the text extractor) were publicly showing BOTH the old and the
new tool again — stacked on one page — for about nineteen hours. Nobody broke them on purpose. A
housekeeping robot that fixes stray markdown punctuation went over the whole site yesterday
lunchtime, found some in the OLD, retired copies of those four tools, "fixed" it there, and in doing
so marked those retired copies as live again. The next routine re-publish then put them back on the
pages. The housekeeping robot's own success check looked only at the punctuation, so it reported a
perfect run.

All four pages were repaired this morning and re-checked on the real website — each shows exactly
one tool again. The deeper problem is filed as bug 360: the page editor should refuse to touch a
retired section at all, and until that is fixed we re-check every retirement by hand at the end of
each swap. One bookkeeping correction that surfaced while proving the damage: the true count is
twenty-seven tools rebuilt, not twenty-eight — a counting slip in our own running tally, now
corrected against the database. Thirty-six remain: thirteen whose logic lives in separate files,
then the big ones.

## 2026-08-25 — the separate-files phase is done; thirty-eight of sixty-three

All the tools whose brains lived in separate script files are now rebuilt — eleven of them; the
twelfth turned out to be the micro-CMS, which is one of your five review-in-person applications,
so it waits for the finale. Two things from this stretch worth a sentence each. First, one of
those tools had never worked a day in its life: its missing script file meant no slider ever
moved on the original site, and its port notes said so plainly — the new one works and keeps the
"share this exact vibe" link promise honestly. Second, our own counting instrument had been
fooled by a code comment that merely *mentioned* a script tag, so the class was twelve, not
thirteen — the census, the correction, and the check that settles it are all written down.
What remains: twenty ordinary larger tools, then your five applications one at a time.

## 2026-08-25, mid-afternoon — the 41st tool, and a thing we have been quietly getting wrong for a week

Text Sanitizer is rebuilt and live. 41 of 63.

It was a good one to catch. The old version had a checkbox labelled "Remove Invisible Unicode",
and it did not remove them — it swapped each one for an ordinary space. That sounds harmless until
you read the tool's own teaching panel, which is *about* how an invisible character glued to a word
makes an AI read a different word. The old tool's fix turned "System" into "Sys tem". It broke the
word in a new way and then reported, in a green tick, that it had cleaned it. Its status bar also
said "Text is clean" on text it had just rewritten, because it counted four kinds of curly quote
while silently fixing twelve. The rebuild deletes the zero-width characters properly, normalises the
genuine spaces to a plain space, tells you exactly which characters it found and how many, and it
can only print "already clean" when the text coming out is character-for-character the text that
went in. That last part is the bit I care about: it is not a better-behaved message, it is a message
that is now impossible to get wrong.

**The other thing is more important, and it is a correction to us.**

Since 08-24 this lane has been carefully naming one or two related articles on each rebuild, so the
new tool gets a sentence woven into a relevant guide — the thing that turns 63 separate tools into a
site that hangs together. I checked today whether those sentences are actually appearing.

They are not, and they never have. Not once, on this site, since the 5th of August. Seventy-one of
these little "go and mention this tool over there" jobs have been created and every single one has
died in one of four ways. The eight I have filed personally are all sitting in a state called
"deferred", which I had inherited a note describing as "their normal gate" — as though it were a
queue they were waiting in. It is not a queue. It is a parking space. The row is created and parked
in the same second, with a flag on it that says, in as many words, do not promote this: dispatching
it would ask a handler to do something it is forbidden to do.

The reason turns out to be sound, and it is another team's work, not a fault. Our /learn/ articles
are marked "owned" — meaning a specific pipeline is responsible for their content and nothing else
may write into them. A guard enforces that. It was built deliberately, it is being actively proven
this week by the session that owns it, and it is doing exactly what it says. Our cross-mention jobs
walk into it and stop.

I nearly talked myself out of this, which is the part worth telling you. When I went to look at the
three articles in question, all three DID contain a link to their tool. For a few minutes that read
as proof the mentions were landing after all. They were not ours: those articles were last written
on the 15th of August, ten days before I filed anything, and the links were already in them — copy
that came across with the original port. A search for the tool's name on the page passes on content
that predates you by a week. The only honest version of that check is to look at when the article
was last written and compare it with when you filed.

So: nothing is broken and nothing is lost. We are still naming the right articles, the jobs are
still being created against the right pages, and when the route for writing into owned pages exists,
those parked jobs are precisely the raw material it needs — if we had skipped the step, there would
be nothing to pick up. What changes is only what we are entitled to say. Filing a rebuild delivers a
correctly-aimed job that is waiting, not a mention on a page. I have corrected our notes and fixed
the check in our runbook, which had been telling sessions to "verify at the artefact" and then
handing them a query that only counted the jobs — a test that could not fail. The team that owns the
guard has the details; whether to go and place those eight tools' mentions by hand is their call and
yours, not mine to fire off at eight articles unasked.
