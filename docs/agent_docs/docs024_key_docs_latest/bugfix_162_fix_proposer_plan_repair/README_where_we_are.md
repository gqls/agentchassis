# Where we are — bug 162, the fix-proposer plan repair loop

## 2026-07-31, late evening

Picked up bug 162. It's a small one with a tidy history, and it turned into one
genuinely useful lesson that has nothing to do with the bug.

**The bug.** When our fix-proposer agent writes a plan for fixing something, that plan
goes through a structural checker first — does every edit name a file, is the same file
edited twice by mistake, that sort of thing. Until today, if the checker said no, the
whole plan was thrown away. Not rejected on its merits by a reviewer: thrown away before
any reviewer saw it, over a rule the agent was never told about. And thrown away
*quietly* — the column a dashboard would look at stayed empty, so the run read as clean.

Another thread already built the cure for this a couple of days ago and switched it on
for one agent (feature-designer). They deliberately left it switched off for
fix-proposer, because fix-proposer belongs to a different lane and changing another
lane's live agent from outside is how sessions tread on each other here. They wrote the
switch-on as a ready-to-run file and filed bug 162 so it wouldn't be forgotten. That was
the right call and it worked — the file was sitting there with instructions.

**What I did.** Checked the bug was still real (it was), checked nobody else was working
it (nobody was — I checked the commit history, then the other sessions' live transcripts
by bug number, then again by the names of the actual code involved, which is the check
that catches someone mid-edit). Confirmed the compiled half of the fix was actually
running on both servers. Then ran the file. Now, when fix-proposer's plan fails the
structural check, it gets handed back to the agent with the exact problems and one
chance to fix them, instead of being binned.

**The check I nearly didn't do, and should have done first.** The switch-on adds a
little router: "plan OK? go to the reviewers. Not OK? go to repair." I'd read the code
that evaluates that question and satisfied myself it worked. But the risk I'd reasoned
about was the wrong one. If the router couldn't *find* the answer at all — because some
results get stored in a wrapper and some don't — then every *good* plan would be sent
for repair, forever, in a loop that nothing stops. Reading the code couldn't tell me
which storage shape we had. Looking at real runs could, and did: it's the safe shape.
Worth remembering that "I read the code and it's fine" and "I looked at the data and
it's fine" are not the same sentence.

**One thing I chose not to do.** The proper way to prove this works end to end is to
deliberately break a plan and watch the repair happen. The method is written down. I
didn't run it, because doing so means temporarily changing a setting on a shared agent
that three other threads had jobs running through at that moment — their jobs would all
have been caught in it. Proving my wiring at their expense isn't a trade I get to make
for them. I've written it up as a gap rather than pretending it's proven; the same
mechanism *is* proven on the other agent, which ran it three times today.

**The part worth telling someone else about.** While checking all this I found that the
shared code had five ways to give up on a plan, and only one of them left any record an
operator could find. Meanwhile a comment in that same file assured the reader that if
you look in the log table and find nothing, nothing was refused. That's not true, and
it's the kind of untrue that makes your next query wrong rather than making it fail.

Fixing it meant changing behaviour that four tests explicitly protect, written days ago
by the other lane. So I put it to the review council rather than deciding alone. Approved
— nine reviewers, three advisory objections, none serious. One of them asked me to prove,
with a test, my claim that this couldn't affect the agents that had opted *out*.

So I wrote that test. It passed. Then I broke the code on purpose to check the test would
notice — and **it didn't**. Nor did the three older tests that say, in so many words,
"this must not touch the database". All four were decorative. The assertion everyone had
used reads like "nothing happened", but it actually means "everything I asked for
happened", and if you ask for nothing, it's always true.

That's the useful bit. It cost nothing to find, because breaking your own code to see if
the test complains takes about a minute. It would have cost a great deal to not find:
I'd have shipped exactly the reassurance the reviewer asked for, with a test beside it
certifying something it never checked. All four are now real, and I've verified they fail
when they should.

**Where it stands.** The bug itself is fixed and live — that half was configuration,
which takes effect immediately. The extra hardening is committed but won't be running
until someone next builds the chassis image; I've said so plainly rather than calling it
done. Ticket closed and moved to bugs_closed.
