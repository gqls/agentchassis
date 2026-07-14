
<!-- SOURCE: U11_traffic_probe.md -->
### VM-hosted backend sites — a new infrastructure class (proposed doc 024)
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** plan "Genuinely new (proposed doc 024 'VM-Hosted Backend Sites (site-engine)', Infrastructure Reference)" — the class runs live for one domain; the reference doc itself was only proposed ("Draft it in this thread once the shape is agreed", HANDOFF).
- **what:** The genuinely-new platform material the probe project surfaced: a persistent, non-reaped, internet-facing VM class and its lifecycle; DNS + public TLS as managed state outside k8s; a data-RETURN path from off-cluster; the off-cluster "commit is deploy" seam and where its credential lives (repo secrets now, adapter later); capability-gate semantics. Everything else was deliberately mapped onto existing mechanisms (adapter skeleton, thunder ssh, thunder_instances→service_instances, scheduled tasks, discovery checks, in-cluster Actions runner). Probe sites remain first-class `sites` rows so the maintenance/improvement loop covers them automatically — the discovery agents scan live sites over HTTP regardless of hosting.
- **sources:** traffic_probe_plan(12).md#framework-integration, HANDOFF#thread-b, traffic_probe_running_notes(28).md#2026-06-11 (integration mapping)
- **relations:** every concept below in this category; improvement-loop; adapters
- **verify-later:** whether docs024 doc "VM-Hosted Backend Sites" was ever written; sites rows with github_repo='vm-sites'

<!-- SOURCE: U11_traffic_probe.md -->
### site-engine — API-only capture backend for the class
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** HANDOFF: "Engine (site-engine, stdlib Go) live on a dedicated EU box for relojistas.com (CPX22, 167.233.33.159)".
- **what:** A single stdlib-only Go binary (zero deps, no go.sum by design) forked from idea.uk's service (kept: App/routes/cors shape, writeJSON, store pattern; dropped: engine/prompts/audience_check/billing). It does only what static files cannot: POST /intent (capture + 303 to THANKS_PATH), GET /api/hit (visit beacon), GET /stats (key-gated summary), GET /health, GET /events (export), GET /access-digest (log digest). nginx serves the chassis-built static site and proxies only these paths; the engine is never exposed directly, keyed by canonical Host, with ACCEPT_HOSTS as optional defence-in-depth. Explicitly class-level: "First feature: visitor-intent capture … the engine … grows by feature (e.g. chat, boards) later." Superseded first cut: a standalone "probe-go" multi-vhost page-serving service (session 1) — page rendering and per-domain content registry removed once the chassis owned the page.
- **sources:** deploy_setup/site-engine/service.go (header), traffic_probe_runbook(13).md#1-2, traffic_probe_running_notes(28).md#session-1-3, deploy_setup/site-engine/site-engine.env
- **relations:** engine store v2, /events endpoint, access-digest, setup.sh provisioning, idea.uk (fork origin)
- **verify-later:** gqls/site-engine repo contents; systemctl status site-engine on 167.233.33.159

<!-- SOURCE: U11_traffic_probe.md -->
### Engine store v2 — daily JSONL events + debounced counters + on-box retention
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-11: "Store v2 (JSONL) … Burst-tested: 300 events + 100 visits"; prune timer installed at go-live (relojistas_notes 2026-06-12 log).
- **what:** Two pre-launch storage cliffs were fixed structurally: v1 rewrote one ever-growing JSON file on every persist and held all events in RAM (superseded). v2 splits by access pattern — events append to daily JSONL (events-YYYYMMDD.jsonl, one line per submission, O(1) at any volume, rotation = the date, retention = delete old files); /stats counters live in a small counters.json flushed by a dirty-flag 5s debounced flusher (crash window ≤5s of visit counts, never events); SIGTERM/SIGINT flush+fsync. Retention enforced on-box by site-engine-prune.timer (daily delete of events files older than RETENTION_DAYS, default 90); explicitly NO logrotate on engine files (move/truncate would race the open handle) — nginx logs keep their own logrotate.
- **sources:** deploy_setup/site-engine/store.go (header), traffic_probe_running_notes(28).md#2026-06-11 (store fix + store v2 + retention), relojistas_notes(8).md#decisions
- **relations:** /events export (tails these lines), intent event record, minimal-data privacy (90-day prune)
- **verify-later:** store.go in site-engine repo; systemctl list-timers site-engine-prune.timer on box

<!-- SOURCE: U11_traffic_probe.md -->
### Probe as Layer 4 build + thin Layer 5 VM deploy (decisions D1–D4)
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** plan "Decisions — RESOLVED 2026-06-10" summary block.
- **what:** The structural framing that killed the standalone-project drift: a probe is a normal chassis-built site whose only differences are the deploy target and one capture component. D1: reuse the modern build-dispatch-loop pipeline (no separate probe workflow; pageflow-builder deprecation is a separate call). D2: a second shared repo for VM sites with the identical domain-folders-at-root layout; sites.github_repo selects the target; the static portfolio-sites repo + B2 Action stay untouched. D3: light per-repo Action now ("commit is deploy", target swapped); the heavier chassis-driven service-deployer is the eventual move. D4 moot: no needs_vm_deploy terminal item — the terminal build item stays target-agnostic (assemble + commit); the one-time per-domain VM setup is a separate provisioning step. Deferred: multi-box routing via deploy_config/service_instances only when relocation matters.
- **sources:** traffic_probe_plan(12).md#decisions-resolved + #decision-1-4 analysis, traffic_probe_running_notes(28).md#2026-06-10 (decisions resolved)
- **relations:** vm-sites repo + Action, github_repo target selector, vmhost adapter (the later heavy path), development-guide (build pipeline)
- **verify-later:** build-dispatch-loop handling a vm-sites-designated site end-to-end

<!-- SOURCE: U11_traffic_probe.md -->
### vm-sites content repo and deploy-to-vm Action
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** plan P2 "*Done: content Action deploy-to-vm.yml + engine Action deploy-engine-to-vm.yml … both validated*"; HANDOFF: "Deploy is 'commit is deploy' via two GitHub Actions … self-hosted runner."
- **what:** A standalone private repo (gqls/vm-sites; created BY HAND because the git-adapter auto-creates repos as PUBLIC; working checkout a sibling of agentchassis, never nested; the docs-tree copy is a reference snapshot only, contextkit pattern). Domain folders at repo ROOT — an assumption bug resolved 2026-06-11: the live sites repo keeps domain folders at root (the `sites/**` variant was a stale copy inside agentchassis/.git/workflows/, which GitHub never reads). The Action is a faithful sibling of the live B2 action: self-hosted runner, dotted-first-segment regex for changed-domain detection (structurally excludes .github/LICENSE/unknown-domain), full-sync fallback on empty diff, secret-presence checks, rsync -az --delete over SSH into /var/www/vm-sites/<domain>; no CF purge; deploys content only for already-provisioned domains. Deletion-propagation gap shared with the B2 action — noted, not fixed.
- **sources:** deploy_setup/vm-deploy/deploy-to-vm.yml (header), traffic_probe_running_notes(28).md#2026-06-11 (layout resolved; live b2 action learned), traffic_probe_runbook(13).md#3.1+5
- **relations:** deployment-github (B2 action sibling), setup.sh WEBROOT_OWNER, debugging lesson #24
- **verify-later:** gqls/vm-sites .github/workflows; Action run history

<!-- SOURCE: U11_traffic_probe.md -->
### site-engine deploy Action and the narrow-sudo privilege model
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** plan P2 done note; runbook §5; running notes 2026-06-12 "the 3.9 engine-seam test now SHIPS the endpoint".
- **what:** On push of **.go/go.mod to the engine repo: build linux/amd64 (static, stripped) → scp to box → run the root-owned hook /usr/local/sbin/site-engine-deploy which atomically swaps the binary and restarts. Privilege model: no root key in CI; setup.sh (when DEPLOY_USER set) installs the hook plus a sudoers rule scoped to ONLY that script — the deploy user can swap the engine and nothing else; the binary runs as the unprivileged site-engine user. Engine and content deploys are deliberately separate workflows so neither touches the other. x86-only constraint: the Action builds GOARCH=amd64 (Arm boxes would need a build-matrix change).
- **sources:** deploy_setup/vm-deploy/deploy-engine-to-vm.yml (header), traffic_probe_running_notes(28).md#2026-06-10 (engine-deploy workflow + privilege model), traffic_probe_runbook(13).md#5
- **relations:** setup.sh (installs the hook), dedicated-vs-shared box policy (x86 constraint)
- **verify-later:** sudoers rule + hook on box; Action run history in gqls/site-engine

<!-- SOURCE: U11_traffic_probe.md -->
### setup.sh — idempotent multi-vhost box provisioning
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** relojistas_notes log 2026-06-12 12:32: "Box provisioned (setup.sh full run)"; cert issued on idempotent re-run at 13:02.
- **what:** Adapted from idea.uk's authoritative setup.sh into the class-level provisioner: non-interactive (env-var params, positional domains fallback), idempotent (re-run IS rebuild; add a domain = extend DOMAINS + re-run; existing domains untouched), parameterised, self-contained (inline nginx conf + systemd unit). Installs per-domain vhosts serving /var/www/vm-sites/<domain> and proxying only the API paths; webroot certbot per domain with graceful HTTP degradation when DNS lags (re-run upgrades to HTTPS); ufw/fail2ban/logrotate/unattended-upgrades/ssh hardening; deploy sudo hook; prune timer; MODE=full|update. Grown options: WEBROOT_OWNER (deploy-user rsync rights), WWW_ALIAS (opt-in www server_name + cert SAN with getent pre-flight; v1 is apex-only), CLOUDFLARE=true (CF real_ip conf), per-domain access logs + adm group for the digest, version-neutral `listen 443 ssl` (nginx ≥1.25 http2 deprecation found in the field). Warning captured: box-takeover semantics (ufw --force reset, removes nginx default site) — why sharing the idea.uk box was declined.
- **sources:** deploy_setup/vm-deploy/setup.sh (header), traffic_probe_running_notes(28).md#2026-06-10 (box artifact) + 2026-06-12 entries, traffic_probe_runbook(13).md#3.5+4
- **relations:** site-engine deploy hook, multi-domain multiplexing, vmhost adapter (automates this later)
- **verify-later:** setup.sh in site-engine or vm-sites repo vs the docs-tree snapshot

<!-- SOURCE: U11_traffic_probe.md -->
### Multi-domain single-binary hosting and domain onboarding/relocation
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** runbook §4 documented + relojistas live; the shared multi-vhost box itself not yet provisioned (wayfaringlondoner "Awaiting a shared box + DNS").
- **what:** One engine binary per box behind many domains: per-domain nginx server_name blocks each serving that domain's web root and proxying the four API paths; the store keys events by host. Onboarding a new domain = DNS first, extend DOMAINS + re-run setup.sh (vhost + cert), deploy content, verify — the one-time step the content Action never does. Relocation = move web root + add to new box's DOMAINS + repoint DNS (instant if CF-proxied) + drop from old box. Design constraint discovered: THANKS_PATH is engine-wide (one env var per box), so all domains on a shared box must share a thanks filename — standard /thanks.html, each domain shipping its own; relojistas keeps /gracias.html on its dedicated box.
- **sources:** traffic_probe_runbook(13).md#4, wayfaringlondoner_notes.md#decisions, traffic_probe_running_notes(28).md#2026-06-13 (THANKS_PATH design point)
- **relations:** setup.sh, dedicated-vs-shared box policy, vmhost adapter (onboard-domain automation)
- **verify-later:** whether the shared box exists; wayfaringlondoner.com DNS/deployment state

<!-- SOURCE: U11_traffic_probe.md -->
### Dedicated vs shared box policy and VM sizing
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** relojistas_notes decisions 2026-06-11 (dedicated VM, hosting); HANDOFF: "no new boxes for now" (2026-06-13).
- **what:** Unknown-traffic experiments get their own box (relojistas: Hetzner CPX22, nbg1, IP 167.233.33.159 — sized by disk/log headroom, not CPU; even the claimed 1.2M visits/mo ≈ 0.5 req/s avg is far inside a small box); low-traffic domains share one multi-vhost box; the live idea.uk box is NOT reused (setup.sh box-takeover semantics + product coupling for a ~€3.49/mo saving). Bandwidth analysis: Hetzner EU cloud includes 20 TB/mo (avoid US/Singapore — slashed allowances); 1.2M visits ≈ 360 GB ≈ 2% of allowance. Stay on x86 (amd64 build). Policy hardened 2026-06-13: use EXISTING boxes only for new domains.
- **sources:** relojistas_notes(8).md#decisions+provenance (coordinates), traffic_probe_running_notes(28).md#2026-06-11 (sizing, bandwidth, box question), HANDOFF#where-things-stand
- **relations:** setup.sh takeover semantics, engine deploy Action x86 constraint
- **verify-later:** Hetzner project inventory; whether a shared box was later provisioned

<!-- SOURCE: U11_traffic_probe.md -->
### Pull-not-push off-cluster data return
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** relojistas_notes decision 2026-06-11 "No third 'collector' VM"; the pulling collector itself still disabled.
- **what:** The serving box only buffers (daily JSONL); the CLUSTER pulls over key-gated HTTPS on a schedule into clients_db. Rationale: pull keeps every credential in the cluster — boxes never hold DB or cluster secrets; a push model or middle VM inverts that, adds an attack surface and a hop for no gain. B2 remains optional cold backup. Collection therefore needs no adapter and no SSH — the engine already speaks key-gated HTTPS through nginx (the "key simplification" of P4). SSH is reserved for provisioning (P5).
- **sources:** relojistas_notes(8).md#decisions, traffic_probe_plan(12).md#P4, traffic_probe_running_notes(28).md#2026-06-11 (no collector VM; integration mapping)
- **relations:** /events endpoint, intent collection topology, vmhost adapter (the SSH half)
- **verify-later:** no box-side push cron/credentials exist; collector egress path

<!-- SOURCE: U11_traffic_probe.md -->
### requires-backend capability gate (Decision 5)
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** plan D5 "Outstanding: apply the planner query change"; component-side tag live (component inserted 2026-06-11); planner gate and audit check not applied.
- **what:** Gating backend-requiring sections off static sites keys on the CLASS (site has a server-side backend), not an instance label or site type. Component side: semantic tag `requires-backend` (on intent-probe; future chat/board sections carry the same). Planner side (to apply): load_components gains `AND NOT (COALESCE(semantic_tags,'[]'::jsonb) ? 'requires-backend')` so such components are opt-in via roadmap section_types only. Site side: deploy_config || {"target":"vm","capabilities":["backend"]} at onboarding. Later: an audit check comparing placed sections' requires-* tags against site capabilities → site_work_items findings. Supersedes the first design (an invented `intent-probe` site type in suitable_site_types + a suitable_site_types='[]' planner gate), corrected on operator feedback: "has a backend" is a property of the deploy target, not a site type.
- **sources:** traffic_probe_plan(12).md#decision-5, intent_probe_component(1).sql#gating, intent_probe_component.sql (family-delta: the superseded layer-1 gate), traffic_probe_running_notes(28).md#2026-06-10 (naming correction)
- **relations:** intent-probe component, site-plan-and-reconciler (build-site-planner load_components), design-composition
- **verify-later:** build-site-planner default_config load_components query; sites.deploy_config on any vm site

<!-- SOURCE: U11_traffic_probe.md -->
### P5 vmhost provisioning adapter and service_instances registry
- **category:** NEW:vm-backend-sites
- **status-signal:** aspirational
- **status-evidence:** plan P5 is entirely future-tense; HANDOFF Thread B lists it as pending; "P5 — registry + provisioning adapter" never marked started.
- **what:** The SSH half of the class, automating what runbook §3 does by hand: a `vmhost` adapter (analyser-adapter README skeleton: cmd/vmhost-adapter, internal/adapters/vmhost/ reusing thunder's ssh package via the shared/ precedent, configs, dockerfile, kustomize overlays, Makefile ×4, KafkaTopic system.adapter.vmhost.requests, 003 envelope contract) for provision-box / run setup.sh / onboard-domain (extend DOMAINS + re-run) / ship engine / decommission. Tracked in a `service_instances` table modelled on thunder_instances MINUS the reaper/uptime cap (persistent boxes are never reaped). Thin request actions + a deployer-family agent. Long-term the adapter holds the deploy SSH credential, retiring the repo-secrets copy.
- **sources:** traffic_probe_plan(12).md#P5 + #framework-integration, HANDOFF#thread-b, traffic_probe_running_notes(28).md#2026-06-11 (integration mapping)
- **relations:** adapters (thunder precedent, 003 envelope), setup.sh (what it automates), backend_unreachable (future handler)
- **verify-later:** any vmhost-adapter code/kustomize; service_instances table existence

<!-- SOURCE: U11_traffic_probe.md -->
### Cloudflare-proxied-in-front option
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** running notes 2026-06-13(f): "Cloudflare: relojistas now PROXIED (operator data: 22,046 SSL reqs/24h …)"; the real_ip conf ("set CLOUDFLARE=true on its next setup.sh re-run") still pending at last entry.
- **what:** Optional per-domain posture: keep DNS on Cloudflare with a proxied record → VM origin. Explicitly NOT a second Worker and not a second content copy (a Worker serving a copy would reintroduce the sync problem — avoid); the VM stays the single source of truth, CF just caches. Adjustments: cache-bypass the API paths; nginx set_real_ip_from CF ranges + real_ip_header CF-Connecting-IP (else rate-limiting throttles all of CF as one client and logs/digest/fail2ban see CF IPs); TLS Full (strict). Bonuses: CF-IPCountry populates the country field for free (engine default GeoHeader), and relocation becomes instant (change the origin IP) instead of DNS-TTL-bound.
- **sources:** traffic_probe_runbook(13).md#8, traffic_probe_running_notes(28).md#2026-06-10 (CF clarification) + 2026-06-13-f, passive_harvest_spec(2).md#cloudflare-note
- **relations:** access-digest (real-IP dependency), setup.sh CLOUDFLARE param, multi-domain relocation
- **verify-later:** relojistas CF zone config; cloudflare-realip.conf on box

<!-- SOURCE: U11_traffic_probe.md -->
### VM-hosted backend sites — a new infrastructure class (proposed doc 024)
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** plan "Genuinely new (proposed doc 024 'VM-Hosted Backend Sites (site-engine)', Infrastructure Reference)" — the class runs live for one domain; the reference doc itself was only proposed ("Draft it in this thread once the shape is agreed", HANDOFF).
- **what:** The genuinely-new platform material the probe project surfaced: a persistent, non-reaped, internet-facing VM class and its lifecycle; DNS + public TLS as managed state outside k8s; a data-RETURN path from off-cluster; the off-cluster "commit is deploy" seam and where its credential lives (repo secrets now, adapter later); capability-gate semantics. Everything else was deliberately mapped onto existing mechanisms (adapter skeleton, thunder ssh, thunder_instances→service_instances, scheduled tasks, discovery checks, in-cluster Actions runner). Probe sites remain first-class `sites` rows so the maintenance/improvement loop covers them automatically — the discovery agents scan live sites over HTTP regardless of hosting.
- **sources:** traffic_probe_plan(12).md#framework-integration, HANDOFF#thread-b, traffic_probe_running_notes(28).md#2026-06-11 (integration mapping)
- **relations:** every concept below in this category; improvement-loop; adapters
- **verify-later:** whether docs024 doc "VM-Hosted Backend Sites" was ever written; sites rows with github_repo='vm-sites'

<!-- SOURCE: U11_traffic_probe.md -->
### site-engine — API-only capture backend for the class
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** HANDOFF: "Engine (site-engine, stdlib Go) live on a dedicated EU box for relojistas.com (CPX22, 167.233.33.159)".
- **what:** A single stdlib-only Go binary (zero deps, no go.sum by design) forked from idea.uk's service (kept: App/routes/cors shape, writeJSON, store pattern; dropped: engine/prompts/audience_check/billing). It does only what static files cannot: POST /intent (capture + 303 to THANKS_PATH), GET /api/hit (visit beacon), GET /stats (key-gated summary), GET /health, GET /events (export), GET /access-digest (log digest). nginx serves the chassis-built static site and proxies only these paths; the engine is never exposed directly, keyed by canonical Host, with ACCEPT_HOSTS as optional defence-in-depth. Explicitly class-level: "First feature: visitor-intent capture … the engine … grows by feature (e.g. chat, boards) later." Superseded first cut: a standalone "probe-go" multi-vhost page-serving service (session 1) — page rendering and per-domain content registry removed once the chassis owned the page.
- **sources:** deploy_setup/site-engine/service.go (header), traffic_probe_runbook(13).md#1-2, traffic_probe_running_notes(28).md#session-1-3, deploy_setup/site-engine/site-engine.env
- **relations:** engine store v2, /events endpoint, access-digest, setup.sh provisioning, idea.uk (fork origin)
- **verify-later:** gqls/site-engine repo contents; systemctl status site-engine on 167.233.33.159

<!-- SOURCE: U11_traffic_probe.md -->
### Engine store v2 — daily JSONL events + debounced counters + on-box retention
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-11: "Store v2 (JSONL) … Burst-tested: 300 events + 100 visits"; prune timer installed at go-live (relojistas_notes 2026-06-12 log).
- **what:** Two pre-launch storage cliffs were fixed structurally: v1 rewrote one ever-growing JSON file on every persist and held all events in RAM (superseded). v2 splits by access pattern — events append to daily JSONL (events-YYYYMMDD.jsonl, one line per submission, O(1) at any volume, rotation = the date, retention = delete old files); /stats counters live in a small counters.json flushed by a dirty-flag 5s debounced flusher (crash window ≤5s of visit counts, never events); SIGTERM/SIGINT flush+fsync. Retention enforced on-box by site-engine-prune.timer (daily delete of events files older than RETENTION_DAYS, default 90); explicitly NO logrotate on engine files (move/truncate would race the open handle) — nginx logs keep their own logrotate.
- **sources:** deploy_setup/site-engine/store.go (header), traffic_probe_running_notes(28).md#2026-06-11 (store fix + store v2 + retention), relojistas_notes(8).md#decisions
- **relations:** /events export (tails these lines), intent event record, minimal-data privacy (90-day prune)
- **verify-later:** store.go in site-engine repo; systemctl list-timers site-engine-prune.timer on box

<!-- SOURCE: U11_traffic_probe.md -->
### Probe as Layer 4 build + thin Layer 5 VM deploy (decisions D1–D4)
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** plan "Decisions — RESOLVED 2026-06-10" summary block.
- **what:** The structural framing that killed the standalone-project drift: a probe is a normal chassis-built site whose only differences are the deploy target and one capture component. D1: reuse the modern build-dispatch-loop pipeline (no separate probe workflow; pageflow-builder deprecation is a separate call). D2: a second shared repo for VM sites with the identical domain-folders-at-root layout; sites.github_repo selects the target; the static portfolio-sites repo + B2 Action stay untouched. D3: light per-repo Action now ("commit is deploy", target swapped); the heavier chassis-driven service-deployer is the eventual move. D4 moot: no needs_vm_deploy terminal item — the terminal build item stays target-agnostic (assemble + commit); the one-time per-domain VM setup is a separate provisioning step. Deferred: multi-box routing via deploy_config/service_instances only when relocation matters.
- **sources:** traffic_probe_plan(12).md#decisions-resolved + #decision-1-4 analysis, traffic_probe_running_notes(28).md#2026-06-10 (decisions resolved)
- **relations:** vm-sites repo + Action, github_repo target selector, vmhost adapter (the later heavy path), development-guide (build pipeline)
- **verify-later:** build-dispatch-loop handling a vm-sites-designated site end-to-end

<!-- SOURCE: U11_traffic_probe.md -->
### vm-sites content repo and deploy-to-vm Action
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** plan P2 "*Done: content Action deploy-to-vm.yml + engine Action deploy-engine-to-vm.yml … both validated*"; HANDOFF: "Deploy is 'commit is deploy' via two GitHub Actions … self-hosted runner."
- **what:** A standalone private repo (gqls/vm-sites; created BY HAND because the git-adapter auto-creates repos as PUBLIC; working checkout a sibling of agentchassis, never nested; the docs-tree copy is a reference snapshot only, contextkit pattern). Domain folders at repo ROOT — an assumption bug resolved 2026-06-11: the live sites repo keeps domain folders at root (the `sites/**` variant was a stale copy inside agentchassis/.git/workflows/, which GitHub never reads). The Action is a faithful sibling of the live B2 action: self-hosted runner, dotted-first-segment regex for changed-domain detection (structurally excludes .github/LICENSE/unknown-domain), full-sync fallback on empty diff, secret-presence checks, rsync -az --delete over SSH into /var/www/vm-sites/<domain>; no CF purge; deploys content only for already-provisioned domains. Deletion-propagation gap shared with the B2 action — noted, not fixed.
- **sources:** deploy_setup/vm-deploy/deploy-to-vm.yml (header), traffic_probe_running_notes(28).md#2026-06-11 (layout resolved; live b2 action learned), traffic_probe_runbook(13).md#3.1+5
- **relations:** deployment-github (B2 action sibling), setup.sh WEBROOT_OWNER, debugging lesson #24
- **verify-later:** gqls/vm-sites .github/workflows; Action run history

<!-- SOURCE: U11_traffic_probe.md -->
### site-engine deploy Action and the narrow-sudo privilege model
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** plan P2 done note; runbook §5; running notes 2026-06-12 "the 3.9 engine-seam test now SHIPS the endpoint".
- **what:** On push of **.go/go.mod to the engine repo: build linux/amd64 (static, stripped) → scp to box → run the root-owned hook /usr/local/sbin/site-engine-deploy which atomically swaps the binary and restarts. Privilege model: no root key in CI; setup.sh (when DEPLOY_USER set) installs the hook plus a sudoers rule scoped to ONLY that script — the deploy user can swap the engine and nothing else; the binary runs as the unprivileged site-engine user. Engine and content deploys are deliberately separate workflows so neither touches the other. x86-only constraint: the Action builds GOARCH=amd64 (Arm boxes would need a build-matrix change).
- **sources:** deploy_setup/vm-deploy/deploy-engine-to-vm.yml (header), traffic_probe_running_notes(28).md#2026-06-10 (engine-deploy workflow + privilege model), traffic_probe_runbook(13).md#5
- **relations:** setup.sh (installs the hook), dedicated-vs-shared box policy (x86 constraint)
- **verify-later:** sudoers rule + hook on box; Action run history in gqls/site-engine

<!-- SOURCE: U11_traffic_probe.md -->
### setup.sh — idempotent multi-vhost box provisioning
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** relojistas_notes log 2026-06-12 12:32: "Box provisioned (setup.sh full run)"; cert issued on idempotent re-run at 13:02.
- **what:** Adapted from idea.uk's authoritative setup.sh into the class-level provisioner: non-interactive (env-var params, positional domains fallback), idempotent (re-run IS rebuild; add a domain = extend DOMAINS + re-run; existing domains untouched), parameterised, self-contained (inline nginx conf + systemd unit). Installs per-domain vhosts serving /var/www/vm-sites/<domain> and proxying only the API paths; webroot certbot per domain with graceful HTTP degradation when DNS lags (re-run upgrades to HTTPS); ufw/fail2ban/logrotate/unattended-upgrades/ssh hardening; deploy sudo hook; prune timer; MODE=full|update. Grown options: WEBROOT_OWNER (deploy-user rsync rights), WWW_ALIAS (opt-in www server_name + cert SAN with getent pre-flight; v1 is apex-only), CLOUDFLARE=true (CF real_ip conf), per-domain access logs + adm group for the digest, version-neutral `listen 443 ssl` (nginx ≥1.25 http2 deprecation found in the field). Warning captured: box-takeover semantics (ufw --force reset, removes nginx default site) — why sharing the idea.uk box was declined.
- **sources:** deploy_setup/vm-deploy/setup.sh (header), traffic_probe_running_notes(28).md#2026-06-10 (box artifact) + 2026-06-12 entries, traffic_probe_runbook(13).md#3.5+4
- **relations:** site-engine deploy hook, multi-domain multiplexing, vmhost adapter (automates this later)
- **verify-later:** setup.sh in site-engine or vm-sites repo vs the docs-tree snapshot

<!-- SOURCE: U11_traffic_probe.md -->
### Multi-domain single-binary hosting and domain onboarding/relocation
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** runbook §4 documented + relojistas live; the shared multi-vhost box itself not yet provisioned (wayfaringlondoner "Awaiting a shared box + DNS").
- **what:** One engine binary per box behind many domains: per-domain nginx server_name blocks each serving that domain's web root and proxying the four API paths; the store keys events by host. Onboarding a new domain = DNS first, extend DOMAINS + re-run setup.sh (vhost + cert), deploy content, verify — the one-time step the content Action never does. Relocation = move web root + add to new box's DOMAINS + repoint DNS (instant if CF-proxied) + drop from old box. Design constraint discovered: THANKS_PATH is engine-wide (one env var per box), so all domains on a shared box must share a thanks filename — standard /thanks.html, each domain shipping its own; relojistas keeps /gracias.html on its dedicated box.
- **sources:** traffic_probe_runbook(13).md#4, wayfaringlondoner_notes.md#decisions, traffic_probe_running_notes(28).md#2026-06-13 (THANKS_PATH design point)
- **relations:** setup.sh, dedicated-vs-shared box policy, vmhost adapter (onboard-domain automation)
- **verify-later:** whether the shared box exists; wayfaringlondoner.com DNS/deployment state

<!-- SOURCE: U11_traffic_probe.md -->
### Dedicated vs shared box policy and VM sizing
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** relojistas_notes decisions 2026-06-11 (dedicated VM, hosting); HANDOFF: "no new boxes for now" (2026-06-13).
- **what:** Unknown-traffic experiments get their own box (relojistas: Hetzner CPX22, nbg1, IP 167.233.33.159 — sized by disk/log headroom, not CPU; even the claimed 1.2M visits/mo ≈ 0.5 req/s avg is far inside a small box); low-traffic domains share one multi-vhost box; the live idea.uk box is NOT reused (setup.sh box-takeover semantics + product coupling for a ~€3.49/mo saving). Bandwidth analysis: Hetzner EU cloud includes 20 TB/mo (avoid US/Singapore — slashed allowances); 1.2M visits ≈ 360 GB ≈ 2% of allowance. Stay on x86 (amd64 build). Policy hardened 2026-06-13: use EXISTING boxes only for new domains.
- **sources:** relojistas_notes(8).md#decisions+provenance (coordinates), traffic_probe_running_notes(28).md#2026-06-11 (sizing, bandwidth, box question), HANDOFF#where-things-stand
- **relations:** setup.sh takeover semantics, engine deploy Action x86 constraint
- **verify-later:** Hetzner project inventory; whether a shared box was later provisioned

<!-- SOURCE: U11_traffic_probe.md -->
### Pull-not-push off-cluster data return
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** relojistas_notes decision 2026-06-11 "No third 'collector' VM"; the pulling collector itself still disabled.
- **what:** The serving box only buffers (daily JSONL); the CLUSTER pulls over key-gated HTTPS on a schedule into clients_db. Rationale: pull keeps every credential in the cluster — boxes never hold DB or cluster secrets; a push model or middle VM inverts that, adds an attack surface and a hop for no gain. B2 remains optional cold backup. Collection therefore needs no adapter and no SSH — the engine already speaks key-gated HTTPS through nginx (the "key simplification" of P4). SSH is reserved for provisioning (P5).
- **sources:** relojistas_notes(8).md#decisions, traffic_probe_plan(12).md#P4, traffic_probe_running_notes(28).md#2026-06-11 (no collector VM; integration mapping)
- **relations:** /events endpoint, intent collection topology, vmhost adapter (the SSH half)
- **verify-later:** no box-side push cron/credentials exist; collector egress path

<!-- SOURCE: U11_traffic_probe.md -->
### requires-backend capability gate (Decision 5)
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** plan D5 "Outstanding: apply the planner query change"; component-side tag live (component inserted 2026-06-11); planner gate and audit check not applied.
- **what:** Gating backend-requiring sections off static sites keys on the CLASS (site has a server-side backend), not an instance label or site type. Component side: semantic tag `requires-backend` (on intent-probe; future chat/board sections carry the same). Planner side (to apply): load_components gains `AND NOT (COALESCE(semantic_tags,'[]'::jsonb) ? 'requires-backend')` so such components are opt-in via roadmap section_types only. Site side: deploy_config || {"target":"vm","capabilities":["backend"]} at onboarding. Later: an audit check comparing placed sections' requires-* tags against site capabilities → site_work_items findings. Supersedes the first design (an invented `intent-probe` site type in suitable_site_types + a suitable_site_types='[]' planner gate), corrected on operator feedback: "has a backend" is a property of the deploy target, not a site type.
- **sources:** traffic_probe_plan(12).md#decision-5, intent_probe_component(1).sql#gating, intent_probe_component.sql (family-delta: the superseded layer-1 gate), traffic_probe_running_notes(28).md#2026-06-10 (naming correction)
- **relations:** intent-probe component, site-plan-and-reconciler (build-site-planner load_components), design-composition
- **verify-later:** build-site-planner default_config load_components query; sites.deploy_config on any vm site

<!-- SOURCE: U11_traffic_probe.md -->
### P5 vmhost provisioning adapter and service_instances registry
- **category:** NEW:vm-backend-sites
- **status-signal:** aspirational
- **status-evidence:** plan P5 is entirely future-tense; HANDOFF Thread B lists it as pending; "P5 — registry + provisioning adapter" never marked started.
- **what:** The SSH half of the class, automating what runbook §3 does by hand: a `vmhost` adapter (analyser-adapter README skeleton: cmd/vmhost-adapter, internal/adapters/vmhost/ reusing thunder's ssh package via the shared/ precedent, configs, dockerfile, kustomize overlays, Makefile ×4, KafkaTopic system.adapter.vmhost.requests, 003 envelope contract) for provision-box / run setup.sh / onboard-domain (extend DOMAINS + re-run) / ship engine / decommission. Tracked in a `service_instances` table modelled on thunder_instances MINUS the reaper/uptime cap (persistent boxes are never reaped). Thin request actions + a deployer-family agent. Long-term the adapter holds the deploy SSH credential, retiring the repo-secrets copy.
- **sources:** traffic_probe_plan(12).md#P5 + #framework-integration, HANDOFF#thread-b, traffic_probe_running_notes(28).md#2026-06-11 (integration mapping)
- **relations:** adapters (thunder precedent, 003 envelope), setup.sh (what it automates), backend_unreachable (future handler)
- **verify-later:** any vmhost-adapter code/kustomize; service_instances table existence

<!-- SOURCE: U11_traffic_probe.md -->
### Cloudflare-proxied-in-front option
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** running notes 2026-06-13(f): "Cloudflare: relojistas now PROXIED (operator data: 22,046 SSL reqs/24h …)"; the real_ip conf ("set CLOUDFLARE=true on its next setup.sh re-run") still pending at last entry.
- **what:** Optional per-domain posture: keep DNS on Cloudflare with a proxied record → VM origin. Explicitly NOT a second Worker and not a second content copy (a Worker serving a copy would reintroduce the sync problem — avoid); the VM stays the single source of truth, CF just caches. Adjustments: cache-bypass the API paths; nginx set_real_ip_from CF ranges + real_ip_header CF-Connecting-IP (else rate-limiting throttles all of CF as one client and logs/digest/fail2ban see CF IPs); TLS Full (strict). Bonuses: CF-IPCountry populates the country field for free (engine default GeoHeader), and relocation becomes instant (change the origin IP) instead of DNS-TTL-bound.
- **sources:** traffic_probe_runbook(13).md#8, traffic_probe_running_notes(28).md#2026-06-10 (CF clarification) + 2026-06-13-f, passive_harvest_spec(2).md#cloudflare-note
- **relations:** access-digest (real-IP dependency), setup.sh CLOUDFLARE param, multi-domain relocation
- **verify-later:** relojistas CF zone config; cloudflare-realip.conf on box
