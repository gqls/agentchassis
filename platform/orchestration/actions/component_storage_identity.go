// FILE: platform/orchestration/actions/component_storage_identity.go
//
// resolveStorageIdentity decides WHERE store_generated_component writes:
// regeneration of an existing base row, or creation of a new one — and, when
// the existing row is depended on by sites other than the requester, DIVERTS
// the write to a site-scoped identity instead of colliding (bugs_open/311).
//
// The defect this closes: selection and storage key on different columns.
// A row the planner could not reuse (dropped by the section template guard,
// or carrying no section_type) could still be FOUND by the store's function
// lookup — so the store treated another site's component as "the row to
// regenerate", and the field-contract guard then (correctly) refused to
// strand that site's content_data. Three attempts, three identical refusals,
// and the requesting site shipped without its section: remortgagecalculator.uk
// with no calculator; loanzy.uk with 7 of 7 tool sections lost. No retry could
// ever succeed — the incumbent could be neither reused nor replaced.
//
// The rule: a base component that OTHER sites depend on is never a
// regeneration target for a build acting for a different site. The store
// mints `<function>-<domain-slug>` instead (deploy_tool_action's fork naming
// convention) as a fresh BASE row. forked_from stays NULL deliberately: the
// template is a fresh generation, not the incumbent's lineage, and every
// selection path filters `forked_from IS NULL` — the new row must be
// selector-visible so the requesting page's rebuild can link it and later
// sites planning the same section_type can reuse it. section_type carries the
// REQUESTED section name (unsuffixed, request vocabulary); function carries
// the storage identity. content_components has no site ownership column;
// ownership is only visible through page_components → pages → sites (and
// site_components), which is exactly the census taken here.

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// storageIdentity is resolveStorageIdentity's answer: the final storage key,
// the existing row (when regenerating), and the diversion record.
type storageIdentity struct {
	FunctionName string // final storage key — site-suffixed when diverted

	ExistingID     string
	ExistingHTML   string
	ExistingSchema string
	ExistingJS     sql.NullString
	IsRegeneration bool

	Diverted         bool     // a foreign collision was re-routed to a site-scoped identity
	DivertedFromID   string   // the incumbent row the write was steered away from
	DivertedFromFunc string   // the function name the generation originally chose
	ForeignDomains   []string // domains depending on the incumbent (for the durable record)
	DivertBlocked    string   // non-empty: foreign collision seen but no scoped identity mintable (requester domain unknown); legacy behaviour applies
}

func resolveStorageIdentity(
	ctx context.Context,
	db *sql.DB,
	functionName string,
	requesterSiteID string,
	requesterDomain string,
	logger *zap.Logger,
) (storageIdentity, error) {

	ident := storageIdentity{FunctionName: functionName}

	found, err := lookupBaseComponent(ctx, db, functionName, &ident)
	if err != nil {
		return ident, err
	}
	if !found {
		return ident, nil // plain creation
	}

	// Requester unknown (direct programmatic invocation, no work item):
	// legacy behaviour — regenerate whatever the function lookup found. The
	// diversion needs a requester to compare the dependent census against.
	if requesterSiteID == "" {
		ident.IsRegeneration = true
		return ident, nil
	}

	foreignIDs, foreignDoms, err := foreignDependents(ctx, db, ident.ExistingID, requesterSiteID)
	if err != nil {
		// Fail towards legacy behaviour, not towards a diversion we cannot
		// ground: an unreadable census must not silently re-route the write.
		// The field-contract guard remains the backstop against stranding.
		logger.Warn("resolveStorageIdentity: dependent census unreadable — proceeding as legacy regeneration",
			zap.String("function", functionName),
			zap.String("existing_id", ident.ExistingID),
			zap.Error(err))
		ident.IsRegeneration = true
		return ident, nil
	}
	if len(foreignIDs) == 0 {
		// No dependents, or only the requester's own pages: a normal
		// same-site regeneration.
		ident.IsRegeneration = true
		return ident, nil
	}

	// Foreign collision. Mint the site-scoped identity.
	domain := requesterDomain
	if domain == "" {
		// Best-effort: the work-item payload normally carries input_data.domain,
		// but a caller that only passed site_id can still be scoped.
		_ = db.QueryRowContext(ctx,
			`SELECT domain FROM sites WHERE id = $1::uuid`, requesterSiteID).Scan(&domain)
	}
	if domain == "" {
		ident.IsRegeneration = true
		ident.DivertBlocked = fmt.Sprintf(
			"existing component %s (function %q) is depended on by other sites (%s) but the requesting site's domain is unknown — proceeding as a legacy regeneration; the field-contract guard is the backstop",
			ident.ExistingID, functionName, strings.Join(foreignDoms, ", "))
		return ident, nil
	}

	suffixed := functionName + "-" + domainSlug(domain)
	div := storageIdentity{
		FunctionName:     suffixed,
		Diverted:         true,
		DivertedFromID:   ident.ExistingID,
		DivertedFromFunc: functionName,
		ForeignDomains:   foreignDoms,
	}
	foundScoped, err := lookupBaseComponent(ctx, db, suffixed, &div)
	if err != nil {
		return div, err
	}
	if !foundScoped {
		return div, nil // creation under the scoped name
	}

	// The scoped name exists too — normally the requester's own earlier
	// diverted row: regenerate it. If it is ALSO foreign-depended (cross-site
	// name squatting, the pathological tail), refuse loudly rather than loop
	// through ever-longer suffixes.
	foreign2, foreignDoms2, err := foreignDependents(ctx, db, div.ExistingID, requesterSiteID)
	if err != nil || len(foreign2) == 0 {
		div.IsRegeneration = true
		return div, nil
	}
	return div, fmt.Errorf(
		"cannot store component %q: the base name is depended on by other sites (%s) and the site-scoped name %q is too (%s) — refusing to overwrite either; repair or rename the existing rows",
		functionName, strings.Join(foreignDoms, ", "), suffixed, strings.Join(foreignDoms2, ", "))
}

// lookupBaseComponent finds the base (forked_from IS NULL) row for a function
// name and fills the identity's Existing* fields. Returns found=false on no
// row. The query and its ordering are the store action's historical lookup,
// moved here verbatim.
//
// Note on the missing is_active filter (changed 2026-05-06): with
// `AND is_active = true` a regeneration of a deactivated row fell through to
// the creation branch and hit the unique-on-name constraint when the old row
// had name == function. Ordering by `is_active DESC, updated_at DESC`
// preserves the old behaviour when active rows exist (the active row sorts
// first) and fixes regeneration when only an inactive row exists. The UPDATE
// branch reactivates deliberately — a regenerated template that passes every
// pre-store gate is by definition healthy, and the common reason for
// inactivity is "broken template, awaiting regeneration" (migration 036).
func lookupBaseComponent(ctx context.Context, db *sql.DB, functionName string, ident *storageIdentity) (bool, error) {
	err := db.QueryRowContext(ctx, `
		SELECT id::text,
		       COALESCE(html_template, ''),
		       COALESCE(input_schema::text, '{}'),
		       js_content
		FROM content_components
		WHERE function = $1 AND forked_from IS NULL
		ORDER BY is_active DESC, updated_at DESC
		LIMIT 1
	`, functionName).Scan(&ident.ExistingID, &ident.ExistingHTML, &ident.ExistingSchema, &ident.ExistingJS)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check existing component: %w", err)
	}
	return true, nil
}

// foreignDependents returns the sites OTHER than the requester whose pages or
// site chrome reference the component. This is the ownership view the schema
// itself does not carry.
func foreignDependents(ctx context.Context, db *sql.DB, componentID, requesterSiteID string) (siteIDs []string, domains []string, err error) {
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT s.id::text, s.domain
		FROM (
			SELECT p.site_id
			FROM page_components pc
			JOIN pages p ON p.id = pc.page_id
			WHERE pc.component_id = $1::uuid
			UNION
			SELECT sc.site_id
			FROM site_components sc
			WHERE sc.component_id = $1::uuid
		) dep
		JOIN sites s ON s.id = dep.site_id
		WHERE s.id::text <> $2
	`, componentID, requesterSiteID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, domain string
		if scanErr := rows.Scan(&id, &domain); scanErr != nil {
			return nil, nil, scanErr
		}
		siteIDs = append(siteIDs, id)
		domains = append(domains, domain)
	}
	return siteIDs, domains, rows.Err()
}
