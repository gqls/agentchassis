
<!-- SOURCE: U04_idea_uk.md -->
### Layer-5 gap = a persistent-service wrapper on already-deployed Thunder plumbing
- **category:** NEW:backend-service-deployment
- **status-signal:** aspirational
- **status-evidence:** "The hard plumbing for Layer 5 already exists and is deployed in production… The remaining gap is a persistent-service wrapper — modest, and largely assembling existing pieces" (2026-06-04); no service-deployer built.
- **what:** The honest reassessment of automated backend deployment: provisioning, ssh_exec, presigned-B2 file transfer, and decommission all exist (Thunder adapter, verified in production), but they're built for **ephemeral** training VMs (18h cap, 15-min reaper, credential-free). A persistent service is the exact opposite shape, so the gap is: persistent-mode provisioning (reaper exemption), credential delivery to the box, DNS+TLS wiring, a `service_instances` table (sibling of thunder_instances), and a parameterised setup script — a `service-deployer` orchestrator modelled on model-trainer, with idea.uk as first consumer. Two distinct things kept clear: deploying the engine binary to a VM (infrastructure) vs expressing the engine as chassis actions (Phase D) — complementary, not alternatives.
- **sources:** idea.uk/PARALLEL_engine_deployment_and_layer5.md; idea.uk/CONSOLIDATION_where_it_all_fits.md (Layer 5)
- **relations:** Thunder adapter (docs033); model-trainer pattern; Path A/setup.sh; 007 box recipe + site_api_routes.
- **verify-later:** thunder_instances table; absence of service_instances; cmd/thunder-adapter actions.

<!-- SOURCE: U04_idea_uk.md -->
### Path A manual VM deploy — setup.sh as the future service-deployer payload
- **category:** NEW:backend-service-deployment
- **status-signal:** deployed
- **status-evidence:** idea.uk LIVE on the Hetzner box 2026-06-05 via this path; setup.sh iterated through real incidents (certbot abort, env comments).
- **what:** "Do it by hand once, and capture the steps as the automation artefact": a single idempotent, non-interactive, parameterised `setup.sh` that converges a fresh Ubuntu box to nginx+TLS+ufw+fail2ban+unattended-upgrades+hardened systemd unit+binary — deliberately written so the chassis service-deployer can later `ssh_exec` the same file (MODE=update = binary swap; re-run = rebuild; anti-lockout guard on SSH password disable). The single-binary model: landing page `go:embed`ded, env in /etc/idea/idea.env, atomic mv-based redeploys.
- **sources:** idea.uk/nginx/README.md; idea.uk/nginx/setup.sh.orig3 (header); idea.uk/nginx/README_setup_box.md; idea.uk/PARALLEL_engine_deployment_and_layer5.md (Path A)
- **relations:** Layer-5 wrapper (Path B); page-serving gotchas; VM launch plan.
- **verify-later:** the live box's drift vs setup.sh (the doc's own rule: fold tweaks back in).

<!-- SOURCE: U04_idea_uk.md -->
### VM launch plan — dedicated hardened box, prior OVH reverse-proxy files audited
- **category:** NEW:backend-service-deployment
- **status-signal:** deployed
- **status-evidence:** Box provisioned 2026-06-04 (Hetzner CX, Nuremberg) following this plan; the year-old files' concrete bugs "all catalogued in the doc".
- **what:** Infrastructure-track decisions: a **dedicated** VM for idea.uk rather than the existing shared OVH multi-domain reverse proxy (blast-radius isolation; the proxy only knows how to reach k8s); reuse of the prior Terraform/nginx/fail2ban/logrotate/prometheus patterns with their specific year-old bugs fixed; secrets confirmed clean before reuse; VM sizing grounded in the engine being I/O-bound (1 vCPU / 512MB–1GB); search-grounded provider comparison (Hetzner vs Oracle vs spot).
- **sources:** idea.uk/nginx/VM_LAUNCH_PLAN.md; idea.uk/running_notes(63).md (2026-06-04 infra checkpoints)
- **relations:** Path A; 007 adoption-pipeline box recipe; Layer-5 wrapper.
- **verify-later:** the OVH proxy box's role for content sites (51.89.148.216 → k8s NodePort).

<!-- SOURCE: U04_idea_uk.md -->
### B2 dead-drop persistence: one-way flow from the exposed box into the framework DB
- **category:** NEW:backend-service-deployment
- **status-signal:** aspirational
- **status-evidence:** "Persistence decisions LOCKED (updated PERSISTENCE_design.md §10)" 2026-06-04 — design settled, but the live service still runs on orders.json only; no idea-ingest agent or idea_orders table evidenced.
- **what:** Standard tiered/DMZ design for internet-facing satellites: the exposed idea.uk box holds NO core-DB credentials and no network path to the cluster; it keeps a local operational store (JSON now; SQLite analysed — would break the stdlib-only property) and writes immutable terminal-event records to a scoped write-only B2 prefix (the dead-drop, reusing Thunder's presigned pattern); a scheduled in-cluster `idea-ingest` agent polls B2 and idempotently INSERTs into framework Postgres — the system of record. Kafka topic / narrow HTTPS ingest / direct PG all rejected (each is an inbound path in).
- **sources:** idea.uk/nginx/PERSISTENCE_design(1).md; idea.uk/running_notes(63).md (persistence checkpoints)
- **relations:** storage-architecture (B2, presigned URLs); scheduler-and-tasks (the ingest schedule); checkpoint-upload plan (same threat model).
- **verify-later:** any idea-events B2 prefix or idea_orders table (expect absent).

<!-- SOURCE: U04_idea_uk.md -->
### VM cutover: nginx front door with reserved tool paths (staging-in-place via DNS)
- **category:** NEW:backend-service-deployment
- **status-signal:** aspirational
- **status-evidence:** Runbook delivered 2026-06-21; "gated on P0 + the site review… deliberate, not done" (TODO P1, 2026-06-26).
- **what:** The go-live mechanism for a chassis-built front site on a VM that already hosts a live paid tool: because idea.uk's DNS (Cloudflare) points at the VM while the chassis deploys to B2, **every chassis build is invisible at the live domain — safe staging-in-place** — and cutover is purely an nginx change: static root for general pages, `location` proxies for the reserved tool paths (/request /confirm /approve /decline /stripe/webhook /internal/* /order/* /op /health /capacity + policy pages), `try_files … =404` so a missed tool path fails loudly, no body rewrites on the webhook location (signature integrity), prove the webhook through nginx BEFORE cutover, rollback = restore one server block. Named biggest risk: reserved-path completeness. Monorepo stays authoritative; the VM is just one more consumer (pull-sync from B2/git or a path-conditional Action push).
- **sources:** idea.uk/RUNBOOK_idea_uk_vm_cutover.md; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Phase 2); idea.uk/TODO_chassis_and_idea_uk(1).md#P1
- **relations:** scheme-to-components P0 (the gate); deployment-github (monorepo → Actions → B2); hybrid build_approach/hosting_trajectory classifier fields.
- **verify-later:** live nginx config on the box; whether cutover has since happened.

<!-- SOURCE: U04_idea_uk.md -->
### Layer-5 gap = a persistent-service wrapper on already-deployed Thunder plumbing
- **category:** NEW:backend-service-deployment
- **status-signal:** aspirational
- **status-evidence:** "The hard plumbing for Layer 5 already exists and is deployed in production… The remaining gap is a persistent-service wrapper — modest, and largely assembling existing pieces" (2026-06-04); no service-deployer built.
- **what:** The honest reassessment of automated backend deployment: provisioning, ssh_exec, presigned-B2 file transfer, and decommission all exist (Thunder adapter, verified in production), but they're built for **ephemeral** training VMs (18h cap, 15-min reaper, credential-free). A persistent service is the exact opposite shape, so the gap is: persistent-mode provisioning (reaper exemption), credential delivery to the box, DNS+TLS wiring, a `service_instances` table (sibling of thunder_instances), and a parameterised setup script — a `service-deployer` orchestrator modelled on model-trainer, with idea.uk as first consumer. Two distinct things kept clear: deploying the engine binary to a VM (infrastructure) vs expressing the engine as chassis actions (Phase D) — complementary, not alternatives.
- **sources:** idea.uk/PARALLEL_engine_deployment_and_layer5.md; idea.uk/CONSOLIDATION_where_it_all_fits.md (Layer 5)
- **relations:** Thunder adapter (docs033); model-trainer pattern; Path A/setup.sh; 007 box recipe + site_api_routes.
- **verify-later:** thunder_instances table; absence of service_instances; cmd/thunder-adapter actions.

<!-- SOURCE: U04_idea_uk.md -->
### Path A manual VM deploy — setup.sh as the future service-deployer payload
- **category:** NEW:backend-service-deployment
- **status-signal:** deployed
- **status-evidence:** idea.uk LIVE on the Hetzner box 2026-06-05 via this path; setup.sh iterated through real incidents (certbot abort, env comments).
- **what:** "Do it by hand once, and capture the steps as the automation artefact": a single idempotent, non-interactive, parameterised `setup.sh` that converges a fresh Ubuntu box to nginx+TLS+ufw+fail2ban+unattended-upgrades+hardened systemd unit+binary — deliberately written so the chassis service-deployer can later `ssh_exec` the same file (MODE=update = binary swap; re-run = rebuild; anti-lockout guard on SSH password disable). The single-binary model: landing page `go:embed`ded, env in /etc/idea/idea.env, atomic mv-based redeploys.
- **sources:** idea.uk/nginx/README.md; idea.uk/nginx/setup.sh.orig3 (header); idea.uk/nginx/README_setup_box.md; idea.uk/PARALLEL_engine_deployment_and_layer5.md (Path A)
- **relations:** Layer-5 wrapper (Path B); page-serving gotchas; VM launch plan.
- **verify-later:** the live box's drift vs setup.sh (the doc's own rule: fold tweaks back in).

<!-- SOURCE: U04_idea_uk.md -->
### VM launch plan — dedicated hardened box, prior OVH reverse-proxy files audited
- **category:** NEW:backend-service-deployment
- **status-signal:** deployed
- **status-evidence:** Box provisioned 2026-06-04 (Hetzner CX, Nuremberg) following this plan; the year-old files' concrete bugs "all catalogued in the doc".
- **what:** Infrastructure-track decisions: a **dedicated** VM for idea.uk rather than the existing shared OVH multi-domain reverse proxy (blast-radius isolation; the proxy only knows how to reach k8s); reuse of the prior Terraform/nginx/fail2ban/logrotate/prometheus patterns with their specific year-old bugs fixed; secrets confirmed clean before reuse; VM sizing grounded in the engine being I/O-bound (1 vCPU / 512MB–1GB); search-grounded provider comparison (Hetzner vs Oracle vs spot).
- **sources:** idea.uk/nginx/VM_LAUNCH_PLAN.md; idea.uk/running_notes(63).md (2026-06-04 infra checkpoints)
- **relations:** Path A; 007 adoption-pipeline box recipe; Layer-5 wrapper.
- **verify-later:** the OVH proxy box's role for content sites (51.89.148.216 → k8s NodePort).

<!-- SOURCE: U04_idea_uk.md -->
### B2 dead-drop persistence: one-way flow from the exposed box into the framework DB
- **category:** NEW:backend-service-deployment
- **status-signal:** aspirational
- **status-evidence:** "Persistence decisions LOCKED (updated PERSISTENCE_design.md §10)" 2026-06-04 — design settled, but the live service still runs on orders.json only; no idea-ingest agent or idea_orders table evidenced.
- **what:** Standard tiered/DMZ design for internet-facing satellites: the exposed idea.uk box holds NO core-DB credentials and no network path to the cluster; it keeps a local operational store (JSON now; SQLite analysed — would break the stdlib-only property) and writes immutable terminal-event records to a scoped write-only B2 prefix (the dead-drop, reusing Thunder's presigned pattern); a scheduled in-cluster `idea-ingest` agent polls B2 and idempotently INSERTs into framework Postgres — the system of record. Kafka topic / narrow HTTPS ingest / direct PG all rejected (each is an inbound path in).
- **sources:** idea.uk/nginx/PERSISTENCE_design(1).md; idea.uk/running_notes(63).md (persistence checkpoints)
- **relations:** storage-architecture (B2, presigned URLs); scheduler-and-tasks (the ingest schedule); checkpoint-upload plan (same threat model).
- **verify-later:** any idea-events B2 prefix or idea_orders table (expect absent).

<!-- SOURCE: U04_idea_uk.md -->
### VM cutover: nginx front door with reserved tool paths (staging-in-place via DNS)
- **category:** NEW:backend-service-deployment
- **status-signal:** aspirational
- **status-evidence:** Runbook delivered 2026-06-21; "gated on P0 + the site review… deliberate, not done" (TODO P1, 2026-06-26).
- **what:** The go-live mechanism for a chassis-built front site on a VM that already hosts a live paid tool: because idea.uk's DNS (Cloudflare) points at the VM while the chassis deploys to B2, **every chassis build is invisible at the live domain — safe staging-in-place** — and cutover is purely an nginx change: static root for general pages, `location` proxies for the reserved tool paths (/request /confirm /approve /decline /stripe/webhook /internal/* /order/* /op /health /capacity + policy pages), `try_files … =404` so a missed tool path fails loudly, no body rewrites on the webhook location (signature integrity), prove the webhook through nginx BEFORE cutover, rollback = restore one server block. Named biggest risk: reserved-path completeness. Monorepo stays authoritative; the VM is just one more consumer (pull-sync from B2/git or a path-conditional Action push).
- **sources:** idea.uk/RUNBOOK_idea_uk_vm_cutover.md; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Phase 2); idea.uk/TODO_chassis_and_idea_uk(1).md#P1
- **relations:** scheme-to-components P0 (the gate); deployment-github (monorepo → Actions → B2); hybrid build_approach/hosting_trajectory classifier fields.
- **verify-later:** live nginx config on the box; whether cutover has since happened.
