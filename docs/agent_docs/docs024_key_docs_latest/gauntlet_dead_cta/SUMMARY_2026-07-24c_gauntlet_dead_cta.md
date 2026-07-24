# SUMMARY — the hosting question grows up: could the framework itself live off-cluster? (2026-07-24c)

Third read-out today, because the question changed again. The morning file
(SUMMARY_2026-07-24) covers the machine-built backend; the midday file (…24b)
covers the corrected bastion design and the bastion-versus-island fork. Since
then the owner escalated the island idea into something bigger, and it holds up.

## What we're trying to do

Give the Gauntlet's debate engine a public home that cannot endanger the
production cluster — and now, additionally, decide how much of our platform that
home should be able to run. The owner's reasoning: the API may become genuinely
busy, and it should be able to use our workflows and agent machinery, not just
answer HTTP — and without Kafka there is no framework. Both halves are correct.

## Where we've come from

The fork on the table this morning was: Route A, tunnel public traffic into the
production cluster behind a hardened bastion; or Route B, a small isolated VM at
Mythic Beasts running just the debate engine and its database. Route B's appeal
was total isolation — public traffic never touches production at all. Its
apparent limit was that it was engine-only: no Kafka, no workflows, no framework.

## What we've done

Measured whether that limit is real, against the live cluster rather than
intuition. Three findings.

First, the framework needs more than a bare VM, but not much more: the agent
system creates and destroys its worker pods on demand — a third of the pods
running in production right now were spawned that way — so it needs a real
Kubernetes underneath. A lightweight one (k3s) on a VPS satisfies it; a plain
docker-compose box cannot.

Second, the framework is far lighter than it looks. Everything except Kafka —
chassis, core manager, scheduler, all the spawned agents — measured under a
couple of cores and about three gigabytes today. Kafka is the heavyweight, but
production's three four-gigabyte brokers are fleet-scale; a single-node Kafka in
about two gigabytes is entirely adequate for an island. The whole framework
island fits a £37-a-month machine; £70 buys generous headroom.

Third, a trap identified in our own repo: we already have a half-built
multi-cluster feature that looks like the obvious answer — production
dispatching agents to a second cluster. It is the wrong tool here, because it
works by sharing production's Kafka and database with the remote cluster, and
our Kafka currently has no internal access control at all. Wiring a
public-facing island in that way would hand a successful attacker production's
entire nervous system. If the island runs the framework, it must be a second,
fully independent instance: its own Kafka, own database, own keys, nothing
shared, talking to production — if ever — only through authenticated APIs.

## Where we are now

The decision is now three-way, all designed, none chosen:

- **Route A — in the cluster, behind the corrected bastion.** Unchanged from
  the midday summary.
- **Route B1 — minimal island.** One ~£8/month VM, engine and database only.
- **Route B2 — framework-ready island.** The same island, but with k3s from day
  one (~£37/month), the engine deployed into it now, and the one-node Kafka plus
  core framework services added in place if and when load or ambition justifies
  them. No re-platforming between B1's capability and the full framework — it is
  an upgrade on the same machine.

The recommendation is B2: it costs a little more than B1, keeps every option
open, and honours the isolation argument completely. A sizing honesty note: the
engine's speed limit is the AI model's latency and cost, not hardware — even the
small box carries a lot of debates. Open items if chosen: an image-registry path
for the island, and the owner takes on bounded operations (roughly a weekend to
stand up, an hour a month after). Meanwhile the build track keeps cycling —
another implementer round failed a build gate mid-afternoon and a further run is
executing as this is written; the deterministic gates continue to hold.

Also named rather than smuggled: a framework island on British-owned compute is
de facto the start of the UK-sovereign-stack exploration the owner parked a
fortnight ago. If that is to be pursued seriously it deserves its own planning
thread; it has been flagged, not absorbed.

## Where we're going

The owner picks A, B1 or B2 and orders the VM accordingly (VPS 12 if B2). Then,
for B2: k3s goes on, the island kustomize overlay and runbook get written, the
tunnel goes live on the same box, and the one configuration tweak (CORS origins
from config) is noted for the pull request. After that the sequence is as
before: the implementer finishes, the PR lands for the owner's review, the
engine deploys by the chosen route, an outside smoke-test proves the API
answers, and the experience loop re-fires with that liveness evidence to govern
the front-end rebuild. The framework services follow onto the island only when
the load is real — that gate is the owner's, later, with the machinery already
under it.
