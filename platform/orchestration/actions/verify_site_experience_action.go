// FILE: platform/orchestration/actions/verify_site_experience_action.go
//
// verify_site_experience — run a bound fork's criteria against the deployed page
// and move the fork, and possibly the entry, along their lifecycles.
//
// This is the piece that makes the register do something. Everything before it
// records intent: an entry says a control must lead somewhere real, a fork says
// which page that is. This runs the resulting check against the page that is
// actually being served, and writes down what happened.
//
// THE STATE MACHINE, AND WHY EACH TRANSITION NEEDS EVIDENCE
//
//	fork:  bound ──green──> verified          ──red──> broken
//	entry: approved ──first green fork──> proven
//
// `proven` is the one worth guarding hardest. It does not mean a reviewer liked
// the entry — that is `approved`, and a council does it. It means the entry's
// criteria have run green against a real page at least once, i.e. that the
// entry is not merely well-formed but describes something achievable. An entry
// that can be `proven` without a green run would make the strongest word in the
// vocabulary the cheapest.
//
// # THE VACUOUS PASS, WHICH IS THE FAILURE THIS ACTION EXISTS TO NOT COMMIT
//
// A criteria document can produce zero executed checks: every check deferred
// (the platform cannot run it), or every check skipped (Tier 2 could not anchor
// the selector, or the id ends in -EDIT). Such a run has no failures. Read as
// "no failures ⇒ pass", it verifies a fork that asserted nothing — which is
// precisely the class of defect the register was built to end, committed by the
// register itself.
//
// So the rule is: **at least one check PASSED and zero FAILED.** Never "no
// failures". A run with zero executed checks is recorded as `inconclusive` and
// changes no status — it is not a pass and it is not a breakage, it is an
// absence of evidence, and the one thing it must not do is look like either.
//
// Registration (registry.go):
//
//	"verify_site_experience": {
//	    Handler:     VerifySiteExperienceAction,
//	    Category:    "experience_register",
//	    Description: "Run a bound fork's criteria against the deployed page and record the outcome",
//	    IsLocal:     true,
//	}
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var VerifySiteExperienceInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{
		"pattern_name", "pattern_name_field", "instance_key",
		"site_id_field", "page_url", "page_url_field", "dry_run",
	},
	Defaults: map[string]interface{}{
		"pattern_name_field": "input_data.pattern_name",
		"site_id_field":      "input_data.site_id",
		"page_url_field":     "input_data.page_url",
		"instance_key":       "default",
		"dry_run":            false,
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("verify_site_experience", VerifySiteExperienceInputSpec)
}

// experienceVerifyTimeout — a page that will not answer in this long is not a
// page a visitor is using either.
const experienceVerifyTimeout = 20 * time.Second

// experienceBindingRE matches the placeholders a fork substitutes. Same syntax
// as the validator's, deliberately: a template that closed at write time must
// substitute here, or the two disagree about what a placeholder is.
func substituteExperienceBindings(doc string, bindings map[string]interface{}) (string, []string) {
	var unresolved []string
	out := experiencePlaceholderRE.ReplaceAllStringFunc(doc, func(m string) string {
		sub := experiencePlaceholderRE.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		v, ok := bindings[sub[1]]
		if !ok {
			unresolved = append(unresolved, sub[1])
			return m
		}
		s := experienceString(v)
		if strings.TrimSpace(s) == "" {
			unresolved = append(unresolved, sub[1])
			return m
		}
		// JSON-escape: the substituted value lands inside a JSON document.
		b, err := json.Marshal(s)
		if err != nil {
			unresolved = append(unresolved, sub[1])
			return m
		}
		return strings.Trim(string(b), `"`)
	})
	return out, unresolved
}

func VerifySiteExperienceAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "verify_site_experience"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	config := params.StepConfig.Config

	siteID := datahelpers.GetStringField(config, "site_id", "")
	if siteID == "" {
		siteID = datahelpers.ExtractNestedFieldString(params.CollectedData,
			datahelpers.GetStringField(config, "site_id_field", "input_data.site_id"))
	}
	patternName := datahelpers.GetStringField(config, "pattern_name", "")
	if patternName == "" {
		patternName = datahelpers.ExtractNestedFieldString(params.CollectedData,
			datahelpers.GetStringField(config, "pattern_name_field", "input_data.pattern_name"))
	}
	if siteID == "" || patternName == "" {
		return nil, fmt.Errorf("verify_site_experience: need site_id and pattern_name")
	}
	instanceKey := datahelpers.GetStringField(config, "instance_key", "default")

	// ── load the fork and its entry together ───────────────────────────────
	var (
		forkID, forkStatus, patternStatus string
		bindingsJSON, criteriaJSON        []byte
		executable                        int
	)
	err := params.DB.QueryRowContext(ctx, `
		SELECT se.id, se.status, se.bindings,
		       p.status, COALESCE(p.criteria_template, '{}'::jsonb), p.executable_checks
		FROM site_experiences se
		JOIN experience_patterns p ON p.name = se.pattern_name
		WHERE se.site_id = $1::uuid AND se.pattern_name = $2 AND se.instance_key = $3`,
		siteID, patternName, instanceKey).
		Scan(&forkID, &forkStatus, &bindingsJSON, &patternStatus, &criteriaJSON, &executable)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("verify_site_experience: no fork of %q on this site (instance %q) — bind it first",
			patternName, instanceKey)
	}
	if err != nil {
		return nil, fmt.Errorf("verify_site_experience: loading fork: %w", err)
	}

	// DRY RUN — evaluate and report, write NOTHING.
	//
	// This exists because of a real gap in the lifecycle, not as a convenience.
	// A fork of a `draft` entry is recorded `proposed`, and a `proposed` fork
	// cannot be verified, because verifying it would attach a live-sounding
	// status to a contract nobody has approved. But per-experience approval is
	// designed and NOT BUILT, so today every entry is draft, every fork is
	// proposed, and the whole path would be unexercisable — which is how a
	// mechanism rots before it is ever used.
	//
	// A dry run breaks that without faking anything: it runs the real checks
	// against the real page and returns the real result, while writing no status
	// and no timestamp. It is the difference between "we cannot check this yet"
	// and "we have never once run it".
	dryRun := datahelpers.GetBoolField(config, "dry_run", false)

	// A `proposed` fork has not been bound against an approved entry; recording
	// a verdict on it would attach a live-sounding status to a contract nobody
	// approved. Reporting on it is fine — claiming is not.
	if forkStatus == "proposed" && !dryRun {
		return nil, fmt.Errorf("verify_site_experience: fork of %q is 'proposed', not bound — its entry is a draft, so there is nothing here it would be honest to verify. Re-run with dry_run to see what the checks WOULD say", patternName)
	}

	bindings := map[string]interface{}{}
	if err := json.Unmarshal(bindingsJSON, &bindings); err != nil {
		return nil, fmt.Errorf("verify_site_experience: decoding bindings: %w", err)
	}

	resolved, unresolved := substituteExperienceBindings(string(criteriaJSON), bindings)
	if len(unresolved) > 0 {
		// Closure was checked at bind time; if it fails now the entry changed
		// underneath the fork. That is a real event and must not be papered over
		// by running the checks anyway.
		return nil, fmt.Errorf("verify_site_experience: %d placeholder(s) no longer resolve (%s) — the entry changed after this fork was bound; re-bind before verifying",
			len(unresolved), strings.Join(unresolved, ", "))
	}

	pageURL := datahelpers.GetStringField(config, "page_url", "")
	if pageURL == "" {
		pageURL = datahelpers.ExtractNestedFieldString(params.CollectedData,
			datahelpers.GetStringField(config, "page_url_field", "input_data.page_url"))
	}
	if pageURL == "" {
		return nil, fmt.Errorf("verify_site_experience: no page_url to check against")
	}

	status, html, fetchErr := fetchExperiencePage(ctx, pageURL)
	if fetchErr != nil {
		// Could not look. That is not a breakage of the experience, and
		// recording it as one would blame the site for our network.
		return experienceInconclusive(ctx, params, logger, forkID, patternName,
			fmt.Sprintf("could not fetch %s: %v", pageURL, fetchErr))
	}

	// Only Tier-2-executable checks may be judged by the Tier 2 evaluator.
	//
	// CORRECTED 2026-07-28, from the first live run. This used to hand the whole
	// document to the static evaluator, which is wrong in the dangerous
	// direction: a check declaring `tier: 4` — or one whose TYPE only exists in
	// the browser runner — would be evaluated statically anyway, and for
	// `selector_count` a static evaluation is *vacuously green* (it confirms the
	// container's anchor and never counts anything). A verification resting on
	// that is the vacuous pass wearing a different hat.
	staticDoc, tier4 := splitExperienceChecksByTier([]byte(resolved))

	ev, err := discovery_checks.EvaluateStaticCriteriaJSON(staticDoc, status, html)
	if err != nil {
		return nil, fmt.Errorf("verify_site_experience: criteria did not parse after substitution: %w", err)
	}
	// Reported, never counted: a check awaiting a browser is not a pass and not
	// a failure, and the whole discipline here is that those stay visible.
	for _, id := range tier4 {
		ev.Skipped = append(ev.Skipped, discovery_checks.StaticCheckOutcome{
			ID: id, Detail: "requires Tier 4 (a real browser); not judged by the static evaluator",
		})
	}

	result := map[string]interface{}{
		"page_url":     pageURL,
		"http_status":  status,
		"passed":       ev.Passed,
		"failed":       ev.Failed,
		"skipped":      ev.Skipped,
		"passed_count": len(ev.Passed),
		"failed_count": len(ev.Failed),
		"deferred_at_write_time": map[string]interface{}{
			"executable_checks": executable,
		},
	}
	resultJSON, _ := json.Marshal(result)

	// ── the decision ───────────────────────────────────────────────────────
	//
	// At least one PASSED and zero FAILED. Not "no failures": a document whose
	// checks were all skipped or all deferred also has no failures, and calling
	// that verified would be the register committing the defect it exists to
	// find.
	verdict := experienceVerdict(ev)

	if dryRun {
		// Side-effect-free ON PURPOSE: not even last_check_result. A dry run
		// that leaves a trace is one somebody later reads as a real result.
		logger.Info("verify_site_experience: DRY RUN (nothing written)",
			zap.String("pattern", patternName),
			zap.String("would_be", verdict),
			zap.Int("passed", len(ev.Passed)),
			zap.Int("failed", len(ev.Failed)),
			zap.Int("skipped", len(ev.Skipped)))
		return map[string]interface{}{
			"dry_run":       true,
			"pattern_name":  patternName,
			"would_be":      verdict,
			"fork_status":   forkStatus + " (unchanged)",
			"passed":        len(ev.Passed),
			"failed":        len(ev.Failed),
			"skipped":       len(ev.Skipped),
			"failed_checks": ev.Failed,
			"summary": fmt.Sprintf("DRY RUN: would be %s — %d passed, %d failed, %d skipped",
				verdict, len(ev.Passed), len(ev.Failed), len(ev.Skipped)),
		}, nil
	}

	switch verdict {
	case "broken":
		return experienceRecord(ctx, params, logger, forkID, patternName, "broken", resultJSON,
			fmt.Sprintf("%d check(s) failed", len(ev.Failed)), ev, false)

	case "inconclusive":
		// Status UNCHANGED. Nothing executed, so there is no evidence in either
		// direction, and the one thing this must not do is look like either.
		return experienceRecord(ctx, params, logger, forkID, patternName, forkStatus, resultJSON,
			fmt.Sprintf("INCONCLUSIVE: nothing executed (%d skipped) — no failures is not a pass", len(ev.Skipped)),
			ev, false)

	default:
		promote := patternStatus == "approved"
		return experienceRecord(ctx, params, logger, forkID, patternName, "verified", resultJSON,
			fmt.Sprintf("%d passed, 0 failed, %d skipped", len(ev.Passed), len(ev.Skipped)),
			ev, promote)
	}
}

// splitExperienceChecksByTier returns the criteria document containing only the
// checks a Tier 2 static evaluation may legitimately judge, plus the ids of
// those it may not.
//
// A check is held back when it is marked DEFERRED, when it declares `tier: 4`,
// or when its type is one only the browser runner implements. Three different
// voices saying the same thing: the register (we already know this cannot run),
// the author (this needs a browser), and the platform (this type only exists at
// Tier 4). Any one of them is sufficient.
//
// A malformed document is returned unchanged — parsing is the evaluator's job
// to complain about, and swallowing the error here would hide it.
func splitExperienceChecksByTier(doc []byte) ([]byte, []string) {
	var parsed map[string]interface{}
	if err := json.Unmarshal(doc, &parsed); err != nil {
		return doc, nil
	}
	checks, ok := parsed["checks"].([]interface{})
	if !ok {
		return doc, nil
	}

	kept := make([]interface{}, 0, len(checks))
	var held []string
	for _, c := range checks {
		cm, ok := c.(map[string]interface{})
		if !ok {
			kept = append(kept, c)
			continue
		}
		id := experienceString(cm["id"])

		// A check the REGISTER ITSELF recorded as deferred must not be executed.
		//
		// Found on the first live run, and it is the sharpest instance of "fix
		// the harness, never the page" in this whole workstream: CC-001's
		// `feed_loads` is marked unexecutable — `asset_loads` only matches the
		// path as text in the page HTML, and that component's loader is an
		// external bundle — yet the consumer ran it anyway and reported a
		// FAILURE. A correct page was called broken by a check we had already
		// written down as impossible. Deferral has to bind the consumer too, or
		// it is only a comment.
		if reason, deferred := experienceDeferralReason(cm); deferred {
			held = append(held, id+" (deferred: "+reason+")")
			continue
		}

		needsBrowser := false

		// The author said so.
		switch t := cm[experienceTierKey].(type) {
		case float64:
			needsBrowser = t > 2
		case string:
			needsBrowser = t == "4"
		}
		// Or the platform says so: the type exists only at Tier 4.
		if tier, known := experienceCheckTiers[experienceString(cm["type"])]; known && tier > 2 {
			needsBrowser = true
		}

		if needsBrowser {
			held = append(held, id)
			continue
		}
		kept = append(kept, c)
	}

	parsed["checks"] = kept
	out, err := json.Marshal(parsed)
	if err != nil {
		return doc, nil
	}
	return out, held
}

// experienceVerdict is the whole decision, in one place so it can be tested
// without a database or a web server — because it is the rule most likely to be
// "simplified" later by someone who reads `len(failed) == 0` as sufficient.
//
//	broken        at least one check FAILED
//	inconclusive  nothing executed — every check skipped or the document empty
//	verified      at least one PASSED and zero FAILED
//
// The middle case is the point. A run with no failures and no passes is not a
// pass; treating it as one verifies a fork that asserted nothing.
func experienceVerdict(ev discovery_checks.StaticEvaluation) string {
	switch {
	case len(ev.Failed) > 0:
		return "broken"
	case len(ev.Passed) == 0:
		return "inconclusive"
	default:
		return "verified"
	}
}

// experienceRecord writes the outcome and, on a first green run of an approved
// entry, promotes it to `proven`.
func experienceRecord(ctx context.Context, params ActionParams, logger *zap.Logger,
	forkID, patternName, forkStatus string, resultJSON []byte, summary string,
	ev discovery_checks.StaticEvaluation, promote bool) (interface{}, error) {

	_, err := params.DB.ExecContext(ctx, `
		UPDATE site_experiences
		SET status = $2, last_checked_at = now(), last_check_result = $3::jsonb, updated_at = now()
		WHERE id = $1::uuid`, forkID, forkStatus, string(resultJSON))
	if err != nil {
		return nil, fmt.Errorf("verify_site_experience: recording result: %w", err)
	}

	promoted := false
	if promote {
		// `proven` is earned by evidence, so the promotion is conditional on the
		// row still being `approved` at write time — a concurrent demotion
		// (someone changed the contract) must win, not be overwritten by a run
		// that was measuring the old one.
		res, err := params.DB.ExecContext(ctx, `
			UPDATE experience_patterns SET status = 'proven', updated_at = now()
			WHERE name = $1 AND status = 'approved'`, patternName)
		if err != nil {
			return nil, fmt.Errorf("verify_site_experience: promoting %q: %w", patternName, err)
		}
		n, _ := res.RowsAffected()
		promoted = n > 0
	}

	fields := []zap.Field{
		zap.String("pattern", patternName),
		zap.String("fork_status", forkStatus),
		zap.Int("passed", len(ev.Passed)),
		zap.Int("failed", len(ev.Failed)),
		zap.Int("skipped", len(ev.Skipped)),
		zap.Bool("entry_promoted_to_proven", promoted),
	}
	if forkStatus == "broken" {
		logger.Warn("verify_site_experience: BROKEN — "+summary, fields...)
	} else {
		logger.Info("verify_site_experience: "+summary, fields...)
	}

	return map[string]interface{}{
		"pattern_name":  patternName,
		"fork_status":   forkStatus,
		"summary":       summary,
		"passed":        len(ev.Passed),
		"failed":        len(ev.Failed),
		"skipped":       len(ev.Skipped),
		"promoted":      promoted,
		"failed_checks": ev.Failed,
	}, nil
}

// experienceInconclusive records that we could not look, without moving the
// fork's status. An absence of evidence must look like neither a pass nor a
// breakage.
func experienceInconclusive(ctx context.Context, params ActionParams, logger *zap.Logger,
	forkID, patternName, detail string) (interface{}, error) {

	body, _ := json.Marshal(map[string]interface{}{"inconclusive": detail})
	if _, err := params.DB.ExecContext(ctx, `
		UPDATE site_experiences
		SET last_checked_at = now(), last_check_result = $2::jsonb, updated_at = now()
		WHERE id = $1::uuid`, forkID, string(body)); err != nil {
		return nil, fmt.Errorf("verify_site_experience: recording inconclusive: %w", err)
	}
	logger.Warn("verify_site_experience: INCONCLUSIVE (status unchanged)",
		zap.String("pattern", patternName), zap.String("detail", detail))
	return map[string]interface{}{
		"pattern_name": patternName,
		"fork_status":  "unchanged",
		"summary":      "INCONCLUSIVE: " + detail,
	}, nil
}

func fetchExperiencePage(ctx context.Context, url string) (int, string, error) {
	ctx, cancel := context.WithTimeout(ctx, experienceVerifyTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, "", err
	}
	// A real browser UA: some fronts serve a different page to an unknown agent,
	// and verifying a page no visitor sees proves nothing about the one they do.
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return resp.StatusCode, "", err
	}
	return resp.StatusCode, string(body), nil
}
