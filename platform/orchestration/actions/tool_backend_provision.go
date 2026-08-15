// FILE: platform/orchestration/actions/tool_backend_provision.go
//
// The deploy-time half of the requires-backend gate (VMB-010; the webdesign.uk
// lane's PLAN_2026-08-11 §2/§4 and the council-approved "site chat intake"
// EXPERIENCE_PLAN's data contract).
//
// Migration 406 gates the SUGGESTION: tool-suggester only offers a
// requires-backend-tagged tool to a site whose deploy_config capabilities
// include 'backend'. Nothing gated the DEPLOY: an add_tool item from any other
// source (a human, an older suggestion, a replayed item) would ship a widget
// that POSTs to an /api/ endpoint the site does not have. This file closes
// that door at the one chokepoint every library-tool deploy passes through,
// and asks — honestly, as a work item a human can see — for the one thing the
// in-cluster pipeline genuinely cannot do: provision a backend instance on the
// box.
//
// WHAT IT DELIBERATELY DOES NOT DO:
//
//   - It does not provision anything. The backend for the one live
//     requires-backend tool (chat-input-box) is a systemd service on the
//     webdesign.uk box, off-cluster by design (the isolation posture in
//     PLAN_2026-08-11 §4 Option A). The chassis has no path to the box and
//     must not grow one for this; the item it raises is the handover.
//   - It does not carry the token. The facts relay authenticates with
//     SITE_FACTS_TOKEN (terraform-owned, in personae-platform-secrets); the
//     item names the secret so the operator knows which one, and never its
//     value — a work-item row is readable by every agent on the fleet.
//   - It does not restate the provisioning runbook. The spec points at the
//     owning lane's RUNBOOK and the box/chat-service source; a second copy of
//     the env contract here would drift from the binary exactly the way
//     compiled-in facts drifted from evidence_base.
//
// THE TWO REFUSALS both mirror an existing predicate byte-for-byte rather
// than approximating it — the pin-predicate-vs-pool-predicate divergence is a
// filed landmine class:
//
//   - capability: `COALESCE(deploy_config->'capabilities','[]'::jsonb) ?
//     'backend'` is 406's own expression. If the two ever disagree, the
//     suggester and the deployer are answering different questions.
//   - facts: the relay (internal/core-manager/handlers/sitefacts.go) serves a
//     404 unless the site has a current evidence_base row with a 'facts' key,
//     and the chat binary REFUSES TO START on zero facts (an empty facts
//     section licenses the model to improvise). A provision item minted for a
//     site with no facts would be unsatisfiable at birth — bugs_open/177's
//     exact class, 9 of 9 items dead in needs_human_review — so the deploy
//     refuses instead, and the error names the missing half.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// siteFactsRelayBase is where a box-hosted backend fetches its site's live
// facts. Advisory config handed to the operator in the provision item — the
// route itself is registered in internal/core-manager/api/server.go and that
// registration is the authority, not this constant.
const siteFactsRelayBase = "http://core-manager.ai-persona-system.svc.cluster.local:8088/api/v1/site-facts/"

func siteFactsRelayURL(domain string) string {
	return siteFactsRelayBase + strings.ToLower(strings.TrimSpace(domain))
}

// toolRequiresBackend reports whether a tool's semantic_tags carry
// "requires-backend". The input is the semantic_tags column as text (the
// deploy action already selects it that way); empty and SQL-null both mean no
// tags, matching COALESCE(semantic_tags,'[]') in 406's query.
//
// A malformed tags column must not silently disarm the gate — the unsafe side
// of this check is "widget deployed against no backend", so on a parse failure
// it falls back to the quoted-element containment the `?` operator would have
// matched, and the caller's log carries the raw value.
func toolRequiresBackend(semanticTagsJSON string) bool {
	raw := strings.TrimSpace(semanticTagsJSON)
	if raw == "" || raw == "null" {
		return false
	}
	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err == nil {
		for _, t := range tags {
			if t == "requires-backend" {
				return true
			}
		}
		return false
	}
	return strings.Contains(raw, `"requires-backend"`)
}

// backendEligibility is what the deploy-time gate needs to know about a site
// before a requires-backend tool may be forked into it.
type backendEligibility struct {
	// Capable: deploy_config capabilities include 'backend' — 406's predicate.
	Capable bool
	// FactsCount: entries in the current evidence_base facts array — what the
	// relay would serve, and what the backend refuses to start without.
	FactsCount int
	// ContactFact: the first fact whose id or kind mentions contact, as a
	// small JSON object {id, claim}, or "" when none exists. The backend fails
	// closed TO the contact details, so the operator needs to know whether the
	// register holds any — absence is not a refusal here (the operator may
	// hold an owner-confirmed value the register does not), but it must be
	// visible in the item rather than discovered at first boot.
	ContactFact string
}

func loadBackendEligibility(ctx context.Context, db *sql.DB, siteID uuid.UUID) (backendEligibility, error) {
	var e backendEligibility
	err := db.QueryRowContext(ctx, `
		SELECT
			COALESCE(s.deploy_config->'capabilities', '[]'::jsonb) ? 'backend',
			COALESCE(jsonb_array_length(eb.data->'facts'), 0),
			COALESCE((SELECT jsonb_build_object('id', f->>'id', 'claim', f->>'claim')::text
			          FROM jsonb_array_elements(eb.data->'facts') f
			          WHERE f->>'id' ILIKE '%contact%' OR f->>'kind' ILIKE '%contact%'
			          LIMIT 1), '')
		FROM sites s
		LEFT JOIN site_specs eb
		  ON eb.site_id = s.id
		 AND eb.aspect = 'evidence_base'
		 AND eb.is_current = true
		 AND eb.data ? 'facts'
		WHERE s.id = $1
	`, siteID).Scan(&e.Capable, &e.FactsCount, &e.ContactFact)
	if err == sql.ErrNoRows {
		return e, fmt.Errorf("site %s not found", siteID)
	}
	return e, err
}

// backendEligibilityRefusal is the gate's decision, separated from the action
// so the refusal can be proven by test rather than asserted by comment. A nil
// return means the deploy may proceed.
func backendEligibilityRefusal(e backendEligibility, toolFunction, domain string) error {
	if !e.Capable {
		return fmt.Errorf(
			"tool %s requires a backend and site %s's deploy_config capabilities do not include 'backend' — "+
				"the widget would POST to an /api/ endpoint the site does not have. "+
				"This mirrors tool-suggester's eligibility gate (migration 406); if the site should be "+
				"backend-capable, that is an owner-level deploy_config change, not a deploy-time override",
			toolFunction, domain)
	}
	if e.FactsCount == 0 {
		return fmt.Errorf(
			"tool %s requires a backend fed by the site-facts relay, and site %s has no current "+
				"evidence_base facts — the relay would serve 404 and the backend refuses to start on zero facts. "+
				"Attest the site's facts (the register trail pattern) before deploying this tool",
			toolFunction, domain)
	}
	return nil
}

// backendProvisionRequest describes one freshly deployed requires-backend tool
// whose backend half now needs provisioning on the box.
type backendProvisionRequest struct {
	siteID       uuid.UUID
	domain       string
	pageID       uuid.UUID
	pageURL      string
	toolFunction string
	displayName  string
	forkID       uuid.UUID
	eligibility  backendEligibility
}

// raiseBackendProvisionItem files the operator-facing provisioning request.
// Raised only on the fresh-deploy arm: an already-deployed tool raised its
// item when it deployed, and re-minting on every backfill re-run would ask the
// operator to provision a backend that is already running.
//
// status is needs_human_review with no handler_agent — the live idiom for
// items whose consumer is a person (195 cta_names_unknown_destination rows
// carry exactly that shape). When box provisioning grows an automated
// consumer, THAT change re-points handler_agent; this emitter does not guess
// at an agent that does not exist.
//
// recurrenceExpected is set because this is an ACTION REQUEST: a completed
// predecessor means a backend was provisioned, and a later legitimate re-ask
// (box rebuilt, fork re-created) must not be branded unresolved at birth by
// the two-strike rule — bugs_open/024's regression, and the third-strike
// branding is terminal and silent (LANDMINES: recurrenceExpected).
//
// Dispositions: "raised", "deduped_open_item", "insert_failed". Non-fatal
// like every follow-on in this action — but the caller MUST surface the
// disposition in its output map: a requires-backend tool whose provision
// request was never filed is a permanently dead widget that every status
// check reads as a successful deploy.
func raiseBackendProvisionItem(ctx context.Context, params ActionParams, logger *zap.Logger, req backendProvisionRequest) string {
	logger = logger.With(
		zap.String("tool_function", req.toolFunction),
		zap.String("domain", req.domain),
	)

	if params.DB == nil {
		logger.Warn("raiseBackendProvisionItem: no database connection, provision item not raised")
		return "insert_failed"
	}

	specFields := map[string]interface{}{
		"domain":                req.domain,
		"tool_function":         req.toolFunction,
		"tool_display_name":     req.displayName,
		"fork_id":               req.forkID.String(),
		"page_id":               req.pageID.String(),
		"page_url":              req.pageURL,
		"facts_relay_url":       siteFactsRelayURL(req.domain),
		"facts_token_secret":    "SITE_FACTS_TOKEN (personae-platform-secrets, terraform 047-base-configs) — reference the secret; never copy its value into a work item",
		"facts_count_at_deploy": req.eligibility.FactsCount,
		"runbook":               "docs/agent_docs/docs024_key_docs_latest/webdesign_uk_build_service/RUNBOOK_webdesign_uk_build_service.md (box provisioning; binary source box/chat-service/)",
		"source":                "tool-deployer",
		"suggestion": fmt.Sprintf(
			"The %s widget is live on %s%s and POSTs same-origin to its backend, which does not exist yet — "+
				"the widget fails closed to the site's contact details until it does. Provision a backend instance "+
				"on the box per the runbook: same binary, parameterised for this site (facts from the relay URL in "+
				"this spec, contact fallback from an owner-confirmed source, the four abuse controls as deployment "+
				"config). Re-read the site's live contact fact at provision time rather than trusting any copy in "+
				"this item.",
			req.displayName, req.domain, req.pageURL),
	}
	// The contact fact travels as a snapshot, named as one. It exists so the
	// operator can see at triage time whether the register holds a contact at
	// all; the suggestion above says to re-read it live before provisioning.
	if req.eligibility.ContactFact != "" {
		specFields["contact_fact_at_deploy"] = json.RawMessage(req.eligibility.ContactFact)
	} else {
		specFields["contact_fact_at_deploy"] = "none found in evidence_base — the operator must supply CONTACT_EMAIL/CONTACT_PHONE from an owner-confirmed source, or the backend will refuse to start"
	}
	spec, _ := json.Marshal(specFields)

	pageID := req.pageID
	inserted, err := withWorkItemTx(ctx, params.DB, logger, workItem{
		siteID:   req.siteID,
		source:   "tool-deployer",
		pipeline: "build",
		itemType: "backend_provision",
		severity: "high",
		summary: fmt.Sprintf("Provision the %s backend for %s — the deployed widget is fail-closed until this is done",
			req.displayName, req.domain),
		spec:               string(spec),
		pageID:             &pageID,
		priority:           70,
		handlerAgent:       "",
		status:             "needs_human_review",
		createdBy:          "tool-deployer",
		itemKey:            fmt.Sprintf("backend_provision:%s:%s", req.toolFunction, req.siteID),
		recurrenceExpected: true,
	})
	if err != nil {
		logger.Warn("raiseBackendProvisionItem: failed to create provision work item — the deployed widget has no backend and no request for one (non-fatal, but act on the disposition)",
			zap.Error(err))
		return "insert_failed"
	}
	if !inserted {
		logger.Info("raiseBackendProvisionItem: an open provision item already holds this key")
		return "deduped_open_item"
	}

	logger.Info("raiseBackendProvisionItem: backend provision item raised",
		zap.Int("facts_count", req.eligibility.FactsCount))
	return "raised"
}
