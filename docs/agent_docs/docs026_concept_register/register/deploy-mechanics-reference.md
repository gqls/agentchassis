# Register — deploy-mechanics-reference

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

1 concept, consolidated from 2 raw extractions (1 unique block, appearing twice
due to exact whole-block duplication in the cluster input file) across unit
U24d.

### DMR-001 — Chassis deploy-mechanism reference (targets A–F)
- **status:** deployed
- **status-evidence:** Live `docubundle/GUIDE_deploy_from_context_packs.md` names six distinct deploy mechanisms (A: chassis image rebuild+rollout, B: DB/SQL migration, C: work-item insert, D: orchestration `orchestrate` trigger via kcat, E: generated static site via git→GitHub Actions→Backblaze, F: idea.uk standalone binary) and a per-project quick reference mapping each named task to its mechanism(s).
- **what:** A structured taxonomy of "what shipping a change actually means" per target: the agent-chassis Kubernetes image is a different deploy surface from the sites it builds (Backblaze-hosted static output) which is different again from the idea.uk box (file-based, cPanel, no k8s/DB). The archived draft only had this half-formed (a looser walkthrough focused on one task, skinner-box); the live version generalized it into the reusable A–F reference.
- **sources:** adoption/docubundle/GUIDE_deploy_from_context_packs(1).md; live docubundle/GUIDE_deploy_from_context_packs.md
- **relations:** adapters (033/035); deployment-github (034); storage-architecture (032)
- **verify-later:** confirm the A–F reference still matches current `makefile.txt`/`kustomization.yaml` targets

### DMR-002 — `make deploy-<service>`: single-service deploy with a registry pre-flight
- **status:** built and in use (first use 2026-07-28: `make deploy-browser-runner-adapter IMAGE_TAG=v1.0.1190`)
- **status-evidence:** pattern rule `deploy-%` in `makefile` (~line 1074, commit `35c8277a8`), with the rationale in the comment block above it. `make -n deploy-agents` verified unchanged — explicit targets keep priority in make, so `deploy-agents`, `deploy-infrastructure` and the numbered `deploy-0NN` targets are unaffected.
- **what:** Deploys ONE named service at `$(IMAGE_TAG)`, mirroring the build side's `build-%-ref` / `build-%-tree` pattern rules. It exists because `deploy-agents` is all-or-nothing: it seds *every* service's kustomization to `$(IMAGE_TAG)` and applies them, which is only safe when every service has been built and pushed at that tag — and on a normal day two or three tags exist across the fleet. Measured while rolling one fix: **2 of 14 backend services had that tag in the registry**, so a `deploy-agents` would have pointed twelve healthy deployments at images that were never pushed and ImagePullBackOff'd them together. The **registry pre-flight is the load-bearing part**: `push-*`/`deploy-*` are git-blind, so nothing downstream of the build checks the tag exists. It earned its place on its first dry run — another session had bumped `IMAGE_TAG` mid-task, so it resolved to `v1.0.1191` against an image built at `v1.0.1190`, and the guard refused instead of rolling onto a missing image.
- **sources:** `makefile` `deploy-%` rule; `docs/agent_docs/docs024_key_docs_latest/HANDOFF_2026-07-28_crash_recovery_open_bugfix_threads.md` §4
- **relations:** DMR-001 (mechanism A, chassis image rebuild+rollout); build-pipeline; multi-session coordination (the tag can move between your build and your deploy — pin `IMAGE_TAG=<what you actually built>`)
