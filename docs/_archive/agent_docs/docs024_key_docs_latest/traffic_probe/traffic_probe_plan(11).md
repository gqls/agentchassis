# Traffic-probe — plan

## How it all fits (plain English)
You own domains that still get visitors but show nothing. For each one we put
up a simple page that looks intentional for what the domain used to be, with
ONE box: "what are you looking for?". Whatever a visitor types is saved. After
a few weeks the saved searches tell you which domains have real demand worth
building a proper site for.

The moving parts, end to end:
1. **A small rented box (VM)** runs nginx + one tiny Go program
   (**site-engine**). nginx serves the page; the engine does the one thing a
   static page can't — save what people type. Each submission becomes one line
   in a daily file on the box; a 1×1 image on the page counts visits.
2. **The page** is, eventually, a normal chassis-built site (the pipeline you
   already run), deployed by "commit is deploy" — same as your B2 sites, just
   rsync-to-the-box instead of b2-sync, via a sibling GitHub Action. For the
   FIRST domain we hand-made the page so nothing blocks go-live.
3. **Reading the data:** now = ssh + a one-liner (commands in the domain notes);
   soon (P4) = the cluster pulls the lines over HTTPS on a schedule into a
   database table, and it becomes a SQL query / report. No extra VM involved.
4. **The framework sees these as normal sites** (a `sites` row whose
   `github_repo` points at the vm-sites repo), so the improvement/maintenance
   loop scans them like everything else; a small health check makes the loop
   notice a dead engine.
5. **Later (P5):** an adapter does over SSH what we're doing by hand now
   (provision a box, add a domain, ship the engine), tracked in a registry —
   at which point new probe domains are fully automated.

Where we ARE: engine + box scripts + both deploy Actions are built, tested,
and high-traffic-safe; the capture component is in the live library; the
relojistas go-live checklist is ready to run. Next: run it, watch the data,
then the P4 collector.


Phased plan with the live decision set and a decision log. Update as we go.

## Goal
Put residual-traffic domains on the chassis as first-class sites whose pages
plausibly reflect the old vertical and capture visitors' stated intent, deployed
to a VM (tiny backend) rather than B2, while staying inside the existing
maintenance/improvement loop. Use captured intent to rank domains worth building.

## Where we are
- Engine is an API-only capture backend; builds and tests pass.
- Framing settled: probe = Layer 4 build + thin Layer 5 VM deploy; not a separate
  project; reuse the build pipeline and the git→Actions seam (target swapped).
- Schema grounded; maintenance/improvement loop is automatic for real `sites`.

## Phases
- **P0 — agree the structural decisions below.** (done 2026-06-10)
- **P1 — manual go-live (Path A)** for 3–5 controlled domains; capture the exact
  box setup as `setup.sh` + nginx conf + systemd unit. (current)
  *2026-06-11: relojistas.com go-live bundle delivered — `relojistas_golive.md`
  (exact commands), `relojistas-site/` (Spanish search-probe page + gracias,
  grounded in the Wayback snapshot: it was a watch FORUM), repos/secrets/VM/
  verification steps. ~~One OPEN ITEM: the sites-repo layout~~ RESOLVED
  2026-06-11: live repo confirmed domain-folders-at-ROOT; deploy-to-vm.yml
  rewritten for root layout (paths-ignore trigger, first-segment detection,
  dot/file/deleted guards) and validated incl. a behavioural simulation. The
  `sites/**` variant was a stale copy in `agentchassis/.git/workflows/`.*
- **P2 — wire the deploy on update**: VM-sites repo + its VM-deploy Action
  ("commit is deploy", target = VM box). Static repo/Action untouched.
  *Done: content Action `deploy-to-vm.yml` + engine Action `deploy-engine-to-vm.yml`
  (+ `setup.sh` `DEPLOY_USER` hook), both validated. New-domain onboarding
  documented (runbook §4). Layout confirmed root-level; content Action rewritten
  to match (2026-06-11).*
- **P3 — make the probe a normal pipeline output**: a way to designate a site as
  VM/probe (its `github_repo` = VM repo selects the target) + the capture
  component the planner can include. No new terminal item (D4 moot); the one-time
  VM setup is a separate provisioning step.
  *Guideline-bound: designate via `sites.github_repo` (no schema change); STEP ZERO
  the component library before creating `intent-probe` (kebab `function`, v2 input
  schema, no-JS form → JS-separation-compliant); verify `git-adapter` reads
  `sites.github_repo` and how `github_repo` is set on a site, before any DB write.*
  *Progress 2026-06-10: traced repo selection — `github_repo` is dormant; 3-touch
  patch specified (upsertSite RETURNING + ensure_site_record map + git_commit
  fallback chain). `intent-probe` component drafted + validated
  (probe-pipeline/intent_probe_component.sql). NEW open decision D5 below.*
  *Progress 2026-06-11: component INSERTED into the live library (idempotency
  confirmed). Names resolved: `vm-sites` + `site-engine`; class-level rename
  applied across engine + deploy artifacts and revalidated (see running notes
  rename map). Remaining for P3: land the chassis patch (resolveGitRepoName
  helper + git_commit + deploy_image_asset + upsertSite/ensure_site_record),
  apply the planner load_components gate (D5), then the onboarding bundle
  (designation UPDATE + trigger payload with roadmap pinning intent-probe).*

### Decision 5 (REVISED 2026-06-10, per operator naming feedback) — gating
backend-requiring sections off static sites
The distinguishing feature is the CLASS (site has a server-side backend), not an
instance label like "intent-probe", and not a site type. Mechanism:
- **Component side:** semantic tag `requires-backend` (set on intent-probe; any
  future chat/board section carries the same tag).
- **Planner gate (to apply):** load_components gains
  `AND NOT (COALESCE(semantic_tags,'[]'::jsonb) ? 'requires-backend')` —
  requires-* components are opt-in via roadmap section_types only.
- **Site side:** `deploy_config || {"target":"vm","capabilities":["backend"]}`
  set at onboarding alongside `github_repo`.
- **Later:** one audit check comparing placed sections' requires-* tags against
  site capabilities → site_work_items findings.
Outstanding: apply the planner query change; confirm repo names (`vm-sites`?
`site-engine`?); decide on neutralising the "probe" infra defaults.
- **P4 — off-box collection + ranking (first chassis integration piece).**
  IN PROGRESS (this chat). Delivered: `intent_events` table migration
  (idempotent via unique engine_event_id; no extra checkpoint storage —
  since=max(event_created_at)) + disabled `intent-collection` scheduled_tasks
  row (improvement-sweep per-row dispatch pattern) + `CollectIntentEventsAction`
  skeleton (reads deploy_config.engine.{base_url,stats_key}, pulls /events
  NDJSON, upserts). Remaining: verify the action's DB/params accessors against
  live code + register it + the intent-collector agent definition; the
  access-log harvest (referer/landing-path/404/UA) as a second collector mode;
  the `backend_unreachable` discovery check; ranking query. Per-site onboarding
  UPDATE extended to carry deploy_config.engine.* (in the migration file).
  Key simplification: collection needs NO adapter and NO SSH. The engine already
  speaks key-gated HTTPS through nginx; add one `GET /events?since=` export
  endpoint to site-engine, then ONE Go action (`vm_collect_intent_events`:
  HTTPS pull + parameterised upsert into a new `intent_events` table) run by a
  scheduled agent (kafka-scheduler + scheduled_tasks, doc 010). Ranking =
  a query over the table. Add the `backend_unreachable` discovery check
  (discovery_checks/ + sweep, doc 004) → site_work_items findings — the
  improvement-loop tie-in. **Ingest validation contract:** parameterised SQL
  only (values are data, never concatenated — injection structurally
  impossible per the house rule); per-line shape checks (JSON parses, kind ∈
  enum, value ≤500 runes, host ∈ accepted set, timestamp sane); burst dedupe
  of identical (host,value) within a minute as bot noise (raw JSONL stays the
  source of truth); **Unicode NFC normalisation + lowercasing for the
  aggregate views happen HERE** (the engine deliberately defers them — it is
  stdlib-only, so it strips Cc/Cf and collapses whitespace at capture while
  NFD combining marks pass through; without NFC, `ñ` in two byte-forms counts
  as two terms); **plus an access-log harvest** — the inbound external referer,
  landing path + query, the forum-404 intent paths, and the user-agent (for
  bot classification) are all already in nginx's combined log and are NOT
  observable at the engine on a static page load; the collector parses the log
  per domain for these (the engine's structured events additionally carry
  `landing_query` on submissions). DB CHECK constraints on the `intent_events`
  table; values escaped at every display surface. Open choice: redact
  email/phone patterns at ingest vs rely on the 90-day prune. *If any of this
  is ever built as a chassis adapter, the 003 envelope contract applies.*
- **P5 — registry + provisioning adapter.** A `vmhost` adapter for what DOES
  need SSH (provision box, run setup.sh, onboard domain = extend DOMAINS +
  re-run, ship engine, decommission), built to the analyser-adapter README
  skeleton: `cmd/vmhost-adapter` + `internal/adapters/vmhost/` (reuse thunder's
  ssh package via the `internal/adapters/shared/` precedent), `configs/`,
  dockerfile, kustomize `services/vmhost-adapter/{base,overlays/production/
  uk_001}`, Makefile ×4, KafkaTopic `system.adapter.vmhost.requests`, 003
  envelope. `service_instances` table modelled on thunder_instances MINUS the
  reaper/uptime cap. Thin actions + a deployer-family agent. Long-term the
  adapter holds the deploy SSH credential, retiring the repo-secrets copy.

## Framework integration — what ISN'T new (mapping, 2026-06-11)
| Need | Existing mechanism |
|---|---|
| Adapter skeleton/wiring | analyser-adapter README pattern (cmd/, internal/adapters/, configs/, dockerfile, kustomize, Makefile, KafkaTopic, 003 envelope) |
| SSH to remote VMs | thunder-adapter `internal/adapters/thunder/ssh` (+ `shared/` precedent) |
| Instance registry | `thunder_instances` → `service_instances` (no reaper) |
| Thin request actions | analyser/thunder action pattern in `platform/orchestration/actions/` |
| Orchestrating agent | deployer family precedent in agent_definitions |
| Periodic collection | kafka-scheduler + scheduled_tasks (doc 010) |
| Health surveillance | discovery_checks/ + 600s sweep → site_work_items (doc 004); `endpoint-health-checker` agent may be reusable (verify) |
| CI runner | in-cluster `github-actions-runner` service exists |
| Data into clients_db | small table + parameterised upsert action |

**Genuinely new (proposed doc 024 “VM-Hosted Backend Sites (site-engine)”,
Infrastructure Reference; numbering operator's):** a persistent, non-reaped,
internet-facing VM class and its lifecycle; DNS + public TLS as managed state
outside k8s; a data-RETURN path from off-cluster; the off-cluster
"commit is deploy" seam and where its credential lives (repo secrets now,
adapter later); the capability gate semantics (D5, partly done).

---

## Decisions — RESOLVED 2026-06-10
Summary: **D1** reuse the `build-dispatch-loop` pipeline (no separate build
workflow). **D2** a second shared repo for VM sites, identical domain-subpath
layout, its own VM-deploy Action; `sites.github_repo` selects target; static
`portfolio-sites` repo + B2 Action untouched. **D3** light per-repo Action,
terminal item target-agnostic. **D4** moot — no `needs_vm_deploy` terminal item;
the one-time per-domain VM setup (provision/DNS/install/register/`ACCEPT_HOSTS`)
is a separate step (Path A now, provisioning step/`service-deployer` later).
Deferred: one VM repo → one Action → one box to start; per-domain routing to
multiple boxes (via `deploy_config`/`service_instances`) only when relocation
matters. Analysis retained below for the record.


### Decision 1 — separate workflow for probe sites, or reuse the pipeline?
Context: the difference between a probe site and a static site is concentrated at
**deploy** (VM vs B2) and one **capture component**; research/plan/write/design/
assemble are the same. There are two build generations in the registry: the older
monolithic `pageflow-builder` (orchestrator with sub-workflows/loops) and the
modern `build-pipeline-trigger → build-dispatch-loop → handler agents` (one work
item per invocation, clean logs) — the modern one already addresses the
"monolithic" concern.
- **Option A — reuse the dispatch-loop pipeline**, add only a capture component +
  a deploy step. Least duplication; improvement loop covers it automatically.
- **Option B — a thin probe entry orchestrator** (sibling to the existing
  `*-builder` agents) that delegates to the SAME specialist/handler agents and
  differs only at the terminal step. Distinct responsibility/clean logs without
  duplicating build logic.
- **Option C — a fully separate build workflow.** Cleanest isolation; highest
  divergence/duplication risk and must re-justify improvement-loop coverage.
- Leaning: A or B (B if a distinct entry point is wanted), not C.

### Decision 2 — repo layout
- **Repo-per-domain:** likely overkill at 100s–1000s (repo sprawl, Actions noise).
- **One shared VM-sites repo, folder per domain:** one Action, scales, clean
  separation from the static-B2 sites which stay as they are. Preferred, *if*
  `git-adapter` can commit per-domain subpaths in a shared repo (verify).
- **Reuse existing `sites` repo arrangement with a deploy_config switch:** least
  new infra; mixes VM and static deploys in one place.

### Decision 3 — deploy mechanism (explained for sign-off)
- **Per-site-repo Action (light, now):** reuse "commit is deploy"; the repo's
  Action ships to the VM instead of B2. Pros: nothing new in the chassis, fast,
  trusted mechanism. Cons: VM IP/SSH creds live in Actions secrets (spread across
  repos unless a shared repo → one Action); relocating a domain edits the Action
  target rather than a DB field; chassis doesn't "know" about the deploy.
- **Chassis-driven `service-deployer` (heavy, later):** the platform provisions/
  ships/restarts over SSH, tracked in `service_instances`. Central, monitorable,
  integrates with provisioning; more to build. Flagged as the eventual move.
- Coupled to Decision 2: shared VM repo + one per-repo Action is the coherent
  light option.

### Decision 4 — is `needs_vm_deploy` a sibling terminal item to `needs_rerender`?
`needs_rerender` is today's terminal item: assemble pages from components → git
commit → (B2) deploy via Action.
- **If we use the per-repo Action (Decision 3 light):** the terminal build item
  can stay target-agnostic — it just assembles + commits; the repo's Action
  decides VM vs B2. Then a *new terminal item is not the right cut*; the real
  VM-specific work is **one-time** (provision, DNS, install engine, register,
  set `ACCEPT_HOSTS`), better modelled as an earlier one-off item/step, not a
  per-build terminal item.
- **If we go chassis-driven (Decision 3 heavy):** `needs_vm_deploy` as a sibling
  terminal item → a `vm-deploy` handler that SSH-ships + restarts (tracked in
  `service_instances`) is the natural shape.
- So Decision 4 follows from Decision 3; do not add `needs_vm_deploy` purely as a
  mirror of `needs_rerender` under the light path.

---

## Decision log
- 2026-06-10: framing settled (probe = Layer 4 + thin Layer 5; reuse build +
  git→Actions; swap target to VM). Engine = API-only backend; page content owned
  by the chassis. Decisions 1–4 opened for discussion.
- 2026-06-10: Decisions 1–4 resolved (see "Decisions — RESOLVED" above). Confirmed
  current build pipeline is `build-dispatch-loop` (site work items);
  `pageflow-builder` deprecation is a separate later call. `git-adapter` already
  writes per-domain subpaths into a shared repo; B2 Action syncs a domain-named
  first-level path in one bucket — VM path mirrors this in a separate repo.
- 2026-06-10: confirmed static deploy is serverless — B2 Action `b2 sync`s
  `sites/<domain>/` to `portfolio-sites`; a Cloudflare Worker serves by
  `hostname+path` → B2 object. VM path replaces both halves (nginx serves +
  engine captures; DNS → box). Wrote `deploy-to-vm.yml` (rsync over SSH, mirror
  of the B2 Action) + `setup.sh WEBROOT_OWNER` param. Box `setup.sh` is multi-vhost
  and `nginx -t`-clean.

## Risks / watch-list
- Divergence if a separate build path doesn't reuse the same handlers (breaks the
  improvement loop). 
- Credential sprawl with per-repo Actions at scale.
- DNS repoint stops parking revenue — choose domains deliberately.
- Health-adjacent domain names need non-clinical framing.
