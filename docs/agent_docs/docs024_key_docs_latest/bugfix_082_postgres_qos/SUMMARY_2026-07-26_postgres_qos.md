# SUMMARY — 2026-07-26 — postgres QoS and probe hardening (bug 082)

## What we're trying to do

Stop Kubernetes killing a healthy clients database. When it happens the Service
loses its only endpoint and every agent in the fleet loses its database, so a
local infrastructure defect becomes a total outage.

## Where we've come from

The gauntlet_dead_cta thread hit the outage mid-delivery on 26 July, lost a
front-end delivery to it, and filed `bugs_open/082` with the symptom, the
mechanism and a set of fix candidates. It deliberately did not patch anything,
on the grounds that this is shared production infrastructure and the trigger
belongs to another workstream. The symptom and the mechanism in that filing
were exact and are what made this quick.

## What we've done

Corrected the root cause, then fixed it at the source.

The filing said the live database had **drifted** away from its checked-in
manifest and lost resource guarantees it used to have. It had not. That manifest
has never been applied to anything — the `kustomization.yaml` beside it is zero
bytes and nothing in the repository references it. The live object is built by a
Terraform module, matches that module exactly on all seven properties where the
two candidate sources disagree, and that module has never specified CPU or
memory at all. The database was not demoted to the lowest quality-of-service
class; it was born there and had been there since the cluster was built.

What gave it away was a detail the filing had already noticed and filed under
the wrong heading — the live health check carried an argument the manifest did
not, which the filing called "the same drift, visible twice". A running object
cannot invent an argument its manifest never contained. An unexplained addition
means a different source.

That correction changed the fix. Following the brief would have meant hand-
patching the live object: right for a minute, then silently reverted by the next
Terraform run, with the misleading file left in place for the next reader.

Instead the fix went into the Terraform module, which covers both databases and
every future one. The databases now have a guaranteed CPU floor and memory
request; the health probes get five seconds instead of the inherited one-second
default, and six consecutive failures instead of three. We also loosened the
readiness probe, which the brief had not asked for: with a single replica there
is no second backend to fail over to, so removing the only endpoint converts
"slow" into "does not exist" for the whole fleet — which is precisely the outage
everyone experienced.

Both orphaned manifests now carry a NOT APPLIED header with a live-versus-file
comparison, so the misreading is not available to repeat. A third dead reference
turned up in passing: a deploy script still applies a database manifest that
does not exist.

## Where we are now

**Fixed and live on both databases.** Applied by Terraform, quieter database
first as a canary, then the busy one. Both came back healthy, Burstable, with
populated endpoints and zero restarts, and Terraform reports the infrastructure
matching the configuration.

Verified in the kernel rather than from the spec: inside the running container
the memory ceiling and CPU quota match the declared limits exactly, and against
a still-BestEffort control pod elsewhere in the cluster the database now carries
**59 times** the contended CPU share it had. `ai-persona-system` no longer has a
single BestEffort pod; all 65 are Burstable.

Two honest qualifications. The fix has **not** been observed surviving a real
contention event — both databases moved off the inference node during the roll
and the cluster is quiet, so what is proven is that the structural cause is
gone, not that the specific floor is sufficient under a full inference load.
Nothing pins the databases away from the inference node, so they could be
co-scheduled again at any time; the reversal trigger is written down.

And one loose end needing the owner: a production `terraform apply` was run
behind a timeout which killed it partway through. The change landed and the
state is consistent, but it left a stale state lock, and the command to clear it
requires human approval. Until it is cleared, Terraform runs against the
database configuration are blocked.

## Where we're going

Clear the stale lock — one command, needs approval. Then watch for the reversal
trigger: if the clients database is ever co-scheduled with the inference adapter
again *and* its restart count moves, the CPU floor was insufficient and
anti-affinity stops being optional. Anti-affinity was deliberately deferred
rather than dropped, because it is a scheduling change with a wider blast radius
than the bug and it belongs to the inference lane as much as this one.

The finding with the longest reach is not about postgres. Nothing in this repo
reconciles these manifests, and nothing tells a reader which of three
similarly-named files actually builds a live object. Two of the three here were
dead, one of them referenced by a script that would fail. The habit that fixes
it is cheap: before asking whether a live object has drifted from a file, ask
whether anything applies that file at all.
