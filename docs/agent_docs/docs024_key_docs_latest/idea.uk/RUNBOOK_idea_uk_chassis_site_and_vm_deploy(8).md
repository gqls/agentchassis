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

## Phase 0 — classifier-only positioning read (cheapest first step, recommended)

Before committing to a full build, run **just `domain-research-classifier`** on idea.uk and read the
specs it writes. This is the minimal, near-zero-cost, zero-deploy way to answer "what does idea.uk do for
a stranger?" — the classifier's output *is* a positioning brief.

**What it writes** (four `site_specs` aspects for the idea.uk site record):
- `identity` — `about_summary` (a 2-3 sentence "what this does"), `services`, `target_audience`,
  `unique_selling_points`, `tagline`, `industry`. The direct answer to the question.
- `classification` — `site_type`, `category`, `industry_tags`, `tone_suggestion`, `suggested_style`,
  `reasoning`.
- `content_direction` — the writing-style guide, including `example_phrases.characteristic` (candidate
  one-liners in the right voice).
- `design_intent` — colour / typography / layout direction.

**How to run just the classifier:**
1. Run `domain-submitter` with `{"domain": "idea.uk"}` — it creates (or finds) the site record and
   returns the `site_id` (it also writes a `submission` spec and a `needs_domain_research` item).
2. Run `domain-research-classifier` on `{"site_id": "<that id>", "domain": "idea.uk"}` — invoke it
   **directly** (its input contract is just `site_id` + `domain`, and handlers run standalone) so it
   stays classifier-only and doesn't chain onward.
3. Read the results: `SELECT aspect, spec_data FROM site_specs WHERE site_id = '<id>'` for the four
   aspects above. (Check the `site_specs` column names first, per the standing rule.)

**Caveats / things to know:**
- **Decision (2026-06-14): leave the live site up** during the run — accept the read being informed by
  the current landing page (the usable option). No suppression; the notes below are kept for reference.
- **Not blank-slate.** It still scrapes the live idea.uk (the Phase 1 caveat), so the read is informed by
  the current landing page. If you want a truly name-only read, you must keep the scrape from reaching the
  live tool — but weigh it first:
  - *Honest warning:* "idea.uk" is a **generic** name. A name-only read will most likely infer a generic
    *ideas / innovation* platform unrelated to the real AI-product-ideas tool. The current landing page is
    the only real description of what idea.uk does, so hiding it removes signal, not bias. For positioning
    you'd actually use, prefer letting it scrape (informed-by-landing-page) or seeding a one-line mission.
  - *If you still want the name-only experiment, do it safely:* idea.uk is live and Stripe posts to
    `/stripe/webhook`, so **do not change DNS and do not stop nginx** (both take the whole site down and
    risk a payment's webhook failing — recoverable via Stripe's ~3-day retries, but needless). Instead add
    a temporary nginx `location = /` that returns an empty page (or 404) while the default `location /`
    keeps proxying everything else to the Go service. The scraper starts at `/` and follows links found
    there; a blank `/` has none, so it gets nothing — while the tool and webhook paths stay up. Revert
    after. (A blank page is marginally better than a 404 — a 404 may be recorded as "site not found".)
    Note `search_domain`'s web search is uncontrollable and may reintroduce a little signal regardless.
- **It writes specs under idea.uk's site record.** Harmless (just data), but check first whether idea.uk
  already exists as a site with specs you'd rather not overwrite.
- **Its terminal step creates a `needs_strategy` work item** — the on-ramp to the rest of the build. If
  the build heartbeat (`build-pipeline-trigger`) is running and that item gets triaged, it will flow into
  a full build on its own. That is harmless (a build deploys to B2, which DNS does not point at — see the
  DNS note), but to stay strictly classifier-only, invoke the classifier out-of-band and/or park that
  `needs_strategy` item.

**Then decide:** if the specs read well, go to Phase 1 (let the full build run and review it). If the
positioning is off (it just parrots the current landing page, or misses the niche), adjust before
building — seed a short `mission_brief`, or arrange a name-only read.

**Phase 0 result (2026-06-14 — ran, with the live site up).** site_id `97ed2f64-65ca-4b67-8a98-dfd8195a0d3a`.
The classifier produced faithful, accurate specs: `identity` (about_summary, tagline "AI product ideas for
your business, tested before we recommend them", target_audience = UK SME owners/founders), `classification`
= **interactive-platform** (category `interactive`, confidence ~0.91 — correctly not a brochure),
`content_direction` (a strong, usable writing-style guide), `design_intent`. The chain then continued
(`strategy` + `briefing` specs written), so a full build is likely now in motion. Findings worth acting on:
- **How the design is ACTUALLY decided (confirmed 2026-06-16 — this corrects an earlier claim).** The look
  is **not** produced by an LLM reading `design_intent`. The active decider is **`site-design-planner`**
  (deterministic, its own description says "no LLM"): it matches a **layout** from a library by industry-tag
  overlap, a **typography** set by font-family match (with fallback), and a **palette** by a cascade
  (`design_reference → mission → design_intent → archetype_default`), then installs the result. For idea.uk
  it resolved to layout **`tool-portal-dark`** (library match, 3 tag overlaps), typography **`sans-modern`**
  (FALLBACK — Fraunces / IBM Plex are not in the library), and palette via **`archetype_default`** (NOT the
  parchment/rust in `design_intent`). So the build did **not** reproduce the current look — it *genericised*
  it. `brand-designer` (5 hardcoded themes on Haiku) and `visual-designer` (image/asset stub) are
  experimental/vestigial, not the decider.
- **Why `design_intent` was ignored (structural).** The classifier writes `design_intent` as **prose** (hex
  buried in a `colour_mood` sentence), but the palette cascade expects **structured** values like adoption's
  `design_reference.reference_values` `{primary, secondary, accent}`. No structured palette to read → it fell
  to `archetype_default`. For a fresh (non-adopted) site the classifier's design intent mostly does not reach
  the installed design. The content was validated and is good; the design is a separate, library-driven
  problem (see the two workstreams below).
- **Duplicate specs.** The classifier was run more than once, so there are two rows each of `identity` /
  `classification` / `content_direction` / `design_intent`. Downstream almost certainly reads the newest per
  aspect, but tidy the stale rows so the planner can't pick up an old one.
- **If pausing to set the design first:** stop the dispatch / park the open work items before
  `build-site-planner` reads the current `design_intent`.

---

## Setting direction before the build (decided 2026-06-14)

Decision: set idea.uk's **direction** (and design) before letting `build-site-planner` run. The Phase 0
read was faithful but backward-looking — it described today's idea.uk and preserved today's design —
because the classifier is a current-state tool and we gave it no aspiration to aim at. The goal for every
domain is to **lead the vertical with the best content for users** (the framework's founding aim), which a
bare classify of the existing site cannot produce. (Does the classifier "get" that ambition? No, not as
run — but it *can* take it from a mission, see below; it is not a classifier defect to fix.)

How aspiration should enter — fix the framework, not hand-seed (refined 2026-06-14):
The aspirational direction has **no generated home** today. `mission_brief` / `roadmap_brief` are the
designated aspirational slots but are **owner-supplied** (nothing generates them); `strategy`
(domain-strategist) is generated and partly forward-looking but monetisation-framed and runs *after* the
classifier, so it cannot shape the classifier's `design_intent`. With the slots empty, the classifier fell
back to the scrape for both content and design — hence the conservative result. Hand-authoring a
`mission_brief` / `design_intent` for idea.uk would work but is the **manual correction we want to avoid**.
- **Preferred fix (framework), refined per the aspect list (2026-06-14):** there is no `vision` /
  `ambition` aspect among the 50, and `aspect` is free-text (no enum) — so rather than add a new aspect,
  make the aspirational direction a **standing default carried in the `mission` / `mission_brief`
  aspect** that every site gets. `domain-submitter` already owns mission persistence and runs *before*
  the classifier, so it is the natural home: always write a `mission_brief` = a fixed platform
  **standing-ambition principle** (lead the field; the most useful / forward-looking content for users;
  build around the site's distinctive tools and surround them with the best material; surpass — don't
  mirror — any existing site), merged with the owner's mission when supplied. The classifier, strategist
  and planner already read `mission_brief` as primary, so forward **content** follows for every site by
  default — correct ordering, no new aspect/agent/reorder. Put the default/merge logic in a Go action,
  not workflow branching (keep the workflow simple).
- **Design is a separate workstream (confirmed — design is deterministic, library-driven).** The
  standing-ambition mission lifts *content* (classifier/strategist/planner read it as primary, via LLM), but
  does **not** touch the design: `site-design-planner` is deterministic and doesn't read mission prose, and
  the palette cascade only takes a palette from `mission` if it is *structured* colours (a text principle has
  none, and we wouldn't want one fixed palette across every site). So design leadership needs its own track:
  (a) fix the prose→structured mismatch so an intended palette actually applies for fresh sites (classifier
  emits a structured palette / `reference_values`, OR the cascade parses hex from prose); (b) curate
  distinctive **layouts**, **typography** sets and **archetype-default palettes** in the libraries (add
  Fraunces / IBM Plex or accept the fallback); (c) decide whether `tool-portal-dark` (what idea.uk got) is a
  leading look or a generic template — needs to be seen rendered.
- **Verify before building (schema / reuse first):** `SELECT DISTINCT aspect FROM site_specs` confirmed —
  no `vision` / `ambition` aspect, and `aspect` is free-text (no enum), so no schema change is needed. The
  design-agent defs are now in hand — `site-design-planner` is the deterministic decider (see above). Still
  to confirm: how `build-site-planner` weights `strategy` vs `mission_brief`, and a reuse-check of
  `domain-submitter`'s `persist_mission_brief` / `write_site_spec` path before writing the Go action.
- **Tidy the duplicate spec rows** regardless, so the planner reads current specs, not stale ones.

Interim (not a framework fix): the one non-"correction" option is the owner supplying idea.uk's real
`mission_brief` as genuine vision — the slot is for exactly that. But the framework fix is the priority.

Open fork (gates the mission): **focused** — the best AI-product-ideas platform (flagship report + audience
check, surrounded by the sharpest content/tools/guides in that niche) — vs **broad** — a frontier
"what's worth building now" platform with AI product ideas as the first flagship among several. Lean:
launch focused, roadmap leaves room to broaden. The existing tools (idea generator, free audience check)
are preserved and launched via the framework regardless; the ambition wraps around them. (Also pin down
the "log handling" tool the user mentioned — planned vs already present.)

Next: confirm where design is decided (design-agent defs) + check whether gamesdesign.co.uk's specs read
ambitious or generic → reuse-check how `domain-submitter` writes `mission_brief` (the `persist_mission_brief`
step / `write_site_spec`) → add the standing-ambition default/merge in a Go action (no schema change; aspect
is free-text) → re-run the front of the pipeline → tidy duplicate specs → planner → VM deploy (Phase 2).

---

## Phase 1 — submit idea.uk to the framework, build to a STAGING target (zero risk to the live tool)

Goal: get the framework's positioning and a built site, reviewed, **without touching the live earner**.

1. **Enter through `domain-submitter` — not the adoption agent, and not `intake-orchestrator`.**
   - `site-adoption-agent` *crawls an existing URL and recreates it* (its archetype step classifies "what
     this site IS, not what it should become"); pointed at live idea.uk it reproduces the current tool
     landing/flow — the *opposite* of a fresh read (that's the seeded approach, below).
   - `intake-orchestrator` is the older, human-in-the-loop path: it uses a *different* classifier
     (`site-classifier`) and a confirm step whose site-type list is dated (`landing/content/portfolio/
     brochure` — no `tools` / `interactive-platform`), then spawns a builder directly. Skip it here.
   - `domain-submitter` feeds the **current** classifier, `domain-research-classifier`, which reads the
     live layout taxonomy, has the adoption-awareness baked in, and whose output flows through the
     work-item build pipeline to `build-site-planner` (the shared convergence point). Submit just
     `{"domain": "idea.uk"}`, with none of its optional `objective` / `mission` / `mission_brief` specs.
   - **Decision: fresh, chosen.** (The seeded build comes later — see "Seeding the existing setup".)
   - **Confirmed caveat — a fresh submit of idea.uk is NOT blank-slate.** `domain-research-classifier`
     web-searches the domain *and* scrapes the live site (up to 3 pages, following about/services/contact/
     team/pricing); it only skips that scrape when an adoption has already run (a `site_archetype` spec is
     present). idea.uk has a live site, so the classifier **will scrape the current tool landing page** and
     the read will substantially reflect/restate today's £29-report positioning rather than invent from
     nothing. Still useful (it re-derives and restructures the current positioning), but if you want a
     truly name-only read you'd have to keep the scrape from seeing the live tool — otherwise accept
     "fresh, informed by the current landing page".
2. **Let the work-item pipeline run.** The chain is: `domain-research-classifier` (writes identity,
   classification, content_direction, design_intent; creates `needs_strategy`) → `domain-strategist` →
   `build-site-planner` (plans the site, syncs pages, populates nav, reconciles → emits `needs_page` × N
   plus a terminal `needs_rerender`) → content / design / imagery handlers → `needs_rerender` assembles
   and commits → static deploy. The `build-pipeline-trigger` heartbeat is what picks up a site with
   pending work items and drives the dispatch loop. Verify with the standard checks (`site_work_items`,
   `page_components`, `site_components`, `pages.build_status`) and the git-adapter logs.
3. **Deploy is the framework's normal static deploy to B2 — already safe.** Because Cloudflare points
   idea.uk at the VM, the B2 copy is invisible at `https://idea.uk`; review it at its B2 URL. No special
   preview subdomain is required, and the live tool is untouched.
4. **Review** the positioning and pages — the deliverable of Phase 1 and the answer to the stranger
   question. Iterate through the framework until the front site is right, before any VM work.

---

## Does the fresh path need adoption's machinery? (capability map)

You asked for "everything we've done in adoption, minus the adoption, converging soon after." Mapping the
adoption agent's steps against the fresh path shows most of it is already there, and the rest is
inherently about crawling an existing site:

| Adoption (`site-adoption-agent`) produces | Fresh path equivalent | Verdict |
|---|---|---|
| Identity (from crawl) | `domain-research-classifier` identity (from search + live-site scrape) | already in fresh |
| `site_archetype` (rich "what it is") | classifier `classification` (site_type / category / tags) | overlapping; `build-site-planner` reads `classification`, so covered |
| `content_direction` (writing style, from real pages) | classifier `content_direction` (from scraped text) | already in fresh (less grounded, but present) |
| `design_intent` (grounded in extracted CSS) | classifier `design_intent` (LLM-inferred) | already in fresh — and for a *fresh* read, inferred is what we want, not the old site's CSS |
| Design **fingerprint** from real CSS | — | crawl-only; N/A to a fresh read (no existing design we want to keep) |
| **Interactive-feature** detection | — | crawl-only; matters for the *seeded* build (detecting the tool), not fresh |
| Pages created + work items | `build-site-planner` `sync_pages` + `reconcile` → `needs_page` + `needs_rerender` | already in fresh (just later in the pipeline) |
| Nav populated | `build-site-planner` `populate_nav` | already in fresh |
| Existing-page **convergence** | `build-site-planner` "preserve existing pages exactly" logic (+ `adoption_locked`) | already shared — converges onto adopted *or* previously-built pages |

**Finding.** The convergence point — `build-site-planner` — is already shared by both paths and already
adoption-aware. The fresh classifier already emits the same rich spec aspects (identity, classification,
content_direction, design_intent) the build consumes. The only adoption capabilities the fresh path lacks
(CSS fingerprint, interactive detection, the full archetype object) are produced *by crawling an existing
site*, so they can't be "ported without the adoption" — they don't apply to a blank read.

**So a new adoption-derived fresh workflow is probably not needed for the fresh read** — running the
existing fresh path reuses all of it. A new self-contained "fresh-build" agent (a copy of adoption with
the crawl swapped for research, plus an orchestrator wrapper) is only worth building if, after seeing the
fresh output, you want adoption's *single-pass, one-agent convergence* instead of the multi-agent
work-item pipeline. That would be a real rework, not a copy: adoption's downstream steps (`analyze_site`,
`classify_archetype`, `derive_content_direction`, `generate_design_intent`) all consume the **crawl**
output, so removing the crawl means re-feeding those steps from research — and the richest of them
(fingerprint, interactive, representative-content) have no fresh equivalent to consume.

**Where adoption's full richness *is* available to idea.uk: the seeded build.** When we want the tool
detected, the real design captured, and fast convergence, that is exactly adoption pointed at the live
idea.uk (next section) — i.e. by *using* adoption, which is fine for the shippable build; "apart from the
adoption" only governs the fresh read.

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

## Open questions — what's answered, and what remains
From the workflow definitions (classifier, planner, trigger, submitter, intake), most earlier questions
are now answered:
- **What `domain-research-classifier` does — answered.** It web-searches the domain *and* scrapes the
  live site (3 pages), skipping the scrape only when an adoption has already run. So the fresh read of
  idea.uk will ingest the current landing page (Phase 1 caveat). It writes identity / classification /
  content_direction / design_intent and creates `needs_strategy`.
- **Entry agent — answered.** Use `domain-submitter` (feeds the current classifier and the work-item
  pipeline to `build-site-planner`). `intake-orchestrator` is the older HITL path on a different, dated
  classifier — not this.
- **`build_approach` / `hosting_trajectory` — note.** These did *not* appear in the classifier or planner
  definitions shown (the classifier emits `site_type` / `category` / `industry_tags`; the planner emits
  `site_type`). They may live in `domain-strategist`, or be a separate concept from the architecture doc.
  Not needed for the fresh read regardless; worth pinning down before the shippable build.

**Resolved empirically (2026-06-14 idea.uk run):** a freshly submitted domain DOES flow through dispatch
end to end without manual triage — the idea.uk run produced not just the classifier's four specs but also
a `strategy` spec (`domain-strategist` ran) and a `briefing` spec (`build-briefing-agent` ran). So the
fresh front segment flows; the work-item status/triage default is whatever makes created items
dispatchable, and it works for the fresh entry. (The chain having reached `briefing` also means the next
handler is `build-site-planner` — a full build is likely in motion unless dispatch is paused.)

**Resolved by the latest definitions:**
- **`domain-strategist`** — reads the specs, writes a `strategy` aspect (domain_type, revenue model,
  site_type, recommended_page_types, tone, value_proposition), and creates a `needs_briefing` item for
  `build-briefing-agent`. So the chain continues: classifier → `needs_strategy` → strategist →
  `needs_briefing` → briefing → (planner → pages → rerender → deploy).
- **`build-dispatch-loop`** — loads up to 5 dispatchable items for a site, and per item: claims it
  atomically, spawns the item's `handler_agent` (dynamic type), calls it, marks complete/failed; the
  trigger re-fires for the rest. This is the same mechanism adoption's work items flow through.

**The decision for you.** Given the capability map, my recommendation is **reuse: run the existing fresh
path** (`domain-submitter` → classifier → strategist → `build-site-planner` → …) and see whether it
converges to adoption-quality output — build a new adoption-derived self-contained "fresh-build" agent +
orchestrator only if that output is materially weaker or the multi-agent convergence proves unreliable,
not pre-emptively. If you'd rather I design that new workflow now regardless, I can: start from
`site-adoption-agent`, drop `crawl_site` plus the fingerprint / interactive / representative-content
steps, and feed `analyze_site` / `derive_content_direction` / `generate_design_intent` from `web_search`
(plus an optional live scrape), with a `fresh-build-orchestrator` wrapper mirroring
`site-adoption-orchestrator`.

## Schema / SQL note
The classifier/briefing fields (`build_approach`, `hosting_trajectory`) and the site / work-item tables
live in `clients_db`. Per the standing rule, check the live schema before writing any SQL for intake or a
deploy-target field — none is written here yet; this is the plan.

---

## Empty index — diagnosis (2026-06-16)

idea.uk's chassis build ran end to end (research → strategy → briefing → site_plan → composition → design →
policy pages + imagery + rerenders), but the **index page has 0 components** and `needs_page:index` ended
**`failed`** ("Claim timed out (attempts exhausted)", 3 attempts). The plan is fine — it lives in the
`site_plans` table (`build-site-planner`, `is_current=t`), not as a spec; the absent `site_plan` *aspect* was
a red herring.

**Most likely cause — the confirmed framework bug in `NOTES_gamesdesign_silent_norebuild`.** The
`page-content-writer`'s `complete` step declares `output_field` (singular) where the coordinator reads
`output_fields` (plural), so the writer's compiled page is dropped at the coordinator boundary (it falls back
to dumping working state → exceeds the 900k `MaxResultSizeBytes` cap → returns a `status:"completed"` stub),
the parent's `save_page_sections` reads null, and 0 components are written. Framework-wide — any multi-section
page on the 2026-06-13 writer `v2` def — and idea.uk's index is multi-section.

**Why idea.uk `failed` where gamesdesign `false-completed`:** the evidence-gate
(`migration_claimed_item_timeout_evidence_v2`, live since 2026-06-04) refuses to complete a 0-component page,
so the claim is reset and retried until attempts exhaust → `failed`. That is the gate working as intended.
**Not yet confirmed for idea.uk** — the decisive read is whether the index's child `page-content-writer` run
produced a full page that the parent then received as the stub
(`page_content.response.message = "...Full result exceeded size limit."`). The alternative is a genuine
handler hang. Don't conflate the two until the orchestration (or a live retry) shows which.

**Separate content gaps on the index (non-blocking deferrals — NOT why it's empty):**
- `needs_section_data` (pricing section): missing `tier_1_name/price/features` sourced from
  `site_specs.pricing.tiers[0].*`; idea.uk has no `pricing` spec. This is a **human** field, so the
  built-but-unwired section-data reconciler would not resolve it — either capture the £29 pricing into
  `site_specs.pricing`, or the pricing section shouldn't be on the page.
- `unresolved_cta` ×2 (hero + call-to-action): "no real page exists to serve as this CTA's destination (no
  eligible content hub)"; the gated template renders no button. idea.uk has only `index` + the 3 policy pages
  — no hub/tool/guide pages for the CTAs to point at. Ties to the thin plan + the known hero-CTA weakness.

**The thin plan.** Only four pages were produced (`index` + `privacy` + `terms` + `refund-policy`). So even
once the empty-page bug is fixed, the index's CTAs and any list sections have nothing real to point at until
the plan includes hub/tool/guide pages — which loops back to the direction/ambition work.

**Fix directions (priority first):**
1. **Coordinator output-drop (framework — fixes every site). Attribution NOT concluded (2026-06-17).** Two
   non-exclusive parts: the coordinator reading `output_fields` (plural) **only** (reportedly since ≈14 Jan)
   is the latent **trap** — and note `output_field` (singular) is the *documented* step-output standard
   (001), with `output_fields` / `extractWorkflowResult` / the 900k cap undocumented here (the result/envelope
   contract was moved to `035_adapter_guide.md` §1, not in hand) — so writers follow `output_field` and the
   coordinator silently diverges; and the `page-content-writer`'s `complete` step emitting `output_field` is
   the **trigger**. The 06-06 gamesdesign build worked with the parent receiving the *flattened* compile
   output, which predates "14 Jan as sole cause" — so something changed between 06-06 and 06-15, the only
   dated candidate being the writer's **06-13** `complete` update. Clean reconciliation: coordinator set the
   trap (≈14 Jan), the 06-13 writer change sprang it. **Decisive reads to conclude:** the writer def *history*
   (is there a pre-06-13 `complete` returning the flattened output? → confirms the trigger, and whether to
   *restore* a contract or *author* one); the coordinator's `output_fields`-only change date; the 06-06
   `page_content.response` shape (flattened keys vs a state-dump that fit). **Fix (lean structural,
   coordinator-side):** make the coordinator honour the documented `output_field` (back-compat), and fail
   loudly / pass by reference on oversize — never a `status:"completed"` stub (fixes the class; image-build-
   handler hit a sibling). Correcting the writer's `complete` to the current contract is the targeted unblock,
   not the fix. Do NOT raise `MaxResultSizeBytes`; do NOT use `output_fields:["page_content"]` (nests wrongly).
2. **Content gaps (idea.uk-specific, after the build works).** Pricing into specs (or drop the section); CTA
   destinations need real hub pages (the plan/direction work).

**Next:** confirm idea.uk's index hit the output-drop (its child writer run + the stub signature) vs a genuine
hang → then recover the writer's pre-06-13 `output_fields` → apply + harden → requeue the index → address the
content gaps.

**Update (2026-06-17) — bucket audit confirms size-is-the-trigger.** Across all agents, `complete_workflow`
steps split: **100 steps / 76 agents** declare `output_fields` (explicit return — safe); **61 bare + 4
`output_field`-only ≈ 59 agents** take the state-dump fallback and almost all run fine (their dumped state is
< 900k). `page-content-writer` is 1 of only 4 in the `output_field`-only bucket (with `site-planner`,
`thunder-reaper`, `training-data-preparer`) and is the one that breaks — a multi-section page is the result
that clears the cap. So the singular key is necessary-but-not-sufficient; the real trigger is **result size**.
Fix is two layers: (1) **coordinator-side class fix** — oversize → fail loud / pass by reference, never a
`status:"completed"` stub (protects all ~59 dump-bucket agents); (2) **writer-side unblock** — give
`page-content-writer.complete` an explicit `output_fields` returning just the compiled page, flat
(`page_content.response.sections_metadata`); the exact value depends on what `compile_page` stores (next read).
NOTE: the child writer runs' `final_result` is null — the stub/real A/B lives in the *parent's*
`collected_data.page_content.response`, not the child's `final_result`.
