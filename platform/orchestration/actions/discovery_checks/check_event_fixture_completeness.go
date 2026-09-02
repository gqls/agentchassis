// FILE: platform/orchestration/actions/discovery_checks/check_event_fixture_completeness.go
//
// bugs_open/427, council REVISE round 2 (08f56b7e), bug_historian HIGH +
// MEDIUM: `queryresolve.resolveUpcomingEvents` deliberately "omits any field
// a fact did not carry rather than inventing a placeholder" — the right
// policy for the RENDER path (never fabricate a venue nobody confirmed), but
// it leaves an incomplete or unevidenced fixture with NO durable signal
// anywhere: a Warn log line is not a record (WRONG_CALLS, "a print statement
// is not a config row"), and this is exactly the recurring shape 016b §9
// names — a field renders blank with no error, no work item, nothing to
// notice it by.
//
// This check is the visibility half, kept OUT of the resolver deliberately:
// a query.* resolver is called at render/plan time and must stay a pure
// read with no side effects (every existing resolver in this package is),
// so completeness auditing runs here instead, on the same discovery-check
// cadence as every other site-wide sweep.
//
// It reads a site's evidence_base register directly (the same raw-map read
// resolveUpcomingEvents uses, RFC_025 §9 Q2 — event_date/venue/participants/
// broadcaster are untyped extra keys, not on the typed EvidenceFact struct)
// rather than importing queryresolve's parsing: queryresolve.go itself
// imports this package (for PageRerenderItemKey), so the reverse import
// would cycle. Deliberately duplicated, same as missingProseLinksOpenItemSQL
// duplicates a status list from package actions for the identical reason —
// keep the two definitions of "does this fact have event_date" in lockstep
// by hand, there being no third package both can import from cheaply.
//
// Files ONE `capability_gap` item per site (the established "found something
// worth a human's attention, no dispatch, no handler" shape — bugs_closed/077,
// reconcile_site_plan_action.go's own capability_gap arm) naming which facts
// are incomplete or unevidenced. Retracts it the moment a run finds none —
// this check is the reason the estate CAN find nothing here, not a claim
// that nothing was ever wrong.
//
// Registration: automatic via init() → Register(&EventFixtureCompletenessCheck{}).
// Inert until named in a live `run_checks.config.checks` array.

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func init() { Register(&EventFixtureCompletenessCheck{}) }

type EventFixtureCompletenessCheck struct{}

func (c *EventFixtureCompletenessCheck) Name() string { return "event_fixture_completeness" }

// eventFixtureVerdict is the pure classification, so the canary and the unit
// tests grade the same code the site query does. Four outcomes:
//
//	""            — not an event fact at all (no event_date key)
//	"unevidenced" — has event_date but no citation url+quote (cannot be
//	                trusted; VerifyAndRegisterCitationsAction should never
//	                produce this shape, so seeing one means a hand edit or a
//	                future writer bypassed it)
//	"incomplete"  — evidenced, but missing venue, broadcaster, or has fewer
//	                than 2 participants — the common, EXPECTED case for an
//	                announcement made before full details are confirmed, not
//	                itself a defect; the defect this check closes is that
//	                nothing durable said so before now
//	"complete"    — evidenced and carries venue, broadcaster and >=2 participants
func eventFixtureVerdict(fact map[string]interface{}) string {
	dateText := strings.TrimSpace(getString(fact, "event_date"))
	if dateText == "" {
		return ""
	}
	src, _ := fact["source"].(map[string]interface{})
	cit, _ := src["citation"].(map[string]interface{})
	if strings.TrimSpace(getString(cit, "url")) == "" || strings.TrimSpace(getString(cit, "quote")) == "" {
		return "unevidenced"
	}
	participants, _ := fact["participants"].([]interface{})
	if strings.TrimSpace(getString(fact, "venue")) == "" ||
		strings.TrimSpace(getString(fact, "broadcaster")) == "" ||
		len(participants) < 2 {
		return "incomplete"
	}
	return "complete"
}

// getString reads a string field off an untyped map without a datahelpers
// import — this file's only two field reads (nested one level) don't
// justify the dependency, and it keeps this check's classification logic
// self-contained and trivially portable if it ever needs to run standalone.
func getString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	s, _ := m[key].(string)
	return s
}

func eventFixtureGapKey(siteID uuid.UUID) string {
	return fmt.Sprintf("event_fixture_completeness:%s", siteID.String())
}

func (c *EventFixtureCompletenessCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	var rawJSON []byte
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'evidence_base' AND is_current = true
	`, dctx.SiteID).Scan(&rawJSON)
	if err != nil {
		// No register, or a read error: nothing to audit either way. Not an
		// error condition for THIS check — most sites have no event facts at
		// all, and that is the ordinary case, not a finding.
		return &CheckResult{}, nil
	}

	var eb map[string]interface{}
	if err := json.Unmarshal(rawJSON, &eb); err != nil {
		dctx.Logger.Warn("event_fixture_completeness: evidence_base did not decode as an object",
			zap.String("site_id", dctx.SiteID.String()), zap.Error(err))
		return &CheckResult{}, nil
	}
	factsRaw, _ := eb["facts"].([]interface{})

	var unevidenced, incomplete, complete []string
	for _, fr := range factsRaw {
		fact, ok := fr.(map[string]interface{})
		if !ok {
			continue
		}
		id := getString(fact, "id")
		switch eventFixtureVerdict(fact) {
		case "unevidenced":
			unevidenced = append(unevidenced, id)
		case "incomplete":
			incomplete = append(incomplete, id)
		case "complete":
			complete = append(complete, id)
		}
	}

	result := &CheckResult{}
	// Census on every run, including a clean one — an instrument that goes
	// quiet is then visible in the run record rather than indistinguishable
	// from "nothing to find" (016b §9).
	result.Findings = append(result.Findings, map[string]interface{}{
		"check":       "event_fixture_completeness",
		"unevidenced": len(unevidenced),
		"incomplete":  len(incomplete),
		"complete":    len(complete),
	})

	gapKey := eventFixtureGapKey(dctx.SiteID)
	if len(unevidenced) == 0 && len(incomplete) == 0 {
		if len(complete) > 0 {
			// Positive observation only (RFC_010) — this site HAS event facts
			// and every one is clean, so any previously-filed gap for it is
			// now stale.
			result.Resolved = append(result.Resolved, ResolvedFinding{
				ItemType: "capability_gap",
				ItemKey:  gapKey,
				Reason:   "every event fact on this site now carries a citation, venue, broadcaster and >=2 participants",
			})
		}
		return result, nil
	}

	specJSON, err := json.Marshal(map[string]interface{}{
		"check":                "event_fixture_completeness",
		"unevidenced_fact_ids": unevidenced,
		"incomplete_fact_ids":  incomplete,
		"note": "unevidenced facts (no citation url+quote) are never rendered by query.upcoming_events and " +
			"will silently never appear on any page — they need a citation added or the fact removed. " +
			"incomplete facts (missing venue/broadcaster/<2 participants) DO render, minus the missing " +
			"fields — often correct (details not yet announced), but nothing else on the estate will tell " +
			"you which fixtures are thin without reading this row.",
	})
	if err != nil {
		return result, nil
	}

	result.WorkItems = append(result.WorkItems, WorkItemSpec{
		SiteID:   dctx.SiteID,
		Source:   "discovery",
		Pipeline: dctx.Pipeline,
		ItemType: "capability_gap",
		Severity: "low",
		Summary: fmt.Sprintf("%d event fact(s) unevidenced, %d incomplete — see spec for which",
			len(unevidenced), len(incomplete)),
		SpecJSON:     string(specJSON),
		Priority:     200,
		HandlerAgent: "", // capability_gap convention: named in spec, not dispatchable
		Status:       "deferred",
		CreatedBy:    dctx.AgentType,
		ItemKey:      gapKey,
		BatchID:      dctx.BatchID,
	})

	dctx.Logger.Info("event_fixture_completeness: complete",
		zap.Int("unevidenced", len(unevidenced)),
		zap.Int("incomplete", len(incomplete)),
		zap.Int("complete", len(complete)),
	)
	return result, nil
}
