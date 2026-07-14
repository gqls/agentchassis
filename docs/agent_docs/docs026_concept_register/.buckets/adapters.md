
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Adapter/response message envelope contract (normative)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** 035 §1 "last verified against code 2026-06-11"; 003(8) now points here
- **what:** Any reply to a chassis request must be recognised as an awaited response or it silently falls to process-as-work (row stuck waiting, ~10-min retries, no error). Load-bearing field: in_response_to_request_id = incoming request_id (request_id fallback only; git adapter's reuse pattern favoured). Three header tiers (validator-enforced / coordinator-needed-but-unvalidated / observability). Body headers MUST be a typed struct with real bools (map[string]string string-bools fail unmarshal and drop the reply pre-claim — the multi-day thunder outage). Send via ProduceWithValidation. Request parsing: action from body, payload at body.data, accept reply-topic from three keys. Sibling race: local dispatches must preRegisterAwaitedRequest before send (confirmed fixed in prod 06-09).
- **sources:** 035 §1; 016 §9 bool-trap + race entries
- **relations:** awaited_requests; O(K²) batch presign
- **verify-later:** ValidateOutgoingMessage field list

<!-- SOURCE: U06_finetuning.md -->
### Adapter design guide (adapter vs agent vs inline; canonical structure)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** FOCUS_adapter_design(3) opening: "Canonical guide for building long-running cluster services… Examples drawn from the working adapters in the repository."
- **what:** The decision rule (one external API + multiple internal callers → adapter; long per-orchestration work → spawned agent; short single-agent call → inline; shared infra like DB/Kafka → nothing) plus the canonical shape: struct fields, ordered NewAdapter with manual cleanup on every failure path, sequential fetch-handle-loop (no goroutine-per-message by default), handleMessage parse/dispatch/respond, health endpoints, sync.Once shutdown, topic conventions (`system.adapter.<name>.requests` for new work), config YAML field-name traps, credentials from env only.
- **sources:** working/flywheel_docs/FOCUS_adapter_design(3).md (whole)
- **relations:** thunder adapter (the guide's newest example); response header tiers; deployment essentials
- **verify-later:** consistency of existing adapters with the guide

<!-- SOURCE: U06_finetuning.md -->
### Adapter response-header tier taxonomy and the validator-coverage gap
- **category:** adapters
- **status-signal:** partial
- **status-evidence:** FOCUS_adapter_design(3) "TODO — tighten validator coverage… Tier-2 fields are necessary for the orchestration to advance but not validated… Tracking issue: not yet filed."
- **what:** Response headers split into Tier 1 (five fields the platform Validator enforces; `is_error=true` bypasses), Tier 2 (what the chassis needs to route the reply to the awaiting orchestration — `in_response_to_request_id`, message_type, status vocabulary complete/error_recoverable/error_unrecoverable, is_complete/is_error, etc.; missing these means a silent AWAITING_RESPONSES hang the validator won't catch), Tier 3 (observability). Known live consequence of the gap: the matcher fix of 2026-05-22 (typed response-header struct so booleans serialise as real bools — a map[string]string sent string bools and the chassis dropped the reply). Proposal to extend the validator exists but is unfiled.
- **sources:** working/flywheel_docs/FOCUS_adapter_design(3).md#sending-responses,#todo; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics (matcher fix)
- **relations:** reply-topic derivation; send-before-register race (same stuck-await symptom family)
- **verify-later:** platform/validation/Validator current coverage

<!-- SOURCE: U06_finetuning.md -->
### Adapter deployment essentials (manifest, cluster resources, RBAC, Makefile)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** FOCUS_adapter_design(3) "Deployment essentials — Real lessons from deploying thunder-adapter Phase 2. Every item below is something the deployment failed without."
- **what:** The complete pre-flight for shipping an adapter: serviceAccountName + imagePullSecrets + `command:` (not `args:` — Dockerfiles use CMD, so args replaces the binary path), required Secrets/SA/Docker-Hub grants, explicit KafkaTopic CRDs (Strimzi auto-create is off; missing reply topics fail only at first response), Recreate strategy, single replica, RBAC trap (resourceNames supports no globs — scope by verbs instead for dynamic names like thunder-ssh-<uuid>), four Makefile insertion points and the newName/newTag overlay split, pre/post-deploy checklists.
- **sources:** working/flywheel_docs/FOCUS_adapter_design(3).md#deployment-essentials; working/flywheel_docs/FOCUS_finetuning_flywheel_changelog_addition.md (phase-2 deploy saga)
- **relations:** adapter design guide; wrong-binary image incident; debugging guide §10
- **verify-later:** thunder-adapter kustomize base vs the checklist

<!-- SOURCE: U08_travelling_docs.md -->
### Tier-4 browser-runner adapter (headless Chromium over Kafka)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "Stage 6 P0 DEPLOYED + SMOKE PASSED" 2026-07-11 (v1.0.1107; §2.15 smoke matched manual inspection; real bool headers; in_response_to_request_id matcher).
- **what:** A dedicated adapter deployment (image = debian-slim + Playwright + Chromium, playwright-go) consuming `system.adapter.browser-runner.requests` (035 Convention A) and replying on the caller's topic with `{results:[{check_id, profile, url, pass, detail}]}`. P0 scope: desktop 1366×900, three check types (`page_status_ok`, `selector_exists` asserted against the LIVE DOM after settle, `no_console_errors`); everything else honestly reported in `skipped[]`, never faked; browser launched per request so a crash poisons one run, not the pod; navigation failure is a check-fail, not an infra error. Contract pinned to the 035 Adapter Guide as normative after a compliance pass (typed header struct with real bools; `in_response_to_request_id` = incoming request_id is THE matcher; ProduceWithValidation never plain Produce). Build gotchas banked: playwright.azureedge.net CDN dead; v0.6100.0 must be required under its declared (pre-rename) module path; the driver ignores XDG_CACHE_HOME — set HOME in the image.
- **sources:** PLAN_tool_acceptance_runner(2).md (whole); RUNBOOK_travelling_docs(38).md#stage-6; RUNNING_NOTES_travelling_docs(39).md#stage-6-built,#2026-07-11; HANDOFF_2026-07-10…md#T3–T6
- **relations:** analyser-adapter mould (pattern source); 035 adapter guide; tool-acceptance-agent (caller).
- **verify-later:** `cmd/browser-runner-adapter/main.go`; `internal/adapters/browserrunner/`; KafkaTopic CR; dockerfile HOME=/pw-home.

<!-- SOURCE: U12_docs024_archives.md -->
### Adapter deployment troubleshooting (ImagePullBackOff / command-vs-args / Kafka topic provisioning)
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** Appears only as "## 10. Adapter & Service Deployment Issues" in `debugging_old/016_debugging_guide_v2(1).md`; absent from every subsequent snapshot and from the live consolidated debugging guide (zero grep hits); the content lives instead in `035_adapter_guide.md`.
- **what:** Covers real thunder-adapter-era deployment failures: diagnosing Docker Hub `ImagePullBackOff`/`insufficient_scope`, the Kubernetes `command:` vs `args:` trap (args silently replaces the entire Dockerfile CMD), the Strimzi `auto.create.topics.enable=false` gotcha requiring an explicit KafkaTopic CRD, and a "deployment essentials checklist" required for every new adapter.
- **sources:** debugging_old/016_debugging_guide_v2(1).md#"10"; docs024_key_docs_latest/035_adapter_guide.md#"2.12-2.13"
- **relations:** adapters (035_adapter_guide.md), deployment-github, single-source relocation convention (below)
- **verify-later:** confirm `035_adapter_guide.md` §2.12/§2.13 still matches the checklist.

<!-- SOURCE: U12_docs024_archives.md -->
### Adapter Response Envelope Contract relocated from 003 to the adapter guide
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** `debugging_old/003_contracts_and_standards_v10/v11.md` contain the full section; live `003_contracts_and_standards(8).md` replaces it with one line: "Moved to 035_adapter_guide.md §1... now the single source for it."
- **what:** Defines how a long-lived adapter must shape its Kafka reply so the chassis recognises it as an awaited response: reuse the incoming `request_id`, fresh `message_id`, `ProduceWithValidation` (never plain `Produce`), and a typed Go struct for response `headers` (not `map[string]string`) so `is_complete`/`is_error` marshal as real JSON booleans. Motivated by a real production incident (thunder-adapter matcher failure, 2026-05-22).
- **sources:** debugging_old/003_contracts_and_standards_v11.md#"Adapter Response Envelope Contract"; docs024_key_docs_latest/035_adapter_guide.md#"1"
- **relations:** adapters, tool-pipeline, single-source relocation convention (below)
- **verify-later:** check adapter Go source for typed `ResponseHeaders` struct vs any remaining `map[string]string` header builders.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### git-adapter as sole write credential holder
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "owner decision, 2026-07-12): the GitHub write credential stays in the git-adapter" (SUMMARY_write_step_position_2026-07-12.md)
- **what:** An architecture decision that the fix-implementer never holds a GitHub write token at all; it sends requests to the git-adapter for `create_branch`, `commit`, and `create_pull_request`, exactly as the site-deploy pipeline already does. Chosen over injecting a write token into the implementer's pod, keeping write credentials entirely out of LLM-driven pods. The implementer pod holds only a read-only token via the isRepoCloningAgent spawn gate.
- **sources:** fixloop_eg_dartsonline/SUMMARY_write_step_position_2026-07-12.md, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 25
- **relations:** write step; git-adapter new actions; isRepoCloningAgent spawn gate; fix-implementer-orchestrator
- **verify-later:** grep/inspect `create_branch`; `commit`; `create_pull_request`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### git-adapter new actions (create_branch, create_pull_request, branch-aware commit)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "Part 1 BUILT (git-adapter, commit 89175383)... 4-test httptest suite green" (NOTES(10)#Turn 25)
- **what:** New git-adapter capabilities: `create_branch` is idempotent (existing branch returns its head rather than erroring); `create_pull_request` defaults its base to the repo's default branch and is the human review terminal; commit gains an optional `branch` parameter with domain-prefixing skipped for platform commits.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#create_branch/commit_files/create_pr, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 25
- **relations:** git-adapter as sole write credential holder; git_adapter_request generic caller
- **verify-later:** grep/inspect `create_branch`; `create_pull_request`; `branch`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Agent-to-adapter capability maturation path (lanes: fast/slow/job)
- **category:** adapters
- **status-signal:** aspirational
- **status-evidence:** "Prove a slow-lane capability as a spawned agent first ... Promote the popular ones to warm adapters ... End-state: a resident chat-orchestrator adapter..." (PLAN_isolated_chat_environment(5).md §11)
- **what:** A general pattern for how agentic capability should mature over time to reduce latency: prove a capability first as a spawned agent, promote popular ones to warm long-running single-replica adapters, and converge on a resident orchestrator adapter that fans out to capability adapters without spawning per request. Framed for chat's three lanes — fast lane (bounded Q&A), slow lane (agentic, latency-warned), job lane (long-running submission + status).
- **sources:** tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#11
- **relations:** Isolated chat/satellite architecture (Y-copy); Building-and-hosting as a service via chat
- **verify-later:** whether any chat capability has been promoted past "spawned agent" to a warm adapter

<!-- SOURCE: U15_docs019_running_notes.md -->
### Adapter response envelope contract
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "003-vs-FOCUS contradiction RESOLVED empirically (2026-06-11) from coordinator.go + git/thunder/dynamic/websearch adapters" (principles(59)).
- **what:** The chassis's normative contract for any Kafka message replier (adapter or agent): body must be a typed struct with real bools (never `map[string]string` string-bools — this exact bug caused a real multi-day thunder production fault), sent via `ProduceWithValidation` (never bare `Produce`), with `in_response_to_request_id` as the primary matcher the coordinator claims on (`request_id` is a fallback — "reuse both" is the safest pattern), and `action`/payload read from the message BODY (not headers). Originally duplicated and drifted between doc 003 and `FOCUS_adapter_design`; single-sourced into `035_adapter_guide.md` after empirical verification against `coordinator.go` and four live adapters (websearch was the deprecated outlier).
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 envelope-contract entries; NOTES_running_synthesis_v2(36).md/v3(32).md (analyser adapter build referencing the same contract).
- **relations:** Analyser adapter build; canonical-doc-home discipline; code-context retrieval infrastructure.
- **verify-later:** `platform/orchestration/types` `ResponseHeaders`/`ResponseMessage`; `platform/validation` `ValidateOutgoingMessage` (the still-open "promote from prose to validator enforcement" TODO).

<!-- SOURCE: U15_docs019_running_notes.md -->
### Analyser adapter build (polyglot code-parsing service)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "MILESTONE 2026-06-12: analyser-adapter DEPLOYED TO PRODUCTION (uk_001)" (principles(59)).
- **what:** A from-scratch Kafka-worker adapter modelled structurally on thunder/git (own image, dockerfile, kustomize base+overlay, config loader, graceful `Shutdown()` with `sync.Once`, health probes) whose one genuine difference is importing the shared chassis-root `internal/analysis` package and holding a dedicated least-privilege, read-only, repo-scoped GitHub token via `secretKeyRef` (never passed through the spawning pod). Fetches via a tarball GET (no git binary, no go-git), not a clone.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11/12 adapter-build entries; NOTES_running_synthesis_v3(32).md (turns 25-27, tarball-fetcher reuse into `internal/reposource`).
- **relations:** Adapter response envelope contract; code-context retrieval infrastructure; GitHub read-token scoping pattern.

<!-- SOURCE: U15_docs019_running_notes.md -->
### GitHub read-token scoping / least-privilege adapter secrets pattern
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** v3(32) DECISIONS: "GitHub read token scoped to the diagnoser via secretKeyRef, not passthrough (turn 25)."
- **what:** `spawn_actions.go` injects `GITHUB_READ_TOKEN` from a shared platform secret only for agent types flagged `isRepoCloningAgent` (currently just `diagnose-agent`), via `secretKeyRef` so the spawning pod itself never holds the token and no other agent type is granted it — the same read-only single-repo PAT the analyser adapter uses.
- **sources:** NOTES_running_synthesis_v3(32).md DECISIONS (turn 25).
- **relations:** Analyser adapter build; diagnose-agent self-contained repo fetch.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Thunder adapter (GPU provisioning, reaper, cost caps, credential boundary)
- **category:** adapters
- **status-signal:** partial
- **status-evidence:** 033 header "Proposal for routing all Thunder Compute interactions through a long-running cluster adapter"; debugging guide v2_44 shows Phases 2–6 progressing
- **what:** A single long-lived `thunder-adapter` Deployment that holds the Thunder API key/B2 creds/SSH keypair store, provisions ephemeral GPU VMs via Kafka actions, and preserves a credential boundary: VMs get only ephemeral SSH keys + hours-expiring presigned URLs. Defence-in-depth: Thunder hard 12h uptime cap + a 15-min reaper + a daily cost cap.
- **sources:** WM/033_thunder_adapter_design.md#tldr, WM/033_thunder_adapter_design.md#preventing-indefinite-running-gpus-defence-in-depth, WM/033_thunder_adapter_design.md#new-schema, WM/016_debugging_guide_v2_44.md#9
- **relations:** fine-tuning flywheel; adapters pattern; multicluster provisioning
- **verify-later:** thunder_instances; thunder_budget_state; model_lifecycle.training_runs; system.adapter.thunder.requests

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Adapter & service deployment debugging (rescued/dropped section)
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** 016 base "Adapter & Service Deployment Issues … Rescued from an earlier guide revision … dropped from the main line"; absent from v2_44
- **what:** A family-delta section present in the base 016 but dropped from the v2_x main line: diagnosing adapter deployment failures (ImagePullBackOff/`insufficient_scope`, immediate crashes from `args:` replacing the whole CMD, `Unknown Topic Or Partition` on first message) and a deployment-essentials checklist. Built from the thunder-adapter Phase 2 debugging.
- **sources:** WM/016_debugging_guide.md#adapter-service-deployment-issues, WM/016_debugging_guide.md#imagepullbackoff-insufficient_scope-authorization-failed
- **relations:** dropped from 016 v2_44 (superseded main line); Thunder adapter; deployment-github
- **verify-later:** kustomize base/deployment.yaml; docker-hub-creds secret; ai-persona-app service account

<!-- SOURCE: U17b_docs019_gofiles.md -->
### analyser-adapter deployment/migration plan
- **category:** adapters
- **status-signal:** partial
- **status-evidence:** README.md marks most destinations "(NEW)"/"(EDIT)" against the real repo tree (tree -d, 2026-06-11), but notes "`NNN_create_code_symbols_index.sql` → workspace root (ALREADY APPLIED — commit for the record, your numbering)"
- **what:** A directory-by-directory migration map from a `chassis-drafts/analyser-adapter` staging area (which does not compile in this tree) to real agentchassis destinations: `cmd/analyser-adapter/main.go`, `internal/adapters/analyser/{adapter,analyse_action,github_source}.go`, `platform/orchestration/actions/{code_symbols_actions,analyser_request_action}.go` (+ registry.go insertion), the code-indexer migration, `configs/analyser-adapter.yaml`, a two-stage Dockerfile, and kustomize base/overlay scaffolding — all following the conventions already used for thunder-adapter. Also flags un-placeable items needing a human call: the `035_adapter_guide.md` doc home, the `system.adapter.analyser.requests` KafkaTopic CRD location, and the `analyser-github-read` Secret (never committed with a real token).
- **sources:** contextkit/README.md, contextkit/README(2).md
- **relations:** code-indexer agent, adapters (033/035 thunder/webscrape pattern), deployment-github (034)
- **verify-later:** build/docker/backend/, deployments/kustomize/services/analyser-adapter/, Makefile — confirm which of the four described insertions actually exist

<!-- SOURCE: U19_sql_tables_components.md -->
### Thunder adapter schema and provisioning gates
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** Full schema with recorded user decisions ($100/day cap, 2 concurrent, 18h uptime, $1.80/hr A100, $25 estimated run); production fix dated 2026-05-22 for identifier recycling; ssh_port verification dated 2026-05-24.
- **what:** GPU VM lifecycle for training: thunder_instances (one row per VM ever provisioned — inserted BEFORE the API call so the reaper always has a record; status machine provisioning→running→decommissioning→decommissioned with reaped/lost/failed terminals; cost snapshot; reaper bookkeeping; FK to model_lifecycle.training_runs), thunder_config singleton (CHECK-enforced single row; caps and pause switch), and computed views thunder_spend_24h (rolling cost incl. running estimates, no drifting counter) and thunder_provision_check (can_provision + denial_reason evaluated at every provision request). Identifier recycling fixed by replacing global uniqueness with a partial unique index over live states only; ssh_port captured at provision so ssh_exec dials directly.
- **sources:** docs/agent_docs/sql_for_tables/042_thunder.sql
- **relations:** thunder unreachable streak; training_runs (flywheel C); 013_thunder_adapter_design.md.
- **verify-later:** adapter reading thunder_provision_check; reaper behaviour.

<!-- SOURCE: U19_sql_tables_components.md -->
### Thunder consecutive-unreachable probe streak
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** Migration 106 with in-transaction verification; rationale documented (single down-probe could be a transient SSH blip; each scheduler tick is a fresh sub-agent that can't hold state in memory).
- **what:** thunder-training-monitor durability: consecutive_unreachable_probes counter (+ last_probe_at) on thunder_instances, bumped/reset by the record_probe_streak action; only after the streak crosses a threshold is the instance treated as 'lost' (fail run + decommission). State lives on the row because monitor ticks are stateless.
- **sources:** docs/agent_docs/sql_for_tables/047_thunder_unreachable_counter.sql
- **relations:** thunder adapter; scheduler tick statelessness.
- **verify-later:** record_probe_streak action; threshold value in monitor config.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Adapter microservice pattern (Kafka/HTTP adapters + secure external-API proxies)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** 0123 codifies the pattern; image-generator, firecrawl/webscrape, playwright, git adapters all follow it; "We will use this exact same pattern for all our Python-based actions."
- **what:** Go agents never embed heavy dependencies: a workflow action produces a Kafka message to `system.adapter.<name>.requests` (or an internal HTTP call); a containerised worker service (Python or Go) in its own Deployment consumes via a shared consumer group, does the work (Playwright, Firecrawl, Stability, git), and replies to the reply_to topic. External GPU/API providers get a dedicated Go proxy adapter that holds the secret key and translates request formats — swap providers by changing one adapter, no workflow changes.
- **sources:** docs003_firecrawl/README.0123.actions_needed_firstdraftpython.md; docs004_website_capture_project/playwright/implementation_roadmap.md; docs001_flow_general/README.097a.imagecreationandstorageflow.md
- **relations:** adapters category anchor (docs 033/035 successor); image adapter; firecrawl adapter; thunder adapter (taxonomy) is a descendant of the "ThunderCompute LLaVA proxy" idea here.
- **verify-later:** internal/adapters/* inventory; which adapter topics exist.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Firecrawl scraping adapter and actions
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** Agent definitions website-capture-firecrawl and webscrape-simple with image tags v1.0.407→v1.0.424 (iteration = real use); v1→v2 migration doc fixing "Unrecognized keys" errors and adding S3 ownership of screenshots/images.
- **what:** Firecrawl API adapter (Kafka consumer on system.adapter.firecrawl.requests) exposing scrape/crawl/extract actions to workflows (firecrawl_scrape, firecrawl_crawl, firecrawl_extract, plus a registered scrape_web action with upload_results to S3). v2 migration: formats array incl. screenshot+links, downloading Google-Cloud-hosted screenshots/images into own S3 (webscrape/client/date/id/ layout) for data ownership since Firecrawl assets expire in 30 days. Chosen over the half-built Playwright adapter to reduce MVP load.
- **sources:** docs003_firecrawl/README.0126.firecrawl_agent_definition.md; docs004_website_capture_project/firecrawl/001claude_initial.md; docs004_website_capture_project/firecrawl/002firecrawl_visual_flow.md; docs003_firecrawl/README.0129.testing_webscrape_message.md
- **relations:** adoption-pipeline crawling (live successor); playwright adapter (the road not taken); storage-architecture.
- **verify-later:** web-scrape-adapter deployment (referenced in initial_messages.txt scale-down list — so it was deployed); FIRECRAWL_API_KEY secret.

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Adapter design pattern (canonical guide)
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** context-pack FOCUS_adapter_design(3).md is a frozen copy; live docs advance to flywheel_docs/FOCUS_adapter_design(3).md and docs024_key_docs_latest/035_adapter_guide.md (38KB, Jun 2026)
- **what:** The canonical guide for building single-replica Kafka-consuming "adapter" services that wrap one external API and hold its credentials (git, web-scrape, image-generator, ollama, thunder). Covers the Adapter struct, NewAdapter cleanup convention, sequential Run loop, handleMessage dispatch, the three-tier response-header contract (Tier-1 validator-required, Tier-2 chassis-routing e.g. `in_response_to_request_id`, Tier-3 observability), health/shutdown, topic naming conventions A vs B, and deployment essentials (serviceAccountName, imagePullSecrets, `command:` not `args:`, Strimzi topic pre-creation).
- **sources:** docubundle/.../FOCUS_adapter_design(3).md#TL;DR, #Responsibilities, #Sending-responses, #Deployment-essentials
- **relations:** superseded by live 035_adapter_guide.md; instantiated by thunder-adapter; response-header contract underpins the reply-topic bugs
- **verify-later:** internal/adapters/*/adapter.go; platform/validation/Validator; 035_adapter_guide.md

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### thunder-adapter — Thunder Compute GPU provisioning adapter
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** STATUS_thunder_adapter_2026-06_04 §1 phases 3.0–3.6; FOCUS(21) §14 "Provision loop verified end-to-end (2026-05-22)"
- **what:** The adapter that wraps the Thunder Compute API to provision/decommission on-demand GPU VMs, holding the Thunder token and B2 keys. Actions: `provision_instance` (spend-check → ed25519 keypair → create → WaitForRunning → INSERT `thunder_instances` with compensating cleanup), `decommission_instance` (idempotent, computes cost from running_since), plus SSH (`ssh_exec`, `ssh_get_status`) and presign actions. Two matcher bugs that blocked it for days: response headers must be a typed struct (not map[string]string) so is_complete/is_error serialise as JSON bools; and `thunder_instance_id` uniqueness must be a partial index on live rows because Thunder recycles numeric ids.
- **sources:** docubundle/.../STATUS_thunder_adapter_2026-06_04.md#1, #3; flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Thunder Compute API specifics)
- **relations:** called by gpu-provisioner, training-launcher, thunder-reaper, thunder-training-monitor; credential boundary for presigned URLs
- **verify-later:** internal/adapters/thunder/api/types.go, provision_action.go, decommission_action.go; thunder_instances, thunder_config, thunder_provision_check

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Thunder Compute API specifics (field/casing/template traps)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §14 "Request/response field shape is asymmetric (verified 2026-05-20 via tnr status --json)"
- **what:** Hard-won Thunder API facts: base URL `https://api.thundercompute.com:8443/v1`; CREATE uses snake_case ints (gpu_type, cpu_cores, num_gpus) but STATUS/LIST returns camelCase with numbers as JSON strings and UPPERCASE enums; real templates are `base`/`ollama`/`comfy-ui`/`forge-neo`/`unsloth` (the OpenAPI `ubuntu-22.04` example is rejected); the login user is `ubuntu` not `root`; SSH needs wait-for-sshd (RUNNING ≠ sshd ready); the SSH port from list is unreliable, use `tnr connect --json`.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Thunder Compute API specifics; Phase 4 SSH item)
- **relations:** underpins thunder-adapter; `IdentifierInt()`/`IsReadyStatus` handling
- **verify-later:** internal/adapters/thunder/api/types.go (CreateInstanceRequest vs Instance)

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### P5 vmhost provisioning adapter (superseded service-deployer)
- **category:** adapters
- **status-signal:** aspirational
- **status-evidence:** plan(11) P5 "A vmhost adapter for what DOES need SSH … built to the analyser-adapter README skeleton"; earlier plan(1)/(4)/(5) P5 "registry + relocation (service_instances) and, eventually, the chassis service-deployer adapter".
- **what:** The eventual automation for what genuinely needs SSH: provision box, run setup.sh, onboard domain, ship engine, decommission — built as a `vmhost` adapter (cmd/vmhost-adapter, internal/adapters/vmhost, reuse thunder's ssh via shared/, kustomize, KafkaTopic system.adapter.vmhost.requests, 003 envelope), with a `service_instances` registry modelled on thunder_instances minus the reaper. Long-term it holds the deploy SSH credential, retiring the repo-secrets copy. Earlier versions named this the "service-deployer" adapter.
- **sources:** traffic_probe_plan(11).md#p5, traffic_probe_plan(1).md#phases, traffic_probe_running_notes(27).md#open-threads
- **relations:** future handler for backend_unreachable; supersedes "service-deployer" naming
- **verify-later:** analyser-adapter README; thunder_instances → service_instances

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### adapter response-envelope contract (request/response wiring conventions)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** adapter(4).go's header states the envelope is "grounded in the orchestrator (coordinator.go) and the three working adapters, not the docs — which disagree (003 vs FOCUS) and were resolved empirically." internal/adapters/analyser/ exists live in the repo, confirming the draft shipped.
- **what:** A reverse-engineered (from working adapters, not from docs) contract for how an adapter must shape its Kafka request/response envelope so the orchestrator actually routes the reply instead of timing out: action comes from `body.action` not headers; `in_response_to_request_id` (echoing the incoming `request_id`) is the load-bearing claim field in `coordinator.go`'s `ProcessResponse`, with `request_id` as fallback; the reply body must use canonical `types.ResponseHeaders` via `ToResponseHeaders` so `is_complete`/`is_error` marshal as real JSON bools (the "bool trap" — a `map[string]string` sending the string `"true"` fails the receiver's struct-bool unmarshal); sends must go via `ProduceWithValidation`, never plain `Produce`. websearch-adapter is flagged as the one adapter still on the deprecated string-bool/plain-Produce path.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/adapter(4).go, docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/analyser_request_action(1).go
- **relations:** contextkit toolchain (above); analyser-adapter service
- **verify-later:** platform/orchestration/coordinator.go ProcessResponse, internal/adapters/analyser/adapter.go, whether websearch-adapter has since been migrated off the string-bool map

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Adapter/response message envelope contract (normative)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** 035 §1 "last verified against code 2026-06-11"; 003(8) now points here
- **what:** Any reply to a chassis request must be recognised as an awaited response or it silently falls to process-as-work (row stuck waiting, ~10-min retries, no error). Load-bearing field: in_response_to_request_id = incoming request_id (request_id fallback only; git adapter's reuse pattern favoured). Three header tiers (validator-enforced / coordinator-needed-but-unvalidated / observability). Body headers MUST be a typed struct with real bools (map[string]string string-bools fail unmarshal and drop the reply pre-claim — the multi-day thunder outage). Send via ProduceWithValidation. Request parsing: action from body, payload at body.data, accept reply-topic from three keys. Sibling race: local dispatches must preRegisterAwaitedRequest before send (confirmed fixed in prod 06-09).
- **sources:** 035 §1; 016 §9 bool-trap + race entries
- **relations:** awaited_requests; O(K²) batch presign
- **verify-later:** ValidateOutgoingMessage field list

<!-- SOURCE: U06_finetuning.md -->
### Adapter design guide (adapter vs agent vs inline; canonical structure)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** FOCUS_adapter_design(3) opening: "Canonical guide for building long-running cluster services… Examples drawn from the working adapters in the repository."
- **what:** The decision rule (one external API + multiple internal callers → adapter; long per-orchestration work → spawned agent; short single-agent call → inline; shared infra like DB/Kafka → nothing) plus the canonical shape: struct fields, ordered NewAdapter with manual cleanup on every failure path, sequential fetch-handle-loop (no goroutine-per-message by default), handleMessage parse/dispatch/respond, health endpoints, sync.Once shutdown, topic conventions (`system.adapter.<name>.requests` for new work), config YAML field-name traps, credentials from env only.
- **sources:** working/flywheel_docs/FOCUS_adapter_design(3).md (whole)
- **relations:** thunder adapter (the guide's newest example); response header tiers; deployment essentials
- **verify-later:** consistency of existing adapters with the guide

<!-- SOURCE: U06_finetuning.md -->
### Adapter response-header tier taxonomy and the validator-coverage gap
- **category:** adapters
- **status-signal:** partial
- **status-evidence:** FOCUS_adapter_design(3) "TODO — tighten validator coverage… Tier-2 fields are necessary for the orchestration to advance but not validated… Tracking issue: not yet filed."
- **what:** Response headers split into Tier 1 (five fields the platform Validator enforces; `is_error=true` bypasses), Tier 2 (what the chassis needs to route the reply to the awaiting orchestration — `in_response_to_request_id`, message_type, status vocabulary complete/error_recoverable/error_unrecoverable, is_complete/is_error, etc.; missing these means a silent AWAITING_RESPONSES hang the validator won't catch), Tier 3 (observability). Known live consequence of the gap: the matcher fix of 2026-05-22 (typed response-header struct so booleans serialise as real bools — a map[string]string sent string bools and the chassis dropped the reply). Proposal to extend the validator exists but is unfiled.
- **sources:** working/flywheel_docs/FOCUS_adapter_design(3).md#sending-responses,#todo; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics (matcher fix)
- **relations:** reply-topic derivation; send-before-register race (same stuck-await symptom family)
- **verify-later:** platform/validation/Validator current coverage

<!-- SOURCE: U06_finetuning.md -->
### Adapter deployment essentials (manifest, cluster resources, RBAC, Makefile)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** FOCUS_adapter_design(3) "Deployment essentials — Real lessons from deploying thunder-adapter Phase 2. Every item below is something the deployment failed without."
- **what:** The complete pre-flight for shipping an adapter: serviceAccountName + imagePullSecrets + `command:` (not `args:` — Dockerfiles use CMD, so args replaces the binary path), required Secrets/SA/Docker-Hub grants, explicit KafkaTopic CRDs (Strimzi auto-create is off; missing reply topics fail only at first response), Recreate strategy, single replica, RBAC trap (resourceNames supports no globs — scope by verbs instead for dynamic names like thunder-ssh-<uuid>), four Makefile insertion points and the newName/newTag overlay split, pre/post-deploy checklists.
- **sources:** working/flywheel_docs/FOCUS_adapter_design(3).md#deployment-essentials; working/flywheel_docs/FOCUS_finetuning_flywheel_changelog_addition.md (phase-2 deploy saga)
- **relations:** adapter design guide; wrong-binary image incident; debugging guide §10
- **verify-later:** thunder-adapter kustomize base vs the checklist

<!-- SOURCE: U08_travelling_docs.md -->
### Tier-4 browser-runner adapter (headless Chromium over Kafka)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "Stage 6 P0 DEPLOYED + SMOKE PASSED" 2026-07-11 (v1.0.1107; §2.15 smoke matched manual inspection; real bool headers; in_response_to_request_id matcher).
- **what:** A dedicated adapter deployment (image = debian-slim + Playwright + Chromium, playwright-go) consuming `system.adapter.browser-runner.requests` (035 Convention A) and replying on the caller's topic with `{results:[{check_id, profile, url, pass, detail}]}`. P0 scope: desktop 1366×900, three check types (`page_status_ok`, `selector_exists` asserted against the LIVE DOM after settle, `no_console_errors`); everything else honestly reported in `skipped[]`, never faked; browser launched per request so a crash poisons one run, not the pod; navigation failure is a check-fail, not an infra error. Contract pinned to the 035 Adapter Guide as normative after a compliance pass (typed header struct with real bools; `in_response_to_request_id` = incoming request_id is THE matcher; ProduceWithValidation never plain Produce). Build gotchas banked: playwright.azureedge.net CDN dead; v0.6100.0 must be required under its declared (pre-rename) module path; the driver ignores XDG_CACHE_HOME — set HOME in the image.
- **sources:** PLAN_tool_acceptance_runner(2).md (whole); RUNBOOK_travelling_docs(38).md#stage-6; RUNNING_NOTES_travelling_docs(39).md#stage-6-built,#2026-07-11; HANDOFF_2026-07-10…md#T3–T6
- **relations:** analyser-adapter mould (pattern source); 035 adapter guide; tool-acceptance-agent (caller).
- **verify-later:** `cmd/browser-runner-adapter/main.go`; `internal/adapters/browserrunner/`; KafkaTopic CR; dockerfile HOME=/pw-home.

<!-- SOURCE: U12_docs024_archives.md -->
### Adapter deployment troubleshooting (ImagePullBackOff / command-vs-args / Kafka topic provisioning)
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** Appears only as "## 10. Adapter & Service Deployment Issues" in `debugging_old/016_debugging_guide_v2(1).md`; absent from every subsequent snapshot and from the live consolidated debugging guide (zero grep hits); the content lives instead in `035_adapter_guide.md`.
- **what:** Covers real thunder-adapter-era deployment failures: diagnosing Docker Hub `ImagePullBackOff`/`insufficient_scope`, the Kubernetes `command:` vs `args:` trap (args silently replaces the entire Dockerfile CMD), the Strimzi `auto.create.topics.enable=false` gotcha requiring an explicit KafkaTopic CRD, and a "deployment essentials checklist" required for every new adapter.
- **sources:** debugging_old/016_debugging_guide_v2(1).md#"10"; docs024_key_docs_latest/035_adapter_guide.md#"2.12-2.13"
- **relations:** adapters (035_adapter_guide.md), deployment-github, single-source relocation convention (below)
- **verify-later:** confirm `035_adapter_guide.md` §2.12/§2.13 still matches the checklist.

<!-- SOURCE: U12_docs024_archives.md -->
### Adapter Response Envelope Contract relocated from 003 to the adapter guide
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** `debugging_old/003_contracts_and_standards_v10/v11.md` contain the full section; live `003_contracts_and_standards(8).md` replaces it with one line: "Moved to 035_adapter_guide.md §1... now the single source for it."
- **what:** Defines how a long-lived adapter must shape its Kafka reply so the chassis recognises it as an awaited response: reuse the incoming `request_id`, fresh `message_id`, `ProduceWithValidation` (never plain `Produce`), and a typed Go struct for response `headers` (not `map[string]string`) so `is_complete`/`is_error` marshal as real JSON booleans. Motivated by a real production incident (thunder-adapter matcher failure, 2026-05-22).
- **sources:** debugging_old/003_contracts_and_standards_v11.md#"Adapter Response Envelope Contract"; docs024_key_docs_latest/035_adapter_guide.md#"1"
- **relations:** adapters, tool-pipeline, single-source relocation convention (below)
- **verify-later:** check adapter Go source for typed `ResponseHeaders` struct vs any remaining `map[string]string` header builders.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### git-adapter as sole write credential holder
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "owner decision, 2026-07-12): the GitHub write credential stays in the git-adapter" (SUMMARY_write_step_position_2026-07-12.md)
- **what:** An architecture decision that the fix-implementer never holds a GitHub write token at all; it sends requests to the git-adapter for `create_branch`, `commit`, and `create_pull_request`, exactly as the site-deploy pipeline already does. Chosen over injecting a write token into the implementer's pod, keeping write credentials entirely out of LLM-driven pods. The implementer pod holds only a read-only token via the isRepoCloningAgent spawn gate.
- **sources:** fixloop_eg_dartsonline/SUMMARY_write_step_position_2026-07-12.md, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 25
- **relations:** write step; git-adapter new actions; isRepoCloningAgent spawn gate; fix-implementer-orchestrator
- **verify-later:** grep/inspect `create_branch`; `commit`; `create_pull_request`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### git-adapter new actions (create_branch, create_pull_request, branch-aware commit)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "Part 1 BUILT (git-adapter, commit 89175383)... 4-test httptest suite green" (NOTES(10)#Turn 25)
- **what:** New git-adapter capabilities: `create_branch` is idempotent (existing branch returns its head rather than erroring); `create_pull_request` defaults its base to the repo's default branch and is the human review terminal; commit gains an optional `branch` parameter with domain-prefixing skipped for platform commits.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#create_branch/commit_files/create_pr, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 25
- **relations:** git-adapter as sole write credential holder; git_adapter_request generic caller
- **verify-later:** grep/inspect `create_branch`; `create_pull_request`; `branch`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Agent-to-adapter capability maturation path (lanes: fast/slow/job)
- **category:** adapters
- **status-signal:** aspirational
- **status-evidence:** "Prove a slow-lane capability as a spawned agent first ... Promote the popular ones to warm adapters ... End-state: a resident chat-orchestrator adapter..." (PLAN_isolated_chat_environment(5).md §11)
- **what:** A general pattern for how agentic capability should mature over time to reduce latency: prove a capability first as a spawned agent, promote popular ones to warm long-running single-replica adapters, and converge on a resident orchestrator adapter that fans out to capability adapters without spawning per request. Framed for chat's three lanes — fast lane (bounded Q&A), slow lane (agentic, latency-warned), job lane (long-running submission + status).
- **sources:** tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#11
- **relations:** Isolated chat/satellite architecture (Y-copy); Building-and-hosting as a service via chat
- **verify-later:** whether any chat capability has been promoted past "spawned agent" to a warm adapter

<!-- SOURCE: U15_docs019_running_notes.md -->
### Adapter response envelope contract
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "003-vs-FOCUS contradiction RESOLVED empirically (2026-06-11) from coordinator.go + git/thunder/dynamic/websearch adapters" (principles(59)).
- **what:** The chassis's normative contract for any Kafka message replier (adapter or agent): body must be a typed struct with real bools (never `map[string]string` string-bools — this exact bug caused a real multi-day thunder production fault), sent via `ProduceWithValidation` (never bare `Produce`), with `in_response_to_request_id` as the primary matcher the coordinator claims on (`request_id` is a fallback — "reuse both" is the safest pattern), and `action`/payload read from the message BODY (not headers). Originally duplicated and drifted between doc 003 and `FOCUS_adapter_design`; single-sourced into `035_adapter_guide.md` after empirical verification against `coordinator.go` and four live adapters (websearch was the deprecated outlier).
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 envelope-contract entries; NOTES_running_synthesis_v2(36).md/v3(32).md (analyser adapter build referencing the same contract).
- **relations:** Analyser adapter build; canonical-doc-home discipline; code-context retrieval infrastructure.
- **verify-later:** `platform/orchestration/types` `ResponseHeaders`/`ResponseMessage`; `platform/validation` `ValidateOutgoingMessage` (the still-open "promote from prose to validator enforcement" TODO).

<!-- SOURCE: U15_docs019_running_notes.md -->
### Analyser adapter build (polyglot code-parsing service)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "MILESTONE 2026-06-12: analyser-adapter DEPLOYED TO PRODUCTION (uk_001)" (principles(59)).
- **what:** A from-scratch Kafka-worker adapter modelled structurally on thunder/git (own image, dockerfile, kustomize base+overlay, config loader, graceful `Shutdown()` with `sync.Once`, health probes) whose one genuine difference is importing the shared chassis-root `internal/analysis` package and holding a dedicated least-privilege, read-only, repo-scoped GitHub token via `secretKeyRef` (never passed through the spawning pod). Fetches via a tarball GET (no git binary, no go-git), not a clone.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11/12 adapter-build entries; NOTES_running_synthesis_v3(32).md (turns 25-27, tarball-fetcher reuse into `internal/reposource`).
- **relations:** Adapter response envelope contract; code-context retrieval infrastructure; GitHub read-token scoping pattern.

<!-- SOURCE: U15_docs019_running_notes.md -->
### GitHub read-token scoping / least-privilege adapter secrets pattern
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** v3(32) DECISIONS: "GitHub read token scoped to the diagnoser via secretKeyRef, not passthrough (turn 25)."
- **what:** `spawn_actions.go` injects `GITHUB_READ_TOKEN` from a shared platform secret only for agent types flagged `isRepoCloningAgent` (currently just `diagnose-agent`), via `secretKeyRef` so the spawning pod itself never holds the token and no other agent type is granted it — the same read-only single-repo PAT the analyser adapter uses.
- **sources:** NOTES_running_synthesis_v3(32).md DECISIONS (turn 25).
- **relations:** Analyser adapter build; diagnose-agent self-contained repo fetch.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Thunder adapter (GPU provisioning, reaper, cost caps, credential boundary)
- **category:** adapters
- **status-signal:** partial
- **status-evidence:** 033 header "Proposal for routing all Thunder Compute interactions through a long-running cluster adapter"; debugging guide v2_44 shows Phases 2–6 progressing
- **what:** A single long-lived `thunder-adapter` Deployment that holds the Thunder API key/B2 creds/SSH keypair store, provisions ephemeral GPU VMs via Kafka actions, and preserves a credential boundary: VMs get only ephemeral SSH keys + hours-expiring presigned URLs. Defence-in-depth: Thunder hard 12h uptime cap + a 15-min reaper + a daily cost cap.
- **sources:** WM/033_thunder_adapter_design.md#tldr, WM/033_thunder_adapter_design.md#preventing-indefinite-running-gpus-defence-in-depth, WM/033_thunder_adapter_design.md#new-schema, WM/016_debugging_guide_v2_44.md#9
- **relations:** fine-tuning flywheel; adapters pattern; multicluster provisioning
- **verify-later:** thunder_instances; thunder_budget_state; model_lifecycle.training_runs; system.adapter.thunder.requests

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Adapter & service deployment debugging (rescued/dropped section)
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** 016 base "Adapter & Service Deployment Issues … Rescued from an earlier guide revision … dropped from the main line"; absent from v2_44
- **what:** A family-delta section present in the base 016 but dropped from the v2_x main line: diagnosing adapter deployment failures (ImagePullBackOff/`insufficient_scope`, immediate crashes from `args:` replacing the whole CMD, `Unknown Topic Or Partition` on first message) and a deployment-essentials checklist. Built from the thunder-adapter Phase 2 debugging.
- **sources:** WM/016_debugging_guide.md#adapter-service-deployment-issues, WM/016_debugging_guide.md#imagepullbackoff-insufficient_scope-authorization-failed
- **relations:** dropped from 016 v2_44 (superseded main line); Thunder adapter; deployment-github
- **verify-later:** kustomize base/deployment.yaml; docker-hub-creds secret; ai-persona-app service account

<!-- SOURCE: U17b_docs019_gofiles.md -->
### analyser-adapter deployment/migration plan
- **category:** adapters
- **status-signal:** partial
- **status-evidence:** README.md marks most destinations "(NEW)"/"(EDIT)" against the real repo tree (tree -d, 2026-06-11), but notes "`NNN_create_code_symbols_index.sql` → workspace root (ALREADY APPLIED — commit for the record, your numbering)"
- **what:** A directory-by-directory migration map from a `chassis-drafts/analyser-adapter` staging area (which does not compile in this tree) to real agentchassis destinations: `cmd/analyser-adapter/main.go`, `internal/adapters/analyser/{adapter,analyse_action,github_source}.go`, `platform/orchestration/actions/{code_symbols_actions,analyser_request_action}.go` (+ registry.go insertion), the code-indexer migration, `configs/analyser-adapter.yaml`, a two-stage Dockerfile, and kustomize base/overlay scaffolding — all following the conventions already used for thunder-adapter. Also flags un-placeable items needing a human call: the `035_adapter_guide.md` doc home, the `system.adapter.analyser.requests` KafkaTopic CRD location, and the `analyser-github-read` Secret (never committed with a real token).
- **sources:** contextkit/README.md, contextkit/README(2).md
- **relations:** code-indexer agent, adapters (033/035 thunder/webscrape pattern), deployment-github (034)
- **verify-later:** build/docker/backend/, deployments/kustomize/services/analyser-adapter/, Makefile — confirm which of the four described insertions actually exist

<!-- SOURCE: U19_sql_tables_components.md -->
### Thunder adapter schema and provisioning gates
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** Full schema with recorded user decisions ($100/day cap, 2 concurrent, 18h uptime, $1.80/hr A100, $25 estimated run); production fix dated 2026-05-22 for identifier recycling; ssh_port verification dated 2026-05-24.
- **what:** GPU VM lifecycle for training: thunder_instances (one row per VM ever provisioned — inserted BEFORE the API call so the reaper always has a record; status machine provisioning→running→decommissioning→decommissioned with reaped/lost/failed terminals; cost snapshot; reaper bookkeeping; FK to model_lifecycle.training_runs), thunder_config singleton (CHECK-enforced single row; caps and pause switch), and computed views thunder_spend_24h (rolling cost incl. running estimates, no drifting counter) and thunder_provision_check (can_provision + denial_reason evaluated at every provision request). Identifier recycling fixed by replacing global uniqueness with a partial unique index over live states only; ssh_port captured at provision so ssh_exec dials directly.
- **sources:** docs/agent_docs/sql_for_tables/042_thunder.sql
- **relations:** thunder unreachable streak; training_runs (flywheel C); 013_thunder_adapter_design.md.
- **verify-later:** adapter reading thunder_provision_check; reaper behaviour.

<!-- SOURCE: U19_sql_tables_components.md -->
### Thunder consecutive-unreachable probe streak
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** Migration 106 with in-transaction verification; rationale documented (single down-probe could be a transient SSH blip; each scheduler tick is a fresh sub-agent that can't hold state in memory).
- **what:** thunder-training-monitor durability: consecutive_unreachable_probes counter (+ last_probe_at) on thunder_instances, bumped/reset by the record_probe_streak action; only after the streak crosses a threshold is the instance treated as 'lost' (fail run + decommission). State lives on the row because monitor ticks are stateless.
- **sources:** docs/agent_docs/sql_for_tables/047_thunder_unreachable_counter.sql
- **relations:** thunder adapter; scheduler tick statelessness.
- **verify-later:** record_probe_streak action; threshold value in monitor config.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Adapter microservice pattern (Kafka/HTTP adapters + secure external-API proxies)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** 0123 codifies the pattern; image-generator, firecrawl/webscrape, playwright, git adapters all follow it; "We will use this exact same pattern for all our Python-based actions."
- **what:** Go agents never embed heavy dependencies: a workflow action produces a Kafka message to `system.adapter.<name>.requests` (or an internal HTTP call); a containerised worker service (Python or Go) in its own Deployment consumes via a shared consumer group, does the work (Playwright, Firecrawl, Stability, git), and replies to the reply_to topic. External GPU/API providers get a dedicated Go proxy adapter that holds the secret key and translates request formats — swap providers by changing one adapter, no workflow changes.
- **sources:** docs003_firecrawl/README.0123.actions_needed_firstdraftpython.md; docs004_website_capture_project/playwright/implementation_roadmap.md; docs001_flow_general/README.097a.imagecreationandstorageflow.md
- **relations:** adapters category anchor (docs 033/035 successor); image adapter; firecrawl adapter; thunder adapter (taxonomy) is a descendant of the "ThunderCompute LLaVA proxy" idea here.
- **verify-later:** internal/adapters/* inventory; which adapter topics exist.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Firecrawl scraping adapter and actions
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** Agent definitions website-capture-firecrawl and webscrape-simple with image tags v1.0.407→v1.0.424 (iteration = real use); v1→v2 migration doc fixing "Unrecognized keys" errors and adding S3 ownership of screenshots/images.
- **what:** Firecrawl API adapter (Kafka consumer on system.adapter.firecrawl.requests) exposing scrape/crawl/extract actions to workflows (firecrawl_scrape, firecrawl_crawl, firecrawl_extract, plus a registered scrape_web action with upload_results to S3). v2 migration: formats array incl. screenshot+links, downloading Google-Cloud-hosted screenshots/images into own S3 (webscrape/client/date/id/ layout) for data ownership since Firecrawl assets expire in 30 days. Chosen over the half-built Playwright adapter to reduce MVP load.
- **sources:** docs003_firecrawl/README.0126.firecrawl_agent_definition.md; docs004_website_capture_project/firecrawl/001claude_initial.md; docs004_website_capture_project/firecrawl/002firecrawl_visual_flow.md; docs003_firecrawl/README.0129.testing_webscrape_message.md
- **relations:** adoption-pipeline crawling (live successor); playwright adapter (the road not taken); storage-architecture.
- **verify-later:** web-scrape-adapter deployment (referenced in initial_messages.txt scale-down list — so it was deployed); FIRECRAWL_API_KEY secret.

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Adapter design pattern (canonical guide)
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** context-pack FOCUS_adapter_design(3).md is a frozen copy; live docs advance to flywheel_docs/FOCUS_adapter_design(3).md and docs024_key_docs_latest/035_adapter_guide.md (38KB, Jun 2026)
- **what:** The canonical guide for building single-replica Kafka-consuming "adapter" services that wrap one external API and hold its credentials (git, web-scrape, image-generator, ollama, thunder). Covers the Adapter struct, NewAdapter cleanup convention, sequential Run loop, handleMessage dispatch, the three-tier response-header contract (Tier-1 validator-required, Tier-2 chassis-routing e.g. `in_response_to_request_id`, Tier-3 observability), health/shutdown, topic naming conventions A vs B, and deployment essentials (serviceAccountName, imagePullSecrets, `command:` not `args:`, Strimzi topic pre-creation).
- **sources:** docubundle/.../FOCUS_adapter_design(3).md#TL;DR, #Responsibilities, #Sending-responses, #Deployment-essentials
- **relations:** superseded by live 035_adapter_guide.md; instantiated by thunder-adapter; response-header contract underpins the reply-topic bugs
- **verify-later:** internal/adapters/*/adapter.go; platform/validation/Validator; 035_adapter_guide.md

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### thunder-adapter — Thunder Compute GPU provisioning adapter
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** STATUS_thunder_adapter_2026-06_04 §1 phases 3.0–3.6; FOCUS(21) §14 "Provision loop verified end-to-end (2026-05-22)"
- **what:** The adapter that wraps the Thunder Compute API to provision/decommission on-demand GPU VMs, holding the Thunder token and B2 keys. Actions: `provision_instance` (spend-check → ed25519 keypair → create → WaitForRunning → INSERT `thunder_instances` with compensating cleanup), `decommission_instance` (idempotent, computes cost from running_since), plus SSH (`ssh_exec`, `ssh_get_status`) and presign actions. Two matcher bugs that blocked it for days: response headers must be a typed struct (not map[string]string) so is_complete/is_error serialise as JSON bools; and `thunder_instance_id` uniqueness must be a partial index on live rows because Thunder recycles numeric ids.
- **sources:** docubundle/.../STATUS_thunder_adapter_2026-06_04.md#1, #3; flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Thunder Compute API specifics)
- **relations:** called by gpu-provisioner, training-launcher, thunder-reaper, thunder-training-monitor; credential boundary for presigned URLs
- **verify-later:** internal/adapters/thunder/api/types.go, provision_action.go, decommission_action.go; thunder_instances, thunder_config, thunder_provision_check

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Thunder Compute API specifics (field/casing/template traps)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §14 "Request/response field shape is asymmetric (verified 2026-05-20 via tnr status --json)"
- **what:** Hard-won Thunder API facts: base URL `https://api.thundercompute.com:8443/v1`; CREATE uses snake_case ints (gpu_type, cpu_cores, num_gpus) but STATUS/LIST returns camelCase with numbers as JSON strings and UPPERCASE enums; real templates are `base`/`ollama`/`comfy-ui`/`forge-neo`/`unsloth` (the OpenAPI `ubuntu-22.04` example is rejected); the login user is `ubuntu` not `root`; SSH needs wait-for-sshd (RUNNING ≠ sshd ready); the SSH port from list is unreliable, use `tnr connect --json`.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Thunder Compute API specifics; Phase 4 SSH item)
- **relations:** underpins thunder-adapter; `IdentifierInt()`/`IsReadyStatus` handling
- **verify-later:** internal/adapters/thunder/api/types.go (CreateInstanceRequest vs Instance)

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### P5 vmhost provisioning adapter (superseded service-deployer)
- **category:** adapters
- **status-signal:** aspirational
- **status-evidence:** plan(11) P5 "A vmhost adapter for what DOES need SSH … built to the analyser-adapter README skeleton"; earlier plan(1)/(4)/(5) P5 "registry + relocation (service_instances) and, eventually, the chassis service-deployer adapter".
- **what:** The eventual automation for what genuinely needs SSH: provision box, run setup.sh, onboard domain, ship engine, decommission — built as a `vmhost` adapter (cmd/vmhost-adapter, internal/adapters/vmhost, reuse thunder's ssh via shared/, kustomize, KafkaTopic system.adapter.vmhost.requests, 003 envelope), with a `service_instances` registry modelled on thunder_instances minus the reaper. Long-term it holds the deploy SSH credential, retiring the repo-secrets copy. Earlier versions named this the "service-deployer" adapter.
- **sources:** traffic_probe_plan(11).md#p5, traffic_probe_plan(1).md#phases, traffic_probe_running_notes(27).md#open-threads
- **relations:** future handler for backend_unreachable; supersedes "service-deployer" naming
- **verify-later:** analyser-adapter README; thunder_instances → service_instances

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### adapter response-envelope contract (request/response wiring conventions)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** adapter(4).go's header states the envelope is "grounded in the orchestrator (coordinator.go) and the three working adapters, not the docs — which disagree (003 vs FOCUS) and were resolved empirically." internal/adapters/analyser/ exists live in the repo, confirming the draft shipped.
- **what:** A reverse-engineered (from working adapters, not from docs) contract for how an adapter must shape its Kafka request/response envelope so the orchestrator actually routes the reply instead of timing out: action comes from `body.action` not headers; `in_response_to_request_id` (echoing the incoming `request_id`) is the load-bearing claim field in `coordinator.go`'s `ProcessResponse`, with `request_id` as fallback; the reply body must use canonical `types.ResponseHeaders` via `ToResponseHeaders` so `is_complete`/`is_error` marshal as real JSON bools (the "bool trap" — a `map[string]string` sending the string `"true"` fails the receiver's struct-bool unmarshal); sends must go via `ProduceWithValidation`, never plain `Produce`. websearch-adapter is flagged as the one adapter still on the deprecated string-bool/plain-Produce path.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/adapter(4).go, docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/analyser_request_action(1).go
- **relations:** contextkit toolchain (above); analyser-adapter service
- **verify-later:** platform/orchestration/coordinator.go ProcessResponse, internal/adapters/analyser/adapter.go, whether websearch-adapter has since been migrated off the string-bool map
