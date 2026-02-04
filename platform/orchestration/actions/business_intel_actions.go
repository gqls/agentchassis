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
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

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
//	    "source_type": "web_scrape",
//	    "source_name": "practice_website",
//	    "source_url": "https://...",
//	    "confidence_score": 0.85,
//	    "extraction_notes": "..."
//	}
func StoreBusinessVerificationAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("StoreBusinessVerificationAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config

	// Extract inputs
	inputFields := []string{"business_id", "verification_result"}
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

	verResult, ok := extracted["verification_result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("verification_result must be an object")
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

	// 3. Store prices (mark old as not current, insert new)
	pricesStored := 0
	if pricesRaw, ok := verResult["prices"].([]interface{}); ok && len(pricesRaw) > 0 {
		// Mark existing prices as not current
		_, _ = tx.ExecContext(ctx, `
			UPDATE business_intel.business_prices 
			SET is_current = FALSE 
			WHERE business_id = $1 AND is_current = TRUE`,
			businessID,
		)

		sourceType, _ := verResult["source_type"].(string)
		sourceURL, _ := verResult["source_url"].(string)

		for _, priceRaw := range pricesRaw {
			price, ok := priceRaw.(map[string]interface{})
			if !ok {
				continue
			}
			err := insertPrice(ctx, tx, businessID, price, sourceType, sourceURL)
			if err != nil {
				params.Logger.Warn("StoreBusinessVerificationAction: Failed to insert price",
					zap.Error(err),
				)
				continue
			}
			pricesStored++
		}
	}

	// 4. Create data observation (provenance record)
	rawDataJSON, _ := json.Marshal(verResult)
	sourceType, _ := verResult["source_type"].(string)
	sourceName, _ := verResult["source_name"].(string)
	sourceURL, _ := verResult["source_url"].(string)
	extractionNotes, _ := verResult["extraction_notes"].(string)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO business_intel.data_observations 
			(business_id, source_type, source_name, source_url, raw_data, 
			 extraction_confidence, extraction_notes, orchestration_id, processed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`,
		businessID, sourceType, sourceName, sourceURL, rawDataJSON,
		confidenceScore, extractionNotes, params.ExecutionContext.OrchestrationID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert data observation: %w", err)
	}

	// 5. Update collection task if orchestration_id is known
	_, _ = tx.ExecContext(ctx, `
		UPDATE business_intel.collection_tasks 
		SET status = 'completed', 
		    completed_at = NOW(),
		    result_summary = $2
		WHERE orchestration_id = $1 AND status = 'in_progress'`,
		params.ExecutionContext.OrchestrationID,
		rawDataJSON,
	)

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	result := map[string]interface{}{
		"stored":         true,
		"business_id":    businessID,
		"updated_fields": updatedFields,
		"vet_updated":    vetUpdated,
		"prices_stored":  pricesStored,
		"stored_at":      time.Now().UTC().Format(time.RFC3339),
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
	rows, err := db.QueryContext(ctx, `
		SELECT service_category, service_name, price_gbp, price_qualifier,
		       includes_vat, source, source_url, observed_at
		FROM business_intel.business_prices
		WHERE business_id = $1 AND is_current = TRUE
		ORDER BY service_category, service_name`, businessID)
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

func updateBusinessFields(ctx context.Context, tx *sql.Tx, businessID string, data map[string]interface{}) (int, error) {
	// Only update known safe columns
	allowedFields := map[string]string{
		"name":          "name",
		"trading_name":  "trading_name",
		"address_line1": "address_line1",
		"address_line2": "address_line2",
		"town":          "town",
		"county":        "county",
		"postcode":      "postcode",
		"phone":         "phone",
		"email":         "email",
		"website_url":   "website_url",
		"business_type": "business_type",
		"group_name":    "group_name",
		"latitude":      "latitude",
		"longitude":     "longitude",
	}

	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	for jsonField, dbCol := range allowedFields {
		if val, ok := data[jsonField]; ok && val != nil {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", dbCol, argIdx))
			args = append(args, val)
			argIdx++
		}
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

	_, err := tx.ExecContext(ctx, `
		INSERT INTO business_intel.business_prices 
			(business_id, service_category, service_name, price_gbp, price_qualifier,
			 includes_vat, source, source_url, is_current)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, TRUE)`,
		businessID, category, name, priceGBP, qualifier, inclVAT, src, url,
	)
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
