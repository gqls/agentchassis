# Where we are — dispatch throughput and the road to thousands of domains

Owner's plain-prose log. Append-only, newest at the bottom.

---

## 2026-08-19 — the workstream is claimed, and the question got bigger

You asked for a deep look at every option for increasing the system's throughput — including
whether we need more than one repo, more than one deploy path, or another cluster — because
the plan is to host and maintain several thousand domains, and when you promote the service
you may get many sign-ups per day in bursts.

What we found, in plain terms:

**The machines are not the problem.** The cluster is running at a fraction of its capacity —
the computers are mostly idle. What limits us is that the work is deliberately taken one
piece at a time in several places: the dispatcher serves one site at a time (about 83 items
an hour for the whole fleet, measured), the agents that talk to outside services mostly
handle one request at a time, and every finished page goes out through a two-slot deploy
queue. These were all sensible safety choices when they were made; they are now the ceiling.

**Money is the other ceiling.** Every piece of work costs AI calls. The account has hit its
monthly spending cap twice this month already, at a fraction of the volume you're aiming
for. Scaled naively, thousands of domains would cost more per domain per month than the
£10 domain fee brings in. So "how fast" and "how cheap per item" have to be solved together —
fewer, smarter items per domain, cheaper models for routine work, and Anthropic's batch
pricing for anything that can wait a few hours.

**Bursts change the priorities.** If a promotion brings fifty sign-ups in a day, each one is
a full site build, and today's queue serves everyone strictly in the order work arrived —
so sign-up number fifty would wait days behind routine maintenance. The fix is not "make
everything faster", it is: a priority lane for new customers, a spending governor so a burst
can't blow the monthly cap mid-build, and an honest "your site is building, ready in about
X" in the sign-up flow. We also found the Cloudflare account has an unpublished limit of
roughly a thousand zones — at fifty domains a day that arrives in weeks, so the already-
agreed "plan B" for DNS may need to be built before the first big promotion, not after.

**Your three questions, answered short:** more repos — no (you already ruled one repo per
site out, and the evidence agrees); another cluster — not for capacity, the current one is
idle; the future shape is self-contained "satellites" when a single cluster genuinely fills
up; the deploy path — yes, that one really is a bottleneck that grows with domain count, and
there is a concrete choice to make about it.

What happens next: a diagnosis run is verifying the dispatcher finding independently (our
own rules require that before we write it down as fact), and the full research document with
every option costed and a decision list for you is being written now. Nothing in the live
system has been changed.
