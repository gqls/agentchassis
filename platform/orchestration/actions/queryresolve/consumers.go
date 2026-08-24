// FILE: platform/orchestration/actions/queryresolve/consumers.go
//
// PageListConsumerPages answers the question the estate had no answer to
// until bugs_open/384: "which pages on this site render a component whose
// input_schema declares a field fed by a page-IMAGE query source?" — that is,
// which pages hold a stored array whose `image` entries were computed from
// the card/hero join at their last section resolve.
//
// WHY A SHARED LOOKUP, NOT A PER-PRODUCER PAGE NAME
//   Two producers already re-resolve their own consumers by name
//   (render_news_section → news-index, render_directory → index). A third
//   hard-coded name would be the third spelling of "who consumes my data",
//   and the one that goes stale when a site adds a second listing page. The
//   consumer set is derivable from content_components.input_schema, so it is
//   derived — once, here — and every producer asks the same question.
//
// WHAT IS EXCLUDED, AND WHY
//   - pages with rebuild_policy='owned': page-rerender's reasoned branch runs
//     save_sections, whose ownership refusal (bugs_open/208, OWNED_PAGE_GUARD)
//     fails the run. Mirrors get_pages_to_build_actions.go's
//     ownedPageExclusionSQL at selection time, which bugs_open/333's lane
//     confirmed is the only place a per-BRANCH refusal can be expressed —
//     the per-agent owned-page door cannot see spec.reason.
//   - removed component rows and inactive pages: nothing renders them.
//   - fields whose source is a query.* base that does NOT read page images
//     (news_archive, the directory kinds, …): their arrays are refreshed by
//     their own producers and a card landing cannot change them.
//   - components whose html_template never RENDERS `.image` (council round
//     c2873f56, guardian seat, 2026-08-24). A consumer that stores the image
//     and never shows it has a stale-but-invisible array; re-resolving it on
//     every landing is a page re-render for no visible change. Measured the
//     same day: loancalculator.co.uk has 26 consumer pages by schema and 1 that
//     renders the image — the other 25 are `tool-cta` strips (58 instances
//     fleet-wide, 0 rendering). With the filter the per-site count is 0–3.
//     When such a template is later changed to render the image, the
//     template_changed re-render re-resolves the array anyway (REB-002).
//
// The LIKE pre-filter is a cheap narrowing only; the decision is made in Go
// against datahelpers.SchemaContentFields (both schema dialects) and
// SourceReadsPageImages (the declared set beside queryHandlers). The
// renders-image predicate is SQL (`html_template ~ '\.image\y'`): `{{.image}}`,
// `{{if .image}}`, `{{$it.image}}` all match; `.image_url` does not (`\y` is a
// word boundary and `_` is a word character).

package queryresolve

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ConsumerPage is one page on a site that renders at least one component
// whose input_schema declares a field fed by a page-image query source.
type ConsumerPage struct {
	ID     uuid.UUID
	Name   string
	URL    string
	Domain string
	// Fields lists every consuming (component, field, source) on the page,
	// sorted by component then field, so callers and tests are deterministic.
	Fields []ConsumerField
}

// ConsumerField is one declared field that consumes a page-image source.
type ConsumerField struct {
	Component string // content_components.name
	Field     string // key in input_schema.fields
	Source    string // the full declaration, e.g. "query.blog_posts"
}

// Sources returns the distinct source declarations on the page, sorted.
func (p ConsumerPage) Sources() []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(p.Fields))
	for _, f := range p.Fields {
		if !seen[f.Source] {
			seen[f.Source] = true
			out = append(out, f.Source)
		}
	}
	sort.Strings(out)
	return out
}

// Queryer is the slice of *sql.DB / *sql.Tx this lookup needs, so it can run
// inside a producer's transaction (flag_page_image_rebuild) or on a bare DB
// (derive_card_asset) without two copies of the query.
type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// pageListConsumerSQL is the ONE definition of the candidate set. Alias `p`
// for pages, `pc` for page_components, `cc` for content_components.
const pageListConsumerSQL = `
	SELECT p.id, p.name, COALESCE(p.url, ''), s.domain, cc.name, cc.input_schema::text
	  FROM page_components pc
	  JOIN pages p  ON p.id = pc.page_id
	  JOIN sites s  ON s.id = p.site_id
	  JOIN content_components cc ON cc.id = pc.component_id
	 WHERE p.site_id = $1
	   AND p.status IN ('active', 'deployed')
	   AND pc.build_status <> 'removed'
	   AND COALESCE(p.rebuild_policy, 'generic') <> 'owned'
	   AND cc.input_schema IS NOT NULL
	   AND cc.input_schema::text LIKE '%query.%'
	   AND cc.html_template ~ '\.image\y'
	 ORDER BY p.name, cc.name`

// PageListConsumerPages returns the site's pages that consume a page-image
// query source, each with the fields that consume it. A component whose
// schema cannot be parsed is logged and skipped — one malformed row must not
// hide every other consumer on the site.
func PageListConsumerPages(ctx context.Context, q Queryer, siteID uuid.UUID, logger *zap.Logger) ([]ConsumerPage, error) {
	if q == nil {
		return nil, fmt.Errorf("PageListConsumerPages: nil queryer")
	}
	if siteID == uuid.Nil {
		return nil, fmt.Errorf("PageListConsumerPages: empty site_id")
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	rows, err := q.QueryContext(ctx, pageListConsumerSQL, siteID)
	if err != nil {
		return nil, fmt.Errorf("PageListConsumerPages query failed: %w", err)
	}
	defer rows.Close()

	byPage := map[uuid.UUID]*ConsumerPage{}
	var order []uuid.UUID
	for rows.Next() {
		var (
			id                                uuid.UUID
			name, url, domain, component, raw string
		)
		if err := rows.Scan(&id, &name, &url, &domain, &component, &raw); err != nil {
			return nil, fmt.Errorf("PageListConsumerPages scan failed: %w", err)
		}
		var schema map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &schema); err != nil {
			logger.Warn("PageListConsumerPages: component input_schema is not a JSON object — skipped, its page may be under-counted",
				zap.String("component", component), zap.String("page", name), zap.Error(err))
			continue
		}
		fields, ok, fromLegacy := datahelpers.SchemaContentFields(schema)
		if !ok {
			continue
		}
		if fromLegacy {
			datahelpers.WarnLegacyDialect(logger, "PageListConsumerPages", component)
		}
		for fieldName, defRaw := range fields {
			def, ok := defRaw.(map[string]interface{})
			if !ok {
				continue
			}
			src, _ := def["source"].(string)
			src = strings.TrimSpace(src)
			lower := strings.ToLower(src)
			if !strings.HasPrefix(lower, "query.") {
				continue
			}
			if !SourceReadsPageImages(strings.TrimPrefix(lower, "query.")) {
				continue
			}
			page, seen := byPage[id]
			if !seen {
				page = &ConsumerPage{ID: id, Name: name, URL: url, Domain: domain}
				byPage[id] = page
				order = append(order, id)
			}
			page.Fields = append(page.Fields, ConsumerField{Component: component, Field: fieldName, Source: src})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("PageListConsumerPages rows iter failed: %w", err)
	}

	out := make([]ConsumerPage, 0, len(order))
	for _, id := range order {
		page := byPage[id]
		sort.Slice(page.Fields, func(i, j int) bool {
			if page.Fields[i].Component != page.Fields[j].Component {
				return page.Fields[i].Component < page.Fields[j].Component
			}
			return page.Fields[i].Field < page.Fields[j].Field
		})
		out = append(out, *page)
	}
	return out, nil
}
