
<!-- SOURCE: U22_recent_small_docs.md -->
### Automated Go action build pipeline (compiler pod)
- **category:** NEW:action-build-pipeline
- **status-signal:** aspirational
- **status-evidence:** "This is a medium-term investment"; whole doc is a design with a numbered ordered rollout, no deployment claim.
- **what:** A design for an in-cluster compiler pod that watches git for LLM-written Go action files, compiles the full chassis, runs tests, has a second-LLM review stage, builds an image via kaniko, and deploys per an HITL dial (manual→staging→auto) with rollback via recorded previous_tag. Uses an `action_build_jobs` job/audit table; git stays the source of truth, replacing GitHub Actions. Closes the loop: LLM identifies missing capability → writes action → compiled/tested/deployed → wires into workflow JSON.
- **sources:** docs020.../002_automated_go_action_create_and_build_pipeline.md
- **relations:** modular discovery-check registry (init() pattern), HITL, tool-lifecycle
- **verify-later:** action_build_jobs table; any compiler-service/ deployment

<!-- SOURCE: U22_recent_small_docs.md -->
### Automated Go action build pipeline (compiler pod)
- **category:** NEW:action-build-pipeline
- **status-signal:** aspirational
- **status-evidence:** "This is a medium-term investment"; whole doc is a design with a numbered ordered rollout, no deployment claim.
- **what:** A design for an in-cluster compiler pod that watches git for LLM-written Go action files, compiles the full chassis, runs tests, has a second-LLM review stage, builds an image via kaniko, and deploys per an HITL dial (manual→staging→auto) with rollback via recorded previous_tag. Uses an `action_build_jobs` job/audit table; git stays the source of truth, replacing GitHub Actions. Closes the loop: LLM identifies missing capability → writes action → compiled/tested/deployed → wires into workflow JSON.
- **sources:** docs020.../002_automated_go_action_create_and_build_pipeline.md
- **relations:** modular discovery-check registry (init() pattern), HITL, tool-lifecycle
- **verify-later:** action_build_jobs table; any compiler-service/ deployment
