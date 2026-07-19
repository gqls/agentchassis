# HANDOFF — robot-hands.com site defects (SUPERSEDED)

> **SUPERSEDED 2026-07-19 — do NOT start a new chat here.**
> Start at **`HANDOFF_2026-07-19_robot_hands_start_here.md`**.
> R1, R2, R3, R5 and R6 are DONE and verified live; R4 is partly done (the
> "blank MatchMatrix" is diagnosed and fixed; two tool pages remain unbuilt).
>
> **Four of this file's stated causes were WRONG** — corrected in
> `PLAN_2026-07-19_robot_hands_site_fixes.md` and evidenced in
> `NOTES_robot_hands_site_fixes.md`. In brief: the blue header predates the
> 2026-07-16 regeneration by a week (it entered 07-09); `component_versions`
> holds NO snapshots to restore from; `hardcoded_section_colors` never modified
> anything (every run fails and is stamped complete anyway — now
> `bugs_open/017…unregistered_action…`); and MatchMatrix was never blank.
> The real cause was three un-migrated palette copies plus deactivated
> colour-baking components, amplified by a fleet-wide `generic_theme` misfire.
>
> Kept unedited below for the R1–R6 framing and its prior-art pointers.

**Filed:** 2026-07-17 after the imagery D13 gate surfaced site-wide problems
that are NOT imagery. The imagery-quality half is the SIBLING handoff:
`../imagery/HANDOFF_2026-07-17_i3_imagery_gate_fixes.md`. Interlocking items
marked ⇄. **Several of these are "we've been here before" — the prior-art
pointers below are load-bearing; read them before fixing.**

## What this project is (fresh-reader paragraph)

agentchassis is an autonomous agent platform operating a fleet of content
websites. One Go runtime executes declarative workflows (JSONB in
`agent_definitions`); agents cooperate over Kafka; work flows through
`site_work_items` (discovery → dispatch → handlers → git deploy →
GitHub Actions → Backblaze B2 behind a Cloudflare worker).
**robot-hands.com** (site_id `00ff3af5-dad8-4770-9f70-3edc267a3c92`) is the
imagery workstream's testbed: rebuilt from scratch 2026-07-08 (33-page plan,
news, tools), layout set to **tool-portal-dark** (dark scheme) via the B7 fix
on 2026-07-10. This week's imagery work ran several improvement-loop passes on
the site; the user's gate then found the site visually and functionally
degraded. DB access:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`.

## The defect list (user gate, 2026-07-17) with verified evidence

### R1. Whole site moved from dark theme to a standard blue brochure look
**WE HAVE BEEN HERE BEFORE — but the mechanism is DIFFERENT this time. Do not
replay the old fix blindly.**
- Prior art (MUST READ): imagery `RUNNING_NOTES` Turn 18 + PLAN "B7
  RESOLVED/COMPLETED" blocks + `SQL_2026-07-10_b7_layout_fix.sql` (industry
  tags) + `SQL_2026-07-10_b7_layout_swap.sql`. Key facts from B7: there is
  **NO runtime re-compose path by design** (install_site_composition refuses
  when a collection exists; fork_theme_from_site install mode removed
  2026-04-19); the sanctioned way to change composition is the **025 FK-swap
  pattern** on the site's own `css_themes` row.
- **VERIFIED 2026-07-17: the B7 state is INTACT.** `css_themes`
  `b1b60faf-ca68-43f5-a1e6-da3a769e4a25` (theme-robot-hands-com) still points
  at layout `tool-portal-dark` (scheme=dark), untouched since 2026-07-10.
  **The served `/assets/css/styles.css` IS tool-portal-dark** (header comment
  confirms). So the stylesheet did NOT regress — do NOT re-swap the FK.
- The regression is at the **component/page level**. Evidence trail
  (site_work_items, 2026-07-16/17):
  - `needs_rerender` "Site component header/footer/head is stale" —
    **completed 2026-07-16 14:18** (during the first D13-era improvement
    loop). The header was REGENERATED — the blue gradient header + standard
    nav in the user's screenshot is the regenerated component not using the
    dark theme's classes/vars. Same failure class as the corrupted-template
    saga (Turns 11–13): a regeneration path that isn't theme-aware.
  - `hardcoded_section_colors` items COMPLETED 2026-07-16 14:19 and
    2026-07-17 09:39 — "brief-explanation-section uses hardcoded fallback
    background #0d0d0d instead of theme var". A fixer that strips hardcoded
    DARK backgrounds on a DARK site plausibly turned sections white. Find
    what its handler actually rewrote (component_versions has pre-regen
    snapshots).
  - Recurring `generic_theme` — "Site using default theme — no
    industry-specific..." (unresolved ×2 attempts, fires every pass): the
    check misjudges tool-portal-dark as "default". If its handler ever runs
    it may try to "improve" the theme. Neutralise or fix the check's
    judgement.
  - `needs_rerender` "37 pages missing header/footer" (unresolved ×2,
    recurring) — the re-render handler fails; pages are inconsistent.
- Fix direction: make header/footer/site-component regeneration theme-aware
  (or restore the pre-2026-07-16 header/footer from `component_versions`
  snapshots); re-render CSS+pages through the known-good kcat triggers
  (webdesign-agent for CSS; rerender-pages with
  `refresh_site_components:true`); audit what hardcoded_section_colors'
  handler changed and revert bad edits; stop the generic_theme misfire.
- **DO NOT re-plan the site** — see memory/doc: re-running build-site-planner
  regresses built pages (`replan-clobbers-built-pages`). DO NOT hand-roll
  spawn+call inline workflows (spawned child runs INIT then idles — Turn 26).

### R2. learning-center-hub.html: yellow text on now-white background
Downstream of R1 (the dark background vanished). **User's fix: restore the
dark background**, not recolour the text. Verify after R1's re-render; there
is also a `forced-text-color-fixer` in the codebase — make sure it doesn't
"fix" the yellow once the dark background returns.

### R3. URL sprawl: /learning-center.html vs /learning-center-hub.html
- Nav (verified on served index) links **/learning-center.html** (the
  category info-card-grid page). **/learning-center-hub.html** (the article
  listing with the new cards) is NOT in the nav — it's only reachable
  directly. Both serve 200. There is ALSO a `learning-center-index` page
  (orphan component residue) — the "learning-center sprawl" has been a
  known cleanup item since the rebuild (PLAN I0 cleanup queue; orphan pages
  how-it-works, selection-guide too).
- Decide the canonical learning-center IA (likely: hub = the listing,
  category grid folded in or kept as landing), fix nav, retire/redirect the
  rest. ⇄ imagery F2/F3: what pages exist determines what `query.blog_posts`
  lists and where cards appear.

### R4. Tools mostly broken; MatchMatrix blank; tools link to descriptions not tools
**Precedent exists — see docs before fixing.** Verified now:
- Nav "MatchMatrix" → `/matchmatrix.html` (200, 31,659B — a CONTENT page
  about the tool; user reports it renders blank → likely an empty tool shell
  or dead JS on it).
- Index/tool-list cards → `/tools/matchmatrix/index.html` → **404** (tool
  page never deployed).
- `/tools/gripper-payload-calculator/index.html` → 200/37KB (that tool
  exists).
- So BOTH failure modes are present: broken tool links AND
  content-page-instead-of-tool. Prior art: `check_misdirected_cta.go`
  (misdirected_pages) — 16 `cta_names_unknown_destination` +
  `unresolved_cta` items already sit in this site's needs_human_review
  queue; and the **experience_loop workstream (started 2026-07-17,
  `../experience_loop/HANDOFF_2026-07-17_experience_loop_start.md`)** covers
  exactly this class fleet-wide (vonc arena/gauntlet broken tools,
  tool_acceptance_due sweeps, page-ownership marker so generic rebuilds
  don't clobber tool pages). **Coordinate with that chat rather than
  duplicating its machinery** — robot-hands may just be its next case set.
- Also: 5 `tool`-role pages exist in the plan (payload calc, cycle time,
  grip force, matchmatrix, robot payload budget) — sweep all 5:
  which /tools/<x>/index.html exist, which nav/card hrefs point where.

### R5. "Load More" button does nothing
`content-listing`'s input_schema has static fields `show_load_more`
(fallback true) + `load_more_text` ("Load More") — the template renders a
button with NO behaviour wired. Options: implement pagination JS (house
pattern: per-site committed `/tools/assets/*.js` bundle), or set the
fallback/content_data to false so the dead button disappears. With only ~3
real articles today, hiding it is the honest v1. ⇄ imagery F2's eligibility
filter changes the item count.

### R6. 6 of the 9 listed articles are 404 on the live site
DB rows are status='active' but never deployed (verified 2026-07-17):
- 3 scaffold/template pages: `/blog/news-post.html`,
  `/blog/learning-center-article.html`, `/blog/learning-center-post.html`
  (placeholder titles — plan-era scaffolding).
- 3 duplicates of the /guides/ articles under /blog/
  (`grip-force-friction-calculator-guide`, `gripper-cycle-time-estimator-guide`,
  `gripper-payload-calculator-guide`) — note one real article ALSO lives at
  `/blog/tool-gripper-payload-calculator-guide.html` (200) while its
  siblings are under /guides/ — the /blog/ vs /guides/ split is itself
  inconsistent.
- Decide per row: build it, or retire/supersede the page row (which also
  fixes what the listing shows — ⇄ imagery F2). Beware: retiring pages
  affects `site_plan_imagery` scope_refs and the entity-linked card assets
  (they'd become orphans — supersede those too).

## Mechanisms a fresh session must know (hard-won this week)
- **Dispatch priority is ASC** (LOWER = sooner; 5 front, 99 last;
  `load_work_item_actions.go:589`).
- **Improvement-loop triage strands items in `detected`** (twice this week);
  hand-promote small batches: status='triaged', triaged_at=now().
- **Zombie claims** >10 min block the whole site's dispatch — but needs_page
  builds legitimately run 20+ min; don't blanket-clear while one is in
  flight.
- **Manual agent triggers** via kcat → `system.agent.generic.requests`
  (`action=orchestrate`, config.agent_type improvement-loop /
  webdesign-agent / rerender-pages; full example imagery notes Turn 18 /
  `sql_for_agents/033_rerender_pages_trigger.sh`). `system.intake` is stale.
- **Verify deploys against the RUNNING POD binary**, never the tag (tags are
  rebuilt in place): `strings /app/agent-chassis | grep -c "<marker>"`.
- The image-landing/article-blank trap is CLOSED (guard live in prod; all 17
  article bodies recovered — `../aaa_fails_to_mend/004…` and `005…`).
- Component regeneration: `needs_component_regeneration` → component-creator;
  put exact field-preservation demands in `spec.description` if the
  pre-store guard rejects. Pre-regen snapshots live in `component_versions`.
- Site work-item state right now includes a large stale triaged backlog
  (page_rerender@35/80, phantom_internal_link@35, undeployed_asset@60...)
  — expect queue friction; don't assume an idle item means broken dispatch
  (check attempt_count vs max, ASC order, zombies).

## Suggested attack order
1. R1 theme restoration (biggest visual damage; R2 falls out of it).
2. R3 nav/IA decision (small, user-facing).
3. R6 build-or-retire the 404 rows (unblocks imagery F2 properly).
4. R4 tools — jointly with the experience_loop chat.
5. R5 load-more (trivial once counts are known).

## Key artifacts
- `../imagery/SQL_2026-07-10_b7_layout_fix.sql` + `SQL_2026-07-10_b7_layout_swap.sql`
  (B7 precedent; the FK-swap pattern — NOT needed this time, state verified intact).
- `../imagery/RUNNING_NOTES_imagery_best_in_class.md` Turns 11–18 (corrupted
  templates + regeneration + B7), Turns 44–47 (this week's runs).
- `../experience_loop/HANDOFF_2026-07-17_experience_loop_start.md` (tools class).
- Memory: `replan-clobbers-built-pages` (fleet-wide landmine — do not re-plan).
- `component_versions` table (pre-regeneration snapshots for header/footer
  restoration).
