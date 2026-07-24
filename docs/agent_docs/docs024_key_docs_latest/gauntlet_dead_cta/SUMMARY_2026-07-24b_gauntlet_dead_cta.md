# SUMMARY — how the public reaches the engine: exposure design corrected, and a fork in the road (2026-07-24b)

Companion to this morning's summary (SUMMARY_2026-07-24), which covers the
machine-built backend. This one covers the half that moved since: how the public
internet reaches the new debate engine without endangering the production cluster.

## What we're trying to do

The Gauntlet's debate engine needs to be callable from a static public website.
That means, for the first time, a path from the open internet to something we run —
and the production cluster behind it builds and serves every site we have. The job
is to open exactly one door, and to be able to say precisely what an attacker gets
if the worst happens on the far side of it.

## Where we've come from

Yesterday's decision was: engine inside the cluster; public traffic through
Cloudflare, then a tunnel, then a small bastion machine we own, then a WireGuard
link into the cluster, with a network rule so the bastion could reach only the
debate engine and nothing else. Three things blocked it, all the owner's: name the
subdomain, provision the bastion machine, approve the VPN peering.

## What we've done

The subdomain is named — **tools.apis.uk** — and the domain is verified properly
live on Cloudflare (the nameserver move worked; one leftover catch-all DNS entry
points at nothing and should be deleted). The step-by-step for creating the tunnel
is written and takes about ten minutes on the new machine.

The owner asked for the peering design to be double-checked against the real
cluster, and it failed the check — in two ways. The VPN inside the cluster rewrites
the sender's address on everything it forwards, so the planned "only the bastion's
address may pass" rule was checking an address that never appears; and that VPN is
the owner's admin door — anything connected to it can reach everything, database
included. Fine for a laptop; not for an internet-facing machine. The corrected
design gives the bastion its own separate, single-purpose doorway, walled three
ways (one key, forwards to one service only, and pinned by the cluster's own
network enforcement — confirmed real and active — so even a fully compromised
doorway reaches nothing else). That design is drafted and on file.

Provider: Mythic Beasts recommended — independently British-owned for twenty-five
years, and the smallest box (about £7–9 a month) is ample. Worth knowing: a Hetzner
box would not be UK-based at all.

Then the owner asked a better question: could the engine run somewhere else
entirely, off the cluster? An audit of the approved build plan says yes, and more
easily than expected: the engine needs no Kafka and no Kubernetes — it is a plain
web service plus one database table. It talks outward only (to Anthropic, and to
the public site for the day's provocation); its single tie to the platform is a
lookup of which sites may call it, which is a one-line change to read from
configuration instead.

## Where we are now

One decision is open, and it is the owner's:

- **Route A — in the cluster, behind the corrected bastion chain.** Public traffic
  terminates inside the production cluster, contained by the three walls above.
- **Route B — an island.** One small UK-owned VM runs the engine, its database, the
  proxy and the tunnel. Public traffic never touches the production cluster in any
  form; there is nothing to tunnel into it and no peering to get right. Worst case
  on the island: the debate records and a deliberately separate, spend-capped AI
  key — the cluster is unreachable from there.

Both routes are fully designed and on file; neither blocks the build, because the
code is identical either way — only where it runs changes. Meanwhile, on the build
track, the platform fix shipped live (v1.0.1155), one more implementer round was
refused for a path deviation our own corrective rule had accidentally suggested
(rule rewritten), and the next round is in flight.

One housekeeping item: while inspecting the live VPN its keys were printed into a
local session log. Low risk — it never left the owner's machine — but the
laptop/phone VPN keys should be re-issued at a convenient moment; five minutes.

## Where we're going

The owner picks Route A or B, and orders the VM (either route needs one — as
bastion or as island). If B: the VM setup files and amended runbook get written,
the one configuration tweak is noted for the pull request, and the database
migration applies to the island's own database. Either way the sequence after the
choice is the same: the implementer finishes, the pull request lands for the
owner's review — the hard gate — the engine deploys by the chosen route, the tunnel
goes live, an outside smoke-test proves the API answers, and the experience loop
re-fires with that liveness evidence to govern the front-end rebuild.
