# README — where we are (bugfix 003, spawn loses child response)

Append-only. Newest at the bottom. Plain prose.

---

**2026-07-20.** Picked up bug 003 — the one where a parent agent sends work to
a child and the answer never comes back, so the parent just sits there until a
cleanup job gives up on it half an hour to an hour and a half later.

The file already had two causes: a network problem reaching one of the Kafka
brokers, and the messaging code acknowledging messages before actually doing
the work (so a restart throws away whatever was in flight). Both still true.
We found a third, and it's the one that explains the worst cases: the code
that is supposed to notice "no reply, try again" is a little timer that lives
inside the pod that sent the request. Deploy a new version — which happens
most days — and every pending timer dies with the old pod. Nothing rebuilds
them. So the retry system only works when nothing disturbs it.

Two things turned out better than feared, one worse. Better: the retry logic
itself is well written, and there's already a once-a-minute background job in
every pod we can extend to do the retrying from the database, which restarts
can't kill. Worse: the health checks on the worker pods are fake — they answer
"I'm fine" unconditionally, so Kubernetes never restarts a pod that's actually
stuck. The probes were pointing at endpoints that always say yes. We're fixing
that now.

Also shipped the first real fix today: the cleanup sweep had a blind spot — a
whole category of stuck orchestration it never looked at. 24 of them had piled
up, one 53 days old. The sweep now covers that category (we checked seven
weeks of history to pick a threshold that would never have hit anything
healthy) and it cleared all 24 within seconds of going live.

One irony worth recording: we sent our new "third cause" theory to the
automated diagnosis system to check us — and that very check got stuck in
exactly the way the bug describes. It'll now be cleaned up by the sweep we
just shipped. We're proceeding on the strength of having read every line of
the relevant code, but the independent verdict is still outstanding.

The network flakiness itself (the thing that triggers all of this) has changed
shape since July 15 — it's now a low-level background flicker hitting all
three brokers from most machines, not one bad route. That's an infrastructure
investigation in its own right, so it's now its own bug (040) and this
workstream concentrates on making the platform shrug off flakes instead of
losing work to them.

---

**2026-07-20, evening.** The health-check fix is live in production — but it
got there by accident rather than by plan, and that's worth writing down.

The change was finished and building cleanly this afternoon, waiting on the
advisory review council before being committed. While it waited, another
session committed its own work with a broad sweep, and my unfinished files
went along for the ride. The new chassis image (v1.0.1140) was then built and
deployed with my change inside it. Nothing was lost and nothing broke — but a
change that was deliberately being held for review shipped without one. This
is the exact hazard the repo's commit rules warn about; I'd been careful not
to sweep up anyone else's work, which does nothing to stop mine being swept up
by someone else.

The good news is that it works, and I checked it properly — against the
running pods rather than against the code. Both the main chassis pod and a
spawned worker pod now answer the health question honestly: they report how
many seconds ago a Kafka broker was last reachable, instead of the old
unconditional "OK". Nothing across the fleet's 44 pods has restarted as a
result. One caveat I want to be straight about: the restart-when-genuinely-
stuck path hasn't actually happened yet, because nothing has been stuck since
the deploy. So the fix is verified in the sense that it reports truthfully; the
part where it rescues a wedged pod is still untested in the wild.

The council did eventually return its third verdict: revise, not approve. Most
of the earlier objections were satisfied — the reviewer who had objected most
strongly to how I'd structured the code now approves it. Of the remainder,
two were about my write-up not showing enough detail rather than about the
code, one I checked and can dismiss with evidence, and one is a genuine hit
against me: I claimed the new code plugs into an existing shared health
mechanism, and it turns out that mechanism has no users at all. The placement
is still right, but the justification I gave was flattering and unchecked. I've
logged that in the wrong-calls file, because the interesting part isn't the
mistake — it's that I carefully verified every claim that might have weakened
my case and skipped the one that strengthened it.

Two things I'd like a decision on rather than deciding myself: whether to run
a deliberate test of the restart path (block one pod's access to Kafka and
watch it recover), and what to do about a 93 MB compiled binary that the sweep
commit accidentally added to the repository — it belongs to another session's
commit, so I've left it alone.

---

**2026-07-20, later that evening.** We ran the restart test, and it failed —
which is the best thing that happened all day.

First, the binary: it's out of the repository, and it turned out to matter more
than tidiness. Because builds work by exporting a snapshot of the repository
into a clean folder, that 93 MB file was being copied into the build folder of
every single service, on every build. It's now removed and blocked from coming
back. (One honest limit: it still exists in the project's history, and removing
that would mean rewriting history, which this repo forbids for good reasons.)
Worth noting the removal took two goes — my first attempt looked like it
worked and silently didn't, because of the same git subtlety the repo's rules
rely on elsewhere.

Then the test. Rather than disturb a working pod, I started a throwaway one
pointed at a dead-end network address, so it couldn't reach Kafka at all. It
was thoroughly stuck — every attempt to talk to Kafka timed out, it couldn't
create any of the channels it needed. And my new health check cheerfully
reported that everything was fine, for six minutes, and never restarted.

So the fix I described this afternoon as "verified live" was broken. Both
things I checked were true — the new code was running, the endpoint returned
real data — but I had only ever exercised the case where the answer is "yes".
I never once asked what it says when the answer should be "no", which is the
only case the thing exists for. Two separate mistakes were hiding behind that:
the check was reading its list of Kafka servers from a different place than the
rest of the program does, so it was watching the wrong Kafka; and it was
testing reachability by simply opening a connection, which in this network
succeeds even against addresses where nothing is listening.

Both are now fixed: it reads the same setting the rest of the program uses, and
it now has a genuine conversation with Kafka rather than just opening a socket.
That's committed but won't take effect until the next image build — and the
restart behaviour is still officially unproven until we re-run this test
afterwards, which is now a test that can actually fail.

Two things for you. There's a policy question I don't want to decide alone: the
check currently reports healthy if it can reach *any* Kafka server, so a pod
that has lost just one of the three — the original symptom in this bug — would
still look fine. Making it stricter would catch that, but risks restarting pods
unnecessarily during routine maintenance. And a caution for anyone repeating
the test: my throwaway pod ended up listening to the main shared channel and
replayed eleven days of old messages. It did no damage (I checked — it wrote
nothing to the database), but the runbook now says how to avoid it.

---

2026-07-25. The two big fixes are built and committed. In plain terms: until
now, if an agent's pod died at the wrong moment, the message it was working on
was gone for good — the queue had already been told it was handled. And when a
child agent never answered, the reminder to chase it lived only inside one
process's memory, so a restart meant nobody ever chased it. Both are now fixed
in code: the queue is only told "handled" after the work actually finishes
(with a claim-and-lease system so a retried message isn't accidentally done
twice), and the chase-up list now lives in the database, where any pod can pick
it up — a request that times out gets retried within about a minute, up to
three times, instead of sitting until the 90-minute sweeper kills the job.

The database side is already live and verified; the code side does nothing
until we ship new images for the chassis and the git adapter, which is the next
step. The review council is looking at the change now — we committed without
waiting, as agreed, and will add the review stamp only if they approve. Two
small extras rode along: two slow agents now get a proper idle timeout instead
of none, and pods now get a full minute to finish up when asked to shut down
instead of being killed mid-drain.

One thing worth knowing: while wiring this we found three subtle traps that
would have made the retries useless if we had shipped the original design as
written — the retry counter wasn't being read on resent requests, the
bookkeeping table refused updates instead of taking them over, and a
chased-and-answered request was being mistaken for a duplicate. All three are
fixed and written up in the bug file.

---

2026-07-25, later. The review council has now looked at the finished work twice
and rejected it twice — and both times the veto came from the same single
reviewer, the guardian, while nearly every other seat approved. It's worth
being precise about what the guardian is and is not saying. It is not saying
the fix is wrong: in round two it accepted our technical case point by point —
the blast radius is now named and mechanically checked, and it agreed the old
message-delivery code "is the defect, not battle-tested plumbing". Its
objection is about process and size: a change that rewrites how every agent
receives and acknowledges messages, across seven files, is in its view an
architecture change, and architecture changes should go through a dedicated
review with a staged rollout plan — a review track which, as it happens, we
don't have. Its fallback suggestion was to ship only the one-line header fix
now and put everything else through that (not-yet-existing) process. It also
noted, fairly, that the database migration was already applied and the code
already committed before the verdict — that followed our own standing rules
(commit early, apply schema before images), but I understand why it reads as
presenting a done deal.

Every concrete evidence gap any reviewer raised has now been closed: we proved
by compiler and by search that the deleted function has no remaining callers
anywhere, tests included, and that nothing else in the codebase was consuming
the expired-request queue. Two small tidy-ups were conceded (a config tweak
shouldn't have ridden the migration; a dead function should be deleted rather
than ported — noted as a follow-on).

So the decision now sitting with you: the code is committed, the schema is
live and safe either way, and the images are built-ready but NOT rolled.
Option one: roll deliberately — git adapter first as a small canary, then the
chassis, with the fault-injection tests run immediately and image rollback as
the escape hatch (the schema was designed to tolerate old binaries, so
rolling back is just re-deploying the previous image). No review stamp on the
commit, since the council did not approve. Option two: follow the guardian —
revert everything except the header fix on the shared branch and stand up the
architecture-review process it wants before shipping the rest; that is real
churn, and the bug keeps costing us roughly thirty-four stranded jobs a day
while it happens. Option three, separable from either: create the
architecture-review track the guardian is asking for, because it is at least
right that we keep pushing platform-wide changes through a gate designed for
point fixes. My own view: the technical case is as verified as it can be
short of running it, the daily damage is real, and I'd take option one with
the canary sequencing — but a fleet-wide roll against two guardian vetoes is
your call, not mine, which is why nothing has shipped.

---

2026-07-25, end of day. A plot twist, then a good ending. You chose the
reviewer's cautious path — undo everything except the small header fix and put
the big redesign through a proper review first. But between that decision and
my acting on it, the facts changed under us: another work session did a
routine fleet deployment, and because our code was already committed to the
shared branch, their build carried it straight to production. That is exactly
the behaviour our own documentation warns about, and this time it worked in
our favour — by the time I checked, the new code had been running the entire
fleet for four and a half hours, and running well. Nineteen timed-out
requests had been retried automatically; seven jobs completed that would
certainly have been silent losses under the old code; six hit the retry limit
and failed loudly instead of sitting in limbo for ninety minutes; nothing
stuck, nothing leaked, no crashes. Faced with that, rolling back working code
to satisfy a process objection made no sense, and when I put the changed
situation back to you, you agreed: keep it.

The process half of your decision stands and is done: there is now an
architecture-review track — a short written procedure for changes of this
size, with a template covering blast radius, staged rollout and rollback —
and its first entry is this very redesign, written up honestly as "reviewed
after the fact". The reviewer that vetoed us twice was asking for exactly
this to exist, and it was right about that even though events overtook its
caution. Still to do: a deliberate kill-the-server-mid-job test to prove the
recovery machinery under controlled conditions (the fleet was too quiet to
run it this afternoon — I have a watcher waiting for the next busy moment), a
re-run of the old health-check restart test, and a look at the weekly numbers
around the 1st of August to confirm the ninety-minute strandings are gone.

---

2026-07-25, evening. The controlled kill test ran, and it earned its keep
twice over. I deleted the main worker server on purpose while three jobs were
mid-flight. The new rescue machinery did exactly what it was built to do —
within half a minute of the first job's deadline passing, a completely
different service's housekeeping loop picked up the orphaned request and set
the work going again. That was the one code path no real traffic had
exercised yet, and it works, across services, with no shared memory at all.

But the test also flushed out an older problem that had been hiding behind
the one we just fixed. Each rescued job kept being re-done every three
minutes, forever: the work itself succeeded every time (including a real
code-repository commit per attempt), but the "job finished" message was being
thrown away by an old guard that only accepts messages for the server that
originally owned the job — and that server was the one I'd killed. Dead
servers never answer, so the guard would have rejected those messages until
the end of time, and the safety sweep that catches stuck jobs can't see this
shape because the loop looks busy. Before our fix this same flaw just meant
a job quietly died within the hour; now it means a tireless loop — which is
louder and, frankly, how we found it. I stopped the loop by clearing the
dead server's name off both jobs (they then finished normally within a
minute — no work lost), filed it properly as bug 075 with the fix mapped
out, and left the fix itself for a fresh thread: it's the same piece of work
that's been blocking us from running more than one worker server, so it was
always coming.

Score for the day: the big delivery rewrite is live, ratified and proven
under deliberate failure; one long-hidden defect found, contained and filed;
and the platform now has the architecture-review process the council was
asking for. Remaining on this workstream: the health-check restart re-test,
and the weekly numbers around the 1st of August.

---

**2026-07-26 evening — bug 075 is fixed in code, and the reason it existed is
worse (and simpler) than we thought.**

I went to fix the guard that threw away those messages and found out something
about it that changes the story. The guard compares the name of the server
holding a job against its own name. I had assumed that name was kept up to
date — that whichever server picked the job up would stamp itself on it. It
isn't. There is a line of code that looks like it does exactly that, and it has
never worked: the database write that follows it simply doesn't include that
column. So the name on every job is the name of the server that first created
it, however long ago, and nothing has ever changed it. There is still a job in
the system stamped with a server that died on the 13th of July.

That makes the guard indefensible rather than just unlucky. It cannot tell a
live server from one that has been gone for a fortnight, because it isn't
looking at anything live. And a reply only ever arrives at one server, so
throwing it away because the name doesn't match doesn't hand it to somebody
else — it destroys it.

So the fix is: take the job over, write down in the log that you did and whose
it was, and get on with it. The protection people thought that guard was
providing is already provided properly elsewhere — two servers cannot pick up
the same reply, because claiming a reply is a single atomic database operation,
and the job's own record is version-checked on every write. Before removing the
guard I checked the thing that would have made this dangerous: whether any
service runs more than one copy of itself and owns jobs. None does.

The second half is the runaway loop itself. The retry counter was being *set*
to one rather than *increased*, on a counter that was always starting from zero
— so it read "attempt 1" for ever and the limit of three could never be
reached. That is now a real count that survives a server dying, which is the
whole point, since a different server may pick up each attempt. I wrote the
tests for it with a deliberate trap: one test encodes the OLD broken rule and
insists it never stops, so if I have fooled myself about what the fix does, the
test suite says so.

None of this is live yet. It is committed, and it does nothing until the next
chassis image is built and rolled — the standing rule is that a bug stays open
until the fix is actually running and has been proven by breaking it on
purpose. I have written that proof out as a script so whoever rolls can run it:
it refuses to run against an image that doesn't have the fix, it checks that the
old code is genuinely gone rather than just that the new code is present, and it
re-runs the deliberate server-kill test that found this bug in the first place.

Two smaller things I chose not to do: a sweep for jobs owned by dead servers
(the database has no way of knowing which servers are alive, so that needs
something built first), and an extra safety net in the cleanup sweep (the bug
report itself says to do that only after the counter is real, which it now is —
so it is a follow-up with a trigger written down, not a loose end).

---

## 26 July, evening — closed it, but only after finding a fourth fault inside our own fix

I sat down to close this case and could not, which turned out to be the useful
part of the evening.

The headline numbers were good and I want to put them down plainly, because they
are the answer to "did any of this work". Jobs that died waiting for an answer
they never got: **2.34 an hour** in the day and a half before the fix went live,
**0.38 an hour** in the day and a half after, and **none at all** on the 26th
across 1,114 completed jobs. Thirty jobs rescued themselves that would previously
have vanished quietly. I had to get those from a history table rather than the
obvious one, because the obvious one only keeps about thirteen days and the
before-picture had already been deleted.

Then I looked at how the retry system was actually behaving rather than whether
it existed, and found the fourth fault. When an answer comes back, the code
"claims" it first so that two servers can't both act on it. Only afterwards does
it do things that can fail. If one of those failed, **the claim was never handed
back**, and — this is the bit that matters — nothing anywhere in the system knew
how to release one. The record sat in a state no cleanup job had ever been told
about. Every recovery path we have looks at other states. So the waiting job
waited for ever, and our own retry system, built specifically to end that, could
not see it.

There were 181 of these sitting in the database, the oldest from 26 June. Two had
a live job still waiting on them. One of those was a web page that would never
have been published.

The fix is one clause in the cleanup routine that already runs every minute, so
it was live in minutes with no software release. The count went 181 to 8, and
the 8 are answers being processed right now. The stranded page-publish job was
picked up by a completely different service and finally got a straight answer.

**Two things went wrong today that I want on the record more than the fix.**

The first: I verified the sister service by checking that some old code had
disappeared from it, got the answer I wanted, and it was meaningless — I was
looking at a path with no program in it at all. What saved me was that I had also
checked for three things that *had* to be there, and all three came back missing
too. Three impossible answers is what exposed it. On its own, "the old code is
gone" is a check that cannot fail, and this is the third time we have logged a
version of that mistake.

The second: our own tool for "is anyone else working on this?" told me nobody
was, so I started writing a fix. Another session had already written it — better
than mine — but hadn't committed yet, and the tool only sees committed work. What
stopped me overwriting them was the editor refusing to write to a file that had
changed under me. A blunter command would have destroyed their work silently.

That second one changed the shape of the fix for the better, oddly. Being locked
out of the file they were holding forced my fix into the database routine instead
of the program code, and that is the more robust answer: it doesn't have to know
about every way the code can bail out, including ones nobody has written yet —
which is exactly how this hole got opened in the first place.

I also re-ran the health-check test that failed so embarrassingly on the 20th,
when a deliberately broken server reported itself perfectly healthy for six
minutes. This time it reported itself broken and **the platform restarted it**,
about two minutes in. That is the behaviour we have been claiming for six days
without ever having proved. It also let me delete a piece of planned work: we
were going to teach servers to kill themselves when stuck, purely as insurance
against the health check being wrong. It isn't wrong. I'd rather have one
mechanism that is proven than two where the second exists because we didn't trust
the first.

So: closed and moved. What's left belongs to other cases and is written down in
them, not left floating here. **One date to keep — 1 August**, a week after the
main fix, to check the numbers hold and that I didn't just catch a quiet weekend.
Nothing needs a decision from you.
