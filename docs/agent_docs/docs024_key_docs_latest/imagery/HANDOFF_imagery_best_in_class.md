# HANDOFF — Imagery best-in-class workstream (continue in a fresh chat)

**Last updated: 2026-07-10 evening (Turn 22). Keep this document updated at
every working turn, alongside the running notes.**

**What this project is about (fresh-reader paragraph).** agentchassis is an
autonomous agent platform that plans, builds, and operates a fleet of content
websites (tools, guides, games, news feeds, articles). Sites are planned by
LLM agents, content/imagery generated through a Kafka-orchestrated work-item
pipeline (discovery checks → site_work_items → dispatch → handler agents),
and assets committed to each site's git repo. The "imagery best-in-class"
workstream (started 2026-07-08, branch 083_imagery then
084_site_improvements_local_ai) raises the fleet's visual quality: brand
consistency, data-accurate infographics, card imagery, sprite-sheet bullets,
product sketches, news imagery, performance budgets, and an audit loop.
**Testbed: robot-hands.com** (site_id `00ff3af5-dad8-4770-9f70-3edc267a3c92`).

## Read these first (all in this directory)
1. `PLAN_imagery_best_in_class.md` — goals G1–G9, decisions D1–D8 (all
   user-confirmed), phases I0–I8 with status blocks. THE map.
2. `RUNNING_NOTES_imagery_best_in_class.md` — Turns 1–22, every decision and
   diagnosis with evidence. Append a turn every session.
3. `RUNBOOK_imagery_best_in_class.md` — the human's tasks (A-rituals,
   B-queue). B1–B3, B7, B8 done; B4–B6, B9 open.
4. SQL artifacts `SQL_2026-07-*.sql` beside these docs — every migration run,
   with backups.

## State of the world (2026-07-10 evening)

**Phase I0 (rebuild + render acceptance): SUBSTANTIALLY COMPLETE.**
- robot-hands rebuilt fresh (33-page plan, news on index + news-index page,
  9 news sources seeded/0 erroring, healthy content layer).
- Per-page imagery renders VERIFIED at scale: 16 distinct git-path heroes
  (`/assets/images/hero-<page>.jpg`), zero expiring S3 URLs, 1 empty image
  slot site-wide (learning-center-index listing card = Phase I3 scope).
- Still open from I0: logo-in-header render path (render_site_components
  doesn't resolve site_assets.logo); Lucide validator unwired.

**Phase I1 (brand consistency): CODE LIVE, ACCEPTANCE IN FLIGHT.**
- `imagery_style_guide` site-spec aspect + `imagery_style_guide.go`:
  generate_image composes per-kind direction (photographic kinds get
  medium+mood+palette; icons palette-only; logos nothing), avoid→negative
  prompt, reference_asset_keys→stable s3:// anchors for Banana. Guide
  SUPERSEDES free-text design_intent.imagery_direction. robot-hands' guide is
  seeded (charcoal/electric-blue, industrial photography, anchors
  hero_canonical+hero_home).
- D5 lock guard live: assets upsert refuses rows with locked_at set.
- ACCEPTANCE PENDING: 3 needs_imagery items queued on robot-hands will be the
  first generations through the guide — check their assets.origin_prompt for
  "industrial photography"/"charcoal" fingerprints. A background monitor was
  watching; re-run the check by hand if the chat ended:
  `SELECT asset_key, origin_prompt FROM assets WHERE site_id='00ff3af5-…' AND created_at > '2026-07-10 16:30';`
- REMAINING I1: logo approve-and-lock flow (B6) + favicon/OG derivation —
  GATED on the leopardess IMAGE_BUCKET fix (below).

**B7 (layout): COMPLETED.** robot-hands moved brochure-formal →
tool-portal-dark. Route: classification lacked industry_tags (fixed,
supersede-row) + there is NO runtime re-compose path by design
(install_site_composition refuses; fork install mode removed 2026-04-19) →
targeted `css_themes.layout_id` swap (025 pattern) + webdesign-agent CSS
deploy. 36 page_rerender items refresh the HTML side, draining behind the
design-audit backlog. USER EYEBALL (A3) requested.

**Self-healing guards built + registered (both live in production):**
- `image_source_unsatisfiable` — component asks for a site_assets.<path>
  nothing supplies → flag item (needs_human_review, no handler).
- `component_template_corrupted` — baked/rendered-output templates (literal
  `<no value>` or zero {{vars}} with schema fields) → auto-emits
  needs_component_regeneration to component-creator. PROVEN end-to-end
  (tool-guide-intro auto-detected → regenerated clean). 7 corrupted actives
  remain fleet-wide (finetuning.uk, vonc.com, leopardessconsulting.co.uk) —
  self-heal when their discovery cycles run; can be forced with the
  improvement-loop trigger (below).
- page-build-handler error routing fixed: real step errors now mark the work
  item `failed` (mark_item_failed step) instead of silently completing.

## Key mechanisms a fresh session needs

**DB access:** `kubectl exec -n ai-persona-system postgres-clients-0 -- psql
-U clients_user -d clients_db` (a PreToolUse hook auto-approves read-only
SELECT/\d via this path; mutations prompt the user).

**Manual agent trigger (system.intake is STALE — do not use):** kcat pod →
`system.agent.generic.requests` with header `action=orchestrate` and body
`{"headers":{…},"config":{"agent_type":"<agent>"},"input_data":{"site_id":…,"domain":…}}`.
Working example in running notes Turn 18 / script precedent
`docs/agent_docs/sql_for_agents/033_rerender_pages_trigger.sh`. Used for:
improvement-loop (full discovery cycle), webdesign-agent (CSS re-render+
deploy), rerender-pages (+`"refresh_site_components":true`).

**Manually-inserted work items are NOT auto-triaged** — insert with
status='triaged', triaged_at=now(). Dedup index: (site_id, item_key) where
status not terminal.

**Zombie claims:** a claimed item >~10min with no progress blocks the WHOLE
site from dispatch. Standing workaround (run when queues stall):
`UPDATE site_work_items SET status='triaged', claimed_by=NULL, claimed_at=NULL, updated_at=now() WHERE status='claimed' AND claimed_at < now() - interval '10 minutes';`
Real fix = reaper cadence bump (user's TODO 6/10/11, RUNBOOK B9).

**Component regeneration:** queue `needs_component_regeneration` →
component-creator (spec: function/component_id/section_type + optional
`description` — the creator prompt renders it; USE IT to demand exact schema
field preservation when the guard rejects a regen).

**Deploys:** user deploys chassis via git → GitHub Actions. Go changes are
inert until then. After deploys: clear zombies, re-trigger interrupted loops.

## Cross-workstream notes (CORRECTED 2026-07-10 evening)
- **The IMAGE_BUCKET scare was RETRACTED by the leopardess workstream**
  (`docs/leopardessconsulting/AUDIT_verified_facts.md` D8): the deploy
  pipeline WORKS by design — the base chassis intentionally carries no
  storage vars; the real pipeline spawns asset-deployer with storage env
  INJECTED (spawn_actions.go isStorageEnabledAgent). Live asset paths serve
  HTTP 200; `assets.url` presigned staleness is cosmetic (83 rows; fix =
  generalise the idea.uk w9_04 url-flip backfill). **B6 (logo flow) is
  therefore NOT gated** — regenerating the logo through the normal
  image-build-handler → asset-deployer chain deploys it correctly. (Verify
  whether robot-hands' logo.png is committed; if not, one needs_imagery item
  for the logo regenerates+deploys it, then approve → set locked_at → done.)
- leopardess also committed a dynamic_adapter routing change
  (logo/illustration/infographic → Banana, honours reference images) —
  deployed with the 2026-07-10 builds; their workstream verifies.

## Next actions (in order — updated Turn 23)
1. Verify the in-flight drains: 36 page_rerenders + 3 needs_imagery + the
   fresh logo item on robot-hands; then the I1 style-guide acceptance check
   (origin_prompt fingerprints on the new assets) and eyeball two
   generations for shared brand voice.
2. User A3 eyeball of tool-portal-dark on robot-hands.com.
3. **B6 logo — ✅ APPROVED + LOCKED (Turn 24, user chose existing logo).**
   assets.logo (May-8, purpose='hero' — LEFT AS-IS so DeployedWebPath
   resolves to the deployed logo.jpg) is locked (locked_at,
   lock_type=permanent, locked_by=user-b6-approval).
4. **I1 is now FEATURE-COMPLETE pending deploy** — style guide, logo lock,
   header logo resolution, and favicon/OG derivation all built. Two Go
   commits await the NEXT CHASSIS DEPLOY: header logo (b00c150b) and
   favicon/OG (derive_brand_head_assets + head injection). POST-DEPLOY
   ACTIVATION for robot-hands, in order:
   a. Trigger `derive_brand_head_assets` via a storage-enabled agent
      (asset-deployer) with {site_id, domain} — commits favicon.png +
      og-card.png. (kcat orchestrate → asset-deployer with an inline
      workflow calling the action, OR add a step; see Turn-18 trigger
      pattern.)
   b. `rerender-pages` with refresh_site_components=true → head gets the
      favicon/OG tags, header gets the locked logo.
   c. Eyeball: logo in header, favicon in the tab, OG card on a link
      preview (or view-source the injected <head>).
4. Remaining I0 stragglers: wire the Lucide validator; orphan old-page
   cleanup (how-it-works, selection-guide, learning-center sprawl).
5. Optionally trigger improvement-loop for finetuning.uk / vonc.com /
   leopardessconsulting.co.uk to sweep the 4 remaining corrupted components
   (archetype-taster-quiz, lobby-grid, provocation-card, tool-cta).
6. Then Phase I2 (sprite sheets — design already locked in
   CONTEXT_PACK_imagery_sprite_sheet.md) and I3 (card imagery / Lane B:
   generic entity_type+entity_id columns on assets — user-confirmed D2).

## Open investigation threads (parked, non-blocking)
- Kafka per-job response-topic partition race ("topic partition not found") —
  transient; now visible as failed items rather than swallowed.
- Runtime re-compose path missing by design — build a site-design-planner
  re-resolve mode if layout changes become routine.
- learning-center-index orphan component row (component_id NULL, pre-rebuild
  residue) — clears with I3 card imagery or a listing rebuild.
