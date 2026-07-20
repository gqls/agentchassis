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
