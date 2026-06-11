# Traffic-probe — plan

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
- **P2 — wire the deploy on update**: VM-sites repo + its VM-deploy Action
  ("commit is deploy", target = VM box). Static repo/Action untouched.
  *Done: content Action `deploy-to-vm.yml` + engine Action `deploy-engine-to-vm.yml`
  (+ `setup.sh` `DEPLOY_USER` hook), both validated. New-domain onboarding
  documented (runbook §4).*
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
- **P4 — off-box collection + ranking**: checkpoint JSON, compute events-per-1k,
  rank domains. *If collection runs as a chassis adapter, it MUST follow the
  Adapter Response Envelope Contract (typed-struct bool headers, reuse request_id,
  message_id, ProduceWithValidation).*
- **P5 — registry + relocation** (`service_instances`) and, eventually, the
  chassis `service-deployer` adapter. *Same adapter-envelope contract applies.*

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
