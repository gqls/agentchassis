// FILE: platform/orchestration/actions/queryresolve/queryresolve.go
//
// Package queryresolve runs named queries against the database for `query.*`
// source resolution in component input_schemas. Components that show lists
// (tool-list, blog-listing, directory-listing, etc.) declare their items
// field with `source: "query.<name>"`, and this package executes that
// resolution.
//
// Initial scope (v1, focus doc FOCUS_directory_builder_and_list_components.md):
//   - Concrete queries: pages_where_type:<type>, pages_under_section:<area>
//   - Called inline from plan_sections_action.go's source resolver
//   - Returns []map[string]interface{} suitable for the "items" field of an
//     array-typed schema field
//
// The resolver is a separate package so that a future directory-builder
// agent (per doc 002) can call the same Resolve() entry point without
// changing the resolver logic.

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
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// QueryRequest describes a single resolution call.
//
// Name is the suffix after "query." in the source declaration. For
// `source: "query.pages_where_type:tool"`, Name is `pages_where_type:tool`.
//
// SiteID scopes the query to a particular site's data. Every query in this
// package is site-scoped — there are no cross-site queries.
//
// Limit (when > 0) caps the number of rows returned. The component schema
// can specify a limit; if zero, the query's default applies.
type QueryRequest struct {
	Name   string
	SiteID uuid.UUID
	Limit  int
}

// queryHandler is the uniform shape every registered query is adapted to.
// Handlers that ignore arg / siteID / limit simply don't read them.
type queryHandler func(ctx context.Context, db *sql.DB, siteID uuid.UUID, arg string, limit int, logger *zap.Logger) (interface{}, error)

// queryHandlers is the ONE home of the `query.*` vocabulary. Resolve
// dispatches through it and IsKnownQueryName answers from it, so the
// dispatcher and any validator asking "would this name resolve?" cannot
// drift — the drift class two hand-maintained lists always grow into
// (bugs_open/309: 7 query names declared in component schemas that the old
// switch did not know, each failing silently at plan time).
var queryHandlers = map[string]queryHandler{
	"pages_where_type": func(ctx context.Context, db *sql.DB, siteID uuid.UUID, arg string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolvePagesWhereType(ctx, db, siteID, arg, limit, false, logger)
	},

	"pages_under_section": func(ctx context.Context, db *sql.DB, siteID uuid.UUID, arg string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolvePagesUnderSection(ctx, db, siteID, arg, limit, logger)
	},

	"section_index_for": func(ctx context.Context, db *sql.DB, siteID uuid.UUID, arg string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveSectionIndexForType(ctx, db, siteID, arg, logger)
	},

	"products": func(ctx context.Context, db *sql.DB, siteID uuid.UUID, arg string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveProducts(ctx, db, siteID, arg, limit, logger)
	},

	// Homepage news cards (latest-news component). Items come from
	// content_feed_items, not pages — see news_items.go for why the
	// selection is shared with the JSON path and the output is escaped.
	"latest_news": func(ctx context.Context, db *sql.DB, siteID uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveLatestNews(ctx, db, siteID, limit, logger)
	},

	// News-index listing (news-listing component). Same machinery,
	// archive depth and window, topics included.
	"news_archive": func(ctx context.Context, db *sql.DB, siteID uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveNewsArchive(ctx, db, siteID, limit, logger)
	},

	// Homepage model-directory snippet (model-directory component).
	// Deliberately NOT site-scoped — see model_directory_items.go's
	// header: directory_entities/directory_claims is one global
	// registry, siteID is unused here.
	"model_directory": func(ctx context.Context, db *sql.DB, _ uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveModelDirectory(ctx, db, limit, logger)
	},

	// Model-directory listing page (model-directory-listing component).
	// Same registry, listing depth.
	"model_directory_full": func(ctx context.Context, db *sql.DB, _ uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveModelDirectoryFull(ctx, db, limit, logger)
	},

	// Homepage adoption-tracker snippet: organisations adopting AI
	// agents, with their cited ROI/rollout claims. Same global register
	// as the model directory, kind='company' — also not site-scoped.
	"adoption_tracker": func(ctx context.Context, db *sql.DB, _ uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveDirectoryKind(ctx, db, "company", limit, 12, logger)
	},

	// Adoption-tracker listing page. Same register, listing depth.
	"adoption_tracker_full": func(ctx context.Context, db *sql.DB, _ uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveDirectoryKind(ctx, db, "company", limit, 50, logger)
	},

	// Agentic-communication protocols (MCP and successors) and their
	// cited uptake. Deliberately a SEPARATE kind rather than a company
	// attribute: a protocol is adopted BY many companies, so modelling
	// it as a company field would force one row per pairing and lose
	// the protocol's own cited facts (spec version, governance, date).
	"protocol_tracker": func(ctx context.Context, db *sql.DB, _ uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveDirectoryKind(ctx, db, "protocol", limit, 12, logger)
	},

	"protocol_tracker_full": func(ctx context.Context, db *sql.DB, _ uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveDirectoryKind(ctx, db, "protocol", limit, 50, logger)
	},

	// Phase B finance kinds (2026-08-13): three more register kinds on the
	// same global directory_entities/directory_claims registry — the same
	// resolveDirectoryKind the adoption/protocol trackers use, never the
	// bespoke model_* functions. Not site-scoped, like every directory arm.
	"mortgage_lender_directory": func(ctx context.Context, db *sql.DB, _ uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveDirectoryKind(ctx, db, "mortgage-lender", limit, 12, logger)
	},

	"mortgage_lender_directory_full": func(ctx context.Context, db *sql.DB, _ uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveDirectoryKind(ctx, db, "mortgage-lender", limit, 50, logger)
	},

	"savings_provider_directory": func(ctx context.Context, db *sql.DB, _ uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveDirectoryKind(ctx, db, "savings-provider", limit, 12, logger)
	},

	"savings_provider_directory_full": func(ctx context.Context, db *sql.DB, _ uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveDirectoryKind(ctx, db, "savings-provider", limit, 50, logger)
	},

	"health_insurer_directory": func(ctx context.Context, db *sql.DB, _ uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveDirectoryKind(ctx, db, "health-insurer", limit, 12, logger)
	},

	"health_insurer_directory_full": func(ctx context.Context, db *sql.DB, _ uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveDirectoryKind(ctx, db, "health-insurer", limit, 50, logger)
	},

	// GENERIC register arms: `directory:<kind>` and `directory_full:<kind>`.
	//
	// Every literal arm above is the same call with a different string —
	// resolveDirectoryKind(kind, 12|50) — because queryHandlers is a literal
	// map while QueryDirectoryEntries has been kind-generic since Phase A.
	// So a SEVENTH kind needed a Go change for no reason other than the map,
	// and the owner's standing objection (2026-09-04) is exactly that: a new
	// directory kind should not be a code change.
	//
	// These two bases take the kind as the `:arg` the vocabulary already
	// supports (`pages_where_type:tool` is the precedent), so a new kind is
	// now a component declaring `query.directory:<kind>` and nothing else.
	//
	// The literal arms are DELIBERATELY LEFT IN PLACE. They are declared by
	// live components' input_schema on shipped pages; retiring them is a
	// separate migration of those declarations, not this change's business —
	// and a rename that silently stopped resolving would empty a served
	// listing with no error (the bugs_open/453 shape).
	//
	// An unknown kind is NOT an error here: QueryDirectoryEntries refuses an
	// EMPTY kind, and a kind with no rows returns an empty list, which is
	// what a component's `on_missing` handling already exists for. Refusing
	// unknown-but-nonempty kinds would mean this map knowing the kind list,
	// which is the coupling being removed.
	"directory": func(ctx context.Context, db *sql.DB, _ uuid.UUID, arg string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveDirectoryKind(ctx, db, arg, limit, 12, logger)
	},
	"directory_full": func(ctx context.Context, db *sql.DB, _ uuid.UUID, arg string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveDirectoryKind(ctx, db, arg, limit, 50, logger)
	},

	// A site's own verified business_intel directory (bugs_open/206).
	// No arg: the vertical is looked up from the site's own
	// directory-export-json config, not a static parameter — see
	// business_directory.go's header for why.
	"business_directory": func(ctx context.Context, db *sql.DB, siteID uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveBusinessDirectory(ctx, db, siteID, limit, logger)
	},

	// Article listings (content-listing, blog-listing components declare
	// `source: "query.blog_posts"`). Fleet convention: articles are pages
	// with page_type 'blog-post', so this is pages_where_type with the
	// type fixed — one vocabulary entry, zero new query machinery.
	//
	// listedOnly (F2.1, 2026-07-17): article listings additionally demand
	// the page actually shipped — plan-era scaffold rows and never-built
	// duplicates sit status='active' and were being listed as 404 links
	// (robot-hands served 6 of them at the D13 gate).
	"blog_posts": func(ctx context.Context, db *sql.DB, siteID uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolvePagesWhereType(ctx, db, siteID, "blog-post", limit, true, logger)
	},

	// Dated, correctable event facts (bugs_open/427) — a site's own
	// evidence_base register, not content_feed_items directly: the feed is
	// the SOURCE the facts are extracted from, but the register is the
	// store this resolver reads, the same way news items are read from
	// content_feed_items rather than re-scraped per page. See
	// upcoming_events.go for why a query source and not a new action.
	"upcoming_events": func(ctx context.Context, db *sql.DB, siteID uuid.UUID, _ string, limit int, logger *zap.Logger) (interface{}, error) {
		return resolveUpcomingEvents(ctx, db, siteID, limit, logger)
	},
}

// SourceDependency names a CLASS OF DATA that a query base reads — the thing a
// producer changes and then has to tell the consumers about.
//
// GENERALISED 2026-08-25 (RFC_052, owner ruling "generalise it now"). Until then
// this was a single boolean per base, `pageImageSources`, meaning "reads the
// page card/hero join". That answered exactly one producer's question. The
// council's architecture seat objected to it on the record (round c2873f56):
// the estate had acquired its first GENERAL "who consumes this data source"
// derivation, and building further sources on a boolean's shape would leave a
// second ad-hoc dependency-tracking mechanism beside the query.* system rather
// than a designed one. This is the designed one: page images are now one
// dependency class among several, and the old boolean is a wrapper over it.
type SourceDependency string

const (
	// DepPageCardImages: items carry an `image` computed from the page's CARD
	// asset (assets.entity_type='page', purpose='card') with the current plan
	// hero as fallback. Produced by derive_card_asset and flag_page_image_rebuild.
	DepPageCardImages SourceDependency = "page_card_images"
	// DepFeedItems: items come from content_feed_items. Produced by render_news_section.
	DepFeedItems SourceDependency = "content_feed_items"
	// DepDirectoryEntities: items come from directory_entities. Produced by render_directory.
	DepDirectoryEntities SourceDependency = "directory_entities"
	// DepBusinessIntel: items come from business_intel. NOT the same store as
	// DepDirectoryEntities — `business_directory` reads business_intel while the
	// model/adoption/protocol/lender trackers read directory_entities, and
	// declaring them as one class would tell the wrong consumers on every publish.
	DepBusinessIntel SourceDependency = "business_intel"
	// DepProducts: items come from the products table. No producer notifies today.
	DepProducts SourceDependency = "products"
	// DepEvidenceBase: items come from a site's site_specs aspect='evidence_base'
	// register (bugs_open/427). Produced by refresh_evidence_base (a citation
	// fact's daily re-verification) and, per-fact, by whatever registers a new
	// dated event fact (news_feed_ingestion's verify_and_register_citations
	// extension).
	//
	// Literal is "site_specs.evidence_base", not the bare "evidence_base" this
	// constant shipped with initially (council REVISE 08f56b7e, guardian +
	// architecture, both low severity but "settle before merge, not left to
	// reviewer preference"): every sibling literal (content_feed_items,
	// directory_entities, business_intel, products) names a real TABLE this
	// dependency reads; evidence_base is not a table, it is one `aspect` value
	// on the shared `site_specs` table, and a bare "evidence_base" would read
	// as if it were its own store to the next person adding a fifth class.
	DepEvidenceBase SourceDependency = "site_specs.evidence_base"
)

// sourceDependencies declares, per query base, WHICH dependency classes its
// resolver reads and WHICH ITEM KEYS each class feeds.
//
// THE ITEM-KEY LIST IS THE LOAD-BEARING GENERALISATION, and it is what lets one
// lookup serve every producer. A dependency that feeds NAMED KEYS (page images
// feed exactly `image`) only matters to a consumer whose template renders one
// of those keys — re-resolving a component that stores the image and never
// shows it is a page re-render for no visible change, which is the bound the
// council added in round 2005a846. A dependency with NO key list governs the
// whole item SET — its membership, order and contents — so EVERY consumer is
// affected and no template filter applies at all. That is why news and
// directory listings need no `.image` test: a new feed item changes which items
// exist, not one field of one item.
//
// PINNED TO THE RESOLVERS BY BEHAVIOUR, NOT BY COMMENT:
// TestSourceDependenciesMatchTheResolvers drives every registered handler
// against a recording sqlmock and checks which store each one's SQL actually
// touches, in BOTH directions. A resolver that starts reading a store without
// declaring it fails; so does a stale declaration. And every registered base
// must appear here — a base with genuinely no notifiable dependency declares an
// EMPTY map rather than being absent, so "nobody thought about it" and "there
// is nothing to think about" stay distinguishable.
//
// Why it exists at all (bugs_open/384): a component field fed by one of these
// sources stores its resolved items in page_components.content_data, and every
// assemble-mode re-render re-ships that stored array verbatim. When the data
// behind the source changes, the pages that consume it must be told to
// re-resolve, and a producer has no other way to learn which sources — and via
// ConsumerPages, which pages — those are.
var sourceDependencies = map[string]map[SourceDependency][]string{
	// Page listings: the card/hero join feeds exactly one item key.
	"pages_where_type":    {DepPageCardImages: {"image"}},
	"pages_under_section": {DepPageCardImages: {"image"}},
	"blog_posts":          {DepPageCardImages: {"image"}},

	// News: a fresh feed item changes the SET, so every consumer is affected.
	"latest_news":  {DepFeedItems: nil},
	"news_archive": {DepFeedItems: nil},

	// Directories: same reasoning, different store.
	// Generic register arms — keyed by BASE, so one entry covers every
	// `directory:<kind>` / `directory_full:<kind>` (SourceReads parses the
	// base before looking up). Without these two a page fed by the generic
	// arm would never be told to re-resolve when its register changed —
	// silently stale, which is bugs_open/384's whole subject.
	"directory":                       {DepDirectoryEntities: nil},
	"directory_full":                  {DepDirectoryEntities: nil},
	"model_directory":                 {DepDirectoryEntities: nil},
	"model_directory_full":            {DepDirectoryEntities: nil},
	"adoption_tracker":                {DepDirectoryEntities: nil},
	"adoption_tracker_full":           {DepDirectoryEntities: nil},
	"protocol_tracker":                {DepDirectoryEntities: nil},
	"protocol_tracker_full":           {DepDirectoryEntities: nil},
	"mortgage_lender_directory":       {DepDirectoryEntities: nil},
	"mortgage_lender_directory_full":  {DepDirectoryEntities: nil},
	"savings_provider_directory":      {DepDirectoryEntities: nil},
	"savings_provider_directory_full": {DepDirectoryEntities: nil},
	"health_insurer_directory":        {DepDirectoryEntities: nil},
	"health_insurer_directory_full":   {DepDirectoryEntities: nil},

	// A different store, despite the name (see DepBusinessIntel).
	"business_directory": {DepBusinessIntel: nil},

	"products": {DepProducts: nil},

	// A new event fact changes the SET (which fixtures exist, in what
	// order) — same reasoning as news/directories, different store.
	"upcoming_events": {DepEvidenceBase: nil},

	// DECLARED, AND DECLARED TO HAVE NONE — not an oversight. section_index_for
	// returns a URL, not an item array, so nothing of it is stored as a snapshot
	// that can go stale in the way this mechanism exists to fix. It reads `pages`
	// like everything else here; that is not a notifiable dependency.
	"section_index_for": {},
}

// SourceReads reports whether a `query.*` source — the part after "query.",
// optional `:arg` included, the same string QueryRequest.Name carries — reads
// the given dependency class. Same normalisation as Resolve, answered from the
// same base.
func SourceReads(name string, dep SourceDependency) bool {
	base, _ := parseQueryName(strings.ToLower(strings.TrimSpace(name)))
	deps, ok := sourceDependencies[base]
	if !ok {
		return false
	}
	_, reads := deps[dep]
	return reads
}

// SourceReadsPageImages is the page-image special case of SourceReads, kept so
// the callers that predate the generalisation read no differently.
func SourceReadsPageImages(name string) bool {
	return SourceReads(name, DepPageCardImages)
}

// SourcesFor returns the query bases that read the given dependency, sorted.
func SourcesFor(dep SourceDependency) []string {
	bases := make([]string, 0, len(sourceDependencies))
	for base, deps := range sourceDependencies {
		if _, reads := deps[dep]; reads {
			bases = append(bases, base)
		}
	}
	sort.Strings(bases)
	return bases
}

// PageImageSources returns the declared page-image bases, sorted, for messages
// and tests. Retained wrapper; SourcesFor is the general form.
func PageImageSources() []string {
	return SourcesFor(DepPageCardImages)
}

// DependencyItemKeys returns the item keys the dependency feeds, sorted and
// de-duplicated across every base that reads it. EMPTY means the dependency
// governs the whole item set rather than named fields — see the note on
// sourceDependencies, and note the difference matters: an empty list makes
// ConsumerPages drop its template filter entirely instead of matching nothing.
func DependencyItemKeys(dep SourceDependency) []string {
	seen := map[string]bool{}
	var keys []string
	for _, deps := range sourceDependencies {
		for _, k := range deps[dep] {
			if !seen[k] {
				seen[k] = true
				keys = append(keys, k)
			}
		}
	}
	sort.Strings(keys)
	return keys
}

// KnownDependencies returns every declared dependency class, sorted.
func KnownDependencies() []SourceDependency {
	seen := map[SourceDependency]bool{}
	var out []SourceDependency
	for _, deps := range sourceDependencies {
		for d := range deps {
			if !seen[d] {
				seen[d] = true
				out = append(out, d)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Resolve dispatches the request to a registered query handler. Returns
// nil + error on unknown query name or DB failure. Returns the resolved
// data on success — currently always []map[string]interface{} but the
// signature is interface{} to allow future query types that return other
// shapes (e.g. a single object, a count).
func Resolve(ctx context.Context, db *sql.DB, req QueryRequest, logger *zap.Logger) (interface{}, error) {
	if db == nil {
		return nil, fmt.Errorf("queryresolve.Resolve: nil db")
	}
	if req.SiteID == uuid.Nil {
		return nil, fmt.Errorf("queryresolve.Resolve: empty site_id")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("queryresolve.Resolve: empty query name")
	}

	// Normalise: lowercase, trim whitespace. Query names are case-insensitive
	// in source declarations to be forgiving (component-creator's prompt may
	// vary capitalisation).
	name := strings.ToLower(strings.TrimSpace(req.Name))

	// Parse the name into base + parameter. The vocabulary uses one of two
	// shapes:
	//   pages_where_type:tool        → base "pages_where_type", arg "tool"
	//   pages_under_section:guides   → base "pages_under_section", arg "guides"
	//   site_specs.case_studies      → not a query (handled by site_specs source)
	//
	// Names without `:` are treated as the whole name being the base, with
	// no argument.
	base, arg := parseQueryName(name)

	handler, ok := queryHandlers[base]
	if !ok {
		return nil, fmt.Errorf("queryresolve.Resolve: unknown query name %q (base %q)", req.Name, base)
	}
	return handler(ctx, db, req.SiteID, arg, req.Limit, logger)
}

// IsKnownQueryName reports whether a `query.*` source declaration would
// dispatch to a registered handler. `name` is the part after "query." —
// the same string QueryRequest.Name carries, optional `:arg` included.
// Same normalisation as Resolve, and answered from the same map, so this
// cannot say yes to a name Resolve would refuse.
func IsKnownQueryName(name string) bool {
	base, _ := parseQueryName(strings.ToLower(strings.TrimSpace(name)))
	_, ok := queryHandlers[base]
	return ok
}

// KnownQueryBases returns the registered query bases, sorted, for use in
// validation messages that must name the real vocabulary.
func KnownQueryBases() []string {
	bases := make([]string, 0, len(queryHandlers))
	for base := range queryHandlers {
		bases = append(bases, base)
	}
	sort.Strings(bases)
	return bases
}

// parseQueryName splits a query name into base and argument. The first
// colon is the separator; subsequent colons are part of the argument
// (rarely needed but harmless).
func parseQueryName(name string) (base, arg string) {
	idx := strings.Index(name, ":")
	if idx < 0 {
		return name, ""
	}
	return name[:idx], name[idx+1:]
}

// ListedPageEligibilitySQL is the ONE definition of "this page actually
// shipped" — a WHERE-clause fragment requiring the alias `p` for pages.
//
// It exists as a shared constant, not as two hand-copied strings, because the
// article listing and the imagery sweep that decides which articles get images
// MUST agree on which articles exist. If they drift, the failure is silent in
// both directions: image-generation credits spent on pages nobody can reach,
// or listed pages that never get an image. That is the same shape as the
// dedup-index ↔ workItemTerminalStatuses contract this platform has already
// paid for once, so the two consumers (here and
// discovery_checks.ContentImageMissingCheck) share the literal.
//
// jsonb_typeof guards the length call: jsonb_array_length raises a Postgres
// ERROR on an object-shaped value — which would abort the whole sweep for a
// site rather than skip one bad row — and returns NULL (silently falsy, a
// silent-drop) when sections is NULL. Every one of the 269 live pages is
// array-shaped today; the guard is here so one malformed row can never take a
// site's listing or imagery loop down with it.
const ListedPageEligibilitySQL = `
		  AND p.deployed_at IS NOT NULL
		  AND jsonb_typeof(p.sections) = 'array'
		  AND jsonb_array_length(p.sections) > 0`

// DeployedPageEligibilitySQL is the weaker sibling for page types whose
// content is NOT in `sections` — tool pages are the case that forced it:
// their substance is the interactive tool committed under /tools/<name>/, and
// 20 of the fleet's 33 deployed tool pages carry zero sections legitimately.
// Requiring sections there would exclude almost every real tool page.
//
// This constant exists for consumers that must not spend money on pages that
// never shipped — today, the imagery sweep. It is deliberately conservative:
// a false negative (skipping a shipped-but-unstamped page) only costs one
// missing image, self-corrects when deployed_at is stamped, and so `deployed_at
// IS NOT NULL` alone is the right floor for a spend guard. Same alias contract
// (`p` for pages) as ListedPageEligibilitySQL.
//
// Listings need a DIFFERENT floor — see FetchablePageEligibilitySQL — because
// their error costs are asymmetric the other way: a false negative delists a
// working page. This constant used to say it was "deliberately NOT applied by
// resolvePagesWhereType" so listings kept a "looser contract" and advertised
// the whole tool directory. That reasoning was bugs_open/052: the loose
// contract advertised never-built pages that 404.
const DeployedPageEligibilitySQL = `
		  AND p.deployed_at IS NOT NULL`

// FetchablePageEligibilitySQL is the floor EVERY page-listing derivation must
// carry: a listing must never advertise a page that would 404 (bugs_open/052 —
// a tool list re-derived from the page set on every render and re-advertised
// two never-built, 404 tool cards; bugs_open/023's derived-field family). Alias
// contract (`p` for pages) as above.
//
// The predicate is "did this page ship / will it serve", validated fleet-wide
// against live HTTP (bugs_open/049 Correction 2, 052 addendum): a page keeps
// serving its old artefact once deployed_at is stamped even after it is flagged
// `needs_rebuild`, so `deployed_at IS NOT NULL` keeps those (the 200s). The
// `OR build_status = 'deployed'` disjunct is load-bearing, not polish: it keeps
// the one fleet page that is `deployed` yet never stamped (idea.uk's
// tool-audience-check, a bugs_open/040 shape — it serves 200 and is a real
// `page_type='tool'` page, so dropping it would delist a working tool, "worse
// than the bug"). A plain `deployed_at IS NOT NULL` floor would do exactly that.
//
// It is a strict superset of what a never-built page passes: `planned` and
// `needs_rebuild` pages that were never deployed have deployed_at NULL and
// build_status <> 'deployed', so both disjuncts reject them — the 404s go.
//
// Stronger than DeployedPageEligibilitySQL (adds the unstamped-deployed keep)
// and weaker than ListedPageEligibilitySQL (does not require `sections`, since
// tool pages legitimately have none). The three are deliberately distinct; each
// comment says why, so the split does not read as accidental drift.
//
// DERIVED, not spelled (bugs_open/185 fix candidate 2, 2026-08-15): the judgement
// here is "has this page shipped", and the estate has ONE definition of that —
// datahelpers.PageHasShippedPredicateFor, i.e. NOT(deployed_at IS NULL AND
// COALESCE(build_status, <empty>) <> 'deployed'). By De Morgan that is exactly the
// old hand-written `(p.deployed_at IS NOT NULL OR p.build_status = 'deployed')`,
// and it was proved so against production before the respell (643/643 pages,
// zero symmetric difference; the idea.uk unstamped keep survives because its
// build_status IS 'deployed'). Until then this constant and the datahelpers
// builder cross-referenced each other by COMMENT only, which is how a
// deliberate split becomes accidental drift on a tree this many sessions
// share. Now a change to the canonical builder reaches every listing floor
// without anyone remembering to read the other file. A `var` because Go cannot
// call a function in a const initialiser; nothing consumes it in a const
// context (checked: two uses, both string concatenation in this file).
var FetchablePageEligibilitySQL = `
		  AND ` + datahelpers.PageHasShippedPredicateFor("p")

// pageListEligibilitySQL returns the WHERE fragment (alias `p`) a page-listing
// query must carry so it never advertises a 404. Extracted as a function so the
// "the generic listing path always has a build-state floor" invariant is
// unit-testable without a database: before bugs_open/052 the listedOnly=false
// path carried NO build-state filter (eligibility was ""), which is precisely
// how tool/game/guide/entity-page listings advertised never-built pages. The
// listedOnly path is the stricter article contract, which already implies the
// fetchability floor.
func pageListEligibilitySQL(listedOnly bool) string {
	if listedOnly {
		return ListedPageEligibilitySQL
	}
	return FetchablePageEligibilitySQL
}

// PageImageProjectionSQL / PageImageJoinsSQL are the shared SQL fragments that give
// every page-listing query its item image (Phase I3, Lane B). Two candidates,
// in preference order:
//   - ca: the page's entity-linked CARD asset (assets.entity_type='page' +
//     entity_id, purpose 'card') — the purpose-built listing crop.
//   - ha: the page's own Lane A plan hero (current plan, page scope) — the
//     always-present fallback until a card is derived.
//
// Alias `p` for pages is required in the enclosing query. The lateral hero
// lookup runs per returned row; the resolvers below cap at 24 rows so this
// stays cheap — a caller that splices these fragments into an UNCAPPED query
// pays the lateral per row (rebuild_blog_listing does; see its note).
//
// EXPORTED 2026-08-25 (bugs_open/384 decision 3). They were package-private
// while the resolvers here were the only readers. They are not any more:
// `rebuild_blog_listing` derives the SAME article set for the SAME blog page
// and was hand-writing `"image": ""` for every article, which made it a second
// writer of a field the 384 seam exists to keep correct. Sharing the fragments
// is the same remedy this package already applies to the eligibility floor
// (ListedPageEligibilitySQL): one definition, so the two cannot disagree about
// what a listing item's image IS.
const PageImageProjectionSQL = `
		    COALESCE(ca.asset_key, '') AS card_key,
		    COALESCE(ha.asset_key, '') AS hero_key,
		    COALESCE(ha.purpose, '')   AS hero_purpose`

const PageImageJoinsSQL = `
		LEFT JOIN assets ca
		  ON ca.site_id = p.site_id AND ca.entity_type = 'page'
		 AND ca.entity_id = p.id AND ca.purpose = 'card' AND ca.status = 'active'
		LEFT JOIN LATERAL (
		    SELECT a.asset_key, a.purpose
		      FROM site_plan_imagery spi
		      JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
		      JOIN assets a ON a.site_id = p.site_id AND a.asset_key = spi.key AND a.status = 'active'
		     WHERE sp.site_id = p.site_id AND spi.kind = 'hero'
		       AND spi.scope = 'page' AND spi.scope_ref = p.name
		     LIMIT 1
		) ha ON true`

// PageListingHardCap is the maximum number of items any page-listing query
// returns, and the bound PageImageJoinsSQL's per-row LATERAL hero lookup was
// designed around.
//
// EXPORTED 2026-08-25 (council round 170147b4, guardian): rebuild_blog_listing
// splices the projection into an UNCAPPED query, so it paid the lateral per row
// AND could disagree with the resolver about the same listing. Measured that
// day: webdesign.co.uk has 40 eligible blog posts, so the resolver would
// produce 24 items and the action 40 — a 16-item divergence between two writers
// of one listing, which is the drift class this file already shares
// ListedPageEligibilitySQL to avoid. One definition, both callers.
const PageListingHardCap = 24

// PageImageCols carries the scanned image candidates for one listing row.
type PageImageCols struct {
	CardKey     string
	HeroKey     string
	HeroPurpose string
}

// WebPath resolves the item's image to a deployed git path: card first, plan
// hero second, empty when the page has neither. Never assets.url — that holds
// an expiring presigned S3 URL.
//
// The card-first/hero-second preference is the WHOLE contract, and a caller
// that wants only the purpose-built crop cannot express that here — measured
// 2026-08-25, loancalculator.co.uk has 0 of 10 tool pages with a card, so its
// listing items resolve to full-bleed page heroes. Changing that means a
// card-only key in the item shape, which is a change to the shared query.*
// seam, not a tweak here.
func (c PageImageCols) WebPath() string {
	if c.CardKey != "" {
		return storage.DeployedWebPath(c.CardKey, "card")
	}
	if c.HeroKey != "" {
		return storage.DeployedWebPath(c.HeroKey, c.HeroPurpose)
	}
	return ""
}

// resolvePagesWhereType returns all pages on the site with the given
// page_type, projected to the standard list-item shape.
//
// Standard list-item shape (consumed by tool-list, blog-listing, etc.):
//
//	{
//	  "name":             "tool-jump-physics",
//	  "title":            "Jump Physics",
//	  "url":              "/tools/jump-physics/index.html",
//	  "meta_description": "...",
//	  "excerpt":          "...",
//	  "nav_label":        "Jump Physics Architect",
//	  "image":            "/assets/images/card-tool-jump-physics.jpg"
//	}
//
// title is the DISPLAY headline — the page's document title with its trailing
// " | <site>" suffix stripped by ListItemTitle. The example above has always
// shown it unsuffixed; until 2026-09-02 the code returned pages.title raw, so
// every card built from this shape displayed the site name inside its own
// headline (bugs_open/425). [MEASURED 2026-09-02] all 19 active components that
// render a range .title render it as visible card text and none of them emits a
// <title> element, so nothing here wants the document title.
//
// excerpt is the one-sentence deck, projected from meta_description and bounded
// by ListItemExcerpt. It is ADDITIVE — meta_description stays, unbounded, for
// any consumer that wants the whole thing. It exists because the component
// library already renders `.excerpt` and nothing on this path ever wrote it:
// the slot rendered as an empty <p> that still took layout, which is read as a
// design fault rather than as the data gap it is.
//
// image (Phase I3, Lane B) is the page's entity-linked CARD asset when one
// exists, else the page's own plan hero (heavier, but present — the card
// supersedes it as derivations land), else "" (components treat a missing
// image as no-thumbnail).
//
// The shape is intentionally generic — components can pick any subset of
// these keys via their `items.items.X.source: "field.X"` declarations
// (handled at template-render time, not here).
//
// Filter: status IN ('active', 'deployed') AND pageListEligibilitySQL. The
// status filter alone does NOT keep unbuilt pages out — a planned, never-built
// page sits status='active', 404s, and was still listed (bugs_open/052). The
// eligibility floor (FetchablePageEligibilitySQL for the generic base) is what
// excludes pages that would 404 while keeping needs_rebuild pages that were
// deployed once and still serve.
//
// listedOnly (F2.1, 2026-07-17) is the STRICTER article contract: it requires
// deployed_at set AND non-empty sections — the page really shipped with real
// content. Used by the blog_posts base: plan-era scaffold rows and never-built
// duplicates pass the status filter and were listed as 404 links. MUST stay in
// lockstep with check_content_image_missing's sweep predicate — the listing
// and the imagery sweep must agree on which articles exist. The generic base
// uses the weaker FetchablePageEligibilitySQL instead (no `sections` demand):
// tool pages are deployed shells whose sections may legitimately be empty.
//
// Note on in_header: we deliberately do NOT filter on in_header here.
// Tool/blog/entity detail pages typically have in_header=false (they are
// individual items, not nav destinations) but they SHOULD appear in their
// parent index's list. The page_type filter is the correct gate; in_header
// is for navigation, not for listings.
func resolvePagesWhereType(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	pageType string,
	limit int,
	listedOnly bool,
	logger *zap.Logger,
) (interface{}, error) {
	if pageType == "" {
		return nil, fmt.Errorf("resolvePagesWhereType: empty page_type argument")
	}

	// Hard cap. Components rarely want more than 12 items in a list; cap at
	// 24 to stop runaway queries. If the schema asks for more, we honour up
	// to 24.
	const hardCap = PageListingHardCap
	if limit <= 0 {
		limit = 12
	}
	if limit > hardCap {
		limit = hardCap
	}

	eligibility := pageListEligibilitySQL(listedOnly)
	rows, err := db.QueryContext(ctx, `
		SELECT
		    p.name,
		    COALESCE(p.title, p.name)               AS title,
		    p.url,
		    COALESCE(p.meta_description, '')        AS meta_description,
		    COALESCE(p.nav_label, p.title, p.name)  AS nav_label,
		    `+PageImageProjectionSQL+`
		FROM pages p
		`+PageImageJoinsSQL+`
		WHERE p.site_id   = $1
		  AND p.page_type = $2
		  AND p.status   IN ('active', 'deployed')`+eligibility+`
		ORDER BY COALESCE(p.nav_order, 100), p.name
		LIMIT $3
	`, siteID, pageType, limit)
	if err != nil {
		return nil, fmt.Errorf("resolvePagesWhereType query failed: %w", err)
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var name, title, url, metaDesc, navLabel string
		var img PageImageCols
		if err := rows.Scan(&name, &title, &url, &metaDesc, &navLabel, &img.CardKey, &img.HeroKey, &img.HeroPurpose); err != nil {
			logger.Warn("resolvePagesWhereType: scan failed", zap.Error(err))
			continue
		}
		items = append(items, map[string]interface{}{
			"name":             name,
			"title":            ListItemTitle(title),
			"url":              url,
			"meta_description": metaDesc,
			"excerpt":          ListItemExcerpt(metaDesc),
			"nav_label":        navLabel,
			"image":            img.WebPath(),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("resolvePagesWhereType rows iter failed: %w", err)
	}

	logger.Info("queryresolve: resolved pages_where_type",
		zap.String("page_type", pageType),
		zap.String("site_id", siteID.String()),
		zap.Int("count", len(items)),
		zap.Int("limit", limit),
	)

	return items, nil
}

// resolvePagesUnderSection returns all pages that belong to the named site
// area (section), projected to the same standard list-item shape as
// resolvePagesWhereType. The <area> argument matches site_areas.name (the
// per-site unique key) via pages.site_area_id; url_prefix is a forgiving
// fallback so `query.pages_under_section:guides` resolves whether the area was
// keyed "guides" or "/guides".
//
// Filter: status IN ('active', 'deployed') AND FetchablePageEligibilitySQL —
// the same build-state floor as the generic resolvePagesWhereType, so a section
// listing cannot advertise a never-built page any more than a type listing can
// (bugs_open/052; the status filter alone does not stop a planned page). This
// resolver has no live consumer in the active schema today, but it is on the
// plan path and is the same page-set derivation, so it gets the same floor.
// in_header is deliberately NOT filtered — item pages under a section often
// have in_header=false but still belong in the section's listing (same
// reasoning as resolvePagesWhereType).
func resolvePagesUnderSection(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	area string,
	limit int,
	logger *zap.Logger,
) (interface{}, error) {
	if area == "" {
		return nil, fmt.Errorf("resolvePagesUnderSection: empty section argument")
	}

	const hardCap = PageListingHardCap
	if limit <= 0 {
		limit = 12
	}
	if limit > hardCap {
		limit = hardCap
	}

	rows, err := db.QueryContext(ctx, `
		SELECT
		    p.name,
		    COALESCE(p.title, p.name)              AS title,
		    p.url,
		    COALESCE(p.meta_description, '')       AS meta_description,
		    COALESCE(p.nav_label, p.title, p.name) AS nav_label,
		    `+PageImageProjectionSQL+`
		FROM pages p
		JOIN site_areas sa
		  ON sa.id = p.site_area_id
		 AND sa.site_id = p.site_id
		`+PageImageJoinsSQL+`
		WHERE p.site_id = $1
		  AND (lower(sa.name) = lower($2)
		       OR sa.url_prefix = $2
		       OR sa.url_prefix = '/' || $2)
		  AND p.status IN ('active', 'deployed')`+FetchablePageEligibilitySQL+`
		ORDER BY COALESCE(p.nav_order, 100), p.name
		LIMIT $3
	`, siteID, area, limit)
	if err != nil {
		return nil, fmt.Errorf("resolvePagesUnderSection query failed: %w", err)
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var name, title, url, metaDesc, navLabel string
		var img PageImageCols
		if err := rows.Scan(&name, &title, &url, &metaDesc, &navLabel, &img.CardKey, &img.HeroKey, &img.HeroPurpose); err != nil {
			logger.Warn("resolvePagesUnderSection: scan failed", zap.Error(err))
			continue
		}
		items = append(items, map[string]interface{}{
			"name":             name,
			"title":            ListItemTitle(title),
			"url":              url,
			"meta_description": metaDesc,
			"excerpt":          ListItemExcerpt(metaDesc),
			"nav_label":        navLabel,
			"image":            img.WebPath(),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("resolvePagesUnderSection rows iter failed: %w", err)
	}

	logger.Info("queryresolve: resolved pages_under_section",
		zap.String("section", area),
		zap.String("site_id", siteID.String()),
		zap.Int("count", len(items)),
		zap.Int("limit", limit),
	)

	return items, nil
}

// resolveProducts returns active products for the site, optionally filtered
// by category (`query.products:gripper` → category "gripper"). Only
// `status = 'active'` rows are returned — a row can exist (e.g. pending
// verification) without being render-eligible yet.
//
// The `specifications` jsonb column is flattened into the returned map
// alongside the fixed fields, so a component's html_template can reference
// any spec key directly (e.g. {{.stroke}}, {{.gripping_force}}) without a
// nested lookup — the schema's `items` block documents which keys a given
// category's rows are expected to carry, but this resolver doesn't enforce
// that shape; it returns whatever specifications each row has.
//
// Provenance (source_url, verified_date) lives in content_data rather than
// dedicated columns — no migration needed, and it travels with the row
// exactly like every other per-product fact a template might show.
func resolveProducts(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	category string,
	limit int,
	logger *zap.Logger,
) (interface{}, error) {
	const hardCap = PageListingHardCap
	if limit <= 0 {
		limit = 12
	}
	if limit > hardCap {
		limit = hardCap
	}

	query := `
		SELECT
		    id::text, name, COALESCE(sku, ''), COALESCE(subcategory, ''),
		    COALESCE(specifications, '{}'::jsonb),
		    COALESCE(price_display, ''),
		    COALESCE(content_data->>'source_url', ''),
		    COALESCE(content_data->>'verified_date', '')
		FROM products
		WHERE site_id = $1
		  AND status = 'active'
	`
	args := []interface{}{siteID}
	if category != "" {
		query += ` AND category = $2 ORDER BY name LIMIT $3`
		args = append(args, category, limit)
	} else {
		query += ` ORDER BY name LIMIT $2`
		args = append(args, limit)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("resolveProducts query failed: %w", err)
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, name, sku, subcategory, priceDisplay, sourceURL, verifiedDate string
		var specsJSON []byte
		if err := rows.Scan(&id, &name, &sku, &subcategory, &specsJSON, &priceDisplay, &sourceURL, &verifiedDate); err != nil {
			logger.Warn("resolveProducts: scan failed", zap.Error(err))
			continue
		}

		item := map[string]interface{}{
			"id":            id,
			"name":          name,
			"sku":           sku,
			"subcategory":   subcategory,
			"price_display": priceDisplay,
			"source_url":    sourceURL,
			"verified_date": verifiedDate,
		}

		var specs map[string]interface{}
		if len(specsJSON) > 0 {
			if err := json.Unmarshal(specsJSON, &specs); err == nil {
				for k, v := range specs {
					item[k] = v
				}
			}
		}

		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("resolveProducts rows iter failed: %w", err)
	}

	logger.Info("queryresolve: resolved products",
		zap.String("category", category),
		zap.String("site_id", siteID.String()),
		zap.Int("count", len(items)),
		zap.Int("limit", limit),
	)

	return items, nil
}
