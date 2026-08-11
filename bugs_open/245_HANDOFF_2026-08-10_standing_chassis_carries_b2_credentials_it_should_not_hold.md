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

---

## UPDATE 2026-08-10 (night) — candidate 1's CODE HALF is implemented, submitted, committed; the overlay half deliberately waits

Owner: *"Please go ahead with 245."* Executed in the bug's own stated order:

- **Code**: the four `os.Getenv` value-copies in the spawn storage block are replaced
  with `secretKeyRef` env entries against `personae-storage-secrets` (name overridable
  via `AGENT_STORAGE_SECRET`, mirroring `AGENT_STORAGE_CONFIGMAP`). References are
  REQUIRED, not optional — a missing key now fails the spawned pod visibly
  (`CreateContainerConfigError`) instead of this bug's silent first-use failure. All
  four key names verified present in the secret before authoring. Commit `e7e3b4e3c`,
  trailer `Council-Submitted: c45c6412-20aa-45ab-b5ae-38fcc2bd7887`. Built against
  clean git-archive HEAD (the working tree carried another session's unrelated compile
  break in `save_page_sections_action.go` + untracked `save_sections_decision_gate.go`
  — not touched).
- **Re-measured at edit time**: the four names' direct `os.Getenv` readers are now ZERO
  (the four spawn lines were the only ones, and they are gone).
- **The overlay's credential lines (77–98) are NOT removed yet.** They stay until this
  binary rolls AND a spawned storage pod's storage OPERATION is proven at the artefact
  (the same-lane CONTRIBUTION above sharpened that bar). Under the old binary the
  spawner still copies from its env; under the new one the chassis lines become inert,
  and only then is the removal safe. **Whoever does the removal: candidate 3's
  verification list above, unchanged.**

**Remains OPEN**: (a) council verdict on `c45c6412` to be read; (b) the roll; (c) the
spawned-pod proof (env sourced by secretKeyRef + a real storage operation, e.g. a
`deploy_image_asset` succeeding at the artefact); (d) the overlay edit; (e) re-run both
greps (`os.Getenv` and `AccessKeyEnvVar`) at removal time.

## UPDATE 2026-08-10 (later) — council APPROVED round 1 (`c45c6412`), and the guardian's medium objection answered with a measurement

Verdict: **APPROVED**, 1 medium + 2 low advisory objections, none high. The commit
already carries `Council-Submitted:`; 098 credits it automatically.

**The medium objection is correct and now quantified.** The touched conditional gates on
`isStorageEnabledAgent(type) || category ∈ {orchestrator, code-driven}` — so the
fail-loud change reaches more than the 13-name storage list. Measured (live
`agent_definitions`, 2026-08-10): **21 orchestrator types + 12 code-driven types** also
pass that gate. Two things bound the risk:

1. **The population receiving credential env is UNCHANGED** — the same broader gate
   applied to the old value-copies; every one of those 33 types' spawned pods was
   already being handed the four values whenever the chassis env had them. What changed
   is only the failure mode when a key goes missing from `personae-storage-secrets`:
   silent-empty-env (break at first storage use) → `CreateContainerConfigError` (break
   visibly at spawn), now for all 33+13 types.
2. **The trigger condition is a secret-key rename/removal** — all four keys verified
   present 2026-08-10, and the overlay's own chassis env references the same keys, so a
   rename would already have broken the chassis deployment itself. But the guardian's
   scenario is real: rotate the secret with a renamed key and every orchestrator spawn
   fails at start until it is fixed. That is the fail-loud design doing its job at a
   wider blast radius than the plan's prose stated; whoever rotates that secret should
   know spawns gate on those exact four key names.

**Low objection (empty-string `AGENT_STORAGE_SECRET`)**: already the safe semantics —
the code tests `== ""`, so set-to-empty and unset both fall back to
`personae-storage-secrets`. **Low objection (stability tracking on spawn_actions.go)**:
noted for the record; both of today's touches were single-block, pattern-mirroring edits.

## UPDATE 2026-08-11 — rolled (v1.0.1284), proven at the spawned pod AND at the artefact; overlay credential lines REMOVED

The remaining steps from the 08-10 night update, executed in order:

- **(b) The roll**: fleet on v1.0.1284 (chassis pods up 09:26Z). Binary proof by pod-grep,
  both replicas: `personae-storage-secrets` 1, `AGENT_STORAGE_SECRET` 1, and the true
  negative `has_aws` (a logging key the commit removed) 0. *Method note: the first negative
  control tried — `Injecting storage credentials` — greps 1 on the FIXED binary, because the
  new code logs `Injecting storage credentials (secretKeyRef)`; a negative control must be a
  string the change removed IN FULL, not a prefix of its replacement.*
- **(c) The spawned-pod proof** (rode `bugs_open/243`'s proof run, work item `ae33ed59…`):
  pod `agent-tool-acceptance-agent-649a6c11-q9mlk`, spec captured live — all four of
  `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `B2_APPLICATION_KEY_ID`,
  `B2_APPLICATION_KEY` are **`valueFrom: secretKeyRef → personae-storage-secrets`**; no
  credential string anywhere in the spec. **And the storage OPERATION at the artefact**: the
  run's `look` step downloaded both acceptance screenshots from B2 inside that pod (2 images
  sent to the vision model, 0 dropped, run `complete` with no step error) — a real
  authenticated B2 READ using the reference-resolved keys. The WRITE path was not separately
  exercised this run; it uses the same client and the same key (`deploy_image_asset` remains
  the canonical write proof if anyone wants belt-and-braces).
- **(e) Both greps re-run at removal time**: direct `os.Getenv` of the four names across
  `platform/ internal/ cmd/ pkg/` → **0**. `AccessKeyEnvVar` sites enumerated: `agentbase`
  (gated on `IMAGE_BUCKET`, never fires on the chassis), `storage_actions.go` (client built
  from step config), `platform/storage/s3.go` (the reader all of them funnel into) — and one
  worth naming: **`prepare_training_data_action.go:88-93` builds a B2 client
  UNCONDITIONALLY** (no bucket gate, hardcoded B2 env names, bucket `finetuning`). Its owning
  agent `training-data-preparer` is already in `storageAgents` (spawned path gets
  secretKeyRef), zero recorded orchestration runs; an inline run on a bare chassis would now
  fail **visibly** at `NewS3Client` ("failed to construct") rather than silently — acceptable
  and recorded.
- **(d) The overlay edit**: lines 76–98 of
  `deployments/kustomize/services/agent-chassis/overlays/production/uk_001/patch-deployment.yaml`
  removed, replaced by a dated tombstone comment pointing here; the stale "credentials above"
  note at the bottom of the 08-08 block corrected. `kubectl kustomize` over the overlay
  builds clean with **0** occurrences of the four names. The change is committed config —
  it reaches the standing deployment at the next `apply -k`/release; until then the running
  chassis pods keep carrying credentials that nothing reads (the safe direction, as the
  08-10b handoff noted).

**Same-class leftover, observed while reading the spawned pod's spec — deliberately not
acted on**: `FIRECRAWL_API_KEY` still travels as a plain `value:` copied from the spawner's
env (`spawn_actions.go:2649-2653`) — the exact `os.Getenv`→Value shape this bug removed for
storage keys, including the silent skip-if-empty. Different key class (third-party SaaS key,
not the storage credentials the owner's directive named), so it is recorded here as scope
for a follow-up decision rather than smuggled into this fix.

**State: everything this file asked for is done and proven except the apply of the overlay
removal, which rides the next release.** After that apply, run candidate 3's residual checks:
`env | grep -c B2_` → 0 on each standing replica.

## UPDATE 2026-08-11 (afternoon) — OWNER DECIDED the FIRECRAWL leftover: convert it. Done, same shape as the storage fix — but via the allow-list, which also fixes a spawner drift

Owner (2026-08-11, in chat): *"convert it to a secretKeyRef."* Implemented not as another
inline secretKeyRef block but by adding `FIRECRAWL_API_KEY` to
`platform/agentenv/provider_keys.go`'s `providerKeyNames` and deleting the value-copy in
`spawn_actions.go` — because the allow-list is THE single place both spawners read
(`bugs_open/112` is why that package exists), and the value-copy had the 112 drift too:
`cmd/remote-job-spawner` injects no Firecrawl key at all, so any Firecrawl-using agent
spawned remotely was already broken silently. `FIRECRAWL_API_URL` stays a value
pass-through (endpoint, not a secret). `personae-default-secrets` verified to hold the
key. Built clean against `git archive HEAD` + the two files (agentenv, actions,
remote-job-spawner packages). Council: `Council-Submitted:
6f13c5ce-91ae-4b4a-8c80-37e8b35436ec` (verdict to be read by whoever is next in the lane).
Inert until the next image roll; after it, a spawned pod's spec should show
`FIRECRAWL_API_KEY` as `secretKeyRef → personae-default-secrets` — same capture method as
the storage proof above.

## UPDATE 2026-08-11 (afternoon) — RESIDUAL PROVEN on v1.0.1286: the standing chassis carries ZERO storage credentials. Nothing further is owed on this bug.

The overlay removal (c2c9e6a18) reached the standing deployment with the 12:02Z
v1.0.1286 release. Measured on BOTH replicas (`agent-chassis-867b7cff84-{l2bwt,twzdn}`):
`env | grep -c '^B2_|^AWS_ACCESS|^AWS_SECRET'` → **0**. With the earlier proofs (spawned
pod secretKeyRef capture, the authenticated B2 read inside the acceptance run, both
source greps, kustomize building clean), every candidate and residual in this file is
now done and live. The FIRECRAWL follow-up has its own trail (`f56abaadf`, corr
`6f13c5ce` — APPROVED on its second round, verdict read 2026-08-11). File stays in
`bugs_open/` per owner practice.
