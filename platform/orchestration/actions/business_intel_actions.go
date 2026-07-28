// FILE: platform/orchestration/actions/business_intel_actions.go
// Actions for the business intelligence data collection pipeline.
// These work against the business_intel schema in clients_db.
//
// Actions:
//   - load_business_record:        Load a single business + vertical-specific details
//   - store_business_verification: Store verification results from an agent run
//   - load_business_batch:         Load next batch of pending collection tasks

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// scrapedDataField is the collected_data key holding the webscrape adapter's
// response — the fetch record that carries the URL actually retrieved and when.
// It is the provenance source for store_business_verification (bugs_open/100).
const scrapedDataField = "scraped_data"

// ---------------------------------------------------------------------------
// load_business_record
// ---------------------------------------------------------------------------
// Loads a business from business_intel.businesses by ID, with optional
// vertical-specific details (e.g. vet_practice_details).
//
// Workflow config:
//
//	{
//	    "action": "load_business_record",
//	    "config": {
//	        "input_fields": ["business_id"],
//	        "include_vet_details": true,
//	        "include_prices": true
//	    },
//	    "output_field": "business_record"
//	}
func LoadBusinessRecordAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("LoadBusinessRecordAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config

	// Extract business_id from collected data
	inputFields := []string{"business_id"}
	if fields, ok := config["input_fields"].([]interface{}); ok {
		inputFields = make([]string, len(fields))
		for i, f := range fields {
			inputFields[i], _ = f.(string)
		}
	}
	extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

	businessID, _ := extracted["business_id"].(string)
	if businessID == "" {
		return nil, fmt.Errorf("business_id is required")
	}

	params.Logger.Info("LoadBusinessRecordAction: Loading business",
		zap.String("business_id", businessID),
	)

	// Load core business record
	business, err := loadBusiness(ctx, params.DB, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to load business %s: %w", businessID, err)
	}

	result := map[string]interface{}{
		"business":    business,
		"business_id": businessID,
		"loaded_at":   time.Now().UTC().Format(time.RFC3339),
	}

	// Optionally load vet-specific details
	includeVet, _ := config["include_vet_details"].(bool)
	if includeVet {
		vetDetails, err := loadVetDetails(ctx, params.DB, businessID)
		if err != nil {
			params.Logger.Warn("LoadBusinessRecordAction: No vet details found",
				zap.String("business_id", businessID),
				zap.Error(err),
			)
		} else if vetDetails != nil {
			result["vet_details"] = vetDetails
		}
	}

	// Optionally load current prices
	includePrices, _ := config["include_prices"].(bool)
	if includePrices {
		prices, err := loadCurrentPrices(ctx, params.DB, businessID)
		if err != nil {
			params.Logger.Warn("LoadBusinessRecordAction: Failed to load prices",
				zap.Error(err),
			)
		} else {
			result["prices"] = prices
		}
	}

	params.Logger.Info("LoadBusinessRecordAction: Loaded",
		zap.String("business_id", businessID),
		zap.String("name", stringFromMap(business, "name")),
	)

	return result, nil
}

// ---------------------------------------------------------------------------
// store_business_verification
// ---------------------------------------------------------------------------
// Stores results from a verification agent run. Updates the business record,
// vet details, prices, and creates a data_observation entry.
//
// Workflow config:
//
//	{
//	    "action": "store_business_verification",
//	    "config": {
//	        "input_fields": ["business_id", "verification_result"]
//	    },
//	    "output_field": "store_result"
//	}
//
// Expected verification_result shape:
//
//	{
//	    "business": { "name", "address_line1", "town", "postcode", "phone", "email", "website_url", ... },
//	    "vet_details": { "species_treated", "accepting_new_clients", ... },
//	    "prices": [ { "service_category", "service_name", "price_gbp", ... } ],
//	    "confidence_score": 0.85,
//	    "extraction_notes": "..."
//	}
//
// NOTE — provenance is NOT in that shape, deliberately (bugs_open/100).
// This comment used to list "source_type" / "source_name" / "source_url" as
// fields of verification_result, and that is what the code read. The prompt has
// never asked for them, so they were never present and all 2,970 observations
// were stored unsourced. Adding them to the prompt is the WRONG repair: it makes
// provenance a model claim about its own evidence, generated by the same call
// that generated the facts, with nothing to check it against.
//
// Provenance is taken from `scraped_data` — the fetch record written by the
// component that performed the fetch — via datahelpers.ExtractFetchProvenance.
// If you are here to make the model emit a source_url, read
// bugs_open/100 §"Why the obvious fix is WRONG" first.
func StoreBusinessVerificationAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("StoreBusinessVerificationAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config

	// Extract inputs.
	//
	// scraped_data is appended unconditionally rather than left to the definition's
	// input_fields list (bugs_open/100). Provenance is not optional, and making it
	// depend on every caller remembering a config key is what produced 2,970 rows
	// with no source: the writer must be able to reach the fetch record whatever the
	// definition says.
	inputFields := []string{"business_id", "verification_result"}
	if fields, ok := config["input_fields"].([]interface{}); ok {
		inputFields = make([]string, len(fields))
		for i, f := range fields {
			inputFields[i], _ = f.(string)
		}
	}
	if !containsString(inputFields, scrapedDataField) {
		inputFields = append(inputFields, scrapedDataField)
	}
	extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

	businessID, _ := extracted["business_id"].(string)
	if businessID == "" {
		return nil, fmt.Errorf("business_id is required")
	}

	taskID, _ := extracted["task_id"].(string)

	verResult, ok := extracted["verification_result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("verification_result must be an object")
	}

	// Provenance comes from the fetch, and ONLY from the fetch (bugs_open/100).
	//
	// This used to read verResult["source_url"] / ["source_type"] / ["source_name"]
	// — the model's own output object. The prompt never asks for those keys, so they
	// were never present and every observation was stored unsourced; and had they
	// been present it would have been worse, because provenance asserted by the same
	// call that produced the facts is not provenance. Those reads are gone rather
	// than demoted to a fallback: a fallback would have quietly restored the old
	// behaviour the moment a model volunteered a plausible-looking URL.
	prov, provOK := datahelpers.ExtractFetchProvenance(extracted[scrapedDataField])
	if !provOK {
		params.Logger.Warn("StoreBusinessVerificationAction: no fetch provenance available — observation will be stored unsourced",
			zap.String("business_id", businessID),
			zap.String("expected_field", scrapedDataField),
			zap.String("consequence", "the row cannot say where it came from and is unpublishable under the sourcing rule"),
			zap.String("ref", "bugs_open/100"),
		)
	}
	if llmClaimed := verResult["source_url"]; llmClaimed != nil {
		// Not used — recorded because it means the prompt has started asking the
		// model for its own provenance, which is the rejected candidate 4.
		params.Logger.Warn("StoreBusinessVerificationAction: model emitted a source_url; it is being IGNORED",
			zap.String("business_id", businessID),
			zap.Any("model_claimed", llmClaimed),
			zap.String("ref", "bugs_open/100 §Why the obvious fix is WRONG"),
		)
	}

	params.Logger.Info("StoreBusinessVerificationAction: Storing results",
		zap.String("business_id", businessID),
	)

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Update core business fields
	updatedFields := 0
	if bizData, ok := verResult["business"].(map[string]interface{}); ok {
		n, err := updateBusinessFields(ctx, tx, businessID, bizData)
		if err != nil {
			return nil, fmt.Errorf("failed to update business: %w", err)
		}
		updatedFields += n
	}

	// 1b. Store individual contact details
	contactsStored := 0
	if bizData, ok := verResult["business"].(map[string]interface{}); ok {
		contactsStored = storeContactDetails(ctx, tx, businessID, bizData,
			params.ExecutionContext.OrchestrationID, params.Logger)
		params.Logger.Info("StoreBusinessVerificationAction: Stored contacts",
			zap.Int("contacts_stored", contactsStored))
	}

	// 1c. Confirm/miss existing contacts based on this extraction
	if bizData, ok := verResult["business"].(map[string]interface{}); ok {
		confirmContacts(ctx, tx, businessID, bizData, params.Logger)
	}

	// Update verification status and confidence
	confidenceScore := 0.0
	if cs, ok := verResult["confidence_score"].(float64); ok {
		confidenceScore = cs
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE business_intel.businesses 
		SET verification_status = 'verified',
		    confidence_score = $2,
		    last_verified_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1`,
		businessID, confidenceScore,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update verification status: %w", err)
	}

	// 2. Upsert vet practice details
	vetUpdated := false
	if vetData, ok := verResult["vet_details"].(map[string]interface{}); ok {
		err := upsertVetDetails(ctx, tx, businessID, vetData)
		if err != nil {
			params.Logger.Warn("StoreBusinessVerificationAction: Failed to upsert vet details",
				zap.Error(err),
			)
		} else {
			vetUpdated = true
		}
	}

	// 3. Store service prices (mark old as not current, insert new).
	// Writes go to the unified products (kind='service') + product_prices
	// schema; business_prices is deprecated (006_unify_prices_schema.sql).
	pricesStored := 0
	if pricesRaw, ok := verResult["prices"].([]interface{}); ok && len(pricesRaw) > 0 {
		// This verification supersedes all prior service prices for the business.
		_, _ = tx.ExecContext(ctx, `
			UPDATE business_intel.product_prices pp
			SET is_current = FALSE
			FROM business_intel.products p
			WHERE pp.product_id = p.id
			  AND p.kind = 'service'
			  AND pp.business_id = $1
			  AND pp.is_current = TRUE`,
			businessID,
		)

		for _, priceRaw := range pricesRaw {
			price, ok := priceRaw.(map[string]interface{})
			if !ok {
				continue
			}
			err := insertPrice(ctx, tx, businessID, price, prov.SourceType, prov.SourceURL)
			if err != nil {
				params.Logger.Warn("StoreBusinessVerificationAction: Failed to insert price",
					zap.Error(err),
				)
				continue
			}
			pricesStored++
		}
	}

	// 3b. Store per-practice medicine prices, if the verifier extracted any
	// (verResult.medicine_prices[]). Same unified schema, kind='medicine'.
	medicinesStored := 0
	if medsRaw, ok := verResult["medicine_prices"].([]interface{}); ok && len(medsRaw) > 0 {
		_, _ = tx.ExecContext(ctx, `
			UPDATE business_intel.product_prices pp
			SET is_current = FALSE
			FROM business_intel.products p
			WHERE pp.product_id = p.id
			  AND p.kind = 'medicine'
			  AND pp.business_id = $1
			  AND pp.is_current = TRUE`,
			businessID,
		)

		for _, medRaw := range medsRaw {
			med, ok := medRaw.(map[string]interface{})
			if !ok {
				continue
			}
			err := insertMedicinePrice(ctx, tx, businessID, med, prov.SourceType, prov.SourceURL)
			if err != nil {
				params.Logger.Warn("StoreBusinessVerificationAction: Failed to insert medicine price",
					zap.Error(err),
				)
				continue
			}
			medicinesStored++
		}
	}

	// 4. Create data observation (provenance record)
	rawDataJSON, _ := json.Marshal(verResult)
	sourceType, sourceName, sourceURL := prov.SourceType, prov.SourceName, prov.SourceURL
	extractionNotes, _ := verResult["extraction_notes"].(string)

	// collected_at is the FETCH time as recorded by the fetcher, not the write time.
	// Passing NULL lets the column default to now(); an observation retrieved hours
	// before it was written should not claim to be as fresh as the write.
	var collectedAt interface{}
	if prov.CapturedAt != "" {
		collectedAt = prov.CapturedAt
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO business_intel.data_observations
			(business_id, source_type, source_name, source_url, raw_data,
			 extraction_confidence, extraction_notes, orchestration_id,
			 collected_at, processed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, COALESCE($9::timestamp, NOW()), NOW())`,
		businessID, sourceType, sourceName, sourceURL, rawDataJSON,
		confidenceScore, extractionNotes, params.ExecutionContext.OrchestrationID,
		collectedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert data observation: %w", err)
	}

	// 4b. Cache search results for discovery mining
	cacheSearchResults(ctx, tx, businessID, params.CollectedData,
		params.ExecutionContext.OrchestrationID, params.Logger)

	// 4c. Bump verification count
	if err := bumpVerificationCount(ctx, tx, businessID); err != nil {
		params.Logger.Warn("Failed to bump verification count", zap.Error(err))
	}

	// 4d. Company number extraction — regex fallback if LLM didn't find it.
	// The LLM may extract registration_number via the prompt, stored by
	// updateBusinessFields. If not, try regex on the scraped content.
	companyNumberFound := ""
	if bizData, ok := verResult["business"].(map[string]interface{}); ok {
		if regNum, ok := bizData["registration_number"].(string); ok && regNum != "" {
			companyNumberFound = regNum
		}
	}

	if companyNumberFound == "" {
		// Try regex on scraped content from collected data
		companyNumberFound = extractCompanyNumberFromCollectedData(params.CollectedData, params.Logger)
		if companyNumberFound != "" {
			// Store the regex-extracted number
			_, _ = tx.ExecContext(ctx, `
				UPDATE business_intel.businesses
				SET company_number_scraped = $1, updated_at = NOW()
				WHERE id = $2 AND (company_number_scraped IS NULL OR company_number_scraped = '')`,
				companyNumberFound, businessID)
			params.Logger.Info("StoreBusinessVerification: extracted company number via regex",
				zap.String("business_id", businessID),
				zap.String("company_number", companyNumberFound))
		}
	}

	// 5. Update collection task - prefer task_id, fall back to orchestration_id
	if taskID != "" {
		_, _ = tx.ExecContext(ctx, `
			UPDATE business_intel.collection_tasks
			SET status = 'completed',
			    completed_at = NOW(),
			    result_summary = $2
			WHERE id = $1 AND status = 'in_progress'`,
			taskID,
			rawDataJSON,
		)
	} else {
		_, _ = tx.ExecContext(ctx, `
			UPDATE business_intel.collection_tasks
			SET status = 'completed',
			    completed_at = NOW(),
			    result_summary = $2
			WHERE orchestration_id = $1 AND status = 'in_progress'`,
			params.ExecutionContext.OrchestrationID,
			rawDataJSON,
		)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	// Post-commit: attempt CH match if we have a company number.
	// Best-effort — doesn't affect the verification result.
	chMatched := false
	if companyNumberFound != "" {
		matchResult, err := params.DB.ExecContext(ctx, `
			UPDATE business_intel.ch_vet_companies
			SET matched_business_id = $1,
				matched_at = NOW(),
				match_confidence = 1.0,
				match_method = 'company_number_scraped',
				updated_at = NOW()
			WHERE company_number = $2
			  AND company_status = 'active'
			  AND matched_business_id IS NULL`,
			businessID, companyNumberFound)
		if err == nil {
			if rows, _ := matchResult.RowsAffected(); rows > 0 {
				chMatched = true
				params.Logger.Info("StoreBusinessVerification: CH matched by company number",
					zap.String("business_id", businessID),
					zap.String("company_number", companyNumberFound))
			}
		}
	}

	result := map[string]interface{}{
		"stored":                 true,
		"business_id":            businessID,
		"updated_fields":         updatedFields,
		"vet_updated":            vetUpdated,
		"prices_stored":          pricesStored,
		"medicine_prices_stored": medicinesStored,
		"contacts_stored":        contactsStored,
		"search_cached":          true,
		"company_number_found":   companyNumberFound,
		"ch_matched":             chMatched,
		"stored_at":              time.Now().UTC().Format(time.RFC3339),
	}

	params.Logger.Info("StoreBusinessVerificationAction: Stored",
		zap.String("business_id", businessID),
		zap.Int("updated_fields", updatedFields),
		zap.Int("prices_stored", pricesStored),
	)

	return result, nil
}

// ---------------------------------------------------------------------------
// load_business_batch
// ---------------------------------------------------------------------------
// Loads next batch of pending collection tasks for processing.
// Claims them by setting status to 'in_progress'.
//
// Workflow config:
//
//	{
//	    "action": "load_business_batch",
//	    "config": {
//	        "batch_size": 10,
//	        "task_type": "initial_verification",
//	        "vertical_slug": "veterinary"
//	    },
//	    "output_field": "batch"
//	}
func LoadBusinessBatchAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("LoadBusinessBatchAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config

	// Config parameters with defaults
	batchSize := 10
	if bs, ok := config["batch_size"].(float64); ok && bs > 0 {
		batchSize = int(bs)
	}
	// Override from input_data if provided (e.g. pipeline passes verify_limit as batch_size)
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if bs, ok := inputData["batch_size"].(float64); ok && bs > 0 {
			batchSize = int(bs)
		}
	}

	taskType, _ := config["task_type"].(string) // optional filter
	verticalSlug, _ := config["vertical_slug"].(string)

	params.Logger.Info("LoadBusinessBatchAction: Loading batch",
		zap.Int("batch_size", batchSize),
		zap.String("task_type", taskType),
		zap.String("vertical_slug", verticalSlug),
	)

	// Claim a batch of pending tasks atomically
	// Uses SELECT ... FOR UPDATE SKIP LOCKED for safe concurrent access
	// (though for now it's single-pod sequential)
	query := `
		WITH claimed AS (
			SELECT ct.id
			FROM business_intel.collection_tasks ct
			WHERE ct.status = 'pending'
			  AND (ct.scheduled_for IS NULL OR ct.scheduled_for <= NOW())
	`
	args := []interface{}{}
	argIdx := 1

	if taskType != "" {
		query += fmt.Sprintf(" AND ct.task_type = $%d", argIdx)
		args = append(args, taskType)
		argIdx++
	}

	if verticalSlug != "" {
		query += fmt.Sprintf(`
			AND ct.vertical_id = (
				SELECT id FROM business_intel.business_verticals WHERE slug = $%d
			)`, argIdx)
		args = append(args, verticalSlug)
		argIdx++
	}

	query += fmt.Sprintf(`
			ORDER BY ct.priority ASC, ct.scheduled_for ASC NULLS LAST
			LIMIT $%d
			FOR UPDATE SKIP LOCKED
		)
		UPDATE business_intel.collection_tasks 
		SET status = 'in_progress', 
		    started_at = NOW(),
		    orchestration_id = $%d::uuid
		FROM claimed
		WHERE business_intel.collection_tasks.id = claimed.id
		RETURNING business_intel.collection_tasks.id, 
		          business_intel.collection_tasks.business_id, 
		          business_intel.collection_tasks.task_type, 
		          business_intel.collection_tasks.priority`,
		argIdx, argIdx+1)
	args = append(args, batchSize, params.ExecutionContext.OrchestrationID)

	rows, err := params.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to claim batch: %w", err)
	}
	defer rows.Close()

	type claimedTask struct {
		TaskID     string `json:"task_id"`
		BusinessID string `json:"business_id"`
		TaskType   string `json:"task_type"`
		Priority   int    `json:"priority"`
	}

	var tasks []claimedTask
	for rows.Next() {
		var t claimedTask
		if err := rows.Scan(&t.TaskID, &t.BusinessID, &t.TaskType, &t.Priority); err != nil {
			params.Logger.Warn("LoadBusinessBatchAction: Failed to scan row", zap.Error(err))
			continue
		}
		tasks = append(tasks, t)
	}

	// For each claimed task, load the business summary
	items := make([]map[string]interface{}, 0, len(tasks))
	for _, t := range tasks {
		biz, err := loadBusinessSummary(ctx, params.DB, t.BusinessID)
		if err != nil {
			params.Logger.Warn("LoadBusinessBatchAction: Failed to load business",
				zap.String("business_id", t.BusinessID),
				zap.Error(err),
			)
			continue
		}
		items = append(items, map[string]interface{}{
			"task_id":     t.TaskID,
			"business_id": t.BusinessID,
			"task_type":   t.TaskType,
			"priority":    t.Priority,
			"business":    biz,
		})
	}

	result := map[string]interface{}{
		"items":      items,
		"batch_size": len(items),
		"has_more":   len(items) == batchSize, // if we got a full batch, there may be more
		"loaded_at":  time.Now().UTC().Format(time.RFC3339),
	}

	params.Logger.Info("LoadBusinessBatchAction: Batch loaded",
		zap.Int("claimed", len(items)),
	)

	return result, nil
}

// ---------------------------------------------------------------------------
// DB helper functions
// ---------------------------------------------------------------------------

func loadBusiness(ctx context.Context, db *sql.DB, businessID string) (map[string]interface{}, error) {
	row := db.QueryRowContext(ctx, `
		SELECT b.id, b.name, b.slug, b.trading_name,
		       b.address_line1, b.address_line2, b.town, b.county, b.postcode, b.country,
		       b.latitude, b.longitude,
		       b.phone, b.email, b.website_url,
		       b.business_type, b.group_name, b.is_independent,
		       b.verification_status, b.confidence_score, b.last_verified_at,
		       b.is_active,
		       bv.slug as vertical_slug, bv.display_name as vertical_name
		FROM business_intel.businesses b
		LEFT JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
		WHERE b.id = $1`, businessID)

	var (
		id, name                                      string
		slug, tradingName                             sql.NullString
		addr1, addr2, town, county, postcode, country sql.NullString
		lat, lng                                      sql.NullFloat64
		phone, email, websiteURL                      sql.NullString
		bizType, groupName                            sql.NullString
		isIndependent                                 sql.NullBool
		verStatus                                     string
		confidence                                    sql.NullFloat64
		lastVerified                                  sql.NullTime
		isActive                                      bool
		verticalSlug, verticalName                    sql.NullString
	)

	err := row.Scan(
		&id, &name, &slug, &tradingName,
		&addr1, &addr2, &town, &county, &postcode, &country,
		&lat, &lng,
		&phone, &email, &websiteURL,
		&bizType, &groupName, &isIndependent,
		&verStatus, &confidence, &lastVerified,
		&isActive,
		&verticalSlug, &verticalName,
	)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":                  id,
		"name":                name,
		"verification_status": verStatus,
		"is_active":           isActive,
	}

	// Only include non-null fields
	setIfValid(result, "slug", slug)
	setIfValid(result, "trading_name", tradingName)
	setIfValid(result, "address_line1", addr1)
	setIfValid(result, "address_line2", addr2)
	setIfValid(result, "town", town)
	setIfValid(result, "county", county)
	setIfValid(result, "postcode", postcode)
	setIfValid(result, "country", country)
	setIfValidFloat(result, "latitude", lat)
	setIfValidFloat(result, "longitude", lng)
	setIfValid(result, "phone", phone)
	setIfValid(result, "email", email)
	setIfValid(result, "website_url", websiteURL)
	setIfValid(result, "business_type", bizType)
	setIfValid(result, "group_name", groupName)
	setIfValidBool(result, "is_independent", isIndependent)
	setIfValidFloat(result, "confidence_score", confidence)
	if lastVerified.Valid {
		result["last_verified_at"] = lastVerified.Time.Format(time.RFC3339)
	}
	setIfValid(result, "vertical_slug", verticalSlug)
	setIfValid(result, "vertical_name", verticalName)

	return result, nil
}

func loadBusinessSummary(ctx context.Context, db *sql.DB, businessID string) (map[string]interface{}, error) {
	row := db.QueryRowContext(ctx, `
		SELECT b.id, b.name, b.postcode, b.town, b.website_url, 
		       b.verification_status, b.phone
		FROM business_intel.businesses b
		WHERE b.id = $1`, businessID)

	var (
		id, name, verStatus     string
		postcode, town, website sql.NullString
		phone                   sql.NullString
	)

	err := row.Scan(&id, &name, &postcode, &town, &website, &verStatus, &phone)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":                  id,
		"name":                name,
		"verification_status": verStatus,
	}
	setIfValid(result, "postcode", postcode)
	setIfValid(result, "town", town)
	setIfValid(result, "website_url", website)
	setIfValid(result, "phone", phone)

	return result, nil
}

func loadVetDetails(ctx context.Context, db *sql.DB, businessID string) (map[string]interface{}, error) {
	row := db.QueryRowContext(ctx, `
		SELECT species_treated, emergency_service, out_of_hours_provider,
		       accepting_new_clients, accepting_new_clients_updated_at,
		       accreditations, num_vets, num_nurses, head_vet_name,
		       has_own_lab, has_imaging, has_surgical_suite,
		       parking_available, wheelchair_accessible
		FROM business_intel.vet_practice_details
		WHERE business_id = $1`, businessID)

	var (
		speciesTreated, accreditations                                  []byte // TEXT[] scanned as bytes
		emergencyService, acceptingNew                                  sql.NullBool
		oohProvider, headVet                                            sql.NullString
		acceptingNewUpdated                                             sql.NullTime
		numVets, numNurses                                              sql.NullInt32
		hasLab, hasImaging, hasSurgical, parkingAvail, wheelchairAccess sql.NullBool
	)

	err := row.Scan(
		&speciesTreated, &emergencyService, &oohProvider,
		&acceptingNew, &acceptingNewUpdated,
		&accreditations, &numVets, &numNurses, &headVet,
		&hasLab, &hasImaging, &hasSurgical,
		&parkingAvail, &wheelchairAccess,
	)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{}
	if speciesTreated != nil {
		result["species_treated"] = pgArrayToSlice(string(speciesTreated))
	}
	setIfValidBool(result, "emergency_service", emergencyService)
	setIfValid(result, "out_of_hours_provider", oohProvider)
	setIfValidBool(result, "accepting_new_clients", acceptingNew)
	if acceptingNewUpdated.Valid {
		result["accepting_new_clients_updated_at"] = acceptingNewUpdated.Time.Format(time.RFC3339)
	}
	if accreditations != nil {
		result["accreditations"] = pgArrayToSlice(string(accreditations))
	}
	if numVets.Valid {
		result["num_vets"] = numVets.Int32
	}
	if numNurses.Valid {
		result["num_nurses"] = numNurses.Int32
	}
	setIfValid(result, "head_vet_name", headVet)
	setIfValidBool(result, "has_own_lab", hasLab)
	setIfValidBool(result, "has_imaging", hasImaging)
	setIfValidBool(result, "has_surgical_suite", hasSurgical)
	setIfValidBool(result, "parking_available", parkingAvail)
	setIfValidBool(result, "wheelchair_accessible", wheelchairAccess)

	return result, nil
}

func loadCurrentPrices(ctx context.Context, db *sql.DB, businessID string) ([]map[string]interface{}, error) {
	// Reads the unified schema: products (kind='service') + product_prices.
	// Output shape is unchanged for callers (service_category/service_name).
	rows, err := db.QueryContext(ctx, `
		SELECT p.category AS service_category, p.name AS service_name,
		       pp.price_gbp, pp.price_qualifier,
		       pp.includes_vat, pp.source, pp.product_url AS source_url, pp.observed_at
		FROM business_intel.product_prices pp
		JOIN business_intel.products p ON p.id = pp.product_id
		WHERE pp.business_id = $1 AND pp.is_current = TRUE AND p.kind = 'service'
		ORDER BY p.category, p.name`, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []map[string]interface{}
	for rows.Next() {
		var (
			category, name string
			price          sql.NullFloat64
			qualifier      sql.NullString
			inclVAT        sql.NullBool
			source         string
			sourceURL      sql.NullString
			observedAt     time.Time
		)
		if err := rows.Scan(&category, &name, &price, &qualifier, &inclVAT, &source, &sourceURL, &observedAt); err != nil {
			continue
		}
		p := map[string]interface{}{
			"service_category": category,
			"service_name":     name,
			"source":           source,
			"observed_at":      observedAt.Format(time.RFC3339),
		}
		if price.Valid {
			p["price_gbp"] = price.Float64
		}
		setIfValid(p, "price_qualifier", qualifier)
		setIfValidBool(p, "includes_vat", inclVAT)
		setIfValid(p, "source_url", sourceURL)
		prices = append(prices, p)
	}

	if prices == nil {
		prices = []map[string]interface{}{}
	}
	return prices, nil
}

// Problem: LLM can return arrays for text fields (e.g. phone: ["028 9047 1361", "028 9065 1729"]),
// numbers as strings, booleans for text fields, etc. The pg driver can't encode []interface{}
// into a text column.
//
// Fix: Add coerceToString helper and use it for all text fields in the allowed list.
// Numeric fields (latitude, longitude) get coerced to float64.

// coerceToString converts an interface{} value to a string suitable for a TEXT column.
// Handles: string (passthrough), []interface{} (join with ", "), []string (join),
// float64/int (fmt), bool (fmt), nil (returns nil to skip).
func coerceToString(val interface{}) interface{} {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case string:
		if v == "" {
			return nil
		}
		return v
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				return s
			}
		}
		return nil
	case []string:
		if len(v) == 0 {
			return nil
		}
		return v[0]
	case float64:
		return fmt.Sprintf("%g", v)
	case int:
		return fmt.Sprintf("%d", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ============================================================================
// Helper: coerceToFloat64
// ============================================================================
func coerceToFloat64(val interface{}) interface{} {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
		return nil
	default:
		return nil
	}
}

// ============================================================================
// Replace existing updateBusinessFields
// ============================================================================
func updateBusinessFields(ctx context.Context, tx *sql.Tx, businessID string, data map[string]interface{}) (int, error) {
	allowedFields := map[string]string{
		"name":                "name",
		"trading_name":        "trading_name",
		"address_line1":       "address_line1",
		"address_line2":       "address_line2",
		"town":                "town",
		"county":              "county",
		"postcode":            "postcode",
		"phone":               "phone",
		"email":               "email",
		"website_url":         "website_url",
		"business_type":       "business_type",
		"group_name":          "group_name",
		"latitude":            "latitude",
		"longitude":           "longitude",
		"registration_number": "company_number_scraped",
	}

	numericFields := map[string]bool{
		"latitude":  true,
		"longitude": true,
	}

	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	for jsonField, dbCol := range allowedFields {
		rawVal, ok := data[jsonField]
		if !ok || rawVal == nil {
			continue
		}

		var coerced interface{}
		if numericFields[jsonField] {
			coerced = coerceToFloat64(rawVal)
		} else {
			coerced = coerceToString(rawVal)
		}

		if coerced == nil {
			continue
		}

		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", dbCol, argIdx))
		args = append(args, coerced)
		argIdx++
	}

	if len(setClauses) == 0 {
		return 0, nil
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	query := fmt.Sprintf("UPDATE business_intel.businesses SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "), argIdx)
	args = append(args, businessID)

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	return len(setClauses) - 1, nil // -1 for updated_at
}

// ============================================================================
// New: storeContactDetails
// ============================================================================
// Extracts phone numbers, emails, and other contact info from the LLM result
// and stores each individually in business_contact_details.
// Uses ON CONFLICT to upsert existing entries.
func storeContactDetails(ctx context.Context, tx *sql.Tx, businessID string,
	bizData map[string]interface{}, orchestrationID string, logger *zap.Logger) int {

	stored := 0
	primaryByType := map[string]bool{} // track which types already have a primary

	insertContact := func(contactType, label, value, source string) {
		if value == "" {
			return
		}
		isPrimary := !primaryByType[contactType]
		_, err := tx.ExecContext(ctx, `
			INSERT INTO business_intel.business_contact_details 
				(business_id, contact_type, label, value, source, orchestration_id, 
				 is_primary, first_seen_at, last_confirmed_at, check_count, missed_count, is_stale, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW(), 1, 0, FALSE, NOW())
			ON CONFLICT (business_id, contact_type, value) 
			DO UPDATE SET 
				label = COALESCE(NULLIF(EXCLUDED.label, ''), business_intel.business_contact_details.label),
				source = EXCLUDED.source,
				orchestration_id = EXCLUDED.orchestration_id,
				last_confirmed_at = NOW(),
				check_count = business_intel.business_contact_details.check_count + 1,
				missed_count = 0,
				is_stale = FALSE,
				updated_at = NOW()`,
			businessID, contactType, label, value, source, orchestrationID, isPrimary,
		)
		if err != nil {
			logger.Warn("storeContactDetails: failed to insert",
				zap.String("type", contactType),
				zap.String("value", value),
				zap.Error(err))
			return
		}
		primaryByType[contactType] = true
		stored++
	}

	// Helper to extract string or string array values
	extractAndStore := func(field interface{}, contactType string) {
		if field == nil {
			return
		}
		switch v := field.(type) {
		case string:
			if v != "" {
				insertContact(contactType, "main", v, "llm_extraction")
			}
		case []interface{}:
			for i, item := range v {
				if s, ok := item.(string); ok && s != "" {
					label := "main"
					if i > 0 {
						label = fmt.Sprintf("additional_%d", i)
					}
					insertContact(contactType, label, s, "llm_extraction")
				}
			}
		case []string:
			for i, s := range v {
				if s != "" {
					label := "main"
					if i > 0 {
						label = fmt.Sprintf("additional_%d", i)
					}
					insertContact(contactType, label, s, "llm_extraction")
				}
			}
		}
	}

	extractAndStore(bizData["phone"], "phone")
	extractAndStore(bizData["email"], "email")
	extractAndStore(bizData["fax"], "fax")

	// Website as a contact entry too
	if ws, ok := bizData["website_url"].(string); ok && ws != "" {
		insertContact("website", "main", ws, "llm_extraction")
	}

	return stored
}

func upsertVetDetails(ctx context.Context, tx *sql.Tx, businessID string, data map[string]interface{}) error {
	// Use INSERT ON CONFLICT for upsert
	_, err := tx.ExecContext(ctx, `
		INSERT INTO business_intel.vet_practice_details (
			business_id, species_treated, emergency_service, out_of_hours_provider,
			accepting_new_clients, accepting_new_clients_updated_at,
			accreditations, num_vets, num_nurses, head_vet_name,
			has_own_lab, has_imaging, has_surgical_suite,
			parking_available, wheelchair_accessible, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW()
		) ON CONFLICT (business_id) DO UPDATE SET
			species_treated = COALESCE(EXCLUDED.species_treated, business_intel.vet_practice_details.species_treated),
			emergency_service = COALESCE(EXCLUDED.emergency_service, business_intel.vet_practice_details.emergency_service),
			out_of_hours_provider = COALESCE(EXCLUDED.out_of_hours_provider, business_intel.vet_practice_details.out_of_hours_provider),
			accepting_new_clients = COALESCE(EXCLUDED.accepting_new_clients, business_intel.vet_practice_details.accepting_new_clients),
			accepting_new_clients_updated_at = CASE 
				WHEN EXCLUDED.accepting_new_clients IS NOT NULL THEN NOW() 
				ELSE business_intel.vet_practice_details.accepting_new_clients_updated_at 
			END,
			accreditations = COALESCE(EXCLUDED.accreditations, business_intel.vet_practice_details.accreditations),
			num_vets = COALESCE(EXCLUDED.num_vets, business_intel.vet_practice_details.num_vets),
			num_nurses = COALESCE(EXCLUDED.num_nurses, business_intel.vet_practice_details.num_nurses),
			head_vet_name = COALESCE(EXCLUDED.head_vet_name, business_intel.vet_practice_details.head_vet_name),
			has_own_lab = COALESCE(EXCLUDED.has_own_lab, business_intel.vet_practice_details.has_own_lab),
			has_imaging = COALESCE(EXCLUDED.has_imaging, business_intel.vet_practice_details.has_imaging),
			has_surgical_suite = COALESCE(EXCLUDED.has_surgical_suite, business_intel.vet_practice_details.has_surgical_suite),
			parking_available = COALESCE(EXCLUDED.parking_available, business_intel.vet_practice_details.parking_available),
			wheelchair_accessible = COALESCE(EXCLUDED.wheelchair_accessible, business_intel.vet_practice_details.wheelchair_accessible),
			updated_at = NOW()`,
		businessID,
		pgArrayFromInterface(data["species_treated"]),
		nullBoolFromInterface(data["emergency_service"]),
		nullStringFromInterface(data["out_of_hours_provider"]),
		nullBoolFromInterface(data["accepting_new_clients"]),
		nullTimeIfBoolPresent(data["accepting_new_clients"]),
		pgArrayFromInterface(data["accreditations"]),
		nullIntFromInterface(data["num_vets"]),
		nullIntFromInterface(data["num_nurses"]),
		nullStringFromInterface(data["head_vet_name"]),
		nullBoolFromInterface(data["has_own_lab"]),
		nullBoolFromInterface(data["has_imaging"]),
		nullBoolFromInterface(data["has_surgical_suite"]),
		nullBoolFromInterface(data["parking_available"]),
		nullBoolFromInterface(data["wheelchair_accessible"]),
	)
	return err
}

// offeringSlugPattern matches 006_unify_prices_schema.sql exactly:
// regexp_replace(lower(...), '[^a-z0-9]+', '-', 'g'). The Go and SQL
// computations MUST stay identical so migrated and live-written rows
// dedupe onto the same products row via ON CONFLICT (slug).
var offeringSlugPattern = regexp.MustCompile(`[^a-z0-9]+`)

func offeringSlug(parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return offeringSlugPattern.ReplaceAllString(strings.ToLower(strings.Join(nonEmpty, "-")), "-")
}

// upsertOfferingPrice writes one price observation against the unified
// schema: upsert the canonical products row (by slug), retire prior current
// observations for this (business, product), insert the new observation.
func upsertOfferingPrice(ctx context.Context, tx *sql.Tx,
	kind, slug, name, category, dosage, businessID string,
	requiresRx bool, priceGBP interface{}, qualifier string, inclVAT bool,
	inStock interface{}, source, sourceURL string) error {

	if source == "" {
		source = "verifier"
	}

	var productID string
	err := tx.QueryRowContext(ctx, `
		INSERT INTO business_intel.products
			(slug, name, category, dosage, kind, requires_prescription, is_active)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5, $6, true)
		ON CONFLICT (slug) DO UPDATE SET updated_at = NOW()
		RETURNING id`,
		slug, name, category, dosage, kind, requiresRx,
	).Scan(&productID)
	if err != nil {
		return fmt.Errorf("upsert product %q: %w", slug, err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE business_intel.product_prices
		SET is_current = FALSE
		WHERE business_id = $1 AND product_id = $2 AND is_current = TRUE`,
		businessID, productID,
	)
	if err != nil {
		return fmt.Errorf("retire prior prices for %q: %w", slug, err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO business_intel.product_prices
			(product_id, business_id, price_gbp, price_qualifier, includes_vat,
			 in_stock, product_url, source, observed_at, is_current)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, NOW(), TRUE)`,
		productID, businessID, priceGBP, qualifier, inclVAT, inStock, sourceURL, source,
	)
	if err != nil {
		return fmt.Errorf("insert price for %q: %w", slug, err)
	}
	return nil
}

// insertPrice stores one service price against the unified schema
// (products kind='service' + product_prices). Replaces the deprecated
// business_prices write path — see 006_unify_prices_schema.sql.
func insertPrice(ctx context.Context, tx *sql.Tx, businessID string, price map[string]interface{}, sourceType, sourceURL string) error {
	category, _ := price["service_category"].(string)
	name, _ := price["service_name"].(string)
	if category == "" || name == "" {
		return fmt.Errorf("service_category and service_name are required")
	}

	priceGBP := nullFloatFromInterface(price["price_gbp"])
	qualifier, _ := price["price_qualifier"].(string)
	inclVAT := true
	if v, ok := price["includes_vat"].(bool); ok {
		inclVAT = v
	}

	src := sourceType
	if s, ok := price["source"].(string); ok && s != "" {
		src = s
	}
	url := sourceURL
	if u, ok := price["source_url"].(string); ok && u != "" {
		url = u
	}

	slug := "service-" + offeringSlug(category, name)
	return upsertOfferingPrice(ctx, tx,
		"service", slug, name, category, "", businessID,
		category == "prescription", priceGBP, qualifier, inclVAT,
		nil, src, url)
}

// insertMedicinePrice stores one per-practice medicine price from the
// verifier's medicine_prices[] output (products kind='medicine').
// Slug shape per HANDOFF_2026-05-18: {name}-{dosage}-{size_variant},
// e.g. apoquel-3-6mg-20-tablets.
func insertMedicinePrice(ctx context.Context, tx *sql.Tx, businessID string, med map[string]interface{}, sourceType, sourceURL string) error {
	name, _ := med["product_name"].(string)
	if name == "" {
		name, _ = med["name"].(string)
	}
	if name == "" {
		return fmt.Errorf("product_name is required")
	}
	dosage, _ := med["dosage"].(string)
	sizeVariant, _ := med["size_variant"].(string)

	priceGBP := nullFloatFromInterface(med["price_gbp"])
	qualifier, _ := med["price_qualifier"].(string)
	inclVAT := true
	if v, ok := med["includes_vat"].(bool); ok {
		inclVAT = v
	}
	var inStock interface{}
	if v, ok := med["in_stock"].(bool); ok {
		inStock = v
	}

	src := sourceType
	if s, ok := med["source"].(string); ok && s != "" {
		src = s
	}
	url := sourceURL
	if u, ok := med["source_url"].(string); ok && u != "" {
		url = u
	}

	slug := offeringSlug(name, dosage, sizeVariant)
	return upsertOfferingPrice(ctx, tx,
		"medicine", slug, name, "", dosage, businessID,
		true, priceGBP, qualifier, inclVAT,
		inStock, src, url)
}

// ============================================================================
// 1. cacheSearchResults - stores raw search results for later discovery mining
// ============================================================================
func cacheSearchResults(ctx context.Context, tx *sql.Tx, businessID string,
	collectedData map[string]interface{}, orchestrationID string, logger *zap.Logger) {

	// Try multiple paths for search results
	var results []interface{}
	var query, provider string

	// Check search_results.response or search_practice.response
	for _, key := range []string{"search_results", "search_practice"} {
		stepData, ok := collectedData[key].(map[string]interface{})
		if !ok {
			continue
		}
		resp, ok := stepData["response"].(map[string]interface{})
		if !ok {
			continue
		}
		if r, ok := resp["results"].([]interface{}); ok && len(r) > 0 {
			results = r
			query, _ = resp["query"].(string)
			provider, _ = resp["provider"].(string)
			break
		}
	}

	if len(results) == 0 {
		logger.Debug("cacheSearchResults: no search results to cache")
		return
	}

	resultsJSON, err := json.Marshal(results)
	if err != nil {
		logger.Warn("cacheSearchResults: failed to marshal results", zap.Error(err))
		return
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO business_intel.search_result_cache 
			(business_id, query, results_json, provider, result_count, orchestration_id)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		businessID, query, resultsJSON, provider, len(results), orchestrationID,
	)
	if err != nil {
		logger.Warn("cacheSearchResults: failed to insert", zap.Error(err))
	} else {
		logger.Info("cacheSearchResults: cached search results",
			zap.String("business_id", businessID),
			zap.String("query", query),
			zap.Int("result_count", len(results)))
	}
}

// ============================================================================
//  2. confirmContacts - bumps last_confirmed_at for contacts found in this run,
//     increments missed_count for contacts NOT found, marks stale after 3 misses
//
// ============================================================================
func confirmContacts(ctx context.Context, tx *sql.Tx, businessID string,
	bizData map[string]interface{}, logger *zap.Logger) {

	// Collect all contact values found in this extraction
	foundContacts := map[string]map[string]bool{} // contact_type -> set of values

	collectValues := func(field interface{}, contactType string) {
		if field == nil {
			return
		}
		if foundContacts[contactType] == nil {
			foundContacts[contactType] = map[string]bool{}
		}
		switch v := field.(type) {
		case string:
			if v != "" {
				foundContacts[contactType][v] = true
			}
		case []interface{}:
			for _, item := range v {
				if s, ok := item.(string); ok && s != "" {
					foundContacts[contactType][s] = true
				}
			}
		case []string:
			for _, s := range v {
				if s != "" {
					foundContacts[contactType][s] = true
				}
			}
		}
	}

	collectValues(bizData["phone"], "phone")
	collectValues(bizData["email"], "email")
	collectValues(bizData["fax"], "fax")

	// For each contact type we searched for, bump confirmed or increment missed
	for _, contactType := range []string{"phone", "email", "fax"} {
		found := foundContacts[contactType]
		if found == nil {
			// We didn't extract this type at all this run — don't penalise.
			// Only penalise if we actively looked and didn't find it.
			// We know we looked if the extraction had the field (even if empty array).
			_, fieldPresent := bizData[contactType]
			if !fieldPresent {
				continue
			}
			found = map[string]bool{} // empty — everything is missed
		}

		// Load existing contacts for this business + type
		rows, err := tx.QueryContext(ctx, `
			SELECT id, value FROM business_intel.business_contact_details 
			WHERE business_id = $1 AND contact_type = $2`,
			businessID, contactType,
		)
		if err != nil {
			logger.Warn("confirmContacts: failed to query existing", zap.Error(err))
			continue
		}

		type existingContact struct {
			id, value string
		}
		var existing []existingContact
		for rows.Next() {
			var c existingContact
			if err := rows.Scan(&c.id, &c.value); err == nil {
				existing = append(existing, c)
			}
		}
		rows.Close()

		for _, c := range existing {
			if found[c.value] {
				// Confirmed — bump timestamps and reset missed count
				_, err := tx.ExecContext(ctx, `
					UPDATE business_intel.business_contact_details 
					SET last_confirmed_at = NOW(), 
					    check_count = check_count + 1,
					    missed_count = 0,
					    is_stale = FALSE,
					    updated_at = NOW()
					WHERE id = $1`, c.id)
				if err != nil {
					logger.Warn("confirmContacts: failed to confirm", zap.String("id", c.id), zap.Error(err))
				}
			} else {
				// Not found this run — increment missed_count, mark stale if >= 3
				_, err := tx.ExecContext(ctx, `
					UPDATE business_intel.business_contact_details 
					SET missed_count = missed_count + 1,
					    is_stale = (missed_count + 1 >= 3),
					    updated_at = NOW()
					WHERE id = $1`, c.id)
				if err != nil {
					logger.Warn("confirmContacts: failed to increment missed", zap.String("id", c.id), zap.Error(err))
				}
			}
		}
	}
}

// ============================================================================
// 3. bumpVerificationCount - increments the business verification counter
// ============================================================================
func bumpVerificationCount(ctx context.Context, tx *sql.Tx, businessID string) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE business_intel.businesses 
		SET verification_count = COALESCE(verification_count, 0) + 1,
		    first_verified_at = COALESCE(first_verified_at, NOW())
		WHERE id = $1`, businessID)
	return err
}

// ---------------------------------------------------------------------------
// Null/type conversion helpers
// ---------------------------------------------------------------------------

func setIfValid(m map[string]interface{}, key string, v sql.NullString) {
	if v.Valid && v.String != "" {
		m[key] = v.String
	}
}

func setIfValidFloat(m map[string]interface{}, key string, v sql.NullFloat64) {
	if v.Valid {
		m[key] = v.Float64
	}
}

func setIfValidBool(m map[string]interface{}, key string, v sql.NullBool) {
	if v.Valid {
		m[key] = v.Bool
	}
}

func stringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func nullStringFromInterface(v interface{}) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	if s, ok := v.(string); ok {
		return sql.NullString{String: s, Valid: true}
	}
	return sql.NullString{}
}

func nullBoolFromInterface(v interface{}) sql.NullBool {
	if v == nil {
		return sql.NullBool{}
	}
	if b, ok := v.(bool); ok {
		return sql.NullBool{Bool: b, Valid: true}
	}
	return sql.NullBool{}
}

func nullIntFromInterface(v interface{}) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{}
	}
	switch n := v.(type) {
	case float64:
		return sql.NullInt32{Int32: int32(n), Valid: true}
	case int:
		return sql.NullInt32{Int32: int32(n), Valid: true}
	}
	return sql.NullInt32{}
}

func nullFloatFromInterface(v interface{}) sql.NullFloat64 {
	if v == nil {
		return sql.NullFloat64{}
	}
	if f, ok := v.(float64); ok {
		return sql.NullFloat64{Float64: f, Valid: true}
	}
	return sql.NullFloat64{}
}

func nullTimeIfBoolPresent(v interface{}) sql.NullTime {
	if v == nil {
		return sql.NullTime{}
	}
	if _, ok := v.(bool); ok {
		return sql.NullTime{Time: time.Now(), Valid: true}
	}
	return sql.NullTime{}
}

// pgArrayFromInterface converts a []interface{} or []string to a PostgreSQL array literal
func pgArrayFromInterface(v interface{}) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}

	var items []string
	switch arr := v.(type) {
	case []interface{}:
		for _, item := range arr {
			if s, ok := item.(string); ok {
				items = append(items, s)
			}
		}
	case []string:
		items = arr
	default:
		return sql.NullString{}
	}

	if len(items) == 0 {
		return sql.NullString{}
	}

	// Build PostgreSQL array literal: {item1,item2}
	return sql.NullString{
		String: "{" + strings.Join(items, ",") + "}",
		Valid:  true,
	}
}

// pgArrayToSlice converts a PostgreSQL array string like {a,b,c} to []string
func pgArrayToSlice(s string) []string {
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

// ============================================================================
// Company number regex extraction
// ============================================================================

// Patterns for extracting UK company registration numbers from website HTML.
// Used as a fallback when the LLM doesn't extract registration_number.
var companyRegNumberPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)company\s*(?:number|no\.?|reg(?:istration)?\.?)\s*[:.]?\s*(\d{7,8})`),
	regexp.MustCompile(`(?i)registered\s+(?:in\s+)?(?:england|wales|scotland|england\s*(?:&|and)\s*wales)[\s,]*(?:number|no\.?|reg\.?)?\s*[:.]?\s*(\d{7,8})`),
	regexp.MustCompile(`(?i)registered\s+company\s*[:.]?\s*(\d{7,8})`),
	regexp.MustCompile(`(?i)registration\s+(?:number|no\.?)\s*[:.]?\s*(\d{7,8})`),
	regexp.MustCompile(`(?i)reg\.?\s*no\.?\s*[:.]?\s*(\d{7,8})`),
	regexp.MustCompile(`\b(SC\d{6})\b`),
	regexp.MustCompile(`\b(NI\d{6})\b`),
}

// extractCompanyNumberFromCollectedData tries to find a company registration number
// in the scraped content from the verification workflow's collected data.
// Looks in scraped_data.content (the text content from scrape_web action).
func extractCompanyNumberFromCollectedData(collectedData map[string]interface{}, logger *zap.Logger) string {
	// Try scraped_data.content (from scrape_web action)
	var content string
	if scrapedData, ok := collectedData["scraped_data"].(map[string]interface{}); ok {
		if c, ok := scrapedData["content"].(string); ok {
			content = c
		}
	}

	if content == "" {
		return ""
	}

	// Focus on the bottom 40% of content (footer region) first
	footerStart := len(content) * 60 / 100
	if footerStart < len(content) {
		if num := extractRegNumber(content[footerStart:]); num != "" {
			return num
		}
	}

	// Fall back to full content
	return extractRegNumber(content)
}

// extractRegNumber applies regex patterns to find a company registration number.
func extractRegNumber(text string) string {
	for _, pattern := range companyRegNumberPatterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) >= 2 {
			num := strings.TrimSpace(matches[1])

			// SC/NI prefixed — return as-is
			if strings.HasPrefix(num, "SC") || strings.HasPrefix(num, "NI") {
				return num
			}

			// Validate: 7 or 8 digits
			if len(num) >= 7 && len(num) <= 8 {
				if len(num) == 7 {
					num = "0" + num
				}
				return num
			}
		}
	}
	return ""
}
