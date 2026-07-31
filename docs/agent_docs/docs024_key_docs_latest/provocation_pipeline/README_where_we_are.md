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
