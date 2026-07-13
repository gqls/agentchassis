# RUNBOOK — idea.uk front site built on the chassis, deployed to the VM (keeping the tool)

Status: **PLAN** (not yet built). Created 2026-06-14.

This is deliberately **separate** from `RUNBOOK_idea_uk.md`. That runbook stays the operational doc for
the live £29 report tool as it runs today. This one covers a new piece of work: using the agentchassis
framework to build idea.uk's **front site** (and, in doing so, work out its positioning), and deploying
that site to the **existing idea.uk VM** alongside the live tool — not to Backblaze B2.

---

## The challenge in one line

idea.uk is not a blank brochure site: it already has a live, earning tool (the £29 report service on
the VM, behind nginx). The framework's normal way to *serve* a finished site is static hosting on
Backblaze B2 with DNS pointed at B2. If we served idea.uk that way, DNS would point away from the VM and
the tool would be bypassed. So we keep idea.uk's DNS on the **VM** and have the VM serve both: nginx
serves the framework's static pages and routes the tool's own paths to the Go service. The tool keeps
running unchanged; the framework owns the front site. (Building and deploying the static files to B2 is
still fine and harmless — see the DNS note — it just isn't what idea.uk's DNS points at.)

## How this maps onto what the framework already has (reuse, don't invent)

- **`build_approach: hybrid`** and **`hosting_trajectory: needs_server`** (or `static_now_api_later`) —
  the site classifier already emits these. idea.uk is exactly a hybrid / needs_server case: static
  pages plus a dynamic tool. Use those values; don't add a new mechanism.
- **"Commit is deploy" → GitHub Actions → a target.** The framework commits the built site to git and an
  Action deploys it. For idea.uk the Action's target is the **VM** (rsync over SSH) instead of B2. This
  is "another adapter" in the sense the system already anticipates (git → Action → host).
- **`deployer-agent` / `site-deployer` / `site-publisher`** are the deploy actors (commit + deploy to
  B2). For idea.uk the build still commits and deploys to B2 as normal; the VM gets the same output by a
  separate sync (Phase 2). The pipeline itself is unchanged.
- **DNS (Cloudflare → the VM).** idea.uk's DNS is in Cloudflare and points at the VM. Useful
  consequence: a **normal static deploy of idea.uk to B2 is harmless** — the files land in B2, but
  because DNS points at the VM, nothing at `https://idea.uk` changes and the live tool is unaffected. So
  the framework's ordinary static deploy doubles as a safe staging copy (review it at its B2 URL), and
  the eventual cutover to serving the framework site at idea.uk `/` is then **only an nginx change on the
  VM** — no DNS change — which also makes rollback trivial.
  (An earlier draft of this runbook said the DNS was at Hetzner; treat Cloudflare → VM as current, and
  reconcile if any Hetzner records linger.)

## What this answers, and what it replaces

1. **"What does idea.uk do for a stranger who's just landed on it?"** Running idea.uk through intake →
   research → briefing → site plan makes the framework form, and write, a positioning. That output *is*
   the answer to the question, and the new front copy.
2. **It replaces the hand-written landing page.** We were about to rewrite the embedded `page.html` by
   hand (cold-visitor copy + a professional design). Instead the framework builds the whole front site,
   so that manual rewrite is no longer the plan — the framework build is idea.uk's front site.

---

## Phase 1 — submit idea.uk to the framework, build to a STAGING target (zero risk to the live tool)

Goal: get the framework's positioning and a built site, reviewed, **without touching the live earner**.

1. **Use the fresh submit path — NOT the adoption agent.** This matters: `site-adoption-agent` *crawls
   an existing URL and recreates it* (its archetype step even classifies "what this site IS, not what it
   should become"). Pointed at the live idea.uk it would reproduce the current tool landing/flow — the
   *opposite* of a fresh read (it is actually the seeded-with-existing approach, below). For a fresh read,
   enter through `domain-submitter` (or `intake-orchestrator` — see open questions): submit just
   `{"domain": "idea.uk"}`, with none of its optional `objective` / `mission` / `mission_brief` specs.
   `domain-submitter` creates a `needs_domain_research` work item handled by `domain-research-classifier`,
   the start of the fresh research → classify → build chain.
   - **Decision: fresh, chosen.** (The seeded build comes later — see "Seeding the existing setup".)
   - **Honest caveat about a fresh read of idea.uk.** idea.uk has little independent web presence, so a
     fresh read is largely *name-based* inference ("idea.uk" → ideas). And if `domain-research-classifier`
     fetches the live site during research, it will see the current tool landing page and may simply
     describe the £29 report — so the read may not be truly "fresh" either way. What the experiment tells
     us depends on what the classifier does (open question below) — worth knowing before reading too much
     into the output.
2. **Let the pipeline run** the normal chain: domain research → briefing → site plan → logo / hero →
   design → content pages → `needs_rerender` (the terminal assemble-and-commit step). Verify with the
   standard checks: `site_work_items`, `page_components`, `site_components`, `pages.build_status`, and the
   git-adapter logs (`kubectl -n ai-persona-system logs -l app=git-adapter | grep idea.uk`).
3. **Deploy is the framework's normal static deploy to B2 — already safe.** Because Cloudflare points
   idea.uk at the VM, the B2 copy is invisible at `https://idea.uk`; review it at its B2 URL. No special
   preview subdomain is required, and the live tool is untouched.
4. **Review** the positioning and pages — the deliverable of Phase 1 and the answer to the stranger
   question. Iterate through the framework until the front site is right, before any VM work.

---

## Seeding the existing setup (for the later, shippable build)

The fresh read above is deliberately unseeded. When we instead want the framework to build a site that
*knows about the existing tool and setup*, "seed" splits into three things — only the first two are build
inputs:

1. **Positioning / content seed (what idea.uk offers).** Two mechanisms already exist:
   - `domain-submitter`'s optional specs — pass `objective` and/or `mission_brief` (also `mission`,
     `roadmap_brief`) with a true statement of the offering (verified AI product ideas, a £29 report, the
     tool entry at `/request`). These are written to `site_specs` and read by the classifier, planner and
     content agents. The lightest seed.
   - **`site-adoption-agent` pointed at the live idea.uk** (`target_url = https://idea.uk`,
     `destination_domain = idea.uk`). This is the *richest* "seed with the existing setup": it crawls the
     current site, extracts a design fingerprint + a writing-style guide, classifies what the site is,
     and — importantly — lists `interactive_features` (it will detect the report form/tool), then writes
     specs + pages + work items to recreate it. In effect adoption *is* the seed-with-the-existing-front
     path, because the tool's web surface is exactly what it captures. (Its `destination_domain` override
     also lets us adopt an *exemplar* vertical site into idea.uk if we ever want best-of-breed seeding.)
2. **Tool awareness (don't rebuild the engine).** The engine + Stripe flow are an existing backend
   service, not something the framework should regenerate. The framework has `tool-*` agents and a tool
   library; the report tool should be represented as an **existing** interactive feature the site links
   to (so `tool-suggester` / `tool-generator` don't try to recreate it), and its paths reserved (Phase 2).
   Registering it fully in the tool library is the larger future job in the risks.
3. **The nginx / VM / deploy "setup" is not a build input.** It is the *deploy target* and routing,
   handled in Phase 2 (VM sync + nginx reserving the tool paths). "Seeding the nginx setup" really means
   configuring that target and reserving paths — there is nothing for the content build to consume there.

Recommended seeded approach when we get to it: **adoption against the live idea.uk** (captures the real
pages, design, and the tool as an interactive feature), optionally plus a short `mission_brief`, then the
Phase 2 VM deploy — so everything the framework knows is grounded in what's actually there.

---

## Phase 2 — deploy to the VM, keep the tool (the cutover)

Goal: serve the framework front site at idea.uk `/` while the tool keeps working at its own paths.

### VM layout
- Static front site: `/var/www/idea.uk` (the framework build, synced here).
- Tool: the existing systemd `idea` service on `127.0.0.1:8080`, binary unchanged. Its embedded
  `page.html` simply stops being what's served at `/`.
- nginx: the router between the two. TLS (Let's Encrypt) stays as-is.

### nginx routing — the heart of it
These are the Go service's **actual** registered routes (from `service.go`). Every one must be proxied
to the tool; everything else is the static site.

- **Reserved tool paths → `proxy_pass http://127.0.0.1:8080`:**
  `/request`, `/audience-check`, `/subscribe`, `/confirm`, `/approve`, `/decline`, `/op`,
  `/stripe/webhook`, `/order/` (success + cancel), `/internal/`, `/health`, `/capacity`, and — see the
  decision below — `/terms`, `/refund-policy`, `/privacy`.
- **Everything else → static** from `/var/www/idea.uk` (`/`, `/about`, `/how-it-works`, `/blog/...`,
  etc.).
- **Protect the money and operator paths.** `/stripe/webhook` (payments) and `/op`, `/confirm`,
  `/approve`, `/decline` (operator approvals) MUST reach the Go service. A routing typo that serves these
  as static silently breaks payments or approvals — test each one explicitly after the change.
- **Reserve these paths in the framework build** so it never generates a page that shadows one of them
  (e.g. it must not create `/request` or anything under `/stripe`).

### Policy pages — a decision
The tool serves `/terms`, `/refund-policy`, `/privacy` today and its emails/flow link to them. Either
keep them on the tool (reserve the paths — simplest, no broken links), or let the framework generate them
and repoint the tool's links. Default: keep them on the tool for now.

### The front-site CTA
The front site's primary call-to-action ("get your report" / "start") links to `/request`, the tool's
entry. Make sure the build uses that exact path.

### Deploy path: framework → VM (getting the same build onto the VM)
The framework does **not** give idea.uk its own repo: each site's static build is committed as a
**subdirectory of the one main site repo** (a monorepo), and that repo's GitHub Actions writes changed
files to B2. We are **not** treating idea.uk as a different repo — it stays a normal subdirectory and
deploys to B2 like every other site (which, per the DNS note, is harmless). The only extra is getting
that same build onto the VM. Two ways, lowest-divergence first:
- **(A) VM pulls the build (recommended).** The VM syncs idea.uk's built files from where the monorepo
  already publishes them (its B2 path, or a `git` checkout of just the `idea.uk/` subdirectory) into
  `/var/www/idea.uk` — a cron / systemd-timer `rsync` or `git pull`, or a small webhook. The monorepo's
  deploy is **untouched**; the VM is just one more consumer of the same output, and nothing about other
  sites changes.
- **(B) The monorepo Action also pushes to the VM.** Add a path-conditional step: when the changed files
  are under `idea.uk/`, also `rsync` them to the VM over SSH (needs a deploy key in the VM's
  `authorized_keys`). More moving parts in the shared Action; only worth it if pull-based lag matters.
- Either way there is **no per-site repo** and no fork of the build model — the VM difference lives
  entirely in *serving* (this sync + nginx), not in the repo. Static files need no nginx reload.

### Order of the cutover (low-risk)
1. Put the reviewed static build in `/var/www/idea.uk`.
2. Write the nginx static-root + tool-path-proxy config and validate it (`nginx -t`); test on a staging
   `server_name` or port first, not by editing the live block in place.
3. Swap idea.uk's `/` from the Go tool to the static site; keep the reserved tool paths proxied.
4. **Test the live tool end to end on the new config:** `/request` → operator `/op` approve → pay-link →
   a test-mode (or small real) payment → `/stripe/webhook` returns 200 → report delivered by email; and
   the `/terms` / `/privacy` links in the report emails resolve.
5. Wire GitHub Actions → VM rsync so future framework changes redeploy automatically.

### Rollback
Keep the previous nginx config. Reverting `/` back to the Go tool's embedded page is a one-line `server`
block change + `nginx -s reload`. The tool binary is untouched throughout, so rollback is nginx-only.

---

## What we need to build / configure (checklist)
- [ ] Framework: confirm idea.uk classifies `hybrid` / `needs_server`; choose fresh vs seeded; run the
      build; review on staging.
- [ ] VM: create `/var/www/idea.uk`; write the nginx static-root + reserved-tool-path proxy; bind the Go
      service to `127.0.0.1:8080` (hardening — it currently binds all interfaces via `:8080`).
- [ ] Deploy: get idea.uk's built subdirectory onto the VM (`/var/www/idea.uk`) — pull-based sync from
      B2/git on the VM (A), or a path-conditional push in the monorepo Action (B). No separate repo.
- [ ] Reserve the tool paths in the framework build so it can't shadow them.
- [ ] Decide policy-page ownership (tool vs framework).
- [ ] Test the full tool flow on the new config — before and after wiring auto-deploy.

## Open decisions / risks (honest)
- **Fresh vs seeded intake** changes what the site says idea.uk is. Fresh answers the stranger question;
  seeded ships the real product. Recommended: both, in that order.
- **Build quality is unknown until we run it** — hence staging-first in Phase 1. Do not build straight
  onto the live VM root.
- **Per-site deploy divergence** (idea.uk → VM, others → B2): a single different workflow; document it.
- **nginx path safety:** the Stripe webhook and operator paths must never be served as static. Reserve
  and test them.
- **The report tool stays a standalone Go service** for now. Folding it into the framework's tool-library
  (so the framework "owns" it too) is a much larger job — out of scope here, captured for later.

## Open questions — to confirm before running the fresh build
These decide whether the fresh path still runs end-to-end and whether `domain-submitter` needs revising.
Answerable by sharing the relevant workflow definitions:
- **What does `domain-research-classifier` actually do?** Web-search the vertical only, or also fetch the
  live idea.uk? This decides how "fresh" the read is (Phase 1 caveat), and it is the likely place
  `build_approach` / `hosting_trajectory` get set — confirm it sets them.
- **`domain-submitter` vs `intake-orchestrator` — which is the current fresh entry?** `domain-submitter`
  looks current (active, updated 2026-06-13) and simply creates the `needs_domain_research` item, so it
  should still kick off a fresh run. `intake-orchestrator` also exists and may supersede it — need its
  definition to choose.
- **Does research → classify → build → static-deploy still flow** from a plain `needs_domain_research`?
  Recent focus has been the adoption path, so the fresh path may have drifted. To confirm, the
  definitions of `domain-research-classifier`, `build-pipeline-trigger` (and/or `build-site-planner`),
  and `intake-orchestrator` would let me say whether we run `domain-submitter` as-is or need a revised
  submitter.
- **Do we need `hybrid` / `needs_server` for the fresh run?** Probably not — for the throwaway fresh
  experiment, letting it classify and deploy to B2 as a normal static site is fine. That distinction
  matters for the shippable VM build, not the fresh read.


The classifier/briefing fields (`build_approach`, `hosting_trajectory`) and the site / work-item tables
live in `clients_db`. Per the standing rule, check the live schema before writing any SQL for intake or a
deploy-target field — none is written here yet; this is the plan.
