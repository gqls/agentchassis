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
