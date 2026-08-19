# 237 — `render-audit-adapter` is in no release path, so it has never been rolled

> # ✅ CLOSED 2026-08-19 — fixed, LIVE and verified on `v1.0.1314`. Owner ruled to close.
>
> **The individual case AND the class.** `render-audit-adapter` was unfrozen 2026-08-10;
> the class fix (one declaration per set + `check-release-coverage`, BLD-022) shipped
> 08-17; the residual six own-lineage services were costed 08-18, ruled by the owner
> the same day, implemented (`b1480f008`) and carried live by the first release after
> the ruling.
>
> **Acceptance, at the artefact** [MEASURED 2026-08-19]: all four check CronJobs and
> both runner Deployments serve `v1.0.1314`; `github-actions-runner` and
> `github-actions-runner-vmsites` both serve the **same** image, which is the
> `<service>:<image>` declaration proven in production; the release was built from
> `d3590ca46`, of which `b1480f008` is an ancestor; and the registry census reads
> **170 / 170, empty diff**, with all four formerly-invisible actions present.
> `-vmsites` moved for the first time since 2026-07-16 and `github-actions-runner`
> for the first time since **April**, closing its missing `rsync`/`ssh` on the same roll.
>
> **What is explicitly NOT claimed:** no behavioural difference was observed, and
> there may be none to observe — see the 2026-08-19 section. An unchanged check
> report here is evidence in **neither** direction.
>
> **Carried forward, deliberately, so closing this does not lose them:**
> `bugs_open/318` (a service whose image the release does not build is still
> invisible to the gate — the door-closer for a *seventh* service, and the owner's
> stated follow-on) and a contribution into `bugs_open/153` (the release set is now
> 19 images and 5 of them carry no provenance stamp — this lane's own change moved
> that denominator).
>
> Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_237_release_enumeration/`
> (standing five; `SUMMARY_2026-08-19_live_and_verified.md` is the read-out).

**(historical status line, kept for the record) Status: FIXED IN THE MAKEFILE
2026-08-10 (owner asked for it directly). STILL OPEN until a release actually
rolls the pod off v1.0.1194** — the repo change cannot move a running pod, and
releases are whole-fleet and owner-run.

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

---

## OWNER RULING 2026-08-17 — the runners' separate cadence is **NOT intended**

Residual 1 above asked whether `github-actions-runner` (v1.0.948) and
`-vmsites` (v1.0.1126) drifting from the fleet was deliberate. **The owner's
answer: it is not.** Every doc that called it "own image lineage, own concern"
(this file's census row, `deploy-github-runners`' makefile comment, the
LANDMINES entry) recorded a *guess about intent* as if it were a design
decision. It was drift.

### The drift is not cosmetic — the two runners have DIFFERENT CAPABILITIES, live

The runner image contains **no platform Go code at all** (Ubuntu 24.04 + the
GitHub Actions runner binary + B2 CLI + one entrypoint script), so "80 tags
behind" sounds harmless. It is not. The only project-owned content in it is the
dockerfile and `github-actions-runner-entrypoint.sh`, and the dockerfile changed
on **2026-07-16** (`6880c669e`) to add **`openssh-client` and `rsync`**.

`-vmsites` was built at that commit (v1.0.1126) and has them.
`github-actions-runner` was built **2026-04-08** (v1.0.948) and does not.

Verified at the running pods, not at the tags [MEASURED 2026-08-17], with
positive controls so the absences cannot be a broken `exec`:

| pod | image | rsync | ssh | git (control) | jq (control) |
|---|---|---|---|---|---|
| `github-actions-runner` | v1.0.948 | **ABSENT** | **ABSENT** | PRESENT | PRESENT |
| `github-actions-runner-vmsites` | v1.0.1126 | PRESENT | PRESENT | PRESENT | PRESENT |

**Consequence:** any workflow on the `gqls/sites` repo runner that needs `rsync`
or `ssh` fails today, and fails looking like a workflow bug rather than a
four-month-old image. The capability was added deliberately in July and has
never reached the runner it was presumably added for.

### Two DIFFERENT mechanisms freeze them, which is why one fix will not do

- **`github-actions-runner` — has a target nobody runs.** `release-github-runner`
  (build + push + `deploy-github-runner` singular) would move it. Its overlay was
  last touched **2026-04-08**; the target has simply not been run in four months.
- **`-vmsites` — has NO retag target at all.** `release-github-runner` and
  `deploy-github-runner` name only the singular service. Its overlay was written
  once, at creation (2026-07-16), and nothing in the makefile can move it. It is
  current only by the accident of having been created on the day the content last
  changed. **This is exactly this bug's class** — a second deployment sharing one
  image, invisible to hand-listed tooling — and `check-release-coverage` does not
  catch it only because the runner image is not in `RELEASE_IMAGES`.

`deploy-github-runners` (plural) applies both manifests every release but
deliberately never seds either tag, so it moves pod-spec changes and never image
changes.

### The real design question this exposes

Rebuilding the runner on every fleet release is not obviously right: the image is
`apt-get` + a pinned upstream runner tarball, so a rebuild at an unchanged commit
still produces a *different* image (newer apt packages), and it would add minutes
to every release for content that changes about twice a year. The honest framing
is that the runners need **a path that fires when their content changes**, not
necessarily the fleet's cadence. Options are costed in the chat of 2026-08-17;
**awaiting the owner's choice** before any code change.

**Do not, in the meantime, "fix" this by running `release-github-runner`**: it
would move the singular runner to today's `IMAGE_TAG` and leave `-vmsites`
untouched and un-movable, which is the state that produced this in the first place.

---

## 2026-08-17 evening — the class fix SURVIVED ITS FIRST REAL RELEASE, and the residual is bigger than two runners

### (a) `v1.0.1307` ran the refactored tooling. Nothing broke.

The fleet rolled while this lane was open, and the release used the new
declaration-driven `deploy-agents`. Verified at the artefact [MEASURED 2026-08-17]:

- **All four of this lane's commits are IN the deployed build.** Build provenance
  (BLD-019) reports `a6d1c53c0` on browser-runner, git-adapter, core-manager,
  auth-service and reasoning-agent — **one revision across all five**, so BLD-020's
  pin held too. `git merge-base --is-ancestor` confirms `6b3524201`, `f0657b466`,
  `e24dc0e6c` and `0cc47437a` are all ancestors of that build.
- **15 services on `v1.0.1307`**, retagged by the single loop that replaced the
  fifteen copy-pasted blocks.
- **The pair still moves together:** `browser-runner-adapter` started `17:05:42Z`,
  `render-audit-adapter` `17:05:44Z` — two seconds apart, on the same image. That
  ordering is the one per-service fact the refactor had to preserve, and it did.

> **What this does NOT prove, stated so nobody upgrades it later.** The tree is
> compliant, so `check-release-coverage` **could only have passed** — this release
> could not have come out otherwise, and a gate that cannot fail proves nothing
> about its discriminating power. That was established separately, by mutation
> (the control that reproduces this bug's original state). What the release proves
> is the *other* half and it is worth having: **the refactor did not break a real
> release.** BLD-022 moves from "armed" to "exercised", not to "proven".

### (b) The real residual: SIX services are frozen, not two — and four of them run PLATFORM GO CODE

Chasing the runner ruling turned up the same shape four more times. Every service
whose image the release does not build is frozen, because each has its own deploy
target and **nothing runs it, and nothing notices**:

| service | pinned | overlay last touched | what it is |
|---|---|---|---|
| `github-actions-runner` | v1.0.948 | 2026-04-08 | CI runner; **missing rsync + ssh** (above) |
| `github-actions-runner-vmsites` | v1.0.1126 | 2026-07-16 | CI runner; **no retag target exists at all** |
| `component-render-check` | v1.0.1258 | 2026-08-06 | daily CronJob check |
| `shared-output-fields-check` | v1.0.1265 | 2026-08-08 | daily CronJob check |
| `removed-config-keys-check` | v1.0.1285 | 2026-08-11 | daily CronJob check |
| `verifier-remit-check` | v1.0.1289 | 2026-08-11 | daily CronJob check |

Each overlay was written **once, on the day the service was created, and never
again** — which is the proof that these targets are not run, rather than an
inference.

**The four checks are the serious half, and they were the quiet one.** Unlike the
runner image (Ubuntu + a pinned upstream tarball, project content changing about
twice a year), all four checks **build from the platform Go source** and import the
estate's most-churned package — `component-render-check` imports
`platform/orchestration/actions`, `verifier-remit-check` imports
`platform/orchestration/actions/discovery_checks`. Measured against HEAD:
**~2,865 / ~2,619 / ~1,778 commits of platform code behind** respectively.

So the estate's own daily immune system is running August-6th logic against a
platform that has moved 2,865 commits. **A fix to a check's own logic does not
reach the check** — it needs someone to remember a target, and for four months
nobody has. This is the same defect as `render-audit-adapter`, one layer out: the
coverage gate deliberately ignores all six, because their images are not in
`RELEASE_IMAGES`, and that scope line is now the thing to revisit.

**Not yet measured, and deliberately not asserted:** whether each check is
*functionally* wrong today. Staleness of a linked package is a strong reason to
look, not proof that behaviour changed. That is a per-service question and it is
the first thing the next session should cost.

> ### ⚠ CORRECTION 2026-08-17, same session — I overstated the checks' staleness
>
> Section (b) above says the four checks are "running August-6th logic". That is
> true of the **platform packages they link** and it is **not** true of their own
> code, and the sentence invites the second reading. Measured properly
> [MEASURED 2026-08-17]: **none of the four has a single unshipped commit to its
> own `cmd/<service>/` directory.** Each was deployed at or after the last change
> to its own logic — `verifier-remit-check`'s two same-day commits (`ef1374426`,
> `74ac4ed3a`) are both ancestors of its overlay commit `74ac4ed3a`, i.e. they
> shipped; the other three have had no own-code changes at all since deploying.
>
> **So the honest risk is narrower than "the immune system is stale".** The checks
> run the logic they were written with. What is stale is the *shared* code they
> link (`platform/orchestration/actions`,
> `platform/orchestration/actions/discovery_checks`) — so if a shared predicate,
> constant or helper they depend on has been corrected since, the correction has
> not reached them, and the check would go on testing the old rule while reporting
> cleanly. That is a real failure mode and it is the reason to fix the mechanism;
> it is not evidence that any check is wrong today.
>
> **Still unmeasured, still not asserted:** whether any linked symbol they actually
> use has changed. That is the per-service question, and it is now a much smaller
> one — the surface is "the symbols these four `cmd/` packages import", not the
> whole 2,865 commits.
>
> Caught by asking a sharper question than the commit count: "has this service's
> OWN code changed since it shipped?" A large diff between a build and HEAD says
> nothing about whether the diff touches the built thing.

---

## 2026-08-18 — the per-service costing, done. Two of the four checks ARE blind today.

The question left open above ("whether any linked symbol they actually use has
changed") is now answered, and the answer is worse than the correction implied.
It is not merely that a linked *helper* moved. **The registry these checks
enumerate is compiled into their binaries, so a frozen image is a frozen view of
what the estate contains.**

### The mechanism, in one paragraph

`cmd/config-key-audit` — which is the binary behind *both*
`shared-output-fields-check` and `removed-config-keys-check`, via different
Dockerfile CMDs — drives its whole audit from
`datahelpers.ListActionInputSpecNames()` (`cmd/config-key-audit/main.go:277`,
with `ListDeclaredConfigKeys()` at `:229` and `ListRemovedConfigKeys()` at
`:260`). That list is populated by ~169 `RegisterActionInputSpec(...)` calls
compiled into the binary. **An action that entered the registry after the image
was built is not in the list, so the loop never visits it** — `GetActionInputSpec`
returns `!ok` and the code does `continue` (`:297-300`). The check does not
report that it skipped anything. It reports clean.

### The census [MEASURED 2026-08-18]

Production registry only (`':(exclude)*_test.go'` — the first cut of this
included test-only registrations and read 201/164; the numbers below are the
corrected ones). `-A1` so multi-line calls are counted; the missing sets are
identical either way, which is the check on the counting method:

```bash
acts () { git grep -h -A1 'RegisterActionInputSpec(' "$1" -- platform/ internal/ \
          ':(exclude)*_test.go' | grep -o '"[a-z0-9_]\+"' | tr -d '"' | sort -u; }
acts HEAD > /tmp/h; acts <build-commit> > /tmp/b; comm -23 /tmp/h /tmp/b
```

| service | tag | build commit | registry in binary | actions it cannot see |
|---|---|---|---|---|
| `component-render-check` | v1.0.1258 | `667757b5d` | **160 / 169** | `ensure_page_section_layout` `evaluate_directory_features` `process_approval_decision` `process_data` `publish_site` `record_vision_finding` `retract_asset_files` `update_page_status` `zip_deliverable` |
| `shared-output-fields-check` | v1.0.1265 | `22ed9aa04` | **161 / 169** | the same, less `ensure_page_section_layout` |
| `removed-config-keys-check` | v1.0.1285 | `a9237f0c9` | **165 / 169** | `evaluate_directory_features` `publish_site` `retract_asset_files` `zip_deliverable` |
| `verifier-remit-check` | v1.0.1289 | `74ac4ed3a` | **165 / 169** | (same four — but see below, it does not read the registry) |

**This measurement could have come out otherwise** — 169 at every commit was the
expected result if the registry had been quiet, and four separate builds agreeing
would have closed the question. They do not agree.

Build commit = the commit that last touched that service's
`overlays/production/uk_001/kustomization.yaml`, which per the freeze evidence is
the day it was created and deployed.

### Per service, because they are not equally affected

- **`shared-output-fields-check` and `removed-config-keys-check` — BLIND, today.**
  Both drive off the compiled registry (proven above at `main.go:277/229/260`).
  Eight and four live actions respectively are outside their field of view, and
  the failure is silent: a skipped action produces no finding, so the check's
  clean report is indistinguishable from a clean estate. Both ran this morning
  (`06:25Z`, `07:10Z`) on the frozen images — confirmed at the cluster, not
  inferred from the repo:
  `kubectl -n ai-persona-system get cronjob -o custom-columns=...,IMAGE:...`
  prints `…/removed-config-keys-check:v1.0.1285` and `…:v1.0.1265`.
- **`component-render-check` — suspect, not proven.** It does not enumerate the
  registry; its one direct internal symbol whose definition changed since its
  build is `actions.RenderContext`. Nine actions are outside its linked registry,
  but whether its own logic depends on that is not established. Someone should
  read `rendercheck.go` against the current `RenderContext` before asserting it.
- **`verifier-remit-check` — least affected, and probably fine.** It references
  only two internal symbols and **neither definition has changed** since
  `74ac4ed3a`. It does not touch `datahelpers` at all. Its 165/169 is a property
  of a package it links, not of anything it reads. Do not fold it in on the
  strength of this table alone.

### `publish_site` and `retract_asset_files` are missing from ALL FOUR

CLAUDE.md already records that these two actions "entered the registry counted as
**ZERO** and were invisible to the check until 2026-08-17" — that is about the
`optional-key-budget-check` cron's hand-maintained literal in `check.py`, a
**different** mechanism with its own parity test
(`cmd/config-key-audit/optional_budget_cron_parity_test.go`). The same two actions
are *also* invisible to these four Go binaries, for the unrelated reason that the
images are frozen. **Two independent blind spots landing on the same pair of
actions is not a coincidence worth ignoring** — both mechanisms fail whenever the
registry grows, and neither notices. Worth checking whether the `publish_site` /
`retract_asset_files` authors were told their action is unaudited by either path.

### What this does NOT show

Nothing here says a check has emitted a *wrong* finding. The demonstrated defect
is **under-coverage** — findings that should have been raised and were not. That
is the harder kind to notice, because the artefact of a blind check and a healthy
estate is the same empty report, and it accumulates: every clean run since
2026-08-08 vouched for actions the binary could not see.

### The complete frozen set is SIX, confirmed — the other CronJobs are not this class

Enumerated every `deployments/kustomize/services/*/overlays/production/uk_001/`
overlay and listed each one not pinned to the fleet tag `v1.0.1309`. Beyond the
six already known, the rows that came back pin **no image at all** —
`optional-key-budget-check`, `single-owner-carriers-check`,
`concept-register-drift-check`, `component-fallback-check`,
`bugs-open-staleness-sweep`, `site-discovery-staleness-check`,
`instance-token-adoption-check`. Those all run `postgres:16-alpine` with their
logic in SQL/ConfigMap, so they carry no build at all and are outside this class
(they have their own staleness mechanism — the ConfigMap literal, see CLAUDE.md
on `check.py`). `ollama-adapter` / `ollama-eval` pin `latest`, which is a
*separate* trap and not part of 237. **So Decision B's scope is the six, and only
the six.**

---

## 2026-08-18 — OWNER RULING on Decision B, and it is implemented (`b1480f008`)

**Ruling: all six fold into the fleet release.** The four checks take option 2 now
with option 1 (the content-change trigger) to follow as its own change; the two
runners take option 2 as well, overriding the lane's recommendation of a one-off
unstick plus a written `RETAG_EXEMPT` entry. One cadence, no exemptions to
maintain. `RETAG_EXEMPT` gains no new entries.

**What shipped** — `RELEASE_IMAGES` gains the four check images plus
`github-actions-runner`; `AGENT_DEPLOY_SERVICES` gains all six services, with
`github-actions-runner-vmsites:github-actions-runner`.

- **`-vmsites` needed no new target.** Both runner overlays pin the *same* image
  (`docker.io/aqls/github-actions-runner`) at different tags — one image, two
  Deployments — so it fits the existing `<service>:<image>` form, the same shape
  as `render-audit-adapter:browser-runner-adapter`. This file's "**no retag target
  exists at all**" was the symptom; it never needed one, it needed *declaring*.
- **The gate's blind spot closes by construction, and this is the part worth
  keeping.** `check-release-coverage` fails a service only when its overlay pins
  an image **the release builds**. All six pinned images the release did not
  build, so no amount of tightening the gate would have caught them — the fix was
  to change which side of the predicate they sit on. Controls, both directions:
  with the **old** declarations none of the six appears in the gate's output
  however hard you probe (`RELEASE_IMAGES="<old 14>" AGENT_DEPLOY_SERVICES="agent-chassis"`);
  with the new ones the same mutation names all six.
- **⚠ NEW HAZARD, written where the old one was.** `deploy-agents` now retags both
  runner overlays to `IMAGE_TAG`. That is safe *only because* the release also
  builds and pushes that image. Remove `github-actions-runner` from
  `RELEASE_IMAGES` (or drop `build-github-runner` from `build-backend`) while
  leaving the runners in `AGENT_DEPLOY_SERVICES` and both runners
  `ImagePullBackOff` **together** — the exact trap the old makefile comment
  warned about, with its premise now inverted. **`check-release-coverage` does not
  catch this direction**: it fails a service that pins a release-built image and is
  in no release path, not one that is in a release path but whose image stopped
  being built. Verified today by set equality (`build-backend` builds ==
  `RELEASE_IMAGES`, neither side over); that equality is policed by nothing.
- `build-github-runner` switched from a bare `docker build … .` (the whole shared
  working tree as context — the pattern inverted for every other backend service
  on 2026-07-17) to `ref_build`. Drop-in: it `COPY`s one tracked file.
- `deploy-github-runners` kept, demoted to its production-strict missing-overlay
  check — the generic loop warns on a missing overlay, that target fails.
- All 21 `AGENT_DEPLOY_SERVICES` entries confirmed to have an overlay at
  `$(OVERLAY_PATH)` with **exactly one** `newTag:` line, so none is silently
  SKIPPED and the whole-file sed cannot clobber a second image.

**Council: none, and deliberately.** Makefile-only submissions are refused
client-side by scope (`platform/`, `internal/`, `pkg/`), so no commit here carries
a review trailer.

### ⚠ THIS IS STILL INERT — the two blind checks are STILL BLIND

Nothing has been built, pushed or rolled. Releases are whole-fleet and the owner
runs `make release`. Until one runs, the frozen tags remain live and the registry
gap is unchanged.

**Acceptance test after the first release that includes this** — and it can fail,
which is the point:

```bash
acts () { git grep -h -A1 'RegisterActionInputSpec(' "$1" -- platform/ internal/ \
          ':(exclude)*_test.go' | grep -o '"[a-z0-9_]\+"' | tr -d '"' | sort -u; }
acts HEAD > /tmp/h; acts <new build commit> > /tmp/b; comm -23 /tmp/h /tmp/b   # expect EMPTY
kubectl -n ai-persona-system get cronjob -o custom-columns='NAME:.metadata.name,IMAGE:.spec.jobTemplate.spec.template.spec.containers[0].image,LAST:.status.lastScheduleTime'
```

Expect 169/169 on all four and the new tag on all four CronJobs. **A green
`check-release-coverage` is not the acceptance test** — it only says the gate
agrees with today's tree.

### What remains before 237 can close

1. The first release carrying this, then the acceptance test above.
2. Option 1 — the content-change trigger (`pinned tag predates the last commit to
   its own sources`), as its own change with its own review. This is what covers a
   *seventh* service without anyone remembering.
3. Open, non-blocking: read `rendercheck.go` against the current
   `actions.RenderContext` and record whether `component-render-check` was ever
   functionally affected. Suspect, unproven, and deliberately not asserted.

---

## 2026-08-19 — IT IS LIVE, and the acceptance test PASSES

`v1.0.1314` carried the fold. All six moved on the first release after the ruling,
with no manual step.

### What is proven [MEASURED 2026-08-19]

1. **All six serve the new tag, read at the cluster.** Four CronJobs on
   `…/component-render-check:v1.0.1314`, `…/shared-output-fields-check:v1.0.1314`,
   `…/removed-config-keys-check:v1.0.1314`, `…/verifier-remit-check:v1.0.1314`,
   all four having run this morning (06:25/06:55/07:10/07:25Z). Both runner
   Deployments on `v1.0.1314`.
2. **The `<service>:<image>` mapping worked in production.** `github-actions-runner`
   AND `github-actions-runner-vmsites` both serve
   `docker.io/aqls/github-actions-runner:v1.0.1314` — one image, two Deployments,
   which is what the declaration says and what the cluster now shows. `-vmsites`
   has moved for the first time since 2026-07-16, and `github-actions-runner` for
   the first time since **April** (it was on v1.0.948), so the missing `rsync`/`ssh`
   gap is closed by the same roll.
3. **The release genuinely carries the fold.** Chassis provenance stamp
   `d3590ca46` (2026-08-18 22:17), and `git merge-base --is-ancestor b1480f008
   d3590ca46` → true.
4. **The registry acceptance test passes: 170 / 170, empty diff.** All four
   actions that were invisible on 08-18 — `evaluate_directory_features`,
   `publish_site`, `retract_asset_files`, `zip_deliverable` — are present in the
   source the release was built from. (The registry itself grew 169 → 170
   overnight, which is exactly the churn that caused the freeze, and it is now
   covered.)

### What is NOT proven, stated plainly

- **No behavioural difference was observed, and there may be none to observe.**
  `removed-config-keys-check`'s `doc_notes` row is byte-identical across the
  frozen (08-18, v1.0.1285) and unfrozen (08-19, v1.0.1314) runs in the field that
  would show it: `keys declared removed:` lists the same four keys both days.
  That is **expected, not a failure** — none of the four newly-visible actions
  declares a removed config key, so this check's output cannot discriminate here.
  The field that did move (`live agent definitions walked: 189 → 191`) reads the
  live DB, not the compiled registry, so it is not evidence either way.
  **Do not cite the identical row as evidence the fix did nothing.**
- **The check pods' logs are gone** (node GC), so the direct before/after at the
  log was unavailable — the completed pods survive, `kubectl logs` on them does not.

### ⚠ RESIDUAL FOUND WHILE VERIFYING — the check images carry NO provenance stamp

`build/docker/backend/*-check.dockerfile` build with a plain
`RUN CGO_ENABLED=0 GOOS=linux go build -o <bin> ./cmd/<pkg>` — **no `buildinfo`
ldflags**. So BLD-019's "ask the binary what commit built it" does **not** work for
these four, nor for `github-actions-runner`. Proving one of them moved therefore
falls back to the image tag plus the release's ancestry, which is exactly the
weaker, inference-shaped proof BLD-019 exists to replace. Now that they are
release images this is worth closing; it is small (add the ldflags to five
dockerfiles) and it is **not** part of 237's class. File it as its own item.
