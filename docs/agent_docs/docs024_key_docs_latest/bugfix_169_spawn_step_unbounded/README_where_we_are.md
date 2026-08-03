# Where we are — bugs_open/169 part A

Plain prose, append-only, newest at the bottom.

## 2026-08-02 — why a build sometimes stopped for 40 minutes, and what has changed

You asked me to pick up 169. Part B of it — the one about some sites never getting
their turn — was already fixed by another session a day earlier. Part A was untouched:
a build job that got stuck for over half an hour doing nothing, holding a piece of work
so nothing else could pick it up.

**The cause is simpler than it looked, and it is general.** When the platform runs one
of its own built-in steps, it does so with no time limit at all. None. If that step
makes a network call — to the database, to the message bus, to Kubernetes — and that
call never comes back, the whole job sits there for ever. Nothing complains, because
from the outside it looks like a step that is still working.

There is a setting called `timeout_seconds` that looks exactly like it should prevent
this. It does not. It is read, stored, and handed to the step — and then no step in the
entire codebase actually looks at it. I checked all 271 of them.

**What I have changed.** Built-in steps now run under a time limit. The default is ten
minutes, chosen from measurement rather than taste: over the last three days the
platform ran 6,951 of these spawn steps, 99% of them finished within 24 seconds, and
exactly one exceeded five minutes — the broken one, which ran for four hours. So ten
minutes catches the broken case with an enormous margin and touches nothing healthy.

There are three ways to escape it, because this changes behaviour for every step in the
platform and I would rather you had levers than certainty: any individual step can set
its own limit, any individual step can turn it off entirely, and there is a single
switch that turns the whole thing off across the fleet without needing a rebuild.

**One thing I want to flag.** A time limit only works on calls that agree to be
interrupted. Most do — database, message bus, Kubernetes all respect it. A step that
deliberately sleeps does not. So this makes the common case survivable; it is not a
guarantee that nothing can ever stick.

**Something I found on the way that is worth knowing.** Jobs that got stuck were
eventually clearing themselves after almost exactly four hours. That is a safety net
that already existed — but it only notices when a new message happens to arrive, it
marks the job failed rather than rescuing it, and it never actually stops the stuck
work. It explains why this was survivable rather than catastrophic, and it is not a fix.

**Where it stands.** The change is committed and reviewed but **not yet live** — it is
program code, so it does nothing until someone builds and deploys a new image. The
ticket stays open until then, with the exact verification steps written into it. I did
not close it, because "committed" and "working in production" are different claims and
this project's own rule is that only the second one closes a ticket.

**Two mistakes of mine, both recorded.** I committed the change one commit before its
register entry, when the rule says they must be together — a small window, but the rule
exists because "I'll do it next commit" is how the entry ends up never written. And I
wrote a commit message containing backticks, which the shell executed; it failed loudly
and cost nothing, but the safe habit is to put long messages in a file instead.

## 2026-08-03 — it is live, and I made it fail on purpose before believing it

The new build went out, so I checked the running containers rather than the version
number: the change is in both of them.

Then I did the part that actually matters. A safety net that has only ever been observed
not-catching-anything is not a proven safety net, so I made one fail deliberately. I
picked the smallest, most reversible thing on the platform — a health check that runs
every minute and simply runs again if it misses one — took a snapshot, gave it an
impossible one-millisecond limit, let it tick once, and put it straight back.

It behaved exactly as designed: the step was cut off, the job failed cleanly instead of
hanging, and the error said which step it was and how to change the limit. Most
usefully, the cancellation reached *inside* the step and stopped its database call —
which was the whole bet behind how I built it. The health check was back to normal on
its next run a minute later.

**And that exercise found a flaw the entire test suite could not have.** The error
message came out saying it timed out on step `""` — a blank. In production the platform
doesn't fill in the step's name on that path, and every test I had written filled it in
by hand, so the tests were blind to it by construction. Fixed, with the name now a
required piece of information the code cannot leave out. That refinement rides along on
the next build; the protection itself is already live.

The ticket is closed. One thing is still open and it is a question for you rather than a
loose end: whether the platform *should* put a time limit on every built-in step by
default. I think yes and that is what is running, but the reviewers were right that a
decision affecting every step deserves your sign-off rather than mine, so it is written
up as a proposal with three options.
