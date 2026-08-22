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
  > **⚠ CORRECTED 2026-08-22, same day, by the planning agent's adversarial pass — the
  > wording above would send a reader to a directory that exists.** Both services DO have
  > `deployments/kustomize/services/<name>/base/`; what they lack is an `overlays/` tree
  > (base-only, applied by hand). The substance survives — a gate globbing `overlays/…`
  > cannot see them — but "no `services/*` overlay on disk" reads as "no directory", and
  > anyone who greps will find one and log this as a wrong call.
- **And one finding of mine was simply wrong, in a way that made the fix better.** I
  recorded that `tools-api` has no production overlay. It has one — at
  `deployments/kustomize/services/tools-api/overlays/production/kustomization.yaml`, with
  **no region directory** — which the gate's fixed `$(OVERLAY_PATH)` glob
  (`production/uk_001`) can never see. It pins a placeholder (`newName: IMAGE_TAG`,
  `newTag: latest`) and has no workload, so it is harmless today. **The corrected lesson
  is worth more than the correction: the enumeration must not bake in the region depth**,
  and the replacement gate walks to any depth because of this.

### 5. What this contribution does NOT claim

The structural half — the gate's self-referential admission test — is **not fixed**.
A new service born outside `RELEASE_IMAGES` is still invisible to
`check-release-coverage` today. That is the design under way in the lane, and its
shape is constrained by 318's own warning: a makefile-only change is refused by the
council client-side, so the predicate is being written as a Go helper the makefile
calls — which also makes it mutation-testable without editing a file forty sessions
share.

---

## 2026-08-22 (later) — **THE MAIN GAP IS CLOSED.** Candidate 2 by construction, candidate 1 REPLACED and its remnant put to the owner

**Council APPROVED, round 1** — `83442a5a-e66d-4772-8872-b445f521d47b`, 3 advisory
objections, none high. Register **BLD-026**. Lane:
`docs/agent_docs/docs024_key_docs_latest/bugfix_318_release_source_coverage/`.

### What shipped

| candidate | state |
|---|---|
| **2** — `build-backend` builds exactly `RELEASE_IMAGES` | **DONE by construction** (`95757b6c2`). Not by adding three names: `build-backend: $(addprefix build-,$(RELEASE_IMAGES))` deletes the second enumeration entirely, so the two sides cannot drift. Set difference 25/25 **by identity**. |
| **1** — the content-change trigger | **RULED OUT by the owner 2026-08-22 — *"we can skip the 18 August staleness build"* — and REPLACED by an accumulation budget on the exemption list. Not deferred: do not re-file it. See the ruling section at the foot. |
| **the main gap** — an image of ours the release does not build | **CLOSED** (`f16daa34a`, `9b87dcfac`). The gate's admission test is INVERTED. |
| **3** — a daily cluster census | **not built**, deliberately: it is detection, it needs RBAC, and it earns its own commit and its own review. |

**The fix in one line:** the predicate moved to `pkg/releaseset` behind a thin
`cmd/releasecheck`, and *"is this overlay's image one the release builds?"* stopped being
the **admission test** and became the **question**. An image of ours that no release
builds is now a violation; the only exit is an explicit `OWN_LINEAGE := <svc>:<target>`
entry — opt-in, unsafe side default OFF, and not architecture-scope under RFC_022's
narrowing.

**Why Go and not more shell, since this file raised the question:** `scripts/council-scope.sh:57`
is `COUNCIL_SCOPE_CODE_RE='^(platform|internal|pkg)/'` — verified verbatim, not asserted.
`cmd/` is outside it, so a helper written wholly in `cmd/` would have drawn **no review at
all**. The layering is what made the round possible. It is also what let the mutation
proofs be table rows: this morning a session mutated the shared makefile in place to prove
the OLD gate discriminates and another session committed the file inside the window.

### `OWN_LINEAGE` was going to ship empty and the first live run found an entry

`admin-dashboard` pins one of our images that `build-backend` does not build — it is a
frontend, built from the working tree with no `ref_build` and no provenance stamp
(BLD-019/BLD-020 scope frontends out explicitly) — and yet it **is** released, as the last
goal of `pinned_sweep` via `release-dashboard`. Genuinely own-lineage, genuinely covered,
and **nothing on disk said so**. The old gate could not have said so: its image is not in
`RELEASE_IMAGES`, and that was the admission test. The entry also records what it does not
cover — `release-backend` omits `release-dashboard`.

That also answers the `guardian` seat's question about the exemption list's completeness:
the gate enumerates the **whole filesystem**, so its first green run over 31 our-registry
overlays **is** the completeness check. Any other out-of-band overlay would have been
named in the same breath.

### Proof, and the instrument shown able to fail

Six mutations on a **copied** tree (never the live makefile): verbatim copy → exit 0;
`agent-chassis` out of the deploy list → `NO RELEASE PATH` (**this reproduces the original
`237` bug state**); `content-loss-check` out of `RELEASE_IMAGES` → `OUR IMAGE, NO RELEASE
BUILDS IT` (**the shape the old gate passed**); `OWN_LINEAGE` emptied → names
`admin-dashboard`; a bare exemption → `EXEMPTION NAMES NO RETAG TARGET`; `RELEASE_IMAGES`
renamed away → `THE CHECK COULD NOT RUN`, never a pass; `RETAG_EXEMPT` renamed away →
**exposes** what it cleared rather than refusing. Nine table cases each name, in their
comment, what a different result would mean.

**A commit-time advisory rides alongside** (`scripts/pattern-check.py`,
`check_unlisted_release_overlay`) because the hard gate does not fire until the next
release, and the makefile's previous remedy — a comment in capitals — failed twice in two
days with the owner ruling in front of both authors. Precision measured against real
history: of **21** commits since 2026-07-15 that added a production overlay it fires on
**7**, and all 7 are the documented incidents (`render-audit-adapter`,
`github-actions-runner-vmsites`, `component-render-check`, `shared-output-fields-check`,
`verifier-remit-check`, `optional-explicit-wires-check`, `commit-sha-exposure-check`).
**Zero false positives** on the other 14.

### The council found a real defect, in the fix, of the fix's own disease

`bug_historian` (medium): the new scanner read only the **first** element of an overlay's
`images:` block. An overlay pinning two images with ours second would produce no `Pin` and
the gate could never flag it — *a check reporting clean about the thing it never looked
at*, reproduced in miniature inside the change written to end it. Measured before acting:
no kustomization anywhere under `deployments/` has more than one element, so it was
**latent**, not live. The cap was **removed** rather than warned about.

### ⚠ WHAT IS STILL OPEN — an owner decision, not work

**Candidate 1 as worded on 2026-08-18 cannot be built**, and re-aiming an owner's stated
intent is his call. Two facts force it:

1. **The pin's history is fiction.** `deploy-agents` seds `newTag` in place and nobody
   commits it — 26 production overlays dirty on 2026-08-22, `agent-chassis` committed at
   `v1.0.1239` against `v1.0.1323` on disk **and in the cluster**. A staleness test dated
   from that history grades the most-rolled service on the estate as the most frozen. Now
   in `LANDMINES.md`.
2. **After P1, the trigger's population is the exemption list, and it has one entry.**
   Every other image of ours is now in the release, and BLD-020 makes a release one
   commit — so *"stale iff its source moved"* can only ever bite an `OWN_LINEAGE` service.

**The recommendation** is to implement it as the standing obligation attached to an
exemption — an `OWN_LINEAGE` entry declares its source closure (the dockerfile's Go target
via `go list -deps`, **plus** `go.mod`, `go.sum`, the dockerfile itself and everything it
`COPY --from=builder`s), and the daily census grades that service's artefact-commit
(BLD-019's OCI label / BLD-023's `git_commit`) against it — rather than as a release gate
over an empty population. **Not built pending the ruling.**

### Two adjacent things, filed not fixed

- `capped-schedule-ordering-check` has an overlay, build/push/deploy targets and release
  membership, and **no CronJob in the cluster**. The next release will create it. Belongs
  to the `316` lane.
- **The next release must run at ≥ `v1.0.1325`.** `v1.0.1324` is contaminated:
  `commit-sha-exposure-check:v1.0.1324` was hand-built from an unpinned commit and three
  overlays already pin 1324, so a same-tag re-push hits the node-cache landmine.

### Close condition

**Not** a green gate — that only says the gate agrees with today's tree. It closes when a
real release has run under it (the `guardian` seat's objection is that this is the single
choke point every deploy passes through, and it is right that the first run should be
watched), and when the owner has ruled on candidate 1's re-aim.

---

## ⚖️ OWNER RULING 2026-08-22 — **SKIP the staleness build; guard the EXEMPTION LIST instead.** Candidate 1 is closed, not parked

> *"we can skip the 18 August staleness build."* — and, on the options put to him,
> **skip it AND add the cheap guard on the excused list.**

This supersedes the section above headed *"⚠ WHAT IS STILL OPEN — an owner decision, not
work"*, and supersedes the recommendation in the lane's
`PLAN_2026-08-22_release_source_coverage.md` §2 that the predicate be re-aimed at the
artefact and attached to `OWN_LINEAGE` entries. **Neither the 2026-08-18 wording nor the
artefact-anchored re-aim is to be built.**

**Read this as CLOSED, not as a gap.** The reason matters, because a later reader who
finds only *"the content-change trigger was never built"* will re-file it — this estate has
a landmine for exactly that shape (*a closed blocker keeps being obeyed*). Two independent
facts had already emptied the idea before the ruling:

1. **As worded it is uncomputable on this tree.** It dates a service's pin from the
   overlay's git history, and `deploy-agents` seds `newTag` in place with nobody committing
   the result — **26 production overlays dirty as of 2026-08-22**, `agent-chassis` recorded
   at `v1.0.1239` against `v1.0.1323` on disk *and* in the cluster. A staleness test on that
   history grades the most-rolled service on the estate as the most frozen. In `LANDMINES.md`.
2. **After the P1 inversion its population is the exemption list, and that list holds ONE
   entry as of 2026-08-22** (`admin-dashboard:deploy-dashboard`). Every other image of ours
   is in the release, and BLD-020 makes a release one commit — so *"stale iff its source
   moved"* could only ever bite an `OWN_LINEAGE` service. A mechanism with one possible
   subject is one nobody exercises, which is this estate's own documented failure mode
   (BLD-023 declined `assert_live_capability()` on precisely that ground).

### What was built instead, and why it is the right trade

**The risk skipping actually carries is not staleness — it is that the exemption list
becomes the new hiding place.** The old one was "not in `RELEASE_IMAGES`"; eight services
found it, two of them within three days of the ruling meant to close it. `OWN_LINEAGE` is
better paperwork for the same shape unless something watches the pile.

So (`8fe69e6c6`), two halves:

- **`ExemptionBudget = 3`** in `pkg/releaseset/predicates.go` — the release fails when the
  excused list grows past it, naming every entry. It polices the **accumulation**, not the
  entry, the same shape as the optional-key budget (RFC_022, owner-set N=10, **WFA-013**).
  N is **lower than 10 deliberately**: an exemption from the release is rarer and costlier
  — what it permits is a service running months-old code with nothing reporting it, which
  is this bug. **It is a judgement and one line; the owner may set it otherwise.**
- **The gate names the standing exemptions on every GREEN run** — *"1 of those is EXCUSED
  from the release (OWN_LINEAGE, budget 3) — admin-dashboard → deploy-dashboard"*. This is
  the half that does the work: a threshold silent until it trips is a threshold nobody
  watches, and this puts the count in front of whoever runs the release, so the **fourth**
  exemption is noticed as it arrives rather than when a number is crossed.

**Stated plainly, because it is the honest objection:** the budget **cannot fire today**.
What distinguishes it from the mechanisms this estate has declined to build is that it
needs no caller — it runs on every `deploy-core` whether or not it fires, so it is
exercised continuously while quiet. That argument is written at the constant so the next
reader can disagree with it rather than guess at it.

**Proven** (`T10` plus a copied-tree run, never the live makefile): at the budget, silent;
one over, reported and naming the whole accumulated set, with a remedy that offers a
**reviewed** raise rather than a bare refusal. Honest detail from the shell control —
fabricating three exemptions trips the *no-overlay* guard instead, because an exemption
naming a service with no overlay is separately a violation; the two guards are independent,
and the clean at-budget case is the unit test, which uses real overlays.

### What still holds after the ruling, so nothing is lost with it

- `OWN_LINEAGE` is unchanged — an explicit, greppable exemption naming the target that
  moves the service. It never depended on the staleness predicate.
- Staleness of a *release-covered* service was never this bug's question and is answered
  elsewhere: **BLD-019** stamps the commit into every image and binary, **BLD-020** makes
  one release one revision, **BLD-023** lets a running pod publish what it can do.
- The **cluster-side census** (candidate 3) is untouched by this ruling. It is a different
  question — *is anything declared but not running, running but undeclared, or off the
  fleet tag?* — and it is the only thing that sees the two shapes found while measuring:
  `capped-schedule-ordering-check` (declared, never applied) and two CronJobs running with
  no overlay on disk. Unbuilt, not owed, its own round if anyone wants it.

### Close condition, restated after the ruling

**One thing remains: a real release under the gate.** The `guardian` seat's objection is
the reason — this is the single choke point every deploy passes through, its admission
logic has been inverted, and it can now fail a release it previously passed. A green run on
today's compliant tree proves only that it *could* pass. When a release has run under it,
`318` closes.

---

## 2026-08-22 — PHASE 2: the cluster census (candidate 3) is built, hand-run

`make release-census` / `go run ./cmd/releasecheck --census`. Register **BLD-026**, **council APPROVED round 2, all reviewers** — `Council-Reviewed: b0883c17-32a1-434d-b0ab-114df4cb04b1`. Round 1 was REVISE on a gating HIGH that found a real defect; see the round record at the foot of this section.

**This closes candidate 3 as a capability, not as a scheduled job** — see the two stated
limits below before describing it as a detector.

### Why it exists, in one sentence

Everything else in this fix reads the **filesystem** and asks *can a release reach this
service?*; the census reads the **cluster** and asks *does what is running match what is
declared?* — because the filesystem and the cluster are two enumerations and **neither is
a superset of the other**, which is exactly why the two shapes found while measuring were
invisible to the gate in both directions.

### The first live run `[MEASURED 2026-08-22]`

```
29 of 45 workloads run a docker.io/aqls/ image; the fleet is on v1.0.1323.  5 findings.
```

| kind | service | independently confirmed by |
|---|---|---|
| AHEAD OF THE FLEET (hand-deployed) | `commit-sha-exposure-check` v1.0.1324 | the `docker images` census earlier the same day |
| AHEAD OF THE FLEET (hand-deployed) | `content-loss-check` v1.0.1324 | same |
| BEHIND THE FLEET TAG | `optional-explicit-wires-check` v1.0.1321 | the overlay census earlier the same day |
| DECLARED BUT NOT RUNNING | `capped-schedule-ordering-check` | found by hand; contributed to the `316` lane |
| DECLARED BUT NOT RUNNING | `component-source-vocabulary-check` | the `309` lane's own commit: *"NOT yet built or deployed"* |

**Zero false positives**, and no `RUNNING BUT NOT DECLARED` — every `aqls` workload in the
cluster is accounted for by a declaration.

### ⚠ The run found a defect in its OWN report, and that is the part worth reading

The two v1.0.1324 findings first came out as **"RUNNING AN OLD FLEET TAG … while the fleet
is on v1.0.1323"**. They are not old. They are **newer**, hand-deployed.

**A report that states the opposite of the truth is worse than one that stays quiet** — a
reader chasing a frozen service finds one that is, if anything, too new, and concludes the
instrument works. Same family as the blind-pass landmine, one rung along: not a check that
misses, but a check that asserts the inverse. Fixed three ways, each with a test:

- the kind splits by **direction**, with **opposite remedies** — a straggler wants a
  release; a hand-deploy wants the tag **never reused**, because a same-tag re-push serves
  the node's cached image;
- tags are ordered **numerically, not lexically**. Not theoretical: this estate is on
  `v1.0.13xx`, so `v1.0.999 → v1.0.1000` is already behind us, and a string comparison
  calls 999 the newer. Pinned by a test, along with the real `v1.0.948 → v1.0.1126` runner
  freeze;
- an **unorderable** tag (`latest`, a date, a different arity) gets its own kind rather
  than a guessed direction.

**What would have caught it earlier: nothing in the unit tests, because the fixtures
covered the direction I had in mind and not its mirror.** Generalisable, and now in the
lane NOTES: *when a finding has a direction, write the fixture for BOTH directions before
running it anywhere* — a one-sided fixture cannot fail on the side it omits.

### Two limits, stated so they are not read as oversights

- **HAND-RUN ONLY.** No CronJob, no RBAC manifest, no `doc_notes` row. Nothing runs it
  unless a person does, and it must **not** be described as a live detector. Scheduling is
  a separate decision with its own round, and the split is deliberate: *"detection works;
  SCHEDULE and DISPATCH do not"* is a measured finding on this estate, so a detector and
  its driver are not claimed together.
- **First container only.** `readWorkloads` reads `containers[0]`; a sidecar carrying one
  of our images would be missed. None exists today. Named because it is the **same shape**
  the council caught in round 1 (the `images:` block cap) — if a sidecar appears, look here.

### What this does NOT change

The close condition is unchanged and is still a single item: **a real release under the
gate.** The census is a detector and cannot substitute for that.

### The census's council rounds — REVISE then APPROVED, and round 1 earned its cost

**Round 1: REVISE**, gating HIGH from `editquality`, and it found a real defect:
`modalTag` broke ties with `tag > best` — a **lexical** comparison, in the same file as
and one screen below the numeric comparator written to fix exactly that. A tie between
five workloads on `v1.0.999` and five on `v1.0.1000` picks `v1.0.999`: lexically higher,
numerically older. And because the fleet tag is **the single value every straggler and
ahead-of-fleet finding is measured against**, that inverts the whole report rather than
one line of it.

**Three occurrences of the same defect, in one file, in one afternoon:** the shipped
report that called two hand-deploys "old" (found by the live run), the first repair for it
(caught before commit), and this tie-break (caught by the council). The lesson is written
at the function: **a helper that does the comparison correctly does not protect the call
sites that do not use it** — `TestCompareTags` was green through all three.

Fixed and mutation-proven the decisive way round: restoring `tag > best` fails the new
test with its own message. The test also pins the end-to-end census on that fleet, that a
clear 3-2 majority still wins outright so the tie-break does not become the rule, and that
an unorderable tie is deterministic across 20 calls.

`reuse_agent`'s objection was **answered with three measurements and its premise was
wrong**: the named CronJob checks do not read the cluster, they read Postgres —
`config-key-audit`'s own dockerfile says *"no kubectl in this image, no pods/exec RBAC"*.
Six inline `kubernetes.NewForConfig` sites and no shared wrapper; zero `.List(` calls for
Deployments/CronJobs/DaemonSets anywhere else in the estate. Nothing to reuse.

**Round 2: APPROVED, all reviewers**, three advisory lows. Two were acted on:

- **`debug_historian`** — the census measures **tags**, and a tag is not the code: a
  same-tag rebuild serves the node's cached image, so a service *on* the fleet tag can
  still be stale. Stated rather than fixed, in the code and in BLD-026: **a clean census
  means "every service is on the tag it should be on", NOT "every service is running the
  code it should be running."** The artefact-side answer belongs to BLD-019/020/023 and
  duplicating it here would give one question two answers that can disagree.
- **`architecture`** — `cluster.go` is the **seventh** inline k8s client bootstrap with no
  shared wrapper, and unwinding seven call sites later costs an RFC. Not extracted here
  (five unrelated call sites inside a bug-fix round is the scope shape the guardian seat
  vetoes) but now tracked in BLD-026 with all seven named.

---

# ✅ CLOSED 2026-08-22 — the close condition is met: a real release ran under the gate, and the census's predictions came true at the artefact

`v1.0.1326`, whole-fleet, owner-run. **20 Deployments and 11 CronJobs all on one tag.**

## The gate ran, and it did not break a real release

`auth-service` and `core-manager` are on `v1.0.1326` `[MEASURED 2026-08-22]`. That is the
evidence that matters: both move only through `deploy-core`, `deploy-core` requires
`update-kustomization-images`, which requires `check-release-coverage`, and a non-zero exit
propagates through `pinned_sweep`'s `|| exit 1`. **The gate ran and passed on a real
whole-fleet release.**

> **⚠ Read that precisely, because BLD-022's entry had to be corrected for exactly this
> overclaim.** The tree was compliant, so the gate **could only have passed**. This
> release proves the inverted predicate does not break a real release; it does **not**
> prove the gate discriminates. Its discriminating power rests on the six mutation
> controls (copied tree, never the live makefile), which is where it should rest.
> *Exercised, not proven.*

## The census: 5 findings → 1, and every prediction confirmed

|  | before the release | after |
|---|---|---|
| workloads on our registry | 29 of 45 | **31 of 47** |
| fleet tag | v1.0.1323 | **v1.0.1326** |
| findings | **5** | **1** (and it is a NEW service, filed since) |

Each of the five resolved exactly as the census said it would `[MEASURED 2026-08-22]`:

| finding | before | after |
|---|---|---|
| BEHIND THE FLEET TAG `optional-explicit-wires-check` | v1.0.1321 | **v1.0.1326** |
| AHEAD (hand-deployed) `commit-sha-exposure-check` | v1.0.1324 | **v1.0.1326** |
| AHEAD (hand-deployed) `content-loss-check` | v1.0.1324 | **v1.0.1326** |
| DECLARED BUT NOT RUNNING `capped-schedule-ordering-check` | no CronJob | **created 15:09:35Z**, v1.0.1326 |
| DECLARED BUT NOT RUNNING `component-source-vocabulary-check` | no CronJob | **created 15:09:37Z**, v1.0.1326 |

**The two creation timestamps are the strongest single piece of evidence here.** This lane
predicted, in writing and in a contribution to the `316` lane, that *"the next release will
CREATE that CronJob, because it is in `AGENT_DEPLOY_SERVICES` and `deploy-agents` applies
overlays."* It did, two seconds apart, in the release's own apply loop. **That prediction
could have come out otherwise** — the release could have aborted at push (which is what
`95757b6c2` fixed), or the overlay could have failed to apply — and it did not.

**The `v1.0.1324` contamination was also avoided as recommended:** the fleet went to
`v1.0.1326`, so the two hand-built images at 1324 were never re-pushed over.

**The one remaining finding is the census working as designed:**
`live-declaration-drift-check`, created by the `363` lane (`18661b3c7`) since the release —
declared and not yet running. It is the *next* one, caught the day it appeared rather than
in three months.

## What is closed, and by what

| candidate | state |
|---|---|
| the main gap — our image, no release builds it | **CLOSED.** Admission test inverted (`f16daa34a`, `9b87dcfac`); council APPROVED `83442a5a`; live and exercised by `v1.0.1326` |
| **2** — `build-backend` == `RELEASE_IMAGES` | **CLOSED by construction** (`95757b6c2`). Cannot drift; the second enumeration is gone |
| **1** — the content-change / staleness trigger | **RULED OUT by the owner 2026-08-22.** Not deferred. Replaced by the exemption accumulation budget (`8fe69e6c6`) |
| **3** — the cluster census | **BUILT, council APPROVED `b0883c17`** (REVISE→approved). Hand-run; scheduling is a separate decision |

## What is deliberately still open, and belongs to nobody as a debt

- **The census has no driver.** Hand-run only; no CronJob, no RBAC, no `doc_notes` row.
  Stated at the makefile target, in `pkg/releaseset/census.go` and in BLD-026. *"Detection
  works; SCHEDULE and DISPATCH do not"* — a detector and its driver were not claimed
  together, and scheduling it is its own decision with its own round.
- **A tag is not the code.** A clean census means *"every service is on the tag it should be
  on"*, **not** *"every service is running the code it should be running"* — a same-tag
  rebuild serves the node's cached image. Raised by the council's `debug_historian` seat;
  stated rather than fixed, because the artefact-side answer is BLD-019/020/023's and two
  answers to one question can disagree.
- **Seven inline `kubernetes.NewForConfig` bootstraps, no shared wrapper.** Raised by the
  `architecture` seat, tracked in BLD-026 with all seven named. Not extracted inside a
  bug-fix round.

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_318_release_source_coverage/`
(standing five). **Register:** **BLD-026**, with BLD-022 corrected in two places.
