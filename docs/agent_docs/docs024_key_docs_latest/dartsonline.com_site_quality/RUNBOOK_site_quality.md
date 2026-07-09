# RUNBOOK — Site Quality Programme (handoff from the builder-route thread)

## THE TASK (read this first if you are new)

**Mission:** every site this platform builds should be *best in class* for its
vertical. **Current reality:** the first site built end-to-end by the work-item
relay (dartsonline.com, site_id `5fe8785b-223d-41a3-88ee-c07187622381`) deploys
successfully but is bare — no navigation, no images, no logo, no graphics, no
interactive tools, no news/feeds, thin content, and an unstyled look. **Your
job:** run a systematic, evidence-first quality programme that closes the gap
between "deploys" and "best in class", leg by leg, without guessing — measure
which legs never dispatched, which dispatched but produced poor output, and
which were never in scope at all, then fix in that order.

**How this platform works (one paragraph):** a Go chassis (`agentchassis`) on
Kubernetes (`-n ai-persona-system`); every agent is an orchestrator owning a
JSON workflow of steps calling Go actions; Kafka messaging; Postgres state
(`sites`, `site_specs`, `site_work_items`, `pages`, `page_components`,
`content_components`); sites deploy via git commits → GitHub Actions →
Backblaze S3. Builds run as a work-item relay: each hop is a `site_work_items`
row naming a `handler_agent`; a 30s scheduler (`build-pipeline-trigger` →
`build-dispatch-loop`) claims items and spawns handlers. The standing working
rules (schema before SQL; snapshots before agent_definitions updates; reuse
before recreate; structural over quick fixes; evidence over assumption; 0-rows
not decisive until the query is checked) apply throughout.

**Sibling documents:** RUNBOOK_builder_route.md (the relay map §B0–§B5, the
spine decision, the §B4 vertical-exemplar researcher — that thread OWNS relay
hops and routing); RUNBOOK_code_retrieval_route.md (the read-only diagnosis
loop — available as an instrument here: symptom in, cited code+data+runtime
diagnosis out); the parallel TOOLS chat (RUNNING_NOTES_travelling_docs) OWNS
tool-pipeline internals, tool docs, and tool-auditor — the planned-tool-page
seam is a JOINT decision already framed in the builder runbook §B5.

## MEASURED BASELINE (2026-07-06, the four rendered pages)

| page | bytes | nav | img | svg | script | css-var refs | stylesheet links |
|---|---|---|---|---|---|---|---|
| index | 14,412 | 0 | 0 | 0 | 0 | 50 | 1 |
| about | 13,125 | 0 | 0 | 0 | 0 | 36 | 1 |
| contact | 4,871 | 0 | 0 | 0 | 0 | 14 | 1 |
| new-arrivals | 5,188 | 0 | 0 | 0 | 0 | 25 | 1 |

Mechanical readings: components USE the `var(--color-*)` convention but the
single stylesheet link points at `/assets/css/styles.css` while the
`needs_design` item (webdesign-agent, "Generate site stylesheet") was last
seen `triaged` — palette referenced, palette likely never delivered ⇒ browser
defaults. Zero `<nav>` on every page ⇒ hypothesis: the RELAY build path lacks
the site-chrome rendering step (pageflow-builder has `render_site_components`
for header/footer/head; build-site-planner's workflow was read in §B3 and has
`populate_nav_tables` — nav DATA — but no chrome-render step was observed).
new-arrivals carries only 3 hrefs (2× /contact.html + the stylesheet) — a
retail listing page with no product links; the reported "broken links" need
verifying (candidates: the stylesheet 404ing; links on other pages).

## THE THREE-WAY SPLIT (fix in this order)

**A. Dispatched but STUCK / never delivered (fix first — nothing looks right
without these):**
- LEG 1 — site chrome: verify the relay path renders header/footer/nav at all
  (site_components rows for this site? does the assembler inject chrome?);
  compare pageflow-builder's render_site_components vs the relay.
- LEG 2 — design: the needs_design item → webdesign-agent → does
  /assets/css/styles.css exist in the sites repo? `sites.style_collection_id`
  was empty at submission — did composition/webdesign ever populate it?
- LEG 3 — imagery: 13 needs_imagery items (logo, heroes, icons) →
  image-build-handler — claimed? failed? handler exists and spawns correctly?
  (Remember the §B4 lesson: seeded/newer handlers can carry the stale-`latest`
  image-tag trap; check the row.)

**B. Delivered but POOR (fix second):**
- LEG 4 — content depth: thin components (contact 4.9KB, no form script);
  page-content-writer prompting, the per-section content family (prototyped,
  experimental — builder runbook §B0e), validate_page_content coverage; the
  unresolved_cta needs_human_review class (about page) — a structural
  CTA-resolution candidate is already parked in the builder runbook.
- LEG 7 — link integrity: full-site href audit vs the pages table + rendered
  set; the user reports broken links on new-arrivals.

**C. Never in SCOPE (fix third — these are PLANNING criteria, not build bugs):**
- LEG 5 — feeds/news/RSS/ticket feeds; graphics/infographics; games where apt:
  the strategist/planner never put them in scope for this site type. The news
  feed pipeline exists (docs 006; content-feed-refresh is enabled 6-hourly)
  but needs per-site attachment; graphics/tools ride the JOINT seam (§B5,
  on hold with the tools chat).
- LEG 6 — the improvement loop: `improvement-sweep` is DISABLED platform-wide
  (builder runbook §B2). Enabling it (per-site first?) brings the auditors
  (design-audit, content-quality, visual-design, site-review) into play —
  coordinate the tool-auditor part with the tools chat.

## FIRST MOVES (numbered; evidence before any change)

1. The three §B6 queries (site 5fe8785b-…): undone/failed build items; the
   design-leg landing check (`style_collection_id` + composition/webdesign
   specs); per-page component count/bytes/thin-count. (Full SQL in the builder
   runbook §B6 block — transplant verbatim.)
2. Does `/assets/css/styles.css` exist in the sites repo for this domain? One
   `git ls-files` / repo browse. If absent → LEG 2 is a dispatch question.
3. `SELECT * FROM site_components WHERE site_id='5fe8785b-…'` (schema first) —
   does the relay path create chrome at all? Then find where the assembler
   injects header/footer (pageflow does; does page-rerender?).
4. needs_imagery + needs_design item states with attempts + any
   agent_error_log rows; check image-build-handler's agent_definitions row
   (image_tag pinned? is_active?).
5. Full link audit across the rendered pages vs `pages` rows.
6. THEN pick the first fix by the pre-stated rule: dispatch problems before
   content problems; scope additions (C) only after A+B stand.

## BOUNDARIES (do not cross without coordinating)
- Tools chat: tool pipeline internals, tool docs, tool-auditor, and (jointly)
  the planned-tool-page seam.
- Builder-route thread: relay hops, routing, the spine, the researcher.
- This programme: legs 1–7 above, per-page quality, and the planning-scope
  criteria additions (LEG 5) — propose planner/strategist prompt changes back
  through the builder thread if they alter relay hops.
