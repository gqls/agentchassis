# HANDOFF — `bugs_closed/237` release-enumeration class fix · 2026-08-17

> # ✅ THIS LANE IS CLOSED — 2026-08-19. Nothing below is live work.
>
> `237` is **fixed, LIVE and verified** on `v1.0.1314` and has moved to
> `bugs_closed/`. Both owner decisions are ruled and implemented; the acceptance
> test (registry census **170/170**, empty diff) passed at the artefact.
>
> **If you are picking something up, it is one of these two — not this file:**
> - **`bugs_open/318`** — a service whose image the release does not build is still
>   invisible to `check-release-coverage`. The owner's follow-on (the content-change
>   trigger) plus the unpoliced `build-backend == RELEASE_IMAGES` invariant.
> - **`bugs_open/153`**, contribution 2026-08-19 — the release set is now 19 images
>   and the 5 this lane added carry no `buildinfo` stamp.
>
> **Read-out for the owner:** `SUMMARY_2026-08-19_live_and_verified.md`.
> **Commands + every trap:** `RUNBOOK_release_enumeration.md`.

**Cold start: read this file, then `bugs_open/237_HANDOFF_2026-08-09_render_audit_adapter_is_in_no_release_path_so_it_has_never_been_rolled.md` from the bottom up.**
The bug file is the shared account and is authoritative; this file is the lane's
own state and the two open owner decisions.

> **The lane's standing five now exist (2026-08-18)** — read these in preference to
> re-deriving anything here: `PLAN_2026-08-18_decision_b_frozen_services.md` (the
> options, costed, with the recommendation), `RUNBOOK_release_enumeration.md` (every
> command, with its gotcha), `NOTES_release_enumeration.md` (the measurement log and
> this session's four missteps), `README_where_we_are.md` (owner's plain-prose log),
> `SUMMARY_2026-08-18_decision_b_costed.md` (the milestone read-out).
> **UPDATED 2026-08-18 — BOTH decisions are now RULED and IMPLEMENTED.** Decision B
> was ruled by the owner the same day: **all six fold into the fleet release**
> (`b1480f008`), the four checks with the content-change trigger to follow, the two
> runners folded in too rather than exempted. **`-vmsites` needed no new target** —
> both runners pin the same image, so it fits the existing `<service>:<image>` form.
> **The work is INERT until a release runs** (whole-fleet, owner runs it), so the two
> blind checks are still blind. Acceptance test, remaining work and the ONE NEW HAZARD
> (removing the runner image from `RELEASE_IMAGES` while leaving the runners in
> `AGENT_DEPLOY_SERVICES` ImagePullBackOffs both — the gate does not catch that
> direction) are in `bugs_open/237`, 2026-08-18 section. **Nothing below is a live
> decision any more; read it as the record.**

---

## 1. What this lane is about, in plain terms

The release tooling used to hold **four separate hand-written lists** of "the
services a release deploys". Nothing checked those lists against what was
actually on disk. So a service could exist, be running in production, and be in
none of the lists — in which case **no release would ever move it**, silently,
for ever. That is what happened to `render-audit-adapter`: it sat 86 image tags
behind the fleet for months, still serving a binary with a known credential leak,
and it was invisible because the normal way of checking ("is the image live?")
looks at the *other* service that owns that image, which was perfectly up to date.

The individual case was fixed on 2026-08-10. This lane fixed **the class**.

## 2. What is DONE and live

All committed and in the deployed build `a6d1c53c0` (v1.0.1307).

| commit | what |
|---|---|
| `6b3524201` | ONE declaration per set (`RELEASE_IMAGES`, `AGENT_DEPLOY_SERVICES` as `<service>[:<image>]`, `RETAG_EXEMPT`) + **`check-release-coverage`** |
| `f0657b466` | the four hand-lists become loops; `deploy-%` resolves the image from the overlay; bespoke `deploy-render-audit-adapter` deleted |
| `e24dc0e6c` | makefile comment correction — the runners' tags are NOT "on purpose" (owner ruling) |
| `0cc47437a`, `43ef000ce` | LANDMINES entries + corrections |
| `4f6485ada`, `61e057a67`, `49e320ff7` | concept register **BLD-022** + index row + status |
| `b1da82a1c` | `WRONG_CALLS.md` — the broken verification recipe (see §5) |
| `843122db0`, `0748b2af5`, `10a04294a`, `6c2afbaa9` | the bug file's running account |

**`check-release-coverage`** enumerates the *filesystem* — every
`deployments/kustomize/services/*/overlays/$(OVERLAY_PATH)/kustomization.yaml` —
and fails the release, naming the service, when an overlay pins an image the
release builds but the service is in no release path. It is a prerequisite of
`update-kustomization-images`, which `deploy-core` calls, which is in the release
sweep, and a non-zero exit propagates (`pinned_sweep` uses `|| exit 1`).

**Exercised by `v1.0.1307` without incident** — 15 services retagged by the one
loop, one revision across the fleet, and `browser-runner-adapter` /
`render-audit-adapter` started two seconds apart (the one ordering fact the
refactor had to preserve). **But the tree was compliant, so the gate could only
have passed.** Its discriminating power comes from the mutation controls, not
from this release. Do not upgrade "exercised" to "proven".

## 3. THE TWO OPEN OWNER DECISIONS

### ~~Decision A — the four shadowing rules~~ → **DONE 2026-08-17 (`4da4c6324`)**

> Implemented as recommended. All four rules and the duplicate `deploy-admin-dashboard`
> are deleted; the pattern rule serves them, now with the pre-flight. Verified: all four
> hit the pre-flight, `vet-intel`/`business-intel` correctly resolve **agent-chassis**
> (not the never-existent `docker.io/aqls/vet-intel`), positive control confirms a valid
> deploy still proceeds, and `quick-scheduler-update` / `admin-dashboard` still resolve.
> **Only Decision B is open.**

#### (original statement of the problem, kept for the record)

The makefile has one generic deploy rule, `deploy-%`, whose valuable part is a
**registry pre-flight**: before touching the cluster it asks whether the image
you are about to deploy was actually built and pushed. Without that, deploying a
tag nobody built leaves the pods in `ImagePullBackOff`.

In GNU Make an **explicitly named rule always beats a wildcard rule**. Four
services are named explicitly — `deploy-vet-intel`, `deploy-business-intel`,
`deploy-remote-job-spawner`, `deploy-kafka-scheduler` — by older three-line rules
that retag and apply and **never ask the registry anything**. So for those four,
the safety check does not exist. Found by running `make deploy-vet-intel` with a
stubbed-failing `docker`: it proceeded to `kubectl apply` instead of refusing.

**Recommendation: delete all four.** Since `f0657b466` the generic rule resolves
the image from the overlay, so it now handles these four correctly, including the
two that run the `agent-chassis` image. Keeping hand-written copies of what the
generic rule already does is precisely this bug's defect — two things that must
stay identical with nothing keeping them identical. Also delete the duplicate
`deploy-admin-dashboard` (defined twice, identically, at ~`:348` and ~`:1272`).

Low risk: removes a way to make a mistake, takes no capability away.

### Decision B — the six frozen services (this is the big one)

**Owner has ruled the runners' separate cadence NOT intended.** Chasing that
turned up the same shape four more times. Every service whose image the release
does not build is frozen, because each has its own deploy target and nothing runs
it and nothing notices:

| service | pinned | overlay last touched | note |
|---|---|---|---|
| `github-actions-runner` | v1.0.948 | 2026-04-08 | **missing `rsync` + `ssh`**, measured at the pod |
| `github-actions-runner-vmsites` | v1.0.1126 | 2026-07-16 | **no retag target exists at all** |
| `component-render-check` | v1.0.1258 | 2026-08-06 | ~2,865 commits of platform code behind |
| `shared-output-fields-check` | v1.0.1265 | 2026-08-08 | ~2,619 behind |
| `removed-config-keys-check` | v1.0.1285 | 2026-08-11 | ~1,778 behind |
| `verifier-remit-check` | v1.0.1289 | 2026-08-11 | ~1,778 behind |

Each overlay was written **once, on the day the service was created, and never
again** — that is the proof the targets are not run, not an inference.

**The four checks are the quiet half.** They build from the platform Go source and
import the estate's most-churned packages (`platform/orchestration/actions`,
`.../discovery_checks`).

> **Sized correctly, after a sharper measurement.** ~~the immune system is running
> August-6th logic~~ — that overstates it. **None of the four has an unshipped
> commit to its own `cmd/` directory**: each was deployed at or after its own last
> code change. They run the logic they were written with. What is stale is the
> **shared code they link** — so a shared predicate or constant corrected since
> their build has not reached them, and the check would go on testing the old rule
> while reporting cleanly. Real failure mode, narrower than it first looked, and
> the surface to measure is "the symbols these four packages import", not 2,865
> commits. A large diff between a build and HEAD says nothing about whether the
> diff touches the built thing.

The runners are different in kind: that image is Ubuntu + a pinned upstream
tarball, project content changing about twice a year, so rebuilding it every
release churns the image (newer apt packages) for nothing.

**Options, and they can differ per group:**

1. **Fire when the content changes.** Extend the coverage gate to fail when an
   overlay's pinned tag predates the last commit to that service's own sources.
   Encodes the real rule — "stale iff its source moved" — and would have caught
   both groups. Most work; closes the door properly. *Recommended for the checks.*
2. **Fold into the fleet release.** Add them to `RELEASE_IMAGES` /
   `AGENT_DEPLOY_SERVICES` so every release builds, pushes and retags them. One
   cadence for everything, simplest to reason about; costs build minutes and, for
   the runners, churns an image that had no reason to change.
3. **Unstick now, decide the mechanism later.** Build and push once, repoint the
   overlays. Fixes the live rsync/ssh gap today, leaves both freeze mechanisms in
   place, so it recurs.

**Do NOT meanwhile just run `release-github-runner`:** it moves the singular
runner only and leaves `-vmsites` untouched and unmovable — the exact state that
produced this.

> **UPDATED 2026-08-18 — the costing is DONE, and two of the four checks are
> provably blind today.** The paragraph below asked the next session to cost this
> per service; that is complete, so do not repeat it. Result: `cmd/config-key-audit`
> drives its audit from a registry of ~169 actions **compiled into the binary**
> (`main.go:277/229/260`; an unknown action is `continue`d at `:297-300` with no
> output), so a frozen image has a frozen **inventory**, not just frozen logic.
> `removed-config-keys-check` sees **165/169** actions, `shared-output-fields-check`
> **161/169** — and both ran on 2026-08-18 on the frozen tags, confirmed at the
> cluster. `component-render-check` (160/169) does not read the registry but one
> symbol it uses, `actions.RenderContext`, has changed — suspect, unproven.
> `verifier-remit-check` is the **least** affected: it reads neither the registry nor
> any symbol whose definition moved, so **do not fold it in on its 165/169 alone**.
> `publish_site` and `retract_asset_files` are absent from all four.
> The frozen set is confirmed to be exactly **six** — every other suspicious
> CronJob runs `postgres:16-alpine` and carries no build of ours.
> Full evidence: `bugs_open/237` (2026-08-18 section) and `NOTES_release_enumeration.md`.
> Options costed and recommended: `PLAN_2026-08-18_decision_b_frozen_services.md`.
> **What is open is the owner's ruling, not a measurement.**

#### (original statement, kept for the record)

**Not yet measured, deliberately not asserted:** whether any check is
*functionally* wrong today. A stale linked package is a reason to look, not proof
behaviour changed. Costing that per service is the first task for whoever
takes Decision B.

## 4. How to verify anything here

```bash
make check-release-coverage                                    # green on a compliant tree
make check-release-coverage AGENT_DEPLOY_SERVICES="agent-chassis"   # MUST name render-audit-adapter
make -n deploy-agents  | grep -c render-audit-adapter          # the retag loop
make -n deploy-core    | grep -c "Release coverage OK"         # the gate, via update-kustomization-images
kubectl -n ai-persona-system logs -l app=<svc> --tail=3000 | grep -m1 'build provenance'
```

## 5. ⚠ TRAPS THIS LANE PAID FOR — read before verifying anything

- **`make -n release` descends into NOTHING.** Since `bugs_open/249` made
  `release` a `pinned_sweep` shell block, `make -n release | grep <service>`
  returns **0 whether or not the service is in the release**. 237's own
  2026-08-10 verification table records this recipe as *passed*; it now always
  says "absent". **Measure at `deploy-agents` / `deploy-core`, never at
  `release`.** In `LANDMINES.md`.
- **`make -n` does not expand a shell `for` loop either.** Counting action lines
  reads **9** where the old copy-pasted blocks read **37**, with the sets
  identical. Compare the **resolved service set** (expand the make variable, or
  run the loop under `kubectl`/`sed` stubs), never the line count. I wrote the
  bad recipe into the bug file myself and hit it twelve minutes later —
  `WRONG_CALLS.md` 2026-08-17.
- **The landmine verifier cannot grade makefile-footprinted entries** — its index
  is Go-only, so it returns `UNVERIFIABLE`. That is not doubt; write a hand-run
  check into the entry.
- **Council refuses makefile-only submissions client-side** (owner ruling
  2026-07-17: scope is `platform/`, `internal/`, `pkg/`). Refusal drawn live this
  lane, no credits, no `FORCE`, and no commit carries a trailer claiming review.

## 6. Close condition for `bugs_open/237`

Not the makefile diff, and **not** a green `check-release-coverage` — that only
says the gate agrees with today's tree. It closes on: ~~Decision A implemented~~ (**done**, `4da4c6324`), and
Decision B ruled and implemented for all six services. **Decision B is the only
thing between this bug and closure.**
