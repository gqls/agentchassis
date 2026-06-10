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
- **P0 — agree the structural decisions below.** (current)
- **P1 — manual go-live (Path A)** for 3–5 controlled domains; capture the exact
  box setup as `setup.sh` + nginx conf + systemd unit.
- **P2 — wire the deploy on update** per Decision 3 (keep "commit is deploy").
- **P3 — make the probe a normal pipeline output**: classification/site_type +
  the capture component the planner can include + the terminal/deploy step per
  Decisions 1 & 4.
- **P4 — off-box collection + ranking**: checkpoint JSON, compute events-per-1k,
  rank domains.
- **P5 — registry + relocation** (`service_instances`) and, eventually, the
  chassis `service-deployer` adapter.

---

## Open decisions (under discussion)

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

## Risks / watch-list
- Divergence if a separate build path doesn't reuse the same handlers (breaks the
  improvement loop). 
- Credential sprawl with per-repo Actions at scale.
- DNS repoint stops parking revenue — choose domains deliberately.
- Health-adjacent domain names need non-clinical framing.
