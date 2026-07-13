# RUNBOOK — idea.uk front site built on the chassis, deployed to the VM (keeping the tool)

Status: **PLAN** (not yet built). Created 2026-06-14.

**What this is, and the task it solves.** The chassis (agentchassis) is the multi-agent system we are building to take a bare domain name and produce a complete multipage website on its own — choosing the design, writing the content, and adding tools and features that suit the industry and the site's objectives, rather than pouring everything into a generic template. This runbook is where that system meets a real, live site. idea.uk already earns money: it sells £29 AI-product-ideas reports through a small Go tool running on its own VM. The task here is to rebuild idea.uk's *front* site through the chassis — so the homepage and the pages around it are properly positioned and designed for what idea.uk actually is — and to deploy that front site onto the same VM, in front of the tool, without disturbing the service that is taking payments (which is why idea.uk's DNS stays on the VM rather than moving to Backblaze B2 the way a normal chassis site would). Because idea.uk is a real domain going through the full pipeline, it doubles as the test case that exposes where the framework falls short — an empty index, a design intent that never reached the build, a layout that doesn't fit the brief — and the working rule throughout is to fix those defects in the framework itself, so every future site benefits, rather than hand-patching this one site.

This is deliberately **separate** from `RUNBOOK_idea_uk.md`. That runbook stays the operational doc for
the live £29 report tool as it runs today. This one covers a new piece of work: using the agentchassis
framework to build idea.uk's **front site** (and, in doing so, work out its positioning), and deploying
that site to the **existing idea.uk VM** alongside the live tool — not to Backblaze B2.

---

## Open items checklist (as of 2026-06-20)

Working backlog. `[x]` done, `[ ]` open. Done so far: the coordinator result-extraction fix, and the classifier structured-`design_intent` migration (applied 2026-06-20). Everything else open.

- [x] **Coordinator result-extraction fix** (`result_spec.go` + `coordinator.go`) — field-validated 2026-06-19 (idea.uk index built + deployed).

idea.uk page content (page builds now; these are the holes in the deployed `index.html`):
- [ ] **Differentiators section renders 7 empty cards.** NOTE: `reconcile_section_data` IS wired (registry.go L914, `ReconcileSectionDataAction`, "re-trigger pages whose deferred section data is now *query-resolvable*") — so this is NOT "wire the reconciler". Establish where differentiator items are sourced (query-resolvable section data / human spec field / writer prose) and fix the empty link. **[being split to a separate chat — see problem statement + bundle below]**
- [ ] **Unresolved CTAs** (hero + call-to-action buttons empty) — point at the real intake (`/request`, the email) or a real destination page; tied to the thin 4-page plan (no hub pages).
- [ ] **Dead contact form** (posts to `#contact`) — wire to the real flow or drop; page already lists email/phone.
- [ ] **Pricing** (`needs_section_data`, sourced from human `site_specs.pricing`) — capture the £29 into specs or drop the section.
- [ ] **Thin nav/footer** (only Home; footer "Our Services" lists only Refund Policy) + **empty meta description** — minor, downstream of the 4-page plan.

Design leadership (palette / typography / layout for fresh builds — idea.uk is the test case):
- [x] **Classifier emits structured `design_intent.palette.reference_values` + `typography.reference_values`** — migration applied 2026-06-20 (`migration_domain_research_classifier_structured_design_intent.sql`: edits the `classify_and_extract` schema + a MANDATORY bullet, `snapshot_agent` backup, live row only). Fresh `design_intent` now lands in the fields the composition cascade and the renderer actually read, instead of the prose `colour_mood` that nothing parsed. Takes effect on the **next** classifier run.
- [ ] **Deploy the dead-slot hardening** (`resolve_composition_reference_helpers.go` + the two `extractPaletteSignal` / `extractTypographySignal` swaps) — a Go change, so it needs a chassis image rebuild + roll of `site-design-planner`; separate from the DB migration above. Makes an adopted `design_reference` a real fallback *after* `design_intent` (it never fired before — wrong key). No agent-def/SQL change.
- [ ] **Validate idea.uk's palette** — the migration does NOT touch idea.uk's existing `design_intent` row (still prose-only). Either re-run the classifier for idea.uk, or patch its `design_intent` with the parchment block (primary `#1A1816` ink, accent `#A8391A` rust, background/surface `#EFE7D6`, text `#1A1816`, text_muted `#5A554C`, border `#1A1816`; body IBM Plex Sans, headings Fraunces) and requeue `needs_composition`. Confirm `palette_source = design_intent` in `resolved_composition.lineage`.
- [ ] **Layout picks ignore light/dark character** — `resolveLayoutByTags` (`fork_theme_composition.go`) matched on `category` + `industry_tags` only, so idea.uk's light-editorial intent landed on `tool-portal-dark`. **Fix written, ready to apply (2026-06-21):** (a) `migration_layouts_scheme_and_light_tool_portal.sql` — adds `layouts.scheme`, sets the two confirmed schemes, inserts a new `tool-portal-light` layout (light counterpart, flat/editorial, reads palette vars, tagged to win for idea.uk); (b) `resolveLayoutByTags_weighted.go` — weighted (IDF tags) + scheme-aware (near-hard constraint) + category/description scoring; needs the caller to pass `siteScheme` (`deriveSchemeFromDesignIntent`), the `layoutResolution` `Scheme`/`IsSchemeMismatch` fields, and a chassis image rebuild + roll of `site-design-planner`. **Validate:** apply (a), deploy (b), then re-resolve idea.uk (`1244516d…`): detach `style_collection_id` (idea.uk is fresh — do NOT reuse the adoption `source_domain` bulk delete) + re-queue `needs_composition`+`needs_design`; confirm `resolved_composition.layout_name = tool-portal-light` with the parchment palette. No auto-layout-generation (agreed) — a varied library + scheme-aware matching is the approach; LLM-judge/pgvector are later, separate tasks.
- [ ] **Components hardcode colours** — tool-portal-dark's header/footer carry baked-in blue rather than reading palette vars; downstream hygiene once layout + palette settle.
- [ ] **Curate layouts / typography / palettes** — add a light-editorial layout (or confirm a brochure layout fits idea.uk), and add Fraunces / IBM Plex to the typography library or accept resolve-on-insert.

Improver-not-rewriter overlay (fix 2 — the larger, separate piece; direction settled 2026-06-20, not started):
- [ ] **webdesign-agent overlay becomes an improver, not a rewriter** — Option 1 + safeguards in the full-palette-plus-diff form. **v1** (no contract change to 025 §5): the `analyze_design` prompt shows the established palette, asks it to keep that as the foundation and change a slot only with a reason, return all slots; the merge diffs the result against the composition base and writes an audit record (slot, old→new, reason) per build. **v2** (deferred, evidence-driven): cap how many core slots a refine may change + optional per-slot denylist; revert the excess to base. Check for an existing audit/log table before adding one. The classifier fix above is the precondition (the overlay now refines a real palette instead of inventing one).

Content ambition / mission:
- [ ] **Standing-ambition default in `domain-submitter`** — Go action carrying a default in the `mission` aspect merged with any owner mission (reuse `persist_mission_brief`/`write_site_spec`, no schema change); owner to finalise the principle wording (draft in notes).

Framework follow-ups (lower priority):
- [ ] (optional) confirm `resolveResultSpec … mode=flatten` / `result built mode=flatten` log on the validated run.
- [ ] broader fix validation: requeue gamesdesign's confirmed-stub page; and/or a full rebuild exercising the writer on every page.
- [ ] (optional) deprecated-key rename migration (`output_field`→`result_from`, `output_fields`→`multiple_output_fields`, `output`→`result_mapping`); old names keep resolving.

Housekeeping:
- [ ] Tidy duplicate idea.uk spec rows + the stale duplicate index work item (`bce22606`).

VM deploy — Phase 2 (hybrid; details in the Phase 2 section below):
- [ ] nginx as front door: static framework pages from `/var/www/idea.uk` + reverse-proxy the reserved tool paths to the Go process on 127.0.0.1:8080; never static-serve `/stripe/webhook` or operator paths; settle who serves `/terms`/`/privacy`/`/refund-policy`; cutover and rollback = one nginx location block; DNS unchanged.

Eventual:
- [ ] Full fresh rebuild of idea.uk as end-to-end confirmation + improved site (after content/design/mission in).

Parked (live £29 service — separate from this chassis work):
- [ ] Rewrite the report + audience-check language/format (less dense, nicer layout) — deferred earlier, not yet picked up.

---

idea.uk is not a blank brochure site: it already has a live, earning tool (the £29 report service on
the VM, behind nginx). The framework's normal way to *serve* a finished site is static hosting on
Backblaze B2 with DNS pointed at B2. If we served idea.uk that way, DNS would point away from the VM and
the tool would be bypassed. So we keep idea.uk's DNS on the **VM** and have the VM serve both: nginx
serves the framework's static pages and routes the tool's own paths to the Go service. The tool keeps
running unchanged; the framework owns the front site. (Building and deploying the static files to B2 is
still fine and harmless — see the DNS note — it just isn't what idea.uk's DNS points at.)

## Running each stage — commands

Inspection-first: every stage leads with read-only commands that confirm the schema and the current
state, and only then mutates — consistent with checking schema before SQL. Where a column name or a
trigger item_type isn't yet confirmed from a `\d`, it's marked `‹confirm›` rather than guessed.
Reminder: `logger.Info` surfaces in the logs, `logger.Debug` does not.

### 0. Basics (used by every stage)

```bash
# namespaces: ai-persona-system = app pods, kafka = Kafka.  DB pod: postgres-clients-0.
# idea.uk site_id: 97ed2f64-65ca-4b67-8a98-dfd8195a0d3a

# psql — interactive shell
kubectl exec -it -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db

# psql — one-off query
kubectl exec -n ai-persona-system postgres-clients-0 -- \
  psql -U clients_user -d clients_db -c "SELECT 1;"

# run a migration file (this is how the classifier migration was applied)
kubectl exec -i -n ai-persona-system postgres-clients-0 -- \
  psql -U clients_user -d clients_db < migration_domain_research_classifier_structured_design_intent.sql

# pods / deployments
kubectl -n ai-persona-system get pods
kubectl -n ai-persona-system get deploy          # find the worker/deployment names used below
kubectl -n kafka get pods                        # cluster: personae-kafka-cluster-...

# logs — find the worker that runs the workflow actions, then follow + filter
kubectl -n ai-persona-system logs -f deploy/‹worker› \
  | grep -iE 'design[_-]?intent|palette|typography|composition|webdesign|site-design-planner'
```

### Launch idioms (how a build is actually triggered)

Confirmed from the production trigger scripts. Two ways work starts:

**1. Orchestrate a static agent** — produce one Kafka message to `system.agent.generic.requests`. Used by
`rerender-pages` (re-assemble every page from existing components), `page-rerender` (one page), and
`page-rebuild` (LLM rewrite of pages in `needs_rebuild`):

```bash
U(){ cat /proc/sys/kernel/random/uuid; }
kubectl -n kafka run -i --rm kcat-$(date +%s) --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H message_type=request -H action=orchestrate -H client_id=demo_client \
  -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H correlation_id=$(U) -H orchestration_id=$(U) -H request_id=$(U) -H message_id=$(U) \
  -H timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ) <<JSON
{"action":"orchestrate","config":{"agent_type":"‹AGENT›"},"input_data":{"site_id":"‹id›","domain":"‹domain›"}}
JSON
```

**2. Launch a dynamic handler via a work item** — `page-build-handler` and the other job-spawned handlers
cannot be reached by a direct orchestrate; the running `build-dispatch-loop` spawns them when it claims a
`site_work_items` row (`status='triaged'`). So you INSERT the row and let the loop spawn it — the
production path, no hand-rolled wrapper (the 081c finding). This is the "reset / insert a work item"
route, and it's how `needs_composition` / `needs_design` get re-run.

**Caveat for the palette:** `rerender-pages` / `page-rerender` only re-assemble page HTML from existing
components, and `page-rebuild` only rewrites content — **none of them re-resolves the composition**, which
is where the palette lives. Changing idea.uk's colours runs through composition (`needs_composition` →
site-design-planner) and CSS render (`needs_design` → webdesign-agent); on an already-built site that
needs the composition reset first (Stage A, Route 2). Firing `rerender-pages` re-stitches the same theme.

### Submission entry points (fresh vs adopt)

Two entry agents, both idiom-1 orchestrate calls, converging on one work item:

- **Fresh domain** → `domain-submitter` (input `{domain, email?, phone?, mission_brief?}`). Creates the site
  row, stores contact/mission, queues `needs_domain_research`.
- **Adopt a source** → `site-adoption-orchestrator` (input `{target_url, destination_domain}`). Spawns
  site-adoption-agent: crawl → fingerprint → analyse → classify_archetype → content_direction →
  apply_adoption_plan (creates the destination site + seeds identity/archetype/design_reference/design_intent)
  → nav → generate_design_intent → queues `needs_domain_research`.

Both then run the **same** cascade: `needs_domain_research` → classifier (read-and-extend) → strategist →
briefing → planner → composition → webdesign → page-build → rerender. A fresh domain is simply "adoption from
the handoff point onward, minus the crawl." The unified trigger `082_submit_domain_unified.sh` picks the right
entry agent (`--from <url>` ⇒ adopt, else fresh) and carries a `--fidelity` field that is **recorded only**
today — doc 028's explicit fidelity input and its `build_policy`/`adoption_meta` aspect and per-item status are
not yet wired (fidelity is currently implicit `high`).

A full **adoption teardown** (for deleting/rebuilding an adopted site) lives in the adoption trigger script's
cleanup block: break `sites.style_collection_id` → `style_collections.css_theme_id` → delete forked library
rows by `source_domain`, preserving the baseline layouts (`tool-portal-dark`, `social-lobby`, `brochure-formal`,
`brochure-bold`, `technical-precise`, `affiliate-hub`). That bulk delete is for **forked** (adopted) rows; a
fresh site like idea.uk uses the lighter detach in Stage A Route 2 instead.

### Stage A — Validate idea.uk's palette

Confirmed by inspection (2026-06-20): the JSONB column is **`data`**; idea.uk's `design_intent` is the
diagnosed case — rich prose naming the parchment palette (`#EFE7D6` paper, `#1A1816` ink, `#A8391A` rust)
and Fraunces / IBM Plex, but **no** structured `palette.reference_values` or `typography.reference_values`,
and `style_direction: "professional-dark"` despite the palette being light. And
`sites.style_collection_id` is **set** (`d98d4a99-686e-46c9-befd-9bb1742307e9`): the composition is
already installed, so editing `design_intent` alone changes nothing on the page until the composition
re-resolves, and `install_site_composition` will not overwrite an existing one (a bare `needs_composition`
requeue no-ops). The migration only affects the **next** classifier run, so idea.uk's existing row is
unchanged by it.

**Route 1 — fresh staging rebuild (recommended; also validates the migration end-to-end).** Submit a fresh
domain through the unified trigger (`082_submit_domain_unified.sh`); the migrated classifier runs and now
writes the structured blocks, composition resolves them from scratch, and you get a clean parchment build —
the existing idea.uk row and its composition left untouched.

```bash
# fresh build = no --from. Use a staging domain to avoid colliding with the existing idea.uk row.
./082_submit_domain_unified.sh ‹staging-domain› --email idea-uk@leopardess.uk
```

Confirm the classifier now emits the structured blocks on the new site:

```bash
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \
"SELECT data->'palette'->'reference_values' AS palette,
        data->'typography'->'reference_values' AS typography
 FROM site_specs ss JOIN sites s ON s.id=ss.site_id
 WHERE s.domain='‹staging-domain›' AND ss.aspect='design_intent' AND ss.is_current;"
```

**Route 2 — fix the existing site in place (more steps; destructive to its current composition).** Patch
`design_intent`, then reset and re-resolve the composition.

Patch — note `palette` does not exist on the row yet, so `jsonb_set(data,'{palette,reference_values}',…)`
would **silently no-op** (`jsonb_set` leaves the target unchanged if any earlier path step is missing).
Use `||` to add the top-level keys. `jsonb_build_object` keeps the quoting clean. Back up first:

```bash
# backup the current design_intent row
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \
"CREATE TABLE IF NOT EXISTS bak_idea_design_intent_20260620 AS
 SELECT * FROM site_specs
 WHERE site_id='97ed2f64-65ca-4b67-8a98-dfd8195a0d3a' AND aspect='design_intent' AND is_current;"

# add structured palette + typography (|| merges new top-level keys; nothing existing is clobbered)
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \
"UPDATE site_specs
 SET data = data || jsonb_build_object(
       'palette', jsonb_build_object('reference_values', jsonb_build_object(
         'primary','#1A1816','secondary','#1A1816','accent','#A8391A',
         'background','#EFE7D6','surface','#EFE7D6','text','#1A1816',
         'text_muted','#5A554C','border','#1A1816')),
       'typography', jsonb_build_object('reference_values', jsonb_build_object(
         'font_family','\"IBM Plex Sans\", system-ui, sans-serif',
         'heading_font','\"Fraunces\", Georgia, serif')))
 WHERE site_id='97ed2f64-65ca-4b67-8a98-dfd8195a0d3a' AND aspect='design_intent' AND is_current;"

# (optional — only bites after a layout re-resolve) correct the light/dark label
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \
"UPDATE site_specs SET data = jsonb_set(data,'{style_direction}','\"modern-light\"')
 WHERE site_id='97ed2f64-65ca-4b67-8a98-dfd8195a0d3a' AND aspect='design_intent' AND is_current;"
```

Reset + re-resolve. idea.uk is a **fresh** build, not an adoption fork — its `css_theme` / `style_collection`
were created by `install_site_composition` and reference the **shared** library palette/layout/typography. So
do **not** reuse the adoption cleanup (which deletes forked rows by `source_domain`); that would risk deleting
shared library rows. The minimal, safe reset detaches the installed composition so install can run again
(`install_site_composition` refuses when `style_collection_id` is already set); deleting the old rows is
optional housekeeping, safe only when nothing else references them.

```bash
# capture the installed composition + check nothing else shares it
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \
"SELECT s.style_collection_id, sc.css_theme_id
 FROM sites s LEFT JOIN style_collections sc ON sc.id=s.style_collection_id
 WHERE s.id='97ed2f64-65ca-4b67-8a98-dfd8195a0d3a';"
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \
"SELECT count(*) AS sites_sharing FROM sites WHERE style_collection_id='‹style_collection_id from above›';"

# minimal reset — detach so install_site_composition will run again
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \
"UPDATE sites SET style_collection_id=NULL WHERE id='97ed2f64-65ca-4b67-8a98-dfd8195a0d3a';"

# optional housekeeping — ONLY if sites_sharing = 1 (the rows are idea.uk's alone)
# kubectl ... -c "DELETE FROM style_collections WHERE id='‹style_collection_id›';"
# kubectl ... -c "DELETE FROM css_themes        WHERE id='‹css_theme_id›';"
```

Then re-queue the design steps via the work-item idiom (the dispatch loop spawns the handlers). In a normal
build `needs_design` depends on `needs_composition`, so insert `needs_composition` first, let site-design-planner
re-install, then insert `needs_design`. Confirm the column set with `\d site_work_items` and mirror an existing
row:

```bash
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \
"INSERT INTO site_work_items (site_id, pipeline, item_type, handler_agent, status, priority, severity, summary, item_key, source)
 VALUES ('97ed2f64-65ca-4b67-8a98-dfd8195a0d3a','build','needs_composition','site-design-planner','triaged',7,'high',
         'Re-resolve composition after design_intent palette fix','needs_composition:97ed-reapply','manual');"

# after that completes (style_collection_id set again), queue the render:
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \
"INSERT INTO site_work_items (site_id, pipeline, item_type, handler_agent, status, priority, severity, summary, item_key, source)
 VALUES ('97ed2f64-65ca-4b67-8a98-dfd8195a0d3a','build','needs_design','webdesign-agent','triaged',8,'high',
         'Re-render CSS after composition re-resolve','needs_design:97ed-reapply','manual');"
```

(Route 1 avoids all of this — prefer it unless you specifically need to fix the existing row.)

**Verify (either route)** — the rendered result is schema-independent:

```bash
# follow the worker as it re-resolves (find the deploy name via: kubectl -n ai-persona-system get deploy)
kubectl -n ai-persona-system logs -f deploy/‹worker› | grep -iE 'composition|palette|typography|design'

# the deployed CSS should carry parchment, not blue
curl -s ‹deployed styles.css URL› | grep -iE '#efe7d6|#1a1816|#a8391a'      # expect these
curl -s ‹deployed styles.css URL› | grep -iE '2563eb|3b82f6|1e3a8a'         # expect nothing (blues)
```

### Stage B — Deploy the dead-slot hardening (Go)

This is a code change in the chassis image (`resolve_composition_reference_helpers.go` plus the two
`extractPaletteSignal` / `extractTypographySignal` swaps), so it ships through the normal pipeline —
github → GitHub Actions build → Backblaze — then a rollout of the worker that runs the actions. No
agent-definition or SQL change.

```bash
# 1. land the code (their CI builds on push; add a tag here if the pipeline keys on tags)
git add platform/orchestration/actions/resolve_composition_reference_helpers.go \
        platform/orchestration/actions/resolve_composition_pallette_action.go \
        platform/orchestration/actions/resolve_composition_typography_action.go
git commit -m "composition: design_reference fingerprint fallback after design_intent (dead slot fix)"
git push        # watch GitHub Actions for the image build

# 2. roll the deployment that runs site-design-planner's actions (find its name first)
kubectl -n ai-persona-system get deploy
kubectl -n ai-persona-system rollout restart deploy/‹worker›
kubectl -n ai-persona-system rollout status  deploy/‹worker›

# 3. confirm the new image is live
kubectl -n ai-persona-system get pods -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[0].image}{"\n"}{end}'
```

A clean way to exercise it is an **adopted** site rebuild (the fallback only fires when `design_reference`
is present and `design_intent` lacks a palette) — gamesdesign is the standing adoption case.

### Stages still to draft (no run-commands yet)

The **layout-scheme fix** (`resolveLayoutByTags` weighing the `design_intent` light/dark character) and
the **improver-not-rewriter overlay** (fix 2) are code changes, not yet written. Once written they ship
the same way: Go parts via the Stage B image-rebuild + rollout; any `analyze_design` prompt change in
`webdesign-agent`'s `agent_definitions` row ships as a `snapshot_agent`-backed SQL migration exactly like
the classifier one.

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
1. **Coordinator result-extraction — FIX FIELD-VALIDATED 2026-06-19 (`result_spec.go` + `coordinator.go`).**
   idea.uk's index built properly after deploy + requeue (needs_page:index claimed 09:51 → completed 10:08,
   ~17m — a real multi-section build, not the prior 41s short-circuit/claim-timeouts) and a fresh page_rerender
   produced a populated `index.html` deployed to B2. A populated page (vs the old empty stub) is the validation;
   final log confirmation = `resolveResultSpec … mode=flatten` / `result built mode=flatten` on that writer run.
   *Diagnosis (settled, git-confirmed):* the coordinator has read `output_fields`
   (plural) only, with the 900k cap + `extractMinimalResult` `status:"completed"` stub, since **commit 06a8c6ef
   (14 Jan)** — unchanged since (the 06-05→06-16 window query was empty; no ~06-13 coordinator change). The
   writer's `complete` was always singular `output_field` (backups Mar–Jun) → always took the fallback dump;
   it passed when the dump was <900k (06-06 gamesdesign; 01-16 logs at 46–147k) and collapsed to the silent
   stub when a big multi-section page's dump cleared 900k. **SIZE was the trigger; singular key
   necessary-but-not-sufficient.** (This corrected an earlier mistaken "coordinator changed ~06-13 / restore
   singular" reading — there was nothing to restore; it never honoured singular.)
   *Fix chosen (overrides the phase5 "don't touch `extractWorkflowResult`" decision — user went structural):*
   centralise the result contract in `resolveResultSpec` (new `result_spec.go`) and apply it in
   `extractWorkflowResult`:
   - **singular** `output_field` (alias of preferred `result_from`) → **FLATTEN**: the named field's map
     CONTENTS become the response body — restores the flat `page_content.response.{page_html,sections_metadata,
     skipped}` shape every consumer reads. (page-build-handler / page-rebuild / site-work-orchestrator;
     model-trainer reads `preparation_result.dataset_uri`.)
   - **plural** `output_fields` (alias `multiple_output_fields`) → FIELDS, nested under each name — UNCHANGED
     for the ~90 plural agents.
   - **`output`** (alias `result_mapping`) → MAPPING `target<-source.path` — previously silently dumped, now
     honoured.
   - none → fallback `skipPatterns` dump (unchanged, discouraged).
   Completion metadata via `setIfAbsent` (flattened payload not clobbered). **Oversize is now loud:**
   `extractWorkflowResultWithSizeLimit` returns an ERROR (not a stub); `notifyParentOfSuccess` routes it to
   `notifyParentOfFailure` with an `Error`-level per-field size breakdown naming the largest field + the remedy.
   `extractMinimalResult` removed. **Pure chassis change — no agent-config edits** (the deprecated keys all
   resolve). Deploy = rebuild image + roll agents. Wiring verified against the uploaded files (3296–3500);
   pre-deploy: grep package `orchestration` for existing `toStringMap`/`setIfAbsent` (redeclaration) + `go build
   ./...` (only 2 of N files seen here). Guardrail: do NOT raise `MaxResultSizeBytes`. **Retest: see "Retest
   after the result_spec fix" below — requeue idea.uk's index, do NOT re-adopt.**
2. **idea.uk deployed-page defects (observed 2026-06-19 in the built `index.html`; the build now works, so
   these are the next work — NOT fix regressions).** (a) *Differentiators rendered SEVEN empty cards* (heading
   present, all `h3`/`p` blank) while the generic method section + 13-item FAQ populated fully → likely a
   section-data source the writer doesn't fill (same class as the pricing `needs_section_data` gap); trace why
   it's empty when siblings populated — may be structural (the section-data reconciler we found unwired).
   (b) *Hero + call-to-action buttons render empty* (the 2 `unresolved_cta`s) — plan made only index + 3 policy
   pages, so no CTA destination; point them at the real intake (`/request` or the email) or a real target page.
   (c) *Contact form posts to `#contact` (dead)* — not wired to the real flow; page already lists email/phone →
   wire or drop. (d) *Pricing* into specs or drop the section. (e) *Thin nav/footer* (only Home; footer "Our
   Services" lists only Refund Policy) — downstream of the 4-page plan. (f) *Empty `meta description`*.
   (g) *Design against intent:* header blue gradient (#3b82f6→#64748b), footer solid blue, rounded corners — the
   3 AVOIDs in `design_intent`; theme/palette block came through EMPTY and header/footer hardcode blue (not via
   palette vars), so parchment/rust never reaches the page = the design-leadership workstream, now visibly
   confirmed. **NB this is the B2 staging build** (idea.uk still served by the Go tool via Cloudflare→VM) — fix
   before any VM cutover; no live exposure now.

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

### Retest after the result_spec fix (requeue, do NOT re-adopt)

The fix is a chassis change → deploy first; then requeue the one page that failed. Re-adopting (full fresh
rebuild) is the wrong first test: slow, many LLM calls / failure points (muddies whether the coordinator fix
worked), and reproduces idea.uk's known weaknesses (thin 4-page plan, CTA→phantom, no pricing spec, design not
leading, mission/standing-ambition pending). A full rebuild is the eventual end-to-end check, AFTER the fix is
proven and the content/design/mission work is in.

0. **Deploy.** Pre-merge grep package `orchestration` for `toStringMap`/`setIfAbsent`; `go build ./...`;
   build/push chassis image; roll agents to the new tag — at least coordinator, `page-content-writer`,
   `page-build-handler`, `site-work-orchestrator`. Confirm pods on the new tag before testing.
1. **(Optional, parallel — sets expectations.)** idea.uk index failure mode from the parent orchestration
   `collected_data.page_content.response`: "…exceeded size limit" stub vs genuine hang. (It failed on
   claim-timeout = the *hang* signature, so the fix may not be its blocker — worth knowing.)
2. **Look, then requeue only index** (target the specific failed row id, not all `needs_page:index`, to avoid
   parallel builds). `SELECT id,status,attempt_count,error FROM site_work_items WHERE
   site_id='97ed2f64-65ca-4b67-8a98-dfd8195a0d3a' AND item_type='needs_page' AND spec->>'page_name'='index';`
   then `UPDATE site_work_items SET status='triaged', attempt_count=0, error=NULL WHERE id='<failed index id>';`
   — **verify the claimable status value against the work-item state machine / the claim query's WHERE** before
   running (`triaged` is the assumed value, unconfirmed).
3. **Watch live** (`page-content-writer` + `page-build-handler`): expect `resolveResultSpec: resolved result
   contract` `mode=flatten matched_key=output_field` → `extractWorkflowResult: result built mode=flatten` at a
   sane size → sections saved >0 / components created → `page_rerender` commits `index.html` to B2. Oversize now
   logs an `Error` naming the largest field (no silent stub).
4. **Verify:** components-on-index query >0; `index.html` present in the bucket.
5. **Unambiguous proof-of-fix option:** also requeue the gamesdesign page with the *confirmed* stub — idea.uk's
   index might be a hang, which would muddy the read; gamesdesign was the known silent-completion case.
6. **Then, separately:** content gaps (pricing into specs or drop; CTA destinations need real hub pages),
   design-leadership, standing-ambition mission — then a full rebuild as the end-to-end confirmation.

### Finding the coordinator commit (git) — for fix-direction #1

**RESULT (2026-06-17):** the file is `platform/orchestration/coordinator.go`. `git log -L` shows only two
commits ever touch the three functions — **11394b22 (11 Jan)** created `extractWorkflowResult` (dumped all
collected_data) and **06a8c6ef (14 Jan, "made extract function respect output fields")** added the plural read
+ skipPatterns + `MaxResultSizeBytes=900000` + `extractWorkflowResultWithSizeLimit` + `extractMinimalResult` +
the `notifyParentOfSuccess` switch. No change since (the 06-05→06-16 window query was empty). So the behaviour
dates to 14 Jan, unchanged — confirming there was no ~06-13 coordinator change (see fix-direction #1). Commands
below kept for reference.

Run where `coordinator.go` lives. Locate first, then narrow precise → broad.

```bash
# 0. locate + confirm symbols exist now
git grep -n -E 'extractWorkflowResult|extractWorkflowResultWithSizeLimit|extractMinimalResult|MaxResultSizeBytes'
F=$(git grep -l 'extractWorkflowResult' -- '*.go' | head -1); echo "$F"

# 1. per-function evolution (precise — the commit that removed the singular branch shows it as a deletion)
git log -L ":extractWorkflowResult:$F"
git log -L ":extractWorkflowResultWithSizeLimit:$F"
git log -L ":extractMinimalResult:$F"
git log -s -L ":extractWorkflowResult:$F" --date=short --pretty='%h %ad %an  %s'   # list only
# if -L :func: can't find Go bounds:  printf '*.go diff=golang\n' >> .gitattributes

# 2. pickaxe (rename-robust, whole repo). "output_field" ⊂ "output_fields" → quote to keep distinct.
git log --oneline -S'"output_fields"'      -- '*.go'   # plural key introduced
git log --oneline -S'MaxResultSizeBytes'   -- '*.go'   # size cap appeared
git log --oneline -S'extractMinimalResult' -- '*.go'   # stub fn appeared
git log --oneline -G'"output_field"'       -- '*.go'   # any diff touching the singular key
git log --oneline -S'OutputFields'         -- '*.go'   # if read via struct field not raw key
git show <hash>                                         # read a candidate in full

# 3. window cross-check (phase2g)
git log -p --since=2026-06-05 --until=2026-06-16 -- "$F"
```

Corroboration for restoring the singular branch (not a workaround): `training-data-preparer`'s `complete`
step description states it uses "singular output_field for clean shape … return fields directly" — the exact
flat-return behaviour that regressed. `thunder-reaper` (`reaper_summary`) and `site-planner` (`validated_plan`)
are the same shape, small enough to escape the size path. `output_fields` (plural) is a real, widely-used
contract (the modern pipeline's ~100 steps), so the fix must honour BOTH — restore singular without disturbing
plural.

## Re-resolve idea.uk's composition onto the light layout (scheme fix) — exact steps

**Progress (2026-06-22):** layouts backed up (`bak_layouts_pre_scheme_20260621`, 17 rows) and `migration_layouts_scheme_and_light_tool_portal.sql` applied (ALTER + tool-portal-dark=dark + soft-editorial=light + `tool-portal-light` inserted, scheme light). Step 1 backups created (`bak_*_idea_20260621`: 1 site / 9 specs / 1 sc / 1 theme / 1 palette / 1 typo) and `reresolve_idea_uk_01b_inspect_only.sql` PASSED: current chain confirmed on `tool-portal-dark` (dark); the **four uniqueness checks are all 0**; `palette-idea-uk` + `typography-idea-uk` both have `source_site_id = idea.uk` (so the step-2 guarded deletes remove them); `tool-portal-light` present and active. **Cleared for step 2.** Do NOT re-run step 1 or the layouts backup (the `CREATE TABLE bak_*` would collide). Remaining gate: deploy the matcher code before step 3 (below).

Goal: make idea.uk (site `1244516d-014d-421c-88c6-090bb1e9552a`) re-pick its layout with the new weighted, scheme-aware matcher so it lands on `tool-portal-light` (parchment), in place, without a full rebuild. Files live in `/mnt/user-data/outputs/` and are run as FILES (never pasted).

Prereqs, in order:
1. Apply the layouts migration (adds `scheme` + the `tool-portal-light` layout). Back up the table first: `kubectl exec -i ... < <(echo "CREATE TABLE bak_layouts_pre_scheme_20260621 AS SELECT * FROM layouts;")`, then `kubectl exec -i ... < migration_layouts_scheme_and_light_tool_portal.sql`.
2. Deploy the matcher code. Merge the three edits (see the header of `resolveLayoutByTags_weighted.go`: struct fields + `math`/`sort` imports + replace the two funcs in `fork_theme_composition.go`; and drop in the merged `resolve_composition_layout_action.go`), rebuild the chassis image, and roll `site-design-planner` per your normal chassis deploy. No workflow JSON change is needed — the new output fields are additive and the scheme is derived inside the action from the specs.

Then the re-resolve (the four numbered files):
**Order note:** the layouts migration (prereq 1) MUST be applied before step 1 — step 1's inspect query reads `layouts.scheme` and checks for `tool-portal-light`, both added by that migration; running step 1 first aborts on `column l.scheme does not exist` (the backups still create, but the uniqueness checks don't print). If that happens: the backups exist — do NOT re-run step 1 (the `CREATE TABLE bak_*` would collide) — apply the migration, then run `reresolve_idea_uk_01b_inspect_only.sql` (inspect + uniqueness, backups skipped).

1. `reresolve_idea_uk_01_backup_and_inspect.sql` — backs up the whole composition chain (`bak_*_idea_20260621`), prints the current chain + palette/typography provenance, and the uniqueness checks. **Eyeball the uniqueness counts: all four MUST be 0.** If any is > 0, a row is shared with another site — stop, don't delete it.
2. `reresolve_idea_uk_02_detach_and_clear.sql` — run only after step 1's counts were all 0. Transactional: NULLs `sites.style_collection_id`, deletes idea.uk's `style_collection` → `css_theme` → `palette` → `typography_set` (palette/typography guarded by `source_site_id = idea.uk`), and supersedes the stale `resolved_composition` spec.
3. **GATE — deploy the matcher code first.** The merged `fork_theme_composition.go` + `resolve_composition_layout_action.go` must be built into the chassis image AND `site-design-planner` rolled to it before this step; otherwise the re-trigger spawns the OLD scheme-blind matcher and re-picks `tool-portal-dark`. Step 2 needs no code and is safe anytime; only this step does. Then `reresolve_idea_uk_03_trigger.sh` — orchestrates `site-design-planner` with `input_data {site_id}` (envelope identical to 082/080c). The planner re-runs composition (now scheme-aware → `tool-portal-light`) and emits its `needs_design` handoff, which the build-dispatch-loop routes to `webdesign-agent` to re-render `styles.css`. If design doesn't advance, re-run the trigger with `AGENT="webdesign-agent"`.
4. `reresolve_idea_uk_04_verify.sql` — expect `layout_name = tool-portal-light` (scheme light), the fresh `resolved_composition` with `palette_source = design_intent_values` and the parchment values, and **no** `needs_new_layout_candidate` queued this run. Then confirm (outside SQL) that the B2 build output re-rendered and the page reads light/parchment.

Rollback: restore from the `bak_*_idea_20260621` tables (re-insert the deleted rows, re-set `sites.style_collection_id`, and flip the old `resolved_composition` spec back to `is_current = true` while superseding the new one). DNS points at the VM, so none of this is visible to the live site during the work.
