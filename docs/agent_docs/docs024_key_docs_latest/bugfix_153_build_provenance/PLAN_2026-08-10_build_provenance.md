# PLAN_2026-08-10_build_provenance — Bug 153, candidate 1 (+4): stamp the commit into every backend image and binary, verify at the pod

**Workstream:** `docs/agent_docs/docs024_key_docs_latest/bugfix_153_build_provenance/`
**Bug:** `bugs_open/153_HANDOFF_2026-07-29_image_tag_does_not_imply_a_rebuild.md`
**Scope decision (already made, do not relitigate):** implement candidate 1 (stamp + verify) generically via the two makefile `define`s and a mechanical rollout to all 14 backend services, plus candidate 4 (widen `verify-agent-images`) cheaply. Candidates 2 and 3 are explicitly deferred — see §8.
**Status of the defect, re-verified 2026-08-10 on `v1.0.1279` (pods `agent-chassis-8496665bb8-{f6svp,sskxd}`):** `strings /app/agent-chassis` contains 0 version strings and 0 40-hex shas; `grep -c ldflags build/docker/backend/agent-chassis.dockerfile makefile` → 0/0. Unfixed and current.
**Author:** drafted by claude-fable-5 (Plan agent), 2026-08-10, from research done in this session — see `NOTES_build_provenance.md` for the ownership/validity check that preceded it.

## 0. Design decisions up front (so the rest reads mechanically)

- **Full sha, not short, for ref builds.** The bug's own positive control says "the pod must report exactly `git rev-parse HEAD` of the ref built" — that is the full 40-hex sha. A 40-hex string is also currently a measured-zero pattern in the binary, so post-fix extraction by pattern is unambiguous. The existing `--short` echo at `makefile:119` stays untouched for humans. (This is a deliberate deviation from the bug file's sketch, which reuses the short sha; the full sha is strictly stronger and costs nothing.)
- **Tree builds never wear a clean sha.** `tree_build` stamps `<shortsha-of-HEAD>-tree`, unconditionally — even when `git status --porcelain` is empty. The `-tree` suffix structurally cannot match a `[0-9a-f]{40}` grep, so a WIP image can never impersonate committed state. Unconditional because "porcelain was empty at build start" is a race, not a guarantee, and a simple rule beats a conditional one.
- **No build timestamp in the binary.** `-X ...BuiltAt=$(date)` would make two builds of the same commit produce different binaries for zero verification power — the image's `.Created` (readable via `docker image inspect`) and the pod's `.status.startTime` already answer "when". We add an `org.opencontainers.image.created` label at the docker CLI (metadata only, no cache impact) so candidate 4 can print it. The binary stays deterministic: sha only.
- **The `--label` goes on the `docker build` command line, not in the Dockerfiles.** This means every image built through `ref_build`/`tree_build` gets the revision label even if a Dockerfile misses its edit — including the two non-service CronJob images (`component-render-check`, `shared-output-fields-check`) that also build via `ref_build` and are NOT getting Dockerfile/main.go edits this round. The Dockerfile edit is then only 2 lines (ARG + ldflags), not 3.
- **The ldflags target is a shared package var, not `main.GitCommit`.** `-X github.com/gqls/agentchassis/pkg/buildinfo.GitCommit=...` is identical across all 14 Dockerfiles regardless of the main package's build style (two services build `main.go` by filename — see §3). One string to copy, no per-service variation.
- **Load-bearing subtlety: `-X` on a package that is not linked in is silently ignored.** The one-line `main.go` import+log (§4) is therefore not cosmetic — it is what forces `pkg/buildinfo` into every binary so the stamp takes. Do not skip a service's main.go edit "because the label is enough".
- **Only build path is the makefile.** Measured today: no `.github/workflows/` directory exists and no `docker build` appears anywhere under `scripts/` — `ref_build`/`tree_build` are the only producers of these images, so editing the two defines covers everything.

## 1. `pkg/buildinfo` — the new shared package

New file `pkg/buildinfo/buildinfo.go` (new directory; `pkg/` currently holds only `diagnose/` and `models/`; no buildinfo/GitCommit mechanism exists anywhere — grepped `platform/`, `pkg/`, `cmd/`):

```go
// Package buildinfo carries build-time provenance, stamped by the makefile's
// ref_build/tree_build via -ldflags "-X ...". See bugs_open/153.
package buildinfo

// GitCommit is the full 40-hex commit sha the binary was built from (ref
// builds), "<shortsha>-tree" for working-tree builds, or "unknown" for any
// build that bypassed the makefile.
var GitCommit = "unknown"
```

That is the whole package. No `BuiltAt` (§0), no init, no dependencies — inert until read, which is what keeps it in the normal council gate rather than an architecture RFC (2026-07-29 owner ruling).

**Serving it over HTTP:** `platform/health/server.go` has `AddHandler(path, handler, methods...)` (line 47) and a `/health` handler (line 54), so a `/buildinfo` endpoint is feasible — but only `agent-chassis`'s main constructs the health server directly; the other services build theirs inside their internal adapter/agent packages, so an endpoint is NOT a uniform 1-line-per-service change. **Decision: log line only this round.** A follow-up could register `/buildinfo` inside `health.Server.Start()` itself (one shared edit, every service inherits on next rebuild) — noted as deferred, severable work; if picked up, verify the registration point at `platform/health/server.go:52-59` then.

## 2. The makefile edit — `ref_build` and `tree_build` only

`ref_build` (`makefile:114-131`). The docker-build stanza at lines 126-130 becomes (compute the sha once, inside the existing `&&` chain, after the ref-verify guard at line 117 has already run):

```make
@CTX=$$(mktemp -d /tmp/ref-ctx-$(1).XXXXXX) && \
trap 'rm -rf "$$CTX"' EXIT && \
GIT_COMMIT=$$(git rev-parse '$(REF)^{commit}') && \
git archive $(REF) | tar -x -C "$$CTX" && \
docker build -t $(REGISTRY)/$(1):$(IMAGE_TAG) \
	--build-arg GIT_COMMIT=$$GIT_COMMIT \
	--label org.opencontainers.image.revision=$$GIT_COMMIT \
	--label org.opencontainers.image.created=$$(date -u +%Y-%m-%dT%H:%M:%SZ) \
	-f "$$CTX/build/docker/backend/$(1).dockerfile" "$$CTX"
```

`tree_build` (`makefile:135-144`). The plain recipe line at 142-143 becomes a single shell command so the marker is computed once:

```make
GIT_COMMIT="$$(git rev-parse --short HEAD 2>/dev/null || echo unknown)-tree" && \
docker build -t $(REGISTRY)/$(1):$(IMAGE_TAG) \
	--build-arg GIT_COMMIT=$$GIT_COMMIT \
	--label org.opencontainers.image.revision=$$GIT_COMMIT \
	--label org.opencontainers.image.created=$$(date -u +%Y-%m-%dT%H:%M:%SZ) \
	-f build/docker/backend/$(1).dockerfile .
```

Notes:
- These two defines are the whole makefile build change — all 16 `$(call ref_build,...)` sites (14 services + 2 CronJob checks) and the `build-%-ref`/`build-%-tree` pattern rules inherit it.
- The two CronJob images get the labels but do not consume the `GIT_COMMIT` build-arg (no Dockerfile edit this round) — docker prints a harmless "build-arg was not consumed" style warning for them. Expected; do not "fix" it by widening scope.
- Do NOT touch `IMAGE_TAG` semantics, the bump requirement, or `push-backend` (`makefile:269-284`) — the bug file's explicit landmine and our deferred candidates 2/3.
- Same-commit rebuild caching is unaffected: an unchanged `GIT_COMMIT` arg value does not invalidate layers, and labels are metadata-only.

## 3. The per-Dockerfile edit — 2 lines × 14 files

Pattern, using `build/docker/backend/agent-chassis.dockerfile` as the reference (current lines 1-6):

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG GIT_COMMIT=unknown
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-X github.com/gqls/agentchassis/pkg/buildinfo.GitCommit=${GIT_COMMIT}" \
    -o agent-chassis ./cmd/agent-chassis
```

Rules: `ARG GIT_COMMIT=unknown` goes in the **builder** stage, immediately before the `go build` line (placing it after `COPY . .` means changing the arg only invalidates from the build step); the ldflags clause is inserted into the **existing** `go build` line, preserving every existing flag; no `LABEL` in any Dockerfile (CLI-applied, §0); no other line changes. `browser-runner-adapter`'s separate playwright-cli build (its line 31) is left untouched.

The 14 files with their verified `go build` lines (each read from the file today — module path `github.com/gqls/agentchassis` confirmed from `go.mod`):

| dockerfile (`build/docker/backend/`) | verified build line target | cmd path |
|---|---|---|
| auth-service.dockerfile:6 | `-o auth-service` | `./cmd/auth-service` |
| core-manager.dockerfile:6 | `-o core-manager` | `./cmd/core-manager` |
| agent-chassis.dockerfile:6 | `-o agent-chassis` | `./cmd/agent-chassis` |
| reasoning-agent.dockerfile:7 | `go build -v -o reasoning-agent` | `./cmd/reasoning-agent` |
| web-search-adapter.dockerfile:6 | `-o web-search-adapter` | `./cmd/web-search-adapter` |
| web-scrape-adapter.dockerfile:6 | `-o web-scrape-adapter` | `./cmd/web-scrape-adapter/` |
| git-adapter.dockerfile:16 | `go build -o git-adapter` | `cmd/git-adapter/main.go` (file build) |
| image-generator-adapter.dockerfile:6 | `-o image-generator-adapter` | `./cmd/image-generator-adapter` |
| thunder-adapter.dockerfile:6 | `-o thunder-adapter` | `./cmd/thunder-adapter` |
| analyser-adapter.dockerfile:6 | `-o analyser-adapter` | `./cmd/analyser-adapter` |
| browser-runner-adapter.dockerfile:23 | `-o browser-runner-adapter` | `./cmd/browser-runner-adapter` |
| content-creator-agent.dockerfile:17 | `-o /app/content-creator-agent` | `./cmd/content-creator-agent` |
| remote-job-spawner.dockerfile:11 | `-o /remote-job-spawner` | `./cmd/remote-job-spawner/main.go` (file build) |
| kafka-scheduler.dockerfile:13 | `-o kafka-scheduler` | `./cmd/scheduler` (NOT cmd/kafka-scheduler) |

The two file-builds (`git-adapter`, `remote-job-spawner`) are fine: both cmd dirs contain only `main.go` (verified), and a file-build still links imported packages, so the `-X` lands once main.go imports buildinfo.

**Binary paths in the final image vary** — matters for the `strings` verification, read each final stage rather than assuming `/app/<name>`: agent-chassis → `/app/agent-chassis`, kafka-scheduler → `/app/kafka-scheduler`, git-adapter → `/root/git-adapter`, remote-job-spawner → `/remote-job-spawner` (all verified); the remaining ten to be read off during implementation.

## 4. The main.go edit — 1 import + 1 log line × 14 files

Two idioms exist (surveyed all 14 today):

**(a) The 12 services using `platform/logger` (zap)** — auth-service, core-manager, agent-chassis, reasoning-agent, web-search-adapter, web-scrape-adapter, git-adapter, image-generator-adapter, thunder-adapter, analyser-adapter, browser-runner-adapter, content-creator-agent. All follow `appLogger, err := logger.New(cfg.Logging.Level)` … `defer appLogger.Sync()`. Immediately after the `defer`:

```go
appLogger.Info("build provenance", zap.String("git_commit", buildinfo.GitCommit))
```

plus `"github.com/gqls/agentchassis/pkg/buildinfo"` in the import block. All 12 already import `zap` or use it transitively — verify per-file; where `zap` is not yet imported directly (some mains only pass the logger on), add it or use `appLogger.Sugar().Infof` — prefer adding the `zap` import for consistency.

**(b) The 2 direct-zap services** — `cmd/remote-job-spawner/main.go` (line ~127: `logger, _ := zap.NewProduction()`) and `cmd/scheduler/main.go` (line ~59: `logger := buildLogger()`, then `logger.Info("kafka-scheduler starting")`). Same line, placed right after their `defer logger.Sync()`; scheduler's can extend the existing "kafka-scheduler starting" Info call with the field instead — implementer's choice, keep it one line either way.

Consistent message string `"build provenance"` and field key `"git_commit"` across all 14, so a fleet-wide log query finds every service the same way.

## 5. `verify-agent-images` widening (candidate 4) — makefile:2100-2111

Append three read-only stanzas after the existing four checks, keeping the target's all-`@`, fail-soft style (`|| echo`), never failing the target:

```make
	@echo "$(CYAN)Local image provenance (docker labels):$(NC)"
	@docker image inspect $(REGISTRY)/agent-chassis:$(IMAGE_TAG) \
		--format 'agent-chassis:$(IMAGE_TAG)  revision={{index .Config.Labels "org.opencontainers.image.revision"}}  created={{.Created}}' \
		2>/dev/null || echo "agent-chassis:$(IMAGE_TAG) not present locally — label check skipped"
	@echo "$(CYAN)Pod binary provenance (agent-chassis):$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) exec deploy/agent-chassis -- \
		sh -c 'strings /app/agent-chassis | grep -E "^[0-9a-f]{40}$$|-tree$$" | head -1' 2>/dev/null \
		|| echo "no provenance stamp in pod binary (image predates the 153 fix)"
	@echo "$(CYAN)Pod imageID + startTime:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get pods -l app=agent-chassis \
		-o custom-columns=NAME:.metadata.name,IMAGE:.spec.containers[0].image,IMAGEID:.status.containerStatuses[0].imageID,STARTED:.status.startTime 2>/dev/null || true
```

Notes: `$$` escaping for the grep anchors is required inside make; use `head -1` rather than `grep -m1` (busybox grep's `-m` support is config-dependent — unverified, so don't depend on it); the 40-hex extraction pattern is safe *today* because the measured count of 40-hex strings in the binary is 0, so post-fix the only match is our stamp — the authoritative check remains grepping for the *expected* sha, this stanza is a display aid. This detects; it does not gate (gating is candidate 2/3, deferred).

## 6. Rollout and test plan (one session, one sitting)

**Order of operations** (house rules: commit per task with explicit pathspec; `make build-*` builds committed HEAD, so nothing can be ref-built before it is committed):

1. **Council submission first** (§7) — `pkg/` is gated; use the `Council-Submitted:` trailer path so commits are not held hostage to the verdict (per the 097 script's own header, 2026-07-30 rule).
2. **Commit 1:** `pkg/buildinfo/buildinfo.go` (new). Local compile check: `go build ./pkg/buildinfo`.
3. **Commit 2:** the 14 `cmd/*/main.go` edits. Check: `go build ./cmd/...` (all 14 + the other cmds must still compile), `go vet ./cmd/...`.
4. **Commit 3:** the 14 Dockerfile edits + the two makefile defines + the `verify-agent-images` widening.
5. **Pre-deploy local smoke (optional, pre-commit-capable):** `make build-agent-chassis-tree` with a scratch `IMAGE_TAG`, then `docker image inspect` (expect `revision=<shortsha>-tree`) and `docker run --rm --entrypoint /bin/sh <image> -c 'strings /app/agent-chassis | grep -- -tree'` (expect the marker). This exercises the ldflags/ARG plumbing without touching the cluster and simultaneously proves the tree-marker rule.
6. **Pilot = agent-chassis** (most-referenced, most-rebuilt, the bug's own worked example). Bump `IMAGE_TAG` at `makefile:17` (currently `v1.0.1279` — take whatever is current at implementation time, another session may have rolled it), commit the bump, then the full cycle with individual targets (not `quick-agent-update`, which drags image-generator-adapter along):
   - `make build-agent-chassis` → build output should echo the short sha as before; local `docker image inspect` shows the full-sha revision label.
   - **Regression guard (before pushing anything):** `make build-agent-chassis REF=<some-older-commit> IMAGE_TAG=<scratch-tag>` must stamp *that* commit's sha (check the label + `docker run … strings`), not HEAD. Delete the scratch tag afterwards.
   - Push + deploy via the normal targets, then verify at the pod:
     `kubectl -n ai-persona-system exec <pod> -- sh -c 'strings /app/agent-chassis | grep -c <full-sha-of-built-ref>'` → must be ≥1, and `verify-agent-images` must print the same sha in its new stanza.
7. **Induced fault (the discriminating test, from the bug file):** bump `IMAGE_TAG` again, run push+deploy **without** build. The pod must come up still reporting the *previous* sha while wearing the new tag — i.e., the retag is now **visible** (candidate 1 detects; it does not refuse — refusing is deferred candidate 2/3). Immediately run the full positive-control cycle (`build` + push + deploy on a further bump) so production does not remain on a lying tag, and record both observations in the bug file.
8. **The other 13 services:** done = Dockerfile + main.go edits committed and compiling. They are NOT rebuilt/rolled this session — each picks up provenance automatically on its next normal roll (the makefile side is already live for them). Add a 14-row checklist to `bugs_open/153` recording which services have rolled with a stamp; per the `/bugs_closed/` bar ("fixed AND live"), 153 stays OPEN with a status note until the fleet has rolled — the pilot proves the mechanism, the checklist tracks liveness.
9. **Docs, same session:** (a) status + checklist + deferred-candidates note appended to `bugs_open/153...`; (b) register the mechanism in `docs/agent_docs/docs026_concept_register/register/build-pipeline.md` as a new BLD entry (last is BLD-017 today; take the next free number at write time); (c) update CLAUDE.md's "Building & deploying images" pod-grep line (line ~474-475) to present the sha grep as the primary check with marker-grep as fallback for pre-fix images — this is the practice the fix exists to retire.

**What can go wrong, planned for:** another session bumps `IMAGE_TAG` or edits the makefile concurrently — rebase, re-read lines 113-144 before applying; the defines are stable-shaped, the edit is local to two stanzas. A service's main.go may have drifted since today's survey — the edit is one line, re-anchor on `defer …Sync()` per file at apply time.

## 7. Council submission shape

Gate trigger: new `pkg/buildinfo` package (pkg/ is in the 2026-07-17 scope ruling). Submit via `docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>`; commit with `Council-Submitted: <correlation>`.

- **rationale (sketch):** "Bug 153: the build/deploy contract's second half — 'provenance verified against the running pod' — is unperformable because images and binaries carry zero provenance (measured: 0 shas, 0 version strings in the live v1.0.1279 binary; no ldflags/ARG/LABEL anywhere under build/docker/backend/). Per-fix marker hunting substitutes for it and produced three false conclusions in one day (153's CONTRIBUTION section). This change stamps the full commit sha ref_build already computes (makefile:119) into every backend binary (ldflags → pkg/buildinfo.GitCommit) and image (OCI revision label at the docker CLI), making 'what is running?' an exact check. Additive and inert-until-read; per the 2026-07-29 owner ruling this goes through the normal gate, not an RFC."
- **plan.edits (6 of ≤8):**
  1. `pkg/buildinfo/buildinfo.go` — add — the stamp target; sketch = the package in §1.
  2. `makefile` symbol `ref_build` — modify — compute full sha once, pass `--build-arg GIT_COMMIT` + two OCI labels; sketch = §2.
  3. `makefile` symbol `tree_build` — modify — stamp `<shortsha>-tree` unconditionally so WIP images can never impersonate commit state; sketch = §2.
  4. `build/docker/backend/agent-chassis.dockerfile` (representative; the other 13 named in the sketch) — modify — `ARG GIT_COMMIT=unknown` + ldflags clause on the existing go build line; sketch = §3 table.
  5. `cmd/agent-chassis/main.go` (representative; other 13 named in sketch) — modify — import buildinfo + one `Info("build provenance", …)` line; load-bearing because `-X` on an unlinked package is silently ignored.
  6. `makefile` symbol `verify-agent-images` — modify — three fail-soft provenance stanzas; sketch = §5.
- **Validator caveat:** `diagnose_persist_fix_plan` validates repo-relative paths and refuses no-op edits (unverified whether it accepts one edit standing for 14 files) — if it balks, keep the representative-file form and enumerate the other 13 inside the sketch text, which stays within the 8-edit cap.
- **grounded_in:** quote `makefile:106-108` (the contract comment), `makefile:119` (sha computed and discarded), the 153 measurement block (`0 shas, 0 version strings`), and the CONTRIBUTION section's "three worked examples" sentence.
- **risks:** docker build-arg warnings on the two CronJob images; a Dockerfile whose go build line drifted since survey; `-X` silently no-oping if a main.go edit is skipped (mitigated by the pilot's pod-level positive control).
- **Consumers to name (owner ruling point 3, name them, don't just measure zero collision):** every session following CLAUDE.md "Building & deploying images" (the pod-grep practice changes from marker-hunt to sha-grep); the `component-render-check` (CGV-030) and `shared-output-fields-check` (RFC_012(d)) CronJob workstreams — their images gain the label but their binaries stay unstamped this round; the `quick-agent-update`, `release`/`release-backend` meta-targets (inherit silently); the 104/138/144 lanes whose handoffs still prescribe marker greps; whoever next rolls `IMAGE_TAG` (their verification recipe just got exact). Notify via the bug-file update + the concept-register entry; no code change lands on any of them.

## 8. Explicit non-goals (deferred, not forgotten)

- **Candidate 2 (tag implies build — sha-suffixed tags or push-refusal on label mismatch):** deferred. Changes the tag-naming/push contract fleet-wide (every kustomization `newTag`, `agent_definitions.image_tag`, dashboards); needs explicit owner sign-off. Candidate 1's stamp is its prerequisite anyway — the label a refusal would compare against now exists.
- **Candidate 3 (build-stamp file gating push):** deferred. There is no `push-%` pattern rule — 14 hand-written push lines inside one `push-backend` target (`makefile:269-284`) would need restructuring; costlier than the value this round, and partially subsumed if candidate 2 ever lands.
- **`/buildinfo` HTTP endpoint via `platform/health`:** deferred (§1) — non-uniform per-service wiring; a future one-line change inside `health.Server.Start()` is the right shape if wanted.
- **Non-service Go images** (`migrator`, `seeder`, `tools-api`, `workflow-monitor`, `platform.dockerfile`, the two CronJob checks): not built via the 14-service path or deliberately out of scope; they inherit the CLI label where they use `ref_build`, and the ldflags pattern is available to them opt-in later.
- **Do not** remove or weaken the `IMAGE_TAG`-bump-per-build requirement (bug file's explicit landmine: a same-tag rebuild ships the node's stale cached binary).

### Critical Files for Implementation
- /home/ant/projects/agentchassis/makefile (ref_build 114-131, tree_build 135-144, verify-agent-images 2100-2111, IMAGE_TAG line 17)
- /home/ant/projects/agentchassis/pkg/buildinfo/buildinfo.go (new — the stamp target)
- /home/ant/projects/agentchassis/build/docker/backend/agent-chassis.dockerfile (pilot; pattern for the other 13)
- /home/ant/projects/agentchassis/cmd/agent-chassis/main.go (pilot import+log; pattern for the other 13)
- /home/ant/projects/agentchassis/bugs_open/153_HANDOFF_2026-07-29_image_tag_does_not_imply_a_rebuild.md (status, checklist, deferred-candidates note)
