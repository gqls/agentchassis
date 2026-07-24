# RUNBOOK — about-page commercial (Phase 1 pilot)

Commands/queries with their gotchas. DB access:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

## The Phase-1 pilot sequence (order matters)

**GOTCHA #0 (the big one, scouted 2026-07-24):** `page_rerender` / `049b` can NEVER
add a section — it only re-renders existing `page_components` rows and escalates the
whole page to the writer if it meets an empty `content_data`
(`rerender_page_sections_action.go:153,186-230`; 016b §"RerenderSinglePageAction").
Adding a section requires the REBUILD path (re-runs `plan_sections` →
`SelectComponentByType` → resolve → render). Do NOT hand-INSERT `page_components`.

1. **Write the `commercial` aspect** (supersede-then-insert, one transaction —
   mirrors `HandleUpdateSiteSpec`, `site_admin_handlers.go:218-233`; data is free-form,
   no schema validation on write; resolver reads `is_current=true` only):
   ```sql
   BEGIN;
   UPDATE site_specs SET is_current=false, superseded_at=NOW()
     WHERE site_id='<SITE_ID>' AND aspect='commercial' AND is_current=true;
   INSERT INTO site_specs (site_id, aspect, data, source, created_by, is_current)
   VALUES ('<SITE_ID>', 'commercial', '<FACTS_JSON>'::jsonb, 'pilot-manual', 'uk', true)
   RETURNING id;
   COMMIT;
   ```
   FACTS_JSON stores **raw facts only** (class, tier, domain, for_sale_requested,
   advertising_active, inventory_open, marketplace_url, advertise_url, built_by_url) —
   gates are computed IN-TEMPLATE (see #2), never precomputed `show_*` booleans
   (write-time derivation = the staleness/write-order bug D5 exists to prevent).
   GOTCHA: store `tier` as a STRING ("1"/"2"/"3").

2. **INSERT the component** (seed file `docs/agent_docs/sql_for_agents/…_about_commercial_block.sql`).
   Selector contract (component_selector.go / 183:94-97): `section_type='about-commercial-block'
   AND component_level='section' AND is_active AND forked_from IS NULL` — new section_type
   ⇒ sole candidate ⇒ deterministic selection, no retire step.
   Template GOTCHAS:
   - Renderer is Go **text/template** with a custom string-comparing `eq`/`ne`
     (call_agent.go:1160) → `{{if eq .tier "1"}}` is safe for JSON numbers AND strings.
   - Allowed funcs ONLY: if/else/range/with/and/or/not (builtins) + eq/ne/default/
     lower/upper/isset/safe (registered). ANY other func ⇒ Parse error ⇒ silent
     fallback to a regex renderer that mangles `{{else if}}` blocks.
   - `missingkey=zero`: absent field = falsy ⇒ `{{if …}}` gates fail closed.
   - Keep the literal closing `</section>` (truncation guard skips components without it).
   - Gates in-template: for-sale = `{{if and (eq .class "portfolio") .for_sale_requested (not .advertising_active)}}`;
     advertise = `{{if and (eq .class "portfolio") .inventory_open (not .for_sale_requested)}}`;
     built-by = `{{if or (eq .class "portfolio") (eq .class "storefront")}}`.
   - Every commercial field's `input_schema` source = `site_specs.commercial.<key>` →
     resolved in plan_sections (:418-420,477-493), merges LAST ⇒ LLM-proof. (These do
     NOT depend on the late resolve_internal_links pass, so the "content_data edits
     don't hold" trap does not apply.)

3. **Add the section to the pilot about page's section list** — in the store that is
   authoritative FOR THAT SITE, checked in priority order:
   `site_plan_sections` row → `site_specs.site_plan` aspect `pages[].sections` (~5 older
   sites) → `pages.sections`. Mirror into `pages.sections` so the cache agrees.

4. **Flag ONLY the about page**: `UPDATE pages SET build_status='needs_rebuild' WHERE id='<ABOUT_PAGE_ID>';`
   (leave siblings `deployed` — only needs_rebuild pages rebuild).

5. **Fire the `page-rebuild` agent** (039_page_rebuild_agent.sql) — skips the site
   planner, per-page plan_sections → select → write → render → deploy.
   GOTCHA: this re-runs the content writer on the page's OTHER sections (hero-about,
   about-content) — pilot must be a page whose copy we can regenerate. There is NO
   path that adds a section without a rebuild.
   GOTCHA: no dispatch within ~300s of a chassis pod restart; queue latency can be
   ~30min — a missing orchestration row is latency, not a drop.

6. **Verify against the LIVE PAGE, not the status row** (bugs_closed/024 lesson):
   curl the served about page and grep for `data-component="about-commercial-block"`
   AND one phrase the template CREATED (e.g. "available to acquire") — never a generic
   CSS property. Check `page_components.rendered_html` length changed. Then check the
   gates: keeper site (or absent aspect) must render NOTHING.

## Pilot-candidate query (sites with an about page and NO open work items)

```sql
SELECT s.id, s.domain, s.status,
       (sp.data->>'site_type') AS site_type,
       p.id AS about_page_id, p.page_type, p.build_status
FROM sites s
JOIN pages p ON p.site_id = s.id
   AND (p.page_type='about' OR p.name ILIKE 'about%' OR p.url ILIKE '%about%')
LEFT JOIN site_specs sp ON sp.site_id=s.id AND sp.aspect='classification' AND sp.is_current=true
WHERE s.status='active'
  AND NOT EXISTS (SELECT 1 FROM site_work_items wi
                  WHERE wi.site_id=s.id
                    AND wi.status NOT IN ('complete','cancelled','rejected'))
ORDER BY s.domain;
```
GOTCHA: `site_work_items.status` is free text with many live values — the broad
NOT IN is deliberate. Site "type" is NOT a column on `sites`; it lives in the
`classification` aspect. Exclude by hand: keepers/clients (leopardess), storefront
(fundamentallyai), test (vonc), idea.uk (VM-served, sitesync lag).

## Honesty preconditions before a line goes LIVE

- For-sale line requires the domain to actually BE listed on Afternic (floor set) —
  otherwise the link is a lie/404. No listing yet ⇒ `for_sale_requested=false`.
- Advertise line requires advertise.co.uk to exist (other thread) ⇒
  `inventory_open=false` until it does.
- Built-by is TRUE today (fundamentallyai.com is live) — the pilot can ship with
  built-by only and the other two gated off, which ALSO proves the gates.
