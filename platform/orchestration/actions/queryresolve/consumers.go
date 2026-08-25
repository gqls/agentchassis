// FILE: platform/orchestration/actions/queryresolve/consumers.go
//
// ConsumerPages answers the question the estate had no answer to until
// bugs_open/384: "which pages on this site render a component whose
// input_schema declares a field fed by a query source that reads X?" — that is,
// which pages hold a stored array that goes stale when X changes.
//
// GENERALISED 2026-08-25 (RFC_052, owner ruling "generalise it now"). It shipped
// image-only: one boolean per source ("reads page images") and one hard-coded
// `.image` template test. That answered one producer's question and, as the
// council's architecture seat put it on the record (round c2873f56), risked
// leaving "a second ad-hoc dependency-tracking mechanism layered next to the
// query.* source system rather than a designed one". The declaration is now a
// per-source DEPENDENCY SET (see sourceDependencies) and this lookup takes the
// dependency as an argument; page images are one class among several.
//
// WHY A SHARED LOOKUP, NOT A PER-PRODUCER QUERY
//   Two producers already re-resolved their own consumers with their own SQL.
//   > CORRECTED 2026-08-25: this comment, and RFC_052 itself, said they
//   > "hard-code their one consumer page". THEY DO NOT — both select by
//   > COMPONENT FUNCTION (`cc.function IN ('latest-news','news-listing')`;
//   > `profile.SnippetComponent/ListingComponent`), so both already handled a
//   > site with two news pages. What was hard-coded is the COMPONENT SET, which
//   > goes stale the day a component is renamed or a second component starts
//   > consuming the same source — a slower failure than the one claimed, and a
//   > real one. Both now call this lookup; measured before migrating, the two
//   > routes select the SAME pages today (news 16 both / 0 / 0; directory 5 both
//   > / 0 / 0, and per-kind 1/1/2/1 with zero on either side).
//
// WHAT IS EXCLUDED, AND WHY
//   - pages with rebuild_policy='owned': page-rerender's reasoned branch runs
//     save_sections, whose ownership refusal (bugs_open/208, OWNED_PAGE_GUARD)
//     fails the run. Mirrors get_pages_to_build_actions.go's
//     ownedPageExclusionSQL at selection time, which bugs_open/333's lane
//     confirmed is the only place a per-BRANCH refusal can be expressed —
//     the per-agent owned-page door cannot see spec.reason.
//   - removed component rows and inactive pages: nothing renders them.
//   - pages that never shipped (PageHasShippedPredicateFor): both producers
//     already carried this floor, so the shared lookup must too or migrating
//     onto it would WIDEN them. Measured no-op today; see consumerSQL.
//   - fields whose source does not read the dependency being asked about: their
//     arrays are refreshed by their own producers.
//   - for a dependency that feeds NAMED ITEM KEYS only: components whose
//     html_template never renders one of those keys (council round c2873f56,
//     guardian seat, 2026-08-24). A consumer that stores a value and never shows
//     it has a stale-but-invisible array, and re-resolving it on every landing
//     is a page re-render for no visible change. This filter does NOT apply to a
//     whole-item-set dependency — see consumerSQL, where getting that backwards
//     would silently return nothing for news and directories.
//     ⚠ It also narrows the SWEEP that shares this lookup, not just the event
//     seam — which silenced check_page_list_stale's own motivating case
//     (WRONG_CALLS, 2026-08-25). When such a template is later changed to render
//     the key, the template_changed re-render re-resolves the array anyway
//     (REB-002) — that is exactly what migration 614 did for `tool-cta`.
//
// The LIKE pre-filter is a cheap narrowing only; the decision is made in Go
// against datahelpers.SchemaContentFields (both schema dialects) and
// SourceReads (the declared set beside queryHandlers).

package queryresolve

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
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

// ConsumesAny reports whether this page consumes any of the given query BASES
// (the part after "query.", `:arg` stripped — "model_directory", not
// "query.model_directory:x").
//
// A producer that is narrower than its dependency class needs this. A
// dependency names a STORE; `render_directory` publishes one KIND at a time out
// of that store, and telling a page that lists mortgage lenders that the model
// directory changed would be a re-render for nothing. ConsumerPages answers
// "who reads this store", and this narrows it to "who reads these sources" —
// which is how the two stay honest: the shared lookup is not quietly taught a
// per-caller special case.
func (p ConsumerPage) ConsumesAny(bases ...string) bool {
	want := make(map[string]bool, len(bases))
	for _, b := range bases {
		if b = strings.ToLower(strings.TrimSpace(b)); b != "" {
			want[b] = true
		}
	}
	if len(want) == 0 {
		return false
	}
	for _, f := range p.Fields {
		src := strings.ToLower(strings.TrimSpace(f.Source))
		base, _ := parseQueryName(strings.TrimPrefix(src, "query."))
		if want[base] {
			return true
		}
	}
	return false
}

// Queryer is the slice of *sql.DB / *sql.Tx this lookup needs, so it can run
// inside a producer's transaction (flag_page_image_rebuild) or on a bare DB
// (derive_card_asset) without two copies of the query.
type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// consumerSQL builds the ONE definition of the candidate set, for one
// dependency class. Alias `p` for pages, `pc` for page_components, `cc` for
// content_components.
//
// `p.status IN ('active','deployed')` is the SAME set resolvePagesWhereType and
// resolvePagesUnderSection select from (a consumer page is chosen by the rule
// its own items are chosen by). Kept for PARITY with the resolvers, not because
// both values occur: [MEASURED 2026-08-24] `pages.status` holds `active` 805 and
// `archived` 66 — `deployed` does not occur today, and dropping it here while the
// resolvers keep it would let the two disagree the day it does.
//
// THE HAS-SHIPPED FLOOR was added 2026-08-25 with the generalisation. Both
// producers that are migrating onto this lookup already carried it
// (datahelpers.PageHasShippedPredicateFor), and without it the shared lookup
// would be the WEAKER of the two — so a migration meant to remove drift would
// have introduced some. It can only narrow, and it is a measured no-op today:
// [MEASURED 2026-08-25] 62 page-image consumer pages fleet-wide before and
// after, 0 excluded. `status` and the build columns answer different questions
// and nothing keeps them in step (bugs_open/098), so both are here.
//
// THE TEMPLATE FILTER IS PER-DEPENDENCY, and its absence is meaningful. When
// the dependency feeds NAMED item keys (page images feed `image`), only a
// consumer whose template actually renders one of them is affected — a stored
// but never-rendered value is a re-render for no visible change (council round
// c2873f56, guardian seat). When the dependency has NO key list it governs the
// whole item set, so EVERY consumer is affected and no filter is applied.
// Getting that backwards would silently return nothing: news and directory
// templates render no `.image`, so the page-image filter would exclude all of
// them.
//
// The renders-key predicate is SQL: `\.(image|…)\y` — `{{.image}}`,
// `{{if .image}}`, `{{$it.image}}` all match; `.image_url` does not (`\y` is a
// word boundary and `_` is a word character). Case-insensitive (`~*`) on the
// council's point (round 2005a846, editquality): a template spelling it
// `.Image` would not read the resolver's lowercase key and so renders nothing
// today, but a silently case-sensitive filter is the kind of bound that defeats
// itself without a tell, and the match costs nothing.
func consumerSQL(dep SourceDependency) string {
	sql := `
	SELECT p.id, p.name, COALESCE(p.url, ''), s.domain, cc.name, cc.input_schema::text
	  FROM page_components pc
	  JOIN pages p  ON p.id = pc.page_id
	  JOIN sites s  ON s.id = p.site_id
	  JOIN content_components cc ON cc.id = pc.component_id
	 WHERE p.site_id = $1
	   AND p.status IN ('active', 'deployed')
	   AND ` + datahelpers.PageHasShippedPredicateFor("p") + `
	   AND pc.build_status <> 'removed'
	   AND COALESCE(p.rebuild_policy, 'generic') <> 'owned'
	   AND cc.input_schema IS NOT NULL
	   AND cc.input_schema::text LIKE '%query.%'`

	if keys := DependencyItemKeys(dep); len(keys) > 0 {
		quoted := make([]string, 0, len(keys))
		for _, k := range keys {
			quoted = append(quoted, regexp.QuoteMeta(k))
		}
		sql += `
	   AND cc.html_template ~* '\.(` + strings.Join(quoted, "|") + `)\y'`
	}

	return sql + `
	 ORDER BY p.name, cc.name`
}

// PageListConsumerPages returns the site's pages that consume a page-image
// query source. The page-image special case of ConsumerPages, kept so the
// callers that predate the generalisation read no differently.
func PageListConsumerPages(ctx context.Context, q Queryer, siteID uuid.UUID, logger *zap.Logger) ([]ConsumerPage, error) {
	return ConsumerPages(ctx, q, siteID, DepPageCardImages, logger)
}

// ConsumerPages returns the site's pages that consume a query source reading
// the given dependency class, each with the fields that consume it. A component
// whose schema cannot be parsed is logged and skipped — one malformed row must
// not hide every other consumer on the site.
//
// An UNDECLARED dependency is an error, not an empty result. Returning nothing
// for a name nobody declared is indistinguishable from "this site has no
// consumers", and a producer would read it as "nobody to tell" — which is
// bugs_open/384 itself, one level up.
func ConsumerPages(ctx context.Context, q Queryer, siteID uuid.UUID, dep SourceDependency, logger *zap.Logger) ([]ConsumerPage, error) {
	if q == nil {
		return nil, fmt.Errorf("ConsumerPages: nil queryer")
	}
	if siteID == uuid.Nil {
		return nil, fmt.Errorf("ConsumerPages: empty site_id")
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	if len(SourcesFor(dep)) == 0 {
		return nil, fmt.Errorf("ConsumerPages: no query base declares dependency %q — declare it in sourceDependencies (known: %v)", dep, KnownDependencies())
	}

	rows, err := q.QueryContext(ctx, consumerSQL(dep), siteID)
	if err != nil {
		return nil, fmt.Errorf("ConsumerPages(%s) query failed: %w", dep, err)
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
			return nil, fmt.Errorf("ConsumerPages scan failed: %w", err)
		}
		var schema map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &schema); err != nil {
			logger.Warn("ConsumerPages: component input_schema is not a JSON object — skipped, its page may be under-counted",
				zap.String("component", component), zap.String("page", name), zap.Error(err))
			continue
		}
		fields, ok, fromLegacy := datahelpers.SchemaContentFields(schema)
		if !ok {
			continue
		}
		if fromLegacy {
			datahelpers.WarnLegacyDialect(logger, "ConsumerPages", component)
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
			if !SourceReads(strings.TrimPrefix(lower, "query."), dep) {
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
		return nil, fmt.Errorf("ConsumerPages rows iter failed: %w", err)
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
