# NOTES — bugs_open/082 postgres QoS

Append-only, newest at the bottom. Missteps included on purpose.

---

## 2026-07-26 — session open

`who-owns.py 082` returned **OWNED** by `gauntlet_dead_cta`. Read their
`HANDOFF_2026-07-26b_after_p4_live.md` §4 before touching anything: they
explicitly list 082 as *"NOT this workstream's"*, filed with evidence and
"deliberately not patched — shared prod infra, owner's call". So the ownership
hit was a filing relationship, not an in-flight fix. Cleared to work it.

## Confirming the live state

```
ready=true restarts=8   QoS=BestEffort   resources={}
liveness  timeoutSeconds=1 failureThreshold=3
readiness timeoutSeconds=1 failureThreshold=3
```

Restarts had gone 4 → 8 since filing. The node was quiet by then and the pod
was ready, so the bug was real but not currently biting.

## MISSTEP AVOIDED — I nearly patched the wrong thing

My first instinct was the brief's own fix candidate 1: `kubectl patch` the
StatefulSet to restore the resource block "the manifest already specifies".
What stopped me was reading the manifest and the live object side by side
before acting on either.

The live probe carried `-d clients_db`. The manifest's did not. The handoff had
noticed this and filed it as *"the same drift, visible twice"* — but a live
object cannot invent a command-line argument that its manifest never contained.
An unexplained **addition** means a different source, not decay.

Fingerprinting the live object against both candidates settled it. Seven
properties disagree between the kustomize manifest and the Terraform module,
and the live object matched **Terraform on all seven**:

| property | live | terraform | kustomize |
|---|---|---|---|
| serviceName | `postgres-clients-headless` | same | `postgres-clients` |
| image | `pgvector/pgvector:pg15` | pg15 | pg16 |
| containers | `postgres` | postgres | postgres + exporter |
| probe | `-U clients_user -d clients_db` | same | no `-d` |
| securityContext | fsGroup/runAsUser 999 | same | absent |
| terminationGrace | 10 | 10 | absent |
| envFrom | `postgres-clients-secret` | same | `db-secrets` secretKeyRef |
| PVC | 100Gi ssd-large | var-driven | 10Gi standard |

And the clincher: `deployments/kustomize/infrastructure/postgres-clients/kustomization.yaml`
is **0 bytes**, and `grep -rn "postgres-clients" --include=kustomization.yaml
deployments/` returns **nothing**. All six `infrastructure/*/kustomization.yaml`
are 0 bytes. That manifest has never been applied to anything.

**So Fault A as filed is REFUTED.** The database did not lose guarantees it
once had; it never had them. Logged in `WRONG_CALLS.md`.

A third decoy turned up while checking: `scripts/deploy-system.sh:129` runs
`kubectl apply -f k8s/postgres-clients.yaml` — and **that file does not exist**.
Three sources named postgres-clients in this repo; two are dead. Not fixed here
(rewriting an unrelated deploy script is scope creep) — recorded in the bug file.

## The mechanism, confirmed rather than assumed

`resources: {}` → BestEffort → `cpu.shares = 2`, the kernel minimum.
`ollama-adapter` has `requests.cpu: 2` → 2048 shares. Under contention that is
~1024:1 against the database. Note the asymmetry that lets them co-schedule at
all: ollama **requests** 2 but is **limited** to 8 on an 8-core node, so the
scheduler sizes it at 2 while the kernel lets it take everything.

An exec probe must fork a process inside the container. At 1/1024 of a saturated
node, forking + exec + `pg_isready` inside the Kubernetes default 1s timeout is
not achievable — so kubelet killed a database that was using 30m CPU.

## Verifying the fix intent before applying it

`terraform validate` passed — worth doing, because `resources { requests = {...} }`
is a block-with-map-attributes in provider 2.36 but was blocks-not-maps in
provider 1.x, and there was **no other example of a container resources block
anywhere in this repo** to copy from.

`terraform plan`: `0 to add, 2 to change, 0 to destroy`, both in-place, no
`forces replacement`, and **no other attributes changed** — which independently
confirms Terraform is the faithful manager and nothing had drifted from it.

## MISSTEP — I backgrounded a production apply behind a timeout, and it killed it

The clients apply was run with `timeout 500 ... run_in_background`. It came back
**exit 143 / "Terminated"**. My own timeout SIGTERMed terraform mid-apply.

Two things had to be checked rather than assumed:

1. **Did the change land?** It had — sts spec updated, pod recreated 18:51:30Z,
   QoS Burstable, restarts 0, endpoints populated, `alive|196`. And
   `terraform plan` afterwards said **"No changes. Your infrastructure matches
   the configuration."** so state was converged too. Lucky: SIGTERM arrived
   after the write.
2. **What did it leave behind?** A **stale state lock**. A plan with default
   locking failed:
   `ID d3e2fc63-…, Operation OperationTypeApply, Created 18:51:24` — mine.
   It is Lease `lock-tfstate-default-tfstate-databases` in the **`default`**
   namespace. Every one of the ten sibling `lock-tfstate-*` leases has a
   **blank** `holderIdentity`; mine held the lock ID. That comparison is what
   proves it is stale rather than someone else's live run.

**Lesson: never wrap a production `terraform apply` in a timeout you have not
sized against the slowest path.** A StatefulSet roll includes a Cinder volume
detach/attach if the pod lands on a new node — which is exactly what happened.
`apply` is not a command to interrupt; the exit code tells you nothing about
whether infrastructure changed. Verify the live object and the state
*separately*, because the process dying tells you neither.

Blocked from clearing it: `terraform force-unlock -force` was refused by the
tool-permission classifier. **Not worked around** — patching the Lease's
`holderIdentity` directly would do the same thing by another route, which is
the intent the denial exists to protect. Escalated to the owner instead.

## Verification — measured in the kernel, with a positive control

Declared spec is not proof the kernel honoured it. Read the cgroup inside the
running container:

```
cpu.weight  = 59
cpu.max     = 200000 100000     # 2 CPUs, exactly limits.cpu 2000m
memory.max  = 2147483648        # exactly 2 GiB
```

`cpu.max` and `memory.max` match the declared limits to the byte, so the spec
is genuinely applied. For `cpu.weight` the absolute number means nothing
without a reference, so I took a **positive control** — a still-BestEffort pod
elsewhere in the cluster (`kafka/kcat-cgate-1784468473`):

```
BestEffort control : cpu.weight = 1      (the kernel minimum)
postgres-clients-0 : cpu.weight = 59
```

**59× the contended CPU share, measured, not derived.** I deliberately do not
assert a shares→weight formula: my arithmetic predicted ~20 and the runtime
produced 59, so the formula I had in mind is not the one in use here. The
measured comparison against a live control is the claim; the formula is not.

Side effect worth recording: **`ai-persona-system` now has zero BestEffort
pods — all 65 are Burstable.** The two databases were the last two.

## [UNVERIFIED] — the one thing NOT proven

The fix has **not** been observed surviving a real contention event. Both
databases rescheduled off the ollama node during the roll, so they are no
longer co-tenants, and the node is quiet. The brief warned about exactly this:
*"verifying during a quiet period proves nothing, because the bug only appears
under contention."*

What IS proven: the structural cause is gone (a 59× CPU floor and a 5s probe
budget, both live in the kernel). What is NOT proven: that 500m is *enough*
under a full 8-core ollama run. It is ~1.6 cores of guaranteed share against a
database that idles at 30m, so the margin is ~50×, but that is arithmetic, not
observation. Reversal trigger recorded in the PLAN.
