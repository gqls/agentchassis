// FILE: platform/orchestration/actions/verify_cited_cardinals_action.go
//
// VerifyCitedCardinalsAction is the deterministic gate between an LLM step that
// emits RANKED, SELF-ATTRIBUTING items and the write that persists them.
//
// The shape it guards: each item carries some prose plus the NAME of the source
// field the prose was drawn from ("from_field", "cited_field", …). That name is
// honesty machinery — a later reader is meant to be able to check the claim by
// reading the field it names. This action makes that checkable rather than
// merely stated: every cardinal quantity in the prose must actually appear in
// the field the item cites, or the item does not ship.
//
// Motivating defect (bugs_open/335). The offer-analyser wrote, at rank 1 on a
// live site, "the same stack that runs eight live sites built by this team",
// stamped from_field: "trust_threshold". The true count was 23; the number came
// from a page meta description passed in the offer surface, and the cited
// premise field contains no number at all. Nothing structural caught it — the
// artefact was well-formed and rank 1 is what a writer consumes first.
//
// WHY THIS IS NOT verify_report_prose. That action is the nearest precedent and
// the numeric-token idea is taken from it, but it could not have caught this
// one: its proseNumRe is `\d[\d,]*\.?\d*`, DIGITS ONLY, and the defect was the
// word "eight". Its own doc comment already flags that hole from the other end
// ("spelled-out numerals … are lower-case English too"). It is also bound to the
// gripper dossier's fact_block, sections and no-match sentence. So the reusable
// half is re-implemented here over a word-aware vocabulary, and this action is
// about ATTRIBUTION (does the number trace to the field this item CITES) rather
// than about one report's fact block.
//
// Config:
//   - object_field  (required) path to the object holding the items array
//   - items_key     (required) key within it holding the array
//   - source_field  (required) path to the object whose keys the items cite;
//                   a JSON-encoded string is parsed (query_database's
//                   `data::text` yields exactly that), and a string that is not
//                   an object is used whole as the source text
//   - text_key      (optional, default "point")       key holding the prose
//   - citation_key  (optional, default "from_field")  key naming the source field
//   - on_violation  (optional, default "fail")        "fail" | "drop"
//   - dropped_key   (optional, default "dropped_unsourced") where drop-mode
//                   records what it removed, inside the returned object
//
// Contract mirrors validate_page_content and verify_report_prose: in "fail" mode
// a violation returns (nil, error) so the workflow's error_step routes to the
// failure path, with the offending tokens in the error text and the log.
//
// "fail" is the DEFAULT because it is the arm that cannot quietly change an
// artefact; "drop" is opt-in per the owner ruling of 2026-08-02 §2 (new
// authority on a seam ships as an opt-in field with the unsafe side off).
// Drop mode is never silent: what it removed is written into the returned
// object under dropped_key, so the persisted artefact carries its own record.

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// VerifyCitedCardinalsInputSpec declares every step-config key this action
// reads under ConfigKeys, which opts it into unknown-config-key detection
// (datahelpers.ActionInputSpec.ConfigKeys). A brand-new action is the cheapest
// possible place to take that: there are no existing definitions to reject, so
// a typo in a future migration is reported rather than silently ignored.
//
// They are NOT in Optional, and that is deliberate rather than an oversight.
// Optional means "extracted by ExtractActionInputs"; this action resolves its
// own paths with ExtractNestedField, exactly as verify_report_prose does. The
// distinction is also what the RFC_022 optional-key budget counts — see
// cmd/config-key-audit/optionalbudget.go, which counts Optional only, on the
// stated ground that ConfigKeys are settings rather than accumulated authority.
// Declaring these in Optional to make the counter see them would misdescribe
// them and inflate a budget meant for a different harm.
var VerifyCitedCardinalsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
	ConfigKeys: []string{
		"object_field", "items_key", "source_field",
		"text_key", "citation_key", "dropped_key", "on_violation",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("verify_cited_cardinals", VerifyCitedCardinalsInputSpec)
}

// cardinalDigitRe matches digit-form quantities, including thousands separators
// and decimals. Same shape as verify_report_prose's proseNumRe — deliberately,
// so the two gates normalise "1,520" and "1520.0" identically.
var cardinalDigitRe = regexp.MustCompile(`\d[\d,]*\.?\d*`)

// cardinalUnits and cardinalTens are the word-numeral vocabulary.
//
// ⚠ "one" and "zero" are ABSENT from cardinalUnits and that is the single most
// load-bearing line in this file. Measured against all 30 live lead_with points
// on 2026-08-21: including them produced 6 flags of which 5 were false —
// "one click away", "in one workflow", "one of those categories", "the one you
// arrived with", "a restart from zero". Every one is an article, a pronoun or an
// idiom, not a quantity claim. Excluding them left exactly 1 flag: the defect
// this gate was written for. A gate at 17% precision is one an operator learns
// to wave through, which is how it stops being a gate at all.
//
// They ARE admitted on the SOURCE side (see cardinalSourceWords) — the asymmetry
// is deliberate. Excluding them from the point means "one" is never CHALLENGED;
// admitting them in the source means a premise that legitimately says "one
// engagement" still licenses a point that says "1 engagement".
var cardinalUnits = map[string]int{
	"two": 2, "three": 3, "four": 4, "five": 5, "six": 6, "seven": 7,
	"eight": 8, "nine": 9, "ten": 10, "eleven": 11, "twelve": 12,
	"thirteen": 13, "fourteen": 14, "fifteen": 15, "sixteen": 16,
	"seventeen": 17, "eighteen": 18, "nineteen": 19,
}

var cardinalTens = map[string]int{
	"twenty": 20, "thirty": 30, "forty": 40, "fifty": 50,
	"sixty": 60, "seventy": 70, "eighty": 80, "ninety": 90,
}

// cardinalSourceWords is the point vocabulary plus "one" and "zero". See the
// note on cardinalUnits for why the two sides differ.
var cardinalSourceWords = func() map[string]int {
	m := make(map[string]int, len(cardinalUnits)+2)
	for k, v := range cardinalUnits {
		m[k] = v
	}
	m["one"] = 1
	m["zero"] = 0
	return m
}()

// cardinalWordRe matches a tens word optionally followed by a unit word
// ("sixty-three", "sixty three"), or a bare unit word. Built for both
// vocabularies so the source side can admit one/zero without the point side
// doing so.
func cardinalWordRe(vocab map[string]int) *regexp.Regexp {
	units := make([]string, 0, len(vocab))
	for w := range vocab {
		units = append(units, w)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(units))) // longest-first: "seventeen" before "seven"
	tens := make([]string, 0, len(cardinalTens))
	for w := range cardinalTens {
		tens = append(tens, w)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(tens)))
	u := strings.Join(units, "|")
	t := strings.Join(tens, "|")
	// The separator class carries the hyphen, the non-breaking hyphen and the
	// en/em dashes: LLM prose writes "2–3" and "sixty–three" with any of them.
	return regexp.MustCompile(`(?i)\b(?:(?:` + t + `)(?:[-\x{2010}-\x{2015}\s](?:` + u + `))?|` + u + `|` + t + `)\b`)
}

var (
	cardinalPointWordRe  = cardinalWordRe(cardinalUnits)
	cardinalSourceWordRe = cardinalWordRe(cardinalSourceWords)
)

// normaliseCardinal canonicalises a digit token so "1,520", "1520" and "1520.0"
// compare equal.
func normaliseCardinal(tok string) string {
	clean := strings.ReplaceAll(tok, ",", "")
	clean = strings.TrimSuffix(clean, ".")
	if f, err := strconv.ParseFloat(clean, 64); err == nil {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return clean
}

// cardinalWordValue sums a matched word numeral ("sixty-three" → "63").
func cardinalWordValue(match string, vocab map[string]int) string {
	total := 0
	for _, part := range regexp.MustCompile(`[-\x{2010}-\x{2015}\s]`).Split(strings.ToLower(match), -1) {
		if v, ok := cardinalTens[part]; ok {
			total += v
			continue
		}
		if v, ok := vocab[part]; ok {
			total += v
		}
	}
	return strconv.Itoa(total)
}

// cardinalsIn returns the normalised numeric values in text. Word numerals are
// reduced to the same representation as digits, so a premise that says
// "sixty-three tools" licenses a point that says "63 tools" and vice versa. That
// relaxation is deliberate: the premise stated the quantity, and which form the
// writer chose is style, not sourcing.
func cardinalsIn(text string, wordRe *regexp.Regexp, vocab map[string]int) map[string]bool {
	out := make(map[string]bool)
	for _, tok := range cardinalDigitRe.FindAllString(text, -1) {
		out[normaliseCardinal(tok)] = true
	}
	for _, m := range wordRe.FindAllString(text, -1) {
		out[cardinalWordValue(m, vocab)] = true
	}
	return out
}

// cardinalSourceText flattens whatever the cited field resolved to into text to
// scan. A nested object (robot-hands' competitive_position is one) is serialised
// rather than skipped — the model is shown the whole premise object, so a value
// nested inside the cited field was genuinely available to it.
func cardinalSourceText(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

// resolveCardinalSource turns the configured source_field into a lookup map.
// query_database selects the strategy spec as `data::text`, so the common case
// is a JSON-encoded STRING rather than an object; parse it. A string that is not
// a JSON object is not an error — it is used whole, under the empty key, so a
// single-field source still works.
func resolveCardinalSource(raw interface{}) (map[string]interface{}, string) {
	switch v := raw.(type) {
	case map[string]interface{}:
		return v, ""
	case string:
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(datahelpers.StripCodeFences(v)), &parsed); err == nil {
			return parsed, ""
		}
		return nil, v
	default:
		return nil, cardinalSourceText(raw)
	}
}

type citedCardinalViolation struct {
	Index       int      `json:"index"`
	CitedField  string   `json:"cited_field"`
	FieldFound  bool     `json:"cited_field_found"`
	Unsourced   []string `json:"unsourced_cardinals"`
	Text        string   `json:"text"`
	Description string   `json:"description"`
}

func VerifyCitedCardinalsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "verify_cited_cardinals"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	objectField, _ := config["object_field"].(string)
	itemsKey, _ := config["items_key"].(string)
	sourceField, _ := config["source_field"].(string)
	if objectField == "" || itemsKey == "" || sourceField == "" {
		return nil, fmt.Errorf("verify_cited_cardinals requires object_field, items_key and source_field in step config")
	}

	textKey := stringOrDefault(config["text_key"], "point")
	citationKey := stringOrDefault(config["citation_key"], "from_field")
	droppedKey := stringOrDefault(config["dropped_key"], "dropped_unsourced")
	onViolation := strings.ToLower(strings.TrimSpace(stringOrDefault(config["on_violation"], "fail")))
	if onViolation != "fail" && onViolation != "drop" {
		return nil, fmt.Errorf("verify_cited_cardinals: on_violation must be \"fail\" or \"drop\", got %q", onViolation)
	}

	objRaw := datahelpers.ExtractNestedField(params.CollectedData, objectField)
	obj, ok := objRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("object_field %q did not resolve to an object (got %T)", objectField, objRaw)
	}
	itemsRaw, present := obj[itemsKey]
	if !present {
		return nil, fmt.Errorf("object at %q has no key %q", objectField, itemsKey)
	}
	items, ok := itemsRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("%s.%s is not an array (got %T)", objectField, itemsKey, itemsRaw)
	}

	sourceMap, sourceWhole := resolveCardinalSource(
		datahelpers.ExtractNestedField(params.CollectedData, sourceField))

	var violations []citedCardinalViolation
	kept := make([]interface{}, 0, len(items))

	for i, raw := range items {
		item, ok := raw.(map[string]interface{})
		if !ok {
			kept = append(kept, raw)
			continue
		}
		text, _ := item[textKey].(string)
		if strings.TrimSpace(text) == "" {
			kept = append(kept, raw)
			continue
		}
		cited, _ := item[citationKey].(string)

		srcText := sourceWhole
		fieldFound := sourceMap == nil && srcText != ""
		if sourceMap != nil {
			if v, present := sourceMap[cited]; present {
				srcText = cardinalSourceText(v)
				fieldFound = true
			} else {
				srcText = ""
			}
		}

		allowed := cardinalsIn(srcText, cardinalSourceWordRe, cardinalSourceWords)
		var unsourced []string
		for c := range cardinalsIn(text, cardinalPointWordRe, cardinalUnits) {
			if !allowed[c] {
				unsourced = append(unsourced, c)
			}
		}
		if len(unsourced) == 0 {
			kept = append(kept, raw)
			continue
		}
		sort.Slice(unsourced, func(a, b int) bool {
			fa, _ := strconv.ParseFloat(unsourced[a], 64)
			fb, _ := strconv.ParseFloat(unsourced[b], 64)
			return fa < fb
		})

		reason := fmt.Sprintf("cites %q, which does not contain %s",
			cited, strings.Join(unsourced, ", "))
		if !fieldFound {
			reason = fmt.Sprintf("cites %q, which is not a field of the source at all; %s unsourced",
				cited, strings.Join(unsourced, ", "))
		}
		violations = append(violations, citedCardinalViolation{
			Index: i, CitedField: cited, FieldFound: fieldFound,
			Unsourced: unsourced, Text: text, Description: reason,
		})
	}

	if len(violations) == 0 {
		logger.Info("verify_cited_cardinals: every cardinal traces to its cited field",
			zap.Int("items_checked", len(items)))
		return map[string]interface{}{
			"verified": true, "checked": len(items), "dropped": 0, "object": obj,
		}, nil
	}

	summary := make([]string, 0, len(violations))
	for _, v := range violations {
		summary = append(summary, fmt.Sprintf("item %d %s", v.Index, v.Description))
	}
	logger.Warn("verify_cited_cardinals: unsourced cardinals found",
		zap.Int("violations", len(violations)),
		zap.String("mode", onViolation),
		zap.Strings("detail", summary))

	if onViolation == "fail" {
		return nil, fmt.Errorf("%d of %d items assert a quantity absent from the field they cite: %s",
			len(violations), len(items), strings.Join(summary, " | "))
	}

	// Drop mode. An artefact with every item removed is a hollow artefact — the
	// class of failure where a row exists, reads as complete, and carries
	// nothing. Refuse it: whatever produced a wholly unsourced set is a defect
	// upstream, and writing the husk would hide it behind a successful step.
	if len(kept) == 0 {
		return nil, fmt.Errorf("all %d items assert unsourced quantities; refusing to write an empty %q: %s",
			len(items), itemsKey, strings.Join(summary, " | "))
	}

	corrected := make(map[string]interface{}, len(obj)+1)
	for k, v := range obj {
		corrected[k] = v
	}
	corrected[itemsKey] = kept
	corrected[droppedKey] = violations

	return map[string]interface{}{
		"verified": false, "checked": len(items),
		"dropped": len(violations), "object": corrected,
	}, nil
}

// stringOrDefault reads an optional string config literal.
func stringOrDefault(v interface{}, def string) string {
	if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
		return s
	}
	return def
}
