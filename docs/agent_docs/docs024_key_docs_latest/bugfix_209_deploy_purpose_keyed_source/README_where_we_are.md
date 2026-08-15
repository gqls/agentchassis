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

---

2026-08-11, later. The logo saga is done (11 of 11), so this session went after
the question the class fix left open: how many OTHER places in the fleet have a
config value that silently does nothing because a built-in default wins?

We built the counter rather than counting by hand — a new mode on the existing
config audit tool that asks the running code for every built-in default and
then checks every live agent's config against them. First fleet-wide run:
about a hundred config entries are dead in the harmless sense (they restate
the default, so nothing differs today), twenty-four say something DIFFERENT
from the default they're shadowed by. The important step was checking each of
those twenty-four by hand: twenty of them turn out fine, because those
actions read their config through a different door that isn't affected. Four
are real.

The four real ones are all the same bug: we have four different auditor
agents (brief fidelity, content quality, site review, visual design), and
each one is supposed to sign its findings with its own name. All four
signatures are dead — every finding any of them has ever filed is signed
"design-audit". About 136 findings back to April are a merged pile nobody can
split by author, and one row from July literally has "brief fidelity" in its
type and "design-audit" in its signature. This matters beyond tidiness: the
213 work (the "wrong grader" bug) depends on that signature to tell producers
apart, so we've told that thread. The fix is small (four lines, one file) but
that file is 213's territory while their round is open, so it's recorded and
offered, not done.

Decisions you may want to weigh in on: (1) whether the four-line per-action
fix ships now or waits for the general "config beats default" change — the
census says the general change would alter behaviour in exactly these four
places and nowhere else, which makes it much less scary than it sounded;
(2) the stale logo.jpg deletion from the last session is still waiting on
your call.

---

2026-08-11, evening — your three decisions, recorded.

You ruled: (1) fix the auditor signatures now with the small four-line change,
and still do the general fix afterwards — not one or the other; (2) ship the
general fix, so that from now on an explicit config value beats a built-in
default everywhere, rather than only being detected; (3) leave the stale
logo.jpg files alone.

Worth saying plainly what (2) buys, because it is the bigger of the two. Today
we can *detect* this class — the new audit fails loudly when someone writes a
config value that a default will silently eat. After (2), that kind of mistake
mostly stops being possible: the config wins, which is what everyone writing
it already assumed. The one cost is that a handful of config values that have
been quietly dead for months will come alive on the day it ships, so the next
session re-runs the census immediately before implementing rather than
trusting yesterday's count — and by then the auditor fix should have taken
those four to zero anyway.

Nothing is implemented yet: this session ended at the decision point. The next
session picks up Task A (the four-line auditor fix, coordinating with the 213
thread who own that file) and then Task B (the general fix, which needs a
council round). Both are written out step-by-step in
HANDOFF_2026-08-11b_census_done_owner_decisions.md.

---

### 2026-08-13 evening → 2026-08-14 morning — the config-beats-default fix is written, tested and committed

The short version: **a step's config can now override a built-in default, which
for months it could not.** That was the whole of bug 231, and it is done in code
and waiting for the next chassis build to go live.

What was actually wrong. Every action declares a small spec — which settings it
accepts and what each falls back to when nobody says otherwise. The code filled in
those fallbacks *first*, then went looking for what the config had asked for — and
every one of those later lookups skipped a setting that already had a value. A
fallback counts as a value. So the fallback always won, and the config was
decoration. The visible damage was the logo one we already knew about: a step that
said "this is a logo" got told "it's a hero", and the file landed under the hero's
name.

The reassuring part is how tightly we could bound the risk. Because the old code
could only ever set one of these settings by one specific route, anything the new
code can touch is something that **did nothing at all before**. That's not an
estimate, it's a property of the code — so there is no way for this to disturb
something that currently works. Counting it across the live fleet: 99 config
entries were inert. 78 of them happened to repeat the fallback word for word, so
nothing changes for those. The remaining 21 belong to actions that were already
reading the config directly by a private back door, so they were already behaving
the way their config said. Net effect on live behaviour today: none. The value is
that the trap is closed for whoever writes the next one.

Two judgement calls worth knowing about. First, a config value containing a dot is
still ignored if it doesn't resolve — because a dotted value is a *pointer to data
elsewhere*, and we have already been badly bitten by treating one as a literal:
that mistake published over 150 broken image links named after the pointer itself.
Same reasoning here, so dots stay off-limits. Second, I deliberately did **not**
do one thing I originally planned. A bare word could in principle be read as
"go and fetch the thing with this name" rather than "use this word". I listed all
48 live cases before deciding, and every single one is plainly a word the author
typed — 'agentchassis', 'main', 'GB', 'high'. So bare words are taken literally,
the other reading is written up as an open question, and the fact that the same
bare word means different things in two situations is recorded as a trap rather
than left for someone to trip over.

A detail I'm glad I checked. The tool that reports this class of problem had to be
updated in the same commit, because two thirds of what it used to report as
"broken" is now working config — a tool describing last week's code is worse than
no tool. After updating it, it reported zero problems. I don't trust a zero from a
checker I just edited, so I took the live configuration, deliberately broke one
entry, and fed it back: the checker caught it and failed as it should. The zero is
a real zero.

Also worth flagging, because it's a coordination thing rather than a code thing:
the first job on my list had **already been done by another session while this was
being planned** — by a different method than the one you chose, and a slightly
stronger one (it makes the bad case fail loudly instead of quietly defaulting).
Nothing was lost, and it's the reason my change ends up altering no live behaviour
at all. I've recorded it rather than argued with it.

Two things I could not clean up and did not pretend to: another session has a file
in the shared tree that doesn't currently compile, and a second one is broken at
the branch tip. Neither is mine and the rule here is to leave other people's work
alone, so I tested against a clean copy of the committed code instead. And my
commit unavoidably carries one line of another session's register entry — I've said
so in the commit message rather than let a reader find it.

Still open: the council review is submitted and I have not read the verdict yet,
and the fix is inert until someone builds a chassis image. The other half of bug
231 — 96 config entries that point at data which sometimes isn't there, and fall
back silently when it isn't — is untouched and stays open by design, because
whether a pointer resolves is only knowable while the thing is running.

---

### 2026-08-14 afternoon — the review came back critical, it was right, and the fix is now live

The council reviewed the config-beats-default change and asked for revisions. Eleven
reviewers, six had nothing to say, and the one that held it up made a point I think
was the best criticism this lane has had: **my argument that the change was safe was
an argument, not a test.** I had proved, by reading the code carefully, that the new
behaviour could only touch settings that previously did nothing. The reviewer's
response was that this is true of the code as it stands today and nothing stops
someone adding a seventh or eighth rule next month that quietly breaks the proof. So
the proof is now a test that fails the moment anyone does that. It cost about twenty
minutes and it is strictly better than what I had.

There is a small embarrassment attached, and it's worth recording. The test says "the
automatic search did not override the default". That sentence is also true if the
search couldn't find anything in the first place — in which case the test guards
nothing and looks like it guards something. I checked, and the search *does* find
values there, so the test is real. But I only checked because we have a standing habit
of asking that question, and without it I'd have shipped a test that reassured
everybody and tested nothing.

Three other reviewer points were also real and are fixed: a step that used both the
old and new spelling of a setting got the *old* one, which is backwards; the change to
that old-spelling path needed tests for the case it wasn't about; and one reviewer
pointed out that when the new code *rejects* a badly-typed setting, it only says so in
the logs, which on this system disappear within minutes. So there's now a mode that
writes a permanent daily record instead. I have not switched it on — it needs a
container image built first, and turning it on before that exists creates a failure
this cluster reports as "still running" rather than "broken", which is worse than not
having it.

**Then the fresh build landed and the fix went live**, and I could prove the code is
running on both machines: the binary carries the exact commit, and — the important
half — a *different*, later commit is correctly *not* in it, so the check can tell the
difference rather than saying yes to everything.

But I could not prove the new behaviour actually happened, and I want to be plain
about that rather than let it read as a clean pass. I went to the logs to watch the
new rule fire and found nothing. The tempting conclusion is "no problems, all good".
The real explanation is that the machine only keeps about **ninety seconds** of logs,
on a machine that had been running five hours. I know that's the explanation rather
than a guess, because an *older* log line that has always been there is missing from
the same window too. So the logs can't answer the question either way, and the honest
status is: the code is definitely running, the behaviour is not yet witnessed. That
check is written down as still owed.

The irony is that this is exactly what the reviewer warned about — I hit their
objection myself, an hour later, trying to check my own work.

One more thing that isn't a code problem: this shared resolver has now been through
**27 reviews**, and **8 of them** flagged that it's a piece of shared machinery nobody
formally owns — and each time it shipped anyway, because each individual change was
well-argued and blocking it would have cost something real. I've written that up as a
proposal for you rather than trying to settle it myself, because the question is
whether this thing gets an owner, and that's a call for you.

---

**2026-08-14, late afternoon — the account ran out of credit mid-review, and it's back.**

The short version: the review round I'd resubmitted died halfway through, and it wasn't
anything wrong with the submission. Our Anthropic account hit its usage cap at 16:38 UTC
and every request the whole system made after that was refused. That's the third time in
fifteen days. You've since restored it, and the round is running again.

A word on how I checked it was actually back, because the obvious check lies. The natural
thing to look at is whether jobs are completing — and they were: sixty-three finished
successfully in the twenty minutes before I looked, none failed. That reading is worthless
here. Those particular jobs are the plumbing — deployment triggers, health checks, page
re-renders — and **none of them ask the AI anything**. So they run perfectly happily
through a total outage, and a healthy-looking job count means nothing about the thing that
was broken. The error log has the mirror-image problem: it falls silent both when calls are
working and when nothing is calling.

The only table that can answer the question is the one that logs each request to the AI
provider with a success flag on it. Every request from 16:05 to 16:42 failed — fifteen of
them, not one success. The next one, at 17:08:40, succeeded. That's the moment it came
back, and it's also what lets me say the dead round died *of this* rather than of something
I did: its death sits inside that window.

So I resubmitted it, unchanged — nothing in it needed fixing — and it's now working through
the reviewers. I'll report the verdict when it lands.

I've also written the trap down where the next person will hit it: there was already a note
about this outage telling people to check the right table, but it didn't say why the wrong
one is worse than merely unhelpful. It now does.

**What's still yours to decide, unchanged from this morning:** whether that shared piece of
machinery gets a formal owner (the 27-reviews question), and whether we switch on the daily
record — which needs a container image built first.

**Later — approved, and the reviewers caught something I'd have carried forward.**

The review came back **approved** at 17:29. Thirteen reviewers, five raised advisory
points, none serious enough to block. That's the change finished: it's approved, it's
running, and the record now says so.

Two of the five were about **my evidence rather than my code**, which is the part worth
telling you about.

One reviewer said the way I'd proved the code was running on the machines was a method
our own notes describe as broken. I checked, and it isn't — the note warns against a
different version of that check, and I'd used the sound one, with the control that makes
it mean something. So that objection doesn't stand.

The other one did stand, and it paid for itself several times over. I'd written "verified
on both machines". There are two machines with the obvious name — but **seventeen** run
that same program under other names, so "both" was a sample I'd described as if it were
everything. Going to count them properly turned up something I wasn't looking for: **the
system had been rebuilt again since this morning**, so the version number I'd certified in
two separate documents was already out of date. The fix survived the rebuild — I checked,
both directions — but I'd have carried a stale number into a third document if a reviewer
hadn't made me widen a check I was confident about.

A third reviewer asked a question I should have answered before submitting: of the
ninety-nine settings this change brings to life, do any of them touch page-building, where
a wrong value could quietly damage a site? **None do.** Twenty-one of the ninety-nine
actually change anything at all; the other seventy-eight set a value to what it already
was. And the three closest to page-building turn out to read their settings by a different
route that never involved the broken machinery in the first place — so the change doesn't
affect them either. That took one command. It should have been in the submission.

Two objections are deliberately left **open** rather than counted as settled, because both
reviewers asked for that: they're the "who owns this shared machinery" question, which is
the one waiting on you. Nothing else is outstanding on this piece of work.

**The decisions waiting on you, in one place (asked for, 2026-08-15).**

Six. Four come out of RFC_028, one is small and operational, one is bigger than this lane.

1. **Does the input resolver get an owner?** It's the one function every action goes through
   to decide where a setting's value comes from. What it guarantees is written in five
   separate places, all currently agreeing, with nothing keeping them so. 27 review rounds
   have touched it and 8 said it needs an architectural owner — then all 27 shipped. The
   deflection rate is the argument: a signal that fires eight times and changes nothing is
   teaching everyone to ignore it.
2. **Should the "a full stop means look it up elsewhere" rule be written once in code?**
   It's currently in comments in two places. That distinction cost 150+ live 404s to learn.
   Cheap to fix, and the one I'd say yes to without hesitation.
3. **Is there a budget on the number of rules in that chain?** You ruled exactly this way
   yesterday for a different accumulation (budget of 10, counter built). Nothing counts
   these; the chain is at eight.
4. **Does the alias precedence need a forward guard?** The safety argument is a census —
   true today, not a constraint. A config written tomorrow with both an old and new name
   silently gets the new behaviour, with one unit test behind it.
5. **Wire up the daily config check?** Built, has run, nothing schedules it. Needs an image
   built and pushed BEFORE the schedule is applied, or the failure reads as "still running".
   I can do it.
6. **The account cap — and a measurement that changes the options.** Third exhaustion in 15
   days. The caching fix from the 10th WORKED: the council gate now serves 97% of its input
   from cache, ~78% cheaper than it would be. The cap still blew, and the reason is that the
   three next-largest consumers — page-content-writer, content-gap-planner, diagnose-agent —
   have no caching at all, and the first of them charges more at full rate than the gate
   does. Options: raise the cap, add a second provider, or apply the technique that worked
   to the three that never got it. I'd do the third regardless, but I have NOT yet checked
   whether their prompts repeat enough to cache well, so I can promise no number until I do.

My own suggestion: 2, 4 and 5 are small and self-contained; then measure the three uncached
agents and bring you a number before anything is spent on 1, 3 or a second provider.

**2026-08-15 — the six decisions, and the one that turned out to be blocked by something that isn't true.**

You ruled on all six. Four of them (the resolver's owner, the "full stop means look it up"
rule written once in code, the limit on how many rules the chain may have, and the guard on
the old-name/new-name collision) are written, reviewed and approved. The reviewers earned
their fee: I'd tested two of the five ways that collision check can be wrong, and one of
them caught it. All five are covered now.

They also caught me repeating myself. I justified a design decision by quoting a count of
past reviews without attaching the query that produced it — which is the *same* objection
the *same* reviewer made about the *same* number one round earlier. I've written that into
the wrong-calls log, because the lesson had already been recorded in this lane's own file,
by me, and I sailed straight past it.

**On the scheduler: you were right, and better than I realised.** There are actually two
scheduling systems here. The six existing config checks all use the older one, which needs a
container image built and shipped for each check. The one you were thinking of already runs
four jobs of exactly the shape we need — which means no image at all, ever: the check lives
inside the program that already holds the information it's checking, and switching it on is a
single database row. That's the way I'll build it.

**On caching, the news is better than I expected and the reason is slightly embarrassing.**

Our system deliberately uses the five-minute cache rather than the one-hour one. The reason
was recorded carefully: a reviewer found that the longer cache needed a special permission
flag we don't send, and that using it without the flag would make *every* AI request in the
estate fail — not just this one improvement. Sound reasoning, and the note said plainly: do
not turn this on until someone confirms the flag.

I confirmed it. **There is no flag.** The longer cache is generally available and needs
nothing special. I sent one real request from inside the system and it not only succeeded,
the response explicitly confirmed it had stored an hour-long cache entry. I checked both of
the AI models we use.

That matters because of one number: the agent with the most to gain from caching reuses its
work **1% of the time within five minutes, and 99.8% within an hour**. Under the old setting,
switching caching on for it would have cost us about **24% more** than not caching at all. The
saving wasn't being left on the table — it was out of reach.

So the setting is flipped and with the reviewers now. Two honest caveats. It's a change to
the program rather than the database, so it does nothing until the next release is built and
rolled — that's your `make release`. And the largest spender of all, the page-content writer,
**cannot** be helped by this: the part of its instructions that repeats sits at the *end*,
and caching only works from the beginning, and even moved it would be too small to qualify.

**One thing I found by accident and think you should know.** Our internal register — the
document the automated reviewers treat as the source of truth — has been stating since the
10th that we use the one-hour cache. We didn't. The correction was actually recorded at the
time, but nine lines further down the same entry, so the document has been contradicting
itself for five days with the *wrong* half at the top where people read it. My change happens
to make it true again, which is exactly why I wrote it down rather than letting it quietly
become correct.

**What's left:** building the scheduled check, and putting the caching marker into the two
agents that can use it — one of which must wait until the release ships, or it costs more
rather than less.

---

**2026-08-15, later the same day — the caching marker is in, and it works. I also got one
thing wrong along the way and want to be straight about it.**

The job was the last piece of the caching work: switching on the saving for the agent that
plans what content a site is missing. Caching here means we pay once to store the long,
unchanging opening section of the instructions we send, and then pay about a tenth as much
each time we send it again. The catch is that stored copy only survives for a set time. Our
old setting kept it five minutes; the release earlier today changed that to an hour. Five
minutes was no good for this agent — it gets asked to do things roughly every ten minutes,
so the copy would always have expired and we'd have paid the storage fee for nothing, about
a quarter more than not bothering at all. At an hour, virtually every repeat lands inside the
window. I checked that myself rather than taking it from the handover: of 391 repeats over
three days, 1% came back within five minutes and 99.7% within the hour.

**Then I found something the handover hadn't accounted for, and it changed the plan.** All
our evidence that the one-hour setting actually works comes from one place — the review
council — and every seat on that council runs a particular model. This content agent was
running a *different, older* model, and while we'd confirmed that model accepts the new
setting without complaining, we'd never actually confirmed it honours it. So switching
caching on for this agent would have been betting the entire saving on something nobody had
checked. I put the choice to you and you said to move the agent onto the proven model first,
then switch caching on. That's what I did.

**Where I went wrong.** Moving to the newer model has a side effect I did spot: the new model
"thinks" before answering by default, where the old one didn't, and the budget we set for its
answer covers the thinking as well. If the thinking eats the budget, the answer gets cut off
halfway. So I raised the budget — and I put the new number in the wrong place. The settings
file has more than one spot where that number can live, and only one of them is actually read.
I'd written it to a spot nothing reads, so the real budget never moved. Worse, the safety check
I'd written to catch exactly this asked "is the number I just wrote big enough?" — which can
only ever answer yes. It congratulated me on a change that hadn't happened.

For about nine minutes the agent was running the new thinking-by-default model on the old
small budget, which is precisely the state I'd written a page of explanation about avoiding.
Nothing was harmed — no work came through in that window, which is luck rather than good
process — and I've fixed it, put the number where it's read, and rewritten the check so it
asks what the system would actually use rather than what I typed. I've logged it in our
running list of mistakes, because the shape of it is more useful than the incident: a check
that reads back your own change can't fail, and it looks exactly like diligence.

**It's working.** The first request through the new setup stored just under 5,000 units of the
repeated section, the answer came back well within budget with no truncation, and nothing
errored. I'm now waiting for the next request to confirm it *reads* that stored copy back
rather than storing it again — that's the bit that proves the saving is real, and a zero there
would be the failure, not the absence of one.

**One honest caveat on the money.** The newer model counts the same text as roughly a third
more units, so before caching this move costs more, not less. Caching more than covers it. But
it's also on introductory pricing until the end of this month, so any figure I quote today
flatters it — worth re-measuring in September rather than trusting today's number.

**2026-08-15, about ten minutes later — it reads back, and that also settles the bigger
question we'd been waiting on all day.**

The next request came through at 14:03 and used the stored copy: it paid the cheap rate for
just under 5,000 units instead of storing them again. So the saving is real, not just
theoretical.

The unexpected part is that the same observation closes the thing the morning's handover said
we'd have to wait hours for. The whole release today hinged on one claim — that stored copies
now survive an hour instead of five minutes — and nothing had actually proved it. The proof
had to be a request that reused a stored copy *more than five minutes after it was stored*,
because under the old setting that copy would simply be gone. The two requests were **8 minutes
and 46 seconds apart**, and the second one read the stored copy. Under the old five-minute
setting that is not a near miss, it is impossible. And I'd already checked the reverse case on
three days of earlier history: 29 times a request came back after more than five minutes, and
29 times out of 29 it had to store a fresh copy. So the check could have come out the other way,
and didn't.

**One thing worth knowing for next time.** We were watching the wrong agent. The review council
was the obvious place to look, but it's busy — requests arrive every minute or so, which means
gaps longer than five minutes almost never happen, so it can essentially never produce this
proof. The content agent I marked today gets asked to do things every ten minutes or so, which
is exactly the window that discriminates. It produced the answer within nine minutes of being
switched on. Going forward that's the one to watch.

Nothing needed reverting: no errors, no cut-off answers, and the readings are where they should
be.
