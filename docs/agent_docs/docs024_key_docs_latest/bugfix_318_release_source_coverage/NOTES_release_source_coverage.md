# NOTES — `bugs_open/318` release source coverage

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep.

---

## 2026-08-22 — session opens: is 318 still valid, and is anyone on it?

**Ownership check first** (`scripts/who-owns.py`, CLAUDE.md § "Before routing work AT
an existing bug"):

- `153` → **OWNED** by `bugfix_153_build_provenance` (10 commits in 14 days, last
  2026-08-17). Its headline fix is FIXED AND LIVE on `v1.0.1283`; what remains is a
  residual list, not an open defect. So this session took the fallback the owner named.
- `318` → filed 2026-08-19 by the `bugs_open/237` lane, **status "OPEN, not started"**,
  no owning lane, no commits against the file since filing. The only other commits in
  its territory are from the `staged_component_build` lane, which *tripped over* the
  blind spot on 08-21/08-22 while creating two new check services and folded them into
  the lists (`67201d125`) — a fix to two instances, explicitly not to the class.
- **Work-item queue**: nothing open matching `release` / `RELEASE_IMAGES` /
  `coverage gate` / `build-backend` / `318`. Query in the RUNBOOK (R1).

**Verdict: 318 is unowned and this session takes it.** Lane created today.

### Why `090` (the diagnosis loop) was NOT run — stated, not omitted

CLAUDE.md's owner ruling of 2026-07-31 says a `bugs_open/` file asserting a
cross-cutting root cause is not filed until it has been through the loop **or the
session states plainly why it substituted equivalent first-hand verification**.

This session is not filing a new root cause. 318's cause was established by the
`bugs_open/237` lane and is already in the concept register as **BLD-022**'s stated
blind spot. Everything added below is a **mechanical measurement with a control**, not
a theory: a set difference from `make -n`, a `docker images` census with two positive
controls, a `kubectl get cronjob` read. Each could have come out otherwise, and one of
them (finding 5) came out **against** the design I expected to recommend. That is the
substitute, declared.

### What was measured `[MEASURED 2026-08-22]`

**1. The gate's admission test is self-referential.** `check-release-coverage`
(makefile `:142`) `continue`s past any overlay whose pinned image is not already in
`RELEASE_IMAGES`:

```make
case " $(foreach i,$(RELEASE_IMAGES),$(REGISTRY)/$(i)) " in \
        *" $$img "*) ;; \
        *) continue;; \
esac;
```

Membership of the list is the gate's own admission criterion, so an image left out at
birth is not *uncovered*, it is *out of scope* — and the gate prints
"Release coverage OK". Already in `LANDMINES.md` (§ "A CHECK SERVICE OMITTED FROM
`RELEASE_IMAGES` AT BIRTH…"), and 8 services have now fallen into it: the original six
(fixed 08-18), then `optional-explicit-wires-check` (born 08-21) and
`commit-sha-exposure-check` (born 08-22).

**2. The `build-backend == RELEASE_IMAGES` invariant is BROKEN ON HEAD.** BLD-022
§(iv) records it as "verified by set equality 2026-08-18 [MEASURED] and policed by
**nothing**". Four days later it is no longer true:

```bash
make -n build-backend 2>/dev/null | grep -oE 'docker build -t [^ ]+' \
  | sed 's#.*/##; s#:.*##' | sort -u > built.txt          # 22
awk '/^RELEASE_IMAGES :=/{f=1} f{print} f&&!/\\$/{exit}' makefile \
  | sed 's/^RELEASE_IMAGES :=//; s/\\//' | tr -s ' \t' '\n' \
  | grep -v '^$' | sort -u > declared.txt                  # 25
comm -23 declared.txt built.txt
```

→ **declared but not built: `optional-explicit-wires-check`,
`commit-sha-exposure-check`, `capped-schedule-ordering-check`.**
Reverse direction (built but not declared): **empty**.

Cause: the names went into `RELEASE_IMAGES` and `AGENT_DEPLOY_SERVICES` but not into
`build-checks`, the `build-backend` prerequisite that fans out to the check images.
`git status --porcelain makefile` is empty and `git diff HEAD -- makefile` is empty, so
this is **HEAD, not local WIP**.

**3. The live consequence, measured at the artefact — the next release aborts.**
`push-backend` (makefile `:476`) loops `RELEASE_IMAGES` with
`docker push … || exit 1`. `IMAGE_TAG` is `v1.0.1324`; the live fleet is `v1.0.1323`.

| image | local tags present | in `build-backend`? |
|---|---|---|
| `optional-explicit-wires-check` | **v1.0.1321 only** | no |
| `capped-schedule-ordering-check` | **none, ever** | no |
| `commit-sha-exposure-check` | v1.0.1324 (hand-built today) | no |
| `agent-chassis` *(control)* | v1.0.1320, 1321, 1322, 1323 | yes |
| `verifier-remit-check` *(control)* | v1.0.1320, 1321, 1322, 1323 | yes |

The two controls are the discriminating half: a service the release really does build
carries an image at **every** recent tag. So `make release IMAGE_TAG=v1.0.1324` builds
22 images (~6 min), then dies on the first `docker push` of an image nobody built —
**before `deploy-core` and `deploy-agents`, so nothing deploys at all.**

**4. Cluster state** (`kubectl -n ai-persona-system get cronjob`):
`optional-explicit-wires-check` runs daily on **v1.0.1321**, three tags behind the
fleet. `capped-schedule-ordering-check` has **no CronJob in the cluster at all** —
overlay, make targets and `RELEASE_IMAGES` membership all exist for a service that was
never applied. `commit-sha-exposure-check` and `content-loss-check` are at v1.0.1324
and have never fired.

**5. ⚠ NEGATIVE RESULT, and it argues against 318's candidate 1 as worded.** Candidate
1 says: *"fail the release when a service's pinned tag predates the last commit to the
sources that service is built from."* Reading "when the pin last moved" out of the
overlay's git history **cannot work on this tree**: `deploy-agents` seds `newTag` in
place and nobody commits the result. Right now **26 production overlays are dirty**,
and `agent-chassis` is committed at `v1.0.1239` while the working tree says
`v1.0.1323` — 84 tags of divergence in one file. Any design that dates the pin from
git is dating a fiction. The honest instrument for "what commit is this artefact
from" is BLD-019's OCI label / binary stamp, or BLD-023's
`service_binary_capabilities.git_commit` — not the overlay.

**6. The source closure IS mechanically derivable.** Every backend dockerfile names its
build target (`./cmd/X`, or `cmd/X/main.go` for `git-adapter` and
`remote-job-spawner`); `go list -deps ./cmd/config-key-audit` returns 309 packages in
0.29 s. But **six images share `./cmd/config-key-audit`** and differ only in CMD flags
and a COPY'd acks JSON, so a closure computed from Go deps alone would call all six
identical and would miss the acks file — which for `commit-sha-exposure-check` is the
check's own definition of a permitted exception.

**7. "Has a dockerfile" is too wide a net.** Five backend dockerfiles have no
production overlay: `tools-api`, `migrator`, `seeder`, `platform`, `workflow-monitor`.
The right enumeration is "an overlay pins a `$(REGISTRY)/…` image".

**8. And the filesystem is not the whole cluster.** `site-discovery-staleness-check`
and `site-locale-unset-check` run as CronJobs with no `services/*` overlay on disk.
Both are `postgres:16-alpine`, so no freeze risk today — but it bounds what any
filesystem-only gate can claim.
