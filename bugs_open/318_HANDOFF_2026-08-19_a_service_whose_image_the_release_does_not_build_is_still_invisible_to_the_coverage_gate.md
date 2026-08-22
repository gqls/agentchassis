# BUG 318 — a service whose image the release does not build is still invisible to the coverage gate

**Filed 2026-08-19 by the `bugs_open/237` lane, on the owner's ruling of 2026-08-18.**
This is `237`'s **follow-on**, not a re-opening of it: 237's six frozen services are
fixed, live and verified, and 237 is closed. This file is the door-closer that stops
a *seventh*.

> **Status: OPEN, not started.** Nothing here is built. The six known cases are
> already fixed by other means, so nothing is on fire — what is missing is the
> mechanism that would catch the next one without anyone remembering.

## The defect, in one paragraph

`check-release-coverage` (BLD-022) fails a release when an overlay pins an image
**the release builds** but the service is in no release path. That predicate closed
`render-audit-adapter`'s case and it is the right predicate for that case. It is
also, by construction, blind to the opposite shape: a service whose image the
release **does not build** is outside the predicate entirely, so the gate ignores it
however long it sits frozen. Six services were in exactly that state — two GitHub
runners frozen since April and July, and four daily check CronJobs — and no amount
of tightening the gate would have found them. They were fixed by moving them to the
other side of the predicate (`b1480f008`), which fixes those six and does nothing
about the seventh.

## Why it is worth building rather than relying on the habit

Each of the six overlays was written **once, on the day its service was created, and
never again**. That is not an inference from commit counts — it is the file history,
and it is what "nobody runs this target" looks like. The failure is silent by
construction: a service with its own deploy target and no caller emits nothing, and
the natural check ("is the image live?") reads healthy because it looks at the pod,
which is running perfectly — just old.

The cost when it bites is not uniform. Two of the four checks compile the estate's
action registry **into the binary**, so a frozen image is a frozen *inventory*: on
2026-08-18 one could see 165 of 169 registered actions and the other 161, and a
skipped action produces no output at all — a blind check and a healthy estate file
the identical empty report. That damage accumulates unnoticed for as long as the
freeze lasts (in that case, since 2026-08-08).

## The fix candidates, ordered by what closes the door

1. **Trigger on content change (the owner's stated intent, 2026-08-18).** Fail the
   release when a service's pinned tag predates the last commit to the sources that
   service is built from. This encodes the real rule — *stale iff its source moved* —
   and it is the only candidate that covers a service the release does not build,
   which is the whole gap. Needs a service→sources map; both halves are already
   derivable (`build/docker/backend/<svc>.dockerfile` names the binary, `go list
   -deps` gives its linked packages), so the work is in the plumbing, not the
   discovery. **⚠ Deriving the map inside the makefile is likely to be the wrong
   shape**, and it decides whether this is reviewable at all: a makefile-only change
   is refused by the council gate on scope (`platform/`/`internal/`/`pkg/`), whereas
   a small Go helper invoked by the makefile is in scope. Decide the shape before
   assuming a review is available.
2. **Assert the release set is self-consistent.** `build-backend` must build exactly
   `RELEASE_IMAGES`. This is currently true, was verified by set equality on
   2026-08-18, and is policed by **nothing** — see the hazard below. Cheap, narrow,
   and does not close the main gap; worth doing anyway.
3. **Periodically census pinned tags against the fleet tag.** A daily job listing
   every overlay not on `IMAGE_TAG`. Detects rather than prevents, and this estate
   already has evidence that a detector with no consumer goes unread — but it is the
   cheapest thing that would have surfaced all six.

## ⚠ The hazard candidate 2 addresses, which is live today

`deploy-agents` now retags **both** runner overlays to `IMAGE_TAG`, and that is safe
*only because* the release also builds and pushes that image. Remove
`github-actions-runner` from `RELEASE_IMAGES`, or drop `build-github-runner` from
`build-backend`, while leaving the runners in `AGENT_DEPLOY_SERVICES`, and both
runners point at an image nobody pushed and `ImagePullBackOff` **together**, taking
CI with them. `check-release-coverage` does **not** catch this direction: it fails a
service that pins a release-built image and is in no release path, never one that is
in a release path but whose image stopped being built.

## How to verify a fix

- The instrument must be shown able to fail. Construct a service whose pinned tag
  predates a commit to its own sources and require the gate to name it; and require
  a compliant tree to stay green. A green gate on a compliant tree proves nothing —
  it could only have passed.
- **Measure at `deploy-agents` / `deploy-core`, never at `release`.** Since
  `bugs_open/249` made `release` a `pinned_sweep` shell block, `make -n release |
  grep <service>` returns 0 whether or not the service is in the release.
- **`make -n` does not expand a shell `for` loop.** Compare the resolved service
  set, never the line count.

## Sources

- `bugs_closed/237_HANDOFF_2026-08-09_render_audit_adapter_is_in_no_release_path_so_it_has_never_been_rolled.md`
  — the parent case, its measurements and the owner rulings.
- `docs/agent_docs/docs024_key_docs_latest/bugfix_237_release_enumeration/` — the
  lane's standing five; `PLAN_2026-08-18_decision_b_frozen_services.md` costs the
  three options, `RUNBOOK_release_enumeration.md` holds every command with its trap.
- Concept register **BLD-022** (`check-release-coverage`) — its "what it
  deliberately does NOT catch" list is where this gap is recorded.
- `LANDMINES.md` — *"The estate's config checks carry a registry COMPILED INTO the
  binary…"* (why the four checks' freeze was expensive) and the two `make -n`
  entries above.
- `bugs_open/153` — the provenance stamp, and the related contribution that 5 of the
  now-19 release images are unstamped.

---

## CONTRIBUTION 2026-08-22 — candidate 2's hazard has FIRED, the next release was broken, and candidate 1 as worded cannot be built

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_318_release_source_coverage/`
(standing five, opened today). This bug is now **OWNED** by that lane.

`090` was not run and this says why, per the owner ruling of 2026-07-31: nothing
below is a new root cause. 318's cause was established by the `237` lane and is
recorded as BLD-022's stated blind spot. Everything here is a **mechanical
measurement with a control** — a set difference from `make -n`, a `docker images`
census with two positive controls, a `kubectl get cronjob` read. Each could have come
out otherwise, and one of them came out **against** the design this session expected
to recommend (finding 3 below). That is the declared substitute for the loop.

### 1. The bug is still valid, and the gate is green while blind

`make check-release-coverage` → *"Release coverage OK"*, exit 0, on a tree where the
class is live. Unchanged predicate: it `continue`s past any overlay whose pinned image
is not already in `RELEASE_IMAGES` (`makefile:150`), so membership of the list is the
gate's own admission criterion.

Two more services fell in since this bug was filed three days ago —
`optional-explicit-wires-check` (born 2026-08-21) and `commit-sha-exposure-check`
(born 2026-08-22) — both born outside both lists, both invisible, both fixed
one-at-a-time by `67201d125` under the standing 08-18 ruling. **Eight instances
now.** The makefile's current remedy is a comment reading *"A NEW CHECK SERVICE MUST
BE ADDED HERE IN THE COMMIT THAT CREATES IT"*, and CLAUDE.md's owner ruling of
2026-08-02 §2 is directly on point: *"a comment is not a control on a tree this many
sessions share."*

### 2. ⚠ THE HAZARD IN §"The hazard candidate 2 addresses" WAS LIVE ON HEAD — the next `make release` would have deployed NOTHING

BLD-022 §(iv) records the invariant *"`build-backend` builds exactly
`RELEASE_IMAGES`"* as **"verified by set equality 2026-08-18 [MEASURED] and policed
by nothing."** Four days later it was false. `[MEASURED 2026-08-22]`, RUNBOOK R2:

```
declared (RELEASE_IMAGES): 25      built (make -n build-backend): 22
declared but NOT built: optional-explicit-wires-check, commit-sha-exposure-check,
                        capped-schedule-ordering-check
built but NOT declared: (none)
```

`git diff HEAD -- makefile` was empty, so this was HEAD and not a local tree.
Cause: two different lanes on 2026-08-22 (`d56fd6b11`, `67201d125`) added names to
`RELEASE_IMAGES` **and** `AGENT_DEPLOY_SERVICES` and to none of the build groups.
Note what that says about the class — the *deploy* side is now one declaration
looped over (237's fix), so the two lanes got that half right automatically; the
*build* side was still a second hand-written enumeration, and that is the half both
got wrong.

**The consequence, measured at the artefact rather than read off the makefile.**
`push-backend` (`makefile:476`) loops `RELEASE_IMAGES` with `docker push … || exit 1`.
`IMAGE_TAG` is `v1.0.1324`; the live fleet is `v1.0.1323`.

| image | local tags | in build path? |
|---|---|---|
| `optional-explicit-wires-check` | **v1.0.1321 only** | no |
| `capped-schedule-ordering-check` | **none, at any tag, ever** | no |
| `commit-sha-exposure-check` | v1.0.1324 (hand-built that morning) | no |
| `agent-chassis` **(control)** | 1320, 1321, 1322, 1323 | yes |
| `verifier-remit-check` **(control)** | 1320, 1321, 1322, 1323 | yes |

The controls are the discriminating half: a service the release genuinely builds
carries an image at *every* recent tag, so the absences mean something. So
`make release IMAGE_TAG=v1.0.1324` would have built 22 images (~6 min), then died on
the first `docker push` of an image nobody built — **before `deploy-core` and
`deploy-agents`, i.e. nothing would have reached the cluster at all.**

**FIXED, narrowly and structurally, in `95757b6c2`** — not by adding three names but
by deleting the second list: `build-backend: $(addprefix build-,$(RELEASE_IMAGES))`,
plus a `build-github-actions-runner` alias for the one image whose target predates the
naming convention. Set difference is now **25/25 IDENTICAL** and cannot drift.
**Mutation-proven on a COPY** (`cp makefile … && make -f`, never the live shared file
— `WRONG_CALLS.md` 2026-08-22, `f016b07ec`): injecting `nobody-built-this-check` into
`RELEASE_IMAGES` gives exit **2**, *"No rule to make target
'build-nobody-built-this-check', needed by 'build-backend'"*; the live makefile under
the identical command gives exit **0**. Exit codes read **without a pipeline** — a
pipeline's status is its last command's, which is the exact defect `editquality` caught
in `bugs_open/153`'s own council round.

This closes candidate 2 in its strongest available form: `build-backend ≠
RELEASE_IMAGES` is now **unrepresentable**, not merely detected. It does **not** touch
the main gap.

### 3. ⚠ NEGATIVE RESULT — candidate 1 CANNOT be built as worded, and this needs an owner decision

Candidate 1 (the owner's stated intent, 2026-08-18) is *"fail the release when a
service's pinned tag predates the last commit to the sources that service is built
from."* The natural reading of "when the pin was set" is the overlay's git history.
**That history is a fiction on this tree.** `deploy-agents` seds `newTag` in place and
nothing commits the result:

```
$ git status --porcelain -- 'deployments/kustomize/services/*/overlays/production/uk_001/kustomization.yaml' | wc -l
26
$ git diff -- .../agent-chassis/.../kustomization.yaml
-    newTag: v1.0.1239
+    newTag: v1.0.1323
```

Eighty-four tags of divergence in one file, and 26 files in that state right now. A
staleness test dated from the overlay would grade `agent-chassis` — the most-rolled
service on the estate — as pinned at `v1.0.1239`.

**The honest instrument already exists and is this bug's sibling:** BLD-019 stamps
`org.opencontainers.image.revision` on every image and `pkg/buildinfo.GitCommit` into
every backend binary, and BLD-023 records `git_commit` per running pod in
`service_binary_capabilities`. "What commit is this artefact from" is a question you
ask the artefact, not git. **Re-aiming candidate 1 at the artefact is a change to an
owner's stated intent and is therefore put to the owner rather than substituted
quietly.**

### 4. Two adjacent facts found while measuring, filed not fixed

- **`capped-schedule-ordering-check` has no CronJob in the cluster at all.** Overlay
  pinned at `v1.0.1324`, build/push/deploy targets, `RELEASE_IMAGES` and
  `AGENT_DEPLOY_SERVICES` membership — and `kubectl -n ai-persona-system get cronjob`
  does not list it. Scaffolded (`d56fd6b11`, the `316` lane) and never applied. The
  release now builds and pushes an image for a service that does not exist in the
  cluster; harmless, but it means overlay-on-disk and workload-in-cluster are two
  different enumerations and **neither is a superset of the other**.
- **The reverse of the same fact:** `site-discovery-staleness-check` and
  `site-locale-unset-check` run as CronJobs with **no `services/*` overlay on disk**.
  Both are `postgres:16-alpine`, so no freeze risk today — but it bounds what any
  filesystem-only gate may claim about coverage.

### 5. What this contribution does NOT claim

The structural half — the gate's self-referential admission test — is **not fixed**.
A new service born outside `RELEASE_IMAGES` is still invisible to
`check-release-coverage` today. That is the design under way in the lane, and its
shape is constrained by 318's own warning: a makefile-only change is refused by the
council client-side, so the predicate is being written as a Go helper the makefile
calls — which also makes it mutation-testable without editing a file forty sessions
share.
