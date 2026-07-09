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

### B6. Logo approval for robot-hands.com (first exercise of A3)
When Phase I1 lands, you'll be asked to approve robot-hands.com's logo — the
first use of the approve-then-lock flow. Budget five minutes to compare the
candidate against the site's brand direction.
**Status 2026-07-08:** getting closer — the rebuild's 20 `needs_imagery` items
(including the fresh logo) are queued behind the page builds; the agent will
flag you when the logo asset exists.

### B7. Layout-gap decision for robot-hands.com (from the 2026-07-08 rebuild)
The composition step found no layout candidate for the site's `scheme=dark`
and applied the `brochure-formal` fallback, raising a
`needs_new_layout_candidate` item with `status='needs_human_review'`.
Your decision, when convenient (not blocking the build):
- **Accept the fallback** — tell the agent; the item gets closed `wont_fix`
  with a note; or
- **Add a dark-scheme layout candidate** — the agent will scope what a new
  layout entry needs and bring you the proposal.
Look at the rebuilt pages first (they render with brochure-formal) — if they
look right, accepting is the cheap and reasonable path.

---

## C. What you should expect the agent to do (so you don't have to)
- All Go/SQL/prompt authoring, migrations drafted with backups + verification
  blocks, deploys through the standard git → GitHub Actions path.
- DB ground-truth checks each session (after your A1).
- Updating `RUNNING_NOTES_imagery_best_in_class.md` every turn and keeping
  `PLAN_imagery_best_in_class.md` in sync with decisions.
- Flagging exactly which of your tasks (A/B numbers) are needed next, rather
  than assuming you've read the whole queue.
