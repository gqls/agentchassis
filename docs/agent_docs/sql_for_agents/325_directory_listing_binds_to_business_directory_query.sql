-- 325 — directory-listing binds to a REAL, registered query; was dangling
--
-- bugs_open/206: `directory-listing`'s `entries.source` read
-- `query.directory_entries` — not a name `queryresolve.Resolve` has ever
-- registered (the real registered names are `model_directory` /
-- `model_directory_full` / `adoption_tracker` / `protocol_tracker` and their
-- siblings, all scoped to the GLOBAL directory_entities/directory_claims
-- registry). Confirmed live 2026-08-06: `usage_count=0` and, more reliably
-- (usage_count is not a trustworthy signal fleet-wide — checked separately),
-- zero pages anywhere carry this component in their `sections`:
--
--     SELECT s.domain, p.name FROM pages p JOIN sites s ON s.id=p.site_id
--     WHERE p.sections @> '"directory-listing"'::jsonb;   -- 0 rows
--
-- So this UPDATE cannot regress a live page — there is not one.
--
-- WHAT THIS DOES. Rebinds the component to `query.business_directory`, the
-- new resolver shipped alongside this migration
-- (queryresolve/business_directory.go) that reads a SITE'S OWN
-- directory-export-json config (business_intel.businesses, filtered
-- exactly as directory_export_action.go's loadDirectoryEntries already
-- does) — so the rendered listing and that site's exported JSON archive can
-- never name a different set of businesses. Fields are rewritten to match
-- what that data actually has (name / postcode / location / website /
-- is_claimed) instead of the AI-directory shape (region / category /
-- category_slug) this component's fields never matched anyway. The
-- unused `filter_categories` field and its category-filter chrome/JS are
-- removed — nothing produces categories for this data.
--
-- Cards now link OUT to the business's own website (external, target=_blank
-- rel=noopener) — there is no per-practice page to link to yet
-- (`entity-page` stays unavailable, `bugs_open/206`) — and show an
-- "Unclaimed listing" note when `is_claimed` is false, matching this site's
-- own homepage "Claim your practice profile" messaging.
--
-- `suitable_page_types` gains `entity-directory` (was `["index","directory"]`
-- only) — that is the actual page_type this component now serves.
--
-- ORDER: DB config, live the moment it commits. Does nothing to any page
-- until a page's plan actually names this component (326 + the Go builder
-- change do that for vetcomparison.uk's directory-index).
--
-- Apply:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < 325_….sql

BEGIN;

CREATE TABLE IF NOT EXISTS content_components_bak_20260806_dirlisting AS
SELECT * FROM content_components WHERE function = 'directory-listing';

UPDATE content_components
SET input_schema = '{
  "fields": {
    "entries": {
      "type": "array",
      "items": {
        "name":       { "type": "text" },
        "postcode":   { "type": "text" },
        "location":   { "type": "text" },
        "website":    { "type": "url" },
        "is_claimed": { "type": "boolean" }
      },
      "source": "query.business_directory",
      "required": true,
      "min_items": 1,
      "on_missing": "skip_section"
    },
    "headline": {
      "type": "text",
      "source": "llm",
      "required": true,
      "llm_guidance": "Section heading for the business directory, e.g. Find a Local Vet Practice, Browse Verified Suppliers"
    }
  }
}'::jsonb,
    html_template = $TPL$<style>
.directory-listing-section {
  padding: var(--spacing-section, 4rem 1.5rem);
}
.directory-listing-section .directory-container {
  max-width: var(--container-max-width, 1200px);
  margin: 0 auto;
}
.directory-listing-section .directory-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}
.directory-listing-section .directory-title {
  color: var(--section-heading, var(--color-heading, #0f172a));
  font-size: 1.75rem;
  font-weight: 700;
  margin: 0;
}
.directory-listing-section .directory-count {
  color: var(--section-text-muted, var(--color-text-muted, #64748b));
  font-size: 0.875rem;
}
.directory-listing-section .directory-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.5rem;
}
.directory-listing-section .directory-card {
  background: var(--color-card-bg, #ffffff);
  border: 1px solid var(--color-border, #e2e8f0);
  border-radius: var(--border-radius, 0.5rem);
  padding: 1.5rem;
  transition: box-shadow 0.2s, border-color 0.2s;
}
.directory-listing-section .directory-card:hover {
  box-shadow: var(--shadow, 0 2px 8px rgba(0,0,0,0.08));
  border-color: var(--color-primary, #3b82f6);
}
.directory-listing-section .directory-card-name {
  color: var(--section-heading, var(--color-heading, #0f172a));
  font-size: 1.125rem;
  font-weight: 600;
  margin: 0 0 0.5rem;
}
.directory-listing-section .directory-card-meta {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 0.75rem;
}
.directory-listing-section .directory-card-location,
.directory-listing-section .directory-card-postcode {
  font-size: 0.8125rem;
  color: var(--section-text-muted, var(--color-text-muted, #64748b));
}
.directory-listing-section .directory-card-link {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-primary, #3b82f6);
  text-decoration: none;
}
.directory-listing-section .directory-card-link:hover {
  text-decoration: underline;
}
.directory-listing-section .directory-card-unclaimed {
  display: inline-block;
  margin-top: 0.5rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--section-text-muted, var(--color-text-muted, #64748b));
  background: var(--color-surface, #f8fafc);
  border: 1px solid var(--color-border, #e2e8f0);
  padding: 0.125rem 0.5rem;
  border-radius: 2rem;
}
@media (max-width: 640px) {
  .directory-listing-section .directory-grid { grid-template-columns: 1fr; }
}
</style>
<section class="directory-listing-section" data-component="directory-listing">
  <div class="directory-container">
    <div class="directory-header">
      <h2 class="directory-title">{{.headline}}</h2>
      {{if .entries}}<span class="directory-count">{{len .entries}} listings</span>{{end}}
    </div>
    <div class="directory-grid">
      {{range .entries}}
      <div class="directory-card">
        <h3 class="directory-card-name">{{.name}}</h3>
        <div class="directory-card-meta">
          {{if .location}}<span class="directory-card-location">{{.location}}</span>{{end}}
          {{if .postcode}}<span class="directory-card-postcode">{{.postcode}}</span>{{end}}
        </div>
        {{if .website}}<a href="{{.website}}" class="directory-card-link" target="_blank" rel="noopener">Visit website</a>{{end}}
        {{if not .is_claimed}}<div><span class="directory-card-unclaimed">Unclaimed listing</span></div>{{end}}
      </div>
      {{end}}
    </div>
  </div>
</section>$TPL$,
    suitable_page_types = '["index", "directory", "entity-directory"]'::jsonb,
    updated_at = now()
WHERE function = 'directory-listing';

DO $$
DECLARE
    schema_txt text;
    tpl text;
BEGIN
    SELECT input_schema::text, html_template INTO schema_txt, tpl
    FROM content_components WHERE function = 'directory-listing';

    IF schema_txt IS NULL THEN
        RAISE EXCEPTION '325: directory-listing not found';
    END IF;
    IF schema_txt LIKE '%directory_entries%' THEN
        RAISE EXCEPTION '325: the dangling query.directory_entries source survives';
    END IF;
    IF schema_txt NOT LIKE '%query.business_directory%' THEN
        RAISE EXCEPTION '325: query.business_directory was not written';
    END IF;
    IF schema_txt LIKE '%filter_categories%' THEN
        RAISE EXCEPTION '325: filter_categories field survives — should be removed';
    END IF;

    IF tpl LIKE '%directory-filter-btn%' THEN
        RAISE EXCEPTION '325: filter chrome/JS survives in the template';
    END IF;
    IF tpl NOT LIKE '%{{.website}}%' THEN
        RAISE EXCEPTION '325: website link was not written';
    END IF;
    IF tpl NOT LIKE '%is_claimed%' THEN
        RAISE EXCEPTION '325: is_claimed handling was not written';
    END IF;

    RAISE NOTICE '325 OK: directory-listing binds to query.business_directory';
END $$;

COMMIT;
