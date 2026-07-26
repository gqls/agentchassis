# RUNBOOK — postgres QoS / probe hardening (bugs_open/082)

Commands that were hard to get right, with the gotcha attached.

## Which file actually builds the running database?

**Terraform, not kustomize.** There are three postgres-clients "sources" in the
repo and two of them are dead:

| path | status |
|---|---|
| `deployments/terraform/modules/postgres-instance/main.tf` | **LIVE** — this is the one |
| `deployments/kustomize/infrastructure/postgres-clients/postgres-clients.yaml` | orphaned; `kustomization.yaml` beside it is 0 bytes, nothing lists it |
| `k8s/postgres-clients.yaml` (referenced by `scripts/deploy-system.sh:129`) | **does not exist** |

Instantiated at
`deployments/terraform/environments/production/uk001/060-databases/main.tf`
(one module call per instance: `postgres_clients_db`, `postgres_templates_db`).

### How to prove which source is live, for any object

Do not trust the filename. Fingerprint the live object against each candidate on
properties where the candidates **disagree** — a property they share proves
nothing. For postgres-clients, seven disagreed and the live object matched
Terraform on all seven:

```bash
kubectl -n ai-persona-system get statefulset postgres-clients \
  -o jsonpath='serviceName={.spec.serviceName}
image={.spec.template.spec.containers[0].image}
containers={range .spec.template.spec.containers[*]}{.name} {end}
podSecCtx={.spec.template.spec.securityContext}
termGrace={.spec.template.spec.terminationGracePeriodSeconds}
envFrom={.spec.template.spec.containers[0].envFrom}
'
```

`serviceName=postgres-clients-headless` (kustomize says `postgres-clients`),
`image=pgvector/pgvector:pg15` (kustomize says pg16), one container (kustomize
has two), `fsGroup 999` (kustomize has none), `termGrace 10` (kustomize omits
it), `envFrom postgres-clients-secret` (kustomize uses `db-secrets` via
`secretKeyRef`). Also PVC `100Gi ssd-large` vs kustomize's `10Gi standard`.

**Gotcha:** the *empty* `kustomization.yaml` is the tell, and it is invisible in
a normal listing because it is 0 bytes, not absent:

```bash
find deployments/kustomize/infrastructure -name kustomization.yaml -exec ls -la {} \;
grep -rn "postgres-clients" --include="kustomization.yaml" deployments/
```

All six `infrastructure/*/kustomization.yaml` are 0 bytes. The grep returns
nothing — no kustomization anywhere references these manifests.

## Running terraform against production

```bash
cd deployments/terraform/environments/production/uk001/060-databases
terraform plan  -lock=false -var-file=terraform.tfvars.secret -no-color
terraform apply -auto-approve -var-file=terraform.tfvars.secret -no-color
```

**Gotchas:**

- `terraform.tfvars.secret` is **not** auto-loaded. Only `terraform.tfvars` and
  `*.auto.tfvars` are. Without `-var-file` it prompts for the DB passwords and
  hangs. `terraform.tfvars` alone holds just `postgres_storage_class`.
- The directory is already initialised (`.terraform/` + `.terraform.lock.hcl`
  are committed-adjacent and present). No `terraform init` needed, and running
  one would be the only thing that writes to the shared tree.
- Backend is a Kubernetes secret: `tfstate-default-tfstate-databases` in the
  **`default`** namespace (not `ai-persona-system`).
- `terraform validate` checks your HCL against the real downloaded provider
  schema without touching state or credentials. Run it before plan — it is what
  confirmed `resources { requests = {...} }` is a block-with-map-attributes in
  provider 2.36 (it was blocks-not-maps in provider 1.x, and there was no other
  example in this repo to copy).
- `terraform fmt` will reformat **pre-existing** misaligned lines you did not
  touch (the `security_context` blocks). Don't run it — that is unrelated churn
  in a tree several sessions share.
- Both probe/resource changes are **in-place** updates to the StatefulSet, not
  replacements. Confirm before applying: the plan must say
  `0 to add, 2 to change, 0 to destroy` and must not contain
  `forces replacement`.

- **NEVER wrap a production `terraform apply` in a timeout.** This cost a stale
  state lock on 2026-07-26. `timeout 500 terraform apply` returned exit **143 /
  "Terminated"** — my own timeout SIGTERMed terraform mid-apply. A StatefulSet
  roll includes a Cinder detach/attach when the pod lands on a new node, which
  is exactly what happened. **The exit code tells you nothing about whether
  infrastructure changed**, so verify the live object and the state
  *separately*:

  ```bash
  kubectl -n ai-persona-system get statefulset postgres-clients -o jsonpath='{.spec.template.spec.containers[0].resources}'
  terraform plan -lock=false -var-file=terraform.tfvars.secret    # want: "No changes"
  ```

  Run it in the background with **no** `timeout`, or in the foreground and wait.

### Clearing a stale state lock

A killed apply leaves the lock held. Any later `terraform` with default locking
fails with `Error acquiring the state lock` and prints the ID.

```bash
terraform force-unlock -force <LOCK_ID>
```

**Confirm it is stale before unlocking, not after.** Two checks:

```bash
# 1. is the state already converged? if yes, the dead run finished its write
terraform plan -lock=false -var-file=terraform.tfvars.secret

# 2. compare against the sibling leases — an UNLOCKED one has a BLANK holder
kubectl -n default get leases | grep lock-tfstate
```

The lock lives in the **`default`** namespace as Lease
`lock-tfstate-default-tfstate-<suffix>`, not in `ai-persona-system`. On
2026-07-26 ten sibling leases showed a blank `holderIdentity` and ours held
`d3e2fc63-…`, which is what proved it was stale rather than a live run by
another session.

**Gotcha:** `terraform force-unlock` is refused by the tool-permission
classifier and needs the owner to approve it. Do **not** route around it by
patching the Lease's `holderIdentity` — that is the same action by another
route, and the denial exists to put a human on it.

### Canary one instance at a time

```bash
terraform apply -auto-approve -var-file=terraform.tfvars.secret \
  -target=module.postgres_templates_db
```

`-target` prints a "not suitable for routine use" warning; that is expected and
is the right tool here. `postgres_templates_db` is the colder database, so a
misbehaving roll shows up on the cheaper one. Verify (below) before applying the
untargeted plan, which then picks up `postgres_clients_db` only.

## Verifying the fix

```bash
kubectl -n ai-persona-system get pod postgres-clients-0 \
  -o custom-columns='NAME:.metadata.name,READY:.status.containerStatuses[0].ready,RESTARTS:.status.containerStatuses[0].restartCount,QOS:.status.qosClass'

kubectl -n ai-persona-system get statefulset postgres-clients -o jsonpath='
resources={.spec.template.spec.containers[0].resources}
livenessTimeout={.spec.template.spec.containers[0].livenessProbe.timeoutSeconds}
readinessTimeout={.spec.template.spec.containers[0].readinessProbe.timeoutSeconds}
'

kubectl -n ai-persona-system get endpoints postgres-clients -o jsonpath='{.subsets}'
```

Expect `QOS=Burstable`, `timeoutSeconds=5` on both probes, and a populated
`addresses` array. **`notReadyAddresses` instead of `addresses` is the outage
symptom** — the Service has nowhere to send traffic.

**Gotcha — `kubectl exec` bypasses the Service.** During the outage the database
answered `kubectl exec ... psql` instantly while every in-cluster client was
failing, which is why this looks survivable from a terminal. Never conclude "the
DB is fine" from an exec. Check the endpoints object.

**Gotcha — verifying while the node is quiet proves nothing.** The bug only
appears under contention. Re-check `kubectl top node` and confirm the noisy
neighbour is still hot when you call it fixed:

```bash
kubectl top node
kubectl top pod -n ai-persona-system --sort-by=cpu | head -5
kubectl -n ai-persona-system get pods -o wide \
  --field-selector spec.nodeName=<node> | sort
```

### Verify in the kernel, not from the spec

A spec field proves what was *asked for*. Read the cgroup inside the running
container to see what the kernel actually did:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- sh -c \
  'cat /sys/fs/cgroup/cpu.weight; cat /sys/fs/cgroup/cpu.max; cat /sys/fs/cgroup/memory.max'
```

Expect `cpu.max = 200000 100000` (= 2 CPUs, matching `limits.cpu 2000m`) and
`memory.max = 2147483648` (= exactly 2 GiB). Those two match the declared
limits to the byte and so are self-evidencing.

**`cpu.weight` needs a positive control** — the number is meaningless alone.
Take one from a pod that is still BestEffort (there are none left in
`ai-persona-system`; use another namespace):

```bash
kubectl get pods -A -o jsonpath='{range .items[*]}{.metadata.namespace} {.metadata.name} {.status.qosClass}{"\n"}{end}' \
  | awk '$3=="BestEffort"'
kubectl -n kafka exec -i <besteffort-pod> -- cat /sys/fs/cgroup/cpu.weight
```

Measured 2026-07-26: BestEffort control = **1** (the kernel minimum),
postgres-clients-0 = **59**. That 59× is the fix, measured.

**Do not derive the expected weight from a shares formula.** The documented
`1 + ((shares-2)*9999)/262142` conversion predicts ~20 for a 500m request; this
runtime produced 59. Compare against a live control instead of arithmetic.

## The mechanism, in one check

Why a 1s probe was unmeetable: a container with **no CPU request** gets
`cpu.shares = 2` (the kernel minimum Kubernetes assigns to BestEffort).
`ollama-adapter` has `requests.cpu: 2` → 2048 shares. Under contention that is
roughly **1024:1** against the database. With `requests.cpu: 500m` the database
gets 512 shares — about 20% of an 8-core node under full contention, ~1.6
cores, against the ~30m it actually uses.

```bash
kubectl -n ai-persona-system get pod -l app=ollama-adapter \
  -o jsonpath='{.items[0].spec.containers[0].resources}'
# {"limits":{"cpu":"8","memory":"20Gi"},"requests":{"cpu":"2","memory":"20Gi"}}
```

Note the asymmetry that lets them co-schedule at all: ollama **requests** 2 but
is **limited** to 8, so the scheduler sizes it at 2 while the kernel lets it
take the whole node.
