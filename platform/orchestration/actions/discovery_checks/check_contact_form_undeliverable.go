// FILE: platform/orchestration/actions/discovery_checks/check_contact_form_undeliverable.go
//
// Detects contact forms that render fine and deliver nothing: a <form> whose
// action submits to the current URL ("#contact", "#", "", absent) or to the
// /contact endpoint that has never existed in the chassis. On a static host
// (B2 / VM nginx) that POST is a 405/404 — the visitor fills the form, presses
// send, and the message is gone with no error they can act on.
//
// Filed as bugs_open/006 §B. The live audit on 2026-07-20 found 10 of 11 sites
// in this state, which is why this is a check and not a one-off repair: the
// value comes from per-component content_data written by the content LLM, so
// nothing stopped it recurring on the next render.
//
// Relationship to the render-path fix (component_library.go, same bug):
// sanitiseFormAction repairs this automatically wherever the site has a real
// contact address. It deliberately does NOT synthesise one, because a mailto
// nobody reads makes the form look repaired while still losing the message.
// So this check is the backstop for the case the renderer will not guess at:
// a site with a form and no address to send it to.
//
// Expected volume changes across the image roll, and that is not a bug in the
// check: it reads DEPLOYED html, so until pages re-render it reports all 10
// affected sites. Once the render fix has shipped AND pages have re-rendered,
// the 6 sites with a real address self-repair and this should settle to the 4
// that have none (robot-hands, relojistas, vetcomparison, vonc on 2026-07-20).
// A count that stays at 10 after a full re-render means the render fix is not
// reaching this component — which is the useful signal, so do not "fix" it by
// narrowing this further.
//
// Deliberate boundaries:
//   - **contact-form components only** (data-component="contact-form"). The
//     first draft matched any <form> and pulled in 6 tool-calculator pages.
//     Tool forms bind submit from an EXTERNAL js bundle, so nothing in
//     rendered_html can clear them — verified 2026-07-20: not one form on the
//     fleet carries an inline addEventListener('submit')/onsubmit/
//     preventDefault, including the tools that demonstrably work. Judging them
//     here would file work items against working tools; broken tool js has its
//     own owners (bugs_open/041, /046). One owner per class, as dead_controls.
//   - mailto: and real POST handlers (idea.uk's /request) are destinations,
//     not defects.
//   - page_components only, matching dead_controls: chrome has its own fixers.
//
// Routing (two branches since 2026-07-24, PLAN_2026-07-24_contact_form_hardening):
//   - site WITH a resolvable sites.email → AUTO-REMEDIATE: emit a page_rerender
//     (handler page-rerender, reason section_data_resolved). The light section
//     re-render re-runs sanitiseFormAction, which deterministically converts the
//     dead form to a mailto to that address — no guess involved, so the
//     objection below does not apply to this branch.
//   - site WITHOUT one → needs_human_review with NO handler, unchanged. The fix
//     is a business decision — supply a contact address, point at a backend, or
//     drop the form — and picking one automatically would guess. Compare
//     dead_controls, same reasoning.
//
// Registration: automatic via init(). Enable by adding
// "contact_form_undeliverable" to a discovery agent's checks array AFTER the
// image carrying this file is live.

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/livespec"
	"go.uber.org/zap"
)

func init() { Register(&ContactFormUndeliverableCheck{}) }

type ContactFormUndeliverableCheck struct{}

func (c *ContactFormUndeliverableCheck) Name() string { return "contact_form_undeliverable" }

type undeliverableFormFinding struct {
	PageID   string `json:"page_id"`
	PageName string `json:"page_name"`
	SlotName string `json:"slot_name"`
	Position int    `json:"position"`
	Action   string `json:"action"`
	Reason   string `json:"reason"`
}

func (c *ContactFormUndeliverableCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.id::text, p.name, COALESCE(pc.slot_name, ''), pc.position,
		       COALESCE(substring(pc.rendered_html from 'action="([^"]*)"'), '') AS action_target
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.rendered_html IS NOT NULL
		  AND pc.locked_at IS NULL
		  -- Scoped to the contact-form component, NOT any <form>. See the
		  -- boundaries note above: tool forms bind submit from an external
		  -- bundle and cannot be cleared from rendered_html, so matching all
		  -- forms files work items against working tools.
		  AND pc.rendered_html ~* 'data-component="contact-form"'
		  AND (
		        pc.rendered_html !~* 'action="[^"]+"'          -- no action at all
		     OR pc.rendered_html ~* 'action="\s*"'             -- empty
		     OR pc.rendered_html ~* 'action="#[^"]*"'          -- self-anchor
		     OR pc.rendered_html ~* 'action="/contact"'        -- never existed
		  )
		ORDER BY p.name, pc.position
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("contact_form_undeliverable query failed: %w", err)
	}
	defer rows.Close()

	var findings []undeliverableFormFinding
	for rows.Next() {
		var f undeliverableFormFinding
		if err := rows.Scan(&f.PageID, &f.PageName, &f.SlotName, &f.Position, &f.Action); err != nil {
			dctx.Logger.Warn("Failed to scan undeliverable form", zap.Error(err))
			continue
		}
		f.Reason = classifyUndeliverableAction(f.Action)
		findings = append(findings, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("contact_form_undeliverable scan failed: %w", err)
	}
	if len(findings) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check":    "contact_form_undeliverable",
			"count":    len(findings),
			"findings": findings,
		}},
	}

	// One work item per page — a page with three dead forms is one decision.
	pageFindings := map[string][]undeliverableFormFinding{}
	for _, f := range findings {
		pageFindings[f.PageID] = append(pageFindings[f.PageID], f)
	}

	// Resolve the site's canonical contact address once. The render seam builds
	// the mailto from the sites.email COLUMN (loadSiteDataFull, and now
	// buildRerenderBaseData too), so a site WITH a resolvable address self-repairs
	// on a re-render — we can auto-remediate it rather than parking it for a
	// human. A site with no resolvable address still parks: inventing one is the
	// guess this file's header warns against, and that objection applies ONLY to
	// the address-less branch. See bugs_open/006 §B / PLAN_2026-07-24_contact_form_hardening.
	siteEmail, siteDomain := resolveSiteContact(dctx)
	resolvable := contactAddressResolvable(siteEmail, siteDomain)

	for pageID, pf := range pageFindings {
		var pageIDPtr *uuid.UUID
		if parsed, err := uuid.Parse(pageID); err == nil {
			pageIDPtr = &parsed
		}

		if resolvable {
			// AUTO-REMEDIATE. rerender_page_sections re-runs sanitiseFormAction,
			// which converts the dead form to a mailto to sites.email. It is a
			// light, in-place, no-LLM re-render, so it does NOT hit the deployed-
			// page rebuild bounce that a full needs_page rebuild risks. reason
			// section_data_resolved routes to the true template re-render via
			// check_rerender_mode WITHOUT the cta_links_stale CTA recompute, and
			// is the conventional "re-render this page's sections to reflect
			// resolved data" reason (render_news_section_html, reconcile_section_data).
			spec := map[string]interface{}{
				"check":     "contact_form_undeliverable",
				"page_id":   pageID,
				"page_name": pf[0].PageName,
				"findings":  pf,
				"fix": fmt.Sprintf("A section re-render converts the dead form to mailto:%s "+
					"(built by sanitiseFormAction from the sites.email column).", siteEmail),
			}
			// The reason described in the comment above, stamped with its routing
			// half by the one helper that defines the pair (bugs_open/440 phase 2).
			//
			// ⚠ THIS CHECK FILES TWO ITEM TYPES — a page_rerender item here and a
			// contact_form_undeliverable review item below — and they build
			// SEPARATE specs (the review item's is the inline map further down).
			// Verified live before stamping, because the routing key belongs only
			// on the rerender item: `SELECT item_type, spec->>'check', count(*)`
			// returned page_rerender 1 AND contact_form_undeliverable 7 for this
			// check name (2026-09-03). If the two specs are ever merged, the
			// stamp must move to a copy used by the rerender item alone.
			livespec.StampRerenderReason(spec, livespec.ReasonSectionDataResolved)
			specJSON, _ := json.Marshal(spec)
			result.WorkItems = append(result.WorkItems, WorkItemSpec{
				SiteID:   dctx.SiteID,
				PageID:   pageIDPtr,
				Source:   "discovery",
				Pipeline: "build",
				ItemType: "page_rerender",
				Severity: "high",
				Summary: fmt.Sprintf("Contact form on page %s submits nowhere — a re-render converts it to mailto:%s",
					pf[0].PageName, siteEmail),
				SpecJSON:     string(specJSON),
				Priority:     30,
				HandlerAgent: "page-rerender",
				Status:       "detected",
				CreatedBy:    dctx.AgentType,
				// Distinct from the human-review key below, so a page that gains an
				// address later is not blocked by a stale needs_human_review row.
				ItemKey: fmt.Sprintf("contact_form_undeliverable_rerender:%s", pageID),
				BatchID: dctx.BatchID,
			})
			continue
		}

		// ADDRESS-LESS: no honest address to convert to, so a re-render cannot
		// fix it. Supplying an address, pointing at a live backend, or dropping
		// the form is a business decision — park for a human (unchanged).
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":     "contact_form_undeliverable",
			"page_id":   pageID,
			"page_name": pf[0].PageName,
			"findings":  pf,
			"remedy": "Set the site's contact address in the sites.email column " +
				"(that is the field the render path reads — loadSiteDataFull's " +
				"COALESCE(si.email,''); content_data.email and the identity " +
				"site_spec are NOT read on render) and re-render — the render path " +
				"then converts the form to a mailto automatically. A bare " +
				"info@<domain> is treated as no address (the synthesised display " +
				"fallback, not a real inbox). Otherwise point the form at a live " +
				"POST handler, or remove it.",
		})
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			PageID:   pageIDPtr,
			Source:   "discovery",
			Pipeline: "content",
			ItemType: "contact_form_undeliverable",
			Severity: "high",
			Summary: fmt.Sprintf("Contact form on page %s submits nowhere (%d form(s); message is silently lost)",
				pf[0].PageName, len(pf)),
			SpecJSON:     string(specJSON),
			Priority:     30,
			HandlerAgent: "",
			Status:       "needs_human_review",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("contact_form_undeliverable:%s", pageID),
			BatchID:      dctx.BatchID,
		})
	}

	return result, nil
}

// resolveSiteContact returns the site's canonical contact address (the
// sites.email COLUMN — the field the render seam reads) and its domain.
func resolveSiteContact(dctx DiscoveryCheckContext) (email, domain string) {
	_ = dctx.DB.QueryRowContext(dctx.Ctx,
		`SELECT COALESCE(email, ''), COALESCE(domain, '') FROM sites WHERE id = $1`,
		dctx.SiteID,
	).Scan(&email, &domain)
	return email, domain
}

// contactAddressResolvable mirrors deliverableFormAction's guard
// (component_library.go): an address the render seam will actually convert into
// a mailto — non-empty, contains '@', and not the synthesised info@<own-domain>
// fallback some render paths set for a site with no configured address. Kept in
// lockstep with that guard; if one changes, change both.
func contactAddressResolvable(email, domain string) bool {
	email = strings.TrimSpace(email)
	if email == "" || !strings.Contains(email, "@") {
		return false
	}
	if strings.EqualFold(email, "info@"+strings.TrimSpace(domain)) {
		return false
	}
	return true
}

// classifyUndeliverableAction names WHY the action delivers nothing, so the
// review queue shows the distinction rather than a bare "broken form".
func classifyUndeliverableAction(action string) string {
	switch {
	case action == "":
		return "no_action_posts_to_self"
	case action == "/contact":
		return "nonexistent_contact_endpoint"
	case len(action) > 0 && action[0] == '#':
		return "self_anchor_posts_to_self"
	default:
		return "undeliverable"
	}
}
