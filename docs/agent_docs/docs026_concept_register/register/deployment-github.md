# Register — deployment-github

5 concepts, consolidated from 26 raw extractions (13 unique blocks, each duplicated
once in the source cluster file) across units U01_docs024_numbered_core,
U03_idea_uk_section_data, U05_content_quality_linking, U09_adoption,
U11_traffic_probe, U17a_docs019_archive_discussions_and_main, U20_legacy_docs_a,
U21_legacy_docs_b, U23_docs_root_vonc, U24c_docs_archive_traffic_probe,
U25_leopardess_social. One raw block ("'Commit is deploy' seam swapped B2→VM +
two GitHub Actions", U24c) was recognised as fully covered — in more detail — by
two vm-backend-sites concepts (the VM content-deploy Action and the site-engine
deploy Action) and was folded there instead of being written up a third time here;
its unique framing ("the seam is preserved, only the destination moves") is noted
below and the full entries live in register/vm-backend-sites.md (VMB-004, VMB-005).

### DGH-001 — Commit-is-deploy: git → GitHub Actions (self-hosted runner) → Backblaze B2 (+ chassis image-tag deploys)
- **status:** deployed
- **status-evidence:** Consistently described as live and in current use across six independent sourcings spanning legacy (docs017/002, docs004) through current (docs/RUNNING_NOTES_vonc) documentation; 102_blog_handoff: "Self-hosted GitHub Actions runner — deployed and running … Runner v2.333.1, pod in ai-persona-system namespace."
- **what:** The platform's standing deploy mechanism, converged on and restated by every era of documentation: individual commits per work item (or per page) are pushed via the git-adapter (a Kafka microservice on `system.adapter.git.requests`, `GitCommitAction` workflow-side) to a shared `sites` repo; a self-hosted GitHub Actions runner (pod in `ai-persona-system`) fires per commit, detects changed root-level domain directories, and runs `b2 sync --delete --skip-newer` to `b2://portfolio-sites/<domain>`, then purges the Cloudflare cache per zone (Cloudflare is DNS/cache-only for this path, not a proxy). There is no separate deploy step — "commit is deploy." Commit SHAs are recorded on pages and work items for traceability; the authoritative workflow lives in `gqls/sites/.github/workflows` (a stray copy under `.git/workflows` is a documented trap that GitHub never reads). Separately, chassis platform code ships as one global `IMAGE_TAG` running every dynamic agent via `agent_definitions` config — `make quick-agent-update IMAGE_TAG=…` (build → push → kustomize → DB image_tag → restart agent-chassis) plus `make update-and-restart-orchestrator` for the generic-orchestrator statefulset; a full `make release` bumps every service; rollback is symmetric and cheap (repoint to the old existing image, no rebuild). Workflow/`default_config` changes are DB-only and take effect immediately, while a chassis code revert only reaches the cluster on the next build/push — these are two different deploy surfaces with different latencies. The identical "commit is deploy" seam was later reused for a VM deploy target (destination swapped, mechanism preserved) — see register/vm-backend-sites.md (VMB-004, VMB-005).
- **sources:** 002(4)#Git commit strategy; 034_github_action.md; README_001_flow_notes.md; docs017_legacy_agent_rules_images_design_keydocs/002_standardising_deployment_implementation_plan.md, 023_maintenance_architecture_unified_v6.md#Git-Commit-Strategy; docs004_website_capture_project/webbuild_pipeline/001pipeline; ED/102_blog_handoff-2026-04-10.md#completed-this-session; docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~18:00, docs/016b_debugging_guide_merged(3).md#orientation
- **relations:** git-adapter; Git-adapter non-fast-forward commit race (DGH-002); sites.github_repo deploy-target selector (DGH-003); Site manifest + external-edit desynchronisation detection (DGH-004); "git committed is not proof of new content"; deployed-binary-predates-disk (Chassis build/deploy practice, DGH-005)
- **verify-later:** gqls/sites .github workflow; B2 bucket layout; Makefile targets; agent_definitions.image_tag column use

### DGH-002 — Git-adapter non-fast-forward commit race (shared sites repo)
- **status:** aspirational
- **status-evidence:** "a latent bug today for concurrent multi-site builds… Fix: retry-on-non-fast-forward in updateRef… tracked as a guardrail"; explicitly noted as unconfirmed in production: "2026-06-04: the missing-homepage lead was NOT this race… remains a latent risk with no confirmed instance yet."
- **what:** `git_commit` publishes to the git-adapter (Kafka key = correlation_id), which does a GitHub Git Data API read-modify-write against the single shared `sites` repo with `force:false` and no retry. Same-site commits serialize safely via Kafka partition (same-site parallelism is git-safe), but different sites can hit different consumer replicas concurrently, and the loser's deploy fails silently on non-fast-forward push. This was initially suspected (but not confirmed) as the cause of a missing-homepage incident before an unrelated auto-complete bug was pinned as the real cause; it remains an open, unconfirmed reliability item. A proposed optimistic-concurrency retry (re-read HEAD, rebuild tree, re-commit) has not been built. Minor cosmetic sibling bug: the page-rerender commit-message template `"Rerender: {{.filename}}"` renders uninterpolated, and full-build deploys also commit as "Rerender: <page>" — the shared message format no longer distinguishes build from rerender.
- **sources:** running_notes_14(26).md#part-10; FOCUS_dispatch_throughput_and_claim_watchdog(3).md#git-deploy-path
- **relations:** Commit-is-deploy (DGH-001); deploy gate on sections_saved; dispatch throughput; Lever A prerequisite
- **verify-later:** git-adapter github_client.go CommitToRepo/updateRef retry behaviour; adapter.go consumer model (2 replicas, group git.adapter.group); commit-message templating scope

### DGH-003 — sites.github_repo deploy-target selector / resolveGitRepoName patch
- **status:** partial
- **status-evidence:** Tracing (2026-06-10) showed `sites.github_repo` DORMANT end-to-end on all 8 sites checked ("upsertSite doesn't SELECT it; ensure_site_record doesn't return it"); the chassis patch to fix it was still listed as pending in the HANDOFF ("Chassis patch (P3, still pending) … Land this so the pipeline can target the VM repo at all").
- **what:** `sites.github_repo` is meant to select the deploy target (e.g. the `vm-sites` repo vs the default `sites` repo) but was found dormant: nothing wrote it back from `upsertSite`, and nothing read it. The specified fix is a private `resolveGitRepoName(config, collected)` helper (config `repo_name` → `site_record.github_repo` → `"sites"`) used by BOTH `git_commit` and `deploy_image_asset` — the latter hardcodes `"sites"` at line 463 and would otherwise split-brain a probe/VM site (pages deploy to the VM repo, images/logos still deploy to the default `sites` repo). The 3-touch patch: (1) `upsertSite` RETURNING += `COALESCE(github_repo,'')`, (2) `EnsureSiteRecordAction` return map += `github_repo`, (3) the `resolveGitRepoName` helper itself. Pre-flight confirmed `github_repo` empty on all 8 sites checked, so the fallback to `"sites"` is safe. `CommitToRepo` already prefixes `<domain>/` for any repo (shared root layout confirmed); `createOrGetRepo` auto-creates missing repos as PUBLIC — a deliberate-visibility trap independent of this patch.
- **sources:** traffic_probe_running_notes(28).md#2026-06-10 (P3 traced; repo surface complete), traffic_probe_running_notes(27).md#2026-06-10-p3-repo-selection,#2026-06-10-p3-pre-flight; traffic_probe_plan(12).md#P3, traffic_probe_plan(11).md#p3; HANDOFF#thread-b
- **relations:** Commit-is-deploy (DGH-001); vm-sites content repo (register/vm-backend-sites.md, VMB-004); requires-backend capability gate (register/vm-backend-sites.md, VMB-010, the other half of onboarding)
- **verify-later:** grep resolveGitRepoName in platform/orchestration/actions/; deploy_image_asset repo_name resolution (line 463); whether the patch ever landed; git_deployer_actions.go, site_db_actions.go

### DGH-004 — Site manifest + external-edit desynchronisation detection
- **status:** aspirational
- **status-evidence:** Design tables (manifest schema, git_hook_adapter, Manifest Sync Agent, status 'desynchronized', HITL review) with no implementation evidence found.
- **what:** Every generated site would carry a `manifest.json` recording what built it (group_type, group_version, brief, build plan, component genes). A git webhook adapter would watch all site repos; a human commit would flag the manifest desynchronised, halting agent edit workflows and queueing a HITL review — protecting human work from being overwritten and agents from acting on stale state.
- **sources:** docs004_website_capture_project/website_analysis/README.004.backend.summary_ideas.md, README.008.evolutionary_algorithm_of_site_portfolio.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion#versioning-model
- **relations:** content-governance locks (the live mechanism that actually protects human edits); Commit-is-deploy (DGH-001); git-adapter
- **verify-later:** any manifest.json in site repos; git webhook receivers

### DGH-005 — Chassis build/deploy practice (local Makefile builds, verify against the pod)
- **status:** deployed
- **status-evidence:** RUNBOOK_minilobby build practice (2026-07-10): "images are built from the local filesystem via the Makefile; commits are at the user's discretion … Do not infer a deployed binary's contents from git history; verify against the running pod."
- **what:** The chassis deployment reality: images build from the local working tree, decoupled from git commits (the image tag is hand-recorded in commit messages), so the deployed binary can lead or lag the repo. Verification is `kubectl exec` + `grep -ac` for symbols in `/app/agent-chassis`, never inferred from git history. Corollary for site work: a committed Go change is inert until a Makefile build+push. Pod logs are ephemeral across rollouts; spawned agents log in their own pods, not agent-chassis.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0 build practice,#2 log-hunting; docs/leopardessconsulting/HANDOFF.md#2-A6; docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10
- **relations:** Commit-is-deploy (DGH-001, the site-content sibling of this chassis-image path); imagery kind routing; operator discipline
- **verify-later:** Makefile build targets; deploy pipeline for the chassis image
