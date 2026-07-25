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
