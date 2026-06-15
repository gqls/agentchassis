# RUNBOOK — idea.uk front site built on the chassis, deployed to the VM (keeping the tool)

Status: **PLAN** (not yet built). Created 2026-06-14.

This is deliberately **separate** from `RUNBOOK_idea_uk.md`. That runbook stays the operational doc for
the live £29 report tool as it runs today. This one covers a new piece of work: using the agentchassis
framework to build idea.uk's **front site** (and, in doing so, work out its positioning), and deploying
that site to the **existing idea.uk VM** alongside the live tool — not to Backblaze B2.

---

## The challenge in one line

idea.uk is not a blank brochure site: it already has a live, earning tool (the £29 report service on
the VM, behind nginx). The framework's normal output is a static site where **commit is deploy** —
GitHub Actions writes the build to Backblaze B2, and Cloudflare/DNS points at B2. If we deployed idea.uk
that way, the tool would be gone. So we build the static front site with the framework but deploy it to
the **VM**, where nginx serves the static pages and routes the tool's own paths to the Go service. The
tool keeps running unchanged; the framework owns the front site.

## How this maps onto what the framework already has (reuse, don't invent)

- **`build_approach: hybrid`** and **`hosting_trajectory: needs_server`** (or `static_now_api_later`) —
  the site classifier already emits these. idea.uk is exactly a hybrid / needs_server case: static
  pages plus a dynamic tool. Use those values; don't add a new mechanism.
- **"Commit is deploy" → GitHub Actions → a target.** The framework commits the built site to git and an
  Action deploys it. For idea.uk the Action's target is the **VM** (rsync over SSH) instead of B2. This
  is "another adapter" in the sense the system already anticipates (git → Action → host).
- **`deployer-agent`** is the deploy actor (git commit + deploy). The VM is just a different destination
  for idea.uk's repo; the rest of the pipeline is unchanged.
- **DNS:** idea.uk's DNS is at Hetzner and already points at the VM. Unlike a B2 site (which needs DNS
  pointed at B2), idea.uk needs **no DNS change** — the VM is already the address. One fewer moving part,
  and it means the cutover is entirely an nginx change we control.

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

1. **Intake.** Submit the domain `idea.uk` through the normal new-site entry (intake orchestrator).
   - **Decision — how much to tell it (this changes what the site says idea.uk is):**
     - *Fresh read* — submit the bare domain and let research + the classifier infer what idea.uk
       is/should be. This is the literal "what would an outsider think idea.uk does"; cheap and direct.
       It may not match our actual tool, especially as idea.uk has little web presence yet.
     - *Seeded build* — give the briefing a short, true statement of the offering: verified AI product
       ideas, a £29 report, and that the tool lives at `/request`. The framework then researches the
       vertical and positions the **real** product, with the front-site CTA pointing at the tool.
     - Recommended: do the **fresh read first** (it is the literal answer to the question), then a
       **seeded build** for the site we would actually ship.
   - Confirm the classifier lands on `build_approach: hybrid`, `hosting_trajectory: needs_server`. If it
     classifies idea.uk as a pure static brochure, the briefing seed (the tool at `/request`) is the
     nudge that corrects it.
2. **Let the pipeline run** the normal chain: domain research → briefing → site plan → logo / hero →
   design → content pages → `needs_rerender` (the terminal assemble-and-commit step). Verify with the
   standard checks: `site_work_items`, `page_components`, `site_components`, `pages.build_status`, and the
   git-adapter logs (`kubectl -n ai-persona-system logs -l app=git-adapter | grep idea.uk`).
3. **Deploy the build to STAGING, not the VM root.** Use the framework's normal B2 path to a staging
   bucket/subpath, or a staging subdomain (e.g. `preview.idea.uk`) — anywhere that is **not** the live
   tool. Review it there.
4. **Review** the positioning and pages. This is the deliverable of Phase 1 and the answer to the
   stranger question. Iterate through the framework (regenerate sections/pages) until the front site is
   right, before going anywhere near the VM.

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

### Deploy path: framework → VM (instead of B2)
- Keep "commit is deploy": the build is committed to idea.uk's site repo as normal.
- Add to idea.uk's repo a **GitHub Actions workflow that rsyncs the built site to the VM**
  (`/var/www/idea.uk`) over SSH on push — instead of writing to B2.
- Needs: an SSH deploy key the Action can use (added to the VM's `authorized_keys`), the target path, and
  no nginx reload for content-only changes (static files).
- This is a **per-site** deploy: idea.uk's repo → VM; other sites' repos → B2 as before. The divergence
  is one workflow file in idea.uk's repo — document it so it isn't "corrected" back to B2 by mistake.

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
- [ ] Deploy: a GitHub Actions workflow in idea.uk's repo that rsyncs the build to the VM over SSH; an
      SSH deploy key on the VM.
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

## Schema / SQL note
The classifier/briefing fields (`build_approach`, `hosting_trajectory`) and the site / work-item tables
live in `clients_db`. Per the standing rule, check the live schema before writing any SQL for intake or a
deploy-target field — none is written here yet; this is the plan.
