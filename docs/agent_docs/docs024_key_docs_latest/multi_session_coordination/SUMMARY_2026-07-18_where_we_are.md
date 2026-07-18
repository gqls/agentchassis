# Multi-session coordination — where we started, what we did, where we are, where we're going

**Written to be read aloud.** 2026-07-16 → 2026-07-18.

---

## Where we started

We had many Claude sessions working on the same thing at the same time: one repository,
one branch, one working directory, one image tag sequence, one live database, one
production cluster. Each session could see everything it did itself. No session could see
anything any other session was doing — not work in progress, not even work already
committed — unless it went and looked, and there was no agreed place to look.

That single condition produced four different-looking problems.

Sessions dispatched expensive work at targets another session was already fixing. Deploys
shipped every session's half-finished code, because the image was built from the shared
working directory rather than from anything committed. Sessions edited the same files with
no announcement, so collisions were discovered by accident. And commits swept up unrelated
threads' work, which made reviewing, reverting or bisecting any single change impossible.

It was not theoretical. The day before we started, a diagnosis run was destroyed
mid-flight: it correctly re-checked the live system before firing, found three broken
pages, and dispatched. Another session had begun repairing those exact pages five minutes
earlier. The evidence was fixed underneath the run while it was reading it, so it concluded
the bug did not exist. The lesson was sharper than "check before you act" — it had checked.
**Checking the live system tells you what is true now, not what is about to become true
because someone else already has work in flight. Checking the pod does not check the queue.**

---

## What we've done

**We stopped sessions firing work at each other's targets.** The diagnosis trigger now
checks, before it spends anything, whether any open work item already touches the thing you
are about to investigate — by page, by site, or by the source files in scope. If something
does, it refuses and shows you what and who. On its first live run it refused immediately,
surfacing two different sessions' open work on a single page. A real collision, caught
before it cost anything.

**We wrote down the working practice where every session will actually see it.** There was
no `CLAUDE.md` in the repository, so nothing told a new session how to behave here. There is
now, and it loads automatically at the start of every session. It carries the commit rules,
the deploy rules, and the hard-won traps. Other threads have since added their own sections
to it, which is exactly what we wanted.

**We inverted how images are built, and this is the most important change.** Previously
`make build` packaged up the entire shared working directory — so an image contained every
session's uncommitted, untested, mid-edit code, whether they knew it or not. The standing
workaround had been to simply not deploy: verified fixes sat unshipped, waiting for someone
else's release. Now a build takes its code from committed history instead. It cannot include
anyone's work in progress, yours or anyone else's. The safe build is the one you get by not
thinking about it, and the old behaviour is still available as a deliberate, clearly-named
escape hatch. If you forget to commit, the build tells you exactly what it is leaving out
and carries on — so the failure is a wasted build, rather than silently shipping someone
else's untested code to production.

**We proved it end to end in production.** We ran a full fleet release: all fourteen backend
services built from committed code and deployed. Every service came up healthy. We then
verified the running pod against the registry by digest — they matched exactly, and matched
a separate earlier build of the same commit, which shows the build is reproducible. The
fleet is live on version 1.0.1133.

**And along the way we found a real production defect.** Asking a simple question — how do
we do continuous deployment when restarting pods interrupts running work? — turned up
something worth more than the answer. The chassis tells Kafka a message has been handled
*before* it handles it. So when a pod stops, anything in flight is not delayed, it is
destroyed, because the message will never be redelivered. That is why restarts break running
work, and it is not really about deploys at all: crashes, evictions and node failures lose
work through exactly the same path. We can measure the damage: twenty-two orchestrations are
wedged right now, the oldest for about fifty-one days.

We filed that, and then we checked our own claim — and it was wrong. We had written that the
fix would be small because a de-duplication layer already existed to make redelivery safe.
It turns out that layer has the identical defect: it records a message as handled before
handling it too. So fixing the first problem on its own would have achieved nothing — the
redelivered message would simply be discarded as a duplicate, and the work lost through a
different door. We corrected the record. The correction is more valuable than the original
finding, and it generalises: **an acknowledgement layer and a de-duplication layer fail the
same way, because they encode the same mistaken belief — that receiving something is the
same as handling it.**

---

## Where we are now

Live and working: the pre-dispatch collision check; the shared working practice every
session loads; committed-code builds as the default across all fourteen services, proven in
production.

Known, documented, and not yet fixed: the chassis loses in-flight work whenever a pod
stops; the de-duplication layer has the same flaw; twenty-two orchestrations are stranded
and nothing sweeps them up. All of it is written down in the open bug queue with the exact
files, line numbers and fix shape, so whoever picks it up does not start from nothing.

And three honest limits worth saying out loud.

The commit discipline is **not self-protecting**. Committing your own work narrowly stops
you sweeping up other people's; it does nothing to stop someone else's broad commit sweeping
up yours. That happened three separate times during this work — including once to the very
change that fixes it, and once, fittingly, to a commit whose message disclosed that it was
carrying someone else's work as a passenger. By the time it ran, we were the passenger. The
only protection available is to commit early and narrowly; a long-lived pile of uncommitted
work is not a private workspace, it is shared mutable state.

We deployed into a busy cluster **by choice**, knowing it would drop about eight running
workflows, because the alternative was waiting indefinitely. That was the right call at the
time, but it is a cost we should not have to keep paying — and the delivery fix above is
what removes it.

And our verification has a boundary: we confirmed the two de-duplication call sites and the
ordering at each, but we have not proved those are the only ways a message can arrive. That
is stated in the bug file rather than assumed away.

---

## Where we're going

**First, fix the delivery guarantee.** Acknowledge messages only after the work is done, and
change the de-duplication layer to record completion rather than receipt — using a claim
with a lease, so that work abandoned mid-flight becomes eligible to run again instead of
being mistaken for work already finished. These two must ship together; either one alone
does nothing. The platform already has this pattern working elsewhere, in the work-item
queue, so we should reuse it rather than invent a second one.

**Then continuous deployment becomes straightforward** — and this is the important ordering.
Building continuous deployment first would make things actively worse: it would deploy more
often, and every deploy would destroy running work. Once a restart means a message is simply
redelivered, most of the ceremony we have been observing — quiet-checks, waiting for the
cluster to drain, choosing deploy windows — becomes an optimisation rather than a
requirement. The remaining pieces are small and well understood: let a pod finish what it is
holding before it dies, do not declare a new pod ready before it is genuinely working, and
run more than one of them.

**On the branching question, the short answer is: yes, but the branch is not the mechanism.**
Branches cannot isolate sessions that share one working directory — and ours all do. If one
session switched branches, every other session's files would change underneath them
mid-edit, which is worse than what we have now. The thing that actually isolates threads is
giving each one its own working directory, using git worktrees, at which point per-thread
branches follow naturally. That would eliminate the whole class of problem where sessions
overwrite and sweep up each other's work.

Two caveats stop it being an obvious win. It only fixes half the problem: the cluster, the
database and the image tags stay shared, so collisions over deployments, migrations and work
queues are entirely untouched by branching — and roughly half of everything that went wrong
in this work was on that side. And it moves cost rather than removing it: today's collisions
become tomorrow's merge conflicts, and at our rate — around a hundred and eighty commits in
two days across twenty-odd workstreams — merging becomes a real ongoing job. Most of those
commits are documentation, which gains little from isolation and would pay the merge cost
anyway.

So the recommendation is a middle path. Give separate working directories to the threads
that change platform code at the same time, where the collisions actually hurt. Leave
documentation-only threads on the shared tree. Keep exactly one branch as the thing that
gets deployed, which is effectively true already — noting that `main` is now two hundred and
seventy commits and five days behind, so it is a label we have stopped maintaining rather
than an integration point we are using. And merge often, because a long-lived branch trades
a collision problem for a drift problem.

**The cheaper alternative, which we should probably do first: a shared git hook.** This
turns out to be far less intrusive than we assumed when it was first declined. The
repository already runs a tracked, shared pre-commit hook for secret-scanning, configured
for every clone — so a guard would be an edit to something that already runs for everyone,
version-controlled, reviewable and revertable, rather than new machinery imposed invisibly.

The honest limit is that a hook and separate working directories fix **different halves** of
the problem, and it is worth being precise about which.

Of the three times work was swept up during this exercise, one was *cross-file*: a commit
about vetinary-medicine exports collected a change to the build file, which that thread had
never touched. That is the class that destroys reviewability — several threads' work
arriving under one thread's message — and it is exactly the original complaint. A hook stops
it. The other two were *same-file*: two threads had both edited the same document, and
whoever committed first took both sets of edits. Both of us used precise, explicit commit
commands and it happened anyway, because git commits whole files. **No hook can prevent
that.** Only separate working directories can, and then it surfaces as a merge conflict —
visible and resolvable — instead of a silent passenger.

One technical caveat shapes what the hook can actually be: git has no hook that fires when
you *stage* a file, so we cannot literally forbid the sweeping command. What we can enforce,
at the moment of commit, is the rule already written down — that a broad commit is allowed
but must announce itself. So: if a commit spans an unusually large number of files or
several unrelated areas, require its message to be labelled a sweep, and otherwise reject it
with an explanation. That is enforceable, matches the documented practice, and catches the
damaging case.

**So the recommendation is: do the hook now, and defer the working directories.** The hook
is perhaps ten lines onto infrastructure that already exists, changes nobody's workflow, and
is instantly revertable. Separate working directories are a real restructuring that trades
today's collisions for tomorrow's merge queue, and they should be earned by evidence — if
same-file collisions on the handful of genuinely hot shared files keep costing real time,
that is the signal to isolate the threads that change platform code, and only those.

**Still genuinely open for you to decide:** whether to put the hook in, and the branching
model. Neither is urgent; the delivery fix above matters more than either.
