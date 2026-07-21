// FILE: platform/orchestration/actions/check_tool_fabrication_action.go
//
// CheckToolFabricationAction is the mechanical net for /bugs_open/020:
// tool-recreation-handler recreating a DATA-BACKED tool by inventing the data.
//
// The prompt half (migration 183) tells the recreation model never to fabricate
// records and to preserve the original data source. This action is the half that
// does NOT rely on the model obeying: it inspects the generated tool and, when it
// finds a synthetic-data fabrication signature, routes the item to
// needs_human_review instead of deploying it.
//
// Live catch it is built from (vetcomparison.uk, 2026-07-18): a real directory of
// 2,109 verified practices loaded from /data/vet-full-index.json was recreated as
// a widget that GENERATES fake practice names and postcodes in the browser —
// Mulberry32 seeded RNG, TOWNS/PREFIXES/SUFFIXES fragment arrays crossed by
// makePostcode()/buildData(), with a comment admitting "we generate a large,
// realistic, deterministic dataset". It shipped live. Every work item said
// `complete`.
//
// PRECISION — the hard part. A blunt "grep for a PRNG" flags every legitimate
// game/simulator (they use randomness for gameplay). So the detector is tiered:
//
//   Tier A (fires on its own — these essentially never appear in a faithful tool):
//     - the model DECLARING it invented data ("synthetic data", "fake records",
//       "realistic, deterministic dataset", "generate a … dataset"); and
//     - synthetic real-world-IDENTIFIER generators the recreation INTRODUCED that
//       the original did not have (makePostcode/randomPhone/fakeAddress/…).
//
//   Tier B (needs CORROBORATION that the original was data-backed and the
//   recreation dropped the fetch — this is exactly the bug, and it is what keeps
//   a legitimate dice game or name-generator from tripping):
//     - a seeded PRNG, a buildData/generateRecords-style corpus builder, ≥2
//       crossed fragment arrays, or a large literal array of entity records; AND
//     - the original tool loaded data from a source (a data-ish fetch/XHR, or the
//       analysis flagged data_source.has_external_data); AND
//     - the recreation does NOT preserve any data fetch of its own.
//
// A false positive here routes to human review, not silent breakage — but the
// corroboration gate keeps Tier B quiet on the common legitimate cases, and the
// action returns the exact signals that fired so a reviewer sees why.
//
// Registration (registry.go):
//   "check_tool_fabrication": {
//       Handler:     CheckToolFabricationAction,
//       Category:    "validation",
//       Description: "Detect invented/synthetic datasets in a recreated tool (bug 020)",
//       IsLocal:     true,
//   },
//
// Workflow: run after check_completeness; a conditional routes
// fabrication_check.fabricated == true to a checkpoint_for_review step (which
// creates a needs_human_review item and completes WITHOUT deploying), else on to
// save_sections. Wiring migration is separate and image-first (the action must
// exist in the pod before the workflow names it).

package actions

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var CheckToolFabricationInputSpec = datahelpers.ActionInputSpec{
	Required: []string{},
	Optional: []string{"html_field", "original_html_field", "analysis_field"},
	Defaults: map[string]interface{}{
		"html_field":          "completeness_check.clean_html",
		"original_html_field": "existing_content.existing_content.raw_html",
		"analysis_field":      "tool_analysis.result",
	},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("check_tool_fabrication", CheckToolFabricationInputSpec)
}

// ── Detection patterns ──────────────────────────────────────────────────

// JS quote characters: single, double, backtick (\x60 is a literal backtick).
const jsQuote = "[\"'\x60]"

var (
	// Tier A — the model declaring it invented data. Two words on the same line:
	// a synthetic qualifier near a data noun, in either order.
	fabQualifierNearData = regexp.MustCompile(`(?is)(synthetic|fabricat\w+|\bfake\b|\bdummy\b|\bmock\b|placeholder|realistic|deterministic|invent\w+|made[-\s]?up)[^\n]{0,48}(dataset|data[-\s]?set|\bdata\b|records|entries|practices|entities|listings|businesses)`)
	fabDataNearQualifier = regexp.MustCompile(`(?is)(dataset|data[-\s]?set|records|entries|practices|entities|listings)[^\n]{0,48}(synthes\w+|generat\w+|fabricat\w+|randomly|seeded|deterministic\w*|invent\w+)`)
	fabGenerateVerbData  = regexp.MustCompile(`(?is)\bgenerat\w+[^\n]{0,64}(dataset|records|entries|list of (?:practices|businesses|entities|items|records|companies|listings))`)

	// Tier A — synthetic real-world IDENTIFIER generators. These fabricate PII
	// (postcodes/phones/addresses/coordinates) and are unambiguous.
	fabSyntheticPII = regexp.MustCompile(`(?i)\b(make|random|generate|gen|fake|create|build|synth|rand)(Postcode|Postcodes|Zip|Zipcode|Zipcodes|Phone|PhoneNumber|Phones|Address|Addresses|Coordinate|Coordinates|Latitude|Longitude|LatLng|LatLon)s?\b`)

	// Tier B — seeded PRNG (named, or the Mulberry32 imul/>>>0 idiom).
	fabNamedPRNG = regexp.MustCompile(`(?i)\b(mulberry32|xmur3|splitmix\d*|sfc32|xoshiro\w*|xorshift\w*|jsf32|tychei)\b`)
	fabImul      = regexp.MustCompile(`\bMath\.imul\b`)
	fabShift0    = regexp.MustCompile(`>>>\s*0\b`)

	// Tier B — corpus builder functions (record-ish nouns only, not generic UI).
	fabCorpusBuilder = regexp.MustCompile(`(?i)\b(build|generate|make|seed|create|assemble|synth)(Data|Dataset|Records|Rows|Entries|Entities|Directory|Catalog|Catalogue|Businesses|Practices|Listings)\b`)

	// Tier B — UPPERCASE fragment arrays typically crossed to fabricate labels.
	fabFragmentArray = regexp.MustCompile(`\b(TOWNS|CITIES|COUNTIES|REGIONS|STREETS|ROADS|PREFIXES|SUFFIXES|FIRST_NAMES|FIRSTNAMES|LAST_NAMES|LASTNAMES|SURNAMES|COMPANY_TYPES|COMPANYTYPES|COMPANY_NAMES|BUSINESS_TYPES|ADJECTIVES|NOUNS|SYLLABLES|WORD_PARTS)\b`)

	// Tier B — a large literal array of entity records: object literals carrying
	// an entity-ish key. Counted; ≥ fabLiteralRecordThreshold means a hardcoded
	// corpus rather than a handful of config items.
	fabRecordObject = regexp.MustCompile(`(?is)\{[^{}]{0,200}?\b(name|title|postcode|address|phone|company|practice|business|listing)\b\s*:`)

	// Corroboration — original tool loaded data from a source.
	fabOriginalDataFetch = regexp.MustCompile(`(?is)\bfetch\s*\(\s*` + jsQuote + `[^"'\x60]*(\.json|/data/|/api/|\.csv|\?)` +
		`|\bXMLHttpRequest\b|\$\.getJSON\b|\baxios\.(get|post)\b|\$\.ajax\b`)

	// A data fetch of ANY kind in the recreation (did it preserve a source?).
	fabAnyFetch = regexp.MustCompile(`(?is)\bfetch\s*\(|\bXMLHttpRequest\b|\$\.getJSON\b|\baxios\.\w+|\$\.ajax\b`)
)

const fabLiteralRecordThreshold = 15

// FabricationResult is the outcome of the detector. Exported for tests.
type FabricationResult struct {
	Fabricated bool     `json:"fabricated"`
	Tier       string   `json:"tier"`    // "declaration" | "corroborated_corpus" | ""
	Signals    []string `json:"signals"` // human-readable tells that fired
	Detail     string   `json:"detail"`
}

// DetectToolFabrication is the pure core of the gate — no DB, no ActionParams —
// so its logic is fully unit-testable. `recreation` is the generated tool HTML;
// `original` is the crawled source (may be empty); `analysisDataBacked` is true
// when the analysis step flagged an external data source.
func DetectToolFabrication(recreation, original string, analysisDataBacked bool) FabricationResult {
	var res FabricationResult
	if strings.TrimSpace(recreation) == "" {
		return res
	}

	// ── Tier A: declaration ────────────────────────────────────────────
	declSignals := []string{}
	if m := fabQualifierNearData.FindString(recreation); m != "" {
		declSignals = append(declSignals, "declared synthetic/fake data: "+snippet(m))
	} else if m := fabDataNearQualifier.FindString(recreation); m != "" {
		declSignals = append(declSignals, "declared generated data: "+snippet(m))
	} else if m := fabGenerateVerbData.FindString(recreation); m != "" {
		declSignals = append(declSignals, "declared it generates a dataset: "+snippet(m))
	}

	// Tier A: synthetic PII generator the recreation INTRODUCED (not inherited
	// from a faithful original that already had it).
	if pii := fabSyntheticPII.FindString(recreation); pii != "" && !fabSyntheticPII.MatchString(original) {
		declSignals = append(declSignals, "synthetic identifier generator introduced: "+pii)
	}

	if len(declSignals) > 0 {
		res.Fabricated = true
		res.Tier = "declaration"
		res.Signals = declSignals
		res.Detail = "The recreation declares or introduces invented data — this is fabrication regardless of the original tool."
		return res
	}

	// ── Tier B: corpus signature + corroboration ───────────────────────
	corpusSignals := []string{}
	if fabNamedPRNG.MatchString(recreation) {
		corpusSignals = append(corpusSignals, "seeded PRNG (named): "+fabNamedPRNG.FindString(recreation))
	} else if fabImul.MatchString(recreation) && fabShift0.MatchString(recreation) {
		corpusSignals = append(corpusSignals, "seeded PRNG idiom (Math.imul + >>>0)")
	}
	if fabCorpusBuilder.MatchString(recreation) {
		corpusSignals = append(corpusSignals, "corpus builder function: "+fabCorpusBuilder.FindString(recreation))
	}
	if frags := fabFragmentArray.FindAllString(recreation, -1); len(dedupe(frags)) >= 2 {
		corpusSignals = append(corpusSignals, "crossed fragment arrays: "+strings.Join(dedupe(frags), ", "))
	}
	if n := len(fabRecordObject.FindAllString(recreation, -1)); n >= fabLiteralRecordThreshold {
		corpusSignals = append(corpusSignals, fmt.Sprintf("large literal record array (~%d entity objects)", n))
	}

	if len(corpusSignals) == 0 {
		return res // no corpus signature — nothing to corroborate
	}

	dataBacked := analysisDataBacked || fabOriginalDataFetch.MatchString(original)
	preserved := fabAnyFetch.MatchString(recreation)

	if dataBacked && !preserved {
		res.Fabricated = true
		res.Tier = "corroborated_corpus"
		res.Signals = corpusSignals
		res.Detail = "The original tool loaded data from a source, the recreation dropped that fetch, and it builds a synthetic record corpus instead — the bug-020 signature."
		return res
	}

	// Corpus signature present but not corroborated (e.g. a legitimate game/sim
	// using randomness, or a recreation that kept its data fetch). Report the
	// signals for observability, but do NOT gate.
	res.Signals = corpusSignals
	res.Detail = "Corpus-generation signals present but not corroborated (original not data-backed, or the recreation preserves a data fetch) — not gated."
	return res
}

func CheckToolFabricationAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "check_tool_fabrication"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, CheckToolFabricationInputSpec, logger)
	if err != nil {
		return nil, err
	}

	htmlField := inputs.Get("html_field")
	if htmlField == "" {
		htmlField = "completeness_check.clean_html"
	}
	originalField := inputs.Get("original_html_field")
	if originalField == "" {
		originalField = "existing_content.existing_content.raw_html"
	}
	analysisField := inputs.Get("analysis_field")
	if analysisField == "" {
		analysisField = "tool_analysis.result"
	}

	recreation := datahelpers.ExtractNestedFieldString(params.CollectedData, htmlField)
	if recreation == "" {
		// No HTML to inspect — nothing to gate. Let the workflow continue; the
		// completeness/validate steps own the empty-output case.
		logger.Warn("check_tool_fabrication: no recreation HTML found — skipping", zap.String("field", htmlField))
		return map[string]interface{}{"fabricated": false, "skipped": true, "reason": "no recreation html"}, nil
	}
	original := datahelpers.ExtractNestedFieldString(params.CollectedData, originalField)

	// data_source.has_external_data from the analysis, when present.
	analysisDataBacked := false
	if ds := datahelpers.ExtractNestedField(params.CollectedData, analysisField+".data_source"); ds != nil {
		analysisDataBacked = dataSourceIsExternal(ds)
	}

	res := DetectToolFabrication(recreation, original, analysisDataBacked)

	logger.Info("check_tool_fabrication",
		zap.Bool("fabricated", res.Fabricated),
		zap.String("tier", res.Tier),
		zap.Int("signal_count", len(res.Signals)),
		zap.Bool("analysis_data_backed", analysisDataBacked),
	)
	if res.Fabricated {
		logger.Warn("check_tool_fabrication: FABRICATION DETECTED — routing to human review",
			zap.String("tier", res.Tier),
			zap.Strings("signals", res.Signals),
		)
	}

	return map[string]interface{}{
		"fabricated": res.Fabricated,
		"tier":       res.Tier,
		"signals":    res.Signals,
		"detail":     res.Detail,
	}, nil
}

// dataSourceIsExternal reads the analysis data_source object. has_external_data
// is authored by the LLM as a free-text field, so treat any truthy/non-"false"
// string, or a non-empty source, as "the original had external data".
func dataSourceIsExternal(ds interface{}) bool {
	m, ok := ds.(map[string]interface{})
	if !ok {
		return false
	}
	switch v := m["has_external_data"].(type) {
	case bool:
		if v {
			return true
		}
	case string:
		low := strings.ToLower(strings.TrimSpace(v))
		if low != "" && low != "false" && low != "no" && !strings.HasPrefix(low, "false") {
			return true
		}
	}
	if src, ok := m["source"].(string); ok && strings.TrimSpace(src) != "" {
		return true
	}
	return false
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func snippet(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 90 {
		return s[:90] + "…"
	}
	return s
}
