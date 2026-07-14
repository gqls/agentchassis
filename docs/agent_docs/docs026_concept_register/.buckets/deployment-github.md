
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Commit IS deploy (git → GitHub Actions → Backblaze B2, Cloudflare DNS-only)
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** 002(4) resolved decision 15; 034 shows the actual workflow
- **what:** Individual commits per work item; GH Actions fires per commit on a self-hosted runner, detects changed root-level domain directories, `b2 sync --delete --skip-newer` each to `b2://portfolio-sites/<domain>`, then purges Cloudflare cache per zone. No separate deploy step. The authoritative workflow lives in gqls/sites/.github/workflows — a stray copy under .git/workflows is a documented trap.
- **sources:** 002(4)#Git commit strategy; 034_github_action.md; 016 §0 item 24
- **relations:** git-adapter; "git committed is not proof of new content"
- **verify-later:** gqls/sites .github workflow; B2 bucket layout

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Chassis and site deploy model (single IMAGE_TAG; git → Actions → B2)
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** README_001_flow_notes: "There's a single global IMAGE_TAG (currently v1.0.1066) and one agent-chassis binary that runs every dynamic agent … Rollback is symmetric and cheap"; checkpoint (uu) repeats the make targets.
- **what:** Two deploy surfaces: (1) chassis code ships as one image tag running every dynamic agent via agent_definitions config — targeted path `make quick-agent-update IMAGE_TAG=…` (build → push → kustomize → DB image_tag → restart agent-chassis) plus `make update-and-restart-orchestrator` for the generic-orchestrator statefulset; full `make release` bumps every service; rollback repoints to the old existing image without rebuild. (2) Site content deploys git → GitHub Actions → Backblaze B2. Operational wrinkle: full-build deploys commit as "Rerender: <page>" — the shared message format no longer distinguishes build from rerender.
- **sources:** README_001_flow_notes.md; running_notes_checkpoint_uu.md#Deploy-rollback; HANDOFF_scheme_to_components_for_claude_code(1).md#Environment; running_notes_scheme_to_components(55).md#Th (hygiene)
- **relations:** deployed-binary-predates-disk; agent re-registration.
- **verify-later:** Makefile targets; agent_definitions.image_tag column use.

<!-- SOURCE: U05_content_quality_linking.md -->
### Git per-page deploy + non-fast-forward race
- **category:** deployment-github
- **status-signal:** partial
- **status-evidence:** running_notes_14(26) Part 10: git-adapter "updateRef is force:false + no-retry, so a concurrent commit to the shared sites repo loses with a silent non-fast-forward (FOCUS_dispatch open item 4)".
- **what:** Pages deploy as one git commit each to the shared sites repo (gqls/sites, git→Cloudflare); concurrent commits during a cascade can silently lose a non-fast-forward push. Suspected (not confirmed) in the missing-homepage case before the auto-complete cause was pinned; remains an open reliability item. Minor cosmetic sibling: page-rerender's commit message template "Rerender: {{.filename}}" renders uninterpolated.
- **sources:** running_notes_14(26).md#part-10; NOTES(44) minor findings
- **relations:** deploy gate on sections_saved; dispatch throughput.
- **verify-later:** git-adapter updateRef retry behaviour; commit-message templating scope.

<!-- SOURCE: U09_adoption.md -->
### Git-adapter cross-site commit race (force:false, no retry, shared sites repo)
- **category:** deployment-github
- **status-signal:** aspirational
- **status-evidence:** "a latent bug today for concurrent multi-site builds… Fix: retry-on-non-fast-forward in updateRef… tracked as a guardrail" ; "(2026-06-04: the missing-homepage lead was NOT this race… remains a latent risk with no confirmed instance yet.)"
- **what:** `git_commit` publishes to the git-adapter (Kafka key = correlation_id), which does a GitHub Git Data API read-modify-write with `force:false` and no retry against a single shared `sites` repo. Same-site commits serialize via partition (git-safe for same-site parallelism); different sites can hit different replicas concurrently and the loser's deploy fails silently on non-fast-forward. Proposed optimistic-concurrency retry (re-read HEAD, rebuild tree, re-commit) not yet built.
- **sources:** FOCUS_dispatch_throughput_and_claim_watchdog(3).md#git-deploy-path
- **relations:** Lever A prerequisite; A4 homepage (initially suspected, exonerated)
- **verify-later:** git-adapter `github_client.go` CommitToRepo/updateRef; adapter.go consumer model (2 replicas, group git.adapter.group)

<!-- SOURCE: U11_traffic_probe.md -->
### sites.github_repo as deploy-target selector (resolveGitRepoName patch)
- **category:** deployment-github
- **status-signal:** aspirational
- **status-evidence:** HANDOFF Thread B: "Chassis patch (P3, still pending) … Land this so the pipeline can target the VM repo at all"; plan P3 "Remaining: land the chassis patch".
- **what:** Tracing showed sites.github_repo is DORMANT end-to-end: git_commit reads config repo_name → default "sites"; upsertSite doesn't SELECT it; ensure_site_record doesn't return it. Specified 3-touch patch: (1) upsertSite RETURNING += COALESCE(github_repo,''), (2) EnsureSiteRecordAction return map += github_repo, (3) a private resolveGitRepoName(config, collected) helper (config repo_name → site_record.github_repo → "sites") used by BOTH git_commit and deploy_image_asset — the latter hardcodes "sites" at line 463 and would otherwise split-brain a probe site (pages → VM repo, images → sites). vet_med_export left alone (dedicated pipeline). Pre-flight confirmed github_repo empty on all 8 sites, so the fallback is safe. CommitToRepo already prefixes <domain>/ for any repo (shared root layout confirmed); createOrGetRepo auto-creates missing repos as PUBLIC — a deliberate-visibility trap.
- **sources:** traffic_probe_running_notes(28).md#2026-06-10 (P3 traced; repo surface complete), traffic_probe_plan(12).md#P3, HANDOFF#thread-b
- **relations:** vm-sites repo, D1–D4, requires-backend gate (the other half of onboarding)
- **verify-later:** grep resolveGitRepoName in platform/orchestration/actions/; deploy_image_asset repo_name resolution; whether the patch ever landed

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Deployment-GitHub / self-hosted runner + deploy path
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** 102_blog_handoff "Self-hosted GitHub Actions runner — deployed and running … Runner v2.333.1, pod in ai-persona-system namespace"
- **what:** The publish path: agents commit generated site files via a git adapter; a self-hosted GitHub Actions runner runs the sites-repo workflow which `b2 sync`s to Backblaze. `needs_rerender` is the terminal build item that assembles pages and triggers deployment.
- **sources:** ED/102_blog_handoff-2026-04-10.md#completed-this-session, WM/001_development_guide(0).md#every-pipeline-must-end-with-assembly-and-deployment, WM/007_adoption_pipeline_v3.md#data-flow-between-layers
- **relations:** blog-listing handoff; storage architecture (B2/S3); site plan reconciler terminal items
- **verify-later:** git-adapter; github-actions-runner dockerfile; needs_rerender handler

<!-- SOURCE: U20_legacy_docs_a.md -->
### Git deployment: commit_to_git + GitHub Action sync to B2
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** webbuild_pipeline/001pipeline: "Deployer commits sites/boxing-tickets.com/index.html to GitHub → GitHub Action automatically syncs that folder to B2 → Site is live."
- **what:** Deployment path: a git-adapter microservice (Kafka topic system.adapter.git.requests) commits generated site files to a repo (per-domain repos in the original design; a sites/<domain>/ folder in practice); a GitHub Action syncs to Backblaze B2 which serves the live site. GitCommitAction is the workflow-side action.
- **sources:** docs004_website_capture_project/webbuild_pipeline/001pipeline; docs004_website_capture_project/website_analysis/README.012.first_agent_definitions_etc.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** deployment-github (docs 034 live successor: git-adapter deploy surface); storage-architecture B2.
- **verify-later:** git-adapter service; the GitHub Action workflow file; sites/ repo layout.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Site manifest + external-edit desynchronisation detection
- **category:** deployment-github
- **status-signal:** aspirational
- **status-evidence:** Design tables in 004 (git_hook_adapter, Manifest Sync Agent, status 'desynchronized', HITL review) and manifest.json "winning genes" tracking in 008/014; no implementation evidence.
- **what:** Every generated site carries a manifest.json recording what built it (group_type, group_version, brief, build plan, component genes). A git webhook adapter watches all site repos; a human commit flags the manifest desynchronised, halting agent edit workflows and queueing HITL review — protecting human work from being overwritten and agents from stale state.
- **sources:** docs004_website_capture_project/website_analysis/README.004.backend.summary_ideas.md; docs004_website_capture_project/website_analysis/README.008.evolutionary_algorithm_of_site_portfolio.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion#versioning-model
- **relations:** content-governance locks (the live mechanism protecting human edits); deployment-github git-adapter.
- **verify-later:** any manifest.json in site repos; git webhook receivers.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Standardized per-page git deployment
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** docs017/002_standardising problem table ("Inconsistent deploy patterns → Standardize to per-page commits"); docs017/023 "Individual git commits per page → each goes live via GitHub Action → S3"; work items store commit_sha in result.
- **what:** Deployment converges on one mechanism: each page is committed individually via git_commit (with file_path override enabling CSS/asset commits), GitHub Actions deploy to hosting (Cloudflare, later S3), and commit SHAs are recorded on pages and work items for traceability. Removed redundant deployer steps whose data paths kept breaking.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/002_standardising_deployment_implementation_plan.md; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Git-Commit-Strategy
- **relations:** git-adapter; deployment-github category; per-page loop.
- **verify-later:** git_commit action file_path config; GitHub Action workflows in site repos.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Sites deployment chain (git → GitHub Actions → Backblaze B2) + image-tag chassis deploys
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** Used throughout ("after Actions propagates (a few minutes)"); 016b orientation: "Deployment is image-tag based... each agent's image_tag is bumped to adopt it; workflow (default_config) changes are DB-only and take effect immediately."
- **what:** Everything site-facing reaches production via git_commit to the 'sites' repo (files map keyed by repo-relative path — pages, tools/assets/*.js, assets/js/snippets.js, data/*.json) → GitHub Actions → Backblaze B2 (public), with the long-running git-adapter handling commits. Platform code ships as a chassis image (GitHub → Actions → image) adopted by bumping per-agent image_tag — so a source revert only reaches the cluster on the next build/push, while agent workflow changes (agent_definitions.default_config) are DB-only and instant. B2 404 NoSuchKey is the "page never deployed" signature.
- **sources:** docs/016b_debugging_guide_merged(3).md#orientation; docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~18:00 (git_commit pattern); docs/RUNBOOK_phase2_provocation_js(29).md#4.2-4.3; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§2
- **relations:** system-architecture; plan-storage revert note (pods keep old behaviour until push)
- **verify-later:** git-adapter; sites repo Actions workflow

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### github_repo target selection + resolveGitRepoName patch
- **category:** deployment-github
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-10 "sites.github_repo is DORMANT end-to-end … The patch (guide's 'small patch' pattern)"; plan(11) P3 "Remaining … land the chassis patch (resolveGitRepoName helper …)".
- **what:** A site's `sites.github_repo` selects deploy target (vm-sites repo vs default "sites"), but was dormant (upsertSite didn't SELECT it, nothing read it). The fix: one `resolveGitRepoName(config, collected)` helper (config repo_name → site_record.github_repo → "sites") used by both `git_commit` and `deploy_image_asset`, plus upsertSite RETURNING + ensure_site_record map additions. `deploy_image_asset` hardcoded "sites" and would split-brain a probe site (pages→VM, logo/hero→sites) without the same fallback.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-p3-repo-selection, traffic_probe_running_notes(27).md#2026-06-10-p3-pre-flight, traffic_probe_plan(11).md#p3
- **relations:** enables P3 pipeline wiring; deploy_image_asset split-brain risk
- **verify-later:** git_deployer_actions.go, site_db_actions.go, upsertSite, EnsureSiteRecordAction, deploy_image_asset:463

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### "Commit is deploy" seam swapped B2→VM + two GitHub Actions
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** runbook(12) §5 "Two separate workflows"; running_notes 2026-06-11 both siblings "rewritten as faithful siblings … Validated".
- **what:** The static "commit is deploy" seam is preserved, only the destination moves. Content Action (`deploy-to-vm.yml` in vm-sites repo): on push, rsync -az --delete each changed root-level `<domain>/` over SSH to `/var/www/vm-sites/<domain>`; self-hosted runner, no CF purge. Engine Action (`deploy-engine-to-vm.yml` in site-engine repo): on push to `**.go`/go.mod, build static stripped linux/amd64, scp, run the narrow `site-engine-deploy` sudo hook (atomic swap + restart). Secrets VM_HOST/VM_USER/VM_SSH_KEY.
- **sources:** traffic_probe_runbook(12).md#5, traffic_probe_running_notes(27).md#2026-06-10-vm-deploy-action, traffic_probe_running_notes(27).md#2026-06-11-live-b2-action
- **relations:** mirrors live deploy-to-b2.yml + Cloudflare Worker; target-agnostic terminal build item
- **verify-later:** vm-sites/.github/workflows/deploy-to-vm.yml, site-engine/.github/workflows/deploy-engine-to-vm.yml

<!-- SOURCE: U25_leopardess_social.md -->
### Chassis build/deploy practice (local Makefile builds, verify against the pod)
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_minilobby build practice (2026-07-10): "images are built from the local filesystem via the Makefile; commits are at the user's discretion … Do not infer a deployed binary's contents from git history; verify against the running pod."
- **what:** The chassis deployment reality: images build from the local working tree, decoupled from git commits (image tag hand-recorded in commit messages), so the deployed binary can lead or lag the repo; verification is kubectl exec + grep -ac for symbols in /app/agent-chassis. Corollary for site work: a committed Go change (e.g. the A6 imagery routing) is inert until a Makefile build+push. Pod logs are ephemeral across rollouts; spawned agents log in their own pods, not agent-chassis.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0 build practice, #2 log-hunting; docs/leopardessconsulting/HANDOFF.md#2-A6; docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10 (evening)
- **relations:** imagery kind routing (A6 awaiting deploy); operator discipline
- **verify-later:** Makefile build targets; deploy pipeline for the chassis image

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Commit IS deploy (git → GitHub Actions → Backblaze B2, Cloudflare DNS-only)
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** 002(4) resolved decision 15; 034 shows the actual workflow
- **what:** Individual commits per work item; GH Actions fires per commit on a self-hosted runner, detects changed root-level domain directories, `b2 sync --delete --skip-newer` each to `b2://portfolio-sites/<domain>`, then purges Cloudflare cache per zone. No separate deploy step. The authoritative workflow lives in gqls/sites/.github/workflows — a stray copy under .git/workflows is a documented trap.
- **sources:** 002(4)#Git commit strategy; 034_github_action.md; 016 §0 item 24
- **relations:** git-adapter; "git committed is not proof of new content"
- **verify-later:** gqls/sites .github workflow; B2 bucket layout

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Chassis and site deploy model (single IMAGE_TAG; git → Actions → B2)
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** README_001_flow_notes: "There's a single global IMAGE_TAG (currently v1.0.1066) and one agent-chassis binary that runs every dynamic agent … Rollback is symmetric and cheap"; checkpoint (uu) repeats the make targets.
- **what:** Two deploy surfaces: (1) chassis code ships as one image tag running every dynamic agent via agent_definitions config — targeted path `make quick-agent-update IMAGE_TAG=…` (build → push → kustomize → DB image_tag → restart agent-chassis) plus `make update-and-restart-orchestrator` for the generic-orchestrator statefulset; full `make release` bumps every service; rollback repoints to the old existing image without rebuild. (2) Site content deploys git → GitHub Actions → Backblaze B2. Operational wrinkle: full-build deploys commit as "Rerender: <page>" — the shared message format no longer distinguishes build from rerender.
- **sources:** README_001_flow_notes.md; running_notes_checkpoint_uu.md#Deploy-rollback; HANDOFF_scheme_to_components_for_claude_code(1).md#Environment; running_notes_scheme_to_components(55).md#Th (hygiene)
- **relations:** deployed-binary-predates-disk; agent re-registration.
- **verify-later:** Makefile targets; agent_definitions.image_tag column use.

<!-- SOURCE: U05_content_quality_linking.md -->
### Git per-page deploy + non-fast-forward race
- **category:** deployment-github
- **status-signal:** partial
- **status-evidence:** running_notes_14(26) Part 10: git-adapter "updateRef is force:false + no-retry, so a concurrent commit to the shared sites repo loses with a silent non-fast-forward (FOCUS_dispatch open item 4)".
- **what:** Pages deploy as one git commit each to the shared sites repo (gqls/sites, git→Cloudflare); concurrent commits during a cascade can silently lose a non-fast-forward push. Suspected (not confirmed) in the missing-homepage case before the auto-complete cause was pinned; remains an open reliability item. Minor cosmetic sibling: page-rerender's commit message template "Rerender: {{.filename}}" renders uninterpolated.
- **sources:** running_notes_14(26).md#part-10; NOTES(44) minor findings
- **relations:** deploy gate on sections_saved; dispatch throughput.
- **verify-later:** git-adapter updateRef retry behaviour; commit-message templating scope.

<!-- SOURCE: U09_adoption.md -->
### Git-adapter cross-site commit race (force:false, no retry, shared sites repo)
- **category:** deployment-github
- **status-signal:** aspirational
- **status-evidence:** "a latent bug today for concurrent multi-site builds… Fix: retry-on-non-fast-forward in updateRef… tracked as a guardrail" ; "(2026-06-04: the missing-homepage lead was NOT this race… remains a latent risk with no confirmed instance yet.)"
- **what:** `git_commit` publishes to the git-adapter (Kafka key = correlation_id), which does a GitHub Git Data API read-modify-write with `force:false` and no retry against a single shared `sites` repo. Same-site commits serialize via partition (git-safe for same-site parallelism); different sites can hit different replicas concurrently and the loser's deploy fails silently on non-fast-forward. Proposed optimistic-concurrency retry (re-read HEAD, rebuild tree, re-commit) not yet built.
- **sources:** FOCUS_dispatch_throughput_and_claim_watchdog(3).md#git-deploy-path
- **relations:** Lever A prerequisite; A4 homepage (initially suspected, exonerated)
- **verify-later:** git-adapter `github_client.go` CommitToRepo/updateRef; adapter.go consumer model (2 replicas, group git.adapter.group)

<!-- SOURCE: U11_traffic_probe.md -->
### sites.github_repo as deploy-target selector (resolveGitRepoName patch)
- **category:** deployment-github
- **status-signal:** aspirational
- **status-evidence:** HANDOFF Thread B: "Chassis patch (P3, still pending) … Land this so the pipeline can target the VM repo at all"; plan P3 "Remaining: land the chassis patch".
- **what:** Tracing showed sites.github_repo is DORMANT end-to-end: git_commit reads config repo_name → default "sites"; upsertSite doesn't SELECT it; ensure_site_record doesn't return it. Specified 3-touch patch: (1) upsertSite RETURNING += COALESCE(github_repo,''), (2) EnsureSiteRecordAction return map += github_repo, (3) a private resolveGitRepoName(config, collected) helper (config repo_name → site_record.github_repo → "sites") used by BOTH git_commit and deploy_image_asset — the latter hardcodes "sites" at line 463 and would otherwise split-brain a probe site (pages → VM repo, images → sites). vet_med_export left alone (dedicated pipeline). Pre-flight confirmed github_repo empty on all 8 sites, so the fallback is safe. CommitToRepo already prefixes <domain>/ for any repo (shared root layout confirmed); createOrGetRepo auto-creates missing repos as PUBLIC — a deliberate-visibility trap.
- **sources:** traffic_probe_running_notes(28).md#2026-06-10 (P3 traced; repo surface complete), traffic_probe_plan(12).md#P3, HANDOFF#thread-b
- **relations:** vm-sites repo, D1–D4, requires-backend gate (the other half of onboarding)
- **verify-later:** grep resolveGitRepoName in platform/orchestration/actions/; deploy_image_asset repo_name resolution; whether the patch ever landed

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Deployment-GitHub / self-hosted runner + deploy path
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** 102_blog_handoff "Self-hosted GitHub Actions runner — deployed and running … Runner v2.333.1, pod in ai-persona-system namespace"
- **what:** The publish path: agents commit generated site files via a git adapter; a self-hosted GitHub Actions runner runs the sites-repo workflow which `b2 sync`s to Backblaze. `needs_rerender` is the terminal build item that assembles pages and triggers deployment.
- **sources:** ED/102_blog_handoff-2026-04-10.md#completed-this-session, WM/001_development_guide(0).md#every-pipeline-must-end-with-assembly-and-deployment, WM/007_adoption_pipeline_v3.md#data-flow-between-layers
- **relations:** blog-listing handoff; storage architecture (B2/S3); site plan reconciler terminal items
- **verify-later:** git-adapter; github-actions-runner dockerfile; needs_rerender handler

<!-- SOURCE: U20_legacy_docs_a.md -->
### Git deployment: commit_to_git + GitHub Action sync to B2
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** webbuild_pipeline/001pipeline: "Deployer commits sites/boxing-tickets.com/index.html to GitHub → GitHub Action automatically syncs that folder to B2 → Site is live."
- **what:** Deployment path: a git-adapter microservice (Kafka topic system.adapter.git.requests) commits generated site files to a repo (per-domain repos in the original design; a sites/<domain>/ folder in practice); a GitHub Action syncs to Backblaze B2 which serves the live site. GitCommitAction is the workflow-side action.
- **sources:** docs004_website_capture_project/webbuild_pipeline/001pipeline; docs004_website_capture_project/website_analysis/README.012.first_agent_definitions_etc.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** deployment-github (docs 034 live successor: git-adapter deploy surface); storage-architecture B2.
- **verify-later:** git-adapter service; the GitHub Action workflow file; sites/ repo layout.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Site manifest + external-edit desynchronisation detection
- **category:** deployment-github
- **status-signal:** aspirational
- **status-evidence:** Design tables in 004 (git_hook_adapter, Manifest Sync Agent, status 'desynchronized', HITL review) and manifest.json "winning genes" tracking in 008/014; no implementation evidence.
- **what:** Every generated site carries a manifest.json recording what built it (group_type, group_version, brief, build plan, component genes). A git webhook adapter watches all site repos; a human commit flags the manifest desynchronised, halting agent edit workflows and queueing HITL review — protecting human work from being overwritten and agents from stale state.
- **sources:** docs004_website_capture_project/website_analysis/README.004.backend.summary_ideas.md; docs004_website_capture_project/website_analysis/README.008.evolutionary_algorithm_of_site_portfolio.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion#versioning-model
- **relations:** content-governance locks (the live mechanism protecting human edits); deployment-github git-adapter.
- **verify-later:** any manifest.json in site repos; git webhook receivers.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Standardized per-page git deployment
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** docs017/002_standardising problem table ("Inconsistent deploy patterns → Standardize to per-page commits"); docs017/023 "Individual git commits per page → each goes live via GitHub Action → S3"; work items store commit_sha in result.
- **what:** Deployment converges on one mechanism: each page is committed individually via git_commit (with file_path override enabling CSS/asset commits), GitHub Actions deploy to hosting (Cloudflare, later S3), and commit SHAs are recorded on pages and work items for traceability. Removed redundant deployer steps whose data paths kept breaking.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/002_standardising_deployment_implementation_plan.md; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Git-Commit-Strategy
- **relations:** git-adapter; deployment-github category; per-page loop.
- **verify-later:** git_commit action file_path config; GitHub Action workflows in site repos.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Sites deployment chain (git → GitHub Actions → Backblaze B2) + image-tag chassis deploys
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** Used throughout ("after Actions propagates (a few minutes)"); 016b orientation: "Deployment is image-tag based... each agent's image_tag is bumped to adopt it; workflow (default_config) changes are DB-only and take effect immediately."
- **what:** Everything site-facing reaches production via git_commit to the 'sites' repo (files map keyed by repo-relative path — pages, tools/assets/*.js, assets/js/snippets.js, data/*.json) → GitHub Actions → Backblaze B2 (public), with the long-running git-adapter handling commits. Platform code ships as a chassis image (GitHub → Actions → image) adopted by bumping per-agent image_tag — so a source revert only reaches the cluster on the next build/push, while agent workflow changes (agent_definitions.default_config) are DB-only and instant. B2 404 NoSuchKey is the "page never deployed" signature.
- **sources:** docs/016b_debugging_guide_merged(3).md#orientation; docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~18:00 (git_commit pattern); docs/RUNBOOK_phase2_provocation_js(29).md#4.2-4.3; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§2
- **relations:** system-architecture; plan-storage revert note (pods keep old behaviour until push)
- **verify-later:** git-adapter; sites repo Actions workflow

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### github_repo target selection + resolveGitRepoName patch
- **category:** deployment-github
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-10 "sites.github_repo is DORMANT end-to-end … The patch (guide's 'small patch' pattern)"; plan(11) P3 "Remaining … land the chassis patch (resolveGitRepoName helper …)".
- **what:** A site's `sites.github_repo` selects deploy target (vm-sites repo vs default "sites"), but was dormant (upsertSite didn't SELECT it, nothing read it). The fix: one `resolveGitRepoName(config, collected)` helper (config repo_name → site_record.github_repo → "sites") used by both `git_commit` and `deploy_image_asset`, plus upsertSite RETURNING + ensure_site_record map additions. `deploy_image_asset` hardcoded "sites" and would split-brain a probe site (pages→VM, logo/hero→sites) without the same fallback.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-p3-repo-selection, traffic_probe_running_notes(27).md#2026-06-10-p3-pre-flight, traffic_probe_plan(11).md#p3
- **relations:** enables P3 pipeline wiring; deploy_image_asset split-brain risk
- **verify-later:** git_deployer_actions.go, site_db_actions.go, upsertSite, EnsureSiteRecordAction, deploy_image_asset:463

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### "Commit is deploy" seam swapped B2→VM + two GitHub Actions
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** runbook(12) §5 "Two separate workflows"; running_notes 2026-06-11 both siblings "rewritten as faithful siblings … Validated".
- **what:** The static "commit is deploy" seam is preserved, only the destination moves. Content Action (`deploy-to-vm.yml` in vm-sites repo): on push, rsync -az --delete each changed root-level `<domain>/` over SSH to `/var/www/vm-sites/<domain>`; self-hosted runner, no CF purge. Engine Action (`deploy-engine-to-vm.yml` in site-engine repo): on push to `**.go`/go.mod, build static stripped linux/amd64, scp, run the narrow `site-engine-deploy` sudo hook (atomic swap + restart). Secrets VM_HOST/VM_USER/VM_SSH_KEY.
- **sources:** traffic_probe_runbook(12).md#5, traffic_probe_running_notes(27).md#2026-06-10-vm-deploy-action, traffic_probe_running_notes(27).md#2026-06-11-live-b2-action
- **relations:** mirrors live deploy-to-b2.yml + Cloudflare Worker; target-agnostic terminal build item
- **verify-later:** vm-sites/.github/workflows/deploy-to-vm.yml, site-engine/.github/workflows/deploy-engine-to-vm.yml

<!-- SOURCE: U25_leopardess_social.md -->
### Chassis build/deploy practice (local Makefile builds, verify against the pod)
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_minilobby build practice (2026-07-10): "images are built from the local filesystem via the Makefile; commits are at the user's discretion … Do not infer a deployed binary's contents from git history; verify against the running pod."
- **what:** The chassis deployment reality: images build from the local working tree, decoupled from git commits (image tag hand-recorded in commit messages), so the deployed binary can lead or lag the repo; verification is kubectl exec + grep -ac for symbols in /app/agent-chassis. Corollary for site work: a committed Go change (e.g. the A6 imagery routing) is inert until a Makefile build+push. Pod logs are ephemeral across rollouts; spawned agents log in their own pods, not agent-chassis.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0 build practice, #2 log-hunting; docs/leopardessconsulting/HANDOFF.md#2-A6; docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10 (evening)
- **relations:** imagery kind routing (A6 awaiting deploy); operator discipline
- **verify-later:** Makefile build targets; deploy pipeline for the chassis image
