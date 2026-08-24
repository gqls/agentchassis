// FILE: platform/orchestration/actions/verify_acceptance_predicates_action.go
//
// VerifyAcceptancePredicatesAction is the deterministic gate between an LLM step
// that emits FINDINGS carrying their own acceptance test and the write that turns
// those findings into work items.
//
// THE SHAPE IT GUARDS. A finding says "this page is wrong" and states, in prose,
// what would make it right (`acceptance_test`). The prose is graded by judgement,
// and nothing checks it: [MEASURED 2026-08-24] of the 37 acceptance tests the
// offer-analyser has written, 4 are refuted RIGHT NOW by a one-line check over
// the exact field the test itself names, and 3 of those 4 sit on work items marked
// `complete` — the page was rebuilt, deployed, and the item closed, while its own
// stated criterion was unmet. The worked case is webdesign.co.uk's index, whose
// test reads "the meta description must state the zero-data or zero-account
// promise before any catalogue count" against a served meta description that
// opens "Sixty-three browser tools…".
//
// So this action lets a finding carry, ALONGSIDE the prose, a small structured
// predicate that a machine can evaluate — and it enforces the two rules that stop
// that predicate becoming a second, worse kind of false confidence:
//
//  1. A PREDICATE MAY ONLY EVER REFUTE. It is a NECESSARY condition of the prose
//     test, never a sufficient one. Satisfying it means "not refuted"; it never
//     means the test is met. Two thirds of live acceptance tests weld a checkable
//     clause to a judgement clause ("does not read as a generic contact-us
//     button"), and a green tick over the cheap half is worse than the prose it
//     replaced. The prose stays the authority.
//
//  2. A PREDICATE MUST REFUTE AT EMISSION, or it is DISCARDED HERE. The finding
//     asserts the page is wrong today, so a condition that genuinely expresses
//     the finding must FAIL today. This is the only property of a predicate that
//     can be checked at the moment it is written, and it is what stops the
//     vacuous case — a needle that appears nowhere, a clause already satisfied —
//     from being stored as honesty machinery and grading green for ever after.
//     Discards are recorded per finding under `acceptance_predicate_rejected`,
//     never dropped silently, so the cost of the rule is visible in the artefact.
//
// Rule 2 is deliberately strict and it does throw away useful predicates: a
// necessary condition that already holds could still refute later if a rewrite
// broke it. That trade is taken on the evidence of bugs_closed/335, where a field
// built to prove sourcing (`from_field`) vouched for a number the cited premise
// never contained. A predicate is exactly the same kind of self-attributing
// artefact — it asserts its own checkability — and the only version of it we can
// PROVE is coupled to the finding in front of us is one that fails in front of us.
//
// WHY NOT THE EXISTING CRITERIA VOCABULARY (selector_exists, attribute_matches,
// has_visible_area — experience_criteria.go). That vocabulary is DOM-shaped and
// browser-evaluated, and it was the first thing considered (the staged-component
// -build lane offered it to this lane on 2026-08-04). It is the wrong instrument
// for this population: these tests are about page METADATA served in the head, and
// the checkable clause is nearly always ordinary text arithmetic over one string.
// Nothing here needs a browser, a fence, or a tier.
//
// WHY NAV IS NOT IN THE VOCABULARY, although 3 of the 37 tests are about the
// header. Because the obvious source is wrong and it fails toward a CONFIDENT
// FALSE REFUTATION. [MEASURED 2026-08-24] leopardessconsulting.co.uk has 13 pages
// with `pages.in_header = true`; its SERVED header renders 7 destinations. A
// predicate reading the column would have refuted "the header nav contains no
// more than seven items" — a test that actually HOLDS at the artefact. The column
// that looks like the answer (`pages.rendered_header`) is empty on all 35 active
// pages of robot-hands.com. Nav needs the served page, so it waits for an
// instrument that reads one.
//
// WHY NOT datahelpers/claims.go's MATCHER, which owns the nearest-looking shape
// ("does this text contain this phrase") and was raised by the council's
// reuse_agent seat (corr ef482d1c). Its compiler is
// `regexp.Compile("(?i)" + p)` over a pattern an AUTHOR wrote into an evidence
// base — a REGEX, deliberately, with boundaries the author's business, and a
// QuoteMeta fallback when it will not compile. The needles here are LITERAL
// phrases an LLM emitted. Feeding one to that compiler makes "(beta)" a capture
// group and "3.5" match "375", or silently changes semantics via the fallback,
// and it drops the word boundaries that stop "we" matching inside "web". Same
// sentence, opposite contract: one interprets, this one quotes.
//
// AND WHY THIS IS NOT THE REVALIDATION FAMILY (revalidate_unverified_claims.go /
// check_unverified_claims.go / ScanDeployedClaims), which the same seat named as
// the closer precedent — correctly, and the boundary is worth stating so a third
// parallel mechanism does not grow here. That family asks "does the DEPLOYED
// component HTML still assert something the site's own evidence register does not
// support", over a fleet-wide register, to RETRACT a finding the page no longer
// supports. This asks "is the condition THIS finding says would make the page
// right false of the page's metadata today", over a per-finding author-supplied
// clause, to decide whether that clause may be STORED. ⚠ The two do converge at
// one point that is NOT built here: a completion-time consumer of these
// predicates would be asking a revalidation-shaped question, and whoever builds
// it should look at `reviewRevalidators` before writing a third loop.
//
// Config:
//   - site_id        (required) path to the site uuid
//   - findings_field (optional, default "audit_result.result") path to the
//     findings array, or to an object carrying one under "findings"
//
// CONTRACT WITH THE WRITE IT GUARDS. This action NEVER fails the workflow over a
// bad predicate and never removes a finding: the finding is the valuable part and
// the predicate is opt-in extra. It returns the findings verbatim apart from the
// two predicate keys.
//
// ⚠ AND IT MUST NOT INVENT AN EMPTY FINDINGS ARRAY. write_audit_findings treats a
// RECOGNISED empty list as the auditor saying "nothing is wrong here", which arms
// silence retraction (bugs_open/213 D1 half two — see parseAuditFindings' third
// return). An unresolvable findings path is NOT that statement, so this action
// omits the `findings` key entirely in that case, which drops write_audit_findings
// onto exactly the path it takes today: findingsRaw == nil, zero items, no
// retraction. An array that resolved and was genuinely empty is passed through AS
// an empty array, which preserves today's retraction behaviour too.
package actions

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// VerifyAcceptancePredicatesInputSpec mirrors WriteAuditFindingsInputSpec, whose
// step this one sits directly in front of: same site_id path, same findings_field
// path, so a migration wiring the pair cannot describe the two steps differently.
//
// There are deliberately NO other knobs. The predicate key names, the items key
// and the vocabulary are fixed literals in this file rather than config: every
// optional key is accumulated authority that some later caller inherits without
// reading this comment (RFC_022's optional-key budget counts exactly that), and
// this action has one consumer. A second consumer can earn a knob when it exists.
var VerifyAcceptancePredicatesInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{"findings_field"},
	Deprecated:  map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("verify_acceptance_predicates", VerifyAcceptancePredicatesInputSpec)
}

// ---------------------------------------------------------------------------
// The vocabulary
// ---------------------------------------------------------------------------

// acceptancePredicateKey / acceptancePredicateRejectedKey are SEPARATE keys, and
// that is the point rather than tidiness. A consumer that reads
// `acceptance_predicate` must never be able to reach a predicate this gate
// refused, so a refused one is not left in place under a status flag — it moves.
// Same shape as verify_cited_cardinals' `dropped_unsourced` sitting beside
// `lead_with` rather than inside it.
const (
	acceptancePredicateKey         = "acceptance_predicate"
	acceptancePredicateRejectedKey = "acceptance_predicate_rejected"
)

// acceptancePredicateFields is the closed key set per predicate type — the
// experienceCheckTypeFields pattern, for the same reason (P7): a key no checker
// reads is an assertion the artefact appears to make and does not. An unknown key
// rejects the predicate rather than being ignored.
var acceptancePredicateFields = map[string]map[string]bool{
	"text_absent":  {"field": true, "page": true, "values": true},
	"text_present": {"field": true, "page": true, "values": true, "min": true},
	"text_order":   {"field": true, "page": true, "before": true, "after": true},
}

// acceptancePredicateTextFields are the only readable fields, and both are served
// verbatim into the page head — verified at the artefact 2026-08-24: the
// `<meta name="description">` served by webdesign.co.uk is byte-identical to
// pages.meta_description. That is what makes a DB-side predicate honest about the
// served page. Page BODY text is deliberately absent: the offer surface passes no
// page content at all (features_open/030 v2(a)), so a predicate over body copy
// would be a claim about something its author never read.
var acceptancePredicateTextFields = map[string]bool{
	"meta_description": true,
	"title":            true,
}

// cardinalNeedle is the one reserved needle. "…before any count of tools" is the
// commonest ordering clause in the live corpus and it does not name a string, so
// it cannot be expressed as one.
const cardinalNeedle = "$cardinal"

// AcceptancePredicateVerdict is what evaluating one predicate produced.
type AcceptancePredicateVerdict string

const (
	// PredicateRefutes — the condition FAILS as things stand. For a predicate
	// emitted beside a live finding this is the WANTED verdict: it is the
	// evidence that the predicate expresses the finding.
	PredicateRefutes AcceptancePredicateVerdict = "refutes"
	// PredicateHolds — the condition is already satisfied, so it says nothing
	// about the defect the finding reports.
	PredicateHolds AcceptancePredicateVerdict = "holds"
	// PredicateInapplicable — malformed, unknown type or key, unreadable field,
	// or a page that is not on this site's surface. Never a pass and never a
	// fail: the predicate was not evaluated at all.
	PredicateInapplicable AcceptancePredicateVerdict = "inapplicable"
)

// AcceptancePredicateSubject is the text one predicate is evaluated over: one
// page's metadata, as stored and as served.
type AcceptancePredicateSubject struct {
	Page            string
	Title           string
	MetaDescription string
}

// AcceptancePredicateRejection is the per-finding record of a predicate this gate
// would not store, and why. It is written into the finding, so the work item
// carries its own account of what was refused.
type AcceptancePredicateRejection struct {
	Verdict   string                 `json:"verdict"`
	Reason    string                 `json:"reason"`
	Predicate map[string]interface{} `json:"predicate"`
}

// ---------------------------------------------------------------------------
// Needle matching
// ---------------------------------------------------------------------------

// needleRe compiles one needle into a word-boundary-anchored, case-insensitive
// matcher.
//
// ⚠ WORD BOUNDARIES ARE LOAD-BEARING, not defensive polish. The live corpus
// contains "no instance of 'we' or 'our'" (2026-08-17 census), and a plain
// substring search finds "we" inside "web", "answer" and "however" — on a site
// about WEB design. That is a false refutation, which is the one failure mode
// this whole design is built to avoid: it grades a passing page as failing, with
// a mechanical air. Boundaries are applied only where the needle's own edge is a
// word character, so "client-side", "no account" and "(beta)" are unaffected.
func needleRe(needle string) (*regexp.Regexp, error) {
	if needle == "" {
		return nil, fmt.Errorf("empty needle")
	}
	pat := regexp.QuoteMeta(needle)
	if isWordEdge(rune(needle[0])) {
		pat = `\b` + pat
	}
	if isWordEdge(rune(needle[len(needle)-1])) {
		pat = pat + `\b`
	}
	return regexp.Compile(`(?i)` + pat)
}

func isWordEdge(r rune) bool {
	return r == '_' || (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// needleMatch returns the earliest byte offset at which needle occurs in text
// together with THE TEXT THAT ACTUALLY MATCHED, or -1 and "".
//
// The matched text is returned, not the needle, because it is the evidence: a
// verdict reading `"$cardinal" appears at 0` tells a reader nothing they can
// check, and `"Sixty-three" appears at 0` tells them everything. That was a real
// defect in this file's first cut, caught by its own test.
//
// The reserved needle resolves to the first cardinal quantity, digits or words,
// using the SAME scanner as the attribution gate — so "63" and "Sixty-three" are
// both counts, and "B2B" and "IPv6" are not.
//
// ⚠ No case folding happens anywhere in this file. Every matcher is
// case-insensitive in its own right ((?i) on the needle and the word-cardinal
// regex, digits for the rest), so offsets are always into the ORIGINAL string —
// which keeps them comparable AND keeps the evidence quotable in the case the
// page actually used. strings.ToLower would break the second and, on non-ASCII,
// the first.
func needleMatch(text, needle string) (int, string, error) {
	if needle == cardinalNeedle {
		pos, matched := firstCardinalMatch(text)
		return pos, matched, nil
	}
	re, err := needleRe(needle)
	if err != nil {
		return -1, "", err
	}
	loc := re.FindStringIndex(text)
	if loc == nil {
		return -1, "", nil
	}
	return loc[0], text[loc[0]:loc[1]], nil
}

// firstCardinalMatch is the position half of cardinalsIn: the earliest offset of
// a quantity in text and the token found there, or (-1, "").
//
// It reuses digitCardinalSpans and cardinalPointWordRe from
// verify_cited_cardinals_action.go rather than re-deriving them. That matters
// because the rules there are not obvious and were each paid for by a live
// defect: digits welded into a word are not quantities ("B2B", "S3", "IPv6"), and
// "one"/"zero" are not quantities in prose ("one click away", "a restart from
// zero") — measured at 17% precision if admitted. A second implementation would
// drift away from both.
func firstCardinalMatch(text string) (int, string) {
	best, matched := -1, ""
	for _, span := range digitCardinalSpans(text) {
		if best == -1 || span[0] < best {
			best, matched = span[0], text[span[0]:span[1]]
		}
	}
	if loc := cardinalPointWordRe.FindStringIndex(text); loc != nil {
		if best == -1 || loc[0] < best {
			best, matched = loc[0], text[loc[0]:loc[1]]
		}
	}
	return best, matched
}

// ---------------------------------------------------------------------------
// Evaluation
// ---------------------------------------------------------------------------

// EvaluateAcceptancePredicate evaluates ONE predicate against ONE page's
// metadata and returns its verdict with a human-readable reason.
//
// Exported deliberately, and with no dependence on ActionParams or a DB handle:
// the second consumer this design expects is a COMPLETION-time check ("the
// handler reported success — does the item's own predicate still refute?"), which
// lives beside complete_work_item_no_change and must not re-derive the semantics.
// That consumer is NOT built here (see the register entry) — but the evaluator it
// needs is one call, not a rewrite.
func EvaluateAcceptancePredicate(pred map[string]interface{}, subject AcceptancePredicateSubject) (AcceptancePredicateVerdict, string) {
	typ, _ := pred["type"].(string)
	typ = strings.ToLower(strings.TrimSpace(typ))
	allowed, known := acceptancePredicateFields[typ]
	if !known {
		return PredicateInapplicable, fmt.Sprintf("unknown predicate type %q; this platform evaluates %s",
			typ, acceptancePredicateTypeList())
	}
	for k := range pred {
		if k == "type" || allowed[k] {
			continue
		}
		return PredicateInapplicable, fmt.Sprintf("key %q is not read by %s, so the predicate asserts less than it appears to", k, typ)
	}

	field, _ := pred["field"].(string)
	field = strings.ToLower(strings.TrimSpace(field))
	if !acceptancePredicateTextFields[field] {
		return PredicateInapplicable, fmt.Sprintf("field %q is not readable; this platform reads meta_description and title", field)
	}
	text := subject.MetaDescription
	if field == "title" {
		text = subject.Title
	}

	switch typ {
	case "text_absent":
		values, err := predicateNeedles(pred, "values")
		if err != nil {
			return PredicateInapplicable, err.Error()
		}
		for _, v := range values {
			if v == cardinalNeedle {
				return PredicateInapplicable, cardinalNeedle + " is only meaningful in text_order's \"after\""
			}
			pos, matched, err := needleMatch(text, v)
			if err != nil {
				return PredicateInapplicable, err.Error()
			}
			if pos >= 0 {
				return PredicateRefutes, fmt.Sprintf("%s of %q contains %q at %d", field, subject.Page, matched, pos)
			}
		}
		return PredicateHolds, fmt.Sprintf("%s of %q contains none of %s", field, subject.Page, quoteList(values))

	case "text_present":
		values, err := predicateNeedles(pred, "values")
		if err != nil {
			return PredicateInapplicable, err.Error()
		}
		minCount := 1
		if raw, present := pred["min"]; present {
			n, ok := predicateInt(raw)
			if !ok || n < 1 {
				return PredicateInapplicable, fmt.Sprintf("min must be a positive whole number, got %v", raw)
			}
			minCount = n
		}
		if minCount > len(values) {
			return PredicateInapplicable, fmt.Sprintf("min %d exceeds the %d values listed, so the predicate can never hold", minCount, len(values))
		}
		var found []string
		for _, v := range values {
			if v == cardinalNeedle {
				return PredicateInapplicable, cardinalNeedle + " is only meaningful in text_order's \"after\""
			}
			pos, _, err := needleMatch(text, v)
			if err != nil {
				return PredicateInapplicable, err.Error()
			}
			if pos >= 0 {
				found = append(found, v)
			}
		}
		if len(found) < minCount {
			return PredicateRefutes, fmt.Sprintf("%s of %q contains %d of the %d required values (%s), fewer than the %d demanded",
				field, subject.Page, len(found), len(values), quoteList(found), minCount)
		}
		return PredicateHolds, fmt.Sprintf("%s of %q already contains %s", field, subject.Page, quoteList(found))

	case "text_order":
		before, err := predicateNeedles(pred, "before")
		if err != nil {
			return PredicateInapplicable, err.Error()
		}
		after, err := predicateNeedles(pred, "after")
		if err != nil {
			return PredicateInapplicable, err.Error()
		}
		for _, v := range before {
			if v == cardinalNeedle {
				return PredicateInapplicable, cardinalNeedle + " is only meaningful in \"after\": \"state a quantity before something\" is not a claim this estate wants to make"
			}
		}
		posBefore := -1
		var firstBefore string
		for _, v := range before {
			pos, matched, err := needleMatch(text, v)
			if err != nil {
				return PredicateInapplicable, err.Error()
			}
			if pos >= 0 && (posBefore == -1 || pos < posBefore) {
				posBefore, firstBefore = pos, matched
			}
		}
		if posBefore == -1 {
			// "State X before Y" is unmet when X is not stated at all. This is
			// the arm that catches an EMPTY meta description, which is a real
			// state on this estate and a genuine failure of such a test.
			return PredicateRefutes, fmt.Sprintf("%s of %q states none of %s, so nothing can precede %s",
				field, subject.Page, quoteList(before), quoteList(after))
		}
		posAfter := -1
		var firstAfter string
		for _, v := range after {
			pos, matched, err := needleMatch(text, v)
			if err != nil {
				return PredicateInapplicable, err.Error()
			}
			if pos >= 0 && (posAfter == -1 || pos < posAfter) {
				posAfter, firstAfter = pos, matched
			}
		}
		if posAfter == -1 {
			return PredicateHolds, fmt.Sprintf("%s of %q states %q and contains none of %s at all",
				field, subject.Page, firstBefore, quoteList(after))
		}
		if posAfter < posBefore {
			return PredicateRefutes, fmt.Sprintf("in %s of %q, %q appears at %d, before %q at %d",
				field, subject.Page, firstAfter, posAfter, firstBefore, posBefore)
		}
		return PredicateHolds, fmt.Sprintf("in %s of %q, %q at %d precedes %q at %d",
			field, subject.Page, firstBefore, posBefore, firstAfter, posAfter)
	}

	return PredicateInapplicable, "unreachable: type validated above"
}

func acceptancePredicateTypeList() string {
	names := make([]string, 0, len(acceptancePredicateFields))
	for k := range acceptancePredicateFields {
		names = append(names, k)
	}
	// Fixed order, so an error message is stable enough to assert on.
	ordered := []string{"text_absent", "text_present", "text_order"}
	if len(ordered) != len(names) {
		return strings.Join(names, ", ")
	}
	return strings.Join(ordered, ", ")
}

// predicateNeedles reads a required non-empty array of non-empty strings.
func predicateNeedles(pred map[string]interface{}, key string) ([]string, error) {
	raw, present := pred[key]
	if !present {
		return nil, fmt.Errorf("%q is required and absent", key)
	}
	list, ok := raw.([]interface{})
	if !ok {
		// A single string where an array belongs is the commonest LLM slip and
		// costs nothing to accept.
		if s, isStr := raw.(string); isStr && strings.TrimSpace(s) != "" {
			return []string{strings.TrimSpace(s)}, nil
		}
		return nil, fmt.Errorf("%q must be an array of strings (got %T)", key, raw)
	}
	out := make([]string, 0, len(list))
	for _, item := range list {
		s, ok := item.(string)
		if !ok || strings.TrimSpace(s) == "" {
			return nil, fmt.Errorf("%q contains a non-string or empty entry", key)
		}
		out = append(out, strings.TrimSpace(s))
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%q is empty", key)
	}
	return out, nil
}

func predicateInt(raw interface{}) (int, bool) {
	switch v := raw.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	// JSON numbers arrive as float64 through every path this action is reached by.
	case float64:
		if v != float64(int(v)) {
			return 0, false
		}
		return int(v), true
	}
	return 0, false
}

func quoteList(vals []string) string {
	if len(vals) == 0 {
		return "(none)"
	}
	q := make([]string, 0, len(vals))
	for _, v := range vals {
		q = append(q, fmt.Sprintf("%q", v))
	}
	return strings.Join(q, ", ")
}

// ---------------------------------------------------------------------------
// The action
// ---------------------------------------------------------------------------

// VerifyAcceptancePredicatesAction validates, evaluates and filters the optional
// acceptance predicate on each finding, and passes every finding through.
func VerifyAcceptancePredicatesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "verify_acceptance_predicates"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, config, VerifyAcceptancePredicatesInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	findingsField := "audit_result.result"
	if f, ok := config["findings_field"].(string); ok && strings.TrimSpace(f) != "" {
		findingsField = strings.TrimSpace(f)
	}

	items, resolved := resolveFindingsList(datahelpers.ExtractNestedField(params.CollectedData, findingsField))
	if !resolved {
		// ⚠ Deliberately NOT an empty array — see the file header. Omitting the
		// key leaves write_audit_findings on its own unresolved-path branch,
		// which files nothing and arms no retraction.
		logger.Warn("verify_acceptance_predicates: findings path did not resolve to a list; passing nothing through",
			zap.String("findings_field", findingsField))
		return map[string]interface{}{
			"checked": 0, "kept": 0, "rejected": 0,
			"findings_resolved": false,
			"reason":            "no findings list at " + findingsField,
		}, nil
	}

	subjects, err := loadAcceptancePredicateSubjects(ctx, params, siteID)
	if err != nil {
		// Every predicate becomes inapplicable rather than the run failing: the
		// findings still deserve to be filed.
		logger.Warn("verify_acceptance_predicates: could not load page metadata; every predicate will be rejected as unevaluable",
			zap.Error(err))
		subjects = nil
	}
	// ⚠ AN EMPTY SUBJECT SET IS THE GATE'S OWN SILENT-INERT FAILURE MODE, and it
	// gets its own signal rather than arriving as N ordinary "page not found"
	// rejections. If the surface query ever stops matching — a lifecycle
	// vocabulary change, a site whose pages are all archived, a site_id that
	// resolves to the wrong row — every predicate is refused for a reason that
	// reads like the model's fault. Raised by two council seats (editquality,
	// debug_historian, corr ef482d1c) as the one thing rule 2 could not protect
	// against, because it fails toward "nothing was storable today", which is a
	// legitimate outcome. `subjects_loaded` is returned on EVERY run, including
	// the clean ones, so "0 pages" is a positive statement in the record rather
	// than an absence a reader has to notice.
	if len(subjects) == 0 {
		logger.Warn("verify_acceptance_predicates: NO pages loaded for this site — every predicate will be refused as unevaluable, and that is a fault here, not a model error",
			zap.String("site_id", siteID.String()),
			zap.Bool("query_errored", err != nil))
	}

	out := make([]interface{}, 0, len(items))
	var kept, rejected int
	rejections := make([]map[string]interface{}, 0)

	for i, raw := range items {
		item, ok := raw.(map[string]interface{})
		if !ok {
			out = append(out, raw)
			continue
		}
		predRaw, present := item[acceptancePredicateKey]
		if !present {
			// Silence is the expected answer for most findings, and it must stay
			// free: no key, no record, nothing to read.
			out = append(out, raw)
			continue
		}

		copied := make(map[string]interface{}, len(item)+1)
		for k, v := range item {
			copied[k] = v
		}
		delete(copied, acceptancePredicateKey)
		delete(copied, acceptancePredicateRejectedKey)

		pred, isObj := predRaw.(map[string]interface{})
		if !isObj {
			rej := AcceptancePredicateRejection{
				Verdict: string(PredicateInapplicable),
				Reason:  fmt.Sprintf("acceptance_predicate must be an object (got %T)", predRaw),
			}
			copied[acceptancePredicateRejectedKey] = rej
			rejected++
			rejections = append(rejections, map[string]interface{}{"index": i, "reason": rej.Reason})
			out = append(out, copied)
			continue
		}

		page := predicatePageName(pred, item)
		subject, found := subjects[page]
		if !found {
			reason := fmt.Sprintf("page %q is not on this site's active surface, so the predicate cannot be evaluated", page)
			if len(subjects) == 0 {
				// Deliberately NOT phrased as "page %q is not on the surface":
				// no page is, and blaming the named one sends the next reader
				// to the model when the fault is this step's.
				reason = "no pages were loaded for this site at all, so NO predicate could be evaluated — this is a fault in the gate's page query or its site_id, not in the predicate"
			}
			rej := AcceptancePredicateRejection{
				Verdict: string(PredicateInapplicable), Reason: reason, Predicate: pred,
			}
			copied[acceptancePredicateRejectedKey] = rej
			rejected++
			rejections = append(rejections, map[string]interface{}{"index": i, "page": page, "reason": reason})
			out = append(out, copied)
			continue
		}

		verdict, reason := EvaluateAcceptancePredicate(pred, subject)
		if verdict == PredicateRefutes {
			// The predicate is stored WITH the evidence that earned it its
			// place: a later reader (or the completion-time consumer) can see
			// that it was evaluated, when, and what it said — rather than
			// having to take "there is a predicate here" on trust.
			stored := make(map[string]interface{}, len(pred)+3)
			for k, v := range pred {
				stored[k] = v
			}
			stored["verdict_at_emission"] = string(PredicateRefutes)
			stored["evidence_at_emission"] = reason
			copied[acceptancePredicateKey] = stored
			kept++
			out = append(out, copied)
			continue
		}

		rej := AcceptancePredicateRejection{Verdict: string(verdict), Reason: reason, Predicate: pred}
		copied[acceptancePredicateRejectedKey] = rej
		rejected++
		rejections = append(rejections, map[string]interface{}{
			"index": i, "page": page, "verdict": string(verdict), "reason": reason,
		})
		out = append(out, copied)
	}

	if rejected > 0 {
		logger.Warn("verify_acceptance_predicates: predicates refused",
			zap.Int("kept", kept), zap.Int("rejected", rejected),
			zap.Any("rejections", rejections))
	} else {
		logger.Info("verify_acceptance_predicates: every predicate offered refutes the page as it stands",
			zap.Int("findings", len(items)), zap.Int("kept", kept))
	}

	return map[string]interface{}{
		"findings":          out,
		"checked":           len(items),
		"kept":              kept,
		"rejected":          rejected,
		"rejections":        rejections,
		"findings_resolved": true,
		"subjects_loaded":   len(subjects),
	}, nil
}

// resolveFindingsList accepts the same two native shapes parseAuditFindings does
// — a list, or an object carrying one under "findings" — and reports whether it
// resolved at all. It does NOT accept a JSON string: this action sits downstream
// of a step whose output is already parsed, and accepting a string here would let
// a malformed LLM reply reach the write as "recognised silence".
func resolveFindingsList(raw interface{}) ([]interface{}, bool) {
	switch v := raw.(type) {
	case []interface{}:
		return v, true
	case map[string]interface{}:
		if inner, ok := v["findings"].([]interface{}); ok {
			return inner, true
		}
	}
	return nil, false
}

// predicatePageName resolves which page a predicate is about: its own "page" if
// it names one, else the finding's. Aliases go through the same map
// write_audit_findings uses, so "homepage" resolves identically in the gate and
// in the write.
func predicatePageName(pred, item map[string]interface{}) string {
	name, _ := pred["page"].(string)
	if strings.TrimSpace(name) == "" {
		name, _ = item["page"].(string)
	}
	name = strings.TrimSpace(name)
	if alias, ok := pageNameAliases[strings.ToLower(name)]; ok {
		return alias
	}
	return name
}

// loadAcceptancePredicateSubjects reads the metadata of every page on the site's
// surface.
//
// ⚠ THE WHERE CLAUSE IS THE OFFER SURFACE'S OWN, deliberately. The model authors
// predicates against the page list it was shown; evaluating them over a DIFFERENT
// population would let a predicate be rejected for naming a page that WAS on the
// surface, or evaluated against a page that was not. The surface query lives in
// `load_offer_surface`'s step config (agent_definitions), so no compiler can hold
// the two together — if it changes, this must change with it.
//
// The lifecycle arm is PageWantedLivePredicateFor rather than a hand-written
// `status = 'active'`, on the landmine's own instruction ("prefer the helper"):
// pages.status has two live values and two of the spellings in circulation are
// INERT — `<> 'deleted'` excludes nothing, `IN ('active','deployed')` works only
// by accident. Two council seats (editquality, debug_historian) raised exactly
// that risk against the hand-written form, on the ground that a filter matching
// nothing would make every predicate "page not on this site's surface" and take
// the whole gate silently inert.
//
// [MEASURED 2026-08-24] the premise of that risk does not hold on this table
// today — `SELECT status, count(*) FROM pages GROUP BY 1` returns exactly
// `active` 805 / `archived` 66, and this query returns 35-137 rows for each of
// the five enrolled sites, never zero. The helper is used anyway, because it is
// the single place the vocabulary is written down, and because the measurement
// above is true of TODAY and the helper stays true after a vocabulary change.
// The silent-inert failure mode is separately made LOUD below — a measurement
// that a hazard is not live today is not a guard against it.
func loadAcceptancePredicateSubjects(ctx context.Context, params ActionParams, siteID uuid.UUID) (map[string]AcceptancePredicateSubject, error) {
	rows, err := params.DB.QueryContext(ctx, `
		SELECT name, COALESCE(title, ''), COALESCE(meta_description, '')
		FROM pages
		WHERE site_id = $1
		  AND `+datahelpers.PageWantedLivePredicateFor("")+`
		  AND NOT (deployed_at IS NULL AND COALESCE(build_status, '') = 'planned')
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]AcceptancePredicateSubject)
	for rows.Next() {
		var s AcceptancePredicateSubject
		if err := rows.Scan(&s.Page, &s.Title, &s.MetaDescription); err != nil {
			continue
		}
		out[s.Page] = s
	}
	return out, rows.Err()
}
