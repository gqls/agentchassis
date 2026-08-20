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

## 2026-08-19, later — the research document is written, and we measured a real build

The full review is now in this directory (RESEARCH_2026-08-18_throughput_to_thousands_of_
domains.md). The headline addition since the morning: we measured what one new domain
actually costs, using loanzy.uk, which was built yesterday. One site: about two hundred
pieces of work, four hundred and ten agent runs (a fifth of which failed internally and
were retried — worth knowing before a burst), about twenty dollars of AI calls, and ten and
a half hours from submission to mostly-live — most of that queueing behind other sites, not
working. We also found the AI bill is dominated not by site work but by the platform's own
review council, which grows with how much we change the platform, not how many domains we
host — good news for per-domain economics, and it means the per-domain running cost still
needs one clean measurement before pricing anything. The document ends with a decision
list; the two that block everything else are: how much upkeep per domain per day do you
actually want, and how big a sign-up burst should we design for, with what promise on
"your site will be ready in…".

## 2026-08-20 — the decision list, explained in plain terms (owner asked)

The two that size everything: D0a is how much upkeep per domain per day you actually want
once a site is built — the whole requirement swings 100× on it (one item/domain/day for
3,000 domains is nearly reachable with config changes; "hundreds of thousands of jobs a
day" needs the structural tier). D0b is the burst: the peak signups/day to design for and
the promise in the signup flow ("ready in about X hours") — the promise sets the required
drain rate, and one measured build (213 items, 410 runs, ~$20 AI, 10.5h) says fifty
signups in a day exceeds what the whole fleet currently does.

The queue (D1–D3): D1 is how many dispatch turns run in parallel — each also multiplies
spend, default stop at 2, and 2 is the safe limit until the adapter work is done. D2 is
who gets served first — today strictly oldest-item-first fleet-wide, which under a burst
puts a paying customer's build behind days of old maintenance; the priority-lane constant
is yours because any priority scheme risks starving someone. D3 is whether batch size and
the scheduler's timeout always move in lockstep (default: yes).

The AI account (D4–D7): D4 is a spending governor — nothing in the code limits AI spend,
the monthly cap was hit twice in eleven days, and without a governor a successful
promotion likely ends in a mid-burst AI outage; it's a promotion prerequisite. D5 is
whether maintenance may fail over to a second provider (the stay-on-Sonnet ruling was made
for the council, where caching makes mixing costlier). D6 is which work classes may wait a
day for half-price batch processing. D7 is the Anthropic account tier itself.

The two forks to pick before code is written: D8 — pages reach production either by the
platform writing storage directly (git becomes the audit record) or by keeping the
Actions-per-commit path with batching and more runners; mutually exclusive investments.
D9 — either fix the scheduler properly or retire polling for workers pulling from the
database queue (a pattern the chassis already proved); same exclusivity; the sibling-row
stopgap is reversible under both.

How we work: D10 — releases are owner-only, serial, no CI; decide on CI/delegation.
D11 — worktrees for code sessions, deferred in July at a quarter of today's commit rate.

The estate: D12 — DNS plan B timing is calendar-shaped: at fifty domains/day the ~1k
Cloudflare zone cap arrives about three weeks into a promotion, so plan B may need
building BEFORE it. D13 — satellites: domains per satellite, the second-satellite
trigger, and whether to build the five cheap seams now.

Hygiene: D14 spot-node floor/autoscaling; D15 a backlog ceiling and whether maintenance
pauses during bursts; D16 retention for the two database tables growing toward the 100 GB
disk. Defaults exist for everything but amount to "stay as we are" — fine for details,
risky for D0b/D4/D12 if a promotion is coming. Rough answers to D0a and D0b turn the rest
into a costed, ordered build queue.
