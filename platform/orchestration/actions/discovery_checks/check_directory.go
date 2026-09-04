// FILE: platform/orchestration/actions/discovery_checks/check_directory.go
//
// Discovery checks for the directory register pipeline — the AI model
// directory (model_directory_pipeline Phase D) and, on identical machinery,
// the company adoption tracker and protocol tracker (Phase E). Cloned
// originally from check_news_feed.go's MissingNewsSectionCheck /
// MissingNewsPageCheck — same opt-in shape
// (site_specs.classification.content_features.<key>), same handler
// (content-gap-planner decides placement, page-build-handler executes).
//
// ONE check pair, one profile per register kind (directoryCheckProfiles).
// Renamed from check_model_directory.go 2026-07-25, when the second kind
// arrived and the alternative was a third near-identical copy of a check
// pair. Two hand-maintained copies of the same logic that must stay
// identical is the drift class this codebase keeps paying for; the profile
// table makes "add a register kind" a data change.
//
// One structural difference from news, all checks account for: there is no
// per-site "missing_..._sources" equivalent to missing_news_sources. The
// register (directory_entities/directory_claims, migration 192) is global,
// not per-site — discovery (the directory-researcher agent, on its own
// scheduled cadence) is what populates it, not anything triggered by a site
// opting in. So the checks gate on "does the global register have ANY
// publishable data OF THIS KIND yet" rather than "does this site have its
// own sources": a site can opt in before the register has data, and the
// checks simply wait (return no finding) rather than creating an empty
// section.
//
// The kind gate is per-kind on purpose. Opting into the adoption tracker
// while only model claims exist must NOT create an adoption-tracker section
// that renders empty — an earlier draft counted all claims regardless of
// kind and would have done exactly that.
//
// Simplification vs. the news precedent (deliberate, not an oversight):
// the page check does NOT do news's "stranded nav page / retype, don't
// duplicate" handling (bugs_open/015). That exists because news pages had
// already been created under wrong page_types (section-index,
// noticias-index, ...) on MANY sites before the news-index gate existed.
// These are new page_types with no such history — there is nothing to retype
// yet. Add the same protection here if that ever changes.
//
// To enable a kind, add its two check names to the
// completeness-discovery-agent's "checks" config array (see
// 194_enable_model_directory_checks.sql for the jsonb_set append pattern).
// Registration here makes a check AVAILABLE; the config array is what makes
// it RUN.

package discovery_checks

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// directoryCheckProfile is everything that differs between register kinds
// at discovery time.
type directoryCheckProfile struct {
	Kind             string // directory_entities.kind
	SpecKey          string // site_specs …content_features.<SpecKey>
	Label            string // human phrase for summaries: "model directory"
	SnippetComponent string // content_components.function for the section
	ListingComponent string // content_components.function for the full page
	PageType         string // pages.page_type for the dedicated page
	PageName         string // pages.name for the dedicated page
	SectionItemType  string // work item type for the missing section
	PageItemType     string // work item type for the missing page
	PageTitle        string // what the created page is for, in one phrase
}

var directoryCheckProfiles = []directoryCheckProfile{
	{
		Kind:             "model",
		SpecKey:          "model_directory",
		Label:            "model directory",
		SnippetComponent: "model-directory",
		ListingComponent: "model-directory-listing",
		PageType:         "model-directory",
		PageName:         "model-directory",
		SectionItemType:  "missing_model_directory_section",
		PageItemType:     "missing_model_directory_page",
		PageTitle:        "a cited directory of AI models with their prices and capabilities",
	},
	{
		Kind:             "company",
		SpecKey:          "adoption_tracker",
		Label:            "AI agent adoption tracker",
		SnippetComponent: "adoption-tracker",
		ListingComponent: "adoption-tracker-listing",
		PageType:         "adoption-tracker",
		PageName:         "adoption-tracker",
		SectionItemType:  "missing_adoption_tracker_section",
		PageItemType:     "missing_adoption_tracker_page",
		PageTitle:        "a cited tracker of which organisations are deploying AI agents, how, and with what measured result",
	},
	{
		Kind:             "protocol",
		SpecKey:          "protocol_tracker",
		Label:            "agent protocol tracker",
		SnippetComponent: "protocol-tracker",
		ListingComponent: "protocol-tracker-listing",
		PageType:         "protocol-tracker",
		PageName:         "protocol-tracker",
		SectionItemType:  "missing_protocol_tracker_section",
		PageItemType:     "missing_protocol_tracker_page",
		PageTitle:        "a cited tracker of agent communication protocols and their uptake",
	},
	// Phase B finance/insurance kinds (2026-08-13). Same machinery, three more
	// profiles. Owner ruling: these directories carry NON-PRICE facts only
	// (regulator status, product types, underwriter, established year — never
	// APR/rates/premiums), which is why no PageTitle below mentions prices.
	{
		Kind:             "mortgage-lender",
		SpecKey:          "mortgage_lender_directory",
		Label:            "mortgage lender directory",
		SnippetComponent: "mortgage-lender-directory",
		ListingComponent: "mortgage-lender-directory-listing",
		PageType:         "mortgage-lenders",
		PageName:         "mortgage-lenders",
		SectionItemType:  "missing_mortgage_lender_directory_section",
		PageItemType:     "missing_mortgage_lender_directory_page",
		PageTitle:        "a cited directory of UK mortgage lenders: regulator status, product types, and lender history",
	},
	{
		Kind:             "savings-provider",
		SpecKey:          "savings_provider_directory",
		Label:            "savings provider directory",
		SnippetComponent: "savings-provider-directory",
		ListingComponent: "savings-provider-directory-listing",
		PageType:         "savings-providers",
		PageName:         "savings-providers",
		SectionItemType:  "missing_savings_provider_directory_section",
		PageItemType:     "missing_savings_provider_directory_page",
		PageTitle:        "a cited directory of UK savings providers: regulator status, protection scheme membership, and product types",
	},
	{
		Kind:             "health-insurer",
		SpecKey:          "health_insurer_directory",
		Label:            "health insurer directory",
		SnippetComponent: "health-insurer-directory",
		ListingComponent: "health-insurer-directory-listing",
		PageType:         "health-insurers",
		PageName:         "health-insurers",
		SectionItemType:  "missing_health_insurer_directory_section",
		PageItemType:     "missing_health_insurer_directory_page",
		PageTitle:        "a cited directory of UK health insurers: regulator status, underwriter, and cover types",
	},
	{
		Kind:             "copywriter",
		SpecKey:          "copywriter_directory",
		Label:            "copywriter directory",
		SnippetComponent: "copywriter-directory",
		ListingComponent: "copywriter-directory-listing",
		PageType:         "copywriter-directory",
		PageName:         "copywriter-directory",
		SectionItemType:  "missing_copywriter_directory_section",
		PageItemType:     "missing_copywriter_directory_page",
		PageTitle:        "a cited directory of UK copywriting organisations, with what each specialises in",
	},
}

func init() {
	for _, p := range directoryCheckProfiles {
		Register(&MissingDirectorySectionCheck{profile: p})
		Register(&MissingDirectoryPageCheck{profile: p})
	}
}

// ListingComponentSpecKeys exposes the per-kind component→SpecKey pairs from
// directoryCheckProfiles (both the listing and the snippet component names), so
// consumers that must stay in lockstep with the kind roster — the bugs_open/444
// plan-time gate — DERIVE it rather than hand-mirroring it. A kind added to the
// profiles above is automatically known to every caller of this function;
// before this existed the gate carried its own copy of these pairs, which is
// exactly the two-hand-maintained-rosters drift class the estate keeps paying
// for (council corr c0990eb3 round 2, architecture + reuse_agent).
func ListingComponentSpecKeys() map[string]string {
	out := make(map[string]string, len(directoryCheckProfiles)*2)
	for _, p := range directoryCheckProfiles {
		out[p.ListingComponent] = p.SpecKey
		out[p.SnippetComponent] = p.SpecKey
	}
	return out
}

// directoryHasPublishableData reports how many current, found claims exist
// for entities of this kind — the "is there anything to show yet" gate a
// per-site sources check would apply, re-keyed to a global register with no
// site_id to check.
//
// On the `de.status = 'active'` filter (council objection, debug_historian
// seat, 2026-07-25): the seat was right to demand the column be enumerated
// before a gate query filters on it — an entity status that LOOKS
// authoritative but whose semantics were never checked is a documented trap
// here (the sites.status informational-column case). Enumerated against live
// data the same day:
//
//	SELECT status, kind, count(*) FROM directory_entities GROUP BY 1,2;
//	active|company|16   active|model|27   active|protocol|4
//
// One value, 'active', across all three kinds — so today the filter is a
// no-op and changes nothing about what this gate counts. It is kept because
// migration 192 defines the column as active|archived|superseded, and an
// archived entity's claims must not make a section look publishable. Stated
// rather than assumed: this is a deliberate no-op-for-now, not a silent
// behaviour change.
func directoryHasPublishableData(dctx DiscoveryCheckContext, kind string) (int, error) {
	var claimCount int
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*)
		FROM directory_claims dc
		JOIN directory_entities de ON de.id = dc.entity_id
		WHERE dc.is_current AND dc.status = 'found'
		  AND de.kind = $1 AND de.status = 'active'
	`, kind).Scan(&claimCount)
	return claimCount, err
}

// directorySpecFlags reads content_features.<key> from a site's current
// classification spec. ok is false when the aspect, key, or site spec
// doesn't exist — treated as "not opted in", not an error.
func directorySpecFlags(dctx DiscoveryCheckContext, specKey string) (recommended, separatePage bool, raw map[string]interface{}, ok bool) {
	var specData []byte
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'classification' AND is_current = true
	`, dctx.SiteID).Scan(&specData)
	if err != nil {
		return false, false, nil, false
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(specData, &spec); err != nil {
		return false, false, nil, false
	}
	contentFeatures, _ := spec["content_features"].(map[string]interface{})
	if contentFeatures == nil {
		return false, false, nil, false
	}
	feature, _ := contentFeatures[specKey].(map[string]interface{})
	if feature == nil {
		return false, false, nil, false
	}
	recommended, _ = feature["recommended"].(bool)
	separatePage, _ = feature["separate_page"].(bool)
	return recommended, separatePage, feature, true
}

// ===========================================================================
// MissingDirectorySectionCheck
// ===========================================================================
// Detects: site spec has content_features.<key>.recommended = true, the
// global register has publishable data of this kind, but no page carries
// this kind's section component.

type MissingDirectorySectionCheck struct{ profile directoryCheckProfile }

func (c *MissingDirectorySectionCheck) Name() string { return c.profile.SectionItemType }

func (c *MissingDirectorySectionCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	p := c.profile

	recommended, _, config, ok := directorySpecFlags(dctx, p.SpecKey)
	if !ok || !recommended {
		return &CheckResult{}, nil
	}

	claimCount, err := directoryHasPublishableData(dctx, p.Kind)
	if err != nil {
		return nil, fmt.Errorf("%s: register query failed: %w", p.SectionItemType, err)
	}
	if claimCount == 0 {
		// Site has opted in, but discovery hasn't produced anything of this
		// kind yet. Wait for the next sweep rather than creating an empty
		// section.
		return &CheckResult{}, nil
	}

	var sectionCount int
	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*) FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND EXISTS (
		      SELECT 1 FROM content_components cc
		      WHERE cc.id = pc.component_id AND cc.function = $2
		  )
	`, dctx.SiteID, p.SnippetComponent).Scan(&sectionCount)
	if err != nil {
		return nil, fmt.Errorf("%s: section query failed: %w", p.SectionItemType, err)
	}
	if sectionCount > 0 {
		return &CheckResult{}, nil
	}

	dctx.Logger.Info("MissingDirectorySectionCheck: register has data but no section",
		zap.String("site_id", dctx.SiteID.String()),
		zap.String("kind", p.Kind),
		zap.Int("claim_count", claimCount))

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":                p.SectionItemType,
		"registry_claim_count": claimCount,
		"directory_kind":       p.Kind,
		"page_name":            "index",
		"feature_config":       config,
		"add_sections":         []string{p.SnippetComponent},
		"description": fmt.Sprintf(
			"Site opted into the %s and the register has publishable entries, but the homepage has no %s section to display them.",
			p.Label, p.SnippetComponent),
		"suggestion": fmt.Sprintf("Add a %s section to the homepage to show the cited %s", p.SnippetComponent, p.Label),
		"category":   "content_completeness",
	})

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":       p.SectionItemType,
			"message":     fmt.Sprintf("Register has publishable %s entries but homepage has no %s section", p.Kind, p.SnippetComponent),
			"claim_count": claimCount,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     p.SectionItemType,
			Severity:     "medium",
			Summary:      fmt.Sprintf("Site opted into the %s (%d cited entries available) but has no homepage section for it", p.Label, claimCount),
			SpecJSON:     string(specJSON),
			Priority:     65,
			HandlerAgent: "content-gap-planner",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("%s:%s", p.SectionItemType, dctx.SiteID),
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

// ===========================================================================
// MissingDirectoryPageCheck
// ===========================================================================
// Detects: content_features.<key>.separate_page = true, the global register
// has publishable data of this kind, but no page of this kind's page_type
// exists.

type MissingDirectoryPageCheck struct{ profile directoryCheckProfile }

func (c *MissingDirectoryPageCheck) Name() string { return c.profile.PageItemType }

func (c *MissingDirectoryPageCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	p := c.profile

	recommended, separatePage, config, ok := directorySpecFlags(dctx, p.SpecKey)
	if !ok || !recommended || !separatePage {
		return &CheckResult{}, nil
	}

	claimCount, err := directoryHasPublishableData(dctx, p.Kind)
	if err != nil {
		return nil, fmt.Errorf("%s: register query failed: %w", p.PageItemType, err)
	}
	if claimCount == 0 {
		return &CheckResult{}, nil
	}

	var pageCount int
	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*) FROM pages
		WHERE site_id = $1 AND page_type = $2
	`, dctx.SiteID, p.PageType).Scan(&pageCount)
	if err != nil {
		return nil, fmt.Errorf("%s: page query failed: %w", p.PageItemType, err)
	}
	if pageCount > 0 {
		return &CheckResult{}, nil
	}

	var domain string
	_ = dctx.DB.QueryRowContext(dctx.Ctx, `SELECT domain FROM sites WHERE id = $1`, dctx.SiteID).Scan(&domain)

	dctx.Logger.Info("MissingDirectoryPageCheck: spec recommends a dedicated page but none exists",
		zap.String("site_id", dctx.SiteID.String()),
		zap.String("kind", p.Kind),
		zap.Int("claim_count", claimCount))

	gapSpec := map[string]interface{}{
		"check":                p.PageItemType,
		"page_name":            p.PageName,
		"page_type":            p.PageType,
		"category":             "content_completeness",
		"registry_claim_count": claimCount,
		"directory_kind":       p.Kind,
		"feature_config":       config,
		"approach":             "new_page",
		"description": fmt.Sprintf(
			"Site %s opted into a dedicated %s page (separate_page=true) and the register has %d cited entries, but no page of page_type '%s' exists.",
			domain, p.PageType, claimCount, p.PageType),
		"suggestion": fmt.Sprintf(
			"Create a new page named '%s' with page_type '%s' — %s — with a hero section, the %s component, and a call-to-action. Add it to header and footer navigation.",
			p.PageName, p.PageType, p.PageTitle, p.ListingComponent),
	}
	specJSON, _ := json.Marshal(gapSpec)

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":   p.PageItemType,
			"message": fmt.Sprintf("Site spec recommends a dedicated %s page but none exists for %s", p.PageType, domain),
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     p.PageItemType,
			Severity:     "medium",
			Summary:      fmt.Sprintf("Site %s needs a dedicated %s page (%d cited entries available)", domain, p.PageType, claimCount),
			SpecJSON:     string(specJSON),
			Priority:     60,
			HandlerAgent: "content-gap-planner",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("%s:%s", p.PageItemType, dctx.SiteID),
			BatchID:      dctx.BatchID,
		}},
	}, nil
}
