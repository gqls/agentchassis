// FILE: platform/orchestration/actions/load_existing_component_action.go
//
// LoadExistingComponentAction is the PRE-GENERATION HALF OF THE BIRTH GATE: it
// tells component-creator's generate_template prompt, before the LLM writes
// anything, what StoreGeneratedComponentAction is going to enforce afterwards.
// It carries two contracts today:
//
//   1. FIELD NAMES (+ the function pin) — the existing component's input_schema
//      field names, so a regeneration REUSES them and does not strand the
//      content_data that dependents are keyed on. Guard: the field-contract
//      block in store_generated_component_action.go.
//   2. SOURCE VOCABULARY — the aspects/queries a declared `source` may name, so
//      the writer does not invent one. Guard: SourceVocabularyIssues in
//      component_source_guard.go (bugs_open/309's birth gate).
//
// WHY BOTH LIVE HERE, AND WHY THEY COME FROM THE GUARD'S OWN CODE (bugs_open/337).
// A gate that enforces a contract its producer was never shown does not prevent
// the defect — it converts it into a retry loop, because the writer regenerates
// with exactly the information that failed last time. Measured 2026-08-22: 101
// component_validation_rejected rows in eight days across four sites, 11 work
// items parked `failed` at attempt 3, and pages shipped without their tool. The
// worked case is a one-character miss — the writer declared
// `site_specs.ctas.primary_url` when the aspect is `cta`, which carries exactly
// `primary_url` and `secondary_url`; it had never been shown a single aspect
// name, because the prompt renders no part of site_specs. So the advisory calls
// LoadKnownSpecAspects and KnownAspectsSorted — the guard's own loader and its
// own formatter — rather than re-querying: offer and enforcement are one
// computation, which is bugs_open/282's remedy and the shape 016b §9/092 already
// shipped and proved (writer's allow-list and gate's accept-set behind one
// predicate).
//
// WHY resolveStorageIdentity AND NOT A SECOND LOOKUP. The section_type query
// below is the SELECTOR's query (component_selector.go) and stays primary so
// nothing that works today changes. But the guard resolves the row it will
// overwrite by FUNCTION, through resolveStorageIdentity — which also carries the
// foreign-dependent diversion. A bare lookup by function would be wrong in two
// live cases and would MANUFACTURE refusals: on a divert-to-create there is no
// contract at all, and naming the incumbent's fields imposes a phantom one; on a
// divert-to-existing-scoped-row the real contract is the SCOPED row's schema, so
// naming the base row's fields makes the writer preserve the wrong set. Calling
// the store's own resolver gets all four cases right by construction. With an
// empty requesterSiteID it degrades to the plain by-function lookup, so nothing
// is lost when the site is unknown.
//
// Advisory by design, and that is load-bearing rather than incidental: a new
// section yields empty output and every prompt block stays dormant; ANY lookup
// problem degrades to blind generation rather than blocking, because the guard
// is still the backstop. So this action never returns an error and always
// returns a well-formed map, so the prompt's {{if ...}} guards are safe even if
// an upstream error_step routed around an earlier step.
//
// ⚠ THE TWO FAIL-OPENS ARE INDEPENDENT ON PURPOSE. The field lookup and the
// vocabulary read can each fail without the other, and each failure must cost
// only its own prompt block — a vocabulary read error must not blank the field
// names. See component_source_guard.go's note that the fail-open is the birth
// gate's POLICY, not the rule's; do not "make the callers consistent".
//
// Registration (registry.go):
//   "load_existing_component": {
//       Handler:     LoadExistingComponentAction,
//       Category:    "site",
//       Description: "Load the birth gate's contracts (field names, function pin, source vocabulary) for the component writer",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// aspectPathCoverageFloor is the minimum number of sites an
// `aspect.key` pair must appear on before it is offered to the writer.
//
// It is not arbitrary and it is not tuning. A generated component is SHARED
// across sites, so a source that resolves on one site renders blank on the rest
// — which is bugs_open/309's damage, arriving quietly. Offering only broadly
// carried paths keeps the default suggestion safe, and the prompt separately
// tells the writer to mark anything short of universal as
// required:false/on_missing:skip_field. Measured 2026-08-22: 446 aspect.key
// pairs exist, 187 of them on >= 5 sites (~5.1 KB of prompt against ~12.6 KB
// unfiltered).
const aspectPathCoverageFloor = 5

// aspectPathSampleLimit bounds how much vocabulary can reach the prompt however
// the estate's data grows. The block is advisory; a truncated list still beats
// the empty one the writer has today, and an unbounded one would let a data
// change silently inflate every component generation's input.
const aspectPathSampleLimit = 250

var LoadExistingComponentInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"section_type"},
	Optional:    []string{},
	Defaults:    map[string]interface{}{},
	Deprecated:  map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("load_existing_component", LoadExistingComponentInputSpec)
}

func LoadExistingComponentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "load_existing_component"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Always return a well-formed map so the prompt's
	// {{if .existing_component.field_names}} guard is safe, and never block
	// generation on a lookup problem (the store-time guard is the backstop).
	empty := map[string]interface{}{"field_names": "", "function": "", "field_count": 0}

	if params.DB == nil {
		logger.Warn("load_existing_component: no DB — generating blind")
		return empty, nil
	}

	// The source vocabulary is INDEPENDENT of whether an existing component is
	// found: a first-time creation invents an unresolvable source exactly as
	// readily as a regeneration does (the live bugs_open/337 failure was a
	// diverted CREATE). So it is loaded once, up front, and merged into every
	// return path below — not nested inside the found-a-row branch.
	vocabulary := loadSourceVocabulary(ctx, params.DB, logger)
	withVocabulary := func(out map[string]interface{}) map[string]interface{} {
		for k, v := range vocabulary {
			out[k] = v
		}
		return out
	}
	empty = withVocabulary(empty)

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, LoadExistingComponentInputSpec, logger,
	)
	if err != nil {
		logger.Warn("load_existing_component: input extraction failed — generating blind", zap.Error(err))
		return empty, nil
	}

	sectionType := inputs.Get("section_type")
	if sectionType == "" {
		logger.Info("load_existing_component: no section_type — generating blind")
		return empty, nil
	}

	// Canonical shared section component for this section_type (matches the
	// selector index: active, non-forked, section-level). If several exist,
	// prefer most-used then most-recent — the row dependents are most likely
	// bound to.
	//
	// "Most-used" is DERIVED from page_components (ComponentUsageSitesSQL), not read
	// from the stored content_components.usage_count column. That column was written
	// on only one of three resolution paths and also counted resolutions that never
	// became bindings, so it ranked by which route found a component rather than by
	// how established it is — and this ORDER BY is not a cosmetic tie-break: it picks
	// the row the store will overwrite and enforce as the contract. bugs_open/378.
	var function string
	var schemaJSON []byte
	err = params.DB.QueryRowContext(ctx, `
		SELECT function, input_schema
		FROM content_components
		WHERE section_type = $1
		  AND forked_from IS NULL
		  AND is_active = true
		  AND component_level = 'section'
		ORDER BY ` + ComponentUsageSitesSQL + ` DESC, updated_at DESC
		LIMIT 1
	`, sectionType).Scan(&function, &schemaJSON)
	if err != nil {
		// No row under the SELECTOR's key. That is not the same as "no
		// contract": the guard resolves by function, and a row whose
		// section_type is NULL, or which is deactivated pending regeneration,
		// is invisible here while still being the row the store will overwrite
		// and enforce. Ask the store's own resolver before concluding the
		// writer is unconstrained (bugs_open/337).
		logger.Info("load_existing_component: no row under section_type — asking the store's identity resolver",
			zap.String("section_type", sectionType), zap.Error(err))
		if fallback, ok := resolveContractViaStorageIdentity(ctx, params, sectionType, logger); ok {
			return withVocabulary(fallback), nil
		}
		return empty, nil
	}

	names := schemaFieldNamesSorted(schemaJSON)
	if len(names) == 0 {
		logger.Info("load_existing_component: existing component has no schema fields",
			zap.String("section_type", sectionType), zap.String("function", function))
		return withVocabulary(map[string]interface{}{"field_names": "", "function": function, "field_count": 0}), nil
	}

	logger.Info("load_existing_component: found existing component — requesting field-name preservation",
		zap.String("section_type", sectionType),
		zap.String("function", function),
		zap.Int("field_count", len(names)))

	return withVocabulary(map[string]interface{}{
		"field_names": strings.Join(names, ", "),
		"function":    function,
		"field_count": len(names),
	}), nil
}

// resolveContractViaStorageIdentity asks the STORE's own identity resolver which
// row a regeneration of this section_type would overwrite, and reports that
// row's field contract. ok=false means "advise nothing" — which covers a real
// creation, a diverted create, and every failure.
//
// The function name is derived exactly as store_generated_component_action.go
// derives it when the model does not supply one (NormaliseToKebab of the
// section_type), and the requester is read from the same two collected_data
// paths the store reads, so the prediction and the enforcement agree by
// construction rather than by coincidence.
//
// It reports ONLY when ident.IsRegeneration. A create has no field contract,
// and advising one there would be worse than silence: it would instruct the
// writer to reproduce an incumbent's schema into a brand-new row, and a writer
// told to reuse names tends to reuse their sources too — which is how an
// advisory manufactures a source-vocabulary refusal that would not otherwise
// have happened.
func resolveContractViaStorageIdentity(
	ctx context.Context, params ActionParams, sectionType string, logger *zap.Logger,
) (map[string]interface{}, bool) {
	functionName := datahelpers.NormaliseToKebab(sectionType)
	if functionName == "" {
		return nil, false
	}

	requesterSiteID := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.site_id")
	requesterDomain := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.domain")

	ident, err := resolveStorageIdentity(ctx, params.DB, functionName, requesterSiteID, requesterDomain, logger)
	if err != nil {
		// Advisory: a resolver problem must degrade to blind generation, never
		// block. The guard still refuses a contract-breaking template.
		logger.Warn("load_existing_component: storage identity resolution failed — generating blind",
			zap.String("section_type", sectionType),
			zap.String("function", functionName),
			zap.Error(err))
		return nil, false
	}
	if !ident.IsRegeneration {
		logger.Info("load_existing_component: storage identity resolves to a creation — no field contract to advise",
			zap.String("section_type", sectionType),
			zap.String("function", ident.FunctionName),
			zap.Bool("diverted", ident.Diverted))
		return nil, false
	}

	names := schemaFieldNamesSorted([]byte(ident.ExistingSchema))
	if len(names) == 0 {
		return nil, false
	}

	logger.Info("load_existing_component: contract recovered from the store's identity resolver — the section_type lookup was blind",
		zap.String("section_type", sectionType),
		zap.String("function", ident.FunctionName),
		zap.Bool("diverted", ident.Diverted),
		zap.Int("field_count", len(names)))

	return map[string]interface{}{
		"field_names": strings.Join(names, ", "),
		"function":    ident.FunctionName,
		"field_count": len(names),
	}, true
}

// loadSourceVocabulary renders the birth gate's OWN source vocabulary as
// prompt-ready strings. Every value here comes from the function the guard
// itself calls, so the offer and the enforcement cannot drift.
//
// Returns an empty map on any failure — the prompt blocks are {{if}}-guarded, so
// the writer simply guesses as it does today and the guard remains the backstop.
// This mirrors the guard's own deliberate fail-open on the same read.
func loadSourceVocabulary(ctx context.Context, db *sql.DB, logger *zap.Logger) map[string]interface{} {
	out := map[string]interface{}{}

	// Query names are compiled in, so this half cannot fail and cannot drift.
	if bases := queryresolve.KnownQueryBases(); len(bases) > 0 {
		out["known_query_bases"] = strings.Join(bases, ", ")
	}

	aspects, err := LoadKnownSpecAspects(ctx, db)
	if err != nil {
		logger.Warn("load_existing_component: site_specs aspect vocabulary unavailable — the writer will guess (guard still enforces)",
			zap.Error(err))
		return out
	}
	if len(aspects) > 0 {
		out["known_aspects"] = strings.Join(KnownAspectsSorted(aspects), ", ")
	}

	paths, err := loadAspectPathCoverage(ctx, db)
	if err != nil {
		logger.Warn("load_existing_component: aspect path coverage unavailable — offering aspect names only",
			zap.Error(err))
		return out
	}
	if len(paths) > 0 {
		out["aspect_paths"] = strings.Join(paths, "\n")
	}
	return out
}

// loadAspectPathCoverage lists the `aspect.key` paths a component may source,
// each with the number of sites carrying it.
//
// The coverage number is the point, not decoration. The guard validates only the
// ASPECT — the first segment — so a writer shown bare aspect names can still
// name a key that exists nowhere, pass the gate, and render blank, which is
// exactly bugs_open/309's damage arriving quietly instead of loudly. Showing the
// real leaf keys is what narrows that, and showing how many sites carry each is
// what lets the writer decide whether the field may be required at all.
func loadAspectPathCoverage(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT ss.aspect, k.key, count(DISTINCT ss.site_id) AS sites
		FROM site_specs ss
		CROSS JOIN LATERAL jsonb_object_keys(
			CASE WHEN jsonb_typeof(ss.data) = 'object' THEN ss.data ELSE '{}'::jsonb END
		) AS k(key)
		WHERE ss.is_current
		GROUP BY ss.aspect, k.key
		HAVING count(DISTINCT ss.site_id) >= $1
		ORDER BY ss.aspect, k.key
		LIMIT $2
	`, aspectPathCoverageFloor, aspectPathSampleLimit)
	if err != nil {
		return nil, fmt.Errorf("loadAspectPathCoverage: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var aspect, key string
		var sites int
		if scanErr := rows.Scan(&aspect, &key, &sites); scanErr != nil {
			return nil, fmt.Errorf("loadAspectPathCoverage: %w", scanErr)
		}
		out = append(out, fmt.Sprintf("       site_specs.%s.%s (%d sites)", aspect, key, sites))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("loadAspectPathCoverage: %w", err)
	}
	return out, nil
}

// schemaFieldNamesSorted extracts the sorted field names from an
// input_schema.fields object. Empty or invalid input → nil.
func schemaFieldNamesSorted(inputSchemaJSON []byte) []string {
	if len(inputSchemaJSON) == 0 {
		return nil
	}
	var schema map[string]interface{}
	if err := json.Unmarshal(inputSchemaJSON, &schema); err != nil {
		return nil
	}
	fields, ok := schema["fields"].(map[string]interface{})
	if !ok {
		return nil
	}
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
