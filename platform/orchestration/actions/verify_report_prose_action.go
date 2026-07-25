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

	violations := verifyReportProse(prose, scoring, contextValues)
	if len(violations) > 0 {
		logger.Warn("VerifyReportProseAction: REJECTED", zap.Strings("violations", violations))
		return nil, fmt.Errorf("report prose failed verification (%d violations): %s",
			len(violations), strings.Join(violations, " | "))
	}

	logger.Info("VerifyReportProseAction: prose verified against fact block")
	return map[string]interface{}{"verified": true, "sections": proseSectionCount}, nil
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
// are request strings (mounting, geometry, budget) the prose may echo.
func verifyReportProse(prose, scoring map[string]interface{}, contextValues []string) []string {
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

	sections := []string{"summary_html", "candidates_html", "integration_html", "vendor_questions_html"}
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
	}

	// (3) the honest no-match contract
	if matchCount == 0 {
		summaryText := proseTagRe.ReplaceAllString(prose["summary_html"].(string), " ")
		if !strings.Contains(summaryText, noMatchSentence) {
			violations = append(violations, fmt.Sprintf("match_count=0 but summary lacks the mandatory sentence %q", noMatchSentence))
		}
		for _, key := range sections {
			raw, _ := prose[key].(string)
			text := proseTagRe.ReplaceAllString(raw, " ")
			// The mandatory negative sentence itself contains "meets the
			// requirement" — remove it before scanning for softening.
			scan := strings.ReplaceAll(text, noMatchSentence, "")
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
