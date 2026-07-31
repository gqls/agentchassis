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
