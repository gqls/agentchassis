// FILE: platform/orchestration/actions/verify_report_prose_action.go
//
// VerifyReportProseAction is the deterministic gate between the LLM prose
// steps and the report page (gripper dossier pilot). validate_page_content's
// evidence-base number check (check 8) cannot serve here — a report is full
// of per-request computed figures that will never be in the site register —
// so this action binds the prose to the per-request fact instead:
//
//   1. Every numeric token in the prose must appear in the allowed set
//      derived from the scoring output's fact_block (formatting-tolerant:
//      "32.7", "32.70" and "1,520" all normalise identically).
//   2. Every candidate/manufacturer name family mentioned must belong to the
//      scored candidate set.
//   3. If match_count == 0, the summary must carry the mandatory sentence
//      verbatim, and no softening/purchase language may appear anywhere.
//   4. No section may be empty.
//   5. The prose step must not have recorded a tolerated truncation — a
//      fragment that parses is still a fragment.
//
// Contract mirrors validate_page_content: violation returns (nil, error) so
// the workflow's error_step routes to the failure path. Offending tokens are
// in the error text and the log.
//
// Config:
//   - prose_field:       dotted path to the prose object {summary_html,
//                        candidates_html, integration_html,
//                        vendor_questions_html}
//   - scoring_field:     dotted path to score_grippers' output
//   - context_field:     optional; the request row, whose string values the
//                        prose may legitimately echo
//   - truncation_field:  optional override for the __truncated marker path
//                        (default: sibling of prose_field, see
//                        truncationMarkerField)

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var VerifyReportProseInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("verify_report_prose", VerifyReportProseInputSpec)
}

var (
	proseTagRe   = regexp.MustCompile(`<[^>]*>`)
	proseNumRe   = regexp.MustCompile(`\d[\d,]*\.?\d*`)
	softeningRes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)meets? (?:the|your) requirement`),
		regexp.MustCompile(`(?i)suitable for your application`),
		regexp.MustCompile(`(?i)we recommend (?:purchasing|buying|ordering)`),
		regexp.MustCompile(`(?i)good (?:fit|match) for your`),
	}
)

const proseSectionCount = 4

func VerifyReportProseAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "verify_report_prose"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	proseField, _ := config["prose_field"].(string)
	scoringField, _ := config["scoring_field"].(string)
	if proseField == "" || scoringField == "" {
		return nil, fmt.Errorf("verify_report_prose requires prose_field and scoring_field in step config")
	}

	proseRaw := datahelpers.ExtractNestedField(params.CollectedData, proseField)
	scoringRaw := datahelpers.ExtractNestedField(params.CollectedData, scoringField)
	prose, ok := proseRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("prose_field %q did not resolve to an object (got %T)", proseField, proseRaw)
	}
	scoring, ok := scoringRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("scoring_field %q did not resolve to an object (got %T)", scoringField, scoringRaw)
	}

	// Truncation guard (bugs_open/012/019 class). ExecuteLLMPromptAction
	// tolerates a cut response and stamps __truncated as a SIBLING of .result
	// — the step then SUCCEEDS, so a consumer that ignores the marker cannot
	// tell a complete dossier from a fragment. A partial that PARSES is still
	// a partial: prose cut mid-section can close into valid JSON with every
	// key present and the honest caveats missing, and this gate's numeric
	// checks would pass it. Refuse it here, where the page is still unwritten.
	//
	// Derived from prose_field by default (diagnose_council_decide precedent)
	// so it cannot drift out of step with it; truncation_field overrides for
	// a prose path that is not the step result's immediate child.
	override, _ := config["truncation_field"].(string)
	if mf := truncationMarkerField(proseField, override); mf != "" {
		if t, ok := datahelpers.ExtractNestedField(params.CollectedData, mf).(bool); ok && t {
			return nil, fmt.Errorf("prose step recorded a TRUNCATION at %q — refusing to publish a fragment", mf)
		}
	}

	// Optional: the raw request row (load_request output). Its string values
	// are context the prose may legitimately echo (mounting standard, part
	// geometry, budget) even though score_grippers never sees them.
	var contextValues []string
	if cf, _ := config["context_field"].(string); cf != "" {
		if m, ok := datahelpers.ExtractNestedField(params.CollectedData, cf).(map[string]interface{}); ok {
			for _, v := range m {
				if s, ok := v.(string); ok && s != "" {
					contextValues = append(contextValues, s)
				}
			}
		}
	}

	violations := verifyReportProse(prose, scoring, contextValues,
		loadKnownVendors(ctx, params.DB, logger))
	if len(violations) > 0 {
		logger.Warn("VerifyReportProseAction: REJECTED", zap.Strings("violations", violations))
		return nil, fmt.Errorf("report prose failed verification (%d violations): %s",
			len(violations), strings.Join(violations, " | "))
	}

	logger.Info("VerifyReportProseAction: prose verified against fact block")
	return map[string]interface{}{"verified": true, "sections": proseSectionCount}, nil
}

// loadKnownVendors returns the vendor universe the prose is checked against:
// every manufacturer the platform has actually indexed, unioned with the
// curated seed list below.
//
// The live half matters because a hardcoded list is a second source of truth
// that drifts the moment someone adds a product (council 7ed137d1 round 2,
// tooling_provenance + editquality: "the sketch hardcodes names rather than
// reading them from an existing vendor/product source"). Sourcing from
// `products` means a vendor becomes checkable the moment it is indexed
// ANYWHERE on the platform, with no code change.
//
// The curated half is still needed and is not redundant: the whole point is to
// catch vendors we have NOT indexed, and those by definition do not appear in
// `products`. The two halves cover opposite gaps.
//
// Fail-open on a DB error is deliberate: the seed list still applies, and a
// transient database problem must not fail an otherwise honest report.
func loadKnownVendors(ctx context.Context, db *sql.DB, logger *zap.Logger) []string {
	vendors := append([]string(nil), seedVertexVendors...)
	if db == nil {
		return vendors
	}
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT specifications->>'manufacturer'
		FROM products
		WHERE specifications->>'manufacturer' IS NOT NULL
		  AND length(trim(specifications->>'manufacturer')) > 2`)
	if err != nil {
		logger.Warn("verify_report_prose: could not load indexed manufacturers; falling back to the seed list only",
			zap.Error(err))
		return vendors
	}
	defer rows.Close()

	seen := make(map[string]bool, len(vendors))
	for _, v := range vendors {
		seen[strings.ToLower(v)] = true
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		name = strings.TrimSpace(name)
		if name == "" || seen[strings.ToLower(name)] {
			continue
		}
		seen[strings.ToLower(name)] = true
		vendors = append(vendors, name)
	}
	return vendors
}

// seedVertexVendors is the curated half: industrial gripper/vacuum vendors we
// have NOT indexed — precisely the names a writer reaches for when padding a
// shortlist, and precisely the ones no query can supply. It is NOT an
// allowlist: a name here is a violation UNLESS the scored candidate set (or
// the fact block / request context) actually contains it, so the check relaxes
// on its own as the index grows.
//
// This closes the plain-word half of the name-fabrication gap the council's
// compliance seat raised at high severity (correlation 7ed137d1): modelNumberRe
// catches an invented SKU because a SKU carries digits, but "we also considered
// Piab" carries none and previously passed untouched — and check_claims:false
// on report pages means validate_page_content's scanner is not backstopping it.
//
// What it still does not cover, stated rather than absorbed: a wholly invented
// vendor name that appears on no list (e.g. "Norgren Robotics"). That residual
// is left to the writer prompt and is the reason the prompt must forbid naming
// any vendor absent from the fact block.
var seedVertexVendors = []string{
	"Applied Robotics", "ATI Industrial Automation", "Bimba", "Camozzi",
	"Coval", "Destaco", "Effecto", "Festo", "Gimatic", "Joulin", "Kosmek",
	"Millibar", "OnRobot", "Piab", "Robotiq", "Röhm", "Schmalz", "Schunk",
	"SMC", "Soft Robotics", "Sommer-automatic", "Vaccon", "Weiss Robotics",
	"Zimmer Group",
}

// truncationMarkerField resolves the collected_data path of the prose step's
// __truncated marker. An explicit override wins; otherwise the marker is the
// sibling of prose_field's terminal segment, per markerFieldFor. A prose_field
// with no parent segment has nowhere to look and yields "" (no guard) — that
// shape does not occur for an execute_llm_prompt result, whose prose always
// sits at "<step>.result".
func truncationMarkerField(proseField, override string) string {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override)
	}
	return markerFieldFor(proseField)
}

// modelNumberRe finds SKU-shaped tokens: a prefix mixing letters and digits
// (adjacent, e.g. "2F", "EGP40", "SGM"+digit) followed by a dashed suffix —
// the realistic name-fabrication risk is the writer inventing a sibling
// model number ("2F-140" when only the 2F-85 is indexed). Pure-numeric
// dashed tokens (ISO clause numbers) are left to the numeric gate.
var modelNumberRe = regexp.MustCompile(`\b[A-Za-z0-9]*(?:[A-Za-z][0-9]|[0-9][A-Za-z])[A-Za-z0-9]*-[A-Za-z0-9-]+\b`)

// verifyReportProse is the pure core (unit-tested directly). contextValues
// are request strings (mounting, geometry, budget) the prose may echo;
// knownVendors is the vendor universe from loadKnownVendors.
func verifyReportProse(prose, scoring map[string]interface{}, contextValues, knownVendors []string) []string {
	var violations []string

	factBlock, _ := scoring["fact_block"].(string)
	if factBlock == "" {
		return []string{"scoring output carries no fact_block"}
	}
	allowedText := factBlock + "\n" + strings.Join(contextValues, "\n")
	allowed := numericSet(allowedText)

	// Candidate and maker names form the allowed-name families.
	var candidateNames []string
	if cands, ok := scoring["candidates"].([]interface{}); ok {
		for _, c := range cands {
			if m, ok := c.(map[string]interface{}); ok {
				if n, _ := m["name"].(string); n != "" {
					candidateNames = append(candidateNames, n)
				}
				if mk, _ := m["maker"].(string); mk != "" {
					candidateNames = append(candidateNames, mk)
				}
			}
		}
	}

	matchCount := -1
	switch v := scoring["match_count"].(type) {
	case int:
		matchCount = v
	case float64:
		matchCount = int(v)
	}

	// The report contract travels WITH the scoring output that produced the
	// match_count — never from a package const or a default here. See the note
	// beside score_grippers' output envelope: a second report type in this
	// package would otherwise be silently checked against the first one's
	// sentence and sections, and this gate would pass it.
	//
	// Absent is a REFUSAL, not a fallback. A gate that defaults its own contract
	// reports success on a report it never actually checked.
	// toStringSlice (select_review_panel_action.go) handles the []interface{}
	// the saga's JSON round-trip produces as well as a native []string.
	sections := toStringSlice(scoring["prose_sections"])
	if len(sections) == 0 {
		return []string{"scoring output carries no prose_sections — the gate cannot know which sections to check"}
	}
	noMatch, _ := scoring["no_match_sentence"].(string)
	if strings.TrimSpace(noMatch) == "" {
		return []string{"scoring output carries no no_match_sentence — the gate cannot enforce the honest-no-match contract"}
	}

	for _, key := range sections {
		raw, _ := prose[key].(string)
		text := strings.TrimSpace(proseTagRe.ReplaceAllString(raw, " "))
		if text == "" {
			violations = append(violations, fmt.Sprintf("%s is empty", key))
			continue
		}

		// (1) numeric discipline
		for _, tok := range proseNumRe.FindAllString(text, -1) {
			if !allowed[normaliseNumericToken(tok)] {
				violations = append(violations, fmt.Sprintf("%s asserts number %q not present in the fact block", key, tok))
			}
		}

		// (2) SKU-shaped tokens must trace to the candidate set, the fact
		// block, or the request context — never a sibling model invented by
		// the writer.
		for _, tok := range modelNumberRe.FindAllString(text, -1) {
			if strings.Contains(allowedText, tok) {
				continue
			}
			known := false
			for _, n := range candidateNames {
				if strings.Contains(n, tok) || strings.Contains(tok, n) {
					known = true
					break
				}
			}
			if !known {
				violations = append(violations, fmt.Sprintf("%s names model-like token %q not in the candidate set or fact block", key, tok))
			}
		}

		// (2b) A vendor from the wider field, named without being assessed.
		// Digit-free, so the SKU check above cannot see it.
		for _, vendor := range knownVendors {
			if !containsFold(text, vendor) {
				continue
			}
			if containsFold(allowedText, vendor) {
				continue
			}
			traced := false
			for _, n := range candidateNames {
				if containsFold(n, vendor) {
					traced = true
					break
				}
			}
			if !traced {
				violations = append(violations, fmt.Sprintf(
					"%s names vendor %q, which is not in the assessed index — the report may only discuss candidates it scored", key, vendor))
			}
		}
	}

	// (3) the honest no-match contract
	if matchCount == 0 {
		// CHECKED assertion. This was `prose["summary_html"].(string)`, which
		// PANICS when the writer omits the key or returns null — and it is
		// reachable only on the match_count==0 path, i.e. exactly the case this
		// gate exists to protect. The prose object comes from an LLM step, so a
		// missing key is an ordinary outcome, not an impossible one. A panic
		// here takes down the report build instead of reporting a violation.
		summaryRaw, _ := prose["summary_html"].(string)
		summaryText := proseTagRe.ReplaceAllString(summaryRaw, " ")
		if !strings.Contains(summaryText, noMatch) {
			violations = append(violations, fmt.Sprintf("match_count=0 but summary lacks the mandatory sentence %q", noMatch))
		}
		for _, key := range sections {
			raw, _ := prose[key].(string)
			text := proseTagRe.ReplaceAllString(raw, " ")
			// The mandatory negative sentence itself contains "meets the
			// requirement" — remove it before scanning for softening.
			scan := strings.ReplaceAll(text, noMatch, "")
			for _, re := range softeningRes {
				if loc := re.FindString(scan); loc != "" {
					violations = append(violations, fmt.Sprintf("%s softens a no-match result (%q)", key, loc))
				}
			}
		}
	}

	sort.Strings(violations)
	return dedupeStrings(violations)
}

// containsFold is a case-insensitive strings.Contains — a vendor named in
// lower case is the same fabrication as one named in title case.
func containsFold(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

// numericSet extracts every numeric token from the fact block, normalised.
func numericSet(factBlock string) map[string]bool {
	set := make(map[string]bool)
	for _, tok := range proseNumRe.FindAllString(factBlock, -1) {
		set[normaliseNumericToken(tok)] = true
	}
	return set
}

// normaliseNumericToken strips thousands separators and canonicalises the
// float form so "1,520", "1520" and "1520.0" compare equal.
func normaliseNumericToken(tok string) string {
	clean := strings.ReplaceAll(tok, ",", "")
	clean = strings.TrimSuffix(clean, ".")
	if f, err := strconv.ParseFloat(clean, 64); err == nil {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return clean
}

func dedupeStrings(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := in[:0]
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
