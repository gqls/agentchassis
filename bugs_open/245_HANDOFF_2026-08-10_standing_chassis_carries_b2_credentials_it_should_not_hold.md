# 245 — the standing agent-chassis deployment carries AWS/B2 storage credentials it should not hold (owner directive), and removing them naively breaks every storage agent's spawn

**Filed 2026-08-10** by `staged_component_build`, at the owner's direction, given in the
same breath as `bugs_open/243`'s fix: *"the agent chassis shouldn't carry b2 credentials,
please write that up as a todo/bug."*

**Status: OPEN — a directive with a landmine in its path, not yet an implemented change.
Nothing here is committed as a fix.** This is deliberately a todo/bug write-up: the
removal itself is small, but doing it in the wrong order breaks storage injection for
every spawned storage agent, silently, at the next spawn.

## What is there today (read from the live overlay, 2026-08-10)

`deployments/kustomize/services/agent-chassis/overlays/production/uk_001/patch-deployment.yaml`
lines 77–98 put four credential env vars on the **standing** chassis deployment, all via
`secretKeyRef` from `personae-storage-secrets`:

- `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`
- `B2_APPLICATION_KEY_ID`, `B2_APPLICATION_KEY`

They predate the 2026-08-05/08 episode (`019cf8d94`) and were **explicitly excluded** from
the `820a033c0` revert — the overlay comment says removing them "is a separate question
with a real blast radius". This file is that separate question, written up.

Confirmed on the running pods (2026-08-10, both replicas): `B2_APPLICATION_KEY` and
`B2_APPLICATION_KEY_ID` present; `IMAGE_BUCKET`/`S3_ENDPOINT` absent — so the standing
chassis holds **credentials without bucket config**: it cannot build its own storage
client (the `agentbase/agent.go:316` gate fails), yet it holds the keys. Worst of both:
no capability, full secret surface, on the deployment whose image runs ~46 pods.

## Why the credentials are there — the blast radius that blocks naive removal

The spawn path **launders credentials through the orchestrator's own environment**.
`spawn_actions.go:2556-2580`: when spawning a storage-enabled agent, the code reads
`os.Getenv("AWS_ACCESS_KEY_ID")`, `os.Getenv("B2_APPLICATION_KEY_ID")` etc. **from the
chassis pod's own env** and injects the *values* into the spawned pod's spec as plain
`Value` env vars.

So the standing chassis carries B2 credentials **because it is the spawner**, and the
spawner copies from itself. Two consequences:

1. **Naive removal breaks spawns silently.** Strip lines 77–98 and the next spawned
   storage agent (site-publisher, image-generator, asset-deployer, … the 13-name list —
   which now includes `tool-acceptance-agent`, `bugs_open/243` fix) gets empty
   credentials: `os.Getenv` returns `""`, the injection block **skips empty values
   without error** (`if awsKeyId != "" {…}`), the pod starts, and every storage
   operation fails at first use. The spawn itself succeeds; nothing fails at the point
   where the mistake was made.
2. **The values-not-references pattern is itself the exposure.** A spawned pod's spec
   embeds the credential *strings* — readable by anything that can `kubectl get pod -o
   yaml`, logged wherever pod specs are logged, and unrotatable without respawning. The
   standing chassis's own copies are `secretKeyRef` (better), but the spawner turns
   references into values downstream.

## Fix candidates, ordered by what closes the door

1. **Convert the spawn injection from env-copying to `secretKeyRef`, then remove the
   chassis env.** The pattern already exists **in the same function**: the
   `GITHUB_READ_TOKEN` injection for repo-cloning agents uses
   `SecretKeyRef{personae-platform-secrets, GITHUB_READ_TOKEN}` (spawn_actions.go ~2540),
   precisely so the token never has to live in the spawner. Do the same for the four
   storage credentials against `personae-storage-secrets` — the secret the chassis
   overlay already references, so it exists in the namespace. Then delete overlay lines
   77–98. Order is load-bearing: **code change → image roll → verify a spawned pod gets
   credentials via secretKeyRef → then the overlay edit.** After both: the spawner never
   holds storage credentials, spawned pods hold references not values, and rotation
   works without respawn-the-fleet.
2. **Also check the other consumers of the chassis's AWS_*/B2_* env before removal.**
   Measured 2026-08-10: `grep os.Getenv` over `platform/ internal/ cmd/ pkg/` for the
   four names returns **exactly the spawn block's four lines** (spawn_actions.go:
   2558-2561) — no other direct reader. Indirect readers exist by *name reference*
   (`config.ObjectStorageConfig{AccessKeyEnvVar: "B2_APPLICATION_KEY_ID", …}` in
   `agentbase/agent.go:311` and the adapters' storage config), but those fire only
   where a client is actually BUILT — spawned storage pods and adapters — never on the
   standing chassis, which has no `IMAGE_BUCKET` and skips construction. A grep proves
   absence only for the spelling it searches: re-run BOTH greps (literal `os.Getenv`
   and `AccessKeyEnvVar`) at removal time; the tree moves ~1,500 commits/week.
3. **Verification for the removal itself**: after the overlay edit rolls, on each
   standing chassis replica `env | grep -c B2_` → 0 (and the negative control:
   `GITHUB_READ_TOKEN` still absent there too); then spawn one storage agent and one
   repo-cloning agent and confirm both work — the storage one proves the secretKeyRef
   path, the git one proves nothing else was disturbed. A spawn that *succeeds* is not
   the proof — the storage OPERATION inside the spawned pod is (candidate 1's failure
   mode is a healthy-looking pod that fails at first use).

## Relations

- `bugs_open/243` — the fix that just widened the storage list by one type; its
  submission's risks block names this file as the sequencing constraint.
- Owner ruling 2026-08-08 (overlay comment; register MDL-040): credentials must not be
  spread across agents. Candidate 1 is that ruling taken to its conclusion: today the
  ruling keeps bucket *config* off the chassis while the chassis still holds the *keys*;
  after candidate 1 the keys live only in the secret and the pods that need them.
- Overlay comment's own caveat (lines ~123–127) and `WRONG_CALLS.md` 2026-08-08: this
  file exists because that comment asked for the question to be handled deliberately.

---

## CONTRIBUTION 2026-08-10 (evening), same lane — the operational half, measured by accident

The "credentials without bucket config" reading above is right, and it has a **visible
operational cost** that this file does not yet record: with `IMAGE_BUCKET` unset, the
standing chassis is not merely *holding keys it cannot use* — it is the consumer of
`system.agent.generic.requests`, so **any orchestration dispatched at that topic which
runs `deploy_image_asset` fails outright.**

Measured tonight, twice, on separate correlations
(`71335ce7-042e-448e-9cf9-9227404c1c14`, `d893fcf4-b364-4a39-a2d3-d2e6bd45451c`):

```
FAILED | deploy_asset | step deploy_asset failed: failed to execute action
                        deploy_image_asset: storage client not available
```

Both replicas confirmed bare, same exec, with a positive control in the same command:

```
IMAGE_BUCKET=[] S3_ENDPOINT=[] B2_KEY_ID_SET=[yes]     # x2 replicas
```

And the pods that DO have the bucket, enumerated across all running pods:
`agent-build-dispatch-loop` (×2) and `image-generator-adapter` (×2) —
`personae-prod-uk001-images`; plus `business-intel` on a different bucket.

**Why this matters for candidate 1's sequencing.** The file already says naive removal
breaks spawned storage agents. The converse is worth stating too: **the chassis's missing
`IMAGE_BUCKET` is not harmless today** — it is why the only documented operator route for
deploying an asset (`docs/leopardessconsulting/scripts/deploy_brand_asset.sh`, which
publishes to exactly that topic) cannot work at all, independently of that script's other
staleness. A reader of this file could reasonably conclude "no capability" means "no
consequence". It means the deploy path has one working route (the work item, via
`build-dispatch-loop`) and one that always fails, with nothing naming which is which.

Not a new fix candidate — candidate 1 still stands and is still correctly ordered. This is
evidence for its *urgency* and a caution for whoever verifies it: **after the change, prove
a spawned storage agent's deploy at the ARTEFACT, not at the spawn.** A pod that starts is
this bug's failure mode, not its refutation.

Full context of how this was found, and the two deploy defects it sits behind:
`bugs_open/248`.

— `staged_component_build`, 2026-08-10
