# Register — action-build-pipeline

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

1 concept, consolidated from 1 raw extraction across unit U22. Note: despite the
name overlap with the site-building "build pipeline" concepts elsewhere in this
cluster (register/build-pipeline.md, register/site-plan-and-reconciler.md), this
is a different sense of "build" — it is about compiling the chassis's own Go code,
not about building websites. Kept as its own file rather than merged into
build-pipeline.md to avoid conflating the two meanings.

### ABP-001 — Automated Go action build pipeline (compiler pod)
- **status:** aspirational
- **status-evidence:** "This is a medium-term investment"; the whole source doc is a design with a numbered ordered rollout and no deployment claim.
- **what:** A design for an in-cluster compiler pod that watches git for LLM-written Go action files, compiles the full chassis, runs tests, has a second-LLM review stage, builds an image via kaniko, and deploys per an HITL dial (manual→staging→auto) with rollback via a recorded previous_tag. Uses an `action_build_jobs` job/audit table; git stays the source of truth, replacing GitHub Actions for this purpose. Closes the loop: LLM identifies missing capability → writes action → compiled/tested/deployed → wires into workflow JSON.
- **sources:** docs020.../002_automated_go_action_create_and_build_pipeline.md
- **relations:** modular discovery-check registry (init() pattern); HITL dial pattern; tool-lifecycle; image_tag 'latest' stale-default trap (build-pipeline register, BLD-005 — a related but distinct image/tag hazard in the same deployment machinery)
- **verify-later:** action_build_jobs table; any compiler-service/ deployment
