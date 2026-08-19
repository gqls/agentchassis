# Where we are — bug 284, the findings that were marked as failures

Plain prose, append-only, newest at the bottom.

---

**2026-08-16, afternoon.**

Picked up bug 284 off the open pile. It had been filed yesterday by the session
working bug 279, which finished and said in as many words that 284 was unowned and
needed a fresh pair of hands. Checked that no one else was on it — the ownership
script said "owned", but the only thing it had to go on was the commit that filed
the bug, which is a known blind spot, so I read the live session logs instead.
Nobody was on it.

**What the bug is.** The platform finds problems on our sites and writes each one
down as a "work item". Most of those are jobs: something is wrong, and we have an
agent that can fix it, so the item names that agent and the machinery sends it
along. But some findings are not jobs at all. Nobody can automatically repaint a
client's brand colours, restart a customer's virtual machine, or decide which page
a duplicated paragraph really belongs on. For those, the checker deliberately
leaves the "who fixes this" field empty and the item just sits there, visible, for
a person to read.

The trouble is that the step which moves new findings into the work queue never
looked at that field. It swept up everything on a site, including the ones with
nobody to send them to. The queue then picked them up, discovered there was no one
to hand them to, and stamped them **blocked — "cannot be routed to any agent"**.

So a perfectly correct observation ends up filed as a machinery failure. And it is
worse than untidy in two ways. Nothing ever unblocks those rows, because the
recovery job only rescues items whose named agent has since appeared, and these
never named one. And a blocked item still counts as "open" for de-duplication
purposes, so the checker that found the problem in the first place is not allowed
to report it again. The finding is frozen in the wrong state and the fresh evidence
is silently dropped.

**How big.** The bug file said 18 rows on 14 sites, all of one kind. When I asked
the question by the error message rather than by the kind of item, it came back
**60 rows across four kinds on at least fifteen sites** — the biggest group being
broken image references, at 40. Another 37 are sitting in the queue today waiting
to be swept up the same way. The reason the original count was low is worth
recording: the search that found the producers looked for a line of code that sets
the field to empty, and the worst offender does not set the field at all — the
programming language fills it in as empty by itself. Invisible to that search.

**What I have done.** The step that promotes findings now asks the same question
the dispatcher will ask a moment later: does this item name someone, and does that
someone exist? If not, it leaves the item alone. Both places now get that question
from one shared piece of code rather than each spelling it out, so they cannot
drift apart later — which is the failure this codebase keeps paying for. It also
now says out loud how many items it held back and of what kind, because a filter
that quietly does less looks exactly like a quiet week.

I also ran it past the diagnosis loop before asserting the cause, as the rules
require, and I am glad I did. It did not reach a verdict, but it found a mistake in
my evidence: I had said a certain marker on the rows could only have been written
by the one step I was accusing, and it turned out three different places write that
marker. The accusation still holds — but on a sharper test, because the other two
always write a fixed value and these rows carry the value only the accused can
produce. That check also turned up something I had missed entirely: two of the 60
rows were not created by the machinery at all. They were inserted by hand, by other
sessions, already in the queue and with no agent named. My fix does not stop that,
and I have said so rather than quietly rounding 58 up to 60.

**Where it stands.** The code is committed and has gone to the review council. It
does nothing until the next chassis release — Go changes only take effect when a
new image is built and rolled. I have deliberately **not** repaired the 60 damaged
rows yet, because until the fix is actually running they would simply be blocked
again within the hour.

**What is left, and one of it is a judgement call.** After the release: repair the
60 rows, then add a database-level rule that makes the bad combination impossible
to write at all — that is the only thing that catches the hand-inserted case,
because roughly twenty places in the code write these rows directly and bypass the
shared front door. That rule has to go in **after** the new code is running, not
before: database changes take effect instantly while code changes wait for a
release, so adding it first would make the old code fail on every site that has one
of these findings, and that would stop the improvement loop across the whole fleet.

One small note in passing: you asked for the plan to be prepared with Fable, and
Fable came back with "you've reached your Fable 5 limit" and produced nothing. I
did the planning myself rather than stall, and said so in the notes so nobody later
mistakes it for Fable's work.

---

**2026-08-16, later.**

The council came back **REVISE** first, then **APPROVED** on the second round.

The first round is the interesting one, because it caught me twice. One seat
pointed out that the marker I was using to blame the two checkers cannot actually
tell you which checker wrote a row — it records the row's own category, and several
things write it. It was right. Re-measured on the marker each checker stamps with
its own name, the answer came out cleaner than what I had submitted: nine rows from
one, nine from the other, exactly the two files I changed and nothing else. A second
seat picked at a loose phrase in my write-up — "three of six producers" — and
following it properly turned up a **sixth** checker I had not found, with
thirty-six live rows, which also never mentions the field at all.

The gating objection was fair too: I had called the change to the shared dispatch
code "semantically identical" without ever quoting the code it was replacing. It
is identical — I quoted both halves in round two, and they differ only by a
nickname and some whitespace — but the reviewer was right that I had asked them to
take it on trust.

Round two passed with four advisory notes, and I have answered all of them rather
than waving them through: the old case in the back-catalogue that looked like a
duplicate is a different member of the same family (a named agent that doesn't
exist, versus no agent named at all); the repair of the sixty rows is now a proper
gated script with its own backup, a check that can fail, and a rollback beside it;
and the deferred work is recorded in the database where the other agents can read
it, not just in our files.

**One thing genuinely needs you.** The two seats that looked hardest at this change
disagree with each other about the same edit. One says I unified too little — there
is a third copy of the same test elsewhere in the codebase, held in step only by a
comment, and I named it and walked past it. The other says I unified too much —
that the shared dispatch code every pipeline depends on shouldn't carry a
refactor that came out of a bug lane, even though it agrees the change is harmless.
Both are reasonable. The rules say a person breaks that tie, so I have left the
approved code as it is and written the disagreement down where you'll meet it.
Either direction is a one-file change whenever you decide.

---

## 2026-08-17 — it is done, and one of the rescued items is yours

The release went out, so everything this lane was holding could finally run.

**The 60 stuck items are repaired.** Each one is back in the state its own checker
files — the "roadmap" ones parked as roadmap entries, the broken-image ones back to
plain findings — and each carries a note saying what happened to it, so none of them
looks like it fixed itself. **One of them was yours**: the note you raised on
2026-08-12 reading fundamentallyai.com — the six Platform Log guides that name tools
and never link to them, and no Tools entry in the top nav. It had been filed as work,
then silently marked as a routing failure two days later, and it has been unreachable
ever since. It is now parked where the roadmap report picks it up. **Nothing has
acted on it — it needs a human, because there has never been an agent that handles
that kind of item.** Flagging it rather than leaving you to find it.

**The door is shut, not just swept.** Beyond the code fix, the database now refuses
outright to put a "nobody can do this" item into the queue of things to do — so a
hand-written insert, or one of the twenty-odd places in the code that bypass the
normal door, can no longer recreate this. I tested that by trying it: the bad shape
is rejected, the two legitimate shapes still go through.

**And I proved the fix works rather than assuming it.** A quiet week would have
looked identical to a working fix, so I picked the site with 36 of these items and
nothing else to do, and ran the exact step that used to break them. It held all 36
back and promoted none — where the old build would have taken every one. That is the
difference between "no complaints" and "checked".

**One thing needs your ruling** (unchanged from the review, nothing was done
unilaterally): two reviewers disagreed about the same edit — one says a third copy of
a shared check should have been unified too, the other says touching that file at all
went beyond the bug. Either way is a small, single-file change whenever you decide.

---

**2026-08-17.**

The build went out and the fix is live and working. It is worth saying how that was
established, because "we deployed it" is not the same claim: another session checked
that the *running* services carry the exact commit (the image's own label, plus a
proof that my commit is an ancestor of it), and I separately asked both running
processes whether my code is inside them, with a deliberate nonsense check alongside
to prove the question could come back "no". Then they proved it *works* rather than
merely exists — they picked a site holding thirty-six of these flag-only findings and
nothing else routable, ran the promoting step at it on purpose, and it held back all
thirty-six and promoted none. Under the old build those thirty-six would have been
queued and then marked as failures.

All sixty damaged rows are repaired, and the last hole is closed: there is now a
database-level rule making the bad combination impossible to write, added in the
correct order (after the build), and tested by trying to break it in three different
ways.

**Two things I got wrong today, both worth you knowing about.** First, another
session had already written the repair by the time I got to mine, because I wrote
mine into our lane's own folder rather than the shared migrations ledger where they
looked — so there are two files that do the same thing, and theirs is the one of
record. Second, and more embarrassing: I checked whether the fix had actually been
exercised, saw that the scheduled job which normally drives it is switched off, and
concluded the code "cannot currently run at all". That was wrong — you can fire the
step directly at one site, which is exactly what they did, and it's a technique
already written down in our own notes. I have corrected that claim where I made it
and logged all of today's misreadings; there were three, and every one of them was me
treating "I found nothing" as "there is nothing".

**One new problem found while checking this one, and it is live.** The same safety
check has two arms: one for a finding with nobody named to fix it, one for a finding
pointed at an agent that doesn't exist. I fixed the first. The second is happening
right now: our tool auditor is filing genuinely useful findings about live tools —
missing input labels, a tool depending on a script that isn't there — and addressing
them to something called "hitl-review", which has never existed. Fourteen of them,
across two sites, and it's growing: five yesterday, fourteen today. Each one is
recorded as a routing failure and, worse, silently blocks the auditor from reporting
that same finding again. I have filed it as bug 291 with the producer identified.
Note that neither my fix nor the new database rule catches it, because the handler is
named rather than blank — so it needs its own decision about what "hitl-review" was
ever meant to be.

**And the judgement call from yesterday is still yours**, untouched: whether to
unify the third copy of that shared test, or to back my change out of the shared
dispatch code. Two reviewers wanted opposite things and I have not picked for you.

---

**2026-08-17, decisions waiting on the owner.** Three, in order of how much they
matter. Bug 291 has been handed to another lane, so it is not one of them. Whether to
re-enable `improvement-sweep` is *not* one either — you already ruled on that in
`bugs_open/083` on 2026-08-15 (it stays paused; triage got its own scheduled task
instead), and this lane's fix does not change that ruling.

**1. The third copy of the shared safety test.** Three places in the code ask the same
question — "does this work item name an agent, and does that agent exist?". Before this
fix, each spelled it out separately and a comment asked the next person to keep them in
step. The fix made two of them read one shared piece of code. The third
(`remit.go`'s `HandlerStepConfig`) still has its own copy, and its comment says out
loud that it must match the dispatcher's. Two reviewers took opposite views: one that
the third should be brought in too, one that the shared dispatch code should not have
been touched at all from a bug lane. Options: (a) unify the third — one definition,
drift impossible, touches a fourth file; (b) revert the dispatch-code half — smallest
footprint, leaves three copies held together by comments; (c) leave exactly as is —
two unified, one not, which is where the reviewers disagreed. **Recommendation: (a),
and not urgently.** The third copy is correct today; the reason to fold it in is that
"keep these in step by hand" has already failed once here, which is the whole reason
this bug existed.

**2. `needs_experience_plan` on fundamentallyai.com — and it is not what it looked
like.** The other lane flagged this to you as a finding with no handler. Checked
directly: the handler **does** exist and is active (`experience-planner`); the row
simply never names it, and no row of this type ever has (5 live, 3 archived, every one
with the field blank). So this is a routing gap, not a missing capability. **But do not
just point the row at the agent:** `bugs_open/227` is still open and says that agent's
prompt hardcodes one site's diagnosis, so running it against fundamentallyai.com would
likely produce another site's plan. Options: (a) leave it deferred as a roadmap row you
read; (b) fix 227 first, then route it; (c) route it now and accept the risk.
**Recommendation: (b) if you want the work done, (a) if you do not** — (c) spends a
run to produce something misleading.

**3. Forty image-reference findings now have no way to close themselves.** The forty
`image_url_404` rows were restored to "recorded, waiting for a human". That check,
unlike its three siblings, has no self-clearing arm — so even after someone repairs a
reference the row stays open, and while it is open the check cannot re-report on that
same path. Options: (a) accept a permanent list a human prunes; (b) have a lane add the
self-clear arm the three sibling checks already have. **Recommendation: (b), low
priority** — it is a small, well-precedented piece of work, and without it the forty
rows will quietly stop being trustworthy as a list.

---

## 2026-08-17 (later) — your fundamentallyai.com note: I checked before filing, and it is already done

You asked me to re-file this as work the machinery can do, and to watch it. I went to
file it and checked the live site first. **All three things you asked for on 12 August
are now true.** I measured the served pages, not the database:

- **The Tools entry is in the top nav.** It reads Home / About / Tools / Contact /
  Platform Log / News / Capabilities.
- **Every guide links to the tool it describes** — all eight of them, between one and
  six links each, in the body text rather than the footer.
- **The tools are linked from the writing.** The homepage points at the tools page and
  at four individual tools; the Platform Log index carries six tool links.

None of that was this item — it was blocked for most of those five days by the bug we
just closed. Other rebuilds did it. **So I have not re-filed anything**: filing work
that is already done wastes a pipeline run and leaves a false record. The item is
closed as satisfied, with the measurements written into it so the next reader can see
what "satisfied" meant and when.

**One honest note about how I nearly got this wrong.** My first measurement said the
guides had no tool links at all — because I guessed their addresses and got six 404
pages, which look exactly like empty pages if you only count links. What caught it was
checking that the pages I was measuring had any content at all. The real addresses are
under /blog/ and /guides/.

**Two things I noticed while measuring, which are NOT your original ask and which I
have not acted on** — say if you want either chased:

1. **Three of the tools appear to have two guides each**, at two different address
   patterns (/blog/…-guide and /guides/tool-…-guide). That may be duplicate content
   competing with itself.
2. **The Platform Log index no longer links to any guide at all.** It links straight to
   the tools instead. That is the mirror image of your original complaint: the tools are
   reachable now, but the writing about them may have been orphaned from its own
   section index.

---

**2026-08-17, later — your three decisions, and one of them turned out to be the wrong
question. Also: the build you deployed shipped no new code.**

**First, the deploy, because it affects everything else.** The pods restarted at 14:42
today but the image tag is unchanged at `v1.0.1305`, and the running binary is the same
one as this morning: I probed it for a string that only exists in a commit made *after*
the build it carries, and it is absent, with a positive and a negative control alongside
to prove the probe works. **215 commits are sitting unshipped**, and another lane
measured the same thing independently. A rebuild at an unchanged tag serves the node's
cached image, so nothing new reaches the fleet. The fix is to bump `IMAGE_TAG` in the
makefile before building. Until that happens, anything I write in Go — including the
unification you just approved — is committed but inert.

**Decision 1 — unify the third copy of the shared test — was already done** by another
session acting on your ruling, and it is correct: there is now one definition, in the
package both callers can reach, with the dispatcher and the checker delegating to it. I
verified it builds and passes, and repaired one thing the merge got wrong: the new
function had been inserted into the middle of another function's comment block, so the
documentation for one was attached to the other. No behaviour change.

**Decision 2 — fix 227 first, then route the row — is recorded and handed over.** 227
belongs to another lane, so I have not touched their fix; I have written the dependency
into their bug file so they know an owner-raised row is queued behind it, and noted the
one thing that makes it easy when they land: the handler they need (`experience-planner`)
already exists and is active, so routing that row is a one-field change once their prompt
fix is in.

**Decision 3 — "add the images" — is the right instinct but the wrong action for most of
them, and I stopped rather than generate thirty pictures nobody needs.** I measured all
thirty findings against what each site actually holds:

- **Eleven already have the image, under exactly the name the page asks for.** Nothing to
  generate; the file simply was never published to the web path. That is a deploy that
  did not run, not a missing picture.
- **Six are heroes where the site has plenty** — lendzy, for instance, has nine, but they
  are named per page (`hero_home`, `hero_about`, `hero_price_cap`…) and one page is asking
  for a plain `hero`, which nothing is named. Generating a tenth would not fix the page;
  pointing the page at its own hero would.
- **Five genuinely need generating** — case-study illustrations on two sites.
- **Eight are the favicon and social-card gaps, and they belong to bug 131**, which is
  owned, so I left them alone.

**And there is a live blocker on the five that do need generating.** The imagery
pipeline is failing today — twelve failures against four successes, all with the deployer
getting a 404 trying to commit to a GitHub branch called `master`. I checked the obvious
explanation (those sites have no repository configured) and it is **not** the cause: the
sites whose images succeeded in the same window have no repository configured either. So
filing the five now would most likely just add five more failure rows. I have put the
whole census and that blocker into bug 114, which is the existing bug this belongs to.

**What I would do next, if you agree:** treat the eleven "already have it" and six
"pointed at the wrong name" as one piece of work in the imagery lane (it is their class,
and it is a repoint/republish, not new art), let 131 finish the favicon and social cards,
and hold the five case-study images until someone understands the deployer's 404. None of
that needs a decision from you today unless you want it prioritised.

---

**2026-08-17, evening — this build is real, and it carries what you approved.**

Unlike the one before it, this deploy genuinely shipped. The tag moved (1305 → **1307**)
and the image is a different one, and I checked it the cheap way this time: the image
itself records which commit it was built from, and I confirmed the copy I inspected is
byte-identical to the one the pods are running before trusting a word of it. It was built
from `a6d1c53c`, and that build contains the routability fix, **the unification you
approved**, and my documentation repair. The backlog of unshipped work is down from 215
commits to 42.

So **decision 1 is done and live**: there is now exactly one definition, fleet-wide, of
"does this work item name an agent that exists". I confirmed it in the running binary
rather than by reading git — the old duplicated form is gone from it and the shared one
is there, and no source file still contains the old spelling.

**A correction I owe you about yesterday.** I told you the previous build had shipped no
new code, and cited a probe of the running binary as evidence. The conclusion was right,
but that probe was worthless: the text I searched for was inside a code comment, and
comments are stripped out when the code is compiled, so it could never have been found in
any build. What actually proved the point was the simpler thing sitting next to it — the
version tag was unchanged and the image was byte-identical to what was already running.
I have logged it, because presenting a test that cannot fail alongside one that can, as
if they were one piece of evidence, is exactly the habit these logs exist to catch.

**On the images, I have not fired anything yet, and here is why.** The imagery pipeline
is mid-retry: another lane's ten failed items were reset and one is being worked right
now. Since the new build went out there have been no failures — but also no images
produced, so that is not good news yet, it is simply no news. Filing my five into a path
whose health is unknown would most likely just produce five more failures. The moment
that in-flight item either produces an image or fails, I will know, and I will either
file the five or tell you the blocker stands.

Two of the three groups are not mine to act on in any case: the eleven "already have the
image, never published" and six "pointed at the wrong name" belong with the imagery lane
(I have given them the full census), and the eight favicon/social-card gaps belong to bug
131.

**Later the same evening — the images: I found out why five of them don't exist, and it
isn't something I should fix by hand.**

The blocker I mentioned has cleared: after this build went out, the imagery pipeline's
retry produced a finished image end to end, and the only failure since is an ordinary
timeout rather than the repeated 404. So the machinery works again.

But when I went to add the five that genuinely need creating, I found they were never
going to be created. The framework's image generator only looks at two kinds of page —
blog posts and tools. The case-study pages are a third kind, and they are simply not in
the list it sweeps. So no amount of running it would ever produce those images, while the
pages themselves carry a reference expecting one. That is the real defect behind those
five findings, and it sits upstream of the report that flagged them.

**I have not fixed it, on purpose, and I'd rather you agreed than assumed.** Two reasons.
Adding the missing page type is a two-line change, but it is fleet-wide: every page of
that kind on every site enters the generator's queue at once, which is real image-
generation spend, so it should be that lane's call with the usual review, and somebody
should count those pages first. And hand-writing the five images myself would mean
composing the prompts by hand — writing site content, which is the framework's job, not
mine — and it would hide the gap so nobody widens the list.

So the position on your third decision is: eleven need publishing, six need repointing,
eight belong to bug 131, and the last five need a small decision from the imagery lane
before anything can generate them. All of it is written up in bug 114 with the file and
line numbers.

---

**2026-08-18, later — your answer on the two-guides-per-tool question, and what it closes.**

You said: two or three articles per tool is fine, it's not a strict rule. So the content-
strategy question I left open is settled — the pairs stay. Nothing to do, and importantly
nothing to undo: I had already declined to file any de-duplication work, because the pairs
turned out not to be duplicates at all (a usage guide and a decision guide per tool), so
there is no cleanup sitting half-done waiting on this. I have recorded your ruling in bug
309, which is where the question was written down, so the next session reading that file
does not re-raise it.

What this does NOT settle is the actual defect on that page, and that is what I am on now:
the Platform Log index lists six articles as plain text with no links in them at all, so
none of the writing is reachable from the site. That is bug 309 and it is still open.

**A correction to my own handoff, straight away.** That handoff told you a diagnosis run
was in flight and that the next session should read its verdict first. There is no such
run. I checked the three places it would have to exist — the intake queue, the diagnosis
artefacts, and the orchestration table — and all three are empty for that identifier, and
no work item of any kind was created in that whole hour. So the trigger never landed. I am
re-firing it, and this time I will confirm the intake row exists rather than trusting the
identifier it printed at me.

**Same evening — the diagnosis came back, and it confirms the whole chain.**

The run I re-fired finished in about eight minutes and the verdict is CONFIRMED. It
worked the mechanism out for itself and cited the two facts everything hangs on: the
site has no "blog" settings entry of the kind the component is asking for — zero rows,
and it printed the site's actual list to show it — and the six link fields are missing
from the page's stored data as a result. That matches what I had measured by hand, from
a different direction, which is the point of running it.

So, in plain terms: the index lists six articles and links none of them, because the
component builds each link from a setting that has never existed on any site we run.
When the setting is missing the framework quietly drops the field, and the template only
draws a link if the field is there — so the link doesn't come out broken, it doesn't come
out at all. Nothing looks wrong on the page, which is why five pieces of writing sat
unreachable without anyone noticing. The second half is that the one routine that would
have filled in real links only looks at pages it recognises as a blog, and this page is
filed as a section index, so it never touches it.

The good news is that this is a migration somebody already did everywhere else. Every
sister component — the tools list, the guides list, the games list — was moved years-of-
commits ago onto a source that reads the real pages, so the title and the link always
come from the same row and can't disagree. This one was left behind. I checked what that
source would return for your site and it gives 8 articles, correctly leaving out the one
that's archived — which also disposes of the odd card advertising a retired page.

**I have not changed anything yet, and I'd like your call on which way to fix it**, because
the obvious fix touches a component another site uses too, so it should go through review
rather than be done quietly. The options are in the bug file, ranked. I've flagged the
choice to you separately.

**Later still — I made the fix you chose, and a safety check stopped it. That is the right outcome, and here is why.**

The change itself is done and live in the configuration: the index component now reads
your real article pages instead of a settings entry that never existed. I proved it
works before and after shipping it — I ran the new template through the rendering
engine on its own first, and then the real pipeline resolved **eight actual articles
with working links**, correctly leaving out the retired one. So the odd card pointing
at an archived page is fixed as a side effect, exactly as I said it would be.

**But the page has not changed, because the save was refused, and nothing was written.**
The framework has a guard that blocks any edit which quietly guts a page's text. The
new cards came out at 42% of the old text and the floor is 50%, so it stopped.

I could have switched that guard off. I didn't, for two reasons. The first is that the
switch isn't per-page — it would have turned the protection off for every page
re-render across every site, to push one page through. That is a bad trade in any
direction. The second is better: **the guard was right, and it found something I had
missed.** Five of your eight articles have no summary text stored at all. The old cards
hid that by having the model make something up; the new ones can only show what is
really there, so those five would have shipped with an empty description.

So the blocker is not the fix — it is a genuine gap in the content, and it now has a
number on it. Fill in those five summaries and the same re-render lands at about 73%,
which clears the floor comfortably. I have not written them myself, because writing
site content is the framework's job, not mine — and the framework already knows about
this: it has flagged 606 missing-page-essentials items across the estate, every single
one sitting undelivered in the same stuck state this whole workstream has been about.
That is the thread to pull, and it is why this hasn't fixed itself.

Everything is recorded, including a warning for whoever migrates the next component
like this, because the error message politely invites you to disable the very guard
that is protecting you.

---

**2026-08-19 — closing this one off.**

The new build is out and carries everything; the fleet is now within three commits of the
branch, which is the healthiest it has been all week. I re-ran the checks on it and got
the cleanest reading this work has produced: **not a single work item anywhere in the
estate is sitting in the "blocked" state**, for any reason at all — while 722 findings of
exactly the kind that used to end up there sit correctly parked, waiting for a human. Four
builds in a row now, no regression.

So yes — **this lane can close, and I have closed it.** The bug is in the closed pile, the
fix is live and has been watched working, the damaged rows are repaired, the hole that
allowed it is shut at the database level, and the decisions you made have all either
landed or been handed to the lane that owns them.

I have written into the handoff the four specific things that would justify anyone opening
it again — a blocked row reappearing, the database rule losing its guarantee, a fourth copy
of the shared check turning up, or the improvement sweep being switched back on (that last
one is not a fault; it is simply the first time the guard would run on its own schedule
rather than being fired by hand, so it is worth one look). Anything short of those is
noise, and I would rather the next person spent their time elsewhere.

What is left over from this work lives with the people who own it: the image census and the
page-type gap with bug 114, the favicon and social cards with 131, and the one row of yours
waiting on 227. None of it needs me.

**2026-08-19 — the new build is live, I checked what it changed, and the answer is nothing on this lane. Which is correct.**

I verified the build properly rather than trusting the version number: the running pod's
image fingerprint matches the one built locally, so this is genuinely new code and not
the cached-image trap that has bitten this estate before. All of yesterday's work is in
it, both mine and the other team's.

The index page is still unchanged, and that was always going to be the case — the part
I fixed is configuration, which went live the moment it was applied, and the part that
is stuck is missing text on five of your articles. A new build of the software changes
neither. I would rather say that plainly than let a fresh deploy imply progress.

**The lane is finished, and the last loose end tied itself off overnight.** There was one
thing I could not prove yesterday: we added a database rule that blocks a bad kind of
record, and I could argue it was safe but had not yet seen anything real pass through
it. Overnight forty-two items went through correctly, every one properly routed. That is
the proof, and it came from ordinary traffic rather than a test I built to reassure
myself.

I also gave myself a scare this morning and want to record it, because it is the same
mistake this estate keeps paying for. I counted the bad-category records and got six,
where yesterday's note says zero — for a moment it read like the fix had come undone.
It had not: all six are from before the fix, and I had dropped the date filter that made
the original number mean anything. The number was real; the question I asked it was
wrong.

**So: one thing left, and it is not mine to do.** Five articles need their summary text
written, by the framework rather than by hand. Then the same command I already ran
finishes the job. I have measured that it would clear the safety guard with room to
spare, so this is not a hope, it is arithmetic.

The reason it has not happened on its own is the thing this whole lane was about. The
system already spotted those missing summaries — six hundred and six times across all
your sites — and every one of those notices is filed under a category with nothing
attached to act on it. The fault we just spent this lane fixing is the same fault
standing between us and the last step, one level up.
