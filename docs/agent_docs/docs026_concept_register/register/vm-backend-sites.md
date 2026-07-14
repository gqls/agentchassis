# Register — vm-backend-sites

14 concepts, consolidated from 26 raw extractions of new:vm-backend-sites (13
unique blocks, each duplicated once in the source cluster file, unit
U11_traffic_probe), plus this file **absorbed new:backend-service-deployment**
(10 raw / 5 unique, unit U04_idea_uk) **and new:persistent-service-deployment**
(4 raw / 2 unique, unit U24e_docs_archive_idea_uk). All three categories turned
out to describe the same emerging architecture — "a persistent, non-reaped,
internet-facing VM class outside k8s" — from different angles: idea.uk was the
pioneering single instance (backend-service-deployment / persistent-service-deployment,
both sourced from the idea.uk archive), and traffic_probe's "VM-Hosted Backend
Sites" is the explicit generalisation of it into a class (its own setup.sh is
literally "adapted from idea.uk's authoritative setup.sh"). Two raw blocks were
additionally judged to belong better in sibling categories and were moved out
after merging with their counterparts there: the proposed automation adapter
("Layer-5 gap" / "service-deployer pattern" / "P5 vmhost provisioning adapter")
now lives as ADP-016 in register/adapters.md, and idea.uk's B2 dead-drop
persistence design now lives merged into STG-008 in register/storage-architecture.md.
One concept (Cloudflare-proxied-in-front option) absorbed a near-identical raw
block that the source cluster file had mis-tagged under the multicluster
category ("Cloudflare-in-front option", unit U24c_docs_archive_traffic_probe).

### VMB-001 — VM-hosted backend sites — a new infrastructure class (proposed doc 024)
- **status:** partial
- **status-evidence:** plan: "Genuinely new (proposed doc 024 'VM-Hosted Backend Sites (site-engine)', Infrastructure Reference)" — the class runs live for one domain (relojistas.com); the reference doc itself was only proposed ("Draft it in this thread once the shape is agreed", HANDOFF).
- **what:** The genuinely-new platform material the traffic_probe project surfaced when it needed a server-side capture backend outside k8s: a persistent, non-reaped, internet-facing VM class and its lifecycle; DNS + public TLS as managed state outside k8s; a data-RETURN path from off-cluster; the off-cluster "commit is deploy" seam and where its credential lives (repo secrets now, an adapter later); and capability-gate semantics for sites that need a backend. Everything else was deliberately mapped onto existing mechanisms (the adapter skeleton, Thunder's ssh precedent, `thunder_instances`→`service_instances`, scheduled tasks, discovery checks, the in-cluster Actions runner). Probe sites remain first-class `sites` rows so the maintenance/improvement loop covers them automatically — discovery agents scan live sites over HTTP regardless of hosting. This class is the direct generalisation of a narrower gap-analysis reached independently a few weeks earlier by the idea.uk project (the "Layer-5" persistent-service-wrapper gap, and its proposed "service-deployer"/"vmhost adapter" solution — both now tracked as register/adapters.md ADP-016, since idea.uk's own setup.sh is the literal ancestor of this class's setup.sh).
- **sources:** traffic_probe_plan(12).md#framework-integration, HANDOFF#thread-b, traffic_probe_running_notes(28).md#2026-06-11
- **relations:** every other concept in this file; vmhost/service-deployer adapter (register/adapters.md, ADP-016); improvement-loop
- **verify-later:** whether docs024 doc "VM-Hosted Backend Sites" was ever written; sites rows with github_repo='vm-sites'

### VMB-002 — site-engine — API-only capture backend for the class
- **status:** deployed
- **status-evidence:** HANDOFF: "Engine (site-engine, stdlib Go) live on a dedicated EU box for relojistas.com (CPX22, 167.233.33.159)."
- **what:** A single stdlib-only Go binary (zero deps, no go.sum by design) forked from idea.uk's service (kept: App/routes/cors shape, writeJSON, store pattern; dropped: engine/prompts/audience_check/billing). It does only what static files cannot: `POST /intent` (capture + 303 to THANKS_PATH), `GET /api/hit` (visit beacon), `GET /stats` (key-gated summary), `GET /health`, `GET /events` (export), `GET /access-digest` (log digest). nginx serves the chassis-built static site and proxies only these paths; the engine is never exposed directly, keyed by canonical Host, with `ACCEPT_HOSTS` as optional defence-in-depth. Explicitly framed as class-level: "First feature: visitor-intent capture … the engine … grows by feature (e.g. chat, boards) later." Superseded a first cut: a standalone "probe-go" multi-vhost page-serving service (session 1) — page rendering and the per-domain content registry were removed once the chassis owned the page.
- **sources:** deploy_setup/site-engine/service.go (header), traffic_probe_runbook(13).md#1-2, traffic_probe_running_notes(28).md#session-1-3, deploy_setup/site-engine/site-engine.env
- **relations:** JSON store scaling evolution (register/storage-architecture.md, STG-007); /events endpoint; access-digest; setup.sh provisioning (VMB-006); idea.uk (fork origin, VMB-012)
- **verify-later:** gqls/site-engine repo contents; systemctl status site-engine on 167.233.33.159

### VMB-003 — Probe as Layer 4 build + thin Layer 5 VM deploy (decisions D1-D4)
- **status:** deployed
- **status-evidence:** plan "Decisions — RESOLVED 2026-06-10" summary block.
- **what:** The structural framing that killed standalone-project drift: a probe is a normal chassis-built site whose only differences are the deploy target and one capture component. D1: reuse the modern build-dispatch-loop pipeline (no separate probe workflow; pageflow-builder deprecation is a separate call). D2: a second shared repo for VM sites with the identical domain-folders-at-root layout; `sites.github_repo` selects the target; the static portfolio-sites repo + B2 Action stay untouched. D3: a light per-repo Action now ("commit is deploy", target swapped); the heavier chassis-driven service-deployer is the eventual move. D4 moot: no `needs_vm_deploy` terminal item — the terminal build item stays target-agnostic (assemble + commit); the one-time per-domain VM setup is a separate provisioning step. Deferred: multi-box routing via `deploy_config`/`service_instances` only when relocation matters.
- **sources:** traffic_probe_plan(12).md#decisions-resolved,#decision-1-4-analysis, traffic_probe_running_notes(28).md#2026-06-10
- **relations:** vm-sites repo + Action (VMB-004); sites.github_repo deploy-target selector (register/deployment-github.md, DGH-003); vmhost adapter (register/adapters.md, ADP-016, the later heavy path)
- **verify-later:** build-dispatch-loop handling a vm-sites-designated site end-to-end

### VMB-004 — vm-sites content repo and deploy-to-vm Action
- **status:** deployed
- **status-evidence:** plan P2: "*Done: content Action deploy-to-vm.yml + engine Action deploy-engine-to-vm.yml … both validated*"; HANDOFF: "Deploy is 'commit is deploy' via two GitHub Actions … self-hosted runner"; runbook(12) §5 confirms both workflows "rewritten as faithful siblings … Validated."
- **what:** A standalone private repo (`gqls/vm-sites`; created BY HAND because the git-adapter auto-creates repos as PUBLIC; working checkout a sibling of agentchassis, never nested) holds domain folders at repo ROOT — an assumption bug resolved 2026-06-11 (the `sites/**` variant was a stale copy inside `agentchassis/.git/workflows/`, which GitHub never reads). The `deploy-to-vm.yml` Action is a faithful sibling of the live B2 action, preserving the same "commit is deploy" seam with only the destination swapped: self-hosted runner, dotted-first-segment regex for changed-domain detection (structurally excludes .github/LICENSE/unknown-domain), full-sync fallback on empty diff, secret-presence checks (VM_HOST/VM_USER/VM_SSH_KEY), `rsync -az --delete` over SSH into `/var/www/vm-sites/<domain>`; no CF purge; deploys content only for already-provisioned domains. The deletion-propagation gap shared with the B2 action is noted but not fixed.
- **sources:** deploy_setup/vm-deploy/deploy-to-vm.yml (header), traffic_probe_running_notes(28).md#2026-06-11, traffic_probe_runbook(13).md#3.1+5; traffic_probe_runbook(12).md#5, traffic_probe_running_notes(27).md#2026-06-10-vm-deploy-action
- **relations:** site-engine deploy Action (VMB-005, its sibling workflow); Commit-is-deploy (register/deployment-github.md, DGH-001, the B2-target original); setup.sh WEBROOT_OWNER (VMB-006)
- **verify-later:** gqls/vm-sites .github/workflows; Action run history; vm-sites/.github/workflows/deploy-to-vm.yml

### VMB-005 — site-engine deploy Action and the narrow-sudo privilege model
- **status:** deployed
- **status-evidence:** plan P2 done note; runbook §5; running notes 2026-06-12: "the 3.9 engine-seam test now SHIPS the endpoint."
- **what:** On push of `**.go`/go.mod to the engine repo: build linux/amd64 (static, stripped) → scp to box → run the root-owned hook `/usr/local/sbin/site-engine-deploy` which atomically swaps the binary and restarts. Privilege model: no root key in CI; `setup.sh` (when `DEPLOY_USER` is set) installs the hook plus a sudoers rule scoped to ONLY that script — the deploy user can swap the engine and nothing else; the binary itself runs as the unprivileged `site-engine` user. Engine and content deploys are deliberately separate workflows so neither touches the other. x86-only constraint: the Action builds `GOARCH=amd64` (Arm boxes would need a build-matrix change).
- **sources:** deploy_setup/vm-deploy/deploy-engine-to-vm.yml (header), traffic_probe_running_notes(28).md#2026-06-10,#2026-06-11-live-b2-action; traffic_probe_runbook(13).md#5; traffic_probe_runbook(12).md#5
- **relations:** vm-sites content repo and deploy-to-vm Action (VMB-004, its sibling workflow); setup.sh (VMB-006, installs the hook); Dedicated vs shared box policy (VMB-008, x86 constraint)
- **verify-later:** sudoers rule + hook on box; Action run history in gqls/site-engine; site-engine/.github/workflows/deploy-engine-to-vm.yml

### VMB-006 — setup.sh — idempotent multi-vhost box provisioning
- **status:** deployed
- **status-evidence:** relojistas_notes log 2026-06-12 12:32: "Box provisioned (setup.sh full run)"; cert issued on idempotent re-run at 13:02.
- **what:** Adapted from idea.uk's authoritative setup.sh (VMB-012) into the class-level provisioner: non-interactive (env-var params, positional domains fallback), idempotent (re-run IS rebuild; adding a domain = extend DOMAINS + re-run; existing domains untouched), self-contained (inline nginx conf + systemd unit). Installs per-domain vhosts serving `/var/www/vm-sites/<domain>` and proxying only the API paths; webroot certbot per domain with graceful HTTP degradation when DNS lags (re-run upgrades to HTTPS); ufw/fail2ban/logrotate/unattended-upgrades/ssh hardening; the deploy sudo hook (VMB-005); the prune timer (register/storage-architecture.md, STG-007); `MODE=full|update`. Options grown beyond idea.uk's original: `WEBROOT_OWNER` (deploy-user rsync rights), `WWW_ALIAS` (opt-in www server_name + cert SAN with getent pre-flight; v1 is apex-only), `CLOUDFLARE=true` (writes cloudflare-realip.conf, see VMB-011), per-domain access logs + adm group for the digest, version-neutral `listen 443 ssl` (nginx ≥1.25 http2-directive deprecation found in the field). Warning captured: box-takeover semantics (`ufw --force reset`, removes nginx default site) — why sharing the idea.uk box was declined (see VMB-008).
- **sources:** deploy_setup/vm-deploy/setup.sh (header), traffic_probe_running_notes(28).md#2026-06-10,#2026-06-12; traffic_probe_runbook(13).md#3.5+4
- **relations:** idea.uk VM deployment / Path A (VMB-012, the original this was adapted from); site-engine deploy Action (VMB-005); Multi-domain single-binary hosting (VMB-007); vmhost adapter (register/adapters.md, ADP-016, automates this later)
- **verify-later:** setup.sh in site-engine or vm-sites repo vs the docs-tree snapshot

### VMB-007 — Multi-domain single-binary hosting and domain onboarding/relocation
- **status:** partial
- **status-evidence:** runbook §4 documented + relojistas live; the shared multi-vhost box itself not yet provisioned as of this documentation (wayfaringlondoner "Awaiting a shared box + DNS").
- **what:** One engine binary per box behind many domains: per-domain nginx `server_name` blocks each serving that domain's web root and proxying the four API paths; the store keys events by host. Onboarding a new domain = DNS first, extend `DOMAINS` + re-run setup.sh (vhost + cert), deploy content, verify — the one-time step the content Action never does. Relocation = move web root + add to new box's DOMAINS + repoint DNS (instant if CF-proxied) + drop from old box. Design constraint discovered: `THANKS_PATH` is engine-wide (one env var per box), so all domains on a shared box must share a thanks filename — standardised on `/thanks.html`, each domain shipping its own; relojistas (on its own dedicated box) keeps `/gracias.html` instead.
- **sources:** traffic_probe_runbook(13).md#4, wayfaringlondoner_notes.md#decisions, traffic_probe_running_notes(28).md#2026-06-13
- **relations:** setup.sh (VMB-006); Dedicated vs shared box policy (VMB-008); vmhost adapter (register/adapters.md, ADP-016, onboard-domain automation)
- **verify-later:** whether the shared box exists; wayfaringlondoner.com DNS/deployment state

### VMB-008 — Dedicated vs shared box policy and VM sizing
- **status:** deployed
- **status-evidence:** relojistas_notes decisions 2026-06-11 (dedicated VM, hosting); HANDOFF: "no new boxes for now" (2026-06-13), hardened to "use EXISTING boxes only for new domains."
- **what:** Unknown-traffic experiments get their own box (relojistas: Hetzner CPX22, nbg1, IP 167.233.33.159 — sized by disk/log headroom, not CPU; even the claimed 1.2M visits/mo ≈ 0.5 req/s avg is far inside a small box); low-traffic domains share one multi-vhost box; the live idea.uk box is NOT reused (setup.sh box-takeover semantics + product coupling, for only a ~€3.49/mo saving — see VMB-013 for idea.uk's own, independently-made dedicated-box decision). Bandwidth analysis: Hetzner EU cloud includes 20 TB/mo (avoid US/Singapore — slashed allowances); 1.2M visits ≈ 360 GB ≈ 2% of allowance. Stay on x86 (amd64 build).
- **sources:** relojistas_notes(8).md#decisions+provenance, traffic_probe_running_notes(28).md#2026-06-11, HANDOFF#where-things-stand
- **relations:** VM launch plan — dedicated hardened box (VMB-013, idea.uk's parallel decision); setup.sh takeover semantics (VMB-006); site-engine deploy Action x86 constraint (VMB-005)
- **verify-later:** Hetzner project inventory; whether a shared box was later provisioned

### VMB-009 — Pull-not-push off-cluster data return
- **status:** partial
- **status-evidence:** relojistas_notes decision 2026-06-11 "No third 'collector' VM"; the pulling collector itself still disabled as of this documentation.
- **what:** The serving box only buffers (daily JSONL, see register/storage-architecture.md STG-007); the CLUSTER pulls over key-gated HTTPS on a schedule into `clients_db`. Rationale: pulling keeps every credential in the cluster — boxes never hold DB or cluster secrets; a push model or a middle VM inverts that, adding an attack surface and a hop for no gain. B2 remains an optional cold backup. Collection therefore needs no adapter and no SSH — the engine already speaks key-gated HTTPS through nginx (the "key simplification" of P4); SSH is reserved for provisioning only. This is a second, independently-arrived-at solution to the same design goal solved by idea.uk's earlier B2-dead-drop pattern (register/storage-architecture.md, STG-008) — kept as a separate concept here because the actual mechanism differs (scheduled HTTPS pull straight from the engine vs. a B2 dead-drop plus a polling ingest agent).
- **sources:** relojistas_notes(8).md#decisions, traffic_probe_plan(12).md#P4, traffic_probe_running_notes(28).md#2026-06-11
- **relations:** Persistence design (register/storage-architecture.md, STG-008, the idea.uk sibling solution); /events endpoint; vmhost adapter (register/adapters.md, ADP-016, the SSH half)
- **verify-later:** no box-side push cron/credentials exist; collector egress path

### VMB-010 — requires-backend capability gate (Decision 5)
- **status:** partial
- **status-evidence:** plan D5 "Outstanding: apply the planner query change"; the component-side semantic tag is live (component inserted 2026-06-11), but the planner gate and audit check were not yet applied.
- **what:** Gating backend-requiring sections off static sites keys on the CLASS (site has a server-side backend), not an instance label or site type. Component side: semantic tag `requires-backend` (on intent-probe; future chat/board sections carry the same). Planner side (to apply): `load_components` gains `AND NOT (COALESCE(semantic_tags,'[]'::jsonb) ? 'requires-backend')` so such components are opt-in via roadmap `section_types` only. Site side: `deploy_config || {"target":"vm","capabilities":["backend"]}` at onboarding. Later: an audit check comparing placed sections' `requires-*` tags against site capabilities → `site_work_items` findings. This design supersedes a first attempt (an invented `intent-probe` site type in `suitable_site_types` + a `suitable_site_types='[]'` planner gate), corrected on operator feedback: "has a backend" is a property of the deploy target, not a site type.
- **sources:** traffic_probe_plan(12).md#decision-5, intent_probe_component(1).sql#gating, intent_probe_component.sql, traffic_probe_running_notes(28).md#2026-06-10
- **relations:** intent-probe component; site-plan-and-reconciler (build-site-planner load_components); sites.github_repo deploy-target selector (register/deployment-github.md, DGH-003, the other half of onboarding)
- **verify-later:** build-site-planner default_config load_components query; sites.deploy_config on any vm site

### VMB-011 — Cloudflare-proxied-in-front option
- **status:** deployed
- **status-evidence:** running_notes 2026-06-13(f): "Cloudflare: relojistas now PROXIED (operator data: 22,046 SSL reqs/24h, 4,416 attacks blocked)" — confirming the option is not just designed but live and carrying real traffic.
- **what:** An optional per-domain posture: keep DNS on Cloudflare with a proxied (orange-cloud) record → VM origin, functioning as a reverse proxy — explicitly NOT a second Worker and not a second content copy (a Worker serving a copy would reintroduce the sync problem the whole class exists to avoid); the VM stays the single source of truth, CF just caches. Required adjustments: cache-bypass the API paths; nginx `set_real_ip_from` CF ranges + `real_ip_header CF-Connecting-IP` (else rate-limiting throttles all of CF as one client, and logs/digest/fail2ban see CF IPs instead of real visitors); TLS Full (strict). Bonuses: `CF-IPCountry` populates the country field for free (engine default GeoHeader), and relocation becomes instant (change the origin IP) instead of DNS-TTL-bound. `setup.sh CLOUDFLARE=true` writes the `cloudflare-realip.conf` adjustment (see VMB-006).
- **sources:** traffic_probe_runbook(13).md#8, traffic_probe_running_notes(28).md#2026-06-10,#2026-06-13-f, passive_harvest_spec(2).md#cloudflare-note, traffic_probe_runbook(12).md#8, traffic_probe_running_notes(27).md#2026-06-10-engine-deploy-workflow,#2026-06-13-f
- **relations:** access-digest (real-IP dependency); setup.sh CLOUDFLARE param (VMB-006); Multi-domain relocation (VMB-007)
- **verify-later:** relojistas CF zone config; /etc/nginx/conf.d/cloudflare-realip.conf on box; setup.sh CLOUDFLARE branch

### VMB-012 — idea.uk VM deployment: Path A manual setup.sh, systemd binary (supersedes the original Docker/S3 plan)
- **status:** deployed
- **status-evidence:** idea.uk LIVE on the Hetzner box since 2026-06-05 via this path; `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)" states plainly that the originally documented Docker/S3 plan was NOT what shipped: "What's actually live differs and is the current truth."
- **what:** The philosophy was "do it by hand once, and capture the steps as the automation artefact": a single idempotent, non-interactive, parameterised `setup.sh` converges a fresh Ubuntu box to nginx+TLS+ufw+fail2ban+unattended-upgrades+hardened systemd unit+binary, iterated through real incidents (certbot abort, env-file comment parsing). This deliberately abandoned the originally documented deploy plan (a containerised `idea-svc` image + S3-hosted static landing page + separate deploy pipeline) in favour of a much simpler shape once real deployment was attempted: one self-contained Go binary (landing page `go:embed`ded, not a separate S3 file), deployed by build → scp → atomic `mv -f` swap → `systemctl restart`, behind nginx + Let's Encrypt on a single Hetzner VM, with env in `/etc/idea/idea.env`. Explicitly flagged in `GUIDE_deploy_from_context_packs.md` as deploy-mechanism **F**, distinct from the chassis's k8s image path (A), DB/SQL path (B), work-items (C), orchestration triggers (D), and generated-static-sites-via-B2 path (E) — "Self-contained Go binary, file-based persistence, not k8s, not Backblaze." Written so the future chassis service-deployer/vmhost adapter (register/adapters.md, ADP-016) can later `ssh_exec` this same file (MODE=update = binary swap; re-run = rebuild; anti-lockout guard on SSH password disable) — the manual run was deliberately treated as that adapter's future payload/capture step, not throwaway work.
- **sources:** idea.uk/nginx/README.md, README_setup_box.md, setup.sh.orig3 (header); idea.uk/PARALLEL_engine_deployment_and_layer5.md (Path A); `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)"; `docubundle_idea_golive/GUIDE_deploy_from_context_packs.md` §F; `running_notes(44).md` (VM provisioning checkpoints, 2026-06-04/05)
- **relations:** setup.sh — idempotent multi-vhost box provisioning (VMB-006, the class-level fork of this same script); VM launch plan (VMB-013); VM cutover (VMB-014); vmhost/service-deployer adapter (register/adapters.md, ADP-016)
- **verify-later:** the live box's drift vs setup.sh; the box at 116.203.204.115 (Hetzner, Nuremberg); /etc/idea/idea.env; systemd unit `idea`

### VMB-013 — VM launch plan (idea.uk): dedicated hardened box, prior OVH reverse-proxy files audited
- **status:** deployed
- **status-evidence:** Box provisioned 2026-06-04 (Hetzner CX, Nuremberg) following this plan; the year-old OVH proxy files' concrete bugs were "all catalogued in the doc."
- **what:** Infrastructure-track decisions for idea.uk's launch: a **dedicated** VM rather than the existing shared OVH multi-domain reverse proxy (blast-radius isolation; the proxy only knows how to reach k8s); reuse of the prior Terraform/nginx/fail2ban/logrotate/prometheus patterns with their specific year-old bugs fixed; secrets confirmed clean before reuse; VM sizing grounded in the engine being I/O-bound (1 vCPU / 512MB-1GB); search-grounded provider comparison (Hetzner vs Oracle vs spot).
- **sources:** idea.uk/nginx/VM_LAUNCH_PLAN.md; idea.uk/running_notes(63).md (2026-06-04 infra checkpoints)
- **relations:** idea.uk VM deployment / Path A (VMB-012); Dedicated vs shared box policy (VMB-008, the traffic_probe project's parallel, independently-made decision); 007 adoption-pipeline box recipe
- **verify-later:** the OVH proxy box's role for content sites (51.89.148.216 → k8s NodePort)

### VMB-014 — VM cutover: nginx front door with reserved tool paths (staging-in-place via DNS)
- **status:** aspirational
- **status-evidence:** Runbook delivered 2026-06-21; "gated on P0 + the site review… deliberate, not done" (TODO P1, 2026-06-26).
- **what:** The go-live mechanism for a chassis-built front site on a VM that already hosts a live paid tool (idea.uk): because idea.uk's DNS (Cloudflare) points at the VM while the chassis deploys to B2, every chassis build is invisible at the live domain — safe staging-in-place — and cutover is purely an nginx change: static root for general pages, `location` proxies for the reserved tool paths (/request /confirm /approve /decline /stripe/webhook /internal/* /order/* /op /health /capacity + policy pages), `try_files … =404` so a missed tool path fails loudly, no body rewrites on the webhook location (signature integrity), prove the webhook through nginx BEFORE cutover, rollback = restore one server block. Named biggest risk: reserved-path completeness. The monorepo stays authoritative; the VM is just one more consumer (pull-sync from B2/git or a path-conditional Action push).
- **sources:** idea.uk/RUNBOOK_idea_uk_vm_cutover.md; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Phase 2); idea.uk/TODO_chassis_and_idea_uk(1).md#P1
- **relations:** scheme-to-components P0 (the gate); Commit-is-deploy (register/deployment-github.md, DGH-001); idea.uk VM deployment (VMB-012)
- **verify-later:** live nginx config on the box; whether cutover has since happened
