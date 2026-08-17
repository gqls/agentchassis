# 237 — `render-audit-adapter` is in no release path, so it has never been rolled

**Status: FIXED IN THE MAKEFILE 2026-08-10 (owner asked for it directly).
STILL OPEN until a release actually rolls the pod off v1.0.1194** — the repo
change cannot move a running pod, and releases are whole-fleet and owner-run.

Found 2026-08-09 while verifying `bugs_open/233`'s fix at the pod: every
storage-touching service had picked up the fix on v1.0.1274 except this one,
which is still serving the pre-fix binary and still has a B2 credential in its
live log buffer.

## Mechanism

`render-audit-adapter` deliberately has no binary of its own — it runs the
**browser-runner** image with a different topic and consumer group
(`deployments/kustomize/services/render-audit-adapter/overlays/production/uk_001/`).
Its overlay pins the tag and its own comment says:

```yaml
# Pin the image tag for this cluster. Keep exactly one tag-pin line — the
# release tooling seds it.
```

**The release tooling does not sed it.** There are two tag-update mechanisms in
the makefile and this service is in neither:

1. Per-service hardcoded lines (`makefile:952-1031+`), one
   `sed -i.bak 's/newTag:.*/…/' …/services/<name>/overlays/…` per service —
   `agent-chassis`, `reasoning-agent`, `web-search-adapter`, `web-scrape-adapter`,
   `git-adapter`, `image-generator-adapter`, `thunder-adapter`, `analyser-adapter`,
   `browser-runner-adapter`, `content-creator-agent`, `remote-job-spawner`,
   `kafka-scheduler`, … **no `render-audit-adapter` line.**
2. `update-kustomization-images` (`makefile:918`), a `for agent in …` loop over
   another **hardcoded list of 11** — also no `render-audit-adapter`.

`grep -n "render-audit" makefile` → **zero hits anywhere in the file.** It has
no build target either; it does not need one (it reuses the browser-runner
image), but that is exactly why nothing ever bumps its tag.

Consequence: the overlay has been untouched since the pod was created
(`git log … -- <overlay>` → single commit `0143a693e`, "feat(render-audit):
give the audit its own pod, own logs, own failure state"), so it is frozen at
**v1.0.1194 while the fleet is on v1.0.1274 — 80 tags behind.**

**This is not only about credentials.** Every browser-runner fix since 1194 —
every render, screenshot and check-vocabulary change — has been reaching the
browser-runner pod and *not* the render-audit pod, which runs the same binary
against a different topic. Any lane that verified a browser-runner fix at the
browser-runner pod and assumed the audit path had it too was wrong.

## Census (with a control, because the first attempt was inert)

All 30 production overlays, tag extracted and compared to the fleet tag:

| service | pinned tag | note |
|---|---|---|
| `render-audit-adapter` | **v1.0.1194** | 80 behind; runs browser-runner binary |
| `component-render-check` | v1.0.1258 | own image; links storage, never calls it |
| `shared-output-fields-check` | v1.0.1265 | own image |
| `github-actions-runner` | **v1.0.948** | huge drift; own image, own concern |
| `github-actions-runner-vmsites` | v1.0.1126 | shares the runner image |
| `ollama-adapter`, `ollama-eval` | `latest` | third-party image, unpinned |
| everything else (22) | v1.0.1274 | current |

> **MISSTEP, recorded because the check was inert, not merely wrong.** My first
> census used `grep -A1 "images:" | grep newTag` and printed `<none>` for **all
> 30 services** — including `render-audit-adapter`, which I had just read as
> `newTag: v1.0.1194`. The `newTag` line is two lines below `images:` (the
> `name:` line sits between), so `-A1` could never reach it: the census would
> have reported "no tags anywhere" whether or not any tag existed. It was caught
> only because I happened to hold a known-positive. **A census needs a row whose
> value you already know, or it cannot come out false.**

## Fix candidates, ordered by what closes the door

1. **Make the tag-update mechanism enumerate the filesystem, not a hand-list.**
   `update-kustomization-images` already globs nothing — replace both hardcoded
   lists with a loop over `deployments/kustomize/services/*/overlays/$(OVERLAY_PATH)/kustomization.yaml`,
   skipping overlays whose image is third-party (`ollama/ollama`). Then a new
   service is covered by existing, not by remembering. This is the only
   candidate that makes the bad state unrepresentable; the others rely on
   someone remembering.
2. **Add the missing `sed` line** for `render-audit-adapter` (and decide
   `github-actions-runner*` separately — 948 is drift of a different order and
   may be deliberate). One line, immediate, but leaves the class open: the next
   image-sharing service repeats this exactly.
3. **Fail loudly instead**: a check that greps every overlay's pinned tag and
   errors when one is more than N tags behind the fleet. Detection, not
   prevention — but it would have caught this months ago, and the estate already
   runs several such CronJob checks.

Do **not** "fix" this by `kubectl delete pod` on render-audit — the overlay pins
v1.0.1194, so it returns on the same stale image (and, per `bugs_open/233`,
writes a fresh plaintext credential line when it restarts).

## Why no 090 run (declared per the OWNER RULING 2026-07-31)

The claim is structural — "a service is in no release path" is a statement
about shared tooling — so the ruling applies. Substituting first-hand
verification, stated plainly rather than omitted:

- `grep -n "render-audit" makefile` → 0 hits (absence proven for that spelling;
  the service dir is spelled `render-audit-adapter` everywhere else, checked).
- Both tag-update mechanisms read directly, top to bottom; both are literal
  enumerations, quoted above with line numbers.
- `git log -- <overlay>` → one commit ever, at creation.
- The running pod confirms the consequence, not merely the config:
  `browser-runner-adapter:v1.0.1194` while the fleet is 1274.
- The census reproduces a known-positive row (see the misstep note).

That is direct reading of the deciding code plus the live artefact, which is
what the loop would have done here; there is no non-obvious cause to hunt.

## Verify a fix

```
grep -c "render-audit" makefile                       # expect >0 after fix 1 or 2
kubectl get pod -n ai-persona-system -l app=render-audit-adapter \
  -o jsonpath='{.items[*].spec.containers[*].image}'   # expect >= v1.0.1274
```
Then re-run `bugs_open/233`'s pod-grep against it: `access_key_present` = 1,
`B2_APPLICATION_KEY from env` = 0. That image has no `strings` — use
`grep -ac <str> /app/browser-runner-adapter`, with `NewS3Client` as the
positive control.

## Relations

- `bugs_open/233` — the credential leak this was found by; carries the
  **rotation-ordering constraint** that makes this bug time-sensitive: roll
  render-audit BEFORE rotating B2 keys, or the new key is logged in plaintext.

---

## FIX APPLIED 2026-08-10 — `render-audit-adapter` is now in the release path

Owner asked for it explicitly: add the service to `release`, `redeploy-agents`
and the steps in between. Fix candidate **2** (add the missing wiring), with the
guessing bug from candidate 1's territory closed where it actually lives.

**Four edits, all in `makefile`:**

1. **`deploy-agents`** — a re-tag-and-apply block, placed immediately after
   `browser-runner-adapter` because the two must move together: the tag applied
   here has to be one the browser-runner was actually built at.
2. **`redeploy-agents`** — a `rollout restart` line, same placement, same reason.
3. **`update-kustomization-images`** — added to the `for agent in …` list, the
   second of the two hand-maintained enumerations this bug is about.
4. **`deploy-render-audit-adapter`** — a NEW explicit target. This one is not
   bookkeeping: the `deploy-%` pattern rule pre-flights
   `docker manifest inspect $(REGISTRY)/<service>:$(IMAGE_TAG)`, which for this
   service asks for `docker.io/aqls/render-audit-adapter` — **an image that has
   never existed** — so the single-service deploy would have refused a perfectly
   valid deploy with "not in the registry". An explicit rule beats a pattern rule
   in GNU Make, and this one checks `browser-runner-adapter` instead.

**Deliberately NOT added to `build-backend` or `push-backend.`** The service has
no binary of its own; adding a build would create a second, divergent image for
one binary. The `deploy-agents` comment says so at the site, so the next reader
does not "complete" the wiring by adding one.

**A guess was removed, not just a name added.** `update-kustomization-images`'s
fallback branch appended `docker.io/aqls/$$agent` when an overlay had no
`images:` block — correct for every service that owns its image and silently
wrong for this one, which would have written a nonexistent image name and
produced an `ImagePullBackOff`. That branch now resolves the image name through
a `case`, so the list and the fallback cannot disagree. Adding the service name
while leaving the fallback guessing would have swapped a frozen tag for a
broken pull, and a comment would not have been a control on a tree this many
sessions share.

**Verified, with a negative control** (all `make -n`, no cluster calls):

| check | result |
|---|---|
| makefile parses | OK |
| `make -n release` reaches the render-audit overlay | **3** actions (sed, dir test, apply) |
| `make -n deploy-agents` applies its overlay | 3 |
| `make -n redeploy-agents` restarts it | 1 |
| `make -n update-kustomization-images` includes it | yes |
| explicit rule beats the pattern rule | `make -n deploy-render-audit-adapter` inspects **`browser-runner-adapter:v1.0.1278`**, not `render-audit-adapter:…` |
| **negative control** — `make -n build-backend` mentions it | **0** (correctly still has no build step) |

**What is still true, and why this stays open.** The pod is still on
`browser-runner-adapter:v1.0.1194`. The makefile change means the *next* release
re-tags and rolls it; it does not roll it now. Until that happens:

- the plaintext-credential leak in `bugs_open/233` is still live on that pod, and
- the **rotation-ordering constraint stands**: roll this pod to ≥ the fixed tag
  **before** rotating the B2 keys, or its next restart writes the *new* key into
  the logs in plaintext.

Close this only against a pod check, not against the makefile diff:
`kubectl get pod -n ai-persona-system -l app=render-audit-adapter -o jsonpath='{.items[*].spec.containers[*].image}'`

---

## CLOSING CONDITION MET 2026-08-10 — the pod rolled, on the first release after the fix

The condition this file set ("close only against a pod check, not the makefile
diff") is satisfied:

```
kubectl get pod -n ai-persona-system -l app=render-audit-adapter \
  -o jsonpath='{.items[*].spec.containers[*].image}'
→ docker.io/aqls/browser-runner-adapter:v1.0.1280      (was v1.0.1194)
```

Started 15:45:16Z, in the same wave as `browser-runner-adapter:v1.0.1280` — the
two now move together, which was the point. The overlay's pin was rewritten from
`v1.0.1194` to `v1.0.1280` by the release, so the `deploy-agents` sed fired: the
fix is proven by the tooling doing the work, not merely by the diff existing.

**86 tags of drift closed in one roll** (1194 → 1280). Every browser-runner
change since 2026-07-28 reached this pod for the first time.

`bugs_open/233`'s credential leak is closed at this pod as a direct consequence —
`access_key_present` 1, `B2_APPLICATION_KEY from env` 0, control `NewS3Client` 3,
and 0 credential lines in its live log buffer against 11 total.

**The CLASS remains open, which is why this file stays in `bugs_open/`.** Both
tag-update mechanisms are still hand-maintained enumerations. The next service
that shares another service's image will repeat this exactly, and it will again
be invisible to the normal proof (pod-grep the image's *owner* and it reads
live). Fix candidate **1** — enumerate the filesystem rather than a hand-list —
is what would retire it. Also still unaddressed and found by the same census:
`github-actions-runner` on **v1.0.948**, `github-actions-runner-vmsites` on
v1.0.1126.

---

## CLASS FIX IN PROGRESS 2026-08-17 — step 1 committed (`6b3524201`), the door is armed

Fix candidate **1**, taken up as its own task. Verified first: the pod is on
`browser-runner-adapter:v1.0.1305` with the fleet (individual case stays closed),
and both mechanisms were still hand-lists at HEAD.

**Live now (`6b3524201`, makefile only):**
- `RELEASE_IMAGES` / `AGENT_DEPLOY_SERVICES` (`name[:image]`) / `RETAG_EXEMPT`
  (`auth-service:deploy-auth-service`, `core-manager:deploy-core-manager`) —
  ONE declaration per set, at the top of the makefile. The image-sharing trio is
  now declared, not special-cased: this census found **vet-intel** and
  **business-intel** pin `agent-chassis` — the old fallback's guess was latently
  wrong for them too; only render-audit-adapter had a `case`.
- **`check-release-coverage`** — enumerates the FILESYSTEM: any overlay pinning a
  release-built image that is in neither list refuses the release, naming the
  service and both remedies. Prereq of `update-kustomization-images` (runs in
  `deploy-core`, so every `make release`). **Mutation-tested**: a fake
  chassis-pinning overlay FAILS it; the original 237 state (list minus
  render-audit-adapter) FAILS it naming render-audit-adapter; current tree PASSES.
- `update-kustomization-images` loops over the declaration (old list was also
  missing thunder/analyser/browser-runner — superset now, seds are idempotent).

**Council:** gate refused client-side — makefile-only touches none of
`platform/`, `internal/`, `pkg/` (owner ruling 2026-07-17; precedent: 249's
banner). Refusal drawn live, no FORCE, no credits.

**Remaining (mechanical, behaviour-equivalence provable by `make -n` set-diff;
baselines saved in the session scratchpad `237/baseline_*.txt` — 37 deploy
actions, 17 restarts, 14 pushes):**
1. `deploy-agents`: replace the ~15 copy-pasted sed+apply blocks with one loop
   over `AGENT_DEPLOY_SERVICES` (hoist the browser-runner topic note + the
   render-audit history comment); keep non-retag applies, DB sync, restarts.
2. `redeploy-agents` (+ its one extra, ollama-adapter) and `push-backend`
   become loops over the declarations.
3. `deploy-%`: preflight the image read FROM the overlay (awk `images:`→`name:`)
   instead of assuming `$(REGISTRY)/$*`; then delete the explicit
   `deploy-render-audit-adapter` rule — the command keeps working via the
   pattern, and the next image-sharing service needs no bespoke rule.
Then: LANDMINES entry update (the hand-list landmine gains its guard) +
`landmines-verify-dispatch.sh`, and a concept-register BLD entry for
`check-release-coverage`. `github-actions-runner*` drift stays a separate
owner call (own image lineage — the coverage check correctly ignores it).

### Steps 2–3 DONE 2026-08-17 (`f0657b466`) — the class fix is complete in the makefile

> **CORRECTION to the step-1 block above.** It told the next reader that behaviour
> equivalence was "provable by `make -n` set-diff" and recorded baselines as
> "37 deploy actions, 17 restarts, 14 pushes". **That instrument cannot work, and
> those figures must not be compared across this change.** `make -n` prints a recipe
> without executing it, and the whole point of the refactor is to replace fifteen
> make lines with a **shell `for` loop** — which is *one* recipe line however many
> services it iterates. The same command now reads **9**, against 37, with the sets
> **identical**. Trusting my own recipe would have read a 76% drop and concluded the
> refactor dropped 28 services from the release. `WRONG_CALLS.md` 2026-08-17.
>
> **The same defect is in this file's own 2026-08-10 verification table**, which
> records "`make -n release` reaches the render-audit overlay → 3 actions". True the
> day it was written; `bugs_open/249` landed the next day, turned `release` into
> `$(call pinned_sweep,…)` — one shell block that loops `make $$goal` internally —
> and `make -n` stopped descending into it at all. **That row now returns 0 on a tree
> where render-audit-adapter is fully wired and rolling correctly.** Do not use it.
> (BLD-020's register entry did flag the general behaviour at the time; what nobody
> updated was the recipe written down *here*, which is where a reader looks.)
> **Measure at a goal the sweep CALLS:** `make -n deploy-agents | grep -c render-audit-adapter`,
> `make -n deploy-core | grep -c "Release coverage OK"`. Now in `LANDMINES.md`.

**What shipped.** `deploy-agents`' fifteen copy-pasted sed+apply blocks, plus
`redeploy-agents` and `push-backend`, are now loops over the declarations. A declared
service with no overlay is **skipped with a named warning** instead of disappearing
into `2>/dev/null || true` — absence looking like success is the same shape as this
bug. `deploy-%`'s registry pre-flight reads the image **from the overlay**, so the
bespoke `deploy-render-audit-adapter` rule is **deleted**: a hand-written duplicate of
a pattern rule is one more pair of things that must stay identical, which is the
defect. The two per-service facts worth keeping (browser-runner's Strimzi topic lives
outside its overlay; render-audit has no image of its own and must move with
browser-runner) are hoisted into the loop header.

**Proven by SETS, with controls** [MEASURED 2026-08-17]:

| check | result |
|---|---|
| `deploy-agents` overlay set, old vs new | 15 vs 15, **identical**; order preserved (render-audit still directly after browser-runner) |
| negative control in that comparison | `auth-service` absent from both |
| `push-backend` / `redeploy-agents` sets | 14/14 and 16/16, identical |
| the loop **executed** under `kubectl`/`sed` stubs | 15 applies, 0 skips (production) |
| positive control for the new warning | development overlay path: **8 SKIPPED named, 7 applied** — the old code said nothing at all |
| `deploy-render-audit-adapter` via the pattern rule | names `browser-runner-adapter:v1.0.1305` + "runs the browser-runner-adapter image"; was: an image that never existed |
| control | `deploy-git-adapter` still names its own image |
| coverage gate, discriminating control | `AGENT_DEPLOY_SERVICES="agent-chassis"` → names `render-audit-adapter`, i.e. **reproduces this bug's original state** |
| negative control | `make -n build-backend` mentions render-audit **0** times |

Registered as **BLD-022** (`register/build-pipeline.md` + the index row).

### ADJACENT DEFECT found while proving this — filed here, NOT fixed

Four explicit rules **shadow** the `deploy-%` pattern rule and carry **no registry
pre-flight at all**: `deploy-vet-intel`, `deploy-business-intel`,
`deploy-remote-job-spawner`, `deploy-kafka-scheduler`. Found because
`make deploy-vet-intel` under a stubbed-failing `docker` proceeded to `kubectl apply`
instead of refusing — an explicit target beats a pattern rule, so the pre-flight the
pattern rule exists to provide is simply absent for those four. A single-service
deploy of any of them can still point the cluster at a never-pushed tag and
ImagePullBackOff, which is exactly what the pre-flight was added to prevent.

Same "two hand-maintained things that must stay identical" family, **different path** —
single-service deploy, not the release sweep — so it is recorded rather than swept
into this change. Fix shape: delete the four rules (the pattern rule now serves them
correctly, image-resolution included), or give each the pre-flight. Whoever takes it
should check the duplicate `deploy-admin-dashboard` definition (two identical alias
rules, `makefile:348` and `:1272`) in the same pass.

### Closing position

The **class is closed in the makefile and the gate is armed**, but it has not yet been
exercised by a real release (releases are whole-fleet and owner-run). It is make-level,
so no roll and no image is needed for it to bite — it takes effect at the next
`deploy-core`. **This file stays open** for two residuals that are genuinely still live:

1. `github-actions-runner` **v1.0.948** and `-vmsites` **v1.0.1126** are still adrift.
   The gate correctly ignores them — they pin an image the release does not build, so
   their lineage is a deliberate separate cadence — but nobody has ruled on whether
   that drift is intended. **An owner call, not a coding task.**
2. The four shadowing rules above.

Close condition: an owner ruling on (1), plus (2) fixed — **not** the makefile diff,
and not a green `make check-release-coverage`, which only says the gate agrees with
today's tree.
