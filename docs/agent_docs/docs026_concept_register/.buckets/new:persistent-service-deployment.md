
<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### idea.uk deployment topology — Docker/S3 plan superseded by systemd binary on a VM
- **category:** NEW:persistent-service-deployment
- **status-signal:** superseded
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)": "The 'Go-live checklist' above describes the original Docker/S3 plan. What's actually live differs and is the current truth: ... idea.uk runs as a single Go binary under systemd on a Hetzner box... — not Docker on a container host, and the landing page is embedded in the binary (`//go:embed page.html`), not a separate file on S3."
- **what:** The originally documented deploy plan (containerised `idea-svc` image + S3-hosted static landing page + separate deploy pipeline) was abandoned in favour of a much simpler shape once real deployment was attempted: one self-contained Go binary (page embedded via `go:embed`), deployed by build → scp → atomic `mv -f` swap → `systemctl restart`, behind nginx + Let's Encrypt on a single Hetzner VM. Explicitly flagged in `GUIDE_deploy_from_context_packs.md` as deploy-mechanism **F**, distinct from the chassis's k8s image path (A), DB/SQL path (B), work-items (C), orchestration triggers (D), and generated-static-sites-via-B2 path (E) — "Self-contained Go binary, file-based persistence, not k8s, not Backblaze."
- **sources:** `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)"; `docubundle_idea_golive/GUIDE_deploy_from_context_packs.md` §F; `running_notes(44).md` (VM provisioning checkpoints, 2026-06-04/05)
- **relations:** deploy-from-context-packs guide (six deploy mechanisms); service-deployer pattern (Path B automation of this same shape)
- **verify-later:** the box at 116.203.204.115 (Hetzner, Nuremberg); `/etc/idea/idea.env`; systemd unit `idea`

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### service-deployer pattern (persistent-VM automation, "Path B")
- **category:** NEW:persistent-service-deployment
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md` "PARALLEL THREAD — Layer 5 reassessed": "THE REAL GAP... A persistent service is the OPPOSITE [of Thunder's ephemeral VMs]: stays up, reaper-EXEMPT, holds its own credentials... So the gap = a persistent-service WRAPPER + credential delivery + DNS/TLS + a service_instances table + a parameterised setup script." Explicitly deferred: "Path A (manual now)... THEN build the service-deployer workflow around the proven script" — Path A was executed manually throughout this archive; Path B (the automated chassis workflow) was never built within it.
- **what:** A proposed chassis-native orchestrator, sibling of `model-trainer`, that would automate what was done by hand for idea.uk: provision a VM in *persistent* mode (reaper-exempt, unlike Thunder's ephemeral 18h-cap training VMs), ship the binary via the existing presigned-B2-URL mechanism, `ssh_exec` a parameterised `setup.sh`, deliver credentials, register in a new `service_instances` table, and health-check. The manual "Path A" run (deploying idea.uk by hand to a Hetzner box, iterating `setup.sh` against real-world failures — placeholder Let's Encrypt emails, systemd `EnvironmentFile` not stripping inline comments, etc.) was deliberately treated as *not throwaway* but as Path B's future payload/capture step.
- **sources:** `running_notes(44).md` ("PARALLEL_engine_deployment_and_layer5.md" summary, "CHECKPOINT 2026-06-04 (continued) — VM deploy artefacts drafted")
- **relations:** Thunder adapter (ephemeral VM precedent, explicitly contrasted); idea.uk deployment topology; deploy-from-context-packs guide (mechanism F)
- **verify-later:** whether `service_instances` table or a `service-deployer` agent definition exists in the live chassis

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### idea.uk deployment topology — Docker/S3 plan superseded by systemd binary on a VM
- **category:** NEW:persistent-service-deployment
- **status-signal:** superseded
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)": "The 'Go-live checklist' above describes the original Docker/S3 plan. What's actually live differs and is the current truth: ... idea.uk runs as a single Go binary under systemd on a Hetzner box... — not Docker on a container host, and the landing page is embedded in the binary (`//go:embed page.html`), not a separate file on S3."
- **what:** The originally documented deploy plan (containerised `idea-svc` image + S3-hosted static landing page + separate deploy pipeline) was abandoned in favour of a much simpler shape once real deployment was attempted: one self-contained Go binary (page embedded via `go:embed`), deployed by build → scp → atomic `mv -f` swap → `systemctl restart`, behind nginx + Let's Encrypt on a single Hetzner VM. Explicitly flagged in `GUIDE_deploy_from_context_packs.md` as deploy-mechanism **F**, distinct from the chassis's k8s image path (A), DB/SQL path (B), work-items (C), orchestration triggers (D), and generated-static-sites-via-B2 path (E) — "Self-contained Go binary, file-based persistence, not k8s, not Backblaze."
- **sources:** `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)"; `docubundle_idea_golive/GUIDE_deploy_from_context_packs.md` §F; `running_notes(44).md` (VM provisioning checkpoints, 2026-06-04/05)
- **relations:** deploy-from-context-packs guide (six deploy mechanisms); service-deployer pattern (Path B automation of this same shape)
- **verify-later:** the box at 116.203.204.115 (Hetzner, Nuremberg); `/etc/idea/idea.env`; systemd unit `idea`

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### service-deployer pattern (persistent-VM automation, "Path B")
- **category:** NEW:persistent-service-deployment
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md` "PARALLEL THREAD — Layer 5 reassessed": "THE REAL GAP... A persistent service is the OPPOSITE [of Thunder's ephemeral VMs]: stays up, reaper-EXEMPT, holds its own credentials... So the gap = a persistent-service WRAPPER + credential delivery + DNS/TLS + a service_instances table + a parameterised setup script." Explicitly deferred: "Path A (manual now)... THEN build the service-deployer workflow around the proven script" — Path A was executed manually throughout this archive; Path B (the automated chassis workflow) was never built within it.
- **what:** A proposed chassis-native orchestrator, sibling of `model-trainer`, that would automate what was done by hand for idea.uk: provision a VM in *persistent* mode (reaper-exempt, unlike Thunder's ephemeral 18h-cap training VMs), ship the binary via the existing presigned-B2-URL mechanism, `ssh_exec` a parameterised `setup.sh`, deliver credentials, register in a new `service_instances` table, and health-check. The manual "Path A" run (deploying idea.uk by hand to a Hetzner box, iterating `setup.sh` against real-world failures — placeholder Let's Encrypt emails, systemd `EnvironmentFile` not stripping inline comments, etc.) was deliberately treated as *not throwaway* but as Path B's future payload/capture step.
- **sources:** `running_notes(44).md` ("PARALLEL_engine_deployment_and_layer5.md" summary, "CHECKPOINT 2026-06-04 (continued) — VM deploy artefacts drafted")
- **relations:** Thunder adapter (ephemeral VM precedent, explicitly contrasted); idea.uk deployment topology; deploy-from-context-packs guide (mechanism F)
- **verify-later:** whether `service_instances` table or a `service-deployer` agent definition exists in the live chassis
