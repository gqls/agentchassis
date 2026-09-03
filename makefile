# Comprehensive Makefile for Agent-Managed Microservices
# date; make release redeploy-agents  ENVIRONMENT=production REGION=uk001; date

# Load environment variables from .env file
include .env
export

export TMPDIR := $(HOME)/kind-tmp

# Project variables - project name is also used for namespaces
PROJECT_NAME := ai-persona-system
ENVIRONMENT ?= production
REGION ?= uk001
REGION_PATH ?= uk_001
REGISTRY ?= docker.io/aqls
#IMAGE_TAG ?= latest
# Bumped 2026-08-17: v1.0.1305 was REUSED. The cluster still serves the cached
# digest for that tag (running sha256:f90a7e88…, built from 6a782274b on 08-16
# 21:53Z) while the locally built v1.0.1305 (sha256:6039e19c…, from 89a0cbeb7)
# carries 252 newer commits, 24 of them touching platform/internal/pkg. A
# same-tag re-release re-serves the cache, so the ONLY remedy is a new tag.
IMAGE_TAG ?= v1.0.1359

# Paths
TERRAFORM_DIR := deployments/terraform/environments/$(ENVIRONMENT)/$(REGION)
KUSTOMIZE_DIR := deployments/kustomize
SCRIPTS_DIR := scripts

# Colors for output
YELLOW := \033[1;33m
GREEN := \033[1;32m
RED := \033[1;31m
NC := \033[0m # No Color

# Default target
.DEFAULT_GOAL := help

ifeq ($(ENVIRONMENT),production)
    OVERLAY_PATH := $(ENVIRONMENT)/$(REGION_PATH)
else
    OVERLAY_PATH := $(ENVIRONMENT)
endif

#################################
# Release enumeration — ONE declaration per set (bugs_open/237)
#
# There used to be FOUR separately hand-maintained lists of "the release's
# services" (deploy-agents' per-service blocks, update-kustomization-images'
# for-list, redeploy-agents' restarts, push-backend's pushes), plus a fallback
# that GUESSED a service's image name from its service name. A service present
# in the filesystem but absent from a list was silently never rolled:
# render-audit-adapter sat 86 tags behind the fleet (v1.0.1194 vs v1.0.1280)
# with a live credential leak (bugs_open/233) because of exactly that. Same
# pattern as bugs_open/066's four unscoped UPDATEs: now there is ONE
# declaration and the targets loop over it.
#
# RELEASE_IMAGES — the images build-backend/push-backend produce at
# $(IMAGE_TAG). An image NOT in this list is never retagged by a release.
#
# ⚠ CORRECTED 2026-08-18 (OWNER RULING, bugs_open/237 Decision B): this comment
# used to end "(check services, github runners, ollama: own lineage, own deploy
# cadence)". For ollama that is still true — it is an upstream image. For the
# four check services and the two github runners it was NOT a cadence, it was a
# FREEZE, and the owner has ruled all six into the fleet release. What the
# separate lineage actually bought:
#   - the four checks each had a deploy target nobody ran, so each overlay was
#     written once on the day the service was created and never again. Because
#     these binaries compile the estate's action registry IN, a frozen image is a
#     frozen INVENTORY: measured 2026-08-18, removed-config-keys-check could see
#     165 of 169 registered actions and shared-output-fields-check 161 — and a
#     check that skips an action emits nothing, so its clean report is
#     indistinguishable from a clean estate;
#   - github-actions-runner sat on v1.0.948 since April, missing rsync and ssh
#     (measured at the pod), and github-actions-runner-vmsites had NO retag
#     target in this file at all.
# Folding them in also closes check-release-coverage's blind spot by
# construction: the gate only polices overlays pinning a RELEASE_IMAGES image,
# so until now it could not see any of these six.
#
# ⚠ AND THE BLIND SPOT IS SELF-REPRODUCING — TWO MORE HAD ALREADY FALLEN INTO IT
# (added 2026-08-22, under the same owner ruling, not as a new decision):
# `optional-explicit-wires-check` (created 08-21) and `commit-sha-exposure-check`
# (created 08-22) were both born OUTSIDE these lists, exactly as the original four
# were. Nothing said so, and nothing could: the coverage gate only polices overlays
# pinning an image that is ALREADY in RELEASE_IMAGES, so a check service omitted at
# birth is invisible to the very mechanism that exists to catch the omission. Both
# run the `config-key-audit` binary, which compiles the estate's action registry IN
# — so for these two a frozen image is a frozen INVENTORY in the precise sense
# measured on 2026-08-18, and their clean reports would slowly stop meaning what
# they say. **A NEW CHECK SERVICE MUST BE ADDED HERE IN THE COMMIT THAT CREATES IT.**
RELEASE_IMAGES := auth-service core-manager agent-chassis reasoning-agent \
	web-search-adapter web-scrape-adapter git-adapter image-generator-adapter \
	thunder-adapter analyser-adapter browser-runner-adapter \
	content-creator-agent remote-job-spawner kafka-scheduler \
	component-render-check shared-output-fields-check \
	removed-config-keys-check verifier-remit-check \
	loop-sitewide-item-key-check brief-negation-check content-loss-check \
	github-actions-runner \
	optional-explicit-wires-check commit-sha-exposure-check \
	capped-schedule-ordering-check component-source-vocabulary-check \
	live-declaration-drift-check finding-code-registry-check \
	ungraded-completions-check render-truncation-check \
	template-input-field-check

# AGENT_DEPLOY_SERVICES — what deploy-agents retags and applies. Entry form is
# <service>[:<image>]; the image defaults to the service name. A service that
# runs ANOTHER service's binary declares that image here — visibly, in one
# place — instead of relying on a buried special case:
#   render-audit-adapter runs the browser-runner image (different topic and
#     consumer group; it deliberately has NO build/push of its own, and none
#     should be added);
#   vet-intel and business-intel run the agent-chassis image;
#   github-actions-runner-vmsites runs the github-actions-runner image — ONE
#     image, two Deployments, which is why -vmsites never needed a build or a
#     retag target of its own and why it went four months with no way to move it
#     (bugs_open/237 Decision B, owner ruling 2026-08-18).
# ORDER is preserved by the loops; render-audit-adapter sits directly after
# browser-runner-adapter because the two must move together — the tag applied
# to it has to be one the browser-runner was actually built at. The same is now
# true of the two runners, for the same reason.
AGENT_DEPLOY_SERVICES := agent-chassis reasoning-agent web-search-adapter \
	web-scrape-adapter git-adapter image-generator-adapter thunder-adapter \
	analyser-adapter browser-runner-adapter \
	render-audit-adapter:browser-runner-adapter content-creator-agent \
	remote-job-spawner kafka-scheduler vet-intel:agent-chassis \
	business-intel:agent-chassis \
	component-render-check shared-output-fields-check \
	removed-config-keys-check verifier-remit-check \
	loop-sitewide-item-key-check brief-negation-check content-loss-check \
	optional-explicit-wires-check commit-sha-exposure-check \
	capped-schedule-ordering-check component-source-vocabulary-check \
	live-declaration-drift-check finding-code-registry-check \
	ungraded-completions-check render-truncation-check \
	template-input-field-check \
	github-actions-runner github-actions-runner-vmsites:github-actions-runner

# RETAG_EXEMPT — overlays that pin a RELEASE_IMAGES image but are retagged by
# their OWN deploy path, named here so check-release-coverage can hold them to
# it. Entry form is <service>:<the make target that retags it>.
RETAG_EXEMPT := auth-service:deploy-auth-service core-manager:deploy-core-manager

# OWN_LINEAGE — the ONLY way an overlay may pin one of OUR images and stay out
# of the release. Entry form is <service>:<the make target that retags it>.
#
# ⚠ EMPTY, AND ITS EMPTINESS IS THE POINT (added 2026-08-22, bugs_closed/318).
# Until today, "the release does not build this image" was the coverage gate's
# ADMISSION TEST — it skipped such an overlay entirely — so a service omitted at
# birth was not uncovered, it was out of scope, and the gate printed "Release
# coverage OK" about the exact omission it exists to catch. Eight services fell
# into that hole, two of them AFTER the owner ruling meant to close it. The
# predicate is now inverted: one of our images that no release builds is a
# VIOLATION, and this list is the explicit, greppable, reviewable way out. It is
# opt-in with the unsafe side default OFF (CLAUDE.md owner ruling 2026-08-02 §2),
# and it must stay a LIST rather than becoming a rule — a predicate that guesses
# which services are legitimately outside the release is one nobody can review.
#
# ⚠ IT WAS GOING TO SHIP EMPTY, AND THE GATE'S FIRST RUN FOUND AN ENTRY FOR IT.
# `admin-dashboard` pins `$(REGISTRY)/admin-dashboard`, which `build-backend`
# does not build — it is a frontend, built by `build-dashboard` from
# `frontends/admin-dashboard/` in the WORKING TREE with no `ref_build`, no
# `REF` and no provenance stamp (BLD-019/BLD-020 scope frontends out
# explicitly). It is nonetheless released: `release-dashboard` is the last goal
# of `pinned_sweep` in `release`. So it is genuinely own-lineage and genuinely
# covered — but until now NOTHING ON DISK SAID SO, and the old gate could not
# have said so, because its image is not in RELEASE_IMAGES and that was the
# gate's admission test. This is the shape bugs_closed/318 is about, found on the
# new predicate's first live run rather than by anyone remembering.
# ⚠ NOTE what the entry does NOT cover: `release-backend` (the no-dashboard
# variant) omits `release-dashboard`, so a backend-only release leaves this
# service on its old tag. That is intended, and it is why the target is named
# here rather than the fact being left to folklore.
OWN_LINEAGE := admin-dashboard:deploy-dashboard

# check-release-coverage (the door-closer for bugs_open/237's class, widened by
# bugs_closed/318): every overlay on disk that pins one of OUR images must be in a
# release path or explicitly exempt, or the release refuses to run.
#
# ⚠ THE PREDICATE MOVED TO GO (2026-08-22) AND THE SHELL COPY WAS DELETED, not
# left beside it — two implementations of one gate is the drift class
# 099_SYNC_gate_roster.py exists for. Two reasons it is Go:
#   1. SCOPE. scripts/council-scope.sh scopes review to platform/, internal/,
#      pkg/ and appliable migrations. A makefile-only gate cannot be reviewed at
#      all (this one never was); the predicates now live in pkg/releaseset and
#      are.
#   2. PROOF. On 2026-08-22 a session mutated THIS FILE in place to show the old
#      gate discriminates, and another session committed it inside the window
#      (WRONG_CALLS.md, f016b07ec). The Go predicates are pure functions, so the
#      mutation proofs are table rows over testdata and no shared file is
#      touched to run them.
# The Go version also fixes two blind spots this recipe had: it walks
# overlays/production to ANY depth (tools-api's real overlay has no region
# directory and this glob could never see it) and it reads `newName` in
# preference to `name`, which is kustomize's own semantics.
.PHONY: check-release-coverage
check-release-coverage: ## Fail if an overlay pins one of our images but is in no release path
	@go run ./cmd/releasecheck --root . --registry $(REGISTRY)

# release-census — the CLUSTER half, and it asks a DIFFERENT question from the
# gate above. Do not read one as the other.
#
#   check-release-coverage  reads the FILESYSTEM: "can a release reach this
#                           service?" It is preventive and it REFUSES.
#   release-census          reads the CLUSTER: "does what is running match what
#                           is declared?" It is a detector and it only REPORTS.
#
# It exists because the filesystem and the cluster are two enumerations and
# NEITHER IS A SUPERSET OF THE OTHER — measured 2026-08-22, one service declared
# everywhere and running nowhere (`capped-schedule-ordering-check`), and two
# running as CronJobs with no overlay on disk. No filesystem gate can see either,
# in either direction, ever.
#
# ⚠ HAND-RUN ONLY, deliberately (bugs_closed/318 phase 1). There is no CronJob, no
# RBAC and no doc_notes row yet — so nothing runs this unless a person does, and
# it must NOT be described as a live detector. Scheduling it is a separate
# decision with its own round, and this estate's own evidence is why the two were
# split: "detection works; SCHEDULE and DISPATCH do not."
#
# Read-only: it LISTS deployments, cronjobs and daemonsets and nothing else.
.PHONY: release-census
release-census: ## Report cluster-vs-declaration drift (read-only; hand-run, no CronJob)
	@go run ./cmd/releasecheck --census --root . --registry $(REGISTRY) --namespace $(PROJECT_NAME)

# print-* — echo-only, and they exist for pkg/releaseset/decl_parity_test.go:
# the Go side reads these declarations with a literal extractor rather than a
# make evaluator, and the parity test asks MAKE for the same lists so the two
# readings cannot drift. Handy by hand too (LANDMINES already reaches for the
# first of them).
.PHONY: print-release-images print-agent-deploy-services print-retag-exempt print-own-lineage
print-release-images: ## Print RELEASE_IMAGES (one line, space separated)
	@echo '$(RELEASE_IMAGES)'
print-agent-deploy-services: ## Print AGENT_DEPLOY_SERVICES
	@echo '$(AGENT_DEPLOY_SERVICES)'
print-retag-exempt: ## Print RETAG_EXEMPT
	@echo '$(RETAG_EXEMPT)'
print-own-lineage: ## Print OWN_LINEAGE (empty today, deliberately)
	@echo '$(OWN_LINEAGE)'

#################################
# Help
#################################
.PHONY: help
help: ## Show this help message
	@echo '$(YELLOW)Personae System - Makefile Commands$(NC)'
	@echo ''
	@echo 'Usage:'
	@echo '  make $(GREEN)<target>$(NC) $(YELLOW)[ENVIRONMENT=production] [REGION=uk001] [IMAGE_TAG=latest]$(NC)'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-30s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

#################################
# Development Environment
#################################
.PHONY: dev-up
dev-up: ## Start local development environment
	@echo "$(YELLOW)Starting local development environment...$(NC)"
	docker-compose -f deployments/docker-compose/docker-compose.yaml up -d

.PHONY: dev-down
dev-down: ## Stop local development environment
	@echo "$(YELLOW)Stopping local development environment...$(NC)"
	docker-compose -f deployments/docker-compose/docker-compose.yaml down

.PHONY: dev-logs
dev-logs: ## Show logs from development environment
	docker-compose -f deployments/docker-compose/docker-compose.yaml logs -f

.PHONY: dev-reset
dev-reset: dev-down ## Reset development environment (removes volumes)
	@echo "$(YELLOW)Resetting development environment...$(NC)"
	docker-compose -f deployments/docker-compose/docker-compose.yaml down -v

#################################
# Building
#################################
.PHONY: build-all
build-all: build-backend build-frontends ## Build all images

# build-backend MUST produce every image in RELEASE_IMAGES. push-backend loops
# that list and exits 1 on the first missing image, so an image declared there
# and not built here fails the release loudly — but the runners are the case
# where that matters most: deploy-agents now retags BOTH runner overlays to
# $(IMAGE_TAG), so a release that retags without building would point them at an
# image nobody pushed and both would ImagePullBackOff together. The build half
# and the retag half must land in the same release
# (bugs_open/237 Decision B, owner ruling 2026-08-18).
#
# ⚠ IT IS DERIVED FROM `RELEASE_IMAGES`, NOT HAND-LISTED (changed 2026-08-22,
# bugs_closed/318). Until today this was a hand-written list of six group targets,
# so "build-backend builds exactly RELEASE_IMAGES" was an invariant two separate
# enumerations had to agree on with nothing keeping them in step — the same shape
# as the four hand-maintained deploy lists `bugs_open/237` removed. BLD-022 §(iv)
# recorded it as "verified by set equality 2026-08-18 [MEASURED] and policed by
# NOTHING", and it was **false four days later**: `optional-explicit-wires-check`,
# `commit-sha-exposure-check` and `capped-schedule-ordering-check` were all added
# to `RELEASE_IMAGES` (and to `AGENT_DEPLOY_SERVICES`) and to none of the build
# groups, by two different lanes on 2026-08-22. Measured the same day: 25
# declared, 22 built, and locally `optional-explicit-wires-check` existed only at
# `v1.0.1321` while `capped-schedule-ordering-check` had never been built at any
# tag — so the next `make release` would have built 22 images and then died on the
# first `docker push` of an image nobody built, BEFORE `deploy-core`, deploying
# nothing at all. The `$(addprefix …)` makes that state unrepresentable rather
# than merely detectable: there is now one list, and the build set IS it.
#
# THE NAMING CONTRACT this relies on: every entry in `RELEASE_IMAGES` has a
# `build-<image>` target. That holds for all 25 today; `github-actions-runner`'s
# real target is the older `build-github-runner`, so the alias below carries it.
# A new image whose target does not follow the convention fails the release with
# "No rule to make target" at the very first build — loudly, and before anything
# is pushed or deployed.
#
# `build-agents` / `build-adapters` / `build-checks` survive as convenience
# targets for a human building one group by hand. They are no longer
# load-bearing, and adding a service to one of them does NOT put it in a release
# — `RELEASE_IMAGES` does, and `check-release-coverage` polices the other side.
.PHONY: build-backend
build-backend: $(addprefix build-,$(RELEASE_IMAGES)) ## Build all backend services (== RELEASE_IMAGES, by construction)

# Alias: the image is `github-actions-runner`, the target predates the naming
# convention. Named here so the derivation above needs no special case.
.PHONY: build-github-actions-runner
build-github-actions-runner: build-github-runner ## Alias for build-github-runner (image name == target name)

# build-checks — the four daily CronJob images. Folded into the release
# 2026-08-18: each previously had a deploy target nobody ran, and because these
# binaries compile the action registry IN, a frozen image silently under-reports
# on the estate it audits (see RELEASE_IMAGES above for the measured figures).
.PHONY: build-checks
build-checks: build-component-render-check build-shared-output-fields-check build-removed-config-keys-check build-verifier-remit-check build-loop-sitewide-item-key-check build-brief-negation-check build-content-loss-check ## Build all seven daily check CronJob images

.PHONY: build-frontends
build-frontends: build-admin-dashboard build-user-portal build-agent-playground ## Build all frontend applications

#################################
# Deploy blast radius (multi_session_coordination HANDOFF 2026-07-16 §3/§7.3)
#
# Many sessions share this ONE working tree. `docker build ... .` sends the
# whole tree as context, so a working-tree image bundles every session's
# uncommitted, untested, mid-edit code — not just yours.
#
# So the DEFAULT is inverted (2026-07-17): build-<service> builds from the
# committed ref $(REF) (HEAD unless pinned) via `git archive` into a clean
# context — it CANNOT bundle anyone's WIP. Your commit is what ships: commit
# your task first (explicit pathspec — see CLAUDE.md), then build.
#
#   make build-<service>            committed HEAD               (safe default)
#   make build-<service> REF=<ref>  a pinned commit
#   make build-<service>-tree       whole working tree, all WIP  (opt-in only)
#
# Failure direction is deliberate: forget to commit and the ref build omits
# YOUR change (a wasted cycle, caught by the pod-grep check) rather than
# silently shipping everyone ELSE'S untested change to production.
#
# push-*/deploy-* are git-blind — they ship whatever is tagged $(IMAGE_TAG).
# Provenance is got right HERE, at build time, and verified against the running
# pod (never git, never the tag).
#################################

REF ?= HEAD

# ref_build,<service> — committed-state build from $(REF). No WIP can enter it.
define ref_build
@test -f build/docker/backend/$(1).dockerfile || \
	{ echo "$(RED)No build/docker/backend/$(1).dockerfile — ref builds cover backend services only (frontends build from frontends/<app>).$(NC)"; exit 1; }
@git rev-parse --verify --quiet '$(REF)^{commit}' >/dev/null || \
	{ echo "$(RED)REF='$(REF)' is not a commit — ref builds must name committed state.$(NC)"; exit 1; }
@echo "$(GREEN)Building $(1) from committed ref $(REF) = $$(git rev-parse --short $(REF)) — working tree NOT included.$(NC)"
@UNSHIPPED=$$(git status --porcelain 2>/dev/null | wc -l); \
if [ "$$UNSHIPPED" -gt 0 ] && [ "$(REF)" = "HEAD" ]; then \
	echo "$(YELLOW)  $$UNSHIPPED uncommitted change(s) are NOT in this image. Commit anything that belongs in it, then rebuild:$(NC)"; \
	git status --porcelain | head -15; \
	if [ "$$UNSHIPPED" -gt 15 ]; then echo "  ... ($$UNSHIPPED total)"; fi; \
fi
@CTX=$$(mktemp -d /tmp/ref-ctx-$(1).XXXXXX) && \
trap 'rm -rf "$$CTX"' EXIT && \
GIT_COMMIT=$$(git rev-parse '$(REF)^{commit}') && \
git archive $(REF) | tar -x -C "$$CTX" && \
docker build -t $(REGISTRY)/$(1):$(IMAGE_TAG) \
	--build-arg GIT_COMMIT=$$GIT_COMMIT \
	--label org.opencontainers.image.revision=$$GIT_COMMIT \
	--label org.opencontainers.image.created=$$(date -u +%Y-%m-%dT%H:%M:%SZ) \
	-f "$$CTX/build/docker/backend/$(1).dockerfile" "$$CTX"
endef

# tree_build,<service> — the escape hatch: build from the WORKING TREE, WIP and
# all. Only when you deliberately want uncommitted code in the image.
define tree_build
@echo "$(RED)Building $(1) from the WORKING TREE — this image bundles uncommitted WIP from ALL sessions, not just yours.$(NC)"
@DIRTY=$$(git status --porcelain 2>/dev/null | wc -l); \
if [ "$$DIRTY" -gt 0 ]; then git status --porcelain | head -25; \
	if [ "$$DIRTY" -gt 25 ]; then echo "  ... ($$DIRTY total)"; fi; \
	echo "$(RED)  ^ all of the above will be in the image. For a committed image instead: make build-$(1)$(NC)"; \
fi
GIT_COMMIT="$$(git rev-parse --short HEAD 2>/dev/null || echo unknown)-tree" && \
docker build -t $(REGISTRY)/$(1):$(IMAGE_TAG) \
	--build-arg GIT_COMMIT=$$GIT_COMMIT \
	--label org.opencontainers.image.revision=$$GIT_COMMIT \
	--label org.opencontainers.image.created=$$(date -u +%Y-%m-%dT%H:%M:%SZ) \
	-f build/docker/backend/$(1).dockerfile .
endef

# Every backend service gets both alternatives for free via two pattern rules.
# (build-<service> below already == the committed-ref build at REF=HEAD; -ref is
# kept as an explicit alias, mainly for passing REF=<ref>.)
build-%-ref: ## committed-ref build (alias of build-<service>): make build-agent-chassis-ref [REF=<ref>]
	$(call ref_build,$*)

build-%-tree: ## WORKING-TREE build — bundles all WIP, opt-in: make build-agent-chassis-tree
	$(call tree_build,$*)

# Backend services — all build from committed HEAD by default (REF=<ref> to pin,
# build-<service>-tree for a working-tree image).
.PHONY: build-auth-service
build-auth-service: ## Build auth-service (committed HEAD; REF=<ref> to pin, -tree for WIP)
	$(call ref_build,auth-service)

.PHONY: build-core-manager
build-core-manager: ## Build core-manager (committed HEAD; REF=<ref> to pin, -tree for WIP)
	$(call ref_build,core-manager)

.PHONY: build-agent-chassis
build-agent-chassis: ## Build agent-chassis (committed HEAD; REF=<ref> to pin, -tree for WIP)
	$(call ref_build,agent-chassis)

.PHONY: build-reasoning-agent
build-reasoning-agent: ## Build reasoning-agent (committed HEAD; REF=<ref> to pin, -tree for WIP)
	$(call ref_build,reasoning-agent)

.PHONY: build-web-search-adapter
build-web-search-adapter: ## Build web-search-adapter (committed HEAD; REF=<ref> to pin, -tree for WIP)
	$(call ref_build,web-search-adapter)

# Not a service — a CronJob image (CGV-030). It ships the component-render-check
# binary AND the baseline it is measured against, so the pair can never disagree.
# Committed-HEAD build matters more than usual here: a working-tree build could
# bake in an uncommitted baseline, i.e. a silenced finding with no diff to review.
.PHONY: build-component-render-check
build-component-render-check: ## Build component-render-check CronJob image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,component-render-check)

# Not a service — a CronJob image (RFC_012 (d), the ONLINE half). It ships the
# config-key-audit binary AND the ack list it ratchets against, for the same
# reason component-render-check ships its baseline: the ack list is the check's
# own definition of "already known", so a working-tree build could bake in an
# unreviewed acknowledgement — a silenced finding with no diff. Committed-HEAD
# build makes that unrepresentable rather than merely discouraged.
.PHONY: build-shared-output-fields-check
build-shared-output-fields-check: ## Build shared-output-fields-check CronJob image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,shared-output-fields-check)

# Same binary as the other config-key-audit checks; different CMD. Flags loop-
# nested create_work_item steps whose item_key is still per-site, so every
# iteration after the first is silently dropped (bugs_open/321).
.PHONY: build-loop-sitewide-item-key-check
build-loop-sitewide-item-key-check: ## Build loop-sitewide-item-key-check CronJob image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,loop-sitewide-item-key-check)

.PHONY: build-template-input-field-check
build-template-input-field-check: ## Build template-input-field-check CronJob image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,template-input-field-check)

# Ships the SAME Go binary the offline audit uses, so the scheduled check walks
# workflow steps with validation.WalkSteps rather than a re-implementation
# (council round 2, corr 3eb0d1f1, reuse_agent gating objection).
# Not a service — a CronJob image (RFC_029 §10.15, CTS-060(5)). Same
# config-key-audit binary as its siblings, different CMD, PLUS the acks file it
# gates on. Committed-HEAD build is load-bearing for the same reason as
# shared-output-fields-check: the acks list is the check's own definition of
# "already known", so a working-tree build could bake in an unreviewed
# acknowledgement — a silenced finding with no diff to review.
.PHONY: build-optional-explicit-wires-check
build-optional-explicit-wires-check: ## Build optional-explicit-wires-check CronJob image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,optional-explicit-wires-check)

# Same family, same argument: the STANDING form of migration 537's apply-time
# guard (bugs_closed/334) ships its acks file with the binary, so the
# committed-HEAD build is what keeps an unreviewed exception out of the image.
.PHONY: build-commit-sha-exposure-check
build-commit-sha-exposure-check: ## Build commit-sha-exposure-check CronJob image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,commit-sha-exposure-check)

.PHONY: build-ungraded-completions-check
build-ungraded-completions-check: ## Build ungraded-completions-check CronJob image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,ungraded-completions-check)

.PHONY: build-render-truncation-check
build-render-truncation-check: ## Build render-truncation-check CronJob image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,render-truncation-check)

# bugs_open/363 phase 2. The DECLARATIONS (platform/livespec) are compiled into
# this binary, so a stale image is a stale SPEC and the check would keep reporting
# clean against yesterday's declarations. Committed-HEAD build is what makes an
# unreviewed declaration change unshippable rather than merely discouraged.
.PHONY: build-live-declaration-drift-check
build-live-declaration-drift-check: ## Build live-declaration-drift-check CronJob image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,live-declaration-drift-check)

# bugs_open/358 phase 2. Same config-key-audit binary as its siblings, different
# CMD, PLUS the finding-code registry it grades against. Committed-HEAD build is
# load-bearing here for the usual reason and one extra: the registry says which
# findings the estate has ACCEPTED as human-evidence-only, so a working-tree build
# could bake in an unreviewed disposition — a silenced finding with no diff to
# review. It runs --no-source (the image ships no repo); the two arms that read Go
# source run at commit time instead, via scripts/check-finding-code-registry.sh.
.PHONY: build-finding-code-registry-check
build-finding-code-registry-check: ## Build finding-code-registry-check CronJob image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,finding-code-registry-check)

.PHONY: build-capped-schedule-ordering-check
build-capped-schedule-ordering-check: ## Build capped-schedule-ordering-check CronJob image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,capped-schedule-ordering-check)

.PHONY: build-component-source-vocabulary-check
build-component-source-vocabulary-check: ## Build component-source-vocabulary-check CronJob image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,component-source-vocabulary-check)

.PHONY: build-removed-config-keys-check
build-removed-config-keys-check: ## Build removed-config-keys-check CronJob image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,removed-config-keys-check)

# Not a service — a CronJob image (WII-015, bugs_open/213 owner ruling D3). It
# ships the class detector for the verifier/producer join, LINKED against the live
# verifier registry: which item types have a verifier, and which of those declare a
# remit, are compiled-in facts, so the alternative was a mirrored list in a Python
# job that goes stale exactly when a new verifier lands. Committed-HEAD build
# matters for the usual reason and one extra: the registry it links IS the
# assertion, so an image built from a working tree could report a remit that is not
# on the branch.
.PHONY: build-verifier-remit-check
build-verifier-remit-check: ## Build verifier-remit-check CronJob image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,verifier-remit-check)

.PHONY: build-content-loss-check
build-content-loss-check: ## Build content-loss-check CronJob image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,content-loss-check)

.PHONY: build-brief-negation-check
build-brief-negation-check: ## Build brief-negation-check CronJob image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,brief-negation-check)

.PHONY: build-web-scrape-adapter
build-web-scrape-adapter: ## Build web-scrape-adapter (committed HEAD; REF=<ref> to pin, -tree for WIP)
	$(call ref_build,web-scrape-adapter)

.PHONY: build-git-adapter
build-git-adapter: ## Build git-adapter (committed HEAD; REF=<ref> to pin, -tree for WIP)
	$(call ref_build,git-adapter)

.PHONY: build-image-generator-adapter
build-image-generator-adapter: ## Build image-generator-adapter (committed HEAD; REF=<ref> to pin, -tree for WIP)
	$(call ref_build,image-generator-adapter)

.PHONY: build-thunder-adapter
build-thunder-adapter: ## Build thunder-adapter (committed HEAD; REF=<ref> to pin, -tree for WIP)
	$(call ref_build,thunder-adapter)

.PHONY: build-analyser-adapter
build-analyser-adapter: ## Build analyser-adapter (committed HEAD; REF=<ref> to pin, -tree for WIP)
	$(call ref_build,analyser-adapter)

.PHONY: build-browser-runner-adapter
build-browser-runner-adapter: ## Build browser-runner-adapter (Tier-4 headless, ~1.2GB Chromium+Playwright; committed HEAD, -tree for WIP)
	$(call ref_build,browser-runner-adapter)

.PHONY: build-content-creator-agent
build-content-creator-agent: ## Build content-creator-agent (committed HEAD; REF=<ref> to pin, -tree for WIP)
	$(call ref_build,content-creator-agent)

.PHONY: build-remote-job-spawner
build-remote-job-spawner: ## Build remote-job-spawner (committed HEAD; REF=<ref> to pin, -tree for WIP)
	$(call ref_build,remote-job-spawner)

.PHONY: build-kafka-scheduler
build-kafka-scheduler: ## Build kafka-scheduler (committed HEAD; REF=<ref> to pin, -tree for WIP)
	$(call ref_build,kafka-scheduler)

# Agent targets
.PHONY: build-agents
build-agents: build-agent-chassis build-reasoning-agent build-content-creator-agent build-remote-job-spawner build-kafka-scheduler ## Build all agents

.PHONY: build-adapters
build-adapters: build-web-search-adapter build-web-scrape-adapter build-git-adapter build-image-generator-adapter build-thunder-adapter build-analyser-adapter build-browser-runner-adapter ## Build all adapters

# Frontend applications
# build-admin-dashboard was declared .PHONY here with no recipe — the line below
# it defines deploy-admin-dashboard — so `make build-frontends` died on the first
# prerequisite. The real target is build-dashboard.
.PHONY: build-admin-dashboard
build-admin-dashboard: build-dashboard ## Build admin-dashboard (alias for build-dashboard)

.PHONY: deploy-admin-dashboard
deploy-admin-dashboard: deploy-dashboard ## Deploy admin-dashboard (alias)

.PHONY: build-user-portal
build-user-portal: ## Build user-portal image
	@echo "$(YELLOW)Building user-portal...$(NC)"
	cd frontends/user-portal
	docker build -t $(REGISTRY)/user-portal:$(IMAGE_TAG) \
		-f frontends/user-portal/Dockerfile frontends/user-portal

.PHONY: build-agent-playground
build-agent-playground: ## Build agent-playground image
	@echo "$(YELLOW)Building agent-playground...$(NC)"
	cd frontends/agent-playground
	docker build -t $(REGISTRY)/agent-playground:$(IMAGE_TAG) \
		-f frontends/agent-playground/Dockerfile frontends/agent-playground

#################################
# Push Images
#################################
.PHONY: push-all
push-all: push-backend push-frontends ## Push all images to registry

.PHONY: push-backend
push-backend: ## Push all backend images
	@echo "$(YELLOW)Pushing backend images...$(NC)"
	@for img in $(RELEASE_IMAGES); do \
		echo "$(CYAN)→ docker push $(REGISTRY)/$$img:$(IMAGE_TAG)$(NC)"; \
		docker push $(REGISTRY)/$$img:$(IMAGE_TAG) || exit 1; \
	done

.PHONY: push-frontends
push-frontends: ## Push all frontend images
	@echo "$(YELLOW)Pushing frontend images...$(NC)"
	docker push $(REGISTRY)/admin-dashboard:$(IMAGE_TAG)
	docker push $(REGISTRY)/user-portal:$(IMAGE_TAG)
	docker push $(REGISTRY)/agent-playground:$(IMAGE_TAG)

#################################
# Infrastructure Deployment
#################################
KUBECONFIG_PATH := $(HOME)/.kube/config_$(ENVIRONMENT)_$(REGION)

.PHONY: deploy-cluster-only
deploy-cluster-only: ## Deploy just the Kubernetes cluster
	@echo "$(GREEN)Deploying Kubernetes cluster...$(NC)"
	@cd $(TERRAFORM_DIR)/010-infrastructure && \
		terraform init && \
		terraform apply -auto-approve -var-file=terraform.tfvars.secret

.PHONY: deploy-infrastructure-old
deploy-infrastructure-old: ## Deploy all infrastructure components
	@echo "$(YELLOW)Deploying infrastructure to $(ENVIRONMENT)/$(REGION)...$(NC)"
	@$(MAKE) deploy-010-infrastructure
	@$(MAKE) deploy-020-ingress
	@$(MAKE) deploy-030-strimzi-operator
	@$(MAKE) deploy-040-kafka-cluster
	@$(MAKE) deploy-045-kafka-users
	@$(MAKE) deploy-047-base-configs
	@$(MAKE) deploy-050-storage
	@$(MAKE) deploy-060-databases
	@$(MAKE) deploy-070-database-schemas
	@$(MAKE) deploy-080-kafka-topics
	@$(MAKE) deploy-090-monitoring
	@$(MAKE) deploy-095-node-config
	@$(MAKE) deploy-096-github-runners
	@$(MAKE) deploy-097-ollama

.PHONY: deploy-infrastructure
deploy-infrastructure: ## Deploy all infrastructure components
	@echo "$(YELLOW)Deploying infrastructure to $(ENVIRONMENT)/$(REGION)...$(NC)"
	@echo "$(GREEN)Step 1: Deploying Kubernetes cluster...$(NC)"
	@cd $(TERRAFORM_DIR)/010-infrastructure && \
		terraform init && \
		terraform apply -auto-approve -var-file=terraform.tfvars.secret && \
		terraform output -raw kubeconfig_raw > $(KUBECONFIG_PATH)
	@echo "$(GREEN)Cluster deployed! Using kubeconfig: $(KUBECONFIG_PATH)$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-020-ingress
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-030-strimzi-operator
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-040-kafka-cluster
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-045-kafka-users
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-047-base-configs
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-050-storage
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-060-databases
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-065-pgbouncer
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-070-database-schemas
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-080-kafka-topics
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-090-monitoring
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-095-node-config
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-096-github-runners
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-097-ollama
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-100-bootstrap-agents
	@echo "$(GREEN)Infrastructure deployment complete!$(NC)"
	@echo "$(YELLOW)To use this cluster, run: export KUBECONFIG=$(KUBECONFIG_PATH)$(NC)"

# Add this new target that skips cluster creation
.PHONY: deploy-infrastructure-from-ingress
deploy-infrastructure-from-ingress: ## Deploy infrastructure starting from ingress (assumes cluster exists)
	@echo "$(YELLOW)Deploying infrastructure from ingress for $(ENVIRONMENT)/$(REGION)...$(NC)"
	@echo "$(GREEN)Using existing kubeconfig: $(KUBECONFIG_PATH)$(NC)"
	@if [ ! -f "$(KUBECONFIG_PATH)" ]; then \
		echo "$(RED)Error: Kubeconfig not found at $(KUBECONFIG_PATH)$(NC)"; \
		echo "$(YELLOW)Manually set up Kubeconfig first - export KUBECONFIG=~/.kube/config_$(ENVIRONMENT)_$(REGION)    $(NC)"; \
		exit 1; \
	fi
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-020-ingress
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-030-strimzi-operator
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-040-kafka-cluster
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-045-kafka-users
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-047-base-configs
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-050-storage
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-060-databases
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-065-pgbouncer
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-070-database-schemas
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-080-kafka-topics
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-090-monitoring
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-095-node-config
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-096-github-runners
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-097-ollama
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-100-bootstrap-agents
	@echo "$(GREEN)Infrastructure deployment complete!$(NC)"
	@echo "$(YELLOW)To use this cluster, run: export KUBECONFIG=$(KUBECONFIG_PATH)$(NC)"

# Quick helper for your current situation
.PHONY: continue-deployment
continue-deployment: deploy-infrastructure-from-ingress ## Continue deployment from where cluster creation finished


# Individual infrastructure components
.PHONY: deploy-010-infrastructure
deploy-010-infrastructure: ## Deploy core infrastructure (Kubernetes cluster)
	@echo "$(GREEN)Deploying 010-infrastructure...$(NC)"
	@cd $(TERRAFORM_DIR)/010-infrastructure && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform init && \
			terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform init && \
			terraform apply -auto-approve; \
		fi

# Export KUBECONFIG for all terraform commands in this section
.PHONY: deploy-020-ingress
deploy-020-ingress: ## Deploy ingress controller
	@echo "$(GREEN)Deploying 020-ingress-nginx...$(NC)"
	@cd $(TERRAFORM_DIR)/020-ingress-nginx && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-030-strimzi-operator
deploy-030-strimzi-operator: ## Deploy Strimzi operator
	@echo "$(GREEN)Deploying 030-strimzi-operator...$(NC)"
	@cd $(TERRAFORM_DIR)/030-strimzi-operator && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-040-kafka-cluster
deploy-040-kafka-cluster: ## Deploy Kafka cluster
	@echo "$(GREEN)Deploying 040-kafka-cluster...$(NC)"
	@cd $(TERRAFORM_DIR)/040-kafka-cluster && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-045-kafka-users
deploy-045-kafka-users: deploy-040-kafka-cluster ## Fixed dependency name
	@echo "$(GREEN)Deploying 045-kafka-users...$(NC)"
	cd $(TERRAFORM_DIR)/045-kafka-users && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-047-base-configs
deploy-047-base-configs: ## Deploy base ConfigMaps and Secrets
	@echo "$(GREEN)Deploying 047-base-configs...$(NC)"
	@cd $(TERRAFORM_DIR)/047-base-configs && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-050-storage
deploy-050-storage: ## Deploy S3/storage buckets
	@echo "$(GREEN)Deploying 050-storage...$(NC)"
	@cd $(TERRAFORM_DIR)/050-storage && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-060-databases
deploy-060-databases: ## Deploy database instances
	@echo "$(GREEN)Deploying 060-databases...$(NC)"
	@cd $(TERRAFORM_DIR)/060-databases && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-065-pgbouncer
deploy-065-pgbouncer: ## Deploy PgBouncer connection pooler
	@echo "$(GREEN)Deploying 065-pgbouncer...$(NC)"
	@# Fetch existing DB passwords and create the userlist secret via temp file
	@CLIENTS_PW=$$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get secret personae-platform-secrets -o jsonpath='{.data.CLIENTS_DB_PASSWORD}' | base64 -d) && \
	 TEMPLATES_PW=$$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get secret personae-platform-secrets -o jsonpath='{.data.TEMPLATES_DB_PASSWORD}' | base64 -d) && \
	 ADMIN_PW=$$(openssl rand -base64 16 | tr -d '=/+' | head -c 20) && \
	 TMPFILE=$$(mktemp) && \
	 echo "\"clients_user\" \"$${CLIENTS_PW}\"" > $$TMPFILE && \
	 echo "\"templates_user\" \"$${TEMPLATES_PW}\"" >> $$TMPFILE && \
	 echo "\"pgbouncer_admin\" \"$${ADMIN_PW}\"" >> $$TMPFILE && \
	 KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) create secret generic pgbouncer-userlist \
		--from-file=userlist.txt=$$TMPFILE \
		--dry-run=client -o yaml | KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -f - && \
	 rm -f $$TMPFILE
	@echo "  PgBouncer userlist secret created"
	@# Apply ConfigMap and Deployment
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -f $(KUSTOMIZE_DIR)/services/pgbouncer/pgbouncer-configmap.yaml
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -f $(KUSTOMIZE_DIR)/services/pgbouncer/pgbouncer-deployment.yaml
	@echo "  Waiting for PgBouncer to be ready..."
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) rollout status deployment/pgbouncer --timeout=60s
	@echo "$(GREEN)PgBouncer deployed on pgbouncer.$(PROJECT_NAME).svc.cluster.local:6432$(NC)"

.PHONY: deploy-070-database-schemas
deploy-070-database-schemas: ## Run database migrations
	@echo "$(GREEN)Deploying 070-database-schemas...$(NC)"
	@cd $(TERRAFORM_DIR)/070-database-schemas && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-080-kafka-topics
deploy-080-kafka-topics: ## Create Kafka topics
	@echo "$(GREEN)Deploying 080-kafka-topics...$(NC)"
	@cd $(TERRAFORM_DIR)/080-kafka-topics && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-090-monitoring
deploy-090-monitoring: ## Deploy monitoring stack
	@echo "$(GREEN)Deploying 090-monitoring...$(NC)"
	@cd $(TERRAFORM_DIR)/090-monitoring && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

# bugs_open/252 / register BLD-021: the node-config DaemonSet as an INSTALL
# step, so a fresh cluster does not come up with the image-GC trigger sitting
# on the eviction line. Terraform wraps the same kustomize overlay that
# deploy-node-config applies day-to-day; both are safe to run in either order
# (same manifest, idempotent apply). Production-only, like the overlay: the
# step is skipped with a notice when the terraform dir does not exist for
# $(ENVIRONMENT)/$(REGION), rather than failing a dev install.
.PHONY: deploy-095-node-config
deploy-095-node-config: ## Deploy 095-node-config (kubelet image-GC DaemonSet, BLD-021)
	@if [ -d "$(TERRAFORM_DIR)/095-node-config" ]; then \
		echo "$(GREEN)Deploying 095-node-config...$(NC)"; \
		cd $(TERRAFORM_DIR)/095-node-config && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi; \
	else \
		echo "$(YELLOW)skipping 095-node-config — no $(TERRAFORM_DIR)/095-node-config (production-only)$(NC)"; \
	fi

# 096/097 — the CI runners and ollama as INSTALL steps (owner directive
# 2026-08-14: the whole framework installs via terraform+kustomize). Each
# terraform root wraps the same overlays deploy-agents applies day-to-day —
# same manifests, idempotent in either order, and NO image sweep in either
# path (runners are pinned to their own tags; ollama runs upstream
# ollama/ollama). Same skip/fail shape as 095.
.PHONY: deploy-096-github-runners
deploy-096-github-runners: ## Deploy 096-github-runners (both CI runner deployments)
	@if [ -d "$(TERRAFORM_DIR)/096-github-runners" ]; then \
		echo "$(GREEN)Deploying 096-github-runners...$(NC)"; \
		cd $(TERRAFORM_DIR)/096-github-runners && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi; \
	else \
		echo "$(YELLOW)skipping 096-github-runners — no $(TERRAFORM_DIR)/096-github-runners (production-only)$(NC)"; \
	fi

.PHONY: deploy-097-ollama
deploy-097-ollama: ## Deploy 097-ollama (ollama-adapter + ollama-eval)
	@if [ -d "$(TERRAFORM_DIR)/097-ollama" ]; then \
		echo "$(GREEN)Deploying 097-ollama...$(NC)"; \
		cd $(TERRAFORM_DIR)/097-ollama && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi; \
	else \
		echo "$(YELLOW)skipping 097-ollama — no $(TERRAFORM_DIR)/097-ollama (production-only)$(NC)"; \
	fi


# bugs_open/066: this is the BOOTSTRAP path — on a fresh cluster there is no
# agent-chassis Deployment to read a tag from, so unlike update-agent-images it
# must seed from $(IMAGE_TAG). Scoped the same way regardless: never touch
# is_snapshot rollback copies (021_model_swap_and_rollback.sql), soft-deleted
# rows, or a row deliberately pinned with default_config.pin_image_tag.
# `2>/dev/null` removed — the failure stays non-fatal for a bootstrap, but a
# silent no-op here is exactly how a stale census goes unnoticed.
.PHONY: deploy-100-bootstrap-agents
deploy-100-bootstrap-agents: ## Deploy bootstrap agents (generic orchestrator) with image updates
	@echo "$(GREEN)Deploying 100-bootstrap-agents...$(NC)"
	@echo "$(YELLOW)First updating agent definitions with current image...$(NC)"
	@printf "%s\n" \
		"UPDATE agent_definitions SET image_repository = :'img_repo', image_tag = :'img_tag', updated_at = NOW()" \
		" WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false) = false" \
		"   AND COALESCE(default_config->'pin_image_tag','false'::jsonb) <> 'true'::jsonb;" \
		"SELECT COUNT(*) as updated_count FROM agent_definitions;" \
	| KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec --request-timeout=5m -i postgres-clients-0 -n $(PROJECT_NAME) -- \
		psql -U clients_user -d clients_db -v img_repo="$(REGISTRY)/agent-chassis" -v img_tag="$(IMAGE_TAG)" || true
	@cd $(TERRAFORM_DIR)/100-bootstrap-agents && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve \
				-var-file=terraform.tfvars.secret \
				-var="image_tag=$(IMAGE_TAG)" \
				-var="registry=$(REGISTRY)"; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve \
				-var="image_tag=$(IMAGE_TAG)" \
				-var="registry=$(REGISTRY)"; \
		fi
	@echo "$(GREEN)Bootstrap agents deployed with image $(REGISTRY)/agent-chassis:$(IMAGE_TAG)$(NC)"


# Add these to the existing Makefile
# Place after the existing Kafka section (near deploy-080-kafka-topics)

#################################
# Kafka Topic Management
#################################

KAFKA_NS := kafka
KAFKA_POD := personae-kafka-cluster-combined-pool-prod-0
KAFKA_BOOTSTRAP := localhost:9092

# Topics required by scheduler-triggered agents (not created by spawn_agent)
# These agents receive messages directly from the kafka-scheduler,
# not via the spawn→job topic pattern. Their process topics must pre-exist.
SCHEDULER_AGENT_TOPICS := \
	system.agent.endpoint-health-checker.process \
	system.agent.build-dispatch-loop.process \
	system.agent.improvement-loop.process

.PHONY: kafka-ensure-scheduler-topics
kafka-ensure-scheduler-topics: ## Create Kafka topics for scheduler-triggered agents
	@echo "$(YELLOW)Ensuring scheduler-triggered agent topics exist...$(NC)"
	@for topic in $(SCHEDULER_AGENT_TOPICS); do \
		echo "  Checking $$topic..."; \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(KAFKA_NS) exec $(KAFKA_POD) -- \
			bin/kafka-topics.sh --describe --topic $$topic \
			--bootstrap-server $(KAFKA_BOOTSTRAP) > /dev/null 2>&1 \
		|| ( \
			echo "  $(YELLOW)Creating $$topic$(NC)"; \
			KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(KAFKA_NS) exec $(KAFKA_POD) -- \
				bin/kafka-topics.sh --create \
				--topic $$topic \
				--partitions 1 \
				--replication-factor 2 \
				--bootstrap-server $(KAFKA_BOOTSTRAP) 2>&1 \
		); \
	done
	@echo "$(GREEN)Scheduler agent topics ready$(NC)"

.PHONY: kafka-list-system-topics
kafka-list-system-topics: ## List all system.agent.* Kafka topics
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(KAFKA_NS) exec $(KAFKA_POD) -- \
		bin/kafka-topics.sh --list --bootstrap-server $(KAFKA_BOOTSTRAP) \
		| grep '^system\.' | sort

.PHONY: kafka-list-job-topics
kafka-list-job-topics: ## List all job.* Kafka topics (dynamic, from spawn_agent)
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(KAFKA_NS) exec $(KAFKA_POD) -- \
		bin/kafka-topics.sh --list --bootstrap-server $(KAFKA_BOOTSTRAP) \
		| grep '^job\.' | sort

#################################
# AI Endpoint Health
#################################

.PHONY: health-status
health-status: ## Show AI endpoint health status
	@echo "$(YELLOW)AI Endpoint Health:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -it postgres-clients-0 -n $(PROJECT_NAME) -- \
		psql -U clients_user -d clients_db -c \
		"SELECT name, CASE WHEN healthy THEN 'UP' ELSE 'DOWN' END as status, \
		 error, last_checked, \
		 CASE WHEN last_checked IS NOT NULL THEN age(now(), last_checked)::text ELSE 'never' END as since_checked \
		 FROM ai_endpoint_health ORDER BY name;"

.PHONY: health-reset-claude
health-reset-claude: ## Manually reset Claude endpoint to healthy (use after credit top-up)
	@echo "$(YELLOW)Resetting Claude endpoint health...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -it postgres-clients-0 -n $(PROJECT_NAME) -- \
		psql -U clients_user -d clients_db -c \
		"UPDATE ai_endpoint_health \
		 SET healthy = true, last_checked = NOW(), last_healthy = NOW(), \
		     error = NULL, updated_at = NOW() \
		 WHERE endpoint_url = 'https://api.anthropic.com/v1/messages'; \
		 SELECT name, healthy, last_checked FROM ai_endpoint_health \
		 WHERE name = 'claude';"
	@echo "$(GREEN)Claude endpoint reset to healthy$(NC)"

.PHONY: health-reset-all
health-reset-all: ## Reset all AI endpoints to healthy
	@echo "$(YELLOW)Resetting all endpoint health...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -it postgres-clients-0 -n $(PROJECT_NAME) -- \
		psql -U clients_user -d clients_db -c \
		"UPDATE ai_endpoint_health \
		 SET healthy = true, last_checked = NOW(), last_healthy = NOW(), \
		     error = NULL, updated_at = NOW(); \
		 SELECT name, healthy, last_checked FROM ai_endpoint_health;"
	@echo "$(GREEN)All endpoints reset$(NC)"

#################################
# Application Deployment (Terraform Workflow)
#################################
# Generic target for deploying any service via Terraform
.PHONY: deploy-all
deploy-all: deploy-infrastructure deploy-core deploy-agents ## deploy-frontends ## Deploy everything

.PHONY: deploy-service
deploy-service:
	@echo "$(GREEN)Deploying service at $(path)...$(NC)"
	@cd $(path) && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init -upgrade && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init -upgrade && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

# Generic target for destroying any service via Terraform
.PHONY: destroy-service
destroy-service:
	@echo "$(RED)Destroying service at $(path)...$(NC)"
	@cd $(path) && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init -upgrade && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform destroy -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init -upgrade && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform destroy -auto-approve; \
		fi

# Core Platform Services
.PHONY: deploy-core
deploy-core: update-kustomization-images deploy-047-base-configs deploy-auth-service deploy-core-manager ## Deploy core platform services using Terraform

.PHONY: deploy-auth-service
deploy-auth-service:  ## Deploy auth-service using Terraform
	# Update the image tag in kustomization.yaml FIRST
	@echo "$(YELLOW)Updating auth-service image tag to $(IMAGE_TAG)...$(NC)"
	@cd $(KUSTOMIZE_DIR)/services/auth-service/overlays/$(OVERLAY_PATH) && \
		sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' kustomization.yaml

	@$(MAKE) deploy-service path=$(TERRAFORM_DIR)/services/core-platform/1110-auth-service

.PHONY: deploy-core-manager
deploy-core-manager:  ## Deploy core-manager using Terraform
# Update the image tag in kustomization.yaml FIRST
	@echo "$(YELLOW)Updating core-manager image tag to $(IMAGE_TAG)...$(NC)"
	@cd $(KUSTOMIZE_DIR)/services/core-manager/overlays/$(OVERLAY_PATH) && \
		sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' kustomization.yaml


	@$(MAKE) deploy-service path=$(TERRAFORM_DIR)/services/core-platform/1120-core-manager


# deploy-remote-job-spawner: DELETED 2026-08-17 (bugs_open/237). It was an explicit rule, and an
# explicit rule BEATS the deploy-% pattern rule — so this service silently opted out
# of the pattern rule's registry pre-flight and could be deployed at a tag nobody
# built. `make deploy-remote-job-spawner` still works and now asks the registry first.

.PHONY: deploy-remote-job-spawner-tf
deploy-remote-job-spawner-tf: ## Deploy remote-job-spawner using Terraform
	@$(MAKE) deploy-service path=$(TERRAFORM_DIR)/services/agents/2220-remote-job-spawner

# deploy-kafka-scheduler: DELETED 2026-08-17 (bugs_open/237). It was an explicit rule, and an
# explicit rule BEATS the deploy-% pattern rule — so this service silently opted out
# of the pattern rule's registry pre-flight and could be deployed at a tag nobody
# built. `make deploy-kafka-scheduler` still works and now asks the registry first.

.PHONY: deploy-kafka-scheduler-tf
deploy-kafka-scheduler-tf: ## Deploy kafka-scheduler using Terraform
	@$(MAKE) deploy-service path=$(TERRAFORM_DIR)/services/agents/2270-kafka-scheduler

.PHONY: push-kafka-scheduler
push-kafka-scheduler: ## Push kafka-scheduler image
	docker push $(REGISTRY)/kafka-scheduler:$(IMAGE_TAG)

.PHONY: quick-scheduler-update
quick-scheduler-update: ## Build, push and deploy kafka-scheduler with current IMAGE_TAG
	@echo "$(YELLOW)Building kafka-scheduler:$(IMAGE_TAG)...$(NC)"
	@$(MAKE) build-kafka-scheduler IMAGE_TAG=$(IMAGE_TAG)
	@echo "$(YELLOW)Pushing kafka-scheduler:$(IMAGE_TAG)...$(NC)"
	@docker push $(REGISTRY)/kafka-scheduler:$(IMAGE_TAG)
	@echo "$(YELLOW)Deploying...$(NC)"
	@$(MAKE) deploy-kafka-scheduler IMAGE_TAG=$(IMAGE_TAG)
	@echo "$(YELLOW)Restarting...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl rollout restart deployment/kafka-scheduler -n ai-persona-system
	@echo "$(GREEN)Scheduler deployed with $(REGISTRY)/kafka-scheduler:$(IMAGE_TAG)$(NC)"

.PHONY: logs-scheduler
logs-scheduler: ## Tail logs from kafka-scheduler
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl logs -f -n $(PROJECT_NAME) -l app=kafka-scheduler

.PHONY: deploy-ollama-adapter
deploy-ollama-adapter: ## Deploy ollama-adapter using kustomize (uses ollama/ollama image, not aqls)
	@echo "$(YELLOW)Deploying ollama-adapter...$(NC)"
	@if [ -d "$(KUSTOMIZE_DIR)/services/ollama-adapter/overlays/$(OVERLAY_PATH)" ]; then \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/ollama-adapter/overlays/$(OVERLAY_PATH); \
	else \
		echo "$(RED)Ollama adapter kustomize directory not found at $(KUSTOMIZE_DIR)/services/ollama-adapter/overlays/$(OVERLAY_PATH)$(NC)"; \
	fi

.PHONY: logs-ollama
logs-ollama: ## Tail logs from ollama-adapter
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl logs -f -n $(PROJECT_NAME) -l app=ollama-adapter


#################################
# WireGuard VPN
#################################

.PHONY: wg-genkeys
wg-genkeys: ## Generate WireGuard keypairs (run once)
	@echo "$(YELLOW)Generating WireGuard keys...$(NC)"
	@mkdir -p .secrets/wireguard
	@wg genkey | tee .secrets/wireguard/server-private.key | wg pubkey > .secrets/wireguard/server-public.key
	@wg genkey | tee .secrets/wireguard/admin-private.key | wg pubkey > .secrets/wireguard/admin-public.key
	@echo "$(GREEN)Keys generated in .secrets/wireguard/$(NC)"
	@echo "Server public key: $$(cat .secrets/wireguard/server-public.key)"
	@echo "Admin public key:  $$(cat .secrets/wireguard/admin-public.key)"
	@echo ""
	@echo "$(YELLOW)Now create terraform.tfvars.secret:$(NC)"
	@echo "  cd $(TERRAFORM_DIR)/048-wireguard"
	@echo "  cp terraform.tfvars.secret.example terraform.tfvars.secret"
	@echo "  # Paste the private/public keys from .secrets/wireguard/"

.PHONY: deploy-048-wireguard
deploy-048-wireguard: ## Deploy WireGuard secret via Terraform
	@echo "$(GREEN)Deploying 048-wireguard secret...$(NC)"
	@cd $(TERRAFORM_DIR)/048-wireguard && \
		KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
		KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret

.PHONY: deploy-wireguard
deploy-wireguard: deploy-048-wireguard ## Deploy WireGuard pod via kustomize
	@echo "$(YELLOW)Deploying WireGuard VPN...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/wireguard/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)WireGuard deployed. Get the NodePort:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get svc wireguard

.PHONY: wg-client-config
wg-client-config: ## Generate WireGuard client config for your laptop
	@echo "$(YELLOW)Generating client config...$(NC)"
	@NODE_IP=$$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}'); \
	if [ -z "$$NODE_IP" ]; then \
		NODE_IP=$$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}'); \
	fi; \
	echo "[Interface]"; \
	echo "Address = 10.8.0.2/24"; \
	echo "PrivateKey = $$(cat .secrets/wireguard/admin-private.key)"; \
	echo "DNS = 10.96.0.10"; \
	echo ""; \
	echo "[Peer]"; \
	echo "PublicKey = $$(cat .secrets/wireguard/server-public.key)"; \
	echo "Endpoint = $$NODE_IP:31820"; \
	echo "AllowedIPs = 10.8.0.0/24, 10.96.0.0/12"; \
	echo "PersistentKeepalive = 25"

.PHONY: wg-status
wg-status: ## Check WireGuard pod status
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) exec deploy/wireguard -- wg show


#################################
# Vet Intel Agent
#################################
# deploy-vet-intel: DELETED 2026-08-17 (bugs_open/237). It was an explicit rule, and an
# explicit rule BEATS the deploy-% pattern rule — so this service silently opted out
# of the pattern rule's registry pre-flight and could be deployed at a tag nobody
# built. `make deploy-vet-intel` still works and now asks the registry first.

.PHONY: logs-vet-intel
logs-vet-intel: ## Tail logs from vet-intel agent
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl logs -f -n $(PROJECT_NAME) -l app=vet-intel

.PHONY: restart-vet-intel
restart-vet-intel: ## Restart vet-intel agent
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl rollout restart deployment/vet-intel -n $(PROJECT_NAME)

#################################
# Business Intel Agent
#################################
# deploy-business-intel: DELETED 2026-08-17 (bugs_open/237). It was an explicit rule, and an
# explicit rule BEATS the deploy-% pattern rule — so this service silently opted out
# of the pattern rule's registry pre-flight and could be deployed at a tag nobody
# built. `make deploy-business-intel` still works and now asks the registry first.

.PHONY: logs-business-intel
logs-business-intel: ## Tail logs from business-intel agent
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl logs -f -n $(PROJECT_NAME) -l app=business-intel

.PHONY: restart-business-intel
restart-business-intel: ## Restart business-intel agent
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl rollout restart deployment/business-intel -n $(PROJECT_NAME)

#################################
# Agent Job Cleanup
#################################
.PHONY: deploy-agent-cleanup
deploy-agent-cleanup: ## Deploy the agent-job-cleanup CronJob
	@echo "$(YELLOW)Deploying agent-job-cleanup CronJob...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -f $(KUSTOMIZE_DIR)/services/agent-job-cleanup/agent-job-cleanup-cronjob.yaml
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -f $(KUSTOMIZE_DIR)/services/agent-job-cleanup/kafka-cleanup-rbac.yaml
	@echo "$(GREEN)Agent cleanup CronJob deployed (runs every 10 min)$(NC)"

.PHONY: agent-cleanup-now
agent-cleanup-now: ## Run agent job cleanup immediately (delete stale spawned jobs and failed pods)
	@echo "$(YELLOW)Cleaning up stale agent jobs...$(NC)"
	@FAILED=$$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl get pods -n $(PROJECT_NAME) --field-selector=status.phase=Failed --no-headers 2>/dev/null | wc -l); \
	if [ "$$FAILED" -gt 0 ]; then \
		echo "Deleting $$FAILED failed pods"; \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl delete pods -n $(PROJECT_NAME) --field-selector=status.phase=Failed; \
	fi
	@for AGENT_TYPE in vet-practice-verifier vet-batch-processor area-sweep-orchestrator area-sweep-discoverer; do \
		COUNT=$$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl get jobs -n $(PROJECT_NAME) -l "spawned-by=orchestrator,agent-type=$$AGENT_TYPE" --no-headers 2>/dev/null | wc -l); \
		if [ "$$COUNT" -gt 0 ]; then \
			echo "Deleting $$COUNT $$AGENT_TYPE jobs"; \
			KUBECONFIG=$(KUBECONFIG_PATH) kubectl delete jobs -n $(PROJECT_NAME) -l "spawned-by=orchestrator,agent-type=$$AGENT_TYPE"; \
		fi; \
	done
	@REMAINING=$$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl get jobs -n $(PROJECT_NAME) -l spawned-by=orchestrator --no-headers 2>/dev/null | wc -l); \
	echo "$(GREEN)Cleanup complete. $$REMAINING spawned jobs remaining$(NC)"

.PHONY: agent-status
agent-status: ## Show spawned agent pod counts and status
	@echo "$(YELLOW)Spawned agent status:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl get pods -n $(PROJECT_NAME) -l spawned-by=orchestrator \
		--no-headers 2>/dev/null | awk '{types[$$1]=$$3} END {for (t in types) print t, types[t]}' || true
	@echo ""
	@echo "By agent type:"
	@for AGENT_TYPE in vet-practice-verifier vet-batch-processor area-sweep-orchestrator area-sweep-discoverer; do \
		RUNNING=$$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl get pods -n $(PROJECT_NAME) -l "agent-type=$$AGENT_TYPE" --field-selector=status.phase=Running --no-headers 2>/dev/null | wc -l); \
		FAILED=$$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl get pods -n $(PROJECT_NAME) -l "agent-type=$$AGENT_TYPE" --field-selector=status.phase=Failed --no-headers 2>/dev/null | wc -l); \
		PENDING=$$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl get pods -n $(PROJECT_NAME) -l "agent-type=$$AGENT_TYPE" --field-selector=status.phase=Pending --no-headers 2>/dev/null | wc -l); \
		echo "  $$AGENT_TYPE: running=$$RUNNING failed=$$FAILED pending=$$PENDING"; \
	done

# Corresponding destroy targets
.PHONY: destroy-core
destroy-core: destroy-core-manager destroy-auth-service ## Destroy core platform services using Terraform

.PHONY: destroy-auth-service
destroy-auth-service: ## Destroy auth-service using Terraform
	@$(MAKE) destroy-service path=$(TERRAFORM_DIR)/services/core-platform/1110-auth-service

.PHONY: destroy-core-manager
destroy-core-manager: ## Destroy core-manager using Terraform
	@$(MAKE) destroy-service path=$(TERRAFORM_DIR)/services/core-platform/1120-core-manager


# Update all agent images. The service list and each service's image are the
# ONE declaration at the top of this file (AGENT_DEPLOY_SERVICES, bugs_open/237)
# — the fallback below writes the DECLARED image, never a guess from the
# service name, which was correct for every image-owning service and silently
# wrong for the three that run another service's binary.
.PHONY: update-kustomization-images
update-kustomization-images: check-release-coverage ## Update image tags in kustomization.yaml files
	@echo "$(YELLOW)Updating kustomization.yaml files with image tag $(IMAGE_TAG)...$(NC)"
	@for entry in $(AGENT_DEPLOY_SERVICES); do \
		agent=$${entry%%:*}; img=$${entry#*:}; [ "$$img" = "$$entry" ] && img=$$agent; \
		kust_file="$(KUSTOMIZE_DIR)/services/$$agent/overlays/$(OVERLAY_PATH)/kustomization.yaml"; \
		if [ -f "$$kust_file" ]; then \
			echo "Updating $$agent kustomization.yaml..."; \
			if grep -q "images:" "$$kust_file"; then \
				sed -i.bak '/images:/,/^[^ ]/{/newTag:/s/newTag:.*/newTag: $(IMAGE_TAG)/}' "$$kust_file"; \
				rm -f "$$kust_file.bak"; \
			else \
				echo "" >> "$$kust_file"; \
				echo "images:" >> "$$kust_file"; \
				echo "  - name: $(REGISTRY)/$$img" >> "$$kust_file"; \
				echo "    newTag: $(IMAGE_TAG)" >> "$$kust_file"; \
			fi; \
		fi; \
	done

# Deploy agents with automatic image update
# Update ConfigMap with new image tag
.PHONY: update-agent-image-tag
update-agent-image-tag: ## Update the agent image tag in ConfigMap
	@echo "$(YELLOW)Updating agent image tag to $(IMAGE_TAG)...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl patch configmap personae-prod-config \
		-n ai-persona-system \
		--type merge \
		-p '{"data":{"AGENT_IMAGE_TAG":"$(IMAGE_TAG)","agent_image_tag":"$(IMAGE_TAG)"}}'

# Deploy agents with automatic image update
.PHONY: deploy-agents
deploy-agents: ## Deploy all agent services with dynamic image tag
	@echo "$(YELLOW)Deploying agent services with image tag $(IMAGE_TAG)...$(NC)"

	# Retag and apply every service in AGENT_DEPLOY_SERVICES (declared ONCE at
	# the top of this file — bugs_open/237). This used to be fifteen
	# copy-pasted six-line blocks, and the copy-paste is precisely how
	# render-audit-adapter came to be in NEITHER tag mechanism: it sat on
	# v1.0.1194 while the fleet ran v1.0.1274 — 80 tags, frozen since the pod
	# was created, invisible to the normal proof because pod-grepping the image
	# OWNER (browser-runner-adapter) read live. check-release-coverage now
	# refuses a release that leaves an overlay out; this loop is what makes the
	# list the only thing to keep right.
	#
	# Two per-service facts that used to live in the deleted comments, kept here
	# because they are still true:
	#
	#  - browser-runner-adapter's REQUESTS TOPIC is a Strimzi KafkaTopic in the
	#    `kafka` namespace and is NOT part of its overlay (the overlay forces
	#    ai-persona-system). Apply it ONCE before the first deploy:
	#      kubectl apply -f $(KUSTOMIZE_DIR)/services/browser-runner-adapter/overlays/$(OVERLAY_PATH)/browser-runner-requests-topic.yaml
	#
	#  - render-audit-adapter has NO IMAGE OF ITS OWN: it runs the browser-runner
	#    binary under a different topic and consumer group, which is why it has no
	#    build-* and no push-* step and why none should be added. It is ordered
	#    directly after browser-runner-adapter because the two must move together:
	#    the tag applied here has to be one the browser-runner was actually built
	#    at.
	#
	# A declared service with no overlay at $(OVERLAY_PATH) is SKIPPED WITH A
	# WARNING, not silently: the old blocks hid that case behind
	# `2>/dev/null || true`, which is the same "absence looks like success"
	# shape as the bug. It is a warning rather than an error because only 9
	# services have a development overlay and 1 a staging one, so a hard failure
	# would break every non-production deploy.
	@for entry in $(AGENT_DEPLOY_SERVICES); do \
		svc=$${entry%%:*}; \
		overlay="$(KUSTOMIZE_DIR)/services/$$svc/overlays/$(OVERLAY_PATH)"; \
		if [ ! -f "$$overlay/kustomization.yaml" ]; then \
			echo "$(YELLOW)  SKIPPED $$svc — declared in AGENT_DEPLOY_SERVICES but no overlay at $$overlay$(NC)"; \
			continue; \
		fi; \
		echo "Updating $$svc to $(IMAGE_TAG)..."; \
		sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' "$$overlay/kustomization.yaml"; \
		rm -f "$$overlay/kustomization.yaml.bak"; \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k "$$overlay"; \
	done


	# Deploy ollama-adapter AND ollama-eval (both use the upstream ollama/ollama
	# image — NOT updated by IMAGE_TAG). ollama-eval added 2026-08-14: it was
	# live in the cluster but applied by hand only, so nothing reconciled it —
	# the same drift hole the runners had before deploy-github-runners.
	@echo "Deploying ollama-adapter + ollama-eval..."
	@for svc in ollama-adapter ollama-eval; do \
		if [ -d "$(KUSTOMIZE_DIR)/services/$$svc/overlays/$(OVERLAY_PATH)" ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/$$svc/overlays/$(OVERLAY_PATH); \
		fi; \
	done

	# Re-apply BOTH github runners' manifests.
	#
	# ⚠ CHANGED 2026-08-18 (owner ruling, bugs_open/237 Decision B): the runners
	# are now IN AGENT_DEPLOY_SERVICES, so the loop above has already retagged and
	# applied both — including -vmsites, which maps to the github-actions-runner
	# image because there is one image and two Deployments. This call is therefore
	# a redundant re-apply and is KEPT ON PURPOSE for the one thing the generic
	# loop does not do: the loop SKIPS a missing overlay with a warning, whereas
	# deploy-github-runners FAILS in production. The runners are production-only,
	# so for them a missing overlay means a release is about to report success
	# while shipping none of the manifests it claims to — worth one idempotent
	# apply to keep. It no longer has anything to do with image tags.
	@$(MAKE) --no-print-directory deploy-github-runners

	# Deploy the node-config DaemonSet (bugs_open/252: kubelet image-GC settings
	# applied per node — the kubelet-config ConfigMap is provider-protected, so
	# node files are the only tenant-reachable home; see the target's comment).
	@$(MAKE) --no-print-directory deploy-node-config

	# Update database agent definitions
	@$(MAKE) update-agent-images-v2 IMAGE_TAG=$(IMAGE_TAG)

	# Force rollout restart to pick up new images
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl rollout restart deployment/agent-chassis -n ai-persona-system 2>/dev/null || true
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl rollout restart deployment/vet-intel -n ai-persona-system 2>/dev/null || true
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl rollout restart deployment/business-intel -n ai-persona-system 2>/dev/null || true

	@echo "$(GREEN)All agents deployed with image tag $(IMAGE_TAG)$(NC)"

#  kubectl apply -k deployments/kustomize/services/agent-chassis/overlays/production/uk_001/

#################################
# Single-service deploy
#
# deploy-agents above is ALL-OR-NOTHING: it seds every service's kustomization to
# $(IMAGE_TAG) and applies them. That is only safe when every service has been
# built and pushed at that tag. It usually has not been — services are built one
# at a time as their code changes, so on a normal day two or three tags exist and
# the rest of the fleet is several behind. Running it then points twelve healthy
# deployments at an image that was never pushed, and they ImagePullBackOff
# together.
#
#   make deploy-<service>                deploy ONE service at $(IMAGE_TAG)
#
# This mirrors the build side, which already solved the same problem with the
# build-%-ref / build-%-tree pattern rules. It is deliberately the same shape:
# one service, named explicitly, no fan-out.
#
# The registry pre-flight is the load-bearing part. push-*/deploy-* are git-blind
# (see the header at the top of this file) — nothing downstream of the build
# checks that the tag you are deploying exists. Asking the registry before
# touching the cluster turns a rolled-back deployment into a refusal that costs
# nothing.
#################################
deploy-%: ## Deploy ONE service at $(IMAGE_TAG): make deploy-browser-runner-adapter
	@OVERLAY="$(KUSTOMIZE_DIR)/services/$*/overlays/$(OVERLAY_PATH)"; \
	test -d "$$OVERLAY" || { \
		echo "$(RED)No overlay at $$OVERLAY — is '$*' a service name?$(NC)"; exit 1; }; \
	test -f "$$OVERLAY/kustomization.yaml" || { \
		echo "$(RED)No kustomization.yaml in $$OVERLAY$(NC)"; exit 1; }; \
	IMG=$$(awk '/^images:/{i=1;next} i&&/name:/{print $$NF;exit}' "$$OVERLAY/kustomization.yaml"); \
	[ -n "$$IMG" ] || IMG="$(REGISTRY)/$*"; \
	BUILD=$${IMG##*/}; \
	if ! docker manifest inspect $$IMG:$(IMAGE_TAG) >/dev/null 2>&1; then \
		echo "$(RED)$$IMG:$(IMAGE_TAG) is not in the registry.$(NC)"; \
		echo "$(YELLOW)  Deploying it would ImagePullBackOff. Build and push first:$(NC)"; \
		echo "    make build-$$BUILD && docker push $$IMG:$(IMAGE_TAG)"; \
		if [ "$$BUILD" != "$*" ]; then \
			echo "$(YELLOW)  (note: $* runs the $$BUILD image — it has no image of its own.)$(NC)"; \
		fi; \
		exit 1; \
	fi; \
	echo "$(GREEN)Deploying $* at $(IMAGE_TAG) — this service only.$(NC)"; \
	sed -i.bak 's|newTag:.*|newTag: $(IMAGE_TAG)|' "$$OVERLAY/kustomization.yaml"; \
	rm -f "$$OVERLAY/kustomization.yaml.bak"; \
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k "$$OVERLAY"
	@echo "$(YELLOW)Verify against the running POD, not the tag — a retag is not a rebuild:$(NC)"
	@echo "  kubectl -n ai-persona-system get pods -l app=$* -o custom-columns=NAME:.metadata.name,START:.status.startTime,IMAGE:.spec.containers[0].image"

# There is deliberately NO explicit deploy-render-audit-adapter rule any more
# (deleted 2026-08-17, bugs_open/237). It existed because the pattern rule's
# pre-flight asked for $(REGISTRY)/<service>:$(IMAGE_TAG), which for a service
# running ANOTHER service's binary names an image that has never existed — so
# the deploy was refused as "not in the registry". The pattern rule now reads
# the image out of the overlay, which is right for every such service without
# anybody writing a bespoke rule; keeping a hand-written duplicate alongside it
# would have been one more pair of things that must stay identical, which is
# the defect this bug is about.

.PHONY: redeploy-agents
redeploy-agents:  ## Forces a rolling restart of all agent deployments
	@echo "$(YELLOW)Forcing rollout restart of agent deployments...$(NC)"
	@for entry in $(AGENT_DEPLOY_SERVICES) ollama-adapter; do \
		svc=$${entry%%:*}; \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl rollout restart deployment $$svc -n ai-persona-system 2>/dev/null \
			|| echo "$(YELLOW)  no deployment/$$svc to restart$(NC)"; \
	done

.PHONY: deploy-frontends
deploy-frontends: ## Deploy all frontend applications
	@echo "$(YELLOW)Deploying frontend applications...$(NC)"
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/admin-dashboard/overlays/$(OVERLAY_PATH)
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/user-portal/overlays/$(OVERLAY_PATH)
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/agent-playground/overlays/$(OVERLAY_PATH)

# (deploy-admin-dashboard is defined once, above — the duplicate here was removed
#  2026-08-17, bugs_open/237.)

.PHONY: deploy-user-portal
deploy-user-portal: ## Deploy user-portal only
	@echo "$(GREEN)Deploying user-portal...$(NC)"
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/user-portal/overlays/$(OVERLAY_PATH)

#################################
# Full Stack Operations
#################################
.PHONY: full-deploy
full-deploy: build-all push-all deploy-all ## Build, push, and deploy everything

.PHONY: quick-deploy
quick-deploy:  ## Deploy applications without building (uses existing images)
	@echo "$(YELLOW)Quick deployment using existing images...$(NC)"
	@$(MAKE) deploy-core
	@$(MAKE) deploy-agents
	@$(MAKE) deploy-frontends

#################################
# Status and Monitoring
#################################
.PHONY: status
status: ## Show status of all deployments
	@echo "$(YELLOW)Deployment Status:$(NC)"
	kubectl get deployments -n $(PROJECT_NAME)
	@echo "\n$(YELLOW)Services:$(NC)"
	kubectl get services -n $(PROJECT_NAME)
	@echo "\n$(YELLOW)Pods:$(NC)"
	kubectl get pods -n $(PROJECT_NAME)

.PHONY: logs
logs: ## Tail logs from all pods
	kubectl logs -f -n $(PROJECT_NAME) -l app.kubernetes.io/part-of=$(PROJECT_NAME) --all-containers=true

.PHONY: logs-auth
logs-auth: ## Tail logs from auth-service
	kubectl logs -f -n $(PROJECT_NAME) -l app=auth-service --all-containers=true

.PHONY: logs-core
logs-core: ## Tail logs from core-manager
	kubectl logs -f -n $(PROJECT_NAME) -l app=core-manager --all-containers=true

#################################
# Rollback Operations
#################################
.PHONY: rollback-auth-service
rollback-auth-service: ## Rollback auth-service deployment
	kubectl rollout undo deployment/auth-service -n $(PROJECT_NAME)

.PHONY: rollback-core-manager
rollback-core-manager: ## Rollback core-manager deployment
	kubectl rollout undo deployment/core-manager -n $(PROJECT_NAME)

#################################
# Testing
#################################
.PHONY: test
test: test-unit test-integration ## Run all tests

.PHONY: test-unit
test-unit: ## Run unit tests
	@echo "$(YELLOW)Running unit tests...$(NC)"
	go test ./... -v -short

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(YELLOW)Running integration tests...$(NC)"
	go test ./tests/integration/... -v

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	@echo "$(YELLOW)Running E2E tests...$(NC)"
	go test ./tests/e2e/... -v

#################################
# Database Operations
#################################
.PHONY: db-migrate
db-migrate: ## Run database migrations
	@echo "$(YELLOW)Running database migrations...$(NC)"
	$(SCRIPTS_DIR)/migration/run-migrations.sh

.PHONY: db-seed
db-seed: ## Seed database with test data
	@echo "$(YELLOW)Seeding database...$(NC)"
	kubectl exec --request-timeout=5m -it deployment/postgres-clients -n $(PROJECT_NAME) -- \
		psql -U postgres -f /scripts/seed-data.sql

#################################
# Utility Commands
#################################
.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -rf dist/
	rm -rf frontends/*/build/
	rm -rf frontends/*/dist/

.PHONY: port-forward-admin
port-forward-admin: dashboard-port-forward ## Port forward admin dashboard (alias for dashboard-port-forward)
# Was `svc/admin-dashboard 3000:80`, which could never connect: the Service
# listens on 8080, not 80. Aliased rather than re-fixed so there is one
# definition to keep correct.

.PHONY: port-forward-grafana
port-forward-grafana: ## Port forward Grafana to localhost:3001
	kubectl port-forward -n $(PROJECT_NAME) svc/grafana 3001:3000

#################################
# Individual Service Builds & Deploys
#################################
# Convenience targets for individual service development
.PHONY: auth-service
auth-service: build-auth-service push-auth-service deploy-auth-service ## Build, push and deploy auth-service

.PHONY: core-manager
core-manager: build-core-manager push-core-manager deploy-core-manager ## Build, push and deploy core-manager

.PHONY: admin-dashboard
admin-dashboard: build-admin-dashboard push-admin-dashboard deploy-admin-dashboard ## Build, push and deploy admin-dashboard

# Push individual services
.PHONY: push-auth-service
push-auth-service: ## Push auth-service image
	docker push $(REGISTRY)/auth-service:$(IMAGE_TAG)

.PHONY: push-core-manager
push-core-manager: ## Push core-manager image
	docker push $(REGISTRY)/core-manager:$(IMAGE_TAG)

.PHONY: push-admin-dashboard
push-admin-dashboard: ## Push admin-dashboard image
	docker push $(REGISTRY)/admin-dashboard:$(IMAGE_TAG)

#################################
# Terraform Operations
#################################
.PHONY: tf-plan
tf-plan: ## Run terraform plan for all infrastructure
	@echo "$(YELLOW)Running Terraform plan...$(NC)"
	@for dir in $(TERRAFORM_DIR)/0*; do \  # This pattern already includes 045-kafka-users
		echo "$(GREEN)Planning $$dir...$(NC)"; \
		cd $$dir && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform plan -var-file=terraform.tfvars.secret; \
		else \
			terraform plan; \
		fi; \
	done

.PHONY: tf-destroy-apps
tf-destroy-apps: ## Destroy all applications (keeps infrastructure)
	@echo "$(RED)Destroying all applications...$(NC)"
	kubectl delete -k $(KUSTOMIZE_DIR)/services --recursive
	kubectl delete -k $(KUSTOMIZE_DIR)/frontends --recursive

.PHONY: tf-destroy-all
tf-destroy-all: ## Destroy everything (WARNING: This will delete everything!)
	@echo "$(RED)WARNING: This will destroy all infrastructure and data!$(NC)"
	@echo "Press Ctrl+C within 5 seconds to cancel..."
	@sleep 5
	@for dir in $$(ls -r $(TERRAFORM_DIR)/); do \
		echo "$(RED)Destroying $$dir...$(NC)"; \
		cd $(TERRAFORM_DIR)/$$dir && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform destroy -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform destroy -auto-approve; \
		fi; \
	done

#################################
# Swagger/Documentation
#################################

# Install swagger tools
.PHONY: install-swagger
install-swagger: ## Install swagger generation tools
	@echo "$(YELLOW)Installing swagger tools...$(NC)"
	go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger documentation for auth-service
.PHONY: swagger-auth
swagger-auth: ## Generate swagger documentation for auth-service
	@echo "$(YELLOW)Generating swagger documentation for auth-service...$(NC)"
	@cd cmd/auth-service && swag init -g main.go -o docs --parseDependency --parseInternal --parseDepth 2
	@echo "$(GREEN)Auth service swagger documentation generated$(NC)"

# Generate swagger documentation for core-manager
.PHONY: swagger-core
swagger-core: ## Generate swagger documentation for core-manager
	@echo "$(YELLOW)Generating swagger documentation for core-manager...$(NC)"
	@cd cmd/core-manager && swag init -g main.go -o docs --parseDependency --parseInternal --parseDepth 2
	@echo "$(GREEN)Core manager swagger documentation generated$(NC)"

# Generate swagger for all services
.PHONY: swagger
swagger: swagger-auth swagger-core ## Generate swagger documentation for all services
	@echo "$(GREEN)All swagger documentation generated$(NC)"

# Backwards compatibility alias
.PHONY: swagger-all
swagger-all: swagger ## Alias for swagger target

# Run the comprehensive documentation generation script
.PHONY: docs
docs: swagger ## Generate comprehensive API documentation
	@echo "$(YELLOW)Running comprehensive documentation generation...$(NC)"
	@if [ -f "$(SCRIPTS_DIR)/docs/generate-docs.sh" ]; then \
		$(SCRIPTS_DIR)/docs/generate-docs.sh; \
	else \
		echo "$(YELLOW)Documentation script not found, skipping$(NC)"; \
	fi

# Start swagger UI servers
.PHONY: swagger-ui
swagger-ui: ## Start Swagger UI, Redoc, and Swagger Editor
	@echo "$(YELLOW)Starting documentation servers...$(NC)"
	@if [ -f "deployments/docker-compose/docker-compose.swagger.yml" ]; then \
		docker-compose -f deployments/docker-compose/docker-compose.swagger.yml up -d; \
		echo "$(GREEN)Documentation servers started:$(NC)"; \
		echo "  • Swagger UI: http://localhost:8082"; \
		echo "  • Redoc: http://localhost:8083"; \
		echo "  • Swagger Editor: http://localhost:8084"; \
	else \
		echo "$(YELLOW)Creating swagger docker-compose file...$(NC)"; \
		$(MAKE) create-swagger-compose; \
		docker-compose -f deployments/docker-compose/docker-compose.swagger.yml up -d; \
	fi

# Create swagger docker-compose file if it doesn't exist
.PHONY: create-swagger-compose
create-swagger-compose: ## Create swagger docker-compose file
	@mkdir -p deployments/docker-compose
	@echo "version: '3.8'" > deployments/docker-compose/docker-compose.swagger.yml
	@echo "services:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "  swagger-ui:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    image: swaggerapi/swagger-ui" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    ports:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "      - \"8082:8080\"" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    environment:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "      SWAGGER_JSON: /docs/swagger.json" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    volumes:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "      - ./../../cmd/auth-service/docs:/docs" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "  redoc:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    image: redocly/redoc" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    ports:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "      - \"8083:80\"" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    environment:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "      SPEC_URL: /docs/swagger.json" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    volumes:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "      - ./../../cmd/auth-service/docs:/usr/share/nginx/html/docs" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "  swagger-editor:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    image: swaggerapi/swagger-editor" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    ports:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "      - \"8084:8080\"" >> deployments/docker-compose/docker-compose.swagger.yml

# Stop swagger UI servers
.PHONY: swagger-down
swagger-down: ## Stop documentation servers
	@echo "$(YELLOW)Stopping documentation servers...$(NC)"
	@if [ -f "deployments/docker-compose/docker-compose.swagger.yml" ]; then \
		docker-compose -f deployments/docker-compose/docker-compose.swagger.yml down; \
	fi

# Validate swagger specs
.PHONY: validate-swagger
validate-swagger: ## Validate swagger specifications
	@echo "$(YELLOW)Validating swagger specifications...$(NC)"
	@if [ -f "cmd/auth-service/docs/swagger.json" ]; then \
		echo "$(GREEN)Validating auth-service swagger...$(NC)"; \
		docker run --rm -v ${PWD}/cmd/auth-service/docs:/spec redocly/cli lint /spec/swagger.json || true; \
	fi
	@if [ -f "cmd/core-manager/docs/swagger.json" ]; then \
		echo "$(GREEN)Validating core-manager swagger...$(NC)"; \
		docker run --rm -v ${PWD}/cmd/core-manager/docs:/spec redocly/cli lint /spec/swagger.json || true; \
	fi

# Generate API documentation (HTML)
.PHONY: generate-api-docs
generate-api-docs: swagger ## Generate HTML API documentation
	@echo "$(YELLOW)Generating HTML API documentation...$(NC)"
	@mkdir -p docs/api
	@if [ -f "cmd/auth-service/docs/swagger.json" ]; then \
		docker run --rm -v ${PWD}:/app redocly/cli build-docs /app/cmd/auth-service/docs/swagger.json -o /app/docs/api/auth-service.html; \
		echo "$(GREEN)Auth service documentation generated at docs/api/auth-service.html$(NC)"; \
	fi
	@if [ -f "cmd/core-manager/docs/swagger.json" ]; then \
		docker run --rm -v ${PWD}:/app redocly/cli build-docs /app/cmd/core-manager/docs/swagger.json -o /app/docs/api/core-manager.html; \
		echo "$(GREEN)Core manager documentation generated at docs/api/core-manager.html$(NC)"; \
	fi

# Serve API documentation locally
.PHONY: serve-docs
serve-docs: ## Serve API documentation locally on port 8080
	@echo "$(YELLOW)Serving API documentation...$(NC)"
	@if command -v python3 > /dev/null; then \
		cd docs/api && python3 -m http.server 8080; \
	else \
		echo "$(RED)Python3 not found. Please install Python3 to serve docs locally.$(NC)"; \
	fi

# Clean swagger generated files
.PHONY: clean-swagger
clean-swagger: ## Clean swagger generated files
	@echo "$(YELLOW)Cleaning swagger files...$(NC)"
	rm -rf cmd/auth-service/docs
	rm -rf cmd/core-manager/docs
	rm -rf docs/api

# Quick documentation workflow
.PHONY: docs-quick
docs-quick: swagger swagger-ui ## Quick swagger generation and UI startup
	@echo "$(GREEN)Documentation ready at http://localhost:8082$(NC)"

# Generate and view documentation
.PHONY: docs-view
docs-view: generate-api-docs ## Generate and open HTML documentation
	@echo "$(GREEN)Opening documentation...$(NC)"
	@if [ -f "docs/api/auth-service.html" ]; then \
		if command -v xdg-open > /dev/null; then \
			xdg-open docs/api/auth-service.html; \
		elif command -v open > /dev/null; then \
			open docs/api/auth-service.html; \
		else \
			echo "$(YELLOW)Please open docs/api/auth-service.html in your browser$(NC)"; \
		fi \
	fi

#################################
# Kind Cluster Management
#################################
.PHONY: kind-create
kind-create: ## Create Kind cluster for development
	@echo "$(YELLOW)Creating Kind cluster using Terraform...$(NC)"
	cd deployments/terraform/environments/development/uk_dev/010-infrastructure && \
		terraform init && \
		terraform apply -auto-approve

.PHONY: kind-delete
kind-delete: ## Delete Kind cluster
	@echo "$(RED)Deleting Kind cluster...$(NC)"
	cd deployments/terraform/environments/development/uk_dev/010-infrastructure && \
		terraform destroy -auto-approve

.PHONY: kind-status
kind-status: ## Check Kind cluster status
	@echo "$(YELLOW)Kind cluster status:$(NC)"
	kind get clusters
	kubectl config use-context kind-personae-dev && kubectl get nodes

.PHONY: kind-load-images
kind-load-images: ## Load Docker images into Kind
	@echo "$(YELLOW)Loading images into Kind...$(NC)"
	@mkdir -p $(TMPDIR)
	kind load docker-image $(REGISTRY)/auth-service:$(IMAGE_TAG) --name personae-dev
	kind load docker-image $(REGISTRY)/core-manager:$(IMAGE_TAG) --name personae-dev
	kind load docker-image $(REGISTRY)/agent-chassis:$(IMAGE_TAG) --name personae-dev
	kind load docker-image $(REGISTRY)/reasoning-agent:$(IMAGE_TAG) --name personae-dev
	kind load docker-image $(REGISTRY)/web-search-adapter:$(IMAGE_TAG) --name personae-dev
	kind load docker-image $(REGISTRY)/web-scrape-adapter:$(IMAGE_TAG) --name personae-dev
	kind load docker-image $(REGISTRY)/git-adapter:$(IMAGE_TAG) --name personae-dev
	kind load docker-image $(REGISTRY)/image-generator-adapter:$(IMAGE_TAG) --name personae-dev
	kind load docker-image $(REGISTRY)/content-creator-agent:$(IMAGE_TAG) --name personae-dev

.PHONY: reload-auth-service
reload-auth-service: ## Rebuild and reload auth-service in Kind
	@echo "$(YELLOW)Rebuilding auth-service...$(NC)"
	@$(MAKE) build-auth-service
	@mkdir -p $(TMPDIR)
	kind load docker-image $(REGISTRY)/auth-service:$(IMAGE_TAG) --name personae-dev
	kubectl delete pod -n ai-persona-system -l app=auth-service
	@echo "$(GREEN)auth-service reloaded$(NC)"

.PHONY: reload-core-manager
reload-core-manager: ## Rebuild and reload core-manager in Kind
	@echo "$(YELLOW)Rebuilding core-manager...$(NC)"
	@$(MAKE) build-core-manager
	@mkdir -p $(TMPDIR)
	kind load docker-image $(REGISTRY)/core-manager:$(IMAGE_TAG) --name personae-dev
	kubectl delete pod -n ai-persona-system -l app=core-manager
	@echo "$(GREEN)core-manager reloaded$(NC)"

# Add a new helper target
.PHONY: kind-load-auth
kind-load-auth: ## Load auth-service image into Kind
	@mkdir -p $(TMPDIR)
	kind load docker-image auth-service:local --name personae-dev

.PHONY: kind-load-core
kind-load-core: ## Load core-manager image into Kind
	@mkdir -p $(TMPDIR)
	kind load docker-image core-manager:local --name personae-dev

#################################
# Environment Specific Helpers
#################################
.PHONY: use-dev-context
use-dev-context: ## Switch to development Kubernetes context
	kubectl config use-context kind-personae-dev

.PHONY: use-prod-context
use-prod-context: ## Switch to production Kubernetes context
	kubectl config use-context personae-$(REGION)-prod-cluster

#################################
# Secrets Management
#################################
.PHONY: create-dev-secrets
create-dev-secrets: ## Create all development secrets (personae-dev-secrets and docker-hub-creds)
	@echo "$(YELLOW)Creating development namespace...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl create namespace $(PROJECT_NAME) --dry-run=client -o yaml | KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -f -
	@echo "$(YELLOW)Creating personae-dev-secrets...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl create secret generic personae-dev-secrets \
		--from-literal=CLIENTS_DB_PASSWORD=$CLIENTS_DB_PASSWORD${} \
		--from-literal=TEMPLATES_DB_PASSWORD=$${TEMPLATES_DB_PASSWORD} \
		--from-literal=AUTH_DB_PASSWORD=$${AUTH_DB_PASSWORD} \
		--from-literal=MINIO_ACCESS_KEY=$${MINIO_ACCESS_KEY} \
		--from-literal=SECRET_KEY=$${SECRET_KEY} \
		--from-literal=JWT_SECRET_KEY=$${JWT_SECRET_KEY} \
		--from-literal=ANTHROPIC_API_KEY=$${ANTHROPIC_API_KEY} \
		--from-literal=SERP_API_KEY=$${SERP_API_KEY} \
		--from-literal=SCRAPING_BEE_API_KEY=$${SCRAPING_BEE_API_KEY} \
		--from-literal=FIRECRAWL_API_KEY=$${FIRECRAWL_API_KEY} \
		--from-literal=STABILITY_API_KEY=$${STABILITY_API_KEY:-not-a-real-key} \
		-n $(PROJECT_NAME) --dry-run=client -o yaml | KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -f -
	@echo "$(GREEN)✓ personae-dev-secrets created$(NC)"
	@echo "$(YELLOW)Creating docker-hub-creds secret...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl create secret docker-registry docker-hub-creds \
		--namespace=$(PROJECT_NAME) \
		--docker-server=docker.io \
		--docker-username="$$(echo $${DOCKER_USERNAME} | tr -d '"')" \
		--docker-password="$$(echo $${DOCKER_PASSWORD} | tr -d '"')" \
		--docker-email="$$(echo $${DOCKER_EMAIL} | tr -d '"')" \
		--dry-run=client -o yaml | KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -f -
	@echo "$(GREEN)✓ docker-hub-creds created$(NC)"
	@echo "$(GREEN)All development secrets created successfully!$(NC)"

# Delete all development secrets
.PHONY: delete-dev-secrets
delete-dev-secrets: ## Delete all development secrets
	@echo "$(YELLOW)Deleting development secrets...$(NC)"
	@kubectl delete secret personae-dev-secrets -n $(PROJECT_NAME) --ignore-not-found
	@kubectl delete secret docker-hub-creds -n $(PROJECT_NAME) --ignore-not-found
	@echo "$(GREEN)Development secrets deleted$(NC)"

# Verify all development secrets
.PHONY: verify-dev-secrets
verify-dev-secrets: ## Verify all development secrets exist
	@echo "$(YELLOW)Verifying development secrets...$(NC)"
	@kubectl get secret personae-dev-secrets -n $(PROJECT_NAME) -o name && echo "$(GREEN)✓ personae-dev-secrets exists$(NC)" || echo "$(RED)✗ personae-dev-secrets missing$(NC)"
	@kubectl get secret docker-hub-creds -n $(PROJECT_NAME) -o name && echo "$(GREEN)✓ docker-hub-creds exists$(NC)" || echo "$(RED)✗ docker-hub-creds missing$(NC)"
	@echo "$(YELLOW)Docker registry config:$(NC)"
	@kubectl get secret docker-hub-creds -n $(PROJECT_NAME) -o jsonpath='{.data.\.dockerconfigjson}' | base64 -d | jq -r '.auths."docker.io" | {username, email}' || true

#################################
# ConfigMap Management
#################################
.PHONY: create-dev-configs
create-dev-configs: ## Create development configmaps
	@echo "$(YELLOW)Creating development configmaps...$(NC)"
	kubectl create namespace ai-persona-system --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f deployments/kustomize/infrastructure/configs/development/configmap-dev.yaml -n ai-persona-system


#################################
# Workflow Monitoring
#################################
.PHONY: build-workflow-monitor
build-workflow-monitor: ## Build workflow-monitor image
	@echo "$(YELLOW)Building workflow-monitor...$(NC)"
	docker build -t $(REGISTRY)/workflow-monitor:$(IMAGE_TAG) \
		-f build/docker/backend/workflow-monitor.dockerfile .

.PHONY: push-workflow-monitor
push-workflow-monitor: ## Push workflow-monitor image
	docker push $(REGISTRY)/workflow-monitor:$(IMAGE_TAG)

# Quick monitoring commands
.PHONY: monitor-workflows
monitor-workflows: ## Run workflow monitor as a one-off command
	@echo "$(YELLOW)Checking workflow status...$(NC)"
	kubectl run workflow-monitor-$(shell date +%s) \
		--image=$(REGISTRY)/workflow-monitor:$(IMAGE_TAG) \
		--rm -it --restart=Never \
		-n $(PROJECT_NAME) \
		--env="DATABASE_URL=postgresql://clients_user:password@postgres-clients:5432/clients_db?sslmode=disable" \
		--env="CLIENT_ID=demo_client" \
		-- /workflow-monitor -stuck-hours=1

.PHONY: monitor-stuck
monitor-stuck: ## Check for stuck workflows
	@echo "$(YELLOW)Checking for stuck workflows...$(NC)"
	kubectl exec --request-timeout=5m -it postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -c \
		"SELECT correlation_id, current_step, status, \
		 EXTRACT(EPOCH FROM (NOW() - updated_at))/3600 as hours_stuck \
		 FROM orchestrator_state \
		 WHERE status IN ('RUNNING', 'AWAITING_RESPONSES') \
		 AND updated_at < NOW() - INTERVAL '1 hour' \
		 ORDER BY updated_at ASC;"

.PHONY: monitor-active
monitor-active: ## Show active workflows
	@echo "$(YELLOW)Active workflows:$(NC)"
	kubectl exec --request-timeout=5m -it postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -c \
		"SELECT correlation_id, current_step, status, \
		 execution_metadata->>'completed_steps' as completed, \
		 execution_metadata->>'total_steps' as total, \
		 ROUND(((execution_metadata->>'completed_steps')::numeric / \
		        NULLIF((execution_metadata->>'total_steps')::numeric, 0)) * 100, 1) as progress_pct \
		 FROM orchestrator_state \
		 WHERE status NOT IN ('COMPLETED', 'FAILED') \
		 ORDER BY updated_at DESC \
		 LIMIT 20;"

.PHONY: monitor-metrics
monitor-metrics: ## Show workflow metrics for last 24 hours
	@echo "$(YELLOW)Workflow metrics (24h):$(NC)"
	kubectl exec --request-timeout=5m -it postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -c \
		"SELECT \
		 COUNT(*) as total, \
		 COUNT(CASE WHEN status = 'COMPLETED' THEN 1 END) as completed, \
		 COUNT(CASE WHEN status = 'FAILED' THEN 1 END) as failed, \
		 COUNT(CASE WHEN status IN ('RUNNING', 'AWAITING_RESPONSES') THEN 1 END) as active, \
		 ROUND(100.0 * COUNT(CASE WHEN status = 'COMPLETED' THEN 1 END) / NULLIF(COUNT(*), 0), 1) as success_rate \
		 FROM orchestrator_state \
		 WHERE created_at > NOW() - INTERVAL '24 hours';"

# Add these targets to your Makefile

#################################
# Database Operations - Runtime Management
#################################

# Quick SQL execution for runtime changes
.PHONY: db-exec-templates
db-exec-templates: ## Execute SQL in templates DB
	@echo "$(YELLOW)Executing SQL in templates DB...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec --request-timeout=5m -it postgres-templates-0 -n $(PROJECT_NAME) -- \
		psql -U templates_user -d templates_db

.PHONY: db-exec-clients
db-exec-clients: ## Execute SQL in clients DB
	@echo "$(YELLOW)Executing SQL in clients DB...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec --request-timeout=5m -it postgres-clients-0 -n $(PROJECT_NAME) -- \
		psql -U clients_user -d clients_db

# Create new agent definition on the fly
.PHONY: agent-create
agent-create: ## Create a new agent definition (usage: make agent-create TYPE=analyzer NAME="Data Analyzer")
	@echo "$(YELLOW)Creating agent definition: $(TYPE)$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec --request-timeout=5m -it postgres-templates-0 -n $(PROJECT_NAME) -- psql -U templates_user -d templates_db -c "\
		INSERT INTO agent_definitions (type, display_name, description, category, default_config, capabilities) VALUES \
		('$(TYPE)', '$(NAME)', '$(DESC)', 'data-driven', \
		'{\"model\": \"claude-3-5-sonnet-20241022\", \"temperature\": 0.5, \"processing_mode\": \"task\", \
		  \"workflow\": {\"start_step\": \"process\", \"steps\": { \
		    \"process\": {\"action\": \"execute_llm_prompt\", \"next_step\": \"complete\"}, \
		    \"complete\": {\"action\": \"complete_workflow\"}}}}', \
		'[\"analysis\", \"$(TYPE)\"]'::jsonb) \
		ON CONFLICT (type) DO UPDATE SET \
		  display_name = EXCLUDED.display_name, \
		  updated_at = NOW() \
		RETURNING id, type, display_name;"

# Update agent configuration
.PHONY: agent-update-config
agent-update-config: ## Update agent config (usage: make agent-update-config TYPE=analyzer CONFIG='{"temperature": 0.7}')
	@echo "$(YELLOW)Updating agent config for: $(TYPE)$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec --request-timeout=5m -it postgres-templates-0 -n $(PROJECT_NAME) -- psql -U templates_user -d templates_db -c "\
		UPDATE agent_definitions \
		SET default_config = default_config || '$(CONFIG)'::jsonb, \
		    updated_at = NOW() \
		WHERE type = '$(TYPE)' \
		RETURNING type, default_config;"

# List all agent definitions
.PHONY: agent-list
agent-list: ## List all agent definitions
	@echo "$(YELLOW)Agent Definitions:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec --request-timeout=5m -it postgres-templates-0 -n $(PROJECT_NAME) -- psql -U templates_user -d templates_db -c "\
		SELECT type, display_name, category, \
		       array_length(capabilities::text[], 1) as cap_count, \
		       is_active, \
		       to_char(updated_at, 'YYYY-MM-DD HH24:MI') as last_updated \
		FROM agent_definitions \
		ORDER BY updated_at DESC;"

# Show agent performance
.PHONY: agent-performance
agent-performance: ## Show agent performance metrics
	@echo "$(YELLOW)Agent Performance Metrics:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec --request-timeout=5m -it postgres-templates-0 -n $(PROJECT_NAME) -- psql -U templates_user -d templates_db -c "\
		SELECT agent_type, \
		       total_tasks, \
		       ROUND(success_rate * 100, 1) || '%' as success_rate, \
		       avg_response_time_ms || 'ms' as avg_time, \
		       ROUND(avg_quality_score, 2) as quality \
		FROM agent_metrics \
		WHERE total_tasks > 0 \
		ORDER BY success_rate DESC;"

# Create agent group dynamically
.PHONY: group-create
group-create: ## Create agent group (usage: make group-create NAME="Analysis Team" TYPE=analysis)
	@echo "$(YELLOW)Creating agent group: $(NAME)$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec --request-timeout=5m -it postgres-templates-0 -n $(PROJECT_NAME) -- psql -U templates_user -d templates_db -c "\
		INSERT INTO agent_groups (name, group_type, agent_configs, orchestration_workflow) \
		VALUES ('$(NAME)', '$(TYPE)', \
		'[{\"role\": \"lead\", \"agent_type\": \"$(TYPE)-leader\"}, \
		  {\"role\": \"worker\", \"agent_type\": \"$(TYPE)-worker\"}]'::jsonb, \
		'{\"start_step\": \"validate\", \"steps\": {}}'::jsonb) \
		RETURNING id, name, group_type;"

# Hot reload agent configuration (notifies running agents)
.PHONY: agent-hot-reload
agent-hot-reload: ## Hot reload agent config (usage: make agent-hot-reload AGENT_ID=xxx CONFIG='{"key": "value"}')
	@echo "$(YELLOW)Hot reloading config for agent: $(AGENT_ID)$(NC)"
	@echo '{"type": "config_update", "agent_id": "$(AGENT_ID)", "config": $(CONFIG)}' | \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec --request-timeout=5m -i kafka-cluster-kafka-0 -n $(PROJECT_NAME) -- \
		/opt/kafka/bin/kafka-console-producer.sh \
		--broker-list localhost:9092 \
		--topic system.agent.$(AGENT_ID).control

# Test discovery functions
.PHONY: agent-discover
agent-discover: ## Test agent discovery (usage: make agent-discover CAPS="analysis,reporting")
	@echo "$(YELLOW)Discovering agents with capabilities: $(CAPS)$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec --request-timeout=5m -it postgres-templates-0 -n $(PROJECT_NAME) -- psql -U templates_user -d templates_db -c "\
		SELECT * FROM find_agents_by_capability('{$(CAPS)}'::text[], 'demo_client');"

# Recommend agents for task
.PHONY: agent-recommend
agent-recommend: ## Get agent recommendations (usage: make agent-recommend TASK=website-builder)
	@echo "$(YELLOW)Recommending agents for task: $(TASK)$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec --request-timeout=5m -it postgres-templates-0 -n $(PROJECT_NAME) -- psql -U templates_user -d templates_db -c "\
		SELECT agent_type, display_name, \
		       ROUND(performance_score * 100) || '%' as score, \
		       recommendation_reason \
		FROM recommend_agents_for_task('$(TASK)', NULL);"

# Quick agent spawn via API call
.PHONY: agent-spawn
agent-spawn: ## Spawn an agent instance (usage: make agent-spawn TYPE=analyzer CLIENT=demo_client)
	@echo "$(YELLOW)Spawning agent: $(TYPE) for client: $(CLIENT)$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl run spawn-agent-$(shell date +%s) --rm -i --restart=Never \
		--image=curlimages/curl -n $(PROJECT_NAME) -- \
		curl -X POST http://core-manager:8088/api/v1/agents/spawn \
		-H "Content-Type: application/json" \
		-d '{"agent_type": "$(TYPE)", "client_id": "$(CLIENT)", "spawn_job": true}'

# Monitor agent jobs
.PHONY: agent-jobs
agent-jobs: ## Show running agent jobs
	@echo "$(YELLOW)Running Agent Jobs:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl get jobs -n $(PROJECT_NAME) -l spawned-by=orchestrator \
		-o custom-columns=NAME:.metadata.name,TYPE:.metadata.labels.agent-type,STATUS:.status.conditions[0].type,AGE:.metadata.creationTimestamp

# Clean up completed agent jobs
.PHONY: agent-cleanup
agent-cleanup: ## Clean up completed agent jobs
	@echo "$(YELLOW)Cleaning up completed agent jobs...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl delete jobs -n $(PROJECT_NAME) -l spawned-by=orchestrator \
		--field-selector status.successful=1

# Add a specific target to just deploy/update the bootstrap agents
.PHONY: bootstrap-agents
bootstrap-agents: deploy-100-bootstrap-agents ## Deploy or update bootstrap agents

# Destroy bootstrap agents if needed
.PHONY: destroy-bootstrap-agents
destroy-bootstrap-agents: ## Destroy bootstrap agents
	@echo "$(RED)Destroying bootstrap agents...$(NC)"
	@cd $(TERRAFORM_DIR)/100-bootstrap-agents && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform destroy -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform destroy -auto-approve; \
		fi

# Check bootstrap agent status
.PHONY: bootstrap-status
bootstrap-status: ## Check status of bootstrap agents
	@echo "$(YELLOW)Bootstrap Agent Status:$(NC)"
	@kubectl get statefulset -n $(PROJECT_NAME) generic-orchestrator
	@echo "\n$(YELLOW)Bootstrap Agent Pods:$(NC)"
	@kubectl get pods -n $(PROJECT_NAME) -l app=generic-orchestrator
	@echo "\n$(YELLOW)Bootstrap Agent Logs (last 20 lines):$(NC)"
	@kubectl logs -n $(PROJECT_NAME) -l app=generic-orchestrator --tail=20

#################################
# Database Backups
#################################

.PHONY: deploy-database-backup
deploy-database-backup: ## Deploy database backup CronJob
	@echo "$(YELLOW)Deploying database backup CronJob...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/database-backup/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)CronJob deployed. Next run:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob database-backup

.PHONY: backup-now
backup-now: ## Trigger an immediate database backup (creates a Job from the CronJob)
	@echo "$(YELLOW)Triggering immediate backup...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) create job \
		--from=cronjob/database-backup \
		database-backup-manual-$$(date +%Y%m%d-%H%M%S)
	@echo "$(GREEN)Backup job created. Watch with:$(NC)"
	@echo "  make backup-logs"

.PHONY: backup-logs
backup-logs: ## Follow logs from the latest backup job
	@LATEST=$$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get pods \
		-l component=backup --sort-by=.metadata.creationTimestamp \
		-o jsonpath='{.items[-1].metadata.name}' 2>/dev/null); \
	if [ -n "$$LATEST" ]; then \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) logs -f "$$LATEST"; \
	else \
		echo "No backup pods found"; \
	fi

.PHONY: backup-status
backup-status: ## Show backup CronJob status and recent jobs
	@echo "$(YELLOW)CronJob status:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob database-backup
	@echo ""
	@echo "$(YELLOW)Recent jobs:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get jobs -l component=backup \
		--sort-by=.metadata.creationTimestamp | tail -5

.PHONY: backup-list-s3
backup-list-s3: ## List recent backups in S3
	@echo "$(YELLOW)Recent backups in S3:$(NC)"
	@B2_KEY_ID=$$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get secret personae-platform-secrets \
		-o jsonpath='{.data.B2_APPLICATION_KEY_ID}' | base64 -d); \
	B2_KEY=$$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get secret personae-platform-secrets \
		-o jsonpath='{.data.B2_APPLICATION_KEY}' | base64 -d); \
	AWS_ACCESS_KEY_ID=$$B2_KEY_ID AWS_SECRET_ACCESS_KEY=$$B2_KEY \
		aws s3 ls s3://personae-prod-uk001-backups/db-backups/ \
		--endpoint-url https://s3.us-east-005.backblazeb2.com \
		| tail -10



#################################
# bugs_open/ Staleness Sweep (RFC_005 §3.3)
#################################

.PHONY: deploy-bugs-open-staleness-sweep
deploy-bugs-open-staleness-sweep: ## Deploy the weekly bugs_open/ staleness sweep CronJob
	@echo "$(YELLOW)Deploying bugs-open-staleness-sweep CronJob...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/bugs-open-staleness-sweep/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)CronJob deployed. Next run:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob bugs-open-staleness-sweep

.PHONY: push-component-render-check
push-component-render-check: ## Push the component-render-check CronJob image
	docker push $(REGISTRY)/component-render-check:$(IMAGE_TAG)

.PHONY: deploy-component-render-check
deploy-component-render-check: ## Deploy the daily component-render-check CronJob (CGV-030)
	@echo "$(YELLOW)Deploying component-render-check CronJob...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/component-render-check/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)CronJob deployed. Next run:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob component-render-check

.PHONY: component-render-check-now
component-render-check-now: ## Trigger an immediate component-render-check run
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) create job \
		--from=cronjob/component-render-check \
		component-render-check-manual-$$(date +%Y%m%d-%H%M%S)

.PHONY: push-shared-output-fields-check
push-shared-output-fields-check: ## Push the shared-output-fields-check CronJob image
	docker push $(REGISTRY)/shared-output-fields-check:$(IMAGE_TAG)

.PHONY: push-optional-explicit-wires-check
push-optional-explicit-wires-check: ## Push the optional-explicit-wires-check CronJob image
	docker push $(REGISTRY)/optional-explicit-wires-check:$(IMAGE_TAG)

.PHONY: push-commit-sha-exposure-check
push-commit-sha-exposure-check: ## Push the commit-sha-exposure-check CronJob image
	docker push $(REGISTRY)/commit-sha-exposure-check:$(IMAGE_TAG)

.PHONY: push-live-declaration-drift-check
push-live-declaration-drift-check: ## Push the live-declaration-drift-check CronJob image
	docker push $(REGISTRY)/live-declaration-drift-check:$(IMAGE_TAG)

.PHONY: push-capped-schedule-ordering-check
push-capped-schedule-ordering-check: ## Push the capped-schedule-ordering-check CronJob image
	docker push $(REGISTRY)/capped-schedule-ordering-check:$(IMAGE_TAG)

.PHONY: push-finding-code-registry-check
push-finding-code-registry-check: ## Push the finding-code-registry-check CronJob image
	docker push $(REGISTRY)/finding-code-registry-check:$(IMAGE_TAG)

.PHONY: push-component-source-vocabulary-check
push-component-source-vocabulary-check: ## Push the component-source-vocabulary-check CronJob image
	docker push $(REGISTRY)/component-source-vocabulary-check:$(IMAGE_TAG)

.PHONY: push-removed-config-keys-check
push-removed-config-keys-check: ## Push the removed-config-keys-check CronJob image
	docker push $(REGISTRY)/removed-config-keys-check:$(IMAGE_TAG)

.PHONY: deploy-optional-explicit-wires-check
deploy-optional-explicit-wires-check: ## Deploy the daily optional-explicit-wires-check CronJob (RFC_029 §10.15 adoption gate)
	@echo "$(YELLOW)Deploying optional-explicit-wires-check CronJob...$(NC)"
	@echo "$(YELLOW)  The image MUST already be pushed at this tag. An absent image gives$(NC)"
	@echo "$(YELLOW)  ImagePullBackOff, which this fleet reports as a Job still RUNNING —$(NC)"
	@echo "$(YELLOW)  never FAILED. Build and push before deploying, not after.$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/optional-explicit-wires-check/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)CronJob deployed. Next run:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob optional-explicit-wires-check

.PHONY: deploy-capped-schedule-ordering-check
deploy-capped-schedule-ordering-check: ## Deploy the daily capped-schedule-ordering-check CronJob (bugs_open/316: a capped step sorting clock-replenished work statically)
	@echo "$(YELLOW)Deploying capped-schedule-ordering-check CronJob...$(NC)"
	@echo "$(YELLOW)  The image MUST already be pushed at this tag. An absent image gives$(NC)"
	@echo "$(YELLOW)  ImagePullBackOff, which this fleet reports as a Job still RUNNING —$(NC)"
	@echo "$(YELLOW)  never FAILED. Build and push before deploying, not after.$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/capped-schedule-ordering-check/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)CronJob deployed. Next run:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob capped-schedule-ordering-check

.PHONY: deploy-finding-code-registry-check
deploy-finding-code-registry-check: ## Deploy the daily finding-code-registry-check CronJob (bugs_open/358: an agent_error_log finding code firing with no declared disposition)
	@echo "$(YELLOW)Deploying finding-code-registry-check CronJob...$(NC)"
	@echo "$(YELLOW)  The image MUST already be pushed at this tag. An absent image gives$(NC)"
	@echo "$(YELLOW)  ImagePullBackOff, which this fleet reports as a Job still RUNNING —$(NC)"
	@echo "$(YELLOW)  never FAILED. Build and push before deploying, not after.$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/finding-code-registry-check/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)CronJob deployed. Next run:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob finding-code-registry-check
	@echo "$(YELLOW)  This target proves NOTHING about the image actually running. Read the artefact:$(NC)"
	@echo "$(YELLOW)  kubectl -n $(PROJECT_NAME) get cronjob finding-code-registry-check -o jsonpath='{.spec.jobTemplate.spec.template.spec.containers[0].image}'$(NC)"

.PHONY: deploy-component-source-vocabulary-check
deploy-component-source-vocabulary-check: ## Deploy the daily component-source-vocabulary-check CronJob (bugs_open/309: an ACTIVE component declaring a data source that resolves nowhere)
	@echo "$(YELLOW)Deploying component-source-vocabulary-check CronJob...$(NC)"
	@echo "$(YELLOW)  The image MUST already be pushed at this tag. An absent image gives$(NC)"
	@echo "$(YELLOW)  ImagePullBackOff, which this fleet reports as a Job still RUNNING —$(NC)"
	@echo "$(YELLOW)  never FAILED. Build and push before deploying, not after.$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/component-source-vocabulary-check/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)CronJob deployed. Next run:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob component-source-vocabulary-check

.PHONY: deploy-commit-sha-exposure-check
deploy-commit-sha-exposure-check: ## Deploy the daily commit-sha-exposure-check CronJob (standing form of 537's guard, bugs_closed/334)
	@echo "$(YELLOW)Deploying commit-sha-exposure-check CronJob...$(NC)"
	@echo "$(YELLOW)  The image MUST already be pushed at this tag. An absent image gives$(NC)"
	@echo "$(YELLOW)  ImagePullBackOff, which this fleet reports as a Job still RUNNING —$(NC)"
	@echo "$(YELLOW)  never FAILED. Build and push before deploying, not after.$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/commit-sha-exposure-check/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)CronJob deployed. Next run:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob commit-sha-exposure-check

.PHONY: deploy-shared-output-fields-check
deploy-shared-output-fields-check: ## Deploy the daily shared-output-fields-check CronJob (RFC_012 (d) online half)
	@echo "$(YELLOW)Deploying shared-output-fields-check CronJob...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/shared-output-fields-check/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)CronJob deployed. Next run:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob shared-output-fields-check

.PHONY: shared-output-fields-check-now
shared-output-fields-check-now: ## Trigger an immediate shared-output-fields-check run
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) create job \
		--from=cronjob/shared-output-fields-check \
		shared-output-fields-check-manual-$$(date +%Y%m%d-%H%M%S)

.PHONY: push-loop-sitewide-item-key-check
push-loop-sitewide-item-key-check: ## Push the loop-sitewide-item-key-check CronJob image
	docker push $(REGISTRY)/loop-sitewide-item-key-check:$(IMAGE_TAG)

.PHONY: deploy-loop-sitewide-item-key-check
deploy-loop-sitewide-item-key-check: ## Deploy the daily loop-sitewide-item-key-check CronJob (bugs_open/321)
	@echo "$(YELLOW)Deploying loop-sitewide-item-key-check CronJob...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/loop-sitewide-item-key-check/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)CronJob deployed. Next run:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob loop-sitewide-item-key-check

.PHONY: loop-sitewide-item-key-check-now
loop-sitewide-item-key-check-now: ## Trigger an immediate loop-sitewide-item-key-check run
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) create job \
		--from=cronjob/loop-sitewide-item-key-check \
		loop-sitewide-item-key-check-manual-$$(date +%Y%m%d-%H%M%S)

.PHONY: push-template-input-field-check
push-template-input-field-check: ## Push the template-input-field-check CronJob image
	docker push $(REGISTRY)/template-input-field-check:$(IMAGE_TAG)

.PHONY: deploy-template-input-field-check
deploy-template-input-field-check: ## Deploy the daily template-input-field-check CronJob (bugs_open/453, WFA-024)
	@echo "$(YELLOW)Deploying template-input-field-check CronJob...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/template-input-field-check/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)CronJob deployed. Next run:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob template-input-field-check

.PHONY: template-input-field-check-now
template-input-field-check-now: ## Trigger an immediate template-input-field-check run
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) create job \
		--from=cronjob/template-input-field-check \
		template-input-field-check-manual-$$(date +%Y%m%d-%H%M%S)

.PHONY: push-verifier-remit-check
push-verifier-remit-check: ## Push the verifier-remit-check CronJob image
	docker push $(REGISTRY)/verifier-remit-check:$(IMAGE_TAG)

.PHONY: deploy-verifier-remit-check
deploy-verifier-remit-check: ## Deploy the daily verifier-remit-check CronJob (WII-015, bugs_open/213 D3)
	@echo "$(YELLOW)Deploying verifier-remit-check CronJob...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/verifier-remit-check/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)CronJob deployed. Next run:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob verifier-remit-check

.PHONY: push-content-loss-check
push-content-loss-check: ## Push the content-loss-check CronJob image
	docker push $(REGISTRY)/content-loss-check:$(IMAGE_TAG)

.PHONY: deploy-content-loss-check
deploy-content-loss-check: ## Deploy the daily content-loss-check CronJob (bugs_open/355, RFC_042 option c)
	@echo "$(YELLOW)Deploying content-loss-check CronJob...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/content-loss-check/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)CronJob deployed. Next run:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob content-loss-check

.PHONY: content-loss-check-now
content-loss-check-now: ## Trigger an immediate content-loss-check run
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) create job \
		--from=cronjob/content-loss-check \
		content-loss-check-manual-$$(date +%Y%m%d-%H%M%S)

.PHONY: verifier-remit-check-now
verifier-remit-check-now: ## Trigger an immediate verifier-remit-check run
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) create job \
		--from=cronjob/verifier-remit-check \
		verifier-remit-check-manual-$$(date +%Y%m%d-%H%M%S)

.PHONY: push-brief-negation-check
push-brief-negation-check: ## Push the brief-negation-check CronJob image
	docker push $(REGISTRY)/brief-negation-check:$(IMAGE_TAG)

.PHONY: deploy-brief-negation-check
deploy-brief-negation-check: ## Deploy the daily brief-negation-check CronJob (bugs_open/305)
	@echo "$(YELLOW)Deploying brief-negation-check CronJob...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/brief-negation-check/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)CronJob deployed. Next run:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob brief-negation-check

.PHONY: brief-negation-check-now
brief-negation-check-now: ## Trigger an immediate brief-negation-check run
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) create job \
		--from=cronjob/brief-negation-check \
		brief-negation-check-manual-$$(date +%Y%m%d-%H%M%S)

.PHONY: bugs-open-staleness-sweep-now
bugs-open-staleness-sweep-now: ## Trigger an immediate sweep run (creates a Job from the CronJob)
	@echo "$(YELLOW)Triggering immediate bugs_open staleness sweep...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) create job \
		--from=cronjob/bugs-open-staleness-sweep \
		bugs-open-staleness-sweep-manual-$$(date +%Y%m%d-%H%M%S)
	@echo "$(GREEN)Job created. Watch with:$(NC)"
	@echo "  make bugs-open-staleness-sweep-logs"

.PHONY: bugs-open-staleness-sweep-logs
bugs-open-staleness-sweep-logs: ## Follow logs from the latest sweep job
	@LATEST=$$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get pods \
		-l app=bugs-open-staleness-sweep --sort-by=.metadata.creationTimestamp \
		-o jsonpath='{.items[-1].metadata.name}' 2>/dev/null); \
	if [ -n "$$LATEST" ]; then \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) logs -f "$$LATEST"; \
	else \
		echo "No sweep pods found"; \
	fi

.PHONY: deploy-instance-token-adoption-check
deploy-instance-token-adoption-check: ## Deploy the RFC_022 expiry tripwire for bugs_open/283 (CLC-016). ⚠ a TRIP is an owed review, not a defect — retire the job once it fires
	@echo "$(YELLOW)Deploying instance-token-adoption-check CronJob...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/instance-token-adoption-check/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)CronJob deployed. Next run:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob instance-token-adoption-check

.PHONY: deploy-concept-register-drift-check
deploy-concept-register-drift-check: ## Deploy the daily concept-register drift CronJob
	@echo "$(YELLOW)Deploying concept-register-drift-check CronJob...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/concept-register-drift-check/overlays/$(OVERLAY_PATH)
	@echo "$(GREEN)CronJob deployed. Next run:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get cronjob concept-register-drift-check

.PHONY: concept-register-drift-check-now
concept-register-drift-check-now: ## Trigger an immediate register drift check (creates a Job from the CronJob)
	@echo "$(YELLOW)Triggering immediate concept-register drift check...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) create job \
		--from=cronjob/concept-register-drift-check \
		concept-register-drift-check-manual-$$(date +%Y%m%d-%H%M%S)
	@echo "$(GREEN)Job created. Watch with:$(NC)"
	@echo "  make concept-register-drift-check-logs"

.PHONY: concept-register-drift-check-logs
concept-register-drift-check-logs: ## Follow logs from the latest register drift check job
	@LATEST=$$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get pods \
		-l app=concept-register-drift-check --sort-by=.metadata.creationTimestamp \
		-o jsonpath='{.items[-1].metadata.name}' 2>/dev/null); \
	if [ -n "$$LATEST" ]; then \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) logs -f "$$LATEST"; \
	else \
		echo "No drift-check pods found"; \
	fi



#################################
# Agent Image Management
#################################

# Sync agent_definitions.image_tag with the tag agent-chassis is RUNNING.
#
# bugs_open/066. There used to be FOUR copies of this UPDATE in this makefile
# (here, update-agent-images-v2, deploy-100-bootstrap-agents, and the inline one
# in the image-generator quick-deploy), every one of them unscoped —
# `SET image_repository = …, image_tag = …` with no WHERE at all. That rewrote
# is_snapshot rollback copies (021_model_swap_and_rollback.sql), resurrected
# soft-deleted rows, and would silently convert an agent that deliberately runs
# some OTHER image onto the chassis image. Measured 2026-07-27: 183 rows hit
# where 180 is correct.
#
# There is now ONE implementation, and it is scoped:
#   scripts/deploy/update-agent-images.sh
# It reads the tag from the live Deployment rather than from $(IMAGE_TAG),
# because the point of the record is what IS running — a make variable is a
# request, not an outcome.
#
# NOTE this is hygiene, not the 066 fix. The rows went four tags stale on
# 2026-07-24 even though deploy-agents already called this, because a sync is a
# property of one deploy PATH, not of the system: `kubectl apply -k` (the
# shortcut written at the foot of deploy-agents), `kubectl set image`
# (scripts/deploy/deploy-agents.sh) and `kubectl rollout undo` all move the
# cluster without it. The fix that closes the door is at spawn time —
# platform/orchestration/actions/agent_image.go.
.PHONY: update-agent-images
update-agent-images: ## Sync agent_definitions.image_tag with the running agent-chassis image
	@KUBECONFIG=$(KUBECONFIG_PATH) NAMESPACE=$(PROJECT_NAME) $(SCRIPTS_DIR)/deploy/update-agent-images.sh

# Kept as a name because deploy-agents and the quick-deploy targets call it.
.PHONY: update-agent-images-v2
update-agent-images-v2: update-agent-images ## Alias of update-agent-images (one implementation, see above)

.PHONY: check-agent-image-drift
check-agent-image-drift: ## Read-only: what the Deployment runs vs what the rows say vs what a spawn will use
	@KUBECONFIG=$(KUBECONFIG_PATH) NAMESPACE=$(PROJECT_NAME) $(SCRIPTS_DIR)/check-agent-image-drift.sh

# Update agent images and restart orchestrator
.PHONY: update-generic-orchestrator
update-generic-orchestrator: ## Update generic orchestrator image to current IMAGE_TAG
	@echo "$(YELLOW)Updating generic orchestrator to $(REGISTRY)/agent-chassis:$(IMAGE_TAG)...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) set image statefulset/generic-orchestrator \
		orchestrator=$(REGISTRY)/agent-chassis:$(IMAGE_TAG)
	@echo "$(GREEN)Waiting for rollout to complete...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) rollout status statefulset/generic-orchestrator --timeout=120s
	@echo "$(GREEN)Generic orchestrator updated to $(IMAGE_TAG)$(NC)"

.PHONY: restart-generic-orchestrator
restart-generic-orchestrator: ## Restart generic orchestrator pod
	@echo "$(YELLOW)Restarting generic orchestrator...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) delete pod generic-orchestrator-0
	@echo "$(GREEN)Waiting for pod to be ready...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) wait --for=condition=ready pod/generic-orchestrator-0 --timeout=120s
	@echo "$(GREEN)Generic orchestrator restarted$(NC)"

.PHONY: update-and-restart-orchestrator
update-and-restart-orchestrator: update-generic-orchestrator restart-generic-orchestrator ## Update and restart generic orchestrator
	@echo "$(GREEN)Generic orchestrator updated and restarted with $(REGISTRY)/agent-chassis:$(IMAGE_TAG)$(NC)"

.PHONY: sync-all-agents
sync-all-agents: update-agent-images-v2 update-generic-orchestrator ## Update database and generic orchestrator to same image
	@echo "$(YELLOW)Cleaning up old agent pods to force respawn with new image...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) delete jobs -l app=dynamic-agent 2>/dev/null || true
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) delete pods -l app=dynamic-agent 2>/dev/null || true
	@echo "$(GREEN)All agents synced to $(REGISTRY)/agent-chassis:$(IMAGE_TAG)$(NC)"

.PHONY: verify-agent-images
verify-agent-images: ## Verify all agent images are consistent
	@echo "$(YELLOW)Checking agent image versions...$(NC)"
	@echo "$(CYAN)Database agent definitions:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec --request-timeout=5m -i postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -t -c \
		"SELECT DISTINCT image_repository || ':' || image_tag as image FROM agent_definitions WHERE is_active = true;" 2>/dev/null || echo "Failed to query database"
	@echo "$(CYAN)Generic orchestrator:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get statefulset generic-orchestrator -o jsonpath='{.spec.template.spec.containers[0].image}' && echo
	@echo "$(CYAN)Running dynamic agents:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get pods -l app=dynamic-agent -o custom-columns=NAME:.metadata.name,IMAGE:.spec.containers[0].image 2>/dev/null || echo "No dynamic agents running"
	@echo "$(CYAN)Agent chassis deployment:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get deployment agent-chassis -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null && echo || echo "No agent-chassis deployment"
	@echo "$(CYAN)Local image provenance (docker labels, bugs_open/153):$(NC)"
	@docker image inspect $(REGISTRY)/agent-chassis:$(IMAGE_TAG) \
		--format 'agent-chassis:$(IMAGE_TAG)  revision={{index .Config.Labels "org.opencontainers.image.revision"}}  created={{.Created}}' \
		2>/dev/null || echo "agent-chassis:$(IMAGE_TAG) not present locally — label check skipped"
	@echo "$(CYAN)Pod binary provenance (agent-chassis):$(NC)"
	@# The `[ -n "$$S" ] && echo` tail is load-bearing, NOT a tidy-up: a pipeline's exit
	@# status is its LAST command's, so `... | grep | head -1` exits 0 even when grep
	@# matched nothing, and the `|| echo` below would never fire for the one case this
	@# check exists to catch — an UNSTAMPED binary would print a silent blank line.
	@# Council `editquality` caught this (corr 44fa6a98, round 1, medium). Capturing to a
	@# variable first makes the test the last command, so absence exits non-zero honestly.
	@# VERIFY a known sha; do NOT try to DISCOVER one. Two traps, both measured 2026-08-10:
	@#  1. `strings` is ABSENT from debian-slim images (browser-runner-adapter). Behind a
	@#     2>/dev/null it returns a silent 0 that is indistinguishable from "no stamp" — it
	@#     made a correctly-stamped service read as unstamped. So: `grep -a` on the binary.
	@#  2. Without `strings`' line boundaries there is nothing to anchor to, so a generic
	@#     "find the 40-hex string" grep matches Go's internal digit table
	@#     (0001020304050607...) and confidently returns the WRONG value on every service.
	@# Hence: ask whether the pod carries THIS ref's sha. EXPECT_SHA=<sha> to check another.
	@# /proc/1/exe resolves the running binary for any image base or binary path.
	@EXPECT_SHA=$${EXPECT_SHA:-$$(git rev-parse HEAD)}; \
	if KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) exec deploy/agent-chassis -- \
		grep -aq "$$EXPECT_SHA" /proc/1/exe 2>/dev/null; then \
		echo "  MATCH — pod binary was built from $$EXPECT_SHA"; \
	else \
		echo "  NO MATCH for $$EXPECT_SHA — the pod was built from a different commit,"; \
		echo "  or predates the 153 fix, or the exec failed. Check the startup log:"; \
		echo "    kubectl -n $(PROJECT_NAME) logs <pod> | grep 'build provenance'"; \
	fi
	@echo "$(CYAN)Pod imageID + startTime:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get pods -l app=agent-chassis \
		-o custom-columns=NAME:.metadata.name,IMAGE:.spec.containers[0].image,IMAGEID:.status.containerStatuses[0].imageID,STARTED:.status.startTime 2>/dev/null || true

.PHONY: quick-agent-update
quick-agent-update: ## Build, push and deploy agent-chassis + image-generator-adapter with current IMAGE_TAG
	@echo "$(YELLOW)Building agent-chassis:$(IMAGE_TAG)...$(NC)"
	@$(MAKE) build-agent-chassis IMAGE_TAG=$(IMAGE_TAG)
	@echo "$(YELLOW)Building image-generator-adapter:$(IMAGE_TAG)...$(NC)"
	@$(MAKE) build-image-generator-adapter IMAGE_TAG=$(IMAGE_TAG)
	@echo "$(YELLOW)Pushing agent-chassis:$(IMAGE_TAG)...$(NC)"
	@docker push $(REGISTRY)/agent-chassis:$(IMAGE_TAG)
	@echo "$(YELLOW)Pushing image-generator-adapter:$(IMAGE_TAG)...$(NC)"
	@docker push $(REGISTRY)/image-generator-adapter:$(IMAGE_TAG)
	@echo "Updating agent-chassis kustomization to $(IMAGE_TAG)...$(NC)"
	@sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' $(KUSTOMIZE_DIR)/services/agent-chassis/overlays/$(OVERLAY_PATH)/kustomization.yaml
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/agent-chassis/overlays/$(OVERLAY_PATH)
	@echo "Updating image-generator-adapter kustomization to $(IMAGE_TAG)...$(NC)"
	@sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' $(KUSTOMIZE_DIR)/services/image-generator-adapter/overlays/$(OVERLAY_PATH)/kustomization.yaml
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/image-generator-adapter/overlays/$(OVERLAY_PATH)
	@echo "$(YELLOW)Deploying...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/agent-chassis/overlays/$(OVERLAY_PATH)
	@echo "$(YELLOW)Updating database...$(NC)"
	@$(MAKE) update-agent-images-v2 IMAGE_TAG=$(IMAGE_TAG)
	@echo "$(YELLOW)Restarting pods...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl rollout restart deployment/agent-chassis -n ai-persona-system
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl rollout restart deployment/image-generator-adapter -n ai-persona-system
	@echo "$(GREEN)Deployment complete with $(REGISTRY)/agent-chassis:$(IMAGE_TAG) + $(REGISTRY)/image-generator-adapter:$(IMAGE_TAG)$(NC)"



# Add these targets to your Makefile

#################################
# PgBouncer Management
#################################

.PHONY: pgbouncer-status
pgbouncer-status: ## Show PgBouncer pod status
	@echo "$(YELLOW)PgBouncer Status:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get pods -l app=pgbouncer
	@echo ""
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get svc pgbouncer

.PHONY: pgbouncer-pools
pgbouncer-pools: ## Show PgBouncer pool statistics
	@echo "$(YELLOW)PgBouncer Pool Stats:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) exec deploy/pgbouncer -- \
		psql -p 6432 -U pgbouncer_admin pgbouncer -c "SHOW POOLS;"

.PHONY: pgbouncer-stats
pgbouncer-stats: ## Show PgBouncer server statistics
	@echo "$(YELLOW)PgBouncer Stats:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) exec deploy/pgbouncer -- \
		psql -p 6432 -U pgbouncer_admin pgbouncer -c "SHOW STATS;"

.PHONY: pgbouncer-clients
pgbouncer-clients: ## Show PgBouncer client connections
	@echo "$(YELLOW)PgBouncer Client Connections:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) exec deploy/pgbouncer -- \
		psql -p 6432 -U pgbouncer_admin pgbouncer -c "SHOW CLIENTS;"

.PHONY: pgbouncer-servers
pgbouncer-servers: ## Show PgBouncer server connections
	@echo "$(YELLOW)PgBouncer Server Connections:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) exec deploy/pgbouncer -- \
		psql -p 6432 -U pgbouncer_admin pgbouncer -c "SHOW SERVERS;"

.PHONY: pgbouncer-logs
pgbouncer-logs: ## Tail PgBouncer logs
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) logs -l app=pgbouncer -f

.PHONY: pgbouncer-restart
pgbouncer-restart: ## Restart PgBouncer
	@echo "$(YELLOW)Restarting PgBouncer...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) rollout restart deployment/pgbouncer
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) rollout status deployment/pgbouncer --timeout=60s
	@echo "$(GREEN)PgBouncer restarted$(NC)"

.PHONY: pgbouncer-test
pgbouncer-test: ## Test connectivity through PgBouncer to both databases
	@echo "$(YELLOW)Testing PgBouncer connectivity...$(NC)"
	@CLIENTS_PW=$$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get secret personae-platform-secrets -o jsonpath='{.data.CLIENTS_DB_PASSWORD}' | base64 -d) && \
	 KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) run pgb-test-$$(date +%s) --rm -i --restart=Never \
		--image=postgres:15-alpine -- \
		psql "postgresql://clients_user:$${CLIENTS_PW}@pgbouncer.$(PROJECT_NAME).svc.cluster.local:6432/clients_db?sslmode=disable" \
		-c "SELECT 'pgbouncer_ok' as status, current_database();"
	@echo "$(GREEN)PgBouncer connectivity test passed$(NC)"

.PHONY: pgbouncer-destroy
pgbouncer-destroy: ## Remove PgBouncer deployment
	@echo "$(RED)Removing PgBouncer...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) delete deployment pgbouncer --ignore-not-found
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) delete svc pgbouncer --ignore-not-found
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) delete configmap pgbouncer-config --ignore-not-found
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) delete secret pgbouncer-userlist --ignore-not-found
	@echo "$(GREEN)PgBouncer removed$(NC)"


#################################
# PostgreSQL Connection Management
#################################

.PHONY: db-check-connections
db-check-connections: ## Check PostgreSQL connection status and limits
	@echo "$(YELLOW)Checking PostgreSQL connection status...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -it postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -c \
		"SELECT 'max_connections' as setting, setting as value FROM pg_settings WHERE name = 'max_connections' \
		 UNION ALL \
		 SELECT 'current_connections', count(*)::text FROM pg_stat_activity WHERE datname = 'clients_db' \
		 UNION ALL \
		 SELECT 'idle_connections', count(*)::text FROM pg_stat_activity WHERE datname = 'clients_db' AND state = 'idle';"

.PHONY: db-connections-by-state
db-connections-by-state: ## Show connections grouped by state
	@echo "$(YELLOW)Connection breakdown by state:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -it postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -c \
		"SELECT state, count(*) as count \
		 FROM pg_stat_activity \
		 WHERE datname = 'clients_db' \
		 GROUP BY state \
		 ORDER BY count DESC;"

.PHONY: db-kill-idle-connections
db-kill-idle-connections: ## Terminate idle connections older than 10 minutes
	@echo "$(YELLOW)Terminating idle connections older than 10 minutes...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -it postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -c \
		"SELECT pg_terminate_backend(pid) \
		 FROM pg_stat_activity \
		 WHERE datname = 'clients_db' \
		   AND state = 'idle' \
		   AND state_change < now() - interval '10 minutes' \
		   AND pid <> pg_backend_pid();"

.PHONY: db-set-max-connections
db-set-max-connections: ## Increase max_connections (requires restart). Usage: make db-set-max-connections MAX=300
	@echo "$(YELLOW)Setting max_connections to $(MAX)...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -it postgres-clients-0 -n $(PROJECT_NAME) -- psql -U postgres -d clients_db -c \
		"ALTER SYSTEM SET max_connections = $(MAX);"
	@echo "$(RED)WARNING: PostgreSQL restart required for this to take effect$(NC)"
	@echo "Run: kubectl rollout restart statefulset/postgres-clients -n $(PROJECT_NAME)"


# ── Topic Cleanup ──────────────────────────────────────────────────────────
# Requires: port-forward to core-manager and a valid admin JWT token
# Usage: make cleanup-topics-dry TOKEN=eyJ...
#        make cleanup-topics TOKEN=eyJ...

cleanup-topics-dry:
	@echo "Dry-run topic cleanup..."
	@curl -s -X POST "http://localhost:8088/api/v1/admin/system/cleanup-topics?dry_run=true" \
		-H "Authorization: Bearer $(TOKEN)" | python3 -m json.tool

cleanup-topics:
	@echo "Running topic cleanup..."
	@curl -s -X POST "http://localhost:8088/api/v1/admin/system/cleanup-topics?batch_size=100" \
		-H "Authorization: Bearer $(TOKEN)" | python3 -m json.tool

cleanup-topics-all:
	@echo "Running topic cleanup (large batch)..."
	@curl -s -X POST "http://localhost:8088/api/v1/admin/system/cleanup-topics?batch_size=500" \
		-H "Authorization: Bearer $(TOKEN)" | python3 -m json.tool



# ── Admin Dashboard (API Gateway + SPA) ────────────────────────────────────
# ── Admin Dashboard ────────────────────────────────────────────────────────
# Follows the same build/push/deploy pattern as other services.
# Uses IMAGE_TAG (same version as core-manager, agents, etc.)

.PHONY: build-dashboard
build-dashboard: ## Build admin-dashboard Docker image
	@echo "$(YELLOW)Building admin-dashboard...$(NC)"
	docker build -t $(REGISTRY)/admin-dashboard:$(IMAGE_TAG) \
		-f frontends/admin-dashboard/Dockerfile frontends/admin-dashboard/
	@echo "Built $(REGISTRY)/admin-dashboard:$(IMAGE_TAG)"

.PHONY: push-dashboard
push-dashboard: ## Push admin-dashboard image
	@echo "$(YELLOW)Pushing admin-dashboard...$(NC)"
	docker push $(REGISTRY)/admin-dashboard:$(IMAGE_TAG)

.PHONY: deploy-dashboard
deploy-dashboard: ## Deploy admin-dashboard (updates image tag in kustomize)
	@echo "$(GREEN)Deploying admin-dashboard...$(NC)"
	@cd $(KUSTOMIZE_DIR)/services/admin-dashboard/overlays/$(OVERLAY_PATH) && \
		sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' kustomization.yaml && \
		rm -f kustomization.yaml.bak
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k \
		$(KUSTOMIZE_DIR)/services/admin-dashboard/overlays/$(OVERLAY_PATH)
	@echo "Dashboard deployed with tag $(IMAGE_TAG)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n ai-persona-system rollout status deployment/admin-dashboard

.PHONY: dashboard-logs
dashboard-logs: ## Tail admin-dashboard logs
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n ai-persona-system logs -l app=admin-dashboard --tail=30 -f

.PHONY: dashboard-port-forward
dashboard-port-forward: ## Port forward admin dashboard to localhost:8080
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n ai-persona-system port-forward svc/admin-dashboard 8080:8080

.PHONY: release-dashboard
release-dashboard: build-dashboard push-dashboard deploy-dashboard ## Build, push and deploy admin-dashboard

#################################
# Full Release (single command)
#
# pinned_sweep,<goals> — resolve $(REF) to ONE commit for the WHOLE release,
# then run <goals> in order underneath it.
#
# WHY THIS IS NOT JUST A PREREQUISITE LIST (bugs_open/249): `ref_build` resolves
# $(REF) afresh inside EVERY service's recipe, so with the default REF=HEAD each
# of the 14 builds asks git what HEAD is at the moment it starts. A release takes
# ~6 minutes and ~40 sessions commit to this one tree, so a release straddles
# whatever lands while it runs. Measured 2026-08-11: v1.0.1284 shipped THREE
# revisions under one tag — 5 services at 55fc8fc35, 1 at e2afedaaf, 8 at
# a41dec8e5, the cut points matching two other sessions' commit times to the
# second. Nothing was broken and nothing looked wrong; the tag was identical
# everywhere and each service's own provenance line was correct for itself.
# Resolving once, here, is what makes "one tag, one revision" TRUE rather than
# merely usual. An explicit REF=<sha> is resolved the same way, so a deliberate
# older-ref release is unaffected.
#
# SCOPE, stated because the echo would otherwise overclaim: this pins the 14
# BACKEND images (everything built through ref_build). `release-dashboard`
# builds from frontends/admin-dashboard/ in the working tree with no ref and no
# stamp, exactly as before — frontends are outside the provenance mechanism
# (BLD-019) and outside this guarantee.
#
# COST, so nobody rediscovers it as a fault: `make -n release` now prints THIS
# SWEEP rather than the docker commands underneath it, because the goals are
# reached through a shell loop instead of a prerequisite list. For the real
# preview use `make -n build-backend` (or `make -n build-<service>`), which is
# unchanged. This is deliberate — the `+` prefix would restore the old preview
# by making the sub-makes run under -n and rely on MAKEFLAGS carrying -n down.
# On an estate where the owner drives releases by hand, a preview command that
# performs a real release if that assumption ever fails is not a trade worth
# taking for tidier output.
#################################
define pinned_sweep
@PINNED=$$(git rev-parse --verify --quiet '$(REF)^{commit}'); \
if [ -z "$$PINNED" ]; then \
	echo "$(RED)REF='$(REF)' is not a commit — a release must name committed state.$(NC)"; \
	exit 1; \
fi; \
echo "$(GREEN)Release pinned to $$PINNED — every BACKEND service in this sweep builds from that one commit (bugs_open/249).$(NC)"; \
for goal in $(1); do \
	echo "$(CYAN)→ make $$goal REF=$$PINNED$(NC)"; \
	$(MAKE) --no-print-directory $$goal REF=$$PINNED || exit 1; \
done
endef

.PHONY: release
release: ## Full release: build, push, deploy everything
	$(call pinned_sweep,build-backend push-backend deploy-core deploy-agents deploy-agent-cleanup release-dashboard)
	@$(MAKE) --no-print-directory release-record
	@echo "$(GREEN)Full release complete with image tag $(IMAGE_TAG)$(NC)"
	@echo "$(YELLOW)Usage: make release IMAGE_TAG=v1.0.xxx ENVIRONMENT=production REGION=uk001$(NC)"

.PHONY: release-backend
release-backend: ## Release backend only (no dashboard)
	$(call pinned_sweep,build-backend push-backend deploy-core deploy-agents deploy-agent-cleanup)
	@$(MAKE) --no-print-directory release-record
	@echo "$(GREEN)Backend release complete with image tag $(IMAGE_TAG)$(NC)"

#################################
# release-record — put the release back INTO git (owner ruling 2026-09-03)
#
# WHY. `deploy-*` rewrites every overlay's `newTag:` with `sed -i` and the
# makefile's IMAGE_TAG is bumped by hand, so a finished release leaves its own
# identity ONLY in the working tree. Measured 2026-09-03: the fleet was running
# `v1.0.1356` while `git show HEAD:` on the chassis overlay said `v1.0.1353` —
# so the running tag existed NOWHERE in git, and the two sanctioned ways to ask
# "did my fix ship?" both failed: the startup stamp had already rotated, and
# there was no committed tag bump to run `git merge-base --is-ancestor` against.
# A release that cannot be located in history is not auditable after its logs
# rotate, which on the chassis is minutes (bugs_open/338 §9, LANDMINES).
#
# ⚠ PATHSPEC, NOT `git add -A`. This runs on a working tree shared by many
# sessions. It commits ONLY tracked files it can name — the makefile and
# `kustomization.yaml` overlays — and never adds an untracked file, so a
# half-finished change belonging to another lane cannot ride along. It takes
# them from the working tree and IGNORES THE INDEX, which is what stops another
# session's staged work being swept in (CLAUDE.md, "Git — commit per task").
#
# ⚠ NON-FATAL BY DESIGN, AND LOUD. It runs AFTER the deploys have succeeded, so
# the images are already live: failing the release at that point would report a
# problem that has not happened. It prints a RED banner with the exact command
# to run by hand instead — deliberately not a silent `|| true`.
#
# Runnable standalone, which is how you record a release that already went out:
#   make release-record IMAGE_TAG=v1.0.1356 REF=<the commit it was built from>
#################################
.PHONY: release-record
release-record: ## Commit the tag bumps a release leaves in the tree (pathspec-scoped, never `add -A`)
	@PINNED=$$(git rev-parse --verify --quiet '$(REF)^{commit}'); 	if [ -z "$$PINNED" ]; then 		echo "$(RED)release-record: REF='$(REF)' is not a commit — cannot record what was built.$(NC)"; 		exit 0; 	fi; 	FILES=$$(git diff --name-only -- makefile $(KUSTOMIZE_DIR) | grep -E '(^makefile$$|kustomization\.yaml$$)' || true); 	if [ -z "$$FILES" ]; then 		echo "$(GREEN)release-record: no tag bumps in the tree — nothing to record.$(NC)"; 		exit 0; 	fi; 	echo "$(CYAN)release-record: committing $$(echo "$$FILES" | wc -l) release artefact(s) for $(IMAGE_TAG), built from $$PINNED$(NC)"; 	git commit $$FILES 		-m "release $(IMAGE_TAG): record the tag bumps this release left in the tree" 		-m "Built from $$PINNED. Written by 'make release-record' so the running tag exists in git: a release whose overlays are only ever dirty cannot be located in history, and the chassis startup stamp rotates within minutes, so there is otherwise nothing to run 'git merge-base --is-ancestor' against (owner ruling 2026-09-03, bugs_open/338 section 9)." 		-m "Pathspec-scoped to the makefile and kustomization.yaml overlays; no untracked file is added and the index is ignored, so no other session's work rides along." 	|| { 		echo "$(RED)=============================================================$(NC)"; 		echo "$(RED)release-record: THE COMMIT FAILED. The release itself is FINE$(NC)"; 		echo "$(RED)and the images are live — only the git record is missing.$(NC)"; 		echo "$(RED)Run this by hand (pathspec kept, so it stays safe):$(NC)"; 		echo "$(YELLOW)  git commit $$(echo $$FILES | tr '\n' ' ') -m 'release $(IMAGE_TAG): record tag bumps (built from $$PINNED)'$(NC)"; 		echo "$(RED)=============================================================$(NC)"; 	}

.PHONY: deploy-services
deploy-services: deploy-core deploy-agents deploy-agent-cleanup deploy-dashboard ## Deploy all services (no build, images must exist)
	@echo "$(GREEN)All services deployed$(NC)"

.PHONY: dev-dashboard
dev-dashboard: ## Run Vite dev server for local dashboard development
	cd frontends/admin-dashboard && npm install && npm run dev


#################################
# GitHub Actions Runner (Self-Hosted)
#################################
.PHONY: build-github-runner
build-github-runner: ## Build github-actions-runner image (committed HEAD; REF=<ref> to pin)
	$(call ref_build,github-actions-runner)
# Was a bare `docker build ... .` until 2026-08-18, i.e. the whole shared working
# tree as build context — the pattern inverted for every other backend service on
# 2026-07-17. It only COPYs one tracked file
# (build/docker/backend/github-actions-runner-entrypoint.sh), which git archive
# includes, so ref_build is a drop-in and the image can no longer carry another
# session's WIP. Changed here because the runner is now a release image
# (RELEASE_IMAGES), and a release image built from the working tree would
# reintroduce exactly the blast radius the ref_build default exists to remove.

.PHONY: push-github-runner
push-github-runner: ## Push github-actions-runner image
	@echo "$(YELLOW)Pushing github-actions-runner...$(NC)"
	docker push $(REGISTRY)/github-actions-runner:$(IMAGE_TAG)

.PHONY: deploy-github-runner
deploy-github-runner: ## Deploy github-actions-runner
	@echo "$(YELLOW)Updating github-actions-runner image tag to $(IMAGE_TAG)...$(NC)"
	@cd $(KUSTOMIZE_DIR)/services/github-actions-runner/overlays/$(OVERLAY_PATH) && \
		sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' kustomization.yaml && \
		rm -f kustomization.yaml.bak
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k \
		$(KUSTOMIZE_DIR)/services/github-actions-runner/overlays/$(OVERLAY_PATH)
	@echo "Runner deployed with tag $(IMAGE_TAG)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n ai-persona-system rollout status deployment/github-actions-runner

.PHONY: release-github-runner
release-github-runner: build-github-runner push-github-runner deploy-github-runner ## Build, push and deploy github-actions-runner

# deploy-github-runners (PLURAL) — apply BOTH runners' manifests, WITHOUT touching
# their image tags. This is what `deploy-agents` calls, so `make release` ships
# runner pod-spec changes the same way it ships every other service's.
#
# ⚠ SUPERSEDED 2026-08-18 (OWNER RULING, bugs_open/237 Decision B) — this target
# no longer decides anything about image tags, and the paragraph below is kept
# only because it explains what the estate looked like before the ruling.
# `github-actions-runner` is now in RELEASE_IMAGES (so build-backend/push-backend
# DO build and push it, from committed HEAD via ref_build) and both runners are
# in AGENT_DEPLOY_SERVICES, with -vmsites mapped to the shared image. So the
# retag now happens in the deploy-agents loop like every other service, and this
# target survives ONLY for its production-strict missing-overlay check.
#
# (historical, pre-ruling) Why it did not sed newTag, when every other block in
# deploy-agents does: the runners did NOT share the platform's image lineage.
# `build-backend` and `push-backend` never built or pushed
# `github-actions-runner` (it had its own build-/push-/release-github-runner
# targets), and the two overlays were pinned to DIFFERENT tags — measured
# 2026-08-13, live: github-actions-runner on v1.0.948 and
# github-actions-runner-vmsites on v1.0.1126, against a platform IMAGE_TAG of
# v1.0.1295.
#
# ⚠ CORRECTED 2026-08-17 (OWNER RULING): this comment used to say those two tags
# differed "on purpose". They do NOT — the owner has ruled the separate cadence
# is NOT intended, and the divergence is FUNCTIONAL, not cosmetic: the 2026-07-16
# dockerfile change added openssh-client and rsync, so v1.0.1126 (-vmsites) HAS
# them and v1.0.948 (github-actions-runner) does NOT — verified at both pods with
# git/jq as controls. The freeze has two different causes: `github-actions-runner`
# has `release-github-runner` and nobody has run it since April; `-vmsites` has NO
# retag target in this file at all, so nothing can move it. ~~The fix is pending
# an owner decision — see bugs_open/237.~~ RULED 2026-08-18: fold both into the
# fleet release (above).
#
# The old warning here read: *Do NOT "tidy" this by seding both tags to
# IMAGE_TAG: retagging them to IMAGE_TAG on release would point both at an image
# that was never built and never pushed, and they would ImagePullBackOff
# together.* **That warning was correct and its premise is now gone** — the
# release builds and pushes the image, so the sed is safe. Note the two halves
# are one change: if anyone removes github-actions-runner from RELEASE_IMAGES (or
# drops build-github-runner from build-backend) while leaving the runners in
# AGENT_DEPLOY_SERVICES, the old trap comes straight back and takes out CI on
# both runners at once. check-release-coverage does NOT catch that direction —
# it fails a service that pins a release-built image and is in no release path,
# not one that is in a release path but whose image stopped being built.
#
# To move a runner to a new image, use `make release-github-runner`, which builds
# and pushes at IMAGE_TAG BEFORE deploy-github-runner (singular) rewrites the tag.
# That ordering is what makes the singular target safe and this one unnecessary
# for image changes.
#
# No rollout restart on purpose: kubectl apply already rolls a Deployment whose
# pod template changed, and a forced restart would interrupt an in-flight CI job
# on every release for no gain.
#
# Missing-overlay handling is deliberately asymmetric, because the two cases mean
# opposite things. The runners are PRODUCTION-ONLY — unlike core-manager,
# reasoning-agent and others, they have no overlays/development — so under
# ENVIRONMENT=development a missing overlay is normal and must not fail the
# deploy. In production it is not normal: it means a release is about to report
# success while shipping none of the manifests it claims to, which is the exact
# failure this target exists to end. So: skip LOUDLY off-production, fail in
# production. A failed `kubectl apply` is always fatal — swallowing that would
# recreate the same silence one layer down.
.PHONY: deploy-github-runners
deploy-github-runners: ## Apply both github runner manifests (pod spec only; image tags left pinned)
	@echo "$(YELLOW)Applying github runner manifests (image tags left as pinned)...$(NC)"
	@APPLIED=0; SKIPPED=0; \
	for svc in github-actions-runner github-actions-runner-vmsites; do \
		OVERLAY="$(KUSTOMIZE_DIR)/services/$$svc/overlays/$(OVERLAY_PATH)"; \
		if [ -d "$$OVERLAY" ]; then \
			echo "  $$svc"; \
			KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k "$$OVERLAY" || exit 1; \
			APPLIED=$$((APPLIED+1)); \
		elif [ "$(ENVIRONMENT)" = "production" ]; then \
			echo "$(RED)  MISSING in production: $$OVERLAY$(NC)"; \
			echo "$(RED)  Refusing to report a release that did not ship these manifests.$(NC)"; \
			exit 1; \
		else \
			echo "$(YELLOW)  skipping $$svc — no $(ENVIRONMENT) overlay (runners are production-only)$(NC)"; \
			SKIPPED=$$((SKIPPED+1)); \
		fi; \
	done; \
	if [ "$$APPLIED" -gt 0 ]; then \
		echo "$(GREEN)Runner manifests applied: $$APPLIED (skipped $$SKIPPED)$(NC)"; \
	else \
		echo "$(YELLOW)No runner manifests applied — all $$SKIPPED skipped for ENVIRONMENT=$(ENVIRONMENT)$(NC)"; \
	fi

#################################
# Node configuration (bugs_open/252)
#################################
# deploy-node-config — apply the node-config DaemonSet, which sets each node's
# kubelet image-GC thresholds (85/80/0s -> 70/60/168h) by editing
# /var/lib/kubelet/config.yaml and restarting kubelet, idempotently.
#
# Why node files and not the kubelet-config ConfigMap: that CM is
# PROVIDER-PROTECTED on this hosted control plane — writes return 200 "patched"
# and revert before the next read (measured 2026-08-14, three write shapes, no
# mutating webhook to explain it). The DaemonSet is also what makes the setting
# survive SPOT node replacement, which a one-off hand edit would not.
# Full mechanism + failure direction: the DaemonSet manifest's header comment.
#
# Like the runners: manifest-only, no image-tag sweep (busybox, own lineage),
# production-only overlay, called from deploy-agents so `make release` ships it.
.PHONY: deploy-node-config
deploy-node-config: ## Apply the node-config DaemonSet (kubelet image-GC settings, per node)
	@OVERLAY="$(KUSTOMIZE_DIR)/services/node-config/overlays/$(OVERLAY_PATH)"; \
	if [ -d "$$OVERLAY" ]; then \
		echo "$(YELLOW)Applying node-config DaemonSet...$(NC)"; \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k "$$OVERLAY" || exit 1; \
		echo "$(GREEN)node-config applied — verify with: make node-config-status$(NC)"; \
	elif [ "$(ENVIRONMENT)" = "production" ]; then \
		echo "$(RED)MISSING in production: $$OVERLAY$(NC)"; exit 1; \
	else \
		echo "$(YELLOW)skipping node-config — no $(ENVIRONMENT) overlay (production-only)$(NC)"; \
	fi

.PHONY: node-config-status
node-config-status: ## Show node-config pods and each node's LIVE kubelet GC values
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n ai-persona-system get pods -l app=node-config -o wide
	@echo "$(YELLOW)live kubelet values per node (the DS is proven at the kubelet, not at its own logs):$(NC)"
	@for n in $$(KUBECONFIG=$(KUBECONFIG_PATH) kubectl get nodes -o jsonpath='{.items[*].metadata.name}'); do \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl get --raw "/api/v1/nodes/$$n/proxy/configz" 2>/dev/null \
		| python3 -c "import json,sys;d=json.load(sys.stdin)['kubeletconfig'];print('  $$n'[-10:],'high',d.get('imageGCHighThresholdPercent'),'low',d.get('imageGCLowThresholdPercent'),'maxAge',d.get('imageMaximumGCAge'))"; \
	done

.PHONY: github-runner-logs
github-runner-logs: ## Tail github-actions-runner logs
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n ai-persona-system logs -l app=github-actions-runner --tail=30 -f

.PHONY: github-runner-status
github-runner-status: ## Show github-actions-runner pod status
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n ai-persona-system get pods -l app=github-actions-runner

.PHONY: github-runner-restart
github-runner-restart: ## Restart github-actions-runner
	@echo "$(YELLOW)Restarting github-actions-runner...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n ai-persona-system rollout restart deployment/github-actions-runner
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n ai-persona-system rollout status deployment/github-actions-runner --timeout=120s
	@echo "$(GREEN)Runner restarted$(NC)"


#################################
# Site chat box (Mythic Beasts VM) — DELIBERATELY NOT part of `release`
#
# The chat box is a plain Go binary on a VM, not a k8s service, so `make
# release` does not carry it and must not: different machine, different
# credential (an ssh key, not the kubeconfig), different blast radius. A
# customer-facing bot should never roll as a side effect of a fleet deploy,
# and a fleet deploy should never fail because an ssh key expired.
#
# These targets exist because the opposite failure is worse and already
# happened (2026-08-18): the deploy lived only in a runbook, so a change was
# committed believing `make release` would ship it. It would not have. An
# invisible deploy path is one nobody remembers to run.
#
#   make box-release     test → build from committed HEAD → push → install → verify
#   make box-status      what is ACTUALLY running on the box (md5, units, binds)
#   make box-verify      compare the box binary against a fresh local build
#
# box-build builds from COMMITTED HEAD through `git archive`, the same rule the
# backend follows, so it cannot bundle another session's WIP on this shared
# tree. box-build-tree is the opt-in escape hatch.
#
# One binary serves every site on the box: webdesign-chat.service plus one
# sitechat@<domain>.service per additional site. They all exec
# /usr/local/bin/sitechat, so a roll restarts all of them.
#################################

BOX_HOST ?= webdesign.vs.mythic-beasts.com
BOX_USER ?= root
BOX_KEY  ?= $(HOME)/.ssh/webdesign_box_ed25519
BOX_SRC  := docs/agent_docs/docs024_key_docs_latest/webdesign_uk_build_service/box/chat-service
BOX_OUT  := $(CURDIR)/dist/sitechat
BOX_SSH   = ssh -i $(BOX_KEY) $(BOX_USER)@$(BOX_HOST)

.PHONY: box-test
box-test: ## Run the site-chat box tests
	@cd $(BOX_SRC) && GOPROXY=off GOTOOLCHAIN=local go test . -count=1

.PHONY: box-build
box-build: box-test ## Build the sitechat binary from committed HEAD (no WIP can enter)
	@git rev-parse --verify --quiet '$(REF)^{commit}' >/dev/null || \
		{ echo "$(RED)REF='$(REF)' is not a commit.$(NC)"; exit 1; }
	@echo "$(GREEN)Building sitechat from committed ref $(REF) = $$(git rev-parse --short $(REF)) — working tree NOT included.$(NC)"
	@UNSHIPPED=$$(git status --porcelain -- $(BOX_SRC) 2>/dev/null | wc -l); \
	if [ "$$UNSHIPPED" -gt 0 ] && [ "$(REF)" = "HEAD" ]; then \
		echo "$(YELLOW)  $$UNSHIPPED uncommitted change(s) under $(BOX_SRC) are NOT in this binary:$(NC)"; \
		git status --porcelain -- $(BOX_SRC); \
	fi
	@mkdir -p $(dir $(BOX_OUT))
	@CTX=$$(mktemp -d /tmp/box-ctx.XXXXXX) && \
	trap 'rm -rf "$$CTX"' EXIT && \
	git archive $(REF) | tar -x -C "$$CTX" && \
	cd "$$CTX/$(BOX_SRC)" && \
	GOPROXY=off GOTOOLCHAIN=local GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -trimpath -ldflags "-X main.buildCommit=$$(cd $(CURDIR) && git rev-parse '$(REF)^{commit}')" -o "$(BOX_OUT)" . && \
	echo "$(GREEN)Built $(BOX_OUT) ($$(stat -c%s "$(BOX_OUT)") bytes, md5 $$(md5sum "$(BOX_OUT)" | cut -d' ' -f1))$(NC)"

.PHONY: box-build-tree
box-build-tree: ## Build sitechat from the WORKING TREE, WIP and all (opt-in escape hatch)
	@echo "$(YELLOW)Building sitechat from the WORKING TREE — uncommitted code WILL ship.$(NC)"
	@mkdir -p $(dir $(BOX_OUT))
	@cd $(BOX_SRC) && GOPROXY=off GOTOOLCHAIN=local GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -trimpath -ldflags "-X main.buildCommit=$$(git rev-parse --short HEAD)-tree" -o "$(BOX_OUT)" . && \
		echo "$(GREEN)Built $(BOX_OUT) (md5 $$(md5sum "$(BOX_OUT)" | cut -d' ' -f1))$(NC)"

.PHONY: box-push
box-push: ## Copy the built sitechat binary to the box (no restart)
	@test -f $(BOX_OUT) || { echo "$(RED)$(BOX_OUT) missing — run 'make box-build' first.$(NC)"; exit 1; }
	scp -i $(BOX_KEY) $(BOX_OUT) $(BOX_USER)@$(BOX_HOST):/root/sitechat

.PHONY: box-deploy
box-deploy: ## Install the pushed binary, restart every site instance, verify
	@echo "$(YELLOW)Installing and restarting all sitechat instances on $(BOX_HOST)...$(NC)"
	@$(BOX_SSH) 'install -m 755 /root/sitechat /usr/local/bin/sitechat && \
		systemctl restart webdesign-chat.service "sitechat@*.service" 2>/dev/null; \
		sleep 2; \
		echo "--- md5 ---"; md5sum /usr/local/bin/sitechat; \
		echo "--- journal ---"; journalctl -u webdesign-chat -n 5 --no-pager -o short-iso; \
		echo "--- binds (must be 127.0.0.1 only) ---"; ss -ltnp | grep sitechat'
	@$(MAKE) --no-print-directory box-verify

.PHONY: box-verify
box-verify: ## Prove the box is running the binary you just built (md5, both ends)
	@test -f $(BOX_OUT) || { echo "$(RED)$(BOX_OUT) missing — nothing to compare against.$(NC)"; exit 1; }
	@LOCAL=$$(md5sum $(BOX_OUT) | cut -d' ' -f1); \
	REMOTE=$$($(BOX_SSH) 'md5sum /usr/local/bin/sitechat' 2>/dev/null | cut -d' ' -f1); \
	echo "  local md5   $$LOCAL"; echo "  box   md5   $$REMOTE"; \
	if [ -z "$$REMOTE" ]; then echo "$(RED)Could not read the box binary — deploy NOT verified.$(NC)"; exit 1; \
	elif [ "$$LOCAL" != "$$REMOTE" ]; then echo "$(RED)MISMATCH — the box is NOT running this build.$(NC)"; exit 1; \
	else echo "$(GREEN)md5 MATCH — the box holds the file just pushed.$(NC)"; fi
	@echo "$(YELLOW)Now asking the RUNNING service what it was built from — md5 cannot answer that:$(NC)"
	@WANT=$$(git rev-parse '$(REF)^{commit}'); \
	GOT=$$($(BOX_SSH) "journalctl -u webdesign-chat --since '-3 min' --no-pager -o cat 2>/dev/null | grep -m1 'build provenance'" 2>/dev/null | sed 's/.*git_commit=//'); \
	echo "  want commit $$WANT"; \
	echo "  box  says   $${GOT:-<none: pre-stamp binary, or the restart is older than the window>}"; \
	if [ "$$GOT" = "$$WANT" ]; then echo "$(GREEN)PROVENANCE MATCH — the running service IS this commit.$(NC)"; \
	elif [ -z "$$GOT" ]; then echo "$(YELLOW)No provenance line — expected on the FIRST roll of the stamp only.$(NC)"; \
	else echo "$(RED)PROVENANCE MISMATCH — running $$GOT, wanted $$WANT.$(NC)"; exit 1; fi

.PHONY: box-release
box-release: box-build box-push box-deploy ## Full chat-box roll: test, build from HEAD, push, install, verify
	@echo "$(GREEN)Chat box released from $$(git rev-parse --short $(REF)).$(NC)"

.PHONY: box-status
box-status: ## Show what is actually running on the chat box
	@$(BOX_SSH) 'echo "--- binary ---"; md5sum /usr/local/bin/sitechat; ls -l /usr/local/bin/sitechat; \
		echo "--- units ---"; systemctl list-units "sitechat@*" webdesign-chat.service --no-pager --no-legend 2>/dev/null; \
		echo "--- binds ---"; ss -ltnp | grep sitechat'
