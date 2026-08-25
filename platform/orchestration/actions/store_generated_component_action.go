// FILE: platform/orchestration/actions/store_generated_component_action.go
//
// StoreGeneratedComponentAction stores an LLM-generated component template
// into the content_components table with full selection metadata.
//
// Used by the component-creator handler agent after execute_llm_prompt
// generates the html_template and input_schema.
//
// Registration:
//   "store_generated_component": {
//       Handler:     StoreGeneratedComponentAction,
//       Category:    "site",
//       Description: "Store a generated component template in the component library",
//       IsLocal:     true,
//   },
//
// Workflow config:
//   "store_component": {
//       "action": "store_generated_component",
//       "config": {
//           "section_type": "input_data.section_type",
//           "site_type": "input_data.site_type",
//           "generated_template": "generate_template"
//       },
//       "next_step": "complete",
//       "output_field": "stored_component"
//   }

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/content"
	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var StoreGeneratedComponentInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"section_type"},
	Optional:    []string{"site_type", "page_context", "description", "design_direction", "generated_template", "advised_identity"},
	Defaults:    map[string]interface{}{},
	Deprecated:  map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("store_generated_component", StoreGeneratedComponentInputSpec)
}

func StoreGeneratedComponentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "store_generated_component"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		StoreGeneratedComponentInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	sectionType := inputs.Get("section_type")
	siteType := inputs.Get("site_type")
	pageContext := inputs.Get("page_context")
	description := inputs.Get("description")

	// Work item source (optional): the originating work item's `source`
	// field, if this action was triggered by one. Used as change_source
	// on component_versions rows so history can be traced back to the
	// audit/triage/manual action that caused the change. Empty string if
	// no work item (e.g. direct programmatic invocation) — the snapshot
	// helper writes NULL in that case.
	workItemSource := datahelpers.ExtractNestedFieldString(
		params.CollectedData, "input_data.source",
	)

	// The LLM output is in collected_data under the output_field of the generate step.
	// extract the generated template — look for the LLM result which contains
	// html_template and input_schema as structured output.
	generatedRaw := inputs.GetRaw("generated_template")

	htmlTemplate, inputSchemaJSON, functionName, isDark, err := parseGeneratedTemplate(generatedRaw, sectionType, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated template: %w", err)
	}

	if htmlTemplate == "" {
		return nil, fmt.Errorf("generated template is empty for section_type %q", sectionType)
	}

	// ── Resolve the storage identity (bugs_open/311) ────────────────────
	// Decided HERE, before anything derives from the function name:
	// separateInlineJS below writes a `/tools/assets/<function>.js` src
	// reference into the template, so a diverted (site-suffixed) identity
	// must be final first. resolveStorageIdentity also carries the
	// regen-vs-create decision that used to live just above the write —
	// including the rule that a base row OTHER sites depend on is never
	// this build's regeneration target (see component_storage_identity.go).
	requesterSiteID := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.site_id")
	requesterDomain := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.domain")

	// THE ADVISED IDENTITY WINS OVER THE ONE THE MODEL WROTE (bugs_open/388).
	//
	// functionName above came from the LLM's own output (parseGeneratedTemplate
	// takes data["function"] when non-empty, else the section_type). Until this
	// existed, that string decided which row this write overwrites — while the
	// field contract the writer was TOLD to preserve came from a different
	// resolver in load_existing_component_action.go. The two were joined by a
	// sentence in the prompt asking the model to echo a name back, which nothing
	// validated and nothing recorded.
	//
	// The pin is the row's PRIMARY KEY, not its name, and that is measured
	// rather than preferred: `function` does not identify a row here
	// (component_storage_identity.go's lookup filters neither component_level
	// nor is_active, and [MEASURED 2026-08-25] 25 function values carry more
	// than one non-forked row, the largest five, spanning component levels).
	//
	// ABSENT IS THE DEFAULT AND ABSENT IS SAFE. No pin — an un-wired workflow, an
	// error_step that routed around the advisory, a genuine creation, any
	// fail-open in the advisory — and the legacy resolution below runs verbatim.
	advisedID, advisedFunction := advisedIdentityPin(inputs.GetRaw("advised_identity"))

	var ident storageIdentity
	if advisedID != "" {
		var pinned bool
		ident, pinned, err = resolveStorageIdentityByID(
			ctx, params.DB, advisedID, requesterSiteID, requesterDomain, logger)
		if err != nil {
			return nil, fmt.Errorf("pinned storage identity resolution failed for %s: %w", advisedID, err)
		}
		if !pinned {
			// The advised row has been deleted, or has become a fork, since the
			// advisory ran. Fall back — but SAY SO, because this race is
			// otherwise invisible and a silent fallback is indistinguishable
			// from the pin never having been wired.
			logger.Warn("store_generated_component: advised component no longer resolvable — falling back to the function lookup",
				zap.String("advised_component_id", advisedID),
				zap.String("function", functionName))
			LogActionFindings(ctx, params, requesterSiteID, "",
				"store_generated_component", []agenterrors.Finding{{
					ErrorCode: "COMPONENT_ADVISED_ROW_VANISHED",
					Severity:  "warning",
					Message: fmt.Sprintf("the pre-generation advisory named component %s for section_type %q, but no non-forked row with that id exists at store time — resolving by function %q instead; the writer was shown a contract for a row that is gone",
						advisedID, sectionType, functionName),
					Context: map[string]interface{}{
						"advised_component_id": advisedID,
						"advised_function":     advisedFunction,
						"emitted_function":     functionName,
						"section_type":         sectionType,
					},
				}}, logger)
			ident, err = resolveStorageIdentity(ctx, params.DB, functionName, requesterSiteID, requesterDomain, logger)
			if err != nil {
				return nil, fmt.Errorf("storage identity resolution failed for %q: %w", functionName, err)
			}
		} else if advisedFunction != "" && functionName != advisedFunction {
			// HARMLESS NOW, AND RECORDED PRECISELY BECAUSE IT IS HARMLESS. The
			// pinned id already decided the row, so the model's disagreement
			// costs nothing — which is what makes this the honest meter for how
			// often it disagrees. Before the pin, obedience could only be
			// sampled: 11 observations, zero failures, which at n=11 bounds the
			// disobedience rate no better than ~24% at 95%. Now it is counted.
			//
			// Compared against the ADVISED name, never against ident.FunctionName
			// — the latter may since have been site-suffixed by the diversion,
			// and comparing with that would report a divergence on every
			// diverted write.
			logger.Info("store_generated_component: the writer chose a different function than it was pinned to — the advised row still wins",
				zap.String("advised_function", advisedFunction),
				zap.String("emitted_function", functionName),
				zap.String("advised_component_id", advisedID))
			LogActionFindings(ctx, params, requesterSiteID, "",
				"store_generated_component", []agenterrors.Finding{{
					ErrorCode: "COMPONENT_FUNCTION_PIN_DIVERGENCE",
					Severity:  "warning",
					Message: fmt.Sprintf("the writer emitted function %q for section_type %q after being pinned to %q (component %s); the pinned row was written, so nothing forked — before bugs_open/388 this would have resolved a different row or created a parallel duplicate",
						functionName, sectionType, advisedFunction, advisedID),
					Context: map[string]interface{}{
						"advised_component_id": advisedID,
						"advised_function":     advisedFunction,
						"emitted_function":     functionName,
						"section_type":         sectionType,
					},
				}}, logger)
		}
	} else {
		ident, err = resolveStorageIdentity(ctx, params.DB, functionName, requesterSiteID, requesterDomain, logger)
		if err != nil {
			return nil, fmt.Errorf("storage identity resolution failed for %q: %w", functionName, err)
		}
		reportAmbiguousUnpinnedRegeneration(ctx, params, ident, functionName, sectionType, requesterSiteID, logger)
	}
	if ident.Diverted {
		logger.Info("store_generated_component: foreign collision — write diverted to site-scoped identity",
			zap.String("requested_function", ident.DivertedFromFunc),
			zap.String("final_function", ident.FunctionName),
			zap.String("incumbent_id", ident.DivertedFromID),
			zap.Strings("incumbent_dependent_domains", ident.ForeignDomains))
		LogActionFindings(ctx, params, requesterSiteID, "",
			"store_generated_component", []agenterrors.Finding{{
				ErrorCode: "COMPONENT_COLLISION_DIVERTED",
				Severity:  "warning",
				Message: fmt.Sprintf("function %q is held by component %s, depended on by %s — write diverted to new base component %q (section_type %q) instead of regenerating another site's row",
					ident.DivertedFromFunc, ident.DivertedFromID,
					strings.Join(ident.ForeignDomains, ", "), ident.FunctionName, sectionType),
				Context: map[string]interface{}{
					"requested_function": ident.DivertedFromFunc,
					"final_function":     ident.FunctionName,
					"incumbent_id":       ident.DivertedFromID,
					"section_type":       sectionType,
				},
			}}, logger)
	}
	if ident.DivertBlocked != "" {
		logger.Warn("store_generated_component: foreign collision seen but not divertible",
			zap.String("function", functionName),
			zap.String("reason", ident.DivertBlocked))
		LogActionFindings(ctx, params, requesterSiteID, "",
			"store_generated_component", []agenterrors.Finding{{
				ErrorCode: "COMPONENT_COLLISION_DIVERT_BLOCKED",
				Severity:  "warning",
				Message:   ident.DivertBlocked,
				Context:   map[string]interface{}{"function": functionName, "section_type": sectionType},
			}}, logger)
	}
	functionName = ident.FunctionName

	// Separate inline <script> blocks into js_content.
	// The template keeps a <script src="/tools/assets/{function}.js"> reference.
	// Components without inline JS are unaffected (jsContent will be empty).
	var jsContent string
	htmlTemplate, jsContent = separateInlineJS(htmlTemplate, functionName)

	if jsContent != "" {
		logger.Info("store_generated_component: extracted inline JS to js_content",
			zap.String("function", functionName),
			zap.Int("js_length", len(jsContent)),
			zap.Int("template_length", len(htmlTemplate)))
	}

	// ── Validate template quality ───────────────────────────────────────
	// Reject templates that are clearly broken — CSS-only output, truncated
	// by token limit, or missing input_schema. Without these checks, broken
	// components enter the DB and silently cause every page using this
	// section type to render empty/CSS-only content.

	// Check 1: Template must contain HTML structure (section or div),
	// not just a <style> block.
	templateLower := strings.ToLower(htmlTemplate)
	if !strings.Contains(templateLower, "<section") && !strings.Contains(templateLower, "<div") {
		return nil, fmt.Errorf(
			"generated template for %q has no HTML structure (<section> or <div>) — likely CSS-only or truncated output",
			sectionType)
	}

	// Check 2: Unclosed <style> tags indicate token-limit truncation.
	// Markup-context count (bugs_open/303): '<style' inside a script body or
	// comment is a mention, not an open tag. Unlike the five-pair gate further
	// down, this check has no 100-char floor — kept as-is.
	for _, tb := range content.StructuralTagCounts(htmlTemplate) {
		if tb.Open == "<style" && tb.Opens > tb.Closes {
			return nil, fmt.Errorf(
				"generated template for %q has %d unclosed <style> tag(s) in markup context — likely truncated by token limit",
				sectionType, tb.Opens-tb.Closes)
		}
	}

	// Check 3: Empty input_schema means the component has no content fields.
	// It can't accept LLM-generated content, so every page using it will
	// render the raw template with no substitution.
	if inputSchemaJSON == "{}" || inputSchemaJSON == "" || inputSchemaJSON == `{"fields":{}}` {
		logger.Warn("store_generated_component: empty input_schema — component has no content fields",
			zap.String("section_type", sectionType),
			zap.String("function", functionName))
		return nil, fmt.Errorf(
			"generated template for %q has empty input_schema — no content fields defined, page builds would produce empty sections",
			sectionType)
	}

	// Check 4: the input_schema must be the house dialect, {"fields": {...}}.
	// content_components refuses the retired JSON-Schema dialect (a top-level
	// "properties" key) by CHECK constraint chk_input_schema_no_legacy_dialect
	// (migration 437, bugs_closed/265), so without this check a JSON-Schema-shaped
	// generation would die on SQLSTATE 23514 at the INSERT/UPDATE below — after
	// deriveRenderMode had read `fields`, found none, and called it "template".
	// Refuse here instead, with the message that names the fix. Same predicate as
	// the constraint, on purpose: refuse exactly the set the table refuses.
	{
		var schemaMap map[string]interface{}
		if err := json.Unmarshal([]byte(inputSchemaJSON), &schemaMap); err == nil &&
			datahelpers.IsLegacyInputSchemaDialect(schemaMap) {
			logger.Warn("store_generated_component: generated input_schema is the retired JSON-Schema dialect — refused before write",
				zap.String("section_type", sectionType),
				zap.String("function", functionName))
			return nil, fmt.Errorf(
				"generated input_schema for %q uses the retired JSON-Schema dialect (top-level \"properties\"); content_components refuses it (chk_input_schema_no_legacy_dialect) — emit the house dialect {\"fields\": {\"<name>\": {\"source\",\"required\",\"type\",...}}}",
				sectionType)
		}
	}

	// Build suitable_site_types from the site_type that triggered the creation
	suitableSiteTypes := []string{}
	if siteType != "" {
		suitableSiteTypes = append(suitableSiteTypes, siteType)
	}
	suitableSiteTypesJSON, _ := json.Marshal(suitableSiteTypes)

	// Build suitable_page_types from page context if available
	suitablePageTypes := []string{}
	if pageContext != "" {
		suitablePageTypes = append(suitablePageTypes, pageContext)
	}
	suitablePageTypesJSON, _ := json.Marshal(suitablePageTypes)

	// Build display name from function
	displayName := datahelpers.FunctionToDisplayName(functionName)

	// Determine category from site_type
	category := "custom"
	if siteType != "" {
		category = siteType
	}

	logger.Info("store_generated_component: storing component",
		zap.String("section_type", sectionType),
		zap.String("function", functionName),
		zap.String("category", category),
		zap.Int("template_length", len(htmlTemplate)),
		zap.Bool("is_dark", isDark))

	// ── Regeneration vs creation ────────────────────────────────────────
	// Decided by resolveStorageIdentity above (component_storage_identity.go
	// carries the lookup, its is_active-ordering history note, and the
	// site-aware diversion rule). If regenerating, the new template will
	// REPLACE the existing one, with the old state snapshotted to
	// component_versions first.
	//
	// Either way, Layer 1 validation below MUST pass before we touch the
	// DB. An existing broken component is NOT grounds to silently accept
	// another broken template.
	existingID := ident.ExistingID
	existingHTML := ident.ExistingHTML
	existingSchema := ident.ExistingSchema
	existingJS := ident.ExistingJS
	isRegeneration := ident.IsRegeneration
	var existingVersion int // max version_number from component_versions; 0 if none

	if isRegeneration {
		// Find the latest version_number so the snapshot we write gets
		// MAX+1. Unique index (component_id, version_number) enforces
		// monotonic numbering.
		if err := params.DB.QueryRowContext(ctx, `
			SELECT COALESCE(MAX(version_number), 0)
			FROM component_versions
			WHERE component_id = $1::uuid
		`, existingID).Scan(&existingVersion); err != nil {
			// Non-fatal: if the query fails we default to 0 and the
			// snapshot INSERT will use version_number=1. Log for visibility.
			logger.Warn("store_generated_component: could not read current max version_number, defaulting to 0",
				zap.String("component_id", existingID),
				zap.Error(err))
			existingVersion = 0
		}
		logger.Info("store_generated_component: regeneration — existing component found",
			zap.String("function", functionName),
			zap.String("existing_id", existingID),
			zap.Int("current_max_version", existingVersion))
	}

	// ── Layer 1 pre-store validation ────────────────────────────────────
	// Before inserting, run the same scoring logic used after insert. This
	// catches templates that pass the structural Check 1/2/3 above but have
	// deeper problems: zero template variables despite a populated schema,
	// schema-template field mismatch, malformed placeholder syntax.
	// Rejecting here prevents broken components entering the DB and being
	// used on pages before the quality auditor catches them.
	//
	// The scoring is a pure function (no DB access) so running it twice —
	// once here, once after INSERT — has no side effects on the first call.
	schemaJSONStr := string(inputSchemaJSON)
	preStoreScore := scoreComponent("", functionName, htmlTemplate, schemaJSONStr, "section")

	// Reject on structural problems that make the component unusable.
	// These are the same conditions that produced quality_score=30 on
	// provocation-feed and archetype-combinations (2026-04-17).
	blockingIssues := []string{}
	if !preStoreScore.TemplateClosed {
		blockingIssues = append(blockingIssues, "template not closed properly")
	}
	if !preStoreScore.HasDataComponent {
		blockingIssues = append(blockingIssues, "missing data-component attribute")
	}
	if preStoreScore.TemplateVariableCount == 0 && preStoreScore.SchemaFieldCount > 0 {
		blockingIssues = append(blockingIssues, fmt.Sprintf(
			"template has 0 {{.var}} placeholders but schema declares %d fields — content would be unreachable",
			preStoreScore.SchemaFieldCount))
	}
	if !preStoreScore.SchemaTemplateSynced && preStoreScore.TemplateVariableCount > 0 {
		orphans, unknownVars, _ := classifySyncIssues(preStoreScore.QualityIssues)
		if len(unknownVars) > 0 {
			// A template var with no schema source renders empty on every page —
			// a real defect. Block, and name EVERY offender (both directions).
			blockingIssues = append(blockingIssues, describeSchemaTemplateMismatch(preStoreScore))
		} else {
			// Orphan-only mismatch (owner ruling 2026-08-25, bugs_open/345): a
			// schema field with no {{.placeholder}} renders nothing whether it is
			// kept or dropped, so DROP it and store rather than refuse the whole
			// component over a harmless surplus. The recorder already scored this
			// class `warning`; the gate used to enforce it as blocking, and 9
			// components died orphan-only over fields the framework would have
			// ignored (against 1 completed) — measured 2026-08-25.
			//
			// Drop ONLY orphans absent from the incumbent schema. An orphan that
			// is also an existing field is left in place: dropping it would remove
			// a name a live page's content_data may be keyed on, which is the
			// stranding guard's concern below, not this branch's. On a creation
			// there is no incumbent, so every orphan is new and droppable — which
			// is the entire measured firing class (all needs_new_component births).
			incumbent := map[string]bool{}
			if isRegeneration {
				incumbent = schemaFieldSet(existingSchema)
			}
			droppable := make([]string, 0, len(orphans))
			for _, f := range orphans {
				if !incumbent[f] {
					droppable = append(droppable, f)
				}
			}
			if len(droppable) > 0 {
				if reduced, derr := dropSchemaFields(inputSchemaJSON, droppable); derr == nil {
					// Keep the stored artefact and every downstream check (source
					// vocabulary, stranding, the INSERT/UPDATE) reading the SAME
					// reduced schema. inputSchemaJSON is what is stored;
					// schemaJSONStr is what the checks below read.
					inputSchemaJSON = reduced
					schemaJSONStr = reduced
					logger.Info("store_generated_component: orphan-only schema/template mismatch — dropped unrendered field(s) and storing",
						zap.String("function", functionName),
						zap.String("section_type", sectionType),
						zap.Strings("dropped_orphan_fields", droppable),
						zap.Int("incumbent_orphans_kept", len(orphans)-len(droppable)))
				} else {
					// If the drop itself fails, fall back to the old behaviour
					// (block) rather than store a schema we could not reduce.
					logger.Warn("store_generated_component: orphan-only mismatch but field drop failed — blocking as before",
						zap.String("function", functionName), zap.Error(derr))
					blockingIssues = append(blockingIssues, describeSchemaTemplateMismatch(preStoreScore))
				}
			} else {
				// Every orphan is an incumbent field: nothing safe to drop, but
				// still not a blocking defect — store as-is (the fields render
				// nothing and the stranding guard below owns the incumbent set).
				logger.Info("store_generated_component: orphan-only schema/template mismatch, all orphans are incumbent fields — storing without change",
					zap.String("function", functionName),
					zap.String("section_type", sectionType),
					zap.Strings("orphan_fields", orphans))
			}
		}
	}

	// Substantive template with no placeholders at all. Catches the case
	// the prior conditions miss: both template AND schema are populated
	// with HTML/CSS but no {{placeholder "..."}} tokens exist. Such a
	// template can render only static markup — no LLM content can be
	// injected, no static fallbacks can be substituted. The threshold
	// (500 chars) excludes legitimately tiny utility components like
	// dividers or spacers.
	const substantiveTemplateThreshold = 500
	if preStoreScore.TemplateVariableCount == 0 && len(htmlTemplate) > substantiveTemplateThreshold {
		blockingIssues = append(blockingIssues, fmt.Sprintf(
			"template is %d chars but has 0 {{placeholder \"...\"}} tokens — no content path exists",
			len(htmlTemplate)))
	}

	// Literal "<no value>" strings in the template are Go text/template
	// render artifacts: the template was rendered against an empty data
	// context (or with default missing-key handling), the unresolved
	// variables produced "<no value>", and that output was stored back
	// as the template. Such a template is permanently broken — the
	// placeholders are gone and can't be restored without regeneration.
	if strings.Contains(htmlTemplate, "<no value>") {
		blockingIssues = append(blockingIssues, fmt.Sprintf(
			"template contains %d '<no value>' artifacts — Go template render output mistakenly stored as source",
			strings.Count(htmlTemplate, "<no value>")))
	}

	// Unterminated structural tag = a generation cut mid-stream (bugs_open/021
	// INSTANCE 1 / bugs_open/046). preStoreScore.TemplateClosed above only
	// balances <section> tags, so a section whose <script>/<style> is cut but
	// whose <section> closes upstream of the cut passes it — which is how 4 of
	// bugs_open/046's 8 casualties (and its one 'section' casualty) slipped
	// through. hasUnbalancedStructuralTags checks all five pairs, counting only
	// markup-context tags (bugs_open/303 — a tag mentioned in a script body or
	// comment is not an unterminated tag). It is the tag-imbalance signal ALONE
	// (NOT the tool path's ends-mid-token check, which over-fires on non-tool
	// components) — calibrated to 0 over-fire fleet-wide, re-run 2026-08-18.
	// The <100-char skip mirrors sectionTemplateValid's stub tolerance; a
	// truncated generation is long, not a tiny intentional stub.
	if len(htmlTemplate) >= 100 && hasUnbalancedStructuralTags(htmlTemplate) {
		blockingIssues = append(blockingIssues,
			"template leaves a structural tag ("+strings.Join(content.UnbalancedStructuralTags(htmlTemplate), ", ")+
				") unterminated in markup context — the signature of a generation cut mid-stream")
	}

	// A template must not SUBSTITUTE A BUSINESS FACT for an absent datum
	// (RFC_009 option B, owner ruling 2026-08-03). contact-info shipped
	// {{else}}+1234567890{{end}} and {{else}}Monday – Friday, 9am – 6pm{{end}} from
	// the library's birth; eight live commercial sites served the invented hours and
	// one served the invented phone as a tel: link, every render "succeeding"
	// (bugs_open/140). A fabricated fact is styled identically to a real one, so
	// nothing downstream — human or machine — can tell them apart.
	//
	// It refuses ONLY the fact-shaped fallback. A LABEL default ({{else}}Read
	// more{{end}}) is legitimate and the library is full of them; the distinction is
	// input_schema's own (on_missing:"skip_field" for facts, an explicit "fallback"
	// for labels), and component_fallback_guard.go's header carries the reasoning.
	//
	// Calibrated before shipping, per the standing instruction in
	// component_write_guard.go's header: 0 findings across all 347 recorded writes
	// (every component_versions row + every content_components row), so it would have
	// refused nothing in the platform's entire history — while still refusing the
	// pre-fix contact-info template, recovered from migration 287's before-image and
	// pinned as a test. Both halves matter: zero-on-the-corpus is also what an inert
	// guard scores.
	//
	// NOTE for whoever hits a refusal on a REGENERATION: a component that already
	// fabricates cannot be repaired by regenerating it into another fabrication, and
	// that is deliberate. It CAN be repaired by a regeneration that gates the
	// element, or by a migration (which is how 140 was fixed). No component is
	// currently trapped by this — 0 active components fabricate as of 2026-08-03.
	if issue := fabricatedFallbackIssue(htmlTemplate); issue != "" {
		blockingIssues = append(blockingIssues, issue)
	}

	// A declared source must be one the platform can resolve (bugs_open/309:
	// blog-listing_pre_037 shipped six required URLs sourced from a site_specs
	// aspect that has never existed on any site, and served an article index
	// whose every card was silently link-less). The aspect half needs the live
	// aspect vocabulary; a failed read skips ONLY that half (fail open — a
	// transient DB error must not block all component generation) — and that
	// skip is recorded DURABLY, not just logged: a sustained outage across
	// store windows would otherwise re-admit phantom-aspect components with
	// the same silent-loss shape this guard exists to close (council
	// fdb032c6, bug_historian). component_source_guard.go carries the
	// reasoning and the calibration.
	knownSpecAspects, aspectsErr := LoadKnownSpecAspects(ctx, params.DB)
	if aspectsErr != nil {
		logger.Warn("store_generated_component: could not load site_specs aspect vocabulary — site_specs source validation skipped for this store",
			zap.Error(aspectsErr))
		LogActionFindings(ctx, params,
			datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.site_id"), "",
			"store_generated_component", []agenterrors.Finding{{
				ErrorCode: "SOURCE_GUARD_ASPECT_SET_UNAVAILABLE",
				Severity:  "warning",
				Message: fmt.Sprintf("site_specs aspect vocabulary unread (%v) — component %q stored with the phantom-aspect half of the source guard SKIPPED; prefix and query-name halves still ran",
					aspectsErr, functionName),
				Context: map[string]interface{}{
					"function":     functionName,
					"section_type": sectionType,
				},
			}}, logger)
	}
	blockingIssues = append(blockingIssues, SourceVocabularyIssues(schemaJSONStr, knownSpecAspects)...)

	// Regeneration must not break the field-name contract that existing
	// dependents' content_data is keyed on. content_data is written to
	// match the component's input_schema field names at build time;
	// renaming or removing a retained field strands that stored content —
	// the renamed placeholder no longer matches the stored key, and
	// RenderTemplate silently strips the unmatched placeholder to "", so
	// the section renders empty with no error (this is the fdd92ad4
	// system-stats failure: e.g. stat_1_number→stat1_value, eyebrow→
	// eyebrow_label renamed in place while dependents' content_data stayed
	// on the old keys). Adding new fields is fine; dropping/renaming an
	// existing one is the damage, so block it here rather than overwrite
	// the shared row and silently empty every dependent. Intentional
	// field-set changes to a shared component must go through a deliberate
	// migration, not an LLM regeneration side effect.
	//
	// This compares old-schema fields to new-schema fields as a proxy for
	// "what dependents have" (content_data is written to match the schema).
	// If exactness is ever needed, swap to querying the affected
	// page_components for the union of their content_data keys.
	if isRegeneration {
		oldFields := schemaFieldSet(existingSchema)
		newFields := schemaFieldSet(schemaJSONStr)
		var stranded []string
		for name := range oldFields {
			if !newFields[name] {
				stranded = append(stranded, name)
			}
		}
		if len(stranded) > 0 {
			sort.Strings(stranded) // deterministic message ordering
			blockingIssues = append(blockingIssues, fmt.Sprintf(
				"regeneration removes/renames %d existing schema field(s) (%s) that dependents' content_data is keyed on — overwriting would strand stored content and render those sections empty; preserve these field names or migrate dependents explicitly",
				len(stranded), strings.Join(stranded, ", ")))
		}
	}

	if len(blockingIssues) > 0 {
		// Persist a structured rejection record to agent_error_log. This
		// makes validation failures queryable across the system — we can
		// run analytics on "which fields does the LLM most often forget
		// to declare?" or "which functions repeatedly fail Direction 2?"
		// without trawling kubectl logs.
		//
		// The act of writing here is best-effort (best handled inside the
		// helper). The next return is the actual rejection.
		recordValidationRejection(
			ctx, params.DB, logger, params,
			functionName, sectionType,
			htmlTemplate, schemaJSONStr,
			preStoreScore, blockingIssues,
		)

		// section_type self-heal on the REJECTION path (bugs_open/337).
		//
		// The UPDATE below carries the same COALESCE, but it only ever runs on
		// a SUCCESSFUL store — so the repair was gated behind the success that
		// its own absence prevents. A row whose section_type is NULL is
		// invisible to BOTH section_type readers: the selector, which is how a
		// page finds a component at all, and load_existing_component, which is
		// how the writer learns the field contract this very rejection is
		// enforcing. Blind writer -> stranded fields -> rejection -> no heal ->
		// blind writer. Measured 2026-08-22: one such component was refused 70
		// times without a single success, while a sibling with four generically
		// named fields escaped on the second attempt by reproducing them by
		// chance.
		//
		// ⚠ GATED ON is_active, AND THAT GATE IS THE WHOLE SAFETY ARGUMENT.
		// Healing section_type makes a component SELECTABLE, and migration 036
		// deactivates broken components precisely so that pages stop choosing
		// them. Healing an inactive row here would take a component that just
		// failed the gate and offer it to page planning — strictly worse than
		// the invisibility being repaired. So this only ever fills a metadata
		// gap on a row the estate already treats as live; it never revives one.
		// It writes a single NULL column, never the template, never is_active.
		healRejectedComponentSectionType(ctx, params.DB, logger, existingID, isRegeneration, sectionType)

		logger.Warn("store_generated_component: rejecting low-quality template",
			zap.String("function", functionName),
			zap.String("section_type", sectionType),
			zap.Int("pre_store_score", preStoreScore.QualityScore),
			zap.Int("template_variable_count", preStoreScore.TemplateVariableCount),
			zap.Int("schema_field_count", preStoreScore.SchemaFieldCount),
			zap.Strings("blocking_issues", blockingIssues),
			zap.Strings("all_issues", preStoreScore.QualityIssues),
		)
		return nil, fmt.Errorf(
			"generated template for %q rejected by pre-store validation: %s",
			sectionType, strings.Join(blockingIssues, "; "))
	}

	logger.Info("store_generated_component: pre-store validation passed",
		zap.String("function", functionName),
		zap.Int("pre_store_score", preStoreScore.QualityScore),
		zap.Int("template_variable_count", preStoreScore.TemplateVariableCount),
		zap.Int("schema_field_count", preStoreScore.SchemaFieldCount),
	)

	// ── Write to DB: UPDATE (regeneration) or INSERT (creation) ─────────
	// Both paths end with scoring + markPagesForRebuild so those stay below.
	var componentID string
	var status string
	var regenPagesMarked int64
	var newVersion int // populated on regeneration; 0 for creation

	if isRegeneration {
		// Snapshot current state to component_versions BEFORE the UPDATE.
		// Best-effort: a failed snapshot logs Warn but does not block the
		// UPDATE — losing history is recoverable; leaving a broken template
		// in place is not.
		newVersion = existingVersion + 1
		snapshotErr := snapshotComponentVersion(
			ctx, params.DB, existingID, newVersion,
			existingHTML, existingSchema,
			nullStringToGo(existingJS),
			"Regenerated by component-creator",
			"component-creator:regen",
			workItemSource,
			logger,
		)
		if snapshotErr != nil {
			logger.Warn("store_generated_component: version snapshot failed, continuing with UPDATE",
				zap.String("component_id", existingID),
				zap.Int("intended_version", newVersion),
				zap.Error(snapshotErr))
			// newVersion still advances — even if we couldn't write the
			// snapshot, we don't want to overwrite version N later and
			// create a gap. The post-UPDATE return reflects what we tried.
		}

		// UPDATE in place: preserves component_id so all foreign key
		// references (page_components, site_components, link_registry,
		// etc.) keep resolving without any relink step.
		// section_type self-heal (bugs_open/311): a manually-seeded row can
		// carry function without section_type, leaving it invisible to the
		// selector for ever. Every successful regeneration repairs that —
		// COALESCE so an already-set value is never overwritten (the two
		// columns legitimately differ on 26 live rows).
		result, err := params.DB.ExecContext(ctx, `
			UPDATE content_components
			SET html_template   = $1,
			    input_schema    = $2::jsonb,
			    js_content      = $3,
			    is_dark_section = $4,
			    render_mode     = $5,
			    section_type    = COALESCE(section_type, $6),
			    is_active       = true,
			    updated_at      = NOW()
			WHERE id = $7::uuid
		`,
			htmlTemplate,
			inputSchemaJSON,
			nullIfEmpty(jsContent),
			isDark,
			deriveRenderMode(inputSchemaJSON), // derived from schema, not hardcoded
			nullIfEmpty(sectionType),
			existingID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update existing component during regeneration: %w", err)
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected != 1 {
			return nil, fmt.Errorf("regeneration UPDATE affected %d rows (expected 1) for component %s",
				rowsAffected, existingID)
		}
		componentID = existingID
		status = "regenerated"

		// Mark dependent page_components pending so the rerender pipeline
		// rebuilds them against the new template. We do NOT overwrite
		// rendered_html — that field holds the last-good render per page
		// and needs per-page variable substitution to regenerate correctly.
		var affectedSiteIDs []string
		regenPagesMarked, affectedSiteIDs = markPagesPendingRebuild(ctx, params.DB, existingID, logger)

		// Raise one needs_rerender work item per affected site so the
		// rerender-pages handler actually regenerates. Without this the
		// build_status=pending flag is informational only — nothing
		// downstream scans page_components for pending rows.
		//
		// The dedup index (site_id, item_key) with item_key scoped by
		// component_id prevents duplicates if the same regen runs twice.
		rerenderItemsCreated := 0
		for _, siteID := range affectedSiteIDs {
			if created := createRerenderWorkItem(
				ctx, params.DB, siteID, existingID, functionName, workItemSource, logger,
			); created {
				rerenderItemsCreated++
			}
		}

		logger.Info("store_generated_component: component regenerated",
			zap.String("component_id", componentID),
			zap.String("function", functionName),
			zap.Int("previous_version", existingVersion),
			zap.Int("new_version", newVersion),
			zap.Int64("pages_marked_rebuild", regenPagesMarked),
			zap.Int("affected_sites", len(affectedSiteIDs)),
			zap.Int("rerender_items_created", rerenderItemsCreated))
	} else {
		// Creation path.
		// usage_count is deliberately NOT in this column list (bugs_closed/378). It was the
		// last writer of a column nothing reads any more: "how proven is this component" is
		// derived from page_components at read time (ComponentUsageSitesSQL). The column still
		// exists and carries DEFAULT 0, so omitting it is behaviour-identical today — and it
		// is omitted precisely so the pending DROP cannot break this statement. That DROP is
		// docs/agent_docs/sql_for_agents/610_content_components_drop_dead_usage_count_HOLD.sql
		// — held, applied by hand, and it names this exact INSERT as its precondition.
		// (609 is the COMMENT-only migration and is already applied; naming it here instead
		//  of 610 was wrong in the first version of this comment, caught by the bugs_open/388
		//  lane. A bare migration NUMBER is ambiguous in this repo — the filename is not.)
		// ⚠ DO NOT re-add usage_count to this column list. Naming a column in an INSERT is
		// enough to make every component creation fail the moment 610 runs.
		err = params.DB.QueryRowContext(ctx, `
			INSERT INTO content_components (
				name, display_name, function, category, component_level,
				section_type, suitable_site_types, suitable_page_types,
				description, html_template, js_content, input_schema,
				is_dark_section, render_mode, created_from, is_active,
				avg_quality_score,
				semantic_tags
			) VALUES (
				$1, $2, $3, $4, 'section',
				$5, $6::jsonb, $7::jsonb,
				$8, $9, $10, $11::jsonb,
				$12, $13, 'generated', true,
				NULL,
				$14::jsonb
			)
			RETURNING id::text
		`,
			functionName,                      // $1 name
			displayName,                       // $2 display_name
			functionName,                      // $3 function
			category,                          // $4 category
			sectionType,                       // $5 section_type
			string(suitableSiteTypesJSON),     // $6 suitable_site_types
			string(suitablePageTypesJSON),     // $7 suitable_page_types
			description,                       // $8 description
			htmlTemplate,                      // $9 html_template (JS extracted)
			nullIfEmpty(jsContent),            // $10 js_content (NULL if no JS)
			inputSchemaJSON,                   // $11 input_schema
			isDark,                            // $12 is_dark_section
			deriveRenderMode(inputSchemaJSON), // $13 render_mode (derived from schema, not hardcoded)
			datahelpers.BuildSemanticTags(sectionType, siteType), // $14 semantic_tags
		).Scan(&componentID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert component: %w", err)
		}
		status = "created"

		// A CREATION THAT NOBODY ADVISED, ON A section_type THAT ALREADY HAS ONE
		// (bugs_open/388). This is the SILENT half of the defect: the loud half
		// is a wrong-cause refusal, and the field-contract guard above is
		// vacuous here because isRegeneration is false — so a parallel duplicate
		// is born with no error, no work item and no trace.
		//
		// It is unrepresentable on the pinned path, so this can only fire on the
		// residual: an un-wired workflow, an advisory that failed open, or an
		// error_step that routed around it. That is exactly what makes it worth
		// recording — it counts the fail-open rather than the fixed case.
		//
		// ⚠ DELIBERATELY A DIFFERENT QUESTION FROM THE ONE THAT RESOLVED THE
		// IDENTITY. The identity was resolved by id (or by function); this asks
		// section_type. bugs_open/324 shipped 32 dangling rows because its
		// completeness check re-grepped the renamer's own patterns — a check
		// that asks the resolver's question can only ever agree with it.
		//
		// Diverted writes are excluded: a site-scoped twin alongside the base
		// row is bugs_open/311 working as designed, not a silent fork, and it
		// already records COMPONENT_COLLISION_DIVERTED above.
		if advisedID == "" && !ident.Diverted && sectionType != "" {
			var incumbents int
			var incumbentFns sql.NullString
			censusErr := params.DB.QueryRowContext(ctx, `
				SELECT count(*), string_agg(function, ', ' ORDER BY function)
				FROM content_components
				WHERE section_type = $1
				  AND forked_from IS NULL
				  AND is_active = true
				  AND component_level = 'section'
				  AND id <> $2::uuid
			`, sectionType, componentID).Scan(&incumbents, &incumbentFns)
			switch {
			case censusErr != nil:
				// Best-effort by design: the row is already written and a census
				// failure must not undo it or fail the build.
				logger.Warn("store_generated_component: parallel-birth census unreadable",
					zap.String("section_type", sectionType), zap.Error(censusErr))
			case incumbents > 0:
				logger.Warn("store_generated_component: created a second component for a section_type that already had one",
					zap.String("section_type", sectionType),
					zap.String("function", functionName),
					zap.String("component_id", componentID),
					zap.Int("incumbents", incumbents),
					zap.String("incumbent_functions", nullStringToGo(incumbentFns)))
				LogActionFindings(ctx, params, requesterSiteID, "",
					"store_generated_component", []agenterrors.Finding{{
						ErrorCode: "COMPONENT_PARALLEL_SECTION_BIRTH",
						Severity:  "warning",
						Message: fmt.Sprintf("created component %q (%s) for section_type %q while %d active non-forked section component(s) already served it (%s), and no advised identity was supplied — the field-contract guard cannot see this case, so the duplicate is silent",
							functionName, componentID, sectionType, incumbents, nullStringToGo(incumbentFns)),
						Context: map[string]interface{}{
							"component_id":        componentID,
							"function":            functionName,
							"section_type":        sectionType,
							"incumbents":          incumbents,
							"incumbent_functions": nullStringToGo(incumbentFns),
						},
					}}, logger)
			}
		}

		// Snapshot version 1 so history is complete from creation onward.
		// Best-effort — a snapshot failure here doesn't undo the INSERT.
		if err := snapshotComponentVersion(
			ctx, params.DB, componentID, 1,
			htmlTemplate, inputSchemaJSON,
			jsContent,
			"Initial version — created by component-creator",
			"component-creator:create",
			workItemSource,
			logger,
		); err != nil {
			logger.Warn("store_generated_component: initial version snapshot failed, continuing",
				zap.String("component_id", componentID),
				zap.Error(err))
		} else {
			newVersion = 1
		}

		logger.Info("store_generated_component: component created",
			zap.String("component_id", componentID),
			zap.String("function", functionName),
			zap.String("section_type", sectionType))
	}

	// Score the resulting row (both paths). Persists to content_components.
	qualityResult := ScoreAndPersistComponent(
		ctx, params.DB,
		componentID, functionName, htmlTemplate, schemaJSONStr, "section",
		logger,
	)

	// Mark pages that were waiting for this section_type as needs_rebuild.
	// When plan_sections can't find a component, the page stays deployed
	// with a gap. Now that the component exists (or has been regenerated),
	// those pages should rebuild.
	markPagesForRebuild(ctx, params.DB, sectionType, logger)

	response := map[string]interface{}{
		"component_id":   componentID,
		"function":       functionName,
		"section_type":   sectionType,
		"display_name":   displayName,
		"category":       category,
		"status":         status,
		"template_size":  len(htmlTemplate),
		"has_js":         jsContent != "",
		"js_size":        len(jsContent),
		"quality_score":  qualityResult.QualityScore,
		"quality_issues": qualityResult.QualityIssues,
	}
	if isRegeneration {
		response["previous_version"] = existingVersion
		response["new_version"] = newVersion
		response["pages_marked_rebuild"] = regenPagesMarked
	} else if newVersion > 0 {
		response["new_version"] = newVersion
	}
	if ident.Diverted {
		response["diverted_from_component_id"] = ident.DivertedFromID
		response["requested_function"] = ident.DivertedFromFunc
	}
	return response, nil
}

// reportAmbiguousUnpinnedRegeneration records that an UN-PINNED regeneration
// chose its target by NAME among several rows that share it (bugs_open/388,
// council round 2, bug_historian's advisory objection).
//
// WHY THIS EXISTS SEPARATELY FROM COMPONENT_PARALLEL_SECTION_BIRTH. That code
// instruments the un-pinned CREATE path, where the damage is a duplicate row.
// This is the un-pinned REGENERATE path, where the damage is the opposite shape
// and worse: an existing row is OVERWRITTEN, and which one was decided by
// `ORDER BY is_active DESC, updated_at DESC LIMIT 1` over a name that does not
// identify a row. [MEASURED 2026-08-25] 25 of 330 non-forked rows' `function`
// values carry more than one row — `site-footer` and `site-header` five each,
// spanning component_level 'section' AND 'site'. The winner can therefore change
// with no code change at all, because anything that touches a sibling's
// updated_at reorders the contest.
//
// ⚠ IT IS DELIBERATELY THE QUESTION THE RESOLVER THREW AWAY, NOT THE ONE IT
// ASKED. lookupBaseComponent asks "which row wins" and its LIMIT 1 structurally
// cannot report that there was a contest; this asks "how many were there". That
// distinction is the whole point — bugs_open/324 shipped 32 dangling rows
// because its completeness check re-ran the resolver's own query and could only
// ever agree with it. A count is not a second opinion on the winner; it is the
// fact the winner's selection discards.
//
// Best-effort and never fatal: the row is about to be written either way, and a
// census failure must not fail a store that would otherwise succeed. Silent on
// the unambiguous case, so a row in agent_error_log always means a real contest.
func reportAmbiguousUnpinnedRegeneration(
	ctx context.Context, params ActionParams, ident storageIdentity,
	functionName, sectionType, requesterSiteID string, logger *zap.Logger,
) {
	// Only a regeneration overwrites. A creation is COMPONENT_PARALLEL_SECTION_BIRTH's
	// case, and a diverted write already records COMPONENT_COLLISION_DIVERTED.
	if !ident.IsRegeneration || ident.Diverted || ident.FunctionName == "" {
		return
	}

	var siblings int
	if err := params.DB.QueryRowContext(ctx, `
		SELECT count(*)
		FROM content_components
		WHERE function = $1 AND forked_from IS NULL
	`, ident.FunctionName).Scan(&siblings); err != nil {
		logger.Warn("store_generated_component: sibling census unreadable — cannot say whether this regeneration was ambiguous",
			zap.String("function", ident.FunctionName), zap.Error(err))
		return
	}
	finding, report := ambiguousRegenerationFinding(ident, siblings, functionName, sectionType)
	if !report {
		return
	}

	logger.Warn("store_generated_component: un-pinned regeneration chose among several rows sharing a function name",
		zap.String("function", ident.FunctionName),
		zap.String("chosen_id", ident.ExistingID),
		zap.Int("rows_sharing_function", siblings))

	LogActionFindings(ctx, params, requesterSiteID, "",
		"store_generated_component", []agenterrors.Finding{finding}, logger)
}

// ambiguousRegenerationFinding is the DECISION, split out as a pure function
// with no DB and no logging — and that split is not tidiness, it is the only way
// the negative case can be tested at all.
//
// ⚠ WHY, RECORDED BECAUSE IT COST A VACUOUS TEST FIRST. LogActionFindings is
// best-effort: it swallows a write error. Under sqlmock an UNEXPECTED
// INSERT INTO agent_error_log therefore returns an error that is swallowed, and
// ExpectationsWereMet() still passes. So a test asserting "no finding is
// recorded" by declaring no expectation for it CANNOT FAIL — and the mutation
// that removes the `siblings <= 1` gate passed against exactly such a test,
// which is how this was found. A positive expectation proves a finding fires; no
// arrangement of expectations proves one does not. The predicate has to be
// reachable directly.
//
// The rule itself: report only when MORE THAN ONE non-forked row holds the name.
// A code that fires on every regeneration is a log line, not a finding.
func ambiguousRegenerationFinding(
	ident storageIdentity, siblings int, emittedFunction, sectionType string,
) (agenterrors.Finding, bool) {
	if siblings <= 1 {
		return agenterrors.Finding{}, false
	}
	return agenterrors.Finding{
		ErrorCode: "COMPONENT_UNPINNED_REGENERATION_AMBIGUOUS",
		Severity:  "warning",
		Message: fmt.Sprintf("no advised identity was supplied, so component %s was chosen to be OVERWRITTEN by function %q — which %d non-forked rows share, ordered by is_active then updated_at; the winner can change without any code change, and section_type %q was the request",
			ident.ExistingID, ident.FunctionName, siblings, sectionType),
		Context: map[string]interface{}{
			"chosen_component_id":   ident.ExistingID,
			"function":              ident.FunctionName,
			"emitted_function":      emittedFunction,
			"rows_sharing_function": siblings,
			"section_type":          sectionType,
		},
	}, true
}

// advisedIdentityPin reads the pre-generation advisory's answer out of
// collected_data: which content_components row this write is FOR, and which
// function name the writer was told to echo.
//
// It is deliberately total and deliberately silent about malformed input. The
// wire is an OPTIONAL-EXPLICIT (`advised_identity?`) config key, and absence is
// its contract — an un-wired workflow, an advisory that failed open, an
// error_step that routed around the load step, and a genuine first creation are
// all legitimate and all arrive here as "" / "". Every one of them means
// "resolve the way you always did", which is the safe default and the reason
// this ships without an ordering constraint (bugs_open/388).
//
// Returning the advised FUNCTION as well as the id is not redundancy: the
// divergence finding must compare what the model emitted against what it was
// TOLD, and ident.FunctionName may since have been site-suffixed by
// bugs_open/311's diversion — comparing against that would report a divergence
// on every diverted write.
func advisedIdentityPin(raw interface{}) (componentID, function string) {
	advised, ok := raw.(map[string]interface{})
	if !ok {
		return "", ""
	}
	if v, ok := advised["component_id"].(string); ok {
		componentID = strings.TrimSpace(v)
	}
	if v, ok := advised["function"].(string); ok {
		function = strings.TrimSpace(v)
	}
	// An id we cannot use is the same as no id: fall back rather than fail. The
	// store's own by-id lookup would reject a non-uuid with a driver error, and
	// an advisory must never be able to break a build.
	if componentID != "" && !isUUIDish(componentID) {
		return "", function
	}
	return componentID, function
}

// isUUIDish is a shape check, not a validator: 8-4-4-4-12 hex with hyphens. It
// exists so a malformed pin degrades to the legacy path instead of reaching
// Postgres as a cast error on a write that had a perfectly good fallback.
func isUUIDish(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			isHex := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
			if !isHex {
				return false
			}
		}
	}
	return true
}

// parseGeneratedTemplate extracts html_template, input_schema, and metadata
// from the LLM output. The LLM is instructed to return JSON with these fields,
// but we handle various formats defensively.
func parseGeneratedTemplate(raw interface{}, sectionType string, logger *zap.Logger) (
	htmlTemplate string, inputSchemaJSON string, functionName string, isDark bool, err error,
) {
	if raw == nil {
		return "", "{}", "", false, fmt.Errorf("generated template is nil")
	}

	// The LLM output might be:
	// 1. A map with "result" containing the JSON string (from execute_llm_prompt)
	// 2. A map with html_template/input_schema directly
	// 3. A string containing JSON (possibly wrapped in markdown code blocks)

	var data map[string]interface{}

	switch v := raw.(type) {
	case map[string]interface{}:
		// Check for execute_llm_prompt's "result" wrapper
		if result, ok := v["result"]; ok {
			switch r := result.(type) {
			case string:
				data = parseJSONStringToMap(r, logger)
			case map[string]interface{}:
				data = r
			}
		} else {
			data = v
		}
	case string:
		data = parseJSONStringToMap(v, logger)
	default:
		return "", "{}", "", false, fmt.Errorf("unexpected type for generated template: %T", raw)
	}

	// Extract html_template
	if ht, ok := data["html_template"].(string); ok {
		htmlTemplate = strings.TrimSpace(ht)
	}

	// Safety check: if htmlTemplate still looks like a JSON blob wrapping an
	// html_template field, extract the actual HTML. This catches cases where
	// json.Unmarshal failed and the fallback stored the entire JSON string.
	htmlTemplate = unwrapJSONBlobIfNeeded(htmlTemplate, logger)

	// Extract input_schema
	inputSchemaJSON = "{}"
	if schema, ok := data["input_schema"]; ok {
		if schemaMap, ok := schema.(map[string]interface{}); ok {
			schemaBytes, _ := json.Marshal(schemaMap)
			inputSchemaJSON = string(schemaBytes)
		} else if schemaStr, ok := schema.(string); ok {
			inputSchemaJSON = schemaStr
		}
	}

	// Extract or derive function name
	if fn, ok := data["function"].(string); ok && fn != "" {
		functionName = fn
	} else {
		// Derive from section_type — prefix with a category hint
		functionName = sectionType
	}

	// Validate kebab-case
	functionName = datahelpers.NormaliseToKebab(functionName)

	// Extract is_dark_section
	if dark, ok := data["is_dark_section"].(bool); ok {
		isDark = dark
	}

	return htmlTemplate, inputSchemaJSON, functionName, isDark, nil
}

// parseJSONStringToMap takes a raw string (possibly with markdown code blocks)
// and tries to parse it as a JSON map. If standard parsing fails, it falls back
// to field-level extraction from broken JSON. If the string is not JSON at all,
// it treats it as raw HTML.
//
// Uses datahelpers.StripCodeFences (shared with content_search, create_tool, etc.)
// and datahelpers.SafeUnmarshalString for safe parsing.
func parseJSONStringToMap(s string, logger *zap.Logger) map[string]interface{} {
	// Strip markdown code blocks using the shared helper from datahelpers
	cleaned := datahelpers.StripCodeFences(s)

	// Try standard JSON parse using the shared SafeUnmarshalString
	var data map[string]interface{}
	if datahelpers.SafeUnmarshalString(cleaned, &data) {
		logger.Info("store_generated_component: parsed LLM output as JSON",
			zap.Int("fields", len(data)))
		return data
	}

	// JSON parse failed — try field-level extraction from broken JSON.
	// LLMs often produce JSON with unescaped characters in HTML/SVG content
	// that breaks json.Unmarshal, but the structure is still recoverable.
	if strings.Contains(cleaned, `"html_template"`) {
		logger.Info("store_generated_component: json.Unmarshal failed, attempting field extraction",
			zap.Int("length", len(cleaned)),
			zap.String("first_80", truncateStr(cleaned, 80)))

		result := map[string]interface{}{}

		if ht, ok := extractJSONStringField(cleaned, "html_template"); ok {
			result["html_template"] = ht
			logger.Info("store_generated_component: extracted html_template from broken JSON",
				zap.Int("length", len(ht)),
				zap.String("first_40", truncateStr(ht, 40)))
		}
		if fn, ok := extractJSONStringField(cleaned, "function"); ok {
			result["function"] = fn
		}
		// is_dark_section is a bool, not a string — check with simple contains
		if strings.Contains(cleaned, `"is_dark_section": true`) || strings.Contains(cleaned, `"is_dark_section":true`) {
			result["is_dark_section"] = true
		}

		if _, hasHTML := result["html_template"]; hasHTML {
			return result
		}
	}

	// Not JSON at all — treat as raw HTML
	logger.Info("store_generated_component: treating as raw HTML",
		zap.Int("length", len(cleaned)),
		zap.String("first_50", truncateStr(cleaned, 50)))
	return map[string]interface{}{
		"html_template": cleaned,
	}
}

// extractJSONStringField extracts the value of a string field from a JSON-like
// string, even when json.Unmarshal fails due to unescaped characters elsewhere.
// It manually scans the JSON string value, handling standard escape sequences.
func extractJSONStringField(s, fieldName string) (string, bool) {
	key := `"` + fieldName + `"`
	keyIdx := strings.Index(s, key)
	if keyIdx == -1 {
		return "", false
	}

	// Find the colon after the key
	rest := s[keyIdx+len(key):]
	colonIdx := strings.IndexByte(rest, ':')
	if colonIdx == -1 {
		return "", false
	}
	rest = rest[colonIdx+1:]

	// Skip whitespace
	rest = strings.TrimLeft(rest, " \t\n\r")

	// Must start with a quote
	if len(rest) == 0 || rest[0] != '"' {
		return "", false
	}
	rest = rest[1:] // skip opening quote

	// Scan for the closing unescaped quote, handling escape sequences
	var result strings.Builder
	result.Grow(len(rest))
	i := 0
	for i < len(rest) {
		if rest[i] == '\\' && i+1 < len(rest) {
			// JSON escape sequence
			switch rest[i+1] {
			case '"':
				result.WriteByte('"')
			case '\\':
				result.WriteByte('\\')
			case '/':
				result.WriteByte('/')
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case 'b':
				result.WriteByte('\b')
			case 'f':
				result.WriteByte('\f')
			default:
				// Unknown escape — preserve as-is
				result.WriteByte(rest[i])
				result.WriteByte(rest[i+1])
			}
			i += 2
		} else if rest[i] == '"' {
			// Unescaped closing quote — end of field value
			return result.String(), true
		} else {
			result.WriteByte(rest[i])
			i++
		}
	}

	// No closing quote found. The JSON is severely broken, but we likely have
	// the content up to the end of the string. Return what we collected.
	extracted := result.String()
	if len(extracted) > 0 {
		return strings.TrimSpace(extracted), true
	}
	return "", false
}

// unwrapJSONBlobIfNeeded checks if a string looks like it's still a JSON blob
// wrapping an html_template field, and extracts the actual HTML if so.
// This is a safety net that catches any case where earlier parsing failed.
func unwrapJSONBlobIfNeeded(s string, logger *zap.Logger) string {
	trimmed := strings.TrimSpace(s)
	if !strings.HasPrefix(trimmed, "{") {
		return s // Not a JSON blob
	}
	if !strings.Contains(trimmed, `"html_template"`) {
		return s // Doesn't contain the field
	}

	logger.Info("store_generated_component: unwrapping JSON blob from html_template",
		zap.Int("length", len(trimmed)),
		zap.String("first_60", truncateStr(trimmed, 60)))

	// Try standard JSON parse first using shared SafeUnmarshalString
	var wrapper map[string]interface{}
	if datahelpers.SafeUnmarshalString(trimmed, &wrapper) {
		if ht, ok := wrapper["html_template"].(string); ok && strings.TrimSpace(ht) != "" {
			return strings.TrimSpace(ht)
		}
	}

	// Standard parse failed — use field extraction
	if ht, ok := extractJSONStringField(trimmed, "html_template"); ok && ht != "" {
		return ht
	}

	return s // Couldn't unwrap — return as-is
}

// markPagesForRebuild finds deployed pages whose sections array references
// the given section_type and marks them for rebuild. This closes the loop
// when a component is created after plan_sections already ran and deferred
// the section.
func markPagesForRebuild(ctx context.Context, db *sql.DB, sectionType string, logger *zap.Logger) {
	res, err := db.ExecContext(ctx, `
		UPDATE pages SET build_status = 'needs_rebuild', updated_at = NOW()
		WHERE `+datahelpers.PageWantedLivePredicateFor("")+`
		  AND build_status = 'deployed'
		  AND EXISTS (
		      SELECT 1 FROM jsonb_array_elements_text(sections) sec
		      WHERE sec = $1
		  )
	`, sectionType)
	if err != nil {
		logger.Warn("store_generated_component: failed to mark pages for rebuild",
			zap.String("section_type", sectionType),
			zap.Error(err))
		return
	}
	if rows, _ := res.RowsAffected(); rows > 0 {
		logger.Info("store_generated_component: marked pages for rebuild",
			zap.String("section_type", sectionType),
			zap.Int64("pages_marked", rows))
	}
}

// truncateStr returns the first n characters of s, appending "..." if truncated.
// Named truncateStr to avoid conflict with any future stdlib truncate.
func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// separateInlineJS extracts inline <script> blocks from an HTML template,
// stores them separately, and replaces with a <script src> reference.
//
// Only extracts <script> tags WITHOUT attributes (inline JS).
// Leaves <script src="...">, <script type="module">, etc. untouched.
//
// Multiple inline script blocks are combined into a single JS content string.
// If no inline scripts found, returns the template unchanged and empty jsContent.
func separateInlineJS(htmlTemplate, functionName string) (cleanHTML, jsContent string) {
	// Match <script> tags with no attributes — these contain inline JS.
	// (?s) enables dot-matches-newline.
	re := regexp.MustCompile(`(?s)<script\s*>(.*?)</script>`)

	var jsBlocks []string
	hasInlineJS := false

	cleanHTML = re.ReplaceAllStringFunc(htmlTemplate, func(match string) string {
		// Safety check: skip if somehow a src= tag matched
		trimmed := strings.TrimSpace(match)
		if len(trimmed) > 20 && strings.Contains(trimmed[:20], "src=") {
			return match
		}

		submatch := re.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}

		js := strings.TrimSpace(submatch[1])
		if js == "" {
			return "" // empty script tag, just remove
		}

		jsBlocks = append(jsBlocks, js)
		hasInlineJS = true
		return "" // remove the inline script block from HTML
	})

	if !hasInlineJS {
		return htmlTemplate, ""
	}

	// Combine all JS blocks
	jsContent = strings.Join(jsBlocks, "\n\n")

	// Add the <script src> reference after </section>
	scriptRef := fmt.Sprintf(`<script src="/tools/assets/%s.js"></script>`, functionName)

	if idx := strings.LastIndex(cleanHTML, "</section>"); idx >= 0 {
		insertAt := idx + len("</section>")
		cleanHTML = cleanHTML[:insertAt] + "\n" + scriptRef + cleanHTML[insertAt:]
	} else {
		cleanHTML = cleanHTML + "\n" + scriptRef
	}

	// Clean up double blank lines left by removed script blocks
	for strings.Contains(cleanHTML, "\n\n\n") {
		cleanHTML = strings.ReplaceAll(cleanHTML, "\n\n\n", "\n\n")
	}

	return cleanHTML, jsContent
}

// ---------------------------------------------------------------------------
// Regeneration helpers
// ---------------------------------------------------------------------------

// snapshotComponentVersion writes the pre-update state of a content_component
// into component_versions so we have history for rollback, diffing, or
// allowing other sites to opt back to an earlier version.
//
// Uses live schema columns: version_number, change_description, changed_by,
// change_source, plus html_template, input_schema, css_template.
// css_template is left NULL for now — the section components store
// everything (CSS + HTML) in html_template and the separate css_template
// column isn't used by the current generator.
//
// changedBy identifies the agent/principal making the change
// (e.g. "component-creator:regen", "tool-improver:auto").
// changeSource identifies the triggering work item or event
// (e.g. the work item's source field, "manual_regen_after_prompt_fix",
// "component-quality-auditor"). May be "" if there is no originating work
// item — the column is nullable, and the helper writes NULL in that case.
//
// The caller is expected to precompute versionNumber as
// MAX(version_number)+1 for this component_id. The unique index
// (component_id, version_number) will reject duplicates, so if two concurrent
// regenerations both compute the same MAX+1, one will fail here and the
// caller logs+continues per best-effort policy.
//
// Returns non-nil error on failure. The caller treats snapshot as
// best-effort: an error here should be logged but not block the UPDATE.
func snapshotComponentVersion(
	ctx context.Context,
	db *sql.DB,
	componentID string,
	versionNumber int,
	htmlTemplate string,
	inputSchemaJSON string,
	jsContent string, // currently not stored in component_versions; passed for future compatibility
	changeDescription string,
	changedBy string,
	changeSource string,
	logger *zap.Logger,
) error {
	_ = jsContent // reserved for when component_versions grows a js_content column

	var changeSourceArg interface{}
	if changeSource == "" {
		changeSourceArg = nil
	} else {
		changeSourceArg = changeSource
	}

	_, err := db.ExecContext(ctx, `
		INSERT INTO component_versions (
			component_id, version_number,
			html_template, input_schema,
			change_description, changed_by, change_source,
			created_at
		) VALUES (
			$1::uuid, $2,
			$3, $4::jsonb,
			$5, $6, $7,
			NOW()
		)
	`,
		componentID,
		versionNumber,
		htmlTemplate,
		inputSchemaJSON,
		changeDescription,
		changedBy,
		changeSourceArg,
	)
	if err != nil {
		return fmt.Errorf("insert component_versions (component=%s, version=%d): %w",
			componentID, versionNumber, err)
	}

	logger.Info("store_generated_component: version snapshot written",
		zap.String("component_id", componentID),
		zap.Int("version_number", versionNumber),
		zap.String("changed_by", changedBy),
		zap.String("change_source", changeSource))
	return nil
}

// markPagesPendingRebuild flips build_status to 'pending' for all
// page_components that use this component_id. Does NOT touch rendered_html
// (that's per-page content; the rerender pipeline will regenerate it with
// variable substitution). Returns the count of pages marked AND the
// distinct site_ids they belong to, so the caller can raise one
// needs_rerender work item per affected site.
//
// Failures are logged but not returned — the UPDATE to content_components
// has already succeeded, and page rebuild eligibility is something the
// auditor can re-check independently. Returns (0, nil) on failure.
func markPagesPendingRebuild(
	ctx context.Context,
	db *sql.DB,
	componentID string,
	logger *zap.Logger,
) (pagesMarked int64, affectedSiteIDs []string) {
	// First, collect the distinct site_ids that will be affected, via the
	// join from page_components → pages. Do this BEFORE the UPDATE so we
	// have a stable set of sites to raise rerender items for.
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT p.site_id::text
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE pc.component_id = $1::uuid
	`, componentID)
	if err != nil {
		logger.Warn("markPagesPendingRebuild: failed to enumerate affected sites",
			zap.String("component_id", componentID),
			zap.Error(err))
		// Still try the UPDATE — pages can be marked pending even if we
		// can't raise rerender items; the auditor will catch them later.
	} else {
		for rows.Next() {
			var sid string
			if scanErr := rows.Scan(&sid); scanErr == nil && sid != "" {
				affectedSiteIDs = append(affectedSiteIDs, sid)
			}
		}
		rows.Close()
	}

	result, err := db.ExecContext(ctx, `
		UPDATE page_components
		SET build_status = 'pending', updated_at = NOW()
		WHERE component_id = $1::uuid
	`, componentID)
	if err != nil {
		logger.Warn("markPagesPendingRebuild: UPDATE failed, page_components left in current state",
			zap.String("component_id", componentID),
			zap.Error(err))
		return 0, affectedSiteIDs
	}
	pagesMarked, _ = result.RowsAffected()
	if pagesMarked > 0 {
		logger.Info("markPagesPendingRebuild: flagged pages for rebuild",
			zap.String("component_id", componentID),
			zap.Int64("pages", pagesMarked),
			zap.Int("sites", len(affectedSiteIDs)))
	}
	return pagesMarked, affectedSiteIDs
}

// nullStringToGo converts a sql.NullString to a plain string, treating
// NULL as "". Keeps call sites readable when we don't care about the
// null-vs-empty distinction.
func nullStringToGo(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

// createRerenderWorkItem inserts a needs_rerender work item for one site,
// scoped to a specific regenerated component. The rerender-pages handler
// picks up these items and rebuilds affected pages.
//
// The item_key is component-scoped (component_regen_rerender:<uuid>) so
// that:
//   - multiple concurrent regens of DIFFERENT components produce
//     distinct work items even within the same site
//   - repeat regens of the SAME component collide with the dedup index
//     idx_swi_dedup (site_id, item_key) and the INSERT is a no-op
//     (excluding completed/failed rows — see index WHERE clause)
//
// Returns true if a row was actually inserted, false if the dedup
// ON CONFLICT path was taken or an error was logged.
//
// Errors are logged but never propagate — an orphaned pending page is
// recoverable by the auditor, and blocking the whole regen on a work
// item write would waste the UPDATE we just completed.
//
// workItemSource is the originating work item's source field, used as
// the `source` column so this synthetic rerender item can be traced back
// to whatever caused the regen. Empty string is safe — `source` has a
// NOT NULL constraint in site_work_items, so we substitute a default
// of "component-creator" when the caller passes "".
func createRerenderWorkItem(
	ctx context.Context,
	db *sql.DB,
	siteID string,
	componentID string,
	functionName string,
	workItemSource string,
	logger *zap.Logger,
) bool {
	sourceField := workItemSource
	if sourceField == "" {
		sourceField = "component-creator"
	}

	itemKey := fmt.Sprintf("component_regen_rerender:%s", componentID)
	summary := fmt.Sprintf("Re-render pages after %s regeneration", functionName)
	// reason=section_data_resolved so the per-page rerender items (created by
	// rerender-pages' create_rerender_items, once it propagates spec.reason)
	// drive a section re-render of this component's dependents rather than an
	// assemble-only re-ship. component_id above is what scopes it to those
	// dependents.
	specJSON := fmt.Sprintf(
		`{"component_id": %q, "function": %q, "reason": "section_data_resolved", "refresh_site_components": false}`,
		componentID, functionName,
	)

	// Insert with a guard that mirrors the dedup index's WHERE clause.
	// The guard on the INSERT is redundant given the unique index but
	// makes the intent readable and avoids a unique-violation error
	// path that would need error-sniffing. Status list from
	// workItemTerminalStatuses so it tracks the index predicate.
	result, err := db.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity,
			summary, priority, handler_agent, status, created_by,
			spec, item_key
		)
		SELECT $1::uuid, $2, 'build', 'needs_rerender', 'medium',
		       $3, 99, 'rerender-pages', 'triaged', 'store_generated_component',
		       $4::jsonb, $5
		WHERE NOT EXISTS (
			SELECT 1 FROM site_work_items
			WHERE site_id = $1::uuid
			  AND item_key = $5
			  AND status NOT IN (%s)
		)
	`, sqlInList(workItemTerminalStatuses)),
		siteID,
		sourceField,
		summary,
		specJSON,
		itemKey,
	)
	if err != nil {
		logger.Warn("createRerenderWorkItem: INSERT failed, site will rely on auditor to catch pending pages",
			zap.String("site_id", siteID),
			zap.String("component_id", componentID),
			zap.Error(err))
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		logger.Info("createRerenderWorkItem: rerender item already pending for this component/site (dedup)",
			zap.String("site_id", siteID),
			zap.String("component_id", componentID),
			zap.String("item_key", itemKey))
		return false
	}

	logger.Info("createRerenderWorkItem: raised needs_rerender work item",
		zap.String("site_id", siteID),
		zap.String("component_id", componentID),
		zap.String("function", functionName),
		zap.String("item_key", itemKey),
		zap.String("source", sourceField))
	return true
}

// healRejectedComponentSectionType fills a NULL section_type on the row a
// REFUSED regeneration was targeting, so the component stops being invisible to
// the two readers that key on that column.
//
// Full rationale at the call site. The two properties that matter here:
//
//   - It is gated on is_active, so a component deactivated as broken is never
//     made selectable by a failed regeneration. Dropping that condition is the
//     mutation the test suite exists to catch.
//   - COALESCE means an already-set section_type is never overwritten. The two
//     columns legitimately differ on live rows, and this repairs an absence
//     only — it is not a reconciliation.
//
// Best-effort: a failure here logs and is otherwise ignored. The caller is
// already returning a rejection, and losing a metadata repair must never turn
// into a second, different error.
func healRejectedComponentSectionType(
	ctx context.Context,
	db *sql.DB,
	logger *zap.Logger,
	existingID string,
	isRegeneration bool,
	sectionType string,
) {
	// Only a regeneration has a row to repair; a refused creation wrote nothing.
	if !isRegeneration || existingID == "" || sectionType == "" || db == nil {
		return
	}

	result, err := db.ExecContext(ctx, `
		UPDATE content_components
		SET section_type = COALESCE(section_type, $1)
		WHERE id = $2::uuid
		  AND is_active = true
		  AND section_type IS NULL
	`, sectionType, existingID)
	if err != nil {
		logger.Warn("store_generated_component: section_type heal on rejection failed (non-fatal)",
			zap.String("component_id", existingID),
			zap.String("section_type", sectionType),
			zap.Error(err))
		return
	}
	if n, rowsErr := result.RowsAffected(); rowsErr == nil && n > 0 {
		logger.Info("store_generated_component: healed NULL section_type on a refused regeneration — the component is no longer invisible to the selector or to the writer's advisory",
			zap.String("component_id", existingID),
			zap.String("section_type", sectionType),
			zap.Int64("rows", n))
	}
}

// ---------------------------------------------------------------------------
// Validation rejection logging
// ---------------------------------------------------------------------------

// orphanSchemaFieldPattern extracts the field name from a Direction-2
// sync issue like:  schema field "card_link_label" has no template variable
// describeSchemaTemplateMismatch names the fields behind a schema/template
// mismatch instead of the bare "do not match" (bugs_open/345, Fable review F4:
// the bare form gave the retry prompt nothing to act on — "change exactly what
// it says was wrong" was unsatisfiable, and item 2396218a burned three FED
// attempts against it on 2026-08-24). The per-field wording is reused VERBATIM
// from QualityIssues — the same strings the classifier below parses — so there
// is one wording with one source, not two implementations of one rule
// (bugs_closed/034). Sorted, because the enriched message now feeds the
// byte-identical repeat detector (345 candidate 2): a message whose field
// order varied between attempts would silently defeat it.
func describeSchemaTemplateMismatch(score ComponentQualityResult) string {
	const bare = "template variables and schema fields do not match"
	mismatch := []string{}
	for _, iss := range score.QualityIssues {
		if orphanSchemaFieldPattern.MatchString(iss) || unknownTemplateVarPattern.MatchString(iss) {
			mismatch = append(mismatch, iss)
		}
	}
	if len(mismatch) == 0 {
		// Sync computed false but no per-field issue matched — keep the old
		// message rather than inventing a second wording for an unknown shape.
		return bare
	}
	sort.Strings(mismatch)
	return bare + ": " + strings.Join(mismatch, "; ")
}

var orphanSchemaFieldPattern = regexp.MustCompile(`^schema field "([^"]+)" has no template variable$`)

// unknownTemplateVarPattern extracts the field name from a Direction-1
// sync issue like:  template var {{.cta_link}} has no schema entry
var unknownTemplateVarPattern = regexp.MustCompile(`^template var \{\{\.([^}]+)\}\} has no schema entry$`)

// recordValidationRejection writes a structured row to agent_error_log
// when pre-store validation rejects an LLM-generated component. This
// gives us a queryable trail of which fields the LLM keeps getting
// wrong, separable from the rest of the chassis log noise.
//
// Best-effort: failures inside this helper are logged at warn level but
// do not affect the caller's return path. The action still returns the
// same rejection error to its caller.
//
// Severity mapping:
//   - "warning"  — Direction-2 bookkeeping mismatch (schema declares a
//     field the template doesn't use, or vice versa). The
//     LLM produced something structurally well-formed but
//     failed list-reconciliation. Common, addressable.
//   - "error"    — Structural failures (template not closed, missing
//     data-component, "<no value>" artifacts, 0-placeholder
//     substantive template). These indicate the LLM
//     produced something broken at a deeper level.
func recordValidationRejection(
	ctx context.Context,
	db *sql.DB,
	logger *zap.Logger,
	params ActionParams,
	functionName string,
	sectionType string,
	htmlTemplate string,
	schemaJSON string,
	score ComponentQualityResult,
	blockingIssues []string,
) {
	if db == nil {
		return
	}

	// Classify issues into orphan-field (bookkeeping) vs other.
	orphanSchemaFields, unknownTemplateVars, otherIssues := classifySyncIssues(score.QualityIssues)

	// Severity: bookkeeping-only failures are warning; anything else is error.
	severity := "warning"
	if len(otherIssues) > 0 || len(unknownTemplateVars) > 0 {
		severity = "error"
	}
	// Severity error if any structural blockers are present (not closed,
	// missing data-component, no-value artifacts, 0-var substantive).
	for _, b := range blockingIssues {
		if strings.Contains(b, "not closed") ||
			strings.Contains(b, "missing data-component") ||
			strings.Contains(b, "<no value>") ||
			strings.Contains(b, "0 {{.var}} placeholders") ||
			strings.Contains(b, "0 {{placeholder") {
			severity = "error"
			break
		}
	}

	// Pull context from the action params.
	workItemID := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.work_item_id")
	siteID := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.site_id")
	domain := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.domain")

	contextPayload := map[string]interface{}{
		"function":                functionName,
		"section_type":            sectionType,
		"pre_store_score":         score.QualityScore,
		"template_variable_count": score.TemplateVariableCount,
		"schema_field_count":      score.SchemaFieldCount,
		"schema_template_synced":  score.SchemaTemplateSynced,
		"template_closed":         score.TemplateClosed,
		"has_data_component":      score.HasDataComponent,
		"template_len":            len(htmlTemplate),
		"schema_len":              len(schemaJSON),
		"orphan_schema_fields":    orphanSchemaFields,
		"unknown_template_vars":   unknownTemplateVars,
		"other_issues":            otherIssues,
		"blocking_issues":         blockingIssues,
		"all_issues":              score.QualityIssues,
	}
	contextJSON, _ := json.Marshal(contextPayload)
	if contextJSON == nil {
		contextJSON = []byte("{}")
	}

	errorMessage := fmt.Sprintf(
		"component validation rejected for function=%q section_type=%q: %s",
		functionName, sectionType, strings.Join(blockingIssues, "; "))

	errorCode := "component_validation_rejected"
	if len(orphanSchemaFields) > 0 && len(otherIssues) == 0 && len(unknownTemplateVars) == 0 {
		errorCode = "component_validation_orphan_schema_field"
	}
	if len(unknownTemplateVars) > 0 {
		errorCode = "component_validation_unknown_template_var"
	}

	// NOTE: component_write_guard.go carries a near-identical insert
	// (recordComponentWriteRejection) for the UPDATE path. Consolidating the
	// two was proposed and deliberately dropped: the council gate's
	// edit-quality and guardian seats both objected that refactoring this
	// BIRTH-path recorder is not required by the bug being fixed (the 012
	// wreck happened on the update path; this guard worked correctly), and
	// that a provenance-literal slip here would silently misfile birth-path
	// rejections fleet-wide. Left standing on purpose — the duplication is
	// cheaper than the blast radius.
	_, err := db.ExecContext(ctx, `
		INSERT INTO agent_error_log (
			site_id, domain, work_item_id, orchestration_id,
			agent_type, agent_id, pod_name, step_name, action,
			error_message, error_code, severity, context
		) VALUES (
			NULLIF($1, '')::uuid, $2, NULLIF($3, '')::uuid, $4,
			$5, $6, $7, $8, $9,
			$10, $11, $12, $13::jsonb
		)
	`,
		siteID,
		domain,
		workItemID,
		params.ExecutionContext.OrchestrationID,
		"component-creator",
		params.ExecutionContext.Sender.AgentID,
		params.ExecutionContext.Sender.PodName,
		"store_component",
		"store_generated_component",
		errorMessage,
		errorCode,
		severity,
		string(contextJSON),
	)
	if err != nil {
		logger.Warn("recordValidationRejection: failed to write to agent_error_log",
			zap.Error(err),
			zap.String("function", functionName))
	}

	// bugs_open/345: the same refusal, on the ROW, in a typed channel the next
	// generation can read. Deliberately the SAME errorMessage and errorCode the
	// insert above uses — one message, two destinations. Composing a second,
	// prettier message here would be two implementations of one rule, which is
	// the drift class bugs_closed/034 closed.
	recordRetryFeedback(ctx, db, logger, workItemID, errorCode, errorMessage,
		params.ExecutionContext.OrchestrationID, "store_component")
}

// recordRetryFeedback is THE ONLY WRITER of site_work_items.retry_feedback
// (migration 561), and that singularity is the entire point of the column.
//
// WHY A DEDICATED COLUMN EXISTS AT ALL. bugs_open/345's shipped fix feeds the
// previous failure back to the writer so a retry can differ from its
// predecessor — and it works (item ceea0c07, 2026-08-22: refused 12:18:43Z,
// re-dispatched 12:51 carrying the text, completed 12:53:07). But it was
// reading `site_work_items.error`, and that column is a general-purpose
// annotation field: [MEASURED 2026-08-22] TWENTY write sites across TEN files
// write it, three of them the human operator HTTP path in
// internal/core-manager/admin/site_admin_handlers.go, plus hand-run SQL. Of the
// 799 rows fleet-wide that passed the loader's gate, 405 were human notes
// ("HELD 2026-08-18 by the loanzy_uk_example_site lane: …") and only 11 were
// validation rejections. The prompt nevertheless told the model "your previous
// output for this component was refused by validation" — false for 6 of the 17
// items (35%) that could actually reach it.
//
// The fix is not a better classifier. Classifying at the reader would mean
// matching error TEXT written freely by twenty producers and by people; a
// column with one writer makes the wrong answer unrepresentable instead of
// merely detectable. `error` keeps its twenty writers and its meaning.
//
// SO: if you are about to add a second writer to this column, don't. Add your
// own producer call to THIS function with your own code, or the guarantee the
// reader depends on is gone and no test will notice.
//
// Best-effort by design, exactly like the agent_error_log insert above: the
// caller is already failing, and losing the feedback costs one blind
// regeneration (pre-345 behaviour), never correctness.
//
// One write per failure EVENT, never on a cadence — WII-018:
// trg_site_work_items_updated_at bumps updated_at on every write and the stale
// reaper keys on it, so a periodic writer makes an open item permanently
// unreapable. updated_at is deliberately NOT set here: the trigger owns it.
func recordRetryFeedback(ctx context.Context, db *sql.DB, logger *zap.Logger,
	workItemID, code, message, orchestrationID, step string) {

	if db == nil || workItemID == "" {
		// No work item means this generation was not dispatched from the queue
		// (a canary, a direct call). There is nothing to feed a retry with.
		return
	}

	// A BLANK message is a NO-OP, never a write (bugs_open/345, at the request of
	// bugs_open/395's second producer). This function REPLACES the column
	// wholesale, and the reader keys on a non-blank message
	// (load_work_item_actions.go: `strings.TrimSpace(lastError.String) != ""`).
	// So writing an empty message does not "clear" anything useful — it CLOBBERS
	// whatever specific feedback a producer already wrote for this attempt with
	// {"message":"",...}, which the reader then drops, and the retry regenerates
	// blind. The invariant "don't pass an empty message" was call-site
	// discipline; a second producer read this function's doc comment and still
	// got the mechanism wrong, so the invariant is enforced HERE where it cannot
	// be. No-op today (store_component always passes a real message); this makes
	// every future producer's empty call harmless rather than harmful.
	if strings.TrimSpace(message) == "" {
		logger.Debug("recordRetryFeedback: blank message — skipped, not written "+
			"(a blank write would clobber existing feedback and the reader would drop it; bugs_open/345)",
			zap.String("work_item_id", workItemID),
			zap.String("code", code))
		return
	}

	// completed_at IS NULL mirrors the reader's gate (load_work_item_actions.go):
	// a completed row keeps its failure text, and writing feedback onto a
	// lifecycle that already ended is how a fresh generation inherits a dead
	// one's refusal — the round-3 hazard, LANDMINES.md:7104.
	_, err := db.ExecContext(ctx, `
		UPDATE site_work_items
		SET retry_feedback = jsonb_build_object(
		        'code',             $2::text,
		        'message',          $3::text,
		        'at',               NOW()::text,
		        'orchestration_id', NULLIF($4::text, ''),
		        'step',             NULLIF($5::text, ''))
		WHERE id = NULLIF($1::text, '')::uuid
		  AND completed_at IS NULL
	`, workItemID, code, message, orchestrationID, step)

	if err != nil {
		logger.Warn("recordRetryFeedback: failed to record typed retry feedback — "+
			"the next attempt will regenerate blind (pre-345 behaviour)",
			zap.Error(err),
			zap.String("work_item_id", workItemID),
			zap.String("code", code))
	}
}

// deriveRenderMode inspects a JSON-encoded input_schema and returns "agent"
// if any field has source="llm", otherwise "template".
//
// This ensures render_mode is always consistent with the schema rather than
// being hardcoded at creation time. The page-content-writer workflow's
// check_render_mode conditional routes sections to LLM generation when
// render_mode == "agent"; without this derivation every component would
// permanently take the template-only path regardless of its content needs.
//
// Called by both the INSERT (creation) and UPDATE (regeneration) paths in
// StoreGeneratedComponentAction so the value is always up to date.
func deriveRenderMode(inputSchemaJSON string) string {
	if inputSchemaJSON == "" || inputSchemaJSON == "{}" {
		return "template"
	}

	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(inputSchemaJSON), &schema); err != nil {
		return "template"
	}

	fields, ok := schema["fields"].(map[string]interface{})
	if !ok {
		return "template"
	}

	for _, v := range fields {
		fieldDef, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if source, ok := fieldDef["source"].(string); ok && source == "llm" {
			return "agent"
		}
	}

	return "template"
}

// schemaFieldSet returns the set of field names declared under
// input_schema.fields. Empty or invalid schema → empty set. Mirrors the
// fields-parse used by deriveRenderMode.
// classifySyncIssues splits scoreComponent's QualityIssues into the two
// schema/template directions plus everything else, using the SAME regexes the
// stored classifier and describeSchemaTemplateMismatch rely on — one parse, one
// source of truth (the bugs_closed/034 drift rule). orphans are schema fields
// with no {{.placeholder}}; unknownVars are {{.placeholders}} with no schema
// entry; others is any remaining issue string.
func classifySyncIssues(qualityIssues []string) (orphans, unknownVars, others []string) {
	orphans = []string{}
	unknownVars = []string{}
	others = []string{}
	for _, issue := range qualityIssues {
		if m := orphanSchemaFieldPattern.FindStringSubmatch(issue); len(m) > 1 {
			orphans = append(orphans, m[1])
			continue
		}
		if m := unknownTemplateVarPattern.FindStringSubmatch(issue); len(m) > 1 {
			unknownVars = append(unknownVars, m[1])
			continue
		}
		others = append(others, issue)
	}
	return orphans, unknownVars, others
}

// dropSchemaFields returns inputSchemaJSON with the named fields removed from
// .fields (house dialect). bugs_open/345: an orphan-only mismatch drops its
// unrendered fields rather than refusing the store. Preserves every other key
// of the schema object untouched; errors if the JSON does not parse so the
// caller can fall back to blocking rather than store a schema it could not
// reduce.
func dropSchemaFields(inputSchemaJSON string, names []string) (string, error) {
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(inputSchemaJSON), &schema); err != nil {
		return "", fmt.Errorf("dropSchemaFields: parse: %w", err)
	}
	fields, ok := schema["fields"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("dropSchemaFields: no .fields object")
	}
	for _, n := range names {
		delete(fields, n)
	}
	schema["fields"] = fields
	out, err := json.Marshal(schema)
	if err != nil {
		return "", fmt.Errorf("dropSchemaFields: marshal: %w", err)
	}
	return string(out), nil
}

func schemaFieldSet(inputSchemaJSON string) map[string]bool {
	out := map[string]bool{}
	if inputSchemaJSON == "" || inputSchemaJSON == "{}" {
		return out
	}
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(inputSchemaJSON), &schema); err != nil {
		return out
	}
	fields, ok := schema["fields"].(map[string]interface{})
	if !ok {
		return out
	}
	for name := range fields {
		out[name] = true
	}
	return out
}
