# Where we are — bug 209, the image deploy that picks its source by "purpose"

Plain-prose log, append-only, newest at the bottom.

---

## 2026-08-08, late evening — first look, and it turned out better and worse than expected

Picked this up because the previous session's handoff said 209 was the only bug in
the open pile that nobody owned. Checked that myself first, since the ownership
script reads commit history and can't see a session that's mid-fix. Three other
sessions were live at the time; none of them was working on this. One was my own
predecessor, which is why the script pointed at me.

The bug is about how we deploy an image. When the system deploys, say, a hero
image, it has to find where that image is stored. One of the ways it looks is by
asking "what's the URL for the *hero* image in this run?" — and that's the
problem: "hero" is a category, not a name. If a run produced two hero images, the
second one quietly overwrites the first, and a deploy asking for "the hero" gets
whichever came last. You'd deploy the wrong picture and nothing would complain.

**The good news: I can't find any way this actually happens today.** I went
through every live workflow that can deploy an image — there are three, plus one
more that was the prime suspect. The two workflows that do handle several images
in one go handle a *hero* and a *logo*, which are different categories, so they
never tread on each other. The suspect workflow does have two steps that both
claim the "hero" slot, but they sit on opposite sides of a fork in the road —
only one can run in any given job. And it doesn't deploy anything itself; it hands
off to a separate deployer, passing the exact image it just saved, by name rather
than by category. I also checked what that deployer actually has in memory when it
runs: the category-keyed values simply aren't there.

So the fault is real in the code but currently unreachable. Latent, not live.
That's a downgrade in urgency, not a clean bill of health — the door is still
open for the next workflow anyone writes that makes two of the same kind of image.

**The bad news, and it's the useful part: the fix the bug file recommends would
cause the very bug it's trying to prevent.** The file's first suggestion is to
delete the category lookup and rely on the image's ID instead. But the two older
workflows don't tell the deploy step which fields to use, so the system falls back
to rummaging through everything in memory looking for anything called "asset_id" —
and the order it rummages in is deliberately randomised by Go. I ran the real code
400 times on identical input: **the step deploying the logo picked up the hero's ID
344 times out of 400.** So the category lookup, which looks like legacy cruft, is
actually the thing keeping those two workflows correct — because for them, the
category *is* the difference between the two images.

Given that, I've deliberately not written a fix tonight. Changing how a shared
piece of machinery resolves things, to fix something that can't currently happen,
on the evening I just discovered the obvious fix is the harmful one, is how you
turn a latent bug into a live one. What I've done instead is lock the behaviour
down with tests that record *why* the lookup was kept, so the next person who
reads the bug file's ranking and reaches for the obvious fix hits the 86% figure
before they ship it.

**One thing I got wrong and want on the record.** I first wrote that those two
older workflows "haven't run in 26 days", because the run table has rows going
back to mid-July. Then I checked what those old rows actually are — and they're
all cancelled or stuck jobs. Completed runs get cleared out after about a day. So
the honest statement is "they didn't run today", which is a much weaker claim, and
I'd have been quoting a made-up month if I hadn't looked. I also tried a second,
longer-lived log to settle it properly, and threw the answer away: it reported
zero runs for a service I'd just watched run sixteen times, so it can't see these
agents at all. A measurement that can't come out the other way isn't evidence.

What I'd want a decision on, when there's appetite: whether those two older
workflows are actually dead. Nothing schedules them any more as far as I can see.
If they're genuinely retired we can delete them, and then the clean fix for this
bug becomes available with no awkward legacy constraint. While they're merely
dormant, we have to keep supporting them.

## 2026-08-09, morning — checked everything again after the new build, still holds

A fresh build went out this morning. Nothing of ours is in it — we shipped tests
and documents, not product code — but the deployment also re-applies the workflow
configurations, and our whole "this can't currently happen" conclusion rests on
those. So I re-read them, comparing content rather than trusting timestamps
(the timestamps all changed, because re-applying identical config still stamps
it). Every fact the conclusion depends on is unchanged, character for character.
The tests still pass. Nobody touched the relevant code overnight. The conclusion
stands on the new build.

One sentence from yesterday deserved tightening, and I've tightened it in all the
records. I said removing the category lookup "swaps a correct lookup for a
mostly-wrong one" — true, but it's the *backup* route that gets swapped, not the
main one. The older workflows normally find their image through a direct pointer
that the proposed fix wouldn't touch; the category lookup is what catches them
when that pointer is missing. So the danger is conditional: the bad fix only
misfires on the day something else has already gone slightly wrong. That doesn't
change the advice — a safety net that grabs the wrong thing exactly when you fall
is worse than useless — but the record should say precisely where the danger
lives, and now it does.

## 2026-08-09, mid-morning — owner's answer received; the problem restated plainly before any fix

**The owner's ruling:** pageflow-builder and site-work-orchestrator are not dead,
but not being worked on. If we need to diverge from them, we can use new actions
and workflows as necessary. Explain the problem clearly before moving further.

So here is the problem, stated to be read aloud.

**What the deploy step has to do.** When a workflow says "deploy this image", the
deploy step must first find where the image's bytes actually live in storage.
It has three ways of finding out, tried in order: first, an address handed to it
directly by the workflow; second, a set of lookups in the run's shared scratchpad,
all keyed by the image's *category* — hero, logo, and so on; third, the image's
own database record, looked up by its unique identity.

**The defect.** The scratchpad has one slot per category, and the last writer
wins. If one run ever saves two images of the same category and then deploys them
through the scratchpad route, both deploys get the second image's bytes. The
first image's address is simply gone — overwritten — and nothing errors, because
a perfectly valid address came back. It is just the wrong image's address.

**Why nothing is broken today.** Measured twice now (and re-checked after this
morning's deployment): no live workflow saves two same-category images and then
deploys through the scratchpad in one run. The two older workflows handle one
hero and one logo each — different categories, different slots, no collision.
The modern path hands the address over directly, by identity, so the scratchpad
is never consulted. The fault is a loaded gun in a drawer nobody currently opens.

**Why the obvious fix is the dangerous one.** The natural instinct — and the bug
file's original recommendation — is to delete the category route and rely on
identity. But identity is only trustworthy when a workflow *names its inputs*,
and the two older workflows don't. For them, "find the identity" means rummaging
through the whole scratchpad for anything called asset_id, in an order the
language deliberately randomises — and with two images saved, that rummage picks
the wrong one 86 times in 100 (measured on the real code, four hundred runs). So
for precisely those two workflows, category is the correct key and identity is
the broken one. Deleting the category route wouldn't hurt them on a normal day —
their direct pointer still works — but it would replace their correct safety net
with one that grabs the wrong image exactly when the direct pointer fails.

**What your ruling means for the fix.** Keeping the older pair alive but
untouched fits the right fix perfectly, because the right fix is divergence, not
surgery. The shape I propose, for approval before anything is built:

1. **An opt-in switch on the deploy step, default off.** Off means today's
   behaviour, bit for bit — the older workflows never set it and never notice.
   On means: resolve by explicit address or identity only, never by category, and
   fail loudly rather than guess. The modern workflows — which already pass
   identity — switch it on. This follows the estate's own 2026-08-02 ruling that
   new authority on shared machinery ships as an opt-in field with the unsafe
   default off. The characterisation tests we shipped yesterday exist precisely
   to prove the off position stays identical.
2. **Alternatively, an entirely new action**, leaving the old one untouched. Zero
   risk to the frozen pair, but it duplicates the download-optimise-commit logic
   and becomes a second thing to maintain. I'd only choose this if you prefer the
   old code literally unedited.
3. Either way, the writers additionally record each image's address under its
   *identity* in the scratchpad, so an in-run identity route exists at all. That
   part touches shared machinery and goes through the register and the council,
   with the other consumers told.

No code has been written. Awaiting your go-ahead on shape 1 versus shape 2.

**Separately, the question you asked about bug 064.** Nobody is working on it —
and strictly, 064 itself is closed (July 24th). What's failing today is a
*recurrence* of its shape, the second one: on the evening of August 8th the
**idea_uk_vm_site** thread shipped its "decision records" change (RFC_015), which
taught the database a new document type — "decision" — but not the Go code's list
of valid types, and a test that exists precisely to catch that split has been red
ever since. Two other threads (this one, and the chrome-divergence one) have both
tripped over the red test and correctly recorded it as not theirs. The
idea_uk_vm_site thread's own notes don't mention it, so they likely don't know.
The fix is one word in one list, plus the closed bug's own checklist — it belongs
to that thread, and they should be told.

## 2026-08-09, late morning — you asked what "into line" would cost, and checking surfaced a real bug

The costing question forced one measurement I hadn't made: when the older
workflows' logo step runs, what category does the deploy code *actually think*
it's handling? The answer is **"hero"** — and it has been since February.

Here's why. The modern input machinery fills in a default of "hero" for the
category before it reads anything else, and every later step of the resolution
refuses to overwrite a value that's already set. The older workflows state their
category as a plain word in their config — "logo" — and a plain word is exactly
the one shape the machinery can never read. So the default wins, silently. The
code's own last-resort check ("if no category was found, read the config") never
fires, because the default means a category is always found. Run the logo step
today and it would resize the logo as if it were a hero image and — worse —
write the logo's bytes over the hero image's file, while the page goes on
pointing at a logo file that never gets written.

I proved this by running the real resolution code with the real live config, not
by reading it, and the proof is pinned as tests. Whether it ever actually
happened on a live site is unknowable — the run records only keep a day, and
these workflows haven't run in that window. Filed as bug 231; the wider question
("how many OTHER workflow configs have a plain-word setting that's silently
losing to a default?") has gone to the diagnosis loop rather than being guessed at.

**What this does to your question.** It roughly halves the honest cost of "into
line", because the biggest piece was already owed:

- **The repair the pair needs anyway** (to be runnable at all, which "not dead"
  implies) is a config-only change: four deploy steps, four pointer-style paths
  each, one migration. That same edit IS the "into line" step — it makes every
  input explicit and deterministic, kills the February bug, and kills the
  86%-wrong rummage in one go.
- **The only cost unique to "into line"** is then a deletion: remove the
  category-lookup fallback from the deploy code (~90 lines), through the council,
  built and rolled *after* the config migration is applied — order stated in the
  plan. That ends bug 209 for everyone with one code path and no opt-in switch.
- **Divergence now buys nothing.** It was designed to avoid touching the frozen
  pair; the February bug means we must touch their config regardless. Once we
  have, the opt-in switch protects nothing the deletion doesn't handle cleaner.
- **What into-line does NOT require:** touching how the workflows generate or
  store images, or the six other places each workflow references these keys —
  those stay as they are (tidying them is optional and can wait indefinitely).
- **Verification:** one throwaway-domain run of each workflow, checking the hero
  and logo files both arrive with different bytes. That also finally satisfies
  bug 209's own "verify on the real workflow" bar.

So: recommendation changed from yesterday. **Into line, two phases (config
migration, then the deletion), with the config phase doubling as the bug-231
repair.** Awaiting your go-ahead.

## 2026-08-09, midday — Phase 1 is applied and verified; Phase 2 is teed up

You approved the into-line fix, Phase 1 first. It's done and live.

What changed: the four deploy steps in the two older workflows no longer say
"my image is the logo" and hope the system agrees — each now points directly at
the exact image its own store step just saved: its category, its storage
address, and its database identity, all read from that one place. The dead
static settings and the old generator-pointer are gone.

How it went in: as a guarded database migration. Before applying it I made the
safety check fail on purpose against the un-migrated rows — a check that has
never failed tells you nothing when it passes — then dry-ran it in a throwaway
transaction, then applied it scoped so the other teams' pending migrations were
untouched. Verified afterwards by reading the four steps back and confirming
the neighbouring steps didn't move. One deliberate behaviour change worth
knowing about: if a store step ever fails, the deploy now skips with a visible
"nothing to deploy" instead of quietly deploying whatever the generator
produced — the old route bypassed the asset record entirely, which is the exact
family of fault this whole effort exists to remove.

This one edit closes three things at once for those workflows: the
February logo-as-hero bug, the random rummage that picked the wrong image's
identity 86 times in 100, and their dependence on the category-keyed lookup.

Also in: the diagnosis loop's verdict on the wider question ("how many other
configs have a plain-word setting silently losing to a default?"). It
independently confirmed the mechanism is real, but couldn't finish the
fleet-wide count — interestingly, because of the same indexing blind spot
another bug (223) already describes: the index that both the verifier and the
diagnosis loop read cannot see plain variable declarations, so the loop
couldn't fetch the very list of defaults it needed. Two consumers of one gap
now; noted on that bug. The fleet-wide count remains open under bug 231.

Still owed, in order: a real proof run — build a throwaway site through each
old workflow and check the hero and logo files both arrive, different bytes
(that also finally satisfies bug 209's own verification bar) — and then
Phase 2, the ~90-line deletion of the category lookup from the deploy code,
which goes through the council and must only roll after today's migration is
confirmed in place (it is; re-check on pickup as rolls re-stamp the config
timestamps without changing content).

---

**2026-08-09, afternoon.** The proof run happened, and it turned up something we
did not go looking for.

First the good news. I built a throwaway site at cookly.uk through the old
pageflow-builder workflow, start to finish — planner, logo, hero, three pages,
design, deploy. Both images arrived in the repo, each from its own source: the
hero step downloaded the hero's own file from storage and stamped the hero's own
database row, with the logo's file sitting right there in the same run untouched.
That is exactly what this morning's migration was supposed to guarantee, and it is
the first time we have watched it happen rather than reasoned about it.

Two things I want to flag about the evidence itself, because both are the sort of
thing that quietly turns into folklore.

One: the test everyone had written down as the bar — "check the two files have
different bytes" — is not actually a test. The system re-encodes the hero as a JPEG
and the logo as a PNG, so the two files are guaranteed to differ no matter which
source each one came from. It would have passed even if the bug were fully present.
The real evidence is which file each step fetched, which I have from the logs.

Two: a caution for anyone doing this next. The per-agent pods keep about ELEVEN
SECONDS of logs. I fetched twenty minutes' worth about a minute after the fact and
got eleven seconds. Start the log capture before you dispatch, or you will be
reconstructing from the database.

Now the thing we did not expect. While checking that our new logo looked right, I
listed every logo this system has ever committed across all the sites. Four are
400×400 PNGs, which is correct. **Eleven are JPEGs at the wrong size** — 1408×768
or 900×900 — and every single one of them was committed by a step that announced
itself as "Deploy **hero** image". The commit message is generated from the same
setting that controls the resizing and the filename, so those messages are not a
labelling quirk: those logos were genuinely processed as though they were hero
photographs. Because JPEG cannot store transparency, every one of those sites is
serving a logo with a solid background baked in, at up to three times the intended
size. The affected list includes idea.uk, robot-hands, dartsonline, fundamentallyai,
webdesign.uk and webdesign.co.uk.

That is bug 231 — the one filed this morning saying a setting written one way in
the config is silently ignored. Its file honestly says "we don't know whether this
ever actually happened". It did, eleven times, and it is on live sites now.

The sting is that **this morning's migration did not fix it.** The migration
repaired the two old workflows, which are the ones nobody runs. The damage came in
through a different door — the evidence points at the shared asset-deployer being
called without the setting it needed, and I can see four database rows whose
filenames are literally an unresolved config path, which is the same failure wearing
a different hat. I have not proved which caller did it, and I am deliberately not
guessing: that question is going to the diagnosis loop, which is what it is for.

So: the fix we shipped this morning is real and now proven, and it is narrower than
the problem. Next is the diagnosis run on the historic cause, the same proof through
the second old workflow, and then the code deletion that was already queued.

---

**2026-08-09, evening.** The lane's original job is finished, and it is live.

The code deletion shipped on the new chassis build. I checked it the only way
that actually proves a deletion: I looked inside the running binary on both
replicas for the text of the line we removed, and it is gone — with two other
lines we kept still present, so I know the search itself works. The council
approved the change, with one minor comment that two of my six edits were
comments rather than code, which is fair and harmless.

Then the logo bug. This morning we knew eleven sites were serving a logo that had
been processed as though it were a photograph, but not why. Today we found out:
one setting, written once, on a piece of the pipeline that handles two different
jobs — logos and page-top photographs. It said "photograph" permanently, so every
logo that went through it came out as a photograph: wrong shape, wrong file type,
and no transparency. The work item passing through it had been saying "this is a
logo" all along; nobody was reading it.

The fix is that it now reads the item. One line of configuration, live
immediately, no rebuild needed. Before changing it I checked the obvious risk —
that the photograph side of the branch might depend on the old hardcoded value —
and it doesn't: every item of both kinds carries its own label. I also proved the
safety check could actually fail before trusting it to pass.

**What that fix does not do is repair the eleven sites already out there.** It
stops new ones. Re-making the existing logos is the next job and it needs care in
a specific order, because the old wrongly-named file doesn't get deleted by a
re-deploy and some pages still point at it. Doing that in the wrong order would
leave a site with a broken image where it currently has an ugly one.

Along the way, three domains turned out to have problems worth knowing about.
cookly.uk is live and serving. www now works on cookly and dartsonline — it never
had, on any of our sites, for a reason nobody had noticed. loanzy.uk had been
pointed at Cloudflare years ago but never actually set up there, which produces a
uniquely unhelpful failure: it looks configured and simply hangs. And lendzy.co.uk
was completely down — every visitor got an error page — with every internal
status green. That last one bothers me most, so I've written it up: we have no
check anywhere that asks whether a finished site can actually be loaded.

Everything is committed and there is a fresh handoff for continuing.

---

**2026-08-10 (late morning).** The logo fix works — proven properly this time, not
just believed. I sent one logo rebuild through the repaired path on cookly.uk and
it came back as a PNG at logo size, through the exact branch that used to get it
wrong. The run recorded its own routing decision, so there is no ambiguity about
which path it took. That is bug 235's fix demonstrated end to end.

Two things about the plan we had written down turned out to be wrong, and both
would have wasted someone's time.

The acceptance test we'd recorded says the logo should come out "400×400". It
doesn't, and shouldn't — logos are fitted into a 400-pixel box keeping their
shape, so a wide wordmark comes out 400×218. Anyone testing for a literal square
would have failed a perfectly good result. What actually distinguishes right from
wrong is that it's a PNG rather than a photograph-shaped JPEG, and that the
database row says "logo" rather than "hero".

And the list of eleven sites needing repair was wrong in two places. idea.uk is
fine — it serves the correct logo and its pages point at it; it just has an old
unused file lying around. Meanwhile relojistas.com has exactly the problem and
wasn't on the list at all. So it's nine sites with damage a visitor would see,
plus two where the database is wrong but the impact isn't clear.

Then the day changed shape. The rebuild I'd queued sat there for twenty minutes
when it should have been picked up in two, so I went looking. **The service that
runs every timer on the platform had been dying and restarting 132 times in
thirteen hours** — alive about one minute in six. Nothing was down, so nothing
alarmed; work was simply arriving at a tenth of the normal rate.

The cause is a blank setting in shared plumbing. Every service here asks Kafka
for information about *every* topic in the cluster, every few seconds, in the
background. That was harmless when there were a few hundred topics. There are now
over twenty-five thousand, because each job step creates a disposable one and
nothing clears them up. You authorised me to clear them, and it worked: the
scheduler has now been up for eleven minutes instead of seventy seconds, using
half the memory it was, and the backlog of overdue timers has drained.

That's a one-off clean-up, not a fix — the setting is still blank and the topics
will start piling up again. I've written up what needs doing, in order, and the
one that actually closes the door needs a review round because it touches
plumbing every service depends on.

I should also own three mistakes, because two of them nearly mattered. I wrote a
number into the bug file before the thing measuring it had finished — the answer
held, but I'd invented a stronger version of it. Worse, the command I was using to
count topics **silently returns a short answer** when there are this many, with no
error at all, and I read a trend into that noise and wrote it up as a correction
before catching it. And I filled up a small disk on the Kafka machine with my own
debugging files, then spent a while diagnosing the resulting nonsense as a
different problem entirely. All three are written down where the next person will
hit them. The near-miss worth naming: my first attempt to delete topics was scoped
by one of those short lists, and it stopped only because of a safety check I'd
added for an unrelated reason. That was luck. It's now a deliberate check.

The logo work itself is unblocked and is the next thing.

---

2026-08-11, morning. The last logo is done — all eleven sites now serve a
proper PNG logo, and relojistas was fixed the right way rather than by hand.

The thing that was beating us on relojistas turned out to be one missing line
of configuration. When the dispatcher hands work to the deploy agent, it
passes the work order nested inside an envelope — and the deploy agent's
"which kind of image is this?" setting was looking on the outside of the
envelope. Finding nothing there, it fell back to its built-in default of
"hero", which overrode the correct answer written plainly inside. The fix
adds one line telling the dispatcher to also copy that answer onto the
outside, where the agent already looks. The sister dispatcher elsewhere in
the system already did exactly this, which is how we know it was an
oversight and not a design decision.

One thing worth owning: the plan I inherited from yesterday evening proposed
a different fix first — and that fix could never have worked. The bug's own
case file, three paragraphs earlier, explains why not, and there is even a
test in the codebase whose name says so. Nobody had joined the two up. I
checked the deciding code before building anything, so it cost nothing, but
it is a sharp example of why we read before we write.

Proof it works: the re-run deploy committed "Deploy logo image" where the
two failed runs had said "Deploy hero image", and the live site now serves a
400-wide PNG with transparency, seconds after the deploy. The site's pages
are being re-rendered now to point at the new file, and the one other site
that borrowed relojistas' logo for its portfolio has been corrected and
queued for the same treatment. The old wrong file stays where it is until
the estate-wide check for stragglers — deleting it early is how you break
someone else's page.

Also from this morning's checks: the memory-limit fix for the Kafka
scheduler went out with the overnight release and the scheduler is healthy.
The twice-daily topic-sweep cron I installed has a blind spot nobody
mentioned: this machine sleeps overnight, and a slept-through slot is simply
skipped, so only the lunchtime run is real. Worth knowing when we read its
log.
