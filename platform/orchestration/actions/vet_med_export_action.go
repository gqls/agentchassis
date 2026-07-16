// FILE: platform/orchestration/actions/vet_med_export_action.go
//
// Exports current med prices as JSON files and commits them to the configured
// target site's git repository via the git-adapter. The target domain is
// config-driven and mandatory — there is deliberately no default domain.
//
// Produces:
//   - data/medicine-prices.json — full price dataset grouped by product
//   - data/medicine-index.json — slim catalog for search/autocomplete
//   - data/medicines_by_letter/{A-Z}.json — price data bucketed by first letter
//   - data/price-metadata.json — export metadata (timestamp, counts, retailers)
//
// Actions:
//   - med_export_json: Export prices to JSON and commit to git

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"go.uber.org/zap"
)

// medExportConfig holds all configurable options for the export.
type medExportConfig struct {
	Domain              string
	RepoName            string
	DataPath            string
	CommitMessagePrefix string
	Filters             medExportFilters
	Outputs             medExportOutputs
}

type medExportFilters struct {
	Species    []string
	Categories []string
	Retailers  []string
}

type medExportOutputs struct {
	Index    bool
	Full     bool
	ByLetter bool
	Metadata bool
}

func parseMedExportConfig(config map[string]interface{}) medExportConfig {
	// Domain has no default: exports must name their target site explicitly
	// (agent config or input_data), so a misconfigured task can never publish
	// to a domain we don't intend.
	ec := medExportConfig{
		RepoName:            "sites",
		DataPath:            "data",
		CommitMessagePrefix: "Update medicine prices",
		Outputs: medExportOutputs{
			Index: true, Full: true, ByLetter: true, Metadata: true,
		},
	}

	if d, ok := config["domain"].(string); ok && d != "" {
		ec.Domain = d
	}
	if r, ok := config["repo_name"].(string); ok && r != "" {
		ec.RepoName = r
	}
	if dp, ok := config["data_path"].(string); ok && dp != "" {
		ec.DataPath = dp
	}
	if cmp, ok := config["commit_message_prefix"].(string); ok && cmp != "" {
		ec.CommitMessagePrefix = cmp
	}

	if filters, ok := config["filters"].(map[string]interface{}); ok {
		ec.Filters.Species = extractMedStringSlice(filters, "species")
		ec.Filters.Categories = extractMedStringSlice(filters, "categories")
		ec.Filters.Retailers = extractMedStringSlice(filters, "retailers")
	}

	if outputs, ok := config["outputs"].(map[string]interface{}); ok {
		if v, ok := outputs["index"].(bool); ok {
			ec.Outputs.Index = v
		}
		if v, ok := outputs["full"].(bool); ok {
			ec.Outputs.Full = v
		}
		if v, ok := outputs["by_letter"].(bool); ok {
			ec.Outputs.ByLetter = v
		}
		if v, ok := outputs["metadata"].(bool); ok {
			ec.Outputs.Metadata = v
		}
	}

	return ec
}

func extractMedStringSlice(m map[string]interface{}, key string) []string {
	val, ok := m[key]
	if !ok {
		return nil
	}
	switch v := val.(type) {
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	case string:
		if v != "" {
			return []string{v}
		}
	}
	return nil
}

// MedExportJSONAction queries current prices and commits JSON files to git.
func MedExportJSONAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("MedExportJSONAction: Starting")

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

	ec := parseMedExportConfig(config)

	if ec.Domain == "" {
		return nil, fmt.Errorf("med export requires an explicit domain in config or input_data; refusing to export without one")
	}

	params.Logger.Info("MedExportJSON: config parsed",
		zap.String("domain", ec.Domain),
		zap.String("data_path", ec.DataPath),
		zap.Strings("filter_species", ec.Filters.Species),
		zap.Strings("filter_categories", ec.Filters.Categories),
		zap.Strings("filter_retailers", ec.Filters.Retailers))

	// 1. Refresh materialized view
	_, err := params.DB.ExecContext(ctx, `REFRESH MATERIALIZED VIEW CONCURRENTLY business_intel.med_price_current`)
	if err != nil {
		params.Logger.Warn("MedExportJSON: view refresh failed, using stale data", zap.Error(err))
	}

	// 2. Query current prices with filters
	prices, err := loadMedPricesForExport(ctx, params.DB, ec.Filters)
	if err != nil {
		return nil, fmt.Errorf("load prices: %w", err)
	}

	params.Logger.Info("MedExportJSON: loaded price data", zap.Int("total_variants", len(prices)))

	if len(prices) == 0 {
		return map[string]interface{}{"status": "complete", "skipped": true, "reason": "no prices to export", "domain": ec.Domain}, nil
	}

	// 3. Group by product
	products := groupMedPricesByProduct(prices)

	params.Logger.Info("MedExportJSON: grouped products", zap.Int("unique_products", len(products)))

	// 4. Build JSON files
	files := make(map[string]interface{})
	basePath := strings.TrimRight(ec.DataPath, "/")

	if ec.Outputs.Index {
		if j, err := buildMedIndex(products); err == nil {
			files[basePath+"/medicine-index.json"] = j
		}
	}
	if ec.Outputs.Full {
		if j, err := buildMedFullPriceJSON(products); err == nil {
			files[basePath+"/medicine-prices.json"] = j
		}
	}
	if ec.Outputs.ByLetter {
		if lf, err := buildMedLetterBucketedJSON(products, basePath); err == nil {
			for path, content := range lf {
				files[path] = content
			}
		}
	}
	if ec.Outputs.Metadata {
		if j, err := buildMedExportMetadata(prices, products, ec); err == nil {
			files[basePath+"/price-metadata.json"] = j
		}
	}

	params.Logger.Info("MedExportJSON: built JSON files", zap.Int("file_count", len(files)))

	// 5. Send to git-adapter
	if params.Producer == nil {
		return nil, fmt.Errorf("kafka producer not available")
	}

	result, err := sendMedExportToGit(ctx, params, ec, files)
	if err != nil {
		return nil, fmt.Errorf("git commit: %w", err)
	}

	// 6. Notify scheduler
	taskName := "med-export-json"
	if tn, ok := config["task_name"].(string); ok && tn != "" {
		taskName = tn
	}
	_, _ = params.DB.ExecContext(ctx, `UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1`, taskName)

	return map[string]interface{}{
		"status": "complete", "domain": ec.Domain,
		"files": len(files), "products": len(products), "variants": len(prices),
		"git_result": result,
	}, nil
}

// ============================================================================
// Data loading with filters
// ============================================================================

type medPriceExportRow struct {
	ListingID       string
	RetailerID      string
	RetailerName    string
	GroupName       string
	RetailerDomain  string
	RetailerURL     string
	ProductName     string // from med_products if matched, else retailer_product_name
	Brand           string
	Category        string
	Species         string
	SizeVariant     string
	Price           float64
	InStock         bool
	TypicalVetPrice float64
	CollectedAt     time.Time
}

func loadMedPricesForExport(ctx context.Context, db *sql.DB, filters medExportFilters) ([]medPriceExportRow, error) {
	query := `
		SELECT DISTINCT ON (ps.listing_id, ps.size_variant)
			ps.listing_id::text, ps.retailer_id, r.name, COALESCE(r.group_name, ''),
			r.domain, l.retailer_url,
			COALESCE(p.name, l.retailer_product_name, 'Unknown'),
			COALESCE(p.brand, ''), COALESCE(p.category, ''),
			COALESCE(array_to_string(p.species, ','), ''),
			COALESCE(ps.size_variant, ''), ps.price,
			COALESCE(ps.in_stock, true), COALESCE(ps.typical_vet_price, 0),
			ps.collected_at
		FROM business_intel.med_price_snapshots ps
		JOIN business_intel.med_retailer_listings l ON l.id = ps.listing_id
		JOIN business_intel.med_retailers r ON r.id = ps.retailer_id
		LEFT JOIN business_intel.med_products p ON p.id = ps.product_id
		WHERE r.is_active = true
		  AND ps.collected_at > NOW() - INTERVAL '14 days'`

	args := []interface{}{}
	argIdx := 1

	if len(filters.Retailers) > 0 {
		ph := make([]string, len(filters.Retailers))
		for i, r := range filters.Retailers {
			ph[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, r)
			argIdx++
		}
		query += fmt.Sprintf(" AND ps.retailer_id IN (%s)", strings.Join(ph, ","))
	}
	if len(filters.Species) > 0 {
		ph := make([]string, len(filters.Species))
		for i, s := range filters.Species {
			ph[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, s)
			argIdx++
		}
		query += fmt.Sprintf(" AND (p.species IS NULL OR p.species && ARRAY[%s]::text[])", strings.Join(ph, ","))
	}
	if len(filters.Categories) > 0 {
		ph := make([]string, len(filters.Categories))
		for i, c := range filters.Categories {
			ph[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, c)
			argIdx++
		}
		query += fmt.Sprintf(" AND (p.category IS NULL OR p.category IN (%s))", strings.Join(ph, ","))
	}

	query += " ORDER BY ps.listing_id, ps.size_variant, ps.collected_at DESC"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []medPriceExportRow
	for rows.Next() {
		var p medPriceExportRow
		if err := rows.Scan(
			&p.ListingID, &p.RetailerID, &p.RetailerName, &p.GroupName,
			&p.RetailerDomain, &p.RetailerURL, &p.ProductName, &p.Brand,
			&p.Category, &p.Species, &p.SizeVariant, &p.Price,
			&p.InStock, &p.TypicalVetPrice, &p.CollectedAt,
		); err != nil {
			continue
		}
		prices = append(prices, p)
	}
	return prices, rows.Err()
}

// ============================================================================
// Grouping — by product name (normalised) across retailers
// ============================================================================

type medExportProduct struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Brand    string            `json:"brand,omitempty"`
	Category string            `json:"category,omitempty"`
	Species  string            `json:"species,omitempty"`
	Options  []medExportOption `json:"options"`
}

type medExportOption struct {
	Retailer        string  `json:"retailer"`
	RetailerGroup   string  `json:"retailer_group,omitempty"`
	URL             string  `json:"url"`
	SizeVariant     string  `json:"size"`
	Price           float64 `json:"price"`
	InStock         bool    `json:"in_stock"`
	TypicalVetPrice float64 `json:"typical_vet_price,omitempty"`
	CollectedAt     string  `json:"collected_at"`
}

func groupMedPricesByProduct(prices []medPriceExportRow) []medExportProduct {
	productMap := make(map[string]*medExportProduct)
	orderKeys := []string{}

	for _, p := range prices {
		key := normaliseMedExportProductName(p.ProductName)
		product, exists := productMap[key]
		if !exists {
			product = &medExportProduct{
				ID: key, Name: p.ProductName, Brand: p.Brand,
				Category: p.Category, Species: p.Species,
			}
			productMap[key] = product
			orderKeys = append(orderKeys, key)
		}
		product.Options = append(product.Options, medExportOption{
			Retailer: p.RetailerName, RetailerGroup: p.GroupName,
			URL: p.RetailerURL, SizeVariant: p.SizeVariant,
			Price: p.Price, InStock: p.InStock,
			TypicalVetPrice: p.TypicalVetPrice,
			CollectedAt:     p.CollectedAt.Format("2006-01-02"),
		})
	}

	sort.Strings(orderKeys)
	products := make([]medExportProduct, 0, len(orderKeys))
	for _, key := range orderKeys {
		p := productMap[key]
		sort.Slice(p.Options, func(i, j int) bool { return p.Options[i].Price < p.Options[j].Price })
		products = append(products, *p)
	}
	return products
}

func normaliseMedExportProductName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	for _, suffix := range []string{" for dogs", " for cats", " for dogs and cats", " for horses"} {
		s = strings.TrimSuffix(s, suffix)
	}
	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevHyphen = false
		} else if !prevHyphen {
			b.WriteRune('-')
			prevHyphen = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// ============================================================================
// JSON builders
// ============================================================================

type medIndexEntry struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Brand    string `json:"brand,omitempty"`
	Category string `json:"category,omitempty"`
	Options  int    `json:"options"`
}

func buildMedicineIndex(products []medExportProduct) (string, error) {
	index := make([]medIndexEntry, len(products))
	for i, p := range products {
		index[i] = medIndexEntry{
			ID:       p.ID,
			Name:     p.Name,
			Brand:    p.Brand,
			Category: p.Category,
			Options:  len(p.Options),
		}
	}
	b, err := json.MarshalIndent(index, "", "  ")
	return string(b), err
}

func buildFullPriceJSON(products []medExportProduct) (string, error) {
	b, err := json.MarshalIndent(products, "", "  ")
	return string(b), err
}

func buildLetterBucketedJSON(products []medExportProduct) (map[string]string, error) {
	buckets := make(map[string][]medExportProduct)

	for _, p := range products {
		letter := "other"
		if len(p.Name) > 0 {
			first := unicode.ToUpper(rune(p.Name[0]))
			if unicode.IsLetter(first) {
				letter = string(first)
			}
		}
		buckets[letter] = append(buckets[letter], p)
	}

	files := make(map[string]string)
	for letter, prods := range buckets {
		b, err := json.MarshalIndent(prods, "", "  ")
		if err != nil {
			return nil, err
		}
		files[fmt.Sprintf("data/medicines_by_letter/%s.json", letter)] = string(b)
	}

	return files, nil
}

type medExportMetadata struct {
	ExportedAt    string            `json:"exported_at"`
	TotalProducts int               `json:"total_products"`
	TotalVariants int               `json:"total_variants"`
	Retailers     []medRetailerStat `json:"retailers"`
	DataFreshness string            `json:"data_freshness"`
}

type medRetailerStat struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Products int    `json:"products"`
	Variants int    `json:"variants"`
}

// ============================================================================
// JSON builders
// ============================================================================

func buildMedIndex(products []medExportProduct) (string, error) {
	type entry struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Brand    string `json:"brand,omitempty"`
		Category string `json:"category,omitempty"`
		Options  int    `json:"options"`
	}
	index := make([]entry, len(products))
	for i, p := range products {
		index[i] = entry{ID: p.ID, Name: p.Name, Brand: p.Brand, Category: p.Category, Options: len(p.Options)}
	}
	b, err := json.MarshalIndent(index, "", "  ")
	return string(b), err
}

func buildMedFullPriceJSON(products []medExportProduct) (string, error) {
	b, err := json.MarshalIndent(products, "", "  ")
	return string(b), err
}

func buildMedLetterBucketedJSON(products []medExportProduct, basePath string) (map[string]string, error) {
	buckets := make(map[string][]medExportProduct)
	for _, p := range products {
		letter := "other"
		if len(p.Name) > 0 {
			first := unicode.ToUpper(rune(p.Name[0]))
			if unicode.IsLetter(first) {
				letter = string(first)
			}
		}
		buckets[letter] = append(buckets[letter], p)
	}
	files := make(map[string]string)
	for letter, prods := range buckets {
		b, err := json.MarshalIndent(prods, "", "  ")
		if err != nil {
			return nil, err
		}
		files[fmt.Sprintf("%s/medicines_by_letter/%s.json", basePath, letter)] = string(b)
	}
	return files, nil
}

func buildMedExportMetadata(prices []medPriceExportRow, products []medExportProduct, ec medExportConfig) (string, error) {
	retailerProducts := make(map[string]map[string]bool)
	retailerVariants := make(map[string]int)
	retailerNames := make(map[string]string)

	for _, p := range prices {
		if retailerProducts[p.RetailerID] == nil {
			retailerProducts[p.RetailerID] = make(map[string]bool)
		}
		retailerProducts[p.RetailerID][p.ListingID] = true
		retailerVariants[p.RetailerID]++
		retailerNames[p.RetailerID] = p.RetailerName
	}

	var retailers []medRetailerStat
	for id, prods := range retailerProducts {
		retailers = append(retailers, medRetailerStat{ID: id, Name: retailerNames[id], Products: len(prods), Variants: retailerVariants[id]})
	}
	sort.Slice(retailers, func(i, j int) bool { return retailers[i].Name < retailers[j].Name })

	var filtersOut interface{}
	if len(ec.Filters.Species) > 0 || len(ec.Filters.Categories) > 0 || len(ec.Filters.Retailers) > 0 {
		filtersOut = map[string]interface{}{"species": ec.Filters.Species, "categories": ec.Filters.Categories, "retailers": ec.Filters.Retailers}
	}

	type metadata struct {
		ExportedAt    string            `json:"exported_at"`
		Domain        string            `json:"domain"`
		TotalProducts int               `json:"total_products"`
		TotalVariants int               `json:"total_variants"`
		Retailers     []medRetailerStat `json:"retailers"`
		DataFreshness string            `json:"data_freshness"`
		Filters       interface{}       `json:"filters,omitempty"`
	}

	meta := metadata{
		ExportedAt: time.Now().UTC().Format(time.RFC3339), Domain: ec.Domain,
		TotalProducts: len(products), TotalVariants: len(prices),
		Retailers: retailers, DataFreshness: "14 days", Filters: filtersOut,
	}
	b, err := json.MarshalIndent(meta, "", "  ")
	return string(b), err
}

// ============================================================================
// Git commit
// ============================================================================

func sendMedExportToGit(ctx context.Context, params ActionParams, ec medExportConfig, files map[string]interface{}) (map[string]interface{}, error) {
	variantCount := 0
	if full, ok := files[strings.TrimRight(ec.DataPath, "/")+"/medicine-prices.json"].(string); ok {
		variantCount = strings.Count(full, `"price"`)
	}
	commitMsg := fmt.Sprintf("%s — %d files, %d variants", ec.CommitMessagePrefix, len(files), variantCount)
	return sendExportFilesToGit(ctx, params, ec.RepoName, ec.Domain, commitMsg, files)
}

func countVariants(files map[string]interface{}) int {
	// Rough count from the full prices file
	if full, ok := files["data/medicine-prices.json"].(string); ok {
		return strings.Count(full, `"price"`)
	}
	return 0
}
