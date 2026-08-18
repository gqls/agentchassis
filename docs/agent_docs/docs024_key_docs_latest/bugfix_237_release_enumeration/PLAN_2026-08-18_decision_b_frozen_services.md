# PLAN — Decision B: the six services no release can move · 2026-08-18

Design, phasing and the reasons. Corrections to the originating brief live here,
marked as corrections.

---

## 1. The problem, stated precisely

`check-release-coverage` (shipped `6b3524201`) enumerates the filesystem and fails
the release when **an overlay pins an image the release builds** but the service is
in no release path. That predicate closed the `render-audit-adapter` case and it is
the right predicate for that case.

It is also, by construction, blind to a service whose image the release **does not
build**. Six such services exist. Each has its own `deploy-*` target, nothing runs
it, and nothing notices — a second freeze mechanism, one layer out from the one the
lane already fixed.

| service | pinned | overlay written | mechanism |
|---|---|---|---|
| `github-actions-runner` | v1.0.948 | 2026-04-08 | target exists, nobody runs it |
| `github-actions-runner-vmsites` | v1.0.1126 | 2026-07-16 | **no retag target exists at all** |
| `component-render-check` | v1.0.1258 | 2026-08-06 | target exists, nobody runs it |
| `shared-output-fields-check` | v1.0.1265 | 2026-08-08 | ” |
| `removed-config-keys-check` | v1.0.1285 | 2026-08-11 | ” |
| `verifier-remit-check` | v1.0.1289 | 2026-08-11 | ” |

Each overlay was written once, on the day its service was created, and never again.
That is the proof the targets are not run — not an inference. Re-verified
2026-08-18: unchanged across the v1.0.1309 roll.

## 2. Cost of the freeze, per group [MEASURED 2026-08-18]

Full evidence in `bugs_open/237` and `NOTES_release_enumeration.md`.

**The four checks — the damage is under-coverage, and two are provably blind.**
`cmd/config-key-audit` drives its audit from `ListActionInputSpecNames()`, a list of
~169 actions **compiled into the binary**. An action registered after the image was
built is never visited (`main.go:297-300`, `continue`, no output). So:

- `removed-config-keys-check` sees 165/169 actions; `shared-output-fields-check`
  161/169. Both ran 2026-08-18 on the frozen tags, confirmed at the cluster.
- `component-render-check` (160/169) does not read the registry; one symbol it does
  use, `actions.RenderContext`, has changed. Suspect, unproven.
- `verifier-remit-check` (165/169) reads neither the registry nor any symbol whose
  definition has changed. **Least affected — do not fold it in on the table alone.**

`publish_site` and `retract_asset_files` are missing from all four.

> **CORRECTION to the 2026-08-17 sizing.** That correction said the honest risk was
> "a shared predicate or constant they link has moved". True, and it understated the
> failure mode: what is frozen is not only the *logic* but the *inventory the logic
> iterates*. A stale linked helper might still compute the right answer; a stale
> registry cannot, because the check never asks the question. Keeping both readings
> on record because the 08-17 correction was itself correcting an overstatement, and
> the trajectory matters more than either figure.

**The two runners — the damage is a live capability gap, not staleness.**
`v1.0.948` is missing `rsync` and `ssh`, measured at the pod. That image is Ubuntu
plus a pinned upstream tarball, with project content changing about twice a year, so
rebuilding it every release churns the image (newer apt packages) for no benefit.
Different problem, different remedy.

## 3. The three options, costed

**Option 1 — trigger on content change.** Extend the gate to fail when an overlay's
pinned tag predates the last commit to the sources that service is built from.

- Encodes the real rule: *stale iff its source moved*. Catches both groups.
- Needs a per-service source→overlay map. That map is already derivable — the
  Dockerfile names the binary, `go list -deps` gives the linked packages — but
  deriving it inside the makefile is the most work of the three.
- Sharpest available predicate, and the one that survives someone adding a seventh
  service.

**Option 2 — fold into the fleet release.** Add them to `RELEASE_IMAGES` /
`AGENT_DEPLOY_SERVICES`, so every release builds, pushes and retags them.

- One cadence for everything; simplest to reason about; nothing new to maintain.
- Costs build minutes per release, and for the runners churns an image with no
  reason to change.
- Ends the *class* for anything folded in — a folded service is covered by the gate
  that already exists, with no new mechanism.

**Option 3 — unstick now, decide later.** Build, push, repoint the overlays once.

- Fixes the live `rsync`/`ssh` gap today. Leaves both freeze mechanisms in place,
  so it recurs, and the next recurrence is silent again.
- Only defensible as a *first step under* 1 or 2, never as the answer.

## 3a. OWNER RULING 2026-08-18 — all six fold into the fleet release

- **The four checks → option 2, then option 1.** Fold in now so the blindness ends
  at the next release; build the content-change trigger afterwards as its own
  change, so a *seventh* check is covered without anyone remembering.
- **The two runners → option 2 as well.** This **overrides the recommendation in
  §4 below**, which argued for a one-off unstick plus a written exemption on the
  grounds that rebuilding an Ubuntu-plus-tarball image every release churns it for
  nothing. The owner has ruled one cadence for everything. Recording the
  disagreement rather than quietly rewriting §4: the churn cost is real and was
  weighed, and a single cadence with no exemptions to maintain is the simpler
  estate. **`RETAG_EXEMPT` therefore gains no new entries.**

**Implemented the same day, `b1480f008`.** What the ruling turned out to need, and
one thing it did not:

- `-vmsites` needed **no new target at all**. Both runner overlays pin the *same*
  image (`docker.io/aqls/github-actions-runner`) at different tags — one image,
  two Deployments — so it fits the existing `<service>:<image>` form exactly, the
  same shape as `render-audit-adapter:browser-runner-adapter`. The handoff's
  "**no retag target exists at all**" was the symptom; the cause is that it never
  needed one, it needed *declaring*.
- **The build half and the retag half are one change.** `deploy-agents` now
  retags both runner overlays to `IMAGE_TAG`; if the image were not also built and
  pushed, both runners would `ImagePullBackOff` together — precisely the trap the
  old makefile comment warned about. So `build-backend` gained `build-checks` and
  `build-github-runner`, verified by **set equality** (what `build-backend` builds
  == `RELEASE_IMAGES`, neither side over). That equality is not policed by any
  gate and is the thing to re-check if anyone edits either list.
- `build-github-runner` was still a bare `docker build … .` — the whole shared
  working tree as build context, the pattern inverted for every other backend
  service on 2026-07-17. Switched to `ref_build`; it only `COPY`s one tracked
  file, so it was a drop-in. A release image built from the working tree would
  have reintroduced exactly that blast radius.
- **The gate's blind spot closes by construction.** `check-release-coverage` only
  polices overlays pinning a `RELEASE_IMAGES` image, so it could not see any of
  these six. Controls run both ways: with the *old* declarations none of the six
  appears however hard you probe; with the new ones a mutated
  `AGENT_DEPLOY_SERVICES` names all six.

**Still inert.** Nothing has been built or rolled — releases are whole-fleet and
the owner runs `make release`. The four checks stay blind until then.

## 4. Recommendation (superseded for the runners — see §3a)

**Split the answer.**

**Four checks → option 2, plus option 1 as the backstop.** Option 2 alone fixes
today and is nearly free to implement; option 1 alone leaves them frozen until
someone acts on a new failure. Together: folding them in means the registry can
never be more than one release behind, and the content-change trigger means a
*seventh* check added next month is covered without anyone remembering. The order
matters — fold in first (fast, fixes the live blindness), build the trigger second.

**Two runners → option 3 once, then an explicit, machine-readable exemption.** The
separate cadence is defensible for this image on its merits; what is not defensible
is that the exemption is currently expressed as *absence* — no entry anywhere,
indistinguishable from an oversight, which is how it reached v1.0.948. Put them in
`RETAG_EXEMPT` (or the equivalent) with the reason inline, so the gate can see a
decision was made. And build `-vmsites` a retag target: a service with no target is
not exempt, it is unreachable.

Rationale for the asymmetry: the checks' cost of being stale is silent
under-coverage of the estate's own immune system, which compounds daily. The
runners' cost is one known capability gap that a single rebuild closes. Same shape,
very different blast radii.

## 5. Phasing, once ruled

1. Build and push the four checks at HEAD; repoint the four overlays; confirm at the
   cluster that the next scheduled run uses the new tag (§6 of the RUNBOOK).
2. Re-run the registry census against the new build commits — expect 169/169 on all
   four. **This is the acceptance test**, and it can fail, which is the point.
3. Add the four to `RELEASE_IMAGES` / `AGENT_DEPLOY_SERVICES`; verify at
   `deploy-agents` / `deploy-core`, never at `release`.
4. Runners: build/push once, repoint, add the exemption entry with its reason, add
   the missing `-vmsites` retag target.
5. Option 1's trigger, as its own change with its own review.
6. Read `rendercheck.go` against the current `actions.RenderContext` and record
   whether `component-render-check` was functionally affected. Open question, not a
   blocker.
7. Close `bugs_open/237`.

## 6. Constraints that shape all of this

- **Council refuses makefile-only submissions client-side** (scope: `platform/`,
  `internal/`, `pkg/`). Steps 1–4 draw no review; step 5, if it touches Go, does.
- **A gate that only agrees with today's tree is not evidence.** A green
  `check-release-coverage` on a compliant tree could only have passed. Discriminating
  power comes from mutation controls — keep using them.
- **Do NOT run `release-github-runner` as an interim.** It moves the singular runner
  only and leaves `-vmsites` untouched and unmovable: the exact state that produced
  this bug.
