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
	// Verb wrapped as a capture group (bugs_open/222) so the negation guard can
	// scan back from the QUALIFIER specifically — see declGenerateVerbData.
	// Nothing else read this group when it was unparenthesised.
	fabGenerateVerbData = regexp.MustCompile(`(?is)(generat\w+)[^\n]{0,64}(dataset|records|entries|list of (?:practices|businesses|entities|items|records|companies|listings))`)

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

// ── Negation guard (bugs_open/222) ──────────────────────────────────────
//
// Tier A convicts a QUALIFIER word near a data noun, with no regard for
// whether the sentence DENIES fabrication rather than declaring it. Real
// incident: a recreation's only matching text was the comment
// "// In-memory portfolio store (no fabricated data — starts empty)" — a
// denial, on a tool that genuinely starts empty. The recreate prompt's own
// Data Integrity section spends ~9 lines forbidding fabrication, so a
// conscientious model echoing that prohibition as a comment is the COMMON
// case, not a freak one (prompt-text-poisons-its-own-detector).
//
// The scanning ALGORITHM is not new: platform/orchestration/datahelpers/
// claims.go built and proved it for the banned-claims layer (CLM-017) —
// bounded backwards window, trimmed to the current clause, tested against a
// cue regex, with a non-obvious multibyte-rune fix already paid for.
// datahelpers.NegationGuard exports exactly that algorithm so this package
// does not reimplement it (CLM-004's anti-drift argument: one algorithm, not
// two that diverge).
//
// The VOCABULARY below is deliberately NOT shared with claims.go's
// negationCueRe, and must never be merged with it in either direction:
// negationCueRe excludes bare "no"/"without" ON PURPOSE, because in marketing
// prose they are intensifiers ("Without exception, every claim is verified")
// whose exclusion is pinned by a dedicated residual test. In a code comment
// about a tool's own data, "no"/"without" are plain denials — the motivating
// payload above is negated by exactly the cue claims.go excludes. Widening
// the shared vocabulary to fix this gate would risk reintroducing the false
// negative CLM-017 was built to prevent, in an unrelated layer. Two
// vocabularies, one algorithm.
var fabNegationCueRe = regexp.MustCompile(
	`(?i)\b(?:no|not|never|nor|none|without|zero|cannot|instead of|rather than)\b` +
		`|[a-z]n['’‘]t\b` +
		`|\b(?:cant|dont|doesnt|didnt|isnt|arent|wasnt|werent|wont|couldnt|shouldnt|wouldnt|mustnt)\b`,
)

// DELIBERATELY ABSENT from fabNegationCueRe, each because it would license a
// REAL declaration rather than deny one:
//
//	"avoid"          — "to avoid an empty state we generate placeholder rows"
//	                    IS a declaration; avoid negates the empty state, not
//	                    the fabrication.
//	"fails to"/"unable to" — the vetcomp incident verbatim: "if the fetch
//	                    fails to load, generate a realistic dataset" — the
//	                    fetch's failure is what LICENSES the fabrication.
//	"rarely"/"seldom" — hedges; "rarely uses mock data" still admits the act.

// fabNegationWindowBytes: half the claims layer's 64. This domain's cues
// include bare "no"/"without", whose false-suppression risk grows with
// distance from the qualifier; 32 clears every observed denial shape
// ("instead of dynamically generating", 24 bytes) with slack for one more
// short adverb. Known residual at this width: "no real data anywhere in the
// tool so we generate…" — Tier B still catches that class (see the risk
// register in PLAN_2026-08-08_negation_aware_declaration_tier.md §7).
const fabNegationWindowBytes = 32

// fabNegationClauseBoundary: the claims-layer boundary set plus brackets —
// this domain scans raw JS/HTML, where a bracket ends a comment's natural
// clause. Each added boundary only SHRINKS the guard's reach, which is the
// conservative direction (toward convicting, not away from it).
const fabNegationClauseBoundary = ".!?;:,<>\n\r\t–—(){}[]"

var fabNegationGuard = datahelpers.NegationGuard{Cue: fabNegationCueRe, Boundary: fabNegationClauseBoundary, Window: fabNegationWindowBytes}

// declPattern pairs a Tier A declaration regex with the submatch index of its
// negatable QUALIFIER token. The three Tier A regexes disagree on where that
// token sits — fabQualifierNearData puts it first, fabDataNearQualifier and
// fabGenerateVerbData put it second — and a negator between the two capture
// groups lands INSIDE the match, invisible to a guard that only scans back
// from the match's overall start ("records are never generated": the match
// starts at "records", and "never" sits inside the span). Scanning from the
// qualifier's own submatch position sees it regardless of which group it is.
type declPattern struct {
	re        *regexp.Regexp
	qualGroup int
}

var (
	declQualifierNearData = declPattern{fabQualifierNearData, 1}
	declDataNearQualifier = declPattern{fabDataNearQualifier, 2}
	declGenerateVerbData  = declPattern{fabGenerateVerbData, 1}
)

// firstAssertedDeclaration returns the first match of p in text whose
// qualifier is NOT negated in its clause, plus every match the guard
// suppressed. Suppressed matches are returned rather than dropped silently —
// a guard that suppresses without a trace is indistinguishable from a gate
// that stopped working (the same argument as ScanBannedClaimsIgnoringNegation
// in claims.go): they are surfaced in FabricationResult.Signals for
// observability, never counted toward Fabricated.
func firstAssertedDeclaration(text string, p declPattern) (asserted string, suppressed []string) {
	for _, m := range p.re.FindAllStringSubmatchIndex(text, -1) {
		qStart := m[2*p.qualGroup]
		if qStart >= 0 && fabNegationGuard.NegatedAt(text, qStart) {
			suppressed = append(suppressed, text[m[0]:m[1]])
			continue
		}
		if asserted == "" {
			asserted = text[m[0]:m[1]]
		}
	}
	return asserted, suppressed
}

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

	// FAIL-SAFE, not fail-open. A safety gate that cannot inspect its input must
	// not silently pass — that silent default is the exact `missingkey=zero` class
	// bug 020 itself belongs to. If there is no recreation HTML to read (a missing
	// field, an upstream extraction bug, or a drifted field path), hold for review
	// rather than deploy unvetted output. This is normally unreachable —
	// check_completeness runs first and guarantees substantial clean_html or errors
	// out — so it will not cause spurious reviews; but if the path ever drifts it
	// fails LOUD (everything held) instead of becoming a silent no-op.
	if strings.TrimSpace(recreation) == "" {
		res.Fabricated = true
		res.Tier = "uninspectable"
		res.Signals = []string{"no recreated tool HTML to inspect — cannot confirm the output is not fabricated"}
		res.Detail = "Fail-safe: the fabrication gate could not read the recreation output, so the item is held for human review rather than deployed."
		return res
	}

	// ── Tier A: declaration ────────────────────────────────────────────
	// Checked in order, first asserted (non-negated) match wins — same
	// priority as the original fabQualifierNearData → fabDataNearQualifier →
	// fabGenerateVerbData chain. Every pattern's suppressed matches are kept
	// for observability regardless of which arm ultimately convicts.
	declSignals := []string{}
	var negatedNotes []string
	for _, arm := range []struct {
		pattern declPattern
		label   string
	}{
		{declQualifierNearData, "declared synthetic/fake data: "},
		{declDataNearQualifier, "declared generated data: "},
		{declGenerateVerbData, "declared it generates a dataset: "},
	} {
		m, suppressed := firstAssertedDeclaration(recreation, arm.pattern)
		negatedNotes = append(negatedNotes, suppressed...)
		if m != "" && len(declSignals) == 0 {
			declSignals = append(declSignals, arm.label+snippet(m))
		}
	}

	// Tier A: synthetic PII generator the recreation INTRODUCED (not inherited
	// from a faithful original that already had it). Not negation-guarded: an
	// identifier generator is code the recreation runs, not deniable prose.
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
	// Nothing convicted at Tier A. Surface any denial the guard suppressed —
	// informational only, never re-considered for gating — so a reviewer of a
	// later Tier B conviction on the same text can see what Tier A declined.
	for _, s := range negatedNotes {
		res.Signals = append(res.Signals, "negated declaration ignored: "+snippet(s))
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

	// Read the field PATHS directly from config (literal dot-paths), then resolve
	// each ONCE. We deliberately do NOT use datahelpers.ExtractActionInputs here:
	// its Strategy 0 resolves any dotted config VALUE against collected_data, which
	// turns html_field="completeness_check.clean_html" into the HTML *content* — and
	// then extracting again with that content as a path yields "" (a silent no-op on
	// the fail-open detector, or a fail-safe over-HOLD of every recreation on this
	// one). check_tool_completeness reads config directly for exactly this reason.
	// Caught live by the bug-020 induced-fault probe, 2026-07-22 (WRONG_CALLS).
	config := params.StepConfig.Config
	htmlField := configStringOr(config, "html_field", "completeness_check.clean_html")
	originalField := configStringOr(config, "original_html_field", "existing_content.existing_content.raw_html")
	analysisField := configStringOr(config, "analysis_field", "tool_analysis.result")

	recreation := datahelpers.ExtractNestedFieldString(params.CollectedData, htmlField)
	if recreation == "" {
		// FAIL-SAFE: missing recreation HTML is anomalous here (check_completeness
		// runs first). Do NOT silently pass — DetectToolFabrication returns an
		// "uninspectable" fabricated=true so the item is held for review, and a
		// drifted field path fails loud instead of becoming a silent no-op.
		logger.Warn("check_tool_fabrication: no recreation HTML at field — holding for review (fail-safe)",
			zap.String("field", htmlField))
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

// configStringOr reads a literal string from step config, or returns def. Used
// for dot-path fields that must NOT be pre-resolved by ExtractActionInputs.
func configStringOr(config map[string]interface{}, key, def string) string {
	if config != nil {
		if v, ok := config[key].(string); ok && v != "" {
			return v
		}
	}
	return def
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
