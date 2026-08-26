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

They are not, and they never have. Not once, on this site, since the 5th of August. Eighty of
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

**Correction, 2026-08-25 15:40, to the entry just above.** The number of those cross-mention jobs is
**eighty**, not seventy-one. I wrote seventy-one an hour earlier from a query that could only see
part of the record — the table holding these jobs only keeps recent ones, and the older, finished
ones move to an archive I had not asked. Another session re-ran my query rather than believing it
and found the missing nine. Nothing about the story changes: still not one of them, in either place,
has ever succeeded. The extra nine are more evidence, not less. I have corrected the figure in the
paragraph above rather than leaving a wrong number sitting there, and this note is here because that
log is meant to be added to rather than edited, and the number was in a paragraph I had written
myself half an hour before.

2026-08-25 — Two thirds done. Forty-one of the sixty-three tools are rebuilt and live; the other
session does about five a day and nothing blocks the rest. This session spent the day on the
platform side: the checker that spots the two most common old-tool faults across every site is
built and before the review council (their first review asked a hard question and answering it
fixed two more checks that were wasting review effort on retired pages). One mistake worth owning:
I briefly recorded a "count changed overnight" that was really two of us using queries one hyphen
apart — caught by the other session refusing to believe a number it couldn't reproduce, corrected
everywhere, and the lesson written down in all three lanes involved. New summary file:
SUMMARY_2026-08-25_two_thirds_done_and_the_platform_pays_us_back.md.

## 2026-08-25, late afternoon — two more tools, two builds lost, and a rule I got wrong and had to take back

43 of 63 now. Cubic Bezier Architect and Golden Ratio Cropper are both rebuilt and live.

The old cubic-bezier had a line of text under its preview saying "Animation loops every 2 seconds",
and the code could not honour that the moment anyone touched it — every twitch of the mouse during
a drag queued another animation on top of the last, so a two-second drag stacked something like a
hundred and twenty overlapping timers. It also played the animation *backwards* half the time,
because it returned the preview to its start using the same easing curve, and for any lopsided curve
backwards looks like a different curve entirely. On a tool whose only job is showing you what a
curve feels like, that is the whole product being wrong.

The old golden-ratio one was the best find of the day. Its dropdown offered three overlays and the
code only knew how to draw two — so choosing "Phi Grid", the option named after the tool's own
subject, silently drew nothing at all. Its "Download Crop" button cropped nothing; it saved your
photograph with the guide lines burned into it, which is the one version a photographer cannot use.
And in small grey type at the bottom of a live public page sat the sentence: *"For this MVP, the
crop is visual. You can screenshot or we can implement full crop logic later."* That is a note from
a developer to a client, left where customers read it. The code comments are worth quoting too,
because they are an argument with nobody: *"In a real crop tool, we would crop. For now, we download
the image with the guide? Or just the image? Usually users want to CROP to this ratio."*

**Now the part where I was wrong.** Both of these tools failed to build on the first try. The
system that writes them has a size limit, and both of my instructions produced tools too big to fit —
the build stops and, annoyingly, the job still reports itself as finished, so you only see it if you
look at the run rather than the job. Nothing was damaged either time; the old tool stayed in place
untouched and the safety check I put in my own retirement step would have refused to proceed anyway.

After the first one I wrote a rule into our runbook: keep the instruction between about two thousand
and three thousand characters. The very next tool disproved it — an instruction of two thousand seven
hundred characters, comfortably inside my own limit, died exactly the same way. I had turned one
success and one failure into a threshold, which is not a measurement, it is a guess with a number in
it. The real problem was different and more useful: I had been spending half of each instruction
explaining what the *old* tool got wrong. The system building the new one has no use for that
history — it needs to know what to build. Once I cut the archaeology and wrote plain requirements,
the same tool described in fifteen hundred characters built first time. I have corrected the runbook
and left the wrong rule visible with its counter-example next to it, rather than quietly deleting it.

**One thing I want to flag rather than bury.** To get golden-ratio under the size limit I dropped the
download button entirely. So the new version has no download at all. My judgement is that this is a
net gain — the visitor loses a button that lied and gains three overlays that work — but it is a
capability removed to fit a budget, not a fault repaired, and those are different things. It is
written down as owed work, along with keyboard access for the cubic-bezier handles, which I cut for
the same reason. Both want their own small follow-up job rather than being smuggled into the next
tool's rebuild.

## 2026-08-25, evening (platform seat) — the proof-it-works check cannot run, because the inspector that would run it has been switched off since the 11th

Some context first. When we shipped the "contract rules" change earlier today — the one that files
one tidy work item per site instead of forty noisy ones — I said the proof would arrive on its own:
the next routine inspection of the webdesign site would exercise the new code, and we would simply
read the result. Tonight I went to read it.

The proof cannot arrive. The platform has several kinds of routine inspection, and the one that
carries our new code is the *design* inspection. That inspector has not visited any site since
August the 11th. It was switched off on purpose that day, during the cost scare, when you ruled
"turn the inspections back on, slowly". We turned the first one back on the next day, someone later
turned on the second — and nobody ever turned on the third. The off switch has now outlived the
reason it was flipped by about two weeks.

Two things made this hard to see, and I have written both up properly. First, my own handoff note
from this afternoon confidently said "the inspections are scheduled" — I never checked the switch,
and I have logged that as a wrong call. Second, the daily watchdog report that exists precisely to
catch a switched-off inspector has been saying "3 of 3 switched on" the whole time. It is
miscounting: a fourth, unrelated task was added later with a similar name, and the report counts it
as one of the three. So the alarm built for exactly this situation went quiet a week into it. That
is a real bug, now filed (number 401), with a suggested one-line-of-thinking fix: count by name,
not by number.

**The decision that is yours.** Turning the design inspector back on is the unfinished last step of
your own "slowly" ruling, so I have not done it without you. The cost of switching it on is the
same shape as the quality one you already approved: it works through the sites over a few days,
then settles to roughly four site-visits a day. Until it is on: our new contract-rules code sits
live but never runs; none of the 43 rebuilt tools has had its routine re-audit; and the design
checks (contrast, broken images, tool health) run on no site at all. The alternative half-measure
is to point one single inspection at the webdesign site by hand, which proves our code works but
leaves the fleet-wide gap standing. My recommendation is the switch: the reason it was off has
ended, and the watchdog that should have told us so is the thing that was broken.

## 2026-08-26, morning (platform seat) — done: the design inspector is back on, and it worked first time

You asked for the switch this morning, and it is flipped. Sixteen seconds after I turned it on,
the scheduler picked the site that had waited longest — agritec, which had never had a design
inspection at all — ran the full set of twenty-four design checks on it, and filed eleven
findings. So the mechanism did not rust while it was off: first pull of the handle, everything
moved.

What to expect now: one site roughly every three hours, longest-neglected first, so the whole
fleet gets its first visit over the next two to three days, then it settles to about four site
visits a day. The webdesign site — the one whose new contract-rules code we have been waiting to
see run — is a day or two down the queue, and I will read the result when its turn comes.

One honest note on cost and behaviour: the world has changed since the pause. Back then, findings
just piled up for reading. Today there is a promoter that picks up well-understood findings every
fifteen minutes and dispatches the repairs automatically. So switching the inspector on also
switches on a stream of small automatic fixes, not just reports. I said this plainly in the
notices I sent to the other fourteen working threads, per your instruction, and two of them wrote
back with useful confirmations rather than surprises — which is what the notices were for.

The watchdog bug that let this sit unnoticed for two weeks (it kept saying "3 of 3 switched on"
by counting a lookalike task) is still open as bug 401 — today removed the harm, not the defect.

## 2026-08-26, later morning (platform seat) — the proof arrived the same morning, and both checks passed

Better news than I promised, and one correction to my own account. The correction first: the
design inspections had in fact already partly resumed overnight, before my morning switch —
the improvement loop you turned back on yesterday evening carries the design inspection with it
on the sites it visits. Two other threads' replies to my notices are what surfaced that, which
is exactly what the notices were for. My morning switch still mattered, but for a different
reason than I wrote at nine: it restores the fair schedule — the guarantee that the
longest-neglected site is visited next — which the loop alone does not give.

And because the loop had already visited the webdesign site at a quarter to four this morning,
the proof we were waiting for is already in, days early. Both checks passed. The new code filed
exactly one tidy work item for the site's remaining un-rebuilt tools — and the number inside it,
twenty, matches the rebuild queue's real remainder to the digit. And the acceptance audit
covered the sixty-nine live tools while ignoring all forty-six retired ones, where the old code
would have listed every ghost. The change we shipped yesterday does what it was built to do, on
the site it was built for.

## 2026-08-26, early afternoon (platform seat) — the acceptance audit mostly fails healthy tools, and a diagnosis is already running

An honest walk-back of one thing I told you this morning. I reported approvingly that the
acceptance audit had covered the sixty-nine live tools. The coverage was real, but the sibling
seat has now shown the verdicts mostly are not: nearly every failure the audit has ever
produced — 110 of 112, across ten sites, back to early July — says a button or input is
"absent" from a page where it plainly exists. The cause looks like a naming mismatch: the
acceptance criteria are written in documents that call an element by its short name, while the
pages as rendered give every element a longer, prefixed name. A checker looking for the short
name reports it missing and fails the tool.

Two consequences. First, thirty-two of those failures already triggered automatic rewrites of
tools across the fleet that may never have been broken — that is being investigated by the
diagnosis loop right now, filed properly rather than asserted. Second, on the webdesign site
forty-one such failures are queued for the same automatic rewriting, aimed at the freshly
rebuilt tools — and the only reason none has fired is that the site's dispatch queue happens
to be stalled. The stall I reported earlier as blocking the rebuild work is, for the moment,
also the thing protecting it. Nobody is touching either half until the diagnosis comes back.

## 2026-08-26, late morning (platform seat) — the diagnosis came back: confirmed, quickly, with the receipts

The investigation I mentioned this morning finished in about forty minutes and confirmed the
whole story. The acceptance checker looks for buttons by their short written name; the
publishing machinery renames every element on the way out (a deliberate step, so two copies of
a tool on one page can't clash); nobody taught the checker about the renaming. So it reports
"button missing" on pages where the button is plainly there. The verdict cites the exact line
of code that does the renaming and a worked example page.

What changes with confirmation: the forty-one queued rewrite jobs on the webdesign site are
now known-false alarms rather than suspected ones — keeping them from running is a necessity,
not a lucky accident of the stalled queue — and the thirty-two rewrites that already ran
across the fleet are confirmed to have been triggered by false failures. The fix belongs to
the team that owns the checker; the evidence trail is on their desk. The rebuilding seat's
own browser-level checks, the ones that actually opened the pages, remain the only real
grading the tools have had — and those all passed.

## 2026-08-26, mid-afternoon (platform seat) — the stalled queue woke up, so the accidental protection is gone

A short correction to this morning's note. I wrote that the only thing stopping the forty-one
false rewrite jobs was that the site's queue happened to be stalled. That stall ended at 15:27 —
the queue is moving again, sixty-six items stand between the front and those forty-one, and the
rewriting worker has been taking jobs across the fleet this past hour. The sibling seat asked
you whether to hold them and you asked for a re-check instead; that re-check is theirs and the
contested action remains untaken. For the record, the two other sites that hit the same false
alarm have both already parked their copies with the release condition written on the row —
that pattern is ready if you want it here. Nothing has been damaged: the rebuilt tools were
re-verified intact this afternoon.

## 2026-08-26, late afternoon (platform seat) — the forty-one are parked, safely and reversibly

Closing the question my last note left open. You asked the rebuilding seat for a re-check
rather than choosing; they re-checked, found the risk live (the queue moving, the rewriting
worker taking nine jobs an hour), found the parking pattern proven on the two other affected
sites, and acted on their own site's queue at a quarter to five. All forty-one false rewrite
jobs are parked, none cancelled — nothing destroyed, the evidence of scale kept, each row
carrying on itself the condition for release: the checker mismatch fixed and the acceptance
audit re-run, and if a failure still reproduces on the real page then it was real after all.
Undoing the whole thing is a single database statement, which is exactly why this shape was
safe to take without waiting. The fifteen similar rows on other sites were deliberately left
alone — those belong to the checker team's cleanup. Nothing else changes: the checker fix
itself is still with the team that owns it, and the rebuild queue remains blocked behind the
sweep backlog.

## 2026-08-26, early evening (rebuilding seat) — the queue cleared, and the forty-fourth tool is built

The last note ended by saying the rebuild queue was still blocked behind the sweep backlog. That
stopped being true within the hour, so here is the correction and what I did with it.

When I picked this up the backlog ahead of a new job had fallen from about a hundred and seven to
ten, and the site was working through them steadily. I waited, watched it drop to eight, and filed
the next rebuild. It was picked up fourteen minutes later and built in three and a half minutes.

The tool is the Monolith Splitter — the one that writes you a careful instruction for an AI
assistant when you want a huge source file broken into smaller pieces. The old version had a real
trap in it: the instruction panel was written once when you pressed a button and then left alone,
so if you afterwards changed the framework, or the file name, or any of the component names, the
page carried on showing an instruction for values that were no longer on screen. You would copy it
and not notice. The new one has no button at all — the instruction is rebuilt from the form every
time you type, so the two cannot disagree.

Three smaller things went with it. The copy button used to decide whether there was anything worth
copying by looking for a sentence of placeholder text on the page, which meant it silently did
nothing and would have broken the moment anyone reworded that sentence; it now looks at whether you
have actually entered any components, and says so on screen when you have not. It used to announce
"Copied!" whether or not the copy had worked; success and failure are now different messages. And
it used to restore itself afterwards to a dark grey it had never been, so a single copy left the
button permanently the wrong colour for the rest of the visit.

I checked all of that by reading the code that makes each decision, not the summary the tool writes
about itself, and I checked each "this is now absent" against the old version to be sure the absence
meant something. One of my checks was worthless and I have said so in the technical log: I was
testing for something that was not in the old version either.

**What is not finished.** The page on the web still shows the old tool. The database is correct and
the old version is retired, but the page only changes when the site next rebuilds that page, and
there are about eighty jobs in front of it — roughly three hours. Nothing is broken or half-done in
the meantime: the page shows one working tool now and will show the new one when it rebuilds. I
have left the exact job reference in the notes for whoever is next.

**One thing worth your attention.** The check we run before filing anything — the one that asks "how
many jobs are ahead of me?" — has been asking the wrong question. It counts jobs that cannot
actually run, because they have already failed twice and are sitting in a retry wait, and at the
same time it ignores a whole category it should count. Today the gap was small, ten against a true
eight, so nothing went wrong. But the two places in the code our runbook cites for how the queue is
ordered turn out to be two different queues entirely, neither of them this one, so the rule we have
been following was written from the wrong source. I have found and recorded the right one and
written a corrected check.

The more useful half of that discovery: what really decides whether your job runs soon is not how
many jobs are ahead of it on your site, but whether your site gets picked at all — and a site is
skipped entirely while any one of its jobs is stuck part-done. I checked whether that is biting us
anywhere today and it is not, on any site, so this is a trap to know about rather than a fire. I
have been careful not to let it explain last night's long silence, because I have no evidence that
it does.

## 2026-08-26, early evening (rebuilding seat) — a second tool the same hour, and the corrected check earned its keep

The queue stayed clear, so I took the next one straight away. Head Architect — the tool that writes
you a complete HTML head block for a page from four fields plus whatever head you paste in.

This one was worth doing carefully, because everything it produces is code the visitor pastes into
their own website, and the old version had two faults that would quietly damage that code.

The first: none of what you typed was made safe before being put into the markup. Type a brand name
with a quotation mark in it — "Bob's \"Best\" Widgets" — and the tag it lands in breaks. Every tag
carrying that name breaks with it, and you would only find out after pasting.

The second is the same problem one level down. The tool writes a block of machine-readable data
about your site for search engines. If your page description happened to contain the characters that
close that block, the block ended early and the rest of your description spilled out as broken
markup. The new version encodes those characters so the data survives intact and the block cannot be
broken out of. It went further than I asked for, which I have noted.

I also found the handoff note I was working from had one defect backwards. It said pasted titles come
out duplicated; reading the code, they are silently thrown away instead — along with any Facebook or
Twitter tags you already had. Those are opposite problems needing opposite fixes, so I corrected the
note and specified the real one. The new tool now says on the page what it keeps and what it replaces,
and reports anything it did not recognise rather than dropping it in silence.

**Speed, which is the visible change.** This one was picked up two minutes and fourteen seconds after
I filed it, built, and the old version retired nineteen seconds after the build finished. Earlier
today the same cycle took fourteen minutes to even start. That is the corrected queue check doing its
job: it now tells me not just how many jobs are ahead but whether this site is the one about to be
served, and I filed at the moment it was first in line.

**Where that leaves us: forty-five of sixty-three.** Both of today's two are waiting on the site to
rebuild their pages before the new versions are visible to the public — the database is correct and
nothing is half-done, but the pages themselves still show the old tools for now. I have left both job
references for whoever picks this up.

One thing I want to be straight about: parking those forty-one false alarms bought time, it did not
fix anything, and the number will keep growing by one for every tool we rebuild. I confirmed that on
today's tool by checking both halves — what the page actually serves and what the checker actually
looks for — and they do not match, for the same reason as all the others.

## 2026-08-26, evening (rebuilding seat) — a third tool, and the queue taught me something by embarrassing me

Forty-six of sixty-three. The third one today is the Asset Formatter: the tool that writes an
instruction telling an AI assistant to use your real hosted images rather than inventing stock
photos. Its old version had a genuinely silly fault — press Copy with nothing filled in and it put
the words "No assets mapped." on your clipboard while telling you the copy had succeeded. It also
promised to make your image names safe and did nothing of the sort, and quietly threw away one of
two images if you gave them the same name. All three are fixed, and I checked each by reading the
code that makes the decision rather than the tool's own description of itself — which matters here,
because this tool's self-description reads as though the old fault had survived, and it hasn't.

**The embarrassing part, which is the useful part.** Earlier this afternoon I found a trap in how
work gets handed out: a site is skipped entirely while any one of its jobs is still in progress. I
checked whether it was biting anyone, found it wasn't on any site, and wrote it down as a thing to
know about rather than a live problem. Thirty-two minutes later it was biting me. I filed the third
tool with nothing ahead of it in the queue and it sat there for ten minutes, because one unrelated
page-building job on the same site was mid-flight.

Nothing was broken — that job was doing exactly what it should, and mine went through the moment it
finished. But it's a clean demonstration that a measurement of *what is happening right now* goes
stale faster than you'd think, and I'd published mine as if it would hold. I've added the second
query to the runbook so the next person checks both: how many jobs are ahead, and whether the site is
currently able to take work at all. Those are different questions and only one of them was being asked.

**One correction to my earlier note.** I said the pages would take about three hours to rebuild. They
are draining faster than that — the site is getting through roughly forty jobs an hour, so it is more
like an hour or two. All three of today's tools are still waiting on that step, and all three pages
are safe in the meantime: each shows one working tool, just not yet the new one.

## 2026-08-26, evening (rebuilding seat) — forty-nine of sixty-three, and the first one is live to the public

Three more since the last note: the CSS Grid Layout Generator, the Insight Injector, and the Privacy
Redactor. Forty-nine of sixty-three done.

**The most important one is the Privacy Redactor**, and it is worth explaining because the old version
was quietly unsafe. It is the tool you paste a log or a support ticket into to strip out personal
details before sharing it. Its switches are labelled Emails, Phones, IPs and Cards, all on by default,
under a heading saying "Intelligent Detection".

I tested it rather than read it: I took its actual detection rules out of the page and ran them over
real examples. **It missed American Express card numbers entirely** — they are fifteen digits and the
rule only ever recognised sixteen — so an Amex number went through untouched with "Cards" ticked. It
missed modern IPv6 addresses completely. And it went the other way too, blacking out ordinary
reference numbers that happened to be the right length. Five of the six examples I tried came out
wrong.

It also never told you the text stays on your machine. It genuinely never sent anything anywhere — but
for a tool whose whole pitch is "paste your sensitive logs here", the one sentence a person needs
before pasting was missing. The new one leads with it, checks card numbers properly, handles both
kinds of IP address, and says plainly that pattern matching cannot catch everything so you should read
the output yourself.

**And the first rebuilt page is now live to the public.** The Asset Formatter's page rebuilt at ten
past seven and I checked the real page on the web: the old tool is gone, the new one is there, and the
specific fix I made is visible in what the public is actually served. That closes the loop for the
first time and gives me confidence in the other five, which are queued and will follow over the next
hour or two.

**Two things I got wrong, both worth telling you about because they had the same shape.** My check of
that live page came back with a perfect score — and it was checking a "page not found" error page. An
error page contains none of the things you are looking for, so it passes every test you have. I had
even printed the "404" on the line above and read straight past it. Separately, my wait-for-the-page
loop treated a missing timestamp as "it's ready", because the command I used quietly returns the
current time when handed nothing.

Neither did any damage, and the page turned out to be fine — the 404 was a few seconds of publishing
lag, which I confirmed by checking two other tools nobody had touched. But both mistakes produced a
confident *green* result rather than an error, which is the kind that gets believed. I have written
both up and made the check refuse to run at all unless the page loads properly.

## 2026-08-26, evening (rebuilding seat) — all six are live to the public, and nothing is left hanging

Following on from the last note: the remaining five pages have now rebuilt, and I have checked every
one of them on the live site rather than in the database. All six of today's tools are serving the new
version, none is serving the old one, and none of the pages is broken. Forty-nine of the sixty-three
are now done and confirmed live.

I checked the specific fixes are actually in what the public receives, not just in what we built —
those are different things, and only the second one matters to a visitor. The escaping in the head
builder, the mobile breakpoint in the layout generator's output, and the card and IP-address handling
in the privacy redactor are all present in the served pages.

**Two of the six checks refused to run, and both refusals were correct.** One page returned "not
found" for a few seconds immediately after publishing — I confirmed it was a brief publishing lag by
checking two other tools nobody had touched, and it was serving normally moments later. The other was
still showing a five-hour-old copy because the new one had not finished uploading; the check spotted
that the page was stale rather than grading the old version by mistake.

That second one is worth a word, because it only worked by luck of timing. An hour earlier I had
found and fixed a bug in that very check: it was reading our database timestamps as if they were in
British Summer Time rather than UTC, which made every page look a full hour fresher than it really
was. Left alone, it would have waved through anything up to an hour out of date — including the stale
page it correctly caught tonight.

I have written that trap up properly and moved the checking script out of my temporary working area
into the project itself, so the next person has it rather than a dead link to a folder that no longer
exists.

**Where that leaves us: forty-nine of sixty-three, all live, nothing owed, and the queue is flowing
normally again.** Fourteen tools remain, five of which are the larger applications you asked to be
done last and one at a time with your review.
