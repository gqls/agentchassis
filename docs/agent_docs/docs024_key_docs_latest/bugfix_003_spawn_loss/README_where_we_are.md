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
