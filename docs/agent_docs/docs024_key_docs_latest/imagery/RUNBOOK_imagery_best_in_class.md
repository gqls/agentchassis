# RUNBOOK — Imagery: best-in-class sites (operator tasks)

**What this project is about (read this first).** agentchassis is an autonomous
agent platform that plans, builds, and operates a fleet of content websites.
Its imagery pipeline generates logos, heroes, illustrations, icons, and
infographics via LLM-planned prompts and two image providers (Stability SDXL,
Google Banana), committing optimised assets to each site's git repo. The
current workstream (`PLAN_imagery_best_in_class.md`) upgrades the fleet's
visual quality to best in class: data-accurate infographics, consistent brand
language, content-linked card images, sprite-sheet bullets, copyright-safe
product sketches, news imagery, and performance budgets. Most of the work is
autonomous, but some steps need a human: credentials, key rotation, database
backups before migrations, visual approvals (logos, sprite sheets), budget
sign-off, and final eyeballing of rendered pages. **This document is the list
of those human tasks, with the commands to run.** robot-hands.com is the
testbed domain.

Conventions used below:
- `psql` means: `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`
- Tasks are grouped: **standing rituals** (repeat every time they apply) and
  **one-off tasks** (checked off once, tracked in the running notes).
- In a Claude Code session you can run any of these yourself by prefixing the
  command with `!` at the prompt, which puts the output straight into the
  conversation.

---

## A. Standing rituals

### A1. Refresh cluster credentials at session start
The DB-backed steps fail with `You must be logged in to the server` when your
kubeconfig auth expires (this is why the seeded bundle's schema sections are
empty). At the start of a working session run your usual cluster login, then
verify:
```
kubectl -n ai-persona-system get pods | head -5
```
If that lists pods, the agent can proceed with DB verification and migrations.

### A2. Backup before any agent-definition or schema migration
Per doc 009 convention, every migration session starts with a timestamped
backup of the rows being touched, e.g.:
```
psql -c "CREATE TABLE agent_definitions_backup_YYYYMMDD_<label> AS
         SELECT * FROM agent_definitions WHERE type = '<agent>' AND is_active = true;"
```
The agent will provide the exact statement per migration; your job is to run it
(or approve it) *before* the migration and confirm the row counts match.

### A3. Visual approval gates
Generated art the agent cannot judge alone. When asked, open the referenced
image/page and answer approve / reject / adjust:
- **Logo approval (per site, once):** the logo is generated, you approve it,
  then it is LOCKED for the life of the site (favicon and social cards derive
  from it). Rejecting triggers regeneration; approving is meant to be final.
- **Sprite sheet check (per site):** confirm each grid cell reads as a clean
  glyph at bullet size and note what each cell should mean (cell meanings are
  assigned after generation, by design).
- **Sampled page eyeball (per phase acceptance):** load the listed
  robot-hands.com pages and confirm the acceptance line in the plan (e.g.
  "cards' images match their articles", "chart axes are real data").

### A4. Refresh the context bundle when starting a fresh agent thread
```
z_bundles/imagery_seed_docs/imagery_bundle.sh
```
Run it *after* A1 so the schema/runtime sections populate. Output lands at
`z_bundles/imagery_bundle.md`. Override site/page with `SITE=… PAGE=…` env vars.

### A5. Discovery/dispatch nudges
Some verification steps need a discovery pass or a stuck item unstuck. The
agent will hand you exact SQL/kubectl one-liners; typical shapes you'll see:
```
-- re-run a discovery pass result check
psql -c "SELECT item_type, status, count(*) FROM site_work_items
         WHERE site_id = '<site>' GROUP BY 1,2 ORDER BY 1;"
```
Nothing here is destructive; run as provided and paste the output back.

---

## B. One-off tasks (current queue)

### B1. ~~Confirm the API-key scrub and rotation happened~~ ✅ DONE 2026-07-08
Keys rotated (user confirmed). Kept for the record: the 2026-05-20 finding was
plaintext logging of `STABILITY_API_KEY`/`BANANA_API_KEY` + B2 keys.

### B2. ~~Confirm robot-hands.com rebuild status~~ ✅ ANSWERED 2026-07-08
Not rebuilt; site is in poor shape. **Decision: rebuild from scratch and add
news at the same time; full latitude on this site.**
**Update (same day, evening):** the re-plan has RUN and completed — new
33-page plan with news in the header nav; content build is working through
the queue unattended. See PLAN Phase I0 status block and running notes Turn 4.

**Update 2026-07-09:** build essentially done — 14 hero images generated and
deployed to git overnight, content layer healthy. Known issue (agent-side, no
action needed from you): heroes don't render on 4 pages because some preset
components ask for image keys the generator doesn't produce
(`site_assets.background`/`product_screenshot`/etc.). Diagnosed; fix approach
in PLAN "I0 finding". You may still see empty image areas on product-detail /
about / gripper-detail / learning-center-index until that fix lands.

**Update 2026-07-10 — RESOLVED.** Two deploys later the imagery rendering is
verified end-to-end: every re-rendered page shows its own committed-path hero
(`/assets/images/hero-<page>.jpg`), zero expiring S3 URLs, and the three
corrupted component templates (content-block-about, product-specs,
info-card-grid) were regenerated via the component-creator path. One empty
image slot remains site-wide (learning-center-index listing card — Phase I3
scope). Full story: running notes Turns 8–14.

### B3. ~~Answer the seven open questions~~ ✅ DONE 2026-07-08
All answers recorded in the running notes (Turn 2) and folded into the plan.

### B4. Data-source API keys (needed before Phase I4)
Decision 2026-07-08: data sources are chosen **per domain** (each domain gets
the source appropriate to its vertical); **free-tier only for now**, paid
sources may be added later. When I4 starts, the agent proposes the source for
the domain in hand (for robot-hands.com: robotics/automation trade data;
generic fallbacks: FRED https://fred.stlouisfed.org/docs/api/, EIA
https://www.eia.gov/opendata/). Your task: register and add the API key(s) as
cluster secrets.

### B5. Budget sign-off for increased generation volume
Best-in-class imagery multiplies per-site image counts (cards, news items,
product sketches). Banana runs on a paid Google Cloud tier. Before Phases
I3/I5/I6 scale beyond the testbed, review expected volumes with the agent and
set per-pass caps you're comfortable with (current default: ~20 generations
per site per discovery pass).

### B6. ~~Logo approval for robot-hands.com~~ ✅ DONE 2026-07-11
You approved the existing (May-8) logo as-is. It is now **locked**
(`assets.locked_at` set, lock_type=permanent) — the store guard refuses any
future overwrite, making it the site's permanent mark. It already serves at
`/assets/images/logo.jpg`. **It will appear IN THE HEADER after the next
chassis deploy** (the header-resolves-logo fix, commit b00c150b, then a
site-component re-render). Remaining logo work is agent-side: favicon + OG
card derived from this locked logo (Phase I1 tail).

### B7. ~~Layout-gap decision for robot-hands.com~~ ✅ COMPLETED 2026-07-10
Your decision: no brochure fallback. Root cause was missing `industry_tags`
in the old-format classification (fixed, supersede-row); the library already
held `tool-portal-dark`. The re-composition path turned out to be unsupported
by design ("re-resolve not supported"), so the switch was executed as a
targeted `css_themes.layout_id` migration (025 pattern,
`SQL_2026-07-10_b7_layout_swap.sql`) + webdesign-agent CSS re-render/deploy.
**👀 YOUR A3 EYEBALL IS NOW DUE: open robot-hands.com — it should read as a
dark, engineered tool-portal (charcoal/electric-blue), not a formal
brochure.** HTML-side re-renders are still draining through dispatch; minor
header/footer refresh may land after your first look.

### B8. ~~Run the check-registration SQL after your NEXT chassis deploy~~ ✅ DONE 2026-07-10
Chassis deployed; `component_template_corrupted` registered on
design-discovery-agent (snapshot taken, verified in live config). The 8
remaining corrupted components fleet-wide (lobby-grid, archetype-*,
game-master-explanation, platform-comparison, provocation-card, tool-cta,
tool-guide-intro — mostly games sites) now regenerate automatically as those
sites' discovery passes run; the agent set a watch for the first emissions.
`image_source_unsatisfiable` was already registered earlier.

### B9. Infra priority nudge (your call, already scoped in your TODO)
The zombie-claim dispatch stall was the single biggest time cost of the
2026-07-09/10 verification — every batch needed manual claim-clearing
(15-min-stuck items block the whole site from dispatch). Your existing TODO
items 6/10/11 (reaper cadence bump, per-item-type circuit breaker) are now the
highest-leverage infra fix for every future build. Recommended before the next
site build.

### B10. ~~Sprite-sheet deploy (jpg switch)~~ ✅ DONE 2026-07-13
The jpg switch deployed; the sheet is live at `/assets/images/sprite-sheet-main.jpg`
(768², 75,745 bytes, under the ≤80KB budget).

### B11. ~~Sprite-sheet EYEBALL GATE~~ ✅ PASSED 2026-07-13/15
You confirmed the nine glyphs (check, gauge, gripper, cog, chart, download,
arrow, info, warning) and later confirmed the themed list bullets render
correctly on a live page. Cell map written back to the plan; **Phase I2 is
COMPLETE** (see B12).

### B12. Sprite bullets — Phase I2 CLOSED (2026-07-15). One pending gate.
I2 is complete and live: the sheet, per-page glyph bullets, a self-healing
fulfilment check (fires once, then quiet), and a "house style" so ordinary
article lists theme themselves. **You chose the neutral ARROW as the default
bullet** (check reserved for explicit use). That change is committed but rides
the next chassis deploy — it self-heals on the next discovery pass.
**👀 AFTER THE NEXT DEPLOY (your A3):** open the friction-calculator guide and
confirm the plain content-list bullets flipped from check to arrow. Fast check:
`curl -s https://robot-hands.com/assets/css/sprites.css | grep -o 'li::before[^}]*background-position:[0-9px -]*}' | head -1`
— the default should read `background-position:0px -40px` (arrow), not `0px 0px`
(check). No action if it hasn't flipped yet; it lands with the format-3 binary.

### B13. Content-loss fixes — NOT imagery; guard now LIVE, recovery still open
While doing I2 the work surfaced pre-existing platform bugs that silently lose
live content (empty product pages; article bodies stored as unparsed JSON; an
image landing that can blank an article). These are handed off to dedicated
threads — you are already driving the article-body / image-landing pair.
Written up in `../aaa_fails_to_mend/004_HANDOFF_image_landing_blanks_article_body.md`,
`../HANDOFF_2026-07-14_article_body_json_envelope.md`, and
`../HANDOFF_2026-07-14_empty_product_sections.md`.
**Update 2026-07-16: the escalate-not-blank guard IS live in prod** (verified in
the running pod) — the standing image-landing hazard is lifted. Still open in
your separate thread: 4 pages remain blanked (finetuning ×3, gamesdesign ×1;
the 004 handoff's §5 correction has the exact list and the fixed detection
query — the old `length=1326` test under-reports).

### B14. Phase I3 (card imagery) — post-deploy acceptance run
Everything for I3 is built and the DB side is already live (assets entity
columns; asset-deployer `content_card` mode; `content_image_missing` check
registered — inert until the binary carrying it deploys). **After your next
chassis deploy:**
1. Trigger an improvement-loop / design-discovery pass for robot-hands.com
   (kcat pattern, notes Turn 18). Two things fire in one pass:
   `sprite_css_missing` (format 2→3 → the B12 arrow default lands) and
   `content_image_missing` (→ up to 9 `needs_content_image` items →
   asset-deployer derives `card-<page>.jpg` crops from each article's hero).
2. When the card items complete, queue a re-render for `learning-center-hub`
   (needs_page — it must re-resolve sections so `query.blog_posts` populates
   the content-listing with items + images; assemble-only won't do that).
3. **👀 A3 gate:** open `/learning-center-hub.html` (hard-refresh) — the
   article listing should show cards, each with its own article's hero-family
   image (800×450 crop, ≤60KB, served from `/assets/images/card-*.jpg`).
   Click through: the article page shows the same image family as its card.

---

## C. What you should expect the agent to do (so you don't have to)
- All Go/SQL/prompt authoring, migrations drafted with backups + verification
  blocks, deploys through the standard git → GitHub Actions path.
- DB ground-truth checks each session (after your A1).
- Updating `RUNNING_NOTES_imagery_best_in_class.md` every turn and keeping
  `PLAN_imagery_best_in_class.md` in sync with decisions.
- Flagging exactly which of your tasks (A/B numbers) are needed next, rather
  than assuming you've read the whole queue.
