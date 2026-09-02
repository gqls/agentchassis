-- 712 — new `event-list` component (query.upcoming_events-sourced).
--
-- bugs_open/427 Phase B, STEP 1 of 2 (this migration is deliberately the safe
-- half only — see NOTES_bugfix_427_event_render.md for why step 2 is held).
-- Everything this component depends on is already live — the
-- query.upcoming_events resolver (council-reviewed, resubmission in
-- progress) and 6 real dated event facts registered for boxingonline.com
-- (news_feed_ingestion, 2026-09-02).
--
-- THIS MIGRATION ONLY CREATES THE COMPONENT IN THE LIBRARY. It attaches to
-- NO page and has NO live effect until a page's `sections` names it — safe
-- to apply on its own.
--
-- STEP 2 (NOT in this migration, deliberately): declaring `event-list` on
-- boxingonline.com's live fight-calendar page. The framework's own
-- check_unresolved_sections.go (LIVE, completeness-discovery-agent) is the
-- designed mechanism for "a page's declared sections name a component that
-- exists but isn't built yet" — it marks such a page `build_status =
-- 'needs_rebuild'`, and get_pages_to_build_actions.go's own status filter
-- (`planned`, `needs_rebuild`) then routes it through the FULL
-- page-build-handler pipeline, the same path a brand-new page takes — NOT
-- the scoped, no-LLM page-rerender path (PBP-013) that only refreshes a
-- query-sourced field on a component ALREADY present. This is going THROUGH
-- the framework, not around it (CLAUDE.md: "EVERY SITE GOES THROUGH THE
-- FRAMEWORK"), and it is the framework's OWN designed answer — but it is
-- also a wider blast radius than "add one section": it is the same pipeline
-- that plans and writes a page from scratch, running against a PAID
-- CUSTOMER'S ALREADY-DEPLOYED page whose other two sections (hero-tool,
-- generic-text-block) carry approved copy. This session verified the
-- MECHANISM (component definition, template, resolver, all render correctly
-- — see the standalone text/template harness run in NOTES) but did NOT
-- verify that a `needs_rebuild` pass on an already-deployed generic page
-- carries forward EXISTING section content untouched rather than
-- re-planning/re-writing it — that is the one open question standing
-- between this and Phase B being complete, named rather than guessed past.
--
-- Every optional field is `{{if}}`-guarded in the template (council REVISE
-- round 2, bug_historian HIGH: an unguarded absent key renders blank with no
-- error under Go's missingkey=zero) — venue, broadcaster, source_url,
-- source_title, and the whole per-item block are each independently gated.
-- The disclaimer is STATIC copy in the template, not read from content_data:
-- it never varies per item, so hand-writing it once is more robust than
-- depending on the resolver's per-item field surviving into every future
-- cache/copy of the row (compliance's "must travel with the claim" ask,
-- satisfied unconditionally rather than conditionally).
--
-- REPLAY GUARD: the component insert is NOT EXISTS on name; the sections
-- update is idempotent jsonb_agg(DISTINCT).

BEGIN;

INSERT INTO content_components (
    name, display_name, description, function, section_type, category,
    render_mode, html_template, input_schema
)
SELECT
    'event-list', 'Event List', 'Query-sourced list of dated, evidenced upcoming events (bugs_open/427)',
    'event-list', 'event-list', 'content-site',
    'template',
    $tmpl$<section data-component="event-list" class="event-list-section" id="event-list">
  <div class="event-list-container">
    {{if .headline}}<h2 class="event-list-title">{{.headline}}</h2>{{end}}
    <div class="event-list-items">
      {{if .items}}{{range .items}}
      <article class="event-list-item">
        <div class="event-list-item-date">{{.date}}</div>
        <div class="event-list-item-content">
          <h3 class="event-list-item-title">{{.title}}</h3>
          <div class="event-list-item-meta">
            {{if .venue}}<span class="event-list-item-venue">{{.venue}}</span>{{end}}
            {{if .broadcaster}}<span class="event-list-item-broadcaster">Watch: {{.broadcaster}}</span>{{end}}
          </div>
          {{if .source_url}}<a class="event-list-item-source" href="{{.source_url}}" target="_blank" rel="noopener noreferrer">{{if .source_title}}{{.source_title}}{{else}}Source{{end}}</a>{{end}}
        </div>
      </article>
      {{end}}{{else}}<p class="event-list-empty">{{if .empty_text}}{{.empty_text}}{{else}}No confirmed fixtures are listed yet — check back soon.{{end}}</p>{{end}}
    </div>
    {{if .items}}<p class="event-list-disclaimer">Schedule details can change after this was checked — confirm with an official source before relying on it.</p>{{end}}
  </div>
</section>
<style>
.event-list-section { padding: 4rem 2rem; background: var(--color-background, #f8fafc); }
.event-list-container { max-width: 760px; margin: 0 auto; }
.event-list-title {
  font-size: clamp(1.5rem, 3vw, 2rem); font-weight: 700; letter-spacing: -0.02em;
  color: var(--color-heading, #0f172a); margin: 0 0 2rem;
}
.event-list-items { display: flex; flex-direction: column; gap: 0; }
.event-list-item {
  display: flex; gap: 1.5rem; padding: 1.5rem 0;
  border-bottom: 1px solid var(--color-border, #e2e8f0);
}
.event-list-item:last-child { border-bottom: none; }
.event-list-item-date {
  flex-shrink: 0; width: 6.5rem; font-weight: 700; font-variant-numeric: tabular-nums;
  color: var(--color-primary, #0f172a);
}
.event-list-item-content { display: flex; flex-direction: column; gap: 0.4rem; }
.event-list-item-title { font-size: 1.125rem; font-weight: 600; line-height: 1.35; margin: 0; color: var(--color-heading, #0f172a); }
.event-list-item-meta { display: flex; flex-wrap: wrap; gap: 1rem; font-size: 0.9rem; color: var(--color-text-muted, #64748b); }
.event-list-item-source { font-size: 0.85rem; color: var(--color-accent, #0369a1); text-decoration: none; }
.event-list-item-source:hover { text-decoration: underline; }
.event-list-empty { padding: 2rem 0; text-align: center; color: var(--color-text-muted, #64748b); }
.event-list-disclaimer { margin-top: 1.5rem; padding-top: 1rem; border-top: 1px dashed var(--color-border, #e2e8f0);
  font-size: 0.8rem; color: var(--color-text-muted, #64748b); }
</style>$tmpl$,
    $schema${
      "fields": {
        "items": {
          "type": "array",
          "items": {
            "fact_id": {"type": "text"}, "title": {"type": "text"}, "date": {"type": "text"},
            "venue": {"type": "text"}, "broadcaster": {"type": "text"},
            "source_title": {"type": "text"}, "source_url": {"type": "url"}
          },
          "limit": 20,
          "source": "query.upcoming_events",
          "required": false,
          "on_missing": "skip_field"
        },
        "headline": {"type": "text", "source": "llm", "required": false,
          "llm_guidance": "Short heading for the fixture list, e.g. 'Upcoming Fights'. Under 8 words. Never state a count or a date — the list itself does."},
        "empty_text": {"type": "text", "source": "llm", "required": false,
          "llm_guidance": "One sentence shown when no fixtures are confirmed yet. Honest, not apologetic — e.g. 'No confirmed fights yet — check back as fight week approaches.'"}
      },
      "notes": "items are server-rendered from query.upcoming_events (bugs_open/427) — a site's own evidence_base register, evidence-gated (citation url+quote required) and re-resolved on every section_data_resolved rerender the register's own refresh queues. The disclaimer is static template copy, not a data field."
    }$schema$::jsonb
WHERE NOT EXISTS (SELECT 1 FROM content_components WHERE name = 'event-list');

COMMIT;

-- STEP 2, deliberately NOT run here — see the header. Once the carry-forward
-- question is answered (or the owner accepts the risk knowingly), the
-- attach step is exactly:
--
-- UPDATE pages
-- SET sections = (SELECT jsonb_agg(DISTINCT x)
--                 FROM jsonb_array_elements(COALESCE(sections, '[]'::jsonb) || '["event-list"]'::jsonb) x)
-- WHERE site_id = 'd2aa5206-73bc-4707-a69c-2702c1eb9152'
--   AND name = 'tool-fight-calendar'
--   AND NOT (sections @> '["event-list"]'::jsonb);

-- VERIFY: component exists, page declares it, exactly once.
SELECT name, function, section_type FROM content_components WHERE name = 'event-list';
SELECT sections FROM pages WHERE site_id = 'd2aa5206-73bc-4707-a69c-2702c1eb9152' AND name = 'tool-fight-calendar';
