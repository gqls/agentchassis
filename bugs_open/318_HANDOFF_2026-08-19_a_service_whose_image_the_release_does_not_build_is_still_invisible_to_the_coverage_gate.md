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
