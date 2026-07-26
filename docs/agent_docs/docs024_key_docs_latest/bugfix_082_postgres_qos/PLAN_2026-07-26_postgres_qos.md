# PLAN — bugs_open/082, postgres QoS and probe hardening

**Opened & closed:** 2026-07-26 · **Status:** FIXED AND LIVE, both instances

## The brief as filed, and the correction

`bugs_open/082` was filed by the gauntlet_dead_cta thread, which hit the outage
mid-delivery. Its symptom section was exact and its mechanism was right. Its
root cause was not.

> **CORRECTED 2026-07-26 — "Fault A" in the filed handoff is REFUTED.**
>
> The handoff read the outage as the live StatefulSet having **drifted** from
> `deployments/kustomize/infrastructure/postgres-clients/postgres-clients.yaml`,
> which does specify `requests: cpu 500m` — and concluded the fix was to
> *reconcile the live object back towards that manifest*, plus item 4, "ask why
> the live object drifted at all".
>
> There was no drift. That kustomize file **has never been applied to
> anything.** Its `kustomization.yaml` is 0 bytes and no kustomization in the
> repo lists it. The live object is built by
> `deployments/terraform/modules/postgres-instance/main.tf`, and matches that
> module exactly on all seven properties where the two candidate sources
> disagree. The module simply never specified `resources` — the database was
> not *demoted* to BestEffort, it was **born** BestEffort and had been so since
> the cluster was built.
>
> **What caught it:** the handoff's own evidence, read once more. It noted in
> passing that "the live probe also carries `-d clients_db`, which the manifest
> does not — the same drift, visible twice." That is not drift visible twice.
> A live object cannot invent an argument the manifest never contained. One
> unexplained *addition* is the signature of a different source, not of decay
> from this one. Grepping for that flag found it in the Terraform module.
>
> **The cheap check that would have caught it first:** never ask "has the live
> object drifted from this file?" before asking **"does anything apply this
> file?"** — one `grep -rn "postgres-clients" --include=kustomization.yaml` and
> one `ls -la` on the sibling `kustomization.yaml` (0 bytes). Both are seconds.

Why the correction matters practically: had the fix followed the brief, it
would have been a `kubectl patch` — correct for about a minute, and silently
reverted by the next `terraform apply`, with the manifest still lying to the
next reader.

## Decisions and their reasons

**1. Fix the Terraform module, not the live object.** It is the only source
that builds anything. It also fixes both databases and every future one from
one edit — `postgres-templates` had the identical defect and nobody had noticed,
because it is colder and had never been co-scheduled with a noisy neighbour.

**2. Give the new resource variables defaults rather than requiring them.**
An instance that forgets to set them must still get a floor. Requiring them
would make "silently BestEffort" reachable again the moment someone adds a third
database in a hurry. The bad state should be unrepresentable, not merely
discouraged.

**3. `failure_threshold` 3 → 6 on the *readiness* probe too**, which the brief
did not ask for. `replicas = 1`, so there is no second backend to fail over to.
Dropping the only endpoint does not route traffic anywhere better — it converts
"queries are slow" into "no such host" for the entire fleet. Aggressive
readiness gating on a single-replica database has no upside that would pay for
the outage it causes. This is the change that actually addresses the *symptom*
the filer saw (`notReadyAddresses` only); the CPU floor addresses the *cause*.

**4. Limits deliberately above the orphaned manifest's numbers.** The manifest
said `limits: 1Gi / 1000m`; we set `2Gi / 2000m`. There was no reviewed prior
state to defer to — that file was never applied, so its numbers carry no
authority. A memory limit that is reached kills a database outright, so against
~210Mi observed RSS the limit is insurance, not a budget.

**5. Did NOT add anti-affinity**, though the brief lists it third. Deferred
with a named reversal trigger rather than dropped — see below.

**6. Did NOT run the council gate.** Out of scope by the owner's 2026-07-17
ruling (`platform/`, `internal/`, `pkg/`); `097` refuses `deployments/`
client-side. Recorded so the `098` coverage report's silence here is explained
rather than mysterious.

## Deferred, with the trigger that should reopen it

**Anti-affinity / keeping the data plane off the inference node.** Not applied.
Two reasons: it is a scheduling change to shared infra whose blast radius is
larger than the bug's, and it belongs to the ollama lane as much as to this one.

As it happens both databases *did* reschedule off the ollama node during the
roll, so they are not co-tenants today. **That is luck, not a fix, and it is
the thing most likely to mislead the next reader.** Nothing pins them apart. A
node drain, an eviction or a chassis roll can put `postgres-clients-0` back
beside `ollama-adapter` at any time.

**Reversal trigger:** if `postgres-clients-0` is ever again co-scheduled with
`ollama-adapter` *and* its restart count moves, the CPU floor was insufficient
and anti-affinity becomes the fix rather than an option. Check with the
topology command in the RUNBOOK.

## What was actually shipped

| change | file | live |
|---|---|---|
| `resources` block, 4 new variables with defaults | `terraform/modules/postgres-instance/{main,variables}.tf` | yes, both instances |
| `timeout_seconds 5`, `failure_threshold 6`, both probes | same | yes, both instances |
| NOT APPLIED headers + live-vs-file table | both orphaned kustomize manifests | n/a (docs) |

Commit `0f87d30c8`. Applied by `terraform apply`, canary on
`postgres_templates_db` first, then the untargeted plan for the clients module.
