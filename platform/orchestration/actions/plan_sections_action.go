// FILE: platform/orchestration/actions/plan_sections_action.go
//
// PlanSectionsAction reads a page's section list, loads each component's
// input_schema (v2 format), resolves data sources, and determines which
// sections can be generated vs which need human input vs which should skip.
//
// This sits between load_page_record and spawn_content_writer in the
// page-build-handler workflow. The content writer only receives sections
// that have all required data available.
//
// Registration:
//   "plan_sections": {
//       Handler:     PlanSectionsAction,
//       Category:    "site",
//       Description: "Resolve section data requirements and triage readiness",
//       IsLocal:     true,
//   },
//
// Workflow config:
//   "plan_sections": {
//       "action": "plan_sections",
//       "config": {
//           "site_id": "site_record.site_id",
//           "sections": "page_record.sections",
//           "page_name": "page_record.name"
//       },
//       "next_step": "check_has_ready_sections",
//       "output_field": "section_plan"
//   }

package actions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/content"
	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/imageryplan"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

var PlanSectionsInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"sections", "section_facts", "section_subjects", "page_name", "pipeline", "work_item_id", "site_type", "page_type"},
	// CheckConfig rather than ConfigKeys: this action reads nothing from
	// params.StepConfig.Config directly — every key reaches it through
	// ExtractActionInputs (:620), which iterates exactly Required ∪ Optional. The
	// spec is therefore already a verified statement of what this action reads,
	// and opting in asserts nothing new about behaviour.
	//
	// The third instance of bugs_open/136's half-landed domain→pipeline rename, and
	// the only one on a path that actually runs: page-build-handler's step carries
	// `domain: "site_record.domain"`, the spec has `pipeline` (which NO live step
	// sets), and the string "domain" does not occur anywhere in this file. Because
	// `domain` is not in the spec, Strategy 0 never resolves that dot-path — the
	// value is fetched by nobody and used by nothing.
	//
	// Opted in HERE specifically because the six agents carrying the other two
	// instances are deliberately quiesced (owner ruling 2026-07-29: the improvement
	// loop is stopped during a heavy development phase), so their warning cannot
	// fire without restarting something that was stopped on purpose. This action
	// ran ~14 times in the preceding 24h, so the detector is exercised by traffic
	// that is already happening.
	CheckConfig: true,
	Defaults:    map[string]interface{}{},
	Deprecated:  map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("plan_sections", PlanSectionsInputSpec)
}

// ============================================================================
// Source resolution
// ============================================================================

// sourceResolver holds cached lookups for a single invocation
// sectionRef identifies one section within a page: its slot name and which
// occurrence of that name it is, counting from zero in page order. A page of
// six illustrated-text-blocks has refs {illustrated-text-block,0} … {…,5}.
//
// Zero value means "no section context" — resolution then behaves exactly as it
// did before per-section binding existed. Callers that genuinely have no
// section (render_site_components passes an empty page name for the same
// reason) get today's page-wide behaviour rather than a wrong guess.
type sectionRef struct {
	Name       string
	Occurrence int
	Known      bool
}

// newSectionRef builds a section identity with the name normalised the way the
// estate's occurrence counter normalises it (lower-cased, trimmed —
// InstanceCounter.NextOccurrence). Both sides of the binding go through here so
// a plan spelling a slot "Article-Body" and a stored row spelling it
// "article-body" are ONE section, which is what the occurrence count already
// assumes. Building the key by hand on one side and through the counter on the
// other is how the two halves of an identity quietly stop matching.
func newSectionRef(name string, occurrence int) sectionRef {
	return sectionRef{
		Name:       strings.ToLower(strings.TrimSpace(name)),
		Occurrence: occurrence,
		Known:      true,
	}
}

type sourceResolver struct {
	siteID        uuid.UUID
	db            *sql.DB
	logger        *zap.Logger
	specs         map[string]map[string]interface{} // aspect → data
	pages         map[string]string                 // page name → url
	assets        map[string]string                 // asset type → url
	siteRow       map[string]string                 // sites-row identity column → value (see ensureSiteRow)
	pageName      string                            // page being planned; scopes per-page asset resolution (hero)
	specsLoaded   bool
	pagesLoaded   bool
	assetsLoaded  bool
	siteRowLoaded bool
	// aliasesUsed records every resolution that came from somewhere other than
	// the path the schema declared: "site_specs.identity.email" →
	// "sites.email". Surfaced in the action result as source_aliases_used so a
	// build record shows which fields changed provenance, rather than that fact
	// living only in a log line.
	aliasesUsed map[string]string
	// sectionAssets binds a section-scope site_plan_imagery row to the ONE
	// section it was planned for, so a page of repeated illustrated blocks can
	// carry a different figure in each. Keyed by the section's identity within
	// the page (slot name + which occurrence of that name), NOT by a position
	// integer: the estate carries two incompatible numbering schemes for the
	// same idea — site_plan_sections.ordering is 0-based and counts site-level
	// slots, page_components.position is 1-based on most pages and neither on
	// 128 of them [MEASURED 2026-08-31] — and a binding computed from one and
	// read through the other is off by a variable amount. The scope_ref ordinal
	// is translated into this key ONCE, in ensureAssets, against the plan that
	// minted it. Both render paths then count occurrences of a slot name in
	// their own order, which is the one thing they agree on.
	//
	// Nil until ensureAssets has run and found at least one section-scope row.
	sectionAssets   map[sectionRef]map[string]string
	planOrder       []string // this page's section list per the current plan
	planOrderLoaded bool
	// liveSectionNames is the section list the CALLING path is actually
	// iterating, in its own order — pages.sections on the build path, the
	// stored page_components slots on the re-render path. The binding is only
	// safe when it agrees with the plan's list (see planSectionOrder), because
	// the ordinal indexes the plan while the occurrence is counted over this.
	// Empty means the caller supplied none, which disables per-section binding.
	liveSectionNames []string
	// storedContent maps slot_name → the page's deployed page_components
	// content_data. It is the carry-forward source for a non-llm field whose
	// declared source resolves nothing (bugs_open/238): a regeneration must not
	// silently lose a resolver-sourced value the live page already carries.
	// Loaded lazily — a plan in which every source resolves never runs the query.
	storedContent       map[string]map[string]interface{}
	storedContentLoaded bool
}

// identityContainerAspects lists the site_specs aspects whose writers group
// fields inside a sub-object, so `<aspect>.<field>` must also be looked for at
// `<aspect>.<container>.<field>`.
//
// `identity` is written by domain-research-classifier, which nests contact
// details: {"contact": {"email":…, "phone":…, "address":…, "location":…}}.
// Component input_schemas ask for the FLAT path (site_specs.identity.email), so
// the two have never agreed. Enumerated rather than a blind deep search: a
// search would make two same-named keys at different depths ambiguous, and the
// site_assets branch below already establishes the enumerated-alias shape.
var identityContainerAspects = map[string][]string{
	"identity": {"contact"},
}

// siteRowIdentityColumns maps a site_specs.identity.<field> leaf onto the sites
// row column holding the SAME fact.
//
// The sites row is the canonical identity store: loadSiteDataFull
// (render_site_components_action.go:337) sources the full-writer render context
// from exactly these columns, and buildRerenderBaseData was changed to prefer
// sites.email over content_data for the same reason (bugs_open/006 §B —
// "making both render paths agree"). plan_sections is the third path and the
// only one that still cannot see them, so an owner-supplied email or phone is
// invisible to every component that declares a site_specs.identity source.
//
// Keys are the spec-side leaf names components actually declare (both spellings
// of address are accepted); values are column names. The set mirrors
// loadSiteDataFull's SELECT, plus contact_address, which holds the fact
// site_specs.identity.address asks for and which no render path reads today.
var siteRowIdentityColumns = map[string]string{
	"email":           "email",
	"phone":           "phone",
	"address":         "contact_address",
	"contact_address": "contact_address",
	"company_name":    "company_name",
	"tagline":         "tagline",
	"logo_text":       "logo_text",
	"logo_url":        "logo_url",
}

// NOTE (signature change): newSourceResolver now takes pageName so site_assets
// resolution can be page-aware (this page's hero rather than a single
// site-wide hero_url). There is one caller (PlanSectionsAction); it passes the
// page_name it already has. An empty pageName degrades safely — the per-page
// hero lookup is skipped and resolution falls back to content_data.
func newSourceResolver(siteID uuid.UUID, db *sql.DB, logger *zap.Logger, pageName string) *sourceResolver {
	return &sourceResolver{
		siteID:   siteID,
		db:       db,
		logger:   logger,
		pageName: pageName,
		specs:    make(map[string]map[string]interface{}),
		pages:    make(map[string]string),
		assets:   make(map[string]string),
	}
}

// withLiveSectionNames records the section list the calling path will iterate,
// which is what per-section imagery binding counts occurrences over. Call it
// before the first resolve; a resolver that is never told stays page-wide.
//
// Deliberately a separate call rather than a constructor argument: the two
// render paths learn their section list at different points (the build path
// after filtering site-level slots out of pages.sections, the re-render path
// after loading the stored rows), and neither has it when the resolver is made.
func (r *sourceResolver) withLiveSectionNames(names []string) *sourceResolver {
	r.liveSectionNames = names
	return r
}

// ensureStoredContent loads this page's deployed content_data rows, once, keyed
// by slot_name (bugs_open/238).
//
// Scoped to `build_status = 'deployed'` deliberately: the carry's contract is
// "what the live page already carries", not "anything ever written". A slot that
// repeats with DIFFERENT content_data is dropped rather than resolved
// arbitrarily — the same conflict rule loadPageSlotComponentIDs applies to slot
// identity, for the same reason: an ambiguous carry source is no carry source.
//
// A query error is not fatal here. Unlike slot identity — where planning against
// a silently-empty map files junk work items — an empty map costs only the carry,
// leaving on_missing to behave exactly as it did before this existed.
func (r *sourceResolver) ensureStoredContent(ctx context.Context) {
	if r.storedContentLoaded {
		return
	}
	r.storedContentLoaded = true
	r.storedContent = make(map[string]map[string]interface{})

	if r.pageName == "" {
		return
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT COALESCE(pc.slot_name, ''), pc.content_data
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1 AND p.name = $2
		  AND pc.build_status = 'deployed'
		  AND pc.content_data IS NOT NULL
	`, r.siteID, r.pageName)
	if err != nil {
		r.logger.Warn("plan_sections: failed to load stored content_data for carry-forward",
			zap.String("page", r.pageName),
			zap.Error(err))
		return
	}
	defer rows.Close()

	seen := make(map[string][]byte)
	conflicted := make(map[string]bool)
	for rows.Next() {
		var slot string
		var dataJSON []byte
		if err := rows.Scan(&slot, &dataJSON); err != nil {
			continue
		}
		if slot == "" {
			continue
		}
		if prior, ok := seen[slot]; ok {
			if !bytes.Equal(prior, dataJSON) {
				conflicted[slot] = true
			}
			continue
		}
		var data map[string]interface{}
		if err := json.Unmarshal(dataJSON, &data); err != nil {
			continue
		}
		seen[slot] = dataJSON
		r.storedContent[slot] = data
	}
	if err := rows.Err(); err != nil {
		r.logger.Warn("plan_sections: stored content_data rows ended early — carry-forward may be partial",
			zap.String("page", r.pageName),
			zap.Error(err))
	}
	for slot := range conflicted {
		delete(r.storedContent, slot)
		r.logger.Warn("plan_sections: slot_name repeats with different content_data — not a carry-forward source",
			zap.String("page", r.pageName),
			zap.String("slot", slot))
	}
}

// storedFieldValue reports the stored value this page already carries for one
// slot/field, and whether it is usable as a carry-forward source.
//
// Emptiness is judged by isEmptyContentValue — the render gate's own predicate —
// because carrying an empty string would carry the defect: `src=""` renders
// identically whether the key is absent or present-and-blank.
func (r *sourceResolver) storedFieldValue(ctx context.Context, slot, field string) (interface{}, bool) {
	r.ensureStoredContent(ctx)
	if len(r.storedContent) == 0 {
		return nil, false
	}

	data, ok := r.storedContent[slot]
	if !ok {
		// Strict fallback, never a rebind: the kebab-normalised form only, for
		// the naming class bugs_open/041 covers.
		data, ok = r.storedContent[NormalizeComponentFunction(slot)]
	}
	if !ok {
		return nil, false
	}

	value, present := data[field]
	if !present || isEmptyContentValue(value) {
		return nil, false
	}
	return value, true
}

// loadSpecs loads all current site_specs for this site (once)
func (r *sourceResolver) ensureSpecs(ctx context.Context) {
	if r.specsLoaded {
		return
	}
	r.specsLoaded = true

	rows, err := r.db.QueryContext(ctx, `
		SELECT aspect, data FROM site_specs
		WHERE site_id = $1 AND is_current = true
	`, r.siteID)
	if err != nil {
		r.logger.Warn("plan_sections: failed to load site_specs", zap.Error(err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var aspect string
		var dataJSON []byte
		if err := rows.Scan(&aspect, &dataJSON); err != nil {
			continue
		}
		var data map[string]interface{}
		if err := json.Unmarshal(dataJSON, &data); err != nil {
			continue
		}
		r.specs[aspect] = data
	}

	r.logger.Info("plan_sections: loaded site_specs",
		zap.Int("aspect_count", len(r.specs)))
}

// ensureSiteRow loads this site's identity columns from the sites row (once).
//
// Deliberately NOT COALESCEd across columns the way loadSiteDataFull is: that
// function needs a non-empty value for a template, so it falls back
// company_name → name → domain. Here an empty value must stay empty, because
// the caller's decision is whether the field resolved AT ALL — substituting the
// domain for a missing company_name would satisfy a `needs_human_review` field
// with a value nobody supplied, which is the failure shape this repo calls a
// defect. Missing stays missing; on_missing then governs.
func (r *sourceResolver) ensureSiteRow(ctx context.Context) {
	if r.siteRowLoaded {
		return
	}
	r.siteRowLoaded = true
	r.siteRow = make(map[string]string)

	var email, phone, address, companyName, tagline, logoText, logoURL string
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(email, ''), COALESCE(phone, ''), COALESCE(contact_address, ''),
		       COALESCE(company_name, ''), COALESCE(tagline, ''),
		       COALESCE(logo_text, ''), COALESCE(logo_url, '')
		FROM sites WHERE id = $1
	`, r.siteID).Scan(&email, &phone, &address, &companyName, &tagline, &logoText, &logoURL)
	if err != nil {
		if err != sql.ErrNoRows {
			r.logger.Warn("plan_sections: failed to load sites identity columns", zap.Error(err))
		}
		return
	}

	for col, val := range map[string]string{
		"email": email, "phone": phone, "contact_address": address,
		"company_name": companyName, "tagline": tagline,
		"logo_text": logoText, "logo_url": logoURL,
	} {
		if val != "" {
			r.siteRow[col] = val
		}
	}
}

// loadPages loads all active pages for URL resolution (once)
func (r *sourceResolver) ensurePages(ctx context.Context) {
	if r.pagesLoaded {
		return
	}
	r.pagesLoaded = true

	rows, err := r.db.QueryContext(ctx, `
		SELECT name, url FROM pages
		WHERE site_id = $1 AND `+datahelpers.PageWantedLivePredicateFor("")+`
	`, r.siteID)
	if err != nil {
		r.logger.Warn("plan_sections: failed to load pages", zap.Error(err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name, url string
		if err := rows.Scan(&name, &url); err != nil {
			continue
		}
		r.pages[name] = url
	}
}

// loadAssets resolves the asset URLs this page's sections may reference, once.
//
// CHANGE: previously this mapped a single site-wide content_data["hero_url"]
// to assets["hero"], so every page shared one hero (and StoreAssetAction
// overwrites hero_url per generation — last-write-wins). It now resolves the
// PAGE'S hero from the current plan's imagery rows joined to the deployed
// asset: site_plan_imagery.key is the asset_key, assets.url is the web path.
// So site_assets.hero on the index page resolves to hero-home.jpg, on
// games-index to hero-games.jpg, etc. The site logo (scope='site',
// kind='logo') resolves the same way. content_data remains a fallback for
// legacy/adopted sites with no plan imagery rows, or assets not yet active.
func (r *sourceResolver) ensureAssets(ctx context.Context) {
	if r.assetsLoaded {
		return
	}
	r.assetsLoaded = true

	// Per-page hero: this page's hero asset from the current plan, joined to
	// the deployed asset row. Skipped when pageName is empty (degrades to the
	// content_data fallback below).
	if r.pageName != "" {
		var assetKey, purpose string
		err := r.db.QueryRowContext(ctx, `
			SELECT a.asset_key, a.purpose
			  FROM site_plan_imagery spi
			  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
			  JOIN assets a ON a.site_id = sp.site_id
			               AND a.asset_key = spi.key
			               AND a.status = 'active'
			 WHERE sp.site_id = $1
			   AND spi.scope = 'page'
			   AND spi.scope_ref = $2
			   AND spi.kind = 'hero'
			 ORDER BY spi.ordering
			 LIMIT 1
		`, r.siteID, r.pageName).Scan(&assetKey, &purpose)
		switch {
		case err == nil && assetKey != "":
			// Resolve to the deployed git path, NOT assets.url (a presigned S3
			// URL that expires and is per-generation).
			r.assets["hero"] = storage.DeployedWebPath(assetKey, purpose)
		case err != nil && err != sql.ErrNoRows:
			r.logger.Warn("plan_sections: per-page hero lookup failed",
				zap.String("page", r.pageName), zap.Error(err))
		}
	}

	// Lane B content hero (Phase I3, D13): a per-article image generated from
	// the article's own content, stored under the literal ContentHeroKey
	// convention with no plan row. The planner's page hero (above) always
	// wins; the site brand hero (below) stays the last resort. This is what
	// makes the article page show the same image family as its listing card.
	if _, ok := r.assets["hero"]; !ok && r.pageName != "" {
		var assetKey, purpose string
		err := r.db.QueryRowContext(ctx, `
			SELECT a.asset_key, a.purpose
			  FROM assets a
			 WHERE a.site_id = $1
			   AND a.asset_key = $2
			   AND a.status = 'active'
			 LIMIT 1
		`, r.siteID, imageryplan.ContentHeroKey(r.pageName)).Scan(&assetKey, &purpose)
		switch {
		case err == nil && assetKey != "":
			r.assets["hero"] = storage.DeployedWebPath(assetKey, purpose)
		case err != nil && err != sql.ErrNoRows:
			r.logger.Warn("plan_sections: content hero lookup failed",
				zap.String("page", r.pageName), zap.Error(err))
		}
	}

	// Site-scope brand hero: fallback when the page has no hero of its own,
	// so image-role-aliased fields still resolve to something brand-consistent
	// rather than nothing. Page-scope (above) always wins.
	if _, ok := r.assets["hero"]; !ok {
		var assetKey, purpose string
		err := r.db.QueryRowContext(ctx, `
			SELECT a.asset_key, a.purpose
			  FROM site_plan_imagery spi
			  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
			  JOIN assets a ON a.site_id = sp.site_id
			               AND a.asset_key = spi.key
			               AND a.status = 'active'
			 WHERE sp.site_id = $1
			   AND spi.scope = 'site'
			   AND spi.kind = 'hero'
			 ORDER BY spi.ordering
			 LIMIT 1
		`, r.siteID).Scan(&assetKey, &purpose)
		switch {
		case err == nil && assetKey != "":
			r.assets["hero"] = storage.DeployedWebPath(assetKey, purpose)
		case err != nil && err != sql.ErrNoRows:
			r.logger.Warn("plan_sections: site-scope hero lookup failed", zap.Error(err))
		}
	}

	// Per-page section imagery: illustrations / icons / infographics requested at
	// section scope for this page (scope_ref = "<page>:<ordinal>"), joined to the
	// deployed asset row. Mapped by KEY (per-key schema paths, e.g. icon sets) and
	// aliased by KIND first-wins (generic paths like site_assets.illustration),
	// mirroring the hero mapping above. Skipped when pageName is empty.
	if r.pageName != "" {
		// scope_ref is selected as well as filtered on: it carries the ordinal
		// that says WHICH section the figure was planned for. Reading it is
		// what lets a page of repeated illustrated sections carry a different
		// figure in each; before this it was matched by prefix and thrown away,
		// and every section on the page resolved the first row of its kind.
		rows, err := r.db.QueryContext(ctx, `
			SELECT spi.kind, spi.scope_ref, a.asset_key, a.purpose
			  FROM site_plan_imagery spi
			  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
			  JOIN assets a ON a.site_id = sp.site_id
			               AND a.asset_key = spi.key
			               AND a.status = 'active'
			 WHERE sp.site_id = $1
			   AND spi.scope = 'section'
			   AND spi.scope_ref LIKE $2 || ':%'
			   AND spi.kind IN ('illustration', 'icon', 'infographic')
			 ORDER BY spi.kind, spi.ordering
		`, r.siteID, r.pageName)
		if err != nil {
			r.logger.Warn("plan_sections: section imagery lookup failed",
				zap.String("page", r.pageName), zap.Error(err))
		} else {
			defer rows.Close()
			type sectionAsset struct{ kind, scopeRef, url string }
			var found []sectionAsset
			for rows.Next() {
				var kind, scopeRef, assetKey, purpose string
				if err := rows.Scan(&kind, &scopeRef, &assetKey, &purpose); err != nil {
					continue
				}
				if assetKey == "" {
					continue
				}
				url := storage.DeployedWebPath(assetKey, purpose)
				// Page-wide map: unchanged. This is what every page carrying a
				// single section figure resolves through today, and a section
				// the ordinal does not name still reaches it here.
				r.assets[assetKey] = url
				if _, exists := r.assets[kind]; !exists {
					r.assets[kind] = url
				}
				found = append(found, sectionAsset{kind: kind, scopeRef: scopeRef, url: url})
			}
			if err := rows.Err(); err != nil {
				r.logger.Warn("plan_sections: section imagery rows error",
					zap.String("page", r.pageName), zap.Error(err))
			}
			if len(found) > 0 {
				// The extra query happens only when this page actually has
				// section-scope figures — a page with none pays nothing, which
				// is every page in the estate bar a handful.
				order := r.planSectionOrder(ctx)
				if len(order) > 0 {
					r.sectionAssets = make(map[sectionRef]map[string]string)
					for _, fa := range found {
						ref, ok := sectionRefForOrdinal(order, fa.scopeRef)
						if !ok {
							// An ordinal that names no section of this page is
							// bugs_open/214's orphan class. Today's behaviour
							// (the page-wide map above) still serves it, so the
							// figure is not lost — but say which one, because
							// "planned, generated, paid for and bound to
							// nothing" is exactly 114's silent shape.
							r.logger.Info("plan_sections: section imagery scope_ref names no section of this page — page-wide fallback only",
								zap.String("page", r.pageName),
								zap.String("scope_ref", fa.scopeRef),
								zap.String("kind", fa.kind))
							continue
						}
						byKind, ok := r.sectionAssets[ref]
						if !ok {
							byKind = make(map[string]string)
							r.sectionAssets[ref] = byKind
						}
						if _, exists := byKind[fa.kind]; !exists {
							byKind[fa.kind] = fa.url
						}
					}
				}
			}
		}
	}

	// Site logo.
	var logoKey, logoPurpose string
	err := r.db.QueryRowContext(ctx, `
		SELECT a.asset_key, a.purpose
		  FROM site_plan_imagery spi
		  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
		  JOIN assets a ON a.site_id = sp.site_id
		               AND a.asset_key = spi.key
		               AND a.status = 'active'
		 WHERE sp.site_id = $1
		   AND spi.scope = 'site'
		   AND spi.kind = 'logo'
		 ORDER BY spi.ordering
		 LIMIT 1
	`, r.siteID).Scan(&logoKey, &logoPurpose)
	switch {
	case err == nil && logoKey != "":
		r.assets["logo"] = storage.DeployedWebPath(logoKey, logoPurpose)
	case err != nil && err != sql.ErrNoRows:
		r.logger.Warn("plan_sections: logo lookup failed", zap.Error(err))
	}

	// Fallback: content_data for anything not resolved above (legacy/adopted
	// sites without plan imagery, or assets not yet active). Gap-fill only —
	// the per-plan values above take precedence.
	var contentDataJSON []byte
	if err := r.db.QueryRowContext(ctx, `
		SELECT content_data FROM sites WHERE id = $1
	`, r.siteID).Scan(&contentDataJSON); err != nil {
		return
	}
	var contentData map[string]interface{}
	if err := json.Unmarshal(contentDataJSON, &contentData); err != nil {
		return
	}
	if _, ok := r.assets["hero"]; !ok {
		if heroURL, ok := contentData["hero_url"].(string); ok && heroURL != "" {
			r.assets["hero"] = heroURL
			// bugs_open/114. Reaching here means every page-scoped route missed and
			// the page will show the site-wide default. That is legitimate for a
			// legacy site with no plan imagery — and it is also what a page whose
			// own generated hero exists looks like, which is the whole of 114's
			// symptom. The two were indistinguishable after the fact: measured
			// 2026-08-15 on mortgagecalculator, six tool pages took this branch
			// while an active, key-matching content_hero_<page> asset sat
			// unreferenced, and by the time anyone looked the orchestration rows
			// were purged, leaving no way to tell which route had been tried.
			//
			// So say which routes were even ELIGIBLE. Three of the four
			// page-scoped routes are gated on pageName, and an empty pageName
			// disables them silently — newSourceResolver's own comment calls that
			// degrading safely, which it is, but safely-and-silently is why this
			// took a purged-history dead end to narrow. A caller that legitimately
			// has no page (render_site_components passes "") is the expected case
			// and says so in the same line.
			r.logger.Info("plan_sections: hero resolved from the site-wide content_data fallback",
				zap.String("page", r.pageName),
				zap.Bool("page_scoped_routes_eligible", r.pageName != ""),
				zap.String("hero_url", heroURL),
				zap.String("bug", "bugs_open/114"))
		}
	}
	if _, ok := r.assets["logo"]; !ok {
		if logoURL, ok := contentData["logo_url"].(string); ok && logoURL != "" {
			r.assets["logo"] = logoURL
		}
	}
}

// planSectionOrder returns this page's section list as the CURRENT plan
// declares it, in ordering order — the same list, including any site-level
// slots, that the scope_ref ordinal was range-checked against when it was
// minted (write_site_plan_imagery_scope.go, `ordinal >= sectionCount`). It is
// the only authority that can turn an ordinal back into a section identity;
// pages.sections is a materialised cache of it and page_components.position is
// a different numbering altogether.
//
// Loaded at most once per resolver, and only when the page has section-scope
// figures to bind. An empty result (no current plan row for this page — an
// adopted or hand-built page) means no per-section binding, and the page-wide
// behaviour that predates this stands unchanged.
func (r *sourceResolver) planSectionOrder(ctx context.Context) []string {
	if r.planOrderLoaded {
		return r.planOrder
	}
	r.planOrderLoaded = true
	if r.pageName == "" {
		return nil
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT sps.component_name
		  FROM site_plan_sections sps
		  JOIN site_plans sp ON sp.id = sps.plan_id AND sp.is_current = true
		 WHERE sp.site_id = $1 AND sps.page_name = $2
		 ORDER BY sps.ordering
	`, r.siteID, r.pageName)
	if err != nil {
		r.logger.Warn("plan_sections: plan section order lookup failed — per-section imagery binding disabled for this page",
			zap.String("page", r.pageName), zap.Error(err))
		return nil
	}
	defer rows.Close()
	// FAIL CLOSED on a bad scan rather than skipping the row (bugs_open/410's
	// shape): a list missing one entry is not a shorter list, it is a list in
	// which every ordinal after the gap names the section BEFORE the one it
	// meant — real figures bound to the wrong sections, on a page that renders
	// and deploys looking correct. There is nothing to salvage from a partial
	// order, so there is no shortfall to measure; the whole binding stands down
	// and the page keeps page-wide resolution.
	var order []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			r.logger.Warn("plan_sections: plan section order scan failed — per-section imagery binding disabled for this page",
				zap.String("page", r.pageName), zap.Error(err))
			return nil
		}
		order = append(order, name)
	}
	if err := rows.Err(); err != nil {
		r.logger.Warn("plan_sections: plan section order rows ended early — no per-section imagery binding for this page",
			zap.String("page", r.pageName), zap.Error(err))
		return nil
	}

	// Locks (RFC_033 / LOCK-008). A human can pin a section onto a live page
	// that no plan tier knows about, and pages.sections carries it — so the
	// live page can hold a section the plan's list does not. This reader must
	// NOT merge those rows into the list: the scope_ref ordinal was minted and
	// range-checked against the PLAN's own list, so merging would move every
	// section after the insertion point and bind figures one section out. But
	// it must not ignore them either, because the consumer paths count
	// occurrences over the LIVE list. So it asks — and where a locked row is
	// not already in the plan, it declines to bind at all and the page keeps
	// the page-wide behaviour that predates this. Degrading is the only safe
	// answer: a shifted binding is a wrong picture on a real page, and it looks
	// exactly like a right one.
	if locked, lerr := datahelpers.LoadLockedPageSlots(ctx, r.db, r.siteID, r.pageName); lerr != nil {
		r.logger.Warn("plan_sections: locked-slot lookup failed — per-section imagery binding disabled for this page",
			zap.String("page", r.pageName), zap.Error(lerr))
		return nil
	} else if _, inserted, _ := datahelpers.MergeLockedPageSlots(order, locked); len(inserted) > 0 {
		r.logger.Info("plan_sections: page carries locked sections the plan does not name — per-section imagery binding disabled, page-wide resolution stands",
			zap.String("page", r.pageName),
			zap.Int("locked_not_in_plan", len(inserted)))
		return nil
	}

	// THE DRIFT GUARD, and it is the general form of the lock case above.
	//
	// The ordinal indexes the PLAN's list; the occurrence is counted over the
	// list the calling path iterates. Those are two orderings maintained
	// independently, and a locked insertion is only one of the ways they come
	// apart — a manual reorder, an earlier section edit, or a re-plan that has
	// not reconciled yet will do it with no lock in sight. Where they disagree,
	// binding does not fail: it silently binds a REAL figure to the WRONG
	// section, which renders, deploys and looks correct.
	//
	// So the two lists are compared, once, and any disagreement stands the whole
	// binding down. The comparison is of the plan's list with its site-level
	// slots removed — the same predicate the build loop filters by, so the two
	// sides are the same kind of list — against the caller's own. Both are
	// normalised the way the occurrence counter normalises, because a spelling
	// difference that the counter would treat as one slot must not read here as
	// a mismatch.
	if !sectionOrderAgrees(order, r.liveSectionNames) {
		r.logger.Info("plan_sections: the plan's section order and the page's live section order disagree — per-section imagery binding disabled, page-wide resolution stands",
			zap.String("page", r.pageName),
			zap.Int("plan_sections", len(order)),
			zap.Int("live_sections", len(r.liveSectionNames)))
		return nil
	}

	r.planOrder = order
	return r.planOrder
}

// sectionOrderAgrees reports whether the plan's section list and the list the
// calling path is iterating describe the same sequence of slots, so that an
// ordinal into the first can be turned into an occurrence counted over the
// second.
//
// Site-level slots (header/footer) are dropped from the PLAN side only: the
// planner may emit them and the build loop filters them out before iterating,
// so they are present in one list and absent from the other by design. Names
// are compared normalised, matching InstanceCounter's own key rule.
//
// An empty live list means the caller never said, which is not agreement.
func sectionOrderAgrees(planOrder, liveNames []string) bool {
	if len(liveNames) == 0 {
		return false
	}
	planned := make([]string, 0, len(planOrder))
	for _, name := range planOrder {
		if isSiteLevelSectionName(name) {
			continue
		}
		planned = append(planned, strings.ToLower(strings.TrimSpace(name)))
	}
	if len(planned) != len(liveNames) {
		return false
	}
	for i, name := range liveNames {
		if planned[i] != strings.ToLower(strings.TrimSpace(name)) {
			return false
		}
	}
	return true
}

// sectionRefForOrdinal translates a scope_ref of the form "<page>:<ordinal>"
// into the identity of the section it names, given that page's plan order.
//
// The ordinal is 0-based and indexes the plan's section list — the same reading
// the mint-side range check uses. Out of range, malformed, or absent means no
// binding: the caller keeps the page-wide behaviour rather than guessing.
func sectionRefForOrdinal(order []string, scopeRef string) (sectionRef, bool) {
	// The ordinal is parsed by the SAME function that range-checks it at write
	// time (write_site_plan_imagery_scope.go). Two hand-written parsers of one
	// field drift, and the drift is silent in both directions.
	ordinal, ok := sectionScopeRefOrdinal(scopeRef)
	if !ok || ordinal >= len(order) {
		return sectionRef{}, false
	}
	// Occurrence comes from the estate's ONE occurrence rule — the same counter
	// that assigns per-instance element-id tokens (InstanceCounter, RFC_032
	// step 3) — walked over the plan's order. A parallel map[string]int would
	// have been four lines and would have disagreed with it on the pages where
	// it matters most: this counter lower-cases and trims, so a plan spelling a
	// slot "Article-Body" and a stored row spelling it "article-body" count as
	// the same slot here and as two different ones under a raw-key map.
	c := NewInstanceCounter()
	var ref sectionRef
	for i := 0; i <= ordinal; i++ {
		occurrence := c.NextOccurrence(order[i])
		if i == ordinal {
			ref = newSectionRef(order[i], occurrence)
		}
	}
	return ref, true
}

// sectionAssetFor returns the figure bound to THIS section for a
// site_assets.<path> lookup, where path is a kind ("illustration") or an asset
// key. Absent binding it returns false and the caller falls through to the
// page-wide map — so a page with one figure, or a section the plan never named,
// resolves exactly as it did before.
func (r *sourceResolver) sectionAssetFor(section sectionRef, path string) (string, bool) {
	if !section.Known || len(r.sectionAssets) == 0 {
		return "", false
	}
	byKind, ok := r.sectionAssets[newSectionRef(section.Name, section.Occurrence)]
	if !ok {
		return "", false
	}
	url, ok := byKind[path]
	return url, ok
}

// sectionDescription returns the purpose/description for a section from the
// site_plan spec. Falls back to page purpose if no section-level description exists.
// Uses already-loaded specs — no extra DB query.
func (r *sourceResolver) sectionDescription(pageName, sectionType string) string {
	plan, ok := r.specs["site_plan"]
	if !ok {
		return ""
	}

	pages, ok := plan["pages"].([]interface{})
	if !ok {
		return ""
	}

	for _, pageRaw := range pages {
		page, ok := pageRaw.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := page["name"].(string)
		if name != pageName {
			continue
		}

		// Check section_descriptions map (if planner provides it)
		if descs, ok := page["section_descriptions"].(map[string]interface{}); ok {
			if desc, ok := descs[sectionType].(string); ok && desc != "" {
				return desc
			}
		}

		// Check section_types array for objects with description
		if sectionTypes, ok := page["section_types"].([]interface{}); ok {
			for _, stRaw := range sectionTypes {
				if st, ok := stRaw.(map[string]interface{}); ok {
					if stName, _ := st["name"].(string); stName == sectionType {
						if desc, _ := st["description"].(string); desc != "" {
							return desc
						}
					}
				}
			}
		}

		// Fall back to page purpose
		if purpose, ok := page["purpose"].(string); ok && purpose != "" {
			return fmt.Sprintf("Section '%s' on page '%s' (purpose: %s)", sectionType, pageName, purpose)
		}
	}

	return ""
}

// resolve checks if a data source has a value available
// Returns: value (if found), found (bool)
func (r *sourceResolver) resolve(ctx context.Context, source string, section sectionRef) (interface{}, bool) {
	if source == "" || source == "llm" || source == "renderer" || source == "static" {
		// These sources don't need resolution — they're generated at render time
		return nil, true
	}

	parts := strings.SplitN(source, ".", 2)
	if len(parts) < 2 {
		return nil, false
	}

	prefix := parts[0]
	path := parts[1]

	switch prefix {
	case "renderer", "static":
		// These are injected at render time — always considered available
		return nil, true

	case "site_specs":
		r.ensureSpecs(ctx)
		if val, ok := r.resolveSpecPath(path); ok {
			return val, true
		}
		// The literal spec path missed. Try the two aliases that name the SAME
		// fact, in order of authority. Exact literal paths always win above, so
		// no path that resolves today changes its value — this branch only adds
		// resolution where the aspect held nothing.
		return r.resolveSpecAlias(ctx, path)

	case "site_assets":
		r.ensureAssets(ctx)
		// This section's OWN figure first: a page of repeated illustrated
		// blocks declares the same source in every one of them, so the
		// page-wide map below can only ever answer them all the same way.
		if url, ok := r.sectionAssetFor(section, path); ok {
			return url, true
		}
		if url, ok := r.assets[path]; ok {
			return url, true
		}
		// Literal key missed — try the image-role alias. Preset/imported
		// components name their image fields freely (site_assets.background,
		// site_assets.product_screenshot, ...) but the pipeline generates
		// per-page heroes; without the alias those fields resolve to nothing
		// and templates render src="". Exact keys above always win, so a
		// future dedicated asset under the literal key takes precedence.
		if role, ok := imageryplan.ImageRoleForPath(path); ok {
			if url, ok := r.assets[role]; ok {
				r.logger.Info("plan_sections: site_assets path resolved via image-role alias",
					zap.String("path", path),
					zap.String("role", role),
					zap.String("page", r.pageName))
				return url, true
			}
		}
		return nil, false

	case "pages":
		r.ensurePages(ctx)
		if url, ok := r.pages[path]; ok {
			return url, true
		}
		// No such page — do NOT fabricate a URL. Returning (nil, false) lets
		// the field's on_missing govern (skip_field drops the field; gated
		// templates then render no button). Fabricating "/<path>.html" here
		// was the phantom-link generator (/contact.html, /services.html on
		// every hero/CTA site-wide).
		r.logger.Info("plan_sections: pages source not found; deferring to on_missing",
			zap.String("page_ref", path),
			zap.String("site_id", r.siteID.String()))
		return nil, false

	case "config":
		r.ensureSpecs(ctx)
		return r.resolveConfigPath(path)

	case "query":
		// Query sources are resolved by the field-loop in planSection via
		// the queryresolve package, BEFORE this method is called. This case
		// is defensive — if a future caller invokes resolveSource directly
		// on a query.* source (instead of going through the field loop),
		// returning (nil, true) keeps the system stable rather than treating
		// it as an unknown source. Real callers should not hit this branch.
		return nil, true

	default:
		return nil, false
	}
}

// resolveSpecPath navigates site_specs: "identity.team" → specs["identity"]["team"]
func (r *sourceResolver) resolveSpecPath(path string) (interface{}, bool) {
	parts := strings.SplitN(path, ".", 2)
	aspect := parts[0]

	specData, ok := r.specs[aspect]
	if !ok {
		return nil, false
	}

	if len(parts) == 1 {
		return specData, true
	}

	// Navigate deeper: "identity.team" → specs["identity"]["team"]
	return navigateMap(specData, parts[1])
}

// resolveSpecAlias is the fallback chain for a site_specs path that missed
// literally. It resolves the same FACT from another store, never a different
// fact, and is consulted only after the literal path has been tried.
//
// Two steps, in order of authority:
//
//  1. the writer's own nested shape — `identity.email` → `identity.contact.email`.
//     Closes the contract mismatch in bugs_open/072: every component schema asks
//     flat, the classifier only ever writes nested, so a site whose identity was
//     never hand-patched cannot resolve a contact field at all.
//  2. the sites row's identity columns — `identity.email` → `sites.email`. This
//     is the canonical store (loadSiteDataFull reads it; buildRerenderBaseData
//     was changed to prefer it, bugs_open/006 §B) and plan_sections was the one
//     path that could not see it.
//
// Only two levels deep by design: an alias applies to `<aspect>.<leaf>`, so a
// deeper path like `identity.team.members` is left alone. Firing is logged with
// the path that won, because a silently-aliased resolve is indistinguishable
// from a literal one in a build record otherwise.
func (r *sourceResolver) resolveSpecAlias(ctx context.Context, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	if len(parts) != 2 {
		return nil, false
	}
	aspect, leaf := parts[0], parts[1]

	// 1. The writer's nested container.
	//
	// NOT redundant with the sites-row step below, though it looks it. The action
	// designed to copy this nested shape into the sites columns —
	// sync_site_identity_action, which reads exactly identity.contact.email/phone
	// — is registered in Go and wired into **zero** live agents (measured
	// 2026-07-31: 0 matching agent_definitions rows, against 9 for plan_sections
	// as a positive control). Its own header says it "should be added as a step in
	// the build flow"; it never was. So nothing else in the pipeline reads the
	// classifier's nested shape, and a new site — which is written nested-only —
	// would resolve nothing without this step. Raised as scope by the council
	// gate's edit-quality seat (corr dd03a73b) and kept on that evidence.
	for _, container := range identityContainerAspects[aspect] {
		if val, ok := r.resolveSpecPath(aspect + "." + container + "." + leaf); ok {
			r.noteAlias("site_specs."+path, "site_specs."+aspect+"."+container+"."+leaf)
			r.logger.Info("plan_sections: site_specs path resolved via the writer's nested shape",
				zap.String("requested", "site_specs."+path),
				zap.String("resolved_from", "site_specs."+aspect+"."+container+"."+leaf),
				zap.String("page", r.pageName))
			return val, true
		}
	}

	// 2. The canonical sites row.
	if aspect != "identity" {
		return nil, false
	}
	col, mapped := siteRowIdentityColumns[leaf]
	if !mapped {
		return nil, false
	}
	r.ensureSiteRow(ctx)
	if val, ok := r.siteRow[col]; ok {
		r.noteAlias("site_specs."+path, "sites."+col)
		r.logger.Info("plan_sections: site_specs path resolved from the canonical sites row",
			zap.String("requested", "site_specs."+path),
			zap.String("resolved_from", "sites."+col),
			zap.String("page", r.pageName))
		return val, true
	}
	return nil, false
}

// noteAlias records a resolution that did not come from the declared path, so
// the build result carries it (source_aliases_used) and not only the log.
func (r *sourceResolver) noteAlias(requested, resolvedFrom string) {
	if r.aliasesUsed == nil {
		r.aliasesUsed = make(map[string]string)
	}
	r.aliasesUsed[requested] = resolvedFrom
}

// resolveConfigPath navigates site content_data config
func (r *sourceResolver) resolveConfigPath(path string) (interface{}, bool) {
	// Config values live in site_specs under various aspects
	// Search across relevant aspects
	for _, aspect := range []string{"site_config", "identity", "design_intent"} {
		if specData, ok := r.specs[aspect]; ok {
			if val, found := navigateMap(specData, path); found {
				return val, true
			}
		}
	}
	return nil, false
}

func navigateMap(data map[string]interface{}, dotPath string) (interface{}, bool) {
	parts := strings.Split(dotPath, ".")
	var current interface{} = data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, exists := v[part]
			if !exists {
				return nil, false
			}
			current = val
		default:
			return nil, false
		}
	}

	// Check if the value is actually populated (not empty string, not empty array)
	switch v := current.(type) {
	case string:
		if v == "" {
			return nil, false
		}
	case []interface{}:
		if len(v) == 0 {
			return nil, false
		}
	case nil:
		return nil, false
	}

	return current, true
}

// ============================================================================
// Section planning result
// ============================================================================

// llmFieldSpec carries per-field metadata for the Step 3 targeted-prompt
// path on page-content-writer. Each entry corresponds to one field whose
// `source` is "llm" in the component's input_schema. The page-content-writer
// prompt template iterates this list instead of dumping the full schema —
// the LLM is asked for exactly the fields it should write, with their
// types and intent, and never given the opportunity to fabricate
// query-resolved data (items, urls, page lists) that the system handles
// elsewhere.
type llmFieldSpec struct {
	Name        string      `json:"name"`
	Type        string      `json:"type,omitempty"` // text | url | image | rich_text | …
	Required    bool        `json:"required,omitempty"`
	Description string      `json:"description,omitempty"` // sourced from input_schema field's `llm_guidance` key
	OnMissing   string      `json:"on_missing,omitempty"`  // skip_field | use_fallback | error
	Fallback    interface{} `json:"fallback,omitempty"`    // value used when on_missing=use_fallback
	// ItemFields lists the field names each element of an array-typed field
	// must contain, from the schema field's `items` (or `item_schema`) map.
	// Empty for non-array fields. Surfaced to the LLM (via the prompt) and to
	// the render-time reconciler so the model emits the exact keys the
	// component template reads, instead of guessing item field names (e.g.
	// title/body) that render empty against a template reading name/description.
	ItemFields []string `json:"item_fields,omitempty"`
	// ValueShape and ItemNotes carry the NESTED element shape that ItemFields
	// cannot express (bugs_open/437). A name list flattens
	// `steps[].branches: array of {body,label}` to the bare name `branches`, and
	// the prompt's exemplar then rendered it `"branches": "..."` — instructing a
	// string, which the writer duly produced and the render gate duly refused,
	// 119 times across six sites in a fortnight. ValueShape is the whole field
	// value as a JSON skeleton; ItemNotes states each structured property's shape
	// and its own schema description. Both are produced by
	// datahelpers.StructuredItemShape, in the same package as the gate that
	// judges the result, and both are EMPTY unless an element property is itself
	// a collection — which is what keeps every other component's prompt
	// byte-identical, and what makes the Go and template halves of the fix safe
	// to deploy in either order.
	ValueShape string   `json:"value_shape,omitempty"`
	ItemNotes  []string `json:"item_notes,omitempty"`
}

type sectionPlanItem struct {
	Name         string                 `json:"name"`
	ComponentID  string                 `json:"component_id"`
	Function     string                 `json:"function"`
	Status       string                 `json:"status"` // "ready", "deferred", "skipped"
	ResolvedData map[string]interface{} `json:"resolved_data,omitempty"`
	LLMFields    []string               `json:"llm_fields,omitempty"`
	// LLMFieldSpecs is the richer counterpart to LLMFields: each spec carries
	// the field's name plus the metadata the targeted-prompt template needs
	// (type, required flag, description, on_missing handling, fallback value).
	// LLMFields stays as a fast lookup of "which fields are LLM-written";
	// LLMFieldSpecs is what page-content-writer's prompt iterates.
	LLMFieldSpecs []llmFieldSpec `json:"llm_field_specs,omitempty"`
	Missing       []missingField `json:"missing,omitempty"`
	Reason        string         `json:"reason,omitempty"`
	// Component carries the full per-section component data as returned by
	// the shared loadSectionComponents helper. Populated when a component
	// was found (Paths 1 and 2). Nil for paths where no component was
	// resolved (Path 3: not_found / selector_unavailable). Downstream
	// consumers — page-content-writer in Step 3 — read input_schema,
	// html_template, render_mode, description, category, content_brief
	// etc. from here instead of re-loading via load_page_section_components.
	Component map[string]interface{} `json:"component,omitempty"`
	// Plan-time fact scoping (bugs_open/151 candidate 1). FactsScoped is
	// true when the plan row carried a non-NULL assigned_fact_ids — the
	// writer prompt then uses AssignedWriterBlock (composed from ONLY the
	// assigned facts, values current at compose time) instead of the
	// whole-site writer_block; an empty AssignedWriterBlock with
	// FactsScoped=true means the section deliberately states no verified
	// facts. All three absent = unscoped = pre-existing behaviour.
	FactsScoped         bool     `json:"facts_scoped,omitempty"`
	AssignedFactIDs     []string `json:"assigned_fact_ids,omitempty"`
	AssignedWriterBlock string   `json:"assigned_writer_block,omitempty"`
	// Subject is the planner's one-line statement of what THIS section
	// specifically covers (per-section subjects build, 2026-08-26; RFC_016
	// §5.1's next structured field). Empty = unassigned = pre-existing
	// behaviour, where every same-named slot shares the page-level brief.
	// Rides to the writer as current_section.subject; the v5 prompt renders
	// it only when non-empty.
	Subject string `json:"subject,omitempty"`
	// CarriedFields names the non-llm fields whose declared source resolved
	// nothing and which were satisfied from this page's own deployed
	// content_data instead (bugs_open/238). StructuralMisses names the required
	// non-llm fields that resolved NOWHERE — neither source nor stored row —
	// and were then omitted under on_missing=skip_field.
	//
	// Both keys are additive and omitempty, so every existing consumer of a
	// section plan entry (the writer prompt, persistSectionSkips, the rerender
	// resolver) is unaffected: absence renders falsy under missingkey=zero and
	// unmarshals to nil.
	CarriedFields    []string       `json:"carried_fields,omitempty"`
	StructuralMisses []missingField `json:"structural_misses,omitempty"`
}

type missingField struct {
	Field     string                 `json:"field"`
	Source    string                 `json:"source"`
	OnMissing string                 `json:"on_missing"`
	Reason    string                 `json:"reason"`
	Type      string                 `json:"type,omitempty"`      // from input_schema field type
	Items     map[string]interface{} `json:"items,omitempty"`     // from input_schema items (array element schema)
	MinItems  int                    `json:"min_items,omitempty"` // from input_schema min_items
}

// ============================================================================
// Main action
// ============================================================================

func PlanSectionsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "plan_sections"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		PlanSectionsInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	pageName := inputs.Get("page_name")
	workItemID := inputs.Get("work_item_id")

	// site_type and page_type for the component selector fallback path.
	// If not provided (existing workflows), the selector still works —
	// it just scores without site/page type relevance bonuses.
	siteType := inputs.Get("site_type")
	pageType := inputs.Get("page_type")
	if pageType == "" {
		pageType = pageName // fall back to page name as page type
	}

	// Parse sections list. sectionFacts is the OPTIONAL plan-time fact
	// assignment (bugs_open/151 candidate 1), an array aligned by index with
	// the sections input — supplied only when the step config wires it (the
	// feature is config-opt-in) and only ever emitted by
	// load_page_sections_from_spec's authoritative tier. Facts are consumed
	// in the SAME walk that accepts a name, so a skipped entry can never
	// shift an assignment onto the wrong section. nil = unscoped.
	sectionsRaw := inputs.GetRaw("sections")
	factsRaw, _ := inputs.GetRaw("section_facts").([]interface{})
	// section_subjects mirrors section_facts: aligned by index, wired by step
	// config, emitted only by load_page_sections_from_spec's authoritative
	// tier. "" = unassigned (pre-existing behaviour).
	subjectsRaw, _ := inputs.GetRaw("section_subjects").([]interface{})
	subjectAt := func(i int) string {
		if i >= len(subjectsRaw) {
			return ""
		}
		s, _ := subjectsRaw[i].(string)
		return strings.TrimSpace(s)
	}
	factsAt := func(i int) []string {
		if i >= len(factsRaw) {
			return nil
		}
		entry, ok := factsRaw[i].([]interface{})
		if !ok {
			return nil // null / wrong shape -> unscoped
		}
		ids := make([]string, 0, len(entry))
		for _, e := range entry {
			if s, ok := e.(string); ok && s != "" {
				ids = append(ids, s)
			}
		}
		return ids
	}
	var sectionNames []string
	var sectionFacts [][]string
	var sectionSubjects []string

	switch v := sectionsRaw.(type) {
	case []interface{}:
		for i, s := range v {
			if name, ok := s.(string); ok {
				sectionNames = append(sectionNames, name)
				sectionFacts = append(sectionFacts, factsAt(i))
				sectionSubjects = append(sectionSubjects, subjectAt(i))
			}
		}
	case []string:
		sectionNames = v
		sectionFacts = make([][]string, len(v))
		sectionSubjects = make([]string, len(v))
		for i := range v {
			sectionFacts[i] = factsAt(i)
			sectionSubjects[i] = subjectAt(i)
		}
	case string:
		// Try JSON parse
		if err := json.Unmarshal([]byte(v), &sectionNames); err != nil {
			return nil, fmt.Errorf("failed to parse sections: %w", err)
		}
		sectionFacts = make([][]string, len(sectionNames))
		sectionSubjects = make([]string, len(sectionNames))
		for i := range sectionNames {
			sectionFacts[i] = factsAt(i)
			sectionSubjects[i] = subjectAt(i)
		}
	}

	// ── Filter out site-level component names ────────────────────────
	// The planner or adoption flow may include header/footer names in the
	// page sections list. These are site-level components handled by
	// InjectHeader/InjectFooter — if we process them here they end up as
	// page_components rows, causing duplicate headers/footers on assembly.
	// Names, facts and subjects are filtered as TRIPLES so alignment survives.
	{
		keptNames := make([]string, 0, len(sectionNames))
		keptFacts := make([][]string, 0, len(sectionFacts))
		keptSubjects := make([]string, 0, len(sectionSubjects))
		for i, s := range sectionNames {
			if isSiteLevelSectionName(s) {
				logger.Info("plan_sections: filtered site-level section",
					zap.String("section", s))
				continue
			}
			keptNames = append(keptNames, s)
			keptFacts = append(keptFacts, sectionFacts[i])
			keptSubjects = append(keptSubjects, sectionSubjects[i])
		}
		sectionNames = keptNames
		sectionFacts = keptFacts
		sectionSubjects = keptSubjects
	}

	if len(sectionNames) == 0 {
		return map[string]interface{}{
			"sections_ready":    []interface{}{},
			"sections_deferred": []interface{}{},
			"sections_skipped":  []interface{}{},
			"ready_count":       0,
			"reason":            "no sections to plan",
		}, nil
	}

	// ── bugs_open/443: repeated component type, no subject to tell the
	// instances apart ────────────────────────────────────────────────────
	// Observe-only, build-side sibling of the planner-side
	// SUBJECT_MISSING_ON_REPEATED_COMPONENT (write_site_plan, rule 17):
	// distinct code because the remedy differs — the planner's says
	// "replan"; this one says the page's SERVING TIER carries no subjects,
	// so N same-typed slots are about to receive one identical brief and
	// the near-duplicate output is the predicted result, not a bad roll.
	// Deliberately NOT gated on any-subject-present: the planner's gate is
	// retro-spam protection for pre-rule-17 plans, and the pages this
	// exists to surface (fallback-tier, no reachable subject store) can
	// never carry one. Bounded: [MEASURED 2026-09-02] 25 repeat-layout
	// pages fleet-wide. Repeats are counted page-wide, not adjacently —
	// non-adjacent repeats duplicate too (443's our-position-on-ai case).
	for _, gap := range repeatedComponentSubjectGaps(sectionNames, sectionSubjects) {
		LogActionEntryInheritingProvenance(ctx, params, agenterrors.Entry{
			SiteID: siteIDStr,
			Action: "plan_sections",
			ErrorMessage: fmt.Sprintf("page %q builds %d %q sections and %d of them carry no subject — identical briefs, near-duplicate output predicted (bugs_open/443)",
				pageName, gap.Repeats, gap.Component, gap.WithoutSubject),
			ErrorCode: "REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT",
			Severity:  "warning",
			Context: map[string]interface{}{
				"page_name":       pageName,
				"component":       gap.Component,
				"repeats":         gap.Repeats,
				"without_subject": gap.WithoutSubject,
				"remedy":          "give the repeated slots distinct subjects at the page's serving tier: site_plan_sections.subject for planned sites (rule 17, seed 640), pages.section_subjects for plan-less pages (bugs_open/443)",
			},
		}, logger)
	}

	// Load component schemas for these sections
	components := loadComponentSchemas(ctx, params.DB, sectionNames, logger)

	// bugs_open/204: on a decomposed site, pages.sections holds POSITIONAL
	// slot names that are no component's name/function, so the map above can
	// never resolve them — but the page's own page_components rows already
	// know exactly which component each slot is. Load that identity map and
	// the schemas for the ids it names; the section loop below tries this
	// FIRST (Path 0), mirroring the re-render path's 182 fix so the two call
	// sites of this judgement cannot drift apart again.
	slotIDs := map[string]string{}
	if pageName != "" {
		var slotErr error
		slotIDs, slotErr = loadPageSlotComponentIDs(ctx, params.DB, siteID, pageName, logger)
		if slotErr != nil {
			return nil, fmt.Errorf("plan_sections: %w", slotErr)
		}
	}
	var byID map[string]componentInfo
	var byIDDrops map[string]string
	if len(slotIDs) > 0 {
		seenIDs := make(map[string]bool, len(slotIDs))
		componentIDs := make([]string, 0, len(slotIDs))
		for _, id := range slotIDs {
			if !seenIDs[id] {
				seenIDs[id] = true
				componentIDs = append(componentIDs, id)
			}
		}
		byID, byIDDrops = loadComponentSchemasByID(ctx, params.DB, componentIDs, logger)
	}

	// Create resolver
	// The list this loop will iterate is what per-section imagery binding counts
	// occurrences over, and it is only safe to bind when it agrees with the
	// plan's own order — sectionNames is post-filter here, which is the form
	// sectionOrderAgrees compares against.
	resolver := newSourceResolver(siteID, params.DB, logger, pageName).
		withLiveSectionNames(sectionNames)

	// Pre-load specs so we can extract design_direction for needs_new_component items.
	// ensureSpecs is idempotent — later calls in planSection() won't re-query.
	resolver.ensureSpecs(ctx)

	// Extract design_direction from design_intent spec (if present).
	// Passed to needs_new_component items so the component-creator knows the visual style.
	designDirection := ""
	if di, ok := resolver.specs["design_intent"]; ok {
		if sd, ok := di["style_direction"].(string); ok && sd != "" {
			designDirection = sd
		}
	}

	// ── Load open data requests for reconciliation after planning ────
	// Used after the planning loop to:
	//   1. Close stale requests for sections that are now ready (component created, data arrived)
	//   2. Skip creating duplicate requests for sections that are still deferred
	openDataRequests := loadOpenSectionDataRequests(ctx, params.DB, siteID, pageName, logger)

	// Plan each section
	var ready []sectionPlanItem
	var deferred []sectionPlanItem
	var skipped []sectionPlanItem

	// Plan-time fact scoping (bugs_open/151 candidate 1): the evidence base
	// is already in the resolver's spec cache (ensureSpecs above), so scoping
	// costs no query. scopeItem is a no-op for an unscoped section (nil
	// facts), which is every section until a plan written WITH assignments
	// reaches this action through the config-wired section_facts input.
	evidenceBase := resolver.specs["evidence_base"]
	scopeItem := func(item *sectionPlanItem, facts []string, subject string) {
		item.Subject = subject // "" = unassigned; omitempty keeps the item byte-identical
		if facts == nil {
			return
		}
		block := composeScopedWriterBlock(evidenceBase, facts, logger, item.Name)
		if len(facts) > 0 && block == "" {
			// A non-empty assignment that composes to NOTHING (every ID
			// unknown, or the evidence base has gone) is a composition
			// anomaly, NOT a deliberately factless section — those carry [].
			// Marking it scoped would render the "state no facts here"
			// branch, making a broken assignment indistinguishable from a
			// deliberate one (council 902a8563, bug_historian's missing
			// case). Degrade to UNSCOPED — the writer falls back to today's
			// site-wide block — and record durably, not just in pod output.
			logger.Warn("plan_sections: fact assignment composed an empty writer block — degrading section to unscoped",
				zap.String("section", item.Name),
				zap.Strings("assigned_fact_ids", facts))
			// The running step IS the right provenance here — plan_sections
			// records its own degradation — so inheritance is declared rather
			// than left to a silent merge.
			LogActionEntryInheritingProvenance(ctx, params, agenterrors.Entry{
				SiteID: siteID.String(),
				Action: "plan_sections",
				ErrorMessage: fmt.Sprintf("fact assignment for section %q composed an empty writer block (%d assigned IDs, none resolvable)",
					item.Name, len(facts)),
				ErrorCode: "FACT_SCOPING_EMPTY_COMPOSITION",
				Severity:  "warning",
				Context: map[string]interface{}{
					"section":           item.Name,
					"page_name":         pageName,
					"assigned_fact_ids": facts,
					"remedy":            "the plan's assigned_fact_ids match no current evidence_base fact with a writer_line; replan the site or correct the register — the section built with the site-wide block meanwhile",
				},
			}, logger)
			return
		}
		item.FactsScoped = true
		item.AssignedFactIDs = facts
		item.AssignedWriterBlock = block
	}

	// Build selector context for the fallback path.
	// This is only used when a section name doesn't match a component function directly.
	selCtx := SelectorContext{
		SiteType: siteType,
		PageType: pageType,
		PageName: pageName,
	}

	// Which occurrence of its own slot name each section is, counted in page
	// order — the section's identity for per-section imagery binding (see
	// sectionRef). It uses the estate's ONE occurrence rule rather than a local
	// tally: InstanceCounter is what assigns per-instance element-id tokens over
	// this same list, and two counters over one iteration are two rules that
	// agree until a spelling or a filter makes them disagree.
	sectionOccurrences := NewInstanceCounter()
	for sectionIdx, sectionName := range sectionNames {
		thisSection := newSectionRef(sectionName, sectionOccurrences.NextOccurrence(sectionName))
		// Path 0: stored-identity lookup (bugs_open/204 — the build-path half
		// of bugs_open/182). The page's own page_components row names exactly
		// which component this slot is; that identity does not depend on slot
		// naming at all, so it is tried first. Semantics mirror the re-render
		// path: id wins over a disagreeing name resolution (observe-only log),
		// an id whose template failed the guard defers LOUDLY rather than
		// falling back to a coincidentally name-matched component (the same
		// silent substitution one level down), and an id that resolves to no
		// active row falls through to the name/selector paths.
		if cid := slotIDs[sectionName]; cid != "" {
			if ci, ok := byID[cid]; ok {
				if nameComp, nameOK := components[sectionName]; nameOK && nameComp.ID != ci.ID {
					logger.Info("plan_sections: component_id and section name resolve to different components (observe-only, id wins)",
						zap.String("section", sectionName),
						zap.String("id_resolved_component", ci.ID),
						zap.String("id_resolved_name", ci.Name),
						zap.String("name_resolved_component", nameComp.ID),
						zap.String("name_resolved_name", nameComp.Name))
				}
				item := planSection(ctx, sectionName, thisSection, ci, resolver, logger)
				scopeItem(&item, sectionFacts[sectionIdx], sectionSubjects[sectionIdx])
				switch item.Status {
				case "ready":
					ready = append(ready, item)
				case "deferred":
					deferred = append(deferred, item)
				case "skipped":
					skipped = append(skipped, item)
				}
				continue
			}
			if _, dropped := byIDDrops[cid]; dropped {
				// The component EXISTS — filing needs_new_component would ask
				// the fleet to build what is already there (the 204 canary
				// filed two junk items per section this way). Defer just this
				// section with a reason a human can act on; the deferred-items
				// writer surfaces it as needs_section_data with this reason as
				// its summary. Unlike the re-render path this is not fatal to
				// the run: planning writes nothing over good HTML, and one
				// broken pinned template must not block the page's other
				// sections.
				deferred = append(deferred, sectionPlanItem{
					Name:        sectionName,
					ComponentID: cid,
					Status:      "deferred",
					Reason:      fmt.Sprintf("stored component %s for slot %q has an invalid/truncated template — repair that component; do not create a new one", cid, sectionName),
					Missing: []missingField{{
						Field:  "html_template",
						Source: "content_components",
						Reason: fmt.Sprintf("stored component %s for slot %q failed the template guard — repair that component; do not create a new one", cid, sectionName),
					}},
				})
				continue
			}
			// id resolved to no active row (retired component, or the row is
			// gone) — the name/selector paths below still get their chance,
			// matching loadContentComponentsByID's stated contract.
		}

		// Path 1: Direct function/name lookup (existing behaviour).
		// All current sites hit this path — their planners output function names.
		comp, ok := components[sectionName]
		if ok {
			item := planSection(ctx, sectionName, thisSection, comp, resolver, logger)
			scopeItem(&item, sectionFacts[sectionIdx], sectionSubjects[sectionIdx])

			switch item.Status {
			case "ready":
				ready = append(ready, item)
			case "deferred":
				deferred = append(deferred, item)
			case "skipped":
				skipped = append(skipped, item)
			}
			continue
		}

		// Path 2: Section type selector.
		// The planner output a section_type (e.g. "provocation-card") rather than
		// a specific function name. The selector queries content_components by
		// section_type, scores candidates, and returns the best match.
		resolved, resolution := resolveSectionComponent(ctx, params.DB, sectionName, selCtx, logger)
		if resolved != nil {
			// Selector found a matching component. Its function flows through the
			// rest of the pipeline exactly as if the planner had specified it directly.
			item := planSection(ctx, resolved.Function, thisSection, *resolved, resolver, logger)
			// Preserve the original section_type name — downstream logging and
			// the content writer use item.Name as the section identifier.
			item.Name = sectionName
			scopeItem(&item, sectionFacts[sectionIdx], sectionSubjects[sectionIdx])
			switch item.Status {
			case "ready":
				ready = append(ready, item)
			case "deferred":
				deferred = append(deferred, item)
			case "skipped":
				skipped = append(skipped, item)
			}
			continue
		}

		// Path 3: No component found anywhere.
		if resolution == "not_found" {
			logger.Info("plan_sections: no component for section_type, creating work item",
				zap.String("section_type", sectionName),
				zap.String("page", pageName))

			// Try to get a meaningful description from the site_plan spec
			// (resolver already has specs loaded — no extra DB query needed)
			description := resolver.sectionDescription(pageName, sectionName)
			if description == "" {
				description = fmt.Sprintf("Component for section type %q on page %q (%s site)", sectionName, pageName, siteType)
			}

			err := CreateNeedsNewComponentItem(
				ctx, params.DB, siteIDStr,
				sectionName, pageName, description,
				designDirection, // extracted from resolver.specs before the loop
				siteType, logger,
			)
			if err != nil {
				logger.Warn("plan_sections: failed to create needs_new_component work item",
					zap.String("section_type", sectionName),
					zap.Error(err))
			}

			deferred = append(deferred, sectionPlanItem{
				Name:   sectionName,
				Status: "deferred",
				Reason: fmt.Sprintf("no component for section_type %q — needs_new_component work item created", sectionName),
			})
		} else {
			// Selector error or unavailable — fall through to content writer (backward compat).
			// This keeps the same behaviour as before for edge cases where the DB query fails.
			logger.Warn("plan_sections: selector unavailable, passing section to content writer as-is",
				zap.String("section", sectionName),
				zap.String("resolution", resolution))
			item := sectionPlanItem{
				Name:   sectionName,
				Status: "ready",
				Reason: "selector unavailable — passing to content writer as-is",
			}
			scopeItem(&item, sectionFacts[sectionIdx], sectionSubjects[sectionIdx])
			ready = append(ready, item)
		}
	}

	// ── Reconcile open data requests with planning results ───────────
	// Close stale requests for sections that are now ready (component
	// created since the request was filed, data now available, etc.)
	if params.DB != nil && len(openDataRequests) > 0 {
		for _, section := range ready {
			if _, wasOpen := openDataRequests[section.Name]; wasOpen {
				closeResolvedDataRequest(ctx, params.DB, siteID, pageName, section.Name, logger)
			}
		}
	}

	// Create work items for deferred sections (skips those that already have open requests)
	if params.DB != nil && len(deferred) > 0 {
		// Filter out sections that already have open data requests — no duplicate items
		var newDeferred []sectionPlanItem
		for _, section := range deferred {
			if _, alreadyOpen := openDataRequests[section.Name]; !alreadyOpen {
				newDeferred = append(newDeferred, section)
			}
		}
		if len(newDeferred) > 0 {
			createDeferredItems(ctx, params.DB, siteID, pageName, workItemID, newDeferred, logger)
		}
	}

	// Build section names lists for the content writer
	readyNames := make([]string, len(ready))
	for i, s := range ready {
		readyNames[i] = s.Name
	}

	// Persist the skip decisions on the page row (bugs_open/040 skip-not-recorded;
	// council corr 164058e6). Until this write, an on_missing=skip_section outcome
	// lived only in this action's result — collected_data, which database-cleanup
	// prunes — so pages.sections permanently promised sections the platform had
	// deliberately declined to render, and the 040 partial-build guard
	// (UpdatePageStatusAction) read every data-gated skip as a build shortfall and
	// parked the page at needs_rebuild on every rebuild. The merge is symmetric:
	// skipped names are added to pages.suppressed_sections (which every
	// completeness reader already excludes) and names that planned ready this
	// build are removed, so a section whose data later arrives is un-suppressed
	// the same build it renders. Deferred sections are deliberately NOT
	// suppressed — their debt stays visible via their needs_human_review item.
	if params.DB != nil && (len(skipped) > 0 || len(ready) > 0) {
		skippedNames := make([]string, len(skipped))
		for i, s := range skipped {
			skippedNames[i] = s.Name
		}
		if persistErr := persistSectionSkips(ctx, params.DB, siteID, pageName, readyNames, skippedNames, logger); persistErr != nil {
			// Warn-not-fail for the BUILD, but the failure itself must be
			// durable: a skip decision that silently fails to persist is the
			// same vanishing-record defect this write exists to fix (council
			// 164058e6, bug_historian). agent_error_log is the durable channel;
			// if even that insert fails, the log line is the last resort.
			logger.Warn("plan_sections: persistSectionSkips failed — skip decisions not durably recorded this build",
				zap.String("page", pageName),
				zap.Strings("skipped", skippedNames),
				zap.Error(persistErr))
			// Running step is the right provenance — declared, not inherited
			// silently (see LogActionEntry's doc).
			LogActionEntryInheritingProvenance(ctx, params, agenterrors.Entry{
				SiteID:       siteID.String(),
				Action:       "plan_sections",
				ErrorMessage: "persistSectionSkips failed: " + persistErr.Error(),
				ErrorCode:    "SKIP_PERSISTENCE_FAILED",
				Severity:     "warning",
				Context: map[string]interface{}{
					"page_name":      pageName,
					"skipped":        skippedNames,
					"ready":          readyNames,
					"remedy":         "the page's suppressed_sections was not updated; the 040 guard may refuse its deploy stamp until the next successful plan — see bugs_open/040 skip-not-recorded",
					"council_review": "164058e6-4630-47a2-b0d7-58659997b291",
				},
			}, logger)
		}
	}

	// bugs_open/238 visibility. Both maps are built from the plan items, so they
	// cover deferred and skipped sections too — a section that carried a key and
	// was then deferred still carried it.
	carriedBySection := collectCarriedFields(ready, deferred, skipped)
	missesBySection := collectStructuralMisses(ready, deferred, skipped)
	recordStructuralKeyCarryMisses(ctx, params, siteID.String(), pageName, missesBySection)

	logger.Info("plan_sections: planning complete",
		zap.Int("ready", len(ready)),
		zap.Int("deferred", len(deferred)),
		zap.Int("skipped", len(skipped)),
		zap.Int("sections_with_carried_keys", len(carriedBySection)),
		zap.Int("sections_with_structural_misses", len(missesBySection)),
		zap.String("page", pageName))

	return map[string]interface{}{
		"sections_ready":    ready,
		"sections_deferred": deferred,
		"sections_skipped":  skipped,
		"ready_names":       readyNames,
		"ready_count":       len(ready),
		"deferred_count":    len(deferred),
		"skipped_count":     len(skipped),
		"total_sections":    len(sectionNames),
		// Which fields resolved from somewhere other than the path they declare
		// (PBP-026). Asked for by the council gate's bug_historian seat on corr
		// dd03a73b: a zap line is not a queryable record of a section whose data
		// provenance changed, and this platform's own history says logs are not a
		// reliable substitute for a structured signal when auditing that. Absent
		// from the payload entirely when nothing aliased, so its presence is the
		// signal. Empty for every build on a site whose specs are populated.
		"source_aliases_used": resolver.aliasesUsed,
		// Which resolver-sourced keys this page's OWN stored content_data had to
		// supply, and which resolved nowhere (bugs_open/238). Same reasoning as
		// source_aliases_used above and the same shape: nil when nothing
		// happened, which marshals to JSON null rather than an empty object — so
		// a quiet build cannot be misread as "we looked and found none".
		"structural_keys_carried": carriedBySection,
		"structural_key_misses":   missesBySection,
	}, nil
}

// collectCarriedFields maps section name → the non-llm fields that section had
// to take from the page's own stored content_data (bugs_open/238). Nil when no
// section carried anything, so a quiet build does not emit an empty object that
// reads like "we looked and found none".
func collectCarriedFields(groups ...[]sectionPlanItem) map[string][]string {
	var out map[string][]string
	for _, group := range groups {
		for _, item := range group {
			if len(item.CarriedFields) == 0 {
				continue
			}
			if out == nil {
				out = make(map[string][]string)
			}
			out[item.Name] = item.CarriedFields
		}
	}
	return out
}

// collectStructuralMisses maps section name → the required non-llm fields that
// resolved neither from their declared source nor from the stored row.
func collectStructuralMisses(groups ...[]sectionPlanItem) map[string][]missingField {
	var out map[string][]missingField
	for _, group := range groups {
		for _, item := range group {
			if len(item.StructuralMisses) == 0 {
				continue
			}
			if out == nil {
				out = make(map[string][]missingField)
			}
			out[item.Name] = item.StructuralMisses
		}
	}
	return out
}

// recordStructuralKeyCarryMisses persists one durable row per section holding a
// REQUIRED non-llm field that resolved nowhere — not from its declared source,
// and not from the page's own deployed content_data — and was then silently
// omitted under on_missing=skip_field.
//
// Durable rather than a log line, for the reason recordFactCarryMisses gives one
// feature over: a plan produced by this omission is byte-identical to a plan for
// a component that never declared the field, so nothing downstream can tell the
// two apart afterwards. bugs_open/238 served five empty <img src=""> and six
// vanished controls for two days with no queryable record of why, while the Info
// log naming the fields had long since rotated away. agent_error_log is the
// channel this action already uses (FACT_SCOPING_EMPTY_COMPOSITION,
// SKIP_PERSISTENCE_FAILED), so "did this page lose a structural key?" becomes one
// error_code query.
//
// Best-effort by construction: a failed write must not change a plan the action
// has already decided on.
func recordStructuralKeyCarryMisses(ctx context.Context, params ActionParams, siteID, pageName string, misses map[string][]missingField) {
	if len(misses) == 0 {
		return
	}
	// Deterministic order: sections arrive from a map, and a findings batch whose
	// row order changes run to run is needlessly hard to diff.
	sections := make([]string, 0, len(misses))
	for section := range misses {
		sections = append(sections, section)
	}
	sort.Strings(sections)

	findings := make([]agenterrors.Finding, 0, len(sections))
	for _, section := range sections {
		fields := misses[section]
		names := make([]string, len(fields))
		sources := make([]string, len(fields))
		for i, f := range fields {
			names[i] = f.Field
			sources[i] = f.Source
		}
		findings = append(findings, agenterrors.Finding{
			ErrorCode: "STRUCTURAL_KEY_CARRY_MISS",
			Severity:  "warning",
			Message: fmt.Sprintf("page %q section %q: %d required non-llm field(s) resolved from neither their declared source nor the page's stored content_data — omitted under on_missing=skip_field",
				pageName, section, len(fields)),
			Context: map[string]interface{}{
				"page_name": pageName,
				"section":   section,
				"fields":    names,
				"sources":   sources,
				"remedy":    "the declared source resolves nothing for this site and no previously-built row held a value, so there was nothing to carry. Populate the source (the site_specs aspect, or the site_plan_imagery row behind site_assets.*) or gate the template on the field. The section built without it — an ungated {{.field}} inside src=/href= ships an empty attribute (bugs_open/238)",
			},
		})
	}
	LogActionFindings(ctx, params, siteID, "", "plan_sections", findings, params.Logger)
}

// ============================================================================
// Load component schemas by section name
// ============================================================================

type componentInfo struct {
	ID          string
	Name        string
	Function    string
	InputSchema map[string]interface{}
	// Raw carries the full per-section component map produced by the
	// shared loadSectionComponents loader. Plan_sections attaches this
	// onto sectionPlanItem.Component so downstream consumers can read
	// html_template, render_mode, description, category, etc. without
	// re-loading from content_components. Step 3 swaps the page-content-
	// writer over to consume this directly.
	Raw map[string]interface{}
}

// loadComponentSchemas is a thin wrapper over the shared loadSectionComponents
// helper. It converts the helper's per-component maps into componentInfo
// records keyed by both name and function (the lookup pattern planSection
// expects), parses input_schema JSON for the field-resolution walk, and
// applies the template-truncation guard that the previous in-line SQL did.
//
// Note: plan_sections doesn't have a pageID at this point in the workflow,
// so brief enrichment is skipped here. Briefs apply at content-write time
// (page-content-writer's load path) where pageID is known. Step 3 may move
// brief loading into the section_plan so plan_sections becomes the single
// source for all per-section content.
func loadComponentSchemas(ctx context.Context, db *sql.DB, sectionNames []string, logger *zap.Logger) map[string]componentInfo {
	result := make(map[string]componentInfo)

	// activeOnly=true preserves the historical is_active=true filter that the
	// inline SQL had. Inactive components stay out of plan_sections so they
	// flow to Path 2 (selector) and may be replaced by a current alternative.
	components := loadSectionComponents(ctx, db, sectionNames, "", true, logger)

	for _, comp := range components {
		ci, ok := componentInfoFromRaw(comp, "plan_sections", logger)
		if !ok {
			continue
		}

		// Index by both name and function for fast lookup in the section loop.
		if ci.Name != "" {
			result[ci.Name] = ci
		}
		if ci.Function != "" && ci.Function != ci.Name {
			result[ci.Function] = ci
		}
	}

	// Alias each raw requested name to its resolved component when the plan asked
	// in snake_case/CamelCase but the component is stored kebab-case
	// (bugs_open/041). loadSectionComponents now resolves such a section, but this
	// map is keyed by the STORED name/function; callers look it up by the
	// REQUESTED name (plan_sections' section loop, rerender's slot_name). Without
	// the alias a "call_to_action" request misses the "call-to-action" entry and
	// falls through to a spurious needs_new_component.
	aliasNormalisedSectionKeys(result, sectionNames)

	return result
}

// componentInfoFromRaw converts one raw per-component map (the shape both
// loadSectionComponents and loadContentComponentsByID produce, scanned by the
// shared scanSectionComponentRow) into a componentInfo, applying the
// template-truncation guard so a truncated/broken template is dropped
// identically regardless of which lookup found the row.
//
// Factored out so the three places that used to do this conversion inline
// (loadComponentSchemas, loadSingleComponentSchema, and now
// loadComponentSchemasByID) cannot drift out of step with each other the way
// bugs_open/024 recorded happening once already.
//
// callerCtx prefixes the truncation-drop log line, so a drop on the rerender
// path is not misattributed to plan_sections in the evidence trail.
//
// ok is false in two cases the caller cannot tell apart from the bool alone:
// the raw map is a name-stub (no component_id — the requested section matched
// no row at all), or the row matched but componentTemplateValid rejected its
// template. loadComponentSchemasByID reports which one via its drops map;
// the two existing callers only ever needed "usable or not".
func componentInfoFromRaw(comp map[string]interface{}, callerCtx string, logger *zap.Logger) (componentInfo, bool) {
	// Stubs have no component_id — nothing matched the requested name/id at all.
	if _, hasID := comp["component_id"]; !hasID {
		return componentInfo{}, false
	}

	// Template truncation guard: components with HTML content but no closing
	// </section> tag are treated as broken and dropped so they flow to the
	// caller's fallback path instead of rendering broken markup. Empty/very-
	// short templates are NOT dropped — they may be intentional stubs.
	//
	// component_level='tool' templates get their own check: a tool is
	// self-contained HTML, not a <section> wrapper, so the '</section>'
	// marker is the wrong truncation signal in BOTH directions there — it
	// dropped healthy tools that end '</script>' (a durable tool fix could
	// never re-render, bugs_open/024) and passed truncated ones that happen
	// to contain '</section>' upstream of the cut.
	htmlTpl, _ := comp["html_template"].(string)
	level, _ := comp["component_level"].(string)
	if !componentTemplateValid(htmlTpl, level) {
		name, _ := comp["name"].(string)
		function, _ := comp["function"].(string)
		// The MEASURED signals ride the log line (council 70cf0da5 round 2,
		// bug_historian advisory: a load-time drop must be diagnosable from
		// the evidence trail, not just visible). The durable detector for
		// this class is the truncated_component discovery sweep, which files
		// the human-review work item; this Warn is the load-time echo.
		logger.Warn(callerCtx+": component template structurally incomplete, skipping (falls back to stored HTML; the truncated_component sweep owns the durable finding)",
			zap.String("function", function),
			zap.String("name", name),
			zap.String("component_level", level),
			zap.Strings("unbalanced_markup_context", content.UnbalancedStructuralTags(htmlTpl)),
			zap.Bool("ends_cleanly", endsCleanly(htmlTpl)))
		return componentInfo{}, false
	}

	var ci componentInfo
	if id, ok := comp["component_id"].(string); ok {
		ci.ID = id
	}
	if name, ok := comp["name"].(string); ok {
		ci.Name = name
	}
	if fn, ok := comp["function"].(string); ok {
		ci.Function = fn
	}
	if schemaStr, ok := comp["input_schema"].(string); ok && schemaStr != "" {
		if err := json.Unmarshal([]byte(schemaStr), &ci.InputSchema); err != nil {
			logger.Warn(callerCtx+": failed to parse input_schema",
				zap.String("component", ci.Name),
				zap.Error(err))
		}
	}
	if ci.InputSchema == nil {
		ci.InputSchema = make(map[string]interface{})
	}
	ci.Raw = comp
	return ci, true
}

// loadComponentSchemasByID resolves components by their OWN identity
// (page_components.component_id) rather than by name/function — the repair
// half of bugs_open/182. A slot_name is only ever a naming convention; a
// site whose slots are positional ("prose-0", "tool-2") has a slot_name that
// is not, and never will be, any component's name/function, so
// loadComponentSchemas can never resolve it. The row already knows exactly
// which component it is; this is what reads that identity instead of
// re-deriving it from a name that may not carry it.
//
// Returns the resolved map keyed by component id, plus a drops map (id ->
// reason, currently only "invalid_template") naming every requested id that
// has a row but was rejected by the template guard — as opposed to an id
// simply absent from the result (not found at all, or is_active=false),
// which the caller cannot distinguish from "no such id" and does not need to:
// both mean "resolve some other way or carry the stored HTML".
func loadComponentSchemasByID(ctx context.Context, db *sql.DB, componentIDs []string, logger *zap.Logger) (map[string]componentInfo, map[string]string) {
	result := make(map[string]componentInfo)
	drops := make(map[string]string)
	if len(componentIDs) == 0 {
		return result, drops
	}

	components := loadContentComponentsByID(ctx, db, componentIDs, logger)
	for _, comp := range components {
		id, _ := comp["component_id"].(string)
		ci, ok := componentInfoFromRaw(comp, "rerender_page_sections(by_id)", logger)
		if !ok {
			// componentInfoFromRaw only fails the template guard here — every row
			// this loop sees came back from a WHERE id = ANY(...) query, so it
			// always carries its own component_id and can never be a name-stub.
			if id != "" {
				drops[id] = "invalid_template"
			}
			continue
		}
		result[ci.ID] = ci
	}
	return result, drops
}

// loadPageSlotComponentIDs reads the page's stored page_components rows and
// returns slot_name → component_id — the identity map the build path needs to
// resolve a POSITIONAL slot name ("prose-0") that is no component's
// name/function (bugs_open/204, the build-path half of bugs_open/182). The
// page is identified by (site_id, name), which pages_site_id_name_key makes
// unique; plan_sections has no page_id in its inputs, but it has always had
// these two.
//
// MOVED to datahelpers 2026-08-21 and this function is now a delegation, not a
// second implementation. A third call site of the same judgement (validate_plan's
// validate_components arm) was found deleting the very names this resolves, and
// the estate had by then grown three private loaders for one question. The load,
// the conflict rule and the log/error strings are unchanged — see
// datahelpers/page_slot_identities.go for the rule and for the consumer list.
//
// A slot_name repeated across rows is NORMAL (generic-text-block used 2-3×
// on one page — measured fleet-wide, 11 legitimate pages; see LANDMINES.md
// "Deduplicating page_components…"). Repeats agreeing on component_id map
// fine; repeats DISAGREEING drop that slot from the map with a warning, so
// resolution falls back to the name path rather than picking a row
// arbitrarily.
//
// No rows (initial build — the page or its components don't exist yet) is
// normal and returns an empty map. A query error is returned to the caller:
// planning against a silently-empty map on a decomposed site files junk
// needs_new_component items (two per section, measured on the 204 canary),
// so a loud transient failure is the cheaper outcome.
func loadPageSlotComponentIDs(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName string, logger *zap.Logger) (map[string]string, error) {
	rows, err := datahelpers.LoadPageSlotRows(ctx, db, siteID, pageName)
	if err != nil {
		return nil, err
	}
	return datahelpers.SlotIDMap(rows, "plan_sections", pageName, logger), nil
}

// aliasNormalisedSectionKeys adds, for each requested section name that is not
// already a key, an alias to the entry stored under its kebab-normalised form.
// Strict superset — it only ADDS keys and never rebinds an existing one, so a
// component whose own name is snake_case (keyed by its raw name) is untouched.
// See /bugs_open/041 (section-lookup-never-normalises).
func aliasNormalisedSectionKeys(result map[string]componentInfo, sectionNames []string) {
	for _, name := range sectionNames {
		if _, ok := result[name]; ok {
			continue
		}
		norm := NormalizeComponentFunction(name)
		if norm == name {
			continue
		}
		if ci, ok := result[norm]; ok {
			result[name] = ci
		}
	}
}

// sectionTemplateValid answers "was this section-level template CUT MID-STREAM?"
//
// ── IT USED TO ASK A DIFFERENT QUESTION, AND GOT IT WRONG 22 TIMES (bugs_open/351)
//
// It mirrored the original SQL CASE, whose "invalid" arm was `html_template NOT
// LIKE '%</section>%'` — using a WRAPPER TAG as a proxy for "not truncated".
// That proxy is wrong for any self-contained widget: a calculator is a
// <div>-wrapped tool with its own <script>, and never contains </section> at
// all. Measured 2026-08-21 over every active section-level calculator
// (**22** as of that date): the old predicate passed **0** of them, while all
// **22** were structurally complete. Consequence: the component selector could
// never reuse a calculator the library already owned, so every site that wanted
// one paid a fresh LLM generation — which is how remortgagecalculator.uk ended
// up in bugs_open/345's retry loop.
//
// This is bugs_open/024's defect one level over, exactly as componentTemplateValid's
// header predicted the class would recur: that fix gave component_level='tool'
// a structural predicate and left 'section' on the marker.
//
// ── WHY IT IS A COPY OF toolTemplateValid RATHER THAN A CALL ───────────────
//
// The two now ask the same question and deliberately agree. They are kept as
// separate functions because componentTemplateValid dispatches on
// component_level and a future divergence (a section-only rule) should have a
// place to live without disturbing tools. If they are still identical when
// someone next reads this, collapsing them is safe — the calibration below is
// what makes that judgement, not preference.
//
// Empty/short templates are still allowed through: they may be intentional
// stubs, and that arm is unchanged.
//
// ── CALIBRATED BOTH DIRECTIONS BEFORE THE CHANGE ──────────────────────────
//
// Over all active section-level templates (**150** as of 2026-08-22): 22
// rescued, **0** regressed. The single regression seen in the 2026-08-21 dry
// run was a FALSE one — a conditionally-wrapped section ending `</section>{{end}}`
// — and is fixed by the matching endsCleanly change rather than accepted.
// ⚠ The marker misclassified BOTH ways, so re-calibrating after any future edit
// must assert the flip SET by id, never a count: a different single row flipping
// hides inside an unchanged count of one.
func sectionTemplateValid(htmlTemplate string) bool {
	if htmlTemplate == "" {
		return true
	}
	if len(htmlTemplate) < 100 {
		return true
	}
	// bugs_open/351: the test is STRUCTURAL, not a wrapper-tag substring.
	// Identical to toolTemplateValid's body — see the header for why the two
	// deliberately agree rather than being collapsed into one function.
	if len(content.UnbalancedStructuralTags(htmlTemplate)) > 0 {
		return false
	}
	return endsCleanly(htmlTemplate)
}

// componentTemplateValid is THE truncation gate for a loaded component, and the
// only one either loader should call.
//
// It exists because there are TWO call sites making this identical judgement —
// loadComponentSchemas (the bulk loader) and loadSingleComponentSchema (the
// by-function loader) — and the first fix for bugs_open/024 patched only the
// bulk one. The council's bug_historian seat predicted the second call site from
// this platform's documented history of the same filter existing twice, and it
// was right: `loadSingleComponentSchema` was still rejecting self-contained tool
// templates on the '</section>' marker and returning nil, silently.
//
// Both loaders now share this predicate so the two cannot drift again. A
// component that is dropped here is invisible downstream — no error, no work
// item — which is what made the original defect cost three fix cycles.
func componentTemplateValid(htmlTemplate, componentLevel string) bool {
	if componentLevel == "tool" {
		return toolTemplateValid(htmlTemplate)
	}
	return sectionTemplateValid(htmlTemplate)
}

// toolTemplateValid is the truncation guard for component_level='tool'
// templates. A tool is self-contained HTML, not a <section> wrapper, so
// sectionTemplateValid's '</section>' marker misclassifies tools in both
// directions: healthy tools ending '</script>' read as truncated (and were
// silently dropped from the schemas map, so a durable tool fix could never
// re-render — bugs_open/024), while genuinely cut templates that contain
// '</section>' upstream of the cut read as whole.
//
// Reuses the component write guard's absolute structural signals instead:
// every paired tag balanced, and the template ends on a closed tag.
//
// Calibrated against all 27 active tool components on 2026-07-20: the 19
// structurally whole templates pass; the 8 truncated rows all fail (each cut
// mid-JavaScript by a pre-guard truncation write, bugs_open/012's class —
// four of which contain '</section>' and so pass sectionTemplateValid today).
// Rejecting those here is load-bearing: it keeps the re-render loop on the
// carry-stored-HTML path for a damaged template instead of deploying broken
// markup from it.
//
// Since bugs_open/303 the tag balance is MARKUP-CONTEXT (content.
// UnbalancedStructuralTags), not a substring count: a tool that MANIPULATES
// HTML mentions tags in its own JavaScript (a comment, a regex
// /<script[^>]*>/), and the substring count read one unpaired mention as a
// truncation — refusing the tool at birth here via create_tool_component, and
// silently dropping it from the schemas map at load (this function's other
// caller), so an affected tool could never re-render. Recalibrated against the
// full live population 2026-08-18: every known casualty still fails, and the
// HTML-manipulating tools that passed by zero margin now pass with headroom.
func toolTemplateValid(htmlTemplate string) bool {
	if htmlTemplate == "" {
		return true
	}
	if len(htmlTemplate) < 100 {
		return true
	}
	if len(content.UnbalancedStructuralTags(htmlTemplate)) > 0 {
		return false
	}
	return endsCleanly(htmlTemplate)
}

// ============================================================================
// Section type resolution via component selector
// ============================================================================

// resolveSectionComponent attempts to find a component for a section name
// that didn't match any function directly. It queries the component selector
// by section_type, which scores candidates against the site/page context.
//
// Returns the resolved componentInfo and a resolution status string:
//   - "selected": selector found a match
//   - "not_found": no components with this section_type exist
//   - "selector_error": DB query failed
func resolveSectionComponent(
	ctx context.Context,
	db *sql.DB,
	sectionName string,
	selCtx SelectorContext,
	logger *zap.Logger,
) (*componentInfo, string) {

	candidate, err := SelectComponentByType(ctx, db, sectionName, selCtx, logger)
	if err != nil {
		logger.Warn("plan_sections: selector query failed",
			zap.String("section", sectionName),
			zap.Error(err))
		return nil, "selector_error"
	}

	if candidate == nil {
		return nil, "not_found"
	}

	// Selector found a match — load the full component info including input_schema.
	// The selector only returns metadata; we need the schema for field resolution.
	comp := loadSingleComponentSchema(ctx, db, candidate.Function, logger)
	if comp == nil {
		logger.Warn("plan_sections: selector matched but component load failed",
			zap.String("section_type", sectionName),
			zap.String("function", candidate.Function))
		return nil, "selector_error"
	}

	// NO usage counter is incremented here, deliberately (bugs_open/378).
	// This is the section_type selector — one of THREE resolution paths in the
	// section loop above (stored component_id, name/function match, and this one).
	// The old IncrementUsageCount lived at exactly this line, so the column it
	// wrote recorded which ROUTE resolved a component rather than whether the
	// component is any good, and it fired HERE — before planSection has decided
	// ready/deferred/skipped and before any page_components row exists — so it
	// also counted resolutions that never became a binding at all.
	// "How proven is this component" is now derived from page_components at read
	// time: see ComponentUsageSitesSQL in component_selector.go.

	logger.Info("plan_sections: resolved via section_type selector",
		zap.String("section_type", sectionName),
		zap.String("resolved_function", candidate.Function),
		zap.Float64("score", candidate.Score))

	return comp, "selected"
}

// loadSingleComponentSchema loads one component's schema by function name.
// Thin wrapper over the shared loadSectionComponents helper. Rejects components
// with truncated templates (missing </section>) so they don't render broken HTML.
// Always uses activeOnly=true to match the original is_active filter.
func loadSingleComponentSchema(ctx context.Context, db *sql.DB, function string, logger *zap.Logger) *componentInfo {
	components := loadSectionComponents(ctx, db, []string{function}, "", true, logger)

	for _, raw := range components {
		// Stubs from the helper (no component_id) mean the function wasn't found.
		if _, hasID := raw["component_id"]; !hasID {
			continue
		}

		// componentInfoFromRaw logs its own drop; preserve this function's
		// original behaviour of failing hard on an invalid template rather than
		// trying another candidate (loadComponentSchemas' bulk loop tries every
		// candidate instead — a deliberate difference kept by not sharing this
		// control flow, only the conversion+guard).
		ci, ok := componentInfoFromRaw(raw, "loadSingleComponentSchema", logger)
		if !ok {
			return nil
		}
		return &ci
	}

	logger.Warn("loadSingleComponentSchema: function not found",
		zap.String("function", function))
	return nil
}

// queryResultLen returns the element count of a resolved query value when it is
// a list, and reports whether it was a list at all. Query list resolvers return
// []map[string]interface{}; the []interface{} case is defensive. Scalars (e.g. a
// URL string from section_index_for) are not lists and never fail a cardinality
// contract here.
func queryResultLen(value interface{}) (int, bool) {
	switch v := value.(type) {
	case []map[string]interface{}:
		return len(v), true
	case []interface{}:
		return len(v), true
	default:
		return 0, false
	}
}

// queryListBelowContract reports whether a successfully-resolved query value is a
// LIST that fails the field's declared cardinality contract: fewer items than
// min_items, or empty when the field is required. A required list with no explicit
// min_items is treated as needing at least one item. Non-list values (scalars,
// nil) never fail here. This is what makes a query-sourced field honour its
// required/min_items/on_missing declaration instead of silently accepting an empty
// slice (bugs_open/054 fix-candidate 2).
func queryListBelowContract(value interface{}, required bool, minItems int) bool {
	n, isList := queryResultLen(value)
	if !isList {
		return false
	}
	floor := minItems
	if required && floor < 1 {
		floor = 1
	}
	return n < floor
}

// ============================================================================
// Plan a single section
// ============================================================================

func planSection(ctx context.Context, sectionName string, section sectionRef, comp componentInfo, resolver *sourceResolver, logger *zap.Logger) sectionPlanItem {
	item := sectionPlanItem{
		Name:        sectionName,
		ComponentID: comp.ID,
		Function:    comp.Function,
		Status:      "ready",
		// Attach the full component data so downstream consumers (Step 3:
		// page-content-writer) can read input_schema, html_template,
		// render_mode, description, category, content_brief etc. without
		// re-loading via load_page_section_components. Nil for sections
		// where no component was resolved.
		Component: comp.Raw,
	}

	// Get fields from schema via datahelpers.SchemaContentFields so a component in
	// the legacy JSON-Schema dialect (`properties`+`required[]`, no `fields`) has
	// its fields planned for — before bugs_open/026 a missed dialect fell through
	// to "all fields from LLM" with no field specs, so a required field the writer
	// was never told about (the news-listing headline) was never generated. This
	// is the generation tripwire: a projected legacy dialect is extinct fleet-wide,
	// so WarnLegacyDialect surfaces a re-seed/restore/creator regression here at
	// build time (the earliest point every component passes through).
	fieldsRaw, ok, fromLegacy := datahelpers.SchemaContentFields(comp.InputSchema)
	if fromLegacy {
		datahelpers.WarnLegacyDialect(logger, "plan_sections", comp.Function)
	}
	if !ok || len(fieldsRaw) == 0 {
		// A self-contained TOOL component legitimately has an empty input_schema:
		// its HTML renders entirely from its own template, with no LLM-authored
		// content fields to supply. Exempt it by the SAME explicit
		// component_level='tool' marker the rerender escalation guard uses
		// (isSelfContainedSection), NEVER by the name heuristic below. Without
		// this, a future tool whose Function name happens to contain
		// "content"/"body"/"article"/… would be marked `deferred` here — carried
		// unchanged, so a durable template fix is computed and silently discarded
		// — the identical end-state as bugs_open/024, reached one function away by
		// a different route (bugs_open/044). Two call sites of the "is this
		// emptiness legitimate?" judgement, now one shared predicate so they
		// cannot drift apart again.
		if isSelfContainedSection(comp) {
			item.Reason = "self-contained tool component — renders from its own template, no content fields"
			return item
		}

		// No v2 schema — component has no declared content fields.
		// Check if the template has actual HTML structure. If it's CSS-only
		// or truncated, it was likely created by a broken component-creator
		// run and will produce empty/broken output.
		if comp.Function != "" {
			funcLower := strings.ToLower(comp.Function)
			// Components that typically need LLM content should not have empty schemas
			needsContent := strings.Contains(funcLower, "article") ||
				strings.Contains(funcLower, "content") ||
				strings.Contains(funcLower, "body") ||
				strings.Contains(funcLower, "text") ||
				strings.Contains(funcLower, "blog")
			if needsContent {
				item.Status = "deferred"
				item.Reason = "component has empty input_schema — needs regeneration with content fields"
				logger.Warn("plan_sections: content component has empty schema, deferring",
					zap.String("function", comp.Function),
					zap.String("section", sectionName))
				return item
			}
		}
		// Non-content components with empty schema (e.g. decorative sections,
		// separators) — treat as fully LLM-generated for backward compat
		item.Reason = "no field schema — all fields from LLM"
		return item
	}

	resolvedData := make(map[string]interface{})
	var llmFields []string
	var llmFieldSpecs []llmFieldSpec
	var missingFields []missingField
	// carriedFields and structuralMisses are the two halves of bugs_open/238's
	// visibility: which resolver-sourced fields the page's own stored row had to
	// supply, and which resolved nowhere at all. Both are surfaced on the plan
	// item, because a carry that silently worked and a carry that silently found
	// nothing produce identical plans otherwise.
	var carriedFields []string
	var structuralMisses []missingField
	shouldSkip := false
	shouldDefer := false

	for fieldName, fieldDefRaw := range fieldsRaw {
		fieldDef, ok := fieldDefRaw.(map[string]interface{})
		if !ok {
			continue
		}

		source, _ := fieldDef["source"].(string)
		required, _ := fieldDef["required"].(bool)
		onMissing, _ := fieldDef["on_missing"].(string)
		if onMissing == "" {
			onMissing = "skip_field"
		}
		fallback := fieldDef["fallback"]
		missingReason, _ := fieldDef["missing_reason"].(string)

		// Extract type info for enriched HITL work items
		fieldType, _ := fieldDef["type"].(string)
		fieldItems, _ := fieldDef["items"].(map[string]interface{})
		fieldMinItems := 0
		if mi, ok := fieldDef["min_items"].(float64); ok {
			fieldMinItems = int(mi)
		}

		// carryStored satisfies a non-llm field from the page's OWN deployed
		// content_data when the declared source yielded nothing (bugs_open/238).
		//
		// Why this exists: a regeneration replaces the row wholesale
		// (save_page_sections DELETEs and re-INSERTs), so a resolver-sourced key
		// that fails to resolve on THIS run is destroyed rather than left alone —
		// while the re-render path, which merges stored ⊕ fresh, has always kept
		// it. That asymmetry is the bug: finetuning.uk's homepage lost all 11 of
		// its non-llm URL keys to one tone_shift and served five <img src=""> plus
		// six vanished controls, after the same values had survived every
		// re-render since 2026-05-01.
		//
		// Live resolution always wins — this runs only after the literal path and
		// every alias have missed — so a repaired source takes precedence on the
		// next build and a carried value cannot outlive its source becoming
		// resolvable. The carried value lands in resolvedData, which
		// RenderComponentAction overlays onto the LLM output (merge_with), so it
		// reaches the rendered HTML and the persisted row through machinery that
		// already exists (PBP-014). No new seam downstream.
		//
		// LLM fields are deliberately NOT carried: a tone_shift rewriting the copy
		// is the regeneration working, and an llm field missing from the writer's
		// output is a different failure with its own guard
		// (missingRequiredLLMFields refuses the render).
		carryStored := func() bool {
			if source == "" || source == "llm" {
				return false
			}
			value, ok := resolver.storedFieldValue(ctx, sectionName, fieldName)
			if !ok {
				return false
			}
			resolvedData[fieldName] = value
			carriedFields = append(carriedFields, fieldName)
			logger.Info("plan_sections: non-llm field carried from stored content_data — declared source resolved nothing",
				zap.String("field", fieldName),
				zap.String("source", source),
				zap.String("section", sectionName),
				zap.String("page", resolver.pageName))
			return true
		}

		// handleMissingField applies the field's declared on_missing policy when
		// its source yielded no usable data — not found, nil, or (for a
		// query-sourced list) an empty/short result that fails the field's
		// required/min_items contract. Defined here so the generic and query.*
		// resolution paths share ONE policy implementation and cannot drift — the
		// drift class bugs_closed/044 was closed for. See bugs_open/054.
		handleMissingField := func() {
			// The carry runs first, at the one point both resolution paths pass
			// through, so it cannot drift from on_missing for the same reason
			// on_missing cannot drift from itself.
			if carryStored() {
				return
			}
			if !required {
				// Optional field missing — apply on_missing
				switch onMissing {
				case "use_fallback":
					if fallback != nil {
						resolvedData[fieldName] = fallback
					}
				case "skip_field":
					// Just omit it
				case "skip_section":
					shouldSkip = true
				case "needs_human_review":
					shouldDefer = true
					missingFields = append(missingFields, missingField{
						Field:     fieldName,
						Source:    source,
						OnMissing: onMissing,
						Reason:    missingReason,
						Type:      fieldType,
						Items:     fieldItems,
						MinItems:  fieldMinItems,
					})
				}
				return
			}

			// Required field missing — apply on_missing
			switch onMissing {
			case "skip_field":
				// Required-but-skippable: honour the schema's declared intent and
				// omit the field instead of deferring the section (mirrors the
				// optional branch; templates gate on the field).
				//
				// That premise is not always true — case-studies-grid renders
				// `src="{{.card1_image_url}}"` at root scope with no gate, so
				// omitting the field ships an empty attribute rather than nothing
				// (bugs_open/238). The behaviour here is unchanged on purpose:
				// deferring instead would be RFC_009's option A, which the owner
				// did NOT take (~90% of fields declare no on_missing, so a
				// declaration-driven gate would block sections fleet-wide while
				// being inert for nine fields in ten). What changed is upstream —
				// the field has already been offered the carry — so reaching here
				// with a non-llm source means it resolved NOWHERE, and that is
				// now recorded durably instead of evaporating into an Info log.
				logger.Info("plan_sections: required field missing with on_missing=skip_field — omitting field",
					zap.String("field", fieldName),
					zap.String("source", source))
				if source != "" && source != "llm" {
					structuralMisses = append(structuralMisses, missingField{
						Field:     fieldName,
						Source:    source,
						OnMissing: onMissing,
						Reason:    missingReason,
						Type:      fieldType,
						Items:     fieldItems,
						MinItems:  fieldMinItems,
					})
				}
			case "use_fallback":
				if fallback != nil {
					resolvedData[fieldName] = fallback
				} else {
					// Required with no fallback — defer
					shouldDefer = true
					missingFields = append(missingFields, missingField{
						Field:     fieldName,
						Source:    source,
						OnMissing: onMissing,
						Reason:    missingReason,
						Type:      fieldType,
						Items:     fieldItems,
						MinItems:  fieldMinItems,
					})
				}
			case "skip_section":
				shouldSkip = true
			case "needs_human_review":
				shouldDefer = true
				missingFields = append(missingFields, missingField{
					Field:     fieldName,
					Source:    source,
					OnMissing: onMissing,
					Reason:    missingReason,
					Type:      fieldType,
					Items:     fieldItems,
					MinItems:  fieldMinItems,
				})
			case "block":
				// Block entire page build — this is handled upstream
				shouldDefer = true
				missingFields = append(missingFields, missingField{
					Field:     fieldName,
					Source:    source,
					OnMissing: "block",
					Reason:    missingReason,
					Type:      fieldType,
					Items:     fieldItems,
					MinItems:  fieldMinItems,
				})
			default:
				// Unknown on_missing — default to defer for safety
				shouldDefer = true
				missingFields = append(missingFields, missingField{
					Field:     fieldName,
					Source:    source,
					OnMissing: onMissing,
					Reason:    missingReason,
					Type:      fieldType,
					Items:     fieldItems,
					MinItems:  fieldMinItems,
				})
			}
		}

		// LLM-generated fields — always available
		if source == "llm" {
			llmFields = append(llmFields, fieldName)
			// bugs_open/437: the nested half of the same question ItemFields
			// answers flatly. Zero values for every field whose elements are
			// scalars, so the emitted spec — and the prompt built from it — is
			// unchanged for all but the nesting components.
			valueShape, itemNotes := datahelpers.StructuredItemShape(fieldDef)
			llmFieldSpecs = append(llmFieldSpecs, llmFieldSpec{
				Name:        fieldName,
				Type:        fieldType,
				Required:    required,
				Description: stringOrEmpty(fieldDef["llm_guidance"]),
				OnMissing:   onMissing,
				Fallback:    fallback,
				ItemFields:  extractArrayItemFields(fieldDef),
				ValueShape:  valueShape,
				ItemNotes:   itemNotes,
			})
			continue
		}

		// Query.* fields — resolve via the queryresolve package.
		// The query resolver runs SQL against the database and returns
		// data shaped for the field. For array-typed fields (lists, grids,
		// directories) the result is []map[string]interface{} that the
		// downstream content writer / template renderer iterates over.
		//
		// Failure handling:
		//   - Unknown query name → log warning, fall through to fallback/skip
		//   - DB error → log warning, fall through to fallback/skip
		//   - Empty result → put empty slice in resolvedData (the component's
		//     html_template should handle empty lists; on_missing applies if
		//     the field is required and the schema treats empty as missing)
		if strings.HasPrefix(source, "query.") {
			queryName := strings.TrimPrefix(source, "query.")

			// Optional limit from the field schema (max items for the list).
			itemLimit := 0
			if l, ok := fieldDef["limit"].(float64); ok {
				itemLimit = int(l)
			}

			req := queryresolve.QueryRequest{
				Name:   queryName,
				SiteID: resolver.siteID,
				Limit:  itemLimit,
			}
			value, qerr := queryresolve.Resolve(ctx, resolver.db, req, resolver.logger)
			if qerr != nil {
				resolver.logger.Warn("plan_sections: query resolution failed",
					zap.String("field", fieldName),
					zap.String("source", source),
					zap.Error(qerr))
				// Resolver ERRORED — distinct from an empty result and must NOT be
				// routed into on_missing, or a genuine failure would be masked as
				// "no data" (the trap bugs_open/054 flags).
				//
				// For a REQUIRED (or min_items-declaring) field, "leave the field
				// unresolved and proceed" shipped hollow listing sections at full
				// page weight (bugs_open/444: resolveBusinessDirectory's designed
				// loud failure, bugs_open/206, evaporated into this Warn line and
				// the section built with only its LLM headline). So: offer the
				// regeneration carry first (a rebuild during a transient error
				// keeps the page's own last-good value; storedFieldValue refuses
				// empties, so a never-filled field cannot be "carried"), else
				// DEFER the section — the loud HITL path — rather than build it
				// without the data it exists to show. on_missing is still never
				// consulted for errors; defer is the third state that keeps
				// error and no-data distinguishable.
				if required || fieldMinItems > 0 {
					if carryStored() {
						continue
					}
					// A declared fallback still beats a deferral — a build that
					// previously succeeded via fallback must keep succeeding
					// (council round 1, render_guardian). Measured 2026-09-02:
					// ZERO of the 17 required/min_items query fields fleet-wide
					// declare one, so this arm is future-proofing, not a live
					// behaviour change.
					if fallback != nil {
						resolvedData[fieldName] = fallback
						continue
					}
					shouldDefer = true
					missingFields = append(missingFields, missingField{
						Field:     fieldName,
						Source:    source,
						OnMissing: onMissing,
						Reason:    fmt.Sprintf("required query source errored: %v", qerr),
						Type:      fieldType,
						Items:     fieldItems,
						MinItems:  fieldMinItems,
					})
					continue
				}
				// Optional field: apply fallback if any, else leave unresolved —
				// behaviour unchanged, but the error is now a DURABLE structural
				// miss on the plan item (the bugs_open/238 visibility channel),
				// not one Warn line in a pod that restarts: an optional query
				// field that errors renders as silently absent, and "silently"
				// was the objection (council corr c0990eb3 round 2,
				// bug_historian HIGH — the recurring silent-empty shape).
				structuralMisses = append(structuralMisses, missingField{
					Field:     fieldName,
					Source:    source,
					OnMissing: onMissing,
					Reason:    fmt.Sprintf("optional query source errored: %v", qerr),
					Type:      fieldType,
					Items:     fieldItems,
					MinItems:  fieldMinItems,
				})
				if fallback != nil {
					resolvedData[fieldName] = fallback
				}
				continue
			}

			// Successful resolve. A query-sourced list that came back empty — or
			// shorter than its declared min_items — fails the field's required/
			// min_items contract, so honour on_missing exactly as the generic path
			// does, rather than storing an empty slice the template ranges over to
			// nothing (bugs_open/054 fix-candidate 2). Scalars and satisfied lists
			// store as before; a nil, non-list value keeps its prior fallback path.
			if queryListBelowContract(value, required, fieldMinItems) {
				// A required / min_items-declaring list resolved empty (or shorter
				// than its floor). Honour on_missing — but log it first, so a
				// required listing that never recovers data is an operator-visible
				// event, not a silent disappearance (bugs_open/054, council R1
				// bug_historian). Errored resolves took the branch above, not this.
				n, _ := queryResultLen(value)
				resolver.logger.Warn("plan_sections: query list below its required/min_items contract — applying on_missing",
					zap.String("field", fieldName),
					zap.String("source", source),
					zap.Int("resolved_items", n),
					zap.Int("min_items", fieldMinItems),
					zap.Bool("required", required),
					zap.String("on_missing", onMissing))
				handleMissingField()
				continue
			}
			if value != nil {
				resolvedData[fieldName] = value
				continue
			}

			// value == nil with no error and not a below-contract list — preserve
			// prior behaviour: apply fallback if any, else leave unresolved.
			if fallback != nil {
				resolvedData[fieldName] = fallback
			}
			continue
		}

		// Renderer/static fields — resolved at render time, not now. A
		// renderer/static source names ANOTHER writer (resolve_internal_links,
		// applyCTARecompute), so on a regeneration the current value exists
		// only in the page's stored content_data — which save_page_sections
		// replaces wholesale. Skipping here therefore destroyed the key on
		// every content_rewrite while every re-render (stored ⊕ fresh merge)
		// kept it: bugs_open/268, 214 CTA anchors gone fleet-wide. So carry
		// the stored value first. This does not reintroduce the migration
		// 091/098 revert problem — the carry re-supplies the page's own last
		// authored/resolved value (storedFieldValue refuses empties), it does
		// not re-derive one from a spec, so it cannot fabricate a destination
		// or overwrite an edit with a recomputation. The early continue
		// itself stays: renderer/static must not reach required-field
		// handling (098's section-readiness contract), and a declared
		// fallback still writes when nothing is stored (181/097b's pin) —
		// but a stored value now beats the fallback, because the fallback is
		// a default and the stored value is the page.
		if source == "renderer" || source == "static" ||
			strings.HasPrefix(source, "renderer.") ||
			strings.HasPrefix(source, "static.") {
			if !carryStored() && fallback != nil {
				resolvedData[fieldName] = fallback
			}
			continue
		}

		// Resolve data source
		value, found := resolver.resolve(ctx, source, section)

		if found && value != nil {
			resolvedData[fieldName] = value
			continue
		}

		// Data not found — apply the field's on_missing rule. Shared with the
		// query.* path above via handleMissingField so the two cannot drift
		// (the drift class bugs_closed/044 was closed for).
		handleMissingField()
	}

	// Set before the skip/defer branches, not after: a section that carried a
	// key and was then deferred for an unrelated field still carried it, and a
	// record that only survives the happy path is the kind of partial evidence
	// this bug was hard to see through in the first place (bugs_open/238).
	item.CarriedFields = carriedFields
	item.StructuralMisses = structuralMisses

	// Skip takes priority over defer
	if shouldSkip {
		item.Status = "skipped"
		if len(missingFields) > 0 {
			item.Reason = fmt.Sprintf("missing data: %s", missingFields[0].Reason)
		} else {
			item.Reason = "on_missing=skip_section triggered"
		}
		item.Missing = missingFields
		return item
	}

	if shouldDefer {
		item.Status = "deferred"
		item.Missing = missingFields
		if len(missingFields) > 0 {
			reasons := make([]string, len(missingFields))
			for i, m := range missingFields {
				if m.Reason != "" {
					reasons[i] = m.Reason
				} else {
					reasons[i] = fmt.Sprintf("%s (from %s)", m.Field, m.Source)
				}
			}
			item.Reason = strings.Join(reasons, "; ")
		}
		return item
	}

	// Authoritative hero aliasing: when this section declares an image-typed
	// field, also write the resolved page hero under the legacy alias keys
	// (hero_url, background_image) unless the schema declares them itself.
	// resolved_data is merged LAST at render time (RenderComponentAction's
	// merge_with overlay — "resolved data wins on conflicts, by design"), so
	// this is what lets the per-page hero defeat the site-wide hero_url that
	// BuildRenderContext still injects for legacy templates: without it,
	// {{or .hero_url .background_image}} picks the site-wide value and every
	// page shows the same image.
	if sectionHasImageField(fieldsRaw) {
		resolver.ensureAssets(ctx)
		if heroURL, ok := resolver.assets["hero"]; ok && heroURL != "" {
			for _, alias := range []string{"hero_url", "background_image"} {
				if _, declared := fieldsRaw[alias]; declared {
					continue // the field's own resolution governs
				}
				if _, already := resolvedData[alias]; !already {
					resolvedData[alias] = heroURL
				}
			}
		}
	}

	// Section is ready
	item.ResolvedData = resolvedData
	item.LLMFields = llmFields
	item.LLMFieldSpecs = llmFieldSpecs
	return item
}

// sectionHasImageField reports whether any declared field in a component's
// input_schema fields map is image-typed (type "image" or "image_url").
func sectionHasImageField(fieldsRaw map[string]interface{}) bool {
	for _, defRaw := range fieldsRaw {
		def, ok := defRaw.(map[string]interface{})
		if !ok {
			continue
		}
		if t, _ := def["type"].(string); t == "image" || t == "image_url" {
			return true
		}
	}
	return false
}

// filterSiteLevelSections removes section names that correspond to site-level
// components (headers, footers). These are managed by site_components and injected
// during page assembly — they should never enter the page_components pipeline.
func isSiteLevelSectionName(s string) bool {
	lower := strings.ToLower(s)
	return strings.Contains(lower, "header") ||
		strings.Contains(lower, "footer") ||
		lower == "site-header" ||
		lower == "site-footer" ||
		lower == "head" ||
		strings.HasPrefix(lower, "head-")
}

// composeScopedWriterBlock builds the verified-facts block for ONE section
// from its plan-time fact assignment (bugs_open/151 candidate 1): the site's
// CURRENT evidence_base filtered to the assigned IDs, rendered by the same
// composeWriterBlock that builds the site-wide block — so values are
// substituted at compose time, and the assignment pins WHICH facts a section
// states, never their numbers. An ID matching no current fact is logged and
// inert: the section simply does not state it. allowed_entities are carried
// unfiltered — scoping exists to stop FACT repetition; entity-relationship
// permission is site-level and unchanged.
func composeScopedWriterBlock(eb map[string]interface{}, assigned []string, logger *zap.Logger, sectionName string) string {
	if len(assigned) == 0 || eb == nil {
		return ""
	}
	factsRaw, _ := eb["facts"].([]interface{})
	byID := make(map[string]interface{}, len(factsRaw))
	for _, fr := range factsRaw {
		fact, ok := fr.(map[string]interface{})
		if !ok {
			continue
		}
		if id := datahelpers.GetStringField(fact, "id", ""); id != "" {
			byID[id] = fr
		}
	}
	subset := make([]interface{}, 0, len(assigned))
	for _, id := range assigned {
		if f, ok := byID[id]; ok {
			subset = append(subset, f)
		} else {
			logger.Warn("plan_sections: assigned fact id matches no current evidence_base fact — inert",
				zap.String("section", sectionName),
				zap.String("fact_id", id))
		}
	}
	if len(subset) == 0 {
		return ""
	}
	filtered := map[string]interface{}{"facts": subset}
	if ents, ok := eb["allowed_entities"]; ok {
		filtered["allowed_entities"] = ents
	}
	// The scoped block REPLACES the site-wide block in that section's prompt,
	// so the site's verbatim guidance must ride along or a fact-scoped section
	// silently loses the NEVER-STATE list (the loss that keeps sites unmanaged
	// — see composeWriterBlock's guidance note; bugs_open/387).
	if g, ok := eb["writer_block_guidance"]; ok {
		filtered["writer_block_guidance"] = g
	}
	return composeWriterBlock(filtered)
}

// ============================================================================
// Check for open data requests (prevents repeated LLM waste)
// ============================================================================

// loadOpenSectionDataRequests returns a map of section_name → reason for all
// sections on this page that have an open needs_section_data work item.
// "Open" means status not in a terminal state (complete, wont_fix, rejected, failed).
func loadOpenSectionDataRequests(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName string, logger *zap.Logger) map[string]string {
	result := make(map[string]string)

	rows, err := db.QueryContext(ctx, `
		SELECT spec->>'section_name', LEFT(summary, 120)
		FROM site_work_items
		WHERE site_id = $1
		  AND item_type = 'needs_section_data'
		  AND spec->>'page_name' = $2
		  AND status NOT IN ('complete', 'wont_fix', 'rejected', 'failed')
	`, siteID, pageName)
	if err != nil {
		logger.Warn("loadOpenSectionDataRequests: query failed", zap.Error(err))
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var sectionName, summary string
		if err := rows.Scan(&sectionName, &summary); err != nil {
			continue
		}
		if sectionName != "" {
			result[sectionName] = summary
		}
	}

	if len(result) > 0 {
		logger.Info("loadOpenSectionDataRequests: found open data requests",
			zap.Int("count", len(result)),
			zap.String("page", pageName))
	}

	return result
}

// closeResolvedDataRequest marks a needs_section_data item as complete when
// plan_sections determines the section is now ready (component created, data
// arrived, etc.). This closes the feedback loop — data requests don't block
// sections forever once the underlying issue is resolved.
func closeResolvedDataRequest(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName, sectionName string, logger *zap.Logger) {
	result, err := db.ExecContext(ctx, `
		UPDATE site_work_items
		SET status = 'complete',
		    completed_at = NOW(),
		    handled_by = 'plan_sections',
		    result = jsonb_build_object('auto_resolved', true, 'reason', 'section now ready — component or data available'),
		    updated_at = NOW()
		WHERE site_id = $1
		  AND item_type = 'needs_section_data'
		  AND spec->>'page_name' = $2
		  AND spec->>'section_name' = $3
		  AND status NOT IN ('complete', 'wont_fix', 'rejected', 'failed')
	`, siteID, pageName, sectionName)
	if err != nil {
		logger.Warn("closeResolvedDataRequest: update failed",
			zap.String("section", sectionName),
			zap.String("page", pageName),
			zap.Error(err))
		return
	}
	if rows, _ := result.RowsAffected(); rows > 0 {
		logger.Info("closeResolvedDataRequest: stale data request auto-closed",
			zap.String("section", sectionName),
			zap.String("page", pageName))
	}
}

// ============================================================================
// Persist skip decisions onto the page row (bugs_open/040 skip-not-recorded)
// ============================================================================

// persistSectionSkips durably records this build's on_missing=skip_section
// outcomes: pages.suppressed_sections := (current − readyNames) ∪ skippedNames.
// The column was verified unused fleet-wide before adoption (0 of 306 pages,
// 2026-07-24) and every completeness reader — the 040 partial-build guard in
// UpdatePageStatusAction, check_empty_sections, check_required_fields_missing —
// already excludes it, so recording here is what stops a legitimately
// data-gated section being counted as a build shortfall forever. Removing
// readyNames makes the record self-healing: when the missing data later
// arrives and the section plans ready, it is un-suppressed the same build.
// Deferred sections are NOT passed here (their needs_human_review item is
// their durable trace). Returns the persistence error rather than swallowing
// it: the caller stays warn-not-fail for the BUILD (a persistence failure
// must not break it) but escalates the failure to agent_error_log — a skip
// decision silently failing to persist would otherwise reproduce the exact
// vanishing-record defect this function exists to fix (council 164058e6,
// bug_historian objection). Values stay plain text names — readers use
// jsonb `?` containment.
func persistSectionSkips(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName string, readyNames, skippedNames []string, logger *zap.Logger) error {
	// nil-safety: a nil slice marshals to JSON null, which would make
	// jsonb_array_length($4::jsonb) fail and disarm the WHERE tail's no-op
	// guard. Always send [] for empty.
	if readyNames == nil {
		readyNames = []string{}
	}
	if skippedNames == nil {
		skippedNames = []string{}
	}
	readyJSON, err := json.Marshal(readyNames)
	if err != nil {
		return fmt.Errorf("marshal ready: %w", err)
	}
	skippedJSON, err := json.Marshal(skippedNames)
	if err != nil {
		return fmt.Errorf("marshal skipped: %w", err)
	}
	// The WHERE tail makes the call a no-op unless there is something to add
	// (skips this build) or possibly remove (an existing non-empty set), so the
	// common all-ready/nothing-suppressed build never rewrites the row.
	result, err := db.ExecContext(ctx, `
		UPDATE pages SET suppressed_sections = (
			SELECT COALESCE(jsonb_agg(v ORDER BY v), '[]'::jsonb) FROM (
				SELECT DISTINCT v FROM (
					SELECT jsonb_array_elements_text(COALESCE(suppressed_sections, '[]'::jsonb)) AS v
					EXCEPT
					SELECT jsonb_array_elements_text($3::jsonb)
					UNION
					SELECT jsonb_array_elements_text($4::jsonb)
				) u
			) s
		), updated_at = NOW()
		WHERE site_id = $1 AND name = $2
		  AND (jsonb_array_length(COALESCE(suppressed_sections, '[]'::jsonb)) > 0
		       OR jsonb_array_length($4::jsonb) > 0)
	`, siteID, pageName, string(readyJSON), string(skippedJSON))
	if err != nil {
		return fmt.Errorf("suppressed_sections merge update: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows > 0 {
		logger.Info("persistSectionSkips: skip decisions persisted to suppressed_sections",
			zap.String("page", pageName),
			zap.Strings("skipped", skippedNames),
			zap.Int("ready_unsuppressed_candidates", len(readyNames)))
	}
	return nil
}

// ============================================================================
// Create work items for deferred sections
// ============================================================================

func createDeferredItems(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName, parentWorkItemID string, deferred []sectionPlanItem, logger *zap.Logger) {
	var parentID *uuid.UUID
	if parentWorkItemID != "" {
		if parsed, err := uuid.Parse(parentWorkItemID); err == nil {
			parentID = &parsed
		}
	}

	for _, section := range deferred {
		// Build missing fields summary
		missingDescs := make([]string, len(section.Missing))
		for i, m := range section.Missing {
			if m.Reason != "" {
				missingDescs[i] = m.Reason
			} else {
				missingDescs[i] = fmt.Sprintf("field '%s' from %s", m.Field, m.Source)
			}
		}

		spec := map[string]interface{}{
			"page_name":    pageName,
			"section_name": section.Name,
			"component_id": section.ComponentID,
			"function":     section.Function,
			"missing":      section.Missing,
			"source":       "plan_sections",
		}
		specJSON, _ := json.Marshal(spec)

		summary := fmt.Sprintf("Section '%s' on %s needs: %s",
			section.Name, pageName, strings.Join(missingDescs, "; "))
		if len(summary) > 250 {
			summary = summary[:247] + "..."
		}

		itemKey := fmt.Sprintf("section_data_%s_%s_%s",
			pageName, sanitiseSectionKey(section.Name), siteID)

		_, err := db.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				spec, priority, status, created_by,
				item_key, parent_item_id
			) VALUES ($1, 'section-planner', 'build', 'needs_section_data', 'medium', $2,
					  $3::jsonb, 50, 'needs_human_review',
					  'plan_sections', $4, $5)
			ON CONFLICT DO NOTHING
		`, siteID, summary, string(specJSON), itemKey, parentID)

		if err != nil {
			logger.Warn("createDeferredItems: failed to insert",
				zap.String("section", section.Name), zap.Error(err))
		} else {
			logger.Info("createDeferredItems: HITL item created",
				zap.String("section", section.Name),
				zap.String("page", pageName))
		}
	}
}

func sanitiseSectionKey(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		if r == ' ' {
			return '_'
		}
		return -1
	}, s)
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}

// stringOrEmpty extracts a string from an interface{}, returning "" when the
// value isn't a string or is nil. Used when reading optional fields off the
// parsed input_schema map (llm_guidance, on_missing) where the schema author
// may simply omit the key.
// extractArrayItemFields returns the sorted field names each element of an
// array-typed input_schema field must contain. Supports both conventions in
// use: `items` (flat name->type map: faq, differentiators, services-grid) and
// `item_schema` (name->{type,...} map: info-card-grid). Returns nil for
// non-array fields or fields with no declared element shape. Sorted because Go
// map iteration is otherwise random and we want stable prompts and specs.
//
// A legacy JSON-Schema `items` ({"type":"object","required":[…],"properties":{…}})
// reaches here verbatim: datahelpers.SchemaContentFields copies `items` through
// unchanged when it projects the legacy dialect onto the v2 field shape. Read
// naively as a flat map its keys are the JSON-Schema KEYWORDS, and those keywords
// then travel into the writer's prompt as the field names to emit — which is how
// mechanism-flow shipped steps keyed properties/required/type and rendered empty
// (bugs_open/240). `properties` being a map is the discriminator: in the flat
// convention a value is a type NAME, never a map.
func extractArrayItemFields(fieldDef map[string]interface{}) []string {
	var fields []string
	if items, ok := fieldDef["items"].(map[string]interface{}); ok {
		if props, isJSONSchema := items["properties"].(map[string]interface{}); isJSONSchema {
			for k := range props {
				fields = append(fields, k)
			}
			sort.Strings(fields)
			return fields
		}
		for k := range items {
			fields = append(fields, k)
		}
	}
	if itemSchema, ok := fieldDef["item_schema"].(map[string]interface{}); ok {
		for k := range itemSchema {
			fields = append(fields, k)
		}
	}
	sort.Strings(fields)
	return fields
}

func stringOrEmpty(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// repeatSubjectGap names one component type that appears more than once in a
// page's FILTERED section list while at least one of its instances carries no
// subject — the bugs_open/443 shape: those instances receive identical briefs
// and the near-duplicate output is the predicted result, not a bad roll.
type repeatSubjectGap struct {
	Component      string
	Repeats        int
	WithoutSubject int
}

// repeatedComponentSubjectGaps is the pure half of the build-side 443 detector,
// separated so its negatives are provable without mock bookkeeping. Repeats are
// counted PAGE-WIDE, not adjacently — non-adjacent instances duplicate too
// (443's our-position-on-ai case). Deliberately not gated on the page carrying
// any subject at all: the planner-side twin's gate is retro-spam protection,
// while the pages this exists to surface can never carry one. Results are
// sorted by component so the emitted rows are deterministic.
func repeatedComponentSubjectGaps(sectionNames, sectionSubjects []string) []repeatSubjectGap {
	repeats := map[string]int{}
	subjectless := map[string]int{}
	for i, n := range sectionNames {
		repeats[n]++
		if i >= len(sectionSubjects) || strings.TrimSpace(sectionSubjects[i]) == "" {
			subjectless[n]++
		}
	}
	var gaps []repeatSubjectGap
	for name, n := range repeats {
		if n < 2 || subjectless[name] == 0 {
			continue
		}
		gaps = append(gaps, repeatSubjectGap{Component: name, Repeats: n, WithoutSubject: subjectless[name]})
	}
	sort.Slice(gaps, func(i, j int) bool { return gaps[i].Component < gaps[j].Component })
	return gaps
}
