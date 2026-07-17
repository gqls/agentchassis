// FILE: platform/orchestration/actions/directory_export_action.go
//
// Generic, config-driven business-directory + price JSON exporter. Serves any
// comparison vertical/domain — nothing site-specific may be hardcoded here.
//
// What it will publish, and nothing else:
//   - directory:   facts for verified businesses (id, name, location,
//                  postcode, website, is_claimed)
//   - aggregates:  k-anonymous area statistics (only where >= min_n
//                  businesses contribute; n always published alongside)
//   - claimed:     per-business prices the business provided itself
//                  (source='claimed_listing', is_claimed=true)
//   - attributed:  per-business prices scraped from the business's OWN
//                  website only (source host must match the business's
//                  website host), always with source URL + observed date;
//                  businesses with publication_optout are excluded
//
// Rows with source='seed_import' (quarantined fabricated data) are excluded
// unconditionally from every output. Policy: docs/agent_docs/
// docs024_key_docs_latest/vetcomparison/ PLAN §Phase 2 + LEGAL §7.
//
// Actions:
//   - directory_export_json: build JSON artefacts and commit to the target
//     site's git repository. Domain is REQUIRED — there is no default.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type dirExportConfig struct {
	VerticalSlug        string
	Domain              string
	RepoName            string
	DataPath            string
	CommitMessagePrefix string
	BusinessTypeILike   string // optional extra filter, e.g. "%vet%"

	Directory         bool
	DirectoryFilename string
	AggEnabled        bool
	AggMinN           int
	ClaimedPrices     bool
	AttributedPrices  bool // fail-closed default false; per-site config enables
}

func parseDirectoryExportConfig(config map[string]interface{}) dirExportConfig {
	ec := dirExportConfig{
		RepoName:            "sites",
		DataPath:            "data",
		CommitMessagePrefix: "Update directory data",
		Directory:           true,
		DirectoryFilename:   "directory.json",
		AggEnabled:          true,
		AggMinN:             3,
		ClaimedPrices:       true,
		AttributedPrices:    false,
	}
	if v, ok := config["vertical"].(string); ok {
		ec.VerticalSlug = v
	}
	if v, ok := config["domain"].(string); ok {
		ec.Domain = v
	}
	if v, ok := config["repo_name"].(string); ok && v != "" {
		ec.RepoName = v
	}
	if v, ok := config["data_path"].(string); ok && v != "" {
		ec.DataPath = v
	}
	if v, ok := config["commit_message_prefix"].(string); ok && v != "" {
		ec.CommitMessagePrefix = v
	}
	if v, ok := config["business_type_ilike"].(string); ok {
		ec.BusinessTypeILike = v
	}
	if outputs, ok := config["outputs"].(map[string]interface{}); ok {
		if v, ok := outputs["directory"].(bool); ok {
			ec.Directory = v
		}
		if v, ok := outputs["directory_filename"].(string); ok && v != "" {
			ec.DirectoryFilename = v
		}
		if agg, ok := outputs["aggregates"].(map[string]interface{}); ok {
			if v, ok := agg["enabled"].(bool); ok {
				ec.AggEnabled = v
			}
			if v, ok := agg["min_n"].(float64); ok && int(v) > 0 {
				ec.AggMinN = int(v)
			}
		}
		if v, ok := outputs["claimed_prices"].(bool); ok {
			ec.ClaimedPrices = v
		}
		if v, ok := outputs["attributed_prices"].(bool); ok {
			ec.AttributedPrices = v
		}
	}
	return ec
}

// DirectoryExportJSONAction builds the JSON artefacts and commits them to the
// target site's repository via the git-adapter.
func DirectoryExportJSONAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("DirectoryExportJSON: starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config
	if config == nil {
		config = make(map[string]interface{})
	}
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		for k, v := range inputData {
			config[k] = v
		}
	}

	ec := parseDirectoryExportConfig(config)
	if ec.Domain == "" {
		return nil, fmt.Errorf("directory export requires an explicit domain; refusing to export without one")
	}
	if ec.VerticalSlug == "" {
		return nil, fmt.Errorf("directory export requires an explicit vertical slug")
	}

	params.Logger.Info("DirectoryExportJSON: config parsed",
		zap.String("domain", ec.Domain), zap.String("vertical", ec.VerticalSlug),
		zap.Bool("attributed_prices", ec.AttributedPrices), zap.Int("agg_min_n", ec.AggMinN))

	files := make(map[string]interface{})
	basePath := trimTrailingSlash(ec.DataPath)
	counts := map[string]int{}

	if ec.Directory {
		entries, err := loadDirectoryEntries(ctx, params.DB, ec)
		if err != nil {
			return nil, fmt.Errorf("load directory: %w", err)
		}
		j, err := json.MarshalIndent(entries, "", " ")
		if err != nil {
			return nil, fmt.Errorf("marshal directory: %w", err)
		}
		files[basePath+"/"+ec.DirectoryFilename] = string(j)
		counts["directory"] = len(entries)
	}

	if ec.AggEnabled {
		aggs, err := loadPriceAggregates(ctx, params.DB, ec)
		if err != nil {
			return nil, fmt.Errorf("load aggregates: %w", err)
		}
		out := map[string]interface{}{
			"generated_at": time.Now().UTC().Format(time.RFC3339),
			"min_n":        ec.AggMinN,
			"note":         "Statistics are only published for areas where at least min_n businesses contribute; n is the number of businesses behind each figure.",
			"aggregates":   aggs,
		}
		j, err := json.MarshalIndent(out, "", " ")
		if err != nil {
			return nil, fmt.Errorf("marshal aggregates: %w", err)
		}
		files[basePath+"/price-aggregates.json"] = string(j)
		counts["aggregate_rows"] = len(aggs)
	}

	if ec.ClaimedPrices {
		claimed, err := loadBusinessPrices(ctx, params.DB, ec, true)
		if err != nil {
			return nil, fmt.Errorf("load claimed prices: %w", err)
		}
		j, err := json.MarshalIndent(claimed, "", " ")
		if err != nil {
			return nil, fmt.Errorf("marshal claimed prices: %w", err)
		}
		files[basePath+"/claimed-prices.json"] = string(j)
		counts["claimed_businesses"] = len(claimed)
	}

	if ec.AttributedPrices {
		attributed, err := loadBusinessPrices(ctx, params.DB, ec, false)
		if err != nil {
			return nil, fmt.Errorf("load attributed prices: %w", err)
		}
		j, err := json.MarshalIndent(attributed, "", " ")
		if err != nil {
			return nil, fmt.Errorf("marshal attributed prices: %w", err)
		}
		files[basePath+"/attributed-prices.json"] = string(j)
		counts["attributed_businesses"] = len(attributed)
	}

	meta := map[string]interface{}{
		"exported_at":        time.Now().UTC().Format(time.RFC3339),
		"domain":             ec.Domain,
		"vertical":           ec.VerticalSlug,
		"counts":             counts,
		"aggregates_min_n":   ec.AggMinN,
		"attributed_enabled": ec.AttributedPrices,
		"policy":             "No price is published without provenance (source URL + capture date) or the business's own consent via a claimed listing. Businesses may opt out of price display. Quarantined data is excluded unconditionally.",
	}
	if j, err := json.MarshalIndent(meta, "", " "); err == nil {
		files[basePath+"/directory-metadata.json"] = string(j)
	}

	params.Logger.Info("DirectoryExportJSON: built files",
		zap.Int("file_count", len(files)), zap.Any("counts", counts))

	if params.Producer == nil {
		return nil, fmt.Errorf("kafka producer not available")
	}
	commitMsg := fmt.Sprintf("%s — %d files", ec.CommitMessagePrefix, len(files))
	gitResult, err := sendExportFilesToGit(ctx, params, ec.RepoName, ec.Domain, commitMsg, files)
	if err != nil {
		return nil, fmt.Errorf("git commit: %w", err)
	}

	taskName := "directory-export-json"
	if tn, ok := config["task_name"].(string); ok && tn != "" {
		taskName = tn
	}
	_, _ = params.DB.ExecContext(ctx, `UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1`, taskName)

	return map[string]interface{}{
		"status": "complete", "domain": ec.Domain, "files": len(files),
		"counts": counts, "git_result": gitResult,
	}, nil
}

func trimTrailingSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

// ---------------------------------------------------------------------------
// Directory
// ---------------------------------------------------------------------------

type directoryEntry struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Postcode  string  `json:"postcode"`
	Location  *string `json:"location"`
	Website   string  `json:"website"`
	IsClaimed bool    `json:"is_claimed"`
}

func loadDirectoryEntries(ctx context.Context, db *sql.DB, ec dirExportConfig) ([]directoryEntry, error) {
	query := `
		SELECT b.id::text,
		       COALESCE(NULLIF(b.trading_name, ''), b.name) AS name,
		       b.postcode,
		       NULLIF(TRIM(BOTH ', ' FROM CONCAT_WS(', ', b.town, b.county)), ''),
		       b.website_url,
		       COALESCE(b.is_claimed, false)
		FROM business_intel.businesses b
		JOIN business_intel.business_verticals v ON v.id = b.vertical_id
		WHERE v.slug = $1
		  AND b.verification_status = 'verified'
		  AND b.website_url IS NOT NULL
		  AND b.postcode IS NOT NULL`
	args := []interface{}{ec.VerticalSlug}
	if ec.BusinessTypeILike != "" {
		query += ` AND b.business_type ILIKE $2`
		args = append(args, ec.BusinessTypeILike)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []directoryEntry
	for rows.Next() {
		var e directoryEntry
		var loc sql.NullString
		if err := rows.Scan(&e.ID, &e.Name, &e.Postcode, &loc, &e.Website, &e.IsClaimed); err != nil {
			return nil, err
		}
		if loc.Valid {
			e.Location = &loc.String
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	if entries == nil {
		entries = []directoryEntry{}
	}
	return entries, rows.Err()
}

// ---------------------------------------------------------------------------
// Aggregates — k-anonymous area statistics
// ---------------------------------------------------------------------------

type priceAggregate struct {
	Area        string  `json:"area"`
	ServiceSlug string  `json:"service_slug"`
	ServiceName string  `json:"service_name"`
	Businesses  int     `json:"n"`
	MedianGBP   float64 `json:"median_gbp"`
	MinGBP      float64 `json:"min_gbp"`
	MaxGBP      float64 `json:"max_gbp"`
}

func loadPriceAggregates(ctx context.Context, db *sql.DB, ec dirExportConfig) ([]priceAggregate, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT upper(substring(b.postcode from '^[A-Za-z]{1,2}')) AS area,
		       p.slug, p.name,
		       COUNT(DISTINCT pp.business_id) AS n,
		       ROUND(percentile_cont(0.5) WITHIN GROUP (ORDER BY pp.price_gbp)::numeric, 2)::float8,
		       MIN(pp.price_gbp)::float8, MAX(pp.price_gbp)::float8
		FROM business_intel.product_prices pp
		JOIN business_intel.products p ON p.id = pp.product_id AND p.kind = 'service'
		JOIN business_intel.businesses b ON b.id = pp.business_id
		JOIN business_intel.business_verticals v ON v.id = b.vertical_id
		WHERE v.slug = $1
		  AND pp.is_current
		  AND pp.price_gbp IS NOT NULL AND pp.price_gbp > 0
		  AND pp.source IS DISTINCT FROM 'seed_import'
		  AND b.verification_status = 'verified'
		  AND b.postcode IS NOT NULL
		GROUP BY 1, 2, 3
		HAVING COUNT(DISTINCT pp.business_id) >= $2
		ORDER BY 1, 4 DESC, 2`,
		ec.VerticalSlug, ec.AggMinN)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	aggs := []priceAggregate{}
	for rows.Next() {
		var a priceAggregate
		if err := rows.Scan(&a.Area, &a.ServiceSlug, &a.ServiceName, &a.Businesses,
			&a.MedianGBP, &a.MinGBP, &a.MaxGBP); err != nil {
			return nil, err
		}
		aggs = append(aggs, a)
	}
	return aggs, rows.Err()
}

// ---------------------------------------------------------------------------
// Per-business prices — claimed (consented) and attributed (own-site sourced)
// ---------------------------------------------------------------------------

type businessPrice struct {
	Slug       string   `json:"slug"`
	Name       string   `json:"name"`
	Category   string   `json:"category,omitempty"`
	Kind       string   `json:"kind"`
	PriceGBP   *float64 `json:"price_gbp"`
	Qualifier  string   `json:"price_qualifier,omitempty"`
	IncludesVAT bool    `json:"includes_vat"`
	PetBand    string   `json:"pet_band"`
	SourceURL  string   `json:"source_url,omitempty"`
	ObservedAt string   `json:"observed_at"`
}

type businessPrices struct {
	BusinessID string          `json:"business_id"`
	Name       string          `json:"name"`
	Provided   bool            `json:"provided_by_business"`
	Prices     []businessPrice `json:"prices"`
}

// loadBusinessPrices returns per-business price lists.
//
// claimed=true  -> only businesses with is_claimed, only rows the business
//                  provided (source='claimed_listing').
// claimed=false -> attributed scrape mode: only unclaimed, non-opted-out
//                  businesses; only rows observed on the business's OWN
//                  website (host match, after stripping www.), each with a
//                  source URL and observation date. seed_import never
//                  qualifies in either mode.
func loadBusinessPrices(ctx context.Context, db *sql.DB, ec dirExportConfig, claimed bool) ([]businessPrices, error) {
	var cond string
	if claimed {
		cond = `
		  AND COALESCE(b.is_claimed, false) = true
		  AND pp.source = 'claimed_listing'`
	} else {
		cond = `
		  AND COALESCE(b.is_claimed, false) = false
		  AND COALESCE(b.publication_optout, false) = false
		  AND pp.source IS DISTINCT FROM 'claimed_listing'
		  AND pp.product_url IS NOT NULL
		  AND pp.observed_at IS NOT NULL
		  AND (
		        hosts.price_host = hosts.site_host
		     OR hosts.price_host LIKE '%.' || hosts.site_host
		  )`
	}

	query := `
		SELECT b.id::text,
		       COALESCE(NULLIF(b.trading_name, ''), b.name),
		       p.slug, p.name, COALESCE(p.category, ''), p.kind,
		       pp.price_gbp::float8, COALESCE(pp.price_qualifier, ''),
		       COALESCE(pp.includes_vat, true), COALESCE(pp.pet_band, 'any'),
		       COALESCE(pp.product_url, ''), pp.observed_at
		FROM business_intel.product_prices pp
		JOIN business_intel.products p ON p.id = pp.product_id
		JOIN business_intel.businesses b ON b.id = pp.business_id
		JOIN business_intel.business_verticals v ON v.id = b.vertical_id
		CROSS JOIN LATERAL (
		    SELECT lower(regexp_replace(substring(pp.product_url from '^(?:https?://)?(?:www\.)?([^/?#]+)'), '^www\.', '')) AS price_host,
		           lower(regexp_replace(substring(b.website_url  from '^(?:https?://)?(?:www\.)?([^/?#]+)'), '^www\.', '')) AS site_host
		) hosts
		WHERE v.slug = $1
		  AND pp.is_current
		  AND pp.source IS DISTINCT FROM 'seed_import'
		  AND b.verification_status = 'verified'` + cond + `
		ORDER BY b.id, p.kind, p.category, p.name`

	rows, err := db.QueryContext(ctx, query, ec.VerticalSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byBusiness := map[string]*businessPrices{}
	var order []string
	for rows.Next() {
		var (
			bizID, bizName          string
			pr                      businessPrice
			priceGBP                sql.NullFloat64
			observedAt              time.Time
		)
		if err := rows.Scan(&bizID, &bizName, &pr.Slug, &pr.Name, &pr.Category, &pr.Kind,
			&priceGBP, &pr.Qualifier, &pr.IncludesVAT, &pr.PetBand,
			&pr.SourceURL, &observedAt); err != nil {
			return nil, err
		}
		if priceGBP.Valid {
			pr.PriceGBP = &priceGBP.Float64
		}
		pr.ObservedAt = observedAt.UTC().Format(time.RFC3339)

		bp, ok := byBusiness[bizID]
		if !ok {
			bp = &businessPrices{BusinessID: bizID, Name: bizName, Provided: claimed}
			byBusiness[bizID] = bp
			order = append(order, bizID)
		}
		bp.Prices = append(bp.Prices, pr)
	}
	out := make([]businessPrices, 0, len(order))
	for _, id := range order {
		out = append(out, *byBusiness[id])
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------------------
// Shared git-adapter sender (used by med + directory exporters)
// ---------------------------------------------------------------------------

func sendExportFilesToGit(ctx context.Context, params ActionParams, repoName, domain, commitMsg string, files map[string]interface{}) (map[string]interface{}, error) {
	correlationID := params.ExecutionContext.CorrelationID
	orchestrationID := params.ExecutionContext.OrchestrationID
	clientID := params.ExecutionContext.ClientID

	responsesTopic := params.ExecutionContext.ResponsesTopic
	if responsesTopic == "" {
		if alt, ok := params.CollectedData["__my_responses_topic__"].(string); ok && alt != "" {
			responsesTopic = alt
		}
	}
	if responsesTopic == "" {
		if parent, ok := params.CollectedData["__parent_responses_topic__"].(string); ok && parent != "" {
			responsesTopic = parent
		}
	}

	newRequestID := uuid.New().String()

	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id": correlationID, "orchestration_id": orchestrationID,
			"client_id": clientID, "step_name": params.ExecutionContext.StepName,
			"step_id": params.ExecutionContext.StepID, "request_id": newRequestID,
			"message_type": "request", "sender_agent_type": params.ExecutionContext.Sender.AgentType,
			"sender_agent_id": orchestrationID, "sender_pod_name": params.ExecutionContext.Sender.PodName,
			"responses_topic": responsesTopic,
		},
		"body": map[string]interface{}{
			"action": "commit",
			"data": map[string]interface{}{
				"repo_name": repoName, "domain": domain,
				"files": files, "commit_message": commitMsg,
			},
		},
	}

	requestBytes, err := json.Marshal(adapterRequest)
	if err != nil {
		return nil, fmt.Errorf("marshal git request: %w", err)
	}

	headers := map[string]string{
		"correlation_id": correlationID, "orchestration_id": orchestrationID,
		"request_id": newRequestID, "message_type": "request", "client_id": clientID,
	}

	if err := params.Producer.Produce(ctx, "system.adapter.git.requests", headers, []byte(correlationID), requestBytes); err != nil {
		return nil, fmt.Errorf("send to git adapter: %w", err)
	}

	params.Logger.Info("ExportToGit: sent to git-adapter",
		zap.String("request_id", newRequestID), zap.String("domain", domain),
		zap.String("repo", repoName), zap.Int("files", len(files)))

	return map[string]interface{}{
		"request_id": newRequestID, "topic": "system.adapter.git.requests",
		"domain": domain, "file_count": len(files), "await_response": true,
	}, nil
}
