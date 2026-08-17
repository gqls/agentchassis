# HANDOFF — `bugs_open/237` release-enumeration class fix · 2026-08-17

**Cold start: read this file, then `bugs_open/237_HANDOFF_2026-08-09_render_audit_adapter_is_in_no_release_path_so_it_has_never_been_rolled.md` from the bottom up.**
The bug file is the shared account and is authoritative; this file is the lane's
own state and the two open owner decisions.

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

### Decision A — the four shadowing rules (no research outstanding; just needs a yes)

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

**The four checks are the serious half and were the quiet one.** They build from
the platform Go source and import the estate's most-churned package
(`platform/orchestration/actions`). So the estate's own daily immune system is
running August-6th logic against a platform 2,865 commits further on. A fix to a
check's logic does not reach the check.

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
says the gate agrees with today's tree. It closes on: Decision A implemented, and
Decision B ruled and implemented for all six services.
