// FILE: platform/orchestration/actions/run_discovery_checks_action.go
//
// AFTER migration — the main function becomes a simple loop.
// This replaces the ~480-line if-chain.
//
// The import of the discovery_checks package triggers all init()
// functions, registering every check automatically.

package actions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"

	// This import triggers init() in every check_*.go file
	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

// --- InputSpec (carried over from existing file) ---

var RunDiscoveryChecksInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{},
	// ConfigKeys, not Optional: both are settings read straight from
	// params.StepConfig.Config below (:66, :73) and never routed through
	// ExtractActionInputs, which is exactly the distinction ConfigKeys exists to
	// draw. Declaring them opts this action into unknown-config-key detection.
	//
	// It is opted in deliberately, to make bugs_open/136 visible at runtime: all
	// three live callers set `check_domain`, which NOTHING reads, while
	// `check_pipeline` — the key that is actually read at :66 — is set by ZERO
	// definitions fleet-wide. The rename landed in Go and never in the data, and
	// because the default at :68 is "design" the mistake is currently invisible:
	// behaviour is correct, and correct by luck.
	//
	// Declaring is safe here because it asserts nothing new about behaviour —
	// ConfigKeys is read only by UnknownConfigKeys/ListDeclaredConfigKeys and the
	// audit tool; ExtractActionInputs never looks at it (verified, not assumed).
	// So this changes what is REPORTED, not what runs. The warning is warn-only
	// (StrictConfig unset), so the three discovery agents keep working unchanged
	// while the defect finally says so out loud.
	// allow_unregistered_checks added 2026-07-30 (bugs_open/149 B4). It MUST be
	// declared here: this spec is opted into unknown-config-key detection, so an
	// undeclared lever would be reported as a stray key by the very audit this
	// list exists to feed — and the first person to set it would be told their
	// correct config was wrong.
	ConfigKeys: []string{"checks", "check_pipeline", "allow_unregistered_checks"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
	// The other half of bugs_open/136, added 2026-08-08. Declaring the key above
	// made the defect VISIBLE; this makes it stop happening. `check_domain` is
	// what all three live callers actually write, and it is now honoured.
	//
	// This IS a behaviour change and the file it changes is this one, so state it
	// here: completeness-discovery-agent asks for "content" and quality- for
	// "build", and until now both silently got the "design" default below. The
	// original filing measured that as harmless. It stopped being harmless on
	// 2026-08-04, when `content_duplication` and `page_canonical_collision` —
	// two checks that propagate dctx.Pipeline rather than hardcoding their own —
	// joined completeness-discovery-agent's list and filed four rows under
	// `design` on an agent whose config says `content`. countDispatchableWorkItems
	// (work_items_common.go) filters on pipeline, so a row under the wrong one is
	// invisible to the count that should see it.
	DeprecatedConfigKeys: map[string]string{"check_domain": "check_pipeline"},
}

func init() {
	datahelpers.RegisterActionInputSpec("run_discovery_checks", RunDiscoveryChecksInputSpec)
}

func RunDiscoveryChecksAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("RunDiscoveryChecksAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// --- Extract inputs (unchanged) ---
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		RunDiscoveryChecksInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	config := params.StepConfig.Config
	checkPipeline, _ := datahelpers.ResolveConfigSetting(
		config, RunDiscoveryChecksInputSpec, "check_pipeline", "design", logger)

	// --- Parse enabled checks from config ---
	enabledChecks := []string{"empty_sections", "undeployed_assets", "missing_css", "duplicate_palette"}
	if configChecks, ok := config["checks"].([]interface{}); ok && len(configChecks) > 0 {
		enabledChecks = make([]string, len(configChecks))
		for i, c := range configChecks {
			enabledChecks[i] = fmt.Sprintf("%v", c)
		}
	}

	var siteDomain string
	_ = params.DB.QueryRowContext(ctx,
		`SELECT domain FROM sites WHERE id = $1`, siteID,
	).Scan(&siteDomain)

	logger.Info("RunDiscoveryChecksAction: Running checks",
		zap.String("site_id", siteIDStr),
		zap.String("domain", siteDomain),
		zap.Strings("checks", enabledChecks),
		zap.Strings("registered", checks.Names()),
	)

	// --- Shared state ---
	batchID := uuid.New()
	agentType := params.ExecutionContext.Sender.AgentType
	if agentType == "" {
		agentType = "discovery-agent"
	}

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	dctx := checks.DiscoveryCheckContext{
		Ctx:       ctx,
		DB:        params.DB,
		TX:        tx,
		SiteID:    siteID,
		Pipeline:  checkPipeline,
		AgentType: agentType,
		BatchID:   batchID,
		Logger:    logger,
	}

	// --- Run each enabled check ---
	//
	// bugs_open/149 B4. Both arms below used to be a Warn and a `continue`, and
	// `checks_run` in the output reported enabledChecks — what was REQUESTED. So a
	// typo in a config array silently shrank the run and the step still reported
	// success, from outside the pod indistinguishable from a clean site. Every
	// Group B measurement in 149 was untrustworthy until this landed, because a
	// zero finding count and a check that never ran look identical.
	//
	// The two arms are treated differently, deliberately:
	//
	//   UNREGISTERED NAME → fail the step. It is a config defect, not a runtime
	//     condition: the registry is fully populated in init() before any run
	//     (including the six directory checks registered per profile at
	//     check_directory.go:111-116, which a grep for literal Name() returns
	//     cannot see), so a name that does not resolve will never resolve. This
	//     matches the contract the platform already has for its sibling — a seed
	//     naming an unregistered ACTION fails at runtime, and an unregistered
	//     CHECK should not be quieter than that.
	//
	//   CHECK ERRORED → report, do not kill the run. Failing all thirty because
	//     one hit a transient DB error would discard twenty-nine checks' findings.
	//     It is no longer silent because checks_failed names it in the output.
	//
	// Verified before shipping: of the 57 check names configured across the three
	// live agents that call this action (completeness-, design- and
	// quality-discovery-agent, measured 2026-07-30), every one resolves. Note
	// maintenance-triage's `checks` array is NOT ours — that agent has no
	// run_discovery_checks step and the array belongs to
	// scan_sites_for_maintenance, whose vocabulary is entirely different.
	var allFindings []interface{}
	inserted := 0
	resolved := 0
	skipped := 0
	ranChecks := make([]string, 0, len(enabledChecks))
	unregisteredChecks := []string{}
	failedChecks := []map[string]string{}

	// allow_unregistered_checks restores the previous tolerate-and-continue
	// behaviour. It exists for the seed-ahead-of-image window: DB config is live
	// immediately and Go is not, so a seed naming a check whose code has not
	// rolled yet would otherwise stop the whole run. Off by default — on by
	// default would leave the defect live — and withdrawable in seconds without
	// an image roll, the same argument repair_internal_links settles.
	allowUnregistered := configBoolOrDefault(config, "allow_unregistered_checks", false)

	for _, checkName := range enabledChecks {
		check := checks.Get(checkName)
		if check == nil {
			unregisteredChecks = append(unregisteredChecks, checkName)
			if !allowUnregistered {
				logger.Error("Unknown discovery check — not registered; failing the step",
					zap.String("check", checkName),
					zap.Strings("registered", checks.Names()))
				return nil, fmt.Errorf(
					"discovery check %q is not registered — the run would have silently omitted it. "+
						"Fix the agent's `checks` array, or set allow_unregistered_checks:true if the "+
						"check's code has not rolled yet. Registered checks: %v",
					checkName, checks.Names())
			}
			logger.Warn("Unknown discovery check — not registered; tolerated by allow_unregistered_checks",
				zap.String("check", checkName))
			continue
		}

		result, err := check.Run(dctx)
		if err != nil {
			logger.Warn("Discovery check failed",
				zap.String("check", checkName),
				zap.Error(err))
			failedChecks = append(failedChecks, map[string]string{
				"check": checkName,
				"error": err.Error(),
			})
			continue
		}
		ranChecks = append(ranChecks, checkName)

		// Append findings
		for _, f := range result.Findings {
			allFindings = append(allFindings, f)
		}

		// Insert work items
		for _, wi := range result.WorkItems {
			ok, err := insertWorkItem(ctx, tx, workItemFromSpec(wi), logger)
			if err != nil {
				logger.Warn("Failed to insert work item",
					zap.String("check", checkName),
					zap.Error(err))
			} else if ok {
				inserted++
			} else {
				skipped++
			}
		}

		// Retractions. AFTER the inserts, and only reachable because the
		// `err != nil` branch above already did `continue` — so a check that
		// FAILED cannot close anything, which is the property that stops a
		// blinded check quietly resolving the estate (RFC_010).
		//
		// In the same transaction as the inserts: a run that dies half-way must
		// not leave findings retracted but their replacements unwritten.
		for _, r := range result.Resolved {
			n, err := resolveWorkItems(ctx, tx, dctx.SiteID, checkName, batchID, r, logger)
			if err != nil {
				// A receipt-bearing retraction that fails here was WITHHELD on
				// purpose, not merely unlucky: resolveWorkItems refuses to close a
				// finding whose resolution destroyed something unless the record of
				// what was destroyed is durable first (bugs_open/469). The flag is
				// what lets a reader tell the two apart in the logs.
				logger.Warn("Failed to resolve work items",
					zap.String("check", checkName),
					zap.String("item_type", r.ItemType),
					zap.Bool("receipt_required", r.Receipt != nil),
					zap.Error(err))
				continue
			}
			resolved += n
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit work items: %w", err)
	}

	// A check that ERRORED is now recorded durably, not just in the step output
	// (bugs_open/149 B4 follow-up; council 2d0dbc2e, bug_historian edit 3).
	//
	// WHY THIS IS NEEDED AT ALL, since the failure is already in checks_failed:
	// pod logs roll and `collected_data` is pruned at ~24h, so "which check
	// silently stopped working, and when" was unanswerable a day later. That is
	// the estate's "a pod log line is not a record" rule (bugs_open/071 gap 3),
	// which the claims floor in the same commit honours with
	// CONTENT_CLAIMS_FLOOR_DETAIL and this path did not. The asymmetry inside one
	// commit was the defect, not the missing feature.
	//
	// AFTER the commit deliberately: the work items are the run's real output and
	// must not be rolled back because a diagnostic write failed. Correspondingly
	// this is best-effort — see writeDiscoveryCheckErrorLog.
	writeDiscoveryCheckErrorLog(ctx, params, siteID, siteDomain, agentType,
		checkPipeline, batchID, failedChecks, logger)

	logger.Info("RunDiscoveryChecksAction: Complete",
		zap.Int("findings", len(allFindings)),
		zap.Int("items_inserted", inserted),
		zap.Int("items_skipped", skipped),
		zap.Int("items_resolved", resolved),
		zap.Int("checks_requested", len(enabledChecks)),
		zap.Int("checks_ran", len(ranChecks)),
		zap.Int("checks_failed", len(failedChecks)),
	)

	// checks_run now reports what ACTUALLY ran, not what was requested
	// (bugs_open/149 B4). checks_requested is kept alongside it so the difference
	// is readable without re-deriving it, and so a consumer that wanted the
	// requested list has somewhere to go rather than silently reading the wrong
	// one. A downstream reader that treated checks_run as "the configured set" is
	// now reading something narrower — which is the correction, not a regression.
	return map[string]interface{}{
		"site_id":             siteIDStr,
		"domain":              siteDomain,
		"checks_run":          ranChecks,
		"checks_requested":    enabledChecks,
		"checks_unregistered": unregisteredChecks,
		"checks_failed":       failedChecks,
		"findings":            allFindings,
		"items_inserted":      inserted,
		"items_skipped":       skipped,
		"items_resolved":      resolved,
		"batch_id":            batchID.String(),
	}, nil
}

// discoveryCheckErrorCode is DELIBERATELY distinct from every other code in
// agent_error_log, for the reason recorded at validate_page_content.go:619-624 and
// repeated in save_sections_claims_guard.go:97-103: the rows answer different
// questions, and a shared code makes "which path caught this" unanswerable — the
// drift bugs_open/097 exists to keep answerable. Checked against the live table
// 2026-07-31: 16 distinct codes in use, none of them this one.
const discoveryCheckErrorCode = "DISCOVERY_CHECK_ERROR"

// writeDiscoveryCheckErrorLog records one row per check that returned an error.
//
// ONE ROW PER CHECK, not one row per run. The question this exists to answer is
// "which check stopped working, and when" — a batch-level row would force every
// reader to parse a list back out of `context` to answer it, and would collapse
// two checks failing for different reasons into one indistinguishable event.
//
// SEVERITY IS `warning`, NOT `error`. A check erroring does not mean the site has
// a defect, and it does not fail the step: the run continues and its other checks
// still report. Grading this `error` would put a harness fault in the same bucket
// as a real content finding, which is the confusion validate_page_content_stats.go
// states as a rule — "severity `error` must never mean 'we could not check this'".
// That rule cuts both ways, and this is its other edge.
//
// BEST-EFFORT BY CONSTRUCTION: every failure path warns and returns. A diagnostic
// write must never be able to fail a run whose actual work — the work items — has
// already been committed.
func writeDiscoveryCheckErrorLog(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	domain, agentType, checkPipeline string,
	batchID uuid.UUID,
	failedChecks []map[string]string,
	logger *zap.Logger,
) {
	if params.DB == nil || len(failedChecks) == 0 {
		return
	}

	// agent_type is NOT NULL on this table. The caller already defaults it to
	// "discovery-agent", but this function is called from tests and future callers
	// too, so it does not rely on that.
	if agentType == "" {
		agentType = "unknown"
	}

	stepName := "run_discovery_checks"
	if params.ExecutionContext != nil && params.ExecutionContext.StepName != "" {
		stepName = params.ExecutionContext.StepName
	}

	findings := make([]agenterrors.Finding, 0, len(failedChecks))
	for _, fc := range failedChecks {
		checkName := fc["check"]
		checkErr := fc["error"]
		findings = append(findings, agenterrors.Finding{
			ErrorCode: discoveryCheckErrorCode,
			Severity:  "warning",
			Message: fmt.Sprintf("Discovery check %q errored and was skipped — the site was NOT checked for this class: %s",
				checkName, checkErr),
			Context: map[string]interface{}{
				"check":          checkName,
				"check_error":    checkErr,
				"check_pipeline": checkPipeline,
				"batch_id":       batchID.String(),
			},
		})
	}

	// agentType and stepName are this function's own, defaulted above — set
	// explicitly so the merge cannot substitute the executing step's identity.
	attempted, recorded := LogActionEntryFindings(ctx, params, agenterrors.Entry{
		SiteID:    siteID.String(),
		Domain:    domain,
		AgentType: agentType,
		StepName:  stepName,
		Action:    "run_discovery_checks",
	}, findings, logger)
	if attempted != recorded {
		logger.Warn("RunDiscoveryChecksAction: failed to write some discovery check error records",
			zap.Int("attempted", attempted),
			zap.Int("recorded", recorded))
	}
}
