# Where we are — bug 066, agents running the wrong version of themselves

(Plain-prose log. Append at the bottom; don't rewrite what's above.)

---

**2026-07-27, afternoon.**

The bug, in plain terms: when the platform starts an agent that needs its own machine — the
ones that clone the repo and write code, like the feature implementer — it looks up which
version of the software to run in a database table, not in the thing that actually got
deployed. So we'd deploy a fix, watch it land, check the main service was running the new
code, and be satisfied. Then the platform would start one of those agents on a version four
releases old, and it would fail on the bug we'd just fixed. That happened on the 24th and
cost the feature-builder thread two rounds before anyone realised the agent wasn't running
the code they were testing.

The case file said the deploy simply never updates that table. **That turned out to be wrong,
and I'd already told you it as fact when you picked between the two routes — sorry.** The
deploy *does* update it, and has all along; there's a line buried a hundred lines into the
deploy recipe that does exactly that.

Which makes the real story more interesting, and honestly more useful. The update runs, and
the table went stale anyway. The reason is that updating the table is something *one
particular way of deploying* does. There are several other ways to push a new version — one
of them is written down as a shortcut in a comment right underneath the deploy step itself —
and none of those touch the table. And the worst case is rolling *back*: if we ship something
bad and undo it, the main service goes back to the old version but the table still points at
the bad one, so the agents keep running the code we just rejected. No amount of updating the
table at deploy time can fix that, because a rollback doesn't go through a deploy.

So the fix is somewhere else entirely: **an agent now runs whatever version the thing that
started it is running.** The chassis asks Kubernetes "what version am I?" and hands that
answer to every agent it starts. There is nothing to remember, nothing to keep in step, and
it works no matter how the new version got there — including a rollback, where the agents now
come back down with it.

Two things I decided to leave deliberately loose. If an agent is set up to run some *other*
piece of software entirely, we don't touch it — the chassis only corrects agents that are
meant to be running the same thing it is. And if someone genuinely wants one agent held on an
old version, there's an explicit way to say so, because the current workaround for this bug
depends on being able to do exactly that and I didn't want to remove it silently.

I also tidied the table-updating side, because even though it no longer decides anything, it's
what people read when they ask "are we exposed?" — and a record that lies is its own problem.
There turned out to be **five** copies of that update scattered around, every one of them
written to update *every row in the table with no conditions at all*. That was quietly
damaging: it also overwrote the frozen backup copies we keep for rolling a model change back,
which are the one thing whose version is supposed to stay put. There's one copy of it now, and
it leaves those alone. There's also a new script that answers the "are we exposed?" question
properly, by separating three things that used to get muddled: what's deployed, what the table
claims, and what the agents are actually running.

**Where this leaves us.** The code is committed and it's been through the review council.
But it does nothing until the next time someone builds and deploys the chassis — the fix is
inside the chassis, so the chassis has to be rebuilt to contain it. You asked me not to touch
production, so I haven't. That means **the bug is still open**, and deliberately so: it's
still reproducible out there right now. I've written the four checks into the case file so
whoever does the next deploy can run them and close it — including the important one, which is
to deliberately break a row and watch the agent come up on the *right* version anyway. A test
that can only pass tells you nothing.

One thing worth knowing while we wait: the old advice — "check the table before you fire one
of these agents" — is still correct until the new chassis is out, and becomes actively
misleading afterwards, because then a stale row is just untidy bookkeeping rather than a
broken agent. I've put that correction in the two places that give the old instruction, with
the one-line check that tells you which of the two worlds you're in.
