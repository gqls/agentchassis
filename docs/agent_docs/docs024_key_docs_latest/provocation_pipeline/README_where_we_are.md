# Where we are — the daily provocation

Plain-prose running log, append-only, newest at the bottom. The owner maintains
this too; add below, never rewrite above.

---

**2026-07-31 — starting from "it never changes"**

You said the provocation hadn't changed that day and never had. That's right, and
the reason turns out to be simpler and slightly worse than a broken job: **nothing
was ever built that could change it.** There's no scheduler, no rotation, no pool
to choose from. The file that feeds the site is a static list written by hand, and
the current provocation is a fixed paragraph inside it. It has been edited six
times since the end of June, always by a person, and the last edit was 26 July.
So "it has never changed on its own" is exactly true.

The bit that makes this worth fixing rather than shrugging at: the site says
*"every day, one provocation"* on the about page and in the arena copy. That's a
promise about how the product behaves, and the product doesn't honour it. It's the
same family as claiming a number that isn't true — just about behaviour instead of
arithmetic.

**Three things I found that change the plan.**

First, a trap. The obvious cheap fix is to put a list of provocations in the file
and let the web page pick one based on today's date. That would break the Gauntlet
in a way that would be genuinely horrible to debug: the game engine on the server
reads that same file and looks for the single entry marked "today". So the page
would display one provocation while the engine argued a different one. Whatever we
build, the choosing has to happen when the file is made, not in the visitor's
browser.

Second, good news that changes the cost of everything. I assumed we'd have to
build the "publish this on a schedule" machinery. **We don't — it already exists
and it ran yesterday afternoon.** The news feed on our other sites works exactly
this way: a scheduled job wakes up, generates a file, and commits it to the site's
repository, which deploys it. I checked it properly rather than trusting the
config — three sites are serving news files stamped within minutes of the job's
last run. So the plumbing for a daily provocation is off the shelf. The only
expensive part is deciding *what* the provocation should be, which was always the
real problem.

Third, we've been here before. There's a plan from 25 June that designs this
whole thing, as a copy of the news pipeline. Half of it was actually built — the
file and the page code we're using today came from it. The half that chooses and
regenerates the content was never started. So this isn't a new idea, it's a
half-finished one, and I'd rather continue it than start again.

**One correction to the handoff I picked up.** It suggested the "See All
Provocations" page might be broken as well, and to check before designing
anything. I checked by actually rendering it in a browser, and **it's fine** —
all eight past provocations are listed properly. The handoff's measurement wasn't
wrong; the conclusion drawn from it was. It had measured that today's provocation
doesn't appear on that page, which is correct and expected, and the small amount
of text on the page got read as evidence of breakage. Worth saying because it
means the archive needs no repair work at all. The only real gap is that today's
provocation never joins the list — which your rule ("archive it when the new one
is published") now fixes.

I nearly made two mistakes of my own in the same half hour, both recorded in the
notes: I read text that was present-but-hidden as if it were visible, and I
matched a fifteen-character phrase that turned out to belong to a completely
different provocation from the one I thought. Both were caught by printing the
surrounding text instead of trusting a count.

**Your decisions this session.** The archive rule is settled — a provocation gets
archived when the next one is published, never during its own day. That's also
the safer answer, which I don't think was obvious: the archive page shows the full
argument for each entry, so archiving today's straight away would have quietly
undone the "sealed" Gauntlet page we built last week, via a third route nobody was
watching. Your rule closes that off without anyone having to remember to.

You've also set the direction: generate provocations with the LLM, behind a check
that they're safe, interesting, current or relevant to our audience, and actually
good provocations rather than just opinions.

**Where I'd push back slightly, on one point.** Two of those four criteria can't
be judged yet. You said it yourself in the same message — until we have live
contestants we don't know what they like. An automated judge of "interesting"
with no audience data will produce a confident score that nothing can check, and
confident-but-uncheckable is the shape most of our expensive mistakes have had. So
my recommendation is: automate the safety and form checks now, keep a person
approving the "is this any good" call, and let that person's job disappear
automatically once we have real engagement data. It's a few minutes a day and it
removes the risk of publishing something false rather than merely gating it. Happy
to be overruled if you'd rather take the speed.

**On the paired provocation — I think this is the best idea in the thread and
it's sequenced too late.** Four things.

It's the Gauntlet turned inside out, and it's the only version where the sealed
reveal genuinely matters. On the public site the seal is a nice touch. In a team,
it's the whole point: if one person can see a colleague's position before
committing their own, you get anchoring and agreement instead of honest
disagreement, which is precisely what the exercise is for.

"Until they've all committed" is the hard part, and it's the same problem as a
sealed-bid auction. You need a deadline, a rule for the person who never replies,
and — importantly — everyone's positions must become visible at the same instant.
If they appear one at a time, whoever answers last has read everyone else's
homework. Your instinct to give the organiser choices about timing is right;
those are the choices.

It's also the answer to your own blocker. You wrote that we'll understand what
people like once we have live contestants. This is how we *get* live contestants —
named people, arriving because someone they know invited them, coming back
repeatedly. That's exactly the data that would let the "is this interesting"
judgement be automated. So the dependency runs the other way from how it looks:
the paired mode unblocks the thing we can't build yet.

And it has a buyer, which the public version doesn't. A team lead has a budget and
a recurring need. An anonymous visitor has neither.

Two cautions. It needs people to have identities, and right now the system has
none at all — the public game doesn't know who anyone is. That's a bigger
prerequisite than the provocation work itself. And once we're holding named
colleagues' opinions and mailing them round their own team, privacy stops being a
preference and becomes something we have to get right: a leak there does real
damage to a real person. Related, and cheap to decide now but expensive later —
an AI verdict saying whose argument was weakest, sent to that person's team, is a
performance review nobody signed up for. Might be better to share the positions
with the group and keep each verdict private to its author.

**On the images idea, one note for when we get there.** Painting or sketching
instead of using a photograph solves the photographer's copyright, but it doesn't
solve the other half — a person's right to control commercial use of their own
likeness is a separate thing, and an original painting doesn't dodge it. In the UK
we're in a more permissive position than the US, but the risk is a viewer
inferring the celebrity endorses us. Also, "only feature celebrities we have free
images of" would mean picking the day's topic based on which pictures we happen to
own, which fights directly against the relevance test we're about to build. The
cheap safe route is to illustrate the idea rather than the person — none of our
nine existing provocations needs a face.

**What I need from you next**, in the plan as open questions: whether to accept
the human-approval recommendation above; what the audience actually is, because
"relevant to our audience" can't be checked until that sentence exists; and
whether you want the paired provocation prototyped quickly to feel the shape, or
specified properly first.

---

**2026-07-31 (later) — Grok, categories, and a paired-mode prototype you can click**

Three answers from you, and I've built the thing you asked for.

**Grok.** You were right on every point — we already query it and the key is
already wired. Better than that: the call we use goes to xAI's newer endpoint with
both a web search *and* an X search tool switched on, which is exactly the
Twitter-readership signal you were describing. So the "where do topics come from"
half of this is largely plumbing we already own. That's now twice in one day this
workstream has costed something as new work and found it already built.

One distinction I'd hold onto, because it's where this could go wrong. Grok tells
us **what is being argued about**, not what the provocation should be. Turning a
live topic into a good provocation — flat, two-sided, in our voice — is a separate
step, and your slop-and-danger filter has to run on **that output**, not on the
topic. The reason matters: X's most-discussed thing on any given day is quite
often a pile-on or a hoax. "It's trending" tells us about volume and nothing else,
so it can't be allowed to stand in for "it's worth arguing".

**Categories.** Agreed, and I'd go further — put the category field in now, while
there's only one. Two reasons. A category isn't really a label, it's a bundle of
settings: politics and pets can't share a safety threshold or an audience, so if
the filter is built assuming one global setting, adding the second category means
rewriting it rather than adding a row. And there's a collision worth knowing about
early: the game engine currently reads exactly one "today's provocation" per site.
Several categories running at once can't be expressed that way at all, so it needs
either a file per category or a change to that shape — and the Gauntlet has to
change at the same time, or the page and the engine will disagree about what's
being argued. That's the other team's code, so it's a conversation before it's a
task. I only spotted it because I'd written the constraint down this morning.

**The narrow first audience.** My suggestion: people who argue online for fun in
places that are already busy — the r/changemyview and Hacker News and tech-Twitter
axis. Three reasons. The share card the other team shipped this morning is
*native* to those places: "here's what I argued and here's how the AI ruled" is a
post format those venues already reward, whereas in a sports or politics feed the
same card reads as bait. It's also the easiest possible first job for the filter,
which matters because we're calibrating it — starting on political opinion would
mean tuning our hardest category with an untested filter, and we've just watched
what an untested filter costs on another site. And it quietly carries the paired
mode's buyers with it, because tech team leads are in that audience.

The honest weakness: that audience is small and sceptical, and it isn't where
"hugely popular" lives. Think of it as the calibration audience, not the
destination. The route to popular runs through your categories; the argument for
this one first is only that it's the place where being wrong is cheap.

**The paired prototype is built and you can play with it.**

```
cd docs/agent_docs/docs024_key_docs_latest/provocation_pipeline/prototype
go run .        # then open http://localhost:8099
```

Make a session, then open each person's link in a separate private window and play
all three parts. Nothing is saved — stop it and it's gone, which is deliberate.

Four decisions I had to take to make it work, all of them arguable and all worth
your view:

*Nobody can read a position before the reveal — including you, the organiser.*
That was the most tempting thing to allow and I think it would wreck the product:
a facilitator who's read the answers can't run the session straight, and people
who suspect you have will hedge. You get who has answered so you know who to
chase, and nothing else.

*Stay silent and you don't get to read the room.* If the deadline passes and Carol
never answered, Carol doesn't see anyone's positions. Without that rule the best
strategy under a deadline is to say nothing and wait — which is the exact
behaviour the whole sealed idea exists to stop.

*Once you commit you can't change it.* Genuinely arguable — a sealed auction lets
you revise, since nobody can see it anyway. I went the other way because "until
they've all committed" needs committing to be a state you can't back out of.

*Everyone opens at the same instant.* If it opened person by person, whoever
looked last would have read everyone else's answer before writing their own.

You get your three timing choices — wait for everyone, open at a quorum of N, or
open at a deadline — plus a force-reveal button for the person who's never going
to reply.

**Something I got wrong, and how it was caught.** I wrote fourteen tests designed
to break the seal, and they all passed. Then I deliberately broke the code three
different ways to check the tests would actually notice. Two were caught
immediately. The third wasn't: my test for "everyone opens at the same instant"
was checking all three people at the same moment on the clock, so it couldn't tell
a genuine simultaneous reveal from one that just stamped the current time whenever
you looked. It was green against a broken implementation. Fixed by checking the
three people at different times, and I re-broke the code to confirm it now fails.

**What I'd still like from you:** whether you accept the human-approval
recommendation from earlier today, which matters more now that the source is an
adversarial one; and your read on those four paired-mode decisions, particularly
the organiser not being able to see positions, since that's the one a customer
will ask you to change.

---

**2026-07-31 (evening) — the builder can now rotate, and what your "no approval" decision obliges**

You agreed the four paired-mode decisions and said we can try without human
approval for now. Both recorded as settled — I'm not going to keep raising the
approval question.

But it does change what has to be built, so I've written that down rather than
leaving it implied. With a person approving, a filter miss costs someone thirty
seconds of attention. With nobody there, a filter miss is a false statement on a
live page. So five things become obligations rather than nice-to-haves:

If nothing passes the filter on a given day, we **publish nothing** and yesterday's
provocation stays up. That's the same broken state we started with — but there's a
real difference between it happening silently for a month and deliberately for a
day with an alert on it. A stale provocation is a broken promise; a false one is a
broken product. The news pipeline already works this way, so we can copy it rather
than invent it.

If the filter *errors* — times out, returns nonsense — that must count as a
rejection, never as a pass. We have already shipped this exact bug once: a check
treated "don't know" as "no objection" and put an unpublished product range on a
live page as a confirmed match. Absence of a verdict is not a favourable verdict.

Somebody has to be able to see what the filter decided, including the rejections.
With an approver, a human sees every candidate as a side effect of approving it.
Take the approver away and nothing observes the filter at all unless we build the
observing — and then the first sign it's broken is a complaint.

A rollback command has to exist and be tested *before* the first automatic
publish, not written in a hurry when it's needed.

And calibrating the filter against our existing nine provocations stops being good
practice and becomes the only evidence it works.

**The builder now rotates.** I rewrote it as a schedule: each provocation carries
the date it goes live, today's is whichever one has most recently arrived, and the
archive is everything published before it. That implements your archive rule as a
property of the structure rather than a step someone has to remember — nothing can
be in both places, ever.

I checked it the strongest way available: run the new builder for today and
compare against what the site is actually serving. Identical, apart from a real
generation timestamp and the two fields (a date and a slug) whose absence is
exactly why the archive got stuck on 5 July.

That comparison caught me quietly rewording live copy. I'd derived the arena's
first card from the archive teaser, but the live card has a longer hand-written
line — so a "no behaviour change" refactor would have changed the words on the
page. Fixed properly, and the card is still derived from one place so it can't go
stale.

**Two things I got wrong, both worth the telling.** My first comparison checked the
card *titles*, found them all identical, and reported a match — the difference was
in a field I hadn't put on my list. And when I built the checker that proves
rotation works, I deliberately broke the builder six ways to see if the checker
would notice: it missed one. Freezing the date to a fixed value sailed straight
through, because I'd checked the date was *present* and never that it was
*correct*. That is precisely the trap I'd written up and filed this morning, about
this same file. Knowing a failure mode turns out not to be the same as checking
for it. Both fixed, and all six breakages are now caught.

**Nothing has been published.** All of this is local. The site is still serving the
26 July provocation, and with only nine provocations scheduled the new builder
produces exactly that same file today — Phase 0 makes rotation *possible*, it
doesn't make the site rotate. That needs new provocations (or the generator) and
the scheduled job. Publishing changes a live site, so I'll wait for you to say go.

---

**2026-07-31 (later) — published, and an honest account of what that did and didn't fix**

The new feed is live. It went out at 15:03 and the site was serving it 45 seconds
later. I checked the three pages afterwards by rendering them rather than reading
the HTML: the home page still shows the provocation, the archive still lists all
eight past ones, and the Gauntlet page is still sealed with no leak. Nothing broke.

Before writing anything I proved I had the right target. The database column that
is supposed to name a site's deploy repository is **empty** for vonc, so I
couldn't confirm it that way — and there's a known trap where publishing to a
plausible-but-wrong repo returns success and changes nothing anyone can see. So I
pulled the file out of the repo and compared it byte for byte with what the site
was serving. Identical, so that path is genuinely the one feeding the site.

**Something I got wrong, and it's the useful part of this update.** I wrote the
publish script so that rolling back is the same script run with an older file —
on the principle that an emergency revert nobody has ever run is not really a
revert. Then I dry-ran the rollback against the backup I'd just taken.

It refused. My own safety check required the new slug field, and the file I'd be
rolling *back to* doesn't have one — because that missing field is precisely the
thing this change fixes. **The escape hatch was blocked by the improvement it
exists to undo.** If that had shipped, I'd have discovered it at the exact moment
it mattered. Now split in two: the checks that prevent an outage run always, the
"is this up to our new standard" checks are skipped when reverting. I tested all
four combinations.

**Now the part I want to be plain about: the site still does not rotate, and it is
still making a false claim.**

What today bought is capability, not behaviour. The builder now picks by date, the
feed carries the fields the archive needs, publishing is one verified command, and
reverting works. But the schedule's last entry is 26 July and nothing rebuilds on a
cadence — so tomorrow the site will serve exactly what it serves today. "Every day,
one provocation" is as untrue this evening as it was this morning.

Two things are needed and both are required — neither alone does it. Provocations
to rotate *to*, which is either the Grok generator or a handful written by hand as
a bridge. And a scheduled job that rebuilds and republishes daily.

One trap I've filed for whoever does that job: the feed now carries a real
generation timestamp, and once something rebuilds daily that timestamp will move
every single day whether or not the provocation changed. So it will *look* like
rotation is working while the site repeats itself — the original bug wearing the
fix as a disguise. The only honest test is whether the provocation's identifier
changed, never the timestamp.

---

## Later that evening — the good news, and then a mistake of mine worth explaining

Two things fixed themselves while I was away, and one thing I'd written down turned
out to be wrong in a way that matters.

**The good news first.** The other team working on the Gauntlet finished their job.
They'd been sealing today's provocation so it can only be read once you step into
the round, and this afternoon the site was in a silly half-state — the lobby card
said "sealed" while the page above it still printed the whole provocation. That's
gone. I re-checked by loading the actual page: today's headline and body now appear
**nowhere** on the home page, the sealed card is there, and the "past provocation"
sample they added renders properly. I'd flagged that last one as unproven earlier,
because the thing I'd matched on could have come from an old archive card just as
easily as the new code. This time I found a label only the new code can produce. So
it's genuinely confirmed now rather than probably-fine.

That also means publishing is safe again, which it wasn't this afternoon.

**Now my mistake.** This morning I wrote the handoff document that tells the next
session what to do, and the first instruction was: write a few provocations by hand,
then add a scheduled job to rebuild and republish daily. I sat down this evening to
do exactly that, and it doesn't work. There is no job to add.

The scheduled jobs on this platform all point at an agent — a named piece of
software that does the work. There isn't one for provocations. There's no code
anywhere in the deployed system that knows how to build this file. The builder I
wrote today is a script sitting in the documentation folder; the cluster can't run
it, and nothing ever ships it. So the "scheduled job" I'd confidently written down
as the next step would have had nothing to point at. You'd get a row in a table that
does nothing.

The reason I got it wrong is worth saying, because it's the sort of error that reads
as perfectly sensible on the page. I'd correctly checked that the *news* feed on our
other sites rebuilds itself on a schedule and commits the result — that's real and
it works. I then let that stand as proof that *our* feed had a path to the same
machinery. It doesn't. It only proved a different feed does. Both halves of my
instruction were true on their own: we do need provocations written, and we do need
a scheduled job. Nothing joined them up, and I never checked the join. One database
query would have told me, and I ran it when I went to *do* the step rather than when
I *wrote* it.

**What it actually takes.** Real, but not alarming, and mostly copying something
that already works. The provocations need to live in the database instead of in a
script. Then a piece of code that picks today's, builds the file and commits it —
and there's a near-identical one already running for the vet medicine directory,
which does exactly this shape of job every two days. Then the agent definition and
the scheduled row, which are small. Writing the provocations comes **last**, because
until the rest exists they're just words in a file nobody can publish.

**One thing I want to flag, because it's a cost of a decision rather than a bug.**
There was a much cheaper way to do all of this: publish a fortnight of provocations
in one go and let the site pick today's by the date. No scheduled job at all, no new
code, and — the real attraction — rotation that *cannot quietly stop working*, which
a nightly job always can. I liked it a lot.

The seal rules it out. Publishing a fortnight ahead puts every one of those
provocations in a file anyone can read, and the whole point of the seal is that
today's isn't readable until you step in — tomorrow's certainly shouldn't be. So the
sealing decision is what forces us into the nightly job. That's not a complaint and
I don't think it changes the decision, but it's a real cost and it was invisible
from inside either piece of work. I've written it down so nobody proposes the cheap
version again in three weeks and has to rediscover why it doesn't fly.

The headline hasn't moved, though, and I don't want it buried under all that:
**the site still doesn't rotate, and the promise on the front page is still false.**

---

## Late evening — the machinery is built

Went and built the thing I said this afternoon was needed. It is done, it is
tested, and the site still doesn't rotate — for two reasons that are both worth
understanding, because neither is a bug.

**What now exists.** The nine provocations live in the database instead of in a
script. There's a piece of code in the platform that picks today's one by date,
builds the whole file, checks it, and commits it to the site's repository. There's
an agent that runs it and a schedule that will call it every six hours. All of it
is committed and the database parts are applied.

**Why the site still doesn't rotate.** Two things, in order.

First, code changes here don't take effect until a new image is built and rolled
out to the cluster, and I deliberately didn't do that. Rolling ships every other
team's committed work at the same time, which isn't one thread's call to make, and
it would have killed the code review I had running. Someone else's roll will carry
it along shortly — that's how this place works. Until then I've left the schedule
switched **off**, so it can't fire at a system that doesn't yet know the command.

Second, and this is the one that matters: **the newest provocation in the pool is
still 26 July.** Even with everything switched on, the machine would correctly
pick that same one every day, because it's the latest one there is. So the
remaining gap is no longer machinery. It's writing provocations.

**The check I'm most pleased with.** There were already rules for this feed, proven
in the Python script. Writing a second implementation in Go is exactly how two
versions of one thing quietly drift apart, each looking fine on its own. So I took
the nine real provocations, built the file both ways, and compared them —
identical. Then I wrote a second test that deliberately breaks one input and
insists the comparison notices, because a comparison that can't fail isn't a check
at all.

**One decision worth flagging, because it looks wrong and isn't.** The job will
*not* commit anything on a day when nothing has changed. It would have been easier
to republish daily regardless, but that would tick the file's "last generated"
timestamp forward every single day while the site repeated itself — which is
precisely the disguise the original bug was wearing. Now a stale timestamp on a
quiet day is the honest answer rather than a symptom to chase.

**Two mistakes.** I put backticks around a word in a commit message, which makes
the shell try to run it as a command; the message is permanently missing a word and
the rules here forbid going back to fix it. It's a trap I had written down for
myself and stepped in anyway. And for a few minutes I thought I'd broken the test
suite, when in fact another team had committed a test file whose other half is
still half-finished in the shared workspace — worth knowing that around here "the
tests are failing" isn't automatically about you.

---

## 2 August, evening — it works, and the site is still wrong

You rolled a chassis build, so the feed publisher finally has somewhere to run.
I checked it was genuinely in the running binary rather than trusting the version
number, turned the schedule on, and it worked first time.

The first run did nothing, on purpose. Before enabling it I ran the old Python
script for today's date and compared its output against what the site is actually
serving — identical. That told me a real run could only conclude "nothing has
changed" and skip, which made it the safest possible moment to switch the thing
on: the whole machine gets exercised, and the site cannot move. That is exactly
what happened, in about a second and a half.

Then I made it do the half that had never run. Skipping is easy; the part that
writes to the website repo had never once executed, and a writer nobody has
watched write is the classic way this platform bites us. Since today's content
was provably identical, I forced a commit — the only thing it could possibly
change was a timestamp. It committed cleanly.

**And that is where it got interesting.** The commit changed all 119 lines of the
file to publish one timestamp. Comparing what our new code wrote against what the
old script used to write, two differences showed up, neither of which breaks
anything: the Go code was turning the italics markup in your headlines into
escape codes, and it writes the sections of the file in alphabetical order rather
than the order a person would choose.

Both of those survive being read by a browser perfectly well, and that is exactly
why nothing caught them. My test compares the two versions *after* reading them
in, so as far as it was concerned the two were the same file. It had been passing
happily the whole time the escaped version was going live. The only thing that
showed it was looking at the actual published artefact.

I have fixed the escaping. I deliberately have not fixed the ordering: that costs
one messy commit when we switch writers, not one per day, and pinning it down
would mean maintaining a second copy of the file's structure that could drift
from the real one — which is the sort of thing that causes the bugs we spend
weeks on. It needs your next build to take effect; it is committed and waiting.

The mildly embarrassing part: the trap where an escape code gets silently turned
back into the character it represents caught me **four separate times** in one
session, including inside the paragraph where I was writing the warning about it.
It is already written down in my own notes as a known hazard. Knowing about a
trap and noticing you are standing in it are apparently different skills.

**Where that leaves us.** Every part of the machine now works and is proven
working. The site is still telling people it publishes a provocation every day
while showing them one from 26 July — because the pool has nothing newer than 26
July. That is now purely a content question. Adding provocations is a database
insert; no code, no build, no deploy. Nothing further happens on its own until
there is something new to publish, and I have not written any, because they go
out as your opinions under your name.

---

## 3 August, morning — the fix is live, and it proved a guess I made yesterday

Your new build carried the escaping fix, and it is now live on the site.

One thing was worth checking before I touched anything: whether the fix would
apply itself. It would not, and the reason is neat enough to be worth a sentence.
The publisher decides whether to bother committing by comparing what it would
write against what the site currently serves — and it does that comparison after
reading both in, which is precisely the step that makes escaped and unescaped
markup look identical. So the same blind spot that let the problem out also stops
the repair getting in. I confirmed the site was still serving the escaped version,
then forced a publish, and put the schedule back exactly as it was.

It worked. The live file now carries proper italics markup again and is back to
10,798 bytes — exactly the size the old Python script used to produce. I checked
it in the repository and then on the actual website, because those can differ.

**The part I am pleased about.** Yesterday I decided *not* to fix the second
problem — the sections of the file being written in alphabetical order — and my
reasoning was that it would cost one ugly commit when the writer changed, rather
than one every day. That was an argument, not evidence. Today's commit changed
**11 lines instead of 119**, which is the measurement that argument needed. The
reasoning held up.

I should be straight about what that does not settle: the old Python script still
writes the sections in a different order, so if we ever fall back to it by hand,
the next automatic run will rewrite the whole file again. That is a real cost of
keeping two ways of writing the same file, and I have written it down rather than
pretended it away.

**Where we are.** The machine is finished, live, and every part of it has now been
watched working. The site is still showing a provocation from 26 July under the
words "Today's Provocation", because that is genuinely the newest one we have. It
is not broken; it is accurately reporting that we have run out of material.

Everything from here is content, and content is your call — they publish as your
opinions under your name. Send me some text with dates and I will load them; the
site will start rotating on its own from that point with nothing further from
either of us.

---

**Wednesday 5 August — you authorised the generator, and it is built. Nothing is
switched on yet, deliberately.**

The site had been promising a daily provocation while serving one from 26 July —
ten days. I checked the machinery first and it was innocent: the publisher ran
that morning, took just over a second, and correctly did nothing, because there
was nothing to do. The pool held nine provocations, all of them yours, all dated
26 July or earlier, and nothing at all dated ahead. So the missing piece was never
the plumbing; it was that nothing has ever written a new provocation.

That is what now exists. Three separate pieces, and the separation is the safety.
One writes candidates, and it is incapable of publishing them — it can only create
drafts. One judges them. One decides which day each runs. The site only shows a
provocation that has been through all three, so no single piece breaking can put
something on the page.

The judge is the part that matters, because you decided months ago that these
publish without anyone reading them first. That decision is what makes the gate
the only thing standing between a model and your home page, so I built it to fail
in the safe direction by construction rather than by care: if the judge times out,
errors, returns nonsense, or gets cut off halfway through a sentence, the answer
is "no". There is exactly one place in the code that can say yes, and it is only
reachable after every check has passed. I tested seven different ways for the
judge to fail and all seven come back as a refusal.

The hardest part of the design was one distinction. A provocation is *supposed* to
be a contestable opinion, so a fact-checker pointed at it would reject every good
one you have ever published. But the supporting paragraph underneath often slips
in ordinary factual claims — your four-day-week entry says the pilots "measure
self-reported output", which is simply true or false. So the opinion is exempt and
the prose beneath it is not. Both mistakes are easy and both are bad: check
everything and you publish nothing, check nothing and you publish falsehoods.

**Three things I got wrong, and the way I found them is the useful part.** Rather
than trusting that my tests worked, I deliberately broke the code to see whether
they would notice. They did not, twice. One test was checking something that
happened to be true for an unrelated reason, so it would have passed even if I had
made the exact mistake it existed to prevent — and a comment I had written
confidently explaining why that could not happen was simply wrong. Another test
was quietly feeding itself different numbers from the ones it thought it was using.
Both are now real tests. I would rather report this than not: a test suite that has
only ever seen working code cannot tell you it would catch broken code.

The third thing was the rollback, and it is worth understanding because it is a
trap rather than a slip. There was already a command to restore the previous
version of the file the site reads. That is not a rollback here. Six hours later
the publisher rebuilds that file from the database, finds the bad provocation still
sitting there approved, and puts it straight back. So the rollback had to work on
the source, not the symptom: it retires the offending provocation, and the previous
one becomes today's automatically. It defaults to showing you what it *would* do
without doing it — and that preview is the real operation, run and then undone,
rather than a description that could drift from what actually happens. I tested it
against live data this morning and put everything back; nothing changed.

**What is not done, and it is the one thing standing between here and a working
daily site.** The tests I have run use a stand-in for the judging model, which
proves the plumbing and the refusal behaviour but says nothing about whether a real
model judges provocations well. Your own instruction was that the gate must be
calibrated against all nine of your real provocations, and against deliberately bad
ones, before it is connected to anything that publishes. That run costs real model
calls and I have not made it. Until it passes, nothing in the system refers to any
of this — it is built, committed, reviewed, and switched off.

So the next step is a calibration run against a real model. If it passes, the site
starts producing its own provocations daily. If it fails, we will have learned that
before anything reached the page, which is the entire point of doing it in this
order.

---

**2026-08-05, later — the categories question, and a collision I nearly caused.**

You asked me to build the gate and the generator. I started, and then found that
another session was already most of the way through both — it had written the gate
about two minutes before I looked. Its work was sitting in the shared working folder
but not yet saved into the project history, which is why nothing I checked first
showed it: the handoff note you pointed me at was written two days earlier and said
that half was unbuilt, and the usual "who is working on this?" tool only knows about
saved work. I found it by looking at the files themselves and then reading the other
session's own running log, where it had just written that the gate was built and it was
moving on to the generator.

That matters more than a bit of wasted effort. Both of us would have been adding a
piece of code with the same name to the same place, and the result would not have been
a tidy clash to sort out later — it would have stopped the project building at all, for
every other session working today. So I stopped and asked you what to do instead, and
you chose categories. The other session has since finished both halves and had them
reviewed and approved, so standing down was the right call.

**On categories, the short version: the thing standing in the way is smaller than it
looks, but it points the opposite way to what you'd expect.** The site can only serve
one provocation per day per site because the game engine goes and reads a single entry
called "today" out of the published file. To have several categories running at once,
that has to change. There were two obvious ways: publish one file per category, or keep
one file and put all the categories inside it.

The second one looks tidier and is the dangerous one. I read the engine's code all the
way through, and it barely checks the file at all — it confirms the "today" entry
exists and isn't blank, and then hands whatever it finds straight to the AI as part of
the question it asks. It never looks inside. So if we changed the structure, nothing
would break, nothing would error, and no alarm would go off. The AI would simply start
arguing against a lump of raw data instead of a provocation, and the only sign would be
that rounds felt slightly wrong. That is about the worst kind of fault to have, because
there is nothing to find. Publishing one file per category can't fail that way: if a
file is missing the engine gets a plain "not found", which it already knows how to
handle by showing the honest error page.

I also checked whether any automatic safety net would catch a mistake here, and there
isn't one. The two halves — the part that writes the file and the part that reads it —
are separate programs that share no common definition of what the file should look
like. They agree only by two comments written in English and kept in step by hand.

**What I have done, and what I have deliberately not done.** I have written the design
up as a formal proposal for your decision, because the code that has to change belongs
to a different workstream and the plan itself says not to design this without agreeing
it with them first. I have written the "reads it but doesn't check it" trap into the
shared warnings file so that anyone who opens that file in future is told before they
have a problem, and I have left a note in the other workstream's own starting document
so they hear it from us rather than discovering it. I have written no code at all.

**What I need from you** is a decision on the proposal — mainly, one file per category
or one file with everything in it, and whether a completed round should record which
category it was arguing about. That last one is cheap today and can't be recovered
later, because rounds are already being published at permanent public links, so a
category we never wrote down is gone for good.

**And the thing that hasn't moved:** the site is still showing a provocation from
26 July under a heading that says today's. That is now ten days. Categories don't fix
it; the new generator will, once it has been tested against a real model and shipped.

---

**Wednesday 5 August, evening — the calibration ran, and it failed. That is the
system working, and it needs two decisions from you.**

Your fresh chassis build carried the gate, so I could finally do the thing I said
was the last obstacle: judge your nine real provocations with a real model. I ran
it inside the cluster against a separate, isolated set of copies, so nothing on the
live site could be touched. I checked afterwards — your nine are exactly as they
were.

**The good half first. The deliberately bad samples were all caught** — the insult,
the invented statistic, the party-political rant, the empty AI-is-changing-
everything filler. Four out of four rejected. That is the direction that matters
most, because that is the direction where something false reaches your home page
with nobody in the loop.

**The bad half: it rejected five of your nine.** A gate that refuses the things you
actually published would quietly starve the site, so this cannot be wired up as it
stands.

**But the gate is not what is wrong, and this is the part worth your attention.** I
went and checked the rule against all nine of your provocations rather than the
handful the rule was written from. The plan I was building to states that every one
of your provocations "makes the case and then makes the counter-case". **That is not
true.** Five of them do. Four of them don't — they make the case and then pivot to a
different question instead of arguing back. One of them has no body text at all in
the database, not a single character.

I had turned that claim into a hard requirement. So the gate did exactly what it
was told, the model read your text correctly, and the result is a well-behaved gate
enforcing a rule that your own published work doesn't follow. There is no way to
tell that apart from a broken gate except by reading the rejected pieces, which is
why I made it record its reasons.

**So the two decisions are yours, and neither is a technical one.**

First: **do you want two-sidedness to be a requirement?** If you do, the gate is
right as built — but you should know it means four of your nine existing pieces
wouldn't clear the bar you're setting for new ones. That's a perfectly reasonable
position; the old ones simply stay as they are. If you don't want it required, I
demote it to something the gate notes but doesn't act on.

Second, a narrower one: your four-day-week piece was rejected because the model
said "the pilots measure self-reported output" overstates things — some pilots used
company figures too. It has a point. Awkwardly, that exact sentence is the example
the plan uses of a factual claim that *should* be allowed through. So: is that
acceptable rhetoric, or is it the kind of overstatement you want caught? Your answer
sets how strict that check is.

**What I have deliberately not done is loosen the rules until it passed.** That
would have given you a green result that meant nothing — a test tuned until it
agrees with itself. This project has a written rule against exactly that and it is
the right rule.

**And one thing I got wrong, which the calibration caught and nothing else did.** My
earlier tests said the gate accepted all nine. They were wrong, and in an
embarrassing way: eight of your nine provocations have no body text stored, so when
I built the test I *wrote the bodies myself* — including tidy "the counter is…"
turns. Then I tested whether the gate spots counter-cases. It did, in prose I had
written to contain them. I had even put a warning in that file about not
paraphrasing your work, and then paraphrased it. The real run scored four out of
nine on the same nine pieces. It is logged with the others.

Everything is written up in a handoff so this can be picked up cold. Nothing is
wired to publish, the site is still showing 26 July, and the next move is your two
answers.
