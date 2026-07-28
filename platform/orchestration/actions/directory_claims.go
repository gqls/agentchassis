// FILE: platform/orchestration/actions/directory_claims.go
//
// Model directory pipeline (docs/agent_docs/docs024_key_docs_latest/
// model_directory_pipeline/), Phase B: a global, cross-site registry of
// cited facts about AI models now, companies/protocols later
// (directory_entities/directory_claims — migration 192).
//
// This file adds NOTHING to how a citation is verified — verifyCitationLive
// and datahelpers.QuoteFoundInText (both already built for the V5
// claims-verification layer, evidence_citations.go) are reused UNCHANGED.
// The only thing new here is WHERE the verified fact lives: a global
// registry rather than a per-site site_specs.evidence_base blob, because a
// model's price is not a fact about any one site.
//
// Two actions:
//
//   verify_and_register_directory_claims — acquisition. Given a batch of
//   candidate claims (each names its entity by kind+slug and carries a
//   citation), find-or-create the entity, verify each citation LIVE, and
//   register only the survivors. The model proposes; the string comparison
//   disposes — identical discipline to verify_and_register_citations.
//   Rejects raise a directory_citation_unverified work item. This registry
//   has no per-site owner, so work items anchor to the system.internal
//   pseudo-site — the same precedent diagnose_triage_action.go already uses
//   for fleet-wide, non-site-scoped findings.
//
//   refresh_directory_claims — freshness. Re-verifies is_current claims
//   whose verified_at has passed their own staleness_days. A claim whose
//   verification STATUS is unchanged (still found) is bumped in place
//   (verified_at only — no version churn for a no-op check). A claim whose
//   status CHANGES (found ⇄ citation_lost ⇄ fetch_error, including
//   recovery) supersedes into a new current row, same idiom as
//   write_site_spec, re-keyed to a relational row instead of a jsonb
//   aspect — a status transition is never a silent in-place edit.
//
// fetch_error is never treated as citation_lost (a paywall going up is not
// evidence the fact is wrong — evidence_citations.go's header makes the
// same distinction); a claim currently fetch_error is still swept next time
// and recovers automatically if the source becomes reachable again.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// directorySystemSiteID anchors registry-wide work items to the
// system.internal pseudo-site: this data has no per-site owner. Same
// constant value as diagnose_triage_action.go's triageSystemSiteID.
const directorySystemSiteID = "eac60db8-b032-432b-b36d-76f37632045d"

// ============================================================================
// ACTION: verify_and_register_directory_claims
// ============================================================================
//
// Config:
//   "candidates": path to the extracted array (required) — each element:
//       {entity_kind (default "model"), entity_slug, entity_name?,
//        entity_owner?, entity_summary?, field, value?, unit?,
//        staleness_days?, quote, url, publisher, title, published?}
//   "dry_run": verify and report, write nothing (optional, default false)

var VerifyAndRegisterDirectoryClaimsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("verify_and_register_directory_claims", VerifyAndRegisterDirectoryClaimsInputSpec)
}

type rejectedDirectoryClaim struct {
	Slug   string `json:"slug,omitempty"`
	Field  string `json:"field,omitempty"`
	URL    string `json:"url,omitempty"`
	Class  string `json:"class"`
	Detail string `json:"detail"`
}

func VerifyAndRegisterDirectoryClaimsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "verify_and_register_directory_claims"))
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	dryRun := configBoolOrDefault(params.StepConfig.Config, "dry_run", false)

	candidatesPath, _ := params.StepConfig.Config["candidates"].(string)
	if candidatesPath == "" {
		return nil, fmt.Errorf("config 'candidates' (dot-path to the extracted claims array) is required")
	}
	rawCandidates := datahelpers.ExtractNestedField(params.CollectedData, candidatesPath)
	candidateList, ok := rawCandidates.([]interface{})
	if !ok {
		return nil, fmt.Errorf("candidates at %q is not an array (got %T)", candidatesPath, rawCandidates)
	}

	today := currentDateString(ctx, params.DB)

	var registered []string
	var failures []rejectedDirectoryClaim

	for _, cr := range candidateList {
		cand, ok := cr.(map[string]interface{})
		if !ok {
			continue
		}
		kind := strings.TrimSpace(datahelpers.GetStringField(cand, "entity_kind", "model"))
		slug := strings.TrimSpace(datahelpers.GetStringField(cand, "entity_slug", ""))
		field := strings.TrimSpace(datahelpers.GetStringField(cand, "field", ""))
		if slug == "" || field == "" {
			failures = append(failures, rejectedDirectoryClaim{Slug: slug, Field: field,
				Class: "citation_invalid", Detail: "candidate missing entity_slug or field"})
			continue
		}

		cit, cerr := datahelpers.ParseCitation(map[string]interface{}{"citation": map[string]interface{}{
			"publisher": cand["publisher"], "title": cand["title"], "url": cand["url"],
			"published": cand["published"], "quote": cand["quote"], "accessed": today,
		}})
		if cerr != nil || cit == nil {
			detail := "candidate missing citation fields"
			if cerr != nil {
				detail = cerr.Error()
			}
			failures = append(failures, rejectedDirectoryClaim{Slug: slug, Field: field,
				URL: datahelpers.GetStringField(cand, "url", ""), Class: "citation_invalid", Detail: detail})
			continue
		}

		// The disposal step: the model proposed this; the source must confirm it.
		outcome := verifyCitationLive(ctx, cit)
		if !outcome.Found {
			failures = append(failures, rejectedDirectoryClaim{Slug: slug, Field: field,
				URL: cit.URL, Class: outcome.FailClass, Detail: outcome.FailDetail})
			continue
		}

		if dryRun {
			registered = append(registered, slug+"."+field)
			continue
		}

		entityID, err := upsertDirectoryEntity(ctx, params.DB, kind, slug, cand)
		if err != nil {
			logger.Warn("verify_and_register_directory_claims: entity upsert failed",
				zap.String("slug", slug), zap.Error(err))
			failures = append(failures, rejectedDirectoryClaim{Slug: slug, Field: field,
				URL: cit.URL, Class: "citation_invalid", Detail: "entity upsert: " + err.Error()})
			continue
		}

		citationJSON, _ := json.Marshal(map[string]interface{}{
			"publisher": cit.Publisher, "title": cit.Title, "url": cit.URL,
			"published": cit.Published, "quote": cit.Quote, "accessed": today,
		})
		staleness := 200
		if v, ok := numericField(cand["staleness_days"]); ok {
			staleness = int(v)
		}
		if err := writeCurrentDirectoryClaim(ctx, params.DB, entityID, field,
			datahelpers.GetStringField(cand, "value", ""), datahelpers.GetStringField(cand, "unit", ""),
			citationJSON, "found", staleness, params.AgentType); err != nil {
			logger.Warn("verify_and_register_directory_claims: claim write failed",
				zap.String("slug", slug), zap.String("field", field), zap.Error(err))
			failures = append(failures, rejectedDirectoryClaim{Slug: slug, Field: field,
				URL: cit.URL, Class: "citation_invalid", Detail: "claim write: " + err.Error()})
			continue
		}
		registered = append(registered, slug+"."+field)
	}

	if !dryRun && len(failures) > 0 {
		if err := createDirectoryCitationFailuresItem(ctx, params, failures, logger); err != nil {
			logger.Warn("verify_and_register_directory_claims: failed to create review item", zap.Error(err))
		}
	}

	logger.Info("verify_and_register_directory_claims: complete",
		zap.Int("candidates", len(candidateList)),
		zap.Int("registered", len(registered)),
		zap.Int("rejected", len(failures)),
		zap.Bool("dry_run", dryRun))

	return map[string]interface{}{
		"candidates": len(candidateList),
		"registered": registered,
		"rejected":   failures,
		"dry_run":    dryRun,
	}, nil
}

// upsertDirectoryEntity finds or creates a directory_entities row by
// (kind, slug) in one atomic statement (no separate SELECT-then-INSERT —
// that would race two concurrent researchers discovering the same entity).
// An existing row's soft fields (name/owner/summary) are refreshed only
// when the candidate supplies a non-empty value, so an absent field on a
// later research pass never blanks an already-curated one. links/attributes
// are shallow-merged (jsonb ||): a candidate need not repeat every link on
// every claim row, and a later pass can add a link key without touching the
// ones already there.
func upsertDirectoryEntity(ctx context.Context, db *sql.DB, kind, slug string, cand map[string]interface{}) (uuid.UUID, error) {
	name := strings.TrimSpace(datahelpers.GetStringField(cand, "entity_name", ""))
	if name == "" {
		name = slug
	}
	owner := datahelpers.GetStringField(cand, "entity_owner", "")
	summary := datahelpers.GetStringField(cand, "entity_summary", "")
	links := jsonObjectField(cand, "entity_links")
	attributes := jsonObjectField(cand, "entity_attributes")

	var id uuid.UUID
	err := db.QueryRowContext(ctx, `
		INSERT INTO directory_entities (kind, slug, name, owner, summary, links, attributes, discovered_by)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6::jsonb, $7::jsonb, $8)
		ON CONFLICT (kind, slug) DO UPDATE SET
			name       = COALESCE(NULLIF(EXCLUDED.name, ''), directory_entities.name),
			owner      = COALESCE(NULLIF(EXCLUDED.owner, ''), directory_entities.owner),
			summary    = COALESCE(NULLIF(EXCLUDED.summary, ''), directory_entities.summary),
			links      = directory_entities.links || EXCLUDED.links,
			attributes = directory_entities.attributes || EXCLUDED.attributes,
			updated_at = now()
		RETURNING id
	`, kind, slug, name, owner, summary, links, attributes, "directory-researcher").Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("upsert entity: %w", err)
	}
	return id, nil
}

// jsonObjectField extracts an optional nested object field (e.g.
// entity_links: {docs, weights, wrapper_url, video_urls}) as a JSON string,
// defaulting to "{}" when absent or the wrong shape — never an error, since
// these are optional enrichments, not verifiable claims.
func jsonObjectField(cand map[string]interface{}, key string) string {
	obj, ok := cand[key].(map[string]interface{})
	if !ok {
		return "{}"
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// writeCurrentDirectoryClaim supersedes any existing current claim for
// (entity_id, field) and inserts the new one as current — the same
// supersede-then-insert idiom write_site_spec uses for site_specs, re-keyed
// to a relational row instead of a jsonb aspect.
func writeCurrentDirectoryClaim(
	ctx context.Context, db *sql.DB, entityID uuid.UUID, field, value, unit string,
	citationJSON []byte, status string, stalenessDays int, createdBy string,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE directory_claims SET is_current = false, superseded_at = now()
		WHERE entity_id = $1 AND field = $2 AND is_current = true
	`, entityID, field); err != nil {
		return fmt.Errorf("supersede claim: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO directory_claims
			(entity_id, field, value, unit, citation, status, staleness_days, verified_at, created_by)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5::jsonb, $6, $7, now(), $8)
	`, entityID, field, value, unit, string(citationJSON), status, stalenessDays, createdBy); err != nil {
		return fmt.Errorf("insert claim: %w", err)
	}

	return tx.Commit()
}

// createDirectoryCitationFailuresItem raises the rejects for a human.
// HITL-terminal, matching citation_unverified: a citation the machine could
// not verify is a decision for a person, never a silent drop and never a
// fact. itemKey is static (no per-batch suffix) — same as
// evidence_citations.go's citation_unverified key — so a second rejection
// while an item is still open does not create item churn.
func createDirectoryCitationFailuresItem(
	ctx context.Context, params ActionParams, failures []rejectedDirectoryClaim, logger *zap.Logger,
) error {
	specJSON, err := json.Marshal(map[string]interface{}{
		"check":    "directory_citation_verification",
		"rejected": failures,
		"fix": "Each rejected candidate either cited a source that does not contain its quote " +
			"(citation_lost / possible hallucination — discard or re-research), could not be fetched " +
			"(fetch_error — retry later, or mark reverifiable:false with a human attestation), or was " +
			"structurally incomplete (citation_invalid). None was registered.",
	})
	if err != nil {
		return err
	}
	siteID, err := uuid.Parse(directorySystemSiteID)
	if err != nil {
		return err
	}
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "research",
		pipeline:     "content",
		itemType:     "directory_citation_unverified",
		severity:     "medium",
		summary:      "Model directory: candidate claims failed live verification and need a human ruling",
		spec:         string(specJSON),
		priority:     35,
		handlerAgent: "human-review",
		status:       "needs_human_review",
		createdBy:    params.AgentType,
		itemKey:      "directory_citation_unverified",
	}, logger)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// ============================================================================
// ACTION: refresh_directory_claims
// ============================================================================
//
// Config:
//   "kind":    optional filter (e.g. "model") — omit to sweep every kind.
//   "dry_run": report only, write nothing (optional, default false)

var RefreshDirectoryClaimsInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{},
	Optional:    []string{"kind"},
	Defaults:    map[string]interface{}{},
	Deprecated:  map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("refresh_directory_claims", RefreshDirectoryClaimsInputSpec)
}

type directoryClaimRefresh struct {
	EntityID string `json:"entity_id"`
	Slug     string `json:"slug"`
	Field    string `json:"field"`
	Outcome  string `json:"outcome"` // fresh | citation_lost | fetch_error | recovered | error
	Detail   string `json:"detail,omitempty"`
}

type dueDirectoryClaim struct {
	id            uuid.UUID
	entityID      uuid.UUID
	slug, field   string
	value, unit   string
	citationJSON  []byte
	status        string
	stalenessDays int
}

func RefreshDirectoryClaimsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "refresh_directory_claims"))
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, RefreshDirectoryClaimsInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	kind := inputs.Get("kind")
	dryRun := configBoolOrDefault(params.StepConfig.Config, "dry_run", false)

	due, err := loadDueDirectoryClaims(ctx, params.DB, kind)
	if err != nil {
		return nil, err
	}
	if len(due) == 0 {
		logger.Info("refresh_directory_claims: nothing due")
		return map[string]interface{}{"checked": 0, "flipped": 0, "dry_run": dryRun, "results": []interface{}{}}, nil
	}

	today := currentDateString(ctx, params.DB)
	var results []directoryClaimRefresh
	flipped := 0

	for _, c := range due {
		entry := directoryClaimRefresh{EntityID: c.entityID.String(), Slug: c.slug, Field: c.field}

		cit, cerr := datahelpers.ParseCitation(map[string]interface{}{"citation": unmarshalJSONObject(c.citationJSON)})
		if cerr != nil || cit == nil {
			entry.Outcome = "error"
			entry.Detail = "stored citation is malformed"
			results = append(results, entry)
			continue
		}

		outcome := verifyCitationLive(ctx, cit)
		newStatus, label, isTransition := classifyDirectoryClaimOutcome(c.status, outcome)

		if !isTransition {
			entry.Outcome = label // "fresh"
			if !dryRun {
				if _, err := params.DB.ExecContext(ctx,
					`UPDATE directory_claims SET verified_at = now() WHERE id = $1`, c.id); err != nil {
					logger.Warn("refresh_directory_claims: verified_at bump failed",
						zap.String("claim_id", c.id.String()), zap.Error(err))
				}
			}
			results = append(results, entry)
			continue
		}

		entry.Outcome = label // "recovered" or newStatus
		entry.Detail = outcome.FailDetail
		if label != "recovered" {
			flipped++
		}

		if !dryRun {
			citMap := unmarshalJSONObject(c.citationJSON)
			citMap["accessed"] = today
			citJSON, _ := json.Marshal(citMap)
			if err := writeCurrentDirectoryClaim(ctx, params.DB, c.entityID, c.field, c.value, c.unit,
				citJSON, newStatus, c.stalenessDays, "directory-freshness"); err != nil {
				logger.Warn("refresh_directory_claims: status transition write failed",
					zap.String("claim_id", c.id.String()), zap.Error(err))
			}
		}
		results = append(results, entry)
	}

	if !dryRun && flipped > 0 {
		if err := createStaleDirectoryClaimItem(ctx, params, results, logger); err != nil {
			logger.Warn("refresh_directory_claims: failed to create stale_directory_claim item", zap.Error(err))
		}
	}

	logger.Info("refresh_directory_claims: complete",
		zap.Int("checked", len(due)), zap.Int("flipped", flipped), zap.Bool("dry_run", dryRun))
	return map[string]interface{}{
		"checked": len(due), "flipped": flipped, "dry_run": dryRun, "results": results,
	}, nil
}

// classifyDirectoryClaimOutcome decides a claim's next status and refresh
// label from a live citation-verify outcome and its previous status. Pure
// and DB-free, so the fetch_error/citation_lost distinction — the one this
// whole layer exists to get right, per evidence_citations.go's header — is
// testable without a live fetch.
//
// isTransition is false only when the status is literally unchanged (the
// "fresh" no-op path — bump verified_at in place, no version churn).
// "recovered" (found again after any non-found status) is itself still a
// transition (a new current row is written, recording the recovery in
// history), but is never counted as a caller's "flipped away from found".
func classifyDirectoryClaimOutcome(prevStatus string, outcome citationVerifyOutcome) (newStatus, label string, isTransition bool) {
	switch {
	case outcome.Found:
		newStatus = "found"
	case outcome.FailClass == "fetch_error":
		newStatus = "fetch_error"
	default:
		newStatus = "citation_lost"
	}

	if newStatus == prevStatus {
		return newStatus, "fresh", false
	}
	if newStatus == "found" {
		return newStatus, "recovered", true
	}
	return newStatus, newStatus, true
}

// loadDueDirectoryClaims selects every is_current claim whose own
// staleness_days has elapsed since it was last verified — the SAME
// selection role citation_lost detection needs, generalised past a single
// site: only check what is actually due, never the whole table every tick.
func loadDueDirectoryClaims(ctx context.Context, db *sql.DB, kind string) ([]dueDirectoryClaim, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT dc.id, dc.entity_id, de.slug, dc.field, dc.value, dc.unit, dc.citation, dc.status, dc.staleness_days
		FROM directory_claims dc
		JOIN directory_entities de ON de.id = dc.entity_id
		WHERE dc.is_current
		  AND (dc.verified_at IS NULL OR dc.verified_at < now() - (dc.staleness_days::text || ' days')::interval)
		  AND ($1 = '' OR de.kind = $1)
	`, kind)
	if err != nil {
		return nil, fmt.Errorf("select due claims: %w", err)
	}
	defer rows.Close()

	var due []dueDirectoryClaim
	for rows.Next() {
		var c dueDirectoryClaim
		var valueN, unitN sql.NullString
		if err := rows.Scan(&c.id, &c.entityID, &c.slug, &c.field, &valueN, &unitN,
			&c.citationJSON, &c.status, &c.stalenessDays); err != nil {
			return nil, fmt.Errorf("scan due claim: %w", err)
		}
		c.value, c.unit = valueN.String, unitN.String
		due = append(due, c)
	}
	return due, rows.Err()
}

// createStaleDirectoryClaimItem raises drift for human ruling. HITL-terminal,
// matching stale_evidence: a claim's verification status changing is a fact
// about the world, but acting on it (re-research, retire, accept) is a
// human decision.
func createStaleDirectoryClaimItem(
	ctx context.Context, params ActionParams, results []directoryClaimRefresh, logger *zap.Logger,
) error {
	var flipped []directoryClaimRefresh
	for _, r := range results {
		if r.Outcome != "fresh" {
			flipped = append(flipped, r)
		}
	}
	specJSON, err := json.Marshal(map[string]interface{}{
		"check":   "directory_claim_freshness",
		"flipped": flipped,
		"fix": "A directory claim's cited source no longer supports it (citation_lost), could not be " +
			"reached (fetch_error — often transient, self-heals if the source recovers), or recovered " +
			"after a prior failure. Review the affected entity/field: re-research a lost claim, leave a " +
			"fetch_error to retry, or note a recovery.",
	})
	if err != nil {
		return err
	}
	siteID, err := uuid.Parse(directorySystemSiteID)
	if err != nil {
		return err
	}
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "scheduled",
		pipeline:     "content",
		itemType:     "stale_directory_claim",
		severity:     "medium",
		summary:      fmt.Sprintf("Model directory freshness: %d claim(s) changed verification status", len(flipped)),
		spec:         string(specJSON),
		priority:     35,
		handlerAgent: "human-review",
		status:       "needs_human_review",
		createdBy:    params.AgentType,
		itemKey:      "stale_directory_claim",
	}, logger)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// unmarshalJSONObject decodes a stored jsonb citation. A malformed row
// (should never happen — this package writes json.Marshal output) degrades
// to an empty map rather than panicking; the caller's ParseCitation then
// reports it as invalid, which is the correct visible failure.
func unmarshalJSONObject(raw []byte) map[string]interface{} {
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return map[string]interface{}{}
	}
	return m
}
